# M5: API Specification

This document specifies the REST APIs, event schemas, and integration patterns for M5 (Telemetry + Device Interface).

---

## 1. Telemetry Service REST API

**Base URL**: `http://localhost:8082/api/v1`

**Authentication**: None (future: JWT tokens)

**Content-Type**: `application/json`

### 1.1 Get Current Battery State

Retrieve the latest telemetry reading for a specific battery.

#### Request

```http
GET /api/v1/telemetry/:batteryId/current
```

**Path Parameters**:
- `batteryId` (string, UUID) - Battery identifier from Asset Management

**Query Parameters**: None

#### Success Response (200 OK)

```json
{
  "id": "state-a1b2c3d4",
  "battery_id": "battery-123",
  "soc": 75.5,
  "power": 25.0,
  "temperature": 28.5,
  "voltage": 800.0,
  "current": 31.25,
  "operation_state": "DISCHARGING",
  "custom_attributes": {
    "vendor": "Tesla",
    "cellVoltageMin": 3.2,
    "thermalZone1Temp": 28.5
  },
  "timestamp": "2025-12-30T10:30:15Z",
  "created_at": "2025-12-30T10:30:15.234Z"
}
```

#### Error Responses

**404 Not Found** - Battery has no telemetry data
```json
{
  "error": "battery not found",
  "message": "No telemetry data exists for battery battery-123"
}
```

**400 Bad Request** - Invalid battery ID format
```json
{
  "error": "invalid request",
  "message": "batteryId must be a valid UUID"
}
```

#### Example

```bash
# Get current state
curl http://localhost:8082/api/v1/telemetry/battery-123/current

# Response
{
  "soc": 85.0,
  "power": -20.0,
  "operation_state": "CHARGING",
  ...
}
```

---

### 1.2 Get Historical Battery States

Query telemetry data within a time range.

#### Request

```http
GET /api/v1/telemetry/:batteryId/history?startTime={ISO8601}&endTime={ISO8601}&limit={int}&offset={int}
```

**Path Parameters**:
- `batteryId` (string, UUID) - Battery identifier

**Query Parameters**:
- `startTime` (string, ISO 8601, optional) - Range start (default: 24h ago)
- `endTime` (string, ISO 8601, optional) - Range end (default: now)
- `limit` (int, optional) - Max records to return (default: 100, max: 1000)
- `offset` (int, optional) - Pagination offset (default: 0)

#### Success Response (200 OK)

```json
{
  "data": [
    {
      "id": "state-xyz789",
      "battery_id": "battery-123",
      "soc": 75.5,
      "power": 25.0,
      "temperature": 28.5,
      "voltage": 800.0,
      "current": 31.25,
      "operation_state": "DISCHARGING",
      "custom_attributes": {},
      "timestamp": "2025-12-30T10:30:15Z",
      "created_at": "2025-12-30T10:30:15.234Z"
    },
    {
      "id": "state-xyz788",
      "battery_id": "battery-123",
      "soc": 75.6,
      "power": 25.0,
      "operation_state": "DISCHARGING",
      "timestamp": "2025-12-30T10:30:14Z",
      ...
    }
  ],
  "pagination": {
    "total": 3600,
    "limit": 100,
    "offset": 0
  }
}
```

#### Error Responses

**400 Bad Request** - Invalid time range
```json
{
  "error": "invalid request",
  "message": "startTime must be before endTime"
}
```

**400 Bad Request** - Invalid limit
```json
{
  "error": "invalid request",
  "message": "limit must be between 1 and 1000"
}
```

#### Examples

```bash
# Last hour of data
curl "http://localhost:8082/api/v1/telemetry/battery-123/history?startTime=2025-12-30T09:00:00Z&endTime=2025-12-30T10:00:00Z"

# Last 100 records (default)
curl http://localhost:8082/api/v1/telemetry/battery-123/history

# Pagination
curl "http://localhost:8082/api/v1/telemetry/battery-123/history?limit=50&offset=100"
```

---

### 1.3 Health Check

Verify Telemetry Service is running.

#### Request

```http
GET /health
```

#### Success Response (200 OK)

```json
{
  "status": "ok",
  "service": "telemetry",
  "version": "1.0.0",
  "timestamp": "2025-12-30T10:30:15Z"
}
```

