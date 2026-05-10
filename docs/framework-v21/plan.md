# Implementation Plan - Framework v21 TestMain Directory Consolidation

Status: In Progress
Created: 2026-05-09
Last Updated: 2026-05-09
Purpose: Consolidate all TestMain orchestration to reusable apps-framework directories, remove duplicated startup logic, and enforce policy with templates and fitness linting.

## Overview

This plan migrates TestMain orchestration for all 10 PS-IDs and all in-scope framework TestMain packages to framework-owned reusable directories under internal/apps-framework/service/. The target is one canonical directory taxonomy with clear separation between lifecycle orchestration and reusable helpers.

Planned directory creation under internal/apps-framework/service/:

Orchestration directories:

1. test_orch_e2e/ - owns docker compose lifecycle, health-wait sequencing, and multi-instance end-to-end startup/shutdown.
2. test_orch_integration/ - owns direct PS-ID server startup, DB wiring, and TB-based cleanup.

Helper directories:

1. test_help_bootstrap/ - provides config/env/bootstrap wiring and startup scaffolding.
2. test_help_barrier/ - provides barrier/unseal fixture composition for DB and API suites.
3. test_help_db/ - provides SQLite/PostgreSQL fixture creation and DB failure-path helpers.
4. test_help_api/ - provides health clients, HTTP assertions, reusable HTTP mocks, and request-level checks.
5. test_help_cli/ - provides deterministic CLI argv/stdout/stderr/exit assertions.
6. test_help_tls/ - provides TLS material, certificate/client construction, and secure/insecure client helpers.

**NOTE: Current implementation temporarily consolidated compose helpers into test_orch_e2e.**
Target architecture remains a dedicated `test_help_compose` helper boundary as defined in the canonical taxonomy and Goal 1A ownership rules.

Execution profiles:

1. Integration profile: start one PS-ID server directly with SQLite and dynamic ports.
2. E2E profile: start one PS-ID docker compose stack with four app instances (2 SQLite + 2 PostgreSQL) plus dependent services.

Canonical directory taxonomy (must all be represented in API design, migration mapping, and policy enforcement):

Orchestration directories (lifecycle ownership — start/wait/shutdown):

1. test_orch_e2e
2. test_orch_integration

Helper directories (consumed by orchestration and suites; no lifecycle ownership):

1. test_help_compose
2. test_help_bootstrap
3. test_help_barrier
4. test_help_db
5. test_help_api
6. test_help_cli
7. test_help_tls

Helper consumer map:

1. test_help_compose: consumed by test_orch_e2e (docker compose orchestration, stack startup, health-wait coupling).
2. test_help_bootstrap: consumed by test_orch_e2e and test_orch_integration (config/env setup, startup wiring).
3. test_help_barrier: consumed by test_orch_integration and fixture-heavy repository/API suites that need unseal/barrier composition helpers.
4. test_help_db: consumed by test_orch_integration and repository-focused TestMain suites (for SQLite fixtures and DB fault-path setup).
5. test_help_api: consumed by test_orch_integration and test_orch_e2e (HTTP clients, health checks, assertions, reusable HTTP mocks).
6. test_help_cli: consumed by client and CLI entry-point test suites that need deterministic command execution/assertion wrappers.
7. test_help_tls: consumed by test_orch_integration and test_orch_e2e for TLS client/certificate helpers and secure/insecure test-client wiring.

Primary outcomes:

1. Build a complete 9-directory test_orch/test_help family.
2. Migrate existing framework testing packages to those modules (or classify as cross-cutting reusable utilities).
3. Migrate every TestMain in internal/apps and internal/apps-framework to the canonical orchestration APIs.
4. Enforce with templates and lint-fitness policies.

## Deep Research Summary (Last 24 Hours)

Evidence reviewed:

1. Same-day framework planning commits touching v21/v22 plan/task/lesson files.
2. Current v21 plan and tasks inventory for directory references and migration coverage.
3. Current in-repo TestMain inventory and directory references already captured in Goal 2 and tasks.

