# M6: Bidding Service - API Specification

## Overview

The Bidding Service is **event-driven only** with no REST API in M6. It subscribes to events from other services and publishes events to signal arbitrage opportunities and commands.

Future M7+ may add a REST API for:
- Query historical decisions
- Override automation mode
- Manual command issuance

## Event Subscriptions

The Bidding Service subscribes to 3 event types from other services:

### 1. battery.state.changed.v1

**Publisher**: Telemetry Service
**Frequency**: 1 Hz (every 1 second)
**Purpose**: Monitor real-time battery state for arbitrage decisions

**Schema**:
```json
{
  "battery_id": "string",
  "soc": "number (0-100)",
  "power": "number (MW, positive=discharge, negative=charge)",
  "temperature": "number (°C)",
  "voltage": "number (V)",
  "current": "number (A)",
  "operation_state": "IDLE | CHARGING | DISCHARGING | FCAS",
  "timestamp": "ISO 8601 string",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "battery_id": "battery-123",
  "soc": 50.0,
  "power": 0.0,
  "temperature": 25.3,
  "voltage": 800.0,
  "current": 0.0,
  "operation_state": "IDLE",
  "timestamp": "2025-12-30T10:00:00Z",
  "event_version": "v1"
}
```

**Handler**: `BiddingEngine.HandleBatteryStateChanged()`
**Actions**:
- Update `BatteryStateCache` with latest state
- Run arbitrage algorithm if price is available
- Publish opportunity events if conditions met

---

### 2. market.price.updated.v1

**Publisher**: Market Data Service
**Frequency**: Variable (on price creation)
**Purpose**: Monitor market price changes for arbitrage opportunities

**Schema**:
```json
{
  "price_id": "string (UUID)",
  "timestamp": "ISO 8601 string",
  "value": "number ($/MWh)",
  "interval": "number (minutes)",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "price_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-12-30T10:00:00Z",
  "value": 30.0,
  "interval": 5,
  "event_version": "v1"
}
```

**Handler**: `BiddingEngine.HandleMarketPriceUpdated()`
**Actions**:
- Update `PriceCache` with latest price
- For each battery in `BatteryStateCache`, run arbitrage algorithm
- Publish opportunity events for all batteries meeting conditions

---

### 3. battery.registered.v1

**Publisher**: Asset Management Service
**Frequency**: On battery creation
**Purpose**: Initialize battery in bidding system

**Schema**:
```json
{
  "battery_id": "string (UUID)",
  "capacity": "number (MWh)",
  "max_power": "number (MW)",
  "ramp_rate": "number (MW/min)",
  "efficiency": "number (0-1)",
  "location": "string (NSW, VIC, QLD, SA, TAS)",
  "manufacturer": "string",
  "constraints": {
    "warranty_eol": "number (MWh)",
    "max_cycles": "number",
    "temp_min": "number (°C)",
    "temp_max": "number (°C)",
    "grid_compliance": ["string (FCAS, ENERGY, ...)"]
  },
  "timestamp": "ISO 8601 string",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "battery_id": "battery-123",
  "capacity": 200.0,
  "max_power": 100.0,
  "ramp_rate": 10.0,
  "efficiency": 0.95,
  "location": "SA",
  "manufacturer": "Tesla",
  "constraints": {
    "warranty_eol": 8000.0,
    "max_cycles": 10000,
    "temp_min": -20.0,
    "temp_max": 60.0,
    "grid_compliance": ["FCAS", "ENERGY"]
  },
  "timestamp": "2025-12-30T09:00:00Z",
  "event_version": "v1"
}
```

**Handler**: `BiddingEngine.HandleBatteryRegistered()`
**Actions**:
- Initialize battery in `BatteryStateCache` with default state (SoC=50%, IDLE)
- Store battery capacity and efficiency for profit calculations
- Battery is now eligible for arbitrage decisions

---

## Event Publications

The Bidding Service publishes 4 new event types:

### 1. charging.opportunity.detected.v1

**Trigger**: Price < $50/MWh AND SoC < 80% AND battery is IDLE
**Subscribers**: Future UI service, Analytics service
**Purpose**: Signal profitable charging opportunity detected

**Schema**:
```json
{
  "battery_id": "string",
  "price": "number ($/MWh)",
  "soc": "number (0-100)",
  "target_soc": "number (0-100, typically 80)",
  "expected_profit": "number ($)",
  "timestamp": "ISO 8601 string",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "battery_id": "battery-123",
  "price": 30.0,
  "soc": 50.0,
  "target_soc": 80.0,
  "expected_profit": 3800.0,
  "timestamp": "2025-12-30T10:00:05Z",
  "event_version": "v1"
}
```

