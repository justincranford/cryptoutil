# Identity Service Extraction Plan

## Executive Summary

Extract `internal/identity` to a standalone service module, enabling independent development, deployment, and versioning of the OAuth 2.1/OIDC identity platform.

**Status**: Planning
**Dependencies**: Tasks 1-9 (taxonomy, dependencies, blueprint, aliases, CLI, utilities, pipeline, tooling, documentation)
**Risk Level**: Medium (119 Go files, 15+ external consumers, complex dependency graph)

## Current State Analysis

### Package Structure

```
internal/identity/
├── apperr/             # Identity-specific error codes (1 file)
├── authz/              # OAuth 2.1 authorization server (16 files)
│   ├── clientauth/     # Client authentication methods (11 files)
│   └── pkce/           # PKCE implementation (3 files)
├── config/             # Identity configuration (5 files)
├── domain/             # Domain models (10 files)
├── idp/                # Identity provider (18 files)
│   ├── auth/           # Authentication methods (7 files)
│   └── userauth/       # User authentication strategies (11 files)
├── integration/        # Integration tests (1 file)
├── issuer/             # JWT/JWE token issuance (7 files)
├── jobs/               # Cleanup jobs (2 files)
├── magic/              # Identity-specific constants (3 files)
├── repository/         # Data access layer (12 files)
│   └── orm/            # GORM repositories (8 files)
├── rs/                 # Resource server (3 files)
├── security/           # Security policies (2 files)
├── server/             # HTTP servers (4 files)
├── storage/            # Storage tests (7 files)
│   ├── fixtures/       # Test fixtures (2 files)
│   └── tests/          # Storage layer tests (3 files)
└── test/               # Test utilities (7 files)
    ├── e2e/            # E2E tests (3 files)
    ├── integration/    # Integration tests (1 file)
    └── testutils/      # Test helpers (1 file)

Total: 119 Go files across 25 subdirectories
```

### External Consumers (Import Dependencies)

**From `cmd/` directory**:

- `cmd/identity/authz/main.go` - AuthZ server entry point (5 imports)
- `cmd/identity/idp/main.go` - IdP server entry point (5 imports)
- `cmd/identity/rs/main.go` - RS server entry point (5 imports)
- `cmd/identity/spa-rp/main.go` - SPA relying party (1 import)

**Internal cross-module references**: None (identity is fully isolated from KMS)

### Importas Aliases (Current)

From `.golangci.yml` lines 289-309:

```yaml
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
- pkg: cryptoutil/internal/identity/issuer
  alias: cryptoutilIdentityIssuer
- pkg: cryptoutil/internal/identity/authz
  alias: cryptoutilIdentityAuthz
- pkg: cryptoutil/internal/identity/authz/clientauth
  alias: cryptoutilIdentityClientAuth
- pkg: cryptoutil/internal/identity/idp
  alias: cryptoutilIdentityIdp
- pkg: cryptoutil/internal/identity/idp/auth
  alias: cryptoutilIdentityAuth
- pkg: cryptoutil/internal/identity/server
  alias: cryptoutilIdentityServer
- pkg: cryptoutil/internal/identity/test/testutils
  alias: cryptoutilIdentityTestTestutils
```

Total: 13 identity-specific importas rules

## Extraction Options

### Option 1: Keep as Internal Package (Status Quo)

**Path**: `cryptoutil/internal/identity` → NO CHANGE

**Pros**:

- No migration work required
- No import path changes
- Existing importas rules remain valid
- Consistent with current architecture (KMS also in `internal/`)

**Cons**:

- Cannot be imported by external projects (Go enforces `internal/` visibility)
- No independent versioning (tied to cryptoutil releases)
- Cannot publish as standalone Go module
- Confusing for users expecting public OAuth 2.1 library

**Recommendation**: ❌ **NOT RECOMMENDED** for long-term; prevents identity platform adoption outside cryptoutil

### Option 2: Promote to `pkg/identity` (Public Library)

**Path**: `cryptoutil/internal/identity` → `cryptoutil/pkg/identity`

**Pros**:

- Public Go package (can be imported by external projects)
- Still part of cryptoutil monorepo
- Independent versioning possible via Go modules replace directives
- Clear public API surface (`pkg/` convention)
- Minimal import path changes

**Cons**:

- Requires updating ALL 119 files with import path changes
- Need to audit API surface for stability (breaking changes impact external users)
- `pkg/` pattern less common in modern Go (internal-first preferred)
- Still coupled to cryptoutil release cycle

**Recommendation**: ⚠️ **CONDITIONAL** - use only if external library consumption is critical

### Option 3: Extract to Separate Module (Go Workspace)

**Path**: `cryptoutil/internal/identity` → `identity/` (sibling to `cryptoutil/`)

