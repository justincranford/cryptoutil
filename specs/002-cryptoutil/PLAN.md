# cryptoutil Implementation Plan

**Version**: 2.0.0
**Date**: December 19, 2025
**Status**: Active

---

## Document Purpose

This plan provides a comprehensive 7-phase implementation approach for cryptoutil, a Go-based cryptographic services platform consisting of 4 independent products (JOSE, Identity, KMS, CA) that can be deployed standalone or as an integrated suite.

**Key Principles**:

- **Phase-based execution**: Strict dependencies between phases (Foundation → Core → Advanced)
- **Evidence-based completion**: Objective proof required (coverage ≥95%, mutation ≥85%, all tests passing)
- **Iterative refinement**: Spec Kit mini-cycles every 3-5 tasks update documents based on implementation insights
- **Constitution authority**: `.specify/memory/constitution.md` defines immutable principles and gates

---

## Document Authority

**Source Documents** (in precedence order):

1. **constitution.md** - Immutable principles, absolute requirements, quality gates
2. **spec.md** - Product requirements, technical specifications
3. **clarify.md** - Authoritative Q&A for resolved ambiguities
4. **plan.md** (this file) - Implementation approach and task breakdown

**Living Document Status**:

- Spec Kit methodology treats ALL documents as living (constitution, spec, clarify, plan)
- Implementation insights trigger document updates via mini-cycle feedback loops
- DETAILED.md Section 2 timeline tracks all document evolution decisions

---

## Table of Contents

