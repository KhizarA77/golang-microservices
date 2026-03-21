package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Task 1: Singleton ===")
	task1Singleton()

	fmt.Println("\n=== Task 2: Logger Factory ===")
	task2LoggerFactory()

	fmt.Println("\n=== Task 3: HTTP Request Builder ===")
	task3RequestBuilder()

	fmt.Println("\n=== Task 4: Functional Options ===")
	task4FunctionalOptions()
}

// =============================================================================
// Task 1 — Thread-Safe Singleton
// =============================================================================

// DatabasePool represents a pool of database connections.
// TODO: Add fields: dsn string, maxConns int, connected bool
type DatabasePool struct {
	// TODO: fields
	dsn       string
	maxConns  int
	connected bool
}

var (
	dbInstance *DatabasePool
	dbOnce     sync.Once
)

// GetDatabasePool returns the singleton DatabasePool instance.
// The pool is initialized exactly once, even under concurrent access.
func GetDatabasePool() *DatabasePool {
	dbOnce.Do(func() {
		// TODO: Initialize dbInstance
		// TODO: Print "Initializing database pool..." to show it runs once
		fmt.Println("Initializing database pool")
		dbInstance = &DatabasePool{
			// TODO: set dsn, maxConns, connected
			dsn:       "postgres://",
			maxConns:  10,
			connected: true,
		}
	})
	return dbInstance
}

// Query simulates running a SQL query.
func (p *DatabasePool) Query(sql string) string {
	// TODO: Return a fake result string like fmt.Sprintf("Result of: %s", sql)
	return fmt.Sprintf("Result of: %s", sql)
}

// Stats returns connection pool information.
func (p *DatabasePool) Stats() string {
	// TODO: Return a string showing dsn, maxConns, connected status
	return fmt.Sprintf("dsn: %s, maxConns: %d, connected: %t", p.dsn, p.maxConns, p.connected)
}

func task1Singleton() {
	var wg sync.WaitGroup
	// Launch 10 goroutines all trying to get the database pool
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pool := GetDatabasePool()
			// TODO: Print pointer address and id to show they're all the same
			fmt.Printf("goroutine %d: %p\n", id, pool)
		}(i)
	}
	wg.Wait()

	pool := GetDatabasePool()
	fmt.Println(pool.Stats())
	fmt.Println(pool.Query("SELECT * FROM users"))
}

// =============================================================================
// Task 2 — Logger Factory
// =============================================================================

// Logger defines the logging interface.
type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

// TODO: Implement ConsoleLogger
// Prints: "[INFO] msg", "[ERROR] msg", "[DEBUG] msg"
type ConsoleLogger struct{}

func (l ConsoleLogger) Info(msg string)  { fmt.Printf("[INFO] %s\n", msg) }
func (l ConsoleLogger) Error(msg string) { fmt.Printf("[ERROR] %s\n", msg) }
func (l ConsoleLogger) Debug(msg string) { fmt.Printf("[DEBUG] %s\n", msg) }

// TODO: Implement JSONLogger
// Prints: {"level":"info","msg":"..."} etc.
type JSONLogger struct{}

func (l JSONLogger) Info(msg string)  { fmt.Printf(`{"level":"info", "msg": %s}`+"\n", msg) }
func (l JSONLogger) Error(msg string) { fmt.Printf(`{"level":"error", "msg": %s}`+"\n", msg) }
func (l JSONLogger) Debug(msg string) { fmt.Printf(`{"level":"debug", "msg": %s}`+"\n", msg) }

// TODO: Implement SilentLogger — discards all output
type SilentLogger struct{}

func (l SilentLogger) Info(msg string)  { /* TODO: do nothing */ }
func (l SilentLogger) Error(msg string) { /* TODO: do nothing */ }
func (l SilentLogger) Debug(msg string) { /* TODO: do nothing */ }

// NewLogger is the factory function that returns a Logger by format name.
// Supported formats: "console", "json", "silent"
// Return an error (or panic) for unknown formats.
func NewLogger(format string) Logger {
	// TODO: switch on format, return appropriate Logger implementation
	switch format {
	case "console":
		return ConsoleLogger{}
	case "json":
		return JSONLogger{}
	case "silent":
		return SilentLogger{}
	}
	return nil
}

func task2LoggerFactory() {
	formats := []string{"console", "json", "silent"}
	for _, fmt := range formats {
		logger := NewLogger(fmt)
		if logger == nil {
			continue
		}
		logger.Info("server started")
		logger.Error("something went wrong")
		logger.Debug("verbose details here")
	}
}

// =============================================================================
// Task 3 — HTTP Request Builder
// =============================================================================

// RequestBuilder builds an *http.Request step by step.
type RequestBuilder struct {
	method  string
	url     string
	headers map[string]string
	body    string
	timeout time.Duration
	// TODO: Add basicAuthUser, basicAuthPass fields
	basicAuthUser string
	basicAuthPass string
}

