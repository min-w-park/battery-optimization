package main

import (
	"context"
	"log"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
)

func main() {
	// Connect to NATS
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer publisher.Close()

	log.Println("Publishing DischargingCommandIssued event...")

	// Publish discharging command event
	event := events.DischargingCommandIssued{
		BatteryID:    "battery-123",
		CommandID:    "cmd-002",
		Power:        25.0,
		Duration:     30,
		StopConditions: events.StopConditions{
			MinSoC: 20.0,
		},
		DecisionMode: "SEMI_AUTO",
		ApprovedBy:   "operator-001",
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := publisher.Publish(ctx, "discharging.command.issued.v1", event); err != nil {
		log.Fatal("Failed to publish event:", err)
	}

	log.Println("Event published successfully!")
	time.Sleep(1 * time.Second) // Give time for subscribers to process
}
