package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
	"github.com/minwook/battery-optimization/services/telemetry/internal/ports"
)

// TelemetryHandler handles HTTP requests for telemetry operations
type TelemetryHandler struct {
	repo ports.TelemetryRepository
	log  *zap.Logger
}

// NewTelemetryHandler creates a new telemetry handler
func NewTelemetryHandler(repo ports.TelemetryRepository, log *zap.Logger) *TelemetryHandler {
	return &TelemetryHandler{
		repo: repo,
		log:  log,
	}
}

// GetCurrentState handles GET /api/v1/telemetry/{batteryId}/current
func (h *TelemetryHandler) GetCurrentState(w http.ResponseWriter, r *http.Request) {
	// Extract batteryId from URL
	vars := mux.Vars(r)
	batteryID := vars["batteryId"]

	// Get current state from repository
	state, err := h.repo.GetCurrentState(r.Context(), batteryID)
	if err != nil {
		if errors.Is(err, domain.ErrBatteryNotFound) {
			h.respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Battery state not found", nil)
			return
		}
		h.log.Error("failed to get current battery state",
			zap.String("battery_id", batteryID),
			zap.Error(err),
		)
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve battery state", nil)
		return
	}

	// Return success response
	h.respondWithJSON(w, http.StatusOK, FromDomain(state))
}

// GetHistory handles GET /api/v1/telemetry/{batteryId}/history
func (h *TelemetryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	// Extract batteryId from URL
	vars := mux.Vars(r)
	batteryID := vars["batteryId"]

	// Parse query parameters
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Validate required parameters
	if startTimeStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "start_time parameter is required", nil)
		return
	}
	if endTimeStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "end_time parameter is required", nil)
		return
	}

	// Parse time parameters (RFC3339 format)
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid start_time format (use RFC3339)", nil)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid end_time format (use RFC3339)", nil)
		return
	}

	// Parse pagination parameters (optional)
	limit := 100 // Default limit
	offset := 0  // Default offset

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid limit parameter", nil)
			return
		}
		limit = parsedLimit
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			h.respondWithError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid offset parameter", nil)
			return
		}
		offset = parsedOffset
	}

	// Build filter
	filter := ports.TimeRangeFilter{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     limit,
		Offset:    offset,
	}

	// Get history from repository
	states, err := h.repo.GetHistory(r.Context(), batteryID, filter)
	if err != nil {
		h.log.Error("failed to get battery state history", zap.String("battery_id", batteryID), zap.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve battery history", nil)
		return
	}

	// Convert to response DTOs
	stateResponses := make([]*BatteryStateResponse, len(states))
	for i, state := range states {
		stateResponses[i] = FromDomain(state)
	}

	// Return success response
	response := BatteryStateHistoryResponse{
		States: stateResponses,
		Total:  len(stateResponses),
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

// respondWithJSON writes a JSON response
func (h *TelemetryHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.log.Error("failed to encode JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// respondWithError writes an error response
func (h *TelemetryHandler) respondWithError(w http.ResponseWriter, status int, code string, message string, details map[string]string) {
	errorResponse := ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
	h.respondWithJSON(w, status, errorResponse)
}
