# Tasks - Consolidated Quality Fixes v7

**Status**: 0 of 47 tasks complete (0%)
**Last Updated**: 2026-02-23
**Created**: 2026-02-23

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark phases or tasks complete without objective evidence

**ALL issues are blockers - NO exceptions.**

---

## Task Checklist

### Phase 1: Critical Bug Fixes

**Phase Objective**: Fix actual bugs that affect correctness

#### Task 1.1: poll.go — Add nil conditionFn guard
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.1
- **Description**: `poll.Until()` panics if `conditionFn` is nil. Add nil guard at function entry.
- **Acceptance Criteria**:
  - [ ] Guard returns error if conditionFn is nil
  - [ ] Test added for nil conditionFn
  - [ ] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.2: poll.go — Add zero/negative timeout and interval validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.2, F-1.3
- **Description**: `poll.Until()` accepts zero/negative timeout and interval without error.
- **Acceptance Criteria**:
  - [ ] Guard returns error if timeout ≤ 0
  - [ ] Guard returns error if interval ≤ 0
  - [ ] Tests added for both cases
  - [ ] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.3: poll.go — Check context before first conditionFn call
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.5
- **Description**: Context cancellation not checked before first `conditionFn` call.
- **Acceptance Criteria**:
  - [ ] Context checked at function entry
  - [ ] Test for pre-canceled context
  - [ ] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.4: poll.go — Wrap timeout error with sentinel
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.4
- **Description**: Timeout error is plain `fmt.Errorf`, not wrapped with a sentinel error for `errors.Is()`.
- **Acceptance Criteria**:
  - [ ] Define `ErrTimeout` sentinel
  - [ ] Wrap timeout error with sentinel
  - [ ] Test using `errors.Is(err, poll.ErrTimeout)`
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.5: ValidateUUIDs — Fix wrong error wrapping
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.4
- **Description**: `ValidateUUIDs()` wraps the wrong error variable.
- **Acceptance Criteria**:
  - [ ] Error wrapping uses correct variable
  - [ ] Test verifies correct error message
  - [ ] Build passes
- **Files**: `internal/shared/util/random/uuid.go`, `internal/shared/util/random/uuid_slice_cache_test.go`

#### Task 1.6: Flaky test — TestAuditLogService_LogOperation_AuditDisabled
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: verified 2026-02-23 test run
- **Description**: Test fails intermittently under `go test ./... -shuffle=on` but passes individually. Likely race condition or shared TestMain state.
- **Acceptance Criteria**:
  - [ ] Root cause identified
  - [ ] Fix applied (isolation, synchronization, or require.Eventually)
  - [ ] Passes with `-count=5 -shuffle=on`
- **Files**: `internal/apps/jose/ja/service/*_test.go`

#### Task 1.7: Verify F-6.5 copy-paste bug
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.5
- **Description**: "sqlite" string reference in PostgreSQL function — verify if actual bug or false positive.
- **Acceptance Criteria**:
  - [ ] Investigate and document whether this is a real bug
  - [ ] Fix if real, close if false positive
- **Files**: TBD (investigate)

---

### Phase 2: Code Quality & Standards

**Phase Objective**: Fix style/standards violations that don't affect correctness but violate project standards

#### Task 2.1: Rename file with space — usernames_passwords_test util.go
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.9
- **Description**: File has space in name: `internal/shared/util/random/usernames_passwords_test util.go`. Go tooling may not pick it up.
- **Acceptance Criteria**:
  - [ ] File renamed to remove space
  - [ ] Build passes
  - [ ] Tests pass
- **Files**: `internal/shared/util/random/usernames_passwords_test util.go`

#### Task 2.2: Error sentinels — change from string to error type
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.8
- **Description**: Error sentinels are typed as `string` instead of `error`. Change type and update callers.
- **Acceptance Criteria**:
  - [ ] Sentinel errors use `errors.New()` or custom error type
  - [ ] All callers updated
  - [ ] Tests pass
- **Files**: TBD (investigate affected files)

#### Task 2.3: Remove //nolint:wsl violations (22 instances)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Dependencies**: None
- **Source**: fixes-v5 F-6.15, verified 22 instances
- **Description**: Per coding instructions: "NEVER use `//nolint:wsl`". Restructure code to eliminate all 22 instances.
- **Acceptance Criteria**:
  - [ ] Zero `//nolint:wsl` or `//nolint:wsl_v5` in codebase
  - [ ] Linting passes
  - [ ] Code restructured (not suppressed)
