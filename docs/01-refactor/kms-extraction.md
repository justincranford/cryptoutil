# KMS Service Extraction Plan

## Executive Summary

Extract `internal/server` → `internal/kms` and consolidate KMS-specific code into a cohesive module structure, separating it from identity and shared utilities.

**Status**: Planning
**Dependencies**: Task 10 (identity extraction complete)
**Risk Level**: High (100+ files, 15+ workflows, complex dependencies on pool/jose/telemetry)

## Current State Analysis

### Package Structure

```
internal/server/                        # RENAME to internal/kms/
├── application/          # 4 files  # → internal/kms/application/
│   ├── application_basic.go            # Basic server lifecycle
│   ├── application_core.go             # Core server functions
│   ├── application_init.go             # Initialization logic
│   └── application_listener.go         # HTTP listener setup
├── barrier/              # 4 subdirs  # → internal/kms/barrier/
│   ├── contentkeysservice/             # Content key management
│   ├── intermediatekeysservice/        # Intermediate key management
│   ├── rootkeysservice/                # Root key management (1 file)
│   └── unsealkeysservice/              # Unseal key management
├── businesslogic/        # 6 files  # → internal/kms/businesslogic/
│   ├── businesslogic.go                # Business logic layer
│   ├── businesslogic_test.go
│   ├── elastic_key_status_state_machine.go  # State machine
│   ├── elastic_key_status_state_machine_test.go
│   ├── oam_orm_mapper.go               # OAM ↔ ORM mappings
│   └── oam_orm_mapper_test.go
├── handler/              # 3 files  # → internal/kms/handler/
│   ├── handler_test.go                 # Handler tests
│   ├── oam_oas_mapper.go               # OAM ↔ OAS mappings
│   └── oas_handlers.go                 # OpenAPI handlers
└── repository/           # 2 subdirs # → internal/kms/repository/
    ├── orm/              # 50+ files  # ORM repository (GORM)
    └── sqlrepository/    # 26 files   # SQL repository (database/sql)

internal/client/                        # RENAME to internal/kms/client/
├── client_oam_mapper.go                # OAM mappings for client
├── client_test.go                      # Client tests
└── client_test_util.go                 # Client test utilities

internal/common/                        # STAYS (shared by KMS + identity)
├── apperr/               # Application errors
├── config/               # Configuration parsing
├── container/            # Container utilities (KMS-specific? TBD)
├── crypto/               # Cryptographic primitives (shared)
├── magic/                # Magic constants (shared)
├── pool/                 # Goroutine pools (KMS-specific? TBD)
├── telemetry/            # OpenTelemetry setup (KMS-specific? TBD)
├── testutil/             # Test utilities (shared)
└── util/                 # General utilities (shared)
```

**Total**: ~100 Go files across 15 subdirectories

### External Consumers

**From `internal/cmd/cryptoutil/server.go`**:

```go
import (
    cryptoutilConfig "cryptoutil/internal/common/config"
    cryptoutilServerApplication "cryptoutil/internal/server/application"
)
```

**From `api/server/openapi_gen_server.go`** (OpenAPI generated code):

- Imports handler package for server interface implementation

**From workflows**:

- Path filters: `internal/server/**` in ci-quality.yml, ci-coverage.yml, ci-e2e.yml

### Shared Utilities Decision (KMS-Specific vs Shared)

#### KMS-Specific → Move to `internal/kms/`

1. **pool** (goroutine pool)
   - Used exclusively by KMS for concurrent key operations
   - Identity has simpler concurrency needs (no pool required)
   - **Decision**: Move to `internal/kms/pool/`

2. **jose** (JWE/JWS operations)
   - Heavy usage in KMS barrier services (key encryption)
   - Identity uses separate issuer package for tokens (different patterns)
   - **Decision**: Move to `internal/kms/jose/` (or keep as `internal/kms/crypto/jose`)

3. **container** (Docker container utilities)
   - Only used in KMS tests for PostgreSQL container setup
   - Identity tests use different patterns
   - **Decision**: Move to `internal/kms/container/`

4. **telemetry** (OpenTelemetry)
   - KMS-specific OTLP setup and metrics
   - Identity would need separate telemetry config
   - **Decision**: **KEEP IN `internal/common/telemetry/`** (shared pattern for future CA service)

#### Shared → Keep in `internal/common/`

