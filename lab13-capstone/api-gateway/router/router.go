package router

import (
	"api-gateway/middleware"
	"api-gateway/proxy"
	"api-gateway/ratelimiter"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func serviceURLs() (user, product, order, notification string) {
	user         = envOr("USER_SERVICE_URL",         "http://localhost:8081")
	product      = envOr("PRODUCT_SERVICE_URL",      "http://localhost:8082")
	order        = envOr("ORDER_SERVICE_URL",         "http://localhost:8083")
	notification = envOr("NOTIFICATION_SERVICE_URL", "http://localhost:8084")
	return
}

// publicPaths don't require authentication
var publicPaths = map[string]bool{
	"/api/users/register": true,
	"/api/users/login":    true,
	"/health":             true,
	"/health/all":         true,
}

func Build(rl *ratelimiter.RateLimiter) http.Handler {
	userURL, productURL, orderURL, notifURL := serviceURLs()

	mux := http.NewServeMux()

	// Health check (gateway's own health)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"api-gateway"}`)
	})

	// Aggregated health — fans out to all downstream services
	mux.HandleFunc("GET /health/all", func(w http.ResponseWriter, r *http.Request) {
		type serviceStatus struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		}
		svcs := []struct{ name, url string }{
			{"user-service", userURL},
			{"product-service", productURL},
			{"order-service", orderURL},
			{"notification-service", notifURL},
		}
		client := &http.Client{Timeout: 2 * time.Second}
		results := []serviceStatus{{Name: "api-gateway", Status: "up"}}
		for _, svc := range svcs {
			status := "down"
			resp, err := client.Get(svc.url + "/health")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					status = "up"
				}
			}
			results = append(results, serviceStatus{Name: svc.name, Status: status})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	// User Service routes
	mux.Handle("/api/users", proxy.Handler(userURL))
	mux.Handle("/api/users/", proxy.Handler(userURL))

	// Product Service routes
	mux.Handle("/api/products", proxy.Handler(productURL))
	mux.Handle("/api/products/", proxy.Handler(productURL))

	// Order Service routes
	mux.Handle("/api/orders", proxy.Handler(orderURL))
	mux.Handle("/api/orders/", proxy.Handler(orderURL))

	// Build middleware chain
	// Outermost: CORS → Recovery → Logging → RequestID → RateLimit → Auth → mux
	var handler http.Handler = mux
	handler = middleware.AuthMiddleware(publicPaths, handler)
	handler = middleware.RateLimitMiddleware(rl, handler)
	handler = middleware.RequestIDMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.RecoveryMiddleware(handler)
	handler = middleware.CORSMiddleware(handler)

	return handler
}
