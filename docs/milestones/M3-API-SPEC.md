# M3: API Specification - REST Endpoints

## 🌐 Base URL

```
http://localhost:8081/api/v1
```

**Note**: Market Data Service runs on port **8081** (Asset Management uses 8080)

## 📋 Endpoints

### 1. Create Market Price

Record a new AEMO price forecast in the system.

**Endpoint**: `POST /prices`

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "region": "NSW",
  "price": 85.50,
  "demand": 8200.0,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z"
}
```

**Field Descriptions**:
- `region` (string, required): NEM region - must be one of: NSW, VIC, QLD, SA, TAS
- `price` (float64, required): Spot price in $/MWh, must be >= 0 (typically $20-$300/MWh)
- `demand` (float64, required): System-wide demand forecast in MW, must be > 0 (typically 1,000-35,000 MW)
- `interval_type` (string, required): Either "5MIN_PREDISPATCH" or "30MIN_PREDISPATCH"
- `interval_start` (string, required): ISO 8601 timestamp for forecast interval (must align with interval_type boundaries)
- `published_at` (string, required): ISO 8601 timestamp when AEMO published this forecast (cannot be in future)

**Success Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "region": "NSW",
  "price": 85.50,
  "demand": 8200.0,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z",
  "created_at": "2025-12-29T10:00:15Z"
}
```

**Error Response** (400 Bad Request - Validation Error):
```json
{
  "code": "VALIDATION_ERROR",
  "message": "price must be non-negative ($/MWh)",
  "details": null
}
```

**Error Response** (400 Bad Request - Invalid Timestamp Format):
```json
{
  "code": "INVALID_INPUT",
  "message": "invalid interval_start format: parsing time \"2025-12-30\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\"",
  "details": null
}
```

**Error Response** (409 Conflict - Duplicate Interval):
```json
{
  "code": "DUPLICATE_INTERVAL",
  "message": "price for this region/interval/time already exists",
  "details": null
}
```

**Error Response** (500 Internal Server Error):
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to save market price",
  "details": null
}
```

---

### 2. Get Market Price by ID

Retrieve a specific market price forecast by its ID.

**Endpoint**: `GET /prices/:id`

**Path Parameters**:
- `id` (string, required): Market price UUID

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "region": "NSW",
  "price": 85.50,
  "demand": 8200.0,
  "interval_type": "5MIN_PREDISPATCH",
  "interval_start": "2025-12-30T10:00:00Z",
  "published_at": "2025-12-29T09:55:00Z",
  "created_at": "2025-12-29T10:00:15Z"
}
```

**Error Response** (404 Not Found):
```json
{
  "code": "NOT_FOUND",
  "message": "Market price not found",
  "details": null
}
```

---

### 3. List Market Prices (Time-Range Query)

Query market price forecasts within a time range with optional filters.

**Endpoint**: `GET /prices`

**Query Parameters**:
- `from` (string, **required**): Start of time range (ISO 8601 format)
- `to` (string, **required**): End of time range (ISO 8601 format)
- `region` (string, optional): Filter by NEM region (NSW, VIC, QLD, SA, TAS)
- `interval_type` (string, optional): Filter by interval type (5MIN_PREDISPATCH or 30MIN_PREDISPATCH)
- `limit` (integer, optional, default: 1000): Maximum number of results
- `offset` (integer, optional, default: 0): Number of results to skip (pagination)

**Example Request**:
```
GET /prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&region=NSW&interval_type=5MIN_PREDISPATCH&limit=50
```

**Success Response** (200 OK):
```json
{
  "prices": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "region": "NSW",
      "price": 85.50,
      "demand": 8200.0,
      "interval_type": "5MIN_PREDISPATCH",
      "interval_start": "2025-12-30T10:05:00Z",
      "published_at": "2025-12-30T10:00:00Z",
      "created_at": "2025-12-30T10:00:15Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "region": "NSW",
      "price": 82.00,
      "demand": 8150.0,
      "interval_type": "5MIN_PREDISPATCH",
      "interval_start": "2025-12-30T10:00:00Z",
      "published_at": "2025-12-30T09:55:00Z",
      "created_at": "2025-12-30T09:55:12Z"
    }
  ],
  "total": 2
}
```

**Note**: Results are ordered by `interval_start` DESC (newest first).

**Error Response** (400 Bad Request - Missing Required Parameters):
```json
{
  "code": "MISSING_PARAMETERS",
  "message": "Required parameters: from (ISO 8601), to (ISO 8601)",
  "details": null
}
```

**Error Response** (400 Bad Request - Invalid Date Format):
```json
{
  "code": "INVALID_DATE_FORMAT",
  "message": "Parameter 'from' must be ISO 8601 format",
  "details": null
}
```

