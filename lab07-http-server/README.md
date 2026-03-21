# Lab 07 — HTTP Server Basics

**Level:** Beginner–Intermediate
**Topic:** `net/http`, Routing, Middleware, JSON Handling

---

## Background

### Go's `net/http` Package

Go ships with a production-quality HTTP server in the standard library. You don't need a framework for basic services.

```go
// Simplest server
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, World!")
})
http.ListenAndServe(":8080", nil)
```

### `http.Handler` Interface

Everything in Go's HTTP server revolves around one interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

A `HandlerFunc` is a function type that also implements `Handler`:
```go
type HandlerFunc func(ResponseWriter, *Request)
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
```

This lets plain functions be used as handlers.

### `http.ServeMux` — Default Router

`http.ServeMux` is Go's built-in request multiplexer (router). It matches request paths to handler functions.

```go
mux := http.NewServeMux()
mux.HandleFunc("/users", usersHandler)
mux.HandleFunc("/users/", specificUserHandler) // trailing slash = subtree
```

As of Go 1.22+, the default mux supports method-based routing:
```go
mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("GET /users/{id}", getUser)  // path parameters
```

### Middleware

Middleware is a function that wraps a handler, adding behavior before and/or after the handler runs.

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// Usage
handler := LoggingMiddleware(mux)
http.ListenAndServe(":8080", handler)
```

Middleware chains are composable:
```go
handler := LoggingMiddleware(AuthMiddleware(RateLimitMiddleware(mux)))
```

### Writing JSON Responses

```go
import "encoding/json"

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    user := User{ID: 1, Name: "Alice"}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(user)
}
```

### Reading JSON Request Bodies

```go
func handler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    // use input.Name
}
```

### Common Response Helpers

```go
// Error response
http.Error(w, "not found", http.StatusNotFound)

// Redirect
http.Redirect(w, r, "/new-path", http.StatusMovedPermanently)

// Status only
w.WriteHeader(http.StatusNoContent)
```

### Graceful Shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// Wait for interrupt
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Start an HTTP server with `net/http`
- Register handlers for specific routes
- Read path parameters and query strings
- Parse and write JSON request/response bodies
- Write reusable middleware (logging, auth, CORS)
- Use `http.Server` for configurable server settings and graceful shutdown

---

## Tasks

### Task 1 — Hello Server

Create a server on port 8080 with these routes:

| Method | Path | Response |
|--------|------|----------|
| GET | `/` | `{"message": "Hello, Go!"}` |
| GET | `/health` | `{"status": "ok", "time": "<current time>"}` |
| GET | `/echo` | Echo back any query parameters as JSON |

Test with:
```bash
curl http://localhost:8080/
curl http://localhost:8080/health
curl "http://localhost:8080/echo?name=Alice&lang=Go"
```

### Task 2 — Middleware Chain

Write three middleware functions:

1. `LoggingMiddleware` — logs `"METHOD PATH STATUS DURATION"` for every request
2. `RequestIDMiddleware` — generates a UUID-like request ID, sets `X-Request-ID` header on response, adds it to request context
3. `RecoveryMiddleware` — catches panics, logs them, returns 500

Add a route `/panic` that intentionally panics to test the recovery middleware.

Chain all middleware around your mux.

### Task 3 — JSON API Routes

Add these routes (all return/accept JSON):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/users` | Return a hardcoded list of 3 users |
| POST | `/api/users` | Accept a user JSON body, validate name is non-empty, return 201 with the created user |
| GET | `/api/users/{id}` | Return user by ID (hardcoded), 404 if not found |

User struct: `ID int`, `Name string`, `Email string`

Validation: if `Name` is empty, return `{"error": "name is required"}` with status 400.

### Task 4 — Custom `http.Server` with Graceful Shutdown

Instead of `http.ListenAndServe`, create an `http.Server` with:
- `ReadTimeout: 5s`
- `WriteTimeout: 10s`
- `IdleTimeout: 120s`
- `MaxHeaderBytes: 1 << 20`

Implement graceful shutdown: listen for `SIGINT`/`SIGTERM`, call `server.Shutdown` with a 10-second context, print `"Server stopped gracefully"`.

---

## Tips

- Go 1.22+ `ServeMux` supports `{id}` wildcards — use `r.PathValue("id")` to extract them.
- Always set `Content-Type: application/json` before writing the body.
- Use `defer r.Body.Close()` after reading the body.
- `http.Error` writes both the body and status code — it's a convenience shortcut for simple error responses.
- Middleware order matters: the outermost middleware runs first on request and last on response.

---

## Running Your Solution

```bash
cd lab07-http-server
go run .
# In another terminal:
curl http://localhost:8080/health
```
