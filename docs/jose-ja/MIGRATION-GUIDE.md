# JOSE-JA Migration Guide

This guide documents the migration from the default tenant pattern to the multi-tenant user registration pattern, covering service-template, cipher-im, and jose-ja services.

## Overview

The cryptoutil project has migrated from a **default tenant** pattern (where services automatically used pre-created tenant/realm IDs) to a **user registration** pattern where:

1. Users explicitly register with `create_tenant=true` to create new tenants.
2. Services no longer auto-create default tenants on startup.
3. Multi-tenant isolation is enforced at the repository layer.
4. Tests use TestMain patterns for per-package tenant setup.

## Breaking Changes

### 1. Removed APIs

- `WithDefaultTenant()` - No longer available in ServerBuilder.
- `EnsureDefaultTenant()` - Removed from codebase.
- Magic constants `DefaultTenantID`, `DefaultRealmID` - Removed from `internal/shared/magic/`.

### 2. New Registration Flow

**Before (Default Tenant)**:
```go
// Service automatically used default tenant
builder := NewServerBuilder(ctx, cfg)
builder.WithDefaultTenant(
    cryptoutilMagic.DefaultTenantID,
    cryptoutilMagic.DefaultRealmID,
)
```

**After (User Registration)**:
```go
// No default tenant - users must register
builder := NewServerBuilder(ctx, cfg)
// WithDefaultTenant() not called

// Client registers with new tenant
POST /service/api/v1/auth/register
{
    "username": "testuser",
    "password": "securepassword",
    "create_tenant": true  // Creates new tenant
}
```

### 3. Test Patterns Changed

**Before**: Tests assumed default tenant existed.

**After**: Tests use TestMain to register users and obtain tenant IDs.

## Migration Steps

### Service-Template Services

If your service extends the service-template:

1. **Remove `WithDefaultTenant()` calls**:
   ```go
   // Before
   builder.WithDefaultTenant(tenantID, realmID)
   
   // After - simply remove the call
   ```

2. **Update tests to use TestMain**:
   ```go
   var (
       testTenantID    googleUuid.UUID
       testRealmID     googleUuid.UUID
       testSessionToken string
   )
   
   func TestMain(m *testing.M) {
       server := startTestServer()
       defer server.Shutdown()
       
       // Register user with new tenant
       resp := registerUser(server, "testuser", "password", true, nil)
       testTenantID = resp.TenantID
       testRealmID = resp.RealmID
       testSessionToken = resp.SessionToken
       
       os.Exit(m.Run())
   }
   ```

3. **Update repository calls to include tenant/realm**:
   ```go
   // Before
   result, err := repo.Find(ctx, id)
   
   // After - include tenant isolation
   result, err := repo.FindByIDAndTenant(ctx, id, tenantID, realmID)
   ```

### Cipher-IM Migration

Cipher-IM was migrated to the new pattern:

1. **Server changes** (`internal/apps/cipher/im/server/server.go`):
   - Removed `builder.WithDefaultTenant()` call.
   - Uses `NewFromConfig()` instead of manual builder setup.

2. **Test changes** (`internal/apps/cipher/im/server/testmain_test.go`):
   - TestMain pattern registers user with `create_tenant=true`.
   - All tests use `testTenantID`, `testRealmID` from registration.

### JOSE-JA Migration

JOSE-JA was refactored with the new pattern:

1. **ElasticJWK Service** - All operations require tenant/realm context.
2. **Repository layer** - Queries filter by tenant_id AND realm_id.
3. **Handler layer** - Extracts session context for tenant identification.

## API Reference

### Registration Endpoint

**POST** `/service/api/v1/auth/register` (headless clients)
**POST** `/browser/api/v1/auth/register` (browser clients)

**Request**:
```json
{
    "username": "string",
    "password": "string",
    "create_tenant": true,      // Creates new tenant if true
    "join_tenant_id": "uuid"    // Join existing tenant if provided
}
```

**Response (create_tenant=true)**:
```json
{
    "user_id": "uuid",
    "tenant_id": "uuid",        // Newly created tenant
    "realm_id": "uuid",         // Default realm in tenant
    "session_token": "string"   // Active session
}
```

