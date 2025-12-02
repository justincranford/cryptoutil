# Speckit Passthru02 Grooming Session 02: Plan & Implementation Strategy

**Purpose**: Structured questions to refine implementation plan, phase priorities, task sequencing, and technical decisions for cryptoutil.
**Created**: 2025-12-02
**Status**: AWAITING ANSWERS

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question. Add comments in the "Notes" field if needed. Multiple selections allowed where indicated.

---

## Section 1: Phase 1 - Identity V2 Production (Q1-10)

### Q1. Phase 1 Timeline

Plan says 2-4 weeks for Phase 1 (Identity completion). Is this realistic?

- [ ] A. Too aggressive - needs 6-8 weeks
- [ ] B. About right - 2-4 weeks with focused effort
- [ ] C. Too conservative - can complete in 1-2 weeks
- [ ] D. Unknown - need to complete first task to estimate

**Notes**:

---

### Q2. Login UI Priority

Login UI is marked HIGH priority. What implementation approach?

- [ ] A. Minimal HTML - server-rendered, no JavaScript
- [ ] B. Simple templates - Go templates with basic CSS
- [ ] C. Tailwind CSS - utility-first styling
- [ ] D. React/Vue - modern SPA approach
- [ ] E. Skip UI - API-only, external apps provide UI

**Notes**:

---

### Q3. Consent UI Requirements

What information must the consent screen display?

- [ ] A. Client name and requested scopes only
- [ ] B. Client name, scopes, and data access summary
- [ ] C. Full OAuth 2.1 compliant disclosure
- [ ] D. Configurable - minimal to comprehensive based on client sensitivity

**Notes**:

---

### Q4. Logout Flow Scope

Which logout flows are required for Phase 1?

- [ ] A. Basic: Clear session, revoke tokens
- [ ] B. Front-channel: Browser redirect to client post_logout_redirect_uri
- [ ] C. Back-channel: Server-to-server logout notifications
- [ ] D. All of the above
- [ ] E. Basic only - defer front/back-channel to later phase

**Notes**:

---

### Q5. Userinfo Endpoint Response Format

What format should `/oidc/v1/userinfo` return?

- [ ] A. JSON only - simple, widely supported
- [ ] B. JWT only - signed for integrity
- [ ] C. Both JSON and JWT - Accept header negotiation
- [ ] D. JSON default, JWT via request parameter

**Notes**:

---

### Q6. Client Secret Storage

How should client secrets be stored?

- [ ] A. Plaintext - simplest, acceptable for dev
- [ ] B. PBKDF2-HMAC-SHA256 - FIPS-compliant hash
- [ ] C. bcrypt - industry standard (NOTE: NOT FIPS-compliant)
- [ ] D. Argon2id - memory-hard (NOTE: NOT FIPS-compliant)

**Notes**:

---

### Q7. Token-User Association

Current implementation has placeholder user association. How to fix?

- [ ] A. Store user_id in token claims
- [ ] B. Store user_id in token database record
- [ ] C. Both - claims for stateless validation, DB for revocation
- [ ] D. External user store - federate to separate user service

**Notes**:

---

### Q8. Token Cleanup Strategy

How should expired tokens be cleaned up?

- [ ] A. Background goroutine - periodic cleanup job
- [ ] B. On-access - check expiration on every token operation
- [ ] C. Cron job - external scheduler
- [ ] D. TTL - database-level expiration (PostgreSQL)
- [ ] E. Hybrid - on-access validation + periodic cleanup

**Notes**:

---

### Q9. Rate Limiting Strategy

What rate limiting approach for Phase 1?

- [ ] A. Per-IP rate limiting - X requests per minute
- [ ] B. Per-client rate limiting - based on client_id
- [ ] C. Per-endpoint rate limiting - different limits per endpoint
- [ ] D. Tiered - IP + client + endpoint combination
- [ ] E. Defer - no rate limiting in Phase 1

**Notes**:

---

### Q10. Audit Logging Scope

What events should be audit logged in Phase 1?

- [ ] A. Authentication events only (login success/failure)
- [ ] B. Authorization events (token issuance, consent)
- [ ] C. Administrative events (client CRUD, secret rotation)
- [ ] D. All of the above
- [ ] E. All + token introspection and revocation

**Notes**:

---

## Section 2: Phase 2 - KMS Stabilization (Q11-15)

