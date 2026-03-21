package main

import (
	"context"
	"fmt"
	"lab08/handlers"
	"lab08/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	// Create store (seeded with sample data)
	s := store.NewProductStore()

	// Create handlers
	h := handlers.NewProductHandler(s)

	// Register routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Apply middleware
	handler := loggingMiddleware(mux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		fmt.Println("Product API server on http://localhost:8080")
		fmt.Println()
		fmt.Println("Endpoints:")
		fmt.Println("  GET    /api/products              List all products")
		fmt.Println("  GET    /api/products?category=X   Filter by category")
		fmt.Println("  GET    /api/products?page=1&limit=2  Paginated")
		fmt.Println("  GET    /api/products/{id}          Get by ID")
		fmt.Println("  POST   /api/products               Create product")
		fmt.Println("  PATCH  /api/products/{id}          Partial update")
		fmt.Println("  DELETE /api/products/{id}          Delete")
		fmt.Println()
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	fmt.Println("Server stopped gracefully")
}
