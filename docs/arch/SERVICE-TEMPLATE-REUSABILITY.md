# Service Template Reusability Documentation

**Purpose**: Document reusable patterns from cipher-im template implementation for migration of remaining 8 services (jose-ja, pki-ca, identity-authz, identity-idp, identity-rs, identity-rp, identity-spa, sm-kms).

**Migration Order**: cipher-im (✅ COMPLETE) → jose-ja → pki-ca → identity services → sm-kms (last)

**Reference Implementation**: `internal/cipher/` and `internal/template/` demonstrate all patterns.

---

## Migration Benefits

### Before (Manual Management)

Services manually managed:

- Database containers in TestMain
- Multiple service dependencies injected into server constructors
- Complex setup/cleanup in test files

Example from `internal/cipher/integration/testmain_integration_test.go`:

```go
func TestMain(m *testing.M) {
    // Manual PostgreSQL container setup.
    sharedPGContainer, sharedConnStr, err = container.SetupSharedPostgresContainer(ctx)
    // ...
    sharedDB, err = gorm.Open(postgresDriver.Open(sharedConnStr), &gorm.Config{})
    // ...
    sharedServer, err = server.New(ctx, cfg, db, repository.DatabaseTypePostgreSQL)
    // Manual cleanup with defer.
}
```

### After (Service Template)

Service-template handles all infrastructure:

```go
func TestMain(m *testing.M) {
    cfg := &config.AppConfig{
        ServerSettings: cryptoutilConfig.ServerSettings{
            DatabaseURL:       "", // Empty = use testcontainer.
            DatabaseContainer: "required",
        },
        JWTSecret: uuid.Must(uuid.NewUUID()).String(),
    }

    // Single function call - handles everything.
    app, err := application.StartApplicationListener(ctx, &application.ApplicationListenerConfig{
        Settings:     &cfg.ServerSettings,
        PublicServer: publicServer,
        AdminServer:  adminServer,
    })

    exitCode := m.Run()
    app.Shutdown(context.Background()) // Automatic cleanup.
    os.Exit(exitCode)
}
```

**Reduction**:

- 50+ lines → 15 lines per TestMain
- Zero manual container management
- Zero manual service initialization
- Automatic cleanup (no defer chains)

---

## 1. Realms Service Pattern

**Purpose**: Domain-agnostic user authentication and session management with tenant isolation.

**Location**: `internal/template/server/realms/`

**Key Components**:

### Schema Lifecycle

```go
// Define domain model
type User struct {
    ID           googleUuid.UUID
    Username     string
    PasswordHash string  // {version}$algorithm$iterations$salt$hash
    TenantID     googleUuid.UUID
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Repository interface (database-agnostic)
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByUsername(ctx context.Context, username string) (*User, error)
    Delete(ctx context.Context, id googleUuid.UUID) error
}

// Service layer (business logic)
type Service interface {
    RegisterUser(ctx context.Context, req *RegisterUserRequest) (*User, error)
    AuthenticateUser(ctx context.Context, req *AuthenticateUserRequest) (*Session, error)
}
```

### Tenant Isolation

**MANDATORY**: All queries MUST include `TenantID` filter.

```go
func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Where("username = ? AND tenant_id = ?", username, r.tenantID).
        First(&user).Error
    if err != nil {
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return &user, nil
}
```

### Generic Interfaces

**Pattern**: Service layer NEVER imports infrastructure (GORM, HTTP handlers).

**Dependencies**:

- ✅ `service` → `repository interface` (dependency injection)
- ✅ `service` → `domain models`
- ❌ `service` → `gorm.DB` (FORBIDDEN - breaks abstraction)
- ❌ `service` → `fiber.Ctx` (FORBIDDEN - presentation layer)

**Example**:

```go
// CORRECT: Service depends on interface
func NewService(repo UserRepository, hashService cryptoutilHash.Service) *Service {
    return &Service{repo: repo, hashService: hashService}
}

// WRONG: Service depends on concrete implementation
func NewService(db *gorm.DB, hashService cryptoutilHash.Service) *Service {
    // ❌ Violates dependency inversion
}
```

