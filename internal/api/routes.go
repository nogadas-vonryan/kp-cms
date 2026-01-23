package api

import (
	"fmt"
	"net/http"

	"main/internal/auth"
	"main/internal/document"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Host            string
	Port            string
	Router          *chi.Mux
	documentService *document.DocumentService
	Config          struct {
		FlagUser string
		FlagPass string
	}
}

func NewServer(host, port, flagUser, flagPass string, documentService *document.DocumentService) *Server {
	s := &Server{
		Host:            host,
		Port:            port,
		Router:          chi.NewRouter(),
		documentService: documentService,
		Config: struct {
			FlagUser string
			FlagPass string
		}{
			FlagUser: flagUser,
			FlagPass: flagPass,
		},
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

	// Public Routes
	s.Router.Get("/", s.handleVersion())
	s.Router.Get("/health", s.handleHealth())

	s.Router.Group(func(r chi.Router) {
		// Auth
		r.Use(auth.AuthMiddleware(s.Config.FlagUser, s.Config.FlagPass))

		r.Route("/documents", func(r chi.Router) {
			// Standard User routes
			r.Get("/", s.handleListDocuments())
			r.Get("/{uuid}", s.handleGetDocumentByUUID())
			r.Get("/code/{code}", s.handleGetDocumentByCode())

			// Admin-only routes
			r.Group(func(admin chi.Router) {
				admin.Use(auth.RequireRole(auth.RoleAdmin))

				admin.Post("/", s.handleCreateDocument())
				admin.Post("/{uuid}/files", s.handleUploadFile())
				admin.Delete("/{uuid}", s.handleDeleteDocument())
			})
		})
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
