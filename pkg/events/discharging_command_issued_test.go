package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDischargingCommandIssued_JSONSerialization(t *testing.T) {
	event := events.DischargingCommandIssued{
		BatteryID: "battery-123",
		Power:     100.0,
		StopConditions: events.CommandStopConditions{
			PriceThreshold: 100.0,
			TargetSoC:      30.0,
			Duration:       0,
			FcasDispatch:   false,
		},
		Reason:       "High price arbitrage: $150/MWh > $100/MWh threshold",
		IssuedBy:     "BIDDING_SERVICE_AUTO",
		Timestamp:    time.Date(2025, 12, 30, 14, 0, 5, 0, time.UTC),
		EventVersion: "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.DischargingCommandIssued
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "battery-123", decoded.BatteryID)
	assert.Equal(t, 100.0, decoded.Power)
	assert.Equal(t, 100.0, decoded.StopConditions.PriceThreshold)
	assert.Equal(t, 30.0, decoded.StopConditions.TargetSoC)
	assert.Equal(t, 0, decoded.StopConditions.Duration)
	assert.Equal(t, false, decoded.StopConditions.FcasDispatch)
	assert.Equal(t, "High price arbitrage: $150/MWh > $100/MWh threshold", decoded.Reason)
	assert.Equal(t, "BIDDING_SERVICE_AUTO", decoded.IssuedBy)
	assert.Equal(t, event.Timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestDischargingCommandIssued_JSONTags(t *testing.T) {
	event := events.DischargingCommandIssued{
		BatteryID: "test-id",
		Power:     100.0,
		StopConditions: events.CommandStopConditions{
			PriceThreshold: 100.0,
			TargetSoC:      30.0,
			Duration:       0,
			FcasDispatch:   false,
		},
		Reason:       "test",
		IssuedBy:     "BIDDING_SERVICE_AUTO",
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	// Verify JSON field names (snake_case)
	assert.Contains(t, jsonMap, "battery_id")
	assert.Contains(t, jsonMap, "power")
	assert.Contains(t, jsonMap, "stop_conditions")
	assert.Contains(t, jsonMap, "reason")
	assert.Contains(t, jsonMap, "issued_by")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")

	// Verify nested stop_conditions structure
	stopConditions := jsonMap["stop_conditions"].(map[string]interface{})
	assert.Contains(t, stopConditions, "price_threshold")
	assert.Contains(t, stopConditions, "target_soc")
	assert.Contains(t, stopConditions, "duration")
	assert.Contains(t, stopConditions, "fcas_dispatch")
}

func TestDischargingCommandIssued_StopConditions(t *testing.T) {
	tests := []struct {
		name           string
		stopConditions events.CommandStopConditions
	}{
		{
			name: "all conditions",
			stopConditions: events.CommandStopConditions{
				PriceThreshold: 100.0,
				TargetSoC:      30.0,
				Duration:       120,
				FcasDispatch:   true,
			},
		},
		{
			name: "price only",
			stopConditions: events.CommandStopConditions{
				PriceThreshold: 100.0,
				TargetSoC:      0,
				Duration:       0,
				FcasDispatch:   false,
			},
		},
		{
			name: "soc only",
			stopConditions: events.CommandStopConditions{
				PriceThreshold: 0,
				TargetSoC:      30.0,
				Duration:       0,
				FcasDispatch:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := events.DischargingCommandIssued{
				BatteryID:      "battery-123",
				Power:          100.0,
				StopConditions: tt.stopConditions,
				Reason:         "test",
				IssuedBy:       "BIDDING_SERVICE_AUTO",
				Timestamp:      time.Now(),
				EventVersion:   "v1",
			}

			data, err := json.Marshal(event)
			require.NoError(t, err)

			var decoded events.DischargingCommandIssued
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.stopConditions, decoded.StopConditions)
		})
	}
}
