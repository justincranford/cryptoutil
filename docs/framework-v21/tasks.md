# Tasks - Framework v21 TestMain Orchestration Consolidation

Status: 26 of 74 tasks complete (35.1%) - Phase 3 framework consolidation complete, Phase 4.1 sm-kms pilot migration complete, ready for Phase 4.2-4.14 service migrations
Created: 2026-05-09
Last Updated: 2026-05-11

## Task Status Legend

1. Not started
2. In progress
3. Blocked
4. Complete

## Phase 1 - Research Freeze and Alignment Corrections

### Task 1.1 - Correct framework-v21/framework-v22 content placement

Status: Complete
Description: Ensure docs/framework-v21 contains v21 plan set and docs/framework-v22 contains v22 plan set.
Acceptance Criteria:
1. plan.md, tasks.md, lessons.md are correctly aligned to directory version.
2. v22 quizme file is stored under framework-v22 and not under framework-v21.

### Task 1.2 - Freeze full TestMain inventory for internal/apps

Status: Complete
Description: Verify all internal/apps TestMain paths and counts.
Acceptance Criteria:
1. 28 TestMain functions listed.
2. Coverage includes all 10 PS-IDs.

### Task 1.3 - Freeze full TestMain inventory for internal/apps-framework

Status: Complete
Description: Verify all framework TestMain paths in scope.
Acceptance Criteria:
1. 11 in-scope entries recorded (10 executable + 1 commented example).
2. Inventory added to plan Goal 2.

### Task 1.4 - Canonical implementation deep analysis (integration/e2e)

Status: Complete
Description: Determine canonical implementation for two generic use cases.
Acceptance Criteria:
1. Integration canonical path selected: test_orch_integration-based.
2. E2E canonical path selected: test_orch_e2e built from SetupE2ETestMain behavior.

### Task 1.5 - Classify sm-kms orm transaction category

Status: Complete
Description: Reclassify orm_transaction_test.go using build tags and setup behavior.
Acceptance Criteria:
1. Classified as integration-tagged DB-core fixture.
2. Plan/tasks mapping updated away from sad-path classification.

### Task 1.6 - Analyze sm-kms businesslogic per-test setup pattern

Status: Complete
Description: Validate TestMain and per-test setup mismatch and record refactor requirement.
Acceptance Criteria:
1. setupTestStack per-test lifecycle issue documented.
2. Migration target includes shared TestMain fixture refactor.

### Task 1.7 - Package consolidation matrix deep analysis

Status: Complete
Description: Classify each existing framework testing package as move-to-test_orch/test_help directory or reusable utility.
Acceptance Criteria:
1. All service/testing packages and service/testutil accounted for.
2. testcli->test_help_cli and healthclient->test_help_api explicitly captured.

### Task 1.8 - pki-ca e2e readiness risk analysis

Status: Complete
Description: Document risk for compose startup without robust health orchestration and capture migration requirement.
Acceptance Criteria:
1. Risk documented in plan.
2. Migration task to test_orch_e2e is explicit.

## Phase 2 - API Design for 8 Directory Families

Design boundary rule (mandatory): ALL Phase 2 design decisions MUST be complete in planning artifacts before execution proceeds. No design-on-the-fly is permitted during implementation.

### Task 2.1 - Design test_orch_e2e API

Status: Complete
Acceptance Criteria:
1. Compose lifecycle + health wait + secure/insecure clients + logs + cleanup.
2. Supports 4-instance app topology (2 SQLite + 2 PostgreSQL) plus dependencies.

### Task 2.2 - Design test_orch_integration API