**Profit Calculation**:
```
price_delta = CHARGE_THRESHOLD - current_price = 50 - 30 = $20/MWh
expected_profit = price_delta * capacity * efficiency
                = 20 * 200 * 0.95 = $3,800
```

---

### 2. discharging.opportunity.detected.v1

**Trigger**: Price > $100/MWh AND SoC > 30% AND battery is IDLE
**Subscribers**: Future UI service, Analytics service
**Purpose**: Signal profitable discharging opportunity detected

**Schema**:
```json
{
  "battery_id": "string",
  "price": "number ($/MWh)",
  "soc": "number (0-100)",
  "target_soc": "number (0-100, typically 30)",
  "expected_profit": "number ($)",
  "stop_conditions": {
    "price_threshold": "number ($/MWh, optional)",
    "duration": "number (minutes, optional)",
    "fcas_dispatch": "boolean"
  },
  "timestamp": "ISO 8601 string",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "battery_id": "battery-123",
  "price": 150.0,
  "soc": 70.0,
  "target_soc": 30.0,
  "expected_profit": 9500.0,
  "stop_conditions": {
    "price_threshold": 100.0,
    "duration": null,
    "fcas_dispatch": false
  },
  "timestamp": "2025-12-30T14:00:05Z",
  "event_version": "v1"
}
```

**Profit Calculation**:
```
price_delta = current_price - DISCHARGE_THRESHOLD = 150 - 100 = $50/MWh
expected_profit = price_delta * capacity * efficiency
                = 50 * 200 * 0.95 = $9,500
```

---

### 3. charging.command.issued.v1

**Trigger**: Charging opportunity detected AND automation mode = FULL_AUTO
**Subscribers**: Device Interface Service
**Purpose**: Command battery to start charging

**Schema**:
```json
{
  "battery_id": "string",
  "target_soc": "number (0-100)",
  "max_power": "number (MW)",
  "reason": "string (human-readable)",
  "issued_by": "string (BIDDING_SERVICE_AUTO | operator_id)",
  "timestamp": "ISO 8601 string",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "battery_id": "battery-123",
  "target_soc": 80.0,
  "max_power": 100.0,
  "reason": "Low price arbitrage: $30/MWh < $50/MWh threshold",
  "issued_by": "BIDDING_SERVICE_AUTO",
  "timestamp": "2025-12-30T10:00:05Z",
  "event_version": "v1"
}
```

**Notes**:
- Only published in **FULL_AUTO** mode
- Device Interface Service executes the command
- `max_power` comes from battery registration (BatteryRegistered event)

---

### 4. discharging.command.issued.v1

**Trigger**: Discharging opportunity detected AND automation mode = FULL_AUTO
**Subscribers**: Device Interface Service
**Purpose**: Command battery to start discharging

**Schema**:
```json
{
  "battery_id": "string",
  "power": "number (MW, positive for discharge)",
  "stop_conditions": {
    "price_threshold": "number ($/MWh, stop if price drops below)",
    "target_soc": "number (0-100, stop if reached)",
    "duration": "number (minutes, stop after elapsed)",
    "fcas_dispatch": "boolean (stop if FCAS dispatch received)"
  },
  "reason": "string (human-readable)",
  "issued_by": "string (BIDDING_SERVICE_AUTO | operator_id)",
  "timestamp": "ISO 8601 string",
  "event_version": "v1"
}
```

**Example**:
```json
{
  "battery_id": "battery-123",
  "power": 100.0,
  "stop_conditions": {
    "price_threshold": 100.0,
    "target_soc": 30.0,
    "duration": null,
    "fcas_dispatch": false
  },
  "reason": "High price arbitrage: $150/MWh > $100/MWh threshold",
  "issued_by": "BIDDING_SERVICE_AUTO",
  "timestamp": "2025-12-30T14:00:05Z",
  "event_version": "v1"
}
```

**Notes**:
- Only published in **FULL_AUTO** mode
- Multiple stop conditions supported (any one triggers stop)
- `price_threshold` enables adaptive discharge (stop if price drops)

---

## Event Flow Diagrams

### Charging Scenario (FULL_AUTO)

