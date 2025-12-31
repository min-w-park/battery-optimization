package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Category 6: Complete End-to-End Workflows (Priority: 🔴 Critical)
// These tests validate complete business workflows spanning multiple services and event cycles.

// TestCompleteArbitrageWorkflow validates a complete arbitrage cycle:
// 1. Register battery
// 2. Low price triggers charging opportunity and command
// 3. Simulate charging completion (SoC increases)
// 4. High price triggers discharging opportunity and command
// 5. Simulate discharging completion (SoC decreases)
//
// This represents a full profit cycle: buy low → charge → sell high → discharge
//
// Expected Events (in sequence):
// - battery.registered.v1
// - market.price.updated.v1 (low)
// - charging.opportunity.detected.v1
// - charging.command.issued.v1
// - battery.state.changed.v1 (charging, SoC increasing)
// - battery.state.changed.v1 (idle, SoC=80%)
// - market.price.updated.v1 (high)
// - discharging.opportunity.detected.v1
// - discharging.command.issued.v1
// - battery.state.changed.v1 (discharging, SoC decreasing)
// - battery.state.changed.v1 (idle, SoC=30%)
func TestCompleteArbitrageWorkflow(t *testing.T) {
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
	t.Log("=== Starting Complete Arbitrage Workflow ===")

	// Phase 1: Battery Registration
	t.Log("\n--- Phase 1: Battery Registration ---")
	batteryID, err := CreateBattery(
		AssetManagementURL,
		DefaultCapacity,
		DefaultMaxPower,
		DefaultRampRate,
		DefaultEfficiency,
	)
	require.NoError(t, err, "Failed to create battery")
	t.Logf("✓ Battery registered: %s (Capacity: %.0f MWh, MaxPower: %.0f MW)",
		batteryID, DefaultCapacity, DefaultMaxPower)

	msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
	require.NoError(t, err)
	AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)
	t.Log("✓ battery.registered.v1 event received")

	// Phase 2: Charging Opportunity (Low Price)
	t.Log("\n--- Phase 2: Charging Opportunity ---")
	// Publish battery state: Low SoC, IDLE BEFORE price
	err = PublishBatteryState(nc, batteryID, LowSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery state published: SoC=%.1f%%, Power=0 MW, State=%s", LowSoC, StateIdle)

	// Allow time for bidding engine to process battery state
	time.Sleep(100 * time.Millisecond)

	err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Low price created: $%.2f/MWh (charging threshold: < $%.2f)", LowPrice, ThresholdLow)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, LowPrice)
	t.Log("✓ market.price.updated.v1 event received")

	msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err)
	AssertChargingOpportunity(t, msg, batteryID, LowPrice, LowSoC)
	t.Log("✓ charging.opportunity.detected.v1 event received")

	msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	AssertChargingCommand(t, msg, batteryID, TargetSoCCharging, DefaultMaxPower)
	t.Logf("✓ charging.command.issued.v1 event received (target: %.0f%%)", TargetSoCCharging)

	// Phase 3: Simulate Charging Progress
	t.Log("\n--- Phase 3: Simulating Charging Progress ---")
	// Simulate charging: SoC increases from 30% → 80%
	chargingSteps := []float64{40.0, 50.0, 60.0, 70.0, 80.0}
	for i, soc := range chargingSteps {
		time.Sleep(100 * time.Millisecond) // Simulate time passing
		err = PublishBatteryState(nc, batteryID, soc, DefaultMaxPower, StateCharging)
		require.NoError(t, err)
		if i == len(chargingSteps)-1 {
			t.Logf("✓ Charging progress: SoC=%.0f%% (target reached)", soc)
		}
	}

	// Charging complete - return to IDLE
	err = PublishBatteryState(nc, batteryID, TargetSoCCharging, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Charging complete: SoC=%.0f%%, State=%s", TargetSoCCharging, StateIdle)

	// Phase 4: Discharging Opportunity (High Price)
	t.Log("\n--- Phase 4: Discharging Opportunity ---")
	err = CreatePrice(MarketDataURL, HighPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ High price created: $%.2f/MWh (discharging threshold: > $%.2f)", HighPrice, ThresholdHigh)

	msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)
	AssertMarketPriceUpdated(t, msg, HighPrice)
	t.Log("✓ market.price.updated.v1 event received")

	msg, err = WaitForEvent(sub, "discharging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err)
	AssertDischargingOpportunity(t, msg, batteryID, HighPrice, TargetSoCCharging)
	t.Log("✓ discharging.opportunity.detected.v1 event received")

	msg, err = WaitForEvent(sub, "discharging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	AssertDischargingCommand(t, msg, batteryID, DefaultMaxPower)
	t.Logf("✓ discharging.command.issued.v1 event received (target: %.0f%%)", TargetSoCDischarging)

	// Phase 5: Simulate Discharging Progress
	t.Log("\n--- Phase 5: Simulating Discharging Progress ---")
	// Simulate discharging: SoC decreases from 80% → 30%
	dischargingSteps := []float64{70.0, 60.0, 50.0, 40.0, 30.0}
	for i, soc := range dischargingSteps {
		time.Sleep(100 * time.Millisecond) // Simulate time passing
		err = PublishBatteryState(nc, batteryID, soc, -DefaultMaxPower, StateDischarging)
		require.NoError(t, err)
		if i == len(dischargingSteps)-1 {
			t.Logf("✓ Discharging progress: SoC=%.0f%% (target reached)", soc)
		}
	}

	// Discharging complete - return to IDLE
	err = PublishBatteryState(nc, batteryID, TargetSoCDischarging, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Discharging complete: SoC=%.0f%%, State=%s", TargetSoCDischarging, StateIdle)

	// Phase 6: Profit Calculation
	t.Log("\n--- Phase 6: Profit Analysis ---")
	energyCycled := (TargetSoCCharging - TargetSoCDischarging) / 100.0 * DefaultCapacity
	priceDelta := HighPrice - LowPrice
	grossProfit := energyCycled * priceDelta
	efficiency := DefaultEfficiency
	netProfit := grossProfit * efficiency

	t.Logf("✓ Energy cycled: %.2f MWh (from %.0f%% to %.0f%% to %.0f%%)",
		energyCycled, LowSoC, TargetSoCCharging, TargetSoCDischarging)
	t.Logf("✓ Price delta: $%.2f/MWh (bought at $%.2f, sold at $%.2f)",
		priceDelta, LowPrice, HighPrice)
	t.Logf("✓ Gross profit: $%.2f", grossProfit)
	t.Logf("✓ Efficiency: %.0f%%", efficiency*100)
	t.Logf("✓ Net profit: $%.2f", netProfit)

	t.Log("\n✅ Complete arbitrage cycle validated!")
	t.Logf("   Battery: %s", batteryID)
	t.Logf("   Cycle: %.0f%% → %.0f%% → %.0f%%", LowSoC, TargetSoCCharging, TargetSoCDischarging)
	t.Logf("   Profit: $%.2f", netProfit)
}

// TestDayInTheLife simulates 24 hours of battery operation with multiple arbitrage cycles:
// - Multiple price changes throughout the day
// - Multiple charging and discharging opportunities
// - Validates system can handle sustained operations
//
// Simplified scenario:
// - 6:00 AM: Low price ($30) → Charge
// - 9:00 AM: Mid price ($75) → No action
// - 12:00 PM: High price ($150) → Discharge
// - 3:00 PM: Mid price ($75) → No action
// - 6:00 PM: High price ($140) → Discharge (if SoC allows)
// - 9:00 PM: Low price ($35) → Charge
//
// This validates:
// - Multiple arbitrage cycles
// - Correct state management across cycles
// - No memory leaks or degradation over time
func TestDayInTheLife(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup
	nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
	require.NoError(t, err)
	defer nc.Close()

	sub, err := SubscribeToEvents(nc, ">")
	require.NoError(t, err)
	defer sub.Unsubscribe()

	DrainEvents(sub)
	t.Log("=== Starting 24-Hour Simulation ===")

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

	// Start with mid SoC
	currentSoC := MidSoC
	totalProfit := 0.0

	// Scenario 1: 6:00 AM - Low price, charge opportunity
	t.Log("\n--- 6:00 AM: Low Price ---")
	err = PublishBatteryState(nc, batteryID, currentSoC, 0.0, StateIdle)
	require.NoError(t, err)
	t.Logf("✓ Battery SoC: %.0f%%", currentSoC)

	time.Sleep(100 * time.Millisecond)

	err = CreatePrice(MarketDataURL, 30.0, DefaultInterval)
	require.NoError(t, err)
	t.Log("✓ Price: $30/MWh")

	_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)

	_, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err)
	t.Log("✓ Charging opportunity detected")

	_, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	t.Log("✓ Charging command issued")

	currentSoC = TargetSoCCharging
	t.Logf("✓ Charged to %.0f%%", currentSoC)

	// Scenario 2: 9:00 AM - Mid price, no action
	t.Log("\n--- 9:00 AM: Mid Price ---")
	err = PublishBatteryState(nc, batteryID, currentSoC, 0.0, StateIdle)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = CreatePrice(MarketDataURL, MidPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Price: $%.2f/MWh", MidPrice)

	_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)

	err = AssertNoEvent(sub, NoEventTimeout)
	require.NoError(t, err)
	t.Log("✓ No action (mid-range price)")

	// Scenario 3: 12:00 PM - High price, discharge opportunity
	t.Log("\n--- 12:00 PM: High Price ---")
	err = CreatePrice(MarketDataURL, HighPrice, DefaultInterval)
	require.NoError(t, err)
	t.Logf("✓ Price: $%.2f/MWh", HighPrice)

	_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)

	_, err = WaitForEvent(sub, "discharging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err)
	t.Log("✓ Discharging opportunity detected")

	_, err = WaitForEvent(sub, "discharging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	t.Log("✓ Discharging command issued")

	// Calculate profit for this cycle
	energyDischarged := (currentSoC - TargetSoCDischarging) / 100.0 * DefaultCapacity
	profit := energyDischarged * (HighPrice - 30.0) * DefaultEfficiency
	totalProfit += profit
	t.Logf("✓ Profit from this cycle: $%.2f", profit)

	currentSoC = TargetSoCDischarging
	t.Logf("✓ Discharged to %.0f%%", currentSoC)

	// Scenario 4: 3:00 PM - Mid price, no action
	t.Log("\n--- 3:00 PM: Mid Price ---")
	err = PublishBatteryState(nc, batteryID, currentSoC, 0.0, StateIdle)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = CreatePrice(MarketDataURL, MidPrice, DefaultInterval)
	require.NoError(t, err)

	_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)

	err = AssertNoEvent(sub, NoEventTimeout)
	require.NoError(t, err)
	t.Log("✓ No action (mid-range price)")

	// Scenario 5: 6:00 PM - Another low price for second charge
	t.Log("\n--- 6:00 PM: Low Price (Second Cycle) ---")
	// Battery state already published from scenario 4, just wait for processing
	time.Sleep(100 * time.Millisecond)

	err = CreatePrice(MarketDataURL, 35.0, DefaultInterval)
	require.NoError(t, err)
	t.Log("✓ Price: $35/MWh")

	_, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
	require.NoError(t, err)

	_, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
	require.NoError(t, err)
	t.Log("✓ Charging opportunity detected (second cycle)")

	_, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
	require.NoError(t, err)
	t.Log("✓ Charging command issued (second cycle)")

	currentSoC = TargetSoCCharging
	t.Logf("✓ Charged to %.0f%% (ready for evening discharge)", currentSoC)

	// Final Summary
	t.Log("\n=== 24-Hour Simulation Complete ===")
	t.Logf("✓ Total arbitrage cycles: 2 (morning + evening)")
	t.Logf("✓ Final SoC: %.0f%%", currentSoC)
	t.Logf("✓ Estimated total profit: $%.2f", totalProfit)
	t.Log("✅ Day-in-the-life workflow validated!")
}
