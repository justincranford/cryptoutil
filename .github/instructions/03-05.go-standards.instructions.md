---
description: "Instructions for Go coding standards: imports, dependencies, formatting, conditionals"
applyTo: "**/*.go"
---
# Go Coding Standards

## Import Alias Conventions

### cryptoutil/** Packages (REQUIRED)
**ALL cryptoutil imports MUST use camelCase aliases starting with "cryptoutil" prefix.**

Common aliases (enforced by importas linter in `.golangci.yml`):
- `cryptoutilOpenapiClient "cryptoutil/api/client"`
- `cryptoutilOpenapiModel "cryptoutil/api/model"`
- `cryptoutilOpenapiServer "cryptoutil/api/server"`
- `cryptoutilClient "cryptoutil/internal/client"`
- `cryptoutilAppErr "cryptoutil/internal/common/apperr"`
- `cryptoutilConfig "cryptoutil/internal/common/config"`
- `cryptoutilContainer "cryptoutil/internal/common/container"`
- `cryptoutilMagic "cryptoutil/internal/common/magic"`
- `cryptoutilPool "cryptoutil/internal/common/pool"`
- `cryptoutilTelemetry "cryptoutil/internal/common/telemetry"`
- `cryptoutilUtil "cryptoutil/internal/common/util"`
- `cryptoutilDateTime "cryptoutil/internal/common/util/datetime"`
- `cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"`
- `cryptoutilServerApplication "cryptoutil/internal/server/application"`
- `cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"`
- `cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"`

### Third-Party Packages
- **Google UUID**: `googleUuid "github.com/google/uuid"` - ALWAYS use `uuid.NewV7()` for time-ordered UUIDs
- **JOSE/JWX**: `joseJwa`, `joseJwe`, `joseJwk`, `joseJws` for `github.com/lestrrat-go/jwx/v3/*`

### Crypto Acronym Exceptions
**Use ALL CAPS for these terms anywhere in identifiers:**
- RSA, EC, ECDSA, ECDH, HMAC, AES, JWA, JWK, JWS, JWE, ED25519, ED448, PKCS8, PKIX, CSR, PEM, DER

**Examples:** `PEMTypeRSAPrivateKey`, `GenerateECDSAKeyPair`, `ValidateJWKHeaders`

## Dependency Management

### Updates
- Check updates: `go list -u -m all | grep '\[.*\]$'`
- Update incrementally: `go get <package>@<version>`
- Clean up: `go mod tidy`
- Test after each: `go test ./... --count=1 -timeout=20m`

### Version Synchronization
- **Synchronize ecosystem packages**: Update related packages to latest compatible version
- **Example**: OpenTelemetry packages (`go.opentelemetry.io/*`), GORM ecosystem
- **Ignore unrelated transients**: Let Go's MVS handle unrelated dependencies
- Use `go mod why <package>` to understand indirect dependencies

### Best Practices
- Update direct dependencies first, then ecosystem-related indirect ones
- Keep related packages at consistent versions
- Review changelog/release notes for breaking changes
- Prefer stable releases; avoid pre-releases
- Update in small batches to isolate issues

## Formatting Standards

- **Encoding**: UTF-8 without BOM, single newline at EOF, no trailing whitespace
- **Indentation**: 4 spaces (Go), 2 spaces (YAML, JSON, Markdown, Dockerfile)
- **Types**: Use `any` not `interface{}`
- **Tool**: gofumpt (strict superset of gofmt)
- **Auto-formatting**: Use `golangci-lint run --fix` (runs gofumpt automatically)

## Conditional Statement Chaining

### Pattern: Chain Mutually Exclusive Conditions

**Prefer chained if/else if/else:**
```go
if ctx == nil {
    return nil, fmt.Errorf("context cannot be nil")
} else if logger == nil {
    return nil, fmt.Errorf("logger cannot be nil")
} else if description == "" {
    return nil, fmt.Errorf("description cannot be empty")
}
```

**Avoid separate if statements:**
```go
if ctx == nil {
    return nil, fmt.Errorf("context cannot be nil")
}

if logger == nil {
    return nil, fmt.Errorf("logger cannot be nil")
}
```

### When NOT to Chain
- Independent conditions (not mutually exclusive)
- Error accumulation patterns
- Cases with early returns
