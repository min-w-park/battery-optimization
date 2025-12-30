# pkg/events - Event Library

Common event library for the Battery Optimization system. This package defines domain events, event publishing/subscribing interfaces, and NATS adapters for event-driven communication between microservices.

## Overview

This library enables asynchronous, decoupled communication between services through a publish/subscribe (pub/sub) pattern using NATS as the message broker.

**Key Concepts**:
- **Domain Events**: Immutable records of things that have happened (past tense)
- **Event Publisher**: Publishes events to NATS subjects
- **Event Subscriber**: Subscribes to NATS subjects and processes events
- **Event Versioning**: Built-in versioning for backward compatibility

## Domain Events

### BatteryRegistered

Published by Asset Management Service when a new battery is created.

**Subject**: `battery.registered.v1`

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
    EventVersion string             `json:"event_version"` // "v1"
}

type BatteryConstraints struct {
    MinSoC               float64 `json:"min_soc"`
    MaxSoC               float64 `json:"max_soc"`
    OperatingTempMin     float64 `json:"operating_temp_min"`
    OperatingTempMax     float64 `json:"operating_temp_max"`
    MaxCycles            int     `json:"max_cycles"`
    WarrantyEoL          float64 `json:"warranty_eol"`
    GridComplianceLevel  string  `json:"grid_compliance_level"`
}
```

**Example**:
```json
{
  "battery_id": "550e8400-e29b-41d4-a716-446655440000",
  "capacity": 100.0,
  "max_power": 50.0,
  "ramp_rate": 10.0,
  "efficiency": 0.95,
  "location": "NSW",
  "manufacturer": "Tesla",
  "constraints": {
    "min_soc": 0.1,
    "max_soc": 0.9,
    "operating_temp_min": -10.0,
    "operating_temp_max": 45.0,
    "max_cycles": 5000,
    "warranty_eol": 0.8,
    "grid_compliance_level": "AS4777"
  },
  "timestamp": "2025-12-30T10:30:15.234Z",
  "event_version": "v1"
}
```

### MarketPriceUpdated

Published by Market Data Service when a new market price is created.

**Subject**: `market.price.updated.v1`

```go
type MarketPriceUpdated struct {
    PriceID       string    `json:"price_id"`
    Region        string    `json:"region"`
    Price         float64   `json:"price"`
    Demand        float64   `json:"demand"`
    IntervalType  string    `json:"interval_type"`
    IntervalStart time.Time `json:"interval_start"`
    PublishedAt   time.Time `json:"published_at"`
    Timestamp     time.Time `json:"timestamp"`
    EventVersion  string    `json:"event_version"` // "v1"
}
```

**Example**:
```json
{
  "price_id": "660e9511-f39c-52e5-b827-557766551111",
  "region": "NSW",
  "price": 75.50,
  "demand": 8500.0,
  "interval_type": "5MIN",
  "interval_start": "2025-12-30T10:30:00Z",
  "published_at": "2025-12-30T10:30:15Z",
  "timestamp": "2025-12-30T10:30:20.567Z",
  "event_version": "v1"
}
```

### BatteryStateChanged

Placeholder for future Telemetry Service (M5). Will be published every 1 second with battery state updates.

**Subject**: `battery.state.changed.v1`

```go
type BatteryStateChanged struct {
    BatteryID    string    `json:"battery_id"`
    SoC          float64   `json:"soc"`
    Power        float64   `json:"power"`
    Status       string    `json:"status"`
    Temperature  float64   `json:"temperature"`
    Timestamp    time.Time `json:"timestamp"`
    EventVersion string    `json:"event_version"` // "v1"
}
```

**Note**: This event is NOT implemented in M4. It will be added in M5.

## Publishing Events

### Publisher Interface

```go
type EventPublisher interface {
    Publish(ctx context.Context, subject string, event interface{}) error
    Close() error
}
```

### Usage

```go
import (
    "context"
    "time"
    "github.com/minwook/battery-optimization/pkg/events"
)

// 1. Connect to NATS
publisher, err := events.NewNATSPublisher("nats://localhost:4222")
if err != nil {
    log.Fatalf("Failed to connect to NATS: %v", err)
}
defer publisher.Close()

// 2. Create event
event := events.BatteryRegistered{
    BatteryID:    battery.ID,
    Capacity:     battery.Capacity,
    MaxPower:     battery.MaxPower,
    RampRate:     battery.RampRate,
    Efficiency:   battery.Efficiency,
    Location:     battery.Location,
    Manufacturer: battery.Manufacturer,
    Constraints: events.BatteryConstraints{
        MinSoC:               battery.Constraints.MinSoC,
        MaxSoC:               battery.Constraints.MaxSoC,
        OperatingTempMin:     battery.Constraints.OperatingTempMin,
        OperatingTempMax:     battery.Constraints.OperatingTempMax,
        MaxCycles:            battery.Constraints.MaxCycles,
        WarrantyEoL:          battery.Constraints.WarrantyEoL,
        GridComplianceLevel:  battery.Constraints.GridComplianceLevel,
    },
    Timestamp:    time.Now(),
    EventVersion: "v1",
}

