package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	fmt.Println("=== Task 1: Adapter ===")
	task1Adapter()

	fmt.Println("\n=== Task 2: Middleware Decorators ===")
	task2Middleware()

	fmt.Println("\n=== Task 3: Caching Proxy ===")
	task3CachingProxy()

	fmt.Println("\n=== Task 4: File System Composite ===")
	task4Composite()
}

// =============================================================================
// Task 1 — Analytics Adapter
// =============================================================================

// ThirdPartyAnalytics is the incompatible external interface we must adapt.
type ThirdPartyAnalytics interface {
	TrackEventV1(category, action, label string, value float64)
	TrackPageViewV1(path, referrer string)
}

// Analytics is our clean internal interface.
type Analytics interface {
	TrackEvent(name string, props map[string]string)
	TrackPageView(path string)
}

// ThirdPartyAnalyticsImpl is a fake implementation of the third-party library.
type ThirdPartyAnalyticsImpl struct{}

func (t *ThirdPartyAnalyticsImpl) TrackEventV1(category, action, label string, value float64) {
	// TODO: Print "[3RD PARTY] event: category=X action=Y label=Z value=V"
	fmt.Printf("[3RD PARTY] event: category=%s action=%s label=%s value=%f\n", category, action, label, value)
}

func (t *ThirdPartyAnalyticsImpl) TrackPageViewV1(path, referrer string) {
	// TODO: Print "[3RD PARTY] pageview: path=X referrer=Y"
	fmt.Printf("[3RD PARTY] pageview: path=%s referrer=%s\n", path, referrer)
}

// AnalyticsAdapter adapts ThirdPartyAnalytics to our Analytics interface.
type AnalyticsAdapter struct {
	// TODO: Add field: third ThirdPartyAnalytics
	third ThirdPartyAnalytics
}

func NewAnalyticsAdapter(third ThirdPartyAnalytics) Analytics {
	// TODO: Return &AnalyticsAdapter{third: third}
	return &AnalyticsAdapter{
		third: third,
	}
}

func (a *AnalyticsAdapter) TrackEvent(name string, props map[string]string) {
	// TODO: Extract category, action, label, value from props
	//       (use defaults if props don't have them)
	category, ok := props["category"]
	if !ok {
		category = "default"
	}
	action, ok := props["action"]
	if !ok {
		action = "default"
	}
	label, ok := props["label"]
	if !ok {
		label = "default"
	}
	var value float64
	if v, ok := props["value"]; ok {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			value = parsed
		}
	}

	a.third.TrackEventV1(category, action, label, value)
}

func (a *AnalyticsAdapter) TrackPageView(path string) {
	// TODO: Delegate to a.third.TrackPageViewV1(path, "")
	a.third.TrackPageViewV1(path, "")
}

func task1Adapter() {
	third := &ThirdPartyAnalyticsImpl{}
	analytics := NewAnalyticsAdapter(third)

	analytics.TrackEvent("button", map[string]string{
		"category": "ui",
		"action":   "click",
		"label":    "signup",
	})
	analytics.TrackPageView("/home")
	analytics.TrackPageView("/pricing")
}

// =============================================================================
// Task 2 — HTTP Middleware Decorators
// =============================================================================

// Request represents an incoming HTTP request.
type Request struct {
	Method  string
	Path    string
	Body    string
	Headers map[string]string
}

// Response represents an HTTP response.
type Response struct {
	Status int
	Body   string
}

// HandlerFunc is a function that processes a request and returns a response.
type HandlerFunc func(req Request) Response

// LoggingMiddleware logs the request method, path, and response status.
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(req Request) Response {
		// TODO: Call next(req)
		// TODO: Print "[LOG] METHOD PATH -> STATUS"
		// TODO: Return the response
		res := next(req)
		fmt.Printf("[LOG] %s %s -> %d\n", req.Method, req.Path, res.Status)
		return res
	}
}

// AuthMiddleware checks that req.Headers["Authorization"] == expectedToken.
// If not, returns Response{Status: 401, Body: "unauthorized"}.
func AuthMiddleware(expectedToken string) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req Request) Response {
			// TODO: Check req.Headers["Authorization"]
			// TODO: If wrong, return 401
			// TODO: Otherwise, call next(req)
			token, ok := req.Headers["Authorization"]
			if !ok || token != expectedToken {
				return Response{401, ""}
			}
			res := next(req)
			return res
		}
	}
}

// RecoveryMiddleware recovers from panics and returns a 500 response.
func RecoveryMiddleware(next HandlerFunc) HandlerFunc {
	return func(req Request) (res Response) {
		defer func() {
			r := recover()
			if r != nil {
				// TODO: Print "[RECOVERY] panic: <r>"
				// Note: you can't return from a deferred function directly
				// Hint: use a named return variable
				fmt.Printf("[RECOVERY] panic: %s\n", r)
				res = Response{500, "Internal Server Error"}
			}
		}()
		// TODO: Call next(req)
		res = next(req)
		return res
	}
}

// myHandler is a real application handler.
func myHandler(req Request) Response {
	if req.Path == "/panic" {
		panic("something terrible happened")
	}
	return Response{Status: 200, Body: "Hello from " + req.Path}
}

