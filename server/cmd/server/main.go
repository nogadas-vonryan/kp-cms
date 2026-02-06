package main

import (
	"flag"
	"kpcms/server/core/api"
	"kpcms/server/core/auth"
	"kpcms/server/core/database"
	"kpcms/server/core/document"
	"kpcms/server/core/document/store"
	"kpcms/server/core/inhabitant"
	inhabitantstore "kpcms/server/core/inhabitant/store"
	"kpcms/server/core/search"
	"log"
	"os"
	"path/filepath"

	"net/http"
)

func main() {
	host := flag.String("host", "0.0.0.0", "Host address to bind the server to")
	port := flag.String("port", "8080", "Port to run the server on")
	user := flag.String("admin", "admin", "Admin username")
	pass := flag.String("pass", "", "Admin password")
	dataPath := flag.String("data", "./data", "Path to the data directory")
	backupPath := flag.String("backup", "./backup", "Path to the backup directory")
	fileInhabitants := flag.Bool("file-inhabitants", false, "Use file-based inhabitant storage instead of SQLite")
	flag.Parse()

	if *pass == "" {
		generatedPass, err := auth.GenerateSecurePassword(12)
		if err != nil {
			log.Fatal("Failed to generate password", err)
		}
		*pass = generatedPass
	}

	namingStrategy := document.NewNamingStrategyPrefixDDDYY("case")
	casesPath := filepath.Join(*dataPath, "cases")
	documentRepository, err := store.New(casesPath, *backupPath, namingStrategy)
	if err != nil {
		log.Fatalf("Failed to create document repository: %v", err)
	}
	// Set backup source to parent data path for full data backup (all entity types)
	documentRepository.SetBackupSource(*dataPath)

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
	authService := auth.NewService(authRepo)

	var inhabitantService *inhabitant.Service
	if *fileInhabitants {
		log.Printf("Using file-based inhabitant storage")
		inhabPath := filepath.Join(*dataPath, "inhabitants")
		namingStrategy := document.NewNamingStrategyPrefixDDDYY("inhabitant")
		inhabitantFileStore, err := inhabitantstore.New(inhabPath, namingStrategy)
		if err != nil {
			log.Fatalf("Failed to create inhabitant file store: %v", err)
		}
		inhabitantService = inhabitant.NewServiceWithStore(inhabitantFileStore)
	} else {
		log.Printf("Using SQLite-based inhabitant storage")
		inhabitantRepo := inhabitant.NewSQLRepository(db.AppDB)
		inhabitantService = inhabitant.NewService(inhabitantRepo)
	}

	documentService := document.NewDocumentService(documentRepository, documentRepository, documentRepository, documentRepository)

	// Create the aggregator service for cross-domain searches
	searchService := search.NewAggregator(inhabitantService, documentService)

	server, err := api.NewServer(
		*host,
		*port,
		*user,
		*pass,
		documentService,
		authService,
		inhabitantService,
		searchService,
	)
	if err != nil {
		log.Fatalf("Failed to set up server: %v", err)
	}

	log.Printf("Server starting on %s", server.Addr())
	log.Printf("Using data directory: %s", *dataPath)
	log.Printf("Password generated: %s", *pass)

	if err := http.ListenAndServe(server.Addr(), server.Router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
