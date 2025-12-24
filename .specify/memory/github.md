# CI/CD Workflow Configuration - Complete Specifications

**Referenced by**: `.github/instructions/04-01.github.instructions.md`

## Cost Efficiency Patterns

**Path Filters**: Trigger workflows only on relevant changes, exclude docs
**Matrix Optimization**: Skip jobs for docs-only changes, conditional steps
**Caching**: Use `cache: true` on `actions/setup-go@v6` (NEVER manual cache actions)

## PostgreSQL Service Configuration - CRITICAL

**Pattern 1: Unit/Integration** (Recommended) - test-containers with randomized credentials
**Pattern 2: E2E Tests** - Docker Compose with Docker secrets (`file:///run/secrets/`)
**Pattern 3: GitHub Workflows** (Legacy) - Use ONLY when test-containers not feasible

**Rationale**: Env vars insecure, test-containers provide isolation/randomization, Docker secrets enforce secure patterns

## Workflow Matrix Reference

| Category | Workflows | Purpose |
|----------|-----------|---------|  
| **CI** | ci-quality, ci-test, ci-coverage, ci-benchmark, ci-mutation, ci-race | Linting, tests, coverage, benchmarks |
| **Security** | ci-sast, ci-gitleaks, ci-dast | Static/dynamic security analysis |
| **Integration** | ci-e2e, ci-load | E2E tests, load testing |

## Configuration Management Patterns

**Production/CI**: ALWAYS use config files (YAML), NEVER env vars
**Tests**: SQLite in-memory (`--dev`) OR test-containers for PostgreSQL
**Secrets**: ALWAYS Docker secrets (`file:///run/secrets/`)

## Artifact Management

**Test Artifacts**: `if: always()`, upload even on failure
**Retention**: Temp logs (1 day), coverage (7 days), security/benchmarks (30 days)

## Act Workflow Testing (Local Execution)

**cmd/workflow Utility**: `go run ./cmd/workflow -workflows=dast,e2e -inputs="key=value"`

**Rules**: NEVER use `-t` timeout, ALWAYS specify `-workflows`, use `-inputs` for params
**Why**: Test workflows locally before push, debug without CI/CD, faster iteration

---

## Diagnostic Logging Standards

**Steps >10s MUST include timing**: Start time, end time, duration in seconds, emoji indicators (📋 start, ⏱️ duration, ✅ success)

---

## Variable Expansion in Heredocs - CRITICAL

**ALWAYS use `${VAR}`** (curly braces) for variable expansion in Bash heredocs

**Rules**: ✅ `${VAR}`, verify expansion, include `cat config.yml` ❌ `$VAR`, skip verification

**Historical Mistake**: ci-dast `$POSTGRES_USER` → literal "$POSTGRES_USER" → "role 'root' does not exist"

---

## Debugging Workflow PostgreSQL Issues

**Checklist**: Download artifacts, check PostgreSQL logs ("role '<username>' does not exist"), verify credentials match, check config for literal $VAR, compare with working workflows

**Evidence Pattern**: 27+ errors (persistent) + 10s interval (health retry) + PostgreSQL OK (credentials issue) + literal $VAR (expansion issue)

**Preventive**: PostgreSQL creds cryptoutil_test/cryptoutil/cryptoutil_test_password, env vars match, `${VAR}` syntax, verification step, health check configured
