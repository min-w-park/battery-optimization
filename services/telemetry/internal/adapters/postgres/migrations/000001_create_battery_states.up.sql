-- Create battery_states table for time-series telemetry data
CREATE TABLE IF NOT EXISTS battery_states (
    id VARCHAR(36) PRIMARY KEY,
    battery_id VARCHAR(36) NOT NULL,
    soc NUMERIC(5,2) NOT NULL CHECK (soc >= 0 AND soc <= 100),
    power NUMERIC(10,2) NOT NULL,
    temperature NUMERIC(5,2) NOT NULL CHECK (temperature >= -20 AND temperature <= 60),
    voltage NUMERIC(10,2) NOT NULL,
    current NUMERIC(10,2) NOT NULL,
    operation_state VARCHAR(20) NOT NULL CHECK (operation_state IN ('IDLE', 'CHARGING', 'DISCHARGING', 'FCAS')),
    custom_attributes JSONB,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Optimized indexes for time-series queries
-- Primary query pattern: Get latest state for battery
CREATE INDEX IF NOT EXISTS idx_battery_states_battery_id_timestamp ON battery_states(battery_id, timestamp DESC);

-- Historical queries by timestamp
CREATE INDEX IF NOT EXISTS idx_battery_states_timestamp ON battery_states(timestamp DESC);

-- Filter by operation state
CREATE INDEX IF NOT EXISTS idx_battery_states_operation_state ON battery_states(operation_state);

-- JSONB GIN index for custom attributes queries (optional, for future use)
CREATE INDEX IF NOT EXISTS idx_battery_states_custom_attributes ON battery_states USING GIN (custom_attributes);
