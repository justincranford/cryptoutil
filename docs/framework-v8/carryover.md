# Framework-v8 Carryover Items

**Created**: 2026-04-05
**Source**: Analysis of `docs/framework-v7/` (plan.md, tasks.md, lessons.md) + `docs/target-structure.md` Section N

Numbered list, prioritized from highest to lowest impact. Items marked ✅ have been moved to
[carryover-completed.md](carryover-completed.md). Items marked ~~BLOCKED~~ have been analyzed
and determined infeasible.

**Completed items** (see [carryover-completed.md](carryover-completed.md)): 2.1, 4, 5, 6, 8

---

## 1. ~~Move `internal/shared/apperr/` → `internal/apps/framework/apperr/`~~ [BLOCKED]

**Current state**: `internal/shared/apperr/` contains `app_errors.go`, `http_errors.go`, and
`http_status_line_and_code.go` — application-level HTTP error abstraction used by all services.

**Why BLOCKED**: Moving `apperr` to `internal/apps/framework/` would create a circular dependency.
The package is imported by `internal/shared/crypto/jose/` (10+ files), `internal/shared/util/random/`,
and `internal/shared/telemetry/` — all of which live in `internal/shared/`. These shared packages
CANNOT import from `internal/apps/framework/` (lower layer cannot depend on upper layer).
The package legitimately IS cross-cutting infrastructure: UUID validation errors and HTTP error
mapping are used at every layer. It correctly belongs in `internal/shared/`.

**Resolution**: Keep `internal/shared/apperr/` in its current location. The original rationale
("HTTP status codes are a framework concern") is incorrect — the error types are used by
shared crypto, telemetry, and utility packages that are below the framework layer.

---

## 2. Create Product-Level Dockerfiles (5 Products) [HIGH]

**Current state**: Product-level deployment directories (`deployments/sm/`, `deployments/jose/`,
`deployments/pki/`, `deployments/identity/`, `deployments/skeleton/`) are missing Dockerfiles.

**Why HIGH**: The `dockerfile_labels` fitness linter and `deployment_dir_completeness` validator
already enforce per-service Dockerfiles. Product-level Dockerfiles are needed to support
product-tier Docker Compose deployments (compose.yml at the product level launches all services
within a product via a single product-built image). Without these, the product-tier deployments
are structurally incomplete.

**Template**: Use the existing `skeleton-template/Dockerfile` as the base. Each product Dockerfile
must include the standard OCI labels (`org.opencontainers.image.title`, etc.) matching the product
display name.

**Action**: Create `deployments/{sm,jose,pki,identity,skeleton}/Dockerfile` (5 files).

---

## 3. Fitness Linter: `usage/service_browser_health_paths` [MEDIUM]

**Current state**: Each PS-ID has a `{ps-id}_usage.go` file with CLI usage strings. These usage
strings MUST mention both `/service/**` and `/browser/**` health paths because every PS-ID exposes
both endpoints. This is not currently enforced by any fitness linter.

**Why MEDIUM**: Discovered during framework-v7 Task 5.5 (deleted 5 product-level usage files).
The deletion of `identity/authz/authz_usage.go` etc. revealed that ensuring PS-ID usage files
contain both health path variants is a code quality invariant that should be automatically verified.

**Pattern**: The linter should scan all `*_usage.go` files under `internal/apps/{PS-ID}/` and
verify they contain both the string `/service/api/v1/health` and `/browser/api/v1/health`.
Files that omit either path are flagged as violations.

**Action**: Create `internal/apps/tools/cicd_lint/lint_fitness/usage_health_path_completeness/`
with `lint.go` + `lint_test.go`, register in the fitness registry.

---

## 7. Load Test Refactoring: All Tiers [LOW]

**Current state**: `test/load/` (Gatling, Java 21, Maven) covers only some service-level
scenarios. Per `docs/target-structure.md` Section I, the target is:
- 10 service-level load test scenarios (one per PS-ID)
- 5 product-level load test scenarios (one per product)
- 1 suite-level load test scenario

**Why LOW**: Load tests do not block CI/CD and require Java/Gatling expertise to extend.
However, the gap means production throughput characteristics at product and suite levels are
unknown until these are created.

**Action**: Extend `test/load/src/` to add product-level and suite-level simulation classes.
Ensure `pom.xml` is updated with the new simulation entry points.
