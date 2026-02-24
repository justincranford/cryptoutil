# Plan: PKI-CA-MERGE0a

**Option**: Move cipher-im to SM product as sm-im (standalone service)
**Recommendation**: ⭐⭐⭐⭐⭐ (Strongly recommended — minimal effort, maximum coherence)
**Created**: 2026-02-23

---

## Concept

cipher-im is re-homed from the "Cipher" product to the "SM" (Secure Materials/Secrets Manager) product as a new standalone service `sm-im`. The service code, API, and functionality are **unchanged** — only the product label and file paths change.

Simultaneously (optional but recommended): rename `jose-ja` → `sm-jwk` and move to SM product, giving SM three cohesive services: sm-kms (keys), sm-im (messages), sm-jwk (JWK authority).

---

## Rationale

**Why cipher-im doesn't fit in "Cipher"**:
- "Cipher" as a product name suggests a protocol, not a use case
- Cipher product has only ONE service (cipher-im) — a 1-service product adds organizational overhead with no benefit
- cipher-im is a MESSAGE store, not a cipher library

**Why cipher-im fits in "SM"**:
- SM = "Secure Materials" — manages anything that must remain confidential
- sm-kms protects KEY MATERIAL; sm-im protects MESSAGE CONTENT — both are "secure materials"
- Same tenant isolation patterns, same barrier encryption substrate
- Natural product family: sm-kms (keys for services), sm-im (messages for users)
- Industry precedent: AWS KMS + end-to-end encrypted services under same security brand

**Why NOT merge into sm-kms**:
- Different domain models (key rings vs messages/recipients)
- Different scaling characteristics (KMS = compute-heavy; IM = storage-heavy)  
- Blast radius: IM outage would take down KMS if merged
- Independent deployability is valuable

---

## Scope of Changes

This is a **mechanical rename** — no business logic changes.

| Change Type | Items |
|------------|-------|
| Directory move | `internal/apps/cipher/im/` → `internal/apps/sm/im/` |
| Import path update | All `cryptoutil/internal/apps/cipher/im/...` → `cryptoutil/internal/apps/sm/im/...` |
| CMD rename | `cmd/cipher-im/` → `cmd/sm-im/` |
| CMD product | `cmd/cipher/main.go` → update to remove im; `cmd/sm/main.go` → add im |
| Deployment | `deployments/cipher-im/` → `deployments/sm-im/` |
| Deployment | `deployments/cipher/` → `deployments/sm/` (add im) |
| Config | `configs/cipher/im/` → `configs/sm/im/` |
| Port assignment | 8700-8799 stays with sm-im (no port change needed) |
| ARCHITECTURE.md | Cipher product: remove im, mark empty or dissolve; SM: add sm-im |
| ci-e2e.yml | Update path reference from cipher-im to sm-im |
| go.sum/go.mod | No changes needed (same module) |

**Optional bundle**: Rename jose-ja → sm-jwk simultaneously (adds ~1h):
| Directory move | `internal/apps/jose/ja/` → `internal/apps/sm/jwk/` |
| CMD rename | `cmd/jose-ja/` → `cmd/sm-jwk/` |
| Deployment | `deployments/jose-ja/` → `deployments/sm-jwk/` |
| ARCHITECTURE.md | JOSE product dissolved; SM gains sm-jwk |

---

## Current State Dependencies

No other service imports from `internal/apps/cipher/im/` (cipher-im is self-contained). Confirmed by:
```bash
grep -r "apps/cipher/im" internal/ --include="*.go" | grep -v "^internal/apps/cipher/im/"
# Expected: empty (no cross-service imports)
```

The rename is purely internal — no client contracts change (OpenAPI paths, TLS certs, Docker DNS names must be updated in deployment configs but service behavior is identical).

---

## Effort Estimate

| Phase | Description | Hours |
|-------|-------------|-------|
| Directory move + import updates | sed/find-replace + go build | 1h |
| cmd/ rename | Update main.go files | 30min |
| Deployment + config file updates | compose.yml, config-*.yml paths | 30min |
| ARCHITECTURE.md update | Port table, product catalog, directory trees | 1h |
| ci-e2e.yml fix | Update paths and service names | 30min |
| Testing (build + E2E verification) | Ensure nothing broken | 1h |
| | **sm-im standalone total** | **~4.5h** |
| *(Optional)* jose-ja → sm-jwk rename | Same steps for jose-ja | +1h |
| | **sm-im + sm-jwk total** | **~5.5h** |

---

## Advantages

- Eliminates the 1-service "Cipher" product (reduces product sprawl)
- SM becomes a coherent family: sm-kms (keys) + sm-im (messages) [+ sm-jwk if done together]
- Zero business logic change — no regression risk
- Minimal testing required (rename-only, same tests pass)
- Sets up SM product for future additions: sm-secrets (static KV), sm-ssh (SSH CA), sm-file (encrypted files)
- Port assignments unchanged (no Docker network config changes)

## Disadvantages

- Breaking change for any external clients using `cipher-im` as a Docker DNS name — must update to `sm-im`
- ARCHITECTURE.md requires significant table updates (port catalog, product catalog, directory tree)
- ci-e2e.yml currently references `cipher-im` service name
- If cipher-im → sm-im AND jose-ja → sm-jwk done together: larger change but better product taxonomy

---

## Risk Assessment

| Risk | Probability | Mitigation |
|------|------------|------------|
| Import path missed in some file | Low | `go build ./...` catches all misses |
| Docker DNS name breaking change | Low | Only affects local/test compose files; update all at once |
| Test failure due to path | Low | E2E tests reference service name; update compose.yml |
| ARCHITECTURE.md inconsistency | Med | Do ARCHITECTURE.md update as final step |

---

## Recommendation: ⭐⭐⭐⭐⭐

STRONGLY RECOMMENDED. This is the highest-value lowest-risk change in the entire research option set:
- Fixes a genuine product taxonomy problem (1-service Cipher product)
- Zero business logic change (pure rename)
- Takes ~4.5h
- Paves the way for sm-secrets, sm-ssh, sm-file extensions

**Can be done independently of jose-ja migration, sm-kms migration debt, and pki-ca migration.**
**Should be done FIRST before any other migration work.**

See [tasks-PKI-CA-MERGE0a.md](tasks-PKI-CA-MERGE0a.md) for implementation tasks.
