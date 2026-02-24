# Implementation Plan - Consolidated Quality Fixes v7

**Status**: Planning
**Created**: 2026-02-23
**Last Updated**: 2026-02-23 (updated per quizme-v1 answers Q1=E, Q2=E, Q3=E)
**Purpose**: Consolidate ALL incomplete work from fixes-v1 through fixes-v6 and implementation-plan-v1 into a single actionable plan. Prior plan directories will be deleted after this plan is created.

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark complete without verification

**ALL issues are blockers - NO exceptions.**

## Overview

This plan consolidates 7 prior plan directories (fixes-v1 through fixes-v6, implementation-plan-v1) into actionable work. Prior plans had significant overlap and many items were already completed. This plan captures ONLY genuinely incomplete work, verified against the current codebase state (2026-02-23).

## Background

**Prior completed work** (verified):
- Build: clean (`go build ./...`, `go build -tags e2e,integration ./...`)
- Linting: clean (`golangci-lint run`, `golangci-lint run --build-tags e2e,integration`)
- Tests: full suite passes (`go test ./... -shuffle=on`, 1 flaky test noted)
- Deployment validators: 65/65 pass
- Poll utility: extracted to `internal/shared/util/poll/` with 100% coverage
- Duration/timeout constants: consolidated in `internal/shared/magic/magic_testing.go`
- Gremlins config: consolidated to single `.gremlins.yaml` with 95% thresholds
- PRODUCT compose files: all 5 created with correct 18XXX port ranges
- Docker Compose fixes: telemetry includes resolved, identity command syntax fixed
- Deployment/config restructuring: complete (SERVICE/PRODUCT/SUITE hierarchy)
- 8 deployment validators: implemented and passing

**Remaining work categories** (from deep analysis):
1. Code quality: bugs, missing guards, file naming, nolint violations
2. Test quality: missing t.Parallel(), flaky tests, missing test coverage
3. Magic constant consolidation: identity/pki magic packages not in shared/magic
4. E2E infrastructure: service startup blockers, compose paths, workflow updates
5. Production coverage: crypto/jose at 89.9% (structural ceiling ~91%)
6. File size: 4 files exceed 500-line hard limit

## Critical Cross-Cutting Findings

These findings affect multiple phases and must be tracked separately:

### 1. sm-kms Migration Debt (~19h)
- `server/application/application_core.go` + `application_basic.go`: Old pre-builder wrappers
- `server/middleware/`: **10 non-test files** of custom middleware (actual count; previously noted as 15):

  | File | Lines | Purpose | Status vs Template |
  |------|-------|---------|-------------------|
  | `claims.go` | 332 | OIDCClaims, ClaimsExtractor, scope/role/permission helpers | Promote to template |
  | `errors.go` | 287 | OAuth2/RFC 7807 error responses, AuthErrorResponder | Promote to template |
  | `introspection.go` | 309 | BatchIntrospector, token introspection with caching | Promote to template |
  | `jwt.go` | 247 | JWT validation with JWK caching | Compare vs `realms/middleware.go:JWTMiddleware` |
  | `jwt_revocation.go` | 417 | JWT revocation checking | Promote to template |
  | `realm_context.go` | 161 | Realm context from request | Compare vs `realms/` package |
  | `scopes.go` | 327 | ScopeValidator, scope enforcement | Promote to template |
  | `service_auth.go` | 436 | AuthMethod enum, mTLS/Bearer service auth | Promote to template |
  | `session.go` | 180 | Session token validation | Likely DUPLICATE of `template/middleware/session.go` |
  | `tenant.go` | 105 | Tenant context extraction | Compare vs `realm/tenant.go` |

  **Decision during migration**: compare each file against template equivalents before choosing to promote, delete, or keep sm-kms-specific. This comparison is part of the ~19h estimate.
- `server.go:35,49`: TODOs for SQLRepository→GORM and GORM+barrier integration
- No integration tests, no E2E tests
- **Note**: sm-kms is LAST per ARCHITECTURE.md migration order, but debt blocks MERGE0b and MERGE2/3 options
- **Tracked in**: research/plan-PKI-CA-MIGRATE.md Phase A, research/plan-PKI-CA-MERGE0b.md Phase Pre

### 2. Template Testing Infrastructure Gap
- No generic `StartServiceFromConfig()` helper in template
- cipher-im uses raw 50×100ms polling loop (`testing/testmain_helper.go`)
- jose-ja uses same raw polling pattern
- **Tracked in**: Task 6.0

