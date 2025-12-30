package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func main() {
	// Get NATS URL from environment or use default
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	// Get subscription pattern from command line args or use default (all events)
	pattern := ">"
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}

	log.SetFlags(log.LstdFlags)
	fmt.Printf("%s=== Event Subscriber Tool ===%s\n", colorCyan, colorReset)
	fmt.Printf("NATS URL: %s\n", natsURL)
	fmt.Printf("Subscription Pattern: %s%s%s\n", colorYellow, pattern, colorReset)
	fmt.Printf("Listening for events... (Press Ctrl+C to stop)\n\n")

	// Connect to NATS
	subscriber, err := events.NewNATSSubscriber(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer subscriber.Close()

	// Create context for subscription
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Event counter
	eventCount := 0

	// Subscribe to events
	handler := func(subject string, data []byte) error {
		eventCount++

		// Pretty-print the event
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("%s[%s] Event #%d%s\n", colorGreen, timestamp, eventCount, colorReset)
		fmt.Printf("%sSubject:%s %s\n", colorBlue, colorReset, subject)

		// Try to pretty-print JSON
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal(data, &prettyJSON); err == nil {
			prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
			fmt.Printf("%sPayload:%s\n%s\n", colorPurple, colorReset, string(prettyBytes))
		} else {
			// Fallback to raw data
			fmt.Printf("%sPayload:%s\n%s\n", colorPurple, colorReset, string(data))
		}

		fmt.Println(colorWhite + "---" + colorReset)
		return nil
	}

	err = subscriber.Subscribe(ctx, pattern, handler)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	fmt.Printf("%sSubscribed successfully!%s\n\n", colorGreen, colorReset)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Printf("\n%sShutting down...%s\n", colorYellow, colorReset)
	fmt.Printf("Total events received: %s%d%s\n", colorGreen, eventCount, colorReset)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
