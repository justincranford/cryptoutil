---
description: "Instructions for Go import alias naming conventions"
applyTo: "**/*.go"
---
# Go Import Alias Naming Instructions

## Import Alias Conventions

Follow these established naming patterns when importing packages with aliases:

### cryptoutil/** Packages
Use `cryptoutil` prefix followed by descriptive module name in camelCase:
- `cryptoutilOpenapiModel "cryptoutil/api/model"`
- `cryptoutilOpenapiServer "cryptoutil/api/server"`
- `cryptoutilAppErr "cryptoutil/internal/common/apperr"`
- `cryptoutilConfig "cryptoutil/internal/common/config"`
- `cryptoutilContainer "cryptoutil/internal/common/container"`
- `cryptoutilDigests "cryptoutil/internal/common/crypto/digests"`
- `cryptoutilJose "cryptoutil/internal/common/crypto/jose"`
- `cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"`
- `cryptoutilPool "cryptoutil/internal/common/pool"`
- `cryptoutilTelemetry "cryptoutil/internal/common/telemetry"`
- `cryptoutilUtil "cryptoutil/internal/common/util"`
- `cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"`
- `cryptoutilCombinations "cryptoutil/internal/common/util/combinations"`
- `cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"`
- `cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"`
- `cryptoutilSQLRepository "cryptoutil/internal/server/repository/sqlrepository"`
- `cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"`
- `cryptoutilContentKeysService "cryptoutil/internal/server/barrier/contentkeysservice"`
- `cryptoutilIntermediateKeysService "cryptoutil/internal/server/barrier/intermediatekeysservice"`
- `cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"`

### Third-Party Packages

#### Google UUID
Use `googleUuid` for the UUID package:
- `googleUuid "github.com/google/uuid"`

#### JOSE/JWX Packages
Use `jose` prefix followed by module name:
- `joseJwa "github.com/lestrrat-go/jwx/v3/jwa"`
- `joseJwe "github.com/lestrrat-go/jwx/v3/jwe"`
- `joseJwk "github.com/lestrrat-go/jwx/v3/jwk"`
- `joseJws "github.com/lestrrat-go/jwx/v3/jws"`

## General Principles

- Use consistent, descriptive aliases that clearly identify the package purpose
- Avoid single-letter or cryptic aliases
- Maintain consistency across the entire codebase
- Prefer camelCase for multi-word aliases
- Use the established prefix patterns for package families