```
Time: 10:00:00
┌──────────────────┐
│  Market Data     │ ──> market.price.updated.v1
│  (Price: $30)    │     { value: 30.0 }
└──────────────────┘
         │
         v
┌──────────────────┐
│ Bidding Service  │
│ Price < $50 ?    │ ──> Yes, but need battery state
└──────────────────┘


Time: 10:00:01 (1 second later)
┌──────────────────┐
│   Telemetry      │ ──> battery.state.changed.v1
│  (SoC: 50%)      │     { soc: 50.0, operation_state: "IDLE" }
└──────────────────┘
         │
         v
┌──────────────────┐
│ Bidding Service  │
│ SoC < 80%? ✅    │
│ State IDLE? ✅   │
│ Price < $50? ✅  │
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Publish Events   │ ──> 1. charging.opportunity.detected.v1
└──────────────────┘     { price: 30.0, soc: 50.0, expected_profit: 3800.0 }
       │
       v
┌──────────────────┐
│ Check Mode       │ ──> FULL_AUTO
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Publish Command  │ ──> 2. charging.command.issued.v1
└──────────────────┘     { target_soc: 80.0, max_power: 100.0 }
       │
       v
┌──────────────────┐
│ Device Interface │ ──> Executes charging command
└──────────────────┘
```

### Discharging Scenario (MANUAL)

```
Time: 14:00:00
┌──────────────────┐
│  Market Data     │ ──> market.price.updated.v1
│  (Price: $150)   │     { value: 150.0 }
└──────────────────┘
         │
         v
┌──────────────────┐
│ Bidding Service  │
│ Price > $100? ✅ │
│ Has battery state│
│ from cache ✅    │
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Check Conditions │
│ SoC > 30%? ✅    │
│ State IDLE? ✅   │
└──────┬───────────┘
       │
       v
┌──────────────────┐
│ Publish Event    │ ──> discharging.opportunity.detected.v1
└──────────────────┘     { price: 150.0, soc: 70.0, expected_profit: 9500.0 }
       │
       v
┌──────────────────┐
│ Check Mode       │ ──> MANUAL (no command issued)
└──────────────────┘
       │
       v
┌──────────────────┐
│  Operator UI     │ ──> Human reviews and manually approves
│  (Future M7+)    │
└──────────────────┘
```

### No Opportunity (Mid-Range Price)

```
Time: 12:00:00
┌──────────────────┐
│  Market Data     │ ──> market.price.updated.v1
│  (Price: $75)    │     { value: 75.0 }
└──────────────────┘
         │
         v
┌──────────────────┐
│ Bidding Service  │
│ Price < $50? ❌  │ (too high for charging)
│ Price > $100? ❌ │ (too low for discharging)
└──────────────────┘
       │
       v
┌──────────────────┐
│ No Action        │ (no events published)
└──────────────────┘
```

---

## NATS Subject Patterns

### Subscriptions

```go
// Subscribe to all battery state changes (1 Hz high-frequency)
subscriber.Subscribe(ctx, "battery.state.changed.v1", handleBatteryStateChanged)

// Subscribe to market price updates
subscriber.Subscribe(ctx, "market.price.updated.v1", handleMarketPriceUpdated)

// Subscribe to battery registrations
subscriber.Subscribe(ctx, "battery.registered.v1", handleBatteryRegistered)
```

### Publications

```go
// Publish opportunity events (always)
publisher.Publish(ctx, "charging.opportunity.detected.v1", opportunityEvent)
publisher.Publish(ctx, "discharging.opportunity.detected.v1", opportunityEvent)

// Publish command events (only if FULL_AUTO)
if automationMode == "FULL_AUTO" {
    publisher.Publish(ctx, "charging.command.issued.v1", commandEvent)
    publisher.Publish(ctx, "discharging.command.issued.v1", commandEvent)
}
```

---

## Configuration

The Bidding Service is configured via environment variables:

```bash
# NATS connection
NATS_URL=nats://localhost:4222

# Automation mode
AUTOMATION_MODE=MANUAL | SEMI_AUTO | FULL_AUTO
# Default: MANUAL

# Logging
LOG_LEVEL=info | debug | warn | error
# Default: info
```

**Docker Compose Example**:
```yaml
bidding:
  build: ./services/bidding
  container_name: bidding
  environment:
    NATS_URL: nats://nats:4222
    AUTOMATION_MODE: FULL_AUTO
    LOG_LEVEL: info
  depends_on:
    - nats
  restart: unless-stopped
```

---

## Event Versioning Strategy

