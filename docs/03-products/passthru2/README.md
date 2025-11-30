# Passthru2: Implementation & Improvement Plan

**Purpose**: Apply lessons learned from `passthru1` to rework demos, solidify best practices, and implement demo parity and developer experience improvements.
**Created**: 2025-11-30
**Status**: DRAFT

---

## Summary of Changes from Passthru1

- Consolidated decision log: use `internal/infra/` + `internal/product/` pattern
- Addressed telemetry coupling and config location inconsistencies in `deployments/` (see `grooming/RESEARCH.md`)
- Created an initial grooming Q&A (25 items) for rapid decision making
- Created new task lists and demo plans aligned with pasthru1 but with immediate DX and E2E improvements

---

## Goals for Passthru2

1. Fix and consolidate differences revealed in `passthru1` (inconsistent docs, missing seeds, telemetry coupling)
2. Ensure product parity: KMS, Identity, JOSE Authority demos have comparable demo experiences (one-command startup, seeded accounts, demo scripts)
3. Implement shared telemetry compose and standardize config locations
4. Implement demo-mode flag and improved `docker compose` patterns for per-product demos
5. Enhance test coverage and CI quality gates

---

## Deliverables

- Updated `README.md`, `TASK-LIST.md`, `DEMO-*.md` per product
- `grooming/GROOMING-QUESTIONS.md` with 25 Q&A items
- `grooming/RESEARCH.md` (already present) updated for `passthru2` insights
- Developer experience concise demo commands and orchestration files

---

## Next steps

1. Answer the 25 grooming questions in `grooming/GROOMING-QUESTIONS.md` to finalize priorities
2. Implement tasks in `TASK-LIST.md` in priority order
3. Start Phase 0: Developer Experience foundation before structural refactors

---

**Status**: WIP
