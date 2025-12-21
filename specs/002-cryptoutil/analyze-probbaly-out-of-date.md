# Gap Analysis - 002-cryptoutil

**Date**: December 17, 2025
**Context**: Fresh analysis after archiving 001-cryptoutil (3710 lines, too much AI slop)
**Status**: 🎯 Identifying MVP Quality Gaps

---

## Executive Summary

**Primary Gaps from 001-cryptoutil**:

1. **Test Performance**: Some packages >30s execution time (target: ≤30s, hard limit 60s)
2. **Coverage Gaps**: Many packages <95% (e.g., identity/authz 66.8%, kms/businesslogic 39.0%)
3. **CI/CD Failures**: Multiple workflow issues (outdated dependencies, mutation timeout, PostgreSQL service containers missing)
4. **Mutation Testing**: No baseline for 98% efficacy target
5. **Hash Architecture**: Scattered implementation, no version management, missing 3 of 4 registries, no FIPS 140-3 validation
6. **Service Template**: 8 services duplicate infrastructure code (no reusable template, 3 template packages needed)
7. **Learn-PS Demonstration**: No reference implementation validating template for customers

---

## Gap 1: Test Performance (≤30s Target)

### Current State

**Baseline needed**: Run `go test -json -v ./... 2>&1 | tee test-output/baseline-timing-002.txt` to identify packages >12s

**Expected Problem Packages** (based on 001-cryptoutil patterns):

- internal/jose: Algorithm variant tests (many crypto operations)
- internal/jose/server: HTTP handler overhead, middleware setup/teardown
- internal/kms/client: Currently uses probabilistic execution, may need tuning
- internal/kms/server/application: Barrier operations, unseal tests
- internal/identity/authz: OAuth 2.1 flows, token generation
- internal/identity/authz/clientauth: mTLS handshake overhead
- internal/identity/idp: MFA flows, consent/login operations
- internal/shared/crypto/keygen: Key generation variants
- internal/shared/crypto/digests: HKDF variants, SHA variants
- internal/shared/crypto/certificate: TLS handshakes, certificate generation

### Gap Analysis

**Missing**:

- Current baseline timing per package
- Identification of specific slow tests within packages
- Profiling data (hotspots, redundant operations)

**Strategy**:

1. Run baseline timing
2. Parse JSON output, extract per-package execution times
3. Profile each package >12s (identify hotspots)
4. Apply probabilistic execution to algorithm variants
5. Optimize redundant operations (TLS handshakes, cert generation)
6. Verify coverage maintained

**Success Criteria**: ALL !integration packages ≤30 seconds (target), ≤60 seconds (hard limit)

---

## Gap 2: Code Coverage (95%+ Production, 98% Infra/Util)

### Current State (from 001-cryptoutil final)

**Production Packages** (Target: 95%+):

| Package | Current | Gap | Priority |
|---------|---------|-----|----------|
| internal/identity/authz | 66.8% | **28.2 points** | 🔥 CRITICAL |
| internal/kms/server/businesslogic | 39.0% | **56.0 points** | 🔥 CRITICAL |
| internal/jose | ~85% | **10 points** | 🔴 HIGH |
| internal/jose/server | ~62% | **33 points** | 🔴 HIGH |
| internal/kms/client | 74.9% | **20.1 points** | 🔴 HIGH |
| internal/kms/server/application | 64.6% | **30.4 points** | 🔴 HIGH |
| internal/identity/idp | Unknown | TBD | 🔴 HIGH |
| internal/identity/rs | Unknown | TBD | 🔴 HIGH |
| internal/identity/rp | Unknown | TBD | 🔴 HIGH |
| internal/identity/spa | Unknown | TBD | 🔴 HIGH |
| internal/identity/authz/clientauth | Unknown | TBD | 🔴 HIGH |
| internal/ca/server | Unknown | TBD | 🔴 HIGH |

**Infrastructure/Utility Packages** (Target: 100%):

