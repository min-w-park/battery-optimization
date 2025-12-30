package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSPublisher implements EventPublisher for NATS message broker.
type NATSPublisher struct {
	conn *nats.Conn
}

// NewNATSPublisher creates a new NATS publisher.
//
// Parameters:
//   - url: NATS server URL (e.g., "nats://localhost:4222")
//
// Returns:
//   - *NATSPublisher: Publisher instance
//   - error: Connection error if NATS is unreachable
//
// Example:
//
//	publisher, err := events.NewNATSPublisher("nats://localhost:4222")
//	if err != nil {
//	    log.Fatalf("Failed to connect to NATS: %v", err)
//	}
//	defer publisher.Close()
func NewNATSPublisher(url string) (*NATSPublisher, error) {
	// Connect to NATS with timeout
	conn, err := nats.Connect(
		url,
		nats.Timeout(10*time.Second),
		nats.Name("event-publisher"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", url, err)
	}

	return &NATSPublisher{
		conn: conn,
	}, nil
}

// Publish sends an event to the specified NATS subject.
//
// The event is JSON-marshaled before publishing. If marshaling fails,
// an error is returned immediately. If NATS publish fails (e.g., connection
// lost), an error is returned.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - subject: NATS subject (e.g., "battery.registered.v1")
//   - event: Event struct to publish (will be JSON-marshaled)
//
// Returns:
//   - error: JSON marshaling error or NATS publish error
//
// Example:
//
//	event := events.BatteryRegistered{...}
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	err := publisher.Publish(ctx, "battery.registered.v1", event)
func (p *NATSPublisher) Publish(ctx context.Context, subject string, event interface{}) error {
	// Check if connection is closed
	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("NATS connection is closed")
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	// Publish to NATS
	// Note: NATS publish is synchronous by default (fire-and-forget)
	// For guaranteed delivery, use Request() or JetStream
	err = p.conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish to subject %s: %w", subject, err)
	}

	// Flush to ensure message is sent
	// Use FlushWithContext if context has deadline, otherwise use Flush
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		if err := p.conn.FlushWithContext(ctx); err != nil {
			return fmt.Errorf("failed to flush message to NATS: %w", err)
		}
	} else {
		// Use simple Flush() for contexts without deadline
		if err := p.conn.Flush(); err != nil {
			return fmt.Errorf("failed to flush message to NATS: %w", err)
		}
	}

	return nil
}

// Close gracefully shuts down the publisher.
//
// It drains pending messages (waits for them to be sent) and then
// closes the NATS connection. Calling Close() multiple times is safe.
//
// Returns:
//   - error: Always returns nil (safe to ignore)
//
// Example:
//
//	defer publisher.Close()
func (p *NATSPublisher) Close() error {
	if p.conn != nil && !p.conn.IsClosed() {
		// Drain waits for pending messages to be sent
		p.conn.Drain()
		// Close the connection
		p.conn.Close()
	}
	return nil
}