- **Files**: Multiple (22 files, see grep results)

#### Task 2.4: TestNegativeDuration — change to time.Duration type
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v5 F-3.2
- **Description**: `TestNegativeDuration = -1` is untyped int, should be `time.Duration`.
- **Acceptance Criteria**:
  - [ ] Changed to `TestNegativeDuration = -1 * time.Nanosecond` or similar
  - [ ] All callers work correctly
  - [ ] Tests pass
- **Files**: `internal/shared/magic/magic_testing.go`

#### Task 2.5: nolint:stylecheck — add bug reference (5 instances)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v5 F-3.4
- **Description**: `//nolint:stylecheck` without GitHub issue reference. Per instructions, only allowed with documented linter bug reference.
- **Acceptance Criteria**:
  - [ ] Each `//nolint:stylecheck` has bug reference OR is removed by fixing the style issue
  - [ ] Linting passes
- **Files**: `internal/shared/magic/magic_testing.go`

#### Task 2.6: pool.go — Convert if/else chain to switch
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.17
- **Description**: `validateConfig` in pool.go has 12-branch if/else chain. Per coding instructions: prefer switch statements.
- **Acceptance Criteria**:
  - [ ] Converted to switch statement
  - [ ] Tests pass
  - [ ] Same behavior
- **Files**: `internal/shared/pool/pool.go`

#### Task 2.7: ValidateUUID — change *string to string parameter
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: Task 1.5
- **Source**: fixes-v6 F-6.44
- **Description**: `ValidateUUID(uuid *googleUuid.UUID, msg *string)` — `msg` parameter should be value, not pointer.
- **Acceptance Criteria**:
  - [ ] Changed `msg *string` to `msg string`
  - [ ] All callers updated
  - [ ] Tests pass
- **Files**: `internal/shared/util/random/uuid.go`, callers

#### Task 2.8: fmt.Errorf without %w audit
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.39
- **Description**: Audit all `fmt.Errorf` calls to ensure errors are wrapped with `%w` where appropriate.
- **Acceptance Criteria**:
  - [ ] Audit complete
  - [ ] All appropriate errors wrapped with `%w`
  - [ ] Tests pass
- **Files**: Multiple (audit results)

---

### Phase 3: Magic Constant Consolidation

**Phase Objective**: Move all domain-specific magic constants to `internal/shared/magic/`

#### Task 3.1: Move identity magic package to shared/magic
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.10
- **Description**: `internal/apps/identity/magic/` has magic_uris.go, magic_scopes.go, magic_adaptive.go, magic_metrics.go. Move to `internal/shared/magic/magic_identity*.go`.
- **Acceptance Criteria**:
  - [ ] Constants moved to `internal/shared/magic/`
  - [ ] Old package deleted
  - [ ] All imports updated
  - [ ] Build + tests pass
- **Files**: `internal/apps/identity/magic/*` → `internal/shared/magic/magic_identity*.go`

#### Task 3.2: Move PKI CA magic package to shared/magic
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.11
- **Description**: `internal/apps/pki/ca/magic/magic.go` (28 lines). Move to `internal/shared/magic/magic_pki_ca.go`.
- **Acceptance Criteria**:
  - [ ] Constants moved to `internal/shared/magic/`
  - [ ] Old package deleted
  - [ ] All imports updated
  - [ ] Build + tests pass
- **Files**: `internal/apps/pki/ca/magic/magic.go` → `internal/shared/magic/magic_pki_ca.go`

#### Task 3.3: Move identity config magic to shared/magic
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.12
- **Description**: `internal/apps/identity/config/magic.go` (62 lines). Move to `internal/shared/magic/`.
- **Acceptance Criteria**:
  - [ ] Constants moved to `internal/shared/magic/`
  - [ ] Old file deleted
  - [ ] All imports updated
  - [ ] Build + tests pass
- **Files**: `internal/apps/identity/config/magic.go` → `internal/shared/magic/`

#### Task 3.4: Move TLS constants to shared/magic
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.25
- **Description**: TLS-related constants scattered outside shared/magic.
- **Acceptance Criteria**:
  - [ ] TLS constants moved to `internal/shared/magic/magic_tls.go`
  - [ ] All imports updated
  - [ ] Build + tests pass
