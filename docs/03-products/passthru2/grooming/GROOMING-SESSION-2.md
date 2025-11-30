# Passthru2 Grooming Session 2: Technical Deep Dive

**Purpose**: Detailed technical questions to finalize implementation decisions for `passthru2`.
**Created**: 2025-11-30
**Status**: AWAITING ANSWERS

---

## Section 1: KMS Realm Implementation (Q1-5)

### Q1. File Realm Configuration Format

What format should the file realm configuration use?

- [ ] A. YAML embedded in main config file
- [ ] B. Separate YAML file (e.g., `realms.yml`)
- [ ] C. JSON file
- [ ] D. Go struct with hardcoded demo values only

Notes:

---

### Q2. File Realm Password Storage

How should file realm passwords be stored?

- [ ] A. Plaintext (demo only, documented as insecure)
- [ ] B. PBKDF2 hashed (consistent with Identity)
- [ ] C. Pre-hashed at config load time
- [ ] D. Support both plaintext (demo) and hashed (prod)

Notes:

---

### Q3. DB Realm Schema Design

For PostgreSQL mode, where should realm users be stored?

- [ ] A. New `kms_realm_users` table (separate from Identity)
- [ ] B. Shared table with Identity users (requires Identity DB)
- [ ] C. Inline in KMS config table (JSON blob)
- [ ] D. DB realm not needed for passthru2 (file realm only)

Notes:

---

### Q4. Realm Priority Order

When multiple realms are enabled, what authentication order should be used?

- [ ] A. File realm first, then DB realm, then federation
- [ ] B. Federation first (external identity authorities)
- [ ] C. Configurable order via config file
- [ ] D. Only one realm can be enabled at a time

Notes:

---

### Q5. Tenant Isolation in KMS

How should tenant isolation be enforced in KMS?

- [ ] A. Database-level isolation (separate schemas/databases)
- [ ] B. Row-level isolation (tenant_id column on all tables)
- [ ] C. Key-level isolation (keys belong to tenants)
- [ ] D. Combination of row-level + key-level

Notes:

---

## Section 2: Token Validation Details (Q6-10)

### Q6. JWKS Caching Strategy

How should KMS cache Identity JWKS?

- [ ] A. In-memory with configurable TTL
- [ ] B. Redis/external cache for distributed KMS
- [ ] C. File-based cache (for single-instance)
- [ ] D. No caching (always fetch on validation)

Notes:

---

### Q7. Token Revocation Check Frequency

How often should KMS check token revocation via introspection?

- [ ] A. Every request (most secure, slowest)
- [ ] B. Only for sensitive operations (encrypt, sign, unwrap)
- [ ] C. Configurable interval (e.g., every 5 minutes)
- [ ] D. Never (rely on token expiry only)

Notes:

---

### Q8. Token Validation Failure Handling

What should KMS return when token validation fails?

- [ ] A. 401 Unauthorized with generic message
- [ ] B. 401 with detailed error (expired, invalid signature, etc.)
- [ ] C. 403 Forbidden for scope issues, 401 for auth issues
- [ ] D. Configurable error detail level

Notes:

---

### Q9. Service-to-Service Authentication

How should KMS authenticate when calling Identity introspection?

- [ ] A. Client credentials (KMS has its own client in Identity)
- [ ] B. mTLS (certificate-based auth)
- [ ] C. API key / static token
- [ ] D. No auth (internal network trust)

Notes:

---

### Q10. Token Claims Extraction

Which token claims should KMS extract and use?

- [ ] A. `sub` (subject/user ID) only
- [ ] B. `sub` + `scope` (for authorization)
- [ ] C. `sub` + `scope` + `tenant_id` (custom claim)
- [ ] D. All standard OIDC claims + custom claims

Notes:

---

## Section 3: Demo Data Details (Q11-15)

### Q11. Demo Key Material

Should demo mode create actual cryptographic keys or placeholders?

- [ ] A. Real keys (fully functional demo)
- [ ] B. Placeholder keys (metadata only, no crypto operations)
- [ ] C. Real keys with warning labels (demo-key-*)
- [ ] D. Configurable (real for local, placeholder for CI)

Notes:

---

### Q12. Demo Data Persistence

Should demo data persist across restarts?

