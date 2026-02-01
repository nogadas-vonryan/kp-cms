package store

import (
	"context"
	"fmt"
	"kpcms/server/core/document"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func (r *Store) ReloadCache(ctx context.Context) ([]document.SyncIssue, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	log.Printf("[Cache] Starting reload of documents from %s", r.basePath)
	start := time.Now()

	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	tempDocs := make(map[string]*document.Document)
	tempCodeToUUID := make(map[string]string)
	var tempSortedCodes []string
	var issues []document.SyncIssue

	for _, entry := range entries {
		if err := ctxErr(ctx); err != nil {
			return nil, err
		}

		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		metaPath := filepath.Join(r.basePath, folderName, "meta.json")

		doc, err := readDocument(metaPath)
		if err != nil {
			issues = append(issues, document.SyncIssue{
				Type:    "MISSING_META",
				Path:    folderName,
				Message: fmt.Sprintf("could not read meta.json: %v", err),
			})
			continue
		}

		folderCode, ok := r.namingStrategy.ExtractCode(folderName)
		if !ok {
			if doc.Code == "" {
				issues = append(issues, document.SyncIssue{
					Type:    "INVALID_FOLDER",
					Path:    folderName,
					Message: "folder name does not match naming strategy and meta.json has no code",
				})
				continue
			}
			folderCode = doc.Code
		}

		// If the meta.json code differs from the folder code, fix it
		if doc.Code != folderCode {
			log.Printf("[Cache] Fixing code mismatch for %s: meta(%s) -> folder(%s)", folderName, doc.Code, folderCode)
			doc.Code = folderCode

			// Safety: Capture variables for the goroutine to prevent race conditions
			go func(path string, d document.Document, folder string) {
				if err := writeDocument(path, &d); err != nil {
					log.Printf("[Cache] Auto-heal failed for %s: %v", folder, err)
				}
			}(metaPath, *doc, folderName)
		}

		if existingUUID, exists := tempCodeToUUID[doc.Code]; exists {
			existing := tempDocs[existingUUID]
			issues = append(issues, document.SyncIssue{
				Type:    "DUPLICATE_CODE",
				Path:    folderName,
				Message: fmt.Sprintf("code '%s' already claimed by folder '%s'", doc.Code, existing.FolderName),
			})
			continue
		}

		if existing, exists := tempDocs[doc.UUID]; exists {
			issues = append(issues, document.SyncIssue{
				Type:    "DUPLICATE_UUID",
				Path:    folderName,
				Message: fmt.Sprintf("UUID '%s' already claimed by folder '%s'", doc.UUID, existing.FolderName),
			})
			continue
		}

		doc.FolderName = folderName
		tempDocs[doc.UUID] = doc
		tempCodeToUUID[doc.Code] = doc.UUID
		tempSortedCodes = append(tempSortedCodes, doc.Code)
	}

	sort.Strings(tempSortedCodes)

	r.mu.Lock()
	r.documents = tempDocs
	r.codeToUUID = tempCodeToUUID
	r.sortedByCode = tempSortedCodes
	r.lastConflicts = issues
	r.mu.Unlock()

	log.Printf("[Cache] Reloaded %d documents from %s in %v (found %d issues)",
		len(tempSortedCodes), r.basePath, time.Since(start), len(issues))

	return issues, nil
}

func (r *Store) ReloadCacheForFolder(ctx context.Context, folderName string) ([]document.SyncIssue, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	log.Printf("[Cache] Starting reload of single document folder: %s", folderName)
	start := time.Now()

	var issues []document.SyncIssue
	metaPath := filepath.Join(r.basePath, folderName, "meta.json")

	// Check if folder exists
	folderPath := filepath.Join(r.basePath, folderName)
	if stat, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("folder does not exist: %s", folderName)
		}
		return nil, fmt.Errorf("failed to access folder: %w", err)
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", folderName)
	}

	// Read document metadata
	doc, err := readDocument(metaPath)
	if err != nil {
		issues = append(issues, document.SyncIssue{
			Type:    "MISSING_META",
			Path:    folderName,
			Message: fmt.Sprintf("could not read meta.json: %v", err),
		})
		return issues, fmt.Errorf("failed to read meta.json: %w", err)
	}

	// Extract and validate folder code
	folderCode, ok := r.namingStrategy.ExtractCode(folderName)
	if !ok {
		if doc.Code == "" {
			issues = append(issues, document.SyncIssue{
				Type:    "INVALID_FOLDER",
				Path:    folderName,
				Message: "folder name does not match naming strategy and meta.json has no code",
			})
			return issues, fmt.Errorf("invalid folder name and missing code in meta.json")
		}
		folderCode = doc.Code
	}

	// Fix code mismatch if needed
	if doc.Code != folderCode {
		log.Printf("[Cache] Fixing code mismatch for %s: meta(%s) -> folder(%s)", folderName, doc.Code, folderCode)
		oldCode := doc.Code
		doc.Code = folderCode

		// Write the fix synchronously for single folder reload
		if err := writeDocument(metaPath, doc); err != nil {
			log.Printf("[Cache] Auto-heal failed for %s: %v", folderName, err)
			issues = append(issues, document.SyncIssue{
				Type:    "AUTO_HEAL_FAILED",
				Path:    folderName,
				Message: fmt.Sprintf("failed to fix code mismatch: %v", err),
			})
		}

		// Update cache - need to remove old code entry if it exists
		r.mu.Lock()
		if oldCode != "" && oldCode != folderCode {
			delete(r.codeToUUID, oldCode)
		}
		r.mu.Unlock()
	}

	// Check for conflicts with existing cache entries
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingUUID, exists := r.codeToUUID[doc.Code]; exists {
		existing := r.documents[existingUUID]
		if existing != nil && existing.FolderName != folderName {
			issues = append(issues, document.SyncIssue{
				Type:    "DUPLICATE_CODE",
				Path:    folderName,
				Message: fmt.Sprintf("code '%s' already claimed by folder '%s'", doc.Code, existing.FolderName),
			})
			return issues, fmt.Errorf("duplicate code conflict: %s", doc.Code)
		}
	}

	if existing, exists := r.documents[doc.UUID]; exists && existing.FolderName != folderName {
		issues = append(issues, document.SyncIssue{
			Type:    "DUPLICATE_UUID",
			Path:    folderName,
			Message: fmt.Sprintf("UUID '%s' already claimed by folder '%s'", doc.UUID, existing.FolderName),
		})
		return issues, fmt.Errorf("duplicate UUID conflict: %s", doc.UUID)
	}

	// Update cache
	doc.FolderName = folderName
	r.addToCache(doc)
	r.rebuildSortedCodesLocked()

	log.Printf("[Cache] Reloaded document folder %s in %v (found %d issues)",
		folderName, time.Since(start), len(issues))

	return issues, nil
}

func (r *Store) GetConflicts(ctx context.Context) ([]document.SyncIssue, error) {
	return r.lastConflicts, nil
}

func (r *Store) addToCache(doc *document.Document) {
	r.documents[doc.UUID] = doc
	r.codeToUUID[doc.Code] = doc.UUID
}

func (r *Store) clearCache() {
	r.documents = make(map[string]*document.Document)
	r.codeToUUID = make(map[string]string)
}

func (r *Store) getCodes() []string {
	codes := make([]string, 0, len(r.codeToUUID))
	for code := range r.codeToUUID {
		codes = append(codes, code)
	}

	return codes
}

// rebuildSortedCodesLocked recalculates the sorted codes slice.
// Caller must hold the write lock.
func (r *Store) rebuildSortedCodesLocked() {
	codes := make([]string, 0, len(r.codeToUUID))
	for code := range r.codeToUUID {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	r.sortedByCode = codes
}
