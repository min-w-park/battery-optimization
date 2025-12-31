package events

import "time"

// ChargingCommandIssued is published when a charging command is issued
// Publisher: Bidding Service (if FULL_AUTO) or Operator Service (if MANUAL/SEMI_AUTO)
// Trigger: Charging opportunity detected AND automation mode allows command execution
type ChargingCommandIssued struct {
	BatteryID    string    `json:"battery_id"`
	TargetSoC    float64   `json:"target_soc"` // Target SoC to charge to (%)
	MaxPower     float64   `json:"max_power"`  // Maximum charging power (MW)
	Reason       string    `json:"reason"`     // Human-readable explanation
	IssuedBy     string    `json:"issued_by"`  // Who issued the command (BIDDING_SERVICE_AUTO or operator ID)
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"`
}
