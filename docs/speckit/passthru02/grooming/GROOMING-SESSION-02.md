# Speckit Passthru02 Grooming Session 02: Plan & Implementation Strategy

**Purpose**: Structured questions to refine implementation plan, phase priorities, task sequencing, and technical decisions for cryptoutil.
**Created**: 2025-12-02
**Status**: ✅ COMPLETED (2025-12-02)

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question. Add comments in the "Notes" field if needed. Multiple selections allowed where indicated.

---

## Section 1: Phase 1 - Identity V2 Production (Q1-10)

### Q1. Phase 1 Timeline

Plan says 2-4 weeks for Phase 1 (Identity completion). Is this realistic?

- [ ] A. Too aggressive - needs 6-8 weeks
- [ ] B. About right - 2-4 weeks with focused effort
- [x] C. Too conservative - can complete in 1-2 weeks
- [ ] D. Unknown - need to complete first task to estimate

**Notes**:

---

### Q2. Login UI Priority

Login UI is marked HIGH priority. What implementation approach?

- [x] A. Minimal HTML - server-rendered, no JavaScript
- [ ] B. Simple templates - Go templates with basic CSS
- [ ] C. Tailwind CSS - utility-first styling
- [ ] D. React/Vue - modern SPA approach
- [ ] E. Skip UI - API-only, external apps provide UI

**Notes**:

---

### Q3. Consent UI Requirements

What information must the consent screen display?

- [ ] A. Client name and requested scopes only
- [x] B. Client name, scopes, and data access summary
- [x] C. Full OAuth 2.1 compliant disclosure
- [x] D. Configurable - minimal to comprehensive based on client sensitivity

**Notes**:

---

### Q4. Logout Flow Scope

Which logout flows are required for Phase 1?

- [ ] A. Basic: Clear session, revoke tokens
- [ ] B. Front-channel: Browser redirect to client post_logout_redirect_uri
- [ ] C. Back-channel: Server-to-server logout notifications
- [x] D. All of the above
- [ ] E. Basic only - defer front/back-channel to later phase

**Notes**:

---

### Q5. Userinfo Endpoint Response Format

What format should `/oidc/v1/userinfo` return?

- [ ] A. JSON only - simple, widely supported
- [x] B. JWT only - signed for integrity
- [ ] C. Both JSON and JWT - Accept header negotiation
- [ ] D. JSON default, JWT via request parameter

**Notes**:
I assume JWT is OAuth 2.1 mandatory requirement, so JWT only.
If I'm wrong, please correct me.

---

### Q6. Client Secret Storage

How should client secrets be stored?

- [ ] A. Plaintext - simplest, acceptable for dev
- [x] B. PBKDF2-HMAC-SHA256 - FIPS-compliant hash
- [ ] C. bcrypt - industry standard (NOTE: NOT FIPS-compliant)
- [ ] D. Argon2id - memory-hard (NOTE: NOT FIPS-compliant)

**Notes**:

---

### Q7. Token-User Association

Current implementation has placeholder user association. How to fix?

- [ ] A. Store user_id in token claims
- [ ] B. Store user_id in token database record
- [x] C. Both - claims for stateless validation, DB for revocation
- [x] D. External user store - federate to separate user service

**Notes**:
KMS needs to support MANDATORY configurable realms for users and clients, file-based and database-based, like Elasticsearch 7.10.
KMS needs to support OPTIONAL configurable federation to realms in IdPs and AuthZs.

---

### Q8. Token Cleanup Strategy

How should expired tokens be cleaned up?

- [ ] A. Background goroutine - periodic cleanup job
- [ ] B. On-access - check expiration on every token operation
- [x] C. Cron job - external scheduler
- [x] D. TTL - database-level expiration (PostgreSQL)
- [x] E. Hybrid - on-access validation + periodic cleanup

**Notes**:

---

### Q9. Rate Limiting Strategy

What rate limiting approach for Phase 1?

- [ ] A. Per-IP rate limiting - X requests per minute
- [ ] B. Per-client rate limiting - based on client_id
- [ ] C. Per-endpoint rate limiting - different limits per endpoint
- [x] D. Tiered - IP + client + endpoint combination
- [ ] E. Defer - no rate limiting in Phase 1

**Notes**:

---

### Q10. Audit Logging Scope

What events should be audit logged in Phase 1?

- [ ] A. Authentication events only (login success/failure)
- [ ] B. Authorization events (token issuance, consent)
- [ ] C. Administrative events (client CRUD, secret rotation)
- [ ] D. All of the above
- [x] E. All + token introspection and revocation

**Notes**:

---

## Section 2: Phase 2 - KMS Stabilization (Q11-15)

### Q11. KMS Demo Reliability

`go run ./cmd/demo kms` reliability goal?

- [x] A. 100% reliable - must never fail
- [ ] B. 95% reliable - occasional failures acceptable
- [ ] C. Best effort - document known issues
- [ ] D. Not important - KMS demo is secondary

**Notes**:

---

### Q12. KMS API Documentation Priority

Which documentation is most needed for KMS?

- [x] A. OpenAPI spec completion - accurate, up-to-date
- [ ] B. Example requests/responses - curl examples
- [ ] C. Error code documentation - comprehensive error reference
- [ ] D. Usage guide - step-by-step tutorials
- [ ] E. All equally important

**Notes**:
Only OpenAPI spec, and minimum executive summary.

---

### Q13. KMS Integration Testing Scope

What E2E scenarios must be tested?

- [x] A. Key lifecycle - create, read, list, rotate
- [x] B. Crypto operations - encrypt, decrypt, sign, verify
- [x] C. Multi-tenant isolation - tenant A cannot access tenant B keys
- [ ] D. Failure scenarios - network errors, DB failures
- [ ] E. All of the above

**Notes**:

---

### Q14. KMS Performance Baseline

What performance targets for KMS?

- [x] A. No specific targets - measure and document
- [ ] B. 100 ops/sec - basic performance
- [ ] C. 1000 ops/sec - production-grade
- [ ] D. 10000 ops/sec - high-performance
- [ ] E. Depends on operation type - different targets for encrypt vs. key creation

**Notes**:

---

### Q15. KMS-Identity Integration Priority

When should KMS federate authentication to Identity?

- [ ] A. Phase 2 - immediate integration
- [x] B. Phase 3 - dedicated integration phase (per plan)
- [ ] C. Phase 4 - after CA foundation
- [ ] D. Never - KMS should have independent authentication

**Notes**:

---

## Section 3: Phase 3 - Integration Demo (Q16-20)

### Q16. Integration Demo Scope

What should `go run ./cmd/demo integration` demonstrate?

- [ ] A. KMS using Identity for auth only
- [ ] B. KMS + Identity + telemetry
- [x] C. Full stack - KMS + Identity + Telemetry + PostgreSQL
- [x] D. All products - JOSE + Identity + KMS (+ CA when ready)

**Notes**:

---

### Q17. OAuth2 Client for KMS

How should KMS be registered as OAuth2 client?

- [x] A. Pre-seeded - demo data includes KMS client
- [x] B. Bootstrap script - automated registration
- [ ] C. Manual registration - documented steps
- [ ] D. Dynamic client registration (RFC 7591)

**Notes**:

---

### Q18. Token Validation Location

Where should KMS validate tokens from Identity?

- [ ] A. Local validation - fetch JWKS, validate locally
- [ ] B. Remote introspection - call Identity introspection endpoint
- [x] C. Hybrid - local first, introspection for revocation check
- [ ] D. Gateway - put API gateway in front of KMS

**Notes**:

---

### Q19. Scope-Based Authorization

How should KMS enforce scopes?

- [ ] A. Simple - read vs. write scopes
- [ ] B. Operation-based - encrypt:*, decrypt:*, sign:*, verify:*
- [x] C. Resource-based - elastickey:read, materialkey:create, etc.
- [x] D. Fine-grained - elastickey:{id}:{operation}

**Notes**:
Hybrid scopes

---

### Q20. Integration Demo Success Criteria

What proves integration demo is "complete"?

- [x] A. `go run ./cmd/demo all` exits 0
- [x] B. All 7/7 demo steps pass
- [x] C. Docker Compose deployment healthy + demo passes
- [ ] D. Documentation complete + above

**Notes**:
C is most important because that is how I will manually test.
If docker compose fails for any reason, or UIs not accessible for any reason, or UI screens like login/logout not working, them demo is a failure.

---

## Section 4: Phase 4 - CA Foundation (Q21-25)

### Q21. CA Domain Charter Scope

What should the CA domain charter define?

- [ ] A. Technical scope only - algorithms, formats, protocols
- [ ] B. Compliance scope - CA/Browser Forum, RFC 5280
- [ ] C. Organizational scope - roles, responsibilities, audit
- [x] D. All of the above

**Notes**:

---

### Q22. CA Crypto Provider Strategy

What crypto provider model for CA?

- [ ] A. Memory only - keys in process memory
- [ ] B. Filesystem - keys in encrypted files
- [x] C. Database - keys in PostgreSQL/SQLite
- [x] D. Pluggable - interface supporting memory/file/DB/HSM
- [ ] E. HSM-first - design around HSM, soft key fallback

**Notes**:

---

### Q23. Certificate Profile Library

How many certificate profiles should be predefined?

- [ ] A. Minimal (5-10) - root, intermediate, server, client, code signing
- [ ] B. Moderate (15-20) - above + email, VPN, device, timestamp
- [x] C. Comprehensive (25+) - per spec.md profile list
- [x] D. Configurable - template-based, user-defined profiles

**Notes**:

---

### Q24. CA Revocation Strategy

Which revocation mechanisms are required?

- [ ] A. CRL only - Certificate Revocation Lists
- [ ] B. OCSP only - Online Certificate Status Protocol
- [ ] C. Both CRL and OCSP
- [x] D. CRL + OCSP + OCSP Stapling
- [x] E. Defer revocation - implement in later phase

