package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
	"github.com/minwook/battery-optimization/services/asset-management/internal/ports"
)

// PostgresRepository implements the BatteryRepository interface using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new battery into the database
func (r *PostgresRepository) Create(ctx context.Context, battery *domain.Battery) error {
	// Marshal constraints to JSONB
	constraintsJSON, err := json.Marshal(battery.Constraints)
	if err != nil {
		return fmt.Errorf("failed to marshal constraints: %w", err)
	}

	query := `
		INSERT INTO batteries (
			id, capacity, max_power, ramp_rate, efficiency,
			location, manufacturer, constraints, status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		battery.ID,
		battery.Capacity,
		battery.MaxPower,
		battery.RampRate,
		battery.Efficiency,
		battery.Location,
		battery.Manufacturer,
		constraintsJSON,
		battery.Status,
		battery.CreatedAt,
		battery.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate key error
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: %s", domain.ErrDuplicateID, err)
			}
		}
		return fmt.Errorf("failed to create battery: %w", err)
	}

	return nil
}

// FindByID retrieves a battery by its ID
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*domain.Battery, error) {
	query := `
		SELECT id, capacity, max_power, ramp_rate, efficiency,
		       location, manufacturer, constraints, status,
		       created_at, updated_at
		FROM batteries
		WHERE id = $1
	`

	var battery domain.Battery
	var constraintsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&battery.ID,
		&battery.Capacity,
		&battery.MaxPower,
		&battery.RampRate,
		&battery.Efficiency,
		&battery.Location,
		&battery.Manufacturer,
		&constraintsJSON,
		&battery.Status,
		&battery.CreatedAt,
		&battery.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find battery: %w", err)
	}

	// Unmarshal constraints from JSONB
	if err := json.Unmarshal(constraintsJSON, &battery.Constraints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal constraints: %w", err)
	}

	return &battery, nil
}

// List retrieves batteries based on the provided filter
func (r *PostgresRepository) List(ctx context.Context, filter ports.ListFilter) ([]*domain.Battery, error) {
	// Build dynamic query with filters
	query := `
		SELECT id, capacity, max_power, ramp_rate, efficiency,
		       location, manufacturer, constraints, status,
		       created_at, updated_at
		FROM batteries
	`

	var conditions []string
	var args []interface{}
	argCount := 0

	// Add location filter if provided
	if filter.Location != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("location = $%d", argCount))
		args = append(args, filter.Location)
	}

	// Add status filter if provided
	if filter.Status != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, filter.Status)
	}

	// Add WHERE clause if there are conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering (consistent ordering for pagination)
	query += " ORDER BY created_at DESC"

	// Add pagination if limit is set
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	// Add offset if set
	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list batteries: %w", err)
	}
	defer rows.Close()

	var batteries []*domain.Battery

	for rows.Next() {
		var battery domain.Battery
		var constraintsJSON []byte

		err := rows.Scan(
			&battery.ID,
			&battery.Capacity,
			&battery.MaxPower,
			&battery.RampRate,
			&battery.Efficiency,
			&battery.Location,
			&battery.Manufacturer,
			&constraintsJSON,
			&battery.Status,
			&battery.CreatedAt,
			&battery.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan battery row: %w", err)
		}

		// Unmarshal constraints from JSONB
		if err := json.Unmarshal(constraintsJSON, &battery.Constraints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal constraints: %w", err)
		}

		batteries = append(batteries, &battery)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating battery rows: %w", err)
	}

	return batteries, nil
}
