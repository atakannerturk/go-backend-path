package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atakannerturk/go-backend-path/internal/config"
	"github.com/atakannerturk/go-backend-path/internal/handler"
	"github.com/atakannerturk/go-backend-path/internal/repository/mysql"
	"github.com/atakannerturk/go-backend-path/internal/service"
	"github.com/atakannerturk/go-backend-path/internal/worker"
	"github.com/atakannerturk/go-backend-path/migrations"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load .env file if it exists (development)
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger.Init(cfg.App.LogLevel, cfg.App.Env)
	logger.Info("Starting application", "env", cfg.App.Env)

	db, err := mysql.NewDB(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	if err := migrations.RunMigrations(db, "./migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Migrations completed successfully")

	userRepo := mysql.NewUserRepository(db)
	txRepo := mysql.NewTransactionRepository(db)
	balanceRepo := mysql.NewBalanceRepository(db)

	workerPool := worker.NewWorkerPool(cfg.Worker.PoolSize, cfg.Worker.TransactionQueueSize)
	workerPool.Start()

	balanceService := service.NewBalanceService(balanceRepo)
	userService := service.NewUserService(userRepo, cfg.Security.JWTSecret)
	txService := service.NewTransactionService(txRepo, balanceService, workerPool)

	logger.Info("Services initialized successfully")

	server := handler.NewServer(userService, txService, balanceService)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      server.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", "host", cfg.Server.Host, "port", cfg.Server.Port)
		serverErrors <- httpServer.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		logger.Info("Shutting down server", "signal", sig.String())

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			httpServer.Close()
			logger.Error("Server shutdown error", "error", err)
		}

		done := make(chan struct{})
		go func() {
			workerPool.Stop()
			close(done)
		}()

		select {
		case <-done:
			logger.Info("Worker pool stopped gracefully")
		case <-shutdownCtx.Done():
			logger.Warn("Worker pool shutdown timeout exceeded")
		}

		total, success, failed := workerPool.Stats()
		logger.Info("Final worker pool stats",
			"total_tasks", total,
			"successful_tasks", success,
			"failed_tasks", failed)
	}

	logger.Info("Server stopped gracefully")
	return nil
}
