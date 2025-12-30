# M5: Domain Specification

This document specifies the domain models, business rules, and validation logic for M5 (Telemetry + Device Interface).

---

## 1. Telemetry Service Domain

### 1.1 BatteryState Aggregate

The BatteryState represents a snapshot of battery telemetry at a specific point in time.

#### Structure

```go
type BatteryState struct {
    ID               string                 // Unique ID for this state record
    BatteryID        string                 // Reference to battery (from Asset Management)
    SoC              float64                // State of Charge (0-100%)
    Power            float64                // MW (positive=discharge, negative=charge)
    Temperature      float64                // °C
    Voltage          float64                // V
    Current          float64                // A
    OperationState   string                 // IDLE, CHARGING, DISCHARGING, FCAS
    CustomAttributes map[string]interface{} // Vendor-specific data (JSONB)
    Timestamp        time.Time              // When state was measured
    CreatedAt        time.Time              // When record was created
}
```

#### Field Definitions

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `ID` | string (UUID) | Unique identifier for this state record | `"a1b2c3d4..."` |
| `BatteryID` | string (UUID) | Foreign key to batteries table | `"battery-123"` |
| `SoC` | float64 | State of Charge percentage | `75.5` (75.5%) |
| `Power` | float64 | Current power flow in MW | `25.0` (discharging) or `-15.0` (charging) |
| `Temperature` | float64 | Battery temperature in Celsius | `28.5` |
| `Voltage` | float64 | Battery voltage in Volts | `800.0` |
| `Current` | float64 | Battery current in Amperes | `31.25` |
| `OperationState` | string | Current operational state | `"CHARGING"` |
| `CustomAttributes` | map | Manufacturer-specific fields | `{"cellVoltageMin": 3.2}` |
| `Timestamp` | time.Time | Measurement timestamp (from adapter) | `2025-12-30T10:30:15Z` |
| `CreatedAt` | time.Time | Database record creation time | `2025-12-30T10:30:15.234Z` |

### 1.2 Validation Rules

#### Rule 1: SoC Range
```go
SoC >= 0 && SoC <= 100
```
**Error**: `ErrInvalidSoC` - "SoC must be between 0 and 100%"

#### Rule 2: Temperature Range
```go
Temperature >= -20 && Temperature <= 60
```
**Error**: `ErrInvalidTemperature` - "Temperature must be between -20°C and 60°C (operating range)"

**Rationale**: Safety limits for lithium-ion batteries

#### Rule 3: OperationState Enum
```go
OperationState ∈ {"IDLE", "CHARGING", "DISCHARGING", "FCAS"}
```
**Error**: `ErrInvalidOperationState` - "OperationState must be one of: IDLE, CHARGING, DISCHARGING, FCAS"

#### Rule 4: Power Consistency
```go
if OperationState == "IDLE" then Power == 0
if OperationState == "CHARGING" then Power < 0
if OperationState == "DISCHARGING" then Power > 0
```
**Error**: `ErrInconsistentPowerState` - "Power sign must match OperationState"

**Rationale**: Sign convention consistency (positive = discharge, negative = charge)

#### Rule 5: Required Fields
```go
BatteryID != "" && ID != "" && Timestamp != zero
```
**Error**: `ErrMissingRequiredField` - "BatteryID, ID, and Timestamp are required"

#### Rule 6: Timestamp Validity
```go
Timestamp <= time.Now() + 1 * time.Second
```
**Error**: `ErrInvalidTimestamp` - "Timestamp cannot be in the future (allowing 1s clock skew)"

**Rationale**: Prevent future-dated telemetry data

### 1.3 Constructor

```go
func NewBatteryState(
    batteryID string,
    soc float64,
    power float64,
    temperature float64,
    voltage float64,
    current float64,
    operationState string,
    customAttributes map[string]interface{},
    timestamp time.Time,
) (*BatteryState, error)
```

**Returns**:
- Valid BatteryState with generated ID and CreatedAt
- Error if any validation fails

### 1.4 Domain Errors

```go
var (
    ErrInvalidSoC              = errors.New("invalid SoC value")
    ErrInvalidTemperature      = errors.New("invalid temperature value")
    ErrInvalidOperationState   = errors.New("invalid operation state")
    ErrInconsistentPowerState  = errors.New("power value inconsistent with operation state")
    ErrMissingRequiredField    = errors.New("required field is missing")
    ErrInvalidTimestamp        = errors.New("timestamp is invalid")
    ErrBatteryNotFound         = errors.New("battery not found")
)
```

