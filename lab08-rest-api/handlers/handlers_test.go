package handlers_test

import (
	"bytes"
	"encoding/json"
	"lab08/handlers"
	"lab08/models"
	"lab08/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer creates a test server with a fresh store.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s := store.NewProductStore()
	h := handlers.NewProductHandler(s)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

// TestListProducts tests GET /api/products returns 200 with a list.
func TestListProducts(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/products")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body models.ListResponse
	// TODO: Assert resp.StatusCode == 200
	assertStatus(t, resp, 200)
	// TODO: Decode body into models.ListResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Assert len(data.Data) > 0
	if len(body.Data) <= 0 {
		t.Fatal("Empty response")
	}
}

// TestGetProductFound tests GET /api/products/1 returns 200.
func TestGetProductFound(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/products/1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// TODO: Assert resp.StatusCode == 200
	assertStatus(t, resp, 200)
	// TODO: Decode body into models.Product
	var body models.Product
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Assert product.ID == 1
	if body.ID != 1 {
		t.Fatal("ID not equal to 1")
	}
}

// TestGetProductNotFound tests GET /api/products/999 returns 404.
func TestGetProductNotFound(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/products/999")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// TODO: Assert resp.StatusCode == 404
	assertStatus(t, resp, 404)
}

// TestCreateProductValid tests POST /api/products with valid body returns 201.
func TestCreateProductValid(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	body := models.CreateProductRequest{
		Name:     "New Laptop",
		Price:    1200.00,
		Stock:    5,
		Category: "electronics",
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := http.Post(
		srv.URL+"/api/products",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// TODO: Assert resp.StatusCode == 201
	assertStatus(t, resp, 201)
	// TODO: Decode body into models.Product
	var product models.Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Assert product.Name == "New Laptop"
	if product.Name != "New Laptop" {
		t.Fatal("Product name mismatch")
	}
	// TODO: Assert resp.Header.Get("Location") != ""
	if resp.Header.Get("Location") == "" {
		t.Fatal("Location is empty")
	}
}

// TestCreateProductInvalid tests POST with empty name returns 400 with validation errors.
func TestCreateProductInvalid(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	body := models.CreateProductRequest{
		// Name is missing
		Price:    -1, // invalid
		Category: "unknown",
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := http.Post(
		srv.URL+"/api/products",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// TODO: Assert resp.StatusCode == 400
	assertStatus(t, resp, 400)
	// TODO: Decode body into models.ErrorResponse
	var error models.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&error)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Assert errorResp.Error == "validation_failed"
	if error.Error != "validation_failed" {
		t.Fatal("response is not validation_failed")
	}
	// TODO: Assert errorResp.Fields["name"] != ""
	if error.Fields["name"] == "" {
		t.Fatal("Field is empty")
	}
}

// TestDeleteProduct tests DELETE /api/products/1 returns 204.
func TestDeleteProduct(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/products/1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// TODO: Assert resp.StatusCode == 204
	assertStatus(t, resp, 204)
	// Verify the product is gone
	getResp, _ := http.Get(srv.URL + "/api/products/1")
	defer getResp.Body.Close()
	// TODO: Assert getResp.StatusCode == 404
	assertStatus(t, getResp, 404)
}

// helper for asserting status codes
func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Errorf("expected status %d, got %d", expected, resp.StatusCode)
	}
}

var _ = assertStatus // suppress unused warning until tests are implemented
