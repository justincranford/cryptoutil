# Tasks - Framework v22 TestMain Orchestration Consolidation

Status: 4 of 77 tasks complete (5.2%)
Created: 2026-05-09
Last Updated: 2026-05-09

## Task Status Legend

1. Not started
2. In progress
3. Blocked
4. Complete

## Phase 1 - Research Freeze and Baseline Evidence

### Task 1.1 - Inventory lock

Status: Complete
Owner: LLM Agent
Description: Freeze all internal/apps TestMain paths and category counts.
Acceptance Criteria:
1. 28 TestMain files listed.
2. Count breakdown by category and PS-ID recorded in plan.md.
3. Evidence artifact created under test-output/completion-verification/.
Files:
4. docs/framework-v22/plan.md
5. test-output/completion-verification/framework-v22-testmain-signals.csv

### Task 1.2 - Pattern signal extraction

Status: Complete
Owner: LLM Agent
Description: Extract behavior signals from each TestMain and persist evidence.
Acceptance Criteria:
1. CSV evidence exists with file + signal columns.
2. Deep analysis appendix references findings.
Files:
3. test-output/completion-verification/framework-v22-testmain-signals.csv
4. docs/framework-v22/plan.md

### Task 1.3 - Distinct type taxonomy

Status: Complete
Owner: LLM Agent
Description: Define canonical TestMain types and duplication findings.
Acceptance Criteria:
1. Appendix includes distinct types and rationale.
2. Goal 2 list includes all TestMain files and purpose summaries.
Files:
3. docs/framework-v22/plan.md

### Task 1.4 - Deep completeness audit of planning artifacts

