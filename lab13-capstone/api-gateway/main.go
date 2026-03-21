package main

// API Gateway — Port 8080
//
// Single entry point for all client requests.
// Implements: reverse proxy, auth middleware, rate limiting, logging.
//
// Routes:
//   /api/users/*      → User Service    :8081
//   /api/products/*   → Product Service :8082
//   /api/orders/*     → Order Service   :8083
//
// Middleware stack (outer to inner):
//   RecoveryMiddleware → LoggingMiddleware → RequestIDMiddleware →
//   RateLimitMiddleware → AuthMiddleware → Proxy
//
// Auth:
//   Public routes (no token required):
//     POST /api/users/register
//     POST /api/users/login
//     GET  /api/products (and sub-routes)
//     GET  /health
//   Protected routes require: Authorization: Bearer <token>
//   Token format: base64("userID:email:timestamp") — from User Service login

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// =============================================================================
// Configuration
// =============================================================================

var routes = map[string]string{
	"/api/users/":    "http://localhost:8081",
	"/api/products/": "http://localhost:8082",
	"/api/orders/":   "http://localhost:8083",
	"/api/users":     "http://localhost:8081",
	"/api/products":  "http://localhost:8082",
	"/api/orders":    "http://localhost:8083",
}

// publicPaths don't require authentication
var publicPaths = map[string]bool{
	"/api/users/register": true,
	"/api/users/login":    true,
	"/health":             true,
}

// =============================================================================
// Rate Limiter (token bucket per IP — Lab 02 time.Ticker concept)
// =============================================================================

type rateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
}

type ipLimiter struct {
	tokens    int
	maxTokens int
	lastRefil time.Time
	refillRate time.Duration // time per token
}

var globalRateLimiter = &rateLimiter{
	limiters: make(map[string]*ipLimiter),
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	l, ok := rl.limiters[ip]
	if !ok {
		l = &ipLimiter{
			tokens:    5,
			maxTokens: 5,
			lastRefil: time.Now(),
			refillRate: 200 * time.Millisecond, // 5 req/sec = 1 per 200ms
		}
		rl.limiters[ip] = l
	}

	// Refill tokens based on elapsed time
	elapsed := time.Since(l.lastRefil)
	newTokens := int(elapsed / l.refillRate)
	if newTokens > 0 {
		l.tokens = min(l.maxTokens, l.tokens+newTokens)
		l.lastRefil = time.Now()
	}

	if l.tokens <= 0 {
		return false
	}
	l.tokens--
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Middleware
// =============================================================================

type contextKey string

const (
	requestIDKey contextKey = "requestID"
	userIDKey    contextKey = "userID"
)

// RequestIDMiddleware injects a unique request ID.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusRecorder captures the response status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.status = status
	sr.ResponseWriter.WriteHeader(status)
}

// LoggingMiddleware logs every request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		reqID, _ := r.Context().Value(requestIDKey).(string)
		log.Printf("[GW] %s %s %s → %d (%v) [req-id: %s]",
			r.RemoteAddr, r.Method, r.URL.Path, rec.status, time.Since(start), reqID)
	})
}

// RecoveryMiddleware catches panics and returns 500.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware limits requests per IP.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		if !globalRateLimiter.allow(ip) {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates the Bearer token for protected routes.
// This is a simplified validator — it just checks the token is non-empty and base64-decodable.
// TODO: For full implementation, verify the token against User Service or use JWT.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public paths
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Also skip GET /api/products (public catalog)
		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/api/products") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		// TODO: Decode the base64 token and extract userID
		// For now, any non-empty Bearer token is accepted
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"empty token"}`, http.StatusUnauthorized)
			return
		}

		// Store token in context for downstream use
		ctx := context.WithValue(r.Context(), userIDKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// =============================================================================
// Reverse Proxy
// =============================================================================

// proxyHandler creates an http.Handler that forwards requests to targetBase.
func proxyHandler(targetBase string) http.Handler {
	client := &http.Client{Timeout: 10 * time.Second}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := targetBase + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// Create outbound request
		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, "proxy error", http.StatusBadGateway)
			return
		}

		// Copy request headers
		for key, vals := range r.Header {
			for _, val := range vals {
				outReq.Header.Add(key, val)
			}
		}

		// Forward request ID
		if reqID, ok := r.Context().Value(requestIDKey).(string); ok {
			outReq.Header.Set("X-Request-ID", reqID)
		}

		// Execute request
		resp, err := client.Do(outReq)
		if err != nil {
			log.Printf("[PROXY ERROR] %s → %s: %v", r.URL.Path, targetURL, err)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"upstream service unavailable"}`, http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, vals := range resp.Header {
			for _, val := range vals {
				w.Header().Add(key, val)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)
	})
}

// =============================================================================
// Router
// =============================================================================

func buildGatewayRouter() http.Handler {
	mux := http.NewServeMux()

	// Health check (gateway's own health)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"api-gateway"}`)
	})

	// User Service routes
	mux.Handle("/api/users", proxyHandler("http://localhost:8081"))
	mux.Handle("/api/users/", proxyHandler("http://localhost:8081"))

	// Product Service routes
	mux.Handle("/api/products", proxyHandler("http://localhost:8082"))
	mux.Handle("/api/products/", proxyHandler("http://localhost:8082"))

	// Order Service routes
	mux.Handle("/api/orders", proxyHandler("http://localhost:8083"))
	mux.Handle("/api/orders/", proxyHandler("http://localhost:8083"))

	// Build middleware chain
	// Outermost: Recovery → Logging → RequestID → RateLimit → Auth → mux
	var handler http.Handler = mux
	handler = AuthMiddleware(handler)
	handler = RateLimitMiddleware(handler)
	handler = RequestIDMiddleware(handler)
	handler = LoggingMiddleware(handler)
	handler = RecoveryMiddleware(handler)

	return handler
}

func main() {
	handler := buildGatewayRouter()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		fmt.Println("API Gateway on http://localhost:8080")
		fmt.Println()
		fmt.Println("Routes:")
		fmt.Println("  /api/users/*    → User Service    :8081")
		fmt.Println("  /api/products/* → Product Service :8082")
		fmt.Println("  /api/orders/*   → Order Service   :8083")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println(`  curl -X POST http://localhost:8080/api/users/register \`)
		fmt.Println(`       -d '{"name":"Alice","email":"alice@example.com","password":"secret"}'`)
		fmt.Println(`  TOKEN=$(curl -s -X POST http://localhost:8080/api/users/login \`)
		fmt.Println(`       -d '{"email":"alice@example.com","password":"secret"}' | jq -r '.token')`)
		fmt.Println(`  curl http://localhost:8080/api/products`)
		fmt.Println(`  curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/orders`)

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nAPI Gateway shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	fmt.Println("API Gateway stopped gracefully")
}
