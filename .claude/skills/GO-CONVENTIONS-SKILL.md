# Go Conventions Skill

## When to use this skill
Use this skill when writing any Go code for this project. It ensures idiomatic Go and consistent code style.

---

## Go Idioms and Best Practices

### 1. Error Handling

**✅ ALWAYS check errors**
```go
// GOOD
data, err := readFile("config.json")
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}

// BAD - ignoring errors
data, _ := readFile("config.json")  // ❌ Never do this
```

**Use %w for error wrapping (Go 1.13+)**
```go
// GOOD - preserves error chain
if err != nil {
    return fmt.Errorf("failed to validate battery: %w", err)
}

// Now caller can use errors.Is() and errors.As()
if errors.Is(err, domain.ErrInvalidCapacity) {
    // handle specific error
}
```

**Return errors, don't panic**
```go
// GOOD
func NewBattery(capacity float64) (*Battery, error) {
    if capacity <= 0 {
        return nil, ErrInvalidCapacity
    }
    return &Battery{Capacity: capacity}, nil
}

// BAD - panic in library code
func NewBattery(capacity float64) *Battery {
    if capacity <= 0 {
        panic("invalid capacity")  // ❌ Only panic for programmer errors
    }
    return &Battery{Capacity: capacity}
}
```

**When to panic:**
- Never in library code
- Only in `main()` or `init()` for unrecoverable setup errors
- Only for programmer errors (e.g., nil pointer that should never happen)

---

### 2. Naming Conventions

**Packages**
```go
// GOOD - lowercase, single word, no underscores
package domain
package postgres
package http

// BAD
package Domain          // ❌ Don't capitalize
package battery_domain  // ❌ No underscores
package batteryDomain   // ❌ No camelCase
```

**Interfaces**
```go
// GOOD - use -er suffix for single-method interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type BatteryRepository interface {  // OK - descriptive name
    Create(ctx context.Context, battery *Battery) error
    FindByID(ctx context.Context, id string) (*Battery, error)
}

// BAD
type IBatteryRepository interface {}  // ❌ No "I" prefix
type BatteryRepositoryInterface {}    // ❌ No "Interface" suffix
```

**Variables**
```go
// GOOD - camelCase, descriptive
batteryID := "123"
maxPower := 100.0
userConfig := loadConfig()

// BAD
battery_id := "123"     // ❌ No snake_case
BatteryID := "123"      // ❌ Only export if needed
mp := 100.0            // ❌ Not descriptive enough
```

**Constants**
```go
// GOOD - MixedCaps (exported) or mixedCaps (unexported)
const MaxRetries = 3
const defaultTimeout = 30 * time.Second

// BAD
const MAX_RETRIES = 3   // ❌ No SCREAMING_SNAKE_CASE
const max_retries = 3   // ❌ No snake_case
```

**Acronyms - keep consistent case**
```go
// GOOD
var httpClient *http.Client  // unexported: all lowercase
var HTTPClient *http.Client  // exported: all uppercase

var batteryID string         // unexported: camelCase with lowercase ID
var BatteryID string         // exported: all caps ID

// BAD
var Http *http.Client        // ❌ Inconsistent
var batteryId string         // ❌ Should be ID not Id
```

---

### 3. Function and Method Design

**Receiver names**
```go
// GOOD - short, consistent (usually 1-2 letters)
func (b *Battery) Validate() error { }
func (b *Battery) SetStatus(s Status) { }

// BAD
func (battery *Battery) Validate() error { }     // ❌ Too long
func (this *Battery) Validate() error { }        // ❌ Not Go style
func (b *Battery) Validate() error { }
func (bat *Battery) SetStatus(s Status) { }      // ❌ Inconsistent receiver name
```

**Pointer vs Value receivers**
```go
// Use pointer receiver if:
// 1. Method modifies the receiver
// 2. Receiver is large struct
// 3. Consistency (if one method uses pointer, all should)

// GOOD - modifies receiver
func (b *Battery) SetStatus(status BatteryStatus) {
    b.Status = status
    b.UpdatedAt = time.Now()
}

// GOOD - large struct, avoid copying
func (b *Battery) Validate() error {  // Battery is large
    // ...
}

// GOOD - small, immutable
func (c Constraints) IsValid() bool {  // Value receiver OK for small types
    return c.TempMin < c.TempMax
}
```

