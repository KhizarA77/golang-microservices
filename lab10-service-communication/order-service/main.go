package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// =============================================================================
// Models
// =============================================================================

// User mirrors the user-service User struct.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Order represents a purchase order.
type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// EnrichedOrder includes the full User object.
type EnrichedOrder struct {
	Order
	User *User `json:"user,omitempty"`
}

// =============================================================================
// User Service Client (Task 2)
// =============================================================================

// UserServiceClient is a typed HTTP client for the user service.
type UserServiceClient struct {
	baseURL string
	client  *http.Client
	cb      *CircuitBreaker
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 3 * time.Second},
		cb:      NewCircuitBreaker(3, 10*time.Second),
	}
}

// GetUser fetches a single user from the user service.
func (c *UserServiceClient) GetUser(ctx context.Context, id string) (*User, error) {
	// TODO: Check circuit breaker: if !c.cb.Allow(), return error "circuit open"
	if !c.cb.Allow() {
		return nil, fmt.Errorf("Circuit open")
	}
	// TODO: Create request: http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users/"+id, nil)
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users/"+id, nil)
	if err != nil {
		return nil, err
	}
	// TODO: Execute with retry: use retryGet
	// TODO: If error, call c.cb.Failure(), return error
	resp, err := retryGet(ctx, c.client, req.URL.String(), 3)
	if err != nil {
		c.cb.Failure()
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: Decode response body into User
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.cb.Failure()
		return nil, err
	}

	// TODO: Call c.cb.Success()
	c.cb.Success()
	// TODO: Return user
	return &user, nil
}

func (c *UserServiceClient) ListUsers(ctx context.Context) ([]*User, error) {
	if !c.cb.Allow() {
		return nil, fmt.Errorf("Circuit open")
	}
	// TODO: Create request: http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users", nil)
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users", nil)
	if err != nil {
		return nil, err
	}
	// TODO: Execute with retry: use retryGet
	// TODO: If error, call c.cb.Failure(), return error
	resp, err := retryGet(ctx, c.client, req.URL.String(), 3)
	if err != nil {
		c.cb.Failure()
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: Decode response body into User
	var users []*User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		c.cb.Failure()
		return nil, err
	}

	// TODO: Call c.cb.Success()
	c.cb.Success()
	// TODO: Return user
	return users, nil
}

// IsHealthy checks if the user service is healthy.
func (c *UserServiceClient) IsHealthy(ctx context.Context) bool {
	// TODO: GET c.baseURL + "/health" with 1-second timeout
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	// TODO: Return true if status 200
	res, err := retryGet(ctx, c.client, req.URL.String(), 3)
	if err != nil {
		c.cb.Failure()
		return false
	}
	if res.StatusCode == 200 {
		c.cb.Success()
		return true
	}
	c.cb.Failure()
	return false
}

// retryGet executes a GET request with exponential backoff retry.
// Retries on network errors and 5xx responses.
func retryGet(ctx context.Context, client *http.Client, url string, maxRetries int) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// TODO: Exponential backoff: 100ms, 200ms, 400ms, ...
			// TODO: Select on time.After and ctx.Done() so we respect cancellation
			backoff := time.Duration(1<<attempt) * 100 * time.Millisecond
			log.Printf("[RETRY] attempt %d, waiting %v", attempt, backoff)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("[RETRY] network error: %v", err)
			continue
		}

		// TODO: If response is 5xx, close body and retry
		// TODO: Otherwise return the response
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			log.Printf("[RETRY] got status %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// =============================================================================
// Circuit Breaker (Task 5)
// =============================================================================

// CircuitBreaker implements a simple circuit breaker.
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailure  time.Time
	state        string // "closed", "open", "half-open"
	mu           sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

// Allow returns true if a request should be attempted.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case "closed":
		return true
	case "open":
		// TODO: If resetTimeout has passed since lastFailure, transition to half-open and return true
		// TODO: Otherwise return false
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = "half-open"
			log.Println("[CB] transitioning to half-open")
			return true
		}
		return false
	case "half-open":
		return true
	}
	return false
}

