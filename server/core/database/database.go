package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"kpcms/server/core/auth"
	"kpcms/server/core/document"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const (
	defaultTokenLength = 32
)

type Database struct {
	db *sql.DB
}

func New(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

func initSchema(db *sql.DB) error {
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

CREATE TABLE IF NOT EXISTS inhabitants (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	middle_name TEXT,
	suffix TEXT,
	birthday TEXT,
	contact_no TEXT,
	address TEXT
);
`); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}

	return nil
}

func (d *Database) Authenticate(ctx context.Context, username, password string) (auth.Identity, error) {
	var hash string
	var role string

	err := d.db.QueryRowContext(ctx,
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

	_, err = d.db.ExecContext(ctx,
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

	err := d.db.QueryRowContext(ctx, `
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
		_ = d.DeleteSession(ctx, token)
		return auth.Identity{}, auth.ErrInvalidCredentials
	}

	return auth.Identity{ID: username, Role: auth.Role(role)}, nil
}

func (d *Database) CreateUser(ctx context.Context, username, password string, role auth.Role) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx,
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
	rows, err := d.db.QueryContext(ctx, `SELECT username, role FROM users ORDER BY username ASC`)
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
	result, err := d.db.ExecContext(ctx,
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
	result, err := d.db.ExecContext(ctx, `DELETE FROM users WHERE username = ?`, username)
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

	result, err := d.db.ExecContext(ctx,
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
	_, err := d.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (d *Database) getUserID(ctx context.Context, username string) (int64, error) {
	var id int64
	if err := d.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = ?`, username).Scan(&id); err != nil {
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

// CreateInhabitant adds a new inhabitant to the database
func (d *Database) CreateInhabitant(ctx context.Context, inhabitant *document.Inhabitant) (int64, error) {
	result, err := d.db.ExecContext(ctx, `
		INSERT INTO inhabitants (first_name, last_name, middle_name, suffix, birthday, contact_no, address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, inhabitant.FirstName, inhabitant.LastName, inhabitant.MiddleName, inhabitant.Suffix,
		inhabitant.Birthday.Format(time.RFC3339), inhabitant.ContactNo, inhabitant.Address)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetInhabitant retrieves an inhabitant by ID
func (d *Database) GetInhabitant(ctx context.Context, id int64) (*document.Inhabitant, error) {
	var firstName, lastName, middleName, suffix, contactNo, address string
	var birthdayStr string

	err := d.db.QueryRowContext(ctx, `
		SELECT first_name, last_name, middle_name, suffix, birthday, contact_no, address
		FROM inhabitants WHERE id = ?
	`, id).Scan(&firstName, &lastName, &middleName, &suffix, &birthdayStr, &contactNo, &address)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("inhabitant not found")
		}
		return nil, err
	}

	inhabitant := &document.Inhabitant{
		ID:         id,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		Suffix:     suffix,
		ContactNo:  contactNo,
		Address:    address,
	}

	if birthdayStr != "" {
		birthday, err := time.Parse(time.RFC3339, birthdayStr)
		if err == nil {
			inhabitant.Birthday = birthday
		}
	}

	return inhabitant, nil
}

// ListInhabitants retrieves all inhabitants with optional pagination
func (d *Database) ListInhabitants(ctx context.Context, limit int, offset int) ([]document.Inhabitant, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := d.db.QueryContext(ctx, `
		SELECT id, first_name, last_name, middle_name, suffix, birthday, contact_no, address
		FROM inhabitants
		ORDER BY last_name ASC, first_name ASC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inhabitants []document.Inhabitant
	for rows.Next() {
		var id int64
		var firstName, lastName, middleName, suffix, contactNo, address string
		var birthdayStr string

		if err := rows.Scan(&id, &firstName, &lastName, &middleName, &suffix, &birthdayStr, &contactNo, &address); err != nil {
			return nil, err
		}

		inhabitant := document.Inhabitant{
			ID:         id,
			FirstName:  firstName,
			LastName:   lastName,
			MiddleName: middleName,
			Suffix:     suffix,
			ContactNo:  contactNo,
			Address:    address,
		}

		if birthdayStr != "" {
			birthday, err := time.Parse(time.RFC3339, birthdayStr)
			if err == nil {
				inhabitant.Birthday = birthday
			}
		}

		inhabitants = append(inhabitants, inhabitant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return inhabitants, nil
}

// UpdateInhabitant updates an existing inhabitant
func (d *Database) UpdateInhabitant(ctx context.Context, inhabitant *document.Inhabitant) error {
	result, err := d.db.ExecContext(ctx, `
		UPDATE inhabitants
		SET first_name = ?, last_name = ?, middle_name = ?, suffix = ?, birthday = ?, contact_no = ?, address = ?
		WHERE id = ?
	`, inhabitant.FirstName, inhabitant.LastName, inhabitant.MiddleName, inhabitant.Suffix,
		inhabitant.Birthday.Format(time.RFC3339), inhabitant.ContactNo, inhabitant.Address, inhabitant.ID)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("inhabitant not found")
	}

	return nil
}

// DeleteInhabitant removes an inhabitant from the database
func (d *Database) DeleteInhabitant(ctx context.Context, id int64) error {
	result, err := d.db.ExecContext(ctx, `DELETE FROM inhabitants WHERE id = ?`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("inhabitant not found")
	}

	return nil
}
