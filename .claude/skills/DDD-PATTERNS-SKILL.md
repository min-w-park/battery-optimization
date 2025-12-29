# Domain-Driven Design (DDD) Patterns Skill

## When to use this skill
Use this skill when designing or implementing domain models, bounded contexts, aggregates, repositories, or any domain-driven design patterns.

---

## Core DDD Concepts

### Bounded Context
A **bounded context** is an explicit boundary within which a domain model exists.

**In this project:**
```
Bounded Contexts:
1. Asset Management - Battery specifications and lifecycle
2. Market Data - Pricing and forecasts
3. Telemetry - Real-time battery state
4. Trading - Bidding and optimization decisions
5. Economics - Financial calculations

Each context has its own:
- Domain model
- Ubiquitous language
- Database
- Service boundaries
```

**Key principle:** Models don't leak across contexts

```go
// GOOD - separate models per context
// Asset Management context
type Battery struct {
    ID       string
    Capacity float64
    Status   BatteryStatus
}

// Telemetry context (different model!)
type BatteryState struct {
    BatteryID string
    SoC       float64
    Power     float64
    Timestamp time.Time
}

// They share BatteryID but are independent models
```

---

### Aggregates

An **aggregate** is a cluster of domain objects treated as a unit. One object is the **aggregate root**.

**Rules:**
1. External objects can only reference the aggregate root
2. Invariants are enforced at aggregate boundaries
3. Transactions should not span aggregates

**Battery Aggregate (example):**
```go
// Battery is the aggregate root
type Battery struct {
    ID          string          // Identity
    Capacity    float64
    Constraints Constraints     // Part of aggregate
    Status      BatteryStatus
    // ... other fields
}

// Constraints is NOT a separate aggregate
// It's a value object within Battery aggregate
type Constraints struct {
    WarrantyEOL float64
    MaxCycles   int
    TempMin     float64
    TempMax     float64
}

// External code interacts with Battery, not Constraints directly
battery.Constraints.WarrantyEOL  // OK - through root
repo.UpdateConstraints(constraints)  // ❌ BAD - bypasses root
```

**Identify aggregate boundaries:**
```
Ask: "What must be consistent together?"

Battery + Constraints → Same aggregate (consistency needed)
Battery + BatteryState → Different aggregates (eventual consistency OK)
Battery + MarketPrice → Different aggregates (independent)
```

---

### Entities vs Value Objects

**Entity** = Identity matters
```go
// Battery is an Entity
// Two batteries with same specs are DIFFERENT batteries
type Battery struct {
    ID       string  // ← Identity
    Capacity float64
}

b1 := &Battery{ID: "123", Capacity: 200}
b2 := &Battery{ID: "456", Capacity: 200}
// b1 != b2 (different IDs = different entities)
```

**Value Object** = Equality by value
```go
// Constraints is a Value Object
// Two constraints with same values are SAME constraints
type Constraints struct {
    WarrantyEOL float64
    MaxCycles   int
}

c1 := Constraints{WarrantyEOL: 0.7, MaxCycles: 10000}
c2 := Constraints{WarrantyEOL: 0.7, MaxCycles: 10000}
// c1 == c2 (same values = equal)
```

**Value Object characteristics:**
- Immutable (create new instead of modifying)
- No identity
- Can be shared
- Equality by value

```go
// GOOD - value object immutability
type Money struct {
    Amount   float64
    Currency string
}

func (m Money) Add(other Money) Money {
    if m.Currency != other.Currency {
        panic("currency mismatch")
    }
    return Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }
}

// Usage
m1 := Money{Amount: 100, Currency: "USD"}
m2 := m1.Add(Money{Amount: 50, Currency: "USD"})
// m1 unchanged (immutable), m2 is new value
```

---

### Domain Services

**When to use:** Logic that doesn't naturally belong to an entity.

```go
// Domain Service - operates on multiple aggregates
type ConflictResolver struct{}

func (r *ConflictResolver) Resolve(
    currentOperation Operation,
    newRequest Request,
    contracts []FcasContract,
) (Resolution, error) {
    // Business logic that involves multiple aggregates
    // Doesn't belong to any single aggregate
}

// NOT a domain service (belongs to Battery)
func ValidateBattery(battery *Battery) error {  // ❌ Should be battery.Validate()
```

