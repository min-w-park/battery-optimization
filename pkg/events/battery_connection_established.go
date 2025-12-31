package events

import "time"

// BatteryConnectionEstablished is published when Device Interface connects to battery hardware
// Publisher: Device Interface Service
// Subscribers: Asset Management Service, Telemetry Service, Operator Dashboard
type BatteryConnectionEstablished struct {
	BatteryID        string                 `json:"battery_id"`
	AdapterType      string                 `json:"adapter_type"` // TeslaLike, BYDLike, etc.
	FirmwareVersion  string                 `json:"firmware_version"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	EventVersion     string                 `json:"event_version"` // "v1"
}
