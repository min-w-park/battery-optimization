package domain_test

import (
	"testing"

	"github.com/minwook/battery-optimization/services/bidding/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Charging Tests
func TestShouldCharge_LowPrice_LowSoC(t *testing.T) {
	// Price < $50, SoC < 80%, state = IDLE
	result := domain.ShouldCharge(30.0, 50.0, "IDLE")
	assert.True(t, result, "Should charge when price is low and SoC is low")
}

func TestShouldCharge_LowPrice_HighSoC(t *testing.T) {
	// Price < $50, but SoC >= 80%
	result := domain.ShouldCharge(30.0, 85.0, "IDLE")
	assert.False(t, result, "Should not charge when SoC is >= 80%")
}

func TestShouldCharge_HighPrice_LowSoC(t *testing.T) {
	// Price >= $50, even though SoC < 80%
	result := domain.ShouldCharge(60.0, 50.0, "IDLE")
	assert.False(t, result, "Should not charge when price is >= $50")
}

func TestShouldCharge_NotIdle(t *testing.T) {
	// All conditions met except state != IDLE
	tests := []struct {
		state string
		name  string
	}{
		{"CHARGING", "already charging"},
		{"DISCHARGING", "currently discharging"},
		{"FCAS", "in FCAS contract"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.ShouldCharge(30.0, 50.0, tt.state)
			assert.False(t, result, "Should not charge when state is %s", tt.state)
		})
	}
}

func TestShouldCharge_ExactThreshold(t *testing.T) {
	// Price exactly at threshold (edge case)
	result := domain.ShouldCharge(50.0, 50.0, "IDLE")
	assert.False(t, result, "Should not charge when price is exactly at threshold")
}

func TestShouldCharge_SoCExactly80(t *testing.T) {
	// SoC exactly at 80% (edge case)
	result := domain.ShouldCharge(30.0, 80.0, "IDLE")
	assert.False(t, result, "Should not charge when SoC is exactly 80%")
}

// Discharging Tests
func TestShouldDischarge_HighPrice_HighSoC(t *testing.T) {
	// Price > $100, SoC > 30%, state = IDLE
	result := domain.ShouldDischarge(150.0, 60.0, "IDLE")
	assert.True(t, result, "Should discharge when price is high and SoC is high")
}

func TestShouldDischarge_HighPrice_LowSoC(t *testing.T) {
	// Price > $100, but SoC <= 30%
	result := domain.ShouldDischarge(150.0, 25.0, "IDLE")
	assert.False(t, result, "Should not discharge when SoC is <= 30%")
}

func TestShouldDischarge_LowPrice_HighSoC(t *testing.T) {
	// Price <= $100, even though SoC > 30%
	result := domain.ShouldDischarge(80.0, 60.0, "IDLE")
	assert.False(t, result, "Should not discharge when price is <= $100")
}

func TestShouldDischarge_NotIdle(t *testing.T) {
	// All conditions met except state != IDLE
	tests := []struct {
		state string
		name  string
	}{
		{"CHARGING", "currently charging"},
		{"DISCHARGING", "already discharging"},
		{"FCAS", "in FCAS contract"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.ShouldDischarge(150.0, 60.0, tt.state)
			assert.False(t, result, "Should not discharge when state is %s", tt.state)
		})
	}
}

func TestShouldDischarge_ExactThreshold(t *testing.T) {
	// Price exactly at threshold (edge case)
	result := domain.ShouldDischarge(100.0, 60.0, "IDLE")
	assert.False(t, result, "Should not discharge when price is exactly at threshold")
}

func TestShouldDischarge_SoCExactly30(t *testing.T) {
	// SoC exactly at 30% (edge case)
	result := domain.ShouldDischarge(150.0, 30.0, "IDLE")
	assert.False(t, result, "Should not discharge when SoC is exactly 30%")
}

// Combined Scenarios
func TestArbitrage_MidRangePrice(t *testing.T) {
	// Price in the middle range ($50-$100)
	price := 75.0
	soc := 60.0
	state := "IDLE"

	shouldCharge := domain.ShouldCharge(price, soc, state)
	shouldDischarge := domain.ShouldDischarge(price, soc, state)

	assert.False(t, shouldCharge, "Should not charge at mid-range price")
	assert.False(t, shouldDischarge, "Should not discharge at mid-range price")
}
