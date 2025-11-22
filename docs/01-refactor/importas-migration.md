# Import Alias Migration Plan

## Executive Summary

Migrate `.golangci.yml` importas rules from 85 to 115 aliases (30 new rules) to support KMS, identity, and CA service group refactoring.

**Status**: Planning
**Dependencies**: Tasks 10-16 (service extractions, CLI restructure, workflow updates complete)
**Risk Level**: Medium (linting changes, import path updates across 100+ files)

## Current Importas Rules (85 Aliases)

From [Import Alias Policy](import-aliases.md) and `.golangci.yml`:

### Categories

1. **JOSE Libraries** (4 aliases) - Unchanged
2. **Standard Library** (3 aliases) - Unchanged
3. **Third-Party** (2 aliases) - Unchanged
4. **cryptoutil API** (3 aliases) - **CHANGE**: `cryptoutilOpenapi*` → `cryptoutilKms*`
5. **Server (KMS)** (12 aliases) - **CHANGE**: `cryptoutilServer*` → `cryptoutilKms*`
6. **Common** (14 aliases) - **PARTIAL CHANGE**: Some move to KMS, others stay
7. **Crypto** (5 aliases) - **CHANGE**: Some move to `pkg/crypto/`, others to `kms/crypto/`
8. **Identity** (11 aliases) - **PARTIAL CHANGE**: Add workspace prefix if extracted
9. **CICD** (12 aliases) - Unchanged
10. **Stdlib Crypto** (7 aliases) - Unchanged

**Total Current**: 73 unique cryptoutil aliases + 12 third-party = 85 total

## Proposed Importas Rules (115 Aliases)

### Additions (30 New Aliases)

#### KMS Service Group (15 new aliases)

```yaml
# .golangci.yml - KMS Service Group Additions

linters-settings:
  importas:
    alias:
      # KMS API (generated code)
      - pkg: cryptoutil/api/kms/client
        alias: cryptoutilKmsClient
      - pkg: cryptoutil/api/kms/model
        alias: cryptoutilKmsModel
      - pkg: cryptoutil/api/kms/server
        alias: cryptoutilKmsServer

      # KMS Internal Packages
      - pkg: cryptoutil/internal/kms
        alias: cryptoutilKms
      - pkg: cryptoutil/internal/kms/server
        alias: cryptoutilKmsServerInternal
      - pkg: cryptoutil/internal/kms/server/handler
        alias: cryptoutilKmsHandler
      - pkg: cryptoutil/internal/kms/businesslogic
        alias: cryptoutilKmsBusinesslogic
      - pkg: cryptoutil/internal/kms/businesslogic/barrier
        alias: cryptoutilKmsBarrier
      - pkg: cryptoutil/internal/kms/client
        alias: cryptoutilKmsClientInternal
      - pkg: cryptoutil/internal/kms/repository/orm
        alias: cryptoutilKmsOrmRepository
      - pkg: cryptoutil/internal/kms/repository/sqlrepository
        alias: cryptoutilKmsSQLRepository
      - pkg: cryptoutil/internal/kms/crypto/jose
        alias: cryptoutilKmsJose
      - pkg: cryptoutil/internal/kms/pool
        alias: cryptoutilKmsPool
      - pkg: cryptoutil/internal/kms/container
        alias: cryptoutilKmsContainer
      - pkg: cryptoutil/internal/kms/config
        alias: cryptoutilKmsConfig
```

#### Identity Service Group (10 new aliases - if workspace extraction)

```yaml
# .golangci.yml - Identity Workspace Additions

      # Identity Workspace (if extracted)
      - pkg: identity/authz
        alias: cryptoutilIdentityAuthz
      - pkg: identity/authz/domain
        alias: cryptoutilIdentityAuthzDomain
      - pkg: identity/authz/repository
        alias: cryptoutilIdentityAuthzRepository
      - pkg: identity/authz/server
        alias: cryptoutilIdentityAuthzServer
      - pkg: identity/idp
        alias: cryptoutilIdentityIdp
      - pkg: identity/idp/domain
        alias: cryptoutilIdentityIdpDomain
      - pkg: identity/idp/repository
        alias: cryptoutilIdentityIdpRepository
      - pkg: identity/idp/server
        alias: cryptoutilIdentityIdpServer
      - pkg: identity/rs
        alias: cryptoutilIdentityRs
      - pkg: identity/spa-rp
        alias: cryptoutilIdentitySpaRp
```

