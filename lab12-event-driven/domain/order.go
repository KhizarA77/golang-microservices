package domain

import (
	"errors"
	"fmt"
	"lab12/eventbus"
	"lab12/events"
	"time"
)

// Order is the order aggregate.
type Order struct {
	ID         string
	UserID     string
	Items      []events.OrderItem
	Status     string // "pending", "shipped", "cancelled"
	TotalPrice float64
	CreatedAt  time.Time
	UpdatedAt  time.Time

	// Uncommitted domain events — published by the command handler after saving.
	uncommittedEvents []eventbus.Event
	version           int
}

// NewOrder creates a new empty order with the given ID.
func NewOrder(id, userID string, items []events.OrderItem) (*Order, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	total := 0.0
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("item %s has invalid quantity", item.ProductID)
		}
		total += float64(item.Quantity) * item.UnitPrice
	}

	o := &Order{
		ID:         id,
		UserID:     userID,
		Items:      items,
		Status:     "pending",
		TotalPrice: total,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// TODO: Record OrderPlaced domain event
	// o.recordEvent(events.OrderPlaced, events.OrderPlacedPayload{...})

	return o, nil
}

// Cancel cancels the order with a reason.
func (o *Order) Cancel(reason string) error {
	if o.Status == "cancelled" {
		return errors.New("order already cancelled")
	}
	if o.Status == "shipped" {
		return errors.New("cannot cancel a shipped order")
	}
	// TODO: Set status to "cancelled", update UpdatedAt
	// TODO: Record OrderCancelled event
	return nil
}

// Ship marks the order as shipped.
func (o *Order) Ship(trackingNumber string) error {
	if o.Status != "pending" {
		return fmt.Errorf("cannot ship order in status %q", o.Status)
	}
	// TODO: Set status to "shipped", update UpdatedAt
	// TODO: Record OrderShipped event
	_ = trackingNumber
	return nil
}

// UncommittedEvents returns all domain events not yet published.
func (o *Order) UncommittedEvents() []eventbus.Event {
	return o.uncommittedEvents
}

// ClearEvents clears all uncommitted events (call after publishing).
func (o *Order) ClearEvents() {
	o.uncommittedEvents = nil
}

// recordEvent adds an event to the uncommitted events list.
func (o *Order) recordEvent(eventType string, payload interface{}) {
	o.version++
	o.uncommittedEvents = append(o.uncommittedEvents, eventbus.Event{
		Type:        eventType,
		AggregateID: o.ID,
		Payload:     payload,
		OccurredAt:  time.Now(),
		Version:     o.version,
	})
}
