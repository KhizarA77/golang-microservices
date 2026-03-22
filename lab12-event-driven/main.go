package main

import (
	"fmt"
	"lab12/domain"
	"lab12/eventbus"
	"lab12/events"
	"lab12/handlers"
	"lab12/query"
)

func main() {
	bus := eventbus.New()

	// Set up the read model projector BEFORE any events are published
	projector := query.NewOrderProjector()
	projector.Subscribe(bus)

	// Register event subscribers
	handlers.SubscribeInventory(bus)
	handlers.SubscribeNotifications(bus)

	// Set up the command handler
	handler := handlers.NewOrderCommandHandler(bus)

	// --- Place 3 orders ---
	fmt.Println("=== Placing Orders ===")

	order1, err := handler.HandlePlaceOrder(domain.PlaceOrderCommand{
		UserID: "user-001",
		Items: []events.OrderItem{
			{ProductID: "P001", ProductName: "Laptop", Quantity: 1, UnitPrice: 999.99},
			{ProductID: "P002", ProductName: "Mouse", Quantity: 2, UnitPrice: 29.99},
		},
	})
	if err != nil {
		fmt.Println("Error placing order 1:", err)
	} else {
		fmt.Printf("Placed: %s (total: $%.2f)\n", order1.ID, order1.TotalPrice)
	}

	order2, err := handler.HandlePlaceOrder(domain.PlaceOrderCommand{
		UserID: "user-002",
		Items: []events.OrderItem{
			{ProductID: "P003", ProductName: "Keyboard", Quantity: 1, UnitPrice: 79.99},
		},
	})
	if err != nil {
		fmt.Println("Error placing order 2:", err)
	} else {
		fmt.Printf("Placed: %s (total: $%.2f)\n", order2.ID, order2.TotalPrice)
	}

	order3, err := handler.HandlePlaceOrder(domain.PlaceOrderCommand{
		UserID: "user-001",
		Items: []events.OrderItem{
			{ProductID: "P001", ProductName: "Laptop", Quantity: 2, UnitPrice: 999.99},
		},
	})
	if err != nil {
		fmt.Println("Error placing order 3:", err)
	} else {
		fmt.Printf("Placed: %s (total: $%.2f)\n", order3.ID, order3.TotalPrice)
	}

	// Wait for async handlers (e.g. inventory reservation) to complete
	bus.WaitForAll()

	// --- Ship order 1 ---
	fmt.Println("\n=== Shipping Order 1 ===")
	if order1 != nil {
		if err := handler.HandleShipOrder(domain.ShipOrderCommand{
			OrderID:        order1.ID,
			TrackingNumber: "TRACK-12345",
		}); err != nil {
			fmt.Println("Error:", err)
		}
	}

	// --- Cancel order 2 ---
	fmt.Println("\n=== Cancelling Order 2 ===")
	if order2 != nil {
		if err := handler.HandleCancelOrder(domain.CancelOrderCommand{
			OrderID: order2.ID,
			Reason:  "Customer requested cancellation",
		}); err != nil {
			fmt.Println("Error:", err)
		}
	}

	// Wait for any remaining async handlers
	bus.WaitForAll()

	// --- Query the read model ---
	fmt.Println("\n=== Order Read Model (Projector) ===")
	for _, view := range projector.ListViews() {
		fmt.Printf("\nOrder %s | User: %s | Status: %s | Total: $%.2f\n",
			view.ID, view.UserID, view.Status, view.TotalPrice)
		for _, entry := range view.EventLog {
			fmt.Printf("  - %s\n", entry)
		}
	}

	// Suppress unused variable warnings until tasks are complete
	_ = order3
}
