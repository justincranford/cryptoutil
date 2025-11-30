# Build Pipeline Impact Assessment

## Overview

This document analyzes the impact of the service group refactoring on GitHub Actions workflows, pre-commit hooks, composite actions, and CI/CD infrastructure. It provides a migration checklist and validation plan.

**Cross-references:**

- [Group Directory Blueprint](./blueprint.md) - Defines target package locations
- [Import Alias Policy](./import-aliases.md) - Import alias migration strategy
- [Shared Utilities Extraction](./shared-utilities.md) - Utility package moves

---

## Workflow Inventory

### Current Workflows (10 Total)

| Workflow | File | Service Orchestration | Affected by Refactor |
|----------|------|----------------------|---------------------|
| **Quality** | `ci-quality.yml` | None | ✅ Yes - Build step references `./cmd/cryptoutil` |
| **Coverage** | `ci-coverage.yml` | None | ✅ Yes - Test paths reference `./internal/...` |
| **Benchmark** | `ci-benchmark.yml` | None | ✅ Yes - Benchmark paths reference `./internal/...` |
| **GitLeaks** | `ci-gitleaks.yml` | None | ❌ No - Scans all files, path-agnostic |
| **SAST** | `ci-sast.yml` | None | ❌ No - Scans all Go files, path-agnostic |
| **Race** | `ci-race.yml` | None | ✅ Yes - Test paths reference `./internal/...` |
| **Fuzz** | `ci-fuzz.yml` | None | ✅ Yes - Fuzz test paths reference `./internal/common/crypto/...` |
| **E2E** | `ci-e2e.yml` | Full Docker stack | ✅ Yes - Test paths reference `./internal/test/e2e/` |
| **DAST** | `ci-dast.yml` | PostgreSQL | ✅ Yes - Docker Compose paths, service references |
| **Load** | `ci-load.yml` | Full Docker stack | ✅ Yes - Docker Compose paths, Java Gatling tests |

---

## Composite Actions Inventory (15 Total)

| Action | Purpose | Affected by Refactor |
|--------|---------|---------------------|
| `workflow-job-begin` | Initialize job (timestamp, Go version) | ❌ No - Generic infrastructure |
| `workflow-job-end` | Finalize job (duration summary) | ❌ No - Generic infrastructure |
| `go-setup` | Set up Go environment | ❌ No - Version-based, path-agnostic |
| `golangci-lint` | Run golangci-lint | ✅ Yes - Uses `.golangci.yml` (importas rules) |
| `custom-cicd-lint` | Run CICD validation commands | ✅ Yes - References `internal/cmd/cicd/` |
| `docker-compose-build` | Build Docker images | ✅ Yes - References `./deployments/Dockerfile` |
| `docker-compose-up` | Start Docker Compose services | ✅ Yes - References `./deployments/compose/compose.yml` |
| `docker-compose-down` | Stop Docker Compose services | ❌ No - Generic command |
| `docker-compose-verify` | Health check services | ❌ No - Generic health checks |
| `docker-compose-logs` | Capture service logs | ❌ No - Generic log collection |
| `docker-images-pull` | Pre-pull Docker images | ❌ No - Generic image pulls |
| `fuzz-test` | Run fuzz tests | ✅ Yes - References `./internal/common/crypto/...` |
| `security-scan-gitleaks` | Run GitLeaks | ❌ No - Path-agnostic scanning |
| `security-scan-trivy` | Run Trivy (legacy) | ❌ No - Path-agnostic scanning |
| `security-scan-trivy2` | Run Trivy (current) | ❌ No - Path-agnostic scanning |

---

## Pre-Commit Hooks Analysis

### .pre-commit-config.yaml

**Hook Categories:**

1. **File Checks** (no impact)
   - `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files`
   - Path-agnostic, no updates needed

2. **Linting Hooks** (medium impact)
   - `golangci-lint-repo-mod` - Uses `.golangci.yml` (needs importas updates)
   - **Action:** Update `.golangci.yml` importas rules with new service group aliases

3. **CICD Hooks** (high impact)
   - `cicd-checks-internal` - Runs `go run ./cmd/cicd` with multiple commands
   - **Action:** Verify `internal/cmd/cicd/` path remains stable
   - **Commands:** `go-enforce-test-patterns`, `go-enforce-any`, `all-enforce-utf8`

4. **Dependency Hooks** (no impact)
   - `go-update-direct-dependencies` - Uses `go.mod`
   - Path-agnostic, no updates needed

5. **Security Hooks** (no impact)
   - `detect-secrets` - Path-agnostic scanning
   - No updates needed

