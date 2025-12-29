CREATE TABLE IF NOT EXISTS batteries (
    id VARCHAR(36) PRIMARY KEY,
    capacity NUMERIC(10,2) NOT NULL CHECK (capacity > 0),
    max_power NUMERIC(10,2) NOT NULL CHECK (max_power > 0 AND max_power <= capacity),
    ramp_rate NUMERIC(10,2) NOT NULL CHECK (ramp_rate > 0 AND ramp_rate <= max_power),
    efficiency NUMERIC(5,4) NOT NULL CHECK (efficiency >= 0 AND efficiency <= 1),
    location VARCHAR(3) NOT NULL CHECK (location IN ('NSW', 'VIC', 'QLD', 'SA', 'TAS')),
    manufacturer VARCHAR(100) NOT NULL,
    constraints JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'REGISTERED',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batteries_status ON batteries(status);
CREATE INDEX idx_batteries_location ON batteries(location);
CREATE INDEX idx_batteries_created_at ON batteries(created_at DESC);
