# Battery Optimization System

A microservices-based battery optimization platform for energy market participation, built with Go and event-driven architecture.

## 🎯 Project Goal

This project demonstrates modern software engineering practices applied to battery energy storage system (BESS) optimization:

- **Domain-Driven Design (DDD)**: Clear bounded contexts and service boundaries
- **Event-Driven Architecture**: Asynchronous communication with eventual consistency
- **Test-Driven Development (TDD)**: Tests written first, driving implementation
- **Microservices**: Independent deployability with DB-per-service pattern
- **Modern Practices**: Continuous delivery mindset, observability, fast feedback

## 🏗️ Architecture Overview

The system consists of several microservices:

1. **Asset Management Service** - Battery specifications and constraints
2. **Market Data Service** - Energy market pricing data
3. **Telemetry Service** - Real-time battery state monitoring
4. **Device Interface Service** - Hardware abstraction layer
5. **Bidding Service** - Real-time bidding decisions based on market conditions

Services communicate via events using NATS, maintaining loose coupling and independent deployability.

## 🔋 Domain Context

This system simulates battery optimization for the Australian National Electricity Market (NEM), focusing on:

- Energy arbitrage (buy low, sell high)
- State of Charge (SoC) management
- Power and ramp rate constraints
- Real-time market participation

## 🚀 Quick Start

**Prerequisites**: Docker, Docker Compose

```bash
# Start all infrastructure (PostgreSQL 18 + NATS)
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

For detailed setup and development commands, see [CLAUDE.md](./CLAUDE.md) and [QUICKSTART.md](./docs/QUICKSTART.md).

## 🤔 Why Event-Driven Architecture?

This project uses event-driven architecture (EDA) over traditional request-response (REST/RPC) patterns for several strategic reasons:

### 1. **Temporal Decoupling**
Services don't need to be online simultaneously. If the Bidding Service is down, Market Data can still publish price updates—Bidding will catch up when it restarts.

**Contrast with REST**: If Market Data tried to POST to Bidding's REST endpoint and Bidding was down, the request would fail and need retry logic.

### 2. **Scalability Through Asynchronous Processing**
High-frequency events (like battery state updates at 1 Hz) can be buffered and processed asynchronously without blocking the publisher.

**Real-world scenario**: Telemetry Service publishes `BatteryStateChanged` every second. Multiple subscribers (Bidding, Telemetry storage, future Alert Service) can process these events at their own pace.

### 3. **Multiple Subscribers Without Coordination**
New services can subscribe to existing events without modifying publishers. This supports the Open-Closed Principle.

**Example**: When we add an Economics Service in the future, it can subscribe to `ChargingOpportunityDetected` events without changing the Bidding Service code.

### 4. **Event Sourcing Readiness**
All state changes are captured as events, making it easy to:
- Rebuild state from event history
- Audit trail for regulatory compliance
- Time-travel debugging ("What was the battery state at 3 PM yesterday?")
- Replay events for testing new algorithms

### 5. **Business Domain Alignment**
The energy market is inherently event-driven:
- AEMO publishes price forecasts (events)
- Batteries transition states (charging started/completed = events)
- Dispatch signals from grid operators (events)

**Modeling the domain as events** makes the system more intuitive and maintainable.

### 6. **Fault Isolation**
If one service has a bug and crashes, it doesn't cascade to other services. Event-driven systems naturally create bulkheads.

**Example**: If Bidding Service has a bug processing price updates, Market Data and Telemetry keep working. When Bidding is fixed, it resumes processing.

### Trade-offs Accepted

Event-driven architecture isn't free:
- **Eventual consistency**: State across services may be briefly out of sync (acceptable for battery optimization)
- **Debugging complexity**: Distributed traces require more tooling (Jaeger/Zipkin in production)
- **Monitoring overhead**: Need to track event lag and dead letters

For this domain (battery optimization), the benefits significantly outweigh the costs.

## 📖 Service Boundaries Explained

Each microservice owns a **bounded context** in the domain:

| Service | Bounded Context | Owns | Publishes Events | Subscribes To |
|---------|----------------|------|------------------|---------------|
| **Asset Management** | Battery specifications | Battery aggregate (capacity, power limits, constraints) | `battery.registered.v1` | None (source of truth) |
| **Market Data** | Energy pricing | MarketPrice aggregate (time-series pricing data) | `market.price.updated.v1` | None (integrates with AEMO) |
| **Telemetry** | Real-time monitoring | BatteryState (SoC, power, temperature) | `battery.state.changed.v1` (1 Hz) | `battery.connection.established.v1` |
| **Device Interface** | Hardware control | BatteryAdapter (vendor abstraction layer) | `charging.started.v1`, `discharging.started.v1`, `charging.completed.v1`, `discharging.completed.v1` | `charging.command.issued.v1`, `discharging.command.issued.v1` |
| **Bidding** | Trading decisions | BiddingDecision (arbitrage algorithm) | `charging.opportunity.detected.v1`, `discharging.opportunity.detected.v1`, `charging.command.issued.v1`, `discharging.command.issued.v1` | `battery.state.changed.v1`, `market.price.updated.v1`, `battery.registered.v1` |

**Key Principle**: Services communicate **only through events** (no direct database access, no synchronous RPC).

### Why DB-per-Service?

Each service has its own PostgreSQL database:
- **Asset Management**: `asset-db` on port 5432
- **Market Data**: `market-db` on port 5433
- **Telemetry**: `telemetry-db` on port 5434
- **Bidding**: No database (stateless, in-memory caches rebuilt from events)
- **Device Interface**: No database (runtime state only)

**Benefits**:
1. **Independent schema evolution**: Market Data can change its schema without coordinating with Asset Management
2. **Technology flexibility**: Could use TimescaleDB for Telemetry without affecting other services
3. **Deployment independence**: Can restart/upgrade services independently
4. **Data ownership**: Clear boundaries prevent accidental coupling

**Trade-off**: Need event-driven synchronization instead of database joins (eventual consistency).

## 🎮 How to Run the System

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (for local development)
- curl (for testing REST APIs)

### Full System Startup

```bash
# 1. Start infrastructure (NATS + 3 PostgreSQL databases)
docker-compose up -d nats asset-db market-db telemetry-db

