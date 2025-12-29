# M4: Event Bus Integration - Domain Specification

## 📖 What are Domain Events?

**Domain Events** are immutable records of significant occurrences within the system. They represent facts about the past that other parts of the system may need to know about.

### Key Characteristics

1. **Past Tense**: Events describe what has happened (e.g., `BatteryRegistered`, not `RegisterBattery`)
2. **Immutable**: Once published, events cannot be changed or deleted
3. **Self-Contained**: Events include all data needed by consumers (no need to call back to publisher)
4. **Timestamped**: Every event has a timestamp marking when it occurred
5. **Versioned**: Events include version information for backward compatibility

### Events vs Commands vs Queries

| Aspect | Command | Event | Query |
|--------|---------|-------|-------|
| **Tense** | Imperative ("Do this") | Past ("This happened") | Present ("What is...?") |
| **Example** | `CreateBattery` | `BatteryRegistered` | `GetBattery` |
| **Can Fail** | Yes | No (already happened) | No (read-only) |
| **Direction** | Request → Service | Service → Subscribers | Request → Service |
| **Timing** | Before action | After action | Anytime |

---

## 🎯 Event Schemas

M4 implements **3 core events** from the full event catalog:

1. **BatteryRegistered** - Published when a battery is added to the system
2. **MarketPriceUpdated** - Published when market price data is received
3. **BatteryStateChanged** - Placeholder for M5 (high-frequency telemetry)

---

### 1. BatteryRegistered

**Published By**: Asset Management Service
**Published When**: After battery successfully saved to database
**NATS Subject**: `battery.registered.v1`
**Frequency**: On-demand (when operator creates battery)

#### Go Struct Definition

```go
package events

import "time"

// BatteryRegistered is published when a new battery is added to the system
type BatteryRegistered struct {
	// Battery identification
	BatteryID string `json:"battery_id"`

	// Battery specifications
	Capacity     float64 `json:"capacity"`      // MWh
	MaxPower     float64 `json:"max_power"`     // MW
	RampRate     float64 `json:"ramp_rate"`     // MW/min
	Efficiency   float64 `json:"efficiency"`    // 0-1

	// Additional metadata
	Location     string `json:"location"`
	Manufacturer string `json:"manufacturer"`

	// Constraints (embedded struct)
	Constraints BatteryConstraints `json:"constraints"`

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}

// BatteryConstraints represents operational limits
type BatteryConstraints struct {
	MinSoC       float64 `json:"min_soc"`        // 0-1 (e.g., 0.2 for 20%)
	MaxSoC       float64 `json:"max_soc"`        // 0-1 (e.g., 0.9 for 90%)
	WarrantyEOL  float64 `json:"warranty_eol"`   // 0-1 (e.g., 0.7 for 70% capacity)
	MaxCycles    int     `json:"max_cycles"`
}
```

#### JSON Example

```json
{
  "battery_id": "9959b197-2b4d-4d25-8f38-592e130f92e2",
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
  },
  "timestamp": "2025-12-30T10:30:00Z",
  "event_version": "v1"
}
```

#### Size Estimate
**~400 bytes per event**

#### Business Rules

