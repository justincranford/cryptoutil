# Implementation Plan - Unified Service-Template Migration (V7)

**Status**: Planning
**Created**: 2026-02-02
**Last Updated**: 2026-02-02
**Purpose**: Properly unify sm-kms, cipher-im, and jose-ja on service-template

## Executive Summary

V6 Phase 13 created OPTIONAL abstraction modes (DisabledDatabaseConfig, DisabledBarrierConfig, etc.) instead of enforcing MANDATORY patterns. This fragmented the architecture rather than unifying it.

**V7 Goal**: All services use the SAME foundation with NO optional disabled modes:
- GORM database (not SQLRepository)
- JWT authentication with realms
- OpenAPI strict server
- Template barrier (not shared/barrier)
- Template migrations (1001-1999) + domain migrations (2001+)

## Technical Context

### Current State (Post-V6)

| Service | Database | Auth | OpenAPI | Barrier | Migrations |
|---------|----------|------|---------|---------|------------|
| **cipher-im** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |
| **jose-ja** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |
| **sm-kms** | SQLRepository ❌ | Basic HTTP ❌ | Manual ❌ | shared/barrier ❌ | Custom ❌ |

### Target State (Post-V7)

| Service | Database | Auth | OpenAPI | Barrier | Migrations |
|---------|----------|------|---------|---------|------------|
| **cipher-im** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |
| **jose-ja** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |
| **sm-kms** | GORM ✅ | JWT/Realms ✅ | Strict ✅ | Template ✅ | Template+Domain ✅ |

## Architectural Principles

### MANDATORY Components (No Exceptions)

1. **GORM Database Layer**
   - PostgreSQL OR SQLite support via GORM
   - NO raw database/sql (except within GORM)
   - Connection pooling, WAL mode, busy timeout (SQLite)
   - MaxOpenConns=5 for GORM transaction patterns

2. **JWT Authentication with Realms**
   - `/service/**` paths: Bearer token authentication
   - `/browser/**` paths: Session cookie authentication
   - Multi-tenant realm isolation
   - Session management via template services

3. **OpenAPI Strict Server**
   - oapi-codegen generated handlers
   - Type-safe request/response validation
   - Consistent error handling
   - SwaggerUI integration

4. **Template Barrier Service**
   - Encryption-at-rest via template's GORM-based barrier
   - Unseal key derivation via HKDF
   - Key hierarchy: unseal → root → intermediate → content

5. **Unified Migration System**
   - Template migrations: 1001-1999 (sessions, barrier, realms, tenants)
   - Domain migrations: 2001+ (service-specific tables)
   - golang-migrate with merged filesystem

### ServerBuilder Simplification

**Remove V6 Optional Modes**:
- ❌ DisabledDatabaseConfig → Database is MANDATORY
- ❌ DisabledBarrierConfig → Barrier is MANDATORY
- ❌ DisabledMigrationConfig → Migrations are MANDATORY
- ❌ JWTAuthDisabled → JWT Auth is MANDATORY (or session auth)
- ❌ RawSQLMode/DualMode → GORM only

**Keep ServerBuilder Core**:
- ✅ TLS configuration
- ✅ Admin + Public HTTPS servers
- ✅ Health endpoints (livez, readyz, shutdown)
- ✅ Domain migrations registration
- ✅ Route registration callbacks
- ✅ Telemetry integration

## Phases

### Phase 0: Research & Discovery (2h)

**Objective**: Understand KMS data model and migration path to GORM

- Analyze KMS SQLRepository queries for GORM equivalents
- Identify KMS-specific barrier requirements vs template barrier
- Document KMS authentication requirements
- Map KMS API endpoints to OpenAPI spec

### Phase 1: Remove V6 Optional Modes (4h)

**Objective**: Clean up ServerBuilder to remove disabled modes

- Remove DisabledDatabaseConfig, DisabledBarrierConfig, DisabledMigrationConfig
- Remove JWTAuthDisabled mode
- Remove RawSQLMode, DualMode database options
- Verify cipher-im and jose-ja still work (they use proper modes)
- Remove KMS builder_adapter.go (temporary V6 hack)

### Phase 2: KMS Data Migration (8h)

**Objective**: Migrate KMS from SQLRepository to GORM

- Create KMS GORM models for existing tables
- Implement GORM repositories matching KMS interfaces
- Create domain migrations (2001+) for KMS tables
- Migrate KMS business logic to use GORM repositories
- Remove SQLRepository and raw database/sql code

