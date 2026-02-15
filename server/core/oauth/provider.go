package oauth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// ProviderInterface defines the interface for OAuth providers
type ProviderInterface interface {
	// GetAuthURL returns the OAuth URL for the given state and scopes
	GetAuthURL(state string, scopes []string) string

	// ExchangeCode exchanges an authorization code for a token
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)

	// RefreshToken refreshes an expired token
	RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)

	// GetHTTPClient returns an HTTP client authorized with the token
	GetHTTPClient(ctx context.Context, token *oauth2.Token) *http.Client

	// GetName returns the provider name
	GetName() Provider

	// HasScope checks if a scope is supported by this provider
	HasScope(scope string) bool
}
