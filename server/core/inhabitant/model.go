package inhabitant

import "time"

type Inhabitant struct {
	ID         int64
	FirstName  string
	LastName   string
	MiddleName string
	Suffix     string
	Birthday   time.Time
	ContactNo  string
	Address    string
}
