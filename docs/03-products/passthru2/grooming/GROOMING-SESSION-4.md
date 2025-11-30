# Passthru2 Grooming Session 4: Edge Cases & Final Details

**Purpose**: Final edge cases and implementation details based on Sessions 1-3 decisions.
**Created**: 2025-11-30
**Status**: AWAITING ANSWERS

---

## Section 1: TLS Infrastructure Details (Q1-5)

### Q1. TLS Package Dependencies

What external dependencies should `internal/infra/tls/` use?

- [x] A. Standard library only (crypto/tls, crypto/x509)
- [x] B. + `golang.org/x/crypto` for additional algorithms
- [ ] C. + `github.com/go-acme/lego` for ACME support (future)
- [x] D. Minimal now, extend as needed

Notes:

---

### Q2. Certificate Storage Format

How should generated certificates be stored/loaded?

- [x] A. PEM files on disk
- [x] B. PKCS#12 bundles
- [ ] C. In-memory only (regenerate on restart)
- [x] D. Configurable (PEM default, PKCS#12 optional)

Notes:
PKCS#11 and YubiKey support needed in future too

---

### Q3. Root CA Trust

Should demo mode use a custom root CA or system trust store?

- [x] A. Custom root CA only (isolated demo)
- [ ] B. System trust store (for real certs)
- [ ] C. Both (configurable)
- [ ] D. Custom root CA with option to add to system store

Notes:
99% always custom CAs
Maybe in future I would support adding system trust store for HTTPS Server front-end UI and CLI certs

---

### Q4. Certificate Validation Strictness

How strict should certificate validation be in demo mode?

- [x] A. Full validation (hostname, expiry, chain)
- [ ] B. Relaxed (skip hostname check for localhost)
- [ ] C. Skip validation entirely in demo (dangerous)
- [ ] D. Configurable validation level

Notes:
CRITICAL ALWAYS FULL VALIDATION

---

### Q5. TLS Version Requirements

What minimum TLS version should be required?

- [ ] A. TLS 1.2 (broad compatibility)
- [x] B. TLS 1.3 only (best security)
- [ ] C. TLS 1.2 minimum, prefer 1.3
- [ ] D. Configurable with secure defaults

Notes:

---

## Section 2: UUIDv4 Tenant ID Implementation (Q6-10)

### Q6. UUIDv4 Generation Source

How should UUIDv4 tenant IDs be generated?

- [ ] A. `crypto/rand` directly
- [ ] B. `github.com/google/uuid` NewRandom()
- [x] C. Match existing UUIDv7 generation pattern but v4
- [ ] D. Use existing project UUID utilities

Notes:

---

### Q7. Tenant ID Validation

How should tenant IDs be validated on input?

- [x] A. Strict UUID format only
- [ ] B. Accept with/without hyphens
- [ ] C. Case-insensitive
- [ ] D. All of the above

Notes:

---

### Q8. Tenant ID Display Format

How should tenant IDs be displayed in logs/responses?

- [x] A. Full UUID with hyphens
- [ ] B. Short form (first 8 chars)
- [ ] C. Configurable (full/short)
- [ ] D. Full in responses, short in logs

Notes:

---

### Q9. Demo Tenant IDs

Should demo tenants have predictable or random UUIDs?

- [ ] A. Predictable (e.g., well-known UUIDs for demo)
- [x] B. Random (regenerated each startup)
- [ ] C. Predictable with option to randomize
- [ ] D. Fixed UUIDs documented in demo docs

Notes:

---

### Q10. Tenant ID in URLs

Should tenant ID be in URL path or header?

- [ ] A. Path: `/api/v1/tenants/{tenant_id}/keys`
- [ ] B. Header: `X-Tenant-ID`
- [ ] C. Query param: `?tenant_id=`
- [ ] D. Configurable (default path)

Notes:
NEVER PATH
NEVER QUERY PARAMETERS
ALWAYS HEADER: ALWAYS linked to HTTP "Authorization" header

- Service APIs: Configurable ;
- User UI/APIs: Session UUIDv7 cookie mapped to server-side session in Redis cache
- Basic
Configurable options in Authz Provider:
- Bearer: issued UUID access token (UUIDv7 or UUIDv4), statefully mapped by issuer to tenant UUIDv4
- Bearer: issued JWT  access token, statelessly mapped by issuer by tenant UUIDv4
Configurable options in KMS File/Database realms:
- Basic: File-realm username/password Base64URL encoded
- Basic: Database-realm username/password Base64URL encoded
- Bearer: Federated to Identity (stateless if JWT access token, stateful if UUID access token)
- TLS Client: Custom SAN extension (uri? other? need to consider options...)

---

## Section 3: Demo CLI Error Handling (Q11-15)

### Q11. Error Aggregation

How should multiple errors be collected and reported?

- [ ] A. Simple list of error messages
- [x] B. Structured error with step/phase info
- [x] C. Error tree (nested errors with context)
- [ ] D. Match existing apperr patterns

Notes:

---

### Q12. Partial Success Handling

If some demo steps succeed but others fail, what should happen?

- [x] A. Report partial success, cleanup successful parts
- [x] B. Leave successful parts running
- [ ] C. Rollback all on any failure
- [x] D. Configurable behavior

Notes:

---

### Q13. Retry Configuration

Should demo CLI support retries for transient failures?

- [ ] A. No retries (fail fast)
- [ ] B. Fixed retry count with delay
- [ ] C. Exponential backoff
- [x] D. Configurable retry strategy

Notes:

---

### Q14. Progress Indication

How should demo CLI show progress?

- [x] A. Simple step counter (1/5, 2/5...)
- [ ] B. Progress bar
- [x] C. Spinner with step description
- [ ] D. Configurable (spinner default)

Notes:

---

### Q15. Exit Codes

What exit code strategy should demo CLI use?

- [ ] A. 0 success, 1 failure (simple)
- [ ] B. Different codes for different failure types
- [x] C. Match sysexits.h conventions
- [x] D. 0 success, 1 partial, 2 complete failure

Notes:
I don't know what is sysexits.h, so I can't decide for sure. Maybe D unless C is better.

---

## Section 4: Benchmark & Coverage Details (Q16-20)

### Q16. Benchmark Baseline

Should benchmarks track baseline/regression?

- [ ] A. No baseline (just current numbers)
- [ ] B. Store baseline in git
- [x] C. Compare against previous run
- [x] D. CI-based regression detection

Notes:
Store baseline in untracked local directory

---

### Q17. Coverage Exclusions

What code should be excluded from coverage requirements?

- [ ] A. Generated code only (api/client, api/server)
- [ ] B. + Test utilities
- [ ] C. + Demo/example code
- [ ] D. Explicit exclusion file

Notes:

---

### Q18. Coverage Trend Storage

Where should coverage trends be stored?

- [ ] A. Git history only (compare commits)
- [ ] B. Dedicated coverage file
- [ ] C. External service (Codecov/Coveralls)
- [ ] D. All of the above

Notes:

---

### Q19. Test Fixture Management

How should large test fixtures be managed?

- [ ] A. Inline in test files
- [ ] B. testdata/ directories
- [ ] C. Generated at test time
- [ ] D. External fixture files with embed

Notes:

---

### Q20. Integration Test Isolation

How should integration tests be isolated from each other?

- [ ] A. Unique database per test
- [ ] B. Unique table prefix per test
- [ ] C. Transaction rollback (where possible)
- [ ] D. UUIDv7 prefix in all test data

Notes:

---

## Section 5: Configuration & Deployment (Q21-25)

### Q21. Config File Format

What config file format should be primary?

- [ ] A. YAML only
- [ ] B. YAML with JSON fallback
- [ ] C. YAML, JSON, TOML all supported
- [ ] D. YAML primary, others via converter

Notes:

---

### Q22. Config Validation

When should config be validated?

- [ ] A. At load time only
- [ ] B. At startup (after load, before use)
- [ ] C. Continuous validation (on change)
- [ ] D. Load + startup validation

Notes:

---

### Q23. Default Config Location

Where should default config files be located?

- [ ] A. Current working directory
- [ ] B. Executable directory
- [ ] C. Standard paths (/etc, ~/.config)
- [ ] D. Configurable with sensible search order

Notes:

---

### Q24. Config Hot Reload

Should config support hot reload?

- [ ] A. No (restart required)
- [ ] B. Yes for all settings
- [ ] C. Yes for non-critical settings only
- [ ] D. Not needed for passthru2 (defer)

Notes:

---

### Q25. Docker Compose Profiles

What compose profiles should be defined?

- [ ] A. dev, demo, ci (from Session 1)
- [ ] B. + prod (production template)
- [ ] C. + minimal (single service)
- [ ] D. dev, demo, ci only (keep simple)

Notes:

---

**Status**: AWAITING YOUR ANSWERS (Change [ ] to [x] as applicable and add notes if needed)
