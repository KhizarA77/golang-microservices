package main

import (
	"fmt"
	"log"
)

// main.go — Wire the gRPC server and client together.
//
// STEP 1: Generate proto code first:
//   protoc --go_out=. --go_opt=paths=source_relative \
//          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
//          proto/product.proto
//
// STEP 2: Run `go mod tidy` to download dependencies.
//
// STEP 3: Uncomment the code below and implement the TODOs in server/ and client/.

func main() {
	fmt.Println("Lab 11 — gRPC Services")
	fmt.Println()
	fmt.Println("Before running this lab:")
	fmt.Println("  1. Install protoc and Go plugins (see README.md)")
	fmt.Println("  2. Run protoc to generate proto/product.pb.go")
	fmt.Println("  3. Run: go mod tidy")
	fmt.Println("  4. Uncomment server and client code, implement TODOs")
	fmt.Println("  5. Run: go run .")
	log.Println("Waiting for implementation...")
}

// Uncomment after proto generation and implementation:
/*
import (
	"context"
	"fmt"
	"lab11/client"
	"lab11/proto"
	"lab11/server"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// TODO: Create logging interceptor
	loggingInterceptor := func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		log.Printf("[gRPC] %s took %v (error: %v)", info.FullMethod, time.Since(start), err)
		return resp, err
	}

	// TODO: Create recovery interceptor
	recoveryInterceptor := func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] recovered in %s: %v", info.FullMethod, r)
				// return internal error
			}
		}()
		return handler(ctx, req)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor, recoveryInterceptor),
	)

	proto.RegisterProductServiceServer(grpcServer, server.NewProductServer())

	go func() {
		log.Printf("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	c, err := client.NewProductClient("localhost:50051")
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// Demo
	fmt.Println("=== Creating Products ===")
	// TODO: Create 3 products, print each

	fmt.Println("\n=== Listing All Products ===")
	// TODO: List all products (streaming), print each

	fmt.Println("\n=== List Electronics ===")
	// TODO: List only electronics

	fmt.Println("\n=== Get Product by ID ===")
	// TODO: Get product with ID 1

	fmt.Println("\n=== Delete Product ===")
	// TODO: Delete product with ID 1

	fmt.Println("\n=== Get Deleted Product (expect NotFound) ===")
	// TODO: Try to get product 1, check for NotFound error

	grpcServer.GracefulStop()
	fmt.Println("\nDone.")
}
*/
