package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBattery_ValidInput(t *testing.T) {
	// Given: Valid battery parameters (Hornsdale-like battery)
	capacity := 200.0 // 200 MWh
	maxPower := 100.0 // 100 MW
	rampRate := 50.0  // 50 MW/min
	efficiency := 0.85
	location := "SA"
	manufacturer := "Tesla"
	constraints := Constraints{
		WarrantyEOL:         0.8,
		MaxCycles:           10000,
		OperatingTempMin:    -10.0,
		OperatingTempMax:    50.0,
		GridComplianceLevel: "AS4777",
	}

	// When: Creating a new battery
	battery, err := NewBattery(capacity, maxPower, rampRate, efficiency, location, manufacturer, constraints)

	// Then: Battery is created successfully
	require.NoError(t, err)
	require.NotNil(t, battery)
	assert.NotEmpty(t, battery.ID, "ID should be generated")
	assert.Equal(t, capacity, battery.Capacity)
	assert.Equal(t, maxPower, battery.MaxPower)
	assert.Equal(t, rampRate, battery.RampRate)
	assert.Equal(t, efficiency, battery.Efficiency)
	assert.Equal(t, location, battery.Location)
	assert.Equal(t, manufacturer, battery.Manufacturer)
	assert.Equal(t, StatusRegistered, battery.Status)
	assert.NotZero(t, battery.CreatedAt)
	assert.NotZero(t, battery.UpdatedAt)
}

func TestNewBattery_InvalidCapacity(t *testing.T) {
	tests := []struct {
		name     string
		capacity float64
	}{
		{"zero capacity", 0},
		{"negative capacity", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: Invalid capacity
			constraints := Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			}

			// When: Creating battery with invalid capacity
			battery, err := NewBattery(tt.capacity, 100.0, 50.0, 0.85, "NSW", "Tesla", constraints)

			// Then: Should return ErrInvalidCapacity
			assert.ErrorIs(t, err, ErrInvalidCapacity)
			assert.Nil(t, battery)
		})
	}
}

func TestNewBattery_InvalidMaxPower(t *testing.T) {
	tests := []struct {
		name     string
		capacity float64
		maxPower float64
		wantErr  error
	}{
		{"zero max power", 200.0, 0, ErrInvalidMaxPower},
		{"negative max power", 200.0, -50, ErrInvalidMaxPower},
		{"max power exceeds capacity", 200.0, 250.0, ErrMaxPowerExceedsCapacity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: Invalid max power
			constraints := Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			}

			// When: Creating battery with invalid max power
			battery, err := NewBattery(tt.capacity, tt.maxPower, 50.0, 0.85, "NSW", "Tesla", constraints)

			// Then: Should return appropriate error
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, battery)
		})
	}
}

func TestNewBattery_InvalidRampRate(t *testing.T) {
	tests := []struct {
		name     string
		rampRate float64
		maxPower float64
		wantErr  error
	}{
		{"zero ramp rate", 0, 100.0, ErrInvalidRampRate},
		{"negative ramp rate", -10, 100.0, ErrInvalidRampRate},
		{"ramp rate exceeds max power", 150.0, 100.0, ErrRampRateExceedsMaxPower},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: Invalid ramp rate
			constraints := Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			}

			// When: Creating battery with invalid ramp rate
			battery, err := NewBattery(200.0, tt.maxPower, tt.rampRate, 0.85, "NSW", "Tesla", constraints)

			// Then: Should return appropriate error
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, battery)
		})
	}
}

func TestNewBattery_InvalidEfficiency(t *testing.T) {
	tests := []struct {
		name       string
		efficiency float64
	}{
		{"negative efficiency", -0.1},
		{"efficiency greater than 1", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: Invalid efficiency
			constraints := Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			}

			// When: Creating battery with invalid efficiency
			battery, err := NewBattery(200.0, 100.0, 50.0, tt.efficiency, "NSW", "Tesla", constraints)

			// Then: Should return ErrInvalidEfficiency
			assert.ErrorIs(t, err, ErrInvalidEfficiency)
			assert.Nil(t, battery)
		})
	}
}

