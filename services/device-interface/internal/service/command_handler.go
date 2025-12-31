package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/device-interface/internal/domain"
)

// CommandHandler processes command events and controls the battery adapter
type CommandHandler struct {
	adapter   domain.BatteryAdapter
	publisher events.EventPublisher
	log       *zap.Logger
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(adapter domain.BatteryAdapter, publisher events.EventPublisher, log *zap.Logger) *CommandHandler {
	return &CommandHandler{
		adapter:   adapter,
		publisher: publisher,
		log:       log,
	}
}

// HandleChargingCommand processes a charging command event
func (h *CommandHandler) HandleChargingCommand(ctx context.Context, event events.ChargingCommandIssued) error {
	h.log.Info("handling charging command",
		zap.String("battery_id", event.BatteryID),
		zap.String("command_id", event.CommandID),
		zap.Float64("power_mw", event.Power),
		zap.Float64("target_soc", event.TargetSoC),
	)

	// Get current state
	state, err := h.adapter.GetState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get battery state: %w", err)
	}

	// Create command
	cmd := domain.Command{
		Type:      domain.CommandCharge,
		Power:     event.Power,
		TargetSoC: event.TargetSoC,
		Duration:  time.Duration(event.Duration) * time.Minute,
	}

	// Send command to adapter
	if err := h.adapter.SendCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to send charging command: %w", err)
	}

	// Publish ChargingStarted event
	startedEvent := events.ChargingStarted{
		BatteryID:    event.BatteryID,
		CommandID:    event.CommandID,
		ActualPower:  state.Power, // Current power (will ramp to target)
		TargetSoC:    event.TargetSoC,
		CurrentSoC:   state.SoC,
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.publisher.Publish(pubCtx, "charging.started.v1", startedEvent); err != nil {
		h.log.Warn("failed to publish charging.started.v1 event", zap.Error(err))
	}

	h.log.Info("charging command accepted",
		zap.Float64("current_soc", state.SoC),
		zap.Float64("target_soc", event.TargetSoC),
	)
	return nil
}

// HandleDischargingCommand processes a discharging command event
func (h *CommandHandler) HandleDischargingCommand(ctx context.Context, event events.DischargingCommandIssued) error {
	h.log.Info("handling discharging command",
		zap.String("battery_id", event.BatteryID),
		zap.String("command_id", event.CommandID),
		zap.Float64("power_mw", event.Power),
	)

	// Get current state
	state, err := h.adapter.GetState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get battery state: %w", err)
	}

	// Create command
	cmd := domain.Command{
		Type:     domain.CommandDischarge,
		Power:    event.Power,
		Duration: time.Duration(event.Duration) * time.Minute,
	}

	// Send command to adapter
	if err := h.adapter.SendCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to send discharging command: %w", err)
	}

	// Publish DischargingStarted event
	startedEvent := events.DischargingStarted{
		BatteryID:    event.BatteryID,
		CommandID:    event.CommandID,
		ActualPower:  state.Power, // Current power (will ramp to target)
		CurrentSoC:   state.SoC,
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.publisher.Publish(pubCtx, "discharging.started.v1", startedEvent); err != nil {
		h.log.Warn("failed to publish discharging.started.v1 event", zap.Error(err))
	}

	h.log.Info("discharging command accepted", zap.Float64("current_soc", state.SoC))
	return nil
}

// HandleConflictResolved processes a conflict resolution event
func (h *CommandHandler) HandleConflictResolved(ctx context.Context, event events.ConflictResolved) error {
	h.log.Info("handling conflict resolved",
		zap.String("conflict_id", event.ConflictID),
		zap.String("chosen_action", event.ChosenAction),
	)

	var cmd domain.Command

	switch event.ChosenAction {
	case "SWITCH_TO_CHARGING":
		cmd = domain.Command{
			Type:  domain.CommandCharge,
			Power: 0, // Will be specified by subsequent ChargingCommandIssued
		}
	case "CONTINUE_CURRENT":
		// Return to idle state
		cmd = domain.Command{
			Type: domain.CommandIdle,
		}
	default:
		return fmt.Errorf("unknown action: %s", event.ChosenAction)
	}

	// Send command to adapter
	if err := h.adapter.SendCommand(ctx, cmd); err != nil {
		return fmt.Errorf("failed to send conflict resolution command: %w", err)
	}

	h.log.Info("conflict resolved", zap.String("chosen_action", event.ChosenAction))
	return nil
}

// OnEvent is the event handler callback for NATS subscriber
func (h *CommandHandler) OnEvent(subject string, data []byte) error {
	ctx := context.Background()

	switch subject {
	case "charging.command.issued.v1":
		var event events.ChargingCommandIssued
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal ChargingCommandIssued: %w", err)
		}
		return h.HandleChargingCommand(ctx, event)

	case "discharging.command.issued.v1":
		var event events.DischargingCommandIssued
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal DischargingCommandIssued: %w", err)
		}
		return h.HandleDischargingCommand(ctx, event)

	case "conflict.resolved.v1":
		var event events.ConflictResolved
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal ConflictResolved: %w", err)
		}
		return h.HandleConflictResolved(ctx, event)

	default:
		h.log.Warn("unknown event subject", zap.String("subject", subject))
		return nil
	}
}
