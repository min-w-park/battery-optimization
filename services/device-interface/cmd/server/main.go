package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/pkg/logger"
	"github.com/minwook/battery-optimization/services/device-interface/internal/adapters"
	"github.com/minwook/battery-optimization/services/device-interface/internal/domain"
	"github.com/minwook/battery-optimization/services/device-interface/internal/service"
)

// Config holds application configuration
type Config struct {
	NatsURL     string
	BatteryID   string
	AdapterType string
	Capacity    float64 // MWh
	MaxPower    float64 // MW
	LogLevel    string
}

func main() {
	// 1. Load configuration from environment
	cfg := loadConfig()

	// 2. Setup structured logging
	log, err := logger.NewFromEnv("device-interface", cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("starting service",
		zap.String("battery_id", cfg.BatteryID),
		zap.String("adapter_type", cfg.AdapterType),
		zap.Float64("capacity_mwh", cfg.Capacity),
		zap.Float64("max_power_mw", cfg.MaxPower),
	)

	// 3. Create battery adapter based on type
	var adapter domain.BatteryAdapter
	switch cfg.AdapterType {
	case "TeslaLike":
		adapter = adapters.NewTeslaLike(cfg.BatteryID, cfg.Capacity, cfg.MaxPower)
		log.Info("created TeslaLike adapter", zap.String("battery_id", cfg.BatteryID))
	case "BYDLike":
		// TODO: Implement BYDLike adapter in Phase 5 extension
		log.Fatal("BYDLike adapter not yet implemented - use TeslaLike for now")
	default:
		log.Fatal("unknown adapter type", zap.String("adapter_type", cfg.AdapterType))
	}

	// Ensure adapter cleanup on exit
	defer func() {
		if teslaAdapter, ok := adapter.(*adapters.TeslaLike); ok {
			teslaAdapter.Stop()
		}
	}()

	// 4. Connect to NATS for event publishing
	var publisher events.EventPublisher
	if cfg.NatsURL != "" {
		natsPublisher, err := events.NewNATSPublisher(cfg.NatsURL)
		if err != nil {
			log.Warn("failed to connect to NATS for publishing",
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

	// 5. Publish BatteryConnectionEstablished event
	if publisher != nil {
		customAttrs := adapter.GetCustomAttributes()
		firmwareVersion := ""
		if fw, ok := customAttrs["firmwareVersion"].(string); ok {
			firmwareVersion = fw
		}

		connectionEvent := events.BatteryConnectionEstablished{
			BatteryID:        cfg.BatteryID,
			AdapterType:      cfg.AdapterType,
			FirmwareVersion:  firmwareVersion,
			CustomAttributes: customAttrs,
			Timestamp:        time.Now(),
			EventVersion:     "v1",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := publisher.Publish(ctx, "battery.connection.established.v1", connectionEvent); err != nil {
			log.Warn("failed to publish BatteryConnectionEstablished", zap.Error(err))
		} else {
			log.Info("published BatteryConnectionEstablished event",
				zap.String("battery_id", cfg.BatteryID),
			)
		}
		cancel()
	}

	// 6. Create command handler
	var commandHandler *service.CommandHandler
	if publisher != nil {
		commandHandler = service.NewCommandHandler(adapter, publisher, log)
		log.Info("command handler initialized")
	}

	// 7. Connect to NATS for event subscription
	var subscriber events.EventSubscriber
	if cfg.NatsURL != "" && commandHandler != nil {
		natsSubscriber, err := events.NewNATSSubscriber(cfg.NatsURL)
		if err != nil {
			log.Warn("failed to connect to NATS for subscription",
				zap.String("nats_url", cfg.NatsURL),
				zap.Error(err),
			)
			log.Info("service will continue without command handling")
		} else {
			subscriber = natsSubscriber
			defer subscriber.Close()
			log.Info("NATS subscriber connected", zap.String("nats_url", cfg.NatsURL))

			// Subscribe to command events
			ctx := context.Background()

			// Subscribe to charging commands
			if err := subscriber.Subscribe(ctx, "charging.command.issued.v1", commandHandler.OnEvent); err != nil {
				log.Fatal("failed to subscribe to charging commands", zap.Error(err))
			}
			log.Info("subscribed to charging.command.issued.v1")

			// Subscribe to discharging commands
			if err := subscriber.Subscribe(ctx, "discharging.command.issued.v1", commandHandler.OnEvent); err != nil {
				log.Fatal("failed to subscribe to discharging commands", zap.Error(err))
			}
			log.Info("subscribed to discharging.command.issued.v1")

			// Subscribe to conflict resolutions
			if err := subscriber.Subscribe(ctx, "conflict.resolved.v1", commandHandler.OnEvent); err != nil {
				log.Fatal("failed to subscribe to conflict resolutions", zap.Error(err))
			}
			log.Info("subscribed to conflict.resolved.v1")

			log.Info("event subscriptions active - ready to process commands")
		}
	}

	// 8. Wait for shutdown signal
	log.Info("device interface service is running - press Ctrl+C to stop")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down device interface service")
	log.Info("service exited")
}

func loadConfig() Config {
	return Config{
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		BatteryID:   getEnv("BATTERY_ID", "battery-001"),
		AdapterType: getEnv("ADAPTER_TYPE", "TeslaLike"),
		Capacity:    getEnvFloat("CAPACITY", 200.0),  // Default: 200 MWh
		MaxPower:    getEnvFloat("MAX_POWER", 100.0), // Default: 100 MW
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var f float64
		if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
			return f
		}
	}
	return defaultValue
}
