package domain

import (
	"time"

	"github.com/google/uuid"
)

// IntervalType defines AEMO market data interval types
type IntervalType string

const (
	Interval5Min  IntervalType = "5MIN_PREDISPATCH"  // 5-minute dispatch intervals
	Interval30Min IntervalType = "30MIN_PREDISPATCH" // 30-minute trading intervals
)

// MarketPrice represents a price forecast for the Australian NEM
type MarketPrice struct {
	ID            string       // UUID
	Region        string       // NSW, VIC, QLD, SA, TAS
	Price         float64      // $/MWh (Australian dollars per megawatt-hour)
	Demand        float64      // MW (system-wide demand forecast)
	IntervalType  IntervalType // 5MIN_PREDISPATCH or 30MIN_PREDISPATCH
	IntervalStart time.Time    // ISO 8601 timestamp
	PublishedAt   time.Time    // When AEMO published this forecast
	CreatedAt     time.Time    // When we received it
}

// NewMarketPrice creates a new MarketPrice with validation
func NewMarketPrice(
	region string,
	price float64,
	demand float64,
	intervalType IntervalType,
	intervalStart time.Time,
	publishedAt time.Time,
) (*MarketPrice, error) {
	// Generate ID (UUID)
	id := uuid.New().String()

	// Create market price
	marketPrice := &MarketPrice{
		ID:            id,
		Region:        region,
		Price:         price,
		Demand:        demand,
		IntervalType:  intervalType,
		IntervalStart: intervalStart,
		PublishedAt:   publishedAt,
		CreatedAt:     time.Now(),
	}

	// Validate
	if err := marketPrice.Validate(); err != nil {
		return nil, err
	}

	return marketPrice, nil
}

// Validate checks all business rules
func (mp *MarketPrice) Validate() error {
	// Price validation
	if mp.Price < 0 {
		return ErrInvalidPrice
	}

	// Demand validation
	if mp.Demand <= 0 {
		return ErrInvalidDemand
	}

	// Region validation
	if !isValidRegion(mp.Region) {
		return ErrInvalidRegion
	}

	// IntervalType validation
	if !isValidIntervalType(mp.IntervalType) {
		return ErrInvalidIntervalType
	}

	// Time constraints
	if mp.IntervalStart.Before(time.Now()) {
		return ErrIntervalInPast
	}

	if mp.PublishedAt.After(time.Now()) {
		return ErrPublishedInFuture
	}

	// Interval alignment validation
	if !validateIntervalAlignment(mp.IntervalStart, mp.IntervalType) {
		return ErrIntervalAlignment
	}

	return nil
}

// isValidRegion checks if region is a valid NEM region
func isValidRegion(region string) bool {
	validRegions := map[string]bool{
		"NSW": true,
		"VIC": true,
		"QLD": true,
		"SA":  true,
		"TAS": true,
	}
	return validRegions[region]
}

// isValidIntervalType checks if interval type is valid
func isValidIntervalType(intervalType IntervalType) bool {
	return intervalType == Interval5Min || intervalType == Interval30Min
}

// validateIntervalAlignment checks if timestamp aligns with interval type
func validateIntervalAlignment(t time.Time, intervalType IntervalType) bool {
	minute := t.Minute()
	second := t.Second()

	if intervalType == Interval5Min {
		// Must be on 5-minute boundary (0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55)
		return minute%5 == 0 && second == 0
	}

	if intervalType == Interval30Min {
		// Must be on 30-minute boundary (0 or 30)
		return (minute == 0 || minute == 30) && second == 0
	}

	return false
}
