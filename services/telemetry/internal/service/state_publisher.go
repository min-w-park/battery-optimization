package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/telemetry/internal/ports"
)

// StatePublisher publishes BatteryStateChanged events at 1 Hz
type StatePublisher struct {
	repo      ports.TelemetryRepository
	publisher events.EventPublisher
	batteryID string
	log       *zap.Logger
}

// NewStatePublisher creates a new StatePublisher instance
func NewStatePublisher(
	repo ports.TelemetryRepository,
	publisher events.EventPublisher,
	batteryID string,
	log *zap.Logger,
) *StatePublisher {
	return &StatePublisher{
		repo:      repo,
		publisher: publisher,
		batteryID: batteryID,
		log:       log,
	}
}

// Start begins publishing battery state changes at 1 Hz
// Runs until the context is cancelled
func (s *StatePublisher) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // 1 Hz
	defer ticker.Stop()

	s.log.Info("state publisher started",
		zap.String("battery_id", s.batteryID),
		zap.String("frequency", "1 Hz"),
	)

	for {
		select {
		case <-ctx.Done():
			s.log.Info("state publisher stopped",
				zap.String("battery_id", s.batteryID),
			)
			return
		case <-ticker.C:
			if err := s.publishOnce(ctx); err != nil {
				// Best-effort: log error but don't crash
				s.log.Warn("failed to publish battery state",
					zap.String("battery_id", s.batteryID),
					zap.Error(err),
				)
			}
		}
	}
}

// publishOnce retrieves current state and publishes event
func (s *StatePublisher) publishOnce(ctx context.Context) error {
	// Get current state from repository
	state, err := s.repo.GetCurrentState(ctx, s.batteryID)
	if err != nil {
		return err
	}

	// Map to BatteryStateChanged event
	event := events.BatteryStateChanged{
		BatteryID:        state.BatteryID,
		SoC:              state.SoC,
		Power:            state.Power,
		Temperature:      state.Temperature,
		Voltage:          state.Voltage,
		Current:          state.Current,
		OperationState:   state.OperationState,
		CustomAttributes: state.CustomAttributes,
		Timestamp:        state.Timestamp,
		EventVersion:     "v1",
	}

	// Publish to NATS with timeout
	pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.publisher.Publish(pubCtx, "battery.state.changed.v1", event); err != nil {
		return fmt.Errorf("failed to publish BatteryStateChanged event: %w", err)
	}

	return nil
}
