# Structured Logging Guide

This document describes the structured logging implementation across all battery optimization services.

## Overview

All services use **structured logging** with [uber-go/zap](https://github.com/uber-go/zap) to provide:
- **JSON-formatted logs** for machine parsing
- **Contextual fields** for better debugging and monitoring
- **Proper log levels** (Info, Warn, Error, Fatal)
- **ISO8601 timestamps** for consistency
- **Caller information** (file:line) for traceability

## Log Format

### Production Logs (JSON)

\`\`\`json
{
  "level": "info",
  "timestamp": "2025-12-31T15:16:47.380+0900",
  "caller": "server/main.go:44",
  "msg": "starting service",
  "service": "asset-management",
  "port": "8080",
  "log_level": "info"
}
\`\`\`

### Development Logs (Console)

When \`LOG_LEVEL=debug\` or \`LOG_LEVEL=development\`, logs use a human-readable console format:

\`\`\`
2025-12-31T15:16:47.380+0900  INFO  server/main.go:44  starting service  {"service": "asset-management", "port": "8080"}
\`\`\`

## Configuration

### Environment Variables

Each service supports the \`LOG_LEVEL\` environment variable:

| Value | Mode | Output Format | Use Case |
|-------|------|---------------|----------|
| \`info\` (default) | Production | JSON | Production deployments, log aggregation |
| \`debug\` | Development | Console | Local development |
| \`development\` | Development | Console | Local development |

**Example**:
\`\`\`bash
# Production (JSON logs)
LOG_LEVEL=info ./bin/asset-management

# Development (console logs)
LOG_LEVEL=debug ./bin/asset-management
\`\`\`

### Service Name Field

All log messages include a \`service\` field identifying the source:

- \`asset-management\`
- \`market-data\`
- \`telemetry\`
- \`bidding\`
- \`device-interface\`

This enables filtering logs by service in log aggregation systems.

## Log Levels

### Info

**Purpose**: Normal operational messages

**When to use**:
- Service startup/shutdown
- Connection establishment
- Successful operations
- State transitions

**Examples**:
\`\`\`go
log.Info("starting service",
    zap.String("port", cfg.Port),
    zap.String("log_level", cfg.LogLevel),
)

log.Info("database connection established")

log.Info("NATS publisher connected", zap.String("nats_url", cfg.NatsURL))
\`\`\`

### Warn

**Purpose**: Unexpected conditions that don't prevent operation

**When to use**:
- Optional dependency failures (e.g., NATS unavailable)
- Degraded operation mode
- Recoverable errors

**Examples**:
\`\`\`go
log.Warn("failed to connect to NATS",
    zap.String("nats_url", cfg.NatsURL),
    zap.Error(err),
)
log.Info("service will continue without event publishing")
\`\`\`

### Error

**Purpose**: Errors that affect specific operations but don't stop the service

**When to use**:
- Request processing failures
- Event publishing failures
- Database query errors (non-fatal)

**Examples**:
\`\`\`go
log.Error("failed to process battery state",
    zap.String("battery_id", event.BatteryID),
    zap.Error(err),
)
\`\`\`

### Fatal

**Purpose**: Errors that prevent service from operating

**When to use**:
- Cannot connect to required dependencies
- Invalid configuration
- Unrecoverable initialization errors

**Examples**:
\`\`\`go
log.Fatal("failed to connect to database", zap.Error(err))
// Service exits with status 1

log.Fatal("failed to run migrations", zap.Error(err))
\`\`\`

## Structured Fields

### Common Field Types

\`\`\`go
// Strings
zap.String("battery_id", "battery-123")
zap.String("nats_url", "nats://localhost:4222")

// Numbers
zap.Int("port", 8080)
zap.Float64("capacity_mwh", 200.0)
zap.Float64("soc_percent", 75.5)

// Errors
zap.Error(err)  // Automatically includes error message

// Durations
zap.Duration("timeout", 2*time.Second)

// Booleans
zap.Bool("is_ready", true)
\`\`\`

### Best Practices

1. **Use descriptive field names**: \`battery_id\` not \`id\`, \`nats_url\` not \`url\`
2. **Include units in field names**: \`capacity_mwh\`, \`max_power_mw\`, \`soc_percent\`
3. **Always include error context**:
   \`\`\`go
   log.Warn("failed to publish event",
       zap.String("event_type", "battery.registered.v1"),
       zap.String("battery_id", batteryID),
       zap.Error(err),
   )
   \`\`\`
4. **Use consistent field names across services**:
   - \`battery_id\` (not \`batteryID\` or \`id\`)
   - \`nats_url\` (not \`natsURL\` or \`url\`)
   - \`log_level\` (not \`logLevel\` or \`level\`)

## Related Documentation

- [CLAUDE.md](../../CLAUDE.md) - Development commands and workflow
- [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture overview
