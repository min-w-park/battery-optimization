package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainErrors_Unwrap(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidSoC", ErrInvalidSoC},
		{"ErrInvalidTemperature", ErrInvalidTemperature},
		{"ErrInvalidOperationState", ErrInvalidOperationState},
		{"ErrInconsistentPowerState", ErrInconsistentPowerState},
		{"ErrMissingRequiredField", ErrMissingRequiredField},
		{"ErrInvalidTimestamp", ErrInvalidTimestamp},
		{"ErrBatteryNotFound", ErrBatteryNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error implements error interface
			assert.Implements(t, (*error)(nil), tt.err)

			// Verify error has a message
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestDomainErrors_IsComparable(t *testing.T) {
	// Test that errors can be compared with errors.Is
	testErr := ErrInvalidSoC

	assert.True(t, errors.Is(testErr, ErrInvalidSoC))
	assert.False(t, errors.Is(testErr, ErrInvalidTemperature))
}
