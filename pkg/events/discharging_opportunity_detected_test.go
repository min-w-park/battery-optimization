package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDischargingOpportunityDetected_JSONSerialization(t *testing.T) {
	event := events.DischargingOpportunityDetected{
		BatteryID:      "battery-123",
		Price:          150.0,
		SoC:            70.0,
		TargetSoC:      30.0,
		ExpectedProfit: 9500.0,
		StopConditions: events.OpportunityStopConditions{
			PriceThreshold: 100.0,
			Duration:       0,
			FcasDispatch:   false,
		},
		Timestamp:    time.Date(2025, 12, 30, 14, 0, 0, 0, time.UTC),
		EventVersion: "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.DischargingOpportunityDetected
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "battery-123", decoded.BatteryID)
	assert.Equal(t, 150.0, decoded.Price)
	assert.Equal(t, 70.0, decoded.SoC)
	assert.Equal(t, 30.0, decoded.TargetSoC)
	assert.Equal(t, 9500.0, decoded.ExpectedProfit)
	assert.Equal(t, 100.0, decoded.StopConditions.PriceThreshold)
	assert.Equal(t, 0, decoded.StopConditions.Duration)
	assert.Equal(t, false, decoded.StopConditions.FcasDispatch)
	assert.Equal(t, event.Timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestDischargingOpportunityDetected_JSONTags(t *testing.T) {
	event := events.DischargingOpportunityDetected{
		BatteryID:      "test-id",
		Price:          150.0,
		SoC:            70.0,
		TargetSoC:      30.0,
		ExpectedProfit: 9500.0,
		StopConditions: events.OpportunityStopConditions{
			PriceThreshold: 100.0,
			Duration:       0,
			FcasDispatch:   false,
		},
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
	assert.Contains(t, jsonMap, "price")
	assert.Contains(t, jsonMap, "soc")
	assert.Contains(t, jsonMap, "target_soc")
	assert.Contains(t, jsonMap, "expected_profit")
	assert.Contains(t, jsonMap, "stop_conditions")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")

	// Verify nested stop_conditions structure
	stopConditions := jsonMap["stop_conditions"].(map[string]interface{})
	assert.Contains(t, stopConditions, "price_threshold")
	assert.Contains(t, stopConditions, "duration")
	assert.Contains(t, stopConditions, "fcas_dispatch")
}

func TestDischargingOpportunityDetected_StopConditions(t *testing.T) {
	tests := []struct {
		name           string
		stopConditions events.OpportunityStopConditions
	}{
		{
			name: "with price threshold",
			stopConditions: events.OpportunityStopConditions{
				PriceThreshold: 100.0,
				Duration:       0,
				FcasDispatch:   false,
			},
		},
		{
			name: "with duration",
			stopConditions: events.OpportunityStopConditions{
				PriceThreshold: 0,
				Duration:       120,
				FcasDispatch:   false,
			},
		},
		{
			name: "with fcas dispatch",
			stopConditions: events.OpportunityStopConditions{
				PriceThreshold: 100.0,
				Duration:       0,
				FcasDispatch:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := events.DischargingOpportunityDetected{
				BatteryID:      "battery-123",
				Price:          150.0,
				SoC:            70.0,
				TargetSoC:      30.0,
				ExpectedProfit: 9500.0,
				StopConditions: tt.stopConditions,
				Timestamp:      time.Now(),
				EventVersion:   "v1",
			}

			data, err := json.Marshal(event)
			require.NoError(t, err)

			var decoded events.DischargingOpportunityDetected
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.stopConditions, decoded.StopConditions)
		})
	}
}
