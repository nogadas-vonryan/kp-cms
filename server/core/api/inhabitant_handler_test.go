package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kpcms/server/core/auth"
	"kpcms/server/core/database"
	"kpcms/server/core/document"
	"kpcms/server/core/document/store"
	"kpcms/server/core/inhabitant"
	inhabitantstore "kpcms/server/core/inhabitant/store"

	"github.com/go-chi/chi/v5"
)

func setupTestServerWithInhabitantStore(t *testing.T) (*Server, *inhabitantstore.InhabitantFileStore, *store.Store) {
	tempDir := t.TempDir()

	strategy := document.NewNamingStrategyCaseDDDD("case")
	docRepo, err := store.New(tempDir, "", strategy)
	if err != nil {
		t.Fatalf("failed to create doc repo: %v", err)
	}

	// Create file-based inhabitant store
	inhabPath := t.TempDir()
	inhabStrategy := document.NewNamingStrategyPrefixDDDYY("inhabitant")
	inhabStore, err := inhabitantstore.New(inhabPath, inhabStrategy)
	if err != nil {
		t.Fatalf("failed to create inhabitant store: %v", err)
	}

	// Create auth database
	authDBPath := t.TempDir() + "/auth.db"
	db, err := database.New(t.TempDir()+"/app.db", authDBPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	authRepo := auth.NewSQLRepository(db.AuthDB)
	authService := auth.NewService(authRepo)
	inhabitantService := inhabitant.NewService(inhabStore)
	documentService := document.NewDocumentService(docRepo, docRepo, nil, nil)

	server, err := NewServer("0.0.0.0", "8080", "admin", "password",
		documentService, authService, inhabitantService, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	return server, inhabStore, docRepo
}

func TestHandleListInhabitants_Success(t *testing.T) {
	server, inhabStore, _ := setupTestServerWithInhabitantStore(t)

	// Create some inhabitants using file-based store
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
			FirstName: "User",
			LastName:  "Name" + string(rune('0'+byte(i))),
		})
		if err != nil {
			t.Fatalf("failed to create inhabitant: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/api/inhabitants", nil)
	w := httptest.NewRecorder()

	server.handleListInhabitants().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var inhabitants []inhabitantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &inhabitants); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(inhabitants) < 3 {
		t.Errorf("expected at least 3 inhabitants, got %d", len(inhabitants))
	}
}

