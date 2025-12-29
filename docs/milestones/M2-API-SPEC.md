# M2: API Specification - REST Endpoints

## 🌐 Base URL

```
http://localhost:8080/api/v1
```

## 📋 Endpoints

### 1. Register Battery

Register a new battery in the system.

**Endpoint**: `POST /batteries`

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "capacity": 200.0,
  "maxPower": 100.0,
  "rampRate": 10.0,
  "efficiency": 0.92,
  "location": "NSW",
  "manufacturer": "Tesla",
  "constraints": {
    "warrantyEOL": 0.7,
    "maxCycles": 10000,
    "tempMin": -10.0,
    "tempMax": 45.0,
    "gridCompliance": ["FCAS", "FFR"]
  }
}
```

**Success Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "capacity": 200.0,
  "maxPower": 100.0,
  "rampRate": 10.0,
  "efficiency": 0.92,
  "location": "NSW",
  "manufacturer": "Tesla",
  "constraints": {
    "warrantyEOL": 0.7,
    "maxCycles": 10000,
    "tempMin": -10.0,
    "tempMax": 45.0,
    "gridCompliance": ["FCAS", "FFR"]
  },
  "status": "REGISTERED",
  "createdAt": "2025-01-15T10:30:00Z",
  "updatedAt": "2025-01-15T10:30:00Z"
}
```

**Error Response** (400 Bad Request):
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      "capacity must be greater than 0",
      "max power cannot exceed capacity"
    ]
  }
}
```

**Error Response** (500 Internal Server Error):
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to register battery"
  }
}
```

---

### 2. Get Battery by ID

Retrieve a specific battery by its ID.

**Endpoint**: `GET /batteries/:id`

**Path Parameters**:
- `id` (string, required): Battery UUID

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "capacity": 200.0,
  "maxPower": 100.0,
  "rampRate": 10.0,
  "efficiency": 0.92,
  "location": "NSW",
  "manufacturer": "Tesla",
  "constraints": {
    "warrantyEOL": 0.7,
    "maxCycles": 10000,
    "tempMin": -10.0,
    "tempMax": 45.0,
    "gridCompliance": ["FCAS", "FFR"]
  },
  "status": "REGISTERED",
  "createdAt": "2025-01-15T10:30:00Z",
  "updatedAt": "2025-01-15T10:30:00Z"
}
```

**Error Response** (404 Not Found):
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Battery not found"
  }
}
```

---

### 3. List Batteries

List all batteries (optional: with pagination).

**Endpoint**: `GET /batteries`

**Query Parameters** (optional):
- `limit` (integer, default: 50): Maximum number of results
- `offset` (integer, default: 0): Number of results to skip
- `location` (string): Filter by NEM region (NSW, VIC, QLD, SA, TAS)
- `status` (string): Filter by status (REGISTERED, OPERATIONAL, etc.)

**Example Request**:
```
GET /batteries?limit=10&location=NSW&status=OPERATIONAL
```

**Success Response** (200 OK):
```json
{
  "batteries": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "capacity": 200.0,
      "maxPower": 100.0,
      "location": "NSW",
      "manufacturer": "Tesla",
      "status": "OPERATIONAL",
      "createdAt": "2025-01-15T10:30:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "capacity": 150.0,
      "maxPower": 75.0,
      "location": "NSW",
      "manufacturer": "BYD",
      "status": "OPERATIONAL",
      "createdAt": "2025-01-14T08:20:00Z"
    }
  ],
  "total": 2,
  "limit": 10,
  "offset": 0
}
```

---

## 📦 Data Transfer Objects (DTOs)

### Request DTOs

```go
package dto

// CreateBatteryRequest represents the request to register a battery
type CreateBatteryRequest struct {
    Capacity     float64            `json:"capacity" binding:"required,gt=0"`
    MaxPower     float64            `json:"maxPower" binding:"required,gt=0"`
    RampRate     float64            `json:"rampRate" binding:"required,gt=0"`
    Efficiency   float64            `json:"efficiency" binding:"required,gte=0,lte=1"`
    Location     string             `json:"location" binding:"required,oneof=NSW VIC QLD SA TAS"`
    Manufacturer string             `json:"manufacturer" binding:"required"`
    Constraints  ConstraintsRequest `json:"constraints" binding:"required"`
}

type ConstraintsRequest struct {
    WarrantyEOL    float64  `json:"warrantyEOL" binding:"required,gte=0,lte=1"`
    MaxCycles      int      `json:"maxCycles" binding:"required,gt=0"`
    TempMin        float64  `json:"tempMin" binding:"required"`
    TempMax        float64  `json:"tempMax" binding:"required"`
    GridCompliance []string `json:"gridCompliance"`
}
```

### Response DTOs

```go
package dto

import "time"

// BatteryResponse represents a battery in API responses
type BatteryResponse struct {
    ID           string               `json:"id"`
    Capacity     float64              `json:"capacity"`
    MaxPower     float64              `json:"maxPower"`
    RampRate     float64              `json:"rampRate"`
    Efficiency   float64              `json:"efficiency"`
    Location     string               `json:"location"`
    Manufacturer string               `json:"manufacturer"`
    Constraints  ConstraintsResponse  `json:"constraints"`
    Status       string               `json:"status"`
    CreatedAt    time.Time            `json:"createdAt"`
    UpdatedAt    time.Time            `json:"updatedAt"`
}

type ConstraintsResponse struct {
    WarrantyEOL    float64  `json:"warrantyEOL"`
    MaxCycles      int      `json:"maxCycles"`
    TempMin        float64  `json:"tempMin"`
    TempMax        float64  `json:"tempMax"`
    GridCompliance []string `json:"gridCompliance"`
}

// ListBatteriesResponse represents paginated battery list
type ListBatteriesResponse struct {
    Batteries []BatteryResponse `json:"batteries"`
    Total     int               `json:"total"`
    Limit     int               `json:"limit"`
    Offset    int               `json:"offset"`
}

// ErrorResponse represents an error
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string   `json:"code"`
    Message string   `json:"message"`
    Details []string `json:"details,omitempty"`
}
```