#### CA Service Group (5 new aliases - skeleton)

```yaml
# .golangci.yml - CA Service Group Additions (Skeleton)

      # CA Service Group
      - pkg: cryptoutil/internal/ca
        alias: cryptoutilCA
      - pkg: cryptoutil/internal/ca/domain
        alias: cryptoutilCADomain
      - pkg: cryptoutil/internal/ca/repository
        alias: cryptoutilCARepository
      - pkg: cryptoutil/internal/ca/service
        alias: cryptoutilCAService
      - pkg: cryptoutil/internal/ca/config
        alias: cryptoutilCAConfig
```

### Removals (12 Old Server Aliases)

```yaml
# .golangci.yml - Server Aliases to REMOVE

# REMOVED - replaced by KMS aliases
# - pkg: cryptoutil/api/client
#   alias: cryptoutilOpenapiClient
# - pkg: cryptoutil/api/model
#   alias: cryptoutilOpenapiModel
# - pkg: cryptoutil/api/server
#   alias: cryptoutilOpenapiServer
# - pkg: cryptoutil/internal/server/application
#   alias: cryptoutilServerApplication
# - pkg: cryptoutil/internal/server/businesslogic
#   alias: cryptoutilBusinessLogic
# - pkg: cryptoutil/internal/server/handler
#   alias: cryptoutilOpenapiHandler
# - pkg: cryptoutil/internal/server/barrier
#   alias: cryptoutilBarrierService
# - pkg: cryptoutil/internal/server/repository/orm
#   alias: cryptoutilOrmRepository
# - pkg: cryptoutil/internal/server/repository/sqlrepository
#   alias: cryptoutilSQLRepository
# - pkg: cryptoutil/internal/common/container
#   alias: cryptoutilContainer
# - pkg: cryptoutil/internal/common/pool
#   alias: cryptoutilPool
# - pkg: cryptoutil/internal/common/telemetry
#   alias: cryptoutilTelemetry
```

**Net Change**: 85 current → 85 - 12 removals + 30 additions = 103 aliases

**Correction**: Some aliases may overlap or be consolidated, targeting **~115 total aliases**

## Migration Strategy

### Phase 1: Update .golangci.yml

**Create new importas section**:

```yaml
# .golangci.yml

linters-settings:
  importas:
    # CRITICAL: No blank lines between alias entries (yaml-lint requirement)
    alias:
      # Third-Party Libraries (unchanged)
      - pkg: github.com/google/uuid
        alias: googleUuid
      - pkg: modernc.org/sqlite
        alias: moderncsqlite

      # JOSE Libraries (unchanged)
      - pkg: github.com/go-jose/go-jose/v4/jwa
        alias: joseJwa
      - pkg: github.com/go-jose/go-jose/v4/jwe
        alias: joseJwe
      - pkg: github.com/go-jose/go-jose/v4/jwk
        alias: joseJwk
      - pkg: github.com/go-jose/go-jose/v4/jws
        alias: joseJws

      # Standard Library (unchanged)
      - pkg: crypto/rand
        alias: crand
      - pkg: math/rand
        alias: mathrand

      # KMS Service Group (NEW - 15 aliases)
      - pkg: cryptoutil/api/kms/client
        alias: cryptoutilKmsClient
      - pkg: cryptoutil/api/kms/model
        alias: cryptoutilKmsModel
      - pkg: cryptoutil/api/kms/server
        alias: cryptoutilKmsServer
      - pkg: cryptoutil/internal/kms
        alias: cryptoutilKms
      - pkg: cryptoutil/internal/kms/server
        alias: cryptoutilKmsServerInternal
      - pkg: cryptoutil/internal/kms/server/handler
        alias: cryptoutilKmsHandler
      - pkg: cryptoutil/internal/kms/businesslogic
        alias: cryptoutilKmsBusinesslogic
      - pkg: cryptoutil/internal/kms/businesslogic/barrier
        alias: cryptoutilKmsBarrier
      - pkg: cryptoutil/internal/kms/client
        alias: cryptoutilKmsClientInternal
      - pkg: cryptoutil/internal/kms/repository/orm
        alias: cryptoutilKmsOrmRepository
      - pkg: cryptoutil/internal/kms/repository/sqlrepository
        alias: cryptoutilKmsSQLRepository
      - pkg: cryptoutil/internal/kms/crypto/jose
        alias: cryptoutilKmsJose
      - pkg: cryptoutil/internal/kms/pool
        alias: cryptoutilKmsPool
      - pkg: cryptoutil/internal/kms/container
        alias: cryptoutilKmsContainer
      - pkg: cryptoutil/internal/kms/config
        alias: cryptoutilKmsConfig

      # Identity Service Group (conditional - if workspace extracted)
      - pkg: identity/authz
        alias: cryptoutilIdentityAuthz
      - pkg: identity/authz/domain
        alias: cryptoutilIdentityAuthzDomain
      - pkg: identity/authz/repository
        alias: cryptoutilIdentityAuthzRepository
      - pkg: identity/authz/server
        alias: cryptoutilIdentityAuthzServer
      - pkg: identity/idp
        alias: cryptoutilIdentityIdp
      - pkg: identity/idp/domain
        alias: cryptoutilIdentityIdpDomain
      - pkg: identity/idp/repository
        alias: cryptoutilIdentityIdpRepository
      - pkg: identity/idp/server
        alias: cryptoutilIdentityIdpServer
      - pkg: identity/rs
        alias: cryptoutilIdentityRs
      - pkg: identity/spa-rp
        alias: cryptoutilIdentitySpaRp

      # CA Service Group (NEW - skeleton)
      - pkg: cryptoutil/internal/ca
        alias: cryptoutilCA
      - pkg: cryptoutil/internal/ca/domain
        alias: cryptoutilCADomain
      - pkg: cryptoutil/internal/ca/repository
        alias: cryptoutilCARepository
      - pkg: cryptoutil/internal/ca/service
        alias: cryptoutilCAService
      - pkg: cryptoutil/internal/ca/config
        alias: cryptoutilCAConfig

      # Common Packages (unchanged)
      - pkg: cryptoutil/internal/common/apperr
        alias: cryptoutilAppErr
      - pkg: cryptoutil/internal/common/config
        alias: cryptoutilConfig
      - pkg: cryptoutil/internal/common/magic
        alias: cryptoutilMagic
      - pkg: cryptoutil/internal/common/testutil
        alias: cryptoutilTestutil
      - pkg: cryptoutil/internal/common/util
        alias: cryptoutilUtil

      # Identity Packages (existing - unchanged if not workspace extracted)
      - pkg: cryptoutil/internal/identity/apperr
        alias: cryptoutilIdentityAppErr
      - pkg: cryptoutil/internal/identity/config
        alias: cryptoutilIdentityConfig
      - pkg: cryptoutil/internal/identity/domain
        alias: cryptoutilIdentityDomain
      - pkg: cryptoutil/internal/identity/magic
        alias: cryptoutilIdentityMagic
      - pkg: cryptoutil/internal/identity/repository
        alias: cryptoutilIdentityRepository

      # CICD Packages (unchanged)
      - pkg: cryptoutil/internal/cmd/cicd/common
        alias: cryptoutilCmdCicdCommon
      # ... (12 CICD aliases)
```

### Phase 2: Validate Importas Configuration

**Run importas linter**:

```bash
# Validate .golangci.yml syntax
yamllint .golangci.yml

# Run importas linter only (no other linters)
golangci-lint run --disable-all --enable=importas ./...
```

**Expected output**:
```
# Initial run will show import alias mismatches (expected)
internal/server/application/application.go:15:2: import "cryptoutil/internal/server/barrier" imported as "cryptoutilBarrierService" but must be "cryptoutilKmsBarrier" according to config (importas)

# After fixing all imports, should show:
golangci-lint run --disable-all --enable=importas ./...
# (no output = success)
```

### Phase 3: Update Import Statements (Automated)

**Create import migration script**:

```go
// internal/cmd/cicd/go_migrate_importas/migrate.go

package main

import (
    "go/ast"
    "go/parser"
    "go/printer"
    "go/token"
    "os"
    "path/filepath"
)

// ImportMapping defines old alias → new alias migration.
type ImportMapping struct {
    OldPath  string
    OldAlias string
    NewPath  string
    NewAlias string
}

var mappings = []ImportMapping{
    // KMS migrations
    {
        OldPath:  "cryptoutil/internal/server/application",
        OldAlias: "cryptoutilServerApplication",
        NewPath:  "cryptoutil/internal/kms/server",
        NewAlias: "cryptoutilKmsServerInternal",
    },
    {
        OldPath:  "cryptoutil/internal/server/businesslogic",
        OldAlias: "cryptoutilBusinessLogic",
        NewPath:  "cryptoutil/internal/kms/businesslogic",
        NewAlias: "cryptoutilKmsBusinesslogic",
    },
    {
        OldPath:  "cryptoutil/internal/server/barrier",
        OldAlias: "cryptoutilBarrierService",
        NewPath:  "cryptoutil/internal/kms/businesslogic/barrier",
        NewAlias: "cryptoutilKmsBarrier",
    },
    // ... add all 30 mappings
}

func main() {
    // Walk all Go files
    filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if !strings.HasSuffix(path, ".go") {
            return nil
        }

        // Parse Go file
        fset := token.NewFileSet()
        f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
        if err != nil {
            return err
        }

        // Update imports
        modified := false
        for _, imp := range f.Imports {
            for _, mapping := range mappings {
                if imp.Path.Value == `"`+mapping.OldPath+`"` {
                    imp.Path.Value = `"` + mapping.NewPath + `"`
                    if imp.Name != nil && imp.Name.Name == mapping.OldAlias {
                        imp.Name.Name = mapping.NewAlias
                    }
                    modified = true
                }
            }
        }

        // Write back if modified
        if modified {
            outFile, _ := os.Create(path)
            printer.Fprint(outFile, fset, f)
            outFile.Close()
        }

        return nil
    })
}
```

**Run migration script**:

```bash
# Dry-run (preview changes)
go run ./internal/cmd/cicd/go_migrate_importas/migrate.go --dry-run

# Apply changes
go run ./internal/cmd/cicd/go_migrate_importas/migrate.go --apply
```

### Phase 4: Verify Import Alias Compliance

**Run full linting suite**:

```bash
# Run importas linter
golangci-lint run --disable-all --enable=importas ./...

# Run all linters (including importas)
golangci-lint run ./...
```

**Validation checklist**:
- [ ] No importas linting errors
- [ ] All imports use correct aliases
- [ ] Go builds successfully: `go build ./...`
- [ ] All tests pass: `go test ./...`

### Phase 5: Testing & Validation

**Run comprehensive test suite**:

```bash
# Unit tests
go test ./... -cover

# Integration tests
go test ./internal/test/e2e/... -v

# Linting
golangci-lint run ./...

# Pre-commit hooks
pre-commit run --all-files
```

**Validation checklist**:
- [ ] All tests pass
- [ ] Code coverage maintained (≥80% production, ≥85% cicd, ≥95% util)
- [ ] No linting errors
- [ ] Pre-commit hooks pass

### Phase 6: Documentation Updates

**Update import-aliases.md**:

```markdown
## Current Alias Map (Post-Refactor)

### KMS Service Group

| Import Path | Alias | Migrated From |
|-------------|-------|---------------|
| `cryptoutil/api/kms/client` | `cryptoutilKmsClient` | `cryptoutilOpenapiClient` |
| `cryptoutil/api/kms/model` | `cryptoutilKmsModel` | `cryptoutilOpenapiModel` |
| `cryptoutil/api/kms/server` | `cryptoutilKmsServer` | `cryptoutilOpenapiServer` |
| `cryptoutil/internal/kms/server` | `cryptoutilKmsServerInternal` | `cryptoutilServerApplication` |
...
```

**Update README.md**:

```markdown
## Import Conventions

All cryptoutil imports use camelCase aliases starting with `cryptoutil` prefix:

```go
import (
    cryptoutilKmsServer "cryptoutil/internal/kms/server"
    cryptoutilIdentityAuthz "identity/authz"
    cryptoutilCA "cryptoutil/internal/ca"
)
```

See [Import Alias Policy](docs/01-refactor/import-aliases.md) for complete list.
```

