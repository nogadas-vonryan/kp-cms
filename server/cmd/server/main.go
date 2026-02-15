package main

import (
	"flag"
	"kpcms/server/core/api"
	"kpcms/server/core/auth"
	"kpcms/server/core/calendar"
	"kpcms/server/core/database"
	"kpcms/server/core/document"
	"kpcms/server/core/document/store"
	"kpcms/server/core/inhabitant"
	"kpcms/server/core/oauth"
	"kpcms/server/core/search"
	"kpcms/server/core/web"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	host := flag.String("host", "0.0.0.0", "Host address to bind the server to")
	port := flag.String("port", "8080", "Port to run the backend server on")
	frontendPort := flag.String("frontend-port", "8081", "Port to run the frontend web server on")
	user := flag.String("admin", "admin", "Admin username")
	pass := flag.String("pass", "", "Admin password")
	dataPath := flag.String("data", "./data", "Path to the data directory")
	backupPath := flag.String("backup", "./backup", "Path to the backup directory")
	frontendPath := flag.String("frontend", "./frontend/dist", "Path to the frontend build directory")
	flag.Parse()

	if *pass == "" {
		generatedPass, err := auth.GenerateSecurePassword(12)
		if err != nil {
			log.Fatal("Failed to generate password", err)
		}
		*pass = generatedPass
	}

	namingStrategy := document.NewNamingStrategyPrefixDDDYY("case")
	documentRepository, err := store.New(*dataPath, *backupPath, namingStrategy)
	if err != nil {
		log.Fatalf("Failed to create document repository: %v", err)
	}

	// Initialize app database
	appDBPath := filepath.Join(*dataPath, "app.db")
	if err := os.MkdirAll(filepath.Dir(appDBPath), 0700); err != nil {
		log.Fatalf("Failed to create app database directory: %v", err)
	}

	// Get platform-specific auth database path
	authDBPath, err := database.GetAuthDBPath()
	if err != nil {
		log.Fatalf("Failed to get auth database path: %v", err)
	}

	db, err := database.New(appDBPath, authDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	authRepo := auth.NewSQLRepository(db.AuthDB)
	inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
	authService := auth.NewService(authRepo)
	inhabitantService := inhabitant.NewService(inhabitantRepo)
	documentService := document.NewDocumentService(documentRepository, documentRepository, documentRepository, documentRepository)

	// Create the aggregator service for cross-domain searches
	searchService := search.NewAggregator(inhabitantService, documentService)

	// Create OAuth service (optional)
	// OAuth callback goes through frontend (port 8081) which proxies to backend
	oauthService, err := oauth.NewService(db.AuthDB, oauth.Config{
		Google: oauth.GoogleConfig{
			CallbackURL: "http://localhost:" + *frontendPort + "/oauth/callback",
		},
	})
	if err != nil {
		log.Printf("Warning: OAuth service not initialized: %v", err)
		log.Println("Calendar features will be disabled. To enable, follow CALENDAR_SETUP.md")
	}

	// Create calendar service using OAuth service
	var calendarService *calendar.Service
	if oauthService != nil {
		calendarService = calendar.New(oauthService, "kpcms-calendar")
	}

	server, err := api.NewServer(
		*host,
		*port,
		*user,
		*pass,
		documentService,
		authService,
		inhabitantService,
		searchService,
		oauthService,
		calendarService,
		db.AppDB,
		appDBPath,
	)
	if err != nil {
		log.Fatalf("Failed to set up server: %v", err)
	}

	// Start backend server in a goroutine
	go func() {
		log.Printf("Backend server starting on %s", server.Addr())
		if err := http.ListenAndServe(server.Addr(), server.Router); err != nil {
			log.Fatalf("Failed to start backend server: %v", err)
		}
	}()

	// Start frontend web server
	webCfg := web.Config{
		FrontendHost:      *host,
		FrontendPort:      0, // Will parse from string
		BackendHost:       *host,
		BackendPort:       0, // Will parse from string
		UseEmbeddedAssets: false,
		FrontendPath:      *frontendPath,
	}

	// Parse ports
	if fp, err := strconv.Atoi(*frontendPort); err == nil {
		webCfg.FrontendPort = fp
	}
	if bp, err := strconv.Atoi(*port); err == nil {
		webCfg.BackendPort = bp
	}

	webServer, err := web.NewServer(webCfg)
	if err != nil {
		log.Fatalf("Failed to create web server: %v", err)
	}

	if err := webServer.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}

	log.Printf("Using data directory: %s", *dataPath)
	log.Printf("Password generated: %s", *pass)
	log.Printf("Access the application at http://localhost:%s", *frontendPort)

	// Keep main goroutine alive
	select {}
}