### Phase 3: KMS Authentication Migration (6h)

**Objective**: Migrate KMS to JWT/Realm authentication

- Define KMS realms and tenant structure
- Implement JWT middleware for KMS API endpoints
- Add session management for browser paths
- Update KMS API handlers for realm context
- Configure /service/** and /browser/** path separation

### Phase 4: KMS OpenAPI Migration (4h)

**Objective**: Migrate KMS to OpenAPI strict server pattern

- Generate KMS OpenAPI spec from existing API
- Use oapi-codegen for strict server handlers
- Migrate existing handlers to strict interface
- Add SwaggerUI for KMS
- Update client generation

### Phase 5: KMS Barrier Migration (4h)

**Objective**: Migrate KMS from shared/barrier to template barrier

- Analyze KMS unseal/seal workflows
- Integrate template barrier with KMS key hierarchy
- Migrate encryption operations to template barrier
- Remove shared/barrier usage from KMS
- Update KMS unseal service integration

### Phase 6: Integration & Testing (6h)

**Objective**: Verify unified architecture works

- All KMS tests pass with new architecture
- All cipher-im tests pass (regression check)
- All jose-ja tests pass (regression check)
- E2E tests for multi-service deployment
- Performance benchmarks vs V6 baseline

### Phase 7: Documentation & Cleanup (2h)

**Objective**: Update documentation and clean up

- Update server-builder.instructions.md
- Update service-template.instructions.md
- Remove obsolete V6 abstraction documentation
- Add migration guide for future services

## Technical Decisions

### Decision 1: KMS Data Model Migration Strategy
- **Options**: 
  A. Incremental (table-by-table)
  B. Big-bang (all at once)
  C. Shadow mode (dual-write then cutover)
- **Needs Research**: Phase 0 will determine best approach

### Decision 2: KMS Barrier Integration
- **Options**:
  A. Replace shared/barrier completely with template barrier
  B. Keep shared/barrier but have template barrier wrap it
  C. Merge shared/barrier into template barrier
- **Needs Research**: Phase 0 will determine compatibility

### Decision 3: KMS API Versioning During Migration
- **Options**:
  A. Breaking change (v2 API)
  B. Backward compatible (v1 with extensions)
  C. Dual support during transition
- **Needs Research**: Client impact analysis

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| KMS data migration breaks existing data | Medium | High | Shadow mode, rollback plan, extensive testing |
| Authentication migration breaks clients | Medium | High | Version API, gradual rollout, client notification |
| Barrier migration corrupts encrypted data | Low | Critical | Test with non-prod data, backup before migration |
| Performance regression | Low | Medium | Benchmark before/after, optimize critical paths |
| cipher-im/jose-ja regression | Low | Medium | Run full test suite after each phase |

## Quality Gates

Each phase MUST pass:
- ✅ All existing tests pass (no regressions)
- ✅ New code has ≥95% coverage
- ✅ Linting clean (`golangci-lint run`)
- ✅ No new TODOs without tracking
- ✅ Mutation testing ≥95% efficacy

Final gate:
- ✅ All three services (kms, cipher-im, jose-ja) use identical ServerBuilder
- ✅ No optional/disabled modes in ServerBuilder
- ✅ E2E tests pass for multi-service deployment
- ✅ Documentation updated

## Success Criteria

- [ ] KMS uses GORM (not SQLRepository)
- [ ] KMS uses JWT/Realm authentication
- [ ] KMS uses OpenAPI strict server
- [ ] KMS uses template barrier
- [ ] KMS uses template+domain migrations
- [ ] ServerBuilder has NO disabled/optional modes
- [ ] All three services are architecturally identical
- [ ] All tests pass across all services
- [ ] Documentation reflects unified architecture

## Estimated Total LOE

| Phase | Estimated | Actual |
|-------|-----------|--------|
| Phase 0 | 2h | |
| Phase 1 | 4h | |
| Phase 2 | 8h | |
| Phase 3 | 6h | |
| Phase 4 | 4h | |
| Phase 5 | 4h | |
| Phase 6 | 6h | |
| Phase 7 | 2h | |
| **Total** | **36h** | |

## Dependencies

- cipher-im: Reference implementation (already correct)
- jose-ja: Reference implementation (already correct)
- service-template: Foundation for all services
- V6 work: Must be cleaned up before V7 can proceed
