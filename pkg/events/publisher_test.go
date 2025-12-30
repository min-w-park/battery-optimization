package events_test

import (
	"context"
	"testing"

	"github.com/minwook/battery-optimization/pkg/events"
)

func TestEventPublisher_Interface(t *testing.T) {
	// This test verifies the interface exists with correct signature
	// The actual implementation tests will be in nats_publisher_test.go
	var _ events.EventPublisher = (*mockPublisher)(nil)
}

// mockPublisher is a test implementation to verify interface compliance
type mockPublisher struct{}

func (m *mockPublisher) Publish(ctx context.Context, subject string, event interface{}) error {
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}