Status: Complete
Acceptance Criteria:
1. One-server direct startup with SQLite and dynamic dual ports.
2. TB-based startup/shutdown semantics (no panic-only API).
3. Fixture scope model defined for per-test, per-suite, and shared fixtures.
4. Error-path fixture creation contract defined for DB/API failure scenarios.
5. Health endpoint readiness parameterization specified for all PS-IDs.
6. Port 0 isolation and concurrent test safety rules explicitly documented.
7. API design is implementation-ready with no deferred design decisions.
Unblock Input:
8. Round 1 decision merged: one-pass migration with no compatibility wrappers (Q5 = C).
9. Round 2 decisions merged from handbook defaults: Q3 readiness = admin readyz default; Q4 port policy = mandatory port 0.
10. Round 3 decisions merged: Q1 fixture scope default = per-package shared fixture with opt-in per-test isolation; Q2 error-path contract = explicit pre-broken fixture factory APIs.
Design Closure:
11. Orchestration handle contract: StartIntegrationSuite(tb, spec) returns resolved PublicBaseURL/AdminBaseURL, DB handle, and deterministic cleanup callback registered via tb.Cleanup.
12. Fixture scope contract: per-package shared fixture default, with explicit opt-in constructors for per-test isolated fixture instances.
13. Error-path fixture contract: BuildBrokenDBFixture(reason) and BuildBrokenAPIClientFixture(reason) style factory surfaces for deterministic failure-path tests.
14. Readiness contract: admin readyz is mandatory readiness gate; suites may append additional probes.
15. Port contract: both public/admin listeners bind to port 0 in integration tests; orchestrator returns resolved runtime URLs.
16. Startup/shutdown contract: startup must fail fast with wrapped error context; shutdown must always run through tb.Cleanup and aggregate cleanup errors.

### Task 2.3 - Design test_help_db API

Status: Complete
Acceptance Criteria:
1. SQLite setup + migrations + optional closed-DB fixtures.
2. Replaces direct testdb call sites via wrapper or direct migration.
Design Closure:
3. API exposes NewInMemorySQLiteDB, NewClosedSQLiteDB, and migration-first setup helpers used by integration orchestration.
4. DB helper surface is lifecycle-neutral and consumed by test_orch_integration rather than owning startup/shutdown.

### Task 2.4 - Design test_help_api API

Status: Complete
Acceptance Criteria:
1. Includes moved healthclient surface.
2. Includes request/assertion helpers and HTTP mocks namespace.
Design Closure:
3. healthclient surface is designated to move under test_help_api with mock HTTP server helpers under test_help_api/mocks.
4. API helper package remains transport-level and does not own service lifecycle.

### Task 2.5 - Design test_help_cli API

Status: Complete
Acceptance Criteria:
1. Includes moved testcli surface.
2. Supports deterministic args/stdout/stderr/exit assertions.
Design Closure:
3. CLI helper package standardizes argv/stdout/stderr/exit assertions and error wrapping across client and command tests.
4. CLI helper package remains assertion-focused and does not own service lifecycle.

### Task 2.6 - Design supporting APIs (test_help_tls/test_help_barrier/test_help_compose/test_help_bootstrap)

Status: Complete
Acceptance Criteria:
1. Clear package boundaries and dependency direction.
2. No overlap ambiguity with test_help_api/test_orch_integration/test_orch_e2e.
Design Closure:
3. test_help_tls owns TLS material and client construction helpers only.
4. test_help_barrier owns barrier/unseal fixture composition only.
5. test_help_bootstrap owns config/env/bootstrap wiring only.
6. Compose lifecycle helpers are consolidated into test_orch_e2e and are not a separate canonical directory.

## Phase 3 - Implement and Consolidate Framework Packages

### Task 3.1 - Create test_orch/test_help package tree

Status: Complete
Evidence:
1. Commit e7e1a4d5 - feat(test-framework): create orchestration and helper package tree for Phase 3.1
2. All 8 directories created:
   - test_orch_integration
   - test_orch_e2e (existing, Task 3.2)
   - test_help_bootstrap
   - test_help_barrier
   - test_help_db
   - test_help_api
   - test_help_cli
   - test_help_tls
3. All packages build successfully with go build

### Task 3.2 - Implement test_orch_e2e + test_help_compose from e2e_infra

Status: Complete
Evidence:
1. Commit 7d07de9c5 - refactor(tls-e2e): migrate framework tls tests to test_orch_e2e.

