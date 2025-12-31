package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/bidding/internal/cache"
	"github.com/minwook/battery-optimization/services/bidding/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockPublisher implements events.EventPublisher for testing
type MockPublisher struct {
	published []PublishedEvent
}

type PublishedEvent struct {
	subject string
	data    interface{}
}

func (m *MockPublisher) Publish(ctx context.Context, subject string, data interface{}) error {
	m.published = append(m.published, PublishedEvent{subject, data})
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}

func (m *MockPublisher) GetPublished(subject string) []interface{} {
	var result []interface{}
	for _, p := range m.published {
		if p.subject == subject {
			result = append(result, p.data)
		}
	}
	return result
}

func (m *MockPublisher) Reset() {
	m.published = nil
}

func TestHandleBatteryStateChanged_ChargingOpportunity(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "FULL_AUTO", zap.NewNop())

	// Set up price (low price = charging opportunity)
	priceCache.Update(30.0, time.Now())

	// Battery state change (low SoC, IDLE)
	event := events.BatteryStateChanged{
		BatteryID:      "battery-123",
		SoC:            50.0,
		Power:          0.0,
		OperationState: "IDLE",
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	err := engine.HandleBatteryStateChanged(event)
	require.NoError(t, err)

	// Should publish ChargingOpportunityDetected
	opportunities := publisher.GetPublished("charging.opportunity.detected.v1")
	require.Len(t, opportunities, 1)

	opportunity := opportunities[0].(events.ChargingOpportunityDetected)
	assert.Equal(t, "battery-123", opportunity.BatteryID)
	assert.Equal(t, 30.0, opportunity.Price)
	assert.Equal(t, 50.0, opportunity.SoC)
	assert.Equal(t, 80.0, opportunity.TargetSoC)

	// Should publish ChargingCommandIssued (FULL_AUTO mode)
	commands := publisher.GetPublished("charging.command.issued.v1")
	require.Len(t, commands, 1)

	command := commands[0].(events.ChargingCommandIssued)
	assert.Equal(t, "battery-123", command.BatteryID)
	assert.Equal(t, 80.0, command.TargetSoC)
	assert.Contains(t, command.Reason, "30")
	assert.Equal(t, "BIDDING_SERVICE_AUTO", command.IssuedBy)
}

func TestHandleBatteryStateChanged_DischargingOpportunity(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "FULL_AUTO", zap.NewNop())

	// Set up price (high price = discharging opportunity)
	priceCache.Update(150.0, time.Now())

	// Battery state change (high SoC, IDLE)
	event := events.BatteryStateChanged{
		BatteryID:      "battery-123",
		SoC:            70.0,
		Power:          0.0,
		OperationState: "IDLE",
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	err := engine.HandleBatteryStateChanged(event)
	require.NoError(t, err)

	// Should publish DischargingOpportunityDetected
	opportunities := publisher.GetPublished("discharging.opportunity.detected.v1")
	require.Len(t, opportunities, 1)

	opportunity := opportunities[0].(events.DischargingOpportunityDetected)
	assert.Equal(t, "battery-123", opportunity.BatteryID)
	assert.Equal(t, 150.0, opportunity.Price)
	assert.Equal(t, 70.0, opportunity.SoC)
	assert.Equal(t, 30.0, opportunity.TargetSoC)

	// Should publish DischargingCommandIssued (FULL_AUTO mode)
	commands := publisher.GetPublished("discharging.command.issued.v1")
	require.Len(t, commands, 1)

	command := commands[0].(events.DischargingCommandIssued)
	assert.Equal(t, "battery-123", command.BatteryID)
	assert.Equal(t, 100.0, command.StopConditions.PriceThreshold)
	assert.Equal(t, 30.0, command.StopConditions.TargetSoC)
	assert.Contains(t, command.Reason, "150")
	assert.Equal(t, "BIDDING_SERVICE_AUTO", command.IssuedBy)
}