**Module structure**:

```
/
├── cryptoutil/          # KMS module (existing)
│   ├── go.mod           # module cryptoutil
│   └── internal/
│       └── kms/         # Renamed from server/
├── identity/            # Identity module (NEW)
│   ├── go.mod           # module identity
│   └── internal/
│       ├── authz/
│       ├── idp/
│       └── rs/
└── go.work              # Go workspace file
```

**Pros**:

- True module independence (separate go.mod, versioning, releases)
- Clear separation of concerns (OAuth 2.1 ≠ KMS)
- Independent CI/CD pipelines possible
- Can publish to separate GitHub repo later
- Go workspace provides monorepo benefits (shared tooling, atomic commits)

**Cons**:

- Highest migration complexity (new module, workspace setup)
- Requires updating 119 files + build scripts + workflows
- Learning curve for Go workspace pattern
- More complex dependency management

**Recommendation**: ⭐ **RECOMMENDED** for scalability and architectural clarity

## Recommended Approach: Option 3 (Go Workspace)

### Rationale

1. **Independent Lifecycle**: OAuth 2.1/OIDC identity platform has different:
   - Feature velocity (identity standards evolve faster than KMS)
   - Security patching requirements
   - Compliance requirements (OAuth 2.1, FAPI, OIDC standards)
   - Consumer base (may want standalone identity without KMS)

2. **Clear Boundaries**: Identity and KMS are fundamentally different services:
   - Identity: User authentication, authorization, token issuance
   - KMS: Cryptographic key management, certificates, signing operations
   - Shared dependency: Only through common crypto primitives (no business logic overlap)

3. **Future-Proof**: Enables:
   - Separate GitHub repository later (`github.com/justincranford/identity`)
   - Independent versioning (identity v2.x, kms v3.x)
   - Standalone deployment architectures
   - External adoption without KMS dependency

### Implementation Phases

#### Phase 1: Go Workspace Setup

**Create workspace structure**:

```bash
# Create go.work at repository root
cat > go.work <<EOF
go 1.25.4

use (
    ./cryptoutil
    ./identity
)
EOF

# Create identity module directory
mkdir -p identity
```

**Initialize identity module**:

```bash
cd identity
go mod init identity
go mod edit -go=1.25.4

# Copy LICENSE, README from cryptoutil
cp ../cryptoutil/LICENSE .
cat > README.md <<EOF
# Identity Service

OAuth 2.1 / OIDC identity platform extracted from cryptoutil.

## Features
- OAuth 2.1 authorization server (RFC 6749, RFC 9126)
- OpenID Connect provider (OIDC Core 1.0)
- Resource server with JWT validation
- Client authentication methods (RFC 8705, RFC 7523, RFC 7521)
- PKCE support (RFC 7636)
EOF
```

#### Phase 2: Code Migration

**Move files**:

```bash
# Move internal/identity → identity/internal/
mv cryptoutil/internal/identity identity/internal/

# Update directory structure
cd identity/internal
mkdir -p identity  # Preserve package name
mv * identity/     # Nest under identity/ to avoid import conflicts
```

**Updated structure**:

```
identity/
├── go.mod
├── go.sum
├── LICENSE
├── README.md
└── internal/
    └── identity/
        ├── apperr/
        ├── authz/
        ├── config/
        ├── domain/
        ├── idp/
        ├── issuer/
        ├── jobs/
        ├── magic/
        ├── repository/
        ├── rs/
        ├── security/
        ├── server/
        ├── storage/
        └── test/
```

#### Phase 3: Import Path Updates

**Search and replace** across ALL files (119 identity files + 4 cmd files + tests):

```bash
# Find all files importing identity packages
find cryptoutil identity -name "*.go" -type f | xargs grep -l "cryptoutil/internal/identity"

# Replace import paths
find cryptoutil identity -name "*.go" -type f -exec sed -i \
  's|cryptoutil/internal/identity|identity/internal/identity|g' {} +
```

**Example change**:

```go
// BEFORE (cryptoutil imports)
import (
    cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
    cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

// AFTER (identity module imports)
import (
    identityAuthz "identity/internal/identity/authz"
    identityConfig "identity/internal/identity/config"
)
```

**Manual verification required**:

- Check cmd/identity/* entry points
- Check test files for relative imports
- Verify go.mod replace directives if needed

#### Phase 4: Importas Rule Updates

**Update `.golangci.yml`** (both modules):

```yaml
# cryptoutil/.golangci.yml - REMOVE identity rules (no longer importing identity)
# DELETE lines 289-309 (13 identity-specific rules)

# identity/.golangci.yml - CREATE new file with identity rules
linters-settings:
  importas:
    no-unaliased: true
    no-extra-aliases: false
    alias:
      # Identity internal packages (now using identity module prefix)
      - pkg: identity/internal/identity/apperr
        alias: identityAppErr
      - pkg: identity/internal/identity/config
        alias: identityConfig
      - pkg: identity/internal/identity/domain
        alias: identityDomain
      - pkg: identity/internal/identity/magic
        alias: identityMagic
      - pkg: identity/internal/identity/repository
        alias: identityRepository
      - pkg: identity/internal/identity/issuer
        alias: identityIssuer
      - pkg: identity/internal/identity/authz
        alias: identityAuthz
      - pkg: identity/internal/identity/authz/clientauth
        alias: identityClientAuth
      - pkg: identity/internal/identity/idp
        alias: identityIdp
      - pkg: identity/internal/identity/idp/auth
        alias: identityAuth
      - pkg: identity/internal/identity/server
        alias: identityServer
      - pkg: identity/internal/identity/repository/orm
        alias: identityORM
      - pkg: identity/internal/identity/authz/pkce
        alias: identityPKCE
      - pkg: identity/internal/identity/rs
        alias: identityRS
      - pkg: identity/internal/identity/security
        alias: identitySecurity
      - pkg: identity/internal/identity/jobs
        alias: identityJobs
      - pkg: identity/internal/identity/test/testutils
        alias: identityTestTestutils
      # Third-party packages (inherited from cryptoutil)
      - pkg: github.com/google/uuid
        alias: googleUuid
      - pkg: github.com/gofiber/fiber/v2
        alias: fiber
      - pkg: gorm.io/gorm
        alias: gorm
      # ... (copy other common third-party aliases)
```

**Alias prefix change**: `cryptoutilIdentity*` → `identity*` (cleaner, module-specific)

#### Phase 5: CI/CD Workflow Updates

**Update workflow path filters**:

```yaml
# .github/workflows/ci-quality.yml
on:
  push:
    paths:
      - 'cryptoutil/**/*.go'
      - 'identity/**/*.go'  # ADD identity paths
      - 'cryptoutil/go.mod'
      - 'identity/go.mod'    # ADD identity go.mod
      - 'go.work'            # ADD workspace file

# .github/workflows/ci-coverage.yml
jobs:
  coverage:
    strategy:
      matrix:
        module: [cryptoutil, identity]  # ADD identity module
    steps:
      - name: Run tests
        run: |
          cd ${{ matrix.module }}
          go test ./... -coverprofile=coverage.out
```

**Create identity-specific workflows** (optional):

```yaml
# .github/workflows/identity-quality.yml
name: Identity Quality
on:
  push:
    paths:
      - 'identity/**'
      - 'go.work'
jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      - name: Lint identity module
        run: |
          cd identity
          golangci-lint run --timeout=10m
      - name: Test identity module
        run: |
          cd identity
          go test ./... -cover
```

**Update composite actions** if they reference identity paths

#### Phase 6: VS Code Configuration

**Update `.vscode/settings.json`**:

```json
{
  "gopls": {
    "build.directoryFilters": [
      "-cryptoutil/vendor",
      "-identity/vendor"
    ],
    "formatting.gofumpt": true
  },
  "go.useLanguageServer": true,
  "go.toolsManagement.autoUpdate": true
}
```

**Update launch configurations** (`cryptoutil/.vscode/launch.json`):

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Identity AuthZ Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cryptoutil/cmd/identity/authz",
      "cwd": "${workspaceFolder}/identity",
      "env": {
        "CONFIG_PATH": "${workspaceFolder}/identity/configs/development.yml"
      }
    },
    {
      "name": "Identity IdP Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cryptoutil/cmd/identity/idp",
      "cwd": "${workspaceFolder}/identity"
    }
  ]
}
```

#### Phase 7: Docker & Compose Updates

**Update Dockerfile** (if identity has separate image):

```dockerfile
# identity/Dockerfile
FROM golang:1.25.4 AS builder
WORKDIR /src
COPY go.work go.work
COPY cryptoutil/go.mod cryptoutil/go.sum cryptoutil/
COPY identity/go.mod identity/go.sum identity/
RUN cd cryptoutil && go mod download
RUN cd identity && go mod download
COPY cryptoutil cryptoutil/
COPY identity identity/
WORKDIR /src/identity
RUN CGO_ENABLED=0 go build -o /app/identity ./cmd/identity
```

**Update docker-compose.yml**:

```yaml
services:
  identity-authz:
    build:
      context: .
      dockerfile: identity/Dockerfile
    volumes:
      - ./identity/configs:/configs:ro
    environment:
      - CONFIG_PATH=/configs/development.yml
```

#### Phase 8: Documentation Updates

**Update root README.md**:

```markdown
# cryptoutil

Multi-service cryptographic platform consisting of:

## Services

### KMS (Key Management Service)
Location: `cryptoutil/`
Module: `cryptoutil`

### Identity (OAuth 2.1 / OIDC)
Location: `identity/`
Module: `identity`

## Workspace Structure

This is a Go workspace containing multiple modules:
- `cryptoutil/` - KMS service
- `identity/` - Identity service
- `go.work` - Workspace configuration

Build instructions: See each module's README.md
```

**Create identity/README.md** (detailed service documentation)

**Update docs/README.md** with workspace patterns

#### Phase 9: Testing & Validation

**Test commands**:

```bash
# Build both modules
go work sync
go build ./cryptoutil/cmd/cryptoutil
go build ./cryptoutil/cmd/identity/authz

# Run tests
cd cryptoutil && go test ./... && cd ..
cd identity && go test ./... && cd ..

# Verify no import errors
go list -m all

# Check coverage (each module independently)
cd cryptoutil && go test ./... -coverprofile=coverage.out && cd ..
cd identity && go test ./... -coverprofile=coverage.out && cd ..
```

**Validation checklist**:

- [ ] All imports resolve correctly (no "cannot find package" errors)
- [ ] All tests pass (`go test ./...` in both modules)
- [ ] golangci-lint passes (both modules)
- [ ] Coverage unchanged (compare before/after)
- [ ] Workflow runs succeed (quality, coverage, benchmarks)
- [ ] Docker builds succeed
- [ ] Launch configurations work in VS Code

#### Phase 10: Rollback Plan

**If extraction fails**, revert with:

```bash
# Delete workspace
rm go.work

# Restore identity to cryptoutil
rm -rf cryptoutil/internal/identity
git checkout HEAD -- cryptoutil/internal/identity

# Revert import path changes
find cryptoutil -name "*.go" -type f -exec sed -i \
  's|identity/internal/identity|cryptoutil/internal/identity|g' {} +

# Restore .golangci.yml
git checkout HEAD -- cryptoutil/.golangci.yml
```

## Risk Assessment

### High Risks

1. **Import Path Updates (119 files)**
   - Mitigation: Automated sed scripts + manual review + test suite validation
   - Rollback: git checkout + sed reversal

2. **Workflow Path Filters**
   - Mitigation: Test with act before pushing
   - Rollback: Revert workflow YAML changes

3. **Importas Rule Synchronization**
   - Mitigation: golangci-lint run before commit
   - Rollback: Restore old .golangci.yml

### Medium Risks

1. **Go Workspace Learning Curve**
   - Mitigation: Document workspace commands in README
   - Fallback: Keep workspace simple (just 2 modules)

2. **Docker Build Complexity**
   - Mitigation: Multi-stage build with proper COPY order
   - Fallback: Revert to single-module Dockerfile

### Low Risks

1. **Coverage Regression**
   - Mitigation: Compare coverage before/after, fix any gaps
   - Acceptable: Workspace overhead may slightly reduce coverage metrics

2. **VS Code Configuration**
   - Mitigation: Test launch configs after migration
   - Fallback: Minimal changes to settings.json

## Success Metrics

- [ ] All 119 identity files migrated to `identity/internal/identity/`
- [ ] All imports updated (`cryptoutil/internal/identity` → `identity/internal/identity`)
- [ ] Zero golangci-lint errors in both modules
- [ ] Test coverage ≥ baseline (identity: 85%, cryptoutil: 80%)
- [ ] All workflows pass (quality, coverage, e2e, dast, load)
- [ ] Docker Compose stack starts successfully
- [ ] VS Code debugging works for both modules

## Timeline

- **Phase 1-2**: Go workspace + code migration (1 day)
- **Phase 3-4**: Import paths + importas rules (1 day)
- **Phase 5-6**: CI/CD + VS Code (1 day)
- **Phase 7-8**: Docker + documentation (1 day)
- **Phase 9**: Testing & validation (1 day)
- **Phase 10**: Rollback plan documentation (buffer)

**Total**: 5-7 days (including buffer for issues)

## Cross-References

- [Service Group Taxonomy](service-groups.md) - Identity service group definition
- [Repository Inventory](dependencies.md) - Identity dependency analysis
- [Directory Blueprint](blueprint.md) - Workspace structure patterns
- [Import Alias Policy](import-aliases.md) - Importas rule patterns
- [Pipeline Impact Assessment](pipeline-impact.md) - CI/CD workflow changes
- [Workspace Tooling Alignment](tooling.md) - VS Code configuration

## Next Steps

After identity extraction:

1. **Task 11**: KMS extraction (internal/server → internal/kms)
2. **Task 12**: CA preparation (internal/ca skeleton)
3. **Task 13-15**: CLI restructuring (service group commands)
