# Device Interface Service

Hardware abstraction layer for battery energy storage systems (BESS) providing vendor-agnostic control and monitoring.

## Purpose

The Device Interface Service provides:
- **Hardware abstraction**: Uniform interface for different battery vendors (Tesla, BYD, etc.)
- **Event-driven command handling**: Reacts to charging/discharging commands via NATS
- **Mock adapters**: Realistic battery simulation for development without physical hardware
- **Lifecycle event publishing**: Publishes operational events (ChargingStarted, DischargingStarted, etc.)

## Architecture

### BatteryAdapter Pattern

```
┌──────────────────────────────────────────────────────┐
│              Command Handler Service                  │
│  (Subscribes to command events, publishes lifecycle)  │
└────────────────┬─────────────────────────────────────┘
                 │
┌────────────────▼─────────────────────────────────────┐
│            BatteryAdapter Interface                   │
│  (Domain: GetState, SendCommand, GetCustomAttributes)│
└────────────────┬─────────────────────────────────────┘
                 │
       ┌─────────┴──────────┐
       ▼                    ▼
┌──────────────┐    ┌──────────────┐
│  TeslaLike   │    │   BYDLike    │
│   Adapter    │    │   Adapter    │
│ (Mock/Real)  │    │ (Mock/Real)  │
└──────────────┘    └──────────────┘
```

### Event Flow

**Subscribed Events**:
1. `charging.command.issued.v1` → Starts charging
2. `discharging.command.issued.v1` → Starts discharging
3. `conflict.resolved.v1` → Resolves FCAS vs. arbitrage conflicts

**Published Events**:
1. `battery.connection.established.v1` → On startup
2. `charging.started.v1` → When charging begins
3. `charging.completed.v1` → When charging stops (future)
4. `discharging.started.v1` → When discharging begins
5. `discharging.completed.v1` → When discharging stops (future)

## BatteryAdapter Interface

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

### BatteryState

```go
type BatteryState struct {
    SoC            float64 // State of Charge (0-100%)
    Power          float64 // MW (positive=discharge, negative=charge)
    Temperature    float64 // °C
    Voltage        float64 // V
    Current        float64 // A
    OperationState string  // IDLE, CHARGING, DISCHARGING, FCAS
}
```

### Command

```go
type CommandType int

const (
    CommandCharge       CommandType = iota // Start charging
    CommandDischarge                       // Start discharging
    CommandIdle                            // Stop all operations
    CommandFcasResponse                    // Respond to FCAS dispatch
)

type Command struct {
    Type      CommandType
    Power     float64       // MW (absolute value)
    TargetSoC float64       // 0-100% (for charging, optional)
    Duration  time.Duration // Max duration (0 = indefinite)
}
```

## Mock Adapters

### TeslaLike Adapter

Simulates Tesla Megapack behavior with realistic characteristics:

**Characteristics**:
- **Ramp Rate**: 5 MW/s (aggressive)
- **Efficiency**: 95% round-trip
- **Temperature Rise**: 0.5°C per MW of power
- **Cooling Rate**: 0.1°C per 100ms when idle
- **Initial SoC**: 50%
- **Initial Temperature**: 25°C
- **Base Voltage**: 800V (varies with SoC: 760V-880V)

**Custom Attributes**:
```json
{
  "vendor": "Tesla",
  "model": "Megapack",
  "firmwareVersion": "1.2.3",
  "cellCount": 4320,
  "thermalZones": 3
}
```

**Simulation Details**:
- Background goroutine updates state every 100ms
- Power ramps gradually towards target (no instant changes)
- SoC increases during charging (accounting for efficiency)
- SoC decreases during discharging
- Temperature increases during operation, cools when idle
- SoC enforced within 0-100% boundaries
- Thread-safe state access with sync.RWMutex

### BYDLike Adapter (Future)

Planned characteristics:
- **Ramp Rate**: 3 MW/s (more conservative)
- **Efficiency**: 92% round-trip
- Different thermal characteristics

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `BATTERY_ID` | `battery-001` | Unique battery identifier |
| `ADAPTER_TYPE` | `TeslaLike` | Adapter type (TeslaLike, BYDLike) |
| `CAPACITY` | `200.0` | Battery capacity (MWh) |
| `MAX_POWER` | `100.0` | Maximum power (MW) |
| `LOG_LEVEL` | `info` | Logging level |

## Development Commands

### Build
```bash
go build -o bin/device-interface cmd/server/main.go
```

### Run
```bash
BATTERY_ID=battery-123 ADAPTER_TYPE=TeslaLike CAPACITY=200 MAX_POWER=100 ./bin/device-interface
```

### Test
```bash
go test ./...
```

### Test Coverage
```bash
go test -cover ./...
```

**Current Coverage**:
- Adapters: 78.5%
- Service Layer: 76.4%

