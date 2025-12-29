package ports

import (
	"context"

	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
)

// BatteryRepository defines the interface for battery persistence operations
type BatteryRepository interface {
	// Create persists a new battery to the repository
	Create(ctx context.Context, battery *domain.Battery) error

	// FindByID retrieves a battery by its unique identifier
	// Returns domain.ErrNotFound if the battery doesn't exist
	FindByID(ctx context.Context, id string) (*domain.Battery, error)

	// List retrieves batteries based on the provided filter
	List(ctx context.Context, filter ListFilter) ([]*domain.Battery, error)
}

// ListFilter defines filtering and pagination options for listing batteries
type ListFilter struct {
	Limit    int                  // Maximum number of results (0 = no limit)
	Offset   int                  // Number of results to skip
	Location string               // Filter by NEM region (empty = no filter)
	Status   domain.BatteryStatus // Filter by status (empty = no filter)
}
