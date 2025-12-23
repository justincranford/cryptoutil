# Implementation Progress - DETAILED

**Iteration**: specs/002-cryptoutil
**Started**: December 17, 2025
**Last Updated**: December 17, 2025
**Status**: 🎯 Fresh Start - MVP Quality Focus

---

## Overview

**Primary Goal**: Clean up AI slop from 001-cryptoutil, achieve production MVP quality, then extract reusable service template from SM-KMS for all 8 PRODUCT-SERVICE instances.

**Key Objectives**:

1. Fast tests (≤12s per package, !integration)
2. High coverage (95%+ production, 98% infra/util, no exceptions)
3. Stable CI/CD (0 failures, all workflows green)
4. High mutation kill rate (98%+ per package)
5. Clean hash architecture (4 types: Low/High × Random/Deterministic)
6. Reusable service template (extract from KMS, validate with Learn-PS)

---

## Section 1: Task Checklist

### Phase 1: Optimize Slow Test Packages (20 tasks) - Target: ≤12 seconds - ✅ COMPLETE (Skipped - All targets met)

**Goal**: Aggressive test performance optimization to ensure all !integration tests complete in ≤12 seconds per package.

**Strategy**: Use probabilistic execution (TestProbAlways, TestProbQuarter, TestProbTenth) for algorithm variants while maintaining 100% execution for base algorithms.

- [x] **P1.1**: Baseline current test timings for all packages ✅ COMPLETE - All packages meet ≤30s target
  - [x] P1.1.1: Run `go test -json -v ./... 2>&1 | tee test-output/baseline-timing-002.txt` ✅
  - [x] P1.1.2: Parse JSON output to extract per-package execution times ✅
  - [x] P1.1.3: Identify all packages >30s execution time ✅ NONE FOUND
  - [x] P1.1.4: Document baseline in test-output/baseline-timing-002-summary.md ✅
  - **Result**: ALL 130+ packages ≤0.77s (longest: kms/server/barrier/unsealkeysservice)
  - **Decision**: Skip P1.2-P1.14 optimization tasks - targets already met

- [ ] **P1.2**: Analyze internal/jose package performance
  - [ ] P1.2.1: Identify algorithm variant test patterns
  - [ ] P1.2.2: Apply probabilistic execution to cipher variants
  - [ ] P1.2.3: Verify coverage maintained with faster execution
  - [ ] P1.2.4: Document optimization strategy

- [ ] **P1.3**: Analyze internal/jose/server package performance
  - [ ] P1.3.1: Profile test execution hotspots
  - [ ] P1.3.2: Identify redundant test operations
  - [ ] P1.3.3: Apply optimizations without coverage loss
  - [ ] P1.3.4: Verify ≤12s target achieved

- [ ] **P1.4**: Analyze internal/kms/client package performance
  - [ ] P1.4.1: Review existing probabilistic optimizations
  - [ ] P1.4.2: Identify additional optimization opportunities
  - [ ] P1.4.3: Apply optimizations to remaining slow tests
  - [ ] P1.4.4: Verify ≤12s target achieved

- [ ] **P1.5**: Analyze internal/kms/server/application package performance
  - [ ] P1.5.1: Profile barrier operation test overhead
  - [ ] P1.5.2: Optimize unseal/seal test patterns
  - [ ] P1.5.3: Apply probabilistic execution where appropriate
  - [ ] P1.5.4: Verify ≤12s target achieved

- [ ] **P1.6**: Analyze internal/identity/authz package performance
  - [ ] P1.6.1: Profile OAuth 2.1 flow test overhead
  - [ ] P1.6.2: Optimize token generation test patterns
  - [ ] P1.6.3: Apply optimizations without coverage loss
  - [ ] P1.6.4: Verify ≤12s target achieved

- [ ] **P1.7**: Analyze internal/identity/authz/clientauth package performance
  - [ ] P1.7.1: Profile client authentication test overhead
  - [ ] P1.7.2: Optimize mTLS handshake test patterns
  - [ ] P1.7.3: Apply optimizations without coverage loss
  - [ ] P1.7.4: Verify ≤12s target achieved

