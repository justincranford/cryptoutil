# Framework v13: Deterministic E2E Testing for All 16 Docker Compose Deployments

## Executive Summary

The current E2E testing infrastructure has critical gaps that prevent consistent, deterministic, and reproducible validation of all 16 Docker Compose deployments (1 suite, 5 product, 10 PS-ID). Only 4 of 10 PS-ID-level deployments have compose-based E2E tests. Zero product-level or suite-level deployments are tested. Tests are copy-pasted across services with massive code duplication, use InsecureSkipVerify instead of validating the actual TLS certificate chain, skip the sqlite-2 instance entirely, and cannot be orchestrated as a single deterministic test suite.

This plan defines a comprehensive, phased approach to deliver Go-based E2E tests that orchestrate and validate all 16 `docker compose` deployments using reusable framework code. The result must be deterministic and reproducible — not dependent on LLM agent non-deterministic behavior from chat session to chat session.

---

## Flaws Identified (Deep Analysis)

### F1. Missing PS-ID E2E Tests (5 of 10 PS-IDs untested)

**Current state**: Only 4 PS-IDs have compose-based E2E tests: `sm-kms`, `sm-im`, `jose-ja`, `skeleton-template`.

**Missing compose-based E2E**: `pki-ca`, `identity-idp`, `identity-rs`, `identity-rp`, `identity-spa`.

**`identity-authz`**: Has `rotation_test.go` in `e2e/` package, but this is NOT a compose E2E test — it creates its own in-memory SQLite DB and does not use Docker Compose at all. It must be reclassified as an integration test and a proper compose-based E2E test must be added.

### F2. Zero Product-Level E2E Tests (0 of 5 products tested)

No tests validate the 5 product-level compose files (`deployments/sm/`, `deployments/jose/`, `deployments/pki/`, `deployments/identity/`, `deployments/skeleton/`). These use recursive includes with `!override` port substitution (SERVICE + 10000), which is a different code path from PS-ID-level compose files. Port conflicts, include resolution, and secret path resolution are untested at this tier.

### F3. Zero Suite-Level E2E Tests (0 of 1 suite tested)

The suite compose file (`deployments/cryptoutil/compose.yml`) includes all 5 products, which recursively include all 10 PS-IDs. This deployment starts 40+ app instances, shared PostgreSQL, shared telemetry, and pki-init with suite-scoped cert generation. It has NEVER been tested end-to-end. Port formula (SERVICE + 20000) is untested.

### F4. Massive Code Duplication in TestMain Files

Each PS-ID's `testmain_e2e_test.go` is ~80 lines, of which ~70 are identical boilerplate:
- Same 4-step lifecycle (Start → WaitForHealth → Run → Stop)
- Same error handling pattern
- Same variable declarations
- Only differences: magic constant names, service-specific container names, port numbers

**Impact**: When a pattern change is needed (e.g., adding compose config validation, adding cleanup verification, adding admin health checks), every TestMain must be updated independently.

### F5. Magic Constant Explosion (~160 E2E constants for 4 PS-IDs)

Each PS-ID requires ~15-20 magic constants for E2E: compose file path, container names (×4), port numbers (×4), health endpoint, timeout. Current 4 PS-IDs already have 146+ E2E-specific magic constants across 6 magic files. Extending to all 10 PS-IDs would require ~400+ constants. Product-level (5) and suite-level (1) would add another ~200+.

**Root cause**: No data-driven test configuration. Each service hardcodes its own constants instead of deriving them from the entity registry (`registry.yaml`).

### F6. sqlite-2 Instance Not Tested

All existing E2E tests test only 3 of 4 instances: sqlite-1, postgres-1, postgres-2. The sqlite-2 instance is deployed but never health-checked or tested. This means 25% of deployed instances have zero E2E validation.

### F7. No Admin Endpoint E2E Testing

E2E tests only validate public health endpoints (`/service/api/v1/health`). Admin endpoints (`/admin/api/v1/livez`, `/admin/api/v1/readyz`) are never tested in the deployed Docker environment. The admin channel uses a different TLS configuration (127.0.0.1:9090 mTLS) that is completely unvalidated.

