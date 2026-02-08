package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

type DocumentService struct {
	docs    DocumentStore
	files   FileStore
	cache   CacheStore
	backups BackupStore
}

func NewDocumentService(
	docs DocumentStore,
	files FileStore,
	cache CacheStore,
	backups BackupStore,
) *DocumentService {
	return &DocumentService{
		docs,
		files,
		cache,
		backups,
	}
}

func (s *DocumentService) Create(ctx context.Context, doc Document) (*Document, error) {
	if err := validateDocumentTitle(doc.Title); err != nil {
		return nil, err
	}

	return s.docs.Create(ctx, &doc)
}

func (s *DocumentService) GetByUUID(ctx context.Context, uuid string) (*Document, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}

	doc, err := s.docs.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("fetching document %s: %w", uuid, err)
	}

	return doc, nil
}

func (s *DocumentService) GetByCode(ctx context.Context, code string) (*Document, error) {
	if code == "" {
		return nil, errors.New("code is required")
	}

	return s.docs.GetByCode(ctx, code)
}

func (s *DocumentService) Update(ctx context.Context, uuid string, doc Document) (*Document, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}
	if err := validateDocumentTitle(doc.Title); err != nil {
		return nil, err
	}

	return s.docs.Update(ctx, uuid, &doc)
}

func (s *DocumentService) Delete(ctx context.Context, uuid string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}

	return s.docs.Delete(ctx, uuid)
}

func (s *DocumentService) List(ctx context.Context, offset int, limit int, sortBy string, sortDesc bool) ([]*Document, error) {
	return s.docs.List(ctx, offset, limit, sortBy, sortDesc)
}

func (s *DocumentService) UploadFile(ctx context.Context, uuid string, fileName string, content io.Reader) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if fileName == "" {
		return errors.New("file name is required")
	}
	if content == nil {
		return errors.New("file content is required")
	}

	return s.files.UploadFile(ctx, uuid, fileName, content)
}

func (s *DocumentService) DownloadFile(ctx context.Context, uuid string, fileName string) (io.ReadCloser, error) {
	if uuid == "" || fileName == "" {
		return nil, errors.New("uuid and fileName are required")
	}
	return s.files.DownloadFile(ctx, uuid, fileName)
}

func (s *DocumentService) UpdateFileMetadata(ctx context.Context, uuid string, fileName string, updates FileMetadataUpdate) error {
	if uuid == "" || fileName == "" {
		return errors.New("uuid and fileName are required")
	}
	return s.files.UpdateFileMetadata(ctx, uuid, fileName, updates)
}

func (s *DocumentService) DeleteFile(ctx context.Context, uuid string, fileName string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if fileName == "" {
		return errors.New("file name is required")
	}

	return s.files.DeleteFile(ctx, uuid, fileName)
}

func (s *DocumentService) UpdateFileContents(ctx context.Context, uuid string, fileName string, content io.Reader) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if fileName == "" {
		return errors.New("file name is required")
	}
	if content == nil {
		return errors.New("file content is required")
	}

	return s.files.UpdateFileContents(ctx, uuid, fileName, content)
}

func (s *DocumentService) RenameFile(ctx context.Context, uuid string, oldName string, newName string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if oldName == "" || newName == "" {
		return errors.New("old and new file names are required")
	}

	return s.files.RenameFile(ctx, uuid, oldName, newName)
}

func (s *DocumentService) GetConflicts(ctx context.Context) ([]SyncIssue, error) {
	return s.cache.GetConflicts(ctx)
}

func (s *DocumentService) ReloadCache(ctx context.Context) ([]SyncIssue, error) {
	return s.cache.ReloadCache(ctx)
}

func (s *DocumentService) ReloadCacheForFolder(ctx context.Context, folderName string) ([]SyncIssue, error) {
	return s.cache.ReloadCacheForFolder(ctx, folderName)
}

func (s *DocumentService) Search(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
	return s.docs.Search(ctx, criteria)
}

// SearchByParticipants searches for documents where any of the provided names
// appears in either the complainants or respondents fields.
// This method is used by the search aggregator for cross-domain searches.
// It uses parallel execution for improved performance.
func (s *DocumentService) SearchByParticipants(ctx context.Context, names []string) ([]*Document, error) {
	if len(names) == 0 {
		return []*Document{}, nil
	}

	type searchResult struct {
		docs []*Document
		err  error
	}

	// Create buffered channel to collect results from all goroutines
	resultsChan := make(chan searchResult, len(names)*2)

	// Use WaitGroup to track completion of all search goroutines
	var wg sync.WaitGroup

	// Launch parallel searches for each name in both fields
	for _, name := range names {
		name := name // capture loop variable

		// Search complainants in parallel
		wg.Add(1)
		go func() {
			defer wg.Done()
			criteria := SearchCriteria{
				FieldFilters: map[string]any{
					"complainants": name,
				},
				Limit: 1000,
			}
			docs, err := s.docs.Search(ctx, criteria)
			resultsChan <- searchResult{docs: docs, err: err}
		}()

		// Search respondents in parallel
		wg.Add(1)
		go func() {
			defer wg.Done()
			criteria := SearchCriteria{
				FieldFilters: map[string]any{
					"respondents": name,
				},
				Limit: 1000,
			}
			docs, err := s.docs.Search(ctx, criteria)
			resultsChan <- searchResult{docs: docs, err: err}
		}()
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results and deduplicate using a map
	uniqueDocs := make(map[string]*Document)
	for result := range resultsChan {
		if result.err != nil {
			return nil, fmt.Errorf("parallel search failed: %w", result.err)
		}
		for _, doc := range result.docs {
			uniqueDocs[doc.UUID] = doc
		}
	}

	// Convert map to slice
	results := make([]*Document, 0, len(uniqueDocs))
	for _, doc := range uniqueDocs {
		results = append(results, doc)
	}

	return results, nil
}

// GetDocumentsByInhabitantID returns all documents linked to the given inhabitant ID.
// This uses the reverse index for O(1) lookup instead of name-based matching.
func (s *DocumentService) GetDocumentsByInhabitantID(ctx context.Context, inhabitantID int64) ([]*Document, error) {
	if inhabitantID <= 0 {
		return nil, errors.New("invalid inhabitant ID")
	}

	// Type-assert to get access to FindByInhabitantID
	store, ok := s.docs.(interface {
		FindByInhabitantID(ctx context.Context, id int64) ([]*Document, error)
	})
	if !ok {
		return nil, errors.New("store does not support inhabitant ID lookup")
	}

	return store.FindByInhabitantID(ctx, inhabitantID)
}

func (s *DocumentService) GetBackupPath() string {
	return s.backups.GetBackupPath()
}

func (s *DocumentService) ListBackups(ctx context.Context) ([]BackupFile, error) {
	return s.backups.ListBackups(ctx)
}

func (s *DocumentService) CreateBackup(ctx context.Context, onProgress func(float64)) (string, error) {
	return s.backups.CreateBackup(ctx, onProgress)
}

func (s *DocumentService) RestoreFromLocalPath(ctx context.Context, fileName string, overwrite bool, onProgress func(float64)) error {
	return s.backups.RestoreFromLocalPath(ctx, fileName, overwrite, onProgress)
}

func validateDocumentTitle(title string) error {
	if title == "" {
		return errors.New("document title is required")
	}

	return nil
}

func getFileExtension(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[i:]
		}
	}
	return ""
}
