# JOSE-JA Refactoring Tasks

## Task Organization

Tasks organized by phase with detailed validation criteria. Each task includes effort estimate (S/M/L), dependencies, and completion evidence requirements.

**Effort Scale**:
- **S (Small)**: <4 hours
- **M (Medium)**: 4-8 hours
- **L (Large)**: 1-2 days

## Phase 1: Database Schema & Repository (Foundation)

### P1.1: Create Migration Files

**Effort**: M (6 hours)

**Dependencies**: None

**Files**:
- `internal/jose/repository/migrations/2001_jose_jwks.up.sql` (create)
- `internal/jose/repository/migrations/2001_jose_jwks.down.sql` (create)
- `internal/jose/repository/migrations/2002_jose_audit_log.up.sql` (create)
- `internal/jose/repository/migrations/2002_jose_audit_log.down.sql` (create)
- `internal/jose/repository/migrations.go` (create - embed.FS wrapper)

**Schema Requirements**:

**jwks table**:
```sql
CREATE TABLE IF NOT EXISTS jwks (
    id TEXT PRIMARY KEY NOT NULL,
    kid TEXT NOT NULL,
    kty TEXT NOT NULL,       -- RSA, EC, OKP, oct
    alg TEXT NOT NULL,       -- Algorithm hint
    use TEXT NOT NULL,       -- sig or enc
    private_jwk TEXT NOT NULL,  -- Encrypted with barrier
    public_jwk TEXT NOT NULL,   -- Plain text
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(kid)
);

CREATE INDEX IF NOT EXISTS idx_jwks_kid ON jwks(kid);
CREATE INDEX IF NOT EXISTS idx_jwks_use ON jwks(use);
```

**jwk_audit_log table**:
```sql
CREATE TABLE IF NOT EXISTS jwk_audit_log (
    id TEXT PRIMARY KEY NOT NULL,
    operation TEXT NOT NULL,  -- generate, get, delete, sign, verify, encrypt, decrypt
    kid TEXT,
    user_id TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_timestamp ON jwk_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_jwk_audit_log_kid ON jwk_audit_log(kid);
```

**migrations.go embed pattern**:
```go
package repository

import (
    "embed"
    "io/fs"
    cryptoutilTemplateServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

func GetMergedMigrationsFS() fs.FS {
    return &mergedMigrations{
        template: cryptoutilTemplateServerRepository.MigrationsFS,
        domain:   MigrationsFS,
    }
}
```

**Validation Criteria**:
- ✅ Migration files apply cleanly (PostgreSQL)
- ✅ Migration files apply cleanly (SQLite)
- ✅ Down migrations rollback correctly
- ✅ Indexes created correctly
- ✅ Schema validation: `SELECT * FROM information_schema.tables`
- ✅ No linting errors

**Evidence Required**:
```bash
# PostgreSQL test
go test ./internal/jose/repository -run TestMigrations_PostgreSQL
# SQLite test
go test ./internal/jose/repository -run TestMigrations_SQLite
# Verify schema
psql -h localhost -U postgres -c "\d jwks"
```

---

### P1.2: Create Domain Models

**Effort**: S (3 hours)

**Dependencies**: P1.1

**Files**:
- `internal/jose/domain/jwk.go` (create)
- `internal/jose/domain/audit.go` (create)

**JWK Model Requirements**:
```go
package domain

import (
    "time"
    googleUuid "github.com/google/uuid"
    joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type JWK struct {
    ID         googleUuid.UUID `gorm:"type:text;primaryKey"`
    KID        googleUuid.UUID `gorm:"type:text;uniqueIndex;not null"`
    KTY        string          `gorm:"type:text;not null"`
    ALG        string          `gorm:"type:text;not null"`
    Use        string          `gorm:"type:text;not null"`
    PrivateJWK string          `gorm:"type:text;not null"`  // Encrypted ciphertext
    PublicJWK  string          `gorm:"type:text;not null"`  // JSON
    CreatedAt  time.Time       `gorm:"autoCreateTime"`
}

func (JWK) TableName() string {
    return "jwks"
}

// Implement JWKModel interface
func (j *JWK) GetID() googleUuid.UUID { return j.ID }
func (j *JWK) GetKID() googleUuid.UUID { return j.KID }
func (j *JWK) GetKTY() string { return j.KTY }
func (j *JWK) GetALG() string { return j.ALG }
func (j *JWK) GetUse() string { return j.Use }
func (j *JWK) GetPrivateJWK() joseJwk.Key { /* Deserialize + decrypt */ }
func (j *JWK) GetPublicJWK() joseJwk.Key { /* Deserialize */ }
```

**AuditLogEntry Model Requirements**:
```go
type AuditLogEntry struct {
    ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
    Operation string          `gorm:"type:text;not null"`
    KID       *googleUuid.UUID `gorm:"type:text"`
    UserID    *googleUuid.UUID `gorm:"type:text"`
    Timestamp time.Time       `gorm:"autoCreateTime"`
    Metadata  string          `gorm:"type:text"`
}

func (AuditLogEntry) TableName() string {
    return "jwk_audit_log"
}
```

