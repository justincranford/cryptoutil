# Lessons - Framework V15: Pre-Flight Gap Fixes + OTel/Grafana mTLS + Public App TLS Trust

**Created**: 2026-04-22
**Last Updated**: 2026-04-21

---

> **MANDATORY per-phase structure** (fill during execution, not planning):
>
> **What Worked** — patterns, approaches, and decisions that proved effective
>
> **What Didn't Work** — approaches tried that failed, wrong assumptions, pitfalls
>
> **Root Causes** — underlying causes for failures; root cause analysis, not symptoms
>
> **Patterns for Future Phases** — lessons to carry forward into the next phase and into
> permanent artifacts (ENG-HANDBOOK.md, agents, skills, instructions)

---

## Phase 0: Pre-Flight Gap Fixes

### What Worked

- **Batch signal-handling fix**: Adding `close(sigChan)` after `signal.Stop(sigChan)` across all
  10 service entry points in a single commit was fast and zero-risk — the pattern is identical in
  every file, making batch editing the right approach.
- **Reference implementation lookup**: `sm-im/im.go` had the correct shutdown timeout pattern.
  Reading it first before editing `sm-kms/kms.go` gave immediate confidence — no guessing.
- **`MustStartAndWaitForDualPorts` helper**: The helper existed and was well-documented. Replacing
  pki-ca's 300-attempt polling loop reduced testmain_test.go by ~20 lines with zero behavior change.
- **`lint-docs` and `lint-deployments` already present**: Task 0.1 only needed the top-level
  permissions block — both CI steps were already wired. Investigation saved from unnecessary work.
- **gofumpt auto-fix**: After adding the `e2e_helpers` import, running `gofumpt -w` auto-fixed
  the import ordering without manual intervention.

### What Didn't Work

- **Task 0.5 scope underestimate**: The plan estimated 1.5h for usage.go deduplication. The actual
  scope (8 `const` blocks → `var`, new shared package, 7 files across 4 product trees) revealed a
  `const`→`var` conversion that introduces non-obvious linter surface. Deferred to V16.
- **V13 stale content in lessons.md**: The lessons.md file had a stale V13 phase list appended
  after the V15 content. Root cause: the file was not cleaned up when V15 planning was created.

### Root Causes

- **`sm-kms` shutdown timeout missing**: The `sm-kms` entry point was created by copying `sm-im`
  but the shutdown block was simplified without carrying forward the `context.WithTimeout` pattern.
  Pattern drift between near-identical files is the recurring cause.
- **`close(sigChan)` missing in all 10 entry points**: The canonical pattern was never established
  in a shared location or checked by a fitness linter. Each file evolved independently.
- **`continue-on-error: true` on coverage gates**: These were added as temporary suppressors during
  initial CI setup but never removed after coverage targets were met. Suppressor debt accumulates
  when there's no automated check that they're removed.
- **`pull-requests: write` over-scope**: The workflow-level permission was copied from a template
  that needed PR comments and never scoped down when that feature was removed.

### Patterns for Future Phases

- **Fitness linter for signal handling pattern**: The `close(sigChan)` gap will recur as new
  services are added. Add a `lint-fitness` sub-linter checking that all `signal.Stop(sigChan)` are
  followed by `close(sigChan)` within 3 lines in service entry points.
- **Coverage gate audit**: Any `continue-on-error: true` on a step that ends with `exit 1` is a
  suppressor that MUST have a removal ticket. Add to Phase 12 knowledge propagation.
- **Shutdown pattern enforcement**: The shutdown pattern (`context.WithTimeout` + error log + cancel
  defer) should be extracted to a shared helper and linted for consistency.
- **`const` vs `var` for CLI strings**: Usage strings are pure data. The correct long-term approach
  is `var` initialized by a parameterized builder, but this is a breaking change for existing
  `const` usage. Introduce in V16 with a fitness linter to enforce the pattern going forward.

---

## Phase 1: pki-init Patch — Cat 2, Cat 3, Cat 4, Cat 8, Cat 9 app

*(To be filled during Phase 1 execution using the 4-section structure above)*

---

## Phase 2: OTel Collector Server TLS

*(To be filled during Phase 2 execution using the 4-section structure above)*

---

## Phase 3: App→OTel Client mTLS

*(To be filled during Phase 3 execution using the 4-section structure above)*

---

## Phase 4: Verify OTel Standalone

*(To be filled during Phase 4 execution using the 4-section structure above)*

---

## Phase 5: Grafana LGTM HTTPS + OTLP Ingest TLS

*(To be filled during Phase 5 execution using the 4-section structure above)*

---

## Phase 6: OTel→Grafana Client mTLS

*(To be filled during Phase 6 execution using the 4-section structure above)*

---

## Phase 7: Verify OTel→Grafana Pipeline

*(To be filled during Phase 7 execution using the 4-section structure above)*

---

## Phase 8: Public PS-ID App Server TLS

*(To be filled during Phase 8 execution using the 4-section structure above)*

---

## Phase 9: Deployment Templates

*(To be filled during Phase 9 execution using the 4-section structure above)*

---

## Phase 10: Deployment Linting

*(To be filled during Phase 10 execution using the 4-section structure above)*

---

## Phase 11: Deployment Verification — Full Telemetry Stack

*(To be filled during Phase 11 execution using the 4-section structure above)*

---

## Phase 12: Knowledge Propagation

*(To be filled during Phase 12 execution using the 4-section structure above)*
