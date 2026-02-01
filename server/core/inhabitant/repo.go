package inhabitant

import "context"

// Repository defines the interface for inhabitant data persistence operations.
type Repository interface {
	// Create adds a new inhabitant to the repository and returns its ID.
	Create(ctx context.Context, inhabitant *Inhabitant) (int64, error)

	// Get retrieves an inhabitant by ID.
	Get(ctx context.Context, id int64) (*Inhabitant, error)

	// List retrieves a paginated list of inhabitants.
	List(ctx context.Context, limit, offset int) ([]Inhabitant, error)

	// Update modifies an existing inhabitant.
	Update(ctx context.Context, inhabitant *Inhabitant) error

	// Delete removes an inhabitant from the repository.
	Delete(ctx context.Context, id int64) error
}
