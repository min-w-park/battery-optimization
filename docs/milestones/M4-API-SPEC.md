# M4: Event Bus Integration - API Specification

## 🔌 NATS Connection Information

### Development Environment

| Service | Port | Purpose |
|---------|------|---------|
| NATS Client | 4222 | Pub/Sub connections |
| NATS Monitoring | 8222 | HTTP monitoring/health |

**Connection URL**: `nats://localhost:4222`
**Monitoring URL**: `http://localhost:8222`

### Docker Compose Configuration

NATS is already running from M0 infrastructure setup:

```yaml
# From docker-compose.yml
nats:
  image: nats:2.10-alpine
  container_name: nats
  ports:
    - "4222:4222"  # Client connections
    - "8222:8222"  # HTTP monitoring
  command: ["-js", "-m", "8222"]
  healthcheck:
    test: ["CMD", "wget", "--spider", "-q", "http://localhost:8222/healthz"]
    interval: 10s
    timeout: 5s
    retries: 3
  restart: unless-stopped
```

---

## 📝 Subject Naming Conventions

### Subject Pattern

```
<service>.<entity>.<action>.<version>
```

**Components**:
- `service`: Source service domain (battery, market, telemetry)
- `entity`: Domain entity type (registered, price, state)
- `action`: Past-tense action (updated, changed, deleted)
- `version`: Schema version (v1, v2, v3)

### Benefits of Hierarchical Subjects

1. **Wildcard Subscriptions**: Subscribe to subsets of events
2. **Service Isolation**: Filter by source service
3. **Version Management**: Run multiple versions simultaneously
4. **Monitoring**: Track events by category

### Subject Examples

| Subject | Publisher | Event Type | Frequency |
|---------|-----------|------------|-----------|
| `battery.registered.v1` | Asset Management | BatteryRegistered | On-demand |
| `battery.updated.v1` | Asset Management | BatteryUpdated | On-demand |
| `battery.deleted.v1` | Asset Management | BatteryDeleted | On-demand |
| `market.price.updated.v1` | Market Data | MarketPriceUpdated | Every 5-30 min |
| `battery.state.changed.v1` | Telemetry (M5) | BatteryStateChanged | Every 1 second |

### Wildcard Patterns

| Pattern | Matches | Use Case |
|---------|---------|----------|
| `battery.registered.v1` | Exact subject | Specific event subscription |
| `battery.*.v1` | All battery events (v1) | Monitor all battery changes |
| `*.price.*.v1` | All price events (v1) | Track pricing across services |
| `battery.>` | All battery events (all versions) | Debug/logging |
| `>` | **All events** | Test subscriber, monitoring |

---

## 📨 Event Message Schemas

### 1. BatteryRegistered

**Subject**: `battery.registered.v1`
**Size**: ~400 bytes
**Frequency**: On-demand (when operator creates battery)

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

---

### 2. MarketPriceUpdated

**Subject**: `market.price.updated.v1`
**Size**: ~250 bytes
**Frequency**: Every 5-30 minutes (depending on interval type)

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

---

### 3. BatteryStateChanged (M5)

**Subject**: `battery.state.changed.v1`
**Size**: ~200 bytes
**Frequency**: 1 Hz (1 event per second per battery)

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

**Note**: Not implemented in M4 - placeholder for M5.

---

## 🚀 Publisher API

### Go Code Example

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
)

func publishBatteryRegistered(publisher events.EventPublisher, battery *domain.Battery) error {
	// Create event from domain aggregate
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

	// Publish with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := publisher.Publish(ctx, "battery.registered.v1", event); err != nil {
		log.Printf("WARN: Failed to publish BatteryRegistered: %v", err)
		return err
	}

	log.Printf("Published BatteryRegistered for battery %s", battery.ID)
	return nil
}
```

### Error Handling: Best-Effort Publishing

**Key Principle**: Event publishing failures should **NOT** fail business operations.

```go
func (h *BatteryHandler) CreateBattery(w http.ResponseWriter, r *http.Request) {
	// 1. Business logic (CRITICAL - must succeed)
	battery, err := domain.NewBattery(...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), battery); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Event publishing (BEST-EFFORT - log failures, don't block)
	go func() {
		if h.publisher != nil {
			if err := publishBatteryRegistered(h.publisher, battery); err != nil {
				log.Printf("WARN: Event publish failed (battery %s): %v", battery.ID, err)
				// Could add to retry queue here
			}
		}
	}()

	// 3. Return success to user (don't wait for event)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(battery)
}
```

**Rationale**:
- User gets immediate HTTP response
- Service remains operational even if NATS is down
- Events are "nice to have" for subscribers, not critical for user request
- Failures are logged for monitoring/alerting

---

## 📥 Subscriber API

### Go Code Example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/minwook/battery-optimization/pkg/events"
)

func main() {
	// Connect to NATS
	subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer subscriber.Close()

	// Define event handler
	handler := func(subject string, data []byte) error {
		switch subject {
		case "battery.registered.v1":
			return handleBatteryRegistered(data)
		case "market.price.updated.v1":
			return handleMarketPriceUpdated(data)
		default:
			log.Printf("Unknown event: %s", subject)
			return nil
		}
	}

	// Subscribe to multiple subjects
	ctx := context.Background()
	if err := subscriber.Subscribe(ctx, "battery.*.v1", handler); err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}
	if err := subscriber.Subscribe(ctx, "market.*.v1", handler); err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	log.Println("Subscribed to battery.*.v1 and market.*.v1")
	select {} // Block forever
}

func handleBatteryRegistered(data []byte) error {
	var event events.BatteryRegistered
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	log.Printf("Battery registered: %s (%.0f MWh, %.0f MW)",
		event.BatteryID, event.Capacity, event.MaxPower)

	// Business logic here...
	return nil
}
```

