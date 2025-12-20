# cryptoutil Constitution

## I. Product Delivery Requirements

### Four Working Products Goal

cryptoutil MUST deliver four Products (9 total services: 8 product services + 1 demo service) that are independently or jointly deployable

| Product | Services | Description | Standalone | United |
|---------|----------|-------------|------------|--------|
| P1: JOSE | 1 service | JSON Object Signing and Encryption (JWK, JWKS, JWE, JWS, JWT) | ✅ | ✅ |
| P2: Identity | 5 services | OAuth 2.1 AuthZ, OIDC 1.0 IdP, Resource Server, Relying Party, Single Page Application | ✅ | ✅ |
| P3: KMS | 1 service | Key Management Service (ElasticKeys, MaterialKeys, encrypt/decrypt, sign/verify, rotation, policies) | ✅ | ✅ |
| P4: CA | 1 service | Certificate Authority (X.509 v3, PKIX RFC 5280, CSR, OCSP, CRL, PKI, EST, SCEP, CMPv2, CMC, ACME) | ✅ | ✅ |
| Demo: Learn-PS | 1 service | Pet Store demonstration service (validates service template reusability) | ✅ | ✅ |

### Complete Service Architecture (9 Services)

#### Product Services (8 Core Services)

| Service Alias | Full Name | Public Port | Admin Port | Description |
|---------------|-----------|-------------|------------|-----------|
| **sm-kms** | Secrets Manager - Key Management Service | 8080 | 9090 | Hierarchical key management with ElasticKeys and MaterialKeys |
| **pki-ca** | Public Key Infrastructure - Certificate Authority | 8380 | 9092 | X.509 certificate lifecycle, EST, OCSP, CRL, time-stamping |
| **jose-ja** | JOSE - JWK Authority | 8280 | 9093 | JWK, JWKS, JWE, JWS, JWT operations |
| **identity-authz** | Identity - Authorization Server | 8180 | 9091 | OAuth 2.1 authorization server, OIDC Discovery |
| **identity-idp** | Identity - Identity Provider | 8181 | 9091 | OIDC authentication, login/consent UI, MFA enrollment |
| **identity-rs** | Identity - Resource Server | 8182 | 9091 | Protected API with token validation (reference implementation) |
| **identity-rp** | Identity - Relying Party | 8183 | 9091 | Backend-for-Frontend pattern (reference implementation) |
| **identity-spa** | Identity - Single Page Application | 8184 | 9091 | Static hosting for SPA clients (reference implementation) |

#### Demonstration Service (1 Service)

| Service Alias | Full Name | Public Port | Admin Port | Description |
|---------------|-----------|-------------|------------|-----------|
| **learn-ps** | Learn - Pet Store | 8580 | 9095 | Educational service demonstrating service template usage (Phase 7) |

### Service Status and Implementation Priority

**Implementation Status by Service**:

| Service | Status | Priority | Deliverables | Notes |
|---------|--------|----------|--------------|-------|
| sm-kms | ✅ COMPLETE | Phase 0 | Dual servers, admin port 9090, public port 8080, DB migrations, telemetry | Reference implementation for dual-server pattern |
| pki-ca | ⚠️ PARTIAL | Phase 2.3 | Missing admin server, public port 8380 only | Needs dual-server migration |
| jose-ja | ⚠️ PARTIAL | Phase 2.1 | Missing admin server, public port 8280 only | Needs dual-server migration |
| identity-authz | ❌ INCOMPLETE | Phase 2.2 | Admin server exists (9091), missing public server (8180) | **CRITICAL BLOCKER**: Missing public HTTP server implementation |
| identity-idp | ❌ INCOMPLETE | Phase 2.2 | Admin server exists (9091), missing public server (8181) | **CRITICAL BLOCKER**: Missing public HTTP server implementation |
| identity-rs | ❌ INCOMPLETE | Phase 2.2 | Admin server exists (9091), missing public server (8182) | **CRITICAL BLOCKER**: Missing public HTTP server implementation |
| identity-rp | ❌ NOT STARTED | Phase 3+ | No servers | Reference implementation, optional deployment |
| identity-spa | ❌ NOT STARTED | Phase 3+ | No servers | Reference implementation, optional deployment |
| learn-ps | ❌ NOT STARTED | Phase 7 | No servers | Validates service template reusability |

