package domain

import (
	"time"

	"github.com/google/uuid"
)

// Operation state constants
const (
	OperationStateIdle        = "IDLE"
	OperationStateCharging    = "CHARGING"
	OperationStateDischarging = "DISCHARGING"
	OperationStateFCAS        = "FCAS"
)

// BatteryState represents a snapshot of battery telemetry at a specific point in time
type BatteryState struct {
	ID               string                 // Unique ID for this state record (UUID)
	BatteryID        string                 // Reference to battery from Asset Management
	SoC              float64                // State of Charge (0-100%)
	Power            float64                // MW (positive=discharge, negative=charge)
	Temperature      float64                // °C
	Voltage          float64                // V
	Current          float64                // A
	OperationState   string                 // IDLE, CHARGING, DISCHARGING, FCAS
	CustomAttributes map[string]interface{} // Vendor-specific data (JSONB)
	Timestamp        time.Time              // When state was measured
	CreatedAt        time.Time              // When record was created
}

// NewBatteryState creates a new BatteryState with validation
func NewBatteryState(
	batteryID string,
	soc float64,
	power float64,
	temperature float64,
	voltage float64,
	current float64,
	operationState string,
	customAttributes map[string]interface{},
	timestamp time.Time,
) (*BatteryState, error) {
	// Generate UUID for ID
	id := uuid.New().String()

	// Create state
	state := &BatteryState{
		ID:               id,
		BatteryID:        batteryID,
		SoC:              soc,
		Power:            power,
		Temperature:      temperature,
		Voltage:          voltage,
		Current:          current,
		OperationState:   operationState,
		CustomAttributes: customAttributes,
		Timestamp:        timestamp,
		CreatedAt:        time.Now(),
	}

	// Validate
	if err := state.Validate(); err != nil {
		return nil, err
	}

	return state, nil
}

// Validate checks all business rules for BatteryState
func (s *BatteryState) Validate() error {
	// Rule 1: Required fields
	if s.BatteryID == "" {
		return ErrMissingRequiredField
	}

	if s.Timestamp.IsZero() {
		return ErrMissingRequiredField
	}

	// Rule 2: SoC range (0-100%)
	if s.SoC < 0 || s.SoC > 100 {
		return ErrInvalidSoC
	}

	// Rule 3: Temperature range (-20 to 60°C operating range)
	if s.Temperature < -20 || s.Temperature > 60 {
		return ErrInvalidTemperature
	}

	// Rule 4: OperationState enum
	validStates := map[string]bool{
		OperationStateIdle:        true,
		OperationStateCharging:    true,
		OperationStateDischarging: true,
		OperationStateFCAS:        true,
	}

	if !validStates[s.OperationState] {
		return ErrInvalidOperationState
	}

	// Rule 5: Power consistency with operation state
	// IDLE must have zero power
	if s.OperationState == OperationStateIdle && s.Power != 0 {
		return ErrInconsistentPowerState
	}

	// CHARGING must have negative power
	if s.OperationState == OperationStateCharging && s.Power >= 0 {
		return ErrInconsistentPowerState
	}

	// DISCHARGING must have positive power
	if s.OperationState == OperationStateDischarging && s.Power <= 0 {
		return ErrInconsistentPowerState
	}

	// FCAS can have any power (responding to frequency signals)

	// Rule 6: Timestamp not in future (allowing 1s clock skew)
	if s.Timestamp.After(time.Now().Add(1 * time.Second)) {
		return ErrInvalidTimestamp
	}

	return nil
}

// IsCharging returns true if battery is currently charging
func (s *BatteryState) IsCharging() bool {
	return s.OperationState == OperationStateCharging
}

// IsDischarging returns true if battery is currently discharging
func (s *BatteryState) IsDischarging() bool {
	return s.OperationState == OperationStateDischarging
}

// IsIdle returns true if battery is idle
func (s *BatteryState) IsIdle() bool {
	return s.OperationState == OperationStateIdle
}
