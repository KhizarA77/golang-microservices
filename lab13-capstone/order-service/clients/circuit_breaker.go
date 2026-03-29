package clients

import (
	"log"
	"sync"
	"time"
)

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
