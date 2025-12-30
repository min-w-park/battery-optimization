package events

import "time"

// DischargingCommandIssued is published when a discharging command is issued
// Publisher: Bidding Service (auto mode) or Operator Service (manual/semi-auto)
// Subscribers: Device Interface Service, Conflict Resolver
type DischargingCommandIssued struct {
	BatteryID       string  `json:"battery_id"`
	CommandID       string  `json:"command_id"`        // UUID
	Power           float64 `json:"power"`             // MW
	Duration        int     `json:"duration"`          // minutes (0 = indefinite)
	StopConditions  StopConditions `json:"stop_conditions"`
	DecisionMode    string  `json:"decision_mode"`     // MANUAL | SEMI_AUTO | FULL_AUTO
	ApprovedBy      string  `json:"approved_by,omitempty"` // operator ID if manual
	Timestamp       time.Time `json:"timestamp"`
	EventVersion    string    `json:"event_version"` // "v1"
}

// StopConditions defines when discharging should stop
type StopConditions struct {
	PriceThreshold  float64 `json:"price_threshold,omitempty"`  // Stop if price < this ($/MWh)
	MinSoC          float64 `json:"min_soc,omitempty"`          // Stop if SoC < this (0-100%)
	FcasDispatch    bool    `json:"fcas_dispatch,omitempty"`    // Stop on FCAS dispatch
	DurationMinutes int     `json:"duration_minutes,omitempty"` // Stop after X minutes
}
