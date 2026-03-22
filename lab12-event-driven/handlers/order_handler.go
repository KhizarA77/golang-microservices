package handlers

import (
	"fmt"
	"lab12/domain"
	"lab12/eventbus"
	"sync"
)

// OrderCommandHandler processes order commands and publishes resulting domain events.
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

// HandlePlaceOrder executes a PlaceOrderCommand:
//  1. Generate an ID, create the order via domain.NewOrder
//  2. Store the order in the in-memory map
//  3. Publish all uncommitted events to the bus
//  4. Clear uncommitted events on the order
func (h *OrderCommandHandler) HandlePlaceOrder(cmd domain.PlaceOrderCommand) (*domain.Order, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// TODO: Generate ID
	id := h.generateID()
	// TODO: Call domain.NewOrder(id, cmd.UserID, cmd.Items)
	order, err := domain.NewOrder(id, cmd.UserID, cmd.Items)
	if err != nil {
		return nil, fmt.Errorf("can't create order: %w", err)
	}
	// TODO: Store order in h.orders
	h.orders[id] = order
	// TODO: Publish uncommitted events and clear them
	for _, ev := range order.UncommittedEvents() {
		h.bus.Publish(ev)
	}
	order.ClearEvents()
	// TODO: Return order
	return order, nil
}

// HandleCancelOrder executes a CancelOrderCommand:
//  1. Look up the order — return error if not found
//  2. Call order.Cancel(cmd.Reason)
//  3. Publish uncommitted events and clear them
func (h *OrderCommandHandler) HandleCancelOrder(cmd domain.CancelOrderCommand) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// TODO: Look up order in h.orders
	order, ok := h.orders[cmd.OrderID]
	if !ok {
		return fmt.Errorf("Order %s not found for cancellation", cmd.OrderID)
	}
	// TODO: Call order.Cancel(cmd.Reason)
	order.Cancel(cmd.Reason)
	// TODO: Publish uncommitted events and clear them
	for _, ev := range order.UncommittedEvents() {
		h.bus.Publish(ev)
	}
	order.ClearEvents()
	return nil
}

// HandleShipOrder executes a ShipOrderCommand:
//  1. Look up the order — return error if not found
//  2. Call order.Ship(cmd.TrackingNumber)
//  3. Publish uncommitted events and clear them
func (h *OrderCommandHandler) HandleShipOrder(cmd domain.ShipOrderCommand) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// TODO: Look up order in h.orders
	order, ok := h.orders[cmd.OrderID]
	if !ok {
		return fmt.Errorf("Order %s not found for shipping\n", cmd.OrderID)
	}
	// TODO: Call order.Ship(cmd.TrackingNumber)
	order.Ship(cmd.TrackingNumber)
	// TODO: Publish uncommitted events and clear them
	for _, ev := range order.UncommittedEvents() {
		h.bus.Publish(ev)
	}
	order.ClearEvents()
	return nil
}
