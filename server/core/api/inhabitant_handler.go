package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kpcms/server/core/inhabitant"

	"github.com/go-chi/chi/v5"
)

type createInhabitantRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Suffix     string `json:"suffix"`
	Birthday   string `json:"birthday"` // ISO 8601 date format
	ContactNo  string `json:"contact_no"`
	Address    string `json:"address"`
}

type inhabitantResponse struct {
	ID         int64  `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Suffix     string `json:"suffix"`
	Birthday   string `json:"birthday"`
	ContactNo  string `json:"contact_no"`
	Address    string `json:"address"`
}

func (s *Server) handleListInhabitants() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 100
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		inhabitants, err := s.inhabitantService.List(r.Context(), limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list inhabitants")
			return
		}

		response := make([]inhabitantResponse, len(inhabitants))
		for i, inh := range inhabitants {
			response[i] = inhabitantResponse{
				ID:         inh.ID,
				FirstName:  inh.FirstName,
				LastName:   inh.LastName,
				MiddleName: inh.MiddleName,
				Suffix:     inh.Suffix,
				Birthday:   inh.Birthday.Format("2006-01-02"),
				ContactNo:  inh.ContactNo,
				Address:    inh.Address,
			}
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func (s *Server) handleGetInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid inhabitant id")
			return
		}

		inhabitant, err := s.inhabitantService.Get(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "inhabitant not found")
			return
		}

		response := inhabitantResponse{
			ID:         inhabitant.ID,
			FirstName:  inhabitant.FirstName,
			LastName:   inhabitant.LastName,
			MiddleName: inhabitant.MiddleName,
			Suffix:     inhabitant.Suffix,
			Birthday:   inhabitant.Birthday.Format("2006-01-02"),
			ContactNo:  inhabitant.ContactNo,
			Address:    inhabitant.Address,
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func (s *Server) handleCreateInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createInhabitantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.FirstName == "" || req.LastName == "" {
			respondError(w, http.StatusBadRequest, "first_name and last_name are required")
			return
		}

		inhabitant := &inhabitant.Inhabitant{
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			MiddleName: req.MiddleName,
			Suffix:     req.Suffix,
			ContactNo:  req.ContactNo,
			Address:    req.Address,
		}

		// Parse birthday if provided
		if req.Birthday != "" {
			parsedDate, err := time.Parse("2006-01-02", req.Birthday)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid birthday format (use YYYY-MM-DD)")
				return
			}
			inhabitant.Birthday = parsedDate
		}

		id, err := s.inhabitantService.Create(r.Context(), inhabitant)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create inhabitant")
			return
		}

		inhabitant.ID = id
		response := inhabitantResponse{
			ID:         inhabitant.ID,
			FirstName:  inhabitant.FirstName,
			LastName:   inhabitant.LastName,
			MiddleName: inhabitant.MiddleName,
			Suffix:     inhabitant.Suffix,
			Birthday:   inhabitant.Birthday.Format("2006-01-02"),
			ContactNo:  inhabitant.ContactNo,
			Address:    inhabitant.Address,
		}

		respondJSON(w, http.StatusCreated, response)
	}
}

func (s *Server) handleUpdateInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid inhabitant id")
			return
		}

		var req createInhabitantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.FirstName == "" || req.LastName == "" {
			respondError(w, http.StatusBadRequest, "first_name and last_name are required")
			return
		}

		inhabitant := &inhabitant.Inhabitant{
			ID:         id,
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			MiddleName: req.MiddleName,
			Suffix:     req.Suffix,
			ContactNo:  req.ContactNo,
			Address:    req.Address,
		}

		// Parse birthday if provided
		if req.Birthday != "" {
			parsedDate, err := time.Parse("2006-01-02", req.Birthday)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid birthday format (use YYYY-MM-DD)")
				return
			}
			inhabitant.Birthday = parsedDate
		}

		if err := s.inhabitantService.Update(r.Context(), inhabitant); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update inhabitant")
			return
		}

		response := inhabitantResponse{
			ID:         inhabitant.ID,
			FirstName:  inhabitant.FirstName,
			LastName:   inhabitant.LastName,
			MiddleName: inhabitant.MiddleName,
			Suffix:     inhabitant.Suffix,
			Birthday:   inhabitant.Birthday.Format("2006-01-02"),
			ContactNo:  inhabitant.ContactNo,
			Address:    inhabitant.Address,
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func (s *Server) handleDeleteInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid inhabitant id")
			return
		}

		if err := s.inhabitantService.Delete(r.Context(), id); err != nil {
			respondError(w, http.StatusNotFound, "inhabitant not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleGetInhabitantDocuments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid inhabitant id")
			return
		}

		// Fetch the inhabitant by ID
		inhabitant, err := s.inhabitantService.Get(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "inhabitant not found")
			return
		}

		// Build the full name for precise matching
		fullName := buildInhabitantFullName(*inhabitant)

		// Search for documents containing this specific inhabitant
		documents, err := s.documentService.SearchByParticipants(r.Context(), []string{fullName})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to search documents")
			return
		}

		respondJSON(w, http.StatusOK, documents)
	}
}

// buildInhabitantFullName constructs a full name from an inhabitant's name parts
// matching the format used in documents
func buildInhabitantFullName(inh inhabitant.Inhabitant) string {
	parts := []string{}

	if inh.FirstName != "" {
		parts = append(parts, inh.FirstName)
	}
	if inh.MiddleName != "" {
		parts = append(parts, inh.MiddleName)
	}
	if inh.LastName != "" {
		parts = append(parts, inh.LastName)
	}
	if inh.Suffix != "" {
		parts = append(parts, inh.Suffix)
	}

	return strings.Join(parts, " ")
}
