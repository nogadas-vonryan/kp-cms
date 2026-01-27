package document

import (
	"context"
	"testing"
)

type mockRepo struct {
	DocumentRepository
	namingStrategy NamingStrategy

	OnGetByUUID   func(ctx context.Context, uuid string) (*Document, error)
	OnGetByCode   func(ctx context.Context, code string) (*Document, error)
	OnGetByFolder func(ctx context.Context, folder string) (*Document, error)
	OnCreate      func(ctx context.Context, doc *Document) (*Document, error)
	OnUpdate      func(ctx context.Context, uuid string, doc *Document) (*Document, error)
	OnDelete      func(ctx context.Context, uuid string) error
	OnList        func(ctx context.Context, offset int, limit int) ([]*Document, error)
}

func (m *mockRepo) GetByUUID(ctx context.Context, uuid string) (*Document, error) {
	if m.OnGetByUUID != nil {
		return m.OnGetByUUID(ctx, uuid)
	}
	return nil, nil
}
func (m *mockRepo) GetByCode(ctx context.Context, code string) (*Document, error) {
	if m.OnGetByCode != nil {
		return m.OnGetByCode(ctx, code)
	}
	return nil, nil
}
func (m *mockRepo) GetByFolderName(ctx context.Context, folder string) (*Document, error) {
	if m.OnGetByFolder != nil {
		return m.OnGetByFolder(ctx, folder)
	}
	return nil, nil
}
func (m *mockRepo) Create(ctx context.Context, doc *Document) (*Document, error) {
	if m.OnCreate != nil {
		return m.OnCreate(ctx, doc)
	}
	return nil, nil
}
func (m *mockRepo) Update(ctx context.Context, uuid string, doc *Document) (*Document, error) {
	if m.OnUpdate != nil {
		return m.OnUpdate(ctx, uuid, doc)
	}
	return nil, nil
}
func (m *mockRepo) Delete(ctx context.Context, uuid string) error {
	if m.OnDelete != nil {
		return m.OnDelete(ctx, uuid)
	}
	return nil
}
func (m *mockRepo) List(ctx context.Context, offset int, limit int) ([]*Document, error) {
	if m.OnList != nil {
		return m.OnList(ctx, offset, limit)
	}
	return nil, nil
}

