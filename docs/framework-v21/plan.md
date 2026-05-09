# Implementation Plan - Framework v21 TestMain Orchestration Consolidation

Status: Planning
Created: 2026-05-09
Last Updated: 2026-05-09
Purpose: Consolidate all TestMain orchestration to reusable apps-framework test_orchestration packages, remove duplicated startup logic, and enforce policy with templates and fitness linting.

## Overview

This plan migrates TestMain orchestration for all 10 PS-IDs and all in-scope framework TestMain packages to framework-owned reusable orchestration modules under internal/apps-framework/service/test_orchestration/. The target is a single canonical approach for:

1. Integration tests: start one PS-ID server directly with SQLite and dynamic ports.
2. E2E tests: start one PS-ID docker compose stack with four app instances (2 SQLite + 2 PostgreSQL) plus dependent services.

Primary outcomes:

1. Build a complete test_orchestration module family.
2. Migrate existing framework testing packages to those modules (or classify as cross-cutting reusable utilities).
3. Migrate every TestMain in internal/apps and internal/apps-framework to the canonical orchestration APIs.
4. Enforce with templates and lint-fitness policies.

## Mutual Exclusivity Guardrails

This document is TestMain-only.

1. Allowed: TestMain inventory, test orchestration APIs, integration/e2e startup patterns, TestMain migration sequencing, and TestMain lint policies.
2. Not allowed: PS-ID template directory-shape convergence, server subdirectory migration plans, mojibake linter work, and template MANIFEST directory enforcement details.
3. Any template-first planning content belongs under docs/framework-v22/ only.

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
2. Canonical integration helper baseline should be testserver.StartAndWait semantics (TB-based cleanup, not panic-based) plus test_tls bundle.
3. e2e_helpers.MustStartAndWaitForDualPorts is widely used but panic-based and should be migrated behind test_integration wrappers.

## Goal 1 - Build test_orchestration Modules

Build the following modules under internal/apps-framework/service/test_orchestration/:

1. test_e2e
2. test_integration
3. test_db
4. test_api
5. test_cli
6. test_tls
7. test_barrier
8. test_compose
9. test_bootstrap

### Goal 1A - Package Consolidation Matrix

Existing packages and target disposition:

1. testing/e2e_infra -> consolidate into test_e2e + test_compose.
2. testing/e2e_helpers -> split by concern:
   - server_start_helpers -> test_integration
   - http_helpers -> test_api + test_tls
   - config_helpers/db_helpers/user_auth_helpers -> test_e2e or test_bootstrap depending on concern
3. testing/testdb -> consolidate into test_db (keep temporary compatibility wrapper).
4. testing/testserver -> consolidate into test_integration + test_tls.
5. testing/healthclient -> move into test_api (as requested).
6. testing/testcli -> move into test_cli (as requested).
7. service/testutil (HTTP mock servers) -> move generic HTTP mocks to test_api/mocks.
8. testing/assertions -> keep reusable beyond orchestration; import from test_api/test_e2e.
9. testing/httpservertests -> keep reusable beyond orchestration; import from test_integration.
10. testing/fixtures -> keep reusable beyond orchestration for seeded DB entities.
11. testing/stubs -> keep reusable beyond orchestration for server/application seams.
12. testing/contract -> empty package; remove or repurpose only with concrete contract tests.

### Goal 1 Completion Criteria

1. All modules exist with package docs and tests.
2. Consolidation matrix implemented and reflected in tasks.
3. Cross-cutting reusable packages explicitly retained and documented.
4. Infrastructure coverage target >=98%.

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
3. sm-im/server/testmain_test.go mock infrastructure uses service/testutil mock servers. These mocks are useful and should be preserved as test_api/mocks, not discarded.
4. pki-ca/e2e/testmain_e2e_test.go has compose-up without health-wait coupling and should migrate to test_e2e to remove readiness race risk.
5. sm-kms/client/testmain_test.go currently uses e2e_helpers server startup helper; it should use test_integration instead.

### Goal 2 Completion Criteria

1. All 39 in-scope TestMains are classified and mapped.
2. Category assignments reflect build tags and real setup behavior.
3. No PS-ID and no framework TestMain omitted.

## Goal 3 - Migration Mapping

### Goal 3A - Canonical Integration and E2E Implementations

1. Integration canonical path: test_integration + test_tls + test_api/assertions.
2. E2E canonical path: test_e2e (built on test_compose + health checks + secure/insecure clients).
3. No direct PS-ID ad-hoc startup loops outside wrappers.

### Goal 3B - internal/apps Migration Summary

1. Identity server TestMain files (authz/idp/rp/rs/spa) migrate from local wait loops to test_integration wrappers.
2. Identity e2e pass-through TestMain files migrate to test_e2e wrappers.
3. jose-ja, skeleton-template, sm-im, sm-kms, pki-ca e2e TestMain files converge on test_e2e facade.
4. sm-im/server and sm-im/client stop using local sm-im/testing helper startup; migrate to test_integration.
5. sm-kms/client moves from e2e_helpers helper to test_integration.
6. sm-kms/businesslogic package moves from per-test setupTestStack to shared TestMain fixture via test_integration/test_db composite helper.
7. sm-kms/repository/orm integration-tagged TestMain migrates to test_integration_db fixture path (test_integration with DB-core fixture hooks).
8. jose-ja/sm-im repository and API fixture TestMain files migrate to test_db/test_api/test_barrier compositions.

### Goal 3C - internal/apps-framework Migration Summary

1. service/server/* TestMain packages migrate to test_integration/test_db/test_barrier/test_api as appropriate.
2. tls/e2e/otel_tls_e2e_test.go migrates to test_e2e facade.
3. Existing framework test packages keep compatibility wrappers during transition, then old entry points are removed.

### Goal 3 Completion Criteria

1. All 39 TestMains migrated or explicitly deprecated with replacement.
2. Client-package full-server startup duplication removed.
3. Integration-tagged tests use integration orchestration and shared TestMain fixtures.
4. pki-ca e2e readiness risk eliminated by test_e2e migration.

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

Phase 2: API design

1. Finalize module API boundaries.
2. Finalize moved-vs-reusable package boundaries.

Phase 3: Implement orchestration modules

1. Create all test_orchestration packages.
2. Port code from existing framework testing packages.

Phase 4: Framework package migration

1. Move testcli -> test_cli.
2. Move healthclient -> test_api.
3. Move service/testutil mocks -> test_api/mocks.
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
   - Mitigation: preserve and migrate service/testutil semantics into test_api/mocks.

## Quality Gates

Per phase gates:

1. go build ./...
2. go build -tags e2e,integration ./...
3. golangci-lint run ./...
4. golangci-lint run --build-tags e2e,integration ./...
5. go test ./... -shuffle=on
6. go run ./cmd/cicd-lint lint-fitness

Coverage/mutation targets:

1. Production >=95%
2. Infrastructure utilities >=98%
3. Mutation >=95% (infrastructure target >=98%)

## Evidence

Planning evidence retained under test-output/completion-verification/ and test-output/framework-docs-swap/.