**Architecture Blocker** (2025-12-20 validation):

Three Identity services (authz, idp, rs) cannot start because they lack public HTTP server implementation:

- `internal/identity/authz/server/server.go` ❌ MISSING (public OAuth 2.1 endpoints)
- `internal/identity/idp/server/server.go` ❌ MISSING (public OIDC endpoints)
- `internal/identity/rs/server/server.go` ❌ MISSING (public resource API)

**Evidence**: 5 E2E workflow failures (workflows 20388807383-20388120287), all fail at "Starting AuthZ server..." with 196-byte logs, immediate container exit. Configuration validated correct (TLS, DSN, secrets, OTEL), but binary has no public server code to start.

**Impact**: Blocks 3/11 workflows (E2E, Load, DAST) until public servers implemented (estimated 3-5 days development).

**Related Documentation**:

- `docs/WORKFLOW-FIXES-ROUND7.md` (commit 1cbf3d34, 228 lines): Round 7 investigation
- `specs/002-cryptoutil/implement/EXECUTIVE.md` (commit 57236a52): Workflow status, limitations
- `specs/002-cryptoutil/implement/DETAILED.md` (2025-12-20): E2E validation session

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

### CGO Ban - ABSOLUTE REQUIREMENT

**!!! CRITICAL: CGO IS BANNED EXCEPT FOR RACE DETECTOR !!!**

- **CGO_ENABLED=0** is MANDATORY for builds, tests, Docker, production deployments
- **ONLY EXCEPTION**: Race detector workflow requires CGO_ENABLED=1 (Go toolchain limitation)
- **NEVER** use dependencies requiring CGO (e.g., `github.com/mattn/go-sqlite3`)
- **ALWAYS** use CGO-free alternatives (e.g., `modernc.org/sqlite`)
- **Go Toolchain Limitation**: Race detector (`-race`) requires C-based ThreadSanitizer from LLVM
- **Rationale**: Maximum portability, static linking, cross-compilation, no C toolchain dependencies for production

**Enforcement in Test Files**:

- ❌ **NEVER** add `isCGOAvailable()` checks or skip tests for CGO availability
- ✅ SQLite tests using `modernc.org/sqlite` MUST ALWAYS run (CGO-free implementation)
- ✅ All database tests MUST pass with `CGO_ENABLED=0`

**Approved Dependencies** (CGO-free only):

| Package | CGO Required? | Status |
|---------|---------------|--------|
| modernc.org/sqlite | ❌ No (pure Go) | ✅ APPROVED |
| github.com/mattn/go-sqlite3 | ✅ Yes (C bindings) | ❌ BANNED |
| gorm.io/gorm | ❌ No | ✅ APPROVED |
| gorm.io/driver/sqlite (with modernc) | ❌ No | ✅ APPROVED |

### FIPS 140-3 Compliance

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
- Mutation tests MANDATORY for quality assurance: gremlins with ≥85% mutation score per package (Phase 4), ≥98% per package (Phase 5+)

**Test Execution Time Targets**:

- Unit test packages: MANDATORY <15 seconds per package (excludes integration/e2e tests)
- Full unit test suite: MANDATORY <180 seconds (3 minutes) total
- Integration/E2E tests: Excluded from strict timing (Docker startup overhead acceptable)
- Probabilistic execution MANDATORY for packages approaching 15s limit

**Probability-Based Test Execution**:

- `TestProbAlways` (100%): Base algorithms (RSA2048, AES256, ES256) - always test
- `TestProbQuarter` (25%): Key size variants (RSA3072, AES192) - statistical sampling
- `TestProbTenth` (10%): Less common variants (RSA4096, AES128) - minimal sampling
- `TestProbNever` (0%): Deprecated or extreme edge cases - skip
- Purpose: Maintain <15s per package timing while preserving comprehensive algorithm coverage
- Rationale: Faster test execution without sacrificing bug detection effectiveness

**main() Function Testability Pattern**:

- ALL main() functions MUST be thin wrappers delegating to co-located testable functions
- Pattern: `main()` calls `internalMain(args, stdin, stdout, stderr) int`
- `internalMain()` accepts injected dependencies for testing
- `main()` 0% coverage acceptable if `internalMain()` ≥95% coverage
- Rationale: Enables testing of exit codes, argument parsing, error handling without terminating test process

**Real Dependencies Preferred Over Mocks**:

- ALWAYS use real dependencies: PostgreSQL test containers, real crypto, real HTTP servers
- ONLY use mocks for: External services that can't run locally (email, SMS, cloud-only APIs)
- Rationale: Real dependencies reveal production bugs; mocks hide integration issues
- Examples:
  - ✅ PostgreSQL: Use test containers (NOT database/sql mocks)
  - ✅ Crypto operations: Use real crypto libraries (NOT mock implementations)
  - ✅ HTTP servers: Use real servers with test clients (NOT httptest mocks unless corner cases)
  - ❌ Email/SMS: Mock (external services)

**Race Condition Prevention - CRITICAL**:

- NEVER write to parent scope in parallel sub-tests, manipulate globals with t.Parallel(), or share sessions
- ALWAYS inline assertions, fresh test data, protect maps/slices with sync.Mutex
- Detection: `go test -race -count=2` (local + ci-race workflow)
- Details: .github/instructions/01-04.testing.instructions.md

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
   - Serves admin/operations endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
   - Bound to localhost only (not externally accessible)
   - Used by Docker health checks, Kubernetes probes, monitoring systems, graceful shutdown
   - TLS required (never plain HTTP)

### Service Examples (All 9 Services)

#### Product Services (8 Core Services)

| Service | Full Name | Public HTTPS | Private HTTPS | Public APIs |
|---------|-----------|--------------|---------------|-------------|
| **sm-kms** | Secrets Manager - KMS | :8080 | 127.0.0.1:9090 | Key operations (encrypt/decrypt, sign/verify), UI |
| **pki-ca** | PKI - Certificate Authority | :8380 | 127.0.0.1:9092 | X.509 cert operations, EST, OCSP, CRL, UI |
| **jose-ja** | JOSE - JWK Authority | :8280 | 127.0.0.1:9093 | JWK/JWKS/JWE/JWS/JWT operations, UI |
| **identity-authz** | Identity - Authorization Server | :8180 | 127.0.0.1:9091 | OAuth 2.1 endpoints, OIDC Discovery, UI |
| **identity-idp** | Identity - Identity Provider | :8181 | 127.0.0.1:9091 | OIDC authentication, login/consent, MFA, UI |
| **identity-rs** | Identity - Resource Server | :8182 | 127.0.0.1:9091 | Protected API with token validation (reference implementation) |
| **identity-rp** | Identity - Relying Party | :8183 | 127.0.0.1:9091 | Backend-for-Frontend pattern (reference implementation) |
| **identity-spa** | Identity - Single Page App | :8184 | 127.0.0.1:9091 | Static hosting for SPA clients (reference implementation) |

#### Demonstration Service (1 Service)

| Service | Full Name | Public HTTPS | Private HTTPS | Public APIs |
|---------|-----------|--------------|---------------|-------------|
| **learn-ps** | Learn - Pet Store | :8580 | 127.0.0.1:9095 | Educational service (Phase 7, validates service template) |

**Admin Port Assignment Strategy**: Each product family gets unique admin port to prevent conflicts in unified deployments.

