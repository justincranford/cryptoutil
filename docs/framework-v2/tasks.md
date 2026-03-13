# Tasks - Framework v2: Service Code Quality Refactoring

**Status**: 0 of 34 tasks complete (0%)
**Last Updated**: 2026-03-12
**Created**: 2026-03-12

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions.**

---

## Task Checklist

### Phase 1: testdb.NewClosedSQLiteDB Helper

**Phase Objective**: Add shared closed-DB helper to service-template testdb package and add a fitness rule to enforce its use.

#### Task 1.1: Implement testdb.NewClosedSQLiteDB

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [fill when complete]
- **Dependencies**: None
- **Description**: Add `NewClosedSQLiteDB(t *testing.T, applyMigrations func(*sql.DB) error) *gorm.DB` to `testdb/testdb.go`. Follows same open/PRAGMA/GORM pattern as `NewInMemorySQLiteDB` but closes the underlying connection before returning.
- **Acceptance Criteria**:
  - [ ] Function added to `internal/apps/template/service/testing/testdb/testdb.go`
  - [ ] Accepts `applyMigrations func(*sql.DB) error` (domain-specific injection)
  - [ ] Registers noop `t.Cleanup()` for consistency
  - [ ] Unit tests: injection of mock openFn for error paths; ≥98% coverage
  - [ ] `go test ./internal/apps/template/service/testing/testdb/...` passes
- **Files**:
  - `internal/apps/template/service/testing/testdb/testdb.go` (modify)
  - `internal/apps/template/service/testing/testdb/testdb_test.go` (modify)

#### Task 1.2: Add no_local_create_closed_database fitness rule

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Add a lint-fitness sub-linter that detects private `createClosedDatabase`, `createClosedDB`, `createClosedServiceDependencies` functions defined outside the `testdb` package. Confirm it FAILS on current jose-ja BEFORE the cleanup, proving the rule works.
- **Acceptance Criteria**:
  - [ ] New sub-linter added to `cmd/cicd/lint_fitness/`
  - [ ] Rule fires on jose-ja `repository/database_error_test.go` (confirmed pre-cleanup)
  - [ ] Rule does NOT fire on `testdb/testdb.go` (allowlist for testdb package)
  - [ ] Rule passes after Phase 2/3/4 cleanups
  - [ ] `go run ./cmd/cicd lint-fitness` includes new rule
  - [ ] Sub-linter tests ≥98% coverage
- **Files**:
  - New file in `cmd/cicd/lint_fitness/` (e.g., `lint_closed_db_pattern.go`)
  - Corresponding test file

