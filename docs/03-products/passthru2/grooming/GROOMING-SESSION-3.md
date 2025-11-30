# Passthru2 Grooming Session 3: Implementation Specifics

**Purpose**: Final implementation details based on Sessions 1 & 2 decisions.
**Created**: 2025-11-30
**Status**: AWAITING ANSWERS

---

## Section 1: TLS/HTTPS Implementation (Q1-5)

### Q1. KMS Cert Utility Location

Where are the KMS cert utility functions that Identity should reuse?

- [ ] A. `internal/crypto/cert/` (existing location)
- [ ] B. `internal/common/tls/` (needs extraction)
- [x] C. `internal/infra/tls/` (create new)
- [ ] D. Other location: _______________

Notes:

---

### Q2. CA Chain Configuration

What should be the default CA chain length?

- [ ] A. 1 (Root CA → Leaf)
- [ ] B. 2 (Root CA → Intermediate CA → Leaf)
- [ ] C. 3 (Root CA → Policy CA → Issuing CA → Leaf)
- [x] D. Configurable with default of 3

Notes:

---

### Q3. Certificate Common Names

What CN pattern should be used for demo certs?

- [ ] A. Service name only (e.g., `kms`, `identity`)
- [ ] B. Service + domain (e.g., `kms.demo.local`)
- [x] C. FQDN style (e.g., `kms.cryptoutil.demo.local`)
- [x] D. Configurable via config file

Notes:

---

### Q4. TLS Client Certificate Mode

Should mTLS be enabled by default for service-to-service communication?

- [x] A. Yes, mTLS required for all internal communication
- [ ] B. Yes for KMS→Identity, optional for others
- [ ] C. No, TLS server-only by default, mTLS opt-in
- [x] D. Configurable per service pair

Notes:

---

### Q5. Certificate Rotation

How should certificate rotation be handled in demo mode?

- [ ] A. No rotation (certs valid for demo duration)
- [ ] B. Long validity (365 days) with manual rotation
- [ ] C. Auto-rotation with configurable interval
- [x] D. Not needed for passthru2 (defer to passthru3)

Notes:
Auto-rotation with configurable interval later

---

## Section 2: Realm Configuration Details (Q6-10)

### Q6. Realm File Location

Where should `realms.yml` be located relative to main config?

- [x] A. Same directory as main config
- [ ] B. Subdirectory: `config/realms/`
- [ ] C. Separate directory: `realms/`
- [ ] D. Embedded in main config with include directive

Notes:

---

### Q7. Realm Password Hash Format

What PBKDF2 hash format should be used?

- [ ] A. `pbkdf2:sha256:iterations:salt:hash` (explicit)
- [ ] B. `$pbkdf2-sha256$iterations$salt$hash` (Modular Crypt Format)
- [ ] C. Base64-encoded struct with all parameters
- [ ] D. Match existing Identity PBKDF2 format

Notes:
Configurable to use other digest, iterations, salt length in bytes

---

### Q8. Realm User Schema

What fields should realm users have?

- [x] A. Minimal: `username`, `password_hash`, `roles`
- [x] B. Standard: + `tenant_id`, `enabled`, `created_at`
- [x] C. Extended: + `email`, `display_name`, `metadata`
- [x] D. Flexible: Core fields + extensible JSON metadata

Notes:

---

### Q9. Role Definition

How should KMS roles be defined?

- [ ] A. Hardcoded in code (admin, tenant-admin, user, service)
- [x] B. Configurable in `realms.yml`
- [x] C. Hierarchical with inheritance
- [x] D. Mapped from Identity roles when federated

Notes:

---

### Q10. Tenant ID Format

What format should tenant IDs use?

- [ ] A. UUIDv7 (consistent with other entities)
- [ ] B. Slug/string (e.g., `tenant-acme`)
- [ ] C. Both (UUID primary, slug alias)
- [ ] D. Configurable per deployment

Notes:
UUIDv4, this is a special case where maximum randomness is preferred for external facing tenant IDs,
not predictable

---

## Section 3: Demo CLI Implementation (Q11-15)

### Q11. Demo CLI Architecture

How should demo CLIs be structured?

- [x] A. Single binary with subcommands (`demo kms`, `demo identity`, `demo all`)
- [ ] B. Separate binaries (`demo-kms`, `demo-identity`, `demo-all`)
- [x] C. Library with thin CLI wrappers
- [ ] D. Cobra-based with shared utilities