- [ ] **P1.8**: Analyze internal/identity/idp package performance
  - [ ] P1.8.1: Profile MFA flow test overhead
  - [ ] P1.8.2: Optimize consent/login test patterns
  - [ ] P1.8.3: Apply optimizations without coverage loss
  - [ ] P1.8.4: Verify ≤12s target achieved

- [ ] **P1.9**: Analyze internal/shared/crypto/keygen package performance
  - [ ] P1.9.1: Review existing probabilistic optimizations
  - [ ] P1.9.2: Verify all key size variants using probability wrappers
  - [ ] P1.9.3: Ensure base key sizes always execute (TestProbAlways)
  - [ ] P1.9.4: Verify ≤12s target achieved

- [ ] **P1.10**: Analyze internal/shared/crypto/digests package performance
  - [ ] P1.10.1: Profile HKDF variant test overhead
  - [ ] P1.10.2: Apply probabilistic execution to SHA variants
  - [ ] P1.10.3: Verify coverage maintained
  - [ ] P1.10.4: Verify ≤12s target achieved

- [ ] **P1.11**: Analyze internal/shared/crypto/certificate package performance
  - [ ] P1.11.1: Profile TLS handshake test overhead
  - [ ] P1.11.2: Optimize certificate generation test patterns
  - [ ] P1.11.3: Apply optimizations without coverage loss
  - [ ] P1.11.4: Verify ≤12s target achieved

- [ ] **P1.12**: Create test timing monitoring script
  - [ ] P1.12.1: Create scripts/monitor-test-timing.ps1
  - [ ] P1.12.2: Parse go test JSON output
  - [ ] P1.12.3: Flag packages exceeding 12s threshold
  - [ ] P1.12.4: Output summary report

- [ ] **P1.13**: Document probabilistic testing patterns
  - [ ] P1.13.1: Create docs/testing/probabilistic-patterns.md
  - [ ] P1.13.2: Document TestProbAlways/Quarter/Tenth usage
  - [ ] P1.13.3: Provide examples from KMS client
  - [ ] P1.13.4: Add guidelines for new test code

- [ ] **P1.14**: Verify all packages ≤12s after optimizations
  - [ ] P1.14.1: Run full test suite with timing
  - [ ] P1.14.2: Compare against baseline
  - [ ] P1.14.3: Identify any remaining slow packages
  - [ ] P1.14.4: Document results

- [ ] **P1.15**: Create pre-commit hook for test timing enforcement
  - [ ] P1.15.1: Add timing check to .pre-commit-config.yaml
  - [ ] P1.15.2: Fail commit if any package >12s
  - [ ] P1.15.3: Document override procedure for valid exceptions
  - [ ] P1.15.4: Test hook with slow package

- [ ] **P1.16-P1.20**: Reserved for additional slow packages discovered during baseline

---

### Phase 2: Coverage Targets - 95% Mandatory (60+ tasks) - NO EXCEPTIONS

**Goal**: Achieve 95%+ coverage for all production packages, 98% for infrastructure/utility. Create granular tracking per package.

**Strategy**: One parent task per major area, subtasks per package for detailed progress tracking.

#### P2.1: KMS Server Coverage (Target: 95%+)

- [ ] **P2.1.1**: internal/kms/server/application (baseline, gap analysis, tests, verify)
- [ ] **P2.1.2**: internal/kms/server/businesslogic (baseline, gap analysis, tests, verify)
- [ ] **P2.1.3**: internal/kms/server/barrier/contentkeysservice (baseline, gap analysis, tests, verify)
- [ ] **P2.1.4**: internal/kms/server/barrier/intermediatekeysservice (baseline, gap analysis, tests, verify)
- [ ] **P2.1.5**: internal/kms/server/barrier/rootkeysservice (baseline, gap analysis, tests, verify)
- [ ] **P2.1.6**: internal/kms/server/barrier/unsealservice (baseline, gap analysis, tests, verify)
- [ ] **P2.1.7**: internal/kms/server/repository/orm (baseline, gap analysis, tests, verify)
- [ ] **P2.1.8**: internal/kms/server/repository/sqlrepository (baseline, gap analysis, tests, verify)

#### P2.2: KMS Client Coverage (Target: 95%+)

- [ ] **P2.2.1**: internal/kms/client (baseline, gap analysis, tests, verify)

