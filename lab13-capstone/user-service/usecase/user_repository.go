package usecase

import "user-service/domain"

// UserRepository defines the port (interface) that the use case needs.
// This interface lives in the use case layer — NOT the repository layer.
// The use case defines what it needs; the repository layer implements it.
type UserRepository interface {
	// FindByID returns a user by their unique ID.
	// Returns ErrUserNotFound if the user does not exist.
	FindByID(id string) (*domain.User, error)

	// FindByEmail returns a user by their email address.
	// Returns ErrUserNotFound if no user has that email.
	FindByEmail(email string) (*domain.User, error)

	// Save persists a new user. The user's ID must already be set.
	Save(user *domain.User) error

	// Update replaces an existing user record.
	// Returns ErrUserNotFound if the user does not exist.
	Update(user *domain.User) error

	// Delete removes a user by ID.
	// Returns ErrUserNotFound if the user does not exist.
	Delete(id string) error

	// List returns all users.
	List() ([]*domain.User, error)
}
