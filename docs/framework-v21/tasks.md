# Tasks - Framework v21 TestMain Orchestration Consolidation

Status: 11 of 74 tasks complete (14.9%)
Created: 2026-05-09
Last Updated: 2026-05-09

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

## Phase 2 - API Design for 9 Directory Families

Design boundary rule (mandatory): ALL Phase 2 design decisions MUST be complete in planning artifacts before execution proceeds. No design-on-the-fly is permitted during implementation.

### Task 2.1 - Design test_orch_e2e API

Status: Complete
Acceptance Criteria:
1. Compose lifecycle + health wait + secure/insecure clients + logs + cleanup.
2. Supports 4-instance app topology (2 SQLite + 2 PostgreSQL) plus dependencies.

### Task 2.2 - Design test_orch_integration API

Status: In progress
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
10. `docs/framework-v21/quizme-v3.md` must be answered and merged to finalize only Q1-Q2 policies and close remaining design ambiguity.

### Task 2.3 - Design test_help_db API

Status: Not started
Acceptance Criteria:
1. SQLite setup + migrations + optional closed-DB fixtures.
2. Replaces direct testdb call sites via wrapper or direct migration.

### Task 2.4 - Design test_help_api API

Status: Not started
Acceptance Criteria:
1. Includes moved healthclient surface.
2. Includes request/assertion helpers and HTTP mocks namespace.

### Task 2.5 - Design test_help_cli API

Status: Not started
Acceptance Criteria:
1. Includes moved testcli surface.
2. Supports deterministic args/stdout/stderr/exit assertions.

### Task 2.6 - Design supporting APIs (test_help_tls/test_help_barrier/test_help_compose/test_help_bootstrap)

Status: Not started
Acceptance Criteria:
1. Clear package boundaries and dependency direction.
2. No overlap ambiguity with test_help_api/test_orch_integration/test_orch_e2e.

## Phase 3 - Implement and Consolidate Framework Packages

### Task 3.1 - Create test_orch/test_help package tree

Status: Not started

### Task 3.2 - Implement test_orch_e2e + test_help_compose from e2e_infra

Status: Complete
Evidence:
1. Commit 7d07de9c5 - refactor(tls-e2e): migrate framework tls tests to test_orch_e2e.

### Task 3.3 - Implement test_orch_integration + test_help_tls from testserver/e2e_helpers

Status: Not started

### Task 3.4 - Implement test_help_db from testdb

Status: Not started

### Task 3.5 - Implement test_help_api and move healthclient into test_help_api

Status: Not started
Acceptance Criteria:
1. healthclient API relocated under test_help_api.
2. No compatibility wrapper retained; callers migrate directly in the same phase.

### Task 3.6 - Implement test_help_cli and move testcli into test_help_cli

Status: Not started
Acceptance Criteria:
1. testcli API relocated under test_help_cli.
2. No compatibility wrapper retained; callers migrate directly in the same phase.

### Task 3.7 - Migrate service/testutil HTTP mocks into test_help_api/mocks

Status: Not started

### Task 3.8 - Classify and keep reusable utility packages (assertions/httpservertests/fixtures/stubs)

Status: Not started
Acceptance Criteria:
1. Explicitly retained and documented as reusable beyond orchestration.
2. Imported by new test_* modules where appropriate.

### Task 3.9 - Remove or repurpose empty contract package

Status: Not started

### Task 3.10 - Package-level tests and one-pass call-site cutover

Status: Not started

## Phase 4 - Migrate internal/apps (28 TestMain files)

### Task 4.1 - Identity server TestMain wrappers to test_orch_integration (5 files)

Status: Not started

### Task 4.2 - Identity e2e TestMain wrappers to test_orch_e2e (5 files)

Status: Not started

### Task 4.3 - jose-ja migrations (e2e/server/repository/service)

Status: Not started

### Task 4.4 - pki-ca migrations (e2e/server)

Status: Not started
Acceptance Criteria:
1. pki-ca e2e no longer uses custom compose start/stop flow.
2. Health-wait orchestration is test_orch_e2e-driven.

### Task 4.5 - skeleton-template migrations (e2e/server)

Status: Not started

### Task 4.6 - sm-im server migration from local SetupTestServer to test_orch_integration

Status: Not started

### Task 4.7 - sm-im client migration to test_orch_integration

Status: Not started

### Task 4.8 - sm-im repository/apis fixture migration to test_help_db/test_help_api/test_help_barrier

Status: Not started

### Task 4.9 - sm-kms server migration to test_orch_integration

Status: Not started

### Task 4.10 - sm-kms e2e migration to test_orch_e2e facade

Status: Not started

### Task 4.11 - sm-kms client migration from e2e_helpers to test_orch_integration

Status: Not started

### Task 4.12 - sm-kms businesslogic refactor to shared TestMain fixture

Status: Not started
Acceptance Criteria:
1. setupTestStack per-test heavy wiring eliminated or reduced behind shared fixture.
2. TestMain pattern drives shared lifecycle.

### Task 4.13 - sm-kms orm integration-tagged migration to integration DB-core fixture

Status: Not started
Acceptance Criteria:
1. Remains integration-tagged.
2. Uses test_orch_integration DB-core fixture hooks.

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

## Phase 7 - Validation and Closure

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