// NewRequestBuilder creates a new builder with method and URL.
func NewRequestBuilder(method, url string) *RequestBuilder {
	return &RequestBuilder{
		method:  method,
		url:     url,
		headers: make(map[string]string),
	}
}

// Header adds a request header. Returns self for chaining.
func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	// TODO: Store header, return b
	b.headers[key] = value
	return b
}

// Body sets the request body as a string. Returns self for chaining.
func (b *RequestBuilder) Body(data string) *RequestBuilder {
	// TODO: Set body, return b
	b.body = data
	return b
}

// Timeout sets the request timeout. Returns self for chaining.
func (b *RequestBuilder) Timeout(d time.Duration) *RequestBuilder {
	// TODO: Set timeout, return b
	b.timeout = d
	return b
}

// BasicAuth sets HTTP Basic Auth credentials. Returns self for chaining.
func (b *RequestBuilder) BasicAuth(user, pass string) *RequestBuilder {
	// TODO: Store user/pass, return b
	b.basicAuthUser = user
	b.basicAuthPass = pass
	return b
}

// Build constructs and returns the *http.Request.
func (b *RequestBuilder) Build() (*http.Request, error) {
	// TODO: Create http.Request using http.NewRequest
	// TODO: Add all stored headers
	req, err := http.NewRequest(b.method, b.url, strings.NewReader(b.body))
	if err != nil {
		return nil, err
	}
	for key, value := range b.headers {
		req.Header.Set(key, value)
	}
	// TODO: If basicAuthUser is set, call req.SetBasicAuth
	if b.basicAuthUser != "" {
		req.SetBasicAuth(b.basicAuthUser, b.basicAuthPass)
	}
	// TODO: Return request and any error
	return req, nil
}

func task3RequestBuilder() {
	req, err := NewRequestBuilder("GET", "https://api.example.com/users").
		Header("Accept", "application/json").
		Header("X-Request-ID", "abc-123").
		Timeout(5 * time.Second).
		Build()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if req != nil {
		fmt.Printf("Method: %s\n", req.Method)
		fmt.Printf("URL: %s\n", req.URL.String())
		fmt.Printf("Headers: %v\n", req.Header)
	}

	// Build a POST with auth
	postReq, err := NewRequestBuilder("POST", "https://api.example.com/login").
		Header("Content-Type", "application/json").
		Body(`{"username":"admin","password":"secret"}`).
		BasicAuth("admin", "secret").
		Build()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if postReq != nil {
		fmt.Printf("POST URL: %s, Auth: %s\n", postReq.URL, postReq.Header.Get("Authorization"))
	}
}

// =============================================================================
// Task 4 — Functional Options
// =============================================================================

// Server represents an HTTP server configuration.
type Server struct {
	host         string
	port         int
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxConns     int
	tlsEnabled   bool
}

// Option is a function that configures a Server.
type Option func(*Server)

// TODO: Implement option functions:
// WithHost(host string) Option
// WithPort(port int) Option
// WithReadTimeout(d time.Duration) Option
// WithWriteTimeout(d time.Duration) Option
// WithMaxConns(n int) Option
// WithTLS() Option

// WithHost sets the server host.
func WithHost(host string) Option {
	return func(s *Server) {
		// TODO: set s.host
		s.host = host
	}
}

func WithPort(port int) Option {
	return func(s *Server) {
		// TODO: set s.port
		s.port = port
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		// TODO
		s.readTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		// TODO
		s.writeTimeout = d
	}
}

func WithMaxConns(n int) Option {
	return func(s *Server) {
		// TODO
		s.maxConns = n
	}
}

func WithTLS() Option {
	return func(s *Server) {
		// TODO: set s.tlsEnabled = true
		s.tlsEnabled = true
	}
}

// NewServer creates a Server with defaults, then applies all provided options.
// Defaults: host="localhost", port=8080, readTimeout=30s, writeTimeout=30s,
//
//	maxConns=100, tlsEnabled=false
func NewServer(opts ...Option) *Server {
	// TODO: Create server with defaults
	s := &Server{
		// TODO: fill defaults
		host:         "localhost",
		port:         8080,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		maxConns:     1,
		tlsEnabled:   false,
	}
	// TODO: Apply each option
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// String returns a human-readable representation of the server config.
func (s *Server) String() string {
	// TODO: Return formatted string showing all fields

	return fmt.Sprintf(
		"host: %s\nport: %d\nreadTimeout: %v\nwriteTimeout:%v\nmaxConns:%d\ntlsEnabled:%t",
		s.host, s.port, s.readTimeout, s.writeTimeout, s.maxConns, s.tlsEnabled)
}

func task4FunctionalOptions() {
	// Default server
	s1 := NewServer()
	fmt.Println("Default:", s1)

	// Custom server
	s2 := NewServer(
		WithHost("0.0.0.0"),
		WithPort(9090),
		WithTLS(),
		WithMaxConns(500),
		WithReadTimeout(10*time.Second),
	)
	fmt.Println("Custom:", s2)

	// Minimal override
	s3 := NewServer(WithPort(3000))
	fmt.Println("Minimal:", s3)
}
