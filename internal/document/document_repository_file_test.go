package document

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDocumentRepository_Create(t *testing.T) {
	tmpBase := t.TempDir()

	repo, err := NewFileDocumentRepository(tmpBase)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{
		UUID:       "d5f78bfc-ccc7-435b-94f2-85027b20fb8c",
		Code:       "DOC-001",
		FolderName: "folder_001",
		Title:      "Test Document",
	}

	err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	metaPath := filepath.Join(tmpBase, "folder_001", "meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("expected meta.json to exist at %s", metaPath)
	}

	retrieved, err := repo.GetByUUID(ctx, doc.UUID)
	if err != nil {
		t.Fatalf("GetByUUID failed: %v", err)
	}
	if retrieved.Title != doc.Title {
		t.Errorf("expected title %s, got %s", doc.Title, retrieved.Title)
	}
}

func TestFileDocumentRepository_ReloadCache(t *testing.T) {
	tmpBase := t.TempDir()

	docFolder := filepath.Join(tmpBase, "existing_doc")
	os.MkdirAll(docFolder, 0755)

	metaData := `{"uuid": "old-uuid", "code": "OLD-01", "folder_name": "existing_doc", "title": "Old Doc"}`
	os.WriteFile(filepath.Join(docFolder, "meta.json"), []byte(metaData), 0644)

	repo, err := NewFileDocumentRepository(tmpBase)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	_, err = repo.GetByUUID(context.Background(), "old-uuid")
	if err != nil {
		t.Errorf("repo failed to load existing document into cache: %v", err)
	}
}