## Event Examples

### Subscribing to Charging Commands

The service automatically subscribes to command events on startup:

```
2025/12/30 14:31:52 Subscribed to: charging.command.issued.v1
2025/12/30 14:31:52 Subscribed to: discharging.command.issued.v1
2025/12/30 14:31:52 Subscribed to: conflict.resolved.v1
2025/12/30 14:31:52 Event subscriptions active - ready to process commands
```

### Handling ChargingCommandIssued

**Incoming Event** (`charging.command.issued.v1`):
```json
{
  "battery_id": "battery-123",
  "command_id": "cmd-001",
  "power": 15.0,
  "target_soc": 80.0,
  "duration": 60,
  "decision_mode": "FULL_AUTO",
  "timestamp": "2025-12-30T14:35:00Z",
  "event_version": "v1"
}
```

**Published Event** (`charging.started.v1`):
```json
{
  "battery_id": "battery-123",
  "command_id": "cmd-001",
  "actual_power": -15.0,
  "target_soc": 80.0,
  "current_soc": 50.0,
  "timestamp": "2025-12-30T14:35:00.100Z",
  "event_version": "v1"
}
```

### Handling DischargingCommandIssued

**Incoming Event** (`discharging.command.issued.v1`):
```json
{
  "battery_id": "battery-123",
  "command_id": "cmd-002",
  "power": 25.0,
  "duration": 30,
  "stop_conditions": {
    "min_soc": 20.0,
    "price_threshold": 50.0
  },
  "decision_mode": "SEMI_AUTO",
  "approved_by": "operator-001",
  "timestamp": "2025-12-30T14:40:00Z",
  "event_version": "v1"
}
```

**Published Event** (`discharging.started.v1`):
```json
{
  "battery_id": "battery-123",
  "command_id": "cmd-002",
  "actual_power": 25.0,
  "current_soc": 75.5,
  "timestamp": "2025-12-30T14:40:00.100Z",
  "event_version": "v1"
}
```

### BatteryConnectionEstablished

Published on startup:

```json
{
  "battery_id": "battery-123",
  "adapter_type": "TeslaLike",
  "firmware_version": "1.2.3",
  "custom_attributes": {
    "vendor": "Tesla",
    "model": "Megapack",
    "cellCount": 4320,
    "thermalZones": 3
  },
  "timestamp": "2025-12-30T14:31:52Z",
  "event_version": "v1"
}
```

## Integration Testing

Use the event subscriber tool to monitor events:

```bash
# Terminal 1: Start event subscriber
cd tools/event-subscriber
go run main.go ">"

# Terminal 2: Start Device Interface Service
cd services/device-interface
BATTERY_ID=battery-123 ./bin/device-interface

# Terminal 3: Publish test command
cd tools/test-publisher
go run main.go
```

**Expected Flow**:
1. `battery.connection.established.v1` published on startup
2. `charging.command.issued.v1` received by service
3. `charging.started.v1` published by service
4. Battery state updates in real-time (check logs)

## Testing TeslaLike Adapter Behavior

### Ramp Rate Test

```bash
# Send 50 MW discharge command
# Power should ramp from 0 to 50 MW at 5 MW/s (10 seconds)

# After 1 second: Power ~5 MW
# After 5 seconds: Power ~25 MW
# After 10 seconds: Power ~50 MW (target reached)
```

### SoC Changes

```bash
# 200 MWh battery, 15 MW charging:
# Time to charge from 50% to 80% = 30% of 200 MWh = 60 MWh
# At 15 MW with 95% efficiency: 60 / (15 * 0.95) = 4.2 hours
```

### Temperature Simulation

```bash
# Charging at 15 MW:
# Temperature rise: 0.5°C/MW * 15 MW = 7.5°C total rise
# From 25°C → 32.5°C when stable

# When idle:
# Cooling: 0.1°C per 100ms → 1°C per second
# Returns to 25°C ambient
```

## Production Deployment

### Real Hardware Adapters

When integrating with actual battery hardware:

1. Implement `BatteryAdapter` interface for your vendor
2. Replace mock simulation with real hardware communication (Modbus, MQTT, proprietary APIs)
3. Add error handling for hardware failures
4. Implement health checks and heartbeat monitoring
5. Add retry logic for transient communication errors

### Safety Considerations

- **Rate limiting**: Prevent command spam that could damage hardware
- **Validation**: Verify commands are within battery specifications
- **Emergency stop**: Implement kill switch for critical failures
- **State verification**: Confirm hardware state matches commanded state
- **Logging**: Comprehensive audit trail of all commands

## Dependencies

- **NATS 2.10**: Event bus for command subscriptions and lifecycle publishing
- **pkg/events**: Shared event library

## License

Proprietary - Battery Optimization System
