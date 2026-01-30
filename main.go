package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
		Height:        700,
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

func StartWebServer(host string, port int, backendHost string, backendPort int) (*http.Server, error) {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == 0 {
		port = 8081
	}
	if backendHost == "" {
		backendHost = "127.0.0.1"
	}
	if backendPort == 0 {
		backendPort = 8080
	}

	// 1. Setup Reverse Proxy
	backendAddr := fmt.Sprintf("http://%s:%d", backendHost, backendPort)
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

	// 2. Prepare Static Assets
	staticPath := "frontend/dist"
	distFS, err := fs.Sub(frontendAssets, staticPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load frontend assets: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	// 3. The "Smart" Catch-All Handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// A. If it's a known API prefix, proxy it immediately
		if strings.HasPrefix(path, "api/") || strings.HasPrefix(path, "auth/") {
			proxy.ServeHTTP(w, r)
			return
		}

		// B. Try to see if the file exists in the static dist folder (CSS, JS, Images)
		f, err := distFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// C. If it's not a file and not an explicit API path,
		// it's likely a frontend route (SPA). Serve index.html.
		data, readErr := fs.ReadFile(distFS, "index.html")
		if readErr != nil {
			// If we can't find index.html, maybe the user is calling an API
			// we didn't explicitly list? Try proxying as a last resort.
			proxy.ServeHTTP(w, r)
			return
		}

		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(data))
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	// Validate address binding
	testListener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, err
	}
	testListener.Close()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Web server error: %v\n", err)
		}
	}()

	return srv, nil
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no network interface found")
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
