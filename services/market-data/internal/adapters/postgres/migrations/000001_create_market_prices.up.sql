CREATE TABLE IF NOT EXISTS market_prices (
    id VARCHAR(36) PRIMARY KEY,
    region VARCHAR(3) NOT NULL CHECK (region IN ('NSW', 'VIC', 'QLD', 'SA', 'TAS')),
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    demand NUMERIC(10,2) NOT NULL CHECK (demand > 0),
    interval_type VARCHAR(20) NOT NULL CHECK (interval_type IN ('5MIN_PREDISPATCH', '30MIN_PREDISPATCH')),
    interval_start TIMESTAMP NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Unique constraint: no duplicate (region, interval_type, interval_start)
CREATE UNIQUE INDEX idx_market_prices_unique_interval
    ON market_prices(region, interval_type, interval_start);

-- Time-range query optimization
CREATE INDEX idx_market_prices_time_range
    ON market_prices(region, interval_type, interval_start DESC);

-- Region filtering
CREATE INDEX idx_market_prices_region
    ON market_prices(region, created_at DESC);
