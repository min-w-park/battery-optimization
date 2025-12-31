# Demo Scenario: End-to-End Arbitrage Flow

This document provides a step-by-step demo of the battery optimization system, demonstrating the full event-driven arbitrage flow from battery registration to automated charging/discharging.

## Prerequisites

- Docker & Docker Compose installed
- curl installed for REST API testing
- Basic understanding of event-driven architecture

## Setup

### Terminal 1: Start All Services

```bash
# Start infrastructure + all services
docker-compose up -d

# Verify all services are healthy
docker-compose ps

# Expected output: All services showing "healthy" status
# - nats
# - asset-db, market-db, telemetry-db
# - asset-management, market-data, telemetry, device-interface, bidding
```

### Terminal 2: Event Subscriber (Monitor Events)

```bash
cd tools/event-subscriber
go run main.go ">"

# This will show ALL events flowing through the system
# Events will be color-coded:
# - Green: Timestamps and counters
# - Blue: Event subjects
# - Purple: JSON payloads
```

### Terminal 3: REST API Commands

Keep this terminal open for running curl commands below.

---

## Step 1: Register a Battery

**Action**: Create a new battery in the Asset Management Service.

```bash
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{
    "capacity": 200.0,
    "maxPower": 100.0,
    "rampRate": 10.0,
    "efficiency": 0.95,
    "location": "SA",
    "manufacturer": "Tesla",
    "constraints": {
      "warrantyEol": 8000.0,
      "maxCycles": 10000,
      "tempMin": -20.0,
      "tempMax": 60.0,
      "gridCompliance": ["FCAS", "ENERGY"]
    }
  }'
```

**Expected Response** (201 Created):
```json
{
  "id": "01JGTX...",
  "capacity": 200.0,
  "maxPower": 100.0,
  ...
}
```

**Expected Events** (in Terminal 2):
```
[Event #1] battery.registered.v1
{
  "battery_id": "01JGTX...",
  "capacity": 200.0,
  "max_power": 100.0,
  "ramp_rate": 10.0,
  "efficiency": 0.95,
  "location": "SA",
  "manufacturer": "Tesla",
  ...
}
```

**What Happened**:
1. Asset Management Service saved battery to `asset-db`
2. Published `battery.registered.v1` event to NATS
3. Bidding Service subscribed to this event and cached battery info
4. Battery is now known to the bidding engine (initial state: SoC=50%, IDLE)

---

## Step 2: Publish Low Market Price (Charging Opportunity)

**Action**: Create a low price in the Market Data Service ($30/MWh, below $50 threshold).

```bash
curl -X POST http://localhost:8081/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2025-12-31T10:00:00Z",
    "value": 30.0,
    "interval": 5
  }'
```

**Expected Response** (201 Created):
```json
{
  "id": "01JGTX...",
  "timestamp": "2025-12-31T10:00:00Z",
  "value": 30.0,
  "interval": 5
}
```

**Expected Events**:
```
[Event #2] market.price.updated.v1
{
  "price_id": "01JGTX...",
  "region": "SA",
  "price": 30.0,
  "demand": 2000.0,
  "interval_type": "5MIN_PREDISPATCH",
  ...
}
```

**What Happened**:
1. Market Data Service saved price to `market-db`
2. Published `market.price.updated.v1` event to NATS
3. Bidding Service subscribed to this event and updated price cache
4. Bidding Service re-evaluated all cached batteries with new price
5. **No opportunity detected yet** - waiting for battery state update

---

## Step 3: Simulate Battery State Update

Since the Bidding Service needs both price AND battery state to make decisions, we need to trigger a battery state update.

**Option A: Using Integration Test Script** (Recommended for Demo)

```bash
# In Terminal 3, run the integration test
cd services/bidding
go run test_integration.go
```

This script will:
1. Register a test battery
2. Publish low price ($30/MWh)
3. Publish battery state (SoC=50%, IDLE)
4. Publish high price ($150/MWh)
5. Publish battery state (SoC=70%, IDLE)

**Option B: Manual Event Publishing** (Advanced)

If you want to manually publish battery state events, you can use the NATS client or modify the integration test script.

**Expected Events** (assuming SoC=50%, IDLE, price=$30):

```
[Event #3] battery.state.changed.v1
{
  "battery_id": "01JGTX...",
  "soc": 50.0,
  "power": 0.0,
  "operation_state": "IDLE",
  "temperature": 25.0,
  ...
}

[Event #4] charging.opportunity.detected.v1
{
  "battery_id": "01JGTX...",
  "price": 30.0,
  "soc": 50.0,
  "target_soc": 80.0,
  "expected_profit": 950.0,  # Simplified calculation
  "timestamp": "2025-12-31T10:00:01Z",
  "event_version": "v1"
}

[Event #5] charging.command.issued.v1  # Only in FULL_AUTO mode
{
  "battery_id": "01JGTX...",
  "target_soc": 80.0,
  "max_power": 100.0,
  "reason": "Low market price: $30.00/MWh (threshold: $50.00/MWh)",
  "issued_by": "bidding-service",
  "timestamp": "2025-12-31T10:00:01Z",
  "event_version": "v1"
}

[Event #6] charging.started.v1  # If Device Interface is running
{
  "battery_id": "01JGTX...",
  "target_soc": 80.0,
  "actual_power": 95.0,  # Considering efficiency
  "estimated_duration": 3600.0,
  ...
}
```

