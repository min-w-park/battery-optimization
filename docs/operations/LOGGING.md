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

```json
{
  "level": "info",
  "timestamp": "2025-12-31T15:16:47.380+0900",
  "caller": "server/main.go:44",
  "msg": "starting service",
  "service": "asset-management",
  "port": "8080",
  "log_level": "info"
}
```

### Development Logs (Console)

When `LOG_LEVEL=debug` or `LOG_LEVEL=development`, logs use a human-readable console format:

```
2025-12-31T15:16:47.380+0900  INFO  server/main.go:44  starting service  {"service": "asset-management", "port": "8080"}
```

## Configuration

### Environment Variables

Each service supports the `LOG_LEVEL` environment variable:

| Value | Mode | Output Format | Use Case |
|-------|------|---------------|----------|
| `info` (default) | Production | JSON | Production deployments, log aggregation |
| `debug` | Development | Console | Local development |
| `development` | Development | Console | Local development |

**Example**:
```bash
# Production (JSON logs)
LOG_LEVEL=info ./bin/asset-management

# Development (console logs)
LOG_LEVEL=debug ./bin/asset-management
```

### Service Name Field

All log messages include a `service` field identifying the source:

- `asset-management`
- `market-data`
- `telemetry`
- `bidding`
- `device-interface`

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
```go
log.Info("starting service",
    zap.String("port", cfg.Port),
    zap.String("log_level", cfg.LogLevel),
)

log.Info("database connection established")

log.Info("NATS publisher connected", zap.String("nats_url", cfg.NatsURL))
```

### Warn

**Purpose**: Unexpected conditions that don't prevent operation

**When to use**:
- Optional dependency failures (e.g., NATS unavailable)
- Degraded operation mode
- Recoverable errors

**Examples**:
```go
log.Warn("failed to connect to NATS",
    zap.String("nats_url", cfg.NatsURL),
    zap.Error(err),
)
log.Info("service will continue without event publishing")
```

### Error

**Purpose**: Errors that affect specific operations but don't stop the service

**When to use**:
- Request processing failures
- Event publishing failures
- Database query errors (non-fatal)

**Examples**:
```go
log.Error("failed to process battery state",
    zap.String("battery_id", event.BatteryID),
    zap.Error(err),
)
```

### Fatal

**Purpose**: Errors that prevent service from operating

**When to use**:
- Cannot connect to required dependencies
- Invalid configuration
- Unrecoverable initialization errors

**Examples**:
```go
log.Fatal("failed to connect to database", zap.Error(err))
// Service exits with status 1

log.Fatal("failed to run migrations", zap.Error(err))
```

## Structured Fields

### Common Field Types

```go
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
```

### Best Practices

1. **Use descriptive field names**: `battery_id` not `id`, `nats_url` not `url`
2. **Include units in field names**: `capacity_mwh`, `max_power_mw`, `soc_percent`
3. **Always include error context**:
   ```go
   log.Warn("failed to publish event",
       zap.String("event_type", "battery.registered.v1"),
       zap.String("battery_id", batteryID),
       zap.Error(err),
   )
   ```
4. **Use consistent field names across services**:
   - `battery_id` (not `batteryID` or `id`)
   - `nats_url` (not `natsURL` or `url`)
   - `log_level` (not `logLevel` or `level`)

## Service-Specific Logging

### Asset Management Service

**Startup logs**:
```json
{"level":"info","msg":"starting service","service":"asset-management","port":"8080","log_level":"info"}
{"level":"info","msg":"database connection established","service":"asset-management"}
{"level":"info","msg":"migrations applied successfully","service":"asset-management"}
{"level":"info","msg":"repository initialized","service":"asset-management"}
{"level":"info","msg":"NATS publisher connected","service":"asset-management","nats_url":"nats://localhost:4222"}
{"level":"info","msg":"HTTP handler initialized","service":"asset-management"}
{"level":"info","msg":"health handler initialized","service":"asset-management"}
{"level":"info","msg":"routes configured","service":"asset-management"}
{"level":"info","msg":"starting HTTP server","service":"asset-management","port":"8080"}
```

### Market Data Service

Similar to Asset Management, with service name `market-data` and default port `8081`.

### Telemetry Service

**Additional logs**:
```json
{"level":"info","msg":"state publisher started","service":"telemetry","battery_id":"battery-123","frequency":"1 Hz"}
```

### Bidding Service

**Event-driven logs**:
```json
{"level":"info","msg":"starting service","service":"bidding","nats_url":"nats://localhost:4222","automation_mode":"MANUAL","log_level":"info"}
{"level":"info","msg":"connected to NATS for publishing","service":"bidding"}
{"level":"info","msg":"connected to NATS for subscribing","service":"bidding"}
{"level":"info","msg":"initialized in-memory caches","service":"bidding"}
{"level":"info","msg":"created bidding engine","service":"bidding","automation_mode":"MANUAL"}
{"level":"info","msg":"subscribed to battery.state.changed.v1","service":"bidding"}
{"level":"info","msg":"subscribed to market.price.updated.v1","service":"bidding"}
{"level":"info","msg":"subscribed to battery.registered.v1","service":"bidding"}
{"level":"info","msg":"all event subscriptions configured successfully","service":"bidding"}
{"level":"info","msg":"bidding service is running in event-driven mode","service":"bidding"}
```

