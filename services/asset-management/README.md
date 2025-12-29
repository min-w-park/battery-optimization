# Asset Management Service

The Asset Management Service is the first microservice in the battery optimization system. It manages battery specifications, constraints, and operational metadata for battery energy storage systems (BESS) participating in the Australian National Electricity Market (NEM).

## Architecture

This service follows **Hexagonal Architecture** (Ports & Adapters pattern):
- **Domain Layer** (`internal/domain/`): Pure business logic, no external dependencies
- **Ports** (`internal/ports/`): Interfaces defining boundaries
- **Adapters** (`internal/adapters/`): Infrastructure implementations
  - HTTP adapter: REST API handlers
  - PostgreSQL adapter: Database persistence

## Features

- ✅ Battery registration with comprehensive validation
- ✅ Battery retrieval by ID
- ✅ Battery listing with filtering (location, status) and pagination
- ✅ 10 business rule validations (capacity, power, efficiency, location, etc.)
- ✅ PostgreSQL persistence with JSONB for nested constraints
- ✅ RESTful API with proper error handling
- ✅ Database migrations using golang-migrate
- ✅ Docker support with multi-stage builds

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### POST /batteries
Register a new battery.

**Request Body**:
```json
{
  "capacity": 200.0,
  "max_power": 100.0,
  "ramp_rate": 50.0,
  "efficiency": 0.85,
  "location": "SA",
  "manufacturer": "Tesla",
  "constraints": {
    "warranty_eol": 0.8,
    "max_cycles": 10000,
    "operating_temp_min": -10.0,
    "operating_temp_max": 50.0,
    "grid_compliance_level": "AS4777"
  }
}
```

**Response**: `201 Created`
```json
{
  "id": "9959b197-2b4d-4d25-8f38-592e130f92e2",
  "capacity": 200.0,
  "max_power": 100.0,
  "ramp_rate": 50.0,
  "efficiency": 0.85,
  "location": "SA",
  "manufacturer": "Tesla",
  "constraints": {
    "warranty_eol": 0.8,
    "max_cycles": 10000,
    "operating_temp_min": -10.0,
    "operating_temp_max": 50.0,
    "grid_compliance_level": "AS4777"
  },
  "status": "REGISTERED",
  "created_at": "2025-12-29T08:14:41.948898Z",
  "updated_at": "2025-12-29T08:14:41.948898Z"
}
```

### GET /batteries/:id
Retrieve a battery by ID.

**Response**: `200 OK` or `404 Not Found`

### GET /batteries
List batteries with optional filtering.

**Query Parameters**:
- `location` (string): Filter by NEM region (NSW, VIC, QLD, SA, TAS)
- `status` (string): Filter by status (REGISTERED, TESTING, OPERATIONAL, etc.)
- `limit` (int): Maximum number of results
- `offset` (int): Pagination offset

**Example**:
```bash
GET /batteries?location=SA&limit=10
```

**Response**: `200 OK`
```json
{
  "batteries": [...],
  "total": 2
}
```

## Validation Rules

The service enforces 10 critical business rules:

