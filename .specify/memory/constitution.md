# cryptoutil Constitution

## Core Principles

### I. FIPS 140-3 Compliance First

All cryptographic operations MUST use NIST FIPS 140-3 approved algorithms. FIPS mode is ALWAYS enabled by default and MUST NEVER be disabled. Approved algorithms include:

- RSA ≥ 2048 bits, AES ≥ 128 bits, EC NIST curves, EdDSA
- PBKDF2-HMAC-SHA256 for password hashing (NEVER bcrypt, scrypt, or Argon2)
- SHA-256 or SHA-512 (NEVER MD5 or SHA-1)

Algorithm agility is required: all crypto operations must support configurable algorithms with FIPS-approved defaults.

### II. Evidence-Based Task Completion

No task is complete without objective, verifiable evidence:

- Code evidence: `go build ./...` clean, `golangci-lint run` clean, coverage ≥80%
- Test evidence: All tests passing, no skips without tracking
- Integration evidence: Core E2E demos work (`go run ./cmd/demo all` 7/7 steps)
- Documentation evidence: PROJECT-STATUS.md updated

Quality gates are MANDATORY - task NOT complete until all checks pass.

### III. Hierarchical Key Security

Multi-layer cryptographic barrier architecture:

- **Unseal secrets** → **Root keys** → **Intermediate keys** → **Content keys**
- All keys encrypted at rest, proper key versioning and rotation
- All cryptoutil instances using the same unseal secrets MUST derive identical JWKs (including KIDs) for interoperability
- NEVER use environment variables for secrets in production; use Docker/Kubernetes secrets

### IV. Code Quality Excellence

ALL linting/formatting errors are MANDATORY to fix - NO EXCEPTIONS:

- Production code, test code, demos, examples, utilities - ALL must pass
- NEVER use `//nolint:` directives except for documented linter bugs
- File size limits: 300 (soft), 400 (medium), 500 (hard → refactor required)
- UTF-8 without BOM for ALL text files
- 80%+ production coverage, 85%+ infrastructure (cicd), 95%+ utility code

### V. Product Architecture Clarity

Clear separation between infrastructure and products:

- **Infrastructure (internal/infra/*)**: Reusable building blocks (config, networking, telemetry, crypto, database)
- **Products (internal/product/*)**: Deployable services built from infrastructure
  - P1: JOSE (JWK, JWKS, JWE, JWS, JWT, OAuth2.1, OIDC1.0)
  - P2: Identity (OAuth 2.1 AuthZ, OIDC IdP, MFA, FIDO2/WebAuthn)
  - P3: KMS (ElasticKey, MaterialKey management, rotation, policies)
  - P4: Certificates (X.509, CSR, OCSP, CRL, PKI, ACME)

## Security Requirements

### Cryptographic Standards

- CA/Browser Forum Baseline Requirements for TLS Server Certificates
- RFC 5280 compliance for X.509 certificates
- Certificate serial numbers: minimum 64 bits CSPRNG, non-sequential, >0, <2^159
- Maximum 398 days validity for subscriber certificates
- Full cert chain validation, MinVersion: TLS 1.2+, never InsecureSkipVerify

### Secret Management

- Docker secrets mounted to `/run/secrets/` with file:// URLs
- Kubernetes secrets mounted as files, not environment variables
- IP allowlisting (IPs & CIDR), per-IP rate limiting
- CORS, CSRF, strict HTTP headers, audit logging

## Quality Gates

### Pre-Commit Gates

1. UTF-8 without BOM enforcement (`cicd all-enforce-utf8`)
2. gofumpt formatting (not gofmt)
3. golangci-lint v2.6.2+ with all enabled linters
4. markdownlint, yamllint for documentation

### Pre-Push Gates

1. All code changes pass `golangci-lint run --fix`
2. All tests pass (`go test ./... -cover`)
3. Coverage maintained at target thresholds
4. cspell spelling verification
5. gitleaks secrets scanning

### Testing Requirements

- Table-driven tests with `t.Parallel()` mandatory
- NEVER hardcode test values - use magic package constants OR runtime-generated UUIDv7
- Dynamic port allocation for server tests (port 0, extract actual assigned port)
- Test file suffixes: `_test.go` (unit), `_bench_test.go` (bench), `_fuzz_test.go` (fuzz), `_integration_test.go` (integration)

## Governance

### Decision Authority

- **Technical decisions**: Follow copilot instructions in `.github/instructions/`
- **Architectural decisions**: Document in ADRs, follow Standard Go Project Layout
- **Compliance decisions**: CA/Browser Forum Baseline Requirements, RFC 5280, NIST SP 800-57

### Work Patterns

- Use built-in tools over terminal commands (create_file, read_file, runTests)
- Commit frequently with conventional commit format
- Work continuously until task complete with evidence
- Progressive validation after every task (TODO scan, test run, coverage, integration, documentation)

### Documentation Standards

- PROJECT-STATUS.md is the ONLY authoritative status source
- README.md for main documentation, docs/README.md for deep dives
- NEVER create separate documentation files for scripts or tools
- Keep docs in 2 main files: README.md (main), docs/README.md (deep dive)

**Version**: 1.0.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-01
