package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/bidding/internal/cache"
	"github.com/minwook/battery-optimization/services/bidding/internal/service"
)

// Config holds application configuration
type Config struct {
	NatsURL        string
	AutomationMode string
	LogLevel       string
}

func main() {
	log.Println("Starting Bidding Service...")

	// 1. Load config from environment variables
	cfg := loadConfig()
	log.Printf("Configuration: NATS_URL=%s, AUTOMATION_MODE=%s, LOG_LEVEL=%s",
		cfg.NatsURL, cfg.AutomationMode, cfg.LogLevel)

	// 2. Connect to NATS
	var publisher events.EventPublisher
	var subscriber *events.NATSSubscriber
	var err error

	if cfg.NatsURL != "" {
		publisher, err = events.NewNATSPublisher(cfg.NatsURL)
		if err != nil {
			log.Printf("WARNING: Failed to connect to NATS for publishing: %v", err)
			log.Println("Service will continue WITHOUT event publishing")
		} else {
			defer publisher.Close()
			log.Println("Connected to NATS for publishing")
		}

		subscriber, err = events.NewNATSSubscriber(cfg.NatsURL)
		if err != nil {
			log.Fatalf("FATAL: Failed to connect to NATS for subscribing: %v", err)
		}
		defer subscriber.Close()
		log.Println("Connected to NATS for subscribing")
	} else {
		log.Println("WARNING: NATS_URL not set, running without event bus")
	}

	// 3. Create caches
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	log.Println("Initialized in-memory caches")

	// 4. Create bidding engine
	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, cfg.AutomationMode)
	log.Printf("Created bidding engine with automation mode: %s", cfg.AutomationMode)

	// 5. Setup event subscriptions
	if subscriber != nil {
		ctx := context.Background()
		setupSubscriptions(ctx, subscriber, engine)
	}

	// 6. Start service (no HTTP server in M6, event-driven only)
	log.Println("Bidding Service is running (event-driven mode)")
	log.Println("Press Ctrl+C to shutdown...")

	// 7. Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Bidding Service...")
	log.Println("Bidding Service stopped")
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		NatsURL:        getEnv("NATS_URL", "nats://localhost:4222"),
		AutomationMode: getEnv("AUTOMATION_MODE", "MANUAL"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv retrieves environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupSubscriptions configures NATS event subscriptions
func setupSubscriptions(ctx context.Context, subscriber *events.NATSSubscriber, engine *service.BiddingEngine) {
	// Subscribe to battery.state.changed.v1
	if err := subscriber.Subscribe(ctx, "battery.state.changed.v1", func(subject string, data []byte) error {
		var event events.BatteryStateChanged
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal BatteryStateChanged: %w", err)
		}
		return engine.HandleBatteryStateChanged(event)
	}); err != nil {
		log.Fatalf("Failed to subscribe to battery.state.changed.v1: %v", err)
	}
	log.Println("Subscribed to: battery.state.changed.v1")

	// Subscribe to market.price.updated.v1
	if err := subscriber.Subscribe(ctx, "market.price.updated.v1", func(subject string, data []byte) error {
		var event events.MarketPriceUpdated
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal MarketPriceUpdated: %w", err)
		}
		return engine.HandleMarketPriceUpdated(event)
	}); err != nil {
		log.Fatalf("Failed to subscribe to market.price.updated.v1: %v", err)
	}
	log.Println("Subscribed to: market.price.updated.v1")

	// Subscribe to battery.registered.v1
	if err := subscriber.Subscribe(ctx, "battery.registered.v1", func(subject string, data []byte) error {
		var event events.BatteryRegistered
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal BatteryRegistered: %w", err)
		}
		return engine.HandleBatteryRegistered(event)
	}); err != nil {
		log.Fatalf("Failed to subscribe to battery.registered.v1: %v", err)
	}
	log.Println("Subscribed to: battery.registered.v1")

	log.Println("All event subscriptions configured successfully")
}
