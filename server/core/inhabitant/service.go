package inhabitant

import (
	"context"
	"fmt"
	"time"
)

// Service provides business logic for inhabitant operations.
type Service struct {
	repo  Repository
	store InhabitantStore
}

// NewService creates a new inhabitant service with a repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// NewServiceWithStore creates a new inhabitant service with a file-based store.
func NewServiceWithStore(store InhabitantStore) *Service {
	return &Service{store: store}
}

// Create adds a new inhabitant and returns it.
func (s *Service) Create(ctx context.Context, inhabitant *Inhabitant) (*Inhabitant, error) {
	if s.store != nil {
		return s.store.Create(ctx, inhabitant)
	}
	if s.repo != nil {
		id, err := s.repo.Create(ctx, inhabitant)
		if err != nil {
			return nil, err
		}
		// Fetch and return the created inhabitant
		return s.repo.Get(ctx, id)
	}
	return nil, nil
}

// GetByUUID retrieves an inhabitant by UUID.
func (s *Service) GetByUUID(ctx context.Context, uuid string) (*Inhabitant, error) {
	if s.store != nil {
		return s.store.GetByUUID(ctx, uuid)
	}
	return nil, nil
}

// GetByID retrieves an inhabitant by ID (SQLite).
func (s *Service) GetByID(ctx context.Context, id int64) (*Inhabitant, error) {
	if s.repo != nil {
		return s.repo.Get(ctx, id)
	}
	return nil, nil
}

// GetByCode retrieves an inhabitant by code (e.g., "001-26").
func (s *Service) GetByCode(ctx context.Context, code string) (*Inhabitant, error) {
	if s.store != nil {
		return s.store.GetByCode(ctx, code)
	}
	return nil, nil
}

// Get retrieves an inhabitant by ID (SQLite).
func (s *Service) Get(ctx context.Context, id int64) (*Inhabitant, error) {
	if s.repo != nil {
		return s.repo.Get(ctx, id)
	}
	return nil, nil
}

// List retrieves a paginated list of inhabitants.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*Inhabitant, error) {
	if s.store != nil {
		return s.store.List(ctx, offset, limit)
	}
	if s.repo != nil {
		results, err := s.repo.List(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		// Convert to slice of pointers
		inhabitants := make([]*Inhabitant, len(results))
		for i := range results {
			inhabitants[i] = &results[i]
		}
		return inhabitants, nil
	}
	return nil, nil
}

// Update modifies an existing inhabitant.
func (s *Service) Update(ctx context.Context, uuid string, inhabitant *Inhabitant) (*Inhabitant, error) {
	if s.store != nil {
		return s.store.Update(ctx, uuid, inhabitant)
	}
	if s.repo != nil {
		// File store uses UUID, SQLite uses int64 ID
		// Return error indicating wrong method for SQLite
		return nil, fmt.Errorf("use UpdateByID for SQLite-based inhabitant storage")
	}
	return nil, fmt.Errorf("no inhabitant store or repository initialized")
}

// Delete removes an inhabitant.
func (s *Service) Delete(ctx context.Context, uuid string) error {
	if s.store != nil {
		return s.store.Delete(ctx, uuid)
	}
	return fmt.Errorf("no inhabitant store initialized for deletion")
}

// DeleteByID removes an inhabitant by ID (SQLite).
func (s *Service) DeleteByID(ctx context.Context, id int64) error {
	if s.repo != nil {
		return s.repo.Delete(ctx, id)
	}
	return nil
}

// UpdateByID updates an inhabitant by ID (SQLite).
func (s *Service) UpdateByID(ctx context.Context, id int64, inh *Inhabitant) error {
	if s.repo != nil {
		inh.ID = id
		return s.repo.Update(ctx, inh)
	}
	return nil
}

// Search performs a full-text search on inhabitant fields.
func (s *Service) Search(ctx context.Context, query string) ([]*Inhabitant, error) {
	if s.store != nil {
		return s.store.Search(ctx, query)
	}
	return nil, nil
}

// FindByName searches for inhabitants by name.
func (s *Service) FindByName(ctx context.Context, name string) ([]*Inhabitant, error) {
	if s.store != nil {
		return s.store.FindByName(ctx, name)
	}
	if s.repo != nil {
		results, err := s.repo.FindByName(ctx, name, 10)
		if err != nil {
			return nil, err
		}
		inhabitants := make([]*Inhabitant, len(results))
		for i := range results {
			inhabitants[i] = &results[i]
		}
		return inhabitants, nil
	}
	return nil, nil
}

// FindPeopleByName searches for inhabitants by partial name match.
// This method is used by the search aggregator for cross-domain searches.
func (s *Service) FindPeopleByName(ctx context.Context, query string, limit int) ([]Inhabitant, error) {
	if s.store != nil {
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
	if s.repo != nil {
		results, err := s.repo.FindByName(ctx, query, limit)
		if err != nil {
			return nil, err
		}
		return results, nil
	}
	return nil, nil
}

// ReloadCache reloads the in-memory cache from disk.
func (s *Service) ReloadCache(ctx context.Context) error {
	if s.store != nil {
		return s.store.ReloadCache(ctx)
	}
	return nil
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
