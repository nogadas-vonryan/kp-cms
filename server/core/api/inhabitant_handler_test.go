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

func setupTestServerWithInhabitantStore(t *testing.T) (*Server, *inhabitantstore.InhabitantFileStore) {
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

	return server, inhabStore
}

func TestHandleListInhabitants_Success(t *testing.T) {
	server, inhabStore := setupTestServerWithInhabitantStore(t)

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
	server, inhabStore := setupTestServerWithInhabitantStore(t)

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
	server, _ := setupTestServerWithInhabitantStore(t)

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
	server, _ := setupTestServerWithInhabitantStore(t)

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
	server, inhabStore := setupTestServerWithInhabitantStore(t)

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
	server, inhabStore := setupTestServerWithInhabitantStore(t)

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
	server, _ := setupTestServerWithInhabitantStore(t)

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