Findings requiring refactor in this plan revision:

1. The overview collapsed scope into only two directories (integration/e2e), which regressed intent from the expanded 9-directory model.
2. Later sections already referenced expanded helper directories, but the plan lacked an explicit top-level taxonomy and dependency model tying all sections together.
3. Directory ownership for HTTP API, CLI, DB, TLS, barrier, compose, and bootstrap needed to be made explicit in canonical mapping and phase deliverables to avoid future drift.

Correction applied in this revision:

1. Promote a two-family directory taxonomy (orchestration directories plus helper directories) to top-level plan scope.
2. Keep integration/e2e as execution profiles, not taxonomy directories.
3. Thread ownership boundaries through goals, migration mapping, phase plan, and completion criteria.

## Scope

In scope:

1. internal/apps-framework/service/test_orch_*/
2. internal/apps-framework/service/test_help_*/
3. internal/apps-framework/service/testing/*and internal/apps-framework/service/testutil/*
4. internal/apps/{10 PS-ID}/** files containing func TestMain
5. internal/apps-framework/** files containing func TestMain
6. api/cryptosuite-registry/templates/internal/apps/**PS_ID**/
7. internal/apps-tools/cicd_lint/lint_fitness/

Out of scope:

1. Business feature behavior unrelated to test orchestration.
2. OpenAPI runtime behavior changes.
3. Production startup architecture changes.

## Technical Context

Language: Go 1.26.1

Current framework testing package inventory:

1. assertions
2. contract (empty)
3. e2e_helpers
4. e2e_infra
5. fixtures
6. healthclient
7. httpservertests
8. stubs
9. testcli
10. testdb
11. testserver
12. service/testutil (outside service/testing, currently used by sm-im/server/testmain_test.go)

Observed canonical baselines from code:

1. Canonical E2E orchestration baseline is e2e_infra.SetupE2ETestMain (already used by jose-ja, sm-im, sm-kms, skeleton-template) and should be wrapped by test_orch_e2e.
2. Canonical integration helper baseline should be testserver.StartAndWait semantics (TB-based cleanup, not panic-based) plus test_help_tls bundle.
3. e2e_helpers.MustStartAndWaitForDualPorts is widely used but panic-based and should be migrated behind test_orch_integration wrappers.

## Goal 1 - Build 9 Directory Families

Build the following directories under internal/apps-framework/service/:

Orchestration directories:
1. test_orch_e2e
2. test_orch_integration

Helper directories:
3. test_help_compose
4. test_help_bootstrap
5. test_help_barrier
6. test_help_db
7. test_help_api
8. test_help_cli
9. test_help_tls

### Goal 1A - Directory Dependency and Ownership Model

Required dependency direction:

1. test_help_compose and test_help_bootstrap provide foundational compose/config wiring.
2. test_orch_e2e composes test_help_compose + test_help_bootstrap + test_help_api + test_help_tls (+ test_help_db only when local fixture helpers are needed).
3. test_orch_integration composes test_help_bootstrap + test_help_db + test_help_tls + test_help_api.
4. test_help_api and test_help_cli remain transport-level helpers, not lifecycle owners.
5. test_help_barrier remains crypto/barrier fixture support consumed by integration/e2e and fixture-heavy repository suites.

Boundary rules:

1. Lifecycle orchestration (start/wait/shutdown) belongs ONLY to test_orch_e2e and test_orch_integration.
2. Docker Compose stack startup/teardown/health-wait helpers belong to test_help_compose.
3. Config/env/startup wiring helpers belong to test_help_bootstrap.
4. Database fixture creation and DB failure-path helpers belong to test_help_db.
5. HTTP assertions/mocks/health checks belong to test_help_api.
6. CLI entry-point execution/assertions belong to test_help_cli.
7. TLS material/client construction helpers belong to test_help_tls.
8. Barrier/unseal fixture helpers belong to test_help_barrier.

