package processor

import (
	"fmt"
	"log"
	"order-service/domain"
	"order-service/eventbus"
	"order-service/events"
	"sync"
	"time"
)

// orderStore is the minimal interface the processor needs — just Update.
type orderStore interface {
	Update(order *domain.Order) error
}

type OrderProcessor struct {
	store orderStore
	bus   *eventbus.EventBus
	queue chan *domain.Order
}

func NewOrderProcessor(store orderStore, bus *eventbus.EventBus) *OrderProcessor {
	return &OrderProcessor{
		store: store,
		bus:   bus,
		queue: make(chan *domain.Order, 50),
	}
}

func (p *OrderProcessor) Start(numWorkers int, wg *sync.WaitGroup) {
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("[WORKER %d] started", workerID)
			for order := range p.queue {
				p.processOrder(workerID, order)
			}
			log.Printf("[WORKER %d] stopped", workerID)
		}(i)
	}
}

func (p *OrderProcessor) Submit(order *domain.Order) bool {
	select {
	case p.queue <- order:
		return true
	default:
		return false
	}
}

func (p *OrderProcessor) Stop() {
	close(p.queue)
}

func (p *OrderProcessor) QueueLen() int {
	return len(p.queue)
}

func (p *OrderProcessor) processOrder(workerID int, order *domain.Order) {
	log.Printf("[WORKER %d] processing order %s", workerID, order.ID)

	// Pending → Processing
	if err := order.Next(); err != nil {
		log.Printf("[WORKER %d] cannot advance order %s: %v", workerID, order.ID, err)
		return
	}
	order.UpdatedAt = time.Now()
	if err := p.store.Update(order); err != nil {
		log.Printf("[WORKER %d] failed to persist order %s: %v", workerID, order.ID, err)
	}
	p.bus.Publish(eventbus.Event{
		Type:        "order.processing",
		AggregateID: order.ID,
		Payload: events.OrderEvent{
			Type:    "order.processing",
			OrderID: order.ID,
			UserID:  order.UserID,
		},
	})

	// Simulate payment processing
	time.Sleep(500 * time.Millisecond)

	// Processing → Shipped
	if err := order.Next(); err != nil {
		log.Printf("[WORKER %d] cannot ship order %s: %v", workerID, order.ID, err)
		return
	}
	tracking := fmt.Sprintf("TRACK-%s-%d", order.ID, time.Now().Unix())
	order.UpdatedAt = time.Now()
	if err := p.store.Update(order); err != nil {
		log.Printf("[WORKER %d] failed to persist order %s: %v", workerID, order.ID, err)
	}
	p.bus.Publish(eventbus.Event{
		Type:        "order.shipped",
		AggregateID: order.ID,
		Payload: events.OrderEvent{
			Type:    "order.shipped",
			OrderID: order.ID,
			UserID:  order.UserID,
			Payload: map[string]string{"tracking": tracking},
		},
	})

	log.Printf("[WORKER %d] order %s shipped (tracking: %s)", workerID, order.ID, tracking)
}
