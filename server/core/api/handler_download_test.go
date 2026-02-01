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

	"kpcms/server/core/auth"
	"kpcms/server/core/database"
	"kpcms/server/core/document"
	"kpcms/server/core/document/store"
)

func setupTestServer(t *testing.T) (*Server, string) {
	tempDir := t.TempDir()

	strategy := document.NewNamingStrategyCaseDDDD("case")
	repo, err := store.New(tempDir, "", strategy)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	memName := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := database.New(fmt.Sprintf("file:%s?mode=memory&cache=shared", memName))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	server, err := NewServer("0.0.0.0", "8080", "admin", "password", document.NewDocumentService(repo, repo, nil, nil), db)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	return server, tempDir
}

func TestHandleDownloadFile_WithSpaces(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	// Create a document
	doc := document.Document{Title: "Test Doc with Spaces"}
	createdDoc, err := server.documentService.Create(ctx, doc)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Upload a file with spaces
	fileName := "test file with spaces.txt"
	fileContent := "hello world from test"
	err = server.documentService.UploadFile(ctx, createdDoc.UUID, fileName, strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	// Test download via HTTP with URL-encoded filename
	urlEncodedFileName := strings.ReplaceAll(fileName, " ", "%20")
	reqURL := fmt.Sprintf("/api/documents/%s/files/%s", createdDoc.UUID, urlEncodedFileName)

	req := httptest.NewRequest("GET", reqURL, nil)
	addAdminSessionCookie(t, server, req)
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; response: %s", w.Code, w.Body.String())
	}

	responseBody, _ := io.ReadAll(w.Body)
	if string(responseBody) != fileContent {
		t.Errorf("expected content %q, got %q", fileContent, string(responseBody))
	}
}

func TestHandleDownloadFile_WithMultipleSpaces(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	doc := document.Document{Title: "Test"}
	createdDoc, err := server.documentService.Create(ctx, doc)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	fileName := "my  document   file.pdf"
	fileContent := "PDF content here"
	err = server.documentService.UploadFile(ctx, createdDoc.UUID, fileName, strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	urlEncodedFileName := strings.ReplaceAll(fileName, " ", "%20")
	reqURL := fmt.Sprintf("/api/documents/%s/files/%s", createdDoc.UUID, urlEncodedFileName)

	req := httptest.NewRequest("GET", reqURL, nil)
	addAdminSessionCookie(t, server, req)
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	responseBody, _ := io.ReadAll(w.Body)
	if string(responseBody) != fileContent {
		t.Errorf("expected %q, got %q", fileContent, string(responseBody))
	}
}

func TestHandleDownloadFile_WithSpecialCharacters(t *testing.T) {
	repo, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	doc := document.Document{Title: "Test"}
	createdDoc, err := repo.documentService.Create(ctx, doc)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Test the exact filename from the user's issue
	fileName := "One Day in the Life of a Rice Farmer [s_kLkOOV3CE].webm"
	fileContent := "Video content here"
	err = repo.documentService.UploadFile(ctx, createdDoc.UUID, fileName, strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	// Test Download
	reader, err := repo.documentService.DownloadFile(ctx, createdDoc.UUID, fileName)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
	defer reader.Close()

	gotContent, _ := io.ReadAll(reader)
	if string(gotContent) != fileContent {
		t.Errorf("expected %q, got %q", fileContent, string(gotContent))
	}
}

// addAdminSessionCookie attaches a valid admin session cookie to the request so it passes auth middleware
func addAdminSessionCookie(t *testing.T, server *Server, req *http.Request) {
	token, err := server.db.CreateSession(context.Background(), auth.Identity{ID: "admin", Role: auth.RoleAdmin}, server.sessionTTL)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
}

// Ensure the HTTP route and middleware path correctly serve filenames with brackets/spaces
func TestHandleDownloadFile_WithBracketsAndSpaces_HTTP(t *testing.T) {
	server, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	doc := document.Document{Title: "HTTP Doc"}
	createdDoc, err := server.documentService.Create(ctx, doc)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	fileName := "One Day in the Life of a Rice Farmer [s_kLkOOV3CE].webm"
	fileContent := "Video content here"
	if err := server.documentService.UploadFile(ctx, createdDoc.UUID, fileName, strings.NewReader(fileContent)); err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	// Create an admin session and attach cookie to bypass auth middleware
	token, err := server.db.CreateSession(context.Background(), auth.Identity{ID: "admin", Role: auth.RoleAdmin}, server.sessionTTL)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	cookieHeader := fmt.Sprintf("%s=%s", auth.SessionCookieName, token)

	// Use the exact URL shape reported by the user (spaces encoded, brackets unencoded)
	encodedPath := "/api/documents/" + createdDoc.UUID + "/files/One%20Day%20in%20the%20Life%20of%20a%20Rice%20Farmer%20[s_kLkOOV3CE].webm"
	req := httptest.NewRequest(http.MethodGet, encodedPath, nil)
	req.Header.Set("Cookie", cookieHeader)

	w := httptest.NewRecorder()
	server.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != fileContent {
		t.Errorf("expected %q, got %q", fileContent, string(body))
	}
}
