package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/minwook/battery-optimization/services/device-interface/internal/domain"
)

func TestNewTeslaLike(t *testing.T) {
	// Given: Battery parameters
	batteryID := "battery-123"
	capacity := 200.0 // MWh
	maxPower := 100.0 // MW

	// When: Creating adapter
	adapter := NewTeslaLike(batteryID, capacity, maxPower)

	// Then: Adapter is configured correctly
	assert.NotNil(t, adapter)
	assert.Equal(t, batteryID, adapter.GetBatteryID())

	customAttrs := adapter.GetCustomAttributes()
	assert.Equal(t, "Tesla", customAttrs["vendor"])
	assert.Equal(t, "Megapack", customAttrs["model"])
	assert.Equal(t, "1.2.3", customAttrs["firmwareVersion"])
	assert.Equal(t, 4320, customAttrs["cellCount"])
	assert.Equal(t, 3, customAttrs["thermalZones"])
}

func TestTeslaLike_GetState_InitialState(t *testing.T) {
	// Given: New adapter
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)

	// When: Getting initial state
	ctx := context.Background()
	state, err := adapter.GetState(ctx)

	// Then: Initial state is correct
	require.NoError(t, err)
	assert.Equal(t, 50.0, state.SoC) // Initial SoC
	assert.Equal(t, 0.0, state.Power)
	assert.Equal(t, 25.0, state.Temperature) // Initial temperature
	assert.Equal(t, domain.OperationStateIdle, state.OperationState)
}

func TestTeslaLike_SendCommand_Charge(t *testing.T) {
	// Given: Adapter with initial state
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)

	// When: Sending charge command
	ctx := context.Background()
	cmd := domain.Command{
		Type:      domain.CommandCharge,
		Power:     15.0, // 15 MW
		TargetSoC: 80.0,
		Duration:  0, // Indefinite
	}
	err := adapter.SendCommand(ctx, cmd)

	// Then: Command accepted
	require.NoError(t, err)

	// Wait for simulation to update
	time.Sleep(200 * time.Millisecond)

	// State should reflect charging
	state, err := adapter.GetState(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.OperationStateCharging, state.OperationState)
	assert.True(t, state.Power < 0, "Power should be negative when charging")
	assert.True(t, state.SoC > 50.0, "SoC should increase")
	assert.True(t, state.Temperature > 25.0, "Temperature should increase")
}

func TestTeslaLike_SendCommand_Discharge(t *testing.T) {
	// Given: Adapter with initial state
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)

	// When: Sending discharge command
	ctx := context.Background()
	cmd := domain.Command{
		Type:     domain.CommandDischarge,
		Power:    25.0, // 25 MW
		Duration: 0,    // Indefinite
	}
	err := adapter.SendCommand(ctx, cmd)

	// Then: Command accepted
	require.NoError(t, err)

	// Wait for simulation to update
	time.Sleep(200 * time.Millisecond)

	// State should reflect discharging
	state, err := adapter.GetState(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.OperationStateDischarging, state.OperationState)
	assert.True(t, state.Power > 0, "Power should be positive when discharging")
	assert.True(t, state.SoC < 50.0, "SoC should decrease")
	assert.True(t, state.Temperature > 25.0, "Temperature should increase")
}

func TestTeslaLike_SendCommand_Idle(t *testing.T) {
	// Given: Adapter in charging state
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)
	ctx := context.Background()

	// Start charging
	chargeCmd := domain.Command{
		Type:  domain.CommandCharge,
		Power: 15.0,
	}
	adapter.SendCommand(ctx, chargeCmd)
	time.Sleep(100 * time.Millisecond)

	// When: Sending idle command
	idleCmd := domain.Command{
		Type: domain.CommandIdle,
	}
	err := adapter.SendCommand(ctx, idleCmd)

	// Then: Command accepted
	require.NoError(t, err)

	// Wait for simulation to update
	time.Sleep(200 * time.Millisecond)

	// State should be idle
	state, err := adapter.GetState(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.OperationStateIdle, state.OperationState)
	assert.Equal(t, 0.0, state.Power)
}

func TestTeslaLike_RampRate(t *testing.T) {
	// Given: Adapter with 5 MW/s ramp rate
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)
	ctx := context.Background()

	// When: Sending 50 MW discharge command
	cmd := domain.Command{
		Type:  domain.CommandDischarge,
		Power: 50.0,
	}
	err := adapter.SendCommand(ctx, cmd)
	require.NoError(t, err)

	// Then: Power should ramp gradually
	// After 1 second, power should be ~5 MW (5 MW/s ramp rate)
	time.Sleep(1 * time.Second)
	state, err := adapter.GetState(ctx)
	require.NoError(t, err)
	// Power should be ramping up (not instantly at 50 MW)
	assert.True(t, state.Power > 0 && state.Power < 50.0, "Power should be ramping: %.2f", state.Power)

	// After 10 seconds, power should reach target 50 MW
	time.Sleep(9 * time.Second)
	state, err = adapter.GetState(ctx)
	require.NoError(t, err)
	assert.InDelta(t, 50.0, state.Power, 5.0, "Power should reach target after ramp time")
}

func TestTeslaLike_SoCBoundaries(t *testing.T) {
	// Given: Adapter starting at 50% SoC
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)
	ctx := context.Background()

	// When: Discharging at high power to test lower boundary
	cmd := domain.Command{
		Type:     domain.CommandDischarge,
		Power:    100.0, // Max power
		Duration: 0,
	}
	err := adapter.SendCommand(ctx, cmd)
	require.NoError(t, err)

	// Wait long enough to potentially go below 0%
	time.Sleep(3 * time.Second)

	// Then: SoC should not go below 0%
	state, err := adapter.GetState(ctx)
	require.NoError(t, err)
	assert.True(t, state.SoC >= 0.0, "SoC should not go below 0%%: %.2f", state.SoC)
	assert.True(t, state.SoC <= 100.0, "SoC should not exceed 100%%: %.2f", state.SoC)
}

func TestTeslaLike_TemperatureCooling(t *testing.T) {
	// Given: Adapter in idle state with elevated temperature
	adapter := NewTeslaLike("battery-123", 200.0, 100.0)
	ctx := context.Background()

	// Heat up the battery by charging
	chargeCmd := domain.Command{
		Type:  domain.CommandCharge,
		Power: 50.0,
	}
	adapter.SendCommand(ctx, chargeCmd)
	time.Sleep(3 * time.Second)

	// When: Switching to idle
	idleCmd := domain.Command{
		Type: domain.CommandIdle,
	}
	adapter.SendCommand(ctx, idleCmd)

	// Wait for power to ramp down to zero (50 MW at 5 MW/s = 10 seconds)
	time.Sleep(11 * time.Second)

	// Get temperature after power has ramped down
	state1, _ := adapter.GetState(ctx)
	temp1 := state1.Temperature

	// Wait for cooling
	time.Sleep(3 * time.Second)

	// Then: Temperature should decrease (cooling)
	state2, _ := adapter.GetState(ctx)
	temp2 := state2.Temperature
	assert.True(t, temp2 < temp1, "Temperature should decrease when idle: %.2f -> %.2f", temp1, temp2)
}
