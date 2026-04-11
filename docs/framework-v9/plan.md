# Plan — Framework v9: Quality & Consistency

**Status**: In Progress (items 1, 4, 7, 14, 22 completed; items 3, 13 removed; items 15-23 added)
**Created**: 2026-04-08
**Updated**: 2026-04-11

---

## Overview

Framework v9 addresses quality improvements, documentation drift, and codebase consistency
issues identified during framework-v8 execution and subsequent handbook review.

Framework-v8 completed 43/43 tasks (100%) across 10 phases. The recursive Docker Compose
include architecture is fully operational at all 3 deployment tiers (SERVICE, PRODUCT, SUITE).

---

## Items

### 1. Dockerfile EXPOSE Port Standardization ✅ [HIGH]

**Source**: Deep research — deployment consistency audit

**Completed**: All 11 Dockerfiles standardized to `EXPOSE 8080` only. Admin port 9090 is
bound to 127.0.0.1 and is never exposed externally per §5.3 (Dual HTTPS Endpoint Pattern).
Compose port mappings handle host-side port differentiation.

### 2. Config File YAML Key Naming Inconsistency [HIGH]

**Source**: Deep research — config consistency audit

**Current state**: Services use two different YAML key naming conventions:
- **kebab-case** (correct per §13.2): jose-ja, pki-ca, skeleton-template
- **snake_case** (incorrect): sm-kms, sm-im, identity-authz, identity-idp, identity-rp,
  identity-rs, identity-spa

**Examples**: `bind_address` vs `bind-public-address`, `service_name` vs `service-name`

**Action**: Migrate snake_case configs to kebab-case. This requires updating both the YAML
files and the Go config parser code that reads them. Services using the framework config
parser already support kebab-case. Services using custom/older parsers need migration.

**Risk**: Config key changes require coordinating config files, Go parsers, deployment
overlay configs, and any documentation referencing config keys.

### ~~3. Unpinned Docker Image Versions~~ REMOVED

**Removed**: Policy reversed — open-source container images (Alpine, PostgreSQL, OTel
Collector, Grafana LGTM) now use `:latest` tags for automatic security patches. All
Dockerfiles use `ARG ALPINE_VERSION=latest` with hadolint DL3007 ignore. All compose
files and CI workflows updated to `postgres:latest`.

### 4. Dockerfile Healthcheck Standardization ✅ [LOW]

**Source**: Deep research — deployment consistency audit

**Completed**: All 11 Dockerfiles standardized to use built-in PS-ID livez CLI instead of
wget. Timing: `--start-period=30s`, `--interval=10s`, `--timeout=30s`. Pattern:
`CMD /app/{PS-ID} livez || exit 1`. Added `dockerfile-healthcheck` fitness linter
(30 test cases) to enforce this pattern. Removed wget from sm-im's apk install.

### 5. testpackage Linter: Resolve Enabled-But-Disabled State [MEDIUM]

**Source**: .golangci.yml investigation (framework-v8 handbook review)

**Current state**: testpackage linter is listed as enabled but `skip-regexp: '.*_test\.go$'`
matches ALL test files, making it a no-op. Documented in §11.3.1.

**Options**:
- **A**: Remove testpackage from enabled linters (honest configuration)
- **B**: Narrow skip-regexp to directories that legitimately need internal package testing
  and migrate other tests to external test packages (`package foo_test`)

**Action**: Evaluate option B first — identify which packages can reasonably use external
test packages. If migration scope is too large, choose option A.

### 6. goheader Linter: Monitor for golangci-lint v2.8+ Fix [LOW]

**Source**: .golangci.yml investigation (framework-v8 handbook review)

**Current state**: goheader is disabled due to a file corruption bug in golangci-lint v2
(replaces file contents instead of reporting violations). Documented in §11.3.1.

**Action**: When golangci-lint v2.8+ is released, test goheader on a branch. If the bug is
fixed, re-enable it. Template: `Copyright (c) {{ YEAR }} Justin Cranford`.

