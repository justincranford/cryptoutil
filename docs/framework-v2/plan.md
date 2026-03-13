# Implementation Plan - Framework v2: Service Code Quality Refactoring

**Status**: Planning
**Created**: 2026-03-12
**Last Updated**: 2026-03-12
**Depends On**: `docs/framework-v1/` (complete)
**Prerequisite For**: `docs/framework-v3/` (v3 phases 3, 4, 6, 7 build on v2 outcomes)
**Purpose**: Systematic code-quality and structural refactoring of three mature services (sm-im, jose-ja, sm-kms) to eliminate patterns that accumulated before service-template existed. Establishes correct target structure for all future services. Runs BEFORE framework-v3 to remove tech debt that would otherwise complicate v3's builder/fitness/extraction work.

---

## Companion Documents

1. **plan.md** (this file) - phases, objectives, decisions
2. **tasks.md** - task checklist per phase
3. **lessons.md** - persistent memory: what worked, what did not, root causes, patterns

---

## Context: Why This Exists

sm-im was implemented first, then used to shape jose-ja, and sm-kms was the original manual implementation later AI-migrated to service-template. All three carry patterns that were reasonable at the time but are now wrong given the service-template and testdb infrastructure that exists.

### Identified Problems

#### Problem 1: Duplicated `createClosedDatabase()` helper (all 3 services)

Every service that wants to test repository error paths re-implements the same boilerplate:
open SQLite → PRAGMA WAL → PRAGMA busy_timeout → GORM → apply migrations → close connection → return broken GORM handle.

| Service | File | Function |
|---------|------|----------|
| jose-ja | `repository/database_error_test.go` | `createClosedDatabase()` |
| jose-ja | `service/database_error_test.go` | `createClosedServiceDependencies()` |
| sm-im | `server/apis/messages_dberror_test.go` | `createClosedDBHandler()` |
| sm-im | `server/apis/messages_errorpaths_test.go` | `createMixedHandler()` |

This belongs in `internal/apps/template/service/testing/testdb` as `NewClosedSQLiteDB(t, applyMigrations)`.

#### Problem 2: Hand-rolled handler DTOs instead of generated models (jose-ja)

`internal/apps/jose/ja/server/apis/jwk_handler.go` defines its own request/response structs
(`CreateElasticJWKRequest`, `ElasticJWKResponse`, `MaterialJWKResponse`) instead of using
`api/jose/models/models.gen.go`. The OpenAPI contract is already code-generated and correct.
Handlers must map between generated DTOs ↔ domain layer; they must NOT invent new DTOs.

sm-kms does this correctly (`cryptoutilKmsServer` from `api/kms/server`). sm-im needs verification.

#### Problem 3: File proliferation via error-path file splitting (jose-ja, sm-kms)

Error-path coverage was achieved by creating separate files per error scenario instead of
table-driven subtests in the same file as the domain component under test:

**jose-ja repository/**:
- `database_error_test.go` - ElasticJWK repo errors
- `database_error_material_test.go` - MaterialJWK repo errors
- `database_error_audit_test.go` - Audit repo errors
- `additional_edge_cases_test.go`, `audit_log_list_test.go` - more splits

**jose-ja service/**:
- `database_error_test.go`, `database_error_corrupt_test.go`, `database_error_corrupt2_test.go`
- `database_error_extra_test.go`, `database_error_jwe_test.go`
- `error_coverage_jwe_jws_test.go`, `error_coverage_jwks_rotation_test.go`, `error_coverage_jwt_test.go`
- `mapping_functions_test.go`, `mapping_functions_parse_test.go`

**sm-kms repository/orm/**:
- `business_entities_additional_errors_test.go`, `business_entities_dead_code_test.go`
- `business_entities_error_mapping_test.go`, `business_entities_get_errors_test.go`
- `business_entities_gorm_errors_test.go`, `business_entities_materialkey_errors_test.go`
- `business_entities_postgres_errors_test.go`, `business_entities_toapperr_test.go`
- `business_entities_update_errors_test.go`, +more

Each domain file (`elastic_jwk_service.go`) should have exactly ONE test file (`elastic_jwk_service_test.go`) containing ALL test cases including error paths as table-driven subtests. Extra files bloat the package and make navigation harder.

#### Problem 4: domain/models.go is a GORM file, not a domain file (jose-ja)

`internal/apps/jose/ja/domain/models.go` contains GORM struct definitions with table annotations
(`gorm:"type:text;primaryKey"`, `TableName()` methods, etc.). These are persistence layer types,
not domain types. In a service with NO separate service/ layer, the domain/ package can hold them
ONLY if the package is renamed to reflect its actual nature (e.g., `model/` or `repository/model/`).
The jose-ja domain/ package currently has GORM types co-located with operation constants — this
needs clarification of intent given v3 D3 (thin wrappers inject only domain logic).

#### Problem 5: sm-kms has pre-template application layer (scope boundary investigation)

sm-kms has `server/application/` with `application_basic.go`, `application_core.go`,
`application_init.go`, `application_listener*.go`, `fiber_middleware_otel_request_logger.go`.
This looks like infrastructure that predates service-template. After framework-v1 builder
migration, this may be dead code or it may still be wired in.

**v2 action**: Audit only. If dead → remove. If still active → flag as v3 Phase 3 work
(builder refactoring eliminates service-owned application layers per D3).

#### Problem 6: sm-kms has auth middleware (v3-owned, do NOT touch)

sm-kms `server/middleware/` contains: JWT validation, claims extraction, revocation checking,
realm context, scopes enforcement, session handling, tenant extraction. Per v3 D1, auth is
100% service-template owned. v2 MUST NOT touch these — they are v3 Phase 3/4/5 scope.

---

## Guiding Principles / Decisions

### D1: testdb.NewClosedSQLiteDB is the ONLY accepted pattern

No service package may contain its own `createClosedDatabase`-style helper function.
All closed-DB error path testing uses `testdb.NewClosedSQLiteDB(t, applyMigrations)`.
After v2, a lint-fitness rule enforces this (see v3 Phase 6 dependency below).

### D2: Handler types come from api/PRODUCT/models/models.gen.go only

Handler request/response structs must NOT be hand-rolled in service packages.
The generated models are the API contract. Handlers map: generated DTO ↔ domain struct.

### D3: One test file per source file, all cases in one file

`elastic_jwk_repository.go` → `elastic_jwk_repository_test.go` only.
Error paths, edge cases, closed-DB paths are all table-driven subtests in the same file.
No `_error_test.go`, `_edge_cases_test.go`, `_corrupt_test.go` split files.

### D4: domain/ packages contain ZERO GORM annotations

If a `domain/` package contains GORM struct tags or `TableName()`, it is a persistence
package and should be named accordingly (e.g., `repository/model/` or just `model/`).
True domain types are API+business agnostic (no GORM, no fiber, no generated models).

### D5: sm-kms middleware is v3-owned, not v2

v2 MUST NOT refactor `server/middleware/`. Touching it in v2 would conflict with v3
D1 (auth service-template migration) and create churn. v2 documents the debt and flags it.

### D6: No scope creep into v3 territory

v2 scope = test infrastructure helpers, file naming, file count reduction, generated models.
v2 MUST NOT: change builder API, move middleware, change auth flows, touch identity services.

---

## v2 ↔ v3 Relationship Analysis

### Overlap Inventory

| v2 Work | v3 Phase | Relationship | Risk |
|---------|----------|-------------|------|
| `testdb.NewClosedSQLiteDB` helper | v3 P6 Task 6.4 (test infra fitness rule) | v2 establishes pattern; v3 enforces via linter | **Low** - pattern defined in v2, rule added in v3 |
| jose-ja handler uses generated models | v3 D3 (thin wrappers) | v2 aligns with v3 goal | **Low** - no conflict, v2 reduces v3 work |
| File proliferation cleanup (jose-ja service/) | v3 P3 (builder refactoring touches same files) | v3 changes registration; v2 changes test organization | **Medium** - same files touched; **v2 must complete before v3 Phase 3** |
| sm-kms `server/application/` audit | v3 P3 (builder refactoring) | v2 audits; v3 removes | **Low** - v2 only reads; v3 writes |
| sm-kms `server/middleware/` | v3 D1 + P3/P4 | v2 MUST NOT touch; v3 owns | **Zero** if v2 respects D5 above |
| jose-ja `domain/` naming | v3 D3 + P7 (extract/reintegrate) | v2 clarifies intent; v3 may extract | **Low** - naming fix is non-breaking |

### Conflicts to Avoid

1. **v2 MUST complete jose-ja service/ test cleanup BEFORE v3 Phase 3** begins builder refactoring.
   Builder refactoring changes how routes are registered in the same handler files v2 is cleaning.
   Parallel work = merge conflicts.

2. **v2 MUST NOT change sm-kms server/middleware/ filenames or move files.**
   v3 D1 migration will delete these entire packages. File renaming creates ghost merges.

3. **v2's handler DTO fix (jose-ja) establishes the pattern for v3 Phase 7/8 reintegration.**
   Document the mapping pattern clearly so identity service reintegration adopts the same approach.

### v3 Adjustments Required After v2 Completes

1. v3 `tasks.md` Phase 3 should note: "jose-ja service/ test cleanup completed in v2, no test migration needed"
2. v3 `tasks.md` Phase 6 Task 6.4: "Enforce testdb.NewClosedSQLiteDB pattern (established in v2)"
3. v3 `plan.md` header: "**Depends On**: `docs/framework-v2/` (complete)"

### Opportunities: Pull v3 Work into v2 (Evaluated)

| v3 Item | Pull into v2? | Decision | Rationale |
|---------|--------------|----------|-----------|
| v3 P6 T6.4: Add test infrastructure fitness rule | YES — partial | Add the `new_closed_sqlite_db_pattern` fitness rule in v2 Phase 1 after helper is established | Rule is tiny, test pattern is defined in v2, natural fit |
| v3 D3: sm-kms `server/application/` removal | NO | v2 audits; v3 Phase 3 removes (builder context needed) | Needs builder context from v3 P3 |
| v3 D1: sm-kms `server/middleware/` migration | NO | Too large, wrong context | v3 owns all auth infra |
| v3 P2: InsecureSkipVerify removal | NO | Out of scope | TLS work is separate concern |

---

## Technical Context

- **Language**: Go 1.26.1
- **Framework**: Fiber v2, service-template builder, GORM
- **Database**: SQLite in-memory (tests), PostgreSQL (production)
- **Generated Code**: `api/PRODUCT/models/models.gen.go` via oapi-codegen
- **Test DB Helper**: `internal/apps/template/service/testing/testdb/testdb.go`
- **Related Plans**: `docs/framework-v3/` (downstream — must coordinate)

---

## Phases

### Phase 1: testdb.NewClosedSQLiteDB Helper (0.5d) [Status: TODO]

**Objective**: Add `NewClosedSQLiteDB(t, applyMigrations)` to service-template testdb package. This is infrastructure work that unlocks all three service cleanups.

- Add `NewClosedSQLiteDB(t *testing.T, applyMigrations func(*sql.DB) error) *gorm.DB` to `testdb/testdb.go`
- Uses same open → PRAGMA WAL → PRAGMA busy_timeout → GORM → apply migrations → close pattern
- Returns `*gorm.DB` with closed underlying connection (for error path tests)
- Cleanup registered via `t.Cleanup()` (noop - DB already closed, but consistent)
- Unit tests: ≥98% coverage (infrastructure utility)
- Add lint-fitness rule `no_local_create_closed_database` (checks for private `createClosedDatabase`, `createClosedDB`, `createClosedServiceDependencies` functions in non-testdb packages)
- Document in ARCHITECTURE.md Section 10.3.6 Shared Test Infrastructure table
- **Success**: Build + tests clean; fitness rule passes on new helper; fitness rule FAILS on current jose-ja (confirmed before cleanup)
- **Post-Mortem**: Update lessons.md.

### Phase 2: jose-ja Cleanup (1.5d) [Status: TODO]

**Objective**: Fix all four identified problems in jose-ja (handler DTOs, closed-DB helpers, file proliferation, domain naming).

- **2.1 Handler DTOs**: Replace hand-rolled `CreateElasticJWKRequest`, `ElasticJWKResponse`, `MaterialJWKResponse` in `server/apis/jwk_handler.go` with types from `api/jose/models/models.gen.go`. Add explicit mapping functions (`toElasticJWKResponse`, `toMaterialJWKResponse`) in the handler.
- **2.2 Migrate closed-DB helpers**: Replace `createClosedDatabase()` (repository) and `createClosedServiceDependencies()` (service) with `testdb.NewClosedSQLiteDB()`.
- **2.3 Merge repository error-path files**: Consolidate `database_error_test.go`, `database_error_material_test.go`, `database_error_audit_test.go`, `additional_edge_cases_test.go`, `audit_log_list_test.go` → error-path subtests in `elastic_jwk_repository_test.go`, `material_jwk_repository_test.go`, `audit_repository_test.go`.
- **2.4 Merge service error-path files**: Consolidate `database_error_test.go`, `database_error_corrupt_test.go`, `database_error_corrupt2_test.go`, `database_error_extra_test.go`, `database_error_jwe_test.go`, `error_coverage_jwe_jws_test.go`, `error_coverage_jwks_rotation_test.go`, `error_coverage_jwt_test.go`, `mapping_functions_test.go`, `mapping_functions_parse_test.go` → subtests in `elastic_jwk_service_test.go`, `jwe_service_test.go`, `jws_service_test.go`, `jwt_service_test.go`, etc.
- **2.5 domain/ naming**: Rename `internal/apps/jose/ja/domain/` → `internal/apps/jose/ja/model/` and update all import aliases. (GORM structs belong in a persistence-aware package, not "domain".)
- **2.6 Quality gates**: lint + tests + coverage ≥95% + fitness rule passes.
- **Success**: jose-ja repository/ ≤5 test files; service/ ≤1 test file per service file; handler uses generated models; fitness rule passes.
- **Post-Mortem**: Update lessons.md.

### Phase 3: sm-im Cleanup (1d) [Status: TODO]

**Objective**: Apply the same cleanup to sm-im (fewer issues than jose-ja, no hand-rolled DTOs).

- **3.1 Verify generated models**: Confirm sm-im handlers use `api/sm/im/models/models.gen.go` (or equivalent). Document result.
- **3.2 Migrate closed-DB helpers**: Replace `createClosedDBHandler()` (`server/apis/messages_dberror_test.go`) and `createMixedHandler()` (`server/apis/messages_errorpaths_test.go`) with `testdb.NewClosedSQLiteDB()` + inline service construction.
- **3.3 Merge repository error-path files**: Evaluate `repository/error_paths_test.go`, `repository/error_returns_test.go`, `repository/concurrent_access_test.go` — merge into `message_repository_test.go` and `message_recipient_jwk_repository_test.go` as table subtests.
- **3.4 Merge handler error-path files**: Evaluate `server/apis/messages_dberror_test.go` and `server/apis/messages_errorpaths_test.go` — merge into `messages_test.go` where appropriate.
- **3.5 domain/ audit**: Confirm sm-im `domain/` contains true domain types (no GORM tags). Document.
- **3.6 Quality gates**: lint + tests + coverage ≥95% + fitness rule passes.
- **Success**: No createClosedDB* functions in sm-im; repository/ and server/apis/ have ≤1 test file per source file; fitness rule passes.
- **Post-Mortem**: Update lessons.md.

### Phase 4: sm-kms Assessment and Safe Cleanup (1d) [Status: TODO]

**Objective**: Audit sm-kms, remove clearly dead code, flag v3-owned items. Do NOT touch middleware or auth.

- **4.1 server/application/ audit**: Compare `server/application/` to what the service-template builder now provides. Determine if it's dead (builder replaced it) or still active. If dead → remove with tests. If active → document as tech debt with v3 Phase 3 as the fix.
- **4.2 server/middleware/ documentation only**: Catalog all middleware files, confirm they replicate service-template auth functionality. Create `docs/framework-v2/sm-kms-middleware-debt.md` documenting what moves where in v3. DO NOT CHANGE CODE.
- **4.3 repository/orm/ file proliferation**: Same pattern as jose-ja. Evaluate merging split error-path files. Apply D3 rule (one test file per source file). Migrate any closed-DB helpers to `testdb.NewClosedSQLiteDB()`.
- **4.4 Verify generated models**: Confirm handler uses `api/kms/server` generated types (expect: already correct). Document.
- **4.5 Quality gates**: lint + tests + coverage ≥95% + fitness rule passes for touched packages.
- **Success**: server/application/ either removed or documented; repository/orm/ file count reduced by ≥30%; no custom closed-DB helpers; fitness rule passes.
- **Post-Mortem**: Update lessons.md.

### Phase 5: Knowledge Propagation (0.5d) [Status: TODO]

**Objective**: Apply lessons to permanent artifacts. Never skip this phase.

- Review lessons.md from all prior phases
- Update ARCHITECTURE.md:
  - Section 10.3.6: Add `testdb.NewClosedSQLiteDB()` to shared test infrastructure table
  - Section 10.2 or 11.2: Document "one test file per source file" rule with example
  - Section 4.4.1 or 3.3: Document that handler types MUST come from `api/PRODUCT/models/models.gen.go`
  - Section 11.2: Document `domain/` naming vs `model/` (GORM structs go in persistence package)
- Update `03-02.testing.instructions.md`: Add "no local createClosedDatabase" rule
- Update `03-03.golang.instructions.md`: Add "no hand-rolled handler DTOs" rule
- Update framework-v3 plan.md + tasks.md: Add "Depends On: framework-v2" header and task notes
- Propagation check: `go run ./cmd/cicd lint-docs validate-propagation`
- **Success**: All artifact updates committed; propagation check passes; v3 plan updated.
- **Post-Mortem**: Update lessons.md.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Merging jose-ja service/ test files misses a test case | Medium | High | Verify line-count and test-count before/after merge; run with -v and compare |
| Handler DTO fix breaks jose-ja API contract (wrong field names/types) | Medium | High | Compare generated model fields against hand-rolled structs before removing; add mapping tests |
| sm-kms server/application/ is still active (not dead code) | Medium | Medium | Audit via call graph before touching; if active, document only |
| v3 Phase 3 starts before v2 Phase 2 completes | Low | Medium | Coordinate start; v3 Phase 3 should reference v2 completion as prerequisite |
| Fitness rule false-positives in testdb package itself | Low | Low | Add exclusion for `testdb` package in fitness rule |

---

## Quality Gates - MANDATORY

**Per-Phase Gates**:
- ✅ `go build ./...` clean
- ✅ `go build -tags e2e,integration ./...` clean
- ✅ `golangci-lint run` clean
- ✅ `golangci-lint run --build-tags e2e,integration` clean
- ✅ `go test ./... -shuffle=on` passes (100%, zero skips)
- ✅ Coverage maintained or improved (production ≥95%, infra ≥98%)
- ✅ Fitness rules pass: `go run ./cmd/cicd lint-fitness`
- ✅ `go test -race -count=2 ./...` clean

---

## Success Criteria

- [ ] `testdb.NewClosedSQLiteDB()` exists in service-template testing package with ≥98% coverage
- [ ] Zero `createClosedDatabase`/`createClosedDB`/`createClosedServiceDependencies` functions outside testdb package (fitness rule enforced)
- [ ] jose-ja `server/apis/jwk_handler.go` imports `api/jose/models` — zero hand-rolled DTOs
- [ ] jose-ja `repository/` and `service/` have ≤1 test file per source file
- [ ] sm-im has zero custom closed-DB helpers
- [ ] sm-kms `server/application/` either removed or documented as v3 tech debt
- [ ] ARCHITECTURE.md updated with new patterns
- [ ] framework-v3 plan updated to reference v2 as prerequisite
- [ ] All fitness rules pass

---

## ARCHITECTURE.md Cross-References

| Topic | Section | When to Reference |
|-------|---------|-------------------|
| Testing Strategy | [Section 10](../../docs/ARCHITECTURE.md#10-testing-architecture) | ALL phases |
| Shared Test Infrastructure | [Section 10.3.6](../../docs/ARCHITECTURE.md#1036-shared-test-infrastructure) | Phase 1 (new helper) |
| Unit Testing (file per test) | [Section 10.2](../../docs/ARCHITECTURE.md#102-unit-testing-strategy) | Phase 2-4 (file merging) |
| Quality Gates | [Section 11.2](../../docs/ARCHITECTURE.md#112-quality-gates) | ALL phases (mandatory) |
| Coding Standards | [Section 13.1](../../docs/ARCHITECTURE.md#131-coding-standards) | Phase 2-4 (file naming) |
| OpenAPI-First Design | [Section 8.1](../../docs/ARCHITECTURE.md#81-openapi-first-design) | Phase 2.1 (handler DTOs) |
| Service Template Pattern | [Section 5.1](../../docs/ARCHITECTURE.md#51-service-template-pattern) | Phase 1, 4 |
| Infrastructure Blockers | [Section 13.7](../../docs/ARCHITECTURE.md#137-infrastructure-blocker-escalation) | ALL phases |
| Post-Mortem & Knowledge Propagation | [Section 13.8](../../docs/ARCHITECTURE.md#138-phase-post-mortem--knowledge-propagation) | Phase 5 (mandatory) |
| Plan Lifecycle | [Section 13.6](../../docs/ARCHITECTURE.md#136-plan-lifecycle-management) | ALL (mandatory) |
