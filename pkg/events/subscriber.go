package events

import "context"

// EventHandler processes a received event.
//
// Parameters:
//   - subject: The NATS subject the event was published to (e.g., "battery.registered.v1")
//   - data: The raw JSON bytes of the event
//
// Returns:
//   - nil: Event processed successfully (message will be ACKed)
//   - error: Event processing failed (message will be NACKed and redelivered)
//
// Example:
//   handler := func(subject string, data []byte) error {
//       var event events.BatteryRegistered
//       if err := json.Unmarshal(data, &event); err != nil {
//           log.Printf("ERROR: Invalid JSON: %v", err)
//           return nil // ACK anyway - malformed JSON won't fix itself
//       }
//
//       // Process event...
//       log.Printf("Received: %s - Battery %s", subject, event.BatteryID)
//       return nil
//   }
type EventHandler func(subject string, data []byte) error

// EventSubscriber subscribes to domain events from NATS.
//
// Implementations:
//   - NATSSubscriber: Subscribes to events from NATS message broker
//
// Usage:
//   subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer subscriber.Close()
//
//   handler := func(subject string, data []byte) error {
//       log.Printf("Received: %s", subject)
//       return nil
//   }
//
//   // Subscribe to all battery events (v1)
//   err = subscriber.Subscribe(ctx, "battery.*.v1", handler)
//
//   // Subscribe to all events
//   err = subscriber.Subscribe(ctx, ">", handler)
type EventSubscriber interface {
	// Subscribe registers a handler for events matching the subject pattern.
	//
	// Subject patterns support NATS wildcards:
	//   - "*" matches a single token (e.g., "battery.*.v1" matches "battery.registered.v1")
	//   - ">" matches multiple tokens (e.g., ">" matches all events)
	//
	// The handler will be called for each matching event.
	// Multiple subscriptions can be active simultaneously.
	//
	// Parameters:
	//   - ctx: Context for cancellation (subscription remains active until Close() or ctx cancellation)
	//   - subject: Subject pattern to subscribe to
	//   - handler: Function to call for each matching event
	//
	// Returns error if subscription fails.
	Subscribe(ctx context.Context, subject string, handler EventHandler) error

	// Close gracefully shuts down the subscriber.
	// It unsubscribes from all subjects and closes the NATS connection.
	Close() error
}
