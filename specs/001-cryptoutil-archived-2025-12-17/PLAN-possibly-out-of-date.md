# Technical Implementation Plan - Post-Consolidation

**Date**: December 7, 2025 (Updated: December 11, 2025)
**Context**: Updated implementation plan after consolidating iteration files and gap analysis
**Status**: ✅ PLAN COMPLETE - 6-phase approach with clear priorities (Phase 0-3 required, 4-5 optional)

---

## Plan Overview

**Goal**: Complete cryptoutil project to 100% functional state with all 4 products working

**Approach**: 6-phase sequential execution (Phases 0-3 required, 4-5 optional, 6 ongoing)

**Timeline**: 24-32 hours work effort, 5-7 calendar days

**Updates (December 11, 2025)**:

- Added Phase 1.5: Identity Admin API Implementation (HIGH priority from gap analysis)
- Added Phase 2.5: CA Production Deployment (HIGH priority from gap analysis)
- Updated test performance targets based on clarifications
- Added workflow execution time SLAs
- Clarified E2E test scope (product workflows, not just health checks)

### CRITICAL: Execution Mandate

**WORK CONTINUOUSLY until ≥950,000 tokens used OR user says "STOP"**:

- Complete task → immediately start next task
- Push changes → immediately continue working
- Update docs → immediately start next task
- NO stopping to provide summaries
- NO asking for permission between tasks
- NO pausing after git operations

---

## Service Architecture - Dual HTTPS Endpoint Pattern

@
**CRITICAL: ALL services MUST implement dual HTTPS endpoints - NO HTTP PORTS**

### Architecture Requirements

Every service implements two HTTPS endpoints:

1. **Public HTTPS Endpoint** (configurable port, default 8080+)
   - Serves business APIs and browser UI
   - TLS required (never HTTP)
   - TWO security middleware stacks on SAME OpenAPI spec:
     - **Service-to-service APIs**: Require OAuth 2.1 client credentials tokens
     - **Browser-to-service APIs/UI**: Require OAuth 2.1 authorization code + PKCE tokens
   - Middleware enforces authorization boundaries

2. **Private HTTPS Endpoint** (always 127.0.0.1:9090 or similar port)
   - Admin/operations endpoints: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
   - Localhost-only binding (not externally accessible)
   - TLS required (never HTTP)
   - Used by Docker health checks, Kubernetes probes, monitoring

### Service Port Assignments

| Service | Public HTTPS | Private HTTPS | Notes |
|---------|--------------|---------------|-------|
| KMS | :8080 | 127.0.0.1:9090 | Key management APIs |
| Identity AuthZ | :8080 | 127.0.0.1:9090 | OAuth 2.1 endpoints |
| Identity IdP | :8081 | 127.0.0.1:9090 | OIDC endpoints |
| Identity RS | :8082 | 127.0.0.1:9090 | Resource server |
| JOSE | :8080 | 127.0.0.1:9092 | JWK/JWT operations |
| CA | :8443 | 127.0.0.1:9443 | Certificate operations |

### Implementation Checklist

- ✅ NO HTTP endpoints on ANY port
- ✅ Health checks use HTTPS with `--no-check-certificate`
- ✅ Admin endpoints bound to 127.0.0.1 only
- ✅ Public endpoints support both service and browser clients
- ✅ Different OAuth token flows per client type

---

## Phase 0: Optimize Slow Test Packages (Day 1, 4-5 hours) 🚀

**Objective**: Enable fast development feedback loop by optimizing 5 slowest test packages

**Rationale**: Currently 430.9s combined execution time blocks rapid iteration. Optimizing these packages first provides foundation for efficient development in subsequent phases.

### Implementation Strategy