### Goal 1B - Package Consolidation Matrix

Existing packages and target disposition:

1. testing/e2e_infra -> consolidate into test_orch_e2e + test_help_compose.
2. testing/e2e_helpers -> split by concern:
   - server_start_helpers -> test_orch_integration
   - http_helpers -> test_help_api + test_help_tls
   - config_helpers/db_helpers/user_auth_helpers -> test_help_bootstrap depending on concern
3. testing/testdb -> consolidate into test_help_db (keep temporary compatibility wrapper).
4. testing/testserver -> consolidate into test_orch_integration + test_help_tls.
5. testing/healthclient -> move into test_help_api (as requested).
6. testing/testcli -> move into test_help_cli (as requested).
7. service/testutil (HTTP mock servers) -> move generic HTTP mocks to test_help_api/mocks.
8. testing/assertions -> keep reusable beyond orchestration; import from test_help_api and test_orch_e2e.
9. testing/httpservertests -> keep reusable beyond orchestration; import from test_orch_integration.
10. testing/fixtures -> keep reusable beyond orchestration for seeded DB entities.
11. testing/stubs -> keep reusable beyond orchestration for server/application seams.
12. testing/contract -> empty package; remove or repurpose only with concrete contract tests.

### Goal 1 Completion Criteria

1. All modules exist with package docs and tests.
2. Consolidation matrix implemented and reflected in tasks.
3. Cross-cutting reusable packages explicitly retained and documented.
4. Infrastructure coverage target >=98%.
5. Directory ownership and dependency direction are validated in package docs and linter policy.

## Goal 2 - Deep Analysis Inventory and Classification

### Goal 2A - Full TestMain Scope

Total TestMain functions in scope: 39

1. internal/apps: 28
2. internal/apps-framework: 11

### Goal 2B - internal/apps (28)

1. identity-authz/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
2. identity-authz/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
3. identity-idp/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
4. identity-idp/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
5. identity-rp/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
6. identity-rp/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
7. identity-rs/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
8. identity-rs/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
9. identity-spa/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
10. identity-spa/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
11. jose-ja/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
12. jose-ja/server/repository/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_barrier, test_help_api.
13. jose-ja/server/service/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
14. jose-ja/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
15. pki-ca/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
16. pki-ca/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_barrier, test_help_api, test_help_tls.
17. skeleton-template/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
18. skeleton-template/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
19. sm-im/client/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_cli, test_help_api, test_help_tls.
20. sm-im/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
21. sm-im/server/apis/messages_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls, test_help_barrier.
22. sm-im/server/repository/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_barrier.
23. sm-im/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls, test_help_barrier.
24. sm-kms/client/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_cli, test_help_api, test_help_tls.
25. sm-kms/e2e/testmain_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
26. sm-kms/server/businesslogic/businesslogic_crud_test.go -> test_orch_integration, test_help_bootstrap, test_help_db.
27. sm-kms/server/repository/orm/orm_transaction_test.go -> test_orch_integration, test_help_bootstrap, test_help_db.
28. sm-kms/server/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.

### Goal 2C - internal/apps-framework (11)

1. service/config/tls_generator/tls_generator_test.go -> test_help_tls.
2. service/server/test_main_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
3. service/server/listener/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_tls.
4. service/server/apis/test_main_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_tls.
5. service/server/apis/registration_integration_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api, test_help_barrier.
6. service/server/businesslogic/tenant_registration_service_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_api.
7. service/server/repository/test_main_test.go -> test_orch_integration, test_help_bootstrap, test_help_db.
8. service/server/repository/orm/testmain_test.go -> test_orch_integration, test_help_bootstrap, test_help_db.
9. service/server/barrier/barrier_service_test.go -> test_orch_integration, test_help_bootstrap, test_help_db, test_help_barrier.
10. service/test_orch_e2e/otel_tls_e2e_test.go -> test_orch_e2e, test_help_compose, test_help_bootstrap, test_help_api, test_help_tls.
11. service/testing/testserver/testserver.go (commented example TestMain; documentation-only and excluded from migration) -> docs-only example of test_orch_integration + test_help_tls wiring.