**Validation Criteria**:
- ✅ Models compile without errors
- ✅ TableName() returns correct table names
- ✅ GORM tags correct (`type:text`, `primaryKey`, `uniqueIndex`)
- ✅ Interface methods implemented correctly
- ✅ No linting errors

**Evidence Required**:
```bash
go build ./internal/jose/domain
golangci-lint run ./internal/jose/domain
```

---

### P1.3: Create JWK Repository Interface

**Effort**: S (2 hours)

**Dependencies**: P1.2

**Files**:
- `internal/jose/repository/jwk_repository.go` (create)

**Interface Requirements**:
```go
package repository

import (
    "context"
    googleUuid "github.com/google/uuid"
    cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
)

type JWKRepository interface {
    Create(ctx context.Context, jwk *cryptoutilJoseDomain.JWK) error
    FindByKID(ctx context.Context, kid googleUuid.UUID) (*cryptoutilJoseDomain.JWK, error)
    FindByUse(ctx context.Context, use string) ([]*cryptoutilJoseDomain.JWK, error)
    List(ctx context.Context) ([]*cryptoutilJoseDomain.JWK, error)
    Delete(ctx context.Context, kid googleUuid.UUID) error
}
```

**GORM Implementation Requirements**:
```go
type GORMJWKRepository struct {
    db *gorm.DB
}

func NewGORMJWKRepository(db *gorm.DB) *GORMJWKRepository {
    return &GORMJWKRepository{db: db}
}

func (r *GORMJWKRepository) Create(ctx context.Context, jwk *cryptoutilJoseDomain.JWK) error {
    return getDB(ctx, r.db).WithContext(ctx).Create(jwk).Error
}

// Transaction support (copy from cipher-im)
type txKey struct{}
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
    return context.WithValue(ctx, txKey{}, tx)
}
func getDB(ctx context.Context, baseDB *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
        return tx
    }
    return baseDB
}
```

**Validation Criteria**:
- ✅ Interface defined correctly
- ✅ GORM implementation compiles
- ✅ Transaction support implemented (WithTransaction, getDB)
- ✅ All CRUD operations present
- ✅ No linting errors

**Evidence Required**:
```bash
go build ./internal/jose/repository
golangci-lint run ./internal/jose/repository
```

---

### P1.4: Create JWK Repository Tests

**Effort**: M (6 hours)

**Dependencies**: P1.3

**Files**:
- `internal/jose/repository/jwk_repository_test.go` (create)

**Test Coverage Requirements**:
- ✅ TestCreate_Success (happy path)
- ✅ TestCreate_Duplicate (UNIQUE constraint violation)
- ✅ TestFindByKID_Found
- ✅ TestFindByKID_NotFound
- ✅ TestFindByUse_MultipleResults
- ✅ TestList_Empty
- ✅ TestList_MultipleRows
- ✅ TestDelete_Success
- ✅ TestDelete_NotFound
- ✅ TestTransactions_RollbackOnError
- ✅ TestTransactions_CommitOnSuccess

**Test Pattern (similar to cipher-im)**:
```go
func TestCreate_Success(t *testing.T) {
    t.Parallel()

    db := setupTestDB(t)
    repo := NewGORMJWKRepository(db)

    jwk := &cryptoutilJoseDomain.JWK{
        ID:  googleUuid.New(),
        KID: googleUuid.New(),
        KTY: "RSA",
        ALG: "RS256",
        Use: "sig",
        PrivateJWK: "encrypted_ciphertext",
        PublicJWK: `{"kty":"RSA","n":"..."}`,
    }

    err := repo.Create(context.Background(), jwk)
    require.NoError(t, err)

    // Verify database state
    var count int64
    db.Model(&cryptoutilJoseDomain.JWK{}).Count(&count)
    require.Equal(t, int64(1), count)
}
```

**Validation Criteria**:
- ✅ All tests pass (PostgreSQL)
- ✅ All tests pass (SQLite)
- ✅ Coverage ≥98% (repository is infrastructure)
- ✅ Tests run in parallel (`t.Parallel()`)
- ✅ No test data leakage (use UUIDv7 for unique values)
- ✅ No linting errors

**Evidence Required**:
```bash
go test ./internal/jose/repository -v -cover
go test -coverprofile=coverage.out ./internal/jose/repository
go tool cover -html=coverage.out
```

---

### P1.5: Create Audit Repository

**Effort**: M (4 hours)

**Dependencies**: P1.2

**Files**:
- `internal/jose/repository/audit_repository.go` (create)
- `internal/jose/repository/audit_repository_test.go` (create)

**Interface Requirements**:
```go
type AuditRepository interface {
    Log(ctx context.Context, entry *cryptoutilJoseDomain.AuditLogEntry) error
    FindByKID(ctx context.Context, kid googleUuid.UUID) ([]*cryptoutilJoseDomain.AuditLogEntry, error)
    FindByOperation(ctx context.Context, operation string) ([]*cryptoutilJoseDomain.AuditLogEntry, error)
    List(ctx context.Context, limit int) ([]*cryptoutilJoseDomain.AuditLogEntry, error)
}
```

