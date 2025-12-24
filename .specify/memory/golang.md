# Go Project Standards - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/03-03.golang.instructions.md`

## Go Version Consistency

**MANDATORY: Use same Go version everywhere** (development, CI/CD, Docker, documentation)

**Current Version**: 1.25.5 (check `go.mod`)
**See**: `.specify/memory/versions.md` for version table

**Enforcement Locations**:
- `go.mod`: `go 1.25.5`
- `.github/workflows/*.yml`: `GO_VERSION: '1.25.5'`
- `Dockerfile`: `FROM golang:1.25.5-alpine`
- `README.md`: Document Go 1.25.5+ requirement

---

## CGO Ban - CRITICAL

**CGO_ENABLED=0 is MANDATORY** (except race detector)

**Rationale**:
- Maximum portability (no C toolchain dependencies)
- Static linking (single binary deployment)
- Cross-compilation (build Linux binaries on Windows/macOS)
- No C library version conflicts

**ONLY Exception**: Race detector (`go test -race`) requires CGO_ENABLED=1 (Go toolchain limitation using C-based ThreadSanitizer from LLVM)

**CGO-Free Alternatives**:
- ✅ `modernc.org/sqlite` (not `github.com/mattn/go-sqlite3`)
- ✅ Pure Go crypto (standard library)
- ✅ Pure Go networking (standard library)

**Detection**:
```bash
# Check for CGO dependencies
go list -u -m all | grep '\[.*\]$'
```

---

## Build Flags and Linking

### Static Linking

**Pattern: Static binaries with debug symbols**

```go
// Makefile or build script
LDFLAGS := -ldflags "-extldflags '-static' -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

go build $(LDFLAGS) -o ./bin/cryptoutil ./cmd/cryptoutil
```

**Validation**:
```bash
# Linux: Verify static linking (should show "statically linked")
ldd ./bin/cryptoutil

# Windows: Check dependencies
dumpbin /dependents ./bin/cryptoutil.exe
```

---

## Go Project Structure

### Standard Layout

