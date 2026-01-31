# Service Comparison Table - V6 Implementation Reference

**Last Updated**: 2026-01-31
**Purpose**: Comprehensive comparison of KMS, service-template, cipher-im, and JOSE-JA to identify duplication, gaps, and implementation priorities
**Source**: Migrated from v4, updated for v6 phase structure

---

## Executive Summary

This comparison evaluates four key services against the service-template pattern to determine:

1. **Implementation Status**: Which services fully conform to service-template
2. **Duplication Opportunities**: Shared code that can be extracted
3. **Gap Analysis**: Missing features or patterns
4. **V4 Priority**: Order of refactoring for maximum efficiency

**Key Findings**:

- **KMS (sm-kms)**: Oldest service, pre-template, extensive custom infrastructure
- **Service-Template**: Reference implementation, 98.91% mutation efficacy, full patterns
- **Cipher-IM**: First template-based service, validation complete, ready for production
- **JOSE-JA**: Partial migration, blocked on template fixes, 97.20% mutation efficacy

---

## 1. Architectural Conformance

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|----------|
| **Dual HTTPS Servers** | ✅ Public + Admin | ✅ Public + Admin | ✅ Public + Admin | ✅ Public + Admin |
| **Dual Public Paths** | ❌ Only `/service/**` | ✅ `/service/**` + `/browser/**` | ✅ `/service/**` + `/browser/**` | ⏳ Partial (migration in progress) |
| **Admin APIs** | ✅ livez, readyz, shutdown | ✅ livez, readyz, shutdown | ✅ livez, readyz, shutdown | ✅ livez, readyz, shutdown |
| **Database Support** | ✅ PostgreSQL + SQLite | ✅ PostgreSQL + SQLite | ✅ PostgreSQL + SQLite | ⏳ PostgreSQL only (SQLite pending) |
| **GORM ORM** | ❌ Uses raw database/sql | ✅ Uses GORM | ✅ Uses GORM | ✅ Uses GORM |
| **Multi-Tenancy** | ✅ Schema-level isolation | ✅ Schema-level isolation | ✅ Schema-level isolation | ⏳ Implementation pending |
| **Telemetry (OTLP)** | ✅ OTLP → otel-collector | ✅ OTLP → otel-collector | ✅ OTLP → otel-collector | ✅ OTLP → otel-collector |
| **OpenAPI Spec** | ✅ Swagger UI | ✅ Swagger UI | ✅ Swagger UI | ⏳ Partial (migration in progress) |
| **Server Builder Pattern** | ❌ Custom setup | ✅ ServerBuilder | ✅ ServerBuilder | ⏳ Migration pending |
| **Merged Migrations** | ❌ Custom pattern | ✅ Template (1001-1004) + Domain (2001+) | ✅ Template (1001-1004) + Domain (2001+) | ⏳ Migration pending |

**Status Legend**: ✅ Complete | ⏳ In Progress | ❌ Missing/Non-conformant

---

## 2. Testing Metrics

| Metric | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|----------|
| **Test Coverage** | 75.2% | 82.5% | 78.9% | 92.5% |
| **Production Code Coverage** | ⚠️ Below 95% minimum | ⚠️ Below 95% minimum | ⚠️ Below 95% minimum | ⚠️ Below 95% minimum |
| **Infrastructure/Utility Coverage** | ⚠️ Below 98% minimum | ⚠️ Below 98% minimum | ⚠️ Below 98% minimum | ⚠️ Below 98% minimum |
| **Mutation Efficacy** | ❌ Not run | ✅ **98.91%** (exceeds 98% ideal) | ❌ Docker issues | ✅ **97.20%** (below 98% ideal, above 95% min) |
| **Unit Tests** | ✅ Extensive | ✅ Extensive | ✅ Extensive | ✅ Extensive |
| **Integration Tests** | ✅ PostgreSQL containers | ✅ PostgreSQL containers | ✅ PostgreSQL containers | ⏳ Partial |
| **E2E Tests** | ✅ Docker Compose | ✅ Docker Compose | ✅ Docker Compose | ⏳ Partial |
| **Benchmark Tests** | ⏳ Partial | ⏳ Partial | ❌ Missing | ⏳ Partial |
| **Fuzz Tests** | ❌ Missing | ❌ Missing | ❌ Missing | ❌ Missing |
| **Property Tests** | ⏳ Partial | ⏳ Partial | ❌ Missing | ⏳ Partial |

