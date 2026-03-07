# Tasks - Framework v1

**Status**: 0 of 44 tasks complete (0%)
**Last Updated**: 2026-03-06
**Created**: 2026-03-06

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

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** - When unknowns discovered, blockers identified, unit/integration/E2E/mutations/fuzz/bench/race/SAST/DAST/load/any tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **Document root causes** - Root cause analysis is part of planning AND implementation, not optional; planning blockers must be resolved during planning, implementation blockers MUST be resolved during implementation
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues
- ✅ **NEVER de-prioritize quality** - Evidence-based verification is ALWAYS highest priority

**Rationale**: Maintaining maximum quality prevents cascading failures and rework.

---

## Task Checklist

### Phase 1: ServiceContract Interface

**Phase Objective**: Define and enforce a Go interface that all 10 services MUST satisfy, providing compile-time framework conformance.

#### Task 1.1: Audit All Service Method Signatures

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Produce a definitive matrix of methods across all 10 services, categorizing them as universal (in interface), common (most services), or service-specific (excluded from interface).
- **Acceptance Criteria**:
  - [ ] All 10 services surveyed: method name, signature, return types
  - [ ] Methods categorized: universal / common / service-specific
  - [ ] KMS divergences documented (Start/Shutdown signatures, IsReady vs SetReady)
  - [ ] Decision: which methods go in the core interface vs optional interfaces
- **Files**:
  - Evidence collected in `test-output/framework-v1/phase1/method-matrix.md`

#### Task 1.2: Define ServiceServer Interface

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Create the Go interface file with compile-time enforcement.
- **Acceptance Criteria**:
  - [ ] Interface defined in `internal/apps/template/service/server/contract.go`
  - [ ] Core interface covers universal methods (Start, Shutdown, DB, PublicPort, AdminPort, SetReady, PublicBaseURL, AdminBaseURL, PublicServerActualPort, AdminServerActualPort)
  - [ ] Optional interface(s) for common-but-not-universal methods (JWKGen, Telemetry, Barrier)
  - [ ] Documentation comments reference ARCHITECTURE.md Section 5.1
  - [ ] Interface is minimal — no methods that only 1-2 services need
  - [ ] File ≤300 lines
- **Files**:
  - `internal/apps/template/service/server/contract.go` (new)

#### Task 1.3: Add Compile-Time Assertions to All Services

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.2
- **Description**: Add `var _ ServiceServer = (*XxxServer)(nil)` compile-time assertions to all 10 services. Fix any signature mismatches.
- **Acceptance Criteria**:
  - [ ] All 9 standard services have `var _ ServiceServer = (*XxxServer)(nil)`
  - [ ] KMS handled appropriately (adapter or separate interface assertion)
  - [ ] `go build ./...` passes (compile-time proof)
  - [ ] Any signature fixes documented (e.g., adding missing methods)
- **Files**:
  - `internal/apps/skeleton/template/server/server.go` (modify)
  - `internal/apps/sm/im/server/server.go` (modify)
  - `internal/apps/sm/kms/server/server.go` (modify)
  - `internal/apps/jose/ja/server/server.go` (modify)
  - `internal/apps/pki/ca/server/server.go` (modify)
  - `internal/apps/identity/authz/server/server.go` (modify)
  - `internal/apps/identity/idp/server/server.go` (modify)
  - `internal/apps/identity/rp/server/server.go` (modify)
  - `internal/apps/identity/rs/server/server.go` (modify)
  - `internal/apps/identity/spa/server/server.go` (modify)

#### Task 1.4: Write Contract Interface Tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Write tests that verify all services satisfy the interface, including type assertions in test code.
- **Acceptance Criteria**:
  - [ ] Test file in `internal/apps/template/service/server/contract_test.go`
  - [ ] Table-driven test iterating over all implementations
  - [ ] `t.Parallel()` on all tests and subtests
  - [ ] Tests pass: `go test ./internal/apps/template/service/server/...`
  - [ ] Coverage ≥95% for the contract file