### Goal 2D - Corrected Classification Decisions

1. sm-kms/server/repository/orm/orm_transaction_test.go is integration-tagged (`//go:build integration`) and should be classified as test_orch_integration + test_help_db, not sad-path/setup-only.
2. sm-kms/server/businesslogic/businesslogic_crud_test.go currently uses TestMain only for env var setup while each test runs full setupTestStack. This requires refactor to a shared test_orch_integration fixture with test_help_bootstrap + test_help_db.
3. sm-im/server/testmain_test.go mock infrastructure uses service/testutil mock servers. These mocks are useful and should be preserved as test_help_api/mocks, not discarded.
4. pki-ca/e2e/testmain_e2e_test.go has compose-up without health-wait coupling and should migrate to test_orch_e2e to remove readiness race risk.
5. sm-kms/client/testmain_test.go currently uses e2e_helpers server startup helper; it should use test_orch_integration instead.
6. api/cryptosuite-registry/templates/internal/apps/**PS_ID**/ is the registration point for the shared TestMain superset; every PS-ID template must expose the same wrapper superset and use the same helper directories in the same pattern.

### Goal 2 Completion Criteria

1. All 39 in-scope TestMains are classified and mapped.
2. Directory assignments reflect build tags and real setup behavior.
3. Classification maps each TestMain to one or more of the canonical directory families.
4. No PS-ID and no framework TestMain omitted.
5. Every PS-ID template under api/cryptosuite-registry/templates/internal/apps/**PS_ID**/ exposes the same TestMain superset and calls the same apps-framework helpers in the same shape.

## Goal 3 - Migration Mapping

### Goal 3A - Canonical Implementations by Directory

Canonical implementation boundaries are defined by the "Canonical directory taxonomy" and "Goal 1A - Directory Dependency and Ownership Model" sections; Goal 3 applies those same boundaries across all 39 in-scope TestMain migrations with no PS-ID ad-hoc lifecycle loops outside orchestration wrappers.

### Goal 3B - internal/apps Migration Summary

1. Identity server TestMain files (authz/idp/rp/rs/spa) migrate from local wait loops to test_orch_integration wrappers.
2. Identity e2e pass-through TestMain files migrate to test_orch_e2e wrappers.
3. jose-ja, skeleton-template, sm-im, sm-kms, pki-ca e2e TestMain files converge on test_orch_e2e facade.
4. sm-im/server and sm-im/client stop using local sm-im/testing helper startup; migrate to test_orch_integration.
5. sm-kms/client moves from e2e_helpers helper to test_orch_integration.
6. sm-kms/businesslogic package moves from per-test setupTestStack to shared TestMain fixture via test_orch_integration (with test_help_bootstrap + test_help_db composite).
7. sm-kms/repository/orm integration-tagged TestMain migrates to test_orch_integration_db fixture path (test_orch_integration with DB-core fixture hooks).
8. jose-ja/sm-im repository and API fixture TestMain files migrate to test_help_db + test_help_api + test_help_barrier compositions.

### Goal 3C - internal/apps-framework Migration Summary

1. service/server/*TestMain packages migrate to test_orch_integration or test_orch_e2e, composing appropriate test_help_* modules.
2. service/test_orch_e2e/otel_tls_e2e_test.go is migrated to test_orch_e2e facade and PS-ID parameterization.
3. Existing framework test packages keep compatibility wrappers during transition, then old entry points are removed.

### Goal 3 Completion Criteria

1. All 39 TestMains migrated or explicitly deprecated with replacement.
2. Client-package full-server startup duplication removed.
3. Integration-tagged tests use test_orch_integration orchestration and shared TestMain fixtures.
4. pki-ca e2e readiness risk eliminated by test_orch_e2e migration.
5. All helper directories (test_help_compose, test_help_bootstrap, test_help_barrier, test_help_db, test_help_api, test_help_cli, test_help_tls) are consumed by at least one migrated orchestrator or suite.
6. Every PS-ID template under api/cryptosuite-registry/templates/internal/apps/**PS_ID**/ exposes the same TestMain superset and calls the same apps-framework helpers in the same shape.

