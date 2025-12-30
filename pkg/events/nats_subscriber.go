package events

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSSubscriber implements EventSubscriber for NATS message broker.
type NATSSubscriber struct {
	conn          *nats.Conn
	subscriptions []*nats.Subscription
	mu            sync.Mutex
}

// NewNATSSubscriber creates a new NATS subscriber.
//
// Parameters:
//   - url: NATS server URL (e.g., "nats://localhost:4222")
//
// Returns:
//   - *NATSSubscriber: Subscriber instance
//   - error: Connection error if NATS is unreachable
//
// Example:
//
//	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
//	if err != nil {
//	    log.Fatalf("Failed to connect to NATS: %v", err)
//	}
//	defer subscriber.Close()
func NewNATSSubscriber(url string) (*NATSSubscriber, error) {
	// Connect to NATS with timeout
	conn, err := nats.Connect(
		url,
		nats.Timeout(10*time.Second),
		nats.Name("event-subscriber"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", url, err)
	}

	return &NATSSubscriber{
		conn:          conn,
		subscriptions: make([]*nats.Subscription, 0),
	}, nil
}

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
// Returns:
//   - error: Subscription error if NATS connection is closed or invalid subject
//
// Example:
//
//	handler := func(subject string, data []byte) error {
//	    var event events.BatteryRegistered
//	    if err := json.Unmarshal(data, &event); err != nil {
//	        log.Printf("ERROR: Invalid JSON: %v", err)
//	        return nil // ACK anyway - malformed JSON won't fix itself
//	    }
//	    log.Printf("Received: %s - Battery %s", subject, event.BatteryID)
//	    return nil
//	}
//	err := subscriber.Subscribe(ctx, "battery.*.v1", handler)
func (s *NATSSubscriber) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if connection is closed
	if s.conn == nil || s.conn.IsClosed() {
		return fmt.Errorf("NATS connection is closed")
	}

	// Wrap the handler to match NATS MsgHandler signature
	natsHandler := func(msg *nats.Msg) {
		// Call the user's handler
		if err := handler(msg.Subject, msg.Data); err != nil {
			// Log error but don't fail - handler errors are application-level
			log.Printf("ERROR: Event handler failed for subject %s: %v", msg.Subject, err)
		}
	}

	// Subscribe to the subject
	sub, err := s.conn.Subscribe(subject, natsHandler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	// Track subscription for cleanup
	s.subscriptions = append(s.subscriptions, sub)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()
		if sub.IsValid() {
			sub.Unsubscribe()
		}
	}()

	return nil
}

// Close gracefully shuts down the subscriber.
//
// It unsubscribes from all subjects, drains pending messages (waits for them
// to be processed), and then closes the NATS connection. Calling Close()
// multiple times is safe.
//
// Returns:
//   - error: Always returns nil (safe to ignore)
//
// Example:
//
//	defer subscriber.Close()
func (s *NATSSubscriber) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil && !s.conn.IsClosed() {
		// Unsubscribe from all subscriptions
		for _, sub := range s.subscriptions {
			if sub.IsValid() {
				sub.Unsubscribe()
			}
		}
		s.subscriptions = nil

		// Drain waits for pending messages to be processed
		s.conn.Drain()
		// Close the connection
		s.conn.Close()
	}
	return nil
}
