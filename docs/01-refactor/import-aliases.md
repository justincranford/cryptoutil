# Import Alias Policy

## Overview

This document defines the import alias policy for the cryptoutil multi-service repository. It extends `.golangci.yml` importas rules to support the new service group structure defined in the blueprint.

**Cross-references:**
- [Service Groups Taxonomy](./service-groups.md) - Defines 43 service groups
- [Dependency Analysis](./dependency-analysis.md) - Identifies coupling risks
- [Group Directory Blueprint](./blueprint.md) - Defines target directory structure

---

## Alias Naming Convention

### Core Rule

**ALL cryptoutil imports MUST use camelCase aliases starting with "cryptoutil" prefix**

**Pattern:** `cryptoutil<ServiceGroup><Package>`

**Examples:**
```go
import cryptoutilKmsServer "cryptoutil/internal/kms/server"
import cryptoutilKmsBarrier "cryptoutil/internal/kms/businesslogic/barrier"
import cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
import cryptoutilCaBusinesslogic "cryptoutil/internal/ca/businesslogic"
```

---

## Current Alias Map (Pre-Refactor)

### API Packages

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/api/client` | `cryptoutilOpenapiClient` | KMS client (will move to api/kms/client) |
| `cryptoutil/api/model` | `cryptoutilOpenapiModel` | KMS models (will move to api/kms/model) |
| `cryptoutil/api/server` | `cryptoutilOpenapiServer` | KMS server (will move to api/kms/server) |

### Server Packages

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/internal/server/application` | `cryptoutilServerApplication` | Will move to kms/server/application |
| `cryptoutil/internal/server/businesslogic` | `cryptoutilBusinessLogic` | Will move to kms/businesslogic |
| `cryptoutil/internal/server/handler` | `cryptoutilOpenapiHandler` | Will move to kms/server/handler |
| `cryptoutil/internal/server/barrier` | `cryptoutilBarrierService` | Will move to kms/businesslogic/barrier |
| `cryptoutil/internal/server/barrier/contentkeysservice` | `cryptoutilContentKeysService` | Will move to kms/businesslogic/barrier/contentkeysservice |
| `cryptoutil/internal/server/barrier/intermediatekeysservice` | `cryptoutilIntermediateKeysService` | Will move to kms/businesslogic/barrier/intermediatekeysservice |
| `cryptoutil/internal/server/barrier/rootkeysservice` | `cryptoutilRootKeysService` | Will move to kms/businesslogic/barrier/rootkeysservice |
| `cryptoutil/internal/server/barrier/unsealkeysservice` | `cryptoutilUnsealKeysService` | Will move to kms/businesslogic/barrier/unsealkeysservice |
| `cryptoutil/internal/server/repository/orm` | `cryptoutilOrmRepository` | Will move to kms/repository/orm |
| `cryptoutil/internal/server/repository/sqlrepository` | `cryptoutilSQLRepository` | Will move to kms/repository/sqlrepository |

### Common Packages

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/internal/common/apperr` | `cryptoutilAppErr` | Stays in common |
| `cryptoutil/internal/common/config` | `cryptoutilConfig` | Stays in common |
| `cryptoutil/internal/common/container` | `cryptoutilContainer` | Will move to kms/container |
| `cryptoutil/internal/common/magic` | `cryptoutilMagic` | Stays in common |
| `cryptoutil/internal/common/pool` | `cryptoutilPool` | Will move to kms/pool |
| `cryptoutil/internal/common/telemetry` | `cryptoutilTelemetry` | Will move to kms/telemetry |
| `cryptoutil/internal/common/testutil` | `cryptoutilTestutil` | Stays in common |
| `cryptoutil/internal/common/util` | `cryptoutilUtil` | Stays in common |
| `cryptoutil/internal/common/util/combinations` | `cryptoutilCombinations` | Stays in common |
| `cryptoutil/internal/common/util/datetime` | `cryptoutilDateTime` | Stays in common |
| `cryptoutil/internal/common/util/files` | `cryptoutilFiles` | Stays in common |
| `cryptoutil/internal/common/util/network` | `cryptoutilNetwork` | Stays in common |
| `cryptoutil/internal/common/util/sysinfo` | `cryptoutilSysinfo` | Stays in common |

### Crypto Packages

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/internal/common/crypto/asn1` | `cryptoutilAsn1` | Will move to pkg/crypto/asn1 |
| `cryptoutil/internal/common/crypto/certificate` | `cryptoutilCertificate` | Will move to pkg/crypto/certificate |
| `cryptoutil/internal/common/crypto/digests` | `cryptoutilDigests` | Will move to pkg/crypto/digests |
| `cryptoutil/internal/common/crypto/jose` | `cryptoutilJose` | Will move to kms/crypto/jose |
| `cryptoutil/internal/common/crypto/keygen` | `cryptoutilKeyGen` | Will move to pkg/crypto/keygen |

