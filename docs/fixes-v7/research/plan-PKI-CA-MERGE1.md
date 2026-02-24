# Research Option: PKI-CA-MERGE1

**Option**: Archive pki-ca; build new pki-ca from jose-ja base + CA logic
**Status**: Research Only (not yet selected)
**Created**: 2026-02-23
**Related**: docs/fixes-v7/research/tasks-PKI-CA-MERGE1.md

---

## Overview

Archive the current `internal/apps/pki/ca/` (move to `internal/apps/archived/pki-ca/`).
Build a new `internal/apps/pki/ca/` from scratch using `jose-ja` as the structural
template, merging only the proven, unique CA business logic from the archived version.

This "greenfield with cherry-picking" approach avoids carrying forward ALL of pki-ca's
current gaps (missing GORM, missing E2E, missing integration tests, local magic, etc.)
and instead starts from a clean, already-template-compliant base (jose-ja).

---

## Rationale

The current pki-ca has significant technical debt:
- In-memory only storage (no GORM, no persistence)
- Local magic package (not in shared/magic)
- No integration test suite
- No E2E test suite
- `SetReady(true)` in wrong place
- TestMain uses raw polling (no template helper)
- README documents Tasks 11-20 as incomplete

Rather than incrementally fixing all these gaps (PKI-CA-MIGRATE approach), this option
rebuilds pki-ca FROM a clean already-compliant base (jose-ja) and cherry-picks only the
verified CA business logic.

---

## jose-ja as Base Structure (Why This Works)

jose-ja and pki-ca both:
- Are cryptographic operation services (jose-ja: JWK operations, pki-ca: certificate operations)
- Use the service-template builder pattern
- Have handler, service, config layers
- Need GORM storage for cryptographic artifacts (jose-ja: JWK sets, pki-ca: certificates)

jose-ja structure that transfers directly to new pki-ca:
- `server/server.go`: Builder pattern (clean template usage)
- `server/config/config.go` + `server/config/config_test_helper.go`: Config test helpers
- `server/testmain_test.go`: TestMain pattern (needs updating to template helper)
- `service/audit_log_service.go`: Audit log pattern applies to CA audit logs

---

## Critical Pre-condition

jose-ja's JWK generation and sign/verify/encrypt/decrypt TODOs MUST be implemented
before using jose-ja as a base. Building new pki-ca from broken jose-ja would import
the same bugs.

See: PKI-CA-MIGRATE Tasks B.1, B.2 — these are REQUIRED prerequisites for this option too.

---

## What to Cherry-Pick from Current pki-ca

### Keep (port directly to new pki-ca)
- `compliance/`: CA/Browser Forum baseline checks — standalone, no infrastructure deps
- `crypto/`: CA crypto provider interface — standalone
- `intermediate/`: Intermediate CA management — standalone if storage layer replaced
- `profile/certificate/` + `profile/subject/`: Profile configuration — standalone
- `security/`: Security policy enforcement — standalone
- `service/issuer/`: EST certificate issuance — needs storage adapter
- `service/revocation/`: CRL + OCSP — needs storage adapter
- `service/timestamp/`: RFC 3161 TSA — standalone
- `api/handler/handler.go`, `handler_est.go`, `handler_certs.go`, `handler_ocsp.go`: API handlers — needs storage adapter
- `cli/cli.go`: Certificate generation CLI — standalone (separate from server)
- `bootstrap/`: CA bootstrap — needs checking

### Replace or Rewrite
- `storage/MemoryStore` → GORM `CertificateRepository` (same interface, GORM impl)
- `magic/` → `internal/shared/magic/magic_pki.go`
- Integration tests → new, using template TestMain helper
- E2E tests → new, using template e2e_infra
- `server/server.go` → new, based on jose-ja's server.go structure

### Discard
- Old polling loop in raw TestMain (replaced by template helper)
- `server/cmd/cmd.go` (clarify vs cmd/pki-ca/main.go)
- Any patterns that conflict with template

---

## Service-Template Consistency Gaps to Fix First

Same as PKI-CA-MIGRATE — before using jose-ja as base, jose-ja must be complete:

### jose-ja Gaps (BLOCKING prerequisites)
1. ❌ **CRITICAL**: JWK generation, sign, verify, encrypt, decrypt stubs unimplemented
2. ❌ TestMain uses legacy polling (not template helper)
3. ❌ No testing/ helper package
4. ❌ No e2e/ directory