1. **apperr** - Application error codes (used by KMS, identity, and future CA)
2. **config** - Configuration parsing (YAML, CLI flags)
3. **crypto** - Crypto primitives (keygen, digests, asn1, certificate)
4. **magic** - Magic constants (shared across services)
5. **testutil** - Test utilities (shared test patterns)
6. **util** - General utilities (datetime, files, network, sysinfo, combinations)

### Importas Aliases (Current)

From `.golangci.yml` lines 269-285:

```yaml
# Cryptoutil internal - server
- pkg: cryptoutil/internal/server/application
  alias: cryptoutilServerApplication
- pkg: cryptoutil/internal/server/businesslogic
  alias: cryptoutilBusinessLogic
- pkg: cryptoutil/internal/server/handler
  alias: cryptoutilOpenapiHandler
- pkg: cryptoutil/internal/server/barrier
  alias: cryptoutilBarrierService
- pkg: cryptoutil/internal/server/barrier/contentkeysservice
  alias: cryptoutilContentKeysService
- pkg: cryptoutil/internal/server/barrier/intermediatekeysservice
  alias: cryptoutilIntermediateKeysService
- pkg: cryptoutil/internal/server/barrier/rootkeysservice
  alias: cryptoutilRootKeysService
- pkg: cryptoutil/internal/server/barrier/unsealkeysservice
  alias: cryptoutilUnsealKeysService
- pkg: cryptoutil/internal/server/repository/orm
  alias: cryptoutilOrmRepository
- pkg: cryptoutil/internal/server/repository/sqlrepository
  alias: cryptoutilSQLRepository
```

**Total**: 11 KMS-specific importas rules

## Extraction Strategy

### Rename Plan

```bash
# Core KMS packages
internal/server/              → internal/kms/
internal/client/              → internal/kms/client/

# KMS-specific utilities (move from internal/common/)
internal/common/pool/         → internal/kms/pool/
internal/common/container/    → internal/kms/container/

# Shared crypto (keep in common, but consider renaming for clarity)
internal/common/crypto/jose/  → internal/kms/crypto/jose/  (OR keep in common)

# Shared telemetry (KEEP in common)
internal/common/telemetry/    → STAYS (shared pattern for multi-service telemetry)
```

### Final Structure

```
internal/
├── kms/                      # NEW (renamed from server/)
│   ├── application/          # Server lifecycle
│   ├── barrier/              # Key barrier services
│   ├── businesslogic/        # Business logic layer
│   ├── client/               # KMS client (renamed from internal/client/)
│   ├── container/            # Docker test utilities (moved from common/)
│   ├── crypto/               # KMS-specific crypto (jose, pool-related crypto)
│   ├── handler/              # OpenAPI handlers
│   ├── pool/                 # Goroutine pool (moved from common/)
│   └── repository/           # Data access layer (orm, sqlrepository)
├── identity/                 # Identity service (from Task 10)
│   └── ...
├── ca/                       # CA service (Task 12 - skeleton only)
│   └── ...
└── common/                   # Shared utilities
    ├── apperr/               # Application errors (SHARED)
    ├── config/               # Configuration (SHARED)
    ├── crypto/               # Crypto primitives (SHARED) - keygen, digests, asn1, certificate
    ├── magic/                # Magic constants (SHARED)
    ├── telemetry/            # OpenTelemetry (SHARED)
    ├── testutil/             # Test utilities (SHARED)
    └── util/                 # General utilities (SHARED)
```

## Implementation Phases

### Phase 1: Rename Core Packages

**Move server → kms**:

```bash
# Rename directory
git mv internal/server internal/kms

# Update package declarations (automated)
find internal/kms -name "*.go" -type f -exec sed -i 's/^package server$/package kms/g' {} +
find internal/kms -name "*.go" -type f -exec sed -i 's/^package server_test$/package kms_test/g' {} +
```

**Move client → kms/client**:

```bash
# Rename directory
git mv internal/client internal/kms/client

# Package declarations already use "client" (no change needed)
```

### Phase 2: Move KMS-Specific Utilities

**Move pool**:

```bash
git mv internal/common/pool internal/kms/pool

# Update imports in kms package
find internal/kms -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/common/pool|cryptoutil/internal/kms/pool|g' {} +
```

**Move container**:

```bash
git mv internal/common/container internal/kms/container

# Update imports in kms package
find internal/kms -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/common/container|cryptoutil/internal/kms/container|g' {} +
```

**Decision on jose**:

- **Option A**: Move to `internal/kms/crypto/jose/` (KMS-specific crypto)
- **Option B**: Keep in `internal/common/crypto/jose/` (shared crypto utility)
- **Recommendation**: **KEEP IN COMMON** (jose is general-purpose JWE/JWS, identity may use later)

### Phase 3: Import Path Updates

**Search and replace** across ALL files:

```bash
# Update imports: internal/server → internal/kms
find . -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/server|cryptoutil/internal/kms|g' {} +

# Update imports: internal/client → internal/kms/client
find . -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/client|cryptoutil/internal/kms/client|g' {} +

# Update pool imports
find . -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/common/pool|cryptoutil/internal/kms/pool|g' {} +

# Update container imports
find . -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/common/container|cryptoutil/internal/kms/container|g' {} +
```

**Files requiring manual review**:

- `internal/cmd/cryptoutil/server.go` - imports server/application
- `api/server/openapi_gen_server.go` - OpenAPI generated imports handler package
- Workflow path filters

### Phase 4: Importas Rule Updates

**Update `.golangci.yml`** (replace KMS aliases):

```yaml
# OLD (internal/server)
- pkg: cryptoutil/internal/server/application
  alias: cryptoutilServerApplication

# NEW (internal/kms)
- pkg: cryptoutil/internal/kms/application
  alias: cryptoutilKmsApplication

# Full replacement list:
- pkg: cryptoutil/internal/kms/application
  alias: cryptoutilKmsApplication
- pkg: cryptoutil/internal/kms/businesslogic
  alias: cryptoutilKmsBusinessLogic
- pkg: cryptoutil/internal/kms/handler
  alias: cryptoutilKmsHandler
- pkg: cryptoutil/internal/kms/barrier
  alias: cryptoutilKmsBarrier
- pkg: cryptoutil/internal/kms/barrier/contentkeysservice
  alias: cryptoutilKmsContentKeys
- pkg: cryptoutil/internal/kms/barrier/intermediatekeysservice
  alias: cryptoutilKmsIntermediateKeys
- pkg: cryptoutil/internal/kms/barrier/rootkeysservice
  alias: cryptoutilKmsRootKeys
- pkg: cryptoutil/internal/kms/barrier/unsealkeysservice
  alias: cryptoutilKmsUnsealKeys
- pkg: cryptoutil/internal/kms/repository/orm
  alias: cryptoutilKmsORM
- pkg: cryptoutil/internal/kms/repository/sqlrepository
  alias: cryptoutilKmsSQL
- pkg: cryptoutil/internal/kms/client
  alias: cryptoutilKmsClient
- pkg: cryptoutil/internal/kms/pool
  alias: cryptoutilKmsPool
- pkg: cryptoutil/internal/kms/container
  alias: cryptoutilKmsContainer
```

**Alias prefix change**: `cryptoutilServer*` → `cryptoutilKms*` (clearer, service-specific)

### Phase 5: Workflow Path Filter Updates

**Update `.github/workflows/`**:

```yaml
# OLD path filters
paths:
  - 'internal/server/**'
  - 'internal/client/**'

# NEW path filters
paths:
  - 'internal/kms/**'
```

**Workflows to update**:

- `ci-quality.yml`
- `ci-coverage.yml`
- `ci-e2e.yml`
- `ci-dast.yml`
- `ci-load.yml`
- `ci-race.yml`
- `ci-fuzz.yml`

**Composite actions** (if they reference server/client paths)

### Phase 6: OpenAPI Code Generation

**Update `api/openapi-gen_config_server.yaml`**:

```yaml
# OLD
package: handler
output: api/server/openapi_gen_server.go

# NEW (if package name changes)
package: handler  # KEEP (handler package name unchanged)
output: api/server/openapi_gen_server.go  # KEEP (output path unchanged)
```

**Regenerate server code**:

```bash
cd api
go generate ./...
```

**Verify**:

- Check generated import paths reference `cryptoutil/internal/kms/handler`

### Phase 7: VS Code Configuration

**Update `.vscode/settings.json`** (if needed):

```json
{
  "gopls": {
    "build.directoryFilters": [
      "-vendor",
      "+internal/kms",
      "+internal/identity",
      "+internal/common"
    ]
  }
}
```

**Update launch configurations**:

```json
{
  "name": "KMS Server",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}/cmd/cryptoutil",
  "args": ["server", "start", "--config", "${workspaceFolder}/configs/test/config.yml"]
}
```

