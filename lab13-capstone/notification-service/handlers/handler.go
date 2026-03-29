package handlers

import (
	"encoding/json"
	"net/http"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("POST /webhook/order", h.handleWebhook)
}

func (h *NotificationHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{
		"status":  "ok",
		"service": "notification-service",
	})
}

func (h *NotificationHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 501, map[string]string{
		"status":  "not implemented",
		"service": "notification-service",
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
