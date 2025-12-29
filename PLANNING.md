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

### ⏳ M3: Market Data Service
**Goal**: Second service + DB-per-service pattern

**Tasks:**
- [ ] Market domain model
  - Price (timestamp, value, interval)
  - Time-series data storage
- [ ] Mock AEMO price data generator
- [ ] REST API
  - `GET /prices?from=X&to=Y` - Query prices
  - `POST /prices` - Input price data
- [ ] Independent PostgreSQL database
- [ ] Prepare `MarketPriceUpdated` event

**Completion Criteria**:
- 2 services running independently
- Market data query API functional
- DB-per-service isolation verified

---

### ⏳ M4: Event Bus Integration (NATS)
**Goal**: Implement event-driven architecture

**Tasks:**
- [ ] NATS installation and Docker Compose integration
- [ ] Event publisher/subscriber common library
- [ ] Implement events from M1
  - `BatteryRegistered`
  - `MarketPriceUpdated`
  - `BatteryStateChanged`
- [ ] Asset Service → publish events
- [ ] Market Service → publish events
- [ ] Simple subscriber test (console output)

**Completion Criteria**:
- NATS pub/sub operational
- Services communicate via events
- Event flow verifiable

---

### ⏳ M5: Telemetry + Device Interface
**Goal**: Hardware abstraction + real-time data

**Tasks:**
- [ ] Define `BatteryAdapter` interface
  ```go
  type BatteryAdapter interface {
      GetState(ctx) (BatteryState, error)
      SendCommand(ctx, Command) error
  }
  ```
- [ ] Implement MockAdapter (2 types: TeslaLike, BYDLike)
- [ ] Telemetry Service
  - Real-time SoC tracking
  - `GET /telemetry/:batteryId/current`
  - Publish `BatteryStateChanged` events
- [ ] Device Interface Service (separate)
  - Subscribe to `ChargeCommandIssued` events
  - Send commands via adapter
- [ ] Handle CustomAttributes

**Completion Criteria**:
- Mock battery simulation running
- State changes emit events
- 2 adapter types are swappable

---

### ⏳ M6: Bidding Service
**Goal**: Real-time decision logic

**Tasks:**
- [ ] Bidding Service
  - Simple arbitrage algorithm
    ```
    if (price > threshold && SoC > 30%):
        discharge bid
    if (price < threshold && SoC < 80%):
        charge bid
    ```
  - Subscribe to `BatteryStateChanged`
  - Subscribe to `MarketPriceUpdated`
  - Publish `BiddingDecisionMade` event
- [ ] End-to-end flow test
  - Price change → Bidding decision → Device command
- [ ] Verify eventual consistency

**Completion Criteria**:
- Price changes trigger automatic bidding decisions
- End-to-end scenario operational
- 4-5 services cooperating via events

---

### ⏳ M7: Documentation + Polish
**Goal**: Ready to share with Seb

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
- Ready to contact Seb

---

## Notes

- Milestones are ordered but timing is flexible
- Complete criteria before moving to next milestone
- If stuck, document blocker and move to parallel task