**GORM Implementation Requirements**:
```go
type GORMAuditRepository struct {
    db *gorm.DB
}

func NewGORMAuditRepository(db *gorm.DB) *GORMAuditRepository {
    return &GORMAuditRepository{db: db}
}

func (r *GORMAuditRepository) Log(ctx context.Context, entry *cryptoutilJoseDomain.AuditLogEntry) error {
    return getDB(ctx, r.db).WithContext(ctx).Create(entry).Error
}
```

**Test Coverage Requirements**:
- ✅ TestLog_Success
- ✅ TestFindByKID_MultipleEntries
- ✅ TestFindByOperation_EmptyResult
- ✅ TestList_Pagination (limit parameter)
- ✅ Coverage ≥98%

**Validation Criteria**:
- ✅ All tests pass (PostgreSQL + SQLite)
- ✅ Coverage ≥98%
- ✅ No linting errors

**Evidence Required**:
```bash
go test ./internal/jose/repository -run TestAudit -v -cover
```

---

## Phase 2: ServerBuilder Integration (Infrastructure)

### P2.1: Create jose-ja Config Extension

**Effort**: S (2 hours)

**Dependencies**: None (Phase 1 complete)

**Files**:
- `internal/jose/server/config/jose_settings.go` (create)

**Config Requirements**:
```go
package config

import (
    cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
)

type JoseSettings struct {
    *cryptoutilConfig.ServiceTemplateServerSettings

    // Jose-specific settings (if any)
    // KeyRotationInterval time.Duration
    // MaxKeysPerUse int
}

func NewJoseSettings() *JoseSettings {
    return &JoseSettings{
        ServiceTemplateServerSettings: &cryptoutilConfig.ServiceTemplateServerSettings{},
    }
}
```

**Validation Criteria**:
- ✅ Config struct compiles
- ✅ Embeds ServiceTemplateServerSettings correctly
- ✅ No linting errors

**Evidence Required**:
```bash
go build ./internal/jose/server/config
golangci-lint run ./internal/jose/server/config
```

---

### P2.2: Refactor server.go with ServerBuilder

**Effort**: L (1-2 days)

**Dependencies**: P2.1, Phase 1 complete

**Files**:
- `internal/jose/server/server.go` (major refactor)
- `internal/jose/server/server_test.go` (update tests)

**Refactoring Pattern** (similar to cipher-im):

**Before** (current - 283 lines):
```go
func NewServer(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings, tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings) (*Server, error) {
    // Manual TLS generation
    tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)

    // Manual telemetry initialization
    telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)

    // Manual JWKGenService initialization
    jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, settings.VerboseMode)

    // Manual keystore
    keyStore := NewKeyStore()

    // Manual Fiber app
    fiberApp := fiber.New(fiber.Config{...})

    return &Server{...}, nil
}
```

**After** (refactored - estimated ~150 lines):
```go
type JoseServer struct {
    app *cryptoutilTemplateServer.Application
    db  *gorm.DB

    // Services
    telemetryService *cryptoutilTelemetry.TelemetryService
    jwkGenService    *cryptoutilJose.JWKGenService
    barrierService   *cryptoutilTemplateBarrier.BarrierService

    // Repositories
    jwkRepo   *repository.GORMJWKRepository
    auditRepo *repository.GORMAuditRepository
}

func NewFromConfig(ctx context.Context, cfg *config.JoseSettings) (*JoseServer, error) {
    builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

    builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

    builder.WithDefaultTenant(
        cryptoutilMagic.JoseDefaultTenantID,
        cryptoutilMagic.JoseDefaultRealmID,
    )

    builder.WithPublicRouteRegistration(func(
        base *cryptoutilTemplateServer.PublicServerBase,
        res *cryptoutilTemplateBuilder.ServiceResources,
    ) error {
        jwkRepo := repository.NewGORMJWKRepository(res.DB)
        auditRepo := repository.NewGORMAuditRepository(res.DB)

        publicServer, err := NewPublicServer(base, res.JWKGenService, res.BarrierService, jwkRepo, auditRepo)
        if err != nil {
            return err
        }

        return publicServer.registerRoutes()
    })

    resources, err := builder.Build()
    if err != nil {
        return nil, err
    }

    jwkRepo := repository.NewGORMJWKRepository(resources.DB)
    auditRepo := repository.NewGORMAuditRepository(resources.DB)

    return &JoseServer{
        app:              resources.Application,
        db:               resources.DB,
        telemetryService: resources.TelemetryService,
        jwkGenService:    resources.JWKGenService,
        barrierService:   resources.BarrierService,
        jwkRepo:          jwkRepo,
        auditRepo:        auditRepo,
    }, nil
}
```

