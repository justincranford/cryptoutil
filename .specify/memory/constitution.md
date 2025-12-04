# cryptoutil Constitution

## Core Principles

### I. FIPS 140-3 Compliance First

All cryptographic operations MUST use NIST FIPS 140-3 approved algorithms. FIPS mode is ALWAYS enabled by default and MUST NEVER be disabled. Approved algorithms include:

- RSA ≥ 2048 bits, AES ≥ 128 bits, EC NIST curves, EdDSA, ECDH, EdDH
- PBKDF2-HMAC-SHA256, PBKDF2-HMAC-SHA384, PBKDF2-HMAC-SHA256 for password hashing (NEVER bcrypt, scrypt, or Argon2)
- SHA-512, SHA-384, or SHA-256 (NEVER MD5 or SHA-1)

Algorithm agility is required: all crypto operations must support configurable algorithms with FIPS-approved defaults.

### II. Evidence-Based Task Completion

No task is complete without objective, verifiable evidence:

- Code evidence: `go build ./...` clean, `golangci-lint run` clean, coverage ≥80%
- Test evidence: All tests passing, no skips without tracking
- Integration evidence: Core E2E demos work (`go run ./cmd/demo all` 7/7 steps)
- Documentation evidence: PROJECT-STATUS.md updated

Quality gates are MANDATORY - task NOT complete until all checks pass.

### III. Code Quality Excellence

ALL linting/formatting errors are MANDATORY to fix - NO EXCEPTIONS:

- Production code, test code, demos, examples, utilities - ALL must pass
- NEVER use `//nolint:` directives except for documented linter bugs
- File size limits: 300 (soft), 400 (medium), 500 (hard → refactor required)
- UTF-8 without BOM for ALL text files
- 80%+ production coverage, 85%+ infrastructure (cicd), 95%+ utility code

### IV. KMS Hierarchical Key Security

Multi-layer KMS cryptographic barrier architecture:

- **Unseal secrets** → **Root keys** → **Intermediate keys** → **Content keys**
- All keys encrypted at rest, proper key versioning and rotation
- All KMS cryptoutil instances using the same unseal secrets MUST derive identical JWKs (including KIDs) for interoperability
- NEVER use environment variables for secrets in production; ALWAYS use Docker/Kubernetes secrets

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
- Full cert chain validation, MinVersion: TLS 1.3+, never InsecureSkipVerify

### Secret Management

- Docker secrets mounted to `/run/secrets/` with file:// URLs
- Kubernetes secrets mounted as files, not environment variables
- IP allowlisting (IPs & CIDR), per-IP rate limiting
- CORS, CSRF, XSS prevention with CSP
- Strict HTTP security headers (HSTS, X-Frame-Options, X-Content-Type-Options, etc.)
- Security header validation and audit logging

## Quality Gates

### Pre-Commit Gates

1. See pre-commit hooks in ./.pre-commit-config.yaml
2. All code builds  `go build ./...`
3. golangci-lint v2.6.2+ with all enabled linters
4. gofumpt formatting (not gofmt)
5. Fix all lint errors in Go, Python (Pylance), Java, Configs, Workflows, etc.

### Pre-Push Gates

1. See pre-push hooks in ./.pre-commit-config.yaml
2. All code changes pass `golangci-lint run --fix`
3. All tests pass (`go test ./... -cover`)
4. Coverage maintained at target thresholds

### Testing Requirements

- Table-driven tests with `t.Parallel()` mandatory
- Test helpers marked with `t.helper()` mandatory
- NEVER hardcode test values - ALWAYS use runtime-generated UUIDv7, or magic values and constants in package `magic`
- ALWAYS support dynamic port allocation for all servers to support concurrent unit, integration, and e2e testing (port 0, extract actual assigned port)
- Test file suffixes: `_test.go` (unit), `_bench_test.go` (bench), `_fuzz_test.go` (fuzz), `_integration_test.go` (integration)

## Governance

### Decision Authority

