package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/minwook/battery-optimization/services/market-data/internal/domain"
	"github.com/minwook/battery-optimization/services/market-data/internal/ports"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Connect to market_data database
	connStr := "postgres://market_user:market_pass@localhost:5433/market_data?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	return db
}

// cleanupTestDB cleans up test data
func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM market_prices")
	require.NoError(t, err, "Failed to clean up test data")
}

func TestCreate_Success(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	intervalStart := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Now().Add(-5 * time.Minute)

	price, err := domain.NewMarketPrice(
		"NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart, publishedAt,
	)
	require.NoError(t, err)

	// When
	err = repo.Create(ctx, price)

	// Then
	assert.NoError(t, err)

	// Verify in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM market_prices WHERE id = $1", price.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreate_DuplicateInterval(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	intervalStart := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Now().Add(-5 * time.Minute)

	// Create first price
	price1, err := domain.NewMarketPrice(
		"NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart, publishedAt,
	)
	require.NoError(t, err)

	err = repo.Create(ctx, price1)
	require.NoError(t, err)

	// Create second price with same region, interval_type, interval_start
	price2, err := domain.NewMarketPrice(
		"NSW", 90.00, 8300.0,
		domain.Interval5Min, intervalStart, publishedAt,
	)
	require.NoError(t, err)

	// When
	err = repo.Create(ctx, price2)

	// Then
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDuplicateInterval)
}

func TestFindByID_Success(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	intervalStart := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2025, 12, 29, 9, 55, 0, 0, time.UTC)

	expectedPrice, err := domain.NewMarketPrice(
		"NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart, publishedAt,
	)
	require.NoError(t, err)

	err = repo.Create(ctx, expectedPrice)
	require.NoError(t, err)

	// When
	actualPrice, err := repo.FindByID(ctx, expectedPrice.ID)

	// Then
	require.NoError(t, err)
	assert.NotNil(t, actualPrice)
	assert.Equal(t, expectedPrice.ID, actualPrice.ID)
	assert.Equal(t, expectedPrice.Region, actualPrice.Region)
	assert.Equal(t, expectedPrice.Price, actualPrice.Price)
	assert.Equal(t, expectedPrice.Demand, actualPrice.Demand)
	assert.Equal(t, expectedPrice.IntervalType, actualPrice.IntervalType)
	assert.Equal(t, expectedPrice.IntervalStart.Unix(), actualPrice.IntervalStart.Unix())
	assert.Equal(t, expectedPrice.PublishedAt.Unix(), actualPrice.PublishedAt.Unix())
}

func TestFindByID_NotFound(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// When
	price, err := repo.FindByID(ctx, "non-existent-id")

	// Then
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, price)
}

func TestListByTimeRange_NoFilters(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert 10 prices across different regions and times
	prices := []*domain.MarketPrice{}
	for i := 0; i < 10; i++ {
		intervalStart := time.Date(2025, 12, 30, 10+i/2, (i%2)*5, 0, 0, time.UTC)
		publishedAt := time.Now().Add(-5 * time.Minute)
		region := []string{"NSW", "VIC", "QLD", "SA", "TAS"}[i%5]

		price, err := domain.NewMarketPrice(
			region, 80.0+float64(i), 8000.0,
			domain.Interval5Min, intervalStart, publishedAt,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, price)
		require.NoError(t, err)

		prices = append(prices, price)
	}

	// When
	filter := ports.TimeRangeFilter{
		From: time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	results, err := repo.ListByTimeRange(ctx, filter)

	// Then
	require.NoError(t, err)
	assert.Len(t, results, 10)

	// Verify order (DESC by interval_start)
	for i := 1; i < len(results); i++ {
		assert.True(t, results[i-1].IntervalStart.After(results[i].IntervalStart) ||
			results[i-1].IntervalStart.Equal(results[i].IntervalStart))
	}
}

func TestListByTimeRange_WithRegionFilter(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert prices for NSW, VIC, SA
	regions := []string{"NSW", "VIC", "SA"}
	for i, region := range regions {
		intervalStart := time.Date(2025, 12, 30, 10, i*5, 0, 0, time.UTC)
		publishedAt := time.Now().Add(-5 * time.Minute)

		price, err := domain.NewMarketPrice(
			region, 80.0+float64(i), 8000.0,
			domain.Interval5Min, intervalStart, publishedAt,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, price)
		require.NoError(t, err)
	}

	// When
	filter := ports.TimeRangeFilter{
		From:   time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Region: "NSW",
	}
	results, err := repo.ListByTimeRange(ctx, filter)

	// Then
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "NSW", results[0].Region)
}

