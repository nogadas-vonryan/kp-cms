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
	documents []*document.Document
	err       error
}

func (m *mockDocumentService) SearchByParticipants(ctx context.Context, names []string) ([]*document.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.documents, nil
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
		mockDocumentSvc := &mockDocumentService{documents: mockDocuments}

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
		mockDocumentSvc := &mockDocumentService{documents: []*document.Document{}}

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
		mockDocumentSvc := &mockDocumentService{}

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
