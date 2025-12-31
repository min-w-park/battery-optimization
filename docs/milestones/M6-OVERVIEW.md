# M6: Bidding Service - Overview

## Big Picture

M6 implements the **Bidding Service** - the first service with real business logic. This service acts as the brain of the battery optimization system, making intelligent arbitrage decisions by analyzing market prices and battery states in real-time.

Unlike previous services (Asset Management, Market Data, Telemetry, Device Interface) which are primarily data management services, the Bidding Service contains the core optimization algorithm that drives profitability.

## What You'll Build

A **stateless, event-driven decision engine** that:

1. **Listens** to real-time battery state updates (1 Hz from Telemetry Service)
2. **Monitors** market price changes (from Market Data Service)
3. **Detects** arbitrage opportunities using a simple algorithm
4. **Publishes** opportunity events for visibility
5. **Issues** charging/discharging commands (in FULL_AUTO mode)

### Example Flow

```
[Market Data Service] ──> MarketPriceUpdated ($30/MWh) ──┐
                                                           │
[Telemetry Service] ──> BatteryStateChanged (SoC=50%) ────┤
                                                           │
                                                           v
                                                [Bidding Service]
                                                   Arbitrage Algorithm
                                                   Price < $50 AND SoC < 80%
                                                           │
                                     ┌─────────────────────┴─────────────────────┐
                                     v                                           v
                     ChargingOpportunityDetected              ChargingCommandIssued
                          (visibility)                          (if FULL_AUTO)
                                     │                                           │
                                     v                                           v
                               [Human UI]                          [Device Interface Service]
                            (future service)                     Executes charge command
```

## Learning Objectives

By completing M6, you'll learn:

### 1. **Event-Driven State Management**
   - Building stateless services that derive state from events
   - Eventual consistency patterns
   - Handling high-frequency events (1 Hz)

### 2. **In-Memory Caching**
   - Thread-safe cache design with `sync.RWMutex`
   - Concurrent read/write patterns
   - Cache invalidation strategies

### 3. **Business Logic Layer**
   - Domain-driven design for decision algorithms
   - Separating decision logic from execution
   - Testable business rules

### 4. **Event Composition**
   - Subscribing to multiple event streams
   - Publishing multiple event types
   - Event versioning and schema design

## Service Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Bidding Service                          │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    NATS Subscriber                      │   │
│  │  - battery.state.changed.v1 (1 Hz)                     │   │
│  │  - market.price.updated.v1                             │   │
│  │  - battery.registered.v1                               │   │
│  └──────────────────────┬──────────────────────────────────┘   │
│                         │                                       │
│                         v                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │               Bidding Engine (Service)                  │   │
│  │                                                         │   │
│  │  HandleBatteryStateChanged() ─┐                        │   │
│  │  HandleMarketPriceUpdated() ───┼─> Arbitrage Algorithm │   │
│  │  HandleBatteryRegistered() ────┘    (Domain Logic)     │   │
│  │                                                         │   │
│  └──────────┬────────────────────────────────────┬─────────┘   │
│             │                                    │             │
│             v                                    v             │
│  ┌─────────────────────┐            ┌─────────────────────┐   │
│  │ BatteryStateCache   │            │    PriceCache       │   │
│  │  (sync.RWMutex)     │            │  (sync.RWMutex)     │   │
│  │                     │            │                     │   │
│  │  batteryID -> State │            │  latestPrice: 30.0  │   │
│  │  battery-123 -> ... │            │  latestTime: ...    │   │
│  └─────────────────────┘            └─────────────────────┘   │
│                         │                                       │
│                         v                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    NATS Publisher                       │   │
│  │  - charging.opportunity.detected.v1                    │   │
│  │  - discharging.opportunity.detected.v1                 │   │
│  │  - charging.command.issued.v1 (if FULL_AUTO)           │   │
│  │  - discharging.command.issued.v1 (if FULL_AUTO)        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  No Database - Stateless Service                               │
└─────────────────────────────────────────────────────────────────┘
```

## Arbitrage Algorithm

### Simple Threshold-Based Strategy

```
CHARGE_THRESHOLD = $50/MWh
DISCHARGE_THRESHOLD = $100/MWh

