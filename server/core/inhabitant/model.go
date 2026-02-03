package inhabitant

import "time"

type Inhabitant struct {
	ID                           int64
	FirstName                    string
	LastName                     string
	MiddleName                   string
	Suffix                       string
	Birthdate                    time.Time
	BirthPlace                   string
	InhabitantType               string
	Sex                          string
	CivilStatus                  string
	Citizenship                  string
	Occupation                   string
	EmailAddress                 string
	HighestEducationalAttainment string
	MotherFirstName              string
	MotherMiddleName             string
	MotherLastName               string
	ContactNo                    string
	Address                      string
}