### 3. ci-e2e.yml Path Bug
- References `deployments/jose/compose.yml` (should be `deployments/jose-ja/compose.yml`)
- All non-cipher-im E2E tests have `SERVICE_TEMPLATE_TODO` comments (disabled)
- **Tracked in**: Task 6.4

### 4. wsl Violations (22 total)
- 2 legacy `//nolint:wsl` at `template/service/telemetry/telemetry_service_helpers.go:134,158` — MUST remove
- 20 `//nolint:wsl_v5` in 5 identity unified files × 4 instances — make genuine effort to fix
- **Tracked in**: Task 2.3

## Architecture Direction

### Product Taxonomy Decision: Option D (3 Products)

User selected **Option D** from [DEEP-RESEARCH.md](research/DEEP-RESEARCH.md): 3 products where SM absorbs kms + im + ja:

| Product | Services |
|---------|----------|
| **SM** (Secret Management) | sm-kms, sm-im (renamed cipher-im), sm-ja (renamed jose-ja), sm-secrets (future) |
| **PKI** | pki-ca, pki-ocsp (future), pki-crldp (future) |
| **Identity** | identity-authz, identity-idp, identity-rs, identity-rp, identity-spa |

### Circular Dependency Concern — RESOLVED

**Concern**: If sm-ja (jose-ja) is inside SM product, and Identity needs JWT issuance (OIDC ID Tokens, OAuth Access Tokens), but SM services may federate authn to Identity → potential circular dependency.

**Resolution: Library vs Service Layer Separation**

The circular dependency does not exist because of the two-layer architecture already in the codebase:

**Layer 1 — Go Libraries (compile-time imports, no runtime dependency)**:
- `internal/shared/crypto/jose/` — JWK/JWS/JWE/JWT operations (already used by identity services)
- `internal/shared/barrier/` — encryption-at-rest (already built into service template for ALL services)
- **Identity services already import `shared/crypto/jose` directly** — verified in identity-idp, identity-rs, identity-spa, identity-authz, identity-rp `server.go` files
- Identity does NOT call sm-ja as an HTTP service

**Layer 2 — HTTP Services (runtime calls)**:
- sm-ja (jose-ja) — HTTP API wrapping the jose library for EXTERNAL consumers
- identity-authz/idp — HTTP APIs for authn/authz

**Dependency flow (no cycles)**:
```
Compile-time (library imports):
  identity-* ──import──→ shared/crypto/jose (library)
  identity-* ──import──→ template/barrier (built-in)
  sm-ja      ──import──→ shared/crypto/jose (library)
  sm-kms     ──import──→ shared/crypto/jose (library)
  pki-ca     ──import──→ shared/crypto/jose (library)

Runtime (HTTP calls, all optional federation):
  sm-kms  ──HTTP──→ identity-authz (optional authn federation)
  sm-im   ──HTTP──→ identity-authz (optional authn federation)
  sm-ja   ──HTTP──→ identity-authz (optional authn federation)
  pki-ca  ──HTTP──→ identity-authz (optional authn federation)

  external ──HTTP──→ sm-ja (JOSE operations API)
  external ──HTTP──→ sm-kms (key management API)
```

**Why no cycle**: Identity imports the jose LIBRARY at compile time. It never calls sm-ja at runtime. SM/PKI services optionally federate to Identity at runtime. Identity never calls back to SM/PKI. All arrows point one direction.

### cipher-im Placement Decision

**Decision**: Move cipher-im under SM PRODUCT as sm-im (standalone service rename).
- SM PRODUCT will contain: kms and im services (and likely more in future: secrets, ja, ssh, etc.)
- **Plan and tasks**: See [research/plan-PKI-CA-MERGE0a.md](research/plan-PKI-CA-MERGE0a.md) and [research/tasks-PKI-CA-MERGE0a.md](research/tasks-PKI-CA-MERGE0a.md)
- **PKI-CA-MERGE0b will NOT be implemented** (merge into sm-kms rejected: 8× effort for same outcome)

### pki-ca Strategy: Under Reconsideration

quizme-v2.md Answer changed to "???" — user is reconsidering pki-ca strategy in light of Option D product taxonomy. The pki-ca migration approach (MIGRATE vs MERGE1) remains open.