**Windows Firewall Prevention - Tests Only**: Unit/integration tests MUST bind to 127.0.0.1 (NOT 0.0.0.0) to prevent Windows Firewall exception prompts during test automation. Docker containers MUST bind to 0.0.0.0 for container networking compatibility.

### Critical Rules

- ❌ **NEVER** create HTTP endpoints on ANY port
- ❌ **NEVER** use plain HTTP for health checks (always HTTPS with --no-check-certificate)
- ❌ **NEVER** expose admin endpoints on public port
- ✅ **ALWAYS** use HTTPS for both public and private endpoints
- ✅ **ALWAYS** bind private endpoints to 127.0.0.1 (not 0.0.0.0)
- ✅ **ALWAYS** implement proper TLS with self-signed certs minimum
- ✅ **ALWAYS** use `wget --no-check-certificate` for Docker health checks

## VI. CI/CD Workflow Requirements

### GitHub Actions Service Dependencies

**MANDATORY: All workflows running `go test` MUST include PostgreSQL service container**

Any workflow executing `go test` on packages that use database repositories (KMS sqlrepository, Identity domain) MUST configure PostgreSQL service:

```yaml
env:
  POSTGRES_HOST: localhost
  POSTGRES_PORT: 5432
  POSTGRES_NAME: cryptoutil_test
  POSTGRES_USER: cryptoutil
  POSTGRES_PASS: cryptoutil_test_password

services:
  postgres:
    image: postgres:18
    env:
      POSTGRES_DB: ${{ env.POSTGRES_NAME }}
      POSTGRES_PASSWORD: ${{ env.POSTGRES_PASS }}
      POSTGRES_USER: ${{ env.POSTGRES_USER }}
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
    ports:
      - 5432:5432
```

**Why Required**: Tests in sqlrepository/domain packages need PostgreSQL; without service = connection refused

**Affected Workflows**: ci-race, ci-mutation, ci-coverage, any workflow running `go test`

**Health Check**: pg_isready (10s interval, 5s timeout, 5 retries = 50s window)

**Details**: .github/instructions/02-01.github.instructions.md

### Service Health Check Requirements

**Docker Compose Health Check Configuration**:

All cryptoutil services in Docker Compose MUST use generous health check timeouts:

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
  start_period: 30s   # Grace period before first check
  interval: 5s        # Time between checks
  timeout: 3s         # Max time per check
  retries: 10         # Max failed checks before unhealthy
  # Total window: 30s + (5s × 10) = 80 seconds