- **Files**: TBD (investigate scattered TLS constants)

#### Task 3.5: Consolidate demo package scattered constants
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v5 F-4.2, fixes-v6 F-6.36
- **Description**: Demo packages have 20+ scattered constants and duplicates. Consolidate to shared/magic.
- **Acceptance Criteria**:
  - [ ] Demo constants moved to `internal/shared/magic/magic_demo.go`
  - [ ] No duplicate constants
  - [ ] All imports updated
  - [ ] Build + tests pass
- **Files**: `internal/apps/*/demo/*.go` → `internal/shared/magic/magic_demo.go`

---

### Phase 4: Test Quality Improvements

**Phase Objective**: Fix test compliance issues

#### Task 4.1: Add t.Parallel() to 35 test files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: None
- **Source**: fixes-v6 F-6.16
- **Description**: 35 test files missing `t.Parallel()` on test functions and/or subtests.
- **Acceptance Criteria**:
  - [ ] All test functions have `t.Parallel()`
  - [ ] All subtests have `t.Parallel()`
  - [ ] Tests pass with `-shuffle=on`
  - [ ] No race conditions introduced
- **Files**: 35 files (see fixes-v6 F-6.16 list)

#### Task 4.2: Replace time.Sleep in PKI server_highcov_test.go
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v5 F-2.11
- **Description**: 4 `time.Sleep` calls in `internal/apps/pki/ca/server/server_highcov_test.go`. Replace with `require.Eventually` or `poll.Until`.
- **Acceptance Criteria**:
  - [ ] All 4 time.Sleep replaced
  - [ ] Tests pass
  - [ ] No flakiness
- **Files**: `internal/apps/pki/ca/server/server_highcov_test.go`

#### Task 4.3: Replace time.Sleep in demo files
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v5 F-2.12-2.14
- **Description**: Demo files use `time.Sleep` for server startup. Replace with polling/readiness checks.
- **Acceptance Criteria**:
  - [ ] Demo time.Sleep calls replaced with readiness polling
  - [ ] Tests pass
- **Files**: `internal/apps/identity/demo/demo.go`, `internal/apps/identity/demo/orchestration_test.go`

#### Task 4.4: Split 4 test files exceeding 500-line limit
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.13
- **Description**: 4 test files exceed 500-line hard limit: `tenant_test.go` (519), `businesslogic_crud_test.go` (514), `oam_orm_mapper_test.go` (506), `tls_error_paths_test.go` (504).
- **Acceptance Criteria**:
  - [ ] All 4 files split to ≤500 lines
  - [ ] Tests pass after split
  - [ ] TestMain patterns preserved
- **Files**: Listed above

#### Task 4.5: Add tests for zero-coverage packages
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: None
- **Source**: fixes-v6 F-6.14, F-6.32, F-6.33
- **Description**: Multiple packages have zero tests: identity server/cmd, identity/rp, identity/spa, pki/ca/domain.
- **Acceptance Criteria**:
  - [ ] At least basic smoke tests for each package
  - [ ] Coverage >0% for all packages
- **Files**: Multiple identity and PKI packages

#### Task 4.6: poll_test.go — Add edge case tests
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 20min
- **Dependencies**: Tasks 1.1-1.4
- **Source**: fixes-v5 F-1.6
- **Description**: After fixing poll.go guards, add edge case tests: nil conditionFn, zero timeout, negative interval, pre-canceled context, immediate success, immediate failure.
- **Acceptance Criteria**:
  - [ ] 7+ new test cases
  - [ ] Table-driven pattern
  - [ ] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll_test.go`

---

### Phase 5: Dependency & Architecture

**Phase Objective**: Fix architectural violations

#### Task 5.1: Fix shared → apps/template dependency inversion
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.7
- **Description**: Shared packages import from `apps/template`, violating dependency direction.
- **Acceptance Criteria**:
  - [ ] No imports from shared → apps/template
  - [ ] Interface extraction or code relocation used
  - [ ] Build passes
- **Files**: TBD (investigate)

#### Task 5.2: Identity poller.go — Use shared poll.Until
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 20min
- **Dependencies**: Tasks 1.1-1.4
- **Source**: fixes-v5 F-4.1
- **Description**: `internal/apps/identity/healthcheck/poller.go` has duplicate polling logic. Replace with `poll.Until()`.
- **Acceptance Criteria**:
  - [ ] poller.go uses poll.Until
  - [ ] Tests pass
  - [ ] No duplicate polling code
- **Files**: `internal/apps/identity/healthcheck/poller.go`

#### Task 5.3: Remove blanket nolint suppressions
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.34
- **Description**: `//nolint:wrapcheck,thelper` blanket suppressions should be replaced with proper error handling.
- **Acceptance Criteria**:
  - [ ] Blanket nolint removed
  - [ ] Proper error wrapping/test helpers used
  - [ ] Linting passes