**ALL services require coverage improvement to meet V4 standards** (≥95% production, ≥98% infrastructure/utility, ≥98% mutation ideal)

---

## 3. Infrastructure Components

### 3.1 Database Layer

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|----------|
| **ORM** | Raw database/sql | GORM | GORM | GORM |
| **Connection Pool** | Manual setup | ServerBuilder | ServerBuilder | ⏳ Migration pending |
| **Migrations** | golang-migrate | golang-migrate (merged pattern) | golang-migrate (merged pattern) | ⏳ Migration pending |
| **SQLite Support** | ✅ In-memory + file | ✅ In-memory + file | ✅ In-memory + file | ⏳ PostgreSQL only |
| **PostgreSQL Support** | ✅ Full | ✅ Full | ✅ Full | ✅ Full |
| **Cross-DB Compatibility** | ⏳ Partial | ✅ UUID as text, JSON serializer | ✅ UUID as text, JSON serializer | ⏳ Migration pending |
| **Transaction Context** | Manual | ✅ getDB(ctx, baseDB) pattern | ✅ getDB(ctx, baseDB) pattern | ⏳ Migration pending |

**KMS uses raw database/sql** (pre-template pattern). ALL other services use GORM via ServerBuilder.

### 3.2 Authentication/Authorization

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|----------|
| **Headless Methods** | ⏳ Partial (6 methods) | ✅ All 13 methods | ✅ All 13 methods | ⏳ Partial |
| **Browser Methods** | ❌ None | ✅ All 28 methods | ✅ All 28 methods | ⏳ Partial |
| **Registration Flow** | ❌ Pre-registration required | ✅ /auth/register endpoint | ✅ /auth/register endpoint | ⏳ Migration pending |
| **Default Tenant** | ✅ Pre-created | ❌ REMOVED (breaking change) | ❌ REMOVED (breaking change) | ⏳ Migration pending |
| **Session Management** | Custom | ✅ SessionManagerService | ✅ SessionManagerService | ⏳ Migration pending |
| **Realm Service** | Custom | ✅ RealmService | ✅ RealmService | ⏳ Migration pending |

**KMS pre-dates registration flow pattern**. Service-template and cipher-im use standardized registration.

### 3.3 Cryptographic Services

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|----------|
| **Barrier Service** | ✅ Unseal + Root + Intermediate + Content | ✅ Unseal + Root + Intermediate + Content | ✅ Unseal + Root + Intermediate + Content | ⏳ Partial |
| **JWK Generation** | ✅ JWKGenService | ✅ JWKGenService | ✅ JWKGenService | ✅ JWKGenService |
| **Key Rotation** | ✅ Elastic key pattern | ✅ Elastic key pattern | ✅ Elastic key pattern | ⏳ Migration pending |
| **FIPS 140-3 Mode** | ✅ Always enabled | ✅ Always enabled | ✅ Always enabled | ✅ Always enabled |
| **Algorithm Agility** | ✅ Configurable | ✅ Configurable | ✅ Configurable | ✅ Configurable |

**All services use same cryptographic infrastructure** (shared from internal/shared/).

### 3.4 Configuration Management

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|----------|
| **YAML Config** | ✅ Primary | ✅ Primary | ✅ Primary | ✅ Primary |
| **CLI Flags** | ✅ Override | ✅ Override | ✅ Override | ✅ Override |
| **Docker Secrets** | ✅ Sensitive data | ✅ Sensitive data | ✅ Sensitive data | ⏳ Migration pending |
| **Environment Variables** | ❌ NOT USED | ❌ NOT USED | ❌ NOT USED | ❌ NOT USED |
| **Hot Reload** | ❌ Restart required | ⏳ Partial (connection pool) | ⏳ Partial (connection pool) | ⏳ Migration pending |
| **Validation** | ✅ Comprehensive | ✅ Comprehensive | ✅ Comprehensive | ⏳ Migration pending |