Charging Opportunity:
  IF price < $50/MWh
  AND SoC < 80%
  AND battery is IDLE
  THEN publish ChargingOpportunityDetected

Discharging Opportunity:
  IF price > $100/MWh
  AND SoC > 30%
  AND battery is IDLE
  THEN publish DischargingOpportunityDetected
```

### Why These Thresholds?

Based on Australian NEM market analysis:
- **$50/MWh**: Below-average price, good for charging
- **$100/MWh**: Above-average price, profitable for discharging
- **SoC 30-80%**: Operating range to preserve battery health
- **IDLE state**: Avoid conflicts with existing FCAS contracts

### Profit Calculation (Simplified)

```
Expected Profit = Price Delta × Capacity × Efficiency

Example:
- Charge at $30/MWh (delta = $20 below threshold)
- Capacity = 200 MWh
- Efficiency = 0.95 (round-trip)
- Profit = $20 × 200 × 0.95 = $3,800 per cycle
```

## Automation Modes

The Bidding Service supports three levels of automation:

### 1. MANUAL Mode
- **Detects** opportunities
- **Publishes** opportunity events
- **Does NOT** issue commands
- Human reviews and approves each trade

**Use Case**: Initial deployment, regulatory compliance, learning phase

### 2. SEMI_AUTO Mode (Not in M6)
- Detects opportunities
- Publishes opportunity events with suggestions
- Waits for human approval before issuing commands
- Future enhancement for M7+

### 3. FULL_AUTO Mode
- Detects opportunities
- Publishes opportunity events (for audit trail)
- **Automatically issues** charging/discharging commands
- No human intervention required

**Use Case**: Production operation, high-frequency trading

## Event Flow Diagram

```
┌──────────────────┐
│ Asset Management │ ──> BatteryRegistered
└──────────────────┘         │
                             v
                    ┌─────────────────┐
                    │ Bidding Service │
                    │ (Initialize)    │
                    └─────────────────┘
                             │
┌──────────────────┐         │
│  Market Data     │ ──> MarketPriceUpdated ($30/MWh)
└──────────────────┘         │
                             v
                    ┌─────────────────┐
                    │ Bidding Service │ ──> ChargingOpportunityDetected
                    │ (Detect Low $)  │
                    └─────────────────┘      │
                             │                │
┌──────────────────┐         │                v
│   Telemetry      │ ──> BatteryStateChanged (SoC=50%, IDLE)
└──────────────────┘         │
                             v
                    ┌─────────────────┐
                    │ Bidding Service │ ──> ChargingCommandIssued
                    │ (Issue Command) │     (if FULL_AUTO)
                    └─────────────────┘      │
                                             v
                                    ┌──────────────────┐
                                    │ Device Interface │
                                    │ (Execute Charge) │
                                    └──────────────────┘