| Package | Current | Target | Strategy | Effort |
|---------|---------|--------|----------|--------|
| `clientauth` | 168s | <30s | Aggressive `t.Parallel()`, split test files by auth method | 2h |
| `jose/server` | 94s | <20s | Parallel subtests, reduce Fiber setup/teardown | 1h |
| `kms/client` | 74s | <20s | Mock KMS dependencies, parallel execution | 1h |
| `jose` | 67s | <15s | Increase coverage 48.8%→95% first, then optimize | 0.5h |
| `kms/server/app` | 28s | <10s | Parallel server tests, dynamic port allocation | 0.5h |

**Success Criteria**:

- All 5 packages execute in <30s each
- Total execution time <100s (down from 430.9s)
- No test failures introduced
- Coverage maintained or improved

---

## Phase 3.5: Server Architecture Unification (Day 4-5, 16-24 hours) 🔴 CRITICAL BLOCKER

**Objective**: Unify Identity, JOSE, and CA servers to match KMS dual-server architecture pattern

**Rationale**: Phase 4 (E2E Tests) and Phase 6 (Demo Videos) are BLOCKED by inconsistent server architectures. Identity has partial implementation (admin servers exist but not integrated), JOSE and CA lack admin servers entirely. This prevents systematic E2E testing and production-quality demos.

### Current State

| Service | Admin Server | Port 9090 | Cmd Integration | Status |
|---------|--------------|-----------|-----------------|--------|
| KMS | ✅ Complete | ✅ Yes | ✅ internal/cmd/cryptoutil | ✅ REFERENCE |
| Identity AuthZ | ✅ Exists | ✅ Yes | ❌ NO | ⚠️ PARTIAL |
| Identity IdP | ✅ Exists | ✅ Yes | ❌ NO | ⚠️ PARTIAL |
| Identity RS | ✅ Exists | ✅ Yes | ❌ NO | ⚠️ PARTIAL |
| JOSE | ❌ Missing | ❌ NO | ❌ NO | ❌ BLOCKED |
| CA | ❌ Missing | ❌ NO | ❌ NO | ❌ BLOCKED |

### Target Architecture

All services follow KMS pattern:

1. **Dual HTTPS Servers**:
   - Public server: 0.0.0.0:<configurable> (APIs + UI)
   - Admin server: 127.0.0.1:9090 (health checks, shutdown)

2. **Unified Command Interface**:
   - `cryptoutil kms|identity|jose|ca <subcommand> [flags]`
   - Implemented in internal/cmd/cryptoutil/<product>/
   - Old standalone binaries (cmd/jose-server, cmd/ca-server) deprecated

3. **Consistent Application Layer**:
   - internal/<product>/server/application.go
   - NewApplication() factory
   - Start()/Stop()/Status()/Health() methods
   - Admin server lifecycle management

### Implementation Tasks

#### P3.5.1: Identity Command Integration (4-6h)

**Current**: Identity has admin servers but NOT in internal/cmd/cryptoutil
**Goal**: `cryptoutil identity start|stop|status|health`

- Create internal/cmd/cryptoutil/identity/ package
- Implement start/stop/status/health subcommands
- Update internal/identity/cmd/main/ to use new structure
- Update cmd/identity-unified/main.go to call internal/cmd/cryptoutil
- Deprecate standalone cmd/identity-compose, cmd/identity-demo
- Update Docker Compose files to use unified command
- Update E2E tests to use unified command

#### P3.5.2: JOSE Admin Server Implementation (6-8h)

**Current**: JOSE has ONLY public server, NO admin endpoints
**Goal**: Dual-server pattern matching KMS

- Create internal/jose/server/admin.go (127.0.0.1:9090)
- Implement /admin/v1/livez, /admin/v1/readyz, /admin/v1/healthz, /admin/v1/shutdown
- Update internal/jose/server/application.go for dual-server lifecycle
- Create internal/cmd/cryptoutil/jose/ package
- Implement start/stop/status/health subcommands
- Update cmd/jose-server/main.go to call internal/cmd/cryptoutil
- Update Docker Compose health checks to use admin endpoints
- Update E2E tests

#### P3.5.3: CA Admin Server Implementation (6-8h)

**Current**: CA has ONLY public server, NO admin endpoints
**Goal**: Dual-server pattern matching KMS

