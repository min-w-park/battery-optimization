package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBatteryState_ValidInput(t *testing.T) {
	// Given: Valid battery state parameters
	batteryID := "battery-123"
	soc := 75.5
	power := 25.0
	temperature := 28.5
	voltage := 800.0
	current := 31.25
	operationState := OperationStateDischarging
	customAttributes := map[string]interface{}{
		"vendor": "Tesla",
		"model":  "Megapack",
	}
	timestamp := time.Now()

	// When: Creating a new battery state
	state, err := NewBatteryState(
		batteryID,
		soc,
		power,
		temperature,
		voltage,
		current,
		operationState,
		customAttributes,
		timestamp,
	)

	// Then: State is created successfully
	require.NoError(t, err)
	assert.NotEmpty(t, state.ID)
	assert.Equal(t, batteryID, state.BatteryID)
	assert.Equal(t, soc, state.SoC)
	assert.Equal(t, power, state.Power)
	assert.Equal(t, temperature, state.Temperature)
	assert.Equal(t, voltage, state.Voltage)
	assert.Equal(t, current, state.Current)
	assert.Equal(t, operationState, state.OperationState)
	assert.Equal(t, customAttributes, state.CustomAttributes)
	assert.Equal(t, timestamp, state.Timestamp)
	assert.False(t, state.CreatedAt.IsZero())
}

func TestNewBatteryState_InvalidSoC(t *testing.T) {
	tests := []struct {
		name string
		soc  float64
	}{
		{"negative SoC", -10.0},
		{"zero SoC is valid", 0.0}, // Actually valid
		{"SoC over 100", 150.0},
		{"SoC slightly over 100", 100.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := NewBatteryState(
				"battery-123",
				tt.soc,
				0.0,
				25.0,
				800.0,
				0.0,
				OperationStateIdle,
				nil,
				time.Now(),
			)

			if tt.name == "zero SoC is valid" {
				assert.NoError(t, err)
				assert.NotNil(t, state)
			} else {
				assert.ErrorIs(t, err, ErrInvalidSoC)
				assert.Nil(t, state)
			}
		})
	}
}

func TestNewBatteryState_InvalidTemperature(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
		wantErr     bool
	}{
		{"temperature too low", -25.0, true},
		{"minimum valid temperature", -20.0, false},
		{"normal temperature", 25.0, false},
		{"maximum valid temperature", 60.0, false},
		{"temperature too high", 65.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := NewBatteryState(
				"battery-123",
				50.0,
				0.0,
				tt.temperature,
				800.0,
				0.0,
				OperationStateIdle,
				nil,
				time.Now(),
			)

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidTemperature)
				assert.Nil(t, state)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, state)
			}
		})
	}
}

func TestNewBatteryState_InvalidOperationState(t *testing.T) {
	tests := []struct {
		name           string
		operationState string
		power          float64
		wantErr        bool
	}{
		{"IDLE is valid", OperationStateIdle, 0.0, false},
		{"CHARGING is valid", OperationStateCharging, -15.0, false},
		{"DISCHARGING is valid", OperationStateDischarging, 25.0, false},
		{"FCAS is valid", OperationStateFCAS, 10.0, false},
		{"invalid state", "INVALID", 0.0, true},
		{"empty state", "", 0.0, true},
		{"lowercase state", "idle", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := NewBatteryState(
				"battery-123",
				50.0,
				tt.power,
				25.0,
				800.0,
				0.0,
				tt.operationState,
				nil,
				time.Now(),
			)

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidOperationState)
				assert.Nil(t, state)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, state)
			}
		})
	}
}

