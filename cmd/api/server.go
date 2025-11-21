package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larissamartinsss/simple-banking-api/internal/adapters/repository/accounts"
	"github.com/larissamartinsss/simple-banking-api/internal/adapters/repository/operationtype"
	"github.com/larissamartinsss/simple-banking-api/internal/adapters/repository/transactions"

	"github.com/larissamartinsss/simple-banking-api/infra/database"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors"
	"github.com/larissamartinsss/simple-banking-api/internal/server"
	"github.com/larissamartinsss/simple-banking-api/internal/server/handlers"
)

// Application holds all application dependencies
type Application struct {
	config Config
	logger *log.Logger
	db     *sql.DB
	server *server.Server
}

// NewApplication creates and initializes a new application instance
func NewApplication(config Config) (*Application, error) {
	app := &Application{
		config: config,
		logger: log.New(os.Stdout, "[SIMPLE-BANKING-API] ", log.LstdFlags|log.Lshortfile),
	}

	if err := app.initializeDatabase(); err != nil {
		return nil, err
	}

	if err := app.initializeDependencies(); err != nil {
		return nil, err
	}

	return app, nil
}

// initializeDatabase connects to database, runs migrations and seeds data
func (app *Application) initializeDatabase() error {
	app.logger.Println("Connecting to database...")
	db, err := database.NewConnection(database.Config{
		DatabasePath: app.config.DatabasePath,
	})
	if err != nil {
		return err
	}
	app.db = db
	app.logger.Println("Database connected successfully")

	// Run migrations
	app.logger.Println("Running database migrations...")
	ctx := context.Background()
	if err := database.RunMigrations(ctx, app.db); err != nil {
		return err
	}
	app.logger.Println("Migrations completed successfully")

	return nil
}

// initializeDependencies sets up the dependency injection chain
func (app *Application) initializeDependencies() error {
	ctx := context.Background()

	// Initialize repositories (Adapters Layer)
	accountRepo := accounts.NewAccountRepository(app.db)
	operationTypeRepo := operationtype.NewOperationTypeRepository(app.db)
	transactionRepo := transactions.NewTransactionRepository(app.db)

	// Seed operation types
	app.logger.Println("Seeding operation types...")
	if err := operationTypeRepo.Seed(ctx); err != nil {
		return err
	}

	// Initialize processors (Business Logic Layer)
	createAccountProcessor := processors.NewCreateAccountProcessor(accountRepo)
	getAccountProcessor := processors.NewGetAccountProcessor(accountRepo)
	createTransactionProcessor := processors.NewCreateTransactionProcessor(
		transactionRepo,
		accountRepo,
		operationTypeRepo,
	)
	getTransactionsProcessor := processors.NewGetTransactionsProcessor(
		transactionRepo,
		accountRepo,
	)

	// Initialize handlers (HTTP Layer)
	createAccountHandler := handlers.NewCreateAccountHandler(createAccountProcessor)
	getAccountHandler := handlers.NewGetAccountHandler(getAccountProcessor)
	createTransactionHandler := handlers.NewCreateTransactionHandler(createTransactionProcessor)
	getTransactionsHandler := handlers.NewGetTransactionsHandler(getTransactionsProcessor)

	// Initialize server (Router)
	app.server = server.NewServer(
		createAccountHandler,
		getAccountHandler,
		createTransactionHandler,
		getTransactionsHandler,
	)

	return nil
}

// Start starts the HTTP server and handles graceful shutdown
func (app *Application) Start() error {
	// Create HTTP server with timeouts
	httpServer := &http.Server{
		Addr:         app.config.ServerAddress,
		Handler:      app.server.GetRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		app.logger.Printf("🌐 Server starting on %s", app.config.ServerAddress)
		app.logger.Println("📋 Available endpoints:")
		app.logger.Println("   POST   /v1/accounts")
		app.logger.Println("   GET    /v1/accounts/{accountId}")
		app.logger.Println("   POST   /v1/transactions")
		app.logger.Println("   GET    /v1/accounts/{accountId}/transactions")
		app.logger.Println("   GET    /health")
		app.logger.Println("")
		app.logger.Println("✨ Server is ready to accept requests!")

		serverErrors <- httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case <-shutdown:
		app.logger.Println("🛑 Shutting down server...")

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			app.logger.Printf("❌ Could not gracefully shutdown: %v", err)
			return err
		}
	}

	app.logger.Println("✅ Server exited gracefully")
	return nil
}

// Shutdown closes all application resources
func (app *Application) Shutdown() {
	if app.db != nil {
		database.Close(app.db)
	}
}
