package oauth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

// Store handles database operations for OAuth tokens
type Store struct {
	db *sql.DB
}

// NewStore creates a new OAuth store
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// SaveToken saves or updates an OAuth token
func (s *Store) SaveToken(token *Token) error {
	scopesJSON, err := json.Marshal(token.Scopes)
	if err != nil {
		return fmt.Errorf("failed to marshal scopes: %w", err)
	}

	query := `
		INSERT INTO oauth_tokens (user_id, provider, access_token, refresh_token, token_type, scopes, expiry, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, provider) DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_type = excluded.token_type,
			scopes = excluded.scopes,
			expiry = excluded.expiry,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = s.db.Exec(query, token.UserID, token.Provider, token.AccessToken,
		token.RefreshToken, token.TokenType, string(scopesJSON), token.Expiry)
	return err
}

// GetToken retrieves an OAuth token for a user and provider
func (s *Store) GetToken(userID int64, provider Provider) (*Token, error) {
	var token Token
	var scopesJSON string

	query := `
		SELECT t.id, t.user_id, u.username, t.provider, t.access_token, t.refresh_token, 
		       t.token_type, t.scopes, t.expiry, t.created_at, t.updated_at
		FROM oauth_tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.user_id = ? AND t.provider = ?
	`
	err := s.db.QueryRow(query, userID, provider).Scan(
		&token.ID, &token.UserID, &token.Username, &token.Provider,
		&token.AccessToken, &token.RefreshToken, &token.TokenType,
		&scopesJSON, &token.Expiry, &token.CreatedAt, &token.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(scopesJSON), &token.Scopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
	}

	return &token, nil
}

// GetTokenByUsername retrieves an OAuth token by username and provider
func (s *Store) GetTokenByUsername(username string, provider Provider) (*Token, error) {
	userID, err := s.getUserIDByUsername(username)
	if err != nil {
		return nil, err
	}
	return s.GetToken(userID, provider)
}

// HasToken checks if a user has a token for a provider
func (s *Store) HasToken(userID int64, provider Provider) bool {
	var count int
	query := "SELECT COUNT(*) FROM oauth_tokens WHERE user_id = ? AND provider = ?"
	err := s.db.QueryRow(query, userID, provider).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// HasTokenByUsername checks if a user has a token by username
func (s *Store) HasTokenByUsername(username string, provider Provider) bool {
	userID, err := s.getUserIDByUsername(username)
	if err != nil {
		return false
	}
	return s.HasToken(userID, provider)
}

// DeleteToken removes an OAuth token
func (s *Store) DeleteToken(userID int64, provider Provider) error {
	query := "DELETE FROM oauth_tokens WHERE user_id = ? AND provider = ?"
	_, err := s.db.Exec(query, userID, provider)
	return err
}

// DeleteTokenByUsername removes a token by username
func (s *Store) DeleteTokenByUsername(username string, provider Provider) error {
	userID, err := s.getUserIDByUsername(username)
	if err != nil {
		return err
	}
	return s.DeleteToken(userID, provider)
}

// GetConnections returns all OAuth connections for a user
func (s *Store) GetConnections(userID int64) ([]*Connection, error) {
	query := `
		SELECT u.username, t.provider, t.scopes, t.created_at
		FROM oauth_tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.user_id = ?
		ORDER BY t.created_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*Connection
	for rows.Next() {
		var conn Connection
		var scopesJSON string
		var createdAt string
		err := rows.Scan(&conn.Username, &conn.Provider, &scopesJSON, &createdAt)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(scopesJSON), &conn.Scopes); err != nil {
			return nil, err
		}
		conn.IsConnected = true
		conn.ConnectedAt, _ = time.Parse(time.RFC3339, createdAt)
		connections = append(connections, &conn)
	}

	return connections, rows.Err()
}

// SaveState saves an OAuth state for CSRF protection
func (s *Store) SaveState(state *State) error {
	scopesJSON, err := json.Marshal(state.Scopes)
	if err != nil {
		return fmt.Errorf("failed to marshal scopes: %w", err)
	}

	query := `
		INSERT INTO oauth_states (state, user_id, provider, scopes, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err = s.db.Exec(query, state.State, state.UserID, state.Provider,
		string(scopesJSON), state.ExpiresAt)
	return err
}

// GetState retrieves and validates an OAuth state
func (s *Store) GetState(stateStr string) (*State, error) {
	var state State
	var scopesJSON string

	query := `
		SELECT s.state, s.user_id, u.username, s.provider, s.scopes, s.expires_at
		FROM oauth_states s
		JOIN users u ON u.id = s.user_id
		WHERE s.state = ? AND s.expires_at > datetime('now')
	`
	err := s.db.QueryRow(query, stateStr).Scan(
		&state.State, &state.UserID, &state.Username, &state.Provider,
		&scopesJSON, &state.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidState
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(scopesJSON), &state.Scopes); err != nil {
		return nil, err
	}

	return &state, nil
}

// DeleteState removes an OAuth state
func (s *Store) DeleteState(stateStr string) error {
	_, err := s.db.Exec("DELETE FROM oauth_states WHERE state = ?", stateStr)
	return err
}

// CleanupExpiredStates removes expired OAuth states
func (s *Store) CleanupExpiredStates() error {
	_, err := s.db.Exec("DELETE FROM oauth_states WHERE expires_at < datetime('now')")
	return err
}

// toOAuth2Token converts our Token to oauth2.Token
func (s *Store) ToOAuth2Token(token *Token) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
}

// FromOAuth2Token converts oauth2.Token to our Token
func FromOAuth2Token(oauthToken *oauth2.Token, userID int64, provider Provider, scopes []string) *Token {
	return &Token{
		UserID:       userID,
		Provider:     provider,
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
		TokenType:    oauthToken.TokenType,
		Scopes:       scopes,
		Expiry:       oauthToken.Expiry,
	}
}

// getUserIDByUsername gets the numeric user ID from username
func (s *Store) getUserIDByUsername(username string) (int64, error) {
	var userID int64
	err := s.db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("user not found: %s", username)
	}
	return userID, err
}

// Errors
var (
	ErrTokenNotFound = errors.New("oauth: token not found")
	ErrInvalidState  = errors.New("oauth: invalid or expired state")
)