---

## 2. Device Interface Service Domain

### 2.1 BatteryAdapter Interface

The core abstraction for hardware access.

```go
type BatteryAdapter interface {
    // GetState retrieves current battery state from hardware
    GetState(ctx context.Context) (BatteryState, error)

    // SendCommand sends a command to the battery hardware
    SendCommand(ctx context.Context, cmd Command) error

    // GetCustomAttributes returns manufacturer-specific attributes
    GetCustomAttributes() map[string]interface{}

    // GetBatteryID returns the battery this adapter controls
    GetBatteryID() string
}
```

#### Method Specifications

**GetState**:
- **Purpose**: Read current telemetry from hardware
- **Frequency**: Called by Telemetry Service at 1 Hz
- **Returns**: BatteryState snapshot
- **Error Cases**: Hardware unreachable, timeout, communication error

**SendCommand**:
- **Purpose**: Execute charge/discharge/idle command
- **Timing**: Must respond within 1 second (FCAS requirement)
- **Returns**: nil on success, error on failure
- **Side Effects**: Changes battery operation state

**GetCustomAttributes**:
- **Purpose**: Provide vendor-specific metadata
- **Returns**: Map of manufacturer-specific fields
- **Examples**: Firmware version, cell count, thermal zone data

**GetBatteryID**:
- **Purpose**: Identify which battery this adapter controls
- **Returns**: Battery UUID from Asset Management Service

### 2.2 BatteryState (Device Domain)

Simpler than Telemetry domain - just the hardware readings.

```go
type BatteryState struct {
    SoC            float64  // 0-100%
    Power          float64  // MW
    Temperature    float64  // °C
    Voltage        float64  // V
    Current        float64  // A
    OperationState string   // IDLE, CHARGING, DISCHARGING, FCAS
}
```

**Note**: No ID, Timestamp, or CreatedAt - those are added by Telemetry Service.

### 2.3 Command Types

```go
type CommandType int

const (
    CommandCharge     CommandType = iota // Start charging
    CommandDischarge                     // Start discharging
    CommandIdle                          // Stop all operations
    CommandFcasResponse                  // Respond to FCAS dispatch
)

type Command struct {
    Type       CommandType   // What to do
    Power      float64       // MW (absolute value)
    TargetSoC  float64       // 0-100% (for charging, optional)
    Duration   time.Duration // Max duration (0 = indefinite)
}
```

#### Command Validation Rules

**Charge Command**:
```go
Type == CommandCharge
Power > 0
TargetSoC > currentSoC (if specified)
```

**Discharge Command**:
```go
Type == CommandDischarge
Power > 0
currentSoC > 0 (can't discharge empty battery)
```

**Idle Command**:
```go
Type == CommandIdle
Power == 0
```

### 2.4 Mock Adapter Behaviors

#### TeslaLike Adapter

**Characteristics**:
- **Ramp Rate**: 5 MW/s (aggressive)
- **Efficiency**: 95% round-trip
- **Temperature Rise**: 0.5°C per MW charging
- **CustomAttributes**:
  ```json
  {
    "vendor": "Tesla",
    "model": "Megapack",
    "firmwareVersion": "1.2.3",
    "cellCount": 4320,
    "thermalZones": 3
  }
  ```

**State Simulation**:
```
Every tick (100ms):
  if charging:
    SoC += (Power * efficiency * tickDuration) / capacity
    Temperature += 0.05 * abs(Power)
  if discharging:
    SoC -= (Power * tickDuration) / capacity
    Temperature += 0.05 * abs(Power)
  if idle:
    Temperature -= 0.1 (cooling)
```

#### BYDLike Adapter

**Characteristics**:
- **Ramp Rate**: 3 MW/s (conservative)
- **Efficiency**: 92% round-trip
- **Temperature Rise**: 0.7°C per MW charging (higher thermal loss)
- **CustomAttributes**:
  ```json
  {
    "vendor": "BYD",
    "model": "Battery-Box",
    "firmwareVersion": "2.0.1",
    "cellCount": 3840,
    "batteryChemistry": "LFP"
  }
  ```

**State Simulation**:
```
Every tick (100ms):
  if charging:
    SoC += (Power * efficiency * tickDuration) / capacity
    Temperature += 0.07 * abs(Power)
  if discharging:
    SoC -= (Power * tickDuration) / capacity
    Temperature += 0.07 * abs(Power)
  if idle:
    Temperature -= 0.08 (slower cooling)
```

