package inhabitant

import (
	"context"
	"time"
)

// Service provides business logic for inhabitant operations.
type Service struct {
	store InhabitantStore
}

// NewService creates a new inhabitant service with a file-based store.
func NewService(store InhabitantStore) *Service {
	return &Service{store: store}
}

// Create adds a new inhabitant and returns it.
func (s *Service) Create(ctx context.Context, inhabitant *Inhabitant) (*Inhabitant, error) {
	return s.store.Create(ctx, inhabitant)
}

// GetByUUID retrieves an inhabitant by UUID.
func (s *Service) GetByUUID(ctx context.Context, uuid string) (*Inhabitant, error) {
	return s.store.GetByUUID(ctx, uuid)
}

// GetByCode retrieves an inhabitant by code (e.g., "001-26").
func (s *Service) GetByCode(ctx context.Context, code string) (*Inhabitant, error) {
	return s.store.GetByCode(ctx, code)
}

// List retrieves a paginated list of inhabitants.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*Inhabitant, error) {
	return s.store.List(ctx, offset, limit)
}

// Update modifies an existing inhabitant.
func (s *Service) Update(ctx context.Context, uuid string, inhabitant *Inhabitant) (*Inhabitant, error) {
	return s.store.Update(ctx, uuid, inhabitant)
}

// Delete removes an inhabitant.
func (s *Service) Delete(ctx context.Context, uuid string) error {
	return s.store.Delete(ctx, uuid)
}

// Search performs a full-text search on inhabitant fields.
func (s *Service) Search(ctx context.Context, query string) ([]*Inhabitant, error) {
	return s.store.Search(ctx, query)
}

// FindByName searches for inhabitants by name.
func (s *Service) FindByName(ctx context.Context, name string) ([]*Inhabitant, error) {
	return s.store.FindByName(ctx, name)
}

// FindPeopleByName searches for inhabitants by partial name match.
// This method is used by the search aggregator for cross-domain searches.
func (s *Service) FindPeopleByName(ctx context.Context, query string, limit int) ([]Inhabitant, error) {
	results, err := s.store.FindByName(ctx, query)
	if err != nil {
		return nil, err
	}
	// Convert to slice of Inhabitant (not pointer)
	inhabitants := make([]Inhabitant, 0, len(results))
	for _, inh := range results {
		if len(inhabitants) >= limit {
			break
		}
		inhabitants = append(inhabitants, *inh)
	}
	return inhabitants, nil
}

// ReloadCache reloads the in-memory cache from disk.
func (s *Service) ReloadCache(ctx context.Context) error {
	return s.store.ReloadCache(ctx)
}

// CreatedAt returns the created timestamp for an inhabitant
func CreatedAt(inh *Inhabitant) time.Time {
	if inh != nil {
		return inh.CreatedAt
	}
	return time.Time{}
}

// UpdatedAt returns the updated timestamp for an inhabitant
func UpdatedAt(inh *Inhabitant) time.Time {
	if inh != nil {
		return inh.UpdatedAt
	}
	return time.Time{}
}
