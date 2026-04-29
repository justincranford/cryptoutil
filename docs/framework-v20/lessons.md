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

*(To be filled at plan completion - numbered links to each phase section with one-sentence outcome.)*

## Actions

*(To be filled at plan completion - numbered list of concrete follow-up items for reviewer.)*

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

*(To be filled during Phase 5 execution using the 4-section structure above.)*

## Phase 6: Historical Document Backfill

*(To be filled during Phase 6 execution using the 4-section structure above.)*

## Phase 7: Verification and Closure

*(To be filled during Phase 7 execution using the 4-section structure above.)*

## Phase 8: Knowledge Propagation

*(To be filled during Phase 8 execution using the 4-section structure above.)*
