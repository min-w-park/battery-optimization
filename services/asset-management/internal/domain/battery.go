package domain

import (
	"time"

	"github.com/google/uuid"
)

// BatteryStatus represents the operational status of a battery
type BatteryStatus string

const (
	StatusRegistered     BatteryStatus = "REGISTERED"
	StatusTesting        BatteryStatus = "TESTING"
	StatusOperational    BatteryStatus = "OPERATIONAL"
	StatusMaintenance    BatteryStatus = "MAINTENANCE"
	StatusDecommissioned BatteryStatus = "DECOMMISSIONED"
)

// Constraints represents the operational constraints and compliance requirements for a battery
type Constraints struct {
	WarrantyEOL         float64 `json:"warranty_eol"`          // Warranty end-of-life threshold (0-1)
	MaxCycles           int     `json:"max_cycles"`            // Maximum charge/discharge cycles
	OperatingTempMin    float64 `json:"operating_temp_min"`    // Minimum operating temperature (°C)
	OperatingTempMax    float64 `json:"operating_temp_max"`    // Maximum operating temperature (°C)
	GridComplianceLevel string  `json:"grid_compliance_level"` // Grid compliance standard (e.g., AS4777)
}

// Validate checks if the constraints are valid
func (c *Constraints) Validate() error {
	// Warranty EOL must be between 0 and 1
	if c.WarrantyEOL < 0 || c.WarrantyEOL > 1 {
		return ErrInvalidWarrantyEOL
	}

	// Max cycles must be positive
	if c.MaxCycles <= 0 {
		return ErrInvalidMaxCycles
	}

	// Temperature min must be less than max
	if c.OperatingTempMin >= c.OperatingTempMax {
		return ErrInvalidTemperatureRange
	}

	return nil
}

// Battery represents the aggregate root for battery asset management
type Battery struct {
	ID           string        // Unique identifier (UUID)
	Capacity     float64       // Energy capacity in MWh
	MaxPower     float64       // Maximum power in MW
	RampRate     float64       // Ramp rate in MW/min
	Efficiency   float64       // Round-trip efficiency (0-1)
	Location     string        // NEM region (NSW, VIC, QLD, SA, TAS)
	Manufacturer string        // Battery manufacturer
	Constraints  Constraints   // Operational constraints
	Status       BatteryStatus // Current operational status
	CreatedAt    time.Time     // Creation timestamp
	UpdatedAt    time.Time     // Last update timestamp
}

// NewBattery creates a new Battery aggregate with validation
func NewBattery(
	capacity float64,
	maxPower float64,
	rampRate float64,
	efficiency float64,
	location string,
	manufacturer string,
	constraints Constraints,
) (*Battery, error) {
	battery := &Battery{
		ID:           uuid.New().String(),
		Capacity:     capacity,
		MaxPower:     maxPower,
		RampRate:     rampRate,
		Efficiency:   efficiency,
		Location:     location,
		Manufacturer: manufacturer,
		Constraints:  constraints,
		Status:       StatusRegistered,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Validate the battery before returning
	if err := battery.Validate(); err != nil {
		return nil, err
	}

	return battery, nil
}

// Validate checks if the battery satisfies all business rules
func (b *Battery) Validate() error {
	// Rule 1: Capacity must be positive
	if b.Capacity <= 0 {
		return ErrInvalidCapacity
	}

	// Rule 2: MaxPower must be positive
	if b.MaxPower <= 0 {
		return ErrInvalidMaxPower
	}

	// Rule 3: MaxPower cannot exceed Capacity
	// (Can't discharge more than capacity in 1 hour)
	if b.MaxPower > b.Capacity {
		return ErrMaxPowerExceedsCapacity
	}

	// Rule 4: RampRate must be positive
	if b.RampRate <= 0 {
		return ErrInvalidRampRate
	}

	// Rule 5: RampRate cannot exceed MaxPower
	if b.RampRate > b.MaxPower {
		return ErrRampRateExceedsMaxPower
	}

	// Rule 6: Efficiency must be between 0 and 1
	if b.Efficiency < 0 || b.Efficiency > 1 {
		return ErrInvalidEfficiency
	}

	// Rule 7: Location must be a valid NEM region
	if !isValidLocation(b.Location) {
		return ErrInvalidLocation
	}

	// Rule 8-10: Validate constraints
	if err := b.Constraints.Validate(); err != nil {
		return err
	}

	return nil
}

// isValidLocation checks if the location is a valid Australian NEM region
func isValidLocation(location string) bool {
	validLocations := map[string]bool{
		"NSW": true, // New South Wales
		"VIC": true, // Victoria
		"QLD": true, // Queensland
		"SA":  true, // South Australia
		"TAS": true, // Tasmania
	}
	return validLocations[location]
}
