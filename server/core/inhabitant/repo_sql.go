package inhabitant

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// SQLRepository is a SQL-based implementation of Repository.
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository creates a new SQL-based inhabitant repository.
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// Create adds a new inhabitant to the repository and returns its ID.
func (r *SQLRepository) Create(ctx context.Context, inhabitant *Inhabitant) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
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

// Get retrieves an inhabitant by ID.
func (r *SQLRepository) Get(ctx context.Context, id int64) (*Inhabitant, error) {
	var firstName, lastName, middleName, suffix, contactNo, address string
	var birthdayStr string

	err := r.db.QueryRowContext(ctx, `
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

// List retrieves a paginated list of inhabitants.
func (r *SQLRepository) List(ctx context.Context, limit int, offset int) ([]Inhabitant, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
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

// Update modifies an existing inhabitant.
func (r *SQLRepository) Update(ctx context.Context, inhabitant *Inhabitant) error {
	result, err := r.db.ExecContext(ctx, `
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

// Delete removes an inhabitant from the repository.
func (r *SQLRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM inhabitants WHERE id = ?`, id)
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
