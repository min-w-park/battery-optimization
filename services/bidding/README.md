# Bidding Service

The Bidding Service is the first microservice with real business logic in the battery optimization system. It implements a simple arbitrage algorithm that makes real-time bidding decisions based on market prices and battery state.

## Purpose

The Bidding Service analyzes market conditions and battery state to identify profitable arbitrage opportunities:
- **Charge when prices are low** (< $50/MWh) and battery SoC < 80%
- **Discharge when prices are high** (> $100/MWh) and battery SoC > 30%
- Publish opportunity events for visibility
- Optionally issue command events based on automation mode

## Architecture

### Event-Driven Design

The Bidding Service is **100% event-driven** with no REST API. It:
- Subscribes to 3 event types from other services
- Maintains in-memory state (eventual consistency)
- Publishes 4 new event types for opportunities and commands
- Runs as a stateless service (no database)

### In-Memory State Management

The service uses two thread-safe caches:
- **BatteryStateCache**: Tracks current state of all batteries (SoC, power, operation state)
- **PriceCache**: Stores the latest market price

Caches are rebuilt from events on startup, demonstrating eventual consistency patterns.

## Arbitrage Algorithm

### Decision Logic

```
Charging Opportunity:
  - Price < $50/MWh (CHARGE_THRESHOLD)
  - SoC < 80% (MAX_SOC_FOR_CHARGE)
  - Battery in IDLE state

Discharging Opportunity:
  - Price > $100/MWh (DISCHARGE_THRESHOLD)
  - SoC > 30% (MIN_SOC_FOR_DISCHARGE)
  - Battery in IDLE state
```

### Implementation

See [internal/domain/arbitrage.go](internal/domain/arbitrage.go) for the core algorithm:
- `ShouldCharge(price, soc, operationState) bool`
- `ShouldDischarge(price, soc, operationState) bool`

## Automation Modes

The service supports two automation modes (configurable via `AUTOMATION_MODE` env var):

### MANUAL Mode (Default)
- Detects opportunities
- Publishes opportunity events only
- Human operator makes final decision
- No command events issued

### FULL_AUTO Mode
- Detects opportunities
- Publishes opportunity events
- **Automatically issues command events**
- Device Interface Service executes commands
- Full end-to-end automation

## Event Subscriptions

The Bidding Service subscribes to:

| Subject | Source | Frequency | Purpose |
|---------|--------|-----------|---------|
| `battery.registered.v1` | Asset Management | On-demand | Initialize battery in cache |
| `battery.state.changed.v1` | Telemetry | 1 Hz | Update battery state, evaluate opportunities |
| `market.price.updated.v1` | Market Data | 5-30 min | Update price, re-evaluate all batteries |

## Event Publications

The Bidding Service publishes:

| Subject | When | Automation Mode |
|---------|------|----------------|
| `charging.opportunity.detected.v1` | Low price + low SoC | MANUAL & FULL_AUTO |
| `charging.command.issued.v1` | After opportunity detected | FULL_AUTO only |
| `discharging.opportunity.detected.v1` | High price + high SoC | MANUAL & FULL_AUTO |
| `discharging.command.issued.v1` | After opportunity detected | FULL_AUTO only |

## Project Structure

```
services/bidding/
├── cmd/server/
│   └── main.go                    # Entry point, NATS subscriptions
├── internal/
│   ├── cache/
│   │   ├── battery_state_cache.go  # Thread-safe battery state cache
│   │   ├── price_cache.go          # Thread-safe price cache
│   │   └── *_test.go               # Cache tests (100% coverage)
│   ├── domain/
│   │   ├── arbitrage.go            # Core arbitrage algorithm
│   │   ├── bidding_decision.go     # BiddingDecision aggregate
│   │   ├── errors.go               # Domain-specific errors
│   │   └── *_test.go               # Domain tests (96.9% coverage)
│   └── service/
│       ├── bidding_engine.go       # Event handlers, decision logic
│       └── bidding_engine_test.go  # Service tests (88.2% coverage)
├── Dockerfile                      # Multi-stage build
├── go.mod                          # Go module definition
└── README.md                       # This file
```

## Development Commands

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...
```

### Build

```bash
# Local binary
go build -o bin/bidding ./cmd/server

# Docker image
docker build -t bidding-service .
```

### Run Locally

```bash
# Manual mode (default)
./bin/bidding

# Full auto mode
AUTOMATION_MODE=FULL_AUTO ./bin/bidding

# With custom NATS URL
NATS_URL=nats://localhost:4222 AUTOMATION_MODE=FULL_AUTO ./bin/bidding
```

### Run with Docker Compose

```bash
# Start infrastructure (NATS + databases)
docker-compose up -d

# Build and start bidding service
docker-compose up -d bidding

# View logs
docker-compose logs -f bidding