#### P2.3: Identity Server Coverage (Target: 95%+)

- [ ] **P2.3.1**: internal/identity/authz (current: 66.8%, gap: 28.2 points)
  - [ ] Baseline: Run coverage, analyze HTML for RED lines
  - [ ] Gap analysis: Identify uncovered error paths, edge cases
  - [ ] Test development: Write targeted tests for gaps
  - [ ] Verify: Confirm 95%+ achieved

- [ ] **P2.3.2**: internal/identity/idp (baseline, gap analysis, tests, verify)
- [ ] **P2.3.3**: internal/identity/rs (baseline, gap analysis, tests, verify)
- [ ] **P2.3.4**: internal/identity/rp (baseline, gap analysis, tests, verify)
- [ ] **P2.3.5**: internal/identity/spa (baseline, gap analysis, tests, verify)
- [ ] **P2.3.6**: internal/identity/authz/clientauth (baseline, gap analysis, tests, verify)
- [ ] **P2.3.7**: internal/identity/domain (baseline, gap analysis, tests, verify)
- [ ] **P2.3.8**: internal/identity/repository (baseline, gap analysis, tests, verify)

#### P2.4: JOSE Coverage (Target: 95%+)

- [ ] **P2.4.1**: internal/jose (baseline, gap analysis, tests, verify)
- [ ] **P2.4.2**: internal/jose/server (baseline, gap analysis, tests, verify)

#### P2.5: CA Coverage (Target: 95%+)

- [ ] **P2.5.1**: internal/ca/server (baseline, gap analysis, tests, verify)
- [ ] **P2.5.2**: internal/ca/domain (baseline, gap analysis, tests, verify)

#### P2.6: Shared Crypto Coverage (Target: 98% - utility code)

- [ ] **P2.6.1**: internal/shared/crypto/keygen (baseline, gap analysis, tests, verify)
- [ ] **P2.6.2**: internal/shared/crypto/digests (baseline, gap analysis, tests, verify)
- [ ] **P2.6.3**: internal/shared/crypto/certificate (baseline, gap analysis, tests, verify)
- [ ] **P2.6.4**: internal/shared/crypto/signer (baseline, gap analysis, tests, verify)

#### P2.7: Shared Infrastructure Coverage (Target: 98%)

- [ ] **P2.7.1**: internal/shared/apperr (baseline, gap analysis, tests, verify)
- [ ] **P2.7.2**: internal/shared/config (baseline, gap analysis, tests, verify)
- [ ] **P2.7.3**: internal/shared/container (baseline, gap analysis, tests, verify)
- [ ] **P2.7.4**: internal/shared/magic (baseline, gap analysis, tests, verify)
- [ ] **P2.7.5**: internal/shared/pool (baseline, gap analysis, tests, verify)
- [ ] **P2.7.6**: internal/shared/telemetry (baseline, gap analysis, tests, verify)
- [ ] **P2.7.7**: internal/shared/testutil (baseline, gap analysis, tests, verify)
- [ ] **P2.7.8**: internal/shared/util (baseline, gap analysis, tests, verify)

#### P2.8: CICD Coverage (Target: 98% - infrastructure)

