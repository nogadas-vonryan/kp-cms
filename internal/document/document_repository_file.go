package document

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FileDocumentRepository struct {
	basePath       string
	namingStrategy NamingStrategy
	mu             sync.RWMutex
	sortedByCode   []string
	cacheByUUID    map[string]*Document
	cacheByCode    map[string]*Document
}

func NewFileDocumentRepository(basePath string, namingStrategy NamingStrategy) (DocumentRepository, error) {
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
		cacheByFolder:  make(map[string]*Document),
		cacheByCode:    make(map[string]*Document),
		sortedByCode:   make([]string, 0),
	}

	// Warning: Lock first (so multiple users can read/write at once without crashing)
	err := repo.reloadCache()
	return repo, err
}

func (r *FileDocumentRepository) reloadCache() error {
	log.Printf("[Cache] Starting reload of documents from %s", r.basePath)
	start := time.Now()

	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return err
	}

	tempUUID := make(map[string]*Document)
	tempCode := make(map[string]*Document)
	tempFolder := make(map[string]*Document)
	var tempSortedCodes []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(r.basePath, entry.Name(), "meta.json")
		doc, err := readDocument(metaPath)
		if err != nil {
			continue
		}

		tempUUID[doc.UUID] = doc
		tempCode[doc.Code] = doc
		tempFolder[doc.FolderName] = doc
		tempSortedCodes = append(tempSortedCodes, doc.Code)
	}

	sort.Strings(tempSortedCodes)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.cacheByUUID = tempUUID
	r.cacheByCode = tempCode
	r.sortedByCode = tempSortedCodes

	log.Printf("[Cache] Reloaded %d documents from %s in %v", len(tempSortedCodes), r.basePath, time.Since(start))

	return nil
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
	doc.CreatedAt = now
	doc.UpdatedAt = now
	if doc.Fields == nil {
		doc.Fields = map[string]any{}
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	r.mu.Lock()
	codes := r.getNextCode()
	nextCode := r.namingStrategy.CalculateNextCode(codes)
	doc.FolderName = r.namingStrategy.GenerateDirName(nextCode, doc.Title)

	doc.Code = nextCode
	r.mu.Unlock()

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

	r.mu.Lock()
	r.addToCache(doc)
	r.mu.Unlock()

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
	updated.CreatedAt = existing.CreatedAt
	updated.UpdatedAt = time.Now().UTC()
	if updated.Fields == nil {
		updated.Fields = map[string]any{}
	}

	metaPath := filepath.Join(r.basePath, existing.FolderName, "meta.json")
	if err := writeDocument(metaPath, &updated); err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.addToCache(doc)
	r.mu.Unlock()

	return doc, nil
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

func (r *FileDocumentRepository) AddFile(ctx context.Context, uuid string, file File) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := filepath.Join(r.basePath, doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []File
	data, err := os.ReadFile(filesJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &files); err != nil {
			return fmt.Errorf("decode files metadata: %w", err)
		}
	}

	// Check if file already exists
	for i, f := range files {
		if f.FileName == file.FileName {
			// Update existing file metadata
			files[i] = file
			return writeFilesMetadata(filesJSONPath, files)
		}
	}

	files = append(files, file)
	return writeFilesMetadata(filesJSONPath, files)
}

func (r *FileDocumentRepository) DeleteFile(ctx context.Context, uuid string, fileName string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := filepath.Join(r.basePath, doc.FolderName)
	filePath := filepath.Join(folderPath, fileName)

	// Delete physical file
	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete physical file: %w", err)
	}

	// Update files.json
	filesJSONPath := filepath.Join(folderPath, "files.json")
	var files []File
	data, err := os.ReadFile(filesJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &files); err != nil {
			return fmt.Errorf("decode files metadata: %w", err)
		}

		// Remove file from metadata
		for i, f := range files {
			if f.FileName == fileName {
				files = append(files[:i], files[i+1:]...)
				break
			}
		}

		return writeFilesMetadata(filesJSONPath, files)
	}

	return nil
}

func (r *FileDocumentRepository) scanPhysicalFolder(folderPath string) ([]File, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var files []File
	for _, entry := range entries {
		// Ignore directories and the metadata file itself
		if entry.IsDir() || entry.Name() == "meta.json" || entry.Name() == "files.json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, File{
			FileName:  entry.Name(),
			Type:      filepath.Ext(entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			// Description and Note will be empty because we are recovering from raw files
		})
	}
	return files, nil
}

func readDocument(metaPath string) (*Document, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fs.ErrNotExist
		}
		return nil, fmt.Errorf("read document: %w", err)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode document: %w", err)
	}

	return &doc, nil
}

func (r *FileDocumentRepository) readFiles(folderName string) ([]File, error) {
	folderPath := filepath.Join(r.basePath, folderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	data, err := os.ReadFile(filesJSONPath)
	if errors.Is(err, os.ErrNotExist) {
		files, err := r.scanPhysicalFolder(folderPath)
		if err != nil {
			return nil, err
		}

		if err := writeFilesMetadata(filesJSONPath, files); err != nil {
			return nil, err
		}
		return files, nil
	}

	var metadataList []File
	if err := json.Unmarshal(data, &metadataList); err != nil {
		return nil, err
	}

	syncedFiles := make([]File, 0, len(metadataList))
	for _, f := range metadataList {
		physicalPath := filepath.Join(folderPath, f.FileName)

		// Get physical info (Size, ModTime) from the OS
		info, err := os.Stat(physicalPath)
		if err == nil {
			// Merge: Physical info + JSON metadata
			f.Size = info.Size()
			f.CreatedAt = info.ModTime()
			f.Type = filepath.Ext(f.FileName)

			syncedFiles = append(syncedFiles, f)
		}
	}

	return syncedFiles, nil
}

func writeDocument(metaPath string, doc *Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode document: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("write document: %w", err)
	}

	return nil
}

func writeFilesMetadata(path string, files []File) error {
	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return fmt.Errorf("encode files metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write files metadata: %w", err)
	}

	return nil
}

func (r *FileDocumentRepository) addToCache(doc *Document) {
	r.cacheByUUID[doc.UUID] = doc
	r.cacheByCode[doc.Code] = doc
}

func (r *FileDocumentRepository) clearCache() {
	r.cacheByUUID = make(map[string]*Document)
	r.cacheByCode = make(map[string]*Document)
}

func (r *FileDocumentRepository) getNextCode() []string {
	codes := make([]string, 0, len(r.cacheByCode))
	for code := range r.cacheByCode {
		codes = append(codes, code)
	}

	return codes
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
