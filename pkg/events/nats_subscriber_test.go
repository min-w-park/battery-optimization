package events_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewNATSSubscriber_Success verifies successful NATS connection
func TestNewNATSSubscriber_Success(t *testing.T) {
	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	require.NotNil(t, subscriber)

	err = subscriber.Close()
	assert.NoError(t, err)
}

// TestNewNATSSubscriber_ConnectionFailure verifies error handling for invalid NATS URL
func TestNewNATSSubscriber_ConnectionFailure(t *testing.T) {
	subscriber, err := events.NewNATSSubscriber("nats://invalid-host:9999")
	assert.Error(t, err)
	assert.Nil(t, subscriber)
	assert.Contains(t, err.Error(), "failed to connect to NATS")
}

// TestNATSSubscriber_Subscribe verifies basic event subscription
func TestNATSSubscriber_Subscribe(t *testing.T) {
	// Create publisher and subscriber
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	// Channel to capture received events
	received := make(chan string, 1)

	// Subscribe to battery events
	ctx := context.Background()
	handler := func(subject string, data []byte) error {
		received <- subject
		return nil
	}
	err = subscriber.Subscribe(ctx, "battery.registered.v1", handler)
	require.NoError(t, err)

	// Give subscriber time to register
	time.Sleep(100 * time.Millisecond)

	// Publish event
	event := events.BatteryRegistered{
		BatteryID:    "BATT-001",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "NSW",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.9,
			WarrantyEOL: 0.8,
			MaxCycles:   5000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = publisher.Publish(pubCtx, "battery.registered.v1", event)
	require.NoError(t, err)

	// Wait for event
	select {
	case subject := <-received:
		assert.Equal(t, "battery.registered.v1", subject)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

// TestNATSSubscriber_WildcardSubscription verifies wildcard pattern matching
func TestNATSSubscriber_WildcardSubscription(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	// Channel to capture received events
	received := make(chan string, 10)

	// Subscribe to all battery events with wildcard (battery.>)
	ctx := context.Background()
	handler := func(subject string, data []byte) error {
		received <- subject
		return nil
	}
	err = subscriber.Subscribe(ctx, "battery.>", handler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish battery.registered.v1 event
	batteryEvent := events.BatteryRegistered{
		BatteryID:    "BATT-001",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "NSW",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.9,
			WarrantyEOL: 0.8,
			MaxCycles:   5000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	pubCtx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	err = publisher.Publish(pubCtx1, "battery.registered.v1", batteryEvent)
	require.NoError(t, err)

	// Small delay to ensure first event is processed
	time.Sleep(50 * time.Millisecond)

	// Publish battery.state.changed.v1 event
	stateEvent := events.BatteryStateChanged{
		BatteryID:      "BATT-001",
		SoC:            0.75,
		Power:          25.0,
		OperationState: "CHARGING",
		Temperature:    25.5,
		Voltage:        800.0,
		Current:        31.25,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	pubCtx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	err = publisher.Publish(pubCtx2, "battery.state.changed.v1", stateEvent)
	require.NoError(t, err)

	// Should receive both events
	receivedSubjects := make(map[string]bool)
	for i := 0; i < 2; i++ {
		select {
		case subject := <-received:
			t.Logf("Received event %d: %s", i+1, subject)
			receivedSubjects[subject] = true
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for events. Received so far: %v", receivedSubjects)
		}
	}

	assert.True(t, receivedSubjects["battery.registered.v1"])
	assert.True(t, receivedSubjects["battery.state.changed.v1"])
}

// TestNATSSubscriber_AllEventsWildcard verifies subscription to all events (>)
func TestNATSSubscriber_AllEventsWildcard(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	// Channel to capture received events
	received := make(chan string, 10)

	// Subscribe to ALL events with ">" wildcard
	ctx := context.Background()
	handler := func(subject string, data []byte) error {
		received <- subject
		return nil
	}
	err = subscriber.Subscribe(ctx, ">", handler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish events from different domains
	batteryEvent := events.BatteryRegistered{
		BatteryID:    "BATT-001",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "NSW",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.9,
			WarrantyEOL: 0.8,
			MaxCycles:   5000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	marketEvent := events.MarketPriceUpdated{
		PriceID:       "PRICE-001",
		Region:        "NSW",
		Price:         75.50,
		Demand:        8500.0,
		IntervalType:  "5MIN",
		IntervalStart: time.Now(),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = publisher.Publish(pubCtx, "battery.registered.v1", batteryEvent)
	require.NoError(t, err)

	err = publisher.Publish(pubCtx, "market.price.updated.v1", marketEvent)
	require.NoError(t, err)

	// Should receive both events from different domains
	receivedSubjects := make(map[string]bool)
	for i := 0; i < 2; i++ {
		select {
		case subject := <-received:
			receivedSubjects[subject] = true
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for events")
		}
	}

	assert.True(t, receivedSubjects["battery.registered.v1"])
	assert.True(t, receivedSubjects["market.price.updated.v1"])
}

// TestNATSSubscriber_HandlerError verifies error handling in event handler
func TestNATSSubscriber_HandlerError(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	// Handler that returns error
	handlerCalled := make(chan bool, 1)
	ctx := context.Background()
	handler := func(subject string, data []byte) error {
		handlerCalled <- true
		return assert.AnError // Return an error
	}
	err = subscriber.Subscribe(ctx, "battery.registered.v1", handler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish event
	event := events.BatteryRegistered{
		BatteryID:    "BATT-001",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "NSW",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.9,
			WarrantyEOL: 0.8,
			MaxCycles:   5000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = publisher.Publish(pubCtx, "battery.registered.v1", event)
	require.NoError(t, err)

	// Handler should still be called even if it returns error
	select {
	case <-handlerCalled:
		// Success - handler was called
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for handler call")
	}
}

// TestNATSSubscriber_MultipleSubscriptions verifies multiple concurrent subscriptions
func TestNATSSubscriber_MultipleSubscriptions(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	// Two different handlers for different subjects
	var mu sync.Mutex
	receivedBattery := false
	receivedMarket := false

	ctx := context.Background()

	// Subscribe to battery events
	batteryHandler := func(subject string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		receivedBattery = true
		return nil
	}
	err = subscriber.Subscribe(ctx, "battery.registered.v1", batteryHandler)
	require.NoError(t, err)

	// Subscribe to market events
	marketHandler := func(subject string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		receivedMarket = true
		return nil
	}
	err = subscriber.Subscribe(ctx, "market.price.updated.v1", marketHandler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish both events
	batteryEvent := events.BatteryRegistered{
		BatteryID:    "BATT-001",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "NSW",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.9,
			WarrantyEOL: 0.8,
			MaxCycles:   5000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	marketEvent := events.MarketPriceUpdated{
		PriceID:       "PRICE-001",
		Region:        "NSW",
		Price:         75.50,
		Demand:        8500.0,
		IntervalType:  "5MIN",
		IntervalStart: time.Now(),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = publisher.Publish(pubCtx, "battery.registered.v1", batteryEvent)
	require.NoError(t, err)

	err = publisher.Publish(pubCtx, "market.price.updated.v1", marketEvent)
	require.NoError(t, err)

	// Wait for both handlers to be called
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.True(t, receivedBattery, "Battery handler should have been called")
	assert.True(t, receivedMarket, "Market handler should have been called")
}

// TestNATSSubscriber_EventDeserialization verifies correct JSON deserialization
func TestNATSSubscriber_EventDeserialization(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	// Channel to capture deserialized event
	received := make(chan events.MarketPriceUpdated, 1)

	ctx := context.Background()
	handler := func(subject string, data []byte) error {
		var event events.MarketPriceUpdated
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		received <- event
		return nil
	}
	err = subscriber.Subscribe(ctx, "market.price.updated.v1", handler)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Publish event with specific values
	originalEvent := events.MarketPriceUpdated{
		PriceID:       "PRICE-123",
		Region:        "VIC",
		Price:         123.45,
		Demand:        9500.0,
		IntervalType:  "30MIN",
		IntervalStart: time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
		PublishedAt:   time.Date(2025, 12, 30, 10, 5, 0, 0, time.UTC),
		Timestamp:     time.Date(2025, 12, 30, 10, 5, 30, 0, time.UTC),
		EventVersion:  "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = publisher.Publish(pubCtx, "market.price.updated.v1", originalEvent)
	require.NoError(t, err)

	// Wait for deserialized event
	select {
	case event := <-received:
		assert.Equal(t, "PRICE-123", event.PriceID)
		assert.Equal(t, "VIC", event.Region)
		assert.Equal(t, 123.45, event.Price)
		assert.Equal(t, 9500.0, event.Demand)
		assert.Equal(t, "30MIN", event.IntervalType)
		assert.Equal(t, "v1", event.EventVersion)
		// Timestamp comparison using Unix() to handle time zones
		assert.Equal(t, originalEvent.IntervalStart.Unix(), event.IntervalStart.Unix())
		assert.Equal(t, originalEvent.PublishedAt.Unix(), event.PublishedAt.Unix())
		assert.Equal(t, originalEvent.Timestamp.Unix(), event.Timestamp.Unix())
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

// TestNATSSubscriber_Close verifies graceful shutdown
func TestNATSSubscriber_Close(t *testing.T) {
	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)

	// Subscribe to an event
	ctx := context.Background()
	handler := func(subject string, data []byte) error {
		return nil
	}
	err = subscriber.Subscribe(ctx, "battery.registered.v1", handler)
	require.NoError(t, err)

	// Close should not error
	err = subscriber.Close()
	assert.NoError(t, err)

	// Second close should also not error
	err = subscriber.Close()
	assert.NoError(t, err)
}
