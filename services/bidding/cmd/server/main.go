package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/pkg/logger"
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
	// 1. Load config from environment variables
	cfg := loadConfig()

	// 2. Setup structured logging
	log, err := logger.NewFromEnv("bidding", cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("starting service",
		zap.String("nats_url", cfg.NatsURL),
		zap.String("automation_mode", cfg.AutomationMode),
		zap.String("log_level", cfg.LogLevel),
	)

	// 3. Connect to NATS
	var publisher events.EventPublisher
	var subscriber *events.NATSSubscriber

	if cfg.NatsURL != "" {
		publisher, err = events.NewNATSPublisher(cfg.NatsURL)
		if err != nil {
			log.Warn("failed to connect to NATS for publishing", zap.Error(err))
			log.Info("service will continue without event publishing")
		} else {
			defer publisher.Close()
			log.Info("connected to NATS for publishing")
		}

		subscriber, err = events.NewNATSSubscriber(cfg.NatsURL)
		if err != nil {
			log.Fatal("failed to connect to NATS for subscribing", zap.Error(err))
		}
		defer subscriber.Close()
		log.Info("connected to NATS for subscribing")
	} else {
		log.Warn("NATS_URL not set - running without event bus")
	}

	// 4. Create caches
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	log.Info("initialized in-memory caches")

	// 5. Create bidding engine
	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, cfg.AutomationMode, log)
	log.Info("created bidding engine", zap.String("automation_mode", cfg.AutomationMode))

	// 6. Setup event subscriptions
	if subscriber != nil {
		ctx := context.Background()
		setupSubscriptions(ctx, subscriber, engine, log)
	}

	// 7. Start service (no HTTP server in M6, event-driven only)
	log.Info("bidding service is running in event-driven mode")
	log.Info("press Ctrl+C to shutdown")

	// 8. Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down bidding service")
	log.Info("bidding service stopped")
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
func setupSubscriptions(ctx context.Context, subscriber *events.NATSSubscriber, engine *service.BiddingEngine, log *zap.Logger) {
	// Subscribe to battery.state.changed.v1
	if err := subscriber.Subscribe(ctx, "battery.state.changed.v1", func(subject string, data []byte) error {
		var event events.BatteryStateChanged
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal BatteryStateChanged: %w", err)
		}
		return engine.HandleBatteryStateChanged(event)
	}); err != nil {
		log.Fatal("failed to subscribe to battery.state.changed.v1", zap.Error(err))
	}
	log.Info("subscribed to battery.state.changed.v1")

	// Subscribe to market.price.updated.v1
	if err := subscriber.Subscribe(ctx, "market.price.updated.v1", func(subject string, data []byte) error {
		var event events.MarketPriceUpdated
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal MarketPriceUpdated: %w", err)
		}
		return engine.HandleMarketPriceUpdated(event)
	}); err != nil {
		log.Fatal("failed to subscribe to market.price.updated.v1", zap.Error(err))
	}
	log.Info("subscribed to market.price.updated.v1")

	// Subscribe to battery.registered.v1
	if err := subscriber.Subscribe(ctx, "battery.registered.v1", func(subject string, data []byte) error {
		var event events.BatteryRegistered
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal BatteryRegistered: %w", err)
		}
		return engine.HandleBatteryRegistered(event)
	}); err != nil {
		log.Fatal("failed to subscribe to battery.registered.v1", zap.Error(err))
	}
	log.Info("subscribed to battery.registered.v1")

	log.Info("all event subscriptions configured successfully")
}
