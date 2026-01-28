package document

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDocumentRepository_UploadFile_ConflictDetection(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload first file
	content1 := bytes.NewReader([]byte("content 1"))
	err = repo.UploadFile(ctx, doc.UUID, "testfile.txt", content1)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Upload second file with same name - should get (Copy) suffix
	content2 := bytes.NewReader([]byte("content 2"))
	err = repo.UploadFile(ctx, doc.UUID, "testfile.txt", content2)
	if err != nil {
		t.Fatalf("UploadFile second with same name failed: %v", err)
	}

	folderPath := filepath.Join(tmpBase, doc.FolderName)

	// Check that both files exist
	if _, err := os.Stat(filepath.Join(folderPath, "testfile.txt")); os.IsNotExist(err) {
		t.Errorf("expected testfile.txt to exist")
	}

	if _, err := os.Stat(filepath.Join(folderPath, "testfile (Copy).txt")); os.IsNotExist(err) {
		t.Errorf("expected testfile (Copy).txt to exist")
	}
}

func TestFileDocumentRepository_UploadFile_MultipleConflicts(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload three files with same name
	for i := 0; i < 3; i++ {
		content := bytes.NewReader([]byte{byte(i)})
		err = repo.UploadFile(ctx, doc.UUID, "file.txt", content)
		if err != nil {
			t.Fatalf("UploadFile iteration %d failed: %v", i, err)
		}
	}

	folderPath := filepath.Join(tmpBase, doc.FolderName)

	// Check that all three exist with proper naming: file.txt, file (Copy).txt, file (Copy 1).txt
	expectedFiles := []string{"file.txt", "file (Copy).txt", "file (Copy 1).txt"}
	for _, expected := range expectedFiles {
		if _, err := os.Stat(filepath.Join(folderPath, expected)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", expected)
		}
	}
}