// 3. Publish event (with timeout)
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

if err := publisher.Publish(ctx, "battery.registered.v1", event); err != nil {
    // Best-effort: Log warning but don't fail the business operation
    log.Printf("WARNING: Failed to publish event: %v", err)
}
```

### Best Practices

1. **Use timeouts**: Always use a context with timeout when publishing
2. **Best-effort publishing**: Don't fail HTTP requests if event publishing fails
3. **Log warnings**: Log publishing errors for debugging
4. **Subject naming**: Follow pattern `<service>.<entity>.<action>.<version>`
5. **Event versioning**: Always set `event_version` field to "v1"

## Subscribing to Events

### Subscriber Interface

```go
type EventHandler func(subject string, data []byte) error

type EventSubscriber interface {
    Subscribe(ctx context.Context, subject string, handler EventHandler) error
    Close() error
}
```

### Usage

```go
import (
    "context"
    "encoding/json"
    "log"
    "github.com/minwook/battery-optimization/pkg/events"
)

// 1. Connect to NATS
subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
if err != nil {
    log.Fatalf("Failed to connect to NATS: %v", err)
}
defer subscriber.Close()

// 2. Define event handler
handler := func(subject string, data []byte) error {
    log.Printf("Received event on subject: %s", subject)

    // Deserialize based on subject
    if subject == "battery.registered.v1" {
        var event events.BatteryRegistered
        if err := json.Unmarshal(data, &event); err != nil {
            return fmt.Errorf("failed to unmarshal event: %w", err)
        }
        log.Printf("Battery registered: %s", event.BatteryID)
    }

    return nil
}

// 3. Subscribe to specific events
ctx := context.Background()
if err := subscriber.Subscribe(ctx, "battery.registered.v1", handler); err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}

// Keep running...
select {}
```

### Wildcard Subscriptions

NATS supports wildcard subscriptions:

- `*` - Matches exactly one token
- `>` - Matches one or more tokens

**Examples**:

```go
// Subscribe to ALL events
subscriber.Subscribe(ctx, ">", handler)

// Subscribe to all battery events
subscriber.Subscribe(ctx, "battery.>", handler)

// Subscribe to all v1 events from any service
subscriber.Subscribe(ctx, "*.*.v1", handler)

// Subscribe to specific event
subscriber.Subscribe(ctx, "battery.registered.v1", handler)
```

**Wildcard Pattern Table**:

| Pattern | Matches |
|---------|---------|
| `>` | All events |
| `battery.>` | `battery.registered.v1`, `battery.state.changed.v1` |
| `market.>` | `market.price.updated.v1` |
| `*.*.v1` | All v1 events (1 token between dots) |
| `battery.*.v1` | `battery.registered.v1` (NOT `battery.state.changed.v1`) |

## Event Versioning

Events include versioning at two levels:

1. **Subject versioning**: `battery.registered.v1` → `battery.registered.v2`
2. **Payload versioning**: `"event_version": "v1"`

### Backward-Compatible Changes (Same Version)

These changes don't require a new version:
- Adding new optional fields
- Making required fields optional

**Example**:
```go
// v1 (original)
type BatteryRegistered struct {
    BatteryID string `json:"battery_id"`
    Capacity  float64 `json:"capacity"`
}

// v1 (backward-compatible change - added field)
type BatteryRegistered struct {
    BatteryID    string  `json:"battery_id"`
    Capacity     float64 `json:"capacity"`
    Manufacturer string  `json:"manufacturer,omitempty"` // NEW: optional
}
```

### Breaking Changes (New Version Required)

These changes require a new version:
- Removing fields
- Renaming fields
- Changing field types
- Making optional fields required

**Example**:
```go
// v1
type BatteryRegistered struct {
    BatteryID string `json:"battery_id"`
    Capacity  float64 `json:"capacity"`
}

// v2 (breaking change - renamed field)
type BatteryRegisteredV2 struct {
    ID       string  `json:"id"` // RENAMED: battery_id → id
    Capacity float64 `json:"capacity"`
}

// Publish to new subject
publisher.Publish(ctx, "battery.registered.v2", eventV2)
```

## Testing

### Running Tests

```bash
cd pkg/events
go test -v ./...
```

### Test Coverage

```bash
go test -cover ./...
```

**Current Coverage**: 88.9% (27/27 tests passing)

### Integration Testing

Use the event subscriber tool for manual testing:

```bash
# Terminal 1: Subscribe to all events
cd tools/event-subscriber
go run main.go