---

## 2. Device Interface Service

**No REST API** - Purely event-driven service.

Communication happens exclusively via NATS events:
- **Subscribes**: ChargingCommandIssued, DischargingCommandIssued
- **Publishes**: ChargingStarted, ChargingCompleted, DischargingStarted, DischargingCompleted

---

## 3. Event Schemas

All events use **NATS pub/sub** with JSON payloads.

### 3.1 BatteryStateChanged (Published by Telemetry)

**Frequency**: 1 Hz (every 1 second per battery)

**Subject**: `battery.state.changed.v1`

**Publisher**: Telemetry Service

**Subscribers**: Bidding Service (M6), Device Interface Service

#### Payload

```json
{
  "battery_id": "battery-123",
  "soc": 75.5,
  "power": 25.0,
  "temperature": 28.5,
  "voltage": 800.0,
  "current": 31.25,
  "operation_state": "DISCHARGING",
  "custom_attributes": {
    "vendor": "Tesla",
    "cellVoltageMin": 3.2
  },
  "timestamp": "2025-12-30T10:30:15Z",
  "event_version": "v1"
}
```

#### Field Types

| Field | Type | Description |
|-------|------|-------------|
| `battery_id` | string (UUID) | Battery identifier |
| `soc` | float | State of Charge (0-100%) |
| `power` | float | MW (positive=discharge, negative=charge) |
| `temperature` | float | °C |
| `voltage` | float | Volts |
| `current` | float | Amperes |
| `operation_state` | string | IDLE, CHARGING, DISCHARGING, FCAS |
| `custom_attributes` | object | Manufacturer-specific data |
| `timestamp` | string (ISO 8601) | When state was measured |
| `event_version` | string | Event schema version ("v1") |

---

### 3.2 ChargingCommandIssued (Subscribed by Device Interface)

**Subject**: `charging.command.issued.v1`

**Publisher**: Bidding Service (M6, future)

**Subscriber**: Device Interface Service

#### Payload

```json
{
  "command_id": "cmd-abc123",
  "battery_id": "battery-123",
  "power": 15.0,
  "target_soc": 90.0,
  "duration_minutes": 60,
  "decision_mode": "SEMI_AUTO",
  "timestamp": "2025-12-30T10:30:15Z",
  "event_version": "v1"
}
```

#### Field Types

| Field | Type | Description |
|-------|------|-------------|
| `command_id` | string (UUID) | Unique command identifier |
| `battery_id` | string (UUID) | Target battery |
| `power` | float | Charging power in MW (absolute value) |
| `target_soc` | float | Stop charging at this SoC (0-100%) |
| `duration_minutes` | int | Max duration (0 = until target) |
| `decision_mode` | string | MANUAL, SEMI_AUTO, FULL_AUTO |
| `timestamp` | string (ISO 8601) | Command issued time |
| `event_version` | string | "v1" |

---

### 3.3 ChargingStarted (Published by Device Interface)

**Subject**: `charging.started.v1`

**Publisher**: Device Interface Service

**Subscribers**: Telemetry Service, Bidding Service

#### Payload

```json
{
  "charging_id": "charge-xyz789",
  "battery_id": "battery-123",
  "command_id": "cmd-abc123",
  "actual_power": 14.5,
  "target_soc": 90.0,
  "initial_soc": 75.0,
  "timestamp": "2025-12-30T10:30:16Z",
  "event_version": "v1"
}
```

#### Field Types

| Field | Type | Description |
|-------|------|-------------|
| `charging_id` | string (UUID) | Unique charging session ID |
| `battery_id` | string (UUID) | Battery identifier |
| `command_id` | string (UUID) | Original command ID |
| `actual_power` | float | Actual charging power (may differ from requested) |
| `target_soc` | float | Target SoC (0-100%) |
| `initial_soc` | float | SoC when charging started |
| `timestamp` | string (ISO 8601) | Charging start time |
| `event_version` | string | "v1" |

---

### 3.4 ChargingCompleted (Published by Device Interface)

**Subject**: `charging.completed.v1`

**Publisher**: Device Interface Service

**Subscribers**: Telemetry Service, Bidding Service

#### Payload

