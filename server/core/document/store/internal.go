package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"kpcms/server/core/document"
	"os"
	"path/filepath"
)

func readDocument(metaPath string) (*document.Document, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fs.ErrNotExist
		}
		return nil, fmt.Errorf("read document: %w", err)
	}

	var doc document.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode document: %w", err)
	}

	return &doc, nil
}

func writeDocument(metaPath string, doc *document.Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode document: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("write document: %w", err)
	}

	return nil
}

func writeFilesMetadata(path string, files []document.File) error {
	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return fmt.Errorf("encode files metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write files metadata: %w", err)
	}

	return nil
}

func (r *Store) getDocumentPath(folderName string) string {
	return filepath.Join(r.basePath, folderName)
}

func (r *Store) getDocumentFilePath(folderName, fileName string) string {
	return filepath.Join(r.basePath, folderName, fileName)
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// rebuildInhabitantIndexLocked rebuilds the inhabitant reverse index from all cached documents.
// Caller must hold the write lock.
func (s *Store) rebuildInhabitantIndexLocked() {
	s.inhabitantToDocuments = make(map[int64][]string)

	for uuid, doc := range s.documents {
		participantIDs := doc.GetAllParticipantIDs()
		for _, id := range participantIDs {
			s.inhabitantToDocuments[id] = append(s.inhabitantToDocuments[id], uuid)
		}
	}
}

// addDocumentToInhabitantIndex adds a document to the inhabitant reverse index.
// Caller must hold the write lock.
func (s *Store) addDocumentToInhabitantIndex(doc *document.Document) {
	participantIDs := doc.GetAllParticipantIDs()
	for _, id := range participantIDs {
		s.inhabitantToDocuments[id] = append(s.inhabitantToDocuments[id], doc.UUID)
	}
}

// removeDocumentFromInhabitantIndex removes a document from the inhabitant reverse index.
// Caller must hold the write lock.
func (s *Store) removeDocumentFromInhabitantIndex(doc *document.Document) {
	participantIDs := doc.GetAllParticipantIDs()
	for _, id := range participantIDs {
		uuids := s.inhabitantToDocuments[id]
		filtered := make([]string, 0, len(uuids))
		for _, uuid := range uuids {
			if uuid != doc.UUID {
				filtered = append(filtered, uuid)
			}
		}
		if len(filtered) > 0 {
			s.inhabitantToDocuments[id] = filtered
		} else {
			delete(s.inhabitantToDocuments, id)
		}
	}
}
