# Implementation Plan - Framework v1

**Status**: Planning
**Created**: 2026-03-06
**Last Updated**: 2026-03-06
**Purpose**: Formalize the cryptoutil service framework with compile-time contracts, shared test infrastructure, architecture fitness functions, cross-service contract tests, simplified builder pattern, and developer tooling (air live reload).

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL documentation must be accurate and complete
- ✅ **Completeness**: NO steps skipped, NO steps de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING** - ALL issues block progress to next phase or task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

## Overview

This plan implements 5 high-ROI framework improvements from `docs/framework-brainstorm/08-recommendations.md`, adapted to the user's specific architectural preferences:

1. **ServiceContract Interface** (P0-1) — Compile-time enforcement that all 10 services conform to framework patterns
2. **air Live Reload** (P0-2) — One root `.air.toml` with per-service targets for 2-3x faster inner development loop
3. **Simplified Builder Pattern** — Remove unnecessary `WithBarrier()`, `WithJWTAuth()`, `WithStrictServer()` calls; all core services get the same infrastructure by default
4. **Architecture Fitness Functions** (P2-2) — `cicd lint-fitness` command integrated into pre-commit hooks + CI workflow
5. **Shared Test Infrastructure** (P1-4) — Consolidate duplicated TestMain/fixture/assertion patterns across 10 services
6. **Cross-Service Contract Test Suite** (P1-2) — One test suite verifying behavioral consistency across all services

### What This Plan Does NOT Include (Explicit Exclusions)

- ❌ **P0-3: Promote Skeleton to Full CRUD Reference** — Not needed if fitness functions + ServiceContract enforce conformance
- ❌ **P1-1: cicd new-service Scaffolding Tool** — Nice-to-have only; 9 services already exist, adding new ones unlikely
- ❌ **P1-3: cicd diff-skeleton Conformance Tool** — Superseded by fitness functions as pre-commit hooks
- ❌ **P2-1: Service Manifest Declaration** — Replaced by simplified builder defaults (services declare add-ons only)
- ❌ **P2-3: OpenAPI-to-Repository Code Generation** — Not wanted
- ❌ **P3-1: Module System (fx/Wire)** — Overkill for the required level of pluggability
- ❌ **P3-2: Extract Framework as Separate Module** — Premature; all 10 services are internal

## Background

The framework brainstorm (`docs/framework-brainstorm/`) analyzed the current state, cross-language patterns, and Go-specific frameworks. Key findings:

- **What works**: ServerBuilder fluent API, ServiceResources DI, dual HTTPS, barrier-by-default, health automation
- **What hurts**: No compile-time contracts, no cross-service conformance checks, duplicated test infrastructure, the `With*()` calls are explicit where they should be implicit defaults
- **Root cause**: Framework is still partially in "library mode" (you call it) rather than "framework mode" (it calls you, you fill in the blanks)

**User's Core Philosophy**: All product-services MUST have the exact same core infrastructure. The ONLY differences between services are domain-specific add-ons (extra config, migrations, OpenAPI APIs). The current `WithBarrier()`, `WithJWTAuth()`, `WithStrictServer()` calls are overkill — these should all be automatic defaults, and services should only declare what's unique.

## Technical Context

- **Language**: Go 1.25.7
- **Framework**: Service template at `internal/apps/template/service/`
- **Database**: PostgreSQL OR SQLite with GORM
- **Existing Services**: 10 total (sm-im, sm-kms, jose-ja, pki-ca, skeleton-template, identity-authz/idp/rp/rs/spa)
- **Existing CICD**: 13 commands (10 linters, 2 formatters, 1 script) via `cmd/cicd`
- **Pre-commit**: `.pre-commit-config.yaml` with comprehensive hooks
- **Key Files Affected**:
  - `internal/apps/template/service/server/builder/` — Builder simplification
  - `internal/apps/template/service/server/` — ServiceContract interface
  - `internal/apps/cicd/` — New `lint_fitness/` command
  - `internal/apps/template/service/testing/` — Shared test infra expansion
  - `.air.toml` — New file, air live reload config
  - `.pre-commit-config.yaml` — Add fitness check hook

