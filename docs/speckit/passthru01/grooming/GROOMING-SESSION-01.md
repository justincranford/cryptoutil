# Grooming Session 01: Core Architecture and FIPS Compliance

## Overview

- **Focus Area**: cryptoutil core architecture, FIPS 140-3 compliance, product structure
- **Related Spec Section**: Constitution Principles I-V, Spec Products P1-P4
- **Prerequisites**: Understanding of cryptographic standards, Go project structure

---

## Questions

### Q1: Which password hashing algorithm is FIPS 140-3 approved and required for cryptoutil?

A) bcrypt with work factor 12
B) Argon2id with memory cost 64MB
C) PBKDF2-HMAC-SHA256
D) scrypt with N=2^14

**Answer**: C
**Explanation**: FIPS 140-3 only approves PBKDF2-HMAC-SHA256 for password hashing. bcrypt, Argon2, and scrypt are NOT FIPS-approved and must NEVER be used in cryptoutil.

---

### Q2: What is the minimum RSA key size allowed under FIPS 140-3 for cryptoutil?

A) 1024 bits
B) 2048 bits
C) 3072 bits
D) 4096 bits

**Answer**: B
**Explanation**: FIPS 140-3 requires RSA keys to be at least 2048 bits. While 3072 and 4096 are also approved, 2048 is the minimum. 1024-bit keys are explicitly prohibited.

---

### Q3: In the cryptoutil key hierarchy, what is the correct order from highest to lowest?

A) Content keys → Intermediate keys → Root keys → Unseal secrets
B) Unseal secrets → Root keys → Intermediate keys → Content keys
C) Root keys → Unseal secrets → Content keys → Intermediate keys
D) Intermediate keys → Root keys → Unseal secrets → Content keys

**Answer**: B
**Explanation**: The hierarchy flows from Unseal secrets (file:///run/secrets/*) → Root keys (derived from unseal) → Intermediate keys (per-tenant isolation) → Content keys (actual data protection).

---

### Q4: What coverage target applies to infrastructure (cicd) code?

A) ≥75%
B) ≥80%
C) ≥85%
D) ≥95%

**Answer**: C
**Explanation**: Infrastructure/cicd code requires ≥85% coverage. Production code requires ≥80%, and utility code requires ≥95%.

---

### Q5: Which of these is a valid reason to use a `//nolint:` directive?

A) The linter rule is inconvenient for this code
B) The code is test code, not production code
C) The linter has a documented bug affecting this code
D) The developer prefers a different coding style

**Answer**: C
**Explanation**: The ONLY valid reason for //nolint: directives is documented linter bugs. Convenience, test code status, or style preferences are NEVER valid reasons.

---

### Q6: What is the hard file size limit requiring mandatory refactoring?

A) 200 lines
B) 300 lines
C) 400 lines
D) 500 lines

**Answer**: D
**Explanation**: File size limits are: 300 (soft warning), 400 (medium/review required), 500 (hard limit requiring refactor).

---

### Q7: Which product handles OAuth 2.1 Authorization Server functionality?

A) P1: JOSE
B) P2: Identity
C) P3: KMS
D) P4: Certificates

**Answer**: B
**Explanation**: P2: Identity includes OAuth 2.1 AuthZ, OIDC IdP, MFA, and FIDO2/WebAuthn. JOSE provides the underlying cryptographic primitives but not the authorization server itself.

---

### Q8: What MUST be true about all cryptoutil instances using the same unseal secrets?

A) They must run on the same physical machine
B) They must derive identical JWKs including KIDs
C) They must use different database backends
D) They must have unique configuration files

**Answer**: B
**Explanation**: For cryptographic interoperability, all instances using the same unseal secrets MUST derive the same JWKs, including KIDs and key materials. Different KIDs would break interoperability.

---

### Q9: What is the maximum validity period for subscriber certificates under CA/Browser Forum requirements?

A) 90 days
B) 180 days
C) 398 days
D) 825 days

**Answer**: C
**Explanation**: CA/Browser Forum Baseline Requirements (post-2020-09-01) limit subscriber certificate validity to maximum 398 days.

---

### Q10: Which hash algorithms are explicitly banned in cryptoutil?

A) SHA-256 and SHA-384
B) MD5 and SHA-1
C) SHA-512 and SHA-3
D) BLAKE2 and BLAKE3

**Answer**: B
**Explanation**: MD5 and SHA-1 are NOT FIPS-approved and explicitly banned. SHA-256, SHA-384, SHA-512 are approved. BLAKE algorithms are not FIPS-approved but aren't explicitly listed as banned (they're simply not used).

---

### Q11: What is the correct test file suffix for benchmark tests?

