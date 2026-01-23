package document

import "time"

// This is saved on meta.json on each document's folder
type Document struct {
	UUID       string         `json:"uuid"`
	Code       string         `json:"code"`
	FolderName string         `json:"folder_name"`
	Title      string         `json:"title"`
	Fields     map[string]any `json:"fields"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DocumentResponse struct {
	UUID       string         `json:"uuid"`
	Code       string         `json:"code"`
	FolderName string         `json:"folder_name"`
	Title      string         `json:"title"`
	Fields     map[string]any `json:"fields"`
	Files      []File         `json:"files"`

	CreatedAt time.Time `json:"created_at"`
}

type File struct {
	FileName    string    `json:"file_name"`
	Description string    `json:"description"`
	Note        string    `json:"note"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}