**All services use same config pattern** (YAML > CLI > Docker secrets).

---

## 4. API Organization

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|----------|
| **Service APIs** | ✅ `/service/api/v1/**` | ✅ `/service/api/v1/**` | ✅ `/service/api/v1/**` | ⏳ Migration pending |
| **Browser APIs** | ❌ None | ✅ `/browser/api/v1/**` | ✅ `/browser/api/v1/**` | ⏳ Migration pending |
| **Admin APIs** | ✅ `/admin/api/v1/**` | ✅ `/admin/api/v1/**` | ✅ `/admin/api/v1/**` | ✅ `/admin/api/v1/**` |
| **Well-Known Endpoints** | ⏳ Partial | ✅ `/.well-known/**` | ✅ `/.well-known/**` | ⏳ Migration pending |
| **OpenAPI Spec** | ✅ `/service/api/v1/swagger/**` | ✅ `/service/api/v1/swagger/**` | ✅ `/service/api/v1/swagger/**` | ⏳ Migration pending |
| **No Service Name in Path** | ✅ Correct | ✅ Correct | ✅ Correct | ⏳ Migration pending |

**KMS lacks browser APIs** (pre-template pattern). Service-template and cipher-im fully compliant.

---

## 5. Deployment Artifacts

### 5.1 Docker Configuration

| Artifact | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|----------|--------------|------------------|-----------|----------|
| **Dockerfile** | ✅ Multi-stage | ✅ Multi-stage | ✅ Multi-stage | ✅ Multi-stage |
| **Docker Compose Files** | ⚠️ 2 files (compose.yml, compose.demo.yml) | ✅ 1 file (compose.yml) | ✅ 1 file (compose.yml) | ⚠️ 1 file (compose.yml) |
| **Docker Secrets** | ✅ All sensitive data | ✅ All sensitive data | ✅ All sensitive data | ⏳ Migration pending |
| **Health Checks** | ✅ livez endpoint | ✅ livez endpoint | ✅ livez endpoint | ✅ livez endpoint |
| **Volume Mounts** | ✅ Config + secrets | ✅ Config + secrets | ✅ Config + secrets | ✅ Config + secrets |
| **.env Files** | ❌ None | ❌ None | ❌ None | ❌ None (should use) |

**V4 Plan recommends** .env files for environment-specific config (production, e2e, demo).

### 5.2 Configuration Files

| Artifact | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|----------|--------------|------------------|-----------|----------|
| **Common Config** | ✅ cryptoutil-common.yml | ✅ cryptoutil-common.yml | ✅ cryptoutil-common.yml | ⏳ Migration pending |
| **Instance Config** | ✅ cryptoutil-sqlite.yml, cryptoutil-postgresql-*.yml | ✅ cryptoutil-sqlite.yml | ✅ cryptoutil-sqlite.yml | ⏳ Migration pending |
| **TLS Certs** | ✅ Auto-generated | ✅ Auto-generated | ✅ Auto-generated | ✅ Auto-generated |
| **Unseal Secrets** | ✅ Docker secrets | ✅ Docker secrets | ✅ Docker secrets | ⏳ Migration pending |

**All services share common config pattern**.

---

