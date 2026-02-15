package oauth

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Service provides OAuth management for all providers
type Service struct {
	store     *Store
	providers map[Provider]ProviderInterface
}

// NewService creates a new OAuth service
func NewService(db *sql.DB, config Config) (*Service, error) {
	store := NewStore(db)

	s := &Service{
		store:     store,
		providers: make(map[Provider]ProviderInterface),
	}

	// Initialize Google provider if credentials exist
	if config.Google.CredentialsPath != "" || defaultCredentialsExist() {
		googleProvider, err := NewGoogleProvider(config.Google.CredentialsPath, config.Google.CallbackURL)
		if err != nil {
			slog.Warn("Failed to initialize Google OAuth provider", "error", err)
		} else {
			s.providers[ProviderGoogle] = googleProvider
			slog.Info("Google OAuth provider initialized")
		}
	}

	// Start cleanup routine
	go s.startCleanupRoutine(context.Background())

	return s, nil
}

// IsProviderAvailable checks if a provider is configured
func (s *Service) IsProviderAvailable(provider Provider) bool {
	_, ok := s.providers[provider]
	return ok
}

// GetAuthURL generates an OAuth URL for a provider
func (s *Service) GetAuthURL(username string, provider Provider, scopes []string) (string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", fmt.Errorf("provider %s not available", provider)
	}

	userID, err := s.getUserID(username)
	if err != nil {
		return "", err
	}

	state := GenerateState()
	oauthState := &State{
		State:     state,
		UserID:    userID,
		Provider:  provider,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.store.SaveState(oauthState); err != nil {
		return "", fmt.Errorf("failed to save state: %w", err)
	}

	return p.GetAuthURL(state, scopes), nil
}

// HandleCallback processes an OAuth callback
func (s *Service) HandleCallback(ctx context.Context, state, code string) error {
	// Get and validate state
	oauthState, err := s.store.GetState(state)
	if err != nil {
		return err
	}

	// Get provider
	p, ok := s.providers[oauthState.Provider]
	if !ok {
		return fmt.Errorf("provider %s not available", oauthState.Provider)
	}

	// Exchange code for token
	oauthToken, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Save token
	token := FromOAuth2Token(oauthToken, oauthState.UserID, oauthState.Provider, oauthState.Scopes)
	if err := s.store.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Clean up state
	_ = s.store.DeleteState(state)

	slog.Info("OAuth connection established",
		"username", oauthState.Username,
		"provider", oauthState.Provider,
		"scopes", oauthState.Scopes)

	return nil
}

// IsConnected checks if a user has connected a provider
func (s *Service) IsConnected(username string, provider Provider) bool {
	return s.store.HasTokenByUsername(username, provider)
}

// GetConnections returns all OAuth connections for a user
func (s *Service) GetConnections(username string) ([]*Connection, error) {
	userID, err := s.getUserID(username)
	if err != nil {
		return nil, err
	}
	return s.store.GetConnections(userID)
}

// Disconnect removes an OAuth connection
func (s *Service) Disconnect(username string, provider Provider) error {
	return s.store.DeleteTokenByUsername(username, provider)
}

// GetHTTPClient returns an HTTP client for a user and provider
func (s *Service) GetHTTPClient(ctx context.Context, username string, provider Provider) (*http.Client, error) {
	// Get token
	token, err := s.store.GetTokenByUsername(username, provider)
	if err != nil {
		return nil, err
	}

	// Check if token needs refresh
	oauthToken := s.store.ToOAuth2Token(token)
	if oauthToken.Expiry.Before(time.Now()) {
		p, ok := s.providers[provider]
		if !ok {
			return nil, fmt.Errorf("provider %s not available", provider)
		}

		newToken, err := p.RefreshToken(ctx, oauthToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Save refreshed token
		refreshedToken := FromOAuth2Token(newToken, token.UserID, provider, token.Scopes)
		if err := s.store.SaveToken(refreshedToken); err != nil {
			slog.Error("Failed to save refreshed token", "error", err)
		}

		oauthToken = newToken
	}

	// Get provider and return client
	p, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not available", provider)
	}

	return p.GetHTTPClient(ctx, oauthToken), nil
}

// HasScope checks if a user has a specific scope
func (s *Service) HasScope(username string, provider Provider, scope string) bool {
	token, err := s.store.GetTokenByUsername(username, provider)
	if err != nil {
		return false
	}

	for _, s := range token.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// startCleanupRoutine periodically cleans up expired OAuth states
func (s *Service) startCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.store.CleanupExpiredStates(); err != nil {
				slog.Error("Failed to cleanup expired OAuth states", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// getUserID gets the numeric user ID from username
func (s *Service) getUserID(username string) (int64, error) {
	// This is a bit of a hack - we need to access the store's private method
	// In production, this should be part of a user service
	return s.store.getUserIDByUsername(username)
}

// defaultCredentialsExist checks if default credentials.json exists
func defaultCredentialsExist() bool {
	path := getDefaultCredentialsPath()
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