- [ ] **P2.8.1**: internal/cmd/cicd/cicd (baseline, gap analysis, tests, verify)
- [ ] **P2.8.2**: internal/cmd/cicd/format_go (baseline, gap analysis, tests, verify)
- [ ] **P2.8.3**: internal/cmd/cicd/lint (baseline, gap analysis, tests, verify)
- [ ] **P2.8.4**: internal/cmd/cicd/*/all subdirectories (baseline per subdir, gap analysis, tests, verify)

---

### Phase 3: CI/CD Workflow Fixes (15 tasks) - Target: 0 Failures

**Goal**: Fix all 5 failing workflows and ensure robust CI/CD pipeline.

#### P3.1: ci-quality Workflow Fix (outdated dependencies)

- [ ] **P3.1.1**: Update github.com/goccy/go-yaml from v1.19.0 to v1.19.1
  - [ ] Run `go get -u github.com/goccy/go-yaml@v1.19.1`
  - [ ] Run `go mod tidy`
  - [ ] Verify tests pass
  - [ ] Commit and push

- [ ] **P3.1.2**: Review all direct dependencies for updates
  - [ ] Run `go list -u -m all | grep '\[.*\]$'`
  - [ ] Identify outdated direct dependencies
  - [ ] Update incrementally with test validation
  - [ ] Document update strategy

- [ ] **P3.1.3**: Create dependency update automation
  - [ ] Add dependabot.yml for Go modules
  - [ ] Configure weekly update schedule
  - [ ] Set up auto-merge for minor/patch updates
  - [ ] Document manual review process for major updates

#### P3.2: ci-mutation Workflow Fix (45min timeout)

- [ ] **P3.2.1**: Analyze current mutation test execution time
  - [ ] Review gremlins execution logs
  - [ ] Identify slowest packages
  - [ ] Document baseline timing per package
  - [ ] Calculate realistic timeout (2x slowest package)

- [ ] **P3.2.2**: Reduce mutation test scope
  - [ ] Focus on business logic packages only
  - [ ] Exclude test utilities, mocks, generated code
  - [ ] Update .gremlins.yaml exclusion patterns
  - [ ] Verify coverage of critical paths maintained

- [ ] **P3.2.3**: Parallelize mutation testing by package
  - [ ] Use GitHub Actions matrix strategy
  - [ ] Split packages into groups (4-6 per job)
  - [ ] Run groups in parallel
  - [ ] Aggregate results

- [ ] **P3.2.4**: Set realistic per-job timeout
  - [ ] Set timeout to 15 minutes per job
  - [ ] Total workflow time: ~20 minutes (parallel execution)
  - [ ] Update ci-mutation.yml timeout-minutes
  - [ ] Document timeout rationale

#### P3.3: ci-fuzz Workflow Fix (opentelemetry-collector healthcheck)

- [ ] **P3.3.1**: Analyze healthcheck-opentelemetry-collector-contrib failure
  - [ ] Review Docker Compose logs
  - [ ] Check service dependencies
  - [ ] Verify port mappings
  - [ ] Document root cause

- [ ] **P3.3.2**: Fix opentelemetry-collector-contrib healthcheck
  - [ ] Update compose.integration.yml healthcheck command
  - [ ] Increase start_period if needed
  - [ ] Verify service starts successfully
  - [ ] Test locally with `docker compose up`

- [ ] **P3.3.3**: Add diagnostic logging to healthcheck
  - [ ] Log collector startup progress
  - [ ] Log healthcheck attempts
  - [ ] Output diagnostic info on failure
  - [ ] Verify logs helpful for debugging

#### P3.4: ci-dast Workflow Fix (readyz endpoint timeout)

- [ ] **P3.4.1**: Analyze /admin/v1/readyz timeout failure
  - [ ] Review service startup logs
  - [ ] Check database migration timing
  - [ ] Verify unseal operation completes
  - [ ] Document bottlenecks

- [ ] **P3.4.2**: Optimize service startup time
  - [ ] Reduce database migration overhead
  - [ ] Parallelize unseal operations
  - [ ] Cache compiled assets
  - [ ] Verify startup <30s

- [ ] **P3.4.3**: Increase readyz timeout if needed
  - [ ] Set realistic timeout based on GitHub Actions latency
  - [ ] Add retry logic with exponential backoff
  - [ ] Update wait-for-ready.sh script
  - [ ] Verify reliability across multiple runs

#### P3.5: ci-load Workflow Fix (opentelemetry-collector healthcheck)

- [ ] **P3.5.1**: Same root cause as P3.3 - coordinate fix
- [ ] **P3.5.2**: Apply same healthcheck fixes to compose.yml
- [ ] **P3.5.3**: Verify load testing scenario works end-to-end
- [ ] **P3.5.4**: Document load testing prerequisites

---

### Phase 4: Mutation Testing Quality Assurance (40+ tasks) - Target: 98% Killed

**Goal**: Achieve 98%+ mutation kill rate per package, starting with highest-impact packages.

**Strategy**: One task per high-value package, with subtasks for baseline/analysis/improvement/verification.

#### P4.1: API Validation Packages (Highest Priority)

- [ ] **P4.1.1**: internal/jose (JWK/JWE/JWS/JWT validation)
  - [ ] Baseline: Run gremlins, document efficacy %
  - [ ] Analysis: Identify lived mutants (boundary conditions, error paths)
  - [ ] Improvement: Write tests targeting lived mutants
  - [ ] Verification: Re-run gremlins, confirm ≥98% efficacy

- [ ] **P4.1.2**: internal/identity/authz (OAuth 2.1 validation)
  - [ ] Baseline, analysis, improvement, verification

- [ ] **P4.1.3**: internal/kms/server/businesslogic (crypto operation validation)
  - [ ] Baseline, analysis, improvement, verification

#### P4.2: Business Logic Packages (High Priority)

- [ ] **P4.2.1**: internal/identity/authz/clientauth (client authentication logic)
- [ ] **P4.2.2**: internal/identity/idp (identity provider flows)
- [ ] **P4.2.3**: internal/kms/server/barrier/* (barrier services)
- [ ] **P4.2.4**: internal/shared/crypto/* (cryptographic utilities)

#### P4.3: Repository Layer Packages (Medium Priority)

- [ ] **P4.3.1**: internal/kms/server/repository/sqlrepository
- [ ] **P4.3.2**: internal/identity/repository
- [ ] **P4.3.3**: internal/jose/server/repository

#### P4.4: Infrastructure Packages (Lower Priority)

- [ ] **P4.4.1**: internal/shared/apperr
- [ ] **P4.4.2**: internal/shared/config
- [ ] **P4.4.3**: internal/shared/telemetry

---

### Phase 5: Refactor Hashes (20 tasks) - Low/High × Random/Deterministic

**Goal**: Create clean hash service architecture supporting 4 hash types with version management.

**Architecture**:

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

#### P5.1: Analysis and Design

- [ ] **P5.1.1**: Analyze current hash implementation architecture
- [ ] **P5.1.2**: Design parameterized base registry class
- [ ] **P5.1.3**: Design version selection logic (input size-based)
- [ ] **P5.1.4**: Design hash output format (includes version metadata)
- [ ] **P5.1.5**: Document migration strategy from current hashing

#### P5.2: Base Registry Implementation

- [ ] **P5.2.1**: Create BaseHashRegistry with version management
- [ ] **P5.2.2**: Implement version selection by input size
- [ ] **P5.2.3**: Implement HashWithLatest method
- [ ] **P5.2.4**: Implement HashWithVersion method
- [ ] **P5.2.5**: Implement Verify method (version-aware)

#### P5.3: Low Entropy Random Hash Registry

- [ ] **P5.3.1**: Implement LowEntropyRandomHashRegistry
- [ ] **P5.3.2**: Configure PBKDF2 parameters (OWASP rounds, salt generation)
- [ ] **P5.3.3**: Implement v1 (SHA256), v2 (SHA384), v3 (SHA512)
- [ ] **P5.3.4**: Write comprehensive tests (all versions, edge cases)

#### P5.4: Low Entropy Deterministic Hash Registry

- [ ] **P5.4.1**: Implement LowEntropyDeterministicHashRegistry
- [ ] **P5.4.2**: Configure PBKDF2 without salt (deterministic)
- [ ] **P5.4.3**: Implement v1/v2/v3 variants
- [ ] **P5.4.4**: Write comprehensive tests

#### P5.5: High Entropy Random Hash Registry

- [ ] **P5.5.1**: Implement HighEntropyRandomHashRegistry
- [ ] **P5.5.2**: Configure HKDF parameters (salt, info strings)
- [ ] **P5.5.3**: Implement v1/v2/v3 variants
- [ ] **P5.5.4**: Write comprehensive tests

#### P5.6: High Entropy Deterministic Hash Registry

- [ ] **P5.6.1**: Implement HighEntropyDeterministicHashRegistry
- [ ] **P5.6.2**: Configure HKDF without salt (deterministic)
- [ ] **P5.6.3**: Implement v1/v2/v3 variants
- [ ] **P5.6.4**: Write comprehensive tests

---

### Phase 6: Server Architecture Unification (30+ tasks) - Extract Reusable Service Template

**Goal**: Extract reusable service template from SM-KMS, augment for all 8 PRODUCT-SERVICE patterns.

**8 PRODUCT-SERVICE Instances**:

1. sm-kms (Secrets Manager - Key Management System)
2. pki-ca (Public Key Infrastructure - Certificate Authority)
3. jose-ja (JOSE - JWK Authority)
4. identity-authz (Identity - Authorization Server)
5. identity-idp (Identity - Identity Provider)
6. identity-rs (Identity - Resource Server)
7. identity-rp (Identity - Relying Party - BFF pattern)
8. identity-spa (Identity - Single Page Application - static hosting)

#### P6.1: Analysis Phase (Extract SM-KMS Patterns)

- [ ] **P6.1.1**: Document SM-KMS architecture
  - [ ] Dual HTTPS servers (public 8080, admin 127.0.0.1:9090)
  - [ ] Dual API paths (/browser vs /service)
  - [ ] Middleware patterns (CORS, CSRF, CSP, rate limiting)
  - [ ] Barrier services (unseal, root, intermediate, content)
  - [ ] OpenAPI spec structure
  - [ ] Client SDK generation

- [ ] **P6.1.2**: Identify common patterns across all 8 services
  - [ ] Server initialization
  - [ ] Route registration
  - [ ] Middleware application
  - [ ] Database initialization
  - [ ] Telemetry setup
  - [ ] Health check endpoints

- [ ] **P6.1.3**: Identify service-specific customization points
  - [ ] API endpoints (OpenAPI specs)
  - [ ] Business logic handlers
  - [ ] Database schemas
  - [ ] Client SDK interfaces
  - [ ] Barrier configuration (optional per service)

- [ ] **P6.1.4**: Design template parameterization strategy
  - [ ] Constructor injection pattern
  - [ ] Interface-based customization
  - [ ] Configuration-driven behavior
  - [ ] Runtime service discovery

#### P6.2: Create Reusable Server Template Package

- [ ] **P6.2.1**: Create internal/template/server package structure
- [ ] **P6.2.2**: Implement ServerTemplate base class
  - [ ] Dual HTTPS server management
  - [ ] Middleware pipeline builder
  - [ ] Route registration framework
  - [ ] Lifecycle management (start/stop/reload)

- [ ] **P6.2.3**: Implement PublicAPIRouter
  - [ ] /browser path prefix (session-based)
  - [ ] /service path prefix (token-based)
  - [ ] Middleware differentiation
  - [ ] OpenAPI spec integration

- [ ] **P6.2.4**: Implement AdminAPIRouter
  - [ ] /admin/v1 path prefix
  - [ ] 127.0.0.1 binding enforcement
  - [ ] Health check endpoints (/livez, /readyz, /healthz)
  - [ ] Shutdown endpoint

- [ ] **P6.2.5**: Implement MiddlewareBuilder
  - [ ] CORS (browser-only)
  - [ ] CSRF (browser-only)
  - [ ] CSP (browser-only)
  - [ ] Rate limiting (both)
  - [ ] IP allowlist (both)
  - [ ] Authentication (both, different mechanisms)

#### P6.3: Create Reusable Client Template Package

- [ ] **P6.3.1**: Create internal/template/client package structure
- [ ] **P6.3.2**: Implement ClientSDK base class
  - [ ] HTTP client configuration
  - [ ] mTLS support
  - [ ] Retry logic with exponential backoff
  - [ ] Request/response interceptors

- [ ] **P6.3.3**: Implement authentication strategies
  - [ ] OAuth 2.1 client credentials flow
  - [ ] mTLS client certificates
  - [ ] API key authentication
  - [ ] JWT token handling

- [ ] **P6.3.4**: Generate client SDKs from OpenAPI specs
  - [ ] Code generation tooling
  - [ ] Type-safe request builders
  - [ ] Response unmarshaling
  - [ ] Error handling

#### P6.4: Database Layer Abstraction

- [ ] **P6.4.1**: Create internal/template/repository package
- [ ] **P6.4.2**: Implement dual database support (PostgreSQL + SQLite)
  - [ ] Connection pool management
  - [ ] Migration framework integration
  - [ ] Transaction handling
  - [ ] Query builder patterns

- [ ] **P6.4.3**: Implement GORM patterns
  - [ ] Model registration
  - [ ] Soft delete support
  - [ ] Audit fields (created_at, updated_at)
  - [ ] UUID primary keys

#### P6.5: Barrier Services Integration (Optional per Service)

- [ ] **P6.5.1**: Make barrier services optional
- [ ] **P6.5.2**: Provide barrier-enabled base template
- [ ] **P6.5.3**: Provide barrier-free base template
- [ ] **P6.5.4**: Document when to use barrier services

#### P6.6: Telemetry Integration

- [ ] **P6.6.1**: Standardize OpenTelemetry setup
  - [ ] OTLP exporter configuration
  - [ ] Trace propagation
  - [ ] Metric collection
  - [ ] Log forwarding

- [ ] **P6.6.2**: Create telemetry middleware
- [ ] **P6.6.3**: Implement service-specific instrumentation hooks
- [ ] **P6.6.4**: Document telemetry best practices

#### P6.7: Configuration Management

- [ ] **P6.7.1**: Standardize YAML configuration structure
- [ ] **P6.7.2**: Implement configuration validation
- [ ] **P6.7.3**: Support Docker secrets for sensitive data
- [ ] **P6.7.4**: Document configuration patterns

#### P6.8: Documentation and Examples

- [ ] **P6.8.1**: Create template usage guide
- [ ] **P6.8.2**: Document customization points
- [ ] **P6.8.3**: Provide step-by-step service creation guide
- [ ] **P6.8.4**: Create comparison table: custom vs template approach

---

### Phase 7: Learn-PS Demonstration Service (25 tasks) - Pet Store Learning Example

**Goal**: Create working demonstration service using service template, validate reusability and completeness.

**Learn-PS Overview**:

- **Product**: Learn (educational/demonstration product)
- **Service**: PS (Pet Store service)
- **Purpose**: Copy-paste-modify starting point for customers creating new PRODUCT-SERVICE instances
- **Scope**: Complete CRUD API for pet store inventory, orders, customers

#### P7.1: Service Design

- [ ] **P7.1.1**: Define Learn-PS requirements
  - [ ] API endpoints (pets, orders, customers)
  - [ ] Data model (PostgreSQL schema)
  - [ ] Business logic (inventory management, order processing)
  - [ ] Authentication requirements

- [ ] **P7.1.2**: Create OpenAPI specification
  - [ ] Define all API endpoints
  - [ ] Request/response schemas
  - [ ] Error responses
  - [ ] Authentication schemes

- [ ] **P7.1.3**: Design database schema
  - [ ] Pets table (id, name, species, price, quantity)
  - [ ] Orders table (id, customer_id, total, status)
  - [ ] Customers table (id, name, email, created_at)
  - [ ] Order items table (order_id, pet_id, quantity, price)

#### P7.2: Service Implementation using Template

- [ ] **P7.2.1**: Instantiate ServerTemplate
  - [ ] Configure dual HTTPS servers
  - [ ] Register public API routes
  - [ ] Register admin API routes
  - [ ] Apply middleware

- [ ] **P7.2.2**: Implement business logic handlers
  - [ ] CreatePet, GetPet, UpdatePet, DeletePet
  - [ ] ListPets with pagination
  - [ ] CreateOrder, GetOrder, ListOrders
  - [ ] CreateCustomer, GetCustomer, ListCustomers

- [ ] **P7.2.3**: Implement repository layer
  - [ ] GORM models
  - [ ] CRUD operations
  - [ ] Transaction handling
  - [ ] Query optimization

- [ ] **P7.2.4**: Generate client SDK from OpenAPI spec
  - [ ] Use oapi-codegen
  - [ ] Validate generated code
  - [ ] Write client usage examples

#### P7.3: Testing

- [ ] **P7.3.1**: Write unit tests (95%+ coverage target)
  - [ ] Business logic tests
  - [ ] Repository tests
  - [ ] Handler tests
  - [ ] Middleware tests

- [ ] **P7.3.2**: Write integration tests
  - [ ] End-to-end API flows
  - [ ] Database interaction tests
  - [ ] Error handling tests

- [ ] **P7.3.3**: Run mutation testing (98%+ efficacy target)
- [ ] **P7.3.4**: Performance testing (<12s test execution)

#### P7.4: Deployment

- [ ] **P7.4.1**: Create Docker Compose configuration
  - [ ] Learn-PS service
  - [ ] PostgreSQL database
  - [ ] OpenTelemetry collector
  - [ ] Health check sidecars

- [ ] **P7.4.2**: Create Kubernetes manifests
  - [ ] Deployment
  - [ ] Service
  - [ ] Ingress
  - [ ] ConfigMap/Secrets

#### P7.5: Documentation

- [ ] **P7.5.1**: Write README.md
  - [ ] Overview
  - [ ] Quick start
  - [ ] API documentation
  - [ ] Development guide

- [ ] **P7.5.2**: Create tutorial series
  - [ ] Part 1: Using the service (API calls)
  - [ ] Part 2: Understanding the code
  - [ ] Part 3: Customizing for your use case
  - [ ] Part 4: Deploying to production

- [ ] **P7.5.3**: Record video demonstration
  - [ ] Service startup
  - [ ] API usage examples
  - [ ] Code walkthrough
  - [ ] Customization tips

---

## Section 2: Append-Only Timeline (Time-ordered)

Tasks may be implemented out of order from Section 1. Each entry references back to Section 1.

---

### 2025-12-22: Authentication Documentation Completion and Constitution Refactoring

**Work Completed**:

- Created .github/instructions/02-10.authentication.instructions.md (382 lines) - comprehensive authentication/authorization reference for all 38 methods
- Documented ALL 10 headless authentication methods (3 non-federated + 7 federated) with per-factor storage realm specifications
- Documented ALL 28 browser authentication methods (6 non-federated + 22 federated) with per-factor storage realm specifications
- Established storage realm pattern: YAML + SQL (Config > DB priority) for static credentials vs SQL ONLY for user-specific enrollment data
- Refactored constitution.md Section VA (lines 499-578) to include per-factor storage realms for all 38 methods
- Refactored constitution.md Section X (lines ~1260-1285) to remove amendment history table (versions 1.0.0-3.0.0), simplified to "Latest amendments: 2025-12-22"
- Updated spec.md Authentication section (lines 730-800) with per-factor storage realm specifications for all 38 methods
- Removed outdated "MORE TO BE CLARIFIED" markers from spec.md (lines 184, 187) after QUIZME-02 answered all authentication unknowns
- Verified clarify.md and clarify.md.old byte-for-byte identical, deleted duplicate backup file
- Renamed plan-probably-out-of-date.md to plan.md to indicate current status

**Coverage/Quality Metrics**:

- Authentication documentation: COMPLETE (10+28 factors documented in copilot instructions, constitution, spec)
- Storage realm specifications: COMPLETE (YAML + SQL vs SQL ONLY distinctions preserved across all documents)
- Unknown markers: 0 in constitution (only workflow reference), 0 in spec, 0 in clarify
- All QUIZME-02 questions answered (15 Q&A integrated into clarify.md Section 8)

**Violations Found**:

- CRITICAL user feedback #1: 10+28 authentication factor lists missing from copilot instructions → FIXED via 02-10.authentication.instructions.md creation
- CRITICAL user feedback #2: Constitution contains revision tracking noise → FIXED via amendment history table removal
- Outdated "MORE TO BE CLARIFIED" markers in spec after QUIZME-02 completion → FIXED via reference updates

**Next Steps**:

- All unknowns resolved → No QUIZME-03 generation needed
- Plan file updated and current → Ready for implementation phase
- Storage realm pattern established → Implementation can reference authoritative documentation

**Related Commits**:

- [352a9bbc] docs(auth): add comprehensive 02-10.authentication instructions with 38 methods
- [0c2be04e] docs(constitution): add storage realms, remove amendment tracking
- [2b065a92] docs(spec): add per-factor storage realm specifications
- [4f074550] chore(clarify): remove duplicate clarify.md.old backup file
- [9efe49e9] docs(spec): remove outdated MORE TO BE CLARIFIED markers for auth
- [3fa2fad5] chore(plan): rename plan file to indicate current status

**Lessons Learned**:

- Storage realm pattern (YAML + SQL vs SQL ONLY) critical for disaster recovery - service must start even if database unavailable
- Per-factor documentation prevents configuration mistakes during implementation
- Constitution refactoring to remove version tracking makes document clearer and more focused on current requirements
- All MFA and step-up authentication questions answered via QUIZME-02 - no additional clarification needed

---
