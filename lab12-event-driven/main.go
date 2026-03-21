package main

import (
	"fmt"
	"lab12/domain"
	"lab12/eventbus"
	"lab12/events"
	"sync"
	"time"
)

// =============================================================================
// Inventory Subscriber
// =============================================================================

func subscribeInventory(bus *eventbus.EventBus) {
	bus.Subscribe(events.OrderPlaced, func(e eventbus.Event) {
		payload, ok := e.Payload.(events.OrderPlacedPayload)
		if !ok {
			return
		}
		// Simulate inventory reservation
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("[INVENTORY] Reserving stock for order %s (%d items)\n",
			payload.OrderID, len(payload.Items))

		// Publish inventory.reserved event back to the bus
		bus.Publish(eventbus.Event{
			Type:        events.InventoryReserved,
			AggregateID: payload.OrderID,
			Payload: events.InventoryReservedPayload{
				OrderID: payload.OrderID,
				Items:   payload.Items,
			},
		})
	})
}

// =============================================================================
// Notification Subscriber
// =============================================================================

func subscribeNotifications(bus *eventbus.EventBus) {
	// TODO: Subscribe to order.placed, order.cancelled, order.shipped
	// For each, print "[NOTIFICATION] Order <id>: <status message>"

	bus.Subscribe(events.OrderPlaced, func(e eventbus.Event) {
		// TODO: Print notification for order placed
	})

	bus.Subscribe(events.OrderCancelled, func(e eventbus.Event) {
		// TODO: Print notification for order cancelled
	})

	bus.Subscribe(events.OrderShipped, func(e eventbus.Event) {
		// TODO: Print notification for order shipped
	})
}

// =============================================================================
// Order Projector (Read Model)
// =============================================================================

type OrderView struct {
	ID         string
	UserID     string
	Status     string
	TotalPrice float64
	EventLog   []string
	UpdatedAt  time.Time
}

type OrderProjector struct {
	views map[string]*OrderView
	mu    sync.RWMutex
}

func NewOrderProjector() *OrderProjector {
	return &OrderProjector{views: make(map[string]*OrderView)}
}

