# Implementation Plan - Framework V20: TLS Enum Split and Policy Surface Migration

**Status**: Not started
**Created**: 2026-04-29
**Last Updated**: 2026-04-29
**Purpose**: Introduce the explicit `TLSProvisionMode` and `TLSClientPolicy` model across handbook, propagated instructions, framework config/schema, runtime listener behavior, deployment artifacts, historical docs, and validation suites.

## Overview

Framework V20 converts one overloaded TLS naming surface into two explicit axes:
1. `TLSProvisionMode` for server certificate material sourcing.
2. `TLSClientPolicy` for TLS-layer client certificate request/require/verification behavior.

The primary goal is correctness of meaning, not only renaming. The current repository mixes:
- provisioning semantics (`static`, `mixed`, `auto`)
- runtime client-auth semantics (`off`, `optional`, `required`, `verified`-style behavior)
- implicit behavior where CA-file presence hardcodes `tls.RequireAndVerifyClientCert`

V20 resolves that ambiguity by updating source-of-truth docs first, propagating the new terminology into instruction files, then aligning code/config/deployment surfaces to the new split.

## Evidence Inputs

- `test-output/FILES.txt`
- `docs/ENG-HANDBOOK.md`
- `docs/tls-structure.md`
- `.github/instructions/02-01.architecture.instructions.md`
- `.github/instructions/02-05.security.instructions.md`
- `docs/required-propagations.yaml`
- `internal/apps-framework/service/config/config.go`
- `internal/apps-framework/service/config/config_parse.go`
- `internal/apps-framework/service/config/config_settings_*.go`
- `internal/apps-framework/service/server/listener/{public,admin,servers}.go`
- `internal/shared/crypto/tls/config.go`
- `internal/apps-tools/cicd_lint/lint_deployments/validate_schema.go`
- `deployments/*/config/*-app-framework-*.yml`
- `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/*`

## Problem Statement

The current repository has four coupled problems:
1. `TLSMode*` is used for certificate provisioning in code but is also used as if it described client-auth runtime behavior in docs.
2. Listener runtime behavior is implicitly derived from CA-file presence instead of an explicit client-policy enum.
3. Deployment overlays and schema do not expose explicit client-policy controls.
4. Propagated instruction content still reflects the old blended terminology.

## Non-Goals

- Re-architect TLS generation internals beyond what is necessary to express `TLSProvisionMode` cleanly.
- Change unrelated PKI category naming, OTel TLS, or PostgreSQL mTLS semantics unless required by the enum split.
- Rewrite archived framework-v16/v17 artifacts unless they are used as active references in V20 scope.

## Key Decisions

1. `docs/ENG-HANDBOOK.md` is the canonical source and must be updated before code/config terminology changes are treated as complete.
2. `TLSProvisionMode` and `TLSClientPolicy` are separate concerns and should propagate to different instruction files by concern.
3. CA bundle path presence must no longer be the policy selector; it is trust material input only.
4. Historical docs (`docs/INVESTIGATION.md`, `docs/framework-v18/plan.md`) are updated last, but they are in scope.

## Affected Files

- Canonical docs and propagation: `docs/ENG-HANDBOOK.md`, `docs/tls-structure.md`, `.github/instructions/{02-01.architecture.instructions.md,02-05.security.instructions.md}`, `docs/required-propagations.yaml` (5 files).
- Provisioning config/type surfaces: `internal/apps-framework/service/config/{config.go,config_settings_a.go,config_parse.go,config_sequential_test.go,config_parse_profiles_test.go,config_test_helper.go,config_test_util.go}` (7 files).
- Provisioning runtime/builder/test surfaces: `internal/apps-framework/service/server/{testutil/helpers.go,listener/{servers.go,servers_test.go,listener_errorpaths_test.go},builder/{server_builder_build.go,server_builder_build_test.go,server_builder_migrations_test.go,server_builder_tls_test.go}}` (9 files).
- Shared provisioning/docs helpers: `internal/apps-framework/service/config/tls_generator/tls_generator.go`, `internal/apps-framework/tls/generator_helpers.go`, `internal/shared/magic/magic_network.go` (3 files).
- Explicit client-policy implementation surfaces: `internal/apps-framework/service/config/{config.go,config_settings_b.go,config_parse.go}`, `internal/shared/crypto/tls/{config.go,tls_test.go,tls_mutation_test.go,tls_load_store_error_paths_test.go}`, `internal/apps-framework/service/server/listener/{public.go,admin.go,public_mtls_test.go,admin_mtls_test.go}`, `internal/apps-framework/service/application/application_listener_test.go`, `internal/apps/identity-authz/server/apis/client_authentication.go` (13 files).
- Deployment schema and config overlays: `internal/apps-tools/cicd_lint/lint_deployments/validate_schema.go`, `deployments/{sm-im,sm-kms,jose-ja,pki-ca,identity-authz,identity-idp,identity-rs,identity-rp,identity-spa,skeleton-template}/config/{*-app-framework-common.yml,*-app-framework-sqlite-1.yml,*-app-framework-sqlite-2.yml,*-app-framework-postgresql-1.yml,*-app-framework-postgresql-2.yml}` (1 + 50 = 51 files).
- Canonical deployment templates and legacy examples: `api/cryptosuite-registry/templates/deployments/__PS_ID__/config/{__PS_ID__-app-framework-common.yml,__PS_ID__-app-framework-sqlite-1.yml,__PS_ID__-app-framework-sqlite-2.yml,__PS_ID__-app-framework-postgresql-1.yml,__PS_ID__-app-framework-postgresql-2.yml}`, `deployments/{sm-im,sm-kms,jose-ja,skeleton-template}/certs/tls-config.yml` (5 + 4 = 9 files).
- Historical documents: `docs/INVESTIGATION.md`, `docs/framework-v18/plan.md` (2 files).