### Device Interface Service

**Hardware adapter logs**:
```json
{"level":"info","msg":"starting service","service":"device-interface","battery_id":"battery-001","adapter_type":"TeslaLike","capacity_mwh":200,"max_power_mw":100}
{"level":"info","msg":"created TeslaLike adapter","service":"device-interface","battery_id":"battery-001"}
{"level":"info","msg":"NATS publisher connected","service":"device-interface","nats_url":"nats://localhost:4222"}
{"level":"info","msg":"published BatteryConnectionEstablished event","service":"device-interface","battery_id":"battery-001"}
{"level":"info","msg":"command handler initialized","service":"device-interface"}
{"level":"info","msg":"NATS subscriber connected","service":"device-interface","nats_url":"nats://localhost:4222"}
{"level":"info","msg":"subscribed to charging.command.issued.v1","service":"device-interface"}
{"level":"info","msg":"subscribed to discharging.command.issued.v1","service":"device-interface"}
{"level":"info","msg":"subscribed to conflict.resolved.v1","service":"device-interface"}
{"level":"info","msg":"event subscriptions active - ready to process commands","service":"device-interface"}
```

## Log Aggregation

### ELK Stack (Future)

Structured JSON logs are ready for ingestion by Elasticsearch/Logstash/Kibana:

**Logstash configuration**:
```ruby
input {
  file {
    path => "/var/log/battery-services/*.log"
    codec => json
  }
}

filter {
  # Logs are already JSON, no parsing needed
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "battery-services-%{+YYYY.MM.dd}"
  }
}
```

**Kibana queries**:
```
service:"asset-management" AND level:"error"
service:"bidding" AND msg:"opportunity detected"
level:"fatal"
```

### CloudWatch Logs (Future)

```bash
# Stream logs to CloudWatch
aws logs create-log-group --log-group-name /battery-optimization/services

# Each service streams to its own log stream
aws logs create-log-stream \
  --log-group-name /battery-optimization/services \
  --log-stream-name asset-management
```

### Grafana Loki (Future)

```yaml
# promtail config for Loki
scrape_configs:
  - job_name: battery-services
    static_configs:
      - targets:
          - localhost
        labels:
          job: battery-services
          __path__: /var/log/battery-services/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            service: service
            msg: msg
```

**LogQL queries**:
```logql
{service="asset-management"} | json | level="error"
{service="bidding"} |= "opportunity"
{service=~"asset-management|market-data"} | json | level="warn"
```

## Troubleshooting

### Issue: Logs not appearing

**Check**:
1. Service is writing to stdout/stderr
2. Log level is appropriate (set to `debug` for development)
3. No log file redirection silencing output

**Solution**:
```bash
# Run service in foreground to see logs
LOG_LEVEL=debug ./bin/asset-management

# Check if logs are being written
docker-compose logs -f asset-management
```

### Issue: Too many logs

**Solution**: Adjust log level to `info` in production
```bash
LOG_LEVEL=info docker-compose up
```

### Issue: Missing contextual information

**Problem**: Logs don't include enough fields for debugging

**Solution**: Add more structured fields:
```go
// Before
log.Info("processing request")

// After
log.Info("processing request",
    zap.String("battery_id", batteryID),
    zap.String("operation", "update_soc"),
    zap.Float64("new_soc", newSoC),
)
```

### Issue: JSON logs hard to read locally

**Solution**: Use development mode
```bash
LOG_LEVEL=development ./bin/asset-management
```

Or pipe JSON logs through jq:
```bash
./bin/asset-management 2>&1 | jq -R 'fromjson?'
```

## Migration from Standard Library log

### Old Pattern (Removed)
```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
log.Printf("Starting service on port %s", port)
log.Fatal("Failed to connect:", err)
```

### New Pattern (Current)
```go
import (
    "go.uber.org/zap"
    "github.com/minwook/battery-optimization/pkg/logger"
)

log, err := logger.NewFromEnv("service-name", cfg.LogLevel)
if err != nil {
    panic(fmt.Sprintf("Failed to initialize logger: %v", err))
}
defer log.Sync()

log.Info("starting service", zap.String("port", port))
log.Fatal("failed to connect", zap.Error(err))
```

## Related Documentation

- [Health Check Endpoints](HEALTH-CHECKS.md) - Service health monitoring
- [CI/CD Pipeline](CI-CD.md) - Automated testing and deployment
- [CLAUDE.md](../../CLAUDE.md) - Development commands and workflow
- [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture overview