**What Happened** (Bidding Engine Logic):
1. Received `battery.state.changed.v1` event
2. Updated battery state cache (SoC=50%, IDLE)
3. Retrieved latest price from cache ($30/MWh)
4. **Arbitrage Algorithm Evaluation**:
   - Price ($30) < CHARGE_THRESHOLD ($50)? ✅ Yes
   - SoC (50%) < MAX_SOC_FOR_CHARGE (80%)? ✅ Yes
   - Operation State = IDLE? ✅ Yes
   - **Result**: Should charge!
5. Published `charging.opportunity.detected.v1`
6. **FULL_AUTO Mode**: Automatically published `charging.command.issued.v1`
7. Device Interface subscribed to command and started charging

---

## Step 4: Publish High Market Price (Discharging Opportunity)

**Action**: Create a high price in Market Data Service ($150/MWh, above $100 threshold).

```bash
curl -X POST http://localhost:8081/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2025-12-31T14:00:00Z",
    "value": 150.0,
    "interval": 5
  }'
```

**Expected Events** (assuming battery is now at SoC=70%, IDLE):

```
[Event #7] market.price.updated.v1
{
  "price": 150.0,
  ...
}

[Event #8] discharging.opportunity.detected.v1
{
  "battery_id": "01JGTX...",
  "price": 150.0,
  "soc": 70.0,
  "target_soc": 30.0,
  "expected_profit": 2000.0,
  "stop_conditions": {
    "min_soc": 30.0,
    "price_threshold": 100.0
  },
  ...
}

[Event #9] discharging.command.issued.v1  # Only in FULL_AUTO mode
{
  "battery_id": "01JGTX...",
  "power": 100.0,
  "stop_conditions": {
    "price_threshold": 100.0,
    "soc_threshold": 30.0,
    "duration_seconds": 14400.0,
    "fcas_dispatch": false
  },
  "reason": "High market price: $150.00/MWh (threshold: $100.00/MWh)",
  ...
}

[Event #10] discharging.started.v1  # If Device Interface is running
{
  "battery_id": "01JGTX...",
  "power": 100.0,
  "estimated_duration": 14400.0,
  ...
}
```

**What Happened** (Bidding Engine Logic):
1. Received `market.price.updated.v1` event
2. Updated price cache ($150/MWh)
3. **Re-evaluated ALL cached batteries** with new price
4. **Arbitrage Algorithm Evaluation**:
   - Price ($150) > DISCHARGE_THRESHOLD ($100)? ✅ Yes
   - SoC (70%) > MIN_SOC_FOR_DISCHARGE (30%)? ✅ Yes
   - Operation State = IDLE? ✅ Yes
   - **Result**: Should discharge!
5. Published `discharging.opportunity.detected.v1`
6. **FULL_AUTO Mode**: Automatically published `discharging.command.issued.v1`
7. Device Interface subscribed to command and started discharging

---

## Step 5: Test Mid-Range Price (No Opportunity)

**Action**: Create a mid-range price ($75/MWh, between $50 and $100).

```bash
curl -X POST http://localhost:8081/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2025-12-31T18:00:00Z",
    "value": 75.0,
    "interval": 5
  }'
```

**Expected Events**:
```
[Event #11] market.price.updated.v1
{
  "price": 75.0,
  ...
}

# NO opportunity events published!
# Price is between $50 and $100, so no arbitrage opportunity exists
```

**What Happened**:
1. Bidding Service updated price cache ($75/MWh)
2. Re-evaluated all batteries
3. **Arbitrage Algorithm Evaluation**:
   - Price ($75) < CHARGE_THRESHOLD ($50)? ❌ No
   - Price ($75) > DISCHARGE_THRESHOLD ($100)? ❌ No
   - **Result**: No action (price is in the "dead zone")
4. No events published

---

## Automation Modes Comparison

### MANUAL Mode (Detect Only)

```bash
# Stop bidding service
docker-compose stop bidding

# Restart with MANUAL mode
docker-compose up -d bidding
# (docker-compose.yml default is FULL_AUTO, so you'd need to modify it or use environment override)

# Or run locally:
cd services/bidding
AUTOMATION_MODE=MANUAL go run cmd/server/main.go
```

**Behavior**:
- ✅ Publishes `charging.opportunity.detected.v1`
- ✅ Publishes `discharging.opportunity.detected.v1`
- ❌ Does NOT publish `*.command.issued.v1` events
- Human operator must manually approve commands