```json
{
  "charging_id": "charge-xyz789",
  "battery_id": "battery-123",
  "final_soc": 90.0,
  "energy_charged_mwh": 15.0,
  "duration_minutes": 58,
  "reason": "TARGET_REACHED",
  "timestamp": "2025-12-30T11:28:16Z",
  "event_version": "v1"
}
```

#### Reason Codes

| Code | Description |
|------|-------------|
| `TARGET_REACHED` | Target SoC achieved |
| `DURATION_EXCEEDED` | Max duration reached |
| `MANUAL_STOP` | Operator intervention |
| `ERROR` | Hardware fault or communication error |
| `CONFLICT` | FCAS dispatch received |

---

### 3.5 DischargingCommandIssued (Subscribed by Device Interface)

**Subject**: `discharging.command.issued.v1`

**Publisher**: Bidding Service (M6, future)

**Subscriber**: Device Interface Service

#### Payload

```json
{
  "command_id": "cmd-def456",
  "battery_id": "battery-123",
  "power": 25.0,
  "minimum_soc": 10.0,
  "duration_minutes": 120,
  "stop_conditions": {
    "price_threshold": 50.0,
    "min_soc": 10.0,
    "max_duration_minutes": 120
  },
  "timestamp": "2025-12-30T10:30:15Z",
  "event_version": "v1"
}
```

---

### 3.6 DischargingStarted (Published by Device Interface)

**Subject**: `discharging.started.v1`

**Publisher**: Device Interface Service

**Subscribers**: Telemetry Service, Bidding Service

#### Payload

```json
{
  "discharging_id": "discharge-uvw321",
  "battery_id": "battery-123",
  "command_id": "cmd-def456",
  "actual_power": 24.5,
  "initial_soc": 90.0,
  "minimum_soc": 10.0,
  "timestamp": "2025-12-30T10:30:16Z",
  "event_version": "v1"
}
```

---

### 3.7 DischargingCompleted (Published by Device Interface)

**Subject**: `discharging.completed.v1`

**Publisher**: Device Interface Service

**Subscribers**: Telemetry Service, Bidding Service

#### Payload

```json
{
  "discharging_id": "discharge-uvw321",
  "battery_id": "battery-123",
  "final_soc": 15.0,
  "energy_discharged_mwh": 75.0,
  "duration_minutes": 120,
  "reason": "DURATION_EXCEEDED",
  "timestamp": "2025-12-30T12:30:16Z",
  "event_version": "v1"
}
```

---

## 4. Event Publishing Patterns

### 4.1 High-Frequency Publishing (BatteryStateChanged)

**Challenge**: 1 Hz = 86,400 events/day per battery

**Pattern**: Background goroutine with ticker

```go
func (s *StatePublisher) Start(ctx context.Context, batteryID string) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            state, err := s.adapter.GetState(ctx)
            if err != nil {
                log.Printf("Failed to get state: %v", err)
                continue
            }

            if err := s.publishState(ctx, batteryID, state); err != nil {
                log.Printf("Failed to publish state: %v", err)
                // Don't crash - best-effort pattern
            }

        case <-ctx.Done():
            return
        }
    }
}
```

**Key Points**:
- Ticker ensures 1 Hz consistency
- Context cancellation for graceful shutdown
- Best-effort: Log errors but continue
- 2-second timeout on publish

### 4.2 Command Event Subscription

**Pattern**: Wildcard subscription with routing

```go
handler := func(subject string, data []byte) error {
    switch subject {
    case "charging.command.issued.v1":
        var event events.ChargingCommandIssued
        json.Unmarshal(data, &event)
        return h.handleChargingCommand(event)

    case "discharging.command.issued.v1":
        var event events.DischargingCommandIssued
        json.Unmarshal(data, &event)
        return h.handleDischargingCommand(event)

    default:
        return fmt.Errorf("unknown subject: %s", subject)
    }
}

subscriber.Subscribe(ctx, "*.command.issued.v1", handler)
```

---

## 5. Performance Considerations

### 5.1 Event Load Estimation

**Scenario**: 10 batteries in operation

