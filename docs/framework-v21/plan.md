# Implementation Plan - Framework v22 TestMain Orchestration Consolidation

Status: Planning
Created: 2026-05-09
Last Updated: 2026-05-09
Purpose: Refactor all PS-ID TestMain orchestration to a single apps-framework test orchestration architecture, remove duplicated startup logic, and enforce allowed patterns through templates and cicd fitness linters.

## Overview

This plan defines a multi-phase migration of all TestMain orchestration in all 10 PS-IDs to reusable framework-level orchestration packages under internal/apps-framework/service/test_orchestration/. The design target is that happy-path startup orchestration is framework-owned and reused everywhere; PS-ID-local TestMain files become thin wrappers or are removed.

Primary outcomes:
1. Build a complete internal/apps-framework/service/test_orchestration/ module family that covers all reusable startup patterns.
2. Inventory and classify every internal/apps/** TestMain in all 10 PS-IDs.
3. Refactor every TestMain in all 10 PS-IDs to use framework orchestration only.
4. Enforce the policy through canonical templates and cicd-lint fitness checks.

## Scope

In scope:
1. internal/apps-framework/service/test_orchestration/test_*/
2. internal/apps/{10 PS-ID}/** test files containing TestMain
3. api/cryptosuite-registry/templates/internal/apps/__PS_ID__/
4. internal/apps-tools/cicd_lint/lint_fitness/ (new or updated linters)
5. test suites and migration docs for orchestration migration

Out of scope:
1. Business feature changes unrelated to test orchestration
2. OpenAPI behavior changes
3. Runtime production startup architecture changes

## Technical Context

Language: Go 1.26.1
Testing strategy baseline: SQLite in-memory for unit/integration, Docker Compose for e2e
Current orchestration baseline:
1. Reusable helpers exist in internal/apps-framework/service/testing/e2e_helpers and e2e_infra
2. Reusable DB helper exists in internal/apps-framework/service/testing/testdb
3. Reusable server helper exists in internal/apps-framework/service/testing/testserver
4. TestMain orchestration is still partially duplicated in PS-ID packages

## Goal 1 - Build Framework Test Orchestration Modules

The first implementation goal is to establish framework-owned reusable orchestration modules under internal/apps-framework/service/test_orchestration/test_*/.

### Goal 1 Numbered List - All test_orchestration modules

1. internal/apps-framework/service/test_orchestration/test_e2e/
   - Canonical E2E TestMain factory for compose lifecycle, health waits, TLS clients, logs, cleanup.
   - Wraps and evolves current e2e_infra setup patterns.

2. internal/apps-framework/service/test_orchestration/test_integration/
   - Canonical integration TestMain factory for app startup with SQLite (public+admin dynamic ports, ready gates, client setup, shutdown).
   - Supports per-PS-ID server constructor injection via function parameters.

3. internal/apps-framework/service/test_orchestration/test_db/
   - Canonical DB-only TestMain factory for SQLite+GORM+migration lifecycle.
   - Supports optional telemetry/JWK/barrier fixture composition.

4. internal/apps-framework/service/test_orchestration/test_api/
   - Canonical API handler/integration orchestration, including middleware wiring, authn/authz stubs, request helpers, error assertion utilities.
   - Supports app.Test-style in-memory HTTP and optional integration HTTP clients.

5. internal/apps-framework/service/test_orchestration/test_cli/
   - Canonical CLI TestMain and command harness orchestration for subcommands (server/client/health/livez/readyz/shutdown/init/compose/e2e).
   - Includes stdout/stderr capture, deterministic argument setup, and exit-code assertions.

6. internal/apps-framework/service/test_orchestration/test_tls/ (candidate identified)
   - Shared TLS fixtures: root pool, public/admin clients, mTLS client cert fixture loading, negative TLS handshake helpers.
   - Extracts repeated TLS client setup from multiple server TestMain files.

7. internal/apps-framework/service/test_orchestration/test_barrier/ (candidate identified)
   - Shared barrier/JWK/telemetry fixture graph for packages that need encrypted content tests without full server startup.
   - Extracts repeated jose-ja and sm-im barrier stack setup.

