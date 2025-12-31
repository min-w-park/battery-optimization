package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestInfrastructure is a simple test to verify the E2E test infrastructure is working
// This test demonstrates the basic usage pattern for all E2E tests
func TestInfrastructure(t *testing.T) {
	// Skip this test if services are not running
	// Remove this skip when running actual E2E tests
	t.Skip("Example test - remove skip to run against live services")

	// 1. Connect to NATS
	nc, err := ConnectToNATS(NATSURL, 5, 1*time.Second)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	// 2. Subscribe to all events
	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	// 3. Drain any existing events
	DrainEvents(sub)

	// 4. Register a battery via REST API
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("Created battery: %s", batteryID)

	// 5. Wait for battery.registered.v1 event
	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive battery.registered.v1 event")

	// 6. Print event for debugging
	PrintEvent(t, msg)

	// 7. Assert event payload
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)

	// 8. Create a market price
	err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
	require.NoError(t, err, "Failed to create price")

	// 9. Wait for market.price.updated.v1 event
	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive market.price.updated.v1 event")
	PrintEvent(t, msg)
	AssertMarketPriceUpdated(t, msg, LowPrice)

	// 10. Publish a battery state event
	err = PublishBatteryState(nc, batteryID, LowSoC, 0.0, StateIdle)
	require.NoError(t, err, "Failed to publish battery state")

	// 11. Wait for charging.opportunity.detected.v1 event
	// (This assumes Bidding Service is running in FULL_AUTO mode)
	msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive charging.opportunity.detected.v1 event")
	PrintEvent(t, msg)
	AssertChargingOpportunity(t, msg, batteryID, LowPrice, LowSoC)

	// 12. Wait for charging.command.issued.v1 event
	msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive charging.command.issued.v1 event")
	PrintEvent(t, msg)
	AssertChargingCommand(t, msg, batteryID, TargetSoCCharging, DefaultMaxPower)

	t.Log("✅ All infrastructure components working correctly!")
}

// TestFixtures verifies that test fixtures are correctly defined
func TestFixtures(t *testing.T) {
	// Test price scenarios
	t.Run("PriceScenarios", func(t *testing.T) {
		require.Less(t, LowPrice, ThresholdLow, "LowPrice should be below charging threshold")
		require.Greater(t, HighPrice, ThresholdHigh, "HighPrice should be above discharging threshold")
		require.Greater(t, MidPrice, ThresholdLow, "MidPrice should be above charging threshold")
		require.Less(t, MidPrice, ThresholdHigh, "MidPrice should be below discharging threshold")
	})

	// Test SoC scenarios
	t.Run("SoCScenarios", func(t *testing.T) {
		require.Less(t, LowSoC, TargetSoCCharging, "LowSoC should be below charging target")
		require.Greater(t, VeryHighSoC, TargetSoCDischarging, "VeryHighSoC should be above discharging target")
	})

	// Test battery specs
	t.Run("BatterySpecs", func(t *testing.T) {
		spec := DefaultBatterySpec()
		require.Equal(t, DefaultCapacity, spec.Capacity)
		require.Equal(t, DefaultMaxPower, spec.MaxPower)
		require.Equal(t, DefaultRampRate, spec.RampRate)
		require.Equal(t, DefaultEfficiency, spec.Efficiency)
	})

	// Test battery states
	t.Run("BatteryStates", func(t *testing.T) {
		idleState := IdleBatteryState()
		require.Equal(t, StateIdle, idleState.OperationState)
		require.Equal(t, 0.0, idleState.Power)

		chargingState := ChargingBatteryState()
		require.Equal(t, StateCharging, chargingState.OperationState)
		require.Greater(t, chargingState.Power, 0.0)

		dischargingState := DischargingBatteryState()
		require.Equal(t, StateDischarging, dischargingState.OperationState)
		require.Less(t, dischargingState.Power, 0.0)
	})

	// Test scenarios
	t.Run("TestScenarios", func(t *testing.T) {
		chargingScenario := GetChargingScenario()
		require.Equal(t, "CHARGING", chargingScenario.ExpectedOpportunity)
		require.True(t, chargingScenario.ExpectedCommand)
		require.Less(t, chargingScenario.Price, ThresholdLow)
		require.Less(t, chargingScenario.InitialState.SoC, TargetSoCCharging)

		dischargingScenario := GetDischargingScenario()
		require.Equal(t, "DISCHARGING", dischargingScenario.ExpectedOpportunity)
		require.True(t, dischargingScenario.ExpectedCommand)
		require.Greater(t, dischargingScenario.Price, ThresholdHigh)
		require.Greater(t, dischargingScenario.InitialState.SoC, TargetSoCDischarging)

		noOpportunityScenario := GetNoOpportunityScenario()
		require.Equal(t, "NONE", noOpportunityScenario.ExpectedOpportunity)
		require.False(t, noOpportunityScenario.ExpectedCommand)
	})
}
