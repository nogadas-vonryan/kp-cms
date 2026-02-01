package inhabitant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func InitSchema(db *sql.DB) error {
	if _, err := db.Exec(`
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

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateInhabitant adds a new inhabitant to the database
func (s *Store) CreateInhabitant(ctx context.Context, inhabitant *Inhabitant) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
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
func (s *Store) GetInhabitant(ctx context.Context, id int64) (*Inhabitant, error) {
	var firstName, lastName, middleName, suffix, contactNo, address string
	var birthdayStr string

	err := s.db.QueryRowContext(ctx, `
		SELECT first_name, last_name, middle_name, suffix, birthday, contact_no, address
		FROM inhabitants WHERE id = ?
	`, id).Scan(&firstName, &lastName, &middleName, &suffix, &birthdayStr, &contactNo, &address)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("inhabitant not found")
		}
		return nil, err
	}

	inhabitant := &Inhabitant{
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
func (s *Store) ListInhabitants(ctx context.Context, limit int, offset int) ([]Inhabitant, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, first_name, last_name, middle_name, suffix, birthday, contact_no, address
		FROM inhabitants
		ORDER BY last_name ASC, first_name ASC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inhabitants []Inhabitant
	for rows.Next() {
		var id int64
		var firstName, lastName, middleName, suffix, contactNo, address string
		var birthdayStr string

		if err := rows.Scan(&id, &firstName, &lastName, &middleName, &suffix, &birthdayStr, &contactNo, &address); err != nil {
			return nil, err
		}

		inhabitant := Inhabitant{
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
func (s *Store) UpdateInhabitant(ctx context.Context, inhabitant *Inhabitant) error {
	result, err := s.db.ExecContext(ctx, `
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
func (s *Store) DeleteInhabitant(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM inhabitants WHERE id = ?`, id)
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
