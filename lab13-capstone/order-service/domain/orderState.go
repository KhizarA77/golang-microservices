package domain

import "fmt"

type OrderState interface {
	Next(o *Order) error
	Cancel(o *Order) error
	Name() string
}

type PendingState struct{}
type ProcessingState struct{}
type ShippedState struct{}
type DeliveredState struct{}
type CancelledState struct{}

func (p *PendingState) Next(o *Order) error {
	o.Status = &ProcessingState{}
	return nil
}

func (p *PendingState) Cancel(o *Order) error {
	o.Status = &CancelledState{}
	return nil
}

func (p *PendingState) Name() string {
	return "pending"
}

func (p *ProcessingState) Next(o *Order) error {
	o.Status = &ShippedState{}
	return nil
}

func (p *ProcessingState) Cancel(o *Order) error {
	o.Status = &CancelledState{}
	return nil
}

func (p *ProcessingState) Name() string {
	return "processing"
}

func (p *ShippedState) Next(o *Order) error {
	o.Status = &DeliveredState{}
	return nil
}

func (p *ShippedState) Cancel(o *Order) error {
	return fmt.Errorf("cannot cancel a shipped order")
}

func (p *ShippedState) Name() string {
	return "shipped"
}

func (p *DeliveredState) Next(o *Order) error {
	return fmt.Errorf("delivered is the final state")
}
func (p *DeliveredState) Cancel(o *Order) error {
	return fmt.Errorf("cannot cancel a delivered order")
}
func (p *DeliveredState) Name() string {
	return "delivered"
}

func (p *CancelledState) Next(o *Order) error {
	return fmt.Errorf("order is cancelled")
}
func (p *CancelledState) Cancel(o *Order) error {
	return fmt.Errorf("order already cancelled")
}
func (p *CancelledState) Name() string {
	return "cancelled"
}
