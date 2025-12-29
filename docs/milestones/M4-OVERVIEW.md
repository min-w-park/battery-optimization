# M4: Event Bus Integration with NATS - Overview

## 🎯 Goal

Transform the existing synchronous microservices architecture into an **event-driven system** by integrating NATS as the event bus. This milestone retrofits the Asset Management Service (M2) and Market Data Service (M3) with event publishing capabilities, establishes a common event library, and creates event-driven communication patterns that enable asynchronous service collaboration.

**Key Deliverables**:
- Common event library (`pkg/events`) with 3 core events
- NATS publisher and subscriber adapters
- Event publishing in Asset Management and Market Data services
- Test subscriber tool for manual verification
- End-to-end event flow verification

**Target Outcomes**:
- Enable asynchronous, loosely-coupled communication between services
- Lay the foundation for real-time event-driven features (M5+)
- Achieve >80% test coverage on event library
- Successful end-to-end event flow from publisher → NATS → subscriber

---

## 🏗️ Big Picture

### Current Architecture (M3)

```
┌─────────────────────┐          ┌─────────────────────┐
│ Asset Management    │          │ Market Data         │
│ Service (M2)        │          │ Service (M3)        │
├─────────────────────┤          ├─────────────────────┤
│ REST API :8080      │          │ REST API :8081      │
│ PostgreSQL :5432    │          │ PostgreSQL :5433    │
└─────────────────────┘          └─────────────────────┘
         ↑                                ↑
         │                                │
    HTTP Request                     HTTP Request
    (Synchronous)                    (Synchronous)
```

**Characteristics**:
- **Synchronous communication**: Services respond immediately to HTTP requests
- **Tight coupling**: Clients must know service URLs and wait for responses
- **No event history**: Past events are lost (no replay capability)
- **Request-response only**: No publish-subscribe patterns

### Target Architecture (M4)

```
┌─────────────────────┐          ┌─────────────────────┐          ┌──────────────────┐
│ Asset Management    │          │ Market Data         │          │ Test Subscriber  │
│ Service (M2+Events) │          │ Service (M3+Events) │          │ Tool             │
├─────────────────────┤          ├─────────────────────┤          ├──────────────────┤
│ REST API :8080      │          │ REST API :8081      │          │ Listens to all   │
│ EventPublisher ─────┼─┐        │ EventPublisher ─────┼─┐        │ events (*.*)     │
│ PostgreSQL :5432    │ │        │ PostgreSQL :5433    │ │        │ Pretty-prints    │
└─────────────────────┘ │        └─────────────────────┘ │        └──────────────────┘
                        │                                │                 ↑
                        └────────────┐      ┌────────────┘                 │
                                     ↓      ↓                              │
                            ┌──────────────────────┐                       │
                            │   NATS Event Bus     │                       │
                            │   :4222 (client)     │                       │
                            │   :8222 (monitoring) │───────────────────────┘
                            └──────────────────────┘
                                     │
                Events Flow:         │
                                     ↓
                    battery.registered.v1 ──────> BatteryRegistered
                    market.price.updated.v1 ────> MarketPriceUpdated
                    battery.state.changed.v1 ───> BatteryStateChanged (M5)
```

**New Characteristics**:
- **Asynchronous communication**: Services publish events without waiting for subscribers
- **Loose coupling**: Publishers don't know who (if anyone) is listening
- **Event-driven**: Services react to events from other services
- **Temporal decoupling**: Publishers and subscribers don't need to be online simultaneously
- **Scalability**: Multiple subscribers can consume the same events

---

## 🏛️ Hexagonal Architecture Extended for Events

### M2/M3 Architecture (Before Events)

```
┌────────────────────────────────────────────────────────┐
│                    HTTP Adapter (Primary)              │
│              REST API ← HTTP Requests from Users       │
└────────────┬───────────────────────────────────────────┘
             │
             ↓
┌────────────────────────────────────────────────────────┐
│                   Application Core                     │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Domain Model (Battery, MarketPrice)             │  │
│  │  • Aggregates with business rules                │  │
│  │  • Value objects                                 │  │
│  │  • Domain events (not yet used)                  │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Ports (Interfaces)                              │  │
│  │  • BatteryRepository                             │  │
│  │  • MarketPriceRepository                         │  │
│  └──────────────────────────────────────────────────┘  │
└────────────┬───────────────────────────────────────────┘
             │
             ↓
┌────────────────────────────────────────────────────────┐
│            PostgreSQL Adapter (Secondary)              │
│         Database ← Persistence Implementation          │
└────────────────────────────────────────────────────────┘
```

### M4 Architecture (After Event Integration)

