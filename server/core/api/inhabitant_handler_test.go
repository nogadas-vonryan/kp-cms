package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"kpcms/server/core/inhabitant"

	"github.com/go-chi/chi/v5"
)

func TestHandleListInhabitants_Success(t *testing.T) {
	server, db := setupTestServerWithDB(t)

	// Create some inhabitants
	ctx := context.Background()
	inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
	for i := 0; i < 3; i++ {
		_, err := inhabitantRepo.Create(ctx, &inhabitant.Inhabitant{
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
	server, db := setupTestServerWithDB(t)

	ctx := context.Background()
	inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
	id, err := inhabitantRepo.Create(ctx, &inhabitant.Inhabitant{
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

	req := httptest.NewRequest("GET", "/inhabitants/"+strconv.FormatInt(id, 10), nil)
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
	server, _ := setupTestServerWithDB(t)

	reqBody := createInhabitantRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		ContactNo: "555-5678",
		Birthday:  "1990-01-15",
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
	server, _ := setupTestServerWithDB(t)

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
	server, db := setupTestServerWithDB(t)

	ctx := context.Background()
	inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
	id, err := inhabitantRepo.Create(ctx, &inhabitant.Inhabitant{
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

	req := httptest.NewRequest("PUT", "/inhabitants/"+strconv.FormatInt(id, 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the update
	updated, _ := inhabitantRepo.Get(ctx, id)
	if updated.FirstName != "Jonathan" {
		t.Errorf("expected first name Jonathan, got %q", updated.FirstName)
	}
}

func TestHandleDeleteInhabitant_Success(t *testing.T) {
	server, db := setupTestServerWithDB(t)

	ctx := context.Background()
	inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
	id, err := inhabitantRepo.Create(ctx, &inhabitant.Inhabitant{
		FirstName: "Alice",
		LastName:  "Adams",
	})
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Delete("/inhabitants/{id}", server.handleDeleteInhabitant())

	req := httptest.NewRequest("DELETE", "/inhabitants/"+strconv.FormatInt(id, 10), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify the deletion
	_, err = inhabitantRepo.Get(ctx, id)
	if err == nil {
		t.Errorf("expected error after deletion")
	}
}

func TestHandleDeleteInhabitant_NotFound(t *testing.T) {
	server, _ := setupTestServerWithDB(t)

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Delete("/inhabitants/{id}", server.handleDeleteInhabitant())

	req := httptest.NewRequest("DELETE", "/inhabitants/9999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
