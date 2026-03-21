package main

// Product Service — Port 8082
//
// INSTRUCTIONS:
// Port your Lab 08 REST API solution here.
// Additions for the capstone:
//   1. Caching proxy (Lab 05): cache individual product responses for 30s
//   2. Concurrent search (Lab 03): SearchProducts fans out to 3 search "strategies"
//   3. Stock management: products have a stock count that decreases when ordered
//
// Endpoints:
//   GET    /health
//   GET    /api/products              list (paginated)
//   GET    /api/products/search?q=... concurrent search
//   GET    /api/products/{id}         get by ID (cached)
//   POST   /api/products              create
//   PATCH  /api/products/{id}         update
//   DELETE /api/products/{id}         delete
//   PATCH  /api/products/{id}/stock   { "delta": -1 } reserve or release stock
//
// The /api/products/{id}/stock endpoint is used by the Order Service
// to reserve stock when an order is placed.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Product represents a catalog product.
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// =============================================================================
// In-memory store
// =============================================================================

type productStore struct {
	mu       sync.RWMutex
	products map[string]*Product
	nextID   int
}

var store = &productStore{products: make(map[string]*Product)}

func (s *productStore) seed() {
	seeds := []Product{
		{Name: "MacBook Pro", Description: "Apple laptop", Price: 1999.99, Stock: 10, Category: "electronics"},
		{Name: "Go Programming", Description: "Learn Go", Price: 39.99, Stock: 100, Category: "books"},
		{Name: "Winter Jacket", Description: "Warm jacket", Price: 89.99, Stock: 30, Category: "clothing"},
		{Name: "Organic Coffee", Description: "Single origin", Price: 14.99, Stock: 200, Category: "food"},
		{Name: "Wireless Mouse", Description: "Ergonomic mouse", Price: 49.99, Stock: 50, Category: "electronics"},
		{Name: "Python Cookbook", Description: "Python recipes", Price: 45.99, Stock: 75, Category: "books"},
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range seeds {
		s.nextID++
		p.ID = fmt.Sprintf("%d", s.nextID)
		p.CreatedAt = time.Now()
		cp := p
		s.products[p.ID] = &cp
	}
}

func (s *productStore) getByID(id string) (*Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	if !ok {
		return nil, false
	}
	cp := *p
	return &cp, true
}

func (s *productStore) list() []*Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Product, 0, len(s.products))
	for _, p := range s.products {
		cp := *p
		result = append(result, &cp)
	}
	return result
}

// reserveStock decrements stock by the given amount.
// Returns error if insufficient stock.
func (s *productStore) reserveStock(id string, qty int) (*Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.products[id]
	if !ok {
		return nil, fmt.Errorf("product not found")
	}
	if p.Stock < qty {
		return nil, fmt.Errorf("insufficient stock: have %d, need %d", p.Stock, qty)
	}
	p.Stock -= qty
	cp := *p
	return &cp, nil
}

// =============================================================================
// Caching proxy (Lab 05 Proxy pattern)
// =============================================================================

type cacheEntry struct {
	product  *Product
	cachedAt time.Time
}

type productCache struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry
	ttl   time.Duration
}

var cache = &productCache{
	cache: make(map[string]cacheEntry),
	ttl:   30 * time.Second,
}

func (c *productCache) get(id string) (*Product, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.cache[id]
	if !ok || time.Since(entry.cachedAt) > c.ttl {
		return nil, false
	}
	return entry.product, true
}

func (c *productCache) set(id string, p *Product) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = cacheEntry{product: p, cachedAt: time.Now()}
}

func (c *productCache) invalidate(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, id)
}

// =============================================================================
// Concurrent Search (Lab 03 Fan-Out/Fan-In)
// =============================================================================

// searchByName searches products where name contains the query.
func searchByName(ctx context.Context, products []*Product, query string) <-chan *Product {
	out := make(chan *Product, 10)
	go func() {
		defer close(out)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(query)) {
				out <- p
			}
		}
	}()
	return out
}

// searchByCategory searches products where category contains the query.
func searchByCategory(ctx context.Context, products []*Product, query string) <-chan *Product {
	out := make(chan *Product, 10)
	go func() {
		defer close(out)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if strings.Contains(strings.ToLower(p.Category), strings.ToLower(query)) {
				out <- p
			}
		}
	}()
	return out
}

// searchByDescription searches products where description contains the query.
func searchByDescription(ctx context.Context, products []*Product, query string) <-chan *Product {
	out := make(chan *Product, 10)
	go func() {
		defer close(out)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if strings.Contains(strings.ToLower(p.Description), strings.ToLower(query)) {
				out <- p
			}
		}
	}()
	return out
}

// TODO: Implement fanIn that merges multiple product channels into one
// (use sync.WaitGroup like in Lab 03)
func fanIn(ctx context.Context, channels ...<-chan *Product) <-chan *Product {
	merged := make(chan *Product, 30)
	var wg sync.WaitGroup

	forward := func(ch <-chan *Product) {
		defer wg.Done()
		for p := range ch {
			select {
			case merged <- p:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()
	return merged
}

// searchProducts fans out the search across multiple strategies and deduplicates results.
func searchProducts(ctx context.Context, query string) []*Product {
	all := store.list()

	// Fan out to 3 concurrent search strategies
	ch1 := searchByName(ctx, all, query)
	ch2 := searchByCategory(ctx, all, query)
	ch3 := searchByDescription(ctx, all, query)

	// Fan in results
	merged := fanIn(ctx, ch1, ch2, ch3)

	// Deduplicate by ID
	seen := make(map[string]bool)
	var results []*Product
	for p := range merged {
		if !seen[p.ID] {
			seen[p.ID] = true
			results = append(results, p)
		}
	}
	return results
}

// =============================================================================
// Handlers
// =============================================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok", "service": "product-service"})
}

func handleListProducts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, store.list())
}

func handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Check cache first
	if p, ok := cache.get(id); ok {
		log.Printf("[CACHE HIT] product %s", id)
		writeJSON(w, 200, p)
		return
	}

	log.Printf("[CACHE MISS] product %s", id)
	p, ok := store.getByID(id)
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "product not found"})
		return
	}

	cache.set(id, p)
	writeJSON(w, 200, p)
}

func handleSearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, 400, map[string]string{"error": "q parameter required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	results := searchProducts(ctx, query)
	if results == nil {
		results = []*Product{}
	}
	writeJSON(w, 200, results)
}

func handleReserveStock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Delta int `json:"delta"` // negative to reserve, positive to release
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	qty := -req.Delta // delta is negative for reservation, so negate for reserveStock
	if req.Delta < 0 {
		// Reserving stock
		p, err := store.reserveStock(id, qty)
		if err != nil {
			writeJSON(w, 409, map[string]string{"error": err.Error()})
			return
		}
		cache.invalidate(id)
		writeJSON(w, 200, p)
	} else {
		// TODO: Releasing stock (return items) — implement similarly
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}
}

func main() {
	store.seed()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/products", handleListProducts)
	mux.HandleFunc("GET /api/products/search", handleSearchProducts)
	mux.HandleFunc("GET /api/products/{id}", handleGetProduct)
	mux.HandleFunc("PATCH /api/products/{id}/stock", handleReserveStock)
	// TODO: Add POST, PATCH (full update), DELETE handlers

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Product Service on http://localhost:8082")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Product Service shutting down...")
}