```
┌────────────────────────────────────────────────────────┐
│                    HTTP Adapter (Primary)              │
│              REST API ← HTTP Requests from Users       │
└────────────┬───────────────────────────────────────────┘
             │
             ↓
┌────────────────────────────────────────────────────────┐
│                   Application Core                     │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Domain Model                                    │  │
│  │  • Aggregates (Battery, MarketPrice)            │  │
│  │  • Domain Events (NEW!) ──────────────────────┐ │  │
│  │    - BatteryRegistered                        │ │  │
│  │    - MarketPriceUpdated                       │ │  │
│  │    - BatteryStateChanged (placeholder)        │ │  │
│  └──────────────────────────────────────────────┼──┘  │
│  ┌──────────────────────────────────────────────┼────┐ │
│  │  Ports (Interfaces)                          │    │ │
│  │  • BatteryRepository                         │    │ │
│  │  • MarketPriceRepository                     │    │ │
│  │  • EventPublisher (NEW!) ←───────────────────┘    │ │
│  │  • EventSubscriber (NEW!)                         │ │
│  └───────────────────────────────────────────────────┘ │
└────────────┬───────────────────┬────────────────────────┘
             │                   │
             ↓                   ↓
┌─────────────────────┐  ┌──────────────────────────────┐
│ PostgreSQL Adapter  │  │  NATS Adapter (NEW!)         │
│ (Secondary)         │  │  • Publisher (Secondary)     │
│ Database            │  │  • Subscriber (Primary)      │
└─────────────────────┘  └──────────────────────────────┘
```

**Key Additions**:
1. **Domain Events**: Immutable records of significant occurrences (BatteryRegistered, MarketPriceUpdated)
2. **EventPublisher Port**: Interface for publishing events to NATS
3. **EventSubscriber Port**: Interface for subscribing to events from NATS
4. **NATS Adapters**: Concrete implementations of publisher/subscriber interfaces
5. **Best-Effort Publishing**: Event failures don't block HTTP responses

---

## 📚 Learning Objectives

By completing M4, you will learn:

### 1. Event-Driven Architecture (EDA)
- **Publish-Subscribe Pattern**: Publishers emit events, subscribers react to them
- **Eventual Consistency**: Services synchronize state through events over time
- **Temporal Decoupling**: Publishers and subscribers operate independently
- **Event-Driven Workflows**: Choreography vs orchestration patterns
- **Trade-offs**: Complexity vs scalability, consistency vs availability

**Example Flow**:
```
User creates battery → HTTP 201 response → Event published (async) → Subscriber reacts
                     ↑                                              ↑
               Immediate response                           Eventual consistency
```

### 2. NATS Integration
- **NATS Core Concepts**: Subjects, subscriptions, pub/sub messaging
- **Subject Hierarchy**: Dot-separated subjects for wildcard subscriptions (`battery.*.v1`)
- **Connection Management**: Connect, publish, subscribe, graceful shutdown
- **Monitoring**: Health checks, server stats, connection tracking
- **Error Handling**: Connection failures, publish failures, subscriber errors

**NATS Subject Examples**:
```
battery.registered.v1          → Specific event
battery.*.v1                   → All battery events (v1)
*.*.v1                         → All events (v1)
>                              → All events (all versions)
```

### 3. Event Versioning
- **Schema Evolution**: Adding fields vs breaking changes
- **Backward Compatibility**: New fields with defaults, optional fields
- **Version Strategy**: Subject-based versioning (`battery.registered.v1`, `.v2`)
- **Migration Patterns**: Running multiple versions simultaneously
- **Version in Payload**: Include `event_version` field for validation

**Versioning Rules**:
```
✅ Backward-compatible (same version):
   - Add new optional field
   - Make required field optional
   - Add new event types

❌ Breaking changes (new version):
   - Remove field
   - Rename field
   - Change field type
   - Change field semantics
```

### 4. Publisher/Subscriber Pattern
- **Decoupling**: Publishers don't know who subscribes
- **Scalability**: Multiple subscribers can consume same events
- **Flexibility**: Add/remove subscribers without changing publishers
- **Fan-out**: One event → many subscribers
- **Fire-and-forget**: Publishers don't wait for subscriber acknowledgment

### 5. Testing Asynchronous Systems
- **Event Flow Verification**: Publish event → verify subscriber receives it
- **Test Subscribers**: Simple tools to inspect event streams
- **Async Assertions**: Wait for eventual consistency
- **Event Serialization**: Test JSON marshaling/unmarshaling
- **Integration Testing**: End-to-end event flows across services

---

## 📊 Comparison: M3 vs M4