1. **Capacity** must be > 0
2. **MaxPower** must be > 0
3. **MaxPower ≤ Capacity** (can't discharge more than capacity in 1 hour)
4. **RampRate** must be > 0
5. **RampRate ≤ MaxPower**
6. **Efficiency** must be between 0 and 1
7. **WarrantyEOL** must be between 0 and 1
8. **MaxCycles** must be > 0
9. **OperatingTempMin < OperatingTempMax**
10. **Location** must be valid NEM region (NSW, VIC, QLD, SA, TAS)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://asset_user:asset_pass@localhost:5432/asset_management?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging verbosity |

## Development

### Prerequisites
- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 18 (via Docker)

### Run Tests
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**Current Coverage**: 79.8% overall
- Domain: 96.7%
- Repository: 87.5%
- HTTP: 66.2%

### Run Locally (Without Docker)
```bash
# Ensure PostgreSQL is running
docker-compose up -d asset-db

# Set environment variables
export DATABASE_URL="postgres://asset_user:asset_pass@localhost:5432/asset_management?sslmode=disable"
export PORT=8080

# Run the service
go run cmd/server/main.go
```

### Run with Docker Compose
```bash
# Build and start
docker-compose up -d asset-management

# View logs
docker-compose logs -f asset-management

# Stop
docker-compose down
```

### Database Migrations

Migrations are managed by [golang-migrate](https://github.com/golang-migrate/migrate) and applied automatically on startup.

**Migration files**:
- `internal/adapters/postgres/migrations/000001_create_batteries.up.sql`
- `internal/adapters/postgres/migrations/000001_create_batteries.down.sql`

**Manual migration**:
```bash
# Connect to database
docker exec -it asset-db psql -U asset_user -d asset_management

# View batteries table
\d batteries

# Query batteries
SELECT * FROM batteries;
```

## Project Structure

```
services/asset-management/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Business logic (pure, no dependencies)
│   │   ├── battery.go          # Battery aggregate
│   │   ├── battery_test.go     # Domain tests (96.7% coverage)
│   │   └── errors.go           # Domain errors
│   ├── ports/                   # Interfaces
│   │   └── repository.go       # Repository interface
│   └── adapters/                # Infrastructure
│       ├── http/                # REST API
│       │   ├── dto.go          # Data transfer objects
│       │   ├── handler.go      # HTTP handlers
│       │   ├── handler_test.go # Handler tests
│       │   └── routes.go       # Route configuration
│       └── postgres/            # Database
│           ├── migrations/     # SQL migrations
│           ├── repository.go   # Repository implementation
│           └── repository_test.go
├── Dockerfile                   # Multi-stage Docker build
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## Testing Examples

### Create Battery (Success)
```bash
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{
    "capacity": 200.0,
    "max_power": 100.0,
    "ramp_rate": 50.0,
    "efficiency": 0.85,
    "location": "SA",
    "manufacturer": "Tesla",
    "constraints": {
      "warranty_eol": 0.8,
      "max_cycles": 10000,
      "operating_temp_min": -10.0,
      "operating_temp_max": 50.0,
      "grid_compliance_level": "AS4777"
    }
  }'
```

### Validation Error (Invalid Capacity)
```bash
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{
    "capacity": 0,
    "max_power": 100.0,
    ...
  }'

# Response: 400 Bad Request
# {"code":"VALIDATION_ERROR","message":"capacity must be greater than 0"}
```

### Get Battery by ID
```bash
curl http://localhost:8080/api/v1/batteries/{battery-id}
```

### List Batteries with Filters
```bash
# Filter by location
curl "http://localhost:8080/api/v1/batteries?location=SA"

# Pagination
curl "http://localhost:8080/api/v1/batteries?limit=10&offset=0"

# Multiple filters
curl "http://localhost:8080/api/v1/batteries?location=NSW&status=REGISTERED&limit=5"
```

## Technology Stack

- **Language**: Go 1.23
- **Database**: PostgreSQL 18
- **HTTP Router**: gorilla/mux
- **Testing**: testify
- **Migrations**: golang-migrate v4
- **Containerization**: Docker (multi-stage build)

## Design Patterns

- **Hexagonal Architecture**: Clear separation between domain and infrastructure
- **Repository Pattern**: Abstract data access with interfaces
- **Dependency Injection**: Constructor-based injection in main.go
- **Table-Driven Tests**: Comprehensive test coverage with clear scenarios
- **Error Wrapping**: Contextual error information with error chains

## Future Enhancements (Out of Scope for M2)

- Event publishing (M4 - NATS integration)
- Update/Delete operations
- Authentication & authorization
- Caching layer
- Metrics & observability (Prometheus)
- Health check endpoint
- API documentation (OpenAPI/Swagger)

## Related Documentation

- [M2 Domain Specification](../../docs/milestones/M2-DOMAIN-SPEC.md)
- [M2 API Specification](../../docs/milestones/M2-API-SPEC.md)
- [M2 Checklist](../../docs/milestones/M2-CHECKLIST.md)
- [EVENTS.md](../../EVENTS.md) - Event catalog for future milestones
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - TDD workflow and development philosophy

## License

Part of the Battery Optimization System project.
