package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"kpcms/server/core/document"
	"kpcms/server/core/inhabitant"

	"github.com/go-chi/chi/v5"
)

type CreateDocumentRequest struct {
	Code       string         `json:"code"`
	FolderName string         `json:"folder_name"`
	Title      string         `json:"title"`
	Fields     map[string]any `json:"fields"`
	CreatedAt  string         `json:"created_at"`
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

type JobType string

const (
	JobTypeBackup  JobType = "backup"
	JobTypeRestore JobType = "restore"
)

type JobStatus struct {
	Type     JobType `json:"type"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
	Error    string  `json:"error,omitempty"`
	Result   string  `json:"result,omitempty"`
}

var activeJobs sync.Map

func (s *Server) handleCreateDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if !isValidDocumentCode(req.Code) {
			respondError(w, http.StatusBadRequest, "invalid document code format")
			return
		}

		createdAt, err := parseDate(req.CreatedAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid created_at date format")
			return
		}

		// Validate inhabitant IDs
		complainantIDs, respondentIDs, err := extractInhabitantIDs(req.Fields)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Validate all IDs exist
		allIDs := append(complainantIDs, respondentIDs...)
		if err := s.validateInhabitantIDs(r.Context(), allIDs); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Denormalize names for display
		if err := s.denormalizeInhabitantNames(r.Context(), req.Fields, complainantIDs, respondentIDs); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to resolve inhabitant names")
			return
		}

		doc := document.Document{
			Code:       req.Code,
			FolderName: req.FolderName,
			Title:      req.Title,
			Fields:     req.Fields,
			CreatedAt:  createdAt,
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
		sortBy := r.URL.Query().Get("sort_by")
		sortDesc := false

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

		if sortDescStr := r.URL.Query().Get("sort_desc"); sortDescStr != "" {
			if parsed, err := strconv.ParseBool(sortDescStr); err == nil {
				sortDesc = parsed
			}
		}

		docs, err := s.documentService.List(r.Context(), offset, limit, sortBy, sortDesc)
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

		if !isValidDocumentCode(req.Code) {
			respondError(w, http.StatusBadRequest, "invalid document code format")
			return
		}

		createdAt, err := parseDate(req.CreatedAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid created_at date format")
			return
		}

		// Validate inhabitant IDs
		complainantIDs, respondentIDs, err := extractInhabitantIDs(req.Fields)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Validate all IDs exist
		allIDs := append(complainantIDs, respondentIDs...)
		if err := s.validateInhabitantIDs(r.Context(), allIDs); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Denormalize names for display
		if err := s.denormalizeInhabitantNames(r.Context(), req.Fields, complainantIDs, respondentIDs); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to resolve inhabitant names")
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

func (s *Server) handleUpdateFileContents() http.HandlerFunc {
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

		// Parse multipart form (32MB max)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			respondError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		if err := s.documentService.UpdateFileContents(r.Context(), uuid, fileName, file); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message":   "file contents updated successfully",
			"file_name": fileName,
		})
	}
}

type RenameFileRequest struct {
	NewName string `json:"new_name"`
}

func (s *Server) handleRenameFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		fileNameParam := chi.URLParam(r, "fileName")
		oldName, err := url.PathUnescape(fileNameParam)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}
		if oldName == "" {
			respondError(w, http.StatusBadRequest, "file name is required")
			return
		}

		var req RenameFileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.NewName == "" {
			respondError(w, http.StatusBadRequest, "new_name is required")
			return
		}

		if err := s.documentService.RenameFile(r.Context(), uuid, oldName, req.NewName); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				respondError(w, http.StatusNotFound, "document or file not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message":  "file renamed successfully",
			"old_name": oldName,
			"new_name": req.NewName,
		})
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

func (s *Server) handleListBackups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backups, err := s.documentService.ListBackups(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]any{
			"backups": backups,
		})
	}
}

func (s *Server) handleDownloadBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileNameParam := chi.URLParam(r, "fileName")

		// Chi parameters may still contain percent-encoding; normalize to the actual filename on disk
		fileName, err := url.PathUnescape(fileNameParam)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}

		if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}

		fullPath := filepath.Join(s.documentService.GetBackupPath(), fileName)

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
		w.Header().Set("Content-Type", "application/zip")
		http.ServeFile(w, r, fullPath)
	}
}

func (s *Server) handleCreateBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := fmt.Sprintf("bak_%d", time.Now().UnixNano())

		// Initial State
		activeJobs.Store(jobID, JobStatus{Type: JobTypeBackup, Progress: 0.0})

		go func() {
			update := func(p float64, status string, err string, res string) {
				activeJobs.Store(jobID, JobStatus{
					Type:     JobTypeBackup,
					Progress: p,
					Status:   status, // Send the status string
					Error:    err,
					Result:   res,
				})
			}

			path, err := s.documentService.CreateBackup(s.ctx, func(p float64) {
				update(p, "processing", "", "")
			})

			if err != nil {
				update(-1.0, "failed", err.Error(), "")
			} else {
				update(100.0, "completed", "", path)
			}
		}()

		respondJSON(w, http.StatusAccepted, map[string]string{
			"job_id":  jobID,
			"message": "Backup started",
		})
	}
}

func (s *Server) handleRestoreBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := r.URL.Query().Get("file_name")
		shouldOverwrite := (r.URL.Query().Get("mode") == "overwrite")

		if fileName == "" {
			respondError(w, http.StatusBadRequest, "file_name is required")
			return
		}

		if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
			respondError(w, http.StatusBadRequest, "invalid file name")
			return
		}

		jobID := fmt.Sprintf("res_%d", time.Now().UnixNano())

		activeJobs.Store(jobID, JobStatus{
			Type:     JobTypeRestore,
			Progress: 0.0,
			Status:   "processing",
		})

		go func() {
			update := func(p float64, status string, err string) {
				activeJobs.Store(jobID, JobStatus{
					Type:     JobTypeRestore,
					Progress: p,
					Status:   status,
					Error:    err,
				})
			}

			err := s.documentService.RestoreFromLocalPath(s.ctx, fileName, shouldOverwrite, func(p float64) {
				update(p, "processing", "")
			})

			if err != nil {
				update(-1.0, "failed", err.Error())
			} else {
				// Restore completed - reopen database connection to read from newly extracted file
				if dbErr := s.ReopenDatabaseConnection(); dbErr != nil {
					// Database reconnection failed - report as partial failure
					update(50.0, "partial_failure", fmt.Sprintf("restore completed but database reconnection failed: %v", dbErr))
					return
				}
				update(100.0, "completed", "")
			}
		}()

		respondJSON(w, http.StatusAccepted, map[string]string{
			"job_id":  jobID,
			"message": "Restore started",
		})
	}
}

func (s *Server) handleGetJobStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "jobID")

		if jobID == "" {
			jobID = r.URL.Query().Get("job_id")
		}

		if jobID == "" {
			respondError(w, http.StatusBadRequest, "Job ID is required")
			return
		}

		val, ok := activeJobs.Load(jobID)
		if !ok {
			respondError(w, http.StatusNotFound, "Job not found")
			return
		}

		respondJSON(w, http.StatusOK, val)
	}
}

func cond(c bool, t, f string) string {
	if c {
		return t
	}
	return f
}

func parseDate(dateStr string) (time.Time, error) {
	// Return zero time if string is empty
	if dateStr == "" {
		return time.Time{}, nil
	}

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

var validCodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^case_[0-9]+$`),            // case_1, case_0001, case_999999
	regexp.MustCompile(`^case_[0-9]{4}_[0-9]+$`),   // case_2026_0001
	regexp.MustCompile(`^[0-9]{3}-[0-9]{2}$`),      // 001-26
	regexp.MustCompile(`^case-[0-9]{3}-[0-9]{2}$`), // case-001-26
	regexp.MustCompile(`^case_[a-fA-F0-9]+$`),      // case_abcd, case_FFFF
}

