package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"order-service/clients"
	"order-service/eventbus"
	"order-service/events"
	"order-service/processor"
	"order-service/usecase"
	"time"
)

type OrderHandler struct {
	productService *clients.ProductServiceClient
	uc             *usecase.OrderUseCase
	proc           *processor.OrderProcessor
	bus            *eventbus.EventBus
}

func NewOrderHandler(
	svc *clients.ProductServiceClient,
	uc *usecase.OrderUseCase,
	proc *processor.OrderProcessor,
	bus *eventbus.EventBus,
) *OrderHandler {
	return &OrderHandler{productService: svc, uc: uc, proc: proc, bus: bus}
}

func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("POST /api/orders", h.handleCreateOrder)
	mux.HandleFunc("GET /api/orders", h.handleListOrders)
	mux.HandleFunc("GET /api/orders/{id}", h.handleGetOrder)
}

func (h *OrderHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]interface{}{
		"status":  "ok",
		"service": "order-service",
		"workers": 3,
		"queue":   h.proc.QueueLen(),
	})
}

func (h *OrderHandler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		ProductID string `json:"productId"`
		Quantity  int    `json:"quantity"`
		UserID    string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.ProductID == "" || req.Quantity <= 0 || req.UserID == "" {
		writeJSON(w, 400, map[string]string{"error": "productId, quantity, and userId are required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	product, err := h.productService.ValidateAndReserveProduct(ctx, req.ProductID, req.Quantity)
	if err != nil {
		status := 422
		if errors.Is(err, clients.ErrProductNotFound) {
			status = 404
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	order, err := h.uc.PlaceOrder(usecase.CreateOrderInput{
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		ProductName: product.Name,
		Quantity:    req.Quantity,
		TotalPrice:  float64(req.Quantity) * product.Price,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to create order"})
		return
	}

	h.bus.Publish(eventbus.Event{
		Type:        "order.placed",
		AggregateID: order.ID,
		Payload: events.OrderEvent{
			Type:    "order.placed",
			OrderID: order.ID,
			UserID:  order.UserID,
			Payload: map[string]interface{}{
				"productName": product.Name,
				"quantity":    req.Quantity,
				"totalPrice":  order.TotalPrice,
			},
		},
	})

	if !h.proc.Submit(order) {
		log.Printf("[WARN] processing queue full, order %s will not be processed", order.ID)
	} else {
		log.Printf("Order %s queued for processing", order.ID)
	}

	writeJSON(w, 202, order)
}

func (h *OrderHandler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.uc.ListOrders()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to list orders"})
		return
	}
	writeJSON(w, 200, orders)
}

func (h *OrderHandler) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	order, err := h.uc.GetOrder(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "order not found"})
		return
	}
	writeJSON(w, 200, order)
}
