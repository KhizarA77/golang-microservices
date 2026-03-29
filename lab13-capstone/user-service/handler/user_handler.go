package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"user-service/domain"
	"user-service/usecase"
)

// UserHandler handles HTTP requests for the user API.
type UserHandler struct {
	uc *usecase.UserUseCase
}

// NewUserHandler creates a new handler with the given use case.
func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

// RegisterRoutes registers all user routes on the given mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users/register", h.Register)
	mux.HandleFunc("POST /api/users/login", h.Login)
	mux.HandleFunc("GET /api/users", h.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", h.GetUser)
	mux.HandleFunc("PUT /api/users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", h.DeleteUser)
	mux.HandleFunc("GET /health", h.Health)
}

// =============================================================================
// Helpers
// =============================================================================

func generateToken(userID, email string) string {
	raw := fmt.Sprintf("%s:%s:%d", userID, email, time.Now().Unix())
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// mapDomainError translates domain errors to HTTP status codes.
func mapDomainError(err error) (int, string) {
	var notFound domain.ErrUserNotFound
	var alreadyExists domain.ErrEmailAlreadyExists
	var validation domain.ErrValidation

	switch {
	case errors.As(err, &notFound):
		return http.StatusNotFound, err.Error()
	case errors.As(err, &alreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, err.Error()
	case errors.As(err, &validation):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// =============================================================================
// Handlers
// =============================================================================

func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok", "service": "user-service"})
}

// Register handles POST /api/users
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	// TODO: Call h.uc.Register(req.Name, req.Email, req.Password)
	user, _err := h.uc.Register(req.Name, req.Email, req.Password)
	// TODO: On error: status, msg := mapDomainError(err); writeJSON(w, status, map...)
	if _err != nil {
		status, msg := mapDomainError(_err)
		writeJSON(w, status, msg)
		return
	}
	// TODO: On success: writeJSON(w, 201, user)
	user.Password = ""
	writeJSON(w, 201, user)
}

// Login handles POST /api/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	// TODO: Call h.uc.Login(req.Email, req.Password)
	user, _err := h.uc.Login(req.Email, req.Password)
	if _err != nil {
		status, msg := mapDomainError(_err)
		writeJSON(w, status, msg)
		return
	}
	// TODO: Map error → status or return 200 with user
	token := generateToken(user.ID, user.Email)
	user.Password = ""
	writeJSON(w, 200, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// ListUsers handles GET /api/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Call h.uc.ListUsers()
	sli, err := h.uc.ListUsers()
	if err != nil {
		status, msg := mapDomainError(err)
		writeJSON(w, status, msg)
		return
	}
	for i := range sli {
		sli[i].Password = ""
	}
	// TODO: Return JSON array of users
	writeJSON(w, 200, sli)
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// TODO: Call h.uc.GetUser(id)
	user, err := h.uc.GetUser(id)
	// TODO: Map ErrUserNotFound → 404 or return 200 with user
	if err != nil {
		status, msg := mapDomainError(err)
		writeJSON(w, status, msg)
		return
	}
	user.Password = ""
	writeJSON(w, 200, user)
}

// UpdateUser handles PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	// TODO: Call h.uc.UpdateUser(id, req.Name)
	user, err := h.uc.UpdateUser(id, req.Name)
	// TODO: Map error or return 200 with updated user
	if err != nil {
		status, msg := mapDomainError(err)
		writeJSON(w, status, msg)
		return
	}
	user.Password = ""
	writeJSON(w, 200, user)
}

// DeleteUser handles DELETE /api/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// TODO: Call h.uc.DeleteUser(id)
	// TODO: Map error or return 204
	err := h.uc.DeleteUser(id)
	if err != nil {
		status, msg := mapDomainError(err)
		writeJSON(w, status, msg)
		return
	}
	writeJSON(w, 200, "Deleted Successfully")
}
