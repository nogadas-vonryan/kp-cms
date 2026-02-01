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
`); err != nil {
		return fmt.Errorf("init auth schema: %w", err)
	}

	return nil
}
