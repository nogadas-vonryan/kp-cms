package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"kpcms/server/core/auth"

	"github.com/go-chi/chi/v5"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createUserRequest struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	Role     auth.Role `json:"role"`
}

type updateRoleRequest struct {
	Role auth.Role `json:"role"`
}

func (s *Server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		identity, err := s.authService.Authenticate(r.Context(), req.Username, req.Password)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		if err := issueSession(r.Context(), w, s.authService, identity, s.sessionTTL); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
	}
}

func (s *Server) handleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Username == "" || req.Password == "" {
			respondError(w, http.StatusBadRequest, "username and password are required")
			return
		}

		if err := s.authService.CreateUser(r.Context(), req.Username, req.Password, auth.RoleUser); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, auth.ErrUserExists) {
				status = http.StatusConflict
			}
			respondError(w, status, err.Error())
			return
		}

		identity := auth.Identity{ID: req.Username, Role: auth.RoleUser}
		if err := issueSession(r.Context(), w, s.authService, identity, s.sessionTTL); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
			_ = s.authService.DeleteSession(r.Context(), cookie.Value)
		}
		clearAuthCookies(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := r.Context().Value(auth.UserKey).(auth.Identity)
		if !ok {
			respondError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"user": identity})
	}
}

func (s *Server) handleListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.authService.ListUsers(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list users")
			return
		}
		respondJSON(w, http.StatusOK, users)
	}
}

func (s *Server) handleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Username == "" || req.Password == "" || req.Role == "" {
			respondError(w, http.StatusBadRequest, "username, password, and role are required")
			return
		}

		if !isValidRole(req.Role) {
			respondError(w, http.StatusBadRequest, "invalid role")
			return
		}

		if err := s.authService.CreateUser(r.Context(), req.Username, req.Password, req.Role); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, auth.ErrUserExists) {
				status = http.StatusConflict
			}
			respondError(w, status, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, auth.UserSummary{Username: req.Username, Role: req.Role})
	}
}

func (s *Server) handleUpdateUserRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		if username == "" {
			respondError(w, http.StatusBadRequest, "username is required")
			return
		}

		var req updateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Role == "" {
			respondError(w, http.StatusBadRequest, "role is required")
			return
		}

		if !isValidRole(req.Role) {
			respondError(w, http.StatusBadRequest, "invalid role")
			return
		}

		if err := s.authService.UpdateRole(r.Context(), username, req.Role); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, auth.ErrUserNotFound) {
				status = http.StatusNotFound
			}
			respondError(w, status, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, auth.UserSummary{Username: username, Role: req.Role})
	}
}

func (s *Server) handleDeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		if username == "" {
			respondError(w, http.StatusBadRequest, "username is required")
			return
		}

		identity, _ := r.Context().Value(auth.UserKey).(auth.Identity)
		if identity.ID == username {
			respondError(w, http.StatusBadRequest, "cannot delete current authenticated user")
			return
		}

		if err := s.authService.DeleteUser(r.Context(), username); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, auth.ErrUserNotFound) {
				status = http.StatusNotFound
			}
			respondError(w, status, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func issueSession(ctx context.Context, w http.ResponseWriter, authSvc *auth.Service, identity auth.Identity, ttl time.Duration) error {
	token, err := authSvc.CreateSession(ctx, identity, ttl)
	if err != nil {
		return err
	}
	csrfToken := auth.NewCSRFToken(32)

	http.SetCookie(w, auth.NewSessionCookie(token, ttl))
	http.SetCookie(w, auth.NewCSRFCookie(csrfToken, ttl))

	respondJSON(w, http.StatusOK, map[string]any{
		"user":       identity,
		"csrf_token": csrfToken,
	})

	return nil
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: auth.SessionCookieName, Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: auth.CSRFCookieName, Value: "", Path: "/", MaxAge: -1})
}

func isValidRole(role auth.Role) bool {
	return role == auth.RoleAdmin || role == auth.RoleUser
}
