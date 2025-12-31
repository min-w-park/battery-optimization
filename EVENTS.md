# Domain Events

## Overview

This document catalogs all domain events in the battery optimization system. Events represent things that have happened in the past and are the primary mechanism for communication between services.

## Event Design Principles

1. **Past tense naming**: Events describe what has happened (e.g., `BatteryRegistered`, not `RegisterBattery`)
2. **Immutable**: Once published, events cannot be changed
3. **Self-contained**: Events should contain all necessary data for consumers
4. **Versioned**: Support backward compatibility through versioning
5. **Business-focused**: Use domain language, not technical jargon

## Implementation Status

Events are organized into categories with implementation status tracked below:

### ✅ Implemented Events (13/25)

**Initialization (2/5)**:
- ✅ **BatteryRegistered** - Asset Management Service (M2)
- ✅ **BatteryConnectionEstablished** - Device Interface Service (M5)
- ⏳ BatteryTestStarted - Planned
- ⏳ BatteryTestCompleted - Planned
- ⏳ BatteryReadyToOperate - Planned

**Market Data (1/3)**:
- ✅ **MarketPriceUpdated** - Market Data Service (M3)
- ⏳ AemoPriceForecastReceived - Future work
- ⏳ InternalPriceForecastGenerated - Future work

**Charging Operations (4/7)**:
- ✅ **ChargingOpportunityDetected** - Bidding Service (M6)
- ✅ **ChargingCommandIssued** - Bidding Service (M6)
- ✅ **ChargingStarted** - Device Interface Service (M5)
- ✅ **ChargingCompleted** - Device Interface Service (M5)
- ⏳ ConflictDetected - Planned
- ⏳ EconomicsCalculationRequested - Future (Economics Service)
- ⏳ EconomicsCalculated - Future (Economics Service)

**Discharging Operations (4/4)**:
- ✅ **DischargingOpportunityDetected** - Bidding Service (M6)
- ✅ **DischargingCommandIssued** - Bidding Service (M6)
- ✅ **DischargingStarted** - Device Interface Service (M5)
- ✅ **DischargingCompleted** - Device Interface Service (M5)

**FCAS (0/4)**:
- ⏳ FcasContractStarted - Future work
- ⏳ FcasContractEnded - Future work
- ⏳ FcasDispatchReceived - Future work
- ⏳ FcasDispatchCompleted - Future work

**Shared Events (2/2)**:
- ✅ **BatteryStateChanged** - Telemetry Service (M5)
- ✅ **ConflictResolved** - Placeholder implementation (M5)

### Event Legend
- ✅ **Implemented**: Event struct defined in `pkg/events`, published and/or subscribed by at least one service
- ⏳ **Planned**: Event schema documented but not yet implemented
- 🔮 **Future**: Requires additional services or external integrations

### Implementation by Service

| Service | Events Published | Events Subscribed |
|---------|-----------------|-------------------|
| **Asset Management** (M2) | BatteryRegistered | None |
| **Market Data** (M3) | MarketPriceUpdated | None |
| **Telemetry** (M5) | BatteryStateChanged | BatteryConnectionEstablished |
| **Device Interface** (M5) | BatteryConnectionEstablished, ChargingStarted, ChargingCompleted, DischargingStarted, DischargingCompleted, ConflictResolved | ChargingCommandIssued, DischargingCommandIssued |
| **Bidding** (M6) | ChargingOpportunityDetected, DischargingOpportunityDetected, ChargingCommandIssued, DischargingCommandIssued | BatteryRegistered, MarketPriceUpdated, BatteryStateChanged |
| **Economics** (Future) | EconomicsCalculated, ConflictResolutionSuggested | ChargingOpportunityDetected, DischargingOpportunityDetected, ConflictDetected |

---

## Event Catalog

### 1. Battery Initialization Events

Events that occur when a new battery is registered and prepared for operation.

#### 1.1 BatteryRegistered
**Publisher**: Asset Management Service  
**Trigger**: Operator completes battery specification input via web UI  
**Subscribers**: Telemetry Service, Device Interface Service