---

## 2. Barrier Service Pattern

**Purpose**: Hierarchical key encryption (Unseal → Root → Intermediate → Content).

**Location**: `internal/template/server/barrier/`

**Key Components**:

### Hierarchical Key Encryption

```go
// Layer 1: Unseal Key (Docker secrets, never stored)
// Layer 2: Root Key (encrypted at rest with Unseal Key)
// Layer 3: Intermediate Key (encrypted with Root Key)
// Layer 4: Content Key (encrypted with Intermediate Key)

type BarrierService interface {
    EncryptWithContentKey(ctx context.Context, plaintext []byte) ([]byte, error)
    DecryptWithContentKey(ctx context.Context, ciphertext []byte) ([]byte, error)
}
```

### Rotation Handlers

```go
// HTTP handlers for key rotation
const (
    MinRotationReasonLength = 10
    MaxRotationReasonLength = 500
)

func HandleRotateRootKey(service RotationService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req RotateRootKeyRequest
        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
        }
        if len(req.Reason) < MinRotationReasonLength || len(req.Reason) > MaxRotationReasonLength {
            return c.Status(400).JSON(fiber.Map{"error": "rotation reason must be 10-500 characters"})
        }
        // ... rotation logic
    }
}
```

### Sentinel Errors

**Pattern**: Return semantic errors instead of `nil+nil`.

```go
var (
    ErrNoRootKeyFound         = errors.New("no root key found")
    ErrNoIntermediateKeyFound = errors.New("no intermediate key found")
)

func (r *GormBarrierRepository) GetRootKeyLatest(ctx context.Context) (*RootKey, error) {
    var key RootKey
    err := r.db.WithContext(ctx).
        Order("created_at DESC").
        First(&key).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrNoRootKeyFound  // ✅ Semantic error
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get latest root key: %w", err)
    }
    return &key, nil
}
```

---

## 3. Hash Service Pattern

**Purpose**: Versioned password hashing (PBKDF2-HMAC-SHA256 with OWASP parameters).

**Location**: `internal/shared/hash/` (NOT direct injection into template)

**Usage**: Indirectly via Realms Service.

### Version-Based Policy

```go
// Hash format: {version}$algorithm$iterations$salt$dk
// Example: {3}$pbkdf2-sha256$600000$base64(salt)$base64(dk)

type Service interface {
    HashSecret(version int, secret string) (string, error)
    VerifySecret(hashedSecret, plainSecret string) (bool, error)
}
```

### PBKDF2 Parameters

```go
const (
    Version3Algorithm  = "pbkdf2-sha256"
    Version3Iterations = 600000  // OWASP 2025 recommendation
    Version3SaltLength = 32      // 256 bits
    Version3DKLength   = 32      // 256 bits
)
```

### Pepper Integration (Phase 7)

**MANDATORY**: All hashes MUST use pepper from Docker secrets.

```go
func (s *Service) HashSecret(version int, secret string) (string, error) {
    peppered := secret + s.pepper  // ✅ Pepper before hashing
    salt := make([]byte, Version3SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", fmt.Errorf("failed to generate salt: %w", err)
    }
    dk := pbkdf2.Key([]byte(peppered), salt, Version3Iterations, Version3DKLength, sha256.New)
    return fmt.Sprintf("{%d}$%s$%d$%s$%s",
        version, Version3Algorithm, Version3Iterations,
        base64.StdEncoding.EncodeToString(salt),
        base64.StdEncoding.EncodeToString(dk)), nil
}
```

---

## 4. Telemetry Pattern

**Purpose**: OTLP export to otel-collector-contrib, then Grafana OTEL-LGTM.

**Location**: `internal/template/server/telemetry/` (if exists), otherwise `internal/shared/telemetry/`

### OTLP Configuration

```yaml
# configs/cryptoutil-common.yml
observability:
  otlp:
    protocol: grpc
    endpoint: opentelemetry-collector-contrib:4317
    service_name: cryptoutil-cipher-im
    service_version: 1.0.0
    insecure: true  # dev only
```

### Sidecar Pattern

