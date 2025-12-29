# M2: Domain Specification - Battery Aggregate

## 🎯 Domain Model

### Battery Aggregate Root

```go
package domain

import (
    "errors"
    "time"
)

// Battery represents a battery energy storage system in the Australian NEM
type Battery struct {
    ID           string
    Capacity     float64    // MWh (Megawatt-hours)
    MaxPower     float64    // MW (Megawatts)
    RampRate     float64    // MW/min
    Efficiency   float64    // Round-trip efficiency (0.0-1.0)
    Location     string     // NSW, VIC, QLD, SA, TAS
    Manufacturer string     // Tesla, BYD, etc.
    Constraints  Constraints
    Status       BatteryStatus
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Constraints defines operational and regulatory constraints
type Constraints struct {
    // Warranty and degradation
    WarrantyEOL float64 // End of life SoC (e.g., 0.7 = 70%)
    MaxCycles   int     // Maximum charge/discharge cycles
    
    // Temperature limits
    TempMin float64 // Minimum operating temperature (°C)
    TempMax float64 // Maximum operating temperature (°C)
    
    // Grid compliance
    GridCompliance []string // e.g., ["FCAS", "FFR", "AEMO_COMPLIANT"]
}

// BatteryStatus represents the current lifecycle status
type BatteryStatus string

const (
    StatusRegistered BatteryStatus = "REGISTERED"  // Just added to system
    StatusTesting    BatteryStatus = "TESTING"     // Undergoing initial tests
    StatusOperational BatteryStatus = "OPERATIONAL" // Ready for operations
    StatusMaintenance BatteryStatus = "MAINTENANCE" // Under maintenance
    StatusDecommissioned BatteryStatus = "DECOMMISSIONED" // No longer in use
)
```

## ✅ Validation Rules

### Business Rules (Invariants)

1. **Capacity Rules**
   - Must be > 0
   - Typical range: 0.2 - 500 MWh
   - Reason: Physical constraint