## Phases

### Phase 1: ServiceContract Interface (~3 days) [Status: ☐ TODO]

**Objective**: Define a Go interface that every cryptoutil service MUST implement, providing compile-time enforcement of framework conformance.

**Analysis of existing service methods** (from codebase research):

Common methods across sm-im, jose-ja, skeleton-template, pki-ca:
- `Start(ctx context.Context) error`
- `Shutdown(ctx context.Context) error`
- `DB() *gorm.DB`
- `App() *Application`
- `PublicPort() int`
- `AdminPort() int`
- `SetReady(ready bool)`
- `PublicBaseURL() string`
- `AdminBaseURL() string`
- `PublicServerActualPort() int`
- `AdminServerActualPort() int`

Additional common methods (sm-im, jose-ja, skeleton):
- `JWKGen() *JWKGenService`
- `Telemetry() *TelemetryService`
- `Barrier() *Service` (barrier)

**KMS divergence (from quizme-v1 Q1 analysis — confirmed minimal):**
- Current: `Start() error` → Target: `Start(ctx context.Context) error` (ctx ignored, uses stored s.ctx)
- Current: `Shutdown()` → Target: `Shutdown(ctx context.Context) error` (ctx ignored, always returns nil)
- Current: `IsReady() bool` (getter) → Needs: `SetReady(bool)` (setter) added alongside
- Missing wrappers (delegate to s.resources): `DB()`, `App()`, `JWKGen()`, `Telemetry()`, `PublicServerActualPort()`, `AdminServerActualPort()`
- 1 call site update in `internal/apps/sm/kms/kms.go` (add ctx arg to Start/Shutdown calls)
- **Decision (quizme-v1 Q1)**: C — Unify KMS (changes confirmed minimal)

**Approach**:
1. Define `ServiceServer` interface in `internal/apps/template/service/server/contract.go`
2. Interface covers the universal set (Start, Shutdown, DB, App, PublicPort, AdminPort, SetReady, PublicBaseURL, AdminBaseURL, PublicServerActualPort, AdminServerActualPort)
3. Add compile-time assertion `var _ ServiceServer = (*SkeletonTemplateServer)(nil)` in all 10 services
4. KMS gets 8 new/modified methods to achieve full conformance (see above)
5. Tests verify all services satisfy the interface

**Success**: All 10 services satisfy `ServiceServer` interface at compile time. Adding a new required method forces all services to implement it.

### Phase 2: Simplified Builder Pattern (~2 days) [Status: ☐ TODO]

**Objective**: Remove ALL explicit `WithBarrier()`, `WithJWTAuth()`, `WithStrictServer()` calls from services. `Build()` auto-configures EVERYTHING. Services declare ONLY domain-specific add-ons.

**Decision (quizme-v1 Q2)**: A — Aggressive. Only `WithDomainMigrations()` and `WithPublicRouteRegistration()` survive as standard service calls. JWTAuth defaults to session mode. StrictServer is auto-configured with default paths from settings.

**Current state** (from codebase):
- `barrierEnabled := true` is already hardcoded in `server_builder_build.go:90`
- sm-im and jose-ja only call `WithDomainMigrations()` and `WithPublicRouteRegistration()` (already close)
- sm-kms calls `WithJWTAuth()` and `WithStrictServer()` — these become auto-configured
- Identity services only call `WithPublicRouteRegistration()` (no domain migrations yet)

**Target state (aggressive)**:
- `NewServerBuilder(ctx, cfg)` auto-configures EVERYTHING: barrier, sessions, TLS, health, realm, registration, JWTAuth (session mode default), StrictServer (paths from settings)
- Services call ONLY: `WithDomainMigrations()` (if domain tables), `WithPublicRouteRegistration()` (always)
- No explicit `WithJWTAuth()` or `WithStrictServer()` calls in any service
- Existing `With*()` methods stay for backward compat but are no longer called by services
- `Build()` internally calls equivalent of `WithJWTAuth(NewDefaultJWTAuthConfig())` + `WithStrictServer(NewDefaultStrictServerConfig().WithPaths(settings))` if not already set

