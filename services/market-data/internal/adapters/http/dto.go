package http

import (
	"fmt"
	"time"

	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
)

// CreateMarketPriceRequest represents the request to create a market price
type CreateMarketPriceRequest struct {
	Region        string  `json:"region"`
	Price         float64 `json:"price"`
	Demand        float64 `json:"demand"`
	IntervalType  string  `json:"interval_type"`
	IntervalStart string  `json:"interval_start"` // ISO 8601
	PublishedAt   string  `json:"published_at"`   // ISO 8601
}

// ToDomain converts CreateMarketPriceRequest to domain.MarketPrice
func (r *CreateMarketPriceRequest) ToDomain() (*domain.MarketPrice, error) {
	// Parse ISO 8601 timestamps
	intervalStart, err := time.Parse(time.RFC3339, r.IntervalStart)
	if err != nil {
		return nil, fmt.Errorf("invalid interval_start format: %w", err)
	}

	publishedAt, err := time.Parse(time.RFC3339, r.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid published_at format: %w", err)
	}

	// Create domain entity (validation happens in NewMarketPrice)
	return domain.NewMarketPrice(
		r.Region,
		r.Price,
		r.Demand,
		domain.IntervalType(r.IntervalType),
		intervalStart,
		publishedAt,
	)
}

// MarketPriceResponse represents a market price in API responses
type MarketPriceResponse struct {
	ID            string  `json:"id"`
	Region        string  `json:"region"`
	Price         float64 `json:"price"`
	Demand        float64 `json:"demand"`
	IntervalType  string  `json:"interval_type"`
	IntervalStart string  `json:"interval_start"` // ISO 8601
	PublishedAt   string  `json:"published_at"`   // ISO 8601
	CreatedAt     string  `json:"created_at"`     // ISO 8601
}

// FromDomain converts domain.MarketPrice to MarketPriceResponse
func FromDomain(price *domain.MarketPrice) *MarketPriceResponse {
	return &MarketPriceResponse{
		ID:            price.ID,
		Region:        price.Region,
		Price:         price.Price,
		Demand:        price.Demand,
		IntervalType:  string(price.IntervalType),
		IntervalStart: price.IntervalStart.Format(time.RFC3339),
		PublishedAt:   price.PublishedAt.Format(time.RFC3339),
		CreatedAt:     price.CreatedAt.Format(time.RFC3339),
	}
}

// ListMarketPricesResponse represents paginated market price list
type ListMarketPricesResponse struct {
	Prices []*MarketPriceResponse `json:"prices"`
	Total  int                    `json:"total"`
}

// ErrorResponse represents an error
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
