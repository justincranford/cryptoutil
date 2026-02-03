# Service Template Blueprint

**Purpose**: Complete blueprint for building services using `internal/apps/template/`.

**Companion Document**: [ARCHITECTURE.md](ARCHITECTURE.md) - Product suite overview and deployment patterns.

---

## ServerBuilder Pattern

### Overview

`ServerBuilder` eliminates **260+ lines of boilerplate** per service by providing:

- Dual HTTPS servers (public + admin)
- TLS certificate configuration
- Database connection and migrations
- Barrier encryption service
- Session management
- Telemetry integration
- Graceful shutdown

### Basic Usage

```go
func NewFromConfig(ctx context.Context, cfg *config.ServiceSettings) (*Server, error) {
    // 1. Create builder with template configuration
    builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

    // 2. Register domain migrations (2001+)
    builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

    // 3. Register domain-specific routes
    builder.WithPublicRouteRegistration(func(
        base *cryptoutilTemplateServer.PublicServerBase,
        res *cryptoutilTemplateBuilder.ServiceResources,
    ) error {
        // Create domain repositories
        msgRepo := repository.NewMessageRepository(res.DB)
        userRepo := repository.NewUserRepository(res.DB)

        // Create domain public server
        publicServer := NewPublicServer(base, msgRepo, userRepo, res.TelemetryService, res.JWKGenService)

        // Register routes on base.App (Fiber instance)
        publicServer.RegisterRoutes(base.App)

        return nil
    })

    // 4. Build complete infrastructure
    resources, err := builder.Build()
    if err != nil {
        return nil, fmt.Errorf("failed to build server: %w", err)
    }

    return &Server{
        resources:   resources,
        application: resources.Application,
    }, nil
}
```

### ServiceResources

`Build()` returns initialized infrastructure:

```go
type ServiceResources struct {
    DB                  *gorm.DB                              // Configured GORM connection
    TelemetryService    *cryptoutilTelemetry.TelemetryService // OpenTelemetry integration
    JWKGenService       *cryptoutilJose.JWKGenService         // JWK generation pool
    UnsealKeysService   *cryptoutilBarrier.UnsealKeysService  // Unseal key management
    BarrierService      *cryptoutilBarrier.BarrierService     // Encryption-at-rest
    SessionManager      *cryptoutilTemplateBusinessLogic.SessionManagerService
    RegistrationService cryptoutilTemplateService.TenantRegistrationService
    RealmService        cryptoutilTemplateService.RealmService
    RealmRepository     cryptoutilTemplateRepository.TenantRealmRepository
    Application         *cryptoutilTemplateServer.Application // Public + Admin servers
    ShutdownCore        func()                                // Graceful shutdown callback
    ShutdownContainer   func()                                // Container cleanup callback
}
```

### Realm Pattern

**CRITICAL**: Realms define authentication METHOD and POLICY, NOT data scoping.

#### RealmService Interface

The template provides `RealmService` for authentication configuration:

```go
type RealmService interface {
    GetRealmConfig(ctx context.Context, realmID uuid.UUID) (*realms.RealmConfig, error)
    ValidatePassword(ctx context.Context, realmID uuid.UUID, password string) error
    CreateSession(ctx context.Context, realmID uuid.UUID, userID uuid.UUID) (*Session, error)
}
```

#### Realm Types (16 Supported)

**Federated Types** (external identity providers):
- `username_password` - Database credentials (default)
- `ldap` - LDAP/Active Directory
- `oauth2` - OAuth 2.0/OIDC provider
- `saml` - SAML 2.0 federation

**Non-Federated Browser Types** (`/browser/**` paths):
- `jwe-session-cookie` - Encrypted JWT cookie
- `jws-session-cookie` - Signed JWT cookie
- `opaque-session-cookie` - Server-side session
- `basic-username-password` - HTTP Basic auth
- `bearer-api-token` - Bearer token
- `https-client-cert` - mTLS client cert

