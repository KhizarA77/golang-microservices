# Lab 12 — Event-Driven Architecture

**Level:** Advanced
**Topic:** Event Bus, Domain Events, Async Processing, CQRS Basics, Outbox Pattern

---

## Background

### What is Event-Driven Architecture?

In event-driven architecture (EDA), components communicate by producing and consuming **events** rather than calling each other directly. This creates **loose coupling** — the producer doesn't know who is consuming its events.

```
[Order Service] ──publishes──► "order.placed" event ──► [Inventory Service]
                                                      ──► [Notification Service]
                                                      ──► [Analytics Service]
```

Compare to direct (synchronous) coupling:
```go
// Tightly coupled — Order service must know about all consumers
func PlaceOrder(order Order) {
    saveOrder(order)
    inventory.Reserve(order)       // direct call
    notifications.Send(order)     // direct call
    analytics.Track(order)        // direct call
}

// Loosely coupled — Order service just publishes an event
func PlaceOrder(order Order) {
    saveOrder(order)
    eventBus.Publish("order.placed", order)  // knows nothing about consumers
}
```

---

### Synchronous vs Asynchronous Events

**Synchronous in-process events** (what this lab covers):
- All event handlers run in the same process
- Useful for decoupling within a service
- Fast, no serialization overhead
- Not durable (events lost if process crashes)

**Asynchronous message queues** (production scale):
- Events flow through a broker (Kafka, RabbitMQ, NATS)
- Durable, survives restarts
- Handlers can be in different services/processes
- Higher complexity (ordering, at-least-once delivery, idempotency)

---

### CQRS — Command Query Responsibility Segregation

CQRS separates read and write operations:

```
             Write side                    Read side
              ─────────                   ──────────
Client ──► Command ──► CommandHandler     Query ──► QueryHandler ──► Read Model
                           │                              │
                           ▼                              │
                        Write Model ──event──► Projector ─┘
```

**Commands** mutate state (PlaceOrder, RegisterUser).
**Queries** read state and never mutate.

The read model can be optimized for reading (denormalized, cached, pre-computed).

---

### Event Sourcing

Instead of storing current state, store the sequence of events that led to it:

```
Normal: User{name: "Alice", email: "alice@example.com"}

Event sourced:
  UserRegistered{name: "Alice", email: "alice@old.com"}
  UserEmailChanged{email: "alice@example.com"}
```

To get current state, replay all events. This gives you a full audit trail.

---

### The Outbox Pattern

Problem: How do you atomically write to your database AND publish an event?
(If you do both and one fails, you get inconsistency.)

Solution: Write the event to an "outbox" table in the same transaction as the data. A separate process reads the outbox and publishes events. This guarantees at-least-once delivery.

```
Transaction:
  INSERT INTO orders ...
  INSERT INTO outbox (event_type, payload) VALUES ('order.placed', ...)

Outbox processor (separate goroutine):
  SELECT * FROM outbox WHERE published = false
  FOR EACH event: publish to broker, mark as published
```

---

## Project Structure

