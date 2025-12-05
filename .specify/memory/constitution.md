# cryptoutil Constitution

## I. Product Delivery Requirements

### Four Working Products Goal

cryptoutil MUST deliver four Products that are independently or jointly deployable

| Product | Description | Standalone | United |
|---------|-------------|------------|--------|
| P1: JOSE | JSON Object Signing and Encryption (JWK, JWKS, JWE, JWS, JWT, OAuth2.1, OIDC1.0) | ✅ | ✅ |
| P2: Identity | Identity (OAuth 2.1 AuthZ, OIDC 1.0 IdP AuthN, Multi-factor authentication, FIDO2/WebAuthn) | ✅ | ✅ |
| P3: KMS | Key Management Service (ElasticKeys contain MaterialKey(s), ElasticKeys support encrypt/decrypt/sign/verify, rotation, policies) | ✅ | ✅ |
| P4: CA | Certificate Authority (X.509 v3, PKIX RFC 5280, CSR, OCSP, CRL, PKI, EST, SCEP, CMPv2, CMC, ACME) | ✅ | ✅ |

### Standalone Mode Requirements

Each product MUST:

- Support start independently in isolation without other products
- Have working Docker Compose deployments that start independently in isolation without other products
- Pass all unit, integration, fuzz, bench, and end-to-end (e2e) tests in isolation without other products
- Support SQLite (dev, in-memory or file-based) and PostgreSQL (dev & prod)
- Support configuration via 1) optional environment variables, 2) optional command line parameters, and 3) optional one or more YAML files; default no settings starts in dev mode

### United Mode Requirements

All four products MUST:

- Support deploy together with 1-3 of the other products, via single Docker Compose without non-overlapping ports
- Share telemetry infrastructure, including a reusable Docker Compose
- Support optional inter-product federation via settings
- Share a reusable crypto implementation (not external)
- Pass all isolated E2E test suites
- Pass all federated E2E test suites

### Architecture Clarity

Clear separation between infrastructure and products:

