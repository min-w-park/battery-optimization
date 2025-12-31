# Planning & Milestones

## Overview

This document tracks progress on building a battery optimization system using microservices architecture and event-driven patterns.

**Approach**: Milestone-based with flexible scheduling

---

## Milestones

### ✅ M0: Project Setup - COMPLETED
**Goal**: Development environment and project structure ready

**Tasks:**
- [x] GitHub repo created
- [x] README initial version
- [x] Go project structure designed
- [x] Docker Compose basic setup
- [x] PostgreSQL container ready
- [x] First commit

**Completion Criteria**:
- ✅ `docker-compose up` starts DB
- ✅ Project structure documented in README

---

### ✅ M1: Event Storming + Domain Events - COMPLETED
**Goal**: Identify what happens in the system as events

**Tasks:**
- [x] Event Storming session
  - [x] List all possible events in the system
  - [x] Sort chronologically
  - [x] Identify service boundaries
- [x] Define core events (25 total)
  - [x] Event names + fields
  - [x] Event schema (JSON)
- [x] Create `EVENTS.md` documentation
- [x] Plan for change management
  - [x] Event versioning strategy
  - [x] Backward compatibility for field additions

**Completion Criteria**:
- ✅ `EVENTS.md` contains event catalog (25 events)
- ✅ Each event specifies publishers/subscribers
- ✅ Event flow diagrams created (7 scenarios)

**Events Identified**:
- Initialization: 5 events
- Market Data: 3 events
- Charging: 7 events
- Discharging: 4 events
- Conflict Resolution: 5 events
- FCAS: 4 events
- Future Work: Identified but deferred

**Key Insights**:
- Charging and discharging are separate capabilities (not unified)
- Conflict resolution requires Economics Service
- FCAS has contract level (planning) and dispatch level (execution)
- Multiple stop conditions for operations
- Battery cannot charge/discharge simultaneously

---

### ✅ M2: Asset Management Service - COMPLETED (2025-12-29)
**Goal**: Build first service using DDD, TDD, and Hexagonal Architecture

**Documentation**:
- [M2 Overview](./docs/milestones/M2-OVERVIEW.md) - Big picture and learning objectives
- [Domain Specification](./docs/milestones/M2-DOMAIN-SPEC.md) - Battery aggregate and validation rules
- [API Specification](./docs/milestones/M2-API-SPEC.md) - REST endpoints and DTOs
- [Implementation Checklist](./docs/milestones/M2-CHECKLIST.md) - Step-by-step tasks
- [Service README](./services/asset-management/README.md) - Complete service documentation

**Phases**:
- [x] **Phase 1**: Project Setup (0.5-1h)
  - ✅ Directory structure (Hexagonal Architecture: cmd, internal/domain, internal/ports, internal/adapters)
  - ✅ Go 1.23 module initialization
  - ✅ Database migrations (golang-migrate with up/down)
  - ✅ Infrastructure verification (PostgreSQL 18)

- [x] **Phase 2**: Domain Layer (2-3h) - **TDD First!**
  - ✅ Domain errors (errors.go)
  - ✅ Battery aggregate with comprehensive tests (battery_test.go → battery.go)
  - ✅ 10 validation rules enforced
  - ✅ Constraints validation
  - ✅ **96.7% test coverage** (exceeds target!)

- [x] **Phase 3**: Repository Layer (2-3h)
  - ✅ Repository interface (ports/repository.go)
  - ✅ PostgreSQL implementation with integration tests
  - ✅ Error mapping (SQL → domain errors)
  - ✅ **87.5% test coverage**

- [x] **Phase 4**: HTTP API Layer (2-3h)
  - ✅ DTOs (CreateBatteryRequest, BatteryResponse, ErrorResponse)
  - ✅ HTTP handlers with comprehensive tests (using mocks)
  - ✅ Routes setup with gorilla/mux (POST, GET by ID, GET list)
  - ✅ **66.2% test coverage**
  - ✅ Error response formatting

