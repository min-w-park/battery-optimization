package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/services/bidding/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate_NewBattery(t *testing.T) {
	c := cache.NewBatteryStateCache()

	state := cache.BatteryState{
		BatteryID:      "battery-123",
		SoC:            50.0,
		Power:          0.0,
		OperationState: "IDLE",
		Temperature:    25.0,
		Voltage:        800.0,
		Current:        0.0,
		Timestamp:      time.Now(),
	}

	c.Update("battery-123", state)

	retrieved, found := c.Get("battery-123")
	require.True(t, found)
	assert.Equal(t, state.BatteryID, retrieved.BatteryID)
	assert.Equal(t, state.SoC, retrieved.SoC)
	assert.Equal(t, state.OperationState, retrieved.OperationState)
}

func TestUpdate_ExistingBattery(t *testing.T) {
	c := cache.NewBatteryStateCache()

	// Initial state
	state1 := cache.BatteryState{
		BatteryID:      "battery-123",
		SoC:            50.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
	}
	c.Update("battery-123", state1)

	// Updated state
	state2 := cache.BatteryState{
		BatteryID:      "battery-123",
		SoC:            55.0,
		OperationState: "CHARGING",
		Timestamp:      time.Now(),
	}
	c.Update("battery-123", state2)

	retrieved, found := c.Get("battery-123")
	require.True(t, found)
	assert.Equal(t, 55.0, retrieved.SoC)
	assert.Equal(t, "CHARGING", retrieved.OperationState)
}

func TestGet_Exists(t *testing.T) {
	c := cache.NewBatteryStateCache()

	state := cache.BatteryState{
		BatteryID:      "battery-123",
		SoC:            50.0,
		OperationState: "IDLE",
		Timestamp:      time.Now(),
	}
	c.Update("battery-123", state)

	retrieved, found := c.Get("battery-123")
	assert.True(t, found)
	assert.Equal(t, "battery-123", retrieved.BatteryID)
}

func TestGet_NotFound(t *testing.T) {
	c := cache.NewBatteryStateCache()

	_, found := c.Get("non-existent")
	assert.False(t, found)
}

func TestList_AllBatteries(t *testing.T) {
	c := cache.NewBatteryStateCache()

	// Add multiple batteries
	state1 := cache.BatteryState{BatteryID: "battery-1", SoC: 50.0, Timestamp: time.Now()}
	state2 := cache.BatteryState{BatteryID: "battery-2", SoC: 60.0, Timestamp: time.Now()}
	state3 := cache.BatteryState{BatteryID: "battery-3", SoC: 70.0, Timestamp: time.Now()}

	c.Update("battery-1", state1)
	c.Update("battery-2", state2)
	c.Update("battery-3", state3)

	list := c.List()
	assert.Len(t, list, 3)

	// Check all batteries are in the list
	ids := make(map[string]bool)
	for _, state := range list {
		ids[state.BatteryID] = true
	}
	assert.True(t, ids["battery-1"])
	assert.True(t, ids["battery-2"])
	assert.True(t, ids["battery-3"])
}

func TestList_EmptyCache(t *testing.T) {
	c := cache.NewBatteryStateCache()

	list := c.List()
	assert.Empty(t, list)
}

func TestUpdate_Concurrent(t *testing.T) {
	c := cache.NewBatteryStateCache()
	const numGoroutines = 100
	const numBatteries = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numBatteries; j++ {
				batteryID := string(rune('A' + j))
				state := cache.BatteryState{
					BatteryID:      batteryID,
					SoC:            float64(id),
					OperationState: "IDLE",
					Timestamp:      time.Now(),
				}
				c.Update(batteryID, state)
			}
		}(i)
	}

	wg.Wait()

	// Verify all batteries exist
	list := c.List()
	assert.Len(t, list, numBatteries)
}

func TestGet_Concurrent(t *testing.T) {
	c := cache.NewBatteryStateCache()

	// Populate cache
	for i := 0; i < 10; i++ {
		batteryID := string(rune('A' + i))
		state := cache.BatteryState{
			BatteryID:      batteryID,
			SoC:            50.0,
			OperationState: "IDLE",
			Timestamp:      time.Now(),
		}
		c.Update(batteryID, state)
	}

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				batteryID := string(rune('A' + j))
				_, found := c.Get(batteryID)
				assert.True(t, found)
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentReadWrite(t *testing.T) {
	c := cache.NewBatteryStateCache()
	const duration = 100 * time.Millisecond

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		start := time.Now()
		i := 0
		for time.Since(start) < duration {
			state := cache.BatteryState{
				BatteryID:      "battery-test",
				SoC:            float64(i % 100),
				OperationState: "IDLE",
				Timestamp:      time.Now(),
			}
			c.Update("battery-test", state)
			i++
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		start := time.Now()
		for time.Since(start) < duration {
			c.Get("battery-test")
			c.List()
		}
	}()

	wg.Wait()
}
