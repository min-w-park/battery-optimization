# Event Subscriber Tool

A command-line utility for monitoring events published to NATS in real-time. This tool is useful for:
- Debugging event flows during development
- Verifying that services are publishing events correctly
- Understanding the event-driven architecture

## Features

- 🎨 **Color-coded output** for easy reading
- 📊 **Pretty-printed JSON** payloads
- 🔍 **Flexible subscription patterns** (wildcards supported)
- 📈 **Event counter** to track total events received
- ⚡ **Real-time monitoring** with timestamps
- 🛑 **Graceful shutdown** (Ctrl+C)

## Usage

### Subscribe to All Events (default)

```bash
go run main.go
```

This subscribes to `>` which matches all events from all services.

### Subscribe to Specific Events

```bash
# All battery events
go run main.go "battery.>"

# Only BatteryRegistered events
go run main.go "battery.registered.v1"

# All market events
go run main.go "market.>"

# All v1 events
go run main.go "*.*.v1"
```

### Custom NATS URL

```bash
NATS_URL="nats://custom-host:4222" go run main.go
```

## Example Output

```
=== Event Subscriber Tool ===
NATS URL: nats://localhost:4222
Subscription Pattern: >
Listening for events... (Press Ctrl+C to stop)

Subscribed successfully!

[10:30:15.234] Event #1
Subject: battery.registered.v1
Payload:
{
  "battery_id": "550e8400-e29b-41d4-a716-446655440000",
  "capacity": 100.0,
  "constraints": {
    "grid_compliance_level": "AS4777",
    "max_cycles": 5000,
    "max_soc": 0.9,
    "min_soc": 0.1,
    "operating_temp_max": 45.0,
    "operating_temp_min": -10.0,
    "warranty_eol": 0.8
  },
  "efficiency": 0.95,
  "event_version": "v1",
  "location": "NSW",
  "manufacturer": "Tesla",
  "max_power": 50.0,
  "ramp_rate": 10.0,
  "timestamp": "2025-12-30T10:30:15.234Z"
}
---

[10:30:20.567] Event #2
Subject: market.price.updated.v1
Payload:
{
  "demand": 8500.0,
  "event_version": "v1",
  "interval_start": "2025-12-30T10:30:00Z",
  "interval_type": "5MIN",
  "price": 75.50,
  "price_id": "660e9511-f39c-52e5-b827-557766551111",
  "published_at": "2025-12-30T10:30:15Z",
  "region": "NSW",
  "timestamp": "2025-12-30T10:30:20.567Z"
}
---

^C
Shutting down...
Total events received: 2
```

## NATS Subject Patterns

The tool supports NATS wildcard subscriptions:

- `*` - Matches a single token
  - Example: `battery.*.v1` matches `battery.registered.v1` but NOT `battery.state.changed.v1`
- `>` - Matches one or more tokens
  - Example: `battery.>` matches `battery.registered.v1` AND `battery.state.changed.v1`
  - Example: `>` matches ALL events

## Common Patterns

| Pattern | Description |
|---------|-------------|
| `>` | All events (default) |
| `battery.>` | All battery-related events |
| `market.>` | All market-related events |
| `battery.registered.v1` | Only BatteryRegistered events |
| `market.price.updated.v1` | Only MarketPriceUpdated events |
| `*.*.v1` | All v1 events from any service |

## Building

```bash
# Build binary
go build -o event-subscriber main.go

# Run binary
./event-subscriber
./event-subscriber "battery.>"
```

## Integration with Services

To see events from the services:

1. **Start NATS** (if not already running):
   ```bash
   docker-compose up -d nats
   ```

2. **Start the subscriber**:
   ```bash
   go run main.go
   ```

3. **Start services** (in separate terminals):
   ```bash
   # Asset Management Service (port 8080)
   cd services/asset-management
   go run cmd/server/main.go

   # Market Data Service (port 8081)
   cd services/market-data
   go run cmd/server/main.go
   ```

4. **Trigger events** by creating resources:
   ```bash
   # Create battery (triggers BatteryRegistered)
   curl -X POST http://localhost:8080/api/v1/batteries \
     -H "Content-Type: application/json" \
     -d '{...}'

   # Create market price (triggers MarketPriceUpdated)
   curl -X POST http://localhost:8081/prices \
     -H "Content-Type: application/json" \
     -d '{...}'
   ```

You should see events appear in the subscriber output in real-time!

## Troubleshooting

### "Failed to connect to NATS"
- Ensure NATS is running: `docker-compose ps nats`
- Check NATS is accessible: `curl http://localhost:8222/healthz`
- Verify NATS_URL environment variable

### No events appearing
- Verify services are publishing events (check service logs)
- Ensure subscription pattern matches event subjects
- Check that services successfully connected to NATS

### Event data truncated or malformed
- This tool displays raw JSON from NATS
- If events look wrong, check service publishing logic
- Verify event serialization in service code
