package cache

import (
	"sync"
	"time"
)

// BatteryState represents the current state of a battery
type BatteryState struct {
	BatteryID      string
	SoC            float64   // State of Charge (%)
	Power          float64   // Current power (MW, positive=discharge, negative=charge)
	OperationState string    // IDLE, CHARGING, DISCHARGING, FCAS
	Temperature    float64   // Temperature (°C)
	Voltage        float64   // Voltage (V)
	Current        float64   // Current (A)
	Timestamp      time.Time // When this state was recorded
}

// BatteryStateCache provides thread-safe caching of battery states
type BatteryStateCache struct {
	mu     sync.RWMutex
	states map[string]BatteryState // batteryID -> state
}

// NewBatteryStateCache creates a new battery state cache
func NewBatteryStateCache() *BatteryStateCache {
	return &BatteryStateCache{
		states: make(map[string]BatteryState),
	}
}

// Update updates the state for a battery (creates if doesn't exist)
func (c *BatteryStateCache) Update(batteryID string, state BatteryState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states[batteryID] = state
}

// Get retrieves the state for a battery
// Returns the state and true if found, zero value and false if not found
func (c *BatteryStateCache) Get(batteryID string) (BatteryState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state, found := c.states[batteryID]
	return state, found
}

// List returns all battery states in the cache
func (c *BatteryStateCache) List() []BatteryState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	states := make([]BatteryState, 0, len(c.states))
	for _, state := range c.states {
		states = append(states, state)
	}
	return states
}
