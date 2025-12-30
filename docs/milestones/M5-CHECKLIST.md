# M5: Implementation Checklist

## 📋 How to Use This Checklist

**⚠️ CRITICAL: READ THESE FIRST**:
1. [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development philosophy and TDD workflow (Kent Beck style)
2. [M5-OVERVIEW.md](M5-OVERVIEW.md) - Hardware abstraction and real-time telemetry architecture
3. [M5-DOMAIN-SPEC.md](M5-DOMAIN-SPEC.md) - BatteryState aggregate, BatteryAdapter interface, validation rules
4. [M5-API-SPEC.md](M5-API-SPEC.md) - REST endpoints, event schemas, 1 Hz publishing patterns

**Then follow this checklist**:
1. Work through tasks in order (top to bottom)
2. **Write tests BEFORE implementation** (Red → Green → Refactor)
3. Check off `[ ]` boxes as you complete each task
4. Each phase should take 2-4 hours
5. If stuck, refer to detailed spec documents
6. Commit after each major milestone

---

## Phase 0: M5 Milestone Documentation (1-1.5 hours) ✅ COMPLETE

### 0.1 Documentation Files
- [x] Create `docs/milestones/M5-OVERVIEW.md`
- [x] Create `docs/milestones/M5-DOMAIN-SPEC.md`
- [x] Create `docs/milestones/M5-API-SPEC.md`
- [x] Create `docs/milestones/M5-CHECKLIST.md`

### 0.2 Review and Commit
- [ ] Review all 4 documentation files
- [ ] Verify event schemas match EVENTS.md
- [ ] Verify database connection matches docker-compose.yml (telemetry-db on 5434)
- [ ] Commit documentation before starting Phase 1
  ```bash
  git add docs/milestones/M5-*.md
  git commit -m "docs(M5): add Telemetry + Device Interface milestone documentation"
  ```

**Checkpoint**: ✅ Documentation complete, ready to start implementation

---

## Phase 1: Telemetry Service - Domain Layer (2-3 hours)

### 1.1 Project Setup
- [ ] Create service directory structure:
  ```bash
  mkdir -p services/telemetry/cmd/server
  mkdir -p services/telemetry/internal/domain
  mkdir -p services/telemetry/internal/ports
  mkdir -p services/telemetry/internal/adapters/postgres/migrations
  mkdir -p services/telemetry/internal/adapters/http
  mkdir -p services/telemetry/internal/service
  ```

- [ ] Initialize Go module:
  ```bash
  cd services/telemetry
  go mod init github.com/minwook/battery-optimization/services/telemetry
  ```

### 1.2 Domain Errors (TDD - Tests First!)

#### Write Tests FIRST (Red Phase)
- [ ] Create `internal/domain/errors_test.go`
- [ ] Write test: `TestDomainErrors_Unwrap`
  - Verify errors implement error interface
  - Test error messages
- [ ] Run tests → Should FAIL ✅

#### Implementation (Green Phase)
- [ ] Create `internal/domain/errors.go`
- [ ] Define domain errors:
  ```go
  var (
      ErrInvalidSoC              = errors.New("invalid SoC value")
      ErrInvalidTemperature      = errors.New("invalid temperature value")
      ErrInvalidOperationState   = errors.New("invalid operation state")
      ErrInconsistentPowerState  = errors.New("power value inconsistent with operation state")
      ErrMissingRequiredField    = errors.New("required field is missing")
      ErrInvalidTimestamp        = errors.New("timestamp is invalid")
      ErrBatteryNotFound         = errors.New("battery not found")
  )
  ```
- [ ] Run tests → All pass ✅

### 1.3 BatteryState Aggregate (TDD - Tests First!)

#### Write Tests FIRST (Red Phase)
- [ ] Create `internal/domain/battery_state_test.go`
- [ ] Write test: `TestNewBatteryState_ValidInput`
  - Create BatteryState with all valid fields
  - Verify ID is generated (UUID)
  - Verify CreatedAt is set
  - Verify all fields match input
- [ ] Write test: `TestNewBatteryState_InvalidSoC`
  - Test SoC < 0
  - Test SoC > 100
  - Verify returns ErrInvalidSoC
- [ ] Write test: `TestNewBatteryState_InvalidTemperature`
  - Test temperature < -20°C
  - Test temperature > 60°C
  - Verify returns ErrInvalidTemperature
- [ ] Write test: `TestNewBatteryState_InvalidOperationState`
  - Test invalid enum values
  - Verify returns ErrInvalidOperationState
- [ ] Write test: `TestNewBatteryState_PowerConsistency`
  - IDLE with non-zero power → error
  - CHARGING with positive power → error
  - DISCHARGING with negative power → error
  - Verify returns ErrInconsistentPowerState
- [ ] Write test: `TestNewBatteryState_MissingRequiredFields`
  - Empty BatteryID → error
  - Zero timestamp → error
- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All BatteryState tests written and failing

#### Implementation (Green Phase)
- [ ] Create `internal/domain/battery_state.go`
- [ ] Define BatteryState struct:
  ```go
  type BatteryState struct {
      ID               string
      BatteryID        string
      SoC              float64
      Power            float64
      Temperature      float64
      Voltage          float64
      Current          float64
      OperationState   string
      CustomAttributes map[string]interface{}
      Timestamp        time.Time
      CreatedAt        time.Time
  }
  ```

- [ ] Define operation state constants:
  ```go
  const (
      OperationStateIdle        = "IDLE"
      OperationStateCharging    = "CHARGING"
      OperationStateDischarging = "DISCHARGING"
      OperationStateFCAS        = "FCAS"
  )
  ```

- [ ] Implement `NewBatteryState()` constructor with validation:
  - Generate UUID for ID
  - Set CreatedAt to time.Now()
  - Validate SoC (0-100)
  - Validate Temperature (-20 to 60)
  - Validate OperationState enum
  - Validate power consistency with state
  - Validate required fields (BatteryID, Timestamp)
  - Validate timestamp not in future

- [ ] Implement `Validate()` method

- [ ] Run tests → All pass ✅ (Green phase)

### 1.4 Test Coverage
- [ ] Run: `go test -cover ./internal/domain/...`
- [ ] Verify coverage >90% (target: 95%+)
- [ ] All validation tests passing

**Checkpoint**: ✅ Domain layer complete with comprehensive validation

**Reference Files**:
- `services/asset-management/internal/domain/battery.go`
- `services/market-data/internal/domain/price.go`

---

## Phase 2: Telemetry Service - Repository Layer (2-3 hours)

### 2.1 Repository Port (Interface)

- [ ] Create `internal/ports/repository.go`
- [ ] Define TelemetryRepository interface:
  ```go
  type TelemetryRepository interface {
      SaveState(ctx context.Context, state *domain.BatteryState) error
      GetCurrentState(ctx context.Context, batteryID string) (*domain.BatteryState, error)
      GetHistory(ctx context.Context, batteryID string, filter TimeRangeFilter) ([]*domain.BatteryState, error)
  }

  type TimeRangeFilter struct {
      StartTime time.Time
      EndTime   time.Time
      Limit     int
      Offset    int
  }
  ```

### 2.2 Database Migrations

- [ ] Create `internal/adapters/postgres/migrations/000001_create_battery_states.up.sql`:
  ```sql
  CREATE TABLE IF NOT EXISTS battery_states (
      id VARCHAR(36) PRIMARY KEY,
      battery_id VARCHAR(36) NOT NULL,
      soc NUMERIC(5,2) NOT NULL CHECK (soc >= 0 AND soc <= 100),
      power NUMERIC(10,2) NOT NULL,
      temperature NUMERIC(5,2) NOT NULL,
      voltage NUMERIC(10,2) NOT NULL,
      current NUMERIC(10,2) NOT NULL,
      operation_state VARCHAR(20) NOT NULL CHECK (operation_state IN ('IDLE', 'CHARGING', 'DISCHARGING', 'FCAS')),
      custom_attributes JSONB,
      timestamp TIMESTAMP NOT NULL,
      created_at TIMESTAMP NOT NULL DEFAULT NOW()
  );

  CREATE INDEX idx_battery_states_battery_id_timestamp ON battery_states(battery_id, timestamp DESC);
  CREATE INDEX idx_battery_states_timestamp ON battery_states(timestamp DESC);
  CREATE INDEX idx_battery_states_operation_state ON battery_states(operation_state);
  ```

- [ ] Create `internal/adapters/postgres/migrations/000001_create_battery_states.down.sql`:
  ```sql
  DROP INDEX IF EXISTS idx_battery_states_operation_state;
  DROP INDEX IF EXISTS idx_battery_states_timestamp;
  DROP INDEX IF EXISTS idx_battery_states_battery_id_timestamp;
  DROP TABLE IF EXISTS battery_states;
  ```

### 2.3 PostgreSQL Repository Tests (TDD - Tests First!)

- [ ] Create `internal/adapters/postgres/repository_test.go`
- [ ] Write test helper: `setupTestDB(t *testing.T) *sql.DB`
  - Connect to telemetry-db: `postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable`
  - Run migrations
  - Return DB connection
- [ ] Write test helper: `cleanupTestDB(t *testing.T, db *sql.DB)`
  - Drop tables
  - Close connection

- [ ] Write test: `TestSaveState_Success`
  - Create valid BatteryState
  - Save to repository
  - Verify no error
  - Verify can retrieve by ID

- [ ] Write test: `TestGetCurrentState_Success`
  - Save multiple states for same battery (different timestamps)
  - Call GetCurrentState
  - Verify returns latest state (highest timestamp)

- [ ] Write test: `TestGetCurrentState_NotFound`
  - Call GetCurrentState for non-existent battery
  - Verify returns ErrBatteryNotFound

- [ ] Write test: `TestGetHistory_TimeRange`
  - Save 10 states across 1 hour
  - Query with startTime/endTime
  - Verify correct states returned
  - Verify ordered by timestamp DESC

- [ ] Write test: `TestGetHistory_Pagination`
  - Save 100 states
  - Query with limit=10, offset=20
  - Verify returns states 21-30

- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All repository tests written and failing

### 2.4 PostgreSQL Repository Implementation (Green Phase)

- [ ] Create `internal/adapters/postgres/repository.go`
- [ ] Implement `PostgresRepository` struct:
  ```go
  type PostgresRepository struct {
      db *sql.DB
  }

  func NewPostgresRepository(db *sql.DB) *PostgresRepository {
      return &PostgresRepository{db: db}
  }
  ```

- [ ] Implement `SaveState`:
  - Marshal CustomAttributes to JSON
  - INSERT statement
  - Error handling (duplicate key, constraint violation)

- [ ] Implement `GetCurrentState`:
  - SELECT with ORDER BY timestamp DESC LIMIT 1
  - Unmarshal CustomAttributes from JSONB
  - Map SQL NULL to nil
  - Return ErrBatteryNotFound if no rows

- [ ] Implement `GetHistory`:
  - SELECT with WHERE battery_id = $1 AND timestamp BETWEEN $2 AND $3
  - ORDER BY timestamp DESC
  - LIMIT and OFFSET for pagination
  - Scan all rows into []*BatteryState

- [ ] Run tests → All pass ✅ (Green phase)

### 2.5 Test Coverage
- [ ] Run: `go test -cover ./internal/adapters/postgres/...`
- [ ] Verify coverage >85% (target: 90%+)

**Checkpoint**: ✅ Repository layer complete with time-series optimization

**Reference Files**:
- `services/asset-management/internal/adapters/postgres/repository.go`
- `services/market-data/internal/adapters/postgres/repository.go`

---

## Phase 3: Telemetry Service - HTTP API Layer (2-3 hours)

### 3.1 DTOs (Data Transfer Objects)

- [ ] Create `internal/adapters/http/dto.go`
- [ ] Define request/response DTOs:
  ```go
  type BatteryStateResponse struct {
      ID               string                 `json:"id"`
      BatteryID        string                 `json:"battery_id"`
      SoC              float64                `json:"soc"`
      Power            float64                `json:"power"`
      Temperature      float64                `json:"temperature"`
      Voltage          float64                `json:"voltage"`
      Current          float64                `json:"current"`
      OperationState   string                 `json:"operation_state"`
      CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
      Timestamp        time.Time              `json:"timestamp"`
      CreatedAt        time.Time              `json:"created_at"`
  }

  type BatteryStateHistoryResponse struct {
      Data       []BatteryStateResponse `json:"data"`
      Pagination PaginationInfo         `json:"pagination"`
  }

  type PaginationInfo struct {
      Total  int `json:"total"`
      Limit  int `json:"limit"`
      Offset int `json:"offset"`
  }

  type ErrorResponse struct {
      Error   string                 `json:"error"`
      Message string                 `json:"message"`
      Details map[string]interface{} `json:"details,omitempty"`
  }
  ```

- [ ] Implement mapping functions:
  - `toStateResponse(state *domain.BatteryState) BatteryStateResponse`

### 3.2 HTTP Handler Tests (TDD - Tests First!)

- [ ] Create `internal/adapters/http/handler_test.go`
- [ ] Create mock repository:
  ```go
  type MockRepository struct {
      SaveStateFunc       func(ctx context.Context, state *domain.BatteryState) error
      GetCurrentStateFunc func(ctx context.Context, batteryID string) (*domain.BatteryState, error)
      GetHistoryFunc      func(ctx context.Context, batteryID string, filter ports.TimeRangeFilter) ([]*domain.BatteryState, error)
  }
  ```

- [ ] Write test: `TestGetCurrentState_Success`
  - Setup mock to return test state
  - Create HTTP request: `GET /api/v1/telemetry/{batteryId}/current`
  - Call handler
  - Verify 200 OK
  - Verify JSON response matches expected

- [ ] Write test: `TestGetCurrentState_NotFound`
  - Setup mock to return ErrBatteryNotFound
  - Call handler
  - Verify 404 Not Found
  - Verify error response format

- [ ] Write test: `TestGetCurrentState_InvalidBatteryID`
  - Request with invalid UUID
  - Verify 400 Bad Request

- [ ] Write test: `TestGetHistory_Success`
  - Setup mock to return array of states
  - Request with query params: `?startTime=...&endTime=...&limit=10`
  - Verify 200 OK
  - Verify pagination info correct

- [ ] Write test: `TestGetHistory_InvalidTimeRange`
  - Request with startTime > endTime
  - Verify 400 Bad Request

- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All handler tests written and failing

### 3.3 HTTP Handler Implementation (Green Phase)

- [ ] Create `internal/adapters/http/handler.go`
- [ ] Implement `TelemetryHandler` struct:
  ```go
  type TelemetryHandler struct {
      repo ports.TelemetryRepository
  }

  func NewTelemetryHandler(repo ports.TelemetryRepository) *TelemetryHandler {
      return &TelemetryHandler{repo: repo}
  }
  ```

- [ ] Implement `GetCurrentState`:
  - Parse batteryId from URL
  - Validate UUID format
  - Call repo.GetCurrentState()
  - Handle errors (404 for not found, 500 for other)
  - Return JSON response

- [ ] Implement `GetHistory`:
  - Parse batteryId from URL
  - Parse query params (startTime, endTime, limit, offset)
  - Validate time range
  - Call repo.GetHistory()
  - Build pagination info
  - Return JSON response

- [ ] Implement error handling helper:
  ```go
  func writeError(w http.ResponseWriter, code int, err string, msg string)
  ```

- [ ] Run tests → All pass ✅ (Green phase)

### 3.4 Routes Setup

- [ ] Create `internal/adapters/http/routes.go`
- [ ] Setup gorilla/mux router:
  ```go
  func SetupRoutes(handler *TelemetryHandler) *mux.Router {
      r := mux.NewRouter()

      // API routes
      api := r.PathPrefix("/api/v1").Subrouter()
      api.HandleFunc("/telemetry/{batteryId}/current", handler.GetCurrentState).Methods("GET")
      api.HandleFunc("/telemetry/{batteryId}/history", handler.GetHistory).Methods("GET")

      // Health check
      r.HandleFunc("/health", handler.HealthCheck).Methods("GET")

      return r
  }
  ```

- [ ] Implement `HealthCheck` handler

### 3.5 Test Coverage
- [ ] Run: `go test -cover ./internal/adapters/http/...`
- [ ] Verify coverage >65% (target: 70%+)

**Checkpoint**: ✅ HTTP API layer complete

**Reference Files**:
- `services/asset-management/internal/adapters/http/handler.go`
- `services/market-data/internal/adapters/http/handler.go`

---

## Phase 4: Telemetry Service - Event Publishing (2-3 hours)

### 4.1 Update pkg/events (if needed)

- [ ] Check `pkg/events/battery_state_changed.go`
- [ ] Remove placeholder comment if present
- [ ] Verify all fields match M5-DOMAIN-SPEC.md:
  - battery_id, soc, power, temperature, voltage, current
  - operation_state, custom_attributes
  - timestamp, event_version

- [ ] Add CustomAttributes field if missing:
  ```go
  type BatteryStateChanged struct {
      BatteryID        string                 `json:"battery_id"`
      SoC              float64                `json:"soc"`
      Power            float64                `json:"power"`
      Temperature      float64                `json:"temperature"`
      Voltage          float64                `json:"voltage"`
      Current          float64                `json:"current"`
      OperationState   string                 `json:"operation_state"`
      CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
      Timestamp        time.Time              `json:"timestamp"`
      EventVersion     string                 `json:"event_version"`
  }
  ```

- [ ] Run: `cd pkg/events && go test -v ./...`
- [ ] Verify all tests still pass

### 4.2 State Publisher Service

- [ ] Create `internal/service/state_publisher.go`
- [ ] Implement `StatePublisher`:
  ```go
  type StatePublisher struct {
      repo      ports.TelemetryRepository
      publisher events.EventPublisher
      batteryID string
  }

  func NewStatePublisher(
      repo ports.TelemetryRepository,
      publisher events.EventPublisher,
      batteryID string,
  ) *StatePublisher {
      return &StatePublisher{
          repo:      repo,
          publisher: publisher,
          batteryID: batteryID,
      }
  }
  ```

- [ ] Implement `Start(ctx context.Context)` method:
  - Create 1 Hz ticker
  - Loop: get current state → publish event
  - Context cancellation support
  - Best-effort error logging
  - 2-second timeout on publish

- [ ] Write tests for StatePublisher:
  - Test ticker fires at 1 Hz
  - Test publishes BatteryStateChanged
  - Test context cancellation
  - Test error handling (don't crash)

### 4.3 Main Application Setup

- [ ] Create `cmd/server/main.go`
- [ ] Implement configuration:
  ```go
  type Config struct {
      DatabaseURL string
      NatsURL     string
      Port        string
      LogLevel    string
  }

  func loadConfig() Config {
      return Config{
          DatabaseURL: getEnv("DATABASE_URL", "postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable"),
          NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
          Port:        getEnv("PORT", "8082"),
          LogLevel:    getEnv("LOG_LEVEL", "info"),
      }
  }
  ```

- [ ] Implement startup sequence:
  1. Load config
  2. Setup logging
  3. Connect to PostgreSQL
  4. Configure connection pool (MaxOpenConns=25, MaxIdleConns=5)
  5. Ping database
  6. Run migrations
  7. Create repository
  8. Connect to NATS (optional - graceful degradation)
  9. Create handler
  10. Setup HTTP routes
  11. **Start State Publisher goroutine** (new for M5)
  12. Start HTTP server
  13. Graceful shutdown (30s timeout)

- [ ] Add State Publisher startup:
  ```go
  // After creating repository and publisher
  if publisher != nil {
      // TODO: Get list of all batteries from Asset Management
      // For now, hardcode test battery
      batteryID := "test-battery-123"

      statePublisher := service.NewStatePublisher(repo, publisher, batteryID)
      go statePublisher.Start(context.Background())
      log.Printf("State publisher started for battery %s (1 Hz)", batteryID)
  }
  ```

### 4.4 Dependencies

- [ ] Update `go.mod`:
  ```bash
  cd services/telemetry
  go get github.com/lib/pq
  go get github.com/gorilla/mux
  go get github.com/golang-migrate/migrate/v4
  go get github.com/google/uuid
  ```

- [ ] Add local replace directive:
  ```go
  replace github.com/minwook/battery-optimization/pkg/events => ../../pkg/events
  ```

- [ ] Run: `go mod tidy`

### 4.5 Manual Testing

- [ ] Start infrastructure: `docker-compose up -d`
- [ ] Verify telemetry-db: `docker exec telemetry-db psql -U telemetry_user -d telemetry -c "SELECT version();"`
- [ ] Start Telemetry Service: `cd services/telemetry && go run cmd/server/main.go`
- [ ] Check logs for "State publisher started"
- [ ] Start event subscriber: `cd tools/event-subscriber && go run main.go "battery.>"`
- [ ] Verify seeing `battery.state.changed.v1` events every 1 second

**Checkpoint**: ✅ BatteryStateChanged events publishing at 1 Hz

**Reference Files**:
- `services/asset-management/cmd/server/main.go`
- `services/market-data/cmd/server/main.go`
- `pkg/events/README.md`

---

## Phase 5: Device Interface Service - Domain & Adapters (3-4 hours)

### 5.1 Project Setup

- [ ] Create service directory structure:
  ```bash
  mkdir -p services/device-interface/cmd/server
  mkdir -p services/device-interface/internal/domain
  mkdir -p services/device-interface/internal/adapters
  mkdir -p services/device-interface/internal/service
  ```

- [ ] Initialize Go module:
  ```bash
  cd services/device-interface
  go mod init github.com/minwook/battery-optimization/services/device-interface
  ```

### 5.2 Domain Layer (BatteryAdapter Interface)

- [ ] Create `internal/domain/battery_adapter.go`
- [ ] Define BatteryAdapter interface:
  ```go
  type BatteryAdapter interface {
      GetState(ctx context.Context) (BatteryState, error)
      SendCommand(ctx context.Context, cmd Command) error
      GetCustomAttributes() map[string]interface{}
      GetBatteryID() string
  }

  type BatteryState struct {
      SoC            float64
      Power          float64
      Temperature    float64
      Voltage        float64
      Current        float64
      OperationState string
  }

  type CommandType int

  const (
      CommandCharge CommandType = iota
      CommandDischarge
      CommandIdle
      CommandFcasResponse
  )

  type Command struct {
      Type       CommandType
      Power      float64
      TargetSoC  float64
      Duration   time.Duration
  }
  ```

### 5.3 TeslaLike Adapter (TDD - Tests First!)

#### Write Tests FIRST (Red Phase)
- [ ] Create `internal/adapters/teslalike_test.go`
- [ ] Write test: `TestTeslaLike_GetState`
  - Create adapter
  - Call GetState
  - Verify returns BatteryState with initial values

- [ ] Write test: `TestTeslaLike_SendCommand_Charge`
  - Send charge command (15 MW)
  - Wait 5 seconds
  - Verify SoC increased
  - Verify Power is negative (charging)
  - Verify OperationState is CHARGING

- [ ] Write test: `TestTeslaLike_SendCommand_Discharge`
  - Send discharge command (25 MW)
  - Wait 5 seconds
  - Verify SoC decreased
  - Verify Power is positive (discharging)

- [ ] Write test: `TestTeslaLike_RampRate`
  - Send command with 50 MW
  - Check power after 1 second
  - Verify power < 50 (ramping up at 5 MW/s)
  - Check after 10 seconds
  - Verify power == 50 (fully ramped)

- [ ] Write test: `TestTeslaLike_TemperatureSimulation`
  - Send charge command
  - Check temperature every second for 10 seconds
  - Verify temperature increases during charging
  - Send idle command
  - Verify temperature decreases

- [ ] Write test: `TestTeslaLike_CustomAttributes`
  - Call GetCustomAttributes
  - Verify contains "vendor": "Tesla"
  - Verify contains "model": "Megapack"

- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All TeslaLike tests written and failing

#### Implementation (Green Phase)
- [ ] Create `internal/adapters/teslalike.go`
- [ ] Implement `TeslaLikeAdapter`:
  ```go
  type TeslaLikeAdapter struct {
      batteryID       string
      capacity        float64
      maxPower        float64
      state           domain.BatteryState
      targetPower     float64
      targetSoC       float64
      commandDuration time.Duration
      commandStartTime time.Time
      mu              sync.RWMutex
      cancel          context.CancelFunc
  }
  ```

- [ ] Implement constructor:
  ```go
  func NewTeslaLikeAdapter(batteryID string, capacity float64, maxPower float64, initialSoC float64) *TeslaLikeAdapter
  ```

- [ ] Implement `GetState()`:
  - Lock mutex (RLock)
  - Return copy of current state

- [ ] Implement `SendCommand()`:
  - Lock mutex
  - Update target power and SoC
  - Start background simulation goroutine if not running

- [ ] Implement background simulation loop:
  - Ticker: 100ms (10 Hz internal simulation)
  - Each tick:
    - Apply ramp rate (5 MW/s for Tesla)
    - Update SoC based on power
    - Update temperature (0.5°C per MW)
    - Check stop conditions (target SoC reached, duration exceeded)
    - Update OperationState

- [ ] Implement `GetCustomAttributes()`:
  ```go
  return map[string]interface{}{
      "vendor": "Tesla",
      "model": "Megapack",
      "firmwareVersion": "1.2.3",
      "cellCount": 4320,
      "thermalZones": 3,
  }
  ```

- [ ] Run tests → All pass ✅ (Green phase)

### 5.4 BYDLike Adapter (Similar to TeslaLike)

- [ ] Create `internal/adapters/bydlike_test.go`
- [ ] Write similar tests as TeslaLike but verify different behaviors:
  - Ramp rate: 3 MW/s (slower than Tesla's 5 MW/s)
  - Temperature rise: 0.7°C per MW (higher than Tesla's 0.5°C)
  - Efficiency: 92% (lower than Tesla's 95%)

- [ ] Run tests → Should FAIL (Red phase) ✅

- [ ] Create `internal/adapters/bydlike.go`
- [ ] Implement with different constants:
  ```go
  const (
      BYDRampRate      = 3.0   // MW/s
      BYDEfficiency    = 0.92  // 92%
      BYDTempRiseRate  = 0.7   // °C per MW
      BYDCoolingRate   = 0.08  // °C per tick when idle
  )
  ```

- [ ] Implement GetCustomAttributes with BYD-specific data:
  ```go
  return map[string]interface{}{
      "vendor": "BYD",
      "model": "Battery-Box",
      "firmwareVersion": "2.0.1",
      "cellCount": 3840,
      "batteryChemistry": "LFP",
  }
  ```

- [ ] Run tests → All pass ✅

### 5.5 Test Coverage
- [ ] Run: `go test -cover ./internal/adapters/...`
- [ ] Verify coverage >85% (target: 90%+)

**Checkpoint**: ✅ Two swappable battery adapters with realistic simulation

**Reference Files**:
- M5-DOMAIN-SPEC.md (adapter specifications)

---

## Phase 6: Device Interface Service - Event Handling (2-3 hours)

### 6.1 Add Missing Events to pkg/events

Check if these events exist in pkg/events, create if missing:

- [ ] `charging_command_issued.go`:
  ```go
  type ChargingCommandIssued struct {
      CommandID       string    `json:"command_id"`
      BatteryID       string    `json:"battery_id"`
      Power           float64   `json:"power"`
      TargetSoC       float64   `json:"target_soc"`
      DurationMinutes int       `json:"duration_minutes"`
      DecisionMode    string    `json:"decision_mode"`
      Timestamp       time.Time `json:"timestamp"`
      EventVersion    string    `json:"event_version"`
  }
  ```

- [ ] `charging_started.go`, `charging_completed.go`
- [ ] `discharging_command_issued.go`
- [ ] `discharging_started.go`, `discharging_completed.go`

- [ ] Write tests for all new events (JSON serialization)
- [ ] Run: `cd pkg/events && go test -v ./...`

### 6.2 Command Handler Service (TDD - Tests First!)

- [ ] Create `internal/service/command_handler_test.go`
- [ ] Write test: `TestHandleChargingCommand_Success`
  - Mock adapter and publisher
  - Create ChargingCommandIssued event
  - Call HandleChargingCommand
  - Verify adapter.SendCommand called
  - Verify ChargingStarted event published

- [ ] Write test: `TestHandleChargingCommand_InvalidPower`
  - Command with negative power
  - Verify returns error
  - Verify no event published

- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ Command handler tests written and failing

- [ ] Create `internal/service/command_handler.go`
- [ ] Implement `CommandHandler`:
  ```go
  type CommandHandler struct {
      adapter   domain.BatteryAdapter
      publisher events.EventPublisher
  }

  func NewCommandHandler(adapter domain.BatteryAdapter, publisher events.EventPublisher) *CommandHandler
  ```

- [ ] Implement `HandleChargingCommand`:
  - Validate command fields
  - Create domain.Command
  - Call adapter.SendCommand()
  - Publish ChargingStarted event
  - Handle errors

- [ ] Implement `HandleDischargingCommand` (similar pattern)

- [ ] Run tests → All pass ✅

### 6.3 Main Application Setup

- [ ] Create `cmd/server/main.go`
- [ ] Implement configuration:
  ```go
  type Config struct {
      NatsURL     string
      BatteryID   string  // Which battery this adapter controls
      AdapterType string  // "tesla" or "byd"
      Capacity    float64
      MaxPower    float64
      InitialSoC  float64
  }
  ```

- [ ] Implement startup sequence:
  1. Load config
  2. Create adapter (TeslaLike or BYDLike based on config)
  3. Connect to NATS
  4. Create publisher
  5. Create subscriber
  6. Create command handler
  7. Subscribe to command events
  8. Publish BatteryConnectionEstablished on startup
  9. Wait for signals (graceful shutdown)

- [ ] Implement event subscription:
  ```go
  handler := func(subject string, data []byte) error {
      switch subject {
      case "charging.command.issued.v1":
          var event events.ChargingCommandIssued
          json.Unmarshal(data, &event)
          return cmdHandler.HandleChargingCommand(event)

      case "discharging.command.issued.v1":
          var event events.DischargingCommandIssued
          json.Unmarshal(data, &event)
          return cmdHandler.HandleDischargingCommand(event)
      }
      return nil
  }

  subscriber.Subscribe(ctx, "*.command.issued.v1", handler)
  ```

### 6.4 Dependencies

- [ ] Update `go.mod`:
  ```bash
  go get github.com/google/uuid
  ```

- [ ] Add local replace directive for pkg/events

- [ ] Run: `go mod tidy`

**Checkpoint**: ✅ Device Interface reacts to commands and publishes lifecycle events

**Reference Files**:
- `pkg/events/README.md` (subscription patterns)
- `services/market-data/cmd/server/main.go` (NATS setup)

---

## Phase 7: Integration Testing (2-3 hours)

### 7.1 Infrastructure Verification

- [ ] Start all infrastructure:
  ```bash
  docker-compose up -d
  ```

- [ ] Verify NATS health:
  ```bash
  curl http://localhost:8222/healthz
  # Expected: {"status":"ok"}
  ```

- [ ] Verify telemetry-db:
  ```bash
  docker exec telemetry-db psql -U telemetry_user -d telemetry -c "SELECT version();"
  # Expected: PostgreSQL 18.x
  ```

- [ ] Check all 3 databases running:
  ```bash
  docker-compose ps
  # asset-db (5432), market-db (5433), telemetry-db (5434)
  ```

### 7.2 Build All Services

- [ ] Build Telemetry Service:
  ```bash
  cd services/telemetry
  go build -o telemetry cmd/server/main.go
  ```

- [ ] Build Device Interface Service:
  ```bash
  cd services/device-interface
  go build -o device-interface cmd/server/main.go
  ```

- [ ] Verify no build errors

### 7.3 End-to-End Event Flow

**Terminal 1: Event Subscriber**
- [ ] Start event subscriber:
  ```bash
  cd tools/event-subscriber
  go run main.go "battery.>"
  ```

**Terminal 2: Telemetry Service**
- [ ] Start Telemetry Service:
  ```bash
  cd services/telemetry
  export DATABASE_URL="postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable"
  export NATS_URL="nats://localhost:4222"
  export PORT="8082"
  go run cmd/server/main.go
  ```
- [ ] Verify logs show "State publisher started"

**Terminal 3: Device Interface Service**
- [ ] Start Device Interface:
  ```bash
  cd services/device-interface
  export NATS_URL="nats://localhost:4222"
  export BATTERY_ID="test-battery-123"
  export ADAPTER_TYPE="tesla"
  export CAPACITY="100"
  export MAX_POWER="50"
  export INITIAL_SOC="50"
  go run cmd/server/main.go
  ```
- [ ] Verify logs show "Battery adapter initialized"

**Verification**:
- [ ] Event subscriber shows `battery.state.changed.v1` every 1 second
- [ ] Events show realistic state (SoC, power, temperature)
- [ ] No errors in service logs

### 7.4 REST API Testing

- [ ] Test health check:
  ```bash
  curl http://localhost:8082/health
  ```

- [ ] Test current state (after telemetry data exists):
  ```bash
  curl http://localhost:8082/api/v1/telemetry/test-battery-123/current
  ```
  - Verify 200 OK response
  - Verify JSON contains soc, power, temperature fields

- [ ] Test history query:
  ```bash
  curl "http://localhost:8082/api/v1/telemetry/test-battery-123/history?limit=10"
  ```
  - Verify returns array of states
  - Verify pagination info

- [ ] Test not found:
  ```bash
  curl http://localhost:8082/api/v1/telemetry/nonexistent-battery/current
  ```
  - Verify 404 Not Found

### 7.5 Test Coverage Verification

- [ ] Telemetry Service coverage:
  ```bash
  cd services/telemetry
  go test -cover ./...
  ```
  - Target: >80% overall
  - Domain: >90%

- [ ] Device Interface Service coverage:
  ```bash
  cd services/device-interface
  go test -cover ./...
  ```
  - Target: >80% overall
  - Adapters: >85%

- [ ] pkg/events coverage (should still be ~89%):
  ```bash
  cd pkg/events
  go test -cover ./...
  ```

### 7.6 Performance Testing

- [ ] Monitor 1 Hz event publishing for 5 minutes
- [ ] Verify stable event rate (no slowdown)
- [ ] Check NATS stats:
  ```bash
  curl http://localhost:8222/varz | jq '.in_msgs, .out_msgs, .in_bytes, .out_bytes'
  ```

- [ ] Verify no memory leaks:
  - Check service memory usage (should be stable <100MB each)
  - Run for 10 minutes, verify memory doesn't grow

**Checkpoint**: ✅ Full event-driven flow working end-to-end

---

## Phase 8: Polish & Documentation (1-2 hours)

### 8.1 Code Quality

- [ ] Run `go fmt` on all packages:
  ```bash
  cd services/telemetry && go fmt ./...
  cd services/device-interface && go fmt ./...
  cd pkg/events && go fmt ./...
  ```

- [ ] Run `go vet` on all packages:
  ```bash
  cd services/telemetry && go vet ./...
  cd services/device-interface && go vet ./...
  cd pkg/events && go vet ./...
  ```

- [ ] Fix any warnings

### 8.2 Service Documentation

- [ ] Create `services/telemetry/README.md`
  - Service purpose and architecture
  - API endpoints with examples
  - Event publishing (1 Hz BatteryStateChanged)
  - Database schema
  - Development commands (build, test, run)
  - Configuration (environment variables)

- [ ] Create `services/device-interface/README.md`
  - Service purpose (hardware abstraction)
  - BatteryAdapter interface documentation
  - Mock adapter descriptions (TeslaLike vs BYDLike)
  - Event subscriptions (charging/discharging commands)
  - Event publishing (started/completed events)
  - CustomAttributes handling
  - Configuration (adapter type, battery specs)

### 8.3 Update Project Documentation

- [ ] Update `PLANNING.md`:
  - Mark M5 as complete
  - Add phase breakdown (0-8)
  - Add test coverage metrics
  - Add key achievements
  - Update "Next Up" to M6

- [ ] Update `CLAUDE.md`:
  - Add M5 services to project status
  - Update Services Running section (add Telemetry:8082, Device Interface:8083)
  - Update completed milestones count

- [ ] Update `README.md`:
  - Add M5 to completed milestones
  - Add Telemetry and Device Interface service descriptions
  - Update Services Running list
  - Update "Next Up" to M6

- [ ] Update `M5-CHECKLIST.md`:
  - Mark all phases 0-8 complete
  - Add implementation notes
  - Add test coverage metrics

### 8.4 Docker Integration (Optional for M5)

Note: Docker deployment can be deferred to M6 or M7.

- [ ] Create `services/telemetry/Dockerfile` (multi-stage build)
- [ ] Create `services/device-interface/Dockerfile`
- [ ] Add services to `docker-compose.yml`:
  ```yaml
  telemetry:
    build: ./services/telemetry
    ports:
      - "8082:8080"
    environment:
      DATABASE_URL: postgres://telemetry_user:telemetry_pass@telemetry-db:5432/telemetry?sslmode=disable
      NATS_URL: nats://nats:4222
    depends_on:
      telemetry-db:
        condition: service_healthy

  device-interface:
    build: ./services/device-interface
    ports:
      - "8083:8080"
    environment:
      NATS_URL: nats://nats:4222
      ADAPTER_TYPE: tesla
  ```

- [ ] Test: `docker-compose up --build`

### 8.5 Final Commit and Tag

- [ ] Review all changes:
  ```bash
  git status
  git diff
  ```

- [ ] Commit M5 implementation:
  ```bash
  git add -A
  git commit -m "feat(M5): Implement Telemetry + Device Interface services

  Phase 0-8 Complete

  Telemetry Service:
  - BatteryState domain with validation (8 rules)
  - PostgreSQL repository with time-series optimization
  - REST API (GET current, GET history)
  - 1 Hz BatteryStateChanged event publishing
  - Test coverage: XX%

  Device Interface Service:
  - BatteryAdapter interface (hardware abstraction)
  - TeslaLike mock adapter (5 MW/s ramp, 95% efficiency)
  - BYDLike mock adapter (3 MW/s ramp, 92% efficiency)
  - Event subscriptions (charging/discharging commands)
  - Lifecycle event publishing (started/completed)
  - Test coverage: XX%

  Key Achievements:
  - High-frequency event publishing (1 Hz stable)
  - Swappable battery adapters
  - CustomAttributes support
  - Time-series database optimization
  - End-to-end event flow verified

  Next: M6 - Bidding Service

  🤖 Generated with [Claude Code](https://claude.com/claude-code)

  Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
  ```

- [ ] Create tag:
  ```bash
  git tag -a m5-complete -m "M5: Telemetry + Device Interface - COMPLETE

  Deliverables:
  ✅ Telemetry Service with 1 Hz event publishing
  ✅ Device Interface Service with hardware abstraction
  ✅ BatteryAdapter interface with 2 mock implementations
  ✅ REST API for querying current and historical state
  ✅ Event-driven command handling
  ✅ CustomAttributes support
  ✅ Test coverage >80% overall

  Services Status:
  - Telemetry: REST API (8082) + 1 Hz Event Publishing
  - Device Interface: Event-driven command handler
  - Infrastructure: NATS (4222), PostgreSQL x3 (5432, 5433, 5434)

  Ready for M6: Bidding Service"
  ```

- [ ] Push (optional):
  ```bash
  git push origin develop --tags
  ```

**Checkpoint**: ✅ M5 complete and ready for M6

---

## ✅ M5 Completion Criteria

- [ ] Mock battery simulation running (TeslaLike and BYDLike adapters)
- [ ] State changes emit events (BatteryStateChanged at 1 Hz)
- [ ] 2 adapter types are swappable without code changes
- [ ] Telemetry Service REST API functional
- [ ] Device Interface reacts to commands within 1 second
- [ ] All tests passing with >80% coverage overall (>85% domain)
- [ ] End-to-end event flow verified
- [ ] Documentation complete

---

## 📊 Test Coverage Summary

**Target Coverage**:
- Overall: >80%
- Domain Layer: >85%
- Repository Layer: >80%
- HTTP Layer: >65%
- Adapter Layer: >85%

**Actual Coverage** (fill in after completion):
- Telemetry Service: ___%
- Device Interface Service: ___%
- pkg/events: 88.9% (maintained)

---

## 🔑 Key Decisions

**Telemetry Service**:
- PostgreSQL for time-series storage
- Indexes on (battery_id, timestamp DESC)
- 1 Hz publishing via background goroutine
- Best-effort event publishing

**Device Interface Service**:
- Stateless (no database)
- In-memory state in adapters
- Background simulation at 10 Hz (100ms ticks)
- Ramp rate enforcement
- Thread-safe state access

**Both Services**:
- TDD workflow (Red → Green → Refactor)
- Hexagonal Architecture
- Event-driven coordination
- Graceful NATS degradation

---

## 📚 Related Documentation

- [M5 Overview](./M5-OVERVIEW.md) - Architecture and big picture
- [M5 Domain Spec](./M5-DOMAIN-SPEC.md) - Aggregates and validation
- [M5 API Spec](./M5-API-SPEC.md) - REST endpoints and events
- [PLANNING.md](../../PLANNING.md) - Project milestone tracking
- [pkg/events README](../../pkg/events/README.md) - Event library usage

---

**Estimated Time**: 12-16 hours

**Difficulty**: Moderate (proven patterns + new concepts)

**Prerequisites**: M0-M4 complete, infrastructure running
