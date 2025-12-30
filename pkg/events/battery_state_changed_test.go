package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatteryStateChanged_JSONSerialization(t *testing.T) {
	// NOTE: This is a placeholder for M5 - not published in M4
	event := events.BatteryStateChanged{
		BatteryID:      "battery-456",
		SoC:            0.75,
		Power:          12.5,
		OperationState: "DISCHARGING",
		Temperature:    25.3,
		Voltage:        800.0,
		Current:        15.625,
		Timestamp:      time.Date(2025, 12, 30, 10, 30, 45, 0, time.UTC),
		EventVersion:   "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.BatteryStateChanged
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "battery-456", decoded.BatteryID)
	assert.Equal(t, 0.75, decoded.SoC)
	assert.Equal(t, 12.5, decoded.Power)
	assert.Equal(t, "DISCHARGING", decoded.OperationState)
	assert.Equal(t, 25.3, decoded.Temperature)
	assert.Equal(t, 800.0, decoded.Voltage)
	assert.Equal(t, 15.625, decoded.Current)
	assert.Equal(t, event.Timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestBatteryStateChanged_JSONTags(t *testing.T) {
	event := events.BatteryStateChanged{
		BatteryID:      "test-id",
		SoC:            0.5,
		Power:          -10.0, // Negative = charging
		OperationState: "CHARGING",
		Temperature:    22.0,
		Voltage:        750.0,
		Current:        -13.33,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	// Verify JSON field names (snake_case)
	assert.Contains(t, jsonMap, "battery_id")
	assert.Contains(t, jsonMap, "soc")
	assert.Contains(t, jsonMap, "power")
	assert.Contains(t, jsonMap, "operation_state")
	assert.Contains(t, jsonMap, "temperature")
	assert.Contains(t, jsonMap, "voltage")
	assert.Contains(t, jsonMap, "current")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")
}

func TestBatteryStateChanged_DifferentStatuses(t *testing.T) {
	statuses := []string{"IDLE", "CHARGING", "DISCHARGING", "FCAS"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			event := events.BatteryStateChanged{
				BatteryID:      "test-battery",
				SoC:            0.6,
				Power:          5.0,
				OperationState: status,
				Temperature:    24.0,
				Voltage:        785.0,
				Current:        6.37,
				Timestamp:      time.Now(),
				EventVersion:   "v1",
			}

			data, err := json.Marshal(event)
			require.NoError(t, err)

			var decoded events.BatteryStateChanged
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, status, decoded.OperationState)
		})
	}
}
