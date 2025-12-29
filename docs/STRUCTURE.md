# Project Structure

This document describes the Go project structure for the battery optimization system.

## Overview

The project follows the **Standard Go Project Layout** with a microservices architecture. Each service is independently deployable with its own database and lifecycle.

## Directory Layout

```
battery-optimization/
├── services/                    # Microservices
│   ├── asset-management/       # Battery specs and constraints
│   ├── market-data/            # Energy market pricing
│   ├── telemetry/              # Real-time battery state
│   ├── device-interface/       # Hardware abstraction
│   └── bidding/                # Real-time bidding logic
├── pkg/                        # Shared libraries
│   ├── events/                 # Event definitions
│   ├── eventbus/               # NATS pub/sub wrapper
│   └── database/               # PostgreSQL utilities
├── docker-compose.yml          # Infrastructure setup
└── docs/                       # Architecture documentation
```

## Service Structure

Each service follows this standard layout:

```
service-name/
├── cmd/
│   └── server/
│       └── main.go            # Application entry point
├── internal/                  # Private application code
│   ├── domain/                # Domain models, aggregates, entities
│   ├── repository/            # Data access layer (implements ports)
│   ├── handler/               # HTTP/API handlers
│   └── service/               # Business logic (application services)
├── Dockerfile                 # Container image definition
└── go.mod                     # Go module definition (independent)
```

### Service-Specific Directories

Some services have specialized internal packages:

- **device-interface/internal/adapter/** - BatteryAdapter implementations (TeslaLike, BYDLike)
- **bidding/internal/strategy/** - Arbitrage algorithms and bidding strategies

## Shared Packages (`pkg/`)

These packages are imported by multiple services:

### `pkg/events`
- Event definitions (structs)
- Event schemas
- Event versioning

Example:
```go
type BatteryRegistered struct {
    BatteryID  string
    Capacity   float64
    MaxPower   float64
    RegisteredAt time.Time
}
```

### `pkg/eventbus`
- NATS connection management
- Publisher interface
- Subscriber interface
- Event serialization/deserialization

Example:
```go
type EventBus interface {
    Publish(topic string, event interface{}) error
    Subscribe(topic string, handler func([]byte)) error
}
```

### `pkg/database`
- PostgreSQL connection pooling
- Migration helpers
- Common database utilities

## Design Principles

### 1. Service Independence
- Each service has its own `go.mod` (no monorepo dependencies)
- Services can be built, tested, and deployed independently
- DB-per-service pattern (no shared databases)

### 2. Internal vs. Pkg
- **`internal/`**: Private to each service, cannot be imported by other services
- **`pkg/`**: Shared libraries that can be imported across services

### 3. Domain-Driven Design
Each service's `internal/domain/` contains:
- **Aggregates**: Consistency boundaries (e.g., Battery, MarketPrice)
- **Entities**: Objects with identity
- **Value Objects**: Immutable, compared by value
- **Domain Events**: What happened in the domain

### 4. Hexagonal Architecture (Ports & Adapters)
- **Handlers**: Incoming adapters (HTTP, gRPC)
- **Repository**: Outgoing adapters (Database)
- **Service**: Core business logic (independent of infrastructure)

## Service Responsibilities

### Asset Management (`services/asset-management`)
- Domain: Battery aggregate (ID, Capacity, MaxPower, RampRate)
- API: POST /batteries, GET /batteries/:id
- Events Published: `BatteryRegistered`
- Database: `asset-db` (port 5432)

### Market Data (`services/market-data`)
- Domain: MarketPrice (Timestamp, Value, Interval)
- API: GET /prices, POST /prices
- Events Published: `MarketPriceUpdated`
- Database: `market-db` (port 5433)

### Telemetry (`services/telemetry`)
- Domain: BatteryState (SoC, Power, Status)
- API: GET /telemetry/:batteryId/current
- Events Published: `BatteryStateChanged`
- Events Subscribed: `ChargeCommandIssued`
- Database: `telemetry-db` (port 5434)

### Device Interface (`services/device-interface`)
- Domain: Command, BatteryAdapter interface
- Events Subscribed: `ChargeCommandIssued`
- No database (stateless)
- Adapters: MockAdapter (TeslaLike, BYDLike)

### Bidding (`services/bidding`)
- Domain: BiddingDecision, ArbitrageStrategy
- Events Subscribed: `BatteryStateChanged`, `MarketPriceUpdated`
- Events Published: `BiddingDecisionMade`, `ChargeCommandIssued`
- No database (stateless decision engine)

## Development Workflow

### Creating a New Service

1. Create directory structure:
   ```bash
   mkdir -p services/new-service/cmd/server
   mkdir -p services/new-service/internal/{domain,repository,handler,service}
   ```

2. Initialize Go module:
   ```bash
   cd services/new-service
   go mod init github.com/yourusername/battery-optimization/services/new-service
   ```

3. Implement domain model in `internal/domain/`

4. Implement repository in `internal/repository/`

5. Implement business logic in `internal/service/`

6. Implement HTTP handlers in `internal/handler/`

7. Wire everything in `cmd/server/main.go`

8. Create Dockerfile

9. Add to docker-compose.yml

### Adding Shared Code

If code needs to be shared between services:

1. Add to appropriate `pkg/` directory
2. Document the public API
3. Version carefully (changes affect all services)

## Testing Strategy

Each service contains its tests alongside the code:

```
internal/
├── domain/
│   ├── battery.go
│   └── battery_test.go      # Table-driven tests
├── repository/
│   ├── postgres.go
│   └── postgres_test.go     # Integration tests with testcontainers
└── service/
    ├── battery_service.go
    └── battery_service_test.go  # Unit tests with mocks
```

## Future Additions

As the project grows, consider adding:

- `services/planning/` - Optimization planning service
- `services/alerts/` - Alerting and monitoring service
- `pkg/observability/` - Prometheus metrics, distributed tracing
- `pkg/auth/` - Authentication/authorization helpers
- `scripts/` - Build and deployment scripts
- `migrations/` - Database migration files per service