## Enforcement Strategy

Template enforcement:

1. Update **PS_ID** templates so server and e2e TestMain wrappers call test_orch_integration/test_orch_e2e.
2. Keep wrappers thin with PS-ID config injection only.

lint-fitness enforcement:

1. testmain-orchestration-policy
2. testmain-location-policy
3. testmain-integration-tag-policy
4. testmain-sadpath-policy
5. template-testmain-compliance

## Multi-Phase Execution Plan

Phase 1: Research and alignment corrections

1. Correct v21/v22 doc placement and metadata.
2. Freeze 39-file TestMain inventory and classification.
3. Freeze package consolidation matrix.
4. Freeze canonical directory taxonomy and ownership boundaries.

**Status**: ✅ COMPLETE (8/8 tasks)

Phase 2: API design (MANDATORY completion in planning phase)

**Critical boundary rule**: No design deferral to execution. ALL design decisions for Phase 2 MUST be fully specified in planning artifacts before implementation continues.

1. Task 2.1: Finalize test_orch_e2e API boundaries - ✅ COMPLETE (implemented in Phase 3.2)
2. Task 2.2: Finalize test_orch_integration API boundaries - 🔄 IN PROGRESS (critical path, blocks Phases 4-8 until design is complete)
3. Finalize moved-vs-reusable package boundaries.
4. Finalize directory dependency direction and no-overlap boundaries.

Phase 2 acceptance gates (must all pass before implementation resumes):
1. test_orch_integration API surface fully specified for all 37 remaining migrations.
2. Fixture scope model defined (per-test vs per-suite vs shared) with deterministic cleanup semantics.
3. Error-path fixture creation mechanism documented for DB/API failure tests.
4. Health endpoint and readiness parameterization defined for all PS-IDs.
5. Port-0 concurrent isolation and startup/shutdown contracts specified.
6. Package boundary and dependency direction rules finalized with no overlap ambiguity.
7. Execution phase receives complete design artifacts with no TBD/TBD-later decisions.

Phase 2 unblock artifact:
1. Round 1 answer merged: Migration compatibility strategy is one-pass direct migration with no compatibility wrappers.
2. `docs/framework-v21/quizme-v2.md` captures unresolved user-choice decisions required to close Task 2.2 and downstream helper API design tasks (2.3-2.6).
3. After answers are provided, merge decisions into plan/tasks and delete quizme-v2.md per lifecycle policy.

**Status**: ⚠️ PARTIAL (1/4 complete, 1 in progress on critical path)

Phase 3: Implement orchestration modules

1. Task 3.1: Create all test_orch_*and test_help_* directories - ⚠️ PARTIAL (only test_orch_e2e created)
2. Task 3.2: Implement test_orch_e2e from e2e_infra - ✅ COMPLETE (Commit 7d07de9c5)
3. Task 3.3: Implement test_orch_integration - ⏳ BLOCKED (Phase 2.2 design dependency)
4. Port code from existing framework testing packages.
5. Validate directory-level package docs and tests.

**Status**: ⚠️ IN PROGRESS (1/5 complete, 1 blocked by Phase 2.2)

**Key Achievement**: Parameterized E2E orchestration for all 10 PS-IDs via NewTLSPSIDSpec factory in test_orch_e2e/tls_psid_spec_e2e.go. All 3 TLS E2E test files are physically relocated under internal/apps-framework/service/test_orch_e2e and TestMain now supports PS-ID selection via CRYPTOUTIL_TLS_E2E_PSID.

Phase 4: Framework package migration

1. Move testcli -> test_help_cli.
2. Move healthclient -> test_help_api.
3. Move service/testutil mocks -> test_help_api/mocks.
4. Enforce one-pass direct migration with no compatibility wrappers.

