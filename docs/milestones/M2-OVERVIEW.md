# M2: Asset Management Service - Overview

## 🎯 Goal

Build the first microservice using DDD, TDD, and event-driven patterns. This service manages battery specifications and constraints - the foundation for all other services.

## 📊 Big Picture

```
┌─────────────────────────────────────────────┐
│  Operator (Web UI / API Client)             │
└───────────────┬─────────────────────────────┘
                │
                ↓ POST /batteries (register)
┌─────────────────────────────────────────────┐
│  Asset Management Service                    │
│  ┌───────────────────────────────────────┐  │
│  │  HTTP Handler (Adapter)               │  │
│  │  - POST /batteries                    │  │
│  │  - GET /batteries/:id                 │  │
│  │  - GET /batteries                     │  │
│  └───────────────┬───────────────────────┘  │
│                  ↓                           │
│  ┌───────────────────────────────────────┐  │
│  │  Domain Layer (Pure Business Logic)  │  │
│  │  - Battery Aggregate                  │  │
│  │  - Validation Rules                   │  │
│  │  - Domain Events                      │  │
│  └───────────────┬───────────────────────┘  │
│                  ↓                           │
│  ┌───────────────────────────────────────┐  │
│  │  Repository (Port/Interface)          │  │
│  └───────────────┬───────────────────────┘  │
│                  ↓                           │
│  ┌───────────────────────────────────────┐  │
│  │  PostgreSQL Adapter                   │  │
│  └───────────────┬───────────────────────┘  │
└──────────────────┼───────────────────────────┘
                   ↓
         ┌─────────────────┐
         │  PostgreSQL DB  │
         │  (asset_db)     │
         └─────────────────┘
                   ↑
         (Future: NATS Event Publishing)
```

## 🏗️ Hexagonal Architecture (Ports & Adapters)

```
Core (Domain):
- Battery aggregate
- Business rules
- No external dependencies

Ports (Interfaces):
- BatteryRepository interface
- (Future) EventPublisher interface

Adapters (Infrastructure):
- HTTP handlers (primary)
- PostgreSQL repository (secondary)
- NATS publisher (future, secondary)
```

**Key Principle**: Domain never depends on infrastructure. Infrastructure depends on domain.

## 📁 Project Structure

```
services/asset-management/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point, dependency injection
│
├── internal/
│   ├── domain/                     # Core business logic (no external deps)
│   │   ├── battery.go             # Battery aggregate
│   │   ├── battery_test.go        # Domain tests (TDD)
│   │   ├── errors.go              # Domain errors
│   │   └── events.go              # Domain events (future)
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
│               └── 001_create_batteries.sql
│
├── Dockerfile
├── go.mod
└── go.sum
```

## 🎓 Learning Objectives

By completing M2, you will:

1. **DDD in Practice**
   - Define a bounded context (Asset Management)
   - Create an aggregate (Battery)
   - Implement domain validation
   - Keep domain pure (no infrastructure leaking in)

2. **TDD Workflow**
   - Write test first, then implementation
   - Use table-driven tests (Go idiom)
   - Achieve >80% test coverage

3. **Hexagonal Architecture**
   - Define ports (interfaces) in domain
   - Implement adapters (HTTP, PostgreSQL)
   - See how adapters can be swapped

4. **Repository Pattern**
   - Abstract data access
   - Test without real database
   - Use transactions properly

5. **Clean Code**
   - Single Responsibility Principle
   - Dependency Inversion Principle
   - Clear separation of concerns

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
(Inside to outside)
```

## 📋 Milestones

### Milestone 2.1: Domain Model (Est: 2-3 hours)
- Battery aggregate with validation
- Domain tests passing
- No external dependencies

### Milestone 2.2: Repository Layer (Est: 2-3 hours)
- Repository interface defined
- PostgreSQL implementation
- Repository tests passing

### Milestone 2.3: HTTP API (Est: 2-3 hours)
- REST endpoints
- Handler tests
- DTOs for request/response

### Milestone 2.4: Integration (Est: 2-3 hours)
- Docker setup
- Database migrations
- End-to-end manual testing
- Documentation

**Total Estimated Time: 10-12 hours**

## 🚫 Out of Scope (for M2)

- Event publishing (deferred to M4)
- Authentication/authorization
- Update/Delete operations (only Create/Read)
- Complex queries
- Caching

## ✅ Definition of Done

- [ ] All unit tests pass (>80% coverage)
- [ ] Service runs in Docker
- [ ] Can register battery via API
- [ ] Can retrieve battery by ID
- [ ] Code follows Go conventions
- [ ] README updated with API docs
- [ ] Manual testing completed

## 🔗 Related Documents

**M2 Specifications**:
- [Domain Specification](./M2-DOMAIN-SPEC.md) - Battery aggregate details
- [API Specification](./M2-API-SPEC.md) - REST API contract
- [Checklist](./M2-CHECKLIST.md) - Step-by-step tasks

**Development Guides** (READ THESE FIRST):
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development philosophy and TDD workflow (Kent Beck style)
- [Battery Domain Skill](../../.claude/skills/BATTERY-DOMAIN-SKILL.md) - Domain concepts

## 🎯 Success Criteria

**You'll know M2 is complete when:**

1. You can `curl` the API and register a battery
2. The battery is stored in PostgreSQL
3. You can retrieve it by ID
4. All tests are green
5. You understand every line of code

**Most importantly**: You should be able to explain:
- Why you structured it this way
- What patterns you used
- How you'd extend it for new features

## 💡 Key Insights to Remember

1. **Domain is king**: Everything else serves the domain
2. **Interfaces at boundaries**: Makes testing and swapping easy
3. **TDD builds confidence**: Tests document behavior
4. **Simple first**: Don't over-engineer, add complexity when needed

---

**Next Steps**: Read the detailed specs, then start with the checklist!
