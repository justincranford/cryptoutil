# Group Directory Blueprint

## Overview

This document defines the target directory structure for cryptoutil's multi-service repository architecture, migration paths from current structure, and import stability strategy during the transition.

**Cross-references:**
- [Service Groups Taxonomy](./service-groups.md) - Defines 43 service groups
- [Dependency Analysis](./dependency-analysis.md) - Identifies coupling risks and migration phases

---

## Target Directory Structure

### Root Layout

```
cryptoutil/
├── cmd/                          # CLI entry points (one per service group)
│   ├── cryptoutil/              # KMS CLI (legacy name, keep for compatibility)
│   ├── kms/                     # KMS CLI (canonical name going forward)
│   ├── identity/                # Identity services CLI bundle
│   │   ├── authz/              # Authorization server CLI
│   │   ├── idp/                # Identity provider CLI
│   │   └── rs/                 # Resource server CLI
│   ├── ca/                      # CA service CLI (future)
│   ├── secrets/                 # Secrets management CLI (future)
│   ├── vault/                   # Vault service CLI (future)
│   └── cicd/                    # CICD utilities CLI (existing)
├── internal/                     # Private application code
│   ├── kms/                     # KMS domain (extracted from server/)
│   │   ├── server/             # KMS server (HTTP handlers, routing)
│   │   ├── businesslogic/      # KMS business logic (barrier, key management)
│   │   ├── repository/         # KMS data access (SQL, GORM)
│   │   ├── client/             # KMS Go client SDK
│   │   ├── crypto/             # KMS-specific crypto (JOSE, JWE/JWS)
│   │   ├── pool/               # KMS worker pools (keygen, crypto ops)
│   │   ├── telemetry/          # KMS observability (OTEL, metrics)
│   │   └── config/             # KMS configuration management
│   ├── identity/                # Identity domain (existing, well-isolated)
│   │   ├── authz/              # Authorization server
│   │   ├── idp/                # Identity provider
│   │   ├── rs/                 # Resource server
│   │   ├── domain/             # Shared identity models
│   │   ├── repository/         # Identity data access (GORM)
│   │   ├── config/             # Identity configuration
│   │   └── server/             # Identity HTTP infrastructure
│   ├── ca/                      # CA domain (future)
│   │   ├── server/             # CA server
│   │   ├── businesslogic/      # CA operations (issuance, revocation)
│   │   ├── repository/         # CA data access
│   │   └── config/             # CA configuration
│   ├── common/                  # Truly shared utilities (minimal)
│   │   ├── config/             # Configuration primitives
│   │   ├── magic/              # Magic constants
│   │   └── util/               # General-purpose utilities (files, network)
│   ├── cmd/                     # Internal command packages
│   │   └── cicd/               # CICD utility implementations
│   └── test/                    # Shared test infrastructure
│       └── e2e/                # End-to-end test framework
├── pkg/                          # Public library code (reusable across services)
│   └── crypto/                  # General-purpose crypto (promoted from internal/common)
│       ├── keygen/             # Key generation (RSA, ECDSA, EdDSA, AES, HMAC)
│       ├── digests/            # Hash functions (SHA, BLAKE)
│       ├── asn1/               # ASN.1 encoding/decoding
│       └── certificate/        # X.509 certificate operations
├── api/                          # OpenAPI specs and generated code
│   ├── kms/                     # KMS OpenAPI specs
│   │   ├── client/             # Generated KMS client
│   │   ├── model/              # Generated KMS models
│   │   └── server/             # Generated KMS server stubs
│   ├── identity/                # Identity OpenAPI specs (existing)
│   │   ├── authz/              # AuthZ server API
│   │   ├── idp/                # IdP API
│   │   └── rs/                 # Resource server API
│   └── ca/                      # CA OpenAPI specs (future)
└── configs/                      # Configuration templates
    ├── kms/                     # KMS configs
    ├── identity/                # Identity configs (existing)
    └── ca/                      # CA configs (future)
```

---

## Migration Paths

### Phase 1: Identity Domain (Already Complete ✅)

**Current state:** Identity is already well-isolated in `internal/identity/`

**Status:** No migration needed - architecture already follows target blueprint

**Validation:**
- ✅ Identity imports only `internal/common/magic` (LOW coupling)
- ✅ depguard + cicd checks enforce isolation
- ✅ Repository structure matches target layout

