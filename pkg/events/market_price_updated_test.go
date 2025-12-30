package events_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketPriceUpdated_JSONSerialization(t *testing.T) {
	// Create event with all fields
	intervalStart := time.Date(2025, 12, 30, 11, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2025, 12, 30, 10, 55, 0, 0, time.UTC)
	timestamp := time.Date(2025, 12, 30, 10, 55, 5, 0, time.UTC)

	event := events.MarketPriceUpdated{
		PriceID:       "price-123",
		Region:        "NSW",
		Price:         85.50,
		Demand:        8200.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: intervalStart,
		PublishedAt:   publishedAt,
		Timestamp:     timestamp,
		EventVersion:  "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal back
	var decoded events.MarketPriceUpdated
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, "price-123", decoded.PriceID)
	assert.Equal(t, "NSW", decoded.Region)
	assert.Equal(t, 85.50, decoded.Price)
	assert.Equal(t, 8200.0, decoded.Demand)
	assert.Equal(t, "5MIN_PREDISPATCH", decoded.IntervalType)
	assert.Equal(t, intervalStart.Unix(), decoded.IntervalStart.Unix())
	assert.Equal(t, publishedAt.Unix(), decoded.PublishedAt.Unix())
	assert.Equal(t, timestamp.Unix(), decoded.Timestamp.Unix())
	assert.Equal(t, "v1", decoded.EventVersion)
}

func TestMarketPriceUpdated_Timestamps(t *testing.T) {
	// Test all three timestamps are preserved correctly
	now := time.Now()
	intervalFuture := now.Add(30 * time.Minute)
	publishedPast := now.Add(-5 * time.Minute)

	event := events.MarketPriceUpdated{
		PriceID:       "test-id",
		Region:        "VIC",
		Price:         100.0,
		Demand:        5000.0,
		IntervalType:  "30MIN_PREDISPATCH",
		IntervalStart: intervalFuture,
		PublishedAt:   publishedPast,
		Timestamp:     now,
		EventVersion:  "v1",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded events.MarketPriceUpdated
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// All timestamps should be preserved
	assert.Equal(t, intervalFuture.Unix(), decoded.IntervalStart.Unix())
	assert.Equal(t, publishedPast.Unix(), decoded.PublishedAt.Unix())
	assert.Equal(t, now.Unix(), decoded.Timestamp.Unix())
}

func TestMarketPriceUpdated_JSONTags(t *testing.T) {
	event := events.MarketPriceUpdated{
		PriceID:       "test-id",
		Region:        "SA",
		Price:         120.0,
		Demand:        1500.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: time.Now(),
		PublishedAt:   time.Now(),
		Timestamp:     time.Now(),
		EventVersion:  "v1",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	// Verify JSON field names (snake_case)
	assert.Contains(t, jsonMap, "price_id")
	assert.Contains(t, jsonMap, "region")
	assert.Contains(t, jsonMap, "price")
	assert.Contains(t, jsonMap, "demand")
	assert.Contains(t, jsonMap, "interval_type")
	assert.Contains(t, jsonMap, "interval_start")
	assert.Contains(t, jsonMap, "published_at")
	assert.Contains(t, jsonMap, "timestamp")
	assert.Contains(t, jsonMap, "event_version")
}

func TestMarketPriceUpdated_DifferentRegions(t *testing.T) {
	regions := []string{"NSW", "VIC", "QLD", "SA", "TAS"}

	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			event := events.MarketPriceUpdated{
				PriceID:       "test-" + region,
				Region:        region,
				Price:         90.0,
				Demand:        7000.0,
				IntervalType:  "5MIN_PREDISPATCH",
				IntervalStart: time.Now(),
				PublishedAt:   time.Now(),
				Timestamp:     time.Now(),
				EventVersion:  "v1",
			}

			data, err := json.Marshal(event)
			require.NoError(t, err)

			var decoded events.MarketPriceUpdated
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, region, decoded.Region)
		})
	}
}
