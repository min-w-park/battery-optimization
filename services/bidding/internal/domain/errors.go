package domain

import "errors"

var (
	// Validation errors
	ErrInvalidSoC            = errors.New("SoC must be between 0 and 100")
	ErrInvalidPrice          = errors.New("price must be greater than 0")
	ErrInvalidDecisionType   = errors.New("decision type must be CHARGE, DISCHARGE, or NO_ACTION")
	ErrInvalidAutomationMode = errors.New("automation mode must be MANUAL, SEMI_AUTO, or FULL_AUTO")
	ErrInvalidOperationState = errors.New("operation state must be IDLE, CHARGING, DISCHARGING, or FCAS")

	// Business rule errors
	ErrBatteryNotReady = errors.New("battery not ready for operation")
	ErrBatteryNotIdle  = errors.New("battery must be in IDLE state")
	ErrEmptyBatteryID  = errors.New("battery ID cannot be empty")
)
