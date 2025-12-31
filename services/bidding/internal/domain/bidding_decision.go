package domain

import (
	"time"

	"github.com/google/uuid"
)

// DecisionType represents the type of bidding decision
type DecisionType string

const (
	DecisionCharge    DecisionType = "CHARGE"
	DecisionDischarge DecisionType = "DISCHARGE"
	DecisionNoAction  DecisionType = "NO_ACTION"
)

// AutomationMode represents the level of automation for decision execution
type AutomationMode string

const (
	ModeManual   AutomationMode = "MANUAL"    // Detect only, no command
	ModeSemiAuto AutomationMode = "SEMI_AUTO" // Detect + suggest, wait for approval
	ModeFullAuto AutomationMode = "FULL_AUTO" // Detect + execute immediately
)

// OperationState represents the current state of battery operation
type OperationState string

const (
	StateIdle        OperationState = "IDLE"
	StateCharging    OperationState = "CHARGING"
	StateDischarging OperationState = "DISCHARGING"
	StateFCAS        OperationState = "FCAS"
)

// BiddingDecision represents an arbitrage decision moment
type BiddingDecision struct {
	ID             string         // UUID
	BatteryID      string         // Which battery this decision is for
	DecisionType   DecisionType   // CHARGE, DISCHARGE, NO_ACTION
	Reason         string         // Human-readable explanation
	Price          float64        // Market price at decision time ($/MWh)
	SoC            float64        // Battery SoC at decision time (%)
	Timestamp      time.Time      // When decision was made
	AutomationMode AutomationMode // MANUAL, SEMI_AUTO, FULL_AUTO
}

// NewBiddingDecision creates a new bidding decision with validation
func NewBiddingDecision(
	batteryID string,
	decisionType string,
	reason string,
	price float64,
	soc float64,
	automationMode string,
) (*BiddingDecision, error) {
	// Validate battery ID
	if batteryID == "" {
		return nil, ErrEmptyBatteryID
	}

	// Validate price
	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	// Validate SoC
	if soc < 0 || soc > 100 {
		return nil, ErrInvalidSoC
	}

	// Validate decision type
	if !isValidDecisionType(decisionType) {
		return nil, ErrInvalidDecisionType
	}

	// Validate automation mode
	if !isValidAutomationMode(automationMode) {
		return nil, ErrInvalidAutomationMode
	}

	return &BiddingDecision{
		ID:             uuid.New().String(),
		BatteryID:      batteryID,
		DecisionType:   DecisionType(decisionType),
		Reason:         reason,
		Price:          price,
		SoC:            soc,
		Timestamp:      time.Now(),
		AutomationMode: AutomationMode(automationMode),
	}, nil
}

// isValidDecisionType checks if decision type is valid
func isValidDecisionType(dt string) bool {
	switch dt {
	case "CHARGE", "DISCHARGE", "NO_ACTION":
		return true
	default:
		return false
	}
}

// isValidAutomationMode checks if automation mode is valid
func isValidAutomationMode(mode string) bool {
	switch mode {
	case "MANUAL", "SEMI_AUTO", "FULL_AUTO":
		return true
	default:
		return false
	}
}
