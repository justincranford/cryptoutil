# Quiz v2: pki-ca Strategic Direction

**Purpose**: One decision needed to proceed with fixes-v7 Phase 6 (pki-ca work)
**Created**: 2026-02-23

---

## Question 1: Which pki-ca strategic option should fixes-v7 Phase 6 implement?

**Question**: Research is complete (see docs/fixes-v7/research/SUMMARY.md). All 4 options share the same ~30h of prerequisites (jose-ja critical TODOs + sm-kms debt). Which option should be targeted for Phase 6?

**Reference**: [docs/fixes-v7/research/SUMMARY.md](research/SUMMARY.md) for full comparison

**Quick Summary**:
- PKI-CA-MIGRATE (~42-66h total) — Migrate pki-ca in-place after fixing prerequisites. Follows ARCHITECTURE.md migration order. Lowest risk, architecturally correct.
- PKI-CA-MERGE1 (~45h total) — Archive current pki-ca; rebuild from jose-ja base + cherry-pick CA logic. Clean slate but porting risk.
- PKI-CA-MERGE2 (~71h total) — Absorb jose-ja + pki-ca into sm-kms. Anti-pattern (product boundary violation).
- PKI-CA-MERGE3 (~87h total) — Absorb ALL into sm-kms. Strongly not recommended (monolith).

**A)** PKI-CA-MIGRATE — Recommended: In-place migration following ARCHITECTURE.md order
**B)** PKI-CA-MERGE1 — Alternative: Archive + rebuild from jose-ja base
**C)** PKI-CA-MERGE2 — Not recommended: Absorb jose-ja + pki-ca into sm-kms
**D)** PKI-CA-MERGE3 — Strongly not recommended: Full monolith into sm-kms
**E)**

**Answer**: B

**Rationale**: This is a strategic architecture decision that affects team structure, deployment topology, and long-term maintenance. The recommended option (A) follows the defined ARCHITECTURE.md migration order and maintains all product boundaries. Option B is viable if the current pki-ca codebase is deemed too inconsistent for incremental migration. Options C and D violate product boundaries.
