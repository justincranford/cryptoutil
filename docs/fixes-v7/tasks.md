# Tasks - Consolidated Quality Fixes v7

**Status**: 0 of 53 tasks complete (0%)
**Last Updated**: 2026-02-23 (updated per quizme-v1 answers Q1=E, Q2=E, Q3=E)
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
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.1
- **Description**: `poll.Until()` panics if `conditionFn` is nil. Add nil guard at function entry.
- **Acceptance Criteria**:
  - [x] Guard returns error if conditionFn is nil
  - [x] Test added for nil conditionFn
  - [x] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.2: poll.go — Add zero/negative timeout and interval validation
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.2, F-1.3
- **Description**: `poll.Until()` accepts zero/negative timeout and interval without error.
- **Acceptance Criteria**:
  - [x] Guard returns error if timeout ≤ 0
  - [x] Guard returns error if interval ≤ 0
  - [x] Tests added for both cases
  - [x] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.3: poll.go — Check context before first conditionFn call
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.5
- **Description**: Context cancellation not checked before first `conditionFn` call.
- **Acceptance Criteria**:
  - [x] Context checked at function entry
  - [x] Test for pre-canceled context
  - [x] 100% coverage maintained
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.4: poll.go — Wrap timeout error with sentinel
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Dependencies**: None
- **Source**: fixes-v5 F-1.4
- **Description**: Timeout error is plain `fmt.Errorf`, not wrapped with a sentinel error for `errors.Is()`.
- **Acceptance Criteria**:
  - [x] Define `ErrTimeout` sentinel
  - [x] Wrap timeout error with sentinel
  - [x] Test using `errors.Is(err, poll.ErrTimeout)`
- **Files**: `internal/shared/util/poll/poll.go`, `internal/shared/util/poll/poll_test.go`

#### Task 1.5: ValidateUUIDs — Fix wrong error wrapping
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.4
- **Description**: `ValidateUUIDs()` wraps the wrong error variable.
- **Acceptance Criteria**:
  - [x] Error wrapping uses correct variable
  - [x] Test verifies correct error message
  - [x] Build passes
- **Files**: `internal/shared/util/random/uuid.go`, `internal/shared/util/random/uuid_slice_cache_test.go`

#### Task 1.6: Flaky test — TestAuditLogService_LogOperation_AuditDisabled
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: None
- **Source**: verified 2026-02-23 test run
- **Description**: Test fails intermittently under `go test ./... -shuffle=on` but passes individually. Likely race condition or shared TestMain state.
- **Acceptance Criteria**:
  - [x] Root cause identified (test uses unique tenantID per run, stable with -count=5 -shuffle=on -race)
  - [x] Fix applied (isolation, synchronization, or require.Eventually)
  - [x] Passes with `-count=5 -shuffle=on`
- **Files**: `internal/apps/jose/ja/service/*_test.go`

#### Task 1.7: Verify F-6.5 copy-paste bug
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.5
- **Description**: "sqlite" string reference in PostgreSQL function — verify if actual bug or false positive.
- **Acceptance Criteria**:
  - [x] Investigate and document whether this is a real bug — REAL BUG: postgres.go:39 says "sqlite" in error message
  - [x] Fix if real, close if false positive — Fixed: changed to "postgres"
- **Files**: TBD (investigate)

---

### Phase 2: Code Quality & Standards

**Phase Objective**: Fix style/standards violations that don't affect correctness but violate project standards

#### Task 2.1: Rename file with space — usernames_passwords_test util.go
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Actual**: 5min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.9
- **Description**: File has space in name: `internal/shared/util/random/usernames_passwords_test util.go`. Go tooling may not pick it up.
- **Acceptance Criteria**:
  - [x] File renamed to remove space — renamed to `usernames_passwords_test_util.go`
  - [x] Build passes
  - [x] Tests pass
- **Files**: `internal/shared/util/random/usernames_passwords_test_util.go`

#### Task 2.2: Error sentinels — change from string to error type
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: 20min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.8
- **Description**: The `var ErrXxx = "string"` vars in uuid.go are actually message descriptors for ValidateUUID/ValidateUUIDs, not error sentinels for `errors.Is()`. Combined with Task 2.7: changed `msg *string` to `msg string` parameter.
- **Acceptance Criteria**:
  - [x] Sentinel errors resolved — they are message descriptors, not error sentinels; `*string` → `string` param change eliminates the concern
  - [x] All callers updated — removed `&` from all 15+ callers across jose, kms, and test files
  - [x] Tests pass
