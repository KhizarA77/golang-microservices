package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Task 1: Event Bus ===")
	task1EventBus()

	fmt.Println("\n=== Task 2: Sorting Strategies ===")
	task2SortingStrategies()

	fmt.Println("\n=== Task 3: Text Editor with Undo/Redo ===")
	task3TextEditor()

	fmt.Println("\n=== Task 4: Order State Machine ===")
	task4StateMachine()
}

// =============================================================================
// Task 1 — Event Bus (Observer)
// =============================================================================

// Event represents something that happened in the system.
type Event struct {
	Type      string
	Payload   interface{}
	Timestamp time.Time
}

// EventHandler is a function called when an event fires.
type EventHandler func(Event)

// EventBus allows publishing events and subscribing to them.
type EventBus struct {
	// TODO: Add handlers map[string][]EventHandler
	// TODO: Add mu sync.RWMutex
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		// TODO: initialize handlers map
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe registers a handler for a given event type.
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	// TODO: Lock, append handler to eb.handlers[eventType]
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish notifies all subscribers of the given event.
// Each handler is called in its own goroutine.
func (eb *EventBus) Publish(event Event, wg *sync.WaitGroup) {
	// TODO: RLock, get handlers slice for event.Type
	// TODO: For each handler, wg.Add(1), launch goroutine: defer wg.Done(), call handler(event)
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	handlers := eb.handlers[event.Type]
	for _, h := range handlers {
		wg.Add(1)
		go func() {
			h(event)
			wg.Done()
		}()
	}
}

func task1EventBus() {
	bus := NewEventBus()
	var wg sync.WaitGroup

	// TODO: Subscribe "user.registered":
	//   - Email service: prints "Sending welcome email to <email>"
	//   - Audit log:     prints "Audit: user <email> registered at <time>"
	bus.Subscribe("user.registered", func(e Event) {
		payload := e.Payload.(map[string]string)
		fmt.Printf("Sending welcome email to %s\n", payload["email"])
	})
	bus.Subscribe("user.registered", func(e Event) {
		payload := e.Payload.(map[string]string)
		fmt.Printf("Audit: user %s registered at time %v\n", payload["email"], e.Timestamp)
	})
	// TODO: Subscribe "order.placed":
	//   - Inventory: prints "Reserving stock for order <id>"
	//   - Notify:    prints "Notifying user of order <id>"

	bus.Subscribe("order.placed", func(e Event) {
		payload := e.Payload.(map[string]string)
		fmt.Printf("Reserving stock for order %s\n", payload["id"])
	})

	bus.Subscribe("order.placed", func(e Event) {
		payload := e.Payload.(map[string]string)
		fmt.Printf("Notifying user of order %s\n", payload["id"])
	})
	// Publish events
	bus.Publish(Event{
		Type:      "user.registered",
		Payload:   map[string]string{"email": "alice@example.com"},
		Timestamp: time.Now(),
	}, &wg)

	bus.Publish(Event{
		Type:      "order.placed",
		Payload:   map[string]string{"id": "ORD-001", "user": "alice"},
		Timestamp: time.Now(),
	}, &wg)

	bus.Publish(Event{
		Type:      "user.registered",
		Payload:   map[string]string{"email": "bob@example.com"},
		Timestamp: time.Now(),
	}, &wg)

	wg.Wait()
	fmt.Println("All events processed")
}

// =============================================================================
// Task 2 — Sorting Strategies
// =============================================================================

// SortStrategy defines the interface for sorting algorithms.
type SortStrategy interface {
	Sort(data []int) []int
	Name() string
}

// BubbleSort implements the bubble sort algorithm.
type BubbleSort struct{}

func (b BubbleSort) Name() string { return "BubbleSort" }
func (b BubbleSort) Sort(data []int) []int {
	// TODO: Make a copy of data
	// TODO: Implement bubble sort
	// TODO: Return sorted copy
	n := len(data)
	cp := make([]int, n)
	copy(cp, data)
	for i := 1; i < n; i++ {
		for j := 0; j <= n-i-1; j++ {
			if cp[j] > cp[j+1] {
				temp := cp[j]
				cp[j] = cp[j+1]
				cp[j+1] = temp
			}
		}
	}
	return cp
}

// SelectionSort implements the selection sort algorithm.
type SelectionSort struct{}

func (s SelectionSort) Name() string { return "SelectionSort" }
func (s SelectionSort) Sort(data []int) []int {
	// TODO: Make a copy of data
	// TODO: Implement selection sort
	// TODO: Return sorted copy
	n := len(data)
	cp := make([]int, n)
	copy(cp, data)
	for i := 0; i < n-1; i++ {
		min_index := i
		for j := i + 1; j < n; j++ {
			if cp[j] < cp[min_index] {
				min_index = j
			}
		}
		temp := cp[i]
		cp[i] = cp[min_index]
		cp[min_index] = temp
	}
	return cp
}

// StdSort wraps Go's built-in sort.
type StdSort struct{}

func (s StdSort) Name() string { return "StdSort" }
func (s StdSort) Sort(data []int) []int {
	cp := make([]int, len(data))
	copy(cp, data)
	sort.Ints(cp)
	return cp
}

// Sorter uses a strategy to sort data.
type Sorter struct {
	strategy SortStrategy
}

func (s *Sorter) SetStrategy(strategy SortStrategy) {
	// TODO: Set s.strategy
	s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
	// TODO: Delegate to s.strategy.Sort
	return s.strategy.Sort(data)
}

func task2SortingStrategies() {
	// Generate random 1000-element slice
	data := make([]int, 1000)
	for i := range data {
		data[i] = rand.Intn(10000)
	}

	sorter := &Sorter{}
	strategies := []SortStrategy{BubbleSort{}, SelectionSort{}, StdSort{}}

	for _, strategy := range strategies {
		sorter.SetStrategy(strategy)
		start := time.Now()
		sorted := sorter.Sort(data)
		elapsed := time.Since(start)
		if sorted != nil {
			fmt.Printf("%s: first 10 = %v, elapsed = %v\n",
				strategy.Name(), sorted[:10], elapsed)
		}
	}
}

// =============================================================================
// Task 3 — Text Editor with Undo/Redo
// =============================================================================

// TextEditor holds the current document content.
type TextEditor struct {
	content string
}

func (e *TextEditor) Content() string { return e.content }

func (e *TextEditor) Insert(pos int, text string) {
	// TODO: Insert text at position pos in e.content
	// Hint: e.content = e.content[:pos] + text + e.content[pos:]
	e.content = e.content[:pos] + text + e.content[pos:]
}

func (e *TextEditor) Delete(start, end int) string {
	// TODO: Delete content[start:end], return the deleted text
	// Hint: deleted = e.content[start:end]; e.content = e.content[:start] + e.content[end:]
	del := e.content[start:end]
	e.content = e.content[:start] + e.content[end:]
	return del
}

// Command is an undoable/redoable action.
type Command interface {
	Execute()
	Undo()
}

// InsertCommand inserts text into the editor.
type InsertCommand struct {
	editor *TextEditor
	pos    int
	text   string
}

func (c *InsertCommand) Execute() {
	// TODO: Call c.editor.Insert(c.pos, c.text)
	c.editor.Insert(c.pos, c.text)
}

func (c *InsertCommand) Undo() {
	// TODO: Call c.editor.Delete(c.pos, c.pos + len(c.text))
	c.editor.Delete(c.pos, c.pos+len(c.text))
}

// DeleteCommand deletes a range from the editor.
type DeleteCommand struct {
	editor      *TextEditor
	start, end  int
	deletedText string // saved for undo
}

func (c *DeleteCommand) Execute() {
	// TODO: c.deletedText = c.editor.Delete(c.start, c.end)
	c.deletedText = c.editor.Delete(c.start, c.end)
}

func (c *DeleteCommand) Undo() {
	// TODO: c.editor.Insert(c.start, c.deletedText)
	c.editor.Insert(c.start, c.deletedText)
}

// CommandHistory manages undo/redo stacks.
type CommandHistory struct {
	undoStack []Command
	redoStack []Command
}

func (h *CommandHistory) Execute(cmd Command) {
	// TODO: Execute cmd
	cmd.Execute()
	// TODO: Push to undoStack
	h.undoStack = append(h.undoStack, cmd)
	// TODO: Clear redoStack (a new action invalidates redo history)
	h.redoStack = make([]Command, 0)
}

func (h *CommandHistory) Undo() {
	// TODO: If undoStack is empty, print "nothing to undo", return
	if len(h.undoStack) == 0 {
		fmt.Println("nothing to undo")
		return
	}
	// TODO: Pop from undoStack, call Undo(), push to redoStack
	cmd := h.undoStack[len(h.undoStack)-1]
	cmd.Undo()
	h.undoStack = h.undoStack[:len(h.undoStack)-1]
	h.redoStack = append(h.redoStack, cmd)
}

func (h *CommandHistory) Redo() {
	// TODO: If redoStack is empty, print "nothing to redo", return
	if len(h.redoStack) == 0 {
		fmt.Println("nothing to redo")
		return
	}
	// TODO: Pop from redoStack, call Execute(), push back to undoStack
	cmd := h.redoStack[len(h.redoStack)-1]
	cmd.Execute()
	h.redoStack = h.redoStack[:len(h.redoStack)-1]
	h.undoStack = append(h.undoStack, cmd)
}

func task3TextEditor() {
	editor := &TextEditor{}
	history := &CommandHistory{}

	// Insert "Hello"
	history.Execute(&InsertCommand{editor: editor, pos: 0, text: "Hello"})
	fmt.Printf("After insert 'Hello':   %q\n", editor.Content())

	// Insert ", World" at position 5
	history.Execute(&InsertCommand{editor: editor, pos: 5, text: ", World"})
	fmt.Printf("After insert ', World': %q\n", editor.Content())

	// Delete positions 5–7 (", ")
	history.Execute(&DeleteCommand{editor: editor, start: 5, end: 7})
	fmt.Printf("After delete [5:7]:     %q\n", editor.Content())

	// Undo delete
	history.Undo()
	fmt.Printf("After undo:             %q\n", editor.Content())

	// Undo second insert
	history.Undo()
	fmt.Printf("After undo:             %q\n", editor.Content())

	// Redo second insert
	history.Redo()
	fmt.Printf("After redo:             %q\n", editor.Content())
}

// =============================================================================
// Task 4 — Order State Machine
// =============================================================================

// Order represents an e-commerce order.
type Order struct {
	ID      string
	state   OrderState
	history []string
}

func NewOrder(id string) *Order {
	o := &Order{ID: id}
	o.state = &PendingState{}
	o.log("created in state: Pending")
	return o
}

func (o *Order) log(msg string) {
	o.history = append(o.history, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
}

func (o *Order) Next() error {
	err := o.state.Next(o)
	if err != nil {
		return err
	}
	o.log("transitioned to: " + o.state.Name())
	return nil
}

func (o *Order) Cancel() error {
	err := o.state.Cancel(o)
	if err != nil {
		return err
	}
	o.log("cancelled: now in " + o.state.Name())
	return nil
}

func (o *Order) PrintHistory() {
	fmt.Printf("Order %s history:\n", o.ID)
	for _, entry := range o.history {
		fmt.Println(" ", entry)
	}
}

// OrderState interface must be implemented by each state.
type OrderState interface {
	Next(o *Order) error
	Cancel(o *Order) error
	Name() string
}

// PendingState — order created, not yet processing
type PendingState struct{}

func (s *PendingState) Name() string { return "Pending" }
func (s *PendingState) Next(o *Order) error {
	// TODO: Set o.state = &ProcessingState{}
	o.state = &ProcessingState{}
	return nil
}
func (s *PendingState) Cancel(o *Order) error {
	// TODO: Set o.state = &CancelledState{}
	o.state = &CancelledState{}
	return nil
}

// ProcessingState — payment received, being fulfilled
type ProcessingState struct{}

func (s *ProcessingState) Name() string { return "Processing" }
func (s *ProcessingState) Next(o *Order) error {
	// TODO: Set o.state = &ShippedState{}
	o.state = &ShippedState{}
	return nil
}
func (s *ProcessingState) Cancel(o *Order) error {
	// TODO: Set o.state = &CancelledState{}
	o.state = &CancelledState{}
	return nil
}

// ShippedState — package is in transit
type ShippedState struct{}

func (s *ShippedState) Name() string { return "Shipped" }
func (s *ShippedState) Next(o *Order) error {
	// TODO: Set o.state = &DeliveredState{}
	o.state = &DeliveredState{}
	return nil
}
func (s *ShippedState) Cancel(o *Order) error {
	// TODO: Return error: "cannot cancel a shipped order"
	return errors.New("cannot cancel a shipped order")
}

// DeliveredState — package received by customer
type DeliveredState struct{}

func (s *DeliveredState) Name() string { return "Delivered" }
func (s *DeliveredState) Next(o *Order) error {
	// TODO: Return error: "order already delivered"
	return errors.New("order already delivered")
}
func (s *DeliveredState) Cancel(o *Order) error {
	// TODO: Return error: "cannot cancel a delivered order"
	return errors.New("cannot cancel a delivered order")
}

// CancelledState — order has been cancelled
type CancelledState struct{}

func (s *CancelledState) Name() string { return "Cancelled" }
func (s *CancelledState) Next(o *Order) error {
	// TODO: Return error: "cannot advance a cancelled order"
	return errors.New("cannot advance a cancelled order")
}
func (s *CancelledState) Cancel(o *Order) error {
	// TODO: Return error: "order already cancelled"
	return errors.New("order already cancelled")
}

func task4StateMachine() {
	// Happy path: Pending -> Processing -> Shipped -> Delivered
	order := NewOrder("ORD-001")
	states := []string{}

	for {
		states = append(states, order.state.Name())
		err := order.Next()
		if err != nil {
			fmt.Println("Stopped:", err)
			break
		}
	}
	states = append(states, order.state.Name())
	fmt.Printf("State progression: %v\n\n", states)
	order.PrintHistory()

	// Error case: try to cancel a delivered order
	fmt.Println()
	order2 := NewOrder("ORD-002")
	order2.Next() // -> Processing
	order2.Next() // -> Shipped
	order2.Next() // -> Delivered
	if err := order2.Cancel(); err != nil {
		fmt.Println("Cancel error:", err)
	}

	// Cancel from Pending
	order3 := NewOrder("ORD-003")
	order3.Cancel()
	fmt.Println("Order 3 state:", order3.state.Name())
	if err := order3.Cancel(); err != nil {
		fmt.Println("Second cancel error:", err)
	}
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
