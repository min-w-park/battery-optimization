# Contributing to Battery Optimization System

## Development Philosophy

This project follows **Test-Driven Development (TDD)** as described by Kent Beck.

**IMPORTANT**: Before writing any code, read our [TDD Guide](./docs/guides/TDD-GUIDE.md).

---

## The Golden Rule

### Red → Green → Refactor

**Always** write tests before implementation:

1. ❌ **RED**: Write a failing test
2. ✅ **GREEN**: Write minimal code to pass
3. ♻️ **REFACTOR**: Clean up while tests protect you

---

## Required Reading

- [TDD Guide (Kent Beck Style)](./docs/guides/TDD-GUIDE.md) - **READ THIS FIRST**
- [Domain Specification](./docs/milestones/M2-DOMAIN-SPEC.md) - What to build
- [API Specification](./docs/milestones/M2-API-SPEC.md) - How API should work

---

## Code Review Checklist

Before submitting any code, verify:

- [ ] Tests were written **before** implementation
- [ ] All tests pass (`go test ./...`)
- [ ] Test coverage > 80% (`go test -cover ./...`)
- [ ] Code follows [TDD patterns](./docs/guides/TDD-GUIDE.md#kent-becks-test-patterns)
- [ ] Tests use Arrange-Act-Assert structure
- [ ] One test tests one thing
- [ ] Tests are readable (clear Given-When-Then)

---

## Workflow Example

### Starting a New Feature

```bash
# 1. Read the spec
cat docs/milestones/M2-DOMAIN-SPEC.md

# 2. Write the test first
# Example: internal/domain/battery_test.go
func TestNewBattery_ValidInput(t *testing.T) {
    battery, err := NewBattery(200.0, 100.0, ...)
    assert.NoError(t, err)
}

# 3. Run test (should FAIL)
go test ./internal/domain/...

# 4. Write minimal code to pass
# Example: internal/domain/battery.go
func NewBattery(...) (*Battery, error) {
    return &Battery{...}, nil
}

# 5. Run test (should PASS)
go test ./internal/domain/...

# 6. Refactor if needed

# 7. Commit
git add .
git commit -m "feat(domain): add NewBattery constructor"
```

---

## Testing Conventions

### File Structure

```
internal/domain/
├── battery.go        # Implementation
└── battery_test.go   # Tests (same package)
```

### Test Naming

```go
// Pattern: Test<Struct>_<Method>_<Scenario>_<ExpectedOutcome>
func TestBattery_Validate_NegativeCapacity_ReturnsError(t *testing.T)
func TestNewBattery_ValidInput_ReturnsNoBatteryNoError(t *testing.T)
```

### Test Structure (Always)

```go
func TestSomething(t *testing.T) {
    // Arrange (Given) - Setup
    input := ...
    expected := ...
    
    // Act (When) - Execute
    result, err := DoSomething(input)
    
    // Assert (Then) - Verify
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

---

## AI Pair Programming

When using Claude Code or similar AI assistants:

### ✅ DO

1. **Share this CONTRIBUTING.md** with the AI
2. **Reference TDD-GUIDE.md** explicitly
3. **Ask for tests first**: "Write the test for X following our TDD guide"
4. **Review AI-generated code**: Understand every line
5. **Request explanations**: "Why did you structure it this way?"

### ❌ DON'T

1. Accept code without tests
2. Let AI write implementation before tests
3. Skip the refactor step
4. Blindly copy-paste without understanding

---

## Example AI Prompt

```
I need to implement Battery.Validate() method.

Please follow our TDD approach (docs/guides/TDD-GUIDE.md):
1. First, write the test for validating positive capacity
2. Make it fail
3. Write minimal code to pass
4. Then we'll add more validation tests one at a time

Refer to docs/milestones/M2-DOMAIN-SPEC.md for validation rules.
```

---

## Go Conventions

- Use `gofmt` before committing
- Run `go vet ./...` to catch issues
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use table-driven tests for variations
- Keep functions small (< 20 lines ideal)

---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(domain): add Battery aggregate with validation
test(domain): add validation tests for Battery
refactor(domain): extract validation helpers
fix(api): handle validation errors correctly
docs(readme): update API examples
```

---

## Questions?

- Read the [TDD Guide](./docs/guides/TDD-GUIDE.md) again
- Check [Domain Spec](./docs/milestones/M2-DOMAIN-SPEC.md)
- Review [API Spec](./docs/milestones/M2-API-SPEC.md)
- Look at existing tests as examples

---

**Remember**: Tests are not overhead. They're the design tool that makes great code possible.
