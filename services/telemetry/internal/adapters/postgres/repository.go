package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
	"github.com/minwook/battery-optimization/services/telemetry/internal/ports"
)

// PostgresRepository implements TelemetryRepository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// SaveState persists a battery state snapshot to the database
func (r *PostgresRepository) SaveState(ctx context.Context, state *domain.BatteryState) error {
	query := `
		INSERT INTO battery_states
		(id, battery_id, soc, power, temperature, voltage, current, operation_state, custom_attributes, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Marshal custom_attributes to JSON
	var customAttrs interface{}
	if state.CustomAttributes != nil {
		customAttrsJSON, err := json.Marshal(state.CustomAttributes)
		if err != nil {
			return fmt.Errorf("failed to marshal custom attributes: %w", err)
		}
		customAttrs = customAttrsJSON
	} else {
		customAttrs = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		state.ID,
		state.BatteryID,
		state.SoC,
		state.Power,
		state.Temperature,
		state.Voltage,
		state.Current,
		state.OperationState,
		customAttrs,
		state.Timestamp,
		state.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save battery state: %w", err)
	}

	return nil
}

// GetCurrentState retrieves the most recent state for a battery
func (r *PostgresRepository) GetCurrentState(ctx context.Context, batteryID string) (*domain.BatteryState, error) {
	query := `
		SELECT id, battery_id, soc, power, temperature, voltage, current, operation_state, custom_attributes, timestamp, created_at
		FROM battery_states
		WHERE battery_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var state domain.BatteryState
	var customAttrsJSON []byte

	err := r.db.QueryRowContext(ctx, query, batteryID).Scan(
		&state.ID,
		&state.BatteryID,
		&state.SoC,
		&state.Power,
		&state.Temperature,
		&state.Voltage,
		&state.Current,
		&state.OperationState,
		&customAttrsJSON,
		&state.Timestamp,
		&state.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: battery_id=%s", domain.ErrBatteryNotFound, batteryID)
		}
		return nil, fmt.Errorf("failed to get current battery state: %w", err)
	}

	// Unmarshal custom_attributes from JSON
	if customAttrsJSON != nil {
		if err := json.Unmarshal(customAttrsJSON, &state.CustomAttributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom attributes: %w", err)
		}
	}

	return &state, nil
}

// GetHistory retrieves battery states within a time range with pagination
func (r *PostgresRepository) GetHistory(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error) {
	query := `
		SELECT id, battery_id, soc, power, temperature, voltage, current, operation_state, custom_attributes, timestamp, created_at
		FROM battery_states
		WHERE battery_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp DESC
	`
	args := []interface{}{batteryID, filter.StartTime, filter.EndTime}
	argCount := 3

	// Pagination
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get battery state history: %w", err)
	}
	defer rows.Close()

	var states []*domain.BatteryState
	for rows.Next() {
		var state domain.BatteryState
		var customAttrsJSON []byte

		err := rows.Scan(
			&state.ID,
			&state.BatteryID,
			&state.SoC,
			&state.Power,
			&state.Temperature,
			&state.Voltage,
			&state.Current,
			&state.OperationState,
			&customAttrsJSON,
			&state.Timestamp,
			&state.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan battery state: %w", err)
		}

		// Unmarshal custom_attributes from JSON
		if customAttrsJSON != nil {
			if err := json.Unmarshal(customAttrsJSON, &state.CustomAttributes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal custom attributes: %w", err)
			}
		}

		states = append(states, &state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating battery states: %w", err)
	}

	return states, nil
}
