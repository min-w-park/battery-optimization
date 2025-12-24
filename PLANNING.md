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

### ⏳ M2: Asset Management Service 
**Goal**: Build first service properly with DDD

**Tasks:**
- [ ] Domain model design (Battery Aggregate)
  - Battery ID, Capacity, Max Power, Ramp Rate
  - Constraint validation
- [ ] Repository pattern implementation
- [ ] REST API (CRUD)
  - `POST /batteries` - Register battery
  - `GET /batteries/:id` - Query specs
- [ ] TDD: Minimum 3 test cases
- [ ] Dockerfile + docker-compose integration
- [ ] Prepare `BatteryRegistered` event (emit in M4)

**Completion Criteria**:
- API works and tests pass
- Can register/query batteries via Postman/curl
- Validation errors returned for constraint violations
- Event emission points marked in code comments

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
