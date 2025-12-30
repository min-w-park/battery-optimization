package events

import "time"

// ChargingCompleted is published when battery charging stops
// Publisher: Device Interface Service
// Subscribers: Bidding Service, Operator Dashboard, Economics Service
type ChargingCompleted struct {
	BatteryID       string    `json:"battery_id"`
	CommandID       string    `json:"command_id"`
	InitialSoC      float64   `json:"initial_soc"`      // 0-100%
	FinalSoC        float64   `json:"final_soc"`        // 0-100%
	EnergyCharged   float64   `json:"energy_charged"`   // MWh
	DurationSeconds int       `json:"duration_seconds"` // Actual duration
	Reason          string    `json:"reason"`           // TARGET_REACHED | COMMAND_STOPPED | ERROR
	Timestamp       time.Time `json:"timestamp"`
	EventVersion    string    `json:"event_version"` // "v1"
}
