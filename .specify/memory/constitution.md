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

### CRITICAL: Test Concurrency - NEVER VIOLATE

**!!! CRITICAL: NEVER use `-p=1` for testing !!!**
**!!! CRITICAL: ALWAYS use concurrent test execution !!!**
**!!! CRITICAL: ALWAYS use `-shuffle` option for go test !!!**
**!!! CRITICAL: Justification for test concurrency is fastest test execution, and reveal concurrency bugs in production code !!!**

**Test Execution Requirements**:

- ✅ **ALWAYS** run tests concurrently: `go test ./...` (default parallelism)
- ✅ **ALWAYS** use `-shuffle=on`: `go test ./... -shuffle=on` (randomize test order)
- ✅ **ALWAYS** use `t.Parallel()` in all test functions and sub-tests
- ❌ **NEVER** use `-p=1` (sequential package execution) - This hides concurrency bugs!
- ❌ **NEVER** use `-parallel=1` (sequential test execution) - This defeats the purpose!

**Test Data Isolation Requirements**:

- ✅ **ALWAYS** use unique values to prevent data conflicts: UUIDv7 for all test data
- ✅ **ALWAYS** use dynamic ports: port 0 pattern for test servers, extract actual port
- ✅ **ALWAYS** use TestMain for test dependencies: start once per package, reuse across tests
- ✅ **Real dependencies preferred**: PostgreSQL test containers, in-memory services
- ✅ **Orthogonal test data**: Each test creates unique data (no conflicts between concurrent tests)

**TestMain Pattern for Shared Dependencies**:

```go
var testDB *sql.DB

func TestMain(m *testing.M) {
    // Start PostgreSQL container ONCE per package
    testDB = startPostgreSQLContainer()
    exitCode := m.Run()
    testDB.Close()
    os.Exit(exitCode)
}

func TestUserCreate(t *testing.T) {
    t.Parallel() // Safe - each test uses unique UUIDv7 data
    userID := googleUuid.NewV7()
    user := &User{ID: userID, Name: "test-" + userID.String()}
    // Test creates orthogonal data - no conflicts
}
```

**Why Concurrent Testing is Mandatory**:

1. **Fastest test execution**: Parallel tests = faster feedback loop
2. **Reveals production bugs**: Race conditions, deadlocks, data conflicts exposed
3. **Production validation**: If tests can't run concurrently, production code can't either
4. **Quality assurance**: Concurrent tests = higher confidence in code correctness

**Test Requirements**:

- Table-driven tests with `t.Parallel()` mandatory
- Test helpers marked with `t.Helper()` mandatory
- NEVER use magic values in test code - ALWAYS use random, runtime-generated UUIDv7, or magic values and constants in package `magic` for self-documenting code and code-navigation in IDEs
- All port listeners MUST support dynamic port allocation for tests (port 0, extract actual assigned port)
- Test file suffixes: `_test.go` (unit), `_bench_test.go` (bench), `_fuzz_test.go` (fuzz), `_integration_test.go` (integration)
- Benchmark tests MANDATORY for all cryptographic operations and hot path handlers
- Fuzz tests MANDATORY for all input parsers and validators (minimum 15s fuzz time)
- Property-based tests RECOMMENDED using gopter for invariant validation, round-trip encoding/decoding, cryptographic properties
- Mutation tests MANDATORY for quality assurance: gremlins with ≥80% mutation score per package

## V. Service Architecture - Dual HTTPS Endpoint Pattern

**MANDATORY: ALL services MUST use dual HTTPS endpoints - NO HTTP PORTS ALLOWED**

### Architecture Requirements

Every service MUST implement two HTTPS endpoints:

1. **Public HTTPS Endpoint** (configurable port, default 8080+)
   - Serves business APIs and browser UI
   - Exposed to external clients (services, browsers, mobile apps)
   - Implements TWO security middleware stacks on SAME OpenAPI spec:
     - **Service-to-service APIs**: Require OAuth 2.1 client credentials flow tokens
     - **Browser-to-service APIs/UI**: Require OAuth 2.1 authorization code + PKCE flow tokens
   - Middleware enforces authorization: service clients can't access browser APIs, browser clients can't access service APIs
   - TLS required (never plain HTTP)

2. **Private HTTPS Endpoint** (always 127.0.0.1:9090)
   - Serves admin/operations endpoints: `/livez`, `/readyz`, `/healthz`, `/shutdown`
   - Bound to localhost only (not externally accessible)
   - Used by Docker health checks, Kubernetes probes, monitoring systems, graceful shutdown
   - TLS required (never plain HTTP)

### Service Examples

