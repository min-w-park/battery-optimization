package events

import "time"

// OpportunityStopConditions defines conditions for stopping a discharge operation
type OpportunityStopConditions struct {
	PriceThreshold float64 `json:"price_threshold"` // Stop if price drops below this ($/MWh)
	Duration       int     `json:"duration"`        // Stop after this many minutes (0 = no limit)
	FcasDispatch   bool    `json:"fcas_dispatch"`   // Stop if FCAS dispatch received
}

// DischargingOpportunityDetected is published when a profitable discharging opportunity is detected
// Publisher: Bidding Service
// Trigger: Price > $100/MWh AND SoC > 30% AND battery is IDLE
type DischargingOpportunityDetected struct {
	BatteryID      string                    `json:"battery_id"`
	Price          float64                   `json:"price"`           // Current market price ($/MWh)
	SoC            float64                   `json:"soc"`             // Current battery SoC (%)
	TargetSoC      float64                   `json:"target_soc"`      // Minimum SoC for discharging (typically 30%)
	ExpectedProfit float64                   `json:"expected_profit"` // Estimated profit ($)
	StopConditions OpportunityStopConditions `json:"stop_conditions"` // Conditions for stopping discharge
	Timestamp      time.Time                 `json:"timestamp"`
	EventVersion   string                    `json:"event_version"`
}
