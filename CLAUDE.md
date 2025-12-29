# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Battery optimization system for energy market participation using microservices architecture and event-driven patterns. Built with Go, targeting the Australian National Electricity Market (NEM). The system simulates battery energy storage system (BESS) optimization through energy arbitrage (buy low, sell high) while managing State of Charge (SoC), power constraints, and ramp rates.

## Architecture

### Microservices Design

The system follows Domain-Driven Design (DDD) with bounded contexts:

1. **Asset Management Service** - Battery specifications, capacity, power limits, ramp rate constraints
2. **Market Data Service** - Energy market pricing data (AEMO price simulation)
3. **Telemetry Service** - Real-time battery state monitoring (SoC tracking)
4. **Device Interface Service** - Hardware abstraction layer via BatteryAdapter interface
5. **Bidding Service** - Real-time bidding decisions based on market conditions

### Event-Driven Communication

Services communicate asynchronously via NATS event bus, maintaining loose coupling and independent deployability. Key architectural patterns:

- **DB-per-service**: Each microservice has its own PostgreSQL database
- **Eventual consistency**: Services react to events published by other services
- **Event versioning**: Field additions maintain backward compatibility
- **Publisher/Subscriber pattern**: Services publish domain events and subscribe to relevant events

### Infrastructure Setup

Three PostgreSQL 18 databases run in separate containers:
- `asset-db` on port 5432 (DB: asset_management, user: asset_user)
- `market-db` on port 5433 (DB: market_data, user: market_user)
- `telemetry-db` on port 5434 (DB: telemetry, user: telemetry_user)

