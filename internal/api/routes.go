package api

import (
	"fmt"
	"net/http"

	"main/internal/document"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Host            string
	Port            string
	Router          *chi.Mux
	documentService *document.DocumentService
}

func NewServer(host, port string, documentService *document.DocumentService) *Server {
	s := &Server{
		Host:            host,
		Port:            port,
		Router:          chi.NewRouter(),
		documentService: documentService,
	}

	s.routes()
	return s
}

func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s *Server) routes() {
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)

	s.Router.Get("/", s.handleVersion())
	s.Router.Get("/health", s.handleHealth())

	// Document routes
	s.Router.Route("/documents", func(r chi.Router) {
		r.Post("/", s.handleCreateDocument())
		r.Get("/", s.handleListDocuments())
		r.Get("/code/{code}", s.handleGetDocumentByCode())
		r.Get("/{uuid}", s.handleGetDocument())
		r.Put("/{uuid}", s.handleUpdateDocument())
		r.Delete("/{uuid}", s.handleDeleteDocument())

		// File routes
		r.Post("/{uuid}/files", s.handleUploadFile())
		r.Delete("/{uuid}/files/{fileName}", s.handleDeleteFile())
	})
}

func (s *Server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Archivist Server - Prototype"))
	}
}

func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
