# cryptoutil Iteration 2 - Clarifications and Answers

This document provides detailed answers to questions arising from the specification. It serves as authoritative guidance for implementation decisions, architectural patterns, and technical trade-offs.

---

## Table of Contents

1. [Architecture and Design Patterns](#architecture-and-design-patterns)
2. [Phase 1: Test Performance Optimization](#phase-1-test-performance-optimization)
3. [Phase 2: Coverage Target Enforcement](#phase-2-coverage-target-enforcement)
4. [Phase 3: CI/CD Workflow Fixes](#phase-3-cicd-workflow-fixes)
5. [Phase 4: Mutation Testing](#phase-4-mutation-testing)
6. [Phase 5: Hash Service Refactoring](#phase-5-hash-service-refactoring)
7. [Phase 6: Service Template Extraction](#phase-6-service-template-extraction)
8. [Phase 7: Learn-PS Demonstration](#phase-7-learn-ps-demonstration)
9. [Cross-Cutting Concerns](#cross-cutting-concerns)

---

## Architecture and Design Patterns

### Q1: What is the dual-server architecture pattern and why is it mandatory?

**A**: ALL services MUST implement dual HTTPS endpoints:

**Public HTTPS Server** (`0.0.0.0:<configurable_port>`):

- Purpose: User-facing APIs and browser UIs
- Ports: 8080 (JOSE/KMS/Identity AuthZ), 8081 (KMS Postgres 1, Identity IdP), 8082 (KMS Postgres 2, Identity RS), 8443 (CA)
- Security: OAuth 2.1 tokens, CORS/CSRF/CSP, rate limiting, TLS 1.3+
- API contexts:
  - `/browser/api/v1/*` - Session-based (HTTP Cookie) for SPA
  - `/service/api/v1/*` - Token-based (HTTP Authorization header) for backends

**Private HTTPS Server** (`127.0.0.1:9090` or `127.0.0.1:9443` for CA):

- Purpose: Internal admin tasks, health checks, metrics
- Security: IP restriction (localhost only), optional mTLS, minimal middleware
- Endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
- NOT exposed in Docker port mappings

**Why Mandatory**:

- **Security**: Admin endpoints isolated from public network
- **Performance**: Health probes don't compete with user traffic
- **Reliability**: Kubernetes/Docker health checks work even if public API overloaded
- **Compliance**: Separation of concerns for audit requirements

**Implementation Status**:

- ✅ **KMS**: Complete reference implementation with dual servers
- ⚠️ **Identity**: Servers exist but not integrated with unified `cryptoutil` command
- ❌ **JOSE**: Missing admin server entirely
- ❌ **CA**: Missing admin server entirely

**Phase 3.5 Deliverable**: All services follow KMS dual-server pattern.

---

### Q2: What is the unified command interface requirement?

**A**: ALL services MUST be accessible via single `cryptoutil` command for consistency:

```bash
# KMS (✅ reference implementation)
cryptoutil kms start --config=kms.yml
cryptoutil kms status

# Identity (⚠️ needs cmd integration)
cryptoutil identity start --config=identity.yml
cryptoutil identity status

# JOSE (❌ needs admin server + cmd integration)
cryptoutil jose start --config=jose.yml
cryptoutil jose status

# CA (❌ needs admin server + cmd integration)
cryptoutil ca start --config=ca.yml
cryptoutil ca status
```

**Implementation Pattern** (from KMS):

- Each service has `internal/cmd/cryptoutil/<service>/` package
- Cobra commands registered in `cmd/cryptoutil/main.go`
- `start` command launches both public and admin servers
- `status` command queries `/admin/v1/healthz` endpoint

**Rationale**:

- Unified operational interface reduces training overhead
- Consistent command structure across all products
- Simplifies deployment automation and troubleshooting

---

### Q3: Why are there multiple instances per service in Docker Compose?

**A**: Multi-instance deployments demonstrate production patterns:

**KMS** (3 instances):

- `kms-sqlite` (port 8080): In-memory SQLite for development/testing
- `kms-postgres-1` (port 8081): PostgreSQL instance 1
- `kms-postgres-2` (port 8082): PostgreSQL instance 2

**JOSE** (planned - 3 instances):

- `jose-sqlite` (port 8080)
- `jose-postgres-1` (port 8081)
- `jose-postgres-2` (port 8082)

**Identity** (3 microservices):

- `identity-authz` (port 8080): Authorization server
- `identity-idp` (port 8081): Identity provider
- `identity-rs` (port 8082): Resource server

**Rationale**:

- **Multi-backend support**: Validate code works with both PostgreSQL and SQLite
- **Database-specific configs**: Each instance has unique CORS origins, OTLP service names
- **Production validation**: Kubernetes typically runs multiple replicas per service
- **Load balancing testing**: E2E tests can distribute load across instances

**Why NOT replicas**: Fixed instances allow database-specific configurations (SQLite vs PostgreSQL connection strings), while replicas assume identical config.

---

### Q4: What is the service mesh topology?

**A**: Services communicate through defined network zones:

```
External Clients (Browsers, Mobile Apps, Services)
    ↓ HTTPS (TLS 1.3+), OAuth 2.1 tokens
Reverse Proxy / API Gateway (Traefik, nginx, Kong - optional)
    ↓
┌────────────┬─────────────┬──────────┬────────┐
│   JOSE     │  Identity   │   KMS    │   CA   │
│ Port: 8090 │AuthZ: 8080  │Port: 8080│Port:8443│
│            │ IdP: 8081   │          │        │
│            │  RS: 8082   │          │        │
└─────┬──────┴──────┬──────┴─────┬────┴───┬────┘
      │ Admin:9092  │ Admin:9090 │Admin:9090│Admin:9443
      │ (127.0.0.1) │ (127.0.0.1)│(127.0.0.1)│(127.0.0.1)
      ↓             ↓            ↓          ↓
Kubernetes / Docker Health Checks
      ↓             ↓            ↓          ↓
PostgreSQL Database (shared dev, isolated prod)
      ↓             ↓            ↓          ↓
OpenTelemetry Collector (Traces, Metrics, Logs)
```

**Network Zones**:

| Zone | Services | Access Control |
|------|----------|----------------|
| **Public** | All (8080-8443) | OAuth 2.1, rate limit, TLS 1.3+ |
| **Admin** | All (9090/9443) | Localhost only, optional mTLS |
| **Database** | PostgreSQL (5432) | Password auth, network isolation |
| **Telemetry** | OTLP (4317/4318) | Service mesh only, no external |

---

## Phase 1: Test Performance Optimization

### Q5: What are the test execution time requirements?

**A**: Test performance targets enforce fast feedback loops:

**Requirements**:

- **Individual package**: <30 seconds per package
- **Total test suite**: <100 seconds (all packages)
- **Race detector**: <200 seconds (slower due to CGO_ENABLED=1 overhead)

**Current Status**: Performance varies by package - optimization needed for slower packages.

**Strategies for Optimization**:

1. **Probability-based execution**: Use `TestProbQuarter` (25%) or `TestProbTenth` (10%) for algorithm/key size variants
2. **Parallel subtests**: ALWAYS use `t.Parallel()` in table-driven tests
3. **Test data efficiency**: Reuse test fixtures, avoid redundant setup
4. **Mock external dependencies**: Replace slow network calls with mocks where appropriate

**Anti-patterns to Avoid**:

- Sequential package execution (`-p=1`) - hides concurrency bugs
- Non-parallel subtests - misses race conditions
- Running all key size variants at 100% - minimal coverage value for 10x time cost

**Validation Commands**:

```bash
# Measure per-package time
go test ./internal/jose -v -count=1 2>&1 | grep -E "^(ok|FAIL)"

# Measure total suite time
time go test ./... -cover -shuffle=on

# Race detector (expected 2x slower)
time go test ./... -race -shuffle=on
```

---

### Q6: What is probability-based test execution?

**A**: Probability-based execution reduces test time while maintaining statistical coverage confidence.

**Magic Constants** (from `internal/shared/magic/magic_testing.go`):

- `TestProbAlways` = 100% (run every time)
- `TestProbHalf` = 50% (run 50% of time)
- `TestProbQuarter` = 25% (run 25% of time)
- `TestProbTenth` = 10% (run 10% of time)
- `TestProbHundredth` = 1% (run 1% of time)

**When to Use**:

- ✅ **Multiple key sizes of same algorithm** (RSA 2048/3072/4096, AES 128/192/256)
- ✅ **Multiple variants of same operation** (HMAC-SHA256/384/512, ECDSA P-256/384/521)
- ✅ **Large test suites** (>50 test cases with redundant coverage)
- ❌ **Fundamentally different algorithms** (RSA vs ECDSA vs EdDSA - always test all)
- ❌ **Business logic branches** (error paths, edge cases - always test all)
- ❌ **Small test suites** (<20 cases - overhead not worth it)

**Pattern**:

```go
tests := []struct {
    name string
    alg  Algorithm
    prob float64  // Execution probability
}{
    {name: "RSA_2048", alg: AlgRS256, prob: magic.TestProbAlways},   // Base case: 100%
    {name: "RSA_3072", alg: AlgRS384, prob: magic.TestProbQuarter},  // Variant: 25%
    {name: "RSA_4096", alg: AlgRS512, prob: magic.TestProbTenth},    // Variant: 10%
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        if rand.Float64() > tt.prob { t.Skip("probabilistic skip") }
        t.Parallel()
        // test logic
    })
}
```

**Rationale**:

- **Faster CI feedback**: 10% execution = 10x faster for variant-heavy tests
- **Statistical confidence**: Bugs eventually caught across multiple CI runs
- **Production validation**: Base algorithms tested 100%, variants sampled

---

### Q7: How do we prevent Windows Firewall exception prompts in tests?

**A**: ALWAYS bind to `127.0.0.1` (loopback only), NEVER `0.0.0.0` (all interfaces).

**Why This Matters**:

- Binding to `0.0.0.0` triggers Windows Firewall exception prompts (blocks CI/CD automation)
- Binding to `127.0.0.1` does NOT trigger firewall prompts (loopback traffic exempt)

**Correct Patterns**:

```go
// ✅ CORRECT: Bind to loopback only (no firewall prompt)
addr := fmt.Sprintf("%s:%d", magic.IPv4Loopback, port) // "127.0.0.1"
listener, err := net.Listen("tcp", addr)

// ❌ WRONG: Bind to all interfaces (triggers firewall prompt)
listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
```

**Configuration Files**:

```yaml
# ✅ CORRECT: Test configs use loopback
bind_address: 127.0.0.1

# ❌ WRONG: Production pattern in test configs
bind_address: 0.0.0.0  # Only for Docker containers, NEVER for local tests
```

**Docker vs Local Development**:

- **Docker containers**: Use `0.0.0.0` (containers isolated by default)
- **Local tests**: Use `127.0.0.1` (avoid firewall prompts)
- **Integration tests**: Use hardcoded ports on `127.0.0.1` (18080, 18081, 18082)

**Enforcement**: Codified in `01-07.security.instructions.md` "Windows Firewall Exception Prevention" section.

---

## Phase 2: Coverage Target Enforcement

### Q8: What are the mandatory coverage targets?

**A**: Coverage targets are BLOCKING requirements with NO EXCEPTIONS:

**Targets**:

- **Production packages** (internal/jose, internal/identity, internal/kms): ≥95%
- **Infrastructure packages** (internal/cmd/cicd/*): ≥100%
- **Utility packages** (internal/shared/*): ≥100%
- **Main functions**: 0% acceptable if `internalMain()` ≥95%

**Why "No Exceptions" Rule Matters**:

- Accepting 70% because "it's better than 60%" leaves 25 points of technical debt
- "This package is mostly error handling" → Add error path tests
- "This is just a thin wrapper" → Still needs 95% coverage
- Incremental improvements accumulate debt; enforce targets strictly

**Enforcement Pattern**:

```bash
# DON'T: Celebrate improvement without meeting target
coverage_before=60.0
coverage_after=70.0
echo "✅ Improved by 10 percentage points!"  # ❌ Still 25 points below target

# DO: Enforce target, reject anything below 95%
if [ "$coverage" -lt 95 ]; then
    echo "❌ BLOCKING: Coverage $coverage% < 95% target"
    echo "Required: Write tests for RED lines in coverage HTML"
    exit 1
fi
```

**Validation Commands**:

```bash
# Generate coverage report
go test ./internal/jose -coverprofile=./test-output/coverage_jose.out

# Check per-function coverage
go tool cover -func ./test-output/coverage_jose.out | Where-Object { $_ -match 'jose.*\.go:' -and $_ -match '(\d+\.\d+)%' -and [double]$Matches[1] -lt 95.0 }

# Visual HTML analysis
go tool cover -html=./test-output/coverage_jose.out -o ./test-output/coverage_jose.html
```

---

### Q9: What is the main() function pattern for maximum coverage?

**A**: ALL main() functions MUST be thin wrappers calling co-located testable functions.

**Pattern**:

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

**Why Mandatory**:

- **95%+ coverage achievable**: main() 0% acceptable when internalMain() ≥95%
- **Dependency injection**: Tests inject mocks for args, stdin, stdout, stderr
- **Exit code testing**: Tests verify return codes without terminating test process
- **Happy/sad path testing**: Test all branches (missing args, invalid input, success)

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

func TestInternalMain_MissingArgs(t *testing.T) {
    args := []string{"cmd"}  // Missing required arg
    stderr := &bytes.Buffer{}

    exitCode := internalMain(args, nil, nil, stderr)

    require.Equal(t, 1, exitCode)
    require.Contains(t, stderr.String(), "usage")
}
```

**Pattern for ALL Commands**:

- `cmd/cicd/cicd.go` → `internalMain()` testable function
- `cmd/cryptoutil/main.go` → `internalMain()` testable function
- `cmd/workflow/main.go` → `internalMain()` testable function
- `internal/cmd/cicd/*/main.go` → `internalMain()` testable function

---

### Q10: What is the mandatory workflow for coverage analysis?

**A**: ALWAYS generate baseline coverage report BEFORE writing tests:

**MANDATORY WORKFLOW - NEVER SKIP**:

1. **Generate baseline coverage**: `go test ./pkg -coverprofile=./test-output/coverage_pkg.out`
2. **Analyze uncovered lines**: `go tool cover -html=./test-output/coverage_pkg.out -o ./test-output/coverage_pkg.html`
3. **Identify specific gaps**: Open HTML report, look for RED (uncovered) lines
4. **Target specific branches**: Write tests ONLY for identified uncovered code
5. **Verify improvement**: Re-run coverage, confirm gaps filled

**Why This Matters**:

- Prevents wasted effort writing tests for already-covered code
- Targets exact missing branches, not guesswork
- Measurable progress per test addition
- Avoids trial-and-error test deletion cycles

**Example Gap Analysis**:

```bash
# Step 1: Generate baseline
go test ./internal/jose -coverprofile=./test-output/coverage_jose_baseline.out

# Step 2: Check per-function coverage
go tool cover -func ./test-output/coverage_jose_baseline.out | Where-Object { $_ -match 'jose.*\.go:' -and $_ -match '(\d+\.\d+)%' -and [double]$Matches[1] -lt 90.0 }

# Step 3: Visual analysis of specific files
go tool cover -html=./test-output/coverage_jose_baseline.out -o ./test-output/coverage_jose_baseline.html
# Open HTML, find RED lines in low-coverage functions

# Step 4: Write targeted tests for RED lines only
# Step 5: Re-run and verify improvement
go test ./internal/jose -coverprofile=./test-output/coverage_jose_new.out
```

**Anti-pattern**: Writing 60+ tests without baseline analysis → 0% coverage improvement because tests duplicated already-covered paths.

---

## Phase 3: CI/CD Workflow Fixes

### Q11: Which workflows require PostgreSQL service containers?

**A**: ANY workflow running `go test` on database-dependent packages MUST include PostgreSQL service.

**MANDATORY Configuration**:

```yaml
env:
  POSTGRES_HOST: localhost
  POSTGRES_PORT: 5432
  POSTGRES_NAME: cryptoutil_test
  POSTGRES_USER: cryptoutil
  POSTGRES_PASS: cryptoutil_test_password

jobs:
  test-job:
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

**Affected Workflows**:

- ✅ **ci-race**: Tests with race detector (requires database)
- ✅ **ci-mutation**: Gremlins mutation testing (requires database)
- ✅ **ci-coverage**: Coverage analysis (requires database)
- ✅ **ci-identity-validation**: Identity-specific tests (requires database)

**Why Required**:

- Tests in `internal/kms/server/repository/sqlrepository` require PostgreSQL
- Tests in `internal/identity/domain/repository` require PostgreSQL
- Without service: Tests fail with "connection refused" after 2.5s timeout
- With service: PostgreSQL ready before tests start (50s startup window)

**Workflows NOT Requiring PostgreSQL**:

- ❌ **ci-quality**: Linting/formatting only
- ❌ **ci-benchmark**: Performance benchmarks (uses mocks)
- ❌ **ci-fuzz**: Fuzz testing (uses in-memory data)
- ❌ **ci-sast/ci-gitleaks**: Static analysis (no runtime)

---

### Q12: What are the CI/CD workflow time targets?

**A**: Workflows have aggressive time targets for fast feedback:

| Workflow | Target | PostgreSQL | Purpose |
|----------|--------|------------|---------|
| ci-quality | <5 min | ❌ | Linting, formatting, builds |
| ci-coverage | <10 min | ✅ | Coverage ≥95% validation |
| ci-race | <15 min | ✅ | Race condition detection |
| ci-mutation | <45 min | ✅ | Mutation efficacy ≥80% |
| ci-benchmark | <10 min | ❌ | Performance benchmarks |
| ci-fuzz | <10 min | ❌ | Fuzz testing |
| ci-sast | <5 min | ❌ | Static security (gosec) |
| ci-gitleaks | <2 min | ❌ | Secrets scanning |
| ci-dast | <15 min | ❌ | Dynamic security (Nuclei, ZAP) |
| ci-e2e | <20 min | ❌ | End-to-end Docker Compose |
| ci-load | <30 min | ❌ | Load testing (Gatling) |
| ci-identity-validation | <5 min | ✅ | Identity-specific tests |

**Critical Path** (quality + coverage + race): <10 minutes
**Full Suite**: <60 minutes for all workflows

**Optimization Strategies**:

- Parallel workflow execution (GitHub Actions concurrency)
- Path filters to skip workflows on docs-only changes
- Cached Go dependencies (`actions/setup-go@v6` with `cache: true`)
- Probability-based test execution for large suites

---

### Q13: What is the health check standardization pattern?

**A**: ALL Docker Compose services MUST use consistent health check patterns:

**Alpine Containers** (use `wget`):

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
  start_period: 10s
  interval: 5s
  retries: 5
  timeout: 5s
```

**Non-Alpine Containers** (use `curl`):

```yaml
healthcheck:
  test: ["CMD", "curl", "-k", "-f", "-s", "https://127.0.0.1:9090/admin/v1/livez"]
  start_period: 10s
  interval: 5s
  retries: 5
  timeout: 5s
```

**Parameters**:

- `start_period: 10s`: Allow 10 seconds before first check (service initialization)
- `interval: 5s`: Check every 5 seconds after start_period
- `retries: 5`: Allow 5 consecutive failures before marking unhealthy
- `timeout: 5s`: Each health check must complete within 5 seconds

**Total Startup Window**: start_period + (interval × retries) = 10s + (5s × 5) = 35 seconds

**Admin Endpoints for Health Checks**:

- `/admin/v1/livez`: Process alive (lightweight check)
- `/admin/v1/readyz`: Dependencies healthy (database, barrier services)
- `/admin/v1/healthz`: Combined health check (both liveness and readiness)

**Why Standardized**:

- Consistent retry logic across all services
- Predictable startup times for orchestration
- Separate liveness (process alive) from readiness (dependencies healthy)

---

## Phase 4: Mutation Testing

### Q14: What are the mutation testing requirements?

**A**: Mutation testing validates test suite effectiveness using gremlins:

**Requirements**:

- **Minimum**: ≥80% gremlins score per package
- **Target**: ≥98% efficacy (mutation detection rate)
- **Focus**: Business logic, parsers, validators, crypto operations

**Configuration** (`.gremlins.yaml`):

```yaml
threshold:
  efficacy: 98  # Target: ≥98% test efficacy
  mutant-coverage: 90  # Target: ≥90% mutant coverage

workers: 4  # Parallel mutant execution
test-cpu: 2  # CPU per test run
timeout-coefficient: 2  # Timeout multiplier
```

**Running Mutation Tests**:

```bash
# Single package
gremlins unleash ./internal/jose

# All packages (excluding integration tests)
gremlins unleash --tags=!integration

# Generate HTML report
gremlins unleash ./internal/jose --output=./test-output/gremlins_jose.html
```

**Metrics**:

- **Efficacy**: Percentage of mutants killed by tests (target ≥98%)
- **Mutant Coverage**: Percentage of code mutated (target ≥90%)
- **Timeout**: Mutant tests that exceed timeout (should be 0%)
- **Not Covered**: Code not reached by tests (should match coverage gaps)

**Interpretation**:

- **High efficacy + high coverage**: Strong test suite
- **High efficacy + low coverage**: Good tests, but incomplete
- **Low efficacy + high coverage**: Tests pass but don't validate behavior
- **Low efficacy + low coverage**: Weak test suite, needs work

**Package-level Parallelization**:

- Use workflow matrix to run gremlins on packages concurrently
- Reduces total mutation testing time from 45+ minutes to <15 minutes

---

### Q15: What mutations does gremlins apply?

**A**: Gremlins applies code mutations to validate test effectiveness:

**Arithmetic Mutations**:

- `+` → `-`, `*`, `/`, `%`
- `-` → `+`, `*`, `/`, `%`
- `*` → `+`, `-`, `/`, `%`
- `/` → `+`, `-`, `*`, `%`

**Conditional Mutations**:

- `==` → `!=`, `<`, `>`, `<=`, `>=`
- `!=` → `==`, `<`, `>`, `<=`, `>=`
- `<` → `>`, `<=`, `>=`, `==`, `!=`
- `&&` → `||`
- `||` → `&&`

**Increment/Decrement Mutations**:

- `i++` → `i--`
- `i--` → `i++`
- `++i` → `--i`
- `--i` → `++i`

**Invert Negatives Mutations**:

- `-x` → `+x`
- `+x` → `-x`

**Remove Mutations**:

- Delete conditionals (if, for, switch)
- Remove function calls (side effects)

**Return Mutations**:

- Change return values (nil → error, true → false, 0 → 1)

**Why Mutation Testing Matters**:

- **Coverage ≠ Quality**: 100% coverage doesn't mean tests validate behavior
- **Detect Weak Assertions**: Tests that pass with incorrect logic
- **Validate Error Paths**: Ensure error handling is actually tested

**Example**:

```go
// Original code
func divide(a, b int) (int, error) {
    if b == 0 { return 0, errors.New("div by zero") }
    return a / b, nil
}

// Mutation: Change `==` to `!=`
func divide(a, b int) (int, error) {
    if b != 0 { return 0, errors.New("div by zero") }  // WRONG
    return a / b, nil
}

// Weak test (passes with mutation)
func TestDivide(t *testing.T) {
    result, _ := divide(10, 2)
    _ = result  // No assertion!
}

// Strong test (kills mutation)
func TestDivide(t *testing.T) {
    result, err := divide(10, 2)
    require.NoError(t, err)
    require.Equal(t, 5, result)  // Fails with mutation
}
```

---

## Phase 5: Hash Service Refactoring

### Q16: What are the 4 hash registry types and when to use each?

**A**: Hash service provides 4 registry types based on input entropy and determinism requirements:

**1. LowEntropyRandomHashRegistry** (PBKDF2-based, salted):

- **Algorithm**: PBKDF2-HMAC-SHA256/384/512 (OWASP recommended rounds)
- **Salt**: Random salt generated per hash
- **Use Case**: Password hashing (user passwords, API keys)
- **Output Format**: Includes version, salt, iterations, hash
- **FIPS 140-3**: ✅ Approved (PBKDF2, HMAC-SHA256/384/512)

**2. LowEntropyDeterministicHashRegistry** (PBKDF2-based, no salt):

- **Algorithm**: PBKDF2-HMAC-SHA256/384/512
- **Salt**: Empty salt (deterministic output)
- **Use Case**: Replay-resistant tokens (CSRF tokens, session IDs)
- **Output Format**: Includes version, iterations, hash
- **FIPS 140-3**: ✅ Approved

**3. HighEntropyRandomHashRegistry** (HKDF-based, salted):

- **Algorithm**: HKDF-HMAC-SHA256/384/512
- **Salt**: Random salt generated per hash
- **Use Case**: Key derivation from high-entropy inputs (master keys → sub-keys)
- **Output Format**: Includes version, salt, hash
- **FIPS 140-3**: ✅ Approved (HKDF, HMAC-SHA256/384/512)

**4. HighEntropyDeterministicHashRegistry** (HKDF-based, no salt):

- **Algorithm**: HKDF-HMAC-SHA256/384/512
- **Salt**: Empty salt (deterministic output)
- **Use Case**: Deterministic key derivation (content-addressed keys, fingerprints)
- **Output Format**: Includes version, hash
- **FIPS 140-3**: ✅ Approved

**Selection Criteria**:

| Input Entropy | Deterministic? | Registry Type | Algorithm |
|---------------|----------------|---------------|-----------|
| Low (password) | No (random salt) | LowEntropyRandom | PBKDF2 |
| Low (password) | Yes (no salt) | LowEntropyDeterministic | PBKDF2 |
| High (key) | No (random salt) | HighEntropyRandom | HKDF |
| High (key) | Yes (no salt) | HighEntropyDeterministic | HKDF |

---

### Q17: What is the version management strategy for hash registries?

**A**: Version management supports algorithm upgrades without breaking existing hashes:

**Version Selection** (automatic based on input size):

- **v1**: 0-31 bytes → SHA-256-based algorithm
- **v2**: 32-47 bytes → SHA-384-based algorithm
- **v3**: 48+ bytes → SHA-512-based algorithm

**Registry API** (consistent across all 4 types):

```go
type HashRegistry interface {
    // HashWithLatest uses current version for new hashes
    HashWithLatest(input []byte) (string, error)

    // HashWithVersion uses specific version (for testing/migration)
    HashWithVersion(input []byte, version int) (string, error)

    // Verify validates input against hash using version metadata
    Verify(input []byte, hashed string) (bool, error)
}
```

**Hash Output Format** (includes version metadata):

```
$pbkdf2-sha256$v1$i=600000$salt=<base64>$hash=<base64>
$hkdf-sha384$v2$salt=<base64>$hash=<base64>
$pbkdf2-sha512$v3$i=600000$hash=<base64>  # deterministic (no salt)
```

**Version Upgrade Workflow**:

1. Deploy new code with v2 support (v1 still default)
2. Verify v2 hashes validate correctly
3. Update config to use v2 as default for new hashes
4. Old v1 hashes still verified successfully (backward compatibility)
5. Optional: Migrate v1 hashes to v2 on user login (opportunistic upgrade)

**Benefits**:

- **Forward Compatibility**: New versions added without breaking old hashes
- **Algorithm Agility**: Upgrade from SHA-256 to SHA-512 without downtime
- **Automatic Selection**: Input size determines version (no manual selection)
- **Backward Compatibility**: Old hashes always verifiable

---

### Q18: Why use PBKDF2 for low-entropy inputs?

**A**: PBKDF2 is FIPS 140-3 approved for password-based key derivation:

**Why PBKDF2**:

- ✅ **FIPS 140-3 Approved**: Meets federal compliance requirements
- ✅ **OWASP Recommended**: 600,000+ iterations for SHA-256 (2023 guidance)
- ✅ **Brute-Force Resistant**: Iteration count increases compute cost for attackers
- ✅ **Industry Standard**: Widely adopted, well-tested, proven secure

**Why NOT bcrypt/scrypt/Argon2**:

- ❌ **NOT FIPS-Approved**: Bcrypt, scrypt, Argon2 banned in FIPS 140-3 mode
- ❌ **Compliance Risk**: Federal/DoD projects MUST use FIPS-approved algorithms

**PBKDF2 Configuration**:

```go
iterations := 600000  // OWASP 2023 recommendation for PBKDF2-HMAC-SHA256
salt := make([]byte, 32)  // 256-bit random salt
_, _ = crand.Read(salt)

hash := pbkdf2.Key(password, salt, iterations, 32, sha256.New)
```

**Iteration Count Guidance**:

| Algorithm | Minimum Iterations | Recommended | Target Time |
|-----------|-------------------|-------------|-------------|
| PBKDF2-HMAC-SHA256 | 600,000 | 600,000+ | 100ms |
| PBKDF2-HMAC-SHA384 | 210,000 | 210,000+ | 100ms |
| PBKDF2-HMAC-SHA512 | 210,000 | 210,000+ | 100ms |

**Source**: OWASP Password Storage Cheat Sheet (2023)

---

### Q19: Why use HKDF for high-entropy inputs?

**A**: HKDF is FIPS 140-3 approved for key derivation from high-entropy keys:

**Why HKDF**:

- ✅ **FIPS 140-3 Approved**: RFC 5869 KDF approved for federal use
- ✅ **High-Entropy Optimized**: No iteration count overhead (input already strong)
- ✅ **Fast Performance**: Suitable for key hierarchies (root → intermediate → material)
- ✅ **Extract-then-Expand**: Separates entropy extraction from key expansion

**HKDF vs PBKDF2**:

| Aspect | HKDF | PBKDF2 |
|--------|------|--------|
| **Input Entropy** | High (≥128 bits) | Low (<128 bits) |
| **Iteration Count** | 1 (no iterations) | 600,000+ |
| **Performance** | Fast (single hash) | Slow (intentional) |
| **Use Case** | Key derivation | Password hashing |
| **FIPS Status** | ✅ Approved | ✅ Approved |

**HKDF Configuration**:

```go
salt := make([]byte, 32)  // 256-bit random salt (or deterministic)
_, _ = crand.Read(salt)

hkdf := hkdf.New(sha256.New, masterKey, salt, info)
derivedKey := make([]byte, 32)
_, _ = io.ReadFull(hkdf, derivedKey)
```

**Use Cases**:

- **Root Key → Intermediate Keys**: KMS key hierarchy
- **Master Key → Per-Tenant Keys**: Multi-tenancy isolation
- **Content-Addressed Keys**: Deterministic key derivation from content hashes

---

## Phase 6: Service Template Extraction

### Q20: What is the service template goal?

**A**: Extract reusable service template from KMS server for 8 PRODUCT-SERVICE instances:

**8 Target Instances**:

1. **sm-kms** - Secrets Manager - Key Management System
2. **pki-ca** - Public Key Infrastructure - Certificate Authority
3. **jose-ja** - JOSE - JWK Authority
4. **identity-authz** - Identity - Authorization Server
5. **identity-idp** - Identity - Identity Provider
6. **identity-rs** - Identity - Resource Server
7. **identity-rp** - Identity - Relying Party (BFF pattern)
8. **identity-spa** - Identity - Single Page Application (static hosting)

**Goal**: All 8 services built from same template, diverge only in:

- API endpoints (OpenAPI specs)
- Business logic handlers
- Database schemas
- Client SDK interfaces
- Barrier services (optional, KMS-specific)

**Benefits**:

- **Faster Development**: Copy-paste-modify instead of build from scratch
- **Consistency**: All services use same infrastructure patterns
- **Maintainability**: Single source of truth for common patterns
- **Quality**: Reuse well-tested, production-hardened components

---

### Q21: What common patterns are extracted to the template?

**A**: Template extracts infrastructure patterns from KMS server:

**Dual HTTPS Servers**:

- Public API server: `0.0.0.0:<configurable_port>`
- Admin API server: `127.0.0.1:9090`
- Health check endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`
- Graceful shutdown: `/admin/v1/shutdown`

**Dual API Paths**:

- `/browser/api/v1/*` - Session-based (HTTP Cookie) for SPA clients
- `/service/api/v1/*` - Token-based (HTTP Authorization header) for service clients
- Shared OpenAPI spec, different request paths and middleware

**Middleware Pipeline**:

- **Browser-only**: CORS, CSRF, CSP (Content Security Policy)
- **Both**: Rate limiting, IP allowlist, authentication, logging
- **Service-only**: mTLS (optional), API key validation

**Database Abstraction**:

- PostgreSQL + SQLite dual support with GORM
- Connection pooling configuration
- Migration management (embedded SQL files)
- Transaction patterns (context-based transaction propagation)

**OpenTelemetry Integration**:

- OTLP traces export (HTTP/gRPC)
- Metrics collection (Prometheus-compatible)
- Logs forwarding (structured logging)
- Correlation IDs for request tracking

**Configuration Management**:

- YAML-based configuration (no environment variables for secrets)
- Docker secrets support (`file:///run/secrets/<secret_name>`)
- Validation on startup (fail-fast for missing required config)

---

### Q22: What are the template packages?

**A**: Template organized into 3 main packages:

**1. ServerTemplate** (`internal/template/server/`):

```
internal/template/server/
├── dual_https.go       # Public + Admin server management
│   - StartPublicServer(ctx, config, handler)
│   - StartAdminServer(ctx, config, handler)
│   - StopServers(ctx)
│   - ReloadConfig(ctx, newConfig)
├── router.go           # Route registration framework
│   - RegisterPublicRoutes(routeFunc)
│   - RegisterAdminRoutes(routeFunc)
│   - RegisterMiddleware(middlewareFunc)
├── middleware.go       # Pipeline builder
│   - ApplyCORS(config)
│   - ApplyCSRF(config)
│   - ApplyCSP(config)
│   - ApplyRateLimit(config)
│   - ApplyIPAllowlist(config)
└── lifecycle.go        # Start/stop/reload lifecycle
    - Graceful shutdown handlers
    - Signal handling (SIGTERM, SIGINT)
    - Cleanup hooks
```

**2. ClientSDK** (`internal/template/client/`):

```
internal/template/client/
├── http_client.go      # HTTP client with mTLS/retry
│   - NewClient(config)
│   - WithAuth(authStrategy)
│   - WithRetry(retryConfig)
│   - WithTimeout(timeout)
├── auth.go             # OAuth 2.1/mTLS/API key strategies
│   - OAuth2Strategy (client_credentials flow)
│   - MTLSStrategy (client certificate)
│   - APIKeyStrategy (X-API-Key header)
└── codegen.go          # OpenAPI-based client generation
    - GenerateClient(openAPISpec)
    - GenerateModels(openAPISpec)
    - GenerateInterfaces(openAPISpec)
```

**3. Database Abstraction** (`internal/template/repository/`):

```
internal/template/repository/
├── dual_db.go          # PostgreSQL + SQLite support
│   - NewDualDB(config) (*gorm.DB, error)
│   - ConfigurePostgreSQL(dsn)
│   - ConfigureSQLite(dsn)
│   - SupportsConcurrentWrites() bool
├── gorm_patterns.go    # Model registration, migrations
│   - RegisterModels(db, models...)
│   - RunMigrations(db, migrationsPath)
│   - AutoMigrate(db, models...)
└── transaction.go      # Transaction handling patterns
    - WithTransaction(ctx, db, fn)
    - GetDB(ctx, baseDB) *gorm.DB
    - RollbackOnError(tx, err)
```

---

### Q23: How does a service customize the template?

**A**: Services customize template through parameterization:

**Constructor Injection**:

```go
template := server.NewServerTemplate(server.Config{
    PublicPort: 8080,
    AdminPort: 9090,
    EnableBarrier: false,  // Service-specific: KMS needs barrier, others don't
    TLSConfig: tlsConfig,
    LogLevel: "info",
})
```

**Route Registration**:

```go
template.RegisterPublicRoutes(func(r fiber.Router) {
    // Service-specific routes
    r.Post("/api/v1/keys", handlers.CreateKey)
    r.Get("/api/v1/keys", handlers.ListKeys)
    r.Get("/api/v1/keys/:id", handlers.GetKey)
    r.Delete("/api/v1/keys/:id", handlers.DeleteKey)
})

template.RegisterAdminRoutes(func(r fiber.Router) {
    // Standard admin endpoints (template-provided)
    // No customization needed - livez/readyz/healthz/shutdown
})
```

**Middleware Configuration**:

```go
template.ApplyMiddleware(middleware.Config{
    CORS: middleware.CORSConfig{
        Origins: []string{"https://jose.example.com"},  // Service-specific
        Methods: []string{"GET", "POST", "PUT", "DELETE"},
        Headers: []string{"Content-Type", "Authorization"},
    },
    RateLimit: middleware.RateLimitConfig{
        RequestsPerMinute: 100,  // Service-specific
        BurstSize: 10,
    },
})
```

**Database Schema**:

```go
// Service-specific GORM models
type ElasticKey struct {
    ID uuid.UUID `gorm:"type:text;primaryKey"`
    Name string `gorm:"type:text;not null"`
    // ... service-specific fields
}

// Register with template
repo := repository.NewDualDB(dbConfig)
repo.RegisterModels(&ElasticKey{}, &MaterialKey{})
repo.RunMigrations("./migrations")
```

**Interface-Based Customization**:

```go
type ServerInterface interface {
    // Service-specific operations
    CreateResource(ctx, req) (resp, error)
    GetResource(ctx, id) (resp, error)
    ListResources(ctx, filters) (resp, error)
}

// Template uses interface, service provides implementation
template.SetHandler(myServiceHandler)
```

---

## Phase 7: Learn-PS Demonstration

### Q24: What is the Learn-PS demonstration service?

**A**: Learn-PS is a working Pet Store service demonstrating template usage:

**Purpose**:

- **Educational**: Show service template in action with realistic API
- **Starting Point**: Customers copy Learn-PS directory, modify for their use case
- **Best Practices**: Production-ready patterns (error handling, testing, deployment)

**Scope**:

- **Product**: Learn (educational/demonstration product line)
- **Service**: PS (Pet Store service)
- **APIs**: 11 endpoints covering pets, orders, customers (full CRUD)

**API Endpoints**:

| Endpoint | Method | Description | Auth Scopes |
|----------|--------|-------------|-------------|
| `/pets` | POST | Create new pet | `write:pets` |
| `/pets` | GET | List pets (paginated) | `read:pets` |
| `/pets/{id}` | GET | Get pet details | `read:pets` |
| `/pets/{id}` | PUT | Update pet | `write:pets` |
| `/pets/{id}` | DELETE | Delete pet | `admin:pets` |
| `/orders` | POST | Create order | `write:orders` |
| `/orders` | GET | List orders | `read:orders` |
| `/orders/{id}` | GET | Get order details | `read:orders` |
| `/customers` | POST | Create customer | `write:customers` |
| `/customers` | GET | List customers | `read:customers` |
| `/customers/{id}` | GET | Get customer details | `read:customers` |

**Authentication**: OAuth 2.1 with scope-based authorization

---

### Q25: What is the Learn-PS database schema?

**A**: Learn-PS uses relational schema with proper foreign keys and constraints:

```sql
-- Pets table
CREATE TABLE pets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Customers table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    total DECIMAL(10,2) NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Order items table
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    pet_id UUID NOT NULL REFERENCES pets(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Design Decisions**:

- **UUIDv7 Primary Keys**: Time-ordered, globally unique identifiers
- **Foreign Key Constraints**: Enforce referential integrity
- **ON DELETE CASCADE**: Auto-delete order items when order deleted
- **CHECK Constraints**: Validate order status enum at database level
- **Timestamps**: Track creation and update times for all entities

---

### Q26: How does Learn-PS use the service template?

**A**: Learn-PS demonstrates template usage with production patterns:

```go
// main.go
func main() {
    // 1. Instantiate ServerTemplate with configuration
    template := server.NewServerTemplate(server.Config{
        PublicPort: 8080,      // Public API server
        AdminPort: 9090,       // Admin API server (health checks)
        EnableBarrier: false,  // No barrier services needed
        TLSConfig: loadTLSConfig(),
        LogLevel: "info",
    })

    // 2. Register API routes
    template.RegisterPublicRoutes(func(r fiber.Router) {
        // Pets endpoints
        r.Post("/api/v1/pets", handlers.CreatePet)
        r.Get("/api/v1/pets", handlers.ListPets)
        r.Get("/api/v1/pets/:id", handlers.GetPet)
        r.Put("/api/v1/pets/:id", handlers.UpdatePet)
        r.Delete("/api/v1/pets/:id", handlers.DeletePet)

        // Orders endpoints
        r.Post("/api/v1/orders", handlers.CreateOrder)
        r.Get("/api/v1/orders", handlers.ListOrders)
        r.Get("/api/v1/orders/:id", handlers.GetOrder)

        // Customers endpoints
        r.Post("/api/v1/customers", handlers.CreateCustomer)
        r.Get("/api/v1/customers", handlers.ListCustomers)
        r.Get("/api/v1/customers/:id", handlers.GetCustomer)
    })

    // 3. Apply middleware (CORS, CSRF, CSP, rate limiting)
    template.ApplyMiddleware(middleware.Config{
        CORS: middleware.CORSConfig{
            Origins: []string{"https://learn-ps.example.com"},
            Methods: []string{"GET", "POST", "PUT", "DELETE"},
            Headers: []string{"Content-Type", "Authorization"},
        },
        RateLimit: middleware.RateLimitConfig{
            RequestsPerMinute: 100,
            BurstSize: 10,
        },
        OAuth: middleware.OAuthConfig{
            Issuer: "https://identity-authz.example.com",
            Audience: "learn-ps",
            RequiredScopes: map[string][]string{
                "POST /api/v1/pets": {"write:pets"},
                "DELETE /api/v1/pets/:id": {"admin:pets"},
            },
        },
    })

    // 4. Start both servers (public + admin)
    template.Start(context.Background())
}
```

**Key Patterns Demonstrated**:

- **Dual-server architecture**: Public API + Admin API separation
- **Route registration**: Clean separation of endpoint handlers
- **Middleware configuration**: Security policies defined in one place
- **OAuth 2.1 integration**: Scope-based authorization per endpoint

---

### Q27: What are the Learn-PS documentation deliverables?

**A**: Learn-PS includes comprehensive documentation for customers:

**1. README.md**:

- Quick start guide (Docker Compose, prerequisites)
- API documentation (endpoint descriptions, request/response examples)
- Authentication guide (OAuth 2.1 token acquisition)
- Development guide (local setup, testing, debugging)

**2. Tutorial Series** (4-part):

- **Part 1: Using Learn-PS** - Deploy service, call APIs, test workflows
- **Part 2: Understanding Learn-PS** - Architecture walkthrough, design decisions
- **Part 3: Customizing Learn-PS** - Modify for custom use case, add endpoints
- **Part 4: Deploying Learn-PS** - Production deployment, monitoring, scaling

**3. Video Demonstration**:

- Service startup and health checks
- API usage examples (create pet, place order, list customers)
- Code walkthrough (template usage, handler implementation)
- Debugging and troubleshooting tips

**4. OpenAPI Specification**:

- Complete API spec with request/response schemas
- OAuth 2.1 security definitions
- Example requests and responses

---

### Q28: What are the Learn-PS quality targets?

**A**: Learn-PS meets same quality standards as production services:

**Quality Targets**:

- **Test Coverage**: ≥95% (production code)
- **Mutation Efficacy**: ≥98% (gremlins score)
- **Test Execution**: ≤12 seconds (full test suite)
- **CI/CD**: Passes all workflows (quality, coverage, race, mutation)

**Why High Standards**:

- **Customer Starting Point**: Customers copy Learn-PS as template
- **Best Practices**: Demonstrates production-ready testing patterns
- **Quality Signal**: Shows cryptoutil project quality standards

**Test Organization**:

- **Unit Tests**: Handler logic, validation, error paths
- **Integration Tests**: Database operations, transaction handling
- **E2E Tests**: Full API workflows (create order → pay → fulfill)
- **Load Tests**: Gatling simulations for capacity planning

**Customer Value**:

- **Working Example**: See service template producing real results
- **Copy-Paste-Modify**: Start with Learn-PS, adapt for use case
- **Production Patterns**: Learn error handling, testing, deployment

---

## Cross-Cutting Concerns

### Q29: What is the CGO ban and its implications?

**A**: CGO is BANNED except for race detector (Go toolchain limitation):

**Rule**: `CGO_ENABLED=0` MANDATORY for builds, tests, Docker, production

**ONLY Exception**: Race detector workflow (`-race` flag) requires `CGO_ENABLED=1`

**Why Banned**:

- **Maximum Portability**: Static linking, cross-compilation, no C toolchain dependencies
- **Simpler Builds**: No C compiler, no platform-specific build issues
- **Smaller Images**: Static binaries without shared library dependencies

**Implications**:

- ❌ **NEVER** use `github.com/mattn/go-sqlite3` (requires CGO)
- ✅ **ALWAYS** use `modernc.org/sqlite` (pure Go, CGO-free)
- ❌ **NEVER** add dependencies requiring CGO
- ✅ **ALWAYS** validate dependencies with `go list -u -m all`

**Race Detector Exception**:

- Go's race detector uses C-based ThreadSanitizer library from LLVM
- Fundamental Go toolchain constraint, not project choice
- Only workflow with `CGO_ENABLED=1` is `ci-race.yml`

---

### Q30: How do we track code quality metrics over time?

**A**: Quality metrics tracked in multiple locations:

**Coverage Tracking**:

- **Baseline Reports**: `./test-output/coverage_<package>_baseline.out`
- **Current Reports**: `./test-output/coverage_<package>_new.out`
- **HTML Reports**: `./test-output/coverage_<package>.html` (visual gap analysis)
- **Per-Function Coverage**: `go tool cover -func ./test-output/coverage_<package>.out`

**Mutation Testing Tracking**:

- **Baseline Reports**: `docs/GREMLINS-TRACKING.md` (tracks efficacy over time)
- **HTML Reports**: `./test-output/gremlins_<package>.html` (detailed mutation results)
- **Package-level Scores**: Track per-package efficacy (target ≥98%)

**CI/CD Metrics**:

- **Workflow Duration**: Track time trends for each workflow (quality, coverage, race)
- **Failure Rates**: Track workflow success/failure rates over time
- **Flaky Test Detection**: Identify intermittent failures with `go test -count=5`

**File Size Tracking**:

- **Soft Limit**: 300 lines (ideal)
- **Medium Limit**: 400 lines (acceptable with justification)
- **Hard Limit**: 500 lines (refactor required)
- **Violation Tracking**: `docs/todos-*.md` lists files exceeding limits

**Documentation**:

- **LESSONS-LEARNED-REGRESSIONS.md**: Tracks historical regressions and solutions
- **P0.X-*.md**: Post-mortems for critical incidents (e.g., format_go self-modification)
- **implement/DETAILED.md**: Implementation timeline with metrics per phase

---

*This clarify.md document is authoritative for implementation decisions. When ambiguities arise, refer to this document first before making assumptions.*
