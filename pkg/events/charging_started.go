package events

import "time"

// ChargingStarted is published when battery hardware begins charging
// Publisher: Device Interface Service
// Subscribers: Telemetry Service, Bidding Service, Operator Dashboard
type ChargingStarted struct {
	BatteryID    string    `json:"battery_id"`
	CommandID    string    `json:"command_id"`
	ActualPower  float64   `json:"actual_power"` // MW (may differ from commanded due to ramp)
	TargetSoC    float64   `json:"target_soc"`   // 0-100%
	CurrentSoC   float64   `json:"current_soc"`  // 0-100%
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