2. **MaxPower Rules**
   - Must be > 0
   - Must be <= Capacity (can't discharge more power than capacity in 1 hour)
   - Typical range: 0.1 - 200 MW
   - Reason: Physical and grid connection constraints

3. **RampRate Rules**
   - Must be > 0
   - Must be <= MaxPower (can't ramp faster than max power)
   - Typical range: 1 - 50 MW/min
   - Reason: Inverter and hardware limitations

4. **Efficiency Rules**
   - Must be between 0.0 and 1.0
   - Typical range: 0.85 - 0.95 (85% - 95%)
   - Reason: Physics - energy is lost during conversion

5. **WarrantyEOL Rules**
   - Must be between 0.0 and 1.0
   - Typical range: 0.6 - 0.8 (60% - 80%)
   - Reason: Manufacturer warranty terms

6. **MaxCycles Rules**
   - Must be > 0
   - Typical range: 5,000 - 15,000 cycles
   - Reason: Battery degradation characteristics

7. **Temperature Rules**
   - TempMin must be < TempMax
   - Typical range: -10°C to 45°C
   - Reason: Battery chemistry constraints

8. **Location Rules**
   - Must be one of: NSW, VIC, QLD, SA, TAS
   - Reason: Australian NEM regions

9. **Status Transitions**
   - REGISTERED → TESTING → OPERATIONAL
   - OPERATIONAL ↔ MAINTENANCE
   - Any → DECOMMISSIONED (final state)

## 🏗️ Constructor and Methods

### NewBattery (Constructor)

```go
// NewBattery creates a new Battery with validation
func NewBattery(
    capacity float64,
    maxPower float64,
    rampRate float64,
    efficiency float64,
    location string,
    manufacturer string,
    constraints Constraints,
) (*Battery, error) {
    // Generate ID (UUID)
    id := generateID()
    
    // Create battery
    battery := &Battery{
        ID:           id,
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
    
    // Validate
    if err := battery.Validate(); err != nil {
        return nil, err
    }
    
    return battery, nil
}
```

### Validate Method

```go
// Validate checks all business rules
func (b *Battery) Validate() error {
    var errs []error
    
    // Capacity validation
    if b.Capacity <= 0 {
        errs = append(errs, ErrInvalidCapacity)
    }
    
    // MaxPower validation
    if b.MaxPower <= 0 {
        errs = append(errs, ErrInvalidMaxPower)
    }
    if b.MaxPower > b.Capacity {
        errs = append(errs, ErrMaxPowerExceedsCapacity)
    }
    
    // RampRate validation
    if b.RampRate <= 0 {
        errs = append(errs, ErrInvalidRampRate)
    }
    if b.RampRate > b.MaxPower {
        errs = append(errs, ErrRampRateExceedsMaxPower)
    }
    
    // Efficiency validation
    if b.Efficiency < 0 || b.Efficiency > 1 {
        errs = append(errs, ErrInvalidEfficiency)
    }
    
    // Location validation
    if !isValidLocation(b.Location) {
        errs = append(errs, ErrInvalidLocation)
    }
    
    // Constraints validation
    if err := b.Constraints.Validate(); err != nil {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return NewValidationError(errs)
    }
    
    return nil
}

// isValidLocation checks if location is a valid NEM region
func isValidLocation(loc string) bool {
    validLocations := []string{"NSW", "VIC", "QLD", "SA", "TAS"}
    for _, valid := range validLocations {
        if loc == valid {
            return true
        }
    }
    return false
}
```

### Constraints Validation

```go
// Validate checks constraints business rules
func (c *Constraints) Validate() error {
    var errs []error
    
    // WarrantyEOL validation
    if c.WarrantyEOL < 0 || c.WarrantyEOL > 1 {
        errs = append(errs, ErrInvalidWarrantyEOL)
    }
    
    // MaxCycles validation
    if c.MaxCycles <= 0 {
        errs = append(errs, ErrInvalidMaxCycles)
    }
    
    // Temperature validation
    if c.TempMin >= c.TempMax {
        errs = append(errs, ErrInvalidTemperatureRange)
    }
    
    if len(errs) > 0 {
        return NewValidationError(errs)
    }
    
    return nil
}
```

## 🚨 Domain Errors

```go
package domain

import "errors"

var (
    // Battery validation errors
    ErrInvalidCapacity          = errors.New("capacity must be greater than 0")
    ErrInvalidMaxPower          = errors.New("max power must be greater than 0")
    ErrMaxPowerExceedsCapacity  = errors.New("max power cannot exceed capacity")
    ErrInvalidRampRate          = errors.New("ramp rate must be greater than 0")
    ErrRampRateExceedsMaxPower  = errors.New("ramp rate cannot exceed max power")
    ErrInvalidEfficiency        = errors.New("efficiency must be between 0 and 1")
    ErrInvalidLocation          = errors.New("location must be NSW, VIC, QLD, SA, or TAS")
    
    // Constraints validation errors
    ErrInvalidWarrantyEOL       = errors.New("warranty EOL must be between 0 and 1")
    ErrInvalidMaxCycles         = errors.New("max cycles must be greater than 0")
    ErrInvalidTemperatureRange  = errors.New("minimum temperature must be less than maximum")
    
    // Business logic errors
    ErrBatteryNotFound          = errors.New("battery not found")
    ErrBatteryAlreadyExists     = errors.New("battery already exists")
)

// ValidationError wraps multiple validation errors
type ValidationError struct {
    Errors []error
}

func NewValidationError(errs []error) *ValidationError {
    return &ValidationError{Errors: errs}
}

func (e *ValidationError) Error() string {
    if len(e.Errors) == 0 {
        return "validation error"
    }
    return e.Errors[0].Error()
}
```

## 🧪 Test Cases (TDD)

### Test Structure (Table-Driven)

```go
package domain

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewBattery_ValidInput(t *testing.T) {
    // Given
    capacity := 200.0
    maxPower := 100.0
    rampRate := 10.0
    efficiency := 0.92
    location := "NSW"
    manufacturer := "Tesla"
    constraints := Constraints{
        WarrantyEOL: 0.7,
        MaxCycles:   10000,
        TempMin:     -10.0,
        TempMax:     45.0,
        GridCompliance: []string{"FCAS", "FFR"},
    }
    
    // When
    battery, err := NewBattery(
        capacity, maxPower, rampRate, efficiency,
        location, manufacturer, constraints,
    )
    
    // Then
    assert.NoError(t, err)
    assert.NotNil(t, battery)
    assert.NotEmpty(t, battery.ID)
    assert.Equal(t, capacity, battery.Capacity)
    assert.Equal(t, StatusRegistered, battery.Status)
}

func TestNewBattery_InvalidInputs(t *testing.T) {
    tests := []struct {
        name        string
        capacity    float64
        maxPower    float64
        rampRate    float64
        efficiency  float64
        location    string
        expectedErr error
    }{
        {
            name:        "negative capacity",
            capacity:    -100.0,
            maxPower:    50.0,
            rampRate:    5.0,
            efficiency:  0.9,
            location:    "NSW",
            expectedErr: ErrInvalidCapacity,
        },
        {
            name:        "zero capacity",
            capacity:    0.0,
            maxPower:    50.0,
            rampRate:    5.0,
            efficiency:  0.9,
            location:    "NSW",
            expectedErr: ErrInvalidCapacity,
        },
        {
            name:        "max power exceeds capacity",
            capacity:    100.0,
            maxPower:    150.0,
            rampRate:    5.0,
            efficiency:  0.9,
            location:    "NSW",
            expectedErr: ErrMaxPowerExceedsCapacity,
        },
        {
            name:        "efficiency too high",
            capacity:    100.0,
            maxPower:    50.0,
            rampRate:    5.0,
            efficiency:  1.5,
            location:    "NSW",
            expectedErr: ErrInvalidEfficiency,
        },
        {
            name:        "invalid location",
            capacity:    100.0,
            maxPower:    50.0,
            rampRate:    5.0,
            efficiency:  0.9,
            location:    "INVALID",
            expectedErr: ErrInvalidLocation,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Given
            constraints := Constraints{
                WarrantyEOL: 0.7,
                MaxCycles:   10000,
                TempMin:     -10.0,
                TempMax:     45.0,
            }
            
            // When
            battery, err := NewBattery(
                tt.capacity, tt.maxPower, tt.rampRate,
                tt.efficiency, tt.location, "Tesla", constraints,
            )
            
            // Then
            assert.Error(t, err)
            assert.Nil(t, battery)
            assert.ErrorIs(t, err, tt.expectedErr)
        })
    }
}

func TestConstraints_Validate(t *testing.T) {
    tests := []struct {
        name        string
        constraints Constraints
        expectedErr error
    }{
        {
            name: "valid constraints",
            constraints: Constraints{
                WarrantyEOL: 0.7,
                MaxCycles:   10000,
                TempMin:     -10.0,
                TempMax:     45.0,
            },
            expectedErr: nil,
        },
        {
            name: "invalid warranty EOL",
            constraints: Constraints{
                WarrantyEOL: 1.5,
                MaxCycles:   10000,
                TempMin:     -10.0,
                TempMax:     45.0,
            },
            expectedErr: ErrInvalidWarrantyEOL,
        },
        {
            name: "invalid temperature range",
            constraints: Constraints{
                WarrantyEOL: 0.7,
                MaxCycles:   10000,
                TempMin:     50.0,
                TempMax:     45.0,
            },
            expectedErr: ErrInvalidTemperatureRange,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // When
            err := tt.constraints.Validate()
            
            // Then
            if tt.expectedErr != nil {
                assert.Error(t, err)
                assert.ErrorIs(t, err, tt.expectedErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 📊 Example Batteries

### Small Commercial Battery
```go
Battery{
    Capacity:     0.5,   // 500 kWh
    MaxPower:     0.25,  // 250 kW
    RampRate:     2.5,   // 2.5 MW/min
    Efficiency:   0.88,
    Location:     "NSW",
    Manufacturer: "BYD",
}
```

### Large Utility-Scale Battery
```go
Battery{
    Capacity:     200.0,  // 200 MWh (like Hornsdale)
    MaxPower:     100.0,  // 100 MW
    RampRate:     10.0,   // 10 MW/min
    Efficiency:   0.92,
    Location:     "SA",
    Manufacturer: "Tesla",
}
```

## 🎯 Key Takeaways

1. **Domain is pure**: No database, no HTTP, just business logic
2. **Validation is strict**: Fail fast with clear errors
3. **Immutability preferred**: Create new instances rather than modify
4. **Tests are documentation**: Each test explains a rule
5. **Go idioms**: Table-driven tests, error handling patterns

---

**Next**: Implement this domain model following TDD!
