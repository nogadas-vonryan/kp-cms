package main

import (
	"flag"
	"log"
	"main/internal/api"
	"main/internal/document"
	"net/http"
)

func main() {
	host := flag.String("host", "0.0.0.0", "Host address to bind the server to")
	port := flag.String("port", "8080", "Port to run the server on")
	dataPath := flag.String("data", "./data", "Path to the data directory")
	flag.Parse()

	documentRepository, err := document.NewFileDocumentRepository(*dataPath)
	if err != nil {
		log.Fatalf("Failed to create document repository: %v", err)
	}

	server := api.NewServer(
		*host,
		*port,
		document.NewDocumentService(documentRepository),
	)

	log.Printf("Server starting on %s", server.Addr())
	log.Printf("Using data directory: %s", *dataPath)

	if err := http.ListenAndServe(server.Addr(), server.Router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
