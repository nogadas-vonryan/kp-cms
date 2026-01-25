package document

import (
	"context"
	"io"
)

type DocumentStore interface {
	Create(ctx context.Context, doc *Document) (*Document, error)
	GetByUUID(ctx context.Context, uuid string) (*Document, error)
	GetByCode(ctx context.Context, code string) (*Document, error)
	Update(ctx context.Context, uuid string, doc *Document) (*Document, error)
	Delete(ctx context.Context, uuid string) error
	List(ctx context.Context, offset int, limit int) ([]*Document, error)
	Search(ctx context.Context, criteria SearchCriteria) ([]*Document, error)
}

type FileStore interface {
	// Adds the metadata of input file to the files.json
	AddFile(ctx context.Context, uuid string, file File) error
	// Uploads the raw content of a multipart form file to the store
	UploadFile(ctx context.Context, uuid string, fileName string, content io.Reader) error
	DeleteFile(ctx context.Context, uuid string, fileName string) error
	GetDocumentFolderPath(ctx context.Context, uuid string) (string, error)
}

type CacheStore interface {
	GetConflicts(ctx context.Context) ([]SyncIssue, error)
	ReloadCache(ctx context.Context) ([]SyncIssue, error)
}

type DocumentRepository interface {
	DocumentStore
	FileStore
	CacheStore
}
