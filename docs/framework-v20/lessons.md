# Lessons - Framework V20: TLS Enum Split and Policy Surface Migration

**Created**: 2026-04-29
**Purpose**: Phase post-mortem lessons for Framework V20. Populate after each phase passes its quality gates.

> **MANDATORY per-phase structure** (4 sections per phase):
>
> **What Worked**: Patterns, tools, or decisions that accelerated the work or prevented issues.
>
> **What Didn't Work**: Friction points, incorrect assumptions, or approaches that required rework.
>
> **Root Causes**: Underlying causes of the "What Didn't Work" items.
>
> **Patterns for Future Phases**: Actionable takeaways for subsequent phases or future plans.

## Executive Summary

1. [Phase 1: Canonical Documentation Split](#phase-1-canonical-documentation-split) — Split TLSProvisionMode and TLSClientPolicy into two independent axes in handbook and tls-structure.md.
2. [Phase 2: Propagation and Documentation Integrity](#phase-2-propagation-and-documentation-integrity) — Propagated both axes to architecture and security instruction files; lint-docs confirmed zero drift.
3. [Phase 3: Provisioning Enum Refactor](#phase-3-provisioning-enum-refactor-in-framework-code) — Renamed TLSMode to TLSProvisionMode with three concrete constants; all call sites, schema text, and tests updated.
4. [Phase 4: Explicit Client Policy Surface](#phase-4-explicit-client-policy-surface) — Added TLSClientPolicy config fields, centralised policy→tls.ClientAuthType mapper, made listeners policy-driven.
5. [Phase 5: Deployment Schema, Config, and Template Alignment](#phase-5-deployment-schema-config-and-template-alignment) — Extended schema allowlists, patched 40 overlay + 4 template files, updated 4 legacy tls-config.yml; lint-deployments 54/54 pass.
6. [Phase 6: Historical Document Backfill](#phase-6-historical-document-backfill) — Replaced stale TLSModeAuto references in INVESTIGATION.md and framework-v18/plan.md; no stale terms remain.
7. [Phase 7: Verification and Closure](#phase-7-verification-and-closure) — All quality gates passed: build, tagged build, full test suite, lint-fitness, lint-go, lint-docs, lint-deployments.

## Actions

1. Completed 2026-05-01: All 10 PS-IDs already had `server-admin-tls-client-policy` and `server-public-tls-client-policy` in their deployment overlays (implemented in Phase 5). Template updated. Forward-looking: new PS-IDs inherit the policy keys from the overlay templates automatically.
2. Completed 2026-05-01: Added TLS client policy comment to `api/cryptosuite-registry/templates/configs/__PS_ID__/__PS_ID__-framework.yml` and all 10 `configs/{PS-ID}/{PS-ID}-framework.yml` instances documenting that the default `TLSClientPolicy` is `none`.
3. Completed 2026-04-29: Added `lint-fitness` rule `config-tls-ca-policy-coupling` that rejects deployment overlay files with `*-tls-ca-file` but missing corresponding `*-tls-client-policy` key.
4. Completed 2026-05-01: Added two-axis TLS model section to `.github/skills/new-service/SKILL.md` and `.claude/skills/new-service/SKILL.md`, covering TLSProvisionMode (automatic) vs TLSClientPolicy (must be set explicitly in deployment overlays).

## Phase 1: Canonical Documentation Split

**What Worked**

- Updating the handbook first established the authoritative split between provisioning and client-certificate policy before any downstream code changes.
- Keeping the first slice limited to the handbook and TLS structure doc made contradiction review fast and local.

**What Didn't Work**

- The execution tracker was not updated in the same slice, which made later completion accounting noisier than it needed to be.

**Root Causes**

- The initial implementation path prioritized source-of-truth prose over plan bookkeeping, even though the bookkeeping is part of the phase gate.

**Patterns for Future Phases**

- Land the source-of-truth change, evidence, and tracker update together in the same semantic unit.

## Phase 2: Propagation and Documentation Integrity

**What Worked**

- Treating `lint-docs` as the first discriminating check caught propagation gaps immediately after the handbook change.
- Splitting the new handbook text into two narrow propagated chunks avoided copying unrelated TLS guidance into both instruction files.

**What Didn't Work**

- The first propagation attempt was incomplete until the registry and target files were updated together.

**Root Causes**

- The propagation system is a three-surface contract: source block, target block, and required-propagations manifest. Missing any one surface invalidates the phase.

**Patterns for Future Phases**

- Treat every new propagated chunk as a three-part change and validate it with `lint-docs` before moving on.

## Phase 3: Provisioning Enum Refactor in Framework Code

**What Worked**

- Using symbol-aware renames for typed Go surfaces reduced the risk of partial `TLSMode` leftovers in the main framework code.
- Narrow package validation across config, listener, TLS helper, and deployment schema gave fast feedback without widening scope.

**What Didn't Work**

- The type rename alone did not clean up error strings, schema descriptions, and test names, so a second manual pass was still required.

**Root Causes**

- Symbol-aware tooling does not update string literals, YAML keys, help text, or human-facing assertions.

**Patterns for Future Phases**

- After a semantic rename, explicitly review strings, schema text, and tests instead of assuming the symbol rename completed the migration.

## Phase 4: Explicit Client Policy Surface

**What Worked**

- Centralizing the mapping from `TLSClientPolicy` to `tls.ClientAuthType` kept the public and admin listener changes consistent and easy to validate.
- Focused tests proved the intended runtime contract: CA bundles provide trust material only, and client-certificate enforcement comes from explicit policy.
- Updating the adjacent application-level expectation test removed stale implicit-mTLS assumptions from nearby repository guidance.

**What Didn't Work**

- The first successful Phase 4 code pass still lacked a phase evidence directory and updated task ledger, which temporarily made completed work look incomplete.

**Root Causes**

- The implementation flow prioritized the code slice and focused tests before the required bookkeeping artifacts.

**Patterns for Future Phases**

- For config-surface changes, finish the phase only after code, tests, evidence, and task-state updates are all present.

## Phase 5: Deployment Schema, Config, and Template Alignment

**What Worked**

- Treating the deployment surface as a single bulk operation (one PowerShell script per sub-task) kept the overlay updates consistent and reduced the risk of per-file mistakes.
- Validating with `lint-deployments` after each sub-task (schema, overlays, templates) gave early confirmation that the deployment contract was intact.
- Adding the new keys to `validate_schema.go` before editing the YAML files meant the schema validator could immediately catch any malformed overlay.

**What Didn't Work**

- PowerShell inline here-strings with CIDR/IPv6 patterns caused repeated quoting failures when embedded in `run_in_terminal` calls; the scripts had to be written to disk and executed as files.
- Incremental regex-based patching of config files produced duplicate lines when the script was re-run after partial failures; deterministic whole-file overwrite was required to recover.
- The `config-instance-minimal` fitness function had a hardcoded allowlist that did not include the two new policy keys, causing lint-fitness to fail after the overlay patch.

**Root Causes**

- The fitness function allowlist is a second allowlist independent of `validate_schema.go`; both must be updated together when new keys are added to the deployment surface.
- Inline script quoting is fragile across shell boundaries on Windows; file-based execution is the only reliable pattern.

**Patterns for Future Phases**

- When adding new config keys, search for every allowlist and whitelist in `config_rules.go`, `validate_schema.go`, and `lint_deployments` to ensure all three surfaces are updated atomically.
- Always write complex PowerShell scripts to disk and execute them as files rather than embedding them inline.

## Phase 6: Historical Document Backfill

**What Worked**

- Using `Select-String` to scan all docs for stale enum names gave a precise list of files to update before starting.
- Limiting the patch to exact stale terms (TLSModeAuto, TLSModeAutoGenerate, TLSModePreGenerated, TLSModeMixed) kept diffs minimal and the historical narrative intact.

**What Didn't Work**

- Nothing material; this was a narrow find-and-replace pass on two files.

**Root Causes**

- Historical docs accumulate stale terminology when enum renames are not immediately backfilled. A search-based scan at the end of the plan is a reliable catch-all.

**Patterns for Future Phases**

- At the end of any enum rename plan, run a workspace-wide literal search for the old names before declaring the plan complete.

## Phase 7: Verification and Closure

**What Worked**

- Running the full pre-commit bundle (`lint-fitness lint-text lint-go lint-go-test lint-golangci lint-compose lint-openapi lint-ports lint-workflow lint-deployments lint-docs`) as a single command surfaced all remaining failures in one pass.
- Fixing the `config-instance-minimal` allowlist and the 8 literal-use violations unblocked all lints without touching any functional code.
- Renaming the `"mixed"` entity-type discriminator in `handlers_parallel_safety_test.go` to `"mixed-ops"` eliminated the false-positive literal-use collision with `TLSProvisionModeMixed`.

**What Didn't Work**

- The first full-bundle lint run revealed a `lint-fitness` failure due to the missed fitness-function allowlist update; this should have been caught during Phase 5.
- String-literal test case names that coincidentally match magic constant values trigger false-positive `literal-use` violations from the linter.

**Root Causes**

- The fitness function has two independent key allowlists (`config_rules.go` and `validate_schema.go`) that must be kept in sync. Phase 5 only updated one.
- The magic-usage linter checks all string literals against all registered magic constants, including test case `name:` fields, which can produce false positives when domain strings happen to match TLS constants.

**Patterns for Future Phases**

- After adding new config keys, run `lint-fitness` explicitly (not just `lint-deployments`) to catch allowlist gaps in `config_rules.go`.
- When naming test cases, use the magic constant via `string(ConstName)` or use a distinct discriminator string that cannot collide with registered magic values.

## Phase 8: Knowledge Propagation

**What Worked**

- The lessons captured during phases 5-7 provided clear, actionable guidance updates for the handbook.
- Adding the "dual allowlist rule" to Section 6.11.5 of the handbook will prevent the same `config-instance-minimal` failure from recurring in future plans.
- Adding the `literal-use` false-positive note to Section 14.1.1 clarifies the correct pattern for using magic constants as test case names.

**What Didn't Work**

- Nothing material; the handbook sections that needed updating were easy to locate and the changes were narrowly scoped.

**Root Causes**

- The dual-allowlist issue was not caught in Phase 5 because `lint-deployments` and `lint-fitness` run separate checks with no cross-reference guard in the workflow.

**Patterns for Future Phases**

- After any deployment config surface addition, always run BOTH `lint-deployments` AND `lint-fitness` (not just one) as part of the phase gate.
- Propagated handbook chunks only need to be re-exported when the `@propagate` region content changes. Non-propagated sections adjacent to propagated chunks can be edited freely.