**Data**:
```json
{
  "batteryId": "string",
  "specs": {
    "capacity": "number (MWh)",
    "maxPower": "number (MW)",
    "rampRate": "number (MW/min)",
    "efficiency": "number (0-1)",
    "temperatureConstraints": {
      "min": "number (°C)",
      "max": "number (°C)"
    }
  },
  "constraints": {
    "warrantyEOL": "number (0-1, e.g., 0.7 for 70%)",
    "maxCycles": "number",
    "safetyLimits": "object (UL9540A, etc.)",
    "gridCompliance": ["string (FCAS, FFR, etc.)"]
  },
  "location": "string",
  "manufacturer": "string",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Battery specs must pass validation
- Capacity, power, and ramp rate are mandatory
- Grid compliance requirements based on market regulations (AEMO for Australia)

---

#### 1.2 BatteryConnectionEstablished
**Publisher**: Device Interface Service  
**Trigger**: Successful communication link with battery hardware  
**Subscribers**: Telemetry Service, Asset Management Service

**Data**:
```json
{
  "batteryId": "string",
  "protocolType": "string (Modbus, CAN, etc.)",
  "connectionStatus": "CONNECTED",
  "timestamp": "ISO 8601"
}
```

---

#### 1.3 BatteryTestStarted
**Publisher**: Device Interface Service  
**Trigger**: Operator initiates battery health check test (manual by default, configurable for automation)  
**Subscribers**: Telemetry Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "testType": "INITIAL_HEALTH_CHECK",
  "testParameters": {
    "chargePower": "number (MW)",
    "dischargePower": "number (MW)",
    "duration": "number (minutes)"
  },
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Default: Manual trigger to avoid affecting other equipment
- Small capacity test (e.g., 1MW for 5 minutes)
- Tests charge/discharge response and ramp rate

---

#### 1.4 BatteryTestCompleted
**Publisher**: Device Interface Service  
**Trigger**: Health check test finishes  
**Subscribers**: Asset Management Service, Alerting Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "testType": "INITIAL_HEALTH_CHECK",
  "status": "PASSED | FAILED",
  "results": {
    "chargeResponse": "object",
    "dischargeResponse": "object",
    "rampRateActual": "number (MW/min)",
    "issues": ["string (if any)"]
  },
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- If FAILED: Alert operations team with detailed results
- If PASSED: Proceed to BatteryReadyToOperate

---

#### 1.5 BatteryReadyToOperate
**Publisher**: Asset Management Service  
**Trigger**: Test passed and all initialization complete  
**Subscribers**: Bidding Service, Market Service, Telemetry Service

**Data**:
```json
{
  "batteryId": "string",
  "operationalStatus": "READY",
  "availableCapacity": "number (MWh)",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Battery is now eligible for bidding
- Telemetry monitoring becomes active
- Can participate in market programs

---

### 2. Market Data Events

Events related to energy market pricing and forecasts.

#### 2.1 AemoPriceForecastReceived
**Publisher**: Market Data Service  
**Trigger**: AEMO publishes price forecast (5-minute or 30-minute pre-dispatch)  
**Subscribers**: Bidding Service, Economics Service

**Data**:
```json
{
  "region": "string (NSW, VIC, etc.)",
  "forecastType": "5MIN_PREDISPATCH | 30MIN_PREDISPATCH",
  "prices": [
    {
      "interval": "ISO 8601",
      "price": "number ($/MWh)",
      "demand": "number (MW)"
    }
  ],
  "publishedAt": "ISO 8601"
}
```

**Business Logic**:
- Updates every 5 minutes (5-min predispatch)
- 30-minute intervals for longer-term forecasts
- Source: AEMO official data

---

#### 2.2 InternalPriceForecastGenerated
**Publisher**: Market Data Service (ML/forecasting module)  
**Trigger**: Internal forecasting model completes prediction  
**Subscribers**: Bidding Service, Economics Service

**Data**:
```json
{
  "region": "string",
  "forecastHorizon": "string (24H, 7D, etc.)",
  "model": "string (ML model version)",
  "prices": [
    {
      "interval": "ISO 8601",
      "price": "number ($/MWh)",
      "confidence": "number (0-1)"
    }
  ],
  "generatedAt": "ISO 8601"
}
```

**Business Logic**:
- Complements AEMO forecasts with longer-term predictions
- Used for strategic planning
- Confidence level indicates prediction reliability

---

#### 2.3 MarketPriceChangedSignificantly
**Publisher**: Market Data Service  
**Trigger**: Price change exceeds threshold (conditional event)  
**Subscribers**: Bidding Service, Alerting Service, Operator Dashboard

**Data**:
```json
{
  "region": "string",
  "previousPrice": "number ($/MWh)",
  "currentPrice": "number ($/MWh)",
  "changePercent": "number",
  "threshold": "number (e.g., $100/MWh)",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Only published if:
  - Price change > 10%, OR
  - Price exceeds absolute threshold (e.g., $100/MWh)
- Triggers immediate bidding strategy review

---

### 3. Operational Events (Charging Scenario)

Events during normal charging operations.

#### 3.1 ChargingOpportunityDetected
**Publisher**: Bidding Service  
**Trigger**: Analysis identifies profitable charging opportunity  
**Subscribers**: Economics Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "currentSoC": "number (0-100)",
  "targetSoC": "number (0-100)",
  "currentPrice": "number ($/MWh)",
  "expectedPeakPrice": "number ($/MWh)",
  "expectedProfit": "number ($)",
  "recommendedAction": "CHARGE",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Current price is low (below threshold)
- Battery has charging capacity (SoC < 80%)
- Future price forecast shows profit opportunity
- Economics calculation shows positive ROI

---

#### 3.2 ChargingCommandIssued
**Publisher**: Bidding Service (if auto mode) or Operator Service (if manual/semi-auto)  
**Trigger**: Charging decision made (operator approval or automatic)  
**Subscribers**: Device Interface Service, Conflict Resolver

**Data**:
```json
{
  "batteryId": "string",
  "commandId": "string (UUID)",
  "power": "number (MW)",
  "targetSoC": "number (0-100)",
  "duration": "number (minutes)",
  "decisionMode": "MANUAL | SEMI_AUTO | FULL_AUTO",
  "approvedBy": "string (operator ID if manual)",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Bidding mode determines approval flow:
  - MANUAL: Operator initiates
  - SEMI_AUTO: AI recommends, operator approves
  - FULL_AUTO: System executes within predefined limits

---

#### 3.3 ConflictDetected
**Publisher**: Bidding Service or Conflict Resolver Service  
**Trigger**: New command conflicts with existing contract/operation  
**Subscribers**: Economics Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "conflictId": "string (UUID)",
  "currentState": {
    "operation": "FCAS_RAISE | FCAS_LOWER | DISCHARGING | IDLE",
    "contract": {
      "type": "string",
      "obligationLevel": "HIGH | MEDIUM | LOW",
      "penalty": "number ($)",
      "endTime": "ISO 8601"
    }
  },
  "conflictingCommand": {
    "type": "CHARGING | DISCHARGING",
    "power": "number (MW)"
  },
  "severity": "HIGH | MEDIUM | LOW",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Battery cannot charge and discharge simultaneously
- Existing FCAS contracts may prevent state change
- High severity if contract has high penalty

---

#### 3.4 EconomicsCalculationRequested
**Publisher**: Conflict Resolver Service  
**Trigger**: ConflictDetected received  
**Subscribers**: Economics Service

**Data**:
```json
{
  "calculationId": "string (UUID)",
  "conflictId": "string",
  "scenarios": [
    {
      "name": "CONTINUE_CURRENT",
      "description": "Maintain FCAS contract"
    },
    {
      "name": "SWITCH_TO_CHARGING",
      "description": "Cancel FCAS, start charging"
    }
  ],
  "timestamp": "ISO 8601"
}
```

---

#### 3.5 EconomicsCalculated
**Publisher**: Economics Service  
**Trigger**: Economic analysis complete  
**Subscribers**: Conflict Resolver Service, Operator Dashboard

**Data**:
```json
{
  "calculationId": "string",
  "conflictId": "string",
  "options": [
    {
      "scenario": "CONTINUE_CURRENT",
      "revenue": "number ($)",
      "penalty": 0,
      "netProfit": "number ($)",
      "risk": "LOW"
    },
    {
      "scenario": "SWITCH_TO_CHARGING",
      "revenue": "number ($)",
      "penalty": "number ($)",
      "netProfit": "number ($)",
      "risk": "HIGH"
    }
  ],
  "recommendation": "CONTINUE_CURRENT",
  "confidence": "number (0-1)",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Calculates revenue, penalties, net profit for each option
- Considers:
  - Contract obligations and penalties
  - Expected arbitrage profit
  - Risk level
- Provides recommendation with confidence score

---

#### 3.6 ConflictResolutionSuggested
**Publisher**: Economics Service  
**Trigger**: After EconomicsCalculated  
**Subscribers**: Operator Dashboard, Auto-Resolver (if enabled)

**Data**:
```json
{
  "conflictId": "string",
  "recommendedAction": "CONTINUE_CURRENT | SWITCH_TO_CHARGING",
  "reasoning": "string",
  "expectedOutcome": {
    "netProfit": "number ($)",
    "risk": "string"
  },
  "requiresApproval": "boolean",
  "timestamp": "ISO 8601"
}
```

---

#### 3.7 ConflictResolved
**Publisher**: Operator Service (manual) or Auto-Resolver Service (auto mode)  
**Trigger**: Final decision made  
**Subscribers**: Device Interface Service, Bidding Service

**Data**:
```json
{
  "conflictId": "string",
  "chosenAction": "CONTINUE_CURRENT | SWITCH_TO_CHARGING",
  "decisionMaker": "OPERATOR | AUTO",
  "operatorId": "string (if manual)",
  "timestamp": "ISO 8601"
}
```

---

#### 3.8 ChargingStarted
**Publisher**: Device Interface Service  
**Trigger**: Battery hardware begins charging  
**Subscribers**: Telemetry Service, Bidding Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "commandId": "string",
  "actualPower": "number (MW)",
  "startingSoC": "number (0-100)",
  "targetSoC": "number (0-100)",
  "timestamp": "ISO 8601"
}
```

---

#### 3.9 BatteryStateChanged
**Publisher**: Telemetry Service  
**Trigger**: Battery state update (every 1 second)  
**Subscribers**: Bidding Service, Device Interface Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "soc": "number (0-100)",
  "power": "number (MW, positive=discharge, negative=charge)",
  "temperature": "number (°C)",
  "voltage": "number (V)",
  "current": "number (A)",
  "operationState": "CHARGING | DISCHARGING | IDLE",
  "customAttributes": "object (manufacturer-specific data)",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Published every 1 second
- High event volume - consumers should filter based on needs
- CustomAttributes handle manufacturer-specific telemetry

---

#### 3.10 ChargingCompleted
**Publisher**: Device Interface Service  
**Trigger**: Target SoC reached or charging command duration elapsed  
**Subscribers**: Bidding Service, Telemetry Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "commandId": "string",
  "finalSoC": "number (0-100)",
  "energyCharged": "number (MWh)",
  "duration": "number (minutes)",
  "averagePower": "number (MW)",
  "completionReason": "TARGET_REACHED | DURATION_ELAPSED | MANUAL_STOP",
  "timestamp": "ISO 8601"
}
```

---

### 4. Operational Events (Discharging Scenario)

Events during normal discharging operations.

#### 4.1 DischargingOpportunityDetected
**Publisher**: Bidding Service  
**Trigger**: High price + sufficient SoC + market obligations checked  
**Subscribers**: Economics Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "currentSoC": "number (0-100)",
  "minimumSoC": "number (calculated based on obligations)",
  "currentPrice": "number ($/MWh)",
  "priceThreshold": "number ($/MWh)",
  "expectedRevenue": "number ($)",
  "obligations": [
    {
      "type": "FCAS_RAISE | FCAS_LOWER",
      "requiredBuffer": "number (MWh)",
      "dispatchProbability": "number (0-1)"
    }
  ],
  "futureOpportunities": {
    "nextPeakPrice": "number ($/MWh)",
    "nextPeakTime": "ISO 8601"
  },
  "recommendedAction": "DISCHARGE",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
1. **Price threshold (first rule)**: Current price > threshold
2. **Market obligations check**:
   - Active FCAS contracts → calculate required SoC buffer
   - High dispatch probability → increase buffer
   - Example: FCAS Lower requires ability to reduce discharge → need headroom
3. **SoC validation**: Current SoC - required buffer > minimum safe SoC
4. **Future opportunity consideration**:
   - Factory production schedule changes
   - Tomorrow's price forecast
   - "Will discharging now prevent better opportunity later?"
5. **Revenue calculation**: Expected profit from discharge
6. If all conditions met → detect opportunity

**Key differences from charging**:
- Must maintain SoC buffer for market obligations
- Consider dispatchability of FCAS programs
- Factory production schedule may require power reserve

---

#### 4.2 DischargingCommandIssued
**Publisher**: Bidding Service (auto) or Operator Service (manual/semi-auto)  
**Trigger**: Discharge decision made (operator approval or automatic)  
**Subscribers**: Device Interface Service, Conflict Resolver

**Data**:
```json
{
  "batteryId": "string",
  "commandId": "string (UUID)",
  "power": "number (MW)",
  "minimumSoC": "number (0-100)",
  "stopConditions": {
    "targetSoC": "number (optional, can stop before this)",
    "duration": "number (minutes, optional)",
    "priceThreshold": "number ($/MWh, optional - stop if price drops below)"
  },
  "decisionMode": "MANUAL | SEMI_AUTO | FULL_AUTO",
  "approvedBy": "string (operator ID if manual)",
  "timestamp": "ISO 8601"
}
```

**Key differences from charging**:
- `minimumSoC` instead of fixed `targetSoC` (can stop early)
- Multiple stop conditions (price, duration, SoC)
- Price threshold for early termination

---

#### 4.3 DischargingStarted
**Publisher**: Device Interface Service  
**Trigger**: Battery hardware begins discharging  
**Subscribers**: Telemetry Service, Bidding Service, Operator Dashboard

**Data**:
```json
{
  "batteryId": "string",
  "commandId": "string",
  "actualPower": "number (MW, positive for discharge)",
  "startingSoC": "number (0-100)",
  "minimumSoC": "number (0-100)",
  "expectedRevenue": "number ($)",
  "timestamp": "ISO 8601"
}
```

---

#### 4.4 DischargingCompleted
**Publisher**: Device Interface Service  
**Trigger**: Any stop condition met or manual stop  
**Subscribers**: Bidding Service, Telemetry Service, Operator Dashboard, Economics Service

**Data**:
```json
{
  "batteryId": "string",
  "commandId": "string",
  "finalSoC": "number (0-100)",
  "energyDischarged": "number (MWh)",
  "duration": "number (minutes)",
  "averagePrice": "number ($/MWh)",
  "actualRevenue": "number ($)",
  "completionReason": "TARGET_SOC_REACHED | DURATION_ELAPSED | PRICE_BELOW_THRESHOLD | FCAS_DISPATCH | OPERATIONAL_CONSTRAINT | MANUAL_STOP | EMERGENCY_STOP",
  "timestamp": "ISO 8601"
}
```

**Completion Reasons**:
- **TARGET_SOC_REACHED**: Minimum SoC reached
- **DURATION_ELAPSED**: Planned discharge duration completed
- **PRICE_BELOW_THRESHOLD**: Price dropped, no longer profitable
- **FCAS_DISPATCH**: FCAS dispatch signal received (must respond)
- **OPERATIONAL_CONSTRAINT**: Factory/site operational requirements changed
- **MANUAL_STOP**: Operator intervention
- **EMERGENCY_STOP**: Safety or equipment issue (future work)

---

### 5. FCAS (Frequency Control Ancillary Services) Events

Events related to grid frequency control services.

#### 5.1 FcasContractStarted
**Publisher**: Bidding Service  
**Trigger**: FCAS contract time block begins  
**Subscribers**: Device Interface Service, Telemetry Service, Economics Service

**Data**:
```json
{
  "contractId": "string (UUID)",
  "batteryId": "string",
  "serviceType": "RAISE | LOWER",
  "serviceCategory": "CONTINGENCY | REGULATION",
  "obligationLevel": "HIGH | MEDIUM | LOW",
  "requiredCapacity": "number (MW)",
  "requiredBuffer": "number (MWh SoC)",
  "penalty": "number ($ if breach)",
  "contractPrice": "number ($/MW/h)",
  "startTime": "ISO 8601",
  "endTime": "ISO 8601",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Contract represents availability commitment (not actual dispatch)
- Battery must maintain SoC buffer to meet potential dispatch
- RAISE: Ability to increase discharge or decrease charge
- LOWER: Ability to increase charge or decrease discharge
- Penalty applies if unable to respond to dispatch

**FCAS Types**:
- **Contingency**: High penalty, grid stability critical
- **Regulation**: Medium penalty, frequency regulation

---

#### 5.2 FcasContractEnded
**Publisher**: Bidding Service  
**Trigger**: FCAS contract time block completed  
**Subscribers**: Device Interface Service, Telemetry Service, Economics Service

**Data**:
```json
{
  "contractId": "string",
  "batteryId": "string",
  "serviceType": "RAISE | LOWER",
  "revenue": "number ($)",
  "dispatchCount": "number (how many times dispatched)",
  "complianceStatus": "COMPLIANT | BREACH",
  "breachPenalty": "number ($ if breach)",
  "timestamp": "ISO 8601"
}
```

---

#### 5.3 FcasDispatchReceived
**Publisher**: Market Data Service (monitoring AEMO API)  
**Trigger**: AEMO issues FCAS dispatch signal  
**Subscribers**: Bidding Service, Device Interface Service

**Data**:
```json
{
  "dispatchId": "string",
  "region": "string (NSW, VIC, QLD, SA, TAS)",
  "serviceType": "RAISE | LOWER",
  "frequencyDeviation": "number (Hz)",
  "requiredResponse": "number (MW)",
  "responseTime": "number (seconds, typically 1-6s)",
  "dispatchTime": "ISO 8601",
  "estimatedDuration": "number (seconds)",
  "timestamp": "ISO 8601"
}
```

**Business Logic**:
- Market-wide signal (not battery-specific)
- Bidding Service checks if battery has active FCAS contract
- If yes → must respond within specified time (typically 1 second)
- Response action depends on current battery state:

**RAISE Dispatch**:
- If charging → reduce charge power
- If idle → start discharge
- If discharging → increase discharge power

**LOWER Dispatch**:
- If discharging → reduce discharge power
- If idle → start charge
- If charging → increase charge power

---

#### 5.4 FcasDispatchCompleted
**Publisher**: Device Interface Service  
**Trigger**: Dispatch response completed or signal ended  
**Subscribers**: Bidding Service, Telemetry Service, Market Data Service

**Data**:
```json
{
  "dispatchId": "string",
  "batteryId": "string",
  "contractId": "string",
  "serviceType": "RAISE | LOWER",
  "requestedResponse": "number (MW)",
  "actualResponse": "number (MW)",
  "responseTime": "number (seconds)",
  "duration": "number (seconds)",
  "complianceStatus": "COMPLIANT | PARTIAL | FAILED",
  "energyProvided": "number (MWh)",
  "timestamp": "ISO 8601"
}
```

**Compliance**:
- **COMPLIANT**: Full response within required time
- **PARTIAL**: Partial response or delayed
- **FAILED**: Unable to respond (may incur penalty)

---

### 6. Future Work Events

These events are identified but deferred to future milestones:

#### 6.1 OperationalConstraintChanged
**Purpose**: Handle factory production schedule changes, site power demand spikes, maintenance requirements  
**Status**: Future work - external system integration required

**Example scenarios**:
- Factory emergency order → need to reserve power
- Site power demand spike → stop battery operations
- Equipment maintenance → safe shutdown required

---

#### 6.2 Emergency Events
**Purpose**: Temperature alerts, equipment failures, safety shutdowns  
**Status**: Future work

**Potential events**:
- BatteryTemperatureExceeded
- EquipmentFaultDetected
- EmergencyShutdownTriggered
- SafetySystemActivated

---

## Event Flow Diagrams

### Scenario 1: Battery Initialization

```
Operator
  ↓ (inputs specs)
BatteryRegistered
  ↓
Device Interface: BatteryConnectionEstablished
  ↓
Operator clicks "Test"
  ↓
BatteryTestStarted
  ↓
  ... test runs ...
  ↓
BatteryTestCompleted (PASSED)
  ↓
BatteryReadyToOperate
  ↓
Bidding Service adds to available batteries
```

---

### Scenario 2: Normal Charging Flow (No Conflict)

```
AEMO publishes forecast
  ↓
AemoPriceForecastReceived
  ↓
Bidding Service analyzes
  ↓
ChargingOpportunityDetected
  ↓
(If semi-auto) Operator approves
  ↓
ChargingCommandIssued
  ↓
No conflict detected
  ↓
ChargingStarted
  ↓
BatteryStateChanged (every 1s)
  ↓
  ... charging continues ...
  ↓
ChargingCompleted (TARGET_REACHED)
```

---

### Scenario 3: Charging with FCAS Conflict

```
ChargingCommandIssued
  ↓
Conflict Resolver checks current state
  ↓
ConflictDetected (FCAS Lower contract active)
  ↓
EconomicsCalculationRequested
  ↓
Economics Service calculates
  ↓
EconomicsCalculated
  ↓
ConflictResolutionSuggested
  ↓
(If manual mode) Operator decides
  ↓
ConflictResolved
  ↓
IF chose SWITCH_TO_CHARGING:
  Cancel FCAS → ChargingStarted
ELSE:
  Reject charging command → FCAS continues
```

---

### Scenario 4: Normal Discharging Flow

```
Market price rises
  ↓
MarketPriceChangedSignificantly
  ↓
Bidding Service analyzes
  - Check SoC > minimum
  - Check FCAS obligations
  - Check future opportunities
  ↓
DischargingOpportunityDetected
  ↓
(If semi-auto) Operator approves
  ↓
DischargingCommandIssued
  ↓
No conflict detected
  ↓
DischargingStarted
  ↓
BatteryStateChanged (every 1s)
  ↓
  ... discharging continues ...
  ↓
Price drops below threshold
  ↓
DischargingCompleted (PRICE_BELOW_THRESHOLD)
```

---

### Scenario 5: FCAS Contract with Dispatch

```
Contract time block starts
  ↓
FcasContractStarted (RAISE)
  ↓
Battery enters standby mode
  - Maintains SoC buffer
  - Monitors for dispatch signal
  ↓
  ... waiting ...
  ↓
Grid frequency drops
  ↓
AEMO issues dispatch
  ↓
FcasDispatchReceived (RAISE)
  ↓
Bidding Service validates contract
  ↓
Current state: Charging at 10MW
  ↓
Device Interface: Reduce charge to 5MW
  (or stop charging and start discharging)
  ↓
BatteryStateChanged (power adjusted)
  ↓
Grid frequency stabilizes
  ↓
FcasDispatchCompleted (COMPLIANT)
  ↓
Resume normal operations or continue standby
  ↓
Contract time ends
  ↓
FcasContractEnded
```

---

### Scenario 6: Discharging Interrupted by FCAS Dispatch

```
DischargingStarted (50MW discharge)
  ↓
Simultaneously: FcasContractStarted (LOWER)
  ↓
  ... discharging continues ...
  ↓
Grid frequency rises (over-frequency)
  ↓
FcasDispatchReceived (LOWER)
  ↓
CONFLICT: Currently discharging but need to reduce discharge
  ↓
Device Interface: Reduce discharge from 50MW to 30MW
  (or stop discharging entirely)
  ↓
DischargingCompleted (FCAS_DISPATCH)
  ↓
FcasDispatchCompleted (COMPLIANT)
  ↓
Check if should resume discharge
  - Is price still high?
  - Is FCAS contract still active?
  ↓
Decision: Resume or wait
```

---

### Scenario 7: Multiple Stop Conditions

```
DischargingStarted
  - Target: 30% SoC
  - Duration: 120 minutes
  - Price threshold: $80/MWh
  ↓
BatteryStateChanged (every 1s)
  ↓
Current: 45% SoC, 60 minutes elapsed
  ↓
Price suddenly drops to $75/MWh
  ↓
MarketPriceChangedSignificantly
  ↓
Bidding Service evaluates
  ↓
Price < threshold → stop discharging
  ↓
DischargingCompleted (PRICE_BELOW_THRESHOLD)
  
Note: Could have also stopped if:
- SoC reached 30%
- 120 minutes elapsed
- FCAS dispatch received
- Operator clicked stop
```

---

## Service Responsibilities

| Service | Events Published | Events Consumed |
|---------|-----------------|-----------------|
| **Asset Management** | BatteryRegistered, BatteryReadyToOperate | BatteryTestCompleted |
| **Device Interface** | BatteryConnectionEstablished, BatteryTestStarted, BatteryTestCompleted, ChargingStarted, ChargingCompleted, DischargingStarted, DischargingCompleted, FcasDispatchCompleted | ChargingCommandIssued, DischargingCommandIssued, ConflictResolved, FcasDispatchReceived |
| **Telemetry** | BatteryStateChanged | BatteryRegistered, BatteryConnectionEstablished, ChargingStarted, DischargingStarted |
| **Market Data** | AemoPriceForecastReceived, InternalPriceForecastGenerated, MarketPriceChangedSignificantly, FcasDispatchReceived | - |
| **Bidding** | ChargingOpportunityDetected, DischargingOpportunityDetected, ChargingCommandIssued, DischargingCommandIssued, ConflictDetected, FcasContractStarted, FcasContractEnded | BatteryReadyToOperate, MarketPriceChangedSignificantly, BatteryStateChanged, FcasDispatchReceived |
| **Economics** | EconomicsCalculated, ConflictResolutionSuggested | EconomicsCalculationRequested, ConflictDetected, ChargingOpportunityDetected, DischargingOpportunityDetected |
| **Conflict Resolver** | EconomicsCalculationRequested | ConflictDetected, ChargingCommandIssued, DischargingCommandIssued |

---

## Event Summary

### Total Events Defined: 25

**Initialization (5)**:
- BatteryRegistered
- BatteryConnectionEstablished
- BatteryTestStarted
- BatteryTestCompleted
- BatteryReadyToOperate

**Market Data (3)**:
- AemoPriceForecastReceived
- InternalPriceForecastGenerated
- MarketPriceChangedSignificantly

**Charging (7)**:
- ChargingOpportunityDetected
- ChargingCommandIssued
- ChargingStarted
- ChargingCompleted
- BatteryStateChanged (shared with discharging)
- ConflictDetected (shared)
- ConflictResolved (shared)

**Discharging (4)**:
- DischargingOpportunityDetected
- DischargingCommandIssued
- DischargingStarted
- DischargingCompleted

**Conflict Resolution (3)**:
- ConflictDetected
- EconomicsCalculationRequested
- EconomicsCalculated
- ConflictResolutionSuggested
- ConflictResolved

**FCAS (4)**:
- FcasContractStarted
- FcasContractEnded
- FcasDispatchReceived
- FcasDispatchCompleted

**Future Work (2+ categories)**:
- OperationalConstraintChanged
- Emergency events (temperature, faults, safety)

---

## Design Decisions

### 1. Event Frequency
- **BatteryStateChanged**: Every 1 second (FCAS requirement)
- **AemoPriceForecastReceived**: Every 5 minutes (AEMO schedule)
- **ConflictDetected**: On-demand (when conflicts occur)

### 2. Conditional Events
- **MarketPriceChangedSignificantly**: Only published when price change exceeds threshold
- Avoids unnecessary event spam while ensuring important changes are communicated

### 3. Conflict Resolution
- **Multi-stage process**: Detection → Economics → Suggestion → Resolution
- Allows for both automated and manual decision-making
- Economics calculation is a separate, reusable service

### 4. Automation Levels
- **MANUAL**: Operator makes all decisions
- **SEMI_AUTO**: AI recommends, operator approves
- **FULL_AUTO**: System executes within predefined safety limits
- Configurable per battery or per site

### 5. Anti-Corruption Layer
- **CustomAttributes** in BatteryStateChanged handles manufacturer-specific data
- Core domain model remains manufacturer-agnostic
- Device Interface Service translates between hardware protocols and domain events

---

## Versioning Strategy

### Event Schema Evolution
1. **Additive changes only** (new fields are optional)
2. **Version in event metadata**: Include `schemaVersion` field
3. **Backward compatibility**: Old consumers ignore unknown fields
4. **Deprecation notice**: 6 months before removing fields

Example:
```json
{
  "eventType": "BatteryStateChanged",
  "schemaVersion": "1.1.0",
  "data": {
    "batteryId": "...",
    "newField": "..." // Added in v1.1.0
  }
}
```

