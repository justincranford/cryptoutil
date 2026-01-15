# Service Template Extraction Plan

**Goal**: Maximize code reuse between cipher-im and cipher-pubsub by extracting ALL common patterns into service-template.

**Principle**: The ONLY differences between cipher-im and cipher-pubsub should be domain-specific (messages vs topics/subscriptions). Everything else (users, tenants, realms, authentication, sessions, migrations, server initialization, TLS, etc.) MUST be in service-template.

---

## Analysis Summary

### Current State

**cipher-im Structure**:
- ✅ Already uses template: Sessions, authentication, users, realms, barrier, telemetry
- ❌ Duplicates in cipher-im: Server initialization (TLS generation, admin/public server setup), database migrations pattern, default tenant creation, repository adapters pattern, API handler registration pattern
- ✅ Domain-specific (should stay): Message domain models, message repository, message handlers, message-specific migrations

**service-template Structure**:
- ✅ Provides: ApplicationCore, ApplicationBasic, PublicServerBase, AdminServer, Sessions, Barrier, Realms, User repository/models
- ❌ Missing: Complete server initialization pattern, migration runner pattern, default tenant seeding pattern, repository adapter pattern, handler registration helpers

---

## Critical Issues Identified

### 1. **Server Initialization Duplication** (CRITICAL)

**Problem**: `cipher-im/server/server.go::NewFromConfig` contains 260+ lines of boilerplate server initialization that would be DUPLICATED in cipher-pubsub.

**Evidence**:
```go
// cipher-im/server/server.go lines 50-310
func NewFromConfig(ctx context.Context, cfg *config.CipherImServerSettings) (*CipherIMServer, error) {
    // 90 lines: TLS config generation for admin server (static/mixed/auto modes)
    // 10 lines: Admin server creation
    // 20 lines: ApplicationCore startup
    // 15 lines: Database type detection
    // 10 lines: Migration application
    // 10 lines: Default tenant seeding
    // 20 lines: Barrier repository and service creation
    // 15 lines: Domain repositories creation
    // 10 lines: Realm repository and service creation
    // 25 lines: SessionManager creation
    // 90 lines: TLS config generation for public server (static/mixed/auto modes)
    // 15 lines: Public server creation
    // 15 lines: Rotation service creation and route registration
    // 10 lines: Status service creation and route registration
    // 10 lines: Application wrapper creation
}
```

**Impact**: cipher-pubsub would duplicate ALL of this boilerplate, changing ONLY:
- Domain repository types (TopicRepository, SubscriptionRepository instead of MessageRepository)
- Public server handler registration (topics/subscriptions instead of messages)

**Solution**: Extract to `template/service/server/builder.go` with factory pattern.

---

### 2. **TLS Configuration Generation Duplication** (CRITICAL)

**Problem**: 180 lines of TLS generation logic (90 for admin + 90 for public) duplicated.

**Evidence**:
```go
// Duplicated pattern in BOTH admin and public TLS generation:
switch tlsMode {
case cryptoutilConfig.TLSModeStatic:
    tlsCfg = &tlsGenerator.TLSGeneratedSettings{
        StaticCertPEM: cfg.TLSStaticCertPEM,
        StaticKeyPEM:  cfg.TLSStaticKeyPEM,
    }
case cryptoutilConfig.TLSModeMixed:
    tlsCfg, err = tlsGenerator.GenerateServerCertFromCA(...)
case cryptoutilConfig.TLSModeAuto:
    tlsCfg, err = tlsGenerator.GenerateAutoTLSGeneratedSettings(...)
}
```

**Impact**: Every service (cipher-im, cipher-pubsub, future services) duplicates this 180-line switch-case pattern.

**Solution**: Extract to `template/service/config/tls_generator/tls_builder.go::GenerateTLSConfigForServer(mode, dnsNames, ipAddresses, ...)`.

---

### 3. **Migration Pattern Duplication** (HIGH)

**Problem**: Each service implements its own migration runner with duplicate error handling, type detection, and execution logic.

