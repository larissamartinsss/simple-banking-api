package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larissamartinsss/simple-banking-api/infra/database"
	"github.com/larissamartinsss/simple-banking-api/internal/adapters/repository"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors"
	"github.com/larissamartinsss/simple-banking-api/internal/server"
	"github.com/larissamartinsss/simple-banking-api/internal/server/handlers"
)

func main() {
	// Configuration
	config := loadConfig()

	// Initialize logger
	logger := log.New(os.Stdout, "[SIMPLE-BANKING-API] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting Simple Banking API...")

	// Initialize database
	logger.Println("Connecting to database...")
	db, err := database.NewConnection(database.Config{
		DatabasePath: config.DatabasePath,
	})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)
	logger.Println("Database connected successfully")

	// Run migrations
	logger.Println("Running database migrations...")
	ctx := context.Background()
	if err := database.RunMigrations(ctx, db); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Println("Migrations completed successfully")

	// Initialize repositories (Adapters)
	accountRepo := repository.NewAccountRepository(db)
	operationTypeRepo := repository.NewOperationTypeRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Seed operation types
	logger.Println("Seeding operation types...")
	if err := operationTypeRepo.Seed(ctx); err != nil {
		logger.Fatalf("Failed to seed operation types: %v", err)
	}

	// Initialize processors
	createAccountProcessor := processors.NewCreateAccountProcessor(accountRepo)
	getAccountProcessor := processors.NewGetAccountProcessor(accountRepo)
	createTransactionProcessor := processors.NewCreateTransactionProcessor(transactionRepo, accountRepo, operationTypeRepo)

	// Initialize handlers (HTTP Layer)
	createAccountHandler := handlers.NewCreateAccountHandler(createAccountProcessor)
	getAccountHandler := handlers.NewGetAccountHandler(getAccountProcessor)
	createTransactionHandler := handlers.NewCreateTransactionHandler(createTransactionProcessor)

	// Initialize server (Router)
	srv := server.NewServer(createAccountHandler, getAccountHandler, createTransactionHandler)

	// Create HTTP server with timeouts
	httpServer := &http.Server{
		Addr:         config.ServerAddress,
		Handler:      srv.GetRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("🌐 Server starting on %s", config.ServerAddress)
		logger.Println("📋 Available endpoints:")
		logger.Println("   POST   /api/v1/accounts")
		logger.Println("   GET    /api/v1/accounts/{accountId}")
		logger.Println("   POST   /api/v1/transactions")
		logger.Println("   GET    /health")
		logger.Println("")
		logger.Println("✨ Server is ready to accept requests!")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	logger.Println("✅ Server exited gracefully")
}

// Config holds application configuration
type Config struct {
	ServerAddress string
	DatabasePath  string
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
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
