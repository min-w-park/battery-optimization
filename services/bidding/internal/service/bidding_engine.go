package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/bidding/internal/cache"
	"github.com/minwook/battery-optimization/services/bidding/internal/domain"
)

// BiddingEngine processes events and makes bidding decisions
type BiddingEngine struct {
	batteryCache   *cache.BatteryStateCache
	priceCache     *cache.PriceCache
	publisher      events.EventPublisher
	automationMode string // MANUAL, SEMI_AUTO, FULL_AUTO
	log            *zap.Logger
}

// NewBiddingEngine creates a new bidding engine
func NewBiddingEngine(
	batteryCache *cache.BatteryStateCache,
	priceCache *cache.PriceCache,
	publisher events.EventPublisher,
	automationMode string,
	log *zap.Logger,
) *BiddingEngine {
	return &BiddingEngine{
		batteryCache:   batteryCache,
		priceCache:     priceCache,
		publisher:      publisher,
		automationMode: automationMode,
		log:            log,
	}
}

// HandleBatteryStateChanged processes battery state change events
func (e *BiddingEngine) HandleBatteryStateChanged(event events.BatteryStateChanged) error {
	// 1. Update battery cache
	batteryState := cache.BatteryState{
		BatteryID:      event.BatteryID,
		SoC:            event.SoC,
		Power:          event.Power,
		OperationState: event.OperationState,
		Temperature:    event.Temperature,
		Voltage:        event.Voltage,
		Current:        event.Current,
		Timestamp:      event.Timestamp,
	}
	e.batteryCache.Update(event.BatteryID, batteryState)

	// 2. Get latest price
	price, _, priceExists := e.priceCache.GetLatest()
	if !priceExists {
		// No price available yet - can't make decisions
		return nil
	}

	// 3. Run arbitrage algorithm and publish events
	return e.evaluateBattery(event.BatteryID, batteryState, price)
}

// HandleMarketPriceUpdated processes market price update events
func (e *BiddingEngine) HandleMarketPriceUpdated(event events.MarketPriceUpdated) error {
	// 1. Update price cache
	e.priceCache.Update(event.Price, event.Timestamp)

	// 2. Evaluate all batteries with new price
	batteries := e.batteryCache.List()
	for _, batteryState := range batteries {
		if err := e.evaluateBattery(batteryState.BatteryID, batteryState, event.Price); err != nil {
			e.log.Warn("failed to evaluate battery", zap.String("battery_id", batteryState.BatteryID), zap.Error(err))
		}
	}

	return nil
}

// HandleBatteryRegistered processes battery registration events
func (e *BiddingEngine) HandleBatteryRegistered(event events.BatteryRegistered) error {
	// Initialize battery in cache with default state
	defaultState := cache.BatteryState{
		BatteryID:      event.BatteryID,
		SoC:            50.0, // Default SoC
		Power:          0.0,
		OperationState: "IDLE",
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		Timestamp:      event.Timestamp,
	}
	e.batteryCache.Update(event.BatteryID, defaultState)

	e.log.Info("battery registered in bidding engine", zap.String("battery_id", event.BatteryID))
	return nil
}

// evaluateBattery runs arbitrage algorithm and publishes appropriate events
func (e *BiddingEngine) evaluateBattery(batteryID string, state cache.BatteryState, price float64) error {
	// Check if should charge
	if domain.ShouldCharge(price, state.SoC, state.OperationState) {
		return e.handleChargingOpportunity(batteryID, state, price)
	}

	// Check if should discharge
	if domain.ShouldDischarge(price, state.SoC, state.OperationState) {
		return e.handleDischargingOpportunity(batteryID, state, price)
	}

	// No opportunity
	return nil
}