### 2.5 Adapter Implementation Requirements

Both adapters must:
1. **Thread-Safe**: Use sync.RWMutex for state access
2. **Real-Time Simulation**: Background goroutine updating state
3. **Ramp Rate Enforcement**: Gradual power changes (not instant)
4. **Temperature Physics**: Realistic heating/cooling
5. **Boundary Checks**: Prevent SoC < 0 or SoC > 100

---

## 3. Shared Domain Concepts

### 3.1 CustomAttributes Pattern

**Purpose**: Store vendor-specific data without polluting core domain

**Storage**: JSONB column in PostgreSQL (Telemetry Service)

**Usage**:
```go
// Tesla-specific attribute
customAttrs := map[string]interface{}{
    "cellVoltageMin": 3.2,
    "cellVoltageMax": 4.1,
    "thermalZone1Temp": 28.5,
}

// BYD-specific attribute
customAttrs := map[string]interface{}{
    "moduleTempAvg": 30.2,
    "bmsVersion": "2.0.1",
    "lfpCycleCount": 1234,
}
```

**Anti-Corruption Layer**: Device Interface translates hardware protocols → domain events

### 3.2 OperationState State Machine

```
IDLE ──charge──> CHARGING ──target reached──> IDLE
IDLE ──discharge──> DISCHARGING ──stop condition──> IDLE
IDLE ──fcas dispatch──> FCAS ──response complete──> IDLE

CHARGING ──stop command──> IDLE
DISCHARGING ──stop command──> IDLE
FCAS ──stop command──> IDLE
```

**Invariant**: No direct transitions between CHARGING ↔ DISCHARGING (must go through IDLE)

### 3.3 Power Sign Convention

**Critical Convention** (matches EVENTS.md):
- **Positive Power**: Discharging (battery → grid)
- **Negative Power**: Charging (grid → battery)
- **Zero Power**: Idle

**Examples**:
- `Power = 25.0` → Discharging at 25 MW
- `Power = -15.0` → Charging at 15 MW
- `Power = 0.0` → Idle

---

## 4. Validation Coverage Summary

| Validation | Telemetry Service | Device Interface | Priority |
|------------|-------------------|------------------|----------|
| SoC Range (0-100) | ✅ | ✅ | Critical |
| Temperature Range | ✅ | ✅ | High |
| OperationState Enum | ✅ | ✅ | Critical |
| Power Consistency | ✅ | ✅ | Critical |
| Required Fields | ✅ | ✅ | Critical |
| Timestamp Validity | ✅ | N/A | Medium |
| Ramp Rate | N/A | ✅ | High |
| Efficiency Bounds | N/A | ✅ | Medium |

---

## 5. Testing Strategy

### Unit Tests (TDD)

**Domain Layer**:
- `TestNewBatteryState_ValidInput` - All fields within ranges
- `TestNewBatteryState_InvalidSoC` - Out of bounds (0-100)
- `TestNewBatteryState_InvalidTemperature` - Out of operating range
- `TestBatteryState_PowerConsistency` - Power sign matches state
- `TestBatteryState_OperationStateTransitions` - State machine

**Adapter Layer**:
- `TestTeslaLike_GetState` - Returns current state
- `TestTeslaLike_SendCommand_Charge` - SoC increases over time
- `TestTeslaLike_SendCommand_Discharge` - SoC decreases
- `TestTeslaLike_RampRate` - Power changes gradually (5 MW/s)
- `TestBYDLike_RampRate` - Different ramp rate (3 MW/s)

### Integration Tests

- End-to-end: Command → Adapter → State Change → Event
- Performance: 1 Hz publishing stability
- Concurrency: Multiple adapters running simultaneously

---

## 6. Related Specifications

- [M5 Overview](./M5-OVERVIEW.md) - Big picture and architecture
- [M5 API Spec](./M5-API-SPEC.md) - REST endpoints and event schemas
- [M5 Checklist](./M5-CHECKLIST.md) - Implementation tasks
- [EVENTS.md](../../EVENTS.md) - BatteryStateChanged event (lines 459-483)

---

**Domain Complexity**: Moderate

**Validation Rules**: 8 critical rules enforced

**Test Coverage Target**: >85% domain layer, >80% overall
