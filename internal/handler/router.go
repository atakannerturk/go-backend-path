package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/atakannerturk/go-backend-path/internal/middleware"
	"github.com/atakannerturk/go-backend-path/internal/service"
)

type Server struct {
	router          *chi.Mux
	userService     *service.UserService
	txService       *service.TransactionService
	balanceService  *service.BalanceService
	authHandler     *AuthHandler
	userHandler     *UserHandler
	txHandler       *TransactionHandler
	balanceHandler  *BalanceHandler
}

func NewServer(
	userService *service.UserService,
	txService *service.TransactionService,
	balanceService *service.BalanceService,
) *Server {
	s := &Server{
		router:         chi.NewRouter(),
		userService:    userService,
		txService:      txService,
		balanceService: balanceService,
	}

	s.authHandler = NewAuthHandler(userService)
	s.userHandler = NewUserHandler(userService)
	s.txHandler = NewTransactionHandler(txService)
	s.balanceHandler = NewBalanceHandler(balanceService)

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Recovery)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Metrics)

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s.router.Use(httprate.LimitByIP(100, 1*time.Minute))
}

func (s *Server) setupRoutes() {
	s.router.Get("/", s.healthCheck)
	s.router.Get("/health", s.healthCheck)
	s.router.Handle("/metrics", promhttp.Handler())

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.authHandler.Register)
			r.Post("/login", s.authHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(s.userService))
				r.Get("/me", s.authHandler.Me)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(middleware.Auth(s.userService))

			r.Get("/", s.userHandler.ListUsers)
			r.Get("/{id}", s.userHandler.GetUser)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Delete("/{id}", s.userHandler.DeleteUser)
			})
		})

		r.Route("/transactions", func(r chi.Router) {
			r.Use(middleware.Auth(s.userService))

			r.Post("/credit", s.txHandler.Credit)
			r.Post("/debit", s.txHandler.Debit)
			r.Post("/transfer", s.txHandler.Transfer)
			r.Get("/history", s.txHandler.GetHistory)
			r.Get("/{id}", s.txHandler.GetTransaction)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Get("/", s.txHandler.ListTransactions)
				r.Post("/{id}/rollback", s.txHandler.Rollback)
			})
		})

		r.Route("/balances", func(r chi.Router) {
			r.Use(middleware.Auth(s.userService))
			r.Get("/current", s.balanceHandler.GetCurrent)
		})
	})
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	middleware.RespondSuccess(w, http.StatusOK, map[string]string{
		"status": "ok",
		"version": "1.0.0",
	})
}

func (s *Server) Handler() http.Handler {
	return s.router
}