- Create internal/ca/server/admin.go (127.0.0.1:9090)
- Implement /admin/v1/livez, /admin/v1/readyz, /admin/v1/healthz, /admin/v1/shutdown
- Update internal/ca/server/application.go for dual-server lifecycle
- Create internal/cmd/cryptoutil/ca/ package
- Implement start/stop/status/health subcommands
- Update cmd/ca-server/main.go to call internal/cmd/cryptoutil
- Update Docker Compose health checks to use admin endpoints
- Update E2E tests

#### P3.5.4: E2E Test Updates (2-3h)

**Goal**: Update all E2E tests to use unified command interface

- Update internal/test/e2e/ helper functions
- Replace standalone binary calls with cryptoutil commands
- Update service startup patterns
- Verify health check patterns
- Update ci-e2e.yml workflow

#### P3.5.5: Documentation Updates (1-2h)

**Goal**: Document unified architecture

- Update docs/README.md with command examples
- Update runbooks for new command structure
- Document migration from standalone binaries
- Update Docker Compose examples
- Create architecture diagram showing dual-server pattern

### Success Criteria

- ✅ All services accessible via `cryptoutil <product> <subcommand>`
- ✅ All services have admin servers on 127.0.0.1:9090
- ✅ All Docker Compose files use admin health checks
- ✅ All E2E tests use unified command interface
- ✅ 98% test coverage for new cmd packages
- ✅ All CI/CD workflows passing with new architecture

### Dependencies

**Blocked By**: Phase 3 (Coverage Targets) - need baseline quality
**Blocks**: Phase 4 (E2E Tests), Phase 6 (Demo Videos)

---

## Phase 1: Fix CI/CD Workflows (Day 6-7, 4-5 hours) ⚠️

**Objective**: Achieve 11/11 workflow pass rate (currently 3/11 passing, 27%)

**Rationale**: CI/CD reliability is critical for continuous integration and deployment confidence.

### Failing Workflows (8 total)

| Workflow | Current | Root Cause | Fix Strategy | Effort |
|----------|---------|------------|--------------|--------|
| ci-dast | Failing | Service connectivity issues | Fix health check timing, HTTPS endpoints | 1h |
| ci-e2e | Failing | Docker Compose setup | Fix service dependencies, wait patterns | 1h |
| ci-load | Failing | Gatling test configuration | Update load test scenarios | 0.5h |
| ci-coverage | Failing | Test execution issues | Fix coverage aggregation | 0.5h |
| ci-race | Failing | Race condition detection | Fix identified race conditions | 1h |
| ci-benchmark | Failing | Benchmark test failures | Update benchmark baselines | 0.5h |
| ci-fuzz | Failing | Fuzz test configuration | Fix fuzz test execution | 0.5h |
| (1 more) | Failing | TBD from analysis | TBD | 0.5h |

**Success Criteria**:

- All 11 workflows passing (100% pass rate)
- CI feedback loop <10 minutes for critical path (quality + coverage + race)
- Full suite <60 minutes total execution
- No flaky tests

### Test Performance SLAs

**Unit/Integration Tests** (`go test ./...`):

- Per package: <30 seconds
- Total suite: <100 seconds
- Measurement: Wall clock time

**Race Detector** (`go test -race ./...`):

- Per package: <60 seconds (2x overhead typical from CGO)
- Total suite: <200 seconds
- Justification: CGO_ENABLED=1 adds 50-100% overhead

**Mutation Testing** (`gremlins unleash`):

- Per package: Varies by complexity
- Total suite: <45 minutes
- Strategy: Separate workflow, not critical path

**Load Testing** (Gatling):

- Per simulation: 5-10 minutes
- Total suite: <30 minutes
- Strategy: Separate workflow, PR approval trigger

---

## Phase 1.5: Identity Admin API Implementation (NEW - HIGH) 🏗️

**Objective**: Implement dual-server architecture for Identity services (AuthZ, IdP, RS) matching KMS pattern

