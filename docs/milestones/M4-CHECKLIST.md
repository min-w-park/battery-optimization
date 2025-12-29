# M4: Implementation Checklist

## 📋 How to Use This Checklist

**⚠️ CRITICAL: READ THESE FIRST**:
1. [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development philosophy and TDD workflow (Kent Beck style)
2. [M4-OVERVIEW.md](M4-OVERVIEW.md) - Event-driven architecture and learning objectives
3. [M4-DOMAIN-SPEC.md](M4-DOMAIN-SPEC.md) - Event schemas, versioning, publisher/subscriber interfaces
4. [M4-API-SPEC.md](M4-API-SPEC.md) - NATS pub/sub API, subject naming, monitoring

**Then follow this checklist**:
1. Work through tasks in order (top to bottom)
2. **Write tests BEFORE implementation** (Red → Green → Refactor)
3. Check off `[ ]` boxes as you complete each task
4. Each phase should take 1-3 hours
5. If stuck, refer to detailed spec documents
6. Commit after each major milestone

---

## Phase 0: Create M4 Milestone Documentation (1-1.5 hours)

### 0.1 Documentation Files
- [ ] Create `docs/milestones/M4-OVERVIEW.md`
- [ ] Create `docs/milestones/M4-DOMAIN-SPEC.md`
- [ ] Create `docs/milestones/M4-API-SPEC.md`
- [ ] Create `docs/milestones/M4-CHECKLIST.md`

### 0.2 Review and Commit
- [ ] Review all 4 documentation files
- [ ] Verify event schemas match EVENTS.md
- [ ] Verify NATS connection info matches docker-compose.yml
- [ ] Commit documentation before starting Phase 1
  ```bash
  git add docs/milestones/M4-*.md
  git commit -m "docs(M4): add Event Bus Integration milestone documentation"
  ```

**Checkpoint**: ✅ Documentation complete, ready to start implementation

---

## Phase 1: Common Event Library (2-3 hours)

### 1.1 Package Structure
- [ ] Create `pkg/events/` directory
- [ ] Initialize Go module:
  ```bash
  cd pkg/events
  go mod init github.com/minwook/battery-optimization/pkg/events
  ```

### 1.2 Event Structs (TDD - Tests First!)

#### Write Tests FIRST (Red Phase)
- [ ] Create `pkg/events/battery_registered_test.go`
- [ ] Write test: `TestBatteryRegistered_JSONSerialization`
  - Create BatteryRegistered event with all fields
  - Marshal to JSON
  - Unmarshal back
  - Verify all fields match
- [ ] Write test: `TestBatteryRegistered_JSONTags`
  - Verify JSON field names (snake_case: `battery_id`, `max_power`, etc.)
- [ ] Run tests → Should FAIL (Red phase) ✅

- [ ] Create `pkg/events/market_price_updated_test.go`
- [ ] Write test: `TestMarketPriceUpdated_JSONSerialization`
- [ ] Write test: `TestMarketPriceUpdated_Timestamps`
  - Verify IntervalStart, PublishedAt, Timestamp all preserved
- [ ] Run tests → Should FAIL (Red phase) ✅

- [ ] Create `pkg/events/battery_state_changed_test.go`
- [ ] Write test: `TestBatteryStateChanged_JSONSerialization`
- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All event tests written and failing

#### Implementation (Green Phase)
- [ ] Create `pkg/events/battery_registered.go`
- [ ] Define `BatteryRegistered` struct with JSON tags:
  ```go
  type BatteryRegistered struct {
      BatteryID    string             `json:"battery_id"`
      Capacity     float64            `json:"capacity"`
      MaxPower     float64            `json:"max_power"`
      RampRate     float64            `json:"ramp_rate"`
      Efficiency   float64            `json:"efficiency"`
      Location     string             `json:"location"`
      Manufacturer string             `json:"manufacturer"`
      Constraints  BatteryConstraints `json:"constraints"`
      Timestamp    time.Time          `json:"timestamp"`
      EventVersion string             `json:"event_version"`
  }

  type BatteryConstraints struct {
      MinSoC      float64 `json:"min_soc"`
      MaxSoC      float64 `json:"max_soc"`
      WarrantyEOL float64 `json:"warranty_eol"`
      MaxCycles   int     `json:"max_cycles"`
  }
  ```

- [ ] Create `pkg/events/market_price_updated.go`
- [ ] Define `MarketPriceUpdated` struct with JSON tags

- [ ] Create `pkg/events/battery_state_changed.go`
- [ ] Define `BatteryStateChanged` struct (placeholder for M5)

- [ ] Run tests → All pass ✅ (Green phase)

### 1.3 Publisher Interface

#### Tests First (Red Phase)
- [ ] Create `pkg/events/publisher_test.go`
- [ ] Write test: `TestEventPublisher_Interface`
  - Verify interface signature
- [ ] Run test → Should FAIL ✅

#### Implementation (Green Phase)
- [ ] Create `pkg/events/publisher.go`
- [ ] Define `EventPublisher` interface:
  ```go
  type EventPublisher interface {
      Publish(ctx context.Context, subject string, event interface{}) error
      Close() error
  }
  ```
- [ ] Run test → Should pass ✅

### 1.4 Subscriber Interface

#### Tests First (Red Phase)
- [ ] Create `pkg/events/subscriber_test.go`
- [ ] Write test: `TestEventSubscriber_Interface`
- [ ] Run test → Should FAIL ✅

#### Implementation (Green Phase)
- [ ] Create `pkg/events/subscriber.go`
- [ ] Define `EventHandler` and `EventSubscriber`:
  ```go
  type EventHandler func(subject string, data []byte) error

  type EventSubscriber interface {
      Subscribe(ctx context.Context, subject string, handler EventHandler) error
      Close() error
  }
  ```
- [ ] Run test → Should pass ✅

### 1.5 Test Coverage
- [ ] Run: `go test -cover ./...`
- [ ] Verify coverage >80% (target: 85%+)
- [ ] All event serialization tests passing

**Checkpoint**: ✅ Common event library complete, interfaces defined, events tested

---

## Phase 2: NATS Publisher Adapter (2-3 hours)

### 2.1 Dependencies
- [ ] Add NATS dependency:
  ```bash
  cd pkg/events
  go get github.com/nats-io/nats.go
  ```

### 2.2 Publisher Tests - Write FIRST! (Red Phase)

- [ ] Create `pkg/events/nats_publisher_test.go`
- [ ] Write test: `TestNewNATSPublisher_Success`
  - Connect to `nats://localhost:4222`
  - Verify no error
  - Verify connection established
- [ ] Write test: `TestNewNATSPublisher_ConnectionFailure`
  - Try to connect to invalid URL
  - Should return error
- [ ] Write test: `TestNATSPublisher_Publish`
  - Create NATSPublisher
  - Publish BatteryRegistered event
  - Verify no error
- [ ] Write test: `TestNATSPublisher_PublishWithTimeout`
  - Publish with context timeout
  - Verify respects timeout
- [ ] Write test: `TestNATSPublisher_Close`
  - Create publisher
  - Call Close()
  - Verify graceful shutdown
- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All publisher tests written and failing

### 2.3 Publisher Implementation (Green Phase)

- [ ] Create `pkg/events/nats_publisher.go`
- [ ] Implement `NATSPublisher` struct:
  ```go
  type NATSPublisher struct {
      conn *nats.Conn
  }
  ```
- [ ] Implement `NewNATSPublisher(url string)`:
  - Connect to NATS server
  - Return publisher or error
- [ ] Implement `Publish(ctx context.Context, subject string, event interface{})`:
  - Marshal event to JSON
  - Publish to NATS subject
  - Respect context timeout
  - Return error if publish fails
- [ ] Implement `Close()`:
  - Drain connection
  - Close connection
- [ ] Run tests → All pass ✅ (Green phase)

### 2.4 Integration Test with Real NATS

- [ ] Create `pkg/events/nats_integration_test.go`
- [ ] Write test: `TestNATSPublisher_Integration`
  - Requires NATS running on localhost:4222
  - Publish event
  - Verify message sent (use subscriber to verify)
- [ ] Mark as integration test:
  ```go
  // +build integration
  ```
- [ ] Run: `go test -tags=integration -v ./...`

### 2.5 Test Coverage
- [ ] Run: `go test -cover ./...`
- [ ] Verify publisher coverage >80%

**Checkpoint**: ✅ NATS publisher working, can publish events

---

## Phase 3: NATS Subscriber Adapter (2-3 hours)

### 3.1 Subscriber Tests - Write FIRST! (Red Phase)

- [ ] Create `pkg/events/nats_subscriber_test.go`
- [ ] Write test: `TestNewNATSSubscriber_Success`
  - Connect to NATS
  - Verify connection
- [ ] Write test: `TestNATSSubscriber_Subscribe`
  - Subscribe to subject
  - Publish event from separate publisher
  - Verify handler called with correct subject and data
- [ ] Write test: `TestNATSSubscriber_WildcardSubscription`
  - Subscribe to `battery.*.v1`
  - Publish to `battery.registered.v1` and `battery.updated.v1`
  - Verify handler receives both
- [ ] Write test: `TestNATSSubscriber_AllEventsWildcard`
  - Subscribe to `>`
  - Publish multiple event types
  - Verify all received
- [ ] Write test: `TestNATSSubscriber_HandlerError`
  - Handler returns error
  - Verify message is NACKed (redelivered)
- [ ] Write test: `TestNATSSubscriber_Close`
  - Subscribe
  - Close subscriber
  - Verify graceful shutdown
- [ ] Run tests → Should FAIL (Red phase) ✅

**TDD Checkpoint**: ✅ All subscriber tests written and failing

### 3.2 Subscriber Implementation (Green Phase)

- [ ] Create `pkg/events/nats_subscriber.go`
- [ ] Implement `NATSSubscriber` struct:
  ```go
  type NATSSubscriber struct {
      conn          *nats.Conn
      subscriptions []*nats.Subscription
      mu            sync.Mutex
  }
  ```
- [ ] Implement `NewNATSSubscriber(url string)`:
  - Connect to NATS
  - Initialize subscriptions slice
- [ ] Implement `Subscribe(ctx context.Context, subject string, handler EventHandler)`:
  - Create NATS subscription
  - Wrap handler to convert nats.Msg to EventHandler signature
  - Store subscription for cleanup
  - Handle errors from handler (NACK on error, ACK on success)
- [ ] Implement `Close()`:
  - Unsubscribe all subscriptions
  - Close connection
- [ ] Run tests → All pass ✅ (Green phase)

### 3.3 Integration Test

- [ ] Create `pkg/events/nats_subscriber_integration_test.go`
- [ ] Write test: `TestNATSSubscriber_EndToEnd`
  - Create publisher and subscriber
  - Subscribe to events
  - Publish BatteryRegistered
  - Verify received
  - Publish MarketPriceUpdated
  - Verify received
- [ ] Mark as integration test: `// +build integration`
- [ ] Run: `go test -tags=integration -v ./...`

### 3.4 Test Coverage
- [ ] Run: `go test -cover ./...`
- [ ] Verify subscriber coverage >80%

**Checkpoint**: ✅ NATS subscriber working, can receive events with wildcards

---

## Phase 4: Asset Service Integration (2-3 hours)

### 4.1 Add Event Dependency

- [ ] Update `services/asset-management/go.mod`:
  ```bash
  cd services/asset-management
  go get github.com/minwook/battery-optimization/pkg/events
  ```

### 4.2 Add EventPublisher Port

- [ ] Create `services/asset-management/internal/ports/event_publisher.go`:
  ```go
  package ports

  import "github.com/minwook/battery-optimization/pkg/events"

  // EventPublisher is a type alias to the events package interface
  type EventPublisher = events.EventPublisher
  ```

### 4.3 Update Handler

- [ ] Modify `services/asset-management/internal/adapters/http/handler.go`:
  - Add `publisher events.EventPublisher` field to `BatteryHandler`
  - Update `NewBatteryHandler()` to accept optional publisher
  - Add `publishBatteryRegistered()` method
  - Call `go h.publishBatteryRegistered(battery)` after DB save in `CreateBattery()`

**Code Example**:
```go
type BatteryHandler struct {
    repo      ports.BatteryRepository
    publisher events.EventPublisher // NEW
}

func NewBatteryHandler(repo ports.BatteryRepository, publisher events.EventPublisher) *BatteryHandler {
    return &BatteryHandler{
        repo:      repo,
        publisher: publisher,
    }
}

func (h *BatteryHandler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    // ... validation ...

    // Save to DB
    if err := h.repo.Create(r.Context(), battery); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Publish event (async, best-effort)
    go h.publishBatteryRegistered(battery)

    // Return response
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(battery)
}

func (h *BatteryHandler) publishBatteryRegistered(battery *domain.Battery) {
    if h.publisher == nil {
        return // Graceful degradation
    }

    event := events.BatteryRegistered{
        BatteryID:    battery.ID,
        Capacity:     battery.Capacity,
        MaxPower:     battery.MaxPower,
        RampRate:     battery.RampRate,
        Efficiency:   battery.Efficiency,
        Location:     battery.Location,
        Manufacturer: battery.Manufacturer,
        Constraints: events.BatteryConstraints{
            MinSoC:      battery.Constraints.MinSoC,
            MaxSoC:      battery.Constraints.MaxSoC,
            WarrantyEOL: battery.Constraints.WarrantyEOL,
            MaxCycles:   battery.Constraints.MaxCycles,
        },
        Timestamp:    time.Now(),
        EventVersion: "v1",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := h.publisher.Publish(ctx, "battery.registered.v1", event); err != nil {
        log.Printf("WARN: Failed to publish BatteryRegistered: %v", err)
    }
}
```

### 4.4 Update Main

- [ ] Modify `services/asset-management/cmd/server/main.go`:
  - Connect to NATS (fatal if fails)
  - Create NATSPublisher
  - Pass publisher to handler
  - Close publisher on shutdown

**Code Example**:
```go
func main() {
    // ... existing DB setup ...

    // Connect to NATS
    natsURL := getEnv("NATS_URL", "nats://localhost:4222")
    publisher, err := events.NewNATSPublisher(natsURL)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer publisher.Close()
    log.Printf("Connected to NATS at %s", natsURL)

    // Create handler with publisher
    batteryHandler := httpAdapter.NewBatteryHandler(batteryRepo, publisher)

    // ... rest of setup ...
}
```

### 4.5 Update Docker Compose

- [ ] Update `docker-compose.yml` to add NATS_URL to asset-management:
  ```yaml
  asset-management:
    environment:
      DATABASE_URL: postgres://asset_user:asset_pass@asset-db:5432/asset_management?sslmode=disable
      PORT: "8080"
      NATS_URL: nats://nats:4222  # NEW
    depends_on:
      asset-db:
        condition: service_healthy
      nats:  # NEW
        condition: service_healthy
  ```

### 4.6 Test Integration

- [ ] Rebuild service: `docker-compose build asset-management`
- [ ] Start service: `docker-compose up -d asset-management`
- [ ] Check logs: `docker-compose logs asset-management | grep NATS`
  - Should see "Connected to NATS at nats://nats:4222"
- [ ] Create battery and verify event published (Phase 7)

**Checkpoint**: ✅ Asset Service publishing BatteryRegistered events

---

## Phase 5: Market Service Integration (2-3 hours)

### 5.1 Add Event Dependency

- [ ] Update `services/market-data/go.mod`:
  ```bash
  cd services/market-data
  go get github.com/minwook/battery-optimization/pkg/events
  ```

### 5.2 Add EventPublisher Port

- [ ] Create `services/market-data/internal/ports/event_publisher.go`:
  ```go
  package ports

  import "github.com/minwook/battery-optimization/pkg/events"

  type EventPublisher = events.EventPublisher
  ```

### 5.3 Update Handler

- [ ] Modify `services/market-data/internal/adapters/http/handler.go`:
  - Add `publisher events.EventPublisher` field
  - Update constructor
  - Add `publishMarketPriceUpdated()` method
  - Call after DB save in `CreateMarketPrice()`

**Code Example**:
```go
func (h *MarketPriceHandler) publishMarketPriceUpdated(price *domain.MarketPrice) {
    if h.publisher == nil {
        return
    }

    event := events.MarketPriceUpdated{
        PriceID:       price.ID,
        Region:        price.Region,
        Price:         price.Price,
        Demand:        price.Demand,
        IntervalType:  string(price.IntervalType),
        IntervalStart: price.IntervalStart,
        PublishedAt:   price.PublishedAt,
        Timestamp:     time.Now(),
        EventVersion:  "v1",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := h.publisher.Publish(ctx, "market.price.updated.v1", event); err != nil {
        log.Printf("WARN: Failed to publish MarketPriceUpdated: %v", err)
    }
}
```

### 5.4 Update Main

- [ ] Modify `services/market-data/cmd/server/main.go`:
  - Connect to NATS
  - Create publisher
  - Pass to handler
  - Close on shutdown

### 5.5 Update Docker Compose

- [ ] Update `docker-compose.yml` to add NATS_URL to market-data:
  ```yaml
  market-data:
    environment:
      DATABASE_URL: postgres://market_user:market_pass@market-db:5432/market_data?sslmode=disable
      PORT: "8080"
      NATS_URL: nats://nats:4222  # NEW
    depends_on:
      market-db:
        condition: service_healthy
      nats:  # NEW
        condition: service_healthy
  ```

### 5.6 Test Integration

- [ ] Rebuild: `docker-compose build market-data`
- [ ] Start: `docker-compose up -d market-data`
- [ ] Check logs: `docker-compose logs market-data | grep NATS`

**Checkpoint**: ✅ Market Service publishing MarketPriceUpdated events

---

## Phase 6: Test Subscriber Tool (1-2 hours)

### 6.1 Create Tool Structure

- [ ] Create `tools/event-subscriber/` directory
- [ ] Initialize Go module:
  ```bash
  cd tools/event-subscriber
  go mod init github.com/minwook/battery-optimization/tools/event-subscriber
  go get github.com/minwook/battery-optimization/pkg/events
  ```

### 6.2 Implement Subscriber

- [ ] Create `tools/event-subscriber/main.go`:

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
)