- Published **AFTER** successful database save (transactional consistency)
- Contains complete battery specification (subscribers don't need to query Asset Service)
- Includes all constraints needed for bidding logic
- Failure to publish is logged but doesn't fail HTTP request (best-effort)

---

### 2. MarketPriceUpdated

**Published By**: Market Data Service
**Published When**: After market price successfully saved to database
**NATS Subject**: `market.price.updated.v1`
**Frequency**: Every 5-30 minutes (depending on AEMO interval type)

#### Go Struct Definition

```go
package events

import "time"

// MarketPriceUpdated is published when new market price data is received
type MarketPriceUpdated struct {
	// Price identification
	PriceID string `json:"price_id"`

	// Market data
	Region        string    `json:"region"`         // NSW, VIC, QLD, SA, TAS
	Price         float64   `json:"price"`          // $/MWh
	Demand        float64   `json:"demand"`         // MW
	IntervalType  string    `json:"interval_type"`  // 5MIN_PREDISPATCH, 30MIN_PREDISPATCH
	IntervalStart time.Time `json:"interval_start"` // When this price applies
	PublishedAt   time.Time `json:"published_at"`   // When AEMO published it

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
```

#### JSON Example

```json
{
  "price_id": "129b6ded-ff74-495d-a52a-0daee54e85b0",
  "region": "NSW",
  "price": 85.50,
  "demand": 8200.0,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z",
  "timestamp": "2025-12-30T09:55:05Z",
  "event_version": "v1"
}
```

#### Size Estimate
**~250 bytes per event**

#### Business Rules

- Published **AFTER** successful database save
- Timestamp vs PublishedAt vs IntervalStart:
  - `interval_start`: When this price applies (future)
  - `published_at`: When AEMO published forecast
  - `timestamp`: When our service processed it
- High frequency: 5MIN_PREDISPATCH updates every 5 minutes → ~12 events/hour/region
- 5 regions × 12 events/hour = 60 events/hour at peak

---

### 3. BatteryStateChanged (Placeholder for M5)

**Published By**: Telemetry Service (M5)
**Published When**: Every 1 second (high-frequency)
**NATS Subject**: `battery.state.changed.v1`
**Frequency**: 1 Hz (1 event per second per battery)

#### Go Struct Definition

```go
package events

import "time"

// BatteryStateChanged is published every second with current battery state
// NOTE: Not implemented in M4 - placeholder for M5
type BatteryStateChanged struct {
	// Battery identification
	BatteryID string `json:"battery_id"`

	// Current state
	SoC         float64 `json:"soc"`          // State of Charge (0-1)
	Power       float64 `json:"power"`        // MW (positive = discharging, negative = charging)
	Status      string  `json:"status"`       // IDLE, CHARGING, DISCHARGING, FCAS
	Temperature float64 `json:"temperature"`  // °C

	// Event metadata
	Timestamp    time.Time `json:"timestamp"`
	EventVersion string    `json:"event_version"` // "v1"
}
```

#### JSON Example

```json
{
  "battery_id": "9959b197-2b4d-4d25-8f38-592e130f92e2",
  "soc": 0.75,
  "power": 12.5,
  "status": "DISCHARGING",
  "temperature": 25.3,
  "timestamp": "2025-12-30T10:30:45Z",
  "event_version": "v1"
}
```

#### Size Estimate
**~200 bytes per event**

#### Performance Considerations

- **High Volume**: 1 event/second × 10 batteries = 10 events/second = 36,000 events/hour
- **Not Persisted**: Events are ephemeral (in-memory NATS streams only)
- **FCAS Requirement**: 1-second granularity needed for frequency control response
- **M4 Scope**: Event struct defined, but NOT published until M5

---

## 🔄 Event Versioning

### Why Versioning Matters

Event schemas evolve over time. Versioning ensures:
- Old subscribers continue working when new fields added
- Breaking changes are explicit (new version)
- Publishers and subscribers can upgrade independently

### Versioning Strategy

M4 uses **subject-based versioning** with **payload confirmation**:

1. **Version in Subject**: `battery.registered.v1`, `battery.registered.v2`
2. **Version in Payload**: `"event_version": "v1"`

#### Backward-Compatible Changes (Same Version)

✅ **Safe changes that don't break existing subscribers**:

```go
// Version 1 (original)
type BatteryRegistered struct {
	BatteryID    string
	Capacity     float64
	MaxPower     float64
	Timestamp    time.Time
	EventVersion string // "v1"
}

// Version 1 (enhanced - still v1!)
type BatteryRegistered struct {
	BatteryID    string
	Capacity     float64
	MaxPower     float64
	// New optional field (backward-compatible)
	SerialNumber string `json:"serial_number,omitempty"`
	Timestamp    time.Time
	EventVersion string // "v1"
}
```

**Why this is safe**:
- Old subscribers ignore `serial_number` (Go zero value: empty string)
- JSON unmarshaling skips unknown fields
- No subscriber behavior changes

**Rules**:
- ✅ Add new optional field
- ✅ Add new event type
- ✅ Make required field optional (with sensible default)

#### Breaking Changes (New Version)

❌ **Changes that require new version**:

```go
// Version 1
type BatteryRegistered struct {
	BatteryID    string
	Capacity     float64  // MWh
	EventVersion string   // "v1"
}

// Version 2 (BREAKING CHANGE)
type BatteryRegistered struct {
	BatteryID    string
	CapacityKWh  float64  // kWh instead of MWh (RENAMED + UNIT CHANGE)
	EventVersion string   // "v2"
}
```

**Why this is breaking**:
- Field renamed: `Capacity` → `CapacityKWh`
- Units changed: MWh → kWh
- Old subscribers would miss the data (looking for `Capacity`)

**Rules**:
- ❌ Remove field
- ❌ Rename field
- ❌ Change field type (e.g., `string` → `int`)
- ❌ Change field semantics (e.g., units, meaning)
- ❌ Make optional field required

**How to handle**:
1. Publish to new subject: `battery.registered.v2`
2. Set `event_version: "v2"` in payload
3. Run v1 and v2 publishers in parallel during migration
4. Subscribers upgrade when ready
5. Deprecate v1 after all subscribers migrated

### Versioning Example: Adding a Field

```go
// events/battery_registered_v1.go
type BatteryRegistered struct {
	BatteryID    string
	Capacity     float64
	MaxPower     float64
	Timestamp    time.Time
	EventVersion string // "v1"
}

// Later: Add warranty info (backward-compatible, still v1)
type BatteryRegistered struct {
	BatteryID    string
	Capacity     float64
	MaxPower     float64
	// New field (optional, backward-compatible)
	WarrantyYears int `json:"warranty_years,omitempty"`
	Timestamp     time.Time
	EventVersion  string // "v1"
}
```

**Migration**:
1. Update `BatteryRegistered` struct
2. Publisher starts including `warranty_years`
3. Old subscribers ignore it (no code changes needed)
4. New subscribers can use it
5. No version bump needed!

---

## 🔌 Publisher Interface

### EventPublisher Interface

```go
package events

import "context"

// EventPublisher publishes domain events to NATS
type EventPublisher interface {
	// Publish sends an event to the specified NATS subject
	// Returns error if publish fails (logged, not fatal)
	Publish(ctx context.Context, subject string, event interface{}) error

	// Close gracefully shuts down the publisher
	Close() error
}
```

### Usage in HTTP Handler

```go
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
)

type BatteryHandler struct {
	repo      ports.BatteryRepository
	publisher events.EventPublisher  // NEW!
}

func (h *BatteryHandler) CreateBattery(w http.ResponseWriter, r *http.Request) {
	// ... parse request, validate ...

	// 1. Save to database (transactional)
	battery, err := domain.NewBattery(req.Capacity, req.MaxPower, ...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), battery); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Publish event (best-effort, non-blocking)
	go h.publishBatteryRegistered(battery)

	// 3. Return HTTP response (don't wait for event)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(battery)
}

func (h *BatteryHandler) publishBatteryRegistered(battery *domain.Battery) {
	if h.publisher == nil {
		return // No publisher configured (graceful degradation)
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
		log.Printf("WARN: Failed to publish BatteryRegistered event: %v", err)
		// Don't fail the HTTP request - best-effort delivery
	}
}
```

### Design Decisions

1. **Goroutine for Publishing**: Use `go h.publishBatteryRegistered()` to avoid blocking HTTP response
2. **Graceful Degradation**: If publisher is nil (NATS unavailable), service remains operational
3. **Logging Failures**: Log publish errors for monitoring, but don't fail business operation
4. **Timeout**: 5-second timeout prevents hanging on NATS issues
5. **After DB Save**: Only publish if database save succeeds (avoid phantom events)

---

## 📥 Subscriber Interface

### EventSubscriber Interface

```go
package events

import "context"

// EventHandler processes a received event
type EventHandler func(subject string, data []byte) error

// EventSubscriber subscribes to domain events from NATS
type EventSubscriber interface {
	// Subscribe registers a handler for events matching the subject pattern
	// Supports wildcards: * (single token), > (multiple tokens)
	// Examples:
	//   - "battery.registered.v1" (exact match)
	//   - "battery.*.v1" (all battery events, version 1)
	//   - ">" (all events)
	Subscribe(ctx context.Context, subject string, handler EventHandler) error

	// Close gracefully shuts down the subscriber
	Close() error
}
```

### Usage in Subscriber Tool

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minwook/battery-optimization/pkg/events"
)