```
lab12-event-driven/
├── go.mod
├── main.go
├── eventbus/
│   └── eventbus.go         ← Thread-safe async event bus
├── events/
│   └── events.go           ← Event type definitions
├── domain/
│   ├── order.go            ← Order aggregate with domain events
│   └── commands.go         ← Command types
├── handlers/
│   ├── order_handler.go    ← Command handlers
│   ├── inventory.go        ← Inventory event subscriber
│   └── notification.go     ← Notification event subscriber
└── query/
    └── order_query.go      ← Read model and projector
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Build a thread-safe, async event bus
- Define domain events as typed structs
- Wire command handlers to domain logic
- Subscribe multiple independent handlers to events
- Build a CQRS read model (projector pattern)
- Implement a simple event store (in-memory)

---

## Tasks

### Task 1 — Event Bus

Build `EventBus` in `eventbus/eventbus.go`:

```go
type EventBus struct {
    subscribers map[string][]func(Event)
    mu          sync.RWMutex
}
```

Methods:
- `Subscribe(eventType string, handler func(Event))` — register handler
- `Publish(event Event)` — dispatch to all handlers **asynchronously** (goroutine per handler)
- `PublishSync(event Event)` — dispatch synchronously (useful for testing)
- `WaitForAll()` — wait for all async handlers to finish (use WaitGroup)

`Event` struct: `Type string`, `AggregateID string`, `Payload interface{}`, `OccurredAt time.Time`, `Version int`

### Task 2 — Domain Events

Define all events in `events/events.go`:
- `OrderPlaced { OrderID, UserID, Items []OrderItem, TotalPrice float64 }`
- `OrderCancelled { OrderID, Reason string }`
- `OrderShipped { OrderID, TrackingNumber string }`
- `InventoryReserved { OrderID, Items []OrderItem }`
- `InventoryReleaseFailed { OrderID, Reason string }`

### Task 3 — Order Aggregate with Domain Events

In `domain/order.go`, build an `Order` aggregate:

```go
type Order struct {
    ID         string
    UserID     string
    Items      []OrderItem
    Status     string
    TotalPrice float64
    events     []Event     // uncommitted events
}
```

Methods:
- `Place(userID string, items []OrderItem) error` — validate, set status "pending", record `OrderPlaced` event
- `Cancel(reason string) error` — validate state, set status "cancelled", record `OrderCancelled`
- `Ship(trackingNumber string) error` — set status "shipped", record `OrderShipped`
- `UncommittedEvents() []Event` — return pending events
- `ClearEvents()` — clear after publishing

The aggregate never directly calls the event bus — it just records events internally. The command handler publishes them.

### Task 4 — Command Handlers and Subscribers

In `handlers/order_handler.go`:
```go
type OrderCommandHandler struct {
    orders map[string]*domain.Order
    bus    *eventbus.EventBus
    mu     sync.Mutex
}

func (h *OrderCommandHandler) HandlePlaceOrder(cmd PlaceOrderCommand) error
func (h *OrderCommandHandler) HandleCancelOrder(cmd CancelOrderCommand) error
func (h *OrderCommandHandler) HandleShipOrder(cmd ShipOrderCommand) error
```

Each handler:
1. Executes the command on the aggregate
2. Publishes uncommitted events to the bus
3. Persists the aggregate (in-memory map)

In `handlers/inventory.go`, subscribe to `order.placed`:
- Print `"[INVENTORY] Reserving stock for order %s"`
- Simulate 100ms processing time
- Publish `inventory.reserved` event

In `handlers/notification.go`, subscribe to all order events:
- Print `"[NOTIFICATION] Order %s: %s"`

### Task 5 — CQRS Read Model

In `query/order_query.go`, build a projector that subscribes to all order events and maintains a read model:

```go
type OrderView struct {
    ID         string
    UserID     string
    Status     string
    TotalPrice float64
    Events     []string  // log of events that happened
    UpdatedAt  time.Time
}

type OrderProjector struct {
    views map[string]*OrderView
    mu    sync.RWMutex
}
```

Subscribe to all order events and update the read model accordingly.
Expose `GetOrderView(id string) (*OrderView, bool)` and `ListOrderViews() []*OrderView`.

### Task 6 — Demo in `main.go`

Wire everything together and demonstrate:
1. Place 3 orders
2. Ship one order
3. Cancel one order
4. Query the read model — show all order views with their event history
5. Show that the event bus dispatched to all subscribers

---

## Tips

- Use `sync.WaitGroup` in the event bus to track in-flight handlers — this lets `WaitForAll()` work correctly.
- The aggregate's `events []Event` slice should be unexported; expose it only through `UncommittedEvents()`.
- Make the projector subscribe before publishing any events — otherwise it misses events.
- For the read model, store `[]string` of event descriptions so you can see the full history.

---

## Running Your Solution

```bash
cd lab12-event-driven
go run .
```