**Domain Service characteristics:**
- Stateless
- Operates on domain objects
- Contains business logic
- Named with verbs (e.g., ConflictResolver, RevenueCalculator)

---

### Repositories

**Repository** = Persistence abstraction for aggregates

```go
// Repository interface (port) - in domain layer
type BatteryRepository interface {
    // Only operates on aggregate roots
    Create(ctx context.Context, battery *Battery) error
    FindByID(ctx context.Context, id string) (*Battery, error)
    Update(ctx context.Context, battery *Battery) error
    Delete(ctx context.Context, id string) error
}

// ✅ GOOD - collection-oriented interface
// Repository feels like an in-memory collection

// ❌ BAD - database-oriented interface
type BatteryRepository interface {
    SaveToDatabase(battery *Battery) error
    ExecuteQuery(sql string) (*Battery, error)
}
```

**Repository rules:**
1. One repository per aggregate root
2. Repository returns domain objects, not DTOs
3. Don't expose database details
4. Repository belongs to domain (interface in domain layer)

```go
// GOOD - domain-centric
battery, err := repo.FindByID(ctx, "123")
if errors.Is(err, domain.ErrBatteryNotFound) {
    // Handle domain error
}

// BAD - database-centric
battery, err := repo.FindByID(ctx, "123")
if err == sql.ErrNoRows {  // ❌ Leaking database details
    // Domain shouldn't know about sql.ErrNoRows
}
```

---

### Domain Events

**Domain Event** = Something that happened in the domain

```go
// Event - past tense naming
type BatteryRegistered struct {
    BatteryID    string
    Capacity     float64
    RegisteredAt time.Time
}

type BatteryStatusChanged struct {
    BatteryID  string
    OldStatus  BatteryStatus
    NewStatus  BatteryStatus
    ChangedAt  time.Time
}

// Aggregates produce events
func (b *Battery) Register() (BatteryRegistered, error) {
    b.Status = StatusRegistered
    b.RegisteredAt = time.Now()
    
    return BatteryRegistered{
        BatteryID:    b.ID,
        Capacity:     b.Capacity,
        RegisteredAt: b.RegisteredAt,
    }, nil
}
```

**Event characteristics:**
- Past tense
- Immutable
- Contains relevant data
- Timestamp included
- Domain concept (not technical event)

---

## Hexagonal Architecture (Ports & Adapters)

**Layers:**
```
┌─────────────────────────────────────────┐
│  Adapters (Infrastructure)              │
│  - HTTP handlers (Primary)              │
│  - PostgreSQL repo (Secondary)          │
│  - NATS publisher (Secondary)           │
└───────────────┬─────────────────────────┘
                │ depends on
┌───────────────▼─────────────────────────┐
│  Ports (Interfaces)                     │
│  - Repository interface                 │
│  - EventPublisher interface             │
└───────────────┬─────────────────────────┘
                │ defined by
┌───────────────▼─────────────────────────┐
│  Domain (Core Business Logic)           │
│  - Battery aggregate                    │
│  - Validation rules                     │
│  - Business invariants                  │
└─────────────────────────────────────────┘
```

**Dependency rule:** Domain depends on NOTHING. Everything depends on domain.

```go
// GOOD - domain defines interface
// internal/ports/repository.go
package ports

type BatteryRepository interface {
    Create(ctx context.Context, battery *domain.Battery) error
}

// Adapter implements interface
// internal/adapters/postgres/repository.go
package postgres

type PostgresRepository struct {
    db *sql.DB
}

func (r *PostgresRepository) Create(ctx context.Context, battery *domain.Battery) error {
    // PostgreSQL-specific implementation
}
```

**Why?** Domain can be tested without database, HTTP, or any infrastructure.

---

## Anti-Corruption Layer (ACL)

**Purpose:** Protect your domain from external models

