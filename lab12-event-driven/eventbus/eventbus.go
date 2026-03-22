package eventbus

import (
	"sync"
	"time"
)

// Event is the base event type passed to all handlers.
type Event struct {
	Type        string
	AggregateID string
	Payload     interface{}
	OccurredAt  time.Time
	Version     int
}

// EventBus dispatches events to registered subscribers.
type EventBus struct {
	subscribers map[string][]func(Event)
	mu          sync.RWMutex
	wg          sync.WaitGroup // tracks in-flight async handlers
}

// New creates a new empty EventBus.
func New() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(Event)),
	}
}

// Subscribe registers a handler for a given event type.
// Handler will be called asynchronously when Publish is called.
func (eb *EventBus) Subscribe(eventType string, handler func(Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// SubscribeAll registers a handler that receives ALL event types.
func (eb *EventBus) SubscribeAll(handler func(Event)) {
	eb.Subscribe("*", handler)
}

// Publish dispatches an event to all subscribers asynchronously.
// Each handler runs in its own goroutine tracked by the internal WaitGroup.
func (eb *EventBus) Publish(event Event) {
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	eb.mu.RLock()
	handlers := make([]func(Event), 0)
	// TODO: Collect handlers for event.Type
	handlers = append(handlers, eb.subscribers[event.Type]...)
	// TODO: Also collect handlers for "*" (SubscribeAll)
	handlers = append(handlers, eb.subscribers["*"]...)
	eb.mu.RUnlock()

	for _, h := range handlers {
		eb.wg.Add(1)
		go func() {
			h(event)
			eb.wg.Done()
		}()
	}
}

// PublishSync dispatches an event synchronously (handlers run in the same goroutine).
// Useful for testing to avoid timing issues.
func (eb *EventBus) PublishSync(event Event) {
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	eb.mu.RLock()
	handlers := make([]func(Event), 0)
	// TODO: Same collection as Publish
	handlers = append(handlers, eb.subscribers[event.Type]...)
	handlers = append(handlers, eb.subscribers["*"]...)
	eb.mu.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}

// WaitForAll blocks until all in-flight async handlers have completed.
func (eb *EventBus) WaitForAll() {
	eb.wg.Wait()
}