## Technical Context

- **Language**: Go 1.25.5
- **Linter**: golangci-lint v2.7.2
- **Database**: PostgreSQL OR SQLite with GORM
- **Mutation Testing**: gremlins with 95% thresholds

## Phases

### Phase 1: Critical Bug Fixes (2h) [Status: ☐ TODO]
**Objective**: Fix actual bugs that affect correctness
- F-6.4: ValidateUUIDs wraps wrong error
- F-6.5: Copy-paste bug — "sqlite" in PostgreSQL function name (verify if real)
- F-6.6: Generic error messages leak JWK context
- F-1.1: poll.go nil conditionFn panic
- F-1.2/F-1.3: poll.go zero/negative timeout/interval validation
- F-1.5: poll.go context not checked before first conditionFn call
- Flaky test: `TestAuditLogService_LogOperation_AuditDisabled` (race condition)
- **Success**: All bugs fixed, tests pass

### Phase 2: Code Quality & Standards (3h) [Status: ☐ TODO]
**Objective**: Fix style/standards violations
- F-6.9: File with space in name: `usernames_passwords_test util.go`
- F-6.8: Error sentinels typed as string not error
- F-6.15: `//nolint:wsl` violations — **Q3=E**: remove all 2 legacy `//nolint:wsl` (v1, must be gone); make genuine effort to fix all 20 `//nolint:wsl_v5` (modern golangci-lint v2); restructure code rather than suppress
- F-3.2: TestNegativeDuration not a `time.Duration` type
- F-3.4: `//nolint:stylecheck` without bug reference (5 instances)
- F-1.4: poll.go timeout error not wrapped with sentinel
- F-6.17: pool.go if/else chain → switch statement
- F-6.44: ValidateUUID takes `*string` pointer unnecessarily
- **Success**: Zero `//nolint:wsl` (legacy v1); all `//nolint:wsl_v5` either fixed by code restructure or documented as structurally required; linting clean

### Phase 3: Magic Constant Consolidation (2h) [Status: ☐ TODO]
**Objective**: Move scattered magic constants to `internal/shared/magic/`
- F-6.10: Identity magic package → shared/magic
- F-6.11: PKI CA magic package → shared/magic
- F-6.12: Identity config magic file → shared/magic
- F-6.25: TLS constants → shared/magic
- F-6.36: Duplicate identity/demo constants
- F-4.2: Demo package 20+ scattered constants
- **Success**: All magic constants in `internal/shared/magic/`, no domain-specific magic packages

### Phase 4: Test Quality Improvements (4h) [Status: ☐ TODO]
**Objective**: Fix test compliance issues
- F-6.16: 35 test files missing `t.Parallel()` (bulk fix)
- F-2.11: PKI `server_highcov_test.go` uses `time.Sleep` (4 instances)
- F-2.12-2.14: Demo files use `time.Sleep` for server startup
- F-6.13: 4 test files exceed 500-line hard limit (split)
- F-6.14: Identity server & cmd packages have zero tests
- F-6.32: identity/rp/ and identity/spa/ have zero tests
- F-6.33: pki/ca/domain/ has zero tests
- F-1.6: poll_test.go missing edge case tests (nil, zero, negative)
- **Success**: All tests have t.Parallel(), no time.Sleep in tests, all files ≤500 lines

### Phase 5: Dependency & Architecture (2h) [Status: ☐ TODO]
**Objective**: Fix architectural violations
- F-6.7: Shared packages import from apps/template (dependency inversion)
- F-4.1: Identity healthcheck/poller.go duplicates poll.Until
- F-6.34: `//nolint:wrapcheck,thelper` blanket suppressions
- F-6.35: jose package name mismatch
- F-6.37: Unused sentinel errors in database/sharding.go
- F-6.38: SQL interpolation in sharding (defense in depth)
- F-6.39: `fmt.Errorf` without `%w` audit
- **Success**: No architectural violations, clean dependency graph

### Phase 6: E2E Infrastructure (4h) [Status: ☐ TODO]
**Objective**: Fix E2E test infrastructure by standardizing service startup patterns **per Q1=E**:

**Step B (do first): Fix cipher-im service startup reliability**
- Make cipher-im service startup reliable in both main code and test code
- Ensure all main/test startup code is reusable via service-template (currently lives in `cipher/im/testing/testmain_helper.go`; needs a generic version in `template/service/testing/`)
- Extract `StartServiceFromConfig()` generic helper into template testing package
- Verify cipher-im E2E tests pass end-to-end