**Non-Federated Service Types** (`/service/**` paths):
- `jwe-session-token` - Encrypted JWT token
- `jws-session-token` - Signed JWT token
- `opaque-session-token` - Server-side token
- `basic-client-id-secret` - Client credentials
- `bearer-api-token` - Bearer token (shared)
- `https-client-cert` - mTLS (shared)

#### Realm Configuration

Each realm has configurable policies via `RealmConfig`:

| Category | Settings | Defaults |
|----------|----------|----------|
| Password | MinLength, Uppercase, Lowercase, Digits, Special | 12, true, true, true, true |
| Session | Timeout, AbsoluteMax, RefreshEnabled | 3600s, 86400s, true |
| MFA | Required, Methods | false, [] |
| Rate Limit | LoginRateLimit, MessageRateLimit | 5/min, 10/min |

#### Realm vs Tenant

**`tenant_id`** scopes ALL data access (keys, sessions, audit logs).

**`realm_id`** scopes ONLY authentication policies (how users authenticate).

```
Tenant (data isolation)
├── Realm A (username_password) → users authenticate differently
├── Realm B (ldap)              → but all see SAME tenant data
└── Realm C (oauth2)            → 
```

---

## Configuration Pattern

### Settings Embedding

Domain settings MUST embed template settings:

```go
type CipherImServerSettings struct {
    // Embed ALL template settings
    ServiceTemplateServerSettings *cryptoutilTemplateConfig.ServiceTemplateServerSettings `yaml:"service_template"`

    // Add domain-specific settings
    MaxMessageSize int `yaml:"max_message_size"`
}
```

### YAML Configuration

```yaml
service_template:
  bind_public_address: "0.0.0.0"
  bind_public_port: 8888
  bind_admin_address: "127.0.0.1"
  bind_admin_port: 9090

  database:
    type: "sqlite"
    dsn: ":memory:"

  telemetry:
    otlp_endpoint: "opentelemetry-collector:4317"
    service_name: "cipher-im"

  # TLS certificates (auto-generated in dev)
  tls_public_cert_pem: ""
  tls_public_key_pem: ""
  tls_admin_cert_pem: ""
  tls_admin_key_pem: ""

# Domain-specific
max_message_size: 65536
```

### Test Settings

```go
func NewTestSettings() *CipherImServerSettings {
    return &CipherImServerSettings{
        ServiceTemplateServerSettings: cryptoutilTemplateTestutil.NewTestSettings(),
        MaxMessageSize:                65536,
    }
}
```

`NewTestSettings()` configures:
- SQLite in-memory
- Port 0 (dynamic allocation)
- Auto-generated TLS certificates
- Disabled telemetry export

---

## Migration Pattern

### Merged Migrations

Template (1001-1999) + Domain (2001+) migrations are merged at runtime:

```go
// Domain migrations (2001+)
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Register with builder
builder.WithDomainMigrations(MigrationsFS, "migrations")
```

**How it works**: `mergedMigrations` implements `fs.FS`, trying domain FS first, falling back to template FS. golang-migrate sees unified stream.

### Migration Numbering

| Range | Owner | Purpose |
|-------|-------|---------|
| 1001 | Template | Sessions table |
| 1002 | Template | Barrier encryption keys |
| 1003 | Template | Realms table |
| 1004 | Template | Tenants table |
| 1005 | Template | Pending users table |
| 2001+ | Domain | Application-specific tables |

### Example Domain Migration

```sql
-- 2001_create_messages.up.sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    encrypted_content BLOB NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

CREATE INDEX idx_messages_tenant ON messages(tenant_id);
CREATE INDEX idx_messages_sender ON messages(sender_id);
```

```sql
-- 2001_create_messages.down.sql
DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_messages_tenant;
DROP TABLE IF EXISTS messages;
```

---

## Route Registration Pattern

### PublicServerBase

Domain code receives `PublicServerBase` with pre-configured middleware:

