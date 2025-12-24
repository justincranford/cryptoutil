# CI/CD Workflow Configuration - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/04-01.github.instructions.md`

## Cost Efficiency Patterns

### Path Filters and Matrix Optimization

**Use path filters to trigger workflows only on relevant changes**:

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'internal/**'
      - '!**/*.md'  # Exclude documentation
  pull_request:
    paths-ignore:
      - 'docs/**'
      - '**.md'
```

**Minimal matrix builds**:
- Skip jobs for docs-only changes
- Use conditional steps: `if: steps.changes.outputs.src == 'true'`
- Combine related checks in single job when possible

**Caching Strategy**:
- Use `cache: true` on `actions/setup-go@v6` (built-in caching)
- NEVER use manual cache actions (actions/cache@v3) for Go modules
- Let GitHub Actions handle Go dependency caching automatically

---

## PostgreSQL Service Configuration - CRITICAL

### Three Deployment Patterns

**Pattern 1: Unit/Integration Tests (Recommended)**

**Use test-containers with randomized database credentials**:

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupPostgres(t *testing.T) *postgres.PostgresContainer {
    ctx := context.Background()
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:18"),
        postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewV7().String())),
        postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewV7().String())),
        postgres.WithPassword(googleUuid.NewV7().String()),
    )
    require.NoError(t, err)
    return container
}
```

**Why This Matters**:
- Docker containers provide isolation (no port conflicts)
- Randomized credentials prevent cross-test contamination
- NEVER use environment variables for test database credentials

**Pattern 2: E2E Tests (Production-Like)**

**Use Docker Compose with Docker secrets**:

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:18
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_user
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB_FILE: /run/secrets/postgres_db

  app:
    secrets:
      - database_url
    command: ["--database-url=file:///run/secrets/database_url"]

secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
  database_url:
    file: ./secrets/database_url.secret
```

**Why This Matters**:
- Mount secrets to `/run/secrets/` in containers
- Application reads credentials from secret files
- Enforces secure patterns for production-like testing

**Pattern 3: GitHub Workflows (Legacy Only)**

**Use ONLY when test-containers not feasible**:

```yaml
services:
  postgres:
    image: postgres:18
    env:
      POSTGRES_DB: cryptoutil_test
      POSTGRES_USER: cryptoutil
      POSTGRES_PASSWORD: cryptoutil_test_password
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
    ports:
      - 5432:5432