```

## Key Design Decisions

### 1. Stateless Architecture (No Database)
**Why**:
- Simplifies deployment (no migrations)
- State rebuilt from events on restart
- Aligns with event-driven architecture

**Trade-off**:
- No historical decision tracking (can add later)
- Must wait for events to populate caches on startup

### 2. In-Memory Caches
**Why**:
- Low-latency decision making (sub-millisecond)
- Handles 1 Hz battery state updates efficiently
- Simple concurrency model

**Trade-off**:
- State lost on restart (acceptable for stateless design)
- Memory usage grows with battery count (manageable for 100s of batteries)

### 3. Event-First Publishing
**Why**:
- Visibility into all decisions (audit trail)
- Opportunity events useful for UI, analytics, debugging
- Separates detection from execution

**Trade-off**:
- More events to manage (4 new event types)
- Slight overhead vs direct command execution

### 4. Simple Threshold Algorithm
**Why**:
- Transparent and debuggable
- Regulatory compliance (explainable decisions)
- Foundation for ML-based algorithms later

**Trade-off**:
- Suboptimal vs sophisticated ML models
- Fixed thresholds don't adapt to market conditions

## Integration with Existing Services

### Dependencies (Subscribes To):
1. **Telemetry Service** → `BatteryStateChanged` (1 Hz)
   - Provides: SoC, Power, OperationState, Temperature

2. **Market Data Service** → `MarketPriceUpdated`
   - Provides: Price, Timestamp, Interval

3. **Asset Management Service** → `BatteryRegistered`
   - Provides: BatteryID, Capacity, MaxPower, Efficiency

### Consumers (Publishes For):
1. **Device Interface Service** → Consumes command events
   - `ChargingCommandIssued` → Start charging
   - `DischargingCommandIssued` → Start discharging

2. **Future UI Service** → Consumes opportunity events
   - `ChargingOpportunityDetected` → Display to operator
   - `DischargingOpportunityDetected` → Display to operator

## Success Criteria

After completing M6, you should be able to:

- ✅ Register a battery via Asset Management API
- ✅ Create a low price ($30/MWh) via Market Data API
- ✅ See `ChargingOpportunityDetected` event published
- ✅ See `ChargingCommandIssued` event (if FULL_AUTO)
- ✅ Create a high price ($150/MWh) via Market Data API
- ✅ See `DischargingOpportunityDetected` event published
- ✅ See `DischargingCommandIssued` event (if FULL_AUTO)
- ✅ Verify caches update correctly with concurrent events
- ✅ All tests pass with >75% coverage

## Next Steps

After M6 completion:
- **M7: Final Polish** - Documentation, demo scenario, production readiness
- System will have **5 microservices** communicating via **12+ event types**
- Full arbitrage flow working end-to-end

## Files You'll Create

```
services/bidding/
├── cmd/server/main.go                                  # Main entry point
├── internal/
│   ├── domain/
│   │   ├── arbitrage.go                               # Business logic
│   │   ├── arbitrage_test.go
│   │   ├── bidding_decision.go                        # Aggregate
│   │   ├── bidding_decision_test.go
│   │   └── errors.go                                  # Domain errors
│   ├── cache/
│   │   ├── battery_state_cache.go                     # Thread-safe cache
│   │   ├── battery_state_cache_test.go
│   │   ├── price_cache.go
│   │   └── price_cache_test.go
│   └── service/
│       ├── bidding_engine.go                          # Event handlers
│       └── bidding_engine_test.go
├── go.mod
├── Dockerfile
└── README.md

pkg/events/
├── charging_opportunity_detected.go                    # New events
├── charging_opportunity_detected_test.go
├── discharging_opportunity_detected.go
├── discharging_opportunity_detected_test.go
├── charging_command_issued.go
├── charging_command_issued_test.go
├── discharging_command_issued.go
└── discharging_command_issued_test.go

docs/milestones/
├── M6-OVERVIEW.md                                      # This file
├── M6-DOMAIN-SPEC.md                                   # Domain model details
├── M6-API-SPEC.md                                      # Event schemas
└── M6-CHECKLIST.md                                     # Implementation steps
```

## Estimated Time

- **Phase 0**: Documentation (1.5h) ← Current phase
- **Phase 1**: Domain Layer (2-3h)
- **Phase 2**: Event Library (1-2h)
- **Phase 3**: Caches (2-3h)
- **Phase 4**: Bidding Engine (2-3h)
- **Phase 5**: Main App (2-3h)
- **Phase 6**: Integration Testing (2-3h)
- **Phase 7**: Polish (1-2h)

**Total**: 10-14 hours (aggressive) to 14-21 hours (conservative)

---

**Ready to build the decision-making brain of the battery optimization system!** 🧠⚡