```go
// External system (AEMO API) has its own model
type AEMOPrice struct {
    RRP              float64
    RegionID         string
    IntervalDateTime string  // Different format
    // ... many other fields we don't need
}

// Anti-Corruption Layer - translates to our domain
type AEMOAdapter struct{}

func (a *AEMOAdapter) FetchPrice() (domain.Price, error) {
    aemoPrice := fetchFromAEMO()  // External call
    
    // Translate to our domain model
    timestamp, _ := time.Parse(time.RFC3339, aemoPrice.IntervalDateTime)
    
    return domain.Price{
        Amount:    aemoPrice.RRP,
        Region:    domain.Region(aemoPrice.RegionID),
        Timestamp: timestamp,
    }, nil
}

// Our domain never sees AEMO's model
```

**Key principle:** External changes don't force domain changes

---

## Ubiquitous Language

**Use domain terms everywhere** - code, docs, conversations

```go
// GOOD - uses domain language
type Battery struct {
    Capacity     float64  // MWh - domain term
    RampRate     float64  // MW/min - domain term
    WarrantyEOL  float64  // End of Life - domain term
}

func (b *Battery) CanParticipateInFCAS() bool {
    return b.RampRate >= 1.0  // FCAS requirement
}

// BAD - technical jargon, not domain language
type Battery struct {
    StorageSize   float64  // ❌ Use domain term "Capacity"
    PowerChangeSpeed float64  // ❌ Use "RampRate"
}

func (b *Battery) CheckMarketEligibility() bool {  // ❌ Use "FCAS" (domain term)
```

**Benefits:**
- Developers and domain experts speak same language
- Code documents domain knowledge
- Reduces translation errors

---

## Patterns for This Project

### Pattern 1: Domain-First Development

**Order:**
1. Model domain (aggregates, value objects)
2. Define ports (interfaces)
3. Implement adapters

```go
// Step 1: Domain
type Battery struct { ... }
func (b *Battery) Validate() error { ... }

// Step 2: Port
type BatteryRepository interface { ... }

// Step 3: Adapter
type PostgresRepository struct { ... }
func (r *PostgresRepository) Create(...) { ... }
```

### Pattern 2: Separate Read/Write Models (CQRS-lite)

```go
// Write model (aggregate) - enforces invariants
type Battery struct {
    ID       string
    Capacity float64
    // ... full model
}

// Read model (projection) - optimized for queries
type BatteryListItem struct {
    ID           string
    Capacity     float64
    Location     string
    Status       string
    // Only fields needed for list view
}

// Repository has both
type BatteryRepository interface {
    Create(ctx context.Context, battery *Battery) error  // Write
    List(ctx context.Context) ([]BatteryListItem, error) // Read
}
```

### Pattern 3: Factory Pattern

```go
// Factory encapsulates complex creation
func NewBattery(
    capacity float64,
    maxPower float64,
    constraints Constraints,
) (*Battery, error) {
    battery := &Battery{
        ID:          uuid.New().String(),
        Capacity:    capacity,
        MaxPower:    maxPower,
        Constraints: constraints,
        Status:      StatusRegistered,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := battery.Validate(); err != nil {
        return nil, err
    }
    
    return battery, nil
}

// Callers don't need to know creation details
battery, err := NewBattery(200, 100, constraints)
```

### Pattern 4: Specification Pattern (for complex queries)

```go
// Specification encapsulates query logic
type BatterySpecification interface {
    IsSatisfiedBy(battery *Battery) bool
}

type OperationalInRegion struct {
    Region string
}

func (s OperationalInRegion) IsSatisfiedBy(b *Battery) bool {
    return b.Status == StatusOperational && b.Location == s.Region
}

// Usage
spec := OperationalInRegion{Region: "NSW"}
for _, battery := range batteries {
    if spec.IsSatisfiedBy(battery) {
        // Use battery
    }
}
```

---

## Testing DDD Code

### Test Domain in Isolation

```go
// GOOD - pure domain test, no infrastructure
func TestBattery_Validate_NegativeCapacity_ReturnsError(t *testing.T) {
    battery := &Battery{Capacity: -100}
    
    err := battery.Validate()
    
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}
// No database, no HTTP, just domain logic
```

