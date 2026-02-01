package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"kpcms/server/core/auth"
	"kpcms/server/core/document"
	"kpcms/server/core/inhabitant"
	"kpcms/server/core/search"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	Host              string
	Port              string
	Router            *chi.Mux
	documentService   *document.DocumentService
	authService       *auth.Service
	inhabitantService *inhabitant.Service
	searchService     *search.AggregatorService
	sessionTTL        time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewServer(host, port, flagUser, flagPass string, documentService *document.DocumentService, authService *auth.Service, inhabitantService *inhabitant.Service, searchService *search.AggregatorService) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	success := false
	defer func() {
		if !success {
			cancel()
		}
	}()

	if err := authService.CreateUser(ctx, flagUser, flagPass, auth.RoleAdmin); err != nil && !errors.Is(err, auth.ErrUserExists) {
		cancel()
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	s := &Server{
		Host:              host,
		Port:              port,
		Router:            chi.NewRouter(),
		documentService:   documentService,
		authService:       authService,
		inhabitantService: inhabitantService,
		searchService:     searchService,
		sessionTTL:        24 * time.Hour,
		ctx:               ctx,
		cancel:            cancel,
	}

	s.routes()

	success = true
	return s, nil
}

func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s *Server) routes() {
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(AccessLogMiddleware)
	s.Router.Use(middleware.Recoverer)

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s.Router.Route("/api", func(r chi.Router) {
		// Public Routes
		r.Get("/", s.handleVersion())
		r.Get("/health", s.handleHealth())
		r.Post("/auth/login", s.handleLogin())
		r.Post("/auth/register", s.handleRegister())
		r.Post("/auth/logout", s.handleLogout())

		r.Group(func(r chi.Router) {
			// Authenticated routes
			r.Use(s.SessionMiddleware())
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

				r.Get("/{uuid}/files/{fileName:.+}", s.handleDownloadFile())

				r.Group(func(admin chi.Router) {
					admin.Use(auth.RequireRole(auth.RoleAdmin))

					admin.Post("/", s.handleCreateDocument())
					admin.Put("/{uuid}", s.handleUpdateDocument())
					admin.Delete("/{uuid}", s.handleDeleteDocument())

					admin.Post("/{uuid}/files", s.handleUploadFile())
					admin.Put("/{uuid}/files/{fileName:.+}", s.handleUpdateFileMetadata())
					admin.Patch("/{uuid}/files/{fileName:.+}/contents", s.handleUpdateFileContents())
					admin.Patch("/{uuid}/files/{fileName:.+}/rename", s.handleRenameFile())
					admin.Delete("/{uuid}/files/{fileName:.+}", s.handleDeleteFile())

					admin.Get("/conflicts", s.handleGetConflicts())
					admin.Post("/reload", s.handleReloadDocuments())
					admin.Post("/reload/{folderName}", s.handleReloadDocument())

					admin.Get("/backup", s.handleListBackups())
					admin.Get("/backup/download/{fileName}", s.handleDownloadBackup())
					admin.Post("/backup", s.handleCreateBackup())
					admin.Post("/backup/restore", s.handleRestoreBackup())
				})
			})

			r.Route("/jobs", func(r chi.Router) {
				r.Get("/{jobID}", s.handleGetJobStatus())
			})

			r.Route("/inhabitants", func(r chi.Router) {
				r.Get("/", s.handleListInhabitants())
				r.Post("/", s.handleCreateInhabitant())
				r.Get("/{id}", s.handleGetInhabitant())
				r.Put("/{id}", s.handleUpdateInhabitant())
				r.Delete("/{id}", s.handleDeleteInhabitant())
			})

			r.Route("/search", func(r chi.Router) {
				r.Get("/advanced", s.handleAdvancedSearch())
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

func (s *Server) Shutdown(ctx context.Context) error {
	s.cancel()
	slog.Info("Server context cancelled, cleaning up...")
	return nil
}
