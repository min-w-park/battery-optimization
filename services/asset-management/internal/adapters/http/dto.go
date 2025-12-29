package http

import (
	"time"

	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
)

// ConstraintsRequest represents the constraints in the API request
type ConstraintsRequest struct {
	WarrantyEOL         float64 `json:"warranty_eol"`
	MaxCycles           int     `json:"max_cycles"`
	OperatingTempMin    float64 `json:"operating_temp_min"`
	OperatingTempMax    float64 `json:"operating_temp_max"`
	GridComplianceLevel string  `json:"grid_compliance_level"`
}

// CreateBatteryRequest represents the request body for creating a battery
type CreateBatteryRequest struct {
	Capacity     float64            `json:"capacity"`
	MaxPower     float64            `json:"max_power"`
	RampRate     float64            `json:"ramp_rate"`
	Efficiency   float64            `json:"efficiency"`
	Location     string             `json:"location"`
	Manufacturer string             `json:"manufacturer"`
	Constraints  ConstraintsRequest `json:"constraints"`
}

// ToDomain converts the DTO to a domain Battery
func (r *CreateBatteryRequest) ToDomain() (*domain.Battery, error) {
	constraints := domain.Constraints{
		WarrantyEOL:         r.Constraints.WarrantyEOL,
		MaxCycles:           r.Constraints.MaxCycles,
		OperatingTempMin:    r.Constraints.OperatingTempMin,
		OperatingTempMax:    r.Constraints.OperatingTempMax,
		GridComplianceLevel: r.Constraints.GridComplianceLevel,
	}

	return domain.NewBattery(
		r.Capacity,
		r.MaxPower,
		r.RampRate,
		r.Efficiency,
		r.Location,
		r.Manufacturer,
		constraints,
	)
}

// ConstraintsResponse represents the constraints in the API response
type ConstraintsResponse struct {
	WarrantyEOL         float64 `json:"warranty_eol"`
	MaxCycles           int     `json:"max_cycles"`
	OperatingTempMin    float64 `json:"operating_temp_min"`
	OperatingTempMax    float64 `json:"operating_temp_max"`
	GridComplianceLevel string  `json:"grid_compliance_level"`
}

// BatteryResponse represents the response body for a battery
type BatteryResponse struct {
	ID           string              `json:"id"`
	Capacity     float64             `json:"capacity"`
	MaxPower     float64             `json:"max_power"`
	RampRate     float64             `json:"ramp_rate"`
	Efficiency   float64             `json:"efficiency"`
	Location     string              `json:"location"`
	Manufacturer string              `json:"manufacturer"`
	Constraints  ConstraintsResponse `json:"constraints"`
	Status       string              `json:"status"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// FromDomain creates a BatteryResponse from a domain Battery
func FromDomain(battery *domain.Battery) *BatteryResponse {
	return &BatteryResponse{
		ID:           battery.ID,
		Capacity:     battery.Capacity,
		MaxPower:     battery.MaxPower,
		RampRate:     battery.RampRate,
		Efficiency:   battery.Efficiency,
		Location:     battery.Location,
		Manufacturer: battery.Manufacturer,
		Constraints: ConstraintsResponse{
			WarrantyEOL:         battery.Constraints.WarrantyEOL,
			MaxCycles:           battery.Constraints.MaxCycles,
			OperatingTempMin:    battery.Constraints.OperatingTempMin,
			OperatingTempMax:    battery.Constraints.OperatingTempMax,
			GridComplianceLevel: battery.Constraints.GridComplianceLevel,
		},
		Status:    string(battery.Status),
		CreatedAt: battery.CreatedAt,
		UpdatedAt: battery.UpdatedAt,
	}
}

// ListBatteriesResponse represents the response for listing batteries
type ListBatteriesResponse struct {
	Batteries []*BatteryResponse `json:"batteries"`
	Total     int                `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
