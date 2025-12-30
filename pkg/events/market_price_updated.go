package events

import "time"

// MarketPriceUpdated is published when new market price data is received.
// Published by: Market Data Service
// Subject: market.price.updated.v1
// Frequency: Every 5-30 minutes (depending on AEMO interval type)
type MarketPriceUpdated struct {
	// Price identification
	PriceID string `json:"price_id"`

	// Market data
	Region        string    `json:"region"`          // NSW, VIC, QLD, SA, TAS
	Price         float64   `json:"price"`           // $/MWh
	Demand        float64   `json:"demand"`          // MW
	IntervalType  string    `json:"interval_type"`   // 5MIN_PREDISPATCH, 30MIN_PREDISPATCH
	IntervalStart time.Time `json:"interval_start"`  // When this price applies
	PublishedAt   time.Time `json:"published_at"`    // When AEMO published it

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
