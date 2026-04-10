# Plan — Framework v9: Quality & Consistency

**Status**: Not Started
**Created**: 2026-04-08
**Updated**: 2026-04-10

---

## Overview

Framework v9 addresses quality improvements, documentation drift, and codebase consistency
issues identified during framework-v8 execution and subsequent handbook review.

Framework-v8 completed 43/43 tasks (100%) across 10 phases. The recursive Docker Compose
include architecture is fully operational at all 3 deployment tiers (SERVICE, PRODUCT, SUITE).

---

## Items

### 1. Dockerfile EXPOSE Port Standardization [HIGH]

**Source**: Deep research — deployment consistency audit

**Current state**: 10 of 11 Dockerfiles use incorrect `EXPOSE` statements. Per §5.3 (Dual
HTTPS Endpoint Pattern) and §3.4 (Port Design), ALL services must internally expose 8080
(public) and 9090 (admin). The compose files handle host port mapping.

**Affected services**: sm-im, jose-ja, skeleton-template, pki-ca, identity-authz,
identity-idp, identity-rp, identity-rs, identity-spa, sm-kms. Only the cryptoutil
Dockerfile already uses the correct `EXPOSE 8080 9090`. Each affected Dockerfile uses
non-standard public/admin port values (host-level ports, wrong admin ports, or missing
the second port entirely).

**Action**: Standardize all Dockerfiles to `EXPOSE 8080 9090`. Verify that compose port
mappings handle the host-side port differentiation.

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

### 3. Unpinned Docker Image Versions [MEDIUM]

**Source**: Deep research — deployment consistency audit

**Current state**: 13 occurrences of `:latest` tags in compose files:
- `alpine:latest` in 11 healthcheck-secrets services
- `otel/opentelemetry-collector-contrib:latest` in shared-telemetry
- `grafana/otel-lgtm:latest` in shared-telemetry

**Action**: Pin all images to specific versions:
- `alpine:latest` → `alpine:3.19`
- Pin OTel collector and Grafana LGTM to current stable versions
- Add a fitness linter to detect unpinned image versions in compose files

### 4. Dockerfile Healthcheck Standardization [LOW]

**Source**: Deep research — deployment consistency audit

**Current state**: `--start-period` varies across Dockerfiles (5s, 10s, 30s). sm-kms uses a
custom healthcheck command (`/app/sm-kms server ready --dev`) while others use `wget` against
the admin livez endpoint.

**Action**: Standardize all healthchecks to:
- `--start-period=30s` (sufficient for TLS cert generation)
- `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez`

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

### 7. TODO P2.4 Tests in jose-ja Repository [MEDIUM]

**Source**: TODO audit

**Current state**: 12 skipped test cases in `jose-ja/server/repository/` with `t.Skip("TODO
P2.4: ...")`. These cover FK constraint tests, mocked database tests, transaction rollback
scenarios, and cascade deletion tests.

**Files affected**:
- `material_jwk_repository_error_test.go` (6 skipped tests)
- `elastic_jwk_repository_error_test.go` (6 skipped tests)

**Action**: Implement the skipped tests. Prioritize FK constraint and cascade deletion tests
(require schema migration changes). Mocked database tests are lower priority.

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

### 12. ENG-HANDBOOK Orphaned Section Coverage [LOW]

**Source**: lint-docs validate-propagation output

**Current state**: 76 of 442 sections (##/### level) are "orphaned" — not referenced by any
`@propagate`/`@source` block. Combined ##/### coverage is 46%.

**Note**: This is informational — orphaned sections are warnings, not errors. Many stubs and
appendix sections intentionally have no propagation targets. Increasing coverage to ~60%
would capture the most-referenced sections.

**Action**: Identify the 10 most-referenced orphaned sections (by cross-reference count in
instruction files) and add `@propagate`/`@source` blocks for them.

### 13. Fitness Linter for Unpinned Docker Image Tags [LOW]

**Source**: Identified during item 3 research

**Current state**: No automated check for `:latest` tags or missing version pins in compose
files. The secrets validator catches inline credentials but not unpinned images.

**Action**: Add a fitness sub-linter to `lint_fitness` that scans compose files for unpinned
image tags (`:latest`, no tag at all). Exclude local build images (`image: cryptoutil:local`).

### 14. context.TODO() Usage in Production Code [LOW]

**Source**: TODO audit — `identity/repository/migrations.go:48`

**Current state**: `context.TODO()` used during migration runs at startup. This is a known
Go pattern for code paths where context isn't yet available, but it should be replaced with
a real context from the startup sequence.

**Action**: Pass the startup context through to migration functions instead of using
`context.TODO()`.
