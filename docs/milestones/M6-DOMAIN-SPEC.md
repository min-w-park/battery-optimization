# M6: Bidding Service - Domain Specification

## Domain Model

The Bidding Service domain revolves around **arbitrage decision-making**. Unlike previous services that manage data (batteries, prices, states), this service contains business logic that drives profitability.

## Aggregates

### 1. BiddingDecision Aggregate

The core aggregate representing a decision moment in the arbitrage process.

```go
type BiddingDecision struct {
    ID              string    // UUID
    BatteryID       string    // Which battery this decision is for
    DecisionType    string    // CHARGE, DISCHARGE, NO_ACTION
    Reason          string    // Human-readable explanation
    Price           float64   // Market price at decision time ($/MWh)
    SoC             float64   // Battery SoC at decision time (%)
    Timestamp       time.Time // When decision was made
    AutomationMode  string    // MANUAL, SEMI_AUTO, FULL_AUTO
}
```

**Invariants**:
- `DecisionType` must be one of: `CHARGE`, `DISCHARGE`, `NO_ACTION`
- `Price` must be > 0
- `SoC` must be between 0 and 100
- `AutomationMode` must be one of: `MANUAL`, `SEMI_AUTO`, `FULL_AUTO`
- `Timestamp` cannot be in the future

**Factory Method**:
```go
func NewBiddingDecision(
    batteryID string,
    decisionType string,
    reason string,
    price float64,
    soc float64,
    automationMode string,
) (*BiddingDecision, error)
```

**Validation Rules**:
1. BatteryID cannot be empty
2. DecisionType must be valid enum value
3. Price must be positive
4. SoC must be in range [0, 100]
5. AutomationMode must be valid enum value

## Domain Services

### 1. Arbitrage Service

Contains the core business logic for detecting arbitrage opportunities.

```go
// Constants
const (
    CHARGE_THRESHOLD_MWH    = 50.0   // Below this: charge opportunity
    DISCHARGE_THRESHOLD_MWH = 100.0  // Above this: discharge opportunity
    MIN_SOC_FOR_DISCHARGE   = 30.0   // Minimum SoC to discharge (%)
    MAX_SOC_FOR_CHARGE      = 80.0   // Maximum SoC to charge (%)
)

// Arbitrage decisions
func ShouldCharge(price float64, soc float64, operationState string) bool
func ShouldDischarge(price float64, soc float64, operationState string) bool
```

**Algorithm Specification**:

#### Charging Opportunity Detection

```
Function: ShouldCharge(price, soc, operationState)

Returns: true if ALL conditions met:
  1. price < CHARGE_THRESHOLD_MWH ($50/MWh)
  2. soc < MAX_SOC_FOR_CHARGE (80%)
  3. operationState == "IDLE"

Returns: false otherwise

Example Scenarios:
  ShouldCharge(30.0, 50.0, "IDLE")        → true   (price low, SoC OK, idle)
  ShouldCharge(30.0, 85.0, "IDLE")        → false  (SoC too high)
  ShouldCharge(60.0, 50.0, "IDLE")        → false  (price too high)
  ShouldCharge(30.0, 50.0, "CHARGING")    → false  (already charging)
  ShouldCharge(30.0, 50.0, "FCAS")        → false  (in FCAS contract)
```

#### Discharging Opportunity Detection

```
Function: ShouldDischarge(price, soc, operationState)

Returns: true if ALL conditions met:
  1. price > DISCHARGE_THRESHOLD_MWH ($100/MWh)
  2. soc > MIN_SOC_FOR_DISCHARGE (30%)
  3. operationState == "IDLE"

Returns: false otherwise

Example Scenarios:
  ShouldDischarge(150.0, 60.0, "IDLE")      → true   (price high, SoC OK, idle)
  ShouldDischarge(150.0, 25.0, "IDLE")      → false  (SoC too low)
  ShouldDischarge(80.0, 60.0, "IDLE")       → false  (price too low)
  ShouldDischarge(150.0, 60.0, "DISCHARGING") → false  (already discharging)
```

**Rationale for Thresholds**:

| Threshold | Value | Rationale |
|-----------|-------|-----------|
| CHARGE_THRESHOLD | $50/MWh | Below-average NEM price; profitable to charge |
| DISCHARGE_THRESHOLD | $100/MWh | Above-average NEM price; profitable to discharge |
| MIN_SOC_FOR_DISCHARGE | 30% | Preserve battery health; avoid deep discharge |
| MAX_SOC_FOR_CHARGE | 80% | Preserve battery health; avoid overcharging |

**Why IDLE State Precondition**:
- Prevents conflicts with existing FCAS contracts (battery already committed)
- Avoids interrupting ongoing charge/discharge operations
- Ensures battery is available for new command

### 2. Profit Estimation (Simplified)

```go
func EstimateProfit(priceDelta float64, capacity float64, efficiency float64) float64 {
    // priceDelta: abs(price - threshold)
    // capacity: Battery capacity in MWh
    // efficiency: Round-trip efficiency (e.g., 0.95)
    return priceDelta * capacity * efficiency
}
```

