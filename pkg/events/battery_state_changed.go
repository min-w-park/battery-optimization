package events

import "time"

// BatteryStateChanged is published every second with current battery state.
// Published by: Telemetry Service (M5)
// Subject: battery.state.changed.v1
// Frequency: 1 Hz (1 event per second per battery)
//
// NOTE: This is a placeholder for M5 - NOT published in M4.
type BatteryStateChanged struct {
	// Battery identification
	BatteryID string `json:"battery_id"`

	// Current state
	SoC         float64 `json:"soc"`         // State of Charge (0-1)
	Power       float64 `json:"power"`       // MW (positive = discharging, negative = charging)
	Status      string  `json:"status"`      // IDLE, CHARGING, DISCHARGING, FCAS
	Temperature float64 `json:"temperature"` // °C

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
