package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
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

	// 2. Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting Device Interface Service...")
	log.Printf("Configuration: BatteryID=%s, AdapterType=%s, Capacity=%.2f MWh, MaxPower=%.2f MW",
		cfg.BatteryID, cfg.AdapterType, cfg.Capacity, cfg.MaxPower)

	// 3. Create battery adapter based on type
	var adapter domain.BatteryAdapter
	switch cfg.AdapterType {
	case "TeslaLike":
		adapter = adapters.NewTeslaLike(cfg.BatteryID, cfg.Capacity, cfg.MaxPower)
		log.Printf("Created TeslaLike adapter for battery %s", cfg.BatteryID)
	case "BYDLike":
		// TODO: Implement BYDLike adapter in Phase 5 extension
		log.Fatal("BYDLike adapter not yet implemented - use TeslaLike for now")
	default:
		log.Fatalf("Unknown adapter type: %s (use TeslaLike or BYDLike)", cfg.AdapterType)
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
			log.Printf("WARNING: Failed to publish BatteryConnectionEstablished: %v", err)
		} else {
			log.Printf("Published BatteryConnectionEstablished event for battery %s", cfg.BatteryID)
		}
		cancel()
	}

	// 6. Create command handler
	var commandHandler *service.CommandHandler
	if publisher != nil {
		commandHandler = service.NewCommandHandler(adapter, publisher)
		log.Println("Command handler initialized")
	}

	// 7. Connect to NATS for event subscription
	var subscriber events.EventSubscriber
	if cfg.NatsURL != "" && commandHandler != nil {
		natsSubscriber, err := events.NewNATSSubscriber(cfg.NatsURL)
		if err != nil {
			log.Printf("WARNING: Failed to connect to NATS for subscription: %v", err)
			log.Println("Service will continue WITHOUT command handling")
		} else {
			subscriber = natsSubscriber
			defer subscriber.Close()
			log.Printf("NATS subscriber connected to %s", cfg.NatsURL)

			// Subscribe to command events
			ctx := context.Background()

			// Subscribe to charging commands
			if err := subscriber.Subscribe(ctx, "charging.command.issued.v1", commandHandler.OnEvent); err != nil {
				log.Fatalf("Failed to subscribe to charging commands: %v", err)
			}
			log.Println("Subscribed to: charging.command.issued.v1")

			// Subscribe to discharging commands
			if err := subscriber.Subscribe(ctx, "discharging.command.issued.v1", commandHandler.OnEvent); err != nil {
				log.Fatalf("Failed to subscribe to discharging commands: %v", err)
			}
			log.Println("Subscribed to: discharging.command.issued.v1")

			// Subscribe to conflict resolutions
			if err := subscriber.Subscribe(ctx, "conflict.resolved.v1", commandHandler.OnEvent); err != nil {
				log.Fatalf("Failed to subscribe to conflict resolutions: %v", err)
			}
			log.Println("Subscribed to: conflict.resolved.v1")

			log.Println("Event subscriptions active - ready to process commands")
		}
	}

	// 8. Wait for shutdown signal
	log.Println("Device Interface Service is running - Press Ctrl+C to stop")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Device Interface Service...")
	log.Println("Service exited")
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
