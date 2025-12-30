package ports

import (
	"context"
	"time"

	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
)

// TimeRangeFilter represents query parameters for historical data
type TimeRangeFilter struct {
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// TelemetryRepository defines the interface for battery state persistence
type TelemetryRepository interface {
	// SaveState persists a battery state snapshot
	SaveState(ctx context.Context, state *domain.BatteryState) error

	// GetCurrentState retrieves the most recent state for a battery
	GetCurrentState(ctx context.Context, batteryID string) (*domain.BatteryState, error)

	// GetHistory retrieves battery states within a time range
	GetHistory(ctx context.Context, batteryID string, filter TimeRangeFilter) ([]*domain.BatteryState, error)
}
