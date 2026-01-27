package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"kpcms/server/core/document"

	"github.com/go-chi/chi/v5"
)

type CreateDocumentRequest struct {
	Code       string         `json:"code"`
	FolderName string         `json:"folder_name"`
	Title      string         `json:"title"`
	Fields     map[string]any `json:"fields"`
}

type UpdateDocumentRequest struct {
	Code      string         `json:"code"`
	Title     string         `json:"title"`
	Fields    map[string]any `json:"fields"`
	CreatedAt string         `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleCreateDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		doc := document.Document{
			Code:       req.Code,
			FolderName: req.FolderName,
			Title:      req.Title,
			Fields:     req.Fields,
		}

		created, err := s.documentService.Create(r.Context(), doc)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, created)
	}
}

func (s *Server) handleListDocuments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		offset := 0
		limit := 10

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		docs, err := s.documentService.List(r.Context(), offset, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, docs)
	}
}

func (s *Server) handleGetDocumentByUUID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		doc, err := s.documentService.GetByUUID(r.Context(), uuid)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, doc)
	}
}

func (s *Server) handleGetDocumentByCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		if code == "" {
			respondError(w, http.StatusBadRequest, "code is required")
			return
		}

		doc, err := s.documentService.GetByCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, doc)
	}
}

func (s *Server) handleUpdateDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		var req UpdateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		createdAt, err := parseDate(req.CreatedAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid created_at date format")
			return
		}

		doc := document.Document{
			Code:      req.Code,
			Title:     req.Title,
			Fields:    req.Fields,
			CreatedAt: createdAt,
		}

		updated, err := s.documentService.Update(r.Context(), uuid, doc)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, updated)
	}
}

func (s *Server) handleDeleteDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		if err := s.documentService.Delete(r.Context(), uuid); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleUploadFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		// Parse multipart form (32MB max)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		metadataUpdate := document.FileMetadataUpdate{}
		if r.MultipartForm != nil {
			if descVals, ok := r.MultipartForm.Value["description"]; ok && len(descVals) > 0 {
				desc := descVals[0]
				metadataUpdate.Description = &desc
			}
			if noteVals, ok := r.MultipartForm.Value["note"]; ok && len(noteVals) > 0 {
				note := noteVals[0]
				metadataUpdate.Note = &note
			}
			if tagVals, ok := r.MultipartForm.Value["tags[]"]; ok && len(tagVals) > 0 {
				metadataUpdate.Tags = &tagVals
			}
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			respondError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		fileName := header.Filename
		if fileName == "" {
			respondError(w, http.StatusBadRequest, "file name is required")
			return
		}

		if err := s.documentService.UploadFile(r.Context(), uuid, fileName, file); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		hasMetadataUpdate := metadataUpdate.Description != nil || metadataUpdate.Note != nil || metadataUpdate.Tags != nil
		if hasMetadataUpdate {
			if err := s.documentService.UpdateFileMetadata(r.Context(), uuid, fileName, metadataUpdate); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					respondError(w, http.StatusNotFound, "document not found")
					return
				}
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		respondJSON(w, http.StatusCreated, map[string]string{
			"message":   "file uploaded successfully",
			"file_name": fileName,
		})
	}
}

type UpdateFileMetadataRequest struct {
	Description *string   `json:"description"`
	Note        *string   `json:"note"`
	Tags        *[]string `json:"tags"`
}

func (s *Server) handleDownloadFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		fileNameParam := chi.URLParam(r, "fileName")

		// Chi parameters may still contain percent-encoding; normalize to the actual filename on disk
		fileName, err := url.PathUnescape(fileNameParam)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}

		if uuid == "" || fileName == "" {
			respondError(w, http.StatusBadRequest, "uuid and file name are required")
			return
		}

		fileReader, err := s.documentService.DownloadFile(r.Context(), uuid, fileName)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "file or document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer fileReader.Close()

		// Set headers to trigger a download in the browser
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := io.Copy(w, fileReader); err != nil {
			// Headers are already sent, so we can't respond with JSON here
			return
		}
	}
}

func (s *Server) handleUpdateFileMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		fileNameParam := chi.URLParam(r, "fileName")
		fileName, err := url.PathUnescape(fileNameParam)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}

		var req UpdateFileMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		updates := document.FileMetadataUpdate{}
		if req.Description != nil {
			updates.Description = req.Description
		}
		if req.Note != nil {
			updates.Note = req.Note
		}
		if req.Tags != nil {
			updates.Tags = req.Tags
		}

		err = s.documentService.UpdateFileMetadata(r.Context(), uuid, fileName, updates)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "file or document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "metadata updated successfully"})
	}
}

func (s *Server) handleDeleteFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		fileNameParam := chi.URLParam(r, "fileName")
		fileName, err := url.PathUnescape(fileNameParam)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}
		if fileName == "" {
			respondError(w, http.StatusBadRequest, "file name is required")
			return
		}

		if err := s.documentService.DeleteFile(r.Context(), uuid, fileName); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleGetConflicts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conflicts, err := s.documentService.GetConflicts(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, conflicts)
	}
}

func (s *Server) handleReloadDocuments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conflicts, err := s.documentService.ReloadCache(r.Context())
		if err != nil {
			http.Error(w, "Failed to initiate reload", 500)
			return
		}

		response := map[string]interface{}{
			"status":    "success",
			"conflicts": conflicts, // User sees these in the JSON
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (s *Server) handleReloadDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folderName := chi.URLParam(r, "folderName")
		if folderName == "" {
			respondError(w, http.StatusBadRequest, "folder name is required")
			return
		}

		conflicts, err := s.documentService.ReloadCacheForFolder(r.Context(), folderName)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response := map[string]interface{}{
			"status":    "success",
			"conflicts": conflicts,
		}
		respondJSON(w, http.StatusOK, response)
	}
}

func (s *Server) handleSearchDocuments() http.HandlerFunc {
	const MaxLimit = 100

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		criteria := document.SearchCriteria{
			UUID:       q.Get("uuid"),
			Title:      q.Get("title"),
			Code:       q.Get("code"),
			FolderName: q.Get("folder_name"),
			FieldKey:   q.Get("field_key"),
			SortBy:     q.Get("sort_by"),
			SortDesc:   q.Get("sort_desc") == "true",
		}

		// Parse date range
		if dateFrom := q.Get("date_from"); dateFrom != "" {
			if t, err := parseDate(dateFrom); err == nil {
				criteria.DateFrom = &t
			}
		}
		if dateTo := q.Get("date_to"); dateTo != "" {
			if t, err := parseDate(dateTo); err == nil {
				criteria.DateTo = &t
			}
		}

		// Parse field filters from query params like field_status=mediation
		criteria.FieldFilters = make(map[string]any)
		for key, values := range q {
			if len(values) > 0 && len(key) > 6 && key[:6] == "field_" {
				fieldName := key[6:]
				criteria.FieldFilters[fieldName] = values[0]
			}
		}

		// Parse pagination
		offset := 0
		limit := 10
		if offsetStr := q.Get("offset"); offsetStr != "" {
			if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
				offset = parsed
			}
		}
		if limitStr := q.Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed

				if limit > MaxLimit {
					limit = MaxLimit
				}
			}
		}
		criteria.Offset = offset
		criteria.Limit = limit

		docs, err := s.documentService.Search(r.Context(), criteria)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, docs)
	}
}

func parseDate(dateStr string) (time.Time, error) {
	// Try multiple date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid date format")
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
