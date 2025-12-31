package domain

// Arbitrage algorithm constants
const (
	CHARGE_THRESHOLD_MWH    = 50.0  // Below this price: charge opportunity
	DISCHARGE_THRESHOLD_MWH = 100.0 // Above this price: discharge opportunity
	MIN_SOC_FOR_DISCHARGE   = 30.0  // Minimum SoC to discharge (%)
	MAX_SOC_FOR_CHARGE      = 80.0  // Maximum SoC to charge (%)
)

// ShouldCharge determines if battery should charge based on arbitrage conditions
//
// Returns true if ALL conditions met:
//   - price < CHARGE_THRESHOLD_MWH ($50/MWh)
//   - soc < MAX_SOC_FOR_CHARGE (80%)
//   - operationState == "IDLE"
func ShouldCharge(price, soc float64, operationState string) bool {
	// Check price threshold
	if price >= CHARGE_THRESHOLD_MWH {
		return false
	}

	// Check SoC threshold
	if soc >= MAX_SOC_FOR_CHARGE {
		return false
	}

	// Check operation state
	if operationState != "IDLE" {
		return false
	}

	return true
}

// ShouldDischarge determines if battery should discharge based on arbitrage conditions
//
// Returns true if ALL conditions met:
//   - price > DISCHARGE_THRESHOLD_MWH ($100/MWh)
//   - soc > MIN_SOC_FOR_DISCHARGE (30%)
//   - operationState == "IDLE"
func ShouldDischarge(price, soc float64, operationState string) bool {
	// Check price threshold
	if price <= DISCHARGE_THRESHOLD_MWH {
		return false
	}

	// Check SoC threshold
	if soc <= MIN_SOC_FOR_DISCHARGE {
		return false
	}

	// Check operation state
	if operationState != "IDLE" {
		return false
	}

	return true
}

// EstimateProfit calculates expected profit from arbitrage opportunity
//
// Formula: price_delta × capacity × efficiency
// Example:
//   - Charging at $30/MWh (delta = $20 below $50 threshold)
//   - Capacity = 200 MWh, Efficiency = 0.95
//   - Profit = $20 × 200 × 0.95 = $3,800
func EstimateProfit(priceDelta, capacity, efficiency float64) float64 {
	return priceDelta * capacity * efficiency
}
