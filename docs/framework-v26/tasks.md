# Framework v26 — Tasks

**Status**: Parked; no implementation tasks are authorized until Phase 1 defines the reload contract.
**Created**: 2026-10-21

## Phase 1: Define Reload Contract

- [ ] Inventory all active and historical PKI component state and its configuration sources.
- [ ] Specify which settings may reload and which authority/key changes remain restart-only or require
      a distinct rollover workflow.
- [ ] Compare supported trigger choices and document authorization, validation, status reporting,
      rollback, in-flight request, and per-replica semantics.
- [ ] Review the proposed contract against the safety invariants in `plan.md` and record decisions.

## Phase 2: Implement Safe Per-Process Reload

- [ ] Construct and validate a complete candidate PKI component generation before activation.
- [ ] Activate candidate state atomically for new requests while allowing in-flight requests to finish
      on their existing generation.
- [ ] Preserve the last known-good generation and expose an explicit failure result when reload fails.
- [ ] Retain historical authority behavior and avoid logging key or credential material.

## Phase 3: Verify and Document Operations

- [ ] Add table-driven tests for successful reload, invalid configuration, construction failures,
      and rollback to the prior generation.
- [ ] Add concurrency tests proving requests cannot observe mixed generations; run under `-race`.
- [ ] Verify historical certificate verification and revocation behavior after active configuration
      changes.
- [ ] Document the chosen trigger, reloadable fields, replica behavior, recovery, and restart fallback.
- [ ] Run normal/tagged builds, focused tests, linters, and applicable repository/CI quality gates.

## Phase 4: Propagate Lessons

- [ ] Update applicable handbook and PKI/configuration guidance with reusable operational rules.
- [ ] Record phase post-mortems, validation evidence, and the completion summary before archiving.
