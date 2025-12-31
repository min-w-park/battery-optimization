package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMarketPrice_ValidInput(t *testing.T) {
	tests := []struct {
		name          string
		region        string
		price         float64
		demand        float64
		intervalType  IntervalType
		intervalStart time.Time
		publishedAt   time.Time
	}{
		{
			name:          "valid 5-minute forecast",
			region:        "NSW",
			price:         85.50,
			demand:        8200.0,
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			publishedAt:   time.Now().Add(-5 * time.Minute),
		},
		{
			name:          "valid 30-minute forecast",
			region:        "SA",
			price:         120.00,
			demand:        1500.0,
			intervalType:  Interval30Min,
			intervalStart: time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
			publishedAt:   time.Now().Add(-10 * time.Minute),
		},
		{
			name:          "valid VIC region with 5-minute",
			region:        "VIC",
			price:         95.75,
			demand:        7500.0,
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 5, 0, 0, time.UTC),
			publishedAt:   time.Now().Add(-3 * time.Minute),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			price, err := NewMarketPrice(
				tt.region, tt.price, tt.demand,
				tt.intervalType, tt.intervalStart, tt.publishedAt,
			)

			// Then
			require.NoError(t, err)
			assert.NotNil(t, price)
			assert.NotEmpty(t, price.ID)
			assert.Equal(t, tt.price, price.Price)
			assert.Equal(t, tt.region, price.Region)
			assert.Equal(t, tt.demand, price.Demand)
			assert.Equal(t, tt.intervalType, price.IntervalType)
			assert.Equal(t, tt.intervalStart, price.IntervalStart)
			assert.Equal(t, tt.publishedAt, price.PublishedAt)
			assert.False(t, price.CreatedAt.IsZero())
		})
	}
}

func TestNewMarketPrice_InvalidPrice(t *testing.T) {
	tests := []struct {
		name  string
		price float64
	}{
		{"negative price", -10.0},
		{"very negative price", -100.0},
		{"slightly negative", -0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			intervalStart := time.Now().Add(1 * time.Hour)
			publishedAt := time.Now()

			// When
			price, err := NewMarketPrice(
				"NSW", tt.price, 8000.0,
				Interval5Min, intervalStart, publishedAt,
			)

			// Then
			assert.Error(t, err)
			assert.Nil(t, price)
			assert.ErrorIs(t, err, ErrInvalidPrice)
		})
	}
}

func TestNewMarketPrice_InvalidDemand(t *testing.T) {
	tests := []struct {
		name   string
		demand float64
	}{
		{"zero demand", 0},
		{"negative demand", -100.0},
		{"slightly negative demand", -0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			intervalStart := time.Now().Add(1 * time.Hour)
			publishedAt := time.Now()

			// When
			price, err := NewMarketPrice(
				"NSW", 85.50, tt.demand,
				Interval5Min, intervalStart, publishedAt,
			)

			// Then
			assert.Error(t, err)
			assert.Nil(t, price)
			assert.ErrorIs(t, err, ErrInvalidDemand)
		})
	}
}

func TestNewMarketPrice_InvalidRegion(t *testing.T) {
	tests := []struct {
		name   string
		region string
	}{
		{"invalid region", "INVALID"},
		{"WA not in NEM", "WA"},
		{"NT not in NEM", "NT"},
		{"empty region", ""},
		{"ACT not in NEM", "ACT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			intervalStart := time.Now().Add(1 * time.Hour)
			publishedAt := time.Now()

			// When
			price, err := NewMarketPrice(
				tt.region, 85.50, 8000.0,
				Interval5Min, intervalStart, publishedAt,
			)

			// Then
			assert.Error(t, err)
			assert.Nil(t, price)
			assert.ErrorIs(t, err, ErrInvalidRegion)
		})
	}
}

