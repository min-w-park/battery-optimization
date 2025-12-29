# API Design Skill

## When to use this skill
Use this skill when designing or implementing REST APIs, HTTP handlers, DTOs, or any API-related code.

---

## REST API Principles

### HTTP Methods (Semantic Usage)

```
POST   - Create new resource
GET    - Retrieve resource(s)
PUT    - Replace entire resource
PATCH  - Partial update
DELETE - Remove resource
```

**For this project (M2 - Asset Management):**
```
POST   /api/v1/batteries           → Create battery
GET    /api/v1/batteries/:id       → Get specific battery
GET    /api/v1/batteries           → List batteries
PUT    /api/v1/batteries/:id       → Replace battery (future)
PATCH  /api/v1/batteries/:id       → Update fields (future)
DELETE /api/v1/batteries/:id       → Delete battery (future)
```

---

### URL Design

**✅ GOOD - RESTful URLs**
```
/api/v1/batteries              # Collection
/api/v1/batteries/:id          # Specific resource
/api/v1/batteries/:id/status   # Sub-resource
/api/v1/batteries?location=NSW # Query parameters for filtering
```

**❌ BAD URLs**
```
/api/v1/getBatteries           # ❌ Verb in URL (use GET method instead)
/api/v1/battery                # ❌ Singular (use plural)
/api/v1/batteries_by_location  # ❌ Snake_case (use query params)
/api/v1/batteries/delete/123   # ❌ Action in URL (use DELETE method)
```

**Rules:**
1. Use nouns, not verbs (verbs are HTTP methods)
2. Use plural for collections
3. Use kebab-case or camelCase (consistent)
4. Version your API (/api/v1/)
5. Keep it simple and predictable

---

### HTTP Status Codes

**Success (2xx):**
```
200 OK              - Successful GET, PUT, PATCH
201 Created         - Successful POST (resource created)
204 No Content      - Successful DELETE (no body)
```

**Client Errors (4xx):**
```
400 Bad Request     - Validation error, malformed JSON
401 Unauthorized    - Missing/invalid authentication
403 Forbidden       - Authenticated but no permission
404 Not Found       - Resource doesn't exist
409 Conflict        - Resource already exists
422 Unprocessable   - Semantic validation error
```

**Server Errors (5xx):**
```
500 Internal Error  - Unexpected server error
503 Service Unavailable - Temporary outage
```

**For this project:**
```go
// GOOD - map domain errors to HTTP status
func (h *Handler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    battery, err := h.service.CreateBattery(...)
    if err != nil {
        statusCode, errResp := mapDomainError(err)
        writeJSON(w, statusCode, errResp)
        return
    }
    writeJSON(w, http.StatusCreated, battery)  // 201 for creation
}

func mapDomainError(err error) (int, ErrorResponse) {
    switch {
    case errors.Is(err, domain.ErrInvalidCapacity):
        return 400, ErrorResponse{Code: "VALIDATION_ERROR", Message: err.Error()}
    case errors.Is(err, domain.ErrBatteryNotFound):
        return 404, ErrorResponse{Code: "NOT_FOUND", Message: "Battery not found"}
    default:
        return 500, ErrorResponse{Code: "INTERNAL_ERROR", Message: "Internal server error"}
    }
}
```

---

## Request/Response Design

### Request Bodies (DTOs)

**✅ GOOD - Clear DTOs**
```go
// CreateBatteryRequest - separate from domain model
type CreateBatteryRequest struct {
    Capacity     float64            `json:"capacity" validate:"required,gt=0"`
    MaxPower     float64            `json:"maxPower" validate:"required,gt=0"`
    RampRate     float64            `json:"rampRate" validate:"required,gt=0"`
    Efficiency   float64            `json:"efficiency" validate:"required,gte=0,lte=1"`
    Location     string             `json:"location" validate:"required,oneof=NSW VIC QLD SA TAS"`
    Manufacturer string             `json:"manufacturer" validate:"required"`
    Constraints  ConstraintsRequest `json:"constraints" validate:"required"`
}

type ConstraintsRequest struct {
    WarrantyEOL    float64  `json:"warrantyEOL" validate:"required,gte=0,lte=1"`
    MaxCycles      int      `json:"maxCycles" validate:"required,gt=0"`
    TempMin        float64  `json:"tempMin" validate:"required"`
    TempMax        float64  `json:"tempMax" validate:"required,gtfield=TempMin"`
    GridCompliance []string `json:"gridCompliance"`
}
```

**Key points:**
- JSON tags in camelCase (JavaScript convention)
- Validation tags for automatic validation
- Separate DTO from domain model (loose coupling)

### Response Bodies