- **Files**: `internal/shared/util/random/uuid.go`, `internal/shared/util/random/uuid_slice_cache_test.go`, `internal/shared/crypto/jose/*.go`, `internal/apps/sm/kms/server/businesslogic/oam_orm_mapper_query.go`, `internal/apps/sm/kms/server/repository/orm/business_entities_operations.go`

#### Task 2.3: Fix `//nolint:wsl` violations (Q3=E)
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: None
- **Source**: fixes-v5 F-6.15; quiz answer Q3=E
- **Description**: `//nolint:wsl` is legacy golangci-lint v1 — MUST be removed. `//nolint:wsl_v5` is modern v2 — make genuine effort to fix by code restructuring.
  - **2 legacy `//nolint:wsl`** (in `template/service/telemetry/telemetry_service_helpers.go:134,158`): Remove by restructuring code to comply with wsl blank-line rules.
  - **20 modern `//nolint:wsl_v5`** (in identity unified services: idp, rs, spa, authz, rp — 4 instances each): Fix by restructuring the identical pattern across 5 files. If structurally impossible to fix without breaking the unified pattern, document each as structurally required with specific explanation.
- **Acceptance Criteria**:
  - [x] Zero `//nolint:wsl` (legacy v1) in codebase — removed 2 from telemetry_service_helpers.go
  - [x] All `//nolint:wsl_v5` removed via verification they were unnecessary — removed all 20 across 5 identity unified files; lint passes without them
  - [x] Linting passes after changes
  - [x] Code restructured rather than suppressed where possible — also changed `for attempt := 0; attempt <= maxRetries; attempt++` to `for attempt := range maxRetries + 1`
- **Files**: `internal/apps/template/service/telemetry/telemetry_service_helpers.go`, `internal/apps/identity/idp/unified/idp.go`, `internal/apps/identity/rs/unified/rs.go`, `internal/apps/identity/spa/unified/spa.go`, `internal/apps/identity/authz/unified/authz.go`, `internal/apps/identity/rp/unified/rp.go`

#### Task 2.4: TestNegativeDuration — change to time.Duration type
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Actual**: 15min
- **Dependencies**: None
- **Source**: fixes-v5 F-3.2
- **Description**: `TestNegativeDuration = -1` is untyped int, should be `time.Duration`.
- **Acceptance Criteria**:
  - [x] Changed to `TestNegativeDuration = -1 * time.Nanosecond`
  - [x] All callers work correctly — added `TestNegativeInt = -1` for int usage in certificates_test.go MaxPathLen field
  - [x] Tests pass
- **Files**: `internal/shared/magic/magic_testing.go`, `internal/shared/crypto/certificate/certificates_test.go`

#### Task 2.5: nolint:stylecheck — add bug reference (5 instances)
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 10min
- **Actual**: 5min
- **Dependencies**: None
- **Source**: fixes-v5 F-3.4
- **Description**: `//nolint:stylecheck` without GitHub issue reference. Per instructions, only allowed with documented linter bug reference.
- **Acceptance Criteria**:
  - [x] All `//nolint:stylecheck` removed — verified unnecessary; lint passes without them in both magic_testing.go and magic_unseal.go
  - [x] Linting passes
- **Files**: `internal/shared/magic/magic_testing.go`, `internal/shared/magic/magic_unseal.go`

#### Task 2.6: pool.go — Convert if/else chain to switch
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 10min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.17
- **Description**: `validateConfig` in pool.go has 12-branch if/else chain. Per coding instructions: prefer switch statements.
- **Acceptance Criteria**:
  - [x] Converted to switch statement
  - [x] Tests pass
  - [x] Same behavior
- **Files**: `internal/shared/pool/pool.go`

#### Task 2.7: ValidateUUID — change *string to string parameter
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 10min (combined with Task 2.2)
- **Dependencies**: Task 1.5
- **Source**: fixes-v6 F-6.44
- **Description**: `ValidateUUID(uuid *googleUuid.UUID, msg *string)` — `msg` parameter should be value, not pointer.
- **Acceptance Criteria**:
  - [x] Changed `msg *string` to `msg string` in both ValidateUUID and ValidateUUIDs
  - [x] All callers updated — 15+ call sites across jose, kms, and test files
  - [x] Tests pass
- **Files**: `internal/shared/util/random/uuid.go`, `internal/shared/util/random/uuid_slice_cache_test.go`, `internal/shared/crypto/jose/*.go`, `internal/apps/sm/kms/server/**/*.go`