**Function parameters**
```go
// GOOD - context first, options last
func Create(ctx context.Context, battery *Battery, opts ...Option) error

// GOOD - related parameters grouped
func NewBattery(capacity, maxPower, rampRate float64) (*Battery, error)

// BAD - too many parameters (use struct or options)
func NewBattery(cap, pow, ramp, eff, temp1, temp2 float64, loc string, ...) // ❌ Too many
```

**Return values**
```go
// GOOD - return (value, error)
func FindByID(id string) (*Battery, error)

// GOOD - return error last
func Validate(battery *Battery) error

// BAD - inconsistent error position
func FindByID(id string) (error, *Battery)  // ❌ Error should be last
```

---

### 4. Structs and Composition

**Struct tags**
```go
// GOOD - json tags in lowercase_snake_case
type Battery struct {
    ID       string  `json:"id"`
    MaxPower float64 `json:"max_power"`
}

// GOOD - db tags match database columns
type Battery struct {
    ID       string  `db:"id"`
    MaxPower float64 `db:"max_power"`
}

// GOOD - multiple tags
type Battery struct {
    ID       string  `json:"id" db:"id" validate:"required"`
    MaxPower float64 `json:"max_power" db:"max_power" validate:"gt=0"`
}
```

**Zero values**
```go
// Design structs so zero value is useful
type Config struct {
    Timeout time.Duration  // Zero value (0) means no timeout
    Retries int           // Zero value (0) means no retries
}

// GOOD - can use without initialization
var cfg Config  // Usable with zero values

// If zero value is not useful, provide constructor
func NewBattery(...) (*Battery, error) {
    // Don't allow zero Battery - constructor required
}
```

**Embedding**
```go
// GOOD - embed for "is-a" relationship
type TimestampedBattery struct {
    *Battery         // Embedded - promotes Battery's methods
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Now can call Battery methods directly
tb := &TimestampedBattery{Battery: battery}
err := tb.Validate()  // Calls Battery.Validate()

// Don't embed if you need "has-a" (composition)
type BatteryService struct {
    repo BatteryRepository  // Has-a, not is-a
}
```

---

### 5. Concurrency Patterns

**Context usage**
```go
// GOOD - pass context as first parameter
func (r *Repository) Create(ctx context.Context, battery *Battery) error {
    // Use ctx for cancellation and deadlines
    return r.db.QueryContext(ctx, ...)
}

// GOOD - check context cancellation
func (s *Service) Process(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // continue processing
    }
}
```

**Goroutines**
```go
// GOOD - use sync.WaitGroup
var wg sync.WaitGroup
for _, battery := range batteries {
    wg.Add(1)
    go func(b *Battery) {
        defer wg.Done()
        process(b)
    }(battery)  // Pass as parameter, not closure
}
wg.Wait()

// BAD - closure over loop variable
for _, battery := range batteries {
    go func() {
        process(battery)  // ❌ All goroutines see last value
    }()
}
```

**Channels**
```go
// GOOD - close channels from sender
func generateBatteries() <-chan *Battery {
    ch := make(chan *Battery)
    go func() {
        defer close(ch)  // Sender closes
        for _, b := range batteries {
            ch <- b
        }
    }()
    return ch
}

// GOOD - check if channel is closed
for battery := range ch {  // Loop exits when ch closes
    process(battery)
}
```

---

### 6. Testing Conventions

**Test function names**
```go
// GOOD - descriptive
func TestBattery_Validate_NegativeCapacity_ReturnsError(t *testing.T)
func TestRepository_Create_DuplicateID_ReturnsError(t *testing.T)

// BAD
func TestValidation(t *testing.T)  // ❌ Not specific enough
func Test1(t *testing.T)           // ❌ Meaningless name
```

**Table-driven tests**
```go
func TestBattery_Validate(t *testing.T) {
    tests := []struct {
        name     string
        battery  *Battery
        wantErr  error
    }{
        {
            name:    "negative capacity",
            battery: &Battery{Capacity: -100},
            wantErr: ErrInvalidCapacity,
        },
        {
            name:    "valid battery",
            battery: &Battery{Capacity: 200, MaxPower: 100},
            wantErr: nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.battery.Validate()
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

**Test helpers**
```go
// GOOD - use t.Helper()
func createTestBattery(t *testing.T) *Battery {
    t.Helper()  // Makes test failure point to caller
    battery, err := NewBattery(200, 100, ...)
    if err != nil {
        t.Fatalf("failed to create test battery: %v", err)
    }
    return battery
}
```

---

### 7. Package Organization

**Internal package**
```
services/asset-management/
├── cmd/                    # Executables
│   └── server/
│       └── main.go
├── internal/              # Private to this service
│   ├── domain/           # Core business logic
│   ├── ports/            # Interfaces
│   └── adapters/         # Implementations
└── pkg/                  # Public libraries (optional)
```

**Use internal/ to prevent external imports**
```go
// internal/domain/battery.go
package domain