---

## 🧪 Example curl Commands

### Register a Battery

```bash
curl -X POST http://localhost:8080/api/v1/batteries \
  -H "Content-Type: application/json" \
  -d '{
    "capacity": 200.0,
    "maxPower": 100.0,
    "rampRate": 10.0,
    "efficiency": 0.92,
    "location": "NSW",
    "manufacturer": "Tesla",
    "constraints": {
      "warrantyEOL": 0.7,
      "maxCycles": 10000,
      "tempMin": -10.0,
      "tempMax": 45.0,
      "gridCompliance": ["FCAS", "FFR"]
    }
  }'
```

### Get Battery by ID

```bash
curl http://localhost:8080/api/v1/batteries/550e8400-e29b-41d4-a716-446655440000
```

### List All Batteries

```bash
curl http://localhost:8080/api/v1/batteries
```

### List Batteries with Filters

```bash
curl "http://localhost:8080/api/v1/batteries?location=NSW&status=OPERATIONAL&limit=10"
```

---

## 🔄 HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Successful GET request |
| 201 | Created | Successful POST (battery registered) |
| 400 | Bad Request | Validation error, malformed JSON |
| 404 | Not Found | Battery ID doesn't exist |
| 500 | Internal Server Error | Database error, unexpected failure |

---

## 🛡️ Error Handling Strategy

### Validation Errors (400)

Map domain validation errors to HTTP 400:

```go
func mapDomainError(err error) (int, ErrorResponse) {
    switch {
    case errors.Is(err, domain.ErrInvalidCapacity),
         errors.Is(err, domain.ErrInvalidMaxPower),
         errors.Is(err, domain.ErrInvalidEfficiency):
        return http.StatusBadRequest, ErrorResponse{
            Error: ErrorDetail{
                Code:    "VALIDATION_ERROR",
                Message: err.Error(),
            },
        }
    case errors.Is(err, domain.ErrBatteryNotFound):
        return http.StatusNotFound, ErrorResponse{
            Error: ErrorDetail{
                Code:    "NOT_FOUND",
                Message: "Battery not found",
            },
        }
    default:
        return http.StatusInternalServerError, ErrorResponse{
            Error: ErrorDetail{
                Code:    "INTERNAL_ERROR",
                Message: "Internal server error",
            },
        }
    }
}
```

### Multiple Validation Errors

If domain returns `ValidationError` with multiple issues:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      "capacity must be greater than 0",
      "efficiency must be between 0 and 1",
      "invalid location: must be NSW, VIC, QLD, SA, or TAS"
    ]
  }
}
```

---

## 🧪 Handler Tests

### Test Structure

```go
package http

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateBattery_Success(t *testing.T) {
    // Given
    mockRepo := new(MockRepository)
    handler := NewHandler(mockRepo)
    
    reqBody := CreateBatteryRequest{
        Capacity:   200.0,
        MaxPower:   100.0,
        RampRate:   10.0,
        Efficiency: 0.92,
        Location:   "NSW",
        Manufacturer: "Tesla",
        Constraints: ConstraintsRequest{
            WarrantyEOL: 0.7,
            MaxCycles:   10000,
            TempMin:     -10.0,
            TempMax:     45.0,
        },
    }
    
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/batteries", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    mockRepo.On("Create", mock.Anything).Return(nil)
    
    // When
    handler.CreateBattery(w, req)
    
    // Then
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var resp BatteryResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NotEmpty(t, resp.ID)
    assert.Equal(t, reqBody.Capacity, resp.Capacity)
}

func TestCreateBattery_ValidationError(t *testing.T) {
    // Given
    handler := NewHandler(nil)
    
    reqBody := CreateBatteryRequest{
        Capacity: -100.0, // Invalid
        MaxPower: 50.0,
        // ... other fields
    }
    
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/batteries", bytes.NewReader(body))
    w := httptest.NewRecorder()
    
    // When
    handler.CreateBattery(w, req)
    
    // Then
    assert.Equal(t, http.StatusBadRequest, w.Code)
    
    var resp ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
}
```

---

## 🎯 API Design Principles

1. **RESTful**: Use proper HTTP verbs and status codes
2. **Consistent**: Same error format across all endpoints
3. **Clear**: Descriptive error messages for developers
4. **Versioned**: `/api/v1/` prefix for future compatibility
5. **JSON**: Standard format, camelCase field names

---

## 📚 OpenAPI/Swagger (Future)

Once API is stable, document with OpenAPI 3.0:

```yaml
openapi: 3.0.0
info:
  title: Battery Optimization API
  version: 1.0.0
paths:
  /batteries:
    post:
      summary: Register a battery
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateBatteryRequest'
      responses:
        '201':
          description: Battery registered successfully
```

---

**Next**: Implement these endpoints following TDD!
