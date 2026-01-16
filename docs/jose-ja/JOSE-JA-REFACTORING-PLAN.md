# JOSE-JA Refactoring Plan

## Executive Summary

Refactor jose-ja to eliminate **~459 lines of duplicated infrastructure** and add database persistence for **audit compliance** and **multi-instance scalability**. Uses service-template ServerBuilder pattern proven in cipher-im refactoring (49% code reduction).

**Approach**: Add database persistence (Option B from analysis) - maintains consistency with other cryptoutil services and enables production compliance.

## Strategic Objectives

### Code Reuse Objectives

1. **Eliminate TLS Duplication**: Remove 4 manual TLS config generations (~80 lines)
2. **Eliminate Admin Server Duplication**: Remove admin.go (259 lines) - use template AdminServer
3. **Eliminate Application Wrapper Duplication**: Replace manual construction (~80 lines)
4. **Eliminate Infrastructure Init**: Remove manual telemetry/JWKGenService setup (~40 lines)

**Total Duplication Eliminated**: ~459 lines (29% of current codebase)

### Architecture Objectives

1. **Add Database Persistence**: Migrate from in-memory KeyStore to database (audit compliance)
2. **Enable Multi-Instance Deployment**: Shared database allows multiple jose-ja instances
3. **Audit Logging**: Track all JWK operations (regulatory compliance: SOC2, PCI-DSS, HIPAA)
4. **Consistent Structure**: Match cipher-im/identity/kms architecture patterns

### Quality Objectives

1. **Zero Linting Errors**: All code passes `golangci-lint run`
2. **Zero Build Errors**: `go build ./...` clean
3. **Test Coverage**: Maintain ≥95% coverage (production code)
4. **Backward Compatibility**: Existing JOSE API contracts unchanged

## Refactoring Phases

### Phase 1: Database Schema & Repository (Foundation)

**Objective**: Create database tables and repository layer WITHOUT changing existing server code.

**Duration**: 2-3 days

**Tasks**:
1. Create migration files (2001_jose_jwks, 2002_jose_audit_log)
2. Create domain models (JWK, AuditLogEntry)
3. Create repository interfaces (JWKRepository, AuditLogRepository)
4. Create GORM repository implementations
5. Unit tests for repositories (≥98% coverage target for infrastructure)

**Validation Criteria**:
- ✅ Migrations apply cleanly (PostgreSQL + SQLite)
- ✅ Repository tests pass (≥98% coverage)
- ✅ Existing server code still compiles (unchanged)
- ✅ `golangci-lint run` clean

**Files Created**:
```
internal/jose/repository/
├── migrations/
│   ├── 2001_jose_jwks.up.sql
│   ├── 2001_jose_jwks.down.sql
│   ├── 2002_jose_audit_log.up.sql
│   └── 2002_jose_audit_log.down.sql
├── migrations.go (embed.FS wrapper)
├── jwk_repository.go
├── jwk_repository_test.go
├── audit_repository.go
└── audit_repository_test.go

internal/jose/domain/
├── jwk.go (JWK model implements domain.JWKModel interface)
└── audit.go (AuditLogEntry model)
```

**Risk Mitigation**: Phase 1 is ADDITIVE only - no changes to existing server code, safe to merge incrementally.

### Phase 2: ServerBuilder Integration (Infrastructure)

**Objective**: Replace manual TLS/telemetry/JWKGenService setup with builder pattern.

**Duration**: 2-3 days

**Tasks**:
1. Create refactored server.go using ServerBuilder
2. Replace manual TLS with builder-generated TLS
3. Replace manual telemetry with builder-injected telemetry
4. Replace manual JWKGenService with builder-injected JWKGenService
5. Remove deprecated `New()` method (use `NewFromConfig` instead)
6. Update server constructor tests

**Validation Criteria**:
- ✅ Server builds without errors
- ✅ TLS certificates valid (curl test passes)
- ✅ Telemetry logs appear (OTLP export working)
- ✅ JWKGenService generates keys correctly
- ✅ All server tests pass

**Files Modified**:
```
internal/jose/server/
├── config/
│   └── jose_settings.go (create, extends ServiceTemplateServerSettings)
├── server.go (refactor to use builder - similar to cipher-im pattern)
└── server_test.go (update for new constructor)
```

**Risk Mitigation**: Keep old `server.go` as `server_old.go` during transition, validate before deleting.

### Phase 3: Admin Server Elimination (Deduplication)

**Objective**: Delete admin.go (259 lines) and use template AdminServer directly.

**Duration**: 1-2 days

**Tasks**:
1. Verify template AdminServer provides all required endpoints
2. Update Application to use template AdminServer
3. Delete internal/jose/server/admin.go
4. Update admin server tests (use template test patterns)