func (p *OrderProjector) Subscribe(bus *eventbus.EventBus) {
	// TODO: Subscribe to all order events and update read model
	// Use bus.SubscribeAll or individual subscriptions

	bus.Subscribe(events.OrderPlaced, func(e eventbus.Event) {
		payload, ok := e.Payload.(events.OrderPlacedPayload)
		if !ok {
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.views[payload.OrderID] = &OrderView{
			ID:         payload.OrderID,
			UserID:     payload.UserID,
			Status:     "pending",
			TotalPrice: payload.TotalPrice,
			EventLog:   []string{fmt.Sprintf("placed at %s", e.OccurredAt.Format("15:04:05"))},
			UpdatedAt:  e.OccurredAt,
		}
	})

	bus.Subscribe(events.OrderCancelled, func(e eventbus.Event) {
		payload, ok := e.Payload.(events.OrderCancelledPayload)
		if !ok {
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if view, exists := p.views[payload.OrderID]; exists {
			// TODO: Update view status and add to EventLog
			view.Status = "cancelled"
			view.EventLog = append(view.EventLog,
				fmt.Sprintf("cancelled at %s: %s", e.OccurredAt.Format("15:04:05"), payload.Reason))
			view.UpdatedAt = e.OccurredAt
		}
	})

	bus.Subscribe(events.OrderShipped, func(e eventbus.Event) {
		payload, ok := e.Payload.(events.OrderShippedPayload)
		if !ok {
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if view, exists := p.views[payload.OrderID]; exists {
			// TODO: Update view status and add to EventLog
			_ = payload
		}
	})
}

func (p *OrderProjector) GetView(id string) (*OrderView, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.views[id]
	return v, ok
}

func (p *OrderProjector) ListViews() []*OrderView {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*OrderView, 0, len(p.views))
	for _, v := range p.views {
		result = append(result, v)
	}
	return result
}

// =============================================================================
// Order Command Handler
// =============================================================================

type OrderCommandHandler struct {
	orders map[string]*domain.Order
	bus    *eventbus.EventBus
	mu     sync.Mutex
	nextID int
}

func NewOrderCommandHandler(bus *eventbus.EventBus) *OrderCommandHandler {
	return &OrderCommandHandler{
		orders: make(map[string]*domain.Order),
		bus:    bus,
	}
}

func (h *OrderCommandHandler) generateID() string {
	h.nextID++
	return fmt.Sprintf("ORD-%03d", h.nextID)
}

func (h *OrderCommandHandler) PlaceOrder(userID string, items []events.OrderItem) (*domain.Order, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.generateID()
	order, err := domain.NewOrder(id, userID, items)
	if err != nil {
		return nil, err
	}

	h.orders[order.ID] = order

	// Publish uncommitted events
	for _, evt := range order.UncommittedEvents() {
		h.bus.Publish(evt)
	}
	order.ClearEvents()

	return order, nil
}

func (h *OrderCommandHandler) CancelOrder(orderID, reason string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	order, ok := h.orders[orderID]
	if !ok {
		return fmt.Errorf("order %s not found", orderID)
	}

	if err := order.Cancel(reason); err != nil {
		return err
	}

	// TODO: Publish uncommitted events and clear them
	return nil
}

func (h *OrderCommandHandler) ShipOrder(orderID, trackingNumber string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	order, ok := h.orders[orderID]
	if !ok {
		return fmt.Errorf("order %s not found", orderID)
	}

	if err := order.Ship(trackingNumber); err != nil {
		return err
	}

	// TODO: Publish uncommitted events and clear them
	return nil
}

// =============================================================================
// Main
// =============================================================================

func main() {
	bus := eventbus.New()
	projector := NewOrderProjector()

	// Subscribe handlers BEFORE publishing events
	projector.Subscribe(bus)
	subscribeInventory(bus)
	subscribeNotifications(bus)

	handler := NewOrderCommandHandler(bus)

	items1 := []events.OrderItem{
		{ProductID: "P001", ProductName: "Laptop", Quantity: 1, UnitPrice: 999.99},
		{ProductID: "P002", ProductName: "Mouse", Quantity: 2, UnitPrice: 29.99},
	}
	items2 := []events.OrderItem{
		{ProductID: "P003", ProductName: "Keyboard", Quantity: 1, UnitPrice: 79.99},
	}
	items3 := []events.OrderItem{
		{ProductID: "P001", ProductName: "Laptop", Quantity: 2, UnitPrice: 999.99},
	}

	fmt.Println("=== Placing Orders ===")
	order1, err := handler.PlaceOrder("user-001", items1)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Placed: %s (total: $%.2f)\n", order1.ID, order1.TotalPrice)
	}

	order2, err := handler.PlaceOrder("user-002", items2)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Placed: %s (total: $%.2f)\n", order2.ID, order2.TotalPrice)
	}

	order3, err := handler.PlaceOrder("user-001", items3)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Placed: %s (total: $%.2f)\n", order3.ID, order3.TotalPrice)
	}

	// Wait for async handlers
	bus.WaitForAll()

	fmt.Println("\n=== Shipping Order 1 ===")
	if order1 != nil {
		if err := handler.ShipOrder(order1.ID, "TRACK-12345"); err != nil {
			fmt.Println("Error:", err)
		}
	}

	fmt.Println("\n=== Cancelling Order 2 ===")
	if order2 != nil {
		if err := handler.CancelOrder(order2.ID, "Customer requested cancellation"); err != nil {
			fmt.Println("Error:", err)
		}
	}

	// Wait for all async handlers to complete
	bus.WaitForAll()

	fmt.Println("\n=== Order Read Model (Projector) ===")
	for _, view := range projector.ListViews() {
		fmt.Printf("\nOrder %s | User: %s | Status: %s | Total: $%.2f\n",
			view.ID, view.UserID, view.Status, view.TotalPrice)
		for _, entry := range view.EventLog {
			fmt.Printf("  - %s\n", entry)
		}
	}
}