func TestHandleGetInhabitant_Success(t *testing.T) {
	server, inhabStore, _ := setupTestServerWithInhabitantStore(t)

	ctx := context.Background()
	created, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "John",
		LastName:  "Doe",
		ContactNo: "555-1234",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Get("/inhabitants/{id}", server.handleGetInhabitant())

	req := httptest.NewRequest("GET", "/inhabitants/"+created.UUID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp inhabitantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.FirstName != "John" {
		t.Errorf("expected first name John, got %q", resp.FirstName)
	}
}

func TestHandleCreateInhabitant_Success(t *testing.T) {
	server, _, _ := setupTestServerWithInhabitantStore(t)

	reqBody := createInhabitantRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		ContactNo: "555-5678",
		Birthdate: "1990-01-15",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/inhabitants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateInhabitant().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp inhabitantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.FirstName != "Jane" {
		t.Errorf("expected first name Jane, got %q", resp.FirstName)
	}
}

func TestHandleCreateInhabitant_MissingRequired(t *testing.T) {
	server, _, _ := setupTestServerWithInhabitantStore(t)

	reqBody := createInhabitantRequest{
		FirstName: "John",
		// Missing LastName
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/inhabitants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateInhabitant().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUpdateInhabitant_Success(t *testing.T) {
	server, inhabStore, _ := setupTestServerWithInhabitantStore(t)

	ctx := context.Background()
	created, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "John",
		LastName:  "Doe",
		ContactNo: "555-1111",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	reqBody := createInhabitantRequest{
		FirstName: "Jonathan",
		LastName:  "Doe",
		ContactNo: "555-2222",
	}

	body, _ := json.Marshal(reqBody)

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Put("/inhabitants/{id}", server.handleUpdateInhabitant())

	req := httptest.NewRequest("PUT", "/inhabitants/"+created.UUID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the update
	updated, _ := inhabStore.GetByUUID(ctx, created.UUID)
	if updated.FirstName != "Jonathan" {
		t.Errorf("expected first name Jonathan, got %q", updated.FirstName)
	}
}

func TestHandleDeleteInhabitant_Success(t *testing.T) {
	server, inhabStore, _ := setupTestServerWithInhabitantStore(t)

	ctx := context.Background()
	created, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "Alice",
		LastName:  "Adams",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Delete("/inhabitants/{id}", server.handleDeleteInhabitant())

	req := httptest.NewRequest("DELETE", "/inhabitants/"+created.UUID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify the deletion
	_, err = inhabStore.GetByUUID(ctx, created.UUID)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestHandleDeleteInhabitant_NotFound(t *testing.T) {
	server, _, _ := setupTestServerWithInhabitantStore(t)

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Delete("/inhabitants/{id}", server.handleDeleteInhabitant())

	req := httptest.NewRequest("DELETE", "/inhabitants/99999999-9999-9999-9999-999999999999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// === Tests for handleGetInhabitantDocuments ===

func TestHandleGetInhabitantDocuments_WithLinkedDocuments(t *testing.T) {
	server, inhabStore, docRepo := setupTestServerWithInhabitantStore(t)

	ctx := context.Background()

	// Create an inhabitant
	created, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "John",
		LastName:  "Doe",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	// Create documents linked to this inhabitant
	inhabitantCode := created.FolderName

	doc1, err := docRepo.Create(ctx, &document.Document{
		Title: "Case 001 - John as Complainant",
		ParticipantIDs: &document.ParticipantLinks{
			Complainants: []string{inhabitantCode},
			Respondents:  []string{"inhabitant-other-26"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create document 1: %v", err)
	}

	doc2, err := docRepo.Create(ctx, &document.Document{
		Title: "Case 002 - John as Respondent",
		ParticipantIDs: &document.ParticipantLinks{
			Complainants: []string{"inhabitant-other2-26"},
			Respondents:  []string{inhabitantCode},
		},
	})
	if err != nil {
		t.Fatalf("failed to create document 2: %v", err)
	}

	// Reload cache to populate the index
	_, err = docRepo.ReloadCache(ctx)
	if err != nil {
		t.Fatalf("failed to reload cache: %v", err)
	}

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Get("/inhabitants/{id}/documents", server.handleGetInhabitantDocuments())

	req := httptest.NewRequest("GET", "/inhabitants/"+created.UUID+"/documents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var docs []document.Document
	if err := json.Unmarshal(w.Body.Bytes(), &docs); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}

	// Verify the documents contain the correct data
	docUUIDs := make(map[string]bool)
	for _, doc := range docs {
		docUUIDs[doc.UUID] = true
		if doc.Title == "" {
			t.Error("document title should not be empty")
		}
	}

	if !docUUIDs[doc1.UUID] {
		t.Error("expected document 1 to be in response")
	}
	if !docUUIDs[doc2.UUID] {
		t.Error("expected document 2 to be in response")
	}
}

func TestHandleGetInhabitantDocuments_NotFound(t *testing.T) {
	server, _, _ := setupTestServerWithInhabitantStore(t)

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Get("/inhabitants/{id}/documents", server.handleGetInhabitantDocuments())

	req := httptest.NewRequest("GET", "/inhabitants/99999999-9999-9999-9999-999999999999/documents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandleGetInhabitantDocuments_NoLinks(t *testing.T) {
	server, inhabStore, docRepo := setupTestServerWithInhabitantStore(t)

	ctx := context.Background()

	// Create an inhabitant
	created, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "Jane",
		LastName:  "Smith",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	// Create a document NOT linked to this inhabitant
	_, err = docRepo.Create(ctx, &document.Document{
		Title: "Case with different participants",
		ParticipantIDs: &document.ParticipantLinks{
			Complainants: []string{"inhabitant-other-26"},
			Respondents:  []string{"inhabitant-another-26"},
		},
	})
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Reload cache to populate the index
	_, err = docRepo.ReloadCache(ctx)
	if err != nil {
		t.Fatalf("failed to reload cache: %v", err)
	}

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Get("/inhabitants/{id}/documents", server.handleGetInhabitantDocuments())

	req := httptest.NewRequest("GET", "/inhabitants/"+created.UUID+"/documents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var docs []document.Document
	if err := json.Unmarshal(w.Body.Bytes(), &docs); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected 0 documents (no links), got %d", len(docs))
	}
}

func TestHandleGetInhabitantDocuments_ByCode(t *testing.T) {
	server, inhabStore, docRepo := setupTestServerWithInhabitantStore(t)

	ctx := context.Background()

	// Create an inhabitant
	created, err := inhabStore.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "Bob",
		LastName:  "Johnson",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	// Create a document linked to this inhabitant
	inhabitantCode := created.FolderName
	_, err = docRepo.Create(ctx, &document.Document{
		Title: "Case linked to Bob",
		ParticipantIDs: &document.ParticipantLinks{
			Complainants: []string{inhabitantCode},
		},
	})
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Reload cache to populate the index
	_, err = docRepo.ReloadCache(ctx)
	if err != nil {
		t.Fatalf("failed to reload cache: %v", err)
	}

	// Use chi router with code instead of UUID
	router := chi.NewRouter()
	router.Get("/inhabitants/{id}/documents", server.handleGetInhabitantDocuments())

	req := httptest.NewRequest("GET", "/inhabitants/"+created.Code+"/documents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var docs []document.Document
	if err := json.Unmarshal(w.Body.Bytes(), &docs); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}
}