### ~~7. TODO P2.4 Tests in jose-ja Repository~~ ✅ COMPLETED

**Source**: TODO audit

**Resolution**: Implemented 13 skipped tests across 3 repository error test files.
All `t.Skip("TODO P2.4: ...")` calls removed. Tests use closedDB for error paths,
testDB for constraint violations, concurrent goroutines for race testing, and GORM
transactions for rollback verification. SQLite FK non-enforcement documented in tests.

### 8. Phase W Migration Handling in Framework [MEDIUM]

**Source**: TODO audit — `application_listener_test.go`

**Current state**: Two TODO comments indicate `StartCoreWithServices` doesn't run migrations
automatically. Tests work around this by running migrations manually.

**Action**: Evaluate whether `StartCoreWithServices` should handle migration execution as
part of its startup sequence. If yes, implement. If no, document the design decision and
remove the TODO comments.

### 9. Rate Limiting Implementation in identity-idp [MEDIUM]

**Source**: TODO audit — `handlers_security_validation_rate_test.go:177`

**Current state**: Comment says "Rate limiting implementation is deferred (MEDIUM priority
TODO)." The framework provides a `RateLimiter` in `internal/apps/framework/service/ratelimit/`
but identity-idp hasn't integrated it yet.

**Action**: Integrate the framework rate limiter into identity-idp handlers per the two-layer
rate limiting architecture documented in §8.5.2.

### 10. Test File Size Violations (3 files) [LOW]

**Source**: File size audit

**Current state**: Three test files exceed the 500-line hard limit:
- `validate_chunks_test.go` — 544 lines
- `jose_seam_injection_test.go` — 509 lines
- `issuer_operations_test.go` — 501 lines

**Action**: Refactor each file to split test cases into semantically named files, each under
500 lines. For example, `jose_seam_injection_test.go` could become
`jose_seam_rsa_test.go` + `jose_seam_ecdsa_test.go`.

### 11. Load Test Coverage: Product and Suite Tiers [LOW]

**Source**: framework-v8 carryover

**Current state**: `test/load/` (Gatling, Java 21, Maven) covers only some service-level
scenarios. Missing: 5 product-level and 1 suite-level load test scenarios.

**Action**: Extend `test/load/src/` with product-level and suite-level simulation classes.
Update `pom.xml` with new simulation entry points.

### ~~12. ENG-HANDBOOK Orphaned Section Coverage~~ ✅ [LOW]

**Source**: lint-docs validate-propagation output