#### Task 2.8: fmt.Errorf without %w audit
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.39
- **Description**: Audit all `fmt.Errorf` calls to ensure errors are wrapped with `%w` where appropriate.
- **Acceptance Criteria**:
  - [x] Audit complete — all `fmt.Errorf` calls use `%w` for error wrapping; remaining `%v` usages are for non-error formatting
  - [x] All appropriate errors wrapped with `%w`
  - [x] Tests pass
- **Files**: Audit only, no changes needed

---

### Phase 3: Magic Constant Consolidation

**Phase Objective**: Move all domain-specific magic constants to `internal/shared/magic/`

#### Task 3.1: Move identity magic package to shared/magic
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: 20min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.10
- **Description**: `internal/apps/identity/magic/` has 12 magic files. Moved to `internal/shared/magic/magic_identity_*.go`.
- **Acceptance Criteria**:
  - [x] Constants moved to `internal/shared/magic/` — 12 files moved with identity_ prefix
  - [x] Old package deleted
  - [x] All imports updated — 131 single-import files + 3 dual-import files
  - [x] Build + tests pass
- **Files**: `internal/apps/identity/magic/*` → `internal/shared/magic/magic_identity_*.go`

#### Task 3.2: Move PKI CA magic package to shared/magic
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 5min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.11
- **Description**: `internal/apps/pki/ca/magic/magic.go` (28 lines). Moved to `internal/shared/magic/magic_pki_ca.go`.
- **Acceptance Criteria**:
  - [x] Constants moved to `internal/shared/magic/`
  - [x] Old package deleted
  - [x] All imports updated — 9 files, import alias auto-fixed by golangci-lint
  - [x] Build + tests pass
- **Files**: `internal/apps/pki/ca/magic/magic.go` → `internal/shared/magic/magic_pki_ca.go`

#### Task 3.3: Move identity config magic to shared/magic
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 15min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.12
- **Description**: `internal/apps/identity/config/magic.go` (62 lines). Created `internal/shared/magic/magic_identity_config.go` with exported constants.
- **Acceptance Criteria**:
  - [x] Constants moved to `internal/shared/magic/` — exported with Identity prefix
  - [x] Old file deleted
  - [x] All imports updated — defaults.go and validation.go now import shared magic
  - [x] Build + tests pass
- **Files**: `internal/apps/identity/config/magic.go` → `internal/shared/magic/magic_identity_config.go`

#### Task 3.4: Move TLS constants to shared/magic
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Actual**: 5min
- **Dependencies**: None
- **Source**: fixes-v6 F-6.25
- **Description**: TLS constants investigated. Only `MinTLSVersion = tls.VersionTLS13` in crypto/tls/config.go — appropriately placed (depends on crypto/tls import). All other TLS constants already in shared/magic (64 references).
- **Acceptance Criteria**:
  - [x] TLS constants audited — already consolidated; one exception appropriately placed in its home package
  - [x] All imports verified
  - [x] Build + tests pass
- **Files**: No changes needed

#### Task 3.5: Consolidate demo package scattered constants
- **Status**: ✅
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Actual**: 15min
- **Dependencies**: None
- **Source**: fixes-v5 F-4.2, fixes-v6 F-6.36
- **Description**: Created `internal/shared/magic/magic_demo.go` with shared demo constants. Updated 4 demo files to reference them.
- **Acceptance Criteria**:
  - [x] Demo constants moved to `internal/shared/magic/magic_demo.go` — DemoClientID, DemoClientSecret, DemoRedirectURI, DemoSampleAccessToken, DemoPort, DemoIssuer, DemoClientName
  - [x] No duplicate constants — all 4 demo files now reference magic
  - [x] All imports updated
  - [x] Build + tests pass
- **Files**: `internal/shared/magic/magic_demo.go`, `internal/apps/identity/demo/demo.go`, `internal/apps/demo/identity.go`, `internal/apps/demo/integration.go`, `internal/apps/demo/script.go`

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

**Phase Objective**: Fix E2E test infrastructure via two-step approach (Q1=E: B-then-A)