| Aspect | M3 (Market Data Service) | M4 (Event Bus Integration) |
|--------|--------------------------|----------------------------|
| **Primary Goal** | New microservice with time-series data | Infrastructure layer for async communication |
| **Communication** | Synchronous REST API | Asynchronous event publishing |
| **New Service?** | Yes (market-data on :8081) | No (retrofits existing services) |
| **New Database?** | Yes (market-db on :5433) | No (events are ephemeral) |
| **Architecture Pattern** | CRUD with REST | Pub/Sub with NATS |
| **Primary Artifact** | MarketPrice aggregate | Event schemas + pub/sub adapters |
| **Testing Focus** | HTTP integration tests | Event flow verification |
| **Deployment** | Docker service on :8081 | Shared library + NATS connection |
| **Data Persistence** | PostgreSQL (time-series) | No persistence (in-memory streams) |
| **Inter-Service Coupling** | Tight (synchronous calls) | Loose (async events) |
| **Failure Mode** | Request fails if service down | Events buffered/redelivered |
| **Key Learning** | Time-series data, interval alignment | Event-driven patterns, NATS pub/sub |

---

## 🛠️ Development Phases

### Pre-Phase: Update M3-CHECKLIST.md (5-10 minutes)
Update M3-CHECKLIST.md with completion status before starting M4.

### Phase 0: Create M4 Milestone Documentation (1-1.5 hours)
- Create M4-OVERVIEW.md (this document)
- Create M4-DOMAIN-SPEC.md (event schemas, versioning, interfaces)
- Create M4-API-SPEC.md (NATS pub/sub API, subject naming)
- Create M4-CHECKLIST.md (implementation guide)

### Phase 1: Common Event Library (2-3 hours)
- Create `pkg/events/` package
- Define event structs (BatteryRegistered, MarketPriceUpdated, BatteryStateChanged)
- Define EventPublisher and EventSubscriber interfaces
- Write comprehensive tests (>80% coverage)

### Phase 2: NATS Publisher Adapter (2-3 hours)
- Implement NATSPublisher struct
- Connect to NATS server
- Implement Publish() method with JSON serialization
- Write integration tests with real NATS

### Phase 3: NATS Subscriber Adapter (2-3 hours)
- Implement NATSSubscriber struct
- Implement Subscribe() method with wildcard support
- Implement EventHandler callback pattern
- Write integration tests

### Phase 4: Asset Service Integration (2-3 hours)
- Add EventPublisher port to Asset Management Service
- Publish BatteryRegistered after battery creation
- Update main.go with NATS connection
- Update docker-compose.yml with NATS_URL

### Phase 5: Market Service Integration (2-3 hours)
- Add EventPublisher port to Market Data Service
- Publish MarketPriceUpdated after price creation
- Update main.go with NATS connection
- Update docker-compose.yml with NATS_URL

### Phase 6: Test Subscriber Tool (1-2 hours)
- Create `tools/event-subscriber/` CLI tool
- Subscribe to all events (`>`)
- Pretty-print JSON events to console
- Implement graceful shutdown

### Phase 7: Integration Testing (2-3 hours)
- End-to-end test: Create battery → verify BatteryRegistered event
- End-to-end test: Create price → verify MarketPriceUpdated event
- Test multiple events
- Verify NATS monitoring endpoints

### Phase 8: Polish & Documentation (1-2 hours)
- Run go fmt, go vet
- Verify test coverage (>80%)
- Create pkg/events/README.md
- Update PLANNING.md and CLAUDE.md
- Git commit and tag

**Total Estimated Time**: 13-18 hours

---

## 🚫 Out of Scope for M4

To keep M4 focused and achievable, the following are explicitly **out of scope**:

### Event Persistence and Replay
- **Not Included**: Storing events in database for replay
- **Rationale**: M4 focuses on pub/sub infrastructure, not event sourcing
- **Future**: Event sourcing in M7+

### Complex Event Subscribers
- **Not Included**: Subscribers with business logic (e.g., bidding logic reacting to prices)
- **Rationale**: M4 establishes infrastructure only
- **Future**: Complex subscribers in M6 (Bidding Service)

### Event Sourcing and CQRS
- **Not Included**: Using events as source of truth, separate read/write models
- **Rationale**: Pattern overkill for current scope
- **Future**: Consider for specific aggregates in M7+

### Dead Letter Queues
- **Not Included**: Handling failed event processing with DLQ
- **Rationale**: M4 uses simple best-effort delivery
- **Future**: Production-grade error handling in M7+

### Event Schema Registry
- **Not Included**: Centralized schema management and validation
- **Rationale**: Simple version-in-subject approach sufficient
- **Future**: Schema registry if many event types emerge

### Multiple NATS Clusters
- **Not Included**: Multi-region, multi-cluster NATS deployments
- **Rationale**: Single NATS instance sufficient for learning
- **Future**: Production deployment patterns in M7+