8. internal/apps-framework/service/test_orchestration/test_compose/ (candidate identified)
   - Compose manager wrappers and readiness orchestration for direct compose start/stop patterns.
   - Used by e2e orchestration and explicit compose-focused tests.

9. internal/apps-framework/service/test_orchestration/test_health/ (candidate identified)
   - Shared health endpoint wait/assert clients for public and admin paths.
   - Consolidates repeated wait loops and health assertions.

10. internal/apps-framework/service/test_orchestration/test_bootstrap/ (candidate identified)
    - Shared test bootstrapping primitives: lifecycle hooks, deterministic cleanup ordering, timeout policy constants, panic-safe teardown.

### Goal 1 Completion Criteria

1. All modules above exist with package-level documentation and tests.
2. Existing framework testing helpers are migrated or wrapped without regression.
3. New orchestration APIs use constructor/function injection, no package-global mutable seams.
4. Coverage targets: >=98% for infrastructure utility packages.

## Goal 2 - Deep Analysis of Every internal/apps/** TestMain

The second implementation goal is complete analysis and classification of each TestMain in every PS-ID.

### Inventory Summary

Total internal/apps TestMain files: 28

Breakdown by functional category:
1. PS-ID e2e package TestMain files: 10
2. PS-ID server package TestMain files: 10
3. PS-ID client package TestMain files: 2
4. PS-ID server sub-package DB/API fixture TestMain files: 4
5. Sad-path or env/setup-only TestMain files: 2

Breakdown by PS-ID:
1. identity-authz: 2
2. identity-idp: 2
3. identity-rp: 2
4. identity-rs: 2
5. identity-spa: 2
6. jose-ja: 4
7. pki-ca: 2
8. skeleton-template: 2
9. sm-im: 5
10. sm-kms: 5

### Goal 2 Numbered List - All internal/apps/** TestMain and what they do

1. internal/apps/identity-authz/e2e/testmain_e2e_test.go
   - Pass-through TestMain only: os.Exit(m.Run()).
   - No compose startup orchestration yet.

2. internal/apps/identity-authz/server/testmain_test.go
   - Starts server with NewTestConfig(loopback, dynamic port, test mode).
   - Runs manual readiness polling and graceful shutdown.

3. internal/apps/identity-idp/e2e/testmain_e2e_test.go
   - Pass-through TestMain only.

4. internal/apps/identity-idp/server/testmain_test.go
   - Starts server with NewTestConfig and manual readiness loop.

5. internal/apps/identity-rp/e2e/testmain_e2e_test.go
   - Pass-through TestMain only.

6. internal/apps/identity-rp/server/testmain_test.go
   - Starts server, creates public/admin TLS clients, waits for readiness, shutdown.

7. internal/apps/identity-rs/e2e/testmain_e2e_test.go
   - Pass-through TestMain only.

8. internal/apps/identity-rs/server/testmain_test.go
   - Starts server with manual readiness loop and shutdown.

9. internal/apps/identity-spa/e2e/testmain_e2e_test.go
   - Pass-through TestMain only.

10. internal/apps/identity-spa/server/testmain_test.go
    - Starts server with DefaultTestConfig and local waitForReady helper.

11. internal/apps/jose-ja/e2e/testmain_e2e_test.go
    - Uses framework SetupE2ETestMain factory, compose profile orchestration, secure/insecure clients.

12. internal/apps/jose-ja/server/repository/testmain_test.go
    - DB-only TestMain: SQLite memory DB, PRAGMA, GORM, migrations.

13. internal/apps/jose-ja/server/service/testmain_test.go
    - DB+services TestMain: SQLite, migrations, telemetry, JWKGen, barrier fixtures.

14. internal/apps/jose-ja/server/testmain_test.go
    - Server integration TestMain: NewFromConfig + MustStartAndWaitForDualPorts + health client setup.

15. internal/apps/pki-ca/e2e/testmain_e2e_test.go
    - Custom compose startup/stop orchestration (not SetupE2ETestMain factory).

