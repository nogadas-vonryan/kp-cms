package store

import (
	"context"
	"errors"
	"fmt"
	"kpcms/server/core/document"
	"os"
	"sync"
)

type Store struct {
	basePath       string
	backupPath     string
	namingStrategy document.NamingStrategy

	mu sync.RWMutex

	documents    map[string]*document.Document
	codeToUUID   map[string]string
	sortedByCode []string

	lastConflicts []document.SyncIssue
}

var _ document.DocumentStore = (*Store)(nil)
var _ document.FileStore = (*Store)(nil)
var _ document.CacheStore = (*Store)(nil)
var _ document.BackupStore = (*Store)(nil)

func New(basePath string, backupPath string, namingStrategy document.NamingStrategy) (*Store, error) {
	if basePath == "" {
		return nil, errors.New("base path is required")
	}
	if namingStrategy == nil {
		return nil, errors.New("naming strategy is required")
	}

	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("ensure base path: %w", err)
	}

	repo := &Store{
		basePath:       basePath,
		backupPath:     backupPath,
		namingStrategy: namingStrategy,
		documents:      make(map[string]*document.Document),
		codeToUUID:     make(map[string]string),
		sortedByCode:   make([]string, 0),
	}

	// Warning: Lock first (so multiple users can read/write at once without crashing)
	_, err := repo.ReloadCache(context.Background())
	return repo, err
}
