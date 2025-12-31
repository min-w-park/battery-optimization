# E2E Test Infrastructure

This directory contains end-to-end (E2E) test infrastructure for the battery optimization system. The tests validate complete workflows across all 5 microservices communicating via NATS events.

## Architecture

The E2E test infrastructure consists of three main components:

### 1. Helpers (`helpers.go`)

Utility functions for interacting with services:

**REST API Helpers**:
- `CreateBattery(baseURL, capacity, maxPower, rampRate, efficiency)` - Register battery via Asset Management
- `CreatePrice(baseURL, price, interval)` - Create market price via Market Data
- `GetBattery(baseURL, batteryID)` - Retrieve battery details

**NATS Helpers**:
- `PublishBatteryState(nc, batteryID, soc, power, state)` - Publish battery state event
- `SubscribeToEvents(nc, pattern)` - Subscribe to NATS events with wildcard
- `WaitForEvent(sub, subject, timeout)` - Wait for specific event
- `WaitForAnyEvent(sub, timeout)` - Wait for any event
- `AssertNoEvent(sub, timeout)` - Verify no event received
- `DrainEvents(sub)` - Clear pending events

**Connection Helpers**:
- `ConnectToNATS(url, retries, retryDelay)` - Connect to NATS with retry logic
- `WaitForHTTPService(baseURL, retries, retryDelay)` - Wait for service readiness

### 2. Assertions (`assertions.go`)

Event validation functions using `testify/assert`:

**Asset Management Events**:
- `AssertBatteryRegistered(t, msg, batteryID, capacity, maxPower)` - Validate battery.registered.v1

**Market Data Events**:
- `AssertMarketPriceUpdated(t, msg, price)` - Validate market.price.updated.v1

**Bidding Service Events**:
- `AssertChargingOpportunity(t, msg, batteryID, price, soc)` - Validate charging.opportunity.detected.v1
- `AssertChargingCommand(t, msg, batteryID, targetSoC, maxPower)` - Validate charging.command.issued.v1
- `AssertDischargingOpportunity(t, msg, batteryID, price, soc)` - Validate discharging.opportunity.detected.v1
- `AssertDischargingCommand(t, msg, batteryID, power)` - Validate discharging.command.issued.v1

**Device Interface Events**:
- `AssertChargingStarted(t, msg, batteryID)` - Validate charging.started.v1
- `AssertChargingCompleted(t, msg, batteryID)` - Validate charging.completed.v1
- `AssertDischargingStarted(t, msg, batteryID)` - Validate discharging.started.v1
- `AssertDischargingCompleted(t, msg, batteryID)` - Validate discharging.completed.v1

**Telemetry Events**:
- `AssertBatteryStateChanged(t, msg, batteryID, soc, state)` - Validate battery.state.changed.v1

**Utilities**:
- `GetEventField(msg, field)` - Extract field from event
- `PrintEvent(t, msg)` - Pretty-print event for debugging

### 3. Fixtures (`fixtures.go`)

Test data and constants:

**Service URLs**:
- `AssetManagementURL` - http://localhost:8080
- `MarketDataURL` - http://localhost:8081
- `TelemetryURL` - http://localhost:8082
- `DeviceInterfaceURL` - http://localhost:8083
- `NATSURL` - nats://localhost:4222

**Price Scenarios**:
- `VeryLowPrice` = $20/MWh
- `LowPrice` = $30/MWh (charging opportunity)
- `EdgeLowPrice` = $49/MWh (edge case)
- `MidPrice` = $75/MWh (no opportunity)
- `EdgeHighPrice` = $101/MWh (edge case)
- `HighPrice` = $150/MWh (discharging opportunity)
- `VeryHighPrice` = $200/MWh

**SoC Scenarios**:
- `VeryLowSoC` = 10%
- `LowSoC` = 30% (charging opportunity)
- `MidSoC` = 50%
- `HighSoC` = 80% (max charging target)
- `VeryHighSoC` = 95% (discharging opportunity)

