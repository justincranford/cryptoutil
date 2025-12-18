# Technical Implementation Plan - 002-cryptoutil

**Date**: December 17, 2025
**Context**: MVP quality focus with hash refactoring and service template extraction
**Status**: 🎯 Fresh Start - 7 Phase Sequential Execution

---

## Plan Overview

**Goal**: Clean up AI slop from 001-cryptoutil, achieve MVP quality, refactor hash service architecture, extract reusable service template

**Approach**: 7-phase sequential execution with strict quality enforcement

**Timeline**: 180-280 hours work effort (210+ tasks)

**Key Principles**:

- **No Exceptions**: 95%+ production, 100% infra/util coverage mandatory
- **Per-Package Enforcement**: Granular tracking, blocking until targets met
- **CI/CD First**: Fix all workflow failures before proceeding
- **Hash Refactoring**: 4 registry types × 3 versions per type, FIPS 140-3 compliant
- **Template Extraction**: Reusable pattern from SM-KMS for 8 services
- **Learn-PS Validation**: Pet Store demo service proves template works

---

## Execution Mandate

**WORK CONTINUOUSLY until user says "STOP"**:

- Complete task → immediately start next task (no summary, no pause)
- Push changes → immediately continue working (no commit echo)
- Update docs → immediately start next task (no acknowledgment)
- NO stopping to provide summaries or status updates
- NO asking for permission between tasks
- NO pausing after git operations or completions
- NO time or token pressure - work can span hours or days
- Correctness over speed - take time to do it right
- Decompose hard/long tasks into smaller subtasks in DETAILED.md

---

## Service Architecture - Dual HTTPS Endpoint Pattern

**CRITICAL: ALL services MUST implement dual HTTPS endpoints - NO HTTP**

### Dual Endpoint Requirements

Every service implements two HTTPS endpoints:

1. **Public HTTPS Endpoint** (configurable port, default 8080+)
   - Serves business APIs and browser UI
   - TLS 1.3+ required (never HTTP)
   - TWO API path prefixes on SAME OpenAPI spec:
     - **`/service/**`**: OAuth 2.1 client credentials tokens (service-to-service)
     - **`/browser/**`**: OAuth 2.1 authorization code + PKCE tokens (browser-to-service)
   - Middleware enforces authorization boundaries per path

2. **Private HTTPS Endpoint** (always 127.0.0.1:9090)
   - Admin/operations endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
   - Localhost-only binding (127.0.0.1, NOT 0.0.0.0 to avoid Windows Firewall prompts)
   - TLS required (never HTTP)
   - Used by Docker health checks, Kubernetes probes, monitoring

### Service Port Assignments

| Service | Public HTTPS | Private HTTPS | Backend |
|---------|--------------|---------------|---------|
| sm-kms | :8080 | 127.0.0.1:9090 | PostgreSQL/SQLite |
| jose-ja | :8280 | 127.0.0.1:9090 | PostgreSQL/SQLite |
| identity-authz | :8180 | 127.0.0.1:9090 | PostgreSQL/SQLite |
| identity-idp | :8181 | 127.0.0.1:9090 | PostgreSQL/SQLite |
| identity-rs | :8182 | 127.0.0.1:9090 | PostgreSQL/SQLite |
| identity-rp | :8183 | 127.0.0.1:9090 | None (BFF) |
| identity-spa | :8184 | 127.0.0.1:9090 | None (static) |
| pki-ca | :8380 | 127.0.0.1:9090 | PostgreSQL/SQLite |

---

## Phase 1: Optimize Slow Test Packages (Day 1-2, 8-12h)

**Objective**: ALL !integration packages execute in ≤30 seconds (relaxed from 12s after realistic analysis)

**Rationale**: Fast feedback loops essential for development velocity. Target relaxed to ≤30s per package after analyzing crypto-heavy test suites.

### Implementation Strategy

**Baseline** (P1.1):

1. Run `go test -json -v ./... 2>&1 | tee test-output/baseline-timing-002.txt`
2. Parse JSON output, extract per-package execution times
3. Identify all packages >12s
4. Document in test-output/baseline-timing-002-summary.md

