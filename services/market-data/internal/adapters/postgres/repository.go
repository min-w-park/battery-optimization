package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
	"github.com/minwook/battery-optimization/services/market-data/internal/ports"
)

// PostgresRepository implements MarketPriceRepository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create saves a new market price to the database
func (r *PostgresRepository) Create(ctx context.Context, price *domain.MarketPrice) error {
	query := `
		INSERT INTO market_prices
		(id, region, price, demand, interval_type, interval_start, published_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		price.ID,
		price.Region,
		price.Price,
		price.Demand,
		string(price.IntervalType),
		price.IntervalStart,
		price.PublishedAt,
		price.CreatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: %s", domain.ErrDuplicateInterval, err)
			}
		}
		return fmt.Errorf("failed to create market price: %w", err)
	}

	return nil
}

// FindByID retrieves a market price by its ID
func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*domain.MarketPrice, error) {
	query := `
		SELECT id, region, price, demand, interval_type, interval_start, published_at, created_at
		FROM market_prices
		WHERE id = $1
	`

	var price domain.MarketPrice
	var intervalType string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&price.ID,
		&price.Region,
		&price.Price,
		&price.Demand,
		&intervalType,
		&price.IntervalStart,
		&price.PublishedAt,
		&price.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: id=%s", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to find market price: %w", err)
	}

	price.IntervalType = domain.IntervalType(intervalType)
	return &price, nil
}

// ListByTimeRange retrieves market prices within a time range with optional filters
func (r *PostgresRepository) ListByTimeRange(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
	query := `
		SELECT id, region, price, demand, interval_type, interval_start, published_at, created_at
		FROM market_prices
		WHERE interval_start >= $1 AND interval_start <= $2
	`
	args := []interface{}{filter.From, filter.To}
	argCount := 2

	// Add optional filters
	if filter.Region != "" {
		argCount++
		query += fmt.Sprintf(" AND region = $%d", argCount)
		args = append(args, filter.Region)
	}

	if filter.IntervalType != "" {
		argCount++
		query += fmt.Sprintf(" AND interval_type = $%d", argCount)
		args = append(args, string(filter.IntervalType))
	}

	// Order by time descending (newest first)
	query += " ORDER BY interval_start DESC"

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
		return nil, fmt.Errorf("failed to list market prices: %w", err)
	}
	defer rows.Close()

	var prices []*domain.MarketPrice
	for rows.Next() {
		var price domain.MarketPrice
		var intervalType string

		err := rows.Scan(
			&price.ID,
			&price.Region,
			&price.Price,
			&price.Demand,
			&intervalType,
			&price.IntervalStart,
			&price.PublishedAt,
			&price.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan market price: %w", err)
		}

		price.IntervalType = domain.IntervalType(intervalType)
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating market prices: %w", err)
	}

	return prices, nil
}
