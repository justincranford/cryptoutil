# CI/CD Workflow Path Filter Updates Plan

## Executive Summary

Update 10 GitHub Actions workflows and 5 composite actions to reflect new directory structure after identity/KMS/CA service group refactoring.

**Status**: Planning
**Dependencies**: Tasks 10-15 (service extractions, CLI restructure complete)
**Risk Level**: Medium (workflow configuration changes, CI/CD disruption possible)

## Current Workflow Inventory

From [Pipeline Impact Assessment](pipeline-impact.md):

### 10 GitHub Actions Workflows

| Workflow | File | Affected | Update Required |
|----------|------|----------|----------------|
| Quality | `ci-quality.yml` | ✅ Yes | Build paths, importas config |
| Coverage | `ci-coverage.yml` | ✅ Yes | Test paths |
| Benchmark | `ci-benchmark.yml` | ✅ Yes | Benchmark paths |
| GitLeaks | `ci-gitleaks.yml` | ❌ No | Path-agnostic |
| SAST | `ci-sast.yml` | ❌ No | Path-agnostic |
| Race | `ci-race.yml` | ✅ Yes | Test paths |
| Fuzz | `ci-fuzz.yml` | ✅ Yes | Crypto package paths |
| E2E | `ci-e2e.yml` | ✅ Yes | Test infrastructure, Docker Compose |
| DAST | `ci-dast.yml` | ✅ Yes | Service endpoints, Docker Compose |
| Load | `ci-load.yml` | ✅ Yes | Service endpoints, Gatling tests |

**Summary**: 7 workflows require updates, 3 are path-agnostic

### 15 Composite Actions

| Action | Affected | Update Required |
|--------|----------|----------------|
| `workflow-job-begin` | ❌ No | Generic infrastructure |
| `workflow-job-end` | ❌ No | Generic infrastructure |
| `go-setup` | ❌ No | Version-based |
| `golangci-lint` | ✅ Yes | `.golangci.yml` importas rules |
| `custom-cicd-lint` | ✅ Yes | `internal/cmd/cicd/` references |
| `docker-compose-build` | ✅ Yes | Dockerfile paths |
| `docker-compose-up` | ✅ Yes | `compose.yml` service names |
| `docker-compose-down` | ❌ No | Generic command |
| `docker-compose-verify` | ❌ No | Generic health checks |
| `docker-compose-logs` | ❌ No | Generic log collection |
| `docker-images-pull` | ❌ No | Generic image pulls |
| `fuzz-test` | ✅ Yes | Crypto package paths |
| `security-scan-gitleaks` | ❌ No | Path-agnostic |
| `security-scan-trivy` | ❌ No | Path-agnostic |
| `security-scan-trivy2` | ❌ No | Path-agnostic |

**Summary**: 5 actions require updates, 10 are path-agnostic

## Update Strategy by Refactor Phase

### Phase 1: Identity Extraction (Task 10 - Go Workspace)

**Directory Changes**:

- Identity module: `cryptoutil/internal/identity` → `identity/` (sibling workspace)
- KMS module: `cryptoutil/internal/server`, `internal/client`, `internal/common` → remains in `cryptoutil/` workspace

**Workflow Updates**:

1. **ci-quality.yml** (Build step)
   - **Current**: `go build ./cmd/cryptoutil`
   - **New**: `go build ./cmd/cryptoutil` (unchanged - still in cryptoutil workspace)
   - **Linting**: Update `.golangci.yml` importas rules (add identity workspace aliases)

2. **ci-coverage.yml** (Test paths)
   - **Current**: `go test ./internal/... -cover`
   - **New**: `go test ./... -cover` (covers both cryptoutil and identity workspaces via go.work)
   - **Coverage filtering**: Separate cryptoutil vs identity coverage reports

3. **ci-e2e.yml** (Test infrastructure)
   - **Current**: `go test ./internal/test/e2e/...`
   - **New**: `go test ./internal/test/e2e/...` (unchanged)
   - **Docker Compose**: Verify identity service references (if identity servers start)

**Composite Actions**:

- `golangci-lint`: Update `.golangci.yml` importas (add `cryptoutilIdentity*` aliases)
- No other action changes needed (identity workspace transparent to CI)

**Validation**:

```bash
# Verify workflows pass after identity extraction
go run ./cmd/workflow -workflows=quality,coverage,e2e -inputs="scan_profile=quick"
```

### Phase 2: KMS Extraction (Task 11 - Rename server→kms)

**Directory Changes**:

- KMS server: `internal/server` → `internal/kms`
- KMS client: `internal/client` → `internal/kms/client`
- Utilities: `internal/common/{pool,container}` → `internal/kms/{pool,container}`

**Workflow Updates**:

1. **ci-quality.yml** (Build step)
   - **Build**: No change (`./cmd/cryptoutil` unchanged)
   - **Linting**: Update `.golangci.yml` importas (`cryptoutilServer*` → `cryptoutilKms*`)

2. **ci-coverage.yml** (Test paths)
   - **Current**: `go test ./internal/... -cover`
   - **New**: `go test ./internal/... -cover` (unchanged - `./internal/...` covers kms)
   - **Coverage thresholds**: Separate KMS vs common targets (if needed)

3. **ci-benchmark.yml** (Benchmark paths)
   - **Current**: `go test -bench=. ./internal/server/...`
   - **New**: `go test -bench=. ./internal/kms/...`

4. **ci-fuzz.yml** (Crypto package paths)
   - **Current**: `go test -fuzz=. ./internal/common/crypto/keygen`
   - **New**: `go test -fuzz=. ./internal/common/crypto/keygen` (unchanged - crypto stays in common)

5. **ci-e2e.yml** (Docker Compose service names)
   - **Current**: Uses `cryptoutil-sqlite`, `cryptoutil-postgres-1`, `cryptoutil-postgres-2`
   - **New**: Keep service names unchanged (Docker service names != internal package names)
   - **Config files**: Update `configs/kms/*.yml` references (was `configs/production/*.yml`)

6. **ci-dast.yml** (Service endpoints)
   - **Current**: `https://127.0.0.1:8080` (cryptoutil-sqlite)
   - **New**: Unchanged (service endpoints independent of internal packages)

7. **ci-load.yml** (Gatling tests)
   - **Current**: `test/load/src/main/scala/` references `/browser/api/v1`
   - **New**: Unchanged (API endpoints independent of internal packages)

**Composite Actions**:

- `golangci-lint`: Update `.golangci.yml` importas (KMS aliases)
- `custom-cicd-lint`: Verify `internal/cmd/cicd/` path stable (unchanged)
- `fuzz-test`: No changes (crypto package paths unchanged)

**Docker Compose**:

- **Service names**: Keep `cryptoutil-sqlite`, `cryptoutil-postgres-1`, `cryptoutil-postgres-2` (no change)
- **Config files**: Update volume mounts from `configs/production/` to `configs/kms/`
- **Health checks**: Unchanged (endpoints same)

**Validation**:

```bash
# Verify all workflows pass after KMS extraction
go run ./cmd/workflow -workflows=all -inputs="scan_profile=quick"
```

### Phase 3: CA Preparation (Task 12 - Skeleton Structure)

**Directory Changes**:

- CA structure: Add `internal/ca/{domain,repository,service,config,magic}`
- No existing code moves (skeleton only)

**Workflow Updates**:

1. **ci-quality.yml** (Build step)
   - **Build**: Add CA CLI build (if `cmd/ca/` created)
   - **Linting**: Update `.golangci.yml` importas (add CA aliases)

2. **ci-coverage.yml** (Test paths)
   - **Current**: `go test ./internal/... -cover`
   - **New**: Unchanged (covers CA automatically)
   - **Coverage thresholds**: Define CA coverage target (future)

**Composite Actions**:

- `golangci-lint`: Update `.golangci.yml` importas (CA aliases)

**Validation**:

```bash
# Verify workflows pass after CA structure added
go run ./cmd/workflow -workflows=quality,coverage -inputs="scan_profile=quick"
```

### Phase 4: CLI Restructure (Tasks 13-15)

**Directory Changes**:

- CLI commands: `internal/cmd/cryptoutil/server.go` → `internal/cmd/cryptoutil/kms/server/server.go`
- CLI commands: Add `internal/cmd/cryptoutil/identity/`, `internal/cmd/cryptoutil/ca/`
- Main dispatcher: Update `internal/cmd/cryptoutil/cryptoutil.go`

**Workflow Updates**:

1. **ci-quality.yml** (Build step)
   - **Current**: `go build ./cmd/cryptoutil`
   - **New**: Unchanged (main entry point `cmd/cryptoutil/main.go` stable)

2. **ci-e2e.yml** (Server commands)
   - **Current**: `./kms cryptoutil server start --dev`
   - **New**: `./cryptoutil kms server start --dev` (legacy alias still works)
   - **Recommendation**: Use new commands to avoid deprecation warnings in CI logs