### F8. TLS Certificate Chain Not Validated

All E2E tests use `NewClientForTest()` with `InsecureSkipVerify: true`. This means:
- TLS certificate chain is never validated (any self-signed cert accepted)
- CA trust chain from pki-init output is never verified
- Certificate SANs are never validated against actual container DNS names
- mTLS client certificates are never presented or verified

The `NewClientForTestWithCA()` function exists but is unused by any E2E test.

### F9. PostgreSQL mTLS Not Verified (Framework v12 Deferred)

Framework v12 deferred Phase 3 (verify leader/follower TLS), Phase 6 (verify app mTLS), Phase 9 (Docker Compose full verification), and Phase 10.5 (verify admin mTLS). These 10 tasks require Docker and were tagged "Docker-deferred." They remain unimplemented.

### F10. No pki-init Output Verification

E2E tests don't validate that pki-init generates the correct certificate tree structure. The 14-category, 90-dir (PS-ID) / 666-dir (suite) cert tree is assumed correct but never verified by any automated test in a deployed Docker environment.

### F11. No Compose Config Validation Before Start

`ComposeManager.Start()` calls `docker compose up` directly without first running `docker compose config` to validate the compose file. Syntax errors, invalid includes, undefined secrets, or port conflicts are only caught at container startup time, not upfront.

### F12. No Cross-Service Communication Testing

Products with multiple PS-IDs (e.g., SM has sm-kms + sm-im; Identity has 5 services) never test inter-service communication. Cross-service mTLS, shared database state, and federation patterns are unvalidated.

### F13. No Parallel E2E Orchestration

E2E tests run sequentially — one compose stack per `go test` invocation. There's no orchestrator that can bring up all 16 deployments (sequentially or parallel) and validate them in a single test run with proper resource isolation.

### F14. No Cleanup Verification

`ComposeManager.Stop()` calls `docker compose down -v` but doesn't verify that all containers were actually removed. Orphaned containers from failed runs can cause port conflicts in subsequent test runs.

### F15. Health Endpoint Path Inconsistency

Different services use different health endpoint constants. Some use `/service/api/v1/health`, which is the actual path. The constants should be standardized and derived from a single source.

### F16. No Test Result Aggregation

When running E2E tests across multiple deployments, there's no aggregated test result report. Each `go test` invocation produces independent output with no correlation between deployment tiers.

### F17. ComposeManager Start Has No Retry Logic

If `docker compose up` fails transiently (e.g., image pull timeout, Docker daemon transient error), the test fails immediately with no retry. This makes tests non-deterministic in CI/CD environments.

### F18. E2E Tests Not Registry-Driven

Tests are not generated or driven from the canonical entity registry (`api/cryptosuite-registry/registry.yaml`). Each PS-ID's E2E test is hand-crafted. When a new PS-ID is added or ports change, E2E tests must be manually updated.

### F19. No Telemetry Verification in E2E

E2E tests deploy OTel Collector and Grafana LGTM but never verify that telemetry data is actually flowing from app instances through the collector to Grafana. The `TestE2E_OtelCollectorHealth` test is skipped in sm-im with a comment about port 13133 not being exposed.

### F20. No Graceful Shutdown Testing

E2E tests never exercise the `/admin/api/v1/shutdown` endpoint. Graceful shutdown behavior under Docker Compose `stop` is unverified.

---

## Architecture Decisions

### D1. Registry-Driven Test Configuration

**Decision**: All E2E test configuration (compose file paths, container names, port numbers, health endpoints) MUST be derived from `api/cryptosuite-registry/registry.yaml` at test runtime or via code generation. NEVER hardcode per-PS-ID magic constants.

**Rationale**: Eliminates the magic constant explosion (F5), ensures consistency with deployment configuration, and makes adding new PS-IDs automatic.

### D2. Shared TestMain Factory

**Decision**: A single reusable `TestMain` factory function in `e2e_infra` package MUST be used by all E2E tests. Service-specific test files only define test cases, not lifecycle management.