```

**Why Generous**: TLS (1-2s) + DB migrations (5-10s) + unseal (1-2s) + OTLP (2-3s) = 10-20s typical, 40s worst-case

**Startup Order**: Config → TLS → DB pool → migrations → unseal → listeners → telemetry (background)

**Details**: docs/TIMEOUT-FIXES-ANALYSIS.md
7. Connect to OTLP telemetry (background, non-blocking)
8. Mark service as ready (/readyz returns HTTP 200)

## VII. Code Quality Excellence

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
- Mutation testing score ≥85% per package Phase 4, ≥98% per package Phase 5+ (gremlins or equivalent)
- ALWAYS fix all pre-commit hook errors; see ./.pre-commit-config.yaml
- ALWAYS fix all pre-commit hook errors; see ./.pre-commit-config.yaml
- All code builds  `go build ./...`, `mvn compile`
- All code changes pass `golangci-lint run --fix`
- All tests pass (`go test ./... -cover`)
- Coverage maintained at target thresholds, and gradually increased

## VIII. Development Workflow and Evidence-Based Completion

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

## VIII. Terminology Standards

**RFC 2119 Keywords** are used throughout this constitution and all specification documents:

- **MUST** / **REQUIRED** / **MANDATORY** / **CRITICAL**: Absolute requirement (all 4 terms are synonymous)
- **MUST NOT** / **SHALL NOT**: Absolute prohibition
- **SHOULD** / **RECOMMENDED**: Strong recommendation (exceptions require documented justification)
- **SHOULD NOT** / **NOT RECOMMENDED**: Strong discouragement (usage requires documented justification)
- **MAY** / **OPTIONAL**: Truly optional (implementer's choice)

**User Intent Clarification**: The terms MUST, REQUIRED, MANDATORY, and CRITICAL are intentionally treated as complete synonyms in this project. All four indicate an absolute, non-negotiable requirement with no exceptions.

**Source**: RFC 2119 "Key words for use in RFCs to Indicate Requirement Levels" + user clarification 2025-12-19

---

## IX. File Size Limits and Code Organization

**File Size Targets** (applies to ALL files: production code, tests, docs, configs):

- **Soft limit**: 300 lines (ideal target for optimal readability)
- **Medium limit**: 400 lines (acceptable with justification in PR)
- **Hard limit**: 500 lines (NEVER EXCEED - refactor required before merge)

**Rationale**:

- Faster LLM agent processing and token usage
- Easier human code review and maintenance
- Better code organization and discoverability
- Forces logical separation of concerns

**Refactoring Strategies When Approaching Limits**:

1. Split by functionality (create_test.go, validate_test.go, extract_test.go)
2. Split by algorithm type (rsa_test.go, ecdsa_test.go, eddsa_test.go)
3. Extract test helpers to *_test_util.go files
4. Move integration tests to *_integration_test.go files

**Service Template Requirement**:

- Phase 6 MUST extract reusable service template from proven implementations (KMS, JOSE, Identity)
- Template includes: Dual HTTPS servers, health checks, graceful shutdown, telemetry, middleware, config management
- Template parameterization: Constructor injection for handlers, middleware, configuration
- All new services MUST use extracted template (consistency, reduced code duplication)

**Learn-PS Demonstration Requirement**:

- Phase 7 MUST implement Learn-PS pet store demonstration service using extracted template
- Purpose: Validate template is truly reusable and production-ready
- Validates: Service stands up, passes health checks, handles requests, integrates with telemetry
- Success criteria: Learn-PS implementation <500 lines (proves template handles infrastructure)

---

## X. Hash Service Architecture and Versioning

**Hash Version Management** (Phase 5 deliverable):

- **Version = Date-Based Policy Revision**: v1 (2020 NIST), v2 (2023 NIST), v3 (2025 OWASP+)
- **Each version contains**: SHA-256/384/512 algorithm selection, PBKDF2 iterations, salt sizes, HKDF info strings
- **Algorithm Selection Within Version**: Based on input size (0-31 bytes→SHA-256, 32-47 bytes→SHA-384, 48+ bytes→SHA-512)
- **Configuration-Driven**: Versions stored in YAML config, not hardcoded in code
- **Hash Output Format**: Prefix format `{v}:base64_hash` (e.g., `{1}:abcd1234...`) for version-aware verification
- **Migration Strategy**: Support multiple versions concurrently during policy transitions

**Hash Registry Types** (4 types × 3 versions each = 12 configurations):

1. **Password Hashing**: PBKDF2-HMAC-SHA (non-deterministic, high entropy)
2. **PII Hashing**: HKDF-SHA (deterministic, searchable)
3. **OTP Hashing**: PBKDF2-HMAC-SHA (non-deterministic, high entropy, short-lived)
4. **Magic Link Hashing**: HKDF-SHA (deterministic, searchable, time-limited)

---

## XI. Governance and Documentation Standards

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
| 3.0.0 | 2025-12-19 | IV, V, VIII, IX, X | Phased mutation targets (85%→98%), test timing (<15s/<180s), probability-based execution, main() pattern, Windows Firewall prevention, admin port assignments, file size limits, service template, Learn-PS, hash versioning, terminology standards (MUST=REQUIRED=MANDATORY=CRITICAL) |

**Version**: 3.0.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-19