3. **ci-dast.yml** (Server commands)
   - **Current**: `docker compose exec cryptoutil-sqlite ./kms cryptoutil server start`
   - **New**: `docker compose exec cryptoutil-sqlite ./cryptoutil kms server start`

**Docker Entrypoint**:

- **Dockerfile**: Update `ENTRYPOINT ["cryptoutil", "kms", "server", "start"]`
- **compose.yml**: Update `command:` directives to use `kms server start` instead of `server start`

**Validation**:

```bash
# Verify CLI restructure doesn't break workflows
go run ./cmd/workflow -workflows=e2e,dast -inputs="scan_profile=quick"
```

## Implementation Phases

### Phase 1: Create Workflow Update Checklist

**Generate automated checklist**:

```go
// internal/cmd/cicd/workflow_migrate/migrate.go

package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    // Scan .github/workflows/*.yml for patterns requiring updates
    workflows := []string{
        "ci-quality.yml",
        "ci-coverage.yml",
        "ci-benchmark.yml",
        "ci-race.yml",
        "ci-fuzz.yml",
        "ci-e2e.yml",
        "ci-dast.yml",
        "ci-load.yml",
    }

    for _, workflow := range workflows {
        path := filepath.Join(".github", "workflows", workflow)
        content, _ := os.ReadFile(path)

        // Check for patterns needing updates
        checkPatterns(path, string(content), []string{
            "internal/server",            // → internal/kms
            "internal/client",            // → internal/kms/client
            "kms cryptoutil server",          // → cryptoutil kms server
            "configs/production",         // → configs/kms
            "cryptoutilServer",          // → cryptoutilKms (importas)
        })
    }
}

func checkPatterns(file, content string, patterns []string) {
    for _, pattern := range patterns {
        if strings.Contains(content, pattern) {
            fmt.Printf("[%s] Found pattern: %s\n", file, pattern)
        }
    }
}
```

**Run checklist generator**:

```bash
go run ./internal/cmd/cicd/workflow_migrate/migrate.go > workflow-migration-checklist.txt
```

### Phase 2: Update .golangci.yml Importas Rules

**Current importas (85 aliases)** → **New importas (115 aliases)**

**Additions** (30 new aliases):

```yaml
# .golangci.yml

linters-settings:
  importas:
    alias:
      # KMS Service Group (15 new aliases)
      - pkg: cryptoutil/internal/kms
        alias: cryptoutilKms
      - pkg: cryptoutil/internal/kms/server/application
        alias: cryptoutilKmsApplication
      - pkg: cryptoutil/internal/kms/server/barrier
        alias: cryptoutilKmsBarrier
      - pkg: cryptoutil/internal/kms/businesslogic
        alias: cryptoutilKmsBusinessLogic
      - pkg: cryptoutil/internal/kms/client
        alias: cryptoutilKmsClient
      - pkg: cryptoutil/internal/kms/container
        alias: cryptoutilKmsContainer
      - pkg: cryptoutil/internal/kms/domain
        alias: cryptoutilKmsDomain
      - pkg: cryptoutil/internal/kms/handler
        alias: cryptoutilKmsHandler
      - pkg: cryptoutil/internal/kms/pool
        alias: cryptoutilKmsPool
      - pkg: cryptoutil/internal/kms/server/repository
        alias: cryptoutilKmsRepository
      - pkg: cryptoutil/internal/kms/server/repository/orm
        alias: cryptoutilKmsOrm
      - pkg: cryptoutil/internal/kms/server/repository/sqlrepository
        alias: cryptoutilKmsSqlRepository
      - pkg: cryptoutil/internal/kms/server
        alias: cryptoutilKmsServer
      - pkg: cryptoutil/internal/kms/service
        alias: cryptoutilKmsService
      - pkg: cryptoutil/internal/kms/service/keygen
        alias: cryptoutilKmsKeygenService

      # Identity Service Group (10 new aliases - if workspace extraction)
      - pkg: identity/authz
        alias: cryptoutilIdentityAuthz
      - pkg: identity/authz/domain
        alias: cryptoutilIdentityAuthzDomain
      - pkg: identity/authz/repository
        alias: cryptoutilIdentityAuthzRepository
      - pkg: identity/authz/server
        alias: cryptoutilIdentityAuthzServer
      - pkg: identity/idp
        alias: cryptoutilIdentityIdp
      - pkg: identity/idp/domain
        alias: cryptoutilIdentityIdpDomain
      - pkg: identity/idp/repository
        alias: cryptoutilIdentityIdpRepository
      - pkg: identity/idp/server
        alias: cryptoutilIdentityIdpServer
      - pkg: identity/rs
        alias: cryptoutilIdentityRs
      - pkg: identity/spa-rp
        alias: cryptoutilIdentitySpaRp

      # CA Service Group (5 new aliases - skeleton)
      - pkg: cryptoutil/internal/ca
        alias: cryptoutilCA
      - pkg: cryptoutil/internal/ca/domain
        alias: cryptoutilCADomain
      - pkg: cryptoutil/internal/ca/repository
        alias: cryptoutilCARepository
      - pkg: cryptoutil/internal/ca/service
        alias: cryptoutilCAService
      - pkg: cryptoutil/internal/ca/config
        alias: cryptoutilCAConfig

      # Remove old server aliases (12 removals)
      # - pkg: cryptoutil/internal/server  # REMOVED
      # - pkg: cryptoutil/internal/server/application  # REMOVED
      # ... etc (see pipeline-impact.md for full list)
```