- **Files**: TBD (investigate)

#### Task 5.4: Clean up unused sentinel errors
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.37
- **Description**: Unused sentinel errors in `database/sharding.go`.
- **Acceptance Criteria**:
  - [ ] Unused sentinels removed
  - [ ] Build passes
- **Files**: `internal/shared/database/sharding.go`

#### Task 5.5: SQL interpolation defense in sharding
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 20min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.38
- **Description**: SQL interpolation in sharding lacks defense-in-depth validation.
- **Acceptance Criteria**:
  - [ ] Input validation added before SQL interpolation
  - [ ] Tests for SQL injection attempts
- **Files**: `internal/shared/database/sharding.go`

---

### Phase 6: E2E Infrastructure

**Phase Objective**: Fix E2E test infrastructure to enable service startup

#### Task 6.1: Fix KMS session JWK algorithm configuration
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v1 blocker 02
- **Description**: KMS services fail with "unsupported JWS algorithm:" — empty algorithm string in session manager config.
- **Acceptance Criteria**:
  - [ ] Session JWK algorithm configured correctly
  - [ ] Service starts successfully
- **Files**: Config files, session manager config

#### Task 6.2: Fix JOSE args routing
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: fixes-v1 blocker 02
- **Description**: JOSE service args routing incorrect — args not stripped properly through product → service → subcommand layers.
- **Acceptance Criteria**:
  - [ ] Args routing matches cipher-im pattern
  - [ ] Service starts successfully
- **Files**: `internal/apps/jose/ja/*.go`

#### Task 6.3: Fix CA service --config flag issue
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 20min
- **Dependencies**: None
- **Source**: fixes-v1 blocker 02
- **Description**: CA service reports "unknown flag: --config". May be entrypoint issue.
- **Acceptance Criteria**:
  - [ ] Root cause identified
  - [ ] Fix applied or documented as non-issue
- **Files**: `internal/apps/pki/ca/*.go`, compose entrypoint

#### Task 6.4: Update CI E2E workflow paths
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 20min
- **Dependencies**: None
- **Source**: implementation-plan-v1 Task 6.4
- **Description**: `ci-e2e.yml` references old compose paths (deployments/ca → pki-ca, etc.).
- **Acceptance Criteria**:
  - [ ] All compose file paths corrected
  - [ ] Workflow uses correct deployment directory names
- **Files**: `.github/workflows/ci-e2e.yml`

#### Task 6.5: Fix identity E2E container names
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: implementation-plan-v1 Task 6.6
- **Description**: Identity E2E container names may not match PRODUCT compose service names.
- **Acceptance Criteria**:
  - [ ] Container names verified and corrected
  - [ ] Magic constants updated if needed
- **Files**: E2E test files, magic constants

---

### Phase 7: Coverage & Mutation Testing

**Phase Objective**: Improve coverage and establish mutation testing baseline

#### Task 7.1: Improve crypto/jose coverage (89.9% → ~91%)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: None
- **Source**: fixes-v4 QG-3
- **Description**: crypto/jose at 89.9% with structural ceiling ~91%. ~111 uncovered stmts mainly from jwx v3 library error paths.
- **Acceptance Criteria**:
  - [ ] Coverage ≥91% (structural ceiling)
  - [ ] New tests use table-driven pattern
  - [ ] No interface wrapping the jwx library
- **Files**: `internal/shared/crypto/jose/*_test.go`

#### Task 7.2: Improve production package coverage to ≥95%
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: None
- **Source**: fixes-v4 QG-4
- **Description**: 17 production packages below 95% coverage target. Focus on packages close to threshold first.
- **Acceptance Criteria**:
  - [ ] Coverage measured for all production packages
  - [ ] Packages close to 95% brought above threshold
  - [ ] Table-driven tests used