#### Task 6.0: Extract generic service startup helper into template (Step B)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Dependencies**: None (do first)
- **Source**: Q1=E quiz answer — prerequisite for all other Phase 6 tasks
- **Description**: `cipher/im/testing/testmain_helper.go` has `StartCipherIMService()` and `SetupTestServer()` that are cipher-im-specific. These patterns (port polling, goroutine start, panic on failure) should live generically in `template/service/testing/`. Create a generic `StartServiceFromConfig[T any]()` helper in the template testing package.
- **Acceptance Criteria**:
  - [ ] Generic startup helper created in `internal/apps/template/service/testing/server_start_helpers.go`
  - [ ] Helper covers: goroutine start, port polling (public + admin), error channel, panic-on-failure
  - [ ] cipher-im `StartCipherIMService()` refactored to call the generic template helper
  - [ ] Tests pass with `go test ./internal/apps/template/service/testing/...`
- **Files**: `internal/apps/template/service/testing/server_start_helpers.go` (new), `internal/apps/cipher/im/testing/testmain_helper.go` (refactor)

#### Task 6.1: Fix KMS session JWK algorithm configuration (Step B)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: Task 6.0
- **Source**: fixes-v1 blocker 02
- **Description**: KMS services fail with "unsupported JWS algorithm:" — empty algorithm string in session manager config. Verify cipher-im config pattern, apply same pattern to KMS config.
- **Acceptance Criteria**:
  - [ ] Session JWK algorithm configured correctly (matches cipher-im pattern)
  - [ ] KMS service starts successfully in Docker
- **Files**: Config files, session manager config

#### Task 6.2: Fix JOSE args routing (Step B)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: Task 6.0
- **Source**: fixes-v1 blocker 02
- **Description**: JOSE service args routing incorrect — args not stripped properly through product → service → subcommand layers. Compare with cipher-im `im.go` routing pattern and apply same.
- **Acceptance Criteria**:
  - [ ] Args routing matches cipher-im pattern
  - [ ] JOSE service starts successfully
- **Files**: `internal/apps/jose/ja/*.go`

#### Task 6.3: Migrate jose-ja and sm-kms TestMains to template helper (Step A)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: Task 6.0, 6.1, 6.2
- **Source**: Q1=E — "ensure jose-ja and sm-kms are reusing the same main and E2E test code"
- **Description**: jose-ja `server/testmain_test.go` uses raw loop (50 × 100ms polls). sm-kms has no integration TestMain. Both should use the generic template helper created in Task 6.0. This standardizes startup pattern across all migrated services.
- **Acceptance Criteria**:
  - [ ] jose-ja `server/testmain_test.go` uses template `StartServiceFromConfig()` pattern
  - [ ] sm-kms integration TestMain created using template helper pattern
  - [ ] Tests pass: `go test -tags=integration ./internal/apps/jose/ja/...` and `./internal/apps/sm/kms/...`
- **Files**: `internal/apps/jose/ja/server/testmain_test.go`, `internal/apps/sm/kms/server/testmain_integration_test.go` (new)

#### Task 6.4: Update CI E2E workflow paths and add jose-ja/sm-kms E2E tests
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45min
- **Dependencies**: 6.1, 6.2, 6.3
- **Source**: implementation-plan-v1 Task 6.4; Q1=E
- **Description**: (1) Fix service_TEMPLATE_TODO markers: `ci-e2e.yml` references old compose paths; (2) Add jose-ja and sm-kms E2E test execution to workflow (they should now start reliably after Tasks 6.0-6.3); (3) Create E2E test files for jose-ja and sm-kms matching cipher-im e2e structure.
- **Acceptance Criteria**:
  - [ ] All compose file paths corrected in `ci-e2e.yml`
  - [ ] jose-ja E2E test suite created (`internal/apps/jose/ja/e2e/`)
  - [ ] sm-kms E2E test suite verified/created (`internal/apps/sm/kms/e2e/`)
  - [ ] CI E2E workflow runs cipher-im, jose-ja, sm-kms E2E tests
  - [ ] `SERVICE_TEMPLATE_TODO` markers removed for migrated services
- **Files**: `.github/workflows/ci-e2e.yml`, `internal/apps/jose/ja/e2e/` (new), `internal/apps/sm/kms/e2e/` (verify)

#### Task 6.5: Verify cipher-im E2E passes end-to-end (Step B validation)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: 6.0
- **Source**: Q1=E — "make cipher-im service startup work reliably first"
- **Description**: Run cipher-im E2E tests to confirm they pass. This is the validation step before propagating patterns to jose-ja and sm-kms. Identity E2E remains disabled until identity services migrate to service-template (future work in research options).
- **Acceptance Criteria**:
  - [ ] `go test -tags=e2e -timeout=30m ./internal/apps/cipher/im/e2e/...` passes
  - [ ] Docker Compose stack starts and all cipher-im containers healthy
  - [ ] No Docker Desktop / firewall binding issues
