# M3: Domain Specification - MarketPrice Aggregate

## 🎯 Domain Model

### MarketPrice Aggregate Root

```go
package domain

import (
    "errors"
    "time"
)

// MarketPrice represents a price forecast for the Australian NEM
type MarketPrice struct {
    ID            string        // UUID
    Region        string        // NSW, VIC, QLD, SA, TAS
    Price         float64       // $/MWh (Australian dollars per megawatt-hour)
    Demand        float64       // MW (system-wide demand forecast)
    IntervalType  IntervalType  // 5MIN_PREDISPATCH or 30MIN_PREDISPATCH
    IntervalStart time.Time     // ISO 8601 timestamp
    PublishedAt   time.Time     // When AEMO published this forecast
    CreatedAt     time.Time     // When we received it
}

// IntervalType defines AEMO market data interval types
type IntervalType string

const (
    Interval5Min  IntervalType = "5MIN_PREDISPATCH"  // 5-minute dispatch intervals
    Interval30Min IntervalType = "30MIN_PREDISPATCH" // 30-minute trading intervals
)
```

## ✅ Validation Rules

### Business Rules (Invariants)

1. **Price Rules**
   - Must be >= 0 (non-negative)
   - Typical range: -$1,000 to $15,000/MWh (NEM can go negative during oversupply)
   - For M3: simplified to >= 0 (negative prices in future iteration)
   - Reason: Market constraint (spot prices can be extreme)

2. **Demand Rules**
   - Must be > 0
   - Typical range: 1,000 - 35,000 MW (system-wide)
   - Reason: Physical constraint (demand must exist)

3. **Region Rules**
   - Must be one of: NSW, VIC, QLD, SA, TAS
   - Same as M2 (Australian NEM regions)
   - Reason: AEMO market structure

4. **IntervalType Rules**
   - Must be either "5MIN_PREDISPATCH" or "30MIN_PREDISPATCH"
   - Reason: AEMO publishes two types of forecasts
   - 5-minute: Real-time dispatch signals
   - 30-minute: Trading interval settlement

5. **IntervalStart Rules (Time Constraints)**
   - **Cannot be in the past** (forecasts are for future intervals)
   - **Must align with IntervalType**:
     - 5MIN: timestamp must be on 5-minute boundaries (e.g., 10:00, 10:05, 10:10)
     - 30MIN: timestamp must be on 30-minute boundaries (e.g., 10:00, 10:30, 11:00)
   - Reason: AEMO forecast structure and market rules

