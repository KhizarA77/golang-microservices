package client

// NOTE: Uncomment after running protoc to generate proto package.

/*
import (
	"context"
	"io"
	"lab11/proto"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// ProductClient wraps the generated gRPC client with typed methods.
type ProductClient struct {
	conn   *grpc.ClientConn
	client proto.ProductServiceClient
}

// NewProductClient creates a new connected client.
func NewProductClient(addr string) (*ProductClient, error) {
	// TODO: Dial addr with insecure credentials (for local dev)
	// grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &ProductClient{
		conn:   conn,
		client: proto.NewProductServiceClient(conn),
	}, nil
}

// Close closes the underlying connection.
func (c *ProductClient) Close() error {
	return c.conn.Close()
}

// GetProduct fetches a product by ID.
func (c *ProductClient) GetProduct(ctx context.Context, id int32) (*proto.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// TODO: Call c.client.GetProduct(ctx, &proto.GetProductRequest{Id: id})
	// TODO: Return product and error
	return nil, nil
}

// ListProducts fetches all products matching the filter, consuming the stream.
func (c *ProductClient) ListProducts(ctx context.Context, category string) ([]*proto.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &proto.ListProductsRequest{Category: category}
	// TODO: Call c.client.ListProducts(ctx, req) — returns a stream
	// TODO: Loop: stream.Recv() until io.EOF
	// TODO: Append each received product to result slice
	// TODO: Return result

	_ = req
	_ = io.EOF
	_ = log.Printf
	return nil, nil
}

// CreateProduct creates a new product.
func (c *ProductClient) CreateProduct(ctx context.Context, req *proto.CreateProductRequest) (*proto.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// TODO: Call c.client.CreateProduct(ctx, req)
	return nil, nil
}

// DeleteProduct deletes a product by ID.
// Returns (success bool, message string, error).
func (c *ProductClient) DeleteProduct(ctx context.Context, id int32) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// TODO: Call c.client.DeleteProduct(ctx, &proto.DeleteProductRequest{Id: id})
	// TODO: Return resp.Success, resp.Message, nil
	return false, "", nil
}

// IsNotFound checks if a gRPC error is a NotFound status.
func IsNotFound(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.NotFound
	}
	return false
}
*/

// Placeholder to keep the package compilable before proto generation.
type placeholder struct{}
