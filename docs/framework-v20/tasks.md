# Tasks - Framework V20: TLS Enum Split and Policy Surface Migration

**Status**: 0 of 30 tasks complete (0%)
**Created**: 2026-04-29
**Last Updated**: 2026-04-29

## Task Status Legend

| Symbol | Meaning |
|--------|---------|
| ❌ | Not started |
| 🔄 | In progress |
| ✅ | Complete |
| ⏳ | Blocked |

## Decision Constraints

1. `docs/ENG-HANDBOOK.md` is the source of truth and must be updated before downstream propagation or code completion claims.
2. `TLSProvisionMode` and `TLSClientPolicy` must remain separate concerns in docs, config, code, and instructions.
3. CA-file presence must not remain the implicit client-policy selector.
4. Deployment and template updates are not complete until `lint-deployments` passes.
5. Historical docs are in scope and must be updated after active docs are correct.

## Phase 1: Canonical Documentation Split

### Task 1.1: Define enum split in handbook

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/ENG-HANDBOOK.md` defines `TLSProvisionMode` and `TLSClientPolicy` separately.
  - [ ] Handbook text clearly states provisioning and client-auth policy are separate axes.
  - [ ] Old `TLSModeOff|On|Mixed` runtime terminology is removed from active handbook guidance.

### Task 1.2: Define propagation chunk boundaries

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Handbook contains propagation-ready chunk boundaries for provisioning guidance.
  - [ ] Handbook contains propagation-ready chunk boundaries for client-policy guidance.
  - [ ] Chunk scope avoids duplicating both full enum blocks into both instruction files.

### Task 1.3: Align companion TLS structure doc

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/tls-structure.md` mirrors handbook terminology for both enums.
  - [ ] Runtime examples use `TLSClientPolicy` vocabulary instead of old runtime `TLSMode*` names.
  - [ ] Companion doc does not contradict the handbook.

### Task 1.4: Phase 1 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Manual contradiction review between handbook and companion TLS doc completed.
  - [ ] Evidence archived in `test-output/v20-phase1/`.

## Phase 2: Propagation and Documentation Integrity

### Task 2.1: Propagate provisioning guidance to architecture instructions

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `.github/instructions/02-01.architecture.instructions.md` receives the handbook provisioning chunk.
  - [ ] Any needed glue text remains outside propagated content.

### Task 2.2: Propagate client-policy guidance to security instructions

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `.github/instructions/02-05.security.instructions.md` receives the handbook client-policy chunk.
  - [ ] Security wording reflects the five-policy model accurately.

### Task 2.3: Update required propagation registry

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/required-propagations.yaml` includes any new chunk IDs/targets.
  - [ ] No propagation target is missing from the registry.

### Task 2.4: Phase 2 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes.
  - [ ] Evidence archived in `test-output/v20-phase2/`.

## Phase 3: Provisioning Enum Refactor

### Task 3.1: Rename framework provisioning type and constants

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `TLSMode` becomes `TLSProvisionMode` in framework config code.
  - [ ] Provisioning constants become `TLSProvisionModeStatic|Mixed|Auto`.
  - [ ] Comments/field docs describe certificate material sourcing, not client-auth policy.

### Task 3.2: Update parser/settings/default surfaces

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] CLI/config registration/help text uses provisioning terminology.
  - [ ] Parser/default paths use the new provisioning type.
  - [ ] Magic/default comments are aligned.

### Task 3.3: Update provisioning call sites and tests

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Builder/listener/testutil/test files compile with `TLSProvisionMode` names.
  - [ ] Targeted tests covering renamed provisioning surfaces pass.

### Task 3.4: Align schema descriptions for provisioning keys

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Deployment schema descriptions/valid values match the real provisioning enum surface.
  - [ ] No schema text still implies `auto|manual` if the actual surface is `static|mixed|auto`.

### Task 3.5: Phase 3 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes.
  - [ ] Targeted provisioning tests pass.
  - [ ] Evidence archived in `test-output/v20-phase3/`.

## Phase 4: Explicit Client Policy Surface

### Task 4.1: Add config fields/settings/parser for `TLSClientPolicy`

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Public and admin client-policy fields exist in framework config.
  - [ ] CLI/config parsing supports the five policy values.
  - [ ] Defaults are explicitly documented.

### Task 4.2: Centralize policy-to-Go TLS mapping

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Shared TLS code exposes a clear mapping from `TLSClientPolicy` to `tls.ClientAuthType`.
  - [ ] Mapping tests cover all five states and fallback behavior.

### Task 4.3: Make public/admin listeners policy-driven

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `applyPublicMTLS()` honors configured client policy instead of always enforcing `RequireAndVerifyClientCert`.
  - [ ] `applyAdminMTLS()` honors configured client policy instead of always enforcing `RequireAndVerifyClientCert`.
  - [ ] CA-file presence is treated as trust material input, not policy selection.

### Task 4.4: Expand runtime behavior tests

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Public/admin listener tests cover all supported client-policy outcomes.
  - [ ] Application listener expectations are updated to explicit policy assertions.
  - [ ] Identity authz note/comments are aligned to the new policy model.

### Task 4.5: Phase 4 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go test` passes for shared TLS and listener packages.
  - [ ] Evidence archived in `test-output/v20-phase4/`.

