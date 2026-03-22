package query

import (
	"fmt"
	"lab12/eventbus"
	"lab12/events"
	"sync"
	"time"
)

// OrderView is the denormalized read model for an order.
// It is updated by the OrderProjector whenever an order event is received.
type OrderView struct {
	ID         string
	UserID     string
	Status     string
	TotalPrice float64
	EventLog   []string // human-readable log of all events that affected this order
	UpdatedAt  time.Time
}

// OrderProjector listens to order events and maintains the OrderView read model.
type OrderProjector struct {
	views map[string]*OrderView
	mu    sync.RWMutex
}

func NewOrderProjector() *OrderProjector {
	return &OrderProjector{views: make(map[string]*OrderView)}
}

// Subscribe registers this projector on the event bus.
// It must be called BEFORE any events are published.
func (p *OrderProjector) Subscribe(bus *eventbus.EventBus) {
	bus.Subscribe(events.OrderPlaced, func(e eventbus.Event) {
		payload, ok := e.Payload.(events.OrderPlacedPayload)
		if !ok {
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		// Create the view first
		p.views[payload.OrderID] = &OrderView{
			ID:         payload.OrderID,
			UserID:     payload.UserID,
			Status:     "pending",
			TotalPrice: payload.TotalPrice,
			EventLog:   []string{},
			UpdatedAt:  time.Now(),
		}
		// Now safe to append
		p.views[payload.OrderID].EventLog = append(
			p.views[payload.OrderID].EventLog,
			fmt.Sprintf("placed at %s", e.OccurredAt.Format("15:04:05")),
		)
	})

	bus.Subscribe(events.OrderCancelled, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.OrderCancelledPayload
		payload, ok := e.Payload.(events.OrderCancelledPayload)
		if !ok {
			return
		}
		// TODO: Lock, look up view by OrderID
		p.mu.Lock()
		defer p.mu.Unlock()
		// TODO: Update status to "cancelled", append to EventLog, update UpdatedAt
		p.views[payload.OrderID].Status = "cancelled"
		p.views[payload.OrderID].EventLog = append(p.views[payload.OrderID].EventLog, "Order has been cancelled")
		p.views[payload.OrderID].UpdatedAt = time.Now()
	})

	bus.Subscribe(events.OrderShipped, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.OrderShippedPayload
		payload, ok := e.Payload.(events.OrderShippedPayload)
		if !ok {
			return
		}
		// TODO: Lock, look up view by OrderID
		p.mu.Lock()
		defer p.mu.Unlock()
		// TODO: Update status to "shipped", append to EventLog with tracking number, update UpdatedAt
		p.views[payload.OrderID].Status = "shipped"
		p.views[payload.OrderID].EventLog = append(p.views[payload.OrderID].EventLog, fmt.Sprintf("Order %s has been shipped with tracking number %s", payload.OrderID, payload.TrackingNumber))
		p.views[payload.OrderID].UpdatedAt = time.Now()

	})

	bus.Subscribe(events.InventoryReserved, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.InventoryReservedPayload
		payload, ok := e.Payload.(events.InventoryReservedPayload)
		if !ok {
			return
		}
		// TODO: Lock, look up view by OrderID
		p.mu.Lock()
		defer p.mu.Unlock()
		// TODO: Append to EventLog: "inventory reserved"
		p.views[payload.OrderID].EventLog = append(p.views[payload.OrderID].EventLog, fmt.Sprintf("Inventory reserved"))
	})
}

// GetView returns the read model for a specific order.
func (p *OrderProjector) GetView(id string) (*OrderView, bool) {
	// TODO: RLock, look up and return p.views[id]
	p.mu.RLock()
	defer p.mu.RUnlock()

	view, ok := p.views[id]
	if !ok {
		return nil, false
	}
	return view, true
}

// ListViews returns all order views.
func (p *OrderProjector) ListViews() []*OrderView {
	// TODO: RLock, collect all views from p.views into a slice and return
	sli := make([]*OrderView, 0)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, value := range p.views {
		sli = append(sli, value)
	}

	return sli
}
