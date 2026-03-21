package server

// NOTE: This file references the generated proto package.
// Run `protoc` to generate proto/product.pb.go and proto/product_grpc.pb.go first.
// See README.md for the protoc command.

// After generating, uncomment the imports and implement the TODOs.

/*
import (
	"context"
	"fmt"
	"lab11/proto"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductServer implements proto.ProductServiceServer.
type ProductServer struct {
	proto.UnimplementedProductServiceServer
	mu       sync.RWMutex
	products map[int32]*proto.Product
	nextID   int32
}

// NewProductServer creates a new server seeded with sample products.
func NewProductServer() *ProductServer {
	s := &ProductServer{
		products: make(map[int32]*proto.Product),
	}

	// Seed with sample data
	seeds := []*proto.CreateProductRequest{
		{Name: "MacBook Pro", Description: "Apple laptop", Price: 1999.99, Stock: 5, Category: "electronics"},
		{Name: "Go Book", Description: "Learn Go", Price: 39.99, Stock: 100, Category: "books"},
		{Name: "Winter Jacket", Description: "Warm jacket", Price: 89.99, Stock: 30, Category: "clothing"},
		{Name: "Organic Coffee", Description: "Single origin", Price: 14.99, Stock: 200, Category: "food"},
		{Name: "Headphones", Description: "Noise cancelling", Price: 299.99, Stock: 15, Category: "electronics"},
	}
	for _, seed := range seeds {
		s.createProduct(seed)
	}
	return s
}

func (s *ProductServer) createProduct(req *proto.CreateProductRequest) *proto.Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	p := &proto.Product{
		Id:          s.nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
	}
	s.products[p.Id] = p
	return p
}

// GetProduct handles unary GetProduct RPC.
func (s *ProductServer) GetProduct(ctx context.Context, req *proto.GetProductRequest) (*proto.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: Look up req.Id in s.products
	// TODO: If not found, return nil, status.Errorf(codes.NotFound, "product %d not found", req.Id)
	// TODO: Return a copy of the product
	_ = req
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// ListProducts handles server streaming ListProducts RPC.
func (s *ProductServer) ListProducts(req *proto.ListProductsRequest, stream proto.ProductService_ListProductsServer) error {
	s.mu.RLock()
	products := make([]*proto.Product, 0)
	for _, p := range s.products {
		// TODO: If req.Category != "", only include matching products
		products = append(products, p)
	}
	s.mu.RUnlock()

	// TODO: If req.PageSize > 0, limit to that many products

	for _, p := range products {
		// Check if client cancelled
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
		}

		// Simulate some processing time
		time.Sleep(50 * time.Millisecond)

		// TODO: stream.Send(p)
		_ = p
		_ = stream
	}
	return nil
}

// CreateProduct handles unary CreateProduct RPC.
func (s *ProductServer) CreateProduct(ctx context.Context, req *proto.CreateProductRequest) (*proto.Product, error) {
	// TODO: Validate req.Name is not empty — return InvalidArgument if so
	// TODO: Validate req.Price > 0
	// TODO: Create product using s.createProduct(req)
	// TODO: Return the product
	_ = req
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DeleteProduct handles unary DeleteProduct RPC.
func (s *ProductServer) DeleteProduct(ctx context.Context, req *proto.DeleteProductRequest) (*proto.DeleteProductResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Check if req.Id exists in s.products
	// TODO: If not found, return response with success=false and message
	// TODO: Delete from map
	// TODO: Return response with success=true
	_ = req
	_ = fmt.Sprintf
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
*/

// Placeholder to keep the package compilable before proto generation.
// Delete this file's content and uncomment the code above after running protoc.
type placeholder struct{}
