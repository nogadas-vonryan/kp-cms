package web

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// EmbedFS is an interface for embedded filesystems
type EmbedFS interface {
	fs.FS
}

// Logger is a simple logger interface
type Logger interface {
	Log(msg string)
}

// defaultLogger writes to stdout
type defaultLogger struct{}

func (l defaultLogger) Log(msg string) {
	fmt.Println(msg)
}

// Server handles frontend serving and reverse proxy to backend
type Server struct {
	frontendPort int
	backendPort  int
	frontendFS   fs.FS
	logger       Logger
	server       *http.Server
}

// Config holds the web server configuration
type Config struct {
	FrontendHost string
	FrontendPort int
	BackendHost  string
	BackendPort  int
	// UseEmbeddedAssets determines whether to use embedded filesystem or disk
	UseEmbeddedAssets bool
	// EmbeddedFS is the embedded filesystem containing frontend assets (only used if UseEmbeddedAssets is true)
	EmbeddedFS fs.FS
	// FrontendPath is the path within EmbeddedFS or disk where frontend files are located
	FrontendPath string
	// Logger is optional; if nil, logs to stdout
	Logger Logger
}

// NewServer creates a new web server
func NewServer(cfg Config) (*Server, error) {
	if cfg.FrontendHost == "" {
		cfg.FrontendHost = "0.0.0.0"
	}
	if cfg.FrontendPort == 0 {
		cfg.FrontendPort = 8081
	}
	if cfg.BackendHost == "" {
		cfg.BackendHost = "127.0.0.1"
	}
	if cfg.BackendPort == 0 {
		cfg.BackendPort = 8080
	}
	if cfg.FrontendPath == "" {
		cfg.FrontendPath = "frontend/dist"
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger{}
	}

	// Validate ports
	if cfg.FrontendPort < 1 || cfg.FrontendPort > 65535 {
		return nil, fmt.Errorf("invalid frontend port %d (must be 1-65535)", cfg.FrontendPort)
	}
	if cfg.BackendPort < 1 || cfg.BackendPort > 65535 {
		return nil, fmt.Errorf("invalid backend port %d (must be 1-65535)", cfg.BackendPort)
	}

	s := &Server{
		frontendPort: cfg.FrontendPort,
		backendPort:  cfg.BackendPort,
		logger:       cfg.Logger,
	}

	// Setup frontend filesystem
	var distFS fs.FS
	var err error

	if cfg.UseEmbeddedAssets && cfg.EmbeddedFS != nil {
		// Use embedded filesystem
		distFS, err = fs.Sub(cfg.EmbeddedFS, cfg.FrontendPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load embedded frontend assets: %v", err)
		}
	} else {
		// Use filesystem
		if _, err := os.Stat(cfg.FrontendPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("frontend path does not exist: %s", cfg.FrontendPath)
		}
		distFS = os.DirFS(cfg.FrontendPath)
	}

	s.frontendFS = distFS

	// Setup reverse proxy to backend
	backendAddr := fmt.Sprintf("http://%s:%d", cfg.BackendHost, cfg.BackendPort)
	target, err := url.Parse(backendAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid backend address: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		r.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Header.Set("X-Forwarded-For", r.RemoteAddr)
	}

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.FS(distFS))

	// Smart catch-all handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// API paths and OAuth callback go to backend
		if s.shouldProxyToBackend(path) {
			proxy.ServeHTTP(w, r)
			return
		}

		// Try to serve static file
		if s.serveStaticFile(w, r, path, fileServer) {
			return
		}

		// SPA fallback: serve index.html
		s.serveSPA(w, r, proxy)
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.FrontendHost, cfg.FrontendPort),
		Handler: mux,
	}

	// Validate address binding
	testListener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to address %s: %v", s.server.Addr, err)
	}
	testListener.Close()

	return s, nil
}

// shouldProxyToBackend returns true if the path should be proxied to backend
func (s *Server) shouldProxyToBackend(path string) bool {
	// API endpoints
	if strings.HasPrefix(path, "api/") || strings.HasPrefix(path, "auth/") {
		return true
	}
	// OAuth callback
	if path == "oauth/callback" {
		return true
	}
	return false
}

// serveStaticFile attempts to serve a static file, returns true if successful
func (s *Server) serveStaticFile(w http.ResponseWriter, r *http.Request, path string, fileServer http.Handler) bool {
	// Try to open the file
	f, err := s.frontendFS.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Check if it's a file (not a directory)
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}

	// Serve the file
	fileServer.ServeHTTP(w, r)
	return true
}

// serveSPA serves the SPA index.html
func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy) {
	// Try to read index.html
	f, err := s.frontendFS.Open("index.html")
	if err != nil {
		// If we can't find index.html, maybe the user is calling an API
		// we didn't explicitly list? Try proxying as a last resort.
		proxy.ServeHTTP(w, r)
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(data))
}

// Start starts the web server in a goroutine
func (s *Server) Start() error {
	go func() {
		s.logger.Log(fmt.Sprintf("Web server starting on http://%s", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Log(fmt.Sprintf("Web server error: %v", err))
		}
	}()
	return nil
}

// Stop gracefully stops the web server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Addr returns the server address
func (s *Server) Addr() string {
	if s.server != nil {
		return s.server.Addr
	}
	return ""
}

// GetPublicURL returns the public URL for OAuth callbacks
func (s *Server) GetPublicURL() string {
	return fmt.Sprintf("http://localhost:%d", s.frontendPort)
}