**Step A (do second): Propagate to jose-ja, sm-kms → unblock their E2E**
- Make jose-ja and sm-kms reuse the same template startup code pattern as cipher-im
- Ensure jose-ja and sm-kms TestMain files use template helper (not raw polling)
- Fix KMS session JWK config blocker (empty algorithm string) using standardized config
- Fix JOSE args routing blocker using standardized routing
- After both services fixed: their E2E startup should unblock in one go
- Update CI E2E workflow (`ci-e2e.yml`) service-specific compose paths

**pki-ca (deferred to research options)**: After this phase, pki-ca will inherit all reliable startup/test code from service-template when it migrates. The CA flag issue is an architectural debt requiring template migration, tracked in research options.

**Success**: cipher-im E2E passes; jose-ja and sm-kms E2E start and run successfully; template has generic startup helper; TestMains use template pattern

### Phase 7: Coverage & Mutation (4h) [Status: ☐ TODO]
**Objective**: Improve coverage and mutation testing **per Q2=E**:
- crypto/jose at 89.9% → push toward structural ceiling (~91%) without interface-wrapping jwx v3
- Document structural ceiling in `docs/fixes-v7/JWX-COV-CEILING.md`: which specific stmts are unreachable and why
- Add `//go:cover-ignore` comments for remaining unreachable error paths after max effort
- Do NOT exempt crypto/jose from ≥98% gate — we are making genuine effort to raise coverage as high as possible without major refactor
- Production packages below 95% (17 packages identified in fixes-v4)
- Run gremlins on all packages meeting ≥95% coverage
- **Success**: crypto/jose reaches ~91% via new tests; JWX-COV-CEILING.md documents remaining ceiling; go:cover-ignore added for genuinely unreachable paths; all production packages ≥95%

### Phase 8: Move cipher-im to SM Product (~4.5h) [Status: ☐ TODO]
**Objective**: Rename cipher-im → sm-im, move under SM product
- Pure mechanical rename — no business logic changes
- Move: `internal/apps/cipher/im/` → `internal/apps/sm/im/`
- Update: cmd/, deployments/, configs/, ARCHITECTURE.md, ci-e2e.yml
- Update all Go import paths referencing cipher/im
- **Detailed plan and tasks**: See [research/plan-PKI-CA-MERGE0a.md](research/plan-PKI-CA-MERGE0a.md) and [research/tasks-PKI-CA-MERGE0a.md](research/tasks-PKI-CA-MERGE0a.md)
- **Success**: cipher-im fully renamed to sm-im, all tests pass, deployments updated, no Cipher product remaining

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Magic consolidation breaks imports | Medium | High | Run build + tests after each move |
| E2E blockers require deep debugging | High | Medium | Focus on one service at a time |
| Flaky test hard to reproduce | Medium | Low | Run with -count=10, check race detector |
| File splitting may break test fixtures | Low | Medium | Verify TestMain patterns preserved |

## Quality Gates - MANDATORY

**Per-Phase Quality Gates**:
- ✅ All tests pass (`go test ./... -shuffle=on`)
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`)
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`)
- ✅ No new TODOs without tracking
- ✅ 65/65 deployment validators pass

**Coverage Targets**:
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code: ≥98% line coverage

**Mutation Testing Targets**:
- ✅ Infrastructure/utility code: ≥98%
- ✅ Production code: ≥95%

## Success Criteria

- [ ] All 8 phases complete with evidence
- [ ] All quality gates passing
- [ ] Zero `//nolint:wsl` (legacy v1) violations
- [ ] All `//nolint:wsl_v5` either fixed by code restructure or documented as structurally required
- [ ] All magic constants in `internal/shared/magic/`
- [ ] All test files have `t.Parallel()`
- [ ] All files ≤500 lines
- [ ] cipher-im E2E passes; jose-ja and sm-kms E2E startup unblocked
- [ ] Template has generic service startup helper for test reuse
- [ ] crypto/jose ≥91%; JWX-COV-CEILING.md documents ceiling; go:cover-ignore for remaining unreachable paths
- [ ] Coverage and mutation targets met
- [ ] cipher-im renamed to sm-im under SM product
- [ ] No Cipher product remaining (SM has kms + im services)
