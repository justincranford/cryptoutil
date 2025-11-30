# Passthru1: Implementation Task List

**Purpose**: Prioritized implementation order for working demos with 85%+ coverage
**Created**: November 29, 2025
**Updated**: November 30, 2025 (Grooming Sessions 5-6 decisions incorporated)
**Timeline**: 1-2 weeks aggressive

---

## Coverage Baseline (2025-01-07)

### KMS Server Coverage (internal/server)

| Package | Coverage | Target | Gap |
|---------|----------|--------|-----|
| application | 65.8% | 85% | +19.2% |
| barrier | 75.5% | 85% | +9.5% |
| contentkeysservice | 81.2% | 85% | +3.8% |
| intermediatekeysservice | 79.2% | 85% | +5.8% |
| rootkeysservice | 79.2% | 85% | +5.8% |
| unsealkeysservice | 49.4% | 85% | +35.6% |
| businesslogic | 37.7% | 85% | +47.3% |
| handler | 27.8% | 85% | +57.2% |
| repository/orm | 90.8% | 85% | ✅ |
| sqlrepository | 77.7% | 85% | +7.3% |

### Identity Server Coverage (internal/identity)

| Package | Coverage | Target | Gap |
|---------|----------|--------|-----|
| authz | 77.1% | 85% | +7.9% |
| authz/clientauth | 78.4% | 85% | +6.6% |
| authz/pkce | 95.5% | 85% | ✅ |
| bootstrap | 82.8% | 85% | +2.2% |
| config | 70.1% | 85% | +14.9% |
| domain | 92.3% | 85% | ✅ |
| healthcheck | 87.1% | 85% | ✅ |
| idp | 57.9% | 85% | +27.1% |
| idp/userauth | 37.1% | 85% | +47.9% |
| issuer | 60.1% | 85% | +24.9% |
| jobs | 89.0% | 85% | ✅ |
| jwks | 77.5% | 85% | +7.5% |
| notifications | 87.8% | 85% | ✅ |
| repository/orm | 67.5% | 85% | +17.5% |
| rotation | 83.7% | 85% | +1.3% |
| rs | 76.4% | 85% | +8.6% |
| security | 100.0% | 85% | ✅ |

### Critical Coverage Gaps (>30% to target)