**Optimization Pattern** (P1.2-P1.11):

1. Profile test execution hotspots
2. Apply probabilistic execution to algorithm variants:
   - Base algorithms: `TestProbAlways` (100% execution)
   - Important variants: `TestProbQuarter` (25% execution)
   - Less critical variants: `TestProbTenth` (10% execution)
3. Optimize redundant operations (TLS handshakes, certificate generation)
4. Verify coverage maintained (no loss acceptable)
5. Verify ≤12s target achieved

**Target Packages** (preliminary, will expand after baseline):

- internal/jose (algorithm variants, crypto operations)
- internal/jose/server (HTTP handler overhead, middleware)
- internal/kms/client (existing probabilistic needs tuning)
- internal/kms/server/application (barrier operations, unseal)
- internal/identity/authz (OAuth 2.1 flows, token generation)
- internal/identity/authz/clientauth (mTLS handshakes)
- internal/identity/idp (MFA flows, consent/login)
- internal/shared/crypto/keygen (key generation variants)
- internal/shared/crypto/digests (HKDF variants, SHA variants)
- internal/shared/crypto/certificate (TLS handshakes, cert generation)

**Infrastructure** (P1.12-P1.15):

- **P1.12**: Create scripts/monitor-test-timing.ps1 (parse JSON, flag >12s)
- **P1.13**: Document patterns in docs/testing/probabilistic-patterns.md
- **P1.14**: Run full verification (baseline comparison, identify remaining slow)
- **P1.15**: Add pre-commit hook (fail if any package >12s)

**Success Criteria**:

- ✅ ALL !integration packages ≤12 seconds
- ✅ Coverage maintained (no losses from optimization)
- ✅ Pre-commit hook enforces timing limit
- ✅ Monitoring script provides real-time feedback

---

## Phase 2: Coverage Targets - 95% Mandatory (Day 3-5, 48-72h)

**Objective**: Production 95%+, infrastructure/utility 100%, NO EXCEPTIONS

**Rationale**: "95%+ with exceptions" in 001-cryptoutil led to accepting 66.8%, 39%, etc. Strict enforcement prevents technical debt accumulation.

### Coverage Enforcement Rules

**BLOCKING Rules**:

1. Coverage < 95% (production) or < 100% (infra/util) = BLOCKING issue
2. "Improvement" is NOT success - only "target met" counts
3. NO rationalization: "This package is mostly error handling" → Add error path tests
4. Per-package enforcement: Can't proceed to next package until current ≥ target

**Per-Package Workflow**:

1. **Baseline**: `go test ./pkg -coverprofile=./test-output/coverage_pkg_baseline.out`
2. **Gap Analysis**: `go tool cover -html=./test-output/coverage_pkg_baseline.out -o ./test-output/coverage_pkg_baseline.html` → Open HTML, identify RED lines
3. **Test Development**: Write tests ONLY for identified RED lines (targeted, not trial-and-error)
4. **Verification**: Re-run coverage, confirm gaps filled, ≥95% or ≥100% achieved

### Major Areas

**P2.1: KMS Server Coverage (95%+)**

- internal/kms/server/application (baseline, gap analysis, tests, verify)
- internal/kms/server/businesslogic (baseline, gap analysis, tests, verify)
- internal/kms/server/barrier/contentkeysservice (baseline, gap analysis, tests, verify)
- internal/kms/server/barrier/intermediatekeysservice (baseline, gap analysis, tests, verify)
- internal/kms/server/barrier/rootkeysservice (baseline, gap analysis, tests, verify)
- internal/kms/server/barrier/unsealservice (baseline, gap analysis, tests, verify)
- internal/kms/server/repository/orm (baseline, gap analysis, tests, verify)
- internal/kms/server/repository/sqlrepository (baseline, gap analysis, tests, verify)

**P2.2: KMS Client Coverage (95%+)**

- internal/kms/client (baseline, gap analysis, tests, verify)

**P2.3: Identity Server Coverage (95%+)**

**CRITICAL**: internal/identity/authz currently 66.8% (28.2 point gap)

