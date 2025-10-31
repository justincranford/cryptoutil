---
description: "Instructions for CI/CD workflow configuration and service connectivity verification"
applyTo: ".github/workflows/*.yml"
---
# CI/CD Workflow Instructions

## CI/CD Cost Efficiency

**Optimize workflow execution to minimize GitHub Actions billed minutes:**

- **Trigger optimization**: Only run workflows on relevant file changes (use path filters)
- **Matrix builds**: Keep job matrices minimal, avoid unnecessary combinations
- **Dependency caching**: Use Go module caching to reduce build times
- **Skip trivial changes**: Use conditionals to skip jobs for docs-only or trivial changes
- **Job filters**: Apply conditionals to avoid wasteful runs (e.g., skip deployment on forks)

**Example patterns:**
```yaml
on:
  push:
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - '.github/workflows/**'
    paths-ignore:
      - '**.md'
      - 'docs/**'
```

## Workflow Architecture Overview

The CI/CD pipeline consists of 6 specialized workflows with different service orchestration approaches:

| Workflow | File | Services | Connectivity Verification | Primary Purpose |
|----------|------|----------|---------------------------|-----------------|
| Quality | `ci-quality.yml` | None | N/A | Code quality, linting, formatting, builds |
| SAST | `ci-sast.yml` | None | N/A | Static security analysis |
| Robustness | `ci-robust.yml` | None | N/A | Concurrency, fuzz, benchmarks |
| DAST | `ci-dast.yml` | Standalone app | Bash/curl | Dynamic security scanning |
| E2E | `ci-e2e.yml` | Full Docker stack | Go test infra | Integration testing |
| Load | `ci-load.yml` | Full Docker stack | Bash/curl | Performance testing |

## Service Connectivity Verification Patterns

### Pattern 1: No Service Verification (Quality, SAST, Robustness)

**When to Use**: Workflows that perform static analysis, unit tests, or builds without starting services

**Workflows**: `ci-quality.yml`, `ci-sast.yml`, `ci-robust.yml`

**Rationale**:
- Static analysis tools don't need running services
- Unit tests and fuzz tests run in isolation
- Container builds verify build process, not runtime behavior

**Implementation**: No connectivity verification steps needed

---

### Pattern 2: Go-Based E2E Infrastructure (E2E)

**When to Use**: Comprehensive integration testing with full Docker Compose orchestration

**Workflow**: `ci-e2e.yml`

**Key Components**:
```go
// internal/test/e2e/infrastructure.go
- Docker Compose orchestration (start/stop services)
- Multi-stage health checking (Docker health + HTTP connectivity)
- Exponential backoff retry logic

// internal/test/e2e/docker_health.go  
- Parse docker compose ps JSON output
- Handle 3 service types: jobs, services with native health, services with healthcheck jobs
- Comprehensive health status determination

// internal/test/e2e/http_utils.go
- CreateInsecureHTTPClient() with InsecureSkipVerify for self-signed certs
- HTTP GET requests with context for timeouts
```

**Verification Flow**:
1. **Docker Health Check**: `docker compose ps --format json` → parse service health status
2. **HTTP Connectivity**: Go HTTP client verifies reachability of all endpoints
3. **Multi-Endpoint Verification**: Public APIs, private health endpoints, observability services

**Why Go Infrastructure**:
- Type-safe service orchestration
- Comprehensive error handling
- Reusable across different test suites
- Integration with Go testing framework
- Better debugging with stack traces

---

### Pattern 3: Bash/Curl Verification (DAST, Load)

**When to Use**: Workflows that need quick service readiness checks before external tools (ZAP, Nuclei, Gatling)

**Workflows**: `ci-dast.yml`, `ci-load.yml`

**CRITICAL: Correct Curl Pattern for HTTPS with Self-Signed Certificates**