**KMS (PROTECT - only add tests, don't refactor):**

- handler: 27.8% → 85% (+57.2%) - NEEDS TESTS
- businesslogic: 37.7% → 85% (+47.3%) - NEEDS TESTS
- unsealkeysservice: 49.4% → 85% (+35.6%) - NEEDS TESTS

**Identity (CAN REFACTOR):**

- idp/userauth: 37.1% → 85% (+47.9%) - NEEDS WORK
- idp: 57.9% → 85% (+27.1%) - NEEDS WORK
- issuer: 60.1% → 85% (+24.9%) - NEEDS WORK

---

## Phase 1: KMS Demo Verification (Day 1-2) - IN PROGRESS

Protect existing manual work while improving demo experience.

### T1.1: Verify KMS Server Starts ✅ COMPLETE

- [x] Start KMS with docker compose
- [x] Verify Swagger UI loads
- [x] Verify health endpoints work
- [x] Document current working state

### T1.2: Verify KMS Browser API ✅ COMPLETE

- [x] Test CORS configuration
- [x] Test CSRF token flow
- [x] Verify browser can make API calls
- [x] Document any issues found

### T1.3: Verify KMS Operations ✅ COMPLETE (2025-11-30)

- [ ] Create key pool via API (not yet tested)
- [x] Create key via API (fixed nil pointer dereference)
- [x] Encrypt data via API
- [x] Decrypt data via API
- [x] Sign data via API
- [x] Verify signature via API

### T1.4: KMS Coverage (Add Tests Only - Don't Refactor Code) - NOT STARTED

- Add tests to handler package (+57.2% needed)
- Add tests to businesslogic package (+47.3% needed)
- Add tests to unsealkeysservice (+35.6% needed)
- Add tests to application package (+19.2% needed)
- Target: All KMS packages ≥85%

---

## Phase 2: Identity Demo Assessment (Day 2-3)

Audit LLM-generated code and identify what's broken.

### T2.1: Identity Code Audit

- List all identity packages
- Identify compilation errors
- Identify runtime errors
- Identify missing implementations
- Create prioritized fix list

### T2.2: Identity Database Setup

- Verify SQLite in-memory works
- Verify PostgreSQL connection works
- Verify migrations run
- Test basic CRUD operations

### T2.3: Identity Domain Models

- Verify User model
- Verify Client model
- Verify Session model
- Verify Token model
- Verify Scope model

### T2.4: Identity Repository Layer

- Fix ORM repositories
- Test Create operations
- Test Read operations
- Test Update operations
- Test Delete operations

### T2.5: Identity Coverage Priority (Fix + Tests)

- Fix and test idp/userauth package (+47.9% needed)
- Fix and test idp package (+27.1% needed)
- Fix and test issuer package (+24.9% needed)
- Fix and test repository/orm (+17.5% needed)
- Fix and test config (+14.9% needed)
- Target: All Identity packages ≥85%

---

## Phase 3: Identity Core Flows (Day 3-5)

Fix the OAuth2.1 authorization flows.

### T3.1: Authorization Endpoint

- Fix /authorize endpoint
- Test authorization code flow
- Test PKCE support
- Test redirect handling

### T3.2: Token Endpoint

- Fix /token endpoint
- Test authorization_code grant
- Test client_credentials grant
- Test refresh_token grant
- Verify token format (JWT)

### T3.3: Token Introspection

- Fix /introspect endpoint
- Test active token introspection
- Test expired token introspection
- Test revoked token introspection

### T3.4: Token Revocation

- Fix /revoke endpoint
- Test access token revocation
- Test refresh token revocation
- Verify revocation cascade

### T3.5: User Authentication

- Fix login flow
- Test username/password auth
- Test session creation
- Test session validation

---

## Phase 4: Identity Demo Polish (Day 5-6)

Make Identity demo self-guided and working.

### T4.1: Identity Server Startup

- Single command startup
- Health endpoint working
- OpenAPI spec serving
- Swagger UI working

### T4.2: Demo Data Seeding

- Pre-create demo users
- Pre-create demo clients
- Pre-create demo scopes
- Pre-create demo sessions

### T4.3: Demo Walkthrough

- Document login flow
- Document token flow
- Document introspection
- Document revocation

---

## Phase 5: Integration Demo (Day 6-7)

Combine KMS and Identity into single demo.

### T5.1: KMS Authentication Setup

- Configure KMS to use Identity tokens
- Add token validation middleware
- Test authenticated requests
- Test unauthorized rejection

### T5.2: Scope-Based Authorization

- Define KMS scopes (read:keys, write:keys, etc.)
- Implement scope checking
- Test scope enforcement
- Document scope requirements

### T5.3: Embedded Identity Option

- Create `identity.New(config)` API
- Embed Identity in KMS process
- Test embedded startup
- Test embedded auth flow

### T5.4: Integration Demo Polish

- Single docker compose for both
- Combined Swagger UI
- Unified demo walkthrough
- Reset capability

---

## Phase 6: Documentation and Cleanup (Day 7)

Final polish and documentation.

### T6.1: Demo Documentation

- Update main README
- Create demo walkthrough guide
- Add troubleshooting section
- Record demo video (optional)

### T6.2: Code Cleanup

- Remove dead code
- Fix remaining lint issues
- Update test coverage
- Tag release

---

## Task Tracking

### Status Legend

- `[ ]` Not started
- `[~]` In progress
- `[X]` Complete
- `[!]` Blocked

### Current Progress

| Phase | Tasks | Done | Status |
|-------|-------|------|--------|
| Phase 1: KMS | 4 | 0 | Not started |
| Phase 2: Identity Assess | 5 | 0 | Not started |
| Phase 3: Identity Flows | 5 | 0 | Not started |
| Phase 4: Identity Polish | 3 | 0 | Not started |
| Phase 5: Integration | 4 | 0 | Not started |
| Phase 6: Docs | 2 | 0 | Not started |
| **TOTAL** | **23** | **0** | **0%** |

### Coverage Milestones

| Milestone | Current | Target | Gap |
|-----------|---------|--------|-----|
| KMS Overall | ~65% | 85% | +20% |
| Identity Overall | ~72% | 85% | +13% |
| Combined | ~68% | 85% | +17% |

---

## Dependencies

```plaintext
Phase 1 (KMS)
    ↓
Phase 2 (Identity Assess)
    ↓
Phase 3 (Identity Flows)
    ↓
Phase 4 (Identity Polish)
    ↓
Phase 5 (Integration)
    ↓
Phase 6 (Docs)
```

All phases are sequential. Each depends on the previous being complete.

---

## Risk Mitigation

### KMS Risk: Breaking Working Code

- **Mitigation**: Test after every change
- **Mitigation**: Commit frequently
- **Mitigation**: Don't refactor architecture yet

### Identity Risk: Too Much Broken Code

- **Mitigation**: Audit before fixing
- **Mitigation**: Fix one component at a time
- **Mitigation**: May need to rewrite some parts

### Integration Risk: Interface Mismatch

- **Mitigation**: Define interfaces early
- **Mitigation**: Build adapter layer if needed
- **Mitigation**: Keep both demos working independently

---

**Status**: READY TO START
**Start With**: Phase 1, Task T1.1
