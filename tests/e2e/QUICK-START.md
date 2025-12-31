# E2E Tests - Quick Start Guide

## ⚡ 30-Second Setup

```bash
# 1. Start everything
cd /path/to/battery-optimization
docker-compose up -d

# 2. Wait 10 seconds for services to initialize
sleep 10

# 3. Run tests
cd tests/e2e
go test -v

# Expected: 13/13 tests pass in ~40 seconds
```

## 🎯 Running Specific Tests

### Quick Smoke Test (10 seconds)

```bash
# Run just the core charging workflow
go test -v -run TestChargingWorkflow
```

### Critical Tests Only (20 seconds)

```bash
# Run Category 1 (Happy Path) - 4 tests
go test -v -run "TestCharging|TestDischarging|TestMultiple|TestPriceChange"
```

### Full Suite (40-50 seconds)

```bash
# Run all 13 tests
go test -v -timeout 2m
```

## 🔍 Verification Checklist

Before running tests, verify:

### 1. Infrastructure Running

```bash
docker-compose ps

# Should show "healthy" for:
# - asset-db
# - market-db
# - telemetry-db
# - nats
```

### 2. Services Running

```bash
docker-compose ps

# Should show "running" for:
# - asset-management (port 8080)
# - market-data (port 8081)
# - telemetry (port 8082)
# - device-interface (port 8083)
# - bidding (no port)
```

### 3. NATS Health

```bash
curl http://localhost:8222/healthz
# Expected: (empty response with 200 OK)
```

### 4. Bidding Automation Mode

```bash
docker-compose logs bidding | grep AUTOMATION_MODE
# Expected: AUTOMATION_MODE=FULL_AUTO
```

## 🐛 Quick Troubleshooting

### Problem: "Failed to connect to NATS"

```bash
# Check NATS is running
docker-compose ps nats
docker-compose logs nats

# Restart NATS
docker-compose restart nats
```

### Problem: "Timeout waiting for event"

```bash
# Check bidding service is running and in FULL_AUTO
docker-compose logs bidding | tail -20

# Restart bidding service
docker-compose restart bidding
```

### Problem: "Failed to create battery"

```bash
# Check asset-management service
curl http://localhost:8080/health
docker-compose logs asset-management | tail -20

# Restart if needed
docker-compose restart asset-management
```

### Problem: "Multiple test failures"

```bash
# Full reset - stop everything
docker-compose down

# Remove old data
docker-compose down -v

# Fresh start
docker-compose up -d

# Wait for health
sleep 15

# Run tests
cd tests/e2e && go test -v
```

## 📊 Understanding Test Output

### ✅ Successful Test

```
=== RUN   TestChargingWorkflow
    category1_happy_path_test.go:35: ✓ Setup complete
    category1_happy_path_test.go:46: ✓ Battery registered: abc123
    category1_happy_path_test.go:52: ✓ Received battery.registered.v1 event
    category1_happy_path_test.go:57: ✓ Market price created: $30.00/MWh
    category1_happy_path_test.go:68: ✓ Published battery state: SoC=30.0%
    category1_happy_path_test.go:74: ✓ Received charging.opportunity.detected.v1
    category1_happy_path_test.go:80: ✓ Received charging.command.issued.v1
    category1_happy_path_test.go:82: ✅ Charging workflow complete!
--- PASS: TestChargingWorkflow (2.34s)
```

### ❌ Failed Test

```
=== RUN   TestChargingWorkflow
    category1_happy_path_test.go:74: Failed to receive charging.opportunity.detected.v1 event (is AUTOMATION_MODE=FULL_AUTO?)
--- FAIL: TestChargingWorkflow (3.12s)
```

**Action**: Check bidding service logs and automation mode

## 🎓 Test Categories

### Category 1: Happy Path (4 tests, ~10s)
Core workflows that must work:
- ✅ Charging workflow
- ✅ Discharging workflow
- ✅ Multiple batteries
- ✅ Price change triggers decision

### Category 2: Edge Cases (4 tests, ~10s)
Boundary conditions:
- ✅ Charging threshold ($49/MWh)
- ✅ Discharging threshold ($101/MWh)
- ✅ SoC boundaries (30%, 80%)
- ✅ No opportunity (mid-range)

