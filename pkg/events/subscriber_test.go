package events_test

import (
	"context"
	"testing"

	"github.com/minwook/battery-optimization/pkg/events"
)

func TestEventSubscriber_Interface(t *testing.T) {
	// This test verifies the interface exists with correct signature
	// The actual implementation tests will be in nats_subscriber_test.go
	var _ events.EventSubscriber = (*mockSubscriber)(nil)
}

// mockSubscriber is a test implementation to verify interface compliance
type mockSubscriber struct{}

func (m *mockSubscriber) Subscribe(ctx context.Context, subject string, handler events.EventHandler) error {
	return nil
}

func (m *mockSubscriber) Close() error {
	return nil
}