func TestNewMarketPrice_InvalidIntervalType(t *testing.T) {
	tests := []struct {
		name         string
		intervalType IntervalType
	}{
		{"invalid type", "INVALID"},
		{"1MIN type", "1MIN_PREDISPATCH"},
		{"empty type", ""},
		{"15MIN type", "15MIN_PREDISPATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			intervalStart := time.Now().Add(1 * time.Hour)
			publishedAt := time.Now()

			// When
			price, err := NewMarketPrice(
				"NSW", 85.50, 8000.0,
				tt.intervalType, intervalStart, publishedAt,
			)

			// Then
			assert.Error(t, err)
			assert.Nil(t, price)
			assert.ErrorIs(t, err, ErrInvalidIntervalType)
		})
	}
}

func TestNewMarketPrice_IntervalInPast(t *testing.T) {
	// Given: interval start is 1 hour ago
	intervalStart := time.Now().Add(-1 * time.Hour)
	publishedAt := time.Now().Add(-2 * time.Hour)

	// When
	price, err := NewMarketPrice(
		"NSW", 85.50, 8000.0,
		Interval5Min, intervalStart, publishedAt,
	)

	// Then
	assert.Error(t, err)
	assert.Nil(t, price)
	assert.ErrorIs(t, err, ErrIntervalInPast)
}

func TestNewMarketPrice_PublishedInFuture(t *testing.T) {
	// Given: published time is 1 hour from now
	intervalStart := time.Now().Add(1 * time.Hour)
	publishedAt := time.Now().Add(1 * time.Hour)

	// When
	price, err := NewMarketPrice(
		"NSW", 85.50, 8000.0,
		Interval5Min, intervalStart, publishedAt,
	)

	// Then
	assert.Error(t, err)
	assert.Nil(t, price)
	assert.ErrorIs(t, err, ErrPublishedInFuture)
}

func TestNewMarketPrice_IntervalAlignment(t *testing.T) {
	tests := []struct {
		name          string
		intervalType  IntervalType
		intervalStart time.Time
		shouldFail    bool
	}{
		{
			name:          "5MIN valid - on 5-minute boundary (10:00)",
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			shouldFail:    false,
		},
		{
			name:          "5MIN valid - on 5-minute boundary (10:05)",
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 5, 0, 0, time.UTC),
			shouldFail:    false,
		},
		{
			name:          "5MIN valid - on 5-minute boundary (10:15)",
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 15, 0, 0, time.UTC),
			shouldFail:    false,
		},
		{
			name:          "5MIN invalid - not on boundary (10:03)",
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 3, 0, 0, time.UTC),
			shouldFail:    true,
		},
		{
			name:          "5MIN invalid - not on boundary (10:01)",
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 1, 0, 0, time.UTC),
			shouldFail:    true,
		},
		{
			name:          "30MIN valid - on 30-minute boundary (10:00)",
			intervalType:  Interval30Min,
			intervalStart: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			shouldFail:    false,
		},
		{
			name:          "30MIN valid - on 30-minute boundary (10:30)",
			intervalType:  Interval30Min,
			intervalStart: time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
			shouldFail:    false,
		},
		{
			name:          "30MIN invalid - not on boundary (10:15)",
			intervalType:  Interval30Min,
			intervalStart: time.Date(2026, 1, 15, 10, 15, 0, 0, time.UTC),
			shouldFail:    true,
		},
		{
			name:          "30MIN invalid - not on boundary (10:45)",
			intervalType:  Interval30Min,
			intervalStart: time.Date(2026, 1, 15, 10, 45, 0, 0, time.UTC),
			shouldFail:    true,
		},
		{
			name:          "5MIN invalid - has seconds (10:00:30)",
			intervalType:  Interval5Min,
			intervalStart: time.Date(2026, 1, 15, 10, 0, 30, 0, time.UTC),
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			publishedAt := time.Now()

			// When
			price, err := NewMarketPrice(
				"NSW", 80.0, 8000.0,
				tt.intervalType, tt.intervalStart, publishedAt,
			)

			// Then
			if tt.shouldFail {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrIntervalAlignment)
				assert.Nil(t, price)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, price)
			}
		})
	}
}

