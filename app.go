package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartBackendServerResult contains the backend server start result
type StartBackendServerResult struct {
	Generated bool   `json:"generated"`
	Password  string `json:"password"`
}

// App struct
type App struct {
	ctx           context.Context
	server        *http.Server
	backendServer *http.Server
	logger        *Logger
	frontendPort  int
	backendPort   int
	config        *Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger = NewLogger(ctx)
	a.logger.Start()
	a.logger.Log("Application started")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		a.logger.Log(fmt.Sprintf("Warning: Failed to load config: %v", err))
		// Use default config
		config = &Config{
			FrontendHost: "0.0.0.0",
			FrontendPort: 8081,
			BackendHost:  "0.0.0.0",
			BackendPort:  8080,
			Username:     "admin",
			Password:     "",
			DataPath:     "./data",
		}
	} else {
		a.logger.Log("Configuration loaded successfully")
	}
	a.config = config
}

func (a *App) SelectFolder() (string, error) {
	a.Log("Opening folder selection dialog...")
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Installation Folder",
		DefaultDirectory: "",
	})

	if err != nil {
		a.Log(fmt.Sprintf("Error selecting folder: %v", err))
		return "", err
	}

	if selection != "" {
		a.Log(fmt.Sprintf("Folder selected: %s", selection))
	} else {
		a.Log("Folder selection cancelled")
	}
	// selection will be an empty string if the user cancels
	return selection, nil
}

func (a *App) GetNetworkInfo() map[string]interface{} {
	localIP, err := GetLocalIP()
	if err != nil {
		localIP = ""
	}

	return map[string]interface{}{
		"localIP":      localIP,
		"frontendPort": a.frontendPort,
		"backendPort":  a.backendPort,
	}
}

func (a *App) GetExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func (a *App) StartWebServer(host string, port int, backendHost string, backendPort int) error {
	a.Log(fmt.Sprintf("Starting frontend web server on %s:%d...", host, port))
	srv, err := StartWebServer(host, port, backendHost, backendPort)
	if err != nil {
		a.Log(fmt.Sprintf("Failed to start frontend server: %v", err))
		return err
	}
	// Store server reference for shutdown later
	a.server = srv
	a.frontendPort = port
	a.Log(fmt.Sprintf("Frontend server started successfully on http://%s:%d", host, port))
	return nil
}

func (a *App) StopWebServer() error {
	if a.server == nil {
		return fmt.Errorf("no server is currently running")
	}

	a.Log("Stopping frontend web server...")
	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.Log(fmt.Sprintf("Error stopping frontend server: %v", err))
		return fmt.Errorf("failed to shutdown server: %v", err)
	}

	a.server = nil
	a.Log("Frontend server stopped successfully")
	return nil
}

func (a *App) StartBackendServer(host string, port int, user, pass, dataPath string) (*StartBackendServerResult, error) {
	a.Log(fmt.Sprintf("Starting backend server on %s:%d with data path: %s...", host, port, dataPath))

	result := &StartBackendServerResult{
		Generated: false,
		Password:  "",
	}

	// Generate password if not provided and log it
	generatedPass, err := GeneratePasswordIfEmpty(pass)
	if err != nil {
		a.Log(fmt.Sprintf("Failed to generate password: %v", err))
		return nil, err
	}
	if pass == "" {
		a.Log(fmt.Sprintf("Generated admin password: %s", generatedPass))
		pass = generatedPass
		result.Generated = true
		result.Password = generatedPass
	}

	srv, err := StartBackendServer(host, port, user, pass, dataPath)
	if err != nil {
		a.Log(fmt.Sprintf("Failed to start backend server: %v", err))
		return nil, err
	}
	// Store server reference for shutdown later
	a.backendServer = srv
	a.backendPort = port
	a.Log(fmt.Sprintf("Backend server started successfully on http://%s:%d", host, port))
	a.Log(fmt.Sprintf("Using data directory: %s", dataPath))
	a.Log(fmt.Sprintf("Admin user: %s", user))

	// If password was generated, auto-save it to config
	if result.Generated {
		updatedConfig := *a.config
		updatedConfig.Password = pass
		if err := SaveConfig(&updatedConfig); err != nil {
			a.Log(fmt.Sprintf("Warning: Failed to auto-save generated password: %v", err))
		} else {
			a.config = &updatedConfig
			a.Log("Generated password auto-saved to configuration")
		}
	}

	// Log network access info
	localIP, err := GetLocalIP()
	if err == nil && localIP != "" {
		a.Log(fmt.Sprintf("Network access: http://%s:%d", localIP, port))
	}
	return result, nil
}

// Log sends a log message to the frontend
func (a *App) Log(msg string) {
	if a.logger != nil {
		a.logger.Log(msg)
	}
}

func (a *App) StopBackendServer() error {
	if a.backendServer == nil {
		return fmt.Errorf("no backend server is currently running")
	}

	a.Log("Stopping backend server...")
	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.backendServer.Shutdown(ctx); err != nil {
		a.Log(fmt.Sprintf("Error stopping backend server: %v", err))
		return fmt.Errorf("failed to shutdown backend server: %v", err)
	}

	a.backendServer = nil
	a.Log("Backend server stopped successfully")
	return nil
}

// LoadSavedConfig returns the currently loaded configuration
func (a *App) LoadSavedConfig() (*Config, error) {
	if a.config == nil {
		return LoadConfig()
	}
	return a.config, nil
}

// SaveConfiguration saves the current configuration to disk
func (a *App) SaveConfiguration(frontendHost string, frontendPort int, backendHost string, backendPort int, username, password, dataPath string) error {
	config := &Config{
		FrontendHost: frontendHost,
		FrontendPort: frontendPort,
		BackendHost:  backendHost,
		BackendPort:  backendPort,
		Username:     username,
		Password:     password,
		DataPath:     dataPath,
	}

	if err := SaveConfig(config); err != nil {
		a.Log(fmt.Sprintf("Failed to save configuration: %v", err))
		return err
	}

	a.config = config
	a.Log("Configuration saved successfully")
	return nil
}

// GetConfigDirectory returns the config directory path for the current OS
func (a *App) GetConfigDirectory() (string, error) {
	return GetConfigDir()
}
