package auth

import (
	"testing"
)

func TestAddUserAndAuthenticate_Success(t *testing.T) {
	username := "testuser"
	password := "testpass"
	role := RoleUser

	err := AddUser(username, password, role)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	user, err := Authenticate(username, password)
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

func TestAuthenticate_InvalidPassword(t *testing.T) {
	username := "user2"
	password := "pass2"
	AddUser(username, password, RoleUser)

	_, err := Authenticate(username, "wrongpass")
	if err == nil {
		t.Error("expected error for invalid password, got nil")
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	_, err := Authenticate("nouser", "nopass")
	if err == nil {
		t.Error("expected error for user not found, got nil")
	}
}

func TestRequireRole(t *testing.T) {
	// This is a basic test for the RequireRole middleware logic.
	// Full HTTP middleware tests would require httptest, but we can check the function signature here.
	mw := RequireRole(RoleAdmin)
	if mw == nil {
		t.Error("RequireRole should return a middleware function")
	}
}