**Accessor Methods Required** (for tests):
```go
func (s *JoseServer) DB() *gorm.DB { return s.db }
func (s *JoseServer) App() *cryptoutilTemplateServer.Application { return s.app }
func (s *JoseServer) JWKGen() *cryptoutilJose.JWKGenService { return s.jwkGenService }
func (s *JoseServer) Telemetry() *cryptoutilTelemetry.TelemetryService { return s.telemetryService }
func (s *JoseServer) PublicPort() int { return s.app.PublicPort() }
func (s *JoseServer) AdminPort() int { return s.app.AdminPort() }
func (s *JoseServer) SetReady(ready bool) { s.app.SetReady(ready) }
```

**Validation Criteria**:
- ✅ Server builds without errors
- ✅ TLS certificates generated correctly
- ✅ Telemetry service initialized
- ✅ JWKGenService initialized
- ✅ Database connection established
- ✅ Migrations applied (template 1001-1004 + jose 2001-2002)
- ✅ All server tests pass
- ✅ No linting errors
- ✅ Accessor methods work in tests

**Evidence Required**:
```bash
go build ./internal/jose/server
go test ./internal/jose/server -v
golangci-lint run ./internal/jose/server

# Verify TLS
curl -k https://127.0.0.1:8080/health

# Verify migrations applied
psql -h localhost -U postgres -c "SELECT * FROM schema_migrations"
```

---

### P2.3: Remove Deprecated New() Method

**Effort**: S (1 hour)

**Dependencies**: P2.2

**Files**:
- `internal/jose/server/server.go` (remove deprecated method)

**Changes**:
```go
// DELETE this method
func New(settings *cryptoutilConfig.ServiceTemplateServerSettings) (*Server, error) {
    // Deprecated: Use NewFromConfig with explicit config instead
}
```

**Validation Criteria**:
- ✅ No references to `New()` in codebase (`grep -r "server.New("`)
- ✅ All callers updated to use `NewFromConfig`
- ✅ All tests pass

**Evidence Required**:
```bash
grep -r "server.New(" internal/ cmd/
# Should return 0 matches
```

---

## Phase 3: Admin Server Elimination (Deduplication)

### P3.1: Verify Template AdminServer

**Effort**: S (2 hours)

**Dependencies**: Phase 2 complete

**Verification Checklist**:
- ✅ Template AdminServer has `/admin/v1/livez` endpoint
- ✅ Template AdminServer has `/admin/v1/readyz` endpoint
- ✅ Template AdminServer has `/admin/v1/shutdown` endpoint
- ✅ Template AdminServer binds to 127.0.0.1:9090
- ✅ Template AdminServer uses TLS
- ✅ Template AdminServer has readiness state management

**Comparison Pattern**:
```bash
# Compare jose admin.go with template admin.go
diff internal/jose/server/admin.go internal/apps/template/service/server/admin.go
```

**Validation Criteria**:
- ✅ Template AdminServer provides ALL features jose-ja needs
- ✅ No missing endpoints
- ✅ No behavioral differences

**Evidence Required**:
```bash
# Test template AdminServer
curl -k https://127.0.0.1:9090/admin/v1/livez
curl -k https://127.0.0.1:9090/admin/v1/readyz
```

---

### P3.2: Update Application to Use Template AdminServer

**Effort**: M (4 hours)

**Dependencies**: P3.1

**Files**:
- `internal/jose/server/application.go` (update to use template AdminServer)

**Changes Required**:
```go
// Before (manual admin server construction)
adminTLSCfg, _ := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)
adminServer, _ := NewAdminHTTPServer(ctx, settings, adminTLSCfg)

// After (use builder-provided admin server)
// Builder already created admin server in resources.Application
// No manual construction needed
```

**Validation Criteria**:
- ✅ Application uses builder-provided admin server
- ✅ Admin endpoints functional
- ✅ Readiness state management works
- ✅ Shutdown triggers correctly
- ✅ All application tests pass

**Evidence Required**:
```bash
go test ./internal/jose/server -run TestApplication
curl -k https://127.0.0.1:9090/admin/v1/livez
curl -k https://127.0.0.1:9090/admin/v1/readyz
```

---

### P3.3: Delete admin.go

**Effort**: S (30 minutes)

**Dependencies**: P3.2

**Files**:
- `internal/jose/server/admin.go` (DELETE - 259 lines)

**Pre-Deletion Checklist**:
- ✅ No references to `NewAdminHTTPServer` in codebase
- ✅ No imports of `internal/jose/server.AdminServer` in codebase
- ✅ All admin tests pass using template AdminServer

**Validation Criteria**:
- ✅ File deleted
- ✅ All tests pass
- ✅ No build errors
- ✅ Admin endpoints still functional

**Evidence Required**:
```bash
grep -r "NewAdminHTTPServer" internal/ cmd/
# Should return 0 matches

git rm internal/jose/server/admin.go
go build ./...
go test ./...
```

---

## Phase 4: KeyStore Migration (Persistence Layer)

