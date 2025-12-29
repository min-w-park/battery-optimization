package domain

import "errors"

// Domain validation errors for Battery aggregate
var (
	// ErrInvalidCapacity indicates the battery capacity is invalid (must be > 0)
	ErrInvalidCapacity = errors.New("capacity must be greater than 0")

	// ErrInvalidMaxPower indicates the max power is invalid (must be > 0)
	ErrInvalidMaxPower = errors.New("max power must be greater than 0")

	// ErrMaxPowerExceedsCapacity indicates max power exceeds capacity (MaxPower <= Capacity constraint)
	ErrMaxPowerExceedsCapacity = errors.New("max power cannot exceed capacity")

	// ErrInvalidRampRate indicates the ramp rate is invalid (must be > 0)
	ErrInvalidRampRate = errors.New("ramp rate must be greater than 0")

	// ErrRampRateExceedsMaxPower indicates ramp rate exceeds max power (RampRate <= MaxPower constraint)
	ErrRampRateExceedsMaxPower = errors.New("ramp rate cannot exceed max power")

	// ErrInvalidEfficiency indicates efficiency is out of valid range (0 <= Efficiency <= 1)
	ErrInvalidEfficiency = errors.New("efficiency must be between 0 and 1")

	// ErrInvalidLocation indicates location is not a valid NEM region
	ErrInvalidLocation = errors.New("location must be a valid NEM region (NSW, VIC, QLD, SA, TAS)")

	// ErrInvalidWarrantyEOL indicates warranty EOL is out of valid range (0 <= WarrantyEOL <= 1)
	ErrInvalidWarrantyEOL = errors.New("warranty EOL must be between 0 and 1")

	// ErrInvalidMaxCycles indicates max cycles is invalid (must be > 0)
	ErrInvalidMaxCycles = errors.New("max cycles must be greater than 0")

	// ErrInvalidTemperatureRange indicates temperature range is invalid (min < max)
	ErrInvalidTemperatureRange = errors.New("operating temperature min must be less than max")
)

// Repository errors
var (
	// ErrNotFound indicates the requested battery was not found
	ErrNotFound = errors.New("battery not found")

	// ErrDuplicateID indicates a battery with this ID already exists
	ErrDuplicateID = errors.New("battery with this ID already exists")
)
