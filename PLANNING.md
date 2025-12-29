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

### ⏳ M2: Asset Management Service (10-12 hours)
**Goal**: Build first service using DDD, TDD, and Hexagonal Architecture

**Documentation**:
- [M2 Overview](./docs/milestones/M2-OVERVIEW.md) - Big picture and learning objectives
- [Domain Specification](./docs/milestones/M2-DOMAIN-SPEC.md) - Battery aggregate and validation rules
- [API Specification](./docs/milestones/M2-API-SPEC.md) - REST endpoints and DTOs
- [Implementation Checklist](./docs/milestones/M2-CHECKLIST.md) - Step-by-step tasks

**Phases**:
- [ ] **Phase 1**: Project Setup (0.5-1h)
  - Directory structure (cmd, internal/domain, internal/ports, internal/adapters)
  - Go module initialization
  - Database migration (batteries table)

- [ ] **Phase 2**: Domain Layer (2-3h) - **TDD First!**
  - Domain errors (errors.go)
  - Battery aggregate with tests (battery_test.go → battery.go)
  - Validation rules (11 business rules)
  - Constraints validation
  - >80% test coverage

- [ ] **Phase 3**: Repository Layer (2-3h)
  - Repository interface (ports/repository.go)
  - PostgreSQL implementation with tests
  - Error mapping (SQL → domain errors)

- [ ] **Phase 4**: HTTP API Layer (2-3h)
  - DTOs (CreateBatteryRequest, BatteryResponse)
  - HTTP handlers with tests (using mocks)
  - Routes setup (POST /batteries, GET /batteries/:id, GET /batteries)
  - Error response formatting

- [ ] **Phase 5**: Main Application (1-2h)
  - Dependency injection (cmd/server/main.go)
  - Configuration (environment variables)
  - Dockerfile (multi-stage build)
  - docker-compose integration

- [ ] **Phase 6**: Integration Testing (1-2h)
  - End-to-end tests with curl
  - Database verification
  - Error case testing
  - API documentation with examples

- [ ] **Phase 7**: Polish & Documentation (1h)
  - Code formatting (go fmt, go vet)
  - Coverage verification
  - README updates
  - Git commit and tag (v0.1.0-m2)

**Key Learning Objectives**:
- Domain-Driven Design (pure domain, no infrastructure leaking)
- Test-Driven Development (write tests first, >80% coverage)
- Hexagonal Architecture (ports & adapters pattern)
- Repository Pattern (abstract data access)
- Clean Code (SRP, DIP, separation of concerns)

**Completion Criteria**:
- ✅ All unit tests pass (domain, repository, handler)
- ✅ Test coverage > 80%
- ✅ Service runs in Docker
- ✅ Can register battery via curl
- ✅ Can retrieve battery by ID
- ✅ Can list batteries with filters
- ✅ Validation errors return 400 with details
- ✅ Code follows Go conventions
- ✅ API documented with curl examples
- ✅ Manual end-to-end testing complete

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