### Identity Packages

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/internal/identity/apperr` | `cryptoutilIdentityAppErr` | Already correct |
| `cryptoutil/internal/identity/config` | `cryptoutilIdentityConfig` | Already correct |
| `cryptoutil/internal/identity/domain` | `cryptoutilIdentityDomain` | Already correct |
| `cryptoutil/internal/identity/magic` | `cryptoutilIdentityMagic` | Already correct |
| `cryptoutil/internal/identity/repository` | `cryptoutilIdentityRepository` | Already correct |
| `cryptoutil/internal/identity/issuer` | `cryptoutilIdentityIssuer` | Already correct |
| `cryptoutil/internal/identity/authz` | `cryptoutilIdentityAuthz` | Already correct |
| `cryptoutil/internal/identity/authz/clientauth` | `cryptoutilIdentityClientAuth` | Already correct |
| `cryptoutil/internal/identity/idp` | `cryptoutilIdentityIdp` | Already correct |
| `cryptoutil/internal/identity/idp/auth` | `cryptoutilIdentityAuth` | Already correct |
| `cryptoutil/internal/identity/server` | `cryptoutilIdentityServer` | Already correct |

### CICD Packages

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/internal/cmd/cicd/common` | `cryptoutilCmdCicdCommon` | Already correct |
| `cryptoutil/internal/cmd/cicd/all_enforce_utf8` | `cryptoutilCmdCicdAllEnforceUtf8` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_check_circular_package_dependencies` | `cryptoutilCmdCicdGoCheckCircularPackageDependencies` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_check_identity_imports` | `cryptoutilCmdCicdGoCheckIdentityImports` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_enforce_any` | `cryptoutilCmdCicdGoEnforceAny` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_enforce_test_patterns` | `cryptoutilCmdCicdGoEnforceTestPatterns` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_fix_all` | `cryptoutilCmdCicdGoFixAll` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_fix_copyloopvar` | `cryptoutilCmdCicdGoFixCopyLoopVar` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_fix_staticcheck_error_strings` | `cryptoutilCmdCicdGoFixStaticcheckErrorStrings` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_fix_thelper` | `cryptoutilCmdCicdGoFixTHelper` | Already correct |
| `cryptoutil/internal/cmd/cicd/go_update_direct_dependencies` | `cryptoutilCmdCicdGoUpdateDirectDependencies` | Already correct |
| `cryptoutil/internal/cmd/cicd/github_workflow_lint` | `cryptoutilCmdCicdGithubWorkflowLint` | Already correct |

---

## Proposed Alias Map (Post-Refactor)

### KMS Domain

| Import Path | Alias | Migration From |
|-------------|-------|----------------|
| `cryptoutil/api/kms/client` | `cryptoutilKmsClient` | `cryptoutilOpenapiClient` |
| `cryptoutil/api/kms/model` | `cryptoutilKmsModel` | `cryptoutilOpenapiModel` |
| `cryptoutil/api/kms/server` | `cryptoutilKmsServer` | `cryptoutilOpenapiServer` |
| `cryptoutil/internal/kms/server` | `cryptoutilKmsServerInternal` | `cryptoutilServerApplication` |
| `cryptoutil/internal/kms/server/handler` | `cryptoutilKmsHandler` | `cryptoutilOpenapiHandler` |
| `cryptoutil/internal/kms/businesslogic` | `cryptoutilKmsBusinesslogic` | `cryptoutilBusinessLogic` |
| `cryptoutil/internal/kms/businesslogic/barrier` | `cryptoutilKmsBarrier` | `cryptoutilBarrierService` |
| `cryptoutil/internal/kms/businesslogic/barrier/contentkeysservice` | `cryptoutilKmsBarrierContentKeys` | `cryptoutilContentKeysService` |
| `cryptoutil/internal/kms/businesslogic/barrier/intermediatekeysservice` | `cryptoutilKmsBarrierIntermediateKeys` | `cryptoutilIntermediateKeysService` |
| `cryptoutil/internal/kms/businesslogic/barrier/rootkeysservice` | `cryptoutilKmsBarrierRootKeys` | `cryptoutilRootKeysService` |
| `cryptoutil/internal/kms/businesslogic/barrier/unsealkeysservice` | `cryptoutilKmsBarrierUnsealKeys` | `cryptoutilUnsealKeysService` |
| `cryptoutil/internal/kms/repository/orm` | `cryptoutilKmsOrmRepository` | `cryptoutilOrmRepository` |
| `cryptoutil/internal/kms/repository/sqlrepository` | `cryptoutilKmsSQLRepository` | `cryptoutilSQLRepository` |
| `cryptoutil/internal/kms/crypto/jose` | `cryptoutilKmsJose` | `cryptoutilJose` |
| `cryptoutil/internal/kms/pool` | `cryptoutilKmsPool` | `cryptoutilPool` |
| `cryptoutil/internal/kms/telemetry` | `cryptoutilKmsTelemetry` | `cryptoutilTelemetry` |
| `cryptoutil/internal/kms/container` | `cryptoutilKmsContainer` | `cryptoutilContainer` |
| `cryptoutil/internal/kms/config` | `cryptoutilKmsConfig` | NEW (extracted from server/config) |
| `cryptoutil/internal/kms/client` | `cryptoutilKmsClientInternal` | NEW (from internal/client) |