### P4.1: Create Database-Backed JWK Service

**Effort**: L (1-2 days)

**Dependencies**: Phase 1, Phase 2 complete

**Files**:
- `internal/jose/server/service/jwk_service.go` (create)
- `internal/jose/server/service/jwk_service_test.go` (create)

**Service Requirements**:
```go
package service

import (
    "context"
    googleUuid "github.com/google/uuid"
    joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
    cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
    cryptoutilJoseRepository "cryptoutil/internal/jose/repository"
    cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
    cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
)

type JWKService struct {
    jwkRepo       cryptoutilJoseRepository.JWKRepository
    auditRepo     cryptoutilJoseRepository.AuditRepository
    barrierService *cryptoutilBarrier.BarrierService
    jwkGenService  *cryptoutilJose.JWKGenService
}

func NewJWKService(
    jwkRepo cryptoutilJoseRepository.JWKRepository,
    auditRepo cryptoutilJoseRepository.AuditRepository,
    barrierService *cryptoutilBarrier.BarrierService,
    jwkGenService *cryptoutilJose.JWKGenService,
) *JWKService {
    return &JWKService{
        jwkRepo:        jwkRepo,
        auditRepo:      auditRepo,
        barrierService: barrierService,
        jwkGenService:  jwkGenService,
    }
}

func (s *JWKService) Generate(ctx context.Context, alg, use string) (*cryptoutilJoseDomain.JWK, error) {
    // 1. Generate JWK using jwkGenService
    // 2. Encrypt private key with barrierService
    // 3. Store in database via jwkRepo
    // 4. Log operation via auditRepo
    // 5. Return domain.JWK
}

func (s *JWKService) Get(ctx context.Context, kid googleUuid.UUID) (*cryptoutilJoseDomain.JWK, error) {
    // 1. Fetch from database via jwkRepo
    // 2. Decrypt private key with barrierService
    // 3. Log operation via auditRepo
    // 4. Return domain.JWK
}

func (s *JWKService) Delete(ctx context.Context, kid googleUuid.UUID) error {
    // 1. Delete from database via jwkRepo
    // 2. Log operation via auditRepo
}

func (s *JWKService) List(ctx context.Context) ([]*cryptoutilJoseDomain.JWK, error) {
    // 1. Fetch all from database via jwkRepo
    // 2. Decrypt private keys with barrierService
    // 3. Return list
}

func (s *JWKService) GetJWKS(ctx context.Context) (joseJwk.Set, error) {
    // 1. Fetch all public keys from database
    // 2. Return JWKS (public keys only)
}
```

**Test Coverage Requirements**:
- ✅ TestGenerate_Success
- ✅ TestGenerate_PrivateKeyEncrypted (verify ciphertext)
- ✅ TestGenerate_AuditLogCreated
- ✅ TestGet_Success (verify decryption)
- ✅ TestGet_AuditLogCreated
- ✅ TestDelete_Success
- ✅ TestDelete_AuditLogCreated
- ✅ TestList_MultipleKeys
- ✅ TestGetJWKS_PublicKeysOnly
- ✅ Coverage ≥95%

**Validation Criteria**:
- ✅ All tests pass
- ✅ Private keys encrypted (verify NOT in plaintext)
- ✅ Audit logs created for ALL operations
- ✅ No linting errors
- ✅ Coverage ≥95%

**Evidence Required**:
```bash
go test ./internal/jose/server/service -v -cover

# Verify encryption
psql -h localhost -U postgres -d jose -c "SELECT kid, private_jwk FROM jwks" | grep -i "BEGIN PRIVATE KEY"
# Should return 0 matches (encrypted, not plaintext PEM)
```

---

### P4.2: Update handlers.go to Use Repository

**Effort**: L (1-2 days)

**Dependencies**: P4.1

**Files**:
- `internal/jose/server/handlers.go` (major refactor)
- `internal/jose/server/handlers_test.go` (update tests)

**Changes Required**:

**Before** (in-memory KeyStore):
```go
func (s *Server) handleJWKGenerate(c *fiber.Ctx) error {
    // ... generate JWK ...
    storedKey := &StoredKey{
        KID:        *kid,
        PrivateJWK: privateJWK,
        PublicJWK:  publicJWK,
        // ...
    }

    if err := s.keyStore.Store(storedKey); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(...)
    }
    // ...
}
```

**After** (database-backed service):
```go
func (s *Server) handleJWKGenerate(c *fiber.Ctx) error {
    // ... parse request ...

    jwk, err := s.jwkService.Generate(c.Context(), req.Algorithm, req.Use)
    if err != nil {
        s.telemetryService.Slogger.Error("Failed to generate JWK", "error", err)
        return c.Status(fiber.StatusInternalServerError).JSON(...)
    }

    // ... return response ...
}
```

**All Handlers to Update**:
- handleJWKGenerate
- handleJWKGet
- handleJWKDelete
- handleJWKList
- handleJWKS (/.well-known/jwks.json)
- handleJWSSign (uses JWK from database)
- handleJWSVerify (uses JWK from database)
- handleJWEEncrypt (uses JWK from database)
- handleJWEDecrypt (uses JWK from database)

