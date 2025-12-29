package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
	"github.com/minwook/battery-optimization/services/market-data/internal/ports"
)

// MarketPriceHandler handles HTTP requests for market prices
type MarketPriceHandler struct {
	repo ports.MarketPriceRepository
}

// NewMarketPriceHandler creates a new market price handler
func NewMarketPriceHandler(repo ports.MarketPriceRepository) *MarketPriceHandler {
	return &MarketPriceHandler{repo: repo}
}

// CreateMarketPrice handles POST /prices
func (h *MarketPriceHandler) CreateMarketPrice(w http.ResponseWriter, r *http.Request) {
	var req CreateMarketPriceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON",
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
			respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR",
				err.Error(), nil)
			return
		}
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
			err.Error(), nil)
		return
	}

	if err := h.repo.Create(r.Context(), price); err != nil {
		if errors.Is(err, domain.ErrDuplicateInterval) {
			respondWithError(w, http.StatusConflict, "DUPLICATE_INTERVAL",
				err.Error(), nil)
			return
		}
		log.Printf("Failed to create market price: %v", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to save market price", nil)
		return
	}

	respondWithJSON(w, http.StatusCreated, FromDomain(price))
}

// GetMarketPrice handles GET /prices/:id
func (h *MarketPriceHandler) GetMarketPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	price, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND",
				"Market price not found", nil)
			return
		}
		log.Printf("Failed to find market price: %v", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to retrieve market price", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, FromDomain(price))
}

// ListMarketPrices handles GET /prices with time-range query
func (h *MarketPriceHandler) ListMarketPrices(w http.ResponseWriter, r *http.Request) {
	// Parse required time range
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_PARAMETERS",
			"Required parameters: from (ISO 8601), to (ISO 8601)", nil)
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_DATE_FORMAT",
			"Parameter 'from' must be ISO 8601 format", nil)
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_DATE_FORMAT",
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
		log.Printf("Failed to list market prices: %v", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to list market prices", nil)
		return
	}

	responses := make([]*MarketPriceResponse, len(prices))
	for i, p := range prices {
		responses[i] = FromDomain(p)
	}

	respondWithJSON(w, http.StatusOK, ListMarketPricesResponse{
		Prices: responses,
		Total:  len(responses),
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":"INTERNAL_ERROR","message":"Failed to encode response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, code, message string, details map[string]string) {
	respondWithJSON(w, statusCode, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