```bash
# ✅ CORRECT: curl with proper flags for HTTPS + self-signed certs
check_endpoint() {
  local url=$1
  local name=$2
  local max_attempts=30
  local attempt=1
  local backoff=1

  while [ $attempt -le $max_attempts ]; do
    echo "Attempt $attempt/$max_attempts: Checking $name... (backoff: ${backoff}s)"

    # -s: silent (no progress bar)
    # -k: insecure (skip TLS certificate verification for self-signed certs)
    # -f: fail on HTTP errors (4xx, 5xx)
    # --connect-timeout 10: connection timeout
    # --max-time 15: total request timeout
    if curl -skf --connect-timeout 10 --max-time 15 "$url" -o /tmp/${name}_response.json 2>/dev/null; then
      if [ -s /tmp/${name}_response.json ]; then
        echo "✅ $name is ready (response size: $(wc -c < /tmp/${name}_response.json) bytes)"
        rm -f /tmp/${name}_response.json
        return 0
      fi
    fi

    sleep $backoff
    # Exponential backoff with max 5 seconds
    backoff=$((backoff < 5 ? backoff + 1 : 5))
    attempt=$((attempt + 1))
  done

  echo "❌ $name failed to respond after $max_attempts attempts"
  return 1
}

# Check HTTPS endpoints (cryptoutil uses self-signed certs in CI)
check_endpoint "https://127.0.0.1:8080/ui/swagger/doc.json" "cryptoutil-sqlite"
check_endpoint "https://127.0.0.1:8081/ui/swagger/doc.json" "cryptoutil-postgres-1"
check_endpoint "https://127.0.0.1:8082/ui/swagger/doc.json" "cryptoutil-postgres-2"
```

**Common Mistakes to AVOID**:

```bash
# ❌ WRONG: wget does not reliably verify HTTPS with self-signed certs
wget --no-check-certificate --spider "$url"

# ❌ WRONG: Missing exponential backoff (hammers service during startup)
while [ $attempt -le 30 ]; do
  curl -skf "$url" && break
  sleep 2  # Fixed delay doesn't adapt to service startup time
done

# ❌ WRONG: Not verifying response body (connection might succeed but return empty)
if curl -skf "$url" >/dev/null 2>&1; then
  echo "Ready"  # Could be ready but returning no data
fi

# ❌ WRONG: Using localhost instead of 127.0.0.1 (IPv6 vs IPv4 ambiguity)
curl -skf "https://localhost:8080/..."  # May resolve to ::1 (IPv6)

# ❌ WRONG: Insufficient timeouts for containerized environments
curl -skf --connect-timeout 2 --max-time 5 "$url"  # Too short for Docker
```

**Why This Pattern Works**:
- `-k` flag handles self-signed TLS certificates (common in CI environments)
- Verifies actual response data (not just connection success)
- Exponential backoff prevents overwhelming services during slow startup
- Generous timeouts (10s connect, 15s total) accommodate containerized environments
- `127.0.0.1` ensures predictable IPv4 resolution (see localhost-vs-ip.instructions.md)

**When Bash/Curl is Preferred Over Go**:
- Simple readiness checks before external tools run
- Workflow already uses bash for service orchestration
- No need for complex retry logic or error handling
- External scanning tools (ZAP, Nuclei) need quick "is it up?" check

---

### Choosing the Right Pattern

**Use Go-Based Infrastructure When**:
- Running comprehensive integration tests
- Need detailed health status reporting
- Multiple service types with different health check mechanisms
- Require type-safe service orchestration
- Tests need to interact with services programmatically

**Use Bash/Curl Verification When**:
- Simple "is it ready?" check before external tools
- Workflow already orchestrates services via bash/docker commands
- External tools (ZAP, Nuclei, Gatling) are the primary test executors
- Minimal infrastructure needed (just readiness verification)

**Use No Verification When**:
- Static analysis or compilation only
- Unit tests run in complete isolation
- No services need to be started

---

## Go Version Consistency
- **ALWAYS use the same Go version as specified in go.mod** for all CI/CD workflows
- Current project Go version: **1.25.1** (check go.mod file)
- Set `GO_VERSION: '1.25.1'` in workflow environment variables
- Use `go-version: ${{ env.GO_VERSION }}` in setup-go actions

## Version Management
- When updating Go version, update ALL workflow files consistently:
  - `.github/workflows/ci.yml`
  - `.github/workflows/dast.yml`  
  - Any other workflows using Go
- Verify go.mod version matches CI/CD workflows before committing

## Best Practices
- Use environment variables for version consistency across jobs
- Pin to specific patch versions (e.g., '1.25.1', not '1.25' or '^1.25')
- Test locally with the same Go version used in CI/CD
- Update Docker base images to match Go version when applicable

## Configuration Management

### Application Configuration (Production/CI Deployments)
- **ALWAYS use config files** for production and CI application deployments
- **Example**: `cryptoutil server start --config configs/production/config.yml`
- **Database**: Config files should specify actual database connections (PostgreSQL for production)
- **CI/CD Pattern**: Copy and modify base config files for different environments rather than using environment variables
- **Why**: Config files are version-controlled, documented, and prevent environment variable naming mistakes

