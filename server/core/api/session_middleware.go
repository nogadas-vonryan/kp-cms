package api

import (
	"context"
	"net/http"

	"kpcms/server/core/auth"
)

// SessionMiddleware validates a session cookie and injects identity into the request context.
func (s *Server) SessionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(auth.SessionCookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			identity, err := s.authService.GetSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), auth.UserKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