## Phase 5: Deployment Schema, Config, and Template Alignment

### Task 5.1: Extend deployment schema for client-policy keys

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Schema includes explicit public/admin client-policy keys.
  - [ ] Validation rules enforce the intended enum values.

### Task 5.2: Update PS-ID deployment overlays

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All 10 common config files are aligned to provisioning terminology where needed.
  - [ ] All 40 instance overlay files set explicit client-policy keys where CA trust files are used.
  - [ ] No overlay relies on implicit `Enforce` semantics.

### Task 5.3: Update canonical deployment templates

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] All 5 canonical `__PS_ID__` config templates reflect the new provisioning and client-policy model.
  - [ ] Template changes remain synchronized with concrete deployment overlays.

### Task 5.4: Update legacy `tls-config.yml` examples

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] The 4 `deployments/*/certs/tls-config.yml` files no longer use old runtime `TLSMode*` names.
  - [ ] Examples are consistent with the handbook terminology.

### Task 5.5: Phase 5 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes.
  - [ ] Evidence archived in `test-output/v20-phase5/`.

## Phase 6: Historical Document Backfill

### Task 6.1: Update investigation doc terminology

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/INVESTIGATION.md` no longer leaves stale active terminology around `TLSModeAuto`.
  - [ ] Historical context remains readable and accurate.

### Task 6.2: Update archived v18 planning references

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/framework-v18/plan.md` no longer references obsolete enum names without explanation.
  - [ ] Archived context still reads coherently.

### Task 6.3: Phase 6 verification

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Historical-doc pass completed and contradictions with active docs removed.
  - [ ] Evidence archived in `test-output/v20-phase6/`.

## Phase 7: Final Verification and Closure

### Task 7.1: Full quality suite

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `go build ./...` passes.
  - [ ] `go build -tags e2e,integration ./...` passes.
  - [ ] `go test ./...` passes.
  - [ ] `golangci-lint run ./...` passes.
  - [ ] `go run ./cmd/cicd-lint lint-fitness` passes.
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes.
  - [ ] `go run ./cmd/cicd-lint lint-deployments` passes.

### Task 7.2: Refresh final inventory and contradiction review

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Final inventory of changed surfaces is archived under `test-output/v20-phase7/`.
  - [ ] `docs/framework-v20/plan.md`, `docs/framework-v20/tasks.md`, and repository reality are consistent.
  - [ ] If any item remains uncertain, mark it unresolved instead of guessing.

## Phase 8: Knowledge Propagation

### Task 8.1: Review execution lessons for durable guidance updates

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] `docs/framework-v20/lessons.md` is reviewed after execution phases complete.
  - [ ] Any new durable guidance for handbook or instructions is explicitly identified.

### Task 8.2: Apply durable documentation/process updates

- **Status**: ❌
- **Acceptance Criteria**:
  - [ ] Durable doc/process updates discovered during V20 are applied.
  - [ ] `go run ./cmd/cicd-lint lint-docs` passes again after those updates.
  - [ ] Evidence archived in `test-output/v20-phase8/`.
