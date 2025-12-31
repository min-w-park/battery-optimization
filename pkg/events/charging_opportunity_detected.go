package events

import "time"

// ChargingOpportunityDetected is published when a profitable charging opportunity is detected
// Publisher: Bidding Service
// Trigger: Price < $50/MWh AND SoC < 80% AND battery is IDLE
type ChargingOpportunityDetected struct {
	BatteryID      string    `json:"battery_id"`
	Price          float64   `json:"price"`           // Current market price ($/MWh)
	SoC            float64   `json:"soc"`             // Current battery SoC (%)
	TargetSoC      float64   `json:"target_soc"`      // Target SoC for charging (typically 80%)
	ExpectedProfit float64   `json:"expected_profit"` // Estimated profit ($)
	Timestamp      time.Time `json:"timestamp"`
	EventVersion   string    `json:"event_version"`
}
