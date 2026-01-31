package store

import (
	"context"
	"kpcms/server/core/document"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupSearchTestRepo(t *testing.T) (*FileDocumentRepository, []string) {
	tmpBase := t.TempDir()
	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()

	// Create test documents with various fields
	docs := []struct {
		title  string
		fields map[string]any
	}{
		{
			title: "Mediation Case",
			fields: map[string]any{
				"status":       "mediation",
				"complainants": []any{"John Doe", "Jane Smith"},
				"priority":     "high",
			},
		},
		{
			title: "Pending Review",
			fields: map[string]any{
				"status":   "pending",
				"reviewer": "Admin User",
			},
		},
		{
			title: "Closed Case",
			fields: map[string]any{
				"status":       "closed",
				"complainants": []any{"Bob Johnson"},
				"resolution":   map[string]any{"type": "settled", "amount": 5000},
			},
		},
		{
			title: "Another Mediation",
			fields: map[string]any{
				"status":   "mediation",
				"priority": "low",
			},
		},
		{
			title: "No Status Field",
			fields: map[string]any{
				"priority": "medium",
			},
		},
	}

	var uuids []string
	for _, d := range docs {
		doc := &document.Document{
			Title:  d.title,
			Fields: d.fields,
		}
		created, err := repo.Create(ctx, doc)
		if err != nil {
			t.Fatalf("failed to create test document: %v", err)
		}
		uuids = append(uuids, created.UUID)

		// Sleep to ensure different CreatedAt timestamps
		time.Sleep(10 * time.Millisecond)
	}

	return repo, uuids
}

func TestSearch_ByUUID(t *testing.T) {
	repo, uuids := setupSearchTestRepo(t)
	ctx := context.Background()

	results, err := repo.Search(ctx, document.SearchCriteria{
		UUID: uuids[0],
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if results[0].UUID != uuids[0] {
		t.Errorf("expected UUID %s, got %s", uuids[0], results[0].UUID)
	}
}

func TestSearch_ByCode(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search by exact code
	results, err := repo.Search(ctx, document.SearchCriteria{
		Code: "0001",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("expected at least 1 result, got %d", len(results))
	}

	// Verify prefix matching works
	results, err = repo.Search(ctx, document.SearchCriteria{
		Code: "000", // should match 0001, 0002, etc.
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 5 {
		t.Errorf("expected at least 5 results for prefix '000', got %d", len(results))
	}
}

func TestSearch_ByFolderName(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search by folder name substring
	results, err := repo.Search(ctx, document.SearchCriteria{
		FolderName: "case_0001",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Test case-insensitive matching
	results, err = repo.Search(ctx, document.SearchCriteria{
		FolderName: "CASE",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}
}

func TestSearch_ByFieldKey(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for documents that have "status" field
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldKey: "status",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("expected 4 documents with 'status' field, got %d", len(results))
	}

	// Search for documents that have "complainants" field
	results, err = repo.Search(ctx, document.SearchCriteria{
		FieldKey: "complainants",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 documents with 'complainants' field, got %d", len(results))
	}
}

func TestSearch_ByFieldValue(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for status = "mediation"
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"status": "mediation",
		},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 documents with status='mediation', got %d", len(results))
	}

	for _, doc := range results {
		if doc.Fields["status"] != "mediation" {
			t.Errorf("expected status='mediation', got %v", doc.Fields["status"])
		}
	}
}

func TestSearch_ByFieldValue_CaseInsensitive(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search with different case
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"status": "MEDIATION",
		},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 documents (case-insensitive), got %d", len(results))
	}
}

func TestSearch_InArray(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for "John Doe" in complainants array
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"complainants": "John Doe",
		},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 document with John Doe in complainants, got %d", len(results))
	}

	if results[0].Title != "Mediation Case" {
		t.Errorf("expected 'Mediation Case', got %s", results[0].Title)
	}
}

func TestSearch_InNestedObject(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for "settled" in nested resolution object
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"resolution": "settled",
		},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 document with 'settled' in resolution, got %d", len(results))
	}

	if results[0].Title != "Closed Case" {
		t.Errorf("expected 'Closed Case', got %s", results[0].Title)
	}
}

func TestSearch_MultipleFieldFilters(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for status="mediation" AND priority="high"
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"status":   "mediation",
			"priority": "high",
		},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 document matching both criteria, got %d", len(results))
	}

	if results[0].Title != "Mediation Case" {
		t.Errorf("expected 'Mediation Case', got %s", results[0].Title)
	}
}

