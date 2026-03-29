package main

// User Service — Port 8081
//
// This service handles user registration, authentication, and profile management.
//
// INSTRUCTIONS:
// Port your Lab 09 Clean Architecture solution here.
// Additions needed for this capstone:
//   1. Login endpoint that returns a signed token string
//   2. Token format: base64("userID:email:timestamp") — simple, no crypto needed for lab
//   3. GET /users/me endpoint using token to identify caller
//
// Endpoints:
//   POST /api/users/register  → {"name":"Alice","email":"...","password":"..."}
//   POST /api/users/login     → returns {"token":"...","user":{...}}
//   GET  /api/users           → list (no auth needed for lab)
//   GET  /api/users/{id}      → get by ID
//   GET  /health
//
// TODO: Copy lab09 domain, usecase, repository, handler packages here.
//       Add login + token generation.
//       Add graceful shutdown.

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/handler"
	"user-service/repository"
	"user-service/usecase"
)

func main() {
	// Seed with a test user
	// Layer 1: Repository (implements usecase.UserRepository interface)
	repo := repository.NewInMemoryUserRepository()

	// Layer 2: Use Case (depends on UserRepository interface)
	uc := usecase.NewUserUseCase(repo)

	// Layer 3: Handler (depends on UserUseCase)
	h := handler.NewUserHandler(uc)

	// Layer 4: Router
	mux := http.NewServeMux()

	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("User Service on http://localhost:8081")
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("User Service shutting down...")
}
