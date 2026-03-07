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

KMS is the exception: `Start()` has no ctx, `Shutdown()` has no error return, `IsReady()` instead of `SetReady()`.

**Approach**:
1. Define `ServiceServer` interface in `internal/apps/template/service/server/contract.go`
2. Interface covers the universal set of methods ALL services share
3. Add compile-time assertion `var _ ServiceServer = (*SkeletonTemplateServer)(nil)` in each service
4. KMS gets special treatment (adaptor or extended interface) until it can be unified
5. Tests verify all services satisfy the interface

**Success**: All 10 services satisfy `ServiceServer` interface at compile time. Adding a new required method forces all services to implement it.

### Phase 2: Simplified Builder Pattern (~2 days) [Status: ☐ TODO]

**Objective**: Remove explicit `WithBarrier()`, `WithJWTAuth()`, `WithStrictServer()` calls from services. Make all standard infrastructure automatic. Services declare ONLY domain-specific add-ons.

**Current state** (from codebase):
- `barrierEnabled := true` is already hardcoded in `server_builder_build.go:90`
- sm-im and jose-ja only call `WithDomainMigrations()` and `WithPublicRouteRegistration()` (minimal)
- sm-kms additionally calls `WithJWTAuth()` and `WithStrictServer()` (it has specialized needs)
- Identity services only call `WithPublicRouteRegistration()` (no domain migrations yet)

**Target state**:
- `NewServerBuilder(ctx, cfg)` auto-configures EVERYTHING (barrier, sessions, TLS, health, realm, registration)
- Services call only: `WithDomainMigrations()` (if they have domain tables), `WithPublicRouteRegistration()` (always, for domain routes), `WithSwaggerUI()` (if enabled)
- `WithJWTAuth()` and `WithStrictServer()` remain for services that NEED non-default behavior (KMS)
- No removal of existing methods — just change defaults so most services don't need them

**Success**: Standard services (sm-im, jose-ja, skeleton, pki-ca, identity-*) need only 2-3 builder calls. Builder auto-configures sensible defaults for all core infrastructure.

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

**NEW fitness functions to add** (in `internal/apps/cicd/lint_fitness/`):
1. **cross-service-import-isolation** — No service package may import another service's internal package (generalize `go-check-identity-imports` to all services)
2. **domain-layer-isolation** — `domain/` must not import `server/`, `client/`, `api/`
3. **file-size-limits** — No file exceeds 500 lines (ARCHITECTURE.md hard limit)
4. **health-endpoint-presence** — All services have health endpoints in their route registration
5. **tls-minimum-version** — All TLS configs use TLS 1.3+
6. **admin-bind-address** — All admin binds use 127.0.0.1 (not 0.0.0.0)
7. **service-contract-compliance** — All services satisfy `ServiceServer` interface (compile-time verified via `var _` assertions, runtime-verified by this linter scanning for the assertion)
8. **migration-range-compliance** — Template migrations 1001-1999, domain migrations 2001+ (extends existing `migration_numbering`)

**CICD Integration** (per ARCHITECTURE.md Section 9.10):
- New command: `lint-fitness`
- Directory: `internal/apps/cicd/lint_fitness/`
- Entry: `Lint(logger)` with `registeredLinters` slice
- Sub-linters: one package per fitness function
- Pre-commit hook: `go run ./cmd/cicd lint-fitness`
- CI workflow: `.github/workflows/ci-fitness.yml`

**Success**: `go run ./cmd/cicd lint-fitness` passes on current codebase. New violations caught at pre-commit time.

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
- A: Remove `WithBarrier()`, `WithJWTAuth()`, `WithStrictServer()` entirely — barrier always on, sessions always on, OpenAPI strict server always on
- B: Keep methods but change defaults — barrier always on (already is), sessions default on, JWT auth default to session mode, strict server default on
- C: Two-tier builder — `NewServerBuilder()` for standard services (everything on), `NewCustomServerBuilder()` for KMS-style services that need opt-out ✓ **SELECTED**
- D: Leave builder as-is, enforce via fitness function that all services call the same With*() methods
- E:

**Decision**: Option C selected — Two-tier builder approach

**Rationale**: Standard services (9 out of 10) get everything automatically. KMS is the only exception that needs custom configuration. Two-tier keeps the escape hatch without cluttering the standard path. The `NewServerBuilder()` method stays the same but auto-configures all standard infrastructure. Services that need to opt-out use explicit methods.

**Impact**: Simplifies standard service `NewFromConfig()` from ~20 builder calls to 2-3. KMS pattern unchanged.

### Decision 2: Fitness Functions Location

**Options**:
- A: New top-level command `fitness-check` (as proposed in brainstorm)
- B: New linter `lint-fitness` following existing naming convention ✓ **SELECTED**
- C: Merge into existing `lint-go` as additional sub-linters
- D: Separate binary `cmd/fitness/main.go`
- E:

**Decision**: Option B selected — `lint-fitness` as new cicd linter command

**Rationale**: Follows established CICD command architecture (Section 9.10): `lint-<target>` naming, `lint_fitness/` directory, `Lint(logger)` entry, `registeredLinters` slice. Consistent with existing 10 linters. Can be independently invoked or included in `all` command.

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
