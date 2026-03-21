package usecase

import (
	"fmt"
	"lab09/domain"
	"time"
)

// UserUseCase contains all business logic related to users.
// It depends only on the UserRepository interface — not on any implementation.
type UserUseCase struct {
	repo UserRepository
}

// NewUserUseCase creates a new UserUseCase with the given repository.
func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

// hashPassword is a simple (non-production) password hasher for this lab.
func hashPassword(password string) string {
	return fmt.Sprintf("hashed:%s", password)
}

func checkPassword(hashed, plain string) bool {
	return hashed == hashPassword(plain)
}

// generateID generates a simple unique ID.
func generateID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}

// Register creates a new user account.
// Returns ErrEmailAlreadyExists if the email is already taken.
// Returns ErrValidation if name or email is empty.
func (uc *UserUseCase) Register(name, email, password string) (*domain.User, error) {
	// TODO: Create a User with given fields, hash the password
	hashPwd := hashPassword(password)
	user := &domain.User{
		Email:    email,
		Name:     name,
		Password: hashPwd,
	}
	err := user.Validate()
	// TODO: Call user.Validate() — return error if invalid
	if err != nil {
		return nil, err
	}
	u, err := uc.repo.FindByEmail(email)
	// TODO: Check if email already exists: uc.repo.FindByEmail(email)
	//       If found, return domain.ErrEmailAlreadyExists{Email: email}
	if u != nil {
		return nil, domain.ErrEmailAlreadyExists{Email: email}
	}
	// TODO: Generate ID, set CreatedAt/UpdatedAt/Active=true
	user.ID = generateID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	// TODO: Save user: uc.repo.Save(user)
	uc.repo.Save(user)
	// TODO: Return user
	return user, nil
}

// Login verifies credentials and returns the user on success.
// Returns ErrUserNotFound if email doesn't exist.
// Returns ErrInvalidCredentials if password is wrong.
func (uc *UserUseCase) Login(email, password string) (*domain.User, error) {
	// TODO: Find user by email: uc.repo.FindByEmail(email)
	user, err := uc.repo.FindByEmail(email)
	// TODO: If not found, return error
	if err != nil {
		return nil, domain.ErrUserNotFound{Identifier: "User not found"}
	}
	// TODO: Check password with checkPassword(user.Password, password)
	// TODO: If wrong, return domain.ErrInvalidCredentials
	if !checkPassword(user.Password, password) {
		return nil, domain.ErrInvalidCredentials
	}
	// TODO: Return user
	return user, nil
}

// GetUser retrieves a user by ID.
// Returns ErrUserNotFound if user doesn't exist.
func (uc *UserUseCase) GetUser(id string) (*domain.User, error) {
	// TODO: Return uc.repo.FindByID(id)
	return uc.repo.FindByID(id)
}

// ListUsers returns all users.
func (uc *UserUseCase) ListUsers() ([]*domain.User, error) {
	// TODO: Return uc.repo.List()
	return uc.repo.List()
}

// UpdateUser updates a user's name.
// Returns ErrUserNotFound if user doesn't exist.
func (uc *UserUseCase) UpdateUser(id, name string) (*domain.User, error) {
	// TODO: Find user by ID
	user, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, domain.ErrUserNotFound{Identifier: fmt.Sprintf("%d User not found", id)}
	}
	// TODO: Update name and UpdatedAt
	user.Name = name
	user.UpdatedAt = time.Now()
	// TODO: Call uc.repo.Update(user)
	uc.repo.Save(user)
	// TODO: Return updated user
	return user, nil
}

// DeleteUser removes a user account.
// Returns ErrUserNotFound if user doesn't exist.
func (uc *UserUseCase) DeleteUser(id string) error {
	// TODO: Return uc.repo.Delete(id)
	return uc.repo.Delete(id)
}