**Validation**:

```bash
# Run golangci-lint to verify importas rules
golangci-lint run ./... --disable-all --enable=importas
```

### Phase 3: Update Docker Compose Configuration

**Current `deployments/compose/compose.yml` service config**:

```yaml
services:
  cryptoutil-sqlite:
    image: cryptoutil:latest
    command: ["kms", "server", "start", "--config", "/app/configs/cryptoutil-sqlite.yml"]  # NEW
    volumes:
      - ./cryptoutil/configs/cryptoutil-common.yml:/app/configs/cryptoutil-common.yml:ro
      - ./cryptoutil/configs/cryptoutil-sqlite.yml:/app/configs/cryptoutil-sqlite.yml:ro
```

**Updates**:

1. **Change command from `server` to `kms server`**:

   ```yaml
   # OLD
   command: ["server", "start", "--config", "/app/configs/cryptoutil-sqlite.yml"]

   # NEW
   command: ["kms", "server", "start", "--config", "/app/configs/cryptoutil-sqlite.yml"]
   ```

2. **Update config volume paths** (if configs/ restructured):

   ```yaml
   # If configs/production/ → configs/kms/
   volumes:
     - ./cryptoutil/configs/kms/cryptoutil-common.yml:/app/configs/cryptoutil-common.yml:ro
     - ./cryptoutil/configs/kms/cryptoutil-sqlite.yml:/app/configs/cryptoutil-sqlite.yml:ro
   ```

3. **Update Dockerfile `ENTRYPOINT`**:

   ```dockerfile
   # OLD
   ENTRYPOINT ["cryptoutil", "server", "start"]

   # NEW
   ENTRYPOINT ["cryptoutil", "kms", "server", "start"]
   ```

**Validation**:

```bash
# Verify Docker Compose services start correctly
docker compose -f deployments/compose/compose.yml up -d
docker compose -f deployments/compose/compose.yml ps
docker compose -f deployments/compose/compose.yml logs cryptoutil-sqlite
```

### Phase 4: Update Workflow Files

**ci-quality.yml** (Build and linting):

```yaml
# No changes needed for build step
- name: Build application
  run: go build -v ./cmd/cryptoutil

# Linting uses updated .golangci.yml (importas rules)
- name: Run golangci-lint
  uses: ./.github/actions/golangci-lint
```

**ci-benchmark.yml** (Benchmark paths):

```yaml
# OLD
- name: Run benchmarks
  run: go test -bench=. ./internal/server/...

# NEW
- name: Run benchmarks
  run: go test -bench=. ./internal/kms/...
```

**ci-e2e.yml** (Server commands):

```yaml
# Update server start command
- name: Run E2E tests
  run: |
    go test ./internal/test/e2e/... -v -timeout=20m
    # E2E test infrastructure already uses correct paths
```

**ci-dast.yml** (Service endpoints):

```yaml
# No changes to endpoints (service names unchanged)
- name: Run Nuclei scans
  run: |
    nuclei -target https://127.0.0.1:8080/ -severity medium,high,critical
```

**ci-load.yml** (Gatling tests):

```yaml
# No changes to Gatling tests (API endpoints unchanged)
- name: Run load tests
  run: |
    cd test/load
    ./mvnw gatling:test
```

### Phase 5: Update Composite Actions

**golangci-lint action** (`.github/actions/golangci-lint/action.yml`):

```yaml
# No changes needed - uses .golangci.yml from repo root
- name: Run golangci-lint
  run: golangci-lint run --config .golangci.yml
```

**custom-cicd-lint action** (`.github/actions/custom-cicd-lint/action.yml`):