### cipher-im Gaps (template improvements, needed for consistency)
1. ❌ `StartCipherIMService()` should use template generic helper
2. ❌ Template lacks generic `StartServiceFromConfig()` helper

### sm-kms Gaps (separate concern, but important for ecosystem health)
1. ❌ Old application_core wrappers
2. ❌ Custom middleware duplication
3. ❌ Missing E2E tests

---

## New pki-ca Structure (After Merge)

```
internal/apps/pki/ca/
├── ca.go                          # RouteService entry (from current ca.go)
├── server/
│   ├── server.go                  # NewFromConfig (jose-ja structure + CA logic)
│   ├── config/
│   │   ├── config.go              # CAServerSettings embedded ServiceTemplateServerSettings
│   │   └── config_test_helper.go  # NewTestConfig(), DefaultTestConfig()
│   ├── apis/                      # handler.go, handler_est.go, handler_certs.go, handler_ocsp.go
│   └── repository/
│       ├── migrations/
│       │   ├── 2001_certificates.up.sql
│       │   └── 2001_certificates.down.sql
│       ├── migrations.go
│       └── certificate_repository.go   # New GORM-backed Store implementation
├── testing/
│   └── testmain_helper.go         # StartCAServer(), SetupTestServer()
├── e2e/
│   ├── testmain_e2e_test.go
│   └── e2e_test.go
├── compliance/                    # Cherry-picked unchanged
├── crypto/                        # Cherry-picked unchanged
├── intermediate/                  # Cherry-picked, adapted for new storage
├── profile/                       # Cherry-picked unchanged
├── security/                      # Cherry-picked unchanged
├── service/issuer/                # Cherry-picked, adapted for GORM storage
├── service/revocation/            # Cherry-picked, adapted for GORM storage
├── service/timestamp/             # Cherry-picked unchanged
├── cli/                           # Cherry-picked unchanged
└── bootstrap/                     # Cherry-picked and verified
```

---

## Advantages of This Option

- ✅ Starts from clean slate — no imported pki-ca technical debt
- ✅ jose-ja structure already validated and template-compliant
- ✅ Forces proper GORM storage (current in-memory storage was never acceptable for production)
- ✅ Forces complete E2E and integration tests (no legacy gaps carry forward)
- ✅ Clean magic constants from day one
- ✅ Correct SetReady pattern from day one
- ✅ Smaller migration surface than full pki-ca migration (cherry-pick proven components)

## Disadvantages

- ❌ Higher risk of breaking proven CA logic during porting
- ❌ Significant upfront work on jose-ja prerequisites
- ❌ Storage replacement requires careful interface preservation
- ❌ "Archive" step is irreversible (though archived code stays in repo)
- ❌ EST handler porting may have subtle bugs if context is lost
- ❌ Parallel development risk: if someone is actively working on pki-ca when it gets archived

---

## Effort Estimate

| Work Item | Estimated Effort |
|-----------|-----------------|
| jose-ja prerequisites (same as PKI-CA-MIGRATE Phase B) | 14h |
| template testing extraction | 2h |
| Archive current pki-ca | 30min |
| New pki-ca skeleton from jose-ja | 3h |
| GORM certificate repository | 4h |
| Port unique CA components (compliance, crypto, service/*) | 6h |
| Port API handlers with storage adapter | 6h |
| testing/ helper package | 2h |
| Integration test suite | 3h |
| E2E test suite | 3h |
| Magic consolidation | 1h |
| **Total** | **44.5h** |

---

## Effort vs PKI-CA-MIGRATE

PKI-CA-MIGRATE: 45h (including sm-kms cleanup 15h + jose-ja TODOs 14h + migration 16h)
PKI-CA-MERGE1: 44.5h (jose-ja prerequisites 14h + greenfield rebuild 30.5h)

Similar total effort. The difference is risk profile:
- MIGRATE: Lower risk on CA logic (less porting), higher risk on sm-kms cleanup
- MERGE1: Higher risk on CA logic porting, but cleaner end state

---

## Recommendation Score: ⭐⭐⭐ (Moderate Recommend)

Best choice if clean architecture is the priority and if someone has deep knowledge of
all pki-ca's existing API handler tests (25 Go test files). The risk of subtle regression
during porting can be mitigated by keeping both versions simultaneously for a period
(current pki-ca → archived/pki-ca, new in pki/ca).

The prerequisite of fixing jose-ja's critical TODOs first is non-negotiable
and adds front-loaded risk.
