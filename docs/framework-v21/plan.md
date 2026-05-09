# Implementation Plan - Framework v21 TestMain Orchestration Consolidation

Status: Planning
Created: 2026-05-09
Last Updated: 2026-05-09
Purpose: Consolidate all TestMain orchestration to reusable apps-framework test_orchestration packages, remove duplicated startup logic, and enforce policy with templates and fitness linting.

## Overview

This plan migrates TestMain orchestration for all 10 PS-IDs and all in-scope framework TestMain packages to framework-owned reusable modules under internal/apps-framework/service/test_orchestration/. The target is one canonical module taxonomy with clear separation between lifecycle orchestration and reusable helpers.

Planned directory creation under internal/apps-framework/service/test_orchestration/:

1. test_e2e/
2. test_integration/
3. test_helpers/
4. test_helpers/test_compose/
5. test_helpers/test_bootstrap/
6. test_helpers/test_db/
7. test_helpers/test_api/
8. test_helpers/test_cli/
9. test_helpers/test_tls/
10. test_helpers/test_barrier/

Execution profiles:

1. Integration profile: start one PS-ID server directly with SQLite and dynamic ports.
2. E2E profile: start one PS-ID docker compose stack with four app instances (2 SQLite + 2 PostgreSQL) plus dependent services.

Canonical module taxonomy (must all be represented in API design, migration mapping, and policy enforcement):

Orchestration modules (lifecycle ownership — start/wait/shutdown):

1. test_e2e
2. test_integration

Helper modules (consumed by orchestration and suites; no lifecycle ownership):

1. test_helpers/test_compose
2. test_helpers/test_bootstrap
3. test_helpers/test_db
4. test_helpers/test_api
5. test_helpers/test_cli
6. test_helpers/test_tls
7. test_helpers/test_barrier

Helper consumer map:

1. test_helpers/test_compose: consumed by test_e2e (docker compose orchestration, stack startup, health-wait coupling).
2. test_helpers/test_bootstrap: consumed by test_e2e and test_integration (config/env setup, startup wiring).
3. test_helpers/test_db: consumed by test_integration and repository-focused TestMain suites (for SQLite fixtures and DB fault-path setup).
4. test_helpers/test_api: consumed by test_integration and test_e2e (HTTP clients, health checks, assertions, reusable HTTP mocks).
5. test_helpers/test_cli: consumed by client and CLI entry-point test suites that need deterministic command execution/assertion wrappers.
6. test_helpers/test_tls: consumed by test_integration and test_e2e for TLS client/certificate helpers and secure/insecure test-client wiring.
7. test_helpers/test_barrier: consumed by integration/e2e plus barrier/repository fixture-heavy suites needing unseal/barrier composition helpers.

Primary outcomes:

1. Build a complete test_orchestration module family.
2. Migrate existing framework testing packages to those modules (or classify as cross-cutting reusable utilities).
3. Migrate every TestMain in internal/apps and internal/apps-framework to the canonical orchestration APIs.
4. Enforce with templates and lint-fitness policies.

## Deep Research Summary (Last 24 Hours)

Evidence reviewed:

1. Same-day framework planning commits touching v21/v22 plan/task/lesson files.
2. Current v21 plan and tasks inventory for category references and migration coverage.
3. Current in-repo TestMain inventory and category references already captured in Goal 2 and tasks.

Findings requiring refactor in this plan revision:

1. The overview collapsed scope into only two categories (integration/e2e), which regressed intent from the expanded category model.
2. Later sections already referenced expanded categories, but the plan lacked an explicit top-level taxonomy and dependency model tying all sections together.
3. Category ownership for HTTP API, CLI, DB, TLS, barrier, compose, and bootstrap needed to be made explicit in canonical mapping and phase deliverables to avoid future drift.

Correction applied in this revision:

1. Promote a two-family module taxonomy (orchestration modules plus helper modules) to top-level plan scope.
2. Keep integration/e2e as execution profiles, not taxonomy categories.
3. Thread ownership boundaries through goals, migration mapping, phase plan, and completion criteria.

## Scope

In scope:

1. internal/apps-framework/service/test_orchestration/test_*/
2. internal/apps-framework/service/testing/*and internal/apps-framework/service/testutil/*
3. internal/apps/{10 PS-ID}/** files containing func TestMain
4. internal/apps-framework/** files containing func TestMain
5. api/cryptosuite-registry/templates/internal/apps/__PS_ID__/
6. internal/apps-tools/cicd_lint/lint_fitness/

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

1. Canonical E2E orchestration baseline is e2e_infra.SetupE2ETestMain (already used by jose-ja, sm-im, sm-kms, skeleton-template).
2. Canonical integration helper baseline should be testserver.StartAndWait semantics (TB-based cleanup, not panic-based) plus test_helpers/test_tls bundle.
3. e2e_helpers.MustStartAndWaitForDualPorts is widely used but panic-based and should be migrated behind test_integration wrappers.

## Goal 1 - Build test_orchestration Module Families

Build the following modules under internal/apps-framework/service/test_orchestration/:

Orchestration modules:
1. test_e2e
2. test_integration

Helper modules (under test_helpers/):
3. test_helpers/test_compose
4. test_helpers/test_bootstrap
5. test_helpers/test_db
6. test_helpers/test_api
7. test_helpers/test_cli
8. test_helpers/test_tls
9. test_helpers/test_barrier

### Goal 1A - Category Dependency and Ownership Model

Required dependency direction:

1. test_helpers/test_compose and test_helpers/test_bootstrap provide foundational compose/config wiring.
2. test_e2e composes test_helpers/test_compose + test_helpers/test_bootstrap + test_helpers/test_api + test_helpers/test_tls (+ test_helpers/test_db only when local fixture helpers are needed).
3. test_integration composes test_helpers/test_bootstrap + test_helpers/test_db + test_helpers/test_tls + test_helpers/test_api.
4. test_helpers/test_api and test_helpers/test_cli remain transport-level helpers, not lifecycle owners.
5. test_helpers/test_barrier remains crypto/barrier fixture support consumed by integration/e2e and fixture-heavy repository suites.

Boundary rules:

1. Lifecycle orchestration (start/wait/shutdown) belongs ONLY to test_e2e and test_integration.
2. Docker Compose stack startup/teardown/health-wait helpers belong to test_helpers/test_compose.
3. Config/env/startup wiring helpers belong to test_helpers/test_bootstrap.
4. Database fixture creation and DB failure-path helpers belong to test_helpers/test_db.
5. HTTP assertions/mocks/health checks belong to test_helpers/test_api.
6. CLI entry-point execution/assertions belong to test_helpers/test_cli.
7. TLS material/client construction helpers belong to test_helpers/test_tls.
8. Barrier/unseal fixture helpers belong to test_helpers/test_barrier.

### Goal 1B - Package Consolidation Matrix

Existing packages and target disposition:

1. testing/e2e_infra -> consolidate into test_e2e + test_helpers/test_compose.
2. testing/e2e_helpers -> split by concern:
   - server_start_helpers -> test_integration
   - http_helpers -> test_helpers/test_api + test_helpers/test_tls
   - config_helpers/db_helpers/user_auth_helpers -> test_helpers/test_bootstrap depending on concern
3. testing/testdb -> consolidate into test_helpers/test_db (keep temporary compatibility wrapper).
4. testing/testserver -> consolidate into test_integration + test_helpers/test_tls.
5. testing/healthclient -> move into test_helpers/test_api (as requested).
6. testing/testcli -> move into test_helpers/test_cli (as requested).
7. service/testutil (HTTP mock servers) -> move generic HTTP mocks to test_helpers/test_api/mocks.
8. testing/assertions -> keep reusable beyond orchestration; import from test_helpers/test_api and test_e2e.
9. testing/httpservertests -> keep reusable beyond orchestration; import from test_integration.
10. testing/fixtures -> keep reusable beyond orchestration for seeded DB entities.
11. testing/stubs -> keep reusable beyond orchestration for server/application seams.
12. testing/contract -> empty package; remove or repurpose only with concrete contract tests.

### Goal 1 Completion Criteria

1. All modules exist with package docs and tests.
2. Consolidation matrix implemented and reflected in tasks.
3. Cross-cutting reusable packages explicitly retained and documented.
4. Infrastructure coverage target >=98%.
5. Category ownership and dependency direction are validated in package docs and linter policy.

## Goal 2 - Deep Analysis Inventory and Classification

### Goal 2A - Full TestMain Scope

Total TestMain functions in scope: 39

1. internal/apps: 28
2. internal/apps-framework: 11

### Goal 2B - internal/apps (28)

1. identity-authz/e2e/testmain_e2e_test.go
2. identity-authz/server/testmain_test.go
3. identity-idp/e2e/testmain_e2e_test.go
4. identity-idp/server/testmain_test.go
5. identity-rp/e2e/testmain_e2e_test.go
6. identity-rp/server/testmain_test.go
7. identity-rs/e2e/testmain_e2e_test.go
8. identity-rs/server/testmain_test.go
9. identity-spa/e2e/testmain_e2e_test.go
10. identity-spa/server/testmain_test.go
11. jose-ja/e2e/testmain_e2e_test.go
12. jose-ja/server/repository/testmain_test.go
13. jose-ja/server/service/testmain_test.go
14. jose-ja/server/testmain_test.go
15. pki-ca/e2e/testmain_e2e_test.go
16. pki-ca/server/testmain_test.go
17. skeleton-template/e2e/testmain_e2e_test.go
18. skeleton-template/server/testmain_test.go
19. sm-im/client/testmain_test.go
20. sm-im/e2e/testmain_e2e_test.go
21. sm-im/server/apis/messages_test.go
22. sm-im/server/repository/testmain_test.go
23. sm-im/server/testmain_test.go
24. sm-kms/client/testmain_test.go
25. sm-kms/e2e/testmain_e2e_test.go
26. sm-kms/server/businesslogic/businesslogic_crud_test.go
27. sm-kms/server/repository/orm/orm_transaction_test.go
28. sm-kms/server/testmain_test.go

### Goal 2C - internal/apps-framework (11)

1. service/config/tls_generator/tls_generator_test.go
2. service/server/test_main_test.go
3. service/server/listener/testmain_test.go
4. service/server/apis/test_main_test.go
5. service/server/apis/registration_integration_test.go
6. service/server/businesslogic/tenant_registration_service_test.go
7. service/server/repository/test_main_test.go
8. service/server/repository/orm/testmain_test.go
9. service/server/barrier/barrier_service_test.go
10. tls/e2e/otel_tls_e2e_test.go
11. service/testing/testserver/testserver.go (commented example TestMain; documentation-only and excluded from migration)

### Goal 2D - Corrected Classification Decisions

1. sm-kms/server/repository/orm/orm_transaction_test.go is integration-tagged (`//go:build integration`) and should be classified as integration DB-core fixture, not sad-path/setup-only.
2. sm-kms/server/businesslogic/businesslogic_crud_test.go currently uses TestMain only for env var setup while each test runs full setupTestStack. This requires refactor to shared TestMain integration fixture.
3. sm-im/server/testmain_test.go mock infrastructure uses service/testutil mock servers. These mocks are useful and should be preserved as test_helpers/test_api/mocks, not discarded.
4. pki-ca/e2e/testmain_e2e_test.go has compose-up without health-wait coupling and should migrate to test_e2e to remove readiness race risk.
5. sm-kms/client/testmain_test.go currently uses e2e_helpers server startup helper; it should use test_integration instead.

### Goal 2 Completion Criteria

1. All 39 in-scope TestMains are classified and mapped.
2. Category assignments reflect build tags and real setup behavior.
3. Classification maps each TestMain to one or more of the canonical module families.
4. No PS-ID and no framework TestMain omitted.

## Goal 3 - Migration Mapping

### Goal 3A - Canonical Implementations by Category

1. Integration canonical path: test_integration (consumes test_helpers/test_bootstrap + test_helpers/test_db + test_helpers/test_tls + test_helpers/test_api/assertions).
2. E2E canonical path: test_e2e (consumes test_helpers/test_compose + test_helpers/test_bootstrap + test_helpers/test_api health checks + secure/insecure clients + test_helpers/test_tls).
3. Compose helper canonical path: test_helpers/test_compose for docker compose stack operations.
4. Bootstrap helper canonical path: test_helpers/test_bootstrap for config/env setup.
5. Database helper canonical path: test_helpers/test_db for database fixtures and error-path setup.
6. API helper canonical path: test_helpers/test_api for health clients, HTTP assertions, and reusable HTTP mocks.
7. CLI helper canonical path: test_helpers/test_cli for deterministic command execution/assertions.
8. TLS helper canonical path: test_helpers/test_tls for TLS certificate/client construction.
9. Barrier helper canonical path: test_helpers/test_barrier for barrier/unseal fixture composition in DB and API suites.
10. No direct PS-ID ad-hoc startup loops outside wrappers.

### Goal 3B - internal/apps Migration Summary

1. Identity server TestMain files (authz/idp/rp/rs/spa) migrate from local wait loops to test_integration wrappers.
2. Identity e2e pass-through TestMain files migrate to test_e2e wrappers.
3. jose-ja, skeleton-template, sm-im, sm-kms, pki-ca e2e TestMain files converge on test_e2e facade.
4. sm-im/server and sm-im/client stop using local sm-im/testing helper startup; migrate to test_integration.
5. sm-kms/client moves from e2e_helpers helper to test_integration.
6. sm-kms/businesslogic package moves from per-test setupTestStack to shared TestMain fixture via test_integration (with test_helpers/test_bootstrap + test_helpers/test_db composite).
7. sm-kms/repository/orm integration-tagged TestMain migrates to test_integration_db fixture path (test_integration with DB-core fixture hooks).
8. jose-ja/sm-im repository and API fixture TestMain files migrate to test_helpers/test_db + test_helpers/test_api + test_helpers/test_barrier compositions.

### Goal 3C - internal/apps-framework Migration Summary

1. service/server/*TestMain packages migrate to test_integration or test_e2e, composing appropriate test_helpers/* modules.
2. tls/e2e/otel_tls_e2e_test.go migrates to test_e2e facade.
3. Existing framework test packages keep compatibility wrappers during transition, then old entry points are removed.

### Goal 3 Completion Criteria

1. All 39 TestMains migrated or explicitly deprecated with replacement.
2. Client-package full-server startup duplication removed.
3. Integration-tagged tests use integration orchestration and shared TestMain fixtures.
4. pki-ca e2e readiness risk eliminated by test_e2e migration.
5. All helper modules (test_helpers/test_compose, test_helpers/test_bootstrap, test_helpers/test_db, test_helpers/test_api, test_helpers/test_cli, test_helpers/test_tls, test_helpers/test_barrier) are consumed by at least one migrated orchestrator or suite.

## Enforcement Strategy

Template enforcement:

1. Update __PS_ID__ templates so server and e2e TestMain wrappers call test_integration/test_e2e.
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
4. Freeze canonical module taxonomy and ownership boundaries.

Phase 2: API design

1. Finalize module API boundaries for all orchestration/helper modules.
2. Finalize moved-vs-reusable package boundaries.
3. Finalize category dependency direction and no-overlap boundaries.

Phase 3: Implement orchestration modules

1. Create all test_orchestration packages.
2. Port code from existing framework testing packages.
3. Validate category-level package docs and tests.

Phase 4: Framework package migration

1. Move testcli -> test_helpers/test_cli.
2. Move healthclient -> test_helpers/test_api.
3. Move service/testutil mocks -> test_helpers/test_api/mocks.
4. Keep compatibility wrappers during migration.

Phase 5: internal/apps PS-ID migration

1. Migrate all 28 PS-ID TestMain files to wrappers/composites.
2. Enforce sm-im and sm-kms client migration to test_integration.
3. Refactor sm-kms businesslogic and orm integration patterns.

Phase 6: internal/apps-framework TestMain migration

1. Migrate all framework TestMain files to same reusable orchestration modules.

Phase 7: Template and linter policy lock

1. Update templates.
2. Add/adjust lint-fitness rules.
3. Enforce canonical category ownership in policy checks.

Phase 8: Validation and rollout

1. Build, lint, test, coverage, mutation, and e2e validation.

## Risks and Mitigations

1. Risk: helper API churn while moving packages.
   - Mitigation: temporary wrappers with strict deprecation tasks.
2. Risk: pki-ca e2e intermittent failures from compose readiness.
   - Mitigation: mandatory migration to test_e2e health-wait orchestration.
3. Risk: integration-tagged suites regress during refactor.
   - Mitigation: dedicated testmain-integration-tag-policy and targeted integration test gates.
4. Risk: mock behavior regressions in sm-im/server tests.
   - Mitigation: preserve and migrate service/testutil semantics into test_helpers/test_api/mocks.
5. Risk: test_helpers/test_compose readiness coupling with test_e2e orchestration.
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

1. Evidence that each canonical module (test_e2e, test_integration, test_helpers/test_compose, test_helpers/test_bootstrap, test_helpers/test_db, test_helpers/test_api, test_helpers/test_cli, test_helpers/test_tls, test_helpers/test_barrier) is represented in API design artifacts, migration mappings, and linter/template policy.

Coverage/mutation targets:

1. Production >=95%
2. Infrastructure utilities >=98%
3. Mutation >=95% (infrastructure target >=98%)

## Evidence

Planning evidence retained under test-output/completion-verification/ and test-output/framework-docs-swap/.

## Deep Re-Research Closure Pass

Second-pass validation checklist for omissions:

1. Re-scan plan for module-collapse language that implies only integration/e2e.
2. Re-verify all canonical modules appear in overview, goals, migration mapping, phases, and quality gates.
3. Re-verify HTTP API, CLI, DB, TLS, barrier, compose, and bootstrap helpers are explicitly called out as first-class modules.
4. Re-verify TestMain migration coverage remains 39 in-scope entries with no module orphan.
5. Re-run grep-based keyword audit before execution begins (test_e2e|test_integration|test_helpers/test_compose|test_helpers/test_bootstrap|test_helpers/test_db|test_helpers/test_api|test_helpers/test_cli|test_helpers/test_tls|test_helpers/test_barrier).