**Validation Criteria**:
- ✅ `/admin/v1/livez` returns 200 OK
- ✅ `/admin/v1/readyz` returns 200 OK (after initialization)
- ✅ `/admin/v1/shutdown` triggers graceful shutdown
- ✅ Admin server binds to 127.0.0.1:9090
- ✅ All admin tests pass

**Files Deleted**:
```
internal/jose/server/
└── admin.go (259 lines) - COMPLETE REMOVAL
```

**Files Modified**:
```
internal/jose/server/
└── application.go (update to use template AdminServer)
```

**Risk Mitigation**: Verify template AdminServer behavior matches jose-ja requirements before deletion.

### Phase 4: KeyStore Migration (Persistence Layer)

**Objective**: Replace in-memory KeyStore with database-backed repository.

**Duration**: 3-4 days

**Tasks**:
1. Create database-backed JWK service (uses repository from Phase 1)
2. Encrypt private keys with BarrierService before storage
3. Update handlers.go to use repository instead of KeyStore
4. Add audit logging to all JWK operations
5. Migrate existing KeyStore tests to repository tests
6. Delete keystore.go (replaced by repository)

**Validation Criteria**:
- ✅ JWK generation persists to database
- ✅ Private keys encrypted with barrier (verify ciphertext)
- ✅ JWK retrieval decrypts correctly
- ✅ Audit log entries created for all operations
- ✅ JWKS endpoint returns all public keys from database
- ✅ Multi-instance test passes (shared database)
- ✅ All handler tests pass

**Files Modified**:
```
internal/jose/server/
├── handlers.go (replace keyStore with jwkRepository + auditRepository)
└── handlers_test.go (update for database-backed storage)
```

**Files Deleted**:
```
internal/jose/server/
└── keystore.go (118 lines) - REPLACED by repository
```

**Risk Mitigation**: Run integration tests with both KeyStore and Repository in parallel during transition.

### Phase 5: Application Wrapper Refactor (Lifecycle)

**Objective**: Use builder-created Application wrapper instead of manual construction.

**Duration**: 1-2 days

**Tasks**:
1. Update application.go to use builder-created Application
2. Remove manual TLS config generation (2 instances)
3. Remove manual server construction
4. Update application tests

**Validation Criteria**:
- ✅ Application.Start() launches both servers
- ✅ Application.Shutdown() stops both servers cleanly
- ✅ PublicPort() and AdminPort() accessors work
- ✅ Error channel aggregation works correctly
- ✅ All application tests pass

**Files Modified**:
```
internal/jose/server/
├── application.go (simplified to use builder pattern)
└── application_test.go (update for builder pattern)
```

**Risk Mitigation**: Validate error handling paths (server startup failures, shutdown timeouts).

### Phase 6: Integration Testing (E2E Validation)

**Objective**: Validate complete refactored system with E2E tests.

**Duration**: 2-3 days

**Tasks**:
1. Create E2E test suite (similar to cipher-im pattern)
2. Test full JWK lifecycle (generate → get → sign/encrypt → delete)
3. Test multi-instance deployment (shared database)
4. Test audit log completeness
5. Test barrier encryption (verify private keys encrypted)
6. Test backward compatibility (existing API contracts)
7. Load testing (Gatling scripts)

**Validation Criteria**:
- ✅ E2E tests pass (both SQLite + PostgreSQL)
- ✅ Multi-instance test passes (2+ jose-ja instances share database)
- ✅ Audit log contains all operations with correct metadata
- ✅ Private keys never appear in plaintext logs
- ✅ Existing API clients work without changes
- ✅ Load test passes (1000 req/s for 60s)
- ✅ Docker Compose deployment works

**Files Created**:
```
test/e2e/jose/
├── jose_e2e_test.go
├── multi_instance_test.go
└── audit_test.go

test/load/jose/
└── JoseLoadTest.scala (Gatling)
```

**Risk Mitigation**: Run E2E tests in CI/CD before merging to main.

### Phase 7: Documentation & Cleanup (Finalization)

**Objective**: Update documentation and remove deprecated code.

**Duration**: 1-2 days

**Tasks**:
1. Update README.md (new architecture, database requirement)
2. Update API documentation (OpenAPI specs)
3. Update deployment guides (Docker Compose, Kubernetes)
4. Create migration guide (for existing deployments)
5. Remove deprecated code (`server_old.go`, old tests)
6. Final linting pass

**Validation Criteria**:
- ✅ All documentation updated
- ✅ Migration guide tested (upgrade path works)
- ✅ No deprecated code remains
- ✅ `golangci-lint run` clean
- ✅ `go build ./...` clean
- ✅ All tests pass (unit + integration + E2E)

