# Task Breakdown - 002-cryptoutil

**Date**: December 17, 2025
**Context**: Fresh restart with MVP quality focus and service template extraction
**Status**: 🎯 210+ tasks identified (ALL MANDATORY)

---

## Task Summary

**CRITICAL**: ALL phases and tasks are MANDATORY for MVP completion and service template validation.

| Phase | Tasks | Priority | Status |
|-------|-------|----------|--------|
| Phase 1: Optimize Slow Test Packages | 20 | 🔥 CRITICAL | ⏳ Not Started |
| Phase 2: Coverage Targets (95%+ strict) | 60+ | 🔥 CRITICAL | ⏳ Not Started |
| Phase 3: CI/CD Workflow Fixes (5 failures) | 15 | 🔥 CRITICAL | ⏳ Not Started |
| Phase 4: Mutation Testing QA (98%+) | 40+ | 🔴 HIGH | ⏳ Not Started |
| Phase 5: Refactor Hashes (4 types) | 20 | 🟡 MEDIUM | ⏳ Not Started |
| Phase 6: Server Architecture Unification | 30+ | 🟡 MEDIUM | ⏳ Not Started |
| Phase 7: Learn-PS Demonstration Service | 25 | 🟢 LOW | ⏳ Not Started |
| **Total** | **210+** | **~160-240h** | **0% Complete** |

---

## Phase Priorities

### 🔥 CRITICAL (Blocking MVP)

**P1: Optimize Slow Test Packages (≤12s target)**

- Fast feedback loops essential for development velocity
- Probabilistic execution for algorithm variants
- Target: ALL packages ≤12 seconds

**P2: Coverage Targets (95%+ strict, NO EXCEPTIONS)**

- Production packages: 95%+ mandatory
- Infrastructure/utility: 100% mandatory
- Per-package granular tracking

**P3: CI/CD Workflow Fixes (0 failures target)**

- ci-quality: outdated dependencies
- ci-mutation: 45min timeout
- ci-fuzz: otel collector healthcheck
- ci-dast: readyz endpoint timeout
- ci-load: otel collector healthcheck

### 🔴 HIGH (Quality Assurance)

**P4: Mutation Testing QA (98%+ killed per package)**

- API validation packages (highest priority)
- Business logic packages (high priority)
- Repository layer (medium priority)
- Infrastructure packages (lower priority)

### 🟡 MEDIUM (Architecture Cleanup)

**P5: Refactor Hashes (4 types, 3 versions)**

- Low/High Entropy × Random/Deterministic
- Version management (SHA256/384/512)
- Clean registry pattern

**P6: Server Architecture Unification**

- Extract reusable service template from SM-KMS
- Validate across 8 PRODUCT-SERVICE instances
- Document customization points

### 🟢 LOW (Validation & Demonstration)

**P7: Learn-PS Demonstration Service**

- Pet Store service using extracted template
- Copy-paste-modify starting point for customers
- Complete CRUD API, Docker deployment, tutorial series

---

## Key Differences from 001-cryptoutil

### What Changed

1. **More Aggressive Test Optimization**: ≤12s (was ≤15s)
2. **Strict Coverage Enforcement**: 95%/100% NO EXCEPTIONS (was "95%+ with exceptions")
3. **Per-Package Task Granularity**: 60+ tasks for P2, 40+ for P4 (was 8-10 coarse tasks)
4. **CI/CD First**: Fix all 5 workflow failures in P3 (was deferred)
5. **98% Mutation Target**: Per package (was 80%)
6. **Hash Refactoring**: 20 tasks, 4 types, 3 versions (was ad-hoc)
7. **Service Template**: 30+ tasks, extract from KMS, validate with Learn-PS (was not planned)

### Why These Changes

- **001-cryptoutil DETAILED.md**: 3710 lines, unmanageable, lost focus
- **Too Many Exceptions**: Accepting 66.8% coverage "because it's better than before" = 28.2 points of debt
- **Hidden Progress Gaps**: Coarse-grained tasks (e.g., "Achieve 95% coverage for identity") hid package-specific failures
- **Accumulated CI/CD Debt**: 5 failing workflows, never prioritized fixes
- **No Reusable Pattern**: 8 services duplicating infrastructure code
- **Hash Architecture Unclear**: 4 types scattered without version management

---

## References

- **Plan**: See PLAN.md for technical approach and architecture
- **Analysis**: See analyze.md for gap analysis and improvement opportunities
- **Clarifications**: See clarify.md for requirement clarifications
- **QA Questions**: See CLARIFY-QA.md for 100 multiple choice validation questions
- **Implementation**: See implement/DETAILED.md for timeline and detailed task descriptions
- **Executive Summary**: See implement/EXECUTIVE.md for stakeholder overview and risk tracking