### General-Purpose Crypto (Promoted to pkg/)

| Import Path | Alias | Migration From |
|-------------|-------|----------------|
| `cryptoutil/pkg/crypto/keygen` | `cryptoutilKeygen` | `cryptoutilKeyGen` |
| `cryptoutil/pkg/crypto/digests` | `cryptoutilDigests` | `cryptoutilDigests` (path change only) |
| `cryptoutil/pkg/crypto/asn1` | `cryptoutilAsn1` | `cryptoutilAsn1` (path change only) |
| `cryptoutil/pkg/crypto/certificate` | `cryptoutilCertificate` | `cryptoutilCertificate` (path change only) |

**Rationale:** These packages are general-purpose cryptographic primitives usable across all service groups (KMS, Identity, CA, future services)

### CA Domain (Future)

| Import Path | Alias | Notes |
|-------------|-------|-------|
| `cryptoutil/api/ca/client` | `cryptoutilCaClient` | NEW |
| `cryptoutil/api/ca/model` | `cryptoutilCaModel` | NEW |
| `cryptoutil/api/ca/server` | `cryptoutilCaServer` | NEW |
| `cryptoutil/internal/ca/server` | `cryptoutilCaServerInternal` | NEW |
| `cryptoutil/internal/ca/businesslogic` | `cryptoutilCaBusinesslogic` | NEW |
| `cryptoutil/internal/ca/businesslogic/issuance` | `cryptoutilCaIssuance` | NEW |
| `cryptoutil/internal/ca/businesslogic/revocation` | `cryptoutilCaRevocation` | NEW |
| `cryptoutil/internal/ca/repository` | `cryptoutilCaRepository` | NEW |
| `cryptoutil/internal/ca/config` | `cryptoutilCaConfig` | NEW |

---

## Migration Strategy

### Phase 1: Add New Aliases (Pre-Migration)

Update `.golangci.yml` importas section with new aliases **BEFORE** moving packages:

```yaml
importas:
  alias:
    # KMS domain (new aliases)
    - pkg: cryptoutil/internal/kms/server
      alias: cryptoutilKmsServerInternal
    - pkg: cryptoutil/internal/kms/businesslogic
      alias: cryptoutilKmsBusinesslogic
    # ... (all KMS aliases)

    # Legacy aliases (deprecated, will remove after migration)
    - pkg: cryptoutil/internal/server/application
      alias: cryptoutilServerApplication
    - pkg: cryptoutil/internal/server/businesslogic
      alias: cryptoutilBusinessLogic
    # ... (all legacy aliases)
```

### Phase 2: Compatibility Period (During Migration)

**Both old and new aliases active simultaneously** for 8-week grace period:

```yaml
importas:
  alias:
    # NEW (post-refactor paths)
    - pkg: cryptoutil/internal/kms/businesslogic/barrier
      alias: cryptoutilKmsBarrier

    # OLD (pre-refactor paths, deprecated)
    - pkg: cryptoutil/internal/server/barrier
      alias: cryptoutilBarrierService  # Deprecated
```

**Allows:**
- Gradual code migration without breaking builds
- CI/CD continues passing during transition
- Test suite remains green throughout

### Phase 3: Remove Legacy Aliases (Post-Migration)

After all code migrated and compatibility shims removed:

```yaml
importas:
  alias:
    # KMS domain (only new aliases remain)
    - pkg: cryptoutil/internal/kms/businesslogic/barrier
      alias: cryptoutilKmsBarrier

    # Legacy aliases REMOVED (breaking change)
```

