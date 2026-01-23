package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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

func (s *DocumentService) GetByFolderName(ctx context.Context, folderName string) (*Document, error) {
	if folderName == "" {
		return nil, errors.New("folder name is required")
	}

	return s.repo.GetByFolderName(ctx, folderName)
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

func (s *DocumentService) List(ctx context.Context, offset int, limit int) ([]*Document, error) {
	return s.repo.List(ctx, offset, limit)
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
