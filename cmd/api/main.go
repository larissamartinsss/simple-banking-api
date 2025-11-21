package main

import (
	"log"
	"os"
)

func main() {
	// Load configuration
	config := LoadConfig()

	// Create logger
	logger := log.New(os.Stdout, "[SIMPLE-BANKING-API] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting Simple Banking API...")

	// Initialize application
	app, err := NewApplication(config)
	if err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Shutdown()

	// Start server (blocks until shutdown signal)
	if err := app.Start(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