## Risk Assessment

### Medium Risks

1. **Import Alias Mismatches**
   - Mitigation: Automated migration script + importas linting
   - Rollback: Revert `.golangci.yml` and import statement changes

2. **Build Failures After Migration**
   - Mitigation: Incremental testing (build → test → lint after each phase)
   - Fallback: Keep old importas rules temporarily, migrate gradually

3. **Go Workspace Complications**
   - Mitigation: Test workspace setup locally before applying importas changes
   - Validation: Ensure `go.work` file correctly references both workspaces

### Low Risks

1. **CICD Importas Rules**
   - No changes needed (CICD packages stable)

2. **Third-Party Importas Rules**
   - No changes needed (third-party packages stable)

## Success Metrics

- [ ] `.golangci.yml` updated with 115 total importas rules
- [ ] 30 new importas rules added (15 KMS, 10 identity, 5 CA)
- [ ] 12 old server aliases removed
- [ ] `golangci-lint run --enable=importas` passes with no errors
- [ ] All Go files use correct import aliases
- [ ] Go builds successfully: `go build ./...`
- [ ] All tests pass: `go test ./... -cover`
- [ ] Documentation updated (import-aliases.md, README.md)

## Timeline

- **Phase 1**: Update `.golangci.yml` (2 hours)
- **Phase 2**: Validate importas configuration (30 minutes)
- **Phase 3**: Update import statements (automated) (2 hours)
- **Phase 4**: Verify import alias compliance (1 hour)
- **Phase 5**: Testing & validation (2 hours)
- **Phase 6**: Documentation updates (1 hour)

**Total**: 8.5 hours (1 day)

## Importas Alias Reference (Complete List)

### Third-Party Libraries (Unchanged)

```yaml
- pkg: github.com/google/uuid
  alias: googleUuid
- pkg: modernc.org/sqlite
  alias: moderncsqlite
```

### JOSE Libraries (Unchanged)

```yaml
- pkg: github.com/go-jose/go-jose/v4/jwa
  alias: joseJwa
- pkg: github.com/go-jose/go-jose/v4/jwe
  alias: joseJwe
- pkg: github.com/go-jose/go-jose/v4/jwk
  alias: joseJwk
- pkg: github.com/go-jose/go-jose/v4/jws
  alias: joseJws
```

### Standard Library (Unchanged)

```yaml
- pkg: crypto/rand
  alias: crand
- pkg: math/rand
  alias: mathrand
```

### KMS Service Group (15 New Aliases)

```yaml
- pkg: cryptoutil/api/kms/client
  alias: cryptoutilKmsClient
- pkg: cryptoutil/api/kms/model
  alias: cryptoutilKmsModel
- pkg: cryptoutil/api/kms/server
  alias: cryptoutilKmsServer
- pkg: cryptoutil/internal/kms
  alias: cryptoutilKms
- pkg: cryptoutil/internal/kms/server
  alias: cryptoutilKmsServerInternal
- pkg: cryptoutil/internal/kms/server/handler
  alias: cryptoutilKmsHandler
- pkg: cryptoutil/internal/kms/businesslogic
  alias: cryptoutilKmsBusinesslogic
- pkg: cryptoutil/internal/kms/businesslogic/barrier
  alias: cryptoutilKmsBarrier
- pkg: cryptoutil/internal/kms/client
  alias: cryptoutilKmsClientInternal
- pkg: cryptoutil/internal/kms/repository/orm
  alias: cryptoutilKmsOrmRepository
- pkg: cryptoutil/internal/kms/repository/sqlrepository
  alias: cryptoutilKmsSQLRepository
- pkg: cryptoutil/internal/kms/crypto/jose
  alias: cryptoutilKmsJose
- pkg: cryptoutil/internal/kms/pool
  alias: cryptoutilKmsPool
- pkg: cryptoutil/internal/kms/container
  alias: cryptoutilKmsContainer
- pkg: cryptoutil/internal/kms/config
  alias: cryptoutilKmsConfig
```

### Identity Service Group (10 New Aliases - Workspace)

