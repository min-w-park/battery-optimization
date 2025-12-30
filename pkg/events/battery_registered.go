package events

import "time"

// BatteryRegistered is published when a new battery is added to the system.
// Published by: Asset Management Service
// Subject: battery.registered.v1
// Frequency: On-demand (when operator creates battery)
type BatteryRegistered struct {
	// Battery identification
	BatteryID string `json:"battery_id"`

	// Battery specifications
	Capacity     float64 `json:"capacity"`      // MWh
	MaxPower     float64 `json:"max_power"`     // MW
	RampRate     float64 `json:"ramp_rate"`     // MW/min
	Efficiency   float64 `json:"efficiency"`    // 0-1

	// Additional metadata
	Location     string `json:"location"`
	Manufacturer string `json:"manufacturer"`

	// Operational constraints
	Constraints BatteryConstraints `json:"constraints"`

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}

// BatteryConstraints represents operational limits for a battery.
type BatteryConstraints struct {
	MinSoC      float64 `json:"min_soc"`       // Minimum State of Charge (0-1)
	MaxSoC      float64 `json:"max_soc"`       // Maximum State of Charge (0-1)
	WarrantyEOL float64 `json:"warranty_eol"`  // Warranty End of Life threshold (0-1)
	MaxCycles   int     `json:"max_cycles"`    // Maximum charge/discharge cycles
}