- [x] **Phase 5**: Main Application (1-2h)
  - ✅ Dependency injection (cmd/server/main.go)
  - ✅ Configuration (environment variables)
  - ✅ Dockerfile (multi-stage build with Go 1.23)
  - ✅ docker-compose integration with health checks

- [x] **Phase 6**: Integration Testing (1-2h)
  - ✅ End-to-end tests with curl (POST, GET, LIST)
  - ✅ Database verification (data persists correctly)
  - ✅ Error case testing (validation, not found)
  - ✅ API filtering (location, status, pagination)

- [x] **Phase 7**: Polish & Documentation (1h)
  - ✅ Code formatting (go fmt, go vet)
  - ✅ Coverage verification (**79.8% overall**)
  - ✅ README created with comprehensive documentation
  - ✅ All phases complete

**Key Learning Objectives**:
- ✅ Domain-Driven Design (pure domain, no infrastructure leaking)
- ✅ Test-Driven Development (Red → Green → Refactor cycle)
- ✅ Hexagonal Architecture (ports & adapters pattern)
- ✅ Repository Pattern (abstract data access)
- ✅ Clean Code (SRP, DIP, separation of concerns)

**Completion Criteria**:
- ✅ All unit tests pass (domain, repository, handler)
- ✅ Test coverage **79.8%** overall (domain: 96.7%, repository: 87.5%, http: 66.2%)
- ✅ Service runs in Docker
- ✅ Can register battery via curl
- ✅ Can retrieve battery by ID
- ✅ Can list batteries with filters
- ✅ Validation errors return 400 with details
- ✅ Code follows Go conventions
- ✅ API documented with curl examples
- ✅ Manual end-to-end testing complete
- ✅ Service README with comprehensive docs

**Out of Scope** (deferred to later milestones):
- Event publishing (M4)
- Authentication/authorization
- Update/Delete operations
- Complex queries
- Caching

---

### ✅ M3: Market Data Service - COMPLETED (2025-12-29)
**Goal**: Second service + DB-per-service pattern

**Documentation**:
- [M3 Overview](./docs/milestones/M3-OVERVIEW.md) - Big picture and learning objectives
- [Domain Specification](./docs/milestones/M3-DOMAIN-SPEC.md) - MarketPrice aggregate and validation rules
- [API Specification](./docs/milestones/M3-API-SPEC.md) - REST endpoints and DTOs
- [Implementation Checklist](./docs/milestones/M3-CHECKLIST.md) - Step-by-step tasks
- [Service README](./services/market-data/README.md) - Complete service documentation

**Phases**:
- [x] **Phase 0**: M3 Milestone Documentation (0.5-1h)
  - ✅ M3-OVERVIEW.md - Service architecture and learning objectives
  - ✅ M3-DOMAIN-SPEC.md - MarketPrice aggregate with 8 validation rules
  - ✅ M3-API-SPEC.md - REST API endpoints and examples
  - ✅ M3-CHECKLIST.md - Step-by-step implementation guide

- [x] **Phase 1**: Project Setup (0.5-1h)
  - ✅ Directory structure (Hexagonal Architecture: cmd, internal/domain, internal/ports, internal/adapters)
  - ✅ Go 1.23 module initialization
  - ✅ Database migrations (golang-migrate v4.18.1)
  - ✅ Infrastructure verification (PostgreSQL 18 on port 5433)

- [x] **Phase 2**: Domain Layer (2-3h) - **TDD First!**
  - ✅ Domain errors (errors.go)
  - ✅ MarketPrice aggregate with comprehensive tests (price_test.go → price.go)
  - ✅ 8 validation rules enforced (price, region, interval alignment, etc.)
  - ✅ Interval alignment validation (5-min and 30-min boundaries)
  - ✅ **96.7% test coverage**

- [x] **Phase 3**: Repository Layer (2-3h)
  - ✅ Repository interface (ports/repository.go)
  - ✅ PostgreSQL implementation with integration tests
  - ✅ Time-range queries with filters (region, interval_type)
  - ✅ Pagination support (limit, offset)
  - ✅ Time-series optimized indexes
  - ✅ **91.1% test coverage**

