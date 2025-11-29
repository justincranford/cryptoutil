# Passthru1: Aggressive Task List

**Purpose**: Prioritized implementation order for working demos
**Created**: November 29, 2025
**Timeline**: 1-2 weeks aggressive

---

## Phase 1: KMS Demo Verification (Day 1-2)

Protect existing manual work while improving demo experience.

### T1.1: Verify KMS Server Starts

- Start KMS with docker compose
- Verify Swagger UI loads
- Verify health endpoints work
- Document current working state

### T1.2: Verify KMS Browser API

- Test CORS configuration
- Test CSRF token flow
- Verify browser can make API calls
- Document any issues found

### T1.3: Verify KMS Operations

- Create key pool via API
- Create key via API
- Encrypt data via API
- Decrypt data via API
- Sign data via API
- Verify signature via API

### T1.4: KMS Demo Enhancements (Optional)

- Add pre-seeded demo accounts
- Add pre-seeded key hierarchies
- Improve Swagger UI descriptions
- Add demo reset capability

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
| Phase 2: Identity Assess | 4 | 0 | Not started |
| Phase 3: Identity Flows | 5 | 0 | Not started |
| Phase 4: Identity Polish | 3 | 0 | Not started |
| Phase 5: Integration | 4 | 0 | Not started |
| Phase 6: Docs | 2 | 0 | Not started |
| **TOTAL** | **22** | **0** | **0%** |

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