### Test Repositories with Real DB (Integration Test)

```go
func TestPostgresRepository_Create_StoresBattery(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupDB(db)
    
    repo := NewPostgresRepository(db)
    battery := createTestBattery()
    
    err := repo.Create(context.Background(), battery)
    
    assert.NoError(t, err)
    // Verify in database
}
```

### Mock Repositories for Application Tests

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*Battery, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Battery), args.Error(1)
}

// Test application logic without database
func TestService_ProcessBattery(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRepo.On("FindByID", mock.Anything, "123").
        Return(testBattery, nil)
    
    service := NewService(mockRepo)
    // Test service logic
}
```

---

## Common DDD Mistakes

### ❌ Mistake 1: Anemic Domain Model

```go
// BAD - no behavior, just data
type Battery struct {
    Capacity float64
    MaxPower float64
}

// Service does all the work
type BatteryService struct{}
func (s *BatteryService) ValidateBattery(b *Battery) error {
    if b.Capacity <= 0 { return error }
    if b.MaxPower > b.Capacity { return error }
}
```

**✅ Fix: Rich domain model**
```go
// GOOD - behavior in domain
type Battery struct {
    Capacity float64
    MaxPower float64
}

func (b *Battery) Validate() error {
    if b.Capacity <= 0 { return error }
    if b.MaxPower > b.Capacity { return error }
    return nil
}
```

### ❌ Mistake 2: Domain Depends on Infrastructure

```go
// BAD - domain imports database
package domain

import "database/sql"  // ❌ Domain shouldn't know about SQL

type Battery struct {
    db *sql.DB  // ❌ Infrastructure in domain
}
```

**✅ Fix: Dependency inversion**
```go
// GOOD - domain defines interface
package domain

type BatteryRepository interface {  // Domain defines what it needs
    Create(battery *Battery) error
}

// Infrastructure implements interface
package postgres

import "database/sql"

type PostgresRepository struct {
    db *sql.DB  // ✅ Infrastructure details here
}

func (r *PostgresRepository) Create(battery *domain.Battery) error {
    // SQL details here, domain doesn't care
}
```

### ❌ Mistake 3: Leaking Domain Events to Infrastructure

```go
// BAD - publishing in domain
package domain

import "github.com/nats-io/nats.go"  // ❌ Domain knows about NATS

func (b *Battery) Register() error {
    b.Status = StatusRegistered
    nats.Publish("battery.registered", b)  // ❌ Infrastructure in domain
}
```

**✅ Fix: Return events, let infrastructure publish**
```go
// GOOD - domain returns event
package domain

func (b *Battery) Register() (BatteryRegistered, error) {
    b.Status = StatusRegistered
    return BatteryRegistered{BatteryID: b.ID}, nil
}

// Infrastructure publishes
package app

func (s *Service) RegisterBattery(battery *Battery) error {
    event, err := battery.Register()  // Get event from domain
    if err != nil { return err }
    
    s.repo.Save(battery)
    s.publisher.Publish(event)  // Infrastructure publishes
    return nil
}
```

---

## DDD Quick Reference

**Entities:** Identity matters
**Value Objects:** Equality by value, immutable
**Aggregates:** Consistency boundary, one root
**Repositories:** Persistence abstraction for aggregates
**Domain Services:** Logic across multiple aggregates
**Domain Events:** Something that happened
**Bounded Context:** Model boundary
**Ubiquitous Language:** Domain terms everywhere

**Dependency direction:** Infrastructure → Domain (never reverse)

---

## For This Project

**Bounded Contexts:**
- Asset Management (M2)
- Market Data (M3)
- Telemetry (M5)
- Trading/Bidding (M6)
- Economics (separate service)

**Aggregates:**
- Battery (Asset Management)
- Price (Market Data)
- BatteryState (Telemetry)
- BiddingStrategy (Trading)

**Value Objects:**
- Constraints
- Money (future)
- TimeRange (future)

**Repositories:**
- BatteryRepository
- PriceRepository (future)
- StateRepository (future)

---

**Remember: DDD is about making implicit domain knowledge explicit in code.**
