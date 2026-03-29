package eventbus

import (
	"sync"
	"time"
)

type Event struct {
	Type        string
	AggregateID string
	Payload     interface{}
	OccurredAt  time.Time
	Version     int
}

type EventBus struct {
	subscribers map[string][]func(Event)
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(Event)),
	}
}

func (eb *EventBus) Subscribe(eventType string, handler func(Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, h := range eb.subscribers[event.Type] {
		eb.wg.Add(1)
		go func() {
			h(event)
			eb.wg.Done()
		}()
	}

}

func (eb *EventBus) WaitForAll() {
	eb.wg.Wait()
}