**Battery States**:
- `IdleBatteryState()` - Mid SoC, idle
- `LowSoCBatteryState()` - Low SoC, ready for charging
- `HighSoCBatteryState()` - High SoC, ready for discharging
- `ChargingBatteryState()` - Currently charging
- `DischargingBatteryState()` - Currently discharging

**Battery Specs**:
- `DefaultBatterySpec()` - Standard 200 MWh battery
- `LargeBatterySpec()` - 500 MWh battery for high-power tests
- `SmallBatterySpec()` - 50 MWh battery for edge cases

**Test Scenarios**:
- `GetChargingScenario()` - Complete charging workflow
- `GetDischargingScenario()` - Complete discharging workflow
- `GetNoOpportunityScenario()` - No arbitrage opportunity
- `GetEdgeCaseChargingScenario()` - Edge case at $49/MWh
- `GetEdgeCaseDischargingScenario()` - Edge case at $101/MWh

## Usage Example

```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestChargingWorkflow(t *testing.T) {
    // 1. Connect to NATS
    nc, err := ConnectToNATS(NATSURL, MaxRetries, ServiceRetryDelay)
    require.NoError(t, err)
    defer nc.Close()

    // 2. Subscribe to all events
    sub, err := SubscribeToEvents(nc, ">")
    require.NoError(t, err)
    defer sub.Unsubscribe()

    // 3. Register battery
    batteryID, err := CreateBattery(AssetManagementURL, DefaultCapacity, DefaultMaxPower, DefaultRampRate, DefaultEfficiency)
    require.NoError(t, err)

    // 4. Wait for battery.registered.v1 event
    msg, err := WaitForEvent(sub, "battery.registered.v1", EventTimeout)
    require.NoError(t, err)
    AssertBatteryRegistered(t, msg, batteryID, DefaultCapacity, DefaultMaxPower)

    // 5. Create low price
    err = CreatePrice(MarketDataURL, LowPrice, DefaultInterval)
    require.NoError(t, err)

    // 6. Wait for market.price.updated.v1 event
    msg, err = WaitForEvent(sub, "market.price.updated.v1", EventTimeout)
    require.NoError(t, err)
    AssertMarketPriceUpdated(t, msg, LowPrice)

    // 7. Publish battery state (low SoC, idle)
    err = PublishBatteryState(nc, batteryID, LowSoC, 0.0, StateIdle)
    require.NoError(t, err)

    // 8. Wait for charging.opportunity.detected.v1 event
    msg, err = WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
    require.NoError(t, err)
    AssertChargingOpportunity(t, msg, batteryID, LowPrice, LowSoC)

    // 9. Wait for charging.command.issued.v1 event (if FULL_AUTO)
    msg, err = WaitForEvent(sub, "charging.command.issued.v1", EventTimeout)
    require.NoError(t, err)
    AssertChargingCommand(t, msg, batteryID, TargetSoCCharging, DefaultMaxPower)
}
```

## Running E2E Tests

### Prerequisites

1. **Start infrastructure**:
   ```bash
   docker-compose up -d
   ```

2. **Start all services** (in separate terminals):
   ```bash
   # Terminal 1: Asset Management
   cd services/asset-management
   go run cmd/server/main.go

   # Terminal 2: Market Data
   cd services/market-data
   go run cmd/server/main.go

   # Terminal 3: Telemetry
   cd services/telemetry
   go run cmd/server/main.go

   # Terminal 4: Device Interface
   cd services/device-interface
   go run cmd/server/main.go

   # Terminal 5: Bidding (with FULL_AUTO)
   cd services/bidding
   AUTOMATION_MODE=FULL_AUTO go run cmd/server/main.go
   ```

3. **Wait for services to be ready** (all health checks passing)

### Run Tests

