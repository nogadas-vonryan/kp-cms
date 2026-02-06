package inhabitant

import (
	"context"
)

// InhabitantStore defines the interface for file-based inhabitant storage.
type InhabitantStore interface {
	// Create adds a new inhabitant and returns it with generated UUID and code.
	Create(ctx context.Context, inh *Inhabitant) (*Inhabitant, error)

	// GetByUUID retrieves an inhabitant by UUID.
	GetByUUID(ctx context.Context, uuid string) (*Inhabitant, error)

	// GetByCode retrieves an inhabitant by code (e.g., "001-26").
	// Critical for linking documents to inhabitants.
	GetByCode(ctx context.Context, code string) (*Inhabitant, error)

	// Update modifies an existing inhabitant.
	Update(ctx context.Context, uuid string, inh *Inhabitant) (*Inhabitant, error)

	// Delete removes an inhabitant by UUID.
	Delete(ctx context.Context, uuid string) error

	// List retrieves a paginated list of inhabitants.
	List(ctx context.Context, offset int, limit int) ([]*Inhabitant, error)

	// Search performs a full-text search on inhabitant fields.
	Search(ctx context.Context, query string) ([]*Inhabitant, error)

	// FindByName searches for inhabitants by name (for migration matching).
	FindByName(ctx context.Context, name string) ([]*Inhabitant, error)

	// ReloadCache reloads the in-memory cache from disk.
	ReloadCache(ctx context.Context) error
}

// CacheStore interface for cache operations.
type CacheStore interface {
	ReloadCache(ctx context.Context) error
}
