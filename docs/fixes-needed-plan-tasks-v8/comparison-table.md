# Service Comparison Table - V8 Implementation Reference

**Last Updated**: 2026-02-03
**Purpose**: Accurate comparison based on CODE ARCHAEOLOGY (not optimistic claims)
**Source**: Deep analysis of V7 claims vs actual code implementation
**V8 Goal**: Complete actual KMS migration that V7 only partially addressed

---

## CRITICAL V8 FINDINGS

**User skepticism about V7 claims was JUSTIFIED**. Code verification reveals:

| Claim (V7) | Reality (Code) |
|------------|----------------|
| Tasks 5.3-5.4 "barrier integration" addressed | ❌ Both marked "Not Started" in tasks.md |
| KMS uses template barrier | ❌ KMS imports `shared/barrier` (4 files still) |
| KMS barrier adapter created | ✅ True but UNUSED - KMS still uses shared/barrier |
| server.go has TODOs about incomplete migration | ✅ 3 TODOs confirm work NOT done |
| Phase 5 complete | ❌ Only Tasks 5.1, 5.2 done (analysis), 5.3, 5.4 NOT done (implementation) |

**Evidence** (verified 2026-02-03):
```bash
$ grep -r "shared/barrier" internal/kms/ --include="*.go" | wc -l
4  # KMS STILL uses shared/barrier!

$ grep "TODO" internal/kms/server/server.go
# TODO(Phase2-5): KMS needs to be migrated to use template's GORM database and barrier.
# TODO(Phase2-5): Replace with template's GORM database and barrier.
# TODO(Phase2-5): Switch to TemplateWithDomain mode once KMS uses template DB.
```

---

## V8 Barrier Architecture Decision

**Per quizme-v1.md Decision E**: Single barrier implementation in `internal/apps/template/service/server/barrier/`

- Template barrier uses GORM (not raw database/sql)
- KMS must migrate from `shared/barrier` to template barrier
- After KMS migration, `internal/shared/barrier/` will be DELETED

**Current Barrier Locations**:
| Location | Used By | Status |
|----------|---------|--------|
| `internal/apps/template/service/server/barrier/` | Template, Cipher-IM, JOSE-JA (via ServerBuilder) | ✅ Target |
| `internal/shared/barrier/` | KMS (directly) | ❌ DELETE after KMS migration |
| `internal/kms/server/barrier/orm_barrier_adapter.go` | KMS (unused adapter) | ❌ DELETE - unused |

---

## Executive Summary

This comparison evaluates four key services against the service-template pattern for V8 unification:

1. **Implementation Status**: Which services fully conform to service-template
2. **V8 Migration Targets**: What must change for unified MANDATORY patterns
3. **Gap Analysis**: Missing features or patterns to address
4. **V8 Priority**: KMS migration is the primary focus

**Key Findings**:

- **KMS (sm-kms)**: Uses ServerBuilder but still imports `shared/barrier` - barrier migration incomplete
- **Service-Template**: Reference implementation, provides template barrier + ServerBuilder
- **Cipher-IM**: First template-based service, fully uses template barrier via ServerBuilder
- **JOSE-JA**: Template-based service, uses ServerBuilder with template barrier

**V8 Executive Decisions** (per quizme-v1.md):
- Q1=E: Single barrier in template only (not shared)
- Q2=E: Delete shared/barrier IMMEDIATELY after KMS migration (no archive period)
- Q3=E: Full testing scope (unit+integration+E2E per phase, mutations at end)
- Q4=E: Incremental doc updates for ACTUALLY-WRONG instructions only

---

