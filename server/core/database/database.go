package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"kpcms/server/core/auth"
	"kpcms/server/core/inhabitant"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const (
	defaultTokenLength = 32
)

// GetAuthDBPath returns the platform-specific path for the auth database
// Windows: %APPDATA%/kpcms/auth.db
// Linux/Mac: ~/.config/kpcms/auth.db
func GetAuthDBPath() (string, error) {
	var basePath string

	switch runtime.GOOS {
	case "windows":
		// Use %APPDATA% on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		basePath = filepath.Join(appData, "kpcms")
	case "linux", "darwin":
		// Use ~/.config on Linux/Mac
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		basePath = filepath.Join(homeDir, ".config", "kpcms")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return "", fmt.Errorf("failed to create auth database directory: %w", err)
	}

	return filepath.Join(basePath, "auth.db"), nil
}

type Database struct {
	appDB           *sql.DB
	authDB          *sql.DB
	InhabitantStore *inhabitant.Store
}

// New creates a new Database instance with separate app and auth databases
// appPath: path to app.db (typically in data directory)
// authPath: path to auth.db (use GetAuthDBPath() for platform-specific default)
func New(appPath, authPath string) (*Database, error) {
	// Open app database
	appDB, err := sql.Open("sqlite", appPath)
	if err != nil {
		return nil, fmt.Errorf("open app sqlite: %w", err)
	}

	appDB.SetMaxOpenConns(1)

	if err := appDB.Ping(); err != nil {
		appDB.Close()
		return nil, fmt.Errorf("ping app sqlite: %w", err)
	}

	if err := initAppSchema(appDB); err != nil {
		appDB.Close()
		return nil, err
	}

	// Open auth database
	authDB, err := sql.Open("sqlite", authPath)
	if err != nil {
		appDB.Close()
		return nil, fmt.Errorf("open auth sqlite: %w", err)
	}

	authDB.SetMaxOpenConns(1)

	if err := authDB.Ping(); err != nil {
		appDB.Close()
		authDB.Close()
		return nil, fmt.Errorf("ping auth sqlite: %w", err)
	}

	if err := initAuthSchema(authDB); err != nil {
		appDB.Close()
		authDB.Close()
		return nil, err
	}

	return &Database{
		appDB:           appDB,
		authDB:          authDB,
		InhabitantStore: inhabitant.NewStore(appDB),
	}, nil
}

func (d *Database) Close() error {
	if d == nil {
		return nil
	}
	var appErr, authErr error
	if d.appDB != nil {
		appErr = d.appDB.Close()
	}
	if d.authDB != nil {
		authErr = d.authDB.Close()
	}
	if appErr != nil {
		return appErr
	}
	return authErr
}

func initAppSchema(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	return inhabitant.InitSchema(db)
}

func initAuthSchema(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	role TEXT NOT NULL,
	password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	token TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	username TEXT NOT NULL,
	expires_at TEXT NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
`); err != nil {
		return fmt.Errorf("init auth schema: %w", err)
	}

	return nil
}

func (d *Database) Authenticate(ctx context.Context, username, password string) (auth.Identity, error) {
	var hash string
	var role string

	err := d.authDB.QueryRowContext(ctx,
		`SELECT password_hash, role FROM users WHERE username = ?`,
		username,
	).Scan(&hash, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Identity{}, auth.ErrUserNotFound
		}
		return auth.Identity{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return auth.Identity{}, auth.ErrInvalidCredentials
	}

	return auth.Identity{ID: username, Role: auth.Role(role)}, nil
}

func (d *Database) CreateSession(ctx context.Context, identity auth.Identity, ttl time.Duration) (string, error) {
	userID, err := d.getUserID(ctx, identity.ID)
	if err != nil {
		return "", err
	}

	token := randomToken(defaultTokenLength)
	expiresAt := time.Now().Add(ttl).UTC().Format(time.RFC3339Nano)

	_, err = d.authDB.ExecContext(ctx,
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

func (d *Database) GetSession(ctx context.Context, token string) (auth.Identity, error) {
	var username string
	var role string
	var expiresAt string

	err := d.authDB.QueryRowContext(ctx, `
SELECT s.username, u.role, s.expires_at
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token = ?`, token).Scan(&username, &role, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Identity{}, auth.ErrInvalidCredentials
		}
		return auth.Identity{}, err
	}

	expiry, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return auth.Identity{}, err
	}

	if time.Now().After(expiry) {
		d.DeleteSession(ctx, token)
		return auth.Identity{}, auth.ErrInvalidCredentials
	}

	return auth.Identity{ID: username, Role: auth.Role(role)}, nil
}

func (d *Database) CreateUser(ctx context.Context, username, password string, role auth.Role) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = d.authDB.ExecContext(ctx,
		`INSERT INTO users (username, role, password_hash) VALUES (?, ?, ?)`,
		username,
		string(role),
		string(hash),
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return auth.ErrUserExists
		}
		return err
	}
	return nil
}

func (d *Database) ListUsers(ctx context.Context) ([]auth.UserSummary, error) {
	rows, err := d.authDB.QueryContext(ctx, `SELECT username, role FROM users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []auth.UserSummary
	for rows.Next() {
		var username string
		var role string
		if err := rows.Scan(&username, &role); err != nil {
			return nil, err
		}
		result = append(result, auth.UserSummary{Username: username, Role: auth.Role(role)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) UpdateRole(ctx context.Context, username string, role auth.Role) error {
	result, err := d.authDB.ExecContext(ctx,
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
		return auth.ErrUserNotFound
	}
	return nil
}

func (d *Database) DeleteUser(ctx context.Context, username string) error {
	result, err := d.authDB.ExecContext(ctx, `DELETE FROM users WHERE username = ?`, username)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return auth.ErrUserNotFound
	}
	return nil
}

func (d *Database) SetPassword(ctx context.Context, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := d.authDB.ExecContext(ctx,
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
		return auth.ErrUserNotFound
	}
	return nil
}

func (d *Database) DeleteSession(ctx context.Context, token string) error {
	_, err := d.authDB.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (d *Database) getUserID(ctx context.Context, username string) (int64, error) {
	var id int64
	if err := d.authDB.QueryRowContext(ctx, `SELECT id FROM users WHERE username = ?`, username).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, auth.ErrUserNotFound
		}
		return 0, err
	}
	return id, nil
}

func randomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		fallback := base64.RawURLEncoding.EncodeToString([]byte(time.Now().String()))
		if len(fallback) >= length {
			return fallback[:length]
		}
		return fallback
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
