package store

import (
	"context"
	"errors"
	"fmt"
	"kpcms/server/core/document"
	"os"
	"sync"
)

type FileDocumentRepository struct {
	basePath       string
	backupPath     string
	namingStrategy document.NamingStrategy
	mu             sync.RWMutex
	sortedByCode   []string
	cacheByUUID    map[string]*document.Document
	cacheByCode    map[string]*document.Document
	lastConflicts  []document.SyncIssue
}

var _ document.DocumentStore = (*FileDocumentRepository)(nil)
var _ document.FileStore = (*FileDocumentRepository)(nil)
var _ document.CacheStore = (*FileDocumentRepository)(nil)
var _ document.BackupStore = (*FileDocumentRepository)(nil)

func NewFileDocumentRepository(basePath string, backupPath string, namingStrategy document.NamingStrategy) (*FileDocumentRepository, error) {
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
		backupPath:     backupPath,
		namingStrategy: namingStrategy,
		cacheByUUID:    make(map[string]*document.Document),
		cacheByCode:    make(map[string]*document.Document),
		sortedByCode:   make([]string, 0),
	}

	// Warning: Lock first (so multiple users can read/write at once without crashing)
	_, err := repo.ReloadCache(context.Background())
	return repo, err
}