| Package | Current | Gap | Priority |
|---------|---------|-----|----------|
| internal/shared/crypto/* | ~90-95% | 5-10 points | 🟡 MEDIUM |
| internal/shared/* (other) | ~85-95% | 5-15 points | 🟡 MEDIUM |
| internal/cmd/cicd/* | ~60-80% | 20-40 points | 🟡 MEDIUM |

### Gap Analysis

**Major Gaps**:

1. **identity/authz (66.8%)**: Missing error path tests, edge case handling
2. **kms/businesslogic (39.0%)**: 18 core operations at 0% (AddElasticKey, Get*, Post*, Update, Delete, Import, Revoke)
3. **jose/server (62.1%)**: HTTP handler tests, middleware tests
4. **kms/client (74.9%)**: Client SDK error paths, retry logic

**Root Causes**:

- "95%+ with exceptions" policy in 001-cryptoutil led to accepting low coverage
- Coarse-grained tasks (e.g., "Achieve 95% for identity") hid per-package gaps
- No baseline → gap analysis → targeted tests workflow enforced

**Strategy**:

1. Per-package baseline: `go test ./pkg -coverprofile=./test-output/coverage_pkg_baseline.out`
2. HTML gap analysis: `go tool cover -html=./test-output/coverage_pkg_baseline.out -o ./test-output/coverage_pkg_baseline.html`
3. Identify RED lines (uncovered code)
4. Write targeted tests ONLY for RED lines (not trial-and-error)
5. Verify: Re-run coverage, confirm ≥95% or ≥98% achieved
6. **BLOCKING**: Can't proceed to next package until current ≥ target

**Success Criteria**: Production 95%+, infrastructure/utility 100%, NO EXCEPTIONS

---

## Gap 3: CI/CD Workflow Failures (5 Active)

### Current State

**Failing Workflows**:

1. **ci-quality**: Outdated dependency (github.com/goccy/go-yaml v1.19.0 → v1.19.1)
2. **ci-mutation**: Timeout after 45 minutes (sequential execution too slow)
3. **ci-fuzz**: opentelemetry-collector-contrib healthcheck exit 1
4. **ci-dast**: /admin/v1/readyz endpoint not ready within timeout
5. **ci-load**: opentelemetry-collector-contrib healthcheck exit 1 (same as ci-fuzz)

### Gap Analysis

**ci-quality Gap**:

- **Issue**: Single outdated dependency failing quality gate
- **Root Cause**: No automated dependency updates (dependabot not configured)
- **Impact**: Quality gate blocked, can't merge PRs
- **Fix**: Update dependency, add dependabot.yml

**ci-mutation Gap**:

- **Issue**: Sequential package execution exceeds 45 minute timeout
- **Root Cause**: Large packages (e.g., identity/authz) take 10+ minutes alone, sequential adds up to 45+min
- **Impact**: Mutation testing incomplete, no efficacy data
- **Fix**: Parallelize by package (GitHub Actions matrix), reduce timeout to 15min/job

**ci-fuzz Gap**:

- **Issue**: Otel collector healthcheck exit 1
- **Root Cause**: Healthcheck command incorrect or service not starting
- **Impact**: Fuzz testing environment not starting
- **Fix**: Update compose.integration.yml healthcheck, add diagnostic logging

**ci-dast Gap**:

- **Issue**: Admin readyz endpoint not ready within timeout
- **Root Cause**: Service startup slower than expected (database migration, unseal)
- **Impact**: DAST scanning cannot proceed
- **Fix**: Optimize startup, increase timeout with exponential backoff

**ci-load Gap**:

- **Issue**: Same as ci-fuzz (otel collector healthcheck)
- **Root Cause**: Same as ci-fuzz
- **Impact**: Load testing environment not starting
- **Fix**: Coordinate with ci-fuzz fix, apply to compose.yml

**Strategy**:

1. Fix ci-quality first (quick win, unblocks merges)
2. Fix ci-fuzz and ci-load together (same root cause)
3. Fix ci-dast (optimize startup, increase timeout)
4. Fix ci-mutation last (requires most work: parallelization)

**Success Criteria**: ALL 5 workflows passing, 0 failures

---

## Gap 4: Mutation Testing (98%+ Efficacy Target)

### Current State

**Baseline Missing**:

- No baseline mutation testing data for 98% efficacy target
- 001-cryptoutil used 80% efficacy target (too lenient)
- No per-package efficacy tracking

### Gap Analysis

**Missing Data**:

- Current efficacy % per package
- Lived mutants analysis (which mutants survived)
- Test quality assessment (boundary conditions, error paths, edge cases)

**Strategy**:

1. Run baseline per package: `gremlins unleash ./pkg` → document efficacy %
2. Analyze lived mutants (gremlins output shows which mutants survived)
3. Write targeted tests for lived mutants (not generic tests)
4. Re-run gremlins, verify ≥98% efficacy achieved
5. **BLOCKING**: Can't proceed to next package until current ≥98%

**Priority Order** (highest-impact packages first):

1. **API Validation**: jose, authz, businesslogic (highest risk)
2. **Business Logic**: clientauth, idp, barrier, crypto
3. **Repository Layer**: sqlrepository, repository
4. **Infrastructure**: apperr, config, telemetry

**Success Criteria**: ALL packages ≥98% efficacy

---

## Gap 5: Hash Service Refactoring (4 Registries, 3 Versions, FIPS 140-3)

### Current State

**Current Implementation**:

- Low Entropy Random Hash Registry (PBKDF2-based) exists
- Missing: Low Entropy Deterministic, High Entropy Random, High Entropy Deterministic registries
- No version management (SHA256/384/512 by input size)
- Scattered implementation (no unified HashRegistry interface)
- No FIPS 140-3 compliance validation documented

### Gap Analysis

**Missing Registries**:

1. **LowEntropyDeterministicHashRegistry** (PBKDF2-HMAC-SHA256/384/512, no salt, deterministic output)
2. **HighEntropyRandomHashRegistry** (HKDF-HMAC-SHA256/384/512, random salt per operation)
3. **HighEntropyDeterministicHashRegistry** (HKDF-HMAC-SHA256/384/512, no salt, deterministic output)

**Missing Features**:

- HashRegistry interface consistent across all 4 registries
- Version management (v1: 0-31 bytes SHA-256, v2: 32-47 bytes SHA-384, v3: 48+ bytes SHA-512)
- Hash output format with version metadata (enables algorithm agility)
- Version-aware Verify method (can verify any version using metadata)
- HashWithVersion method (specify version explicitly v1/v2/v3)
- Migration strategy documentation
- FIPS 140-3 compliance validation (PBKDF2/HKDF approved algorithms documented)

**Strategy**:

1. **P5.1**: Analysis and design (parameterized base registry, version selection)
2. **P5.2**: Base registry implementation (version management, HashWithLatest, HashWithVersion, Verify)
3. **P5.3**: Low Entropy Random (PBKDF2, salt, v1/v2/v3)
4. **P5.4**: Low Entropy Deterministic (PBKDF2, no salt, v1/v2/v3)
5. **P5.5**: High Entropy Random (HKDF, salt, v1/v2/v3)
6. **P5.6**: High Entropy Deterministic (HKDF, no salt, v1/v2/v3)

**Success Criteria**: 4 registries implemented, 3 versions per registry, HashRegistry interface consistent, hash output format with version metadata, migration strategy documented, FIPS 140-3 compliance validated

---

## Gap 6: Service Template Extraction (8 PRODUCT-SERVICE Instances, 3 Template Packages)

### Current State

**8 Services Duplicating Code**:

1. sm-kms (Secrets Manager - KMS)
2. jose-ja (JOSE - JWK Authority)
3. pki-ca (PKI - Certificate Authority)
4. identity-authz (Identity - Authorization Server)
5. identity-idp (Identity - Identity Provider)
6. identity-rs (Identity - Resource Server)
7. identity-rp (Identity - Relying Party - BFF)
8. identity-spa (Identity - SPA - static hosting)

**Duplicated Infrastructure**:

- Dual HTTPS servers (public, admin)
- Dual API paths (/browser, /service)
- Middleware (CORS, CSRF, CSP, rate limiting, IP allowlist)
- Database abstraction (PostgreSQL + SQLite)
- Telemetry integration (OTLP → Otel Collector)
- Configuration management (YAML, validation, secrets)

### Gap Analysis

**No Reusable Template**:

- Each service implements dual HTTPS independently
- Middleware patterns duplicated across services
- Database layer duplicated (same GORM patterns)
- Telemetry setup duplicated (same OTLP config)
- Configuration validation duplicated

**Consequences**:

- High maintenance burden (8× effort for infrastructure changes)
- Inconsistency risk (services diverge over time)
- New service creation slow (copy-paste-modify error-prone)
- No validated starting point for customers

**Strategy**:

1. **P6.1**: Analyze SM-KMS (extract patterns)
2. **P6.2**: Create server template package (internal/template/server: dual HTTPS, routes, middleware, lifecycle)
3. **P6.3**: Create client template package (internal/template/client: HTTP client, auth strategies, SDK generation)
4. **P6.4**: Create repository template package (internal/template/repository: dual database, GORM patterns, transactions)
5. **P6.5**: Barrier services integration (optional per service)
6. **P6.6**: Telemetry integration (OTLP patterns)
7. **P6.7**: Configuration management (YAML, validation, secrets)
8. **P6.8**: Refactor all 8 services to use template, document usage guide

**Success Criteria**: 3 template packages created, 8 services refactored to use template, template usage guide documented, Learn-PS validates reusability

---

## Gap 7: Documentation (001-cryptoutil Too Long)

### Current State

**001-cryptoutil DETAILED.md**:

- **Size**: 3710 lines
- **Problem**: Unmanageable, lost focus, hard to navigate
- **Root Cause**: Append-only timeline grew without bounds

### Gap Analysis

**Documentation Issues**:

- Section 1 (task checklist): Too many completed tasks (clutter)
- Section 2 (timeline): Too many session entries (3710 lines total)
- No clear separation between active work and historical reference

**Strategy**:

1. **Reset 002-cryptoutil**: Fresh start, all tasks unchecked
2. **Clear timeline**: Section 2 starts empty, append as work progresses
3. **Archive 001-cryptoutil**: Preserve history without cluttering new spec
4. **Strict task structure**: Per-package granularity, no hiding progress gaps
5. **Append-only discipline**: Only add timeline entries for completed work, never edit history

**Success Criteria**: DETAILED.md remains <2000 lines, timeline entries concise

---

## Gap 8: Gremlins Windows Panic (Known Issue)

### Current State

**Issue**: gremlins v0.6.0 crashes with "panic: error, this is temporary" after coverage gathering on Windows

**Impact**:

- Mutation testing cannot run locally on Windows
- CI/CD Linux environment works perfectly
- Baseline data available in docs/gremlins/MUTATION-TESTING-BASELINE.md

### Gap Analysis

**Workaround**:

- Use CI/CD for mutation testing (Linux-based runners)
- Baseline data sufficient for Phase 4 planning

**Permanent Fix**:

- Track gremlins upstream issue
- Consider alternative mutation testing tools (go-mutesting, mutagen)
- Re-evaluate after gremlins v0.7.0 release

**Strategy**: Accept workaround, proceed with CI/CD-based mutation testing

**Success Criteria**: Phase 4 mutation testing completes successfully in CI/CD

---

## Improvement Opportunities

### Quick Wins (High Impact, Low Effort)

1. **Fix ci-quality** (30 min): Update dependency, add dependabot.yml
2. **Test timing baseline** (1 hour): Identify packages >12s
3. **Coverage baseline** (2 hours): Run baseline per major area, identify critical gaps

### High-Impact Improvements (High Impact, Medium Effort)

1. **Fix identity/authz coverage** (8-12 hours): 28.2 point gap, critical path
2. **Fix kms/businesslogic coverage** (16-24 hours): 56 point gap, 18 core ops at 0%
3. **Parallelize ci-mutation** (4-6 hours): Reduce timeout from 45min to 20min

### Strategic Improvements (High Impact, High Effort)

1. **Service template extraction** (48-72 hours): Eliminate duplication across 8 services
2. **Learn-PS demonstration** (40-60 hours): Validate template, provide customer starting point
3. **Hash refactoring** (16-24 hours): Clean architecture, version management, 4 types

---

## Risk Assessment

### High-Risk Areas

1. **kms/businesslogic (39% coverage)**: 18 core operations untested, high cryptographic risk
2. **CI/CD failures (5 active)**: Quality gates not enforced, technical debt accumulating
3. **Service template extraction**: Complex refactoring, potential regressions if not carefully tested

### Medium-Risk Areas

1. **identity/authz (66.8% coverage)**: OAuth 2.1 flows critical, error paths undertested
2. **Mutation testing**: No baseline, 98% target aggressive, may require extensive test development
3. **Hash refactoring**: Affects all services, migration strategy must be carefully documented

### Low-Risk Areas

1. **Test performance optimization**: Probabilistic execution proven pattern from 001-cryptoutil
2. **Learn-PS demonstration**: Isolated validation service, minimal impact on existing services
3. **Documentation restructure**: Clear improvement, low regression risk

---

## Prioritization Rationale

### Phase 1 (Test Performance) First

**Why**: Fast feedback loops essential for all subsequent work. Optimizing tests first ensures efficient development in P2-P7.

### Phase 2 (Coverage) Second

**Why**: Must achieve 95%+/98% targets before mutations. Mutation testing ineffective without good coverage baseline.

### Phase 3 (CI/CD Fixes) Third

**Why**: Quality gates must be green before proceeding to mutations and architecture work. Can't merge PRs without passing CI/CD.

### Phase 4 (Mutations) Fourth

**Why**: Requires high coverage from P2. Validates test quality before architecture refactoring begins.

### Phase 5 (Hash Refactoring) Fifth

**Why**: Simpler than service template extraction. Proves refactoring discipline before tackling larger P6.

### Phase 6 (Service Template) Sixth

**Why**: Complex refactoring. Requires all quality gates passing (P1-P4) before attempting.

### Phase 7 (Learn-PS) Last

**Why**: Validates template from P6. Can only proceed after template extraction complete.

---

## Success Metrics

### Phase Completion Metrics

| Phase | Success Metric | Target |
|-------|----------------|--------|
| P1 | Test execution time | ≤30s per package (target), ≤60s (hard limit) |
| P2 | Code coverage | 95%+ production, 98% infra/util, NO EXCEPTIONS |
| P3 | CI/CD health | 0 failures, PostgreSQL service containers added |
| P4 | Mutation efficacy | 98%+ efficacy per package |
| P5 | Hash registries | 4 registries × 3 versions, FIPS 140-3 validated |
| P6 | Service template | 3 packages, 8 services refactored, documented |
| P7 | Learn-PS | Operational, 95% coverage, 98% mutations, ≤12s tests |

### Overall MVP Metrics

- ✅ Fast tests (≤30s per package target, <100s total !integration suite)
- ✅ High coverage (95%+ production, 98% infra/util, NO EXCEPTIONS)
- ✅ Stable CI/CD (0 failures, all workflows green, time targets met)
- ✅ High mutation kill (98%+ efficacy per package, ALL packages)
- ✅ Clean hash architecture (4 registries × 3 versions, FIPS 140-3 validated)
- ✅ Reusable service template (3 packages, 8 services using template, documented, Learn-PS validates)
- ✅ Customer demonstrability (Learn-PS operational, 4-part tutorial series, video demonstration)

---

## Conclusion

**002-cryptoutil represents a strategic reset**: Clean up AI slop from 001-cryptoutil, enforce strict quality targets without exceptions, extract reusable patterns before scaling to multiple services.

**Key Differences from 001-cryptoutil**:

1. **No Exceptions**: 95%+/98% coverage mandatory, 98% mutations mandatory
2. **Per-Package Enforcement**: Granular tracking, BLOCKING until targets met
3. **CI/CD First**: Fix all 5 failures before proceeding
4. **Template Extraction**: Reusable pattern for 8 services, validated with Learn-PS
5. **Strict Documentation**: DETAILED.md <2000 lines, clear separation of active vs historical

**Expected Outcome**: Production MVP quality, reusable service template, validated demonstration service, no technical debt.