# Terminal 2: Start services
docker-compose up -d
cd services/asset-management
go run cmd/server/main.go

# Terminal 3: Trigger events
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{
    "capacity": 100.0,
    "max_power": 50.0,
    "ramp_rate": 10.0,
    "efficiency": 0.95,
    "location": "NSW",
    "manufacturer": "Tesla",
    "constraints": {
      "min_soc": 0.1,
      "max_soc": 0.9,
      "operating_temp_min": -10.0,
      "operating_temp_max": 45.0,
      "max_cycles": 5000,
      "warranty_eol": 0.8,
      "grid_compliance_level": "AS4777"
    }
  }'
```

You should see the `BatteryRegistered` event appear in Terminal 1!

## NATS Connection Details

**Server URL**: `nats://localhost:4222`
**Monitoring URL**: `http://localhost:8222`

### Health Check

```bash
curl http://localhost:8222/healthz
# Output: {"status":"ok"}
```

### Server Stats

```bash
curl http://localhost:8222/varz | jq
```

### Active Connections

```bash
curl http://localhost:8222/connz | jq
```

## Error Handling

### Publisher Errors

1. **Connection Failure**: Fatal error - service should not start
2. **Publish Failure**: Log warning, continue with business operation
3. **Serialization Failure**: Log error, investigate event structure

### Subscriber Errors

1. **Connection Failure**: Fatal error - subscriber should not start
2. **Handler Error**: Log error, ACK message (prevent redelivery loop)
3. **Deserialization Failure**: Log warning, ACK message

### Graceful Degradation

Services continue working even if NATS is unavailable:

```go
// In service startup
var publisher events.EventPublisher
if cfg.NatsURL != "" {
    natsPublisher, err := events.NewNATSPublisher(cfg.NatsURL)
    if err != nil {
        log.Printf("WARNING: Failed to connect to NATS: %v", err)
        log.Println("Service will continue WITHOUT event publishing")
    } else {
        publisher = natsPublisher
        defer publisher.Close()
    }
}

// In handler
if h.publisher != nil {
    if err := h.publisher.Publish(ctx, subject, event); err != nil {
        log.Printf("WARNING: Failed to publish event: %v", err)
    }
}
```

## Dependencies

```bash
go get github.com/nats-io/nats.go@v1.48.0
```

## Architecture

This library follows **Hexagonal Architecture** (Ports & Adapters):

**Core**:
- Event structs (domain entities)

**Ports**:
- `EventPublisher` interface (secondary port)
- `EventSubscriber` interface (primary port)

**Adapters**:
- `NATSPublisher` (secondary adapter - infrastructure)
- `NATSSubscriber` (primary adapter - infrastructure)

Services depend on interfaces, not concrete implementations. This enables:
- **Testability**: Mock publishers/subscribers in tests
- **Flexibility**: Swap NATS for Kafka/RabbitMQ without changing service code
- **Decoupling**: Services don't know about NATS implementation details

## Troubleshooting

### "Failed to connect to NATS"

**Problem**: Publisher/subscriber can't connect to NATS server.

**Solutions**:
1. Check NATS is running: `docker-compose ps nats`
2. Verify NATS health: `curl http://localhost:8222/healthz`
3. Check NATS URL: `NATS_URL=nats://localhost:4222`

### "Context requires a deadline"

**Problem**: Using `context.Background()` with `Publish()`.

**Solution**: Use `context.WithTimeout()`:
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
publisher.Publish(ctx, subject, event)
```

### No events received by subscriber

**Problem**: Subscriber not receiving published events.

**Solutions**:
1. Verify publisher is connected and publishing (check logs)
2. Check subject pattern matches: Use `>` to debug
3. Ensure subscriber is running before publisher
4. Check NATS monitoring: `curl http://localhost:8222/connz`

### Events appear truncated or malformed

**Problem**: JSON deserialization fails or data looks wrong.

**Solutions**:
1. Verify JSON tags match field names (snake_case)
2. Check event struct definition
3. Use event subscriber tool to see raw JSON
4. Validate timestamp format is RFC3339

## Related Documentation

- **Event Catalog**: [EVENTS.md](../../EVENTS.md) - Complete catalog of 25 domain events
- **M4 Milestone**: [docs/milestones/M4-OVERVIEW.md](../../docs/milestones/M4-OVERVIEW.md) - Event bus integration guide
- **Test Subscriber Tool**: [tools/event-subscriber/README.md](../../tools/event-subscriber/README.md) - Manual event monitoring
- **Architecture**: [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - System architecture diagrams

## Support

For issues or questions:
1. Check [docs/milestones/M4-CHECKLIST.md](../../docs/milestones/M4-CHECKLIST.md) for common problems
2. Review NATS documentation: https://docs.nats.io
3. Check service logs for warnings/errors
