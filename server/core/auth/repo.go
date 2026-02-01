package auth

import (
	"context"
	"time"
)

// Repository defines the interface for auth data persistence operations.
type Repository interface {
	// Authenticate verifies user credentials and returns their identity if valid.
	Authenticate(ctx context.Context, username, password string) (Identity, error)

	// CreateSession creates a new session token for the given identity with the specified TTL.
	CreateSession(ctx context.Context, identity Identity, ttl time.Duration) (string, error)

	// GetSession retrieves the identity associated with the given session token.
	GetSession(ctx context.Context, token string) (Identity, error)

	// CreateUser creates a new user with the given credentials and role.
	CreateUser(ctx context.Context, username, password string, role Role) error

	// ListUsers returns all users in the system.
	ListUsers(ctx context.Context) ([]UserSummary, error)

	// UpdateRole updates the role for a specific user.
	UpdateRole(ctx context.Context, username string, role Role) error

	// DeleteUser removes a user from the system.
	DeleteUser(ctx context.Context, username string) error

	// SetPassword updates the password for a specific user.
	SetPassword(ctx context.Context, username, password string) error

	// DeleteSession invalidates a session token.
	DeleteSession(ctx context.Context, token string) error
}