func TestNewBatteryState_PowerConsistency(t *testing.T) {
	tests := []struct {
		name           string
		power          float64
		operationState string
		wantErr        bool
	}{
		// IDLE must have zero power
		{"IDLE with zero power", 0.0, OperationStateIdle, false},
		{"IDLE with positive power", 10.0, OperationStateIdle, true},
		{"IDLE with negative power", -10.0, OperationStateIdle, true},

		// CHARGING must have negative power
		{"CHARGING with negative power", -15.0, OperationStateCharging, false},
		{"CHARGING with zero power", 0.0, OperationStateCharging, true},
		{"CHARGING with positive power", 15.0, OperationStateCharging, true},

		// DISCHARGING must have positive power
		{"DISCHARGING with positive power", 25.0, OperationStateDischarging, false},
		{"DISCHARGING with zero power", 0.0, OperationStateDischarging, true},
		{"DISCHARGING with negative power", -25.0, OperationStateDischarging, true},

		// FCAS can have any power (responding to frequency signals)
		{"FCAS with positive power", 10.0, OperationStateFCAS, false},
		{"FCAS with negative power", -10.0, OperationStateFCAS, false},
		{"FCAS with zero power", 0.0, OperationStateFCAS, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := NewBatteryState(
				"battery-123",
				50.0,
				tt.power,
				25.0,
				800.0,
				0.0,
				tt.operationState,
				nil,
				time.Now(),
			)

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInconsistentPowerState)
				assert.Nil(t, state)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, state)
			}
		})
	}
}

func TestNewBatteryState_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		batteryID string
		timestamp time.Time
		wantErr   bool
	}{
		{"empty battery ID", "", time.Now(), true},
		{"zero timestamp", "battery-123", time.Time{}, true},
		{"valid required fields", "battery-123", time.Now(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := NewBatteryState(
				tt.batteryID,
				50.0,
				0.0,
				25.0,
				800.0,
				0.0,
				OperationStateIdle,
				nil,
				tt.timestamp,
			)

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrMissingRequiredField)
				assert.Nil(t, state)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, state)
			}
		})
	}
}

func TestNewBatteryState_InvalidTimestamp(t *testing.T) {
	// Timestamp in the future (allowing 1s clock skew)
	futureTime := time.Now().Add(5 * time.Second)

	state, err := NewBatteryState(
		"battery-123",
		50.0,
		0.0,
		25.0,
		800.0,
		0.0,
		OperationStateIdle,
		nil,
		futureTime,
	)

	assert.ErrorIs(t, err, ErrInvalidTimestamp)
	assert.Nil(t, state)
}

func TestNewBatteryState_TimestampWithinClockSkew(t *testing.T) {
	// Timestamp slightly in future but within 1s clock skew tolerance
	almostFutureTime := time.Now().Add(500 * time.Millisecond)

	state, err := NewBatteryState(
		"battery-123",
		50.0,
		0.0,
		25.0,
		800.0,
		0.0,
		OperationStateIdle,
		nil,
		almostFutureTime,
	)

	// Should be valid (within clock skew tolerance)
	assert.NoError(t, err)
	assert.NotNil(t, state)
}

func TestBatteryState_Validate(t *testing.T) {
	// Create a valid state first
	state, err := NewBatteryState(
		"battery-123",
		75.0,
		25.0,
		28.5,
		800.0,
		31.25,
		OperationStateDischarging,
		nil,
		time.Now(),
	)
	require.NoError(t, err)

	// Validate should pass for valid state
	err = state.Validate()
	assert.NoError(t, err)

	// Modify to invalid SoC
	state.SoC = 150.0
	err = state.Validate()
	assert.ErrorIs(t, err, ErrInvalidSoC)
}

func TestBatteryState_HelperMethods(t *testing.T) {
	tests := []struct {
		name           string
		operationState string
		power          float64
		expectedIdle   bool
		expectedCharge bool
		expectedDisch  bool
	}{
		{
			name:           "IDLE state",
			operationState: OperationStateIdle,
			power:          0.0,
			expectedIdle:   true,
			expectedCharge: false,
			expectedDisch:  false,
		},
		{
			name:           "CHARGING state",
			operationState: OperationStateCharging,
			power:          -15.0,
			expectedIdle:   false,
			expectedCharge: true,
			expectedDisch:  false,
		},
		{
			name:           "DISCHARGING state",
			operationState: OperationStateDischarging,
			power:          25.0,
			expectedIdle:   false,
			expectedCharge: false,
			expectedDisch:  true,
		},
		{
			name:           "FCAS state",
			operationState: OperationStateFCAS,
			power:          10.0,
			expectedIdle:   false,
			expectedCharge: false,
			expectedDisch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := NewBatteryState(
				"battery-123",
				50.0,
				tt.power,
				25.0,
				800.0,
				0.0,
				tt.operationState,
				nil,
				time.Now(),
			)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedIdle, state.IsIdle())
			assert.Equal(t, tt.expectedCharge, state.IsCharging())
			assert.Equal(t, tt.expectedDisch, state.IsDischarging())
		})
	}
}
