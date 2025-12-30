package domain

import (
	"context"
	"time"
)

// BatteryAdapter is the core abstraction for hardware access
type BatteryAdapter interface {
	// GetState retrieves current battery state from hardware
	GetState(ctx context.Context) (BatteryState, error)

	// SendCommand sends a command to the battery hardware
	SendCommand(ctx context.Context, cmd Command) error

	// GetCustomAttributes returns manufacturer-specific attributes
	GetCustomAttributes() map[string]interface{}

	// GetBatteryID returns the battery this adapter controls
	GetBatteryID() string
}

// BatteryState represents hardware readings (simpler than Telemetry domain)
// No ID, Timestamp, or CreatedAt - those are added by Telemetry Service
type BatteryState struct {
	SoC            float64 // State of Charge (0-100%)
	Power          float64 // MW (positive=discharge, negative=charge)
	Temperature    float64 // °C
	Voltage        float64 // V
	Current        float64 // A
	OperationState string  // IDLE, CHARGING, DISCHARGING, FCAS
}

// CommandType represents the type of command to send to the battery
type CommandType int

const (
	CommandCharge       CommandType = iota // Start charging
	CommandDischarge                       // Start discharging
	CommandIdle                            // Stop all operations
	CommandFcasResponse                    // Respond to FCAS dispatch
)

// Command represents a command to send to the battery hardware
type Command struct {
	Type      CommandType   // What to do
	Power     float64       // MW (absolute value)
	TargetSoC float64       // 0-100% (for charging, optional)
	Duration  time.Duration // Max duration (0 = indefinite)
}

// Operation state constants
const (
	OperationStateIdle        = "IDLE"
	OperationStateCharging    = "CHARGING"
	OperationStateDischarging = "DISCHARGING"
	OperationStateFCAS        = "FCAS"
)
