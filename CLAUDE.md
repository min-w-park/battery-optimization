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

Three PostgreSQL databases run in separate containers:
- `asset-db` on port 5432 (DB: asset_management, user: asset_user)
- `market-db` on port 5433 (DB: market_data, user: market_user)
- `telemetry-db` on port 5434 (DB: telemetry, user: telemetry_user)

NATS event bus:
- Client connections: port 4222
- HTTP monitoring: port 8222 (health check: http://localhost:8222/healthz)

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
# Connect to Asset Management DB
docker exec -it asset-db psql -U asset_user -d asset_management

# Connect to Market Data DB
docker exec -it market-db psql -U market_user -d market_data

# Connect to Telemetry DB
docker exec -it telemetry-db psql -U telemetry_user -d telemetry

# Check database version
docker exec -it asset-db psql -U asset_user -d asset_management -c "SELECT version();"
```

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

Core domain events (see EVENTS.md when created):
- `BatteryRegistered` - Asset Management publishes when battery added
- `MarketPriceUpdated` - Market Data publishes on price changes
- `BatteryStateChanged` - Telemetry publishes on SoC/power changes
- `BiddingDecisionMade` - Bidding Service publishes charge/discharge decisions
- `ChargeCommandIssued` - Device Interface executes via BatteryAdapter

## Service Development Pattern

When implementing a new service:

1. **Domain Model First**: Define aggregates, entities, value objects
2. **Repository Pattern**: Abstract data access with interfaces
3. **REST API**: Expose service capabilities (CRUD operations)
4. **Event Emission**: Publish domain events at aggregate boundaries
5. **Event Subscription**: React to relevant events from other services
6. **Dockerfile**: Containerize the service
7. **docker-compose.yml**: Integrate into infrastructure

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

Currently in M0 (Project Setup). See PLANNING.md for milestone tracking:
- M0: Infrastructure setup (Docker Compose, PostgreSQL, NATS)
- M1: Event Storming + Domain Events documentation
- M2-M6: Service implementation (Asset → Market → Events → Telemetry → Bidding)
- M7: Documentation and polish

## Key Design Decisions

- **Go as primary language**: Strong concurrency primitives, explicit error handling
- **NATS over Kafka**: Lightweight, simpler ops for learning context
- **DB-per-service**: True service isolation, independent schemas
- **Event-driven over RPC**: Temporal decoupling, better fault isolation
- **Mock adapters**: Hardware simulation without physical dependencies
- **Australian NEM context**: Real-world market constraints and pricing patterns