func TestSearch_DateRange(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)
	hourFromNow := now.Add(1 * time.Hour)

	// Search for documents created after an hour ago (should get all)
	results, err := repo.Search(ctx, document.SearchCriteria{
		DateFrom: &hourAgo,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 documents created after hour ago, got %d", len(results))
	}

	// Search for documents created before an hour from now (should get all)
	results, err = repo.Search(ctx, document.SearchCriteria{
		DateTo: &hourFromNow,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 documents created before hour from now, got %d", len(results))
	}

	// Search for documents created in the future (should get none)
	futureDate := now.Add(24 * time.Hour)
	results, err = repo.Search(ctx, document.SearchCriteria{
		DateFrom: &futureDate,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 documents in the future, got %d", len(results))
	}
}

func TestSearch_Pagination(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Get first 2 results
	results, err := repo.Search(ctx, document.SearchCriteria{
		Offset: 0,
		Limit:  2,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results with limit=2, got %d", len(results))
	}

	// Get next 2 results
	results, err = repo.Search(ctx, document.SearchCriteria{
		Offset: 2,
		Limit:  2,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for second page, got %d", len(results))
	}

	// Get results beyond available documents
	results, err = repo.Search(ctx, document.SearchCriteria{
		Offset: 10,
		Limit:  2,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results beyond available docs, got %d", len(results))
	}
}

func TestSearch_CombinedCriteria(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Combine code prefix, field existence, and field value
	results, err := repo.Search(ctx, document.SearchCriteria{
		Code:     "000",
		FieldKey: "status",
		FieldFilters: map[string]any{
			"status": "mediation",
		},
		Limit: 10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results matching all criteria, got %d", len(results))
	}

	for _, doc := range results {
		if doc.Fields["status"] != "mediation" {
			t.Errorf("expected status='mediation', got %v", doc.Fields["status"])
		}
	}
}

func TestSearch_NoResults(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for non-existent UUID
	results, err := repo.Search(ctx, document.SearchCriteria{
		UUID: "non-existent-uuid",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for non-existent UUID, got %d", len(results))
	}

	// Search for non-existent field value
	results, err = repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"status": "non-existent-status",
		},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for non-existent status, got %d", len(results))
	}
}

func TestSearch_EmptyCriteria(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Empty criteria should return all documents
	results, err := repo.Search(ctx, document.SearchCriteria{})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results for empty criteria, got %d", len(results))
	}
}

func TestSearch_WithActualFolders(t *testing.T) {
	tmpBase := t.TempDir()
	strategy := document.NewNamingStrategyCaseDDDD("case")

	// Create folders with actual structure
	doc1Folder := filepath.Join(tmpBase, "case_0001_mediation")
	os.MkdirAll(doc1Folder, 0755)
	metaData1 := `{"uuid": "uuid-1", "code": "0001", "folder_name": "case_0001_mediation", "title": "First Case", "fields": {"status": "mediation"}, "created_at": "2026-01-25T10:00:00Z", "updated_at": "2026-01-25T10:00:00Z"}`
	os.WriteFile(filepath.Join(doc1Folder, "meta.json"), []byte(metaData1), 0644)
	os.WriteFile(filepath.Join(doc1Folder, "files.json"), []byte("[]"), 0644)

	doc2Folder := filepath.Join(tmpBase, "case_0002_pending")
	os.MkdirAll(doc2Folder, 0755)
	metaData2 := `{"uuid": "uuid-2", "code": "0002", "folder_name": "case_0002_pending", "title": "Second Case", "fields": {"status": "pending"}, "created_at": "2026-01-25T11:00:00Z", "updated_at": "2026-01-25T11:00:00Z"}`
	os.WriteFile(filepath.Join(doc2Folder, "meta.json"), []byte(metaData2), 0644)
	os.WriteFile(filepath.Join(doc2Folder, "files.json"), []byte("[]"), 0644)

	repo, err := NewFileDocumentRepository(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()

	// Search by folder name
	results, err := repo.Search(ctx, document.SearchCriteria{
		FolderName: "mediation",
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for folder name 'mediation', got %d", len(results))
	}

	if results[0].Title != "First Case" {
		t.Errorf("expected 'First Case', got %s", results[0].Title)
	}
}

func TestSearch_SortByCode(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Sort by code ascending (default)
	results, err := repo.Search(ctx, document.SearchCriteria{
		SortBy: "code",
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// Verify ascending order
	for i := 0; i < len(results)-1; i++ {
		if results[i].Code > results[i+1].Code {
			t.Errorf("expected code %s <= %s (ascending order)", results[i].Code, results[i+1].Code)
		}
	}

	// Sort by code descending
	results, err = repo.Search(ctx, document.SearchCriteria{
		SortBy:   "code",
		SortDesc: true,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify descending order
	for i := 0; i < len(results)-1; i++ {
		if results[i].Code < results[i+1].Code {
			t.Errorf("expected code %s >= %s (descending order)", results[i].Code, results[i+1].Code)
		}
	}
}

func TestSearch_SortByTitle(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Sort by title ascending
	results, err := repo.Search(ctx, document.SearchCriteria{
		SortBy: "title",
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// Verify ascending alphabetical order
	for i := 0; i < len(results)-1; i++ {
		titleI := strings.ToLower(results[i].Title)
		titleJ := strings.ToLower(results[i+1].Title)
		if titleI > titleJ {
			t.Errorf("expected title %s <= %s (ascending order)", titleI, titleJ)
		}
	}

	// Sort by title descending
	results, err = repo.Search(ctx, document.SearchCriteria{
		SortBy:   "title",
		SortDesc: true,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify descending alphabetical order
	for i := 0; i < len(results)-1; i++ {
		titleI := strings.ToLower(results[i].Title)
		titleJ := strings.ToLower(results[i+1].Title)
		if titleI < titleJ {
			t.Errorf("expected title %s >= %s (descending order)", titleI, titleJ)
		}
	}
}

func TestSearch_SortByCreatedAt(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Sort by created_at ascending
	results, err := repo.Search(ctx, document.SearchCriteria{
		SortBy: "created_at",
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// Verify ascending chronological order (newest should be last)
	for i := 0; i < len(results)-1; i++ {
		if results[i].CreatedAt.After(results[i+1].CreatedAt) {
			t.Errorf("expected created_at in ascending order at index %d and %d", i, i+1)
		}
	}

	// Sort by created_at descending (newest first)
	results, err = repo.Search(ctx, document.SearchCriteria{
		SortBy:   "created_at",
		SortDesc: true,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify descending chronological order (newest should be first)
	for i := 0; i < len(results)-1; i++ {
		if results[i].CreatedAt.Before(results[i+1].CreatedAt) {
			t.Errorf("expected created_at in descending order at index %d and %d", i, i+1)
		}
	}
}

func TestSearch_SortByFolderName(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Sort by folder_name ascending
	results, err := repo.Search(ctx, document.SearchCriteria{
		SortBy: "folder_name",
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// Verify ascending alphabetical order
	for i := 0; i < len(results)-1; i++ {
		folderI := strings.ToLower(results[i].FolderName)
		folderJ := strings.ToLower(results[i+1].FolderName)
		if folderI > folderJ {
			t.Errorf("expected folder_name %s <= %s (ascending order)", folderI, folderJ)
		}
	}

	// Sort by folder_name descending
	results, err = repo.Search(ctx, document.SearchCriteria{
		SortBy:   "folder_name",
		SortDesc: true,
		Limit:    10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify descending alphabetical order
	for i := 0; i < len(results)-1; i++ {
		folderI := strings.ToLower(results[i].FolderName)
		folderJ := strings.ToLower(results[i+1].FolderName)
		if folderI < folderJ {
			t.Errorf("expected folder_name %s >= %s (descending order)", folderI, folderJ)
		}
	}
}

func TestSearch_SortWithFilter(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// Search for mediation documents and sort by title ascending
	results, err := repo.Search(ctx, document.SearchCriteria{
		FieldFilters: map[string]any{
			"status": "mediation",
		},
		SortBy: "title",
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 mediation results, got %d", len(results))
	}

	// Should be sorted alphabetically
	if results[0].Title > results[1].Title {
		t.Errorf("expected title %s <= %s (ascending order)", results[0].Title, results[1].Title)
	}
}

func TestSearch_DefaultSort(t *testing.T) {
	repo, _ := setupSearchTestRepo(t)
	ctx := context.Background()

	// When SortBy is empty, should default to "code"
	results, err := repo.Search(ctx, document.SearchCriteria{
		Limit: 10,
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// Verify default ascending code order
	for i := 0; i < len(results)-1; i++ {
		if results[i].Code > results[i+1].Code {
			t.Errorf("expected default sort by code in ascending order")
		}
	}
}
