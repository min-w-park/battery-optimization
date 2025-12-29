# TDD (Test-Driven Development) Skill

## When to use this skill
This skill should be used whenever writing Go code for the battery optimization project. It enforces Kent Beck's Test-Driven Development methodology.

## Core Principles

### The Three Laws of TDD
1. **No production code without a failing test**
2. **Write only enough test to fail**
3. **Write only enough code to pass the test**

### The Rhythm: Red → Green → Refactor

```
RED:    Write a failing test
GREEN:  Write minimal code to pass
REFACTOR: Clean up (tests still green)
```

---

## Mandatory Workflow

### Step 1: Understand the Requirement
Before writing ANY code:
- Read the relevant spec (e.g., `docs/milestones/M2-DOMAIN-SPEC.md`)
- Identify ONE specific behavior to implement
- Ask: "What's the simplest test for this?"

### Step 2: Write the Test First (RED)
```go
func TestBattery_Validate_NegativeCapacity_ReturnsError(t *testing.T) {
    // Arrange
    battery := &Battery{Capacity: -100.0}
    
    // Act
    err := battery.Validate()
    
    // Assert
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}
```

**Key points:**
- Test name format: `Test<Type>_<Method>_<Scenario>_<Expected>`
- Use Arrange-Act-Assert structure
- Test ONE thing only
- Write the test you WISH you had (even if types don't exist yet)

### Step 3: Run the Test (Verify RED)
```bash
go test ./internal/domain/...
# Should FAIL with clear error
```

**If it doesn't fail correctly, fix the test first**

### Step 4: Write Minimal Code (GREEN)
```go
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    return nil  // JUST enough to pass this ONE test
}
```

**Rules:**
- Write the SIMPLEST code that passes
- It's OK to "fake it" initially
- Don't implement features not yet tested
- No premature optimization

### Step 5: Run Test Again (Verify GREEN)
```bash
go test ./internal/domain/...
# Should PASS
```

### Step 6: Refactor (Keep GREEN)
Now that tests protect you, clean up:
- Extract functions
- Improve names
- Remove duplication
- Simplify logic

**Run tests after EVERY refactor!**

### Step 7: Commit
```bash
git add .
git commit -m "test(domain): add negative capacity validation"
```

### Step 8: Repeat
Go back to Step 1 for the next behavior.

---

## Test Patterns

### Pattern 1: Table-Driven Tests (for multiple cases)
```go
func TestBattery_Validate_InvalidInputs(t *testing.T) {
    tests := []struct {
        name     string
        capacity float64
        wantErr  error
    }{
        {"negative", -100.0, ErrInvalidCapacity},
        {"zero", 0.0, ErrInvalidCapacity},
        {"too large", 1000000.0, ErrCapacityTooLarge},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            battery := &Battery{Capacity: tt.capacity}
            
            err := battery.Validate()
            
            assert.ErrorIs(t, err, tt.wantErr)
        })
    }
}
```

### Pattern 2: Given-When-Then (clarity)
```go
func TestNewBattery_ValidInput_ReturnsNoBatteryNoError(t *testing.T) {
    // Given (Arrange)
    capacity := 200.0
    maxPower := 100.0
    
    // When (Act)
    battery, err := NewBattery(capacity, maxPower, ...)
    
    // Then (Assert)
    assert.NoError(t, err)
    assert.NotNil(t, battery)
    assert.Equal(t, capacity, battery.Capacity)
}
```

### Pattern 3: Test Behavior, Not Implementation
```go
// ❌ BAD - tests internal details
func TestBattery_HasValidationFlag(t *testing.T) {
    battery.Validate()
    assert.True(t, battery.wasValidated) // internal state
}

// ✅ GOOD - tests observable behavior
func TestBattery_Validate_ReturnsNoError(t *testing.T) {
    err := battery.Validate()
    assert.NoError(t, err) // public outcome
}
```

---

## What to Test

### ✅ DO Test:
- Business logic
- Validation rules
- Error cases
- Edge cases (boundary conditions)
- State transitions

### ❌ DON'T Test:
- Simple getters/setters (no logic)
- Framework code (trust the framework)
- Third-party libraries
- Obvious code (e.g., `return true`)

---

## Example: Building Battery Validation Step-by-Step

### Iteration 1: Happy Path
```go
// Test
func TestBattery_Validate_ValidBattery_ReturnsNoError(t *testing.T) {
    battery := &Battery{Capacity: 200.0, MaxPower: 100.0}
    err := battery.Validate()
    assert.NoError(t, err)
}

// Implementation (fake it!)
func (b *Battery) Validate() error {
    return nil
}
```

### Iteration 2: First Validation
```go
// Test
func TestBattery_Validate_NegativeCapacity_ReturnsError(t *testing.T) {
    battery := &Battery{Capacity: -100.0}
    err := battery.Validate()
    assert.ErrorIs(t, err, ErrInvalidCapacity)
}

// Implementation (make it real)
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    return nil
}
```

### Iteration 3: Second Validation
```go
// Test
func TestBattery_Validate_MaxPowerExceedsCapacity_ReturnsError(t *testing.T) {
    battery := &Battery{Capacity: 100.0, MaxPower: 150.0}
    err := battery.Validate()
    assert.ErrorIs(t, err, ErrMaxPowerExceedsCapacity)
}

// Implementation (add check)
func (b *Battery) Validate() error {
    if b.Capacity <= 0 {
        return ErrInvalidCapacity
    }
    if b.MaxPower > b.Capacity {
        return ErrMaxPowerExceedsCapacity
    }
    return nil
}
```

### Iteration 4: Refactor
```go
// Tests still green, now clean up
func (b *Battery) Validate() error {
    if err := b.validateCapacity(); err != nil {
        return err
    }
    if err := b.validatePower(); err != nil {
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

func (b *Battery) validatePower() error {
    if b.MaxPower > b.Capacity {
        return ErrMaxPowerExceedsCapacity
    }
    return nil
}
```

---

## Common Mistakes to Avoid

### Mistake 1: Writing Implementation First
```go
// ❌ WRONG ORDER
1. Write Battery.Validate() code
2. Write tests to verify it works

// ✅ CORRECT ORDER
1. Write test for Battery.Validate()
2. Watch it fail
3. Write code to pass
```

### Mistake 2: Testing Too Much at Once
```go
// ❌ BAD - tests multiple things
func TestBattery_Validate(t *testing.T) {
    battery := &Battery{
        Capacity: -100,    // wrong
        MaxPower: 200,     // wrong
        Efficiency: 1.5,   // wrong
    }
    err := battery.Validate()
    assert.Error(t, err)  // Which error?!
}

// ✅ GOOD - one test per scenario
func TestBattery_Validate_NegativeCapacity(t *testing.T) { ... }
func TestBattery_Validate_InvalidMaxPower(t *testing.T) { ... }
func TestBattery_Validate_InvalidEfficiency(t *testing.T) { ... }
```

### Mistake 3: Skipping Refactor
```go
// Tests are green but code is ugly
func (b *Battery) Validate() error {
    if b.Capacity <= 0 { return ErrInvalidCapacity }
    if b.MaxPower <= 0 { return ErrInvalidMaxPower }
    if b.MaxPower > b.Capacity { return ErrMaxPowerExceedsCapacity }
    if b.Efficiency < 0 || b.Efficiency > 1 { return ErrInvalidEfficiency }
    if b.RampRate <= 0 { return ErrInvalidRampRate }
    // ... gets worse ...
    return nil
}

// ✅ Refactor NOW while tests protect you
func (b *Battery) Validate() error {
    validators := []func() error{
        b.validateCapacity,
        b.validatePower,
        b.validateEfficiency,
        b.validateRampRate,
    }
    
    for _, validate := range validators {
        if err := validate(); err != nil {
            return err
        }
    }
    return nil
}
```

---

## Integration with Project

### File Locations
- Tests go next to implementation: `battery.go` → `battery_test.go`
- Same package (can test unexported functions)
- Test files have `_test.go` suffix

### Running Tests
```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/domain/...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestBattery_Validate ./internal/domain/...

# Verbose output
go test -v ./...
```

### Coverage Goals
- Aim for >80% coverage
- 100% is not always necessary
- Focus on business logic

---

## Quick Reference Card

```
┌─────────────────────────────────────────────┐
│  TDD Cycle - ALWAYS FOLLOW THIS ORDER      │
├─────────────────────────────────────────────┤
│  1. ❌ RED:  Write failing test             │
│  2. ✅ GREEN: Minimal code to pass          │
│  3. ♻️  REFACTOR: Clean up (keep green)     │
│  4. 🔁 REPEAT: Next behavior                │
└─────────────────────────────────────────────┘

REMEMBER:
- Tests document what code should do
- Tests enable fearless refactoring
- If hard to test → probably bad design
- Commit on green
```

---

## When You're Stuck

1. **Make the test smaller** - test one tiny thing
2. **Fake the implementation** - hard-code if needed
3. **Write the obvious test** - even if boring
4. **Take a break** - TDD requires focus

---

## Project-Specific Notes

For this battery optimization project:

- **Domain layer** (internal/domain): Pure TDD, 100% tested
- **Repository layer**: Mock for unit tests, real DB for integration tests  
- **HTTP layer**: Use httptest for handler tests
- **End-to-end**: Separate integration tests (optional)

---

## Success Metrics

You're doing TDD right when:
- ✅ Every function has tests written FIRST
- ✅ Tests pass immediately after implementation
- ✅ You refactor confidently (tests protect you)
- ✅ Coverage is >80% without thinking about it
- ✅ Code is clean (refactor step works)

---

**Remember: TDD is not about testing. It's about design.**

Write the test you wish you had. Then make it work.
