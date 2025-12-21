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

**Q**: Which specific packages require 95% vs 98% coverage?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A2, CLARIFY-QUIZME2 A2.1, 2025-12-19):

**Answer**: D - Case-by-case per package (document each in clarify.md)

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

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A4, constitution.md VA, spec.md, 2025-12-20):

**Service Discovery Mechanisms**:

1. **Configuration File** (Preferred): Static YAML with explicit URLs
2. **Docker Compose**: Service names resolve via Docker network DNS
3. **Kubernetes**: Service discovery via cluster DNS
4. **Environment Variables**: Override config file settings

**Configuration Pattern**:

```yaml
# Example: KMS federation configuration
federation:
  identity_url: "https://identity-authz:8180"
  identity_enabled: true
  identity_timeout: 10s

  jose_url: "https://jose-server:8280"
  jose_enabled: true
  jose_timeout: 10s

  ca_url: "https://ca-server:8380"
  ca_enabled: false  # Optional
  ca_timeout: 10s
```

**Q**: Where should federation configuration be stored?

**A** (Source: CLARIFY-QUIZME2 A4.1, constitution.md VA, 2025-12-20):

- **Decision**: Each service has own federation section in service-specific YAML
- Example: `kms.yml` has `federation.identity_url`, `federation.jose_url`
- Rationale: Decouples services, allows independent configuration

**Q**: How should services handle federated service unavailability?

**A** (Source: CLARIFY-QUIZME2 A4.2, constitution.md VA, spec.md, 2025-12-20):

- **Decision**: Graceful degradation with circuit breaker patterns

**Fallback Modes**:

- **Identity Unavailable**: `local_validation` (cached keys), `reject_all` (strict), `allow_all` (dev only)
- **JOSE Unavailable**: `internal_crypto` (use service's own JWE/JWS)
- **CA Unavailable**: `self_signed` (dev), `cached_certs` (production)

**Circuit Breaker**:

- Open circuit after N consecutive failures (default: 5)
- Reset circuit after timeout (default: 60s)
- Test N requests before closing (default: 3)

**Retry Strategies**:

- Exponential backoff: 1s, 2s, 4s, 8s, 16s (max 5 retries)
- Timeout escalation: 1.5x per retry (10s → 15s → 22.5s)
- Health check before retry: Poll `/admin/v1/healthz`

**Combined Implementation**:

- Each service YAML has `federation:` section with service URLs and timeouts
- Each service YAML has `federation_fallback:` section with fallback modes
- Services start even if federated services unreachable
- Federated features disabled until dependencies available
- Log warnings for unavailable federated services
- Periodic retry with exponential backoff
- Circuit breaker prevents cascade failures
- Health monitoring tracks federated service availability

**Testing Requirements** (Source: constitution.md VA, spec.md):

- Integration tests: Mock federated services, test graceful degradation, test circuit breaker
- E2E tests: Deploy full stack, test cross-service communication, verify health checks

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

**Q**: Should we enforce strict coverage targets (95%/98%) or allow gradual improvement?

**A** (Source: SPECKIT-CONFLICTS-ANALYSIS A2, 2025-12-19):

- **Strict Enforcement**: 95%+ production, 98%+ infrastructure/utility, NO EXCEPTIONS
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

### Service Template Extraction and Reuse

**Q**: Why must service template be extracted and reused instead of copying code?

**A** (Source: constitution.md Section IX, spec.md Phase 6, architecture instructions, 2025-12-21):

**Requirements**:

- **Phase 6 MUST extract reusable template** from KMS reference implementation
- **ALL new services MUST use template** (consistency, reduced code duplication)
- **Learn-PS MUST demonstrate template reusability** (Phase 7 validation)
- **Template success criteria**: Service implementation <500 lines

**Template Components**:

- Dual HTTPS servers (public + admin)
- Health check endpoints (/livez, /readyz, /healthz)
- Graceful shutdown with context cancellation
- Middleware pipeline (CORS, CSRF, CSP, rate limiting, auth)
- Database abstraction (PostgreSQL + SQLite)
- OpenTelemetry integration (traces, metrics, logs)
- TLS configuration (separate public/admin)
- Config management (YAML + CLI flags + Docker secrets)

**NEVER DO**:

- ❌ Copy-paste service infrastructure code between services
- ❌ Duplicate dual-server pattern, health checks, shutdown logic
- ❌ Reimplement middleware pipeline or telemetry integration

**ALWAYS DO**:

- ✅ Extract template from proven KMS implementation
- ✅ Parameterize template for service-specific customization
- ✅ Use constructor injection for handlers, middleware, config
- ✅ Separate business logic from infrastructure concerns

**Rationale**:

- **Consistency**: All services follow same architectural patterns
- **Maintainability**: Infrastructure fixes propagate to all services
- **Quality**: Proven patterns reduce bugs in new services
- **Velocity**: Template reduces implementation from ~1500 lines to <500 lines

**Reference**:

- See constitution.md Section IX "Service Template Requirement"
- See spec.md Section "Phase 6: Service Template Extraction"
- See architecture instructions "Service Template Requirement - MANDATORY"

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

## Service Architecture Continued

### Q6: Are the testing coverage targets (95% production, 98% infrastructure/utility) consistently specified?

**Answer**: A - Yes, consistently specified

- Production packages: ≥95% (internal/{jose,identity,kms,ca})
- Infrastructure packages: ≥98% (internal/cmd/cicd/*)
- Utility packages: ≥98% (internal/shared/*, pkg/*)
- Main functions: 0% acceptable if internalMain() ≥95%
- **Documented in**: .github/instructions/01-04.testing.instructions.md

### Q7: Are the test timing requirements (<15s unit, <180s total) realistic?

**Answer**: A - Yes, realistic

- Per-package timeout: <15 seconds for unit tests
- Full suite timeout: <180 seconds (3 minutes) for all unit tests
- Integration/E2E excluded from strict timing (Docker startup overhead acceptable)
- Probabilistic execution MANDATORY for packages approaching 15s limit
- **Documented in**: .github/instructions/01-04.testing.instructions.md

### Q8: Are FIPS 140-3 restrictions and approved algorithms clearly documented?

**Answer**: A - Yes, clearly documented

- FIPS 140-3 mode ALWAYS enabled (MANDATORY, NEVER disabled)
- Approved algorithms: RSA ≥2048, AES ≥128, EC NIST curves, EdDSA, PBKDF2-HMAC-SHA256
- BANNED algorithms: bcrypt, scrypt, Argon2, MD5, SHA-1
- Password hashing: PBKDF2-HMAC-SHA256 ONLY
- Algorithm agility: All operations support configurable algorithms with FIPS defaults
- **Documented in**: .github/instructions/01-09.cryptography.instructions.md

### Q9: Are TLS configuration patterns (TLS 1.3+, no InsecureSkipVerify) adequately specified?

**Answer**: A - Yes, adequately specified

- MinVersion: TLS 1.3+ (MANDATORY)
- Full cert chain validation (MANDATORY)
- NEVER InsecureSkipVerify (ABSOLUTE PROHIBITION)
- Separate TLS config for public/admin endpoints
- Client cert authentication configurable per endpoint
- **Documented in**: .github/instructions/01-07.security.instructions.md, .github/instructions/01-10.pki.instructions.md

### Q10: Are phase dependencies (Phase 1 → Phase 2 → Phase 3) logically ordered?

**Answer**: A - Yes, logically ordered

- Phase 1 Foundation: Domain models, schema, CRUD (≥95% coverage, ≥80% mutation)
- Phase 2 Core: Business logic, APIs, auth (E2E works, zero CRITICAL TODOs)
- Phase 3 Advanced: MFA, WebAuthn (ONLY after Phase 1+2 complete)
- Strict sequence enforcement: NEVER start Phase 3 before Phase 2 complete
- **Documented in**: .github/instructions/05-01.evidence-based-completion.instructions.md

### Q11: Is cross-database compatibility (PostgreSQL + SQLite) adequately specified?

**Answer**: A - Yes, adequately specified

- UUID type: TEXT (not native UUID) for cross-DB compatibility
- JSON fields: GORM `serializer:json` (not `type:json`)
- Nullable UUIDs: NullableUUID type (not pointer `*googleUuid.UUID`)
- SQLite connection pool: MaxOpenConns=5 for GORM transactions
- WAL mode + busy_timeout for concurrent writes
- **Documented in**: .github/instructions/01-06.database.instructions.md

### Q12: Is naming consistency (camelCase cryptoutil prefix for imports) enforced?

**Answer**: A - Yes, enforced

- Pattern: `cryptoutil<PackageName>` (e.g., cryptoutilMagic, cryptoutilCmdCicdCommon)
- Defined in .golangci.yml importas section (source of truth)
- Common third-party: googleUuid, joseJwa/Jwe/Jwk/Jws, `crand "crypto/rand"`
- Crypto acronyms: ALL CAPS (RSA, EC, ECDSA, ECDH, HMAC, AES, JWA, JWK, JWS, JWE, ED25519, PKCS8, PEM, DER)
- **Documented in**: .github/instructions/01-05.golang.instructions.md

### Q13: Are success criteria for service template extraction (<500 lines) clear?

**Answer**: A - Yes, clear

- Service implementation <500 lines after template extraction
- Template components: Dual servers, middleware, DB abstraction, telemetry, health checks, shutdown, config, TLS
- Phase 6: Extract from KMS reference
- Phase 7: Validate with Learn-PS
- ALL new services MUST use template
- **Documented in**: .github/instructions/01-01.architecture.instructions.md

### Q14: Are health check patterns (/livez, /readyz, /healthz) adequately specified?

**Answer**: A - Yes, adequately specified

- /admin/v1/livez: Liveness probe (service running)
- /admin/v1/readyz: Readiness probe (service ready to accept traffic)
- /admin/v1/healthz: Health check (service healthy)
- /admin/v1/shutdown: Graceful shutdown trigger
- All on admin endpoint (127.0.0.1:9090)
- Consumed by Docker health checks, Kubernetes probes, monitoring systems
- **Documented in**: .github/instructions/01-01.architecture.instructions.md

### Q15: Is log aggregation (OTLP → otel-collector → Grafana) clearly specified?

**Answer**: A - Yes, clearly specified

- All telemetry forwarded through otel-contrib sidecar (MANDATORY)
- Application: OTLP gRPC:4317 or HTTP:4318 → otel-collector → Grafana OTLP:14317/14318
- Collector self-monitoring: Internal → Grafana OTLP HTTP:14318
- Grafana receives telemetry only (no Prometheus scraping of collector metrics)
- **Documented in**: .github/instructions/02-03.observability.instructions.md

### Q16: Are SPOFs (Single Points of Failure) in architecture adequately mitigated?

**Answer**: A - Yes, adequately mitigated

- PostgreSQL: Multiple instances (cryptoutil-postgres-1, cryptoutil-postgres-2) share database
- SQLite: In-memory, acceptable for dev/test (not production)
- Otel-collector: Sidecar pattern (one per deployment, not globally shared)
- Graceful degradation: Circuit breaker, fallback modes for federated services
- Health monitoring: Regular checks, metrics/alerts
- **Documented in**: .github/instructions/01-01.architecture.instructions.md, .github/instructions/02-03.observability.instructions.md

### Q17: Are performance scaling considerations (horizontal scaling) adequately addressed?

**Answer**: B - Partially addressed, needs horizontal scaling guidance

- **Current Coverage**: Vertical scaling (resource limits, connection pools, concurrent requests)
- **Missing**: Horizontal scaling patterns (load balancing, session affinity, distributed caching, database sharding)
- **Required Updates**:
  - Add load balancing patterns to constitution.md
  - Add session state management for horizontal scaling to spec.md
  - Add distributed caching strategy to spec.md
  - Add database scaling patterns (read replicas, sharding) to spec.md
- **Action Required**: Update constitution.md Section X (Performance & Scaling) with horizontal scaling guidance

### Q18: Are backup and recovery procedures adequately specified?

**Answer**: A - Yes, covered by database migrations, key versioning, and rotation

- Database migrations: Embedded SQL with golang-migrate (versioned schema changes)
- Key rotation: Version-based key management (KeyRing pattern with activeKeyID)
- Disaster recovery: Restore from migrations + key backups
- PostgreSQL: Standard pg_dump/pg_restore patterns
- SQLite: File-based backups
- **Documented in**: .github/instructions/01-06.database.instructions.md (migrations), .github/instructions/01-09.cryptography.instructions.md (key rotation)

### Q19: Is integration testing (Docker Compose E2E) comprehensive?

**Answer**: A - Yes, comprehensive

- Full stack deployment: All services + dependencies (PostgreSQL, otel-collector, Grafana)
- Health check validation: Service startup, readiness checks
- Cross-service communication: Federation patterns, service discovery
- Failure scenarios: Graceful degradation, circuit breaker, retry logic
- **Documented in**: .github/instructions/01-01.architecture.instructions.md (federation testing), .github/instructions/02-01.github.instructions.md (ci-e2e workflow)

### Q20: Is documentation maintenance strategy (continuous updates, DETAILED.md timeline) clear?

**Answer**: A - Yes, clear

- Constitution/spec/clarify: Continuous updates during implementation (MANDATORY)
- DETAILED.md Section 2: Append-only timeline (authoritative implementation log)
- Mini-cycle feedback: Update specs every 2-3 tasks (not end of phase)
- Session documentation: Append to DETAILED.md (NEVER create standalone session docs)
- **Documented in**: .github/instructions/06-01.speckit.instructions.md

---

## Identity Service Architecture

### Q1.3: Are identity-rp and identity-spa services optional or mandatory?

**Answer**: B+C - Optional services (not all deployments need them), Docker Compose includes all services

- **identity-rp** (Relying Party): Backend-for-Frontend pattern, optional reference implementation
- **identity-spa** (SPA): Static hosting for single-page apps, optional reference implementation
- **Core services**: identity-authz (OAuth 2.1 server), identity-idp (OIDC authentication) are MANDATORY
- **Docker Compose**: Includes all 5 Identity services for complete reference architecture
- **Production**: Deployments may include only authz+idp, omitting rp+spa if using alternative client patterns
- **Update Required**: Constitution.md Section III.4 clarify optional vs mandatory services

### Q1.4: Is learn-ps (Pet Store) intended for production use or dev/test only?

**Answer**: B+B - Dev/test environment only, isolated stack deployment

- **Purpose**: Educational service demonstrating service template usage (Phase 7)
- **NOT for production**: Reference implementation only, demonstrates patterns
- **Deployment**: Isolated stack (not included in production compose files)
- **Docker Compose**: Separate compose file for learn-ps demonstration
- **Update Required**: Constitution.md Section III.5 clarify learn-ps as dev/test reference only

---

## Authentication and Authorization

### Q2.1: Are all authentication methods (password, MFA, WebAuthn, SSO) required?

**Answer**: D+C - All authentication methods mandatory, MFA enrollment tiered by user risk

- **All authentication methods MANDATORY**: Password (PBKDF2), MFA (TOTP/SMS), WebAuthn (FIDO2), SSO (OIDC federation)
- **MFA enrollment**: Tiered based on user risk level (high-risk mandatory, medium-risk recommended, low-risk optional)
- **Rationale**: Comprehensive auth suite supports diverse deployment scenarios
- **Update Required**: Spec.md Section 4.2 clarify MFA enrollment tiers

### Q2.2: Is there a default authentication method fallback?

**Answer**: D+A - No default authentication method, first-match priority ordering

- **No default fallback**: Configuration MUST specify authentication methods explicitly
- **Priority ordering**: First matching method in configuration is used
- **Configuration example**:

```yaml
authentication:
  methods:
    - webauthn  # Highest priority
    - mfa       # Second priority
    - password  # Lowest priority
  require_mfa: true  # Force MFA after password auth
```

- **Rationale**: Explicit configuration prevents security misconfigurations
- **Update Required**: Spec.md Section 4.2 clarify authentication method ordering

### Q2.3: How is session state managed (database, Redis, in-memory)?

**Answer**: D+C - Session storage configurable (database/Redis/in-memory), database storage is default

- **Session format**: Configurable (JWT, opaque tokens, hybrid)
- **Storage backend**: Database (default), Redis (high-performance), in-memory (dev/test only)
- **Configuration example**:

```yaml
session:
  format: jwt  # or opaque, hybrid
  storage: database  # or redis, memory
  duration: 3600  # seconds
```

- **Rationale**: Different deployment scenarios require different session strategies
- **Update Required**: Spec.md Section 4.3 clarify session storage options

---

## Database Architecture

### Q3.1: Does the architecture support active-active database clustering?

**Answer**: E+E - Active-active PostgreSQL cluster pattern supported, no automatic database failover (manual intervention required)

- **PostgreSQL**: Multiple instances (cryptoutil-postgres-1, cryptoutil-postgres-2) share same database
- **NOT active-active cluster**: Instances connect to same PostgreSQL server, not distributed cluster
- **Schema initialization**: First instance initializes schema, others wait via health checks
- **Failover**: Manual intervention required (update database DSN in configuration)
- **Rationale**: Simplifies deployment, avoids distributed consensus complexity
- **Update Required**: Constitution.md Section V.3 clarify active-active vs shared database pattern

### Q3.2: Is feature parity required between SQLite and PostgreSQL?

**Answer**: A+A - Strict feature parity required, except SQLite connection pool differences

- **Feature parity**: ALL business logic, migrations, queries MUST work on both
- **Exceptions**: SQLite MaxOpenConns=5 (vs PostgreSQL defaults), SQLite WAL mode specific
- **Cross-DB compatibility**: TEXT type for UUID, GORM serializer:json, no read-only transactions
- **Testing**: MANDATORY tests on both SQLite (unit) and PostgreSQL (integration)
- **Update Required**: Constitution.md Section V.3 clarify strict parity requirement

### Q3.3: Do all services share a single database or have independent databases?

**Answer**: B+D - Independent databases per service, sequential startup with health check dependencies

- **Database isolation**: Each service has independent database (kms_db, identity_db, jose_db, ca_db)
- **Schema ownership**: Service owns schema, migrations, data
- **Sequential startup**: First instance initializes schema, others wait via `depends_on: service_healthy`
- **Rationale**: Microservices isolation, independent scaling, schema evolution
- **Update Required**: Constitution.md Section V.3 clarify per-service database isolation

---

## Cryptography and FIPS Compliance

### Q4.1: Is FIPS 140-3 compliance strictly enforced or aspirational?

**Answer**: C+C - Aspirational goal pending Go standard library FIPS-validated crypto, document current status

- **Current Status**: Algorithm-level compliance (use FIPS-approved algorithms only)
- **NOT FIPS-validated**: Go crypto libraries not FIPS 140-3 validated (no official CMVP certificate)
- **Compliance Strategy**: Use FIPS-approved algorithms (RSA, AES, ECDSA, PBKDF2) while waiting for Go FIPS validation
- **Future**: Adopt FIPS-validated Go crypto when available (e.g., google/go-fips, microsoft/go-crypto-openssl)
- **Documentation**: Clearly state \"algorithm-level FIPS compliance, not FIPS-validated\"
- **Update Required**: Constitution.md Section VI clarify aspirational vs strict FIPS compliance

### Q4.2: Can different unseal key versions coexist or must they be synchronized?

**Answer**: B+A - Same product instances only (KMS-to-KMS share keys), fail fast on version mismatch across products

- **Intra-product**: All KMS instances MUST use same unseal secrets (deterministic key derivation)
- **Inter-product**: KMS unseal keys ≠ Identity unseal keys (independent key hierarchies)
- **Version mismatch**: Fail fast with clear error message (prevents data corruption)
- **Rationale**: Cryptographic interoperability requires shared keys within product
- **Update Required**: Constitution.md Section VI clarify unseal key synchronization requirements

### Q4.3: Are hash version updates (v1 → v2) automatic or manual?

**Answer**: A+D - Manual version updates only (operator decision), force re-authentication for migration

- **Version updates**: Manual configuration change (update `current_version: 2` in config.yaml)
- **Migration strategy**: Gradual (new hashes use v2, old hashes verify on v1)
- **NO automatic migration**: Old hashes stay on original version until user re-authenticates
- **Force migration**: Invalidate sessions + force re-authentication to trigger re-hash with new version
- **Rationale**: Controlled migration prevents surprise password hash updates
- **Update Required**: Constitution.md Section VI clarify manual hash version management

---

## Testing and Quality Assurance

### Q5.1: Are coverage/mutation/timing requirements CI/CD-blocking or warnings?

**Answer**: A+C - CI/CD failure on violations (blocking), PR merge gating enforced

- **Coverage**: <95% production, <98% infrastructure/utility = CI/CD failure
- **Mutation**: <85% Phase 4, <98% Phase 5+ = CI/CD failure
- **Timing**: >15s unit tests per-package, >180s total suite = CI/CD failure
- **PR merge**: MUST pass all quality gates before merge to main
- **Rationale**: Enforce quality standards early, prevent technical debt accumulation
- **Update Required**: Constitution.md Section VIII clarify CI/CD blocking enforcement

### Q5.2: Is there a grace period for test timing violations?

**Answer**: A+A - No grace period (immediate CI/CD failure), immediate enforcement on all PRs

- **Timing violations**: Immediate CI/CD failure (no warnings, no grace period)
- **Enforcement**: ALL PRs MUST meet timing requirements
- **Mitigation**: Use probabilistic execution (TestProbTenth, TestProbQuarter) for slow packages
- **Rationale**: Prevent slow test accumulation, maintain fast feedback loop
- **Update Required**: Constitution.md Section VIII clarify zero-tolerance timing enforcement

### Q5.3: Do generated code files have lower coverage requirements?

**Answer**: C+A - Generated code has lower coverage target (80%), integration tests excluded from timing requirements

- **Generated code**: Target 80% coverage (vs 95% production)
- **Rationale**: Generated code (OpenAPI client/server) is less error-prone, testing focuses on integration
- **Integration tests**: Excluded from <15s per-package timing (Docker startup overhead acceptable)
- **E2E tests**: Excluded from <180s total timing (full stack startup overhead acceptable)
- **Update Required**: Constitution.md Section VIII clarify generated code coverage targets

---

## CI/CD and Workflows

### Q6.1: Are dependency update PRs (Dependabot/Renovate) auto-merged or require manual review?

**Answer**: D+A - PR created with test results notification only (no auto-merge), strict gating (must pass all checks)

- **Auto-merge**: DISABLED (too risky for security-critical project)
- **Workflow**: Dependabot/Renovate creates PR → CI/CD runs all checks → human reviews → manual merge
- **Notification**: Test results posted to PR (pass/fail visibility)
- **Gating**: MUST pass coverage, mutation, timing, security scans before merge consideration
- **Rationale**: Manual review required for dependency changes (security, compatibility)
- **Update Required**: Constitution.md Section IX.5 clarify dependency update workflow

### Q6.2: Are health checks tiered (livez → readyz → healthz) or flat?

**Answer**: B+A - Tiered health checks (livez → readyz → healthz), max 60 seconds from cold start to healthy

- **Livez**: Process running (immediate response)
- **Readyz**: Dependencies ready (DB connected, migrations applied)
- **Healthz**: Service fully healthy (can accept traffic)
- **Timing**: Max 60 seconds from container start to healthz=true
- **Configuration**:

```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/admin/v1/livez"]
  start_period: 10s
  interval: 5s
  retries: 5  # Max 10+25=35s to healthy
```

- **Update Required**: Constitution.md Section IX.3 clarify tiered health check pattern

---

## Documentation and Workflow

### Q7.1: Is amendment of constitution.md allowed or is it immutable?

**Answer**: A - Amendment allowed with justification (living document), version tracking not required (DON'T CARE)

- **Constitution is LIVING DOCUMENT**: Continuous updates during implementation (MANDATORY)
- **Amendment process**: Discover constraint → document in DETAILED.md timeline → update constitution.md → commit with reference
- **Version tracking**: Not required (git history provides versioning)
- **Rationale**: Implementation reality requires constitution evolution (not static prerequisite)
- **Update Required**: Constitution.md header clarify \"LIVING DOCUMENT\" status

### Q7.2: Is clarify.md regenerated from scratch or updated continuously?

**Answer**: B+C - Continuous updates (topical organization maintained), hybrid regeneration (merge new Q&A into existing topics)

- **Update pattern**: Append new Q&A to existing topical sections (not chronological)
- **Regeneration**: Periodic reorganization to maintain topical structure
- **NEVER**: Create new clarify-ROUND-N.md files (single clarify.md is source of truth)
- **Hybrid approach**: Add new questions to existing topics, reorganize when topics become unwieldy
- **Update Required**: Speckit instructions clarify continuous clarify.md update pattern

### Q7.3: Is CLARIFY-QUIZME.md continuously updated or one-shot?

**Answer**: C+C - Continuous updates (add questions as unknowns arise), user provides answers for batch update to clarify.md

- **Workflow**: Discover unknown → add to CLARIFY-QUIZME.md → user answers → move to clarify.md → update constitution/spec
- **Continuous**: Questions added throughout implementation (not one-shot at beginning)
- **Batch processing**: User answers multiple questions → agent integrates all into clarify.md in one update
- **NEVER**: Pre-fill answers in CLARIFY-QUIZME.md (violates core principle)
- **Update Required**: Speckit instructions clarify continuous CLARIFY-QUIZME.md workflow

---

## Observability and Telemetry

### Q8.1: What are resource limits for OTLP collector?

**Answer**: B+D - 512Mi memory limit, adaptive sampling based on load

- **Memory limit**: 512Mi (prevents OOM in constrained environments)
- **CPU limit**: 500m (0.5 CPU cores)
- **Sampling strategy**: Adaptive based on throughput (100% at low load, 10% at high load)
- **Configuration example**:

```yaml
processors:
  probabilistic_sampler:
    sampling_percentage: 10  # High load
    hash_seed: 42
```

- **Rationale**: Balance telemetry completeness with resource constraints
- **Update Required**: Observability instructions clarify OTLP resource limits and sampling

---

## Security and Secrets Management

### Q9.1: What are the required permissions for Docker secrets files?

**Answer**: A+E - 400 permissions (r--------), Dockerfile MUST include validation job

- **File permissions**: 400 (read-only for owner) or 440 (read-only for owner+group)
- **Rationale**: Prevent unauthorized access to secrets
- **Dockerfile validation pattern** (from KMS):

```dockerfile
# Validation stage - verify secrets exist with correct permissions
FROM alpine:3.19 AS validator
COPY --from=builder /run/secrets/ /run/secrets/
RUN ls -la /run/secrets/ && \
    test -r /run/secrets/database_url_secret && \
    chmod 440 /run/secrets/*
```

- **Enforcement**: CI/CD workflow validates Dockerfile includes secrets validation job
- **Update Required**: Docker instructions add 440 permissions requirement, Dockerfile validation job pattern

---

## Identity and Multi-Tenancy

### Q10.1: Does federated authentication share sessions across services?

**Answer**: D - No session sharing (each service validates tokens independently), UX implications unknown

- **Session isolation**: Each service (authz, idp, rs, rp, spa) manages own sessions
- **Token validation**: Services validate OAuth 2.1 tokens independently (no session sharing)
- **UX impact**: Unknown (may require multiple logins if sessions not federated)
- **Future consideration**: Implement SSO token exchange for seamless UX
- **Update Required**: Identity architecture spec clarify session isolation pattern

### Q10.2: How is multi-tenant data isolation implemented?

**Answer**: B - Schema-level tenant isolation preferred, table-level isolation acceptable

- **Preferred**: Separate PostgreSQL schemas per tenant (tenant_a.users, tenant_b.users)
- **Acceptable**: Tenant ID column in shared tables with row-level security (RLS)
- **NOT SUPPORTED**: Separate databases per tenant (too many connections)
- **Configuration**:

```yaml
multi_tenancy:
  isolation: schema  # or table
  tenant_id_header: X-Tenant-ID
```

- **Update Required**: Identity architecture spec clarify tenant isolation patterns

### Q10.3: Are custom certificate profiles for different client types supported?

**Answer**: B - Custom certificate profiles allowed (DV, OV, EV)

- **Certificate profiles**: DV (Domain Validation), OV (Organization Validation), EV (Extended Validation)
- **Per-client configuration**: Client can request specific profile based on trust requirements
- **Policy enforcement**: CA policy engine enforces profile constraints
- **Configuration example**:

```yaml
certificate_profiles:
  - name: dv
    validation: domain_only
    validity: 90_days
  - name: ov
    validation: organization
    validity: 397_days
  - name: ev
    validation: extended
    validity: 397_days
```

- **Update Required**: CA architecture spec clarify certificate profile customization

---

*This clarify.md document is authoritative for implementation decisions. When ambiguities arise, refer to this document first before making assumptions.*