1. [Overview](#overview)
2. [Phase 0: Foundation (Complete)](#phase-0-foundation-complete)
3. [Phase 1: Core Infrastructure](#phase-1-core-infrastructure)
4. [Phase 2: Service Completion](#phase-2-service-completion)
5. [Phase 3: Advanced Features](#phase-3-advanced-features)
6. [Phase 4: Quality Gates](#phase-4-quality-gates)
7. [Phase 5: Production Hardening](#phase-5-production-hardening)
8. [Phase 6: Service Template](#phase-6-service-template)
9. [Phase 7: Learn-PS Demonstration](#phase-7-learn-ps-demonstration)
10. [Dependencies and Sequencing](#dependencies-and-sequencing)
11. [Risk Management](#risk-management)
12. [Success Criteria](#success-criteria)

---

## Overview

### Product Suite Architecture

cryptoutil delivers 4 independent products deployable standalone or unified:

| Product | Status | Public Port | Admin Port | Description |
|---------|--------|-------------|------------|-------------|
| **P1: JOSE** | ✅ Complete | 9443-9449 | 127.0.0.1:9090 | JSON Object Signing/Encryption Authority |
| **P2: Identity** | ⚠️ Partial | 18000-18409 | 127.0.0.1:9090 | OAuth 2.1 AuthZ + OIDC IdP (5 services) |
| **P3: KMS** | ✅ Complete | 8080-8089 | 127.0.0.1:9090 | Hierarchical Key Management Service |
| **P4: CA** | ✅ Complete | 8443-8449 | 127.0.0.1:9090 | X.509 Certificate Authority |

### Core Requirements

**CGO Ban** (CRITICAL):

- CGO_ENABLED=0 MANDATORY for builds, tests, Docker, production
- ONLY EXCEPTION: Race detector workflow (Go toolchain limitation)
- Use CGO-free alternatives (modernc.org/sqlite, NOT github.com/mattn/go-sqlite3)

**FIPS 140-3 Compliance** (CRITICAL):

- ALWAYS enabled, NEVER disabled
- Approved algorithms: RSA ≥2048, AES ≥128, ECDSA (P-256/384/521), EdDSA, PBKDF2-HMAC-SHA256
- BANNED algorithms: bcrypt, scrypt, Argon2, MD5, SHA-1, DES, 3DES

**Dual HTTPS Endpoint Pattern** (MANDATORY):

- **Public HTTPS** (0.0.0.0:configurable): Browser APIs (/browser/*) + Service APIs (/service/*)
- **Private HTTPS** (127.0.0.1:909X): Admin APIs (/admin/v1/livez, readyz, healthz, shutdown)
- Admin port assignments: KMS=9090, Identity=9091, CA=9092, JOSE=9093

**Testing Requirements** (CRITICAL):

- ALWAYS concurrent: `go test ./... -shuffle=on` (NEVER `-p=1` or `-parallel=1`)
- Coverage targets: 95% production, 100% infrastructure, 100% utility
- Mutation testing: ≥85% Phase 4, ≥98% Phase 5+
- Test timing: <15s per unit test package, <180s total unit suite
- Real dependencies preferred (PostgreSQL containers, real crypto, real HTTPS servers)

---

## Phase 0: Foundation (Complete)

**Status**: ✅ Complete
**Duration**: Completed
**Goal**: Establish project infrastructure, build system, and core utilities

### Completed Deliverables

#### Build and Development Infrastructure

- [x] Go 1.25.5+ toolchain configuration
- [x] golangci-lint v2.6.2+ configuration (.golangci.yml)
- [x] Pre-commit hooks (.pre-commit-config.yaml)
- [x] Docker multi-stage builds with static linking
- [x] Docker Compose deployments (compose.yml)
- [x] Makefile or equivalent build automation

#### CI/CD Workflows

- [x] ci-quality (linting, formatting, build validation)
- [x] ci-coverage (test coverage analysis)
- [x] ci-race (race condition detection with CGO_ENABLED=1)
- [x] ci-mutation (gremlins mutation testing)
- [x] ci-benchmark (performance benchmarks)
- [x] ci-fuzz (fuzz testing for parsers/validators)
- [x] ci-sast (static security analysis with gosec)
- [x] ci-gitleaks (secrets scanning)
- [x] ci-dast (dynamic security testing with Nuclei/ZAP)
- [x] ci-e2e (end-to-end Docker Compose tests)
- [x] ci-load (Gatling load testing - Service API only)

#### Project Structure

- [x] Standard Go Project Layout (cmd, internal, pkg, configs, deployments)
- [x] Magic constants organization (internal/shared/magic)
- [x] Error handling patterns and utilities
- [x] Configuration management (YAML-based)
- [x] Logging and observability foundation

#### Documentation

- [x] README.md (main project documentation)
- [x] docs/README.md (deep dive documentation)
- [x] docs/DEV-SETUP.md (development environment setup)
- [x] .github/copilot-instructions.md (LLM agent directives)
- [x] .github/instructions/*.instructions.md (19 instruction files)

---

## Phase 1: Core Infrastructure

**Status**: ⚠️ In Progress
**Priority**: HIGHEST
**Dependencies**: Phase 0 complete
**Goal**: Complete shared infrastructure used by all 4 products

### 1.1 Shared Crypto Library (internal/shared/crypto)

**Status**: ✅ Complete

**Deliverables**:

- [x] FIPS 140-3 compliant algorithm implementations
- [x] Key generation pools (keygen package) for RSA, ECDSA, ECDH, EdDSA, AES, HMAC
- [x] JWK/JWKS/JWE/JWS/JWT primitives (internal/jose)
- [x] Deterministic key derivation (HKDF) for interoperability
- [x] Certificate chain validation (TLS 1.3+, full chain, NEVER InsecureSkipVerify)
- [x] Secure random number generation (crypto/rand, NEVER math/rand)

**Quality Gates**:

- [x] 95%+ test coverage
- [x] 85%+ mutation efficacy (gremlins)
- [x] Fuzz tests for all parsers and validators
- [x] Benchmark tests for crypto operations

**Evidence**: All crypto packages pass CI/CD, coverage reports in test-output/

---

### 1.2 Database Repository Layer (internal/shared/repository)

**Status**: ✅ Complete

**Deliverables**:

- [x] PostgreSQL support (production/development/testing)
- [x] SQLite support (development/testing with modernc.org/sqlite)
- [x] GORM ORM integration
- [x] Schema migrations (golang-migrate with embedded SQL)
- [x] Connection pooling (PostgreSQL max 50, SQLite max 1)
- [x] Transaction context pattern (ctx.Value for tx propagation)
- [x] WAL mode and busy_timeout for SQLite concurrency

**Quality Gates**:

- [x] 100%+ test coverage (infrastructure requirement)
- [x] PostgreSQL and SQLite dual testing
- [x] Concurrent test execution with t.Parallel()
- [x] Integration tests with test containers

**Evidence**: All repository tests pass, no CGO dependencies, PostgreSQL service in workflows

---

### 1.3 Telemetry Infrastructure (internal/shared/telemetry)

**Status**: ✅ Complete

**Deliverables**:

- [x] OpenTelemetry instrumentation (traces, metrics, logs)
- [x] OTLP export configuration (gRPC:4317, HTTP:4318)
- [x] Structured logging with context propagation
- [x] Health endpoint telemetry (/admin/v1/livez, readyz, healthz)
- [x] Grafana dashboards (Loki, Tempo, Prometheus)
- [x] Docker Compose telemetry stack (otel-collector, grafana-otel-lgtm)

**Quality Gates**:

- [x] 100%+ test coverage
- [x] Telemetry forwarding validated (services → otel-collector → Grafana)
- [x] No sensitive data in logs (PII, passwords, keys)
- [x] Proper log levels (debug, info, warn, error)

**Evidence**: Telemetry working in Docker Compose, Grafana dashboards accessible

---

### 1.4 Configuration Management (internal/shared/config)

**Status**: ✅ Complete

**Deliverables**:

- [x] YAML-based configuration (NO environment variables for secrets)
- [x] CLI flag support for overrides
- [x] Docker secrets integration (file:///run/secrets/*)
- [x] Kubernetes secrets support (mounted files)
- [x] Configuration validation on startup
- [x] Feature flags framework

**Quality Gates**:

- [x] 100%+ test coverage
- [x] Validation errors prevent startup
- [x] Default config supports dev mode (no required settings)
- [x] No hardcoded secrets in code or configs

**Evidence**: All services start with valid YAML configs, secrets mounted correctly

---

### 1.5 Networking and HTTP Framework (internal/shared/server)

**Status**: ⚠️ Partial (KMS complete, Identity/JOSE/CA need migration)

**Deliverables**:

- [x] Dual HTTPS server pattern (public + admin)
- [x] TLS 1.3+ minimum (NEVER plain HTTP)
- [x] Dynamic and static TLS certificate support
- [x] HTTP/2 support via Fiber framework
- [x] Graceful shutdown on /admin/v1/shutdown
- [x] Health check endpoints (livez, readyz, healthz)
- [x] Middleware pipeline (CORS, CSRF, CSP, rate limiting, IP allowlist)

**Quality Gates**:

- [x] 95%+ test coverage
- [x] TLS handshake tests with real certificates
- [x] Health check retry logic tested
- [x] Graceful shutdown validated (no interrupted requests)

**Evidence**: KMS reference implementation complete, Docker health checks working

**Blockers**:

- Identity services need admin server migration (currently use /health on public port)
- JOSE needs admin server implementation
- CA needs admin server implementation

---

## Phase 2: Service Completion

**Status**: ⚠️ In Progress
**Priority**: HIGH
**Dependencies**: Phase 1 complete
**Goal**: Complete all 4 products with dual-server architecture and unified command interface

### 2.1 JOSE Admin Server Implementation

**Status**: IN PROGRESS
**Estimate**: 2-3 days
**Dependencies**: Phase 1.5 complete

**Tasks**:

- [ ] Create internal/jose/server/admin package
- [ ] Implement private HTTPS server on port 9093
- [ ] Add /admin/v1/livez, readyz, healthz, shutdown endpoints
- [ ] Update Docker Compose health checks to use admin endpoints
- [ ] Integrate with cmd/cryptoutil/jose subcommand
- [ ] Update all JOSE test files to use dual-server pattern

**Quality Gates**:

- [ ] 95%+ test coverage for admin server code
- [ ] Docker health checks pass (wget --no-check-certificate <https://127.0.0.1:9093/admin/v1/livez>)
- [ ] `cryptoutil jose start` command working
- [ ] All JOSE tests passing with new architecture

**Evidence**: JOSE service starts with dual servers, health checks working, cryptoutil command functional

---

### 2.2 Identity Admin Server Migration

**Status**: ⚠️ Partial
**Estimate**: 3-4 days
**Dependencies**: Phase 1.5 complete

**Tasks**:

- [ ] Migrate identity-authz to dual-server pattern (public 8180, admin 9091)
- [ ] Migrate identity-idp to dual-server pattern (public 8181, admin 9091)
- [ ] Migrate identity-rs to dual-server pattern (public 8182, admin 9091)
- [ ] Integrate with cmd/cryptoutil/identity subcommand
- [ ] Update Docker Compose health checks to use admin endpoints
- [ ] Update all Identity test files to use dual-server pattern

**Quality Gates**:

- [ ] 95%+ test coverage for admin server code
- [ ] All 3 Identity services use admin port 9091
- [ ] Docker health checks pass for all Identity services
- [ ] `cryptoutil identity start` command working
- [ ] All Identity tests passing with new architecture

**Evidence**: Identity services start with dual servers, health checks working, cryptoutil command functional

---

### 2.3 CA Admin Server Migration

**Status**: ❌ Blocked
**Estimate**: 2-3 days
**Dependencies**: Phase 1.5 complete

**Tasks**:

- [ ] Migrate CA service to dual-server pattern (public 8380, admin 9092)
- [ ] Implement /admin/v1/livez, readyz, healthz, shutdown endpoints
- [ ] Integrate with cmd/cryptoutil/ca subcommand
- [ ] Update Docker Compose health checks to use admin endpoints
- [ ] Update all CA test files to use dual-server pattern

**Quality Gates**:

- [ ] 95%+ test coverage for admin server code
- [ ] Docker health checks pass (wget --no-check-certificate <https://127.0.0.1:9092/admin/v1/livez>)
- [ ] `cryptoutil ca start` command working
- [ ] All CA tests passing with new architecture

**Evidence**: CA service starts with dual servers, health checks working, cryptoutil command functional

---

### 2.4 Service Federation Configuration

**Status**: ❌ Not Started
**Estimate**: 3-4 days
**Dependencies**: Phase 2.1, 2.2, 2.3 complete

**Tasks**:

- [ ] Add federation section to each service YAML config
- [ ] Implement graceful degradation for unavailable federated services
- [ ] Add periodic retry with exponential backoff for federated service discovery
- [ ] Add federation status to /admin/v1/readyz endpoint
- [ ] Document federation patterns in docs/README.md

**Configuration Example** (kms.yml):

```yaml
federation:
  enabled: true
  identity:
    authz_url: https://identity-authz:8180
    idp_url: https://identity-idp:8181
    retry:
      initial_backoff: 1s
      max_backoff: 60s
      max_retries: 10
  jose:
    ja_url: https://jose-ja:8280
    retry:
      initial_backoff: 1s
      max_backoff: 60s
      max_retries: 10
```

**Quality Gates**:

- [ ] 95%+ test coverage for federation code
- [ ] Services start successfully when federated services unavailable
- [ ] Federated features auto-enable when dependencies become available
- [ ] Federation status visible in /admin/v1/readyz response

**Evidence**: All services start in standalone and federated modes, graceful degradation working

---

### 2.5 Unified Command Interface

**Status**: ⚠️ Partial (KMS complete)
**Estimate**: 2-3 days
**Dependencies**: Phase 2.1, 2.2, 2.3 complete

**Tasks**:

- [ ] Complete cmd/cryptoutil/jose subcommand integration
- [ ] Complete cmd/cryptoutil/identity subcommand integration
- [ ] Complete cmd/cryptoutil/ca subcommand integration
- [ ] Add `cryptoutil status` command (checks all running services)
- [ ] Add `cryptoutil version` command (shows version, Go version, build date)
- [ ] Update documentation with command usage examples

**Command Examples**:

```bash
# Start services
cryptoutil kms start --config=configs/kms/kms-postgres-1.yml
cryptoutil identity start --config=configs/identity/authz.yml
cryptoutil jose start --config=configs/jose/jose-sqlite.yml
cryptoutil ca start --config=configs/ca/ca-postgres-1.yml

# Check status
cryptoutil status
cryptoutil kms status
cryptoutil identity status

# Version info
cryptoutil version
```

**Quality Gates**:

- [ ] All 4 products accessible via cryptoutil command
- [ ] Status commands query /admin/v1/readyz endpoints
- [ ] 95%+ test coverage for command handlers
- [ ] Help text accurate and complete

**Evidence**: All cryptoutil commands working, status queries functional, help text complete

---

## Phase 3: Advanced Features

**Status**: ⚠️ In Progress
**Priority**: MEDIUM
**Dependencies**: Phase 2 complete
**Goal**: Implement advanced authentication, authorization, and cryptographic features

### 3.1 Advanced MFA Factors (Identity)

**Status**: ✅ Complete

**Deliverables**:

- [x] Hardware Security Keys (U2F/FIDO via WebAuthn)
- [x] Push Notifications (mobile app approval tokens)
- [x] Phone Call OTP (voice call delivery)
- [x] Email OTP (with rate limiting 5/10min)
- [x] SMS OTP (NIST deprecated but mandatory)
- [x] HOTP (HMAC-based OTP with counter sync)
- [x] Recovery Codes (10-code single-use)

**Quality Gates**:

- [x] 95%+ test coverage for each MFA factor
- [x] Integration tests with mock providers (email, SMS, push)
- [x] Rate limiting tests (5 OTPs per 10 minutes)
- [x] Replay attack prevention validated

**Evidence**: All MFA factors implemented with 6-15 tests each, mock providers working

---

### 3.2 Advanced Client Authentication (Identity)

**Status**: ✅ Complete

**Deliverables**:

- [x] client_secret_jwt (RFC 7523 Section 3 with jti replay protection)
- [x] private_key_jwt (RFC 7523 Section 3 with JWKS support)
- [x] tls_client_auth (mutual TLS with CA validation)
- [x] self_signed_tls_client_auth (self-signed cert validation)
- [x] session_cookie (browser session for SPA)

**Quality Gates**:

- [x] 95%+ test coverage for each auth method
- [x] JTI replay protection validated (10-minute assertion lifetime)
- [x] Certificate revocation checking tests
- [x] SHA-256 fingerprint verification tests

**Evidence**: All auth methods implemented with 6-11 tests each, RFC 7523 compliance validated

---

### 3.3 OAuth 2.1 Advanced Flows (Identity)

**Status**: ✅ Complete

**Deliverables**:

- [x] Device Authorization Grant (RFC 8628 - 18 tests passing)
- [x] Pushed Authorization Requests (RFC 9126 - 16 tests passing)
- [x] Token Introspection (RFC 7662)
- [x] Token Revocation (RFC 7009)
- [x] Client Secret Rotation (with grace period)

**Quality Gates**:

- [x] 95%+ test coverage for each flow
- [x] RFC compliance validated (8628, 9126, 7662, 7009)
- [x] Grace period tests for secret rotation
- [x] Audit trail for rotation events

**Evidence**: All OAuth 2.1 flows implemented, RFC compliance validated, tests passing

---

### 3.4 Hash Service Architecture (Phase 5 - Deferred)

**Status**: ❌ Not Started
**Estimate**: 1-2 weeks
**Dependencies**: Phase 4 complete
**Priority**: DEFERRED to Phase 5

**Deliverables**:

- [ ] Unified hash service architecture (4 registries × 3 versions = 12 configs)
- [ ] LowEntropyRandomHashRegistry (PBKDF2-based, salted)
- [ ] LowEntropyDeterministicHashRegistry (PBKDF2-based, no salt)
- [ ] HighEntropyRandomHashRegistry (HKDF-based, salted)
- [ ] HighEntropyDeterministicHashRegistry (HKDF-based, no salt)
- [ ] Version-based policy management (v1=2020 NIST, v2=2023 NIST, v3=2025 OWASP)
- [ ] Automatic version selection based on input size (0-31→SHA-256, 32-47→SHA-384, 48+→SHA-512)
- [ ] Hash output format with version metadata: `{version}:{algorithm}:{params}:base64_hash`
- [ ] Backward compatibility (verify against all versions until match)

**Configuration Example** (hash-service.yml):

```yaml
hash_service:
  registries:
    password_hashing:  # Low entropy, random (PBKDF2, salted)
      versions:
        - id: 1
          algorithm: PBKDF2-HMAC-SHA256
          params:
            rounds: 600000  # OWASP 2023 recommendation
            salt_size: 16
        - id: 2
          algorithm: PBKDF2-HMAC-SHA384
          params:
            rounds: 600000
            salt_size: 24
    pii_hashing:  # High entropy, deterministic (HKDF, no salt)
      versions:
        - id: 1
          algorithm: HKDF-HMAC-SHA256
          params:
            info: "pii-hash-v1"
        - id: 2
          algorithm: HKDF-HMAC-SHA384
          params:
            info: "pii-hash-v2"
```

**Quality Gates**:

- [ ] 100%+ test coverage (utility code requirement)
- [ ] Version migration tests (hash with v1, verify with v2)
- [ ] FIPS 140-3 compliance validated (PBKDF2, HKDF approved)
- [ ] 98%+ mutation efficacy (Phase 5 requirement)

**Evidence**: All hash registries implemented, version migration working, FIPS compliance validated

---

## Phase 4: Quality Gates

**Status**: ⚠️ In Progress
**Priority**: HIGHEST
**Dependencies**: Phase 3 complete
**Goal**: Achieve 95%+ coverage, 85%+ mutation efficacy, pass all quality gates

### 4.1 Coverage Analysis and Gap Remediation

**Status**: ⚠️ In Progress
**Estimate**: 2-3 weeks

**Tasks**:

- [ ] Generate baseline coverage reports for all packages
- [ ] Analyze coverage HTML for RED (uncovered) lines
- [ ] Create coverage gap remediation plan per package
- [ ] Write targeted tests for identified gaps (NOT blind test addition)
- [ ] Validate coverage improvement per test batch
- [ ] Achieve 95%+ production, 100%+ infrastructure/utility coverage

**Package-Level Tracking** (examples):

| Package | Current Coverage | Target | Status | Gaps |
|---------|------------------|--------|--------|------|
| internal/jose | 84.2% | 95% | ❌ | Unused functions (23%), Is*/Extract* defaults (83-86%) |
| internal/identity/domain | 87.5% | 95% | ❌ | Error paths, edge cases |
| internal/kms/server | 91.2% | 95% | ❌ | Encryption corner cases |
| internal/ca | 88.9% | 95% | ❌ | Certificate validation paths |
| internal/shared/crypto | 96.3% | 95% | ✅ | None |
| internal/shared/repository | 100% | 100% | ✅ | None |
| internal/cmd/cicd | 92.1% | 100% | ❌ | CLI error handling |

**Quality Gates**:

- [ ] NO PACKAGE below 95% (production) or 100% (infrastructure/utility)
- [ ] Coverage reports in test-output/ directory
- [ ] Baseline → HTML analysis → targeted tests → verify improvement cycle documented
- [ ] NO blind test addition (always analyze baseline first)

**Evidence**: Coverage reports showing 95%/100% targets met, test-output/ artifacts committed

---

### 4.2 Mutation Testing Baseline

**Status**: ⚠️ In Progress
**Estimate**: 1-2 weeks

**Tasks**:

- [ ] Run gremlins on all packages (exclude generated code, vendor, tests)
- [ ] Generate baseline mutation reports (≥85% efficacy target for Phase 4)
- [ ] Document efficacy scores in docs/GREMLINS-TRACKING.md
- [ ] Identify low-efficacy packages (<85%)
- [ ] Create mutation gap remediation plan (priority order: API validation → business logic → repository → infrastructure)
- [ ] Add tests to kill surviving mutants
- [ ] Achieve ≥85% efficacy per package

**Configuration** (.gremlins.yaml):

```yaml
threshold:
  efficacy: 85  # Phase 4 target (Phase 5+ raises to 98%)
  mutant-coverage: 90

workers: 4
test-cpu: 2
timeout-coefficient: 2

exclude:
  - "*_test.go"
  - "testdata/*"
  - "vendor/*"
  - "*/api/client/*"  # Generated code
  - "*/api/model/*"   # Generated code
  - "*/api/server/*"  # Generated code
```

**Quality Gates**:

- [ ] ≥85% efficacy per package (Phase 4 requirement)
- [ ] Baseline reports committed to docs/gremlins/
- [ ] Efficacy tracking in docs/GREMLINS-TRACKING.md
- [ ] Windows compatibility validated (v0.7.0+ test)

**Evidence**: Gremlins reports showing ≥85% efficacy, tracking document updated

---

### 4.3 Test Timing Optimization

**Status**: ⚠️ In Progress
**Estimate**: 1-2 weeks

**Tasks**:

- [ ] Measure baseline test execution times per package
- [ ] Identify slow packages (>15s per unit test package)
- [ ] Apply probabilistic execution (TestProbTenth, TestProbQuarter) to slow packages
- [ ] Consolidate redundant table-driven test cases
- [ ] Verify coverage maintained after optimization (<1% drop acceptable if >5s faster)
- [ ] Achieve <15s per unit test package, <180s total unit suite

**Probabilistic Execution Strategy**:

- **TestProbAlways** (100%): Base algorithms (RSA2048, AES256, ES256) - always test
- **TestProbQuarter** (25%): Key size variants (RSA3072, AES192) - statistical sampling
- **TestProbTenth** (10%): Less common variants (RSA4096, AES128) - minimal sampling
- **TestProbNever** (0%): Deprecated or extreme edge cases - skip

**Quality Gates**:

- [ ] All unit test packages <15s (warning if violated, not blocking)
- [ ] Total unit test suite <180s (hard limit)
- [ ] Coverage drop <1% after consolidation
- [ ] Mutation efficacy unchanged after optimization

**Evidence**: Timing reports showing <15s/<180s targets met, coverage maintained

---

### 4.4 Linting and Formatting Compliance

**Status**: ✅ Complete
**Estimate**: Ongoing maintenance

**Tasks**:

- [x] Fix all golangci-lint violations (golangci-lint run --fix)
- [x] Enforce UTF-8 without BOM for all text files
- [x] Apply file size limits (300 soft, 400 medium, 500 hard)
- [x] Validate conventional commit messages
- [x] Run pre-commit hooks on all commits

**Quality Gates**:

- [x] `golangci-lint run` passes with 0 violations
- [x] `go build ./...` passes with 0 errors
- [x] Pre-commit hooks pass (trailing whitespace, markdown lint, etc.)
- [x] All files ≤500 lines (refactor if exceeded)

**Evidence**: CI/CD workflows passing (ci-quality), no lint violations

---

## Phase 5: Production Hardening

**Status**: ❌ Not Started
**Priority**: HIGH
**Dependencies**: Phase 4 complete
**Goal**: Achieve 98%+ mutation efficacy, implement hash service, strengthen security

### 5.1 Mutation Testing Excellence

**Status**: ❌ Not Started
**Estimate**: 2-3 weeks
**Dependencies**: Phase 4.2 complete (≥85% baseline established)

**Tasks**:

- [ ] Raise mutation efficacy target from 85% to 98% per package
- [ ] Update .gremlins.yaml threshold to 98%
- [ ] Re-run gremlins on all packages
- [ ] Identify surviving mutants in <98% packages
- [ ] Add tests to kill surviving mutants (focus on business logic, edge cases)
- [ ] Document efficacy improvements in docs/GREMLINS-TRACKING.md
- [ ] Achieve ≥98% efficacy per package

**Quality Gates**:

- [ ] ≥98% efficacy per package (Phase 5+ requirement)
- [ ] NO EXCEPTIONS - all packages must meet target
- [ ] Gremlins reports showing <2% surviving mutants
- [ ] Efficacy tracking document updated with Phase 5 results

**Evidence**: Gremlins reports showing ≥98% efficacy, all packages meeting target

---

### 5.2 Hash Service Implementation

**Status**: ❌ Not Started
**Estimate**: 1-2 weeks
**Dependencies**: Phase 4 complete

**See detailed spec in Phase 3.4 - Deferred to Phase 5 for proper sequencing**

**Rationale for Phase 5 Placement**:

- Requires mature codebase (Phase 4 quality gates passed)
- Needs high mutation coverage (98%) for cryptographic correctness
- Benefits from lessons learned during Phase 1-4 implementation

---

### 5.3 Security Hardening

**Status**: ❌ Not Started
**Estimate**: 1-2 weeks
**Dependencies**: Phase 4 complete

**Tasks**:

- [ ] STRIDE threat modeling for all 4 products
- [ ] Penetration testing (Nuclei, ZAP, custom scripts)
- [ ] Security audit of cryptographic implementations
- [ ] TLS configuration hardening (ciphersuites, perfect forward secrecy)
- [ ] Rate limiting tuning (per-IP, per-client, per-user)
- [ ] IP allowlist validation and CIDR support
- [ ] Audit logging for all security-sensitive operations

**Quality Gates**:

- [ ] DAST scans pass with 0 critical/high findings
- [ ] SAST scans pass with 0 critical/high findings
- [ ] Secrets scanning passes (gitleaks)
- [ ] All TLS connections use TLS 1.3+ (NEVER TLS 1.2 or lower)
- [ ] Audit logs capture all auth/authz/crypto operations

**Evidence**: Security scan reports in dast-reports/, STRIDE model documented

---

### 5.4 Load Testing and Performance Tuning

**Status**: ⚠️ Partial (Service API only)
**Estimate**: 1-2 weeks
**Dependencies**: Phase 4 complete

**Tasks**:

- [x] Gatling load tests for Service API (/service/api/v1/*)
- [ ] Gatling load tests for Browser API (/browser/api/v1/*)
- [ ] Gatling load tests for Admin API (/admin/v1/*)
- [ ] Multi-product integration workflows (OAuth flow, cert issuance, KMS operations)
- [ ] Performance baseline establishment (throughput, latency, error rate)
- [ ] Bottleneck identification and remediation
- [ ] Connection pooling tuning (PostgreSQL, HTTP clients)
- [ ] Telemetry performance impact analysis

**Load Test Scenarios**:

| Scenario | Target RPS | Duration | Success Criteria |
|----------|-----------|----------|------------------|
| KMS Encrypt/Decrypt | 1000 RPS | 10 min | <100ms p99, <1% errors |
| Identity Token Issuance | 500 RPS | 10 min | <200ms p99, <1% errors |
| CA Certificate Issuance | 100 RPS | 10 min | <500ms p99, <1% errors |
| JOSE Sign/Verify | 2000 RPS | 10 min | <50ms p99, <1% errors |

**Quality Gates**:

- [ ] All load tests pass with success criteria met
- [ ] No memory leaks under sustained load
- [ ] Graceful degradation under overload (rate limiting kicks in)
- [ ] Load test reports committed to test-output/

**Evidence**: Gatling reports showing performance targets met, no regressions

---

### 5.5 E2E Workflow Testing

**Status**: ⚠️ Partial (infrastructure only)
**Estimate**: 1-2 weeks
**Dependencies**: Phase 4 complete

**Current Coverage**:

- [x] Docker Compose lifecycle tests (startup, health checks, shutdown)
- [x] Container log collection

**Missing Coverage**:

- [ ] OAuth 2.1 authorization code flow (browser → AuthZ → IdP → consent → token)
- [ ] Certificate issuance workflow (CSR → CA → certificate → CRL/OCSP)
- [ ] KMS key generation, encryption/decryption, rotation workflow
- [ ] JOSE token signing and verification workflow
- [ ] Multi-product integration (KMS → Identity federation, CA → JOSE signing)

**E2E Test Structure** (internal/test/e2e):

```
e2e_test.go           # Main test orchestrator
oauth_flow_test.go    # OAuth 2.1 flows
cert_lifecycle_test.go # CA certificate workflows
kms_operations_test.go # KMS encrypt/decrypt/rotate
jose_operations_test.go # JOSE sign/verify/encrypt/decrypt
federation_test.go    # Multi-product integration
```

**Quality Gates**:

- [ ] All product workflows covered by E2E tests
- [ ] E2E tests use real Docker Compose stack (not mocks)
- [ ] E2E tests complete in <240s (total)
- [ ] E2E tests validate end-to-end functionality (not just HTTP 200)

**Evidence**: E2E test suite passing, workflow coverage complete

---

## Phase 6: Service Template

**Status**: ❌ Not Started
**Priority**: MEDIUM
**Dependencies**: Phase 5 complete
**Goal**: Extract reusable service template from proven implementations

### 6.1 Template Design and Extraction

**Status**: ❌ Not Started
**Estimate**: 2-3 weeks

**Tasks**:

- [ ] Analyze KMS, JOSE, Identity, CA implementations for common patterns
- [ ] Design ServerTemplate abstraction (internal/template/server/)
- [ ] Design ClientSDK abstraction (internal/template/client/)
- [ ] Design Repository abstraction (internal/template/repository/)
- [ ] Extract dual HTTPS server pattern
- [ ] Extract middleware pipeline builder (CORS/CSRF/CSP/rate limit)
- [ ] Extract database abstraction (PostgreSQL + SQLite dual support)
- [ ] Extract OpenTelemetry integration patterns
- [ ] Extract health check and graceful shutdown patterns

**Template Architecture**:

```
internal/template/
├── server/          # ServerTemplate base class
│   ├── dual_https.go       # Public + Admin server management
│   ├── router.go           # Route registration framework
│   ├── middleware.go       # Pipeline builder
│   └── lifecycle.go        # Start/stop/reload
├── client/          # ClientSDK base class
│   ├── http_client.go      # HTTP client with mTLS/retry
│   ├── auth.go             # OAuth 2.1/mTLS/API key
│   └── codegen.go          # OpenAPI client generation
└── repository/      # Database abstraction
    ├── dual_db.go          # PostgreSQL + SQLite support
    ├── gorm_patterns.go    # Model registration, migrations
    └── transaction.go      # Transaction handling
```

**Customization Points**:

- **API Endpoints**: Custom OpenAPI specs per service
- **Business Logic Handlers**: Service-specific request processing
- **Database Schemas**: Custom GORM models per service
- **Client SDK Generation**: Service-specific client interfaces
- **Barrier Services**: Optional (KMS-specific, not needed for other services)

**Quality Gates**:

- [ ] 100%+ test coverage (utility code requirement)
- [ ] Template validated against all 4 existing products
- [ ] Parameterization allows full customization
- [ ] Documentation complete (usage guide, examples)

**Evidence**: Template code passing tests, validated against KMS/JOSE/Identity/CA

---

### 6.2 Product Refactoring to Use Template

**Status**: ❌ Not Started
**Estimate**: 2-3 weeks
**Dependencies**: Phase 6.1 complete

**Tasks**:

- [ ] Refactor KMS to use ServerTemplate
- [ ] Refactor JOSE to use ServerTemplate
- [ ] Refactor Identity (AuthZ, IdP, RS) to use ServerTemplate
- [ ] Refactor CA to use ServerTemplate
- [ ] Validate all tests still passing after refactoring
- [ ] Measure code reduction (target: 30-50% less boilerplate per service)
- [ ] Document migration guide for future services

**Quality Gates**:

- [ ] All refactored services pass existing test suites
- [ ] Coverage maintained at 95%/100% targets
- [ ] Mutation efficacy maintained at 98%
- [ ] No functionality regressions

**Evidence**: All services refactored, tests passing, code reduction measured

---

## Phase 7: Learn-PS Demonstration

**Status**: ❌ Not Started
**Priority**: LOW
**Dependencies**: Phase 6 complete
**Goal**: Create working Pet Store service using template to validate reusability

### 7.1 Learn-PS Service Implementation

**Status**: ❌ Not Started
**Estimate**: 1-2 weeks

**Tasks**:

- [ ] Create internal/learn/ps/ package structure
- [ ] Design Pet Store API (pets, orders, customers)
- [ ] Implement OpenAPI spec for Pet Store
- [ ] Use ServerTemplate for dual HTTPS servers
- [ ] Implement CRUD handlers using template patterns
- [ ] Add PostgreSQL/SQLite database schemas
- [ ] Implement business logic (inventory, orders, payments)
- [ ] Add OAuth 2.1 integration with Identity product

**API Endpoints**:

| Endpoint | Method | Description | Scope |
|----------|--------|-------------|-------|
| `/pets` | POST | Create pet | write:pets |
| `/pets` | GET | List pets | read:pets |
| `/pets/{id}` | GET | Get pet | read:pets |
| `/pets/{id}` | PUT | Update pet | write:pets |
| `/pets/{id}` | DELETE | Delete pet | admin:pets |
| `/orders` | POST | Create order | write:orders |
| `/orders` | GET | List orders | read:orders |
| `/orders/{id}` | GET | Get order | read:orders |
| `/customers` | POST | Create customer | write:customers |
| `/customers` | GET | List customers | read:customers |

**Quality Gates**:

- [ ] 95%+ test coverage
- [ ] 98%+ mutation efficacy
- [ ] Service starts with `cryptoutil learn-ps start`
- [ ] All CRUD operations working
- [ ] OAuth 2.1 integration functional

**Evidence**: Learn-PS service passing all tests, docker-compose deployment working

---

### 7.2 Learn-PS Documentation and Tutorials

**Status**: ❌ Not Started
**Estimate**: 1 week
**Dependencies**: Phase 7.1 complete

**Tasks**:

- [ ] Write README.md for Learn-PS (quick start, API docs, dev guide)
- [ ] Create 4-part tutorial series:
  1. Using Learn-PS (run, test, interact with API)
  2. Understanding Learn-PS (architecture, design decisions)
  3. Customizing Learn-PS (modify for your use case)
  4. Deploying Learn-PS (Docker, Kubernetes, production)
- [ ] Record video demonstration (15-20 minutes)
  - Service startup
  - API usage (Postman/curl)
  - Code walkthrough (handlers, database, middleware)
  - Customization demo (add new endpoint)
- [ ] Create deployment manifests (Docker Compose, Kubernetes)

**Quality Gates**:

- [ ] Documentation complete and accurate
- [ ] Tutorials validated by following step-by-step
- [ ] Video uploaded and accessible
- [ ] Deployment manifests working

**Evidence**: Documentation published, video available, tutorials validated

---

## Dependencies and Sequencing

### Critical Path

```
Phase 0 (Foundation) → Phase 1 (Core Infrastructure) → Phase 2 (Service Completion)
    ↓
Phase 3 (Advanced Features) → Phase 4 (Quality Gates) → Phase 5 (Production Hardening)
    ↓
Phase 6 (Service Template) → Phase 7 (Learn-PS Demonstration)
```

### Phase Dependencies Matrix

| Phase | Depends On | Blocks | Duration |
|-------|-----------|--------|----------|
| Phase 0 | None | All | Complete |
| Phase 1 | Phase 0 | Phase 2 | 3-4 weeks |
| Phase 2 | Phase 1 | Phase 3 | 2-3 weeks |
| Phase 3 | Phase 2 | Phase 4 | 2-3 weeks |
| Phase 4 | Phase 3 | Phase 5 | 3-4 weeks |
| Phase 5 | Phase 4 | Phase 6 | 3-4 weeks |
| Phase 6 | Phase 5 | Phase 7 | 4-6 weeks |
| Phase 7 | Phase 6 | None | 2-3 weeks |

**Total Estimated Duration**: 19-25 weeks (4.75-6.25 months)

### Parallel Work Opportunities

**Phase 1-2 Overlap**:

- Service completion (Phase 2) can start while infrastructure finalization (Phase 1) ongoing
- Caveat: Only if core patterns (dual HTTPS, middleware) are stable

**Phase 3-4 Overlap**:

- Advanced features (Phase 3) and quality gates (Phase 4) can run in parallel
- Caveat: Coverage/mutation must be measured after feature completion

**Phase 6-7 Overlap**:

- Learn-PS design (Phase 7) can start while template extraction (Phase 6) ongoing
- Caveat: Template API must be stable before Learn-PS implementation

---

## Risk Management

### High-Risk Items

#### 1. Mutation Testing Performance (Phase 4-5)

**Risk**: Mutation testing takes >45 minutes per package on slow machines

**Mitigation**:

- Use GitHub Actions matrix strategy for parallel package execution
- Set per-package timeout (30 minutes max)
- Exclude generated code, vendor directories
- Optimize test execution before mutation runs

**Contingency**: If mutation testing blocks CI/CD, move to nightly builds

---

#### 2. Windows Firewall Test Failures (Ongoing)

**Risk**: Tests binding to 0.0.0.0 trigger Windows Firewall exception prompts

**Mitigation**:

- ALWAYS bind to 127.0.0.1 in tests and local development
- Use 0.0.0.0 ONLY in Docker containers
- Document pattern in all test files

**Contingency**: Use Act for local workflow testing (Linux containers)

---

#### 3. Coverage Target Enforcement Resistance (Phase 4)

**Risk**: Developers rationalize <95% coverage ("mostly error handling", "thin wrapper")

**Mitigation**:

- NO EXCEPTIONS policy in constitution.md
- Coverage < 95% is BLOCKING issue in PR reviews
- Automate coverage checks in CI/CD (fail build if <95%)

**Contingency**: Add coverage enforcement to pre-push hooks

---

#### 4. Hash Service Migration Complexity (Phase 5)

**Risk**: Migrating existing hashes to new version format breaks authentication

**Mitigation**:

- Output format includes version prefix ({version}:{algorithm}:{params}:hash)
- Reject unprefixed hashes (force re-hash on next authentication)
- Document migration strategy in DETAILED.md

**Contingency**: Add backward compatibility for unprefixed hashes if user demands

---

#### 5. Service Template Abstraction Leakage (Phase 6)

**Risk**: Template abstraction too rigid, forcing services into unnatural patterns

**Mitigation**:

- Design template with customization points (handlers, middleware, config)
- Validate template against all 4 existing products before finalizing
- Allow "escape hatches" for service-specific requirements

**Contingency**: Provide multiple template variants (simple, advanced, custom)

---

### Medium-Risk Items

#### 1. Federation Configuration Complexity

**Risk**: Service federation configuration becomes unwieldy with 4+ products

**Mitigation**: YAML-based config with clear defaults and validation

---

#### 2. Docker Compose Health Check Timing

**Risk**: Health checks timeout in GitHub Actions due to shared CPU resources

**Mitigation**: Generous start_period (30-60s), diagnostic logging, exponential backoff

---

#### 3. Probabilistic Test Execution Flakiness

**Risk**: Statistical sampling misses bugs (e.g., TestProbTenth skips failing test)

**Mitigation**: Run TestProbAlways on all base algorithms, use TestProb* only for variants

---

## Success Criteria

### Phase Completion Checklist

Each phase is NOT complete until ALL criteria are met:

#### Phase 0 (Foundation) - ✅ Complete

- [x] All CI/CD workflows passing
- [x] Docker Compose deployments working
- [x] Documentation complete
- [x] Pre-commit hooks configured

#### Phase 1 (Core Infrastructure) - ⚠️ In Progress

- [ ] All infrastructure packages 100% coverage
- [ ] CGO ban enforced (NO CGO dependencies)
- [ ] FIPS 140-3 compliance validated
- [ ] Dual HTTPS pattern implemented for ALL services
- [ ] Telemetry working (services → otel-collector → Grafana)

#### Phase 2 (Service Completion) - ❌ Blocked

- [ ] All 4 products use dual HTTPS architecture
- [ ] Unified command interface working (`cryptoutil <product> start`)
- [ ] Federation configuration implemented
- [ ] All services accessible via cryptoutil command
- [ ] Docker Compose deployments updated

#### Phase 3 (Advanced Features) - ⚠️ Partial

- [x] All MFA factors implemented and tested
- [x] All client authentication methods implemented
- [x] OAuth 2.1 advanced flows implemented
- [ ] Hash service architecture designed and documented

#### Phase 4 (Quality Gates) - ⚠️ In Progress

- [ ] 95%+ coverage for ALL production packages
- [ ] 100%+ coverage for ALL infrastructure/utility packages
- [ ] ≥85% mutation efficacy for ALL packages
- [ ] <15s per unit test package, <180s total unit suite
- [ ] All golangci-lint violations fixed

#### Phase 5 (Production Hardening) - ❌ Not Started

- [ ] ≥98% mutation efficacy for ALL packages
- [ ] Hash service implemented and tested
- [ ] Security hardening complete (STRIDE, DAST, SAST)
- [ ] Load testing complete (all API types)
- [ ] E2E workflow testing complete

#### Phase 6 (Service Template) - ❌ Not Started

- [ ] ServerTemplate extracted and tested
- [ ] All 4 products refactored to use template
- [ ] Code reduction measured (30-50% target)
- [ ] Migration guide documented

#### Phase 7 (Learn-PS Demonstration) - ❌ Not Started

- [ ] Learn-PS service implemented and tested
- [ ] Documentation and tutorials complete
- [ ] Video demonstration recorded
- [ ] Deployment manifests working

---

### Final Acceptance Criteria

cryptoutil is ready for production when:

1. **All 7 phases complete** with evidence documented in DETAILED.md
2. **All quality gates passed**:
   - 95%+ coverage (production), 100%+ (infrastructure/utility)
   - ≥98% mutation efficacy per package
   - <15s per unit test package, <180s total
   - All CI/CD workflows passing
3. **All 4 products deployable** standalone and unified:
   - Docker Compose deployments working
   - Kubernetes manifests validated
   - Unified command interface functional
4. **Security hardening complete**:
   - DAST/SAST scans passing
   - TLS 1.3+ enforced
   - Audit logging comprehensive
5. **Documentation complete**:
   - README.md accurate
   - Runbooks for operations
   - Tutorials for developers
   - API documentation published
6. **Performance validated**:
   - Load tests passing
   - No memory leaks
   - Graceful degradation under load

---

## Appendix: Task Tracking

### Task Breakdown by Phase

**Phase 1: Core Infrastructure** (22 tasks)

- 1.1 Shared Crypto Library: 6 tasks (✅ complete)
- 1.2 Database Repository Layer: 7 tasks (✅ complete)
- 1.3 Telemetry Infrastructure: 6 tasks (✅ complete)
- 1.4 Configuration Management: 6 tasks (✅ complete)
- 1.5 Networking and HTTP Framework: 7 tasks (⚠️ partial)

**Phase 2: Service Completion** (18 tasks)

- 2.1 JOSE Admin Server: 6 tasks (❌ blocked)
- 2.2 Identity Admin Server Migration: 6 tasks (⚠️ partial)
- 2.3 CA Admin Server Migration: 6 tasks (❌ blocked)
- 2.4 Service Federation Configuration: 5 tasks (❌ not started)
- 2.5 Unified Command Interface: 6 tasks (⚠️ partial)

**Phase 3: Advanced Features** (12 tasks)

- 3.1 Advanced MFA Factors: 7 tasks (✅ complete)
- 3.2 Advanced Client Authentication: 5 tasks (✅ complete)
- 3.3 OAuth 2.1 Advanced Flows: 5 tasks (✅ complete)
- 3.4 Hash Service Architecture: 9 tasks (❌ deferred to Phase 5)

**Phase 4: Quality Gates** (18 tasks)

- 4.1 Coverage Analysis: 6 tasks (⚠️ in progress)
- 4.2 Mutation Testing Baseline: 7 tasks (⚠️ in progress)
- 4.3 Test Timing Optimization: 6 tasks (⚠️ in progress)
- 4.4 Linting Compliance: 5 tasks (✅ complete)

**Phase 5: Production Hardening** (25 tasks)

- 5.1 Mutation Testing Excellence: 7 tasks (❌ not started)
- 5.2 Hash Service Implementation: 9 tasks (❌ not started)
- 5.3 Security Hardening: 7 tasks (❌ not started)
- 5.4 Load Testing: 8 tasks (⚠️ partial)
- 5.5 E2E Workflow Testing: 7 tasks (⚠️ partial)

**Phase 6: Service Template** (12 tasks)

- 6.1 Template Design: 9 tasks (❌ not started)
- 6.2 Product Refactoring: 6 tasks (❌ not started)

**Phase 7: Learn-PS Demonstration** (10 tasks)

- 7.1 Service Implementation: 9 tasks (❌ not started)
- 7.2 Documentation: 7 tasks (❌ not started)

**Total Tasks**: 117 tasks across 7 phases

---

## Post-Implementation Activities

After all 7 phases complete:

1. **Documentation Updates**: Update docs/README.md with 002-cryptoutil outcomes
2. **Lessons Learned**: Document in specs/002-cryptoutil/implement/EXECUTIVE.md
3. **Post-Mortem**: Create post-mortem for any P0 incidents encountered
4. **Archive Decision**: If starting 003-cryptoutil iteration, archive 002-cryptoutil
5. **Release Tagging**: `git tag -a v0.2.0 -m "MVP quality release with service template"`
6. **Final Push**: Ensure all commits pushed to GitHub

---

*End of Plan Document*
