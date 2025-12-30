# M5: Telemetry + Device Interface - Overview

## 🎯 Goal

Implement hardware abstraction and real-time battery state monitoring to enable:
1. **Telemetry Service** - Real-time battery state tracking with 1 Hz event publishing
2. **Device Interface Service** - Hardware abstraction layer with vendor-agnostic command interface

## 🏗️ Big Picture

M5 bridges the gap between physical battery hardware and the optimization logic by providing:
- **Hardware Abstraction**: BatteryAdapter interface supporting multiple vendors (Tesla, BYD, etc.)
- **Real-Time Telemetry**: High-frequency state monitoring (1 Hz) for FCAS compliance
- **Event-Driven Commands**: React to charging/discharging decisions from Bidding Service

### Services Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Physical Battery Hardware                 │
│              (Tesla Megapack, BYD Battery-Box, etc.)        │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ Hardware Protocol (Modbus, CAN, etc.)
                       ▼
         ┌─────────────────────────────────┐
         │   BatteryAdapter Interface      │
         │   (Anti-corruption Layer)       │
         ├─────────────────────────────────┤
         │  TeslaLike Mock Implementation  │
         │   BYDLike Mock Implementation   │
         └─────────┬───────────┬───────────┘
                   │           │
        GetState() │           │ SendCommand()
                   │           │
        ┌──────────▼─────┐   ┌▼──────────────────────┐
        │   Telemetry    │   │  Device Interface     │
        │    Service     │   │      Service          │
        ├────────────────┤   ├───────────────────────┤
        │ • REST API     │   │ • Event Subscriber    │
        │ • PostgreSQL   │   │ • Command Handler     │
        │ • 1 Hz Events  │   │ • Stateless           │
        └────────┬───────┘   └────────┬──────────────┘
                 │                    │
                 │ Publishes          │ Publishes
                 │ BatteryState       │ ChargingStarted
                 │ Changed (1 Hz)     │ DischargingStarted
                 │                    │
                 ▼                    ▼
         ┌───────────────────────────────────┐
         │         NATS Event Bus            │
         └───────────────────────────────────┘
                          │
                          │ Subscribes
                          ▼
              ┌─────────────────────┐
              │   Bidding Service   │  ← M6
              │ (Arbitrage Logic)   │
              └─────────────────────┘
```

## 📚 Learning Objectives

### 1. Hardware Abstraction Pattern
- **Adapter Interface**: Define hardware-agnostic API (GetState, SendCommand)
- **Anti-Corruption Layer**: Translate vendor protocols to domain events
- **Swappable Implementations**: TeslaLike vs BYDLike behaviors
- **CustomAttributes**: Handle manufacturer-specific data without domain pollution

### 2. High-Frequency Event Publishing
- **1 Hz Publishing**: BatteryStateChanged every 1 second per battery
- **FCAS Requirement**: Fast frequency response (1-second timing)
- **Best-Effort Pattern**: Don't block on event failures
- **Performance**: Handle 86,400 events/day per battery

### 3. Stateless Service Design
- **Device Interface**: No database, purely event-driven
- **In-Memory State**: Adapter maintains current battery state
- **Resilience**: Restart-safe command handling

### 4. Event-Driven Coordination
- **Command Subscription**: React to ChargingCommandIssued, DischargingCommandIssued
- **Lifecycle Events**: Publish ChargingStarted, ChargingCompleted, etc.
- **State Monitoring**: Continuous telemetry → optimization decisions

## 🏛️ Service Breakdown

### Telemetry Service

**Responsibility**: Monitor and record battery state in real-time

**Components**:
- **Domain**: BatteryState aggregate (SoC, Power, Temperature, Voltage, Current)
- **Repository**: PostgreSQL time-series storage
- **HTTP API**:
  - `GET /api/v1/telemetry/:batteryId/current` - Latest state
  - `GET /api/v1/telemetry/:batteryId/history` - Historical query
- **Event Publisher**: BatteryStateChanged (1 Hz)

**Database**: telemetry-db (PostgreSQL 18 on port 5434)

**Port**: 8082

**Key Pattern**: Time-series optimization with indexed queries

---

### Device Interface Service

**Responsibility**: Abstract hardware and execute commands

**Components**:
- **Domain**: BatteryAdapter interface, Command types
- **Adapters**:
  - TeslaLike - Tesla Megapack simulation
  - BYDLike - BYD Battery-Box simulation
- **Event Subscriber**: ChargingCommandIssued, DischargingCommandIssued
- **Event Publisher**: ChargingStarted, ChargingCompleted, DischargingStarted, DischargingCompleted

**Database**: None (stateless)

**Port**: 8083

**Key Pattern**: Hexagonal architecture with ports & adapters

---

## 🔄 Event Flow

### 1. Battery Registration Flow
```
Asset Management → BatteryRegistered
     ↓
Device Interface → BatteryConnectionEstablished
     ↓
Telemetry → Starts 1 Hz polling
     ↓