#### Task 1.3: Phase 1 quality gate

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Tasks 1.1, 1.2
- **Description**: Full quality gate + post-mortem.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go test ./internal/apps/template/service/testing/testdb/... -cover` ≥98%
  - [ ] `golangci-lint run ./internal/apps/template/service/testing/testdb/...` clean
  - [ ] `go run ./cmd/cicd lint-fitness` passes (new rule present, fires on jose-ja as expected)
  - [ ] lessons.md updated with Phase 1 post-mortem

---

### Phase 2: jose-ja Cleanup

**Phase Objective**: Fix all four problems in jose-ja (handler DTOs, closed-DB helpers, file proliferation in repository/ and service/, domain naming).

#### Task 2.1: Replace hand-rolled handler DTOs with generated models

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [fill when complete]
- **Dependencies**: None (can start independently of Phase 1)
- **Description**: `server/apis/jwk_handler.go` defines `CreateElasticJWKRequest`, `ElasticJWKResponse`, `MaterialJWKResponse` instead of using `api/jose/models/models.gen.go`. Replace with generated types and add explicit mapping functions.
- **Acceptance Criteria**:
  - [ ] `server/apis/jwk_handler.go` imports `api/jose/models` — zero hand-rolled request/response structs
  - [ ] Explicit mapping functions `toElasticJWKResponse`, `toMaterialJWKResponse` added as unexported helpers
  - [ ] Handler tests updated to use generated types
  - [ ] `go test ./internal/apps/jose/ja/server/...` passes
  - [ ] No API behavior change (all existing tests still pass)
- **Files**:
  - `internal/apps/jose/ja/server/apis/jwk_handler.go` (modify)
  - `internal/apps/jose/ja/server/apis/jwk_handler_test.go` (modify)
  - `internal/apps/jose/ja/server/apis/jwk_handler_material.go` (modify)

#### Task 2.2: Migrate closed-DB helpers to testdb.NewClosedSQLiteDB

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Replace `createClosedDatabase()` in `repository/database_error_test.go` and `createClosedServiceDependencies()` in `service/database_error_test.go` with calls to `testdb.NewClosedSQLiteDB(t, applyMigrations)`.
- **Acceptance Criteria**:
  - [ ] No `createClosedDatabase` function in `jose/ja/repository/` package
  - [ ] No `createClosedServiceDependencies` function in `jose/ja/service/` package
  - [ ] All error-path tests use `testdb.NewClosedSQLiteDB()`
  - [ ] `go test ./internal/apps/jose/ja/repository/... ./internal/apps/jose/ja/service/...` passes
  - [ ] Fitness rule passes (no more violations in jose-ja)

#### Task 2.3: Merge repository/ error-path test files

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Consolidate split error-path files into domain-named test files as table subtests. Files to merge: `database_error_test.go`, `database_error_material_test.go`, `database_error_audit_test.go`, `additional_edge_cases_test.go`, `audit_log_list_test.go`.
- **Acceptance Criteria**:
  - [ ] `database_error_test.go` deleted; cases in `elastic_jwk_repository_test.go`
  - [ ] `database_error_material_test.go` deleted; cases in `material_jwk_repository_test.go`
  - [ ] `database_error_audit_test.go` deleted; cases in `audit_repository_test.go`
  - [ ] `additional_edge_cases_test.go` and `audit_log_list_test.go` deleted; cases distributed to appropriate domain test files
  - [ ] Test count before == test count after (no test loss)
  - [ ] `go test ./internal/apps/jose/ja/repository/...` passes

#### Task 2.4: Merge service/ error-path test files

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Actual**: [fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Merge 10+ split service error-path files into the relevant service test files.
- Target merges:
  - `database_error_test.go` → `elastic_jwk_service_test.go`
  - `database_error_corrupt_test.go` + `database_error_corrupt2_test.go` → `elastic_jwk_service_test.go`
  - `database_error_extra_test.go` → relevant service test file
  - `database_error_jwe_test.go` → `jwe_service_test.go`
  - `error_coverage_jwe_jws_test.go` → `jwe_service_test.go` + `jws_service_test.go`
  - `error_coverage_jwks_rotation_test.go` → `jwks_service_test.go` or `material_rotation_service_test.go`
  - `error_coverage_jwt_test.go` → `jwt_service_test.go`
  - `mapping_functions_test.go` + `mapping_functions_parse_test.go` → appropriate service test file
- **Acceptance Criteria**:
  - [ ] All merged source files deleted
  - [ ] Test count before == test count after (no test loss)
  - [ ] Each test file name matches its source file name
  - [ ] `go test ./internal/apps/jose/ja/service/...` passes

#### Task 2.5: Rename domain/ → model/

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [fill when complete]
- **Dependencies**: Tasks 2.1, 2.3, 2.4 (all files that import domain package must be updated simultaneously)
- **Description**: `internal/apps/jose/ja/domain/` contains GORM structs, not true domain types. Rename to `model/` and update all import paths and aliases.
- **Acceptance Criteria**:
  - [ ] `internal/apps/jose/ja/domain/` replaced by `internal/apps/jose/ja/model/`
  - [ ] All import paths updated across jose-ja packages
  - [ ] Import alias updated: `cryptoutilAppsJoseJaModel` (was `cryptoutilAppsJoseJaDomain`)
  - [ ] `go build ./internal/apps/jose/ja/...` clean
  - [ ] `go test ./internal/apps/jose/ja/...` passes

#### Task 2.6: Phase 2 quality gate

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Tasks 2.1-2.5
- **Description**: Full quality gate + post-mortem.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go test ./internal/apps/jose/ja/... -shuffle=on` passes
  - [ ] `golangci-lint run ./internal/apps/jose/ja/...` clean
  - [ ] Coverage maintained: `go test -cover ./internal/apps/jose/ja/...` ≥95%
  - [ ] `go run ./cmd/cicd lint-fitness` passes (no jose-ja violations)
  - [ ] jose-ja repository/ has ≤5 test files total
  - [ ] jose-ja service/ has ≤1 test file per source file
  - [ ] lessons.md updated with Phase 2 post-mortem