```

**When to Use**:
- Legacy tests requiring specific PostgreSQL configuration
- Tests not yet migrated to test-containers pattern
- Temporary workaround during migration phase

**Rationale**: Environment variables are insecure even for testing; test-containers provide isolated, randomized credentials; Docker secrets enforce secure patterns.

---

## Workflow Matrix Reference

### Continuous Integration Workflows

| Workflow | Trigger | Purpose | Artifacts |
|----------|---------|---------|-----------|
| ci-quality | Push to main, PR | Linting, formatting, build validation | golangci-lint reports |
| ci-test | Push to main, PR | Unit and integration tests | Test reports, coverage |
| ci-coverage | Push to main, PR | Test coverage analysis | Coverage HTML/JSON |
| ci-benchmark | Manual, weekly | Performance benchmarks | Benchmark results |
| ci-mutation | Manual | Mutation testing with gremlins | Gremlins reports |
| ci-race | Push to main, PR | Race condition detection | Race detector logs |

### Security Workflows

| Workflow | Trigger | Purpose | Artifacts |
|----------|---------|---------|-----------|
| ci-sast | Push to main, PR | Static security analysis (gosec, semgrep) | SARIF reports |
| ci-gitleaks | Push to main, PR | Secrets scanning | Gitleaks reports |
| ci-dast | Manual | Dynamic security testing (Nuclei, ZAP) | DAST scan results |

### Integration and Load Testing

| Workflow | Trigger | Purpose | Artifacts |
|----------|---------|---------|-----------|
| ci-e2e | Manual, nightly | End-to-end integration tests | E2E test logs |
| ci-load | Manual | Load and stress testing (Gatling) | Gatling reports |

---

## Configuration Management Patterns

### Config Files vs Environment Variables

**MANDATORY Rules**:
- **Production/CI**: ALWAYS use config files (YAML, TOML)
- **Tests**: Use SQLite in-memory (`--dev` flag) OR test-containers for PostgreSQL
- **Secrets**: ALWAYS use Docker secrets (NEVER environment variables)

**Examples**:

```yaml
# Production config (configs/production.yml)
database-url: "file:///run/secrets/database_url"
tls-cert: "file:///run/secrets/tls_cert"
tls-key: "file:///run/secrets/tls_key"
```

```go
// Test config (in-memory SQLite)
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    return db
}
```

---

## Artifact Management

### Upload Patterns

**Test artifacts** (always upload even on failure):

```yaml
- name: Upload test reports
  if: always()  # CRITICAL - upload even if tests fail
  uses: actions/upload-artifact@v4
  with:
    name: test-reports-${{ matrix.os }}
    path: ./test-output/*.json
    retention-days: 1  # Temporary artifacts

- name: Upload coverage reports
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: coverage-reports-${{ matrix.os }}
    path: ./test-output/coverage*.html
    retention-days: 7  # Keep for analysis
```

**Security scan results** (upload to GitHub Security tab):

```yaml
- name: Upload SARIF to Security tab
  if: always()
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: ./test-output/gosec.sarif
    category: gosec

- name: Upload Semgrep SARIF
  if: always()
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: ./test-output/semgrep.sarif
    category: semgrep
```

### Retention Policies

| Artifact Type | Retention | Rationale |
|---------------|-----------|-----------|
| Temporary test logs | 1 day | Quick debugging, disposable |
| Coverage reports | 7 days | Analysis, comparison |
| Security scan results | 30 days | Compliance, trending |
| Benchmark results | 30 days | Performance tracking |
| DAST scan results | 30 days | Security audits |

---

## Act Workflow Testing (Local Execution)

### cmd/workflow Utility Usage

**Execute workflows locally** (no Docker Desktop required):

```bash
# Quick DAST scan (3-5 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Full DAST scan (10-15 minutes)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=full"

# Multiple workflows sequentially
go run ./cmd/workflow -workflows=e2e,dast

# E2E with specific profile
go run ./cmd/workflow -workflows=e2e -inputs="compose_profile=sqlite"
```

**CRITICAL Rules**:
- NEVER use `-t` timeout flag (ALWAYS let cmd/workflow complete naturally)
- ALWAYS specify `-workflows` flag with comma-separated workflow names
- Use `-inputs` for workflow-specific parameters (key=value format)

**Why cmd/workflow Matters**:
- Test workflows locally before pushing to GitHub
- Debug workflow issues without triggering CI/CD
- Faster iteration cycle (no git commit/push required)

---

## Diagnostic Logging Standards

### Timing Requirements for Long-Running Steps

**Steps >10 seconds MUST include timing**:

```yaml
- name: Build Docker image
  run: |
    START_TIME=$(date +%s)
    echo "📋 Starting Docker build: $(date -u +'%Y-%m-%d %H:%M:%S UTC')"

    docker build -t cryptoutil:latest .

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo "⏱️ Docker build completed in ${DURATION}s"
```

### Emoji Standards for Workflow Logs

| Emoji | Meaning | Usage |
|-------|---------|-------|
| 📋 | Start | Step initiation |
| 📅 | Timestamps | Start/end times |
| ⏱️ | Duration | Elapsed time |
| ✅ | Success | Step completed successfully |
| ❌ | Error | Step failed |
| 🔍 | Diagnostic | Debug information |
| 🏗️ | Build | Build operations |
| 🧪 | Test | Test execution |
| 🔒 | Security | Security scans |

**Rationale**: Visual scanning of GitHub Actions logs, quick identification of bottlenecks and failures

---

## GitHub CLI for Workflow Diagnostics

### Essential gh Commands

**List and View Workflows**:

```bash
# List recent workflow runs
gh run list --limit 10

# View specific workflow run details
gh run view <run-id>

# View failed workflow logs ONLY
gh run view <run-id> --log-failed

# Watch a running workflow (live updates)
gh run watch <run-id>
```

**Artifact Management**:

```bash
# Download workflow artifacts
gh run download <run-id>

# Download specific artifact
gh run download <run-id> --name coverage-reports
```

**Re-run and Cancel Operations**:

```bash
# Re-run ALL jobs
gh run rerun <run-id>

# Re-run ONLY failed jobs
gh run rerun <run-id> --failed

# Cancel stuck runs
gh run cancel <run-id>
```

**Workflow Management**:

```bash
# List all workflows
gh workflow list

# View workflow YAML
gh workflow view <workflow-name>

# Manually trigger workflow
gh workflow run <workflow-name>

# Check workflow status
gh workflow view <workflow-name> --yaml
```

### Common Diagnostic Patterns

**Find Recent Failures**:

```bash
# Recent failures on main branch
gh run list --branch main --status failure --limit 5

# All failures in last 7 days
gh run list --status failure --limit 50
```

**Debug Specific Failure**:

```bash
# Get full logs for debugging
gh run view <run-id> --log

# Download logs to file
gh run view <run-id> --log > workflow-logs.txt

# View only failed jobs
gh run view <run-id> --log-failed > failed-jobs.txt
```

**Bulk Operations**:

```bash
# Cancel all running workflows
gh run list --status in_progress --json databaseId --jq '.[].databaseId' | xargs -I {} gh run cancel {}

# Re-run all failed workflows from today
gh run list --status failure --json databaseId --jq '.[].databaseId' | head -10 | xargs -I {} gh run rerun {} --failed
```

---

## Variable Expansion in Heredocs - CRITICAL

### Bash Heredoc Pattern

**ALWAYS use curly braces `${VAR}` syntax for variable expansion**:

```yaml
- name: Generate config
  run: |
    cat > ./configs/test/config.yml <<EOF
    database-url: "postgres://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_NAME}?sslmode=disable"
    bind-public-address: "${APP_BIND_PUBLIC_ADDRESS}"
    bind-public-port: ${APP_BIND_PUBLIC_PORT}
    EOF
```

**CRITICAL Rules**:
- ✅ ALWAYS use `${VAR}` syntax (curly braces) for explicit variable expansion
- ✅ ALWAYS verify generated config files have expanded values (not literal $VAR strings)
- ✅ ALWAYS test config generation with `cat config.yml` step in workflow
- ❌ NEVER use `$VAR` syntax in heredocs (may write literal "$VAR" to file)
- ❌ NEVER rely on implicit variable expansion behavior (shell-dependent, error-prone)

### Historical Mistakes (CI-DAST Lessons)

**Problem**: ci-dast used `$POSTGRES_USER` instead of `${POSTGRES_USER}` in heredoc
**Result**: Heredoc wrote literal string "$POSTGRES_USER" to config.yml
**Impact**: Application read literal "$POSTGRES_USER" as username, defaulted to 'root'
**PostgreSQL Error**: "FATAL: role 'root' does not exist" (27+ occurrences)
**Fix**: Change all `$VAR` → `${VAR}` in heredoc (7 variables affected)

**Evidence**:
- Workflow logs showed: `database-url: ***$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_NAME` (literal dollar signs)
- PostgreSQL logs showed: "role 'root' does not exist" (wrong username)
- Config verification step revealed literal $VAR strings instead of expanded values

---

## Debugging Workflow PostgreSQL Issues

### Root Cause Analysis Checklist

**When PostgreSQL connection errors occur**:

1. **Download workflow artifacts** - container logs, PostgreSQL logs
2. **Check PostgreSQL logs** - look for "FATAL: role '<username>' does not exist" patterns
3. **Verify credentials** - compare workflow env vars vs service container env vars
4. **Check config generation** - search workflow logs for generated config output
5. **Verify variable expansion** - ensure config has expanded values, not literal $VAR strings
6. **Compare with working workflows** - ci-coverage, ci-mutation, ci-race use correct pattern

### Evidence-Based Pattern Recognition

**Error Frequency Analysis**:
- 27+ occurrences over 5 minutes = persistent misconfiguration (NOT transient failure)
- Every 10 seconds = health check retry loop
- Connection rejected immediately = credentials issue (NOT network/timing)

**Log Correlation**:
- PostgreSQL startup succeeded = database server healthy
- Connections rejected = credentials mismatch
- Literal "$POSTGRES_HOST" in logs = variable expansion issue

### Preventive Checklist for All Services

**Before deploying new service or modifying workflow**:

- [ ] PostgreSQL credentials use cryptoutil_test/cryptoutil/cryptoutil_test_password
- [ ] Workflow env vars match PostgreSQL service container env vars
- [ ] Config generation uses `${VAR}` syntax (curly braces) for ALL variables
- [ ] Config generation includes verification step (`cat config.yml` in workflow logs)
- [ ] Workflow tested locally with Docker Compose before pushing
- [ ] Workflow logs reviewed after first run to verify config correctness
- [ ] PostgreSQL service health check configured (pg_isready, interval 10s)
- [ ] Application startup includes database connection verification

---

## Cross-References

**Related Documentation**:
- PostgreSQL service requirements: See `.github/instructions/04-01.github.instructions.md`
- Service template health checks: See `.specify/memory/service-template.md`
- Database configuration patterns: See `.specify/memory/database.md`
- Docker Compose secrets: See `.specify/memory/docker.md`
