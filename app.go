package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	server        *http.Server
	backendServer *http.Server
	logger        *Logger
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

func (a *App) StartWebServer(host string, port int) error {
	a.Log(fmt.Sprintf("Starting frontend web server on %s:%d...", host, port))
	srv, err := StartWebServer(host, port)
	if err != nil {
		a.Log(fmt.Sprintf("Failed to start frontend server: %v", err))
		return err
	}
	// Store server reference for shutdown later
	a.server = srv
	a.Log(fmt.Sprintf("Frontend server started successfully on %s:%d", host, port))
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

func (a *App) StartBackendServer(host string, port int, user, pass, dataPath string) error {
	a.Log(fmt.Sprintf("Starting backend server on %s:%d with data path: %s...", host, port, dataPath))

	// Generate password if not provided and log it
	generatedPass, err := GeneratePasswordIfEmpty(pass)
	if err != nil {
		a.Log(fmt.Sprintf("Failed to generate password: %v", err))
		return err
	}
	if pass == "" {
		a.Log(fmt.Sprintf("Generated admin password: %s", generatedPass))
		pass = generatedPass
	}

	srv, err := StartBackendServer(host, port, user, pass, dataPath)
	if err != nil {
		a.Log(fmt.Sprintf("Failed to start backend server: %v", err))
		return err
	}
	// Store server reference for shutdown later
	a.backendServer = srv
	a.Log(fmt.Sprintf("Backend server started successfully on %s:%d", host, port))
	return nil
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