### Phase 8: Docker & Compose Updates

**Update Dockerfile** (if paths referenced):

```dockerfile
# Build KMS server
COPY internal/kms internal/kms
COPY internal/common internal/common
COPY cmd/cryptoutil cmd/cryptoutil
RUN go build -o /app/cryptoutil ./cmd/cryptoutil
```

**Update compose.yml** (if paths referenced in comments/docs)

### Phase 9: Documentation Updates

**Update README.md**:

```markdown
## Project Structure

```

internal/
├── kms/           # Key Management Service (previously server/)
├── identity/      # OAuth 2.1 / OIDC Identity Platform
├── ca/            # Certificate Authority (future)
└── common/        # Shared utilities

```
```

**Update docs/README.md** with new structure

**Update PLANS-INDEX.md** to reflect refactor completion

### Phase 10: Testing & Validation

**Test commands**:

```bash
# Build
go build ./cmd/cryptoutil
go build ./cmd/identity/authz

# Tests
go test ./internal/kms/... -cover
go test ./internal/common/... -cover

# Lint
golangci-lint run --timeout=10m

# Coverage
go test ./... -coverprofile=test-output/coverage_kms.out
```

**Validation checklist**:

- [ ] All imports resolve (`go list -m all`)
- [ ] All tests pass (`go test ./...`)
- [ ] golangci-lint passes
- [ ] Coverage unchanged
- [ ] Workflows pass (quality, coverage, e2e)
- [ ] Docker builds succeed
- [ ] Server starts: `./cryptoutil server start --dev`

## Risk Assessment

### High Risks

1. **Import Path Updates (100+ files)**
   - Mitigation: Automated sed + manual review + test suite
   - Rollback: `git revert` + sed reversal

2. **OpenAPI Regeneration**
   - Mitigation: Test generated code before commit
   - Rollback: Revert api/server/ directory

3. **Workflow Path Filters**
   - Mitigation: Test with act
   - Rollback: Revert workflow YAML

### Medium Risks

1. **Pool/Container Dependency Analysis**
   - Mitigation: grep analysis confirms KMS-only usage
   - Fallback: Keep in `internal/common/` if shared usage found

2. **Importas Synchronization**
   - Mitigation: golangci-lint run before commit
   - Rollback: Restore old .golangci.yml

### Low Risks

1. **VS Code Configuration**
   - Mitigation: Minimal changes (gopls handles renames)

2. **Docker Builds**
   - Mitigation: Test build before commit

## Success Metrics

- [ ] `internal/server` renamed to `internal/kms`
- [ ] `internal/client` renamed to `internal/kms/client`
- [ ] `internal/common/pool` moved to `internal/kms/pool`
- [ ] `internal/common/container` moved to `internal/kms/container`
- [ ] All imports updated (`internal/server` → `internal/kms`)
- [ ] Importas rules updated (11 KMS aliases)
- [ ] Zero golangci-lint errors
- [ ] Test coverage ≥ baseline (85%+ for KMS infrastructure)
- [ ] All workflows pass
- [ ] OpenAPI server code regenerated successfully
- [ ] Server starts: `./cryptoutil server start --dev`

## Timeline

- **Phase 1**: Rename core packages (2 hours)
- **Phase 2**: Move utilities (2 hours)
- **Phase 3**: Import path updates (2 hours)
- **Phase 4**: Importas rules (1 hour)
- **Phase 5**: Workflow updates (1 hour)
- **Phase 6**: OpenAPI regeneration (1 hour)
- **Phase 7-8**: VS Code + Docker (1 hour)
- **Phase 9**: Documentation (2 hours)
- **Phase 10**: Testing & validation (2 hours)

**Total**: 14 hours (1-2 days)

## Cross-References

- [Task 10: Identity Extraction](identity-extraction.md) - Identity module separation
- [Service Group Taxonomy](service-groups.md) - KMS service group definition
- [Shared Utilities Extraction](shared-utilities.md) - Pool/container analysis
- [Import Alias Policy](import-aliases.md) - Importas rule patterns
- [Pipeline Impact Assessment](pipeline-impact.md) - Workflow updates
- [Directory Blueprint](blueprint.md) - Final structure vision

## Next Steps

After KMS extraction:

1. **Task 12**: CA structure preparation
2. **Task 13-15**: CLI restructuring (kms, identity, ca commands)
3. **Task 16-18**: Infrastructure updates (workflows, importas, telemetry)
