# Telemetry Service

Real-time battery state monitoring and historical data storage service for the Battery Optimization System.

## Purpose

The Telemetry Service provides:
- Real-time battery state monitoring (SoC, power, temperature, voltage, current)
- Historical telemetry data storage with time-series optimization
- REST API for querying current and historical battery states
- Event publishing for `BatteryStateChanged` (placeholder for future 1 Hz publishing)

## Architecture

### Hexagonal Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP API Layer                        │
│  (Adapters/HTTP: DTOs, Handlers, Routes)                │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                   Domain Layer                           │
│  (Core Business Logic: BatteryState, Validation)        │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│                Repository Port                           │
│  (Interface: TelemetryRepository)                       │
└────────────────┬────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────┐
│            PostgreSQL Adapter                            │
│  (Implementation: Time-series queries, Migrations)      │
└─────────────────────────────────────────────────────────┘
```

### Database Schema

**Table: `battery_states`**

```sql
CREATE TABLE battery_states (
    id VARCHAR(36) PRIMARY KEY,
    battery_id VARCHAR(36) NOT NULL,
    soc NUMERIC(5,2) NOT NULL CHECK (soc >= 0 AND soc <= 100),
    power NUMERIC(10,2) NOT NULL,
    temperature NUMERIC(5,2) NOT NULL CHECK (temperature >= -20 AND temperature <= 60),
    voltage NUMERIC(10,2) NOT NULL,
    current NUMERIC(10,2) NOT NULL,
    operation_state VARCHAR(20) NOT NULL CHECK (operation_state IN ('IDLE', 'CHARGING', 'DISCHARGING', 'FCAS')),
    custom_attributes JSONB,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Indexes** (Time-Series Optimized):
- `idx_battery_states_battery_id_timestamp` - Primary query pattern (get latest state for battery)
- `idx_battery_states_timestamp` - Historical queries by timestamp
- `idx_battery_states_operation_state` - Filter by operation state
- `idx_battery_states_custom_attributes` - JSONB GIN index for custom attributes

## REST API

### Endpoints

#### Get Current State

```http
GET /api/v1/telemetry/:batteryId/current
```

**Response** (200 OK):
```json
{
  "id": "state-uuid",
  "battery_id": "battery-123",
  "soc": 75.5,
  "power": 25.0,
  "temperature": 28.5,
  "voltage": 825.0,
  "current": 30303.0,
  "operation_state": "DISCHARGING",
  "custom_attributes": {
    "vendor": "Tesla",
    "model": "Megapack"
  },
  "timestamp": "2025-12-30T14:35:00Z"
}
```

**Error** (404 Not Found):
```json
{
  "code": "NOT_FOUND",
  "message": "No telemetry data found for battery battery-123"
}
```

#### Get Historical Data

```http
GET /api/v1/telemetry/:batteryId/history?start_time=2025-12-30T10:00:00Z&end_time=2025-12-30T14:00:00Z&limit=100
```

**Query Parameters**:
- `start_time` (required): ISO 8601 timestamp
- `end_time` (required): ISO 8601 timestamp
- `limit` (optional): Max results (default: 1000)
- `offset` (optional): Pagination offset (default: 0)

**Response** (200 OK):
```json
{
  "battery_id": "battery-123",
  "states": [
    {
      "id": "state-1",
      "soc": 50.0,
      "power": 0.0,
      "temperature": 25.0,
      "voltage": 800.0,
      "current": 0.0,
      "operation_state": "IDLE",
      "timestamp": "2025-12-30T10:00:00Z"
    },
    ...
  ],
  "total": 240
}
```

## Event Publishing

### BatteryStateChanged (v1)

**Subject**: `battery.state.changed.v1`
**Frequency**: Placeholder for future 1 Hz publishing

**Payload**:
```json
{
  "battery_id": "battery-123",
  "soc": 75.5,
  "power": 25.0,
  "operation_state": "DISCHARGING",
  "temperature": 28.5,
  "voltage": 825.0,
  "current": 30303.0,
  "custom_attributes": {
    "vendor": "Tesla"
  },
  "timestamp": "2025-12-30T14:35:00Z",
  "event_version": "v1"
}
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable` | PostgreSQL connection string |
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `PORT` | `8082` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level |

## Development Commands

### Build
```bash
go build -o bin/telemetry cmd/server/main.go
```

### Run
```bash
PORT=8082 ./bin/telemetry
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
- Domain: 100.0%
- PostgreSQL Adapter: 81.5%
- HTTP Adapter: 76.7%

### Database Setup

The service automatically runs migrations on startup. To manually verify the database:

```bash
docker exec -it telemetry-db psql -U telemetry_user -d telemetry

# List tables
\dt

# Query states
SELECT battery_id, soc, power, operation_state, timestamp
FROM battery_states
ORDER BY timestamp DESC
LIMIT 10;
```

## Dependencies

- **PostgreSQL 18**: Time-series data storage
- **NATS 2.10**: Event bus for publishing `BatteryStateChanged`
- **gorilla/mux**: HTTP routing
- **golang-migrate/migrate**: Database migrations
- **lib/pq**: PostgreSQL driver

## Integration

### With Device Interface Service

Device Interface publishes `BatteryStateChanged` events that Telemetry Service subscribes to (future enhancement for 1 Hz telemetry).

### With Bidding Service

Bidding Service queries current battery state via REST API to make bidding decisions.

### With Operator Dashboard

Dashboard queries historical telemetry data for visualization and analysis.

## Production Considerations

### High-Frequency Data (1 Hz Publishing)

- **86,400 events/day** per battery at 1 Hz
- Consider partitioning `battery_states` table by timestamp (daily/weekly)
- Implement data retention policy (e.g., keep raw 1 Hz data for 7 days, downsample to 1-minute aggregates for historical analysis)
- Monitor NATS throughput and backpressure

### Database Tuning

- Increase `shared_buffers` for time-series workload
- Tune `work_mem` for sorting large result sets
- Consider TimescaleDB extension for advanced time-series features

### API Rate Limiting

- Implement rate limiting on historical queries to prevent database overload
- Cache recent states in Redis for high-traffic `/current` endpoint

## License

Proprietary - Battery Optimization System