**Structural approach**: Track a `configured` flag per With*() call. `Build()` applies defaults for any not-yet-configured aspects.

**Success**: Every standard service's `NewFromConfig` is ≤10 lines. Only 2 builder calls: `WithDomainMigrations()` (optional) + `WithPublicRouteRegistration()`.

### Phase 3: air Live Reload (~0.5 days) [Status: ☐ TODO]

**Objective**: Configure `air` for hot reload during development with one root `.air.toml` that supports multiple service targets.

**Approach**:
- Single `.air.toml` at project root with configurable service target
- Default: `go build -o tmp/main ./cmd/${SERVICE}/` where `${SERVICE}` is set via env var
- Watch: `internal/`, `cmd/`, `pkg/`, `api/`
- Exclude: `_test.go`, `tmp/`, `test-output/`, `docs/`, `.git/`
- Usage: `SERVICE=sm-im air` or `SERVICE=jose-ja air`

**Success**: `SERVICE=sm-im air` watches code changes and auto-rebuilds in <3s.

### Phase 4: Architecture Fitness Functions (~2 weeks) [Status: ☐ TODO]

**Objective**: Automated enforcement of ARCHITECTURE.md constraints as a `cicd lint-fitness` command, integrated into pre-commit hooks and CI.

**Key insight**: Many fitness functions ALREADY exist in the current cicd linters:
- `lint-go/circular_deps` → Dependency isolation (partial)
- `lint-go/cgo_free_sqlite` → Security constraint
- `lint-go/crypto_rand` → Security constraint
- `lint-go/insecure_skip_verify` → Security constraint
- `lint-go/non_fips_algorithms` → Security constraint
- `lint-go/migration_numbering` → Framework versioning
- `lint-go/product_structure` → Service structure
- `lint-go/service_structure` → Service structure
- `lint-gotest/bind_address_safety` → Test quality (port 0)
- `lint-gotest/parallel_tests` → Test quality (t.Parallel)
- `lint-gotest/no_hardcoded_passwords` → Security
- `lint-skeleton/check_skeleton_placeholders` → Service conformance

**Decision (quizme-v1 Q3)**: A — Full migration. ALL architecture-enforcement checks move from lint_go/lint_gotest/lint_skeleton to lint_fitness. lint_go/lint_gotest keep only Go language quality checks. lint_skeleton is dissolved.

**Complete inventory: 23 sub-linters total (15 migrated + 8 new)**

**Existing checks MIGRATING to lint_fitness** (from lint_go):
1. `cgo_free_sqlite` — Architecture: CGO ban enforcement
2. `circular_deps` — Architecture: dependency isolation
3. `cmd_main_pattern` — Architecture: thin main() pattern
4. `crypto_rand` — Security/FIPS: use crypto/rand, not math/rand
5. `insecure_skip_verify` — Security: TLS hardening
6. `migration_numbering` — Architecture: migration file naming
7. `non_fips_algorithms` — Security/FIPS: banned algorithm detection
8. `product_structure` — Architecture: service directory layout
9. `product_wiring` — Architecture: service wiring patterns
10. `service_structure` — Architecture: service layer patterns

**Existing checks MIGRATING to lint_fitness** (from lint_gotest):
11. `bind_address_safety` — Architecture: port 0 in tests, loopback binding
12. `no_hardcoded_passwords` — Security: no secrets in tests
13. `parallel_tests` — Architecture: t.Parallel() usage
14. `test_patterns` — Architecture: table-driven test patterns

**Existing checks MIGRATING to lint_fitness** (from lint_skeleton — dissolves):
15. `check_skeleton_placeholders` — Architecture: placeholder detection in new services