**Error Response** (500 Internal Server Error):
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Failed to list market prices",
  "details": null
}
```

---

## 📦 Data Transfer Objects (DTOs)

### Request DTOs

```go
package http

import (
    "fmt"
    "time"

    "github.com/minwook/battery-optimization/services/market-data/internal/domain"
)

// CreateMarketPriceRequest represents the request to create a market price
type CreateMarketPriceRequest struct {
    Region        string  `json:"region"`
    Price         float64 `json:"price"`
    Demand        float64 `json:"demand"`
    IntervalType  string  `json:"interval_type"`
    IntervalStart string  `json:"interval_start"` // ISO 8601
    PublishedAt   string  `json:"published_at"`   // ISO 8601
}

// ToDomain converts CreateMarketPriceRequest to domain.MarketPrice
func (r *CreateMarketPriceRequest) ToDomain() (*domain.MarketPrice, error) {
    // Parse ISO 8601 timestamps
    intervalStart, err := time.Parse(time.RFC3339, r.IntervalStart)
    if err != nil {
        return nil, fmt.Errorf("invalid interval_start format: %w", err)
    }

    publishedAt, err := time.Parse(time.RFC3339, r.PublishedAt)
    if err != nil {
        return nil, fmt.Errorf("invalid published_at format: %w", err)
    }

    // Create domain entity (validation happens in NewMarketPrice)
    return domain.NewMarketPrice(
        r.Region,
        r.Price,
        r.Demand,
        domain.IntervalType(r.IntervalType),
        intervalStart,
        publishedAt,
    )
}
```

### Response DTOs

```go
package http

import (
    "time"

    "github.com/minwook/battery-optimization/services/market-data/internal/domain"
)

// MarketPriceResponse represents a market price in API responses
type MarketPriceResponse struct {
    ID            string  `json:"id"`
    Region        string  `json:"region"`
    Price         float64 `json:"price"`
    Demand        float64 `json:"demand"`
    IntervalType  string  `json:"interval_type"`
    IntervalStart string  `json:"interval_start"` // ISO 8601
    PublishedAt   string  `json:"published_at"`   // ISO 8601
    CreatedAt     string  `json:"created_at"`     // ISO 8601
}

// FromDomain converts domain.MarketPrice to MarketPriceResponse
func FromDomain(price *domain.MarketPrice) *MarketPriceResponse {
    return &MarketPriceResponse{
        ID:            price.ID,
        Region:        price.Region,
        Price:         price.Price,
        Demand:        price.Demand,
        IntervalType:  string(price.IntervalType),
        IntervalStart: price.IntervalStart.Format(time.RFC3339),
        PublishedAt:   price.PublishedAt.Format(time.RFC3339),
        CreatedAt:     price.CreatedAt.Format(time.RFC3339),
    }
}

// ListMarketPricesResponse represents paginated market price list
type ListMarketPricesResponse struct {
    Prices []*MarketPriceResponse `json:"prices"`
    Total  int                     `json:"total"`
}

// ErrorResponse represents an error
type ErrorResponse struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Details map[string]string `json:"details,omitempty"`
}
```

---

## 🧪 Example curl Commands

### Create Market Price (5-minute forecast)

```bash
curl -X POST http://localhost:8081/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "region": "NSW",
    "price": 85.50,
    "demand": 8200.0,
    "interval_type": "5MIN_PREDISPATCH",
    "interval_start": "2025-12-30T10:00:00Z",
    "published_at": "2025-12-29T09:55:00Z"
  }'
```

**Expected**: 201 Created with price ID

### Create Market Price (30-minute forecast)

```bash
curl -X POST http://localhost:8081/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "region": "SA",
    "price": 120.00,
    "demand": 1500.0,
    "interval_type": "30MIN_PREDISPATCH",
    "interval_start": "2025-12-30T11:00:00Z",
    "published_at": "2025-12-29T10:30:00Z"
  }'