- [ ] A. Yes (SQLite file / Postgres)
- [ ] B. No (in-memory only, re-seeded on restart)
- [ ] C. Configurable via flag
- [ ] D. Different per profile (dev=persist, ci=ephemeral)

Notes:

---

### Q13. Demo User Passwords

What passwords should demo users have?

- [ ] A. Predictable (e.g., `demo-admin-password`)
- [ ] B. Generated and logged on startup
- [ ] C. Same as username (e.g., admin/admin)
- [ ] D. Documented in config file with clear warnings

Notes:

---

### Q14. Demo Client Secrets

For confidential demo clients, how should secrets be handled?

- [ ] A. Predictable secrets documented in demo docs
- [ ] B. Generated and logged on startup
- [ ] C. Stored in Docker secrets even for demo
- [ ] D. No confidential clients in demo (public only)

Notes:

---

### Q15. Demo Data Cleanup

Should there be a way to reset demo data?

- [ ] A. Yes, `--reset-demo` flag
- [ ] B. Yes, DELETE endpoint (admin only)
- [ ] C. No (restart service to reset)
- [ ] D. Not needed (demo data is ephemeral)

Notes:

---

## Section 4: Compose & Deployment (Q16-20)

### Q16. Health Check Implementation

What should health checks verify?

- [ ] A. HTTP 200 from `/livez` only
- [ ] B. `/livez` + `/readyz` (separate liveness/readiness)
- [ ] C. Database connectivity + service health
- [ ] D. Full dependency chain (DB + Identity + Telemetry)

Notes:

---

### Q17. Compose Network Architecture

How should Docker Compose networks be structured?

- [ ] A. Single shared network for all services
- [ ] B. Per-product networks + shared telemetry network
- [ ] C. Frontend/backend network separation
- [ ] D. Service mesh style (each service isolated)

Notes:

---

### Q18. Volume Strategy

How should persistent data be handled in Compose?

- [ ] A. Named volumes for all data
- [ ] B. Bind mounts for development, volumes for demo
- [ ] C. No persistence (ephemeral containers)
- [ ] D. Configurable per profile

Notes:

---

### Q19. Port Allocation

What port scheme should be used?

- [ ] A. Standard ports (8080 for all, different hosts)
- [ ] B. Product-specific ports (KMS=8081, Identity=8082)
- [ ] C. Dynamic ports (let Docker assign)
- [ ] D. Configurable via environment variables

Notes:

---

### Q20. TLS in Demo Mode

Should demo mode use TLS?

- [ ] A. Yes, self-signed certs (same as production)
- [ ] B. No TLS for simplicity
- [ ] C. Optional via flag
- [ ] D. TLS for external access, plain HTTP internally

Notes:

---

## Section 5: Testing Strategy (Q21-25)

### Q21. Demo Integration Test Scope

What should demo integration tests verify?

- [ ] A. Service startup and health checks only
- [ ] B. Basic CRUD operations with demo data
- [ ] C. Full demo flow (login → operation → logout)
- [ ] D. All of the above

Notes:

---

### Q22. E2E Test Environment

Where should E2E tests run?

- [ ] A. Docker Compose (same as demo)
- [ ] B. Kubernetes (minikube/kind)
- [ ] C. Testcontainers in Go tests
- [ ] D. Combination (Compose for local, Testcontainers for CI)

Notes:

---

### Q23. Test Data Isolation

How should tests be isolated from each other?

- [ ] A. Unique prefixes per test (demo-test-123-*)
- [ ] B. Separate databases per test
- [ ] C. Transaction rollback (if supported)
- [ ] D. Cleanup after each test

Notes:

---

### Q24. Performance Testing Scope

Should passthru2 include performance tests?

- [ ] A. Yes, basic benchmarks for critical paths
- [ ] B. Yes, load tests with configurable concurrency
- [ ] C. No, defer to passthru3
- [ ] D. Only if time permits

Notes:

---

### Q25. Test Documentation

What test documentation is required?

- [ ] A. Test coverage reports only
- [ ] B. Test case descriptions in code
- [ ] C. Separate test plan document
- [ ] D. All of the above

Notes:

---

**Status**: AWAITING YOUR ANSWERS (Change [ ] to [x] as applicable and add notes if needed)