A) `_test.go`
B) `_bench_test.go`
C) `_benchmark_test.go`
D) `_perf_test.go`

**Answer**: B
**Explanation**: Test file conventions: `_test.go` (unit), `_bench_test.go` (bench), `_fuzz_test.go` (fuzz), `_integration_test.go` (integration).

---

### Q12: In table-driven tests, why is `t.Parallel()` mandatory?

A) It makes tests run faster
B) It reduces code duplication
C) It reveals real concurrency bugs
D) It simplifies test maintenance

**Answer**: C
**Explanation**: `t.Parallel()` is a FEATURE that reveals race conditions that sequential tests would hide. Failing parallel tests indicate production bugs to fix, not a reason to remove parallelism.

---

### Q13: Where should secrets be stored in production Docker deployments?

A) Environment variables
B) Docker secrets at `/run/secrets/`
C) Configuration files in the image
D) Command line arguments

**Answer**: B
**Explanation**: NEVER use environment variables for secrets in production. Docker secrets mounted to `/run/secrets/` with file:// URLs are the required approach.

---

### Q14: What is the correct way to handle database DSN in Go code?

A) Use environment variables exclusively
B) Use `localhost` for database DSN
C) Use `127.0.0.1` for database DSN
D) Hard-code the DSN in source code

**Answer**: B
**Explanation**: For Go code database DSN, `localhost` is acceptable. However, for Go server binding addresses, `127.0.0.1` is required.

---

### Q15: Which infrastructure component handles OpenTelemetry instrumentation?

A) I1: Configuration
B) I5: Telemetry
C) I9: Deployment
D) I6: Crypto

**Answer**: B
**Explanation**: I5: Telemetry handles logging, metrics, tracing, monitoring, OpenTelemetry, and Grafana integration.

---

### Q16: What is the minimum bits required for certificate serial numbers?

A) 32 bits
B) 64 bits
C) 128 bits
D) 256 bits

**Answer**: B
**Explanation**: CA/Browser Forum requires certificate serial numbers have minimum 64 bits from CSPRNG, be non-sequential, >0, and <2^159.

---

### Q17: What status indicates a feature is implemented but has issues?

A) ✅
B) ⚠️
C) ❌
D) 🔄

**Answer**: B
**Explanation**: In spec status indicators: ✅ = fully working, ⚠️ = partial/issues, ❌ = not implemented.

---

### Q18: Which Go version is required for cryptoutil?

A) 1.21+
B) 1.22+
C) 1.23+
D) 1.25.4+

**Answer**: D
**Explanation**: cryptoutil requires Go 1.25.4+ as specified in the version requirements.

---

### Q19: What must happen before any task can be marked complete?

A) Code review approval
B) Manager sign-off
C) Objective, verifiable evidence
D) Documentation update only

**Answer**: C
**Explanation**: Evidence-Based Task Completion requires objective evidence: clean build, clean lint, test pass, coverage met, and documentation updated.

---

### Q20: Which PKCE code challenge method is required for OAuth 2.1?

A) plain
B) S256
C) S384
D) S512

**Answer**: B
**Explanation**: OAuth 2.1 requires PKCE with S256 (SHA-256) code challenge method. The "plain" method is explicitly prohibited.

---

### Q21: What is the port for the cryptoutil-sqlite service public API?

A) 8080
B) 8081
C) 8082
D) 9090

**Answer**: A
**Explanation**: cryptoutil-sqlite uses port 8080 for public API. PostgreSQL instances use 8081/8082. Admin API uses 9090.

---

### Q22: Which tool should be used for Go code formatting?

A) gofmt
B) gofumpt
C) goimports
D) go fmt

**Answer**: B
**Explanation**: gofumpt (not gofmt) is the required formatter. It's a stricter superset of gofmt.

---

### Q23: What is the correct import alias convention for cryptoutil packages?

A) `cu<PackageName>` (e.g., `cuMagic`)
B) `cryptoutil<PackageName>` (e.g., `cryptoutilMagic`)
C) `<PackageName>` (no alias)
D) `pkg<PackageName>` (e.g., `pkgMagic`)

**Answer**: B
**Explanation**: ALL `cryptoutil/**` imports MUST use camelCase aliases with "cryptoutil" prefix (e.g., `cryptoutilMagic`, `cryptoutilConfig`).

---

### Q24: Which authentication method is NOT supported in cryptoutil Identity?

A) client_secret_basic
B) client_secret_post
C) private_key_jwt
D) client_secret_jwt

**Answer**: C
**Explanation**: private_key_jwt is ❌ Not Implemented. client_secret_basic and client_secret_post are ✅ Working. client_secret_jwt is ⚠️ Not Tested.

