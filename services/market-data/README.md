# Market Data Service

The Market Data Service manages AEMO (Australian Energy Market Operator) price forecasts and market data for the Australian National Electricity Market (NEM). It provides a REST API for storing and querying time-series market price data with support for 5-minute and 30-minute predispatch forecasts.

## Features

- **Time-Series Price Data**: Store and query market prices with interval-based forecasts
- **AEMO Integration**: Support for 5-minute and 30-minute predispatch intervals
- **Region Support**: All NEM regions (NSW, VIC, QLD, SA, TAS)
- **Validation**: 8 domain validation rules for price data integrity
- **Time-Range Queries**: Efficient querying with time range, region, and interval type filters
- **PostgreSQL**: Dedicated database with time-series optimized indexes

## Architecture

This service follows **Hexagonal Architecture** (Ports & Adapters pattern):

```
internal/
├── domain/           # Core business logic (MarketPrice aggregate)
├── ports/            # Repository interface definitions
└── adapters/
    ├── http/         # REST API handlers
    └── postgres/     # PostgreSQL repository implementation
```

## API Endpoints

### Create Market Price

```bash
POST /api/v1/prices
Content-Type: application/json

{
  "region": "NSW",
  "price": 85.50,
  "demand": 8200.0,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z"
}

Response: 201 Created
{
  "id": "129b6ded-ff74-495d-a52a-0daee54e85b0",
  "region": "NSW",
  "price": 85.5,
  "demand": 8200,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z",
  "created_at": "2025-12-29T15:26:35Z"
}
```

### Get Market Price by ID

```bash
GET /api/v1/prices/{id}

Response: 200 OK
{
  "id": "129b6ded-ff74-495d-a52a-0daee54e85b0",
  "region": "NSW",
  "price": 85.5,
  "demand": 8200,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z",
  "created_at": "2025-12-29T15:26:35Z"
}
```

### List Market Prices with Time Range

```bash
GET /api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z

Response: 200 OK
{
  "prices": [
    {
      "id": "2dea8fd3-fb32-42d3-8f4d-df5ea87a7dd0",
      "region": "SA",
      "price": 120,
      "demand": 1500,
      "interval_type": "30MIN_PREDISPATCH",
      "interval_start": "2025-12-30T11:00:00Z",
      "published_at": "2025-12-29T10:30:00Z",
      "created_at": "2025-12-29T15:27:04Z"
    },
    {
      "id": "129b6ded-ff74-495d-a52a-0daee54e85b0",
      "region": "NSW",
      "price": 85.5,
      "demand": 8200,
      "interval_type": "5MIN_PREDISPATCH",
      "interval_start": "2025-12-30T10:00:00Z",
      "published_at": "2025-12-29T09:55:00Z",
      "created_at": "2025-12-29T15:26:35Z"
    }
  ],
  "total": 2
}
```

### Query Parameters

- `from` (required): Start of time range (ISO 8601)
- `to` (required): End of time range (ISO 8601)
- `region` (optional): Filter by NEM region (NSW, VIC, QLD, SA, TAS)
- `interval_type` (optional): Filter by interval type (5MIN_PREDISPATCH, 30MIN_PREDISPATCH)
- `limit` (optional): Maximum number of results (pagination)
- `offset` (optional): Skip first N results (pagination)

### Filter Examples

```bash
# Filter by region
GET /api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&region=NSW

# Filter by interval type
GET /api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&interval_type=5MIN_PREDISPATCH

# Pagination
GET /api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&limit=10&offset=0
```

## Domain Model

### MarketPrice Aggregate

```go
type MarketPrice struct {
    ID            string        // UUID
    Region        string        // NSW, VIC, QLD, SA, TAS
    Price         float64       // $/MWh (non-negative)
    Demand        float64       // MW (positive)
    IntervalType  IntervalType  // 5MIN_PREDISPATCH or 30MIN_PREDISPATCH
    IntervalStart time.Time     // ISO 8601 timestamp
    PublishedAt   time.Time     // When AEMO published this forecast
    CreatedAt     time.Time     // When we received it
}
```

### Validation Rules

