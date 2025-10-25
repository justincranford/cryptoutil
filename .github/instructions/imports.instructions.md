---
description: "Instructions for Go import alias naming conventions"
applyTo: "**/*.go"
---
# Go Import Alias Naming Instructions

## Import Alias Conventions

Follow these established naming patterns when importing packages with aliases:

### cryptoutil/** Packages (REQUIRED)

**ALL cryptoutil imports MUST use camelCase aliases starting with "cryptoutil" prefix.**
**NEVER use unaliased imports for cryptoutil packages.**

The importas linter in `.golangci.yml` enforces consistent aliasing for all cryptoutil packages.

Common cryptoutil import aliases:
- `cryptoutilOpenapiClient "cryptoutil/api/client"`
- `cryptoutilOpenapiModel "cryptoutil/api/model"`
- `cryptoutilOpenapiServer "cryptoutil/api/server"`
- `cryptoutilClient "cryptoutil/internal/client"`
- `cryptoutilCmd "cryptoutil/internal/cmd"`
- `cryptoutilAppErr "cryptoutil/internal/common/apperr"`
- `cryptoutilConfig "cryptoutil/internal/common/config"`
- `cryptoutilContainer "cryptoutil/internal/common/container"`
- `cryptoutilMagic "cryptoutil/internal/common/magic"`
- `cryptoutilPool "cryptoutil/internal/common/pool"`
- `cryptoutilTelemetry "cryptoutil/internal/common/telemetry"`
- `cryptoutilUtil "cryptoutil/internal/common/util"`
- `cryptoutilCombinations "cryptoutil/internal/common/util/combinations"`
- `cryptoutilDateTime "cryptoutil/internal/common/util/datetime"`
- `cryptoutilNetwork "cryptoutil/internal/common/util/network"`
- `cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"`
- `cryptoutilAsn1 "cryptoutil/internal/common/crypto/asn1"`
- `cryptoutilCertificate "cryptoutil/internal/common/crypto/certificate"`
- `cryptoutilDigests "cryptoutil/internal/common/crypto/digests"`
- `cryptoutilJose "cryptoutil/internal/common/crypto/jose"`
- `cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"`
- `cryptoutilServerApplication "cryptoutil/internal/server/application"`
- `cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"`
- `cryptoutilOpenapiHandler "cryptoutil/internal/server/handler"`
- `cryptoutilBarrierService "cryptoutil/internal/server/barrier"`
- `cryptoutilContentKeysService "cryptoutil/internal/server/barrier/contentkeysservice"`
- `cryptoutilIntermediateKeysService "cryptoutil/internal/server/barrier/intermediatekeysservice"`
- `cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"`
- `cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"`
- `cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"`
- `cryptoutilSQLRepository "cryptoutil/internal/server/repository/sqlrepository"`

**See `.golangci.yml` importas section for the complete list of required aliases.**

### Third-Party Packages


#### Google UUID
Use `googleUuid` for the UUID package:
- `googleUuid "github.com/google/uuid"`

#### UUID Versioning
- **ALWAYS use `uuid.NewV7()` instead of `uuid.New()` for UUIDs**
	- Version 7 (time-ordered) is preferred over version 4 (random)
- Use the alias `googleUuid` for `github.com/google/uuid` in all imports

#### JOSE/JWX Packages
Use `jose` prefix followed by module name:
- `joseJwa "github.com/lestrrat-go/jwx/v3/jwa"`
- `joseJwe "github.com/lestrrat-go/jwx/v3/jwe"`
- `joseJwk "github.com/lestrrat-go/jwx/v3/jwk"`
- `joseJws "github.com/lestrrat-go/jwx/v3/jws"`

## Go Naming Convention Notes

Follow standard Go naming conventions with cryptographic exceptions:

### Standard Go Rules
- **Acronyms at the beginning or end**: ALL CAPS (e.g., `JWKParser`, `ParseJWK`)
- **Acronyms in the middle**: camelCase (e.g., `parseRSAKey`, `generateAESToken`)

### Cryptographic Exceptions ⚡
**Use ALL CAPS for these standard crypto terms anywhere in identifiers:**
- `RSA` - Rivest-Shamir-Adleman
- `EC` - Elliptic Curve  
- `ECDSA` - Elliptic Curve Digital Signature Algorithm
- `ECDH` - Elliptic Curve Diffie-Hellman
- `HMAC` - Hash-based Message Authentication Code
- `AES` - Advanced Encryption Standard
- `JWA` - JSON Web Algorithm
- `JWK` - JSON Web Key
- `JWS` - JSON Web Signature
- `JWE` - JSON Web Encryption
- `ED25519` / `ED448` - Edwards-curve Digital Signature Algorithm
- `PKCS8` / `PKIX` - Public Key Cryptography Standards
- `CSR` - Certificate Signing Request
- `PEM` - Privacy Enhanced Mail
- `DER` - Distinguished Encoding Rules

**Examples:** `PEMTypeRSAPrivateKey`, `GenerateECDSAKeyPair`, `ValidateJWKHeaders`

**Rationale:** Cryptographic terms should remain recognizable to security professionals, even when appearing in the middle of identifiers.

## General Principles

- Use consistent, descriptive aliases that clearly identify the package purpose
- Avoid single-letter or cryptic aliases
- Maintain consistency across the entire codebase
- Prefer camelCase for multi-word aliases, except for crypto acronyms listed above
- Use the established prefix patterns for package families
