# Service-Template Strategy CLARIFY-QUIZME v1

**Created**: 2026-01-18
**Purpose**: Identify unknowns, risks, incompleteness, and conflicts before executing fixes-needed-PLAN and fixes-needed-TASKS
**Scope**: Service-template reusability for all 9 product-services (cipher-im, jose-ja, pki-ca, identity-authz, identity-idp, identity-rs, identity-rp, identity-spa, sm-kms)

**Instructions**: Answer ALL questions. For multiple-choice, select ONE option (A-D) or provide write-in answer (E). For write-ins, be specific and detailed.

---

## SECTION 1: Multi-Tenancy Registration Flow Architecture

### Q1.1: Registration Endpoint Return Semantics

**Context**: Current plan specifies TWO different registration response patterns:

**Pattern A** (fixes-needed-PLAN.md line 300): Returns 201 (Create) vs 202 (Join)
```json
Response 200 (Create):
{
    "user_id": "uuidv7",
    "tenant_id": "uuidv4",
    "status": "pending"
}
```

**Pattern B** (Current implementation in `tenant_registration_service.go`): Returns Tenant object or error
```go
func RegisterUserWithTenant(...) (*cryptoutilTemplateRepository.Tenant, error)
```

**Question**: Which registration response pattern is CORRECT for all 9 services?

**A)** Pattern A (201 Create with session_token, 202 Join with join_request_id) - HTTP status differentiation
**B)** Pattern B (Always return Tenant object, caller determines HTTP status) - Service layer agnostic
**C)** Hybrid: Service returns RegisterResult struct (union of both patterns), handler chooses HTTP status
**D)** Different patterns per service type (browser vs service clients have different needs)
**E)** Write-in: I specified a new Pattern A format that I want you to use. User must receive HTTP 403 for all endpoints requiring authn until approved, or HTTP 401 if they are rejected. User is not actually saved in users table, they are saved in pending_users table, and not moved to users table unless approved.

Answer: E

**Follow-up**: If join request is rejected, should user be notified via:
- Polling `/browser/api/v1/auth/join-requests/:id`?
- Webhook/callback URL provided during registration?
- Email notification?
- All API calls return 403 Forbidden until approved?


---

### Q1.2: Session Token Issuance for Join Requests

**Context**: Current TenantRegistrationService implementation shows:

**Create Tenant Flow** (fixes-needed-PLAN.md line 300):
- Returns: `session_token` (user can immediately use the service)

**Join Existing Tenant Flow** (fixes-needed-PLAN.md line 305):
- Returns: `join_request_id` with status "pending"
- **UNCLEAR**: Can user get a session token before admin approval?

**Question**: When joining EXISTING tenant (not creating new), when does user receive a valid session token?

**A)** NEVER until admin approves join request (user cannot access ANY endpoints until approval)
**B)** Immediately with limited permissions (read-only access, cannot create resources)
**C)** Immediately with full permissions to tenant's data (admin can revoke later)
**D)** Depends on tenant's join policy setting (auto-approve, manual-approve, invite-only)
**E)** Write-in: I removed session_token and join_request_id from response. See answer in Q1.1. For Q1.2, answer is A, as per Q1.1 answer.

Answer: E

**Implications for cipher-im**:
- If A: cipher-im messages cannot be created until admin approves
- If B: Need permission system BEFORE jose-ja (scope creep)
- If C: Security risk (unauthorized data access)
- If D: Need tenant configuration table and join policy logic

---

### Q1.3: Realm Creation During Tenant Registration

**Context**: Copilot instructions say "realms are authn only, NOT data scope filtering" (02-02.service-template.instructions.md).

Current plan shows:
```go
Response 201 (Create):
{
    "realm_id": "uuid"  // <-- Realm is created
}
```

**But UNCLEAR**:
- What realm type is created? (File-based users? DB-based users? OAuth federated?)
- What is the default realm name?
- Can multiple realms be created for same tenant?
- Do ALL 9 services create identical realms, or service-specific?

