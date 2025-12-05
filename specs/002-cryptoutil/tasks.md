# cryptoutil Tasks - Iteration 2

## Task Breakdown

### Phase 1: JOSE Authority

| ID | Title | Description | Priority | Status | LOE |
|----|-------|-------------|----------|--------|-----|
| JOSE-1 | OpenAPI Specification | Create `api/jose/openapi_spec_*.yaml` for JOSE Authority endpoints | HIGH | ❌ Not Started | 2h |
| JOSE-2 | Server Scaffolding | Create Fiber server in `internal/jose/server/` with middleware | HIGH | ❌ Not Started | 4h |
| JOSE-3 | Key Generation Handler | POST `/jose/v1/keys` - Generate new JWK | HIGH | ❌ Not Started | 4h |
| JOSE-4 | Key Retrieval Handlers | GET `/jose/v1/keys/{kid}`, GET `/jose/v1/keys` | HIGH | ❌ Not Started | 3h |
| JOSE-5 | JWKS Endpoint | GET `/jose/v1/jwks` - Public JWKS | HIGH | ❌ Not Started | 2h |
| JOSE-6 | JWS Sign Handler | POST `/jose/v1/sign` - Create JWS | HIGH | ❌ Not Started | 3h |
| JOSE-7 | JWS Verify Handler | POST `/jose/v1/verify` - Verify JWS | HIGH | ❌ Not Started | 3h |
| JOSE-8 | JWE Encrypt Handler | POST `/jose/v1/encrypt` - Create JWE | MEDIUM | ❌ Not Started | 3h |
| JOSE-9 | JWE Decrypt Handler | POST `/jose/v1/decrypt` - Decrypt JWE | MEDIUM | ❌ Not Started | 3h |
| JOSE-10 | JWT Issue Handler | POST `/jose/v1/jwt/issue` - Issue JWT with claims | HIGH | ❌ Not Started | 4h |
| JOSE-11 | JWT Validate Handler | POST `/jose/v1/jwt/validate` - Validate JWT | HIGH | ❌ Not Started | 4h |
| JOSE-12 | Integration Tests | E2E tests for all JOSE endpoints | HIGH | ❌ Not Started | 6h |
| JOSE-13 | Docker Integration | Add jose-authority to Docker Compose | MEDIUM | ❌ Not Started | 2h |

**Phase 1 Total**: 13 tasks, ~43 hours

---

### Phase 2: CA Server REST API

| ID | Title | Description | Priority | Status | LOE |
|----|-------|-------------|----------|--------|-----|
| CA-1 | OpenAPI Specification | Create `api/ca/openapi_spec_paths.yaml` | HIGH | ❌ Not Started | 4h |
| CA-2 | Server Scaffolding | Fiber server with mTLS in `internal/ca/server/` | HIGH | ❌ Not Started | 4h |
| CA-3 | Health Handler | GET `/ca/v1/health` | HIGH | ❌ Not Started | 1h |
| CA-4 | CA List Handler | GET `/ca/v1/ca` - List available CAs | HIGH | ❌ Not Started | 3h |
| CA-5 | CA Details Handler | GET `/ca/v1/ca/{ca_id}` - CA details + chain | HIGH | ❌ Not Started | 3h |
| CA-6 | CRL Handler | GET `/ca/v1/ca/{ca_id}/crl` - Download CRL | HIGH | ❌ Not Started | 3h |
| CA-7 | Certificate Issue Handler | POST `/ca/v1/certificate` - Issue from CSR | HIGH | ❌ Not Started | 6h |
| CA-8 | Certificate Get Handler | GET `/ca/v1/certificate/{serial}` | MEDIUM | ❌ Not Started | 2h |
| CA-9 | Certificate Revoke Handler | POST `/ca/v1/certificate/{serial}/revoke` | HIGH | ❌ Not Started | 4h |
| CA-10 | Certificate Status Handler | GET `/ca/v1/certificate/{serial}/status` | MEDIUM | ❌ Not Started | 2h |
| CA-11 | OCSP Handler | POST `/ca/v1/ocsp` - RFC 6960 responder | HIGH | ❌ Not Started | 6h |
| CA-12 | Profile List Handler | GET `/ca/v1/profiles` | MEDIUM | ❌ Not Started | 2h |
| CA-13 | Profile Details Handler | GET `/ca/v1/profiles/{profile_id}` | MEDIUM | ❌ Not Started | 2h |
| CA-14 | EST cacerts Handler | GET `/ca/v1/est/cacerts` | MEDIUM | ❌ Not Started | 2h |
| CA-15 | EST simpleenroll Handler | POST `/ca/v1/est/simpleenroll` | MEDIUM | ❌ Not Started | 4h |
| CA-16 | EST simplereenroll Handler | POST `/ca/v1/est/simplereenroll` | LOW | ❌ Not Started | 3h |
| CA-17 | EST serverkeygen Handler | POST `/ca/v1/est/serverkeygen` | LOW | ❌ Not Started | 3h |
| CA-18 | TSA Handler | POST `/ca/v1/tsa/timestamp` - RFC 3161 | MEDIUM | ❌ Not Started | 4h |
| CA-19 | Integration Tests | E2E tests for CA endpoints | HIGH | ❌ Not Started | 8h |
| CA-20 | Docker Integration | Add ca-server to Docker Compose | MEDIUM | ❌ Not Started | 2h |

