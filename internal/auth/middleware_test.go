package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_Success(t *testing.T) {
	username := "mwuser"
	password := "mwpass"
	AddUser(username, password, RoleUser)

	h := AuthMiddleware("", "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := r.Context().Value(UserKey).(Identity)
		if !ok {
			t.Error("user not found in context")
		}
		if identity.ID != username {
			t.Errorf("expected ID %s, got %s", username, identity.ID)
		}
		if identity.Role != RoleUser {
			t.Errorf("expected role %s, got %s", RoleUser, identity.Role)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth(username, password)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rw.Code)
	}
}

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	h := AuthMiddleware("", "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rw.Code)
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	username := "adminuser"
	password := "adminpass"
	AddUser(username, password, RoleAdmin)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := AuthMiddleware("", "")(RequireRole(RoleAdmin)(baseHandler))

	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth(username, password)
	rw := httptest.NewRecorder()

	mw.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rw.Code)
	}

	// Now test forbidden for user role
	AddUser("useronly", "userpass", RoleUser)
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.SetBasicAuth("useronly", "userpass")
	rw2 := httptest.NewRecorder()
	mw.ServeHTTP(rw2, req2)
	if rw2.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", rw2.Code)
	}
}