**Question**: When user creates NEW tenant via registration, what realm configuration is created?

**A)** Single DB-based username/password realm named "default" (all 9 services use this pattern)
**B)** Single file-based username/password realm named "default" (all 9 services use this pattern)
**C)** Service-specific realm type (cipher-im uses DB, jose-ja uses file, identity-* uses OAuth)
**D)** NO realm created during registration (admin must configure realm before users can login)
**E)** Write-in: I removed realm_id from response. See answer in Q1.1. For Q1.3, new user creates new tenant with new DB-based username/password realm".

Answer: E

**Follow-up**: If tenant has 5 realms (3 LDAP, 1 OAuth, 1 DB), how does registration endpoint know which realm to use?
- Always create user in DB realm?
- User provides `realm_id` parameter in registration request?
- Admin pre-configures "primary realm" for tenant?

---

## SECTION 2: ServerBuilder Registration Flow Integration

### Q2.1: Registration Service Availability Before First Request

**Context**: ServerBuilder creates RegistrationService during `Build()`:

```go
func (b *ServerBuilder) Build() (*ServiceResources, error) {
    // ...
    registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(
        core.DB,
        tenantRepo,
        userRepo,
        joinRequestRepo,
    )
    // ...
}
```

**But registration endpoints need to exist BEFORE any tenant exists**:
- `/browser/api/v1/auth/register` (public, no authentication)
- `/service/api/v1/auth/register` (public, no authentication)

**Question**: How are registration endpoints registered in ServerBuilder if they don't require authentication?

**A)** Always registered in PublicServerBase (template infrastructure, not domain-specific)
**B)** Registered in `WithPublicRouteRegistration` callback (domain-specific, each service registers own)
**C)** Registered in separate "unauthenticated routes" callback (new builder method needed)
**D)** Registration endpoints are NOT part of service-template (each service implements own)
**E)** Write-in: Endpoints must be changed to `/browser/api/v1/register` and `/service/api/v1/register`, and be unauthenticated. Enforce strong rate limiting. Registration logic MUST BE template infrastructure.

Answer:

**Implications**:
- If A: Registration logic becomes template infrastructure (coupled to service-template)
- If B: Every service duplicates registration route registration (9× duplication)
- If C: New builder method needed (breaks existing pattern)
- If D: Defeats purpose of service-template reusability

---

### Q2.2: Demo Tenant Creation for cipher-im

**Context**: Current cipher-im implementation has:

```go
// internal/apps/cipher/im/server/public_server.go line 99+
func NewPublicServer(...) (*PublicServer, error) {
    // First user registration - create demo tenant
    tenant, err := s.registrationService.RegisterUserWithTenant(
        ctx,
        dummyUserID,
        "Cipher-IM Demo Tenant",
        true, // createTenant = true
    )
}
```

**This creates "demo tenant" at server startup** (before ANY client registration).

**Conflict**: Plan says "NO default tenant pattern" but cipher-im creates demo tenant at startup.

**Question**: Should cipher-im have a demo tenant for E2E testing and demonstrations?

**A)** YES - Keep demo tenant for cipher-im ONLY (jose-ja and other 8 services do NOT create demo tenant)
**B)** YES - All 9 services create demo tenant for E2E testing (configurable via `--demo-mode` flag)
**C)** NO - Remove demo tenant from cipher-im (E2E tests use registration API to create tenant)
**D)** YES - But move demo tenant creation to E2E test setup (TestMain), NOT server startup
**E)** Write-in:

Answer: C; E2E tests more complex is design intent

**Implications for fixes-needed-PLAN**:
- If A or B: Plan contradicts "NO default tenant" principle
- If C: E2E tests more complex (must register tenant before testing)
- If D: Cleaner separation (server has NO tenants at startup, tests create as needed)

---

## SECTION 3: Service-Template Directory Structure

### Q3.1: Template Code Location Ambiguity

**Context**: Current codebase shows THREE different locations for "service-template" code:

**Location 1**: `internal/apps/template/service/`
- Contains: ServerBuilder, ServiceResources, PublicServerBase, Application
- Used by: cipher-im, jose-ja (planned)

**Location 2**: `docs/service-template/SERVICE-TEMPLATE-v4.md`
- Documents: Template extraction from cipher-im
- Contains: Migration instructions, validation criteria

**Location 3**: Copilot instructions reference "service-template" but path unclear
- `02-02.service-template.instructions.md` says "service-template" but doesn't specify internal/apps/template

**Question**: Is `internal/apps/template/` the CANONICAL location for reusable service infrastructure?

**A)** YES - `internal/apps/template/` is the ONLY location for reusable service infrastructure
**B)** NO - Should be `internal/shared/template/` (shared across all apps, not "app-specific")
**C)** NO - Should be `pkg/template/` (public library for external consumers)
**D)** SPLIT - Infrastructure in `internal/shared/`, app patterns in `internal/apps/template/`
**E)** Write-in:

Answer: A; `internal/apps/template/` the CANONICAL location, and shared+reused by all internal/apps/PRODUCT/SERVICE`

**Implications**:
- If B: Requires moving 50+ files from `internal/apps/template/` to `internal/shared/template/`
- If C: Exposes service-template as public API (versioning commitments, breaking changes impact external users)
- If D: Need clear split criteria (what goes in shared vs apps)

---

### Q3.2: Domain-Specific vs Template Migrations Numbering

**Context**: Current migration numbering:

**Template Migrations** (1001-1004):
- 1001: Sessions tables
- 1002: Barrier encryption keys
- 1003: Realm tables
- 1004: Multi-tenancy (tenants, tenant_realms)

**Plan adds**:
- 1005: tenant_join_requests (template infrastructure)

**Domain Migrations** (2001+):
- cipher-im: 2001 (messages, message_recipients, message_recipient_jwks)
- jose-ja: 2001-2005 (elastic_jwks, material_keys, jwks_config, audit_config, audit_log)

**Question**: Is tenant_join_requests (1005) TEMPLATE infrastructure or DOMAIN logic?

**A)** TEMPLATE infrastructure (1005) - All 9 services use join requests (reusable)
**B)** DOMAIN logic (2001+ for each service) - Only some services need join requests
**C)** OPTIONAL template (10XX range) - Services opt-in via builder configuration
**D)** Split: Join request TABLE is template (1005), but join request LOGIC is domain (each service customizes approval workflow)
**E)** Write-in:

Answer: A; ALL SERVICES NEED MULTI-TENANT JOIN!!!!

**Implications**:
- If A: tenant_join_requests becomes mandatory for all 9 services (even if some don't need multi-tenant join)
- If B: Duplicate migration files across 9 services (defeats template reusability)
- If C: Need versioned migration sets (1001-1004 = base, 1005-1010 = optional features)
- If D: Table reused, approval workflow customized (reasonable compromise)

---

## SECTION 4: Realms as Authentication-Only

### Q4.1: Realm ID Filtering in Repository WHERE Clauses

**Context**: Copilot instructions say:

> "Realms are authn only (removed from repository WHERE clauses)" (fixes-needed-TASKS.md line 9)

**Current codebase shows**:
- Some repositories filter by `tenant_id` only (correct)
- NO repositories currently filter by `realm_id` (already compliant)

**But UNCLEAR**:
- If realms are authn only, why are they in tenant_realms table?
- If user switches realms within same tenant, do they see different data?
- If client uses OAuth realm vs password realm, are they isolated?

**Question**: If realms are ONLY for authentication, what is their purpose beyond login method selection?

**A)** Realms determine HOW user authenticates (password, OAuth, LDAP, WebAuthn) but ALL realms in tenant see SAME data
**B)** Realms provide logical data isolation within tenant (OAuth users see different messages than password users)
**C)** Realms are legacy from multi-realm design, should be REMOVED entirely (tenant_id is sufficient)
**D)** Realms enable SSO federation (SAML, OIDC) while keeping tenant data unified
**E)** Write-in:

Answer: A

**Implications for fixes-needed-PLAN**:
- If A: Plan is correct (realms = authn only, no data filtering)
- If B: Plan is WRONG (need realm_id in all repository WHERE clauses)
- If C: Need Phase 0 task to remove tenant_realms table entirely
- If D: Need federation configuration in realms (SAML IdP URL, OIDC discovery endpoint)

---

### Q4.2: Session Token and Realm Association

**Context**: SessionManagerService has methods like:

```go
IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID) (string, error)
```

**Session token contains `realm_id`** for tracking which realm authenticated the user.

**Question**: If user has sessions in 2 different realms (password + OAuth), do they see SAME data or DIFFERENT data?

**A)** SAME data (realm_id is for audit logging only, not data filtering)
**B)** DIFFERENT data (realm_id acts as data partition key)
**C)** Depends on service implementation (cipher-im ignores realm, jose-ja uses it)
**D)** Realm switching requires new session (user must logout and re-authenticate)
**E)** Write-in:

Answer: A


---

## SECTION 5: Password Hashing Strategy

### Q5.1: Hash Service Injection Point

**Context**: Current plan shows:

**SERVICE-TEMPLATE-v4.md**:
```go
// Inject Hash Service into service-template realms UserServiceImpl
type UserServiceImpl struct {
    hashService *cryptoutilHash.Service
}
```

**But ServerBuilder does NOT create Hash Service**:
- ServerBuilder creates: DB, Telemetry, JWKGen, Barrier, SessionManager
- **MISSING**: Hash Service

**Question**: Where should Hash Service be created and injected?

**A)** ServerBuilder creates Hash Service (adds to ServiceResources) - All services share same hash config
**B)** Each service creates own Hash Service - Allows service-specific hash policies (different PBKDF2 iterations)
**C)** RealmService creates Hash Service per realm - Different realms can have different hash policies
**D)** Hash Service is global singleton - Created once in main(), passed to all services
**E)** Write-in:

Answer: A

**Implications**:
- If A: Need to add Hash Service to ServerBuilder.Build()
- If B: Defeats template reusability (9× duplication)
- If C: Complex (each realm has different pepper, iterations, versioning)
- If D: Global state violates dependency injection pattern

---

### Q5.2: Pepper Storage for Multi-Tenant Deployments

**Context**: Copilot instructions say:

> "Docker secrets > YAML > ENV priority" (fixes-needed-TASKS.md line 4)

**But pepper is SECRET and SHARED across all tenants**:
- Pepper file: `/run/secrets/hash_pepper`
- All tenants use SAME pepper (security requirement for deterministic hashing)

**Question**: Should pepper be TENANT-SPECIFIC or GLOBAL for all tenants?

**A)** GLOBAL pepper (all tenants use same pepper) - Simpler, one Docker secret
**B)** TENANT-SPECIFIC pepper (each tenant has unique pepper) - Better isolation, more complex
**C)** HYBRID: Global pepper for password hashing, tenant-specific pepper for PII hashing
**D)** REALM-SPECIFIC pepper (each realm has unique pepper) - Maximum isolation
**E)** Write-in:

Answer: A

**Security implications**:
- If A: Pepper compromise affects ALL tenants
- If B: Need pepper storage table (peppers encrypted at rest with barrier service)
- If C: Two pepper management systems (complexity)
- If D: Massive complexity (100s of peppers for large deployments)

---

## SECTION 6: Testing Strategy Conflicts

### Q6.1: TestMain Pattern for Registration Flow

**Context**: Copilot instructions mandate:

> "TestMain Pattern: Tests MUST use TestMain with registration for proper multi-tenancy" (03-08.server-builder.instructions.md line 91)

**Example**:
```go
func TestMain(m *testing.M) {
    // Start server
    testServer, _ = server.NewFromConfig(ctx, cfg)
    go testServer.Start()

    // Register test tenant through API
    resp := registerTestUser(testServer.PublicBaseURL())
    testTenantID = resp.TenantID

    // Run tests
    os.Exit(m.Run())
}
```

**Problem**: If ALL tests share SAME tenant (created in TestMain), tests are NOT isolated.

**Question**: Should TestMain create ONE shared tenant for all tests, or should EACH test create own tenant?

**A)** ONE shared tenant in TestMain (all tests use same testTenantID) - Faster, but tests interfere
**B)** EACH test creates own tenant (t.Run() creates unique tenant) - Isolated, but 100× slower
**C)** EACH test PACKAGE creates own tenant (TestMain per package) - Balanced isolation
**D)** Hybrid: Shared tenant for read-only tests, unique tenant for write tests
**E)** Write-in:

Answer: C; users are created per test, not shared across tests, and users are unique and isolated per test via UUIDv7 in usernames, passwords, and tenant IDs

**Implications**:
- If A: Test fixtures pollute each other (user from Test1 visible in Test2)
- If B: Test suite takes 10+ minutes (1000 tests × 100ms tenant creation each)
- If C: Package-level isolation sufficient for most cases
- If D: Need test categorization (which tests are read-only?)

---

### Q6.2: E2E Test Registration vs Production Registration

**Context**: fixes-needed-PLAN shows:

**Production Registration** (fixes-needed-PLAN.md line 300):
```
POST /browser/api/v1/auth/register
{
    "username": "user@example.com",
    "password": "securepassword",
    "tenant_id": null  // Creates new tenant
}
```

**E2E Test Registration** (current cipher-im implementation):
```go
// Create demo tenant at server startup
tenant, err := s.registrationService.RegisterUserWithTenant(
    ctx,
    dummyUserID,
    "Cipher-IM Demo Tenant",
    true, // createTenant = true
)
```

**CONFLICT**: E2E tests bypass HTTP registration endpoint (call service directly).

**Question**: Should E2E tests use HTTP registration endpoint or call service methods directly?

**A)** HTTP registration endpoint (realistic E2E simulation) - Tests full stack
**B)** Direct service method calls (faster, bypasses HTTP) - Unit test approach
**C)** Hybrid: HTTP for first tenant, direct calls for subsequent users - Balanced
**D)** Depends on test type (smoke tests use HTTP, performance tests use direct calls)
**E)** Write-in:

Answer: A; E2E tests must simulate real user interactions via HTTP endpoints to ensure full stack validation.

**Implications**:
- If A: E2E tests require running server (Docker Compose, slower)
- If B: Not true E2E (misses HTTP layer bugs)
- If C: Inconsistent test patterns
- If D: Need clear test categorization

---

## SECTION 7: jose-ja Specific Requirements

### Q7.1: JWK Storage Multi-Tenancy

**Context**: jose-ja stores Elastic JWKs (key rings) with Material Keys (versioned keys).

**Plan shows** (fixes-needed-PLAN.md Phase 2):
```sql
CREATE TABLE IF NOT EXISTS elastic_jwk (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- Multi-tenant isolation
    kid TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    ...
)
```

**Question**: If two tenants create JWKs with SAME kid (key ID), how are they differentiated?

**A)** kid is GLOBALLY unique (enforced at application level) - Cross-tenant collision check
**B)** kid is UNIQUE per tenant (database constraint: UNIQUE(tenant_id, kid)) - Tenant-scoped uniqueness
**C)** kid can duplicate across tenants (NO uniqueness constraint) - kid is descriptive only
**D)** kid MUST include tenant_id prefix (e.g., "T1-mykey-2024") - Enforced naming convention
**E)** Write-in:

Answer: A; kid must be globally unique across all tenants to prevent collisions.

---

### Q7.2: Cross-Tenant JWKS Access

**Context**: Plan mentions:

> "Cross-tenant via API (not DB config)" (fixes-needed-PLAN.md line 12)

**But cipher-im implementation shows**:
```go
// AllowCrossTenant field in JWKSConfig
type JWKSConfig struct {
    AllowCrossTenant bool
}
```

**Question**: Should jose-ja allow cross-tenant access to JWKs for JWT verification?

**A)** YES - Public keys are public (any tenant can verify JWTs signed by any tenant)
**B)** NO - Each tenant's JWKs are isolated (tenant A cannot verify tenant B's JWTs)
**C)** CONFIGURABLE - Admin enables/disables cross-tenant JWKS access per tenant
**D)** NEVER for production, ALWAYS for dev/testing
**E)** Write-in:

Answer: C

---

## SECTION 8: Blocker Identification

### Q8.1: Phase 0 Completion Criteria

**Context**: Plan says:

> "CRITICAL: Phases 0-1 MUST complete before Phase 2 begins. Service-template and cipher-im are blocking issues for ALL future services." (fixes-needed-PLAN.md line 53)

**But current codebase shows**:
- ✅ ServerBuilder exists
- ✅ TenantRegistrationService exists
- ✅ Registration handlers exist
- ✅ Join request tables/repos exist
- ❌ WithDefaultTenant() still referenced in copilot instructions (03-08.server-builder.instructions.md)

**Question**: What is the ACTUAL blocker preventing jose-ja implementation RIGHT NOW?

**A)** Nothing - jose-ja can start immediately (Phase 0 already done)
**B)** Documentation - Copilot instructions still reference WithDefaultTenant (confusing)
**C)** Testing - No E2E tests validate registration flow works correctly
**D)** Hash Service - Not integrated into ServerBuilder (passwords cannot be hashed)
**E)** Write-in:

Answer: B; fix copilot instructions to remove all references to default tenants

**If A**: Plan's Phase 0 is redundant
**If B**: Simple documentation fix
**If C**: Need comprehensive test suite
**If D**: Need ServerBuilder changes

---

### Q8.2: cipher-im Alignment Validation

**Context**: Plan requires:

> "cipher-im and jose-ja are in perfect alignment with my strategy/architecture/design goals for service-template" (user request)

**Current cipher-im issues**:
1. Demo tenant created at startup (violates "NO default tenant")
2. Direct service calls in E2E tests (bypasses HTTP registration)
3. No Hash Service integration (uses RealmService password hashing)
4. No join request E2E tests (only tenant creation tested)

**Question**: Which cipher-im issues MUST be fixed before jose-ja implementation?

**A)** ALL 4 issues (perfect alignment required)
**B)** Issues 1+2 only (demo tenant and E2E patterns)
**C)** Issue 3 only (Hash Service integration is blocker)
**D)** NONE (cipher-im is reference implementation, imperfections acceptable)
**E)** Write-in:

Answer: A

**Implications**:
- If A: 3-5 days additional cipher-im work before jose-ja
- If B: 1-2 days cipher-im cleanup
- If C: 4-6 hours ServerBuilder changes
- If D: jose-ja starts immediately, inherits cipher-im patterns

---

## SECTION 9: Risk Assessment

### Q9.1: Migration Rollback Strategy

**Context**: Plan involves migrating 9 services sequentially:

> "cipher-im FIRST (validate template) → jose-ja → pki-ca → identity services → sm-kms LAST" (02-02.service-template.instructions.md)

**Question**: If jose-ja migration reveals fundamental template design flaw, what is rollback strategy?

**A)** Revert jose-ja changes, fix template, re-migrate jose-ja
**B)** Keep jose-ja with workaround, fix template for pki-ca
**C)** Pause all migrations, redesign template, re-migrate cipher-im AND jose-ja
**D)** No rollback - Template design is final (cipher-im validation sufficient)
**E)** Write-in:

Answer: D

---

### Q9.2: Performance Impact of Registration Flow

**Context**: New registration flow requires:
1. Tenant creation (INSERT into tenants)
2. Realm creation (INSERT into tenant_realms)
3. User creation (INSERT into users)
4. Join request creation (INSERT into tenant_join_requests, if joining)
5. Session token issuance (INSERT into sessions)

**Question**: What is acceptable latency for registration endpoint?

**A)** <100ms (same as login) - High performance required
**B)** <500ms (user can wait) - Reasonable for one-time operation
**C)** <2000ms (with progress indicator) - Acceptable for complex setup
**D)** <5000ms (async job) - Background processing, email confirmation
**E)** Write-in:

Answer: B

**Implications for database**:
- If A: Need caching, optimistic locking, connection pooling
- If B: Standard transactional approach sufficient
- If C: Can afford sequential operations (no parallelization needed)
- If D: Need job queue (Redis, PostgreSQL LISTEN/NOTIFY)

---

## SECTION 10: Completion Criteria

### Q10.1: Definition of "Perfect Alignment"

**Context**: User requirement:

> "cipher-im and jose-ja in perfect alignment with my strategy/architecture/design goals for service-template to contain and be used for maximum reuse in all 9 product-services"

**Question**: What constitutes "perfect alignment" for cipher-im and jose-ja?

**A)** Zero code duplication (all infrastructure in service-template)
**B)** Consistent API paths (/service/api/v1/*, /browser/api/v1/*, /admin/api/v1/*)
**C)** Identical registration flow (both use tenant join requests)
**D)** All copilot instructions followed (no violations in either service)
**E)** Write-in: A, B, C

Answer: E

**Validation checklist** (which apply?):
- [ ] Both services use ServerBuilder
- [ ] Both services use TenantRegistrationService
- [ ] Both services have NO default tenant
- [ ] Both services filter by tenant_id ONLY (not realm_id)
- [ ] Both services use Hash Service for passwords
- [ ] Both services use cryptoutilMagic.TestPassword in tests
- [ ] Both services pass all quality gates (≥95% coverage, ≥85% mutation)
- [ ] Both services have identical directory structure
- [ ] Both services have E2E tests for registration flow
- [ ] Both services documented in service-template reusability guide

---

### Q10.2: Success Criteria for QUIZME Resolution

**Context**: This QUIZME document identifies 40+ unknowns across 10 sections.

**Question**: What percentage of questions must be answered to proceed with fixes-needed-PLAN execution?

**A)** 100% (all questions answered) - No ambiguity allowed
**B)** 80% (critical questions answered) - Some design decisions can be deferred
**C)** 50% (blocking questions answered) - Iterative refinement acceptable
**D)** 0% (execute plan, refine on discovery) - Agile approach
**E)** Write-in:

Answer: A

**Critical questions** (MUST answer before proceeding):
- Q1.1: Registration response pattern
- Q1.2: Session token issuance timing
- Q1.3: Realm creation during registration
- Q2.2: Demo tenant for cipher-im
- Q3.1: Service-template directory location
- Q5.1: Hash Service injection point
- Q8.1: Phase 0 completion criteria
- Q8.2: cipher-im alignment requirements

**Non-critical questions** (can defer):
- Q7.2: Cross-tenant JWKS access (jose-ja specific)
- Q9.1: Rollback strategy (risk mitigation)
- Q9.2: Performance targets (optimization)

---

## SUBMISSION INSTRUCTIONS

**Format**: Create `CLARIFY-ANSWERS-v1.md` with:

```markdown
# CLARIFY-ANSWERS v1

## Q1.1: Registration Endpoint Return Semantics
**Answer**: C
**Rationale**: ...

## Q1.2: Session Token Issuance for Join Requests
**Answer**: A
**Rationale**: ...

[etc for all 40+ questions]
```

**Next Steps**:
1. User answers ALL questions in CLARIFY-ANSWERS-v1.md
2. Agent analyzes answers for conflicts/incompleteness
3. Agent updates fixes-needed-PLAN.md and fixes-needed-TASKS.md
4. Agent executes updated plan with full context

**Timeline**: Expect 2-3 hours for comprehensive answers (thoughtful, not rushed).