**Current state**: 76 of 442 sections (##/### level) are "orphaned" — not referenced by any
`@propagate`/`@source` block. Combined ##/### coverage is 46%.

**Analysis**: All 76 orphaned sections have zero cross-references from any instruction file.
These sections are truly unreferenced — they consist of appendix/reference tables (A1-A3,
B1-B8, C1-C4), structural metadata sections, and parent-level headings whose child subsections
are already propagated. No actionable candidates exist for adding new @propagate/@source blocks.
Instruction files already cover all substantive content via existing propagation targets.

### ~~13. Fitness Linter for Unpinned Docker Image Tags~~ REMOVED

**Removed**: Policy reversed — open-source images now intentionally use `:latest` tags.
Instead, a `dockerfile-healthcheck` fitness linter was added to validate HEALTHCHECK
uses built-in PS-ID livez CLI (not wget/curl).

### 14. context.TODO() Usage in Production Code ✅ [LOW]

**Source**: TODO audit — `identity/repository/migrations.go:48`

**Completed**: Added `ctx context.Context` parameter to `Migrate()` function. Callers
(`AutoMigrate`) now pass startup context through instead of using `context.TODO()`.
Also replaced `context.TODO()` with `context.Background()` in identity test files.

### 15. Dockerfile Template Enforcement [CRITICAL]

**Source**: Deep research — Dockerfile consistency audit (deployment-templates.md)

**Current state**: 3 fundamentally different Dockerfile patterns exist across 10 PS-IDs:
- **Pattern A** (sm-kms, identity-authz/idp/rp/rs): 4-stage, WORKDIR /app/run, curl installed,
  GOMODCACHE/GOCACHE env vars, USER commented out, individual LABEL lines
- **Pattern B** (jose-ja, pki-ca, skeleton-template): 3-stage (no runtime-deps), adduser-based,
  compact LABEL, CMD with config path
- **Pattern C** (sm-im): 2-stage (no validation), user 1000:1000, no BuildKit caches, no static link check

**Target state**: ONE canonical 4-stage template (validation → builder → runtime-deps → final)
as defined in `docs/deployment-templates.md` Section B. All 10 PS-ID Dockerfiles MUST match
this template exactly (parameterized by PS-ID, DISPLAY_NAME).

**Action**: Rewrite all 10 PS-ID Dockerfiles to match the canonical template. Fix:
- skeleton-template: Remove all jose-ja copy-paste artifacts (header, username, paths)
- identity-spa: Fix COPY bug (copies /app/cryptoutil instead of /app/identity-spa)
- sm-im: Add validation stage, BuildKit caches, static link check, fix user to 65532
- Pattern A services: Remove curl, uncomment USER, remove GOMODCACHE/GOCACHE, fix WORKDIR
- All: Use compact LABEL block, no CMD, ENTRYPOINT with tini

**Risk**: Binary name changes (CMD removal) require compose.yml command verification.

### 16. Dockerfile Enforcement Linters [HIGH]

**Source**: cicd-lint gap analysis

**Current state**: Only 2 Dockerfile-specific fitness linters exist (`dockerfile-healthcheck`,
`dockerfile-labels`). The massive divergence in Section 15 was undetectable.

**Action**: Implement 8 new fitness linters as specified in `docs/deployment-templates.md`
Section N.1: `dockerfile_structure`, `dockerfile_binary_name`, `dockerfile_user`,
`dockerfile_entrypoint`, `dockerfile_no_cmd`, `dockerfile_no_curl`, `dockerfile_workdir`,
`dockerfile_no_goenv`.

### 17. Config Key Naming Migration (snake_case → kebab-case) [HIGH]

**Source**: NOTE — This supersedes and expands Item 2.

**Current state**: 7 of 10 PS-IDs use snake_case config keys. 3 use kebab-case (correct).
This affects `configs/{PS-ID}/{PS-ID}.yml` AND `deployments/{PS-ID}/config/*.yml`.

**Action**: Migrate all snake_case configs to kebab-case. Items affected:
- Standalone configs: sm-kms, sm-im (completely different schema from framework services)
- Deployment configs: sm-kms (common has mixed keys including `security: csrf_enabled`),
  sm-im (minimal - just bind + credentials)
- Go config parsers: Verify framework parser handles kebab-case (it does), verify
  sm-kms/sm-im custom parsers are updated
- Identity service standalone configs: identity-authz, identity-idp, identity-rp,
  identity-rs, identity-spa
- Identity deployment configs: all 5 services

**Risk**: Config key changes break running services. Coordinate config files + parsers + tests.

### 18. Deployment Config Overlay Standardization [MEDIUM]

**Source**: Deep research — config overlay structure audit

**Current state**: Deployment config overlays vary wildly:
- sm-kms: Common has TLS, unseal, credentials, CORS, CSRF. Instance files minimal.
- sm-im: Common has only bind + credentials. Instance files have bind + otlp only.
- jose-ja: Common has OTel, security-headers, rate-limiting, credentials. Instance files
  DUPLICATE common settings (security-headers, rate-limiting).
- skeleton-template: Copy-paste of jose-ja (says "JOSE Common Configuration" in header).

**Target state**: All deployment config overlays follow the template in
`docs/deployment-templates.md` Section D. Common file has shared settings (bind, TLS,
unseal, credentials, allowed-ips, CSRF). Instance files have ONLY instance-specific
settings (cors-origins, otlp-service, otlp-hostname, database-url for SQLite).

**Action**: Rewrite all deployment config overlays to match template.

### 19. Standalone Config Standardization [MEDIUM]

**Source**: Deep research — standalone config structure audit

**Current state**: Standalone configs have different schemas:
- sm-kms/sm-im: Deeply nested (service/database/server/admin/tls/features/rate_limit/cors),
  all snake_case, extensive inline documentation
- jose-ja: Flat kebab-case (bind-public-address, tls-enabled, cors-origins, otlp-*, etc.)
- skeleton-template: Copy-paste of jose-ja (header says "JOSE Authority Server",
  otlp-service says "skeleton-template-ja")

**Target state**: All standalone configs follow the template in `docs/deployment-templates.md`
Section E. Flat kebab-case keys matching the framework config parser.

**Action**: Rewrite standalone configs. sm-kms and sm-im require schema migration from deep
nested to flat kebab-case.

### 20. Suite Dockerfile Fix [LOW]

**Source**: Deep research — Dockerfile audit

**Current state**: cryptoutil suite Dockerfile does not install/copy tini. ENTRYPOINT is bare
`["/app/cryptoutil"]` without tini wrapper.

**Action**: Add tini to runtime-deps stage. Update ENTRYPOINT to
`["/sbin/tini", "--", "/app/cryptoutil"]`.

### 21. Config Enforcement Linters [MEDIUM]

**Source**: cicd-lint gap analysis

**Current state**: No linter validates config YAML key naming or config content structure.
The `validate_schema.go` in lint-deployments validates hardcoded schema but doesn't
check all the structural rules in deployment-templates.md.

**Action**: Implement 4 new fitness linters as specified in `docs/deployment-templates.md`
Section N.1: `config_key_naming`, `config_header_identity`, `config_instance_minimal`,
`config_common_complete`.

### 22. Deployment Templates Documentation ✅ [LOW]

**Source**: Planning — companion documentation

**Completed**: Created `docs/deployment-templates.md` as companion to
`docs/target-structure.md` (directory layout) and `docs/tls-structure.md` (TLS certs).
Defines canonical file content templates for all Dockerfiles, compose.yml files,
config files, and secrets at PS-ID, PRODUCT, and SUITE levels. Includes 24 Dockerfile
rules, 20 compose rules, 12 config rules, template-comparison linter architecture,
canonical template file catalog, and comprehensive inconsistency inventory.

### 23. Canonical Template Architecture (api/cryptosuite-registry/templates/) [HIGH]

**Source**: Architecture decision — template enforcement strategy

**Decision**: All canonical deployment templates are stored as parameterized template
files in `api/cryptosuite-registry/templates/`. This makes the templates
machine-readable and linter-consumable.

**Template files**:
- `Dockerfile.tmpl` — PS-ID Dockerfile (×10)
- `compose.yml.tmpl` — PS-ID compose (×10)
- `config-common.yml.tmpl` — Deployment common config (×10)
- `config-sqlite.yml.tmpl` — Deployment SQLite instance config (×20)
- `config-postgresql.yml.tmpl` — Deployment PostgreSQL instance config (×20)
- `standalone-config.yml.tmpl` — Standalone dev config (×10)
- `product-compose.yml.tmpl` — Product compose (×5)
- `suite-compose.yml.tmpl` — Suite compose (×1)
- `suite-Dockerfile.tmpl` — Suite Dockerfile (×1)

**Enforcement**: Linters instantiate templates in-memory by substituting registry
values, then compare byte-for-byte against actual files on disk. Any deviation is
a BLOCKING linting error with unified diff output.

**Relationship to existing docs**:
- `deployment-templates.md` describes the templates and their rules (human docs)
- `templates/` directory IS the templates (machine source of truth)
- `registry.yaml` provides the parameter values for instantiation

**Action**: Create the `api/cryptosuite-registry/templates/` directory and all 9
template files. Implement template-comparison linters in `cicd-lint lint-fitness`.
