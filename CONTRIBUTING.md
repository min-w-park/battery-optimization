# Contributing to Battery Optimization System

## Quick Start

**New to the project?**
1. Read [Skills](./.claude/skills/) for detailed coding standards
2. Check [M2 Overview](./docs/milestones/M2-OVERVIEW.md) for current milestone
3. Follow the [Checklist](./docs/milestones/M2-CHECKLIST.md)

---

## Development Philosophy

This project follows **Modern Software Engineering** practices:

- **TDD (Test-Driven Development)** - Tests before code
- **DDD (Domain-Driven Design)** - Domain at the center
- **Hexagonal Architecture** - Infrastructure serves domain
- **Event-Driven** - Services communicate via events

**Detailed guides:** See `.claude/skills/` directory

---

## The TDD Workflow

```
1. ❌ RED:  Write a failing test
2. ✅ GREEN: Write minimal code to pass
3. ♻️  REFACTOR: Clean up (tests protect you)
4. 🔁 REPEAT: Next behavior
```

Full guide: [TDD-SKILL](./.claude/skills/TDD-SKILL.md)

---

## Project Structure

```
services/asset-management/
├── cmd/server/           # Entry point
├── internal/
│   ├── domain/          # Pure business logic (no dependencies)
│   ├── ports/           # Interfaces
│   └── adapters/        # Implementations (HTTP, PostgreSQL)
```

Full guide: [DDD-PATTERNS-SKILL](./.claude/skills/DDD-PATTERNS-SKILL.md)

---

## Development Workflow

### Starting a Feature

```bash
# 1. Read the spec
cat docs/milestones/M2-DOMAIN-SPEC.md

# 2. Write test FIRST
# internal/domain/battery_test.go
func TestNewBattery_ValidInput(t *testing.T) {
    battery, err := NewBattery(200.0, 100.0, ...)
    assert.NoError(t, err)
}

# 3. Run test (should FAIL)
go test ./internal/domain/...

# 4. Write minimal implementation
# internal/domain/battery.go

# 5. Run test (should PASS)
go test ./internal/domain/...

# 6. Refactor if needed

# 7. Commit
git add .
git commit -m "feat(domain): add NewBattery constructor"
```

---

## Using Claude Code (AI Pair Programming)

Claude Code automatically loads guides from `.claude/skills/`

**Good prompts:**
```
"Implement Battery.Validate() following TDD-SKILL and BATTERY-DOMAIN-SKILL"

"Create POST /batteries endpoint following API-DESIGN-SKILL"

"Write repository tests using GO-CONVENTIONS-SKILL"
```

**Reference specific skills explicitly for best results.**

---

## Code Standards

**All detailed in Skills:**
- [TDD-SKILL](./.claude/skills/TDD-SKILL.md) - Red-Green-Refactor
- [GO-CONVENTIONS-SKILL](./.claude/skills/GO-CONVENTIONS-SKILL.md) - Go idioms
- [DDD-PATTERNS-SKILL](./.claude/skills/DDD-PATTERNS-SKILL.md) - Domain design
- [BATTERY-DOMAIN-SKILL](./.claude/skills/BATTERY-DOMAIN-SKILL.md) - Domain knowledge
- [API-DESIGN-SKILL](./.claude/skills/API-DESIGN-SKILL.md) - REST APIs

---

## Before Committing

```bash
# Format
go fmt ./...

# Vet
go vet ./...

# Tests
go test ./...

# Coverage (aim for >80%)
go test -cover ./...
```

---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(domain): add Battery aggregate
test(domain): add validation tests
refactor(api): extract DTO conversion
fix(repo): handle duplicate IDs
docs(readme): update API examples
```

---

## Pull Request Checklist

- [ ] Tests written before implementation
- [ ] All tests pass
- [ ] Coverage >80%
- [ ] Follows patterns in Skills
- [ ] Code formatted (`go fmt`)
- [ ] No lint errors (`go vet`)
- [ ] Conventional commit messages

---

## Getting Help

**For coding standards:** Check `.claude/skills/`  
**For domain rules:** See `BATTERY-DOMAIN-SKILL.md`  
**For current tasks:** See `docs/milestones/M2-CHECKLIST.md`  
**For architecture:** See `docs/ARCHITECTURE.md`

---

**Happy coding! 🚀**