**Validation Criteria**:
- ✅ All handlers use JWKService instead of KeyStore
- ✅ All operations persist to database
- ✅ All operations create audit logs
- ✅ Private keys encrypted at rest
- ✅ All handler tests pass
- ✅ E2E flow works (generate → get → sign → verify)
- ✅ No linting errors

**Evidence Required**:
```bash
go test ./internal/jose/server -run TestHandlers -v
go test ./internal/jose/server -run TestE2E_JWK_Lifecycle -v

# Verify database state
psql -h localhost -U postgres -d jose -c "SELECT COUNT(*) FROM jwks"
psql -h localhost -U postgres -d jose -c "SELECT COUNT(*) FROM jwk_audit_log"
```

---

### P4.3: Delete keystore.go

**Effort**: S (30 minutes)

**Dependencies**: P4.2

**Files**:
- `internal/jose/server/keystore.go` (DELETE - 118 lines)
- `internal/jose/server/keystore_test.go` (DELETE)

**Pre-Deletion Checklist**:
- ✅ No references to `NewKeyStore` in codebase
- ✅ No references to `KeyStore` struct in codebase
- ✅ All handlers use JWKService
- ✅ All tests pass

**Validation Criteria**:
- ✅ Files deleted
- ✅ All tests pass
- ✅ No build errors
- ✅ All JOSE operations functional

**Evidence Required**:
```bash
grep -r "KeyStore" internal/jose/
# Should return 0 matches

git rm internal/jose/server/keystore.go
git rm internal/jose/server/keystore_test.go
go build ./...
go test ./...
```

---

## Phase 5: Application Wrapper Refactor (Lifecycle)

### P5.1: Simplify application.go

**Effort**: M (4 hours)

**Dependencies**: Phase 3, Phase 4 complete

**Files**:
- `internal/jose/server/application.go` (simplify)

**Changes Required**:

**Before** (manual construction):
```go
func NewApplication(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*Application, error) {
    // Manual TLS for public server
    publicTLSCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)

    // Manual TLS for admin server
    adminTLSCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(...)

    // Manual server construction
    publicServer, err := NewServer(ctx, settings, publicTLSCfg)
    adminServer, err := NewAdminHTTPServer(ctx, settings, adminTLSCfg)

    return &Application{
        publicServer: publicServer,
        adminServer:  adminServer,
    }, nil
}
```

**After** (builder-provided):
```go
// Application wrapper is already created by builder
// This file may not be needed anymore - builder provides everything
```

**Decision Point**: Keep application.go or delete it?

**Option A**: Delete application.go (recommended)
- Builder creates Application wrapper directly
- No additional wrapper needed
- Consistent with template pattern

**Option B**: Keep as thin wrapper
- Provides jose-ja specific methods
- Maintains backward compatibility
- Additional layer of indirection

**Validation Criteria**:
- ✅ Application lifecycle works (Start, Shutdown)
- ✅ Both servers start correctly
- ✅ Error handling works
- ✅ Accessor methods functional
- ✅ All application tests pass

**Evidence Required**:
```bash
go test ./internal/jose/server -run TestApplication -v
curl -k https://127.0.0.1:8080/health
curl -k https://127.0.0.1:9090/admin/v1/livez
```

---

## Phase 6: Integration Testing (E2E Validation)

### P6.1: Create E2E Test Suite

**Effort**: L (2 days)

**Dependencies**: Phase 5 complete

**Files**:
- `test/e2e/jose/jose_e2e_test.go` (create)

**Test Scenarios**:

**JWK Lifecycle**:
```go
func TestE2E_JWK_Lifecycle(t *testing.T) {
    // 1. Start jose-ja server
    // 2. Generate JWK via POST /jose/v1/jwk/generate
    // 3. Verify JWK persisted to database
    // 4. Retrieve JWK via GET /jose/v1/jwk/{kid}
    // 5. List all JWKs via GET /jose/v1/jwk
    // 6. Delete JWK via DELETE /jose/v1/jwk/{kid}
    // 7. Verify JWK deleted from database
    // 8. Verify all operations in audit log
}
```

**JWS Operations**:
```go
func TestE2E_JWS_SignVerify(t *testing.T) {
    // 1. Generate signing key
    // 2. Sign payload via POST /jose/v1/jws/sign
    // 3. Verify signature via POST /jose/v1/jws/verify
    // 4. Verify audit log entries
}
```

**JWE Operations**:
```go
func TestE2E_JWE_EncryptDecrypt(t *testing.T) {
    // 1. Generate encryption key
    // 2. Encrypt payload via POST /jose/v1/jwe/encrypt
    // 3. Decrypt ciphertext via POST /jose/v1/jwe/decrypt
    // 4. Verify audit log entries
}
```

