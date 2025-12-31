package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
	"github.com/minwook/battery-optimization/services/asset-management/internal/ports"
)

// MockRepository is a mock implementation of BatteryRepository for testing
type MockRepository struct {
	CreateFunc   func(ctx context.Context, battery *domain.Battery) error
	FindByIDFunc func(ctx context.Context, id string) (*domain.Battery, error)
	ListFunc     func(ctx context.Context, filter ports.ListFilter) ([]*domain.Battery, error)
}

func (m *MockRepository) Create(ctx context.Context, battery *domain.Battery) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, battery)
	}
	return nil
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*domain.Battery, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context, filter ports.ListFilter) ([]*domain.Battery, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return nil, nil
}

// Helper function to create a valid test battery
func createTestBattery() *domain.Battery {
	constraints := domain.Constraints{
		WarrantyEOL:         0.8,
		MaxCycles:           10000,
		OperatingTempMin:    -10.0,
		OperatingTempMax:    50.0,
		GridComplianceLevel: "AS4777",
	}

	battery, _ := domain.NewBattery(200.0, 100.0, 50.0, 0.85, "SA", "Tesla", constraints)
	return battery
}

func TestCreateBattery_Success(t *testing.T) {
	// Given: A mock repository and handler
	mockRepo := &MockRepository{
		CreateFunc: func(ctx context.Context, battery *domain.Battery) error {
			return nil
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	requestBody := CreateBatteryRequest{
		Capacity:     200.0,
		MaxPower:     100.0,
		RampRate:     50.0,
		Efficiency:   0.85,
		Location:     "SA",
		Manufacturer: "Tesla",
		Constraints: ConstraintsRequest{
			WarrantyEOL:         0.8,
			MaxCycles:           10000,
			OperatingTempMin:    -10.0,
			OperatingTempMax:    50.0,
			GridComplianceLevel: "AS4777",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batteries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When: Creating a battery
	handler.CreateBattery(rec, req)

	// Then: Should return 201 Created
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response BatteryResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, 200.0, response.Capacity)
	assert.Equal(t, "SA", response.Location)
	assert.Equal(t, "REGISTERED", response.Status)
}

func TestCreateBattery_InvalidJSON(t *testing.T) {
	// Given: A handler
	mockRepo := &MockRepository{}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/batteries", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When: Posting invalid JSON
	handler.CreateBattery(rec, req)

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_JSON", response.Code)
}

func TestCreateBattery_ValidationFailure(t *testing.T) {
	// Given: A handler
	mockRepo := &MockRepository{}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	requestBody := CreateBatteryRequest{
		Capacity:     0, // Invalid: must be > 0
		MaxPower:     100.0,
		RampRate:     50.0,
		Efficiency:   0.85,
		Location:     "SA",
		Manufacturer: "Tesla",
		Constraints: ConstraintsRequest{
			WarrantyEOL:      0.8,
			MaxCycles:        10000,
			OperatingTempMin: -10.0,
			OperatingTempMax: 50.0,
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batteries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When: Creating battery with invalid capacity
	handler.CreateBattery(rec, req)

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Code)
	assert.Contains(t, response.Message, "capacity")
}

func TestCreateBattery_RepositoryError(t *testing.T) {
	// Given: A repository that returns an error
	mockRepo := &MockRepository{
		CreateFunc: func(ctx context.Context, battery *domain.Battery) error {
			return errors.New("database error")
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	requestBody := CreateBatteryRequest{
		Capacity:     200.0,
		MaxPower:     100.0,
		RampRate:     50.0,
		Efficiency:   0.85,
		Location:     "SA",
		Manufacturer: "Tesla",
		Constraints: ConstraintsRequest{
			WarrantyEOL:      0.8,
			MaxCycles:        10000,
			OperatingTempMin: -10.0,
			OperatingTempMax: 50.0,
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batteries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// When: Repository fails
	handler.CreateBattery(rec, req)

	// Then: Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetBattery_Success(t *testing.T) {
	// Given: A repository with an existing battery
	testBattery := createTestBattery()
	mockRepo := &MockRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.Battery, error) {
			if id == testBattery.ID {
				return testBattery, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/batteries/"+testBattery.ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": testBattery.ID})
	rec := httptest.NewRecorder()

	// When: Getting battery by ID
	handler.GetBattery(rec, req)

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatteryResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, testBattery.ID, response.ID)
	assert.Equal(t, testBattery.Capacity, response.Capacity)
}

func TestGetBattery_NotFound(t *testing.T) {
	// Given: A repository without the battery
	mockRepo := &MockRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*domain.Battery, error) {
			return nil, domain.ErrNotFound
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/batteries/non-existent", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "non-existent"})
	rec := httptest.NewRecorder()

	// When: Getting non-existent battery
	handler.GetBattery(rec, req)

	// Then: Should return 404 Not Found
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

func TestListBatteries_NoFilters(t *testing.T) {
	// Given: A repository with batteries
	battery1 := createTestBattery()
	battery2 := createTestBattery()

	mockRepo := &MockRepository{
		ListFunc: func(ctx context.Context, filter ports.ListFilter) ([]*domain.Battery, error) {
			return []*domain.Battery{battery1, battery2}, nil
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/batteries", nil)
	rec := httptest.NewRecorder()

	// When: Listing all batteries
	handler.ListBatteries(rec, req)

	// Then: Should return 200 OK with list
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ListBatteriesResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Batteries, 2)
	assert.Equal(t, 2, response.Total)
}

func TestListBatteries_WithFilters(t *testing.T) {
	// Given: A repository that respects filters
	battery := createTestBattery()

	mockRepo := &MockRepository{
		ListFunc: func(ctx context.Context, filter ports.ListFilter) ([]*domain.Battery, error) {
			// Verify filter parameters
			assert.Equal(t, "SA", filter.Location)
			assert.Equal(t, 10, filter.Limit)
			assert.Equal(t, 5, filter.Offset)
			return []*domain.Battery{battery}, nil
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/batteries?location=SA&limit=10&offset=5", nil)
	rec := httptest.NewRecorder()

	// When: Listing with filters
	handler.ListBatteries(rec, req)

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListBatteries_RepositoryError(t *testing.T) {
	// Given: A repository that returns an error
	mockRepo := &MockRepository{
		ListFunc: func(ctx context.Context, filter ports.ListFilter) ([]*domain.Battery, error) {
			return nil, errors.New("database error")
		},
	}
	handler := NewBatteryHandler(mockRepo, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/batteries", nil)
	rec := httptest.NewRecorder()

	// When: Repository fails
	handler.ListBatteries(rec, req)

	// Then: Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