func isValidDocumentCode(code string) bool {
	if code == "" {
		return false
	}
	for _, pattern := range validCodePatterns {
		if pattern.MatchString(code) {
			return true
		}
	}
	return false
}

// validateInhabitantIDs checks if all inhabitant IDs exist in the database.
// Returns error if any ID is invalid or doesn't exist.
func (s *Server) validateInhabitantIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("invalid inhabitant ID: %d", id)
		}

		_, err := s.inhabitantService.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("inhabitant with ID %d not found", id)
		}
	}

	return nil
}

// extractInhabitantIDs extracts inhabitant IDs from document fields
func extractInhabitantIDs(fields map[string]any) (complainants []int64, respondents []int64, err error) {
	if fields == nil {
		return nil, nil, nil
	}

	// Extract complainant_ids
	if raw, exists := fields["complainant_ids"]; exists {
		complainants, err = parseIDArrayFromInterface(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid complainant_ids format: %w", err)
		}
	}

	// Extract respondent_ids
	if raw, exists := fields["respondent_ids"]; exists {
		respondents, err = parseIDArrayFromInterface(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid respondent_ids format: %w", err)
		}
	}

	return complainants, respondents, nil
}

// parseIDArrayFromInterface converts various numeric array types to []int64
func parseIDArrayFromInterface(raw any) ([]int64, error) {
	switch v := raw.(type) {
	case []int64:
		return v, nil
	case []any:
		result := make([]int64, 0, len(v))
		for _, item := range v {
			id, err := toInt64FromInterface(item)
			if err != nil {
				return nil, err
			}
			result = append(result, id)
		}
		return result, nil
	case []int:
		result := make([]int64, 0, len(v))
		for _, id := range v {
			result = append(result, int64(id))
		}
		return result, nil
	case []float64:
		result := make([]int64, 0, len(v))
		for _, id := range v {
			result = append(result, int64(id))
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", raw)
	}
}

// toInt64FromInterface converts a numeric value to int64
func toInt64FromInterface(v any) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// denormalizeInhabitantNames resolves inhabitant IDs to names and stores them
// in the complainants and respondents fields for fast display without lookups.
func (s *Server) denormalizeInhabitantNames(ctx context.Context, fields map[string]any, complainantIDs, respondentIDs []int64) error {
	if fields == nil {
		return nil
	}

	// Resolve complainant names
	if len(complainantIDs) > 0 {
		complainantNames := make([]string, 0, len(complainantIDs))
		for _, id := range complainantIDs {
			inhabitant, err := s.inhabitantService.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to resolve complainant %d: %w", id, err)
			}
			complainantNames = append(complainantNames, buildFullName(inhabitant))
		}
		fields["complainants"] = complainantNames
	}

	// Resolve respondent names
	if len(respondentIDs) > 0 {
		respondentNames := make([]string, 0, len(respondentIDs))
		for _, id := range respondentIDs {
			inhabitant, err := s.inhabitantService.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to resolve respondent %d: %w", id, err)
			}
			respondentNames = append(respondentNames, buildFullName(inhabitant))
		}
		fields["respondents"] = respondentNames
	}

	return nil
}

// buildFullName constructs a full name from an inhabitant's name parts
func buildFullName(inh *inhabitant.Inhabitant) string {
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