**✅ GOOD - Consistent response structure**
```go
// Success response
type BatteryResponse struct {
    ID           string              `json:"id"`
    Capacity     float64             `json:"capacity"`
    MaxPower     float64             `json:"maxPower"`
    RampRate     float64             `json:"rampRate"`
    Efficiency   float64             `json:"efficiency"`
    Location     string              `json:"location"`
    Manufacturer string              `json:"manufacturer"`
    Constraints  ConstraintsResponse `json:"constraints"`
    Status       string              `json:"status"`
    CreatedAt    time.Time           `json:"createdAt"`
    UpdatedAt    time.Time           `json:"updatedAt"`
}

// Error response
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string   `json:"code"`           // Machine-readable
    Message string   `json:"message"`        // Human-readable
    Details []string `json:"details,omitempty"` // Optional details
}
```

**Example responses:**
```json
// Success (201)
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "capacity": 200.0,
  "maxPower": 100.0,
  "status": "REGISTERED",
  "createdAt": "2025-01-15T10:30:00Z"
}

// Validation error (400)
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      "capacity must be greater than 0",
      "efficiency must be between 0 and 1"
    ]
  }
}

// Not found (404)
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Battery not found"
  }
}
```

---

## Pagination

**For list endpoints:**
```go
type ListBatteriesRequest struct {
    Limit  int    `json:"limit"`   // Max results (default: 50, max: 100)
    Offset int    `json:"offset"`  // Skip N results
    Location string `json:"location"` // Filter
    Status   string `json:"status"`   // Filter
}

type ListBatteriesResponse struct {
    Batteries []BatteryResponse `json:"batteries"`
    Total     int               `json:"total"`     // Total count
    Limit     int               `json:"limit"`
    Offset    int               `json:"offset"`
    HasMore   bool              `json:"hasMore"`   // More pages?
}
```

**URL example:**
```
GET /api/v1/batteries?limit=10&offset=20&location=NSW&status=OPERATIONAL
```

---

## Validation

### Input Validation Strategy

**Layer 1: JSON Schema Validation**
```go
// Automatic with validation tags
type CreateBatteryRequest struct {
    Capacity float64 `json:"capacity" validate:"required,gt=0"`
    // Catches: missing, wrong type, basic constraints
}
```

**Layer 2: Domain Validation**
```go
// Business rules in domain
func (b *Battery) Validate() error {
    if b.MaxPower > b.Capacity {
        return ErrMaxPowerExceedsCapacity
    }
    // Catches: complex business rules
}
```

**Full validation flow:**
```go
func (h *Handler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    var req CreateBatteryRequest
    
    // Parse JSON
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, 400, "INVALID_JSON", "Malformed JSON")
        return
    }
    
    // Layer 1: Struct validation
    if err := h.validator.Struct(req); err != nil {
        writeError(w, 400, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // Convert DTO to domain
    battery := requestToDomain(req)
    
    // Layer 2: Domain validation
    if err := battery.Validate(); err != nil {
        writeError(w, 400, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // Save
    if err := h.repo.Create(r.Context(), battery); err != nil {
        writeError(w, 500, "INTERNAL_ERROR", "Failed to create battery")
        return
    }
    
    writeJSON(w, 201, domainToResponse(battery))
}
```

---

## Error Handling

### Consistent Error Format

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string   `json:"code"`    // VALIDATION_ERROR, NOT_FOUND, etc.
    Message string   `json:"message"` // Human-readable description
    Details []string `json:"details,omitempty"` // Optional additional info
}
```

### Error Response Examples

**Validation error (multiple issues):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      "capacity must be greater than 0",
      "maxPower cannot exceed capacity",
      "location must be NSW, VIC, QLD, SA, or TAS"
    ]
  }
}
```

**Not found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Battery with ID '123' not found"
  }
}
```

**Internal error (don't leak details!):**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An internal error occurred"
  }
}
```

---

## DTO ↔ Domain Conversion

**Separation of concerns:** API DTOs ≠ Domain models

```go
// Request DTO → Domain
func requestToDomain(req CreateBatteryRequest) *domain.Battery {
    return &domain.Battery{
        ID:           uuid.New().String(),
        Capacity:     req.Capacity,
        MaxPower:     req.MaxPower,
        RampRate:     req.RampRate,
        Efficiency:   req.Efficiency,
        Location:     req.Location,
        Manufacturer: req.Manufacturer,
        Constraints: domain.Constraints{
            WarrantyEOL:    req.Constraints.WarrantyEOL,
            MaxCycles:      req.Constraints.MaxCycles,
            TempMin:        req.Constraints.TempMin,
            TempMax:        req.Constraints.TempMax,
            GridCompliance: req.Constraints.GridCompliance,
        },
        Status:    domain.StatusRegistered,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// Domain → Response DTO
func domainToResponse(battery *domain.Battery) BatteryResponse {
    return BatteryResponse{
        ID:           battery.ID,
        Capacity:     battery.Capacity,
        MaxPower:     battery.MaxPower,
        RampRate:     battery.RampRate,
        Efficiency:   battery.Efficiency,
        Location:     battery.Location,
        Manufacturer: battery.Manufacturer,
        Constraints: ConstraintsResponse{
            WarrantyEOL:    battery.Constraints.WarrantyEOL,
            MaxCycles:      battery.Constraints.MaxCycles,
            TempMin:        battery.Constraints.TempMin,
            TempMax:        battery.Constraints.TempMax,
            GridCompliance: battery.Constraints.GridCompliance,
        },
        Status:    string(battery.Status),
        CreatedAt: battery.CreatedAt,
        UpdatedAt: battery.UpdatedAt,
    }
}
```

