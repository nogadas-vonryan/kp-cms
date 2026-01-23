package api

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strconv"

	"main/internal/document"

	"github.com/go-chi/chi/v5"
)

type CreateDocumentRequest struct {
	Code       string         `json:"code"`
	FolderName string         `json:"folder_name"`
	Title      string         `json:"title"`
	Fields     map[string]any `json:"fields"`
}

type UpdateDocumentRequest struct {
	Code   string         `json:"code"`
	Title  string         `json:"title"`
	Fields map[string]any `json:"fields"`
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
		limit := 15

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

		doc := document.Document{
			Code:   req.Code,
			Title:  req.Title,
			Fields: req.Fields,
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

		respondJSON(w, http.StatusCreated, map[string]string{
			"message":   "file uploaded successfully",
			"file_name": fileName,
		})
	}
}

func (s *Server) handleDeleteFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			respondError(w, http.StatusBadRequest, "uuid is required")
			return
		}

		fileName := chi.URLParam(r, "fileName")
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
