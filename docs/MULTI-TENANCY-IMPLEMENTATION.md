# Multi-Tenancy Implementation - Work Tracking Document

**Status**: In Progress
**Started**: 2026-01-09

## Requirements Summary

### Core Multi-Tenancy Requirements

1. **Tenant-User/Client Association**
   - All browser users MUST be associated with a tenant
   - All non-browser clients MUST be associated with a tenant
   - Tenants must have at least one user OR client (cannot drop to zero)

2. **Tenant Registration**
   - Tenants stored in database
   - Users and clients associated with tenants
   - Registration API allows selecting existing tenant OR creating new tenant

3. **New Tenant Registration**
   - Registering user/client automatically approved
   - Associated with newly created tenant
   - First user/client given administrator role

4. **Existing Tenant Registration**
   - User/client stored in unverified table
   - Must be verified by admin
   - Admin assigns role(s) to move from unverified to verified table
   - Unverified users/clients auto-expire after N hours (default 72)

5. **Realm Configuration Per Tenant**
   - Each tenant has configurable realms
   - Sessions associated with tenant's realm ID
   - Realm types (username/password, etc.) manageable per tenant
   - YAML config OR database management
   - New tenants via API get database realm types only
   - File realm types require operator intervention

### SessionManager Integration

1. SessionManager MUST be reusable across Service-Template
2. Cipher-IM MUST use SessionManager for session tokens
3. Support OPAQUE, JWE JWT, and JWS JWT session tokens
4. Browser users and non-browser clients both get sessions on authentication

## Current State Analysis

### Existing Session Infrastructure

**Location**: `internal/apps/template/service/server/`
- ✅ SessionManager implementation exists
- ✅ Supports OPAQUE, JWE, JWS session types
- ✅ Browser and Service session separation
- ✅ Database schema for sessions
- ⚠️ Realm field exists but is just a string identifier (NOT multi-tenancy)

**Problem with Current Realm Implementation**:
The comment in `0003_add_session_manager_tables.up.sql` says:
```sql
realm TEXT,  -- Realm identifier for multi-tenancy
```

This is INCORRECT - storing a string realm identifier is NOT multi-tenancy. True multi-tenancy requires:
- Tenant table with tenant metadata
- User/Client tables with foreign keys to tenants
- Realm configurations per tenant
- Verification workflows for new users/clients

### Files Requiring Changes

**New Tables Required**:
1. `tenants` - Tenant information
2. `users` - Verified users with tenant_id FK
3. `clients` - Verified clients with tenant_id FK
4. `unverified_users` - Pending user registrations
5. `unverified_clients` - Pending client registrations
6. `tenant_realms` - Realm configurations per tenant
7. `roles` - Role definitions
8. `user_roles` - User-role associations
9. `client_roles` - Client-role associations

**Migration Files**:
- Template: New migration for tenant tables
- Cipher-IM: New migration for tenant tables

**Code Changes**:
- SessionManager: Update to use tenant_id instead of realm string
- APIs: Registration, verification, tenant management
- Middleware: Tenant extraction from requests
- Repository: Tenant CRUD operations

## Risks and Uncertainties

### Identified Risks

1. **Scope Creep**: Full multi-tenancy is a large feature
   - Mitigation: Implement incrementally, ensure tests pass at each step

2. **Schema Changes**: Altering existing session tables
   - Mitigation: New migrations, test with both SQLite and PostgreSQL

3. **Backward Compatibility**: Existing services may break
   - Mitigation: Make multi-tenancy optional with feature flag

4. **Performance**: Multiple tenant queries
   - Mitigation: Proper indexing, connection pooling

### Open Questions

1. **Default Tenant**: Should there be a default "system" tenant?
   - Decision: YES - create during initialization

2. **Tenant Isolation**: How strict should data isolation be?
   - Decision: Schema-level isolation per tenant

3. **Admin Users**: Can admin users manage multiple tenants?
   - Decision: YES - super-admin role can manage all tenants

4. **Realm Types**: What realm types are supported?
   - Decision: username/password (DB), LDAP (file config), OAuth2 (file config)

## Implementation Plan

### Phase 1: Database Schema
- [ ] Create tenant tables migration (template)
- [ ] Create user/client tables with tenant FK
- [ ] Create unverified tables
- [ ] Create realm configuration tables
- [ ] Create role tables
- [ ] Apply migrations to both SQLite and PostgreSQL
- [ ] Verify schema with integration tests

### Phase 2: Repository Layer
- [ ] TenantRepository interface and implementation
- [ ] UserRepository with tenant filtering
- [ ] ClientRepository with tenant filtering
- [ ] UnverifiedUserRepository
- [ ] UnverifiedClientRepository
- [ ] RealmRepository per-tenant
- [ ] Unit tests for all repositories

### Phase 3: Business Logic
- [ ] TenantService for tenant CRUD
- [ ] RegistrationService for user/client registration
- [ ] VerificationService for admin approval
- [ ] RealmService for realm configuration
- [ ] Update SessionManager to use tenant_id
- [ ] Unit tests for all services

### Phase 4: API Layer
- [ ] Tenant management APIs
- [ ] User registration API (new tenant or existing)
- [ ] Client registration API (new tenant or existing)
- [ ] Admin verification APIs
- [ ] Realm management APIs
- [ ] Integration tests for API flows

### Phase 5: Cipher-IM Integration
- [ ] Apply tenant migrations to cipher-im
- [ ] Update cipher-im to use SessionManager
- [ ] Create cipher-im specific tenant logic
- [ ] Update docker compose with tenant examples
- [ ] E2E tests with multi-tenant scenarios

### Phase 6: Testing and Documentation
- [ ] Coverage validation (≥95% target)
- [ ] Mutation testing (≥85% target)
- [ ] E2E tests with Docker Compose
- [ ] API documentation
- [ ] User guide for multi-tenancy setup

## Timeline Tracking

### 2026-01-09: Initial Analysis
- Read existing session manager implementation
- Identified incorrect realm usage (string vs true multi-tenancy)
- Created work tracking document
- Defined implementation phases