- **Files**:
  - `internal/apps/template/service/server/contract_test.go` (new)

#### Task 1.5: Phase 1 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 1.1-1.4
- **Description**: Run all quality gates for Phase 1 and collect evidence.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./internal/apps/template/service/server/...` passing
  - [ ] `golangci-lint run ./internal/apps/template/service/server/...` clean
  - [ ] No new TODOs
  - [ ] Evidence in `test-output/framework-v1/phase1/`
  - [ ] Git commit: `feat(framework): add ServiceContract interface`

---

### Phase 2: Simplified Builder Pattern

**Phase Objective**: Make core infrastructure automatic defaults in ServerBuilder so standard services need only domain-specific builder calls.

#### Task 2.1: Analyze Current Builder Defaults

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete
- **Description**: Map what Build() already auto-configures vs what requires explicit With*() calls. Identify which With*() calls can become defaults.
- **Acceptance Criteria**:
  - [ ] Current auto-configured items documented
  - [ ] Current explicit-required items documented
  - [ ] Each With*() method categorized: always-on / default-on / opt-in
  - [ ] KMS special needs identified (what it opts out of or customizes)
  - [ ] Evidence in `test-output/framework-v1/phase2/builder-analysis.md`

#### Task 2.2: Implement Builder Default Enhancement

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Modify `ServerBuilder.Build()` to auto-configure standard infrastructure. Add `WithoutBarrier()`, `WithoutSessions()` opt-out methods for exceptional services (if any truly need to opt out). Note: most With*() methods stay for backward compat, but calling them becomes optional.
- **Acceptance Criteria**:
  - [ ] `Build()` auto-configures: barrier, sessions, realm, registration, TLS, health
  - [ ] Existing With*() methods still work (backward compatible)
  - [ ] Services that don't call With*() still get standard infrastructure
  - [ ] KMS still builds correctly with its custom config
  - [ ] All existing tests pass (zero regressions)
  - [ ] File changes ≤500 lines per file
- **Files**:
  - `internal/apps/template/service/server/builder/server_builder.go` (modify)
  - `internal/apps/template/service/server/builder/server_builder_build.go` (modify)

#### Task 2.3: Simplify Standard Service NewFromConfig

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Update standard services (sm-im, jose-ja, skeleton, pki-ca, identity-*) to remove unnecessary explicit With*() calls, keeping only domain-specific calls.
- **Acceptance Criteria**:
  - [ ] Standard services use only: `WithDomainMigrations()` (if needed) + `WithPublicRouteRegistration()` + `WithSwaggerUI()` (if needed)
  - [ ] All removed With*() calls were actually auto-configured defaults
  - [ ] KMS keeps its custom With*() calls unchanged
  - [ ] All existing tests pass (zero regressions)
  - [ ] Each service's NewFromConfig is ≤30 lines
- **Files**:
  - `internal/apps/skeleton/template/server/server.go` (modify)
  - `internal/apps/sm/im/server/server.go` (modify)
  - `internal/apps/jose/ja/server/server.go` (modify)
  - `internal/apps/pki/ca/server/server.go` (modify)
  - `internal/apps/identity/authz/server/server.go` (modify)
  - `internal/apps/identity/idp/server/server.go` (modify)
  - `internal/apps/identity/rp/server/server.go` (modify)
  - `internal/apps/identity/rs/server/server.go` (modify)
  - `internal/apps/identity/spa/server/server.go` (modify)

#### Task 2.4: Phase 2 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.1-2.3
- **Description**: Run all quality gates for Phase 2 and collect evidence.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./internal/apps/...` passing (ALL services)
  - [ ] `golangci-lint run` clean
  - [ ] No regressions in any existing test
  - [ ] Evidence in `test-output/framework-v1/phase2/`
  - [ ] Git commit: `refactor(builder): auto-configure standard infrastructure defaults`

---

### Phase 3: air Live Reload

**Phase Objective**: Configure air for hot reload during development with one root config file.

