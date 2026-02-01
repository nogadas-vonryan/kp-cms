package auth

import (
	"context"
	"time"
)

// Service provides business logic for authentication operations.
type Service struct {
	repo Repository
}

// NewService creates a new authentication service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Authenticate validates user credentials and returns their identity.
func (s *Service) Authenticate(ctx context.Context, username, password string) (Identity, error) {
	return s.repo.Authenticate(ctx, username, password)
}

// CreateSession creates a new session token for the given identity.
func (s *Service) CreateSession(ctx context.Context, identity Identity, ttl time.Duration) (string, error) {
	return s.repo.CreateSession(ctx, identity, ttl)
}

// GetSession retrieves the identity associated with a session token.
func (s *Service) GetSession(ctx context.Context, token string) (Identity, error) {
	return s.repo.GetSession(ctx, token)
}

// CreateUser creates a new user in the system.
func (s *Service) CreateUser(ctx context.Context, username, password string, role Role) error {
	return s.repo.CreateUser(ctx, username, password, role)
}

// ListUsers returns all users in the system.
func (s *Service) ListUsers(ctx context.Context) ([]UserSummary, error) {
	return s.repo.ListUsers(ctx)
}

// UpdateRole changes the role for a user.
func (s *Service) UpdateRole(ctx context.Context, username string, role Role) error {
	return s.repo.UpdateRole(ctx, username, role)
}

// DeleteUser removes a user from the system.
func (s *Service) DeleteUser(ctx context.Context, username string) error {
	return s.repo.DeleteUser(ctx, username)
}

// SetPassword updates a user's password.
func (s *Service) SetPassword(ctx context.Context, username, password string) error {
	return s.repo.SetPassword(ctx, username, password)
}

// DeleteSession invalidates a session token.
func (s *Service) DeleteSession(ctx context.Context, token string) error {
	return s.repo.DeleteSession(ctx, token)
}