---

### Phase 2: KMS Domain Extraction (High Priority ⚠️)

**Current state:** KMS scattered across `internal/server/`, `internal/client/`, `cmd/cryptoutil/`

**Target state:** KMS consolidated under `internal/kms/`

**Migration sequence:**

#### Step 2.1: Extract KMS-Specific Utilities

Move KMS-coupled utilities from `internal/common/` to `internal/kms/`:

| Source | Destination | Rationale |
|--------|-------------|-----------|
| `internal/common/crypto/jose/` | `internal/kms/crypto/jose/` | KMS-specific JOSE operations |
| `internal/common/pool/` | `internal/kms/pool/` | KMS worker pools (keygen, crypto) |
| `internal/common/telemetry/` | `internal/kms/telemetry/` | KMS observability patterns |
| `internal/common/container/` | `internal/kms/container/` | KMS dependency injection |

**Import updates:**
```go
// Before
import cryptoutilJose "cryptoutil/internal/common/crypto/jose"
import cryptoutilPool "cryptoutil/internal/common/pool"

// After
import cryptoutilKmsJose "cryptoutil/internal/kms/crypto/jose"
import cryptoutilKmsPool "cryptoutil/internal/kms/pool"
```

#### Step 2.2: Promote General-Purpose Crypto to pkg/

Move reusable crypto from `internal/common/crypto/` to `pkg/crypto/`:

| Source | Destination | Rationale |
|--------|-------------|-----------|
| `internal/common/crypto/keygen/` | `pkg/crypto/keygen/` | General key generation |
| `internal/common/crypto/digests/` | `pkg/crypto/digests/` | Hash functions (any service) |
| `internal/common/crypto/asn1/` | `pkg/crypto/asn1/` | ASN.1 encoding (any service) |
| `internal/common/crypto/certificate/` | `pkg/crypto/certificate/` | X.509 operations (CA, KMS, Identity) |

**Import updates:**
```go
// Before
import cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
import cryptoutilDigests "cryptoutil/internal/common/crypto/digests"

// After
import "cryptoutil/pkg/crypto/keygen"
import "cryptoutil/pkg/crypto/digests"
```

**Why pkg/ promotion:**
- CA service will need keygen, certificate, asn1
- Identity may need digests for password hashing
- These are general-purpose crypto primitives (not KMS-specific)

#### Step 2.3: Relocate KMS Server Packages

Consolidate KMS server code under `internal/kms/`:

| Source | Destination | Rationale |
|--------|-------------|-----------|
| `internal/server/barrier/` | `internal/kms/businesslogic/barrier/` | KMS barrier unsealing logic |
| `internal/server/businesslogic/` | `internal/kms/businesslogic/` | KMS key operations |
| `internal/server/handler/` | `internal/kms/server/handler/` | KMS HTTP handlers |
| `internal/server/repository/` | `internal/kms/repository/` | KMS data access |
| `internal/server/config/` | `internal/kms/config/` | KMS configuration |
| `internal/client/` | `internal/kms/client/` | KMS Go client SDK |

**Import updates:**
```go
// Before
import cryptoutilBarrier "cryptoutil/internal/server/barrier"
import cryptoutilBusinesslogic "cryptoutil/internal/server/businesslogic"

// After
import cryptoutilKmsBarrier "cryptoutil/internal/kms/businesslogic/barrier"
import cryptoutilKmsBusinesslogic "cryptoutil/internal/kms/businesslogic"
```

#### Step 2.4: Update API Generation

Reorganize OpenAPI generated code:

| Source | Destination | Rationale |
|--------|-------------|-----------|
| `api/client/` | `api/kms/client/` | KMS-specific API client |
| `api/model/` | `api/kms/model/` | KMS data models |
| `api/server/` | `api/kms/server/` | KMS server stubs |

**Import updates:**
```go
// Before
import cryptoutilOpenapiClient "cryptoutil/api/client"
import cryptoutilOpenapiModel "cryptoutil/api/model"

// After
import cryptoutilKmsClient "cryptoutil/api/kms/client"
import cryptoutilKmsModel "cryptoutil/api/kms/model"
```

---

### Phase 3: CA Domain Preparation (Future)

**Current state:** CA planned but not implemented

**Target state:** `internal/ca/` structure ready for implementation

