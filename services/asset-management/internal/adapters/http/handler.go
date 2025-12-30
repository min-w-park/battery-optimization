package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
	"github.com/minwook/battery-optimization/services/asset-management/internal/ports"
)

// BatteryHandler handles HTTP requests for battery operations
type BatteryHandler struct {
	repo      ports.BatteryRepository
	publisher events.EventPublisher // Optional - service works without events
}

// NewBatteryHandler creates a new battery handler
func NewBatteryHandler(repo ports.BatteryRepository, publisher events.EventPublisher) *BatteryHandler {
	return &BatteryHandler{
		repo:      repo,
		publisher: publisher,
	}
}

// CreateBattery handles POST /api/v1/batteries
func (h *BatteryHandler) CreateBattery(w http.ResponseWriter, r *http.Request) {
	var req CreateBatteryRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format", nil)
		return
	}

	// Convert DTO to domain model (includes validation)
	battery, err := req.ToDomain()
	if err != nil {
		// Check if it's a domain validation error
		if errors.Is(err, domain.ErrInvalidCapacity) ||
			errors.Is(err, domain.ErrInvalidMaxPower) ||
			errors.Is(err, domain.ErrMaxPowerExceedsCapacity) ||
			errors.Is(err, domain.ErrInvalidRampRate) ||
			errors.Is(err, domain.ErrRampRateExceedsMaxPower) ||
			errors.Is(err, domain.ErrInvalidEfficiency) ||
			errors.Is(err, domain.ErrInvalidLocation) ||
			errors.Is(err, domain.ErrInvalidWarrantyEOL) ||
			errors.Is(err, domain.ErrInvalidMaxCycles) ||
			errors.Is(err, domain.ErrInvalidTemperatureRange) {
			respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create battery", nil)
		return
	}

	// Persist battery
	if err := h.repo.Create(r.Context(), battery); err != nil {
		log.Printf("Failed to create battery: %v", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save battery", nil)
		return
	}

	// Publish BatteryRegistered event (best-effort, don't fail HTTP request)
	if h.publisher != nil {
		event := events.BatteryRegistered{
			BatteryID:    battery.ID,
			Capacity:     battery.Capacity,
			MaxPower:     battery.MaxPower,
			RampRate:     battery.RampRate,
			Efficiency:   battery.Efficiency,
			Location:     battery.Location,
			Manufacturer: battery.Manufacturer,
			Constraints: events.BatteryConstraints{
				// Default SoC constraints (10%-90% operational range)
				MinSoC:              0.1,
				MaxSoC:              0.9,
				// Physical/warranty constraints from domain
				WarrantyEOL:         battery.Constraints.WarrantyEOL,
				MaxCycles:           battery.Constraints.MaxCycles,
				OperatingTempMin:    battery.Constraints.OperatingTempMin,
				OperatingTempMax:    battery.Constraints.OperatingTempMax,
				GridComplianceLevel: battery.Constraints.GridComplianceLevel,
			},
			Timestamp:    time.Now(),
			EventVersion: "v1",
		}

		// Use context with timeout for event publishing
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := h.publisher.Publish(ctx, "battery.registered.v1", event); err != nil {
			// Log warning but don't fail the HTTP request
			log.Printf("WARNING: Failed to publish BatteryRegistered event for battery %s: %v", battery.ID, err)
		} else {
			log.Printf("Published BatteryRegistered event for battery %s", battery.ID)
		}
	}

	// Return success response
	respondWithJSON(w, http.StatusCreated, FromDomain(battery))
}

// GetBattery handles GET /api/v1/batteries/{id}
func (h *BatteryHandler) GetBattery(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Find battery
	battery, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Battery not found", nil)
			return
		}
		log.Printf("Failed to find battery: %v", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve battery", nil)
		return
	}

	// Return battery
	respondWithJSON(w, http.StatusOK, FromDomain(battery))
}

// ListBatteries handles GET /api/v1/batteries
func (h *BatteryHandler) ListBatteries(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := ports.ListFilter{
		Location: r.URL.Query().Get("location"),
		Status:   domain.BatteryStatus(r.URL.Query().Get("status")),
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Get batteries from repository
	batteries, err := h.repo.List(r.Context(), filter)
	if err != nil {
		log.Printf("Failed to list batteries: %v", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list batteries", nil)
		return
	}

	// Convert to response DTOs
	batteryResponses := make([]*BatteryResponse, len(batteries))
	for i, battery := range batteries {
		batteryResponses[i] = FromDomain(battery)
	}

	// Return list
	response := ListBatteriesResponse{
		Batteries: batteryResponses,
		Total:     len(batteryResponses),
	}
	respondWithJSON(w, http.StatusOK, response)
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, code, message string, details map[string]string) {
	respondWithJSON(w, statusCode, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