**Files Updated**:
```
docs/jose-ja/
├── MIGRATION-GUIDE.md (create)
├── README.md (update)
└── API.md (update)

deployments/jose/
├── compose.yml (update for database requirement)
└── kubernetes/ (update manifests)
```

**Risk Mitigation**: Test migration guide with actual jose-ja deployment upgrade.

## Risk Assessment

### High Risk Areas

1. **Database Migration**: First time jose-ja uses database - ensure connection pooling, transaction handling correct
   - **Mitigation**: Use exact patterns from cipher-im (proven to work)

2. **Barrier Encryption**: Private JWKs must be encrypted - verify no plaintext leaks
   - **Mitigation**: Add specific tests for ciphertext verification, audit log scanning

3. **Multi-Instance Coordination**: Database shared across instances - prevent race conditions
   - **Mitigation**: Use GORM transactions, optimistic locking for key rotation

4. **Backward Compatibility**: Existing API contracts must not change
   - **Mitigation**: Contract tests, versioned API endpoints

### Medium Risk Areas

1. **Performance Regression**: Database queries slower than in-memory
   - **Mitigation**: Add caching layer, benchmark tests (baseline vs refactored)

2. **Test Coverage Gaps**: New repository layer needs ≥98% coverage
   - **Mitigation**: Incremental coverage analysis, mutation testing

3. **Docker Compose Changes**: Database dependency changes deployment
   - **Mitigation**: Update compose files in parallel with code, test before merge

### Low Risk Areas

1. **TLS/Telemetry**: Well-tested in cipher-im refactoring
2. **Admin Server**: Template pattern proven to work
3. **Application Wrapper**: Minimal changes (builder handles complexity)

## Success Criteria

### Code Quality Gates

- ✅ **Zero Linting Errors**: `golangci-lint run` clean
- ✅ **Zero Build Errors**: `go build ./...` clean
- ✅ **Test Coverage**: ≥95% production code, ≥98% repository/infrastructure
- ✅ **Mutation Score**: ≥85% production code, ≥98% repository/infrastructure
- ✅ **No Skipped Tests**: All tests enabled and passing

### Functional Gates

- ✅ **All JOSE Operations Work**: JWK/JWS/JWE/JWT endpoints functional
- ✅ **Audit Logging Complete**: All operations logged with metadata
- ✅ **Multi-Instance Support**: 2+ instances share database without conflicts
- ✅ **Backward Compatibility**: Existing API clients work without changes
- ✅ **Performance Acceptable**: <10% regression vs baseline

### Documentation Gates

- ✅ **Migration Guide**: Upgrade path documented and tested
- ✅ **API Documentation**: OpenAPI specs updated
- ✅ **Deployment Guides**: Docker Compose + Kubernetes updated
- ✅ **README**: Architecture diagram updated

## Rollback Plan

### Phase 1-2 Rollback (Low Risk)
- **Action**: Delete new files (repository, migrations), revert server.go changes
- **Impact**: None (old server still functional)
- **Duration**: <1 hour

### Phase 3-4 Rollback (Medium Risk)
- **Action**: Restore admin.go, keystore.go from git history
- **Impact**: Manual TLS setup restored
- **Duration**: 2-4 hours (requires testing)

### Phase 5-7 Rollback (High Risk)
- **Action**: Full git revert to pre-refactor commit
- **Impact**: All refactoring work lost
- **Duration**: 1 day (requires full regression testing)

**Recommendation**: Merge phases incrementally (1-2 → 3 → 4 → 5-7) to minimize rollback risk.

## Timeline Estimate

| Phase | Duration | Dependencies | Risk Level |
|-------|----------|--------------|------------|
| Phase 1: Database Schema & Repository | 2-3 days | None | Low |
| Phase 2: ServerBuilder Integration | 2-3 days | Phase 1 | Low |
| Phase 3: Admin Server Elimination | 1-2 days | Phase 2 | Low |
| Phase 4: KeyStore Migration | 3-4 days | Phase 1, 3 | Medium |
| Phase 5: Application Wrapper Refactor | 1-2 days | Phase 2, 3, 4 | Medium |
| Phase 6: Integration Testing | 2-3 days | Phase 5 | High |
| Phase 7: Documentation & Cleanup | 1-2 days | Phase 6 | Low |
| **TOTAL** | **12-19 days** | Sequential | Medium |

**Note**: Timeline assumes single developer working full-time. Parallel work on independent phases could reduce to 8-12 days.

## Cross-References

- **Analysis Document**: [JOSE-JA-ANALYSIS.md](JOSE-JA-ANALYSIS.md)
- **ServerBuilder Pattern**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **cipher-im Reference**: [internal/apps/cipher/im/server/](../../internal/apps/cipher/im/server/)
- **Template Builder**: [internal/apps/template/service/server/builder/server_builder.go](../../internal/apps/template/service/server/builder/server_builder.go)
