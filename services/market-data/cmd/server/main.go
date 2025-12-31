package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/pkg/logger"
	httpAdapter "github.com/minwook/battery-optimization/services/market-data/internal/adapters/http"
	postgresAdapter "github.com/minwook/battery-optimization/services/market-data/internal/adapters/postgres"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	NatsURL     string
	Port        string
	LogLevel    string
}

func main() {
	// 1. Load configuration from environment
	cfg := loadConfig()

	// 2. Setup structured logging
	log, err := logger.NewFromEnv("market-data", cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("starting service",
		zap.String("port", cfg.Port),
		zap.String("log_level", cfg.LogLevel),
	)

	// 3. Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping database", zap.Error(err))
	}
	log.Info("database connection established")

	// 4. Run migrations using golang-migrate
	if err := runMigrations(db, cfg.DatabaseURL, log); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// 5. Create repository
	repo := postgresAdapter.NewPostgresRepository(db)
	log.Info("repository initialized")

	// 6. Connect to NATS (optional - service works without events)
	var publisher events.EventPublisher
	if cfg.NatsURL != "" {
		natsPublisher, err := events.NewNATSPublisher(cfg.NatsURL)
		if err != nil {
			log.Warn("failed to connect to NATS",
				zap.String("nats_url", cfg.NatsURL),
				zap.Error(err),
			)
			log.Info("service will continue without event publishing")
		} else {
			publisher = natsPublisher
			defer publisher.Close()
			log.Info("NATS publisher connected", zap.String("nats_url", cfg.NatsURL))
		}
	} else {
		log.Info("NATS_URL not set - running without event publishing")
	}

	// 7. Create handlers
	handler := httpAdapter.NewMarketPriceHandler(repo, publisher)
	log.Info("HTTP handler initialized")

	healthHandler := httpAdapter.NewHealthHandler(db, nil)
	log.Info("health handler initialized")

	// 8. Setup HTTP router
	router := httpAdapter.SetupRoutes(handler, healthHandler)
	log.Info("routes configured")

	// 9. Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 10. Graceful shutdown on SIGINT/SIGTERM
	go func() {
		log.Info("starting HTTP server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exited gracefully")
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	cfg := Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://market_user:market_pass@localhost:5433/market_data?sslmode=disable"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		Port:        getEnv("PORT", "8081"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// runMigrations applies database migrations
func runMigrations(db *sql.DB, databaseURL string, log *zap.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/adapters/postgres/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("migrations applied successfully")
	return nil
}
