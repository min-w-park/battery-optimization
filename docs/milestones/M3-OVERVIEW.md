# M3: Market Data Service - Overview

## 🎯 Goal

Build the second microservice using DDD, TDD, and event-driven patterns. This service manages AEMO price forecasts and market data for the Australian National Electricity Market (NEM) - providing critical pricing information for battery optimization decisions.

## 📊 Big Picture

```
┌─────────────────────────────────────────────┐
│  AEMO Market Operator / Price Simulator     │
└───────────────┬─────────────────────────────┘
                │
                ↓ Price Forecast Data
┌─────────────────────────────────────────────┐
│  Market Data Service                         │
│  ┌───────────────────────────────────────┐  │
│  │  HTTP Handler (Adapter)               │  │
│  │  - POST /prices                       │  │
│  │  - GET /prices/:id                    │  │
│  │  - GET /prices?from=X&to=Y            │  │
│  └───────────────┬───────────────────────┘  │
│                  ↓                           │
│  ┌───────────────────────────────────────┐  │
│  │  Domain Layer (Pure Business Logic)  │  │
│  │  - MarketPrice Aggregate              │  │
│  │  - Validation Rules (8 rules)         │  │
│  │  - Interval Alignment Logic           │  │
│  └───────────────┬───────────────────────┘  │
│                  ↓                           │
│  ┌───────────────────────────────────────┐  │
│  │  Repository (Port/Interface)          │  │
│  └───────────────┬───────────────────────┘  │
│                  ↓                           │
│  ┌───────────────────────────────────────┐  │
│  │  PostgreSQL Adapter                   │  │
│  │  (Time-series optimized queries)      │  │
│  └───────────────┬───────────────────────┘  │
└──────────────────┼───────────────────────────┘
                   ↓
         ┌─────────────────┐
         │  PostgreSQL DB  │
         │  (market_db)    │
         │  Port: 5433     │
         └─────────────────┘
                   ↑
         (Future: NATS Event Publishing)
```

## 🏗️ Hexagonal Architecture (Ports & Adapters)

```
Core (Domain):
- MarketPrice aggregate
- Interval alignment validation
- Time-series business rules
- No external dependencies

Ports (Interfaces):
- MarketPriceRepository interface (with time-range queries)
- (Future) EventPublisher interface

Adapters (Infrastructure):
- HTTP handlers (primary)
- PostgreSQL repository (secondary, time-series optimized)
- NATS publisher (future, secondary)
```

**Key Principle**: Domain never depends on infrastructure. Infrastructure depends on domain.

## 📁 Project Structure

```
services/market-data/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point, dependency injection
│
├── internal/
│   ├── domain/                     # Core business logic (no external deps)
│   │   ├── price.go               # MarketPrice aggregate
│   │   ├── price_test.go          # Domain tests (TDD)
│   │   └── errors.go              # Domain errors
│   │
│   ├── ports/                      # Interfaces defined by domain
│   │   └── repository.go          # Repository interface
│   │
│   └── adapters/
│       ├── http/                   # Primary adapter (drives domain)
│       │   ├── handler.go         # HTTP handlers
│       │   ├── handler_test.go    # Handler tests
│       │   ├── routes.go          # Route definitions
│       │   └── dto.go             # Data transfer objects
│       │
│       └── postgres/               # Secondary adapter (driven by domain)
│           ├── repository.go      # Repository implementation
│           ├── repository_test.go # Repository tests
│           └── migrations/        # SQL migrations
│               ├── 000001_create_market_prices.up.sql
│               └── 000001_create_market_prices.down.sql
│
├── Dockerfile
├── go.mod
└── go.sum
```

## 🎓 Learning Objectives

By completing M3, you will:

1. **Time-Series Data Modeling**
   - Handle time-based intervals (5-minute, 30-minute)
   - Implement interval alignment validation
   - Optimize queries for time ranges
   - Design immutable time-series records

2. **DB-per-Service Pattern**
   - Second independent PostgreSQL database (market-db)
   - Verify service isolation (no cross-database queries)
   - Understand bounded context boundaries

3. **TDD on Second Service**
   - Replicate M2 TDD patterns
   - Apply Red → Green → Refactor workflow
   - Achieve >80% test coverage again

