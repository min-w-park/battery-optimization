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
	httpAdapter "github.com/minwook/battery-optimization/services/telemetry/internal/adapters/http"
	postgresAdapter "github.com/minwook/battery-optimization/services/telemetry/internal/adapters/postgres"
	"github.com/minwook/battery-optimization/services/telemetry/internal/service"
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
	log, err := logger.NewFromEnv("telemetry", cfg.LogLevel)
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

	// 4. Run migrations
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
	handler := httpAdapter.NewTelemetryHandler(repo, log)
	log.Info("HTTP handler initialized")

	healthHandler := httpAdapter.NewHealthHandler(db, nil)
	log.Info("health handler initialized")

	// 8. Setup HTTP router
	router := httpAdapter.SetupRoutes(handler, healthHandler, log)
	log.Info("routes configured")

	// 9. Start State Publisher (1 Hz event publishing)
	if publisher != nil {
		// TODO: Get list of all batteries from Asset Management Service
		// For M5, we'll use a hardcoded battery ID for demonstration
		batteryIDs := []string{"battery-123"}

		for _, batteryID := range batteryIDs {
			statePublisher := service.NewStatePublisher(repo, publisher, batteryID, log)
			go statePublisher.Start(context.Background())
		}
	}

	// 10. Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 11. Start server in a goroutine
	go func() {
		log.Info("HTTP server listening", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()

	// 12. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server exited")
}

func loadConfig() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		Port:        getEnv("PORT", "8082"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigrations(db *sql.DB, databaseURL string, log *zap.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/adapters/postgres/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	log.Info("database migrations completed successfully")
	return nil
}
