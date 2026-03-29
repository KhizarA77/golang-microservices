package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"product-service/cache"
	"product-service/models"
	"product-service/store"
	"strconv"
	"time"
)

// ProductHandler holds all HTTP handlers for the product API.
type ProductHandler struct {
	store *store.ProductStore
	cache *cache.ProductCache
}

// NewProductHandler creates a new handler with the given store.
func NewProductHandler(s *store.ProductStore, c *cache.ProductCache) *ProductHandler {
	return &ProductHandler{store: s, cache: c}
}

// RegisterRoutes registers all product routes on the given mux.
func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/products", h.handleListProducts)
	mux.HandleFunc("POST /api/products", h.handleCreateProduct)
	mux.HandleFunc("GET /api/products/{id}", h.handleGetProduct)
	mux.HandleFunc("GET /api/products/search", h.handleSearchProducts)
	mux.HandleFunc("PATCH /api/products/{id}", h.handleUpdateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", h.handleDeleteProduct)
	mux.HandleFunc("PATCH /api/products/{id}/stock", h.handleReserveStock)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok", "service": "product-service"})
	})
}

// =============================================================================
// Helpers
// =============================================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errCode, msg string) {
	writeJSON(w, status, models.ErrorResponse{
		Error:   errCode,
		Message: msg,
	})
}

func writeValidationError(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
		Error:   "validation_failed",
		Message: "One or more fields are invalid",
		Fields:  fields,
	})
}

func parseIDParam(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func parsePagination(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return
}

// =============================================================================
// Handlers
// =============================================================================

func (h *ProductHandler) handleSearchProducts(w http.ResponseWriter, r *http.Request) {
	// call search prods here
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, 400, map[string]string{"error": "q parameter required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	results := h.store.SearchProducts(ctx, query)

	if results == nil {
		results = []*models.Product{}
	}
	writeJSON(w, 200, results)
}

func (h *ProductHandler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	category := r.URL.Query().Get("category")

	var products []*models.Product
	var total int

	if category != "" {
		// TODO: Call h.store.GetByCategory(category, page, limit)
		products, total = h.store.GetByCategory(category, page, limit)
	} else {
		// TODO: Call h.store.GetAll(page, limit)
		products, total = h.store.GetAll(page, limit)
	}

	// TODO: Write JSON response with models.ListResponse{Data, Total, Page, Limit}
	writeJSON(w, 200, models.ListResponse{Data: products, Total: total, Page: page, Limit: limit})
}

// func (h *ProductHandler) handleListProducts(w http.ResponseWriter, r *http.Request) {
// 	// TODO: Write JSON response with models.ListResponse{Data, Total, Page, Limit}
// 	writeJSON(w, 200, h.store.List())
// }

// GetProduct handles GET /api/products/{id}
func (h *ProductHandler) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "ID must be a number")
		return
	}

	p, ok := h.cache.Get(id)
	if ok {
		log.Printf("[CACHE HIT] product %d\n", id)
		writeJSON(w, 200, p)
		return
	}
	log.Printf("[CACHE MISS] product %d\n", id)
	// TODO: h.store.GetByID(id)
	p, ok = h.store.GetByID(id)
	// TODO: If not found: 404 with error "product_not_found"
	if !ok {
		writeError(w, 404, "product_not_found", "Product not found")
		return
	}
	// TODO: If found: 200 with product JSON
	h.cache.Set(id, p)
	writeJSON(w, 200, p)
}

// CreateProduct handles POST /api/products
func (h *ProductHandler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return
	}

	// TODO: Validate req; if errors, return 400 with writeValidationError
	errors := req.Validate()
	if len(errors) > 0 {
		writeValidationError(w, errors)
		return
	}
	// TODO: Create product: product := h.store.Create(req)
	product := h.store.Create(req)
	h.cache.Set(product.ID, product)
	// TODO: Set Location header: w.Header().Set("Location", fmt.Sprintf("/api/products/%d", product.ID))
	w.Header().Set("Location", fmt.Sprintf("/api/products/%d", product.ID))
	// TODO: Write 201 response with product
	writeJSON(w, 201, product)
}

// UpdateProduct handles PATCH /api/products/{id}
func (h *ProductHandler) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "ID must be a number")
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return
	}

	// TODO: h.store.Update(id, req)
	p, ok := h.store.Update(id, req)
	// TODO: If not found: 404
	if !ok {
		writeError(w, 404, "product_not_found", "Product not found")
		return
	}
	// TODO: If found: 200 with updated product
	h.cache.Set(id, p)
	writeJSON(w, 200, p)
}

// DeleteProduct handles DELETE /api/products/{id}
func (h *ProductHandler) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "ID must be a number")
		return
	}

	// TODO: h.store.Delete(id)
	ok := h.store.Delete(id)
	// TODO: If not found: 404
	if !ok {
		writeError(w, 404, "product_not_found", "Product not found")
		return
	}
	h.cache.Invalidate(id)
	// TODO: If deleted: 204 No Content (w.WriteHeader(http.StatusNoContent))
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) handleReserveStock(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "ID must be a number")
	}
	var req struct {
		Delta int `json:"delta"` // negative to reserve, positive to release
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	qty := -req.Delta // delta is negative for reservation, so negate for reserveStock
	if req.Delta < 0 {
		// Reserving stock
		p, err := h.store.ReserveStock(id, qty)
		if err != nil {
			writeJSON(w, 409, map[string]string{"error": err.Error()})
			return
		}
		h.cache.Invalidate(id)
		writeJSON(w, 200, p)
	} else {
		// TODO: Releasing stock (return items) — implement similarly
		p, err := h.store.ReleaseStock(id, qty)
		if err != nil {
			writeJSON(w, 409, map[string]string{"error": err.Error()})
			return
		}
		h.cache.Invalidate(id)
		writeJSON(w, 200, p)
	}
}