**Evidence**:
```go
// cipher-im/repository/migrations.go
func ApplyCipherIMMigrations(sqlDB *sql.DB, dbType DatabaseType) error {
    // 40 lines: DatabaseType detection logic
    // 20 lines: Migration source setup (embed.FS)
    // 30 lines: golang-migrate driver creation
    // 20 lines: Migration execution with error handling
}
```

**Impact**: cipher-pubsub duplicates entire migration runner pattern.

**Solution**: Extract to `template/service/server/repository/migrations.go` with parameterized migration source.

---

### 4. **Default Tenant/Realm Seeding** (MEDIUM)

**Problem**: `ensureDefaultTenant` is service-specific but pattern is reusable.

**Evidence**:
```go
// cipher-im/server/server.go:377
func ensureDefaultTenant(db *gorm.DB) error {
    // Check if tenant exists
    // Create if missing with hardcoded UUID from magic constants
}
```

**Impact**: cipher-pubsub needs identical pattern with different magic UUIDs.

**Solution**: Extract to `template/service/server/repository/seeding.go::EnsureDefaultTenant(db, tenantID, realmID)`.

---

### 5. **Repository Adapter Pattern** (MEDIUM)

**Problem**: `UserRepositoryAdapter` wraps template UserRepository for cipher-im-specific needs. Pattern is reusable but duplicated.

**Evidence**:
```go
// cipher-im/repository/user_repository_adapter.go
type UserRepositoryAdapter struct {
    userRepo *UserRepository
}

func (a *UserRepositoryAdapter) FindByUsername(ctx context.Context, username string) (cryptoutilTemplateRealms.UserModel, error) {
    user, err := a.userRepo.FindByUsername(ctx, username)
    return user, err
}
```

**Impact**: This adapter pattern converts cipher-im UserRepository to template's UserModel interface. cipher-pubsub would need identical adapter.

**Solution**: Make template UserRepository implement UserModel interface directly, eliminating need for adapter.

---

### 6. **Public Server Route Registration** (LOW)

**Problem**: Each service manually registers routes with middleware. Pattern is consistent but manual.

**Evidence**:
```go
// cipher-im/server/public_server.go:127-145
func (s *PublicServer) registerRoutes() {
    // Session endpoints (no middleware)
    app.Post("/service/api/v1/sessions/issue", sessionHandler.IssueSession)
    
    // User endpoints (authentication, no middleware)
    app.Post("/service/api/v1/users/register", s.authnHandler.HandleRegisterUser())
    
    // Business logic endpoints (session middleware required)
    app.Put("/service/api/v1/messages/tx", serviceSessionMiddleware, s.messageHandler.HandleSendMessage())
}
```

**Impact**: cipher-pubsub duplicates route registration pattern with topics/subscriptions instead of messages.

**Solution**: Provide helper in template for standard routes, services register ONLY domain-specific routes.

---

## Extraction Plan

### Phase 1: Server Initialization Builder (HIGHEST PRIORITY)

**Objective**: Eliminate 260 lines of duplication in server initialization.

**Deliverables**:

1. **`template/service/server/builder/server_builder.go`**:
   - `ServerBuilder` struct with fluent API
   - `NewServerBuilder(ctx, config) *ServerBuilder`
   - `WithDomainMigrations(migrationFS, migrationsPath) *ServerBuilder`
   - `WithDomainRepositories(factoryFunc) *ServerBuilder`
   - `WithPublicServerHandlers(registrationFunc) *ServerBuilder`
   - `WithDefaultTenant(tenantID, realmID) *ServerBuilder`
   - `Build() (*Application, error)` - creates complete application with admin + public servers

