# JOSE-JA Structure Analysis

## Executive Summary

jose-ja is a **stateless cryptographic service** providing JWK/JWS/JWE/JWT operations via in-memory KeyStore. Current architecture has **~709 lines of duplicated infrastructure** (TLS, telemetry, admin server, application wrapper) that can be eliminated via service-template ServerBuilder pattern.

**Critical Architectural Decision**: jose-ja is STATELESS (no database, no migrations) unlike cipher-im (stateful with database/migrations/tenants). This requires either:
- **Option A**: Keep stateless, create lightweight builder variant (no DB/migrations/tenants/sessions/barrier)
- **Option B**: Add database persistence (JWK storage, audit logs), use full ServerBuilder

## Current Architecture (internal/jose/server/)

### File Structure

```
internal/jose/server/
├── server.go           (283 lines) - Main server with manual TLS/telemetry
├── application.go      (167 lines) - Application wrapper (partial template duplication)
├── admin.go            (259 lines) - Admin server (FULL template duplication)
├── handlers.go         (776 lines) - JOSE operation handlers
├── keystore.go         (118 lines) - In-memory JWK storage
├── keystore_test.go
└── server_test.go
```

**Total**: ~1603 lines (excluding tests)

### server.go (283 lines) - Main Server

**Purpose**: Public HTTPS server for JOSE operations

**Key Characteristics**:
- ✅ Uses fiber framework (consistent with template)
- ❌ Manual TLS generation (`GenerateAutoTLSGeneratedSettings` + `GenerateTLSMaterial`)
- ❌ Manual telemetry initialization (`NewTelemetryService`)
- ❌ Manual JWKGenService initialization (`NewJWKGenService`)
- ❌ In-memory KeyStore (no database persistence)
- ✅ Proper shutdown handling (listener close, Fiber shutdown)
- ❌ Deprecated `New()` method (backward compatibility)

**Routes**:
- `/health`, `/livez`, `/readyz` - Health endpoints
- `/.well-known/jwks.json` - Public key discovery (OIDC standard)
- `/jose/v1/jwk/*` - JWK operations (generate, get, delete, list)
- `/jose/v1/jws/*` - JWS operations (sign, verify)
- `/jose/v1/jwe/*` - JWE operations (encrypt, decrypt)
- `/jose/v1/jwt/*` - JWT operations (issue, validate)

**Duplication Pattern**:
```go
// jose-ja server.go (manual TLS)
tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)
tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)

// vs. service-template builder (automated TLS)
builder.Build() // Handles TLS generation automatically
```

### application.go (167 lines) - Application Wrapper

**Purpose**: Manages both public + admin servers with unified lifecycle

**Key Characteristics**:
- ✅ Start() launches both servers in goroutines (template pattern)
- ✅ Shutdown() stops both servers with error aggregation (template pattern)
- ✅ PublicPort() and AdminPort() accessor methods (template pattern)
- ❌ Manual TLS config generation for BOTH servers (duplication)
- ❌ Manual server construction (NewServer + NewAdminHTTPServer calls)

**Duplication Pattern**:
```go
// jose-ja application.go (manual server construction)
publicTLSCfg, _ := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)
adminTLSCfg, _ := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)
publicServer, _ := NewServer(ctx, settings, publicTLSCfg)
adminServer, _ := NewAdminHTTPServer(ctx, settings, adminTLSCfg)

// vs. service-template builder (automated construction)
resources, _ := builder.Build() // Handles BOTH servers + TLS
```

### admin.go (259 lines) - Admin Server

**Purpose**: Private HTTPS server for admin operations (`/admin/v1/livez`, `/readyz`, `/shutdown`)

**Key Characteristics**:
- ❌ **COMPLETE DUPLICATION** of template AdminServer pattern
- ❌ Manual TLS: `GenerateAutoTLSGeneratedSettings` + `GenerateTLSMaterial`
- ❌ Manual Fiber app creation (identical config to template)
- ❌ Routes: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`
- ❌ Readiness state management: `mu.RWMutex`, `ready bool`, `shutdown bool`
- ❌ Identical shutdown handling pattern to template

**Duplication Evidence**:
```go
// jose-ja admin.go (259 lines of duplication)
type AdminServer struct {
    mu       sync.RWMutex
    ready    bool
    shutdown bool
    // ... manual TLS setup, Fiber app, route handlers
}