```go
type PublicServerBase struct {
    App              *fiber.App                // Register routes here
    TelemetryService *cryptoutilTelemetry.TelemetryService
    JWKGenService    *cryptoutilJose.JWKGenService
}
```

### Route Registration

```go
func (s *PublicServer) RegisterRoutes(app *fiber.App) {
    // Browser paths (session cookies, CSRF)
    browser := app.Group("/browser/api/v1")
    browser.Get("/messages", s.ListMessages)
    browser.Post("/messages", s.SendMessage)

    // Service paths (Bearer tokens)
    service := app.Group("/service/api/v1")
    service.Get("/messages", s.ListMessages)
    service.Post("/messages", s.SendMessage)
}
```

### Handler Pattern

```go
func (s *PublicServer) SendMessage(c *fiber.Ctx) error {
    ctx := c.UserContext()

    // Get tenant from session (injected by middleware)
    tenantID, err := cryptoutilTemplateMiddleware.GetTenantID(ctx)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "unauthorized",
        })
    }

    // Parse request
    var req SendMessageRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request",
        })
    }

    // Business logic (scoped by tenant_id)
    msg, err := s.messageRepo.Create(ctx, tenantID, req)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to send message",
        })
    }

    return c.Status(fiber.StatusCreated).JSON(msg)
}
```

---

## Repository Pattern

### Domain Repository

```go
type MessageRepository struct {
    db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
    return &MessageRepository{db: db}
}

// CRITICAL: Always filter by tenant_id
func (r *MessageRepository) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]Message, error) {
    var messages []Message
    if err := r.db.WithContext(ctx).
        Where("tenant_id = ?", tenantID).
        Order("created_at DESC").
        Find(&messages).Error; err != nil {
        return nil, fmt.Errorf("failed to list messages: %w", err)
    }
    return messages, nil
}

func (r *MessageRepository) Create(ctx context.Context, msg *Message) error {
    if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
        return fmt.Errorf("failed to create message: %w", err)
    }
    return nil
}
```

### Transaction Context Pattern

For cross-repository transactions:

```go
func (s *MessageService) SendWithRecipients(ctx context.Context, msg *Message, recipients []Recipient) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Create message
        if err := tx.Create(msg).Error; err != nil {
            return err
        }

        // Create recipients
        for _, r := range recipients {
            r.MessageID = msg.ID
            if err := tx.Create(&r).Error; err != nil {
                return err
            }
        }

        return nil
    })
}
```

---

## Testing Patterns (MANDATORY)

### TestMain Pattern

**ALL integration tests MUST use TestMain**:

```go
var (
    testDB     *gorm.DB
    testServer *Server
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Create server with test configuration
    cfg := config.NewTestSettings()
    var err error
    testServer, err = NewFromConfig(ctx, cfg)
    if err != nil {
        log.Fatalf("Failed to create test server: %v", err)
    }

    // Start server
    go func() {
        if err := testServer.Start(); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    // Wait for ready
    if err := testServer.WaitForReady(ctx, 10*time.Second); err != nil {
        log.Fatalf("Server not ready: %v", err)
    }

    // Run tests
    exitCode := m.Run()

    // Cleanup
    testServer.Shutdown(ctx)
    os.Exit(exitCode)
}
```

### Table-Driven Tests (MANDATORY)