func TestValidateIntervalAlignment(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    time.Time
		intervalType IntervalType
		expected     bool
	}{
		// 5-minute valid cases
		{"5MIN - 10:00", time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), Interval5Min, true},
		{"5MIN - 10:05", time.Date(2026, 1, 15, 10, 5, 0, 0, time.UTC), Interval5Min, true},
		{"5MIN - 10:10", time.Date(2026, 1, 15, 10, 10, 0, 0, time.UTC), Interval5Min, true},
		{"5MIN - 10:15", time.Date(2026, 1, 15, 10, 15, 0, 0, time.UTC), Interval5Min, true},
		{"5MIN - 10:20", time.Date(2026, 1, 15, 10, 20, 0, 0, time.UTC), Interval5Min, true},
		{"5MIN - 10:55", time.Date(2026, 1, 15, 10, 55, 0, 0, time.UTC), Interval5Min, true},

		// 5-minute invalid cases
		{"5MIN - 10:01", time.Date(2026, 1, 15, 10, 1, 0, 0, time.UTC), Interval5Min, false},
		{"5MIN - 10:03", time.Date(2026, 1, 15, 10, 3, 0, 0, time.UTC), Interval5Min, false},
		{"5MIN - 10:07", time.Date(2026, 1, 15, 10, 7, 0, 0, time.UTC), Interval5Min, false},
		{"5MIN - 10:00:01", time.Date(2026, 1, 15, 10, 0, 1, 0, time.UTC), Interval5Min, false},

		// 30-minute valid cases
		{"30MIN - 10:00", time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), Interval30Min, true},
		{"30MIN - 10:30", time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC), Interval30Min, true},
		{"30MIN - 11:00", time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC), Interval30Min, true},

		// 30-minute invalid cases
		{"30MIN - 10:15", time.Date(2026, 1, 15, 10, 15, 0, 0, time.UTC), Interval30Min, false},
		{"30MIN - 10:45", time.Date(2026, 1, 15, 10, 45, 0, 0, time.UTC), Interval30Min, false},
		{"30MIN - 10:05", time.Date(2026, 1, 15, 10, 5, 0, 0, time.UTC), Interval30Min, false},
		{"30MIN - 10:30:01", time.Date(2026, 1, 15, 10, 30, 1, 0, time.UTC), Interval30Min, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateIntervalAlignment(tt.timestamp, tt.intervalType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidRegion(t *testing.T) {
	tests := []struct {
		region   string
		expected bool
	}{
		// Valid NEM regions
		{"NSW", true},
		{"VIC", true},
		{"QLD", true},
		{"SA", true},
		{"TAS", true},

		// Invalid regions
		{"WA", false},
		{"NT", false},
		{"ACT", false},
		{"INVALID", false},
		{"", false},
		{"nsw", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := isValidRegion(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidIntervalType(t *testing.T) {
	tests := []struct {
		intervalType IntervalType
		expected     bool
	}{
		// Valid interval types
		{Interval5Min, true},
		{Interval30Min, true},

		// Invalid interval types
		{"INVALID", false},
		{"1MIN_PREDISPATCH", false},
		{"15MIN_PREDISPATCH", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.intervalType), func(t *testing.T) {
			result := isValidIntervalType(tt.intervalType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarketPrice_ZeroPrice(t *testing.T) {
	// Given: zero price (valid during oversupply) with properly aligned timestamp
	intervalStart := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Now().Add(-5 * time.Minute)

	// When
	price, err := NewMarketPrice(
		"NSW", 0.0, 8000.0,
		Interval5Min, intervalStart, publishedAt,
	)

	// Then: should be valid (price >= 0)
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.Equal(t, 0.0, price.Price)
}

func TestMarketPrice_AllNEMRegions(t *testing.T) {
	regions := []string{"NSW", "VIC", "QLD", "SA", "TAS"}

	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			intervalStart := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
			publishedAt := time.Now().Add(-5 * time.Minute)

			price, err := NewMarketPrice(
				region, 85.50, 8000.0,
				Interval5Min, intervalStart, publishedAt,
			)

			assert.NoError(t, err)
			assert.NotNil(t, price)
			assert.Equal(t, region, price.Region)
		})
	}
}
