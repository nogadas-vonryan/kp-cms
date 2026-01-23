package auth

import (
	"context"
	"net/http"
)

type contextKey string

const UserKey contextKey = "user"

func AuthMiddleware(expectedUser, expectedPass string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()

			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := Authenticate(username, password)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, Identity{
				ID:   user.Username,
				Role: user.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
