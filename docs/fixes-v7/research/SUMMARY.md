# Research Summary: pki-ca Strategic Options

**Created**: 2026-02-23
**Purpose**: Compare all strategic options for pki-ca migration and cipher-im placement

---

## Quick Reference Comparison

**cipher-im placement options** (independent of pki-ca option):

| Option | Description | Effort | Recommendation |
|--------|-------------|--------|----------------|
| PKI-CA-MERGE0a | Move cipher-im → sm-im (standalone rename) | ~4.5h | ⭐⭐⭐⭐⭐ |
| PKI-CA-MERGE0b | Merge cipher-im into sm-kms monolith | ~35.5h | ⭐⭐ |

**pki-ca options** (choose one; all require same prerequisites):

| Option | Description | Effort | Recommendation |
|--------|-------------|--------|----------------|
| PKI-CA-MIGRATE | Migrate pki-ca in-place to template | 42-66h | ⭐⭐⭐⭐ |
| PKI-CA-MERGE1 | Archive pki-ca; rebuild from jose-ja base | ~45h | ⭐⭐⭐ |
| PKI-CA-MERGE2 | Archive jose-ja + pki-ca; absorb into sm-kms | ~71h | ⭐⭐ |
| PKI-CA-MERGE3 | Archive cipher-im + jose-ja + pki-ca; absorb all into sm-kms | ~87h | ⭐ |

---

## Deep Research

For exhaustive analysis of all possible product groupings, adjacent/missing services, and industry comparisons:

**See**: [DEEP-RESEARCH.md](DEEP-RESEARCH.md)

Key findings from deep research:
- **8 Product Taxonomy Schemas** (A–H) evaluated with ratings
- **14 Adjacent/Missing Services** identified (sm-secrets is highest priority gap)
- **jose-ja / sm-kms overlap**: both implement elastic key ring — intentional duplication or consolidation target?
- **Recommended target architecture**: 4 products / 11 services (SM: kms+jwk+im+secrets; PKI: ca+ocsp; Identity ×5; Audit: server)

---

## Critical Cross-Cutting Findings (ALL Options)

These issues exist **REGARDLESS** of which pki-ca option is chosen:

### 1. jose-ja Critical TODOs (BLOCKING for all options)
- `jwk_handler.go:358,368`: JWK generation stubs — NOT implemented
- `jwk_handler_material.go:234,244,254,264`: sign/verify/encrypt/decrypt — NOT implemented
- **Estimated fix**: 11h
- **Impact**: jose-ja is structurally migrated but NOT functionally complete

