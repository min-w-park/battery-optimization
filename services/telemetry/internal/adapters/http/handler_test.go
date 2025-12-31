package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
	"github.com/minwook/battery-optimization/services/telemetry/internal/ports"
)

// MockRepository is a mock implementation of TelemetryRepository for testing
type MockRepository struct {
	SaveStateFunc       func(ctx context.Context, state *domain.BatteryState) error
	GetCurrentStateFunc func(ctx context.Context, batteryID string) (*domain.BatteryState, error)
	GetHistoryFunc      func(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error)
}

func (m *MockRepository) SaveState(ctx context.Context, state *domain.BatteryState) error {
	if m.SaveStateFunc != nil {
		return m.SaveStateFunc(ctx, state)
	}
	return nil
}

func (m *MockRepository) GetCurrentState(ctx context.Context, batteryID string) (*domain.BatteryState, error) {
	if m.GetCurrentStateFunc != nil {
		return m.GetCurrentStateFunc(ctx, batteryID)
	}
	return nil, nil
}

func (m *MockRepository) GetHistory(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(ctx, batteryID, filter)
	}
	return nil, nil
}

// Helper function to create a valid test battery state
func createTestBatteryState() *domain.BatteryState {
	state, _ := domain.NewBatteryState(
		"battery-123",
		75.5,
		25.0,
		28.5,
		800.0,
		31.25,
		domain.OperationStateDischarging,
		map[string]interface{}{
			"vendor": "Tesla",
			"model":  "Megapack",
		},
		time.Now().Add(-1*time.Minute),
	)
	return state
}

func TestGetCurrentState_Success(t *testing.T) {
	// Given: A mock repository with test state
	testState := createTestBatteryState()
	mockRepo := &MockRepository{
		GetCurrentStateFunc: func(ctx context.Context, batteryID string) (*domain.BatteryState, error) {
			assert.Equal(t, "battery-123", batteryID)
			return testState, nil
		},
	}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	// Setup router with path variable
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/current", handler.GetCurrentState)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/battery-123/current", nil)
	rec := httptest.NewRecorder()

	// When: Getting current state
	router.ServeHTTP(rec, req)

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatteryStateResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, testState.ID, response.ID)
	assert.Equal(t, testState.BatteryID, response.BatteryID)
	assert.Equal(t, testState.SoC, response.SoC)
	assert.Equal(t, testState.Power, response.Power)
	assert.Equal(t, testState.OperationState, response.OperationState)
}

func TestGetCurrentState_NotFound(t *testing.T) {
	// Given: A mock repository that returns not found error
	mockRepo := &MockRepository{
		GetCurrentStateFunc: func(ctx context.Context, batteryID string) (*domain.BatteryState, error) {
			return nil, domain.ErrBatteryNotFound
		},
	}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/current", handler.GetCurrentState)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/battery-999/current", nil)
	rec := httptest.NewRecorder()

	// When: Getting non-existent battery state
	router.ServeHTTP(rec, req)

	// Then: Should return 404 Not Found
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", response.Code)
}

func TestGetCurrentState_InternalError(t *testing.T) {
	// Given: A mock repository that returns internal error
	mockRepo := &MockRepository{
		GetCurrentStateFunc: func(ctx context.Context, batteryID string) (*domain.BatteryState, error) {
			return nil, errors.New("database connection failed")
		},
	}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/current", handler.GetCurrentState)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/battery-123/current", nil)
	rec := httptest.NewRecorder()

	// When: Repository error occurs
	router.ServeHTTP(rec, req)

	// Then: Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", response.Code)
}

func TestGetHistory_Success(t *testing.T) {
	// Given: A mock repository with historical states
	baseTime := time.Now().Add(-10 * time.Minute)
	testStates := []*domain.BatteryState{}
	for i := 0; i < 3; i++ {
		state, _ := domain.NewBatteryState(
			"battery-456",
			float64(50+i*10),
			0.0,
			25.0,
			800.0,
			0.0,
			domain.OperationStateIdle,
			nil,
			baseTime.Add(time.Duration(i)*time.Minute),
		)
		testStates = append(testStates, state)
	}

	mockRepo := &MockRepository{
		GetHistoryFunc: func(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error) {
			assert.Equal(t, "battery-456", batteryID)
			assert.NotZero(t, filter.StartTime)
			assert.NotZero(t, filter.EndTime)
			return testStates, nil
		},
	}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/history", handler.GetHistory)

	startTime := baseTime.Format(time.RFC3339)
	endTime := baseTime.Add(5 * time.Minute).Format(time.RFC3339)
	reqURL := "/api/v1/telemetry/battery-456/history?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime)
	req := httptest.NewRequest(http.MethodGet, reqURL, nil)
	rec := httptest.NewRecorder()

	// When: Getting history
	router.ServeHTTP(rec, req)

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response BatteryStateHistoryResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.States, 3)
	assert.Equal(t, 3, response.Total)
}

func TestGetHistory_MissingStartTime(t *testing.T) {
	// Given: A handler with mock repository
	mockRepo := &MockRepository{}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/history", handler.GetHistory)

	// Missing start_time parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/battery-456/history?end_time=2025-12-30T10:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When: Calling with missing start_time
	router.ServeHTTP(rec, req)

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response.Code)
}

func TestGetHistory_MissingEndTime(t *testing.T) {
	// Given: A handler with mock repository
	mockRepo := &MockRepository{}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/history", handler.GetHistory)

	// Missing end_time parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/battery-456/history?start_time=2025-12-30T09:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When: Calling with missing end_time
	router.ServeHTTP(rec, req)

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response.Code)
}

func TestGetHistory_InvalidTimeFormat(t *testing.T) {
	// Given: A handler with mock repository
	mockRepo := &MockRepository{}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/history", handler.GetHistory)

	// Invalid time format
	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/battery-456/history?start_time=invalid&end_time=2025-12-30T10:00:00Z", nil)
	rec := httptest.NewRecorder()

	// When: Calling with invalid time format
	router.ServeHTTP(rec, req)

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response.Code)
}

func TestGetHistory_WithPagination(t *testing.T) {
	// Given: A mock repository
	mockRepo := &MockRepository{
		GetHistoryFunc: func(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error) {
			assert.Equal(t, 10, filter.Limit)
			assert.Equal(t, 5, filter.Offset)
			return []*domain.BatteryState{}, nil
		},
	}
	handler := NewTelemetryHandler(mockRepo, zap.NewNop())

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/telemetry/{batteryId}/history", handler.GetHistory)

	baseTime := time.Now().Add(-10 * time.Minute)
	startTime := baseTime.Format(time.RFC3339)
	endTime := baseTime.Add(5 * time.Minute).Format(time.RFC3339)
	reqURL := "/api/v1/telemetry/battery-456/history?start_time=" + url.QueryEscape(startTime) + "&end_time=" + url.QueryEscape(endTime) + "&limit=10&offset=5"
	req := httptest.NewRequest(http.MethodGet, reqURL, nil)
	rec := httptest.NewRecorder()

	// When: Getting history with pagination
	router.ServeHTTP(rec, req)

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)
}
