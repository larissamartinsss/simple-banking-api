package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/larissamartinsss/simple-banking-api/internal/server/handlers"
)

type Server struct {
	router                   *chi.Mux
	createAccountHandler     *handlers.CreateAccountHandler
	getAccountHandler        *handlers.GetAccountHandler
	createTransactionHandler *handlers.CreateTransactionHandler
}

func NewServer(createAccountHandler *handlers.CreateAccountHandler, getAccountHandler *handlers.GetAccountHandler, createTransactionHandler *handlers.CreateTransactionHandler) *Server {
	s := &Server{
		router:                   chi.NewRouter(),
		createAccountHandler:     createAccountHandler,
		getAccountHandler:        getAccountHandler,
		createTransactionHandler: createTransactionHandler,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
	s.router.Use(middleware.SetHeader("Content-Type", "application/json"))
}

// setupRoutes configures all RESTful routes
func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", s.createAccountHandler.Handle)
			r.Get("/{accountId}", s.getAccountHandler.Handle)
		})

		r.Route("/transactions", func(r chi.Router) {
			r.Post("/", s.createTransactionHandler.Handle)
		})
	})
}

func (s *Server) GetRouter() http.Handler {
	return s.router
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
