package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

// handleAdvancedSearch performs a cross-domain search joining inhabitants and documents.
func (s *Server) handleAdvancedSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract query parameter
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		// Extract optional limit parameter
		limit := 20 // default limit
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "invalid limit parameter", http.StatusBadRequest)
				return
			}
			if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}

		// Execute the aggregated search
		result, err := s.searchService.ExecuteAdvancedSearch(ctx, query, limit)
		if err != nil {
			slog.Error("advanced search failed",
				"query", query,
				"error", err)
			http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			slog.Error("failed to encode search response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
