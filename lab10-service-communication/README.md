# Lab 10 — Service-to-Service Communication

**Level:** Intermediate–Advanced
**Topic:** HTTP Microservices, Service Clients, Retry Logic, Health Checks, Circuit Breaker

---

## Background

### Microservices vs Monolith

A **monolith** is a single deployable unit where all functionality shares memory and function calls. A **microservice** is an independently deployable service that owns its own data and communicates over the network.

**Trade-offs:**
| Concern | Monolith | Microservices |
|---------|----------|---------------|
| Deployment | Simple | Complex |
| Scaling | Scale all or nothing | Scale individual services |
| Failure isolation | One failure can crash all | Failures are isolated |
| Network latency | None | Present |
| Data consistency | Easy (shared DB) | Hard (each owns its data) |

---

### Service Discovery

In production, services find each other via:
- **DNS** — service name resolves to an IP (e.g., Kubernetes services)
- **Service registry** — services register themselves (Consul, etcd)
- **Environment variables** — simple for dev: `USER_SERVICE_URL=http://user-svc:8081`

In this lab, we'll use hardcoded URLs and environment variables.

---

### HTTP Client Best Practices

Always configure timeouts on your HTTP client. The default Go client has no timeout:

```go
client := &http.Client{
    Timeout: 5 * time.Second,  // total request timeout
}
```

Use a shared client — creating a new one per request is wasteful (loses connection pooling).

---

### Retry with Exponential Backoff

Transient failures (timeouts, 503) are expected in distributed systems. Retry with backoff:

```go
func retryDo(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(attempt) * 200 * time.Millisecond
            time.Sleep(backoff)
        }
        resp, err := client.Do(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        if err != nil { lastErr = err }
    }
    return nil, lastErr
}
```

**Note:** Only retry idempotent operations (GET, DELETE, PUT). Never blindly retry POST.

---

### Circuit Breaker Pattern

A circuit breaker tracks failures and "opens" when failures exceed a threshold, returning errors immediately without trying the remote service. After a timeout, it "half-opens" to try again.

```
Closed ──(failures exceed threshold)──► Open ──(timeout)──► Half-Open
  ▲                                                               │
  └──────────────────────(success)────────────────────────────────┘
```

States:
- **Closed** — normal operation, requests pass through
- **Open** — circuit tripped, return error immediately (fail fast)
- **Half-Open** — testing recovery, let one request through

---

### Health Check Endpoints

Every service should expose `GET /health`:
```json
{
  "status": "ok",
  "service": "user-service",
  "version": "1.0.0",
  "dependencies": {
    "database": "ok"
  }
}
```

This allows load balancers and orchestrators (Kubernetes) to know if a service is ready to receive traffic.

---

## Project Structure

```
lab10-service-communication/
├── user-service/
│   ├── go.mod
│   └── main.go          ← Runs on :8081
├── order-service/
│   ├── go.mod
│   └── main.go          ← Runs on :8082, calls user-service
└── shared/
    └── client/
        └── client.go    ← Reusable HTTP client with retry
```

---

## Service Design

**User Service (port 8081)**:
```
GET  /health
GET  /users/{id}      → return user details
GET  /users           → list users
POST /users           → create user
```

**Order Service (port 8082)**:
```
GET  /health
GET  /orders          → list orders (enriched with user data from user-service)
POST /orders          → create order (validates user exists via user-service)
GET  /orders/{id}     → get order
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Run multiple Go services on different ports
- Build a typed HTTP client for service-to-service calls
- Implement retry logic with exponential backoff
- Handle service-down scenarios gracefully
- Implement a simple circuit breaker
- Write health check endpoints

---

## Tasks

### Task 1 — User Service

Build the user service on port 8081:
- In-memory user store with 3 seeded users
- `GET /health` — returns `{"status":"ok","service":"user-service"}`
- `GET /users` — list all users
- `GET /users/{id}` — get by ID, 404 if not found
- `POST /users` — create user, 201

### Task 2 — User Service Client

In `order-service/main.go` (or a `client` package), build a `UserServiceClient`:

```go
type UserServiceClient struct {
    baseURL string
    client  *http.Client
}

func (c *UserServiceClient) GetUser(ctx context.Context, id string) (*User, error)
func (c *UserServiceClient) ListUsers(ctx context.Context) ([]*User, error)
```

Use `context.Context` for request propagation. Respect context cancellation.

### Task 3 — Order Service

Build the order service on port 8082:
- In-memory order store
- `GET /health` — also checks if user-service is healthy
- `POST /orders` — body: `{userId, productName, quantity, price}`
  - Validate the user exists by calling user service
  - If user service is down, return 503 with `{"error": "user service unavailable"}`
- `GET /orders` — list all orders, enrich with user data
- `GET /orders/{id}` — get single order enriched with user data

### Task 4 — Retry Logic

Add a `retryGet(ctx context.Context, url string, maxRetries int)` function that:
- Retries on network errors or 5xx responses
- Uses exponential backoff: `100ms, 200ms, 400ms, ...`
- Respects context cancellation (don't retry if context is done)
- Logs each retry attempt

Test it by temporarily returning 503 from user service every other request.

### Task 5 — Simple Circuit Breaker

Implement a `CircuitBreaker` struct:
```go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    failures     int
    lastFailure  time.Time
    state        string  // "closed", "open", "half-open"
    mu           sync.Mutex
}
```

Methods:
- `Allow() bool` — returns false if circuit is open
- `Success()` — record success, close circuit if half-open
- `Failure()` — increment failures, open circuit at threshold

Wrap the user service client calls with the circuit breaker.

---

## Running the Lab

```bash
# Terminal 1
cd lab10-service-communication/user-service
go run .

# Terminal 2
cd lab10-service-communication/order-service
go run .

# Test
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8082/orders
curl -X POST http://localhost:8082/orders \
  -H "Content-Type: application/json" \
  -d '{"userId":"1","productName":"Laptop","quantity":1,"price":999.99}'
```
