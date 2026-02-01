package database

import (
	"context"
	"testing"
	"time"

	"kpcms/server/core/auth"
)

func setupTestDB(t *testing.T) *Database {
	db, err := New("file:test_auth?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestCreateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.CreateUser(ctx, "testuser", "password123", auth.RoleUser)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.CreateUser(ctx, "testuser", "password123", auth.RoleUser)
	if err != nil {
		t.Fatalf("first CreateUser failed: %v", err)
	}

	err = db.CreateUser(ctx, "testuser", "password456", auth.RoleUser)
	if err != auth.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	err := db.CreateUser(ctx, username, password, auth.RoleAdmin)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	identity, err := db.Authenticate(ctx, username, password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if identity.ID != username {
		t.Errorf("expected ID %q, got %q", username, identity.ID)
	}

	if identity.Role != auth.RoleAdmin {
		t.Errorf("expected role %q, got %q", auth.RoleAdmin, identity.Role)
	}
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	err := db.CreateUser(ctx, username, password, auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, err = db.Authenticate(ctx, username, "wrongpassword")
	if err != auth.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.Authenticate(ctx, "nonexistent", "password")
	if err != auth.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateSession_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	err := db.CreateUser(ctx, username, password, auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	identity := auth.Identity{ID: username, Role: auth.RoleUser}
	token, err := db.CreateSession(ctx, identity, 24*time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Errorf("expected non-empty token")
	}
}

func TestGetSession_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	err := db.CreateUser(ctx, username, password, auth.RoleAdmin)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	identity := auth.Identity{ID: username, Role: auth.RoleAdmin}
	token, err := db.CreateSession(ctx, identity, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	retrieved, err := db.GetSession(ctx, token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.ID != username {
		t.Errorf("expected ID %q, got %q", username, retrieved.ID)
	}

	if retrieved.Role != auth.RoleAdmin {
		t.Errorf("expected role %q, got %q", auth.RoleAdmin, retrieved.Role)
	}
}

func TestGetSession_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.GetSession(ctx, "nonexistent_token")
	if err != auth.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestGetSession_Expired(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	err := db.CreateUser(ctx, username, password, auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	identity := auth.Identity{ID: username, Role: auth.RoleUser}
	token, err := db.CreateSession(ctx, identity, -1*time.Second) // Already expired
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	_, err = db.GetSession(ctx, token)
	if err != auth.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for expired session, got %v", err)
	}
}

func TestDeleteSession_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	err := db.CreateUser(ctx, username, password, auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	identity := auth.Identity{ID: username, Role: auth.RoleUser}
	token, err := db.CreateSession(ctx, identity, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = db.DeleteSession(ctx, token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = db.GetSession(ctx, token)
	if err != auth.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials after delete, got %v", err)
	}
}

func TestListUsers_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	users := []struct {
		username string
		role     auth.Role
	}{
		{"user1", auth.RoleUser},
		{"user2", auth.RoleAdmin},
		{"user3", auth.RoleUser},
	}

	for _, u := range users {
		err := db.CreateUser(ctx, u.username, "password", u.role)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	summaries, err := db.ListUsers(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(summaries) != len(users) {
		t.Errorf("expected %d users, got %d", len(users), len(summaries))
	}

	for i, u := range users {
		if i < len(summaries) {
			if summaries[i].Username != u.username {
				t.Errorf("expected username %q, got %q", u.username, summaries[i].Username)
			}
			if summaries[i].Role != u.role {
				t.Errorf("expected role %q, got %q", u.role, summaries[i].Role)
			}
		}
	}
}

func TestUpdateRole_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	err := db.CreateUser(ctx, username, "password", auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = db.UpdateRole(ctx, username, auth.RoleAdmin)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	identity, err := db.Authenticate(ctx, username, "password")
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	if identity.Role != auth.RoleAdmin {
		t.Errorf("expected role %q, got %q", auth.RoleAdmin, identity.Role)
	}
}

func TestUpdateRole_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.UpdateRole(ctx, "nonexistent", auth.RoleAdmin)
	if err != auth.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	err := db.CreateUser(ctx, username, "password", auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = db.DeleteUser(ctx, username)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = db.Authenticate(ctx, username, "password")
	if err != auth.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound after delete, got %v", err)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.DeleteUser(ctx, "nonexistent")
	if err != auth.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestSetPassword_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	username := "testuser"
	oldPassword := "oldpassword"
	newPassword := "newpassword"

	err := db.CreateUser(ctx, username, oldPassword, auth.RoleUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = db.SetPassword(ctx, username, newPassword)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Old password should not work
	_, err = db.Authenticate(ctx, username, oldPassword)
	if err != auth.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials with old password, got %v", err)
	}

	// New password should work
	_, err = db.Authenticate(ctx, username, newPassword)
	if err != nil {
		t.Fatalf("expected no error with new password, got %v", err)
	}
}

func TestSetPassword_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.SetPassword(ctx, "nonexistent", "password")
	if err != auth.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