**Why separate?**
- API can evolve independently of domain
- Domain stays pure (no JSON tags, validation tags)
- Different representations (e.g., list view vs detail view)

---

## Content Negotiation

**Always set Content-Type:**
```go
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(data)
}
```

**Check Content-Type on requests:**
```go
func (h *Handler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("Content-Type") != "application/json" {
        writeError(w, 415, "UNSUPPORTED_MEDIA_TYPE", "Expected application/json")
        return
    }
    // ...
}
```

---

## Security Considerations

### Input Validation
```go
// GOOD - validate all inputs
func (h *Handler) GetBattery(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    
    // Validate UUID format
    if _, err := uuid.Parse(id); err != nil {
        writeError(w, 400, "INVALID_ID", "Invalid battery ID format")
        return
    }
    
    battery, err := h.repo.FindByID(r.Context(), id)
    // ...
}
```

### Don't Leak Internal Details
```go
// BAD - leaking database errors
func (h *Handler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    err := h.repo.Create(...)
    if err != nil {
        // ❌ Exposes database details
        writeError(w, 500, "DB_ERROR", err.Error())
    }
}

// GOOD - generic error
func (h *Handler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    err := h.repo.Create(...)
    if err != nil {
        // Log detailed error internally
        log.Error("database error", "err", err)
        // Return generic error to client
        writeError(w, 500, "INTERNAL_ERROR", "Failed to create battery")
    }
}
```

---

## Testing HTTP Handlers

### Use httptest
```go
func TestHandler_CreateBattery_Success(t *testing.T) {
    // Setup
    mockRepo := new(MockRepository)
    handler := NewHandler(mockRepo)
    
    reqBody := `{
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
            "tempMax": 45.0
        }
    }`
    
    req := httptest.NewRequest("POST", "/api/v1/batteries", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Battery")).
        Return(nil)
    
    // Execute
    handler.CreateBattery(w, req)
    
    // Assert
    assert.Equal(t, 201, w.Code)
    
    var resp BatteryResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, 200.0, resp.Capacity)
    assert.Equal(t, "REGISTERED", resp.Status)
}
```

### Test Error Cases
```go
func TestHandler_CreateBattery_InvalidJSON(t *testing.T) {
    handler := NewHandler(nil)
    
    req := httptest.NewRequest("POST", "/api/v1/batteries", 
        strings.NewReader(`{invalid json`))
    w := httptest.NewRecorder()
    
    handler.CreateBattery(w, req)
    
    assert.Equal(t, 400, w.Code)
    
    var resp ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, "INVALID_JSON", resp.Error.Code)
}

func TestHandler_GetBattery_NotFound(t *testing.T) {
    mockRepo := new(MockRepository)
    handler := NewHandler(mockRepo)
    
    mockRepo.On("FindByID", mock.Anything, "123").
        Return(nil, domain.ErrBatteryNotFound)
    
    req := httptest.NewRequest("GET", "/api/v1/batteries/123", nil)
    w := httptest.NewRecorder()
    
    handler.GetBattery(w, req)
    
    assert.Equal(t, 404, w.Code)
}
```

---

## API Documentation

### Document with Examples
```go
// CreateBattery registers a new battery in the system.
//
// POST /api/v1/batteries
//
// Request:
//   {
//     "capacity": 200.0,
//     "maxPower": 100.0,
//     "location": "NSW",
//     ...
//   }
//
// Success Response (201):
//   {
//     "id": "uuid",
//     "capacity": 200.0,
//     "status": "REGISTERED",
//     ...
//   }
//
// Error Responses:
//   400 - Validation error
//   500 - Internal error
func (h *Handler) CreateBattery(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

---

## Best Practices Summary

**✅ DO:**
- Use proper HTTP methods and status codes
- Validate all inputs (JSON + domain)
- Return consistent error format
- Separate DTOs from domain models
- Test all endpoints (success + errors)
- Version your API (/api/v1/)
- Use plural nouns for collections
- Set Content-Type headers

**❌ DON'T:**
- Expose internal errors to clients
- Use verbs in URLs
- Mix domain and API concerns
- Return 200 for errors
- Forget to validate IDs
- Leak database details
- Use snake_case in JSON (use camelCase)

---

## Quick Reference

```
POST   → 201 Created (with body)
GET    → 200 OK
PUT    → 200 OK
PATCH  → 200 OK
DELETE → 204 No Content

Validation error → 400 Bad Request
Not found        → 404 Not Found
Conflict         → 409 Conflict
Server error     → 500 Internal Error

URL: /api/v1/batteries
     └─ version
              └─ resource (plural)

JSON: camelCase
Domain: MixedCaps or mixedCaps
```

---

**Remember: A good API is predictable, consistent, and follows REST principles.**
