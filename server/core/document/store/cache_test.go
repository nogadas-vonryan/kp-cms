package store

import (
	"context"
	"kpcms/server/core/document"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDocumentRepository_ReloadCache(t *testing.T) {
	tmpBase := t.TempDir()

	docFolder := filepath.Join(tmpBase, "existing_doc")
	os.MkdirAll(docFolder, 0755)

	metaData := `{"uuid": "old-uuid", "code": "OLD-01", "folder_name": "existing_doc", "title": "Old Doc"}`
	os.WriteFile(filepath.Join(docFolder, "meta.json"), []byte(metaData), 0644)

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := New(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	_, err = repo.GetByUUID(context.Background(), "old-uuid")
	if err != nil {
		t.Errorf("repo failed to load existing document into cache: %v", err)
	}
}

func TestFileDocumentRepository_ReloadCacheForFolder_Success(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := New(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	folderName := "case_0001"
	folderPath := filepath.Join(tmpBase, folderName)
	if err := os.MkdirAll(folderPath, 0o755); err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}

	metaData := `{"uuid": "uuid-1", "code": "0001", "folder_name": "case_0001", "title": "Folder Test"}`
	if err := os.WriteFile(filepath.Join(folderPath, "meta.json"), []byte(metaData), 0644); err != nil {
		t.Fatalf("failed to write meta.json: %v", err)
	}

	issues, err := repo.ReloadCacheForFolder(context.Background(), folderName)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}

	// Verify the doc is now in cache by code
	doc, err := repo.GetByCode(context.Background(), "0001")
	if err != nil {
		t.Fatalf("expected document to be retrievable by code, got error: %v", err)
	}
	if doc.Code != "0001" {
		t.Errorf("expected code 0001, got: %s", doc.Code)
	}
}

func TestFileDocumentRepository_ReloadCacheForFolder_MissingMeta(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := New(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	folderName := "case_0002"
	folderPath := filepath.Join(tmpBase, folderName)
	if err := os.MkdirAll(folderPath, 0o755); err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}

	issues, err := repo.ReloadCacheForFolder(context.Background(), folderName)
	if err == nil {
		t.Fatalf("expected error when meta.json is missing")
	}
	if len(issues) == 0 {
		t.Fatalf("expected issues to be returned when meta.json is missing")
	}
	found := false
	for _, it := range issues {
		if it.Type == "MISSING_META" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected MISSING_META issue, got: %v", issues)
	}
}