All events use `v1` versioning in M6.

**Backward-Compatible Changes** (keep v1):
- Add new optional fields
- Add new enum values (if consumers ignore unknown values)
- Extend nested objects with optional fields

**Breaking Changes** (create v2):
- Remove fields
- Rename fields
- Change field types
- Make optional fields required
- Change enum value meanings

**Example Evolution** (Future):
```go
// M6: v1
type ChargingOpportunityDetected struct {
    BatteryID      string  `json:"battery_id"`
    Price          float64 `json:"price"`
    SoC            float64 `json:"soc"`
    // ...
    EventVersion   string  `json:"event_version"` // "v1"
}

// M7: v1 (backward compatible - added optional field)
type ChargingOpportunityDetected struct {
    BatteryID      string  `json:"battery_id"`
    Price          float64 `json:"price"`
    SoC            float64 `json:"soc"`
    MLConfidence   float64 `json:"ml_confidence,omitempty"` // NEW (optional)
    // ...
    EventVersion   string  `json:"event_version"` // still "v1"
}

// M8: v2 (breaking change - renamed field)
type ChargingOpportunityDetected struct {
    BatteryID      string  `json:"battery_id"`
    MarketPrice    float64 `json:"market_price"` // RENAMED from "price"
    CurrentSoC     float64 `json:"current_soc"`  // RENAMED from "soc"
    // ...
    EventVersion   string  `json:"event_version"` // "v2"
}
```

---

## High-Frequency Event Handling

The Bidding Service receives `BatteryStateChanged` events at **1 Hz** (every second). This requires efficient processing:

### Performance Considerations

1. **In-Memory Caching**: All state stored in RAM (no database lookups)
2. **Thread-Safe Access**: `sync.RWMutex` for concurrent reads/writes
3. **Fast Arbitrage Algorithm**: O(1) decision logic (simple threshold checks)
4. **Buffered Channels**: Handle burst traffic
5. **Best-Effort Publishing**: Don't block on event publication failures

### Example Processing Time Target

```
Event arrives → Update cache → Run algorithm → Publish events
   |               |              |               |
   |<─ 5ms ───────>|<─ 1ms ───>|<─── 10ms ────>|
   |                                              |
   |<──────────── Total: < 20ms ─────────────────>|
```

**Why this matters**:
- 1 Hz = 1000ms between events
- Target 20ms processing = 2% CPU utilization
- Leaves 98% for handling bursts and multiple batteries

### Scalability

```
1 battery @ 1 Hz  = 86,400 events/day = 3,600 events/hour
10 batteries      = 864,000 events/day = 36,000 events/hour
100 batteries     = 8,640,000 events/day = 360,000 events/hour

Estimated throughput: 10,000 events/second (with proper caching)
```

---

## Testing Event Flows

Use the `tools/event-subscriber` CLI to monitor events:

```bash
# Terminal 1: Subscribe to all events
cd tools/event-subscriber
go run main.go ">"

# Terminal 2: Subscribe to only opportunity events
go run main.go "*.opportunity.*"

# Terminal 3: Subscribe to only command events
go run main.go "*.command.*"

# Terminal 4: Subscribe to all bidding events
go run main.go "charging.>,discharging.>"
```

**Example Output**:
```
[2025-12-30 10:00:05] market.price.updated.v1
{
  "value": 30.0,
  "timestamp": "2025-12-30T10:00:00Z"
}

[2025-12-30 10:00:05] charging.opportunity.detected.v1
{
  "battery_id": "battery-123",
  "price": 30.0,
  "soc": 50.0,
  "expected_profit": 3800.0
}

[2025-12-30 10:00:05] charging.command.issued.v1
{
  "battery_id": "battery-123",
  "target_soc": 80.0,
  "issued_by": "BIDDING_SERVICE_AUTO"
}
```

---

## Future REST API (M7+)

Potential REST endpoints for future milestones:

```
GET  /api/v1/decisions              # List recent bidding decisions
GET  /api/v1/decisions/:id          # Get specific decision
GET  /api/v1/opportunities/active   # List current opportunities
POST /api/v1/commands/charging      # Manual charging command
POST /api/v1/commands/discharging   # Manual discharging command
PUT  /api/v1/config/automation-mode # Change automation mode
GET  /api/v1/health                 # Health check
```

**Not included in M6** - focus is on event-driven architecture first.

---

**The Bidding Service communicates exclusively through events, enabling loose coupling and independent scalability.** 🚀📡