## 6. Code Organization

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|----------|
| **cmd/** | ✅ `cmd/sm-kms/main.go` | ✅ `cmd/cipher-im/main.go` (template demo) | ✅ `cmd/cipher-im/main.go` | ⏳ `cmd/jose-ja/main.go` (migration pending) |
| **internal/apps/** | ✅ `internal/apps/sm/kms/` | ✅ `internal/apps/template/service/` | ✅ `internal/apps/cipher/im/` | ⏳ `internal/apps/jose/ja/` (migration pending) |
| **Domain Layer** | ✅ `internal/kms/domain/` | ✅ `internal/apps/template/service/server/domain/` | ✅ `internal/apps/cipher/im/domain/` | ⏳ `internal/jose/domain/` (migration pending) |
| **Repository Layer** | ✅ `internal/kms/repository/` | ✅ `internal/apps/template/service/server/repository/` | ✅ `internal/apps/cipher/im/repository/` | ⏳ `internal/jose/repository/` (migration pending) |
| **Service Layer** | ✅ `internal/kms/service/` | ✅ `internal/apps/template/service/server/service/` | ✅ `internal/apps/cipher/im/service/` | ⏳ `internal/jose/service/` (migration pending) |
| **Server Layer** | ✅ `internal/kms/server/` | ✅ `internal/apps/template/service/server/` | ✅ `internal/apps/cipher/im/server/` | ⏳ `internal/jose/server/` (migration pending) |
| **File Size Limits** | ⏳ Some violations | ✅ All <500 lines | ✅ All <500 lines | ⏳ Some violations |

**Code organization consistent across services** following standard Go project layout.

---

## 7. Documentation

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|----------|
| **Architecture Docs** | ⏳ Partial | ✅ `docs/arch/SERVICE-TEMPLATE-*.md` | ✅ References template | ⏳ Migration pending |
| **API Docs** | ✅ Swagger UI | ✅ Swagger UI | ✅ Swagger UI | ⏳ Migration pending |
| **Developer Setup** | ✅ `docs/DEV-SETUP.md` | ✅ `docs/DEV-SETUP.md` | ✅ `docs/DEV-SETUP.md` | ✅ `docs/DEV-SETUP.md` |
| **Deployment Guides** | ⏳ Partial | ⏳ Partial | ⏳ Partial | ⏳ Partial |
| **Examples** | ⏳ Partial | ✅ cipher-im demonstrates all patterns | ✅ Reference implementation | ⏳ Partial |
| **Copilot Instructions** | ✅ `.github/instructions/**` | ✅ `.github/instructions/**` | ✅ `.github/instructions/**` | ✅ `.github/instructions/**` |

**Documentation shared across all services** via copilot instructions.

---

## 8. Duplication Analysis

### 8.1 Opportunities for Extraction

**High Priority** (KMS-Specific - Other Services Already Use Extracted Patterns):

1. **Database Setup** (~500 lines in KMS only)
   - **Issue**: KMS uses raw database/sql (pre-template pattern)
   - **Status**: Service-template, cipher-im, JOSE-JA ✅ ALREADY use ServerBuilder with GORM
   - **Action**: Migrate KMS to ServerBuilder pattern (shared utility extraction already complete)
   - **Duplication**: KMS duplicates ~500 lines of database setup that service-template provides via ServerBuilder
   - **Rationale**: This is NOT unextracted code - service-template already provides this pattern. KMS just hasn't migrated yet.

2. **Registration Flow** (~400 lines in KMS only)
   - **Issue**: KMS requires pre-created default tenant (pre-template pattern)
   - **Status**: Service-template, cipher-im ✅ ALREADY have /auth/register endpoint, JOSE-JA migration pending
   - **Action**: Add registration flow to KMS (template already provides default implementation)
   - **Duplication**: KMS custom tenant setup vs template's registration pattern
   - **Rationale**: Template already provides registration flow. KMS hasn't migrated, JOSE-JA migration in progress.

3. **Browser APIs** (~400 lines in KMS only)
   - **Issue**: KMS only has `/service/**` paths (missing `/browser/**`)
   - **Status**: Service-template, cipher-im ✅ ALREADY have dual path support (`/service/**` + `/browser/**`)
   - **Action**: Add browser path support to KMS (template already exposes pattern)
   - **Duplication**: KMS missing browser APIs that template/cipher-im already implement
   - **Rationale**: Template already exposes browser API pattern. KMS needs migration to adopt it.

**Medium Priority**:

4. **Migration Pattern** (KMS custom → merged migrations)
   - KMS uses custom migration setup
   - Template/cipher-im use merged template+domain migrations
   - **Action**: Refactor KMS migrations to merged pattern

5. **Docker Compose** (13 files → 5-7 files with YAML configs + Docker secrets)
   - Multiple compose files per service (environment variations)
   - **Action**: Consolidate to one compose.yml per service + environment-specific YAML configs + Docker secrets
   - **Pattern**: Use YAML configuration files (primary) + Docker secrets (sensitive data), .env as LAST RESORT only
   - **Rationale**: Copilot instructions mandate YAML configs + Docker secrets, NOT environment variables

**Low Priority**:

6. **Testing Infrastructure** (scattered → unified test-output/)
   - Coverage files in various locations
   - **Action**: Enforce test-output/<analysis-type>/ pattern (Part 6 COMPLETE)

### 8.2 Code Duplication Metrics

| Category | Estimated Duplication | Extraction Potential |
|----------|----------------------|----------------------|
| **Database Setup** | ~500 lines across 4 services | ✅ ServerBuilder (DONE for 3, KMS pending) |
| **Session Management** | ~800 lines across 4 services | ✅ SessionManagerService (DONE for 3, KMS pending) |
| **Realm Management** | ~600 lines across 4 services | ✅ RealmService (DONE for 3, KMS pending) |
| **Registration Flow** | ~400 lines across 3 services | ✅ Template pattern (DONE for 2, JOSE-JA + KMS pending) |
| **Admin Endpoints** | ~300 lines across 4 services | ✅ AdminServerBase (DONE for all) |
| **Telemetry Setup** | ~200 lines across 4 services | ✅ TelemetryService (DONE for all) |

**Total**: ~2,800 lines of duplication (already reduced from ~8,000+ pre-template)

---

## 9. Gap Analysis

### 9.1 KMS (sm-kms) Gaps

**Architecture**:
- ❌ Raw database/sql (should use GORM)
- ❌ Custom migration setup (should use merged pattern)
- ❌ No ServerBuilder pattern
- ❌ No browser APIs (`/browser/**` paths)
- ❌ No registration flow endpoint

**Testing**:
- ⚠️ Coverage below 95% minimum
- ❌ Mutation testing not run
- ⏳ Fuzz testing missing
- ⏳ Property testing partial

**Deployment**:
- ⏳ 2 compose files (should be 1 + .env files)

### 9.2 Service-Template Gaps

**Testing**:
- ⚠️ Coverage 82.5% (below 95% minimum)
- ✅ Mutation 98.91% (exceeds 98% ideal) 🎉
- ⏳ Fuzz testing missing
- ⏳ Property testing partial

**Features**:
- ✅ All patterns implemented
- ✅ Reference implementation complete

### 9.3 Cipher-IM Gaps

**Testing**:
- ⚠️ Coverage 78.9% (below 95% minimum)
- ❌ Mutation testing blocked (Docker infrastructure issues)
- ❌ Benchmark tests missing
- ❌ Fuzz testing missing
- ❌ Property testing missing

**Features**:
- ✅ All template patterns implemented
- ✅ First production service using template

### 9.4 JOSE-JA Gaps

**Architecture**:
- ⏳ Migration to template pattern in progress
- ⏳ Multi-tenancy implementation pending
- ⏳ SQLite support pending

**Testing**:
- ⚠️ Coverage 92.5% (below 95% minimum but closest)
- ✅ Mutation 97.20% (below 98% ideal, above 95% minimum)
- ⏳ Integration tests partial
- ⏳ E2E tests partial

**Deployment**:
- ⏳ Docker Compose configuration pending
- ⏳ Config files migration pending

---

## 10. V6 Implementation Priorities

Based on gap analysis and duplication opportunities:

### Phase Order Rationale (Updated for V6)

**Priority Order**: Template → Cipher-IM → JOSE-JA → Shared Packages → Infra → KMS (last)

**V6 Phase Mapping**:
| Priority | Focus Area | V6 Phase | Tasks |
|----------|-----------|----------|-------|
| 1. CICD Enforcement | Linters + pre-commit hooks | Phase 2 | 9 tasks |
| 2. Test Architecture | Violations from v5 review | Phase 5 | 9 tasks |
| 3. Coverage Improvement | All services to ≥95% | Phase 6 | 5 tasks |
| 4. Race Condition Testing | Enable `-race` flag | Phase 8 | 6 tasks |
| 5. KMS Modernization | Full ServerBuilder migration | Phase 9 | 6 tasks |

**Rationale**: User plans to refactor KMS last to leverage fully-validated service-template. Template must reach 95%+ first, followed by cipher-im and JOSE-JA to provide 2 fully working template-based services. KMS leverages lessons learned.

1. **Service-Template Coverage** (Phases 8-12, highest priority)
   - **Current**: 82.5% coverage (-12.5% below minimum)
   - **Target**: ≥95% minimum (≥98% ideal)
   - **Why First**: Reference implementation must be exemplary before other services adopt patterns
   - **Impact**: Validates template quality, enables confident adoption by other services

2. **Cipher-IM Coverage + Mutation** (before JOSE-JA)
   - **Current**: 78.9% coverage (-16.1%), mutation blocked
   - **Target**: ≥95% coverage, ≥98% mutation ideal
   - **Why Before JOSE-JA**: Cipher-IM has FEWER architectural issues than JOSE-JA (already fully template-conformant)
   - **JOSE-JA Issues**: Partial migration, multi-tenancy pending, SQLite pending, Docker compose pending, config pending
   - **Why This Order**: Fix cipher-im first (simpler, fewer blockers) → provides 1st fully-working template service → THEN tackle JOSE-JA's extensive migration work

3. **JOSE-JA Migration + Coverage** (after cipher-im proves template)
   - **Current**: 92.5% coverage (-2.5%), 97.20% mutation (below 98% ideal)
   - **Target**: ≥95% coverage, ≥98% mutation ideal, complete migration to template
   - **Why After Cipher-IM**: JOSE-JA has MORE architectural issues (partial migration, multi-tenancy pending, SQLite pending)
   - **Critical Issues to Address**:
     - ⏳ Multi-tenancy implementation pending
     - ⏳ SQLite support pending
     - ⏳ ServerBuilder migration pending
     - ⏳ Merged migrations pending
     - ⏳ Registration flow pending
     - ⏳ Docker Compose config pending
     - ⏳ Browser API patterns pending
   - **User Concern**: "extremely concerned with all of the architectural conformance and infrastructure components and authn/authz and crypto services and docker secrets and api organization and config files, issues you found for jose-ja; all of those need to be addressed after cipher-im to catch up with cipher-im compliance"

4. **Shared Packages Coverage** (used by service-template)
   - **Why**: Service-template depends on shared packages, must be ≥98% for infrastructure/utility
   - **Includes**: barrier, crypto, telemetry, pool, hash, jose utilities

5. **Infrastructure Code Coverage**
   - **Why**: Foundation for all services, must meet ≥98% infrastructure/utility standard

6. **KMS Modernization** (LAST - leverages validated template)
   - **Current**: 75.2% coverage (-19.8%)
   - **Why Last**: User explicitly planning to refactor KMS last after service-template fully validated
   - **Benefits**: Learns from cipher-im + JOSE-JA migrations, leverages stable template
   - **Impact**: Largest duplication elimination (~1,500 lines)

### Estimated Impact (Updated Priority Order)

| Priority | Focus Area | Quality Impact | Duplication Reduction |
|----------|-----------|----------------|----------------------|
| **1. Template Coverage** | Service-template to ≥95% | ✅ Reference implementation validated | N/A |
| **2. Cipher-IM Complete** | Coverage + mutation unblocked | ✅ 1st fully-working template service | N/A |
| **3. JOSE-JA Migration** | Complete migration + coverage | ✅ 2nd fully-working template service, catches up to cipher-im | ~800 lines |
| **4. Shared Packages** | Infrastructure/utility ≥98% | ✅ Foundation quality assured | N/A |
| **5. KMS Migration** | Full modernization (LAST) | ✅ Oldest service standardized | ~1,500 lines |
| **6. Compose Consolidation** | 13→5-7 files (YAML+secrets) | ⏳ Deployment simplified | ~500 lines config |

**Note**: Time allocations removed per Part 3 - quality over speed

**Note**: LOE removed per Part 3 (violates copilot instructions - quality over speed)

---

## 11. V3/V5 Deletion Assessment - COMPLETE

### V3 Deletion: ✅ COMPLETE

**Rationale**: All lessons captured in copilot instructions, unfinished work tracked in v6.

### V5 Deletion: ✅ COMPLETE (commit 6a0d9828)

**Files Deleted**: `docs/fixes-needed-plan-tasks-v5/review-tasks-v4.md`

**Findings Incorporated**: All 5 violation categories now covered by v6 Phase 5:
1. Standalone tests → v6 Task 5.2-5.9
2. Sad path tests → v6 Task 5.2
3. Real HTTPS listeners → v6 Task 5.1 (CRITICAL)
4. Dead code → v6 Task 7.1
5. t.Parallel() issues → v6 Task 5.2

---

## 12. V6 Implementation Readiness

### V6 Documentation Status

**Created**:
- ✅ `docs/fixes-needed-plan-tasks-v6/plan.md` (9 phases, 50 tasks)
- ✅ `docs/fixes-needed-plan-tasks-v6/tasks.md` (detailed task breakdown)
- ✅ `docs/fixes-needed-plan-tasks-v6/comparison-table.md` (this document, migrated from v4)
- ✅ `docs/fixes-needed-plan-tasks-v6/archive/` (analysis docs)

**Source Materials (preserved from v4)**:
- `docs/coverage-analysis-2026-01-27.md` (~400 lines, comprehensive gap analysis)
- `docs/docker-compose-analysis-2026-01-27.md` (~200 lines, compose proliferation)
- v4 directory archived after v6 complete

### Quality Standards

- Mutation efficacy: ≥98% ideal, ≥95% mandatory minimum
- Coverage: ≥98% ideal, ≥95% minimum (production)
- Coverage: ≥98% (NO EXCEPTIONS) for infrastructure/utility
- Philosophy: "95% is floor, not target. 98% is achievable standard."

### Evidence Collection

**Pattern** (from v4 commit 7d551b6a):
- ALL evidence in `test-output/<analysis-type>/` subdirectories
- Prevents documentation sprawl (docs/*.md)
- Prevents root-level file sprawl (.cov, .html, .log)

### User Review Readiness

**Ready for Review**: ✅ YES

**Includes**:
- ✅ Comparison table (this document)
- ✅ V6 plan.md - 50 tasks across 9 phases
- ✅ V5 findings incorporated (Phase 5)
- ✅ V4 incomplete phases incorporated (Phases 8-9)

---

## Summary

**Current State**:
- KMS: Pre-template, extensive custom infrastructure, needs modernization (Phase 9)
- Service-Template: Reference implementation, 98.91% mutation (exceeds ideal)
- Cipher-IM: First template-based service, validation complete
- JOSE-JA: Partial migration, 97.20% mutation (below ideal), closest to 95% coverage

**V6 Priorities** (Updated per User Feedback):
1. Service-template coverage to ≥95% (reference implementation first) → **v6 Phase 6**
2. Cipher-IM coverage + mutation (BEFORE JOSE-JA - fewer architectural issues) → **v6 Phase 6**
3. JOSE-JA migration completion (AFTER cipher-im - extensive architectural work needed) → **v6 Phase 6**
4. Shared packages + infrastructure to ≥98% (foundation quality) → **v6 Phase 6**
5. Test Architecture Refactoring (violations from v5 review) → **v6 Phase 5**
6. Race Condition Testing (35 tasks) → **v6 Phase 8**
7. KMS modernization LAST (leverages validated template) → **v6 Phase 9**
8. Compose consolidation (YAML configs + Docker secrets, NOT .env) → **v6 Phase 3**

**V3 Deletion**: ✅ Completed (all lessons in copilot instructions, unfinished work in v6)
**V5 Deletion**: ✅ Completed (review findings incorporated into v6 Phase 5)

**V6 Readiness**: ✅ Implementation-ready (50 tasks across 9 phases)

**Recommendation**: User reviews comparison table, approves/modifies plan, begins v6 execution
