package events

// Event type constants — use these as Event.Type
const (
	OrderPlaced            = "order.placed"
	OrderCancelled         = "order.cancelled"
	OrderShipped           = "order.shipped"
	InventoryReserved      = "inventory.reserved"
	InventoryReleaseFailed = "inventory.release_failed"
)

// OrderItem represents a line item in an order.
type OrderItem struct {
	ProductID   string
	ProductName string
	Quantity    int
	UnitPrice   float64
}

// OrderPlacedPayload is the payload for order.placed events.
type OrderPlacedPayload struct {
	OrderID    string
	UserID     string
	Items      []OrderItem
	TotalPrice float64
}

// OrderCancelledPayload is the payload for order.cancelled events.
type OrderCancelledPayload struct {
	OrderID string
	Reason  string
}

// OrderShippedPayload is the payload for order.shipped events.
type OrderShippedPayload struct {
	OrderID        string
	TrackingNumber string
}

// InventoryReservedPayload is the payload for inventory.reserved events.
type InventoryReservedPayload struct {
	OrderID string
	Items   []OrderItem
}

// InventoryReleaseFailedPayload is the payload for inventory.release_failed events.
type InventoryReleaseFailedPayload struct {
	OrderID string
	Reason  string
}
