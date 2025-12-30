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
	Capacity   float64 `json:"capacity"`   // MWh
	MaxPower   float64 `json:"max_power"`  // MW
	RampRate   float64 `json:"ramp_rate"`  // MW/min
	Efficiency float64 `json:"efficiency"` // 0-1

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
// Combines SoC constraints (for bidding logic) with physical constraints (from domain).
type BatteryConstraints struct {
	// State of Charge constraints (for bidding/optimization)
	MinSoC float64 `json:"min_soc"` // Minimum State of Charge (0-1)
	MaxSoC float64 `json:"max_soc"` // Maximum State of Charge (0-1)

	// Physical/warranty constraints (from Asset Management domain)
	WarrantyEOL         float64 `json:"warranty_eol"`          // Warranty End of Life threshold (0-1)
	MaxCycles           int     `json:"max_cycles"`            // Maximum charge/discharge cycles
	OperatingTempMin    float64 `json:"operating_temp_min"`    // Minimum operating temperature (°C)
	OperatingTempMax    float64 `json:"operating_temp_max"`    // Maximum operating temperature (°C)
	GridComplianceLevel string  `json:"grid_compliance_level"` // Grid compliance standard (e.g., AS4777)
}
