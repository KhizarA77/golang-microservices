package store

import (
	"lab08/models"
	"sort"
	"sync"
	"time"
)

// ProductStore is a thread-safe in-memory store for products.
type ProductStore struct {
	mu       sync.RWMutex
	products map[int]*models.Product
	nextID   int
}

// NewProductStore creates a new store pre-seeded with sample products.
func NewProductStore() *ProductStore {
	s := &ProductStore{
		products: make(map[int]*models.Product),
	}

	// Seed with 5 sample products
	seeds := []models.CreateProductRequest{
		{Name: "MacBook Pro", Description: "Apple laptop", Price: 1999.99, Stock: 5, Category: "electronics"},
		{Name: "Go Programming Book", Description: "Learn Go from scratch", Price: 39.99, Stock: 100, Category: "books"},
		{Name: "Winter Jacket", Description: "Warm for cold climates", Price: 89.99, Stock: 30, Category: "clothing"},
		{Name: "Organic Coffee", Description: "Single origin beans", Price: 14.99, Stock: 200, Category: "food"},
		{Name: "Wireless Headphones", Description: "Noise cancelling", Price: 299.99, Stock: 15, Category: "electronics"},
	}

	for _, seed := range seeds {
		s.Create(seed)
	}

	return s
}

// GetAll returns a paginated slice of all products and the total count.
func (s *ProductStore) GetAll(page, limit int) ([]*models.Product, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: Collect all products into a slice
	sli := make([]*models.Product, 0)
	for _, val := range s.products {
		sli = append(sli, val)
	}
	sort.Slice(sli, func(i, j int) bool { return sli[i].ID < sli[j].ID })
	// TODO: Apply pagination: start = (page-1)*limit, end = min(start+limit, len)
	start := (page - 1) * limit
	if start >= len(sli) {
		return []*models.Product{}, len(sli)
	}
	end := min(start+limit, len(sli))
	// TODO: Return the page slice and total count
	return sli[start:end], len(sli)
}

// GetByCategory returns products filtered by category with pagination.
func (s *ProductStore) GetByCategory(category string, page, limit int) ([]*models.Product, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: Filter products where product.Category == category
	sli := make([]*models.Product, 0)
	for _, val := range s.products {
		if val.Category == category {
			sli = append(sli, val)
		}
	}
	sort.Slice(sli, func(i, j int) bool { return sli[i].ID < sli[j].ID })
	// TODO: Apply pagination
	start := (page - 1) * limit
	if start >= len(sli) {
		return []*models.Product{}, len(sli)
	}
	end := min(start+limit, len(sli))
	// TODO: Return page slice and total count for that category
	return sli[start:end], len(sli)
}

// GetByID returns a product by its ID.
func (s *ProductStore) GetByID(id int) (*models.Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res, ok := s.products[id]
	// TODO: Return s.products[id], ok
	if !ok {
		return nil, false
	}
	return res, true
}

// Create creates a new product from the given request.
func (s *ProductStore) Create(req models.CreateProductRequest) *models.Product {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	p := &models.Product{
		// TODO: Set all fields from req + ID, CreatedAt, UpdatedAt
		ID:          s.nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// TODO: Store in s.products[p.ID]
	s.products[p.ID] = p
	// TODO: Return p
	return p
}

// Update applies a partial update to a product.
// Returns the updated product and true, or nil and false if not found.
func (s *ProductStore) Update(id int, req models.UpdateProductRequest) (*models.Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.products[id]
	if !ok {
		return nil, false
	}

	// TODO: Update only non-nil fields from req:
	//   if req.Name != nil { p.Name = *req.Name }
	//   etc.
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Category != nil {
		p.Category = *req.Category
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}
	// TODO: Update p.UpdatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return p, true
}

// Delete removes a product by ID.
// Returns true if deleted, false if not found.
func (s *ProductStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Check if product exists
	_, ok := s.products[id]
	// TODO: Delete from s.products
	if !ok {
		return false
	}
	delete(s.products, id)
	// TODO: Return true/false
	return true
}
