# JOSE-JA Refactoring Plan v2 (Quiz-Informed)

**Last Updated**: 2026-01-16
**Based On**: Quiz answers from JOSE-JA-QUIZME.md

## Executive Summary

Refactor jose-ja to:
1. **Eliminate ~459 lines of duplicated infrastructure** using service-template ServerBuilder pattern
2. **Add MANDATORY database persistence** (ALL services MUST use database per service-template requirements)
3. **Implement multi-tenancy** (tenant_id + realm_id isolation)
4. **Add comprehensive audit logging** (all operations, per-tenant configurable)
5. **Adopt /browser/** and `/service/**` path split (MANDATORY for all services)

**Key Decisions from Quiz**:

| Decision Area | Choice | Rationale |
|---------------|--------|-----------|
| Private Key Storage | Barrier-encrypted JWE (like cipher-im/sm-kms) | Consistent with existing services |
| Public Key Storage | Barrier-encrypted JWE | User requested encryption for both |
| Key Metadata | JWE encrypted blob + minimal metadata (kid, timestamp) | Follows cipher-im pattern |
| Multi-Tenancy | MANDATORY (tenant_id + realm_id) | Service-template requirement |
| Default Tenant | NO defaults - users register/join tenants | Service-template registration flow |
| Audit Logging | ALL operations, per-tenant configurable | Compliance requirement |
| Audit Attribution | Link to user/client PK + session PK | Service-template integration |
| Migration Strategy | BREAKING CHANGE (no backward compat) | Alpha project, clean slate |
| Export/Import | CLI: `jose-ja client export/import` | Service-template command pattern |
| SessionManager | MANDATORY integration | ALL service-template features required |
| Path Split | `/browser/**` + `/service/**` MANDATORY | Service-template requirement |
| JWKS Path | `/service/api/v1/jose/.well-known/jwks.json` | Consistent routing + per-JWK scoping |
| Test Isolation | UUIDv7 for all test data | Parallel test safety |
| Mock Strategy | Real database (SQLite), NO mocks | Test real behavior |
| Caching | NO caching (always query database) | Simplicity, consistency |
| Connection Pool | Inherit from service-template | Avoid duplication |
| Phase 1 Scope | Full JWK lifecycle + JWS/JWE/JWT | Production-ready completeness |
| Builder Timing | Incremental (DB/telemetry Phase 1, admin Phase 2) | Gradual adoption |
| Highest Risk | Private key encryption with barrier | Data loss potential |
| Rollback | Not applicable (alpha project) | No production deployments |
| Service-Template Features | ALL features MANDATORY (no exceptions) | Baseline for all 9 services |
| Documentation | Plan + tasks only (no product docs yet) | Incremental documentation |

## Strategic Objectives

### Code Reuse Objectives (UNCHANGED)

1. **Eliminate TLS Duplication**: ~80 lines
2. **Eliminate Admin Server**: admin.go (259 lines)
3. **Eliminate Application Wrapper**: ~80 lines
4. **Eliminate Infrastructure Init**: ~40 lines

**Total**: ~459 lines (29% reduction)

### Architecture Objectives (UPDATED)

1. **MANDATORY Database Persistence**: Service-template requires ALL services use database
2. **MANDATORY Multi-Tenancy**: tenant_id + realm_id isolation per service-template
3. **MANDATORY SessionManager**: Integration required for all services
4. **MANDATORY Path Split**: `/browser/**` + `/service/**` routing
5. **Comprehensive Audit Logging**: All operations, per-tenant configurable
6. **Elastic JWK Pattern**: Proxy JWK containing time-ordered material JWKs
7. **Per-JWK JWKS Scoping**: Each elastic JWK has own JWKS endpoint

### Quality Objectives (UNCHANGED)

1. Zero linting errors
2. Zero build errors
3. ≥95% coverage (production), ≥98% (infrastructure)
4. Backward compatibility NOT required (alpha project)

## CRITICAL Architecture Changes from Quiz

### 1. Multi-Tenancy is MANDATORY

**OLD Assumption**: jose-ja could be single-tenant
**CORRECTED**: ALL services MUST support multi-tenancy per service-template

**Database Schema Changes**:
```sql
CREATE TABLE IF NOT EXISTS jwks (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,  -- NEW: tenant isolation
    realm_id TEXT NOT NULL,   -- NEW: realm isolation
    kid TEXT NOT NULL,        -- Elastic JWK ID (UUIDv7)
    kty TEXT NOT NULL,
    alg TEXT NOT NULL,
    use TEXT NOT NULL,
    private_jwk_jwe TEXT NOT NULL,  -- UPDATED: Barrier-encrypted JWE
    public_jwk_jwe TEXT NOT NULL,   -- UPDATED: Barrier-encrypted JWE (even for public)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
    UNIQUE(tenant_id, realm_id, kid)
);

CREATE INDEX IF NOT EXISTS idx_jwks_tenant_realm ON jwks(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_jwks_kid ON jwks(kid);
CREATE INDEX IF NOT EXISTS idx_jwks_use ON jwks(use);
```

### 2. No Default Tenant Pattern

**OLD Assumption**: Use `cryptoutilMagic.JoseDefaultTenantID`
**CORRECTED**: NO default tenants - users register or join tenants via service-template flow

**Service-Template Registration Flow**:
1. User/client registers via `/browser/api/v1/auth/register` or `/service/api/v1/auth/register`
2. Choose: Create new tenant OR join existing tenant
3. If new tenant: User becomes admin, others request to join (requires admin approval)
4. If join existing: Requires admin authorization

**Implication**: jose-ja MUST NOT hardcode default tenant creation in migrations

### 3. Audit Logging - All Operations, Per-Tenant Configurable

**Schema**:
```sql
CREATE TABLE IF NOT EXISTS jwk_audit_log (
    id TEXT PRIMARY KEY NOT NULL,
    tenant_id TEXT NOT NULL,
    realm_id TEXT NOT NULL,
    operation TEXT NOT NULL,  -- generate, get, list, delete, sign, verify, encrypt, decrypt
    kid TEXT,
    user_id TEXT NOT NULL,     -- FK to users table (service-template)
    session_id TEXT NOT NULL,  -- FK to sessions table (service-template)
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT,
    FOREIGN KEY (tenant_id, realm_id) REFERENCES tenant_realms(tenant_id, realm_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_tenant_realm ON jwk_audit_log(tenant_id, realm_id);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_timestamp ON jwk_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_kid ON jwk_audit_log(kid);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_user ON jwk_audit_log(user_id);
```

**Per-Tenant Configuration**:
```sql
CREATE TABLE IF NOT EXISTS tenant_audit_config (
    tenant_id TEXT NOT NULL,
    operation TEXT NOT NULL,  -- generate, get, list, delete, sign, verify, encrypt, decrypt
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (tenant_id, operation),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

### 4. Path Routing MANDATORY Change

**OLD**: `/jose/v1/**`
**NEW**: `/browser/api/v1/jose/**` + `/service/api/v1/jose/**`

**Middleware Differences**:
- `/browser/**`: CSRF, CORS, CSP, session cookies
- `/service/**`: API key/bearer tokens, no CSRF

**JWKS Path Change**:
- OLD: `/.well-known/jwks.json`
- NEW: `/service/api/v1/jose/{kid}/.well-known/jwks.json` (per elastic JWK)

**Rationale**: Each elastic JWK (proxy JWK) contains multiple material JWKs. JWKS endpoint scoped per elastic JWK returns only public material JWKs for that elastic JWK. If elastic JWK is secret/symmetric, JWKS is empty.

### 5. Elastic JWK Pattern (New Concept)

**Elastic JWK** = Proxy JWK containing time-ordered list of material JWKs

**Database Model**:
```sql
CREATE TABLE IF NOT EXISTS jwk_materials (
    id TEXT PRIMARY KEY NOT NULL,
    elastic_jwk_id TEXT NOT NULL,  -- FK to jwks(id)
    material_kid TEXT NOT NULL,     -- Material JWK KID (UUIDv7)
    private_jwk_jwe TEXT NOT NULL,  -- Barrier-encrypted material private JWK
    public_jwk_jwe TEXT NOT NULL,   -- Barrier-encrypted material public JWK
    active BOOLEAN NOT NULL DEFAULT FALSE,  -- Only 1 active per elastic JWK
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    retired_at TIMESTAMP,
    FOREIGN KEY (elastic_jwk_id) REFERENCES jwks(id),
    UNIQUE(elastic_jwk_id, material_kid)
);

CREATE INDEX IF NOT EXISTS idx_jwk_materials_elastic ON jwk_materials(elastic_jwk_id);
CREATE INDEX IF NOT EXISTS idx_jwk_materials_active ON jwk_materials(elastic_jwk_id, active);
```

**Usage**:
- **Sign/Encrypt**: Use active material JWK, embed material_kid in output
- **Verify/Decrypt**: Lookup material JWK by embedded material_kid
- **Rotation**: Generate new material JWK, set as active, retire old (keep for decryption/verification)

### 6. SessionManager Integration MANDATORY

**OLD**: jose-ja stateless API
**NEW**: SessionManager required for both `/browser/**` and `/service/**` paths

**SessionManager Usage**:
- Browser users: Session cookies via `/browser/api/v1/auth/login`
- Service clients: API key sessions via `/service/api/v1/auth/authenticate`
- Rate limiting: Per-session tracking
- Audit logs: Link to session_id

## Refactoring Phases (UPDATED)

### Phase 1: Database Schema & Repository (4-5 days)

**NEW Tasks**:
1. Create migrations with multi-tenancy (tenant_id, realm_id)
2. Create elastic_jwk + jwk_materials tables
3. Create tenant_audit_config table
4. Create domain models (JWK, JWKMaterial, AuditLogEntry)
5. Create repositories (JWKRepository, JWKMaterialRepository, AuditRepository, AuditConfigRepository)
6. Unit tests (≥98% coverage)

**Validation**:
- ✅ Migrations apply (PostgreSQL + SQLite)
- ✅ Multi-tenant isolation enforced (tenant_id + realm_id FK)
- ✅ Elastic JWK pattern supported
- ✅ Repository tests pass (≥98% coverage)

### Phase 2: ServerBuilder Integration (3-4 days)

**NEW Tasks**:
1. Create JoseSettings (extends ServiceTemplateServerSettings)
2. Refactor server.go with builder
3. Register domain migrations via `builder.WithDomainMigrations()`
4. Register public routes via `builder.WithPublicRouteRegistration()`
5. Implement SessionManager integration
6. Update tests

**Validation**:
- ✅ Builder creates complete infrastructure
- ✅ SessionManager integrated
- ✅ Both `/browser/**` and `/service/**` paths functional
- ✅ All tests pass

### Phase 3: Multi-Tenancy & Session Integration (3-4 days)

**NEW Tasks**:
1. Implement tenant registration flow (create tenant or join existing)
2. Implement admin authorization for join requests
3. Add tenant_id + realm_id to all JWK operations
4. Add session_id extraction middleware
5. Update handlers to enforce tenant isolation
6. Tests for multi-tenant scenarios

**Validation**:
- ✅ Tenant registration works
- ✅ JWKs isolated by tenant + realm
- ✅ Cross-tenant access blocked
- ✅ Session tracking works

### Phase 4: Audit Logging with Per-Tenant Config (2-3 days)

**NEW Tasks**:
1. Implement audit config service (per-tenant operation toggle)
2. Add audit logging to ALL JWK operations
3. Link audit logs to user_id + session_id
4. Admin API for audit config management
5. Tests for audit logging

**Validation**:
- ✅ All operations logged (when enabled)
- ✅ Per-tenant config works
- ✅ Audit logs link to users + sessions
- ✅ Compliance requirements met

### Phase 5: Elastic JWK Implementation (3-4 days)

**NEW Tasks**:
1. Implement elastic JWK service
2. Implement material JWK rotation
3. Update sign/encrypt to use active material
4. Update verify/decrypt to lookup by material_kid
5. Per-JWK JWKS endpoint: `/service/api/v1/jose/{kid}/.well-known/jwks.json`
6. Tests for key rotation

**Validation**:
- ✅ Elastic JWK pattern works
- ✅ Material rotation works
- ✅ Sign/verify with historical material works
- ✅ Per-JWK JWKS endpoint works

### Phase 6: Path Migration & Middleware (2-3 days)

**NEW Tasks**:
1. Migrate all endpoints to `/browser/**` and `/service/**`
2. Add CSRF middleware to `/browser/**` paths
3. Add CORS middleware to `/browser/**` paths
4. Update OpenAPI specs for new paths
5. Tests for middleware behavior

**Validation**:
- ✅ All endpoints migrated
- ✅ CSRF protection works on `/browser/**`
- ✅ `/service/**` has no CSRF (correct)
- ✅ OpenAPI specs updated

### Phase 7: Integration & E2E Testing (3-4 days)

**Tasks**:
1. E2E: Full JWK lifecycle (multi-tenant)
2. E2E: Elastic JWK rotation
3. E2E: Multi-instance deployment
4. E2E: Audit log verification
5. E2E: SessionManager integration
6. Load testing (Gatling)

**Validation**:
- ✅ All E2E tests pass
- ✅ Multi-instance works
- ✅ Performance acceptable
- ✅ Audit logs complete

### Phase 8: Documentation & Cleanup (2-3 days)

**Tasks**:
1. Create migration guide
2. Update API documentation
3. Update deployment guides
4. Final cleanup (linting, TODOs)

**Validation**:
- ✅ Documentation complete
- ✅ No deprecated code
- ✅ All quality gates pass

## Timeline Estimate (UPDATED)

| Phase | Duration | Risk |
|-------|----------|------|
| Phase 1: Database Schema & Repository | 4-5 days | Low |
| Phase 2: ServerBuilder Integration | 3-4 days | Low |
| Phase 3: Multi-Tenancy & Sessions | 3-4 days | Medium |
| Phase 4: Audit Logging | 2-3 days | Low |
| Phase 5: Elastic JWK | 3-4 days | Medium |
| Phase 6: Path Migration | 2-3 days | Low |
| Phase 7: E2E Testing | 3-4 days | High |
| Phase 8: Documentation | 2-3 days | Low |
| **TOTAL** | **22-30 days** | Medium |

**Note**: 50% longer than original estimate due to mandatory multi-tenancy, elastic JWK pattern, session integration requirements discovered in quiz.

## Risk Assessment (UPDATED)

### CRITICAL Risks (NEW)

1. **Multi-Tenant Data Isolation**: Cross-tenant data leakage would be security incident
   - **Mitigation**: Row-level security tests, tenant_id enforcement in all queries

2. **Elastic JWK Rotation**: Material key management complexity
   - **Mitigation**: Comprehensive rotation tests, material_kid tracking

3. **Session Integration**: SessionManager dependency on service-template correctness
   - **Mitigation**: Integration tests with service-template, verify session lifecycle

### High Risks (UNCHANGED)

1. **Barrier Encryption**: Private key encryption/decryption correctness
2. **Multi-Instance Coordination**: Database locking, race conditions

### Medium Risks (UPDATED)

1. **Path Migration**: Existing API contracts change (but alpha project = acceptable)
2. **Performance**: Database queries + barrier encryption overhead
3. **Audit Log Volume**: High-frequency operations (sign/verify) generate large logs

## Success Criteria (UPDATED)

### Functional Gates

- ✅ Multi-tenancy enforced (tenant + realm isolation)
- ✅ Elastic JWK pattern works (rotation + historical material)
- ✅ SessionManager integrated (browser + service paths)
- ✅ Audit logging complete (all operations, per-tenant config)
- ✅ Path split implemented (`/browser/**` + `/service/**`)
- ✅ Per-JWK JWKS endpoints functional

### Quality Gates (UNCHANGED)

- ✅ Zero linting/build errors
- ✅ ≥95% coverage (production), ≥98% (infrastructure)
- ✅ ≥85% mutation score (production), ≥98% (infrastructure)

### Documentation Gates (UNCHANGED)

- ✅ Migration guide
- ✅ API documentation
- ✅ Deployment guides

## Open Questions for Round 2 QUIZME

Based on deep analysis, these areas need clarification:

1. **Elastic JWK Material Cleanup**: When to delete retired material JWKs?
2. **Audit Log Retention**: Per-tenant retention policies?
3. **JWKS Endpoint Caching**: Should per-JWK JWKS be cached?
4. **Cross-Tenant Key Sharing**: Should tenants be able to share public keys?
5. **Tenant Registration UI**: Browser UI for tenant management?
6. **Session Timeout**: Different timeouts for browser vs service sessions?
7. **Rate Limiting**: Per-tenant or per-session rate limits?
8. **Barrier Key Rotation**: How to rotate barrier keys without re-encrypting all JWKs?

## Cross-References

- **Quiz Answers**: [JOSE-JA-QUIZME.md](JOSE-JA-QUIZME.md)
- **Service-Template**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **cipher-im Reference**: [internal/apps/cipher/im/](../../internal/apps/cipher/im/)
- **Multi-Tenancy**: [02-10.authn.instructions.md](../../.github/instructions/02-10.authn.instructions.md)