**JWKS Discovery**:
```go
func TestE2E_JWKS_Discovery(t *testing.T) {
    // 1. Generate multiple keys (signing + encryption)
    // 2. Fetch JWKS via GET /.well-known/jwks.json
    // 3. Verify all public keys present
    // 4. Verify private keys NOT exposed
}
```

**Validation Criteria**:
- ✅ All E2E tests pass (SQLite)
- ✅ All E2E tests pass (PostgreSQL)
- ✅ Private keys encrypted at rest
- ✅ Audit log complete
- ✅ No linting errors

**Evidence Required**:
```bash
go test ./test/e2e/jose -v -tags=e2e
```

---

### P6.2: Multi-Instance Test

**Effort**: M (6 hours)

**Dependencies**: P6.1

**Files**:
- `test/e2e/jose/multi_instance_test.go` (create)

**Test Scenario**:
```go
func TestE2E_MultiInstance_SharedDatabase(t *testing.T) {
    // 1. Start PostgreSQL database
    // 2. Start jose-ja instance 1 on port 8080
    // 3. Start jose-ja instance 2 on port 8081
    // 4. Generate JWK via instance 1
    // 5. Retrieve JWK via instance 2 (verify shared database)
    // 6. Sign payload via instance 1
    // 7. Verify signature via instance 2 (verify key sharing)
    // 8. Shutdown both instances
}
```

**Validation Criteria**:
- ✅ Both instances share database
- ✅ JWK generated on instance 1 visible on instance 2
- ✅ Signature created on instance 1 verifiable on instance 2
- ✅ No database locking conflicts
- ✅ Audit log contains entries from both instances

**Evidence Required**:
```bash
go test ./test/e2e/jose -run TestE2E_MultiInstance -v -tags=e2e
```

---

### P6.3: Barrier Encryption Verification

**Effort**: M (4 hours)

**Dependencies**: P6.1

**Files**:
- `test/e2e/jose/barrier_test.go` (create)

**Test Scenario**:
```go
func TestE2E_Barrier_PrivateKeysEncrypted(t *testing.T) {
    // 1. Start jose-ja server
    // 2. Generate JWK
    // 3. Query database directly
    // 4. Verify private_jwk column contains ciphertext (NOT plaintext PEM)
    // 5. Verify ciphertext starts with barrier prefix
    // 6. Retrieve JWK via API
    // 7. Verify private key decrypts correctly
}
```

**Validation Criteria**:
- ✅ Private keys NOT in plaintext in database
- ✅ Private keys decrypt correctly via API
- ✅ Barrier ciphertext format correct
- ✅ No plaintext leaks in logs

**Evidence Required**:
```bash
go test ./test/e2e/jose -run TestE2E_Barrier -v -tags=e2e

# Manual verification
psql -h localhost -U postgres -d jose -c "SELECT kid, private_jwk FROM jwks" | grep "BEGIN PRIVATE KEY"
# Should return 0 matches
```

---

### P6.4: Load Testing

**Effort**: M (6 hours)

**Dependencies**: P6.1

**Files**:
- `test/load/jose/JoseLoadTest.scala` (create - Gatling script)

**Load Test Scenarios**:
1. **JWK Generation**: 100 concurrent users generating keys (1000 req/s for 60s)
2. **JWS Signing**: 500 concurrent users signing payloads (5000 req/s for 60s)
3. **JWE Encryption**: 500 concurrent users encrypting payloads (5000 req/s for 60s)
4. **JWKS Discovery**: 1000 concurrent users fetching JWKS (10000 req/s for 60s)

**Validation Criteria**:
- ✅ P95 latency <100ms (JWK generation)
- ✅ P95 latency <50ms (JWS/JWE operations)
- ✅ P95 latency <10ms (JWKS discovery)
- ✅ Error rate <0.1%
- ✅ No database connection pool exhaustion

**Evidence Required**:
```bash
# Run Gatling load test
cd test/load/jose
mvn gatling:test
# Review HTML report in target/gatling/results/
```

---

## Phase 7: Documentation & Cleanup (Finalization)

### P7.1: Create Migration Guide

**Effort**: M (4 hours)

**Dependencies**: Phase 6 complete

**Files**:
- `docs/jose-ja/MIGRATION-GUIDE.md` (create)

**Guide Content**:
1. **Overview**: Stateless → Stateful migration
2. **Prerequisites**: PostgreSQL/SQLite database setup
3. **Migration Steps**:
   - Database setup (create schema, apply migrations)
   - Configuration changes (database URL, barrier unseal key)
   - Deployment (Docker Compose / Kubernetes)
   - Verification (health checks, audit log)
4. **Rollback Plan**: How to revert to old version
5. **Troubleshooting**: Common issues and solutions

**Validation Criteria**:
- ✅ Guide tested with real jose-ja deployment
- ✅ All steps verified to work
- ✅ No ambiguous instructions
- ✅ Examples provided for PostgreSQL + SQLite

**Evidence Required**:
```bash
# Follow migration guide step-by-step
# Verify successful upgrade
curl -k https://127.0.0.1:8080/health
curl -k https://127.0.0.1:9090/admin/v1/livez
```

