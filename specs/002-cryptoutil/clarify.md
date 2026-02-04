# cryptoutil Iteration 2 - Clarifications and Answers

**Last Updated**: December 24, 2025
**Purpose**: Authoritative Q&A for implementation decisions, architectural patterns, and technical trade-offs
**Organization**: Topical (consolidated from previous iterations)

---

## Table of Contents

1. [Architecture and Service Design](#architecture-and-service-design)
2. [Testing Strategy and Quality Assurance](#testing-strategy-and-quality-assurance)
3. [Cryptography and Hash Service](#cryptography-and-hash-service)
4. [Observability and Telemetry](#observability-and-telemetry)
5. [Deployment and Docker](#deployment-and-docker)
6. [CI/CD and Automation](#cicd-and-automation)
7. [Documentation and Workflow](#documentation-and-workflow)
8. [Authentication and Authorization](#authentication-and-authorization)
9. [Service Template and Migration Strategy](#service-template-and-migration-strategy)
10. [Federation and Service Integration (Session 2025-12-23)](#federation-and-service-integration-session-2025-12-23)

---

## Architecture and Service Design

### Dual-Server Architecture Pattern

**Q**: What is the dual-server architecture pattern and why is it mandatory?

**A**: ALL services MUST implement dual HTTPS endpoints:

**Public HTTPS Server** (`<configurable_address>:<configurable_port>`):

- Purpose: User-facing APIs and browser UIs
- Ports: 8080 (KMS), 8180-8184 (Identity services), 8280 (JOSE), 8380 (CA)
- Security: OAuth 2.1 tokens, CORS/CSRF/CSP, rate limiting, TLS 1.3+
- API contexts:
  - `/browser/api/v1/*` - Session-based (HTTP Cookie) for SPA
  - `/service/api/v1/*` - Token-based (HTTP Authorization header) for backends

**Private HTTPS Server** (Admin endpoints):

- Purpose: Internal admin tasks, health checks, metrics
- Admin Port: 127.0.0.1:9090 (ALL services, all instances)
- Security: IP restriction (localhost only), optional mTLS, minimal middleware
- Endpoints: `/admin/api/v1/livez`, `/admin/api/v1/readyz`, `/admin/api/v1/healthz`, `/admin/api/v1/shutdown`
- NOT exposed in Docker port mappings

**Rationale for Shared Admin Port**:

- Admin ports bound to 127.0.0.1 only (not externally accessible)
- Docker Compose: Each service instance = separate container with isolated network namespace
- Same admin port (9090) can be reused across ALL services without collision
- Multiple instances: Admin port 0 in all unit tests, Admin internal 9090 in Docker Compose, Admin unique external port mapping per instance if needed

**Implementation Status**:

- ✅ KMS: Complete reference implementation
- ✅ Identity AuthZ: Dual servers implemented (commit 04317efd 2025-12-21)
- ✅ Identity IdP: Dual servers implemented (commit 04317efd 2025-12-21)
- ✅ Identity RS: Public server implemented (commit 04317efd 2025-12-21)
- ❌ Identity RP: Not started
- ❌ Identity SPA: Not started
- ❌ JOSE: Missing admin server
- ❌ CA: Missing admin server

---

### Package Coverage Classification

**Q**: Which specific packages require 95% vs 98% coverage?

**A**: Case-by-case per package (document each in clarify.md)

**Initial Classification**:

- **Production (95%)**: internal/{jose,identity,kms,ca}
- **Infrastructure (98%)**: internal/cmd/cicd/*
- **Utility (98%)**: internal/shared/*, pkg/*

**Rationale**: Package complexity varies - some "production" packages have simpler logic warranting 98%, while some "utility" packages have complex error handling justifying 95%. Document each package's target in this clarify.md as implementation progresses.

**Documentation Pattern**:

- Add new entries to this section as packages are analyzed
- Justify any deviation from initial classification
- Update constitution.md if patterns emerge

---

### Service Federation Configuration

**Q**: How should services discover and configure federated services (Identity, JOSE)?

**A**: Services discover and communicate with other cryptoutil services via **configuration** (NEVER hardcoded URLs).

**Service Discovery Mechanisms**:

1. **Configuration File** (Preferred): Static YAML with explicit URLs

   ```yaml
   federation:
     identity_url: "https://identity.example.com:8180"
     jose_url: "https://jose.example.com:8280"
   ```

2. **Docker Compose**: Service names resolve via Docker network DNS

   ```yaml
   federation:
     identity_url: "https://identity-authz:8180"  # Service name from compose.yml
   ```

3. **Kubernetes**: Service discovery via cluster DNS

   ```yaml
   federation:
     identity_url: "https://identity-authz.cryptoutil-ns.svc.cluster.local:8180"
   ```

4. **Environment Variables** (Overrides config file):

   ```bash
   CRYPTOUTIL_FEDERATION_IDENTITY_URL="https://identity:8180"
   ```

**Graceful Degradation Patterns**:

**Circuit Breaker**: Automatically disable federated service after N consecutive failures

**Fallback Modes**:

- **Identity Unavailable**: Local token validation (cached public keys), reject all (strict), allow all (development only)
- **JOSE Unavailable**: Internal crypto implementation (use KMS's own JWE/JWS)
- **CA Unavailable**: Self-signed TLS certificates (development), cached certificates (production)

**Retry Strategies**:

- **Exponential Backoff**: 1s, 2s, 4s, 8s, 16s (max 5 retries)
- **Timeout Escalation**: Increase timeout 1.5x per retry (10s → 15s → 22.5s)
- **Health Check Before Retry**: Poll `/admin/api/v1/livez` endpoint (fast liveness check) before resuming traffic

---

### Connection Pool Sizing Configuration

**Q**: What is the REQUIRED formula for determining connection pool size?

**A**: Configurable values with hot-reloadable configuration (no fixed formula).

**Configuration Pattern**:

```yaml
database:
  driver: postgres
  connection_pool:
    max_open_connections: 25      # Maximum concurrent connections
    max_idle_connections: 10      # Idle connections kept alive
    connection_max_lifetime: 3600s  # Connection reuse limit (1 hour)
    connection_max_idle_time: 300s  # Idle timeout (5 minutes)
```

**Recommended Starting Values**:

- **PostgreSQL**: max_open=25, max_idle=10
- **SQLite**: max_open=5, max_idle=1 (single-writer limitation)
- **High-traffic deployments**: Increase max_open based on observed contention

**Hot Reload Support**:

- Configuration changes apply without service restart
- Monitor connection pool metrics (queue depth, wait time, utilization)
- Adjust based on production telemetry data

**Monitoring Metrics** (OpenTelemetry):

```yaml
metrics:
  - db.connection_pool.max_open
  - db.connection_pool.max_idle
  - db.connection_pool.in_use
  - db.connection_pool.idle
  - db.connection_pool.wait_count
  - db.connection_pool.wait_duration
```

**Rationale**: Connection pool sizing depends on workload patterns, hardware resources, and database capabilities. Configurable values with hot reload enable production tuning without downtime. Fixed formulas over-simplify complex trade-offs between connection overhead and query concurrency.

---

### Read Replica Strategy

**Q**: What replication lag is acceptable and when should reads fall back to primary?

**A**: Remove read replicas entirely - all reads go to primary only.

**Architecture Decision**:

- All read queries directed to primary database
- No read replica configuration or routing logic
- Simplifies architecture and eliminates replication lag concerns

**Rejected Pattern**:

```yaml
# ❌ NOT IMPLEMENTED
database:
  primary:
    dsn: "postgres://primary:5432/cryptoutil"
  read_replicas:
    - dsn: "postgres://replica1:5432/cryptoutil"
    - dsn: "postgres://replica2:5432/cryptoutil"
  max_replication_lag: 5s
```

**Rationale for Removal**:

- Read replicas add operational complexity (monitoring lag, failover, consistency)
- Cryptoutil services are write-heavy (sessions, audit logs, key operations)
- Read-heavy workloads better served by caching layers (if needed in future)
- Database sharding (Phase 4) provides better scalability than read replicas
- Connection pooling sufficient for read concurrency in Phases 1-3

**Future Consideration**:

- If read performance becomes bottleneck: Add caching layer (Redis/Memcached) for specific read-heavy queries
- If write scalability needed: Implement database sharding (Phase 4 with multi-tenancy)

---

### Service Template Extraction

**Q**: When should the service template be extracted and how should it be validated?

**A**: Extract template in Phase 2, validate with cipher-im demonstration service in Phase 3.

**Template Components** (extracted from KMS reference implementation):

- Two HTTPS servers (public + admin)
- Two public API paths (`/browser/api/v1/*` vs `/service/api/v1/*`)
- Three admin endpoints (`/admin/api/v1/livez`, `/admin/api/v1/readyz`, `/admin/api/v1/shutdown`)
- Database abstraction (PostgreSQL || SQLite dual support, GORM)
- OpenTelemetry integration (OTLP traces, metrics, logs)
- Config management (YAML files + CLI flags, Docker secrets support)

**Template Parameterization**:

- Constructor injection for configuration, handlers, middleware
- Service-specific OpenAPI specs passed to template
- Business logic separated from infrastructure concerns

**Validation Strategy**:

- cipher-im service MUST use extracted template
- cipher-im MUST pass all unit/integration/E2E tests
- Deep analysis MUST show no blockers to migrate existing services
- Only after cipher-im succeeds can production services migrate

**Migration Priority**:

1. **cipher-im FIRST** (Phase 3) - Validate template reusability
2. **One service at a time** - Sequentially refactor jose-ja (Phase 4), pki-ca (Phase 5), identity services (Phase 6+)
3. **sm-kms LAST** - Only after ALL other services running excellently on template

---

### Session State Management for Horizontal Scaling

**Q**: Which session state management pattern(s) should be implemented for horizontal scaling?

**A**: SQL database-backed sessions ONLY (JWS/OPAQUE/JWE formats), NO Redis, NO sticky sessions.

**Implementation Priority** (all stored in SQL database):

1. **JWS Sessions** (HIGHEST priority): Stateless token validation, cryptographic signature verification
2. **OPAQUE Sessions** (MEDIUM priority): Database lookup for every request, maximum revocation control
3. **JWE Sessions** (LOWER priority): Encrypted session data, requires decryption on every request

**Deployment Priority** (security preference):

1. **JWE Sessions** (HIGHEST security): Encrypted session data prevents inspection
2. **OPAQUE Sessions** (MEDIUM security): Database-backed with immediate revocation
3. **JWS Sessions** (LOWER security): Signed but not encrypted, readable session data

**Rejected Patterns**:

- ❌ Redis cluster: Adds operational complexity, NOT supported
- ❌ Sticky sessions: Load balancer affinity, prevents true horizontal scaling
- ❌ Stateless-only: All session formats stored in SQL for auditability and revocation

**Rationale**: Database-backed sessions (PostgreSQL/SQLite) provide consistent storage, auditability, and revocation capabilities across all three token formats. Implementation priority favors simplest-to-implement (JWS) first, deployment priority favors most secure (JWE) for production.

---

### Database Sharding Strategy

**Q**: When should database sharding be implemented and what partition strategy should be used?

**A**: Phase 4 with multi-tenancy features, partition by tenant ID.

**Implementation Plan**:

- **Timeline**: Phase 4 (alongside multi-tenancy implementation)
- **Partition Strategy**: Tenant ID-based sharding
- **Rationale**: Aligns with multi-tenancy isolation, natural partition boundary

**Deferred Until Phase 4**:

- Read replicas and connection pooling sufficient for Phases 1-3
- Sharding adds complexity only justified by multi-tenant deployments
- Partition by tenant ID provides clean isolation and scalability

---

### Multi-Tenancy Isolation Pattern

**Q**: Which multi-tenancy isolation pattern MUST be implemented?

**A**: Dual-layer isolation - per-row tenant_id (all DBs) + schema-level (PostgreSQL only).

**Implementation Pattern**:

**Layer 1: Per-Row Tenant ID** (PostgreSQL + SQLite):

```sql
CREATE TABLE tenants (
  id UUID PRIMARY KEY,  -- UUIDv4
  name TEXT NOT NULL
);

CREATE TABLE users (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  username TEXT NOT NULL
);

CREATE TABLE sessions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  user_id UUID NOT NULL
);
```

**Layer 2: Schema-Level Isolation** (PostgreSQL only):

```sql
-- Tenant A
CREATE SCHEMA tenant_a;
CREATE TABLE tenant_a.users (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL CHECK (tenant_id = 'UUID-for-tenant-a'),
  username TEXT NOT NULL
);

-- Tenant B
CREATE SCHEMA tenant_b;
CREATE TABLE tenant_b.users (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL CHECK (tenant_id = 'UUID-for-tenant-b'),
  username TEXT NOT NULL
);
```

**Rejected Patterns**:

- ❌ Schema-level isolation ONLY: Doesn't work for SQLite (no schema support)
- ❌ Row-Level Security (RLS): Adds query complexity and performance overhead
- ❌ Database-level isolation: Too heavy-weight for multi-tenancy

**Rationale**: Per-row tenant_id works everywhere (PostgreSQL + SQLite). Schema-level adds defense-in-depth for PostgreSQL deployments. Both layers together prevent tenant data leakage.

---

## Testing Strategy and Quality Assurance

### Coverage Targets by Package Type

**Q**: What are the exact coverage targets and how strictly are they enforced?

**A**: Coverage targets are MANDATORY with NO EXCEPTIONS.

**Coverage Targets**:

- **Production packages** (internal/{jose,identity,kms,ca}): ≥95%
- **Infrastructure packages** (internal/cmd/cicd/*): ≥98%
- **Utility packages** (internal/shared/*, pkg/*): ≥98%
- **Main functions**: 0% acceptable if internalMain() ≥95%

**Enforcement Pattern**:

```bash
# ❌ WRONG: Celebrate improvement without meeting target
coverage_before=60.0
coverage_after=70.0
echo "✅ Improved by 10 percentage points!"  # Still 25 points below target

# ✅ CORRECT: Enforce target, reject anything below 95%
if [ "$coverage" -lt 95 ]; then
    echo "❌ BLOCKING: Coverage $coverage% < 95% target"
    echo "Required: Write tests for RED lines in coverage HTML"
    exit 1
fi
```

**Why "No Exceptions" Rule Matters**:

- Accepting 70% because "it's better than 60%" leaves 25 points of technical debt
- "This package is mostly error handling" → Add error path tests
- "This is just a thin wrapper" → Still needs 95% coverage
- Incremental improvements accumulate debt; enforce targets strictly

---

### main() Function Pattern for Maximum Coverage

**Q**: How should main() functions be structured to achieve coverage targets?

**A**: ALL main() functions MUST be thin wrappers calling co-located testable functions.

**Pattern** (MANDATORY for ALL commands):

```go
// CORRECT - Thin main() delegates to testable internalMain()
func main() {
    os.Exit(internalMain(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// internalMain is testable - accepts injected dependencies
func internalMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    // All logic here - fully testable with mocks
    if len(args) < 2 {
        fmt.Fprintln(stderr, "usage: cmd <arg>")
        return 1
    }
    // ... business logic
    return 0
}
```

**Why This Pattern is MANDATORY**:

- **95%+ coverage achievable**: main() 0% is acceptable when internalMain() is 95%+
- **Dependency injection**: Tests inject mocks for args, stdin, stdout, stderr
- **Exit code testing**: Tests verify return codes without terminating test process
- **Happy/sad path testing**: Test all branches (missing args, invalid input, success cases)

**Testing Pattern**:

```go
func TestInternalMain_HappyPath(t *testing.T) {
    args := []string{"cmd", "arg1"}
    stdin := strings.NewReader("")
    stdout := &bytes.Buffer{}
    stderr := &bytes.Buffer{}

    exitCode := internalMain(args, stdin, stdout, stderr)

    require.Equal(t, 0, exitCode)
    require.Contains(t, stdout.String(), "success")
}
```

---

### Test Execution Time Targets

**Q**: What are the time limits for test execution?

**A**: Strict timing targets with probabilistic execution for large test suites.

**Test Execution Time Targets**:

- **MANDATORY**: <15 seconds per unit test package
- **MANDATORY**: <180 seconds (3 minutes) for full unit test suite
- Integration/E2E tests excluded from strict timing (Docker startup overhead acceptable)
- Probabilistic execution MANDATORY for packages approaching 15s limit

**Probability-Based Test Execution**:

Use probability-based execution for table-driven tests with algorithm/key size variants:

**Magic Constants** (defined in `internal/shared/magic/magic_testing.go`):

- `TestProbAlways` (100%) - Base algorithms (RSA2048, AES256, ES256)
- `TestProbQuarter` (25%) - Important variants (RSA3072, AES192, ES384)
- `TestProbTenth` (10%) - Comprehensive variants (RSA4096, AES128, ES521)

**When to Use Probability-Based Testing**:

- ✅ Multiple key sizes of same algorithm (RSA 2048/3072/4096, AES 128/192/256)
- ✅ Multiple variants of same operation (HMAC-SHA256/384/512, ECDSA P-256/384/521)
- ✅ Large test suites (>50 test cases with redundant coverage)
- ❌ Fundamentally different algorithms (RSA vs ECDSA vs EdDSA - always test all)
- ❌ Business logic branches (error paths, edge cases - always test all)
- ❌ Small test suites (<20 cases - overhead not worth it)

---

### GitHub Actions Performance Considerations

**Q**: How should timeouts be configured for GitHub Actions vs local development?

**A**: Apply 2.5-3× multiplier to local timings for GitHub Actions.

**Performance Multipliers**:

- **Typical**: GitHub Actions 2.5-3.3× slower than local development
- **Extreme Cases**: Up to 150× slower for certain operations
- **Root Cause**: Shared CPU resources, network latency, cold starts, container overhead

**Timing Strategy**:

**Local Development**:

- Fast iteration with minimal timeouts
- Unit tests: 1-5s typical per package
- Network operations: 2-5s typical

**GitHub Actions**:

- Apply 2.5-3× multiplier minimum to local timings
- Add 50-100% safety margin for reliability
- Unit tests: 5-15s per package target
- Network operations: 5-10s (general), 10-15s (TLS handshakes)
- Health checks: 300s (5 minutes) for full Docker Compose stack

---

### Mutation Testing Requirements

**Q**: What are the mutation testing targets and how should gremlins be executed?

**A**: Phased mutation targets with package-level parallelization.

**Mutation Testing Targets**:

- **Phase 4**: ≥85% gremlins score per package
- **Phase 5+**: ≥98% gremlins score per package

**Recommended Configuration** (`.gremlins.yaml`):

```yaml
threshold:
  efficacy: 85  # Phase 4 target, raise to 98 in Phase 5+
  mutant-coverage: 90  # Target: ≥90% mutant coverage

workers: 4  # Parallel mutant execution
test-cpu: 2  # CPU per test run
timeout-coefficient: 2  # Timeout multiplier
```

**Optimization Strategies**:

- **Package-level parallelization**: Run gremlins on packages concurrently using GitHub Actions matrix strategy
- **Per-package timeout**: Fail fast for slow packages (prevents CI/CD blocking)
- **Exclude tests, generated code, vendor directories** for <20min total execution
- Focus on business logic, parsers, validators, crypto operations

---

### Race Detector with Probabilistic Test Execution

**Q**: Should race detector workflow disable probabilistic execution and accept longer runtimes?

**A**: Keep probabilistic execution enabled - accept that some race conditions may not be caught in single run.

**Race Detector Behavior**:

- Race detector adds ~10× runtime overhead
- Probabilistic tests (TestProbQuarter, TestProbTenth) still active
- Accept longer per-package timeout (150s vs 15s for unit tests)
- Different test cases execute on each run (shuffle + probability)

**Trade-offs**:

- ✅ Stays under GitHub Actions time limits (even with 10× overhead)
- ✅ Different code paths tested across multiple CI runs
- ⚠️ Some race conditions may not appear in specific run
- ❌ Disabling probabilistic execution → >25 minutes per workflow (unacceptable)

**Mitigation**:

- Run race detector on every commit (broad coverage over time)
- Shuffle tests for different execution orders
- Local developers can disable probabilistic execution for deep race analysis

**Rationale**: Maintaining <15 minute CI workflows is critical for development velocity. Probabilistic execution with race detector still provides valuable coverage across multiple runs. Organizations requiring 100% deterministic race detection can run exhaustive local tests.

---

### E2E Test API Path Coverage

**Q**: Which API paths MUST be covered by E2E tests for each product?

**A**: Test BOTH `/service/api/v1/*` and `/browser/api/v1/*` paths with separate E2E scenarios per path type.

**Priority Order**:

1. **`/service/api/v1/*` paths** (HIGHEST priority):
   - Headless client authentication (Bearer tokens, mTLS, Client ID/Secret)
   - Simpler test implementation (no browser automation)
   - Backend-to-backend integration scenarios
   - Examples: API key validation, certificate issuance, JOSE signing

2. **`/browser/api/v1/*` paths** (MEDIUM priority):
   - Browser client authentication (session cookies, OAuth flows)
   - Full middleware stack (CORS, CSRF, CSP)
   - Browser automation required (Playwright/Puppeteer)
   - Examples: OAuth authorization code flow, OIDC authentication, SPA session management

**Coverage Requirements**:

- ALL products MUST test BOTH path types eventually
- Initial E2E (Phase 2): `/service/*` paths only for JOSE, CA, KMS
- Expanded E2E (Phase 3+): Add `/browser/*` paths for all products
- Identity product: `/service/*` AND `/browser/*` in same phase (OAuth flows require browser)

**Rationale**: Dual API paths serve fundamentally different client types with different authentication patterns. Both MUST be tested to ensure complete coverage. `/service/*` priority enables faster initial E2E implementation.

---

### Mutation Testing Generated Code Exemption

**Q**: What code qualifies as "generated code" eligible for mutation testing exemption?

**A**: OpenAPI-generated code + GORM auto-migration code + protobuf (if used).

**Exemption-Eligible Code**:

1. **OpenAPI-generated models/clients** (oapi-codegen output):
   - `api/model/*` - OpenAPI schema-generated Go structs
   - `api/client/*` - OpenAPI-generated HTTP client code
   - `api/server/*` - OpenAPI-generated server interfaces

2. **GORM auto-migration code**:
   - `internal/*/models/*_gen.go` - GORM code generation output
   - Database migration files if generated

3. **Protobuf-generated code** (if used):
   - `*.pb.go` - protoc compiler output
   - `*_grpc.pb.go` - gRPC service definitions

**NOT Exemption-Eligible**:

- Hand-written business logic (even if simple)
- Test helpers and utilities
- Configuration parsers
- Third-party libraries (vendor/ directory already excluded)

**Ramp-Up Requirement**:

- Generated code MAY start below 85% mutation coverage
- MUST have plan to reach ≥85% through:
  - Additional integration tests exercising generated code paths
  - Custom validation logic for edge cases
  - Error handling tests for generated error paths
- Document exemptions in clarify.md with justification and timeline

**Rationale**: Generated code often has boilerplate that's difficult to meaningfully test with mutations. However, generated code still needs sufficient testing through integration tests that exercise the full generated code paths.

---

## Cryptography and Hash Service

### Hash Registry Architecture

**Q**: What is the hash registry architecture and why are there four registries?

**A**: Four registries based on entropy level and determinism requirements.

**Supported Registries**:

1. **LowEntropyDeterministicHashRegistry** - PII lookup (searchable, no decryption)
   - Use case: Username/email lookup, IP address tracking
   - Algorithm: PBKDF2(input || pepper, fixedSalt, HIGH_iterations, 256)
   - Protection: Query rate limits, abuse detection, audit logs

2. **LowEntropyRandomHashRegistry** - Password hashing (non-searchable, no decryption)
   - Use case: Password verification
   - Algorithm: PBKDF2(password || pepper, randomSalt, OWASP_MIN_iterations, 256)
   - Protection: Random salt per password, pepper in secrets

3. **HighEntropyDeterministicHashRegistry** - Config blob hash (searchable, no decryption)
   - Use case: Configuration deduplication
   - Algorithm: HKDF-Extract(fixedSalt, input || pepper) → HKDF-Expand(PRK, "config-blob-hash", 256)
   - Protection: Fixed salt for determinism, pepper in secrets

4. **HighEntropyRandomHashRegistry** - API key hashing (non-searchable, no decryption)
   - Use case: API key verification
   - Algorithm: HKDF-Extract(randomSalt, apiKey || pepper) → HKDF-Expand(PRK, "api-key-hash", 256)
   - Protection: Random salt per key, pepper in secrets

**Entropy Threshold**: 128 bits entropy (256-bit search space)

---

### Hash Output Format and Versioning

**Q**: How are hashes formatted and versioned?

**A**: Version-based policy framework with tuple of (policy revision, pepper).

**Hash Output Format** (MANDATORY):

```
{version}:{algorithm}:{iterations}:base64(randomSalt):base64(hash)
```

**Examples**:

```
{1}:PBKDF2-HMAC-SHA256:rounds=600000:abcd1234...
{2}:PBKDF2-HMAC-SHA384:rounds=600000:efgh5678...
{3}:HKDF-SHA512:info=api-key,salt=xyz:ijkl9012...
```

**Version Update Triggers**:

- New NIST or OWASP policy published
- Pepper rotation required (1 year policy, compromise)
- Algorithm strength increase (e.g., SHA-256 → SHA-384)

**Backward Compatibility**:

- Old hashes stay on original version (v1, v2, etc.)
- New hashes use current_version
- Gradual migration (no forced re-hash)
- Rehash next time cleartext value presented

**Configuration Example**:

```yaml
hash_service:
  password_registry:
    current_version: 4  # New passwords use v4
    # Old v3, v2, v1 passwords still verified correctly
```

---

### mTLS Certificate Revocation Checking

**Q**: What revocation checking mechanisms MUST be implemented for mTLS client certificates?

**A**: BOTH CRLDP and OCSP MUST be implemented with immediate CRL publication.

**CRLDP Requirements** (CRITICAL):

- Each CRLDP HTTPS URL MUST contain ONLY ONE certificate serial number
- Serial numbers MUST be base64-url-encoded (RFC 4648) - uses `-_` instead of `+/`, no padding `=`
- CRLs MUST be signed and available IMMEDIATELY upon revocation (NOT batched)
- CRLs MUST NOT be delayed by 24 hours or any batch processing window
- HTTPS endpoints MUST be reliable and highly available
- Example: Certificate serial `0x123ABC` → base64-url encode → `EjOrvA` → `https://ca.example.com/crl/EjOrvA.crl`

**OCSP Requirements**:

- OCSP responder MUST be implemented for all issued certificates
- OCSP stapling is NICE-TO-HAVE but NOT a blocker for initial implementation

**Implementation Priority**:

1. CRLDP with immediate publication (MANDATORY Phase 2)
2. OCSP responder (MANDATORY Phase 2)
3. OCSP stapling (OPTIONAL Phase 3+)

**Rationale**: Immediate CRL publication provides fastest revocation propagation. Per-certificate CRL URLs enable fine-grained revocation without requiring clients to download massive CRLs. OCSP provides additional validation path. OCSP stapling reduces client-side lookup latency but not critical for initial release.

---

### Pepper Requirements

**Q**: How should pepper be managed and rotated?

**A**: Pepper MUST be mutually exclusive from hashed values storage, associated with hash version.

**Pepper Storage** (NEVER store pepper in DB or source code):

- **VALID OPTIONS IN ORDER OF PREFERENCE**:
  1. Docker/Kubernetes Secret (preferred for production)
  2. Configuration file (acceptable for development)
  3. Environment variable (discouraged, but supported)
- **MUST** be mutually exclusive from hashed values storage (pepper in secrets/config, hashes in DB)
- **MUST** be associated with hash version (different pepper per version)

**Pepper Rotation**:

- Pepper CANNOT be rotated silently (requires re-hash all records)
- Changing pepper REQUIRES version bump, even if no other hash parameters changed
- Example: v3 pepper compromised → bump to v4 with new pepper, re-hash all v3 records

**Additional Protections for LowEntropyDeterministicHashRegistry**:

- **MANDATORY** (prevents oracle attacks):
  - Query rate limits (prevent brute-force enumeration)
  - Abuse detection (detect suspicious query patterns)
  - Audit logs (track all hash queries for forensics)
  - Strict access control (limit who can query hashes)
- **RECOMMENDED**: Apply same protections to all 4 registries for consistency

---

### Unseal Secrets Key Management

**Q**: Should unseal secrets derive keys deterministically OR store pre-generated JWKs directly?

**A**: Support BOTH approaches with configuration selection per deployment.

**Option A: Derived Keys** (HKDF-based):

```yaml
unseal:
  mode: derive
  secret: <base64-encoded-secret>
  # Derives JWKs deterministically using HKDF
```

**Benefits**: Reproducible across environments, smaller secret size, algorithm-agile

**Option B: Pre-Generated JWKs**:

```yaml
unseal:
  mode: jwks
  jwks_file: /run/secrets/unseal_jwks.json
  # Uses exact JWKs from file
```

**Benefits**: Full control over key material, supports imported keys, easier audit

**Configuration Selection**:

- Development/Testing: Prefer derivation (reproducibility, simpler setup)
- Production: Support both (organization-specific requirements)
- Migration: Switch between modes by re-encrypting data with new unseal JWKs

**Rationale**: Different organizations have different key management policies. Supporting both approaches provides maximum flexibility without sacrificing security.

---

### Pepper Rotation Operational Procedure

**Q**: What is the REQUIRED procedure for rotating pepper without service interruption?

**A**: Lazy migration - re-hash records opportunistically as users re-authenticate.

**Lazy Migration Pattern**:

1. **Deploy new pepper version**: Bump hash version (v3 → v4), add new pepper to configuration
2. **Dual-version support**: Service accepts BOTH v3 (old pepper) and v4 (new pepper) hashes
3. **Re-hash on re-authentication**: When user authenticates with v3 hash, verify with old pepper, then re-hash with v4 pepper and update database
4. **Gradual migration**: Records naturally migrate over time as users authenticate
5. **Monitor migration**: Track percentage of records still on old version
6. **Deprecation**: After 90 days (or sufficient migration %), remove v3 support

**Configuration Example**:

```yaml
hash_service:
  password_registry:
    current_version: 4  # New hashes use v4 + new pepper
    supported_versions: [3, 4]  # Accept v3 + old pepper during grace period
    pepper_v3: <old-pepper-base64>
    pepper_v4: <new-pepper-base64>
```

**Advantages**:

- Zero downtime (no service interruption)
- No forced password resets (user convenience)
- No mass re-hashing job (avoids database load spike)
- Natural migration tied to user activity

**Rationale**: Lazy migration balances security (new pepper deployed immediately for new hashes) with operational simplicity (no downtime or mass updates). Most active users migrate within days/weeks; inactive accounts migrate over months.

---

## Observability and Telemetry

### Telemetry Architecture

**Q**: How should telemetry be forwarded and aggregated?

**A**: All telemetry MUST be forwarded through otel-contrib sidecar to upstream platforms.

**Telemetry Forwarding Architecture**:

```
cryptoutil services (OTLP gRPC:4317 or HTTP:4318)
  → OpenTelemetry Collector Contrib
  → Grafana-OTEL-LGTM (OTLP gRPC:14317 or HTTP:14318)
```

**Push-Based Telemetry Flow**:

**Application Telemetry**:

- Protocol: OTLP (OpenTelemetry Protocol) - push-based
- Supported Protocols:
  - GRPC: `grpc://host:port` (efficient binary, default for internal)
  - HTTP: `http://host:port` or `https://host:port` (firewall-friendly)
- Data: Crypto operations, API calls, business logic telemetry

**Collector Self-Monitoring**:

- Protocol: OTLP - push-based (collector exports its own telemetry)
- Data: Collector throughput, error rates, queue depths, resource usage

**Configuration Requirements**:

- `cryptoutil-otel.yml` MUST point to `opentelemetry-collector:4317`
- **NEVER** configure cryptoutil to bypass otel-collector-contrib sidecar
- The otel-contrib sidecar handles processing, filtering, routing before forwarding

**Rationale**:

- Ensures centralized telemetry processing and filtering
- Maintains consistent architecture across environments
- Enables future enhancements (sampling, aggregation, etc.)
- Prevents direct coupling between services and telemetry platforms

---

### Telemetry Sampling Strategy Configuration

**Q**: What are the exact thresholds and algorithm for adaptive sampling?

**A**: All sampling strategies in OpenTelemetry Collector config as commented options, tail-based sampling uncommented by default.

**OpenTelemetry Collector Configuration Pattern** (`configs/observability/otel-collector-config.yaml`):

```yaml
processors:
  # OPTION A: Head-based sampling (commented out)
  # probabilistic_sampler:
  #   sampling_percentage: 100  # 100% at low load
  #   # Decrease to 10% at high load via config hot-reload

  # OPTION B: Tail-based sampling (ACTIVE - uncommented)
  tail_sampling:
    decision_wait: 10s
    num_traces: 100
    expected_new_traces_per_sec: 10
    policies:
      - name: sample-errors
        type: status_code
        status_code: {status_codes: [ERROR]}
      - name: sample-slow
        type: latency
        latency: {threshold_ms: 1000}
      - name: sample-probabilistic
        type: probabilistic
        probabilistic: {sampling_percentage: 10}

  # OPTION C: Adaptive sampling (commented out)
  # adaptive_sampler:
  #   initial_sampling_percentage: 100
  #   target_spans_per_second: 1000

  # OPTION D: Always-on (commented out)
  # noop: {}
```

**Default Strategy**: Tail-based sampling (Option B)

- Sample ALL errors (status_code: ERROR)
- Sample ALL slow requests (latency >1s)
- Sample 10% of remaining traces probabilistically

**Customization**:

- Administrator can uncomment desired strategy
- Swap between strategies by commenting/uncommenting blocks
- Hot-reload config without service restart (OpenTelemetry Collector feature)

**Rationale**: Tail-based sampling provides best balance of observability (catch all errors and slow requests) and cost (reduce trace volume). Configuration-driven approach enables deployment-specific tuning without code changes.

---

### Health Check Failure Tolerance and Recovery

**Q**: What action should orchestrator take when health checks fail after all retries?

**A**: Kubernetes: Remove from load balancer + restart pod. Docker Compose: Mark unhealthy + continue running.

**Kubernetes Behavior**:

- **Liveness probe failure**: Restart pod (assumes service is deadlocked/crashed)
- **Readiness probe failure**: Remove from Service load balancer (stop routing traffic)
- Both use same health check endpoint (`/admin/api/v1/livez` and `/admin/api/v1/readyz`)
- Configuration:

  ```yaml
  livenessProbe:
    httpGet:
      path: /admin/api/v1/livez
      port: 9090
      scheme: HTTPS
    initialDelaySeconds: 10
    periodSeconds: 5
    failureThreshold: 5  # 5 retries = 25s before restart

  readinessProbe:
    httpGet:
      path: /admin/api/v1/readyz
      port: 9090
      scheme: HTTPS
    initialDelaySeconds: 5
    periodSeconds: 5
    failureThreshold: 3  # 3 retries = 15s before removing from LB
  ```

**Docker Compose Behavior**:

- **Health check failure**: Mark container as "unhealthy" in `docker ps` output
- **NO automatic restart**: Container continues running (manual intervention required)
- **NO load balancer removal**: Docker Compose doesn't provide service mesh load balancing
- Configuration:

  ```yaml
  healthcheck:
    test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
    start_period: 10s
    interval: 5s
    timeout: 2s
    retries: 5  # 5 retries = 25s before marking unhealthy
  ```

**Rationale**: Kubernetes and Docker Compose have different orchestration capabilities. Kubernetes restarts failed pods and manages load balancer registration. Docker Compose marks health status but requires manual intervention for recovery.

---

## Deployment and Docker

### Docker Compose Latency Hiding Strategies

**Q**: How can Docker Compose startup time be minimized?

**A**: Single build shared image, schema initialization by first instance, health check dependencies.

**MANDATORY Optimizations**:

1. **Single Build, Shared Image**:

```yaml
services:
  builder:
    build: ./
    image: cryptoutil:local

  cryptoutil-postgres-1:
    image: cryptoutil:local  # Reuses built image
    depends_on:
      builder:
        condition: service_completed_successfully
```

**Rationale**: Build once, all instances use same image. Prevents 3× build time.

1. **Schema Initialization by First Instance**:

```yaml
cryptoutil-postgres-1:
  depends_on:
    postgres:
      condition: service_healthy

cryptoutil-postgres-2:
  depends_on:
    cryptoutil-postgres-1:
      condition: service_healthy  # Waits for schema init
```

**Rationale**: First instance initializes DB schema, others wait. Prevents contention.

1. **Health Check Dependencies**:

```yaml
cryptoutil-sqlite:
  healthcheck:
    test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/api/v1/livez"]
    start_period: 10s
    interval: 5s
    retries: 5

otel-collector:
  depends_on:
    cryptoutil-sqlite:
      condition: service_healthy
```

**Rationale**: Services start only after dependencies healthy, not just started.

**Expected Startup Times**:

| Service | Expected Time | Strategy |
|---------|--------------|----------|
| builder | 30-60s | One-time build, cached for all instances |
| postgres | 5-30s | start_period=5s + (5s×5 retries) = max 30s |
| cryptoutil (first) | 10-35s | start_period=10s + (5s×5 retries) + unseal |
| cryptoutil (others) | 5-15s | Schema initialized, just unseal |
| otel-collector | 10-40s | Waits for cryptoutil, 10s sleep + 15 retries |

**Total Expected**: 60-150s for full stack in optimal conditions
**GitHub Actions**: Add 50-100% margin for shared CPU, network latency, cold starts

---

### Docker Secrets Permissions

**Q**: What permissions should Docker secrets have?

**A**: 440 permissions (r--r-----) MANDATORY for all secrets.

**MANDATORY: All secrets files MUST have 440 permissions**:

```bash
# Correct permissions
chmod 440 deployments/compose/*/secrets/*.secret
ls -la deployments/compose/*/secrets/
# Should show: -r--r----- for all .secret files
```

**Rationale**: Prevents unauthorized access while allowing group read (Docker daemon group).

**Dockerfile Secrets Validation** (MANDATORY pattern):

```dockerfile
# Validation stage - verify secrets exist with correct permissions
FROM alpine:3.19 AS validator
WORKDIR /validation

# Copy secrets from builder stage (if applicable)
COPY --from=builder /run/secrets/ /run/secrets/ 2>/dev/null || true

# Validate secrets existence and permissions
RUN echo "🔍 Validating Docker secrets..." && \
    ls -la /run/secrets/ || echo "⚠️ No secrets found" && \
    if [ -d /run/secrets/ ]; then \
        for secret in database_url unseal_key tls_cert tls_key; do \
            if [ -f "/run/secrets/$secret" ]; then \
                chmod 440 "/run/secrets/$secret" 2>/dev/null || true; \
            fi; \
        done; \
    fi
```

---

## CI/CD and Automation

### PostgreSQL Service Requirements

**Q**: How should PostgreSQL be configured for unit, integration, and E2E tests?

**A**: Different strategies for different test types.

**Unit/Integration Tests**: MUST use test-containers with randomized credentials

- Use test-containers library for PostgreSQL
- Generate unique database name, username, password per test suite
- Docker containers provide isolation, no port conflicts
- NEVER use environment variables for credentials

**E2E Tests**: MUST use Docker Compose with Docker secrets

- Use Docker Compose for full-stack E2E testing
- Configure PostgreSQL via Docker secrets, not environment variables
- Mount secrets to `/run/secrets/` in containers
- Application reads credentials from secret files
- Example: `database-url: file:///run/secrets/postgres_url`

**GitHub Workflows**: May use PostgreSQL service container for legacy tests

```yaml
services:
  postgres:
    image: postgres:18
    env:
      POSTGRES_DB: cryptoutil_test
      POSTGRES_USER: cryptoutil
      POSTGRES_PASSWORD: cryptoutil_test_password
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
    ports:
      - 5432:5432
```

**Rationale**: Test-containers provide isolated, randomized credentials; Docker secrets enforce secure patterns for production-like E2E tests.

---

### PostgreSQL Service Container Workflows

**Q**: Which workflows beyond ci-race/ci-mutation/ci-coverage REQUIRE PostgreSQL service container?

**A**: OPTIONAL for all workflows - prefer test-containers library over PostgreSQL service container.

**Migration Strategy**:

- **Current State**: Some workflows use PostgreSQL service container for unit/integration tests
- **Target State**: ALL workflows use test-containers library for database dependencies
- **Transition**: Gradually migrate workflows from service containers to test-containers

**Test-Containers Benefits**:

- ✅ Isolated database per test suite (no shared state)
- ✅ Randomized credentials (no hardcoded passwords)
- ✅ Automatic cleanup (containers removed after tests)
- ✅ Works locally and in CI/CD identically
- ✅ No port conflicts (random host ports)

**PostgreSQL Service Container Pattern** (DEPRECATED):

```yaml
# ❌ DEPRECATED - Use test-containers instead
services:
  postgres:
    image: postgres:18
    env:
      POSTGRES_DB: cryptoutil_test
      POSTGRES_USER: cryptoutil
      POSTGRES_PASSWORD: cryptoutil_test_password
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
    ports:
      - 5432:5432
```

**Test-Containers Pattern** (PREFERRED):

```go
// ✅ PREFERRED - Test-containers provide isolation
func TestWithPostgreSQL(t *testing.T) {
    ctx := context.Background()
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:18"),
        postgres.WithDatabase("cryptoutil_test"),
        postgres.WithUsername("cryptoutil_"+uuid.New().String()),
        postgres.WithPassword(uuid.New().String()),
    )
    require.NoError(t, err)
    defer container.Terminate(ctx)

    connStr, err := container.ConnectionString(ctx)
    require.NoError(t, err)
    // Use connStr for test
}
```

**Rationale**: Test-containers library provides better isolation, security, and developer experience than GitHub Actions service containers. Service containers acceptable during transition period but all new tests MUST use test-containers.

---

### Docker Image Pre-Pull Strategy

**Q**: When should workflows use docker-images-pull action vs on-demand pulling?

**A**: Only pre-pull for workflows that use Docker - not all workflows need Docker images.

**Pre-Pull Decision Matrix**:

| Workflow Type | Pre-Pull? | Rationale |
|---------------|-----------|-----------|
| E2E tests (Docker Compose) | ✅ YES | 10+ images, parallel pull saves 2-5 minutes |
| Load tests (Docker Compose) | ✅ YES | Same as E2E, critical for timeout avoidance |
| Unit tests (no Docker) | ❌ NO | No Docker images needed |
| Integration tests (test-containers) | ⚠️ OPTIONAL | Test-containers pulls images on-demand, pre-pull may not help |
| CI/CD validation (no Docker) | ❌ NO | Linting, formatting, no Docker needed |

**Pre-Pull Action Usage**:

```yaml
# ✅ CORRECT - E2E workflow with Docker Compose
jobs:
  e2e-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Pre-pull Docker images
        uses: ./.github/actions/docker-images-pull
        with:
          compose-files: |
            deployments/compose.integration.yml

      - name: Run E2E tests
        run: docker compose -f deployments/compose.integration.yml up --abort-on-container-exit
```

**On-Demand Pulling** (NO pre-pull):

```yaml
# ✅ CORRECT - Unit test workflow without Docker
jobs:
  unit-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5

      - name: Run unit tests
        run: go test ./...
        # No Docker images needed, no pre-pull
```

**Rationale**: Pre-pulling images in parallel reduces E2E test startup time by 50-70% (from 5-10 minutes to 2-4 minutes). However, workflows without Docker don't need pre-pull overhead. Test-containers manages its own image pulling, so pre-pull provides minimal benefit.

---

### Variable Expansion in Heredocs

**Q**: How should variables be expanded in Bash heredocs for workflow config generation?

**A**: ALWAYS use curly braces `${VAR}` syntax for explicit variable expansion.

**CRITICAL RULES**:

- ✅ ALWAYS use `${VAR}` syntax (curly braces) for explicit variable expansion
- ✅ ALWAYS verify generated config files have expanded values (not literal $VAR strings)
- ✅ ALWAYS test config generation with `cat config.yml` step in workflow
- ❌ NEVER use `$VAR` syntax in heredocs (may write literal "$VAR" to file)
- ❌ NEVER rely on implicit variable expansion behavior (shell-dependent)

**Correct Pattern**:

```yaml
- name: Generate config
  run: |
    cat > ./configs/test/config.yml <<EOF
    database-url: "postgres://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_NAME}?sslmode=disable"
    bind-public-address: "${APP_BIND_PUBLIC_ADDRESS}"
    bind-public-port: ${APP_BIND_PUBLIC_PORT}
    EOF
```

**Historical Mistakes**:

- ci-dast used `$POSTGRES_USER` instead of `${POSTGRES_USER}` in heredoc
- Heredoc wrote literal string "$POSTGRES_USER" to config.yml
- Application read literal "$POSTGRES_USER" as username, defaulted to 'root'
- PostgreSQL rejected connection: "role 'root' does not exist"
- Fix: Change all `$VAR` → `${VAR}` in heredoc

---

## Documentation and Workflow

### Session Documentation Strategy

**Q**: How should session work be documented?

**A**: NEVER create standalone session files, ALWAYS append to DETAILED.md Section 2 timeline.

**Append-Only Timeline Pattern** (Required):

```markdown
### YYYY-MM-DD: Brief Session Title
- Work completed: Summary of tasks (commit hashes)
- Key findings: Important discoveries or blockers
- Coverage/quality metrics: Before/after numbers
- Violations found: Any issues discovered
- Next steps: Outstanding work or follow-up needed
- Related commits: [abc1234] description
```

**Violations to Avoid**:

- ❌ `docs/SESSION-2025-12-14-*.md` (standalone session doc)
- ❌ `docs/session-*.md` (any dated session documentation)
- ❌ `docs/analysis-*.md` (standalone analysis documents)
- ❌ `docs/work-log-*.md` (separate work logs)

**Why This Matters**:

- Prevents documentation bloat (dozens of orphaned session files)
- Single source of truth for implementation timeline
- Easier to search and review work history
- Maintains chronological narrative flow

**When to Create New Documentation**:

**ONLY create new docs for**:

- Permanent feature specifications (`specs/*/README.md`, `TASKS.md`)
- Reference guides users need (`docs/DEMO-GUIDE.md`, `docs/DEV-SETUP.md`)
- Post-mortem analysis requiring deep dive (`docs/P0.X-*.md`)
- Architecture Decision Records (ADRs)

---

### Git Workflow Patterns

**Q**: Should commits be incremental or amended during implementation?

**A**: ALWAYS commit incrementally (NOT amend) to preserve history and enable bisect.

**Why Incremental Commits Matter**:

- Preserves full timeline of changes and decisions
- Enables git bisect to identify when bugs were introduced
- Allows selective revert of specific fixes
- Shows thought process and iterative improvement
- Easier to review each logical change independently

**NEVER use `git commit --amend` repeatedly**:

```bash
# ❌ WRONG: Amend repeatedly (loses history)
git commit -m "fix"
git add more_fixes
git commit --amend
git add even_more_fixes
git commit --amend  # Original fix context lost!
```

**ALWAYS commit incrementally**:

```bash
# ✅ CORRECT: Commit each logical unit independently
git commit -m "fix(format_go): restore clean baseline from 07192eac"
# Run tests, verify baseline works
git commit -m "fix(format_go): add defensive check with filepath.Abs()"
# Run tests, verify defensive check works
git commit -m "test(format_go): verify self_modification_test catches regressions"
# Clear progression, easy to bisect, reviewable history
```

**When to Use Amend** (Rare Cases):

- Fixing typos in commit message IMMEDIATELY after commit (before push)
- Adding forgotten files to most recent commit (within 1 minute)
- NEVER amend after push (breaks shared history)
- NEVER amend repeatedly during debugging session

---

### Restore from Clean Baseline Pattern

**Q**: When fixing regressions, should fixes be applied to current HEAD?

**A**: ALWAYS restore from clean baseline FIRST, then apply targeted fixes.

**Why This Matters**:

- HEAD may be corrupted by previous failed attempts
- Incremental fixes on corrupted base compound the problem
- Clean baseline ensures you start from known-good state
- Prevents "fixing" code that's already broken

**ALWAYS restore clean baseline FIRST**:

```bash
# ✅ CORRECT: Restore clean baseline, THEN apply targeted fixes
# 1. Find last known-good commit
git log --oneline --grep="baseline" | head -5

# 2. Restore ENTIRE package from clean commit
git checkout <clean-commit-hash> -- path/to/package/

# 3. Verify baseline works
go test ./path/to/package/
git status  # Should show only restored files

# 4. Apply ONLY the new fix (minimal change)
# Edit specific file with targeted change

# 5. Verify fix works independently
go test ./path/to/package/

# 6. Commit as NEW commit (not amend!)
git commit -m "fix(package): add defensive check for X"
```

**Common Mistakes**:

- Assuming HEAD is correct (may be corrupted from previous attempts)
- Applying "one more fix" on top of corrupted code
- Mixing baseline restoration with new fixes in same commit
- Using amend instead of new commits (loses restoration evidence)

---

## Authentication and Authorization

**Source**: QUIZME-02 (December 22, 2025) and AUTH-AUTHZ-SINGLE-FACTORS.md

### Single Factor Authentication Methods

**Q**: What single factor authentication (SFA) methods are supported for headless-based and browser-based clients?

**A** (Source: QUIZME-02 Q1-Q2, AUTH-AUTHZ-SINGLE-FACTORS.md):

**Headless-Based Clients** (`/service/*` paths): 10 methods

- Non-Federated (3): Basic (Client ID/Secret), Bearer (API Token), HTTPS Client Certificate
- Federated (7): Basic (Client ID/Secret), Bearer (API Token), HTTPS Client Certificate, JWE OAuth 2.1 Access Token, JWS OAuth 2.1 Access Token, Opaque OAuth 2.1 Access Token, Opaque OAuth 2.1 Refresh Token

**Browser-Based Clients** (`/browser/*` paths): 28 methods

- Non-Federated (6): JWE Session Cookie, JWS Session Cookie, Opaque Session Cookie, Basic (Username/Password), Bearer (API Token), HTTPS Client Certificate
- Federated (22): All non-federated (6) + TOTP + HOTP + Recovery Codes + WebAuthn with Passkeys + WebAuthn without Passkeys + Push Notification + Basic (Email/Password) + Magic Link via Email + Magic Link via SMS + Random OTP via Email + Random OTP via SMS + Random OTP via Phone

**Multi-Factor Authentication (MFA)**:

- MFA = Combination of 2+ single factor authentication methods
- Factor priority order: Passkey > TOTP > Hardware Keys > Email OTP > SMS OTP > HOTP > Recovery Codes > Push Notifications > Phone Call OTP

**Storage Realms**: Config (YAML) > SQL (GORM) for disaster recovery priority

**Note**: Rate limiting and IP allowlist removed from AUTH-AUTHZ-SINGLE-FACTORS - see KMS reference implementation for pattern.

---

### Authorization Methods

**Q**: What authorization methods are supported?

**A** (Source: AUTH-AUTHZ-SINGLE-FACTORS.md):

**Headless-Based Clients** (2 methods):

- Scope-Based Authorization
- Role-Based Access Control (RBAC)

**Browser-Based Clients** (4 methods):

- Scope-Based Authorization
- Role-Based Access Control (RBAC)
- Resource-Level Access Control
- Consent Tracking (scope+resource tuples)

---

### Session Token Format Configuration

**Q**: How is session token format determined - is it fixed per product or configurable?

**A** (Source: QUIZME-02 Q3):

Session token format is **configuration-driven**:

**Non-Federated Mode**: Product-specific configuration determines format

```yaml
session:
  token_format: opaque  # or jwe, jws
```

**Federated Mode**: Identity Provider configuration determines format

```yaml
federation:
  identity:
    session_token_format: jwe  # or jws, opaque
```

**Admin configures** via YAML or environment variables per deployment.

---

### Session Storage Backend

**Q**: What database backends are supported for session storage? Should Redis be used?

**A** (Source: QUIZME-02 Q4):

**Supported Backends**:

- **SQLite**: Single-node deployments
- **PostgreSQL**: Distributed/high-availability deployments with shared session data
- **NO Redis**: NOT supported (adds operational complexity)

**Configuration Examples**:

```yaml
# Single-node
database:
  driver: sqlite
  dsn: "file:sessions.db?cache=shared"

# Distributed/HA
database:
  driver: postgres
  dsn: "postgres://user:pass@host:5432/sessions?sslmode=require"
```

---

### Session Cookie Security Attributes

**Q**: What HttpOnly, Secure, SameSite settings should be used for session cookies?

**A** (Source: QUIZME-02 Q5):

**Deferred to KMS reference implementation** - See `application_listener.go` for Swagger UI cookie settings pattern.

**Rationale**: Cookie security attributes depend on deployment context (HTTPS availability, cross-origin requirements, browser compatibility). KMS implementation provides validated production patterns.

---

### MFA Step-Up Authentication Triggers

**Q**: What triggers re-authentication (step-up) in an active session?

**A** (Source: QUIZME-02 Q6):

**Time-Based Re-Authentication** (30-minute intervals):

- Re-authentication MANDATORY every 30 minutes for sensitive resources
- Applies to operations: key rotation, client secret rotation, admin actions
- Session remains valid for low-sensitivity operations

**Rationale**: Time-based provides consistent security posture regardless of operation type, preventing session hijacking exploitation window.

---

### MFA Enrollment Workflow for New Users

**Q**: Is MFA enrollment mandatory during initial user setup?

**A** (Source: QUIZME-02 Q7):

**Optional Enrollment with Limited Access**:

- Enrollment **OPTIONAL** during initial setup
- Access **LIMITED** until additional factors enrolled (read-only access)
- User MUST enroll at least one factor for write operations
- Only one identifying factor required for initial login

**Rationale**: Balances security with user onboarding friction. Users can explore product with read-only access before committing to full MFA enrollment.

---

### MFA Factor Fallback Strategy

**Q**: What happens when user's primary MFA factor is unavailable (e.g., lost phone for TOTP)?

**A** (Source: QUIZME-02 Q8):

**Any Identifying Factor Sufficient**:

- Any factor that uniquely identifies the user is sufficient for first MFA factor
- User can select any enrolled factor from login UI
- No automatic fallback hierarchy (user chooses explicitly)

**Rationale**: Provides flexibility while maintaining security. User controls which enrolled factor to use per authentication attempt.

---

### OAuth 2.1 Access Token vs Session Token Distinction

**Q**: What is the relationship between OAuth 2.1 Access Tokens and session cookies in Federated Identity mode?

**A** (Source: QUIZME-02 Q9):

**SEPARATE TOKENS** (Exactly B!!!):

- OAuth 2.1 Access Token exchanged for internal session cookie
- Backend-for-Frontend (BFF) pattern
- OAuth token used for IdP federation, session cookie used for application state
- NO token nesting or reuse

**Rationale**: Decouples external OAuth flow from internal session management. Session cookie can have different format/lifetime than OAuth token.

---

### Realm Type Failover Behavior

**Q**: How should authentication realm failover work when database is unavailable?

**A** (Source: QUIZME-02 Q10):

**Admin-Configured Priority List**:

```yaml
realms:
  priority_list:
    - type: file
      realm: config.yaml
    - type: database
      realm: postgresql_production
    - type: database
      realm: sqlite_fallback
```

**Behavior**:

- System tries each Realm+Type in priority order
- Continue until one succeeds or all fail
- Supports flexible failover and multi-realm setups

**Rationale**: Configuration-driven failover allows operators to define deployment-specific priorities without code changes.

---

### Authorization Decision Caching Strategy

**Q**: Should authorization decisions be cached or evaluated on every request?

**A** (Source: QUIZME-02 Q11):

**Zero Trust - No Caching**:

- Authorization decisions MUST be evaluated on EVERY request
- NO caching of authorization decisions (prevents stale permissions)
- Performance via efficient policy evaluation, not caching

**Rationale**: Eliminates risk of stale permissions after role/policy changes. Real-time policy evaluation ensures security, even if slightly slower.

---

### Cross-Service Authorization Propagation

**Q**: How should authorization be handled when Identity passes requests to KMS/JOSE/CA?

**A** (Source: QUIZME-02 Q12):

**Direct Token Validation**:

- Session token passed between federated services via HTTP headers
- Each service independently validates token and enforces authorization
- NO token transformation or delegation

**Rationale**: Each service maintains autonomy for authorization decisions. Prevents cascading authorization failures and simplifies debugging.

---

### Rate Limiting Configuration

**Q**: What are the rate limiting requirements for authentication endpoints?

**A** (Source: QUIZME-02 Q13):

**Removed from AUTH-AUTHZ-SINGLE-FACTORS** - See KMS reference implementation.

Refer to `application_listener.go` and `config.go` for production-validated rate limiting patterns.

---

### IP Allowlist Configuration

**Q**: What are the IP allowlist requirements for admin endpoints?

**A** (Source: QUIZME-02 Q14):

**Removed from AUTH-AUTHZ-SINGLE-FACTORS** - See KMS reference implementation.

Refer to `application_listener.go` and `config.go` for production-validated IP allowlist patterns.

---

### Consent Tracking Granularity

**Q**: At what granularity should consent be tracked (per-scope or per-resource)?

**A** (Source: QUIZME-02 Q15):

**Scope+Resource Tuples**:

- Tracked as `(scope, resource)` tuples
- Example: `("read:keys", "key-123")` separate from `("read:keys", "key-456")`
- Enables fine-grained consent revocation per resource

**Rationale**: Provides maximum user control over data access. User can revoke access to specific resources without affecting entire scope.

---

## Service Template and Migration Strategy

### Service Template Migration Priority

**Q**: Should identity services (authz, idp, rs) be refactored to use the extracted service template immediately after cipher-im validation, or later?

**A** (Source: CLARIFY-QUIZME-01 Q1, 2025-12-22):

Identity services will be migrated **LAST** in the following sequence:

1. **cipher-im** (Phase 3): Validate service template first
2. **JOSE and CA** (Phases 4-5): Migrate next, one at a time, to allow adjustments to the service template to accommodate JOSE and CA service patterns
3. **Identity services** (Phase 6+): Migrate last, ordered by Authz → IdP → RS → RP → SPA

**Rationale**: cipher-im will validate the service template first, then JOSE and CA migrations will drive template refinements to support different service patterns. Identity services migrate last to benefit from a mature, battle-tested template.

---

### Cipher-IM Service Specification

**Service Name**: cipher-im (short form), Cipher-InstantMessenger (full descriptive name)

**Q**: What are the detailed requirements for the Cipher-InstantMessenger demonstration service?

**A**: Encrypted messaging service that validates service template reusability and demonstrates crypto library integration.

**Service Overview**:

- **Purpose**: Secure messaging between users with end-to-end encryption
- **Deployment**: Single-tenant (no multi-tenancy support)
- **Authentication**: BASIC username/password ONLY
- **Authorization**: Sender+receivers mapping controls message access

**Message Encryption Architecture**:

1. Sender calls `PUT /tx` with plaintext message + receiver list
2. System generates random AES-256-GCM JWK for this message
3. Message encrypted with AES-256-GCM JWK
4. For each receiver:
   - Fetch receiver's public ECDH key
   - Encrypt AES-256-GCM JWK using ECDH-AESGCMKW with receiver's public key
   - Store encrypted JWK + encrypted message + receiver ID
5. Return message ID (UUIDv7)

**API Specification**:

- `PUT /tx` - Send message
  - Request: `{"receivers": ["user1", "user2"], "message": "plaintext"}`
  - Response: `{"id": "UUIDv7", "timestamp": "ISO8601"}`
  - Auth: Sender identity from BASIC auth

- `GET /tx?m={UUIDv7}` - List sent messages (sender's outbox)
  - Optional filter: `m=UUIDv7` for specific message
  - Sorting: Timestamp DESC (hardcoded, no pagination)
  - Auth: Only sender can list their sent messages

- `DELETE /tx/{id}` - Delete sent message
  - Effect: Removes sender's copy only
  - Receivers still have their copies (independent deletion)
  - Auth: Only sender can delete their sent message

- `GET /rx?m={UUIDv7}` - List received messages (receiver's inbox)
  - Decrypt: Use receiver's private ECDH key to unwrap AES-256-GCM JWK, then decrypt message
  - Optional filter: `m=UUIDv7` for specific message
  - Sorting: Timestamp DESC (hardcoded, no pagination)
  - Auth: Only receiver can list their received messages

- `DELETE /rx/{id}` - Delete received message
  - Effect: Removes this receiver's copy only
  - Sender and other receivers unaffected (independent deletion)
  - Auth: Only receiver can delete their received message

**Database Schema**:

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,  -- UUIDv7
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,  -- PBKDF2-HMAC-SHA256
  public_key_jwk JSONB NOT NULL,  -- ECDH public key
  private_key_jwk_encrypted JSONB NOT NULL,  -- Encrypted with password-derived key
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE messages (
  id UUID PRIMARY KEY,  -- UUIDv7
  sender_id UUID NOT NULL REFERENCES users(id),
  message_encrypted BYTEA NOT NULL,  -- AES-256-GCM ciphertext
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE message_receivers (
  id UUID PRIMARY KEY,  -- UUIDv7
  message_id UUID NOT NULL REFERENCES messages(id),
  receiver_id UUID NOT NULL REFERENCES users(id),
  jwk_encrypted JSONB NOT NULL,  -- AES-256-GCM JWK encrypted for this receiver
  deleted_by_receiver BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE(message_id, receiver_id)
);

CREATE INDEX idx_messages_sender ON messages(sender_id, created_at DESC);
CREATE INDEX idx_message_receivers_receiver ON message_receivers(receiver_id, created_at DESC);
```

**Crypto Libraries Used**:

- JWE encryption/decryption (shared `internal/shared/crypto/jwe`)
- ECDH key agreement (shared `internal/shared/crypto/ecdh`)
- AES-GCM encryption (shared `internal/shared/crypto/aes`)
- PBKDF2 password hashing (shared `internal/shared/crypto/hashes`)

**Service Template Validation**:

- Dual HTTPS servers (public 8070 + admin 9090)
- PostgreSQL + SQLite support
- OTLP telemetry integration
- Health checks (`/admin/api/v1/livez`, `/admin/api/v1/readyz`)
- Graceful shutdown (`/admin/api/v1/shutdown`)
- Docker Compose deployment
- Configuration management (YAML + env vars + CLI params)

**Success Criteria**:

- Service implementation <1000 lines (template handles infrastructure)
- All unit tests pass (coverage ≥95%)
- E2E tests validate encryption/decryption flow

---

### Shared Package Organization - CRITICAL

**Q**: Why must reusable code be in `internal/shared/` packages, and what are the consequences of placing it in service-specific locations?

**A** (Source: Course correction 2025-12-25):

**MANDATORY**: All reusable code MUST be in `internal/shared/` packages.

**Current Problems**:

1. **internal/jose/crypto**: Contains JWE/JWS/JWK utilities needed by MULTIPLE services (sm-kms, jose-ja, cipher-im), but is currently in JOSE service-specific location
2. **Service Template TLS**: Duplicating TLS cert generation code instead of using existing `internal/shared/crypto/certificate/` infrastructure

**Consequences of Current Organization**:

- **Blocks cipher-im implementation**: cipher-im needs JWE encryption from `internal/jose/crypto`, but importing service-specific packages creates circular dependencies
- **Technical debt**: Service template duplicates TLS cert generation code instead of reusing shared infrastructure
- **Hard-coding**: Service template has hard-coded values instead of parameter injection patterns
- **Migration delays**: Every service migration will encounter same issues, wasting time on rework

**Course Corrections**:

**Phase 1.1: Move JOSE Crypto** (NEW, BLOCKING):

- Move `internal/jose/crypto/*` → `internal/shared/crypto/jose/`
- Update all imports in sm-kms, jose-ja, service template
- Verify tests pass with no coverage regression
- **Rationale**: Enables cipher-im to use JWE without circular dependencies

**Phase 1.2: Refactor Template TLS** (NEW, BLOCKING):

- Remove duplicated TLS code from service template
- Use `internal/shared/crypto/certificate/` and `internal/shared/crypto/keygen/`
- Implement parameter injection for all TLS configuration
- Support all 3 TLS modes: static certs, mixed (static+generated), auto-generated
- **Rationale**: Prevents technical debt in all service migrations

**Required Shared Packages** (from `spec.md`):

| Package | Purpose | Used By | Status |
|---------|---------|---------|--------|
| `internal/shared/crypto/jose/` | JWK/JWE/JWS utilities | sm-kms, jose-ja, cipher-im | **MUST MOVE** from internal/jose/crypto |
| `internal/shared/crypto/certificate/` | TLS cert chains | All services | ✅ Exists, MUST BE USED |
| `internal/shared/telemetry/` | OTLP integration | All services | ✅ Exists |
| `internal/shared/magic/` | Magic constants | All packages | ✅ Exists |
| `internal/shared/crypto/digests/` | Hash algorithms | All services | ✅ Exists |
| `internal/shared/crypto/hash/` | Password hashing, hash registry | sm-kms, identity | ✅ Exists |
| `internal/shared/util/` | Utilities | All packages | ✅ Exists |

**Quality Requirements**:

- All shared packages MUST have ≥98% coverage (infrastructure/utility code standard)
- All shared packages MUST have ≥98% mutation score (infrastructure/utility code standard)
- All shared packages MUST have comprehensive documentation with usage examples

**Migration Dependencies**:

- Phase 1.1 (Move JOSE Crypto) is **BLOCKING** Phase 2 (Template Extraction)
- Phase 1.2 (Refactor Template TLS) is **BLOCKING** Phase 3 (Cipher-IM Implementation)
- All production service migrations (Phases 4-7) depend on clean shared package organization
- Docker Compose deployment works (SQLite + PostgreSQL modes)
- Demonstrates crypto library integration without external dependencies

---

### Monitoring and Metrics Architecture

**Q**: Should admin ports expose `/admin/api/v1/metrics` endpoint for external monitoring tools (Prometheus, Grafana)?

**A** (Source: CLARIFY-QUIZME-01 Q2, 2025-12-22):

**CRITICAL**: `/admin/api/v1/metrics` endpoint is a **MISTAKE** and MUST be removed from the project entirely.

**Correct Architecture**:

- ALL services MUST use OTLP protocol to **push** metrics, tracing, and logging to OpenTelemetry Collector Contrib
- **NEVER** use pull or scrape patterns (no Prometheus scraping of service endpoints)
- OpenTelemetry Collector Contrib uses OTLP to forward metrics, tracing, and logging to Grafana LGTM

**Action Required**:

- Remove all references to `/admin/api/v1/metrics` from codebase
- Remove Prometheus scraping configurations
- Update documentation to clarify push-only telemetry architecture

---

### SQLite Production Readiness

**Q**: Should SQLite be supported for production single-instance deployments, or remain strictly development-only?

**A** (Source: CLARIFY-QUIZME-01 Q3, 2025-12-22):

SQLite is **acceptable** for production single-instance deployments with **<1000 requests/day**.

**Requirements**:

- MUST NOT forbid SQLite in constitution.md, spec.md, or copilot instructions for low-traffic production deployments
- Recommended: Use PostgreSQL for production deployments
- Acceptable: Use SQLite for small-scale production deployments with <1000 requests/day

**Rationale**: Small-scale deployments benefit from SQLite's simplicity (no separate database server, zero-configuration). Traffic threshold ensures SQLite's single-writer limitation isn't violated.

---

### MFA Factor Implementation Priority

**Q**: What is the mandatory implementation sequence for MFA factors? Should deprecated factors (SMS OTP) be implemented?

**A** (Source: CLARIFY-QUIZME-01 Q4, 2025-12-22):

**All factors including deprecated ones MUST be implemented for backward compatibility**.

**MFA Factors** (in priority order, ALL MANDATORY):

1. **Passkey** (WebAuthn with discoverable credentials) - HIGHEST priority, FIDO2 standard
2. **TOTP** (Time-based One-Time Password) - HIGH priority, RFC 6238, authenticator apps
3. **Hardware Security Keys** (WebAuthn without passkeys) - HIGH priority, FIDO U2F/FIDO2
4. **Email OTP** (One-Time Password via email) - MEDIUM priority, email delivery required
5. **Recovery Codes** (Pre-generated backup codes) - MEDIUM priority, account recovery
6. **SMS OTP** (NIST deprecated but MANDATORY) - MEDIUM priority, backward compatibility
7. **Phone Call OTP** (NIST deprecated but MANDATORY) - LOW priority, backward compatibility
8. **Magic Link** (Time-limited authentication link via email/SMS) - LOW priority
9. **Push Notification** (Mobile app push-based approval) - LOW priority

**Action Required**: Constitution.md MUST be updated to reflect this requirement, including listing all authentication methods in priority order.

**Rationale**: Even though SMS OTP and Phone Call OTP are NIST deprecated, many organizations still rely on them for legacy compatibility and user accessibility.

---

### Certificate Profile Extensibility

**Q**: Should the CA support custom certificate profiles beyond the 24 predefined profiles?

**A** (Source: CLARIFY-QUIZME-01 Q5, 2025-12-22):

**Support custom profiles via YAML configuration files** (file-based extensibility).

**Implementation**:

- 24 predefined profiles cover most use cases
- Organizations with specific needs can define custom profiles in YAML configuration
- Profiles loaded at runtime from configuration directory
- No database-driven or plugin-based extensibility needed at this time

**Rationale**: File-based configuration strikes balance between flexibility and simplicity. Most organizations won't need custom profiles; those that do can manage YAML files via version control.

---

### Telemetry Data Retention and Privacy

**Q**: What data retention policy should be enforced for telemetry data? Should sensitive fields be redacted?

**A** (Source: CLARIFY-QUIZME-01 Q6, 2025-12-22):

**Retain telemetry data for 90 days with NO redaction of any fields by default**.

**Configuration**:

- Default retention: 90 days
- Redaction: None by default (full observability)
- Operators MAY configure custom redaction patterns per deployment if needed for compliance

**Rationale**: Full observability is preferred for troubleshooting and forensics. Compliance requirements (GDPR, CCPA) vary by deployment; operators can enable redaction via configuration when needed.

---

### Federation Fallback Mode for Production

**Q**: What is the MANDATORY fallback mode for production deployments when the Identity service is unavailable?

**A** (Source: CLARIFY-QUIZME-01 Q7, 2025-12-22):

**reject_all** (strict mode) is **MANDATORY** for production deployments.

**Fallback Behavior**:

- **Production**: `reject_all` - Deny all requests until Identity service recovers (maximum security)
- **Development**: `allow_all` - Allow all requests during development (convenience)
- **Local validation**: NOT allowed in production (risk of stale cached keys)

**Rationale**: Security over availability. If the Identity service is down, it's better to reject traffic than risk unauthorized access with stale cached credentials.

---

### Docker Secrets vs Kubernetes Secrets Priority

**Q**: Should the codebase prioritize Docker secrets or Kubernetes secrets integration?

**A** (Source: CLARIFY-QUIZME-01 Q8, 2025-12-22):

**Docker secrets ONLY** - Kubernetes deployments must use Docker-compatible secret mounting.

**Implementation Pattern**:

- All services read secrets from `file:///run/secrets/*` paths
- Kubernetes deployments mount secrets as files using same paths
- No special Kubernetes secret handling (env vars, volume mounts with different paths)

**Rationale**: Single secret handling implementation reduces complexity. Kubernetes supports Docker-compatible secret mounting via volumeMounts, so no separate code path needed.

---

### Load Testing Target Performance Metrics

**Q**: What are the target performance metrics for load testing across all API types?

**A** (Source: CLARIFY-QUIZME-01 Q9, 2025-12-22):

**No hard targets** - Load tests validate scalability trends and identify bottlenecks only.

**Approach**:

- Establish baseline performance metrics through initial load testing
- Iteratively improve performance over time
- Track trends (requests/second, latency percentiles, error rates)
- No specific numeric targets (e.g., "1000 req/s") at this time

**Rationale**: Performance requirements vary by deployment scale and hardware. Focus on identifying bottlenecks and improving trends rather than arbitrary numeric targets.

---

### E2E Test Workflow Coverage Priority

**Q**: What is the minimum viable E2E test coverage for Phase 2 completion?

**A** (Source: CLARIFY-QUIZME-01 Q10, 2025-12-22):

**JOSE + CA + KMS** (Identity later).

**E2E Coverage Sequence**:

1. **Phase 2**: JOSE signing/verification, CA certificate issuance, KMS encryption/decryption
2. **Phase 3+**: OAuth 2.1 authorization code flow, OIDC authentication flow, token validation

**Rationale**: JOSE, CA, and KMS are standalone products with clear E2E scenarios. Identity product has complex multi-service interactions that benefit from later implementation after other products stabilize.

---

### Mutation Testing Enforcement Strategy

**Q**: Should mutation testing targets be enforced strictly per package, or allow exemptions?

**A** (Source: CLARIFY-QUIZME-01 Q11, 2025-12-22):

**Allow exemptions for generated code** (e.g., OpenAPI-generated models) with ramp-up plan.

**Enforcement Strategy**:

- Generated code (OpenAPI models, protobuf) may start below 85% mutation coverage
- MUST be ramped up to ≥85% over time through additional tests
- Document exemptions in clarify.md with justification and timeline
- Business logic and infrastructure packages: Strict ≥85%/≥98% enforcement

**Rationale**: Generated code often has boilerplate that's hard to mutate meaningfully. Allow initial exemption but require improvement over time.

---

### Probabilistic Testing Seed Management

**Q**: Should probabilistic test execution use fixed seeds or random seeds?

**A** (Source: CLARIFY-QUIZME-01 Q12, 2025-12-22):

**Always random seed** - Probabilistic test execution is a performance-only optimization, not a reproducibility feature.

**Implementation**:

- Use random seed per test run (time-based or Go's default random seed)
- Do NOT use fixed seeds (SEED=12345)
- Do NOT use date-based seeds (YYYYMMDD)

**Rationale**: Probabilistic testing is purely for reducing test execution time (<15s per package target). It's not intended for reproducibility. Tests that need reproducibility should NOT use probabilistic execution.

---

## Federation and Service Integration (Session 2025-12-23)

### Circuit Breaker Retry Behavior

**Q**: When circuit breaker opens after 5 failures, does retry mechanism continue running?

**A**: Stop retrying immediately - fail-fast until half-open state after 60s timeout

- Circuit breaker states: Closed (normal), Open (fail-fast), Half-Open (testing)
- Retry mechanisms ONLY active in Closed and Half-Open states
- Open state: All requests fail immediately without retry attempts
- After timeout (60s), transition to Half-Open for testing
- Prevents resource exhaustion and cascading failures

---

### Admin Port Collision in Unified Deployments

**Q**: How do multiple services avoid admin port collisions in unified deployments?

**A**: Containerization requirement - each container has isolated localhost namespace, non-containerized unified deployments not supported

- Admin ports fixed at 127.0.0.1:9090 for all services
- Containerization REQUIRED for multi-service deployments (each container isolates localhost)
- Non-containerized unified deployments NOT SUPPORTED (would cause port collisions)
- Single-service standalone deployments can run non-containerized
- Rationale: Container isolation enables consistent admin port across all services

---

### Session Token Format Selection

**Q**: What determines session token format selection between opaque, JWE, and JWS?

**A**: Configuration-driven per service deployment - admin configures via YAML, default opaque for browser/JWS for headless, all three formats must be supported

- Format selection: Administrator-configured via YAML deployment configuration
- Default behavior: Opaque tokens for browser-based clients, JWS tokens for headless clients
- Mandatory support: All services MUST implement all three formats (opaque, JWE, JWS)
- Rationale: Enables deployment flexibility, security/performance tradeoffs per environment
- See: Session Token Format section and [.github/instructions/02-10.authentication.instructions.md](../../.github/instructions/02-10.authentication.instructions.md) for configuration examples

---

### Federation Mode Transition Session Handling

**Q**: When service transitions from non-federated to federated mode, how are existing sessions handled?

**A**: Grace period dual-format support - accept BOTH formats during transition (e.g., 24h), old tokens expire naturally, new tokens issued for new logins

- Grace period: Accept both old-format (non-federated) and new-format (federated) tokens during transition
- Default grace period: 24 hours (configurable)
- Old token handling: Expire naturally according to their TTL (no forced invalidation)
- New token issuance: New logins immediately receive federated-format tokens
- Prevents: Service disruption and forced user re-authentication
- See: Federation Configuration section for migration configuration examples

---

### Introspection Result Caching

**Q**: How should introspection results be cached?

**A**: Cache positive results with configurable TTL, cache negative results for 1 minute - provides operational flexibility while maintaining security

- Positive results (active=true): Configurable TTL per deployment (default: match token expiry)
- Negative results (active=false): Fixed 1-minute cache to prevent abuse and reduce load
- Rationale: Positive caching reduces load on authorization server, negative caching prevents attackers from overwhelming introspection endpoint
- Configuration example:

  ```yaml
  introspection:
    cache_positive_ttl: 3600  # 1 hour (or match token TTL)
    cache_negative_ttl: 60    # 1 minute (fixed)
  ```

- See: Token Validation section for implementation details

---

### Federation Timeout Configuration Granularity

**Q**: Should federation timeouts be configurable per federated service or use single global timeout?

**A**: Per-service timeout configuration REQUIRED (identity_timeout, jose_timeout, ca_timeout separate).

**Configuration Pattern**:

```yaml
federation:
  identity:
    url: "https://identity-authz:8180"
    timeout: 10s  # Identity-specific timeout
    retry:
      max_attempts: 3
      backoff: exponential

  jose:
    url: "https://jose:8280"
    timeout: 15s  # JOSE-specific timeout (longer for crypto ops)
    retry:
      max_attempts: 3
      backoff: exponential

  ca:
    url: "https://ca:8380"
    timeout: 30s  # CA-specific timeout (longer for cert issuance)
    retry:
      max_attempts: 2
      backoff: linear
```

**Rationale for Per-Service Timeouts**:

- **Identity**: Fast token validation (<10s typical)
- **JOSE**: Moderate crypto operations (10-15s for signing/verification)
- **CA**: Slow certificate issuance (20-30s for full CA chain validation)
- Different services have different performance characteristics
- Global timeout forces lowest-common-denominator (too short for CA, too long for Identity)

**Rejected Pattern**:

- ❌ Single global `federation_timeout: 10s` - Doesn't accommodate service-specific needs

---

### Cross-Service API Versioning Strategy

**Q**: How should API versioning be handled across federated services deployed independently?

**A**: Backward compatible - support N-1 version (current and previous version).

**Compatibility Requirements**:

- Services MUST support TWO API versions simultaneously (current + previous)
- Example: Identity v2 MUST work with JOSE v1 OR JOSE v2
- Forward compatibility NOT required (Identity v1 does NOT need to work with JOSE v2)

**API Version Negotiation**:

```yaml
# Service advertises supported versions via /admin/api/v1/version endpoint
GET https://jose:8280/admin/api/v1/version
Response:
{
  "current_version": "v2",
  "supported_versions": ["v1", "v2"],
  "deprecated_versions": ["v1"]  # v1 supported but deprecated
}
```

**Client Version Selection**:

```go
// Client prefers latest supported version
if serverSupports("v2") {
    useAPIVersion("v2")
} else if serverSupports("v1") {
    useAPIVersion("v1")
} else {
    return ErrVersionMismatch
}
```

**Deprecation Timeline**:

- New version released: Support v(N) and v(N-1) for 90 days
- After 90 days: Drop v(N-1) support, add v(N+1) support
- Requires coordinated rollout but allows independent deployments within 90-day window

**Rationale**: N-1 compatibility enables gradual rollout of new API versions without requiring synchronized deployments. 90-day overlap provides sufficient time for all services to upgrade.

---

### Service Discovery DNS Caching Behavior

**Q**: How should services handle DNS caching for federated service URLs?

**A**: No caching - perform DNS lookup on every request.

**DNS Resolution Pattern**:

- Resolve federated service hostname on EVERY request
- Do NOT cache DNS lookup results
- Do NOT respect DNS TTL (even if low)
- Use Go's default DNS resolver (respects /etc/resolv.conf)

**Configuration Example**:

```yaml
federation:
  identity:
    url: "https://identity-authz:8180"  # DNS lookup on every request
  jose:
    url: "https://jose.cryptoutil.svc.cluster.local:8280"  # DNS lookup on every request
```

**Trade-offs**:

- ✅ Highest reliability (immediate failover to new IP)
- ✅ No stale DNS cache issues
- ✅ Kubernetes service mesh updates reflected immediately
- ⚠️ Slightly higher latency (DNS lookup overhead ~1-5ms)

**Rejected Patterns**:

- ❌ Cache DNS for 5 minutes - Delays failover, stale IPs during rolling updates
- ❌ Respect DNS TTL with 30s minimum - Still delays Kubernetes service updates
- ❌ Service mesh (Istio/Linkerd) - Adds operational complexity, out of scope for Phase 2

**Rationale**: In Kubernetes and Docker Compose environments, service IPs change frequently during rolling updates and scaling events. No-cache DNS ensures immediate failover without stale connection attempts.

---

## Status Summary

**Last Review**: December 24, 2025
**Next Actions**:

1. Update constitution.md with architectural decisions (session state, multi-tenancy, database sharding, mTLS revocation, unseal secrets)
2. Update spec.md with finalized requirements (connection pools, read replicas, API versioning, DNS caching, sampling strategy)
3. Update copilot instructions to reflect new clarifications (probabilistic testing with race detector, E2E path coverage)
4. Generate plan.md and tasks.md based on all clarifications
5. Begin implementation with evidence-based validation

**Key Insights**:

- Dual-server architecture is critical for all services
- Coverage targets are strict (95%/98%) with no exceptions (except generated code with ramp-up plan)
- Probabilistic testing uses random seeds for broader coverage over time
- Federation uses reject_all in production (security over availability)
- Docker Compose optimizations can reduce startup time by 50%+
- Session documentation belongs in DETAILED.md, not standalone files
- MFA: All factors (including NIST deprecated) are MANDATORY for backward compatibility
- Telemetry: Push-only via OTLP (NEVER scrape/pull patterns)
- SQLite acceptable for production <1000 req/day deployments
- Circuit breaker: Fail-fast in open state, retry only in closed/half-open states
- Session token format: Configuration-driven with all three formats supported
- Introspection caching: Positive results configurable TTL, negative results 1-minute fixed

**New Clarifications (Session 2025-12-24)**:

- Session state: SQL-backed JWS/OPAQUE/JWE (implementation priority JWS>OPAQUE>JWE, deployment priority JWE>OPAQUE>JWS)
- Database sharding: Phase 4 with multi-tenancy, partition by tenant ID
- Multi-tenancy: Schema-level isolation only
- mTLS revocation: CRLDP+OCSP both required, CRLDP immediate (not batched), OCSP stapling nice-to-have
- Unseal secrets: Support BOTH derivation and pre-generated JWKs
- Pepper rotation: Lazy migration (re-hash on re-authentication)
- Race detector: Keep probabilistic execution enabled
- E2E coverage: Test BOTH /service/* and /browser/*, priority /service/*
- Generated code exemption: OpenAPI + GORM + protobuf
- Sampling: All strategies in OTLP config, tail-based uncommented
- Health check failure: K8s remove LB + restart; Docker mark unhealthy + continue
- Federation timeouts: Per-service required (identity_timeout, jose_timeout, ca_timeout)
- API versioning: Backward compatible (support N-1)
- DNS caching: No caching, lookup every request
- Connection pools: Configurable, hot reloadable
- Read replicas: Removed, all reads to primary
- PostgreSQL workflows: Optional, use test-containers
- Docker pre-pull: Only for workflows using Docker
