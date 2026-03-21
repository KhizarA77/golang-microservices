package domain

import (
	"errors"
	"time"
)

// User is the core domain entity.
// It has no dependencies on external packages.
type User struct {
	ID        string
	Email     string
	Name      string
	Password  string // stored hashed
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate returns an error if the user entity is in an invalid state.
func (u *User) Validate() error {
	// TODO: Check Name is not empty
	if u.Name == "" {
		return errors.New("Name cant be empty")
	}
	// TODO: Check Email is not empty (basic check — no regex needed)
	if u.Email == "" {
		return errors.New("Email cant be empty")
	}
	return nil
}

// =============================================================================
// Domain Errors
// These are defined in the domain layer so use cases can return them and
// handler layer can translate them to HTTP status codes.
// =============================================================================

// ErrUserNotFound is returned when a user cannot be found by ID or email.
type ErrUserNotFound struct {
	Identifier string
}

func (e ErrUserNotFound) Error() string {
	return "user not found: " + e.Identifier
}

// ErrEmailAlreadyExists is returned when trying to register with a taken email.
type ErrEmailAlreadyExists struct {
	Email string
}

func (e ErrEmailAlreadyExists) Error() string {
	return "email already registered: " + e.Email
}

// ErrInvalidCredentials is returned on login with wrong password.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrValidation is returned when domain validation fails.
type ErrValidation struct {
	Message string
}

func (e ErrValidation) Error() string {
	return "validation error: " + e.Message
}