# Stop service
docker-compose down bidding
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `AUTOMATION_MODE` | `MANUAL` | `MANUAL` or `FULL_AUTO` |
| `LOG_LEVEL` | `info` | Logging level |

## Example Event Flows

### Charging Opportunity Flow

```
1. Market Data Service publishes: market.price.updated.v1 (price=$30)
   ↓
2. Bidding Service receives event, updates PriceCache
   ↓
3. For each battery in cache, evaluate arbitrage algorithm
   ↓
4. Battery has SoC=50%, IDLE → ShouldCharge() = true
   ↓
5. Publish: charging.opportunity.detected.v1
   ↓
6. If FULL_AUTO: Publish charging.command.issued.v1
   ↓
7. Device Interface Service receives command, starts charging
```

### Discharging Opportunity Flow

```
1. Telemetry Service publishes: battery.state.changed.v1 (SoC=70%)
   ↓
2. Bidding Service receives event, updates BatteryStateCache
   ↓
3. Get latest price from PriceCache (price=$150)
   ↓
4. ShouldDischarge(150, 70, IDLE) = true
   ↓
5. Publish: discharging.opportunity.detected.v1
   ↓
6. If FULL_AUTO: Publish discharging.command.issued.v1
   ↓
7. Device Interface Service receives command, starts discharging
```

## Integration Testing

See [test_integration.go](test_integration.go) for a complete integration test that:
1. Registers a battery
2. Tests charging opportunity (low price)
3. Tests discharging opportunity (high price)
4. Tests no opportunity (mid-range price)

Run with:
```bash
go run test_integration.go
```

Expected output:
- Battery registered in cache
- Charging opportunity detected + command issued
- Discharging opportunity detected + command issued
- No events for mid-range price

## Test Coverage

| Package | Coverage | Notes |
|---------|----------|-------|
| `internal/cache` | 100% | Thread-safe caches, concurrent tests |
| `internal/domain` | 96.9% | Arbitrage algorithm, decision logic |
| `internal/service` | 88.2% | Event handlers, business logic |
| **Overall** | **94%** | Exceeds 75% target |

All tests pass with `-race` flag (no race conditions detected).

## Design Decisions

### 1. Stateless Service (No Database)
- **Rationale**: In-memory caches rebuilt from events on startup
- **Benefit**: Simpler deployment, no migrations, eventual consistency
- **Trade-off**: State lost on restart (recovers from event replay)

### 2. Thread-Safe Caches
- **Pattern**: `sync.RWMutex` for concurrent reads/writes
- **Rationale**: High-frequency battery state updates (1 Hz)
- **Benefit**: Multiple goroutines can safely access shared state

### 3. Simple Arbitrage Algorithm
- **Current**: Fixed price thresholds ($50 charge, $100 discharge)
- **Future**: ML-based dynamic thresholds, multi-factor optimization
- **Rationale**: Start simple, validate architecture first

### 4. Event-First Design
- **Always**: Publish opportunity events (visibility)
- **Conditionally**: Publish command events (automation mode)
- **Benefit**: Audit trail, manual override capability

### 5. Best-Effort Event Publishing
- **Pattern**: Log warnings on publish failures, don't crash
- **Rationale**: Service degrades gracefully if NATS unavailable
- **Benefit**: More resilient to infrastructure issues

## Future Enhancements

1. **Dynamic Pricing Thresholds**: ML model to predict optimal charge/discharge prices
2. **Multi-Battery Optimization**: Coordinate multiple batteries for maximum profit
3. **FCAS Integration**: Consider frequency control contracts in bidding decisions
4. **REST API**: Health endpoint, manual override triggers
5. **Metrics**: Prometheus metrics for opportunities detected, commands issued
6. **Event Replay**: Rebuild state from event history on startup

## Related Documentation

- [M6 Overview](../../docs/milestones/M6-OVERVIEW.md) - Big picture and learning objectives
- [M6 Domain Spec](../../docs/milestones/M6-DOMAIN-SPEC.md) - Arbitrage algorithm specification
- [M6 API Spec](../../docs/milestones/M6-API-SPEC.md) - Event schemas and flows
- [M6 Checklist](../../docs/milestones/M6-CHECKLIST.md) - Implementation steps
- [EVENTS.md](../../EVENTS.md) - Complete event catalog
- [ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - System architecture

## Troubleshooting

### Service won't start
- Check NATS is running: `curl http://localhost:8222/healthz`
- Verify NATS_URL environment variable is correct
- Check logs for connection errors

### No events being processed
- Verify other services are publishing events
- Check NATS message count: `curl http://localhost:8222/varz | jq .in_msgs`
- Enable verbose logging: `LOG_LEVEL=debug`

### Opportunities not detected
- Check price thresholds in `internal/domain/arbitrage.go`
- Verify battery is in IDLE state
- Check SoC is within bounds (30-80%)
- Review automation mode setting

## License

This project is part of the Battery Optimization System educational project.
