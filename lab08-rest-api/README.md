# Lab 08 — REST API Design

**Level:** Intermediate
**Topic:** RESTful API, CRUD, In-Memory Store, Validation, Error Handling

---

## Background

### What is REST?

REST (Representational State Transfer) is an architectural style for APIs. Key constraints:

1. **Stateless** — every request contains all information needed; server holds no session
2. **Uniform Interface** — consistent resource identification and manipulation via HTTP methods
3. **Resource-based** — URL identifies a resource; HTTP method defines the action

### HTTP Methods and Their Semantics

| Method | Action | Idempotent? | Body? |
|--------|--------|-------------|-------|
| GET | Read | Yes | No |
| POST | Create | No | Yes |
| PUT | Replace entirely | Yes | Yes |
| PATCH | Partial update | No | Yes |
| DELETE | Delete | Yes | No |

### RESTful URL Design

```
GET    /products           → list all products
POST   /products           → create a product
GET    /products/{id}      → get product by ID
PUT    /products/{id}      → replace product
PATCH  /products/{id}      → partially update product
DELETE /products/{id}      → delete product
```

Sub-resources: `GET /products/{id}/reviews`

### HTTP Status Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 | OK | Successful GET, PUT, PATCH |
| 201 | Created | Successful POST (include `Location` header) |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid input |
| 401 | Unauthorized | Missing/invalid auth |
| 403 | Forbidden | Valid auth, but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Validation errors |
| 500 | Internal Server Error | Unexpected server error |

### Consistent Error Response Format

Pick a format and use it everywhere:

```json
{
  "error": "validation_failed",
  "message": "Name is required",
  "fields": {
    "name": "must not be empty",
    "price": "must be greater than 0"
  }
}
```

### In-Memory Store with Locking

For this lab, use a map protected by a `sync.RWMutex`:

```go
type Store struct {
    mu       sync.RWMutex
    products map[int]*Product
    nextID   int
}

func (s *Store) Get(id int) (*Product, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    p, ok := s.products[id]
    return p, ok
}

func (s *Store) Create(p *Product) *Product {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.nextID++
    p.ID = s.nextID
    s.products[p.ID] = p
    return p
}
```

### Pagination

For list endpoints, support `?page=1&limit=10` query parameters:

```go
page, _ := strconv.Atoi(r.URL.Query().Get("page"))
if page < 1 { page = 1 }
limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
if limit < 1 || limit > 100 { limit = 10 }

start := (page - 1) * limit
end := min(start + limit, len(items))
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Design clean, RESTful URL structures
- Implement a complete CRUD API with proper HTTP status codes
- Validate request bodies and return structured errors
- Build a thread-safe in-memory repository
- Support pagination and filtering on list endpoints
- Write integration-style tests using `httptest`

---

## Project Structure

```
lab08-rest-api/
├── go.mod
├── main.go
├── models/
│   └── product.go     ← Product struct and validation
├── store/
│   └── store.go       ← In-memory store
└── handlers/
    └── handlers.go    ← HTTP handler functions
```

---

## Tasks

### Task 1 — Product Model and Validation

In `models/product.go`, define:

```go
type Product struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Stock       int       `json:"stock"`
    Category    string    `json:"category"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    Stock       int     `json:"stock"`
    Category    string  `json:"category"`
}

type UpdateProductRequest struct {
    Name        *string  `json:"name"`        // pointer = optional
    Description *string  `json:"description"`
    Price       *float64 `json:"price"`
    Stock       *int     `json:"stock"`
    Category    *string  `json:"category"`
}
```

Implement `(r CreateProductRequest) Validate() map[string]string` that returns a map of field name → error message for:
- `name` must not be empty
- `price` must be > 0
- `stock` must be >= 0
- `category` must be one of: "electronics", "clothing", "food", "books"

### Task 2 — In-Memory Store

In `store/store.go`, implement a `ProductStore` with:
- `GetAll(page, limit int) ([]*Product, int)` — paginated list, returns items and total count
- `GetByID(id int) (*Product, bool)`
- `GetByCategory(category string, page, limit int) ([]*Product, int)` — filter by category
- `Create(req CreateProductRequest) *Product`
- `Update(id int, req UpdateProductRequest) (*Product, bool)` — partial update (only update non-nil fields)
- `Delete(id int) bool`

Seed with 5 sample products in the constructor.

### Task 3 — HTTP Handlers

In `handlers/handlers.go`, implement all CRUD handlers:

| Handler | Method | Route |
|---------|--------|-------|
| `ListProducts` | GET | `/api/products` |
| `GetProduct` | GET | `/api/products/{id}` |
| `CreateProduct` | POST | `/api/products` |
| `UpdateProduct` | PATCH | `/api/products/{id}` |
| `DeleteProduct` | DELETE | `/api/products/{id}` |

Response envelope for list:
```json
{
  "data": [...],
  "total": 10,
  "page": 1,
  "limit": 10
}
```

Support `?category=electronics` filtering on the list endpoint.

### Task 4 — Wire Everything in `main.go`

- Create the store (seeded with products)
- Create handlers
- Register routes on `http.NewServeMux()`
- Add logging middleware
- Start server on port 8080

### Task 5 — Test with `httptest`

In `handlers/handlers_test.go`, write tests for:
- GET /api/products → 200 with list
- POST /api/products with valid body → 201
- POST /api/products with invalid body → 400 with validation errors
- GET /api/products/{id} for existing id → 200
- GET /api/products/{id} for non-existent id → 404
- DELETE /api/products/{id} → 204

Use `httptest.NewRecorder()` and `httptest.NewServer()`.

---

## Tips

- `PATCH` means partial update — only update fields that are present (non-nil pointers let you tell the difference between "not provided" and "set to zero value").
- Always call `defer r.Body.Close()` after reading request body.
- Set `Location: /api/products/{id}` header in the 201 response.
- For test isolation, each test should create its own fresh store instance.
- `json.NewDecoder(r.Body).Decode(&input)` returns an error on invalid JSON — always check it.

---

## Running Your Solution

```bash
cd lab08-rest-api
go run .

# Test endpoints
curl http://localhost:8080/api/products
curl http://localhost:8080/api/products/1
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":999.99,"stock":10,"category":"electronics"}'
curl -X DELETE http://localhost:8080/api/products/1
```

Run tests:
```bash
go test ./handlers/... -v
```
