# Test-Driven Development Guide (Kent Beck Style)

> "I'm not a great programmer; I'm just a good programmer with great habits." - Kent Beck

## The TDD Mantra

**Red → Green → Refactor**

This is not just a process. It's a rhythm. A way of thinking.

---

## The Three Laws of TDD

### 1. You are not allowed to write any production code unless it is to make a failing unit test pass.

**Wrong**:
```go
// Writing implementation first
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    // ... more validation
}
```

**Right**:
```go
// Write the test first
func TestBattery_Validate_NegativeCapacity(t *testing.T) {
    battery := &Battery{Capacity: -100}
    
    err := battery.Validate()
    
    assert.Error(t, err)
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}

// NOW write the minimal code to pass
```

### 2. You are not allowed to write any more of a unit test than is sufficient to fail.

**Wrong**:
```go
func TestBattery_Validate(t *testing.T) {
    // Testing everything at once
    battery := &Battery{
        Capacity: -100,    // Wrong
        MaxPower: -50,     // Wrong
        Efficiency: 1.5,   // Wrong
    }
    err := battery.Validate()
    // Now what? Too many things to fix!
}
```

**Right**:
```go
// One test, one failure, one fix
func TestBattery_Validate_NegativeCapacity(t *testing.T) {
    battery := &Battery{Capacity: -100}
    err := battery.Validate()
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}

// Next test after this passes
func TestBattery_Validate_ZeroCapacity(t *testing.T) {
    battery := &Battery{Capacity: 0}
    err := battery.Validate()
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}
```

### 3. You are not allowed to write any more production code than is sufficient to pass the one failing unit test.

**Wrong**:
```go
// Over-engineering to handle future cases
func (b *Battery) Validate() error {
    var errors []error
    
    // Implementing all validations even though tests don't require them yet
    if b.Capacity <= 0 { errors = append(errors, ErrInvalidCapacity) }
    if b.MaxPower <= 0 { errors = append(errors, ErrInvalidMaxPower) }
    if b.Efficiency < 0 || b.Efficiency > 1 { errors = append(errors, ...) }
    // ... etc
}
```

**Right**:
```go
// Minimal code to pass the current failing test
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    return nil
}
// Add more ONLY when next test requires it
```

---

## The TDD Cycle (Detailed)

### Step 1: RED - Write a Failing Test

**Think**: "What's the next simplest behavior?"

```go
func TestNewBattery_ValidCapacity(t *testing.T) {
    // Given
    capacity := 200.0
    
    // When
    battery, err := NewBattery(capacity, ...)
    
    // Then
    assert.NoError(t, err)
    assert.Equal(t, capacity, battery.Capacity)
}
```

