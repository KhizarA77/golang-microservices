package main

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
/*
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
*/
// Uncomment after proto generation and implementation:

func main() {
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
			r := recover()
			if r != nil {
				log.Printf("[PANIC] recovered in %s: %v", info.FullMethod, r)
				// return internal error
				err = status.Errorf(codes.Internal, "internal server error")

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
	c.CreateProduct(ctx, &proto.CreateProductRequest{
		Name: "Macbook", Price: 599.99, Description: "Laptop", Stock: 10, Category: "Tech"})
	c.CreateProduct(ctx, &proto.CreateProductRequest{
		Name: "Airpods", Price: 99.99, Description: "Earbuds", Stock: 10, Category: "Tech",
	})
	c.CreateProduct(ctx, &proto.CreateProductRequest{
		Name: "BMW M3", Price: 119999.99, Description: "Car", Stock: 3, Category: "Cars",
	})
	fmt.Println("\n=== Listing All Products ===")
	// TODO: List all products (streaming), print each
	sli, err := c.ListProducts(ctx, "")
	if err != nil {
		fmt.Printf("No products exist")
	} else {
		fmt.Println(sli)
	}
	fmt.Println("\n=== List Electronics ===")
	// TODO: List only electronics
	sli, err = c.ListProducts(ctx, "Tech")
	if err != nil {
		fmt.Printf("No products in Tech")
	} else {
		fmt.Println(sli)
	}
	fmt.Println("\n=== Get Product by ID ===")
	// TODO: Get product with ID 1
	prod, err := c.GetProduct(ctx, 1)
	if err == nil {
		fmt.Println(prod)
	}
	fmt.Println("\n=== Delete Product ===")
	// TODO: Delete product with ID 1
	success, msg, err := c.DeleteProduct(ctx, 1)
	if err != nil {
		log.Printf("Delete error: %v", err)
	} else {
		fmt.Printf("Deleted: success=%v message=%s\n", success, msg)
	}

	fmt.Println("\n=== Get Deleted Product (expect NotFound) ===")
	// TODO: Try to get product 1, check for NotFound error
	_, err = c.GetProduct(ctx, 1)
	if client.IsNotFound(err) {
		fmt.Println("Correctly got NotFound error for deleted product")
	} else if err != nil {
		log.Printf("Unexpected error: %v", err)
	}

	grpcServer.GracefulStop()
	fmt.Println("\nDone.")
}