type Battery struct { ... }  // Can't be imported outside this module
```

---

### 8. Comments and Documentation

**Package comments**
```go
// Package domain implements the core battery optimization business logic.
// It provides the Battery aggregate and related domain entities.
package domain
```

**Exported function comments**
```go
// NewBattery creates a new Battery with the given specifications.
// It validates all inputs and returns an error if any constraint is violated.
//
// Parameters:
//   - capacity: Energy storage capacity in MWh (must be > 0)
//   - maxPower: Maximum charge/discharge rate in MW (must be <= capacity)
//
// Returns an error if validation fails.
func NewBattery(capacity, maxPower float64) (*Battery, error) {
    // ...
}
```

**Don't comment obvious code**
```go
// BAD - obvious comments
// Set capacity to input capacity
battery.Capacity = capacity  // ❌ Comment adds no value

// GOOD - explain WHY, not WHAT
// MaxPower must not exceed Capacity to prevent thermal damage
if maxPower > capacity {
    return ErrMaxPowerExceedsCapacity
}
```

---

### 9. Common Go Patterns

**Options pattern (for complex constructors)**
```go
type BatteryOption func(*Battery)

func WithManufacturer(m string) BatteryOption {
    return func(b *Battery) {
        b.Manufacturer = m
    }
}

func NewBattery(capacity float64, opts ...BatteryOption) *Battery {
    b := &Battery{Capacity: capacity}
    for _, opt := range opts {
        opt(b)
    }
    return b
}

// Usage
battery := NewBattery(200, 
    WithManufacturer("Tesla"),
    WithLocation("NSW"),
)
```

**Functional options for configs**
```go
type Config struct {
    timeout time.Duration
    retries int
}

type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) { c.timeout = d }
}

func NewService(opts ...Option) *Service {
    cfg := &Config{
        timeout: 30 * time.Second,  // Default
        retries: 3,
    }
    for _, opt := range opts {
        opt(cfg)
    }
    return &Service{config: cfg}
}
```

---

### 10. Code Organization

**One type per file (for important types)**
```
internal/domain/
├── battery.go         # Battery type and methods
├── constraints.go     # Constraints type
├── errors.go          # Domain errors
└── events.go          # Domain events
```

**Group related functionality**
```go
// battery.go
type Battery struct { ... }
func NewBattery(...) (*Battery, error) { ... }
func (b *Battery) Validate() error { ... }

// Don't scatter Battery methods across multiple files
```

---

### 11. Performance Tips

**String concatenation**
```go
// GOOD - use strings.Builder for loops
var sb strings.Builder
for _, s := range strings {
    sb.WriteString(s)
}
result := sb.String()

// BAD - string concatenation in loop
result := ""
for _, s := range strings {
    result += s  // ❌ Allocates on every iteration
}
```

**Slice preallocation**
```go
// GOOD - preallocate if size known
batteries := make([]*Battery, 0, expectedCount)

// OK - let it grow if size unknown
batteries := []*Battery{}
```

**Defer in loops**
```go
// BAD - defer in loop (defers pile up)
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close()  // ❌ All close at end of function
}

// GOOD - use function to scope defer
for _, file := range files {
    func() {
        f, _ := os.Open(file)
        defer f.Close()  // Closes each iteration
        process(f)
    }()
}
```

---

## Project-Specific Conventions

### File naming
```
battery.go           # Types
battery_test.go      # Tests
battery_mock.go      # Mocks (if using mockgen)
```

### Import grouping
```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External dependencies
    "github.com/google/uuid"
    "github.com/lib/pq"
    
    // Internal packages
    "github.com/user/battery-optimization/internal/domain"
    "github.com/user/battery-optimization/internal/ports"
)
```

### Error variables
```go
// Prefix with Err
var (
    ErrInvalidCapacity = errors.New("invalid capacity")
    ErrBatteryNotFound = errors.New("battery not found")
)
```

---

## Tools to Use

**Before committing:**
```bash
# Format code
go fmt ./...

# Vet for issues
go vet ./...

# Run tests
go test ./...

# Optional: golangci-lint
golangci-lint run
```

---

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

---

**Remember: Idiomatic Go is simple, clear, and readable. When in doubt, choose clarity over cleverness.**
