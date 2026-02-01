package inhabitant

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "file:test_inhabitant?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestCreateInhabitant_Success(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	inhabitant := &Inhabitant{
		FirstName:  "John",
		LastName:   "Doe",
		MiddleName: "Michael",
		Suffix:     "Jr.",
		ContactNo:  "555-1234",
		Address:    "123 Main St",
	}

	id, err := store.CreateInhabitant(ctx, inhabitant)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
}

func TestGetInhabitant_Success(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	inhabitant := &Inhabitant{
		FirstName:  "Jane",
		LastName:   "Smith",
		MiddleName: "Marie",
		ContactNo:  "555-5678",
		Address:    "456 Oak Ave",
	}

	id, err := store.CreateInhabitant(ctx, inhabitant)
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	retrieved, err := store.GetInhabitant(ctx, id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.FirstName != inhabitant.FirstName {
		t.Errorf("expected first name %q, got %q", inhabitant.FirstName, retrieved.FirstName)
	}

	if retrieved.LastName != inhabitant.LastName {
		t.Errorf("expected last name %q, got %q", inhabitant.LastName, retrieved.LastName)
	}
}

func TestGetInhabitant_NotFound(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	_, err := store.GetInhabitant(ctx, 9999)
	if err == nil {
		t.Errorf("expected error for non-existent inhabitant")
	}
}

func TestListInhabitants_Success(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	inhabitants := []*Inhabitant{
		{FirstName: "Alice", LastName: "Adams", ContactNo: "555-0001"},
		{FirstName: "Bob", LastName: "Brown", ContactNo: "555-0002"},
		{FirstName: "Carol", LastName: "Clark", ContactNo: "555-0003"},
	}

	for _, inh := range inhabitants {
		_, err := store.CreateInhabitant(ctx, inh)
		if err != nil {
			t.Fatalf("failed to create inhabitant: %v", err)
		}
	}

	list, err := store.ListInhabitants(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(list) < len(inhabitants) {
		t.Errorf("expected at least %d inhabitants, got %d", len(inhabitants), len(list))
	}
}

func TestUpdateInhabitant_Success(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	inhabitant := &Inhabitant{
		FirstName: "David",
		LastName:  "Davis",
		ContactNo: "555-9999",
	}

	id, err := store.CreateInhabitant(ctx, inhabitant)
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	inhabitant.ID = id
	inhabitant.FirstName = "Daniel"
	inhabitant.ContactNo = "555-8888"

	err = store.UpdateInhabitant(ctx, inhabitant)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrieved, err := store.GetInhabitant(ctx, id)
	if err != nil {
		t.Fatalf("failed to get inhabitant: %v", err)
	}

	if retrieved.FirstName != "Daniel" {
		t.Errorf("expected first name Daniel, got %q", retrieved.FirstName)
	}

	if retrieved.ContactNo != "555-8888" {
		t.Errorf("expected contact number 555-8888, got %q", retrieved.ContactNo)
	}
}

func TestDeleteInhabitant_Success(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	inhabitant := &Inhabitant{
		FirstName: "Eve",
		LastName:  "Evans",
		ContactNo: "555-7777",
	}

	id, err := store.CreateInhabitant(ctx, inhabitant)
	if err != nil {
		t.Fatalf("failed to create inhabitant: %v", err)
	}

	err = store.DeleteInhabitant(ctx, id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = store.GetInhabitant(ctx, id)
	if err == nil {
		t.Errorf("expected error after delete")
	}
}

func TestDeleteInhabitant_NotFound(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	err := store.DeleteInhabitant(ctx, 9999)
	if err == nil {
		t.Errorf("expected error for non-existent inhabitant")
	}
}