2. **Usage Pattern**:
```go
// cipher-im/server/server.go
func NewFromConfig(ctx context.Context, cfg *config.CipherImServerSettings) (*CipherIMServer, error) {
    app, core, err := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings).
        WithDomainMigrations(repository.CipherIMMigrations, "migrations").
        WithDefaultTenant(cryptoutilMagic.CipherIMDefaultTenantID, cryptoutilMagic.CipherIMDefaultRealmID).
        WithPublicServerHandlers(registerCipherIMRoutes).
        Build()
    
    if err != nil {
        return nil, err
    }
    
    // Extract domain-specific repositories from app
    messageRepo := repository.NewMessageRepository(app.DB())
    messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(app.DB(), app.BarrierService())
    
    return &CipherIMServer{
        app:         app,
        messageRepo: messageRepo,
        // ...
    }, nil
}

func registerCipherIMRoutes(publicServer *PublicServerBase, services *DomainServices) error {
    messageHandler := apis.NewMessageHandler(services.MessageRepo, ...)
    publicServer.RegisterBusinessRoutes("/messages/tx", PUT, sessionMiddleware, messageHandler.HandleSendMessage())
    // ...
}
```

**Code Reduction**: 260 lines → 20 lines per service (92% reduction).

---

### Phase 2: TLS Configuration Helpers (HIGH PRIORITY)

**Objective**: Eliminate 180 lines of TLS generation duplication.

**Deliverables**:

1. **`template/service/config/tls_generator/tls_builder.go`**:
   - `GenerateTLSConfigForServer(mode, certPEM, keyPEM, caCertPEM, caKeyPEM, dnsNames, ipAddresses) (*TLSGeneratedSettings, error)`
   - Handles all three modes (static, mixed, auto) in single function

2. **Usage Pattern**:
```go
adminTLSCfg, err := tlsGenerator.GenerateTLSConfigForServer(
    cfg.TLSPrivateMode,
    cfg.TLSStaticCertPEM,
    cfg.TLSStaticKeyPEM,
    cfg.TLSMixedCACertPEM,
    cfg.TLSMixedCAKeyPEM,
    cfg.TLSPrivateDNSNames,
    cfg.TLSPrivateIPAddresses,
)
```

**Code Reduction**: 90 lines → 10 lines per TLS config (89% reduction).

---

### Phase 3: Migration Runner Abstraction (HIGH PRIORITY)

**Objective**: Eliminate migration runner duplication.

**Deliverables**:

1. **`template/service/server/repository/migrations_runner.go`**:
   - `ApplyMigrations(sqlDB, migrationFS, migrationsPath, databaseURL) error`
   - Auto-detects database type from URL
   - Handles golang-migrate driver setup and execution

2. **Usage Pattern**:
```go
//go:embed migrations/*.sql
var CipherIMMigrations embed.FS

// In server initialization:
err = cryptoutilTemplateRepository.ApplyMigrations(sqlDB, repository.CipherIMMigrations, "migrations", cfg.DatabaseURL)
```

**Code Reduction**: 110 lines → 5 lines per service (95% reduction).

---

### Phase 4: Default Tenant Seeding Helper (MEDIUM PRIORITY)

**Objective**: Standardize default tenant/realm creation pattern.

**Deliverables**:

1. **`template/service/server/repository/seeding.go`**:
   - `EnsureDefaultTenant(ctx, db, tenantID, realmID, tenantName) error`
   - Creates tenant + realm if missing
   - Idempotent (safe to call multiple times)

2. **Usage Pattern**:
```go
err = cryptoutilTemplateRepository.EnsureDefaultTenant(
    ctx,
    core.DB,
    cryptoutilMagic.CipherIMDefaultTenantID,
    cryptoutilMagic.CipherIMDefaultRealmID,
    "cipher-im-default",
)
```

**Code Reduction**: 30 lines → 7 lines per service (77% reduction).

---

### Phase 5: Repository Interface Standardization (MEDIUM PRIORITY)

**Objective**: Eliminate need for UserRepositoryAdapter by making template UserRepository implement UserModel interface.

**Deliverables**:

1. **Refactor `template/service/server/repository/user_repository.go`**:
   - Make `User` struct implement `realms.UserModel` interface directly
   - Add interface methods if missing

2. **Remove from cipher-im**:
   - Delete `cipher-im/repository/user_repository.go` (uses template directly)
   - Delete `cipher-im/repository/user_repository_adapter.go` (no longer needed)

**Code Reduction**: 80 lines eliminated per service.

---

### Phase 6: Standard Route Registration Helpers (LOW PRIORITY)