**Rationale**: Identity currently uses single public server with `/health`. Spec requires private admin server on 127.0.0.1:9090 for health probes, matching KMS/JOSE/CA pattern.

### Implementation Strategy (8-10 hours)

**Step 1: Create Private Server Infrastructure** (3 hours)

- Add `PrivateTLSServer` struct to each service (AuthZ, IdP, RS)
- Implement admin endpoint handlers: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`, `/admin/v1/shutdown`
- Configure TLS for private server (self-signed cert for localhost)
- Bind to `127.0.0.1:9090` (not externally accessible)

**Step 2: Update Server Startup Logic** (2 hours)

- Modify `StartAuthZServer()`, `StartIDPServer()`, `StartRSServer()` to launch both servers
- Public server: 0.0.0.0:8080 (AuthZ), :8081 (IdP), :8082 (RS)
- Private server: 127.0.0.1:9090 for all 3 services
- Add graceful shutdown coordination

**Step 3: Update Tests** (2 hours)

- Migrate integration tests from `/health` to `/admin/v1/healthz`
- Update `internal/identity/**/integration_test.go` files
- Add admin endpoint test coverage
- Verify health check responses (200 OK, JSON body)

**Step 4: Update Deployments** (2 hours)

- Update Docker Compose health checks: `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/v1/livez`
- Update `deployments/identity/compose.yml`, `compose.advanced.yml`, `compose.simple.yml`
- Update GitHub workflows `ci-dast.yml`, `ci-e2e.yml` to use admin endpoints
- Verify Docker health checks pass locally

**Step 5: Backward Compatibility (1 hour)**

- Keep public `/health` endpoint for backward compatibility (optional deprecation later)
- Update documentation to recommend admin endpoints
- Add migration notes to README

**Success Criteria**:

- ✅ Each Identity service (AuthZ, IdP, RS) has private admin server on 127.0.0.1:9090
- ✅ Admin endpoints return 200 OK: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/healthz`
- ✅ Docker Compose health checks use admin endpoints and pass
- ✅ All integration tests pass using new admin endpoints
- ✅ GitHub workflows updated and passing

---

## Phase 2: Complete Deferred I2 Features (Day 2 + 4, 8-10 hours) 🔧

**Objective**: Finish ALL 8 deferred Iteration 2 features (EST serverkeygen MANDATORY)

### Day 2: JOSE E2E Tests (3-4 hours)

**Scope**: Comprehensive E2E test suite for JOSE Authority service

**Implementation**:

- Create `internal/jose/server/*_integration_test.go` files
- Test all 10 JOSE API endpoints end-to-end
- Validate JWK generation, JWKS endpoints, JWS sign/verify, JWE encrypt/decrypt, JWT issue/validate
- Integration with Docker Compose deployment

**Success Criteria**:

- All 10 JOSE endpoints tested end-to-end
- Tests run in <2 minutes total
- Coverage >95% for JOSE server package

### Day 4: CA OCSP + Docker Integration (3-4 hours)

**CA OCSP Responder** (2 hours):

- Implement OCSP responder endpoint `/ca/v1/ocsp`
- Support RFC 6960 OCSP protocol
- Return certificate status (good, revoked, unknown)
- Integration with CRL generation

**EST Serverkeygen** (2 hours) - MANDATORY:

- Research and integrate CMS/PKCS#7 library (github.com/github/smimesign or similar)
- Implement `/ca/v1/est/serverkeygen` endpoint per RFC 7030
- Generate key pair server-side, wrap private key in PKCS#7/CMS
- Return encrypted private key and certificate to client
- E2E tests for serverkeygen flow
- Update SPECKIT-PROGRESS.md I3.1.4 status ⚠️ → ✅

**JOSE Docker Integration** (1-2 hours):

- Add `jose-sqlite`, `jose-postgres-1`, `jose-postgres-2` services to Docker Compose
- Configure ports (8080-8082 public, 9090 admin)
- Health checks via wget
- Integration with common infrastructure (postgres, otel-collector)