**Notes**:
D, but defer implementation to layer stage

---

### Q25. CA Timeline Assessment

Plan says 4-8 weeks for CA foundation. Is this realistic?

- [ ] A. Too aggressive - needs 12+ weeks
- [x] B. About right - 4-8 weeks with focused effort
- [ ] C. Too conservative - can complete in 2-4 weeks
- [ ] D. Unknown - depends on prior phase completion

**Notes**:

---

## Section 5: Technical Decisions (Q26-32)

### Q26. Database Migration Strategy

How should database schema changes be managed?

- [ ] A. GORM AutoMigrate - automatic schema updates
- [ ] B. golang-migrate - versioned SQL migrations
- [ ] C. Both - AutoMigrate for dev, migrations for prod
- [ ] D. Manual - hand-written migration scripts

**Notes**:
I think golang-migrate is used now? Whatever is used now, I assume that is ok.

---

### Q27. Configuration File Format

What format for configuration files?

- [x] A. YAML only - current standard
- [ ] B. TOML - simpler, less ambiguous
- [ ] C. JSON - machine-friendly
- [ ] D. Multiple - support YAML, TOML, JSON with autodetection

**Notes**:

---

### Q28. Error Response Format

What error response format for APIs?

- [x] A. RFC 7807 Problem Details - standard error format
- [ ] B. Simple JSON - `{"error": "message", "code": "ERR_001"}`
- [x] C. OAuth 2.1 format - `{"error": "invalid_request", "error_description": "..."}`
- [x] D. Per-product - OAuth errors for Identity, custom for KMS/CA

**Notes**:

---

### Q29. API Versioning Strategy

How should APIs be versioned?

- [x] A. URL path - `/v1/`, `/v2/`
- [ ] B. Header - `Accept-Version: v1`
- [ ] C. No versioning - single version, breaking changes allowed
- [ ] D. Per-product - Identity uses v1, KMS unversioned

**Notes**:

---

### Q30. Concurrent Connection Handling

How many concurrent connections should each service support?

- [x] A. Low (100) - development/demo use
- [ ] B. Medium (1000) - production single-node
- [ ] C. High (10000) - production multi-node
- [x] D. Configurable - per-deployment tuning

**Notes**:
A default, configurable

---

### Q31. TLS Certificate Strategy

How should TLS certificates be managed for deployments?

- [ ] A. Self-signed - generated at startup
- [ ] B. Pre-generated - mounted via Docker secrets
- [ ] C. Let's Encrypt - ACME automatic
- [x] D. Internal CA - use P4 (CA) to issue certificates
- [ ] E. Configurable - support all above

**Notes**:
D is highest priority.
C would be nice later.
A NEVER!!!!
B maybe in far future

---

### Q32. Health Check Strategy

What should health endpoints check?

- [ ] A. Liveness - process alive only
- [ ] B. Readiness - database connected
- [ ] C. Deep health - database + external dependencies
- [x] D. Configurable - shallow for k8s probes, deep for monitoring

**Notes**:
B default, configurable to be A B or C

---

## Section 6: Risk & Blockers (Q33-35)

### Q33. Phase 1 Blockers

What could block Phase 1 completion? (Select all that apply)

- [ ] A. UI complexity - HTML/CSS/JS learning curve
- [x] B. Token-user association - database schema changes
- [ ] C. Secret hashing migration - breaking change for existing secrets
- [ ] D. Rate limiting complexity - need distributed rate limiter
- [ ] E. Audit logging volume - storage and performance concerns

**Notes**:

---

### Q34. Technical Debt Tolerance

How much technical debt is acceptable during Phase 1?

- [x] A. Zero - all code must be production-quality
- [ ] B. Low - minimal TODOs, tracked in PROJECT-STATUS.md
- [ ] C. Medium - TODOs acceptable, fix in Phase 5
- [ ] D. High - ship features fast, refactor later

**Notes**:

---

### Q35. Go-Live Criteria

What criteria must be met before "production ready"?

- [ ] A. All tests pass, coverage targets met
- [x] B. Above + security review completed
- [ ] C. Above + performance benchmarks pass
- [ ] D. Above + external audit (penetration test)
- [ ] E. Above + documentation complete + runbooks

**Notes**:
B, you will do the security review as part of plan

---

## Summary & Action Items

After completing this grooming session:

1. Review answers for Phase 1 priorities
2. Identify any dependencies between decisions
3. Update plan.md with refined estimates
4. Update spec.md if scope changes
5. Create detailed task breakdown for next sprint
6. Share answers with Copilot for implementation guidance

---

## Priority Matrix (Optional)

Fill in your top priorities:

| Rank | Area | Task/Decision |
|------|------|---------------|
| 1 | | |
| 2 | | |
| 3 | | |
| 4 | | |
| 5 | | |

---

*Session Created: 2025-12-02*
*Expected Completion: [DATE]*