---

## Phase Details

### Phase 1: Optimize Slow Test Packages (20 tasks)

**Goal**: ALL !integration tests complete in ≤12 seconds per package

**Strategy**:

- Baseline current test timings (identify all packages >12s)
- Apply probabilistic execution (TestProbAlways/Quarter/Tenth) to algorithm variants
- Profile hotspots, optimize redundant operations
- Create monitoring script and pre-commit hook

**Target Packages** (preliminary list, will expand after baseline):

- internal/jose (algorithm variants)
- internal/jose/server (HTTP handler overhead)
- internal/kms/client (existing probabilistic, needs tuning)
- internal/kms/server/application (barrier operations)
- internal/identity/authz (OAuth flows)
- internal/identity/authz/clientauth (mTLS handshakes)
- internal/identity/idp (MFA flows)
- internal/shared/crypto/keygen (key generation variants)
- internal/shared/crypto/digests (HKDF variants)
- internal/shared/crypto/certificate (TLS handshakes)

### Phase 2: Coverage Targets - 95% Mandatory (60+ tasks)

**Goal**: Production 95%+, Infrastructure/Utility 100%, NO EXCEPTIONS

**Strategy**:

- One parent task per major area (KMS, Identity, JOSE, CA, Shared)
- Subtasks per package (baseline, gap analysis, tests, verify)
- Per-package enforcement: BLOCKING until target met

**Major Areas**:

- **P2.1**: KMS Server (8 packages, 95%+ target)
- **P2.2**: KMS Client (1 package, 95%+ target)
- **P2.3**: Identity Server (8 packages, 95%+ target)
  - **CRITICAL**: internal/identity/authz currently 66.8% (28.2 point gap)
- **P2.4**: JOSE (2 packages, 95%+ target)
- **P2.5**: CA (2 packages, 95%+ target)
- **P2.6**: Shared Crypto (4 packages, 100%+ target - utility code)
- **P2.7**: Shared Infrastructure (8 packages, 100%+ target)
- **P2.8**: CICD (10+ packages, 100%+ target - infrastructure)

### Phase 3: CI/CD Workflow Fixes (15 tasks)

**Goal**: 0 workflow failures, all quality gates green

**Strategy**:

- Fix ci-quality (outdated dependencies)
- Fix ci-mutation (timeout, parallelization)
- Fix ci-fuzz (otel collector healthcheck)
- Fix ci-dast (readyz endpoint timeout)
- Fix ci-load (otel collector healthcheck)

**Workflows**:

- **P3.1**: ci-quality (3 subtasks: update dep, review all, automate)
- **P3.2**: ci-mutation (4 subtasks: analyze timing, reduce scope, parallelize, set timeout)
- **P3.3**: ci-fuzz (3 subtasks: analyze failure, fix healthcheck, add logging)
- **P3.4**: ci-dast (3 subtasks: analyze timeout, optimize startup, increase timeout)
- **P3.5**: ci-load (4 subtasks: coordinate with P3.3, apply fixes, verify E2E, document)

### Phase 4: Mutation Testing QA (40+ tasks)

**Goal**: 98%+ mutation kill rate per package

**Strategy**:

- Start with highest-impact packages (API validation)
- Per-package: baseline, analysis, improvement, verification
- Prioritize by impact: API validation → business logic → repository → infrastructure

**Priority Order**:

- **P4.1**: API Validation (jose, authz, businesslogic)
- **P4.2**: Business Logic (clientauth, idp, barrier, crypto)
- **P4.3**: Repository Layer (sqlrepository, repository, jose/server/repository)
- **P4.4**: Infrastructure (apperr, config, telemetry)

### Phase 5: Refactor Hashes (20 tasks)

**Goal**: Clean hash architecture with 4 types, 3 versions

**Architecture**:

```
HashService
├── LowEntropyRandomHashRegistry (PBKDF2-based)
│   ├── v1: 0-31 bytes → PBKDF2-HMAC-SHA256
│   ├── v2: 32-47 bytes → PBKDF2-HMAC-SHA384
│   └── v3: 48+ bytes → PBKDF2-HMAC-SHA512
├── LowEntropyDeterministicHashRegistry (PBKDF2, no salt)
├── HighEntropyRandomHashRegistry (HKDF-based)
└── HighEntropyDeterministicHashRegistry (HKDF, no salt)
```

**Tasks**:

- **P5.1**: Analysis and Design (5 subtasks)
- **P5.2**: Base Registry Implementation (5 subtasks)
- **P5.3**: Low Entropy Random (4 subtasks)
- **P5.4**: Low Entropy Deterministic (4 subtasks)
- **P5.5**: High Entropy Random (4 subtasks)
- **P5.6**: High Entropy Deterministic (4 subtasks)