**MANDATORY**: ALWAYS forward through sidecar (NEVER bypass).

```
cryptoutil → otel-collector:4317 → grafana-otel-lgtm:14317
```

### Structured Logging

```go
import "go.uber.org/zap"

logger.Info("User registered",
    zap.String("user_id", userID.String()),
    zap.String("tenant_id", tenantID.String()),
    zap.Duration("duration", elapsed),
)
```

---

## 5. Repository Patterns

**Purpose**: Database abstraction with PostgreSQL and SQLite dual support.

**Location**: `internal/template/server/realms/gorm_*_repository.go`

### Cross-Database Compatibility

```go
// UUID fields: ALWAYS type:text (SQLite has no native UUID)
type User struct {
    ID       googleUuid.UUID `gorm:"type:text;primaryKey"`
    TenantID googleUuid.UUID `gorm:"type:text;index"`
}

// Nullable UUIDs: Use NullableUUID type (NOT *googleUuid.UUID)
type Session struct {
    UserID googleUuid.UUID         `gorm:"type:text"`
    ClientProfileID NullableUUID  `gorm:"type:text;index"`  // ✅ Custom type
}

// JSON Arrays: ALWAYS serializer:json (NOT type:json)
type Client struct {
    AllowedScopes []string `gorm:"serializer:json"`  // ✅ Works on both
}
```

### SQLite Configuration

```go
// MANDATORY for concurrent operations
sqlDB, _ := sql.Open("sqlite", dsn)
sqlDB.Exec("PRAGMA journal_mode=WAL;")  // Concurrent reads + 1 writer
sqlDB.Exec("PRAGMA busy_timeout = 30000;")  // 30s retry on lock

// GORM requires multiple connections for transactions
sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)  // 5
```

