package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
	"github.com/minwook/battery-optimization/services/market-data/internal/ports"
)

// MarketPriceHandler handles HTTP requests for market prices
type MarketPriceHandler struct {
	repo      ports.MarketPriceRepository
	publisher events.EventPublisher // Optional - service works without events
	log       *zap.Logger
}

// NewMarketPriceHandler creates a new market price handler
func NewMarketPriceHandler(repo ports.MarketPriceRepository, publisher events.EventPublisher, log *zap.Logger) *MarketPriceHandler {
	return &MarketPriceHandler{
		repo:      repo,
		publisher: publisher,
		log:       log,
	}
}

// CreateMarketPrice handles POST /prices
func (h *MarketPriceHandler) CreateMarketPrice(w http.ResponseWriter, r *http.Request) {
	var req CreateMarketPriceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_JSON",
			"Invalid JSON format", nil)
		return
	}

	price, err := req.ToDomain()
	if err != nil {
		// Map domain validation errors to 400 Bad Request
		if errors.Is(err, domain.ErrInvalidPrice) ||
			errors.Is(err, domain.ErrInvalidDemand) ||
			errors.Is(err, domain.ErrInvalidRegion) ||
			errors.Is(err, domain.ErrInvalidIntervalType) ||
			errors.Is(err, domain.ErrIntervalInPast) ||
			errors.Is(err, domain.ErrPublishedInFuture) ||
			errors.Is(err, domain.ErrIntervalAlignment) {
			h.respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR",
				err.Error(), nil)
			return
		}
		h.respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			err.Error(), nil)
		return
	}

	if err := h.repo.Create(r.Context(), price); err != nil {
		if errors.Is(err, domain.ErrDuplicateInterval) {
			h.respondWithError(w, http.StatusConflict, "DUPLICATE_INTERVAL",
				err.Error(), nil)
			return
		}
		h.log.Error("failed to create market price", zap.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to save market price", nil)
		return
	}

	// Publish MarketPriceUpdated event (best-effort, don't fail HTTP request)
	if h.publisher != nil {
		event := events.MarketPriceUpdated{
			PriceID:       price.ID,
			Region:        price.Region,
			Price:         price.Price,
			Demand:        price.Demand,
			IntervalType:  string(price.IntervalType),
			IntervalStart: price.IntervalStart,
			PublishedAt:   price.PublishedAt,
			Timestamp:     time.Now(),
			EventVersion:  "v1",
		}

		// Use context with timeout for event publishing
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := h.publisher.Publish(ctx, "market.price.updated.v1", event); err != nil {
			// Log warning but don't fail the HTTP request
			h.log.Warn("failed to publish market.price.updated.v1 event",
				zap.String("price_id", price.ID),
				zap.Error(err),
			)
		} else {
			h.log.Info("published market.price.updated.v1 event",
				zap.String("price_id", price.ID),
				zap.String("region", price.Region),
				zap.Float64("price", price.Price),
			)
		}
	}

	h.respondWithJSON(w, http.StatusCreated, FromDomain(price))
}

// GetMarketPrice handles GET /prices/:id
func (h *MarketPriceHandler) GetMarketPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	price, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "NOT_FOUND",
				"Market price not found", nil)
			return
		}
		h.log.Error("failed to find market price",
			zap.String("price_id", id),
			zap.Error(err),
		)
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to retrieve market price", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, FromDomain(price))
}

// ListMarketPrices handles GET /prices with time-range query
func (h *MarketPriceHandler) ListMarketPrices(w http.ResponseWriter, r *http.Request) {
	// Parse required time range
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "MISSING_PARAMETERS",
			"Required parameters: from (ISO 8601), to (ISO 8601)", nil)
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_DATE_FORMAT",
			"Parameter 'from' must be ISO 8601 format", nil)
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_DATE_FORMAT",
			"Parameter 'to' must be ISO 8601 format", nil)
		return
	}

	filter := ports.TimeRangeFilter{
		From:         from,
		To:           to,
		Region:       r.URL.Query().Get("region"),
		IntervalType: domain.IntervalType(r.URL.Query().Get("interval_type")),
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		if limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, _ := strconv.Atoi(offsetStr)
		if offset >= 0 {
			filter.Offset = offset
		}
	}

	prices, err := h.repo.ListByTimeRange(r.Context(), filter)
	if err != nil {
		h.log.Error("failed to list market prices", zap.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to list market prices", nil)
		return
	}

	responses := make([]*MarketPriceResponse, len(prices))
	for i, p := range prices {
		responses[i] = FromDomain(p)
	}

	h.respondWithJSON(w, http.StatusOK, ListMarketPricesResponse{
		Prices: responses,
		Total:  len(responses),
	})
}

// respondWithJSON sends a JSON response
func (h *MarketPriceHandler) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.log.Error("failed to encode JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":"INTERNAL_ERROR","message":"Failed to encode response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

// respondWithError sends an error response
func (h *MarketPriceHandler) respondWithError(w http.ResponseWriter, statusCode int, code, message string, details map[string]string) {
	h.respondWithJSON(w, statusCode, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
