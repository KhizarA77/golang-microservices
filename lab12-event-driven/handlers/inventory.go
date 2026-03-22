package handlers

import (
	"fmt"
	"lab12/eventbus"
	"lab12/events"
	"time"
)

// SubscribeInventory registers the inventory service handlers on the event bus.
// It listens for order.placed and reserves stock, then publishes inventory.reserved.
func SubscribeInventory(bus *eventbus.EventBus) {
	bus.Subscribe(events.OrderPlaced, func(e eventbus.Event) {
		// TODO: Type-assert e.Payload to events.OrderPlacedPayload
		//       If assertion fails, return early
		payload, ok := e.Payload.(events.OrderPlacedPayload)
		if !ok {
			return
		}

		// TODO: Simulate 100ms processing time
		time.Sleep(100 * time.Millisecond)
		// TODO: Print "[INVENTORY] Reserving stock for order <OrderID> (<n> items)"

		fmt.Printf("[INVENTORY] Reserving stock for order %s (%d items)\n", payload.OrderID, len(payload.Items))

		// TODO: Publish an inventory.reserved event back to the bus:
		ev := eventbus.Event{
			Type:        events.InventoryReserved,
			AggregateID: payload.OrderID,
			Payload:     events.InventoryReservedPayload{OrderID: payload.OrderID, Items: payload.Items},
		}
		bus.Publish(ev)
	})
}