**Success Criteria**:

- OCSP responder returns valid responses
- EST serverkeygen endpoint operational with CMS encryption
- JOSE services start and pass health checks
- Docker Compose deployment working end-to-end

---

## Phase 2.5: CA Production Deployment (NEW - HIGH) 🚀

**Objective**: Create production-ready Docker Compose configuration for CA services matching JOSE/KMS patterns

**Rationale**: CA only has `compose.simple.yml` (dev-only). Need multi-instance PostgreSQL deployment for production readiness.

### Implementation Strategy (4-6 hours)

**Step 1: Create Production Compose File** (2 hours)

- Create `deployments/ca/compose.yml` based on `deployments/kms/compose.yml` template
- Define 3 CA instances:
  - `ca-sqlite`: Development instance (SQLite in-memory, port 8443)
  - `ca-postgres-1`: Production instance 1 (PostgreSQL, port 8444)
  - `ca-postgres-2`: Production instance 2 (PostgreSQL, port 8445)
- Each instance binds admin port 9443 internally (127.0.0.1 only)

**Step 2: Configure PostgreSQL Backend** (1 hour)

- Add PostgreSQL service definition
- Configure database initialization scripts
- Add secrets for database credentials (file-based, not environment variables)
- Add health checks for PostgreSQL service

**Step 3: Integrate Telemetry Services** (1 hour)

- Include OpenTelemetry collector configuration
- Add grafana-otel-lgtm for observability
- Configure OTLP exporters in CA service configs
- Add collector health checks

**Step 4: Configure CA-Specific Requirements** (1 hour)

- Add CRL distribution volume mounts
- Configure OCSP responder endpoints
- Add EST endpoint configurations
- Configure certificate profile directories

**Step 5: Testing and Validation** (1 hour)

- Test `docker compose up -d` startup
- Verify all 3 CA instances start successfully
- Test health checks pass (admin endpoints)
- Verify certificate issuance works on all 3 instances
- Test CRL generation and OCSP responses

**Success Criteria**:

- ✅ `deployments/ca/compose.yml` exists and is production-ready
- ✅ 3 CA instances (sqlite, postgres-1, postgres-2) start successfully
- ✅ Health checks pass: `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9443/admin/v1/livez`
- ✅ PostgreSQL backend works for CA instances
- ✅ Telemetry integration functional (OTLP, Grafana)
- ✅ CRL distribution and OCSP responder operational

---

## Phase 3: Achieve Coverage Targets (Day 5, 2-3 hours) 📊

**Objective**: Get 5/5 packages to ≥95% coverage

### Critical Gaps (2 packages, 2 hours)

| Package | Current | Gap | Strategy | Effort |
|---------|---------|-----|----------|--------|
| `ca/handler` | 47.2% | -47.8% | Add handler tests for all endpoints | 1h |
| `auth/userauth` | 42.6% | -52.4% | Add authentication flow tests | 1h |

### Secondary Targets (2 packages, 1 hour)

| Package | Current | Gap | Strategy | Effort |
|---------|---------|-----|----------|--------|
| `unsealkeysservice` | 78.2% | -16.8% | Add edge case tests | 0.5h |
| `network` | 88.7% | -6.3% | Add error path tests | 0.5h |

**Success Criteria**:

- All 5 packages ≥95% coverage
- No coverage regressions in other packages
- Tests use `t.Parallel()` and UUIDv7 patterns

---

## Phase 4: Advanced Testing & E2E Workflows (Upgraded Priority - HIGH) 🧪

**Objective**: Add comprehensive E2E workflow tests and advanced testing methodologies

**Rationale**: Current E2E tests only validate Docker health checks. Need end-to-end product workflow validation and load test coverage for Browser API.

### E2E Workflow Tests (4-6 hours, HIGH PRIORITY)

**Location**: Extend `internal/test/e2e/e2e_test.go`

**Minimum Viable E2E Workflows**:

1. **OAuth 2.1 Authorization Code Flow** (2 hours)
   - Browser → AuthZ `/oauth2/v1/authorize` → redirect to IdP `/oidc/v1/login`
   - User login → `/oidc/v1/consent` → consent granted
   - Redirect back to AuthZ with code → `/oauth2/v1/token` with PKCE verifier
   - Validate access token, ID token, refresh token returned
   - Test token introspection and revocation

2. **KMS Encrypt/Decrypt Workflow** (1 hour)
   - Create ElasticKey → generate MaterialKey
   - Encrypt plaintext → decrypt ciphertext
   - Verify plaintext matches original
   - Test key rotation (generate new MaterialKey, decrypt with old key fails)

3. **CA Certificate Lifecycle** (2 hours)
   - Generate CSR → submit to `/ca/v1/certificate`
   - Receive issued certificate → validate chain
   - Revoke certificate → verify OCSP revoked status
   - Check CRL contains revoked certificate

4. **JOSE JWT Sign/Verify** (1 hour)
   - Generate JWK → sign JWT with claims
   - Verify JWT signature → validate claims
   - Test token expiration and invalid signature rejection

**Success Criteria**:

- All 4 E2E workflows pass in CI/CD
- Tests use real Docker stack (PostgreSQL, all services)
- Execution time: <10 minutes total
- Coverage of critical user journeys

### Load Testing Expansion (3-4 hours, HIGH PRIORITY)

**Browser API Gatling Simulation** (3 hours):

- Create `test/load/src/test/java/cryptoutil/BrowserApiSimulation.java`
- Test OAuth authorization code flow under load
- Test certificate request workflows
- Test UI endpoints (`/ui/swagger/doc.json`, health checks)
- Simulate 50-100 concurrent users
- Measure response times (p95 <500ms target)

**Admin API Load Test** (DEFER - LOW PRIORITY):

- Admin endpoints are operational, not high-throughput
- Defer unless specific monitoring concerns arise

**Success Criteria**:

- Browser API load test exists and passes
- Performance baselines established (p95, p99 latency)
- Load test runs in <10 minutes

### Advanced Testing Methodologies (4-6 hours, OPTIONAL)

**Benchmark Tests** (2 hours):

- Add `_bench_test.go` files to cryptographic packages
- Benchmark key generation, encryption, signing operations
- Establish performance baselines

**Fuzz Tests Expansion** (2 hours):

- Add `_fuzz_test.go` files to parsers and validators
- Fuzz JWT parsing, certificate parsing, OAuth token introspection
- Run for 15s minimum per test

**Property-Based Tests** (2 hours):

- Add `_property_test.go` files using gopter
- Test round-trip properties (encrypt→decrypt, sign→verify)
- Validate cryptographic invariants

**Success Criteria**:

- Benchmarks established for all crypto operations
- Fuzz tests run without failures
- Property tests validate core invariants

---

## Phase 5: Documentation & Demo (Optional, 8-12 hours) 📹

**Objective**: Create demo videos showcasing each product

### Demo Video Plan

| Demo | Duration | Content | Effort |
|------|----------|---------|--------|
| JOSE Authority | 5-10 min | JWK generation, signing, encryption | 2h |
| Identity Server | 10-15 min | OAuth flow, OIDC, MFA | 2-3h |
| KMS | 10-15 min | Key hierarchy, encryption, rotation | 2-3h |
| CA Server | 10-15 min | Certificate issuance, revocation, OCSP | 2-3h |
| Integration | 15-20 min | All 4 products working together | 3-4h |
| Unified Suite | 20-30 min | Complete deployment walkthrough | 3-4h |

**Success Criteria**:

- All 6 demo videos recorded and published
- Demos show working functionality
- Narration explains architecture and features

---

## Success Metrics

### Required for Completion (Phases 0-3 + NEW)

