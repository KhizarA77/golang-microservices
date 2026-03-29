package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ErrProductNotFound is returned when the product service responds with 404.
var ErrProductNotFound = errors.New("product not found")

type ProductServiceClient struct {
	baseURL string
	client  *http.Client
	cb      *CircuitBreaker
}

type ProductInfo struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func NewProductServiceClient(baseURL string) *ProductServiceClient {
	return &ProductServiceClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 3 * time.Second},
		cb:      NewCircuitBreaker(3, 10*time.Second),
	}
}

func (p *ProductServiceClient) ValidateAndReserveProduct(ctx context.Context, productID string, qty int) (*ProductInfo, error) {
	if !p.cb.Allow() {
		return nil, fmt.Errorf("circuit open: product service unavailable")
	}

	// GET product info — no wasteful request object, retryGet builds its own
	url := fmt.Sprintf("%s/api/products/%s", p.baseURL, productID)
	resp, err := retryGet(ctx, p.client, url, 3)
	if err != nil {
		p.cb.Failure()
		return nil, fmt.Errorf("product service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Service is up, product just doesn't exist — not a circuit-breaker failure
		return nil, ErrProductNotFound
	}
	if resp.StatusCode != 200 {
		p.cb.Failure()
		return nil, fmt.Errorf("product service error: %d", resp.StatusCode)
	}

	var product ProductInfo
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		p.cb.Failure()
		return nil, err
	}

	if product.Stock < qty {
		// Business rule failure — service is healthy
		return nil, fmt.Errorf("insufficient stock: have %d, need %d", product.Stock, qty)
	}

	// PATCH to reserve stock — also tracked by the circuit breaker
	body, _ := json.Marshal(map[string]int{"delta": -qty})
	stockReq, err := http.NewRequestWithContext(ctx, "PATCH",
		fmt.Sprintf("%s/api/products/%s/stock", p.baseURL, productID),
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	stockReq.Header.Set("Content-Type", "application/json")

	stockResp, err := p.client.Do(stockReq)
	if err != nil {
		p.cb.Failure()
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}
	defer stockResp.Body.Close()

	if stockResp.StatusCode == 409 {
		// Concurrent reservation conflict — service is healthy
		return nil, fmt.Errorf("insufficient stock (concurrent reservation)")
	}
	if stockResp.StatusCode != 200 && stockResp.StatusCode != 204 {
		p.cb.Failure()
		return nil, fmt.Errorf("stock reservation failed: %d", stockResp.StatusCode)
	}

	// Only mark success after the full operation completes cleanly
	p.cb.Success()
	return &product, nil
}

func retryGet(ctx context.Context, client *http.Client, url string, maxRetries int) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
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
