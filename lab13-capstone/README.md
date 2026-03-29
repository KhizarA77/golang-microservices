# Lab 13 — Capstone: Microservices E-Commerce Platform

**Level:** Expert
**Topic:** Integrating Concurrency + Design Patterns + Clean Architecture + Microservices + Event-Driven

---

## Overview

This capstone integrates everything you've learned into a realistic (simplified) e-commerce platform. You'll build five services that work together:

```
                        ┌─────────────────┐
                        │   API Gateway   │ :8080
                        └────────┬────────┘
                                 │ routes requests
              ┌──────────────────┼──────────────────┐
              │                  │                  │
     ┌────────▼──────┐  ┌────────▼──────┐  ┌───────▼────────┐
     │ User Service  │  │Product Service│  │ Order Service  │
     │    :8081      │  │    :8082      │  │    :8083       │
     └───────────────┘  └───────────────┘  └────────┬───────┘
                                                     │ publishes events
                                           ┌─────────▼──────────┐
                                           │Notification Service│
                                           │     :8084          │
                                           └────────────────────┘
```

---

## What Each Service Does

### API Gateway (:8080)
- Single entry point for clients
- Routes `/api/users/*` → User Service
- Routes `/api/products/*` → Product Service
- Routes `/api/orders/*` → Order Service
- JWT authentication middleware (validates token, passes user ID downstream)
- Rate limiting middleware (5 requests/second per IP)
- Request ID injection
- Logging

### User Service (:8081)
- Register, login (returns JWT token), get profile, update profile
- Clean architecture: domain → usecase → repository → handler
- In-memory repository

### Product Service (:8082)
- Product catalog: CRUD for products
- Concurrent search (fan-out to multiple "data sources")
- In-memory repository
- Caching proxy pattern (cache GET responses for 30s)

### Order Service (:8083)
- Place orders (validates product stock via Product Service)
- Order state machine (Pending → Processing → Shipped → Delivered)
- Domain events published to shared event bus
- Worker pool for processing orders concurrently
- Clean architecture

### Notification Service (:8084)
- Subscribes to order events
- Simulates sending email/SMS notifications
- Uses goroutines for async processing
- Rate limited to prevent notification spam

---

## Concepts Applied

| Concept | Where Used |
|---------|-----------|
| **Goroutines & Channels** | Worker pool in Order Service; async notifications |
| **Context & Cancellation** | All HTTP clients; graceful shutdown |
| **Worker Pool** | Order processing queue in Order Service |
| **Pipeline** | Product search fan-out/fan-in |
| **Singleton** | Shared event bus instance |
| **Factory** | Repository factory (memory/future DB) |
| **Builder** | HTTP request builder for service clients |
| **Functional Options** | Server configuration |
| **Adapter** | Service clients adapt HTTP responses to domain types |
| **Decorator/Middleware** | Auth, logging, rate limiting in API Gateway |
| **Proxy/Cache** | Product service response caching |
| **Observer/Event Bus** | Order events → Notification Service |
| **Strategy** | Pluggable order pricing strategy |
| **State Machine** | Order status transitions |
| **Clean Architecture** | User Service, Order Service layer structure |
| **Repository Pattern** | All services |
| **REST API** | All services |
| **Service Communication** | API Gateway → services, Order → Product |
| **Circuit Breaker** | Order Service → Product Service calls |
| **Event-Driven** | Order events to Notification Service |

---

## Project Structure

```
lab13-capstone/
├── README.md
├── shared/                           ← Shared types and utilities
│   ├── go.mod
│   ├── events/events.go              ← Event type constants and payloads
│   ├── middleware/middleware.go       ← Shared HTTP middleware
│   └── client/client.go             ← Reusable HTTP client with retry
│
├── api-gateway/
│   ├── go.mod
│   └── main.go
│
├── user-service/
│   ├── go.mod
│   ├── domain/user.go
│   ├── usecase/user_usecase.go
│   ├── repository/memory.go
│   ├── handler/handler.go
│   └── main.go
│
├── product-service/
│   ├── go.mod
│   ├── domain/product.go
│   ├── usecase/product_usecase.go
│   ├── repository/memory.go
│   ├── cache/proxy.go
│   ├── handler/handler.go
│   └── main.go
│
├── order-service/
│   ├── go.mod
│   ├── domain/order.go             ← Order aggregate + state machine
│   ├── usecase/order_usecase.go
│   ├── repository/memory.go
│   ├── worker/pool.go              ← Worker pool for order processing
│   ├── handler/handler.go
│   └── main.go
│
└── notification-service/
    ├── go.mod
    └── main.go
```

---

## Implementation Guide

### Step 1 — Start Simple: User Service

Copy your solution from Lab 09 (Clean Architecture) into `user-service/`. Add a login endpoint that issues a simple JWT (or a signed token string — you can use `golang-jwt/jwt` or simulate with `base64`).

**Login response:**
```json
{
  "token": "eyJ...",
  "user": { "id": "...", "name": "...", "email": "..." }
}
```

### Step 2 — Product Service

Adapt Lab 08 (REST API) into `product-service/`. Add:
- **Caching proxy** from Lab 05: cache `GetProduct` for 30 seconds
- **Concurrent search**: `SearchProducts(query string)` fans out to 3 "data sources" simultaneously, merges results (Lab 03 fan-out/fan-in)

