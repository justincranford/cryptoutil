# Workflow Testing Guidelines

## Overview

This document provides guidelines for testing GitHub Actions workflows locally before pushing to GitHub, reducing iteration cycles and preventing CI/CD failures.

## Local Testing Tools

### 1. Act (Recommended for Simple Workflows)

**Installation**:

```powershell
winget install nektos.act
```

**Basic Usage**:

```bash
# Test specific workflow
act -W .github/workflows/ci-quality.yml

# Test workflow with specific event
act push -W .github/workflows/ci-quality.yml

# Test all workflows for push event
act push

# List available workflows
act -l
```

**Limitations**:

- Docker container overhead (slow startup)
- Limited service support (PostgreSQL, OTEL require Docker Compose)
- Windows-specific issues (path mappings, permissions)

### 2. cmd/workflow (Project-Specific Tool)

**Recommended Tool**: Custom Go-based workflow runner with Docker Compose integration

**Usage**:

```bash
# Test single workflow
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"  # 3-5 min

# Test full scan
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=full"   # 10-15 min

# Test multiple workflows
go run ./cmd/workflow -workflows=e2e,dast

# Test all workflows
go run ./cmd/workflow -workflows=all
```

**Advantages**:

- Native Docker Compose support
- Service health checks built-in
- OTEL collector integration
- Workflow-specific configurations
- Faster than Act (no container-in-container overhead)

**NEVER use `-t` timeout flag** - ALWAYS let cmd/workflow complete naturally

## Workflow Testing Strategy

### Phase 1: Unit Tests (No Services)

Test workflows that don't require Docker services:

```bash
# Quality checks (linting, formatting, builds)
go run ./cmd/workflow -workflows=quality

# Coverage collection
go run ./cmd/workflow -workflows=coverage

# Benchmark tests
go run ./cmd/workflow -workflows=benchmark

# Fuzz tests (15s per test)
go run ./cmd/workflow -workflows=fuzz

# Race detection
go run ./cmd/workflow -workflows=race

# SAST (static security analysis)
go run ./cmd/workflow -workflows=sast

# GitLeaks (secrets scanning)
go run ./cmd/workflow -workflows=gitleaks
```

**Expected Duration**: 2-5 minutes total

### Phase 2: Integration Tests (Require Services)

Test workflows that require Docker Compose services:

```bash
# E2E tests (full Docker stack)
go run ./cmd/workflow -workflows=e2e

# Load tests (full Docker stack)
go run ./cmd/workflow -workflows=load

# DAST (quick scan)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# DAST (full scan)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=full"
```

**Expected Duration**:

- E2E: 5-10 minutes
- Load: 5-10 minutes
- DAST (quick): 3-5 minutes
- DAST (full): 10-15 minutes

## Pre-Push Checklist

**Before pushing changes that affect workflows**:

1. ✅ Test unit workflows locally (quality, coverage, race)
2. ✅ Test integration workflows if service configs changed (e2e, load, dast)
3. ✅ Verify Docker Compose health checks pass
4. ✅ Check workflow logs for errors
5. ✅ Validate service connectivity (curl/wget health endpoints)
6. ✅ Push changes to GitHub
7. ✅ Monitor workflow runs via `gh run watch` or GitHub UI

## Common Workflow Failures

### 1. Dependency Version Conflicts

**Symptom**:

```
Error: github.com/goccy/go-yaml@v1.18.7 conflicts with parent requirement ^1.19.0
```

**Fix**:

```bash
go get -u github.com/goccy/go-yaml@latest
go get -u all  # Update all transitive dependencies
go mod tidy
go test ./...  # Verify tests pass
```

**Prevention**: Regularly run `go get -u all` before major releases

### 2. Container Startup Failures

**Symptom**:

```
Container compose-identity-authz-e2e-1  Error
dependency failed to start: container compose-identity-authz-e2e-1 exited (1)
```

**Diagnosis Steps**:

1. Download container logs from CI artifacts:

   ```bash
   gh run download <run-id> --name e2e-container-logs-<run-id>
   ```

2. Extract and view logs:

   ```powershell
   Expand-Archive container-logs_*.zip
   Get-Content compose-identity-authz-e2e-1.log
   ```

3. Identify root cause from actual error message (not just exit code 1)

