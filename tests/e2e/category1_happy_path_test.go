package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Category 1: Happy Path Workflows (Priority: 🔴 Critical)
// These tests validate the core end-to-end workflows that must work for the system to be functional.

// TestChargingWorkflow validates the complete charging arbitrage workflow:
// 1. Register battery via Asset Management API
// 2. Create low price via Market Data API
// 3. Publish battery state (low SoC, IDLE)
// 4. Bidding Service detects charging opportunity
// 5. Bidding Service issues charging command (if FULL_AUTO)
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1
// - charging.opportunity.detected.v1
// - charging.command.issued.v1 (if FULL_AUTO mode)
func TestChargingWorkflow(t *testing.T) {
	// Skip if not running against live services
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// 1. Connect to NATS
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	// 2. Subscribe to all events
	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	// 3. Drain any pending events from previous tests
	DrainEvents(sub)
	t.Log("✓ Setup complete - connected to NATS and subscribed to events")

	// 4. Register battery via Asset Management API
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("✓ Battery registered: %s", batteryID)

	// 5. Wait for battery.registered.v1 event
	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive battery.registered.v1 event")
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)
	t.Log("✓ Received battery.registered.v1 event")

	// 6. Publish battery state (low SoC, IDLE) BEFORE price to ensure bidding engine has current state
	err = PublishBatteryState(nc, batteryID, LowSoC, 0.0, StateIdle)
	require.NoError(t, err, "Failed to publish battery state")
	t.Logf("✓ Published battery state: SoC=%.1f%%, State=%s", LowSoC, StateIdle)

	// Allow time for bidding engine to process battery state
	time.Sleep(100 * time.Millisecond)

	// 7. Create low price to trigger charging opportunity
	err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
	require.NoError(t, err, "Failed to create price")
	t.Logf("✓ Market price created: $%.2f/MWh", LowPrice)

	// 8. Wait for market.price.updated.v1 event
	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive market.price.updated.v1 event")
	AssertMarketPriceUpdated(t, msg, LowPrice)
	t.Log("✓ Received market.price.updated.v1 event")

	// 9. Wait for charging.opportunity.detected.v1 event
	msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive charging.opportunity.detected.v1 event")
	AssertChargingOpportunity(t, msg, batteryID, LowPrice, LowSoC)
	t.Log("✓ Received charging.opportunity.detected.v1 event")

	// 10. Wait for charging.command.issued.v1 event (if FULL_AUTO mode)
	// Note: This assumes Bidding Service is running in FULL_AUTO mode
	msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive charging.command.issued.v1 event (is AUTOMATION_MODE=FULL_AUTO?)")
	AssertChargingCommand(t, msg, batteryID, TargetSoCCharging, DefaultMaxPower)
	t.Log("✓ Received charging.command.issued.v1 event")

	t.Log("✅ Charging workflow complete - all events validated!")
}

// TestDischargingWorkflow validates the complete discharging arbitrage workflow:
// 1. Register battery via Asset Management API
// 2. Create high price via Market Data API
// 3. Publish battery state (high SoC, IDLE)
// 4. Bidding Service detects discharging opportunity
// 5. Bidding Service issues discharging command (if FULL_AUTO)
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1
// - discharging.opportunity.detected.v1
// - discharging.command.issued.v1 (if FULL_AUTO mode)
func TestDischargingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// 1. Connect to NATS
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	// 2. Subscribe to all events
	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	// 3. Drain pending events
	DrainEvents(sub)
	t.Log("✓ Setup complete - connected to NATS and subscribed to events")

	// 4. Register battery
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("✓ Battery registered: %s", batteryID)

	// 5. Wait for battery.registered.v1 event
	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive battery.registered.v1 event")
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)
	t.Log("✓ Received battery.registered.v1 event")

	// 6. Create high price to trigger discharging opportunity
	err = CreatePrice(MarketDataURL, HighPrice, DefaultInterval)
	require.NoError(t, err, "Failed to create price")
	t.Logf("✓ Market price created: $%.2f/MWh", HighPrice)

	// 7. Wait for market.price.updated.v1 event
	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive market.price.updated.v1 event")
	AssertMarketPriceUpdated(t, msg, HighPrice)
	t.Log("✓ Received market.price.updated.v1 event")

	// 8. Publish battery state (high SoC, IDLE) to trigger decision
	err = PublishBatteryState(nc, batteryID, VeryHighSoC, 0.0, StateIdle)
	require.NoError(t, err, "Failed to publish battery state")
	t.Logf("✓ Published battery state: SoC=%.1f%%, State=%s", VeryHighSoC, StateIdle)

	// 9. Wait for discharging.opportunity.detected.v1 event
	msg, err = WaitForEvent(sub, "discharging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive discharging.opportunity.detected.v1 event")
	AssertDischargingOpportunity(t, msg, batteryID, HighPrice, VeryHighSoC)
	t.Log("✓ Received discharging.opportunity.detected.v1 event")

	// 10. Wait for discharging.command.issued.v1 event (if FULL_AUTO mode)
	msg, err = WaitForEvent(sub, "discharging.command.issued.v1", EventTimeout)
	require.NoError(t, err, "Failed to receive discharging.command.issued.v1 event (is AUTOMATION_MODE=FULL_AUTO?)")
	AssertDischargingCommand(t, msg, batteryID, DefaultMaxPower)
	t.Log("✓ Received discharging.command.issued.v1 event")

	t.Log("✅ Discharging workflow complete - all events validated!")
}

