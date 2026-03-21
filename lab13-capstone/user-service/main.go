package main

// User Service — Port 8081
//
// This service handles user registration, authentication, and profile management.
//
// INSTRUCTIONS:
// Port your Lab 09 Clean Architecture solution here.
// Additions needed for this capstone:
//   1. Login endpoint that returns a signed token string
//   2. Token format: base64("userID:email:timestamp") — simple, no crypto needed for lab
//   3. GET /users/me endpoint using token to identify caller
//
// Endpoints:
//   POST /api/users/register  → {"name":"Alice","email":"...","password":"..."}
//   POST /api/users/login     → returns {"token":"...","user":{...}}
//   GET  /api/users           → list (no auth needed for lab)
//   GET  /api/users/{id}      → get by ID
//   GET  /health
//
// TODO: Copy lab09 domain, usecase, repository, handler packages here.
//       Add login + token generation.
//       Add graceful shutdown.

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// SimpleUser is a minimal user for this scaffold.
// Replace with your Lab 09 domain.User when porting.
type SimpleUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	password  string    // unexported — not serialised
	CreatedAt time.Time `json:"created_at"`
}

// generateToken creates a simple non-secure token for the lab.
// Format: base64("id:email:timestamp")
func generateToken(userID, email string) string {
	raw := fmt.Sprintf("%s:%s:%d", userID, email, time.Now().Unix())
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// =============================================================================
// In-memory store (replace with Lab 09 repository)
// =============================================================================

type store struct {
	mu     sync.RWMutex
	users  map[string]*SimpleUser
	byEmail map[string]*SimpleUser
	nextID int
}

var globalStore = &store{
	users:   make(map[string]*SimpleUser),
	byEmail: make(map[string]*SimpleUser),
}

func (s *store) register(name, email, password string) (*SimpleUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byEmail[email]; exists {
		return nil, fmt.Errorf("email already registered")
	}
	s.nextID++
	u := &SimpleUser{
		ID:        fmt.Sprintf("%d", s.nextID),
		Name:      name,
		Email:     email,
		password:  fmt.Sprintf("hashed:%s", password),
		CreatedAt: time.Now(),
	}
	s.users[u.ID] = u
	s.byEmail[email] = u
	return u, nil
}

func (s *store) login(email, password string) (*SimpleUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byEmail[email]
	if !ok || u.password != fmt.Sprintf("hashed:%s", password) {
		return nil, fmt.Errorf("invalid credentials")
	}
	return u, nil
}

func (s *store) list() []*SimpleUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*SimpleUser, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		result = append(result, &cp)
	}
	return result
}

func (s *store) getByID(id string) (*SimpleUser, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

// =============================================================================
// Handlers
// =============================================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, 400, map[string]string{"error": "name, email, and password are required"})
		return
	}

	user, err := globalStore.register(req.Name, req.Email, req.Password)
	if err != nil {
		writeJSON(w, 409, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, user)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	user, err := globalStore.login(req.Email, req.Password)
	if err != nil {
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}

	token := generateToken(user.ID, user.Email)
	writeJSON(w, 200, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, globalStore.list())
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, ok := globalStore.getByID(id)
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, 200, user)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok", "service": "user-service"})
}

func main() {
	// Seed with a test user
	globalStore.register("Alice Smith", "alice@example.com", "password123")
	globalStore.register("Bob Jones", "bob@example.com", "password456")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/users/register", handleRegister)
	mux.HandleFunc("POST /api/users/login", handleLogin)
	mux.HandleFunc("GET /api/users", handleListUsers)
	mux.HandleFunc("GET /api/users/{id}", handleGetUser)

	srv := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("User Service on http://localhost:8081")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// TODO: When you have more time, port your Lab 09 implementation here
	// and replace the globalStore with clean architecture layers.

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("User Service shutting down...")
}
