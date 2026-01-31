package document

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type DocumentService struct {
	repo DocumentRepository
}

func NewDocumentService(repo DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) Create(ctx context.Context, doc Document) (*Document, error) {
	if err := validateDocumentTitle(doc.Title); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &doc)
}

func (s *DocumentService) GetByUUID(ctx context.Context, uuid string) (*Document, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}

	doc, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("fetching document %s: %w", uuid, err)
	}

	return doc, nil
}

func (s *DocumentService) GetByCode(ctx context.Context, code string) (*Document, error) {
	if code == "" {
		return nil, errors.New("code is required")
	}

	return s.repo.GetByCode(ctx, code)
}

func (s *DocumentService) Update(ctx context.Context, uuid string, doc Document) (*Document, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}
	if err := validateDocumentTitle(doc.Title); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, uuid, &doc)
}

func (s *DocumentService) Delete(ctx context.Context, uuid string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}

	return s.repo.Delete(ctx, uuid)
}

func (s *DocumentService) List(ctx context.Context, offset int, limit int, sortBy string, sortDesc bool) ([]*Document, error) {
	return s.repo.List(ctx, offset, limit, sortBy, sortDesc)
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

	// Delegate to repository which handles path construction, security, and file I/O
	return s.repo.UploadFile(ctx, uuid, fileName, content)
}

func (s *DocumentService) DownloadFile(ctx context.Context, uuid string, fileName string) (io.ReadCloser, error) {
	if uuid == "" || fileName == "" {
		return nil, errors.New("uuid and fileName are required")
	}
	return s.repo.DownloadFile(ctx, uuid, fileName)
}

func (s *DocumentService) UpdateFileMetadata(ctx context.Context, uuid string, fileName string, updates FileMetadataUpdate) error {
	if uuid == "" || fileName == "" {
		return errors.New("uuid and fileName are required")
	}
	return s.repo.UpdateFileMetadata(ctx, uuid, fileName, updates)
}

func (s *DocumentService) DeleteFile(ctx context.Context, uuid string, fileName string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if fileName == "" {
		return errors.New("file name is required")
	}

	return s.repo.DeleteFile(ctx, uuid, fileName)
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

	return s.repo.UpdateFileContents(ctx, uuid, fileName, content)
}

func (s *DocumentService) RenameFile(ctx context.Context, uuid string, oldName string, newName string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if oldName == "" || newName == "" {
		return errors.New("old and new file names are required")
	}

	return s.repo.RenameFile(ctx, uuid, oldName, newName)
}

func (s *DocumentService) GetConflicts(ctx context.Context) ([]SyncIssue, error) {
	return s.repo.GetConflicts(ctx)
}

func (s *DocumentService) ReloadCache(ctx context.Context) ([]SyncIssue, error) {
	return s.repo.ReloadCache(ctx)
}

func (s *DocumentService) ReloadCacheForFolder(ctx context.Context, folderName string) ([]SyncIssue, error) {
	return s.repo.ReloadCacheForFolder(ctx, folderName)
}

func (s *DocumentService) Search(ctx context.Context, criteria SearchCriteria) ([]*Document, error) {
	return s.repo.Search(ctx, criteria)
}

func (s *DocumentService) GetBackupPath() string {
	return s.repo.GetBackupPath()
}

func (s *DocumentService) ListBackups(ctx context.Context) ([]BackupFile, error) {
	return s.repo.ListBackups(ctx)
}

func (s *DocumentService) CreateBackup(ctx context.Context) (string, error) {
	return s.repo.CreateBackup(ctx)
}

func (s *DocumentService) RestoreFromLocalPath(ctx context.Context, fileName string, overwrite bool, onProgress func(float64)) error {
	return s.repo.RestoreFromLocalPath(ctx, fileName, overwrite, onProgress)
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
