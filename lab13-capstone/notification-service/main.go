package main

// Notification Service — Port 8084
//
// This service polls for order events via HTTP long-polling or webhook.
//
// Since all services in this lab are separate Go modules (separate processes),
// they can't share an in-process channel directly. Instead, this service
// subscribes to the Order Service via HTTP webhook callbacks.
//
// DESIGN: Simple polling approach (easier than webhooks for this lab)
//   - Notification service calls GET /api/orders every 2 seconds
//   - Tracks which orders it has already notified about
//   - When order status changes, sends notification
//
// For a production implementation, use a message broker (Kafka, NATS, RabbitMQ).
//
// Endpoints:
//   GET /health
//   POST /webhook/order  → Order Service calls this when order status changes (bonus)

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"notification-service/handlers"
	"notification-service/poller"
	"notification-service/workers"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const OrderServiceURL string = "http://localhost:8083"

func main() {
	// Start notification worker pool
	worker := workers.NewNotificationWorker(50)
	handler := handlers.NewNotificationHandler()
	poller := poller.NewOrderPoller(OrderServiceURL, worker)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	var pollerWG sync.WaitGroup
	pollerWG.Add(1)
	go func() {
		defer pollerWG.Done()
		poller.Start(ctx, 2*time.Second)
	}()
	go worker.Start(2, &wg)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8084",
		Handler: mux,
	}

	go func() {
		fmt.Println("Notification Service on http://localhost:8084")
		fmt.Println("Polling Order Service every 2 seconds...")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()        // 1. tell poller to stop
	pollerWG.Wait() // 2. wait for poller to fully exit

	worker.Stop() // 3. now safe to close the queue
	wg.Wait()     // 4. drain in-flight notifications

	// 5. shut down HTTP
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

}