# 2. Verify infrastructure health
curl http://localhost:8222/healthz  # NATS health check
docker-compose ps                    # All should be "healthy"

# 3. Start all services
docker-compose up -d asset-management market-data telemetry device-interface bidding

# 4. View aggregated logs
docker-compose logs -f

# 5. View specific service logs
docker-compose logs -f bidding
```

### Running Services Locally (Development)

```bash
# Terminal 1: Asset Management (port 8080)
cd services/asset-management
go run cmd/server/main.go

# Terminal 2: Market Data (port 8081)
cd services/market-data
go run cmd/server/main.go

# Terminal 3: Telemetry (port 8082)
cd services/telemetry
go run cmd/server/main.go

# Terminal 4: Device Interface (port 8083)
cd services/device-interface
BATTERY_ID=battery-123 ADAPTER_TYPE=TeslaLike CAPACITY=200 MAX_POWER=100 go run cmd/server/main.go

# Terminal 5: Bidding (event-driven, no port)
cd services/bidding
AUTOMATION_MODE=FULL_AUTO go run cmd/server/main.go

# Terminal 6: Event Monitor (see all events)
cd tools/event-subscriber
go run main.go ">"  # Subscribe to all events
```

### Testing the System End-to-End

#### Manual Testing with DEMO.md

See [DEMO.md](./docs/DEMO.md) for a comprehensive step-by-step demo scenario with expected events and troubleshooting.

#### Automated End-to-End Tests

The system includes 13 comprehensive E2E tests that validate complete workflows across all 5 microservices:

**Prerequisites**: All infrastructure services must be running (NATS + 3 PostgreSQL databases)

```bash
# Run all E2E tests
cd tests/e2e
go test -v

