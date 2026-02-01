package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SQLRepository is a SQL-based implementation of Repository.
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository creates a new SQL-based auth repository.
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// Authenticate verifies user credentials and returns their identity if valid.
func (r *SQLRepository) Authenticate(ctx context.Context, username, password string) (Identity, error) {
	var hash string
	var role string

	err := r.db.QueryRowContext(ctx,
		`SELECT password_hash, role FROM users WHERE username = ?`,
		username,
	).Scan(&hash, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Identity{}, ErrUserNotFound
		}
		return Identity{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return Identity{}, ErrInvalidCredentials
	}

	return Identity{ID: username, Role: Role(role)}, nil
}

// CreateSession creates a new session token for the given identity with the specified TTL.
func (r *SQLRepository) CreateSession(ctx context.Context, identity Identity, ttl time.Duration) (string, error) {
	userID, err := r.getUserID(ctx, identity.ID)
	if err != nil {
		return "", err
	}

	token := NewSessionToken(32)
	expiresAt := time.Now().Add(ttl).UTC().Format(time.RFC3339Nano)

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, username, expires_at) VALUES (?, ?, ?, ?)`,
		token,
		userID,
		identity.ID,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetSession retrieves the identity associated with the given session token.
func (r *SQLRepository) GetSession(ctx context.Context, token string) (Identity, error) {
	var username string
	var role string
	var expiresAt string

	err := r.db.QueryRowContext(ctx, `
SELECT s.username, u.role, s.expires_at
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token = ?`, token).Scan(&username, &role, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Identity{}, ErrInvalidCredentials
		}
		return Identity{}, err
	}

	expiry, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return Identity{}, err
	}

	if time.Now().After(expiry) {
		_ = r.DeleteSession(ctx, token)
		return Identity{}, ErrInvalidCredentials
	}

	return Identity{ID: username, Role: Role(role)}, nil
}

// CreateUser creates a new user with the given credentials and role.
func (r *SQLRepository) CreateUser(ctx context.Context, username, password string, role Role) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO users (username, role, password_hash) VALUES (?, ?, ?)`,
		username,
		string(role),
		string(hash),
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

// ListUsers returns all users in the system.
func (r *SQLRepository) ListUsers(ctx context.Context) ([]UserSummary, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT username, role FROM users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserSummary
	for rows.Next() {
		var username string
		var role string
		if err := rows.Scan(&username, &role); err != nil {
			return nil, err
		}
		result = append(result, UserSummary{Username: username, Role: Role(role)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateRole updates the role for a specific user.
func (r *SQLRepository) UpdateRole(ctx context.Context, username string, role Role) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = ? WHERE username = ?`,
		string(role),
		username,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser removes a user from the system.
func (r *SQLRepository) DeleteUser(ctx context.Context, username string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE username = ?`, username)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetPassword updates the password for a specific user.
func (r *SQLRepository) SetPassword(ctx context.Context, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE username = ?`,
		string(hash),
		username,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteSession invalidates a session token.
func (r *SQLRepository) DeleteSession(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// getUserID retrieves the user ID for a given username.
func (r *SQLRepository) getUserID(ctx context.Context, username string) (int64, error) {
	var id int64
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = ?`, username).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return id, nil
}

// isUniqueConstraint checks if an error is due to a unique constraint violation.
func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
