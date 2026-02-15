package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"kpcms/server/core/api"
	"kpcms/server/core/auth"
	"kpcms/server/core/calendar"
	"kpcms/server/core/database"
	"kpcms/server/core/document"
	"kpcms/server/core/document/store"
	"kpcms/server/core/inhabitant"
	"kpcms/server/core/oauth"
	"kpcms/server/core/search"

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

var appLogger *Logger

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
	log.Print(msg)
	l.logChannel <- msg
}

func SetLogger(l *Logger) {
	appLogger = l
}

func logMessage(msg string) {
	if appLogger != nil {
		appLogger.Log(msg)
		return
	}
	log.Print(msg)
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "KPCMS",
		Width:         460,
		Height:        540,
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
		logMessage(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}

// GetPreferredIP returns the local IP address the host uses to communicate
// with the rest of the network.
func GetPreferredIP() (string, error) {
	// Use a public address to find the active routing interface.
	conn, err := net.Dial("udp4", "google.com:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String(), nil
	}

	// Scan interfaces (If offline or dial fails)
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// Filter out interfaces that are down or are loopbacks
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip IPv6 and Loopback for compatibility
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}

			// Filter out common "useless" virtual IPs
			ipStr := ip.String()
			if strings.HasPrefix(ipStr, "169.254") { // APIPA (No connection)
				continue
			}

			return ipStr, nil
		}
	}

	return "", fmt.Errorf("could not determine local IP")
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

func StartBackendServer(host string, port int, user, pass, dataPath string, backupPath string) (*http.Server, error) {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == 0 {
		port = 8080
	}

	api.SetLogSink(logMessage)

	// Validate port range
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port number: %d (must be between 1-65535)", port)
	}

	// Set default data path
	if dataPath == "" {
		dataPath = "./data"
	}

	if backupPath == "" {
		backupPath = "./backup"
	}

	// Create document repository
	namingStrategy := document.NewNamingStrategyPrefixDDDYY("case")
	documentRepository, err := store.New(dataPath, backupPath, namingStrategy)
	if err != nil {
		return nil, fmt.Errorf("failed to create document repository: %v", err)
	}

	// Initialize app database
	appDBPath := filepath.Join(dataPath, "app.db")
	if err := os.MkdirAll(filepath.Dir(appDBPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Get platform-specific auth database path
	authDBPath, err := database.GetAuthDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth database path: %v", err)
	}

	db, err := database.New(appDBPath, authDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	// Create repositories
	authRepo := auth.NewSQLRepository(db.AuthDB)
	inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)

	// Create services
	authService := auth.NewService(authRepo)
	inhabitantService := inhabitant.NewService(inhabitantRepo)

	// Create search aggregator service
	searchService := search.NewAggregator(inhabitantService, document.NewDocumentService(documentRepository, documentRepository, documentRepository, documentRepository))

	// Create OAuth service (optional - can be nil if credentials not configured)
	// OAuth callback goes through frontend (port 8081) which proxies to backend
	oauthService, err := oauth.NewService(db.AuthDB, oauth.Config{
		Google: oauth.GoogleConfig{
			CallbackURL: "http://localhost:8081/oauth/callback",
		},
	})
	if err != nil {
		logMessage(fmt.Sprintf("Warning: OAuth service not initialized: %v", err))
		logMessage("Calendar features will be disabled. To enable, follow CALENDAR_SETUP.md")
	}

	// Create calendar service using OAuth service
	var calendarService *calendar.Service
	if oauthService != nil {
		calendarService = calendar.New(oauthService, "kpcms-calendar")
	}

	// Create API server
	apiServer, err := api.NewServer(
		host,
		fmt.Sprintf("%d", port),
		user,
		pass,
		document.NewDocumentService(documentRepository, documentRepository, documentRepository, documentRepository),
		authService,
		inhabitantService,
		searchService,
		oauthService,
		calendarService,
		db.AppDB,
		appDBPath,
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
	srv.RegisterOnShutdown(func() {
		_ = db.Close()
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logMessage(fmt.Sprintf("Backend server error: %v", err))
		}
	}()

	return srv, nil
}
