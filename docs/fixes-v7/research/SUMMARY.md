# Research Summary: pki-ca Strategic Options

**Created**: 2026-02-23
**Purpose**: Compare 4 strategic options for pki-ca service-template migration/integration

---

## Quick Reference Comparison

| Option | Description | Effort | Recommendation |
|--------|-------------|--------|----------------|
| PKI-CA-MIGRATE | Migrate pki-ca in-place to template | 42-66h | ⭐⭐⭐⭐ |
| PKI-CA-MERGE1 | Archive pki-ca; rebuild from jose-ja base | ~45h | ⭐⭐⭐ |
| PKI-CA-MERGE2 | Archive jose-ja + pki-ca; absorb into sm-kms | ~71h | ⭐⭐ |
| PKI-CA-MERGE3 | Archive cipher-im + jose-ja + pki-ca; absorb all into sm-kms | ~87h | ⭐ |

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

### 5. wsl Violations (22 total — Task 2.3 in fixes-v7)
- 2 legacy `//nolint:wsl` at `template/service/telemetry/telemetry_service_helpers.go:134,158` — MUST remove
- 20 `//nolint:wsl_v5` in 5 identity unified files × 4 instances — make genuine effort to fix

---

## Option Deep Dive

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
jose-ja TODOs (11h) ─────┐
                         ├──→ PKI-CA-MIGRATE Phase C (12-36h)
sm-kms debt (19h) ───────┘    PKI-CA-MERGE1 Phases 1-5 (~32h)
                              PKI-CA-MERGE2 Phases 1-5 (~39h)
                              PKI-CA-MERGE3 Phases 1-6 (~52h)

Template startup helper (2h) ─→ All E2E test tasks
```

All options share the same 30h of prerequisite work. The options differ only in what is done with pki-ca after prerequisites are complete.

---

## Files Index

| File | Description |
|------|-------------|
| plan-PKI-CA-MIGRATE.md | RECOMMENDED: In-place migration plan |
| tasks-PKI-CA-MIGRATE.md | RECOMMENDED: 15 tasks (~42-66h) |
| plan-PKI-CA-MERGE1.md | Alternative: Rebuild from jose-ja base |
| tasks-PKI-CA-MERGE1.md | Alternative: 16 tasks (~45h) |
| plan-PKI-CA-MERGE2.md | Not recommended: Absorb into sm-kms |
| tasks-PKI-CA-MERGE2.md | Not recommended: 18 tasks (~71h) |
| plan-PKI-CA-MERGE3.md | Strongly not recommended: Full monolith |
| tasks-PKI-CA-MERGE3.md | Strongly not recommended: ~28 tasks (~87h) |
| SUMMARY.md | This file |
