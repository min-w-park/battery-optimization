package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
	"github.com/minwook/battery-optimization/services/telemetry/internal/ports"
)

// MockTelemetryRepository is a mock implementation of ports.TelemetryRepository
type MockTelemetryRepository struct {
	mock.Mock
}

func (m *MockTelemetryRepository) SaveState(ctx context.Context, state *domain.BatteryState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *MockTelemetryRepository) GetCurrentState(ctx context.Context, batteryID string) (*domain.BatteryState, error) {
	args := m.Called(ctx, batteryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BatteryState), args.Error(1)
}

func (m *MockTelemetryRepository) GetHistory(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error) {
	args := m.Called(ctx, batteryID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BatteryState), args.Error(1)
}

// MockEventPublisher is a mock implementation of events.EventPublisher
type MockEventPublisher struct {
	mock.Mock
	PublishedEvents []PublishedEvent
}

type PublishedEvent struct {
	Subject string
	Event   interface{}
}

func (m *MockEventPublisher) Publish(ctx context.Context, subject string, event interface{}) error {
	m.PublishedEvents = append(m.PublishedEvents, PublishedEvent{Subject: subject, Event: event})
	args := m.Called(ctx, subject, event)
	return args.Error(0)
}

func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewStatePublisher(t *testing.T) {
	// Given: Mock dependencies
	repo := new(MockTelemetryRepository)
	publisher := new(MockEventPublisher)
	batteryID := "battery-123"

	// When: Creating state publisher
	sp := NewStatePublisher(repo, publisher, batteryID, zap.NewNop())

	// Then: Publisher is created correctly
	assert.NotNil(t, sp)
	assert.Equal(t, repo, sp.repo)
	assert.Equal(t, publisher, sp.publisher)
	assert.Equal(t, batteryID, sp.batteryID)
}

func TestStatePublisher_PublishOnce_Success(t *testing.T) {
	// Given: Mock dependencies
	repo := new(MockTelemetryRepository)
	publisher := new(MockEventPublisher)
	batteryID := "battery-123"

	// Mock repository to return a battery state
	testState := &domain.BatteryState{
		ID:            "state-1",
		BatteryID:     batteryID,
		SoC:           75.5,
		Power:         25.0,
		Temperature:   28.5,
		Voltage:       825.0,
		Current:       30303.0,
		OperationState: domain.OperationStateDischarging,
		CustomAttributes: map[string]interface{}{
			"vendor": "Tesla",
			"model":  "Megapack",
		},
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}
	repo.On("GetCurrentState", mock.Anything, batteryID).Return(testState, nil)

	// Mock publisher to succeed
	publisher.On("Publish", mock.Anything, "battery.state.changed.v1", mock.Anything).Return(nil)

	sp := NewStatePublisher(repo, publisher, batteryID, zap.NewNop())

	// When: Publishing once
	ctx := context.Background()
	err := sp.publishOnce(ctx)

	// Then: No error and event published
	require.NoError(t, err)
	repo.AssertExpectations(t)
	publisher.AssertExpectations(t)

	// Verify published event
	require.Len(t, publisher.PublishedEvents, 1)
	assert.Equal(t, "battery.state.changed.v1", publisher.PublishedEvents[0].Subject)

	event, ok := publisher.PublishedEvents[0].Event.(events.BatteryStateChanged)
	require.True(t, ok, "Event should be BatteryStateChanged")
	assert.Equal(t, batteryID, event.BatteryID)
	assert.Equal(t, 75.5, event.SoC)
	assert.Equal(t, 25.0, event.Power)
	assert.Equal(t, 28.5, event.Temperature)
	assert.Equal(t, 825.0, event.Voltage)
	assert.Equal(t, 30303.0, event.Current)
	assert.Equal(t, "DISCHARGING", event.OperationState)
	assert.Equal(t, "v1", event.EventVersion)
}

func TestStatePublisher_PublishOnce_NoState(t *testing.T) {
	// Given: Mock dependencies
	repo := new(MockTelemetryRepository)
	publisher := new(MockEventPublisher)
	batteryID := "battery-123"

	// Mock repository to return no state (battery not found)
	repo.On("GetCurrentState", mock.Anything, batteryID).Return(nil, domain.ErrBatteryNotFound)

	sp := NewStatePublisher(repo, publisher, batteryID, zap.NewNop())

	// When: Publishing once with no state
	ctx := context.Background()
	err := sp.publishOnce(ctx)

	// Then: Error is returned (should not publish)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrBatteryNotFound, err)
	repo.AssertExpectations(t)
	// Publisher should not be called
	assert.Len(t, publisher.PublishedEvents, 0)
}

