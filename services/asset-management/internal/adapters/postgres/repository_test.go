package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
	"github.com/minwook/battery-optimization/services/asset-management/internal/ports"
)

const testDatabaseURL = "postgres://asset_user:asset_pass@localhost:5432/asset_management?sslmode=disable"

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testDatabaseURL)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	// Run migrations
	err = runTestMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// runTestMigrations applies the schema
func runTestMigrations(db *sql.DB) error {
	// Create batteries table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS batteries (
			id VARCHAR(36) PRIMARY KEY,
			capacity NUMERIC(10,2) NOT NULL CHECK (capacity > 0),
			max_power NUMERIC(10,2) NOT NULL CHECK (max_power > 0 AND max_power <= capacity),
			ramp_rate NUMERIC(10,2) NOT NULL CHECK (ramp_rate > 0 AND ramp_rate <= max_power),
			efficiency NUMERIC(5,4) NOT NULL CHECK (efficiency >= 0 AND efficiency <= 1),
			location VARCHAR(3) NOT NULL CHECK (location IN ('NSW', 'VIC', 'QLD', 'SA', 'TAS')),
			manufacturer VARCHAR(100) NOT NULL,
			constraints JSONB NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'REGISTERED',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

// cleanupTestDB removes all data from the test database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS batteries")
	require.NoError(t, err, "Failed to cleanup test database")
	db.Close()
}

// createTestBattery creates a valid battery for testing
func createTestBattery(t *testing.T) *domain.Battery {
	constraints := domain.Constraints{
		WarrantyEOL:         0.8,
		MaxCycles:           10000,
		OperatingTempMin:    -10.0,
		OperatingTempMax:    50.0,
		GridComplianceLevel: "AS4777",
	}

	battery, err := domain.NewBattery(200.0, 100.0, 50.0, 0.85, "SA", "Tesla", constraints)
	require.NoError(t, err)
	return battery
}

func TestCreate_Success(t *testing.T) {
	// Given: A test database and repository
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	battery := createTestBattery(t)

	// When: Creating a battery
	err := repo.Create(context.Background(), battery)

	// Then: Should succeed
	require.NoError(t, err)

	// And: Battery should exist in database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM batteries WHERE id = $1", battery.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreate_DuplicateID(t *testing.T) {
	// Given: A test database with an existing battery
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	battery := createTestBattery(t)

	err := repo.Create(context.Background(), battery)
	require.NoError(t, err)

	// When: Creating the same battery again
	err = repo.Create(context.Background(), battery)

	// Then: Should return duplicate error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestFindByID_Success(t *testing.T) {
	// Given: A test database with an existing battery
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)
	original := createTestBattery(t)

	err := repo.Create(context.Background(), original)
	require.NoError(t, err)

	// When: Finding by ID
	found, err := repo.FindByID(context.Background(), original.ID)

	// Then: Should return the battery
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, original.ID, found.ID)
	assert.Equal(t, original.Capacity, found.Capacity)
	assert.Equal(t, original.MaxPower, found.MaxPower)
	assert.Equal(t, original.RampRate, found.RampRate)
	assert.Equal(t, original.Efficiency, found.Efficiency)
	assert.Equal(t, original.Location, found.Location)
	assert.Equal(t, original.Manufacturer, found.Manufacturer)
	assert.Equal(t, original.Status, found.Status)
	assert.Equal(t, original.Constraints.WarrantyEOL, found.Constraints.WarrantyEOL)
	assert.Equal(t, original.Constraints.MaxCycles, found.Constraints.MaxCycles)
}

func TestFindByID_NotFound(t *testing.T) {
	// Given: A test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)

	// When: Finding a non-existent battery
	found, err := repo.FindByID(context.Background(), "non-existent-id")

	// Then: Should return ErrNotFound
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, found)
}

func TestList_NoFilters(t *testing.T) {
	// Given: A test database with multiple batteries
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)

	// Create test batteries
	battery1 := createTestBattery(t)
	battery2 := createTestBattery(t)
	battery2.Location = "NSW"

	err := repo.Create(context.Background(), battery1)
	require.NoError(t, err)
	err = repo.Create(context.Background(), battery2)
	require.NoError(t, err)

	// When: Listing all batteries
	filter := ports.ListFilter{}
	batteries, err := repo.List(context.Background(), filter)

	// Then: Should return all batteries
	require.NoError(t, err)
	assert.Len(t, batteries, 2)
}