### Environment Variables (When Necessary)
- **ONLY use for exceptional cases** when Docker/Kubernetes secrets or config files cannot be used
- **NEVER use environment variables for secrets in production** - always prefer Docker secrets or Kubernetes secrets
- Application uses Viper with `CRYPTOUTIL_` prefix: `CRYPTOUTIL_DATABASE_URL`, `CRYPTOUTIL_LOG_LEVEL`, etc.
- **NEVER use non-standard environment variable names** like `POSTGRES_URL` - they will be ignored by the application
- **Use sparingly**: Only for emergency overrides or local development when secrets infrastructure isn't available
- Check `config.go` for the exact setting names and their corresponding environment variables

### Test Configuration (Development/Testing)
- **Tests ALWAYS use SQLite in-memory databases** regardless of config file database-url settings
- When running `cryptoutil server start --dev`, the application automatically switches to SQLite for development/testing
- **Config files in tests**: Can specify any database URL (even PostgreSQL) - tests will ignore it and use SQLite
- **CI/CD test workflows should NOT include PostgreSQL services** - tests use SQLite automatically for isolation and speed
- **Why**: Ensures test isolation, faster execution, and eliminates database setup complexity in CI/CD

## Go Module Caching Best Practices

### Use `cache: true` on `setup-go` Action
- **Preferred**: `cache: true` on `actions/setup-go@v6`
- **Why**: Automatic, self-healing, prevents tar extraction conflicts
- **Avoid**: Manual `actions/cache@v4` for Go modules (brittle, requires workarounds)

### Cache Key Strategy
- Use `go.sum` hash for cache invalidation
- Include OS in key for cross-platform compatibility
- Consider dependency count for large monorepos

### Troubleshooting
- Cache misses: Check `go.sum` changes
- Cache corruption: Let `setup-go` handle it automatically
- Performance issues: Monitor cache hit rates in workflow logs

## Build Flags and Linking

### Static Linking Requirement
- **ALWAYS use static linking** for both CI and Docker builds to ensure maximum portability
- Use `-extldflags '-static'` in ldflags for static linking
- Validate static linking in Docker builds with `ldd` check

### Debug Symbols vs Size Trade-offs
- **Performance and diagnostics prioritized over binary size**
- **CI builds**: Use `-s -extldflags '-static'` (strip symbol table but keep DWARF debug symbols with `-w` removed)
  - Static linking for maximum portability across CI environments
  - Retains debug symbols for troubleshooting test failures and CI diagnostics
  - Smaller than full debug build but still debuggable
- **Docker builds**: Use `-s -extldflags '-static'` (strip symbol table but keep DWARF debug symbols)
  - Static linking for container portability
  - Debug symbols retained for production troubleshooting
- **NEVER use `-w`** in either context (removes DWARF debug symbols, hurts diagnostics)

### Flag Explanations
- `-s`: Strip symbol table (reduces size, keeps DWARF debug symbols)
- `-w`: Strip DWARF debug symbols (breaks debugging, never use)
- `-extldflags '-static'`: Force static linking with external linker

## Artifact Management Best Practices

### Actions Tab Artifacts (Downloadable)
- **ALWAYS upload artifacts to Actions tab** for any generated reports, logs, or outputs that users/developers might need to download
- Use `actions/upload-artifact@v4` with descriptive names and `if: always()` condition
- Include retention policies: `retention-days: 2` for reports, `retention-days: 1` for temporary files
- **Examples**:
  - Test coverage reports
  - Security scan reports (SARIF files)
  - Build artifacts
  - Log files and diagnostics
  - SBOM files
  - Mutation testing reports

### Security Tab Integration (SARIF)
- **ALWAYS upload security findings to GitHub Security tab** using `github/codeql-action/upload-sarif@v3`
- Upload SARIF files for SAST, DAST, container scanning, and dependency analysis results
- Use `if: always() && hashFiles('file.sarif') != ''` to avoid failures when no issues found
- **Dual upload pattern**: Upload both to Security tab (for visualization) AND Actions tab (for downloadability)
- **Examples**:
  - Staticcheck SARIF results
  - Trivy vulnerability scans
  - CodeQL analysis results
  - OWASP ZAP reports
  - Nuclei security scans

### Additional Best Practices
- Use consistent artifact naming conventions across workflows
- Include timestamps or run IDs in artifact names when multiple runs possible
- Compress large artifacts to reduce storage costs
- Document artifact contents in workflow comments
- Clean up temporary artifacts within workflows to avoid clutter
