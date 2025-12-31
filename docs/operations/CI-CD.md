# CI/CD Pipeline Documentation

## Overview

The battery optimization system uses GitHub Actions for continuous integration and continuous delivery. Every commit triggers automated testing, and successful builds produce versioned Docker images ready for deployment.

## Workflows

### 1. Test Workflow (`test.yml`)

**Triggers**:
- Push to `develop` or `main` branches
- Pull requests targeting `develop` or `main`

**Jobs**:

#### Test Job
Runs all unit and integration tests with full infrastructure:

**Services**:
- 3x PostgreSQL 18 databases (ports 5432, 5433, 5434)
- NATS event bus (ports 4222, 8222)

**Steps**:
1. Checkout code
2. Set up Go 1.23
3. Download dependencies
4. Run `go vet` (static analysis)
5. Run tests with race detection: `go test -v -race -coverprofile=coverage.out ./...`
6. Upload coverage to Codecov (optional, requires `CODECOV_TOKEN` secret)
7. Generate coverage report

**Environment Variables**:
```bash
ASSET_DB_URL=postgres://asset_user:asset_pass@localhost:5432/asset_management?sslmode=disable
MARKET_DB_URL=postgres://market_user:market_pass@localhost:5433/market_data?sslmode=disable
TELEMETRY_DB_URL=postgres://telemetry_user:telemetry_pass@localhost:5434/telemetry?sslmode=disable
NATS_URL=nats://localhost:4222
```

#### Lint Job
Runs `golangci-lint` for code quality checks:

**Checks Include**:
- Dead code detection
- Unused variables/imports
- Code complexity metrics
- Go best practices enforcement
- Security vulnerability scanning

**Configuration**: Uses `golangci-lint` defaults with 5-minute timeout

---

### 2. Build Workflow (`build.yml`)

**Triggers**:
- Push to `develop` or `main` branches
- Push of version tags (`v*`)
- Pull requests (build only, no push)

**Jobs**:

#### Build Services Job
Builds Docker images for all 5 microservices in parallel:

**Services Built**:
1. `asset-management`
2. `market-data`
3. `telemetry`
4. `device-interface`
5. `bidding`

**Image Tagging Strategy**:
- Branch pushes: `<service>:develop`, `<service>:main`
- Pull requests: `<service>:pr-123`
- Version tags: `<service>:v1.2.3`, `<service>:1.2`
- Commit SHA: `<service>:develop-abc1234`

**Registry**: GitHub Container Registry (`ghcr.io`)

**Build Features**:
- Multi-stage Docker builds (optimized image sizes)
- Layer caching via GitHub Actions cache
- Build args: `SERVICE_NAME`, `VERSION`, `COMMIT_SHA`

**Permissions Required**:
- `contents: read` - Read repository code
- `packages: write` - Push to ghcr.io

#### Verify Build Job
Runs after all images are built successfully (non-PR only):
- Confirms all 5 services built
- Displays image registry URLs
- Shows version tags and commit SHA

---

## Running Locally

### Run Tests (Match CI Environment)

```bash
# Full test suite with race detection
go test -v -race -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./services/asset-management/...

# Run specific test
go test -v -run TestChargingWorkflow ./tests/e2e/
```

### Run Static Analysis

```bash
# Go vet (included in Go toolchain)
go vet ./...

# Install golangci-lint (one-time)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run --timeout=5m

# Auto-fix issues where possible
golangci-lint run --fix
```

### Build Docker Images Locally

```bash
# Build single service
docker build -t battery-optimization/asset-management:local \
  -f services/asset-management/Dockerfile .

# Build all services (using docker-compose)
docker-compose build

# Build with build args
docker build \
  --build-arg SERVICE_NAME=asset-management \
  --build-arg VERSION=local \
  --build-arg COMMIT_SHA=$(git rev-parse HEAD) \
  -f services/asset-management/Dockerfile \
  -t battery-optimization/asset-management:local .
```

---

## Secrets Configuration

### Required Secrets

**Optional (for coverage reporting)**:
- `CODECOV_TOKEN` - Codecov.io API token for coverage uploads

**Automatic (GitHub-provided)**:
- `GITHUB_TOKEN` - Automatically available, used for ghcr.io authentication

### Setting Up Secrets

1. Navigate to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add `CODECOV_TOKEN` (if using Codecov):
   - Sign up at https://codecov.io
   - Link your repository
   - Copy the upload token
   - Paste into GitHub secret

---

## Workflow Status Badges

Add to README.md:

```markdown
![Test Status](https://github.com/YOUR_USERNAME/battery-optimization/workflows/Test/badge.svg)
![Build Status](https://github.com/YOUR_USERNAME/battery-optimization/workflows/Build/badge.svg)
```

---

## Troubleshooting

### Tests Fail in CI but Pass Locally

**Problem**: Race conditions or timing issues in CI environment

**Solutions**:
1. Run tests locally with `-race` flag: `go test -race ./...`
2. Check for hardcoded timeouts that might be too short
3. Review E2E test timeouts (currently 3 seconds for events)
4. Add retries for flaky network operations

### Docker Build Fails

**Problem**: Context size too large or missing files

**Solutions**:
1. Check `.dockerignore` file includes build artifacts
2. Verify Dockerfile path: `services/<service>/Dockerfile`
3. Build locally to reproduce: `docker build -f services/asset-management/Dockerfile .`
4. Check GitHub Actions logs for specific error

### golangci-lint Timeout

**Problem**: Linter takes > 5 minutes

**Solutions**:
1. Increase timeout in workflow: `args: --timeout=10m`
2. Disable slow linters in `.golangci.yml`
3. Run linter on changed files only (future enhancement)

### Image Push Permission Denied

**Problem**: Cannot push to ghcr.io

**Solutions**:
1. Verify workflow has `packages: write` permission
2. Enable GitHub Packages in repository settings
3. Check `GITHUB_TOKEN` has correct scopes
4. For personal repos: Enable "Improved container support" in settings

---

## Future Enhancements

**Planned Improvements**:
1. E2E test workflow (run against live docker-compose stack)
2. Performance benchmarking (`go test -bench`)
3. Dependency vulnerability scanning (Dependabot, Snyk)
4. Automated deployment to staging environment
5. Smoke tests after deployment
6. Slack/Discord notifications for failures

**Phase 2 (Health Checks)**: Will add health check validation to build workflow

**Phase 3 (Structured Logging)**: Will add log format validation

**Phase 4 (Metrics)**: Will add Prometheus metric validation

---

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [golangci-lint](https://golangci-lint.run/)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Codecov Documentation](https://docs.codecov.com/)