### Wildcard Subscriptions

```go
// Subscribe to all battery events (v1 only)
subscriber.Subscribe(ctx, "battery.*.v1", handler)

// Subscribe to all events from all services (v1 only)
subscriber.Subscribe(ctx, "*.*.v1", handler)

// Subscribe to everything (all versions)
subscriber.Subscribe(ctx, ">", handler)
```

---

## 🔗 Integration with REST APIs

### HTTP Flow with Event Publishing

```
┌─────────────────────────────────────────────────────────┐
│ 1. HTTP Request                                         │
│    POST /api/v1/batteries                               │
│    { "capacity": 50, "max_power": 25, ... }            │
└────────────────┬────────────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Domain Validation                                    │
│    battery := NewBattery(...)                           │
│    ✅ Validate capacity, power, constraints             │
└────────────────┬────────────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Database Save (Transactional)                        │
│    repo.Create(battery)                                 │
│    ✅ Persist to PostgreSQL                             │
└────────────────┬────────────────────────────────────────┘
                 │
                 ├─────────────────────────────┐
                 │                             │
                 ↓                             ↓
┌────────────────────────────┐  ┌──────────────────────────┐
│ 4a. HTTP Response          │  │ 4b. Publish Event        │
│     201 Created            │  │     (Async, Best-Effort) │
│     { battery JSON }       │  │     go publish(...)      │
└────────────────────────────┘  └─────────┬────────────────┘
                                          │
                                          ↓
                                ┌──────────────────────────┐
                                │ NATS Event Bus           │
                                │ battery.registered.v1    │
                                └─────────┬────────────────┘
                                          │
                                          ↓
                                ┌──────────────────────────┐
                                │ Subscribers Receive      │
                                │ (Test Tool, M6 Bidding)  │
                                └──────────────────────────┘
```

### Code Example: Handler with Event Publishing

```go
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/minwook/battery-optimization/pkg/events"
	"github.com/minwook/battery-optimization/services/asset-management/internal/domain"
	"github.com/minwook/battery-optimization/services/asset-management/internal/ports"
)

type BatteryHandler struct {
	repo      ports.BatteryRepository
	publisher events.EventPublisher // Optional (nil if NATS unavailable)
}

func NewBatteryHandler(repo ports.BatteryRepository, publisher events.EventPublisher) *BatteryHandler {
	return &BatteryHandler{
		repo:      repo,
		publisher: publisher,
	}
}

func (h *BatteryHandler) CreateBattery(w http.ResponseWriter, r *http.Request) {
	var req CreateBatteryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	// Domain logic
	battery, err := domain.NewBattery(
		req.Capacity,
		req.MaxPower,
		req.RampRate,
		req.Efficiency,
		req.Location,
		req.Manufacturer,
		req.Constraints,
	)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Persist
	if err := h.repo.Create(r.Context(), battery); err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
		return
	}

	// Publish event (async, best-effort)
	go h.publishBatteryRegistered(battery)

	// Success response
	respondJSON(w, http.StatusCreated, battery)
}

func (h *BatteryHandler) publishBatteryRegistered(battery *domain.Battery) {
	if h.publisher == nil {
		return // Graceful degradation: NATS not configured
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
		log.Printf("WARN: Failed to publish BatteryRegistered for %s: %v", battery.ID, err)
	}
}
```

### Key Design Decisions

1. **Publisher is Optional**: `h.publisher` can be `nil` (NATS unavailable)
2. **Goroutine for Publishing**: `go h.publishBatteryRegistered(battery)` avoids blocking HTTP response
3. **Timeout**: 5-second context timeout prevents hanging
4. **Logging Only**: Publish failures are logged, not returned to user
5. **After DB Save**: Events published **after** successful persistence (no phantom events)

---

## ⚠️ Error Handling

### Scenario 1: NATS Connection Fails (Startup)

