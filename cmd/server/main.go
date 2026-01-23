package main

import (
	"flag"
	"log"
	"main/internal/api"
	"main/internal/auth"
	"main/internal/document"

	"net/http"
)

func main() {
	host := flag.String("host", "0.0.0.0", "Host address to bind the server to")
	port := flag.String("port", "8080", "Port to run the server on")
	user := flag.String("admin", "admin", "Admin username")
	pass := flag.String("pass", "", "Admin password")
	dataPath := flag.String("data", "./data", "Path to the data directory")
	flag.Parse()

	if *pass == "" {
		generatedPass, err := auth.GenerateSecurePassword(8)
		if err != nil {
			log.Fatal("Failed to generate password", err)
		}
		*pass = generatedPass
	}

	documentRepository, err := document.NewFileDocumentRepository(*dataPath)
	if err != nil {
		log.Fatalf("Failed to create document repository: %v", err)
	}

	server := api.NewServer(
		*host,
		*port,
		*user,
		*pass,
		document.NewDocumentService(documentRepository),
	)

	log.Printf("Server starting on %s", server.Addr())
	log.Printf("Using data directory: %s", *dataPath)
	log.Printf("Password generated: %s", *pass)

	// Create the admin account
	auth.AddUser(*user, *pass, auth.RoleAdmin)

	if err := http.ListenAndServe(server.Addr(), server.Router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
