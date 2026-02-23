# Quiz Me v1 — Consolidated Quality Fixes v7

## Question 1: E2E Infrastructure Priority

**Question**: E2E service startup blockers (KMS, JOSE, CA) have persisted across fixes-v1 through fixes-v6 without resolution. How should we handle them in v7?

**A)** Fix all 3 E2E blockers in Phase 6 as planned (KMS session JWK, JOSE args routing, CA flag issue) — full effort ~1.5h
**B)** Focus only on cipher-im E2E (already working) and defer other services until service template migration is complete
**C)** Create a separate E2E-focused plan (docs/e2e-v1/) and remove E2E tasks from fixes-v7 entirely
**D)** Fix only the easiest blocker (CA flag issue) and document the others as architectural debt requiring service template migration
**E)**

**Answer**:

**Rationale**: E2E blockers have been carried forward across 6 prior plans without resolution. Need to decide if they belong in a quality fixes plan or a separate E2E infrastructure plan.

---

## Question 2: crypto/jose Coverage Ceiling

**Question**: `internal/shared/crypto/jose` is at 89.9% with a structural ceiling of ~91%. The ~111 uncovered stmts are unreachable error paths in the jwx v3 library. How should we handle this?

**A)** Accept 89.9% as effective 100% — document structural ceiling, add `//go:cover-ignore` comments, exempt from ≥98% gate
**B)** Push to ~91% by testing additional error paths, accept any remaining as structural ceiling
**C)** Interface-wrap the jwx v3 library to enable mocking (significant refactor, ~8h)
**D)** Accept 89.9% with no further work — it's already well-tested and the gap is purely structural
**E)**

**Answer**:

**Rationale**: Per fixes-v4, the uncovered statements are jwk.Set/Import/json.Marshal errors on valid objects, unreachable type-switch defaults, and jwe/jws errors with valid input. These genuinely cannot be reached without interface-wrapping.

---

## Question 3: nolint:wsl Remediation Approach

**Question**: There are 22 `//nolint:wsl` or `//nolint:wsl_v5` instances. The coding instructions say "NEVER use //nolint:wsl". Some are in identity unified service files (idp, rs, rp, spa, authz) that follow a similar pattern. How thoroughly should these be fixed?

**A)** Fix all 22 instances by restructuring the code — no `//nolint:wsl` exceptions
**B)** Fix the 2 in template/telemetry (easy), document the 20 in identity/unified as needing architectural refactor
**C)** Fix all 22 but allow `//nolint:wsl_v5` (only prohibit `//nolint:wsl` without version suffix)
**D)** Remove `wsl` from linter config entirely — it's too noisy and the violations are cosmetic
**E)**

**Answer**:

**Rationale**: The identity unified service files (idp.go, rs.go, rp.go, spa.go, authz.go) have structural patterns that conflict with wsl's blank line requirements. Fixing them may require significant restructuring of the unified service pattern.
