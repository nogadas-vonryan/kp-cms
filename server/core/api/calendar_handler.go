package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"kpcms/server/core/auth"
	"kpcms/server/core/calendar"
	"kpcms/server/core/oauth"

	"github.com/go-chi/chi/v5"
)

// getUsernameFromContext extracts the username from the request context
func getUsernameFromContext(r *http.Request) (string, bool) {
	identity, ok := r.Context().Value(auth.UserKey).(auth.Identity)
	if !ok {
		return "", false
	}
	return identity.ID, true
}

// parseInt64 parses a string to int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// CalendarHandler handles calendar-related HTTP requests
type CalendarHandler struct {
	service *calendar.Service
}

// NewCalendarHandler creates a new calendar handler
func NewCalendarHandler(service *calendar.Service) *CalendarHandler {
	return &CalendarHandler{service: service}
}

// handleGetConnectionStatus returns the calendar connection status
func (h *CalendarHandler) handleGetConnectionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		connection, err := h.service.GetConnectionStatus(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"is_connected": connection.IsConnected,
			"connection":   connection,
		})
	}
}

// handleCreateEvent creates a new calendar event
func (h *CalendarHandler) handleCreateEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req calendar.CreateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Summary == "" {
			http.Error(w, "Summary is required", http.StatusBadRequest)
			return
		}

		if req.StartTime.IsZero() || req.EndTime.IsZero() {
			http.Error(w, "Start and end times are required", http.StatusBadRequest)
			return
		}

		if req.EndTime.Before(req.StartTime) {
			http.Error(w, "End time must be after start time", http.StatusBadRequest)
			return
		}

		event, err := h.service.CreateEvent(r.Context(), username, &req)
		if err != nil {
			if errors.Is(err, oauth.ErrTokenNotFound) || errors.Is(err, calendar.ErrNotConnected) {
				http.Error(w, "Calendar not connected. Please connect your calendar first.", http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(event)
	}
}

// handleListEvents lists calendar events
func (h *CalendarHandler) handleListEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse query parameters
		opts := calendar.ListOptions{}

		if startStr := r.URL.Query().Get("start_time"); startStr != "" {
			t, err := time.Parse(time.RFC3339, startStr)
			if err != nil {
				http.Error(w, "Invalid start_time format. Use RFC3339.", http.StatusBadRequest)
				return
			}
			opts.StartTime = t
		}

		if endStr := r.URL.Query().Get("end_time"); endStr != "" {
			t, err := time.Parse(time.RFC3339, endStr)
			if err != nil {
				http.Error(w, "Invalid end_time format. Use RFC3339.", http.StatusBadRequest)
				return
			}
			opts.EndTime = t
		}

		// Validate that end_time is after start_time if both are provided
		if !opts.StartTime.IsZero() && !opts.EndTime.IsZero() && opts.EndTime.Before(opts.StartTime) {
			http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
			return
		}

		opts.CalendarID = r.URL.Query().Get("calendar_id")

		if maxStr := r.URL.Query().Get("max_results"); maxStr != "" {
			if max, err := parseInt64(maxStr); err == nil {
				opts.MaxResults = max
			}
		}

		if showDeleted := r.URL.Query().Get("show_deleted"); showDeleted == "true" {
			opts.ShowDeleted = true
		}

		response, err := h.service.ListEvents(r.Context(), username, opts)
		if err != nil {
			if errors.Is(err, oauth.ErrTokenNotFound) || errors.Is(err, calendar.ErrNotConnected) {
				http.Error(w, "Calendar not connected. Please connect your calendar first.", http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleListCalendars lists all available calendars
func (h *CalendarHandler) handleListCalendars() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := getUsernameFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		calendars, err := h.service.ListCalendars(r.Context(), username)
		if err != nil {
			if errors.Is(err, oauth.ErrTokenNotFound) || errors.Is(err, calendar.ErrNotConnected) {
				http.Error(w, "Calendar not connected. Please connect your calendar first.", http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"calendars": calendars,
		})
	}
}

// RegisterCalendarRoutes registers calendar routes with the router
func (s *Server) RegisterCalendarRoutes(handler *CalendarHandler) {
	// Protected calendar routes
	s.Router.Route("/api/calendar", func(r chi.Router) {
		r.Use(s.SessionMiddleware())
		r.Use(auth.CSRFMiddleware())

		r.Get("/status", handler.handleGetConnectionStatus())
		r.Get("/calendars", handler.handleListCalendars())

		r.Route("/events", func(events chi.Router) {
			events.Get("/", handler.handleListEvents())
			events.Post("/", handler.handleCreateEvent())
		})
	})
}
