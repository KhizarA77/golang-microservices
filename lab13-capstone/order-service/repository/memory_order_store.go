package repository

import (
	"fmt"
	"order-service/domain"
	"sync"
)

type MemoryOrderStore struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
	nextID int
}

func NewMemoryOrderStore() *MemoryOrderStore {
	return &MemoryOrderStore{
		orders: make(map[string]*domain.Order),
	}
}

func (s *MemoryOrderStore) Save(order *domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	order.ID = fmt.Sprintf("ORD-%03d", s.nextID)
	cp := *order
	s.orders[order.ID] = &cp
	return nil
}

func (s *MemoryOrderStore) FindByID(id string) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, fmt.Errorf("order %s not found", id)
	}
	cp := *o
	return &cp, nil
}

func (s *MemoryOrderStore) Update(order *domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orders[order.ID]; !ok {
		return fmt.Errorf("order %s not found", order.ID)
	}
	cp := *order
	s.orders[order.ID] = &cp
	return nil
}

func (s *MemoryOrderStore) List() ([]*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*domain.Order, 0, len(s.orders))
	for _, o := range s.orders {
		cp := *o
		result = append(result, &cp)
	}
	return result, nil
}
