package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Host   string
	Port   string
	Router *chi.Mux
}

func NewServer(host, port string) *Server {
	s := &Server{
		Host:   host,
		Port:   port,
		Router: chi.NewRouter(),
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
	s.Router.Post("/test", s.handleTest())
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