### Test-Containers Pattern

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Start PostgreSQL test-container ONCE per package
    container, _ := postgres.RunContainer(ctx,
        postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewV7().String())),
        postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewV7().String())),
    )
    defer container.Terminate(ctx)

    connStr, _ := container.ConnectionString(ctx)
    testDB, _ = gorm.Open(postgres.Open(connStr), &gorm.Config{})

    exitCode := m.Run()
    os.Exit(exitCode)
}
```

**Benefits**:

- ✅ Heavyweight setup runs ONCE (not per test)
- ✅ Automatic cleanup via defer
- ✅ Isolated databases (randomized credentials)
- ✅ No Docker Compose dependencies

---

## 6. Test Patterns

**Purpose**: Consistent testing across all services.

**Location**: `internal/template/server/*_test.go`

### TestMain Pattern

**When to Use**: PostgreSQL containers, HTTP servers, heavyweight dependencies (>100ms startup).

**When NOT to Use**: Simple unit tests, mocks, lightweight helpers.

```go
var (
    testDB     *gorm.DB
    testServer *Server
)

func TestMain(m *testing.M) {
    // Setup ONCE
    testDB = setupDatabase()
    testServer = setupServer(testDB)
    go testServer.Start()
    defer testServer.Shutdown()

    // Run all tests
    exitCode := m.Run()

    // Cleanup via defer
    os.Exit(exitCode)
}
```

### NewTestConfig Pattern

**MANDATORY**: ALWAYS use for ServerSettings creation in tests.

```go
// ✅ CORRECT: Safe bind address
settings := cryptoutilConfig.NewTestConfig(
    cryptoutilMagic.IPv4Loopback,  // "127.0.0.1"
    0,  // Dynamic port allocation
    true,  // DevMode
)

// ❌ WRONG: Triggers Windows Firewall
settings := &cryptoutilConfig.ServerSettings{
    BindPublicAddress: "",  // Defaults to 0.0.0.0
    BindPublicPort: 0,
}
```

### t.Cleanup Pattern

```go
func TestSomething(t *testing.T) {
    db := setupTestDB(t)
    t.Cleanup(func() {
        sqlDB, _ := db.DB()
        sqlDB.Close()
    })

    // Test logic using db
}
```

**Benefits**:

- ✅ Automatic cleanup (even on panic/failure)
- ✅ LIFO order (reverse of registration)
- ✅ Scoped to subtest (`t.Run`)

### Dynamic Port Allocation

```go
listener, err := net.Listen("tcp", fmt.Sprintf("%s:0", cryptoutilMagic.IPv4Loopback))
require.NoError(t, err)

actualPort := listener.Addr().(*net.TCPAddr).Port
resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api", actualPort))
```

**Why**: Avoids port conflicts in parallel tests.

---

## Migration Readiness Checklist

### Completed Phases

✅ **Phase 1: FIPS Compliance** (Commit: f092a2ce)

- Replaced bcrypt with PBKDF2-HMAC-SHA256
- 600,000 iterations (OWASP 2025)
- Versioned format: `{3}$pbkdf2-sha256$600000$salt$dk`

✅ **Phase 2: Windows Firewall Prevention** (Commits: e824a46c, e84eca64, 7137a11c, d28184d0)

- 4-layer defense: Runtime validation, CICD linter, test helper, documentation
- Mandatory 127.0.0.1 binding in tests
- NewTestConfig() enforces safe defaults

✅ **Phase 3: Template Linting** (Commit: d5040fd7)

- 0 golangci-lint violations in `internal/template/`
- mnd: Named constants (MinRotationReasonLength, MaxRotationReasonLength)
- nilnil: Sentinel errors (ErrNoRootKeyFound, ErrNoIntermediateKeyFound)
- unused: Removed 137 lines of dead mock code
- errcheck: Defer with error checking
- noctx: ExecContext for context propagation
- wrapcheck: Error wrapping with fmt.Errorf("%w", err)

### Pending Phases

⏳ **Phase 5: CICD Non-FIPS Algorithm Linter**

- Detect: bcrypt, scrypt, Argon2, MD5, SHA-1, DES, 3DES, RC4
- Integrate into pre-commit hooks
- Reject banned algorithms at commit time

⏳ **Phase 6: Windows Firewall Root Cause Prevention**

- Research additional trigger patterns (multicast, IPv6, broadcast)
- Deep diagnostic analysis (scan test executables)
- Comprehensive prevention strategy

⏳ **Phase 7: Pepper Implementation**

- Docker secrets for pepper storage (MANDATORY OWASP)
- 32-byte secure random pepper
- Hash format includes pepper: `PBKDF2(password || pepper, salt, ...)`

---

## Service Migration Order

| # | Service | Status | Notes |
|---|---------|--------|-------|
| 1 | cipher-im | ✅ COMPLETE | Blueprint for remaining services |
| 2 | jose-ja | ⏳ PENDING | JOSE/JWK Authority |
| 3 | pki-ca | ⏳ PENDING | Certificate Authority |
| 4 | identity-authz | ⏳ PENDING | OAuth 2.1 Authorization Server |
| 5 | identity-idp | ⏳ PENDING | OIDC Identity Provider |
| 6 | identity-rs | ⏳ PENDING | Resource Server |
| 7 | identity-rp | ⏳ PENDING | Relying Party |
| 8 | identity-spa | ⏳ PENDING | Single Page Application |
| 9 | sm-kms | ⏳ PENDING | Secrets Manager KMS (LAST) |

**Rationale**: sm-kms is most mature service, used as template extraction source. Migrate LAST to validate template handles all edge cases.

---

## Key Takeaways

1. **Realms Service**: Generic user authentication with tenant isolation, NEVER imports infrastructure
2. **Barrier Service**: Hierarchical key encryption with sentinel errors and rotation handlers
3. **Hash Service**: Versioned PBKDF2 with MANDATORY pepper from Docker secrets
4. **Telemetry**: OTLP sidecar pattern (NEVER bypass otel-collector)
5. **Repository**: Cross-DB compatible (type:text for UUIDs, serializer:json for arrays)
6. **Testing**: TestMain for heavyweight, NewTestConfig for safety, t.Cleanup for automatic cleanup

**Migration Pattern**: Extract → Verify → Apply to next service → Iterate

**Reference**: See `internal/cipher/` and `internal/template/` for complete implementation examples.