func TestDocumentService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should fail if all are empty", func(t *testing.T) {
		service := NewDocumentService(nil)
		doc := Document{Title: "", Code: "", FolderName: ""}

		_, err := service.Create(ctx, doc)
		if err == nil {
			t.Errorf("expected error, got: %v", err)
		}
	})

	t.Run("should fail if title is empty", func(t *testing.T) {
		service := NewDocumentService(nil) // Repo not needed for title validation
		doc := Document{Title: "", Code: "0001", FolderName: "case_0001"}

		_, err := service.Create(ctx, doc)
		if err == nil || err.Error() != "document title is required" {
			t.Errorf("expected title error, got: %v", err)
		}
	})

	t.Run("should create document successfully", func(t *testing.T) {
		mockRepo := &mockRepo{
			OnGetByFolder: func(ctx context.Context, folder string) (*Document, error) {
				return nil, nil
			},
			OnGetByCode: func(ctx context.Context, code string) (*Document, error) {
				return nil, nil
			},
			OnCreate: func(ctx context.Context, doc *Document) (*Document, error) {
				return doc, nil
			},
		}
		service := NewDocumentService(mockRepo)
		doc := Document{Title: "Test Title", Code: "0001", FolderName: "case_0001"}

		_, err := service.Create(ctx, doc)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

func TestDocumentService_GetByUUID(t *testing.T) {
	ctx := context.Background()

	t.Run("should fail if uuid is empty", func(t *testing.T) {
		service := NewDocumentService(nil)

		_, err := service.GetByUUID(ctx, "")
		if err == nil || err.Error() != "uuid is required" {
			t.Errorf("expected uuid required error, got: %v", err)
		}
	})

	t.Run("should return document by uuid", func(t *testing.T) {
		expectedDoc := &Document{UUID: "test-uuid", Code: "0001", Title: "Test"}
		mockRepo := &mockRepo{
			OnGetByUUID: func(ctx context.Context, uuid string) (*Document, error) {
				return expectedDoc, nil
			},
		}
		service := NewDocumentService(mockRepo)

		doc, err := service.GetByUUID(ctx, "test-uuid")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if doc != expectedDoc {
			t.Errorf("expected document to match")
		}
	})
}

func TestDocumentService_GetByCode(t *testing.T) {
	ctx := context.Background()

	t.Run("should fail if code is empty", func(t *testing.T) {
		service := NewDocumentService(nil)

		_, err := service.GetByCode(ctx, "")
		if err == nil || err.Error() != "code is required" {
			t.Errorf("expected code required error, got: %v", err)
		}
	})

	t.Run("should return document by code", func(t *testing.T) {
		expectedDoc := &Document{UUID: "test-uuid", Code: "0001", Title: "Test"}
		mockRepo := &mockRepo{
			OnGetByCode: func(ctx context.Context, code string) (*Document, error) {
				return expectedDoc, nil
			},
		}
		service := NewDocumentService(mockRepo)

		doc, err := service.GetByCode(ctx, "0001")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if doc != expectedDoc {
			t.Errorf("expected document to match")
		}
	})
}

func TestDocumentService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("should fail if uuid is empty", func(t *testing.T) {
		service := NewDocumentService(nil)

		_, err := service.Update(ctx, "", Document{})
		if err == nil || err.Error() != "uuid is required" {
			t.Errorf("expected uuid required error, got: %v", err)
		}
	})

	t.Run("should fail if title is empty", func(t *testing.T) {
		service := NewDocumentService(nil)

		_, err := service.Update(ctx, "test-uuid", Document{Title: ""})
		if err == nil || err.Error() != "document title is required" {
			t.Errorf("expected title required error, got: %v", err)
		}
	})

	t.Run("should update document successfully", func(t *testing.T) {
		mockRepo := &mockRepo{
			OnUpdate: func(ctx context.Context, uuid string, doc *Document) (*Document, error) {
				if uuid != "test-uuid" {
					t.Errorf("expected uuid test-uuid, got: %s", uuid)
				}
				return doc, nil
			},
		}
		service := NewDocumentService(mockRepo)

		_, err := service.Update(ctx, "test-uuid", Document{Title: "Updated Title"})
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

func TestDocumentService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should fail if uuid is empty", func(t *testing.T) {
		service := NewDocumentService(nil)

		err := service.Delete(ctx, "")
		if err == nil || err.Error() != "uuid is required" {
			t.Errorf("expected uuid required error, got: %v", err)
		}
	})

	t.Run("should delete document successfully", func(t *testing.T) {
		deletedUUID := ""
		mockRepo := &mockRepo{
			OnDelete: func(ctx context.Context, uuid string) error {
				deletedUUID = uuid
				return nil
			},
		}
		service := NewDocumentService(mockRepo)

		err := service.Delete(ctx, "test-uuid")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if deletedUUID != "test-uuid" {
			t.Errorf("expected uuid test-uuid to be deleted, got: %s", deletedUUID)
		}
	})
}

func TestDocumentService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("should return empty list", func(t *testing.T) {
		mockRepo := &mockRepo{
			OnList: func(ctx context.Context, offset, limit int) ([]*Document, error) {
				return []*Document{}, nil
			},
		}
		service := NewDocumentService(mockRepo)

		docs, err := service.List(ctx, 0, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(docs) != 0 {
			t.Errorf("expected empty list, got: %v", docs)
		}
	})

	t.Run("should return list of documents", func(t *testing.T) {
		expectedDocs := []*Document{
			{UUID: "uuid-1", Code: "0001", Title: "Test 1"},
			{UUID: "uuid-2", Code: "0002", Title: "Test 2"},
		}
		mockRepo := &mockRepo{
			OnList: func(ctx context.Context, offset, limit int) ([]*Document, error) {
				return expectedDocs, nil
			},
		}
		service := NewDocumentService(mockRepo)

		docs, err := service.List(ctx, 0, 100)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(docs) != 2 {
			t.Errorf("expected 2 documents, got: %d", len(docs))
		}
		if len(docs) != len(expectedDocs) {
			t.Errorf("expected documents to match")
		}
		for i, doc := range docs {
			if doc != expectedDocs[i] {
				t.Errorf("expected documents to match")
				break
			}
		}
	})
}