1. **Price ≥ 0** - Prices must be non-negative ($/MWh)
2. **Demand > 0** - System-wide demand must be positive (MW)
3. **Valid Region** - Must be one of: NSW, VIC, QLD, SA, TAS
4. **Valid Interval Type** - Must be: 5MIN_PREDISPATCH or 30MIN_PREDISPATCH
5. **Interval Start Not in Past** - Forecasts are for future intervals
6. **Published At Not in Future** - Sanity check on data source
7. **No Duplicate Intervals** - Unique constraint on (region, interval_type, interval_start)
8. **Interval Alignment**:
   - 5MIN: Timestamp must be on 5-minute boundaries (e.g., 10:00, 10:05, 10:10)
   - 30MIN: Timestamp must be on 30-minute boundaries (e.g., 10:00, 10:30, 11:00)

## Error Responses

All errors return JSON with `code`, `message`, and optional `details`:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "price must be non-negative ($/MWh)"
}
```

### Error Codes

- `400 Bad Request`:
  - `INVALID_JSON` - Malformed JSON request
  - `VALIDATION_ERROR` - Domain validation failed
  - `INVALID_INPUT` - Invalid input format
  - `MISSING_PARAMETERS` - Required query parameters missing
  - `INVALID_DATE_FORMAT` - Date not in ISO 8601 format
- `404 Not Found`:
  - `NOT_FOUND` - Market price not found
- `409 Conflict`:
  - `DUPLICATE_INTERVAL` - Price for this region/interval/time already exists
- `500 Internal Server Error`:
  - `INTERNAL_ERROR` - Server error

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (default: `postgres://market_user:market_pass@localhost:5433/market_data?sslmode=disable`)
- `PORT`: HTTP server port (default: `8080`)
- `LOG_LEVEL`: Logging level (default: `info`)

## Development

### Prerequisites

- Go 1.23.2
- PostgreSQL 18
- Docker & Docker Compose

### Running Locally

```bash
# Start PostgreSQL
docker-compose up -d market-db

# Run tests
go test -cover ./...

# Run service
go run cmd/server/main.go
```

### Running in Docker

```bash
# Build and start service
docker-compose up -d market-data

# View logs
docker-compose logs -f market-data

# Stop service
docker-compose down
```

### Database

The service uses a dedicated PostgreSQL database (`market_data`) on port 5433:

```bash
# Connect to database
docker exec -it market-db psql -U market_user -d market_data

# Check table structure
\d market_prices

# Query data
SELECT * FROM market_prices ORDER BY interval_start DESC LIMIT 10;
```

### Migrations

Migrations are automatically applied on service startup using `golang-migrate`. Migration files are in:

```
internal/adapters/postgres/migrations/
├── 000001_create_market_prices.up.sql
└── 000001_create_market_prices.down.sql
```

## Testing

The service follows **Test-Driven Development (TDD)** with comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# View detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

- **Domain Layer**: 96.7% coverage
- **Repository Layer**: 91.1% coverage
- **HTTP Layer**: 66.0% coverage
- **Overall**: 84.6% coverage

### Test Structure

- `internal/domain/price_test.go` - Domain validation tests (table-driven)
- `internal/adapters/postgres/repository_test.go` - Integration tests with real PostgreSQL
- `internal/adapters/http/handler_test.go` - Unit tests with manual mocks

## Database Schema

```sql
CREATE TABLE market_prices (
    id VARCHAR(36) PRIMARY KEY,
    region VARCHAR(3) NOT NULL CHECK (region IN ('NSW', 'VIC', 'QLD', 'SA', 'TAS')),
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    demand NUMERIC(10,2) NOT NULL CHECK (demand > 0),
    interval_type VARCHAR(20) NOT NULL CHECK (interval_type IN ('5MIN_PREDISPATCH', '30MIN_PREDISPATCH')),
    interval_start TIMESTAMP NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Unique constraint: no duplicate (region, interval_type, interval_start)
CREATE UNIQUE INDEX idx_market_prices_unique_interval
    ON market_prices(region, interval_type, interval_start);

-- Time-range query optimization
CREATE INDEX idx_market_prices_time_range
    ON market_prices(region, interval_type, interval_start DESC);
```

## Dependencies

- `github.com/gorilla/mux` - HTTP routing
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation
- `github.com/golang-migrate/migrate/v4` - Database migrations
- `github.com/stretchr/testify` - Testing assertions

## Project Status

- ✅ M3 Implementation Complete
- ✅ Domain layer with 8 validation rules
- ✅ PostgreSQL repository with time-series queries
- ✅ REST API with error handling
- ✅ Docker support
- ✅ Comprehensive test coverage (>80%)

## Next Steps

- M4: Event Bus Integration (NATS pub/sub)
  - Publish `MarketPriceUpdated` events
  - Enable event-driven communication with other services