Total planned touch surface: `5 + 7 + 9 + 3 + 13 + 51 + 9 + 2 = 99 files`.

## Phase Summary

### Phase 1: Canonical Documentation Split

**Objective**: Define the authoritative semantics in `docs/ENG-HANDBOOK.md` and align `docs/tls-structure.md`.

- Replace blended runtime terminology with explicit `TLSClientPolicy` vocabulary.
- Replace provisioning names with `TLSProvisionMode` vocabulary.
- Define handbook propagation chunk boundaries for each enum.
- Ensure runtime examples distinguish trust material from client-auth policy.

**Success**:
- Handbook explicitly defines both enums and their responsibilities.
- Companion TLS structure doc mirrors the same model without contradiction.

### Phase 2: Propagation and Documentation Integrity

**Objective**: Push the handbook changes into instruction artifacts and validate drift.

- Propagate `TLSProvisionMode` guidance into architecture instructions.
- Propagate `TLSClientPolicy` guidance into security instructions.
- Update `docs/required-propagations.yaml` if chunk targets or IDs change.
- Run propagation validation.

**Success**:
- `lint-docs` passes with no drift.
- Instruction files reflect the new split without duplicating both enum blocks everywhere.

### Phase 3: Provisioning Enum Refactor in Framework Code

**Objective**: Rename provisioning code surfaces from `TLSMode` to `TLSProvisionMode`.

- Update framework config type/fields/comments.
- Update parser/settings/defaults/tests.
- Update server builder/listener provisioning switching.
- Update provisioning helper comments and schema descriptions.

**Success**:
- No code path still uses `TLSMode` for certificate provisioning.
- Tests and defaults compile with `TLSProvisionMode` terminology.

### Phase 4: Explicit Client Policy Surface

**Objective**: Introduce explicit `TLSClientPolicy` configuration and mapping to Go TLS client-auth behavior.

- Add config fields/settings/parser support.
- Add central policy-to-`tls.ClientAuthType` mapping.
- Update listener runtime behavior to honor configured policy.
- Expand tests to cover the five policy states.

**Success**:
- Public/admin listener behavior is policy-driven, not inferred from CA-file presence.
- Shared TLS helpers and listener tests cover all policy states.

### Phase 5: Deployment Schema, Config, and Template Alignment

**Objective**: Surface the new enum model in deployment config overlays and templates.

- Update deployment schema for provisioning and client-policy fields.
- Update 10 PS-ID common configs if provisioning terminology changes there.
- Update 40 instance overlays to add explicit public/admin client-policy keys where CA-file paths are already present.
- Update 5 canonical deployment template config files.
- Update the 4 legacy `tls-config.yml` doc/example files that still use old runtime names.

**Success**:
- Deployment overlays no longer rely on implicit `Enforce` semantics.
- Template and concrete deployment artifacts remain aligned.

### Phase 6: Historical Document Backfill

**Objective**: Normalize old names in investigative and planning docs after active surfaces are correct.

- Update `docs/INVESTIGATION.md` terminology where the old names are no longer acceptable.
- Update `docs/framework-v18/plan.md` archived references to the new naming where appropriate.

**Success**:
- Historical docs no longer undermine current terminology.
- Any retained historical phrasing is explicit and intentional.

### Phase 7: Verification and Closure

**Objective**: Prove the repo is internally consistent after the enum split.

- Run build, tests, lint, docs propagation, and deployment validation.
- Refresh the FILES-style inventory if needed.
- Reconcile V20 plan/tasks/evidence with actual repository state.

**Success**:
- Build/lint/docs/deployment validators pass.
- No unresolved contradiction remains between docs, config schema, code behavior, and deployment overlays.

### Phase 8: Knowledge Propagation

**Objective**: Fold execution lessons back into durable repo guidance after the implementation stabilizes.

- Review `docs/framework-v20/lessons.md` across all prior phases.
- Update durable handbook, instruction, or skill guidance if V20 execution uncovers new repeatable patterns.
- Update this plan/tasks pair if execution reveals scope drift that must be made explicit.
- Re-run propagation validation after any durable guidance update.

**Success**:
- Lessons are converted into durable guidance where warranted.
- No new propagation drift is introduced by post-implementation documentation updates.

## Quality Gates

- `go build ./...`
- `go build -tags e2e,integration ./...`
- `go test ./...`
- `golangci-lint run ./...`
- `go run ./cmd/cicd-lint lint-docs`
- `go run ./cmd/cicd-lint lint-fitness`
- `go run ./cmd/cicd-lint lint-deployments` (mandatory when config/template/deployment/schema files change)

## Evidence Strategy

Archive V20 execution evidence under:
- `test-output/v20-phase1/`
- `test-output/v20-phase2/`
- `test-output/v20-phase3/`
- `test-output/v20-phase4/`
- `test-output/v20-phase5/`
- `test-output/v20-phase6/`
- `test-output/v20-phase7/`
- `test-output/v20-phase8/`

## Risks and Watchpoints

1. Propagation drift risk: handbook chunk changes will break instruction copies if not updated in the same phase.
2. Schema/template drift risk: deployment config changes must be mirrored into canonical templates immediately.
3. Behavior-regression risk: listener runtime semantics may change unintentionally if policy defaults are not specified carefully.
4. Naming-scope risk: docs may incorrectly restate runtime policy as provisioning or vice versa if the split is not enforced consistently.
5. Historical-doc risk: stale old names in investigative docs can reintroduce ambiguity during later archaeology.
