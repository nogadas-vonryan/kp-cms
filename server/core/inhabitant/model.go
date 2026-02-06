package inhabitant

import "time"

// Inhabitant represents a person in the system.
// For file-based storage, UUID, Code, FolderName, CreatedAt, and UpdatedAt are used.
// ID is retained for SQLite migration compatibility.
type Inhabitant struct {
	// New fields for file-based storage
	UUID       string    `json:"uuid"`
	Code       string    `json:"code"`        // e.g., "001-26"
	FolderName string    `json:"folder_name"` // e.g., "inhabitant-001-26"
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Existing fields (for display and compatibility)
	ID                           int64     `json:"id,omitempty"` // Deprecated: for migration only
	FirstName                    string    `json:"first_name"`
	LastName                     string    `json:"last_name"`
	MiddleName                   string    `json:"middle_name,omitempty"`
	Suffix                       string    `json:"suffix,omitempty"`
	Birthdate                    time.Time `json:"birthdate,omitempty"`
	BirthPlace                   string    `json:"birth_place,omitempty"`
	InhabitantType               string    `json:"inhabitant_type,omitempty"`
	Sex                          string    `json:"sex,omitempty"`
	CivilStatus                  string    `json:"civil_status,omitempty"`
	Citizenship                  string    `json:"citizenship,omitempty"`
	Occupation                   string    `json:"occupation,omitempty"`
	EmailAddress                 string    `json:"email_address,omitempty"`
	HighestEducationalAttainment string    `json:"highest_educational_attainment,omitempty"`
	MotherFirstName              string    `json:"mother_first_name,omitempty"`
	MotherMiddleName             string    `json:"mother_middle_name,omitempty"`
	MotherLastName               string    `json:"mother_last_name,omitempty"`
	ContactNo                    string    `json:"contact_no,omitempty"`
	Address                      string    `json:"address,omitempty"`
}
