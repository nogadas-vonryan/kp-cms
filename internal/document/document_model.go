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
	KPFormType  KPForm    `json:"kp_form_type"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

// FileMetadataUpdate carries optional updates to a file's metadata.
// A nil field means "leave as-is"; non-nil values overwrite existing data.
type FileMetadataUpdate struct {
	Description *string `json:"description,omitempty"`
	Note        *string `json:"note,omitempty"`
	KPFormType  *KPForm `json:"kp_form_type,omitempty"`
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

type KPForm uint8

func (k KPForm) IsValid() bool {
	return k <= 27
}
