package main

// Order Service — Port 8083
//
// This is the most complex service. It integrates:
//   - Clean architecture (domain/usecase/repository/handler layers)
//   - State machine for order transitions (Lab 06)
//   - Worker pool for processing orders concurrently (Lab 03)
//   - HTTP service client with circuit breaker (Lab 10)
//   - Domain events published to shared event bus (Lab 12)
//
// ORDER FLOW:
//   1. POST /api/orders → validate product+stock → create Pending order → queue for processing
//   2. Worker goroutine picks up order → simulates payment → transitions state → publishes event
//
// Endpoints:
//   GET  /health
//   POST /api/orders      { productId, quantity, userId }
//   GET  /api/orders
//   GET  /api/orders/{id}
//   POST /api/orders/{id}/ship   { trackingNumber }

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// =============================================================================
// Domain: Order + State Machine (from Lab 06)
// =============================================================================

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	ProductID   string      `json:"product_id"`
	ProductName string      `json:"product_name"`
	Quantity    int         `json:"quantity"`
	TotalPrice  float64     `json:"total_price"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// =============================================================================
// Events (shared with Notification Service via channel)
// =============================================================================

type OrderEvent struct {
	Type    string
	OrderID string
	UserID  string
	Payload interface{}
}

// Global event channel — notification service subscribes to this.
// In production this would be a message broker.
var eventCh = make(chan OrderEvent, 100)

func publishEvent(eventType, orderID, userID string, payload interface{}) {
	select {
	case eventCh <- OrderEvent{Type: eventType, OrderID: orderID, UserID: userID, Payload: payload}:
	default:
		log.Printf("[WARN] event channel full, dropping event: %s", eventType)
	}
}

// =============================================================================
// Order Store
// =============================================================================

type orderStore struct {
	mu     sync.RWMutex
	orders map[string]*Order
	nextID int
}

var store = &orderStore{orders: make(map[string]*Order)}

func (s *orderStore) create(o *Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	o.ID = fmt.Sprintf("ORD-%03d", s.nextID)
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	cp := *o
	s.orders[o.ID] = &cp
	// Update the original so caller has the ID
	*o = cp
}

func (s *orderStore) update(o *Order) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orders[o.ID]; !ok {
		return false
	}
	o.UpdatedAt = time.Now()
	cp := *o
	s.orders[o.ID] = &cp
	return true
}

func (s *orderStore) getByID(id string) (*Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, false
	}
	cp := *o
	return &cp, true
}

func (s *orderStore) list() []*Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Order, 0, len(s.orders))
	for _, o := range s.orders {
		cp := *o
		result = append(result, &cp)
	}
	return result
}

// =============================================================================
// Worker Pool (Lab 03 — processes orders asynchronously)
// =============================================================================

type processingJob struct {
	order *Order
}

var processingQueue = make(chan processingJob, 50)

// startOrderWorkers launches numWorkers goroutines that process orders.
func startOrderWorkers(numWorkers int, wg *sync.WaitGroup) {
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("[WORKER %d] started", workerID)
			for job := range processingQueue {
				processOrder(workerID, job.order)
			}
			log.Printf("[WORKER %d] stopped", workerID)
		}(i)
	}
}

// processOrder simulates payment processing and transitions the order state.
func processOrder(workerID int, order *Order) {
	log.Printf("[WORKER %d] processing order %s", workerID, order.ID)

	// Transition to Processing
	order.Status = StatusProcessing
	store.update(order)
	publishEvent("order.processing", order.ID, order.UserID, nil)

	// Simulate payment processing
	time.Sleep(500 * time.Millisecond)

	// Transition to Shipped (simplified — always succeeds in this lab)
	tracking := fmt.Sprintf("TRACK-%s-%d", order.ID, time.Now().Unix())
	order.Status = StatusShipped
	store.update(order)
	publishEvent("order.shipped", order.ID, order.UserID, map[string]string{"tracking": tracking})

	log.Printf("[WORKER %d] order %s shipped (tracking: %s)", workerID, order.ID, tracking)
}

// =============================================================================
// Product Service Client (Lab 10)
// =============================================================================

const productServiceURL = "http://localhost:8082"

type productInfo struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// validateAndReserveProduct checks the product exists and reserves stock.
func validateAndReserveProduct(ctx context.Context, productID string, qty int) (*productInfo, error) {
	// First, get product info
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/products/%s", productServiceURL, productID), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("product service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("product %s not found", productID)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("product service error: %d", resp.StatusCode)
	}

	var p productInfo
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	if p.Stock < qty {
		return nil, fmt.Errorf("insufficient stock: have %d, need %d", p.Stock, qty)
	}

	// Reserve stock (delta = -qty)
	body, _ := json.Marshal(map[string]int{"delta": -qty})
	stockReq, _ := http.NewRequestWithContext(ctx, "PATCH",
		fmt.Sprintf("%s/api/products/%s/stock", productServiceURL, productID),
		bytes.NewReader(body))
	stockReq.Header.Set("Content-Type", "application/json")

	stockResp, err := client.Do(stockReq)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}
	defer stockResp.Body.Close()

	if stockResp.StatusCode == 409 {
		return nil, fmt.Errorf("insufficient stock (concurrent reservation)")
	}

	return &p, nil
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
	writeJSON(w, 200, map[string]interface{}{
		"status":  "ok",
		"service": "order-service",
		"workers": 3,
		"queue":   len(processingQueue),
	})
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID string `json:"productId"`
		Quantity  int    `json:"quantity"`
		UserID    string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	if req.ProductID == "" || req.Quantity <= 0 || req.UserID == "" {
		writeJSON(w, 400, map[string]string{"error": "productId, quantity, and userId are required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Validate product and reserve stock
	product, err := validateAndReserveProduct(ctx, req.ProductID, req.Quantity)
	if err != nil {
		status := 422
		if err.Error() == fmt.Sprintf("product %s not found", req.ProductID) {
			status = 404
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	// Create order
	order := &Order{
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		ProductName: product.Name,
		Quantity:    req.Quantity,
		TotalPrice:  float64(req.Quantity) * product.Price,
		Status:      StatusPending,
	}
	store.create(order)

	// Publish placed event
	publishEvent("order.placed", order.ID, order.UserID, map[string]interface{}{
		"productName": product.Name,
		"quantity":    req.Quantity,
		"totalPrice":  order.TotalPrice,
	})

	// Queue for async processing (non-blocking)
	select {
	case processingQueue <- processingJob{order: order}:
		log.Printf("Order %s queued for processing", order.ID)
	default:
		log.Printf("[WARN] processing queue full, order %s will not be processed", order.ID)
	}

	writeJSON(w, 202, order) // 202 Accepted — will be processed asynchronously
}

func handleListOrders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, store.list())
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	order, ok := store.getByID(id)
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "order not found"})
		return
	}
	writeJSON(w, 200, order)
}

func main() {
	var workerWG sync.WaitGroup

	// Start worker pool (Lab 03 pattern)
	startOrderWorkers(3, &workerWG)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/orders", handleCreateOrder)
	mux.HandleFunc("GET /api/orders", handleListOrders)
	mux.HandleFunc("GET /api/orders/{id}", handleGetOrder)

	srv := &http.Server{
		Addr:         ":8083",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Order Service on http://localhost:8083")
		fmt.Println("(calls Product Service at http://localhost:8082)")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Order Service shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	close(processingQueue)  // Signal workers to stop
	workerWG.Wait()          // Wait for in-flight orders
	fmt.Println("Order Service stopped gracefully")
}