**NEW fitness sub-linters** (in `internal/apps/cicd/lint_fitness/`):
16. `cross_service_import_isolation` — No service package imports another service's internal package
17. `domain_layer_isolation` — `domain/` must not import `server/`, `client/`, `api/`
18. `file_size_limits` — No file exceeds 500 lines (ARCHITECTURE.md Section 11.2.6 hard limit)
19. `health_endpoint_presence` — All services register health endpoints
20. `tls_minimum_version` — All TLS configs use TLS 1.3+
21. `admin_bind_address` — All admin binds use 127.0.0.1 (not 0.0.0.0)
22. `service_contract_compliance` — All services have `var _ ServiceServer = (*XxxServer)(nil)` assertion
23. `migration_range_compliance` — Template 1001-1999, domain 2001+ (extends migrated migration_numbering)

**Checks REMAINING in lint_go** (language quality only — 7 checks):
- `leftover_coverage`, `magic_aliases`, `magic_duplicates`, `magic_usage`, `no_unaliased_cryptoutil_imports`, `test_presence`, `common`

**Checks REMAINING in lint_gotest** (language quality only — 2 checks):
- `require_over_assert`, `common`

**CICD Integration** (per ARCHITECTURE.md Section 9.10):
- New command: `lint-fitness`
- Directory: `internal/apps/cicd/lint_fitness/`
- Entry: `Lint(logger)` with `registeredLinters` slice
- Sub-linters: one package per fitness function
- Pre-commit hook: `go run ./cmd/cicd lint-fitness`
- Migration order: New checks first, then existing checks migrate

**Success**: `go run ./cmd/cicd lint-fitness` passes on current codebase. New violations caught at pre-commit time. lint_skeleton command removed from all invocation paths.

### Phase 5: Shared Test Infrastructure (~1 week) [Status: ☐ TODO]

**Objective**: Consolidate duplicated test setup patterns into a shared package that all services import.

**Current duplication** (from codebase research):
- `internal/apps/template/service/testing/e2e_helpers/` — 7 files (config, DB, HTTP, server, auth helpers)
- `internal/apps/template/service/testing/e2e_infra/` — 3 files (compose manager, docker health)
- `internal/apps/template/service/testutil/` — 1 file (HTTP test helpers)
- `internal/apps/template/service/server/testutil/` — 3 files (database, helpers, mocks)
- `internal/apps/sm/im/testing/` — 2 files (testmain helper)
- Each service has its own TestMain with ~50 lines of setup code

**Target**: Expand `internal/apps/template/service/testing/` to provide:
1. `NewInMemorySQLiteDB(t)` — Quick unit test DB
2. `NewPostgresTestContainer(ctx, t)` — Integration test DB
3. `NewTestServer(t, cfg)` — Standardized server setup with port 0
4. `StandardFixtures` — Create tenant, realm, user fixtures
5. `StandardAssertions` — HTTP response validation, error format checking
6. `HealthClient` — Reusable client for health endpoint contract tests

**Success**: New service's TestMain is <10 lines. Test helper changes propagate to all services automatically.

### Phase 6: Cross-Service Contract Test Suite (~1 week) [Status: ☐ TODO]

**Objective**: One test suite that verifies ALL services behave consistently for core framework behavior.

**Depends on**: Phase 1 (ServiceContract interface), Phase 5 (shared test infra)

**Contract tests** (run against each service that satisfies `ServiceServer`):
1. `/admin/api/v1/livez` returns 200 OK
2. `/admin/api/v1/readyz` returns 200 OK when server is ready, 503 when not
3. `/admin/api/v1/shutdown` triggers graceful shutdown
4. `/service/api/v1/*` rejects unauthenticated requests with 401
5. `/browser/api/v1/*` rejects cross-origin requests appropriately
6. Error responses always contain `code` + `message` fields
7. All services include `trace_id` in error responses
8. `/browser/api/v1/health` and `/service/api/v1/health` return 200

**Approach**:
- Package: `internal/apps/template/service/testing/contract/`
- Test function: `RunContractTests(t *testing.T, server ServiceServer)`
- Each service's integration test calls `RunContractTests` with its server instance
- Table-driven subtests for each contract
- Uses shared test infrastructure from Phase 5

