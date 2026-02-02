# Service Comparison Table - V7 Implementation Reference

**Last Updated**: 2026-02-02
**Purpose**: Comprehensive comparison of KMS, service-template, cipher-im, and JOSE-JA to track V7 migration progress
**Source**: Migrated from v6, updated for v7 unified architecture plan
**V7 Goal**: ALL services use MANDATORY patterns - NO optional disabled modes

---

## Executive Summary

This comparison evaluates four key services against the service-template pattern for V7 unification:

1. **Implementation Status**: Which services fully conform to service-template
2. **V7 Migration Targets**: What must change for unified MANDATORY patterns
3. **Gap Analysis**: Missing features or patterns to address
4. **V7 Priority**: KMS migration is the primary focus

**Key Findings**:

- **KMS (sm-kms)**: Oldest service, pre-template, requires FULL migration to GORM, JWT, OpenAPI, template barrier
- **Service-Template**: Reference implementation, 98.91% mutation efficacy, full patterns
- **Cipher-IM**: First template-based service, fully conformant, validation target
- **JOSE-JA**: Template-based service, needs validation that V7 changes don't regress

**V7 Key Changes** (per quizme answers):
- Q1: Fresh start (no data migration needed)
- Q2: Merge shared/barrier INTO template barrier (feature parity required)
- Q3: Internal only (no API versioning needed)
- Q4: Correctness first (no shortcuts)
- Q5: Full regression + E2E + coverage; mutation testing LAST
- Q6: Continuous documentation updates

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

---

## 10. V7 Implementation Priorities

Based on comparison analysis and quizme decisions:

### Phase Sequence

| Phase | Focus | Purpose |
|-------|-------|---------|
| **1** | Analysis | Current state documentation |
| **2** | V6 Cleanup | Delete optional modes (BarrierModeDisabled, etc.) |
| **3** | KMS Foundation | GORM models, migrations, repositories |
| **4** | KMS Migration | Authentication, OpenAPI, barrier integration |
| **5** | Barrier Merge | shared/barrier INTO template |
| **6** | Testing | Regression, E2E, coverage, mutation (LAST) |
| **7** | Documentation | Finalization and consistency |

### Quality Gates (Per Phase)

- ✅ All tests pass (`runTests`)
- ✅ Coverage ≥95% production, ≥98% infrastructure
- ✅ Linting clean (`golangci-lint run`)
- ✅ Documentation updated (continuous per Q6)
- ✅ Mutation testing ≥95% (Phase 6 ONLY - per Q5)

---

## 11. V4/V6 Superseded Work

### V6 Phase 13 (Optional Modes): ❌ ARCHIVED

**Created**: BarrierModeDisabled, JWTAuthModeDisabled, MigrationModeDisabled, DatabaseModeRawSQL
**Problem**: Optional abstraction modes fragment architecture
**V7 Fix**: MANDATORY patterns - delete optional modes, enforce unified approach

### V4 Incomplete Phases: ✅ INCORPORATED

- Phase 0.4 (KMS Modernization) → V7 Phase 3-4
- Phase 0.5 (Mutation Testing) → V7 Phase 6.7
- Phase 0.6 (Documentation) → V7 Phase 7

### V6 Task 5.1 (BLOCKED): ✅ ADDRESSED

**Original Block**: StartApplicationListener not implemented
**V7 Resolution**: Complete migration removes need for partial abstractions

---

## 12. V7 References

**V7 Documentation**:
- `docs/fixes-needed-plan-tasks-v7/plan.md` - Implementation plan (7 phases, 40 tasks)
- `docs/fixes-needed-plan-tasks-v7/tasks.md` - Detailed task breakdown

**Architecture**:
- `.github/instructions/02-01.architecture.instructions.md`
- `.github/instructions/02-02.service-template.instructions.md`
- `.github/instructions/03-08.server-builder.instructions.md`

---

## Summary

**Current State**:
- KMS: Pre-template, SQLRepository/raw database/sql, needs FULL migration
- Service-Template: Reference, 98.91% mutation
- Cipher-IM: Template-based, fully conformant
- JOSE-JA: Template-based, fully conformant

**V7 Approach** (quizme decisions):
1. Q1=D: Fresh start (no data migration)
2. Q2=C: Merge barriers (shared → template)
3. Q3=D: Internal only (no API versioning)
4. Q4=A: Correctness first
5. Q5=E: Full testing, mutation LAST
6. Q6=C: Continuous documentation

**V7 Priorities**: 7 phases, 40 tasks - sequential execution with quality gates

**V4/V6 Status**: Superseded - incorporated or archived
