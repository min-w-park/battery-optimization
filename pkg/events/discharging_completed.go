package events

import "time"

// DischargingCompleted is published when battery discharging stops
// Publisher: Device Interface Service
// Subscribers: Bidding Service, Operator Dashboard, Economics Service
type DischargingCompleted struct {
	BatteryID        string    `json:"battery_id"`
	CommandID        string    `json:"command_id"`
	InitialSoC       float64   `json:"initial_soc"`       // 0-100%
	FinalSoC         float64   `json:"final_soc"`         // 0-100%
	EnergyDischarged float64   `json:"energy_discharged"` // MWh
	DurationSeconds  int       `json:"duration_seconds"`  // Actual duration
	Reason           string    `json:"reason"`            // PRICE_BELOW_THRESHOLD | FCAS_DISPATCH | MIN_SOC_REACHED | COMMAND_STOPPED | ERROR
	Timestamp        time.Time `json:"timestamp"`
	EventVersion     string    `json:"event_version"` // "v1"
}