16. internal/apps/pki-ca/server/testmain_test.go
    - Server integration TestMain with helper-based dual-port startup and TLS test client.

17. internal/apps/skeleton-template/e2e/testmain_e2e_test.go
    - Uses framework SetupE2ETestMain factory.

18. internal/apps/skeleton-template/server/testmain_test.go
    - Server integration TestMain with helper-based dual-port startup and public/admin clients.

19. internal/apps/sm-im/client/testmain_test.go
    - Client package TestMain starts full sm-im server via StartSmIMService.
    - Duplicates happy-path server startup in client package.

20. internal/apps/sm-im/e2e/testmain_e2e_test.go
    - Uses framework SetupE2ETestMain factory.

21. internal/apps/sm-im/server/apis/messages_test.go
    - DB+services TestMain: SQLite, migrations, telemetry, JWKGen, barrier, repositories, handler wiring.

22. internal/apps/sm-im/server/repository/testmain_test.go
    - DB+services TestMain: SQLite, migrations, telemetry, JWKGen, barrier.

23. internal/apps/sm-im/server/testmain_test.go
    - Server integration TestMain uses PS-ID helper SetupTestServer, then assigns package globals.

24. internal/apps/sm-kms/client/testmain_test.go
    - Client package TestMain starts full KMS server directly.
    - Duplicates happy-path server startup in client package.

25. internal/apps/sm-kms/e2e/testmain_e2e_test.go
    - Uses framework SetupE2ETestMain factory.

26. internal/apps/sm-kms/server/businesslogic/businesslogic_crud_test.go
    - Sad-path/environment setup TestMain with env var mutation and passthrough run.
    - Does not orchestrate full service lifecycle.

27. internal/apps/sm-kms/server/repository/orm/orm_transaction_test.go
    - Core+DB orchestration via StartCore, migrations, AutoMigrate, repository setup.
    - Primarily DB/service fixture orchestration, no long-lived HTTP server usage.

28. internal/apps/sm-kms/server/testmain_test.go
    - Server integration TestMain: starts KMS server with SQLite and helper-based dual-port wait.

### Goal 2 Completion Criteria