```

**Expected**: 201 Created

### Get Market Price by ID

```bash
curl http://localhost:8081/api/v1/prices/550e8400-e29b-41d4-a716-446655440000
```

**Expected**: 200 OK with price data

### List Market Prices (Time Range - All Regions)

```bash
curl "http://localhost:8081/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z"
```

**Expected**: 200 OK with list of prices across all regions

### List Market Prices (Filtered by Region)

```bash
curl "http://localhost:8081/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&region=NSW"
```

**Expected**: 200 OK with only NSW prices

### List Market Prices (Filtered by Interval Type)

```bash
curl "http://localhost:8081/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&interval_type=5MIN_PREDISPATCH"
```

**Expected**: 200 OK with only 5-minute forecasts

### List with Multiple Filters + Pagination

```bash
curl "http://localhost:8081/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z&region=NSW&interval_type=5MIN_PREDISPATCH&limit=10&offset=0"
```

**Expected**: 200 OK with first 10 results

---

## 🔄 HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET request |
| 201 | Created | Successful POST (market price created) |
| 400 | Bad Request | Validation error, malformed JSON, invalid timestamp, missing parameters |
| 404 | Not Found | Market price ID doesn't exist |
| 409 | Conflict | Duplicate interval (same region/interval_type/interval_start) |
| 500 | Internal Server Error | Database error, unexpected failure |

---

## 🛡️ Error Handling Strategy

### Domain Validation Errors → 400 Bad Request

Map domain validation errors to HTTP 400:

```go
func (h *MarketPriceHandler) CreateMarketPrice(w http.ResponseWriter, r *http.Request) {
    var req CreateMarketPriceRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "INVALID_JSON",
            "Invalid JSON format", nil)
        return
    }

    price, err := req.ToDomain()
    if err != nil {
        // Timestamp parsing errors or validation errors
        if errors.Is(err, domain.ErrInvalidPrice) ||
           errors.Is(err, domain.ErrInvalidDemand) ||
           errors.Is(err, domain.ErrInvalidRegion) ||
           errors.Is(err, domain.ErrInvalidIntervalType) ||
           errors.Is(err, domain.ErrIntervalInPast) ||
           errors.Is(err, domain.ErrPublishedInFuture) ||
           errors.Is(err, domain.ErrIntervalAlignment) {
            respondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR",
                err.Error(), nil)
            return
        }
        respondWithError(w, http.StatusBadRequest, "INVALID_INPUT",
            err.Error(), nil)
        return
    }

    if err := h.repo.Create(r.Context(), price); err != nil {
        if errors.Is(err, domain.ErrDuplicateInterval) {
            respondWithError(w, http.StatusConflict, "DUPLICATE_INTERVAL",
                err.Error(), nil)
            return
        }
        log.Printf("Failed to create market price: %v", err)
        respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
            "Failed to save market price", nil)
        return
    }

    respondWithJSON(w, http.StatusCreated, FromDomain(price))
}
```

### Error Response Format

All errors follow this consistent format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": null
}
```

**Error Codes**:
- `INVALID_JSON` - Malformed JSON in request body
- `VALIDATION_ERROR` - Domain validation failed (price < 0, invalid region, etc.)
- `INVALID_INPUT` - Timestamp parsing error or other input issues
- `MISSING_PARAMETERS` - Required query parameters missing (from/to)
- `INVALID_DATE_FORMAT` - Timestamp not in ISO 8601 format
- `DUPLICATE_INTERVAL` - Price already exists for this region/interval/time
- `NOT_FOUND` - Market price ID not found
- `INTERNAL_ERROR` - Database or unexpected error

---

## 🧪 Handler Tests

### Test Structure

```go
package http

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gorilla/mux"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/minwook/battery-optimization/services/market-data/internal/domain"
    "github.com/minwook/battery-optimization/services/market-data/internal/ports"
)

// MockRepository is a manual mock for testing
type MockRepository struct {
    CreateFunc          func(ctx context.Context, price *domain.MarketPrice) error
    FindByIDFunc        func(ctx context.Context, id string) (*domain.MarketPrice, error)
    ListByTimeRangeFunc func(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error)
}

func (m *MockRepository) Create(ctx context.Context, price *domain.MarketPrice) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, price)
    }
    return nil
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*domain.MarketPrice, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(ctx, id)
    }
    return nil, nil
}

func (m *MockRepository) ListByTimeRange(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
    if m.ListByTimeRangeFunc != nil {
        return m.ListByTimeRangeFunc(ctx, filter)
    }
    return nil, nil
}

func TestCreateMarketPrice_Success(t *testing.T) {
    // Given
    mockRepo := &MockRepository{
        CreateFunc: func(ctx context.Context, price *domain.MarketPrice) error {
            return nil
        },
    }
    handler := NewMarketPriceHandler(mockRepo)

    reqBody := CreateMarketPriceRequest{
        Region:        "NSW",
        Price:         85.50,
        Demand:        8200.0,
        IntervalType:  "5MIN_PREDISPATCH",
        IntervalStart: "2025-12-30T10:00:00Z",
        PublishedAt:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
    }

    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/prices", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    // When
    handler.CreateMarketPrice(rec, req)

    // Then
    assert.Equal(t, http.StatusCreated, rec.Code)

    var resp MarketPriceResponse
    err := json.Unmarshal(rec.Body.Bytes(), &resp)
    require.NoError(t, err)
    assert.NotEmpty(t, resp.ID)
    assert.Equal(t, 85.50, resp.Price)
    assert.Equal(t, "NSW", resp.Region)
}

func TestCreateMarketPrice_ValidationError(t *testing.T) {
    // Given
    mockRepo := &MockRepository{}
    handler := NewMarketPriceHandler(mockRepo)

    reqBody := CreateMarketPriceRequest{
        Region:        "NSW",
        Price:         -10.0, // Invalid: must be >= 0
        Demand:        8200.0,
        IntervalType:  "5MIN_PREDISPATCH",
        IntervalStart: "2025-12-30T10:00:00Z",
        PublishedAt:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
    }

    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/prices", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    // When
    handler.CreateMarketPrice(rec, req)

    // Then
    assert.Equal(t, http.StatusBadRequest, rec.Code)

    var resp ErrorResponse
    err := json.Unmarshal(rec.Body.Bytes(), &resp)
    require.NoError(t, err)
    assert.Equal(t, "VALIDATION_ERROR", resp.Code)
    assert.Contains(t, resp.Message, "price")
}

func TestListMarketPrices_WithTimeRange(t *testing.T) {
    // Given
    price1, _ := domain.NewMarketPrice("NSW", 85.50, 8200.0,
        domain.Interval5Min,
        time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
        time.Now().Add(-5*time.Minute))

    mockRepo := &MockRepository{
        ListByTimeRangeFunc: func(ctx context.Context, filter ports.TimeRangeFilter) ([]*domain.MarketPrice, error) {
            assert.Equal(t, "2025-12-30T00:00:00Z", filter.From.Format(time.RFC3339))
            assert.Equal(t, "2025-12-31T00:00:00Z", filter.To.Format(time.RFC3339))
            return []*domain.MarketPrice{price1}, nil
        },
    }
    handler := NewMarketPriceHandler(mockRepo)

    req := httptest.NewRequest(http.MethodGet,
        "/api/v1/prices?from=2025-12-30T00:00:00Z&to=2025-12-31T00:00:00Z", nil)
    rec := httptest.NewRecorder()

    // When
    handler.ListMarketPrices(rec, req)

    // Then
    assert.Equal(t, http.StatusOK, rec.Code)

    var resp ListMarketPricesResponse
    err := json.Unmarshal(rec.Body.Bytes(), &resp)
    require.NoError(t, err)
    assert.Len(t, resp.Prices, 1)
    assert.Equal(t, 1, resp.Total)
}
```

