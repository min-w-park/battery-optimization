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

### Continuous Integration

The project uses GitHub Actions for automated testing and building:

**Automated Tests** (`.github/workflows/test.yml`):
- Runs on every push to `develop` and `main` branches
- Runs on all pull requests
- Executes all unit tests with race detection (`-race`)
- Generates coverage reports
- Runs `go vet` for static analysis
- Runs `golangci-lint` for code quality

**Docker Image Builds** (`.github/workflows/build.yml`):
- Builds all 5 service Docker images
- Tags images with branch name, PR number, or semantic version
- Pushes images to GitHub Container Registry (ghcr.io)
- Uses Docker layer caching for faster builds

**Running Tests Locally** (matches CI environment):
```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -func=coverage.out

# Run static analysis
go vet ./...

# Run linter (requires golangci-lint installed)
golangci-lint run --timeout=5m
```

### Health Check Endpoints

All HTTP-based services expose health check endpoints for monitoring:

```bash
# Liveness check - is the service running?
curl http://localhost:8080/health/live

# Readiness check - is the service ready to handle traffic?
curl http://localhost:8080/health/ready
```

**Service Ports**:
- Asset Management: 8080
- Market Data: 8080
- Telemetry: 8082

**Response Examples**:
```json
// Liveness
{"status": "ok", "service": "asset-management"}

// Readiness (healthy)
{
  "status": "ready",
  "service": "asset-management",
  "checks": {
    "database": "healthy",
    "nats": "not configured"
  }
}

// Readiness (unhealthy)
{
  "status": "unavailable",
  "service": "asset-management",
  "checks": {
    "database": "unhealthy: connection timeout",
    "nats": "not checked"
  }
}
```

See [docs/operations/HEALTH-CHECKS.md](docs/operations/HEALTH-CHECKS.md) for detailed documentation on health check endpoints, Kubernetes integration, and troubleshooting.

See [docs/operations/CI-CD.md](docs/operations/CI-CD.md) for detailed workflow documentation.

## Development Approach

### Test-Driven Development (TDD)

**CRITICAL**: This project follows strict TDD (Test-Driven Development). See [CONTRIBUTING.md](CONTRIBUTING.md) for complete workflow and [.claude/skills/TDD-SKILL.md](.claude/skills/TDD-SKILL.md) for detailed patterns.

**The Golden Rule**: Red → Green → Refactor
1. ❌ Write a failing test
2. ✅ Write minimal code to pass
3. ♻️ Refactor while tests protect you

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

**Core Documentation**:
- **[STRUCTURE.md](docs/STRUCTURE.md)**: Go project structure, service layout, and development workflow
- **[EVENTS.md](EVENTS.md)**: Complete event catalog (25 events) with JSON schemas and flow diagrams
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)**: System architecture with mermaid diagrams, service boundaries, and design patterns
- **[PLANNING.md](PLANNING.md)**: Milestone tracking and project status
- **[QUICKSTART.md](docs/QUICKSTART.md)**: Quick start guide for development setup
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Development philosophy and workflow (TDD, code review)

**Development Guides**:
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Development philosophy and TDD workflow (Kent Beck style) - **READ THIS FIRST**
- **[Battery Domain Skill](.claude/skills/BATTERY-DOMAIN-SKILL.md)**: Domain concepts, validation rules, examples
- **[TDD Skill](.claude/skills/TDD-SKILL.md)**: TDD workflow, patterns, and anti-patterns

**Milestone Documentation**:
- **M2 (Asset Management Service)**:
  - [M2 Overview](docs/milestones/M2-OVERVIEW.md): Big picture and learning objectives
  - [M2 Domain Spec](docs/milestones/M2-DOMAIN-SPEC.md): Battery aggregate and validation rules
  - [M2 API Spec](docs/milestones/M2-API-SPEC.md): REST endpoints and DTOs
  - [M2 Checklist](docs/milestones/M2-CHECKLIST.md): Step-by-step implementation guide
- **M3 (Market Data Service)**:
  - [M3 Overview](docs/milestones/M3-OVERVIEW.md): Time-series data and DB-per-service pattern
  - [M3 Domain Spec](docs/milestones/M3-DOMAIN-SPEC.md): MarketPrice aggregate and validation
  - [M3 API Spec](docs/milestones/M3-API-SPEC.md): Time-range queries and filtering
  - [M3 Checklist](docs/milestones/M3-CHECKLIST.md): Implementation guide
- **M4 (Event Bus Integration)**:
  - [M4 Overview](docs/milestones/M4-OVERVIEW.md): Event-driven architecture with NATS
  - [M4 Domain Spec](docs/milestones/M4-DOMAIN-SPEC.md): Event schemas and versioning
  - [M4 API Spec](docs/milestones/M4-API-SPEC.md): NATS pub/sub patterns
  - [M4 Checklist](docs/milestones/M4-CHECKLIST.md): Event integration guide
  - [pkg/events README](pkg/events/README.md): Event library usage guide

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

