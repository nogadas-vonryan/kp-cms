package store

import (
	"context"
	"encoding/json"
	"kpcms/server/core/document"
	"os"
	"path/filepath"
	"testing"
)

// TestManuallyAddedFile_AutomaticRegistrationAndUpdate verifies that when a file
// is added manually via file explorer (not through the API), it gets automatically
// registered in files.json when the document is accessed, and metadata updates
// work correctly afterwards.
func TestManuallyAddedFile_AutomaticRegistrationAndUpdate(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()

	// Create a document
	doc := &document.Document{Title: "Test Document"}
	created, err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Manually add a file to the directory (simulating file explorer addition)
	folderPath := repo.getDocumentPath(created.FolderName)
	manualFilePath := filepath.Join(folderPath, "manual_file.txt")
	manualContent := []byte("This file was added manually via file explorer")
	err = os.WriteFile(manualFilePath, manualContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write manual file: %v", err)
	}

	// Verify the file is NOT in files.json yet
	filesJSONPath := filepath.Join(folderPath, "files.json")
	initialData, err := os.ReadFile(filesJSONPath)
	if err != nil {
		t.Fatalf("Failed to read files.json: %v", err)
	}
	var initialFiles []document.File
	if err := json.Unmarshal(initialData, &initialFiles); err != nil {
		t.Fatalf("Failed to parse files.json: %v", err)
	}

	// Confirm the manual file is not in the metadata yet
	for _, f := range initialFiles {
		if f.FileName == "manual_file.txt" {
			t.Fatalf("Manual file should not be in files.json yet, but it is")
		}
	}
	t.Log("✓ Confirmed: manual file not in files.json initially")

	// Now access the document via GetByUUID (which calls readFiles)
	// This should trigger automatic registration of the manual file
	retrievedDoc, err := repo.GetByUUID(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetByUUID failed: %v", err)
	}

	// Verify the manual file is now in the returned files list
	foundInMemory := false
	for _, f := range retrievedDoc.Files {
		if f.FileName == "manual_file.txt" {
			foundInMemory = true
			if f.Size != int64(len(manualContent)) {
				t.Errorf("Expected file size %d, got %d", len(manualContent), f.Size)
			}
			break
		}
	}
	if !foundInMemory {
		t.Fatalf("Manual file not found in GetByUUID response")
	}
	t.Log("✓ Manual file appears in GetByUUID response")

	// Verify the file was automatically written to files.json
	updatedData, err := os.ReadFile(filesJSONPath)
	if err != nil {
		t.Fatalf("Failed to read files.json after GetByUUID: %v", err)
	}
	var updatedFiles []document.File
	if err := json.Unmarshal(updatedData, &updatedFiles); err != nil {
		t.Fatalf("Failed to parse updated files.json: %v", err)
	}

	foundInJSON := false
	for _, f := range updatedFiles {
		if f.FileName == "manual_file.txt" {
			foundInJSON = true
			break
		}
	}
	if !foundInJSON {
		t.Fatalf("Manual file was not automatically synced to files.json")
	}
	t.Log("✓ Manual file was automatically synced to files.json")

	// Now test that updating metadata works (this was the original bug)
	description := "This is a test description"
	note := "Important note"
	tags := []string{"manual", "test", "important"}

	updates := document.FileMetadataUpdate{
		Description: &description,
		Note:        &note,
		Tags:        &tags,
	}

	err = repo.UpdateFileMetadata(ctx, created.UUID, "manual_file.txt", updates)
	if err != nil {
		t.Fatalf("UpdateFileMetadata failed for manually added file: %v", err)
	}
	t.Log("✓ UpdateFileMetadata succeeded for manually added file")

	// Verify the metadata was actually updated
	finalDoc, err := repo.GetByUUID(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetByUUID failed after update: %v", err)
	}

	var updatedFile *document.File
	for i, f := range finalDoc.Files {
		if f.FileName == "manual_file.txt" {
			updatedFile = &finalDoc.Files[i]
			break
		}
	}

	if updatedFile == nil {
		t.Fatalf("Could not find manual file after metadata update")
	}

	if updatedFile.Description != description {
		t.Errorf("Description not updated. Expected: %s, Got: %s", description, updatedFile.Description)
	}
	if updatedFile.Note != note {
		t.Errorf("Note not updated. Expected: %s, Got: %s", note, updatedFile.Note)
	}
	if len(updatedFile.Tags) != len(tags) {
		t.Errorf("Tags count mismatch. Expected: %d, Got: %d", len(tags), len(updatedFile.Tags))
	} else {
		for i, tag := range tags {
			if updatedFile.Tags[i] != tag {
				t.Errorf("Tag[%d] mismatch. Expected: %s, Got: %s", i, tag, updatedFile.Tags[i])
			}
		}
	}

	t.Log("✓ All metadata updates were successfully applied")
	t.Log("✅ TEST PASSED: Manually added files are automatically registered and can be updated")
}

// TestManuallyAddedFile_MultipleFiles verifies that multiple manually added files
// are all correctly registered when the document is accessed
func TestManuallyAddedFile_MultipleFiles(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, "", strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()

	// Create a document
	doc := &document.Document{Title: "Test Document"}
	created, err := repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Manually add multiple files
	folderPath := repo.getDocumentPath(created.FolderName)
	manualFiles := []string{"file1.txt", "file2.pdf", "file3.doc"}

	for _, filename := range manualFiles {
		filePath := filepath.Join(folderPath, filename)
		err = os.WriteFile(filePath, []byte("content of "+filename), 0644)
		if err != nil {
			t.Fatalf("Failed to write manual file %s: %v", filename, err)
		}
	}

	// Access the document to trigger auto-registration
	retrievedDoc, err := repo.GetByUUID(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetByUUID failed: %v", err)
	}

	// Verify all manual files are present
	foundCount := 0
	for _, filename := range manualFiles {
		found := false
		for _, f := range retrievedDoc.Files {
			if f.FileName == filename {
				found = true
				foundCount++
				break
			}
		}
		if !found {
			t.Errorf("Manual file %s not found in GetByUUID response", filename)
		}
	}

	if foundCount != len(manualFiles) {
		t.Fatalf("Expected %d manual files, found %d", len(manualFiles), foundCount)
	}

	// Update metadata on one of them to ensure it works
	desc := "Updated description"
	err = repo.UpdateFileMetadata(ctx, created.UUID, "file2.pdf", document.FileMetadataUpdate{
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("UpdateFileMetadata failed for manually added file: %v", err)
	}

	t.Log("✅ All manually added files were registered and can be updated")
}