- **Infrastructure (internal/infra/*)**: Reusable building blocks (config, networking, telemetry, crypto, database)
- **Products (internal/product/*)**: Deployable services built from infrastructure

## II. Cryptographic Compliance and Standards

All cryptographic operations MUST use NIST FIPS 140-3 approved algorithms. FIPS mode is ALWAYS enabled and MUST NEVER be disabled. Approved algorithms include:

- RSA ≥ 2048 bits, AES ≥ 128 bits, EC NIST curves, EdDSA, ECDH, EdDH for ciphers and signatures; NEVER 3DES, DES
- PBKDF2-HMAC-SHA256, PBKDF2-HMAC-SHA384, PBKDF2-HMAC-SHA256 for password hashing; NEVER bcrypt, scrypt, or Argon2)
- SHA-512, SHA-384, or SHA-256; NEVER MD5 or SHA-1

Algorithm agility is required: all crypto operations must support configurable algorithms with FIPS-approved secure defaults.

- Cryptographically secure entropy and random number generation
- CA/Browser Forum Baseline Requirements for TLS Server Certificates
- RFC 5280 compliance for X.509 certificates and CRLs
- Certificate serial numbers: minimum 64 bits CSPRNG, non-sequential, >0, <2^159
- Maximum 398 days validity for subscriber certificates
- Full cert chain validation, MinVersion: TLS 1.3+, never InsecureSkipVerify

All data at rest that is secret (e.g. Passwords, Keys) or sensitive (e.g. Personally Identifiable Information) MUST be encrypted or hashed

- Data is SEARCHABLE and DOESN'T need decryption (e.g. Magic Links): MUST use Deterministic Hash; use HKDF or PBKDF2 algorithm with keys in an enclave (e.g. PII)
- Data is SEARCHABLE and DOES need decryption (e.g. PII): MUST be Deterministic Cipher; use convergent encryption AES-GCM-IV algorithm with keys and IV in an enclave
- Data is NON-SEARCHABLE and DOESN'T need decryption (e.g. Passwords, OTPs): MUST use high-entropy, Non-Deterministic Hash
- Data is NON-SEARCHABLE and DOES need decryption (e.g. Keys): MUST use high-entropy, Non-Deterministic Cipher; AES-GCM preferred, or AES-CBC

All secret or sensitive data used by containers for configurations and parameters MUST use Docker/Kubernetes secrets:

- NEVER use environment variables
- Docker secrets mounted to `/run/secrets/` with file:// URLs; NEVER use environment variables
- Kubernetes secrets mounted as files; NEVER use environment variables

## III. KMS Hierarchical Key Security

Multi-layer KMS cryptographic barrier architecture:

- **Unseal secrets** → **Root keys** → **Intermediate keys** → **Content keys**
- All keys encrypted at rest, proper key versioning and rotation
- All KMS cryptoutil instances sharing a database MUST use the same unseal secrets for interoperability; derive same JWKs with same kids, or use same JWKs in enclave (e.g. PKCS#11, PKCS#12, HSM, TPM 2.0, Yubikey)
- NEVER use environment variables for secrets in all deployment; ALWAYS use Docker/Kubernetes secrets, including development, because it needs to reproduce production security

## IV. Go Testing Requirements

- Table-driven tests with `t.Parallel()` mandatory
- Test helpers marked with `t.helper()` mandatory
- NEVER use magic values in test code - ALWAYS use random, runtime-generated UUIDv7, or magic values and constants in package `magic` for self-documenting code and code-navigation in IDEs
- All port listeners MUST support dynamic port allocation for tests (port 0, extract actual assigned port); maximum use of test concurrency in unit, integration, and e2e tests serves to validate robustness of main code against mult-thread and multi-process bugs
- Test file suffixes: `_test.go` (unit), `_bench_test.go` (bench), `_fuzz_test.go` (fuzz), `_integration_test.go` (integration)

## V. Code Quality Excellence

- ALWAYS fix linting/formatting errors - NO EXCEPTIONS - Production code, test code, demos, examples, utilities, configuration, documentation, workflows - ALL must pass
- NEVER use `//nolint:` directives except for documented linter bugs
- ALWAYS use UTF-8 without BOM for ALL text file encoding; never use UTF-16, UTF-32, CP-1252, ASCII
- File size limits: 300 (soft), 400 (medium), 500 (hard → refactor required); ideal for user development and reviews, and LLM agent development and reviews
- 85%+ production coverage, 90%+ infrastructure (cicd), 100% utility code
- ALWAYS fix all pre-commit hook errors; see ./.pre-commit-config.yaml
- ALWAYS fix all pre-push hook errors; see ./.pre-commit-config.yaml
- All code builds  `go build ./...`, `mvn compile`
- All code changes pass `golangci-lint run --fix`
- All tests pass (`go test ./... -cover`)
- Coverage maintained at target thresholds, and gradually increased

## VI. Development Workflow and Evidence-Based Completion

### Evidence-Based Task Completion

No task is complete without objective, verifiable evidence:

- Code evidence: `go build ./...` clean, `golangci-lint run` clean, coverage ≥85%
- Test evidence: All tests passing, no skips without tracking
- Integration evidence: Core E2E demos work (`go run ./cmd/demo all` 7/7 steps)
- Documentation evidence: PROJECT-STATUS.md updated

Quality gates are MANDATORY - task NOT complete until all checks pass.

### Work Patterns

- ALWAYS Use Copilot Extension's built-in tools over terminal commands (create_file, read_file, runTests)
- Commit frequently with conventional commit format, and fix all pre-commit errors
- Work continuously until task complete with evidence
- Progressive validation after every task (TODO scan, test run, coverage, integration, documentation)

### Spec Kit Iteration Lifecycle

#### Iteration Workflow (MANDATORY)

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

#### Pre-Implementation Gates

Before running `/speckit.implement`:

- [ ] All `[NEEDS CLARIFICATION]` markers resolved in spec.md
- [ ] `/speckit.clarify` executed if spec was created/modified
- [ ] `/speckit.analyze` executed after `/speckit.tasks`
- [ ] All requirements have corresponding tasks
- [ ] No orphan tasks without requirement traceability

#### Post-Implementation Gates

Before marking iteration complete:

- [ ] `go build ./...` produces no errors
- [ ] `go test ./...` passes with 0 failures (not just "pass individually")
- [ ] `golangci-lint run` passes with no violations
- [ ] `/speckit.checklist` executed and all items verified
- [ ] Coverage targets maintained (85% production, 90% infrastructure)
- [ ] All spec.md status markers accurate and up-to-date
- [ ] No deferred items without documented justification

#### Iteration Completion Criteria

An iteration is NOT COMPLETE until:

1. **All workflow steps executed** (1-8 above)
2. **All gates passed** (pre and post implementation)
3. **Evidence documented** in PROGRESS.md
4. **Status markers updated** in spec.md
5. **No build errors** in `go build ./...`
6. **No test failures** in `go test ./...`
7. **No lint errors** in `golangci-lint run`

#### Gate Failure Protocol

When a gate fails:

1. **STOP** - Do not proceed to next step
2. **Document** - Record failure in PROGRESS.md
3. **Fix** - Address the root cause
4. **Retest** - Re-run the gate
5. **Evidence** - Document passing evidence

**NEVER** mark an iteration complete with failing gates.

## X. Governance and Documentation Standards

### Decision Authority

- **Technical decisions**: Follow copilot instructions in `.github/instructions/`
- **Architectural decisions**: Document in ADRs, follow Standard Go Project Layout
- **Compliance decisions**: CA/Browser Forum Baseline Requirements, RFC 5280, FIPS 140-3, NIST SP 800-57

### Documentation Standards

- PROJECT-STATUS.md is the ONLY authoritative status source for speckit
- Keep docs in 2 main files: README.md (main), docs/README.md (deep dive)
- NEVER create separate documentation files for scripts or tools

**Version**: 1.1.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-04
