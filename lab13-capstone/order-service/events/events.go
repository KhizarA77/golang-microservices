package events

type OrderEvent struct {
	Type    string
	OrderID string
	UserID  string
	Payload interface{}
}