---

### P7.2: Update Documentation

**Effort**: M (4 hours)

**Dependencies**: P7.1

**Files**:
- `docs/jose-ja/README.md` (update)
- `docs/jose-ja/API.md` (update)
- `README.md` (update jose-ja section)

**Documentation Updates**:
1. **Architecture Diagram**: Show database persistence
2. **Deployment Guide**: Add database requirement
3. **API Documentation**: Update with audit log examples
4. **Configuration Reference**: Add database URL, barrier settings
5. **Multi-Instance Guide**: Document shared database pattern

**Validation Criteria**:
- ✅ All documentation accurate
- ✅ Examples tested
- ✅ Diagrams updated
- ✅ No broken links

**Evidence Required**:
```bash
# Verify documentation
markdownlint docs/jose-ja/*.md
```

---

### P7.3: Update Deployment Configurations

**Effort**: M (4 hours)

**Dependencies**: P7.2

**Files**:
- `deployments/jose/compose.yml` (update)
- `deployments/jose/kubernetes/*.yaml` (update)

**Compose Changes**:
```yaml
services:
  jose-ja:
    image: cryptoutil/jose-ja:latest
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/jose
      - BARRIER_UNSEAL_KEY=file:///run/secrets/barrier_unseal_key
    secrets:
      - barrier_unseal_key
    depends_on:
      - postgres

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=jose
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
```

**Kubernetes Changes**:
- Add PostgreSQL StatefulSet
- Add PersistentVolumeClaim for database
- Add ConfigMap for jose-ja config
- Add Secret for barrier unseal key

**Validation Criteria**:
- ✅ Docker Compose deployment works
- ✅ Kubernetes deployment works
- ✅ Database initialized correctly
- ✅ jose-ja connects to database

**Evidence Required**:
```bash
docker compose -f deployments/jose/compose.yml up -d
docker compose -f deployments/jose/compose.yml ps
# Verify all services running

kubectl apply -f deployments/jose/kubernetes/
kubectl get pods -n jose
# Verify all pods ready
```

---

### P7.4: Final Cleanup

**Effort**: S (2 hours)

**Dependencies**: P7.3

**Cleanup Tasks**:
1. Delete deprecated files (`server_old.go`, old tests)
2. Remove all `//nolint` comments (verify fixes are correct)
3. Remove all TODO/FIXME comments (create issues if needed)
4. Final linting pass (`golangci-lint run`)
5. Final build pass (`go build ./...`)
6. Final test pass (`go test ./...`)

**Validation Criteria**:
- ✅ No deprecated code
- ✅ No `//nolint` comments
- ✅ No TODO/FIXME comments (except tracked issues)
- ✅ `golangci-lint run` clean
- ✅ `go build ./...` clean
- ✅ `go test ./...` passes (≥95% coverage)

**Evidence Required**:
```bash
grep -r "//nolint" internal/jose/
grep -r "TODO\|FIXME" internal/jose/ | grep -v "github.com"
golangci-lint run ./...
go build ./...
go test ./...
```

---

## Summary Statistics

### Total Tasks: 29

**Phase Breakdown**:
- Phase 1 (Database): 5 tasks (2-3 days)
- Phase 2 (Builder): 3 tasks (2-3 days)
- Phase 3 (Admin): 3 tasks (1-2 days)
- Phase 4 (Persistence): 3 tasks (3-4 days)
- Phase 5 (Wrapper): 1 task (1 day)
- Phase 6 (E2E): 4 tasks (2-3 days)
- Phase 7 (Docs): 4 tasks (2-3 days)

**Effort Breakdown**:
- Small (S): 10 tasks (~20 hours)
- Medium (M): 13 tasks (~60 hours)
- Large (L): 6 tasks (~50 hours)
- **TOTAL**: ~130 hours (16-19 days)

### Code Metrics

**Before Refactoring**:
- Total Lines: ~1603 lines
- Duplication: ~459 lines (29%)
- Database: None (in-memory)
- Audit Logs: None

**After Refactoring**:
- Total Lines: ~900 lines (44% reduction)
- Duplication: 0 lines (eliminated)
- Database: PostgreSQL/SQLite
- Audit Logs: Complete

### Coverage Targets

- Production Code: ≥95%
- Infrastructure/Repository: ≥98%
- Mutation Score: ≥85% production, ≥98% infrastructure
- E2E Coverage: All critical paths tested

## Cross-References

- **Plan Document**: [JOSE-JA-REFACTORING-PLAN.md](JOSE-JA-REFACTORING-PLAN.md)
- **Analysis Document**: [JOSE-JA-ANALYSIS.md](JOSE-JA-ANALYSIS.md)
- **ServerBuilder Pattern**: [03-08.server-builder.instructions.md](../../.github/instructions/03-08.server-builder.instructions.md)
- **cipher-im Reference**: [internal/apps/cipher/im/server/](../../internal/apps/cipher/im/server/)
