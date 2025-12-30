package domain

import "errors"

// Domain errors for Telemetry Service
var (
	// ErrInvalidSoC indicates State of Charge is outside valid range (0-100%)
	ErrInvalidSoC = errors.New("invalid SoC value")

	// ErrInvalidTemperature indicates temperature is outside operating range (-20 to 60°C)
	ErrInvalidTemperature = errors.New("invalid temperature value")

	// ErrInvalidOperationState indicates operation state is not a valid enum value
	ErrInvalidOperationState = errors.New("invalid operation state")

	// ErrInconsistentPowerState indicates power sign doesn't match operation state
	ErrInconsistentPowerState = errors.New("power value inconsistent with operation state")

	// ErrMissingRequiredField indicates a required field is missing or empty
	ErrMissingRequiredField = errors.New("required field is missing")

	// ErrInvalidTimestamp indicates timestamp is invalid (e.g., in the future)
	ErrInvalidTimestamp = errors.New("timestamp is invalid")

	// ErrBatteryNotFound indicates no telemetry data exists for the battery
	ErrBatteryNotFound = errors.New("battery not found")
)