**Situation**: NATS server is down when service starts.

**Behavior**:
```go
func main() {
	// Try to connect to NATS
	publisher, err := events.NewNATSPublisher(natsURL)
	if err != nil {
		log.Fatalf("FATAL: Cannot connect to NATS: %v", err)
		// Service exits - event infrastructure is critical
	}
	defer publisher.Close()

	// Continue with HTTP server setup...
}
```

**Decision**: **FAIL FAST** - Service should not start without NATS connection.

**Rationale**:
- Event infrastructure is foundational in M4+
- Better to fail at startup than run in degraded mode silently
- Docker/Kubernetes will restart service until NATS is available

---

### Scenario 2: Publish Fails (Runtime)

**Situation**: NATS is running, but publish operation fails (network issue, timeout).

**Behavior**:
```go
if err := h.publisher.Publish(ctx, subject, event); err != nil {
	log.Printf("WARN: Failed to publish %s: %v", subject, err)
	// Don't fail HTTP request
	// Could add to retry queue here
}
```

**Decision**: **LOG WARNING** - Continue serving HTTP requests.

**Rationale**:
- User's HTTP request already succeeded (data saved to DB)
- Events are "nice to have" for async workflows, not critical for user
- Service remains operational (degraded mode: no events published)

---

### Scenario 3: Invalid Event Format

**Situation**: Subscriber receives malformed JSON that cannot be unmarshaled.

**Behavior**:
```go
func handler(subject string, data []byte) error {
	var event events.BatteryRegistered
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ERROR: Invalid event format on %s: %v", subject, err)
		log.Printf("Raw data: %s", string(data))
		return nil // ACK message (don't redelivery - it will never succeed)
	}

	// Process valid event...
	return nil
}
```

**Decision**: **ACK MESSAGE** - Don't redelivery (would fail again).

**Rationale**:
- Malformed JSON won't fix itself
- Redelivery would create infinite loop
- Log error for debugging (shows in monitoring)

---

### Scenario 4: Event Processing Fails

**Situation**: Subscriber successfully unmarshals event but business logic fails (e.g., database error).

**Behavior**:
```go
func handler(subject string, data []byte) error {
	var event events.BatteryRegistered
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("ERROR: Unmarshal failed: %v", err)
		return nil // ACK (don't retry)
	}

	// Business logic
	if err := processBatteryRegistered(event); err != nil {
		log.Printf("ERROR: Processing failed for %s: %v", event.BatteryID, err)
		return err // NACK - will be redelivered
	}

	return nil // ACK
}
```

**Decision**: **NACK MESSAGE** - Return error to trigger redelivery.

**Rationale**:
- Temporary errors (DB connection) may resolve on retry
- NATS will redeliver message (with backoff)
- Subscriber must be idempotent (handle duplicates)

---

## 📊 NATS Monitoring Endpoints

### Health Check

```bash
curl http://localhost:8222/healthz
```

**Response** (healthy):
```json
{
  "status": "ok"
}
```

**Response** (unhealthy):
```json
{
  "status": "unavailable"
}
```

**HTTP Status**:
- `200 OK`: NATS is healthy
- `503 Service Unavailable`: NATS is down

---

### Server Variables (varz)

```bash
curl http://localhost:8222/varz | jq
```

**Response**:
```json
{
  "server_id": "NCVS4S6O7DTGQ2WJWXKZ3XZFZX3T5LNVZXZLZ",
  "server_name": "nats",
  "version": "2.10.0",
  "proto": 1,
  "go": "go1.21.0",
  "host": "0.0.0.0",
  "port": 4222,
  "max_connections": 65536,
  "ping_interval": 120000000000,
  "ping_max": 2,
  "http_host": "0.0.0.0",
  "http_port": 8222,
  "max_control_line": 4096,
  "max_payload": 1048576,
  "start": "2025-12-30T01:23:45.123456Z",
  "now": "2025-12-30T10:45:12.654321Z",
  "uptime": "9h21m27s",
  "mem": 12582912,
  "cores": 8,
  "cpu": 0.5,
  "connections": 3,
  "total_connections": 15,
  "routes": 0,
  "remotes": 0,
  "leafnodes": 0,
  "in_msgs": 1234,
  "out_msgs": 3702,
  "in_bytes": 123456,
  "out_bytes": 370200,
  "slow_consumers": 0
}
```

**Key Fields**:
- `connections`: Current active connections (should be 3 in M4: 2 publishers + 1 subscriber)
- `in_msgs` / `out_msgs`: Message counts
- `slow_consumers`: Number of slow subscribers (should be 0)

---

### Connection Info (connz)

```bash
curl http://localhost:8222/connz | jq
```

