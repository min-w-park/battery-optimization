#!/bin/bash
set -e

echo "=== E2E Test Runner with Proper Isolation ==="
echo ""

# Function to clean databases and restart bidding service
cleanup_and_restart() {
    echo "🧹 Cleaning databases and restarting bidding service..."
    docker exec asset-db psql -U asset_user -d asset_management -c "TRUNCATE batteries CASCADE;" > /dev/null
    docker exec market-db psql -U market_user -d market_data -c "TRUNCATE market_prices CASCADE;" > /dev/null
    docker-compose restart bidding > /dev/null 2>&1
    sleep 3
    echo "✓ Cleanup complete"
    echo ""
}

# Initial cleanup
cleanup_and_restart

# Category 1: Happy Path Tests
echo "=== Category 1: Happy Path Tests ==="
cd /Users/minwook/work/energy/battery-optimization/tests/e2e

echo "Running TestChargingWorkflow..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestChargingWorkflow$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestDischargingWorkflow..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestDischargingWorkflow$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestMultipleBatteriesWorkflow..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestMultipleBatteriesWorkflow$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestPriceChangeTriggersDecision..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestPriceChangeTriggersDecision$"
if [ $? -ne 0 ]; then exit 1; fi

echo ""

# Category 2: Edge Cases
echo "=== Category 2: Edge Cases ==="

echo "Running TestEdgeCaseChargingThreshold..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestEdgeCaseChargingThreshold$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestEdgeCaseDischargingThreshold..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestEdgeCaseDischargingThreshold$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestEdgeCaseSoCBoundaries/HighSoC_NoCharging..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestEdgeCaseSoCBoundaries/HighSoC_NoCharging$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestEdgeCaseSoCBoundaries/LowSoC_NoDischarging..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestEdgeCaseSoCBoundaries/LowSoC_NoDischarging$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestNoOpportunityMidRange..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestNoOpportunityMidRange$"
if [ $? -ne 0 ]; then exit 1; fi

echo ""

# Category 3: Error Handling
echo "=== Category 3: Error Handling ==="

echo "Running TestInvalidBatteryState/AlreadyCharging_NoNewCommand..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestInvalidBatteryState/AlreadyCharging_NoNewCommand$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestInvalidBatteryState/AlreadyDischarging_NoNewCommand..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestInvalidBatteryState/AlreadyDischarging_NoNewCommand$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestEventPublishFailure..."
cleanup_and_restart
go test -v -timeout 60s -count=1 -run "^TestEventPublishFailure$"
if [ $? -ne 0 ]; then exit 1; fi

echo ""

# Category 6: Complete Workflows
echo "=== Category 6: Complete Workflows ==="

echo "Running TestCompleteArbitrageWorkflow..."
cleanup_and_restart
go test -v -timeout 120s -count=1 -run "^TestCompleteArbitrageWorkflow$"
if [ $? -ne 0 ]; then exit 1; fi

echo "Running TestDayInTheLife..."
cleanup_and_restart
go test -v -timeout 120s -count=1 -run "^TestDayInTheLife$"
if [ $? -ne 0 ]; then exit 1; fi

echo ""

# Final cleanup
echo "=== Final Cleanup ==="
cleanup_and_restart

echo "✅ All E2E tests passed!"