Status: Complete
Owner: LLM Agent
Description: Verify Goal 1/2/3 lists are complete and no PS-ID/TestMain was omitted.
Acceptance Criteria:
1. Goal 2 section includes all 28 TestMain files from internal/apps/**.
2. Goal 3 section maps all 28 TestMain files to refactor outcomes.
3. Audit evidence files created in test-output/completion-verification/.
Files:
4. docs/framework-v22/plan.md
5. test-output/completion-verification/framework-v22-plan-completeness-check.txt
6. test-output/completion-verification/framework-v22-goal2-goal3-coverage.txt

## Phase 2 - Orchestration API Design

### Task 2.1 - Design test_e2e package API

Status: Not started
Description: Define API contract for compose lifecycle + health wait + client setup.
Acceptance Criteria:
1. API draft reviewed.
2. Supports PS-ID-specific config injection.

### Task 2.2 - Design test_integration package API

Status: Not started
Description: Define integration startup contract for app+SQLite lifecycle.
Acceptance Criteria:
1. Supports constructor injection and dual-port wait.
2. Supports client fixture return values.

### Task 2.3 - Design test_db package API

Status: Not started
Description: Define DB fixture contract for SQLite/GORM/migrations.
Acceptance Criteria:
1. Supports migration callback injection.
2. Supports teardown and closed-DB error fixture path.

### Task 2.4 - Design test_api package API

Status: Not started
Description: Define API test orchestration for handler graph and request helpers.
Acceptance Criteria:
1. Supports in-memory app.Test path.
2. Supports integration HTTP client path when needed.

### Task 2.5 - Design test_cli package API

Status: Not started
Description: Define CLI harness orchestration.
Acceptance Criteria:
1. Argument and IO capture helpers defined.
2. Exit-code and output assertion helpers defined.

### Task 2.6 - Design additional candidate APIs

Status: Not started
Description: Define test_tls, test_barrier, test_compose, test_health, test_bootstrap.
Acceptance Criteria:
1. Candidate package list finalized.
2. Scope boundaries documented.

## Phase 3 - Implement Framework Orchestration Packages

### Task 3.1 - Create test_orchestration directory tree

Status: Not started
Description: Add package skeletons for all approved test_* modules.
Acceptance Criteria:
1. All planned package directories exist.
2. Each package has package docs and initial tests.

### Task 3.2 - Implement test_bootstrap primitives

Status: Not started
Description: Lifecycle hooks, teardown ordering, timeout constants.

### Task 3.3 - Implement test_db core

Status: Not started
Description: SQLite setup, PRAGMA, GORM wrapper, migration pipeline.

### Task 3.4 - Implement test_barrier fixtures

Status: Not started
Description: Telemetry + JWK + barrier reusable fixture graph.

### Task 3.5 - Implement test_tls fixtures

Status: Not started
Description: TLS pool and client builders for public/admin/mTLS test paths.

### Task 3.6 - Implement test_health helpers

Status: Not started
Description: Unified readiness and endpoint assertion helpers.

### Task 3.7 - Implement test_integration orchestration

Status: Not started
Description: Server startup, dual-port wait, ready gate, cleanup orchestration.

### Task 3.8 - Implement test_compose orchestration

Status: Not started
Description: Compose manager abstraction and wait orchestration.

### Task 3.9 - Implement test_e2e orchestration

Status: Not started
Description: PS-ID e2e TestMain factory facade built on test_compose and test_health.

### Task 3.10 - Implement test_api orchestration

Status: Not started
Description: Handler graph fixture and request harness.

### Task 3.11 - Implement test_cli orchestration

Status: Not started
Description: CLI invocation harness with deterministic assertions.

### Task 3.12 - Add package-level tests and export_test seams

Status: Not started
Description: Ensure all orchestration packages have high coverage and robust tests.

## Phase 4 - Migrate Existing Framework Callers

### Task 4.1 - Add compatibility wrappers in existing framework testing packages

Status: Not started

### Task 4.2 - Repoint existing e2e_infra call sites to test_e2e facade

Status: Not started

### Task 4.3 - Repoint server helper call sites to test_integration facade

Status: Not started

### Task 4.4 - Validate no behavior regressions in framework package tests

Status: Not started

## Phase 5 - PS-ID Migration (All 10 PS-IDs)

### Identity Family

### Task 5.1 - identity-authz server TestMain migration

Status: Not started

### Task 5.2 - identity-authz e2e TestMain migration

Status: Not started

### Task 5.3 - identity-idp server TestMain migration

Status: Not started

### Task 5.4 - identity-idp e2e TestMain migration

Status: Not started

### Task 5.5 - identity-rp server TestMain migration

Status: Not started

### Task 5.6 - identity-rp e2e TestMain migration

Status: Not started

### Task 5.7 - identity-rs server TestMain migration

Status: Not started

### Task 5.8 - identity-rs e2e TestMain migration

Status: Not started

### Task 5.9 - identity-spa server TestMain migration

Status: Not started

### Task 5.10 - identity-spa e2e TestMain migration

Status: Not started

### JOSE

### Task 5.11 - jose-ja e2e facade migration

Status: Not started

### Task 5.12 - jose-ja server TestMain migration

Status: Not started

### Task 5.13 - jose-ja repository TestMain migration to test_db

Status: Not started

### Task 5.14 - jose-ja service TestMain migration to test_db + test_barrier

Status: Not started

### PKI

### Task 5.15 - pki-ca e2e custom compose migration to test_e2e

Status: Not started

### Task 5.16 - pki-ca server migration to test_integration + test_tls

Status: Not started

### Skeleton

### Task 5.17 - skeleton-template e2e facade migration

Status: Not started

### Task 5.18 - skeleton-template server facade migration

Status: Not started

### SM-IM

### Task 5.19 - sm-im server migration from local SetupTestServer to test_integration

Status: Not started

### Task 5.20 - sm-im e2e facade migration

Status: Not started

### Task 5.21 - sm-im client TestMain migration to framework orchestration

Status: Not started

### Task 5.22 - sm-im repository TestMain migration to test_db + test_barrier

Status: Not started

### Task 5.23 - sm-im API TestMain migration to test_api + test_barrier

Status: Not started

### SM-KMS

### Task 5.24 - sm-kms server TestMain migration to test_integration

Status: Not started

### Task 5.25 - sm-kms e2e facade migration

Status: Not started

### Task 5.26 - sm-kms client TestMain migration to framework orchestration

Status: Not started

### Task 5.27 - sm-kms businesslogic sad-path orchestration migration to test_bootstrap

Status: Not started

### Task 5.28 - sm-kms orm transaction TestMain migration to test_db/test_api fixture

Status: Not started

## Phase 6 - Duplicate Fixture Removal and Cleanup

### Task 6.1 - Remove duplicated SQLite fixture code from jose-ja packages

Status: Not started

### Task 6.2 - Remove duplicated SQLite+barrier fixture code from sm-im packages

Status: Not started

### Task 6.3 - Remove duplicated client-package happy-path server startup logic

Status: Not started

### Task 6.4 - Remove obsolete helper paths after migration

Status: Not started

## Phase 7 - Template and PS-ID Instantiation Alignment

### Task 7.1 - Update __PS_ID__ template TestMain wrappers

Status: Not started

### Task 7.2 - Add integration_testing/e2e_testing wrapper strategy in template policy

Status: Not started

### Task 7.3 - Sync all 10 PS-ID instantiated files

Status: Not started

### Task 7.4 - Validate apps-ps-id-template exact-match compliance

Status: Not started

## Phase 8 - Linter Enforcement

### Task 8.1 - Implement lint-fitness testmain-orchestration-policy

Status: Not started

### Task 8.2 - Implement lint-fitness testmain-location-policy

Status: Not started

### Task 8.3 - Implement lint-fitness testmain-sadpath-policy

Status: Not started

### Task 8.4 - Implement lint-fitness template-testmain-compliance check

Status: Not started

### Task 8.5 - Add linter tests for pass/fail scenarios

Status: Not started

### Task 8.6 - Integrate checks into standard cicd-lint flows

Status: Not started

## Phase 9 - Validation and Rollout

### Task 9.1 - Build validation across all tags

Status: Not started

### Task 9.2 - Lint validation across all tags

Status: Not started

### Task 9.3 - Unit and integration test validation

Status: Not started

### Task 9.4 - E2E suite validation per PS-ID

Status: Not started

### Task 9.5 - Coverage and mutation target validation

Status: Not started

### Task 9.6 - Final policy conformance report

Status: Not started

## Cross-Cutting Quality Tasks

### Task Q1 - Ensure no happy-path startup outside framework orchestration

Status: Not started

### Task Q2 - Ensure sad-path startup intent is explicit and lint-validated

Status: Not started

### Task Q3 - Ensure all wrappers are thin and deterministic

Status: Not started

### Task Q4 - Ensure all new framework orchestration packages meet >=98% coverage

Status: Not started

### Task Q5 - Ensure no new TODO/FIXME introduced untracked

Status: Not started

## Verification Checklist for Planning Completeness

1. Goal 1 list complete with required modules and identified additional candidates.
2. Goal 2 list includes all 28 internal/apps TestMain files.
3. Goal 3 list maps all 28 files to migration outcomes.
4. No PS-ID missing from any goal-level list.
5. Enforcement path includes template updates and cicd lint-fitness checks.

## Evidence Archive

1. test-output/completion-verification/framework-v22-testmain-signals.csv
2. test-output/completion-verification/framework-v22-testmain-signals.txt
3. docs/framework-v22/plan.md
