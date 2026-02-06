package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kpcms/server/core/document"
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

		// Try by code as fallback
		inh, err = s.inhabitantService.GetByCode(r.Context(), idOrUUID)
		if err == nil && inh != nil {
			s.respondInhabitantJSON(w, http.StatusOK, inh)
			return
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

		updated, err := s.inhabitantService.Update(r.Context(), idOrUUID, inh)
		if err != nil {
			if err.Error() == "file does not exist" {
				respondError(w, http.StatusNotFound, "inhabitant not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		s.respondInhabitantJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeleteInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idOrUUID := chi.URLParam(r, "id")

		err := s.inhabitantService.Delete(r.Context(), idOrUUID)
		if err != nil {
			if err.Error() == "file does not exist" {
				respondError(w, http.StatusNotFound, "inhabitant not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleGetInhabitantDocuments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idOrUUID := chi.URLParam(r, "id")

		// Try to get inhabitant by UUID first
		inh, err := s.inhabitantService.GetByUUID(r.Context(), idOrUUID)
		if err != nil {
			// Try by code as fallback
			inh, err = s.inhabitantService.GetByCode(r.Context(), idOrUUID)
			if err != nil {
				respondError(w, http.StatusNotFound, "inhabitant not found")
				return
			}
		}

		// Build the inhabitant code for lookup
		// The format is "inhabitant-XXX-YY" or just "XXX-YY"
		inhabitantCode := inh.FolderName
		if inhabitantCode == "" {
			inhabitantCode = "inhabitant-" + inh.Code
		}

		// Get documents linked to this inhabitant
		docs, err := s.documentService.GetDocumentsByInhabitantCode(r.Context(), inhabitantCode)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get documents")
			return
		}

		// Convert to response format
		response := make([]document.Document, len(docs))
		for i, doc := range docs {
			response[i] = document.Document{
				UUID:           doc.UUID,
				Code:           doc.Code,
				FolderName:     doc.FolderName,
				Title:          doc.Title,
				Fields:         doc.Fields,
				ParticipantIDs: doc.ParticipantIDs,
				Files:          doc.Files,
				CreatedAt:      doc.CreatedAt,
				UpdatedAt:      doc.UpdatedAt,
			}
		}

		respondJSON(w, http.StatusOK, response)
	}
}
