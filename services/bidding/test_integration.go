// +build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
)

func main() {
	log.Println("M6 Integration Test: Bidding Service")
	log.Println("=====================================")

	// Connect to NATS
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer publisher.Close()

	ctx := context.Background()

	// Test 1: Battery Registration
	log.Println("\n[Test 1] Publishing BatteryRegistered event...")
	batteryEvent := events.BatteryRegistered{
		BatteryID:    "test-battery-123",
		Capacity:     200.0,
		MaxPower:     100.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "SA",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:              0.1,
			MaxSoC:              0.9,
			WarrantyEOL:         0.8,
			MaxCycles:           10000,
			OperatingTempMin:    -20.0,
			OperatingTempMax:    60.0,
			GridComplianceLevel: "AS4777",
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	if err := publisher.Publish(ctx, "battery.registered.v1", batteryEvent); err != nil {
		log.Fatalf("Failed to publish battery registered: %v", err)
	}
	log.Printf("✓ Battery registered: %s", batteryEvent.BatteryID)
	time.Sleep(500 * time.Millisecond)

	// Test 2: Low Market Price (should trigger charging opportunity)
	log.Println("\n[Test 2] Publishing MarketPriceUpdated (LOW price)...")
	lowPriceEvent := events.MarketPriceUpdated{
		PriceID:       "price-low-1",
		Region:        "SA",
		Price:         30.0, // Below $50 threshold
		Demand:        2000.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: time.Now().Add(5 * time.Minute),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	if err := publisher.Publish(ctx, "market.price.updated.v1", lowPriceEvent); err != nil {
		log.Fatalf("Failed to publish low price: %v", err)
	}
	log.Printf("✓ Low price published: $%.2f/MWh", lowPriceEvent.Price)
	time.Sleep(500 * time.Millisecond)

	// Test 3: Battery State Changed (with low SoC)
	log.Println("\n[Test 3] Publishing BatteryStateChanged (low SoC, IDLE)...")
	batteryStateEvent := events.BatteryStateChanged{
		BatteryID:      "test-battery-123",
		SoC:            50.0, // Low SoC
		Power:          0.0,
		OperationState: "IDLE",
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	if err := publisher.Publish(ctx, "battery.state.changed.v1", batteryStateEvent); err != nil {
		log.Fatalf("Failed to publish battery state: %v", err)
	}
	log.Printf("✓ Battery state published: SoC=%.1f%%, State=%s", batteryStateEvent.SoC, batteryStateEvent.OperationState)
	log.Println("  → Expected: Bidding service should detect CHARGING OPPORTUNITY and issue CHARGING COMMAND")
	time.Sleep(1 * time.Second)

	// Test 4: High Market Price (should trigger discharging opportunity)
	log.Println("\n[Test 4] Publishing MarketPriceUpdated (HIGH price)...")
	highPriceEvent := events.MarketPriceUpdated{
		PriceID:       "price-high-1",
		Region:        "SA",
		Price:         150.0, // Above $100 threshold
		Demand:        2500.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: time.Now().Add(5 * time.Minute),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	if err := publisher.Publish(ctx, "market.price.updated.v1", highPriceEvent); err != nil {
		log.Fatalf("Failed to publish high price: %v", err)
	}
	log.Printf("✓ High price published: $%.2f/MWh", highPriceEvent.Price)
	time.Sleep(500 * time.Millisecond)

	// Test 5: Battery State Changed (with high SoC)
	log.Println("\n[Test 5] Publishing BatteryStateChanged (high SoC, IDLE)...")
	batteryStateEventHigh := events.BatteryStateChanged{
		BatteryID:      "test-battery-123",
		SoC:            70.0, // High SoC
		Power:          0.0,
		OperationState: "IDLE",
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	if err := publisher.Publish(ctx, "battery.state.changed.v1", batteryStateEventHigh); err != nil {
		log.Fatalf("Failed to publish battery state: %v", err)
	}
	log.Printf("✓ Battery state published: SoC=%.1f%%, State=%s", batteryStateEventHigh.SoC, batteryStateEventHigh.OperationState)
	log.Println("  → Expected: Bidding service should detect DISCHARGING OPPORTUNITY and issue DISCHARGING COMMAND")
	time.Sleep(1 * time.Second)

	// Test 6: Mid-range Price (should NOT trigger any opportunity)
	log.Println("\n[Test 6] Publishing MarketPriceUpdated (MID-RANGE price)...")
	midPriceEvent := events.MarketPriceUpdated{
		PriceID:       "price-mid-1",
		Region:        "SA",
		Price:         75.0, // Between $50 and $100
		Demand:        2200.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: time.Now().Add(5 * time.Minute),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	if err := publisher.Publish(ctx, "market.price.updated.v1", midPriceEvent); err != nil {
		log.Fatalf("Failed to publish mid price: %v", err)
	}
	log.Printf("✓ Mid-range price published: $%.2f/MWh", midPriceEvent.Price)
	log.Println("  → Expected: NO opportunity detected (price between thresholds)")
	time.Sleep(1 * time.Second)

	log.Println("\n=====================================")
	log.Println("Integration test events published successfully!")
	log.Println("Check bidding service logs for:")
	log.Println("  - Battery registered in cache")
	log.Println("  - Charging opportunity detected (Test 3)")
	log.Println("  - Charging command issued (FULL_AUTO mode)")
	log.Println("  - Discharging opportunity detected (Test 5)")
	log.Println("  - Discharging command issued (FULL_AUTO mode)")
	log.Println("  - No events for mid-range price (Test 6)")
}

// Helper to print event as JSON
func printEvent(name string, event interface{}) {
	data, _ := json.MarshalIndent(event, "", "  ")
	fmt.Printf("\n[%s]\n%s\n", name, string(data))
}
