package store

import (
	"context"
	"strings"

	"kpcms/server/core/inhabitant"
)

// Search performs a full-text search on inhabitant fields.
func (s *InhabitantFileStore) Search(ctx context.Context, query string) ([]*inhabitant.Inhabitant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	var results []*inhabitant.Inhabitant

	for _, inh := range s.inhabitants {
		if matchesQuery(inh, query) {
			results = append(results, inh)
		}
	}

	return results, nil
}

// FindByName searches for inhabitants by name (for migration matching).
func (s *InhabitantFileStore) FindByName(ctx context.Context, name string) ([]*inhabitant.Inhabitant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	name = normalizeSearchName(name)
	var results []*inhabitant.Inhabitant

	// Direct lookup in name index
	if uuids, exists := s.nameIndex[name]; exists {
		for _, uuid := range uuids {
			if inh, ok := s.inhabitants[uuid]; ok {
				results = append(results, inh)
			}
		}
	}

	return results, nil
}

// matchesQuery checks if an inhabitant matches the search query.
func matchesQuery(inh *inhabitant.Inhabitant, query string) bool {
	// Check name fields
	normalizedName := normalizeName(inh)
	if strings.Contains(normalizedName, query) {
		return true
	}

	// Check individual fields
	fields := []string{
		inh.FirstName,
		inh.LastName,
		inh.MiddleName,
		inh.Suffix,
		inh.Address,
		inh.Occupation,
	}

	for _, field := range fields {
		if strings.ToLower(field) == query {
			return true
		}
	}

	return false
}

// normalizeSearchName normalizes a search name for index lookup.
func normalizeSearchName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