**Preparation steps:**

1. **Directory scaffold:**
   ```
   internal/ca/
   ├── server/           # CA HTTP server
   ├── businesslogic/    # Certificate issuance/revocation
   ├── repository/       # CA data access (certificates, CRLs)
   ├── config/           # CA configuration
   └── crypto/           # CA-specific crypto (if needed)
   ```

2. **CLI scaffold:**
   ```
   cmd/ca/
   ├── main.go          # CA CLI entry point
   ├── server.go        # Server management commands
   ├── issue.go         # Certificate issuance commands
   ├── revoke.go        # Revocation commands
   └── audit.go         # Audit log queries
   ```

3. **API scaffold:**
   ```
   api/ca/
   ├── openapi_spec.yaml      # CA API specification
   ├── client/                # Generated CA client
   ├── model/                 # Generated CA models
   └── server/                # Generated CA server stubs
   ```

**Import conventions:**
```go
import cryptoutilCaServer "cryptoutil/internal/ca/server"
import cryptoutilCaBusinesslogic "cryptoutil/internal/ca/businesslogic"
import cryptoutilCaClient "cryptoutil/api/ca/client"
```

---

## Import Stability Strategy

### Backward Compatibility During Migration

**Goal:** Allow gradual migration without breaking existing code

**Approach:** Temporary import aliases in `.golangci.yml` + deprecation warnings

#### Compatibility Shims (Temporary)

Create forwarding packages during migration:

```go
// internal/server/barrier/compat.go (temporary)
package barrier

import impl "cryptoutil/internal/kms/businesslogic/barrier"

// Deprecated: Use cryptoutil/internal/kms/businesslogic/barrier instead
type Barrier = impl.Barrier

// Deprecated: Use cryptoutil/internal/kms/businesslogic/barrier instead
func NewBarrier(ctx context.Context, logger *slog.Logger) (*Barrier, error) {
    return impl.NewBarrier(ctx, logger)
}
```

**Deprecation timeline:**
- Week 1-2: Create new structure, add compatibility shims
- Week 3-4: Update all internal imports to new paths
- Week 5-6: Update tests, configs, documentation
- Week 7-8: Remove compatibility shims (breaking change)

#### Import Alias Updates

Update `.golangci.yml` importas rules incrementally:

```yaml
# Phase 1: Add new aliases
linters-settings:
  importas:
    alias:
      # KMS domain (new)
      - pkg: cryptoutil/internal/kms/server
        alias: cryptoutilKmsServer
      - pkg: cryptoutil/internal/kms/businesslogic
        alias: cryptoutilKmsBusinesslogic

      # Legacy (deprecated, remove after migration)
      - pkg: cryptoutil/internal/server/barrier
        alias: cryptoutilBarrier  # Deprecated
```

**Migration validation:**
- Run `golangci-lint run` after each package move
- Use `go list -m all` to verify module consistency
- Test full build: `go build ./...`
- Run full test suite: `go test ./...`

---

## Migration Sequence Diagrams

### Phase 2: KMS Extraction Flow

```
[Current Structure]
internal/
├── server/
│   ├── barrier/
│   ├── businesslogic/
│   ├── handler/
│   └── repository/
├── client/
└── common/
    ├── crypto/
    │   ├── jose/
    │   ├── keygen/
    │   └── digests/
    └── pool/

        ↓
   [Step 2.1: Extract KMS-Specific Utils]

internal/
├── kms/
│   ├── crypto/
│   │   └── jose/        ← from internal/common/crypto/jose
│   ├── pool/            ← from internal/common/pool
│   └── telemetry/       ← from internal/common/telemetry
└── common/
    └── crypto/
        ├── keygen/      (remains temporarily)
        └── digests/     (remains temporarily)

        ↓
   [Step 2.2: Promote General Crypto to pkg/]

pkg/
└── crypto/
    ├── keygen/          ← from internal/common/crypto/keygen
    ├── digests/         ← from internal/common/crypto/digests
    ├── asn1/            ← from internal/common/crypto/asn1
    └── certificate/     ← from internal/common/crypto/certificate

        ↓
   [Step 2.3: Relocate KMS Server Packages]

internal/
└── kms/
    ├── server/
    │   └── handler/     ← from internal/server/handler
    ├── businesslogic/
    │   ├── barrier/     ← from internal/server/barrier
    │   └── (core)       ← from internal/server/businesslogic
    ├── repository/      ← from internal/server/repository
    ├── client/          ← from internal/client
    ├── config/          ← from internal/server/config
    ├── crypto/
    │   └── jose/
    ├── pool/
    └── telemetry/

        ↓
   [Step 2.4: Update API Generation]

api/
└── kms/
    ├── client/          ← from api/client
    ├── model/           ← from api/model
    └── server/          ← from api/server

        ↓
   [Final State: Clean KMS Domain]

internal/kms/            ✅ Isolated KMS domain
pkg/crypto/              ✅ Reusable crypto primitives
internal/common/         ✅ Minimal shared utilities
```