#### Task 3.1: Create .air.toml Configuration

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: None (independent of other phases)
- **Description**: Create a single `.air.toml` at project root targeting `cmd/${SERVICE}/main.go`.
- **Acceptance Criteria**:
  - [ ] `.air.toml` at project root
  - [ ] `SERVICE` env var selects which service binary to build
  - [ ] Watches: `internal/`, `cmd/`, `pkg/`, `api/`
  - [ ] Excludes: `_test.go`, `tmp/`, `test-output/`, `docs/`, `.git/`, `node_modules/`
  - [ ] Rebuild command: `go build -o ./tmp/main ./cmd/${SERVICE}`
  - [ ] Run command: `./tmp/main server --dev`
  - [ ] Kill delay: 500ms for graceful shutdown
- **Files**:
  - `.air.toml` (new)

#### Task 3.2: Add air to .gitignore

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Ensure `tmp/` directory (air build output) is in `.gitignore`.
- **Acceptance Criteria**:
  - [ ] `tmp/` in `.gitignore` (if not already present)
  - [ ] `tmp/` not tracked by git
- **Files**:
  - `.gitignore` (modify if needed)

#### Task 3.3: Document air Usage

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Add air usage to project documentation.
- **Acceptance Criteria**:
  - [ ] Usage documented (install, run, env var override)
  - [ ] Example: `SERVICE=sm-im air` or `SERVICE=jose-ja air`
  - [ ] Prerequisites noted (Go install of air: `go install github.com/air-verse/air@latest`)
- **Files**:
  - Documentation file (existing or new dev setup docs)

