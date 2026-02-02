package document

import (
	"context"
	"errors"
	"sync"
	"testing"
)

var ErrDocumentNotFound = errors.New("document not found")

// Mock document store for SearchByParticipants testing
type mockSearchStore struct {
	DocumentStore
	searchFunc func(ctx context.Context, criteria SearchCriteria) ([]*Document, error)
	mu         sync.Mutex
	callCount  int
}

func (m *mockSearchStore) Search(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()

	if m.searchFunc != nil {
		return m.searchFunc(ctx, criteria)
	}
	return []*Document{}, nil
}

func TestSearchByParticipants(t *testing.T) {
	ctx := context.Background()

	t.Run("returns empty for no names", func(t *testing.T) {
		mockStore := &mockSearchStore{}
		service := NewDocumentService(mockStore, nil, nil, nil)

		results, err := service.SearchByParticipants(ctx, []string{})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("searches for single name in both fields", func(t *testing.T) {
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Return different documents based on field being searched
				if _, ok := criteria.FieldFilters["complainants"]; ok {
					return []*Document{
						{UUID: "doc-1", Title: "Case 1", Code: "case-001"},
					}, nil
				}
				if _, ok := criteria.FieldFilters["respondents"]; ok {
					return []*Document{
						{UUID: "doc-2", Title: "Case 2", Code: "case-002"},
					}, nil
				}
				return []*Document{}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)
		results, err := service.SearchByParticipants(ctx, []string{"John Doe"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have 2 documents (one from complainants, one from respondents)
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		// Verify both searches were called (2 per name: complainants + respondents)
		if mockStore.callCount != 2 {
			t.Errorf("expected 2 search calls, got %d", mockStore.callCount)
		}
	})

	t.Run("deduplicates results by UUID", func(t *testing.T) {
		// Mock that returns the same document for both complainants and respondents
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				return []*Document{
					{UUID: "doc-1", Title: "Same Doc", Code: "case-001"},
				}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)
		results, err := service.SearchByParticipants(ctx, []string{"John Doe"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have only 1 document despite being returned twice
		if len(results) != 1 {
			t.Errorf("expected 1 deduplicated result, got %d", len(results))
		}

		if results[0].UUID != "doc-1" {
			t.Errorf("expected UUID doc-1, got %s", results[0].UUID)
		}
	})

	t.Run("searches multiple names in parallel", func(t *testing.T) {
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Return unique documents based on search term
				for field, value := range criteria.FieldFilters {
					if field == "complainants" {
						if value == "John Doe" {
							return []*Document{{UUID: "doc-1", Title: "John as Complainant"}}, nil
						}
						if value == "Jane Smith" {
							return []*Document{{UUID: "doc-2", Title: "Jane as Complainant"}}, nil
						}
					}
					if field == "respondents" {
						if value == "John Doe" {
							return []*Document{{UUID: "doc-3", Title: "John as Respondent"}}, nil
						}
						if value == "Jane Smith" {
							return []*Document{{UUID: "doc-4", Title: "Jane as Respondent"}}, nil
						}
					}
				}
				return []*Document{}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)
		results, err := service.SearchByParticipants(ctx, []string{"John Doe", "Jane Smith"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have 4 unique documents
		if len(results) != 4 {
			t.Errorf("expected 4 results, got %d", len(results))
		}

		// Verify parallel execution: 2 names × 2 fields = 4 calls
		if mockStore.callCount != 4 {
			t.Errorf("expected 4 parallel search calls, got %d", mockStore.callCount)
		}
	})

	t.Run("handles search errors gracefully", func(t *testing.T) {
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				return nil, ErrDocumentNotFound
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)
		_, err := service.SearchByParticipants(ctx, []string{"John Doe"})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("handles complex name variations", func(t *testing.T) {
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Simulate finding documents for various name formats
				return []*Document{
					{UUID: "doc-1", Title: "Found Doc", Code: "case-001"},
				}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)

		// Test with multiple name variations that might come from generateSearchKeys
		names := []string{
			"John Dabba Doe", // Full name
			"John Doe",       // Short name
			"Doe, John",      // Formal
			"Doe, John D.",   // Formal with middle initial
		}

		results, err := service.SearchByParticipants(ctx, names)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should deduplicate the same document found multiple times
		if len(results) != 1 {
			t.Errorf("expected 1 deduplicated result, got %d", len(results))
		}

		// Verify all name variations were searched (4 names × 2 fields = 8 calls)
		if mockStore.callCount != 8 {
			t.Errorf("expected 8 search calls for name variations, got %d", mockStore.callCount)
		}
	})

	t.Run("concurrent safety test", func(t *testing.T) {
		// Test that parallel execution doesn't cause race conditions
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Simulate some work
				return []*Document{
					{UUID: "doc-1", Title: "Test", Code: "test-001"},
				}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)

		// Run multiple searches concurrently
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := service.SearchByParticipants(ctx, []string{"Test Name"})
				if err != nil {
					t.Errorf("unexpected error in concurrent execution: %v", err)
				}
			}()
		}
		wg.Wait()
	})
}

// TestMiddleNameHandling tests the specific scenario where searching for "John Doe"
// should match documents containing "John Dabba Doe" or "John D. Doe"
func TestMiddleNameHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("John Doe matches John Dabba Doe in documents", func(t *testing.T) {
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Simulate documents with full middle names
				if _, ok := criteria.FieldFilters["complainants"]; ok {
					return []*Document{
						{UUID: "doc-1", Title: "Case with John Dabba Doe", Code: "case-001",
							Fields: map[string]any{"complainants": []any{"John Dabba Doe"}}},
					}, nil
				}
				return []*Document{}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)
		// Search for "John Doe" should find "John Dabba Doe"
		results, err := service.SearchByParticipants(ctx, []string{"John D. Doe"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) == 0 {
			t.Error("expected to find John Dabba Doe when searching for John Doe")
		}
	})

	t.Run("John Doe matches John D. Doe in documents", func(t *testing.T) {
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Simulate documents with middle initial
				if _, ok := criteria.FieldFilters["respondents"]; ok {
					return []*Document{
						{UUID: "doc-2", Title: "Case with John D. Doe", Code: "case-002",
							Fields: map[string]any{"respondents": []any{"John D. Doe"}}},
					}, nil
				}
				return []*Document{}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)
		// Search for "John Doe" should find "John D. Doe"
		results, err := service.SearchByParticipants(ctx, []string{"John Doe"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) == 0 {
			t.Error("expected to find John D. Doe when searching for John Doe")
		}
	})

	t.Run("generateSearchKeys creates John Doe variant for John Dabba Doe", func(t *testing.T) {
		// This test verifies that the aggregator's generateSearchKeys function
		// creates both "John Dabba Doe" and "John Doe" as search keys
		// This ensures documents with either format will be found
		mockStore := &mockSearchStore{
			searchFunc: func(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
				// Count how many different name variations are being searched
				return []*Document{
					{UUID: "doc-1", Title: "Test", Code: "test-001"},
				}, nil
			},
		}

		service := NewDocumentService(mockStore, nil, nil, nil)

		// When aggregator generates keys for "John Dabba Doe", it should include:
		// - "John Dabba Doe" (full)
		// - "John Doe" (short)
		// - "Doe, John" (formal)
		// - "Doe, John D." (formal with initial)
		names := []string{"John Dabba Doe", "John Doe", "Doe, John", "Doe, John D."}

		results, err := service.SearchByParticipants(ctx, names)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should deduplicate to 1 document
		if len(results) != 1 {
			t.Errorf("expected 1 deduplicated result, got %d", len(results))
		}

		// Verify all 4 name formats were searched (4 names × 2 fields = 8 searches)
		if mockStore.callCount != 8 {
			t.Errorf("expected 8 searches for all name variations, got %d", mockStore.callCount)
		}
	})
}
