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
	FirstName                    string `json:"first_name"`
	LastName                     string `json:"last_name"`
	MiddleName                   string `json:"middle_name"`
	Suffix                       string `json:"suffix"`
	Birthdate                    string `json:"birthdate"` // ISO 8601 date format
	BirthPlace                   string `json:"birth_place"`
	InhabitantType               string `json:"inhabitant_type"`
	Sex                          string `json:"sex"`
	CivilStatus                  string `json:"civil_status"`
	Citizenship                  string `json:"citizenship"`
	Occupation                   string `json:"occupation"`
	EmailAddress                 string `json:"email_address"`
	HighestEducationalAttainment string `json:"highest_educational_attainment"`
	MotherFirstName              string `json:"mother_first_name"`
	MotherMiddleName             string `json:"mother_middle_name"`
	MotherLastName               string `json:"mother_last_name"`
	ContactNo                    string `json:"contact_no"`
	Address                      string `json:"address"`
}

type inhabitantResponse struct {
	UUID                         string `json:"uuid"`
	Code                         string `json:"code"`
	FolderName                   string `json:"folder_name"`
	FirstName                    string `json:"first_name"`
	LastName                     string `json:"last_name"`
	MiddleName                   string `json:"middle_name"`
	Suffix                       string `json:"suffix"`
	Birthdate                    string `json:"birthdate"`
	BirthPlace                   string `json:"birth_place"`
	InhabitantType               string `json:"inhabitant_type"`
	Sex                          string `json:"sex"`
	CivilStatus                  string `json:"civil_status"`
	Citizenship                  string `json:"citizenship"`
	Occupation                   string `json:"occupation"`
	EmailAddress                 string `json:"email_address"`
	HighestEducationalAttainment string `json:"highest_educational_attainment"`
	MotherFirstName              string `json:"mother_first_name"`
	MotherMiddleName             string `json:"mother_middle_name"`
	MotherLastName               string `json:"mother_last_name"`
	ContactNo                    string `json:"contact_no"`
	Address                      string `json:"address"`
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

		// Check if there's a search query
		query := r.URL.Query().Get("q")

		var inhabitants []*inhabitant.Inhabitant
		var err error

		if query != "" {
			// Use search functionality
			inhabitants, err = s.inhabitantService.Search(r.Context(), query)
		} else {
			// Use regular list with pagination
			inhabitants, err = s.inhabitantService.List(r.Context(), limit, offset)
		}

		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list inhabitants")
			return
		}

		response := make([]inhabitantResponse, len(inhabitants))
		for i, inh := range inhabitants {
			birthdateStr := ""
			if !inh.Birthdate.IsZero() {
				birthdateStr = inh.Birthdate.Format("2006-01-02")
			}
			response[i] = inhabitantResponse{
				UUID:                         inh.UUID,
				Code:                         inh.Code,
				FolderName:                   inh.FolderName,
				FirstName:                    inh.FirstName,
				LastName:                     inh.LastName,
				MiddleName:                   inh.MiddleName,
				Suffix:                       inh.Suffix,
				Birthdate:                    birthdateStr,
				BirthPlace:                   inh.BirthPlace,
				InhabitantType:               inh.InhabitantType,
				Sex:                          inh.Sex,
				CivilStatus:                  inh.CivilStatus,
				Citizenship:                  inh.Citizenship,
				Occupation:                   inh.Occupation,
				EmailAddress:                 inh.EmailAddress,
				HighestEducationalAttainment: inh.HighestEducationalAttainment,
				MotherFirstName:              inh.MotherFirstName,
				MotherMiddleName:             inh.MotherMiddleName,
				MotherLastName:               inh.MotherLastName,
				ContactNo:                    inh.ContactNo,
				Address:                      inh.Address,
			}
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func (s *Server) handleGetInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idOrUUID := chi.URLParam(r, "id")

		// Try UUID first (file-based store)
		inh, err := s.inhabitantService.GetByUUID(r.Context(), idOrUUID)
		if err == nil && inh != nil {
			s.respondInhabitantJSON(w, http.StatusOK, inh)
			return
		}

		// Try ID (SQLite-based)
		if id, parseErr := strconv.ParseInt(idOrUUID, 10, 64); parseErr == nil {
			inh, err = s.inhabitantService.GetByID(r.Context(), id)
			if err == nil && inh != nil {
				s.respondInhabitantJSON(w, http.StatusOK, inh)
				return
			}
		}

		respondError(w, http.StatusNotFound, "inhabitant not found")
	}
}

// respondInhabitantJSON writes an inhabitant response to the client.
func (s *Server) respondInhabitantJSON(w http.ResponseWriter, status int, inh *inhabitant.Inhabitant) {
	birthdateStr := ""
	if !inh.Birthdate.IsZero() {
		birthdateStr = inh.Birthdate.Format("2006-01-02")
	}

	response := inhabitantResponse{
		UUID:                         inh.UUID,
		Code:                         inh.Code,
		FolderName:                   inh.FolderName,
		FirstName:                    inh.FirstName,
		LastName:                     inh.LastName,
		MiddleName:                   inh.MiddleName,
		Suffix:                       inh.Suffix,
		Birthdate:                    birthdateStr,
		BirthPlace:                   inh.BirthPlace,
		InhabitantType:               inh.InhabitantType,
		Sex:                          inh.Sex,
		CivilStatus:                  inh.CivilStatus,
		Citizenship:                  inh.Citizenship,
		Occupation:                   inh.Occupation,
		EmailAddress:                 inh.EmailAddress,
		HighestEducationalAttainment: inh.HighestEducationalAttainment,
		MotherFirstName:              inh.MotherFirstName,
		MotherMiddleName:             inh.MotherMiddleName,
		MotherLastName:               inh.MotherLastName,
		ContactNo:                    inh.ContactNo,
		Address:                      inh.Address,
	}

	respondJSON(w, status, response)
}

