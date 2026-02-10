package inhabitant

import (
	"context"
	"database/sql"
)

// Service provides business logic for inhabitant operations.
type Service struct {
	repo Repository
}

// NewService creates a new inhabitant service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create adds a new inhabitant and returns its ID.
func (s *Service) Create(ctx context.Context, inhabitant *Inhabitant) (int64, error) {
	return s.repo.Create(ctx, inhabitant)
}

// Get retrieves an inhabitant by ID.
func (s *Service) Get(ctx context.Context, id int64) (*Inhabitant, error) {
	return s.repo.Get(ctx, id)
}

// List retrieves a paginated list of inhabitants.
func (s *Service) List(ctx context.Context, limit, offset int) ([]Inhabitant, error) {
	return s.repo.List(ctx, limit, offset)
}

// Update modifies an existing inhabitant.
func (s *Service) Update(ctx context.Context, inhabitant *Inhabitant) error {
	return s.repo.Update(ctx, inhabitant)
}

// Delete removes an inhabitant.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// FindPeopleByName searches for inhabitants whose names partially match the query.
// This method is used by the search aggregator to find people for cross-domain searches.
func (s *Service) FindPeopleByName(ctx context.Context, query string, limit int) ([]Inhabitant, error) {
	return s.repo.FindByName(ctx, query, limit)
}

// UpdateDB updates the database connection used by the repository.
// This is necessary after backup restore to reconnect to the newly extracted database.
func (s *Service) UpdateDB(db *sql.DB) {
	if s != nil && s.repo != nil {
		s.repo.UpdateDB(db)
	}
}

// Reload validates the database connection is working (e.g., after a restore operation).
// Returns nil if the connection is healthy, error otherwise.
func (s *Service) Reload(ctx context.Context) error {
	// Test the connection with a simple query
	_, err := s.repo.List(ctx, 1, 0)
	return err
}