**Success**: Adding a new contract test automatically tests ALL services. Behavioral divergence caught in CI.

### Phase 7: Quality Gates & Evidence (~2 days) [Status: ☐ TODO]

**Objective**: Verify all phases meet quality gates, collect evidence, commit.

- All tests pass (`go test ./...`)
- Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`)
- Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`)
- Coverage ≥95% (production), ≥98% (infrastructure/utility)
- Mutation testing ≥95% minimum
- No new TODOs without tracking
- Race detector clean (`go test -race -count=2 ./...`)
- Pre-commit hooks pass (`pre-commit run --all-files`)
- Fitness functions pass (`go run ./cmd/cicd lint-fitness`)

**Success**: All quality gates pass with evidence collected in `test-output/framework-v1/`.

## Executive Decisions

### Decision 1: Builder Simplification Strategy

**Options**:
- A: Aggressive — `Build()` auto-configures everything. Only `WithDomainMigrations()` and `WithPublicRouteRegistration()` survive. JWTAuth defaults to session mode. StrictServer auto-configured from settings. ✓ **SELECTED**
- B: Keep methods but change defaults — barrier always on (already is), sessions default on, JWT auth default to session mode, strict server default on
- C: Two-tier builder — `NewServerBuilder()` for standard services (everything on), `NewCustomServerBuilder()` for KMS-style services that need opt-out
- D: Leave builder as-is, enforce via fitness function that all services call the same With*() methods
- E:

**Decision**: Option A selected — Aggressive simplification (quizme-v1 Q2, Answer A)

**Rationale**: User's core philosophy: all services have identical core. No service should need to think about infrastructure configuration. `Build()` applies all defaults. `configured` flags in builder prevent double-application. Existing `With*()` methods kept for backward compat but no service calls them.

**Impact**: Every standard service's `NewFromConfig` reduces to 1-2 builder calls. KMS aligns with standard after Q1 unification.

### Decision 2: Fitness Functions Migration Strategy

**Options**:
- A: Full migration — ALL architecture-enforcement checks move from lint_go/lint_gotest/lint_skeleton to lint_fitness. lint_skeleton dissolved. ✓ **SELECTED**
- B: Dual home — leave existing checks in place, add only new checks to lint_fitness
- C: Reference only — lint_fitness calls into lint_go/lint_gotest for shared implementation
- D: New only — lint_fitness contains ONLY checks that don't exist yet
- E:

**Decision**: Option A selected — Full migration (quizme-v1 Q3, Answer A)

**Rationale**: Architecture fitness checks belong together in one place. Having them split across lint_go, lint_gotest, lint_skeleton is confusing. Full migration provides clean separation: lint_go = Go language quality, lint_fitness = architecture constraints. lint_skeleton dissolves (its 1 check migrates).

**Implementation order**: New checks added first (Phase 4 tasks 4.3-4.10), existing checks migrate after (Phase 4 tasks 4.13-4.15). Pre-commit hooks updated at end (Phase 4 task 4.11).

### Decision 3: ServiceContract Interface Location

**Options**:
- A: New package `internal/apps/template/service/framework/contract.go`
- B: Existing package `internal/apps/template/service/server/contract.go` ✓ **SELECTED**
- C: New package `internal/apps/template/service/contract/contract.go`
- D: Shared package `internal/shared/framework/contract.go`
- E:

**Decision**: Option B selected — In existing server package

**Rationale**: The interface references types from the server package (`Application`, `PublicServerBase`). Placing it in the same package avoids circular imports. All services already import this package. No new package needed.

### Decision 4: Skeleton/Scaffolding (P0-3 / P1-1)

**Options**:
- A: Implement both P0-3 (skeleton CRUD) and P1-1 (scaffolding tool)
- B: Implement P0-3 only (skeleton as reference) without scaffolding
- C: Skip both — rely on ServiceContract + fitness functions for enforcement ✓ **SELECTED**
- D: Defer to future framework version after this work validates the approach
- E:

**Decision**: Option C selected — Skip, enforce via contract + fitness functions

**Rationale**: User's 9 services already exist. ServiceContract provides compile-time enforcement. Fitness functions provide pre-commit enforcement. Contract tests provide behavioral enforcement. Adding new services is a nice-to-have only. The combination of Phases 1, 4, 5, and 6 provides stronger conformance guarantees than a reference implementation alone.

### Decision 5: KMS Interface Compliance Strategy

**Options**:
- A: Adapter pattern — `KMSAdapter` wrapper forwards calls with expected signatures
- B: Two interfaces — `ServiceServer` (core) + `ServiceServerWithContext` (extended)
- C: Unify KMS — Modify KMS to match standard signatures ✓ **SELECTED**
- D: Exclude KMS from ServiceServer contract entirely
- E:

**Decision**: Option C selected — Unify KMS (quizme-v1 Q1, Answer C; confirmed minimal after analysis)

**KMS changes required** (all minimal, 1 file each):
1. `Start(ctx context.Context) error` — add ctx param (ctx stored as s.ctx, ignored)
2. `Shutdown(ctx context.Context) error` — add ctx + error return (return nil)
3. `SetReady(bool)` — add setter alongside existing `IsReady()` getter
4. Delegate methods (7 additions): `DB()`, `App()`, `JWKGen()`, `Telemetry()`, `PublicServerActualPort()`, `AdminServerActualPort()`
5. Update 1 call site in `internal/apps/sm/kms/kms.go`

**Rationale**: Changes are truly minimal — no architectural change to KMS. KMS internal behavior unchanged; only the external method signatures conform to the standard contract.

### Decision 6: Shared Test Infrastructure Migration Scope

**Options**:
- A: All 10 services
- B: Core 3 (sm-im, jose-ja, skeleton-template)
- C: Template + 1
- D: Create only — no migration
- E: Core 4 — sm-im, jose-ja, sm-kms, skeleton-template ✓ **SELECTED**

**Decision**: Option E selected — Core 4 services (quizme-v1 Q4, Answer E)

**Rationale**: sm-kms included because it benefits from KMS unification (Phase 1) and shared DB helpers. skeleton-template is the reference. sm-im is the most feature-complete. jose-ja validates the JWK-heavy path.

### Decision 7: Contract Test Depth

**Options**:
- A: Shallow (6-8 tests, health only)
- B: Medium (12-15 tests, health + auth)
- C: Deep (20+ tests, health + auth + domain patterns) ✓ **SELECTED**
- D: Progressive (shallow now, more later)
- E:

**Decision**: Option C selected — Deep 21+ contracts (quizme-v1 Q5, Answer C)

**Rationale**: The contract test framework itself is the main investment. Writing 21 contracts vs 8 is marginal additional effort once the framework is in place. Deep contracts catch the most divergence and provide full validation of framework behavioral guarantees.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| KMS builder divergence breaks contract | Medium | High | Phase 1 includes KMS-specific interface; adapter pattern if needed |
| Fitness functions false positives | Medium | Medium | Conservative initial rules; error vs warning severity levels |
| Shared test infra breaks existing tests | Low | High | Incremental migration; existing tests continue to work during transition |
| air.toml conflicts with CI/CD | Low | Low | air is dev-only; CI uses `go build` directly |
| Cross-service tests too slow | Medium | Medium | Use SQLite in-memory, port 0, `app.Test()` — no real servers needed |

## Quality Gates - MANDATORY

**Per-Action Quality Gates**:
- ✅ All tests pass (`go test ./...`) - 100% passing, zero skips
- ✅ Build clean (`go build ./...` AND `go build -tags e2e,integration ./...`) - zero errors
- ✅ Linting clean (`golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`) - zero warnings
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets (from copilot instructions)**:
- ✅ Production code: ≥95% line coverage
- ✅ Infrastructure/utility code: ≥98% line coverage
- ✅ main() functions: 0% acceptable if internalMain() ≥95%
- ✅ Generated code: Excluded from coverage (OpenAPI stubs, GORM models, protobuf)

