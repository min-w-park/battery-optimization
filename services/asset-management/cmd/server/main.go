package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/minwook/battery-optimization/pkg/events"
	httpAdapter "github.com/minwook/battery-optimization/services/asset-management/internal/adapters/http"
	postgresAdapter "github.com/minwook/battery-optimization/services/asset-management/internal/adapters/postgres"
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

	// 2. Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting Asset Management Service...")
	log.Printf("Configuration: Port=%s, LogLevel=%s", cfg.Port, cfg.LogLevel)

	// 3. Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Database connection established")

	// 4. Run migrations using golang-migrate
	if err := runMigrations(db, cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 5. Create repository
	repo := postgresAdapter.NewPostgresRepository(db)
	log.Println("Repository initialized")

	// 6. Connect to NATS (optional - service works without events)
	var publisher events.EventPublisher
	if cfg.NatsURL != "" {
		natsPublisher, err := events.NewNATSPublisher(cfg.NatsURL)
		if err != nil {
			log.Printf("WARNING: Failed to connect to NATS at %s: %v", cfg.NatsURL, err)
			log.Println("Service will continue WITHOUT event publishing")
		} else {
			publisher = natsPublisher
			defer publisher.Close()
			log.Printf("NATS publisher connected to %s", cfg.NatsURL)
		}
	} else {
		log.Println("NATS_URL not set - running without event publishing")
	}

	// 7. Create handler
	handler := httpAdapter.NewBatteryHandler(repo, publisher)
	log.Println("HTTP handler initialized")

	// 8. Setup HTTP router
	router := httpAdapter.SetupRoutes(handler)
	log.Println("Routes configured")

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
		log.Printf("Starting HTTP server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	cfg := Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://asset_user:asset_pass@localhost:5432/asset_management?sslmode=disable"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		Port:        getEnv("PORT", "8080"),
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
func runMigrations(db *sql.DB, databaseURL string) error {
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

	log.Println("Migrations applied successfully")
	return nil
}
