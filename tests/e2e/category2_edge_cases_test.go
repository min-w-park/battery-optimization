package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Category 2: Edge Cases (Priority: 🟡 High)
// These tests validate boundary conditions and edge cases in the arbitrage algorithm.

// TestEdgeCaseChargingThreshold validates behavior at the charging price threshold ($49/MWh):
// - Price just below $50/MWh should trigger charging
// - Tests the arbitrage algorithm boundary condition
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1 ($49/MWh)
// - charging.opportunity.detected.v1 (should trigger)
// - charging.command.issued.v1 (if FULL_AUTO)
func TestEdgeCaseChargingThreshold(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	DrainEvents(sub)
	t.Log("✓ Setup complete")

	// Register battery
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("✓ Battery registered: %s", batteryID)

	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err)
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)

	// Create edge case price ($49/MWh - just below $50 threshold)
	err = CreatePrice(MarketDataURL, EdgeLowPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Edge case price created: $%.2f/MWh (threshold is $%.2f)", EdgeLowPrice, ThresholdLow)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, EdgeLowPrice)

	// Publish battery state (low SoC, IDLE)
	err = PublishBatteryState(nc, batteryID, LowSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery state: SoC=%.1f%%, State=%s", LowSoC, StateIdle)

	// Should trigger charging opportunity (< $50)
	msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err, "Expected charging opportunity at $49/MWh")
	AssertChargingOpportunity(t, msg, batteryID, EdgeLowPrice, LowSoC)
	t.Log("✓ Charging opportunity detected at edge case price (correct)")

	// Should issue charging command
	msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	AssertChargingCommand(t, msg, batteryID, TargetSoCCharging, DefaultMaxPower)
	t.Log("✓ Charging command issued at edge case price (correct)")

	t.Log("✅ Edge case charging threshold test passed!")
}