**Objective**: Simplify route registration for standard endpoints (sessions, users).

**Deliverables**:

1. **`template/service/server/routes/standard_routes.go`**:
   - `RegisterSessionRoutes(app, sessionManager)`
   - `RegisterUserAuthRoutes(app, userService, sessionManager)`
   - Auto-registers `/service` and `/browser` variants

2. **Usage Pattern**:
```go
func (s *PublicServer) registerRoutes() {
    // Standard routes (sessions, authentication) - handled by template
    cryptoutilTemplateRoutes.RegisterSessionRoutes(s.base.App(), s.sessionManagerService)
    cryptoutilTemplateRoutes.RegisterUserAuthRoutes(s.base.App(), s.authnHandler, s.sessionManagerService)
    
    // Domain-specific routes only
    app := s.base.App()
    app.Put("/service/api/v1/messages/tx", serviceSessionMiddleware, s.messageHandler.HandleSendMessage())
    app.Get("/service/api/v1/messages/rx", serviceSessionMiddleware, s.messageHandler.HandleReceiveMessages())
    app.Delete("/service/api/v1/messages/:id", serviceSessionMiddleware, s.messageHandler.HandleDeleteMessage())
    // Mirror for /browser paths...
}
```

**Code Reduction**: 20 lines → 8 lines for standard routes.

---

## Refactoring Checklist

### Must Extract to Template (CRITICAL)

- [ ] **Server initialization builder** - `builder/server_builder.go`
- [ ] **TLS config helper** - `config/tls_generator/tls_builder.go`
- [ ] **Migration runner** - `repository/migrations_runner.go`
- [ ] **Default tenant seeding** - `repository/seeding.go`
- [ ] **Repository interface alignment** - Make User implement UserModel

### Must Remain in cipher-im (Domain-Specific)

- [x] **Message domain model** - `domain/message.go`
- [x] **MessageRecipientJWK domain model** - `domain/recipient_message_jwk.go`
- [x] **Message repository** - `repository/message_repository.go`
- [x] **MessageRecipientJWK repository** - `repository/message_recipient_jwk_repository.go`
- [x] **Message handlers** - `server/apis/messages.go`
- [x] **Message migrations** - `repository/migrations/*.sql` (message tables only)

### Can Delete from cipher-im (Now in Template)

After extraction:
- [ ] **UserRepository** - Use template version directly
- [ ] **UserRepositoryAdapter** - No longer needed after Phase 5
- [ ] **ensureDefaultTenant** - Use template helper
- [ ] **ApplyCipherIMMigrations boilerplate** - Use template runner (keep embed.FS + path only)
- [ ] **TLS generation switch-case** - Use template helper
- [ ] **90% of server.go::NewFromConfig** - Use builder pattern

---

## Implementation Priority

### Sprint 1 (Highest Impact - 85% Code Reduction)

1. **Phase 1**: Server initialization builder
2. **Phase 2**: TLS configuration helpers
3. **Phase 3**: Migration runner abstraction

**Expected Outcome**: cipher-im `server.go::NewFromConfig` reduces from 260 lines to ~40 lines.

### Sprint 2 (Medium Impact - Repository Cleanup)

4. **Phase 4**: Default tenant seeding helper
5. **Phase 5**: Repository interface standardization

**Expected Outcome**: Delete 110 lines of adapter/wrapper code per service.

### Sprint 3 (Polish - Route Registration)

6. **Phase 6**: Standard route registration helpers

**Expected Outcome**: Cleaner route registration, less duplication.

---

## Validation Criteria

### Code Reuse Metrics

**Before Extraction**:
- cipher-im service-specific code: ~600 lines
- Duplicated template-eligible code: ~480 lines (80%)
- True domain-specific code: ~120 lines (20%)

**After Extraction**:
- cipher-im service-specific code: ~150 lines
- Template-provided code: ~450 lines (reusable across N services)
- True domain-specific code: ~150 lines (100%)

**Target**: ≥75% reduction in per-service boilerplate (480 → 120 lines).

### cipher-pubsub Creation Test