// TestMultipleBatteriesWorkflow validates that the system can handle multiple batteries
// with different states and prices simultaneously:
// 1. Register two batteries
// 2. Create low price
// 3. Publish different states for each battery:
//   - Battery 1: Low SoC (should trigger charging)
//   - Battery 2: High SoC (should NOT trigger charging)
//
// 4. Verify only Battery 1 gets charging opportunity
//
// Expected Events:
// - 2x battery.registered.v1
// - 1x market.price.updated.v1
// - 1x charging.opportunity.detected.v1 (only for Battery 1)
// - 1x charging.command.issued.v1 (only for Battery 1, if FULL_AUTO)
func TestMultipleBatteriesWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// 1. Setup
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	DrainEvents(sub)
	t.Log("✓ Setup complete")

	// 2. Register first battery (will have low SoC)
	battery1ID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery 1")
	t.Logf("✓ Battery 1 registered: %s", battery1ID)

	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err)
	AssertBatteryRegistered(t, msg, battery1ID, DefaultCapacity, DefaultMaxPower)

	// 3. Register second battery (will have high SoC)
	battery2ID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery 2")
	t.Logf("✓ Battery 2 registered: %s", battery2ID)

	msg, err = WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err)
	AssertBatteryRegistered(t, msg, battery2ID, DefaultCapacity, DefaultMaxPower)

	// 4. Create low price (charging opportunity)
	err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Market price created: $%.2f/MWh", LowPrice)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, LowPrice)

	// 5. Publish Battery 1 state: Low SoC, IDLE (should trigger charging)
	err = PublishBatteryState(nc, battery1ID, LowSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery 1 state: SoC=%.1f%%, State=%s (should trigger charging)", LowSoC, StateIdle)

	// 6. Publish Battery 2 state: High SoC, IDLE (should NOT trigger charging)
	err = PublishBatteryState(nc, battery2ID, VeryHighSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery 2 state: SoC=%.1f%%, State=%s (should NOT trigger charging)", VeryHighSoC, StateIdle)

	// 7. Wait for charging.opportunity.detected.v1 - should be for Battery 1 only
	msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err)
	AssertChargingOpportunity(t, msg, battery1ID, LowPrice, LowSoC)
	t.Logf("✓ Charging opportunity detected for Battery 1 (correct)")

	// 8. Wait for charging.command.issued.v1 - should be for Battery 1 only
	msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	AssertChargingCommand(t, msg, battery1ID, TargetSoCCharging, DefaultMaxPower)
	t.Logf("✓ Charging command issued for Battery 1 (correct)")

	// 9. Verify NO additional charging events for Battery 2
	err = AssertNoEvent(sub, NoEventTimeout)
	require.NoError(t, err, "Unexpected event received - Battery 2 should NOT trigger charging")
	t.Log("✓ No charging events for Battery 2 (correct)")

	t.Log("✅ Multiple batteries workflow complete - correct filtering!")
}

// TestPriceChangeTriggersDecision validates that price changes trigger re-evaluation
// of arbitrage opportunities:
// 1. Register battery
// 2. Publish battery state (low SoC, IDLE)
// 3. Create mid-range price (no opportunity)
// 4. Verify no opportunity detected
// 5. Update price to low price (charging opportunity)
// 6. Verify charging opportunity detected
//
// Expected Events:
// - battery.registered.v1
// - market.price.updated.v1 (mid price) → no opportunity
// - market.price.updated.v1 (low price) → charging opportunity + command
func TestPriceChangeTriggersDecision(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// 1. Setup
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err, "Failed to connect to NATS")
	defer nc.Close()

	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err, "Failed to subscribe to events")
	defer sub.Unsubscribe()

	DrainEvents(sub)
	t.Log("✓ Setup complete")

	// 2. Register battery
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

	// 3. Publish battery state (low SoC, IDLE)
	err = PublishBatteryState(nc, batteryID, LowSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery state: SoC=%.1f%%, State=%s", LowSoC, StateIdle)

	// 4. Create mid-range price (no opportunity expected)
	err = CreatePrice(MarketDataURL, MidPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Market price created: $%.2f/MWh (mid-range, no opportunity)", MidPrice)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, MidPrice)

	// 5. Verify no opportunity events (mid-range price)
	err = AssertNoEvent(sub, NoEventTimeout)
	require.NoError(t, err, "Unexpected opportunity event for mid-range price")
	t.Log("✓ No opportunity detected for mid-range price (correct)")

	// 6. Update price to low price (charging opportunity expected)
	err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Market price updated: $%.2f/MWh (low, charging opportunity)", LowPrice)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, LowPrice)

	// 7. Wait for charging.opportunity.detected.v1
	msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err, "Expected charging opportunity after price drop")
	AssertChargingOpportunity(t, msg, batteryID, LowPrice, LowSoC)
	t.Log("✓ Charging opportunity detected after price change")

	// 8. Wait for charging.command.issued.v1
	msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	AssertChargingCommand(t, msg, batteryID, TargetSoCCharging, DefaultMaxPower)
	t.Log("✓ Charging command issued after price change")

	t.Log("✅ Price change workflow complete - re-evaluation working!")
}
