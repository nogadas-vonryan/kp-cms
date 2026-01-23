package auth

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Identity struct {
	ID   string
	Role Role
}
