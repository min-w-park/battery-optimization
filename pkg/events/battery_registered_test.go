package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatteryRegistered_JSONSerialization(t *testing.T) {
	// Create event with all fields
	event := events.BatteryRegistered{
		BatteryID:    "test-battery-123",
		Capacity:     50.0,
		MaxPower:     25.0,
		RampRate:     5.0,
		Efficiency:   0.92,
		Location:     "Sydney, NSW",
		Manufacturer: "Tesla Megapack",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.2,
			MaxSoC:      0.9,
			WarrantyEOL: 0.7,
			MaxCycles:   10000,
		},
		Timestamp:    time.Date(2025, 12, 30, 10, 30, 0, 0, time.UTC),
		EventVersion: "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.BatteryRegistered
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "test-battery-123", decoded.BatteryID)
	assert.Equal(t, 50.0, decoded.Capacity)
	assert.Equal(t, 25.0, decoded.MaxPower)
	assert.Equal(t, 5.0, decoded.RampRate)
	assert.Equal(t, 0.92, decoded.Efficiency)
	assert.Equal(t, "Sydney, NSW", decoded.Location)
	assert.Equal(t, "Tesla Megapack", decoded.Manufacturer)
	assert.Equal(t, 0.2, decoded.Constraints.MinSoC)
	assert.Equal(t, 0.9, decoded.Constraints.MaxSoC)
	assert.Equal(t, 0.7, decoded.Constraints.WarrantyEOL)
	assert.Equal(t, 10000, decoded.Constraints.MaxCycles)
	assert.Equal(t, event.Timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestBatteryRegistered_JSONTags(t *testing.T) {
	event := events.BatteryRegistered{
		BatteryID:    "test-id",
		Capacity:     100.0,
		MaxPower:     50.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "Melbourne, VIC",
		Manufacturer: "BYD",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.1,
			MaxSoC:      0.95,
			WarrantyEOL: 0.8,
			MaxCycles:   15000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)

	// Verify JSON field names (snake_case)
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	// Check top-level fields
	assert.Contains(t, jsonMap, "battery_id")
	assert.Contains(t, jsonMap, "capacity")
	assert.Contains(t, jsonMap, "max_power")
	assert.Contains(t, jsonMap, "ramp_rate")
	assert.Contains(t, jsonMap, "efficiency")
	assert.Contains(t, jsonMap, "location")
	assert.Contains(t, jsonMap, "manufacturer")
	assert.Contains(t, jsonMap, "constraints")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")

	// Check constraints fields
	constraints := jsonMap["constraints"].(map[string]interface{})
	assert.Contains(t, constraints, "min_soc")
	assert.Contains(t, constraints, "max_soc")
	assert.Contains(t, constraints, "warranty_eol")
	assert.Contains(t, constraints, "max_cycles")
}

func TestBatteryRegistered_EmptyEvent(t *testing.T) {
	// Test zero values can be marshaled
	event := events.BatteryRegistered{}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded events.BatteryRegistered
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "", decoded.BatteryID)
	assert.Equal(t, 0.0, decoded.Capacity)
}
