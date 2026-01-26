package api

import (
	"fmt"
	"net/http"
	"time"

	"main/internal/auth"
	"main/internal/document"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	Host            string
	Port            string
	Router          *chi.Mux
	documentService *document.DocumentService
	userStore       *auth.UserStore
	sessionManager  *auth.SessionManager
	sessionTTL      time.Duration
}

func NewServer(host, port, flagUser, flagPass string, documentService *document.DocumentService) (*Server, error) {
	userStore := auth.NewUserStore()
	if err := userStore.AddUser(flagUser, flagPass, auth.RoleAdmin); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	s := &Server{
		Host:            host,
		Port:            port,
		Router:          chi.NewRouter(),
		documentService: documentService,
		userStore:       userStore,
		sessionManager:  auth.NewSessionManager(24 * time.Hour),
		sessionTTL:      24 * time.Hour,
	}

	s.routes()
	return s, nil
}

func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s *Server) routes() {
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public Routes
	s.Router.Get("/", s.handleVersion())
	s.Router.Get("/health", s.handleHealth())
	s.Router.Post("/auth/login", s.handleLogin())
	s.Router.Post("/auth/register", s.handleRegister())
	s.Router.Post("/auth/logout", s.handleLogout())

	s.Router.Group(func(r chi.Router) {
		// Authenticated routes
		r.Use(auth.SessionMiddleware(s.sessionManager))
		r.Use(auth.CSRFMiddleware())

		r.Get("/auth/me", s.handleMe())

		r.Route("/auth/admin", func(admin chi.Router) {
			admin.Use(auth.RequireRole(auth.RoleAdmin))
			admin.Get("/users", s.handleListUsers())
			admin.Post("/users", s.handleCreateUser())
			admin.Post("/users/{username}/role", s.handleUpdateUserRole())
			admin.Delete("/users/{username}", s.handleDeleteUser())
		})

		r.Route("/documents", func(r chi.Router) {
			r.Get("/", s.handleListDocuments())
			r.Get("/search", s.handleSearchDocuments())
			r.Get("/{uuid}", s.handleGetDocumentByUUID())
			r.Get("/code/{code}", s.handleGetDocumentByCode())

			r.Get("/{uuid}/files/{fileName}", s.handleDownloadFile())

			r.Group(func(admin chi.Router) {
				admin.Use(auth.RequireRole(auth.RoleAdmin))

				admin.Post("/", s.handleCreateDocument())
				admin.Put("/{uuid}", s.handleUpdateDocument())
				admin.Delete("/{uuid}", s.handleDeleteDocument())

				admin.Post("/{uuid}/files", s.handleUploadFile())
				admin.Put("/{uuid}/files/{fileName}", s.handleUpdateFileMetadata())
				admin.Delete("/{uuid}/files/{fileName}", s.handleDeleteFile())

				admin.Get("/conflicts", s.handleGetConflicts())
				admin.Post("/reload", s.handleReloadDocuments())
				admin.Post("/reload/{folderName}", s.handleReloadDocument())
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