// handleChargingOpportunity publishes charging opportunity and command events
func (e *BiddingEngine) handleChargingOpportunity(batteryID string, state cache.BatteryState, price float64) error {
	targetSoC := 80.0
	expectedProfit := calculateExpectedProfit(price, domain.CHARGE_THRESHOLD_MWH, 200.0, 0.95)

	// Always publish opportunity event
	opportunityEvent := events.ChargingOpportunityDetected{
		BatteryID:      batteryID,
		Price:          price,
		SoC:            state.SoC,
		TargetSoC:      targetSoC,
		ExpectedProfit: expectedProfit,
		Timestamp:      time.Now(),
		EventVersion:   "v1",
	}

	if err := e.publishEvent("charging.opportunity.detected.v1", opportunityEvent); err != nil {
		e.log.Warn("failed to publish charging opportunity", zap.Error(err))
	}

	// Publish command event if FULL_AUTO mode
	if e.automationMode == "FULL_AUTO" {
		commandEvent := events.ChargingCommandIssued{
			BatteryID:    batteryID,
			TargetSoC:    targetSoC,
			MaxPower:     100.0,
			Reason:       fmt.Sprintf("Low price ($%.2f/MWh) detected, charging to %.0f%% SoC", price, targetSoC),
			IssuedBy:     "BIDDING_SERVICE_AUTO",
			Timestamp:    time.Now(),
			EventVersion: "v1",
		}

		if err := e.publishEvent("charging.command.issued.v1", commandEvent); err != nil {
			e.log.Warn("failed to publish charging command", zap.Error(err))
		}
	}

	return nil
}

// handleDischargingOpportunity publishes discharging opportunity and command events
func (e *BiddingEngine) handleDischargingOpportunity(batteryID string, state cache.BatteryState, price float64) error {
	targetSoC := 30.0
	expectedProfit := calculateExpectedProfit(price, domain.DISCHARGE_THRESHOLD_MWH, 200.0, 0.95)

	// Always publish opportunity event
	opportunityEvent := events.DischargingOpportunityDetected{
		BatteryID:      batteryID,
		Price:          price,
		SoC:            state.SoC,
		TargetSoC:      targetSoC,
		ExpectedProfit: expectedProfit,
		StopConditions: events.OpportunityStopConditions{
			PriceThreshold: domain.DISCHARGE_THRESHOLD_MWH,
			Duration:       60,
			FcasDispatch:   true,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	if err := e.publishEvent("discharging.opportunity.detected.v1", opportunityEvent); err != nil {
		e.log.Warn("failed to publish discharging opportunity", zap.Error(err))
	}

	// Publish command event if FULL_AUTO mode
	if e.automationMode == "FULL_AUTO" {
		commandEvent := events.DischargingCommandIssued{
			BatteryID: batteryID,
			Power:     100.0,
			StopConditions: events.CommandStopConditions{
				PriceThreshold: domain.DISCHARGE_THRESHOLD_MWH,
				TargetSoC:      targetSoC,
				Duration:       60,
				FcasDispatch:   true,
			},
			Reason:       fmt.Sprintf("High price ($%.2f/MWh) detected, discharging to %.0f%% SoC", price, targetSoC),
			IssuedBy:     "BIDDING_SERVICE_AUTO",
			Timestamp:    time.Now(),
			EventVersion: "v1",
		}

		if err := e.publishEvent("discharging.command.issued.v1", commandEvent); err != nil {
			e.log.Warn("failed to publish discharging command", zap.Error(err))
		}
	}

	return nil
}

// publishEvent publishes an event with best-effort delivery
func (e *BiddingEngine) publishEvent(subject string, event interface{}) error {
	if e.publisher == nil {
		return nil // No publisher configured
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return e.publisher.Publish(ctx, subject, event)
}

// calculateExpectedProfit estimates profit from arbitrage opportunity
func calculateExpectedProfit(price, threshold, capacity, efficiency float64) float64 {
	priceDelta := price - threshold
	if priceDelta < 0 {
		priceDelta = -priceDelta
	}
	return priceDelta * capacity * efficiency
}
