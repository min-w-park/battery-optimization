package events

import "time"

// ChargingCommandIssued is published when a charging command is issued
// Publisher: Bidding Service (auto mode) or Operator Service (manual/semi-auto)
// Subscribers: Device Interface Service, Conflict Resolver
type ChargingCommandIssued struct {
	BatteryID    string  `json:"battery_id"`
	CommandID    string  `json:"command_id"`    // UUID
	Power        float64 `json:"power"`         // MW
	TargetSoC    float64 `json:"target_soc"`    // 0-100%
	Duration     int     `json:"duration"`      // minutes
	DecisionMode string  `json:"decision_mode"` // MANUAL | SEMI_AUTO | FULL_AUTO
	ApprovedBy   string  `json:"approved_by,omitempty"` // operator ID if manual
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