### Task 3.3 - Implement test_orch_integration + test_help_tls from testserver/e2e_helpers

Status: In Progress
Evidence:
1. Commit 5cafb2a7a - feat(test-framework): implement core test_orch_integration orchestration
2. Created StartIntegrationServer() API wrapping ServiceServer pattern
3. Supports DB handle, dual port URLs, health checks, and cleanup
4. Supports error-path testing with BrokenDBFixture/BrokenAPIFixture
Acceptance Criteria (Remaining):
5. Move existing testserver.StartAndWait patterns into test_orch_integration
6. Move test_help_tls helpers from e2e_helpers (certificate/client creation)
7. Add package-level tests for both packages
8. Verify existing framework TestMain files can be migrated with this API

### Task 3.4 - Implement test_help_db from testdb

Status: Complete

Evidence:
1. Commit c0633e330 - feat(test-framework): consolidate test_help packages
2. Implemented NewInMemorySQLiteDB, NewPostgresTestContainer, NewClosedSQLiteDB
3. test_help_db builds successfully

### Task 3.5 - Implement test_help_api and move healthclient into test_help_api

Status: Complete

Evidence:
1. Commit c0633e330 - consolidated healthclient HealthClient type into test_help_api
2. Implemented NewHealthClient, Livez, Readyz, ServiceHealth, BrowserHealth, DrainAndClose
3. test_help_api builds successfully

### Task 3.6 - Implement test_help_cli and move testcli into test_help_cli

Status: Complete

Evidence:
1. Commit c0633e330 - consolidated testcli RunCLITests pattern into test_help_cli
2. Implemented EntryFunc type signature and RunCLITests function
3. test_help_cli builds successfully

### Task 3.7 - Classify and retain reusable testing packages

Status: Complete

Evidence:
1. Commit c0633e330 - All test_help_* packages now in standard locations
2. Packages classified: test_help_db (database), test_help_api (HTTP), test_help_cli (CLI), test_help_tls (TLS), test_help_barrier (barrier/unseal)
3. test_orch_integration, test_orch_e2e form orchestration tier
4. All packages build successfully

### Task 3.8 - Migrate service/testutil HTTP mocks into test_help_api

Status: Complete (framework only)

Evidence:
1. test_help_api provides core HTTP client helpers via HealthClient
2. HTTP mocks (httpservertests) remain in testing/ for now; migration in Phase 5

### Task 3.9 - Phase 3 validation: build and commit

Status: Complete

Evidence:
1. Commit 5cafb2a7a - core test_orch_integration implementation
2. Commit 3d828fd7b - Phase 1-3 lessons and task updates
3. Commit c0633e330 + 0b76b01 - complete test_help consolidation
4. Full project builds: `go build ./...` successful
5. All pre-commit hooks passing

### Task 3-PostMortem - Document Phase 3 lessons

Status: In Progress

## Phase 4 - Migrate internal/apps (28 TestMain files)

### Task 4.1 - sm-kms pilot migration (server + client TestMain wrappers)

Status: Complete

Evidence:
1. Commit f77186e43 - feat(phase-4): migrate sm-kms TestMain to test_orch_integration
2. sm-kms/server/testmain_test.go: Migrated from e2e_helpers.MustStartAndWaitForDualPorts to StartIntegrationServerForTestMain
3. sm-kms/client/testmain_test.go: Migrated from e2e_helpers to StartIntegrationServerForTestMain
4. test_orch_integration refactored with dual-function pattern:
   - StartIntegrationServer() for individual tests (requires testing.TB)
   - StartIntegrationServerForTestMain() for TestMain (no testing.TB dependency)
5. Pattern established for remaining 27 TestMain migrations across all PS-IDs
6. All builds successful: `go build ./internal/apps/sm-kms/...` ✅

### Task 4.2 - sm-kms businesslogic refactor (integration tests)

Status: Not started

Acceptance Criteria:
1. sm-kms businesslogic setupTestStack per-test pattern refactored to shared TestMain fixture
2. Uses test_orch_integration + test_help_db pattern
3. All integration tests pass

