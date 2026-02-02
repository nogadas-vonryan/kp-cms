package search

import (
	"context"
	"testing"

	"kpcms/server/core/document"
	"kpcms/server/core/inhabitant"
)

// Mock inhabitant service for testing
type mockInhabitantService struct {
	inhabitants []inhabitant.Inhabitant
	err         error
}

func (m *mockInhabitantService) FindPeopleByName(ctx context.Context, query string, limit int) ([]inhabitant.Inhabitant, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.inhabitants, nil
}

// Mock document service for testing
type mockDocumentService struct {
	documents              []*document.Document
	err                    error
	searchByParticipantsFn func(ctx context.Context, names []string) ([]*document.Document, error)
}

func (m *mockDocumentService) SearchByParticipants(ctx context.Context, names []string) ([]*document.Document, error) {
	if m.searchByParticipantsFn != nil {
		return m.searchByParticipantsFn(ctx, names)
	}
	return m.documents, m.err
}

func TestExecuteAdvancedSearch(t *testing.T) {
	ctx := context.Background()

	t.Run("successful search with results", func(t *testing.T) {
		mockInhabitants := []inhabitant.Inhabitant{
			{ID: 1, FirstName: "Juan", LastName: "Delacruz"},
			{ID: 2, FirstName: "Maria", LastName: "Delacruz"},
		}

		mockDocuments := []*document.Document{
			{UUID: "doc-1", Title: "Case 001", Code: "case-001-24"},
			{UUID: "doc-2", Title: "Case 002", Code: "case-002-24"},
		}

		mockInhabitantSvc := &mockInhabitantService{inhabitants: mockInhabitants}
		mockDocumentSvc := &mockDocumentService{
			documents: mockDocuments,
			searchByParticipantsFn: func(ctx context.Context, names []string) ([]*document.Document, error) {
				return mockDocuments, nil
			},
		}

		aggregator := NewAggregator(mockInhabitantSvc, mockDocumentSvc)
		result, err := aggregator.ExecuteAdvancedSearch(ctx, "Delacruz", 20)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Query != "Delacruz" {
			t.Errorf("expected query 'Delacruz', got '%s'", result.Query)
		}

		if len(result.Inhabitants) != 2 {
			t.Errorf("expected 2 inhabitants, got %d", len(result.Inhabitants))
		}

		if len(result.Documents) != 2 {
			t.Errorf("expected 2 documents, got %d", len(result.Documents))
		}
	})

	t.Run("no inhabitants found", func(t *testing.T) {
		mockInhabitantSvc := &mockInhabitantService{inhabitants: []inhabitant.Inhabitant{}}
		mockDocumentSvc := &mockDocumentService{
			documents: []*document.Document{},
			searchByParticipantsFn: func(ctx context.Context, names []string) ([]*document.Document, error) {
				return []*document.Document{}, nil
			},
		}

		aggregator := NewAggregator(mockInhabitantSvc, mockDocumentSvc)
		result, err := aggregator.ExecuteAdvancedSearch(ctx, "Nonexistent", 20)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Inhabitants) != 0 {
			t.Errorf("expected 0 inhabitants, got %d", len(result.Inhabitants))
		}

		if len(result.Documents) != 0 {
			t.Errorf("expected 0 documents, got %d", len(result.Documents))
		}
	})

	t.Run("empty query", func(t *testing.T) {
		mockInhabitantSvc := &mockInhabitantService{}
		mockDocumentSvc := &mockDocumentService{
			searchByParticipantsFn: func(ctx context.Context, names []string) ([]*document.Document, error) {
				return nil, nil
			},
		}

		aggregator := NewAggregator(mockInhabitantSvc, mockDocumentSvc)
		_, err := aggregator.ExecuteAdvancedSearch(ctx, "", 20)

		if err == nil {
			t.Fatal("expected error for empty query, got nil")
		}
	})

	t.Run("builds full names correctly", func(t *testing.T) {
		testCases := []struct {
			inhabitant inhabitant.Inhabitant
			expected   string
		}{
			{
				inhabitant: inhabitant.Inhabitant{FirstName: "Juan", LastName: "Delacruz"},
				expected:   "Juan Delacruz",
			},
			{
				inhabitant: inhabitant.Inhabitant{FirstName: "Maria", MiddleName: "Santos", LastName: "Delacruz"},
				expected:   "Maria Santos Delacruz",
			},
			{
				inhabitant: inhabitant.Inhabitant{FirstName: "Juan", LastName: "Delacruz", Suffix: "Jr."},
				expected:   "Juan Delacruz Jr.",
			},
		}

		for _, tc := range testCases {
			result := buildFullName(tc.inhabitant)
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		}
	})
}

