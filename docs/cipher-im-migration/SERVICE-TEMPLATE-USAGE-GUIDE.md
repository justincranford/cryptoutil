# Using Service Template Application Layer in Tests

This guide shows how to use the new service-template application layer to simplify test setup in cipher-im (and future services).

## Quick Reference: Old vs New

### OLD: Manual Database Management (50+ lines)

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Manual PostgreSQL container setup (10+ lines).
    sharedPGContainer, sharedConnStr, err = container.SetupSharedPostgresContainer(ctx)
    if err != nil {
        panic(fmt.Sprintf("failed to setup PostgreSQL container: %v", err))
    }
    defer func() {
        if err := sharedPGContainer.Terminate(ctx); err != nil {
            fmt.Fprintf(os.Stderr, "failed to terminate PostgreSQL container: %v\n", err)
        }
    }()

    // Verify connection (5+ lines).
    if err := container.VerifyPostgresConnection(sharedConnStr); err != nil {
        panic(fmt.Sprintf("failed to verify PostgreSQL connection: %v", err))
    }

    // Create database connection (5+ lines).
    sharedDB, err = gorm.Open(postgresDriver.Open(sharedConnStr), &gorm.Config{})
    if err != nil {
        panic(fmt.Sprintf("failed to open database connection: %v", err))
    }

    // Create server instance (10+ lines).
    sharedServer, err = server.New(ctx, cfg, sharedDB, repository.DatabaseTypePostgreSQL)
    if err != nil {
        panic(fmt.Sprintf("failed to create cipher server: %v", err))
    }

    exitCode := m.Run()
    os.Exit(exitCode)
}
```

### NEW: Automatic Infrastructure (15 lines)

```go
func TestMain(m *testing.M) {
    ctx := context.Background()

    // Single config - service-template handles everything.
    cfg := &config.AppConfig{
        ServerSettings: cryptoutilConfig.ServerSettings{
            DatabaseURL:       "",           // Empty = use testcontainer.
            DatabaseContainer: "required",   // Require PostgreSQL testcontainer.
            BindPublicPort:    0,            // Dynamic port allocation.
            BindPrivatePort:   0,            // Dynamic port allocation.
            DevMode:           true,
        },
        JWTSecret: uuid.Must(uuid.NewUUID()).String(),
    }

    // Single function call creates everything (database, telemetry, servers).
    sharedServer, err := server.NewFromConfig(ctx, cfg)
    if err != nil {
        panic(fmt.Sprintf("failed to create server: %v", err))
    }

    exitCode := m.Run()

    // Automatic cleanup (no manual defer chains needed).
    _ = sharedServer.Shutdown(context.Background())

    os.Exit(exitCode)
}
```

**Benefits**:

- **50+ lines → 15 lines** (70% reduction)
- **Zero manual container management**
- **Zero manual service initialization**
- **Automatic cleanup** (no defer chains)

## Integration Tests Example

File: `internal/cipher/integration/testmain_integration_test.go`

```go
package integration

import (
 "context"
 "fmt"
 "os"
 "testing"

 "github.com/google/uuid"

 "cryptoutil/internal/cipher/server"
 "cryptoutil/internal/cipher/server/config"
 cryptoutilConfig "cryptoutil/internal/shared/config"
)

var sharedServer *server.CipherIMServer

func TestMain(m *testing.M) {
 ctx := context.Background()

 // Configure automatic PostgreSQL testcontainer.
 cfg := &config.AppConfig{
  ServerSettings: cryptoutilConfig.ServerSettings{
   DatabaseURL:       "",           // Empty = use testcontainer.
   DatabaseContainer: "required",   // Require PostgreSQL testcontainer.
   BindPublicPort:    0,            // Dynamic port allocation.
   BindPrivatePort:   0,            // Dynamic port allocation.
   DevMode:           true,
  },
  JWTSecret: uuid.Must(uuid.NewUUID()).String(),
 }

 // Create server with automatic infrastructure (PostgreSQL, telemetry, etc.).
 var err error

 sharedServer, err = server.NewFromConfig(ctx, cfg)
 if err != nil {
  panic(fmt.Sprintf("failed to create server: %v", err))
 }

 // Run tests.
 exitCode := m.Run()

 // Cleanup.
 _ = sharedServer.Shutdown(context.Background())

 os.Exit(exitCode)
}
```

## E2E Tests Example

File: `internal/cipher/e2e/testmain_e2e_test.go`

```go
package e2e_test

