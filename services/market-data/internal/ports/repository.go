package ports

import (
	"context"
	"time"

	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
)

// MarketPriceRepository defines the interface for market price persistence
type MarketPriceRepository interface {
	Create(ctx context.Context, price *domain.MarketPrice) error
	FindByID(ctx context.Context, id string) (*domain.MarketPrice, error)
	ListByTimeRange(ctx context.Context, filter TimeRangeFilter) ([]*domain.MarketPrice, error)
}

// TimeRangeFilter defines filters for time-range queries
type TimeRangeFilter struct {
	Region       string              // Optional: filter by region
	IntervalType domain.IntervalType // Optional: filter by interval type
	From         time.Time           // Required: start of time range
	To           time.Time           // Required: end of time range
	Limit        int                 // Optional: max results (default: 1000)
	Offset       int                 // Optional: pagination offset
}
