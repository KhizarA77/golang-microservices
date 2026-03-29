package usecase_test

import (
	"errors"
	"testing"
	"user-service/domain"
	"user-service/repository"
	"user-service/usecase"
)

// newUseCase creates a fresh UserUseCase backed by an in-memory repository.
func newUseCase() *usecase.UserUseCase {
	repo := repository.NewInMemoryUserRepository()
	return usecase.NewUserUseCase(repo)
}

func TestRegister_Success(t *testing.T) {
	uc := newUseCase()

	user, err := uc.Register("Alice", "alice@example.com", "password123")
	// TODO: Assert err == nil
	// TODO: Assert user != nil
	// TODO: Assert user.ID != ""
	// TODO: Assert user.Name == "Alice"
	// TODO: Assert user.Email == "alice@example.com"
	// TODO: Assert user.Active == true
	_ = user
	_ = err
}

func TestRegister_DuplicateEmail(t *testing.T) {
	uc := newUseCase()

	_, err := uc.Register("Alice", "alice@example.com", "pass1")
	if err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}

	_, err = uc.Register("Alice2", "alice@example.com", "pass2")
	// TODO: Assert err is not nil
	// TODO: Assert errors.As(err, &domain.ErrEmailAlreadyExists{}) is true

	var alreadyExists domain.ErrEmailAlreadyExists
	_ = errors.As(err, &alreadyExists)
	_ = err
}

func TestLogin_Success(t *testing.T) {
	uc := newUseCase()

	_, err := uc.Register("Bob", "bob@example.com", "mypassword")
	if err != nil {
		t.Fatal(err)
	}

	user, err := uc.Login("bob@example.com", "mypassword")
	// TODO: Assert err == nil
	// TODO: Assert user.Email == "bob@example.com"
	_ = user
	_ = err
}

func TestLogin_WrongPassword(t *testing.T) {
	uc := newUseCase()

	_, err := uc.Register("Carol", "carol@example.com", "correct")
	if err != nil {
		t.Fatal(err)
	}

	_, err = uc.Login("carol@example.com", "wrong")
	// TODO: Assert err == domain.ErrInvalidCredentials
	_ = err
}

func TestGetUser_NotFound(t *testing.T) {
	uc := newUseCase()

	_, err := uc.GetUser("nonexistent-id")
	// TODO: Assert errors.As(err, &domain.ErrUserNotFound{}) is true

	var notFound domain.ErrUserNotFound
	_ = errors.As(err, &notFound)
	_ = err
}

func TestUpdateUser(t *testing.T) {
	uc := newUseCase()

	user, err := uc.Register("Dave", "dave@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := uc.UpdateUser(user.ID, "David")
	// TODO: Assert err == nil
	// TODO: Assert updated.Name == "David"
	_ = updated
	_ = err
}

func TestDeleteUser(t *testing.T) {
	uc := newUseCase()

	user, err := uc.Register("Eve", "eve@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	err = uc.DeleteUser(user.ID)
	// TODO: Assert err == nil

	_, err = uc.GetUser(user.ID)
	// TODO: Assert err is ErrUserNotFound
	_ = err
}