**Example Calculation**:
```
Charging Opportunity:
  Market Price:  $30/MWh
  Threshold:     $50/MWh
  Price Delta:   $20/MWh
  Capacity:      200 MWh
  Efficiency:    0.95

  Expected Profit = $20 * 200 * 0.95 = $3,800 per full cycle

Discharging Opportunity:
  Market Price:  $150/MWh
  Threshold:     $100/MWh
  Price Delta:   $50/MWh
  Capacity:      200 MWh
  Efficiency:    0.95

  Expected Profit = $50 * 200 * 0.95 = $9,500 per full discharge
```

## Value Objects

### 1. AutomationMode

```go
type AutomationMode string

const (
    MANUAL     AutomationMode = "MANUAL"      // Detect only
    SEMI_AUTO  AutomationMode = "SEMI_AUTO"   // Detect + suggest
    FULL_AUTO  AutomationMode = "FULL_AUTO"   // Detect + execute
)
```

**Behavior by Mode**:

| Mode | Detects Opportunity | Publishes Opportunity Event | Issues Command Event |
|------|---------------------|----------------------------|---------------------|
| MANUAL | ✅ | ✅ | ❌ |
| SEMI_AUTO | ✅ | ✅ | ⏸️ (after approval) |
| FULL_AUTO | ✅ | ✅ | ✅ (immediately) |

### 2. DecisionType

```go
type DecisionType string

const (
    CHARGE      DecisionType = "CHARGE"
    DISCHARGE   DecisionType = "DISCHARGE"
    NO_ACTION   DecisionType = "NO_ACTION"
)
```

### 3. OperationState

```go
type OperationState string

const (
    IDLE        OperationState = "IDLE"
    CHARGING    OperationState = "CHARGING"
    DISCHARGING OperationState = "DISCHARGING"
    FCAS        OperationState = "FCAS"
)
```

## Domain Errors

```go
package domain

import "errors"

var (
    // Validation errors
    ErrInvalidSoC           = errors.New("SoC must be between 0 and 100")
    ErrInvalidPrice         = errors.New("price must be greater than 0")
    ErrInvalidDecisionType  = errors.New("decision type must be CHARGE, DISCHARGE, or NO_ACTION")
    ErrInvalidAutomationMode = errors.New("automation mode must be MANUAL, SEMI_AUTO, or FULL_AUTO")
    ErrInvalidOperationState = errors.New("operation state must be IDLE, CHARGING, DISCHARGING, or FCAS")

    // Business rule errors
    ErrBatteryNotReady      = errors.New("battery not ready for operation")
    ErrBatteryNotIdle       = errors.New("battery must be in IDLE state")
    ErrEmptyBatteryID       = errors.New("battery ID cannot be empty")
)
```

## Business Rules

### Rule 1: Charging Preconditions

```
GIVEN a battery with current state
WHEN evaluating charging opportunity
THEN charge ONLY IF:
  - Market price < $50/MWh
  - Battery SoC < 80%
  - Battery OperationState == IDLE
```

**Violations**:
- Price >= $50/MWh → Not profitable
- SoC >= 80% → Battery health risk
- State != IDLE → Conflict with existing operation

### Rule 2: Discharging Preconditions

```
GIVEN a battery with current state
WHEN evaluating discharging opportunity
THEN discharge ONLY IF:
  - Market price > $100/MWh
  - Battery SoC > 30%
  - Battery OperationState == IDLE
```

**Violations**:
- Price <= $100/MWh → Not profitable
- SoC <= 30% → Battery health risk
- State != IDLE → Conflict with existing operation

### Rule 3: Automation Mode Behavior

```
GIVEN an arbitrage opportunity detected
WHEN automation mode is:
  - MANUAL: Publish opportunity event only
  - SEMI_AUTO: Publish opportunity + wait for approval
  - FULL_AUTO: Publish opportunity + issue command immediately
```

### Rule 4: State Consistency

```
GIVEN battery state updates arriving at 1 Hz
WHEN making a decision
THEN use the LATEST state from cache
  - Ignore stale states (based on timestamp)
  - Handle out-of-order events gracefully
```

## Decision Flow

```
┌─────────────────────────────────────────────────────────────┐
│  Event: BatteryStateChanged or MarketPriceUpdated          │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        v
        ┌───────────────────────────────┐
        │  Update In-Memory Cache       │
        │  - BatteryStateCache          │
        │  - PriceCache                 │
        └───────────────┬───────────────┘
                        │
                        v
        ┌───────────────────────────────┐
        │  Get Latest State + Price     │
        │  from caches                  │
        └───────────────┬───────────────┘
                        │
                        v
        ┌───────────────────────────────┐
        │  Arbitrage Algorithm          │
        │  ShouldCharge()?              │
        │  ShouldDischarge()?           │
        └───────────────┬───────────────┘
                        │
        ┌───────────────┴───────────────┐
        │                               │
        v                               v
  Opportunity                     No Opportunity
  Detected                        (no action)
        │                               │
        v                               v
┌───────────────┐               ┌───────────────┐
│ Publish Event │               │   Return      │
│ - Opportunity │               └───────────────┘
└───────┬───────┘
        │
        v
┌───────────────────────────────┐
│ Check Automation Mode         │
└───────┬───────────────────────┘
        │
        ┌─────────────┬─────────────────┬───────────────┐
        v             v                 v               v
     MANUAL       SEMI_AUTO         FULL_AUTO      (invalid)
        │             │                 │
        v             v                 v
    Return        Wait for          Issue Command
                  Approval          Event
                     │                 │
                     v                 v
                Issue Command    ┌─────────────────┐
                Event            │ Publish Command │
                     │           └─────────────────┘
                     v
                ┌─────────────────┐
                │ Publish Command │
                └─────────────────┘
```

