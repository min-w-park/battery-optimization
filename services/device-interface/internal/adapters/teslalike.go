package adapters

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/minwook/battery-optimization/services/device-interface/internal/domain"
)

// TeslaLike is a mock battery adapter simulating Tesla-like behavior
type TeslaLike struct {
	batteryID string
	capacity  float64 // MWh
	maxPower  float64 // MW

	// State (protected by mutex)
	mu               sync.RWMutex
	soc              float64 // 0-100%
	power            float64 // MW (current power, may be ramping)
	targetPower      float64 // MW (target power from command)
	temperature      float64 // °C
	operationState   string
	voltage          float64 // V
	current          float64 // A
	customAttributes map[string]interface{}

	// Simulation control
	stopChan chan struct{}
	ticker   *time.Ticker
}

const (
	// Tesla characteristics
	teslaRampRate      = 5.0   // MW/s
	teslaEfficiency    = 0.95  // 95% round-trip
	teslaTempRisePerMW = 0.5   // °C per MW (divided by 10 for 100ms ticks)
	teslaCoolingRate   = 0.1   // °C per 100ms when idle
	teslaInitialSoC    = 50.0  // %
	teslaInitialTemp   = 25.0  // °C
	teslaBaseVoltage   = 800.0 // V
	teslaTickDuration  = 100 * time.Millisecond
)

// NewTeslaLike creates a new Tesla-like battery adapter
func NewTeslaLike(batteryID string, capacity, maxPower float64) *TeslaLike {
	adapter := &TeslaLike{
		batteryID:      batteryID,
		capacity:       capacity,
		maxPower:       maxPower,
		soc:            teslaInitialSoC,
		power:          0.0,
		targetPower:    0.0,
		temperature:    teslaInitialTemp,
		operationState: domain.OperationStateIdle,
		voltage:        teslaBaseVoltage,
		current:        0.0,
		stopChan:       make(chan struct{}),
		customAttributes: map[string]interface{}{
			"vendor":          "Tesla",
			"model":           "Megapack",
			"firmwareVersion": "1.2.3",
			"cellCount":       4320,
			"thermalZones":    3,
		},
	}

	// Start simulation goroutine
	adapter.ticker = time.NewTicker(teslaTickDuration)
	go adapter.simulate()

	return adapter
}

// GetState retrieves current battery state
func (a *TeslaLike) GetState(ctx context.Context) (domain.BatteryState, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return domain.BatteryState{
		SoC:            a.soc,
		Power:          a.power,
		Temperature:    a.temperature,
		Voltage:        a.voltage,
		Current:        a.current,
		OperationState: a.operationState,
	}, nil
}

// SendCommand sends a command to the battery
func (a *TeslaLike) SendCommand(ctx context.Context, cmd domain.Command) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch cmd.Type {
	case domain.CommandCharge:
		a.operationState = domain.OperationStateCharging
		a.targetPower = -cmd.Power // Negative for charging
	case domain.CommandDischarge:
		a.operationState = domain.OperationStateDischarging
		a.targetPower = cmd.Power // Positive for discharging
	case domain.CommandIdle:
		a.operationState = domain.OperationStateIdle
		a.targetPower = 0.0
	case domain.CommandFcasResponse:
		a.operationState = domain.OperationStateFCAS
		a.targetPower = cmd.Power // Can be positive or negative
	}

	return nil
}

// GetCustomAttributes returns vendor-specific attributes
func (a *TeslaLike) GetCustomAttributes() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.customAttributes
}

// GetBatteryID returns the battery ID
func (a *TeslaLike) GetBatteryID() string {
	return a.batteryID
}

// Stop stops the simulation goroutine
func (a *TeslaLike) Stop() {
	close(a.stopChan)
	a.ticker.Stop()
}

// simulate runs the battery state simulation
func (a *TeslaLike) simulate() {
	for {
		select {
		case <-a.stopChan:
			return
		case <-a.ticker.C:
			a.tick()
		}
	}
}

// tick updates battery state (called every 100ms)
func (a *TeslaLike) tick() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Ramp power towards target (5 MW/s = 0.5 MW per 100ms)
	rampStep := teslaRampRate * teslaTickDuration.Seconds()
	if math.Abs(a.targetPower-a.power) <= rampStep {
		a.power = a.targetPower
	} else if a.targetPower > a.power {
		a.power += rampStep
	} else {
		a.power -= rampStep
	}

	// Update SoC based on power
	if a.power < 0 {
		// Charging: SoC increases, accounting for efficiency
		energyCharged := math.Abs(a.power) * teslaEfficiency * teslaTickDuration.Hours()
		socIncrease := (energyCharged / a.capacity) * 100.0
		a.soc += socIncrease
	} else if a.power > 0 {
		// Discharging: SoC decreases
		energyDischarged := a.power * teslaTickDuration.Hours()
		socDecrease := (energyDischarged / a.capacity) * 100.0
		a.soc -= socDecrease
	}

	// Boundary checks for SoC
	if a.soc < 0 {
		a.soc = 0
		a.power = 0
		a.targetPower = 0
		a.operationState = domain.OperationStateIdle
	}
	if a.soc > 100 {
		a.soc = 100
		a.power = 0
		a.targetPower = 0
		a.operationState = domain.OperationStateIdle
	}

	// Update temperature
	if a.power != 0 {
		// Temperature rises when charging or discharging
		// 0.5°C per MW becomes 0.05°C per MW per 100ms tick
		tempIncrease := teslaTempRisePerMW * math.Abs(a.power) * teslaTickDuration.Seconds() * 10
		a.temperature += tempIncrease
	} else {
		// Temperature decreases when idle (cooling)
		a.temperature -= teslaCoolingRate
		if a.temperature < teslaInitialTemp {
			a.temperature = teslaInitialTemp
		}
	}

	// Update voltage (simplified: slightly varies with SoC)
	a.voltage = teslaBaseVoltage * (0.95 + 0.1*(a.soc/100.0))

	// Update current (I = P/V, convert MW to W and V)
	if a.voltage > 0 {
		a.current = (a.power * 1e6) / a.voltage // MW to W, then I = P/V
	} else {
		a.current = 0
	}
}
