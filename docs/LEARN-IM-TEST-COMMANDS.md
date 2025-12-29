
# learn-im Test Commands

## Unit Tests with Coverage

```powershell
# Run unit tests for all learn-im packages
go test ./internal/learn/... -short -coverprofile=./test-output/coverage_learn_unit.out

# Generate HTML coverage report
go tool cover -html=./test-output/coverage_learn_unit.out -o ./test-output/coverage_learn_unit.html
```

**Note**: Unit tests use `-short` flag to skip long-running integration tests.

## Integration Tests with Coverage

```powershell
# Run integration tests (uses PostgreSQL test-containers)
go test ./internal/learn/integration/... -coverprofile=./test-output/coverage_learn_integration.out

# Generate HTML coverage report
go tool cover -html=./test-output/coverage_learn_integration.out -o ./test-output/coverage_learn_integration.html
```

**Requirements**:

- Docker runtime available (for test-containers)
- Windows: May fail with "Docker rootless not supported" - use CI/CD for validation

## E2E Tests with Coverage

```powershell
# Run end-to-end tests (service + browser paths)
go test ./internal/learn/e2e/... -coverprofile=./test-output/coverage_learn_e2e.out

# Generate HTML coverage report
go tool cover -html=./test-output/coverage_learn_e2e.out -o ./test-output/coverage_learn_e2e.html
```

**Test Coverage**:

- 4 service path tests (`/service/api/v1/*`)
- 3 browser path tests (`/browser/api/v1/*`)

## Race Condition Detection

```powershell
# Run all tests with race detector (requires CGO_ENABLED=1)
# Note: 10× slower than normal execution
go test -race -count=2 ./internal/learn/...
```

**CGO Limitation**: Race detector requires CGO_ENABLED=1 and GCC compiler. Use CI/CD workflows for race detection if GCC not available locally.

## Docker Compose (Start/Use/Stop)

```powershell
# Start learn-im service
docker compose -f deployments/learn/compose.yml up -d

# Check service status
docker compose -f deployments/learn/compose.yml ps

# View logs (follow mode)
docker compose -f deployments/learn/compose.yml logs -f learn-im

# Test registration endpoint
curl -k https://localhost:8888/service/api/v1/users/register `
  -H 'Content-Type: application/json' `
  -d '{\"username\":\"alice\",\"password\":\"SecurePass123!\"}'

# Stop all services
docker compose -f deployments/learn/compose.yml down
```

## Demo Application (CLI)

```powershell
# Start learn-im in dev mode (SQLite in-memory)
go run ./cmd/learn-im -d

# In separate terminal - test registration
curl -k https://localhost:8888/service/api/v1/users/register `
  -H 'Content-Type: application/json' `
  -d '{\"username\":\"bob\",\"password\":\"SecurePass456!\"}'

# Stop with Ctrl+C
```

**CGO Limitation**: `go run` requires GCC compiler for modernc.org/sqlite. Use Docker Compose for local testing if GCC not available.

## Coverage Targets

| Code Type | Minimum Coverage |
|-----------|------------------|
| Production code | ≥95% |
| Infrastructure/utility code | ≥98% |

**Verify Coverage**:

```powershell
# Check coverage by function
go tool cover -func=./test-output/coverage_learn_unit.out
```

## Test Patterns

- **TestMain**: Used for heavyweight dependencies (PostgreSQL test-containers, HTTP servers)
- **Table-Driven Tests**: Used for multiple test cases with t.Parallel()
- **test-containers**: PostgreSQL containers for integration tests
- **Dynamic Ports**: All test servers use port 0 (dynamic allocation)
- **Unique Test Data**: UUIDv7 or magic constants (NEVER hardcoded values)