---

### Q25: What is the correct approach for test data values?

A) Hardcode UUID strings for consistency
B) Use magic package constants OR runtime-generated UUIDv7
C) Generate new UUIDv7 for each assertion
D) Use sequential integers starting from 1

**Answer**: B
**Explanation**: NEVER hardcode test values. Use magic package constants OR generate UUIDv7 once and reuse. Calling NewV7() twice expecting same result is wrong.

---

### Q26: What encoding must ALL text files use?

A) UTF-16 with BOM
B) UTF-8 with BOM
C) UTF-8 without BOM
D) ASCII

**Answer**: C
**Explanation**: UTF-8 without BOM is required for ALL text files, enforced by `cicd all-enforce-utf8`.

---

### Q27: Which curve types are approved for EC keys in FIPS 140-3?

A) secp256k1, secp384r1, secp521r1
B) P-256, P-384, P-521 (NIST curves)
C) Curve25519, Curve448
D) brainpoolP256r1, brainpoolP384r1

**Answer**: B
**Explanation**: FIPS 140-3 approves NIST curves: P-256, P-384, P-521. secp256k1 (Bitcoin curve), Curve25519, and brainpool curves are NOT FIPS-approved.

---

### Q28: What is the admin API health endpoint for liveness probes?

A) `/health`
B) `/livez`
C) `/readyz`
D) `/status`

**Answer**: B
**Explanation**: Admin API at port 9090 exposes `/livez` for liveness and `/readyz` for readiness probes.

---

### Q29: Which database backend is used for development/testing?

A) PostgreSQL only
B) SQLite only
C) Both PostgreSQL and SQLite
D) MySQL

**Answer**: C
**Explanation**: SQLite is used for development/testing (in-memory), PostgreSQL for production. Both must be supported with identical behavior.

---

### Q30: What should be the MaxOpenConns setting for SQLite with GORM?

A) 1
B) 5
C) 10
D) Unlimited

**Answer**: B
**Explanation**: For GORM with SQLite, MaxOpenConns should be 5 to allow transactions (which need multiple connections). Setting to 1 causes deadlocks with explicit transactions.

---

### Q31: What is the TLS minimum version required?

A) TLS 1.0
B) TLS 1.1
C) TLS 1.2
D) TLS 1.3

**Answer**: C
**Explanation**: TLS 1.2 is the minimum version. Full cert chain validation is required, and InsecureSkipVerify must NEVER be used.

---

### Q32: Which tool is used for OpenAPI code generation?

A) swagger-codegen
B) openapi-generator
C) oapi-codegen
D) go-swagger

**Answer**: C
**Explanation**: oapi-codegen with strict server pattern is the required tool for OpenAPI code generation.

---

### Q33: What commit message format is required?

A) GitHub style (imperative with issue reference)
B) Conventional Commits format
C) JIRA ticket prefix
D) Free-form descriptive

**Answer**: B
**Explanation**: Conventional Commits format is required: `<type>[optional scope]: <description>`. Types include feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.

---

### Q34: What is the golangci-lint version requirement?

A) v1.50.0+
B) v1.55.0+
C) v2.0.0+
D) v2.6.2+

**Answer**: D
**Explanation**: golangci-lint v2.6.2+ is required, with specific v2 configuration changes (wsl_v5 config key, removed settings).

---

### Q35: Which MFA factor is fully implemented and working?

A) Email OTP
B) SMS OTP
C) TOTP
D) All of the above

**Answer**: C
**Explanation**: TOTP is ✅ Working. Passkey/WebAuthn is ⚠️ Partial. Email OTP and SMS OTP are ❌ Not Implemented.

---

### Q36: What is the purpose of the `.specify/memory/constitution.md` file?

A) Store project configuration
B) Define immutable project principles
C) Track implementation progress
D) Document API endpoints

**Answer**: B
**Explanation**: The constitution defines foundational, immutable project principles that should NEVER be modified without explicit stakeholder approval.

---

### Q37: What is the correct PRAGMA for SQLite concurrent writes?

A) `PRAGMA journal_mode=DELETE`
B) `PRAGMA journal_mode=WAL`
C) `PRAGMA journal_mode=MEMORY`
D) `PRAGMA journal_mode=TRUNCATE`

**Answer**: B
**Explanation**: WAL (Write-Ahead Logging) mode enables better concurrency with multiple readers and one writer.

---

### Q38: Which product is currently in PLANNING status?

A) P1: JOSE
B) P2: Identity
C) P3: KMS
D) P4: Certificates

