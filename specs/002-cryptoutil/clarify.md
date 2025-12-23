# cryptoutil Iteration 2 - Clarifications and Answers

**Last Updated**: December 22, 2025
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
- Admin Port Assignments:
  - KMS: 9090 (all KMS instances share, bound to 127.0.0.1)
  - Identity: 9091 (all 5 Identity services share)
  - CA: 9092 (all CA instances share)
  - JOSE: 9093 (all JOSE instances share)
- Security: IP restriction (localhost only), optional mTLS, minimal middleware
- Endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
- NOT exposed in Docker port mappings

**Rationale for Unique Admin Ports**:

- Admin ports bound to 127.0.0.1 only (not externally accessible)
- Docker Compose: Each service instance = separate container with isolated network namespace
- Same admin port can be reused across instances of same product without collision
- Multiple instances: Admin port 0 in all unit tests, Admin internal 9090/9091/9092/9093 port in docker compose, Admin unique external port mapping per instance

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
- **Health Check Before Retry**: Poll `/admin/v1/livez` endpoint (fast liveness check) before resuming traffic

---

### Service Template Extraction

**Q**: When should the service template be extracted and how should it be validated?

**A**: Extract template in Phase 6, validate with Learn-PS demonstration service.

**Template Components** (extracted from KMS reference implementation):

- Two HTTPS servers (public + admin)
- Two public API paths (`/browser/api/v1/*` vs `/service/api/v1/*`)
- Three admin endpoints (`/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`)
- Database abstraction (PostgreSQL || SQLite dual support, GORM)
- OpenTelemetry integration (OTLP traces, metrics, logs)
- Config management (YAML files + CLI flags, Docker secrets support)

**Template Parameterization**:

- Constructor injection for configuration, handlers, middleware
- Service-specific OpenAPI specs passed to template
- Business logic separated from infrastructure concerns

**Validation Strategy**:

- Learn-PS service MUST use extracted template
- Learn-PS MUST pass all unit/integration/E2E tests
- Deep analysis MUST show no blockers to migrate existing services
- Only after Learn-PS succeeds can production services migrate

**Migration Priority**:

1. **learn-ps FIRST** (Phase 7) - Validate template reusability
2. **One service at a time** - Sequentially refactor jose-ja, pki-ca, identity services
3. **sm-kms LAST** - Only after ALL other services running excellently on template

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
    test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
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

---

## Service Template and Migration Strategy

### Service Template Migration Priority

**Q**: Should identity services (authz, idp, rs) be refactored to use the extracted service template immediately after Learn-PS validation, or later?

**A** (Source: CLARIFY-QUIZME-01 Q1, 2025-12-22):

Identity services will be migrated **LAST** in the following sequence:

1. **learn-ps** (Phase 7): Validate service template first
2. **JOSE and CA** (Phases after learn-ps): Migrate next, one at a time, to allow adjustments to the service template to accommodate JOSE and CA service patterns
3. **Identity services** (Final phase): Migrate last, ordered by Authz → IdP → RS → RP → SPA

**Rationale**: Learn-PS will validate the service template first, then JOSE and CA migrations will drive template refinements to support different service patterns. Identity services migrate last to benefit from a mature, battle-tested template.

---

### Monitoring and Metrics Architecture

**Q**: Should admin ports expose `/admin/v1/metrics` endpoint for external monitoring tools (Prometheus, Grafana)?

**A** (Source: CLARIFY-QUIZME-01 Q2, 2025-12-22):

**CRITICAL**: `/admin/v1/metrics` endpoint is a **MISTAKE** and MUST be removed from the project entirely.

**Correct Architecture**:

- ALL services MUST use OTLP protocol to **push** metrics, tracing, and logging to OpenTelemetry Collector Contrib
- **NEVER** use pull or scrape patterns (no Prometheus scraping of service endpoints)
- OpenTelemetry Collector Contrib uses OTLP to forward metrics, tracing, and logging to Grafana LGTM

**Action Required**:

- Remove all references to `/admin/v1/metrics` from codebase
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

## Status Summary

**Last Review**: December 22, 2025
**Next Actions**:

1. Update constitution.md with architectural decisions (service template migration, MFA factors, federation fallback)
2. Update spec.md with finalized requirements (SQLite production, cert profiles, load testing, E2E coverage, telemetry retention)
3. Update copilot instructions to remove `/admin/v1/metrics` references
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