## Example Scenarios

### Scenario 1: Charging Opportunity (FULL_AUTO)

```
Initial State:
  - Battery SoC: 50%
  - Battery State: IDLE
  - Market Price: $80/MWh
  - Automation Mode: FULL_AUTO

Event 1: MarketPriceUpdated (price = $30/MWh)
  → Update PriceCache
  → Check: ShouldCharge(30, 50, IDLE)
  → Result: true (30 < 50, 50 < 80, IDLE)
  → Publish: ChargingOpportunityDetected
  → Publish: ChargingCommandIssued (FULL_AUTO)

Event 2: BatteryStateChanged (state = CHARGING)
  → Update BatteryStateCache
  → Check: ShouldCharge(30, 55, CHARGING)
  → Result: false (not IDLE)
  → No action
```

### Scenario 2: Discharging Opportunity (MANUAL)

```
Initial State:
  - Battery SoC: 70%
  - Battery State: IDLE
  - Market Price: $50/MWh
  - Automation Mode: MANUAL

Event 1: MarketPriceUpdated (price = $150/MWh)
  → Update PriceCache
  → Check: ShouldDischarge(150, 70, IDLE)
  → Result: true (150 > 100, 70 > 30, IDLE)
  → Publish: DischargingOpportunityDetected
  → NO command issued (MANUAL mode)
  → Human operator reviews and manually approves
```

### Scenario 3: No Opportunity (Mid-Range Price)

```
Initial State:
  - Battery SoC: 60%
  - Battery State: IDLE
  - Market Price: $50/MWh

Event 1: MarketPriceUpdated (price = $75/MWh)
  → Update PriceCache
  → Check: ShouldCharge(75, 60, IDLE)
  → Result: false (75 >= 50)
  → Check: ShouldDischarge(75, 60, IDLE)
  → Result: false (75 <= 100)
  → No opportunity detected
  → No events published
```

### Scenario 4: FCAS Conflict Avoidance

```
Initial State:
  - Battery SoC: 60%
  - Battery State: FCAS (in FCAS contract)
  - Market Price: $30/MWh (great charging opportunity!)

Event 1: BatteryStateChanged (state = FCAS, soc = 60%)
  → Update BatteryStateCache
  → Check: ShouldCharge(30, 60, FCAS)
  → Result: false (state != IDLE)
  → No opportunity detected (avoid FCAS conflict)
  → Battery continues FCAS contract undisturbed
```

## Testing Strategy

### Unit Tests (Domain Layer)

**Arbitrage Algorithm Tests**:
```go
// Test charging conditions
TestShouldCharge_LowPrice_LowSoC          → true
TestShouldCharge_LowPrice_HighSoC         → false (SoC >= 80%)
TestShouldCharge_HighPrice_LowSoC         → false (price >= 50)
TestShouldCharge_NotIdle                  → false (state != IDLE)

// Test discharging conditions
TestShouldDischarge_HighPrice_HighSoC     → true
TestShouldDischarge_HighPrice_LowSoC      → false (SoC <= 30%)
TestShouldDischarge_LowPrice_HighSoC      → false (price <= 100)
TestShouldDischarge_NotIdle               → false (state != IDLE)

// Edge cases
TestShouldCharge_ExactThreshold           → false (30 == 50)
TestShouldCharge_SoCExactly80             → false (80 == 80)
TestShouldDischarge_SoCExactly30          → false (30 == 30)
```

**BiddingDecision Tests**:
```go
TestNewBiddingDecision_ValidInput         → success
TestNewBiddingDecision_InvalidSoC         → error
TestNewBiddingDecision_InvalidPrice       → error
TestNewBiddingDecision_InvalidMode        → error
TestNewBiddingDecision_EmptyBatteryID     → error
```

### Integration Tests

**Event-Driven Scenarios**:
```go
TestBiddingEngine_ChargingOpportunity     → publishes correct events
TestBiddingEngine_DischargingOpportunity  → publishes correct events
TestBiddingEngine_NoOpportunity           → no events published
TestBiddingEngine_ManualMode              → opportunity only
TestBiddingEngine_FullAutoMode            → opportunity + command
```

## Coverage Targets

- **Domain Layer**: >90% (critical business logic)
- **Cache Layer**: >85% (thread-safety is critical)
- **Service Layer**: >80% (event handlers)
- **Overall**: >75%

---

**This domain model provides the foundation for intelligent, profitable battery arbitrage.** 📊💰