---

### Phase 3: sm-im Cleanup

**Phase Objective**: Apply same cleanup to sm-im (fewer issues, no hand-rolled DTOs expected).

#### Task 3.1: Verify sm-im handler uses generated models

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: None
- **Description**: Confirm `server/apis/messages.go` uses `api/sm/im/` generated models exclusively. Document result.
- **Acceptance Criteria**:
  - [ ] Audit results documented in `test-output/framework-v2/sm-im-model-audit.md`
  - [ ] If violations found: create new task 3.1b to fix (block Phase 3)

#### Task 3.2: Migrate sm-im closed-DB helpers

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Replace `createClosedDBHandler()` and `createMixedHandler()` with `testdb.NewClosedSQLiteDB()` + inline service construction.
- **Acceptance Criteria**:
  - [ ] No `createClosedDBHandler` function in `sm/im/server/apis/`
  - [ ] No `createMixedHandler` function in `sm/im/server/apis/`
  - [ ] All error-path tests use `testdb.NewClosedSQLiteDB()` or inline setup
  - [ ] `go test ./internal/apps/sm/im/server/apis/...` passes
  - [ ] Fitness rule passes (no sm-im violations)

#### Task 3.3: Merge sm-im repository/ error-path files

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [fill when complete]
- **Dependencies**: Task 3.2
- **Description**: Merge `repository/error_paths_test.go`, `repository/error_returns_test.go`, `repository/concurrent_access_test.go` into domain-named test files.
- **Acceptance Criteria**:
  - [ ] `error_paths_test.go` deleted; cases in `message_repository_test.go`
  - [ ] `error_returns_test.go` deleted; cases distributed to domain test files
  - [ ] `concurrent_access_test.go` cases merged into appropriate domain test file
  - [ ] Test count before == test count after
  - [ ] `go test ./internal/apps/sm/im/repository/...` passes

#### Task 3.4: Merge sm-im server/apis/ error-path files

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: [fill when complete]
- **Dependencies**: Task 3.2
- **Description**: Evaluate `messages_dberror_test.go` and `messages_errorpaths_test.go` — merge error cases into `messages_test.go`.
- **Acceptance Criteria**:
  - [ ] `messages_dberror_test.go` deleted; cases in `messages_test.go`
  - [ ] `messages_errorpaths_test.go` deleted; cases in `messages_test.go`
  - [ ] Test count before == test count after
  - [ ] `go test ./internal/apps/sm/im/server/apis/...` passes

#### Task 3.5: sm-im domain/ audit

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: None
- **Description**: Confirm sm-im `domain/` contains only true domain types (no GORM tags, no fiber, no generated models). If violations: extend this task or create 3.5b.
- **Acceptance Criteria**:
  - [ ] Audit results documented
  - [ ] If domain is clean: mark done. If GORM present: rename to `model/` (same as jose-ja task 2.5).