## 1. Architectural Conformance

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|---------|
| **Dual HTTPS Servers** | ✅ Public + Admin | ✅ Public + Admin | ✅ Public + Admin | ✅ Public + Admin |
| **Dual Public Paths** | ❌ Only `/service/**` | ✅ `/service/**` + `/browser/**` | ✅ `/service/**` + `/browser/**` | ⏳ Partial |
| **Admin APIs** | ✅ livez, readyz, shutdown | ✅ livez, readyz, shutdown | ✅ livez, readyz, shutdown | ✅ livez, readyz, shutdown |
| **Database Support** | ✅ PostgreSQL + SQLite | ✅ PostgreSQL + SQLite | ✅ PostgreSQL + SQLite | ⏳ PostgreSQL only |
| **GORM ORM** | ❌ Uses raw database/sql | ✅ Uses GORM | ✅ Uses GORM | ✅ Uses GORM |
| **Multi-Tenancy** | ✅ Schema-level isolation | ✅ Schema-level isolation | ✅ Schema-level isolation | ⏳ Implementation pending |
| **Telemetry (OTLP)** | ✅ OTLP → otel-collector | ✅ OTLP → otel-collector | ✅ OTLP → otel-collector | ✅ OTLP → otel-collector |
| **OpenAPI Spec** | ✅ Swagger UI | ✅ Swagger UI | ✅ Swagger UI | ⏳ Partial |
| **Server Builder Pattern** | ⏳ Uses but incomplete | ✅ ServerBuilder | ✅ ServerBuilder | ✅ ServerBuilder |
| **Merged Migrations** | ❌ Custom pattern | ✅ Template (1001-1004) + Domain (2001+) | ✅ Template + Domain | ⏳ Migration pending |
| **Uses Template Barrier** | ❌ Uses shared/barrier | ✅ Template barrier | ✅ Via ServerBuilder | ✅ Via ServerBuilder |

**Status Legend**: ✅ Complete | ⏳ In Progress | ❌ Missing/Non-conformant

---

## 2. Testing Metrics

| Metric | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|---------|
| **Test Coverage** | 75.2% | 82.5% | 78.9% | 92.5% |
| **Production Code Coverage** | ⚠️ Below 95% minimum | ⚠️ Below 95% minimum | ⚠️ Below 95% minimum | ⚠️ Below 95% minimum |
| **Infrastructure/Utility Coverage** | ⚠️ Below 98% minimum | ⚠️ Below 98% minimum | ⚠️ Below 98% minimum | ⚠️ Below 98% minimum |
| **Mutation Efficacy** | ❌ Not run | ✅ **98.91%** (exceeds 98% ideal) | ❌ Docker issues | ✅ **97.20%** |
| **Unit Tests** | ✅ Extensive | ✅ Extensive | ✅ Extensive | ✅ Extensive |
| **Integration Tests** | ✅ PostgreSQL containers | ✅ PostgreSQL containers | ✅ PostgreSQL containers | ⏳ Partial |
| **E2E Tests** | ✅ Docker Compose | ✅ Docker Compose | ✅ Docker Compose | ⏳ Partial |
| **Benchmark Tests** | ⏳ Partial | ⏳ Partial | ❌ Missing | ⏳ Partial |
| **Fuzz Tests** | ❌ Missing | ❌ Missing | ❌ Missing | ❌ Missing |
| **Property Tests** | ⏳ Partial | ⏳ Partial | ❌ Missing | ⏳ Partial |

**ALL services require coverage improvement to meet standards** (≥95% production, ≥98% infrastructure/utility, ≥98% mutation ideal)

---

## 3. Infrastructure Components

### 3.1 Database Layer

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|---------|
| **ORM** | Raw database/sql | GORM | GORM | GORM |
| **Connection Pool** | Manual setup | ServerBuilder | ServerBuilder | ServerBuilder |
| **Migrations** | golang-migrate | golang-migrate (merged pattern) | golang-migrate (merged) | ⏳ Migration pending |
| **SQLite Support** | ✅ In-memory + file | ✅ In-memory + file | ✅ In-memory + file | ⏳ PostgreSQL only |
| **PostgreSQL Support** | ✅ Full | ✅ Full | ✅ Full | ✅ Full |
| **Cross-DB Compatibility** | ⏳ Partial | ✅ UUID as text, JSON serializer | ✅ UUID as text, JSON serializer | ⏳ Migration pending |
| **Transaction Context** | Manual | ✅ getDB(ctx, baseDB) pattern | ✅ getDB(ctx, baseDB) pattern | ⏳ Migration pending |

**KMS uses raw database/sql** (pre-template pattern). ALL other services use GORM via ServerBuilder.

### 3.2 Cryptographic Services (Barrier)

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|---------|
| **Barrier Source** | ❌ `shared/barrier` | ✅ Template barrier | ✅ Template barrier (via SB) | ✅ Template barrier (via SB) |
| **Barrier Storage** | Raw SQL | GORM | GORM | GORM |
| **Key Hierarchy** | ✅ Unseal → Root → Intermediate → Content | ✅ Same | ✅ Same | ✅ Same |
| **JWK Generation** | ✅ JWKGenService | ✅ JWKGenService | ✅ JWKGenService | ✅ JWKGenService |
| **Key Rotation** | ✅ Elastic key pattern | ✅ Elastic key pattern | ✅ Elastic key pattern | ⏳ Migration pending |
| **FIPS 140-3 Mode** | ✅ Always enabled | ✅ Always enabled | ✅ Always enabled | ✅ Always enabled |