func TestNewBattery_InvalidLocation(t *testing.T) {
	// Given: Invalid location (not a valid NEM region)
	constraints := Constraints{
		WarrantyEOL:      0.8,
		MaxCycles:        10000,
		OperatingTempMin: -10.0,
		OperatingTempMax: 50.0,
	}

	// When: Creating battery with invalid location
	battery, err := NewBattery(200.0, 100.0, 50.0, 0.85, "INVALID", "Tesla", constraints)

	// Then: Should return ErrInvalidLocation
	assert.ErrorIs(t, err, ErrInvalidLocation)
	assert.Nil(t, battery)
}

func TestNewBattery_AllValidLocations(t *testing.T) {
	// Test all valid NEM regions
	validLocations := []string{"NSW", "VIC", "QLD", "SA", "TAS"}

	for _, location := range validLocations {
		t.Run(location, func(t *testing.T) {
			// Given: Valid NEM region
			constraints := Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			}

			// When: Creating battery with valid location
			battery, err := NewBattery(200.0, 100.0, 50.0, 0.85, location, "Tesla", constraints)

			// Then: Should succeed
			require.NoError(t, err)
			assert.Equal(t, location, battery.Location)
		})
	}
}

func TestConstraints_Validate(t *testing.T) {
	tests := []struct {
		name        string
		constraints Constraints
		wantErr     error
	}{
		{
			name: "valid constraints",
			constraints: Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			},
			wantErr: nil,
		},
		{
			name: "warranty EOL negative",
			constraints: Constraints{
				WarrantyEOL:      -0.1,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			},
			wantErr: ErrInvalidWarrantyEOL,
		},
		{
			name: "warranty EOL greater than 1",
			constraints: Constraints{
				WarrantyEOL:      1.5,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			},
			wantErr: ErrInvalidWarrantyEOL,
		},
		{
			name: "max cycles zero",
			constraints: Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        0,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			},
			wantErr: ErrInvalidMaxCycles,
		},
		{
			name: "max cycles negative",
			constraints: Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        -100,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			},
			wantErr: ErrInvalidMaxCycles,
		},
		{
			name: "temp min greater than max",
			constraints: Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: 60.0,
				OperatingTempMax: 50.0,
			},
			wantErr: ErrInvalidTemperatureRange,
		},
		{
			name: "temp min equal to max",
			constraints: Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: 50.0,
				OperatingTempMax: 50.0,
			},
			wantErr: ErrInvalidTemperatureRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: Validating constraints
			err := tt.constraints.Validate()

			// Then: Should return expected error
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBattery_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		capacity float64
		maxPower float64
		rampRate float64
		wantErr  bool
	}{
		{
			name:     "maxPower equals capacity (boundary)",
			capacity: 200.0,
			maxPower: 200.0,
			rampRate: 50.0,
			wantErr:  false,
		},
		{
			name:     "rampRate equals maxPower (boundary)",
			capacity: 200.0,
			maxPower: 100.0,
			rampRate: 100.0,
			wantErr:  false,
		},
		{
			name:     "efficiency is 0 (boundary)",
			capacity: 200.0,
			maxPower: 100.0,
			rampRate: 50.0,
			wantErr:  false, // Will test with efficiency 0 in the actual call
		},
		{
			name:     "efficiency is 1 (boundary)",
			capacity: 200.0,
			maxPower: 100.0,
			rampRate: 50.0,
			wantErr:  false, // Will test with efficiency 1 in the actual call
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints := Constraints{
				WarrantyEOL:      0.8,
				MaxCycles:        10000,
				OperatingTempMin: -10.0,
				OperatingTempMax: 50.0,
			}

			efficiency := 0.85 // default
			if tt.name == "efficiency is 0 (boundary)" {
				efficiency = 0.0
			} else if tt.name == "efficiency is 1 (boundary)" {
				efficiency = 1.0
			}

			battery, err := NewBattery(tt.capacity, tt.maxPower, tt.rampRate, efficiency, "NSW", "Tesla", constraints)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, battery)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, battery)
			}
		})
	}
}
