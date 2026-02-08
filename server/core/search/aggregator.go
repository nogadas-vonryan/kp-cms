package search

import (
	"context"
	"fmt"

	"kpcms/server/core/document"
	"kpcms/server/core/inhabitant"
)

// InhabitantSearcher defines the interface for searching inhabitants.
type InhabitantSearcher interface {
	FindPeopleByName(ctx context.Context, query string, limit int) ([]inhabitant.Inhabitant, error)
}

// DocumentSearcher defines the interface for searching documents.
type DocumentSearcher interface {
	SearchByParticipants(ctx context.Context, names []string) ([]*document.Document, error)
	GetDocumentsByInhabitantID(ctx context.Context, inhabitantID int64) ([]*document.Document, error)
}

// AggregatorService orchestrates cross-domain searches between inhabitants and documents.
// It implements the Application-Level Join pattern.
type AggregatorService struct {
	inhabitantService InhabitantSearcher
	documentService   DocumentSearcher
}

// NewAggregator creates a new aggregator service with the given dependencies.
func NewAggregator(inhabitantService InhabitantSearcher, documentService DocumentSearcher) *AggregatorService {
	return &AggregatorService{
		inhabitantService: inhabitantService,
		documentService:   documentService,
	}
}

// SearchResult represents the combined result of a cross-domain search.
type SearchResult struct {
	Query       string                  `json:"query"`
	Inhabitants []inhabitant.Inhabitant `json:"inhabitants"`
	Documents   []*document.Document    `json:"documents"`
}

// ExecuteAdvancedSearch performs an application-level join between inhabitants and documents.
// It searches for inhabitants by name, then finds all documents linked to those inhabitants
// using their IDs for exact matching.
func (a *AggregatorService) ExecuteAdvancedSearch(ctx context.Context, query string, maxPeople int) (*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if maxPeople <= 0 {
		maxPeople = 20 // Default limit
	}

	// Step 1: Fetch inhabitants matching the query
	inhabitants, err := a.inhabitantService.FindPeopleByName(ctx, query, maxPeople)
	if err != nil {
		return nil, fmt.Errorf("searching inhabitants: %w", err)
	}

	// Step 2: If no inhabitants found, return early
	if len(inhabitants) == 0 {
		return &SearchResult{
			Query:       query,
			Inhabitants: []inhabitant.Inhabitant{},
			Documents:   []*document.Document{},
		}, nil
	}

	// Step 3: Find documents linked to each inhabitant by ID
	documentsMap := make(map[string]*document.Document) // UUID -> Document for deduplication
	for _, inh := range inhabitants {
		docs, err := a.documentService.GetDocumentsByInhabitantID(ctx, inh.ID)
		if err != nil {
			// Log but don't fail - continue with other inhabitants
			continue
		}

		// Deduplicate documents by UUID
		for _, doc := range docs {
			documentsMap[doc.UUID] = doc
		}
	}

	// Convert map to slice
	documents := make([]*document.Document, 0, len(documentsMap))
	for _, doc := range documentsMap {
		documents = append(documents, doc)
	}

	// Step 4: Return combined results
	return &SearchResult{
		Query:       query,
		Inhabitants: inhabitants,
		Documents:   documents,
	}, nil
}