### Q11. KMS Demo Reliability

`go run ./cmd/demo kms` reliability goal?

- [ ] A. 100% reliable - must never fail
- [ ] B. 95% reliable - occasional failures acceptable
- [ ] C. Best effort - document known issues
- [ ] D. Not important - KMS demo is secondary

**Notes**:

---

### Q12. KMS API Documentation Priority

Which documentation is most needed for KMS?

- [ ] A. OpenAPI spec completion - accurate, up-to-date
- [ ] B. Example requests/responses - curl examples
- [ ] C. Error code documentation - comprehensive error reference
- [ ] D. Usage guide - step-by-step tutorials
- [ ] E. All equally important

**Notes**:

---

### Q13. KMS Integration Testing Scope

What E2E scenarios must be tested?

- [ ] A. Key lifecycle - create, read, list, rotate
- [ ] B. Crypto operations - encrypt, decrypt, sign, verify
- [ ] C. Multi-tenant isolation - tenant A cannot access tenant B keys
- [ ] D. Failure scenarios - network errors, DB failures
- [ ] E. All of the above

**Notes**:

---

### Q14. KMS Performance Baseline

What performance targets for KMS?

- [ ] A. No specific targets - measure and document
- [ ] B. 100 ops/sec - basic performance
- [ ] C. 1000 ops/sec - production-grade
- [ ] D. 10000 ops/sec - high-performance
- [ ] E. Depends on operation type - different targets for encrypt vs. key creation

**Notes**:

---

### Q15. KMS-Identity Integration Priority

When should KMS federate authentication to Identity?

- [ ] A. Phase 2 - immediate integration
- [ ] B. Phase 3 - dedicated integration phase (per plan)
- [ ] C. Phase 4 - after CA foundation
- [ ] D. Never - KMS should have independent authentication

**Notes**:

---

## Section 3: Phase 3 - Integration Demo (Q16-20)

### Q16. Integration Demo Scope

What should `go run ./cmd/demo integration` demonstrate?

- [ ] A. KMS using Identity for auth only
- [ ] B. KMS + Identity + telemetry
- [ ] C. Full stack - KMS + Identity + Telemetry + PostgreSQL
- [ ] D. All products - JOSE + Identity + KMS (+ CA when ready)

**Notes**:

---

### Q17. OAuth2 Client for KMS

How should KMS be registered as OAuth2 client?

- [ ] A. Pre-seeded - demo data includes KMS client
- [ ] B. Bootstrap script - automated registration
- [ ] C. Manual registration - documented steps
- [ ] D. Dynamic client registration (RFC 7591)

**Notes**:

---

### Q18. Token Validation Location

Where should KMS validate tokens from Identity?

- [ ] A. Local validation - fetch JWKS, validate locally
- [ ] B. Remote introspection - call Identity introspection endpoint
- [ ] C. Hybrid - local first, introspection for revocation check
- [ ] D. Gateway - put API gateway in front of KMS

**Notes**:

---

### Q19. Scope-Based Authorization

How should KMS enforce scopes?

- [ ] A. Simple - read vs. write scopes
- [ ] B. Operation-based - encrypt:*, decrypt:*, sign:*, verify:*
- [ ] C. Resource-based - elastickey:read, materialkey:create, etc.
- [ ] D. Fine-grained - elastickey:{id}:{operation}

**Notes**:

---

### Q20. Integration Demo Success Criteria

What proves integration demo is "complete"?

- [ ] A. `go run ./cmd/demo all` exits 0
- [ ] B. All 7/7 demo steps pass
- [ ] C. Docker Compose deployment healthy + demo passes
- [ ] D. Documentation complete + above

**Notes**:

---

## Section 4: Phase 4 - CA Foundation (Q21-25)

### Q21. CA Domain Charter Scope

What should the CA domain charter define?

- [ ] A. Technical scope only - algorithms, formats, protocols
- [ ] B. Compliance scope - CA/Browser Forum, RFC 5280
- [ ] C. Organizational scope - roles, responsibilities, audit
- [ ] D. All of the above

**Notes**:

---

### Q22. CA Crypto Provider Strategy

What crypto provider model for CA?

- [ ] A. Memory only - keys in process memory
- [ ] B. Filesystem - keys in encrypted files
- [ ] C. Database - keys in PostgreSQL/SQLite
- [ ] D. Pluggable - interface supporting memory/file/DB/HSM
- [ ] E. HSM-first - design around HSM, soft key fallback

