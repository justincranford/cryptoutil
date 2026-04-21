# Framework V15 — Gap Analysis

**Generated**: 2026-04-21
**Source**: Deep analysis of all workflows, main code, unit tests, integration tests, E2E tests, docs, pre-commit/push checks, Copilot+Claude instructions/agents/skills, and TLS structure documentation.
**Purpose**: Catalogue ALL findings from deep analysis for future prioritization and remediation.

---

## Summary

| Severity | Category | Count | Fixed |
|----------|----------|-------|-------|
| CRITICAL | CI/CD | 1 | ❌ |
| HIGH | CI/CD | 2 | ❌ |
| HIGH | Main Code | 2 | ❌ |
| HIGH | Tests | 1 | ✅ |
| MEDIUM | CI/CD | 5 | ❌ |
| MEDIUM | Main Code | 2 | ❌ |
| MEDIUM | Pre-commit | 2 | ❌ |
| MEDIUM | Docs | 6 | ❌ |
| CODE QUALITY | Tests | 1 | ❌ |
| DESIGN CHOICE | Tests | 1 | (intentional) |
| FALSE POSITIVE | Tests | 1 | N/A |

---

## 1. CI/CD Workflow Gaps

### 1.1 — CRITICAL: `lint-docs` and `lint-deployments` Not in Any CI/CD Workflow

**Finding**: `go run ./cmd/cicd-lint lint-docs` and `go run ./cmd/cicd-lint lint-deployments` are only run via pre-commit hooks — they do NOT appear in any `.github/workflows/ci-*.yml` file. The `ci-fitness.yml` workflow runs `lint-fitness` standalone, but neither `lint-docs` (which validates `@propagate`/`@source` drift) nor `lint-deployments` (which validates deployment structure and config schema) are enforced in CI/CD. Developers who bypass `--no-verify` can push propagation drift and broken deployment configs that won't be caught until local pre-commit runs.

**Affected**: All contributors. Documentation drift and deployment config errors can reach `main`.
**Recommended Fix**: Add `lint-docs` and `lint-deployments` steps to `ci-quality.yml` (or a dedicated `ci-lint.yml`).
**Status**: ❌ Not fixed

---

### 1.2 — HIGH: `ci-coverage.yml` Uses `continue-on-error: true`

**Finding**: The coverage enforcement step in `ci-coverage.yml` has `continue-on-error: true`. This means coverage threshold violations (below 95%/98%) do NOT block CI/CD — the workflow marks the step yellow but continues to the next step and ultimately reports green. This defeats the purpose of having coverage gates.

**Recommended Fix**: Remove `continue-on-error: true` from coverage enforcement step.
**Status**: ❌ Not fixed

---

### 1.3 — HIGH: `ci-identity-validation.yml` Overly Permissive Permissions

**Finding**: `ci-identity-validation.yml` declares `contents: write, pull-requests: write` at the workflow level. The `pull-requests: write` permission is broader than needed for most jobs. Per ENG-HANDBOOK.md Section 9.7, permissions should be scoped to minimum required per job (`contents: read` by default).

**Also**: This workflow hardcodes `GO_VERSION` instead of consuming the version from `workflow-job-begin` outputs, violating the "same version everywhere" principle.

**Recommended Fix**: Scope permissions to minimum required per job; consume GO_VERSION from shared workflow output.
**Status**: ❌ Not fixed

---

### 1.4 — MEDIUM: Unpinned Docker Images in 6+ Workflows

**Finding**: Multiple workflows use unpinned Docker image tags (`postgres:latest`, `zaproxy:stable`) in service container configurations. Per security-first principles, Docker images should be pinned to specific digest or semantic version to prevent unexpected updates from breaking CI/CD.

**Affected workflows**: `ci-e2e.yml`, `ci-dast.yml`, `ci-load.yml`, and others using PostgreSQL containers.
**Recommended Fix**: Pin all Docker images to specific versions (`postgres:17.2`, `zaproxy/zap-stable:2.15.0`).
**Status**: ❌ Not fixed

