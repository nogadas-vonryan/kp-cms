package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleProvider implements OAuth for Google
type GoogleProvider struct {
	config *oauth2.Config
}

// GoogleCredentials represents the structure of credentials.json
type GoogleCredentials struct {
	Installed struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
	} `json:"installed"`
	Web struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
	} `json:"web"`
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(credentialsPath, callbackURL string) (*GoogleProvider, error) {
	if credentialsPath == "" {
		credentialsPath = getDefaultCredentialsPath()
	}

	config, err := loadGoogleCredentials(credentialsPath)
	if err != nil {
		return nil, err
	}

	if callbackURL != "" {
		config.RedirectURL = callbackURL
	}

	return &GoogleProvider{
		config: config,
	}, nil
}

// GetAuthURL returns the OAuth URL for Google
func (p *GoogleProvider) GetAuthURL(state string, scopes []string) string {
	// Ensure we have at least userinfo scope
	hasUserInfo := false
	for _, scope := range scopes {
		if scope == ScopeUserInfo {
			hasUserInfo = true
			break
		}
	}
	if !hasUserInfo {
		scopes = append([]string{ScopeUserInfo}, scopes...)
	}

	// Update config with requested scopes
	config := &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		RedirectURL:  p.config.RedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges an authorization code for a token
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// RefreshToken refreshes an expired token
func (p *GoogleProvider) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := p.config.TokenSource(ctx, token)
	return tokenSource.Token()
}

// GetHTTPClient returns an HTTP client authorized with the token
func (p *GoogleProvider) GetHTTPClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return p.config.Client(ctx, token)
}

// GetName returns the provider name
func (p *GoogleProvider) GetName() Provider {
	return ProviderGoogle
}

// HasScope checks if a scope is supported
func (p *GoogleProvider) HasScope(scope string) bool {
	supportedScopes := []string{
		ScopeCalendar,
		ScopeCalendarEvents,
		ScopeDrive,
		ScopeDriveFile,
		ScopeGmailSend,
		ScopeUserInfo,
	}
	for _, s := range supportedScopes {
		if s == scope {
			return true
		}
	}
	return false
}

// GenerateState generates a random state string for CSRF protection
func GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// loadGoogleCredentials loads OAuth2 credentials from a JSON file
func loadGoogleCredentials(path string) (*oauth2.Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials.json not found at: %s\n\n"+
			"Please create a Google Cloud OAuth2 credential and place credentials.json in the same directory as the executable.\n"+
			"See: https://developers.google.com/calendar/api/quickstart/go", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds GoogleCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials.json format: %w", err)
	}

	var clientID, clientSecret string
	var redirectURIs []string

	if creds.Installed.ClientID != "" {
		clientID = creds.Installed.ClientID
		clientSecret = creds.Installed.ClientSecret
		redirectURIs = creds.Installed.RedirectURIs
	} else if creds.Web.ClientID != "" {
		clientID = creds.Web.ClientID
		clientSecret = creds.Web.ClientSecret
		redirectURIs = creds.Web.RedirectURIs
	} else {
		return nil, fmt.Errorf("credentials.json missing client_id")
	}

	redirectURI := "http://localhost:8080/oauth/callback"
	if len(redirectURIs) > 0 {
		redirectURI = redirectURIs[0]
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{ScopeUserInfo},
		Endpoint:     google.Endpoint,
	}, nil
}

// getDefaultCredentialsPath returns the default path for credentials.json
func getDefaultCredentialsPath() string {
	// Look in the platform-specific config directory
	configDir, err := getConfigDir()
	if err == nil {
		configPath := filepath.Join(configDir, "credentials.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Fallback to current working directory
	return "./credentials.json"
}

// getConfigDir returns the platform-specific config directory
func getConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "kpcms")
	case "linux", "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		configDir = filepath.Join(homeDir, ".config", "kpcms")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return configDir, nil
}
