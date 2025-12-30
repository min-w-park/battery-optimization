package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
	"github.com/minwook/battery-optimization/services/market-data/internal/ports"
)

// MockRepository is a manual mock for testing
type MockRepository struct {
	CreateFunc          func(ctx context.Context, price *domain.MarketPrice) error
	FindByIDFunc        func(ctx context.Context, id string) (*domain.MarketPrice, error)
	ListByTimeRangeFunc func(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error)
}

func (m *MockRepository) Create(ctx context.Context, price *domain.MarketPrice) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, price)
	}
	return nil
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*domain.MarketPrice, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) ListByTimeRange(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
	if m.ListByTimeRangeFunc != nil {
		return m.ListByTimeRangeFunc(ctx, filter)
	}
	return nil, nil
}

func TestCreateMarketPrice_Success(t *testing.T) {
	// Given
	mockRepo := &MockRepository{
		CreateFunc: func(ctx context.Context, price *domain.MarketPrice) error {
			return nil
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	reqBody := CreateMarketPriceRequest{
		Region:        "NSW",
		Price:         85.50,
		Demand:        8200.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: "2025-12-30T10:00:00Z",
		PublishedAt:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When
	handler.CreateMarketPrice(rec, req)

	// Then
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp MarketPriceResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, 85.50, resp.Price)
	assert.Equal(t, "NSW", resp.Region)
}

func TestCreateMarketPrice_InvalidJSON(t *testing.T) {
	// Given
	mockRepo := &MockRepository{}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/prices", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When
	handler.CreateMarketPrice(rec, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_JSON", resp.Code)
}

func TestCreateMarketPrice_ValidationFailure(t *testing.T) {
	// Given
	mockRepo := &MockRepository{}
	handler := NewMarketPriceHandler(mockRepo, nil)

	reqBody := CreateMarketPriceRequest{
		Region:        "NSW",
		Price:         -10.0, // Invalid: must be >= 0
		Demand:        8200.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: "2025-12-30T10:00:00Z",
		PublishedAt:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When
	handler.CreateMarketPrice(rec, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Contains(t, resp.Message, "price")
}

func TestCreateMarketPrice_DuplicateInterval(t *testing.T) {
	// Given
	mockRepo := &MockRepository{
		CreateFunc: func(ctx context.Context, price *domain.MarketPrice) error {
			return domain.ErrDuplicateInterval
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	reqBody := CreateMarketPriceRequest{
		Region:        "NSW",
		Price:         85.50,
		Demand:        8200.0,
		IntervalType:  "5MIN_PREDISPATCH",
		IntervalStart: "2025-12-30T10:00:00Z",
		PublishedAt:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/prices", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When
	handler.CreateMarketPrice(rec, req)

	// Then
	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "DUPLICATE_INTERVAL", resp.Code)
}

func TestGetMarketPrice_Success(t *testing.T) {
	// Given
	intervalStart := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2025, 12, 29, 9, 55, 0, 0, time.UTC)

	testPrice, _ := domain.NewMarketPrice("NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart, publishedAt)

	mockRepo := &MockRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.MarketPrice, error) {
			if id == testPrice.ID {
				return testPrice, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prices/"+testPrice.ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": testPrice.ID})
	rec := httptest.NewRecorder()

	// When
	handler.GetMarketPrice(rec, req)

	// Then
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp MarketPriceResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, testPrice.ID, resp.ID)
	assert.Equal(t, testPrice.Price, resp.Price)
}

func TestGetMarketPrice_NotFound(t *testing.T) {
	// Given
	mockRepo := &MockRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.MarketPrice, error) {
			return nil, domain.ErrNotFound
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prices/non-existent", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "non-existent"})
	rec := httptest.NewRecorder()

	// When
	handler.GetMarketPrice(rec, req)

	// Then
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

func TestListMarketPrices_WithTimeRange(t *testing.T) {
	// Given
	intervalStart := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2025, 12, 29, 9, 55, 0, 0, time.UTC)

	price1, _ := domain.NewMarketPrice("NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart, publishedAt)

	mockRepo := &MockRepository{
		ListByTimeRangeFunc: func(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
			assert.Equal(t, "2025-12-30T00:00:00Z", filter.From.Format(time.RFC3339))
			assert.Equal(t, "2025-12-31T00:00:00Z", filter.To.Format(time.RFC3339))
			return []*domain.MarketPrice{price1}, nil
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When
	handler.ListMarketPrices(rec, req)

	// Then
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp ListMarketPricesResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Prices, 1)
	assert.Equal(t, 1, resp.Total)
}

func TestListMarketPrices_WithFilters(t *testing.T) {
	// Given
	intervalStart := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2025, 12, 29, 9, 55, 0, 0, time.UTC)

	price1, _ := domain.NewMarketPrice("NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart, publishedAt)

	mockRepo := &MockRepository{
		ListByTimeRangeFunc: func(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
			// Verify filter parameters
			assert.Equal(t, "NSW", filter.Region)
			assert.Equal(t, domain.Interval5Min, filter.IntervalType)
			assert.Equal(t, 10, filter.Limit)
			assert.Equal(t, 5, filter.Offset)
			return []*domain.MarketPrice{price1}, nil
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&region=NSW&interval_type=5MIN_PREDISPATCH&limit=10&offset=5", nil)
	rec := httptest.NewRecorder()

	// When
	handler.ListMarketPrices(rec, req)

	// Then
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListMarketPrices_MissingFromParameter(t *testing.T) {
	// Given
	mockRepo := &MockRepository{}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prices?to=2025-12-31T00:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When
	handler.ListMarketPrices(rec, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "MISSING_PARAMETERS", resp.Code)
}

func TestListMarketPrices_InvalidDateFormat(t *testing.T) {
	// Given
	mockRepo := &MockRepository{}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/prices?from=invalid-date&to=2025-12-31T00:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When
	handler.ListMarketPrices(rec, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_DATE_FORMAT", resp.Code)
}

func TestListMarketPrices_RepositoryError(t *testing.T) {
	// Given
	mockRepo := &MockRepository{
		ListByTimeRangeFunc: func(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
			return nil, assert.AnError
		},
	}
	handler := NewMarketPriceHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When
	handler.ListMarketPrices(rec, req)

	// Then
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