**Response (join_tenant_id provided)**:
```json
{
    "user_id": "uuid",
    "join_request_id": "uuid",  // Pending admin approval
    "status": "pending"
}
```

### Join Request Management (Admin)

**GET** `/service/api/v1/admin/join-requests?tenant_id=uuid`
**POST** `/service/api/v1/admin/join-requests/:id/approve`
**POST** `/service/api/v1/admin/join-requests/:id/reject`

## Path Migration

All services now use versioned API paths:

| Old Path | New Path (Service) | New Path (Browser) |
|----------|-------------------|-------------------|
| `/api/jose/*` | `/service/api/v1/jose/*` | `/browser/api/v1/jose/*` |
| `/api/im/*` | `/service/api/v1/im/*` | `/browser/api/v1/im/*` |
| `/admin/*` | `/admin/v1/*` | `/admin/v1/*` |

### Example: JOSE-JA Endpoints

```
# Service API (headless clients)
POST /service/api/v1/jose/jwk/generate
GET  /service/api/v1/jose/jwk/list
POST /service/api/v1/jose/jws/sign
POST /service/api/v1/jose/jws/verify
POST /service/api/v1/jose/jwe/encrypt
POST /service/api/v1/jose/jwe/decrypt

# Browser API (browser clients)
POST /browser/api/v1/jose/jwk/generate
...

# JWKS Endpoint
GET /service/api/v1/jose/elastic-jwks/:kid/.well-known/jwks.json
GET /browser/api/v1/jose/elastic-jwks/:kid/.well-known/jwks.json
```

## Multi-Tenant Isolation

### Repository Pattern

All repositories enforce tenant isolation:

```go
func (r *ElasticJWKRepository) FindByIDAndTenant(
    ctx context.Context,
    id googleUuid.UUID,
    tenantID googleUuid.UUID,
    realmID googleUuid.UUID,
) (*ElasticJWK, error) {
    var result ElasticJWK
    err := r.db.WithContext(ctx).
        Where("id = ? AND tenant_id = ? AND realm_id = ?", id, tenantID, realmID).
        First(&result).Error
    return &result, err
}
```

### Testing Pattern

```go
func TestMultiTenantIsolation(t *testing.T) {
    // Create two tenants
    resp1 := registerUser(server, "user1", "pass1", true, nil)
    resp2 := registerUser(server, "user2", "pass2", true, nil)
    
    // Create resource in tenant1
    jwk := createJWK(resp1.TenantID, resp1.RealmID)
    
    // Verify tenant2 cannot access tenant1's resource
    _, err := getJWK(jwk.ID, resp2.TenantID, resp2.RealmID)
    require.ErrorIs(t, err, ErrNotFound)  // Proper isolation
}
```

## Audit Logging

JOSE-JA includes per-tenant audit logging:

```go
// Set audit config for tenant
POST /admin/v1/audit/config
{
    "tenant_id": "uuid",
    "operation": "jws:sign",
    "enabled": true,
    "sampling_rate": 100  // 100% = log all
}

// Get audit logs
GET /admin/v1/audit/logs?tenant_id=uuid&operation=jws:sign
```

## FAQ

### Q: Can I still use a single tenant for simple deployments?

Yes. Register a single user with `create_tenant=true` and use that tenant for all operations. The multi-tenant infrastructure is transparent for single-tenant use cases.

### Q: How do I migrate existing data?

Existing data must be associated with a valid tenant_id and realm_id. Create a migration that:
1. Creates a new tenant via the registration API.
2. Updates existing records with the new tenant_id/realm_id.
3. Creates any necessary user associations.

### Q: What happens to anonymous/unauthenticated requests?

Services that allow unauthenticated requests (e.g., public JWKS endpoint) handle them appropriately. The JWKS endpoint extracts the kid from the URL path and returns public keys regardless of authentication.

## References

- [Service Template Instructions](/.github/instructions/03-08.server-builder.instructions.md)
- [Testing Instructions](/.github/instructions/03-02.testing.instructions.md)
- [Authentication Patterns](/.github/instructions/02-10.authn.instructions.md)