### Phase 6: Server Architecture Unification (30+ tasks)

**Goal**: Extract reusable service template from SM-KMS for 8 PRODUCT-SERVICE instances

**8 Instances**:

1. sm-kms (Secrets Manager - KMS)
2. pki-ca (PKI - Certificate Authority)
3. jose-ja (JOSE - JWK Authority)
4. identity-authz (Identity - Authorization Server)
5. identity-idp (Identity - Identity Provider)
6. identity-rs (Identity - Resource Server)
7. identity-rp (Identity - Relying Party - BFF)
8. identity-spa (Identity - SPA - static hosting)

**Template Features**:

- Dual HTTPS servers (public 8xxx, admin 127.0.0.1:9090)
- Dual API paths (/browser session-based, /service token-based)
- Middleware pipeline (CORS, CSRF, CSP, rate limiting, IP allowlist)
- Database abstraction (PostgreSQL + SQLite)
- Telemetry integration (OTLP → Otel Collector)
- Optional barrier services (for KMS-like services)

**Tasks**:

- **P6.1**: Analysis (extract SM-KMS patterns, identify commonalities)
- **P6.2**: Server Template Package (dual HTTPS, routers, middleware)
- **P6.3**: Client Template Package (SDK base, auth strategies, generation)
- **P6.4**: Database Layer Abstraction (PostgreSQL + SQLite)
- **P6.5**: Barrier Services Integration (optional per service)
- **P6.6**: Telemetry Integration (OTLP patterns)
- **P6.7**: Configuration Management (YAML, validation, secrets)
- **P6.8**: Documentation and Examples (usage guide, customization)

### Phase 7: Learn-PS Demonstration Service (25 tasks)

**Goal**: Working Pet Store service using extracted template as validation

**Purpose**:

- Validate template completeness and usability
- Provide copy-paste-modify starting point for customers
- Demonstrate all template features in production context

**Scope**:

- Complete CRUD API (pets, orders, customers)
- PostgreSQL schema, GORM models
- Business logic (inventory, order processing)
- Client SDK generation
- Unit tests (95%+), integration tests, mutation tests (98%+)
- Docker Compose deployment
- Tutorial series, video demonstration

**Tasks**:

- **P7.1**: Service Design (requirements, OpenAPI, database schema)
- **P7.2**: Service Implementation (instantiate template, handlers, repository)
- **P7.3**: Testing (unit 95%+, integration, mutation 98%+, performance ≤12s)
- **P7.4**: Deployment (Docker Compose, Kubernetes manifests)
- **P7.5**: Documentation (README, tutorials, video walkthrough)

---

## Success Criteria

### Phase 1 Success (Fast Tests)

- ✅ ALL !integration packages ≤12 seconds execution time
- ✅ Probabilistic testing patterns documented
- ✅ Pre-commit hook enforces timing limit
- ✅ Monitoring script provides real-time feedback

### Phase 2 Success (High Coverage)

- ✅ Production packages: 95%+ coverage (NO EXCEPTIONS)
- ✅ Infrastructure/utility: 100% coverage (NO EXCEPTIONS)
- ✅ Per-package verification complete
- ✅ Gap analysis documented for each package

### Phase 3 Success (Stable CI/CD)

- ✅ ALL 5 workflow failures resolved
- ✅ ci-quality passing (dependencies current)
- ✅ ci-mutation passing (parallel, ≤20min total)
- ✅ ci-fuzz passing (otel collector healthy)
- ✅ ci-dast passing (readyz responsive)
- ✅ ci-load passing (E2E load tests successful)

### Phase 4 Success (High Mutation Kill)

- ✅ API validation packages: 98%+ efficacy
- ✅ Business logic packages: 98%+ efficacy
- ✅ Repository layer packages: 98%+ efficacy
- ✅ Infrastructure packages: 98%+ efficacy

### Phase 5 Success (Clean Hash Architecture)

- ✅ 4 hash types implemented (Low/High × Random/Deterministic)
- ✅ 3 versions per type (SHA256/384/512)
- ✅ Version-aware Verify method
- ✅ Migration strategy documented

### Phase 6 Success (Reusable Template)

- ✅ Server template extracted from SM-KMS
- ✅ Client template with SDK generation
- ✅ Database abstraction (PostgreSQL + SQLite)
- ✅ Template documentation complete
- ✅ Customization points identified

### Phase 7 Success (Learn-PS Validation)

- ✅ Pet Store service operational
- ✅ 95%+ coverage, 98%+ mutation efficacy
- ✅ Docker Compose deployment working
- ✅ Tutorial series complete
- ✅ Video demonstration recorded