// vs. service-template AdminServer (already exists)
// IDENTICAL pattern in internal/apps/template/service/server/admin.go
```

**Elimination Opportunity**: admin.go can be COMPLETELY REMOVED by using template AdminServer directly.

### handlers.go (776 lines) - JOSE Operations

**Purpose**: HTTP handlers for JWK/JWS/JWE/JWT operations

**Key Characteristics**:
- ✅ Domain-specific logic (NOT duplication, keep as-is)
- ✅ Uses JWKGenService for key generation
- ✅ Uses KeyStore for in-memory storage
- ✅ Proper error handling and logging
- ✅ Algorithm validation (RSA/EC/OKP/oct with various key sizes)
- ✅ JWE encryption parameters (alg + enc mapping)
- ✅ JWS signature algorithms
- ✅ Serializes public keys for JWKS responses

**NOT Duplication**: This is core business logic, MUST remain in jose-ja.

### keystore.go (118 lines) - In-Memory Storage

**Purpose**: Thread-safe in-memory JWK storage

**Key Characteristics**:
- ✅ Simple map-based storage (`map[string]*StoredKey`)
- ✅ Thread-safe (`sync.RWMutex`)
- ✅ CRUD operations (Store, Get, Delete, List)
- ✅ JWKS generation (GetJWKS returns joseJwk.Set)
- ❌ **NO PERSISTENCE** - all keys lost on restart
- ❌ **NO AUDIT LOGS** - no record of key generation/usage
- ❌ **NO MULTI-INSTANCE** - each instance has separate KeyStore

**Stateless Implications**:
1. **Compliance Risk**: No audit trail for key operations (regulatory requirement for many industries)
2. **Availability Risk**: Restart = lose all keys (downstream services may cache kid references)
3. **Scalability Limitation**: Can't run multiple jose-ja instances (keystore not shared)

**Database Migration Considerations**:
- **Option A (Keep Stateless)**: Accept limitations, use for dev/testing only
- **Option B (Add Database)**:
  - Persist keys to `jwks` table (id, kid, kty, alg, use, private_jwk, public_jwk, created_at)
  - Add audit logs to `jwk_audit_log` table (operation, kid, timestamp, user_id)
  - Enable multi-instance deployment (shared database)
  - Support key rotation with history (keep old keys for signature verification)

## Duplication Analysis

### TLS Infrastructure (COMPLETE DUPLICATION)

**Duplicated in**: server.go (1 instance), application.go (2 instances), admin.go (1 instance) = **4 total TLS config generations**

**Pattern**:
```go
// Repeated 4 times across jose-ja files
tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
    []string{"localhost", "jose-server"},
    []string{"127.0.0.1", "::1"},
    cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
)
if err != nil { return nil, err }

tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)
if err != nil { return nil, err }
```

**Elimination**: ServerBuilder handles TLS for BOTH servers automatically.

### Telemetry Infrastructure (COMPLETE DUPLICATION)

**Duplicated in**: server.go (1 instance)

**Pattern**:
```go
// Manual initialization
telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
if err != nil { return nil, err }
```

**Elimination**: ServerBuilder initializes telemetry automatically.

### JWKGenService Infrastructure (COMPLETE DUPLICATION)

**Duplicated in**: server.go (1 instance)

**Pattern**:
```go
// Manual initialization
jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, settings.VerboseMode)
if err != nil {
    telemetryService.Shutdown()
    return nil, err
}
```

**Elimination**: ServerBuilder initializes JWKGenService automatically.

### Admin Server (COMPLETE DUPLICATION)

**Duplicated in**: admin.go (259 lines - ENTIRE FILE)

**Identical Pattern to Template**:
- Readiness state management (`mu.RWMutex`, `ready`, `shutdown`)
- Routes: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown`
- Fiber app configuration
- TLS listener setup
- Shutdown handling

**Elimination**: Replace `internal/jose/server/admin.go` with template AdminServer import.

### Application Wrapper (PARTIAL DUPLICATION)

**Duplicated in**: application.go (167 lines - PARTIAL)

**Template Pattern Already Used**:
- ✅ Start() with goroutines for both servers
- ✅ Shutdown() with error aggregation
- ✅ PublicPort() and AdminPort() accessors

**Still Duplicated**:
- ❌ Manual TLS config generation (2 instances)
- ❌ Manual server construction (NewServer + NewAdminHTTPServer calls)

**Elimination**: Use template Application pattern (builder creates Application wrapper).

## Quantified Duplication

| Category | Lines | Files | Elimination Strategy |
|----------|-------|-------|---------------------|
| **TLS Generation** | ~80 | server.go, application.go, admin.go | Builder handles automatically |
| **Telemetry Init** | ~20 | server.go | Builder handles automatically |
| **JWKGenService Init** | ~20 | server.go | Builder handles automatically |
| **Admin Server** | 259 | admin.go | Use template AdminServer directly |
| **Application Wrapper** | ~80 | application.go | Builder creates Application |
| **TOTAL** | **~459 lines** | 3 files | ServerBuilder pattern |

**NOTE**: This excludes handlers.go (776 lines) and keystore.go (118 lines) which are domain-specific and MUST remain.

## Comparison: jose-ja vs cipher-im vs service-template