- internal/identity/authz (baseline, gap analysis, tests, verify)
- internal/identity/idp (baseline, gap analysis, tests, verify)
- internal/identity/rs (baseline, gap analysis, tests, verify)
- internal/identity/rp (baseline, gap analysis, tests, verify)
- internal/identity/spa (baseline, gap analysis, tests, verify)
- internal/identity/authz/clientauth (baseline, gap analysis, tests, verify)
- internal/identity/domain (baseline, gap analysis, tests, verify)
- internal/identity/repository (baseline, gap analysis, tests, verify)

**P2.4: JOSE Coverage (95%+)**

- internal/jose (baseline, gap analysis, tests, verify)
- internal/jose/server (baseline, gap analysis, tests, verify)

**P2.5: CA Coverage (95%+)**

- internal/ca/server (baseline, gap analysis, tests, verify)
- internal/ca/domain (baseline, gap analysis, tests, verify)

**P2.6: Shared Crypto Coverage (100% - utility code)**

- internal/shared/crypto/keygen (baseline, gap analysis, tests, verify)
- internal/shared/crypto/digests (baseline, gap analysis, tests, verify)
- internal/shared/crypto/certificate (baseline, gap analysis, tests, verify)
- internal/shared/crypto/signer (baseline, gap analysis, tests, verify)

**P2.7: Shared Infrastructure Coverage (100%)**

- internal/shared/apperr (baseline, gap analysis, tests, verify)
- internal/shared/config (baseline, gap analysis, tests, verify)
- internal/shared/container (baseline, gap analysis, tests, verify)
- internal/shared/magic (baseline, gap analysis, tests, verify)
- internal/shared/pool (baseline, gap analysis, tests, verify)
- internal/shared/telemetry (baseline, gap analysis, tests, verify)
- internal/shared/testutil (baseline, gap analysis, tests, verify)
- internal/shared/util (baseline, gap analysis, tests, verify)

**P2.8: CICD Coverage (100% - infrastructure)**

