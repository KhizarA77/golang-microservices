# Lab 06 — Behavioral Design Patterns

**Level:** Advanced
**Topic:** Observer, Strategy, Command (with Undo), State Machine

---

## Background

Behavioral patterns deal with **algorithms and communication between objects** — how responsibilities are assigned and how objects interact.

---

### Observer

Defines a one-to-many dependency: when one object (the subject) changes state, all its dependents (observers) are notified automatically.

```go
type Event struct {
    Type    string
    Payload interface{}
}

type Handler func(event Event)

type EventBus struct {
    handlers map[string][]Handler
    mu       sync.RWMutex
}

func (eb *EventBus) Subscribe(eventType string, h Handler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.handlers[eventType] = append(eb.handlers[eventType], h)
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    handlers := eb.handlers[event.Type]
    eb.mu.RUnlock()
    for _, h := range handlers {
        go h(event) // async notification
    }
}
```

**When to use:** GUI event systems, domain events, notification systems, decoupling components.

---

### Strategy

Defines a family of algorithms, encapsulates each one, and makes them interchangeable. The strategy is selected at runtime.

```go
type SortStrategy interface {
    Sort(data []int) []int
}

type BubbleSort struct{}
type QuickSort  struct{}

func (b BubbleSort) Sort(data []int) []int { /* ... */ }
func (q QuickSort)  Sort(data []int) []int { /* ... */ }

type Sorter struct {
    strategy SortStrategy
}

func (s *Sorter) SetStrategy(strategy SortStrategy) {
    s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
    return s.strategy.Sort(data)
}
```

In Go, strategy can also be a `func` type instead of an interface — even simpler:

```go
type SortFunc func([]int) []int

type Sorter struct { sortFn SortFunc }
```

**When to use:** Multiple algorithms for the same task, switching behaviors at runtime, A/B testing different implementations.

---

### Command

Encapsulates a request as an object. Allows parameterizing methods, queuing requests, and supporting undoable operations.

```go
type Command interface {
    Execute()
    Undo()
}

type TextEditor struct { content string }

type InsertCommand struct {
    editor *TextEditor
    text   string
    pos    int
}

func (c *InsertCommand) Execute() {
    // insert c.text at c.pos
}
func (c *InsertCommand) Undo() {
    // remove the inserted text
}

// History stack
type History struct {
    commands []Command
}

func (h *History) Execute(cmd Command) {
    cmd.Execute()
    h.commands = append(h.commands, cmd)
}

func (h *History) Undo() {
    if len(h.commands) == 0 { return }
    last := h.commands[len(h.commands)-1]
    last.Undo()
    h.commands = h.commands[:len(h.commands)-1]
}
```

**When to use:** Undo/redo, transaction logs, job queues, macro recording.

---

### State Machine

An object's behavior changes based on its internal state. Instead of a giant switch statement, each state is its own type implementing a common interface.

```go
type OrderState interface {
    Next(order *Order) error
    Cancel(order *Order) error
    String() string
}

type PendingState   struct{}
type ProcessingState struct{}
type ShippedState   struct{}

func (p PendingState) Next(o *Order) error {
    o.state = ProcessingState{}
    return nil
}
// etc.
```

**When to use:** Order processing, connection lifecycle, game states, UI form wizards.

---

## Learning Objectives

By the end of this lab you will be able to:

- Build a thread-safe event bus using the Observer pattern
- Implement swappable algorithms using the Strategy pattern
- Create undoable operations with the Command pattern
- Model complex state transitions with a State Machine

---

## Tasks

### Task 1 — Event Bus (Observer)

Build a thread-safe `EventBus`:
- `Subscribe(eventType string, handler func(Event))` — register a listener
- `Publish(event Event)` — notify all subscribers for that event type
- `Unsubscribe(eventType string, handler func(Event))` — remove a listener

`Event` struct: `Type string`, `Payload interface{}`, `Timestamp time.Time`

Create these event types and subscribers:
- `"user.registered"` → Email service prints `"Sending welcome email to <email>"`
- `"user.registered"` → Audit log prints `"Audit: user registered at <time>"`
- `"order.placed"` → Inventory service prints `"Reserving stock for order <id>"`
- `"order.placed"` → Notification service prints `"Notifying user of order <id>"`

Publish several events and verify all subscribers are notified. Use `sync.WaitGroup` to wait for async notifications.

### Task 2 — Sorting Strategies

Implement three sort strategies:
1. `BubbleSort` — O(n²), simple but slow
2. `SelectionSort` — O(n²), fewer swaps
3. `StdSort` — wraps Go's `sort.Ints`

`SortStrategy` interface: `Sort(data []int) []int` (return a sorted copy).

Build a `Sorter` that holds a `SortStrategy`. Add `SetStrategy(SortStrategy)` and `Sort([]int) []int`.

Benchmark each strategy against a 1000-element random slice using `time.Since`. Print the sorted first 10 elements and elapsed time.

### Task 3 — Text Editor with Undo/Redo

Build a simple text editor with undo/redo:

`TextEditor` struct: holds `content string`.
Methods: `Insert(pos int, text string)`, `Delete(start, end int)`, `Content() string`

`Command` interface: `Execute()`, `Undo()`

Implement:
- `InsertCommand` — records position and text, Execute inserts, Undo removes
- `DeleteCommand` — records range and deleted text, Execute deletes, Undo re-inserts

`CommandHistory` struct:
- `Execute(cmd Command)` — executes and adds to undo stack
- `Undo()` — undoes last command, pushes to redo stack
- `Redo()` — redoes last undone command

Demo:
```
Insert "Hello" -> "Hello"
Insert ", World" at 5 -> "Hello, World"
Delete 5-7 -> "Hello World"
Undo -> "Hello, World"
Undo -> "Hello"
Redo -> "Hello, World"
```

### Task 4 — Order State Machine

Model an e-commerce order with these states:

```
Pending -> Processing -> Shipped -> Delivered
    \           \
     \           -> Cancelled
      -> Cancelled
```

Each state implements:
```go
type OrderState interface {
    Next(o *Order) error   // transition to next state
    Cancel(o *Order) error // cancel the order
    Name() string          // state name
}
```

`Order` struct: `ID string`, `state OrderState`, `history []string` (log of state transitions)

Implement all four states. Illegal transitions should return an error (e.g., can't cancel a delivered order).

Demo: create an order, walk it through all states, then try to cancel a delivered order (should error).

---

## Tips

- For the event bus, use `sync.RWMutex` — many goroutines read handlers concurrently, but Subscribe modifies them exclusively.
- For Unsubscribe in Go, you can't directly compare `func` values. Use an ID-based approach or store handlers in a `sync.Map` with string keys.
- For the sort strategies, make a copy of the input slice inside `Sort()` so the original is not mutated.
- For command undo: the `DeleteCommand` must save the deleted text *before* executing, so it can restore it on undo.
- For the state machine: store a reference to the `*Order` in state transitions to update `order.state` directly.

---

## Running Your Solution

```bash
cd lab06-behavioral-patterns
go run .
```