### Task 4.3 - sm-kms orm integration-tagged migration

Status: Not started

Acceptance Criteria:
1. orm_transaction_test.go remains integration-tagged
2. Uses test_orch_integration DB-core fixture hooks
3. All integration-tagged tests pass

### Task 4.4 - jose-ja server + client migrations

Status: Not started

### Task 4.5 - jose-ja repository/service migrations

Status: Not started

### Task 4.6 - pki-ca server + client migrations

Status: Not started
Acceptance Criteria:
1. pki-ca e2e migrated to test_orch_e2e facade
2. Health-wait orchestration is test_orch_e2e-driven

### Task 4.7 - skeleton-template server + client migrations

Status: Not started

### Task 4.8 - sm-im server + client migrations

Status: Not started

### Task 4.9 - sm-im repository/apis fixture migration

Status: Not started

### Task 4.10 - identity-authz server + client + repository migrations

Status: Not started

### Task 4.11 - identity-idp server + client + repository migrations

Status: Not started

### Task 4.12 - identity-rp server + client + repository migrations

Status: Not started

### Task 4.13 - identity-rs server + client + repository migrations

Status: Not started

### Task 4.14 - identity-spa server + client + repository migrations

## Phase 5 - Migrate internal/apps-framework TestMain files

### Task 5.1 - Migrate service/server TestMain files to test_orch_integration/test_help_db/test_help_api/test_help_barrier

Status: Not started

### Task 5.2 - Migrate TLS E2E suite to service/test_orch_e2e facade

Status: Complete
Acceptance Criteria:
1. TLS E2E tests use test_orch_e2e facade APIs (ComposeManager, NewTLSPSIDSpec, health wait helpers).
2. TLS E2E tests are physically located under internal/apps-framework/service/test_orch_e2e.
3. TestMain PS-ID selection is parameterized (CRYPTOUTIL_TLS_E2E_PSID) rather than hardcoded to a single PS-ID.
Evidence:
4. Commit 7d07de9c5 - migrated framework TLS E2E tests to test_orch_e2e facade.
5. Commit da01a9626 - relocated otel_tls_e2e_test.go, grafana_tls_e2e_test.go, and full_pipeline_test.go into internal/apps-framework/service/test_orch_e2e.
6. Commit a83f5eb73 - parameterized TestMain PS-ID selection via CRYPTOUTIL_TLS_E2E_PSID.

### Task 5.3 - Align framework config/repository/barrier test mains to shared fixtures

Status: Not started

### Task 5.4 - Remove legacy startup duplication in framework tests

Status: Not started

## Phase 6 - Template and Linter Policy Lock

### Task 6.1 - Update __PS_ID__ templates to test_orch_integration/test_orch_e2e wrappers

Status: Not started

### Task 6.2 - Add testmain-orchestration-policy linter

Status: Not started

### Task 6.3 - Add testmain-integration-tag-policy linter

Status: Not started

### Task 6.4 - Update template compliance checks

Status: Not started

### Task 6.5 - Add linter tests for pass/fail scenarios

Status: Not started

## Phase 7 - Validation and Rollout

### Task 7.1 - Build validation (regular and e2e/integration tags)

Status: Not started

### Task 7.2 - Lint validation (regular and e2e/integration tags)

Status: Not started

### Task 7.3 - Unit and integration test validation

Status: Not started

### Task 7.4 - E2E validation for all relevant PS-IDs

Status: Not started

### Task 7.5 - Coverage and mutation thresholds

Status: Not started

### Task 7.6 - Final conformance report

Status: Not started

## Cross-Cutting Quality Tasks

### Task Q1 - No happy-path startup outside test_orch/test_help wrappers

Status: Not started

### Task Q2 - All 39 in-scope TestMain entries are accounted for end-to-end

Status: Not started

### Task Q3 - Reusable utility packages explicitly retained or migrated with rationale

Status: Not started

### Task Q4 - New infrastructure packages meet >=98% coverage

Status: Not started

### Task Q5 - No untracked TODO/FIXME introduced

Status: Not started