**Answer**: D
**Explanation**: P4: Certificates is planned with 20 tasks defined in docs/05-ca/README.md but not yet implemented.

---

### Q39: What is the busy_timeout PRAGMA value for SQLite?

A) 5000 (5 seconds)
B) 10000 (10 seconds)
C) 30000 (30 seconds)
D) 60000 (60 seconds)

**Answer**: C
**Explanation**: busy_timeout should be 30000ms (30 seconds) to handle concurrent write operations gracefully.

---

### Q40: What must be done after every code change?

A) Run `go build ./...` only
B) Run `golangci-lint run --fix`
C) Update documentation only
D) Create a pull request

**Answer**: B
**Explanation**: ALWAYS run `golangci-lint run --fix` FIRST after changes. It handles formatting, imports, and auto-fixable linters.

---

### Q41: Which port is used for OTLP gRPC telemetry export?

A) 4317
B) 4318
C) 9090
D) 13133

**Answer**: A
**Explanation**: OTLP uses port 4317 for gRPC and 4318 for HTTP. Port 13133 is the collector health check.

---

### Q42: What is the correct behavior when a parallel test fails?

A) Remove `t.Parallel()` from the test
B) Mark the test as skipped
C) Fix the production bug revealed
D) Run tests sequentially instead

**Answer**: C
**Explanation**: Failing parallel tests reveal production bugs (race conditions). Fix the bug, don't remove parallelism.

---

### Q43: What is the single source of truth for project status?

A) README.md
B) PROJECT-STATUS.md
C) CHANGELOG.md
D) GitHub Issues

**Answer**: B
**Explanation**: PROJECT-STATUS.md is the ONLY authoritative source for project status.

---

### Q44: Which infrastructure component handles IP allowlisting?

A) I2: Networking
B) I6: Crypto
C) I16: Security
D) I1: Configuration

**Answer**: A
**Explanation**: I2: Networking handles HTTP, HTTPS, gRPC, REST, load balancing, and firewalls including IP allowlisting.

---

### Q45: What algorithm agility requirement applies to cryptoutil?

A) Support only one algorithm per operation type
B) Support configurable algorithms with FIPS-approved defaults
C) Support all algorithms without restriction
D) Support legacy algorithms for backward compatibility

**Answer**: B
**Explanation**: Algorithm agility is required: all crypto operations must support configurable algorithms with FIPS-approved defaults.

---

### Q46: What is the correct way to declare default values in Go code?

A) Inline literals in function calls
B) Named variables (e.g., `var defaultConfigFiles = []string{}`)
C) Constants in the same file
D) Environment variables

**Answer**: B
**Explanation**: ALWAYS declare default values as named variables rather than inline literals, following the established pattern.

---

### Q47: Which endpoint returns the OpenID Connect Discovery document?

A) `/oauth2/v1/authorize`
B) `/.well-known/oauth-authorization-server`
C) `/.well-known/openid-configuration`
D) `/oauth2/v1/discovery`

**Answer**: C
**Explanation**: OpenID Connect Discovery is at `/.well-known/openid-configuration`. JWKS is at `/.well-known/jwks.json`.

---

### Q48: What type should be used for nullable UUID foreign keys in GORM?

A) `*googleUuid.UUID`
B) `sql.NullString`
C) `NullableUUID` (custom type)
D) `string`

**Answer**: C
**Explanation**: Pointer UUIDs cause "row value misused" errors in SQLite. Use the custom NullableUUID type that implements sql.Scanner and driver.Valuer.

---

### Q49: What is the correct GORM annotation for JSON array fields?

A) `gorm:"type:json"`
B) `gorm:"serializer:json"`
C) `gorm:"type:text"`
D) `gorm:"type:jsonb"`

**Answer**: B
**Explanation**: Use `serializer:json` instead of `type:json` for cross-DB compatibility. SQLite lacks native JSON type, so GORM handles encoding/decoding for TEXT columns.

---

### Q50: What must happen before pushing code changes?

A) Only `go build` must pass
B) All pre-push hooks must pass
C) Documentation must be complete
D) PR must be approved

**Answer**: B
**Explanation**: All pre-push gates must pass: golangci-lint, tests, coverage, cspell spelling, and gitleaks scanning.

---

## Session Summary

**Topics Covered**:

- FIPS 140-3 cryptographic requirements
- Key hierarchy and security architecture
- Code quality and coverage standards
- Product structure (P1-P4)
- Infrastructure components (I1-I16)
- Database configuration (SQLite/PostgreSQL)
- Testing patterns and conventions
- Deployment and telemetry

**Next Session**: GROOMING-SESSION-02 - Identity OAuth 2.1 Deep Dive
