# Lab 04 — Creational Design Patterns

**Level:** Beginner–Intermediate
**Topic:** Singleton, Factory Method, Builder, Functional Options

---

## Background

Creational patterns deal with **object creation** — how you construct instances in a controlled, flexible way.

---

### Singleton

Ensures only **one instance** of a type exists in the application. In Go, the thread-safe way is `sync.Once`:

```go
type DB struct { conn string }

var (
    instance *DB
    once     sync.Once
)

func GetDB() *DB {
    once.Do(func() {
        instance = &DB{conn: "postgres://..."}
        fmt.Println("DB connected")
    })
    return instance
}
```

`sync.Once` guarantees the function runs exactly once, even if called from multiple goroutines simultaneously. This is preferred over `init()` for lazy initialization or when the resource is expensive to create.

**When to use:** Database connections, logger instances, configuration stores, connection pools.

---

### Factory Method

Defines an interface for creating objects but lets subclasses (or implementations) decide which type to instantiate. In Go, this is typically a function that returns an interface.

```go
type Logger interface {
    Log(msg string)
}

type JSONLogger  struct{}
type TextLogger  struct{}

func (j JSONLogger) Log(msg string)  { fmt.Printf(`{"msg":"%s"}`+"\n", msg) }
func (t TextLogger) Log(msg string)  { fmt.Println("[TEXT]", msg) }

// Factory function
func NewLogger(format string) Logger {
    switch format {
    case "json":  return JSONLogger{}
    case "text":  return TextLogger{}
    default:      panic("unknown format")
    }
}
```

**When to use:** When you need to create different implementations of an interface based on a configuration value or runtime condition.

---

### Builder

Constructs a complex object step by step. Useful when an object has many optional fields and telescoping constructors become unwieldy.

```go
type QueryBuilder struct {
    table  string
    where  []string
    limit  int
    offset int
}

func NewQueryBuilder(table string) *QueryBuilder {
    return &QueryBuilder{table: table}
}

func (b *QueryBuilder) Where(cond string) *QueryBuilder {
    b.where = append(b.where, cond)
    return b  // return self for chaining
}

func (b *QueryBuilder) Limit(n int) *QueryBuilder {
    b.limit = n
    return b
}

func (b *QueryBuilder) Build() string {
    // build and return SQL string
}
```

**When to use:** Constructing SQL queries, HTTP requests, complex configuration objects.

---

### Functional Options (Go Idiomatic)

The most idiomatic Go pattern for optional configuration. Instead of many constructor parameters, you pass variadic option functions.

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithHost(host string) Option {
    return func(s *Server) { s.host = host }
}

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func NewServer(opts ...Option) *Server {
    s := &Server{host: "localhost", port: 8080, timeout: 30 * time.Second} // defaults
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
srv := NewServer(WithPort(9090), WithHost("0.0.0.0"))
```

**Why it's great:** Defaults are centralized, callers only specify what they need, adding new options is backwards-compatible.

---

## Learning Objectives

By the end of this lab you will be able to:

- Implement a thread-safe Singleton using `sync.Once`
- Create a Factory that returns different interface implementations
- Build a fluent Builder for complex object construction
- Apply the Functional Options pattern for flexible constructors

---

## Tasks

### Task 1 — Thread-Safe Singleton (Database Connection)

Implement a `DatabasePool` singleton:
- Struct with fields: `dsn string`, `maxConns int`, `connected bool`
- A `GetDatabasePool()` function using `sync.Once`
- Method `Query(sql string) string` that returns a fake result
- Method `Stats() string` that returns connection info

Test it by calling `GetDatabasePool()` from 10 concurrent goroutines — all should get the same pointer and the connection message should print only once.

### Task 2 — Logger Factory

Define a `Logger` interface with `Info(msg string)`, `Error(msg string)`, `Debug(msg string)`.

Implement three loggers:
- `ConsoleLogger` — prints with a simple prefix like `[INFO]`
- `JSONLogger` — prints JSON format `{"level":"info","msg":"..."}`
- `SilentLogger` — discards all output (useful for testing)

Write a `NewLogger(format string) Logger` factory function.

Test by creating each type and calling all methods.

### Task 3 — HTTP Request Builder

Build a `RequestBuilder` for constructing HTTP requests:
- Required: `method string`, `url string`
- Optional (chaining): `Header(key, value string)`, `Body(data string)`, `Timeout(d time.Duration)`, `BasicAuth(user, pass string)`
- `Build() (*http.Request, error)` — constructs the actual `*http.Request`

Example usage:
```go
req, err := NewRequestBuilder("GET", "https://api.example.com/users").
    Header("Accept", "application/json").
    Timeout(5 * time.Second).
    Build()
```

### Task 4 — Server Config with Functional Options

Define a `Server` struct:
```go
type Server struct {
    host         string
    port         int
    readTimeout  time.Duration
    writeTimeout time.Duration
    maxConns     int
    tlsEnabled   bool
}
```

Defaults: `host="localhost"`, `port=8080`, `readTimeout=30s`, `writeTimeout=30s`, `maxConns=100`, `tlsEnabled=false`

Implement `Option` type and these option functions:
- `WithHost(host string) Option`
- `WithPort(port int) Option`
- `WithReadTimeout(d time.Duration) Option`
- `WithWriteTimeout(d time.Duration) Option`
- `WithMaxConns(n int) Option`
- `WithTLS() Option`

`NewServer(opts ...Option) *Server` applies options over defaults.

Add a `String() string` method for pretty printing config.

Test with various combinations.

---

## Tips

- `sync.Once` is zero-value safe — you can declare `var once sync.Once` at the package level without initialization.
- Returning the builder pointer from each method (method chaining) works because the methods take a pointer receiver.
- Functional options are better than a config struct parameter because they allow you to keep adding options without breaking existing call sites.
- For the factory, return an interface not a concrete type — callers shouldn't know or care what concrete type they get.

---

## Running Your Solution

```bash
cd lab04-creational-patterns
go run .
```
