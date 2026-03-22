package handlers

import (
	"fmt"
	"lab12/eventbus"
	"lab12/events"
)

// SubscribeNotifications registers notification handlers for all order events.
// Each prints "[NOTIFICATION] Order <id>: <message>"
func SubscribeNotifications(bus *eventbus.EventBus) {
	bus.Subscribe(events.OrderPlaced, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.OrderPlacedPayload
		payload, ok := e.Payload.(events.OrderPlacedPayload)
		if !ok {
			return
		}
		// TODO: Print "[NOTIFICATION] Order <OrderID>: order placed by <UserID>"
		fmt.Printf("[NOTIFICATION] Order %s: order placed by %s\n", payload.OrderID, payload.UserID)
	})

	bus.Subscribe(events.OrderCancelled, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.OrderCancelledPayload
		payload, ok := e.Payload.(events.OrderCancelledPayload)
		if !ok {
			return
		}
		// TODO: Print "[NOTIFICATION] Order <OrderID>: cancelled — <Reason>"
		fmt.Printf("[NOTIFICATION] Order %s: cancelled - %s\n", payload.OrderID, payload.Reason)
	})

	bus.Subscribe(events.OrderShipped, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.OrderShippedPayload
		payload, ok := e.Payload.(events.OrderShippedPayload)
		if !ok {
			return
		}
		// TODO: Print "[NOTIFICATION] Order <OrderID>: shipped with tracking <TrackingNumber>"
		fmt.Printf("[NOTIFICATION] Order %s: shipped with tracking %s\n", payload.OrderID, payload.TrackingNumber)
	})

	bus.Subscribe(events.InventoryReserved, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.InventoryReservedPayload
		payload, ok := e.Payload.(events.InventoryReservedPayload)
		if !ok {
			return
		}
		// TODO: Print "[NOTIFICATION] Order <OrderID>: inventory reserved"
		fmt.Printf("[NOTIFICATION] Order %s: inventory reserved\n", payload.OrderID)
	})
}