func TestFileDocumentRepository_UploadFile_MetadataPreservation(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload file
	content := bytes.NewReader([]byte("test content"))
	err = repo.UploadFile(ctx, doc.UUID, "document.pdf", content)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Read files.json and verify metadata
	folderPath := filepath.Join(tmpBase, doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	data, err := os.ReadFile(filesJSONPath)
	if err != nil {
		t.Fatalf("failed to read files.json: %v", err)
	}

	var files []File
	err = json.Unmarshal(data, &files)
	if err != nil {
		t.Fatalf("failed to unmarshal files.json: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].FileName != "document.pdf" {
		t.Errorf("expected filename document.pdf, got %s", files[0].FileName)
	}

	if files[0].Type != ".pdf" {
		t.Errorf("expected type .pdf, got %s", files[0].Type)
	}

	if files[0].Size != 12 {
		t.Errorf("expected size 12, got %d", files[0].Size)
	}
}

func TestFileDocumentRepository_UpdateFileContents(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload initial file
	initialContent := bytes.NewReader([]byte("initial content"))
	err = repo.UploadFile(ctx, doc.UUID, "myfile.txt", initialContent)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Update file contents
	newContent := bytes.NewReader([]byte("updated content with more text"))
	err = repo.UpdateFileContents(ctx, doc.UUID, "myfile.txt", newContent)
	if err != nil {
		t.Fatalf("UpdateFileContents failed: %v", err)
	}

	// Read the file and verify content
	folderPath := filepath.Join(tmpBase, doc.FolderName)
	filePath := filepath.Join(folderPath, "myfile.txt")

	updatedData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	if string(updatedData) != "updated content with more text" {
		t.Errorf("expected 'updated content with more text', got %q", string(updatedData))
	}
}

func TestFileDocumentRepository_UpdateFileContents_PreservesMetadata(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload initial file
	initialContent := bytes.NewReader([]byte("initial"))
	err = repo.UploadFile(ctx, doc.UUID, "myfile.txt", initialContent)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Add custom metadata
	folderPath := filepath.Join(tmpBase, doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []File
	data, _ := os.ReadFile(filesJSONPath)
	json.Unmarshal(data, &files)

	files[0].Description = "My important file"
	files[0].Note = "This is a note"
	files[0].Tags = []string{"important", "archived"}

	filesData, _ := json.Marshal(files)
	os.WriteFile(filesJSONPath, filesData, 0644)

	// Update file contents
	newContent := bytes.NewReader([]byte("new content"))
	err = repo.UpdateFileContents(ctx, doc.UUID, "myfile.txt", newContent)
	if err != nil {
		t.Fatalf("UpdateFileContents failed: %v", err)
	}

	// Verify metadata is preserved
	data, _ = os.ReadFile(filesJSONPath)
	json.Unmarshal(data, &files)

	if files[0].Description != "My important file" {
		t.Errorf("expected description 'My important file', got %q", files[0].Description)
	}

	if files[0].Note != "This is a note" {
		t.Errorf("expected note 'This is a note', got %q", files[0].Note)
	}

	if len(files[0].Tags) != 2 || files[0].Tags[0] != "important" {
		t.Errorf("expected tags preserved, got %v", files[0].Tags)
	}
}

func TestFileDocumentRepository_RenameFile(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload file
	content := bytes.NewReader([]byte("test content"))
	err = repo.UploadFile(ctx, doc.UUID, "oldname.txt", content)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Rename file
	err = repo.RenameFile(ctx, doc.UUID, "oldname.txt", "newname.txt")
	if err != nil {
		t.Fatalf("RenameFile failed: %v", err)
	}

	folderPath := filepath.Join(tmpBase, doc.FolderName)

	// Check old file doesn't exist
	if _, err := os.Stat(filepath.Join(folderPath, "oldname.txt")); err == nil {
		t.Errorf("expected oldname.txt to be deleted")
	}

	// Check new file exists
	if _, err := os.Stat(filepath.Join(folderPath, "newname.txt")); os.IsNotExist(err) {
		t.Errorf("expected newname.txt to exist")
	}

	// Check files.json is updated
	filesJSONPath := filepath.Join(folderPath, "files.json")
	data, _ := os.ReadFile(filesJSONPath)
	var files []File
	json.Unmarshal(data, &files)

	if len(files) != 1 || files[0].FileName != "newname.txt" {
		t.Errorf("expected files.json to reference newname.txt")
	}
}

func TestFileDocumentRepository_RenameFile_ConflictResolution(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload two files
	content1 := bytes.NewReader([]byte("content 1"))
	err = repo.UploadFile(ctx, doc.UUID, "file1.txt", content1)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	content2 := bytes.NewReader([]byte("content 2"))
	err = repo.UploadFile(ctx, doc.UUID, "file2.txt", content2)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Try to rename file1 to file2 (conflict)
	err = repo.RenameFile(ctx, doc.UUID, "file1.txt", "file2.txt")
	if err != nil {
		t.Fatalf("RenameFile with conflict failed: %v", err)
	}

	folderPath := filepath.Join(tmpBase, doc.FolderName)

	// Check that file1 was renamed to "file2 (Copy).txt"
	if _, err := os.Stat(filepath.Join(folderPath, "file2 (Copy).txt")); os.IsNotExist(err) {
		t.Errorf("expected file2 (Copy).txt to exist after conflict resolution")
	}

	// Check metadata is updated correctly
	filesJSONPath := filepath.Join(folderPath, "files.json")
	data, _ := os.ReadFile(filesJSONPath)
	var files []File
	json.Unmarshal(data, &files)

	var found bool
	for _, f := range files {
		if f.FileName == "file2 (Copy).txt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected files.json to contain 'file2 (Copy).txt'")
	}
}

func TestFileDocumentRepository_RenameFile_PreservesMetadata(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Upload file
	content := bytes.NewReader([]byte("test"))
	err = repo.UploadFile(ctx, doc.UUID, "document.txt", content)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Add custom metadata
	folderPath := filepath.Join(tmpBase, doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []File
	data, _ := os.ReadFile(filesJSONPath)
	json.Unmarshal(data, &files)

	files[0].Description = "Important document"
	files[0].Note = "Critical note"
	files[0].Tags = []string{"urgent"}

	filesData, _ := json.Marshal(files)
	os.WriteFile(filesJSONPath, filesData, 0644)

	// Rename file
	err = repo.RenameFile(ctx, doc.UUID, "document.txt", "renamed.txt")
	if err != nil {
		t.Fatalf("RenameFile failed: %v", err)
	}

	// Verify metadata is preserved
	data, _ = os.ReadFile(filesJSONPath)
	json.Unmarshal(data, &files)

	if files[0].Description != "Important document" {
		t.Errorf("expected description preserved, got %q", files[0].Description)
	}

	if files[0].Note != "Critical note" {
		t.Errorf("expected note preserved, got %q", files[0].Note)
	}

	if len(files[0].Tags) != 1 || files[0].Tags[0] != "urgent" {
		t.Errorf("expected tags preserved, got %v", files[0].Tags)
	}
}

func TestFileDocumentRepository_UniqueFileName(t *testing.T) {
	tmpBase := t.TempDir()

	// Create some test files
	testFile1 := filepath.Join(tmpBase, "test.txt")
	os.WriteFile(testFile1, []byte("test"), 0644)

	testFile2 := filepath.Join(tmpBase, "test (Copy).txt")
	os.WriteFile(testFile2, []byte("test"), 0644)

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Test uniqueFileName generates correct names
	tests := []struct {
		baseName string
		expected string
	}{
		{"nonexistent.txt", "nonexistent.txt"},
		{"test.txt", "test (Copy 1).txt"}, // test.txt and test (Copy).txt exist, so should return test (Copy 1).txt
	}

	for _, tt := range tests {
		result, err := repo.uniqueFileName(tmpBase, tt.baseName)
		if err != nil {
			t.Errorf("uniqueFileName failed for %s: %v", tt.baseName, err)
		}
		if result != tt.expected {
			t.Errorf("uniqueFileName(%s) = %s, expected %s", tt.baseName, result, tt.expected)
		}
	}
}

func TestFileDocumentRepository_UpdateFileContents_CreatesIfNotExists(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update a file that doesn't exist yet
	content := bytes.NewReader([]byte("new file content"))
	err = repo.UpdateFileContents(ctx, doc.UUID, "newfile.txt", content)
	if err != nil {
		t.Fatalf("UpdateFileContents for new file failed: %v", err)
	}

	// Check file was created
	folderPath := filepath.Join(tmpBase, doc.FolderName)
	filePath := filepath.Join(folderPath, "newfile.txt")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected newfile.txt to be created")
	}

	// Check metadata was added
	filesJSONPath := filepath.Join(folderPath, "files.json")
	data, _ := os.ReadFile(filesJSONPath)
	var files []File
	json.Unmarshal(data, &files)

	if len(files) != 1 || files[0].FileName != "newfile.txt" {
		t.Errorf("expected metadata for newfile.txt to be added")
	}
}

func TestFileDocumentRepository_PathTraversalProtection(t *testing.T) {
	tmpBase := t.TempDir()

	strategy := NewNamingStrategyCaseDDDD("case")
	repo, err := NewFileDocumentRepository(tmpBase, strategy)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	ctx := context.Background()
	doc := &Document{Title: "Test Document"}
	_, err = repo.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test various path traversal attempts
	maliciousNames := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32\\config",
		"/etc/passwd",
		"C:\\Windows\\System32",
		"./../../secret.txt",
	}

	for _, name := range maliciousNames {
		content := bytes.NewReader([]byte("test"))
		err := repo.UploadFile(ctx, doc.UUID, name, content)
		if err == nil {
			t.Errorf("UploadFile should reject path traversal: %s", name)
		}
	}

	for _, name := range maliciousNames {
		content := bytes.NewReader([]byte("test"))
		err := repo.UpdateFileContents(ctx, doc.UUID, name, content)
		if err == nil {
			t.Errorf("UpdateFileContents should reject path traversal: %s", name)
		}
	}

	for _, name := range maliciousNames {
		err := repo.RenameFile(ctx, doc.UUID, name, "safe.txt")
		if err == nil {
			t.Errorf("RenameFile should reject path traversal: %s", name)
		}
	}
}
