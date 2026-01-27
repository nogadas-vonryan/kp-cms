package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRole_AllowsCorrectRole(t *testing.T) {
	handler := RequireRole(RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	ctx = contextWithIdentity(ctx, Identity{ID: "admin", Role: RoleAdmin})
	req = req.WithContext(ctx)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rw.Code)
	}
}

func TestRequireRole_RejectsWrongRole(t *testing.T) {
	handler := RequireRole(RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	ctx = contextWithIdentity(ctx, Identity{ID: "user", Role: RoleUser})
	req = req.WithContext(ctx)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
	if rw.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", rw.Code)
	}
}

// Helper to set Identity in context
func contextWithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, UserKey, identity)
}
