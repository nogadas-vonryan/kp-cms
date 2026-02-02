package search

import (
	"context"
	"fmt"
	"strings"

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
// It searches for inhabitants by name, then finds all documents where those inhabitants
// appear as complainants or respondents.
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

	// Step 3: Generate all search key variations and deduplicate
	searchKeysMap := make(map[string]struct{})
	for _, inh := range inhabitants {
		keys := generateSearchKeys(inh)
		for _, key := range keys {
			searchKeysMap[key] = struct{}{}
		}
	}

	// Convert deduplicated map to slice
	searchKeys := make([]string, 0, len(searchKeysMap))
	for key := range searchKeysMap {
		searchKeys = append(searchKeys, key)
	}

	// Step 4: Search documents containing any of these names
	documents, err := a.documentService.SearchByParticipants(ctx, searchKeys)
	if err != nil {
		return nil, fmt.Errorf("searching documents: %w", err)
	}

	// Step 5: Return combined results
	return &SearchResult{
		Query:       query,
		Inhabitants: inhabitants,
		Documents:   documents,
	}, nil
}

// buildFullName constructs a full name from an inhabitant's name parts.
func buildFullName(inh inhabitant.Inhabitant) string {
	parts := []string{}

	if inh.FirstName != "" {
		parts = append(parts, inh.FirstName)
	}
	if inh.MiddleName != "" {
		parts = append(parts, inh.MiddleName)
	}
	if inh.LastName != "" {
		parts = append(parts, inh.LastName)
	}
	if inh.Suffix != "" {
		parts = append(parts, inh.Suffix)
	}

	return strings.Join(parts, " ")
}

// generateSearchKeys creates multiple search key variations for an inhabitant.
// This includes: Full Name (John Dabba Doe), Short Name (John Doe), and Formal (Doe, John).
func generateSearchKeys(inh inhabitant.Inhabitant) []string {
	keys := []string{}

	// Format 1: Full Name (FirstName MiddleName LastName Suffix)
	fullName := buildFullName(inh)
	if fullName != "" {
		keys = append(keys, fullName)
	}

	// Format 2: Short Name (FirstName LastName) - without middle name and suffix
	if inh.FirstName != "" && inh.LastName != "" {
		shortName := inh.FirstName + " " + inh.LastName
		keys = append(keys, shortName)
	}

	// Format 3: Formal Name (LastName, FirstName)
	if inh.LastName != "" && inh.FirstName != "" {
		formalName := inh.LastName + ", " + inh.FirstName
		keys = append(keys, formalName)
	}

	// Format 4: Formal with middle initial (LastName, FirstName M.)
	if inh.LastName != "" && inh.FirstName != "" && inh.MiddleName != "" {
		middleInitial := string([]rune(inh.MiddleName)[0]) + "."
		formalWithMiddle := inh.LastName + ", " + inh.FirstName + " " + middleInitial
		keys = append(keys, formalWithMiddle)
	}

	return keys
}
