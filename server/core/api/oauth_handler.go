package api

import (
	"encoding/json"
	"net/http"

	"kpcms/server/core/auth"
	"kpcms/server/core/oauth"

	"github.com/go-chi/chi/v5"
)

// OAuthHandler handles OAuth-related HTTP requests
type OAuthHandler struct {
	service *oauth.Service
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(service *oauth.Service) *OAuthHandler {
	return &OAuthHandler{service: service}
}

// handleGetAuthURL returns the OAuth URL for connecting a provider
func (h *OAuthHandler) handleGetAuthURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get provider and scopes from query parameters
		providerStr := r.URL.Query().Get("provider")
		if providerStr == "" {
			providerStr = string(oauth.ProviderGoogle)
		}
		provider := oauth.Provider(providerStr)

		// Get requested scopes or use defaults
		var scopes []string
		if scopeStr := r.URL.Query().Get("scopes"); scopeStr != "" {
			// Parse scopes from comma-separated list
			scopes = parseScopes(scopeStr)
		} else {
			// Use calendar scopes by default
			scopes = oauth.CalendarScopes()
		}

		authURL, err := h.service.GetAuthURL(username, provider, scopes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"auth_url": authURL,
		})
	}
}

// handleOAuthCallback handles the OAuth callback from providers
func (h *OAuthHandler) handleOAuthCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")
		errorParam := r.URL.Query().Get("error")

		if errorParam != "" {
			http.Error(w, "OAuth error: "+errorParam, http.StatusBadRequest)
			return
		}

		if state == "" || code == "" {
			http.Error(w, "Missing state or code parameter", http.StatusBadRequest)
			return
		}

		if err := h.service.HandleCallback(r.Context(), state, code); err != nil {
			if err == oauth.ErrInvalidState {
				http.Error(w, "Invalid or expired OAuth state", http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "OAuth connection established successfully",
		})
	}
}

// handleGetConnectionStatus returns the OAuth connection status
func (h *OAuthHandler) handleGetConnectionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get provider from query
		providerStr := r.URL.Query().Get("provider")
		if providerStr == "" {
			providerStr = string(oauth.ProviderGoogle)
		}
		provider := oauth.Provider(providerStr)

		connections, err := h.service.GetConnections(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		isConnected := h.service.IsConnected(username, provider)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"is_connected": isConnected,
			"connections":  connections,
		})
	}
}

// handleDisconnect disconnects an OAuth provider
func (h *OAuthHandler) handleDisconnect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get provider from query
		providerStr := r.URL.Query().Get("provider")
		if providerStr == "" {
			providerStr = string(oauth.ProviderGoogle)
		}
		provider := oauth.Provider(providerStr)

		if err := h.service.Disconnect(username, provider); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "OAuth connection disconnected successfully",
		})
	}
}

// parseScopes parses a comma-separated scope string
func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return nil
	}
	scopes := []string{}
	start := 0
	for i, ch := range scopeStr {
		if ch == ',' {
			if i > start {
				scopes = append(scopes, scopeStr[start:i])
			}
			start = i + 1
		}
	}
	if start < len(scopeStr) {
		scopes = append(scopes, scopeStr[start:])
	}
	return scopes
}

// RegisterOAuthRoutes registers OAuth routes with the router
func (s *Server) RegisterOAuthRoutes(handler *OAuthHandler) {
	// OAuth callback is public (no auth required)
	s.Router.Get("/oauth/callback", handler.handleOAuthCallback())

	// Protected OAuth routes
	s.Router.Route("/api/oauth", func(r chi.Router) {
		r.Use(s.SessionMiddleware())
		r.Use(auth.CSRFMiddleware())

		r.Get("/auth-url", handler.handleGetAuthURL())
		r.Get("/status", handler.handleGetConnectionStatus())
		r.Post("/disconnect", handler.handleDisconnect())
	})
}
