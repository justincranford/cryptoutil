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
| Demo: Learn | 1 service | InstantMessenger demonstration service - encrypted messaging between users (validates service template reusability, crypto lib integration) | ✅ | ✅ |

### Service Catalog (9 Services Total)

| Service | Product | Public Ports | Admin Port | Status | Notes |
|---------|---------|--------------|------------|--------|-------|
| sm-kms | Secrets Manager | 8080-8089 | 9090 | ✅ COMPLETE | Reference implementation |
| pki-ca | PKI | 8443-8449 | 9090 | ⚠️ PARTIAL | Needs dual-server |
| jose-ja | JOSE | 9443-9449 | 9090 | ⚠️ PARTIAL | Needs dual-server |
| identity-authz | Identity | 18000-18009 | 9090 | ✅ COMPLETE | Dual servers |
| identity-idp | Identity | 18100-18109 | 9090 | ✅ COMPLETE | Dual servers |
| identity-rs | Identity | 18200-18209 | 9090 | ⏳ IN PROGRESS | Public server pending |
| identity-rp | Identity | 18300-18309 | 9090 | ❌ NOT STARTED | Reference implementation |
| identity-spa | Identity | 18400-18409 | 9090 | ❌ NOT STARTED | Reference implementation |
| learn-im | Learn | 8888-8889 | 9090 | ❌ NOT STARTED | Phase 3 validation |

**Implementation Priority**: sm-kms (✅) → template extraction (Phase 2) → learn-im (Phase 3, validates template) → jose-ja (Phase 4) → pki-ca (Phase 5) → identity services (authz ✅, idp ✅, rs ⏳, rp ❌, spa ❌, Phase 6)

**See**: `architecture.md` for complete service catalog and federation patterns

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
- **mTLS MUST implement BOTH CRLDP and OCSP for certificate revocation checking**
- **CRLDP MUST provide immediate revocation checks (NOT batched or delayed)**
- Rationale: Defense in depth - OCSP for online checks, CRLDP as fallback

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
- **Unseal secrets MUST support BOTH strategies**: (1) Key derivation from master secret, (2) Pre-generated JWKs stored in enclave
- **Pepper rotation MUST use lazy migration strategy**: Re-hash passwords ONLY on re-authentication (NOT batch migration)
- Rationale: Lazy migration avoids service downtime and preserves user sessions
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
- ✅ **Race detector MUST keep probabilistic execution enabled** (NOT disabled for performance)
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

**MANDATORY: All Services in All Products MUST support run as containers; this is preferred for production and end-to-end testing**

**MANDATORY: All Services in All Products MUST support Two HTTPS Endpoints**

Separate HTTPS endpoints for public operations vs private administration MUST be supported. TLS server certificate authentication MUST be enforced for both endpoints; TLS client certificate authentication may be enabled per endpoint, and set to either preferred or required, via configuration; HTTP is NEVER allowed.

### Deployment Environments

**Production Deployments**:

- Public endpoints MUST support configurable bind address (container default: 0.0.0.0, test/dev default: 127.0.0.1)
- Public endpoints bind address configuration pattern: `<configurable_address>:<configurable_port>`
- Container deployments typically use 0.0.0.0 IPv4 bind address (enables external access)
- Test/dev deployments typically use 127.0.0.1 IPv4 bind address (prevents Windows Firewall prompts)
- Private endpoints MUST ALWAYS use 127.0.0.1:9090 (never configurable, not mapped outside containers)
- No IPv6 inside containers: All endpoints must use IPv4 inside containers, due to dual-stack limitations in container runtimes (e.g. Docker Desktop for Windows)

**Development/Test Environments**:

For address binding:

- Public and private endpoints MUST use 127.0.0.1 IPv4 bind address (prevents Windows Firewall prompts)
- Rationale: 0.0.0.0 binding triggers Windows Firewall exception prompts, blocking automated execution of tests

For port binding:

- Public and private endpoints MUST use port 0 (dynamic allocation) to avoid port collisions
- Rationale: static ports cause port collisions during parallel test automation

### CA Architecture Pattern

#### TLS Issuing CA Configurations

Examples of Issuing CA Cert Chains, in order of highest to lowest preference:

- Offline Root CA -> Online Root CA -> Online Issuing CA
- Online Root CA -> Online Issuing CA
- Online Root CA
- Online Root CA -> Policy Root CA -> Online Issuing CA
- Offline Root CA -> Online Root CA -> Policy CA -> Online Issuing CA

### Two Endpoint HTTPS Architecture Pattern

#### TLS Certificate Configuration

All services in all products MUST support separate configuration for two HTTPS endpoints, including separate configuration for TLS Server and TLS client of each endpoint.

These main options MUST be supported via configuration from the point of view (POV) of the HTTPS Issuing CA certificate chain:

1. All Externally; useful for production, where HTTPS Issuing CA certificate chain is provided without private key, and the HTTPS Server certificate chain is provided with private key
2. Mixed External and Auto-Generated; where Issuing CA chain is provided with private key, and HTTPS Server certificate and private key are generated and signed by the Issuing CA; useful for per-product development and testing
3. All Auto-Generated; where HTTPS Server certificate chain and private key are all generated by the service instance; useful for standalone service development and testing, not production

HTTPS Issuing CA certificate chains for TLS Server MAY BE per-suite, per-product, or per-service type.

HTTPS Issuing CA certificate chains for TLS Client MUST BE per-service type (preferred).

**Static TLS Certificates** (Externally Provided):

These main options MUST be supported via configuration from the point of view (POV) of the HTTPS Server certificate chain and private key:

- Private keys stored in Docker Secrets (production and development)
- Certificate chains provided via file paths or PEM-encoded data in configuration files
- Trusted CA certificates configurable for client verification
- Required for production deployments with organizational PKI

HTTPS Issuing CA for TLS Server Certs SHOULD BE shared across all products (preferred), all services of a product, per-service instance type, or per-service instance.

HTTPS Issuing CA for TLS Client Certs MUST BE shared per per-service instance type.

#### 2. Private HTTPS Endpoint

**Purpose**: Administration, health checks, graceful shutdown

**Configuration**:

- Production port: 127.0.0.1:9090 (static binding)
- Test port: 0 (dynamic allocation)
- Bind address: ALWAYS 127.0.0.1 (IPv4 loopback only)
- TLS: HTTPS is MANDATORY (never HTTP)
- External access: NEVER (127.0.0.1-only)

**Endpoints**:

- `/admin/v1/livez` - Liveness probe (lightweight check: service running, process alive)
- `/admin/v1/readyz` - Readiness probe (heavyweight check: dependencies healthy, ready for traffic)
- `/admin/v1/shutdown` - Graceful shutdown trigger

**Health Check Semantics**:

- **livez**: Fast, lightweight check (~1ms) - verifies process is alive, TLS server responding
- **readyz**: Slow, comprehensive check (~100ms+) - verifies database connectivity, downstream services, resource availability
- **Use livez for**: Docker healthchecks (fast, frequent), liveness probes (restart on failure)
- **Use readyz for**: Kubernetes readiness probes (remove from load balancer), deployment validation

**Consumers**: Docker health checks, Kubernetes probes, monitoring systems, orchestration tools

#### 3. Public HTTPS Endpoint (Public Server)

**Purpose**: Business APIs, browser UIs, external client access

**Terminology Clarification** (Two HTTPS Endpoint Strategy):

- **Admin Port**: Non-exposed, ALWAYS 127.0.0.1:9090 (never configurable)
- **Exported Port**:
  - **Inside Container**: 0.0.0.0:8080 (container default, enables external access)
  - **Outside Container**:
    - **Tests**: Mapped to 127.0.0.1 with unique static port range per service (prevents Windows Firewall prompts, avoids port collisions)
    - **Production**: Mapped to `<configurable_address>:<configurable_port>` with unique static port range per service

**Configuration**:

- Production ports: Service-specific ranges (8080-8089 for KMS, 8180-8189 for Identity, etc.)
- Bind address: Configurable (container default: 0.0.0.0, test/dev default: 127.0.0.1)
- Pattern: `<configurable_address>:<configurable_port>` (NEVER hardcode 0.0.0.0 in documentation)
- TLS: HTTPS MANDATORY (never HTTP)
- External access: YES (exposed to clients)

**Request Path Prefixes and Middlewares**:

For public HTTPS endpoint, all services implement TWO security middleware stacks, which reuse an OpenAPI specification per service but enforce different authentication, authorization, and access control policies based on request path prefixes:

**Service-to-Service APIs** (`/service/**` prefix):

- Access: Service clients ONLY (browsers blocked by middleware)
- Middleware: IP allowlist, rate limiting, request logging

**Browser-to-Service APIs/UI** (`/browser/**` prefix):

- Access: Browser clients ONLY (service clients blocked by middleware)
- Middleware: CSRF protection, CORS policies, CSP headers, IP allowlist, rate limiting
- Additional content: HTML pages, JavaScript, CSS, images, fonts, etc.

**API Consistency**:

- SAME OpenAPI specification served at both `/service/**` and `/browser/**` paths
- Middleware enforces mutual exclusivity (service tokens can't access browser paths, vice versa)
- Prevents unauthorized cross-client access patterns

**Port Allocation**: See service catalog table above for port ranges. Admin ports: 9090 (KMS), 9091 (Identity), 9092 (CA), 9093 (JOSE), 9095 (Learn)

**Windows Firewall Prevention**: Tests bind 127.0.0.1 (not 0.0.0.0). See `https-ports.md` for complete binding patterns

### Critical Rules

- ❌ **NEVER** create HTTP endpoints on ANY port
- ❌ **NEVER** use plain HTTP for health checks (always HTTPS with --no-check-certificate)
- ❌ **NEVER** expose admin endpoints on public port
- ✅ **ALWAYS** use HTTPS for both public and private endpoints
- ✅ **ALWAYS** bind private endpoints to 127.0.0.1 (not 0.0.0.0)
- ✅ **ALWAYS** implement proper TLS with self-signed certs minimum
- ✅ **ALWAYS** use `wget --no-check-certificate` for Docker health checks

## VA. Service Federation and Discovery - CRITICAL

**MANDATORY: Services MUST support configurable federation for cross-service communication**

### Authentication and Authorization Architecture - CRITICAL

**Source**: QUIZME-02 answers (Q1-Q2, Q3-Q4, Q6-Q7, Q10-Q12, Q15)
**Reference**: See `.specify/memory/authn-authz-factors.md` for authoritative authentication/authorization factor list

**Single Factor Authentication Methods** (SFA):

- **Headless-Based Clients** (`/service/*` paths): 10 methods (3 non-federated + 7 federated)
- **Browser-Based Clients** (`/browser/*` paths): 28 methods (6 non-federated + 22 federated)
- **Complete list with per-factor storage realms**: `.specify/memory/authn-authz-factors.md`

**Storage Realm Pattern**:

- **YAML + SQL (Config > DB priority)**: Static credentials, provider configs (enables service start without database)
- **SQL ONLY**: User-specific enrollment data, one-time tokens/codes (dynamic per-user)
- **Details**: See `.specify/memory/authn-authz-factors.md` Section "Storage Realm Specifications"

**Multi-Factor Authentication (MFA)**:

- Combine 2+ single factors (e.g., Password + TOTP, Client ID/Secret + mTLS)
- Common combinations: See `.specify/memory/authn-authz-factors.md` Section "Multi-Factor Authentication"

**Authorization Methods**:

- Headless-Based: Scope-Based Authorization, Role-Based Access Control (RBAC)
- Browser-Based: Scope-Based Authorization, RBAC, Resource-Level Access Control, Consent Tracking (scope+resource tuples)

**Session Token Format** (Q3 - Configuration-Driven):

- Non-Federated mode: Product-specific config determines format (opaque, JWE, JWS)
- Federated mode: Identity Provider config determines format
- Admin configures via YAML/environment variable

**Session Storage Backend** (Q4 - PostgreSQL/SQLite Only):

- SQLite: Single-node deployments
- PostgreSQL: Distributed/high-availability deployments with shared session data
- **CRITICAL: Session state MUST use SQL database only (NO Redis)**
- **Session Token Formats**: JWS (signed), OPAQUE (database reference), JWE (encrypted)
- Rationale: SQL ensures ACID compliance, transaction support, consistent backups

**MFA Step-Up Authentication** (Q6 - Time-Based):

- Re-authentication MANDATORY every 30 minutes for sensitive resources
- Applies regardless of operation type
- Session remains valid for low-sensitivity operations

**MFA Enrollment Workflow** (Q7 - Optional with Limited Access):

- OPTIONAL enrollment during initial setup
- Access LIMITED until additional factors enrolled
- Only one identifying factor required for initial login

**Realm Failover Behavior** (Q10 - Priority List):

- Admin configures priority list of Realm+Type tuples
- System tries each in priority order until success or all fail
- Example: `[(File, YAML), (Database, PostgreSQL), (Database, SQLite)]`

**Zero Trust Authorization** (Q11 - No Caching):

- Authorization MUST be evaluated on EVERY request
- NO caching of authorization decisions (prevents stale permissions)
- Performance via efficient policy evaluation, not caching

**Cross-Service Authorization** (Q12 - Direct Token Validation):

- Session Token passed between federated services
- Each service independently validates token and enforces authorization
- NO token transformation or delegation

**Consent Tracking Granularity** (Q15 - Scope+Resource Tuples):

- Tracked as (scope, resource) tuples
- Example: ("read:keys", "key-123") separate from ("read:keys", "key-456")

### Federation Architecture

Services discover and communicate with other cryptoutil services via **configuration** (NEVER hardcoded URLs):

```yaml
# Example KMS federation configuration
federation:
  # Identity service for OAuth 2.1 authentication
  identity_url: "https://identity-authz:8180"
  identity_enabled: true
  identity_timeout: 10s  # MUST be per-service configurable

  # JOSE service for external JWE/JWS operations
  jose_url: "https://jose-server:8280"
  jose_enabled: true
  jose_timeout: 10s  # MUST be per-service configurable

  # CA service for TLS certificate operations
  ca_url: "https://ca-server:8380"
  ca_enabled: false  # Optional - KMS can use internal TLS certs
  ca_timeout: 10s  # MUST be per-service configurable

# Federation Requirements (MANDATORY)
# - Timeouts MUST be per-service configurable (identity_timeout, jose_timeout, ca_timeout)
# - API versioning MUST support N-1 backward compatibility (rolling upgrades)
# - DNS caching MUST be disabled (lookup every request for dynamic service discovery)

# Graceful degradation settings
federation_fallback:
  # When identity service unavailable
  identity_fallback_mode: "local_validation"  # or "reject_all", "allow_all" (dev only)

  # When JOSE service unavailable
  jose_fallback_mode: "internal_crypto"  # Use internal JWE/JWS implementation

  # When CA service unavailable
  ca_fallback_mode: "self_signed"  # Generate self-signed TLS certs
```

### Service Discovery Mechanisms

**Configuration File** (Preferred for static deployments):

- Explicit URLs in YAML configuration files
- Example: `federation.identity_url: "https://identity.example.com:8180"`

**Docker Compose Service Names**:

- Docker networks provide automatic DNS resolution
- Example: `federation.identity_url: "https://identity-authz:8180"` (service name from compose.yml)

**Kubernetes Service Discovery**:

- Kubernetes DNS provides automatic service resolution
- Example: `federation.identity_url: "https://identity-authz.cryptoutil-ns.svc.cluster.local:8180"`

**Environment Variables** (Overrides config file):

- `CRYPTOUTIL_FEDERATION_IDENTITY_URL="https://identity:8180"`
- `CRYPTOUTIL_FEDERATION_JOSE_URL="https://jose:8280"`

### Graceful Degradation Patterns

**Circuit Breaker**: Automatically disable federated service after N consecutive failures

- Failure threshold: Open circuit after N failures
- Timeout: Reset circuit after timeout period
- Half-open requests: Test N requests before closing circuit

**Fallback Modes** (MANDATORY Production Requirements per CLARIFY-QUIZME-01 Q7, 2025-12-22):

- **Identity Unavailable Production**: `reject_all` (strict mode) - MANDATORY for production deployments (security over availability)
- **Identity Unavailable Development**: `allow_all` - Acceptable for development only (convenience)
- **Identity Unavailable BANNED**: `local_validation` (cached public keys) - NOT allowed in production (risk of stale cached keys)
- **JOSE Unavailable**: Internal crypto implementation (use service's own JWE/JWS)
- **CA Unavailable**: Self-signed TLS certificates (development), cached certificates (production)

**Rationale**: If Identity service is down, it's better to reject traffic than risk unauthorized access with stale cached credentials.

**Retry Strategies**:

- **Exponential Backoff**: 1s, 2s, 4s, 8s, 16s (max 5 retries)
- **Timeout Escalation**: Increase timeout 1.5x per retry (10s → 15s → 22.5s)
- **Health Check Before Retry**: Poll `/admin/v1/healthz` endpoint before resuming traffic

### Federation Health Monitoring

**Regular Health Checks**:

- Check federated service health every 30 seconds
- Log warnings when federated services become unhealthy
- Activate fallback mode when health checks fail

**Metrics and Alerts**:

- `federation_request_duration_seconds{service="identity"}` - Latency tracking
- `federation_request_failures_total{service="identity"}` - Error rate
- `federation_circuit_breaker_state{service="identity"}` - Circuit state (closed/open/half-open)

### Cross-Service Authentication

**Service-to-Service mTLS** (Preferred):

```yaml
federation:
  identity_url: "https://identity-authz:8180"
  identity_client_cert: "file:///run/secrets/kms_client_cert"
  identity_client_key: "file:///run/secrets/kms_client_key"
  identity_ca_cert: "file:///run/secrets/identity_ca_cert"
```

**OAuth 2.1 Client Credentials** (Alternative):

```yaml
federation:
  identity_url: "https://identity-authz:8180"
  identity_client_id: "kms-service"
  identity_client_secret: "file:///run/secrets/kms_client_secret"
  identity_token_endpoint: "https://identity-authz:8180/service/token"
```

### MFA Factor Priority and Implementation (MANDATORY)

**Decision Source**: CLARIFY-QUIZME-01 Q4, 2025-12-22

**ALL factors including deprecated ones MUST be implemented for backward compatibility**.

**MFA Factors** (in priority order, ALL MANDATORY for Phase 2+ Identity product):

1. **Passkey** (WebAuthn with discoverable credentials) - HIGHEST priority, FIDO2 standard, phishing-resistant
2. **TOTP** (Time-based One-Time Password) - HIGH priority, RFC 6238, authenticator apps (Google Authenticator, Authy)
3. **Hardware Security Keys** (WebAuthn without passkeys) - HIGH priority, FIDO U2F/FIDO2, phishing-resistant
4. **Email OTP** (One-Time Password via email) - MEDIUM priority, email delivery required, backup factor
5. **Recovery Codes** (Pre-generated backup codes) - MEDIUM priority, account recovery when primary factor unavailable
6. **SMS OTP** (NIST deprecated but MANDATORY) - MEDIUM priority, backward compatibility, accessibility for non-technical users
7. **Phone Call OTP** (NIST deprecated but MANDATORY) - LOW priority, backward compatibility, accessibility alternative
8. **Magic Link** (Time-limited authentication link via email/SMS) - LOW priority, passwordless alternative
9. **Push Notification** (Mobile app push-based approval) - LOW priority, requires mobile app integration

**Rationale**: Even though SMS OTP and Phone Call OTP are NIST deprecated (NIST SP 800-63B Revision 3), many organizations still rely on them for legacy compatibility and user accessibility (e.g., users without smartphones, users in low-tech environments).

**Implementation Notes**:

- Phase 2.1: Passkey, TOTP, Hardware Security Keys (core factors)
- Phase 2.2: Email OTP, Recovery Codes, SMS OTP (backward compatibility)
- Phase 2.3: Phone Call OTP, Magic Link, Push Notification (optional factors)

### Federation Testing Requirements

**Integration Tests MUST**:

- Test each federated service independently (mock others)
- Test graceful degradation when federated service unavailable
- Test circuit breaker behavior (failure thresholds, timeouts, recovery)
- Test retry logic (exponential backoff, max retries)
- Verify timeout configurations prevent cascade failures

**E2E Tests MUST**:

- Deploy full stack (all federated services)
- Test cross-service communication paths
- **Cover BOTH /service/* and /browser/* request paths (priority: /service/* first)**
- Test federation with Docker Compose service discovery
- Verify health checks detect service failures
- Test failover and recovery scenarios
- **Generated code exemptions**: OpenAPI client/server + GORM models + protobuf (excluded from coverage)

---

## VB. Performance, Scaling, and Resource Management

### Vertical Scaling

**Resource Limits** (Per-service configuration):

- CPU limits: 500m-2000m (0.5-2 CPU cores)
- Memory limits: 256Mi-1Gi (configurable per service)
- Connection pool sizing: Based on workload (PostgreSQL 10-50, SQLite 5)
- Concurrent request handling: Configurable (default: 100 concurrent requests)

**Resource Monitoring**:

- OTLP metrics: CPU usage, memory usage, goroutine count
- **Telemetry sampling**: All strategies in OTLP config (always-on, probabilistic, rate-limiting), tail-based sampling default
- Health checks: Resource exhaustion detection
- **Health check failure handling**:
  - **Kubernetes**: MUST remove from load balancer + restart pod
  - **Docker**: MUST mark container unhealthy + continue running (manual intervention required)
- Graceful degradation: Circuit breaker when resources depleted

### Horizontal Scaling

**Load Balancing Patterns**:

- **Layer 7 (HTTP/HTTPS)**: Use reverse proxy (nginx, Traefik, Envoy) for path-based routing
- **Layer 4 (TCP)**: Use TCP load balancer for raw connection distribution
- **DNS-based**: Round-robin DNS for simple load distribution
- **Service mesh**: Istio/Linkerd for advanced traffic management

**Session State Management for Horizontal Scaling**:

- **Stateless sessions** (Preferred): JWT tokens, no server-side storage
- **Sticky sessions**: Load balancer affinity based on session cookie
- **Distributed session store**: Redis cluster for shared session state
- **Database-backed sessions**: PostgreSQL with connection pooling

**Database Scaling Patterns**:

- **Read replicas**: NOT USED - all reads MUST go to primary database only (prevents stale data)
- **Connection pooling**: PgBouncer/pgpool-II for connection multiplexing
- **Database sharding**: MUST be implemented in Phase 4 with tenant ID partitioning strategy
- **Multi-tenancy**:
  - For PostgreSQL+SQLite: MUST use per-row tenant_id column (UUIDv4, FK to tenants.id) in all tables
  - For PostgreSQL only: MUST ALSO separate tenants into separate schemas (schema name 'tenant_UUID')
  - NEVER use row-level security (RLS) - per-row tenant_id provides sufficient isolation
- **Connection pools**: MUST be configurable and hot-reloadable without service restart
- **Caching**: Redis/Memcached for frequently accessed data (NOT for session state)

**Distributed Caching Strategy**:

- **Cache invalidation**: TTL-based expiration, event-driven invalidation
- **Cache consistency**: Write-through, write-behind, or cache-aside patterns
- **Cache tiers**: L1 (in-memory), L2 (Redis), L3 (database)

**Deployment Patterns**:

- **Blue-Green**: Zero-downtime deployments with instant rollback
- **Canary**: Gradual rollout to subset of users
- **Rolling updates**: Kubernetes-style progressive replacement

### Backup and Recovery

**Database Backup**:

- **PostgreSQL**: `pg_dump` for logical backups, `pg_basebackup` for physical backups
- **SQLite**: File-based backups (copy .db file)
- **Backup frequency**: Daily automated backups, retain 30 days
- **Backup validation**: Test restore procedure monthly

**Disaster Recovery**:

- **Database migrations**: Embedded SQL with golang-migrate provides schema versioning
- **Key rotation**: Version-based key management (KeyRing pattern) enables key recovery
- **Configuration backups**: YAML configs stored in version control
- **Recovery procedure**: Restore database from backup + apply migrations + restore keys from backup

**Documented in**:

- .github/instructions/01-06.database.instructions.md (database migrations)
- .github/instructions/01-09.cryptography.instructions.md (key versioning and rotation)

### Quality Tracking Documentation

**MANDATORY: Use docs/QUALITY-TODOs.md for coverage/gremlins challenges**

**Pattern for documenting quality improvements**:

```markdown
## Phase N: [Package Name]

### Priority 1: Critical Coverage Gaps (Target: 95%+)

**Package**: internal/[package]/[subpackage]
**Current Coverage**: X.X%
**Target Coverage**: 95.0%
**Gap**: Functions with <90% coverage

**Challenges**:
- Uncovered line ranges (file:lineStart-lineEnd)
- Reason for difficulty (e.g., error paths, edge cases, concurrency)

**Lessons Learned**:
- What worked (e.g., table-driven tests, property-based tests)
- What didn't work (e.g., mocking external dependencies)
- Recommendations for similar packages

### Priority 2: Mutation Testing Improvements (Target: 85%/98%)

**Package**: internal/[package]/[subpackage]
**Current Mutation Score**: X.X%
**Target Mutation Score**: 85.0% (Phase 4) / 98.0% (Phase 5+)
**Gap**: Mutants not killed

**Challenges**:
- Surviving mutants (specific mutation operators)
- Reason for difficulty (e.g., complex business logic, crypto operations)

**Lessons Learned**:
- Mutation-killing strategies (e.g., boundary value tests, error injection)
```

**Update docs/QUALITY-TODOs.md continuously** as challenges are discovered during implementation.

---

## VI. CI/CD Workflow Requirements

### GitHub Actions Service Dependencies

**MANDATORY: PostgreSQL workflows MUST use test-containers (NOT GitHub Actions service containers)**

**CRITICAL: GitHub Actions service containers NOT ALLOWED for PostgreSQL testing**

All workflows executing `go test` on packages that use database repositories MUST use test-containers library:

```go
// Example: PostgreSQL test container in TestMain
func TestMain(m *testing.M) {
    ctx := context.Background()
    postgresContainer, _ := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:18"),
        postgres.WithDatabase("cryptoutil_test"),
        postgres.WithUsername("cryptoutil"),
        postgres.WithPassword("cryptoutil_test_password"),
    )
    defer postgresContainer.Terminate(ctx)
    os.Exit(m.Run())
}
```

**Why test-containers Required**:

- Better test isolation (each package gets dedicated PostgreSQL instance)
- Automatic cleanup (containers terminated after tests)
- Local and CI parity (same test execution pattern)
- Avoids GitHub Actions service container limitations

**Docker Pre-Pull Requirements**:

- **ONLY for workflows that use Docker images** (e.g., E2E tests with Docker Compose)
- **NOT for unit/integration tests** (test-containers handles image pulling)
- Pre-pull step: `docker pull postgres:18 && docker pull otel/opentelemetry-collector-contrib:latest`

**Affected Workflows**: ci-race, ci-mutation, ci-coverage, any workflow running `go test`

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
- 95%+ production coverage, 98% infrastructure (cicd), 98% utility code
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

- Code evidence: `go build ./...` clean, `golangci-lint run` clean, coverage ≥95% (production), ≥98% (infrastructure/utility)
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
- [ ] Coverage targets maintained (95% production, 98% infrastructure/utility)
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

**Service Template Implementation Details**:

- **Admin Endpoints** (127.0.0.1:9090):
  - `/livez`, `/readyz`, `/shutdown` endpoints MANDATORY
  - Admin prefix MUST be configurable (default: `/admin/v1`)
  - Implementation: gofiber middleware (reference: sm-kms `internal/kms/server/application/application_listener.go`)
- **Health Check Requirements**:
  - OpenTelemetry Collector Contrib MUST use separate health check job (does NOT expose external health endpoint)
  - Reference implementation: KMS Docker Compose `deployments/compose/compose.yml` (working pattern)
- **Docker Secrets Validation**:
  - Docker Compose MUST include dedicated job to validate Docker Secrets presence and mounting
  - Fast-fail check before starting all other jobs and services (prevents cryptic runtime errors)

**Service Template Migration Priority** (HIGH PRIORITY):

**Decision Source**: CLARIFY-QUIZME-01 Q1, 2025-12-22

1. **learn-ps FIRST** (Phase 7):
   - CRITICAL: Implement learn-ps using service template
   - Iterative implementation, testing, validation, analysis
   - GUARANTEE ALL service template requirements met before migrating production services
   - Validates template is production-ready and truly reusable
2. **JOSE and CA NEXT** (one at a time, Phases 8-9):
   - MUST refactor JOSE (jose-ja) and CA (pki-ca) sequentially after learn-ps validation
   - Purpose: Drive template refinements to accommodate different service patterns
   - Identify and fix issues in service template to unblock remaining service migrations
   - Order: jose-ja → pki-ca (allow adjustments between migrations)
3. **Identity services LAST** (Phases 10-14):
   - MUST refactor identity services AFTER JOSE and CA migrations complete
   - Benefit from mature, battle-tested template refined by JOSE/CA migrations
   - Order: identity-authz → identity-idp → identity-rs → identity-rp → identity-spa
4. **sm-kms NEVER**:
   - KMS remains on current implementation indefinitely (reference implementation)
   - Only migrate KMS if ALL other 8 services running excellently on template
   - Prevents disrupting reference implementation

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

1. **LowEntropyRandomHashRegistry**: PBKDF2-HMAC-SHA (non-deterministic, salted, password hashing)
2. **LowEntropyDeterministicHashRegistry**: PBKDF2-HMAC-SHA (deterministic, fixed + derived salt, replay-resistant tokens)
3. **HighEntropyRandomHashRegistry**: HKDF-SHA (non-deterministic, salted, key derivation)
4. **HighEntropyDeterministicHashRegistry**: HKDF-SHA (deterministic, fixed + derived salt, deterministic key derivation)

**Pepper Requirements** (CRITICAL - ALL 4 Registries):

- **MANDATORY: All 4 hash registries MUST use pepper for additional security layer**
- **Pepper Storage** (NEVER store pepper in DB or source code):
  - VALID OPTIONS IN ORDER OF PREFERENCE: 1. Docker Secret, 2. Configuration file, 3. Environment variable
  - MUST be mutually exclusive from hashed values storage (pepper in secrets/config, hashes in DB)
  - MUST be associated with hash version (different pepper per version)
- **Pepper Rotation**:
  - Pepper CANNOT be rotated silently (requires re-hash all records)
  - Changing pepper REQUIRES version bump, even if no other hash parameters changed
  - Example: v1 pepper compromised → bump to v2 with new pepper, re-hash all v1 records
- **Additional Protections for LowEntropyDeterministicHashRegistry** (deterministic PII hashing):
  - MANDATORY (prevents deterministic hashing oracle attacks):
    - Query rate limits (prevent brute-force enumeration)
    - Abuse detection (detect suspicious query patterns)
    - Audit logs (track all hash queries for forensics)
    - Strict access control (limit who can query hashes)
  - RECOMMENDED: Apply same protections to all 4 registries for consistency

**Hash Registry Implementations**:

- **LowEntropyDeterministicHashRegistry** (PII Lookup) - ⚠️ Allowed with pepper + high cost:

  ```go
  hash = PBKDF2(
      input || pepper,
      fixedSalt,
      iterations = HIGH,  // Much higher than random registry
      outputLength = 256 bits
  )
  ```

  - Format: `{version}:base64(hash)`
  - Rationale: Deterministic for PII lookup, pepper prevents rainbow tables, high iteration cost mitigates brute force

- **HighEntropyDeterministicHashRegistry** (Config Blob Hash) - Good security:

  ```go
  PRK = HKDF-Extract(
      salt = fixedSalt,
      IKM = input || pepper
  )
  hash = HKDF-Expand(
      PRK,
      info = "config-blob-hash",
      L = 256 bits
  )
  ```

  - Format: `{version}:base64(hash)`
  - Rationale: High-entropy inputs, HKDF faster than PBKDF2, pepper provides domain separation

- **LowEntropyRandomHashRegistry** (Password Hashing) - Best practice:

  ```go
  hash = PBKDF2(
      password || pepper,
      randomSalt,
      iterations = OWASP_MINIMUM,
      outputLength = 256 bits
  )
  ```

  - Format: `{version}:{algorithm}:{iterations}:base64(randomSalt):base64(hash)`
  - Rationale: Random salt prevents rainbow tables, pepper adds secret key layer, OWASP iterations

- **HighEntropyRandomHashRegistry** (API Key Hashing) - Best practice:

  ```go
  PRK = HKDF-Extract(
      salt = randomSalt,
      IKM = apiKey || pepper
  )
  hash = HKDF-Expand(
      PRK,
      info = "api-key-hash",
      L = 256 bits
  )
  ```

  - Format: `{version}:{algorithm}:base64(randomSalt):base64(hash)`
  - Rationale: High-entropy inputs, HKDF faster than PBKDF2, random salt + pepper for defense in depth

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
5. **Documentation**: Update document with changes
6. **Communication**: Announce to all stakeholders

**Ratified**: 2025-12-01 | **Latest amendments**: 2025-12-22