---

### 1.5 — MEDIUM: `ci-race.yml` Missing Build Tag Exclusions

**Finding**: `ci-race.yml` runs `go test -race ./...` without build tag exclusions (`-tags=!integration,!bench,!fuzz,!e2e`). This means benchmark and fuzz test files are compiled but excluded by their function name pattern, while integration test files (which do NOT have build tags by design) ARE compiled and run under the race detector. If integration tests start real servers with `TestMain`, this significantly increases CI time.

**Note**: Per `test-file-suffix-rules.yaml`, integration build tags are OPTIONAL by design — the 8 services without build tags on their integration tests is intentional. Running them under the race detector may be desirable.
**Recommended Fix**: Evaluate whether `--tags=!bench,!fuzz,!e2e` (without `!integration`) is appropriate.
**Status**: ❌ Under investigation

---

### 1.6 — MEDIUM: `ci-mutation.yml` Short Artifact Retention

**Finding**: `ci-mutation.yml` uses `retention-days: 7` for mutation testing artifacts. Per ENG-HANDBOOK.md Section 9.7 (artifact management), security testing artifacts should be retained for 30 days. Mutation test results are a form of test quality assurance that informs security-relevant decisions.

**Recommended Fix**: Change retention-days from 7 to 30 in `ci-mutation.yml`.
**Status**: ❌ Not fixed

---

### 1.7 — MEDIUM: `ci-quality.yml` Missing Top-Level Permissions Block

**Finding**: `ci-quality.yml` does not declare an explicit top-level `permissions:` block. Per ENG-HANDBOOK.md Section 9.7 security requirements, all workflows MUST declare explicit minimum permissions. Without an explicit block, GitHub uses the organization default (typically `contents: read`, but this is not guaranteed).

**Recommended Fix**: Add `permissions: { contents: read }` block to `ci-quality.yml`.
**Status**: ❌ Not fixed

---

### 1.8 — MEDIUM: `ci-load.yml` No Artifact Retention Configuration

**Finding**: `ci-load.yml` uploads artifacts without `retention-days` configuration, resulting in the organization default (90 days). Load test artifacts are large (Gatling HTML reports) and should be explicitly configured with a shorter retention period (7 days per ENG-HANDBOOK.md Section 9.7 for temp logs).

**Recommended Fix**: Add `retention-days: 7` to load test artifact upload steps.
**Status**: ❌ Not fixed

---

### 1.9 — MEDIUM: `ci-sast.yml` Redundant Maven Cache

**Finding**: `ci-sast.yml` caches Maven dependencies even though the primary SAST analysis (CodeQL, gosec) does not use Maven. The Maven cache is only needed for the Java Gatling load test compilation subpath. This adds unnecessary CI/CD overhead and cache storage usage.

**Recommended Fix**: Move Maven dependency caching to `ci-load.yml` where it's actually needed.
**Status**: ❌ Not fixed

---

## 2. Main Code Gaps

### 2.1 — HIGH: Duplicate `usage.go` Files (8 files, 4 pairs)

**Finding**: 4 pairs of nearly identical `usage.go` files exist at both product-level and service-level entry points. The product-level files delegate to service-level commands but duplicate the usage string generation logic. This creates maintenance burden when adding new CLI flags or subcommands.

**Affected files** (4 pairs):
- `internal/apps/sm/usage.go` + `internal/apps/sm-kms/usage.go`
- `internal/apps/sm/usage.go` + `internal/apps/sm-im/usage.go`
- `internal/apps/jose/usage.go` + `internal/apps/jose-ja/usage.go`
- `internal/apps/pki/usage.go` + `internal/apps/pki-ca/usage.go`

**Recommended Fix**: Extract shared usage generation to `internal/apps/framework/service/usage/` with product/service composition.
**Status**: ❌ Not fixed

---

### 2.2 — HIGH: `sm-kms` Shutdown Missing Timeout Context