**Status**: ⏳ BLOCKED (depends on Phase 2.2 completion)

Phase 5: internal/apps PS-ID migration

1. Migrate all 28 PS-ID TestMain files to wrappers/composites.
2. Enforce sm-im and sm-kms client migration to test_orch_integration.
3. Refactor sm-kms businesslogic and orm integration patterns.

**Status**: ⏳ BLOCKED (depends on Phase 2.2 completion, 37 TestMain files pending migration)

Phase 6: internal/apps-framework TestMain migration

1. Migrate all framework TestMain files to same reusable orchestration directories.

**Status**: ⏳ NOT STARTED (depends on Phase 4 completion)

Phase 7: Template and linter policy lock

1. Update templates.
2. Add/adjust lint-fitness rules.
3. Enforce canonical directory ownership in policy checks.

**Status**: ⏳ NOT STARTED (depends on Phase 5 completion)

Phase 8: Validation and rollout

1. Build, lint, test, coverage, mutation, and e2e validation.

**Status**: ⏳ NOT STARTED (depends on Phase 7 completion)

## Risks and Mitigations

1. Risk: helper API churn while moving packages.
   - Mitigation: one-pass migration playbook with explicit call-site cutover sequencing and package-level verification gates.
2. Risk: pki-ca e2e intermittent failures from compose readiness.
   - Mitigation: mandatory migration to test_orch_e2e health-wait orchestration.
3. Risk: integration-tagged suites regress during refactor.
   - Mitigation: dedicated testmain-integration-tag-policy and targeted integration test gates.
4. Risk: mock behavior regressions in sm-im/server tests.
   - Mitigation: preserve and migrate service/testutil semantics into test_help_api/mocks.
5. Risk: test_help_compose readiness coupling with test_orch_e2e orchestration.
   - Mitigation: mandate health-wait semantics at compose helper API layer (not optional).

## Quality Gates

Per phase gates:

1. go build ./...
2. go build -tags e2e,integration ./...
3. golangci-lint run ./...
4. golangci-lint run --build-tags e2e,integration ./...
5. go test ./... -shuffle=on
6. go run ./cmd/cicd-lint lint-fitness

Module coverage gate (plan-level):

1. Evidence that each canonical directory (test_orch_e2e, test_orch_integration, test_help_compose, test_help_bootstrap, test_help_barrier, test_help_db, test_help_api, test_help_cli, test_help_tls) is represented in API design artifacts, migration mappings, and linter/template policy.

Coverage/mutation targets:

1. Production >=95%
2. Infrastructure utilities >=98%
3. Mutation >=95% (infrastructure target >=98%)

## Evidence

Planning evidence retained under test-output/completion-verification/ and test-output/framework-docs-swap/.

## Deep Re-Research Closure Pass

Second-pass validation checklist for omissions:

1. Re-scan plan for directory-collapse language that implies only integration/e2e.
2. Re-verify all canonical directories appear in overview, goals, migration mapping, phases, and quality gates.
3. Re-verify HTTP API, CLI, DB, TLS, barrier, compose, and bootstrap helpers are explicitly called out as first-class directories.
4. Re-verify TestMain migration coverage remains 39 in-scope entries with no directory orphan.
5. Re-run grep-based keyword audit before execution begins (test_orch_e2e|test_orch_integration|test_help_compose|test_help_bootstrap|test_help_barrier|test_help_db|test_help_api|test_help_cli|test_help_tls).

## Quizme Round 1 (2026-05-09)

1. Question: Default fixture scope model for test_orch_integration?
   - Answer: E
2. Question: Error-path fixture creation contract?
   - Answer: E
3. Question: Readiness endpoint contract for integration orchestration?
   - Answer: E
4. Question: Port allocation and concurrency safety contract?
   - Answer: E
5. Question: Migration compatibility strategy?
   - Answer: C (remove wrappers immediately; require direct migration in one pass)