// Success records a successful request.
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	// TODO: Reset failures to 0
	// TODO: Set state to "closed"
	cb.failures = 0
	cb.state = "closed"
	log.Println("[CB] success, circuit closed")
}

// Failure records a failed request.
func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	// TODO: Increment failures
	// TODO: Set lastFailure = time.Now()
	// TODO: If failures >= maxFailures, set state = "open"
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.maxFailures {
		cb.state = "open"
		log.Printf("[CB] circuit opened after %d failures", cb.failures)
	}
}

// =============================================================================
// Order Store
// =============================================================================

type OrderStore struct {
	mu     sync.RWMutex
	orders map[string]*Order
	nextID int
}

func NewOrderStore() *OrderStore {
	return &OrderStore{orders: make(map[string]*Order)}
}

func (s *OrderStore) create(o Order) *Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	o.ID = fmt.Sprintf("%d", s.nextID)
	o.CreatedAt = time.Now()
	o.Status = "pending"
	cp := o
	s.orders[o.ID] = &cp
	return &cp
}

func (s *OrderStore) list() []*Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Order, 0, len(s.orders))
	for _, o := range s.orders {
		cp := *o
		result = append(result, &cp)
	}
	return result
}

func (s *OrderStore) getByID(id string) (*Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, false
	}
	cp := *o
	return &cp, true
}

// =============================================================================
// Handlers
// =============================================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

var (
	orderStore *OrderStore
	userClient *UserServiceClient
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	userServiceOK := userClient.IsHealthy(ctx)

	status := "ok"
	if !userServiceOK {
		status = "degraded"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  status,
		"service": "order-service",
		"dependencies": map[string]string{
			"user-service": map[bool]string{true: "ok", false: "unavailable"}[userServiceOK],
		},
	})
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID      string  `json:"userId"`
		ProductName string  `json:"productName"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// TODO: Validate user exists by calling userClient.GetUser(ctx, input.UserID)
	_, err := userClient.GetUser(ctx, input.UserID)
	// TODO: If error (user not found or service unavailable), return 503 or 404
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "User not found"})
		return
	}
	// TODO: Create order in orderStore
	order := orderStore.create(Order{UserID: input.UserID, ProductName: input.ProductName,
		Quantity: input.Quantity, Price: input.Price})
	// TODO: Return 201 with order
	writeJSON(w, 201, order)
}

func handleListOrders(w http.ResponseWriter, r *http.Request) {
	orders := orderStore.list()

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// TODO: For each order, try to enrich with user data
	// TODO: If user service fails, still return order but with user = nil
	// Build enriched orders
	enriched := make([]*EnrichedOrder, 0, len(orders))
	for _, o := range orders {
		e := &EnrichedOrder{Order: *o}
		user, err := userClient.GetUser(ctx, o.UserID)
		if err == nil {
			e.User = user
		}
		enriched = append(enriched, e)
	}

	writeJSON(w, http.StatusOK, enriched)
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	order, ok := orderStore.getByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// TODO: Enrich with user data
	enriched := &EnrichedOrder{Order: *order}
	user, err := userClient.GetUser(ctx, order.UserID)
	if err == nil {
		enriched.User = user
	}

	writeJSON(w, http.StatusOK, enriched)
}

func main() {
	userServiceURL := "http://localhost:8081"

	orderStore = NewOrderStore()
	userClient = NewUserServiceClient(userServiceURL)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /orders", handleCreateOrder)
	mux.HandleFunc("GET /orders", handleListOrders)
	mux.HandleFunc("GET /orders/{id}", handleGetOrder)

	fmt.Println("Order Service running on http://localhost:8082")
	fmt.Printf("Connecting to user-service at %s\n", userServiceURL)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ok := userClient.IsHealthy(ctx)
	if !ok {
		log.Fatalf("Failed to connect to user-service at %s\n", userServiceURL)
	}
	fmt.Println("Successfully connected to user service")
	log.Fatal(http.ListenAndServe(":8082", mux))
}