import (
 "context"
 "crypto/tls"
 "fmt"
 "net/http"
 "os"
 "testing"

 "github.com/google/uuid"

 "cryptoutil/internal/cipher/server"
 "cryptoutil/internal/cipher/server/config"
 cryptoutilConfig "cryptoutil/internal/shared/config"
 cryptoutilMagic "cryptoutil/internal/shared/magic"
)

var (
 sharedHTTPClient   *http.Client
 testCipherIMServer *server.CipherIMServer
 baseURL            string
 adminURL           string
)

func TestMain(m *testing.M) {
 ctx := context.Background()

 // Configure SQLite in-memory for fast E2E tests.
 cfg := &config.AppConfig{
  ServerSettings: cryptoutilConfig.ServerSettings{
   DatabaseURL:       "file::memory:?cache=shared", // SQLite in-memory.
   DatabaseContainer: "disabled",                   // No container for E2E.
   BindPublicPort:    0,                            // Dynamic port allocation.
   BindPrivatePort:   0,                            // Dynamic port allocation.
   DevMode:           true,
  },
  JWTSecret: uuid.Must(uuid.NewUUID()).String(),
 }

 // Create server with automatic infrastructure (SQLite, telemetry, etc.).
 var err error

 testCipherIMServer, err = server.NewFromConfig(ctx, cfg)
 if err != nil {
  panic(fmt.Sprintf("failed to create server: %v", err))
 }

 // Setup HTTP client for tests.
 sharedHTTPClient = &http.Client{
  Transport: &http.Transport{
   TLSClientConfig: &tls.Config{
    InsecureSkipVerify: true, //nolint:gosec // Test environment only.
   },
  },
  Timeout: cryptoutilMagic.CipherDefaultTimeout,
 }

 // Get server URLs.
 baseURL = fmt.Sprintf("https://127.0.0.1:%d", testCipherIMServer.PublicPort())

 adminPort, _ := testCipherIMServer.AdminPort()
 adminURL = fmt.Sprintf("https://127.0.0.1:%d", adminPort)

 // Run tests.
 exitCode := m.Run()

 // Cleanup.
 _ = testCipherIMServer.Shutdown(context.Background())

 os.Exit(exitCode)
}
```

## Database Configuration Matrix

| DatabaseURL | DatabaseContainer | Result |
|-------------|------------------|--------|
| `""` (empty) | `"required"` | PostgreSQL testcontainer (fails if unavailable) |
| `""` (empty) | `"preferred"` | PostgreSQL testcontainer with fallback to error |
| `"file::memory:?cache=shared"` | Any | SQLite in-memory |
| `"postgres://user:pass@host:5432/db"` | `"disabled"` | External PostgreSQL |
| `"postgres://user:pass@host:5432/db"` | `"preferred"` | Testcontainer, fallback to external |

## Migration Checklist

### Step 1: Update Integration Tests

- [  ] Replace manual container setup with `NewFromConfig`
- [ ] Set `DatabaseContainer: "required"` for PostgreSQL
- [ ] Remove manual cleanup (handled by `Shutdown`)

### Step 2: Update E2E Tests

- [ ] Replace manual DB setup with `NewFromConfig`
- [ ] Set `DatabaseURL: "file::memory:?cache=shared"` for SQLite
- [ ] Set `DatabaseContainer: "disabled"`
- [ ] Remove manual cleanup (handled by `Shutdown`)

### Step 3: Validate

- [ ] Run integration tests: `go test ./internal/cipher/integration -v`
- [ ] Run E2E tests: `go test ./internal/cipher/e2e -v`
- [ ] Verify automatic cleanup (no hanging containers)

## Future Enhancements

### Main.go Support

The same pattern will work in `cmd/cipher-im/main.go`:

```go
func main() {
 cfg := loadConfig() // Load from flags/env/file.

 ctx := context.Background()

 server, err := server.NewFromConfig(ctx, cfg)
 if err != nil {
  log.Fatalf("failed to create server: %v", err)
 }

 // Start server (blocks until shutdown).
 if err := server.Start(ctx); err != nil {
  log.Fatalf("server failed: %v", err)
 }
}
```

### Other Services

Once cipher-im validates the pattern, migrate:

1. jose-ja (Phase 2)
2. pki-ca (Phase 2)
3. identity services (Phase 2)
4. sm-kms (Phase 3 - after all others validated)

See `internal/template/server/application/README.md` for complete migration plan.
