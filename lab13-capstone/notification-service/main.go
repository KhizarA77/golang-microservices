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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// OrderStatus mirrors order-service types
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID     string      `json:"id"`
	UserID string      `json:"user_id"`
	Status OrderStatus `json:"status"`
}

// =============================================================================
// Notification Worker Pool (Lab 03)
// =============================================================================

type Notification struct {
	OrderID string
	UserID  string
	Type    string
	Message string
}

var notificationQueue = make(chan Notification, 100)

func startNotificationWorkers(n int) {
	for i := 1; i <= n; i++ {
		go func(workerID int) {
			for notif := range notificationQueue {
				// Simulate sending notification (email/SMS)
				time.Sleep(50 * time.Millisecond)
				log.Printf("[NOTIFICATION WORKER %d] %s for order %s (user: %s): %s",
					workerID, notif.Type, notif.OrderID, notif.UserID, notif.Message)
			}
		}(i)
	}
}

// =============================================================================
// Order Poller
// =============================================================================

type orderPoller struct {
	mu           sync.Mutex
	lastStatuses map[string]OrderStatus // orderID → last known status
}

func newOrderPoller() *orderPoller {
	return &orderPoller{lastStatuses: make(map[string]OrderStatus)}
}

func (p *orderPoller) poll() {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:8083/api/orders")
	if err != nil {
		log.Printf("[POLLER] could not reach order service: %v", err)
		return
	}
	defer resp.Body.Close()

	var orders []Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		log.Printf("[POLLER] decode error: %v", err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, order := range orders {
		lastStatus, seen := p.lastStatuses[order.ID]

		if !seen {
			// New order — send "placed" notification
			p.lastStatuses[order.ID] = order.Status
			notificationQueue <- Notification{
				OrderID: order.ID,
				UserID:  order.UserID,
				Type:    "email",
				Message: fmt.Sprintf("Your order %s has been placed!", order.ID),
			}
			continue
		}

		if lastStatus != order.Status {
			// Status changed — send notification
			p.lastStatuses[order.ID] = order.Status
			msg := statusMessage(order.Status, order.ID)
			if msg != "" {
				notificationQueue <- Notification{
					OrderID: order.ID,
					UserID:  order.UserID,
					Type:    statusNotificationType(order.Status),
					Message: msg,
				}
			}
		}
	}
}

func statusMessage(status OrderStatus, orderID string) string {
	switch status {
	case StatusProcessing:
		return fmt.Sprintf("Your order %s is being processed.", orderID)
	case StatusShipped:
		return fmt.Sprintf("Great news! Your order %s has shipped.", orderID)
	case StatusDelivered:
		return fmt.Sprintf("Your order %s has been delivered. Enjoy!", orderID)
	case StatusCancelled:
		return fmt.Sprintf("Your order %s has been cancelled.", orderID)
	}
	return ""
}

func statusNotificationType(status OrderStatus) string {
	if status == StatusShipped || status == StatusDelivered {
		return "sms"
	}
	return "email"
}

// =============================================================================
// Webhook Handler (bonus — for when Order Service calls us)
// =============================================================================

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse incoming event from Order Service
	// TODO: Queue notification based on event type
	// This is an exercise — the polling approach already works above
	w.WriteHeader(http.StatusOK)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"ok","service":"notification-service"}`)
}

func main() {
	// Start notification worker pool
	startNotificationWorkers(2)

	// Start order poller
	poller := newOrderPoller()
	pollTicker := time.NewTicker(2 * time.Second)
	go func() {
		for range pollTicker.C {
			poller.poll()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /webhook/order", handleWebhook)

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

	pollTicker.Stop()
	close(notificationQueue)
	fmt.Println("Notification Service stopped")
}