6. **Mutation Testing** (medium impact)
   - `gremlins` - Tests `internal/...` packages
   - May need exclusion updates if package structure changes significantly

---

## Path Filter Analysis

### paths-ignore Configuration

**Current Pattern (used in ALL workflows):**

```yaml
paths-ignore:
  - 'docs/**'
  - '**/*.md'
  - '.github/copilot-instructions.md'
  - '.github/instructions/**'
  - 'workflow-reports/**'
  - 'nohup.out'
  - 'LICENSE'
  - '.editorconfig'
  - '.gitignore'
  - '.gitattributes'
  - '.github/ISSUE_TEMPLATE/**'
  - '.github/pull_request_template.md'
  - '.github/dependabot.yml'
  - '**/*.log'
  - '**/*.sarif'
```

**Refactor Impact:**

- **No changes needed** - Filters are file type/directory-based, not package-based
- After docs/ reorganization (Task 9), verify `docs/**` filter still covers all documentation

**Recommendation:**

- Add `test-output/**` to paths-ignore (temporary test artifacts)
- Consider adding `workflow-reports/**` (already present)

---

## Required Workflow Updates by Phase

### Phase 1: Identity Extraction (Task 10)

**Workflows Affected:**

1. **ci-quality.yml**
   - Build step: No change (still `./cmd/cryptoutil`)
   - Linting: Update `.golangci.yml` importas rules

2. **ci-coverage.yml**
   - Test paths: No change (`./internal/...` covers identity)
   - Coverage filtering: No change

3. **ci-e2e.yml**
   - Test paths: No change (`./internal/test/e2e/` stable)
   - Docker Compose: Verify identity service references

**Pre-Commit Hooks:**

- Update `.golangci.yml` importas (add identity service group aliases)
- No cicd command changes needed

**Validation:**

```bash
# Verify workflows pass after identity extraction
go run ./cmd/workflow -workflows=quality,coverage,e2e
```

---

### Phase 2: KMS Extraction (Task 11)

**Workflows Affected:**

1. **ci-quality.yml**
   - Build step: Update `./cmd/cryptoutil` to `./cmd/kms/cryptoutil` (if CLI moves)
   - Linting: Update `.golangci.yml` importas rules (KMS aliases)

2. **ci-coverage.yml**
   - Test paths: Update `./internal/...` to `./internal/kms/...` + `./internal/...`
   - Coverage thresholds: Separate KMS vs common coverage targets

3. **ci-benchmark.yml**
   - Benchmark paths: Update `./internal/server/...` to `./internal/kms/server/...`

4. **ci-fuzz.yml**
   - Fuzz test paths: Update crypto package paths
   - `./internal/common/crypto/keygen` → `./pkg/crypto/keygen`
   - `./internal/common/crypto/digests` → `./pkg/crypto/digests`

5. **ci-e2e.yml**
   - Test infrastructure: Update `internal/test/e2e/` references
   - Docker Compose: Update service names (`cryptoutil-*` → `kms-*`)

6. **ci-dast.yml**
   - Service endpoints: Update container names and URLs
   - Nuclei targets: Update from `cryptoutil-sqlite` to `kms-sqlite`

7. **ci-load.yml**
   - Gatling tests: Update service endpoints
   - Docker Compose: Update service references

**Pre-Commit Hooks:**

- Update `.golangci.yml` importas (KMS service group aliases)
- Update `custom-cicd-lint` if cicd package moves

**Composite Actions:**

- `golangci-lint`: Auto-updated via `.golangci.yml` changes
- `custom-cicd-lint`: Verify `internal/cmd/cicd/` path stable
- `fuzz-test`: Update crypto package paths

**Docker Compose:**

- Update `./deployments/compose/compose.yml` service names
- Update health check endpoints
- Update Swagger UI URLs

**Validation:**

```bash
# Verify all workflows pass after KMS extraction
go run ./cmd/workflow -workflows=all
```

---

### Phase 3: CA Preparation (Task 12)

**Workflows Affected:**

1. **ci-quality.yml**
   - Build step: Add CA CLI build (if implemented)
   - Linting: Update `.golangci.yml` importas rules (CA aliases)

2. **ci-coverage.yml**
   - Test paths: Add `./internal/ca/...` coverage
   - Coverage thresholds: Define CA coverage target

3. **ci-e2e.yml**
   - Add CA service orchestration (if applicable)
   - Update test suite to include CA tests

**Pre-Commit Hooks:**

- Update `.golangci.yml` importas (CA service group aliases)

**Validation:**

```bash
# Verify workflows pass after CA structure added
go run ./cmd/workflow -workflows=quality,coverage
```

