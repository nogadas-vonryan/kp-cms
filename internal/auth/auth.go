package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
	Role     Role
}

// Global user store for now
var Users = make(map[string]User)

func AddUser(username, password string, role Role) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	Users[username] = User{
		Username: username,
		Password: string(hash),
		Role:     role,
	}

	return nil
}

func Authenticate(username, password string) (User, error) {
	user, exists := Users[username]
	if !exists {
		return User{}, errors.New("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, errors.New("invalid credentials")
	}

	return user, nil
}
