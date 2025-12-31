package domain_test

import (
	"testing"
	"time"

	"github.com/minwook/battery-optimization/services/bidding/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBiddingDecision_ValidInput(t *testing.T) {
	decision, err := domain.NewBiddingDecision(
		"battery-123",
		"CHARGE",
		"Low price arbitrage",
		30.0,
		50.0,
		"FULL_AUTO",
	)

	require.NoError(t, err)
	assert.NotNil(t, decision)
	assert.NotEmpty(t, decision.ID)
	assert.Equal(t, "battery-123", decision.BatteryID)
	assert.Equal(t, string(domain.DecisionCharge), string(decision.DecisionType))
	assert.Equal(t, "Low price arbitrage", decision.Reason)
	assert.Equal(t, 30.0, decision.Price)
	assert.Equal(t, 50.0, decision.SoC)
	assert.Equal(t, string(domain.ModeFullAuto), string(decision.AutomationMode))
	assert.False(t, decision.Timestamp.IsZero())
}

func TestNewBiddingDecision_InvalidSoC(t *testing.T) {
	tests := []struct {
		name string
		soc  float64
	}{
		{"negative SoC", -10.0},
		{"SoC above 100", 110.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewBiddingDecision(
				"battery-123",
				"CHARGE",
				"test",
				30.0,
				tt.soc,
				"MANUAL",
			)

			assert.ErrorIs(t, err, domain.ErrInvalidSoC)
		})
	}
}

func TestNewBiddingDecision_InvalidPrice(t *testing.T) {
	tests := []struct {
		name  string
		price float64
	}{
		{"zero price", 0.0},
		{"negative price", -50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewBiddingDecision(
				"battery-123",
				"CHARGE",
				"test",
				tt.price,
				50.0,
				"MANUAL",
			)

			assert.ErrorIs(t, err, domain.ErrInvalidPrice)
		})
	}
}

func TestNewBiddingDecision_InvalidDecisionType(t *testing.T) {
	_, err := domain.NewBiddingDecision(
		"battery-123",
		"INVALID_TYPE",
		"test",
		30.0,
		50.0,
		"MANUAL",
	)

	assert.ErrorIs(t, err, domain.ErrInvalidDecisionType)
}

func TestNewBiddingDecision_InvalidAutomationMode(t *testing.T) {
	_, err := domain.NewBiddingDecision(
		"battery-123",
		"CHARGE",
		"test",
		30.0,
		50.0,
		"INVALID_MODE",
	)

	assert.ErrorIs(t, err, domain.ErrInvalidAutomationMode)
}

func TestNewBiddingDecision_EmptyBatteryID(t *testing.T) {
	_, err := domain.NewBiddingDecision(
		"",
		"CHARGE",
		"test",
		30.0,
		50.0,
		"MANUAL",
	)

	assert.ErrorIs(t, err, domain.ErrEmptyBatteryID)
}

func TestNewBiddingDecision_AllDecisionTypes(t *testing.T) {
	types := []string{"CHARGE", "DISCHARGE", "NO_ACTION"}

	for _, decisionType := range types {
		t.Run(decisionType, func(t *testing.T) {
			decision, err := domain.NewBiddingDecision(
				"battery-123",
				decisionType,
				"test reason",
				100.0,
				50.0,
				"MANUAL",
			)

			require.NoError(t, err)
			assert.Equal(t, decisionType, string(decision.DecisionType))
		})
	}
}

func TestNewBiddingDecision_AllAutomationModes(t *testing.T) {
	modes := []string{"MANUAL", "SEMI_AUTO", "FULL_AUTO"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			decision, err := domain.NewBiddingDecision(
				"battery-123",
				"CHARGE",
				"test reason",
				30.0,
				50.0,
				mode,
			)

			require.NoError(t, err)
			assert.Equal(t, mode, string(decision.AutomationMode))
		})
	}
}

func TestBiddingDecision_TimestampIsRecent(t *testing.T) {
	before := time.Now()

	decision, err := domain.NewBiddingDecision(
		"battery-123",
		"CHARGE",
		"test",
		30.0,
		50.0,
		"MANUAL",
	)

	after := time.Now()

	require.NoError(t, err)
	assert.True(t, decision.Timestamp.After(before) || decision.Timestamp.Equal(before))
	assert.True(t, decision.Timestamp.Before(after) || decision.Timestamp.Equal(after))
}
