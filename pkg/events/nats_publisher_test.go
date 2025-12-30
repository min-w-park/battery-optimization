package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNATSPublisher_Success(t *testing.T) {
	// NOTE: Requires NATS running on localhost:4222
	// Run: docker-compose up -d nats
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	require.NotNil(t, publisher)
	defer publisher.Close()
}

func TestNewNATSPublisher_ConnectionFailure(t *testing.T) {
	// Try to connect to invalid URL
	publisher, err := events.NewNATSPublisher("nats://invalid-host:9999")
	assert.Error(t, err)
	assert.Nil(t, publisher)
}

func TestNATSPublisher_Publish(t *testing.T) {
	// NOTE: Requires NATS running on localhost:4222
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	// Create test event
	event := events.BatteryRegistered{
		BatteryID:    "test-battery",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "Test Location",
		Manufacturer: "Test Manufacturer",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.95,
			WarrantyEOL: 0.8,
			MaxCycles:   15000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	// Publish event
	ctx := context.Background()
	err = publisher.Publish(ctx, "battery.registered.v1", event)
	assert.NoError(t, err)
}

func TestNATSPublisher_PublishWithTimeout(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	event := events.MarketPriceUpdated{
		PriceID:       "test-price",
		Region:        "NSW",
		Price:         100.0,
		Demand:        5000.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: time.Now(),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	// Publish with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = publisher.Publish(ctx, "market.price.updated.v1", event)
	assert.NoError(t, err)
}

func TestNATSPublisher_PublishInvalidEvent(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	// Try to publish something that can't be JSON-marshaled
	// (channels can't be marshaled)
	invalidEvent := make(chan int)

	ctx := context.Background()
	err = publisher.Publish(ctx, "test.subject", invalidEvent)
	assert.Error(t, err, "Should fail to marshal channel to JSON")
}

func TestNATSPublisher_Close(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)

	// Close should not error
	err = publisher.Close()
	assert.NoError(t, err)

	// Second close should be safe (idempotent)
	err = publisher.Close()
	assert.NoError(t, err)
}

func TestNATSPublisher_PublishAfterClose(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)

	// Close connection
	publisher.Close()

	// Try to publish after close
	event := events.BatteryRegistered{
		BatteryID:    "test",
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	ctx := context.Background()
	err = publisher.Publish(ctx, "test.subject", event)
	assert.Error(t, err, "Should fail to publish after close")
}

func TestNATSPublisher_MultipleEvents(t *testing.T) {
	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
	require.NoError(t, err)
	defer publisher.Close()

	ctx := context.Background()

	// Publish multiple events in sequence
	for i := 0; i < 10; i++ {
		event := events.BatteryRegistered{
			BatteryID:    "battery-" + string(rune('0'+i)),
			Capacity:     float64(50 + i),
			Timestamp:    time.Now(),
			EventVersion: "v1",
		}

		err = publisher.Publish(ctx, "battery.registered.v1", event)
		assert.NoError(t, err)
	}
}
