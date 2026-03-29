package usecase

import "order-service/domain"

type OrderRepository interface {
	FindByID(id string) (*domain.Order, error)
	Save(order *domain.Order) error
	Update(order *domain.Order) error
	List() ([]*domain.Order, error)
}