#### Task 3.6: Phase 3 quality gate

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Tasks 3.1-3.5
- **Description**: Full quality gate + post-mortem.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go test ./internal/apps/sm/im/... -shuffle=on` passes
  - [ ] `golangci-lint run ./internal/apps/sm/im/...` clean
  - [ ] Coverage maintained ≥95%
  - [ ] `go run ./cmd/cicd lint-fitness` passes (no sm-im violations)
  - [ ] lessons.md updated with Phase 3 post-mortem

---

### Phase 4: sm-kms Assessment and Safe Cleanup

**Phase Objective**: Audit sm-kms, remove dead code, document v3-owned debt. No middleware changes.

#### Task 4.1: server/application/ audit

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [fill when complete]
- **Dependencies**: None
- **Description**: Determine if `server/application/` is dead code (replaced by service-template builder) or still active. Use call graph tracing from `kms.go` / `server.go` entrypoints.
- **Acceptance Criteria**:
  - [ ] Audit documented in `test-output/framework-v2/sm-kms-application-audit.md`
  - [ ] If dead: create Task 4.1b (remove with tests)
  - [ ] If active: document as v3 Phase 3 tech debt

#### Task 4.2: server/middleware/ documentation (NO CODE CHANGES)

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [fill when complete]
- **Dependencies**: None
- **Description**: Catalog sm-kms `server/middleware/` files. Map each to its future home in service-template (per v3 D1). Document in plan — do NOT change code.
- **Acceptance Criteria**:
  - [ ] Catalog written to `test-output/framework-v2/sm-kms-middleware-debt.md`
  - [ ] Each middleware file mapped to service-template counterpart (or "no counterpart yet")
  - [ ] v3 tasks.md updated with findings (Phase 3 task notes)
  - [ ] Zero code changes in this task

#### Task 4.3: repository/orm/ file proliferation cleanup

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Apply D3 rule to `repository/orm/` — merge split error-path files into domain-named test files. Migrate any closed-DB helpers to `testdb.NewClosedSQLiteDB()`.
- Files to evaluate (merge into domain-named test files):
  - `business_entities_additional_errors_test.go`
  - `business_entities_dead_code_test.go`
  - `business_entities_error_mapping_test.go`
  - `business_entities_get_errors_test.go`
  - `business_entities_gorm_errors_test.go`
  - `business_entities_materialkey_errors_test.go`
  - `business_entities_postgres_errors_test.go`
  - `business_entities_toapperr_test.go`
  - `business_entities_update_errors_test.go`
  - `business_entities_filters_uncovered_test.go`
- **Acceptance Criteria**:
  - [ ] Each merged file deleted
  - [ ] Test count before == test count after
  - [ ] `go test ./internal/apps/sm/kms/server/repository/orm/...` passes
  - [ ] Fitness rule passes (no sm-kms violations)

#### Task 4.4: Verify sm-kms handler uses generated models

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: None
- **Description**: Confirm `server/handler/` imports from `api/kms/server` generated types. Expected to already be correct (sm-kms was manually created with this in mind). Document.
- **Acceptance Criteria**:
  - [ ] Audit results documented
  - [ ] If violations found: create Task 4.4b to fix

#### Task 4.5: Phase 4 quality gate

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Tasks 4.1-4.4
- **Description**: Full quality gate + post-mortem.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go test ./internal/apps/sm/kms/... -shuffle=on` passes
  - [ ] `golangci-lint run ./internal/apps/sm/kms/...` clean
  - [ ] Coverage maintained ≥95%
  - [ ] `go run ./cmd/cicd lint-fitness` passes (no sm-kms violations in scope)
  - [ ] lessons.md updated with Phase 4 post-mortem

---

### Phase 5: Knowledge Propagation

**Phase Objective**: Propagate all lessons and patterns to permanent artifacts. NEVER skip.

