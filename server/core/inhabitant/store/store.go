package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"kpcms/server/core/document"
	"kpcms/server/core/inhabitant"
)

// InhabitantFileStore implements file-based storage for inhabitants.
type InhabitantFileStore struct {
	basePath       string
	namingStrategy document.NamingStrategy
	mu             sync.RWMutex

	// In-memory indexes
	inhabitants map[string]*inhabitant.Inhabitant // UUID -> Inhabitant
	codeToUUID  map[string]string                 // Code -> UUID
	nameIndex   map[string][]string               // normalized_name -> []UUID
}

// New creates a new InhabitantFileStore.
func New(basePath string, namingStrategy document.NamingStrategy) (*InhabitantFileStore, error) {
	store := &InhabitantFileStore{
		basePath:       basePath,
		namingStrategy: namingStrategy,
		inhabitants:    make(map[string]*inhabitant.Inhabitant),
		codeToUUID:     make(map[string]string),
		nameIndex:      make(map[string][]string),
	}

	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	// Load all inhabitants into memory
	if err := store.ReloadCache(context.Background()); err != nil {
		return nil, err
	}

	return store, nil
}

// generateFolderName generates a new folder name with a unique code.
func (s *InhabitantFileStore) generateFolderName() string {
	// Get existing codes
	var existingCodes []string
	for code := range s.codeToUUID {
		existingCodes = append(existingCodes, code)
	}

	// Calculate next code
	nextCode := s.namingStrategy.CalculateNextCode(existingCodes)

	// Generate folder name using prefix
	return s.namingStrategy.GenerateDirName(nextCode, "")
}

// Create adds a new inhabitant to the store.
func (s *InhabitantFileStore) Create(ctx context.Context, inh *inhabitant.Inhabitant) (*inhabitant.Inhabitant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	inh.UUID = uuid.New().String()
	inh.CreatedAt = now
	inh.UpdatedAt = now

	// Generate code and folder name
	folderName := s.generateFolderName()
	inh.FolderName = folderName
	code, _ := s.namingStrategy.ExtractCode(folderName)
	inh.Code = code

	// Ensure required fields have defaults
	if inh.FirstName == "" {
		inh.FirstName = "Unknown"
	}
	if inh.LastName == "" {
		inh.LastName = "Unknown"
	}

	// Write to disk
	if err := s.writeToFile(inh); err != nil {
		return nil, err
	}

	// Update indexes
	s.inhabitants[inh.UUID] = inh
	s.codeToUUID[inh.Code] = inh.UUID

	normalizedName := normalizeName(inh)
	s.nameIndex[normalizedName] = append(s.nameIndex[normalizedName], inh.UUID)

	if inh.LastName != "" && inh.FirstName != "" {
		shortName := strings.ToLower(inh.FirstName + " " + inh.LastName)
		s.nameIndex[shortName] = append(s.nameIndex[shortName], inh.UUID)

		formalName := strings.ToLower(inh.LastName + ", " + inh.FirstName)
		s.nameIndex[formalName] = append(s.nameIndex[formalName], inh.UUID)

		// Also index with suffix if present
		if inh.Suffix != "" {
			shortNameWithSuffix := strings.ToLower(inh.FirstName + " " + inh.LastName + " " + inh.Suffix)
			s.nameIndex[shortNameWithSuffix] = append(s.nameIndex[shortNameWithSuffix], inh.UUID)

			formalNameWithSuffix := strings.ToLower(inh.LastName + ", " + inh.FirstName + " " + inh.Suffix)
			s.nameIndex[formalNameWithSuffix] = append(s.nameIndex[formalNameWithSuffix], inh.UUID)
		}
	}

	return inh, nil
}

// GetByUUID retrieves an inhabitant by UUID.
func (s *InhabitantFileStore) GetByUUID(ctx context.Context, uuid string) (*inhabitant.Inhabitant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inh, exists := s.inhabitants[uuid]
	if !exists {
		return nil, os.ErrNotExist
	}
	return inh, nil
}

// GetByCode retrieves an inhabitant by code (e.g., "001-26").
func (s *InhabitantFileStore) GetByCode(ctx context.Context, code string) (*inhabitant.Inhabitant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uuid, exists := s.codeToUUID[code]
	if !exists {
		return nil, os.ErrNotExist
	}

	inh, exists := s.inhabitants[uuid]
	if !exists {
		return nil, os.ErrNotExist
	}
	return inh, nil
}