**Phase 2 Total**: 20 tasks, ~68 hours

---

### Phase 3: Unified Suite

| ID | Title | Description | Priority | Status | LOE |
|----|-------|-------------|----------|--------|-----|
| UNIFIED-1 | Docker Compose Update | Add all 4 services to compose.yml | HIGH | ❌ Not Started | 4h |
| UNIFIED-2 | Shared Secrets Config | Unified unseal secret configuration | HIGH | ❌ Not Started | 2h |
| UNIFIED-3 | Service Discovery | Inter-service communication setup | HIGH | ❌ Not Started | 4h |
| UNIFIED-4 | Health Checks | All services report healthy | HIGH | ❌ Not Started | 2h |
| UNIFIED-5 | E2E Demo Script | `cmd/demo unified` command | MEDIUM | ❌ Not Started | 6h |
| UNIFIED-6 | Documentation | Architecture docs and runbooks | MEDIUM | ❌ Not Started | 4h |

**Phase 3 Total**: 6 tasks, ~22 hours

---

## Summary

| Phase | Tasks | LOE |
|-------|-------|-----|
| Phase 1: JOSE Authority | 13 | ~43h |
| Phase 2: CA Server | 20 | ~68h |
| Phase 3: Unified Suite | 6 | ~22h |
| **Total** | **39** | **~133h** |

---

## Task Dependencies

```
JOSE-1 ─→ JOSE-2 ─→ JOSE-3 ─→ JOSE-4
                    │         │
                    ├→ JOSE-5 │
                    │         │
                    ├→ JOSE-6 ─→ JOSE-10
                    │         │
                    ├→ JOSE-7 ─→ JOSE-11
                    │
                    ├→ JOSE-8
                    │
                    └→ JOSE-9

JOSE-10, JOSE-11 ─→ JOSE-12 ─→ JOSE-13

CA-1 ─→ CA-2 ─→ CA-3 ─→ CA-4 ─→ CA-5
                        │
                        ├→ CA-6
                        │
                        ├→ CA-7 ─→ CA-8
                        │         │
                        ├→ CA-9 ─→ CA-10
                        │
                        ├→ CA-11
                        │
                        ├→ CA-12 ─→ CA-13
                        │
                        ├→ CA-14 ─→ CA-15 ─→ CA-16 ─→ CA-17
                        │
                        └→ CA-18

CA-* ─→ CA-19 ─→ CA-20

JOSE-13, CA-20 ─→ UNIFIED-1 ─→ UNIFIED-2 ─→ UNIFIED-3 ─→ UNIFIED-4 ─→ UNIFIED-5 ─→ UNIFIED-6
```

---

## Priority Matrix

### HIGH Priority (Must Complete)

- JOSE-1, JOSE-2, JOSE-3, JOSE-4, JOSE-5, JOSE-6, JOSE-7, JOSE-10, JOSE-11, JOSE-12
- CA-1, CA-2, CA-3, CA-4, CA-5, CA-6, CA-7, CA-9, CA-11, CA-19
- UNIFIED-1, UNIFIED-2, UNIFIED-3, UNIFIED-4

### MEDIUM Priority (Should Complete)

- JOSE-8, JOSE-9, JOSE-13
- CA-8, CA-10, CA-12, CA-13, CA-14, CA-15, CA-18, CA-20
- UNIFIED-5, UNIFIED-6

### LOW Priority (Nice to Have)

- CA-16, CA-17

---

*Tasks Version: 2.0.0*
*Created: January 15, 2025*
