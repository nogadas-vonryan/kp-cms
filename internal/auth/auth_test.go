package auth

import (
	"testing"
)

func TestUserStore_AddUserAndAuthenticate_Success(t *testing.T) {
	store := NewUserStore()
	username := "testuser"
	password := "testpass"
	role := RoleUser

	err := store.AddUser(username, password, role)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	user, err := store.Authenticate(username, password)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if user.Username != username {
		t.Errorf("expected username %s, got %s", username, user.Username)
	}
	if user.Role != role {
		t.Errorf("expected role %s, got %s", role, user.Role)
	}
}

func TestUserStore_Authenticate_InvalidPassword(t *testing.T) {
	store := NewUserStore()
	username := "user2"
	password := "pass2"
	store.AddUser(username, password, RoleUser)

	_, err := store.Authenticate(username, "wrongpass")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserStore_Authenticate_UserNotFound(t *testing.T) {
	store := NewUserStore()
	_, err := store.Authenticate("nouser", "nopass")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserStore_AddUser_Duplicate(t *testing.T) {
	store := NewUserStore()
	username := "dupuser"
	password := "pass"

	err := store.AddUser(username, password, RoleUser)
	if err != nil {
		t.Fatalf("first AddUser failed: %v", err)
	}

	err = store.AddUser(username, password, RoleUser)
	if err != ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestUserStore_UpdateRole(t *testing.T) {
	store := NewUserStore()
	username := "roleuser"
	store.AddUser(username, "pass", RoleUser)

	err := store.UpdateRole(username, RoleAdmin)
	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	user, _ := store.GetUser(username)
	if user.Role != RoleAdmin {
		t.Errorf("expected role %s, got %s", RoleAdmin, user.Role)
	}
}

func TestUserStore_DeleteUser(t *testing.T) {
	store := NewUserStore()
	username := "deluser"
	store.AddUser(username, "pass", RoleUser)

	err := store.DeleteUser(username)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = store.GetUser(username)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound after delete, got %v", err)
	}
}