# Run specific test categories
go test -v -run TestCompleteArbitrageWorkflow  # Full charging/discharging cycle
go test -v -run TestDayInTheLife               # Multi-battery scenario
go test -v -run TestChargingWorkflow           # Charging opportunity detection
```

**Test Categories**:

1. **Happy Path Tests** (4 tests):
   - `TestChargingWorkflow` - Low price triggers charging opportunity
   - `TestDischargingWorkflow` - High price triggers discharging opportunity
   - `TestMultipleBatteriesWorkflow` - Multiple batteries operate independently
   - `TestPriceChangeTriggersDecision` - Price updates trigger new decisions

2. **Edge Cases** (4 tests):
   - `TestEdgeCaseChargingThreshold` - Price exactly at $50/MWh threshold
   - `TestEdgeCaseDischargingThreshold` - Price exactly at $100/MWh threshold
   - `TestEdgeCaseSoCBoundaries` - SoC at 0%, 50%, 100%
   - `TestNoOpportunityMidRange` - Mid-range price ($75/MWh) triggers no action

3. **Error Handling** (3 tests):
   - `TestInvalidBatteryState` - Invalid SoC values handled gracefully
   - `TestServiceUnavailable` - Services handle NATS disconnection
   - `TestEventPublishFailure` - Best-effort publishing doesn't block operations

4. **Complete Workflows** (2 tests):
   - `TestCompleteArbitrageWorkflow` - Full cycle: charge low → discharge high → revenue
   - `TestDayInTheLife` - 24-hour simulation with multiple batteries and price changes

**What E2E Tests Validate**:
- ✅ Event-driven communication across all services
- ✅ NATS pub/sub reliability
- ✅ Business logic correctness (charging/discharging decisions)
- ✅ Service integration (Asset Management → Telemetry → Bidding → Device Interface)
- ✅ Error handling and graceful degradation
- ✅ Multi-battery scenarios
- ✅ Real-world arbitrage workflows

**E2E Test Infrastructure**:
- **Helpers** (`helpers.go`): REST API calls, NATS publishing, event waiting
- **Assertions** (`assertions.go`): Event validation with structured checks
- **Fixtures** (`fixtures.go`): Test data (price scenarios, SoC levels, battery configs)

For detailed E2E test documentation, see [tests/e2e/README.md](./tests/e2e/README.md).

### Cleanup

```bash
# Stop services (preserve data)
docker-compose down