// Update modifies an existing inhabitant.
func (s *InhabitantFileStore) Update(ctx context.Context, uuid string, inh *inhabitant.Inhabitant) (*inhabitant.Inhabitant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.inhabitants[uuid]
	if !exists {
		return nil, os.ErrNotExist
	}

	// Preserve generated fields
	inh.UUID = existing.UUID
	inh.Code = existing.Code
	inh.FolderName = existing.FolderName
	inh.CreatedAt = existing.CreatedAt
	inh.UpdatedAt = time.Now()

	// Write to disk
	if err := s.writeToFile(inh); err != nil {
		return nil, err
	}

	// Update indexes
	s.inhabitants[uuid] = inh

	// Refresh name index to reflect any name changes
	refreshNameIndex(s, existing, inh)

	return inh, nil
}

// Delete removes an inhabitant by UUID.
func (s *InhabitantFileStore) Delete(ctx context.Context, uuid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inh, exists := s.inhabitants[uuid]
	if !exists {
		return os.ErrNotExist
	}

	// Remove from disk
	folderPath := filepath.Join(s.basePath, inh.FolderName)
	if err := os.RemoveAll(folderPath); err != nil {
		return err
	}

	// Remove from indexes
	// Remove from codeToUUID
	delete(s.codeToUUID, inh.Code)

	// Remove from nameIndex (all variations)
	removeFromNameIndex(s, inh)

	delete(s.inhabitants, uuid)

	return nil
}

// List retrieves a paginated list of inhabitants.
func (s *InhabitantFileStore) List(ctx context.Context, offset int, limit int) ([]*inhabitant.Inhabitant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]*inhabitant.Inhabitant, 0, len(s.inhabitants))
	for _, inh := range s.inhabitants {
		all = append(all, inh)
	}

	if offset >= len(all) {
		return []*inhabitant.Inhabitant{}, nil
	}

	if offset+limit > len(all) {
		limit = len(all) - offset
	}

	return all[offset : offset+limit], nil
}

