package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionMiddleware_Success(t *testing.T) {
	sessions := NewSessionManager(1 * time.Hour)
	username := "mwuser"
	identity := Identity{ID: username, Role: RoleUser}
	session := sessions.Create(identity)

	h := SessionMiddleware(sessions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(UserKey).(Identity)
		if !ok {
			t.Error("identity not found in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if id.ID != username {
			t.Errorf("expected ID %s, got %s", username, id.ID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if id.Role != RoleUser {
			t.Errorf("expected role %s, got %s", RoleUser, id.Role)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: session.Token})
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rw.Code)
	}
}

func TestSessionMiddleware_NoSession(t *testing.T) {
	sessions := NewSessionManager(1 * time.Hour)

	h := SessionMiddleware(sessions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rw.Code)
	}
}

func TestSessionMiddleware_ExpiredSession(t *testing.T) {
	sessions := NewSessionManager(1 * time.Millisecond)
	identity := Identity{ID: "user", Role: RoleUser}
	session := sessions.Create(identity)

	time.Sleep(10 * time.Millisecond)

	h := SessionMiddleware(sessions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: session.Token})
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for expired session, got %d", rw.Code)
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	sessions := NewSessionManager(1 * time.Hour)

	adminIdentity := Identity{ID: "admin", Role: RoleAdmin}
	adminSession := sessions.Create(adminIdentity)

	userIdentity := Identity{ID: "user", Role: RoleUser}
	userSession := sessions.Create(userIdentity)

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := SessionMiddleware(sessions)(RequireRole(RoleAdmin)(baseHandler))

	// Test admin access
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: adminSession.Token})
	rw := httptest.NewRecorder()

	mw.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin, got %d", rw.Code)
	}

	// Test user denied access
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(&http.Cookie{Name: SessionCookieName, Value: userSession.Token})
	rw2 := httptest.NewRecorder()
	mw.ServeHTTP(rw2, req2)
	if rw2.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-admin user, got %d", rw2.Code)
	}
}

func TestCSRFMiddleware_Success(t *testing.T) {
	csrfToken := "test_csrf_token"

	h := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rw.Code)
	}
}

func TestCSRFMiddleware_MissingToken(t *testing.T) {
	h := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", nil)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for missing CSRF token, got %d", rw.Code)
	}
}

func TestCSRFMiddleware_InvalidToken(t *testing.T) {
	h := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "cookie_token"})
	req.Header.Set("X-CSRF-Token", "header_token")
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for mismatched CSRF tokens, got %d", rw.Code)
	}
}

func TestCSRFMiddleware_SkipsGetRequests(t *testing.T) {
	h := CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 OK for GET request, got %d", rw.Code)
	}
}