func main() {
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	// Connect to NATS
	subscriber, err := events.NewNATSSubscriber(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer subscriber.Close()

	log.Printf("Connected to NATS at %s", natsURL)

	// Event handler
	handler := func(subject string, data []byte) error {
		// Pretty-print JSON
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err != nil {
			log.Printf("ERROR: Invalid JSON on %s: %v", subject, err)
			return nil // ACK anyway
		}

		fmt.Printf("\n[%s] %s\n", time.Now().Format(time.RFC3339), subject)
		fmt.Println(pretty.String())
		return nil
	}

	// Subscribe to all events
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := subscriber.Subscribe(ctx, ">", handler); err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	log.Println("Subscribed to: >")
	log.Println("Listening for events... (Ctrl+C to exit)")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

### 6.3 Create README

- [ ] Create `tools/event-subscriber/README.md`:

```markdown
# Event Subscriber Tool

Simple CLI tool to subscribe to all NATS events and pretty-print them to the console.

## Usage

```bash
# Run locally
go run main.go

# Build binary
go build -o event-subscriber

# Run with custom NATS URL
NATS_URL=nats://localhost:4222 ./event-subscriber
```

## Output Example

```
[2025-12-30T10:30:15Z] battery.registered.v1
{
  "battery_id": "9959b197-2b4d-4d25-8f38-592e130f92e2",
  "capacity": 50.0,
  "max_power": 25.0,
  ...
}

[2025-12-30T10:35:20Z] market.price.updated.v1
{
  "price_id": "129b6ded-ff74-495d-a52a-0daee54e85b0",
  "region": "NSW",
  "price": 85.50,
  ...
}
```
```

### 6.4 Test Tool

- [ ] Run tool: `go run main.go`
- [ ] Verify connects to NATS
- [ ] Leave running for Phase 7 tests

**Checkpoint**: ✅ Test subscriber tool working, listening for events

---

## Phase 7: Integration Testing (2-3 hours)

### 7.1 End-to-End Event Flow

#### Setup
- [ ] Start all services:
  ```bash
  docker-compose up -d
  ```
- [ ] Start test subscriber in separate terminal:
  ```bash
  cd tools/event-subscriber
  go run main.go
  ```

#### Test BatteryRegistered Event
- [ ] Create battery via API:
  ```bash
  curl -X POST http://localhost:8080/api/v1/batteries \
    -H "Content-Type: application/json" \
    -d '{
      "capacity": 50.0,
      "max_power": 25.0,
      "ramp_rate": 5.0,
      "efficiency": 0.92,
      "location": "Sydney, NSW",
      "manufacturer": "Tesla Megapack",
      "constraints": {
        "min_soc": 0.2,
        "max_soc": 0.9,
        "warranty_eol": 0.7,
        "max_cycles": 10000
      }
    }'
  ```
- [ ] Verify HTTP response: 201 Created with battery ID
- [ ] Verify test subscriber shows BatteryRegistered event
- [ ] Verify event has correct battery_id, capacity, constraints

#### Test MarketPriceUpdated Event
- [ ] Create market price via API:
  ```bash
  curl -X POST http://localhost:8081/api/v1/prices \
    -H "Content-Type: application/json" \
    -d '{
      "region": "NSW",
      "price": 85.50,
      "demand": 8200.0,
      "interval_type": "5MIN_PREDISPATCH",
      "interval_start": "2025-12-30T11:00:00Z",
      "published_at": "2025-12-30T10:55:00Z"
    }'
  ```
- [ ] Verify HTTP response: 201 Created
- [ ] Verify test subscriber shows MarketPriceUpdated event
- [ ] Verify event has correct region, price, interval_type

#### Test Multiple Events
- [ ] Create 3 batteries
- [ ] Create 5 market prices
- [ ] Verify all 8 events received by subscriber
- [ ] Verify events in correct order per publisher

### 7.2 NATS CLI Verification

- [ ] Install NATS CLI (if not installed):
  ```bash
  brew install nats-io/nats-tools/nats
  ```

- [ ] Subscribe to battery events:
  ```bash
  nats sub "battery.*.v1"
  ```
- [ ] Create battery → Verify event appears

- [ ] Subscribe to all events:
  ```bash
  nats sub ">"
  ```
- [ ] Create battery and price → Verify both appear

### 7.3 NATS Monitoring Verification

- [ ] Check NATS health:
  ```bash
  curl http://localhost:8222/healthz
  ```
  → Expected: `{"status":"ok"}` with HTTP 200

- [ ] Check connections:
  ```bash
  curl http://localhost:8222/connz | jq '.num_connections'
  ```
  → Expected: 3 (asset-management, market-data, test-subscriber)

- [ ] View connection details:
  ```bash
  curl http://localhost:8222/connz | jq '.connections[] | {name, subscriptions}'
  ```
  → Verify each connection has subscriptions

- [ ] Check server stats:
  ```bash
  curl http://localhost:8222/varz | jq '{in_msgs, out_msgs, connections}'
  ```
  → Verify in_msgs and out_msgs > 0

### 7.4 Error Scenarios

#### Scenario 1: NATS Down at Service Start
- [ ] Stop NATS: `docker-compose stop nats`
- [ ] Try to start asset-management: `docker-compose up asset-management`
- [ ] Verify service fails with "Failed to connect to NATS" error
- [ ] Start NATS: `docker-compose start nats`
- [ ] Verify service starts successfully

#### Scenario 2: NATS Fails During Runtime
- [ ] Start all services
- [ ] Create battery → Verify event published
- [ ] Stop NATS: `docker-compose stop nats`
- [ ] Create another battery
- [ ] Verify HTTP response: 201 Created (service still works!)
- [ ] Check logs: Should see "WARN: Failed to publish" message
- [ ] Start NATS: `docker-compose start nats`
- [ ] Create another battery → Events working again

#### Scenario 3: Publisher Timeout
- [ ] Create battery with very short timeout (modify code temporarily)
- [ ] Verify publish fails gracefully (logged, not fatal)

**Checkpoint**: ✅ All integration tests passing, end-to-end event flow verified

---

## Phase 8: Polish & Documentation (1-2 hours)

### 8.1 Code Quality

- [ ] Run `go fmt ./...` in pkg/events → **All files formatted** ✅
- [ ] Run `go vet ./...` in pkg/events → **No warnings** ✅
- [ ] Run `go fmt ./...` in services/asset-management → **Formatted** ✅
- [ ] Run `go fmt ./...` in services/market-data → **Formatted** ✅
- [ ] Run `go fmt ./...` in tools/event-subscriber → **Formatted** ✅
- [ ] Add comments to exported functions and types

### 8.2 Test Coverage

- [ ] Run: `go test -cover ./...` in pkg/events
  - Target: >80% overall
  - Event structs: >90%
  - Publisher: >80%
  - Subscriber: >80%
- [ ] Run integration tests: `go test -tags=integration -v ./...`
- [ ] All tests pass ✅

### 8.3 Documentation

- [ ] Create `pkg/events/README.md`:
  - Package purpose
  - Event types supported
  - Usage examples (publisher and subscriber)
  - Testing instructions
  - Integration with services

Example README structure:
```markdown
# Events Package

Common event library for battery optimization system.

## Events

- **BatteryRegistered**: Published when battery added to system
- **MarketPriceUpdated**: Published when market price received
- **BatteryStateChanged**: Placeholder for M5 (telemetry)

## Usage

### Publishing Events

```go
publisher, _ := events.NewNATSPublisher("nats://localhost:4222")
event := events.BatteryRegistered{...}
publisher.Publish(ctx, "battery.registered.v1", event)
```

### Subscribing to Events

```go
subscriber, _ := events.NewNATSSubscriber("nats://localhost:4222")
handler := func(subject string, data []byte) error {
    // Process event...
    return nil
}
subscriber.Subscribe(ctx, "battery.*.v1", handler)
```

## Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires NATS)
go test -tags=integration -v ./...
```
```

- [ ] Update `PLANNING.md`:
  - Mark M4 as completed
  - Add completion date
  - Add test coverage metrics
  - Note any deviations from plan

- [ ] Update `CLAUDE.md`:
  - Add "Event Publishing" section
  - Document event catalog (3 events)
  - Explain best-effort publishing pattern
  - Add NATS monitoring section

Example CLAUDE.md addition:
```markdown
## Event Publishing

Services publish domain events to NATS event bus for asynchronous communication.

### Event Catalog

- `battery.registered.v1` - Asset Management publishes after battery creation
- `market.price.updated.v1` - Market Data publishes after price creation
- `battery.state.changed.v1` - Telemetry publishes every 1 second (M5)

### Best-Effort Publishing

Events are published asynchronously and failures don't block HTTP responses:
- DB save succeeds → HTTP 201 returned immediately
- Event publish happens in goroutine (async)
- Publish failures are logged, not fatal

See [pkg/events/README.md](pkg/events/README.md) for usage.
```

### 8.4 Git

- [ ] Commit all changes:
  ```bash
  git add pkg/events/ services/asset-management/ services/market-data/ tools/event-subscriber/ docs/milestones/M4-*.md docker-compose.yml PLANNING.md CLAUDE.md
  git commit -m "feat(M4): implement Event Bus Integration with NATS

  - Add common event library (pkg/events) with 3 event types
  - Implement NATS publisher and subscriber adapters
  - Integrate event publishing in Asset Management Service
  - Integrate event publishing in Market Data Service
  - Add test subscriber tool for manual verification
  - End-to-end event flow verified

  M4 complete: Event-driven architecture established
  - BatteryRegistered published after battery creation
  - MarketPriceUpdated published after price creation
  - pkg/events coverage: >80%
  - 3 NATS connections verified (2 publishers + 1 subscriber)
  - Best-effort publishing with graceful degradation

  Events:
  - battery.registered.v1 (~400 bytes, on-demand)
  - market.price.updated.v1 (~250 bytes, every 5-30 min)
  - battery.state.changed.v1 (placeholder for M5)

  Ready for M5: Telemetry Service with high-frequency events"
  ```

- [ ] Push to GitHub:
  ```bash
  git push origin develop
  ```

- [ ] Tag release:
  ```bash
  git tag m4-complete
  git push origin m4-complete
  ```

**Checkpoint**: ✅ M4 complete, code polished, documentation updated, changes committed

---

## 🎯 Definition of Done

All items below must be true:

**Functionality**:
- [ ] pkg/events library created with 3 event types (BatteryRegistered, MarketPriceUpdated, BatteryStateChanged)
- [ ] EventPublisher and EventSubscriber interfaces defined
- [ ] NATSPublisher implemented and tested
- [ ] NATSSubscriber implemented and tested (wildcard support)
- [ ] Asset Management Service publishes BatteryRegistered events
- [ ] Market Data Service publishes MarketPriceUpdated events
- [ ] Test subscriber tool receives all events

**Testing**:
- [ ] pkg/events test coverage >80%
- [ ] Event serialization tests pass (JSON marshal/unmarshal)
- [ ] Publisher integration tests pass (with real NATS)
- [ ] Subscriber integration tests pass (wildcard subscriptions)
- [ ] End-to-end test: Create battery → BatteryRegistered received
- [ ] End-to-end test: Create price → MarketPriceUpdated received

**Architecture**:
- [ ] Services remain loosely coupled via events
- [ ] REST APIs continue working (synchronous communication preserved)
- [ ] Event versioning infrastructure in place (subject + payload version)
- [ ] Graceful degradation when NATS unavailable (publish failures logged, not fatal)
- [ ] NATS monitoring shows 3 connections (2 publishers + 1 subscriber)

**Code Quality**:
- [ ] Code formatted (`go fmt`)
- [ ] No vet warnings (`go vet`)
- [ ] Exported functions documented
- [ ] Error handling throughout (best-effort publishing)

**Documentation**:
- [ ] All 4 M4 milestone docs created (OVERVIEW, DOMAIN-SPEC, API-SPEC, CHECKLIST)
- [ ] pkg/events/README.md created
- [ ] PLANNING.md updated with M4 completion
- [ ] CLAUDE.md updated with event publishing section
- [ ] tools/event-subscriber/README.md created

**Infrastructure**:
- [ ] docker-compose.yml updated (NATS_URL for services)
- [ ] Services depend on NATS (fails at startup if unavailable)
- [ ] NATS health check verified: `curl http://localhost:8222/healthz`
- [ ] NATS connections verified: `curl http://localhost:8222/connz | jq '.num_connections'` returns 3

**Git**:
- [ ] Changes committed with descriptive message
- [ ] Git tag created (m4-complete)
- [ ] Pushed to GitHub

---

## 🐛 Common Issues & Solutions

### NATS Connection Refused
```
Error: nats: no servers available for connection
Solution:
- Check NATS is running: docker-compose ps nats
- Verify NATS_URL correct: nats://nats:4222 (in Docker) or nats://localhost:4222 (local)
- Check NATS health: curl http://localhost:8222/healthz
```

### Event Not Received by Subscriber
```
Problem: Published event but subscriber doesn't receive
Solution:
- Verify subscriber is connected: Check test subscriber logs
- Verify subject matches: "battery.registered.v1" exact match or wildcard
- Check NATS stats: curl http://localhost:8222/varz | jq '.in_msgs, .out_msgs'
- Use NATS CLI to verify: nats sub ">" (should see all events)
```

### Publish Timeout
```
Error: context deadline exceeded
Solution:
- Increase timeout: ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
- Check NATS latency: curl http://localhost:8222/varz | jq '.cpu, .mem'
- Verify NATS not slow_consumers: curl http://localhost:8222/varz | jq '.slow_consumers'
```

### JSON Unmarshal Error
```
Error: json: cannot unmarshal string into Go struct field
Solution:
- Verify JSON tags match: `json:"battery_id"` not `json:"batteryID"`
- Check event_version field: Should be "v1" (string)
- Verify timestamps in RFC3339 format: time.Now().Format(time.RFC3339)
```

---

## 📊 Progress Tracking

**Estimated Time**: 13-18 hours
**Actual Time**: ___ hours (completed YYYY-MM-DD)

**Phases Completed**:
- [ ] Phase 0: Documentation (1-1.5h)
- [ ] Phase 1: Common Event Library (2-3h)
- [ ] Phase 2: NATS Publisher (2-3h)
- [ ] Phase 3: NATS Subscriber (2-3h)
- [ ] Phase 4: Asset Integration (2-3h)
- [ ] Phase 5: Market Integration (2-3h)
- [ ] Phase 6: Test Subscriber (1-2h)
- [ ] Phase 7: Integration Testing (2-3h)
- [ ] Phase 8: Polish (1-2h)

**Test Coverage Achieved**:
- pkg/events: ___% (target: >80%)
- Event serialization: ___% (target: >90%)
- Publisher: ___% (target: >80%)
- Subscriber: ___% (target: >80%)

**Implementation Notes**:
```
(Add notes here as you implement)
- NATS connection: ...
- Event versioning: ...
- Challenges faced: ...
- Key decisions: ...
```

---

## ✅ Ready for M5

Once M4 is complete:
- ✅ Event infrastructure established
- ✅ Can easily add new events (just add struct to pkg/events)
- ✅ Publishers and subscribers working
- ✅ NATS monitoring in place
- **M5**: Add BatteryStateChanged (high-frequency, 1 event/second)
- **M6**: Bidding Service subscribes to all 3 events

**Great job! 🎉**

---

**Next Steps**:
1. Review what you learned (event-driven patterns, NATS pub/sub)
2. Test event flow manually (3-terminal setup)
3. Start M5 (Telemetry Service) when ready

**Key Learnings from M4**:
- Events enable loose coupling between services
- Best-effort publishing keeps services available
- Event versioning critical for backward compatibility
- NATS wildcards provide flexible subscriptions
- Async communication trades consistency for scalability