**V8 Action**: Migrate KMS from `shared/barrier` to template barrier, then DELETE `shared/barrier`.

### 3.3 Authentication/Authorization

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|---------|
| **Headless Methods** | ⏳ Partial (6 methods) | ✅ All 13 methods | ✅ All 13 methods | ⏳ Partial |
| **Browser Methods** | ❌ None | ✅ All 28 methods | ✅ All 28 methods | ⏳ Partial |
| **Registration Flow** | ❌ Pre-registration required | ✅ /auth/register endpoint | ✅ /auth/register endpoint | ⏳ Migration pending |
| **Default Tenant** | ✅ Pre-created | ❌ REMOVED (breaking change) | ❌ REMOVED (breaking change) | ⏳ Migration pending |
| **Session Management** | Custom | ✅ SessionManagerService | ✅ SessionManagerService | ⏳ Migration pending |
| **Realm Service** | Custom | ✅ RealmService | ✅ RealmService | ⏳ Migration pending |

**KMS pre-dates registration flow pattern**. Service-template and cipher-im use standardized registration.

### 3.4 Configuration Management

| Component | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|-----------|--------------|------------------|-----------|---------|
| **YAML Config** | ✅ Primary | ✅ Primary | ✅ Primary | ✅ Primary |
| **CLI Flags** | ✅ Override | ✅ Override | ✅ Override | ✅ Override |
| **Docker Secrets** | ✅ Sensitive data | ✅ Sensitive data | ✅ Sensitive data | ⏳ Migration pending |
| **Environment Variables** | ❌ NOT USED | ❌ NOT USED | ❌ NOT USED | ❌ NOT USED |
| **Hot Reload** | ❌ Restart required | ⏳ Partial | ⏳ Partial | ⏳ Migration pending |
| **Validation** | ✅ Comprehensive | ✅ Comprehensive | ✅ Comprehensive | ⏳ Migration pending |

**All services use same config pattern** (YAML > CLI > Docker secrets).

---

## 4. API Organization

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|---------|
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
|----------|--------------|------------------|-----------|---------|
| **Dockerfile** | ✅ Multi-stage | ✅ Multi-stage | ✅ Multi-stage | ✅ Multi-stage |
| **Docker Compose Files** | ⚠️ 2 files | ✅ 1 file | ✅ 1 file | ⚠️ 1 file |
| **Docker Secrets** | ✅ All sensitive data | ✅ All sensitive data | ✅ All sensitive data | ⏳ Migration pending |
| **Health Checks** | ✅ livez endpoint | ✅ livez endpoint | ✅ livez endpoint | ✅ livez endpoint |
| **Volume Mounts** | ✅ Config + secrets | ✅ Config + secrets | ✅ Config + secrets | ✅ Config + secrets |

---

## 6. Code Organization