func main() {
	// Connect to NATS
	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer subscriber.Close()

	// Subscribe to all events
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(subject string, data []byte) error {
		// Pretty-print JSON
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err != nil {
			log.Printf("ERROR: Invalid JSON: %v", err)
			return nil // ACK anyway (don't redelivery)
		}

		fmt.Printf("\n[%s] %s\n", time.Now().Format(time.RFC3339), subject)
		fmt.Println(pretty.String())
		return nil
	}

	// Subscribe to all events (wildcard)
	if err := subscriber.Subscribe(ctx, ">", handler); err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	log.Println("Listening for events... (Ctrl+C to exit)")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
```

### Wildcard Subscriptions

NATS supports hierarchical subjects with wildcards:

| Pattern | Matches | Example |
|---------|---------|---------|
| `battery.registered.v1` | Exact subject only | `battery.registered.v1` |
| `battery.*.v1` | All battery events (v1) | `battery.registered.v1`, `battery.updated.v1` |
| `*.price.*.v1` | All price events (v1) | `market.price.updated.v1` |
| `>` | **All events** | Everything |

---

## 🔁 Event Ordering and Idempotency

### Event Ordering

**NATS Guarantee**: Events published by a single publisher to a single subscriber are delivered in order.

**No Global Ordering**: Events from different publishers (e.g., Asset Service + Market Service) may be interleaved.

**Example**:
```
Publisher A: Event 1 → Event 2 → Event 3
Publisher B: Event X → Event Y

Subscriber may see: Event 1, Event X, Event 2, Event Y, Event 3
                    ✅ A's events ordered
                    ✅ B's events ordered
                    ❌ No global ordering
```

### Idempotency

**At-Least-Once Delivery**: NATS may deliver the same event multiple times (e.g., network retry).

**Subscribers Must Be Idempotent**: Processing the same event twice should have the same effect as processing it once.

#### Idempotent Subscriber Example

```go
type BatteryEventHandler struct {
	processedEvents map[string]bool // Event ID → processed
	mu              sync.Mutex
}

func (h *BatteryEventHandler) Handle(subject string, data []byte) error {
	var event events.BatteryRegistered
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	// Deduplication check
	h.mu.Lock()
	if h.processedEvents[event.BatteryID] {
		h.mu.Unlock()
		log.Printf("WARN: Duplicate BatteryRegistered for %s (ignoring)", event.BatteryID)
		return nil // Already processed, skip
	}
	h.processedEvents[event.BatteryID] = true
	h.mu.Unlock()

	// Process event (idempotent)
	log.Printf("Processing BatteryRegistered: %s", event.BatteryID)
	// ... business logic ...

	return nil
}
```

**Production Pattern**: Use database unique constraint or distributed cache (Redis) for deduplication.

---

## ✅ Test Cases (TDD)

### Domain Event Tests

```go
// pkg/events/battery_registered_test.go
package events_test

func TestBatteryRegistered_JSONSerialization(t *testing.T) {
	event := events.BatteryRegistered{
		BatteryID:    "test-id",
		Capacity:     50.0,
		MaxPower:     25.0,
		RampRate:     5.0,
		Efficiency:   0.92,
		Location:     "Sydney, NSW",
		Manufacturer: "Tesla Megapack",
		Constraints: events.BatteryConstraints{
			MinSoC:      0.2,
			MaxSoC:      0.9,
			WarrantyEOL: 0.7,
			MaxCycles:   10000,
		},
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	// Marshal to JSON
	data, err := json.Marshal(event)
	require.NoError(t, err)

	// Unmarshal back
	var decoded events.BatteryRegistered
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, "test-id", decoded.BatteryID)
	assert.Equal(t, 50.0, decoded.Capacity)
	assert.Equal(t, "v1", decoded.EventVersion)
}
```

### Publisher Tests

```go
// pkg/events/nats_publisher_test.go
func TestNATSPublisher_Publish(t *testing.T) {
	// Setup NATS test server
	natsURL := "nats://localhost:4222"
	publisher, err := events.NewNATSPublisher(natsURL)
	require.NoError(t, err)
	defer publisher.Close()

	// Publish event
	event := events.BatteryRegistered{
		BatteryID:    "test-id",
		Capacity:     50.0,
		Timestamp:    time.Now(),
		EventVersion: "v1",
	}

	ctx := context.Background()
	err = publisher.Publish(ctx, "battery.registered.v1", event)
	assert.NoError(t, err)
}
```

### Subscriber Tests

```go
// pkg/events/nats_subscriber_test.go
func TestNATSSubscriber_Receive(t *testing.T) {
	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	require.NoError(t, err)
	defer subscriber.Close()

	received := make(chan bool, 1)
	handler := func(subject string, data []byte) error {
		assert.Equal(t, "battery.registered.v1", subject)
		received <- true
		return nil
	}

	// Subscribe
	ctx := context.Background()
	err = subscriber.Subscribe(ctx, "battery.registered.v1", handler)
	require.NoError(t, err)

	// Publish event (from separate connection)
	publisher, _ := events.NewNATSPublisher("nats://localhost:4222")
	defer publisher.Close()

	event := events.BatteryRegistered{BatteryID: "test"}
	publisher.Publish(ctx, "battery.registered.v1", event)

	// Wait for event
	select {
	case <-received:
		// Success!
	case <-time.After(5 * time.Second):
		t.Fatal("Event not received")
	}
}
```

---

## 📈 Coverage Goals

| Layer | Target Coverage | Purpose |
|-------|-----------------|---------|
| Event Serialization | >90% | Ensure JSON marshaling works |
| Publisher | >80% | Publish logic + error handling |
| Subscriber | >80% | Subscribe logic + wildcards |
| Integration | 100% | End-to-end event flow |

---

## 🎯 Key Takeaways

1. **Events are Past Tense**: `BatteryRegistered`, not `RegisterBattery`
2. **Immutability is Critical**: Once published, events cannot change
3. **Self-Contained Data**: Events include everything subscribers need
4. **Versioning from Day One**: Use subject-based versioning (`battery.registered.v1`)
5. **Backward Compatibility**: Add fields without breaking changes
6. **Best-Effort Publishing**: Don't fail business operations if publish fails
7. **Idempotent Subscribers**: Handle duplicate events gracefully
8. **Wildcards for Flexibility**: `battery.*.v1` subscribes to all battery events

---

**Next**: [M4-API-SPEC.md](M4-API-SPEC.md) - NATS pub/sub API details