---

## Validation Tests

### Alias Enforcement Test

Create `internal/cmd/cicd/go_check_import_aliases/` to validate import alias compliance:

```go
// Verify all cryptoutil imports use correct aliases
func TestImportAliasCompliance(t *testing.T) {
    t.Parallel()

    goFiles := findAllGoFiles(".")
    for _, file := range goFiles {
        aliases := extractImportAliases(file)
        for pkg, alias := range aliases {
            if strings.HasPrefix(pkg, "cryptoutil/") {
                // MUST use cryptoutil prefix
                if !strings.HasPrefix(alias, "cryptoutil") {
                    t.Errorf("%s: package %s uses invalid alias %s (must start with 'cryptoutil')", file, pkg, alias)
                }

                // Check against approved alias map
                expected := getExpectedAlias(pkg)
                if alias != expected {
                    t.Errorf("%s: package %s uses alias %s, expected %s", file, pkg, alias, expected)
                }
            }
        }
    }
}
```

### Migration Progress Tracking

```go
// Track migration progress by counting legacy vs new imports
func TestMigrationProgress(t *testing.T) {
    t.Parallel()

    legacyImports := countImportsMatching("cryptoutil/internal/server/")
    newKmsImports := countImportsMatching("cryptoutil/internal/kms/")

    progress := float64(newKmsImports) / float64(legacyImports+newKmsImports) * 100
    t.Logf("KMS migration progress: %.1f%% (%d new / %d total)", progress, newKmsImports, legacyImports+newKmsImports)

    if legacyImports > 0 && time.Now().After(migrationDeadline) {
        t.Errorf("Migration incomplete: %d legacy imports remain after deadline", legacyImports)
    }
}
```

---

## golangci.yml Updates

### Current State (85 aliases)

See `.golangci.yml` lines 153-326 for current importas configuration

**Coverage:**
- ✅ JOSE libraries (4 aliases)
- ✅ Standard library (3 aliases)
- ✅ Third-party (2 aliases)
- ✅ Cryptoutil API (3 aliases)
- ✅ Cryptoutil server (12 aliases)
- ✅ Cryptoutil common (14 aliases)
- ✅ Cryptoutil crypto (5 aliases)
- ✅ Cryptoutil identity (11 aliases)
- ✅ Cryptoutil CICD (12 aliases)
- ✅ Stdlib crypto (7 aliases)

### Required Additions (Post-Refactor)

**KMS domain:** +17 new aliases
**pkg/crypto:** +4 aliases (path changes)
**CA domain:** +9 new aliases (future)

**Total post-refactor:** 85 current + 30 new = **115 aliases**

### Implementation Plan

1. **Task 4 (Current):** Document proposed aliases in this file ✅
2. **Task 11 (Phase 2):** Add new KMS aliases to `.golangci.yml` before package moves
3. **Task 11 (Phase 2):** Keep legacy aliases during 8-week migration
4. **Task 11 (Phase 2):** Update import statements in moved packages
5. **Task 11 (Phase 3):** Remove legacy aliases after compatibility shims removed
6. **Task 12 (Phase 3):** Add CA aliases when CA implementation begins

---

## Cross-References

- **Service Groups Taxonomy:** [docs/01-refactor/service-groups.md](./service-groups.md)
- **Dependency Analysis:** [docs/01-refactor/dependency-analysis.md](./dependency-analysis.md)
- **Group Directory Blueprint:** [docs/01-refactor/blueprint.md](./blueprint.md)
- **golangci.yml:** [.golangci.yml](../../.golangci.yml) - importas configuration
- **Golang Instructions:** [.github/instructions/01-03.golang.instructions.md](../../.github/instructions/01-03.golang.instructions.md)

---

## Notes

- **Alias consistency:** All cryptoutil packages MUST use `cryptoutil` prefix (enforced by golangci-lint importas)
- **CamelCase convention:** Package names converted to camelCase in aliases (e.g., `businesslogic` → `Businesslogic`)
- **Service group prefix:** New structure uses service group in alias (e.g., `cryptoutilKms`, `cryptoutilIdentity`, `cryptoutilCa`)
- **Crypto acronyms:** Use ALL CAPS for RSA, EC, ECDSA, etc. (e.g., `cryptoutilRSAKeys`, `cryptoutilECDSASign`)
- **Backward compatibility:** Legacy aliases remain active during 8-week migration window
- **Linter enforcement:** importas linter fails builds on alias violations (strict enforcement)
