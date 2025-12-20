# Workflow Fixes - 2025-12-20

## Summary

4 workflows failed from commit `6b1a6174` (docs(spec): add all 9 services architecture + create CLARIFY-QUIZME.md):

1. **CI - Quality Testing** (BLOCKING - halts deployments)
2. **CI - End-to-End Testing**  
3. **CI - Load Testing**
4. **CI - DAST Security Testing**

## Task List

### Task 1: Fix Outdated Go Dependencies (CI - Quality Testing)

**Workflow**: `.github/workflows/ci-quality.yml`  
**Failure**: `lint-go-mod` detected outdated dependencies  
**Root Cause**: Two direct dependencies have newer versions available:

- `github.com/goccy/go-yaml` v1.19.0 → v1.19.1
- `modernc.org/sqlite` v1.40.1 → v1.41.0

**Proposed Fix**:

```bash
# Update dependencies
go get github.com/goccy/go-yaml@v1.19.1
go get modernc.org/sqlite@v1.41.0
go mod tidy

# Verify fix
go run ./cmd/cicd lint-go-mod

# Run tests to ensure no breaking changes
go test ./...
```

**Priority**: CRITICAL (blocks all deployments)  
**Estimated Time**: 5 minutes

---

### Task 2: Fix Identity AuthZ Service Startup (CI - E2E Testing, Load Testing, DAST)

**Workflows**:

- `.github/workflows/ci-e2e.yml`
- `.github/workflows/ci-load.yml`
- `.github/workflows/ci-dast.yml`

**Failure**: `compose-identity-authz-e2e-1` container exits with code 1  
**Error**: `dependency failed to start: container compose-identity-authz-e2e-1 exited (1)`

**Root Cause Analysis**:

From workflow logs:

```
Container compose-identity-authz-e2e-1  Error
dependency failed to start: container compose-identity-authz-e2e-1 exited (1)
```

From local Docker logs attempt:

```
failed to connect to the docker API at npipe:////./pipe/dockerDesktopLinuxEngine
open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified.
```

**Hypothesis**: The `identity-authz-e2e` service is failing to start in GitHub Actions CI environment, but the root cause is not clear from logs. Need to:

1. Check `deployments/compose/compose.yml` for identity-authz-e2e configuration issues
2. Check `deployments/identity/config/authz-e2e.yml` for config file issues
3. Verify command arguments: `["identity", "start", "--service=authz", "--config=/app/config/authz-e2e.yml", "-u", "file:///run/secrets/postgres_url.secret"]`
4. Check healthcheck endpoint: `https://127.0.0.1:8080/health` (should this be `/admin/v1/healthz` or `/admin/v1/livez`?)
5. Verify database connection (depends on `identity-postgres-e2e` being healthy)

**Debugging Steps**:

1. Read identity-authz-e2e service definition in compose.yml (lines 669-730)
2. Read authz-e2e.yml config file
3. Check if healthcheck endpoint is correct
4. Compare with working KMS service healthcheck patterns
5. Check if command arguments are correct for identity unified binary
6. Verify postgres_url.secret is being passed correctly

**Proposed Fix**: (TBD after investigation)

**Priority**: HIGH (blocks E2E, Load, and DAST workflows)  
**Estimated Time**: 30-60 minutes

---

### Task 3: Fix DAST Application Readiness Timeout

**Workflow**: `.github/workflows/ci-dast.yml`  
**Failure**: Application fails to become ready within timeout (30 attempts × 5s backoff = 150 seconds)  
**Error**:

```
Attempt 30/30 (backoff: 5s)
Testing: https://127.0.0.1:9090/admin/v1/readyz
❌ Not ready: https://127.0.0.1:9090/admin/v1/readyz
❌ Application failed to become ready within timeout
```

**Root Cause**: This is a **symptom** of Task 2 (identity-authz-e2e failing to start). The DAST workflow is waiting for cryptoutil services to become healthy, but they can't start because identity-authz-e2e dependency failed.

**Proposed Fix**: Fix Task 2 first - this should automatically resolve DAST timeout.

**Alternative**: If Task 2 fix doesn't resolve DAST, investigate:

1. Check if `/admin/v1/readyz` endpoint is correct (should it be `/admin/v1/livez`?)
2. Increase timeout attempts (30 → 60) or increase backoff (5s → 10s)
3. Check if DAST is testing against the wrong service (should test cryptoutil-sqlite, not identity-authz)

**Priority**: MEDIUM (likely resolves with Task 2 fix)  
**Estimated Time**: 5 minutes (if Task 2 resolves it) or 15-30 minutes (if separate issue)

---

## Execution Plan

1. **Fix Task 1 immediately** (outdated dependencies) - commit and push
2. **Wait for Quality workflow to pass** before proceeding
3. **Investigate Task 2** (identity-authz startup failure) - read configs, compare patterns
4. **Apply Task 2 fix** - commit and push
5. **Wait for E2E, Load, and DAST workflows** to verify Task 2 fix resolves all three
6. **If DAST still fails**: Apply Task 3 alternative fixes
7. **Re-run workflow monitoring cycle** until all workflows pass

---

## Next Steps

Once all workflows pass:

1. User answers CLARIFY-QUIZME.md questions
2. Re-run `/speckit.clarify` with user answers
3. Re-run `/speckit.analyze` to validate constitution, spec, clarify, and copilot instructions
4. Update incomplete spec-kit documents (PLAN-incomplete.md → PLAN.md, etc.)
