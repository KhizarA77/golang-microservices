package main

// Product Service — Port 8082
//
// INSTRUCTIONS:
// Port your Lab 08 REST API solution here.
// Additions for the capstone:
//   1. Caching proxy (Lab 05): cache individual product responses for 30s
//   2. Concurrent search (Lab 03): SearchProducts fans out to 3 search "strategies"
//   3. Stock management: products have a stock count that decreases when ordered
//
// Endpoints:
//   GET    /health
//   GET    /api/products              list (paginated)
//   GET    /api/products/search?q=... concurrent search
//   GET    /api/products/{id}         get by ID (cached)
//   POST   /api/products              create
//   PATCH  /api/products/{id}         update
//   DELETE /api/products/{id}         delete
//   PATCH  /api/products/{id}/stock   { "delta": -1 } reserve or release stock
//
// The /api/products/{id}/stock endpoint is used by the Order Service
// to reserve stock when an order is placed.

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"product-service/cache"
	"product-service/handlers"
	"product-service/store"
	"syscall"
	"time"
)

func main() {
	store := store.NewProductStore()
	cache := cache.NewProductCache()

	productHandler := handlers.NewProductHandler(store, cache)

	mux := http.NewServeMux()

	productHandler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Product Service on http://localhost:8082")
		fmt.Println("=========== ROUTES ===========")
		fmt.Println("GET    /health")
		fmt.Println("GET    /api/products              list (paginated)")
		fmt.Println("GET    /api/products/search?q=... concurrent search")
		fmt.Println("GET    /api/products/{id}         get by ID (cached)")
		fmt.Println("POST   /api/products              create")
		fmt.Println("PATCH  /api/products/{id}         update")
		fmt.Println("DELETE /api/products/{id}         delete")
		fmt.Println(`PATCH  /api/products/{id}/stock   { "delta": -1 } reserve or release stock`)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Product Service shutting down...")
}
