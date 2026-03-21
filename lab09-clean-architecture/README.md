# Lab 09 — Clean Architecture

**Level:** Advanced
**Topic:** Layered Architecture, Repository Pattern, Use Cases, Dependency Injection

---

## Background

### What is Clean Architecture?

Clean Architecture (Robert C. Martin) organizes code in concentric layers where dependencies only point **inward**. The inner layers know nothing about the outer layers.

```
┌─────────────────────────────────────────────┐
│  Frameworks & Drivers (HTTP, DB, CLI)        │  ← outermost
│  ┌───────────────────────────────────────┐  │
│  │  Interface Adapters (Handlers, Repos) │  │
│  │  ┌─────────────────────────────────┐ │  │
│  │  │  Use Cases (Application Logic)  │ │  │
│  │  │  ┌───────────────────────────┐  │ │  │
│  │  │  │  Entities (Domain Models)  │  │ │  │
│  │  │  └───────────────────────────┘  │ │  │
│  │  └─────────────────────────────────┘ │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**The Dependency Rule:** Source code dependencies must point inward. Nothing in an inner circle can know about something in an outer circle.

---

### Layer Responsibilities

**Domain (Entities)**
- Pure business data structures and rules
- No imports from other project packages
- No framework dependencies

```go
// domain/user.go
type User struct {
    ID        string
    Email     string
    Name      string
    CreatedAt time.Time
}

func (u *User) IsValid() bool { return u.Email != "" && u.Name != "" }
```

**Use Cases (Application Business Rules)**
- Orchestrates the flow of data to/from entities
- Defines interfaces it needs from outer layers (ports)
- Knows about domain, not about HTTP or databases

```go
// usecase/user_usecase.go
type UserRepository interface {        // port (defined here, implemented outside)
    FindByID(id string) (*domain.User, error)
    Save(user *domain.User) error
    FindByEmail(email string) (*domain.User, error)
}

type UserUseCase struct {
    repo UserRepository
}

func (uc *UserUseCase) Register(name, email string) (*domain.User, error) {
    // business logic here
}
```

**Interface Adapters (Repository implementations, HTTP handlers)**
- Translates between use case interfaces and external formats
- Repository implementations (in-memory, SQL, etc.)
- HTTP handlers that call use cases

**Frameworks & Drivers**
- The outermost layer: `main.go`, web framework config, database drivers
- Wires everything together (dependency injection)

---

### Repository Pattern

The repository abstracts data persistence behind an interface. Use cases depend on the interface, not the implementation. This makes use cases testable without a database.

```go
// Interface (in usecase layer)
type UserRepository interface {
    FindByID(id string) (*domain.User, error)
    Save(u *domain.User) error
    Delete(id string) error
    List() ([]*domain.User, error)
}

// In-memory implementation (in repository layer)
type InMemoryUserRepository struct {
    mu    sync.RWMutex
    store map[string]*domain.User
}
```

---

### Dependency Injection

Instead of use cases creating their own dependencies, inject them via constructors:

```go
// BAD: use case creates its own repo
func NewUserUseCase() *UserUseCase {
    return &UserUseCase{repo: NewSQLRepo()} // coupled to SQL
}

// GOOD: inject the dependency
func NewUserUseCase(repo UserRepository) *UserUseCase {
    return &UserUseCase{repo: repo}
}
```

In `main.go` (the composition root):
```go
repo := repository.NewInMemoryUserRepository()
uc   := usecase.NewUserUseCase(repo)
h    := handler.NewUserHandler(uc)
```

---

### Errors in Clean Architecture

Define domain-specific error types so use cases communicate intent:

```go
type ErrNotFound struct { ID string }
func (e ErrNotFound) Error() string { return fmt.Sprintf("user %s not found", e.ID) }

