package document

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) (*FileDocumentRepository, string) {
	tempDir, err := os.MkdirTemp("", "doc_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	repo, err := NewFileDocumentRepository(tempDir, nil)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	return repo, tempDir
}

func TestFileDocumentRepository_DownloadFile(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	doc, _ := repo.Create(ctx, &Document{Title: "Test Doc"})

	fileName := "test.txt"
	content := "hello world"
	repo.UploadFile(ctx, doc.UUID, fileName, strings.NewReader(content))

	// Test Download
	reader, err := repo.DownloadFile(ctx, doc.UUID, fileName)
	if err != nil {
		t.Errorf("DownloadFile failed: %v", err)
	}
	defer reader.Close()

	gotContent, _ := io.ReadAll(reader)
	if string(gotContent) != content {
		t.Errorf("expected %q, got %q", content, string(gotContent))
	}
}

func TestFileDocumentRepository_UpdateFileMetadata(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	doc, _ := repo.Create(ctx, &Document{Title: "Meta Test"})

	fileName := "data.csv"
	repo.UploadFile(ctx, doc.UUID, fileName, strings.NewReader("1,2,3"))

	// Update Metadata
	newDesc := "New Description"
	newNote := "Important note"
	err := repo.UpdateFileMetadata(ctx, doc.UUID, fileName, newDesc, newNote)
	if err != nil {
		t.Fatalf("UpdateFileMetadata failed: %v", err)
	}

	// Verify update by fetching the document again
	updatedDoc, _ := repo.GetByUUID(ctx, doc.UUID)
	found := false
	for _, f := range updatedDoc.Files {
		if f.FileName == fileName {
			if f.Description != newDesc || f.Note != newNote {
				t.Errorf("metadata mismatch: got desc=%q, note=%q", f.Description, f.Note)
			}
			found = true
		}
	}

	if !found {
		t.Error("file entry not found in document after metadata update")
	}
}

func TestFileDocumentRepository_DeleteFile(t *testing.T) {
	repo, tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	doc, _ := repo.Create(ctx, &Document{Title: "Delete Test"})

	fileName := "trash.txt"
	repo.UploadFile(ctx, doc.UUID, fileName, strings.NewReader("delete me"))

	// Delete
	err := repo.DeleteFile(ctx, doc.UUID, fileName)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	// 1. Verify file is gone from Document view
	updatedDoc, _ := repo.GetByUUID(ctx, doc.UUID)
	if len(updatedDoc.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(updatedDoc.Files))
	}

	// 2. Verify physical file is gone
	folderPath, _ := repo.GetDocumentFolderPath(ctx, doc.UUID)
	filePath := filepath.Join(folderPath, fileName)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("physical file still exists after deletion")
	}
}