```yaml
# Verify internal/cmd/cicd/ path still valid
- name: Run CICD checks
  run: |
    go run ./cmd/cicd go-enforce-test-patterns
    go run ./cmd/cicd go-enforce-any
    go run ./cmd/cicd all-enforce-utf8
```

**fuzz-test action** (`.github/actions/fuzz-test/action.yml`):

```yaml
# No changes - crypto package paths unchanged
- name: Run fuzz tests
  run: |
    go test -fuzz=. -fuzztime=15s ./internal/common/crypto/keygen
    go test -fuzz=. -fuzztime=15s ./internal/common/crypto/digests
```

**docker-compose-build action** (`.github/actions/docker-compose-build/action.yml`):

```yaml
# Update Dockerfile path if moved
- name: Build Docker images
  run: |
    docker compose -f deployments/compose/compose.yml build
```

**docker-compose-up action** (`.github/actions/docker-compose-up/action.yml`):

```yaml
# No changes - service names unchanged
- name: Start services
  run: |
    docker compose -f deployments/compose/compose.yml up -d
```

### Phase 6: Testing & Validation

**Run full workflow suite**:

```bash
# Test all workflows locally via act
go run ./cmd/workflow -workflows=all -inputs="scan_profile=quick"
```

**Validation checklist**:

- [ ] `ci-quality.yml` passes (build succeeds, linting passes with new importas)
- [ ] `ci-coverage.yml` passes (tests run, coverage reported)
- [ ] `ci-benchmark.yml` passes (benchmarks run with new paths)
- [ ] `ci-race.yml` passes (race tests run)
- [ ] `ci-fuzz.yml` passes (fuzz tests run)
- [ ] `ci-e2e.yml` passes (services start, tests run)
- [ ] `ci-dast.yml` passes (Nuclei/ZAP scans succeed)
- [ ] `ci-load.yml` passes (Gatling tests run)
- [ ] Docker Compose services start correctly
- [ ] No deprecation warnings in CI logs

### Phase 7: Documentation Updates

**Update README.md**:

```markdown
## CI/CD Workflows

After refactoring:
- All workflows use new `kms`, `identity`, `ca` directory structure
- Docker Compose services use `cryptoutil kms server start` command
- `.golangci.yml` importas rules updated for new service groups
```

**Update workflow documentation**:

```bash
# Update WORKFLOWS.md with new paths
# Update docs/pre-commit-hooks.md with new importas rules
```

## Risk Assessment

### Medium Risks

1. **Workflow Configuration Errors**
   - Mitigation: Test all workflows locally via `act` before pushing
   - Rollback: Revert workflow file changes

2. **Docker Compose Service Startup Failures**
   - Mitigation: Test Docker Compose locally after each change
   - Rollback: Revert `compose.yml` and Dockerfile changes

3. **Importas Linting Failures**
   - Mitigation: Run `golangci-lint run --disable-all --enable=importas` locally
   - Rollback: Revert `.golangci.yml` changes

### Low Risks

1. **Path-Agnostic Workflows**
   - No risk: GitLeaks, SAST, Trivy scans are path-independent

2. **Artifact Uploads**
   - Low risk: Artifacts use `workflow-reports/` (stable directory)

## Success Metrics

- [ ] All 10 workflows pass in GitHub Actions
- [ ] Docker Compose services start successfully
- [ ] No linting errors with new importas rules
- [ ] No deprecation warnings in CI logs
- [ ] E2E tests pass with new service structure
- [ ] DAST scans succeed with new endpoints
- [ ] Load tests pass with new configuration

## Timeline

- **Phase 1**: Create workflow update checklist (1 hour)
- **Phase 2**: Update `.golangci.yml` importas rules (2 hours)
- **Phase 3**: Update Docker Compose configuration (1 hour)
- **Phase 4**: Update workflow files (2 hours)
- **Phase 5**: Update composite actions (1 hour)
- **Phase 6**: Testing & validation (3 hours)
- **Phase 7**: Documentation updates (1 hour)

**Total**: 11 hours (1.5 days)

## Cross-References

- [Pipeline Impact Assessment](pipeline-impact.md) - Workflow inventory and analysis
- [KMS Extraction](kms-extraction.md) - Package rename details
- [CLI Restructure](cli-restructure.md) - Command structure changes
- [Importas Migration](importas-migration.md) - Import alias updates (Task 17)

## Next Steps

After workflow updates:

1. **Task 17**: Importas migration (update all import statements)
2. **Task 18**: Observability updates (OTLP service names)
3. **Task 19-20**: Integration testing, documentation finalization