func TestList_WithLocationFilter(t *testing.T) {
	// Given: A test database with batteries in different locations
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)

	// Create batteries in different locations
	battery1 := createTestBattery(t) // SA
	battery2 := createTestBattery(t)
	battery2.Location = "NSW"
	battery3 := createTestBattery(t)
	battery3.Location = "VIC"

	require.NoError(t, repo.Create(context.Background(), battery1))
	require.NoError(t, repo.Create(context.Background(), battery2))
	require.NoError(t, repo.Create(context.Background(), battery3))

	// When: Filtering by NSW
	filter := ports.ListFilter{Location: "NSW"}
	batteries, err := repo.List(context.Background(), filter)

	// Then: Should return only NSW battery
	require.NoError(t, err)
	assert.Len(t, batteries, 1)
	assert.Equal(t, "NSW", batteries[0].Location)
}

func TestList_WithStatusFilter(t *testing.T) {
	// Given: A test database with batteries in different statuses
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)

	// Create batteries with different statuses
	battery1 := createTestBattery(t) // REGISTERED
	battery2 := createTestBattery(t)
	battery2.Status = domain.StatusOperational

	// Insert directly with different status
	require.NoError(t, repo.Create(context.Background(), battery1))

	// Update battery2 status in database
	_, err := db.Exec("INSERT INTO batteries (id, capacity, max_power, ramp_rate, efficiency, location, manufacturer, constraints, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		battery2.ID, battery2.Capacity, battery2.MaxPower, battery2.RampRate, battery2.Efficiency,
		battery2.Location, battery2.Manufacturer, `{"warranty_eol": 0.8, "max_cycles": 10000, "operating_temp_min": -10, "operating_temp_max": 50, "grid_compliance_level": "AS4777"}`,
		battery2.Status, battery2.CreatedAt, battery2.UpdatedAt)
	require.NoError(t, err)

	// When: Filtering by OPERATIONAL status
	filter := ports.ListFilter{Status: domain.StatusOperational}
	batteries, err := repo.List(context.Background(), filter)

	// Then: Should return only OPERATIONAL battery
	require.NoError(t, err)
	assert.Len(t, batteries, 1)
	assert.Equal(t, domain.StatusOperational, batteries[0].Status)
}

func TestList_WithPagination(t *testing.T) {
	// Given: A test database with multiple batteries
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)

	// Create 5 batteries
	for i := 0; i < 5; i++ {
		battery := createTestBattery(t)
		battery.Manufacturer = fmt.Sprintf("Manufacturer-%d", i)
		err := repo.Create(context.Background(), battery)
		require.NoError(t, err)
	}

	// When: Requesting with limit and offset
	filter := ports.ListFilter{Limit: 2, Offset: 1}
	batteries, err := repo.List(context.Background(), filter)

	// Then: Should return 2 batteries starting from offset 1
	require.NoError(t, err)
	assert.Len(t, batteries, 2)
}

func TestList_MultipleFilters(t *testing.T) {
	// Given: A test database with various batteries
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewPostgresRepository(db)

	// Create batteries with different attributes
	battery1 := createTestBattery(t) // SA, REGISTERED
	battery2 := createTestBattery(t)
	battery2.Location = "SA"
	battery2.Status = domain.StatusOperational
	battery3 := createTestBattery(t)
	battery3.Location = "NSW"
	battery3.Status = domain.StatusRegistered

	require.NoError(t, repo.Create(context.Background(), battery1))

	// Insert battery2 with OPERATIONAL status
	_, err := db.Exec("INSERT INTO batteries (id, capacity, max_power, ramp_rate, efficiency, location, manufacturer, constraints, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		battery2.ID, battery2.Capacity, battery2.MaxPower, battery2.RampRate, battery2.Efficiency,
		battery2.Location, battery2.Manufacturer, `{"warranty_eol": 0.8, "max_cycles": 10000, "operating_temp_min": -10, "operating_temp_max": 50, "grid_compliance_level": "AS4777"}`,
		battery2.Status, battery2.CreatedAt, battery2.UpdatedAt)
	require.NoError(t, err)

	require.NoError(t, repo.Create(context.Background(), battery3))

	// When: Filtering by SA location AND OPERATIONAL status
	filter := ports.ListFilter{Location: "SA", Status: domain.StatusOperational}
	batteries, err := repo.List(context.Background(), filter)

	// Then: Should return only SA + OPERATIONAL battery
	require.NoError(t, err)
	assert.Len(t, batteries, 1)
	assert.Equal(t, "SA", batteries[0].Location)
	assert.Equal(t, domain.StatusOperational, batteries[0].Status)
}
