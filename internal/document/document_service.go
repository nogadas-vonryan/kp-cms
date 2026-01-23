package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/google/uuid"
)

type DocumentService struct {
	repo DocumentRepository
}

func NewDocumentService(repo DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) Create(ctx context.Context, doc Document) error {
	if err := validateDocumentTitle(doc.Title); err != nil {
		return err
	}
	if doc.Code == "" {
		return errors.New("document code is required")
	}
	if doc.FolderName == "" {
		return errors.New("document folder name is required")
	}

	existsByFolder, err := s.repo.GetByFolderName(ctx, doc.FolderName)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if existsByFolder != nil {
		return errors.New("document by that folder name already exists")
	}

	existsByCode, err := s.repo.GetByCode(ctx, doc.Code)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if existsByCode != nil {
		return errors.New("document by that code already exists")
	}

	now := time.Now().UTC()
	doc.UUID = uuid.NewString()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	if doc.Fields == nil {
		doc.Fields = map[string]any{}
	}

	return s.repo.Create(ctx, &doc)
}

func (s *DocumentService) GetByUUID(ctx context.Context, uuid string) (*DocumentResponse, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required")
	}

	doc, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("fetching document %s: %w", uuid, err)
	}

	return doc, nil
}

func (s *DocumentService) GetByCode(ctx context.Context, code string) (*DocumentResponse, error) {
	if code == "" {
		return nil, errors.New("code is required")
	}

	return s.repo.GetByCode(ctx, code)
}

func (s *DocumentService) GetByFolderName(ctx context.Context, folderName string) (*DocumentResponse, error) {
	if folderName == "" {
		return nil, errors.New("folder name is required")
	}

	return s.repo.GetByFolderName(ctx, folderName)
}

func (s *DocumentService) Update(ctx context.Context, uuid string, doc Document) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}
	if err := validateDocumentTitle(doc.Title); err != nil {
		return err
	}

	doc.UpdatedAt = time.Now().UTC()
	if doc.Fields == nil {
		doc.Fields = map[string]any{}
	}

	return s.repo.Update(ctx, uuid, &doc)
}

func (s *DocumentService) Delete(ctx context.Context, uuid string) error {
	if uuid == "" {
		return errors.New("uuid is required")
	}

	return s.repo.Delete(ctx, uuid)
}

func (s *DocumentService) List(ctx context.Context) ([]*Document, error) {
	return s.repo.List(ctx)
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

	doc, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("fetching document: %w", err)
	}

	fileRepo, ok := s.repo.(*FileDocumentRepository)
	if !ok {
		return errors.New("repository does not support file operations")
	}

	folderPath := fileRepo.basePath + "/" + doc.FolderName
	filePath := folderPath + "/" + fileName

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	writtenBytes, err := io.Copy(outFile, content)
	if err != nil {
		return fmt.Errorf("write file content: %w", err)
	}

	fileInfo, err := outFile.Stat()
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	file := File{
		FileName:  fileName,
		Type:      getFileExtension(fileName),
		Size:      writtenBytes,
		CreatedAt: fileInfo.ModTime(),
	}

	return s.repo.AddFile(ctx, uuid, file)
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