BatteryStateChanged (every 1 second)
```

### 2. Charging Command Flow
```
Bidding Service → ChargingCommandIssued
     ↓
Device Interface → Validates + adapter.SendCommand()
     ↓
Device Interface → ChargingStarted
     ↓
Telemetry → BatteryStateChanged (SoC increasing)
     ↓
Device Interface → ChargingCompleted (target reached)
```

### 3. State Monitoring Loop
```
Telemetry → GetState() from adapter (1 Hz)
     ↓
Save to PostgreSQL
     ↓
Publish BatteryStateChanged
     ↓
Bidding Service receives → Make decisions
```

## 🎓 Key Architectural Patterns

### 1. Adapter Pattern (Gang of Four)
```go
// Port (domain interface)
type BatteryAdapter interface {
    GetState(ctx context.Context) (BatteryState, error)
    SendCommand(ctx context.Context, cmd Command) error
}

// Adapter (infrastructure implementation)
type TeslaLikeAdapter struct {
    state BatteryState
    mu    sync.RWMutex
}

func (a *TeslaLikeAdapter) GetState(ctx context.Context) (BatteryState, error) {
    // Translate hardware protocol → domain model
}
```

**Why**: Isolates hardware-specific code from business logic

### 2. Anti-Corruption Layer (DDD)
- Device Interface translates between hardware protocols and domain events
- CustomAttributes field contains vendor-specific data
- Core domain remains vendor-agnostic

### 3. Time-Series Data (Telemetry)
```sql
CREATE INDEX idx_battery_states_battery_id_timestamp
ON battery_states(battery_id, timestamp DESC);
```
**Why**: Optimized for "latest state" and "time-range" queries

### 4. Stateless Service (Device Interface)
- No database needed
- Adapter holds in-memory state
- Restart-safe through event sourcing (future)

## 📊 Performance Considerations

### High-Frequency Publishing (1 Hz)

**Challenge**: 86,400 events/day per battery × N batteries

**Solutions**:
1. **Best-effort publishing**: 2-second timeout, log failures
2. **NATS efficiency**: Lightweight protocol, ~1KB per event
3. **Subscriber filtering**: Only process needed events
4. **Batching (future)**: Aggregate multiple states

**Expected Load** (10 batteries):
- Events/day: 864,000
- Events/second: 10
- Bandwidth: ~10 KB/s

**NATS Capacity**: Handles millions of messages/second

### Database Strategy

**Write Load**: 1 insert/second per battery

**Optimization**:
- Partitioning by battery_id and date (future)
- Archiving old data (>30 days) to cold storage
- Index on (battery_id, timestamp DESC) for latest queries

## 🔧 Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Language | Go 1.23 | Concurrency primitives for 1 Hz goroutines |
| Database | PostgreSQL 18 | Time-series support, JSONB for CustomAttributes |
| Event Bus | NATS 2.10 | Lightweight, high-throughput messaging |
| Testing | Testify | Table-driven tests, mock assertions |
| Migrations | golang-migrate v4 | Versioned schema changes |

## 📈 Success Metrics

### Completion Criteria
- ✅ Mock battery simulation running (TeslaLike and BYDLike)
- ✅ BatteryStateChanged events publishing at 1 Hz
- ✅ 2 adapter types swappable without code changes
- ✅ Telemetry REST API functional
- ✅ Device Interface reacts to commands within 1 second
- ✅ Test coverage >80% overall (>85% domain)
- ✅ End-to-end event flow verified

### Performance Targets
- **Event Publishing**: 1 Hz stable for 5+ minutes
- **Command Response**: <100ms from event received to adapter.SendCommand()
- **API Latency**: <50ms for current state query
- **Memory**: <100MB per service

## 🚀 What's Next

### M6: Bidding Service (Future)
- Subscribe to BatteryStateChanged, MarketPriceUpdated, BatteryRegistered
- Implement arbitrage algorithm (buy low, sell high)
- Publish ChargingCommandIssued, DischargingCommandIssued
- Complete the optimization loop!

### M7: Documentation Polish
- Demo scenario (end-to-end walkthrough)
- Production deployment guide
- Performance tuning recommendations

## 📝 Related Documentation

- [M5 Domain Spec](./M5-DOMAIN-SPEC.md) - Aggregates, validation rules, adapter interface
- [M5 API Spec](./M5-API-SPEC.md) - REST endpoints, event schemas, patterns
- [M5 Checklist](./M5-CHECKLIST.md) - Step-by-step implementation guide
- [EVENTS.md](../../EVENTS.md) - BatteryStateChanged event specification (lines 459-483)
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - TDD workflow and development philosophy

---

**Estimated Time**: 12-16 hours across 8 phases

**Difficulty**: Moderate (replicates M2/M3/M4 patterns + new concepts: adapters, high-frequency events)

**Prerequisites**: M0-M4 complete, NATS running, PostgreSQL 18 × 3 instances
