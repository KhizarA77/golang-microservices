# Lab 11 — gRPC Services

**Level:** Advanced
**Topic:** Protocol Buffers, gRPC, Streaming, Interceptors

---

## Background

### Why gRPC?

gRPC is a high-performance RPC framework from Google. Compared to REST:

| | REST | gRPC |
|-|------|------|
| Protocol | HTTP/1.1 | HTTP/2 |
| Format | JSON (text) | Protobuf (binary) |
| Contract | OpenAPI (optional) | `.proto` file (required) |
| Streaming | Limited | First-class |
| Type safety | Manual | Generated |
| Performance | Good | Excellent |

**Use gRPC for:** Internal service-to-service communication, real-time streaming, performance-critical paths.
**Use REST for:** Public APIs, browser clients, simple integrations.

---

### Protocol Buffers (protobuf)

Protobuf is a language-neutral schema definition language. You define messages and services in a `.proto` file:

```protobuf
syntax = "proto3";
package product;
option go_package = "./proto;proto";

message Product {
  int32  id    = 1;
  string name  = 2;
  double price = 3;
  int32  stock = 4;
}

message GetProductRequest {
  int32 id = 1;
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product);
  rpc ListProducts(ListRequest) returns (stream Product);  // server streaming
  rpc CreateProduct(Product) returns (Product);
}
```

The `protoc` compiler with the Go plugin generates Go code from this file.

---

### gRPC Service Types

| Type | Description |
|------|-------------|
| Unary | Single request → single response (like HTTP) |
| Server streaming | Single request → stream of responses |
| Client streaming | Stream of requests → single response |
| Bidirectional streaming | Stream of requests ↔ stream of responses |

---

### Error Handling in gRPC

gRPC uses status codes, not HTTP status codes:

```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

// Return an error
return nil, status.Errorf(codes.NotFound, "product %d not found", id)

// Check error type on client
if st, ok := status.FromError(err); ok {
    if st.Code() == codes.NotFound {
        // handle not found
    }
}
```

Common codes: `OK`, `NotFound`, `InvalidArgument`, `Internal`, `Unavailable`, `AlreadyExists`, `Unauthenticated`

---

### Interceptors (Middleware for gRPC)

```go
// Unary server interceptor
func loggingInterceptor(ctx context.Context, req interface{},
    info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

    start := time.Now()
    resp, err := handler(ctx, req)
    log.Printf("%s took %v err=%v", info.FullMethod, time.Since(start), err)
    return resp, err
}

// Apply to server
grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
```

---

### Context and Metadata

gRPC uses `context.Context` for cancellation and deadlines. Metadata is like HTTP headers:

```go
// Client: attach metadata
md := metadata.Pairs("authorization", "Bearer token")
ctx := metadata.NewOutgoingContext(context.Background(), md)

// Server: read metadata
md, ok := metadata.FromIncomingContext(ctx)
```

---

## Project Structure

```
lab11-grpc-services/
├── go.mod
├── proto/
│   └── product.proto         ← Service definition
│   └── product.pb.go         ← Generated (run protoc)
│   └── product_grpc.pb.go    ← Generated (run protoc)
├── server/
│   └── server.go             ← gRPC server implementation
├── client/
│   └── client.go             ← gRPC client
└── main.go                   ← Starts server in background, runs client
```

---

## Setup

Install tools:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Install protoc: https://grpc.io/docs/protoc-installation/

Generate code:
```bash
cd lab11-grpc-services
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/product.proto
```

Or on Windows with protoc in PATH.

---

## Learning Objectives

By the end of this lab you will be able to:

- Define a gRPC service in a `.proto` file
- Implement a gRPC server
- Connect with a gRPC client
- Use server streaming to emit a sequence of responses
- Add unary interceptors (logging, error recovery)
- Handle gRPC status codes

---

## Tasks

### Task 1 — Proto Definition

Write `proto/product.proto`:

Messages:
- `Product { id, name, description, price, stock, category }`
- `GetProductRequest { id }`
- `ListProductsRequest { category (optional), page_size }`
- `CreateProductRequest { name, description, price, stock, category }`
- `DeleteProductRequest { id }`
- `DeleteProductResponse { success, message }`

Service `ProductService`:
- `rpc GetProduct(GetProductRequest) returns (Product)` — unary
- `rpc ListProducts(ListProductsRequest) returns (stream Product)` — server streaming
- `rpc CreateProduct(CreateProductRequest) returns (Product)` — unary
- `rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse)` — unary

### Task 2 — gRPC Server

In `server/server.go`, implement `ProductServiceServer`:
- In-memory product store (seeded with 5 products)
- `GetProduct` — find by ID, return `NotFound` if missing
- `ListProducts` — stream all products (filter by category if provided)
  - Add a small `time.Sleep(50*time.Millisecond)` between sends to show streaming
- `CreateProduct` — validate, assign ID, store and return
- `DeleteProduct` — remove from store

Add two unary interceptors:
1. Logging interceptor — logs method name, duration, error
2. Recovery interceptor — recovers from panics, returns `Internal` error

### Task 3 — gRPC Client

In `client/client.go`, implement a `ProductClient` that:
- Connects to the gRPC server
- `GetProduct(id int32) (*proto.Product, error)`
- `ListProducts(category string) ([]*proto.Product, error)` — consumes the stream
- `CreateProduct(name, description, category string, price float64, stock int32) (*proto.Product, error)`
- `DeleteProduct(id int32) (bool, error)`

### Task 4 — Wire in `main.go`

In `main.go`:
1. Start the gRPC server in a goroutine (port 50051)
2. Wait for it to be ready
3. Create a client and run a demo:
   - Create 3 products
   - List all products (show streaming)
   - Get product by ID
   - List by category
   - Delete a product
   - Try to get the deleted product (should return NotFound)
4. Shutdown cleanly

---

## Tips

- Run `go mod tidy` after adding gRPC dependencies — they'll be in `go.sum`
- The generated `.pb.go` and `_grpc.pb.go` files should be committed to source control
- Use `grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))` for unencrypted local connections
- Server streaming: call `stream.Send(&product)` in a loop, return nil when done
- Check `stream.Context().Done()` inside the stream loop to respect client cancellation

---

## Dependencies

```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

After running `go mod tidy`, your `go.mod` will include these dependencies.
