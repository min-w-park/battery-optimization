package cache

import (
	"sync"
	"time"
)

// PriceCache provides thread-safe caching of the latest market price
type PriceCache struct {
	mu          sync.RWMutex
	latestPrice float64
	latestTime  time.Time
	initialized bool
}

// NewPriceCache creates a new price cache
func NewPriceCache() *PriceCache {
	return &PriceCache{
		initialized: false,
	}
}

// Update updates the latest price
func (c *PriceCache) Update(price float64, timestamp time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latestPrice = price
	c.latestTime = timestamp
	c.initialized = true
}

// GetLatest retrieves the latest price
// Returns the price, timestamp, and true if a price exists
// Returns zero values and false if no price has been set
func (c *PriceCache) GetLatest() (float64, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latestPrice, c.latestTime, c.initialized
}