// TestEdgeCaseDischargingThreshold validates behavior at the discharging price threshold ($101/MWh):
// - Price just above $100/MWh should trigger discharging
// - Tests the arbitrage algorithm boundary condition
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1 ($101/MWh)
// - discharging.opportunity.detected.v1 (should trigger)
// - discharging.command.issued.v1 (if FULL_AUTO)
func TestEdgeCaseDischargingThreshold(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	DrainEvents(sub)
	t.Log("✓ Setup complete")

	// Register battery
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("✓ Battery registered: %s", batteryID)

	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err)
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)

	// Create edge case price ($101/MWh - just above $100 threshold)
	err = CreatePrice(MarketDataURL, EdgeHighPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Edge case price created: $%.2f/MWh (threshold is $%.2f)", EdgeHighPrice, ThresholdHigh)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, EdgeHighPrice)

	// Publish battery state (high SoC, IDLE)
	err = PublishBatteryState(nc, batteryID, VeryHighSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery state: SoC=%.1f%%, State=%s", VeryHighSoC, StateIdle)

	// Should trigger discharging opportunity (> $100)
	msg, err = WaitForEvent(sub, "discharging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err, "Expected discharging opportunity at $101/MWh")
	AssertDischargingOpportunity(t, msg, batteryID, EdgeHighPrice, VeryHighSoC)
	t.Log("✓ Discharging opportunity detected at edge case price (correct)")

	// Should issue discharging command
	msg, err = WaitForEvent(sub, "discharging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	AssertDischargingCommand(t, msg, batteryID, DefaultMaxPower)
	t.Log("✓ Discharging command issued at edge case price (correct)")

	t.Log("✅ Edge case discharging threshold test passed!")
}

// TestEdgeCaseSoCBoundaries validates behavior at SoC boundaries:
// - SoC at 80% (max charging target) - should NOT trigger charging
// - SoC at 30% (min discharging target) - should NOT trigger discharging
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1
// - NO opportunity events (boundary conditions prevent arbitrage)
func TestEdgeCaseSoCBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Test 1: SoC at 80% with low price - should NOT charge
	t.Run("HighSoC_NoCharging", func(t *testing.T) {
		nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
		require.NoError(t, err)
		defer nc.Close()

		sub, err := SubscribeToEvents(nc, ">")
		require.NoError(t, err)
		defer sub.Unsubscribe()

		DrainEvents(sub)

		// Register battery
		batteryID, err := CreateBattery(
			AssetManagementURL,
			DefaultCapacity,
			DefaultMaxPower,
			DefaultRampRate,
			DefaultEfficiency,
		)
		require.NoError(t, err)
		t.Logf("✓ Battery registered: %s", batteryID)

		_, err = WaitForEvent(sub, "battery.registered.v1", EventTimeout)
		require.NoError(t, err)

		// Low price (charging opportunity)
		err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
		require.NoError(t, err)
		t.Logf("✓ Low price: $%.2f/MWh", LowPrice)

		_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
		require.NoError(t, err)

		// Publish battery state: SoC at 80% (boundary)
		err = PublishBatteryState(nc, batteryID, HighSoC, 0.0, StateIdle)
		require.NoError(t, err)
		t.Logf("✓ Battery SoC: %.1f%% (at charging target boundary)", HighSoC)

		// Should NOT trigger charging (SoC >= 80%)
		err = AssertNoEvent(sub, NoEventTimeout)
		require.NoError(t, err, "Should NOT charge when SoC >= 80%")
		t.Log("✓ No charging at SoC=80% (correct)")
	})

	// Test 2: SoC at 30% with high price - should NOT discharge
	t.Run("LowSoC_NoDischarging", func(t *testing.T) {
		nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
		require.NoError(t, err)
		defer nc.Close()

		sub, err := SubscribeToEvents(nc, ">")
		require.NoError(t, err)
		defer sub.Unsubscribe()

		DrainEvents(sub)

		// Register battery
		batteryID, err := CreateBattery(
			AssetManagementURL,
			DefaultCapacity,
			DefaultMaxPower,
			DefaultRampRate,
			DefaultEfficiency,
		)
		require.NoError(t, err)
		t.Logf("✓ Battery registered: %s", batteryID)

		_, err = WaitForEvent(sub, "battery.registered.v1", EventTimeout)
		require.NoError(t, err)

		// High price (discharging opportunity)
		err = CreatePrice(MarketDataURL, HighPrice, DefaultInterval)
		require.NoError(t, err)
		t.Logf("✓ High price: $%.2f/MWh", HighPrice)

		_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
		require.NoError(t, err)

		// Publish battery state: SoC at 30% (boundary)
		err = PublishBatteryState(nc, batteryID, MinSoCForDischarge, 0.0, StateIdle)
		require.NoError(t, err)
		t.Logf("✓ Battery SoC: %.1f%% (at discharging minimum boundary)", MinSoCForDischarge)

		// Should NOT trigger discharging (SoC <= 30%)
		err = AssertNoEvent(sub, NoEventTimeout)
		require.NoError(t, err, "Should NOT discharge when SoC <= 30%")
		t.Log("✓ No discharging at SoC=30% (correct)")
	})

	t.Log("✅ SoC boundary tests passed!")
}

// TestNoOpportunityMidRange validates that mid-range prices and SoC don't trigger arbitrage:
// - Mid-range price ($75/MWh)
// - Mid-range SoC (50%)
// - Should produce NO opportunity events
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1
// - NO opportunity events
func TestNoOpportunityMidRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	DrainEvents(sub)
	t.Log("✓ Setup complete")

	// Register battery
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("✓ Battery registered: %s", batteryID)

	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err)
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)

	// Create mid-range price (no opportunity)
	err = CreatePrice(MarketDataURL, MidPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Mid-range price: $%.2f/MWh (between $%.2f and $%.2f)", MidPrice, ThresholdLow, ThresholdHigh)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, MidPrice)

	// Publish battery state (mid SoC, IDLE)
	err = PublishBatteryState(nc, batteryID, MidSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery state: SoC=%.1f%%, State=%s (mid-range)", MidSoC, StateIdle)

	// Should NOT trigger any opportunity
	err = AssertNoEvent(sub, NoEventTimeout)
	require.NoError(t, err, "Should NOT trigger any arbitrage opportunity in mid-range")
	t.Log("✓ No opportunity in mid-range (correct)")

	t.Log("✅ No opportunity mid-range test passed!")
}
