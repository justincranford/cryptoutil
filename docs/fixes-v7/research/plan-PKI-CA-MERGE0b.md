# Plan: PKI-CA-MERGE0b

**Option**: Merge sm-im into sm-kms (combined KMS + messaging monolith)
**Recommendation**: ⭐⭐ (Not recommended — domain collision, scale coupling)
**Created**: 2026-02-23

---

## Concept

sm-im's messaging functionality is absorbed into sm-kms. sm-kms becomes a combined service handling both key management operations AND encrypted message storage. All sm-im routes, domain models, and repositories are merged into sm-kms.

---

## Why a User Might Consider This

- Reduces deployment units (9 services → 8 services)
- Both services use GORM, barrier, tenancy — identical infrastructure substrate
- Both protect confidential data — thematic coherence
- Operational simplicity: one service to monitor, one database connection pool, one set of TLS certs

---

## Domain Analysis: sm-kms vs sm-im

| Dimension | sm-kms | sm-im |
|-----------|--------|-----------|
| Primary data | Elastic key rings + material keys | Messages + RecipientJWKs |
| Data retention | Potentially forever (keys in use) | Messages may have TTL/expiry |
| Access pattern | Few records, heavily accessed | Many records, append-heavy |
| Scaling driver | Compute (crypto ops) | Storage (message volume) |
| Audit requirement | Every key operation (security audit) | Message send/read (compliance) |
| Regulatory context | FIPS 140-3, key management | Communication records retention |
| Consumers | Backend services (machine-to-machine) | End users (human-facing) |
| API surface | Keys: CRUD + crypto ops | Messages: send + receive + list |

**Finding**: sm-kms and sm-im serve fundamentally different consumers with fundamentally different data models and scaling requirements. Merging them couples two unrelated concerns.

---

## Proposed Changes

If merged, sm-kms would need:

1. **New GORM models**: `Message`, `MessageRecipientJWK` tables added to sm-kms migrations
2. **New route handlers**: `/messages` endpoints wired into sm-kms server
3. **New OpenAPI spec sections**: merged server spec with message routes
4. **New service logic**: message storage + retrieval logic in sm-kms business logic layer
5. **New repository**: message_repository.go, message_recipient_jwk_repository.go added to sm-kms

---

## Merged sm-kms Structure

```
internal/apps/sm/kms/
├── server/
│   ├── businesslogic/
│   │   ├── businesslogic.go       (existing)
│   │   ├── businesslogic_crypto.go (existing)
│   │   └── businesslogic_messages.go (NEW: sm-im logic)
│   ├── handler/ (existing)
│   ├── middleware/ (existing — 15 files with existing debt)
│   ├── repository/
│   │   ├── orm/ (existing)
│   │   ├── message_repository.go      (NEW from sm-im)
│   │   └── message_recipient_jwk_repository.go (NEW from sm-im)
│   └── apis/
│       ├── messages.go   (NEW from sm-im)
│       └── sessions.go   (NEW from sm-im)
```

---

## Prerequisites (SAME as MIGRATE/MERGE1 for sm-kms debt)

The 15-file custom middleware debt in sm-kms is an EXISTING problem that must be fixed before adding more complexity:
- Remove `application_core.go` / `application_basic.go` wrappers
- Consolidate 15 custom middleware files → template session
- Fix `server.go:35,49` TODOs
- **~19h of prerequisites before any merge work**

---

## Effort Estimate

| Component | Hours |
|-----------|-------|
| sm-kms migration debt (prerequisite) | 19h |
| Merge domain models + migrations | 2h |
| Merge repositories | 2h |
| Merge API handlers + OpenAPI spec | 3h |
| Merge business logic | 2h |
| Testing (unit + integration) | 4h |
| ARCHITECTURE.md + deployment updates | 1h |
| **Total** | **~33h** |

---

## Advantages

- One fewer deployment unit (9 → 8 services)
- Single GORM DB connection for both scenarios (marginally simpler ops)
- No separate container for sm-im (smaller Docker Compose file)

## Disadvantages

- **Domain collision**: Key material management and messaging are unrelated concerns in one service
- **Scale coupling**: KMS is compute-bound (crypto ops); IM is storage-bound (message volume). Cannot scale independently.
- **Blast radius**: KMS failure = messaging failure and vice versa. Two unrelated features, one SPOF.
- **Audit confusion**: KMS audit trail (every crypto op) mixed with message send/receive audit trail
- **Code size**: sm-kms already at 9,536 LOC + 19h migration debt; adding 2,309 LOC of messaging = ~12K LOC service with dual purpose
- **OpenAPI bloat**: KMS API + messaging API in single spec — confusing for clients
- **Prerequisites**: Must fix sm-kms 19h of migration debt BEFORE merge (cannot add more complexity to broken service)
- **More effort than MERGE0a**: 33h vs 4.5h for identical user-facing benefit

---

## Recommendation: ⭐⭐

**NOT RECOMMENDED**. The effort (33h) is 7× that of MERGE0a (4.5h) for dubious benefit (one fewer service). The domain collision forces two fundamentally different concerns — key lifecycle management and encrypted messaging — into a single service with a single failure domain.

**Choose MERGE0a instead**: Move sm-im to SM as sm-im (standalone). Same product grouping benefit, 7× less work, maintains independent scaling and blast radius isolation.

**If the real goal is "reduce service count"**: sm-kms and jose-ja are the natural merge candidates (both manage elastic key rings). See DEEP-RESEARCH.md Schema E for that analysis.

See [tasks-PKI-CA-MERGE0b.md](tasks-PKI-CA-MERGE0b.md) for implementation tasks.
