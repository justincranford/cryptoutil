# Workflow Test Guideline

**Purpose**: Comprehensive guide for testing individual workflows locally, identifying gaps in local test strategy, and methodologies for analyzing testing effectiveness, quality, and results.

## Local Workflow Testing Methods

### 1. Manual Workflow Execution (`go run ./cmd/workflow`)

**Primary Method**: Use the workflow command to execute individual workflows locally.

```bash
# Execute single workflow
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Execute multiple workflows
go run ./cmd/workflow -workflows=dast,e2e -inputs="scan_profile=full"

# List available workflows
go run ./cmd/workflow -list

# Get workflow help
go run ./cmd/workflow -help
```

**Supported Workflows**:

- `ci-quality`: Linting, formatting, builds
- `ci-coverage`: Test coverage analysis
- `ci-benchmark`: Performance benchmarks
- `ci-fuzz`: Fuzz testing
- `ci-race`: Race condition detection
- `ci-sast`: Static security analysis
- `ci-gitleaks`: Secrets scanning
- `ci-dast`: Dynamic security testing (requires services)
- `ci-e2e`: End-to-end integration testing (requires services)
- `ci-load`: Load testing (requires services)

### 2. Direct Tool Execution

**For workflows without service dependencies**:

```bash
# Quality checks (equivalent to ci-quality)
golangci-lint run
go build ./...
go mod tidy

# Coverage (equivalent to ci-coverage)
go test ./... -coverprofile=./test-output/coverage.out
go tool cover -html=./test-output/coverage.out -o ./test-output/coverage.html

# Race detection (equivalent to ci-race)
go test -race -count=3 ./...

# Fuzz testing (equivalent to ci-fuzz)
go test -fuzz=FuzzTestName -fuzztime=15s ./pkg/path

# Benchmarking (equivalent to ci-benchmark)
go test -bench=. -benchmem ./pkg/path
```

### 3. Service-Dependent Workflow Testing

**Requires Docker Compose infrastructure**:

```bash
# Start services for DAST/E2E/Load testing
docker compose -f ./deployments/compose/compose.yml up -d

# Wait for services to be healthy
# Check health endpoints manually or use scripts

# Run workflow
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Stop services
docker compose -f ./deployments/compose/compose.yml down -v
```

**Service Health Verification**:

```bash
# KMS (SQLite)
curl -k https://localhost:8080/admin/v1/healthz

# KMS (PostgreSQL 1)
curl -k https://localhost:8081/admin/v1/healthz

# KMS (PostgreSQL 2)
curl -k https://localhost:8082/admin/v1/healthz

# Identity services
curl -k https://localhost:8180/admin/v1/healthz  # authz
curl -k https://localhost:8181/admin/v1/healthz  # idp
curl -k https://localhost:8182/admin/v1/healthz  # rs
```

### 4. Selective Package Testing

**Test individual packages before full workflow**:

```bash
# Test specific package
go test ./internal/identity/authz/...

# With coverage
go test -cover ./internal/identity/authz/...

# With race detection
go test -race ./internal/identity/authz/...

# With verbose output
go test -v ./internal/identity/authz/...
```

### 5. Pre-commit Hook Testing

**Test pre-commit hooks locally**:

```bash
# Run all hooks on staged files
pre-commit run

# Run specific hook
pre-commit run markdownlint-cli2

# Run on specific files
pre-commit run --files .github/copilot-instructions.md

# Test in CI environment
pre-commit run --all-files
```

## Gaps in Local Test Strategy

### 1. Service Startup Automation

**Gap**: Manual service startup for integration tests

**Current**: `docker compose up -d` + manual health checks
**Missing**: Automated service readiness verification
**Impact**: Unreliable test execution, false failures

**Recommended Solution**:

```bash
#!/bin/bash
# automated-service-start.sh
docker compose -f ./deployments/compose/compose.yml up -d

# Wait for services with timeout
timeout=300
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if curl -k -f https://localhost:8080/admin/v1/healthz > /dev/null 2>&1; then
        echo "Services ready"
        exit 0
    fi
    sleep 5
    elapsed=$((elapsed + 5))
done

echo "Services failed to start within $timeout seconds"
exit 1
```

### 2. Test Result Correlation

**Gap**: No correlation between local test results and CI workflow results

**Current**: Local tests run in isolation
**Missing**: Result comparison and delta analysis
**Impact**: Hard to predict CI failures from local testing

**Recommended Solution**: Create result comparison script