```bash
# Run all E2E tests
cd tests/e2e
go test -v -timeout 30s

# Run specific test
go test -v -run TestChargingWorkflow

# Run with race detector
go test -v -race

# Run with coverage
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Organization

E2E tests are organized into 6 categories:

### Category 1: Happy Path Workflows (Critical 🔴)
1. **TestChargingWorkflow** - Complete charging flow
2. **TestDischargingWorkflow** - Complete discharging flow
3. **TestMultipleBatteriesWorkflow** - Multiple batteries simultaneously
4. **TestPriceChangeTriggersDecision** - Price update triggers new decision

### Category 2: Edge Cases (High 🟡)
5. **TestEdgeCaseChargingThreshold** - Price exactly at $50/MWh
6. **TestEdgeCaseDischargingThreshold** - Price exactly at $100/MWh
7. **TestEdgeCaseSoCBoundaries** - SoC at 30% and 80%
8. **TestNoOpportunityMidRange** - Mid-range price and SoC

### Category 3: Error Handling (High 🟡)
9. **TestInvalidBatteryState** - Non-IDLE state
10. **TestServiceUnavailable** - Handle service failures
11. **TestEventPublishFailure** - NATS publish failures

### Category 4: Concurrency & Race Conditions (Medium 🟢)
12. **TestConcurrentPriceUpdates** - Rapid price changes
13. **TestConcurrentBatteryStateUpdates** - Multiple batteries updating
14. **TestRaceConditionStateCache** - Cache consistency

### Category 5: State Management & Event Ordering (Medium 🟢)
15. **TestEventOrdering** - Out-of-order events
16. **TestStateRecovery** - Service restart
17. **TestCacheConsistency** - Cache matches events

### Category 6: Complete End-to-End Workflows (Critical 🔴)
18. **TestCompleteArbitrageWorkflow** - Full cycle (register → charge → discharge)
19. **TestDayInTheLife** - 24-hour simulation

### Automation Mode Testing
20. **TestAutomationModes** - MANUAL vs FULL_AUTO behavior

## Debugging

**Enable event logging**:
```go
func TestChargingWorkflow(t *testing.T) {
    // ... setup ...

    msg, err := WaitForEvent(sub, "charging.opportunity.detected.v1", EventTimeout)
    require.NoError(t, err)

    // Print event for debugging
    PrintEvent(t, msg)

    AssertChargingOpportunity(t, msg, batteryID, LowPrice, LowSoC)
}
```

**Use event subscriber tool** to monitor events during tests:
```bash
cd tools/event-subscriber
go run main.go ">"
```

**Check NATS connection**:
```bash
curl http://localhost:8222/connz
curl http://localhost:8222/varz
```

## Best Practices

1. **Always drain events** before starting a new test to avoid leftover events:
   ```go
   DrainEvents(sub)
   ```

2. **Use appropriate timeouts**:
   - `EventTimeout` (3s) for expected events
   - `NoEventTimeout` (1s) for verifying no event
   - `ServiceStartTimeout` (30s) for service readiness

3. **Clean up resources**:
   ```go
   defer nc.Close()
   defer sub.Unsubscribe()
   ```

4. **Test isolation**: Each test should be independent and not rely on state from previous tests

5. **Use fixtures**: Prefer predefined scenarios from `fixtures.go` over hardcoded values

6. **Validate all event fields**: Don't just check a few fields - validate the complete event structure

## Troubleshooting

**Problem**: Tests timeout waiting for events
- **Solution**: Check that all services are running and healthy
- **Solution**: Verify NATS connection: `curl http://localhost:8222/healthz`
- **Solution**: Check service logs for errors

**Problem**: Events arrive out of order
- **Solution**: Use `WaitForEvent()` to wait for specific events by subject
- **Solution**: Don't assume sequential event delivery

**Problem**: Race conditions in cache tests
- **Solution**: Use `go test -race` to detect race conditions
- **Solution**: Add delays between concurrent operations

**Problem**: Flaky tests
- **Solution**: Increase timeouts in `fixtures.go`
- **Solution**: Add retry logic for service connections
- **Solution**: Ensure services are fully initialized before tests start

## Contributing

When adding new E2E tests:

1. Add helper functions to `helpers.go` if needed
2. Add assertion functions to `assertions.go` for new events
3. Add test data to `fixtures.go` for reusability
4. Follow the existing test structure (setup → action → assert)
5. Add documentation to this README
6. Ensure tests pass with `-race` flag
7. Target >80% code coverage for critical workflows
