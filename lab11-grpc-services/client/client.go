package client

// NOTE: Uncomment after running protoc to generate proto package.

import (
	"context"
	"io"
	"lab11/proto"
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
	prod, err := c.client.GetProduct(ctx, &proto.GetProductRequest{Id: id})
	// TODO: Return product and error
	if err != nil {
		return nil, err
	}
	return prod, nil
}

// ListProducts fetches all products matching the filter, consuming the stream.
func (c *ProductClient) ListProducts(ctx context.Context, category string) ([]*proto.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &proto.ListProductsRequest{Category: category}
	// TODO: Call c.client.ListProducts(ctx, req) — returns a stream
	stream, err := c.client.ListProducts(ctx, req)
	if err != nil {
		return nil, err
	}
	sli := make([]*proto.Product, 0)
	// TODO: Loop: stream.Recv() until io.EOF
	// TODO: Append each received product to result slice
	for {
		product, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		sli = append(sli, product)
	}
	// TODO: Return result
	return sli, nil
}

// CreateProduct creates a new product.
func (c *ProductClient) CreateProduct(ctx context.Context, req *proto.CreateProductRequest) (*proto.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// TODO: Call c.client.CreateProduct(ctx, req)
	product, err := c.client.CreateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// DeleteProduct deletes a product by ID.
// Returns (success bool, message string, error).
func (c *ProductClient) DeleteProduct(ctx context.Context, id int32) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// TODO: Call c.client.DeleteProduct(ctx, &proto.DeleteProductRequest{Id: id})
	resp, err := c.client.DeleteProduct(ctx, &proto.DeleteProductRequest{Id: id})
	if err != nil {
		return false, "", err
	}
	// TODO: Return resp.Success, resp.Message, nil
	return resp.Success, resp.Message, nil
}

// IsNotFound checks if a gRPC error is a NotFound status.
func IsNotFound(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.NotFound
	}
	return false
}

// Placeholder to keep the package compilable before proto generation.
type placeholder struct{}
