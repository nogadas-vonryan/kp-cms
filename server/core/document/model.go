package document

import "time"

// This is saved on meta.json on each document's folder
type Document struct {
	UUID       string         `json:"uuid"`
	Code       string         `json:"code"`
	FolderName string         `json:"folder_name"`
	Title      string         `json:"title"`
	Fields     map[string]any `json:"fields"`
	Files      []File         `json:"files"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type File struct {
	FileName    string    `json:"file_name"`
	Description string    `json:"description"`
	Note        string    `json:"note"`
	Tags        []string  `json:"tags"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

// FileMetadataUpdate carries optional updates to a file's metadata.
// A nil field means "leave as-is"; non-nil values overwrite existing data.
type FileMetadataUpdate struct {
	Description *string   `json:"description,omitempty"`
	Note        *string   `json:"note,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
}

type SearchCriteria struct {
	UUID         string         // exact match for UUID
	Title        string         // match for title
	Code         string         // prefix match for code
	FolderName   string         // substring match for folder name (case-insensitive)
	FieldKey     string         // search for presence of a field
	FieldFilters map[string]any // search for field key-value pairs
	DateFrom     *time.Time     // filter documents created after this date
	DateTo       *time.Time     // filter documents created before this date
	SortBy       string         // sort by: code, created_at, updated_at, title, folder_name
	SortDesc     bool           // sort descending when true
	Offset       int            // pagination offset
	Limit        int            // pagination limit
}

type SyncIssue struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type BackupFile struct {
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// GetComplainantIDs extracts complainant IDs from the document fields.
// Returns an empty slice if the field is missing or has wrong type.
func (d *Document) GetComplainantIDs() []int64 {
	if d.Fields == nil {
		return []int64{}
	}
	raw, ok := d.Fields["complainant_ids"]
	if !ok {
		return []int64{}
	}
	return parseIDArray(raw)
}

// GetRespondentIDs extracts respondent IDs from the document fields.
// Returns an empty slice if the field is missing or has wrong type.
func (d *Document) GetRespondentIDs() []int64 {
	if d.Fields == nil {
		return []int64{}
	}
	raw, ok := d.Fields["respondent_ids"]
	if !ok {
		return []int64{}
	}
	return parseIDArray(raw)
}

// GetAllParticipantIDs returns all unique participant IDs (complainants + respondents).
func (d *Document) GetAllParticipantIDs() []int64 {
	complainants := d.GetComplainantIDs()
	respondents := d.GetRespondentIDs()

	seen := make(map[int64]bool)
	result := make([]int64, 0, len(complainants)+len(respondents))

	for _, id := range complainants {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	for _, id := range respondents {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}

// parseIDArray converts various numeric types to []int64.
// Handles []any containing int, int64, float64, etc.
func parseIDArray(raw any) []int64 {
	switch v := raw.(type) {
	case []int64:
		return v
	case []any:
		result := make([]int64, 0, len(v))
		for _, item := range v {
			if id := toInt64(item); id != 0 {
				result = append(result, id)
			}
		}
		return result
	case []int:
		result := make([]int64, 0, len(v))
		for _, id := range v {
			result = append(result, int64(id))
		}
		return result
	case []float64:
		result := make([]int64, 0, len(v))
		for _, id := range v {
			result = append(result, int64(id))
		}
		return result
	default:
		return []int64{}
	}
}

// toInt64 converts a numeric value to int64, returns 0 if conversion fails.
func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	default:
		return 0
	}
}