| Aspect | KMS (sm-kms) | Service-Template | Cipher-IM | JOSE-JA |
|--------|--------------|------------------|-----------|---------|
| **cmd/** | ✅ `cmd/sm-kms/main.go` | ✅ `cmd/cipher-im/main.go` | ✅ `cmd/cipher-im/main.go` | ⏳ `cmd/jose-ja/main.go` |
| **internal/apps/** | ✅ `internal/kms/` | ✅ `internal/apps/template/` | ✅ `internal/apps/cipher/im/` | ⏳ `internal/apps/jose/ja/` |
| **Domain Layer** | ✅ `internal/kms/domain/` | ✅ Template domain | ✅ `internal/apps/cipher/im/domain/` | ⏳ Migration pending |
| **Repository Layer** | ✅ `internal/kms/repository/` | ✅ Template repository | ✅ `internal/apps/cipher/im/repository/` | ⏳ Migration pending |
| **Service Layer** | ✅ `internal/kms/service/` | ✅ Template service | ✅ `internal/apps/cipher/im/service/` | ⏳ Migration pending |
| **Server Layer** | ✅ `internal/kms/server/` | ✅ Template server | ✅ `internal/apps/cipher/im/server/` | ⏳ Migration pending |
| **File Size Limits** | ⏳ Some violations | ✅ All <500 lines | ✅ All <500 lines | ⏳ Some violations |

---

## 7. Gap Analysis

### 7.1 KMS (sm-kms) Gaps - V8 PRIMARY TARGET

**Barrier Migration** (CRITICAL):
- ❌ Still imports `shared/barrier` (4 files)
- ❌ Has unused `orm_barrier_adapter.go`
- ✅ Uses ServerBuilder but with incomplete migration
- **V8 Action**: Complete barrier migration to template barrier

**Database**:
- ❌ Raw database/sql (should use GORM)
- ❌ Custom migration setup (should use merged pattern)
- **V8 Action**: Already using ServerBuilder, just needs barrier completion

**API**:
- ❌ No browser APIs (`/browser/**` paths)
- ❌ No registration flow endpoint
- **V8 Action**: Add after barrier migration

### 7.2 Service-Template Gaps

**Testing**:
- ⚠️ Coverage 82.5% (below 95% minimum)
- ✅ Mutation 98.91% (exceeds 98% ideal) 🎉
- ⏳ Fuzz testing missing
- ⏳ Property testing partial

**Features**: ✅ All patterns implemented, reference implementation complete

### 7.3 Cipher-IM Gaps

**Testing**:
- ⚠️ Coverage 78.9% (below 95% minimum)
- ❌ Mutation testing blocked (Docker infrastructure issues)
- ❌ Benchmark tests missing
- ❌ Fuzz testing missing

**Features**: ✅ All template patterns implemented, first production service using template

### 7.4 JOSE-JA Gaps

**Architecture**:
- ⏳ Migration to template pattern in progress
- ⏳ Multi-tenancy implementation pending
- ⏳ SQLite support pending

**Testing**:
- ⚠️ Coverage 92.5% (below 95% minimum but closest)
- ✅ Mutation 97.20% (below 98% ideal, above 95% minimum)
- ⏳ Integration tests partial
- ⏳ E2E tests partial

---

## 8. V8 Implementation Priorities

Based on gap analysis and executive decisions:

### Phase Sequence

| Phase | Focus | Purpose |
|-------|-------|---------|
| **1** | Research & Analysis | Code archaeology, accurate state documentation |
| **2** | KMS Barrier Migration | Complete migration from shared/barrier to template |
| **3** | Testing & Validation | Unit + Integration + E2E for migrated code |
| **4** | Delete shared/barrier | Remove unused code IMMEDIATELY (per Q2=E) |
| **5** | Mutation Testing | Final quality gate (grouped at end per Q3=E) |

### Quality Gates (Per Phase)

- ✅ All tests pass (`runTests`)
- ✅ Coverage ≥95% production, ≥98% infrastructure
- ✅ Linting clean (`golangci-lint run`)
- ✅ Incremental doc updates for ACTUALLY-WRONG instructions only (per Q4=E)
- ✅ Mutation testing ≥95% minimum (Phase 5 ONLY - per Q3=E)

---

## 9. References

**V8 Documentation**:
- `docs/fixes-needed-plan-tasks-v8/plan.md` - Implementation plan (5 phases, 16 tasks)
- `docs/fixes-needed-plan-tasks-v8/tasks.md` - Detailed task breakdown

**Architecture**:
- `.github/instructions/02-01.architecture.instructions.md`
- `.github/instructions/02-02.service-template.instructions.md`
- `.github/instructions/03-08.server-builder.instructions.md`

**Barrier Implementations**:
- `internal/apps/template/service/server/barrier/` - TARGET (GORM-based)
- `internal/shared/barrier/` - TO BE DELETED after KMS migration
- `internal/kms/server/barrier/` - TO BE DELETED (unused adapter)

---

## Summary

**Current State** (verified 2026-02-03):
- KMS: Uses ServerBuilder but still imports shared/barrier (4 files) - migration incomplete
- Service-Template: Reference, 98.91% mutation, provides template barrier
- Cipher-IM: Template-based, fully uses template barrier via ServerBuilder
- JOSE-JA: Template-based, uses ServerBuilder with template barrier

**V8 Approach** (quizme decisions):
1. Q1=E: Single barrier in template only
2. Q2=E: Delete shared/barrier IMMEDIATELY after KMS migration
3. Q3=E: Full testing scope (unit+integration+E2E per phase, mutations at end)
4. Q4=E: Incremental doc updates for ACTUALLY-WRONG instructions only

**V8 Priorities**: 5 phases, 16 tasks - sequential execution with quality gates
