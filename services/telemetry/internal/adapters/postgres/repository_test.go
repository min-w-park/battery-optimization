package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/minwook/battery-optimization/services/telemetry/internal/domain"
	"github.com/minwook/battery-optimization/services/telemetry/internal/ports"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Connect to telemetry database (port 5434)
	connStr := "postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	return db
}

// cleanupTestDB cleans up test data
func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM battery_states")
	require.NoError(t, err, "Failed to clean up test data")
}

func TestSaveState_Success(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	state, err := domain.NewBatteryState(
		"battery-123",
		75.5,
		25.0,
		28.5,
		800.0,
		31.25,
		domain.OperationStateDischarging,
		map[string]interface{}{
			"vendor": "Tesla",
			"model":  "Megapack",
		},
		time.Now(),
	)
	require.NoError(t, err)

	// When
	err = repo.SaveState(ctx, state)

	// Then
	assert.NoError(t, err)

	// Verify in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM battery_states WHERE id = $1", state.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSaveState_MultipleStates(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	batteryID := "battery-456"
	baseTime := time.Now().Add(-5 * time.Minute)

	// Save multiple states for same battery
	for i := 0; i < 3; i++ {
		state, err := domain.NewBatteryState(
			batteryID,
			float64(50+i*10),
			0.0,
			25.0,
			800.0,
			0.0,
			domain.OperationStateIdle,
			nil,
			baseTime.Add(time.Duration(i)*time.Second),
		)
		require.NoError(t, err)

		err = repo.SaveState(ctx, state)
		require.NoError(t, err)
	}

	// Then: Verify all states saved
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM battery_states WHERE battery_id = $1", batteryID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestGetCurrentState_Success(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	batteryID := "battery-789"
	baseTime := time.Now().Add(-5 * time.Minute)

	// Save 3 states at different times
	var latestState *domain.BatteryState
	for i := 0; i < 3; i++ {
		state, err := domain.NewBatteryState(
			batteryID,
			float64(50+i*10),
			0.0,
			25.0+float64(i),
			800.0,
			0.0,
			domain.OperationStateIdle,
			nil,
			baseTime.Add(time.Duration(i)*time.Minute),
		)
		require.NoError(t, err)
		err = repo.SaveState(ctx, state)
		require.NoError(t, err)

		if i == 2 {
			latestState = state
		}
	}

	// When
	result, err := repo.GetCurrentState(ctx, batteryID)

	// Then
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, latestState.ID, result.ID)
	assert.Equal(t, latestState.SoC, result.SoC)
	assert.Equal(t, latestState.Temperature, result.Temperature)
}

func TestGetCurrentState_NotFound(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// When
	result, err := repo.GetCurrentState(ctx, "nonexistent-battery")

	// Then
	assert.ErrorIs(t, err, domain.ErrBatteryNotFound)
	assert.Nil(t, result)
}

func TestGetHistory_TimeRange(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	batteryID := "battery-history"
	// Use time in the past to avoid timestamp validation errors
	baseTime := time.Now().Add(-10 * time.Minute)

	// Save 5 states at 1-minute intervals
	for i := 0; i < 5; i++ {
		state, err := domain.NewBatteryState(
			batteryID,
			float64(50+i*5),
			0.0,
			25.0,
			800.0,
			0.0,
			domain.OperationStateIdle,
			nil,
			baseTime.Add(time.Duration(i)*time.Minute),
		)
		require.NoError(t, err)
		err = repo.SaveState(ctx, state)
		require.NoError(t, err)
	}

	// When: Query middle 4 states (index 1-4, inclusive range)
	filter := ports.TimeRangeFilter{
		StartTime: baseTime.Add(1 * time.Minute),
		EndTime:   baseTime.Add(4 * time.Minute),
		Limit:     10,
		Offset:    0,
	}
	results, err := repo.GetHistory(ctx, batteryID, filter)

	// Then
	assert.NoError(t, err)
	assert.Len(t, results, 4)
	// Results should be in descending order (latest first)
	assert.Equal(t, 70.0, results[0].SoC) // Index 4
	assert.Equal(t, 65.0, results[1].SoC) // Index 3
	assert.Equal(t, 60.0, results[2].SoC) // Index 2
	assert.Equal(t, 55.0, results[3].SoC) // Index 1
}

func TestGetHistory_Pagination(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	batteryID := "battery-pagination"
	baseTime := time.Now().Add(-5 * time.Minute)

	// Save 10 states
	for i := 0; i < 10; i++ {
		state, err := domain.NewBatteryState(
			batteryID,
			float64(i*10),
			0.0,
			25.0,
			800.0,
			0.0,
			domain.OperationStateIdle,
			nil,
			baseTime.Add(time.Duration(i)*time.Second),
		)
		require.NoError(t, err)
		err = repo.SaveState(ctx, state)
		require.NoError(t, err)
	}

	// When: Get first page (limit 3, offset 0)
	filter := ports.TimeRangeFilter{
		StartTime: baseTime,
		EndTime:   baseTime.Add(20 * time.Second),
		Limit:     3,
		Offset:    0,
	}
	page1, err := repo.GetHistory(ctx, batteryID, filter)

	// Then
	assert.NoError(t, err)
	assert.Len(t, page1, 3)

	// When: Get second page (limit 3, offset 3)
	filter.Offset = 3
	page2, err := repo.GetHistory(ctx, batteryID, filter)

	// Then
	assert.NoError(t, err)
	assert.Len(t, page2, 3)
	// Pages should not overlap
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestGetHistory_EmptyResult(t *testing.T) {
	// Given
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	batteryID := "battery-empty"
	baseTime := time.Now().Add(-30 * time.Minute)

	// Save state at T+0
	state, err := domain.NewBatteryState(
		batteryID,
		50.0,
		0.0,
		25.0,
		800.0,
		0.0,
		domain.OperationStateIdle,
		nil,
		baseTime,
	)
	require.NoError(t, err)
	err = repo.SaveState(ctx, state)
	require.NoError(t, err)

	// When: Query future time range (no data)
	filter := ports.TimeRangeFilter{
		StartTime: baseTime.Add(10 * time.Minute),
		EndTime:   baseTime.Add(20 * time.Minute),
		Limit:     10,
		Offset:    0,
	}
	results, err := repo.GetHistory(ctx, batteryID, filter)

	// Then
	assert.NoError(t, err)
	assert.Empty(t, results)
}
