package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"notification-service/workers"
	"sync"
	"time"
)

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

type OrderPoller struct {
	orderServiceURL string
	worker          *workers.NotificationWorker
	mu              sync.Mutex
	lastStatuses    map[string]OrderStatus
}

func NewOrderPoller(orderServiceURL string, worker *workers.NotificationWorker) *OrderPoller {
	return &OrderPoller{
		orderServiceURL: orderServiceURL,
		worker:          worker,
		lastStatuses:    make(map[string]OrderStatus),
	}
}

func (p *OrderPoller) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.poll()
		case <-ctx.Done():
			return
		}
	}
}

func (p *OrderPoller) poll() {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(p.orderServiceURL + "/api/orders")
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
			p.lastStatuses[order.ID] = order.Status
			p.worker.Submit(workers.Notification{
				OrderID: order.ID,
				UserID:  order.UserID,
				Type:    "email",
				Message: fmt.Sprintf("Your order %s has been placed!", order.ID),
			})
			continue
		}
		if lastStatus != order.Status {
			p.lastStatuses[order.ID] = order.Status
			msg := statusMessage(order.Status, order.ID)
			if msg != "" {
				p.worker.Submit(workers.Notification{
					OrderID: order.ID,
					UserID:  order.UserID,
					Type:    statusNotificationType(order.Status),
					Message: msg,
				})
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
