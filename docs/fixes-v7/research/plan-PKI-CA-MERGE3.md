# Plan: PKI-CA-MERGE3

**Option**: Archive sm-im + jose-ja + pki-ca; absorb all three into sm-kms as "Crypto Monolith"
**Recommendation**: ⭐ (Strongly not recommended — architectural anti-pattern)
**Created**: 2026-02-23

---

## Concept

sm-kms absorbs ALL other cryptoutil services:
- sm-im's messaging encryption (2,309 LOC)
- jose-ja's JWK operations (4,406 LOC)
- pki-ca's CA operations (11,418 LOC)

Result: a single "Cryptoutil Monolith" at ~27K+ LOC, serving all cryptographic operations.

---

## Current State

Same prerequisites as MERGE2, PLUS:

### sm-im Migration Debt (minimal — reference impl)
| Gap | Current | Target | Effort |
|-----|---------|--------|--------|
| E2E timeout reliability | Occasional | Reliable | 1.5h |
| TestMain uses raw polling | Custom 50×100ms | template WaitForServerPort | 1h |

---

## Proposed Architecture

```
sm-kms (renamed or as-is) becomes:
internal/apps/sm/kms/
├── server/
│   ├── api/
│   │   ├── handler/     (existing KMS handlers)
│   │   ├── sm/      (from sm-im: messaging)
│   │   ├── jwk/         (from jose-ja: JWK)
│   │   └── ca/          (from pki-ca: CA)
│   ├── service/
│   │   ├── kms/         (existing)
│   │   ├── sm/      (from sm-im)
│   │   ├── jwk/         (from jose-ja)
│   │   └── ca/          (from pki-ca)
│   └── repository/
│       ├── key_repository.go     (existing)
│       ├── message_repository.go (from sm-im)
│       ├── jwk_repository.go     (from jose-ja)
│       └── cert_repository.go    (from pki-ca: GORM)

archived/
├── sm-im/
├── jose-ja/
└── pki-ca/
```

---

## Product Boundary Analysis

ARCHITECTURE.md defines 5 distinct products:
| Product | Service | Purpose |
|---------|---------|---------|
| SM | sm-kms | Secret/Key Management |
| Cipher | sm-im | Encrypted Messaging |
| JOSE | jose-ja | JWK Authority |
| PKI | pki-ca | Certificate Authority |
| Identity | (multiple) | AuthN/AuthZ |

This option **COLLAPSES 4 separate products into 1**, destroying the product boundary design.

---

## OpenAPI Implications

Must merge 4 separate OpenAPI specs into one:
- /service/api/v1/keys (KMS)
- /service/api/v1/messages (sm-im)
- /service/api/v1/jwks (jose-ja)
- /service/api/v1/certs (pki-ca)

The merged spec would be extremely large and hard to maintain.

---

## Advantages

- Absolute minimum deployment units (4 services → 1, excluding Identity)
- Single container to manage
- One set of TLS certs, one healthcheck, one CI pipeline

## Disadvantages

- **DESTROYS ALL PRODUCT BOUNDARIES**: SM ≠ JOSE ≠ PKI
- Creates a ~27K LOC monolith — impossible to partition, maintain, or reason about
- Single point of failure for ALL non-identity crypto operations
- Each domain has radically different scaling characteristics:
  - sm-im: messaging throughput (horizontally scalable)
  - jose-ja: JWK operations (read-heavy, low write)
  - pki-ca: cert issuance (write-heavy, compliance requirements)
  - sm-kms: key operations (security-critical, audit-heavy)
- Impossible to independently audit/compliance-check CA operations
- CA/Browser Forum compliance requires isolated CA infrastructure
- Team ownership undefined — single service owned by everyone = owned by no-one
- All cryptoutil clients must be updated to new combined service endpoints
- Breaking change with zero benefit over PKI-CA-MIGRATE

---

## Effort Estimate

| Component | Hours |
|-----------|-------|
| All MERGE2 prerequisites and work | ~71h |
| Port sm-im (additional) | 8h |
| Merge sm-im OpenAPI into combined spec | 2h |
| sm-im unit/integration/E2E tests | 4h |
| **Total** | **~85h** |

Most expensive option. Highest risk. Worst architectural outcome.

---

## Recommendation: ⭐

**STRONGLY NOT RECOMMENDED**. This option:

1. Violates ALL product boundary separations defined in ARCHITECTURE.md
2. Creates an unmaintainable ~27K LOC monolith
3. Violates CA/Browser Forum requirements for CA isolation
4. Has the highest effort of all 4 options (~85h)
5. Provides essentially zero technical benefit over PKI-CA-MIGRATE
6. Breaks all existing client service endpoints

**This option exists only for completeness of the option space analysis.** It should not be implemented under any circumstances without complete architectural redesign of the entire product suite and explicit organizational decision to abandon product separation.

Compare with PKI-CA-MIGRATE (⭐⭐⭐⭐): architecturally correct, lowest risk, maintains all product boundaries.
