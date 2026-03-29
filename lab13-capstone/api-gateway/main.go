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
	"api-gateway/ratelimiter"
	"api-gateway/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rateLimiter := ratelimiter.NewRateLimiter()
	handler := router.Build(rateLimiter)

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
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("API Gateway shutdown error: %v", err)
	}
	fmt.Println("API Gateway stopped gracefully")
}