---

## ✅ Definition of Done

M4 is complete when **all** of the following are true:

### Functional Requirements
- [ ] `pkg/events` library created with 3 event types (BatteryRegistered, MarketPriceUpdated, BatteryStateChanged)
- [ ] EventPublisher and EventSubscriber interfaces defined
- [ ] NATSPublisher implemented and tested
- [ ] NATSSubscriber implemented and tested
- [ ] Asset Management Service publishes BatteryRegistered events
- [ ] Market Data Service publishes MarketPriceUpdated events
- [ ] Test subscriber tool receives all events
- [ ] NATS monitoring shows 3 connections (2 publishers + 1 subscriber)

### Testing Requirements
- [ ] pkg/events test coverage >80%
- [ ] Event serialization/deserialization tests pass
- [ ] Publisher integration tests pass (with real NATS)
- [ ] Subscriber integration tests pass (with wildcard subscriptions)
- [ ] End-to-end test: Create battery → BatteryRegistered received
- [ ] End-to-end test: Create price → MarketPriceUpdated received

### Architecture Requirements
- [ ] Services remain loosely coupled via events
- [ ] REST APIs continue working (synchronous communication preserved)
- [ ] Event versioning infrastructure in place (subject + payload version)
- [ ] Graceful degradation when NATS unavailable (publish failures logged, not fatal)
- [ ] Ready for M5: Can easily add BatteryStateChanged event

### Documentation Requirements
- [ ] All 4 M4 milestone docs created (OVERVIEW, DOMAIN-SPEC, API-SPEC, CHECKLIST)
- [ ] pkg/events/README.md with usage guide
- [ ] PLANNING.md updated with M4 completion
- [ ] CLAUDE.md updated with event publishing section
- [ ] Git committed and tagged (m4-complete)

---

## 🎯 Success Criteria

You successfully completed M4 if you can:

1. **Demonstrate Event Flow**:
   ```bash
   # Terminal 1: Start test subscriber
   go run tools/event-subscriber/main.go

   # Terminal 2: Create battery
   curl -X POST http://localhost:8080/api/v1/batteries -d '{ ... }'

   # Terminal 1: See BatteryRegistered event logged
   ```

2. **Explain Event Decoupling**:
   - Why publishers don't know about subscribers
   - How new subscribers can be added without changing publishers
   - Trade-offs: eventual consistency vs immediate consistency

3. **Describe Event Versioning**:
   - How to add a new field to BatteryRegistered (backward-compatible)
   - When to create `.v2` version (breaking changes)
   - How subscribers handle multiple versions

4. **Monitor NATS**:
   ```bash
   # Check health
   curl http://localhost:8222/healthz

   # View connections
   curl http://localhost:8222/connz | jq
   ```

5. **Add New Event (Future)**:
   - Add BatteryStateChanged to pkg/events
   - Publish from Telemetry Service (M5)
   - Subscribe in Bidding Service (M6)

---

## 🔄 Relationship to Other Milestones

### M0 (Project Setup)
- **Provides**: NATS infrastructure (docker-compose.yml)
- **M4 Uses**: Existing NATS on ports 4222, 8222

### M1 (Event Storming)
- **Provides**: 25 domain events documented
- **M4 Implements**: 3 events (BatteryRegistered, MarketPriceUpdated, BatteryStateChanged placeholder)

### M2 (Asset Management Service)
- **Provides**: Battery aggregate, REST API
- **M4 Extends**: Add event publishing after battery creation

### M3 (Market Data Service)
- **Provides**: MarketPrice aggregate, REST API
- **M4 Extends**: Add event publishing after price creation

### M5 (Telemetry + Device Interface) - NEXT
- **Will Use**: pkg/events library and NATS publisher
- **Will Publish**: BatteryStateChanged every 1 second (high-frequency events)
- **Depends On**: M4 event infrastructure

### M6 (Bidding Service)
- **Will Use**: NATS subscriber to react to events
- **Will Subscribe**: BatteryRegistered, MarketPriceUpdated, BatteryStateChanged
- **Depends On**: M4 event infrastructure

---

## 🚀 Ready to Start?

Once you've reviewed this overview:

1. Read [M4-DOMAIN-SPEC.md](M4-DOMAIN-SPEC.md) for event schemas and validation
2. Read [M4-API-SPEC.md](M4-API-SPEC.md) for NATS pub/sub API details
3. Follow [M4-CHECKLIST.md](M4-CHECKLIST.md) step-by-step

**Next Step**: Phase 1 - Common Event Library (`pkg/events`)

---

**Last Updated**: 2025-12-30
**Status**: 🚧 In Progress
**Target Completion**: 13-18 hours
