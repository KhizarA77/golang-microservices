# Lab 05 — Structural Design Patterns

**Level:** Intermediate
**Topic:** Adapter, Decorator, Proxy, Composite

---

## Background

Structural patterns deal with **how types and objects are composed** to form larger structures.

---

### Adapter

Wraps an incompatible interface to make it compatible with another. Think of it as a translator between two interfaces.

```go
// Old system's interface
type OldPrinter interface {
    PrintOld(s string)
}

// New system expects this
type NewPrinter interface {
    Print(s string)
}

// Adapter wraps OldPrinter to satisfy NewPrinter
type PrinterAdapter struct {
    old OldPrinter
}

func (a *PrinterAdapter) Print(s string) {
    a.old.PrintOld(s)  // delegate to old
}
```

**When to use:** Integrating third-party libraries, wrapping legacy code, connecting incompatible APIs.

---

### Decorator

Adds behavior to an object without modifying its type. In Go, this is done by wrapping an interface implementation:

```go
type Handler interface {
    Handle(req string) string
}

// Base implementation
type BaseHandler struct{}
func (h BaseHandler) Handle(req string) string { return "response" }

// Logging decorator
type LoggingHandler struct {
    next Handler
}
func (h LoggingHandler) Handle(req string) string {
    fmt.Printf(">> %s\n", req)
    resp := h.next.Handle(req)
    fmt.Printf("<< %s\n", resp)
    return resp
}

// Compose decorators
handler := LoggingHandler{next: AuthHandler{next: BaseHandler{}}}
```

In HTTP middleware, this is the standard pattern. Each middleware wraps the next handler in the chain.

**When to use:** HTTP middleware, logging, caching, rate limiting, metrics collection — any cross-cutting concern.

---

### Proxy

Controls access to another object. The proxy has the same interface as the real object. Common types:

- **Virtual Proxy:** Delays expensive initialization (lazy loading)
- **Caching Proxy:** Returns cached results instead of re-computing
- **Protection Proxy:** Adds access control checks
- **Remote Proxy:** Represents an object in a different address space

```go
type ImageLoader interface {
    Load(path string) []byte
}

type CachingImageLoader struct {
    real  ImageLoader
    cache map[string][]byte
}

func (c *CachingImageLoader) Load(path string) []byte {
    if data, ok := c.cache[path]; ok {
        fmt.Println("cache hit:", path)
        return data
    }
    data := c.real.Load(path)
    c.cache[path] = data
    return data
}
```

**When to use:** Caching, rate limiting, access control, lazy initialization of expensive resources.

---

### Composite

Composes objects into tree structures to represent hierarchies. Clients treat individual objects and compositions uniformly.

```go
type Component interface {
    Display(indent string)
    Size() int
}

type File struct { name string; size int }
func (f File) Display(indent string) { fmt.Println(indent + f.name) }
func (f File) Size() int { return f.size }

type Folder struct {
    name     string
    children []Component
}
func (f *Folder) Add(c Component) { f.children = append(f.children, c) }
func (f *Folder) Display(indent string) {
    fmt.Println(indent + f.name + "/")
    for _, c := range f.children { c.Display(indent + "  ") }
}
func (f *Folder) Size() int {
    total := 0
    for _, c := range f.children { total += c.Size() }
    return total
}
```

**When to use:** File systems, UI component trees, organization charts, menu systems.

---

## Learning Objectives

By the end of this lab you will be able to:

- Adapt an incompatible interface to a new one without modifying the original
- Layer behavior on top of an interface using decorators
- Intercept and modify access to an object using a proxy
- Build tree structures where leaves and nodes share an interface

---

## Tasks

### Task 1 — Analytics Adapter

You have a third-party analytics library with this interface:
```go
type ThirdPartyAnalytics interface {
    TrackEventV1(category, action, label string, value float64)
    TrackPageViewV1(path, referrer string)
}
```

Your application uses a cleaner internal interface:
```go
type Analytics interface {
    TrackEvent(name string, props map[string]string)
    TrackPageView(path string)
}
```

Implement:
1. `ThirdPartyAnalyticsImpl` — a fake implementation of `ThirdPartyAnalytics` that prints what it would track
2. `AnalyticsAdapter` — wraps `ThirdPartyAnalytics` and implements `Analytics`
3. Test it: use `Analytics` interface in your code, backed by the adapter

### Task 2 — HTTP Middleware Decorators

Define a handler function type:
```go
type HandlerFunc func(req Request) Response
```

Where:
```go
type Request  struct { Method, Path, Body string; Headers map[string]string }
type Response struct { Status int; Body string }
```

Implement three middleware decorators (each wraps a `HandlerFunc`):
1. `LoggingMiddleware(next HandlerFunc) HandlerFunc` — logs `"[LOG] METHOD PATH -> STATUS"`
2. `AuthMiddleware(token string) func(HandlerFunc) HandlerFunc` — checks `req.Headers["Authorization"]`, returns 401 if wrong
3. `RecoveryMiddleware(next HandlerFunc) HandlerFunc` — recovers from panics, returns 500

Write a real handler and compose all three:
```go
handler := LoggingMiddleware(AuthMiddleware("secret")(RecoveryMiddleware(myHandler)))
```

Test with valid auth, invalid auth, and a handler that panics.

### Task 3 — Caching Proxy

Define a `WeatherService` interface:
```go
type WeatherService interface {
    GetWeather(city string) (string, error)
}
```

Implement:
1. `RealWeatherService` — simulates an API call with 200ms delay, returns `"Sunny, 25°C"` for any city
2. `CachingWeatherProxy` — wraps `RealWeatherService`, caches results for 30 seconds
   - Cache hit: return immediately, print `"[CACHE HIT] city"`
   - Cache miss: call real service, store result with timestamp, print `"[CACHE MISS] city"`

Test by calling the same city multiple times and verifying cache hits.

### Task 4 — File System Composite

Build a simple file system composite:

- `Component` interface: `Name() string`, `Size() int`, `Display(indent string)`
- `File` struct: has `name` and `size` (in bytes)
- `Directory` struct: has `name` and `children []Component`
  - `Add(c Component)` to add children
  - `Size()` returns total size of all children recursively
  - `Display()` shows tree with indentation

Create this structure and display it:
```
root/ (size: X bytes)
  docs/ (size: X bytes)
    readme.txt (1024 bytes)
    spec.pdf (40960 bytes)
  src/ (size: X bytes)
    main.go (2048 bytes)
    utils.go (1536 bytes)
  go.mod (256 bytes)
```

---

## Tips

- Adapter: the adapter struct holds a reference to the adaptee — delegation does the translation.
- Decorator: the key is that the decorator and the thing it decorates share the same interface — this enables unlimited stacking.
- Proxy: use `time.Since(cachedAt) < ttl` to check cache freshness.
- Composite: the `Display` method passes an increasing indent string through recursion.

---

## Running Your Solution

```bash
cd lab05-structural-patterns
go run .
```
