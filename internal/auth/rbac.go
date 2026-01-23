package auth

import "net/http"

func RequireRole(requiredRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := r.Context().Value(UserKey).(Identity)
			if !ok || identity.Role != requiredRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