6. **PublishedAt Rules**
   - Cannot be in the future (sanity check)
   - Must be before or equal to current time
   - Reason: Logical constraint (can't publish in future)

7. **Duplicate Prevention**
   - No duplicate (Region, IntervalType, IntervalStart) combinations
   - Enforced at repository level with unique index
   - Reason: Each forecast interval is unique per region and type

8. **Interval Alignment Validation**
   - **5-minute alignment**: minute % 5 == 0 AND second == 0
     - Valid: 10:00:00, 10:05:00, 10:10:00, 10:15:00, etc.
     - Invalid: 10:03:00, 10:01:30
   - **30-minute alignment**: (minute == 0 OR minute == 30) AND second == 0
     - Valid: 10:00:00, 10:30:00, 11:00:00
     - Invalid: 10:15:00, 10:45:00
   - Reason: AEMO market dispatch schedule

## 🏗️ Constructor and Methods

### NewMarketPrice (Constructor)

```go
// NewMarketPrice creates a new MarketPrice with validation
func NewMarketPrice(
    region string,
    price float64,
    demand float64,
    intervalType IntervalType,
    intervalStart time.Time,
    publishedAt time.Time,
) (*MarketPrice, error) {
    // Generate ID (UUID)
    id := uuid.New().String()

    // Create market price
    marketPrice := &MarketPrice{
        ID:            id,
        Region:        region,
        Price:         price,
        Demand:        demand,
        IntervalType:  intervalType,
        IntervalStart: intervalStart,
        PublishedAt:   publishedAt,
        CreatedAt:     time.Now(),
    }

    // Validate
    if err := marketPrice.Validate(); err != nil {
        return nil, err
    }

    return marketPrice, nil
}
```

### Validate Method

```go
// Validate checks all business rules
func (mp *MarketPrice) Validate() error {
    // Price validation
    if mp.Price < 0 {
        return ErrInvalidPrice
    }

    // Demand validation
    if mp.Demand <= 0 {
        return ErrInvalidDemand
    }

    // Region validation
    if !isValidRegion(mp.Region) {
        return ErrInvalidRegion
    }

    // IntervalType validation
    if !isValidIntervalType(mp.IntervalType) {
        return ErrInvalidIntervalType
    }

    // Time constraints
    if mp.IntervalStart.Before(time.Now()) {
        return ErrIntervalInPast
    }

    if mp.PublishedAt.After(time.Now()) {
        return ErrPublishedInFuture
    }

    // Interval alignment validation
    if !validateIntervalAlignment(mp.IntervalStart, mp.IntervalType) {
        return ErrIntervalAlignment
    }

    return nil
}

// isValidRegion checks if region is a valid NEM region
func isValidRegion(region string) bool {
    validRegions := map[string]bool{
        "NSW": true,
        "VIC": true,
        "QLD": true,
        "SA":  true,
        "TAS": true,
    }
    return validRegions[region]
}

// isValidIntervalType checks if interval type is valid
func isValidIntervalType(intervalType IntervalType) bool {
    return intervalType == Interval5Min || intervalType == Interval30Min
}

// validateIntervalAlignment checks if timestamp aligns with interval type
func validateIntervalAlignment(t time.Time, intervalType IntervalType) bool {
    minute := t.Minute()
    second := t.Second()

    if intervalType == Interval5Min {
        // Must be on 5-minute boundary (0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55)
        return minute%5 == 0 && second == 0
    }

    if intervalType == Interval30Min {
        // Must be on 30-minute boundary (0 or 30)
        return (minute == 0 || minute == 30) && second == 0
    }

    return false
}
```

## 🚨 Domain Errors

```go
package domain

import "errors"

var (
    // MarketPrice validation errors
    ErrInvalidPrice        = errors.New("price must be non-negative ($/MWh)")
    ErrInvalidDemand       = errors.New("demand must be greater than 0 (MW)")
    ErrInvalidRegion       = errors.New("region must be valid NEM region (NSW, VIC, QLD, SA, TAS)")
    ErrInvalidIntervalType = errors.New("interval type must be 5MIN_PREDISPATCH or 30MIN_PREDISPATCH")
    ErrIntervalInPast      = errors.New("interval start cannot be in the past")
    ErrPublishedInFuture   = errors.New("published time cannot be in the future")
    ErrIntervalAlignment   = errors.New("interval start must align with interval type (5-min or 30-min boundaries)")

    // Repository errors
    ErrNotFound          = errors.New("market price not found")
    ErrDuplicateInterval = errors.New("price for this region/interval/time already exists")
)
```

## 🧪 Test Cases (TDD)

### Test Structure (Table-Driven)

```go
package domain

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewMarketPrice_ValidInput(t *testing.T) {
    tests := []struct {
        name          string
        region        string
        price         float64
        demand        float64
        intervalType  IntervalType
        intervalStart time.Time
        publishedAt   time.Time
    }{
        {
            name:          "valid 5-minute forecast",
            region:        "NSW",
            price:         85.50,
            demand:        8200.0,
            intervalType:  Interval5Min,
            intervalStart: time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
            publishedAt:   time.Now().Add(-5 * time.Minute),
        },
        {
            name:          "valid 30-minute forecast",
            region:        "SA",
            price:         120.00,
            demand:        1500.0,
            intervalType:  Interval30Min,
            intervalStart: time.Date(2025, 12, 30, 10, 30, 0, 0, time.UTC),
            publishedAt:   time.Now().Add(-10 * time.Minute),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // When
            price, err := NewMarketPrice(
                tt.region, tt.price, tt.demand,
                tt.intervalType, tt.intervalStart, tt.publishedAt,
            )

            // Then
            require.NoError(t, err)
            assert.NotNil(t, price)
            assert.NotEmpty(t, price.ID)
            assert.Equal(t, tt.price, price.Price)
            assert.Equal(t, tt.region, price.Region)
        })
    }
}

func TestNewMarketPrice_InvalidPrice(t *testing.T) {
    tests := []struct {
        name  string
        price float64
    }{
        {"negative price", -10.0},
        {"very negative price", -100.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Given
            intervalStart := time.Now().Add(1 * time.Hour)
            publishedAt := time.Now()

            // When
            price, err := NewMarketPrice(
                "NSW", tt.price, 8000.0,
                Interval5Min, intervalStart, publishedAt,
            )

            // Then
            assert.Error(t, err)
            assert.Nil(t, price)
            assert.ErrorIs(t, err, ErrInvalidPrice)
        })
    }
}

func TestNewMarketPrice_IntervalAlignment(t *testing.T) {
    tests := []struct {
        name          string
        intervalType  IntervalType
        intervalStart time.Time
        shouldFail    bool
    }{
        {
            name:          "5MIN valid - on 5-minute boundary",
            intervalType:  Interval5Min,
            intervalStart: time.Date(2025, 12, 30, 10, 5, 0, 0, time.UTC),
            shouldFail:    false,
        },
        {
            name:          "5MIN invalid - not on boundary",
            intervalType:  Interval5Min,
            intervalStart: time.Date(2025, 12, 30, 10, 3, 0, 0, time.UTC),
            shouldFail:    true,
        },
        {
            name:          "30MIN valid - on 30-minute boundary",
            intervalType:  Interval30Min,
            intervalStart: time.Date(2025, 12, 30, 10, 30, 0, 0, time.UTC),
            shouldFail:    false,
        },
        {
            name:          "30MIN invalid - not on boundary",
            intervalType:  Interval30Min,
            intervalStart: time.Date(2025, 12, 30, 10, 15, 0, 0, time.UTC),
            shouldFail:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // When
            price, err := NewMarketPrice(
                "NSW", 80.0, 8000.0,
                tt.intervalType, tt.intervalStart, time.Now(),
            )

            // Then
            if tt.shouldFail {
                assert.Error(t, err)
                assert.ErrorIs(t, err, ErrIntervalAlignment)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, price)
            }
        })
    }
}
```

## 📊 Example Market Prices

### 5-Minute Dispatch Forecast (FCAS Response)
```go
MarketPrice{
    Region:        "NSW",
    Price:         85.50,    // $/MWh
    Demand:        8200.0,   // MW (system demand)
    IntervalType:  Interval5Min,
    IntervalStart: "2025-12-30T10:05:00Z", // 5-min boundary
    PublishedAt:   "2025-12-30T10:00:00Z", // Published 5 min before
}
```

### 30-Minute Trading Interval Forecast (Energy Arbitrage)
```go
MarketPrice{
    Region:        "SA",
    Price:         120.00,   // $/MWh (higher price, good for discharge)
    Demand:        1500.0,   // MW
    IntervalType:  Interval30Min,
    IntervalStart: "2025-12-30T11:00:00Z", // 30-min boundary
    PublishedAt:   "2025-12-30T10:30:00Z", // Published 30 min before
}
```

### Low Price (Charging Opportunity)
```go
MarketPrice{
    Region:        "VIC",
    Price:         35.00,    // $/MWh (low price, good for charging)
    Demand:        6500.0,   // MW
    IntervalType:  Interval30Min,
    IntervalStart: "2025-12-30T02:00:00Z", // Off-peak
    PublishedAt:   "2025-12-30T01:30:00Z",
}
```

## 🎯 Key Takeaways

1. **Time-series immutability**: Prices are forecasts, never updated once created
2. **Interval alignment is critical**: AEMO operates on strict time boundaries
3. **Validation is strict**: Fail fast with clear errors
4. **Timestamps use UTC**: Avoid timezone issues
5. **Region isolation**: Each region has independent pricing
6. **Go idioms**: Table-driven tests, error handling patterns

## 🔍 Key Differences from M2 (Battery)

| Aspect | M2 Battery | M3 MarketPrice |
|--------|-----------|----------------|
| Entity Type | Long-lived entity | Immutable time-series record |
| Primary Complexity | Nested constraints | Interval alignment |
| Validation Rules | 10 rules | 8 rules |
| Update Pattern | Status changes (future) | Never updated (immutable) |
| Key Uniqueness | UUID only | UUID + (region, interval, time) |
| Time Sensitivity | Created/Updated timestamps | Interval boundaries critical |

---

**Next**: Implement this domain model following TDD!
