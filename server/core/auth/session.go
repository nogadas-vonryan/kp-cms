package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	SessionCookieName = "session_id"
	CSRFCookieName    = "csrf_token"
)

type Session struct {
	Token     string
	Identity  Identity
	ExpiresAt time.Time
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]Session
	ttl      time.Duration
}

func NewSessionManager(ttl time.Duration) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]Session),
		ttl:      ttl,
	}
}

func (m *SessionManager) Create(identity Identity) Session {
	token := randomToken(32)
	session := Session{
		Token:     token,
		Identity:  identity,
		ExpiresAt: time.Now().Add(m.ttl),
	}

	m.mu.Lock()
	m.sessions[token] = session
	m.mu.Unlock()

	return session
}

func (m *SessionManager) Get(token string) (Identity, bool) {
	m.mu.RLock()
	session, ok := m.sessions[token]
	m.mu.RUnlock()

	if !ok {
		return Identity{}, false
	}

	if time.Now().After(session.ExpiresAt) {
		m.Delete(token)
		return Identity{}, false
	}

	return session.Identity, true
}

func (m *SessionManager) Delete(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

func (m *SessionManager) TTL() time.Duration {
	return m.ttl
}

func randomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		fallback := base64.RawURLEncoding.EncodeToString([]byte(time.Now().String()))
		if len(fallback) >= length {
			return fallback[:length]
		}
		return fallback
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

// NewSessionToken generates a new session token with the specified length.
func NewSessionToken(length int) string {
	return randomToken(length)
}

func NewCSRFToken(length int) string {
	return randomToken(length)
}

func NewSessionCookie(token string, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	}
}

func NewCSRFCookie(token string, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	}
}
