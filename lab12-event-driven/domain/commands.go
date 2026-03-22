package domain

import "lab12/events"

// PlaceOrderCommand is issued when a user wants to place a new order.
type PlaceOrderCommand struct {
	UserID string
	Items  []events.OrderItem
}

// CancelOrderCommand is issued when a user wants to cancel an existing order.
type CancelOrderCommand struct {
	OrderID string
	Reason  string
}

// ShipOrderCommand is issued when an order is ready to be shipped.
type ShipOrderCommand struct {
	OrderID        string
	TrackingNumber string
}