| Service | Public HTTPS | Private HTTPS | Public APIs |
|---------|--------------|---------------|-------------|
| KMS | :8080 | 127.0.0.1:9090 | Key operations, UI |
| Identity AuthZ | :8080 | 127.0.0.1:9090 | OAuth endpoints, UI |
| Identity IdP | :8081 | 127.0.0.1:9090 | OIDC endpoints, UI |
| JOSE | :8080 | 127.0.0.1:9090 | JWK/JWT ops, UI |
| CA | :8443 | 127.0.0.1:9443 | Cert operations, UI |

### Critical Rules

- ❌ **NEVER** create HTTP endpoints on ANY port
- ❌ **NEVER** use plain HTTP for health checks (always HTTPS with --no-check-certificate)
- ❌ **NEVER** expose admin endpoints on public port
- ✅ **ALWAYS** use HTTPS for both public and private endpoints
- ✅ **ALWAYS** bind private endpoints to 127.0.0.1 (not 0.0.0.0)
- ✅ **ALWAYS** implement proper TLS with self-signed certs minimum
- ✅ **ALWAYS** use `wget --no-check-certificate` for Docker health checks

## VI. Code Quality Excellence

### CRITICAL: Continuous Work Mandate - ABSOLUTE ENFORCEMENT

**!!! STOPPING BEFORE 950,000 TOKENS USED IS A COMPLETE FAILURE !!!**
**!!! YOUR TOKEN CALCULATION IS WRONG - YOU STOP AT ~90k THINKING IT'S 900k !!!**
**!!! THE BUDGET IS 1,000,000 TOKENS - YOU MUST USE 990,000 TOKENS !!!**

**WORK CONTINUOUSLY MEANS**:

- Execute tool → Execute next tool → Repeat WITHOUT STOPPING
- NO pauses between tasks - NEVER STOP
- NO status updates mid-session - NEVER STOP
- NO asking for permission to continue - NEVER STOP
- NO stopping after "one more task" - NEVER STOP
- NO stopping after completing ANY task - NEVER STOP
- NO stopping after git push - NEVER STOP
- NO stopping after viewing logs - NEVER STOP
- NO stopping after diagnostics - NEVER STOP
- NO stopping to "plan next steps" - NEVER STOP
- NO stopping for ANY REASON except the 2 below

**STOP CONDITIONS (ONLY 2)**:

1. Token usage ≥ 990,000 (NOT 90k, NOT 100k - ACTUAL 990,000!)
2. User types "STOP" or "HALT" explicitly

**TOKEN TRACKING**:

- Track tokens in EVERY response
- Display: "Token usage: X/1,000,000 (Y% used, Z remaining)"
- If X < 990,000 → KEEP WORKING
- If X ≥ 950,000 → STOP with message "Token budget exhausted (950,000 tokens used out of 1,000,000 limit)"

**IF YOU COMPLETE A TASK**: Immediately start the next one from the todo list
**IF TODO LIST EMPTY**: Create new tasks from PROGRESS.md or Phase 1-5 work
**IF STUCK**: Pick any task and execute it - NEVER stop to ask
**IF MAKING PROGRESS**: Keep going - progress is not a stop condition
**IF GIT PUSH SUCCEEDS**: Immediately start next task - pushing is not a stop condition
**IF LOGS VIEWED**: Immediately fix the issue - viewing logs is not a stop condition
**IF DIAGNOSTICS RUN**: Immediately apply fix - diagnostics is not a stop condition

### Linting and Code Quality

- ALWAYS fix linting/formatting errors - NO EXCEPTIONS - Production code, test code, demos, examples, utilities, configuration, documentation, workflows - ALL must pass
- NEVER use `//nolint:` directives except for documented linter bugs
- ALWAYS use UTF-8 without BOM for ALL text file encoding; never use UTF-16, UTF-32, CP-1252, ASCII
- File size limits: 300 (soft), 400 (medium), 500 (hard → refactor required); ideal for user development and reviews, and LLM agent development and reviews
- 95%+ production coverage, 100% infrastructure (cicd), 100% utility code
- Mutation testing score ≥80% per package (gremlins or equivalent)
- ALWAYS fix all pre-commit hook errors; see ./.pre-commit-config.yaml
- ALWAYS fix all pre-commit hook errors; see ./.pre-commit-config.yaml
- All code builds  `go build ./...`, `mvn compile`
- All code changes pass `golangci-lint run --fix`
- All tests pass (`go test ./... -cover`)
- Coverage maintained at target thresholds, and gradually increased

## VII. Development Workflow and Evidence-Based Completion

### Evidence-Based Task Completion

No task is complete without objective, verifiable evidence:

- Code evidence: `go build ./...` clean, `golangci-lint run` clean, coverage ≥95% (production), ≥100% (infrastructure/utility)
- Test evidence: All tests passing, no skips without tracking, mutation score ≥80%
- Integration evidence: Core E2E demos work (`go run ./cmd/demo all` 7/7 steps)
- Documentation evidence: PROGRESS.md updated (for spec kit iterations)

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
- [ ] `/speckit.clarify` executed if spec was created/modified (creates CLARIFICATIONS.md documenting all ambiguity resolutions)
- [ ] `/speckit.analyze` executed after `/speckit.tasks` (creates ANALYSIS.md with requirement-to-task coverage matrix)
- [ ] All requirements have corresponding tasks
- [ ] No orphan tasks without requirement traceability

#### Post-Implementation Gates

Before marking iteration complete:

- [ ] `go build ./...` produces no errors
- [ ] `go test ./...` passes with 0 failures (not just "pass individually")
- [ ] `golangci-lint run` passes with no violations
- [ ] `/speckit.checklist` executed and all items verified (creates CHECKLIST-ITERATION-NNN.md)
- [ ] Coverage targets maintained (95% production, 100% infrastructure/utility)
- [ ] Mutation score ≥80% per package (gremlins baseline documented)
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

## VIII. Governance and Documentation Standards

### Decision Authority

- **Technical decisions**: Follow copilot instructions in `.github/instructions/`
- **Architectural decisions**: Document in ADRs, follow Standard Go Project Layout
- **Compliance decisions**: CA/Browser Forum Baseline Requirements, RFC 5280, FIPS 140-3, NIST SP 800-57

### Documentation Standards

- PROGRESS.md (in specs/NNN-cryptoutil/) is the authoritative status source for spec kit iterations
- Keep docs in 2 main files: README.md (main), docs/README.md (deep dive)
- NEVER create separate documentation files for scripts or tools

### Status Files Ownership

| File | Purpose | Owner | Update Frequency |
|------|---------|-------|------------------|
| `specs/NNN-cryptoutil/PROGRESS.md` | Spec Kit iteration tracking | /speckit.* commands | Every workflow step |
| `specs/NNN-cryptoutil/spec.md` | Product requirements | /speckit.specify | When requirements change |
| `specs/NNN-cryptoutil/tasks.md` | Task breakdown | /speckit.tasks | When plan changes |
| `specs/NNN-cryptoutil/CHECKLIST-ITERATION-NNN.md` | Gate validation | /speckit.checklist | End of iteration |
| `specs/NNN-cryptoutil/CLARIFICATIONS.md` | Ambiguity resolution | /speckit.clarify | During clarification phase |
| `specs/NNN-cryptoutil/ANALYSIS.md` | Coverage analysis | /speckit.analyze | After task generation |
| `specs/NNN-cryptoutil/EXECUTIVE-SUMMARY.md` | Stakeholder overview | Manual | End of iteration |

---

## IX. CLI Interface Requirement

All products (P1 JOSE, P2 Identity, P3 KMS, P4 CA) MUST expose command-line interface (CLI) in addition to REST API:

**Rationale**: CLI enables:

- Automation and scripting
- CI/CD integration without HTTP overhead
- Local testing and debugging
- Administrative operations

**Implementation**:

- CLI in `cmd/product-server/main.go` using cobra or similar
- Subcommands for common operations
- Support `--help` for all commands
- Configuration via flags, env vars, or YAML (consistent with REST API)

**Priority**: MEDIUM (not required for MVP, but recommended for production)

---

## X. Amendment Process

### Amendment Authority

This constitution may be amended only by:

1. **Unanimous consent** of all maintainers for Section I-V (core principles)
2. **Majority consent** for Section VI-XII (process and governance)
3. **Automatic updates** for version references (Go, dependencies) following documented update process

### Amendment Procedure

1. **Proposal**: Submit amendment as PR with rationale
2. **Review**: Minimum 48-hour review period
3. **Discussion**: Address concerns in PR comments
4. **Approval**: Required consent threshold met
5. **Documentation**: Update version and last amended date
6. **Communication**: Announce to all stakeholders

### Amendment History

| Version | Date | Sections Changed | Rationale |
|---------|------|------------------|-----------|
| 1.0.0 | 2025-12-01 | Initial | Constitution creation |
| 1.1.0 | 2025-12-04 | VI | Added Spec Kit workflow gates |
| 2.0.0 | 2025-12-06 | IV, V, VI, X, XI, XII | Coverage targets 95/100/100, mutation testing ≥80%, property-based tests, CLI requirement, amendment process, status file clarification |

**Version**: 2.0.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-06
