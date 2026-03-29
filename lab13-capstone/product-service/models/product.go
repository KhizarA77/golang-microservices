package models

import "time"

// Product represents a product in the catalog.
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest is the request body for creating a product.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
}

// UpdateProductRequest uses pointer fields so we can distinguish
// "field not provided" (nil) from "field set to zero value".
type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	Category    *string  `json:"category"`
}

// ValidCategories lists all allowed product categories.
var ValidCategories = map[string]bool{
	"electronics": true,
	"clothing":    true,
	"food":        true,
	"books":       true,
}

// Validate validates the create request.
// Returns a map of field → error message for each invalid field.
// Returns nil if everything is valid.
func (r CreateProductRequest) Validate() map[string]string {
	errors := make(map[string]string)

	// TODO: Validate Name is not empty
	if r.Name == "" {
		errors["name"] = "Name must not be empty"
	}
	// TODO: Validate Price is > 0
	if r.Price <= 0 {
		errors["price"] = "Price must be > 0"
	}
	// TODO: Validate Stock is >= 0
	if r.Stock < 0 {
		errors["stock"] = "Must be >= 0"
	}
	// TODO: Validate Category is one of ValidCategories
	_, ok := ValidCategories[r.Category]
	if !ok {
		errors["category"] = "must be one of: electronics, clothing, food, books"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ListResponse wraps a paginated list of products.
type ListResponse struct {
	Data  []*Product `json:"data"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}
