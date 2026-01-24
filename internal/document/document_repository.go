package document

import (
	"context"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) (*Document, error)
	GetByUUID(ctx context.Context, uuid string) (*Document, error)
	GetByCode(ctx context.Context, code string) (*Document, error)
	Update(ctx context.Context, uuid string, doc *Document) (*Document, error)
	Delete(ctx context.Context, uuid string) error
	List(ctx context.Context, offset int, limit int) ([]*Document, error)

	AddFile(ctx context.Context, uuid string, file File) error
	DeleteFile(ctx context.Context, uuid string, fileName string) error
}
