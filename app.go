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
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectFolder() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Installation Folder",
		DefaultDirectory: "",
	})

	if err != nil {
		return "", err
	}

	// selection will be an empty string if the user cancels
	return selection, nil
}

func (a *App) StartWebServer(host string, port int) error {
	srv, err := StartWebServer(host, port)
	if err != nil {
		return err
	}
	// Store server reference for shutdown later
	a.server = srv
	return nil
}

func (a *App) StopWebServer() error {
	if a.server == nil {
		return fmt.Errorf("no server is currently running")
	}

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %v", err)
	}

	a.server = nil
	return nil
}

func (a *App) StartBackendServer(host string, port int, user, pass, dataPath string) error {
	srv, err := StartBackendServer(host, port, user, pass, dataPath)
	if err != nil {
		return err
	}
	// Store server reference for shutdown later
	a.backendServer = srv
	return nil
}

func (a *App) StopBackendServer() error {
	if a.backendServer == nil {
		return fmt.Errorf("no backend server is currently running")
	}

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.backendServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown backend server: %v", err)
	}

	a.backendServer = nil
	return nil
}
