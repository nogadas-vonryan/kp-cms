package api

import (
	"context"
	"encoding/json"
	"fmt"
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
	ID                           int64  `json:"id"`
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

type documentRefreshError struct {
	UUID  string `json:"uuid"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

type updateInhabitantResponse struct {
	inhabitantResponse
	DocumentsUpdated int                    `json:"documents_updated"`
	DocumentErrors   []documentRefreshError `json:"document_errors,omitempty"`
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

		var inhabitants []inhabitant.Inhabitant
		var err error

		if query != "" {
			// Use search functionality with tokenized matching
			inhabitants, err = s.inhabitantService.FindPeopleByName(r.Context(), query, limit)
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
			response[i] = inhabitantResponse{
				ID:                           inh.ID,
				FirstName:                    inh.FirstName,
				LastName:                     inh.LastName,
				MiddleName:                   inh.MiddleName,
				Suffix:                       inh.Suffix,
				Birthdate:                    inh.Birthdate.Format("2006-01-02"),
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
			ID:                           inhabitant.ID,
			FirstName:                    inhabitant.FirstName,
			LastName:                     inhabitant.LastName,
			MiddleName:                   inhabitant.MiddleName,
			Suffix:                       inhabitant.Suffix,
			Birthdate:                    inhabitant.Birthdate.Format("2006-01-02"),
			BirthPlace:                   inhabitant.BirthPlace,
			InhabitantType:               inhabitant.InhabitantType,
			Sex:                          inhabitant.Sex,
			CivilStatus:                  inhabitant.CivilStatus,
			Citizenship:                  inhabitant.Citizenship,
			Occupation:                   inhabitant.Occupation,
			EmailAddress:                 inhabitant.EmailAddress,
			HighestEducationalAttainment: inhabitant.HighestEducationalAttainment,
			MotherFirstName:              inhabitant.MotherFirstName,
			MotherMiddleName:             inhabitant.MotherMiddleName,
			MotherLastName:               inhabitant.MotherLastName,
			ContactNo:                    inhabitant.ContactNo,
			Address:                      inhabitant.Address,
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
			FirstName:                    req.FirstName,
			LastName:                     req.LastName,
			MiddleName:                   req.MiddleName,
			Suffix:                       req.Suffix,
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

		// Parse birthdate if provided
		if req.Birthdate != "" {
			parsedDate, err := time.Parse("2006-01-02", req.Birthdate)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid birthdate format (use YYYY-MM-DD)")
				return
			}
			inhabitant.Birthdate = parsedDate
		}

		id, err := s.inhabitantService.Create(r.Context(), inhabitant)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create inhabitant")
			return
		}

		inhabitant.ID = id
		response := inhabitantResponse{
			ID:                           inhabitant.ID,
			FirstName:                    inhabitant.FirstName,
			LastName:                     inhabitant.LastName,
			MiddleName:                   inhabitant.MiddleName,
			Suffix:                       inhabitant.Suffix,
			Birthdate:                    inhabitant.Birthdate.Format("2006-01-02"),
			BirthPlace:                   inhabitant.BirthPlace,
			InhabitantType:               inhabitant.InhabitantType,
			Sex:                          inhabitant.Sex,
			CivilStatus:                  inhabitant.CivilStatus,
			Citizenship:                  inhabitant.Citizenship,
			Occupation:                   inhabitant.Occupation,
			EmailAddress:                 inhabitant.EmailAddress,
			HighestEducationalAttainment: inhabitant.HighestEducationalAttainment,
			MotherFirstName:              inhabitant.MotherFirstName,
			MotherMiddleName:             inhabitant.MotherMiddleName,
			MotherLastName:               inhabitant.MotherLastName,
			ContactNo:                    inhabitant.ContactNo,
			Address:                      inhabitant.Address,
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
			ID:                           id,
			FirstName:                    req.FirstName,
			LastName:                     req.LastName,
			MiddleName:                   req.MiddleName,
			Suffix:                       req.Suffix,
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

		// Parse birthdate if provided
		if req.Birthdate != "" {
			parsedDate, err := time.Parse("2006-01-02", req.Birthdate)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid birthdate format (use YYYY-MM-DD)")
				return
			}
			inhabitant.Birthdate = parsedDate
		}

		if err := s.inhabitantService.Update(r.Context(), inhabitant); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update inhabitant")
			return
		}

		updatedCount, refreshErrors := s.refreshInhabitantDocumentNames(r.Context(), inhabitant.ID)

		response := updateInhabitantResponse{
			inhabitantResponse: inhabitantResponse{
				ID:         inhabitant.ID,
				FirstName:  inhabitant.FirstName,
				LastName:   inhabitant.LastName,
				MiddleName: inhabitant.MiddleName,
				Suffix:     inhabitant.Suffix,
				Birthdate:  inhabitant.Birthdate.Format("2006-01-02"),
				ContactNo:  inhabitant.ContactNo,
				Address:    inhabitant.Address,
			},
			DocumentsUpdated: updatedCount,
			DocumentErrors:   refreshErrors,
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func (s *Server) refreshInhabitantDocumentNames(ctx context.Context, inhabitantID int64) (int, []documentRefreshError) {
	documents, err := s.documentService.GetDocumentsByInhabitantID(ctx, inhabitantID)
	if err != nil {
		return 0, []documentRefreshError{{
			Error: fmt.Sprintf("failed to load linked documents: %v", err),
		}}
	}

	updatedCount := 0
	refreshErrors := make([]documentRefreshError, 0)

	for _, doc := range documents {
		complainantIDs := doc.GetComplainantIDs()
		respondentIDs := doc.GetRespondentIDs()

		if err := s.denormalizeInhabitantNames(ctx, doc.Fields, complainantIDs, respondentIDs); err != nil {
			refreshErrors = append(refreshErrors, documentRefreshError{
				UUID:  doc.UUID,
				Code:  doc.Code,
				Error: err.Error(),
			})
			continue
		}

		updatedDoc := document.Document{
			Code:      doc.Code,
			Title:     doc.Title,
			Fields:    doc.Fields,
			CreatedAt: doc.CreatedAt,
		}

		if _, err := s.documentService.Update(ctx, doc.UUID, updatedDoc); err != nil {
			refreshErrors = append(refreshErrors, documentRefreshError{
				UUID:  doc.UUID,
				Code:  doc.Code,
				Error: err.Error(),
			})
			continue
		}

		updatedCount++
	}

	return updatedCount, refreshErrors
}

func (s *Server) handleDeleteInhabitant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid inhabitant id")
			return
		}

		// Check if inhabitant has linked documents (soft-block)
		linkedDocs, err := s.documentService.GetDocumentsByInhabitantID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to check linked documents")
			return
		}

		if len(linkedDocs) > 0 {
			respondError(w, http.StatusConflict, fmt.Sprintf("cannot delete inhabitant: linked to %d document(s)", len(linkedDocs)))
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

		// Fetch documents directly by inhabitant ID using the reverse index
		documents, err := s.documentService.GetDocumentsByInhabitantID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to fetch documents")
			return
		}

		respondJSON(w, http.StatusOK, documents)
	}
}