---

## Artifact and Caching Strategy

### Current Caching

**Go Module Caching:**

- `go-setup` action uses `actions/setup-go@v6` with `cache: true`
- Cache key: `go.sum` hash + OS
- **Impact:** None - Go module cache is path-agnostic

**Docker Layer Caching:**

- No explicit layer caching currently
- **Opportunity:** Add Docker BuildKit cache to speed up builds post-refactor

**Recommendation:**

```yaml
- name: Set up Docker Buildx with cache
  uses: docker/setup-buildx-action@v4
  with:
    buildkitd-flags: --allow-insecure-entitlement network.host
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Artifact Uploads

**Current Artifacts (by workflow):**

| Workflow | Artifact | Path | Retention | Impact |
|----------|----------|------|-----------|--------|
| quality | `golangci-lint-report` | `workflow-reports/golangci-lint/` | 1 day | None |
| coverage | `coverage-html-report` | `workflow-reports/coverage/coverage.html` | 30 days | None |
| coverage | `coverage-func-report` | `workflow-reports/coverage/coverage-func.txt` | 30 days | None |
| benchmark | `benchmark-results` | `workflow-reports/benchmark/` | 7 days | None |
| race | `race-test-results` | `workflow-reports/race/` | 7 days | None |
| fuzz | `fuzz-test-corpus` | `workflow-reports/fuzz/` | 7 days | Path may change (crypto moves) |
| e2e | `e2e-docker-logs` | `workflow-reports/e2e/docker-logs/` | 3 days | None |
| dast | `nuclei-sarif` | `workflow-reports/dast/nuclei.sarif` | 30 days | None |
| dast | `zap-sarif` | `workflow-reports/dast/zap.sarif` | 30 days | None |
| load | `gatling-results` | `test/load/results/` | 7 days | None |

**Refactor Impact:**

- **Low:** Most artifacts use `workflow-reports/` directory (stable)
- **Medium:** Fuzz test corpus path may change with crypto package moves

**Recommendation:**

- Standardize all artifacts under `workflow-reports/<workflow-name>/`
- Update artifact paths in `ci-fuzz.yml` after crypto promotion

---

## SARIF Upload Strategy

### Current SARIF Uploads

| Workflow | Tool | SARIF File | Security Tab Integration |
|----------|------|------------|-------------------------|
| quality | Trivy (Dockerfile) | `workflow-reports/trivy-dockerfile.sarif` | ✅ Yes |
| quality | Trivy (image) | `workflow-reports/trivy-image.sarif` | ✅ Yes |
| coverage | Trivy (coverage) | `workflow-reports/trivy-coverage.sarif` | ✅ Yes |
| sast | gosec | `workflow-reports/sast/gosec.sarif` | ✅ Yes |
| sast | Trivy (sast) | `workflow-reports/trivy-sast.sarif` | ✅ Yes |
| dast | Nuclei | `workflow-reports/dast/nuclei.sarif` | ✅ Yes |
| dast | ZAP | `workflow-reports/dast/zap.sarif` | ✅ Yes |
| gitleaks | GitLeaks | `workflow-reports/gitleaks/gitleaks.sarif` | ✅ Yes |

**Refactor Impact:**

- **None** - SARIF uploads are path-agnostic
- All uploads use `github/codeql-action/upload-sarif@v3` (stable API)

---

## importas Migration Strategy

### Current .golangci.yml Importas Rules (85 Aliases)

**Categories:**

1. **JOSE Libraries** (4 aliases)
   - `joseJwa`, `joseJwe`, `joseJwk`, `joseJws`
   - **Impact:** None (third-party packages)

2. **Standard Library** (3 aliases)
   - `crand`, `mathrand`, `cryptorand`
   - **Impact:** None

3. **Third-Party** (2 aliases)
   - `googleUuid`, `moderncsqlite`
   - **Impact:** None

4. **cryptoutil API** (3 aliases)
   - `cryptoutilOpenapiClient`, `cryptoutilOpenapiModel`, `cryptoutilOpenapiServer`
   - **Impact:** Low (generated code, stable paths)

5. **Server (KMS)** (12 aliases)
   - `cryptoutilApplication`, `cryptoutilBusinessLogic`, `cryptoutilDomain`, etc.
   - **Impact:** HIGH - All change to `cryptoutilKms*` pattern

6. **Common** (14 aliases)
   - `cryptoutilApperr`, `cryptoutilConfig`, `cryptoutilContainer`, etc.
   - **Impact:** MEDIUM - Some move to `cryptoutilKms*`, some stay

7. **Crypto** (5 aliases)
   - `cryptoutilAsn1`, `cryptoutilCertificate`, `cryptoutilDigests`, etc.
   - **Impact:** HIGH - Most move to `pkg/crypto/*`

8. **Identity** (11 aliases)
   - `cryptoutilIdentityAuthz`, `cryptoutilIdentityIdp`, etc.
   - **Impact:** LOW - Already isolated, may add `cryptoutilIdentity*` prefix

9. **CICD** (12 aliases)
   - `cryptoutilCicd`, `cryptoutilCicdCommon`, etc.
   - **Impact:** None (stable paths)

10. **Stdlib Crypto** (7 aliases)
    - `cryptoAes`, `cryptoEcdh`, `cryptoEcdsa`, etc.
    - **Impact:** None

### Proposed importas Updates (30 New Aliases)

**KMS Service Group** (12 new aliases):

```yaml
# internal/kms/application → cryptoutilKmsApplication
- pkg: "cryptoutil/internal/kms/application"
  alias: "cryptoutilKmsApplication"

# internal/kms/businesslogic → cryptoutilKmsBusinessLogic
- pkg: "cryptoutil/internal/kms/businesslogic"
  alias: "cryptoutilKmsBusinessLogic"

# internal/kms/domain → cryptoutilKmsDomain
- pkg: "cryptoutil/internal/kms/domain"
  alias: "cryptoutilKmsDomain"

# ... (10 more KMS aliases)
```

**Crypto Promotion** (5 new aliases):

```yaml
# pkg/crypto/keygen → cryptoutilKeygen
- pkg: "cryptoutil/pkg/crypto/keygen"
  alias: "cryptoutilKeygen"

# pkg/crypto/digests → cryptoutilDigests
- pkg: "cryptoutil/pkg/crypto/digests"
  alias: "cryptoutilDigests"

# pkg/crypto/asn1 → cryptoutilAsn1
- pkg: "cryptoutil/pkg/crypto/asn1"
  alias: "cryptoutilAsn1"

# pkg/crypto/certificate → cryptoutilCertificate
- pkg: "cryptoutil/pkg/crypto/certificate"
  alias: "cryptoutilCertificate"
```

**KMS Utilities** (4 new aliases):

```yaml
# internal/kms/crypto/jose → cryptoutilKmsJose
- pkg: "cryptoutil/internal/kms/crypto/jose"
  alias: "cryptoutilKmsJose"

# internal/kms/pool → cryptoutilKmsPool
- pkg: "cryptoutil/internal/kms/pool"
  alias: "cryptoutilKmsPool"

# internal/kms/telemetry → cryptoutilKmsTelemetry
- pkg: "cryptoutil/internal/kms/telemetry"
  alias: "cryptoutilKmsTelemetry"

# internal/kms/container → cryptoutilKmsContainer
- pkg: "cryptoutil/internal/kms/container"
  alias: "cryptoutilKmsContainer"
```

**CA Service Group** (9 new aliases):

```yaml
# internal/ca/application → cryptoutilCaApplication
- pkg: "cryptoutil/internal/ca/application"
  alias: "cryptoutilCaApplication"

# ... (9 more CA aliases)
```

**Total:** 85 current + 30 new = 115 aliases (matches import-aliases.md proposal)

---

## Workflow Execution Testing Strategy

### Pre-Refactor Baseline

**Step 1: Capture baseline workflow execution times**

```bash
# Run all workflows and capture durations
go run ./cmd/workflow -workflows=all

# Expected results:
# - quality: 3-5 minutes
# - coverage: 8-10 minutes
# - benchmark: 5-7 minutes
# - gitleaks: 1-2 minutes
# - sast: 2-3 minutes
# - race: 10-12 minutes
# - fuzz: 3-5 minutes
# - e2e: 8-10 minutes
# - dast: 10-15 minutes (quick profile)
# - load: 12-15 minutes
```

**Step 2: Document baseline coverage**

```bash
# Capture coverage before refactoring
go test ./... -coverprofile=test-output/coverage_baseline.out
go tool cover -func=test-output/coverage_baseline.out > test-output/coverage_baseline_func.txt
```

**Step 3: Tag baseline commit**

```bash
git tag refactor-baseline-$(date +%Y%m%d)
git push origin refactor-baseline-$(date +%Y%m%d)
```

---

### During Refactor Validation

**After Each Phase:**

1. **Update .golangci.yml importas rules**
   - Add new service group aliases
   - Run `golangci-lint run ./...` to verify

2. **Update workflow paths** (if needed)
   - Test paths, Docker Compose references, artifact paths

3. **Run affected workflows**
   - Phase 1 (Identity): `quality,coverage,e2e`
   - Phase 2 (KMS): `all`
   - Phase 3 (CA): `quality,coverage`

4. **Compare coverage**

   ```bash
   go test ./... -coverprofile=test-output/coverage_post_phaseN.out
   go tool cover -func=test-output/coverage_post_phaseN.out > test-output/coverage_post_phaseN_func.txt
   diff test-output/coverage_baseline_func.txt test-output/coverage_post_phaseN_func.txt
   ```

5. **Check workflow durations**
   - Ensure no significant performance regression (±10% acceptable)

---

### Post-Refactor Final Validation

**Step 1: Run full workflow suite**

```bash
# Full workflow execution (30-45 minutes total)
go run ./cmd/workflow -workflows=all

# Verify all workflows pass
```

**Step 2: Coverage comparison**

```bash
# Compare final coverage to baseline
go test ./... -coverprofile=test-output/coverage_final.out
go tool cover -func=test-output/coverage_final.out > test-output/coverage_final_func.txt

# Diff analysis
diff test-output/coverage_baseline_func.txt test-output/coverage_final_func.txt

# Acceptable: ±2% variance
# Investigate: >5% coverage drop in any package
```

**Step 3: Integration testing**

```bash
# Run E2E tests
go run ./cmd/workflow -workflows=e2e

# Run DAST tests (quick profile)
go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"

# Run load tests
go run ./cmd/workflow -workflows=load
```

**Step 4: Tag final commit**

```bash
git tag refactor-complete-$(date +%Y%m%d)
git push origin refactor-complete-$(date +%Y%m%d)
```

---

## Migration Checklist

### Phase 1: Identity Extraction

- [ ] Update `.golangci.yml` importas (add identity aliases)
- [ ] Run `golangci-lint run ./...` - verify no importas errors
- [ ] Run `go run ./cmd/workflow -workflows=quality,coverage,e2e`
- [ ] Compare coverage: baseline vs post-identity
- [ ] Commit: `refactor: extract identity service group`

### Phase 2: KMS Extraction

- [ ] Update `.golangci.yml` importas (add KMS aliases, update common aliases)
- [ ] Update workflow test paths (`./internal/...` → `./internal/kms/...`)
- [ ] Update Docker Compose service names (`cryptoutil-*` → `kms-*`)
- [ ] Update DAST endpoints (Nuclei, ZAP targets)
- [ ] Update fuzz test paths (crypto package moves)
- [ ] Run `golangci-lint run ./...` - verify no importas errors
- [ ] Run `go run ./cmd/workflow -workflows=all`
- [ ] Compare coverage: baseline vs post-KMS
- [ ] Commit: `refactor: extract KMS service group`

### Phase 3: CA Preparation

- [ ] Update `.golangci.yml` importas (add CA aliases)
- [ ] Update workflow test paths (add `./internal/ca/...`)
- [ ] Run `golangci-lint run ./...` - verify no importas errors
- [ ] Run `go run ./cmd/workflow -workflows=quality,coverage`
- [ ] Compare coverage: baseline vs post-CA
- [ ] Commit: `refactor: prepare CA service group structure`

### Post-Refactor Cleanup

- [ ] Remove compatibility shims (after 8-week grace period)
- [ ] Update `.golangci.yml` importas (remove legacy aliases)
- [ ] Run `go run ./cmd/workflow -workflows=all` - final validation
- [ ] Update documentation (README, import-aliases.md, blueprint.md)
- [ ] Tag release: `git tag refactor-complete-$(date +%Y%m%d)`

---

## Cross-References

- **Group Directory Blueprint:** [docs/01-refactor/blueprint.md](./blueprint.md)
- **Import Alias Policy:** [docs/01-refactor/import-aliases.md](./import-aliases.md)
- **Shared Utilities Extraction:** [docs/01-refactor/shared-utilities.md](./shared-utilities.md)
- **CLI Strategy Framework:** [docs/01-refactor/cli-strategy.md](./cli-strategy.md)

---

## Notes

- **Workflow paths are mostly stable** - `./internal/...` pattern covers refactored packages
- **importas updates are CRITICAL** - Workflows will fail linting without updated aliases
- **Docker Compose service renaming is optional** - Can keep `cryptoutil-*` names for backward compatibility
- **Artifact paths are stable** - Most use `workflow-reports/` directory (no changes needed)
- **SARIF uploads are path-agnostic** - Security tab integration unchanged
- **Fuzz test paths may change** - Update after crypto promotion to `pkg/crypto/`
- **Coverage tracking is essential** - Validate no regression after each phase