**Run the test**: It should FAIL (function doesn't exist yet)

**Key Question**: "Can I make it fail for the right reason?"
- Good fail: "NewBattery function not defined"
- Bad fail: Random panic, unclear error

### Step 2: GREEN - Make It Pass (Quickly!)

**Goal**: Get to green as fast as possible. Don't worry about elegance yet.

```go
// Even this is OK if it makes the test pass!
func NewBattery(capacity float64, ...) (*Battery, error) {
    return &Battery{Capacity: capacity}, nil
}
```

**Kent Beck's advice**: "Make the test work quickly, committing whatever sins necessary in the process."

**Fake It Till You Make It**:
```go
// First test: battery with capacity 200
func TestNewBattery_FirstBattery(t *testing.T) {
    battery, _ := NewBattery(200.0)
    assert.Equal(t, 200.0, battery.Capacity)
}

// Simplest implementation that passes:
func NewBattery(capacity float64) (*Battery, error) {
    return &Battery{Capacity: 200.0}, nil  // Hard-coded!
}

// Next test forces you to make it real:
func TestNewBattery_DifferentCapacity(t *testing.T) {
    battery, _ := NewBattery(150.0)
    assert.Equal(t, 150.0, battery.Capacity)
}

// Now you must use the parameter:
func NewBattery(capacity float64) (*Battery, error) {
    return &Battery{Capacity: capacity}, nil
}
```

### Step 3: REFACTOR - Clean Up

**Now** make it beautiful. Tests are green, so you can refactor safely.

**Before**:
```go
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    if b.MaxPower <= 0 {
        return ErrInvalidMaxPower
    }
    if b.MaxPower > b.Capacity {
        return ErrMaxPowerExceedsCapacity
    }
    if b.Efficiency < 0 || b.Efficiency > 1 {
        return ErrInvalidEfficiency
    }
    // ... many more
    return nil
}
```

**After** (extract helper):
```go
func (b *Battery) Validate() error {
    if err := b.validateCapacity(); err != nil {
        return err
    }
    if err := b.validatePower(); err != nil {
        return err
    }
    if err := b.validateEfficiency(); err != nil {
        return err
    }
    return nil
}

func (b *Battery) validateCapacity() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    return nil
}
```

**Run tests again**: Still green? Good. Commit.

---

## Kent Beck's Test Patterns

### Pattern 1: Arrange-Act-Assert (Given-When-Then)

```go
func TestBattery_Validate_Success(t *testing.T) {
    // Arrange (Given) - Set up test data
    battery := &Battery{
        Capacity:   200.0,
        MaxPower:   100.0,
        Efficiency: 0.92,
    }
    
    // Act (When) - Execute the behavior
    err := battery.Validate()
    
    // Assert (Then) - Verify the outcome
    assert.NoError(t, err)
}
```

### Pattern 2: Table-Driven Tests (for variations)

```go
func TestBattery_Validate_InvalidInputs(t *testing.T) {
    tests := []struct {
        name          string
        capacity      float64
        expectedError error
    }{
        {
            name:          "negative capacity",
            capacity:      -100.0,
            expectedError: ErrInvalidCapacity,
        },
        {
            name:          "zero capacity",
            capacity:      0.0,
            expectedError: ErrInvalidCapacity,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            battery := &Battery{Capacity: tt.capacity}
            
            err := battery.Validate()
            
            assert.ErrorIs(t, err, tt.expectedError)
        })
    }
}
```

### Pattern 3: Test One Thing at a Time

**Wrong**:
```go
func TestBattery_Everything(t *testing.T) {
    battery, err := NewBattery(...)
    assert.NoError(t, err)
    
    err = battery.Validate()
    assert.NoError(t, err)
    
    battery.Status = StatusOperational
    assert.Equal(t, StatusOperational, battery.Status)
    
    // Too much! What if it fails?
}
```

**Right**:
```go
func TestNewBattery_ValidInput_ReturnsNoBatteryNoError(t *testing.T) {
    battery, err := NewBattery(...)
    assert.NoError(t, err)
    assert.NotNil(t, battery)
}

func TestBattery_Validate_ValidBattery_ReturnsNoError(t *testing.T) {
    battery := &Battery{...valid data...}
    err := battery.Validate()
    assert.NoError(t, err)
}

func TestBattery_SetStatus_SetsStatusCorrectly(t *testing.T) {
    battery := &Battery{}
    battery.Status = StatusOperational
    assert.Equal(t, StatusOperational, battery.Status)
}
```

### Pattern 4: Test Behavior, Not Implementation

**Wrong** (tests internal state):
```go
func TestBattery_Validate_SetsInternalFlag(t *testing.T) {
    battery := &Battery{...}
    battery.Validate()
    
    // Testing internal implementation detail
    assert.True(t, battery.wasValidated)
}
```

**Right** (tests observable behavior):
```go
func TestBattery_Validate_ValidInput_ReturnsNoError(t *testing.T) {
    battery := &Battery{...valid...}
    
    err := battery.Validate()
    
    assert.NoError(t, err)  // Observable outcome
}
```

---

## The TDD Workflow in Practice

### Starting a New Feature: Battery Validation

**Step 1**: What's the simplest test?
```go
// Start with the happy path
func TestBattery_Validate_ValidBattery_ReturnsNoError(t *testing.T) {
    battery := &Battery{Capacity: 200.0, MaxPower: 100.0}
    err := battery.Validate()
    assert.NoError(t, err)
}
```

**Step 2**: Make it pass (fake it!)
```go
func (b *Battery) Validate() error {
    return nil  // Simplest thing that works
}
```

**Step 3**: Next test (first edge case)
```go
func TestBattery_Validate_NegativeCapacity_ReturnsError(t *testing.T) {
    battery := &Battery{Capacity: -100.0}
    err := battery.Validate()
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}
```

**Step 4**: Make it pass
```go
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    return nil
}
```

**Step 5**: Continue adding tests one at a time
- Zero capacity
- Max power validation
- Efficiency validation
- ... etc

Each test adds one behavior. Each implementation adds minimal code.

---

## Kent Beck's Advice

### "Test Until Fear Turns to Boredom"

Write tests until you're confident. Not paranoid, not careless. Confident.

```go
// Confident after these tests:
func TestNewBattery_ValidInput(t *testing.T) { ... }
func TestNewBattery_NegativeCapacity(t *testing.T) { ... }
func TestNewBattery_ZeroCapacity(t *testing.T) { ... }
func TestNewBattery_MaxPowerExceedsCapacity(t *testing.T) { ... }

// Paranoid (too many):
func TestNewBattery_Capacity_0_0001(t *testing.T) { ... }
func TestNewBattery_Capacity_0_0002(t *testing.T) { ... }
// Stop. You're testing floats, not your logic.
```

### "If it's hard to test, it's probably a design problem"

```go
// Hard to test (too many dependencies)
func (s *Service) ProcessBattery(id string) error {
    battery := s.db.FindBattery(id)      // DB dependency
    price := s.aemo.GetPrice()           // External API
    decision := s.optimizer.Decide(...)  // Complex logic
    s.publisher.Publish(...)             // Event bus
    // How do you test this?!
}

// Easy to test (dependency injection)
func (s *Service) ProcessBattery(
    battery *Battery,
    price float64,
    optimizer Optimizer,
) error {
    decision := optimizer.Decide(battery, price)
    return decision
}
// Now you can test with mocks!
```

### "Test-first is not test-obsessed"

Don't test:
- ❌ Getters/setters (unless they have logic)
- ❌ Framework code (trust the framework)
- ❌ Database queries (test your logic, not SQL)

Do test:
- ✅ Business logic
- ✅ Edge cases
- ✅ Error handling
- ✅ State transitions

---

## Common Mistakes

### Mistake 1: Writing Tests After Code

**Problem**: Tests become "rubber stamps" - they just verify what you already wrote.

**Solution**: Tests first. Always.

### Mistake 2: Big Jumps

**Problem**: Writing a test that requires 50 lines of code to pass.

**Solution**: Smaller steps. If a test needs more than ~10 lines of code, break it down.

### Mistake 3: Testing Implementation

**Problem**: Tests break when you refactor, even though behavior didn't change.

```go
// Bad - tests internal structure
func TestBattery_HasValidatedFlag(t *testing.T) {
    battery.Validate()
    assert.True(t, battery.isValidated)
}

// Good - tests behavior
func TestBattery_Validate_ReturnsNoError(t *testing.T) {
    err := battery.Validate()
    assert.NoError(t, err)
}
```

### Mistake 4: Not Refactoring

**Problem**: Tests are green, code is ugly, "I'll clean it later" (you won't).

**Solution**: Refactor NOW while tests protect you.

---

## TDD for This Project

### Domain Layer (Pure Business Logic)

**Every** domain function should have tests first:

```go
// 1. Write test
func TestNewBattery_ValidInput(t *testing.T) { ... }

// 2. Implement
func NewBattery(...) (*Battery, error) { ... }

// 3. Refactor
// Extract validation, improve names, etc.
```

### Repository Layer (Database)

Use **test doubles** (mocks) for unit tests:

```go
func TestHandler_CreateBattery_Success(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRepo.On("Create", mock.Anything).Return(nil)
    
    handler := NewHandler(mockRepo)
    // Test handler without real database
}
```

### HTTP Layer (Integration)

Test with `httptest`:

```go
func TestHandler_CreateBattery_Returns201(t *testing.T) {
    req := httptest.NewRequest("POST", "/batteries", body)
    w := httptest.NewRecorder()
    
    handler.CreateBattery(w, req)
    
    assert.Equal(t, 201, w.Code)
}
```

---

## Remember

> "TDD is not about testing. It's about design." - Kent Beck

Tests are:
1. **Specification**: What should it do?
2. **Documentation**: How does it work?
3. **Safety Net**: Can I refactor safely?
4. **Design Tool**: Is my code testable? (If not, it's probably not well-designed)

---

## The Zen of TDD

- Red → Green → Refactor
- One test at a time
- Simplest thing that works
- Clean code when tests are green
- Commit on green

**Trust the process. It works.**

---

## Quick Reference

**Starting a new feature:**
1. Write simplest test (happy path)
2. Make it pass (fake it if needed)
3. Write next test (first edge case)
4. Make it pass (use real logic now)
5. Refactor
6. Repeat 3-5 until confident

**When stuck:**
- Make the test smaller
- Fake the implementation
- Write the obvious (boring) test
- Take a break

**When tests are green:**
- Refactor mercilessly
- Extract functions
- Improve names
- Remove duplication
- Commit

---

**Now go write some tests! 🧪**