func TestHandleBatteryStateChanged_NoOpportunity(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "FULL_AUTO", zap.NewNop())

	// Set up mid-range price (no opportunity)
	priceCache.Update(75.0, time.Now())

	// Battery state change
	event := events.BatteryStateChanged{
		BatteryID:      "battery-123",
		SoC:            60.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	err := engine.HandleBatteryStateChanged(event)
	require.NoError(t, err)

	// Should NOT publish any opportunity events
	assert.Empty(t, publisher.GetPublished("charging.opportunity.detected.v1"))
	assert.Empty(t, publisher.GetPublished("discharging.opportunity.detected.v1"))
	assert.Empty(t, publisher.GetPublished("charging.command.issued.v1"))
	assert.Empty(t, publisher.GetPublished("discharging.command.issued.v1"))
}

func TestHandleMarketPriceUpdated_TriggersDecision(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "FULL_AUTO", zap.NewNop())

	// Add battery to cache first
	batteryCache.Update("battery-123", cache.BatteryState{
		BatteryID:      "battery-123",
		SoC:            50.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
	})

	// Price update (low price)
	event := events.MarketPriceUpdated{
		PriceID:      "price-1",
		Price:        30.0,
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	err := engine.HandleMarketPriceUpdated(event)
	require.NoError(t, err)

	// Should publish ChargingOpportunityDetected
	opportunities := publisher.GetPublished("charging.opportunity.detected.v1")
	assert.Len(t, opportunities, 1)

	// Should publish ChargingCommandIssued (FULL_AUTO)
	commands := publisher.GetPublished("charging.command.issued.v1")
	assert.Len(t, commands, 1)
}

func TestHandleBatteryRegistered_AddsToCache(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "MANUAL", zap.NewNop())

	event := events.BatteryRegistered{
		BatteryID:    "battery-123",
		Capacity:     200.0,
		MaxPower:     100.0,
		RampRate:     10.0,
		Efficiency:   0.95,
		Location:     "SA",
		Manufacturer: "Tesla",
		Constraints: events.BatteryConstraints{
			MinSoC:              0.1,
			MaxSoC:              0.9,
			WarrantyEOL:         0.8,
			MaxCycles:           10000,
			OperatingTempMin:    -20.0,
			OperatingTempMax:    60.0,
			GridComplianceLevel: "AS4777",
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	err := engine.HandleBatteryRegistered(event)
	require.NoError(t, err)

	// Battery should be in cache with default state
	state, found := batteryCache.Get("battery-123")
	require.True(t, found)
	assert.Equal(t, "battery-123", state.BatteryID)
	assert.Equal(t, 50.0, state.SoC) // Default SoC
	assert.Equal(t, "IDLE", state.OperationState)
}

func TestAutomationMode_Manual(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	// MANUAL mode
	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "MANUAL", zap.NewNop())

	priceCache.Update(30.0, time.Now())

	event := events.BatteryStateChanged{
		BatteryID:      "battery-123",
		SoC:            50.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	err := engine.HandleBatteryStateChanged(event)
	require.NoError(t, err)

	// Should publish opportunity event
	opportunities := publisher.GetPublished("charging.opportunity.detected.v1")
	assert.Len(t, opportunities, 1)

	// Should NOT publish command event (MANUAL mode)
	commands := publisher.GetPublished("charging.command.issued.v1")
	assert.Empty(t, commands)
}

func TestAutomationMode_FullAuto(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache()
	publisher := &MockPublisher{}

	// FULL_AUTO mode
	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "FULL_AUTO", zap.NewNop())

	priceCache.Update(30.0, time.Now())

	event := events.BatteryStateChanged{
		BatteryID:      "battery-123",
		SoC:            50.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	err := engine.HandleBatteryStateChanged(event)
	require.NoError(t, err)

	// Should publish opportunity event
	opportunities := publisher.GetPublished("charging.opportunity.detected.v1")
	assert.Len(t, opportunities, 1)

	// Should publish command event (FULL_AUTO mode)
	commands := publisher.GetPublished("charging.command.issued.v1")
	assert.Len(t, commands, 1)
}

func TestNoPriceInCache(t *testing.T) {
	batteryCache := cache.NewBatteryStateCache()
	priceCache := cache.NewPriceCache() // Empty
	publisher := &MockPublisher{}

	engine := service.NewBiddingEngine(batteryCache, priceCache, publisher, "FULL_AUTO", zap.NewNop())

	event := events.BatteryStateChanged{
		BatteryID:      "battery-123",
		SoC:            50.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	err := engine.HandleBatteryStateChanged(event)
	require.NoError(t, err)

	// Should not publish anything (no price available)
	assert.Empty(t, publisher.published)
}
