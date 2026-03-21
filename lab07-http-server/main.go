package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// =============================================================================
// Models
// =============================================================================

// User represents an application user.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Hardcoded user store for this lab
var users = []User{
	{ID: 1, Name: "Alice Smith", Email: "alice@example.com"},
	{ID: 2, Name: "Bob Jones", Email: "bob@example.com"},
	{ID: 3, Name: "Carol White", Email: "carol@example.com"},
}

// =============================================================================
// Middleware
// =============================================================================

// LoggingMiddleware logs method, path, status code, and duration.
// Hint: use a custom ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		// TODO: Call next.ServeHTTP(rec, r)
		next.ServeHTTP(rec, r)
		// TODO: Log: log.Printf("METHOD PATH STATUS DURATION")
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// RequestIDMiddleware generates a request ID and attaches it to the response header.
type contextKey string

const requestIDKey contextKey = "requestID"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Generate a simple request ID: fmt.Sprintf("%d", time.Now().UnixNano())
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		// TODO: Set w.Header().Set("X-Request-ID", id)
		w.Header().Set("X-Request-ID", requestID)
		// TODO: Store in context: ctx = context.WithValue(r.Context(), requestIDKey, id)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		// TODO: Call next.ServeHTTP(w, r.WithContext(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware catches panics, logs them, returns 500.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				// TODO: Log the panic: log.Printf("[PANIC] %v", err)
				log.Printf("[PANIC] %v", err)
				// TODO: Respond with 500
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		// TODO: Call next.ServeHTTP(w, r)
		next.ServeHTTP(w, r)
	})
}

// =============================================================================
// Helper functions
// =============================================================================

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// TODO: json.NewEncoder(w).Encode(v)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// =============================================================================
// Task 1 — Basic Routes
// =============================================================================

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// TODO: Write JSON: {"message": "Hello, Go!"}
	writeJSON(w, 200, map[string]string{
		"message": "Hello, Go!",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	// TODO: Write JSON: {"status": "ok", "time": time.Now().Format(time.RFC3339)}
	writeJSON(w, 200, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse r.URL.Query() into a map[string]string
	// TODO: Write it as JSON
	// Hint: r.URL.Query() returns url.Values (map[string][]string)
	//       Convert to map[string]string by taking the first value of each key
	query := r.URL.Query()
	params := make(map[string]string)
	for key, values := range query {
		params[key] = values[0]
	}
	writeJSON(w, 200, params)
}

// =============================================================================
// Task 3 — User API
// =============================================================================

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Write the users slice as JSON with status 200
	writeJSON(w, 200, users)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Decode JSON body into a User struct
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		writeError(w, 400, "Bad Request")
		return
	}
	// TODO: Validate: Name must not be empty → 400 with {"error": "name is required"}
	if user.Name == "" {
		writeError(w, 400, "name is required")
	}
	// TODO: Assign a new ID (len(users) + 1)
	user.ID = len(users) + 1
	// TODO: Append to users slice
	users = append(users, user)
	// TODO: Respond with 201 and the new user
	writeJSON(w, 201, user)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract path parameter "id" using r.PathValue("id")
	path := r.PathValue("id")
	// TODO: Convert to int with strconv.Atoi
	id, err := strconv.Atoi(path)
	if err != nil {
		writeError(w, 400, "Bad request")
		return
	}
	// TODO: Find user in users slice by ID
	for _, el := range users {
		if el.ID == id {
			writeJSON(w, 200, el)
			return
		}
	}
	// TODO: If not found: 404 with {"error": "user not found"}
	writeError(w, 404, "user not found")
}

// Panic handler for testing RecoveryMiddleware
func handlePanic(w http.ResponseWriter, r *http.Request) {
	panic("intentional panic for testing!")
}

// =============================================================================
// Task 4 — Server Setup with Graceful Shutdown
// =============================================================================

func buildRouter() http.Handler {
	mux := http.NewServeMux()

	// Task 1 routes
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /echo", handleEcho)

	// Task 2 — panic route
	mux.HandleFunc("GET /panic", handlePanic)

	// Task 3 routes
	mux.HandleFunc("GET /api/users", handleListUsers)
	mux.HandleFunc("POST /api/users", handleCreateUser)
	mux.HandleFunc("GET /api/users/{id}", handleGetUser)

	// TODO: Wrap mux with middleware chain:
	// RecoveryMiddleware -> RequestIDMiddleware -> LoggingMiddleware -> mux
	// (outermost is RecoveryMiddleware so it catches panics in other middleware too)
	var handler http.Handler = mux
	handler = LoggingMiddleware(handler)
	handler = RequestIDMiddleware(handler)
	handler = RecoveryMiddleware(handler)

	return handler
}

func main() {
	handler := buildRouter()

	// Task 4: Custom server with timeouts
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
		// TODO: Set ReadTimeout, WriteTimeout, IdleTimeout, MaxHeaderBytes
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 3000,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Server starting on http://localhost:8080")
		fmt.Println("Try: curl http://localhost:8080/health")
		fmt.Println("     curl http://localhost:8080/api/users")
		fmt.Println("     curl -X POST http://localhost:8080/api/users -d '{\"name\":\"Dave\",\"email\":\"dave@example.com\"}'")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// TODO: Set up signal handling for SIGINT, SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")

	// TODO: Create context with 10-second timeout
	// TODO: Call srv.Shutdown(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	fmt.Println("Server stopped gracefully")
}
