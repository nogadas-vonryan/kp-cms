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
	documents                    []*document.Document
	err                          error
	searchByParticipantsFn       func(ctx context.Context, names []string) ([]*document.Document, error)
	getDocumentsByInhabitantIDFn func(ctx context.Context, id int64) ([]*document.Document, error)
}

func (m *mockDocumentService) SearchByParticipants(ctx context.Context, names []string) ([]*document.Document, error) {
	if m.searchByParticipantsFn != nil {
		return m.searchByParticipantsFn(ctx, names)
	}
	return m.documents, m.err
}

func (m *mockDocumentService) GetDocumentsByInhabitantID(ctx context.Context, id int64) ([]*document.Document, error) {
	if m.getDocumentsByInhabitantIDFn != nil {
		return m.getDocumentsByInhabitantIDFn(ctx, id)
	}
	return m.documents, m.err
}

func TestExecuteAdvancedSearch(t *testing.T) {
	ctx := context.Background()

	t.Run("successful search with results using ID-based lookup", func(t *testing.T) {
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
			getDocumentsByInhabitantIDFn: func(ctx context.Context, id int64) ([]*document.Document, error) {
				// Return documents for each inhabitant
				if id == 1 {
					return []*document.Document{mockDocuments[0]}, nil
				}
				if id == 2 {
					return []*document.Document{mockDocuments[1]}, nil
				}
				return []*document.Document{}, nil
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
			getDocumentsByInhabitantIDFn: func(ctx context.Context, id int64) ([]*document.Document, error) {
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
			getDocumentsByInhabitantIDFn: func(ctx context.Context, id int64) ([]*document.Document, error) {
				return nil, nil
			},
		}

		aggregator := NewAggregator(mockInhabitantSvc, mockDocumentSvc)
		_, err := aggregator.ExecuteAdvancedSearch(ctx, "", 20)

		if err == nil {
			t.Fatal("expected error for empty query, got nil")
		}
	})

	t.Run("deduplicates documents from multiple inhabitants", func(t *testing.T) {
		// Create inhabitants that both link to the same document
		mockInhabitants := []inhabitant.Inhabitant{
			{ID: 1, FirstName: "John", LastName: "Doe"},
			{ID: 2, FirstName: "Jane", LastName: "Doe"},
		}

		sharedDoc := &document.Document{UUID: "doc-1", Title: "Shared Case", Code: "case-001"}

		mockDocumentSvc := &mockDocumentService{
			getDocumentsByInhabitantIDFn: func(ctx context.Context, id int64) ([]*document.Document, error) {
				// Both inhabitants link to the same document
				return []*document.Document{sharedDoc}, nil
			},
		}

		mockInhabitantSvc := &mockInhabitantService{inhabitants: mockInhabitants}
		aggregator := NewAggregator(mockInhabitantSvc, mockDocumentSvc)

		result, err := aggregator.ExecuteAdvancedSearch(context.Background(), "Doe", 20)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify document deduplication: should have only 1 unique document
		if len(result.Documents) != 1 {
			t.Errorf("expected 1 deduplicated document, got %d", len(result.Documents))
		}

		if result.Documents[0].UUID != "doc-1" {
			t.Errorf("expected document UUID 'doc-1', got '%s'", result.Documents[0].UUID)
		}
	})
}
