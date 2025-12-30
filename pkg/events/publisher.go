package events

import "context"

// EventPublisher publishes domain events to NATS.
//
// Implementations:
//   - NATSPublisher: Publishes events to NATS message broker
//
// Usage:
//   publisher, err := events.NewNATSPublisher("nats://localhost:4222")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer publisher.Close()
//
//   event := events.BatteryRegistered{...}
//   err = publisher.Publish(ctx, "battery.registered.v1", event)
type EventPublisher interface {
	// Publish sends an event to the specified NATS subject.
	// The event will be JSON-serialized before publishing.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - subject: NATS subject (e.g., "battery.registered.v1")
	//   - event: Event struct to publish (will be JSON-marshaled)
	//
	// Returns error if:
	//   - JSON marshaling fails
	//   - NATS publish fails
	//   - Context timeout/cancellation
	Publish(ctx context.Context, subject string, event interface{}) error

	// Close gracefully shuts down the publisher.
	// It drains pending messages and closes the NATS connection.
	Close() error
}