NATS event bus:
- Client connections: port 4222
- HTTP monitoring: port 8222 (health check: http://localhost:8222/healthz)

All infrastructure services are defined in docker-compose.yml and can be started with `docker-compose up -d`.

## Development Commands

### Infrastructure Management

```bash
# Start all infrastructure services (databases + NATS)
docker-compose up -d

# Check service health status
docker-compose ps

# View logs for all services
docker-compose logs -f

# View logs for specific service
docker-compose logs -f asset-db

# Stop services (preserve data)
docker-compose down

# Stop and remove all data volumes
docker-compose down -v
```

### Database Operations

```bash
# Connect to Asset Management DB (PostgreSQL 18)
docker exec -it asset-db psql -U asset_user -d asset_management

# Connect to Market Data DB
docker exec -it market-db psql -U market_user -d market_data

# Connect to Telemetry DB
docker exec -it telemetry-db psql -U telemetry_user -d telemetry

# Check database version (should show PostgreSQL 18.x)
docker exec asset-db psql -U asset_user -d asset_management -c "SELECT version();"
```

**Note**: Remove the `-it` flag when running commands in scripts or CI/CD pipelines to avoid TTY errors.

### NATS Monitoring

```bash
# Health check
curl http://localhost:8222/healthz

# View NATS server variables
curl http://localhost:8222/varz

# View connection info
curl http://localhost:8222/connz
```

## Development Approach

### Test-Driven Development (TDD)

Tests are written first, driving implementation. Use table-driven tests following Go conventions:

```go
func TestBatteryValidation(t *testing.T) {
    tests := []struct {
        name    string
        battery Battery
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Domain-Driven Design

Each service owns its domain model as aggregates with clear boundaries:
- Battery Aggregate: ID, Capacity, MaxPower, RampRate with constraint validation
- MarketPrice: Timestamp, Value, Interval for time-series data
- BatteryState: SoC, Power, Status for real-time telemetry

### Event Catalog

**25 domain events** are fully documented in [EVENTS.md](EVENTS.md). Key event categories:

**Initialization (5 events)**:
- `BatteryRegistered` - Asset Management publishes when battery added
- `BatteryConnectionEstablished` - Device Interface confirms hardware connection
- `BatteryTestStarted/Completed` - Health check test lifecycle
- `BatteryReadyToOperate` - Battery approved for market participation

**Market Data (3 events)**:
- `AemoPriceForecastReceived` - AEMO 5-min/30-min predispatch forecasts
- `InternalPriceForecastGenerated` - ML-based price predictions
- `MarketPriceChangedSignificantly` - Conditional event for major price movements

**Charging Operations (7 events)**:
- `ChargingOpportunityDetected` - Bidding Service identifies profitable charging
- `ChargingCommandIssued` - Command to start charging (manual/semi-auto/full-auto)
- `ConflictDetected` - Charging conflicts with existing FCAS contract
- `EconomicsCalculationRequested/Calculated` - ROI analysis for conflict resolution
- `ConflictResolutionSuggested/Resolved` - Economics-driven decision
- `ChargingStarted/Completed` - Charging lifecycle

**Discharging Operations (4 events)**:
- `DischargingOpportunityDetected` - High price + sufficient SoC validation
- `DischargingCommandIssued` - Command with multiple stop conditions
- `DischargingStarted/Completed` - Discharge lifecycle with reason codes

**FCAS (Frequency Control) (4 events)**:
- `FcasContractStarted/Ended` - Contract lifecycle (planning level)
- `FcasDispatchReceived` - AEMO dispatch signal (execution level)
- `FcasDispatchCompleted` - Compliance status tracking

**Shared Events**:
- `BatteryStateChanged` - Published every 1 second by Telemetry Service

See [EVENTS.md](EVENTS.md) for complete event schemas, flow diagrams, and service responsibilities.

## Documentation

- **[STRUCTURE.md](docs/STRUCTURE.md)**: Go project structure, service layout, and development workflow
- **[EVENTS.md](EVENTS.md)**: Complete event catalog (25 events) with JSON schemas and flow diagrams
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)**: System architecture with mermaid diagrams, service boundaries, and design patterns
- **[PLANNING.md](PLANNING.md)**: Milestone tracking and project status
- **[QUICKSTART.md](docs/QUICKSTART.md)**: Quick start guide for development setup

**M2 Milestone Documentation**:
- **[M2 Overview](docs/milestones/M2-OVERVIEW.md)**: Big picture and learning objectives
- **[M2 Domain Spec](docs/milestones/M2-DOMAIN-SPEC.md)**: Battery aggregate and validation rules
- **[M2 API Spec](docs/milestones/M2-API-SPEC.md)**: REST endpoints and DTOs
- **[M2 Checklist](docs/milestones/M2-CHECKLIST.md)**: Step-by-step implementation guide

## Service Development Pattern

When implementing a new service:

1. **Domain Model First**: Define aggregates, entities, value objects (see [STRUCTURE.md](docs/STRUCTURE.md))
2. **Repository Pattern**: Abstract data access with interfaces
3. **REST API**: Expose service capabilities (CRUD operations)
4. **Event Emission**: Publish domain events at aggregate boundaries (see [EVENTS.md](EVENTS.md))
5. **Event Subscription**: React to relevant events from other services
6. **TDD**: Write tests first using table-driven test pattern
7. **Dockerfile**: Containerize the service
8. **docker-compose.yml**: Integrate into infrastructure

Refer to [STRUCTURE.md](docs/STRUCTURE.md) for detailed directory layout and service-specific patterns.

For M2 (Asset Management Service), follow the detailed guide in [M2 Checklist](docs/milestones/M2-CHECKLIST.md).

## Hardware Abstraction

BatteryAdapter interface enables multiple battery vendor implementations:

```go
type BatteryAdapter interface {
    GetState(ctx context.Context) (BatteryState, error)
    SendCommand(ctx context.Context, cmd Command) error
}
```

Mock adapters simulate different vendor behaviors (TeslaLike, BYDLike) with CustomAttributes support.

## Project Status

**✅ Completed Milestones**:
- **M0: Project Setup** - Infrastructure running (Docker Compose, PostgreSQL 18, NATS)
- **M1: Event Storming** - 25 events documented with schemas and flow diagrams

**🚧 Current**: M2 - Asset Management Service

See [PLANNING.md](PLANNING.md) for detailed milestone tracking and task breakdowns.

**Upcoming Milestones**:
- M2: Asset Management Service (Battery domain model, REST API, TDD)
- M3: Market Data Service (DB-per-service pattern)
- M4: Event Bus Integration (NATS pub/sub)
- M5: Telemetry + Device Interface (Hardware abstraction)
- M6: Bidding Service (Arbitrage algorithm)
- M7: Documentation polish

## Key Design Decisions

- **Go as primary language**: Strong concurrency primitives, explicit error handling
- **PostgreSQL 18**: Latest stable version with improved performance
- **NATS over Kafka**: Lightweight, simpler ops for learning context
- **DB-per-service**: True service isolation, independent schemas (3 separate PostgreSQL instances)
- **Event-driven over RPC**: Temporal decoupling, better fault isolation
- **Mock adapters**: Hardware simulation without physical dependencies
- **Australian NEM context**: Real-world market constraints and pricing patterns
- **Standard Go Project Layout**: `cmd/`, `internal/`, `pkg/` structure for maintainability

## Key Event-Driven Insights

From the Event Storming session (M1), several critical architectural insights emerged:

1. **Charging vs Discharging**: Separate capabilities with different constraints
   - Charging: Fixed target SoC
   - Discharging: Multiple stop conditions (price threshold, FCAS dispatch, SoC, duration)

2. **FCAS Two-Level Architecture**:
   - **Contract level** (planning): FcasContractStarted/Ended
   - **Dispatch level** (execution): FcasDispatchReceived/Completed
   - Must respond to dispatch within 1 second

3. **Economics Service**: Critical for conflict resolution between arbitrage opportunities and existing contracts

4. **Automation Levels**: MANUAL, SEMI_AUTO, FULL_AUTO - configurable per battery/site

5. **Event Frequency**:
   - `BatteryStateChanged`: Every 1 second (FCAS requirement)
   - `AemoPriceForecastReceived`: Every 5 minutes (AEMO schedule)
   - Conditional events only when thresholds exceeded

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design patterns and [EVENTS.md](EVENTS.md) for complete event flows.