- internal/cmd/cicd/cicd (baseline, gap analysis, tests, verify)
- internal/cmd/cicd/format_go (baseline, gap analysis, tests, verify)
- internal/cmd/cicd/lint (baseline, gap analysis, tests, verify)
- internal/cmd/cicd/*/all subdirectories (baseline per subdir, gap analysis, tests, verify)

**Success Criteria**:

- ✅ Production packages: ALL ≥95% coverage
- ✅ Infrastructure/utility: ALL ≥100% coverage
- ✅ Per-package verification complete
- ✅ Gap analysis documented for each package

---

## Phase 3: CI/CD Workflow Fixes (Day 6, 6-8h)

**Objective**: 0 workflow failures, all quality gates green

**Rationale**: Accumulated CI/CD debt from 001-cryptoutil (5 failing workflows). Fix all before proceeding to avoid compounding issues.

### Workflow Failures

**P3.1: ci-quality (outdated dependencies)**

**Issue**: github.com/goccy/go-yaml v1.19.0 outdated (v1.19.1 available)

**Fix**:

1. Run `go get -u github.com/goccy/go-yaml@v1.19.1`
2. Run `go mod tidy`
3. Verify tests pass
4. Commit and push
5. Review all direct dependencies: `go list -u -m all | grep '\[.*\]$'`
6. Update incrementally with test validation
7. Add dependabot.yml for automated updates

**P3.2: ci-mutation (45min timeout)**

**Issue**: Mutation testing exceeds 45 minute timeout

**Root Cause**:

- Sequential package execution
- Large packages (e.g., identity/authz) take 10+ minutes alone
- Total time: 45+ minutes (timeout)

**Fix**:

1. Analyze current execution time per package (gremlins logs)
2. Reduce mutation test scope (focus on business logic, exclude test utilities)
3. Parallelize by package using GitHub Actions matrix:
   - Split into groups (4-6 packages per job)
   - Run groups in parallel
   - Aggregate results
4. Set realistic per-job timeout (15 minutes)
5. Total workflow time: ~20 minutes (parallel execution)

**P3.3: ci-fuzz (opentelemetry-collector healthcheck)**

**Issue**: healthcheck-opentelemetry-collector-contrib exit 1

**Root Cause**:

- Healthcheck command incorrect or service not starting
- Port mapping issue or dependency not ready

**Fix**:

1. Review Docker Compose logs for otel-collector-contrib
2. Check service dependencies in compose.integration.yml
3. Verify port mappings (4317 gRPC, 4318 HTTP, 13133 healthcheck)
4. Update healthcheck command if needed
5. Increase start_period if service needs more time
6. Add diagnostic logging to healthcheck
7. Test locally: `docker compose -f deployments/compose/compose.integration.yml up`

**P3.4: ci-dast (/admin/v1/readyz timeout)**

**Issue**: Admin readyz endpoint not ready within timeout

**Root Cause**:

- Service startup slower than expected (database migration, unseal)
- Timeout too short for GitHub Actions environment

**Fix**:

1. Review service startup logs (database migration timing)
2. Check unseal operation completes
3. Profile bottlenecks (migration, unseal, health check)
4. Optimize service startup:
   - Reduce database migration overhead
   - Parallelize unseal operations
   - Cache compiled assets
5. Increase readyz timeout with exponential backoff
6. Update wait-for-ready.sh script
7. Verify reliability across multiple runs

**P3.5: ci-load (opentelemetry-collector healthcheck)**

**Issue**: Same as P3.3 (opentelemetry-collector-contrib exit 1)

**Fix**:

1. Coordinate with P3.3 fix (same root cause)
2. Apply same healthcheck fixes to compose.yml (not just compose.integration.yml)
3. Verify load testing scenario works end-to-end
4. Document load testing prerequisites

**Success Criteria**:

- ✅ ci-quality: passing (dependencies current)
- ✅ ci-mutation: passing (parallel, ≤20min total)
- ✅ ci-fuzz: passing (otel collector healthy)
- ✅ ci-dast: passing (readyz responsive)
- ✅ ci-load: passing (E2E load tests successful)

---

## Phase 4: Mutation Testing QA (Day 7-9, 32-48h)

**Objective**: 98%+ mutation kill rate per package

**Rationale**: 80% efficacy in 001-cryptoutil was too lenient. 98%+ ensures high test quality.

### Mutation Testing Strategy

**Priority Order** (highest-impact packages first):

1. **API Validation Packages**: JWK/JWE/JWS/JWT, OAuth 2.1, crypto operations
2. **Business Logic Packages**: client auth, IdP, barrier services, crypto utilities
3. **Repository Layer**: sqlrepository, identity repository, jose repository
4. **Infrastructure**: apperr, config, telemetry

**Per-Package Workflow**:

1. **Baseline**: Run `gremlins unleash ./pkg` → document efficacy %
2. **Analysis**: Identify lived mutants (boundary conditions, error paths, edge cases)
3. **Improvement**: Write tests targeting lived mutants (not generic tests)
4. **Verification**: Re-run gremlins, confirm ≥98% efficacy achieved

### Major Areas

**P4.1: API Validation Packages (Highest Priority)**

- internal/jose (JWK/JWE/JWS/JWT validation)
- internal/identity/authz (OAuth 2.1 validation)
- internal/kms/server/businesslogic (crypto operation validation)

**P4.2: Business Logic Packages (High Priority)**

- internal/identity/authz/clientauth (client authentication logic)
- internal/identity/idp (identity provider flows)
- internal/kms/server/barrier/* (barrier services)
- internal/shared/crypto/* (cryptographic utilities)

**P4.3: Repository Layer Packages (Medium Priority)**

- internal/kms/server/repository/sqlrepository
- internal/identity/repository
- internal/jose/server/repository

**P4.4: Infrastructure Packages (Lower Priority)**

- internal/shared/apperr
- internal/shared/config
- internal/shared/telemetry

**Success Criteria**:

- ✅ API validation packages: ≥98% efficacy
- ✅ Business logic packages: ≥98% efficacy
- ✅ Repository layer: ≥98% efficacy
- ✅ Infrastructure: ≥98% efficacy

---

## Phase 5: Hash Service Refactoring (Day 21-26, 24-36h)

**Objective**: 4 hash registry types × 3 versions per type, FIPS 140-3 compliant

**Rationale**: Current hash implementation lacks algorithm agility. Refactor to support low/high entropy inputs with deterministic/random variants.

### Hash Architecture Design

```
HashService
├── LowEntropyRandomHashRegistry (PBKDF2-based)
│   ├── v1: 0-31 bytes → PBKDF2-HMAC-SHA256 (OWASP rounds)
│   ├── v2: 32-47 bytes → PBKDF2-HMAC-SHA384
│   └── v3: 48+ bytes → PBKDF2-HMAC-SHA512
├── LowEntropyDeterministicHashRegistry (PBKDF2-based, no salt)
├── HighEntropyRandomHashRegistry (HKDF-based)
│   ├── v1: 0-31 bytes → HKDF-HMAC-SHA256
│   ├── v2: 32-47 bytes → HKDF-HMAC-SHA384
│   └── v3: 48+ bytes → HKDF-HMAC-SHA512
└── HighEntropyDeterministicHashRegistry (HKDF-based, no salt)
```

**API per registry**:

- `HashWithLatest(input []byte) (string, error)` - Uses current version
- `HashWithVersion(input []byte, version int) (string, error)` - Uses specific version
- `Verify(input []byte, hashed string) (bool, error)` - Verifies against any version

### Implementation Phases

**P5.1: Analysis and Design**

1. Analyze current hash implementation architecture
2. Design parameterized base registry class
3. Design version selection logic (input size-based)
4. Design hash output format (includes version metadata)
5. Document migration strategy from current hashing

**P5.2: Base Registry Implementation**

1. Create BaseHashRegistry with version management
2. Implement version selection by input size
3. Implement HashWithLatest method
4. Implement HashWithVersion method
5. Implement Verify method (version-aware)

**P5.3: Low Entropy Random Hash Registry**

1. Implement LowEntropyRandomHashRegistry
2. Configure PBKDF2 parameters (OWASP rounds, salt generation)
3. Implement v1 (SHA256), v2 (SHA384), v3 (SHA512)
4. Write comprehensive tests (all versions, edge cases)

**P5.4: Low Entropy Deterministic Hash Registry**

1. Implement LowEntropyDeterministicHashRegistry
2. Configure PBKDF2 without salt (deterministic)
3. Implement v1/v2/v3 variants
4. Write comprehensive tests

**P5.5: High Entropy Random Hash Registry**

1. Implement HighEntropyRandomHashRegistry
2. Configure HKDF parameters (salt, info strings)
3. Implement v1/v2/v3 variants
4. Write comprehensive tests

**P5.6: High Entropy Deterministic Hash Registry**

1. Implement HighEntropyDeterministicHashRegistry
2. Configure HKDF without salt (deterministic)
3. Implement v1/v2/v3 variants
4. Write comprehensive tests

**Success Criteria**:

- ✅ 4 hash registries implemented (Low/High × Random/Deterministic)
- ✅ 3 versions per registry (v1 SHA-256, v2 SHA-384, v3 SHA-512)
- ✅ HashRegistry interface consistent across all types
- ✅ Hash output format includes version metadata
- ✅ Version-aware Verify method
- ✅ Migration strategy documented
- ✅ FIPS 140-3 compliance validated (PBKDF2/HKDF approved)
- ✅ ≥95% coverage, ≥98% mutation efficacy

---

## Phase 6: Service Template Extraction (Day 27-38, 48-72h)

**Objective**: Extract reusable service template from SM-KMS for 8 PRODUCT-SERVICE instances

**Rationale**: 8 services duplicate infrastructure code. Reusable template eliminates duplication, ensures consistency.

### 8 PRODUCT-SERVICE Instances

1. **sm-kms**: Secrets Manager - Key Management System
2. **pki-ca**: Public Key Infrastructure - Certificate Authority
3. **jose-ja**: JOSE - JWK Authority
4. **identity-authz**: Identity - Authorization Server
5. **identity-idp**: Identity - Identity Provider
6. **identity-rs**: Identity - Resource Server
7. **identity-rp**: Identity - Relying Party (BFF pattern)
8. **identity-spa**: Identity - Single Page Application (static hosting)

### Template Features

**Server Template**:

- Dual HTTPS servers (public 8xxx, admin 127.0.0.1:9090)
- Dual API paths (/browser session-based, /service token-based)
- Middleware pipeline (CORS, CSRF, CSP, rate limiting, IP allowlist)
- Route registration framework
- Lifecycle management (start/stop/reload)

**Client Template**:

- HTTP client configuration (mTLS, timeouts, retries)
- Authentication strategies (OAuth 2.1, mTLS, API key, JWT)
- Request/response interceptors
- SDK generation from OpenAPI specs

**Database Template**:

- Dual database support (PostgreSQL + SQLite)
- Connection pool management
- Migration framework integration
- Transaction handling
- GORM patterns (UUID PKs, soft delete, audit fields)

**Barrier Services** (Optional per Service):

- Unseal/Seal operations
- Root/Intermediate/Content key hierarchy
- Encrypted-at-rest protection
- Optional integration (not all services need barriers)

**Telemetry Template**:

- OTLP exporter configuration
- Trace propagation
- Metric collection
- Log forwarding
- Service-specific instrumentation hooks

**Configuration Template**:

- YAML configuration structure
- Configuration validation
- Docker secrets support
- Environment-specific overrides

### Implementation Phases

**P6.1: Analysis Phase (Extract SM-KMS Patterns)**

1. Document SM-KMS architecture (dual HTTPS, dual paths, middleware, barriers, OpenAPI)
2. Identify common patterns across all 8 services
3. Identify service-specific customization points
4. Design template parameterization strategy

**P6.2: Create Reusable Server Template Package**

1. Create internal/template/server package structure
2. Implement ServerTemplate base class
3. Implement PublicAPIRouter (/browser vs /service)
4. Implement AdminAPIRouter (/admin/v1)
5. Implement MiddlewareBuilder

**P6.3: Create Reusable Client Template Package**

1. Create internal/template/client package structure
2. Implement ClientSDK base class
3. Implement authentication strategies
4. Generate client SDKs from OpenAPI specs

**P6.4: Database Layer Abstraction**

1. Create internal/template/repository package
2. Implement dual database support (PostgreSQL + SQLite)
3. Implement GORM patterns

**P6.5: Barrier Services Integration (Optional per Service)**

1. Make barrier services optional
2. Provide barrier-enabled base template
3. Provide barrier-free base template
4. Document when to use barrier services

**P6.6: Telemetry Integration**

1. Standardize OpenTelemetry setup
2. Create telemetry middleware
3. Implement service-specific instrumentation hooks
4. Document telemetry best practices

**P6.7: Configuration Management**

1. Standardize YAML configuration structure
2. Implement configuration validation
3. Support Docker secrets for sensitive data
4. Document configuration patterns

**P6.8: Documentation and Examples**

1. Create template usage guide
2. Document customization points
3. Provide step-by-step service creation guide
4. Create comparison table: custom vs template approach

**Success Criteria**:

- ✅ Server template extracted from SM-KMS
- ✅ Client template with SDK generation
- ✅ Database abstraction (PostgreSQL + SQLite)
- ✅ Barrier services optional integration
- ✅ Telemetry standardization
- ✅ Configuration patterns documented
- ✅ Template usage guide complete

---

## Phase 7: Learn-PS Demonstration (Day 39-47, 36-54h)

**Objective**: Working Pet Store service validating template reusability

**Rationale**: Prove template works for NEW services (not just refactored existing ones). Learn-PS serves as copy-paste-modify starting point for customers.

### Learn-PS Overview

- **Product**: Learn (educational/demonstration product)
- **Service**: PS (Pet Store service)
- **Purpose**: Validate template completeness, provide customer starting point
- **Scope**: Complete CRUD API for pet store inventory, orders, customers

### Implementation Phases

**P7.1: Service Design**

1. Define requirements (API endpoints, data model, business logic, authentication)
2. Create OpenAPI specification (all endpoints, schemas, error responses)
3. Design database schema (pets, orders, customers, order_items tables)

**P7.2: Service Implementation using Template**

1. Instantiate ServerTemplate (dual HTTPS, routes, middleware)
2. Implement business logic handlers (CreatePet, GetPet, UpdatePet, DeletePet, CreateOrder, etc.)
3. Implement repository layer (GORM models, CRUD, transactions)
4. Generate client SDK from OpenAPI spec

**P7.3: Testing**

1. Write unit tests (95%+ coverage target)
2. Write integration tests (E2E API flows, database interaction)
3. Run mutation testing (98%+ efficacy target)
4. Performance testing (<12s test execution)

**P7.4: Deployment**

1. Create Docker Compose configuration (Learn-PS, PostgreSQL, Otel Collector, health checks)
2. Create Kubernetes manifests (Deployment, Service, Ingress, ConfigMap/Secrets)

**P7.5: Documentation**

1. Write README.md (overview, quick start, API docs, development guide)
2. Create tutorial series (4 parts: using, understanding, customizing, deploying)
3. Record video demonstration (startup, API usage, code walkthrough, customization tips)

**Success Criteria**:

- ✅ Pet Store service operational
- ✅ 95%+ coverage, 98%+ mutation efficacy
- ✅ Tests execute in <12s
- ✅ Docker Compose deployment working
- ✅ Tutorial series complete (4 parts)
- ✅ Video demonstration recorded

---

## Quality Gates (Enforced at Each Phase)

**Test Performance**:

- ✅ ALL !integration packages ≤30 seconds execution time (target)
- ✅ Hard limit: 60 seconds (blocking failure)
- ✅ Total !integration suite <100s
- ✅ Probabilistic testing patterns applied where appropriate
- ✅ Coverage maintained (no losses from optimization)

**Code Coverage**:

- ✅ Production packages: ≥95% coverage (NO EXCEPTIONS)
- ✅ Infrastructure/utility: ≥100% coverage (NO EXCEPTIONS)
- ✅ Per-package verification complete
- ✅ Gap analysis documented for each package

**Mutation Testing**:

- ✅ API validation packages: ≥98% efficacy
- ✅ Business logic packages: ≥98% efficacy
- ✅ Repository layer: ≥98% efficacy
- ✅ Infrastructure: ≥98% efficacy

**CI/CD Health**:

- ✅ ALL workflows passing (0 failures)
- ✅ Quality gates enforced
- ✅ Dependencies current

**Linting**:

- ✅ golangci-lint passing (ALL packages)
- ✅ NO `//nolint:` directives (except documented linter bugs)
- ✅ ALL files UTF-8 without BOM

---

## Success Criteria (Overall)

**MVP Quality Achieved**:

- ✅ Fast tests (≤30s per package, <100s total !integration suite)
- ✅ High coverage (95%+ production, 100% infra/util, NO EXCEPTIONS)
- ✅ Stable CI/CD (0 failures, time targets met)
- ✅ High mutation kill (98%+ per package)
- ✅ Clean hash architecture (4 registries × 3 versions, FIPS 140-3)

**Service Template Ready**:

- ✅ Reusable server template extracted
- ✅ Reusable client template with SDK generation
- ✅ Database abstraction working
- ✅ Template documentation complete
- ✅ Learn-PS demonstration operational

**Customer Deliverables**:

- ✅ 4 products operational (SM-KMS, JOSE-JA, Identity, PKI-CA)
- ✅ Docker Compose deployment working
- ✅ Learn-PS demo service as starting point
- ✅ Tutorial series (4 parts)
- ✅ Video demonstration

---

## Post-Implementation Checklist

- [ ] Update docs/README.md with 002-cryptoutil results
- [ ] Document lessons learned in implement/EXECUTIVE.md
- [ ] Create post-mortem for any P0 incidents
- [ ] Archive 002-cryptoutil if starting 003-cryptoutil
- [ ] Push all commits
- [ ] Tag release: `git tag -a v0.2.0 -m "MVP quality release with service template"`
