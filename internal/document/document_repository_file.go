package document

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SyncIssue struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type FileDocumentRepository struct {
	basePath       string
	namingStrategy NamingStrategy
	mu             sync.RWMutex
	sortedByCode   []string
	cacheByUUID    map[string]*Document
	cacheByCode    map[string]*Document
	lastConflicts  []SyncIssue
}

var _ DocumentRepository = (*FileDocumentRepository)(nil)

func NewFileDocumentRepository(basePath string, namingStrategy NamingStrategy) (*FileDocumentRepository, error) {
	if basePath == "" {
		return nil, errors.New("base path is required")
	}
	if namingStrategy == nil {
		return nil, errors.New("naming strategy is required")
	}

	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("ensure base path: %w", err)
	}

	repo := &FileDocumentRepository{
		basePath:       basePath,
		namingStrategy: namingStrategy,
		cacheByUUID:    make(map[string]*Document),
		cacheByCode:    make(map[string]*Document),
		sortedByCode:   make([]string, 0),
	}

	// Warning: Lock first (so multiple users can read/write at once without crashing)
	_, err := repo.ReloadCache(context.Background())
	return repo, err
}

func (r *FileDocumentRepository) Create(ctx context.Context, doc *Document) (*Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("document is nil")
	}

	now := time.Now().UTC()
	doc.UUID = uuid.NewString()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	if doc.Fields == nil {
		doc.Fields = map[string]any{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	codes := r.getNextCode()
	nextCode := r.namingStrategy.CalculateNextCode(codes)
	doc.FolderName = r.namingStrategy.GenerateDirName(nextCode, doc.Title)
	doc.Code = nextCode

	// All file I/O inside lock to prevent race condition
	folderPath := filepath.Join(r.basePath, doc.FolderName)
	metaPath := filepath.Join(folderPath, "meta.json")

	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}

	if _, err := os.Stat(metaPath); err == nil {
		return nil, fmt.Errorf("document already exists: %w", fs.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("check existing document: %w", err)
	}

	if err := writeDocument(metaPath, doc); err != nil {
		return nil, err
	}

	filesPath := filepath.Join(folderPath, "files.json")
	if err := os.WriteFile(filesPath, []byte("[]"), 0644); err != nil {
		return nil, fmt.Errorf("failed to initialize files.json: %w", err)
	}

	r.addToCache(doc)
	r.rebuildSortedCodesLocked()

	return doc, nil
}

func (r *FileDocumentRepository) GetByUUID(ctx context.Context, uuidValue string) (*Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, exists := r.cacheByUUID[uuidValue]
	if !exists {
		return nil, fs.ErrNotExist
	}

	files, err := r.readFiles(doc.FolderName)
	if err != nil {
		return nil, err
	}

	resp := &Document{
		UUID:       doc.UUID,
		Code:       doc.Code,
		FolderName: doc.FolderName,
		Title:      doc.Title,
		Fields:     doc.Fields,
		Files:      files,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
	return resp, nil
}

func (r *FileDocumentRepository) GetByCode(ctx context.Context, code string) (*Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, exists := r.cacheByCode[code]
	if !exists {
		return nil, fs.ErrNotExist
	}

	files, err := r.readFiles(doc.FolderName)
	if err != nil {
		return nil, err
	}

	resp := &Document{
		UUID:       doc.UUID,
		Code:       doc.Code,
		FolderName: doc.FolderName,
		Title:      doc.Title,
		Fields:     doc.Fields,
		Files:      files,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
	return resp, nil
}

func (r *FileDocumentRepository) Update(ctx context.Context, uuidValue string, doc *Document) (*Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("document is nil")
	}

	existing, err := r.GetByUUID(ctx, uuidValue)
	if err != nil {
		return nil, err
	}

	updated := *doc
	updated.UUID = existing.UUID
	updated.FolderName = existing.FolderName
	updated.Code = existing.Code
	if updated.CreatedAt.IsZero() {
		updated.CreatedAt = existing.CreatedAt
	}
	updated.UpdatedAt = time.Now().UTC()
	if updated.Fields == nil {
		updated.Fields = map[string]any{}
	}

	metaPath := filepath.Join(r.basePath, existing.FolderName, "meta.json")
	if err := writeDocument(metaPath, &updated); err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.addToCache(&updated)
	r.rebuildSortedCodesLocked()
	r.mu.Unlock()

	return &updated, nil
}

func (r *FileDocumentRepository) Delete(ctx context.Context, uuidValue string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuidValue)
	if err != nil {
		return err
	}

	folderPath := filepath.Join(r.basePath, doc.FolderName)
	if err := os.RemoveAll(folderPath); err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	r.mu.Lock()
	delete(r.cacheByUUID, uuidValue)
	delete(r.cacheByCode, doc.Code)
	r.rebuildSortedCodesLocked()
	r.mu.Unlock()

	return nil
}

func (r *FileDocumentRepository) List(ctx context.Context, offset int, limit int) ([]*Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.sortedByCode)
	if offset >= total || offset < 0 {
		return []*Document{}, nil
	}

	end := offset + limit
	if end > total || limit <= 0 {
		end = total
	}

	pageKeys := r.sortedByCode[offset:end]
	docs := make([]*Document, 0, len(pageKeys))

	for _, code := range pageKeys {
		doc, exists := r.cacheByCode[code]
		if !exists {
			continue
		}

		files, err := r.readFiles(doc.FolderName)
		if err != nil {
			return nil, err
		}

		doc.Files = files
		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *FileDocumentRepository) GetDocumentFolderPath(ctx context.Context, uuid string) (string, error) {
	if err := ctxErr(ctx); err != nil {
		return "", err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return "", fmt.Errorf("get document: %w", err)
	}

	return r.getDocumentPath(doc.FolderName), nil
}