Notes:

---

### Q12. Demo CLI Output Format

What output format should demo CLIs use?

- [x] A. Human-readable with colors/emojis
- [x] B. JSON for machine parsing
- [x] C. Both (default human, `--json` flag)
- [x] D. Structured logging format (compatible with OTLP)

Notes:

---

### Q13. Demo Flow Execution

How should demo flows handle failures?

- [ ] A. Stop on first error with detailed message
- [x] B. Continue on error, report summary at end
- [ ] C. Retry with backoff, then fail
- [x] D. Configurable behavior via flag

Notes:

---

### Q14. Health Check Waiting

How long should demo CLI wait for services to be healthy?

- [ ] A. Fixed timeout (60 seconds)
- [x] B. Configurable timeout with default (30 seconds)
- [ ] C. Infinite wait with progress indicator
- [ ] D. Exponential backoff with max attempts

Notes:

---

### Q15. Demo Data Verification

Should demo CLI verify seeded data after startup?

- [x] A. Yes, query and validate all demo entities
- [ ] B. Yes, but only critical entities (users, clients)
- [ ] C. No, rely on health checks only
- [ ] D. Configurable via `--verify` flag

Notes:

---

## Section 4: Token Validation Implementation (Q16-20)

### Q16. JWKS Cache Implementation

What caching library should be used for JWKS?

- [ ] A. Simple map with mutex (no external deps)
- [ ] B. `patrickmn/go-cache` (popular, simple)
- [ ] C. `dgraph-io/ristretto` (high performance)
- [ ] D. Custom implementation matching existing patterns

Notes:
No preference, I don't know which one to use

---

### Q17. Token Introspection Batching

Should introspection requests be batched?

- [X] A. No, one introspection per token per request
- [X] B. Yes, batch multiple tokens in single request
- [X] C. Yes, with request deduplication
- [ ] D. Not needed for passthru2 (defer optimization)

Notes:
Both

---

### Q18. Error Response Structure

What structure should token validation errors use?

- [ ] A. RFC 6749 OAuth 2.0 error response
- [ ] B. RFC 7807 Problem Details for HTTP APIs
- [ ] C. Custom error structure matching existing apperr
- [X] D. Hybrid (OAuth errors for auth, Problem Details for others)

Notes:

---

### Q19. Scope Parsing

How should scope strings be parsed and validated?

- [ ] A. Simple string split on space
- [X] B. Structured parser with validation
- [ ] C. Regex-based with format enforcement
- [ ] D. Match Identity's scope parsing implementation

Notes:

---

### Q20. Claims Context Propagation

How should extracted claims be propagated through the request context?

- [X] A. Custom context key with typed struct
- [X] B. Standard OIDC claims struct
- [ ] C. Map[string]interface{} for flexibility
- [ ] D. Protobuf-style generated types

Notes:

---

## Section 5: Testing & Quality (Q21-25)

### Q21. Test Data Factory Pattern

Should passthru2 use a factory pattern for test data?

- [X] A. Yes, dedicated `testutil` package with factories
- [X] B. Yes, per-package test helpers
- [ ] C. No, inline test data generation
- [ ] D. Combination (factories for complex, inline for simple)

Notes:

---

### Q22. Benchmark Targets

What operations should have benchmarks?

- [X] A. Crypto operations only (encrypt, decrypt, sign)
- [X] B. + Token validation (JWT parsing, introspection)
- [X] C. + Database operations (CRUD)
- [X] D. All public API endpoints

Notes:

---

### Q23. E2E Test Parallelization

Should E2E tests run in parallel?

- [X] A. No, sequential for predictability
- [ ] B. Yes, with proper isolation (UUIDv7 prefixes)
- [X] C. Configurable via test flag
- [X] D. Parallel for independent flows, sequential for dependent

Notes:

---

### Q24. Test Coverage Reporting

What coverage reporting format should be used?

- [ ] A. Go native coverage profile
- [ ] B. HTML report for local review
- [ ] C. Codecov/Coveralls integration
- [X] D. All of the above

Notes:

---

### Q25. Integration Test Timeout

What should be the default timeout for integration tests?

- [ ] A. 30 seconds per test
- [X] B. 60 seconds per test
- [ ] C. 5 minutes for full suite
- [X] D. Configurable with sensible defaults

Notes:

---

**Status**: ✅ COMPLETED (2025-11-30)
