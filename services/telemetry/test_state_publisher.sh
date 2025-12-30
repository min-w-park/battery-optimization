#!/bin/bash

# Test script for StatePublisher - inserts test battery state and verifies 1 Hz publishing

echo "Inserting test battery state into database..."

# Insert a test battery state
docker exec telemetry-db psql -U telemetry_user -d telemetry <<EOF
-- Insert test battery state
INSERT INTO battery_states (
    id, battery_id, soc, power, temperature, voltage, current,
    operation_state, custom_attributes, timestamp, created_at
) VALUES (
    'state-test-001',
    'battery-123',
    75.5,
    25.0,
    28.5,
    825.0,
    30303.0,
    'DISCHARGING',
    '{"vendor": "Tesla", "model": "Megapack"}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (id) DO UPDATE SET
    soc = EXCLUDED.soc,
    power = EXCLUDED.power,
    temperature = EXCLUDED.temperature,
    timestamp = EXCLUDED.timestamp;

-- Verify the insert
SELECT battery_id, soc, power, operation_state, timestamp
FROM battery_states
WHERE battery_id = 'battery-123'
ORDER BY timestamp DESC
LIMIT 1;
EOF

echo ""
echo "Test battery state inserted successfully!"
echo ""
echo "Now you can:"
echo "1. Start the event subscriber in another terminal:"
echo "   cd tools/event-subscriber && go run main.go 'battery.>'"
echo ""
echo "2. Start the telemetry service:"
echo "   cd services/telemetry && PORT=8082 ./bin/telemetry"
echo ""
echo "3. You should see 'battery.state.changed.v1' events published every 1 second"
