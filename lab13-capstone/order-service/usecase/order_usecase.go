package usecase

import (
	"fmt"
	"order-service/domain"
	"time"
)

type CreateOrderInput struct {
	UserID      string
	ProductID   string
	ProductName string
	Quantity    int
	TotalPrice  float64
}

type OrderUseCase struct {
	repo OrderRepository
}

func NewOrderUseCase(repo OrderRepository) *OrderUseCase {
	return &OrderUseCase{repo: repo}
}

// PlaceOrder creates a new order in Pending state and persists it.
func (uc *OrderUseCase) PlaceOrder(input CreateOrderInput) (*domain.Order, error) {
	if input.UserID == "" || input.ProductID == "" || input.Quantity <= 0 {
		return nil, fmt.Errorf("userID, productID, and quantity are required")
	}
	order := &domain.Order{
		UserID:      input.UserID,
		ProductID:   input.ProductID,
		ProductName: input.ProductName,
		Quantity:    input.Quantity,
		TotalPrice:  input.TotalPrice,
		Status:      &domain.PendingState{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := uc.repo.Save(order); err != nil {
		return nil, fmt.Errorf("saving order: %w", err)
	}
	return order, nil
}

// GetOrder retrieves an order by ID.
func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.FindByID(id)
}

// ListOrders returns all orders.
func (uc *OrderUseCase) ListOrders() ([]*domain.Order, error) {
	return uc.repo.List()
}

// AdvanceOrder transitions the order to the next state (Pending→Processing→Shipped→Delivered).
func (uc *OrderUseCase) AdvanceOrder(id string) error {
	order, err := uc.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := order.Next(); err != nil {
		return err
	}
	order.UpdatedAt = time.Now()
	return uc.repo.Update(order)
}

// CancelOrder cancels an order if its current state allows it.
func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := order.Cancel(); err != nil {
		return err
	}
	order.UpdatedAt = time.Now()
	return uc.repo.Update(order)
}