1. Every TestMain in internal/apps/** accounted for and categorized.
2. Each file mapped to orchestration intent: e2e, integration app, DB fixture, API fixture, CLI fixture, or sad-path.
3. No PS-ID omitted from inventory.

## Goal 3 - Refactor Every TestMain to Framework Orchestration Reuse

The third implementation goal is execution mapping: each existing TestMain refactored to reusable framework orchestration modules.

### Refactor Policy

1. Happy-path service startup in PS-ID packages must only occur through test_orchestration/test_integration (integration) or test_orchestration/test_e2e (e2e).
2. DB fixture startup must use test_orchestration/test_db or test_orchestration/test_barrier.
3. API-specific fixture startup must use test_orchestration/test_api.
4. CLI-specific fixture startup must use test_orchestration/test_cli.
5. Sad-path startup tests are allowed only when server/service start is expected to fail and must use dedicated fail-intent orchestration helpers.

### Goal 3 Numbered List - How every TestMain will be refactored

1. identity-authz/e2e/testmain_e2e_test.go
   - Replace pass-through with test_e2e factory wrapper and PS-ID config object.

2. identity-authz/server/testmain_test.go
   - Replace manual startup/readiness/shutdown with test_integration orchestration call.

3. identity-idp/e2e/testmain_e2e_test.go
   - Replace pass-through with test_e2e orchestration wrapper.

4. identity-idp/server/testmain_test.go
   - Replace manual startup with test_integration orchestration wrapper.

5. identity-rp/e2e/testmain_e2e_test.go
   - Replace pass-through with test_e2e orchestration wrapper.

6. identity-rp/server/testmain_test.go
   - Migrate startup and TLS client setup to test_integration + test_tls.

7. identity-rs/e2e/testmain_e2e_test.go
   - Replace pass-through with test_e2e orchestration wrapper.

8. identity-rs/server/testmain_test.go
   - Replace manual startup with test_integration orchestration wrapper.

9. identity-spa/e2e/testmain_e2e_test.go
   - Replace pass-through with test_e2e orchestration wrapper.

10. identity-spa/server/testmain_test.go
    - Replace local wait helper and startup with test_integration orchestration wrapper.

11. jose-ja/e2e/testmain_e2e_test.go
    - Keep e2e factory usage but migrate from legacy e2e_infra API to test_e2e package facade.

12. jose-ja/server/repository/testmain_test.go
    - Replace local SQLite setup with test_db factory for repository fixtures.

13. jose-ja/server/service/testmain_test.go
    - Replace local fixture graph with test_db + test_barrier composite fixture.

14. jose-ja/server/testmain_test.go
    - Replace current helper usage with test_integration package for consistent API.

15. pki-ca/e2e/testmain_e2e_test.go
    - Replace custom compose start/stop with test_e2e factory.

16. pki-ca/server/testmain_test.go
    - Replace package-local orchestration with test_integration + test_tls.

17. skeleton-template/e2e/testmain_e2e_test.go
    - Repoint to test_e2e facade package.

18. skeleton-template/server/testmain_test.go
    - Repoint to test_integration + test_tls facade package.

19. sm-im/client/testmain_test.go
    - Remove direct happy-path server startup.
    - Reuse shared integration orchestration from test_integration (or move tests to integration package).

20. sm-im/e2e/testmain_e2e_test.go
    - Repoint to test_e2e facade package.

21. sm-im/server/apis/messages_test.go
    - Replace local DB+barrier fixture bootstrapping with test_api + test_barrier.

22. sm-im/server/repository/testmain_test.go
    - Replace local DB+barrier fixture bootstrapping with test_db + test_barrier.

23. sm-im/server/testmain_test.go
    - Replace PS-ID-local SetupTestServer path with test_integration factory.

24. sm-kms/client/testmain_test.go
    - Remove direct happy-path server startup.
    - Reuse test_integration orchestration from framework.

25. sm-kms/e2e/testmain_e2e_test.go
    - Repoint to test_e2e facade package.

26. sm-kms/server/businesslogic/businesslogic_crud_test.go
    - Keep as sad-path/setup-only TestMain if no happy-path startup; migrate env setup helper into test_bootstrap.

27. sm-kms/server/repository/orm/orm_transaction_test.go
    - Replace custom core+DB lifecycle with test_db or test_api DB-core fixture orchestrator.

28. sm-kms/server/testmain_test.go
    - Repoint to test_integration orchestration.

### Goal 3 Completion Criteria

1. All 28 TestMain files migrated or removed according to mapping.
2. No PS-ID package performs ad-hoc happy-path startup logic outside framework orchestration packages.
3. Remaining local TestMain files are wrappers only, with PS-ID config injection.
4. Sad-path-only startup tests explicitly marked and validated by linter allowlist rules.

## Enforcement Strategy - Templates and cicd Linters

### Template Enforcement

1. Update canonical template tree at api/cryptosuite-registry/templates/internal/apps/__PS_ID__/.
2. Add integration and e2e orchestration wrapper templates under template package layout.
3. Make server/testmain_test.go generated wrappers thin, calling framework test_orchestration modules.
4. Ensure all 10 PS-ID instantiations remain in exact-match compliance where required.

### cicd-lint Fitness Enforcement

Add or extend lint-fitness checks:
1. testmain-orchestration-policy
   - Detect all func TestMain under internal/apps/**.
   - Classify startup operations.
   - Fail if happy-path startup bypasses internal/apps-framework/service/test_orchestration/test_*.

2. testmain-location-policy
   - Enforce allowed happy-path startup package locations.
   - Disallow ad-hoc startup in client/repository/service/apis packages unless using framework orchestration facade.

3. testmain-sadpath-policy
   - Require fail-intent marker/comment or explicit helper when start is expected to fail.
   - Block ambiguous local startup patterns.

4. template-testmain-compliance
   - Ensure template-generated TestMain wrappers reference framework orchestration packages.

5. psid-template-instantiation-sync
   - Verify all 10 PS-ID instantiations remain aligned with updated canonical templates.

## Multi-Phase Execution Plan

Phase 1: Research Freeze and Baseline Evidence
1. Freeze inventory of all TestMain in internal/apps/**.
2. Record behavior classification and duplicate-pattern analysis.
3. Capture evidence artifacts in test-output/completion-verification/.

Phase 2: Orchestration API Design
1. Define package APIs for test_e2e/test_integration/test_db/test_api/test_cli and additional candidates.
2. Define lifecycle contracts (setup, ready wait, clients, teardown).
3. Define standard timeout constants and naming conventions.

Phase 3: Implement Framework Orchestration Packages
1. Create each test_orchestration package.
2. Port reusable logic from current framework testing helpers.
3. Add package tests and coverage.

Phase 4: Migrate Existing Framework Callers
1. Migrate existing e2e_infra consumers to new test_e2e facade.
2. Migrate server startup helper consumers to test_integration facade.
3. Keep compatibility wrappers temporarily to avoid broad breakage.

Phase 5: PS-ID-by-PS-ID TestMain Refactor
1. identity family migration (authz, idp, rp, rs, spa).
2. jose-ja migration.
3. pki-ca migration.
4. skeleton-template migration.
5. sm-im migration.
6. sm-kms migration.

Phase 6: Remove Duplicate Local Fixture Graphs
1. Remove duplicated SQLite PRAGMA/GORM setup from local TestMain files.
2. Replace local barrier fixture setup with framework orchestrators.
3. Replace client-package full-server startup with integration orchestrator reuse.

Phase 7: Template Updates
1. Update template files and MANIFEST constraints as required.
2. Re-instantiate or sync all 10 PS-ID generated files.
3. Ensure lint-fitness template conformance passes.

Phase 8: Linter Implementation and Policy Lock
1. Implement new lint-fitness validators.
2. Add positive and negative tests for policy enforcement.
3. Wire checks into bulk cicd-lint runs.

Phase 9: Validation and Rollout
1. Run full go build/lint/test gates.
2. Run e2e for all relevant PS-IDs.
3. Confirm no policy violations remain.
4. Document migration completion and enforce ongoing guardrails.

## Risks and Mitigations

1. Risk: Hidden startup dependencies in local TestMain files.
   - Mitigation: staged migration with compatibility wrappers and per-package tests.

2. Risk: Over-generalizing orchestration API and losing PS-ID edge cases.
   - Mitigation: function-parameter injection and PS-ID config structs.

3. Risk: Template drift across 10 PS-IDs.
   - Mitigation: strict template sync checks and explicit lint-fitness verification.

4. Risk: Build-tag or package-boundary test regressions.
   - Mitigation: keep wrappers local to package while orchestration logic is framework-owned.

5. Risk: New linter false positives.
   - Mitigation: signal-based detection plus curated allowlists for sad-path tests.

## Quality Gates

Per phase mandatory checks:
1. go build ./...
2. go build -tags e2e,integration ./...
3. golangci-lint run ./...
4. golangci-lint run --build-tags e2e,integration ./...
5. go test ./... -shuffle=on
6. go test -race -count=2 ./... (where applicable)
7. go run ./cmd/cicd-lint lint-fitness
8. go run ./cmd/cicd-lint lint-deployments (only if deployment files change)

Coverage and mutation targets:
1. Production packages >=95%.
2. Infrastructure utility packages >=98%.
3. Mutation efficacy >=95% minimum, >=98% for infrastructure utility.

## Evidence

Current evidence artifacts for this planning phase:
1. test-output/completion-verification/framework-v22-testmain-signals.txt
2. test-output/completion-verification/framework-v22-testmain-signals.csv
3. test-output/completion-verification/framework-v22-plan-completeness-check.txt
4. test-output/completion-verification/framework-v22-goal2-goal3-coverage.txt

Quizme status:
1. quizme-v1.md not created for this planning cycle.
2. Reason: no unresolved architectural unknowns blocked construction of Goal 1/2/3 exhaustive lists.

## Appendix B - Plan and Task Deep Completeness Audit

Audit objective: verify that plan.md and tasks.md fully cover all 10 PS-IDs and all internal/apps TestMain files, with no omissions in Goal 1/2/3 numbered lists.

Audit checks executed:
1. Enumerated canonical TestMain inventory from source tree using grep over internal/apps/**.
2. Verified every discovered TestMain path appears in plan.md.
3. Verified every discovered TestMain path appears in Goal 2 section.
4. Verified every discovered TestMain path appears in Goal 3 section.
5. Verified PS-ID coverage matrix includes all 10 PS-IDs.
6. Verified tasks.md phase/task structure includes Goal 1 build-out, Goal 2 analysis, Goal 3 migration, and template+linter enforcement work.

Audit results:
1. Total discovered TestMain files in internal/apps/**: 28.
2. Missing TestMain path references in plan.md: 0.
3. Missing TestMain entries in Goal 2 list: 0.
4. Missing TestMain entries in Goal 3 list: 0.
5. Missing PS-IDs in plan coverage: 0.
6. Missing goal-level implementation tracks in tasks.md: 0.

Audit conclusion:
1. The three required numbered lists are complete and aligned with repository baseline evidence.
2. The plan is ready for implementation execution under docs/framework-v22/tasks.md.

## Appendix A - Deep Analysis (Baseline Research)

### A.1 Distinct TestMain Functional Types Identified

1. Integration happy-path server startup (direct Go app startup + SQLite + dynamic ports).
2. E2E happy-path service startup (Docker Compose with PostgreSQL and SQLite instances).
3. DB-only fixture startup (SQLite/GORM/migrations and optional service fixtures).
4. API fixture startup (handler and repository graph without full service startup).
5. Sad-path/setup-only startup (environment setup or fail-intent without happy-path service run).

### A.2 Current Duplication Findings

1. Server startup logic is duplicated across multiple PS-ID server TestMain files.
2. Client package TestMain in sm-im and sm-kms starts full happy-path server, duplicating integration startup.
3. DB fixture setup (SQLite PRAGMA, GORM setup, migrations, telemetry/JWK/barrier) repeated across jose-ja and sm-im sub-packages.
4. E2E startup is inconsistent: some PS-IDs use framework factory, one uses custom compose startup, identity PS-IDs currently pass through only.

### A.3 Distinct Patterns by Category

1. Framework e2e factory users:
   - jose-ja/e2e
   - sm-kms/e2e
   - sm-im/e2e
   - skeleton-template/e2e

2. Custom compose e2e startup:
   - pki-ca/e2e

3. E2E pass-through placeholders:
   - identity-authz/e2e
   - identity-idp/e2e
   - identity-rp/e2e
   - identity-rs/e2e
   - identity-spa/e2e

4. Server startup with helper-based dual-port wait:
   - jose-ja/server
   - pki-ca/server
   - skeleton-template/server
   - sm-kms/server

5. Server startup with local manual wait loops:
   - identity-authz/server
   - identity-idp/server
   - identity-rp/server
   - identity-rs/server
   - identity-spa/server

6. PS-ID local helper based startup:
   - sm-im/server (SetupTestServer)
   - sm-im/client (StartSmIMService)

7. DB-only fixture TestMain set:
   - jose-ja/server/repository
   - jose-ja/server/service
   - sm-im/server/repository
   - sm-im/server/apis/messages_test.go
   - sm-kms/server/repository/orm

8. Sad-path or setup-only TestMain:
   - sm-kms/server/businesslogic/businesslogic_crud_test.go

### A.4 Why Centralization is Required

1. Determinism: one orchestration implementation for startup/ready/teardown semantics.
2. Maintainability: avoid repeated fixes across 10 PS-IDs.
3. Policy compliance: codify what is allowed as happy-path startup.
4. Linter enforceability: easier static checks when orchestration entrypoints are centralized.
5. Test reliability: unified timeout, keepalive, TLS client, and cleanup behavior.

### A.5 Baseline Decision

Adopt framework-owned orchestration modules in internal/apps-framework/service/test_orchestration/test_*/ as the only allowed happy-path startup path. PS-ID TestMain files remain only as thin wrappers or are deleted when wrappers are unnecessary.