**Finding**: The `sm-kms` service calls `server.Shutdown()` with `context.Background()` (no timeout). If the server hangs during shutdown (e.g., long-running requests, connection draining), the process never terminates. All other services use a context with a 5-second timeout for graceful shutdown.

**Affected**: `internal/apps/sm-kms/server/server.go` `Shutdown()` call path.
**Recommended Fix**: Add `context.WithTimeout(ctx, magic.DefaultDataServerShutdownTimeout)` to sm-kms shutdown.
**Status**: ❌ Not fixed

---

### 2.3 — MEDIUM: Port Type Inconsistency in `identity-authz`

**Finding**: `identity-authz` server binding code has a missing `uint16` cast on the port value compared to all other services. This results in an implicit int-to-uint16 conversion that passes linting but is inconsistent with the established pattern across all 9 other services.

**Affected**: `internal/apps/identity-authz/server/server.go`.
**Recommended Fix**: Add explicit `uint16(port)` cast consistent with other services.
**Status**: ❌ Not fixed

---

### 2.4 — MEDIUM: Signal Handling Cleanup Inconsistency

**Finding**: Signal handling (`SIGTERM`, `SIGINT`) cleanup patterns differ across service entry points. Some services use `signal.Stop(sigChan)` + `close(sigChan)` in defer, others use only `signal.Stop()`. The inconsistency creates potential goroutine leaks if the signal channel is not properly closed on cleanup paths.

**Recommended Fix**: Standardize signal handling cleanup pattern across all 10 service entry points using the framework's canonical pattern.
**Status**: ❌ Not fixed

---

## 3. Unit Test Gaps

### 3.1 — HIGH: ✅ FIXED — Sequential Test With Parallel Subtests (`application_init_test.go`)

**Finding**: `internal/apps/sm-kms/server/application/application_init_test.go` — `TestServerInit_HappyPath` is marked `// Sequential: uses os.Chdir (global process state)` on the parent function (no `t.Parallel()` call). However, the table-driven subtest inside the loop called `t.Parallel()`, which would cause a race condition on the process-level working directory if additional test cases were added to the table.

**Fix Applied**: Removed `t.Parallel()` from subtest; added `// Sequential:` comment inside subtest body for clarity.
**Commit**: `fix(test): remove t.Parallel from subtests inside Sequential test with os.Chdir`
**Status**: ✅ Fixed in this session

---

### 3.2 — FALSE POSITIVE: `user_auth_test.go` `DisableKeepAlives` Concern

**Finding**: Analysis flagged that `internal/apps/framework/service/client/user_auth_test.go` lacks `DisableKeepAlives: true` on HTTP transport. However, this file uses `httptest.NewServer` (a mock HTTP server from the Go standard library), NOT a real Fiber/fasthttp server. The 90-second shutdown hang from `DisableKeepAlives` only applies to Fiber/fasthttp servers. Standard `httptest.NewServer` closes cleanly.

**Conclusion**: No fix needed. This is a false positive.
**Status**: ✅ Confirmed false positive — no action required

---

## 4. Pre-Commit/Pre-Push Gaps

### 4.1 — MEDIUM: `golangci-lint` Not Pinned in `.pre-commit-config.yaml`

**Finding**: `.pre-commit-config.yaml` specifies `golangci-lint` as a pre-commit hook without pinning to a specific version. Per ENG-HANDBOOK.md Section 11.3.3 ("ALWAYS pin golangci-lint to specific version"), the pre-commit hook MUST pin to the same version used in CI/CD (v2.7.2+). Without version pinning, different developers may run different linter versions, causing inconsistent pre-commit results.

**Recommended Fix**: Pin `golangci-lint` to `v2.7.2` (or higher if updated) in `.pre-commit-config.yaml`.
**Status**: ❌ Not fixed

---

### 4.2 — MEDIUM: `lint-docs` Not Enforced in CI/CD