```go
func TestSendMessage_Validation(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        request SendMessageRequest
        wantErr string
    }{
        {
            name:    "empty content",
            request: SendMessageRequest{Content: ""},
            wantErr: "content required",
        },
        {
            name:    "no recipients",
            request: SendMessageRequest{Content: "hello", Recipients: nil},
            wantErr: "at least one recipient",
        },
        {
            name: "valid request",
            request: SendMessageRequest{
                Content:    "hello",
                Recipients: []string{googleUuid.NewV7().String()},
            },
            wantErr: "",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            // Use unique test data
            tenantID := googleUuid.NewV7()

            err := testServer.SendMessage(ctx, tenantID, tc.request)

            if tc.wantErr != "" {
                require.Error(t, err)
                require.Contains(t, err.Error(), tc.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Handler Tests with app.Test()

**ALWAYS use Fiber's in-memory testing**:

```go
func TestListMessages_Handler(t *testing.T) {
    t.Parallel()

    // Create standalone Fiber app
    app := fiber.New(fiber.Config{DisableStartupMessage: true})

    // Register handler under test
    msgRepo := repository.NewMessageRepository(testDB)
    handler := NewPublicServer(nil, msgRepo, nil, nil, nil)
    app.Get("/browser/api/v1/messages", handler.ListMessages)

    // Create HTTP request (no network call)
    req := httptest.NewRequest("GET", "/browser/api/v1/messages", nil)
    req.Header.Set("X-Tenant-ID", testTenantID.String())

    // Test handler in-memory
    resp, err := app.Test(req, -1)
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 200, resp.StatusCode)
}
```

### Dynamic Test Data (UUIDv7)

```go
func TestCreate_UniqueConstraint(t *testing.T) {
    t.Parallel()

    // Generate unique IDs per test
    tenantID := googleUuid.NewV7()
    userID := googleUuid.NewV7()
    msgID := googleUuid.NewV7()

    msg := &Message{
        ID:       msgID,
        TenantID: tenantID,
        SenderID: userID,
        Content:  fmt.Sprintf("test_%s", msgID),
    }

    err := testRepo.Create(ctx, msg)
    require.NoError(t, err)
}
```

### SQLite DateTime (CRITICAL)

**ALWAYS use `.UTC()` when comparing with SQLite timestamps**:

```go
// ❌ WRONG: time.Now() without .UTC()
if session.CreatedAt.After(time.Now()) { ... }

// ✅ CORRECT: Always use .UTC()
if session.CreatedAt.After(time.Now().UTC()) { ... }
```

**Pre-commit hook auto-converts** `time.Now()` → `time.Now().UTC()`.

---

## Server Lifecycle

### Start

```go
func (s *Server) Start() error {
    return s.application.Start()
}
```

### Shutdown

```go
func (s *Server) Shutdown(ctx context.Context) error {
    // Template handles graceful shutdown:
    // 1. Stop accepting new requests
    // 2. Drain active requests (30s timeout)
    // 3. Close database connections
    // 4. Release telemetry resources
    return s.application.Shutdown(ctx)
}
```

### WaitForReady

```go
func (s *Server) WaitForReady(ctx context.Context, timeout time.Duration) error {
    return s.application.WaitForReady(ctx, timeout)
}
```

### Accessor Methods (Required for Tests)

```go
func (s *Server) PublicPort() int       { return s.application.PublicPort() }
func (s *Server) AdminPort() int        { return s.application.AdminPort() }
func (s *Server) PublicBaseURL() string { return s.application.PublicBaseURL() }
func (s *Server) AdminBaseURL() string  { return s.application.AdminBaseURL() }
func (s *Server) SetReady(ready bool)   { s.application.SetReady(ready) }
```

---

## E2E Testing

### ComposeManager

Use `internal/apps/template/testing/e2e/compose.go`:

```go
func TestE2E_SendMessage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    ctx := context.Background()

    // Start Docker Compose stack
    manager := e2e.NewComposeManager(t, "../../../deployments/cipher")
    manager.Up(ctx)
    defer manager.Down(ctx)

    // Wait for service healthy
    manager.WaitForHealthy(ctx, "cipher-im", 60*time.Second)

    // Get TLS-enabled HTTP client
    client := manager.HTTPClient()

    // Test API
    resp, err := client.Post(
        manager.ServiceURL("cipher-im") + "/browser/api/v1/messages",
        "application/json",
        strings.NewReader(`{"content":"hello","recipients":["user-id"]}`),
    )
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, 201, resp.StatusCode)
}
```

---

## File Structure Template

```
internal/apps/<product>/<service>/
├── domain/
│   ├── models.go          # Domain entities (Message, Recipient)
│   └── errors.go          # Domain-specific errors
├── repository/
│   ├── migrations/
│   │   ├── 2001_create_messages.up.sql
│   │   └── 2001_create_messages.down.sql
│   ├── embed.go           # //go:embed migrations/*.sql
│   ├── message_repo.go    # MessageRepository
│   └── user_repo.go       # UserRepository (if domain-specific)
├── server/
│   ├── config/
│   │   └── settings.go    # ServiceSettings (embeds template)
│   ├── apis/
│   │   ├── public.go      # PublicServer (route registration)
│   │   └── handlers.go    # HTTP handlers
│   └── server.go          # Server (uses ServerBuilder)
├── client/
│   └── client.go          # API client (optional)
├── e2e/
│   └── e2e_test.go        # E2E tests (Docker Compose)
├── integration/
│   └── integration_test.go # Integration tests (TestMain)
└── testutil/
    └── testutil.go        # Test helpers