**Notes**:

---

### Q23. Certificate Profile Library

How many certificate profiles should be predefined?

- [ ] A. Minimal (5-10) - root, intermediate, server, client, code signing
- [ ] B. Moderate (15-20) - above + email, VPN, device, timestamp
- [ ] C. Comprehensive (25+) - per spec.md profile list
- [ ] D. Configurable - template-based, user-defined profiles

**Notes**:

---

### Q24. CA Revocation Strategy

Which revocation mechanisms are required?

- [ ] A. CRL only - Certificate Revocation Lists
- [ ] B. OCSP only - Online Certificate Status Protocol
- [ ] C. Both CRL and OCSP
- [ ] D. CRL + OCSP + OCSP Stapling
- [ ] E. Defer revocation - implement in later phase

**Notes**:

---

### Q25. CA Timeline Assessment

Plan says 4-8 weeks for CA foundation. Is this realistic?

- [ ] A. Too aggressive - needs 12+ weeks
- [ ] B. About right - 4-8 weeks with focused effort
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

---

### Q27. Configuration File Format

What format for configuration files?

- [ ] A. YAML only - current standard
- [ ] B. TOML - simpler, less ambiguous
- [ ] C. JSON - machine-friendly
- [ ] D. Multiple - support YAML, TOML, JSON with autodetection

**Notes**:

---

### Q28. Error Response Format

What error response format for APIs?

- [ ] A. RFC 7807 Problem Details - standard error format
- [ ] B. Simple JSON - `{"error": "message", "code": "ERR_001"}`
- [ ] C. OAuth 2.1 format - `{"error": "invalid_request", "error_description": "..."}`
- [ ] D. Per-product - OAuth errors for Identity, custom for KMS/CA

**Notes**:

---

### Q29. API Versioning Strategy

How should APIs be versioned?

- [ ] A. URL path - `/v1/`, `/v2/`
- [ ] B. Header - `Accept-Version: v1`
- [ ] C. No versioning - single version, breaking changes allowed
- [ ] D. Per-product - Identity uses v1, KMS unversioned

**Notes**:

---

### Q30. Concurrent Connection Handling

How many concurrent connections should each service support?

- [ ] A. Low (100) - development/demo use
- [ ] B. Medium (1000) - production single-node
- [ ] C. High (10000) - production multi-node
- [ ] D. Configurable - per-deployment tuning

**Notes**:

---

### Q31. TLS Certificate Strategy

How should TLS certificates be managed for deployments?

- [ ] A. Self-signed - generated at startup
- [ ] B. Pre-generated - mounted via Docker secrets
- [ ] C. Let's Encrypt - ACME automatic
- [ ] D. Internal CA - use P4 (CA) to issue certificates
- [ ] E. Configurable - support all above

**Notes**:

---

### Q32. Health Check Strategy

What should health endpoints check?

- [ ] A. Liveness - process alive only
- [ ] B. Readiness - database connected
- [ ] C. Deep health - database + external dependencies
- [ ] D. Configurable - shallow for k8s probes, deep for monitoring

**Notes**:

---

## Section 6: Risk & Blockers (Q33-35)

### Q33. Phase 1 Blockers

What could block Phase 1 completion? (Select all that apply)

- [ ] A. UI complexity - HTML/CSS/JS learning curve
- [ ] B. Token-user association - database schema changes
- [ ] C. Secret hashing migration - breaking change for existing secrets
- [ ] D. Rate limiting complexity - need distributed rate limiter
- [ ] E. Audit logging volume - storage and performance concerns

**Notes**:

---

### Q34. Technical Debt Tolerance

How much technical debt is acceptable during Phase 1?

- [ ] A. Zero - all code must be production-quality
- [ ] B. Low - minimal TODOs, tracked in PROJECT-STATUS.md
- [ ] C. Medium - TODOs acceptable, fix in Phase 5
- [ ] D. High - ship features fast, refactor later

**Notes**:

---

### Q35. Go-Live Criteria

What criteria must be met before "production ready"?

- [ ] A. All tests pass, coverage targets met
- [ ] B. Above + security review completed
- [ ] C. Above + performance benchmarks pass
- [ ] D. Above + external audit (penetration test)
- [ ] E. Above + documentation complete + runbooks

**Notes**:

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
