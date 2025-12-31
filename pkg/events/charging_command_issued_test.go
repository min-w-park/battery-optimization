package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChargingCommandIssued_JSONSerialization(t *testing.T) {
	event := events.ChargingCommandIssued{
		BatteryID:    "battery-123",
		TargetSoC:    80.0,
		MaxPower:     100.0,
		Reason:       "Low price arbitrage: $30/MWh < $50/MWh threshold",
		IssuedBy:     "BIDDING_SERVICE_AUTO",
		Timestamp:    time.Date(2025, 12, 30, 10, 0, 5, 0, time.UTC),
		EventVersion: "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.ChargingCommandIssued
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "battery-123", decoded.BatteryID)
	assert.Equal(t, 80.0, decoded.TargetSoC)
	assert.Equal(t, 100.0, decoded.MaxPower)
	assert.Equal(t, "Low price arbitrage: $30/MWh < $50/MWh threshold", decoded.Reason)
	assert.Equal(t, "BIDDING_SERVICE_AUTO", decoded.IssuedBy)
	assert.Equal(t, event.Timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestChargingCommandIssued_JSONTags(t *testing.T) {
	event := events.ChargingCommandIssued{
		BatteryID:    "test-id",
		TargetSoC:    80.0,
		MaxPower:     100.0,
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
	assert.Contains(t, jsonMap, "target_soc")
	assert.Contains(t, jsonMap, "max_power")
	assert.Contains(t, jsonMap, "reason")
	assert.Contains(t, jsonMap, "issued_by")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")
}

func TestChargingCommandIssued_DifferentIssuers(t *testing.T) {
	issuers := []string{"BIDDING_SERVICE_AUTO", "operator-123", "MANUAL_OVERRIDE"}

	for _, issuer := range issuers {
		t.Run(issuer, func(t *testing.T) {
			event := events.ChargingCommandIssued{
				BatteryID:    "battery-123",
				TargetSoC:    80.0,
				MaxPower:     100.0,
				Reason:       "test",
				IssuedBy:     issuer,
				Timestamp:    time.Now(),
				EventVersion: "v1",
			}

			data, err := json.Marshal(event)
			require.NoError(t, err)

			var decoded events.ChargingCommandIssued
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, issuer, decoded.IssuedBy)
		})
	}
}
