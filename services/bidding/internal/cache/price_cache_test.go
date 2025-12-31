package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/services/bidding/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestPriceUpdate_NewPrice(t *testing.T) {
	c := cache.NewPriceCache()

	now := time.Now()
	c.Update(30.0, now)

	price, timestamp, found := c.GetLatest()
	assert.True(t, found)
	assert.Equal(t, 30.0, price)
	assert.Equal(t, now.Unix(), timestamp.Unix())
}

func TestPriceUpdate_OverwritesOldPrice(t *testing.T) {
	c := cache.NewPriceCache()

	time1 := time.Now()
	c.Update(30.0, time1)

	time2 := time.Now().Add(1 * time.Minute)
	c.Update(150.0, time2)

	price, timestamp, found := c.GetLatest()
	assert.True(t, found)
	assert.Equal(t, 150.0, price)
	assert.Equal(t, time2.Unix(), timestamp.Unix())
}

func TestPriceGetLatest_Exists(t *testing.T) {
	c := cache.NewPriceCache()

	now := time.Now()
	c.Update(75.0, now)

	price, _, found := c.GetLatest()
	assert.True(t, found)
	assert.Equal(t, 75.0, price)
}

func TestPriceGetLatest_NotFound(t *testing.T) {
	c := cache.NewPriceCache()

	_, _, found := c.GetLatest()
	assert.False(t, found)
}

func TestPriceUpdate_Concurrent(t *testing.T) {
	c := cache.NewPriceCache()
	const numGoroutines = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			price := float64(id)
			timestamp := time.Now()
			c.Update(price, timestamp)
		}(i)
	}

	wg.Wait()

	// Should have some price (last write wins)
	_, _, found := c.GetLatest()
	assert.True(t, found)
}

func TestPriceGetLatest_Concurrent(t *testing.T) {
	c := cache.NewPriceCache()

	// Initialize with a price
	c.Update(50.0, time.Now())

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			price, _, found := c.GetLatest()
			assert.True(t, found)
			assert.Equal(t, 50.0, price)
		}()
	}

	wg.Wait()
}

func TestPriceConcurrentReadWrite(t *testing.T) {
	c := cache.NewPriceCache()
	const duration = 100 * time.Millisecond

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		start := time.Now()
		i := 0
		for time.Since(start) < duration {
			price := float64(i % 200)
			c.Update(price, time.Now())
			i++
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		start := time.Now()
		for time.Since(start) < duration {
			c.GetLatest()
		}
	}()

	wg.Wait()
}

func TestPriceZeroValue(t *testing.T) {
	c := cache.NewPriceCache()

	// Update with zero price (valid price)
	now := time.Now()
	c.Update(0.0, now)

	price, timestamp, found := c.GetLatest()
	assert.True(t, found)
	assert.Equal(t, 0.0, price)
	assert.Equal(t, now.Unix(), timestamp.Unix())
}