func TestListByTimeRange_WithIntervalTypeFilter(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert 5MIN and 30MIN prices
	intervalStart5Min := time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC)
	publishedAt := time.Now().Add(-5 * time.Minute)

	price5Min, err := domain.NewMarketPrice(
		"NSW", 85.50, 8200.0,
		domain.Interval5Min, intervalStart5Min, publishedAt,
	)
	require.NoError(t, err)
	err = repo.Create(ctx, price5Min)
	require.NoError(t, err)

	intervalStart30Min := time.Date(2025, 12, 30, 10, 30, 0, 0, time.UTC)
	price30Min, err := domain.NewMarketPrice(
		"NSW", 90.00, 8300.0,
		domain.Interval30Min, intervalStart30Min, publishedAt,
	)
	require.NoError(t, err)
	err = repo.Create(ctx, price30Min)
	require.NoError(t, err)

	// When
	filter := ports.TimeRangeFilter{
		From:         time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:           time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		IntervalType: domain.Interval5Min,
	}
	results, err := repo.ListByTimeRange(ctx, filter)

	// Then
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, domain.Interval5Min, results[0].IntervalType)
}

func TestListByTimeRange_WithPagination(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert 25 prices
	for i := 0; i < 25; i++ {
		intervalStart := time.Date(2025, 12, 30, 10, i*5, 0, 0, time.UTC)
		publishedAt := time.Now().Add(-5 * time.Minute)

		price, err := domain.NewMarketPrice(
			"NSW", 80.0+float64(i), 8000.0,
			domain.Interval5Min, intervalStart, publishedAt,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, price)
		require.NoError(t, err)
	}

	// When: First page (limit=10, offset=0)
	filter1 := ports.TimeRangeFilter{
		From:   time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Limit:  10,
		Offset: 0,
	}
	results1, err := repo.ListByTimeRange(ctx, filter1)

	// Then
	require.NoError(t, err)
	assert.Len(t, results1, 10)

	// When: Second page (limit=10, offset=10)
	filter2 := ports.TimeRangeFilter{
		From:   time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Limit:  10,
		Offset: 10,
	}
	results2, err := repo.ListByTimeRange(ctx, filter2)

	// Then
	require.NoError(t, err)
	assert.Len(t, results2, 10)

	// Verify different results
	assert.NotEqual(t, results1[0].ID, results2[0].ID)
}

func TestListByTimeRange_OrderedByTime(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert prices in random order
	times := []time.Time{
		time.Date(2025, 12, 30, 10, 15, 0, 0, time.UTC),
		time.Date(2025, 12, 30, 10, 5, 0, 0, time.UTC),
		time.Date(2025, 12, 30, 10, 25, 0, 0, time.UTC),
		time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 30, 10, 10, 0, 0, time.UTC),
	}

	publishedAt := time.Now().Add(-5 * time.Minute)

	for i, intervalStart := range times {
		price, err := domain.NewMarketPrice(
			"NSW", 80.0+float64(i), 8000.0,
			domain.Interval5Min, intervalStart, publishedAt,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, price)
		require.NoError(t, err)
	}

	// When
	filter := ports.TimeRangeFilter{
		From: time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	results, err := repo.ListByTimeRange(ctx, filter)

	// Then
	require.NoError(t, err)
	assert.Len(t, results, 5)

	// Verify DESC order (newest first) - compare Unix timestamps to avoid timezone issues
	assert.Equal(t, time.Date(2025, 12, 30, 10, 25, 0, 0, time.UTC).Unix(), results[0].IntervalStart.Unix())
	assert.Equal(t, time.Date(2025, 12, 30, 10, 15, 0, 0, time.UTC).Unix(), results[1].IntervalStart.Unix())
	assert.Equal(t, time.Date(2025, 12, 30, 10, 10, 0, 0, time.UTC).Unix(), results[2].IntervalStart.Unix())
	assert.Equal(t, time.Date(2025, 12, 30, 10, 5, 0, 0, time.UTC).Unix(), results[3].IntervalStart.Unix())
	assert.Equal(t, time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC).Unix(), results[4].IntervalStart.Unix())
}

func TestListByTimeRange_MultipleFilters(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert various prices
	testCases := []struct {
		region       string
		intervalType domain.IntervalType
		minute       int
	}{
		{"NSW", domain.Interval5Min, 0},
		{"NSW", domain.Interval30Min, 30},
		{"VIC", domain.Interval5Min, 5},
		{"NSW", domain.Interval5Min, 10},
	}

	publishedAt := time.Now().Add(-5 * time.Minute)

	for i, tc := range testCases {
		intervalStart := time.Date(2025, 12, 30, 10, tc.minute, 0, 0, time.UTC)
		price, err := domain.NewMarketPrice(
			tc.region, 80.0+float64(i), 8000.0,
			tc.intervalType, intervalStart, publishedAt,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, price)
		require.NoError(t, err)
	}

	// When: Filter by region=NSW AND interval_type=5MIN
	filter := ports.TimeRangeFilter{
		From:         time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
		To:           time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Region:       "NSW",
		IntervalType: domain.Interval5Min,
	}
	results, err := repo.ListByTimeRange(ctx, filter)

	// Then
	require.NoError(t, err)
	assert.Len(t, results, 2) // Only NSW with 5MIN (index 0 and 3)

	for _, result := range results {
		assert.Equal(t, "NSW", result.Region)
		assert.Equal(t, domain.Interval5Min, result.IntervalType)
	}
}