| Aspect | jose-ja (Current) | cipher-im (Refactored) | service-template |
|--------|-------------------|------------------------|------------------|
| **Database** | ❌ None (in-memory) | ✅ GORM (SQLite/PostgreSQL) | ✅ GORM support |
| **Migrations** | ❌ None | ✅ Domain migrations (2001+) | ✅ Template migrations (1001-1004) |
| **Multi-Tenancy** | ❌ None | ✅ Tenant/realm architecture | ✅ Tenant/realm support |
| **Sessions** | ❌ None | ✅ SessionManager | ✅ SessionManager |
| **Barrier Encryption** | ❌ None | ✅ BarrierService | ✅ BarrierService |
| **TLS** | ❌ Manual (4 instances) | ✅ Builder-generated | ✅ Builder-generated |
| **Telemetry** | ❌ Manual | ✅ Builder-generated | ✅ Builder-generated |
| **JWKGenService** | ❌ Manual | ✅ Builder-injected | ✅ Builder-injected |
| **Admin Server** | ❌ Duplicated (259 lines) | ✅ Template-provided | ✅ Provided |
| **Application Wrapper** | ❌ Partial duplication | ✅ Template-provided | ✅ Provided |
| **Total Lines** | ~1603 (before refactor) | 579 → 298 (49% reduction) | N/A (template) |
| **Estimated Post-Refactor** | **~900 lines** (44% reduction) | 298 lines | N/A |

## Architectural Decision Matrix

### Option A: Lightweight Builder (Stateless Services)

**Approach**: Create builder variant that supports stateless services (no DB/migrations/tenants/sessions/barrier).

**Pros**:
- ✅ Matches jose-ja current architecture (no breaking changes)
- ✅ Eliminates TLS/telemetry/admin server duplication
- ✅ Faster startup (no database initialization)
- ✅ Simpler deployment (no database dependencies)

**Cons**:
- ❌ No audit logging (compliance risk)
- ❌ Can't run multiple instances (keystore not shared)
- ❌ Keys lost on restart (availability risk)
- ❌ Requires builder variant maintenance (two builder patterns)

**Use Case**: Dev/test environments, non-production JOSE services

### Option B: Add Database Persistence (Stateful Service)

**Approach**: Add database tables for JWK storage + audit logs, use full ServerBuilder.

**Pros**:
- ✅ Audit trail (regulatory compliance)
- ✅ Multi-instance support (shared database)
- ✅ Key persistence (survive restarts)
- ✅ Uses existing ServerBuilder (no variant needed)
- ✅ Consistent with other services (cipher-im, identity, kms)

**Cons**:
- ❌ Database dependency (PostgreSQL/SQLite required)
- ❌ Migration files required (2001+)
- ❌ More complex deployment (database initialization)
- ❌ Slower startup (database connection + migrations)

**Use Case**: Production JOSE services, compliance-required deployments

## Recommended Decision: Option B (Add Database Persistence)

**Rationale**:

1. **Compliance**: Audit logging is MANDATORY for most production crypto services (SOC2, PCI-DSS, HIPAA)
2. **Scalability**: Multi-instance deployment is essential for high-availability services
3. **Consistency**: All other cryptoutil services use database (cipher-im, identity, kms) - jose-ja should match
4. **Builder Simplicity**: Avoids maintaining two builder variants (stateful vs stateless)
5. **Future-Proofing**: Database enables future features (key rotation, expiration, access control)

**Migration Tables**:

**2001_jose_jwks.up.sql**:
```sql
CREATE TABLE IF NOT EXISTS jwks (
    id TEXT PRIMARY KEY NOT NULL,
    kid TEXT NOT NULL,
    kty TEXT NOT NULL,  -- Key type (RSA, EC, OKP, oct)
    alg TEXT NOT NULL,  -- Algorithm hint
    use TEXT NOT NULL,  -- Key use (sig, enc)
    private_jwk TEXT NOT NULL,  -- Encrypted with barrier
    public_jwk TEXT NOT NULL,   -- Plain text (safe to expose)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(kid)
);

CREATE INDEX IF NOT EXISTS idx_jwks_kid ON jwks(kid);
CREATE INDEX IF NOT EXISTS idx_jwks_use ON jwks(use);
```

**2002_jose_audit_log.up.sql**:
```sql
CREATE TABLE IF NOT EXISTS jwk_audit_log (
    id TEXT PRIMARY KEY NOT NULL,
    operation TEXT NOT NULL,  -- generate, get, delete, sign, verify, encrypt, decrypt
    kid TEXT,
    user_id TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT  -- JSON with additional details
);

CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_timestamp ON jwk_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_kid ON jwk_audit_log(kid);
```

## Next Steps

Based on **Option B (Add Database Persistence)** recommendation:

1. ✅ **Analysis Complete** - This document
2. ⏳ **Create Refactoring Plan** - Detailed phases and tasks
3. ⏳ **Create Refactoring Tasks** - Task breakdown with validation criteria
4. ⏳ **Iterative Review** - Multiple deep analysis passes
5. ⏳ **Create QUIZME** - Validate understanding before implementation
6. ⏳ **Execute Refactoring** - Follow approved plan

## Cross-References

- **ServerBuilder Pattern**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **Merged Migrations Pattern**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md#merged-migrations-pattern)
- **cipher-im Refactoring**: [internal/apps/cipher/im/server/](../../internal/apps/cipher/im/server/) (reference implementation)
- **Template Builder**: [internal/apps/template/service/server/builder/server_builder.go](../../internal/apps/template/service/server/builder/server_builder.go)
