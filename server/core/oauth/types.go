package oauth

import (
	"time"
)

// Provider represents an OAuth provider (Google, Microsoft, etc.)
type Provider string

const (
	ProviderGoogle Provider = "google"
)

// Token represents an OAuth token stored in the database
type Token struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	Provider     Provider  `json:"provider"`
	AccessToken  string    `json:"-"` // Never expose in JSON
	RefreshToken string    `json:"-"` // Never expose in JSON
	TokenType    string    `json:"token_type"`
	Scopes       []string  `json:"scopes"`
	Expiry       time.Time `json:"expiry"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Connection represents a user's OAuth connection status
type Connection struct {
	Username    string    `json:"username"`
	Provider    string    `json:"provider"`
	Scopes      []string  `json:"scopes"`
	IsConnected bool      `json:"is_connected"`
	ConnectedAt time.Time `json:"connected_at"`
}

// Config represents OAuth configuration
type Config struct {
	Google GoogleConfig
}

// GoogleConfig represents Google OAuth configuration
type GoogleConfig struct {
	CredentialsPath string
	CallbackURL     string
}

// State represents an OAuth state for CSRF protection
type State struct {
	State     string    `json:"state"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Provider  Provider  `json:"provider"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthURLRequest represents a request to generate an OAuth URL
type AuthURLRequest struct {
	Provider Provider `json:"provider"`
	Scopes   []string `json:"scopes"`
}

// CallbackRequest represents an OAuth callback
type CallbackRequest struct {
	Provider Provider `json:"provider"`
	State    string   `json:"state"`
	Code     string   `json:"code"`
}

// Scope constants for Google APIs
const (
	ScopeCalendar       = "https://www.googleapis.com/auth/calendar"
	ScopeCalendarEvents = "https://www.googleapis.com/auth/calendar.events"
	ScopeDrive          = "https://www.googleapis.com/auth/drive"
	ScopeDriveFile      = "https://www.googleapis.com/auth/drive.file"
	ScopeGmailSend      = "https://www.googleapis.com/auth/gmail.send"
	ScopeUserInfo       = "https://www.googleapis.com/auth/userinfo.email"
)

// DefaultScopes returns default scopes for a provider
func DefaultScopes(provider Provider) []string {
	switch provider {
	case ProviderGoogle:
		return []string{ScopeUserInfo}
	default:
		return nil
	}
}

// CalendarScopes returns scopes needed for calendar access
func CalendarScopes() []string {
	return []string{ScopeCalendar, ScopeCalendarEvents}
}

// DriveScopes returns scopes needed for drive access
func DriveScopes() []string {
	return []string{ScopeDrive, ScopeDriveFile}
}