# Stop and remove ALL data (fresh start)
docker-compose down -v
```

## 📚 Documentation

**Core Documentation**:
- **[PLANNING.md](./PLANNING.md)** - Milestone tracking and project status (M0-M6 ✅)
- **[EVENTS.md](./EVENTS.md)** - Complete event catalog (25 events with schemas and flow diagrams)
- **[ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - System architecture with mermaid diagrams
- **[STRUCTURE.md](./docs/STRUCTURE.md)** - Go project structure and development workflow
- **[QUICKSTART.md](./docs/QUICKSTART.md)** - Quick start guide for development setup
- **[CLAUDE.md](./CLAUDE.md)** - Guidance for Claude Code (development commands, patterns)
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - **Development philosophy and TDD workflow**

**Development Guides**:
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Development philosophy and TDD workflow (Kent Beck style) - **READ FIRST**
- **[Domain Guide](./.claude/skills/BATTERY-DOMAIN-SKILL.md)** - Battery domain concepts and validation rules

**Milestone Documentation**:
- **M2 (Asset Management Service)**:
  - [M2 Overview](./docs/milestones/M2-OVERVIEW.md) - Big picture and learning objectives
  - [M2 Domain Spec](./docs/milestones/M2-DOMAIN-SPEC.md) - Battery aggregate and validation rules
  - [M2 API Spec](./docs/milestones/M2-API-SPEC.md) - REST endpoints and DTOs
  - [M2 Checklist](./docs/milestones/M2-CHECKLIST.md) - Step-by-step implementation guide
- **M3 (Market Data Service)**:
  - [M3 Overview](./docs/milestones/M3-OVERVIEW.md) - Time-series data and DB-per-service pattern
  - [M3 Domain Spec](./docs/milestones/M3-DOMAIN-SPEC.md) - MarketPrice aggregate
  - [M3 API Spec](./docs/milestones/M3-API-SPEC.md) - Time-range queries
  - [M3 Checklist](./docs/milestones/M3-CHECKLIST.md) - Implementation guide
- **M4 (Event Bus Integration)**:
  - [M4 Overview](./docs/milestones/M4-OVERVIEW.md) - Event-driven architecture
  - [M4 Domain Spec](./docs/milestones/M4-DOMAIN-SPEC.md) - Event schemas and versioning
  - [M4 API Spec](./docs/milestones/M4-API-SPEC.md) - NATS pub/sub patterns
  - [M4 Checklist](./docs/milestones/M4-CHECKLIST.md) - Event integration guide
  - [pkg/events README](./pkg/events/README.md) - Event library usage guide
- **M5 (Telemetry + Device Interface)**:
  - [M5 Overview](./docs/milestones/M5-OVERVIEW.md) - Hardware abstraction and real-time monitoring
  - [M5 Domain Spec](./docs/milestones/M5-DOMAIN-SPEC.md) - BatteryAdapter interface and BatteryState aggregate
  - [M5 API Spec](./docs/milestones/M5-API-SPEC.md) - Telemetry endpoints and Device Interface events
  - [M5 Checklist](./docs/milestones/M5-CHECKLIST.md) - Implementation guide
  - [Telemetry Service README](./services/telemetry/README.md) - Complete service documentation
  - [Device Interface README](./services/device-interface/README.md) - Hardware adapter documentation
- **M6 (Bidding Service)**:
  - [M6 Overview](./docs/milestones/M6-OVERVIEW.md) - Arbitrage algorithm and business logic
  - [M6 Domain Spec](./docs/milestones/M6-DOMAIN-SPEC.md) - Bidding decision algorithm
  - [M6 API Spec](./docs/milestones/M6-API-SPEC.md) - Event subscriptions and publications
  - [M6 Checklist](./docs/milestones/M6-CHECKLIST.md) - Implementation guide
  - [Bidding Service README](./services/bidding/README.md) - Complete service documentation

## 🎓 Learning Focus

This project prioritizes:
1. High-level architectural patterns over language-specific syntax
2. Understanding trade-offs in distributed systems
3. Modern engineering practices (TDD, CD, observability)
4. Domain-driven design in a real-world context

## 🏢 Inspiration

Inspired by battery optimization platforms in the Australian energy market, such as Hachiko and Evergen.

## 📝 Current Status

**✅ Completed Milestones**:

- **M0: Project Setup** - Infrastructure running (Docker Compose, PostgreSQL 18, NATS)
- **M1: Event Storming** - 25 events documented with schemas and flow diagrams

- **M2: Asset Management Service** - Production-ready REST API with TDD and Hexagonal Architecture
  - 79.8% test coverage (domain: 96.7%)
  - 3 RESTful endpoints (POST, GET, LIST)
  - PostgreSQL persistence with automatic migrations
  - Docker deployment ready

- **M3: Market Data Service** - Time-series data service with DB-per-service pattern
  - 84.6% test coverage (domain: 96.7%, repository: 91.1%)
  - Time-range queries with filtering (region, interval_type)
  - Pagination support (limit, offset)
  - Separate PostgreSQL instance (market-db on port 5433)
  - AEMO market data concepts (5MIN/30MIN intervals)

- **M4: Event Bus Integration** - Event-driven architecture with NATS
  - pkg/events library with 88.9% test coverage (27/27 tests passing)
  - NATS publisher/subscriber adapters
  - 2 services publishing events (battery.registered.v1, market.price.updated.v1)
  - Test subscriber tool for real-time monitoring
  - Event versioning strategy (subject + payload)
  - Best-effort publishing pattern (graceful degradation)

- **M5: Telemetry + Device Interface** - Hardware abstraction and real-time monitoring
  - **Telemetry Service**: 86.1% test coverage (domain: 100%, postgres: 81.5%, http: 76.7%)
    - PostgreSQL time-series storage with optimized indexes
    - REST API: GET current state, GET history with time-range queries
    - BatteryState domain model with comprehensive validation
    - JSONB support for CustomAttributes
  - **Device Interface Service**: 77.5% test coverage (adapters: 78.5%, service: 76.4%)
    - BatteryAdapter interface for vendor-agnostic hardware control
    - TeslaLike mock adapter (5 MW/s ramp rate, 95% efficiency, temperature physics)
    - Real-time simulation with 100ms tick rate
    - Event-driven command handling (charging/discharging via NATS)
    - Lifecycle event publishing (ChargingStarted, DischargingStarted)
  - 7 new events added to pkg/events library
  - End-to-end event flow validated

- **M6: Bidding Service** - Real-time arbitrage decision engine
  - **94% test coverage** overall (domain: 96.9%, cache: 100%, service: 88.2%)
  - Simple arbitrage algorithm (charge < $50/MWh, discharge > $100/MWh)
  - In-memory state management with thread-safe caches (sync.RWMutex)
  - Event-driven architecture (subscribes to 3 event types, publishes 4)
  - Automation modes: MANUAL (detect only) and FULL_AUTO (detect + execute)
  - Stateless service (no database, rebuilt from events)
  - 4 new events: ChargingOpportunityDetected, DischargingOpportunityDetected, ChargingCommandIssued, DischargingCommandIssued
  - Integration testing validated (end-to-end event flow)

**🎉 MVP Complete!** All core services implemented (M0-M6)

**Services Running**:
- Asset Management: REST API (8080) + Event Publishing
- Market Data: REST API (8081) + Event Publishing
- Telemetry: REST API (8082) + Time-series storage
- Device Interface: Event-driven command handler (port 8083)
- Bidding: Event-driven arbitrage engine (stateless)
- Infrastructure: NATS (4222), PostgreSQL x3 (5432, 5433, 5434)

See [PLANNING.md](./PLANNING.md) for detailed milestone tracking and next steps.

## 🔗 Tech Stack

- **Language**: Go 1.23+
- **Event Bus**: NATS 2.10
- **Database**: PostgreSQL 18 (DB-per-service: 3 independent instances)
- **Containers**: Docker & Docker Compose
- **Testing**: Go testing package with table-driven tests (TDD approach)

## 🎯 Business Value Alignment

### 1. Solid Optimization Fundamentals (Current Focus)
**Foundation First Approach**: Before bankability or fast deployment can deliver value, the optimization engine must work correctly.

This project builds:
- Accurate battery state management and constraints
- Real-time market data integration
- Optimal charging/discharging decisions
- FCAS (Frequency Control Ancillary Services) support
- Event-driven architecture for scalability

**Why this matters**: Predictable revenue (bankability) is only possible with consistent, reliable optimization.

---

### 2. Future Enhancement: Bankability Layer

Once optimization fundamentals are solid, the next phase would add financial predictability features:

#### Revenue Predictability
- Historical revenue tracking per operation
- Multi-scenario forecasting (conservative, expected, optimistic)
- Confidence intervals for investor reporting
- Revenue volatility analysis

#### Investment Metrics
- ROI, IRR, Payback period calculations
- Risk-adjusted returns
- Expected vs. actual performance tracking

#### Lender Dashboard
- Standardized financial reporting for banks/investors
- Compliance documentation
- Real-time performance vs. forecast
- "Bankability score" calculation

**Business Impact**: Makes C&I batteries investable by providing lenders with predictable revenue projections backed by proven optimization performance.

---

### 3. Future Enhancement: Fast Deployment via AI Simulation

#### The Pre-Sales Problem
Traditional battery project sales require:
- Manual site modeling (weeks)
- Custom revenue projections
- Multiple sales meetings
- High touch sales process

This limits deployment velocity.

#### The Solution: Public AI-Driven Simulation Tool
A self-service tool that transforms the sales pipeline:

**User Flow:**
1. **Upload historical data** (CSV from any battery/solar monitor)
2. **AI normalizes data** (handles different formats, fills gaps)
3. **Run optimization** (same engine as live operations)
4. **See results instantly**: "You could have earned $45,000 more last year"

**Architecture Advantage:**
This project's **core optimization engine** is designed to be data-source agnostic:
```
┌─────────────────────────────┐
│ Simulation Interface        │ ← Future: CSV upload
├─────────────────────────────┤
│ Live Operation Interface    │ ← Current: Real-time APIs
├─────────────────────────────┤
│ Optimization Engine (CORE)  │ ← This project
│ • Market analysis           │
│ • Constraint validation     │
│ • Revenue calculation       │
│ • FCAS integration          │
└─────────────────────────────┘
```

**Dual Usage:**
- **Simulation Mode**: Historical data → "What if" analysis (pre-sales)
- **Live Mode**: Real-time telemetry → Actual optimization (production)

**Business Impact:**
- **Lead Magnet**: Prospects engage immediately with value
- **Sales Efficiency**: Automate pre-sales modeling
- **Trust Through Transparency**: Show actual historical performance
- **Scalability**: Small team handles thousands of sites

#### AI Translation Layer (Future)
- **Data Preprocessing**: Normalize CSV columns from different OEMs
- **Gap Filling**: Use historical weather/market data to complete datasets
- **Quality Detection**: Flag poor quality data before simulation
- **Privacy**: De-identification and secure upload protocols

---

## 🏗️ Design Philosophy

**Build the Foundation Right:**
This project focuses on creating a robust optimization engine that:
1. **Works correctly** (accurate decisions)
2. **Scales efficiently** (handles multiple sites)
3. **Remains flexible** (adapts to different use cases)

Once this foundation is solid:
- **Bankability features** add financial predictability on top
- **Simulation tools** reuse the same optimization logic for pre-sales
- **Multi-OEM support** (via hardware abstraction layer) enables universal deployment

**Key Insight**: The optimization engine is the "brain" - whether analyzing historical data (simulation) or controlling live batteries (production), the core logic remains the same. Build it once, use it everywhere.
