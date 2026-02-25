# Plan: PKI-CA-MERGE2

**Option**: Archive jose-ja + pki-ca; absorb both into sm-kms as "Crypto Operations Service"
**Recommendation**: ⭐⭐ (Not recommended — product boundary violation)
**Created**: 2026-02-23

---

## Concept

sm-kms (already the most complete service at 9,536 LOC) absorbs:
- jose-ja's JWK operations (4,406 LOC)
- pki-ca's CA operations (11,418 LOC)

Result: a single "Crypto Operations Service" combining KMS + JWK + CA as ~25K LOC service.

---

## Current State

### sm-kms Migration Debt (must fix regardless)
| Gap | Current | Target | Effort |
|-----|---------|--------|--------|
| application_core wrappers | Old pattern | Direct ServiceResources | 2h |
| 15 custom middleware files | Custom JWT/claims | Template session.go | 4h |
| SQLRepository in server.go:35 | Custom SQL | GORM via template | 6h |
| No integration tests | None | TestMain pattern | 4h |
| No E2E tests | None | Template e2e_infra | 3h |

### jose-ja Critical TODOs (must fix regardless)
| Gap | Location | Effort |
|-----|----------|--------|
| JWK generation | jwk_handler.go:358,368 | 3h |
| sign/verify/encrypt/decrypt | jwk_handler_material.go:234-264 | 8h |

### pki-ca Gaps
| Gap | Current | Target | Effort |
|-----|---------|--------|--------|
| Storage | MemoryStore only | GORM | 4h |
| SetReady(true) location | Before Start() | After binding | 30min |
| Local magic/ package | internal/apps/pki/ca/magic/ | shared/magic | 1h |
| No integration tests | None | TestMain pattern | 3h |
| No E2E tests | None | Template e2e_infra | 3h |

---

## Proposed Architecture

```
sm-kms → rename to crypto-ops (or keep sm-kms for compat)
internal/apps/sm/kms/
├── server/
│   ├── api/
│   │   ├── handler/ (existing KMS handlers)
│   │   ├── jwk/    (from jose-ja: JWK handlers)
│   │   └── ca/     (from pki-ca: CA handlers)
│   ├── middleware/ (cleaned up, using template session)
│   ├── repository/
│   │   ├── key_repository.go   (existing)
│   │   ├── jwk_repository.go   (from jose-ja)
│   │   └── cert_repository.go  (from pki-ca: GORM)
│   └── service/
│       ├── kms/  (existing)
│       ├── jwk/  (from jose-ja)
│       └── ca/   (from pki-ca)

archived/
├── jose-ja/ (after merge)
└── pki-ca/  (after merge)
```

---

## What Gets Merged

| Source | Component | Action |
|--------|-----------|--------|
| jose-ja | api/handler/ | → sm-kms/server/api/jwk/ |
| jose-ja | service/ | → sm-kms/server/service/jwk/ |
| jose-ja | openapi/ | → sm-kms/server/openapi/ (merged spec) |
| pki-ca | api/handler/ | → sm-kms/server/api/ca/ |
| pki-ca | service/issuer, revocation, timestamp | → sm-kms/server/service/ca/ |
| pki-ca | compliance/, crypto/, profile/ | → sm-kms/server/ca/ (new subdir) |
| pki-ca | storage/ | → sm-kms/server/repository/cert_repository.go (GORM) |

---

## OpenAPI Implications

- sm-kms, jose-ja, pki-ca each have separate OpenAPI specs
- Merged service needs single unified spec with 3 API sections:
  - /service/api/v1/keys (KMS)
  - /service/api/v1/jwks (JWK)
  - /service/api/v1/certs (CA)
- oapi-codegen generates 3 sets of server stubs from single spec → complex

---

## Prerequisites (ALL must be done before merge)

1. ✅ sm-kms migration debt (all items above)
2. ✅ jose-ja critical TODOs (sign/verify/encrypt/decrypt)
3. ✅ template generic startup helper (Task 6.0)

---

## Advantages

- Fewest deployment units (3 services → 1)
- Fewer container resources in production
- Single service to monitor and maintain
- Cryptographic operations colocated (potential for cross-operation optimization)

## Disadvantages

- **VIOLATES PRODUCT BOUNDARIES**: SM (Secret Management) ≠ JOSE (JWK) ≠ PKI (CA)
- Creates ~25K LOC service (3× sm-im, 2.6× jose-ja)
- Single service becomes single point of failure for ALL crypto operations
- Merged OpenAPI spec is complex and harder to maintain
- Independent scaling impossible (KMS and CA have very different load profiles)
- ARCHITECTURE.md explicitly defines separate products: SM, JOSE, PKI
- Client teams expect separate service endpoints (breaking API change)
- CI/CD test isolation lost (one service → all tests must pass for any change)
- Team ownership unclear (who owns crypto-ops?)

---

## Effort Estimate

| Component | Hours |
|-----------|-------|
| sm-kms cleanup (prerequisite) | 19h |
| jose-ja TODOs (prerequisite) | 11h |
| pki-ca GORM storage (done during merge) | 4h |
| API merger (handler ports, routing) | 8h |
| OpenAPI spec merger | 4h |
| Service integration and wiring | 4h |
| Testing (unit, integration, E2E) | 8h |
| CLI routing updates | 2h |
| Archive jose-ja + pki-ca | 1h |
| **Total** | **~61h** |

---

## Recommendation: ⭐⭐

**NOT RECOMMENDED**. This option creates an architectural anti-pattern:
1. Violates the established product boundary separation in ARCHITECTURE.md
2. Creates a ~25K LOC service that's harder to test, maintain, and deploy independently
3. All the prerequisites (sm-kms cleanup, jose-ja TODOs) must be done anyway — net savings are minimal (only avoids separate deployments)
4. Breaking API change for any existing clients expecting separate service endpoints

**Only viable if**: Organization decision to collapse all crypto operations into a single deployment unit for operational simplicity, and product boundaries are intentionally dissolved.

Compare with PKI-CA-MIGRATE (⭐⭐⭐⭐): same prerequisites, architecturally correct, maintains product boundaries.