4. **Advanced Query Patterns**
   - Time-range filtering (from/to timestamps)
   - Optional filters (region, interval type)
   - Pagination with time-series data
   - Index optimization for timestamps

5. **Domain Validation Complexity**
   - Timestamp boundary validation
   - Interval alignment rules (5-min and 30-min boundaries)
   - Duplicate detection across dimensions
   - Time-based business rules

## 🔄 Development Flow (TDD)

```
For each feature:

1. Write failing test (RED)
   → Test describes what should happen

2. Write minimal code to pass (GREEN)
   → Just enough to make test pass

3. Refactor (REFACTOR)
   → Clean up, extract functions

4. Repeat

Order:
Domain → Repository → HTTP Handler
(Inside to outside, same as M2)
```

## 📋 Milestones

### Milestone 3.1: Domain Model (Est: 2-3 hours)
- MarketPrice aggregate with 8 validation rules
- Interval alignment logic
- Domain tests passing
- No external dependencies

### Milestone 3.2: Repository Layer (Est: 2-3 hours)
- Repository interface with time-range queries
- PostgreSQL implementation
- Time-series indexes
- Repository tests passing

### Milestone 3.3: HTTP API (Est: 2-3 hours)
- REST endpoints (time-range query support)
- Handler tests
- DTOs for request/response
- ISO 8601 timestamp handling

### Milestone 3.4: Integration (Est: 2-3 hours)
- Docker setup
- Database migrations
- End-to-end manual testing
- Documentation
- Verify 2 services running independently

**Total Estimated Time: 11-13 hours**

## 🚫 Out of Scope (for M3)

- Event publishing (deferred to M4)
- ML-based price forecasting
- Update/Delete operations (only Create/Read)
- Complex analytics queries
- Price change detection (conditional events in M4)
- AEMO API integration (simulation only)

## ✅ Definition of Done

- [ ] All unit tests pass (>80% coverage)
- [ ] Service runs in Docker on port 8081
- [ ] Can create market price via API
- [ ] Can query prices by time range
- [ ] Code follows Go conventions
- [ ] README updated with API docs
- [ ] Manual testing completed
- [ ] Two services running independently (Asset + Market)

## 🔗 Related Documents

**M3 Specifications**:
- [Domain Specification](./M3-DOMAIN-SPEC.md) - MarketPrice aggregate details
- [API Specification](./M3-API-SPEC.md) - REST API contract
- [Checklist](./M3-CHECKLIST.md) - Step-by-step tasks

**Development Guides** (READ THESE FIRST):
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development philosophy and TDD workflow (Kent Beck style)
- [EVENTS.md](../../EVENTS.md) - Market data events (section 2)

## 🎯 Success Criteria

**You'll know M3 is complete when:**

1. You can `curl` the API and create a market price
2. The price is stored in market-db (PostgreSQL)
3. You can query prices by time range (from/to)
4. All tests are green
5. Two services run independently (asset-management + market-data)
6. No cross-database dependencies

**Most importantly**: You should be able to explain:
- Why time-series data requires different modeling
- How interval alignment validation works
- Why DB-per-service pattern matters
- How you'd extend it for event publishing (M4)

## 💡 Key Insights to Remember

1. **Time-series is different**: Immutable, time-range queries, interval alignment
2. **DB-per-service isolation**: Each service owns its data completely
3. **Duplicate prevention**: Unique constraint on (region, interval_type, interval_start)
4. **Indexes matter**: Time-range queries need proper indexing

## 🔍 Key Differences from M2

| Aspect | M2 (Asset Management) | M3 (Market Data) |
|--------|----------------------|------------------|
| Domain Model | Entity (Battery) | Time-series (MarketPrice) |
| Primary Key | UUID only | UUID + unique constraint on interval |
| Queries | Simple filters (location, status) | Time-range (from/to timestamps) |
| Updates | Future: status changes | Immutable (no updates) |
| Validation | 10 rules (capacity, power, etc.) | 8 rules (interval alignment, time constraints) |
| Database Port | 5432 (asset-db) | 5433 (market-db) |
| Service Port | 8080 | 8081 |

---

**Next Steps**: Read the detailed specs, then start with the checklist!