```go
// Framework provides this:
func RunE2ETestMain(m *testing.M, config E2EConfig) {
    // Standard 4-step lifecycle with all quality checks
}

// Each PS-ID test uses it:
func TestMain(m *testing.M) {
    e2e_infra.RunE2ETestMain(m, e2e_infra.E2EConfig{
        DeploymentTier: "ps-id",
        DeploymentID:   "sm-kms",
    })
}
```

**Rationale**: Eliminates code duplication (F4), ensures consistent lifecycle management, and makes pattern changes global.

### D3. Three-Tier Test Organization

**Decision**: E2E tests MUST be organized in three tiers matching the deployment hierarchy:

| Tier | Location | Compose File | Port Range |
|------|----------|-------------|-----------|
| PS-ID | `internal/apps/{PS-ID}/e2e/` | `deployments/{PS-ID}/compose.yml` | 8000-8999 |
| Product | `internal/apps/{PRODUCT}/e2e/` | `deployments/{PRODUCT}/compose.yml` | 18000-18999 |
| Suite | `internal/apps/cryptoutil/e2e/` | `deployments/cryptoutil/compose.yml` | 28000-28999 |

**Rationale**: Tests mirror the deployment hierarchy. Each tier validates different concerns (PS-ID: single service correctness, Product: multi-service integration, Suite: full system).

### D4. All 4 Instances Must Be Tested

**Decision**: Every PS-ID E2E test MUST validate all 4 instances: sqlite-1, sqlite-2, postgres-1, postgres-2. No instance may be skipped.

**Rationale**: Fixes F6. The sqlite-2 instance exists to validate multi-instance isolation. Skipping it defeats the purpose of the 4-instance deployment pattern.

### D5. TLS Certificate Chain Validation Required

**Decision**: E2E tests MUST validate the actual TLS certificate chain from pki-init output. Use `NewClientForTestWithCA()` with the generated CA certificate, NOT `InsecureSkipVerify: true`.

**Rationale**: Fixes F8. The pki-init output generates a full CA chain. Tests must verify that the chain is correct by using it for TLS verification.

### D6. Compose Config Validation Before Start

**Decision**: `ComposeManager.Start()` MUST run `docker compose config --quiet` before `docker compose up` to detect syntax errors, invalid includes, undefined secrets, and port conflicts early.

**Rationale**: Fixes F11. Catches configuration errors before expensive container builds/starts.

### D7. Admin + Public Endpoint Testing

**Decision**: E2E tests MUST validate both public endpoints (`/service/api/v1/health`, `/browser/api/v1/health`) AND admin endpoints (`/admin/api/v1/livez`, `/admin/api/v1/readyz`) on every deployed instance.

**Rationale**: Fixes F7. The admin channel uses different TLS (mTLS on 127.0.0.1:9090), and admin endpoint failures would be invisible without testing.

**Note**: Admin endpoints are bound to `127.0.0.1:9090` inside containers and are NOT exposed to the host. E2E tests MUST use `docker compose exec` to reach admin endpoints from inside the container network, or use a test-specific compose override that exposes admin ports.

### D8. Deterministic Orchestrator Command

**Decision**: A single Go command (`go run ./cmd/cicd-workflow -workflows=e2e`) MUST be capable of orchestrating all 16 E2E deployments in a deterministic order: PS-ID first (smallest), then Product, then Suite. Each deployment's compose stack MUST be fully torn down before the next starts (to avoid port conflicts), unless resource isolation is achieved via Docker networks.

**Rationale**: Fixes F13. Creates reproducible test execution independent of LLM agent behavior.

### D9. pki-init Output Verification

**Decision**: PS-ID E2E tests MUST include a test that verifies the pki-init-generated cert tree matches the expected directory count and structure for the deployment tier (90 dirs for PS-ID, varies for Product, 666 for Suite).

**Rationale**: Fixes F10. Certificate structure errors are silent unless explicitly verified.

### D10. Cleanup Verification and Orphan Detection

