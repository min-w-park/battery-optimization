package events

import "time"

// CommandStopConditions defines conditions for stopping a discharge operation
type CommandStopConditions struct {
	PriceThreshold float64 `json:"price_threshold"` // Stop if price drops below this ($/MWh)
	TargetSoC      float64 `json:"target_soc"`      // Stop if SoC reaches this (%)
	Duration       int     `json:"duration"`        // Stop after this many minutes (0 = no limit)
	FcasDispatch   bool    `json:"fcas_dispatch"`   // Stop if FCAS dispatch received
}

// DischargingCommandIssued is published when a discharging command is issued
// Publisher: Bidding Service (if FULL_AUTO) or Operator Service (if MANUAL/SEMI_AUTO)
// Trigger: Discharging opportunity detected AND automation mode allows command execution
type DischargingCommandIssued struct {
	BatteryID      string                `json:"battery_id"`
	CommandID      string                `json:"command_id"`      // Unique command identifier
	Power          float64               `json:"power"`           // Discharge power (MW, positive for discharge)
	Duration       int                   `json:"duration"`        // Command duration in minutes
	StopConditions CommandStopConditions `json:"stop_conditions"` // Conditions for stopping discharge
	Reason         string                `json:"reason"`          // Human-readable explanation
	IssuedBy       string                `json:"issued_by"`       // Who issued the command
	Timestamp      time.Time             `json:"timestamp"`
	EventVersion   string                `json:"event_version"`
}