#### Task 5.1: Update ARCHITECTURE.md

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [fill when complete]
- **Dependencies**: Phases 1-4 complete
- **Description**: Update ARCHITECTURE.md with all new patterns.
- **Acceptance Criteria**:
  - [ ] Section 10.3.6: `testdb.NewClosedSQLiteDB()` added to shared infra table
  - [ ] Section 10.2 or new subsection: "one test file per source file" rule with example
  - [ ] Section 8.1 or 11.2: "Handler DTOs must come from api/PRODUCT/models/models.gen.go"
  - [ ] Note clarifying `domain/` vs `model/` naming for GORM structs
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes

#### Task 5.2: Update instruction files

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Task 5.1
- **Description**: Propagate new rules to instruction files.
- **Acceptance Criteria**:
  - [ ] `03-02.testing.instructions.md`: No-local-createClosedDatabase rule added
  - [ ] `03-03.golang.instructions.md`: Handler DTOs from generated models rule added (or `02-04.openapi.instructions.md`)
  - [ ] Propagation markers consistent with ARCHITECTURE.md

#### Task 5.3: Update framework-v3 plan and tasks

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Phases 1-4 complete
- **Description**: Update framework-v3 documents to reference v2 as completed prerequisite and adjust affected tasks.
- **Acceptance Criteria**:
  - [ ] `docs/framework-v3/plan.md` header: "**Depends On**: `docs/framework-v2/` (complete)"
  - [ ] `docs/framework-v3/tasks.md` Phase 3 notes: jose-ja/sm-im/sm-kms test cleanup done in v2
  - [ ] `docs/framework-v3/tasks.md` Phase 6 Task 6.4: references `no_local_create_closed_database` rule established in v2

#### Task 5.4: Phase 5 quality gate (final)

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [fill when complete]
- **Dependencies**: Tasks 5.1-5.3
- **Description**: Final quality gate for entire plan.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go test ./... -shuffle=on` passes (zero regressions across all services)
  - [ ] `golangci-lint run ./...` clean
  - [ ] `go run ./cmd/cicd lint-fitness` passes (all new rules active)
  - [ ] `go run ./cmd/cicd lint-docs validate-propagation` passes
  - [ ] lessons.md updated with Phase 5 post-mortem
  - [ ] Git: all changes committed in semantic groups

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥95% coverage (production), ≥98% (infrastructure/utility)
- [ ] No skipped tests
- [ ] Race detector clean: `go test -race ./...`
- [ ] Test count before == test count after for all file merges

### Code Quality
- [ ] Linting passes across all modified packages
- [ ] No new TODOs without tracking
- [ ] Fitness rules pass: `go run ./cmd/cicd lint-fitness`

### Documentation
- [ ] ARCHITECTURE.md updated (Task 5.1)
- [ ] Instruction files updated (Task 5.2)
- [ ] framework-v3 updated (Task 5.3)

---

## Notes / Deferred Work

- **sm-kms server/middleware/**: Cataloged in Phase 4, NOT changed. Owned by framework-v3 D1.
- **sm-kms server/application/**: Audited in Phase 4. Either removed (if dead) or flagged for v3 Phase 3.
- **identity services**: Not in scope for v2. framework-v3 Phase 7/8 handles identity restructuring.
- **pki-ca**: Not in scope; its domain is still partial (framework-v3 Phase 8 Stage 4).

---

## Evidence Archive

- `test-output/framework-v2/sm-im-model-audit.md` - Task 3.1
- `test-output/framework-v2/sm-kms-application-audit.md` - Task 4.1
- `test-output/framework-v2/sm-kms-middleware-debt.md` - Task 4.2
- `test-output/framework-v2/phase1/` - testdb helper evidence
- `test-output/framework-v2/phase2/` - jose-ja cleanup evidence
- `test-output/framework-v2/phase3/` - sm-im cleanup evidence
- `test-output/framework-v2/phase4/` - sm-kms assessment evidence
- `test-output/framework-v2/phase5/` - knowledge propagation evidence