- [x] **Phase 4**: HTTP API Layer (2-3h)
  - ✅ DTOs (CreateMarketPriceRequest, MarketPriceResponse, ErrorResponse)
  - ✅ HTTP handlers with comprehensive tests (using manual mocks)
  - ✅ Routes setup with gorilla/mux (POST, GET by ID, GET list with time-range)
  - ✅ ISO 8601 timestamp handling
  - ✅ **66.0% test coverage**

- [x] **Phase 5**: Main Application (1h)
  - ✅ Dependency injection (cmd/server/main.go)
  - ✅ Configuration (environment variables)
  - ✅ Dockerfile (multi-stage build with Go 1.23)
  - ✅ docker-compose integration with market-db on port 8081

- [x] **Phase 6**: Integration Testing (1h)
  - ✅ End-to-end tests with curl (POST, GET, LIST with time-range)
  - ✅ Database verification (data persists correctly)
  - ✅ Error case testing (negative price, invalid region, duplicate interval)
  - ✅ Time-range query testing
  - ✅ Filter testing (region, interval_type)
  - ✅ Pagination testing (limit, offset)
  - ✅ DB-per-service isolation verified

- [x] **Phase 7**: Polish & Documentation (1h)
  - ✅ Code formatting (go fmt, go vet)
  - ✅ Coverage verification (**84.6% overall**)
  - ✅ README created with comprehensive documentation
  - ✅ PLANNING.md updated
  - ✅ All phases complete

**Key Learning Objectives**:
- ✅ Time-series data modeling (interval alignment)
- ✅ DB-per-service pattern (market-db separate from asset-db)
- ✅ Time-range queries with multiple filters
- ✅ AEMO market data concepts (5MIN/30MIN predispatch)
- ✅ Replicating M2 patterns successfully

**Completion Criteria**:
- ✅ All unit tests pass (domain, repository, handler)
- ✅ Test coverage **84.6%** overall (domain: 96.7%, repository: 91.1%, http: 66.0%)
- ✅ Service runs in Docker on port 8081
- ✅ Can create market prices via curl
- ✅ Can retrieve prices by ID
- ✅ Can query prices by time range
- ✅ Filtering by region and interval_type works
- ✅ Pagination works correctly
- ✅ Validation errors return 400 with details
- ✅ Duplicate interval returns 409 Conflict
- ✅ Data persists in PostgreSQL
- ✅ 2 services running independently (asset-management:8080, market-data:8081)
- ✅ DB-per-service isolation verified (asset-db on 5432, market-db on 5433)
- ✅ Service README with comprehensive docs

**Out of Scope** (deferred to later milestones):
- Event publishing (M4)
- AEMO data ingestion automation
- Price forecast calculations
- Authentication/authorization
- Update/Delete operations

---

### ✅ M4: Event Bus Integration (NATS) - COMPLETED (2025-12-30)
**Goal**: Implement event-driven architecture

**Documentation**:
- [M4 Overview](./docs/milestones/M4-OVERVIEW.md) - Event-driven architecture and learning objectives
- [Domain Specification](./docs/milestones/M4-DOMAIN-SPEC.md) - Event schemas and versioning strategy
- [API Specification](./docs/milestones/M4-API-SPEC.md) - NATS pub/sub patterns and subject naming
- [Implementation Checklist](./docs/milestones/M4-CHECKLIST.md) - Step-by-step tasks
- [Events Library README](./pkg/events/README.md) - Complete usage guide

**Phases**:
- [x] **Phase 0**: M4 Milestone Documentation (1.5h)
  - ✅ M4-OVERVIEW.md - Event-driven architecture overview
  - ✅ M4-DOMAIN-SPEC.md - Event schemas and versioning
  - ✅ M4-API-SPEC.md - NATS subject patterns and error handling
  - ✅ M4-CHECKLIST.md - Phase-by-phase implementation guide

- [x] **Phase 1**: Common Event Library (3h)
  - ✅ pkg/events package created
  - ✅ BatteryRegistered event with BatteryConstraints
  - ✅ MarketPriceUpdated event
  - ✅ BatteryStateChanged event (placeholder for M5)
  - ✅ EventPublisher and EventSubscriber interfaces
  - ✅ Comprehensive event tests (JSON serialization)