```bash
#!/bin/bash
# compare-results.sh
# Compare local test results with CI artifacts

# Download CI artifacts
gh run download <run-id> --dir ci-results

# Compare coverage
diff <(sort local-coverage.out) <(sort ci-results/coverage.out)

# Compare test outputs
diff local-test.log ci-results/test.log
```

### 3. Environment Parity

**Gap**: Local environment differs from CI environment

**Current**: Local uses host Go/PostgreSQL, CI uses containers
**Missing**: Containerized local testing
**Impact**: Environment-specific failures not caught locally

**Recommended Solution**: Add containerized test option

```bash
# Test in containers like CI
docker run --rm -v $(pwd):/src -w /src golang:1.25.5 \
    go test ./... -coverprofile=coverage.out
```

### 4. Performance Regression Detection

**Gap**: No local performance baseline comparison

**Current**: Benchmarks run but not tracked
**Missing**: Historical performance comparison
**Impact**: Performance regressions not detected until CI

**Recommended Solution**: Performance tracking script

```bash
#!/bin/bash
# benchmark-track.sh

# Run benchmarks
go test -bench=. -benchmem ./... > benchmark.txt

# Compare with baseline
if [ -f benchmark-baseline.txt ]; then
    benchstat benchmark-baseline.txt benchmark.txt
fi

# Update baseline if improved
cp benchmark.txt benchmark-baseline.txt
```

### 5. Security Testing Depth

**Gap**: Limited local security testing scope

**Current**: Basic linting and gosec
**Missing**: Comprehensive security scanning
**Impact**: Security issues not caught locally

**Recommended Solution**: Enhanced local security testing

```bash
# Run comprehensive security checks
gosec ./...
trivy fs --security-checks vuln,config,secret .
semgrep --config auto .
```

### 6. Integration Test Isolation

**Gap**: Integration tests affect each other

**Current**: All services start together
**Missing**: Per-service integration testing
**Impact**: Hard to isolate integration failures

**Recommended Solution**: Service-specific test scripts

```bash
#!/bin/bash
# test-identity-only.sh

# Start only identity services
docker compose -f ./deployments/compose/compose.yml up -d identity-authz identity-idp identity-rs

# Test identity workflows
go run ./cmd/workflow -workflows=e2e -inputs="services=identity"

# Cleanup
docker compose -f ./deployments/compose/compose.yml down -v
```

## Testing Effectiveness Methodologies

### 1. Coverage Analysis

**Mutation Score Tracking**:

```bash
# Run mutation testing
gremlins unleash --tags=!integration

# Track scores over time
echo "$(date): $(grep 'Mutation score' gremlins-report.txt)" >> mutation-history.txt
```

**Coverage Gap Analysis**:

```bash
# Find uncovered functions
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%" | sort -k3 -n
```

### 2. Test Quality Metrics

**Test Execution Time Analysis**:

```bash
# Measure test execution time
time go test ./...

# Identify slow tests
go test -v ./... 2>&1 | grep -E "^=== RUN|^--- PASS" | \
    awk '/=== RUN/{test=$2} /--- PASS/{print test, $3}' | \
    sort -k2 -n
```

**Flaky Test Detection**:

```bash
# Run tests multiple times to detect flakes
for i in {1..5}; do
    echo "Run $i:"
    go test -run TestFlaky ./... || echo "FAILED"
done
```

### 3. Result Quality Assessment

**Test Result Consistency**:

```bash
# Compare test results across runs
go test -json ./... > test-run-1.json
go test -json ./... > test-run-2.json

# Compare outputs
diff test-run-1.json test-run-2.json
```

**Error Pattern Analysis**:

```bash
# Analyze common failure patterns
go test -v ./... 2>&1 | grep -i "fail\|error" | \
    sed 's/.*\(FAIL\|ERROR\).*/\1/' | sort | uniq -c | sort -nr
```

### 4. Integration Test Effectiveness

**Service Interaction Coverage**:

```bash
# Check which service combinations are tested
grep -r "federation\|service.*url" test/ | \
    grep -o "https://[a-z0-9-]*:[0-9]*" | sort | uniq
```

**API Contract Verification**:

```bash
# Verify API contracts match between services
find . -name "*.yaml" -exec grep -l "openapi\|swagger" {} \; | \
    xargs -I {} sh -c 'echo "=== {} ==="; head -20 {}'
```

## Quality Assurance Methodologies

### 1. Test Suite Health Metrics

**Test Suite Statistics**:

```bash
# Overall test statistics
go test -v ./... 2>&1 | grep -E "^=== RUN|^--- (PASS|FAIL|SKIP)" | \
    awk '
        /=== RUN/ {tests++}
        /--- PASS/ {passes++}
        /--- FAIL/ {fails++}
        /--- SKIP/ {skips++}
        END {
            print "Total tests:", tests
            print "Passed:", passes
            print "Failed:", fails
            print "Skipped:", skips
            print "Pass rate:", (passes/tests)*100 "%"
        }
    '
```

### 2. Code Quality Correlation

**Linting vs Test Quality**:

```bash
# Compare linting issues with test failures
golangci-lint run --out-format=json > lint-results.json
go test -json ./... > test-results.json

# Correlate issues
jq -s '.[0] as $lint | .[1] as $test |
    $lint[] | select(.Pos.Filename | contains("test")) |
    . + {test_failures: ($test | map(select(.Package == (.Pos.Filename | sub("_test.go"; ""))) | select(.Action == "fail") | length))} |
    select(.test_failures > 0)' lint-results.json test-results.json
```

### 3. Performance Benchmarking

**Benchmark Result Analysis**:

```bash
# Analyze benchmark results for regressions
go test -bench=. -benchmem ./... | \
    awk '/^Benchmark/ {print $1, $3, $5}' | \
    sort -k2 -n | tail -10  # Slowest benchmarks
```

**Memory Leak Detection**:

```bash
# Check for memory leaks in benchmarks
go test -bench=. -benchmem -memprofile=mem.prof ./...
go tool pprof -top mem.prof
```

## Result Analysis Frameworks

### 1. Automated Test Reporting

**Generate Comprehensive Reports**:

```bash
#!/bin/bash
# generate-test-report.sh

echo "# Test Report - $(date)" > test-report.md
echo "" >> test-report.md

# Coverage summary
echo "## Coverage" >> test-report.md
go test -cover ./... | grep -E "coverage|ok|FAIL" >> test-report.md
echo "" >> test-report.md

# Test timing
echo "## Test Timing" >> test-report.md
go test -v ./... 2>&1 | grep -E "^=== RUN|^--- PASS" | \
    awk '/=== RUN/{test=$2; start=$0} /--- PASS/{print test, $3}' | \
    sort -k2 -nr | head -10 >> test-report.md
echo "" >> test-report.md

# Mutation score
echo "## Mutation Testing" >> test-report.md
if command -v gremlins &> /dev/null; then
    gremlins unleash --tags=!integration | grep "Mutation score" >> test-report.md
fi
```

### 2. CI/CD Result Comparison

**Compare Local vs CI Results**:

```bash
#!/bin/bash
# compare-ci-local.sh

# Get latest CI run ID
CI_RUN_ID=$(gh run list --workflow=ci-quality --limit=1 --json databaseId --jq '.[0].databaseId')

# Download CI artifacts
gh run download $CI_RUN_ID --dir ci-artifacts

# Compare results
echo "Coverage comparison:"
diff <(sort coverage.out) <(sort ci-artifacts/coverage.out) || echo "Coverage differs"

echo "Test result comparison:"
diff test-results.json ci-artifacts/test-results.json || echo "Test results differ"
```

### 3. Trend Analysis

**Track Metrics Over Time**:

```bash
#!/bin/bash
# track-metrics.sh

DATE=$(date +%Y-%m-%d)

# Coverage trend
COVERAGE=$(go test -cover ./... 2>&1 | grep "coverage" | awk '{print $5}')
echo "$DATE,coverage,$COVERAGE" >> metrics.csv

# Test count
TEST_COUNT=$(go test -v ./... 2>&1 | grep -c "^=== RUN")
echo "$DATE,test_count,$TEST_COUNT" >> metrics.csv

# Execution time
EXEC_TIME=$(time go test ./... 2>&1 | grep real | awk '{print $2}')
echo "$DATE,exec_time,$EXEC_TIME" >> metrics.csv
```

## Recommendations for Improvement

### 1. Automated Testing Infrastructure

- Implement automated service startup/shutdown scripts
- Add containerized testing options for environment parity
- Create per-service integration test suites

### 2. Result Analysis Tools

- Develop scripts for comparing local vs CI results
- Implement performance regression detection
- Add automated test quality metrics collection

### 3. Security Testing Enhancement

- Integrate comprehensive security scanning tools locally
- Add security test result correlation with CI
- Implement security testing in pre-commit hooks

### 4. Quality Assurance Framework

- Establish test suite health dashboards
- Implement automated flaky test detection
- Add code quality correlation analysis

### 5. Documentation and Training

- Document all local testing workflows
- Create troubleshooting guides for common issues
- Train team on effective local testing practices

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
- Shared OTEL collector ports (4317, 4318, 8070, 13133)
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
