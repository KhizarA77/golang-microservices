package main

import (
	"fmt"
	"lab09/handler"
	"lab09/repository"
	"lab09/usecase"
	"log"
	"net/http"
)

func main() {
	// ==========================================================================
	// Composition Root — wire all layers together here
	// The dependency flow: main -> handler -> usecase -> repository -> domain
	// ==========================================================================

	// Layer 1: Repository (implements usecase.UserRepository interface)
	repo := repository.NewInMemoryUserRepository()

	// Layer 2: Use Case (depends on UserRepository interface)
	uc := usecase.NewUserUseCase(repo)

	// Layer 3: Handler (depends on UserUseCase)
	h := handler.NewUserHandler(uc)

	// Layer 4: Router
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	fmt.Println("Clean Architecture User API on http://localhost:8080")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  POST   /api/users           Register")
	fmt.Println("  POST   /api/auth/login       Login")
	fmt.Println("  GET    /api/users            List users")
	fmt.Println("  GET    /api/users/{id}       Get user")
	fmt.Println("  PUT    /api/users/{id}       Update user")
	fmt.Println("  DELETE /api/users/{id}       Delete user")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  curl -X POST http://localhost:8080/api/users \`)
	fmt.Println(`       -H "Content-Type: application/json" \`)
	fmt.Println(`       -d '{"name":"Alice","email":"alice@example.com","password":"secret"}'`)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