Follow [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```
cryptoutil/
├── cmd/                    # Main applications (entry points)
│   ├── cryptoutil/         # Main CLI tool
│   ├── cicd/               # CI/CD utilities
│   └── demo/               # Demo applications
├── internal/               # Private application code
│   ├── identity/           # Identity domain (authn/authz)
│   ├── jose/               # JOSE domain (JWT/JWE/JWS)
│   ├── kms/                # KMS domain (key management)
│   ├── shared/             # Shared utilities
│   └── cmd/                # Internal command logic
├── pkg/                    # Public library code
├── api/                    # OpenAPI specs, generated code
├── configs/                # Configuration files
├── scripts/                # Build/test scripts
├── deployments/            # Docker Compose, Kubernetes manifests
├── test/                   # Additional test files (load tests, etc.)
└── docs/                   # Documentation
```

**Key Rules**:
- ❌ Avoid `/src` directory (redundant in Go)
- ❌ Avoid deep nesting (>3 levels indicates design issue)
- ✅ Use `/internal` for private code (enforced by compiler)
- ✅ Use `/pkg` for public libraries (safe for external import)

---

## Application Architecture

### Layered Architecture

```
main() [cmd/]
  → Application [internal/*/application/]
    → Business Logic [internal/*/service/, internal/*/domain/]
      → Repositories [internal/*/repository/]
        → Database/External Systems
```

**Dependency Rules**:
- Main depends on Application (creates instance)
- Application depends on Business Logic (orchestrates services)
- Business Logic depends on Repositories (data access interface)
- Repositories implement data access (concrete implementations)

**Configuration Management**:
- ✅ YAML files + CLI flags (explicit, version-controlled)
- ✅ Docker/Kubernetes secrets (sensitive data)
- ❌ Environment variables (NOT used for configuration)

**Design Patterns**:
- Constructor injection: `NewService(logger, repo, config)`
- Context propagation: Pass `context.Context` to all long-running ops
- Graceful shutdown: Listen for signals, close resources cleanly
- Factory pattern: Create instances with dependencies injected
- Error propagation: Wrap errors with context (`fmt.Errorf("failed to X: %w", err)`)

---

## Import Alias Conventions

### Internal Package Aliases

**Pattern: `cryptoutil<PackageName>` in camelCase**

**Defined in**: `.golangci.yml` importas section (source of truth)

**Common Aliases**:
```go
import (
    // Shared packages
    cryptoutilMagic "cryptoutil/internal/shared/magic"
    cryptoutilServer "cryptoutil/internal/server"
    cryptoutilIdentity "cryptoutil/internal/identity"

    // CICD packages
    cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"

    // Domain packages
    cryptoutilAuthz "cryptoutil/internal/identity/authz"
    cryptoutilJose "cryptoutil/internal/jose"
)
```

### Third-Party Package Aliases

**Standard Library**:
```go
import (
    crand "crypto/rand"  // Avoid conflict with math/rand
)
```

**External Packages**:
```go
import (
    // UUID
    googleUuid "github.com/google/uuid"

    // JOSE
    jose "github.com/go-jose/go-jose/v4"
    joseJwa "github.com/go-jose/go-jose/v4/jwa"
    joseJwe "github.com/go-jose/go-jose/v4/jwe"
    joseJwk "github.com/go-jose/go-jose/v4/jwk"
    joseJws "github.com/go-jose/go-jose/v4/jws"
    joseJwt "github.com/go-jose/go-jose/v4/jwt"
)
```

### Crypto Acronym Conventions

**ALWAYS use ALL CAPS for crypto acronyms**:
- RSA, EC, ECDSA, ECDH, HMAC, AES
- JWA, JWK, JWS, JWE
- ED25519, PKCS8, PEM, DER

**Examples**:
```go
// ✅ CORRECT
func NewRSAKey() {}
func GenerateECDSAKeyPair() {}
func ParsePKCS8PrivateKey() {}

// ❌ WRONG
func NewRsaKey() {}
func GenerateEcdsaKeyPair() {}
func ParsePkcs8PrivateKey() {}
```

---

## Magic Values Management

### Storage Locations

**Shared Constants**: `internal/shared/magic/magic_*.go`
- `magic_network.go` - Ports, timeouts, buffer sizes
- `magic_database.go` - Connection pool sizes, query timeouts
- `magic_cryptography.go` - Key sizes, iteration counts, salt lengths
- `magic_testing.go` - Test probabilities (`TestProbAlways`, `TestProbTenth`, etc.)

**Domain-Specific Constants**: `internal/<domain>/magic*.go`
- `internal/identity/magic/magic_authn.go` - Authentication timeouts
- `internal/identity/magic/magic_authz.go` - Authorization constants

### Naming Conventions

**Pattern: Descriptive names grouped by category**

```go
package magic

// Network constants
const (
    DefaultHTTPPort      = 8080
    DefaultAdminPort     = 9090
    DefaultReadTimeout   = 30 * time.Second
    DefaultWriteTimeout  = 30 * time.Second
    IPv4Loopback        = "127.0.0.1"
)

// Database constants
const (
    DBSQLiteBusyTimeout       = 30000  // milliseconds
    SQLiteMaxOpenConnections  = 5      // GORM transaction support
    PostgreSQLMaxOpenConns    = 25
    PostgreSQLMaxIdleConns    = 10
)

// Cryptography constants
const (
    DefaultPBKDF2Iterations = 600000   // OWASP 2023 recommendation
    DefaultHMACSaltLength   = 32       // 256 bits
    DefaultAESKeySize       = 256      // bits
)

// Testing constants
const (
    TestProbAlways  = 1.0   // 100% execution (base algorithms)
    TestProbQuarter = 0.25  // 25% execution (key size variants)
    TestProbTenth   = 0.1   // 10% execution (redundant variants)
)
```

---

## Key Takeaways

1. **Version Consistency**: Same Go version (1.25.5) everywhere (dev, CI/CD, Docker)
2. **CGO Ban**: CGO_ENABLED=0 always (except race detector)
3. **Static Linking**: Single binary deployment with debug symbols
4. **Project Structure**: Follow golang-standards/project-layout (cmd, internal, pkg, api)
5. **Import Aliases**: `cryptoutil<PackageName>` for internal, defined in `.golangci.yml`
6. **Crypto Acronyms**: ALL CAPS (RSA, ECDSA, HMAC, JWK, etc.)
7. **Magic Values**: Shared in `internal/shared/magic/`, domain-specific in `internal/<domain>/magic*.go`
8. **Architecture**: Layered (main → app → business → repo), dependency injection, graceful shutdown
