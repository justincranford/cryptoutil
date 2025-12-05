# cryptoutil Tasks - Iteration 2

## Task Breakdown

### Phase 1: JOSE Authority

| ID | Title | Description | Priority | Status | LOE |
|----|-------|-------------|----------|--------|-----|
| JOSE-1 | OpenAPI Specification | Create `api/jose/openapi_spec.yaml` for JOSE Authority endpoints | HIGH | ✅ Complete | 2h |
| JOSE-2 | Server Scaffolding | Create Fiber server in `internal/jose/server/` with middleware | HIGH | ✅ Complete | 4h |
| JOSE-3 | Key Generation Handler | POST `/jose/v1/jwk/generate` - Generate new JWK | HIGH | ✅ Complete | 4h |
| JOSE-4 | Key Retrieval Handlers | GET `/jose/v1/jwk/{kid}`, GET `/jose/v1/jwk`, DELETE `/jose/v1/jwk/{kid}/delete` | HIGH | ✅ Complete | 3h |
| JOSE-5 | JWKS Endpoint | GET `/jose/v1/jwks` and `/.well-known/jwks.json` - Public JWKS | HIGH | ✅ Complete | 2h |
| JOSE-6 | JWS Sign Handler | POST `/jose/v1/jws/sign` - Create JWS | HIGH | ✅ Complete | 3h |
| JOSE-7 | JWS Verify Handler | POST `/jose/v1/jws/verify` - Verify JWS | HIGH | ✅ Complete | 3h |
| JOSE-8 | JWE Encrypt Handler | POST `/jose/v1/jwe/encrypt` - Create JWE | MEDIUM | ✅ Complete | 3h |
| JOSE-9 | JWE Decrypt Handler | POST `/jose/v1/jwe/decrypt` - Decrypt JWE | MEDIUM | ✅ Complete | 3h |
| JOSE-10 | JWT Issue Handler | POST `/jose/v1/jwt/sign` - Issue JWT with claims | HIGH | ✅ Complete | 4h |
| JOSE-11 | JWT Validate Handler | POST `/jose/v1/jwt/verify` - Validate JWT | HIGH | ✅ Complete | 4h |
| JOSE-12 | Integration Tests | E2E tests for all JOSE endpoints | HIGH | ✅ Complete | 6h |
| JOSE-13 | Docker Integration | Add jose-authority to Docker Compose | MEDIUM | ❌ Not Started | 2h |

**Phase 1 Total**: 13 tasks, 12 complete, ~43 hours

---

### Phase 2: CA Server REST API

| ID | Title | Description | Priority | Status | LOE |
|----|-------|-------------|----------|--------|-----|
| CA-1 | OpenAPI Specification | `api/ca/openapi_spec_enrollment.yaml` | HIGH | ✅ Complete | 4h |
| CA-2 | Server Scaffolding | Handler struct in `internal/ca/api/handler/` | HIGH | ✅ Complete | 4h |
| CA-3 | Health Handler | Health check endpoints | HIGH | ❌ Not Started | 1h |
| CA-4 | CA List Handler | GET `/ca` - List available CAs | HIGH | ❌ Not Started | 3h |
| CA-5 | CA Details Handler | GET `/ca/{ca_id}` - CA details + chain | HIGH | ❌ Not Started | 3h |
| CA-6 | CRL Handler | GET `/ca/{ca_id}/crl` - Download CRL | HIGH | ❌ Not Started | 3h |
| CA-7 | Certificate Issue Handler | POST `/enroll` - Issue from CSR | HIGH | ✅ Complete | 6h |
| CA-8 | Certificate Get Handler | GET `/certificates/{serial}` + list + chain | MEDIUM | ✅ Complete | 2h |
| CA-9 | Certificate Revoke Handler | POST `/certificate/{serial}/revoke` | HIGH | ❌ Not Started | 4h |
| CA-10 | Certificate Status Handler | GET `/enroll/{requestId}` - Enrollment status | MEDIUM | ⚠️ Partial (returns NotImplemented) | 2h |
| CA-11 | OCSP Handler | POST `/ocsp` - RFC 6960 responder | HIGH | ❌ Not Started | 6h |
| CA-12 | Profile List Handler | GET `/profiles` | MEDIUM | ✅ Complete | 2h |
| CA-13 | Profile Details Handler | GET `/profiles/{profile_id}` | MEDIUM | ✅ Complete | 2h |
| CA-14 | EST cacerts Handler | GET `/est/cacerts` | MEDIUM | ❌ Not Started | 2h |
| CA-15 | EST simpleenroll Handler | POST `/est/simpleenroll` | MEDIUM | ❌ Not Started | 4h |
| CA-16 | EST simplereenroll Handler | POST `/est/simplereenroll` | LOW | ❌ Not Started | 3h |
| CA-17 | EST serverkeygen Handler | POST `/est/serverkeygen` | LOW | ❌ Not Started | 3h |
| CA-18 | TSA Handler | POST `/tsa/timestamp` - RFC 3161 | MEDIUM | ⚠️ Exists in `service/timestamp` | 4h |
| CA-19 | Integration Tests | E2E tests for CA endpoints | HIGH | ⚠️ Partial (handler_test.go exists but low coverage) | 8h |
| CA-20 | Docker Integration | Add ca-server to Docker Compose | MEDIUM | ❌ Not Started | 2h |

**Phase 2 Total**: 20 tasks, 6 complete, 2 partial, ~68 hours

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

## Summary

| Phase | Tasks | Complete | Partial | Remaining |
|-------|-------|----------|---------|-----------|
| Phase 1: JOSE Authority | 13 | 12 | 0 | 1 |
| Phase 2: CA Server | 20 | 6 | 2 | 12 |
| Phase 3: Unified Suite | 6 | 0 | 0 | 6 |
| **Total** | **39** | **18** | **2** | **19** |

**Progress**: 46% complete, 5% partial

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

- ✅ JOSE-1 through JOSE-12 (complete)
- JOSE-13 (Docker Integration) - remaining
- CA-3, CA-4, CA-5, CA-6, CA-9, CA-11 - not started
- CA-19 (Integration Tests) - partial
- UNIFIED-1, UNIFIED-2, UNIFIED-3, UNIFIED-4 - not started

### MEDIUM Priority (Should Complete)

- ✅ CA-12, CA-13 (complete)
- CA-8, CA-10 (partial - return NotImplemented)
- CA-14, CA-15, CA-18, CA-20 - not started
- UNIFIED-5, UNIFIED-6 - not started

### LOW Priority (Nice to Have)

- CA-16, CA-17 - not started

---

*Tasks Version: 2.1.0*
*Last Updated: December 5, 2025*