func TestStatePublisher_PublishOnce_PublishError(t *testing.T) {
	// Given: Mock dependencies
	repo := new(MockTelemetryRepository)
	publisher := new(MockEventPublisher)
	batteryID := "battery-123"

	testState := &domain.BatteryState{
		ID:            "state-1",
		BatteryID:     batteryID,
		SoC:           50.0,
		Power:         0.0,
		Temperature:   25.0,
		Voltage:       800.0,
		Current:       0.0,
		OperationState: domain.OperationStateIdle,
		Timestamp:     time.Now(),
		CreatedAt:     time.Now(),
	}
	repo.On("GetCurrentState", mock.Anything, batteryID).Return(testState, nil)

	// Mock publisher to fail
	publishErr := errors.New("NATS connection lost")
	publisher.On("Publish", mock.Anything, "battery.state.changed.v1", mock.Anything).Return(publishErr)

	sp := NewStatePublisher(repo, publisher, batteryID, zap.NewNop())

	// When: Publishing fails
	ctx := context.Background()
	err := sp.publishOnce(ctx)

	// Then: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish")
	repo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestStatePublisher_Start_ContextCancellation(t *testing.T) {
	// Given: Mock dependencies
	repo := new(MockTelemetryRepository)
	publisher := new(MockEventPublisher)
	batteryID := "battery-123"

	testState := &domain.BatteryState{
		ID:            "state-1",
		BatteryID:     batteryID,
		SoC:           50.0,
		Power:         0.0,
		Temperature:   25.0,
		Voltage:       800.0,
		Current:       0.0,
		OperationState: domain.OperationStateIdle,
		Timestamp:     time.Now(),
		CreatedAt:     time.Now(),
	}

	// Allow multiple calls
	repo.On("GetCurrentState", mock.Anything, batteryID).Return(testState, nil).Maybe()
	publisher.On("Publish", mock.Anything, "battery.state.changed.v1", mock.Anything).Return(nil).Maybe()

	sp := NewStatePublisher(repo, publisher, batteryID, zap.NewNop())

	// When: Starting with cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start in goroutine
	done := make(chan bool)
	go func() {
		sp.Start(ctx)
		done <- true
	}()

	// Let it run long enough for at least one tick (ticker fires after 1 second)
	time.Sleep(1200 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for completion (with timeout)
	select {
	case <-done:
		// Success - Start() returned
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not exit after context cancellation")
	}

	// Then: Should have published at least once (ticker fired after 1 second)
	assert.GreaterOrEqual(t, len(publisher.PublishedEvents), 1, "Should have published at least one event")
}

func TestStatePublisher_Start_TickerFrequency(t *testing.T) {
	// Given: Mock dependencies
	repo := new(MockTelemetryRepository)
	publisher := new(MockEventPublisher)
	batteryID := "battery-123"

	testState := &domain.BatteryState{
		ID:            "state-1",
		BatteryID:     batteryID,
		SoC:           50.0,
		Power:         0.0,
		Temperature:   25.0,
		Voltage:       800.0,
		Current:       0.0,
		OperationState: domain.OperationStateIdle,
		Timestamp:     time.Now(),
		CreatedAt:     time.Now(),
	}

	repo.On("GetCurrentState", mock.Anything, batteryID).Return(testState, nil).Maybe()
	publisher.On("Publish", mock.Anything, "battery.state.changed.v1", mock.Anything).Return(nil).Maybe()

	sp := NewStatePublisher(repo, publisher, batteryID, zap.NewNop())

	// When: Running for ~2.5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	go sp.Start(ctx)

	// Wait for context timeout
	<-ctx.Done()

	// Give it a moment to finish
	time.Sleep(100 * time.Millisecond)

	// Then: Should have published approximately 2-3 times (1 Hz = 1 per second)
	// Allow some tolerance for timing
	eventCount := len(publisher.PublishedEvents)
	assert.GreaterOrEqual(t, eventCount, 2, "Should publish at least 2 events in 2.5 seconds")
	assert.LessOrEqual(t, eventCount, 4, "Should not publish more than 4 events in 2.5 seconds")
}