**Common Root Causes**:

- TLS cert file required but not configured
- Database DSN required but not provided
- Credential mismatch (app vs database)
- Missing public HTTP server implementation

**Prevention**: Test Docker Compose locally before pushing:

```bash
docker compose -f deployments/compose/compose.yml up -d
docker compose ps  # Verify all services healthy
docker compose logs <service>  # Check for errors
docker compose down -v
```

### 3. Service Health Check Failures

**Symptom**:

```
Attempt 30/30 (backoff: 5s)
Testing: https://127.0.0.1:9090/admin/v1/readyz
❌ Not ready: https://127.0.0.1:9090/admin/v1/readyz
❌ Application failed to become ready within timeout
```

**Diagnosis**:

1. Check if services started successfully
2. Verify health check endpoint exists
3. Test health endpoint manually:

   ```bash
   curl -k https://127.0.0.1:9090/admin/v1/livez
   curl -k https://127.0.0.1:9090/admin/v1/readyz
   curl -k https://127.0.0.1:9090/admin/v1/healthz
   ```

**Common Root Causes**:

- Wrong healthcheck endpoint path (/health vs /admin/v1/livez)
- Service startup dependency issues
- Insufficient health check timeout
- TLS configuration mismatch (http vs https)

**Prevention**: Use consistent healthcheck patterns across all services

### 4. Port Conflicts (Docker Compose)

**Symptom**:

```
Error response from daemon: driver failed programming external connectivity on endpoint opentelemetry-collector-contrib: Bind for 0.0.0.0:4317 failed: port is already allocated
```

**Diagnosis**:

1. Check for duplicate port mappings:

   ```bash
   grep -r "ports:" deployments/compose/*.yml
   ```

2. Verify no services expose same ports to host

**Common Root Causes**:

- Multiple services include same telemetry compose file
- Shared OTEL collector ports (4317, 4318, 8888, 13133)
- Attempting to run multiple product stacks simultaneously

**Prevention**:

- Use container-to-container networking (no host port mappings for shared services)
- Test sequential deployments (CA, then JOSE, then Identity)
- Remove host port mappings for shared infrastructure

## Code Archaeology Pattern (Critical Discovery)

**When to Use**: Container crashes with zero symptom change after configuration fixes

**Steps**:

1. Download container logs from last 3-5 workflow runs
2. Compare log byte counts across runs:

   ```powershell
   Get-ChildItem *.log | Select-Object Name, Length
   ```

3. If byte count IDENTICAL despite fixes → implementation issue, not config
4. Compare with working service (e.g., CA vs Identity):

   ```bash
   tree internal/ca/server
   tree internal/identity/authz/server
   ```

5. Identify missing files (server.go, application.go, service.go)
6. Review Application.Start() code for missing initialization
7. Check NewApplication() for complete setup

**Pattern Recognition**:

- **Cascading errors**: Each fix changes error message (TLS → DSN → credentials)
- **Zero symptom change**: Fix applied but SAME crash = missing code
- **Decreasing byte count**: 331 → 313 → 196 bytes = earlier crash = deeper problem

**Time Saved**: 9 minutes (code archaeology) vs 60 minutes (config debugging)

**Reference**: See docs/WORKFLOW-FIXES-CONSOLIDATED.md Rounds 3-7

## Diagnostic Commands

### GitHub CLI Workflow Diagnostics

```bash
# List recent workflow runs
gh run list --limit 10

# View specific workflow run details
gh run view <run-id>

# View failed workflow logs
gh run view <run-id> --log-failed

# Download workflow artifacts
gh run download <run-id>

# Watch a running workflow
gh run watch <run-id>

# Re-run failed jobs
gh run rerun <run-id> --failed

# List workflows
gh workflow list

# View workflow file
gh workflow view <workflow-name>
```

### Docker Compose Diagnostics

```bash
# View service status
docker compose ps

# View service logs
docker compose logs <service>

# View logs with timestamps
docker compose logs -t <service>

# Follow logs in real-time
docker compose logs -f <service>

# Execute command in running container
docker compose exec <service> <command>

# View service health checks
docker compose ps --format json | jq '.[] | {name: .Name, status: .Status, health: .Health}'

# Restart specific service
docker compose restart <service>

# Stop and remove all containers
docker compose down -v
```

