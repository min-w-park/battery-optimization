-- Drop indexes first
DROP INDEX IF EXISTS idx_battery_states_custom_attributes;
DROP INDEX IF EXISTS idx_battery_states_operation_state;
DROP INDEX IF EXISTS idx_battery_states_timestamp;
DROP INDEX IF EXISTS idx_battery_states_battery_id_timestamp;

-- Drop table
DROP TABLE IF EXISTS battery_states;