**Mutation Testing Targets (from copilot instructions)**:
- ✅ Infrastructure/utility code: ≥98% (NO EXCEPTIONS)

**Per-Phase Quality Gates**:
- ✅ Unit + integration tests complete before moving to next phase
- ✅ Race detector clean (`go test -race -count=2 ./...`)
- ✅ Deployment validators pass (when deployments/ or configs/ changed)

**Overall Project Quality Gates**:
- ✅ All phases complete with evidence
- ✅ All test categories passing (unit, integration)
- ✅ Coverage and mutation targets met
- ✅ CI/CD workflows green
- ✅ Documentation updated

## Success Criteria

- [ ] All 10 services satisfy `ServiceServer` interface (compile-time verified)
- [ ] Standard services use ≤3 builder calls (down from 5-7)
- [ ] `SERVICE=sm-im air` provides <3s rebuild loop
- [ ] `go run ./cmd/cicd lint-fitness` catches 8+ architectural constraint categories
- [ ] Pre-commit hook runs fitness checks on every commit
- [ ] Shared test infrastructure eliminates 300+ lines of duplicated TestMain code
- [ ] Contract test suite covers 8+ behavioral contracts across all services
- [ ] All quality gates passing
- [ ] Evidence archived in `test-output/framework-v1/`
- [ ] Evidence archived (test output, logs, analysis)

## ARCHITECTURE.md Cross-References - MANDATORY

| Topic | ARCHITECTURE.md Section | When Referenced |
|-------|------------------------|----------------|
| Testing Strategy | [Section 10](../../docs/ARCHITECTURE.md#10-testing-architecture) | Phases 5, 6, 7 |
| Unit Testing | [Section 10.2](../../docs/ARCHITECTURE.md#102-unit-testing-strategy) | All phases with tests |
| Coverage Ceiling | [Section 10.2.3](../../docs/ARCHITECTURE.md#1023-coverage-targets) | Phase 7 quality gates |
| Integration Testing | [Section 10.3](../../docs/ARCHITECTURE.md#103-integration-testing-strategy) | Phase 6 contract tests |
| Mutation Testing | [Section 10.5](../../docs/ARCHITECTURE.md#105-mutation-testing-strategy) | Phase 7 quality gates |
| Quality Gates | [Section 11.2](../../docs/ARCHITECTURE.md#112-quality-gates) | ALL phases (mandatory) |
| Code Quality | [Section 11.3](../../docs/ARCHITECTURE.md#113-code-quality-standards) | All phases with new code |
| Coding Standards | [Section 13.1](../../docs/ARCHITECTURE.md#131-coding-standards) | ALL phases with implementation |
| Version Control | [Section 13.2](../../docs/ARCHITECTURE.md#132-version-control) | ALL phases (commit strategy) |
| CICD Command Architecture | [Section 9.10](../../docs/ARCHITECTURE.md#910-cicd-command-architecture) | Phase 4 (fitness functions) |
| Pre-Commit Hooks | [Section 9.9](../../docs/ARCHITECTURE.md#99-pre-commit-hook-architecture) | Phase 4 |
| Service Template | [Section 5.1](../../docs/ARCHITECTURE.md#51-service-template-pattern) | Phases 1, 2 |
| Service Builder | [Section 5.2](../../docs/ARCHITECTURE.md#52-service-builder-pattern) | Phase 2 |
| Infrastructure Blockers | [Section 13.7](../../docs/ARCHITECTURE.md#137-infrastructure-blocker-escalation) | All phases |
| Plan Lifecycle | [Section 13.6](../../docs/ARCHITECTURE.md#136-plan-lifecycle-management) | ALL phases (mandatory) |
| Security Architecture | [Section 6](../../docs/ARCHITECTURE.md#6-security-architecture) | Phase 4 (security fitness functions) |
| File Size Limits | [Section 11.2.6](../../docs/ARCHITECTURE.md#1126-file-size-limits) | Phase 4 (file size fitness function) |