func task2Middleware() {
	// Compose the middleware chain (innermost first, outermost last)
	// RecoveryMiddleware -> AuthMiddleware -> LoggingMiddleware -> myHandler
	// Reading right-to-left: request hits LoggingMiddleware first
	handler := LoggingMiddleware(
		AuthMiddleware("secret")(myHandler))

	// Test 1: Valid auth
	fmt.Println("--- Valid auth ---")
	resp := handler(Request{
		Method:  "GET",
		Path:    "/users",
		Headers: map[string]string{"Authorization": "secret"},
	})
	fmt.Printf("Response: %d %s\n", resp.Status, resp.Body)

	// Test 2: Invalid auth
	fmt.Println("--- Invalid auth ---")
	resp = handler(Request{
		Method:  "GET",
		Path:    "/users",
		Headers: map[string]string{"Authorization": "wrong"},
	})
	fmt.Printf("Response: %d %s\n", resp.Status, resp.Body)

	// Test 3: Panic recovery
	fmt.Println("--- Panic path ---")
	resp = handler(Request{
		Method:  "GET",
		Path:    "/panic",
		Headers: map[string]string{"Authorization": "secret"},
	})
	fmt.Printf("Response: %d %s\n", resp.Status, resp.Body)
}

// =============================================================================
// Task 3 — Caching Proxy
// =============================================================================

// WeatherService fetches weather data for a city.
type WeatherService interface {
	GetWeather(city string) (string, error)
}

// RealWeatherService simulates an external API call.
type RealWeatherService struct{}

func (r *RealWeatherService) GetWeather(city string) (string, error) {
	// TODO: Simulate 200ms delay
	// TODO: Return fmt.Sprintf("Sunny, 25°C in %s", city), nil
	time.Sleep(200 * time.Millisecond)
	return fmt.Sprintf("Sunny, 25°C in %s", city), nil
}

// cacheEntry stores a cached result with a timestamp.
type cacheEntry struct {
	result   string
	cachedAt time.Time
}

// CachingWeatherProxy wraps WeatherService and caches results.
type CachingWeatherProxy struct {
	// TODO: Add fields: real WeatherService, cache map[string]cacheEntry, ttl time.Duration
	real  WeatherService
	cache map[string]cacheEntry
	ttl   time.Duration
}

func NewCachingWeatherProxy(real WeatherService, ttl time.Duration) WeatherService {
	// TODO: Return &CachingWeatherProxy{real: real, cache: make(map[string]cacheEntry), ttl: ttl}
	return &CachingWeatherProxy{
		real:  real,
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

func (p *CachingWeatherProxy) GetWeather(city string) (string, error) {
	// TODO: Check if city is in cache AND cache entry is still fresh (< ttl)
	//       Print "[CACHE HIT] city" and return cached result
	// TODO: Otherwise, call p.real.GetWeather(city)
	//       Print "[CACHE MISS] city"
	//       Store result in cache with current timestamp
	//       Return result
	res, ok := p.cache[city]
	if ok && (time.Since(res.cachedAt) <= p.ttl) {
		fmt.Printf("[CACHE HIT] %s\n", city)
		return res.result, nil
	}
	fmt.Printf("[CACHE MISS] %s\n", city)
	result, err := p.real.GetWeather(city)
	if err != nil {
		return "", err
	}
	p.cache[city] = cacheEntry{result: result, cachedAt: time.Now()}
	return p.cache[city].result, nil
}

func task3CachingProxy() {
	real := &RealWeatherService{}
	proxy := NewCachingWeatherProxy(real, 30*time.Second)

	cities := []string{"London", "Paris", "London", "Tokyo", "Paris", "London"}

	for _, city := range cities {
		weather, err := proxy.GetWeather(city)
		if err != nil {
			fmt.Printf("Error for %s: %v\n", city, err)
			continue
		}
		fmt.Printf("%s: %s\n", city, weather)
	}
}

// =============================================================================
// Task 4 — File System Composite
// =============================================================================

// Component is the common interface for both files and directories.
type Component interface {
	Name() string
	Size() int
	Display(indent string)
}

// File is a leaf node in the composite tree.
type File struct {
	name string
	size int
}

func NewFile(name string, size int) *File {
	return &File{name: name, size: size}
}

func (f *File) Name() string { return f.name }
func (f *File) Size() int    { return f.size }
func (f *File) Display(indent string) {
	// TODO: Print indent + name + " (" + size + " bytes)"
	fmt.Printf("%s %s (%d bytes)\n", indent, f.name, f.size)
}

// Directory is a composite node that can contain other Components.
type Directory struct {
	name     string
	children []Component
}

func NewDirectory(name string) *Directory {
	return &Directory{name: name}
}

func (d *Directory) Add(c Component) {
	// TODO: Append c to d.children
	d.children = append(d.children, c)
}

func (d *Directory) Name() string { return d.name }

func (d *Directory) Size() int {
	total := 0
	for _, c := range d.children {
		total += c.Size()
	}
	return total
}

func (d *Directory) Display(indent string) {
	fmt.Printf("%s%s/ (%d bytes)\n", indent, d.name, d.Size())
	for _, c := range d.children {
		c.Display(indent + "  ")
	}
}

func task4Composite() {
	// Build the file system:
	// root/
	//   docs/
	//     readme.txt  (1024 bytes)
	//     spec.pdf    (40960 bytes)
	//   src/
	//     main.go     (2048 bytes)
	//     utils.go    (1536 bytes)
	//   go.mod        (256 bytes)

	root := NewDirectory("root")

	docs := NewDirectory("docs")
	// TODO: Add readme.txt and spec.pdf to docs
	docs.Add(NewFile("readme.txt", 1024))
	docs.Add(NewFile("spec.pdf", 40960))
	// TODO: Add docs to root
	root.Add(docs)

	src := NewDirectory("src")
	// TODO: Add main.go and utils.go to src
	// TODO: Add src to root
	src.Add(NewFile("main.go", 2048))
	src.Add(NewFile("utils.go", 1536))
	root.Add(src)
	// TODO: Add go.mod to root
	root.Add(NewFile("go.mod", 256))
	root.Display("")
	fmt.Printf("\nTotal size: %d bytes\n", root.Size())
}