- **Technical decisions**: Follow copilot instructions in `.github/instructions/`
- **Architectural decisions**: Document in ADRs, follow Standard Go Project Layout
- **Compliance decisions**: CA/Browser Forum Baseline Requirements, RFC 5280, FIPS 140-3, NIST SP 800-57

### Work Patterns

- ALWAYS Use Copilot Extension's built-in tools over terminal commands (create_file, read_file, runTests)
- Commit frequently with conventional commit format
- Work continuously until task complete with evidence
- Progressive validation after every task (TODO scan, test run, coverage, integration, documentation)

### Documentation Standards

- PROJECT-STATUS.md is the ONLY authoritative status source
- README.md for main documentation, docs/README.md for deep dives
- NEVER create separate documentation files for scripts or tools
- Keep docs in 2 main files: README.md (main), docs/README.md (deep dive)

---

## VI. Spec Kit Iteration Lifecycle

### Iteration Workflow (MANDATORY)

Every iteration MUST follow this sequence:

```
1. /speckit.constitution  → Review/update principles (first iteration only)
2. /speckit.specify       → Define/update requirements (spec.md)
3. /speckit.clarify       → Resolve ALL ambiguities
4. /speckit.plan          → Technical implementation plan
5. /speckit.tasks         → Generate task breakdown
6. /speckit.analyze       → Coverage check (before implement)
7. /speckit.implement     → Execute implementation
8. /speckit.checklist     → Validate completion (after implement)
```

**CRITICAL**: Steps 3 and 6-8 are MANDATORY, not optional.

### Pre-Implementation Gates

Before running `/speckit.implement`:

- [ ] All `[NEEDS CLARIFICATION]` markers resolved in spec.md
- [ ] `/speckit.clarify` executed if spec was created/modified
- [ ] `/speckit.analyze` executed after `/speckit.tasks`
- [ ] All requirements have corresponding tasks
- [ ] No orphan tasks without requirement traceability

### Post-Implementation Gates

Before marking iteration complete:

- [ ] `go test ./...` passes with 0 failures (not just "pass individually")
- [ ] `go build ./...` produces no errors
- [ ] `golangci-lint run` passes with no new violations
- [ ] `/speckit.checklist` executed and all items verified
- [ ] Coverage targets maintained (80% production, 85% infrastructure)
- [ ] All spec.md status markers accurate and up-to-date
- [ ] No deferred items without documented justification

### Iteration Completion Criteria

An iteration is NOT COMPLETE until:

1. **All workflow steps executed** (1-8 above)
2. **All gates passed** (pre and post implementation)
3. **Evidence documented** in PROGRESS.md
4. **Status markers updated** in spec.md
5. **No test failures** in `go test ./...`
6. **No lint errors** in `golangci-lint run`

### Gate Failure Protocol

When a gate fails:

1. **STOP** - Do not proceed to next step
2. **Document** - Record failure in PROGRESS.md
3. **Fix** - Address the root cause
4. **Retest** - Re-run the gate
5. **Evidence** - Document passing evidence

**NEVER** mark an iteration complete with failing gates.

---

## VII. Product Delivery Requirements

### Four Working Products Goal

cryptoutil MUST deliver four independently deployable products:

| Product | Description | Standalone | United |
|---------|-------------|------------|--------|
| P1: JOSE | JSON Object Signing and Encryption | ✅ | ✅ |
| P2: Identity | OAuth 2.1 + OIDC IdP | ✅ | ✅ |
| P3: KMS | Key Management Service | ✅ | ✅ |
| P4: CA | Certificate Authority | ✅ | ✅ |

### Standalone Mode Requirements

Each product MUST:

- Start independently without other products
- Use embedded JOSE (not external)
- Support SQLite (dev) and PostgreSQL (prod)
- Pass all tests in isolation
- Have working Docker Compose deployment

### United Mode Requirements

All four products MUST:

- Deploy together in single Docker Compose
- Share telemetry infrastructure
- Support optional inter-product federation
- Pass full E2E test suite

**Version**: 1.1.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-03