type ErrAlreadyExists struct { Email string }
func (e ErrAlreadyExists) Error() string { return fmt.Sprintf("user with email %s already exists", e.Email) }
```

HTTP handlers translate these domain errors into appropriate HTTP status codes.

---

## Project Structure

```
lab09-clean-architecture/
├── go.mod
├── main.go                         ← composition root
├── domain/
│   └── user.go                     ← User entity, domain errors
├── usecase/
│   ├── user_repository.go          ← UserRepository interface (port)
│   └── user_usecase.go             ← Business logic
├── repository/
│   └── memory_user_repository.go   ← In-memory implementation
└── handler/
    └── user_handler.go             ← HTTP handlers
```

---

## Learning Objectives

By the end of this lab you will be able to:

- Organize code into domain, use case, repository, and handler layers
- Define interfaces (ports) at the use case layer
- Implement the repository pattern
- Apply dependency injection via constructors
- Translate domain errors into HTTP status codes
- Test use cases in isolation (no HTTP, no database)

---

## Tasks

### Task 1 — Domain Layer (`domain/user.go`)

Define the `User` entity:
```go
type User struct {
    ID        string
    Email     string
    Name      string
    Password  string  // hashed
    Active    bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

And domain errors:
- `ErrUserNotFound`
- `ErrEmailAlreadyExists`
- `ErrInvalidCredentials`

Add a method `(u *User) Validate() error` checking Name and Email are not empty.

### Task 2 — Use Case Layer

In `usecase/user_repository.go`, define the `UserRepository` interface with:
- `FindByID(id string) (*domain.User, error)`
- `FindByEmail(email string) (*domain.User, error)`
- `Save(user *domain.User) error`
- `Update(user *domain.User) error`
- `Delete(id string) error`
- `List() ([]*domain.User, error)`

In `usecase/user_usecase.go`, implement `UserUseCase` with:
- `Register(name, email, password string) (*domain.User, error)` — check duplicate email, hash password, save
- `Login(email, password string) (*domain.User, error)` — find by email, verify password
- `GetUser(id string) (*domain.User, error)`
- `ListUsers() ([]*domain.User, error)`
- `UpdateUser(id, name string) (*domain.User, error)`
- `DeleteUser(id string) error`

For password hashing, use a simple approach: `hash = fmt.Sprintf("hashed:%s", password)` (we'll avoid bcrypt for simplicity).

### Task 3 — Repository Layer (`repository/memory_user_repository.go`)

Implement the `UserRepository` interface using an in-memory map:
- Use `sync.RWMutex` for thread safety
- Use `uuid`-like IDs: `fmt.Sprintf("user-%d", time.Now().UnixNano())`

### Task 4 — Handler Layer (`handler/user_handler.go`)

HTTP handlers that call use cases. Each handler should:
- Parse JSON body or path parameters
- Call the appropriate use case method
- Map domain errors to HTTP status codes:
  - `ErrUserNotFound` → 404
  - `ErrEmailAlreadyExists` → 409
  - `ErrInvalidCredentials` → 401
  - Other errors → 500

Endpoints:
```
POST   /api/users           Register
POST   /api/auth/login      Login
GET    /api/users           List
GET    /api/users/{id}      Get by ID
PUT    /api/users/{id}      Update name
DELETE /api/users/{id}      Delete
```

### Task 5 — Wire in `main.go` and Write Use Case Tests

Wire all layers in `main.go` using dependency injection.

Write unit tests for `UserUseCase` in `usecase/user_usecase_test.go`:
- Test successful registration
- Test duplicate email returns `ErrEmailAlreadyExists`
- Test login with correct credentials succeeds
- Test login with wrong password returns `ErrInvalidCredentials`
- Test get non-existent user returns `ErrUserNotFound`

Use the real in-memory repository — no mocking needed for this lab.

---

## Tips

- Each layer only imports the layer directly inside it — never imports outer layers.
- Use `errors.As` to check error types: `var notFound ErrUserNotFound; errors.As(err, &notFound)`
- The handler layer is the only place that knows about HTTP — use cases know nothing about `http.Request`.
- Put `UserRepository` interface in the `usecase` package (not `repository`) — the use case defines what it needs, the repository implements it.

---

## Running Your Solution

```bash
cd lab09-clean-architecture
go run .
go test ./usecase/... -v
```