- [x] **Phase 2**: NATS Publisher Adapter (3h)
  - ✅ NATSPublisher implementation
  - ✅ Smart flush handling (FlushWithContext vs Flush)
  - ✅ Context timeout support
  - ✅ JSON serialization
  - ✅ Error handling and connection management
  - ✅ Publisher tests (8/8 passing)

- [x] **Phase 3**: NATS Subscriber Adapter (3h)
  - ✅ NATSSubscriber implementation
  - ✅ Wildcard subscription support (`*` and `>`)
  - ✅ Event handler pattern
  - ✅ Multiple concurrent subscriptions
  - ✅ Error handling in handlers
  - ✅ Subscriber tests (11/11 passing)

- [x] **Phase 4**: Asset Service Integration (3h)
  - ✅ Added EventPublisher port
  - ✅ Updated BatteryHandler to publish BatteryRegistered
  - ✅ Modified main.go with NATS connection
  - ✅ Best-effort publishing (warnings logged, HTTP not blocked)
  - ✅ All tests updated and passing

- [x] **Phase 5**: Market Service Integration (3h)
  - ✅ Added EventPublisher port
  - ✅ Updated MarketPriceHandler to publish MarketPriceUpdated
  - ✅ Modified main.go with NATS connection (port 8081)
  - ✅ Best-effort publishing pattern
  - ✅ All tests updated and passing

- [x] **Phase 6**: Test Subscriber Tool (2h)
  - ✅ Created tools/event-subscriber/ CLI tool
  - ✅ Color-coded output (timestamps, subjects, payloads)
  - ✅ Pretty-printed JSON formatting
  - ✅ Wildcard subscription support
  - ✅ Event counter tracking
  - ✅ Graceful shutdown (Ctrl+C)
  - ✅ Comprehensive README with usage examples

- [x] **Phase 7**: Integration Testing (3h)
  - ✅ Verified NATS health and connectivity
  - ✅ All pkg/events tests passing (27/27)
  - ✅ Test coverage: **88.9%** (exceeds >80% target)
  - ✅ Both services build successfully
  - ✅ Infrastructure health verified
  - ✅ Event flow tested end-to-end

- [x] **Phase 8**: Polish & Documentation (2h)
  - ✅ Code formatting (go fmt on all packages)
  - ✅ Static analysis (go vet on all packages)
  - ✅ pkg/events/README.md created (comprehensive usage guide)
  - ✅ PLANNING.md updated with M4 completion
  - ✅ All phases complete

**Key Learning Objectives**:
- ✅ Event-Driven Architecture (pub/sub patterns, eventual consistency)
- ✅ NATS Integration (connect, publish, subscribe, monitor)
- ✅ Event Versioning (backward compatibility strategy)
- ✅ Publisher/Subscriber Pattern (loose coupling)
- ✅ Testing Asynchronous Systems (test subscribers, event flow verification)