### Category 3: Error Handling (3 tests, ~8s)
Resilience:
- ✅ Invalid battery state
- ⏭️ Service unavailable (manual)
- ✅ Event publish failure

### Category 6: Complete Workflows (2 tests, ~12s)
Full business cycles:
- ✅ Complete arbitrage (charge → discharge)
- ✅ Day in the life (24-hour simulation)

## 🚀 Advanced Usage

### Run with Race Detector

```bash
go test -v -race
# Detects concurrent access issues
```

### Run Specific Category

```bash
# Category 1 only
go test -v -run "^TestCharging|^TestDischarging|^TestMultiple|^TestPriceChange"

# Category 2 only
go test -v -run "^TestEdgeCase|^TestNoOpportunity"

# Category 6 only
go test -v -run "^TestComplete|^TestDay"
```

### Monitor Events During Tests

```bash
# Terminal 1: Run event subscriber
cd tools/event-subscriber
go run main.go ">"

# Terminal 2: Run tests
cd tests/e2e
go test -v -run TestChargingWorkflow
```

### Run Tests in Loop (Stability Testing)

```bash
# Run 10 times to check for flakiness
for i in {1..10}; do
  echo "Run $i"
  go test -v -run TestChargingWorkflow || break
done
```

## 📈 Expected Results

### All Tests Pass

```
ok      github.com/minwook/battery-optimization/tests/e2e    42.156s
```

### Summary

```
PASS
TestChargingWorkflow                    2.3s ✅
TestDischargingWorkflow                 2.4s ✅
TestMultipleBatteriesWorkflow           3.2s ✅
TestPriceChangeTriggersDecision         3.5s ✅
TestEdgeCaseChargingThreshold           2.1s ✅
TestEdgeCaseDischargingThreshold        2.3s ✅
TestEdgeCaseSoCBoundaries               3.8s ✅
TestNoOpportunityMidRange               2.2s ✅
TestInvalidBatteryState                 3.1s ✅
TestEventPublishFailure                 2.0s ✅
TestCompleteArbitrageWorkflow           5.4s ✅
TestDayInTheLife                        7.9s ✅
```

**Total**: 13 tests, ~42s, 100% pass rate

## 🎯 Common Workflows

### First Time Running Tests

```bash
# 1. Clone repo and setup
git clone <repo>
cd battery-optimization

# 2. Start infrastructure
docker-compose up -d

# 3. Wait for healthy state
sleep 15
docker-compose ps  # All should be "healthy"

# 4. Run tests
cd tests/e2e
go test -v
```

### Daily Development

```bash
# Start services (if not running)
docker-compose up -d

# Run quick smoke test
cd tests/e2e
go test -v -run TestChargingWorkflow

# If pass, run full suite
go test -v
```

### Debugging Failed Test

```bash
# 1. Run single test with verbose output
go test -v -run TestChargingWorkflow

# 2. Check service logs
docker-compose logs bidding | tail -50
docker-compose logs asset-management | tail -50

# 3. Check NATS
curl http://localhost:8222/varz | jq

# 4. Restart problematic service
docker-compose restart bidding

# 5. Re-run test
go test -v -run TestChargingWorkflow
```

## ✅ Success Indicators

You'll know tests are working when:

1. ✅ All services show "healthy" in `docker-compose ps`
2. ✅ NATS responds to health check
3. ✅ Tests complete in ~40-50 seconds
4. ✅ All 13 tests show PASS
5. ✅ No error messages in service logs (warnings OK)

## 📚 More Information

- **Detailed docs**: [README.md](README.md)
- **Test summary**: [TEST-SUMMARY.md](TEST-SUMMARY.md)
- **Implementation status**: [IMPLEMENTATION-STATUS.md](IMPLEMENTATION-STATUS.md)
- **System demo**: [../../docs/DEMO.md](../../docs/DEMO.md)

## 🆘 Getting Help

If tests still fail after troubleshooting:

1. Check service logs: `docker-compose logs <service>`
2. Verify environment: `docker-compose config`
3. Full reset: `docker-compose down -v && docker-compose up -d`
4. Review test output for specific error messages
5. Check [TEST-SUMMARY.md](TEST-SUMMARY.md) troubleshooting section
