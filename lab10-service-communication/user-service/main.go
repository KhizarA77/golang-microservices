package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// User represents a user in this service.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserStore is a thread-safe in-memory user store.
type UserStore struct {
	mu     sync.RWMutex
	users  map[string]*User
	nextID int
}

func NewUserStore() *UserStore {
	s := &UserStore{users: make(map[string]*User)}
	// Seed with sample data
	s.create(User{Name: "Alice Smith", Email: "alice@example.com"})
	s.create(User{Name: "Bob Jones", Email: "bob@example.com"})
	s.create(User{Name: "Carol White", Email: "carol@example.com"})
	return s
}

func (s *UserStore) create(u User) *User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	u.ID = fmt.Sprintf("%d", s.nextID)
	u.CreatedAt = time.Now()
	cp := u
	s.users[u.ID] = &cp
	return &cp
}

func (s *UserStore) getByID(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (s *UserStore) list() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		result = append(result, &cp)
	}
	return result
}

// =============================================================================
// Handlers
// =============================================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

var store *UserStore

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "user-service",
		"version": "1.0.0",
	})
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Return all users as JSON array
	writeJSON(w, http.StatusOK, store.list())
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// TODO: store.getByID(id)
	user, ok := store.getByID(id)
	// TODO: If not found, return 404 with {"error": "user not found"}
	if !ok {
		writeJSON(w, 404, map[string]string{"error": "user not found"})
		return
	}
	// TODO: If found, return 200 with user JSON
	writeJSON(w, 200, user)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	defer r.Body.Close()

	// TODO: Validate name and email are not empty
	if input.Name == "" || input.Email == "" {
		writeJSON(w, 400, map[string]string{"error": "Invalid input"})
		return
	}
	// TODO: store.create(User{Name, Email})
	user := store.create(User{Name: input.Name, Email: input.Email})
	// TODO: Return 201 with new user
	writeJSON(w, 201, user)
}

func main() {
	store = NewUserStore()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /users", handleListUsers)
	mux.HandleFunc("GET /users/{id}", handleGetUser)
	mux.HandleFunc("POST /users", handleCreateUser)

	fmt.Println("User Service running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