---

## Package Relocation Rules

### Rule 1: Domain-Specific Code → internal/<group>/

**Applies to:** Code tightly coupled to one service group

**Examples:**
- KMS barrier unsealing → `internal/kms/businesslogic/barrier/`
- Identity OAuth handlers → `internal/identity/authz/handler/`
- CA certificate issuance → `internal/ca/businesslogic/issuance/`

**Test:** If removing the service group makes the code useless → domain-specific

---

### Rule 2: Reusable Libraries → pkg/

**Applies to:** Code usable across multiple service groups or external projects

**Examples:**
- Crypto key generation → `pkg/crypto/keygen/`
- ASN.1 encoding → `pkg/crypto/asn1/`
- X.509 certificate parsing → `pkg/crypto/certificate/`

**Test:** If code could be extracted into standalone module → promote to pkg/

---

### Rule 3: Shared Utilities → internal/common/

**Applies to:** Internal-only code used by 2+ service groups

**Examples:**
- Magic constants → `internal/common/magic/`
- File utilities → `internal/common/util/files/`
- Network utilities → `internal/common/util/network/`

**Test:** If code is internal implementation detail shared across groups → keep in common/

---

## Validation Checklist

### Pre-Migration Validation

- [ ] Run full test suite: `go test ./... --count=1 -timeout=10m`
- [ ] Generate baseline coverage: `go test ./... -coverprofile=test-output/coverage_baseline.out`
- [ ] Run linters: `golangci-lint run ./...`
- [ ] Document current import graph: `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...`
- [ ] Backup current structure: `git tag pre-migration-$(date +%Y%m%d)`

### Per-Package Migration Validation

- [ ] Move package files to new location
- [ ] Update import statements in moved files
- [ ] Update import statements in files importing moved package
- [ ] Update `.golangci.yml` importas rules
- [ ] Run tests for moved package: `go test ./<new-path>`
- [ ] Run tests for packages importing moved package
- [ ] Run full test suite: `go test ./...`
- [ ] Verify coverage unchanged: `go test ./... -coverprofile=test-output/coverage_new.out`
- [ ] Run linters: `golangci-lint run ./...`
- [ ] Update documentation references
- [ ] Commit migration: `git commit -m "refactor: move <package> to <new-path>"`

### Post-Migration Validation

- [ ] Run full test suite: `go test ./... --count=1 -timeout=10m`
- [ ] Compare coverage: baseline vs post-migration
- [ ] Run all workflows: `go run ./cmd/workflow -workflows=all`
- [ ] Verify CLI commands work: `go run ./cmd/kms`, `go run ./cmd/identity`
- [ ] Update README with new import paths
- [ ] Remove compatibility shims (after grace period)
- [ ] Document migration in CHANGELOG
- [ ] Tag release: `git tag v<version>-refactor`

---

## Cross-References

- **Service Groups Taxonomy:** [docs/01-refactor/service-groups.md](./service-groups.md)
- **Dependency Analysis:** [docs/01-refactor/dependency-analysis.md](./dependency-analysis.md)
- **Import Alias Policy:** [docs/01-refactor/import-aliases.md](./import-aliases.md) (future)
- **CLI Strategy Framework:** [docs/01-refactor/cli-strategy.md](./cli-strategy.md) (future)

---

## Notes

- **Import stability:** Use compatibility shims during migration to avoid breaking consumers
- **Testing discipline:** Run full test suite after each package move
- **Incremental migration:** Move one package at a time, validate, commit, repeat
- **Documentation updates:** Update README and docs/ hierarchy as packages move
- **Linter enforcement:** Add new importas rules immediately after package moves