**Response**:
```json
{
  "server_id": "NCVS4S6O7DTGQ2WJWXKZ3XZFZX3T5LNVZXZLZ",
  "now": "2025-12-30T10:45:12.654321Z",
  "num_connections": 3,
  "total": 3,
  "offset": 0,
  "limit": 1024,
  "connections": [
    {
      "cid": 1,
      "ip": "172.18.0.5",
      "port": 52314,
      "start": "2025-12-30T10:30:00Z",
      "last_activity": "2025-12-30T10:45:10Z",
      "uptime": "15m12s",
      "idle": "2s",
      "pending_bytes": 0,
      "in_msgs": 150,
      "out_msgs": 0,
      "in_bytes": 15000,
      "out_bytes": 0,
      "subscriptions": 0,
      "name": "asset-management-publisher",
      "lang": "go",
      "version": "1.31.0"
    },
    {
      "cid": 2,
      "ip": "172.18.0.6",
      "port": 52315,
      "subscriptions": 1,
      "name": "market-data-publisher"
    },
    {
      "cid": 3,
      "ip": "172.18.0.7",
      "port": 52316,
      "subscriptions": 1,
      "name": "test-subscriber"
    }
  ]
}
```

**Key Fields**:
- `num_connections`: Should be 3 in M4 (asset-management, market-data, test-subscriber)
- `subscriptions`: Number of active subscriptions per connection
- `in_msgs` / `out_msgs`: Per-connection message counts

---

## 🧪 Testing with NATS CLI

### Install NATS CLI

```bash
# macOS
brew install nats-io/nats-tools/nats

# Linux
curl -L https://github.com/nats-io/natscli/releases/download/v0.1.1/nats-0.1.1-linux-amd64.zip -o nats.zip
unzip nats.zip
sudo mv nats /usr/local/bin/
```

### Subscribe to Events

```bash
# Subscribe to all events
nats sub ">"

# Subscribe to battery events only (v1)
nats sub "battery.*.v1"

# Subscribe to specific event
nats sub "battery.registered.v1"
```

**Output**:
```
[#1] Received on "battery.registered.v1"
{"battery_id":"9959b197-2b4d-4d25-8f38-592e130f92e2","capacity":50,"max_power":25,...}

[#2] Received on "market.price.updated.v1"
{"price_id":"129b6ded-ff74-495d-a52a-0daee54e85b0","region":"NSW","price":85.5,...}
```

### Publish Test Events

```bash
# Publish test BatteryRegistered event
nats pub battery.registered.v1 '{"battery_id":"test-123","capacity":100,"timestamp":"2025-12-30T10:00:00Z","event_version":"v1"}'

# Publish test MarketPriceUpdated event
nats pub market.price.updated.v1 '{"price_id":"test-456","region":"NSW","price":120,"timestamp":"2025-12-30T10:00:00Z","event_version":"v1"}'
```

---

## 🧭 Example: 3-Terminal Test Flow

### Terminal 1: Start Test Subscriber

```bash
cd tools/event-subscriber
go run main.go
```

**Output**:
```
2025-12-30T10:30:00Z Connected to NATS at nats://localhost:4222
2025-12-30T10:30:00Z Subscribed to: >
2025-12-30T10:30:00Z Listening for events... (Ctrl+C to exit)
```

---

### Terminal 2: Create Battery (Asset Management API)

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

**HTTP Response**:
```json
{
  "id": "9959b197-2b4d-4d25-8f38-592e130f92e2",
  "capacity": 50.0,
  "max_power": 25.0,
  ...
}
```

---

### Terminal 1: See Event Logged

```
[2025-12-30T10:30:15Z] battery.registered.v1
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
  "timestamp": "2025-12-30T10:30:15Z",
  "event_version": "v1"
}
```

---

### Terminal 3: Create Market Price (Market Data API)

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

---

### Terminal 1: See MarketPriceUpdated Event

```
[2025-12-30T10:35:20Z] market.price.updated.v1
{
  "price_id": "129b6ded-ff74-495d-a52a-0daee54e85b0",
  "region": "NSW",
  "price": 85.50,
  "demand": 8200.0,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T11:00:00Z",
  "published_at": "2025-12-30T10:55:00Z",
  "timestamp": "2025-12-30T10:35:20Z",
  "event_version": "v1"
}
```

---

## 🎯 Success Criteria

M4 API integration is successful when:

1. **NATS Healthy**: `curl http://localhost:8222/healthz` returns `200 OK`
2. **3 Connections**: `curl http://localhost:8222/connz | jq '.num_connections'` returns `3`
3. **Events Published**: Create battery → BatteryRegistered event appears in test subscriber
4. **Events Published**: Create price → MarketPriceUpdated event appears in test subscriber
5. **HTTP Still Works**: Services respond to HTTP requests even if NATS is down
6. **Graceful Degradation**: Publish failures logged, not fatal

---

**Next**: [M4-CHECKLIST.md](M4-CHECKLIST.md) - Step-by-step implementation guide
