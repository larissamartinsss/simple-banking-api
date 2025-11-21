package main

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ServerAddress string
	DatabasePath  string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() Config {
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8080"
	}

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "./data/banking.db"
	}

	return Config{
		ServerAddress: serverAddress,
		DatabasePath:  databasePath,
	}
}