---

## 🎯 API Design Principles

1. **RESTful**: Use proper HTTP verbs and status codes
2. **Time-Range Queries**: Required `from` and `to` parameters for efficient time-series queries
3. **ISO 8601 Timestamps**: Consistent date/time format across all endpoints
4. **Consistent Error Format**: Same error structure across all endpoints
5. **Clear Error Messages**: Descriptive messages for developers
6. **Versioned**: `/api/v1/` prefix for future compatibility
7. **JSON**: Standard format with snake_case field names (AEMO convention)
8. **Immutable Records**: POST only (no PUT/PATCH/DELETE for time-series data)

---

## 🔍 Key Differences from M2 (Battery API)

| Aspect | M2 Battery API | M3 Market Price API |
|--------|---------------|---------------------|
| **Port** | 8080 | 8081 |
| **Base URL** | `/api/v1/batteries` | `/api/v1/prices` |
| **Primary Query** | Location/status filters | Time-range (from/to) + optional filters |
| **Create Response** | 201 with status field | 201 without status (immutable) |
| **Update Support** | Yes (PUT/PATCH planned) | No (immutable time-series) |
| **Delete Support** | Yes (DELETE planned) | No (immutable time-series) |
| **Unique Constraint** | ID only | (region, interval_type, interval_start) |
| **Error Code for Duplicate** | N/A | 409 Conflict |
| **Timestamp Format** | createdAt/updatedAt | interval_start/published_at/created_at |
| **Field Naming** | camelCase | snake_case (AEMO convention) |

---

## 📚 OpenAPI/Swagger (Future)

Once API is stable, document with OpenAPI 3.0:

```yaml
openapi: 3.0.0
info:
  title: Market Data Service API
  version: 1.0.0
  description: AEMO price forecast management for Australian NEM
servers:
  - url: http://localhost:8081/api/v1
paths:
  /prices:
    post:
      summary: Create market price forecast
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateMarketPriceRequest'
      responses:
        '201':
          description: Market price created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MarketPriceResponse'
        '400':
          description: Validation error
        '409':
          description: Duplicate interval
    get:
      summary: List market prices by time range
      parameters:
        - name: from
          in: query
          required: true
          schema:
            type: string
            format: date-time
        - name: to
          in: query
          required: true
          schema:
            type: string
            format: date-time
        - name: region
          in: query
          schema:
            type: string
            enum: [NSW, VIC, QLD, SA, TAS]
        - name: interval_type
          in: query
          schema:
            type: string
            enum: [5MIN_PREDISPATCH, 30MIN_PREDISPATCH]
      responses:
        '200':
          description: List of market prices
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListMarketPricesResponse'
```

---

**Next**: Implement these endpoints following TDD!