**Completion Criteria**:
- ✅ All unit tests pass (events, publisher, subscriber)
- ✅ Test coverage **88.9%** for pkg/events (exceeds >80% target)
- ✅ NATS running and healthy (docker-compose)
- ✅ Asset Service publishes BatteryRegistered events
- ✅ Market Service publishes MarketPriceUpdated events
- ✅ Test subscriber tool receives and displays events
- ✅ Wildcard subscriptions working (`>`, `battery.>`, `*.*.v1`)
- ✅ Event versioning implemented (subject + payload version)
- ✅ Best-effort publishing pattern (graceful degradation)
- ✅ NATS monitoring endpoints accessible (http://localhost:8222)
- ✅ Services continue working if NATS unavailable
- ✅ Comprehensive documentation (README, API examples, troubleshooting)

**Key Achievements**:
- Event library with 3 events (BatteryRegistered, MarketPriceUpdated, BatteryStateChanged placeholder)
- NATS publisher/subscriber adapters with comprehensive error handling
- 2 services retrofitted with event publishing (Asset Management, Market Data)
- Test subscriber tool for real-time event monitoring
- Event versioning strategy implemented from day one
- 27 tests passing with 88.9% coverage

**Out of Scope** (deferred to later milestones):
- Event persistence/replay (future: event sourcing)
- Complex subscribers with business logic (M6: Bidding Service)
- Dead letter queues
- Event schema registry
- CQRS patterns

---

### ✅ M5: Telemetry + Device Interface - COMPLETED
**Goal**: Hardware abstraction + real-time data

**Tasks:**
- [x] Define `BatteryAdapter` interface
  ```go
  type BatteryAdapter interface {
      GetState(ctx) (BatteryState, error)
      SendCommand(ctx, Command) error
      GetCustomAttributes() map[string]interface{}
      GetBatteryID() string
  }
  ```
- [x] Implement TeslaLike mock adapter with realistic simulation
  - 5 MW/s ramp rate, 95% efficiency, temperature physics
  - Background goroutine (100ms ticks) for state updates
  - Thread-safe with sync.RWMutex
- [x] Telemetry Service
  - PostgreSQL time-series storage
  - `GET /telemetry/:batteryId/current`
  - `GET /telemetry/:batteryId/history`
  - Event publishing integration (placeholder for 1 Hz)
- [x] Device Interface Service
  - Subscribe to `charging.command.issued.v1`
  - Subscribe to `discharging.command.issued.v1`
  - Subscribe to `conflict.resolved.v1`
  - Publish `battery.connection.established.v1`
  - Publish `charging.started.v1` / `discharging.started.v1`
- [x] Event library extended with 7 new events
- [x] Command handler service with event-driven architecture
- [x] Handle CustomAttributes (JSONB in database)

**Completion Criteria**:
- ✅ Mock battery simulation running (TeslaLike adapter)
- ✅ State changes tracked (SoC, power, temperature updating)
- ✅ Event-driven command flow working end-to-end
- ✅ Test coverage: Telemetry 81.5%, Device Interface 78.5%
- ✅ Integration testing validated

---

### ✅ M6: Bidding Service - COMPLETED (2025-12-31)
**Goal**: Real-time decision logic with arbitrage algorithm

**Documentation**:
- [M6 Overview](./docs/milestones/M6-OVERVIEW.md) - Big picture and learning objectives
- [Domain Specification](./docs/milestones/M6-DOMAIN-SPEC.md) - Arbitrage algorithm and business rules
- [API Specification](./docs/milestones/M6-API-SPEC.md) - Event schemas and flows
- [Implementation Checklist](./docs/milestones/M6-CHECKLIST.md) - Step-by-step tasks
- [Service README](./services/bidding/README.md) - Complete service documentation

**Phases**:
- [x] **Phase 0**: M6 Milestone Documentation (1.5h)
  - ✅ M6-OVERVIEW.md - Service architecture and learning objectives
  - ✅ M6-DOMAIN-SPEC.md - Arbitrage algorithm specification (CHARGE_THRESHOLD=$50, DISCHARGE_THRESHOLD=$100)
  - ✅ M6-API-SPEC.md - Event subscriptions and publications
  - ✅ M6-CHECKLIST.md - Implementation guide

- [x] **Phase 1**: Domain Layer - Arbitrage Algorithm (2-3h) - **TDD First!**
  - ✅ Domain errors (errors.go)
  - ✅ Arbitrage algorithm with comprehensive tests (arbitrage_test.go → arbitrage.go)
  - ✅ ShouldCharge() and ShouldDischarge() decision logic
  - ✅ BiddingDecision aggregate with validation
  - ✅ **96.9% test coverage**

- [x] **Phase 2**: Event Library Extensions (1-2h)
  - ✅ ChargingOpportunityDetected event
  - ✅ DischargingOpportunityDetected event
  - ✅ ChargingCommandIssued event (updated to M6 schema)
  - ✅ DischargingCommandIssued event (updated to M6 schema)
  - ✅ All events tested with JSON serialization

- [x] **Phase 3**: In-Memory State Caches (2-3h)
  - ✅ BatteryStateCache with sync.RWMutex (thread-safe)
  - ✅ PriceCache with sync.RWMutex (thread-safe)
  - ✅ Comprehensive concurrent testing (100+ goroutines)
  - ✅ **100% test coverage**
  - ✅ No race conditions (verified with -race flag)

- [x] **Phase 4**: Bidding Engine Service (2-3h)
  - ✅ BiddingEngine with event handlers
  - ✅ HandleBatteryStateChanged - updates cache, runs arbitrage, publishes events
  - ✅ HandleMarketPriceUpdated - updates price, re-evaluates all batteries
  - ✅ HandleBatteryRegistered - initializes battery in cache
  - ✅ MANUAL and FULL_AUTO automation modes
  - ✅ **88.2% test coverage**

- [x] **Phase 5**: Main Application Setup (2-3h)
  - ✅ main.go with dependency injection
  - ✅ NATS subscriptions (3 event types)
  - ✅ Configuration (NATS_URL, AUTOMATION_MODE, LOG_LEVEL)
  - ✅ Dockerfile (multi-stage build)
  - ✅ docker-compose integration
  - ✅ Graceful shutdown

- [x] **Phase 6**: Integration Testing (2-3h)
  - ✅ Infrastructure verification (NATS healthy)
  - ✅ Event-driven flow tested (battery registration → price updates → opportunities)
  - ✅ Charging opportunity detection (low price)
  - ✅ Discharging opportunity detection (high price)
  - ✅ No opportunity for mid-range prices
  - ✅ FULL_AUTO mode command issuance verified
  - ✅ Integration test script created (test_integration.go)

- [x] **Phase 7**: Polish & Documentation (1-2h)
  - ✅ Code formatting (go fmt, go vet)
  - ✅ Coverage verification (**94% overall**: domain 96.9%, cache 100%, service 88.2%)
  - ✅ Service README created (comprehensive documentation)
  - ✅ PLANNING.md updated
  - ✅ All phases complete

**Key Learning Objectives**:
- ✅ Event-Driven Business Logic (first service with real decision-making)
- ✅ In-Memory State Management (eventual consistency, thread-safe caches)
- ✅ Arbitrage Algorithm (simple threshold-based trading logic)
- ✅ Automation Modes (MANUAL vs FULL_AUTO)
- ✅ Stateless Service Design (no database, rebuilt from events)

**Completion Criteria**:
- ✅ All unit tests pass (domain, cache, service)
- ✅ Test coverage **94% overall** (exceeds >75% target)
- ✅ No race conditions detected (go test -race)
- ✅ Service subscribes to 3 event types
- ✅ Service publishes 4 event types (opportunity + command)
- ✅ MANUAL mode: only publishes opportunity events
- ✅ FULL_AUTO mode: publishes opportunity + command events
- ✅ Arbitrage algorithm correctly detects opportunities
- ✅ End-to-end event flow verified with integration test
- ✅ Comprehensive documentation complete

**Key Achievements**:
- First service with real business logic (not just CRUD)
- Stateless, event-driven architecture
- Thread-safe in-memory caches (sync.RWMutex patterns)
- Simple but functional arbitrage algorithm
- Configurable automation modes
- 4 new event types for bidding operations
- Integration test demonstrating full event flow
- Exceeds all test coverage targets

---

### ⏳ M7: Documentation + Polish
**Goal**: Ready to share

**Tasks:**
- [ ] Complete architecture diagram
- [ ] Finalize README
  - Why event-driven?
  - Service boundaries explanation
  - How to run locally
- [ ] Production TODO list
  - Planning Service
  - Alert Service
  - Observability (Prometheus/Grafana)
  - CI/CD pipeline
  - Kubernetes deployment
- [ ] Document demo scenario
- [ ] (Optional) Demo video or animated GIF

**Completion Criteria**:
- External person can understand project from README
- System runs with `docker-compose up`

---

## Notes

- Milestones are ordered but timing is flexible
- Complete criteria before moving to next milestone
- If stuck, document blocker and move to parallel task
