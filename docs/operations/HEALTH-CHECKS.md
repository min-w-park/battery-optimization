# Health Check Endpoints

This document describes the health check endpoints available for monitoring service health and readiness.

## Overview

All HTTP-based services expose two health check endpoints:
- `/health/live` - Liveness probe
- `/health/ready` - Readiness probe

These endpoints follow Kubernetes health check conventions and can be used for:
- Kubernetes liveness and readiness probes
- Load balancer health checks
- Monitoring and alerting systems
- Manual service verification

## Services with Health Endpoints

| Service | Port | Liveness | Readiness |
|---------|------|----------|-----------|
| Asset Management | 8080 | `/health/live` | `/health/ready` |
| Market Data | 8080 | `/health/live` | `/health/ready` |
| Telemetry | 8082 | `/health/live` | `/health/ready` |

**Note**: Device Interface and Bidding services are event-driven only and do not expose HTTP endpoints.

## Endpoint Details

### Liveness Probe: `/health/live`

**Purpose**: Indicates if the service process is running and can handle requests.

**When to use**:
- Kubernetes liveness probes
- Detecting if a service needs to be restarted
- Basic "is the service up?" checks

**Response**:
```json
{
  "status": "ok",
  "service": "asset-management"
}
```

**Status Codes**:
- `200 OK` - Service is alive and running

**Failure Conditions**:
- Service process has crashed or is unresponsive
- Should **never** fail unless the service is completely down

### Readiness Probe: `/health/ready`

**Purpose**: Indicates if the service is ready to accept traffic and handle requests.

**When to use**:
- Kubernetes readiness probes
- Load balancer backend pool checks
- Determining if a service is ready to receive traffic
- Checking if dependencies are available

**Response (Healthy)**:
```json
{
  "status": "ready",
  "service": "asset-management",
  "checks": {
    "database": "healthy",
    "nats": "healthy"
  }
}
```

**Response (Unhealthy)**:
```json
{
  "status": "unavailable",
  "service": "asset-management",
  "checks": {
    "database": "unhealthy: connection timeout",
    "nats": "not checked"
  }
}
```

**Status Codes**:
- `200 OK` - Service is ready to handle traffic
- `503 Service Unavailable` - Service dependencies are unhealthy

**Checks Performed**:
1. **Database Connection**: 2-second ping timeout
   - `healthy` - Database is reachable
   - `unhealthy: <error>` - Database ping failed

2. **NATS Connection** (Optional):
   - `healthy` - NATS is connected
   - `unhealthy (optional)` - NATS is disconnected (service continues to work)
   - `not configured` - NATS URL not provided

**Failure Conditions**:
- Database connection is unavailable
- Database query timeout exceeds 2 seconds
- Critical dependencies are not reachable

## Usage Examples

### Manual Testing

```bash
# Test liveness
curl http://localhost:8080/health/live

# Test readiness
curl http://localhost:8080/health/ready

# Test with pretty-print
curl -s http://localhost:8080/health/ready | jq .

# Check status code only
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health/ready
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: asset-management
spec:
  containers:
  - name: asset-management
    image: asset-management:latest
    ports:
    - containerPort: 8080

    # Liveness probe: Restart if service is dead
    livenessProbe:
      httpGet:
        path: /health/live
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 30
      timeoutSeconds: 5
      failureThreshold: 3

    # Readiness probe: Remove from load balancer if not ready
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
      timeoutSeconds: 2
      failureThreshold: 2
      successThreshold: 1
```

### Docker Compose Health Check

```yaml
services:
  asset-management:
    image: asset-management:latest
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health/ready"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 40s
```

### Monitoring Script

```bash
#!/bin/bash
# check-all-services.sh

SERVICES=(
  "asset-management:8080"
  "market-data:8080"
  "telemetry:8082"
)

for service in "${SERVICES[@]}"; do
  name="${service%:*}"
  port="${service#*:}"

  status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${port}/health/ready")

  if [ "$status" -eq 200 ]; then
    echo "✅ $name is healthy"
  else
    echo "❌ $name is unhealthy (HTTP $status)"
  fi
done
```

## Best Practices

### Liveness vs Readiness

**Use liveness** to detect:
- Process crashes
- Deadlocks
- Infinite loops
- Unrecoverable errors

**Use readiness** to detect:
- Database connection issues
- Dependency unavailability
- Startup/shutdown periods
- Resource exhaustion

### Timeouts

- **Liveness**: Longer timeout (5s), less frequent (30s interval)
  - Prevents unnecessary restarts due to temporary slowness

- **Readiness**: Shorter timeout (2s), more frequent (10s interval)
  - Quickly removes unhealthy instances from load balancer

### Failure Thresholds

- **Liveness**: Higher threshold (3 failures) before restart
  - Avoids restart storms during temporary issues

- **Readiness**: Lower threshold (2 failures) before removal
  - Quickly stops routing traffic to degraded instances

## Troubleshooting

### Service Shows Unhealthy But Works

**Symptoms**: `/health/ready` returns 503, but service handles requests fine.

**Possible Causes**:
1. Database connection pool exhausted
2. Slow database queries affecting health check timeout
3. Network latency to dependencies

**Resolution**:
```bash
# Check database connectivity
docker exec asset-db psql -U asset_user -d asset_management -c "SELECT 1;"

# Check NATS connectivity
curl http://localhost:8222/healthz
```

### Health Check Times Out

**Symptoms**: Health check request hangs or times out.

**Possible Causes**:
1. Service is under heavy load
2. Goroutine/thread pool exhaustion
3. Middleware blocking requests

**Resolution**:
```bash
# Check if service is running
docker ps | grep asset-management

# Check service logs
docker logs asset-management

# Check resource usage
docker stats asset-management
```

### Frequent Restarts (Liveness Failing)

**Symptoms**: Kubernetes restarts pods frequently.

**Possible Causes**:
1. Liveness timeout too aggressive
2. Service genuinely crashing
3. Resource limits too low

**Resolution**:
1. Increase `initialDelaySeconds` and `timeoutSeconds`
2. Check service logs for panic/crash
3. Increase CPU/memory limits

## Monitoring Integration

### Prometheus Metrics (Future)

Health check metrics will be available via Prometheus:

```prometheus
# Service uptime
up{service="asset-management"} 1

# Health check latency
health_check_duration_seconds{endpoint="/health/ready"} 0.002

# Dependency health
dependency_health{service="asset-management",dependency="database"} 1
```

### Alerting Rules (Future)

```yaml
groups:
- name: health_checks
  rules:
  - alert: ServiceDown
    expr: up{job="battery-services"} == 0
    for: 1m
    annotations:
      summary: "Service {{ $labels.service }} is down"

  - alert: DatabaseUnhealthy
    expr: dependency_health{dependency="database"} == 0
    for: 2m
    annotations:
      summary: "Database health check failing for {{ $labels.service }}"
```

## Related Documentation

- [CI/CD Pipeline](CI-CD.md) - Automated testing and deployment
- [CLAUDE.md](../../CLAUDE.md) - Development commands and workflow
- [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture overview