func (s *Server) handleCreateInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createInhabitantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.FirstName == "" || req.LastName == "" {
			respondError(w, http.StatusBadRequest, "first_name and last_name are required")
			return
		}

		// Parse birthdate if provided
		var birthdate time.Time
		if req.Birthdate != "" {
			var err error
			birthdate, err = time.Parse("2006-01-02", req.Birthdate)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid birthdate format, use YYYY-MM-DD")
				return
			}
		}

		inh := &inhabitant.Inhabitant{
			FirstName:                    req.FirstName,
			LastName:                     req.LastName,
			MiddleName:                   req.MiddleName,
			Suffix:                       req.Suffix,
			Birthdate:                    birthdate,
			BirthPlace:                   req.BirthPlace,
			InhabitantType:               req.InhabitantType,
			Sex:                          req.Sex,
			CivilStatus:                  req.CivilStatus,
			Citizenship:                  req.Citizenship,
			Occupation:                   req.Occupation,
			EmailAddress:                 req.EmailAddress,
			HighestEducationalAttainment: req.HighestEducationalAttainment,
			MotherFirstName:              req.MotherFirstName,
			MotherMiddleName:             req.MotherMiddleName,
			MotherLastName:               req.MotherLastName,
			ContactNo:                    req.ContactNo,
			Address:                      req.Address,
		}

		created, err := s.inhabitantService.Create(r.Context(), inh)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create inhabitant")
			return
		}

		s.respondInhabitantJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleUpdateInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idOrUUID := chi.URLParam(r, "id")

		var req createInhabitantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Parse birthdate if provided
		var birthdate time.Time
		if req.Birthdate != "" {
			var err error
			birthdate, err = time.Parse("2006-01-02", req.Birthdate)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid birthdate format, use YYYY-MM-DD")
				return
			}
		}

		inh := &inhabitant.Inhabitant{
			FirstName:                    req.FirstName,
			LastName:                     req.LastName,
			MiddleName:                   req.MiddleName,
			Suffix:                       req.Suffix,
			Birthdate:                    birthdate,
			BirthPlace:                   req.BirthPlace,
			InhabitantType:               req.InhabitantType,
			Sex:                          req.Sex,
			CivilStatus:                  req.CivilStatus,
			Citizenship:                  req.Citizenship,
			Occupation:                   req.Occupation,
			EmailAddress:                 req.EmailAddress,
			HighestEducationalAttainment: req.HighestEducationalAttainment,
			MotherFirstName:              req.MotherFirstName,
			MotherMiddleName:             req.MotherMiddleName,
			MotherLastName:               req.MotherLastName,
			ContactNo:                    req.ContactNo,
			Address:                      req.Address,
		}

		// Try UUID first (file-based store)
		updated, err := s.inhabitantService.Update(r.Context(), idOrUUID, inh)
		if err == nil && updated != nil {
			s.respondInhabitantJSON(w, http.StatusOK, updated)
			return
		}

		// Check if error indicates wrong method (store not initialized)
		if err != nil && strings.Contains(err.Error(), "UpdateByID") {
			// Try ID (SQLite-based)
			if id, parseErr := strconv.ParseInt(idOrUUID, 10, 64); parseErr == nil {
				err = s.inhabitantService.UpdateByID(r.Context(), id, inh)
				if err == nil {
					// Fetch the updated inhabitant
					updated, _ := s.inhabitantService.GetByID(r.Context(), id)
					if updated != nil {
						s.respondInhabitantJSON(w, http.StatusOK, updated)
						return
					}
				}
				// Check if it was a not found error
				if err != nil {
					respondError(w, http.StatusNotFound, "inhabitant not found")
					return
				}
			}
		}

		// Check if it was a not found error from file store
		if err != nil && strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "inhabitant not found")
			return
		}

		// Handle other errors
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondError(w, http.StatusInternalServerError, "failed to update inhabitant")
	}
}

func (s *Server) handleDeleteInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idOrUUID := chi.URLParam(r, "id")

		// Check if it looks like a UUID (contains dashes)
		isUUID := strings.Contains(idOrUUID, "-")

		if isUUID {
			// Try UUID first (file-based store)
			err := s.inhabitantService.Delete(r.Context(), idOrUUID)
			if err == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// UUID not found in file store - return 404
			respondError(w, http.StatusNotFound, "inhabitant not found")
			return
		}

		// Try ID (SQLite-based)
		if id, parseErr := strconv.ParseInt(idOrUUID, 10, 64); parseErr == nil {
			err := s.inhabitantService.DeleteByID(r.Context(), id)
			if err == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// ID not found
			respondError(w, http.StatusNotFound, "inhabitant not found")
			return
		}

		// Invalid ID format
		respondError(w, http.StatusBadRequest, "invalid inhabitant ID")
	}
}

func (s *Server) handleGetInhabitantDocuments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This endpoint would require the search aggregator to find documents by inhabitant
		// For now, return an empty list as it's not fully implemented
		respondJSON(w, http.StatusOK, []map[string]string{})
	}
}
