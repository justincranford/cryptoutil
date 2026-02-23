# Implementation Plan - Consolidated Quality Fixes v7

**Status**: Planning
**Created**: 2026-02-23
**Last Updated**: 2026-02-23
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
- F-6.15: 22 `//nolint:wsl` violations
- F-3.2: TestNegativeDuration not a `time.Duration` type
- F-3.4: `//nolint:stylecheck` without bug reference (5 instances)
- F-1.4: poll.go timeout error not wrapped with sentinel
- F-6.17: pool.go if/else chain → switch statement
- F-6.44: ValidateUUID takes `*string` pointer unnecessarily
- **Success**: All code quality issues resolved, linting clean

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

### Phase 6: E2E Infrastructure (3h) [Status: ☐ TODO]
**Objective**: Fix E2E test infrastructure 
- E2E service startup blockers (KMS session JWK config, JOSE args routing, CA flag issue)
- Task 6.4: Update CI E2E workflow (`ci-e2e.yml` compose paths)
- Task 6.6: Fix identity E2E container names
- fixes-v3 E2E config issues (authz-e2e.yml, idp-e2e.yml as real YAML files)
- **Success**: E2E tests can start and run in Docker environment

### Phase 7: Coverage & Mutation (4h) [Status: ☐ TODO]
**Objective**: Improve coverage and mutation testing
- crypto/jose at 89.9% → push toward structural ceiling (~91%)
- Production packages below 95% (17 packages identified in fixes-v4)
- Run gremlins on all packages meeting ≥95% coverage
- **Success**: Coverage targets met, mutation testing baseline established

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

- [ ] All 7 phases complete with evidence
- [ ] All quality gates passing
- [ ] Zero `//nolint:wsl` violations
- [ ] All magic constants in `internal/shared/magic/`
- [ ] All test files have `t.Parallel()`
- [ ] All files ≤500 lines
- [ ] E2E infrastructure functional
- [ ] Coverage and mutation targets met
