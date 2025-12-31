package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChargingOpportunityDetected_JSONSerialization(t *testing.T) {
	event := events.ChargingOpportunityDetected{
		BatteryID:      "battery-123",
		Price:          30.0,
		SoC:            50.0,
		TargetSoC:      80.0,
		ExpectedProfit: 3800.0,
		Timestamp:      time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
		EventVersion:   "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.ChargingOpportunityDetected
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "battery-123", decoded.BatteryID)
	assert.Equal(t, 30.0, decoded.Price)
	assert.Equal(t, 50.0, decoded.SoC)
	assert.Equal(t, 80.0, decoded.TargetSoC)
	assert.Equal(t, 3800.0, decoded.ExpectedProfit)
	assert.Equal(t, event.Timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestChargingOpportunityDetected_JSONTags(t *testing.T) {
	event := events.ChargingOpportunityDetected{
		BatteryID:      "test-id",
		Price:          30.0,
		SoC:            50.0,
		TargetSoC:      80.0,
		ExpectedProfit: 3800.0,
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
	assert.Contains(t, jsonMap, "price")
	assert.Contains(t, jsonMap, "soc")
	assert.Contains(t, jsonMap, "target_soc")
	assert.Contains(t, jsonMap, "expected_profit")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")
}
