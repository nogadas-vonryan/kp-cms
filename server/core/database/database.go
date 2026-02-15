package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"kpcms/server/core/inhabitant"

	_ "modernc.org/sqlite"
)

// Database manages SQL database connections.
type Database struct {
	AppDB  *sql.DB
	AuthDB *sql.DB
}

// New creates a new Database instance with separate app and auth databases.
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
		AppDB:  appDB,
		AuthDB: authDB,
	}, nil
}

// Close closes both database connections.
func (d *Database) Close() error {
	if d == nil {
		return nil
	}
	var appErr, authErr error
	if d.AppDB != nil {
		appErr = d.AppDB.Close()
	}
	if d.AuthDB != nil {
		authErr = d.AuthDB.Close()
	}
	if appErr != nil {
		return appErr
	}
	return authErr
}

// ReopenAppDB closes the current AppDB connection and opens a new one to the same path.
// This is useful after backup restore to ensure we're reading from the newly extracted database file.
func (d *Database) ReopenAppDB(appPath string) error {
	if d == nil {
		return fmt.Errorf("database instance is nil")
	}

	// Close the existing connection
	if d.AppDB != nil {
		if err := d.AppDB.Close(); err != nil {
			return fmt.Errorf("failed to close existing app database connection: %w", err)
		}
	}

	// Open a new connection
	appDB, err := sql.Open("sqlite", appPath)
	if err != nil {
		return fmt.Errorf("failed to reopen app sqlite: %w", err)
	}

	appDB.SetMaxOpenConns(1)

	if err := appDB.Ping(); err != nil {
		appDB.Close()
		return fmt.Errorf("failed to ping reopened app sqlite: %w", err)
	}

	if err := initAppSchema(appDB); err != nil {
		appDB.Close()
		return fmt.Errorf("failed to initialize schema in reopened database: %w", err)
	}

	d.AppDB = appDB
	return nil
}

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

// initAppSchema initializes the app database schema.
func initAppSchema(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	return inhabitant.InitSchema(db)
}

// initAuthSchema initializes the auth database schema.
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

-- Generic OAuth tokens table (replaces calendar_tokens)
CREATE TABLE IF NOT EXISTS oauth_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	provider TEXT NOT NULL DEFAULT 'google',
	access_token TEXT NOT NULL,
	refresh_token TEXT NOT NULL,
	token_type TEXT DEFAULT 'Bearer',
	scopes TEXT DEFAULT '["https://www.googleapis.com/auth/userinfo.email"]',
	expiry DATETIME NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(user_id, provider)
);

-- Generic OAuth states table (replaces calendar_oauth_states)
CREATE TABLE IF NOT EXISTS oauth_states (
	state TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	provider TEXT NOT NULL DEFAULT 'google',
	scopes TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user 
ON oauth_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_tokens_provider 
ON oauth_tokens(provider);

CREATE INDEX IF NOT EXISTS idx_oauth_states_expiry 
ON oauth_states(expires_at);

-- Legacy tables kept for migration compatibility
CREATE TABLE IF NOT EXISTS calendar_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	calendar_id TEXT,
	access_token TEXT NOT NULL,
	refresh_token TEXT NOT NULL,
	token_type TEXT DEFAULT 'Bearer',
	expiry DATETIME NOT NULL,
	is_primary BOOLEAN DEFAULT FALSE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(user_id, calendar_id)
);

CREATE TABLE IF NOT EXISTS calendar_oauth_states (
	state TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_calendar_tokens_user 
ON calendar_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_calendar_oauth_states_expiry 
ON calendar_oauth_states(expires_at);
`); err != nil {
		return fmt.Errorf("init auth schema: %w", err)
	}

	return nil
}