- **Files**: Multiple production packages

#### Task 7.3: Run gremlins mutation testing baseline
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Tasks 7.1-7.2
- **Source**: fixes-v4 QG-6
- **Description**: Run gremlins on packages meeting ≥95% coverage to establish mutation testing baseline.
- **Acceptance Criteria**:
  - [ ] Gremlins run on all qualifying packages
  - [ ] Results documented
  - [ ] Packages below 95% efficacy identified for improvement
- **Files**: `.gremlins.yaml`, mutation results

---

## Cross-Cutting Tasks

### Testing
- [ ] All tests pass with `-shuffle=on -count=1`
- [ ] Race detector clean: `go test -race -count=2 ./...`
- [ ] No flaky tests

### Code Quality
- [ ] Linting passes: `golangci-lint run` and `golangci-lint run --build-tags e2e,integration`
- [ ] Zero `//nolint:wsl` violations
- [ ] No files >500 lines
- [ ] All magic constants in `internal/shared/magic/`

### Deployment
- [ ] 65/65 deployment validators pass
- [ ] E2E infrastructure functional

---

## Provenance — Source Mapping

This table maps each task back to its original plan/finding:

| Task | Original Source | Original Finding ID |
|------|----------------|---------------------|
| 1.1 | fixes-v5 | F-1.1 |
| 1.2 | fixes-v5 | F-1.2, F-1.3 |
| 1.3 | fixes-v5 | F-1.5 |
| 1.4 | fixes-v5 | F-1.4 |
| 1.5 | fixes-v6 | F-6.4 |
| 1.6 | verified 2026-02-23 | test run |
| 1.7 | fixes-v6 | F-6.5 |
| 2.1 | fixes-v6 | F-6.9 |
| 2.2 | fixes-v6 | F-6.8 |
| 2.3 | fixes-v6 | F-6.15 |
| 2.4 | fixes-v5 | F-3.2 |
| 2.5 | fixes-v5 | F-3.4 |
| 2.6 | fixes-v6 | F-6.17 |
| 2.7 | fixes-v6 | F-6.44 |
| 2.8 | fixes-v6 | F-6.39 |
| 3.1 | fixes-v6 | F-6.10 |
| 3.2 | fixes-v6 | F-6.11 |
| 3.3 | fixes-v6 | F-6.12 |
| 3.4 | fixes-v6 | F-6.25 |
| 3.5 | fixes-v5/v6 | F-4.2, F-6.36 |
| 4.1 | fixes-v6 | F-6.16 |
| 4.2 | fixes-v5 | F-2.11 |
| 4.3 | fixes-v5 | F-2.12-2.14 |
| 4.4 | fixes-v6 | F-6.13 |
| 4.5 | fixes-v6 | F-6.14, F-6.32, F-6.33 |
| 4.6 | fixes-v5 | F-1.6 |
| 5.1 | fixes-v6 | F-6.7 |
| 5.2 | fixes-v5 | F-4.1 |
| 5.3 | fixes-v6 | F-6.34 |
| 5.4 | fixes-v6 | F-6.37 |
| 5.5 | fixes-v6 | F-6.38 |
| 6.1 | fixes-v1 | blocker 02 |
| 6.2 | fixes-v1 | blocker 02 |
| 6.3 | fixes-v1 | blocker 02 |
| 6.4 | impl-plan-v1 | Task 6.4 |
| 6.5 | impl-plan-v1 | Task 6.6 |
| 7.1 | fixes-v4 | QG-3 |
| 7.2 | fixes-v4 | QG-4 |
| 7.3 | fixes-v4 | QG-6 |

## Items Verified Complete (NOT carried forward)

The following items from prior plans were verified as already completed:
- fixes-v4: Poll utility extraction, duration/timeout constants consolidation, gremlins config consolidation
- fixes-v2: All 7 phases complete (deployment restructuring, validators, config validation)
- fixes-v3: All 6 phases complete (configs/deployments CICD rigor, 8 validators, 65/65 passing)
- impl-plan-v1: Phases 1-3 complete (SERVICE compose), Phase 4 complete (PRODUCT compose), Phase 5 complete (magic constants)
- fixes-v1: Docker Compose fixes (telemetry, identity commands, ports), documentation verification
- fixes-v4: QG-1 linting, QG-2 flaky test fix, infrastructure coverage ≥98% (except crypto/jose)
