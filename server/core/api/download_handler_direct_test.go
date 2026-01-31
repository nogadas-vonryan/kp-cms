package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"kpcms/server/core/document"

	"github.com/go-chi/chi/v5"
)

func TestDownloadFile_WithSpaces_Direct(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := document.NewFileDocumentRepository(tempDir, "", strategy)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	docService := document.NewDocumentService(repo, repo, nil, nil)
	ctx := context.Background()

	// Create a document
	doc := document.Document{Title: "Test Doc"}
	createdDoc, err := docService.Create(ctx, doc)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Upload a file with spaces
	fileName := "file with spaces.txt"
	fileContent := "test content"
	err = docService.UploadFile(ctx, createdDoc.UUID, fileName, strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("failed to upload: %v", err)
	}

	// Create a test server with minimal routing
	server := &Server{
		documentService: docService,
	}

	// Create a chi router with just the download handler
	router := chi.NewRouter()
	router.Get("/{uuid}/files/{fileName}", server.handleDownloadFile())

	// Test with URL-encoded filename
	urlEncodedFileName := strings.ReplaceAll(fileName, " ", "%20")
	testURL := fmt.Sprintf("/%s/files/%s", createdDoc.UUID, urlEncodedFileName)

	req := httptest.NewRequest("GET", testURL, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != fileContent {
		t.Errorf("expected %q, got %q", fileContent, string(body))
	}
}

func TestDownloadFile_WithMultipleSpaces_Direct(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := document.NewFileDocumentRepository(tempDir, "", strategy)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	docService := document.NewDocumentService(repo, repo, nil, nil)
	ctx := context.Background()

	doc := document.Document{Title: "Test"}
	createdDoc, err := docService.Create(ctx, doc)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	fileName := "my  document   file.pdf"
	fileContent := "PDF content"
	err = docService.UploadFile(ctx, createdDoc.UUID, fileName, strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("failed to upload: %v", err)
	}

	server := &Server{
		documentService: docService,
	}

	router := chi.NewRouter()
	router.Get("/{uuid}/files/{fileName}", server.handleDownloadFile())

	urlEncodedFileName := strings.ReplaceAll(fileName, " ", "%20")
	testURL := fmt.Sprintf("/%s/files/%s", createdDoc.UUID, urlEncodedFileName)

	req := httptest.NewRequest("GET", testURL, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != fileContent {
		t.Errorf("expected %q, got %q", fileContent, string(body))
	}
}
