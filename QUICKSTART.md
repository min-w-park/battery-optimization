# Quick Start Guide

## Prerequisites

- Docker & Docker Compose installed
- Go 1.21+ installed
- Make (optional, for convenience commands)

## Step 1: Verify Setup

```bash
# Check Docker is running
docker --version
docker-compose --version

# Check Go is installed
go version
```

## Step 2: Start Infrastructure

```bash
# Start all databases and NATS
docker-compose up -d

# Wait for services to be healthy
sleep 10

# Check status
docker-compose ps

# Verify databases are ready
make db-status

# Check NATS
curl http://localhost:8222/healthz
```

Expected output: All services should show "healthy" status.

## Step 3: Initialize Service Directories

```bash
# Create service directory structure
make init-services

# Verify structure
tree -L 3 services/
```

## Step 4: Test Database Connections

```bash
# Test Asset DB
docker exec -it asset-db psql -U asset_user -d asset_management -c "SELECT version();"

# Test Market DB
docker exec -it market-db psql -U market_user -d market_data -c "SELECT version();"

# Test Telemetry DB
docker exec -it telemetry-db psql -U telemetry_user -d telemetry -c "SELECT version();"
```

## Step 5: View Logs

```bash
# Follow all logs
make logs

# Or specific service
docker-compose logs -f asset-db
```

## Stopping Services

```bash
# Stop services (keep data)
make down

# Stop and remove all data
make clean
```

## Troubleshooting

### Port Already in Use

If you see "port already in use" errors:

```bash
# Check what's using the port
lsof -i :5432
# or
netstat -an | grep 5432

# Stop the service or change docker-compose.yml ports
```

### Database Not Ready

If database health checks fail:

```bash
# View database logs
docker-compose logs asset-db

# Restart specific service
docker-compose restart asset-db
```

### NATS Connection Issues

```bash
# Check NATS logs
docker-compose logs nats

# Test NATS monitoring endpoint
curl http://localhost:8222/varz
```

## Next Steps

Once M0 is complete and all services are healthy:

1. ✅ Update PLANNING.md: Mark M0 tasks as complete
2. 🎨 Begin M1: Event Storming session
3. 📝 Create EVENTS.md with initial event catalog

## M0 Completion Checklist

- [ ] Docker Compose starts all services
- [ ] All database health checks pass
- [ ] NATS is accessible on port 4222
- [ ] Service directories created
- [ ] Can connect to each database
- [ ] Project structure documented
- [ ] README, PLANNING.md, EVENTS.md exist

**When all items are checked, M0 is complete!**
