package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/device-interface/internal/domain"
)

// MockBatteryAdapter is a mock implementation of domain.BatteryAdapter
type MockBatteryAdapter struct {
	mock.Mock
}

func (m *MockBatteryAdapter) GetState(ctx context.Context) (domain.BatteryState, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.BatteryState), args.Error(1)
}

func (m *MockBatteryAdapter) SendCommand(ctx context.Context, cmd domain.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockBatteryAdapter) GetCustomAttributes() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockBatteryAdapter) GetBatteryID() string {
	args := m.Called()
	return args.String(0)
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

func TestNewCommandHandler(t *testing.T) {
	// Given: Mock dependencies
	adapter := new(MockBatteryAdapter)
	publisher := new(MockEventPublisher)

	// When: Creating command handler
	handler := NewCommandHandler(adapter, publisher)

	// Then: Handler is created correctly
	assert.NotNil(t, handler)
	assert.Equal(t, adapter, handler.adapter)
	assert.Equal(t, publisher, handler.publisher)
}

func TestCommandHandler_HandleChargingCommand(t *testing.T) {
	// Given: Mock dependencies
	adapter := new(MockBatteryAdapter)
	publisher := new(MockEventPublisher)
	handler := NewCommandHandler(adapter, publisher)

	// Mock adapter state
	adapter.On("GetState", mock.Anything).Return(domain.BatteryState{
		SoC:            45.0,
		Power:          0.0,
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		OperationState: domain.OperationStateIdle,
	}, nil)

	// Mock command execution
	adapter.On("SendCommand", mock.Anything, mock.MatchedBy(func(cmd domain.Command) bool {
		return cmd.Type == domain.CommandCharge && cmd.Power == 15.0
	})).Return(nil)

	// Mock event publishing
	publisher.On("Publish", mock.Anything, "charging.started.v1", mock.Anything).Return(nil)

	// When: Handling charging command event
	eventData := events.ChargingCommandIssued{
		BatteryID:    "battery-123",
		CommandID:    "cmd-456",
		Power:        15.0,
		TargetSoC:    80.0,
		Duration:     60,
		DecisionMode: "FULL_AUTO",
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	ctx := context.Background()
	err := handler.HandleChargingCommand(ctx, eventData)

	// Then: Command processed successfully
	require.NoError(t, err)
	adapter.AssertExpectations(t)
	publisher.AssertExpectations(t)

	// Verify ChargingStarted event was published
	require.Len(t, publisher.PublishedEvents, 1)
	assert.Equal(t, "charging.started.v1", publisher.PublishedEvents[0].Subject)

	startedEvent, ok := publisher.PublishedEvents[0].Event.(events.ChargingStarted)
	require.True(t, ok, "Event should be ChargingStarted")
	assert.Equal(t, "battery-123", startedEvent.BatteryID)
	assert.Equal(t, "cmd-456", startedEvent.CommandID)
	assert.Equal(t, 45.0, startedEvent.CurrentSoC)
	assert.Equal(t, 80.0, startedEvent.TargetSoC)
}

func TestCommandHandler_HandleDischargingCommand(t *testing.T) {
	// Given: Mock dependencies
	adapter := new(MockBatteryAdapter)
	publisher := new(MockEventPublisher)
	handler := NewCommandHandler(adapter, publisher)

	// Mock adapter state
	adapter.On("GetState", mock.Anything).Return(domain.BatteryState{
		SoC:            75.0,
		Power:          0.0,
		Temperature:    25.0,
		Voltage:        820.0,
		Current:        0.0,
		OperationState: domain.OperationStateIdle,
	}, nil)

	// Mock command execution
	adapter.On("SendCommand", mock.Anything, mock.MatchedBy(func(cmd domain.Command) bool {
		return cmd.Type == domain.CommandDischarge && cmd.Power == 25.0
	})).Return(nil)

	// Mock event publishing
	publisher.On("Publish", mock.Anything, "discharging.started.v1", mock.Anything).Return(nil)

	// When: Handling discharging command event
	eventData := events.DischargingCommandIssued{
		BatteryID:    "battery-123",
		CommandID:    "cmd-789",
		Power:        25.0,
		Duration:     30,
		DecisionMode: "SEMI_AUTO",
		ApprovedBy:   "operator-001",
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	ctx := context.Background()
	err := handler.HandleDischargingCommand(ctx, eventData)

	// Then: Command processed successfully
	require.NoError(t, err)
	adapter.AssertExpectations(t)
	publisher.AssertExpectations(t)

	// Verify DischargingStarted event was published
	require.Len(t, publisher.PublishedEvents, 1)
	assert.Equal(t, "discharging.started.v1", publisher.PublishedEvents[0].Subject)

	startedEvent, ok := publisher.PublishedEvents[0].Event.(events.DischargingStarted)
	require.True(t, ok, "Event should be DischargingStarted")
	assert.Equal(t, "battery-123", startedEvent.BatteryID)
	assert.Equal(t, "cmd-789", startedEvent.CommandID)
	assert.Equal(t, 75.0, startedEvent.CurrentSoC)
}

func TestCommandHandler_HandleConflictResolved(t *testing.T) {
	tests := []struct {
		name            string
		chosenAction    string
		expectedCmdType domain.CommandType
	}{
		{
			name:            "Switch to charging",
			chosenAction:    "SWITCH_TO_CHARGING",
			expectedCmdType: domain.CommandCharge,
		},
		{
			name:            "Continue current (idle)",
			chosenAction:    "CONTINUE_CURRENT",
			expectedCmdType: domain.CommandIdle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: Mock dependencies
			adapter := new(MockBatteryAdapter)
			publisher := new(MockEventPublisher)
			handler := NewCommandHandler(adapter, publisher)

			// Mock command execution
			adapter.On("SendCommand", mock.Anything, mock.MatchedBy(func(cmd domain.Command) bool {
				return cmd.Type == tt.expectedCmdType
			})).Return(nil)

			// When: Handling conflict resolved event
			eventData := events.ConflictResolved{
				ConflictID:    "conflict-123",
				ChosenAction:  tt.chosenAction,
				DecisionMaker: "AUTO",
				Timestamp:     time.Now(),
				EventVersion:  "v1",
			}

			ctx := context.Background()
			err := handler.HandleConflictResolved(ctx, eventData)

			// Then: Command processed successfully
			require.NoError(t, err)
			adapter.AssertExpectations(t)
		})
	}
}

func TestCommandHandler_OnEvent(t *testing.T) {
	// Given: Mock dependencies and handler
	adapter := new(MockBatteryAdapter)
	publisher := new(MockEventPublisher)
	handler := NewCommandHandler(adapter, publisher)

	tests := []struct {
		name    string
		subject string
		event   interface{}
		setup   func()
	}{
		{
			name:    "Charging command event",
			subject: "charging.command.issued.v1",
			event: events.ChargingCommandIssued{
				BatteryID:    "battery-123",
				CommandID:    "cmd-001",
				Power:        20.0,
				TargetSoC:    90.0,
				EventVersion: "v1",
			},
			setup: func() {
				adapter.On("GetState", mock.Anything).Return(domain.BatteryState{SoC: 50.0}, nil)
				adapter.On("SendCommand", mock.Anything, mock.Anything).Return(nil)
				publisher.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "Discharging command event",
			subject: "discharging.command.issued.v1",
			event: events.DischargingCommandIssued{
				BatteryID:    "battery-123",
				CommandID:    "cmd-002",
				Power:        30.0,
				EventVersion: "v1",
			},
			setup: func() {
				adapter.On("GetState", mock.Anything).Return(domain.BatteryState{SoC: 80.0}, nil)
				adapter.On("SendCommand", mock.Anything, mock.Anything).Return(nil)
				publisher.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "Conflict resolved event",
			subject: "conflict.resolved.v1",
			event: events.ConflictResolved{
				ConflictID:   "conflict-001",
				ChosenAction: "CONTINUE_CURRENT",
				EventVersion: "v1",
			},
			setup: func() {
				adapter.On("SendCommand", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			adapter = new(MockBatteryAdapter)
			publisher = new(MockEventPublisher)
			handler = NewCommandHandler(adapter, publisher)

			// Setup expectations
			tt.setup()

			// Marshal event to JSON
			data, err := json.Marshal(tt.event)
			require.NoError(t, err)

			// When: Processing event through OnEvent
			err = handler.OnEvent(tt.subject, data)

			// Then: Event processed successfully
			require.NoError(t, err)
		})
	}
}
