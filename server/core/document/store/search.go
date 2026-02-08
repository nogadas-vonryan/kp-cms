package store

import (
	"context"
	"kpcms/server/core/document"
	"sort"
	"strings"
)

func (r *Store) Search(ctx context.Context, criteria document.SearchCriteria) ([]*document.Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*document.Document

	// Iterate through cache
	for _, doc := range r.documents {
		if !matchesCriteria(doc, criteria) {
			continue
		}

		// Load files for matching documents
		files, err := r.readFiles(doc.FolderName)
		if err != nil {
			continue
		}

		result := &document.Document{
			UUID:       doc.UUID,
			Code:       doc.Code,
			FolderName: doc.FolderName,
			Title:      doc.Title,
			Fields:     doc.Fields,
			Files:      files,
			CreatedAt:  doc.CreatedAt,
			UpdatedAt:  doc.UpdatedAt,
		}
		results = append(results, result)
	}

	// Apply sorting
	sortResults(results, criteria.SortBy, criteria.SortDesc)

	// Apply pagination
	return paginate(results, criteria.Offset, criteria.Limit), nil
}

func matchesCriteria(doc *document.Document, criteria document.SearchCriteria) bool {
	// UUID filter (exact match)
	if criteria.UUID != "" && doc.UUID != criteria.UUID {
		return false
	}

	if criteria.Title != "" {
		if !strings.Contains(strings.ToLower(doc.Title), strings.ToLower(criteria.Title)) {
			return false
		}
	}

	// Code filter (prefix match)
	if criteria.Code != "" && !strings.HasPrefix(doc.Code, criteria.Code) {
		return false
	}

	// FolderName filter (case-insensitive substring match)
	if criteria.FolderName != "" {
		if !strings.Contains(
			strings.ToLower(doc.FolderName),
			strings.ToLower(criteria.FolderName),
		) {
			return false
		}
	}

	// Date range filter
	if criteria.DateFrom != nil && doc.CreatedAt.Before(*criteria.DateFrom) {
		return false
	}
	if criteria.DateTo != nil && doc.CreatedAt.After(*criteria.DateTo) {
		return false
	}

	// Field key existence check
	if criteria.FieldKey != "" {
		if _, exists := doc.Fields[criteria.FieldKey]; !exists {
			return false
		}
	}

	// Field key-value filters
	for key, expectedValue := range criteria.FieldFilters {
		actualValue, exists := doc.Fields[key]
		if !exists {
			return false
		}

		// Deep search in arrays (e.g., complainants)
		if !matchesFieldValue(actualValue, expectedValue) {
			return false
		}
	}

	return true
}

func matchesFieldValue(actual, expected any) bool {
	// Direct equality check
	if actual == expected {
		return true
	}

	// String contains (case-insensitive) with tokenized matching
	if actualStr, ok := actual.(string); ok {
		if expectedStr, ok := expected.(string); ok {
			return tokenizedMatch(actualStr, expectedStr)
		}
	}

	// Array search (for fields like complainants: [])
	if actualSlice, ok := actual.([]any); ok {
		for _, item := range actualSlice {
			if matchesFieldValue(item, expected) {
				return true
			}
		}
	}

	// Map/object search (for nested structures)
	if actualMap, ok := actual.(map[string]any); ok {
		for _, value := range actualMap {
			if matchesFieldValue(value, expected) {
				return true
			}
		}
	}

	return false
}

// tokenizedMatch checks if all tokens (words) in the search term exist in the target string.
// This allows "John Doe" to match "John Dabba Doe" or "Doe, John".
func tokenizedMatch(actual, expected string) bool {
	// Normalize strings: lowercase and trim
	actualLower := strings.ToLower(strings.TrimSpace(actual))
	expectedLower := strings.ToLower(strings.TrimSpace(expected))

	// Simple substring check first (optimization for exact matches)
	if strings.Contains(actualLower, expectedLower) {
		return true
	}

	// Tokenize the search term by splitting on whitespace and commas
	tokens := strings.FieldsFunc(expectedLower, func(r rune) bool {
		return r == ' ' || r == ','
	})

	// Empty search matches everything
	if len(tokens) == 0 {
		return true
	}

	// Check if all tokens exist in the actual string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if !strings.Contains(actualLower, token) {
			return false
		}
	}

	return true
}

func sortResults(docs []*document.Document, sortBy string, sortDesc bool) {
	if sortBy == "" {
		sortBy = "code" // default sort
	}

	sort.SliceStable(docs, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "created_at":
			less = docs[i].CreatedAt.Before(docs[j].CreatedAt)
		case "updated_at":
			less = docs[i].UpdatedAt.Before(docs[j].UpdatedAt)
		case "title":
			less = strings.ToLower(docs[i].Title) < strings.ToLower(docs[j].Title)
		case "folder_name":
			less = strings.ToLower(docs[i].FolderName) < strings.ToLower(docs[j].FolderName)
		case "code":
			fallthrough
		default:
			less = docs[i].Code < docs[j].Code
		}

		if sortDesc {
			return !less
		}
		return less
	})
}

func paginate(docs []*document.Document, offset, limit int) []*document.Document {
	if offset >= len(docs) || offset < 0 {
		return []*document.Document{}
	}
	end := offset + limit
	if end > len(docs) || limit <= 0 {
		end = len(docs)
	}
	return docs[offset:end]
}

// FindByInhabitantID returns all documents linked to the given inhabitant ID.
func (r *Store) FindByInhabitantID(ctx context.Context, inhabitantID int64) ([]*document.Document, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	uuids, exists := r.inhabitantToDocuments[inhabitantID]
	if !exists || len(uuids) == 0 {
		return []*document.Document{}, nil
	}

	results := make([]*document.Document, 0, len(uuids))
	for _, uuid := range uuids {
		doc, exists := r.documents[uuid]
		if !exists {
			continue
		}

		files, err := r.readFiles(doc.FolderName)
		if err != nil {
			continue
		}

		result := &document.Document{
			UUID:       doc.UUID,
			Code:       doc.Code,
			FolderName: doc.FolderName,
			Title:      doc.Title,
			Fields:     doc.Fields,
			Files:      files,
			CreatedAt:  doc.CreatedAt,
			UpdatedAt:  doc.UpdatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}
