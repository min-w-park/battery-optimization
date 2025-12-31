package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Event Assertion Helpers

// AssertBatteryRegistered validates a battery.registered.v1 event
func AssertBatteryRegistered(t *testing.T, msg *nats.Msg, expectedBatteryID string, expectedCapacity, expectedMaxPower float64) {
	require.NotNil(t, msg, "Expected battery.registered.v1 event, got nil")
	assert.Equal(t, "battery.registered.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"])
	assert.Equal(t, expectedCapacity, event["capacity"])
	assert.Equal(t, expectedMaxPower, event["max_power"])
	assert.NotEmpty(t, event["timestamp"])
	assert.Equal(t, "v1", event["event_version"])
}

// AssertMarketPriceUpdated validates a market.price.updated.v1 event
func AssertMarketPriceUpdated(t *testing.T, msg *nats.Msg, expectedPrice float64) {
	require.NotNil(t, msg, "Expected market.price.updated.v1 event, got nil")
	assert.Equal(t, "market.price.updated.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedPrice, event["value"])
	assert.NotEmpty(t, event["timestamp"])
	assert.Equal(t, "v1", event["event_version"])
}

// AssertChargingOpportunity validates a charging.opportunity.detected.v1 event
func AssertChargingOpportunity(t *testing.T, msg *nats.Msg, expectedBatteryID string, expectedPrice, expectedSoC float64) {
	require.NotNil(t, msg, "Expected charging.opportunity.detected.v1 event, got nil")
	assert.Equal(t, "charging.opportunity.detected.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.Equal(t, expectedPrice, event["price"], "Price mismatch")
	assert.Equal(t, expectedSoC, event["soc"], "SoC mismatch")
	assert.Equal(t, 80.0, event["target_soc"], "Target SoC should be 80%")
	assert.NotEmpty(t, event["expected_profit"], "Expected profit should be calculated")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertChargingCommand validates a charging.command.issued.v1 event
func AssertChargingCommand(t *testing.T, msg *nats.Msg, expectedBatteryID string, expectedTargetSoC, expectedMaxPower float64) {
	require.NotNil(t, msg, "Expected charging.command.issued.v1 event, got nil")
	assert.Equal(t, "charging.command.issued.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.Equal(t, expectedTargetSoC, event["target_soc"], "Target SoC mismatch")
	assert.Equal(t, expectedMaxPower, event["max_power"], "Max power mismatch")
	assert.NotEmpty(t, event["reason"], "Reason missing")
	assert.Contains(t, event["reason"], "Low price", "Reason should mention low price")
	assert.NotEmpty(t, event["issued_by"], "Issued by missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertDischargingOpportunity validates a discharging.opportunity.detected.v1 event
func AssertDischargingOpportunity(t *testing.T, msg *nats.Msg, expectedBatteryID string, expectedPrice, expectedSoC float64) {
	require.NotNil(t, msg, "Expected discharging.opportunity.detected.v1 event, got nil")
	assert.Equal(t, "discharging.opportunity.detected.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.Equal(t, expectedPrice, event["price"], "Price mismatch")
	assert.Equal(t, expectedSoC, event["soc"], "SoC mismatch")
	assert.Equal(t, 30.0, event["target_soc"], "Target SoC should be 30%")
	assert.NotEmpty(t, event["expected_profit"], "Expected profit should be calculated")
	assert.NotEmpty(t, event["stop_conditions"], "Stop conditions missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertDischargingCommand validates a discharging.command.issued.v1 event
func AssertDischargingCommand(t *testing.T, msg *nats.Msg, expectedBatteryID string, expectedPower float64) {
	require.NotNil(t, msg, "Expected discharging.command.issued.v1 event, got nil")
	assert.Equal(t, "discharging.command.issued.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.Equal(t, expectedPower, event["power"], "Power mismatch")
	assert.NotEmpty(t, event["reason"], "Reason missing")
	assert.Contains(t, event["reason"], "High price", "Reason should mention high price")
	assert.NotEmpty(t, event["issued_by"], "Issued by missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])

	// Validate stop conditions exist
	stopConditions, ok := event["stop_conditions"].(map[string]interface{})
	require.True(t, ok, "Stop conditions should be an object")
	assert.NotEmpty(t, stopConditions, "Stop conditions should not be empty")
}

// AssertBatteryStateChanged validates a battery.state.changed.v1 event
func AssertBatteryStateChanged(t *testing.T, msg *nats.Msg, expectedBatteryID string, expectedSoC float64, expectedState string) {
	require.NotNil(t, msg, "Expected battery.state.changed.v1 event, got nil")
	assert.Equal(t, "battery.state.changed.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.Equal(t, expectedSoC, event["soc"], "SoC mismatch")
	assert.Equal(t, expectedState, event["operation_state"], "Operation state mismatch")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertChargingStarted validates a charging.started.v1 event
func AssertChargingStarted(t *testing.T, msg *nats.Msg, expectedBatteryID string) {
	require.NotNil(t, msg, "Expected charging.started.v1 event, got nil")
	assert.Equal(t, "charging.started.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.NotEmpty(t, event["power"], "Power missing")
	assert.NotEmpty(t, event["target_soc"], "Target SoC missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertChargingCompleted validates a charging.completed.v1 event
func AssertChargingCompleted(t *testing.T, msg *nats.Msg, expectedBatteryID string) {
	require.NotNil(t, msg, "Expected charging.completed.v1 event, got nil")
	assert.Equal(t, "charging.completed.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.NotEmpty(t, event["final_soc"], "Final SoC missing")
	assert.NotEmpty(t, event["energy_charged"], "Energy charged missing")
	assert.NotEmpty(t, event["duration"], "Duration missing")
	assert.NotEmpty(t, event["stop_reason"], "Stop reason missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertDischargingStarted validates a discharging.started.v1 event
func AssertDischargingStarted(t *testing.T, msg *nats.Msg, expectedBatteryID string) {
	require.NotNil(t, msg, "Expected discharging.started.v1 event, got nil")
	assert.Equal(t, "discharging.started.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.NotEmpty(t, event["power"], "Power missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// AssertDischargingCompleted validates a discharging.completed.v1 event
func AssertDischargingCompleted(t *testing.T, msg *nats.Msg, expectedBatteryID string) {
	require.NotNil(t, msg, "Expected discharging.completed.v1 event, got nil")
	assert.Equal(t, "discharging.completed.v1", msg.Subject)

	var event map[string]interface{}
	err := json.Unmarshal(msg.Data, &event)
	require.NoError(t, err, "Failed to unmarshal event")

	assert.Equal(t, expectedBatteryID, event["battery_id"], "Battery ID mismatch")
	assert.NotEmpty(t, event["final_soc"], "Final SoC missing")
	assert.NotEmpty(t, event["energy_discharged"], "Energy discharged missing")
	assert.NotEmpty(t, event["duration"], "Duration missing")
	assert.NotEmpty(t, event["stop_reason"], "Stop reason missing")
	assert.NotEmpty(t, event["timestamp"], "Timestamp missing")
	assert.Equal(t, "v1", event["event_version"])
}

// GetEventField extracts a field from a NATS event message
func GetEventField(msg *nats.Msg, field string) (interface{}, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	value, ok := event[field]
	if !ok {
		return nil, fmt.Errorf("field %s not found in event", field)
	}

	return value, nil
}

// PrintEvent pretty-prints a NATS event for debugging
func PrintEvent(t *testing.T, msg *nats.Msg) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		t.Logf("Failed to unmarshal event: %v", err)
		return
	}

	prettyJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		t.Logf("Failed to pretty-print event: %v", err)
		return
	}

	t.Logf("Event [%s]:\n%s", msg.Subject, string(prettyJSON))
}