## Event Publishing and Subscribing

All services use the **pkg/events** library for event-driven communication via NATS. See [pkg/events/README.md](pkg/events/README.md) for complete usage guide.

### Publishing Events

Services publish domain events when aggregates change state:

```go
import "github.com/minwook/battery-optimization/pkg/events"

// 1. In main.go - Connect to NATS
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

// 2. In handler - Publish event after DB save
if h.publisher != nil {
    event := events.BatteryRegistered{
        BatteryID:    battery.ID,
        Capacity:     battery.Capacity,
        // ... map all fields ...
        Timestamp:    time.Now(),
        EventVersion: "v1",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := h.publisher.Publish(ctx, "battery.registered.v1", event); err != nil {
        // Best-effort: Log warning but don't fail the HTTP request
        log.Printf("WARNING: Failed to publish event: %v", err)
    }
}
```

**Key Patterns**:
- **Best-effort publishing**: Event failures don't block HTTP responses
- **Optional publisher**: Services work gracefully when NATS unavailable
- **Context timeouts**: Always use 2-second timeout for publishing
- **Subject versioning**: Include version in subject (`battery.registered.v1`)
- **Payload versioning**: Set `event_version` field to "v1"

### Subscribing to Events

Services subscribe to relevant events from other services:

```go
import "github.com/minwook/battery-optimization/pkg/events"

// 1. Connect to NATS
subscriber, err := events.NewNATSSubscriber("nats://localhost:4222")
if err != nil {
    log.Fatalf("Failed to connect to NATS: %v", err)
}
defer subscriber.Close()

// 2. Define handler
handler := func(subject string, data []byte) error {
    if subject == "battery.registered.v1" {
        var event events.BatteryRegistered
        if err := json.Unmarshal(data, &event); err != nil {
            return fmt.Errorf("failed to unmarshal: %w", err)
        }
        log.Printf("Battery registered: %s", event.BatteryID)
    }
    return nil
}

// 3. Subscribe with wildcard
ctx := context.Background()
if err := subscriber.Subscribe(ctx, "battery.>", handler); err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}
```

**Wildcard Patterns**:
- `>` - All events
- `battery.>` - All battery events
- `market.>` - All market events
- `*.*.v1` - All v1 events from any service

### Implemented Events (M4)

**Currently Publishing**:
- `battery.registered.v1` - Published by Asset Management Service when battery created
- `market.price.updated.v1` - Published by Market Data Service when price created
- `battery.state.changed.v1` - Placeholder for M5 (Telemetry Service)

**Event Versioning Strategy**:
- Backward-compatible changes (add fields): Keep same version
- Breaking changes (remove/rename fields): Create new version (v2)

### Testing Events

Use the event subscriber tool for manual testing:

```bash
# Terminal 1: Subscribe to all events
cd tools/event-subscriber
go run main.go

# Terminal 2: Trigger event
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{ ... }'

# See event appear in Terminal 1
```

For automated testing, see [pkg/events/README.md](pkg/events/README.md).

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
- **M2: Asset Management Service** - Battery domain model, REST API, 79.8% test coverage
- **M3: Market Data Service** - Time-series data, DB-per-service pattern, 84.6% test coverage
- **M4: Event Bus Integration** - NATS pub/sub, event library (88.9% coverage), 2 services publishing events
- **M5: Telemetry + Device Interface** - Hardware abstraction, real-time monitoring
  - Telemetry Service: 86.1% coverage, time-series storage, REST API
  - Device Interface Service: 77.5% coverage, BatteryAdapter interface, TeslaLike mock adapter
  - 7 new events, end-to-end event flow validated
- **M6: Bidding Service** - Real-time arbitrage decision engine
  - 94% test coverage overall (domain: 96.9%, cache: 100%, service: 88.2%)
  - Simple arbitrage algorithm (charge < $50/MWh, discharge > $100/MWh)
  - In-memory state management with thread-safe caches (sync.RWMutex)
  - Event-driven architecture (subscribes to 3 event types, publishes 4)
  - Automation modes: MANUAL and FULL_AUTO
  - Stateless service (no database)
  - 4 new events: ChargingOpportunityDetected, DischargingOpportunityDetected, ChargingCommandIssued, DischargingCommandIssued

**🚧 Current**: M7 - Documentation polish and final verification

See [PLANNING.md](PLANNING.md) for detailed milestone tracking and task breakdowns.

**System Status**: MVP Complete! 5 microservices implemented with event-driven architecture.
- Services: Asset Management, Market Data, Telemetry, Device Interface, Bidding
- Events: 13/25 events implemented (see [EVENTS.md](EVENTS.md) for full status)
- Infrastructure: NATS 2.10, PostgreSQL 18 (3 databases), Docker Compose

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
