# Products Plan Grooming Session 2

**Purpose**: Deep-dive questions based on Grooming Session 1 analysis
**Created**: November 29, 2025
**Status**: AWAITING ANSWERS

---

## Instructions

Select your answer by changing `[ ]` to `[x]` for each question.

---

## Section 1: Demonstrability & Developer Experience (Q1-5)

### Q1. Demo Format

What does "easy to demonstrate" look like to you?

- [X] Web UI with login flow (visual, impressive)
- [ ] CLI commands with JSON output (scriptable, clear)
- [ ] Swagger UI interactive API (already exists, just needs polish)
- [ ] Video/GIF walkthrough in README
- [ ] Jupyter-style notebook with runnable examples
- [ ] All of the above

**Notes**:

---

### Q2. Demo Audience

Who is the primary demo audience?

- [X] Myself - learning/validation that it works
- [ ] Potential open source contributors
- [X] Potential employers/clients (portfolio piece)
- [ ] Conference talks/presentations
- [ ] Technical blog posts/tutorials
- [ ] All of the above

**Notes**:

---

### Q3. One-Command Demo

What should `make demo` or `go run ./cmd/demo` do?

- [X] Start all services (KMS + Identity + deps) in Docker
- [ ] Start minimal services (just Identity authz for OAuth demo)
- [X] Start with sample data pre-loaded
- [ ] Interactive CLI wizard
- [X] Print instructions and URLs to open

**Notes**:

---

### Q4. Demo Data

What sample data should be included for demos?

- [ ] Pre-configured OAuth clients (demo-client, test-client)
- [X] Sample users with passwords
- [X] Sample keys in KMS
- [X] Sample certificates
- [X] Sample tokens (pre-generated JWTs)
- [ ] None - always start clean

**Notes**:

---

### Q5. Demo Documentation

Where should demo documentation live? Nowhere. It should be 100% intuitive when I open first UI link

- [ ] README.md "Quick Start" section
- [ ] Separate DEMO.md file
- [ ] docs/GETTING-STARTED.md
- [ ] Interactive docs in Swagger UI
- [ ] Video linked from README

**Notes**:

---

## Section 2: KMS MVP Definition (Q6-10)

### Q6. KMS Minimum Scope

What's the absolute minimum viable KMS?

- [X] Key generation only (generate and return key material)
- [X] Key generation + persistent storage (generate, store, retrieve)
- [ ] Above + key rotation (versioned keys)
- [X] Above + key hierarchy (root → intermediate → content)
- [ ] Above + policy/access control
- [ ] Full feature set (all above + HSM integration)

**Notes**:

---

### Q7. Key Types for MVP

Which key types are MUST-HAVE for KMS MVP?

**Asymmetric:**

- [X] RSA-2048 (signing, encryption)
- [ ] RSA-4096 (signing, encryption)
- [X] ECDSA P-256 (signing)
- [ ] ECDSA P-384 (signing)
- [X] ECDH P-256 (key agreement)
- [X] Ed25519 (signing)

**Symmetric:**

- [ ] AES-128-GCM (encryption)
- [X] AES-256-GCM (encryption)
- [ ] HMAC-SHA256 (MAC)
- [X] HMAC-SHA512 (MAC)

**Notes**:

---

### Q8. KMS API Style

What API style for KMS?

- [X] REST only (OpenAPI-generated)
- [ ] gRPC only (protobuf)
- [ ] REST + gRPC (both)
- [X] CLI-first (command line primary interface)
- [X] Library-first (Go package for embedding)

**Notes**:

---

### Q9. KMS Storage Backend

What storage backend for MVP?

- [ ] SQLite only (simple, portable)
- [ ] PostgreSQL only (production-ready)
- [X] Both SQLite and PostgreSQL (configurable)
- [ ] In-memory only (for demos/testing)
- [ ] Pluggable (interface with multiple implementations)

**Notes**:

---

### Q10. KMS Key Hierarchy

How should key hierarchy work?

- [ ] Flat (all keys at same level)
- [ ] Two-tier (root keys → content keys)
- [X] Three-tier (root → intermediate → content)
- [ ] N-tier (unlimited nesting)
- [ ] Configurable (admin chooses depth)

**Notes**:

---

## Section 3: Identity Integration (Q11-15)

### Q11. Identity Auth for KMS

When KMS uses Identity for authentication, what's the primary auth model?

- [ ] API keys (simple, static secrets)
- [ ] OAuth 2.1 client credentials (service-to-service)
- [ ] OAuth 2.1 authorization code (user context)
- [ ] mTLS (certificate-based mutual auth)
- [X] Multiple options (configurable per deployment)

**Notes**:

---

### Q12. Identity Embedding

Should Identity be embeddable in KMS?

- [X] Yes - KMS can embed Identity for standalone deployment
- [ ] No - always separate services
- [ ] Optional - KMS can use embedded OR external Identity
- [ ] Haven't decided yet

