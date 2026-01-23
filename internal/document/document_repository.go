package document

import (
	"context"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByUUID(ctx context.Context, uuid string) (*DocumentResponse, error)
	GetByCode(ctx context.Context, code string) (*DocumentResponse, error)
	GetByFolderName(ctx context.Context, folderName string) (*DocumentResponse, error)
	Update(ctx context.Context, uuid string, doc *Document) error
	Delete(ctx context.Context, uuid string) error
	List(ctx context.Context) ([]*Document, error)

	AddFile(ctx context.Context, uuid string, file File) error
	DeleteFile(ctx context.Context, uuid string, fileName string) error
}