```

---

## Checklist for New Service

1. **Create directory structure** per template above
2. **Define domain models** in `domain/models.go`
3. **Create migrations** in `repository/migrations/` (start at 2001)
4. **Embed migrations** in `repository/embed.go`
5. **Implement repositories** in `repository/`
6. **Create settings** embedding `ServiceTemplateServerSettings`
7. **Implement server** using `ServerBuilder` pattern
8. **Register routes** via `WithPublicRouteRegistration`
9. **Add TestMain** in `integration/`
10. **Add E2E tests** in `e2e/`
11. **Create Docker Compose** in `deployments/<product>/`
12. **Create CLI entry point** in `cmd/<product>-<service>/main.go`
13. **Verify quality gates**: coverage ≥95%, mutation ≥85%, lint clean

---

## *FromSettings Factory Pattern (PREFERRED)

Services should use settings-based factories for testability and consistency:

```go
// ✅ PREFERRED: Settings-based factory
type UnsealKeysSettings struct {
    KeyPaths []string `yaml:"key_paths"`
}

func NewUnsealKeysServiceFromSettings(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    if settings == nil {
        return nil, errors.New("settings required")
    }
    return &UnsealKeysService{
        keyPaths: settings.KeyPaths,
    }, nil
}

// Usage in ServerBuilder
builder.WithUnsealKeysService(func(settings *UnsealKeysSettings) (*UnsealKeysService, error) {
    return NewUnsealKeysServiceFromSettings(settings)
})
```

**Benefits**:
- All configuration in one struct
- Easy to test (pass test settings)
- Consistent initialization across codebase
- Self-documenting dependencies

---

## Test Settings Factory

Every service config should have a test settings factory:

```go
// NewTestSettings returns configuration suitable for testing
func NewTestSettings() *CipherImServerSettings {
    return &CipherImServerSettings{
        ServiceTemplateServerSettings: cryptoutilTemplateTestutil.NewTestSettings(),
        MaxMessageSize:                65536,
    }
}
```

**NewTestSettings() configures**:
- SQLite in-memory (`:memory:`)
- Port 0 (dynamic allocation, no conflicts)
- Auto-generated TLS certificates
- Disabled telemetry export
- Short timeouts for fast tests

---

## Configuration Priority

**Load Order** (highest to lowest):

1. **Docker Secrets** (`file:///run/secrets/`) - Passwords, keys, tokens
2. **YAML Configuration** (`--config=`) - Primary settings
3. **CLI Parameters** - Overrides

**CRITICAL: Environment variables NOT supported** (security, auditability).

Example precedence:
```bash
# Docker secret overrides YAML overrides CLI default
cipher-im server \
    --database-url=file:///run/secrets/database_url \  # Docker secret (wins)
    --config=/etc/cipher/im.yml                         # YAML config
```