**Decision**: `ComposeManager.Stop()` MUST verify that all containers are removed after `docker compose down -v`. Before `Start()`, MUST check for orphaned containers from previous runs and clean them up.

**Rationale**: Fixes F14. Orphaned containers cause non-deterministic port-conflict failures.

---

## Phased Implementation Plan

### Phase 0: Prerequisite Cleanup and Framework v12 Completion

**Goal**: Complete the deferred Docker-dependent tasks from Framework v12 and fix structural issues before building the new E2E framework.

**Tasks**:
1. Reclassify `identity-authz/e2e/rotation_test.go` — move to integration test (it's not a compose E2E test)
2. Complete Framework v12 Phase 3: Verify PostgreSQL leader/follower TLS in Docker
3. Complete Framework v12 Phase 6: Verify app-to-PostgreSQL mTLS in Docker
4. Complete Framework v12 Phase 9: Docker Compose deployment verification
5. Complete Framework v12 Phase 10.5: Verify admin mTLS in deployment
6. Verify all 10 PS-ID compose files pass `docker compose config --quiet`
7. Verify all 5 Product compose files pass `docker compose config --quiet`
8. Verify suite compose file passes `docker compose config --quiet`

### Phase 1: E2E Test Configuration Registry

**Goal**: Replace per-PS-ID magic constants with a registry-driven configuration system.

**Tasks**:
1. Define `E2EDeploymentConfig` struct in `e2e_infra` package:
   ```go
   type E2EDeploymentConfig struct {
       Tier           string // "ps-id", "product", "suite"
       ID             string // "sm-kms", "sm", "cryptoutil"
       ComposeFile    string // relative to project root
       Instances      []InstanceConfig
       HealthEndpoint string
       HealthTimeout  time.Duration
       Profiles       []string
   }
   type InstanceConfig struct {
       Name      string // container service name
       PublicPort uint16
       AdminPort  uint16 // 0 if not host-exposed
       DBBackend  string // "sqlite" or "postgresql"
   }
   ```
2. Create `e2e_infra.LoadDeploymentConfig(tier, id)` that derives config from registry.yaml + port assignment rules
3. Create `e2e_infra.LoadAllDeploymentConfigs()` returning all 16 configs
4. Migrate existing sm-kms E2E to use registry-driven config (validate migration)
5. Remove per-PS-ID E2E magic constants from `internal/shared/magic/` (after all E2E tests are migrated)

### Phase 2: Shared TestMain Factory and ComposeManager Enhancements

**Goal**: Create a reusable TestMain that eliminates boilerplate duplication.

**Tasks**:
1. Enhance `ComposeManager` with compose config validation (`docker compose config --quiet` before `up`)
2. Add retry logic to `ComposeManager.Start()` (1 retry with cleanup between attempts)
3. Add orphan container detection and cleanup before `Start()`
4. Add cleanup verification after `Stop()` (verify all containers removed)
5. Create `RunE2ETestMain(m *testing.M, config E2EDeploymentConfig)` factory:
   - Step 0: Validate compose config
   - Step 1: Detect and clean orphans
   - Step 2: Start compose stack (with retry)
   - Step 3: Wait for all instance health (public + admin)
   - Step 4: Run tests
   - Step 5: Stop and verify cleanup
6. Create `E2ETestContext` struct that TestMain provides to test functions via package-level var:
   ```go
   type E2ETestContext struct {
       Config     E2EDeploymentConfig
       HTTPClient *http.Client // TLS-verified client
       Instances  map[string]InstanceTestEndpoints
   }
   ```

### Phase 3: Shared E2E Test Library

**Goal**: Create reusable test functions that work across all 16 deployments.

**Tasks**:
1. Create `e2e_helpers.TestAllInstancesHealthy(t, ctx)` — validates all 4 instances' public health
2. Create `e2e_helpers.TestAllInstancesAdminHealthy(t, ctx)` — validates admin livez/readyz (via `docker compose exec`)
3. Create `e2e_helpers.TestSQLiteIsolation(t, ctx)` — verifies sqlite-1 and sqlite-2 are isolated
4. Create `e2e_helpers.TestPostgreSQLSharedState(t, ctx)` — verifies postgres-1 and postgres-2 share state
5. Create `e2e_helpers.TestRegistrationFlow(t, ctx)` — tests both `/browser/` and `/service/` registration
6. Create `e2e_helpers.TestPKIInitCertTree(t, ctx)` — verifies cert directory structure
7. Create `e2e_helpers.TestTLSCertificateChain(t, ctx)` — validates TLS cert chain against pki-init CA
8. Create `e2e_helpers.TestPublicAndAdminEndpoints(t, ctx)` — combined public+admin health validation

### Phase 4: PS-ID E2E Tests for All 10 Services

**Goal**: Every PS-ID has a compose-based E2E test using the shared framework.

**Tasks**:
1. Migrate `sm-kms` E2E to shared TestMain factory (validate no regression)
2. Migrate `sm-im` E2E to shared TestMain factory (preserve registration tests)
3. Migrate `jose-ja` E2E to shared TestMain factory
4. Migrate `skeleton-template` E2E to shared TestMain factory
5. Create `pki-ca` E2E test using shared TestMain factory
6. Create `identity-authz` compose-based E2E test (replace rotation test with integration classification)
7. Create `identity-idp` E2E test using shared TestMain factory
8. Create `identity-rs` E2E test using shared TestMain factory
9. Create `identity-rp` E2E test using shared TestMain factory
10. Create `identity-spa` E2E test using shared TestMain factory

**Each PS-ID test MUST include**: Health checks (4 instances × public + admin), SQLite isolation, PostgreSQL shared state, TLS cert chain validation, pki-init tree verification.

### Phase 5: Product-Level E2E Tests (5 Products)

**Goal**: Validate product-level compose files with `!override` port substitution.

**Tasks**:
1. Create `internal/apps/sm/e2e/` — tests SM product deployment (sm-kms + sm-im, ports 18000-18199)
2. Create `internal/apps/jose/e2e/` — tests JOSE product deployment (jose-ja, ports 18200-18299)
3. Create `internal/apps/pki/e2e/` — tests PKI product deployment (pki-ca, ports 18300-18399)
4. Create `internal/apps/identity/e2e/` — tests Identity product deployment (5 services, ports 18400-18899)
5. Create `internal/apps/skeleton/e2e/` — tests Skeleton product deployment (skeleton-template, ports 18900-18999)

**Each product test MUST include**: All PS-ID instance health checks at product-level ports, product-level pki-init verification (correct cert tree scope), cross-service communication (for multi-PS-ID products), port formula validation (SERVICE + 10000).

### Phase 6: Suite-Level E2E Test

**Goal**: Validate the full suite deployment with all 40 app instances.

**Tasks**:
1. Create `internal/apps/cryptoutil/e2e/` — tests full suite deployment (all 10 PS-IDs, ports 28000-28999)
2. Verify all 40 app instances are healthy (10 PS-IDs × 4 instances)
3. Verify shared PostgreSQL handles all 10 PS-IDs' app connections
4. Verify shared telemetry (OTel + Grafana) receives data from all services
5. Verify suite-level pki-init generates 666 directories
6. Port formula validation (SERVICE + 20000)

### Phase 7: E2E Orchestrator Integration

**Goal**: Integrate all 16 E2E deployments into the `cicd-workflow` command for CI/CD.

**Tasks**:
1. Update `cicd-workflow` to support tiered E2E execution: `--e2e-tier=ps-id`, `--e2e-tier=product`, `--e2e-tier=suite`, `--e2e-tier=all`
2. Implement sequential execution with proper cleanup between deployments
3. Add aggregated test result reporting across all 16 deployments
4. Add CI/CD workflow for E2E: `.github/workflows/ci-e2e.yml`
5. Add timing reporting per deployment tier

### Phase 8: PostgreSQL mTLS E2E Verification (Framework v12 Completion)

**Goal**: Complete the deferred PostgreSQL mTLS verification tasks from Framework v12.

**Tasks**:
1. Verify PostgreSQL leader accepts TLS connections (Cat 11 server cert validated)
2. Verify PostgreSQL follower accepts TLS connections (Cat 11 server cert validated)
3. Verify leader↔follower replication uses mTLS (Cat 13 replication certs)
4. Verify app instances connect to PostgreSQL with mTLS (Cat 14 app client certs)
5. Verify non-mTLS connections are rejected by PostgreSQL (pg_hba.conf enforcement)
6. Verify admin channel mTLS works in deployed containers (Cat 6 + Cat 7 certs)

### Phase 9: Quality Gates and Documentation

**Goal**: Ensure all E2E tests meet quality standards and documentation is updated.

**Tasks**:
1. All 16 E2E tests pass deterministically (run 3 times consecutively, all pass)
2. All E2E tests use TLS certificate chain validation (no InsecureSkipVerify)
3. E2E test code coverage for `e2e_infra` and `e2e_helpers` packages ≥95%
4. All old per-PS-ID E2E magic constants removed from `internal/shared/magic/`
5. Update ENG-HANDBOOK.md Section 10.4 (E2E Testing Strategy) with new patterns
6. Update docs/deployment-templates.md with E2E test requirements per tier
7. Create fitness linter `e2e-coverage` that validates every PS-ID has E2E tests
8. Update CI/CD workflows to run tiered E2E

---

## Directory Count Verification Matrix

Each E2E tier MUST verify the pki-init cert tree directory count:

| Tier | Deployment | Expected Dirs (2 realms) | Formula |
|------|------------|--------------------------|---------|
| PS-ID | Any PS-ID | 90 | 30 global + 60 PS-ID |
| Product | sm (2 PS-IDs) | 150 | 30 global + 120 (60×2) |
| Product | jose (1 PS-ID) | 90 | 30 global + 60 (60×1) |
| Product | pki (1 PS-ID) | 90 | 30 global + 60 (60×1) |
| Product | identity (5 PS-IDs) | 330 | 30 global + 300 (60×5) |
| Product | skeleton (1 PS-ID) | 90 | 30 global + 60 (60×1) |
| Suite | cryptoutil (10 PS-IDs) | 666 | 30 global + 636 (varies due to shared Cat 9 admin/infra) |

---

## Port Assignment Verification Matrix

| Tier | Port Formula | Example (sm-kms sqlite-1) |
|------|-------------|--------------------------|
| PS-ID (SERVICE) | Base ports | 8000 |
| Product | SERVICE + 10000 | 18000 |
| Suite | SERVICE + 20000 | 28000 |

E2E tests at each tier MUST verify that instances respond on the correct tier-specific ports.

---

## Resource Requirements

- **Docker Desktop**: Required for all E2E tests
- **Memory**: Suite deployment requires ~16GB RAM (40 app instances + PostgreSQL + OTel + Grafana)
- **Disk**: pki-init generates ~666 directories with ~2000 files for suite tier
- **Time**: PS-ID E2E ~3-5 min each, Product ~5-10 min, Suite ~15-20 min
- **CI/CD**: E2E workflow should run PS-ID tier on every push, Product/Suite on nightly/weekly

---

## Success Criteria

1. All 16 Docker Compose deployments (10 PS-ID + 5 Product + 1 Suite) have E2E tests
2. All E2E tests are deterministic (same result on every run given same code)
3. All E2E tests use registry-driven configuration (no per-PS-ID magic constants)
4. All E2E tests validate all 4 instances per PS-ID (sqlite-1, sqlite-2, postgres-1, postgres-2)
5. All E2E tests validate TLS certificate chain (no InsecureSkipVerify)
6. All E2E tests validate both public and admin health endpoints
7. All E2E tests verify pki-init cert tree structure
8. Single command orchestrates all 16 deployments: `go run ./cmd/cicd-workflow -workflows=e2e`
9. PostgreSQL mTLS is verified in deployed Docker environment
10. E2E test framework code has ≥95% unit test coverage