After extraction, creating cipher-pubsub should require ONLY:

1. **Domain Models** (~40 lines):
   - `domain/topic.go`, `domain/subscription.go`

2. **Domain Repositories** (~80 lines):
   - `repository/topic_repository.go`, `repository/subscription_repository.go`

3. **Domain Handlers** (~150 lines):
   - `server/apis/topics.go`, `server/apis/subscriptions.go`

4. **Domain Migrations** (~30 lines SQL):
   - `repository/migrations/0001_topics.up.sql`
   - `repository/migrations/0002_subscriptions.up.sql`

5. **Server Initialization** (~30 lines):
   - `server/server.go` using builder pattern

**Total**: ~330 lines of domain-specific code.

**Without Extraction**: ~810 lines (330 domain + 480 duplicated boilerplate).

**Validation**: Creating cipher-pubsub requires ≤350 lines of new code (domain-specific only).

---

## Risks and Mitigations

### Risk 1: Over-Abstraction

**Concern**: Builder pattern might become too rigid for future services with different needs.

**Mitigation**: Provide "escape hatches" - allow services to access `ApplicationCore` directly and build custom initialization if needed. Builder is convenience, not requirement.

### Risk 2: Breaking Changes

**Concern**: Refactoring cipher-im might break existing tests/workflows.

**Mitigation**: Implement extraction incrementally with continuous validation:
- Phase 1: Extract builder, run cipher-im tests ✅
- Phase 2: Extract TLS helpers, run cipher-im tests ✅
- Each phase validated before proceeding.

### Risk 3: Template Complexity

**Concern**: Template might become too complex with too many options.

**Mitigation**: Follow "convention over configuration" - provide sensible defaults, require config ONLY for deviations. Most services use standard pattern (80/20 rule).

---

## Success Metrics

1. **Code Reduction**: ≥75% reduction in per-service boilerplate
2. **Time to New Service**: cipher-pubsub creation <2 hours (from blueprint)
3. **Consistency**: 100% of services use same initialization pattern
4. **Maintainability**: Template changes propagate to all services automatically
5. **Test Coverage**: Template builder ≥98% coverage, service integration ≥95%

---

## Next Steps

1. **Approve Plan**: Review and approve extraction strategy
2. **Implement Sprint 1**: Server builder + TLS helpers + migration runner
3. **Validate**: Refactor cipher-im to use new builder, run all tests
4. **Document**: Update service-template README with usage patterns
5. **Blueprint**: Create cipher-pubsub to validate template reusability
6. **Iterate**: Continue with Sprint 2 and 3 based on learnings

---

## Appendix: File Organization

### New Template Files (To Create)

```
internal/apps/template/service/
├── server/
│   ├── builder/
│   │   ├── server_builder.go          # Phase 1: Server initialization builder
│   │   └── server_builder_test.go
│   ├── repository/
│   │   ├── migrations_runner.go       # Phase 3: Migration runner
│   │   ├── migrations_runner_test.go
│   │   ├── seeding.go                 # Phase 4: Default tenant seeding
│   │   └── seeding_test.go
│   ├── routes/
│   │   ├── standard_routes.go         # Phase 6: Route registration helpers
│   │   └── standard_routes_test.go
├── config/
│   └── tls_generator/
│       ├── tls_builder.go             # Phase 2: TLS config helper
│       └── tls_builder_test.go
```

### Modified Template Files

```
internal/apps/template/service/
├── server/
│   ├── repository/
│   │   ├── user_repository.go         # Phase 5: Implement UserModel interface
```

### Deleted cipher-im Files (After Extraction)

```
internal/apps/cipher/im/
├── repository/
│   ├── user_repository.go             # Phase 5: Use template version
│   └── user_repository_adapter.go     # Phase 5: No longer needed
```

### Simplified cipher-im Files

```
internal/apps/cipher/im/
├── server/
│   ├── server.go                      # Phases 1-4: 260 → 40 lines
│   └── public_server.go               # Phase 6: Simplified route registration
├── repository/
│   └── migrations.go                  # Phase 3: 110 → 10 lines
```

---

**End of Analysis**
