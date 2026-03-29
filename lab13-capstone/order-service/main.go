package main

// Order Service — Port 8083
//
// This is the most complex service. It integrates:
//   - Clean architecture (domain/usecase/repository/handler layers)
//   - State machine for order transitions (Lab 06)
//   - Worker pool for processing orders concurrently (Lab 03)
//   - HTTP service client with circuit breaker (Lab 10)
//   - Domain events published to shared event bus (Lab 12)
//
// ORDER FLOW:
//   1. POST /api/orders → validate product+stock → create Pending order → queue for processing
//   2. Worker goroutine picks up order → simulates payment → transitions state → publishes event
//
// Endpoints:
//   GET  /health
//   POST /api/orders      { productId, quantity, userId }
//   GET  /api/orders
//   GET  /api/orders/{id}

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"order-service/clients"
	"order-service/eventbus"
	"order-service/handlers"
	"order-service/processor"
	"order-service/repository"
	"order-service/usecase"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	bus := eventbus.NewEventBus()
	store := repository.NewMemoryOrderStore()
	uc := usecase.NewOrderUseCase(store)
	proc := processor.NewOrderProcessor(store, bus)
	client := clients.NewProductServiceClient("http://localhost:8082")
	handler := handlers.NewOrderHandler(client, uc, proc, bus)
	var workerWG sync.WaitGroup
	proc.Start(3, &workerWG)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	srv := &http.Server{
		Addr:         ":8083",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Order Service on http://localhost:8083")
		fmt.Println("(calls Product Service at http://localhost:8082)")
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Order Service shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	proc.Stop()      // Signal workers to stop accepting new jobs
	workerWG.Wait()  // Wait for in-flight orders to finish
	bus.WaitForAll() // Wait for any pending event handlers
	fmt.Println("Order Service stopped gracefully")
}