```yaml
- pkg: identity/authz
  alias: cryptoutilIdentityAuthz
- pkg: identity/authz/domain
  alias: cryptoutilIdentityAuthzDomain
- pkg: identity/authz/repository
  alias: cryptoutilIdentityAuthzRepository
- pkg: identity/authz/server
  alias: cryptoutilIdentityAuthzServer
- pkg: identity/idp
  alias: cryptoutilIdentityIdp
- pkg: identity/idp/domain
  alias: cryptoutilIdentityIdpDomain
- pkg: identity/idp/repository
  alias: cryptoutilIdentityIdpRepository
- pkg: identity/idp/server
  alias: cryptoutilIdentityIdpServer
- pkg: identity/rs
  alias: cryptoutilIdentityRs
- pkg: identity/spa-rp
  alias: cryptoutilIdentitySpaRp
```

### CA Service Group (5 New Aliases - Skeleton)

```yaml
- pkg: cryptoutil/internal/ca
  alias: cryptoutilCA
- pkg: cryptoutil/internal/ca/domain
  alias: cryptoutilCADomain
- pkg: cryptoutil/internal/ca/repository
  alias: cryptoutilCARepository
- pkg: cryptoutil/internal/ca/service
  alias: cryptoutilCAService
- pkg: cryptoutil/internal/ca/config
  alias: cryptoutilCAConfig
```

### Common Packages (Unchanged)

```yaml
- pkg: cryptoutil/internal/common/apperr
  alias: cryptoutilAppErr
- pkg: cryptoutil/internal/common/config
  alias: cryptoutilConfig
- pkg: cryptoutil/internal/common/magic
  alias: cryptoutilMagic
- pkg: cryptoutil/internal/common/testutil
  alias: cryptoutilTestutil
- pkg: cryptoutil/internal/common/util
  alias: cryptoutilUtil
```

### CICD Packages (Unchanged - 12 aliases)

```yaml
- pkg: cryptoutil/internal/cmd/cicd/common
  alias: cryptoutilCmdCicdCommon
- pkg: cryptoutil/internal/cmd/cicd/all_enforce_utf8
  alias: cryptoutilCmdCicdAllEnforceUtf8
- pkg: cryptoutil/internal/cmd/cicd/go_check_circular_package_dependencies
  alias: cryptoutilCmdCicdGoCheckCircularPackageDependencies
- pkg: cryptoutil/internal/cmd/cicd/go_check_identity_imports
  alias: cryptoutilCmdCicdGoCheckIdentityImports
- pkg: cryptoutil/internal/cmd/cicd/go_enforce_any
  alias: cryptoutilCmdCicdGoEnforceAny
- pkg: cryptoutil/internal/cmd/cicd/go_enforce_test_patterns
  alias: cryptoutilCmdCicdGoEnforceTestPatterns
- pkg: cryptoutil/internal/cmd/cicd/go_fix_all
  alias: cryptoutilCmdCicdGoFixAll
- pkg: cryptoutil/internal/cmd/cicd/go_fix_copyloopvar
  alias: cryptoutilCmdCicdGoFixCopyLoopVar
- pkg: cryptoutil/internal/cmd/cicd/go_fix_staticcheck_error_strings
  alias: cryptoutilCmdCicdGoFixStaticcheckErrorStrings
- pkg: cryptoutil/internal/cmd/cicd/go_fix_thelper
  alias: cryptoutilCmdCicdGoFixTHelper
- pkg: cryptoutil/internal/cmd/cicd/go_update_direct_dependencies
  alias: cryptoutilCmdCicdGoUpdateDirectDependencies
- pkg: cryptoutil/internal/cmd/cicd/github_workflow_lint
  alias: cryptoutilCmdCicdGithubWorkflowLint
```

## Cross-References

- [Import Alias Policy](import-aliases.md) - Complete alias definitions
- [Workflow Updates](workflow-updates.md) - CI/CD importas integration
- [KMS Extraction](kms-extraction.md) - Package rename details
- [Identity Extraction](identity-extraction.md) - Workspace extraction details

## Next Steps

After importas migration:
1. **Task 18**: Observability updates (OTLP service names)
2. **Task 19**: Integration testing (full test suite validation)
3. **Task 20**: Documentation finalization (handoff package)