| Event | Frequency | Events/sec | Events/day | Bandwidth |
|-------|-----------|------------|------------|-----------|
| BatteryStateChanged | 1 Hz × 10 | 10 | 864,000 | ~10 KB/s |
| ChargingStarted | ~5/day × 10 | 0.001 | 50 | Negligible |
| ChargingCompleted | ~5/day × 10 | 0.001 | 50 | Negligible |
| DischargingStarted | ~5/day × 10 | 0.001 | 50 | Negligible |
| DischargingCompleted | ~5/day × 10 | 0.001 | 50 | Negligible |

**Total**: ~10 events/sec, ~10 KB/s bandwidth

**NATS Capacity**: Millions of messages/second

**Conclusion**: Well within NATS capabilities

### 5.2 Database Write Load

**Telemetry Service**: 10 inserts/second (1 Hz × 10 batteries)

**PostgreSQL Capacity**: 10,000+ writes/second on modern hardware

**Optimization**: Batch inserts (future)

### 5.3 API Response Times

**Target Latencies**:
- `GET /telemetry/:id/current` - <50ms (single row query)
- `GET /telemetry/:id/history` - <200ms (time-range query)

**Optimization**: Index on `(battery_id, timestamp DESC)`

---

## 6. Error Handling

### 6.1 REST API Error Format

All errors return:

```json
{
  "error": "error_code",
  "message": "Human-readable description",
  "details": {} // Optional additional context
}
```

### 6.2 Event Publishing Errors

**Strategy**: Best-effort with logging

```go
if err := publisher.Publish(ctx, subject, event); err != nil {
    log.Printf("WARNING: Failed to publish %s: %v", subject, err)
    // Don't fail business operation
}
```

**Rationale**: Event publishing failures shouldn't block telemetry collection or command execution.

---

## 7. Integration Examples

### 7.1 End-to-End Charging Flow

```bash
# Terminal 1: Subscribe to all events
cd tools/event-subscriber
go run main.go ">"

# Terminal 2: Create battery (Asset Management)
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{"capacity": 100, "max_power": 50, ...}'

# See event: battery.registered.v1

# Terminal 3: Start Telemetry Service
cd services/telemetry
go run cmd/server/main.go

# See events: battery.state.changed.v1 (every 1 second)

# Terminal 4: Send charging command (manual test)
# (In M6, Bidding Service will publish this automatically)
nats pub charging.command.issued.v1 '{"battery_id": "...", "power": 15, ...}'

# See events in sequence:
# 1. charging.started.v1
# 2. battery.state.changed.v1 (SoC increasing)
# 3. charging.completed.v1 (target reached)
```

### 7.2 Query Current State

```bash
# Get latest telemetry
curl http://localhost:8082/api/v1/telemetry/battery-123/current

# Response shows real-time state
{
  "soc": 85.5,
  "power": -15.0,
  "operation_state": "CHARGING"
}
```

---

## 8. Testing Endpoints

### 8.1 Manual Testing with curl

```bash
# Health check
curl http://localhost:8082/health

# Current state (after telemetry data exists)
curl http://localhost:8082/api/v1/telemetry/battery-123/current

# History (last hour)
curl "http://localhost:8082/api/v1/telemetry/battery-123/history?limit=10"

# With time range
START=$(date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8082/api/v1/telemetry/battery-123/history?startTime=$START&endTime=$END"
```

### 8.2 Event Testing with NATS CLI

```bash
# Install NATS CLI
brew install nats-io/nats-tools/nats

# Subscribe to all battery events
nats sub "battery.>"

# Publish test event
nats pub battery.state.changed.v1 '{"battery_id": "test", "soc": 50, ...}'
```

---

## 9. Related Documentation

- [M5 Overview](./M5-OVERVIEW.md) - Architecture and big picture
- [M5 Domain Spec](./M5-DOMAIN-SPEC.md) - Domain models and validation
- [M5 Checklist](./M5-CHECKLIST.md) - Implementation steps
- [pkg/events README](../../pkg/events/README.md) - Event library usage
- [EVENTS.md](../../EVENTS.md) - Complete event catalog

---

**API Complexity**: Low (2 endpoints) + High-frequency events

**Event Count**: 7 events (3 publishing, 2 subscribing, 2 lifecycle)

**Performance**: Optimized for 1 Hz publishing and sub-50ms API responses