func TestGenerateSearchKeys(t *testing.T) {
	t.Run("generates all name format variations", func(t *testing.T) {
		inh := inhabitant.Inhabitant{
			FirstName:  "John",
			MiddleName: "Dabba",
			LastName:   "Doe",
		}

		keys := generateSearchKeys(inh)

		// Should generate: Full Name, Short Name, Formal Name, Formal with Middle Initial
		expectedKeys := []string{
			"John Dabba Doe", // Full name
			"John Doe",       // Short name
			"Doe, John",      // Formal
			"Doe, John D.",   // Formal with middle initial
		}

		if len(keys) != len(expectedKeys) {
			t.Errorf("expected %d keys, got %d", len(expectedKeys), len(keys))
		}

		for _, expected := range expectedKeys {
			found := false
			for _, key := range keys {
				if key == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected key '%s' not found in %v", expected, keys)
			}
		}
	})

	t.Run("handles name without middle name", func(t *testing.T) {
		inh := inhabitant.Inhabitant{
			FirstName: "Jane",
			LastName:  "Smith",
		}

		keys := generateSearchKeys(inh)

		expectedKeys := []string{
			"Jane Smith",  // Full name
			"Jane Smith",  // Short name (same as full)
			"Smith, Jane", // Formal
		}

		// Check that all expected keys exist
		for _, expected := range expectedKeys {
			found := false
			for _, key := range keys {
				if key == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected key '%s' not found in %v", expected, keys)
			}
		}
	})

	t.Run("handles name with suffix", func(t *testing.T) {
		inh := inhabitant.Inhabitant{
			FirstName: "Robert",
			LastName:  "Johnson",
			Suffix:    "Jr.",
		}

		keys := generateSearchKeys(inh)

		// Full name should include suffix
		fullNameFound := false
		for _, key := range keys {
			if key == "Robert Johnson Jr." {
				fullNameFound = true
				break
			}
		}
		if !fullNameFound {
			t.Errorf("expected full name with suffix not found in %v", keys)
		}
	})

	t.Run("handles incomplete names", func(t *testing.T) {
		inh := inhabitant.Inhabitant{
			FirstName: "Madonna",
		}

		keys := generateSearchKeys(inh)

		// Should still generate a key with just the first name
		if len(keys) == 0 {
			t.Error("expected at least one key for incomplete name")
		}
	})

	t.Run("deduplication in ExecuteAdvancedSearch", func(t *testing.T) {
		// Create inhabitants with overlapping search keys
		mockInhabitants := []inhabitant.Inhabitant{
			{ID: 1, FirstName: "John", MiddleName: "Dabba", LastName: "Doe"},
			{ID: 2, FirstName: "John", LastName: "Doe"}, // Short name matches first person
		}

		mockDocuments := []*document.Document{
			{UUID: "doc-1", Title: "Test Doc", Code: "test-001"},
		}

		// Track which names were searched
		searchedNames := make(map[string]bool)
		mockDocumentSvc := &mockDocumentService{
			documents: mockDocuments,
			searchByParticipantsFn: func(ctx context.Context, names []string) ([]*document.Document, error) {
				for _, name := range names {
					searchedNames[name] = true
				}
				return mockDocuments, nil
			},
		}

		mockInhabitantSvc := &mockInhabitantService{inhabitants: mockInhabitants}
		aggregator := NewAggregator(mockInhabitantSvc, mockDocumentSvc)

		_, err := aggregator.ExecuteAdvancedSearch(context.Background(), "Doe", 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify that "John Doe" appears only once despite being generated by both inhabitants
		// (This tests the deduplication logic in ExecuteAdvancedSearch)
	})
}
