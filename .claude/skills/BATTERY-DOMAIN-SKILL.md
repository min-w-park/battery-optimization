# Battery Optimization Domain Skill

## When to use this skill
Use this skill when working on battery optimization domain logic, including Battery aggregates, validation rules, and energy market concepts.

## Domain Context

This project simulates battery optimization for the **Australian National Electricity Market (NEM)**.

### Key Concepts

**Battery Energy Storage System (BESS)**:
- Stores energy (MWh - Megawatt Hours)
- Discharges at power rate (MW - Megawatts)
- Participates in energy markets
- Subject to physical and regulatory constraints

**State of Charge (SoC)**:
- Battery's current energy level (%)
- 0% = empty, 100% = full capacity
- Critical for operations (can't discharge when empty)

**FCAS (Frequency Control Ancillary Services)**:
- Grid stability services in Australia
- Battery responds to frequency deviations
- Two types: Raise (increase frequency), Lower (decrease frequency)
- High-value revenue stream

**Energy Arbitrage**:
- Buy energy when price is low (charge battery)
- Sell energy when price is high (discharge battery)
- Core revenue strategy

---

## Domain Model

### Battery Aggregate

```go
type Battery struct {
    ID           string        // Unique identifier
    Capacity     float64       // MWh - total energy storage
    MaxPower     float64       // MW - maximum charge/discharge rate
    RampRate     float64       // MW/min - how fast power can change
    Efficiency   float64       // 0.0-1.0 - round-trip efficiency
    Location     string        // NEM region: NSW, VIC, QLD, SA, TAS
    Manufacturer string        // Tesla, BYD, etc.
    Constraints  Constraints   // Operational limits
    Status       BatteryStatus // Lifecycle state
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Validation Rules (Business Invariants)

**These are CRITICAL - they reflect real physics and regulations:**

1. **Capacity > 0** (can't have negative energy storage)
2. **MaxPower > 0** (must be able to charge/discharge)
3. **MaxPower ≤ Capacity** (can't discharge more than capacity in 1 hour)
4. **RampRate > 0** (must be able to change power)
5. **RampRate ≤ MaxPower** (can't ramp faster than max power)
6. **0 ≤ Efficiency ≤ 1** (physics constraint)
7. **WarrantyEOL between 0-1** (manufacturer warranty)
8. **MaxCycles > 0** (battery degrades over cycles)
9. **TempMin < TempMax** (operating temperature range)
10. **Location must be valid NEM region** (NSW, VIC, QLD, SA, TAS)

### Constraints

```go
type Constraints struct {
    WarrantyEOL    float64   // End-of-life SoC (e.g., 0.7 = 70%)
    MaxCycles      int       // Before warranty expires
    TempMin        float64   // Minimum operating temp (°C)
    TempMax        float64   // Maximum operating temp (°C)
    GridCompliance []string  // e.g., ["FCAS", "FFR"]
}
```

**Real-world context:**
- **WarrantyEOL**: Batteries degrade. Warranty typically guarantees 70-80% capacity
- **MaxCycles**: Typical range 5,000-15,000 cycles
- **Temperature**: Batteries fail outside operating range (-10°C to 45°C typical)
- **GridCompliance**: Must meet AEMO standards for market participation

### Battery Status Lifecycle

```go
type BatteryStatus string

const (
    StatusRegistered     BatteryStatus = "REGISTERED"      // Just added
    StatusTesting        BatteryStatus = "TESTING"         // Initial tests
    StatusOperational    BatteryStatus = "OPERATIONAL"     // Ready for market
    StatusMaintenance    BatteryStatus = "MAINTENANCE"     // Offline for service
    StatusDecommissioned BatteryStatus = "DECOMMISSIONED"  // Retired
)
```

**Valid Transitions:**
- REGISTERED → TESTING → OPERATIONAL
- OPERATIONAL ↔ MAINTENANCE
- Any → DECOMMISSIONED (final state)

---

## Implementation Patterns

### Pattern 1: Constructor with Validation

```go
func NewBattery(
    capacity float64,
    maxPower float64,
    rampRate float64,
    efficiency float64,
    location string,
    manufacturer string,
    constraints Constraints,
) (*Battery, error) {
    battery := &Battery{
        ID:           generateID(), // UUID
        Capacity:     capacity,
        MaxPower:     maxPower,
        RampRate:     rampRate,
        Efficiency:   efficiency,
        Location:     location,
        Manufacturer: manufacturer,
        Constraints:  constraints,
        Status:       StatusRegistered,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    if err := battery.Validate(); err != nil {
        return nil, err
    }
    
    return battery, nil
}
```

**Key points:**
- Always validate in constructor
- Fail fast - don't create invalid objects
- Return error, not nil battery

### Pattern 2: Validation with Multiple Errors

```go
func (b *Battery) Validate() error {
    var errs []error
    
    if b.Capacity <= 0 {
        errs = append(errs, ErrInvalidCapacity)
    }
    if b.MaxPower > b.Capacity {
        errs = append(errs, ErrMaxPowerExceedsCapacity)
    }
    // ... more validations
    
    if len(errs) > 0 {
        return NewValidationError(errs)
    }
    return nil
}
```

**Why collect errors?**
- User gets all problems at once
- Better UX than one-at-a-time
- Especially important for API responses

### Pattern 3: Domain Errors (Not Generic Errors)

```go
var (
    ErrInvalidCapacity         = errors.New("capacity must be greater than 0")
    ErrMaxPowerExceedsCapacity = errors.New("max power cannot exceed capacity")
    ErrInvalidLocation         = errors.New("location must be NSW, VIC, QLD, SA, or TAS")
)
```

**Why specific errors?**
- Tests can assert exact error
- HTTP layer can map to proper status codes
- Better debugging
- Clear business meaning

---

## Example Batteries (Real-World Scale)

### Small Commercial (e.g., shopping center)
```go
Battery{
    Capacity:     0.5,    // 500 kWh
    MaxPower:     0.25,   // 250 kW (C-rate = 0.5)
    RampRate:     2.5,    // 2.5 MW/min
    Efficiency:   0.88,   // 88% round-trip
    Location:     "NSW",
    Manufacturer: "BYD",
}
```

### Medium Industrial (e.g., factory)
```go
Battery{
    Capacity:     5.0,    // 5 MWh
    MaxPower:     2.5,    // 2.5 MW (C-rate = 0.5)
    RampRate:     5.0,    // 5 MW/min
    Efficiency:   0.90,   // 90% round-trip
    Location:     "VIC",
    Manufacturer: "Tesla",
}
```

### Large Utility-Scale (e.g., Hornsdale Power Reserve)
```go
Battery{
    Capacity:     200.0,  // 200 MWh
    MaxPower:     100.0,  // 100 MW (C-rate = 0.5)
    RampRate:     10.0,   // 10 MW/min
    Efficiency:   0.92,   // 92% round-trip
    Location:     "SA",
    Manufacturer: "Tesla",
}
```

**C-rate**: MaxPower / Capacity
- 1C = full charge/discharge in 1 hour
- 0.5C = 2 hours to full charge/discharge
- Higher C-rate = more stress on battery

---

## Common Validation Scenarios

### Scenario 1: Negative Capacity
```go
// Test
func TestNewBattery_NegativeCapacity_ReturnsError(t *testing.T) {
    _, err := NewBattery(-100.0, 50.0, ...)
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}
```

**Why it matters:** Physics - can't store negative energy

### Scenario 2: Power Exceeds Capacity
```go
// Test
func TestNewBattery_PowerExceedsCapacity_ReturnsError(t *testing.T) {
    _, err := NewBattery(100.0, 150.0, ...) // 100 MWh, 150 MW
    assert.ErrorIs(t, err, ErrMaxPowerExceedsCapacity)
}
```

**Why it matters:** Power rating determines how fast you can charge/discharge. If MaxPower > Capacity, battery would fully charge/discharge in <1 hour, which creates thermal and degradation issues.

### Scenario 3: Invalid NEM Region
```go
// Test
func TestNewBattery_InvalidLocation_ReturnsError(t *testing.T) {
    _, err := NewBattery(100.0, 50.0, ..., "CALIFORNIA", ...)
    assert.ErrorIs(t, err, ErrInvalidLocation)
}
```

**Why it matters:** This project is Australian NEM specific. Different regions have different market rules.

### Scenario 4: Temperature Range
```go
// Test
func TestConstraints_MinTempExceedsMax_ReturnsError(t *testing.T) {
    constraints := Constraints{TempMin: 50.0, TempMax: 45.0}
    err := constraints.Validate()
    assert.ErrorIs(t, err, ErrInvalidTemperatureRange)
}
```

**Why it matters:** Battery chemistry has strict temperature limits. Operating outside range causes damage or fire risk.

---

## Anti-Patterns to Avoid

### ❌ Anti-Pattern 1: Validation in Multiple Places
```go
// BAD - validation scattered
func NewBattery(...) (*Battery, error) {
    if capacity <= 0 { return nil, err }  // Validation here
    return &Battery{...}, nil
}

func (s *Service) RegisterBattery(...) error {
    if capacity <= 0 { return err }  // AND here (duplication!)
    // ...
}
```

**✅ Better:** Validation ONLY in domain (Battery.Validate())

### ❌ Anti-Pattern 2: Generic Validation Errors
```go
// BAD
return errors.New("invalid input")  // Which input?!
```

**✅ Better:**
```go
return ErrInvalidCapacity  // Specific, testable
```

### ❌ Anti-Pattern 3: Exposing Internal State
```go
// BAD
type Battery struct {
    Capacity      float64
    validatedFlag bool  // Implementation detail leaking
}
```

**✅ Better:** Keep validation internal, only expose behavior

### ❌ Anti-Pattern 4: Validating in Getters
```go
// BAD
func (b *Battery) Capacity() float64 {
    if b.capacity <= 0 {
        panic("invalid capacity")  // Too late!
    }
    return b.capacity
}
```

**✅ Better:** Validate in constructor, guarantee invariants always hold

---

## Integration with Events (Future)

When battery state changes, domain events are published:

```go
// Not implemented in M2, but design for it
type BatteryRegistered struct {
    BatteryID    string
    Capacity     float64
    MaxPower     float64
    RegisteredAt time.Time
}

func (b *Battery) Register() (BatteryRegistered, error) {
    // Change state
    b.Status = StatusRegistered
    
    // Return event
    return BatteryRegistered{
        BatteryID: b.ID,
        Capacity:  b.Capacity,
        MaxPower:  b.MaxPower,
        RegisteredAt: time.Now(),
    }, nil
}
```

**Design principle:** Domain objects return events, infrastructure publishes them

---

## Testing Strategy

### Unit Tests (Domain Layer)
```go
// Test business rules in isolation
func TestBattery_Validate_AllRules(t *testing.T) {
    // Test each validation rule
}
```

### Property-Based Tests (Optional, Advanced)
```go
// Test invariants hold for any valid input
func TestBattery_ValidatedBattery_NeverHasNegativeCapacity(t *testing.T) {
    // Generate random valid batteries
    // Assert capacity always > 0
}
```

### Example-Based Tests (Required)
```go
// Test realistic scenarios
func TestBattery_SmallCommercial_ValidatesCorrectly(t *testing.T) {
    battery, err := NewBattery(0.5, 0.25, ...)
    assert.NoError(t, err)
}
```

---

## Key Takeaways

1. **Domain is pure** - No dependencies on DB, HTTP, etc.
2. **Validate aggressively** - Fail fast, clear errors
3. **Business rules are explicit** - Every rule has a reason
4. **Real-world constraints** - Physics and regulations matter
5. **Tests document domain** - They explain "why" rules exist

---

## References

For detailed specifications, see:
- `docs/milestones/M2-DOMAIN-SPEC.md` - Complete Battery aggregate spec
- `docs/milestones/M2-API-SPEC.md` - How domain maps to API
- `EVENTS.md` - Domain events catalog

---

**Remember: The domain is the heart of the system. Get this right, everything else follows.**
