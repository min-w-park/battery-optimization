# M2: Implementation Checklist

## 📋 How to Use This Checklist

**⚠️ CRITICAL: READ THESE FIRST**:
1. [TDD Guide](../guides/TDD-GUIDE.md) - Test-Driven Development (Kent Beck style)
2. [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development philosophy and workflow
3. [Battery Domain Skill](../../.claude/skills/BATTERY-DOMAIN-SKILL.md) - Domain concepts

**Then follow this checklist**:
1. Work through tasks in order (top to bottom)
2. **Write tests BEFORE implementation** (Red → Green → Refactor)
3. Check off `[ ]` boxes as you complete each task
4. Each phase should take 2-3 hours
5. If stuck, refer to detailed spec documents
6. Commit after each major milestone

---

## Phase 1: Project Setup (30-60 min)

### 1.1 Directory Structure
- [ ] Create `services/asset-management/` directory
- [ ] Create subdirectories:
  ```
  cmd/server/
  internal/domain/
  internal/ports/
  internal/adapters/http/
  internal/adapters/postgres/
  ```
- [ ] Initialize Go module: `go mod init github.com/[user]/battery-optimization/services/asset-management`

### 1.2 Dependencies
- [ ] Add dependencies to go.mod:
  ```bash
  go get github.com/google/uuid
  go get github.com/stretchr/testify
  go get github.com/lib/pq
  go get github.com/gorilla/mux  # or chi/echo
  ```

### 1.3 Database Setup
- [ ] Create `migrations/001_create_batteries.sql`
- [ ] Add batteries table schema (see below)
- [ ] Update docker-compose.yml to mount migrations

**SQL Migration**:
```sql
CREATE TABLE IF NOT EXISTS batteries (
    id UUID PRIMARY KEY,
    capacity FLOAT NOT NULL CHECK (capacity > 0),
    max_power FLOAT NOT NULL CHECK (max_power > 0),
    ramp_rate FLOAT NOT NULL CHECK (ramp_rate > 0),
    efficiency FLOAT NOT NULL CHECK (efficiency >= 0 AND efficiency <= 1),
    location VARCHAR(10) NOT NULL,
    manufacturer VARCHAR(100) NOT NULL,
    warranty_eol FLOAT NOT NULL,
    max_cycles INTEGER NOT NULL,
    temp_min FLOAT NOT NULL,
    temp_max FLOAT NOT NULL,
    grid_compliance TEXT[],
    status VARCHAR(20) NOT NULL DEFAULT 'REGISTERED',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batteries_location ON batteries(location);
CREATE INDEX idx_batteries_status ON batteries(status);
```

---

## Phase 2: Domain Layer (2-3 hours)

### 2.1 Domain Errors
- [ ] Create `internal/domain/errors.go`
- [ ] Define all error variables (see domain spec)
- [ ] Implement `ValidationError` type
- [ ] Write tests for ValidationError

### 2.2 Battery Aggregate - Tests First!
- [ ] Create `internal/domain/battery_test.go`
- [ ] Write test: `TestNewBattery_ValidInput`
- [ ] Write test: `TestNewBattery_InvalidCapacity`
- [ ] Write test: `TestNewBattery_InvalidMaxPower`
- [ ] Write test: `TestNewBattery_InvalidEfficiency`
- [ ] Write test: `TestNewBattery_InvalidLocation`
- [ ] Write test: `TestConstraints_Validate`

**TDD Cycle**: Write all tests first (they will fail)

### 2.3 Battery Aggregate - Implementation
- [ ] Create `internal/domain/battery.go`
- [ ] Define `Battery` struct
- [ ] Define `Constraints` struct
- [ ] Define `BatteryStatus` type and constants
- [ ] Implement `NewBattery()` constructor
- [ ] Implement `Battery.Validate()` method
- [ ] Implement `Constraints.Validate()` method
- [ ] Implement `isValidLocation()` helper
- [ ] Run tests → All should pass ✅

### 2.4 Domain Tests Coverage
- [ ] Run: `go test -cover ./internal/domain/...`
- [ ] Verify coverage > 80%
- [ ] Add edge case tests if needed

**Checkpoint**: Domain layer complete, all tests green

---

## Phase 3: Repository Layer (2-3 hours)

### 3.1 Repository Interface
- [ ] Create `internal/ports/repository.go`
- [ ] Define `BatteryRepository` interface:
  ```go
  type BatteryRepository interface {
      Create(ctx context.Context, battery *domain.Battery) error
      FindByID(ctx context.Context, id string) (*domain.Battery, error)
      List(ctx context.Context, filter ListFilter) ([]*domain.Battery, error)
  }
  ```
- [ ] Define `ListFilter` struct (location, status, pagination)

### 3.2 PostgreSQL Repository - Tests First!
- [ ] Create `internal/adapters/postgres/repository_test.go`
- [ ] Setup test database connection helper
- [ ] Write test: `TestCreate_Success`
- [ ] Write test: `TestCreate_DuplicateID`
- [ ] Write test: `TestFindByID_Success`
- [ ] Write test: `TestFindByID_NotFound`
- [ ] Write test: `TestList_WithFilters`

### 3.3 PostgreSQL Repository - Implementation
- [ ] Create `internal/adapters/postgres/repository.go`
- [ ] Implement `PostgresRepository` struct
- [ ] Implement `Create()` method
- [ ] Implement `FindByID()` method
- [ ] Implement `List()` method
- [ ] Handle SQL errors → domain errors mapping
- [ ] Run tests → All should pass ✅

**Tips**:
- Use `sqlx` or `database/sql`
- Handle `sql.ErrNoRows` → `domain.ErrBatteryNotFound`
- Use prepared statements for SQL injection protection

**Checkpoint**: Repository layer complete, can save/retrieve batteries

---

## Phase 4: HTTP API Layer (2-3 hours)

### 4.1 DTOs
- [ ] Create `internal/adapters/http/dto.go`
- [ ] Define `CreateBatteryRequest` struct with validation tags
- [ ] Define `BatteryResponse` struct
- [ ] Define `ListBatteriesResponse` struct
- [ ] Define `ErrorResponse` struct
- [ ] Implement conversion functions:
  - `requestToDomain(req CreateBatteryRequest) *domain.Battery`
  - `domainToResponse(battery *domain.Battery) BatteryResponse`

### 4.2 HTTP Handler - Tests First!
- [ ] Create `internal/adapters/http/handler_test.go`
- [ ] Setup mock repository
- [ ] Write test: `TestCreateBattery_Success`
- [ ] Write test: `TestCreateBattery_InvalidJSON`
- [ ] Write test: `TestCreateBattery_ValidationError`
- [ ] Write test: `TestGetBattery_Success`
- [ ] Write test: `TestGetBattery_NotFound`
- [ ] Write test: `TestListBatteries_Success`

### 4.3 HTTP Handler - Implementation
- [ ] Create `internal/adapters/http/handler.go`
- [ ] Implement `Handler` struct (holds repository)
- [ ] Implement `CreateBattery(w http.ResponseWriter, r *http.Request)`
- [ ] Implement `GetBattery(w http.ResponseWriter, r *http.Request)`
- [ ] Implement `ListBatteries(w http.ResponseWriter, r *http.Request)`
- [ ] Implement error mapping helper
- [ ] Run tests → All should pass ✅

### 4.4 Routes
- [ ] Create `internal/adapters/http/routes.go`
- [ ] Setup router (mux/chi/echo)
- [ ] Register routes:
  - `POST /api/v1/batteries`
  - `GET /api/v1/batteries/:id`
  - `GET /api/v1/batteries`
- [ ] Add middleware (logging, CORS if needed)

**Checkpoint**: HTTP layer complete, API defined

---

## Phase 5: Main Application (1-2 hours)

### 5.1 Main Entry Point
- [ ] Create `cmd/server/main.go`
- [ ] Implement dependency injection:
  ```go
  func main() {
      // 1. Load config
      // 2. Connect to database
      // 3. Create repository
      // 4. Create handler
      // 5. Setup routes
      // 6. Start server
  }
  ```
- [ ] Add graceful shutdown
- [ ] Add basic logging

### 5.2 Configuration
- [ ] Support environment variables:
  - `DATABASE_URL`
  - `PORT`
  - `LOG_LEVEL`
- [ ] Add `.env.example` file

### 5.3 Docker
- [ ] Create `Dockerfile`:
  ```dockerfile
  FROM golang:1.21 as builder
  WORKDIR /app
  COPY go.* ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 go build -o server ./cmd/server
  
  FROM alpine:latest
  COPY --from=builder /app/server /server
  EXPOSE 8080
  CMD ["/server"]
  ```
- [ ] Update `docker-compose.yml` to include asset-management service

**Checkpoint**: Service runs in Docker

---

## Phase 6: Integration Testing (1-2 hours)

### 6.1 End-to-End Test
- [ ] Start services: `docker-compose up`
- [ ] Verify database connection
- [ ] Test POST /batteries with curl
- [ ] Verify battery in database
- [ ] Test GET /batteries/:id
- [ ] Test GET /batteries (list)

### 6.2 Error Cases
- [ ] Test invalid capacity (should return 400)
- [ ] Test invalid location (should return 400)
- [ ] Test non-existent ID (should return 404)

### 6.3 Documentation
- [ ] Add API examples to README
- [ ] Document environment variables
- [ ] Add "How to run" instructions

**Example Commands**:
```bash
# Start services
docker-compose up -d asset-db asset-management

# Register a battery
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{"capacity": 200.0, "maxPower": 100.0, ...}'

# Get battery
curl http://localhost:8080/api/v1/batteries/{id}
```

---

## Phase 7: Polish & Documentation (1 hour)

### 7.1 Code Quality
- [ ] Run `go fmt ./...`
- [ ] Run `go vet ./...`
- [ ] Run `golangci-lint run` (if installed)
- [ ] Add comments to exported functions
- [ ] Review code for improvements

### 7.2 Tests
- [ ] All tests pass: `go test ./...`
- [ ] Coverage report: `go test -cover ./...`
- [ ] Verify >80% coverage

### 7.3 Documentation
- [ ] Update `PLANNING.md` with M2 completion
- [ ] Add M2 details to main README
- [ ] Create API documentation (curl examples)
- [ ] Add architecture diagrams (if needed)

### 7.4 Git
- [ ] Commit all changes
- [ ] Push to GitHub
- [ ] Tag: `git tag v0.1.0-m2`

---

## 🎯 Definition of Done

All items below must be true:

- [x] All domain tests pass
- [x] All repository tests pass
- [x] All handler tests pass
- [x] Service runs in Docker
- [x] Can register battery via curl
- [x] Can retrieve battery by ID
- [x] Can list batteries
- [x] Error handling works correctly
- [x] Code is formatted and linted
- [x] Documentation is updated
- [x] Changes are committed to Git

---

## 🐛 Common Issues & Solutions

### Database Connection Fails
```
Error: could not connect to database
Solution: Check docker-compose is running, verify DATABASE_URL
```

### Tests Fail on CI
```
Error: postgres not available in tests
Solution: Use test database or mock repository for unit tests
```

### Import Cycle Error
```
Error: import cycle not allowed
Solution: Ensure domain doesn't import adapters, only ports
```

---

## 📊 Progress Tracking

**Estimated Time**: 10-12 hours
**Actual Time**: ___ hours

**Phases Completed**:
- [ ] Phase 1: Setup (0.5-1h)
- [ ] Phase 2: Domain (2-3h)
- [ ] Phase 3: Repository (2-3h)
- [ ] Phase 4: HTTP API (2-3h)
- [ ] Phase 5: Main App (1-2h)
- [ ] Phase 6: Integration (1-2h)
- [ ] Phase 7: Polish (1h)

**Blockers / Notes**:
```
(Add any issues you encountered and how you solved them)
```

---

## ✅ Ready for M3

Once M2 is complete:
- You have a working Asset Management Service
- You understand DDD, TDD, and hexagonal architecture
- You're ready to build Market Data Service (similar pattern)
- You can explain your architecture to Seb

**Great job! 🎉**

---

**Next Steps**:
1. Take a break
2. Review what you learned
3. Start M3 when ready