**Finding**: `lint-docs` validates `@propagate`/`@source` drift between `ENG-HANDBOOK.md` and instruction files, and validates agent pairs. This linter runs in pre-commit hooks but is absent from all CI/CD workflows. See also Gap 1.1 (Critical) — this is a subset of that issue.

**Status**: ❌ Not fixed (see Gap 1.1)

---

## 5. Integration / E2E Test Gaps

### 5.1 — CODE QUALITY: `pki-ca` TestMain Uses Manual Polling Loop

**Finding**: `internal/apps/pki-ca/server/testmain_test.go` uses a manual 300-attempt polling loop to wait for the test server to start (checking HTTP health endpoints). All other services that start real servers in TestMain (jose-ja, skeleton-template, sm-kms, sm-im) use the shared `MustStartAndWaitForDualPorts` helper from `internal/apps/framework/service/testing/e2e_helpers/`.

**Recommended Fix**: Replace the manual loop with:
```go
cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(
    testServer, func() error { return testServer.Start(ctx) })
```
**Status**: ❌ Not fixed (deferred — no impact on correctness, only consistency)

---

### 5.2 — DESIGN CHOICE: Integration Test Build Tags Are Optional

**Finding**: Eight services (pki-ca, jose-ja, skeleton-template, identity-authz, identity-idp, identity-rp, identity-rs, identity-spa) have `*_integration_test.go` files WITHOUT `//go:build integration` build tags. Initial analysis flagged this as a "CRITICAL" issue, but deeper investigation revealed this is intentional.

**Evidence**: `internal/apps/tools/cicd_lint/lint_fitness/test_file_suffix_structure/test-file-suffix-rules.yaml` explicitly states:
> "Build tags are optional (some integration tests use TestMain pattern without tags, others declare `//go:build integration` or `//go:build !integration`)."

**Conclusion**: The `required_build_tags` field is empty for `_integration_test.go` files. Both patterns are accepted. SM-KMS and SM-IM use the integration tag pattern; other services use the tagless TestMain pattern. No action needed.
**Status**: ✅ Confirmed intentional design — no action required

---

## 6. TLS/PKI Documentation Gaps

*(See also: [docs/tls-structure-suggestions.md](../tls-structure-suggestions.md) for detailed recommendations)*

### 6.1 — MEDIUM: Admin CA Bundle (`issuing-ca.pem`) Undocumented in `tls-structure.md`

**Finding**: `tls-structure.md` documents the `private-https-admin-server-entity-{PS-ID}/` certificate but does not mention the admin CA bundle file (`issuing-ca.pem`) that must be distributed to clients verifying the admin TLS endpoint. The admin CA trust chain is only partially described.

**Status**: ❌ Documented in tls-structure-suggestions.md

---

### 6.2 — MEDIUM: `tls-config.yml` Pattern Undocumented

**Finding**: The `TLSModeMixed` dynamic cert generation pattern — where a service reads `tls-config.yml` to determine which pre-generated certs to use vs. which to generate dynamically — is not documented in `tls-structure.md`. Developers working on cert initialization do not know when to use `tls-config.yml` vs. pre-generated certs.

**Status**: ❌ Documented in tls-structure-suggestions.md

---

### 6.3 — MEDIUM: Realm Dynamic Binding Explanation Missing

**Finding**: `tls-structure.md` references "Decision 8: Realm dynamic binding from registry.yaml" but does not explain the binding mechanism. How does the service discover which realm to bind to based on the registry? This gap affects developers implementing new services.

**Status**: ❌ Documented in tls-structure-suggestions.md

---

### 6.4 — MEDIUM: PostgreSQL Naming Ambiguity (`postgres` vs `postgres-1`/`postgres-2`)

**Finding**: Certificate categories in `tls-structure.md` use inconsistent PostgreSQL naming:
- Cat 4–5: Use `postgres` (shared domain, e.g., `public-https-client-issuing-ca-{PS-ID}-postgres/`)
- Cat 6–7, 14: Use `postgres-1`/`postgres-2` (individual containers, e.g., `private-https-server-entity-postgres-1/`)