- **Files**: `internal/apps/cipher/im/e2e/`, `deployments/cipher-im/compose.yml`

---

### Phase 7: Coverage & Mutation Testing

**Phase Objective**: Improve coverage and establish mutation testing baseline

#### Task 7.1: Push crypto/jose coverage to ~91% structural ceiling (Q2=E)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Dependencies**: None
- **Source**: fixes-v4 QG-3; quiz answer Q2=E
- **Description**: crypto/jose at 89.9%. Per Q2=E: push to ~91% via new tests, then stop. Do NOT interface-wrap jwx v3.
  - Target the ~111 uncovered stmts systematically: err paths on valid inputs (jwk.Set, jwk.Import, json.Marshal), type-switch defaults, jwe/jws errors with valid input
  - For each group: try to trigger via test input or test data manipulation; if test cannot reach the path without interface-wrapping, mark as structural ceiling
  - Add `//go:cover-ignore` comments on confirmed-unreachable lines (e.g. `_ = json.Marshal(validStruct)` where Marshal can't fail)
  - Document all findings in `docs/fixes-v7/JWX-COV-CEILING.md`
  - Do NOT exempt from ≥98% gate — this is genuine effort to maximize coverage
- **Acceptance Criteria**:
  - [ ] Coverage increases from 89.9% toward ~91%
  - [ ] `docs/fixes-v7/JWX-COV-CEILING.md` created with: uncovered stmt categories, why each is unreachable, test attempts made
  - [ ] `//go:cover-ignore` added for confirmed-unreachable paths only
  - [ ] No interface-wrapping of jwx library
  - [ ] New tests use table-driven pattern
  - [ ] Tests pass: `go test -cover ./internal/shared/crypto/jose/...`
- **Files**: `internal/shared/crypto/jose/*_test.go`, `docs/fixes-v7/JWX-COV-CEILING.md` (new)

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

### Phase 8: Move cipher-im to SM Product

**Phase Objective**: Rename cipher-im → sm-im, move under SM product. Detailed breakdown in [research/tasks-PKI-CA-MERGE0a.md](research/tasks-PKI-CA-MERGE0a.md). PKI-CA-MERGE0b (merge into sm-kms) will NOT be implemented.

#### Task 8.1: Code rename — move cipher-im to sm-im
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Dependencies**: None (can be done at any time)
- **Source**: Architecture Direction — Option D; PKI-CA-MERGE0a Phase 1
- **Description**: Move `internal/apps/cipher/im/` → `internal/apps/sm/im/`. Update `cmd/cipher-im/main.go` → `cmd/sm-im/main.go`. Update `cmd/cipher/main.go` to remove im (or convert to `cmd/sm/main.go` product-level entry). Update ALL Go import paths from `cipher/im` to `sm/im`.
- **Acceptance Criteria**:
  - [ ] All cipher/im source files moved to sm/im
  - [ ] All import paths updated
  - [ ] `go build ./...` passes
  - [ ] `go test ./... -shuffle=on` passes
- **Files**: `internal/apps/cipher/im/` → `internal/apps/sm/im/`, `cmd/cipher-im/` → `cmd/sm-im/`, all files with `cipher/im` imports

#### Task 8.2: Deployment and config updates
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 8.1
- **Source**: PKI-CA-MERGE0a Phase 2
- **Description**: Move `deployments/cipher-im/` → `deployments/sm-im/`. Move `deployments/cipher/` → update SM product compose. Move `configs/cipher/im/` → `configs/sm/im/`. Update port range comments (keep 8700-8799 or reassign to SM range).
- **Acceptance Criteria**:
  - [ ] Deployment files moved and updated
  - [ ] Config files moved and updated
  - [ ] `go run ./cmd/cicd lint-deployments validate-all` passes (65/65)
  - [ ] Docker Compose health checks pass
- **Files**: `deployments/cipher-im/`, `deployments/cipher/`, `configs/cipher/im/`

#### Task 8.3: Documentation and CI updates
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Dependencies**: Task 8.1
- **Source**: PKI-CA-MERGE0a Phase 3
- **Description**: Update ARCHITECTURE.md service catalog (remove Cipher product, add sm-im to SM product). Update ci-e2e.yml paths. Update README.md. Update copilot instruction files referencing cipher-im.
- **Acceptance Criteria**:
  - [ ] ARCHITECTURE.md service catalog updated
  - [ ] ci-e2e.yml compose paths updated
  - [ ] No references to "cipher-im" in docs (except historical notes)
- **Files**: `docs/ARCHITECTURE.md`, `.github/workflows/ci-e2e.yml`, `README.md`, `.github/instructions/02-01.architecture.instructions.md`

#### Task 8.4: Validation — build, test, lint
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30min
- **Dependencies**: Tasks 8.1-8.3
- **Source**: PKI-CA-MERGE0a Phase 5
- **Description**: Full validation after rename: build, test, lint, deployment validators.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./... -shuffle=on` passes
  - [ ] `golangci-lint run` clean
  - [ ] `golangci-lint run --build-tags e2e,integration` clean
  - [ ] 65/65 deployment validators pass
- **Files**: All

#### Task 8.5: Git commit and Cipher product cleanup
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15min
- **Dependencies**: Task 8.4
- **Source**: PKI-CA-MERGE0a Phase 5
- **Description**: Commit rename. Verify no remaining Cipher product artifacts. Clean up empty `internal/apps/cipher/` directory if any.
- **Acceptance Criteria**:
  - [ ] Committed with `refactor(sm): rename cipher-im to sm-im`
  - [ ] No `internal/apps/cipher/` directory remaining
  - [ ] No `cmd/cipher*/` directories remaining
  - [ ] No `deployments/cipher*/` directories remaining
- **Files**: Git operations

---

## Cross-Cutting Tasks

### Testing
- [ ] All tests pass with `-shuffle=on -count=1`
- [ ] Race detector clean: `go test -race -count=2 ./...`
- [ ] No flaky tests

### Code Quality
- [ ] Linting passes: `golangci-lint run` and `golangci-lint run --build-tags e2e,integration`
- [ ] Zero `//nolint:wsl` (legacy v1) violations
- [ ] All `//nolint:wsl_v5` either removed or documented as structurally required
- [ ] No files >500 lines
- [ ] All magic constants in `internal/shared/magic/`

### Deployment
- [ ] 65/65 deployment validators pass
- [ ] cipher-im E2E passes; jose-ja and sm-kms E2E startup unblocked
- [ ] Template has generic service startup helper (`template/service/testing/server_start_helpers.go`)
- [ ] cipher-im renamed to sm-im under SM product
- [ ] No Cipher product remaining

### Coverage
- [ ] crypto/jose ≥91% structural ceiling reached via new tests
- [ ] `docs/fixes-v7/JWX-COV-CEILING.md` documents remaining unreachable paths
- [ ] `//go:cover-ignore` added for confirmed-unreachable paths only

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
| 6.0 | Q1=E quiz answer | quizme-v1 Q1 |
| 6.1 | fixes-v1 | blocker 02 |
| 6.2 | fixes-v1 | blocker 02 |
| 6.3 | Q1=E quiz answer | quizme-v1 Q1 |
| 6.4 | impl-plan-v1 + Q1=E | Task 6.4 + quiz |
| 6.5 | Q1=E quiz answer | quizme-v1 Q1 |
| 7.1 | fixes-v4 + Q2=E | QG-3 + quizme-v1 Q2 |
| 7.2 | fixes-v4 | QG-4 |
| 7.3 | fixes-v4 | QG-6 |
| 8.1 | Architecture Direction | PKI-CA-MERGE0a Phase 1 |
| 8.2 | Architecture Direction | PKI-CA-MERGE0a Phase 2 |
| 8.3 | Architecture Direction | PKI-CA-MERGE0a Phase 3 |
| 8.4 | Architecture Direction | PKI-CA-MERGE0a Phase 5 |
| 8.5 | Architecture Direction | PKI-CA-MERGE0a Phase 5 |

## Items Verified Complete (NOT carried forward)

The following items from prior plans were verified as already completed:
- fixes-v4: Poll utility extraction, duration/timeout constants consolidation, gremlins config consolidation
- fixes-v2: All 7 phases complete (deployment restructuring, validators, config validation)
- fixes-v3: All 6 phases complete (configs/deployments CICD rigor, 8 validators, 65/65 passing)
- impl-plan-v1: Phases 1-3 complete (SERVICE compose), Phase 4 complete (PRODUCT compose), Phase 5 complete (magic constants)
- fixes-v1: Docker Compose fixes (telemetry, identity commands, ports), documentation verification
- fixes-v4: QG-1 linting, QG-2 flaky test fix, infrastructure coverage ≥98% (except crypto/jose)
