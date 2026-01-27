package auth

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Role     Role   `json:"role"`
}

type UserSummary struct {
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

// UserStore is a simple in-memory store with mutex protection.
type UserStore struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewUserStore() *UserStore {
	return &UserStore{users: make(map[string]User)}
}

func (s *UserStore) AddUser(username, password string, role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.users[username] = User{Username: username, Password: string(hash), Role: role}
	return nil
}

func (s *UserStore) Authenticate(username, password string) (User, error) {
	s.mu.RLock()
	user, exists := s.users[username]
	s.mu.RUnlock()

	if !exists {
		return User{}, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserStore) GetUser(username string) (User, error) {
	s.mu.RLock()
	user, exists := s.users[username]
	s.mu.RUnlock()

	if !exists {
		return User{}, ErrUserNotFound
	}

	return user, nil
}

func (s *UserStore) DeleteUser(username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; !exists {
		return ErrUserNotFound
	}

	delete(s.users, username)
	return nil
}

func (s *UserStore) UpdateRole(username string, role Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return ErrUserNotFound
	}

	user.Role = role
	s.users[username] = user
	return nil
}

func (s *UserStore) SetPassword(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[username]
	if !exists {
		return ErrUserNotFound
	}

	user.Password = string(hash)
	s.users[username] = user
	return nil
}

func (s *UserStore) ListUsers() []UserSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]UserSummary, 0, len(s.users))
	for _, user := range s.users {
		result = append(result, UserSummary{Username: user.Username, Role: user.Role})
	}
	return result
}
