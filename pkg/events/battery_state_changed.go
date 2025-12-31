package events

import "time"

// BatteryStateChanged is published every second with current battery state.
// Published by: Telemetry Service (M5)
// Subject: battery.state.changed.v1
// Frequency: 1 Hz (1 event per second per battery)
type BatteryStateChanged struct {
	// Battery identification
	BatteryID string `json:"battery_id"`

	// Current state
	SoC              float64                `json:"soc"`                         // State of Charge (0-100%)
	Power            float64                `json:"power"`                       // MW (positive = discharging, negative = charging)
	OperationState   string                 `json:"operation_state"`             // IDLE, CHARGING, DISCHARGING, FCAS
	Temperature      float64                `json:"temperature"`                 // °C
	Voltage          float64                `json:"voltage"`                     // V
	Current          float64                `json:"current"`                     // A
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"` // Vendor-specific data

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