// writeToFile writes an inhabitant to disk.
func (s *InhabitantFileStore) writeToFile(inh *inhabitant.Inhabitant) error {
	folderPath := filepath.Join(s.basePath, inh.FolderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(inh, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(folderPath, "inhabitant.json")
	return os.WriteFile(filePath, data, 0644)
}

// reloadCache loads all inhabitants from disk into memory.
func (s *InhabitantFileStore) ReloadCache(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing indexes
	s.inhabitants = make(map[string]*inhabitant.Inhabitant)
	s.codeToUUID = make(map[string]string)
	s.nameIndex = make(map[string][]string)

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		inhabitantPath := filepath.Join(s.basePath, entry.Name(), "inhabitant.json")
		data, err := os.ReadFile(inhabitantPath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var inh inhabitant.Inhabitant
		if err := json.Unmarshal(data, &inh); err != nil {
			continue // Skip malformed files
		}

		s.inhabitants[inh.UUID] = &inh
		s.codeToUUID[inh.Code] = inh.UUID

		// Build name index
		normalizedName := normalizeName(&inh)
		s.nameIndex[normalizedName] = append(s.nameIndex[normalizedName], inh.UUID)

		// Also index by parts
		if inh.LastName != "" && inh.FirstName != "" {
			shortName := strings.ToLower(inh.FirstName + " " + inh.LastName)
			s.nameIndex[shortName] = append(s.nameIndex[shortName], inh.UUID)

			formalName := strings.ToLower(inh.LastName + ", " + inh.FirstName)
			s.nameIndex[formalName] = append(s.nameIndex[formalName], inh.UUID)

			// Also index with suffix if present
			if inh.Suffix != "" {
				shortNameWithSuffix := strings.ToLower(inh.FirstName + " " + inh.LastName + " " + inh.Suffix)
				s.nameIndex[shortNameWithSuffix] = append(s.nameIndex[shortNameWithSuffix], inh.UUID)

				formalNameWithSuffix := strings.ToLower(inh.LastName + ", " + inh.FirstName + " " + inh.Suffix)
				s.nameIndex[formalNameWithSuffix] = append(s.nameIndex[formalNameWithSuffix], inh.UUID)
			}
		}
	}

	return nil
}

// GetPath returns the folder path for an inhabitant.
func (s *InhabitantFileStore) GetPath(uuid string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if inh, exists := s.inhabitants[uuid]; exists {
		return filepath.Join(s.basePath, inh.FolderName)
	}
	return ""
}

// normalizeName returns a normalized full name for indexing.
func normalizeName(inh *inhabitant.Inhabitant) string {
	var parts []string
	if inh.FirstName != "" {
		parts = append(parts, inh.FirstName)
	}
	if inh.MiddleName != "" {
		parts = append(parts, inh.MiddleName)
	}
	if inh.LastName != "" {
		parts = append(parts, inh.LastName)
	}
	if inh.Suffix != "" {
		parts = append(parts, inh.Suffix)
	}
	return strings.ToLower(strings.Join(parts, " "))
}

// removeFromNameIndex removes an inhabitant from all name index entries.
func removeFromNameIndex(s *InhabitantFileStore, inh *inhabitant.Inhabitant) {
	// Remove full name
	normalizedName := normalizeName(inh)
	if uuids, exists := s.nameIndex[normalizedName]; exists {
		var newUUIDs []string
		for _, id := range uuids {
			if id != inh.UUID {
				newUUIDs = append(newUUIDs, id)
			}
		}
		if len(newUUIDs) == 0 {
			delete(s.nameIndex, normalizedName)
		} else {
			s.nameIndex[normalizedName] = newUUIDs
		}
	}

	// Remove short name (FirstName LastName)
	if inh.FirstName != "" && inh.LastName != "" {
		shortName := strings.ToLower(inh.FirstName + " " + inh.LastName)
		if uuids, exists := s.nameIndex[shortName]; exists {
			var newUUIDs []string
			for _, id := range uuids {
				if id != inh.UUID {
					newUUIDs = append(newUUIDs, id)
				}
			}
			if len(newUUIDs) == 0 {
				delete(s.nameIndex, shortName)
			} else {
				s.nameIndex[shortName] = newUUIDs
			}
		}
	}

	// Remove formal name (LastName, FirstName)
	if inh.FirstName != "" && inh.LastName != "" {
		formalName := strings.ToLower(inh.LastName + ", " + inh.FirstName)
		if uuids, exists := s.nameIndex[formalName]; exists {
			var newUUIDs []string
			for _, id := range uuids {
				if id != inh.UUID {
					newUUIDs = append(newUUIDs, id)
				}
			}
			if len(newUUIDs) == 0 {
				delete(s.nameIndex, formalName)
			} else {
				s.nameIndex[formalName] = newUUIDs
			}
		}

		// Remove with suffix if present
		if inh.Suffix != "" {
			shortNameWithSuffix := strings.ToLower(inh.FirstName + " " + inh.LastName + " " + inh.Suffix)
			if uuids, exists := s.nameIndex[shortNameWithSuffix]; exists {
				var newUUIDs []string
				for _, id := range uuids {
					if id != inh.UUID {
						newUUIDs = append(newUUIDs, id)
					}
				}
				if len(newUUIDs) == 0 {
					delete(s.nameIndex, shortNameWithSuffix)
				} else {
					s.nameIndex[shortNameWithSuffix] = newUUIDs
				}
			}

			formalNameWithSuffix := strings.ToLower(inh.LastName + ", " + inh.FirstName + " " + inh.Suffix)
			if uuids, exists := s.nameIndex[formalNameWithSuffix]; exists {
				var newUUIDs []string
				for _, id := range uuids {
					if id != inh.UUID {
						newUUIDs = append(newUUIDs, id)
					}
				}
				if len(newUUIDs) == 0 {
					delete(s.nameIndex, formalNameWithSuffix)
				} else {
					s.nameIndex[formalNameWithSuffix] = newUUIDs
				}
			}
		}
	}
}

// refreshNameIndex removes old name entries and adds new ones for an inhabitant.
func refreshNameIndex(s *InhabitantFileStore, existing *inhabitant.Inhabitant, updated *inhabitant.Inhabitant) {
	// Remove old entries first
	removeFromNameIndex(s, existing)
	// Add new entries
	addToNameIndex(s, updated)
}

// addToNameIndex adds an inhabitant to all name index entries.
func addToNameIndex(s *InhabitantFileStore, inh *inhabitant.Inhabitant) {
	normalizedName := normalizeName(inh)
	s.nameIndex[normalizedName] = append(s.nameIndex[normalizedName], inh.UUID)

	if inh.FirstName != "" && inh.LastName != "" {
		shortName := strings.ToLower(inh.FirstName + " " + inh.LastName)
		s.nameIndex[shortName] = append(s.nameIndex[shortName], inh.UUID)

		formalName := strings.ToLower(inh.LastName + ", " + inh.FirstName)
		s.nameIndex[formalName] = append(s.nameIndex[formalName], inh.UUID)

		// Also index with suffix if present
		if inh.Suffix != "" {
			shortNameWithSuffix := strings.ToLower(inh.FirstName + " " + inh.LastName + " " + inh.Suffix)
			s.nameIndex[shortNameWithSuffix] = append(s.nameIndex[shortNameWithSuffix], inh.UUID)

			formalNameWithSuffix := strings.ToLower(inh.LastName + ", " + inh.FirstName + " " + inh.Suffix)
			s.nameIndex[formalNameWithSuffix] = append(s.nameIndex[formalNameWithSuffix], inh.UUID)
		}
	}
}