### 2. sm-kms Migration Debt (significant for all options)
- `server/application/application_core.go` + `application_basic.go`: Old pre-builder wrappers
- `server/middleware/`: 15 non-test files of custom JWT/claims/session middleware (vs template's 1 file)
- `server.go:35,49`: TODOs for SQLRepository→GORM and GORM+barrier integration
- No integration tests, no E2E tests
- **Estimated fix**: 19h
- **Note**: sm-kms is LAST per ARCHITECTURE.md migration order, but debt still must be cleaned up

### 3. Template Testing Infrastructure Gap
- No generic `StartServiceFromConfig()` helper exists in template
- cipher-im uses raw 50×100ms polling loop (`testing/testmain_helper.go`)
- jose-ja uses same raw polling pattern
- Template has `WaitForServerPort()` in e2e_helpers but it's not wired generically
- **Estimated fix**: 2h (Task 6.0 in main fixes-v7)

### 4. ci-e2e.yml Path Bug
- References `deployments/jose/compose.yml` (should be `deployments/jose-ja/compose.yml`)
- All non-cipher-im E2E tests have `SERVICE_TEMPLATE_TODO` comments (disabled)

### 5. cipher-im Product Problem (cipher product has only 1 service)
- 1-service product `Cipher` is organizational overhead
- cipher-im is a MESSAGE STORE, not a cipher library — belongs in SM product
- **Options**: MERGE0a (recommended, 4.5h pure rename) or MERGE0b (not recommended, ~35h merge)
- **Independent of pki-ca option** — can be done at any time

### 6. wsl Violations (22 total — Task 2.3 in fixes-v7)
- 2 legacy `//nolint:wsl` at `template/service/telemetry/telemetry_service_helpers.go:134,158` — MUST remove
- 20 `//nolint:wsl_v5` in 5 identity unified files × 4 instances — make genuine effort to fix

---

## Option Deep Dive

### PKI-CA-MERGE0a ⭐⭐⭐⭐⭐ (STRONGLY RECOMMENDED — do first)

**Approach**: Pure mechanical rename of cipher-im → sm-im. No business logic changes. Copy files, update imports, update deployments/configs/ARCHITECTURE.md.
**Who it's for**: Anyone who wants to move cipher-im to the SM product with minimal risk and zero prerequisites.
**Why strongly recommended**:
- Zero prerequisites (can be done before any other migration work)
- ~4.5h total effort
- Eliminates the 1-service Cipher product
- Paves way for future SM extensions (sm-secrets, sm-ssh, sm-file)
- Pure rename — no logic changes possible means no logic bugs introduced

**See**: [plan-PKI-CA-MERGE0a.md](plan-PKI-CA-MERGE0a.md) + [tasks-PKI-CA-MERGE0a.md](tasks-PKI-CA-MERGE0a.md)

---

### PKI-CA-MERGE0b ⭐⭐ (NOT RECOMMENDED)

**Approach**: Remove cipher-im as a separate service; absorb all message handling, repositories, and API routes directly into sm-kms.
**Why not recommended**:
- Requires sm-kms migration debt (19h) as prerequisite BEFORE any merge work
- Different domain concerns: sm-kms is compute-bound machine-to-machine; cipher-im is storage-bound human-facing
- ~8× the effort of MERGE0a for the same product grouping outcome
- Degrades independent scalability

**See**: [plan-PKI-CA-MERGE0b.md](plan-PKI-CA-MERGE0b.md) + [tasks-PKI-CA-MERGE0b.md](tasks-PKI-CA-MERGE0b.md)

---

### PKI-CA-MIGRATE ⭐⭐⭐⭐ (RECOMMENDED)

**Approach**: Fix pre-existing gaps in cipher-im/jose-ja/sm-kms first, then migrate pki-ca in-place following the established ARCHITECTURE.md migration order.

**Who it's for**: Teams that want to follow the defined migration strategy, minimize architectural risk, and keep service boundaries clean.

**Key phases**:
- Phase A: sm-kms debt cleanup (19h)
- Phase B: jose-ja critical TODOs (11h)
- Phase C: pki-ca in-place migration (12-36h):
  - GORM certificate storage (replaces MemoryStore)
  - Fix SetReady(true) placement (ca.go:caServerStart)
  - Consolidate magic/ to shared/magic
  - Add integration + E2E tests
  - Enable in ci-e2e.yml

**Why recommended**:
- Architecturally correct per ARCHITECTURE.md migration order
- Maintains all product boundaries (SM, PKI, JOSE, Cipher, Identity remain separate)
- Low risk: incremental changes on existing codebase
- Same prerequisites as MERGE1 but less total effort

**See**: [plan-PKI-CA-MIGRATE.md](plan-PKI-CA-MIGRATE.md) + [tasks-PKI-CA-MIGRATE.md](tasks-PKI-CA-MIGRATE.md)

---

### PKI-CA-MERGE1 ⭐⭐⭐ (ALTERNATIVE)

**Approach**: Archive current pki-ca (proven CA logic preserved); build new pki-ca from jose-ja base (clean template skeleton) + cherry-pick CA business logic.

**Who it's for**: Teams that believe current pki-ca's structure is so inconsistent that a clean rebuild is safer than incremental migration.

**Key advantage over MIGRATE**: Guarantees clean template architecture from day 1 (no remnant pre-template patterns).

**Key risk**: Careful porting required — must preserve all proven CA logic from archived version. jose-ja must be complete before starting (Pre.1+Pre.2 = 11h blocking).

**Why not top choice**: Same prerequisites as MIGRATE (~11h jose-ja TODOs), slightly higher total effort (~45h vs 42-66h), higher risk of introducing bugs during porting.

**See**: [plan-PKI-CA-MERGE1.md](plan-PKI-CA-MERGE1.md) + [tasks-PKI-CA-MERGE1.md](tasks-PKI-CA-MERGE1.md)

---

### PKI-CA-MERGE2 ⭐⭐ (NOT RECOMMENDED)

**Approach**: Archive jose-ja + pki-ca; absorb both into sm-kms as "Crypto Operations Service".

**Why not recommended**: Violates product boundary between SM (Secret Management), JOSE (JWK Authority), and PKI (Certificate Authority). Creates ~25K LOC service. ARCHITECTURE.md explicitly defines these as separate products. Independent scaling impossible. Highest effort of all useful options (~71h).

**Only viable if**: Organization explicitly decides to collapse all crypto operations into single deployment unit and dissolves product boundaries.

**See**: [plan-PKI-CA-MERGE2.md](plan-PKI-CA-MERGE2.md) + [tasks-PKI-CA-MERGE2.md](tasks-PKI-CA-MERGE2.md)

---

### PKI-CA-MERGE3 ⭐ (STRONGLY NOT RECOMMENDED)

**Approach**: Archive cipher-im + jose-ja + pki-ca; absorb ALL into sm-kms as crypto monolith.

**Why strongly not recommended**: Collapses 4 separate products (SM, Cipher, JOSE, PKI) into 1 service. Creates ~27K LOC monolith. Violates CA/Browser Forum requirements for CA isolation. Worst architectural outcome. Highest effort (~87h). Exists only for completeness of option space analysis.

**See**: [plan-PKI-CA-MERGE3.md](plan-PKI-CA-MERGE3.md) + [tasks-PKI-CA-MERGE3.md](tasks-PKI-CA-MERGE3.md)

---

## Decision Guidance

**For cipher-im placement — choose PKI-CA-MERGE0a** unless there is a specific operational reason to merge into sm-kms (almost none exist).

**Choose PKI-CA-MIGRATE if**:
- You want to follow ARCHITECTURE.md migration order
- You want minimal architectural risk
- You prefer incremental improvements over clean-slate rebuilds

**Choose PKI-CA-MERGE1 if**:
- You believe current pki-ca structure is too inconsistent for incremental improvement
- You've already completed jose-ja critical TODOs
- You're willing to accept porting risk for guaranteed clean architecture

**Do NOT choose PKI-CA-MERGE2 or MERGE3** unless the organization has made an explicit strategic decision to dissolve product boundaries.

---

## Prerequisite Dependency Map

```
PKI-CA-MERGE0a (4.5h) ─────────────────────────→ DONE (no prereqs)

PKI-CA-MERGE0b (35.5h):
  sm-kms debt (19h) ──→ merge work (14h)

jose-ja TODOs (11h) ─────┐
                         ├──→ PKI-CA-MIGRATE Phase C (12-36h)
sm-kms debt (19h) ───────┘    PKI-CA-MERGE1 Phases 1-5 (~32h)
                              PKI-CA-MERGE2 Phases 1-5 (~39h)
                              PKI-CA-MERGE3 Phases 1-6 (~52h)

Template startup helper (2h) ─→ All E2E test tasks
```

All pki-ca options share the same 30h of prerequisite work. MERGE0a is completely independent of all other work.

---

## Files Index

| File | Description |
|------|-------------|
| DEEP-RESEARCH.md | Deep research: 8 taxonomy schemas, 14 missing services, industry comparisons |
| plan-PKI-CA-MERGE0a.md | STRONGLY RECOMMENDED: cipher-im → sm-im standalone rename |
| tasks-PKI-CA-MERGE0a.md | STRONGLY RECOMMENDED: 17 tasks (~4.5h, no prereqs) |
| plan-PKI-CA-MERGE0b.md | Not recommended: cipher-im merged into sm-kms |
| tasks-PKI-CA-MERGE0b.md | Not recommended: 16 tasks (~35.5h) |
| plan-PKI-CA-MIGRATE.md | RECOMMENDED for pki-ca: In-place migration plan |
| tasks-PKI-CA-MIGRATE.md | RECOMMENDED for pki-ca: 15 tasks (~42-66h) |
| plan-PKI-CA-MERGE1.md | Alternative for pki-ca: Rebuild from jose-ja base |
| tasks-PKI-CA-MERGE1.md | Alternative for pki-ca: 16 tasks (~45h) |
| plan-PKI-CA-MERGE2.md | Not recommended: Absorb jose-ja+pki-ca into sm-kms |
| tasks-PKI-CA-MERGE2.md | Not recommended: 18 tasks (~71h) |
| plan-PKI-CA-MERGE3.md | Strongly not recommended: Full monolith |
| tasks-PKI-CA-MERGE3.md | Strongly not recommended: ~28 tasks (~87h) |
| SUMMARY.md | This file |
