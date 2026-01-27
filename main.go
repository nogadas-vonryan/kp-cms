package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"kpcms/server/core/api"
	"kpcms/server/core/auth"
	"kpcms/server/core/document"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:panel/dist
var panelAssets embed.FS

//go:embed all:frontend/dist
var frontendAssets embed.FS

// Helper function to check if error message contains substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type Logger struct {
	ctx        context.Context
	logChannel chan string
}

func NewLogger(ctx context.Context) *Logger {
	return &Logger{
		ctx:        ctx,
		logChannel: make(chan string),
	}
}

func (l *Logger) Start() {
	go func() {
		for msg := range l.logChannel {
			// Send log message to frontend
			runtime.EventsEmit(l.ctx, "log", msg)
		}
	}()
}

func (l *Logger) Log(msg string) {
	l.logChannel <- msg
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "KPCMS",
		Width:         440,
		Height:        620,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: panelAssets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func StartWebServer(host string, port int) (*http.Server, error) {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == 0 {
		port = 8081
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port number: %d (must be between 1-65535)", port)
	}

	staticPath := "frontend/dist"

	distFS, err := fs.Sub(frontendAssets, staticPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load frontend assets: %v", err)
	}

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.FS(distFS))

	// API Routes
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status": "online"}`)
	})

	// SPA Handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		f, err := distFS.Open(path)

		// If path is empty (root) or file doesn't exist (SPA route)
		if err != nil || path == "" {
			data, readErr := fs.ReadFile(distFS, "index.html")
			if readErr != nil {
				http.Error(w, "index file not found", http.StatusInternalServerError)
				return
			}
			http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(data))
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	// Test if we can bind to the address before starting the server
	// This catches "address already in use" and permission errors early

	// Try to create a test listener to validate the address
	testListener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		if contains(err.Error(), "address already in use") {
			return nil, fmt.Errorf("port %d is already in use", port)
		} else if contains(err.Error(), "permission denied") {
			return nil, fmt.Errorf("permission denied: cannot bind to port %d (ports below 1024 require elevated privileges)", port)
		}
		return nil, fmt.Errorf("failed to bind to %s:%d: %v", host, port, err)
	}
	testListener.Close()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		}
	}()

	return srv, nil
}

func GeneratePasswordIfEmpty(pass string) (string, error) {
	if pass == "" {
		generatedPass, err := auth.GenerateSecurePassword(12)
		if err != nil {
			return "", fmt.Errorf("failed to generate password: %v", err)
		}
		return generatedPass, nil
	}
	return pass, nil
}

func StartBackendServer(host string, port int, user, pass, dataPath string) (*http.Server, error) {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == 0 {
		port = 8080
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port number: %d (must be between 1-65535)", port)
	}

	// Set default data path
	if dataPath == "" {
		dataPath = "./data"
	}

	// Create document repository
	namingStrategy := document.NewNamingStrategyPrefixDDDYY("case")
	documentRepository, err := document.NewFileDocumentRepository(dataPath, namingStrategy)
	if err != nil {
		return nil, fmt.Errorf("failed to create document repository: %v", err)
	}

	// Create API server
	apiServer, err := api.NewServer(
		host,
		fmt.Sprintf("%d", port),
		user,
		pass,
		document.NewDocumentService(documentRepository),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up server: %v", err)
	}

	// Test if we can bind to the address before starting the server
	addr := fmt.Sprintf("%s:%d", host, port)
	testListener, err := net.Listen("tcp", addr)
	if err != nil {
		if contains(err.Error(), "address already in use") {
			return nil, fmt.Errorf("port %d is already in use", port)
		} else if contains(err.Error(), "permission denied") {
			return nil, fmt.Errorf("permission denied: cannot bind to port %d (ports below 1024 require elevated privileges)", port)
		}
		return nil, fmt.Errorf("failed to bind to %s:%d: %v", host, port, err)
	}
	testListener.Close()

	srv := &http.Server{
		Addr:    addr,
		Handler: apiServer.Router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		}
	}()

	return srv, nil
}