**Notes**:

---

### Q13. Token Format

What token format should Identity issue for KMS auth?

- [ ] JWT access tokens (standard, inspectable)
- [ ] Opaque tokens (reference tokens, requires introspection)
- [X] Both (configurable)
- [ ] Doesn't matter - Identity handles it

**Notes**:

---

### Q14. Identity Scopes for KMS

What OAuth scopes should KMS require?

- [ ] Single scope: `kms:admin` (full access)
- [X] Resource-based: `kms:keys:read`, `kms:keys:write`, etc.
- [X] Key-specific: `kms:key:{keyId}:read`
- [X] Hierarchical: `kms:*`, `kms:keys:*`, `kms:keys:{id}:*`
- [ ] No scopes - use separate RBAC system

**Notes**:

---

### Q15. Session Management

Does KMS need session management (from Identity)?

- [X] Yes - stateful sessions with refresh
- [ ] No - stateless JWT validation only
- [ ] Optional - support both modes
- [ ] Haven't thought about this

**Notes**:

---

## Section 4: JOSE Product Clarification (Q16-18)

### Q16. JOSE Standalone Value

What's the standalone value of JOSE as a product?

- [ ] CLI tool for generating/validating JWTs
- [X] Library for other Go projects to import
- [ ] HTTP service for JWT operations
- [ ] All of the above
- [ ] Reconsider - maybe JOSE should be infrastructure

**Notes**:

---

### Q17. JOSE CLI Commands

What CLI commands should JOSE product have?

- [X] `jose jwk generate` - generate JWK/JWKS
- [X] `jose jwt sign` - sign JWT with key
- [X] `jose jwt verify` - verify JWT signature
- [X] `jose jwt decode` - decode and display JWT
- [X] `jose jwe encrypt` - encrypt with JWE
- [X] `jose jwe decrypt` - decrypt JWE
- [X] All of the above

**Notes**:

---

### Q18. JOSE vs Identity Relationship

How does JOSE relate to Identity?

- [ ] Identity imports JOSE library (JOSE is dependency)
- [ ] JOSE and Identity are independent
- [ ] JOSE can use Identity for key storage
- [X] Identity IS the JWT issuer (JOSE is just library)

**Notes**:

---

## Section 5: Certificates Product Scope (Q19-22)

### Q19. Certificates MVP

What's the minimum viable Certificates product?

- [ ] X.509 certificate generation only
- [ ] Above + CA hierarchy (root → intermediate → leaf)
- [X] Above + CSR handling
- [ ] Above + revocation (CRL/OCSP)
- [ ] Above + ACME protocol (Let's Encrypt style)
- [ ] Full PKI suite

**Notes**:

---

### Q20. Certificates Use Cases

Primary use case for Certificates product?

- [X] Internal mTLS (service-to-service)
- [X] TLS server certificates (HTTPS)
- [ ] Code signing certificates
- [X] Client authentication certificates
- [ ] All of the above

**Notes**:

---

### Q21. Certificates + KMS Integration

How should Certificates integrate with KMS?

- [ ] Certificates uses KMS for CA key storage
- [X] Independent (Certificates has own key management)
- [ ] Configurable (can use KMS or standalone)

**Notes**:

---

### Q22. Certificates + Identity Integration

How should Certificates integrate with Identity?

- [ ] Identity issues certificates as credentials
- [X] Certificates uses Identity for admin auth
- [ ] mTLS certificates authenticate to Identity
- [ ] All of the above
- [ ] Independent

**Notes**:
I haven't thought too far ahead here

---

## Section 6: Technical Decisions (Q23-25)

### Q23. Database Schema Strategy

How should database schemas be organized?

- [ ] Single database, shared tables (simple)
- [ ] Single database, separate schemas per product
- [X] Separate databases per product
- [ ] Pluggable (configurable per deployment)

**Notes**:
Separate DB servers instance for prod and preprod
Singe DB server instance for staging and dev, but separate logical DBs
---

### Q24. Configuration Format

What configuration format for all products?

- [ ] YAML only (current approach)
- [ ] JSON only
- [ ] TOML only
- [X] Multiple formats supported
- [ ] Environment variables primary
- [ ] Mix (YAML files + env var overrides)

**Notes**:
Prefer configure over ENV

---

### Q25. Logging Strategy

What logging strategy across products?

- [ ] Structured JSON logs (machine-readable)
- [ ] Human-readable text logs (dev-friendly)
- [ ] Configurable (JSON for prod, text for dev)
- [X] OpenTelemetry native (traces + logs unified)

**Notes**:

---

## Summary Section

### Clarifications Needed

After answering, list any areas still unclear:

1.
2.
3.

### Dependencies Identified

List product/component dependencies discovered:

1.
2.
3.

### Scope Adjustments

List any scope changes from original plan:

1.
2.
3.

---

**Status**: AWAITING ANSWERS
**Next Step**: Complete answers, then request final refined plan