```go
func (s *ProductUseCase) SearchProducts(ctx context.Context, query string) []*Product {
    // Fan out to 3 search strategies concurrently
    // 1. Name search
    // 2. Category search
    // 3. Description search
    // Fan in and deduplicate
}
```

### Step 3 — Order Service

This is the most complex service. Combines:
- **Clean architecture** (domain/usecase/repository/handler)
- **State machine** (Pending → Processing → Shipped → Delivered — from Lab 06)
- **Worker pool** (Lab 03): orders are placed in a queue, a pool of 3 workers processes them
- **Service client** (Lab 10): validates product exists and has stock before placing order
- **Event publishing**: publishes `order.placed`, `order.shipped`, `order.cancelled`

**Order processing flow:**
```
POST /orders
  │
  ├── Validate: product exists (call Product Service)
  ├── Validate: sufficient stock
  ├── Create order in Pending state
  ├── Put in processing queue (buffered channel)
  └── Return 202 Accepted

Worker goroutine:
  ├── Receive order from queue
  ├── Transition to Processing
  ├── Simulate payment processing (200ms)
  ├── Transition to Shipped (or Cancelled if payment fails)
  └── Publish event to event bus
```

### Step 4 — Notification Service

Subscribes to the shared in-process event bus (or HTTP webhooks if you want to simulate a real message broker):
- `order.placed` → "Email: Your order <id> has been placed"
- `order.shipped` → "SMS: Your order <id> has shipped! Track: <number>"
- `order.cancelled` → "Email: Your order <id> has been cancelled"

Use a worker pool with 2 notification workers to process notifications asynchronously.

### Step 5 — API Gateway

The gateway is a reverse proxy with middleware:

```go
type Gateway struct {
    userServiceURL    string
    productServiceURL string
    orderServiceURL   string
    rateLimiter       map[string]*RateLimiter  // per-IP
}

func (g *Gateway) proxyRequest(targetURL string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Build outbound request to target service
        // Copy headers, body
        // Forward request
        // Copy response back
    }
}
```

Middleware stack (outermost to innermost):
1. `RecoveryMiddleware` — catch panics
2. `LoggingMiddleware` — log all requests
3. `RequestIDMiddleware` — inject X-Request-ID
4. `RateLimitMiddleware` — 5 req/sec per IP (use time.Ticker)
5. `AuthMiddleware` — validate JWT for protected routes (skip `/api/users/register`, `/api/users/login`)

### Step 6 — Graceful Shutdown

All services must handle `SIGINT`/`SIGTERM`:
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
// Give in-flight requests 10 seconds to complete
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
server.Shutdown(ctx)
```

---

## Frontend Testing Dashboard

A browser-based UI is included at `frontend/index.html` for testing all services interactively — no npm or build step required.

```
lab13-capstone/
└── frontend/
    └── index.html   ← open directly in a browser
```

**Features:**

| Tab | What you can do |
|-----|-----------------|
| **Health** | Live up/down status for all 5 services (routed via gateway) |
| **Auth** | Register, login, logout — token auto-saved to `localStorage` |
| **Products** | List (paginated), search, get by ID, create, update, delete |
| **Orders** | Place an order, list all orders with status badges, get by ID |
| **Users** | List, get by ID, update name, delete |

All requests go through the API Gateway on `:8080`. The gateway has CORS enabled so the browser can call it freely. Open `frontend/index.html` in any browser after starting the services.

---

## Running the Capstone

```bash
# Start services (each in a separate terminal)
cd lab13-capstone/user-service    && go run .  # :8081
cd lab13-capstone/product-service && go run .  # :8082
cd lab13-capstone/order-service   && go run .  # :8083
cd lab13-capstone/notification-service && go run .  # :8084
cd lab13-capstone/api-gateway     && go run .  # :8080

# Or use a script to start all:
./start-all.sh
```

```bash
# Register and get a token
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret"}'

TOKEN=$(curl -s -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret"}' | jq -r '.token')

# List products
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/products

# Place an order
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"productId":"1","quantity":2}'

# List orders
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/orders
```

---

## Acceptance Criteria

- [ ] All 5 services start without errors
- [ ] User can register and login, receive a token
- [ ] Token is required for protected endpoints
- [ ] Products can be listed and viewed
- [ ] Orders can be placed; product stock is validated
- [ ] Notification service logs events when orders are processed
- [ ] API Gateway routes correctly and logs all requests
- [ ] Pressing Ctrl+C on any service shuts it down gracefully
- [ ] No data races (`go run -race .`)

---

## Stretch Goals

- Add a `docker-compose.yml` to start all services
- Implement real JWT signing with `golang-jwt/jwt`
- Add pagination to all list endpoints
- Implement an order search endpoint using fan-out
- Add Prometheus metrics middleware
- Write integration tests using `httptest`

---

## Congratulations

If you've made it here and have a working capstone, you've covered:
- Advanced Go concurrency (goroutines, channels, select, context, worker pools, pipelines)
- Design patterns (Singleton, Factory, Builder, Adapter, Decorator, Proxy, Observer, Strategy, State)
- Backend development (REST, middleware, clean architecture, testing)
- Microservices (service communication, circuit breakers, event-driven, API gateway)

You're ready to build production Go services.
