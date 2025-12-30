package http

import (
	"time"

	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
)

// BatteryStateResponse represents the API response for battery state
type BatteryStateResponse struct {
	ID               string                 `json:"id"`
	BatteryID        string                 `json:"battery_id"`
	SoC              float64                `json:"soc"`
	Power            float64                `json:"power"`
	Temperature      float64                `json:"temperature"`
	Voltage          float64                `json:"voltage"`
	Current          float64                `json:"current"`
	OperationState   string                 `json:"operation_state"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	CreatedAt        time.Time              `json:"created_at"`
}

// FromDomain creates a BatteryStateResponse from a domain BatteryState
func FromDomain(state *domain.BatteryState) *BatteryStateResponse {
	return &BatteryStateResponse{
		ID:               state.ID,
		BatteryID:        state.BatteryID,
		SoC:              state.SoC,
		Power:            state.Power,
		Temperature:      state.Temperature,
		Voltage:          state.Voltage,
		Current:          state.Current,
		OperationState:   state.OperationState,
		CustomAttributes: state.CustomAttributes,
		Timestamp:        state.Timestamp,
		CreatedAt:        state.CreatedAt,
	}
}

// BatteryStateHistoryResponse represents the response for historical states
type BatteryStateHistoryResponse struct {
	States []*BatteryStateResponse `json:"states"`
	Total  int                     `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