### Service Health Check Verification

```bash
# Test admin endpoints (HTTPS with self-signed cert)
curl -k https://127.0.0.1:9090/admin/v1/livez
curl -k https://127.0.0.1:9090/admin/v1/readyz
curl -k https://127.0.0.1:9090/admin/v1/healthz

# Test public endpoints (HTTPS)
curl -k https://127.0.0.1:8080/ui/swagger/doc.json  # KMS SQLite
curl -k https://127.0.0.1:8081/ui/swagger/doc.json  # KMS PostgreSQL 1
curl -k https://127.0.0.1:8082/ui/swagger/doc.json  # KMS PostgreSQL 2

# Test with wget (Alpine containers)
wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/v1/livez
```

## Workflow Timing Expectations

| Workflow | Services | Expected Duration | Notes |
|----------|----------|-------------------|-------|
| ci-quality | None | 2-3 minutes | Linting, formatting, builds |
| ci-coverage | None | 3-5 minutes | Test coverage collection |
| ci-benchmark | None | 2-4 minutes | Performance benchmarks |
| ci-fuzz | None | 5-10 minutes | Fuzz testing (15s per test) |
| ci-race | None | 5-10 minutes | Race detection (10x overhead) |
| ci-sast | None | 2-3 minutes | Static security analysis |
| ci-gitleaks | None | 1-2 minutes | Secrets scanning |
| ci-mutation | None | 15-20 minutes | Mutation testing (parallel) |
| ci-e2e | Full stack | 5-10 minutes | E2E integration tests |
| ci-load | Full stack | 5-10 minutes | Load testing |
| ci-dast (quick) | Full stack | 3-5 minutes | Quick security scan |
| ci-dast (full) | Full stack | 10-15 minutes | Comprehensive scan |

**Notes**:

- GitHub Actions runners are shared resources (variable CPU steal time)
- Add 50-100% margin to expected times in CI/CD vs local
- Parallel tests increase timing variability
- Docker service startup adds 1-2 minutes overhead

## Best Practices

### 1. Iterative Testing

**DO**:

- Test workflows locally BEFORE pushing
- Fix one issue at a time
- Verify fix works before moving to next issue
- Commit each fix independently

**DON'T**:

- Apply multiple fixes simultaneously
- Push without local verification
- Batch unrelated fixes in single commit
- Skip local testing for "simple" changes

### 2. Log Analysis

**DO**:

- Download container logs from CI artifacts
- Compare logs across multiple runs
- Look for actual error messages (not just exit codes)
- Track byte count changes (indicates earlier/later crash)

**DON'T**:

- Assume exit code 1 is enough diagnosis
- Apply fixes without reading actual error messages
- Ignore log byte count trends
- Skip log comparison across runs

### 3. Configuration vs Implementation

**DO**:

- Verify complete architecture exists BEFORE debugging config
- Compare with working services (e.g., CA)
- Check for missing files (server.go, application.go)
- Use code archaeology when zero symptom change occurs

**DON'T**:

- Keep applying config fixes when symptoms don't change
- Assume container crash is always configuration
- Debug configuration before verifying implementation complete
- Waste time on config when code is missing

### 4. Workflow Monitoring

**DO**:

- Push changes to GitHub after local validation
- Monitor workflow runs via `gh run watch`
- Check workflow status after 5-10 minutes
- Download artifacts if failures occur

**DON'T**:

- Push without local testing
- Ignore workflow failures
- Assume workflows will pass without verification
- Wait hours before checking workflow status

## Summary

**Local Testing Priority**:

1. **ALWAYS test locally first** - saves 5-10 minutes per iteration
2. **Use cmd/workflow for integration tests** - faster than Act
3. **Download and analyze container logs** - actual errors, not assumptions
4. **Code archaeology for zero symptom change** - missing code vs config
5. **Monitor GitHub workflows** - verify fixes work in CI/CD

**Time Investment**:

- Local testing: 2-5 minutes (unit) + 5-15 minutes (integration)
- GitHub workflow: 5-10 minute wait per push
- Savings: 3-6 iterations avoided = 15-60 minutes saved

**Quality Benefits**:

- Faster iteration cycles
- Earlier error detection
- Better diagnosis (actual error messages)
- Reduced CI/CD load
- Cleaner commit history
