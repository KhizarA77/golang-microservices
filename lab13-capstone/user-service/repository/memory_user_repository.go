package repository

import (
	"sync"
	"user-service/domain"
)

// InMemoryUserRepository implements usecase.UserRepository using a map.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

// NewInMemoryUserRepository creates a new empty in-memory repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

// FindByID returns a user by ID.
func (r *InMemoryUserRepository) FindByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: Look up id in r.users
	user, ok := r.users[id]
	// TODO: If not found, return nil, domain.ErrUserNotFound{Identifier: id}
	if !ok {
		return nil, domain.ErrUserNotFound{Identifier: id}
	}
	// TODO: Return a copy: copy := *r.users[id]; return &copy, nil
	copy := *user

	return &copy, nil
}

// FindByEmail returns a user by email address.
func (r *InMemoryUserRepository) FindByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: Iterate r.users, find where user.Email == email
	for _, user := range r.users {
		if user.Email == email {
			copy := *user
			return &copy, nil
		}
	}
	// TODO: If not found, return nil, domain.ErrUserNotFound{Identifier: email}
	return nil, domain.ErrUserNotFound{Identifier: email}
}

// Save stores a new user.
func (r *InMemoryUserRepository) Save(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: Store a copy: cp := *user; r.users[user.ID] = &cp
	cp := *user
	r.users[user.ID] = &cp
	return nil
}

// Update replaces an existing user.
func (r *InMemoryUserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: Check user exists
	_, ok := r.users[user.ID]
	if !ok {
		return domain.ErrUserNotFound{Identifier: user.ID}
	}
	// TODO: Replace: cp := *user; r.users[user.ID] = &cp
	cp := *user
	r.users[user.ID] = &cp
	return nil
}

// Delete removes a user by ID.
func (r *InMemoryUserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: Check exists, delete, return nil
	_, ok := r.users[id]
	if !ok {
		return domain.ErrUserNotFound{Identifier: id}
	}
	delete(r.users, id)
	return nil
}

// List returns all users.
func (r *InMemoryUserRepository) List() ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: Collect all users into a slice, return it
	sli := make([]*domain.User, 0)

	for _, user := range r.users {
		sli = append(sli, user)
	}

	return sli, nil
}