- ✅ Phase 0: All 5 slow packages <30s execution, total <100s
- ✅ Phase 1: 11/11 CI/CD workflows passing, feedback loop <10 min
- ✅ **Phase 1.5: Identity admin API implemented (dual-server pattern)**
- ✅ Phase 2: 7/8 deferred features complete (EST serverkeygen optional if PKCS#7 blocked)
- ✅ **Phase 2.5: CA production deployment (multi-instance PostgreSQL)**
- ✅ Phase 3: 5/5 packages ≥95% coverage

### High Priority Enhancements (Phase 4)

- ✅ **E2E workflow tests (OAuth, KMS, CA, JOSE) - 4 workflows minimum**
- ✅ **Browser API load testing (Gatling simulation)**
- ⚠️ Advanced testing methodologies (benchmarks, fuzz expansion, property tests)

### Optional Enhancements (Phase 5)

- ⚠️ Phase 5: Demo videos created (6 videos)

### Quality Gates

- All linting passes (`golangci-lint run`)
- All tests pass (`go test ./... -cover -shuffle=on`)
- Race detector passes (`go test -race ./...` <200s total)
- No CRITICAL/HIGH security vulnerabilities
- Integration demos work (`go run ./cmd/demo all`)
- Docker Compose deployments work for all 4 products

---

## Risk Management

| Risk | Impact | Mitigation | Status |
|------|--------|------------|--------|
| Test optimization breaks tests | HIGH | Incremental changes, verify after each package | Ongoing |
| CI/CD fixes introduce flakiness | MEDIUM | Use health checks, exponential backoff patterns | Ongoing |
| Coverage improvements reduce quality | LOW | Require meaningful tests, not just coverage numbers | Ongoing |
| EST serverkeygen needs CMS library | MEDIUM | Research github.com/github/smimesign or similar, MANDATORY | Planned |
| **Identity admin API breaks deployments** | HIGH | Keep `/health` for backward compat, staged migration | NEW |
| **CA deployment complexity** | MEDIUM | Copy KMS pattern, test incrementally | NEW |
| **E2E tests require full stack** | MEDIUM | Use Docker Compose, ensure PostgreSQL available | NEW |

---

## Dependencies

**External**:

- CMS/PKCS#7 library (MANDATORY for EST serverkeygen)
- Java 21 LTS (for Gatling load tests)
- Docker Compose v2+ (for deployments)
- PostgreSQL 18+ (for production backends)

**Internal**:

- Phase 0 must complete before Phases 1-3 (foundation)
- Phase 1.5 depends on Phase 1 (CI/CD stable first)
- Phase 2 (JOSE E2E) before Phase 1 (some CI workflows depend on it)
- Phase 2.5 depends on Phase 2 (CA features complete first)
- Phase 3 can run in parallel with Phases 1-2
- Phase 4 only after Phases 0-3 complete
- Phase 5 only after all other phases complete

**Critical Path** (for minimum viable completion):

Phase 0 → Phase 1 → Phase 1.5 → Phase 2 → Phase 2.5 → Phase 3 → Phase 4 (E2E only)

**Estimated Total**: 24-32 hours work effort, 5-7 calendar days

---

## Conclusion

**Plan Status**: ✅ **COMPLETE AND EXECUTABLE**

This plan provides:

- Clear 5-phase sequential approach
- Effort estimates per task
- Success criteria per phase
- Risk mitigation strategies
- Dependency management

**Next Step**: Execute /speckit.tasks to generate detailed task breakdown.

---

*Technical Implementation Plan Version: 2.0.0*
*Author: GitHub Copilot (Claude Sonnet 4.5)*
*Created: December 7, 2025*
*Updated: December 11, 2025*

**Update Summary (v2.0.0)**:

- Added Phase 1.5: Identity Admin API Implementation (8-10 hours)
- Added Phase 2.5: CA Production Deployment (4-6 hours)
- Upgraded Phase 4 to HIGH priority with E2E workflows and Browser API load tests
- Added test performance SLAs (unit <100s, race <200s, mutation <45min)
- Updated success metrics and risk management
- Increased total effort estimate: 24-32 hours (from 16-24 hours)
- Increased timeline: 5-7 calendar days (from 3-5 days)