This ambiguity makes it unclear whether a specific cert directory targets the shared PostgreSQL cluster or an individual PostgreSQL container.

**Status**: ❌ Documented in tls-structure-suggestions.md

---

### 6.5 — MEDIUM: Directory Count Formula Derivation Missing

**Finding**: `tls-structure.md` states "639 total cert directories" without showing the derivation formula. Per ENG-HANDBOOK.md Section 14.1.2 ("Derive directory/file counts from pattern expansion"), counts MUST include the formula (e.g., `30 global + 60 per-PS-ID × 10 = 630`). Reviewers cannot verify the count without the breakdown.

**Status**: ❌ Documented in tls-structure-suggestions.md

---

### 6.6 — MEDIUM: V12/V13 Phase Dependency Graph Missing from `pki-init-order.md`

**Finding**: `pki-init-order.md` describes V12 and V13 phases but lacks a visual dependency graph showing which phases can run in parallel and which have strict ordering. The cross-plan dependency section is also contradictory: it states "V13 Phase 0 can be done in parallel with V12 Phases 1-9" but V13 Phase 0 requires V12 Phase 0 cert categories to exist as prerequisites.

**Status**: ❌ Documented in tls-structure-suggestions.md

---

## 7. Copilot + Claude Instructions/Agents/Skills Gaps

### 7.1 — ✅ FIXED — `fix-workflows` Agents Missing ENG-HANDBOOK Section 9 Reference

**Finding**: Both `.github/agents/fix-workflows.agent.md` and `.claude/agents/fix-workflows.md` lacked cross-references to ENG-HANDBOOK.md Section 9 (CI/CD Workflow Architecture). Per `06-02.agent-format.instructions.md`, "Agents modifying CI/CD workflows or infrastructure MUST reference infrastructure architecture (Section 9)."

**Fix Applied**: Added cross-references to Sections 9.7, 9.9, 9.10, and 14.7 to both agent files.
**Commit**: `docs(agents): add ENG-HANDBOOK Section 9 cross-references to fix-workflows agents`
**Status**: ✅ Fixed in this session

---

## 8. Low Priority / Future Work

### 8.1 — LOW: `DEV-SETUP.md` Missing ENG-HANDBOOK Cross-References

**Finding**: `docs/DEV-SETUP.md` does not reference ENG-HANDBOOK.md sections for architecture decisions it describes. Developers reading DEV-SETUP.md have no way to find the canonical source of truth for the patterns they're setting up.

**Recommended Fix**: Add "See ENG-HANDBOOK.md §X for..." cross-references to relevant sections in DEV-SETUP.md.
**Status**: ❌ Low priority, not fixed

---

## Prioritization Recommendation

**Fix next (unblocking CI/CD integrity):**
1. Gap 1.1 (CRITICAL): Add `lint-docs` + `lint-deployments` to CI/CD workflow
2. Gap 1.2 (HIGH): Remove `continue-on-error: true` from coverage enforcement
3. Gap 2.2 (HIGH): Add shutdown timeout to sm-kms

**Fix in next sprint:**
4. Gap 4.1 (MEDIUM): Pin golangci-lint version in `.pre-commit-config.yaml`
5. Gap 1.3 (HIGH): Scope permissions in `ci-identity-validation.yml`
6. Gap 2.1 (HIGH): Extract duplicate `usage.go` files to framework package
7. Gap 5.1 (CODE QUALITY): Replace manual loop with `MustStartAndWaitForDualPorts` in pki-ca testmain
8. Gaps 6.1–6.6: Update tls-structure.md per suggestions doc

**Lower priority:**
9. Gaps 1.4–1.9 (MEDIUM): Various CI/CD improvements
10. Gaps 2.3–2.4 (MEDIUM): Port cast and signal handling cleanup
11. Gap 8.1 (LOW): DEV-SETUP.md cross-references
