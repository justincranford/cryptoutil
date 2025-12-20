# cryptoutil Iteration 2 - Clarifications and Answers

**Last Updated**: December 19, 2025
**Purpose**: Authoritative Q&A for implementation decisions, architectural patterns, and technical trade-offs
**Organization**: Topical (merged Round 1 + Round 2 clarifications)

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

**Public HTTPS Server** (`0.0.0.0:<configurable_port>`):

- Purpose: User-facing APIs and browser UIs
- Ports: 8080 (KMS), 8180-8184 (Identity services), 8280 (JOSE), 8380 (CA)
- Security: OAuth 2.1 tokens, CORS/CSRF/CSP, rate limiting, TLS 1.3+
- API contexts:
  - `/browser/api/v1/*` - Session-based (HTTP Cookie) for SPA
  - `/service/api/v1/*` - Token-based (HTTP Authorization header) for backends

**Private HTTPS Server** (Admin endpoints):

- Purpose: Internal admin tasks, health checks, metrics
- **Admin Port Assignments** (Source: SPECKIT-CONFLICTS-ANALYSIS C4, 2025-12-19):
  - KMS: 9090 (all KMS instances share, bound to 127.0.0.1)
  - Identity: 9091 (all 5 Identity services share)
  - CA: 9092 (all CA instances share)
  - JOSE: 9093 (all JOSE instances share)
- Security: IP restriction (localhost only), optional mTLS, minimal middleware
- Endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
- NOT exposed in Docker port mappings

**Rationale for Unique Admin Ports** (Source: CLARIFY-QUIZME2 C4.1, 2025-12-19):

- Admin ports bound to 127.0.0.1 only (not externally accessible)
- Docker Compose: Each service instance = separate container with isolated network namespace
- Same admin port can be reused across instances of same product without collision
- **Multiple instances**: Admin port 0 in all unit tests, Admin internal 9090/9091/9092/9093 port in docker compose, Admin unique external port mapping per instance

**Implementation Status**:

- ✅ KMS: Complete reference implementation
- ⚠️ Identity: Servers exist but not fully integrated
- ❌ JOSE: Missing admin server
- ❌ CA: Missing admin server

---

### Package Coverage Classification

**Q**: Which specific packages require 95% vs 100% coverage?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A2, CLARIFY-QUIZME2 A2.1, 2025-12-19):

**Answer**: D - Case-by-case per package (document each in clarify.md)

**Initial Classification**:

- **Production (95%)**: internal/{jose,identity,kms,ca}
- **Infrastructure (100%)**: internal/cmd/cicd/*
- **Utility (100%)**: internal/shared/*, pkg/*

**Rationale**: Package complexity varies - some "production" packages have simpler logic warranting 100%, while some "utility" packages have complex error handling justifying 95%. Document each package's target in this clarify.md as implementation progresses.

**Documentation Pattern**:

- Add new entries to this section as packages are analyzed
- Justify any deviation from initial classification
- Update constitution.md if patterns emerge

---

### Service Federation Configuration

**Q**: How should services discover and configure federated services (Identity, JOSE)?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A4, 2025-12-19):

- **Round 1 Decision**: A - Static YAML configuration (federation.identity_authz_url, federation.jose_ja_url)

**Q**: Where should federation configuration be stored?

**A** (Source: CLARIFY-QUIZME2 A4.1, 2025-12-19):

- **Round 2 Decision**: A - Each service has own federation section (kms.yml has federation.identity_url, federation.jose_url)

**Q**: How should services handle federated service unavailability?

**A** (Source: CLARIFY-QUIZME2 A4.2, 2025-12-19):

- **Round 2 Decision**: B - Graceful degradation (start but disable federated features)
- **Rationale**: "I assume this is best practice for microservices, if not them requires elaboration and reconsideration"

**Combined Implementation**:

- Each service YAML has `federation:` section
- Service starts even if federated services unreachable
- Federated features disabled until dependencies available
- Log warnings for unavailable federated services
- Periodic retry with exponential backoff

---

### CA Deployment Architecture

**Q**: How many CA instances should we deploy?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS C7, 2025-12-19):

- **Round 1 Decision**: A - 3 instances (matches KMS/JOSE/Identity pattern for consistency)

**Q**: Does CA have different database schema requirements than KMS/JOSE?

**A** (Source: CLARIFY-QUIZME2 C7.1, 2025-12-19):

- **Round 2 Decision**: E - "CA follows same repository patterns as KMS/JOSE, but also needs significant differences (certificates, CRLs, OCSP)"

**Combined Implementation**:

- 3 CA instances: ca-sqlite (8380), ca-postgres-1 (8381), ca-postgres-2 (8382)
- Admin port: 9092 (shared across all CA instances)
- Schema: Shares base repository patterns (audit, config, migrations)
- Schema: Adds certificate-specific tables (certificates, CRLs, OCSP responses)
- Schema: Custom migrations in internal/ca/server/repository/migrations/

---

### Docker Compose Instance Naming

**Q**: How should we name multiple instances in Docker Compose?

**A** (Source: CLARIFY-QUIZME2, 2025-12-19):

- **Decision**: E - Service-product-instance-number (sm-kms-1, sm-kms-2, sm-kms-3) with backend in config

**Naming Convention**:

- Format: `<service>-<product>-<instance>` (e.g., sm-kms-1, pki-ca-2, jose-ja-3)
- Backend (SQLite/PostgreSQL) specified in config, not name
- Instance numbers start at 1 (not 0)
- Consistent across all products

**Examples**:

- sm-kms-1, sm-kms-2, sm-kms-3 (KMS instances)
- pki-ca-1, pki-ca-2, pki-ca-3 (CA instances)
- jose-ja-1, jose-ja-2, jose-ja-3 (JOSE instances)
- identity-authz-1, identity-idp-1, identity-rs-1, identity-rp-1, identity-spa-1

---

## Testing Strategy and Quality Assurance

### Coverage Target Enforcement

**Q**: Should we enforce strict coverage targets (95%/100%) or allow gradual improvement?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A2, 2025-12-19):

- **Strict Enforcement**: 95%+ production, 100%+ infrastructure/utility, NO EXCEPTIONS
- Coverage < 95% is BLOCKING issue requiring immediate remediation
- "Improvement" is NOT success - only "target met" counts

**Why "No Exceptions" Rule Matters**:

- Accepting 70% because "it's better than 60%" leaves 25 points of technical debt
- "This package is mostly error handling" → Add error path tests
- "This is just a thin wrapper" → Still needs 95% coverage
- Incremental improvements accumulate debt; enforce targets strictly

---

### Mutation Testing Thresholds

**Q**: Single high threshold (98%) or phased approach?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS C2, 2025-12-19):

- **Round 1 Decision**: E - "85% for Phase 4, then 98% for Phase 5+"

**Q**: When exactly does mutation score requirement change from 85% to 98%?

**A** (Source: CLARIFY-QUIZME2 C2.1, 2025-12-19):

- **Round 2 Decision**: A - At start of Phase 5 (all packages must reach 98% before Phase 5 work begins)

**Implementation**:

- Phase 4 goal: ≥85% efficacy per package (validate mutation testing infrastructure)
- Phase 5+ goal: ≥98% efficacy per package (mature codebase quality)
- **Transition**: All packages must reach 98% before Phase 5 starts (hard gate)
- Track progress in docs/GREMLINS-TRACKING.md

---

### Test Execution Timing Targets

**Q**: Should we set strict timing targets for test execution?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS C3, 2025-12-19):

- **Round 1 Decision**: E - "<15 seconds per unit test package, <180 seconds for full unit test suite (integration/e2e excluded from strict timing)"

**Q**: How should we enforce test timing targets in CI/CD?

**A** (Source: CLARIFY-QUIZME2 C3.1, 2025-12-19):

- **Round 2 Decision**: B - Warning only (log slow packages but don't fail build)

**Q**: Should we set any timing targets for integration/e2e tests?

**A** (Source: CLARIFY-QUIZME2 C3.2, 2025-12-19):

- **Round 2 Decision**: E - "<25sec integration per-package, <180s all integrations, <240sec all e2e"

**Combined Implementation**:

- **Unit tests**: <15s per package (warning), <180s total (hard limit)
- **Integration tests**: <25s per package (warning), <180s total (warning)
- **E2E tests**: <240s total (warning)
- CI/CD logs slow packages but doesn't fail build
- Use probabilistic execution (TestProbTenth, TestProbQuarter) for packages approaching limits

---

### Test Consolidation Strategy

**Q**: Should we consolidate redundant test cases to improve timing?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q1.1, 2025-12-19):

- **Round 1 Decision**: C - "Yes, consolidate table-driven test cases where truly redundant (same code path, different data)"

**Q**: How do we ensure consolidation doesn't reduce coverage?

**A** (Source: CLARIFY-QUIZME2 Q1.1.1, 2025-12-19):

- **Round 2 Decision**: B - Allow minor coverage drops (<1%) if timing improves significantly (>5s faster)

**Implementation**:

- Consolidate only truly redundant tests (same code path, different data values)
- Validate coverage before/after consolidation
- Allow <1% coverage drop if timing improves >5s
- Use mutation testing to validate quality unchanged
- Document consolidations in commit messages

---

### TestMain Pattern for Server Sharing

**Q**: Should we use TestMain to start servers once per package?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q1.2, 2025-12-19):

- **Round 1 Decision**: A - "Yes, TestMain for heavyweight dependencies (PostgreSQL, servers) - start once per package"

**Q**: How should TestMain handle packages testing multiple services?

**A** (Source: CLARIFY-QUIZME2 Q1.2.1, 2025-12-19):

- **Round 2 Decision**: C for integration tests using TestMain, B for e2e tests using docker compose

**Combined Implementation**:

- **Unit tests**: No TestMain (fast, isolated)
- **Integration tests**: TestMain starts single service + dependencies (PostgreSQL, etc.)
  - Split tests into separate packages (one package per service)
  - Each package's TestMain starts only required service
- **E2E tests**: Docker Compose starts full multi-service stack
  - Use TestMain to wait for stack readiness
  - Tests use docker compose network for service communication

---

### Real Dependencies vs Mocks

**Q**: Should we prefer real dependencies (PostgreSQL containers, crypto) or mocks?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS O1, Q2.1, 2025-12-19):

- **Round 1 Decision**: C - "ALWAYS prefer real dependencies (test containers, real crypto, real HTTP servers) over mocks"

**Q**: At what coverage threshold should we add httptest mocks for corner cases?

**A** (Source: CLARIFY-QUIZME2 Q2.1.1, 2025-12-19):

- **Round 2 Decision**: C and D
  - C - Use httptest for error injection (network failures, timeout simulation)
  - D - Use httptest for security testing (malformed requests, boundary conditions)

**Combined Implementation**:

- **ALWAYS prefer real dependencies**: PostgreSQL containers, real HTTPS servers, real crypto
- **Mocks ONLY for**:
  - Hard-to-reach error injection (network failures, timeouts)
  - Security testing (malformed requests, boundary conditions)
  - External services that can't run locally (email, SMS, cloud-only APIs)
- **Rationale**: Real dependencies reveal production bugs; mocks hide integration issues

---

### main() Function Testability Pattern

**Q**: How should we structure main() functions for testability?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS O2, 2025-12-19):

- **Decision**: E - "Thin main() delegates to co-located internalMain(args, stdin, stdout, stderr) - fully testable"

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

**Rationale**:

- main() is untestable (os.Args, os.Stdout, os.Exit hardcoded)
- internalMain() accepts injected dependencies (testable with mocks)
- 95%+ coverage achievable (main() 0% is acceptable when internalMain() 95%+)
- Enables testing all branches (happy path, error cases, exit codes)

---

## Cryptography and Hash Service

### Hash Service Version Architecture

**Q**: What does "version" mean in hash service architecture?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS O4, Q5.1, 2025-12-19):

- **Round 1 Decision**: E - "Version = date-based policy revision (v1=2020 NIST, v2=2023 NIST, v3=2025 OWASP)"

**Q**: What triggers hash version updates in production?

**A** (Source: CLARIFY-QUIZME2 Q5.1.1, 2025-12-19):

- **Round 2 Decision**: A - Manual operator decision (update config, restart service, new hashes use v2)

**Combined Architecture**:

- **Version = Date-Based Policy Revision**: v1 (2020 NIST), v2 (2023 NIST), v3 (2025 OWASP)
- **Algorithm Selection Within Version**: Input size-based (0-31→SHA-256, 32-47→SHA-384, 48+→SHA-512)
- **4 Registries × 3 Versions = 12 Configurations**: Each registry supports v1/v2/v3
- **Version Update Trigger**: Manual operator decision (update config, restart, new hashes use new version)
- **Gradual Migration**: Old hashes stay on v1, new hashes use v2

---

### Hash Output Format

**Q**: What format should hash outputs use?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q5.2, 2025-12-19):

- **Round 1 Decision**: A - "Prefix format: {v}:base64_hash"

**Q**: How do we verify old hashes during version migration?

**A** (Source: CLARIFY-QUIZME2 Q5.1.2, 2025-12-19):

- **Round 2 Decision**: B - "Hash output includes version prefix, algorithm, and algorithm parameters"

**Q**: How do we handle existing hashes without version prefix?

**A** (Source: CLARIFY-QUIZME2 Q5.2.1, 2025-12-19):

- **Round 2 Decision**: D - Reject unprefixed hashes (force re-hash on next authentication)

**Combined Implementation**:

- **Output Format**: `{version}:{algorithm}:{params}:base64_hash`
  - Example: `{1}:PBKDF2-HMAC-SHA256:rounds=600000:abcd1234...`
  - Example: `{2}:PBKDF2-HMAC-SHA384:rounds=600000:efgh5678...`
- **Verification**: Parse prefix, use specified algorithm/params
- **Backward Compatibility**: None - reject unprefixed hashes, force re-hash on next auth
- **Rationale**: Enforces version metadata, prevents ambiguous verification attempts

---

## Observability and Telemetry

### Diagnostic Logging for Startup Phases

**Q**: Should we add diagnostic logging for /readyz timeout issues?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q3.2, 2025-12-19):

- **Round 1 Decision**: D - "Yes, add diagnostic logging (timestamps, phase names like TLS/DB/unseal)"

**Q**: What level of diagnostic detail should we add?

**A** (Source: CLARIFY-QUIZME2 Q3.2.1, 2025-12-19):

- **Round 2 Decision**: B - Timestamps + component names (2025-12-19 10:15:30 [TLS] Server started)

**Combined Implementation**:

- **Format**: `YYYY-MM-DD HH:MM:SS [COMPONENT] Message`
- **Components**: TLS, Database, Migration, Unseal, Health
- **Example**:

  ```
  2025-12-19 10:15:28 [TLS] Generating server certificate...
  2025-12-19 10:15:30 [TLS] Server certificate ready (2.1s)
  2025-12-19 10:15:31 [Database] Connecting to PostgreSQL...
  2025-12-19 10:15:36 [Database] Connection established (5.3s)
  2025-12-19 10:15:37 [Migration] Running schema migrations...
  2025-12-19 10:15:38 [Migration] Migrations complete (1.2s)
  2025-12-19 10:15:39 [Unseal] Deriving unseal keys...
  2025-12-19 10:15:41 [Unseal] Service unsealed (1.8s)
  2025-12-19 10:15:41 [Health] Service ready (total 13.2s)
  ```

---

### Otel Collector Health Check

**Q**: Should we use sidecar health check for otel-collector-contrib?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q3.3, 2025-12-19):

- **Round 1 Decision**: "D IS ONLY SOLUTION that works"

**Q**: Should we investigate otel-collector internal health endpoints?

**A** (Source: CLARIFY-QUIZME2 Q3.3.1, 2025-12-19):

- **Round 2 Decision**: B - Yes, but keep sidecar as fallback (belt-and-suspenders approach)

**Combined Implementation**:

- **Primary**: Alpine sidecar container with wget healthcheck (working solution)
- **Investigation**: Check otel-collector /healthz or /metrics endpoints
- **Fallback**: Keep sidecar if internal endpoints don't exist or unreliable
- **Documentation**: Document findings in DETAILED.md

---

## Deployment and Docker

### Docker Compose Health Check Start Period

**Q**: Should we increase health check start_period based on diagnostic logging results?

**A** (Source: CLARIFY-QUIZME2, 2025-12-19):

- **Decision**: D - Dynamic per service (KMS: 30s, Identity: 45s, CA: 60s)

**Per-Service Start Periods**:

- **KMS**: 30s (TLS + unseal + basic startup)
- **Identity**: 45s (TLS + DB + migrations + unseal)
- **CA**: 60s (TLS + DB + migrations + certificate chain validation + unseal)
- **JOSE**: 30s (TLS + unseal + basic startup)
- **Telemetry (OTLP)**: 45s (waiting for services + initialization)

**Rationale**: Services have different startup complexities - CA validates full certificate chains, Identity runs migrations, KMS/JOSE are simpler.

---

## CI/CD and Automation

### Gremlins Windows Compatibility

**Q**: Should we investigate gremlins Windows compatibility or use CI/CD workaround?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A6, 2025-12-19):

- **Round 1 Decision**: E - "Investigate in Phase 4, but use CI/CD workaround for now"

**Q**: What level of effort should we invest in Windows compatibility?

**A** (Source: CLARIFY-QUIZME2, 2025-12-19):

- **Round 2 Decision**: B - Medium priority - Investigate in Phase 4, document findings, keep CI/CD workaround
- **User Emphasis**: "MUST: TIME ALWAYS PERMITS!!! BIAS MUST ALWAYS PRIORITIZE ACCURACY OVER COMPLETION!!!"

**Combined Implementation**:

- **Current**: CI/CD runs gremlins on Linux (working)
- **Phase 4**: Dedicate time to investigate Windows compatibility
  - Test gremlins v0.7.0+ for fixes
  - Analyze root cause of Windows panics
  - Document findings in docs/GREMLINS-TRACKING.md
- **Fallback**: Keep CI/CD workaround if Windows compatibility not achievable
- **Priority**: Medium (important for developer experience, but not blocking)

---

### Coverage Baseline Artifact Retention

**Q**: Where should we store coverage baseline artifacts?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q8.2, 2025-12-19):

- **Round 1 Decision**: B - "CI/CD workflow artifacts (download before/after for comparison)"

**Q**: How long should we retain coverage baseline artifacts?

**A** (Source: CLARIFY-QUIZME2, 2025-12-19):

- **Round 2 Decision**: E - "1 day - this is non-released product, so no need to retain any longer than 1 day"

**Combined Implementation**:

- **Storage**: CI/CD workflow artifacts (GitHub Actions)
- **Retention**: 1 day (non-released product, rapid iteration)
- **Usage**: Download artifacts for trend analysis during active development
- **Exception**: Tag specific baseline artifacts for release milestones (download and commit to git)

---

## Documentation and Workflow

### Clarify.md Organization Strategy

**Q**: How should we organize clarify.md as it grows with Round 2 answers?

**A** (Source: CLARIFY-QUIZME2, 2025-12-19):

- **Decision**: B - Topical reorganization (merge Round 1 + Round 2 into unified topic sections)

**Implementation**:

- **This document** uses topical organization
- Merged Round 1 (SPECKIT-CONFLICTS-ANALYSIS) and Round 2 (CLARIFY-QUIZME2) answers
- Organized by functional area (Architecture, Testing, Cryptography, etc.)
- Sub-questions grouped under parent topic
- Combined Q&A shows evolution from Round 1 → Round 2

**Deleted Files After Merge**:

- specs/002-cryptoutil/CLARIFY-QUIZME.md (merged)
- specs/002-cryptoutil/CLARIFY-QUIZME2.md (merged)

---

### Spec Kit Feedback Loop Timing

**Q**: How frequently should we update constitution/spec during implementation?

**A** (Source: CLARIFY-QUIZME2, 2025-12-19):

- **Decision**: C - When implementation insights contradict spec (as-needed basis)

**Implementation**:

- Update constitution/spec when implementation reveals design flaws
- Don't wait for phase completion if contradiction discovered
- Document feedback loop in DETAILED.md Section 2 timeline
- Minor clarifications can accumulate for batch update
- Major contradictions require immediate spec update

---

### Service Template Initialization

**Q**: How should services initialize with template?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q6.1, 2025-12-19):

- **Round 1 Decision**: A - "Constructor injection (NewService(handlers, middleware, config))"

**Q**: Should service template use builder pattern or direct constructor?

**A** (Source: CLARIFY-QUIZME2 Q6.1.1, 2025-12-19):

- **Round 2 Decision**: D - Configuration struct (NewServiceTemplate(&ServiceConfig{Handlers: ..., Middleware: ...}))

**Combined Implementation**:

```go
type ServiceConfig struct {
    Handlers    []Handler
    Middleware  []Middleware
    Telemetry   TelemetryConfig
    Security    SecurityConfig
    Database    DatabaseConfig
}

func NewServiceTemplate(config *ServiceConfig) (*Service, error) {
    // Validate config
    // Initialize components
    // Return configured service
}
```

**Rationale**: Configuration struct provides clearest initialization semantics, easier to extend, and best IDE support for field names.

---

### Service Template Customization Points

**Q**: Which customization points should template expose?

**A** (Source: CLARIFY-QUIZME2 Q6.1.2, 2025-12-19):

- **Decision**: D - All aspects customizable (fully parameterized template)

**Customization Points**:

- **Handlers**: Service-specific API handlers
- **Middleware**: Custom middleware chain (CORS, CSRF, auth, rate limiting)
- **Telemetry**: OTLP config, service name, trace sampling
- **Security**: TLS config, auth methods, IP allowlist
- **Database**: Connection string, pool size, migrations path

**Rationale**: Maximum flexibility for 8 different PRODUCT-SERVICE instances (sm-kms, pki-ca, jose-ja, identity-*).

---

### SDK Generation Automation Timing

**Q**: When should go:generate run for SDK generation?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS Q6.3, 2025-12-19):

- **Round 1 Decision**: "B (go:generate), but user/LLM agent can do A (oapi-codegen directly) during development"

**Q**: When should go:generate run for SDK generation?

**A** (Source: CLARIFY-QUIZME2 Q6.3.1, 2025-12-19):

- **Round 2 Decision**: A - Manually during development (developer runs `go generate ./...` when OpenAPI changes)

**Combined Implementation**:

- **Manual execution**: Developers run `go generate ./...` when OpenAPI changes
- **Pre-commit hook**: Optional (can validate SDK up-to-date)
- **CI/CD**: NOT automatic (avoids commit pollution)
- **Flexibility**: Developers choose `go generate ./...` or direct `oapi-codegen` for faster iteration

---

## Windows Firewall Prevention

**Q**: How should we prevent Windows Firewall exception prompts during testing?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS O3, 2025-12-19):

- **Decision**: E - "MANDATORY: ALWAYS bind to 127.0.0.1 (NEVER 0.0.0.0) in tests/local dev"

**Implementation**:

- **Local tests**: ALWAYS use `127.0.0.1` binding (no firewall prompts)
- **Docker containers**: Use `0.0.0.0` binding (isolated network)
- **Integration tests**: Use hardcoded ports on `127.0.0.1` (18080, 18081, 18082)
- **Test configs**: Verify `bind_address: 127.0.0.1` in all test YAML files

**Rationale**:

- Binding to `0.0.0.0` (all interfaces) triggers Windows Firewall prompts
- Binding to `127.0.0.1` (loopback only) does NOT trigger prompts
- Tests MUST run without user interaction

---

*This clarify.md document is authoritative for implementation decisions. When ambiguities arise, refer to this document first before making assumptions.*