**Use Case**: Production environments where human oversight is required before executing trades.

### FULL_AUTO Mode (Detect + Execute)

```bash
# Default mode in docker-compose.yml
docker-compose up -d bidding

# Or run locally:
cd services/bidding
AUTOMATION_MODE=FULL_AUTO go run cmd/server/main.go
```

**Behavior**:
- ✅ Publishes `charging.opportunity.detected.v1`
- ✅ Publishes `discharging.opportunity.detected.v1`
- ✅ Publishes `charging.command.issued.v1`
- ✅ Publishes `discharging.command.issued.v1`
- Device Interface automatically executes commands

**Use Case**: Lights-out operation for trusted algorithms.

---

## Event Flow Diagram

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│ Asset Mgmt      │       │ Market Data     │       │ Telemetry       │
│ Service         │       │ Service         │       │ Service         │
└────────┬────────┘       └────────┬────────┘       └────────┬────────┘
         │                         │                         │
         │ battery.registered.v1   │                         │
         ├────────────────────────►│                         │
         │                         │ market.price.updated.v1 │
         │                         ├────────────────────────►│
         │                         │                         │ battery.state.changed.v1
         │                         │                         ├───────────────────────┐
         │                         │                         │                       │
         │                         │                         │                       ▼
         │                         │                  ┌──────┴──────────────┐
         │                         │                  │ Bidding Service     │
         │                         │                  │ (Arbitrage Engine)  │
         │                         │                  └──────┬──────────────┘
         │                         │                         │
         │                         │ charging.opportunity.detected.v1
         │                         │◄────────────────────────┤
         │                         │                         │
         │                         │ charging.command.issued.v1 (FULL_AUTO)
         │                         │◄────────────────────────┤
         │                         │                         │
         │                         │                         ▼
         │                         │                  ┌──────────────────┐
         │                         │                  │ Device Interface │
         │                         │                  │ Service          │
         │                         │                  └──────┬───────────┘
         │                         │                         │
         │                         │ charging.started.v1     │
         │                         │◄────────────────────────┤
         │                         │                         │
```

---

## Troubleshooting

### No Events Appearing in Subscriber?

```bash
# 1. Check NATS health
curl http://localhost:8222/healthz
# Expected: 200 OK

# 2. Check NATS message count
curl http://localhost:8222/varz | jq .in_msgs
# Should be increasing as events flow

# 3. Check service logs
docker-compose logs bidding
docker-compose logs asset-management

# 4. Verify services are running
docker-compose ps
# All should show "Up" status
```

### Opportunities Not Detected?

```bash
# 1. Check price thresholds
cat services/bidding/internal/domain/arbitrage.go | grep "THRESHOLD"
# CHARGE_THRESHOLD_MWH = 50.0
# DISCHARGE_THRESHOLD_MWH = 100.0

# 2. Verify battery state
# - Must be in IDLE state (not already charging/discharging)
# - SoC must be < 80% for charging
# - SoC must be > 30% for discharging

# 3. Check bidding service logs for warnings
docker-compose logs bidding | grep WARNING

# 4. Verify automation mode
docker-compose exec bidding printenv | grep AUTOMATION_MODE
# Should show FULL_AUTO or MANUAL
```

### Commands Not Executed?

```bash
# 1. Check if Device Interface is running
docker-compose ps device-interface

# 2. Verify subscription to command events
docker-compose logs device-interface | grep "Subscribed"
# Should see: "Subscribed to: charging.command.issued.v1"

# 3. Check for command processing errors
docker-compose logs device-interface | grep ERROR
```

### Fresh Start

```bash
# Stop all services and remove data
docker-compose down -v

# Restart from scratch
docker-compose up -d

# Wait for healthy status
docker-compose ps
```

---

## Next Steps

After completing this demo:

1. **Explore Event Catalog**: See [EVENTS.md](../EVENTS.md) for complete event schemas
2. **Read Service Documentation**: Each service has a README (e.g., `services/bidding/README.md`)
3. **Modify Arbitrage Algorithm**: Try changing price thresholds in `services/bidding/internal/domain/arbitrage.go`
4. **Add New Subscribers**: Create a custom service that subscribes to opportunity events
5. **Test Automation Modes**: Compare MANUAL vs FULL_AUTO behavior

---

## Summary

This demo showcased:
- ✅ Event-driven microservices architecture
- ✅ DB-per-service pattern (3 PostgreSQL databases)
- ✅ NATS pub/sub event bus
- ✅ Real-time arbitrage decision engine
- ✅ Automation modes (MANUAL vs FULL_AUTO)
- ✅ End-to-end event flow from registration to command execution

**Key Insight**: Services are loosely coupled via events. Adding new services or modifying existing ones doesn't require changing other services—just subscribe to the events you need.
