package domain

import (
	"encoding/json"
	"time"
)

type Order struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	ProductID   string     `json:"product_id"`
	ProductName string     `json:"product_name"`
	Quantity    int        `json:"quantity"`
	TotalPrice  float64    `json:"total_price"`
	Status      OrderState `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MarshalJSON serializes Status as its string name instead of the raw interface struct.
func (o *Order) MarshalJSON() ([]byte, error) {
	status := ""
	if o.Status != nil {
		status = o.Status.Name()
	}
	return json.Marshal(struct {
		ID          string    `json:"id"`
		UserID      string    `json:"user_id"`
		ProductID   string    `json:"product_id"`
		ProductName string    `json:"product_name"`
		Quantity    int       `json:"quantity"`
		TotalPrice  float64   `json:"total_price"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}{
		ID:          o.ID,
		UserID:      o.UserID,
		ProductID:   o.ProductID,
		ProductName: o.ProductName,
		Quantity:    o.Quantity,
		TotalPrice:  o.TotalPrice,
		Status:      status,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	})
}

func (o *Order) Next() error {
	return o.Status.Next(o)
}

func (o *Order) Cancel() error {
	return o.Status.Cancel(o)
}