#### Task 3.4: Phase 3 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1-3.3
- **Description**: Verify air config works and commit.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean (air config doesn't break build)
  - [ ] `.air.toml` parses correctly (`air` command starts without config errors)
  - [ ] Evidence in `test-output/framework-v1/phase3/`
  - [ ] Git commit: `feat(dx): add air live reload configuration`

---

### Phase 4: Architecture Fitness Functions

**Phase Objective**: Create `cicd lint-fitness` command with 8+ fitness sub-linters, integrated into pre-commit hooks and CI.

#### Task 4.1: Create lint_fitness Package Structure

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 complete (for service-contract-compliance check)
- **Description**: Create the `lint_fitness/` package following CICD command architecture (ARCHITECTURE.md Section 9.10).
- **Acceptance Criteria**:
  - [ ] Directory: `internal/apps/cicd/lint_fitness/`
  - [ ] Entry: `lint_fitness.go` with `Lint(logger)` function + `registeredLinters` slice
  - [ ] Follows same pattern as `lint_go/`, `lint_gotest/`, etc.
  - [ ] Empty linter slice initially (sub-linters added in subsequent tasks)
- **Files**:
  - `internal/apps/cicd/lint_fitness/lint_fitness.go` (new)

#### Task 4.2: Register lint-fitness in CICD Dispatch

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Register `lint-fitness` in the CICD command dispatch chain: magic constants, switch cases, imports.
- **Acceptance Criteria**:
  - [ ] `internal/shared/magic/magic_cicd.go` — Add `lint-fitness` to ValidCommands
  - [ ] `internal/apps/cicd/cicd.go` — Add case `"lint-fitness"` in switch, import lint_fitness package
  - [ ] `go run ./cmd/cicd lint-fitness` runs successfully (zero sub-linters, no errors)
  - [ ] `go run ./cmd/cicd all` includes lint-fitness in the linter run
- **Files**:
  - `internal/shared/magic/magic_cicd.go` (modify)
  - `internal/apps/cicd/cicd.go` (modify)

#### Task 4.3: Fitness Sub-Linter: Cross-Service Import Isolation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Generalize the existing `go-check-identity-imports` to all services. No service package may import another service's internal package.
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/cross_service_import_isolation/`
  - [ ] Checks: `internal/apps/<product>/<service>/` must not import `internal/apps/<other-product>/<other-service>/`
  - [ ] Template imports are allowed (all services import template)
  - [ ] Shared imports are allowed (all services import shared)
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase
- **Files**:
  - `internal/apps/cicd/lint_fitness/cross_service_import_isolation/cross_service_import_isolation.go` (new)
  - `internal/apps/cicd/lint_fitness/cross_service_import_isolation/cross_service_import_isolation_test.go` (new)

#### Task 4.4: Fitness Sub-Linter: Domain Layer Isolation

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Verify `domain/` packages don't import `server/`, `client/`, or `api/`.
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/domain_layer_isolation/`
  - [ ] Checks: `*/domain/` must not import `*/server/`, `*/client/`, `*/api/`
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase
- **Files**:
  - `internal/apps/cicd/lint_fitness/domain_layer_isolation/domain_layer_isolation.go` (new)
  - `internal/apps/cicd/lint_fitness/domain_layer_isolation/domain_layer_isolation_test.go` (new)

#### Task 4.5: Fitness Sub-Linter: File Size Limits

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Enforce ARCHITECTURE.md Section 11.2.6 file size limits (soft: 300, medium: 400, hard: 500).
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/file_size_limits/`
  - [ ] Reports: warning at >300 lines, error at >500 lines
  - [ ] Excludes: generated files (`*_gen.go`), test files, magic constants (`internal/shared/magic/`)
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Documents any existing violations (without failing) for future remediation
- **Files**:
  - `internal/apps/cicd/lint_fitness/file_size_limits/file_size_limits.go` (new)
  - `internal/apps/cicd/lint_fitness/file_size_limits/file_size_limits_test.go` (new)

#### Task 4.6: Fitness Sub-Linter: Health Endpoint Presence

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Verify all services register health endpoints.
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/health_endpoint_presence/`
  - [ ] Checks: each service under `internal/apps/` has references to livez, readyz, health paths
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase
- **Files**:
  - `internal/apps/cicd/lint_fitness/health_endpoint_presence/health_endpoint_presence.go` (new)
  - `internal/apps/cicd/lint_fitness/health_endpoint_presence/health_endpoint_presence_test.go` (new)

#### Task 4.7: Fitness Sub-Linter: TLS Minimum Version

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Verify all TLS configurations use TLS 1.3+ minimum.
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/tls_minimum_version/`
  - [ ] Checks: scan Go files for `tls.Config` with `MinVersion` < `tls.VersionTLS13`
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase
- **Files**:
  - `internal/apps/cicd/lint_fitness/tls_minimum_version/tls_minimum_version.go` (new)
  - `internal/apps/cicd/lint_fitness/tls_minimum_version/tls_minimum_version_test.go` (new)

#### Task 4.8: Fitness Sub-Linter: Admin Bind Address

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Verify all admin server bindings use 127.0.0.1 (not 0.0.0.0).
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/admin_bind_address/`
  - [ ] Checks: scan config files and Go code for admin bind address patterns
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase
- **Files**:
  - `internal/apps/cicd/lint_fitness/admin_bind_address/admin_bind_address.go` (new)
  - `internal/apps/cicd/lint_fitness/admin_bind_address/admin_bind_address_test.go` (new)

#### Task 4.9: Fitness Sub-Linter: Service Contract Compliance

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2, Phase 1 (ServiceContract interface exists)
- **Description**: Verify all services have `var _ ServiceServer = (*XxxServer)(nil)` compile-time assertion.
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/service_contract_compliance/`
  - [ ] Checks: scan each service's server.go for compile-time interface assertion
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase (after Phase 1)
- **Files**:
  - `internal/apps/cicd/lint_fitness/service_contract_compliance/service_contract_compliance.go` (new)
  - `internal/apps/cicd/lint_fitness/service_contract_compliance/service_contract_compliance_test.go` (new)

#### Task 4.10: Fitness Sub-Linter: Migration Range Compliance

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Verify template migrations are 1001-1999, domain migrations are 2001+.
- **Acceptance Criteria**:
  - [ ] Sub-linter: `internal/apps/cicd/lint_fitness/migration_range_compliance/`
  - [ ] Checks: template migration files are numbered 1001-1999, domain migration files are 2001+
  - [ ] Extends existing `migration_numbering` check in lint_go (complementary, not duplicate)
  - [ ] Tests: ≥95% coverage, table-driven, t.Parallel()
  - [ ] Passes on current codebase
- **Files**:
  - `internal/apps/cicd/lint_fitness/migration_range_compliance/migration_range_compliance.go` (new)
  - `internal/apps/cicd/lint_fitness/migration_range_compliance/migration_range_compliance_test.go` (new)

#### Task 4.11: Add Pre-Commit Hook for lint-fitness

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.3-4.10 (at least some sub-linters exist)
- **Description**: Add `go run ./cmd/cicd lint-fitness` to `.pre-commit-config.yaml`.
- **Acceptance Criteria**:
  - [ ] Hook added to `.pre-commit-config.yaml` following existing patterns
  - [ ] Hook runs on relevant file types (`.go`, `.sql`, `.yml`)
  - [ ] `pre-commit run lint-fitness --all-files` passes
- **Files**:
  - `.pre-commit-config.yaml` (modify)

#### Task 4.12: Phase 4 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1-4.11
- **Description**: Run all quality gates for Phase 4 and collect evidence.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./internal/apps/cicd/lint_fitness/...` passing (≥98% coverage for cicd utility)
  - [ ] `golangci-lint run` clean
  - [ ] `go run ./cmd/cicd lint-fitness` passes on current codebase
  - [ ] All 8 sub-linters have ≥95% test coverage
  - [ ] No new TODOs without tracking
  - [ ] Evidence in `test-output/framework-v1/phase4/`
  - [ ] Git commit: `feat(cicd): add lint-fitness command with 8 architecture fitness sub-linters`

---

### Phase 5: Shared Test Infrastructure

**Phase Objective**: Consolidate duplicated test setup patterns into shared packages, drastically reducing TestMain boilerplate across services.

#### Task 5.1: Audit Current Test Helpers

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None (independent of Phase 4)
- **Description**: Catalog all existing test helpers across services, identify duplication, and plan consolidation.
- **Acceptance Criteria**:
  - [ ] All test helper files listed across all services
  - [ ] Duplication identified (similar setup code in multiple TestMain functions)
  - [ ] Consolidation targets documented: what moves where
  - [ ] Evidence in `test-output/framework-v1/phase5/test-helper-audit.md`

#### Task 5.2: Create Shared Database Test Helpers

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1
- **Description**: Create shared helpers for SQLite in-memory DB setup and PostgreSQL test containers.
- **Acceptance Criteria**:
  - [ ] `NewInMemorySQLiteDB(t)` — Returns `*gorm.DB` with WAL mode, configured for tests
  - [ ] `NewPostgresTestContainer(ctx, t)` — Returns `*gorm.DB` with test container
  - [ ] Both helpers handle cleanup via `t.Cleanup()`
  - [ ] Tests: ≥98% coverage (infrastructure utility)
  - [ ] Located in appropriate package under `internal/apps/template/service/testing/`
- **Files**:
  - `internal/apps/template/service/testing/testdb/testdb.go` (new)
  - `internal/apps/template/service/testing/testdb/testdb_test.go` (new)

#### Task 5.3: Create Shared Server Test Helper

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2, Phase 1 (ServiceServer interface)
- **Description**: Create a helper that sets up a test server using port 0, SQLite in-memory, and all standard infrastructure.
- **Acceptance Criteria**:
  - [ ] `NewTestServer(t, opts...)` — Returns a running server for testing (port 0)
  - [ ] Uses `ServiceServer` interface as return type
  - [ ] Handles server shutdown via `t.Cleanup()`
  - [ ] Default opts: SQLite in-memory, dev TLS certs, port 0
  - [ ] Tests: ≥98% coverage (infrastructure utility)
- **Files**:
  - `internal/apps/template/service/testing/testserver/testserver.go` (new)
  - `internal/apps/template/service/testing/testserver/testserver_test.go` (new)

#### Task 5.4: Create Shared Fixture Helpers

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.2
- **Description**: Create helpers for common test fixtures (tenants, realms, users).
- **Acceptance Criteria**:
  - [ ] `CreateTestTenant(t, db)` — Creates and returns a test tenant
  - [ ] `CreateTestRealm(t, db, tenantID)` — Creates and returns a test realm
  - [ ] `CreateTestUser(t, db, tenantID, realmID)` — Creates and returns a test user
  - [ ] All use `googleUuid.NewV7()` for unique IDs (no hardcoded data)
  - [ ] Tests: ≥98% coverage (infrastructure utility)
- **Files**:
  - `internal/apps/template/service/testing/fixtures/fixtures.go` (new)
  - `internal/apps/template/service/testing/fixtures/fixtures_test.go` (new)

#### Task 5.5: Create Shared Assertion Helpers

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Create helpers for common test assertions (HTTP response validation, error format, health check).
- **Acceptance Criteria**:
  - [ ] `AssertHealthy(t, resp)` — Validates 200 OK health response
  - [ ] `AssertErrorResponse(t, resp, expectedCode)` — Validates error JSON (code, message)
  - [ ] `AssertTraceID(t, resp)` — Validates trace_id in response headers/body
  - [ ] `AssertJSONContentType(t, resp)` — Validates Content-Type: application/json
  - [ ] Tests: ≥98% coverage (infrastructure utility)
- **Files**:
  - `internal/apps/template/service/testing/assertions/assertions.go` (new)
  - `internal/apps/template/service/testing/assertions/assertions_test.go` (new)

#### Task 5.6: Create Health Client Helper

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.5
- **Description**: Create a reusable HTTPS client for testing health endpoints across services.
- **Acceptance Criteria**:
  - [ ] `NewHealthClient(baseURL)` — Returns configured HTTP client (TLS skip verify for test certs)
  - [ ] Methods: `Livez()`, `Readyz()`, `PublicHealth()`, `ServiceHealth()`, `BrowserHealth()`
  - [ ] Returns structured responses for easy assertion
  - [ ] Tests: ≥98% coverage (infrastructure utility)
- **Files**:
  - `internal/apps/template/service/testing/healthclient/healthclient.go` (new)
  - `internal/apps/template/service/testing/healthclient/healthclient_test.go` (new)

#### Task 5.7: Migrate Existing Services to Shared Helpers

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 4h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.2-5.6
- **Description**: Update existing service TestMain functions and test helpers to use the new shared packages.
- **Acceptance Criteria**:
  - [ ] At least 2-3 services migrated to use shared helpers (sm-im, skeleton, jose-ja)
  - [ ] Remaining services documented for future migration
  - [ ] All migrated tests pass
  - [ ] Net line reduction measured and documented
  - [ ] No regressions in any existing test

#### Task 5.8: Phase 5 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 5.1-5.7
- **Description**: Run all quality gates for Phase 5 and collect evidence.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./internal/apps/template/service/testing/...` passing (≥98% coverage)
  - [ ] All migrated services' tests still pass
  - [ ] `golangci-lint run` clean
  - [ ] Evidence in `test-output/framework-v1/phase5/`
  - [ ] Git commit: `feat(testing): add shared test infrastructure package`

---

### Phase 6: Cross-Service Contract Test Suite

**Phase Objective**: One test suite verifying ALL services behave consistently for core framework behavior.

#### Task 6.1: Design Contract Test Architecture

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Phase 1 (ServiceServer interface), Phase 5 (shared test helpers)
- **Description**: Design the contract test package structure and identify contracts to verify.
- **Acceptance Criteria**:
  - [ ] Package location decided: `internal/apps/template/service/testing/contract/`
  - [ ] API: `RunContractTests(t *testing.T, server ServiceServer)` designed
  - [ ] 8+ contracts identified (health, auth rejection, error format, trace_id, etc.)
  - [ ] Test execution strategy documented (SQLite in-memory, port 0, app.Test())
  - [ ] Evidence in `test-output/framework-v1/phase6/contract-design.md`

#### Task 6.2: Implement Health Contract Tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.1
- **Description**: Implement contract tests for health endpoints.
- **Acceptance Criteria**:
  - [ ] `/admin/api/v1/livez` returns 200 OK
  - [ ] `/admin/api/v1/readyz` returns 200 OK when ready, 503 when not
  - [ ] `/browser/api/v1/health` returns 200 OK (if registered)
  - [ ] `/service/api/v1/health` returns 200 OK (if registered)
  - [ ] Table-driven with t.Parallel()
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/apps/template/service/testing/contract/health_contracts.go` (new)
  - `internal/apps/template/service/testing/contract/health_contracts_test.go` (new)

#### Task 6.3: Implement Auth Contract Tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.1
- **Description**: Implement contract tests for authentication behavior.
- **Acceptance Criteria**:
  - [ ] `/service/api/v1/*` rejects unauthenticated requests with 401
  - [ ] Error response contains `code` and `message` fields
  - [ ] Table-driven with t.Parallel()
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/apps/template/service/testing/contract/auth_contracts.go` (new)
  - `internal/apps/template/service/testing/contract/auth_contracts_test.go` (new)

#### Task 6.4: Implement Error Format Contract Tests

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 6.1
- **Description**: Implement contract tests for error response format consistency.
- **Acceptance Criteria**:
  - [ ] All error responses contain `code` (string) and `message` (string)
  - [ ] All error responses include `requestId` (UUID)
  - [ ] All responses include `trace_id` in headers or body
  - [ ] Content-Type is application/json for all error responses
  - [ ] Table-driven with t.Parallel()
  - [ ] Tests: ≥95% coverage
- **Files**:
  - `internal/apps/template/service/testing/contract/error_contracts.go` (new)
  - `internal/apps/template/service/testing/contract/error_contracts_test.go` (new)

#### Task 6.5: Integrate Contract Tests with Services

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 3h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 6.2-6.4
- **Description**: Add `RunContractTests(t, server)` call to at least 3 services' existing test suites.
- **Acceptance Criteria**:
  - [ ] Contract tests integrated into sm-im, skeleton-template, jose-ja test suites
  - [ ] Contract tests pass for all integrated services
  - [ ] Remaining services documented for future integration
  - [ ] Evidence of behavioral consistency across services

#### Task 6.6: Phase 6 Quality Gate

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 6.1-6.5
- **Description**: Run all quality gates for Phase 6 and collect evidence.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./internal/apps/template/service/testing/contract/...` passing
  - [ ] All integrated services' contract tests passing
  - [ ] `golangci-lint run` clean
  - [ ] Evidence in `test-output/framework-v1/phase6/`
  - [ ] Git commit: `feat(testing): add cross-service contract test suite`

---

### Phase 7: Final Quality Gates & Evidence

**Phase Objective**: Verify ALL phases meet quality gates, collect comprehensive evidence, final commit.

#### Task 7.1: Full Build Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: All previous phases
- **Description**: Full build and test verification.
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `go build -tags e2e,integration ./...` clean
  - [ ] `go test ./...` passing (100%, zero skips)
  - [ ] `go test -race -count=2 ./...` clean
  - [ ] `golangci-lint run` clean
  - [ ] `golangci-lint run --build-tags e2e,integration` clean

#### Task 7.2: Coverage Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.1
- **Description**: Verify coverage targets.
- **Acceptance Criteria**:
  - [ ] New production code: ≥95% line coverage
  - [ ] New infrastructure code (test helpers, cicd): ≥98% line coverage
  - [ ] No coverage regressions in existing packages
  - [ ] Coverage report in `test-output/framework-v1/phase7/coverage/`

#### Task 7.3: Mutation Testing

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.2
- **Description**: Run mutation testing on new packages.
- **Acceptance Criteria**:
  - [ ] `gremlins unleash --tags=!integration` on new packages
  - [ ] ≥95% mutation score for production packages
  - [ ] ≥98% mutation score for infrastructure/utility packages
  - [ ] Results in `test-output/framework-v1/phase7/mutation/`

#### Task 7.4: Fitness Functions Self-Check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.1
- **Description**: Run the new fitness functions against the entire codebase.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd lint-fitness` passes with zero errors
  - [ ] All 8+ sub-linters executed
  - [ ] Results in `test-output/framework-v1/phase7/fitness-check.log`

#### Task 7.5: Pre-Commit Hook Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: [Fill when complete]
- **Dependencies**: Task 7.4
- **Description**: Verify pre-commit hooks work end-to-end.
- **Acceptance Criteria**:
  - [ ] `pre-commit run --all-files` passes
  - [ ] lint-fitness hook executes as part of pre-commit
  - [ ] Results in `test-output/framework-v1/phase7/pre-commit.log`

#### Task 7.6: Final Git Commit & Evidence Archive

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 7.1-7.5
- **Description**: Final commit with all evidence archived.
- **Acceptance Criteria**:
  - [ ] All evidence collected in `test-output/framework-v1/`
  - [ ] Plan.md updated with final status
  - [ ] Tasks.md updated with completion percentages and actual LOE
  - [ ] Clean working tree
  - [ ] Git commit: `docs(framework-v1): complete framework v1 implementation`

---

## Cross-Cutting Tasks

### Testing

- [ ] Unit tests ≥95% coverage (production), ≥98% (infrastructure/utility)
- [ ] Integration tests pass
- [ ] No skipped tests (except documented exceptions)
- [ ] Race detector clean: `go test -race -count=2 ./...`
- [ ] All tests use t.Parallel()
- [ ] All tests are table-driven (multi-case)
- [ ] No hardcoded UUIDs (use googleUuid.NewV7())

### Code Quality

- [ ] Linting passes: `golangci-lint run ./...` and `golangci-lint run --build-tags e2e,integration ./...`
- [ ] No new TODOs without tracking
- [ ] No security vulnerabilities
- [ ] Formatting clean: `gofumpt -s -w ./`
- [ ] Imports organized: `goimports -w ./`
- [ ] All files ≤500 lines

### Documentation

- [ ] air usage documented
- [ ] ServiceContract interface documented
- [ ] Fitness functions documented (what each checks)
- [ ] Shared test infrastructure usage documented

### Deployment

- [ ] No deployment changes required (all changes are build-time/test-time)
- [ ] Pre-commit hooks updated for lint-fitness

---

## Notes / Deferred Work

### Items Explicitly Excluded (User Decision)

- P0-3: Skeleton CRUD reference implementation — Not needed with contract + fitness enforcement
- P1-1: cicd new-service scaffolding — Only 9 services exist, unlikely to add more
- P1-3: cicd diff-skeleton conformance — Superseded by fitness functions
- P2-1: Service manifest declaration — Replaced by simplified builder defaults
- P2-3: OpenAPI-to-Repository codegen — Not wanted
- P3-1: Module system (fx/Wire) — Overkill
- P3-2: Extract framework module — Premature

### Future Considerations

- If a 10th+ service is added, revisit P1-1 scaffolding decision
- If fitness function false positives are high, add severity levels (warn vs error)
- Consider moving more lint_go/lint_gotest checks into lint-fitness for cleaner organization

---

## Evidence Archive

- `test-output/framework-v1/phase1/` - ServiceContract interface evidence
- `test-output/framework-v1/phase2/` - Builder simplification evidence
- `test-output/framework-v1/phase3/` - air live reload evidence
- `test-output/framework-v1/phase4/` - Fitness functions evidence
- `test-output/framework-v1/phase5/` - Shared test infrastructure evidence
- `test-output/framework-v1/phase6/` - Contract test suite evidence
- `test-output/framework-v1/phase7/` - Final quality gates evidence
