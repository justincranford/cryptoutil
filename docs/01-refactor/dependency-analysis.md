# Repository Inventory and Coupling Analysis

**Last Updated**: 2025-11-21
**Purpose**: Map package dependencies and identify cross-group coupling risks for repository refactor
**Scope**: Analysis of internal/**, cmd/**, api/** packages

---

## Executive Summary

**Total Packages Analyzed**: 85 packages across 4 major domains (KMS, Identity, CICD, Shared)

**Key Findings**:
- ✅ **Identity domain isolation working well** - only imports `internal/common/magic` as intended
- ⚠️ **KMS/Server domain highly coupled** - `internal/server/**` imports from common crypto, jose, telemetry
- ⚠️ **Common utilities need extraction** - several utilities used across domains (crypto, pool, telemetry)
- ✅ **CICD utilities properly isolated** - cicd commands import only common utilities, no cross-domain coupling

**Coupling Risk Assessment**: MEDIUM
- Identity → Common: LOW (only magic constants)
- KMS → Common: HIGH (crypto, jose, telemetry, pool, config, container, util)
- CICD → Common: LOW (magic, util/files)

---

## Part 1: Package Inventory by Domain

### Domain 1: KMS (Key Management Service)

**Root Packages**: 12 packages

| Package Path | Purpose | External Dependencies | Internal Dependencies |
|--------------|---------|----------------------|----------------------|
| `internal/server/application` | Fiber HTTP server, API orchestration | fiber/v2, otelfiber, swagger | server/barrier, server/handler, server/businesslogic, server/repository, common/config, common/crypto/certificate, common/crypto/jose, common/telemetry, common/util |
| `internal/server/barrier` | Hierarchical key management (unseal → root → intermediate → content) | None | barrier/*, common/crypto/jose, common/telemetry, server/repository/orm |
| `internal/server/barrier/unsealkeysservice` | Unseal key derivation (3-of-5, 5-of-5 modes) | jwx/v3 | common/config, common/crypto/digests, common/crypto/jose, common/crypto/keygen, common/magic, common/telemetry, common/util/combinations, common/util/sysinfo |
| `internal/server/barrier/rootkeysservice` | Root key generation and storage | jwx/v3, gorm | common/apperr, common/crypto/jose, common/telemetry, barrier/unsealkeysservice, server/repository/orm |
| `internal/server/barrier/intermediatekeysservice` | Intermediate key generation | jwx/v3, gorm | common/apperr, common/crypto/jose, common/telemetry, barrier/rootkeysservice, server/repository/orm |
| `internal/server/barrier/contentkeysservice` | Content key management (tenant keys) | jwx/v3 | common/crypto/jose, common/telemetry, barrier/intermediatekeysservice, server/repository/orm |
| `internal/server/businesslogic` | Business logic layer (key operations) | jwx/v3 | api/model, common/crypto/jose, common/magic, common/telemetry, common/util, server/barrier, server/repository/orm |
| `internal/server/handler` | HTTP handler implementations | None | api/model, api/server, common/apperr, server/businesslogic |
| `internal/server/repository/orm` | GORM ORM layer (PostgreSQL, SQLite) | gorm, pgx/v5, modernc.org/sqlite | api/model, common/apperr, common/config, common/crypto/jose, common/magic, common/telemetry, common/util, server/repository/sqlrepository |
| `internal/server/repository/sqlrepository` | SQL migrations and database providers | gorm, golang-migrate/migrate/v4, pgx/v5, modernc.org/sqlite | common/apperr, common/config, common/container, common/magic, common/telemetry |
| `internal/client` | KMS Go client SDK | None | api/client, api/model, common/crypto/jose, common/magic |
| `cmd/cryptoutil` | KMS CLI application | None | common/config, server/application |

**Key Observations**:
- Server packages heavily depend on `common/crypto/jose` for JWE/JWS operations
- All layers use `common/telemetry` for observability (OTEL)
- Barrier services form clear hierarchy: unseal → root → intermediate → content
- Repository layer abstracts database operations (PostgreSQL, SQLite)

---

### Domain 2: Identity (OAuth 2.1 / OIDC)

**Root Packages**: 23 packages

| Package Path | Purpose | External Dependencies | Internal Dependencies |
|--------------|---------|----------------------|----------------------|
| `internal/identity/authz` | OAuth 2.1 Authorization Server | fiber/v2, jwx/v3 | identity/apperr, identity/authz/clientauth, identity/authz/pkce, identity/config, identity/domain, identity/issuer, identity/magic, identity/repository |
| `internal/identity/idp` | OIDC Identity Provider | fiber/v2 | identity/config, identity/domain, identity/idp/auth, identity/issuer, identity/magic, identity/repository |
| `internal/identity/rs` | Resource Server | fiber/v2 | identity/config, identity/issuer, identity/magic |
| `internal/identity/authz/clientauth` | Client authentication (secret, JWT, mTLS) | jwx/v3 | identity/apperr, identity/domain, identity/magic, identity/repository |
| `internal/identity/authz/pkce` | Proof Key for Code Exchange | None | identity/magic |
| `internal/identity/idp/auth` | User authentication (password, TOTP, SMS) | bcrypt | identity/apperr, identity/domain, identity/repository |
| `internal/identity/idp/userauth` | TOTP/MFA implementation | bcrypt | identity/domain, identity/magic, identity/repository |
| `internal/identity/issuer` | Token issuance service | None | identity/apperr, identity/config, identity/domain, identity/magic |
| `internal/identity/security` | Security utilities (rate limiting, IP allowlisting) | None | identity/domain, identity/magic |
| `internal/identity/jobs` | Background jobs (session cleanup, token expiry) | None | identity/repository |
| `internal/identity/repository` | Database abstraction (GORM) | gorm, golang-migrate/migrate/v4, pgx/v5, modernc.org/sqlite | identity/apperr, identity/config, identity/domain, identity/repository/orm |
| `internal/identity/repository/orm` | GORM repository implementations | gorm | identity/apperr, identity/domain |
| `internal/identity/domain` | Domain models (User, Client, Token, Session, AuthFlow, AuthProfile, MFAFactor) | gorm | common/magic (NullableUUID only) |
| `internal/identity/config` | Configuration loading and validation | gopkg.in/yaml.v3 | common/magic |
| `internal/identity/server` | HTTP server orchestration | fiber/v2 | identity/authz, identity/config, identity/idp, identity/issuer, identity/magic, identity/repository, identity/rs |
| `cmd/identity/authz` | AuthZ CLI application | None | identity/config, identity/issuer, identity/magic, identity/repository, identity/server |
| `cmd/identity/idp` | IdP CLI application | None | identity/config, identity/issuer, identity/magic, identity/repository, identity/server |
| `cmd/identity/rs` | RS CLI application | None | identity/config, identity/issuer, identity/magic, identity/server |
| `cmd/identity/spa-rp` | SPA Relying Party demo | None | identity/magic |
| `api/identity/authz` | OpenAPI specs (future) | None | None |
| `api/identity/idp` | OpenAPI specs (future) | None | None |
| `api/identity/rs` | OpenAPI specs (future) | None | None |
| `internal/identity/test/e2e` | E2E testing infrastructure | None | common/magic |

**Key Observations**:
- ✅ **Domain isolation enforced** - Identity ONLY imports `internal/common/magic` from common
- ✅ **Self-contained crypto** - Uses stdlib crypto (bcrypt, SHA-256, HMAC) instead of common/crypto
- ✅ **No KMS coupling** - Identity does NOT import server, client, or api packages
- Clean layered architecture: HTTP → Services → Repository → Domain
- GORM-based persistence with cross-DB compatibility (PostgreSQL, SQLite)

---

### Domain 3: CICD (Build & Code Quality Tools)

**Root Packages**: 12 packages

| Package Path | Purpose | External Dependencies | Internal Dependencies |
|--------------|---------|----------------------|----------------------|
| `internal/cmd/cicd` | CICD dispatcher | None | cmd/cicd/*, common/magic, common/util/files |
| `internal/cmd/cicd/all_enforce_utf8` | UTF-8 encoding enforcement | None | cmd/cicd/common, common/magic |
| `internal/cmd/cicd/go_enforce_any` | Replace `interface{}` with `any` | None | cmd/cicd/common, common/magic, common/util/files |
| `internal/cmd/cicd/go_enforce_test_patterns` | Test pattern enforcement (table-driven) | None | cmd/cicd/common, common/magic |
| `internal/cmd/cicd/go_check_circular_package_dependencies` | Circular dependency detection | None | cmd/cicd/common, common/magic, common/util/files |
| `internal/cmd/cicd/go_check_identity_imports` | Identity domain isolation validation | None | cmd/cicd/common, common/magic |
| `internal/cmd/cicd/go_fix_staticcheck_error_strings` | Error string capitalization fixes | None | cmd/cicd/common |
| `internal/cmd/cicd/go_fix_copyloopvar` | Loop variable capture fixes | None | cmd/cicd/common |
| `internal/cmd/cicd/go_fix_thelper` | Test helper annotation fixes | None | cmd/cicd/common |
| `internal/cmd/cicd/go_fix_all` | Orchestrator for all fixes | None | cmd/cicd/go_fix_* |
| `internal/cmd/cicd/go_update_direct_dependencies` | Go dependency updates (direct only) | None | cmd/cicd/common, common/magic, common/util/files |
| `internal/cmd/cicd/github_workflow_lint` | GitHub Actions workflow validation | None | cmd/cicd/common, cmd/cicd/go_update_direct_dependencies, common/magic |

**Key Observations**:
- ✅ **No cross-domain coupling** - CICD tools only import common utilities
- ✅ **Self-exclusion pattern** - Commands exclude their own subdirectories from processing
- Tools organized as flat snake_case subdirectories (NOT categorized)
- Each command has production code (`*.go`) and test code (`*_test.go`) in same subdirectory

---

### Domain 4: Shared Utilities (Common)

**Root Packages**: 20 packages

| Package Path | Purpose | External Dependencies | Internal Dependencies |
|--------------|---------|----------------------|----------------------|
| `internal/common/apperr` | Application error types (HTTP status mapping) | None | None |
| `internal/common/config` | Configuration loading (YAML, env vars, flags) | viper, pflag | common/magic |
| `internal/common/container` | Testcontainers integration (Docker) | testcontainers-go | common/magic, common/telemetry |
| `internal/common/crypto/asn1` | ASN.1/DER/PEM utilities | None | common/magic |
| `internal/common/crypto/certificate` | Certificate generation (CA, TLS, CSR) | None | common/crypto/keygen, common/magic, common/util/network |
| `internal/common/crypto/digests` | HKDF key derivation, SHA-256/512 | golang.org/x/crypto/hkdf | common/magic |
| `internal/common/crypto/jose` | JWE/JWS operations (encryption, signing) | jwx/v3, circl/sign/ed448 | api/model, common/apperr, common/crypto/keygen, common/magic, common/pool, common/telemetry, common/util |
| `internal/common/crypto/keygen` | Key generation (RSA, ECDSA, ECDH, EdDSA, AES, HMAC, UUIDv7) | circl/sign/ed448 | common/magic, common/util |
| `internal/common/crypto/keygenpooltest` | Key generation pool test utilities | None | common/apperr, common/crypto/keygen, common/magic, common/pool, common/telemetry, common/util |
| `internal/common/magic` | Magic constants (timeouts, buffer sizes, network settings) | None | None |
| `internal/common/pool` | Generic resource pool (concurrent key generation) | otel/metric | common/magic, common/telemetry |
| `internal/common/telemetry` | OpenTelemetry integration (traces, metrics, logs) | otel/*, slog-multi | common/apperr, common/config, common/magic |
| `internal/common/testutil` | Test utilities (temp directories, file assertions) | testify/require | common/magic |
| `internal/common/util` | General utilities (JSON, YAML, UUID, random) | goccy/go-yaml | common/apperr |
| `internal/common/util/combinations` | Combinatorics (nCr for unseal keys) | None | None |
| `internal/common/util/datetime` | Date/time formatting | None | None |
| `internal/common/util/files` | File I/O utilities | None | None |
| `internal/common/util/network` | Network utilities (TLS, HTTP clients) | None | common/magic |
| `internal/common/util/sysinfo` | System information (CPU, memory, hostname) | gopsutil | common/magic, common/util |
| `internal/common/util/thread` | Thread-safe utilities (sync helpers) | None | None |

**Key Observations**:
- ✅ **Shared foundation** - Used by KMS and CICD (NOT Identity, except magic)
- `common/crypto/jose` is KMS-specific (should move to KMS domain)
- `common/pool` is KMS-specific (should move to KMS domain)
- `common/telemetry` could remain shared (used by KMS, could be used by Identity in future)
- `common/magic` is truly shared (used by all domains)
- `common/util` is general-purpose (could remain shared or split by domain)

---

## Part 2: Coupling Analysis

### Cross-Domain Import Matrix

| From Domain | To: KMS | To: Identity | To: CICD | To: Common |
|-------------|---------|--------------|----------|------------|
| **KMS** | ✅ Internal | ❌ None | ❌ None | ⚠️ HIGH (crypto, jose, pool, telemetry, config, container, util) |
| **Identity** | ❌ None | ✅ Internal | ❌ None | ✅ LOW (magic only) |
| **CICD** | ❌ None | ❌ None | ✅ Internal | ✅ LOW (magic, util/files) |
| **Common** | ❌ None | ❌ None | ❌ None | ✅ Internal |

### Coupling Risk Assessment

#### Risk Level: LOW (Identity → Common)
- **Import**: `internal/common/magic` ONLY
- **Purpose**: Shared constants (timeouts, buffer sizes, file permissions)
- **Impact**: Read-only dependency; no business logic coupling
- **Mitigation**: None needed - intentional design

#### Risk Level: MEDIUM (CICD → Common)
- **Import**: `internal/common/magic`, `internal/common/util/files`
- **Purpose**: Magic constants + file I/O utilities
- **Impact**: Utility dependency; no domain logic coupling
- **Mitigation**: None needed - tools legitimately need file utilities

#### Risk Level: HIGH (KMS → Common)
- **Import**: `common/crypto/jose`, `common/pool`, `common/telemetry`, `common/config`, `common/container`, `common/util/*`, `common/magic`
- **Purpose**: Cryptographic operations, resource pooling, observability, configuration
- **Impact**: Heavy coupling to shared utilities; refactor risk when moving KMS packages
- **Mitigation Options**:
  1. **Move KMS-specific utilities to KMS domain**: `crypto/jose`, `pool` → `internal/kms/crypto/jose`, `internal/kms/pool`
  2. **Keep truly shared utilities in common**: `telemetry`, `config`, `magic`, `util` (general-purpose)
  3. **Extract domain-agnostic crypto to pkg**: `crypto/keygen`, `crypto/digests` → `pkg/crypto/*`

---

## Part 3: Proposed Mitigation Strategies

### Strategy 1: Extract KMS-Specific Utilities

**Move to `internal/kms/crypto/jose`**:
- `internal/common/crypto/jose` (JWE/JWS is KMS-specific, not used by Identity)

**Move to `internal/kms/pool`**:
- `internal/common/pool` (Resource pooling for concurrent key generation - KMS only)

**Rationale**: These utilities are tightly coupled to KMS business logic; no other domain uses them.

---

### Strategy 2: Promote General-Purpose Crypto to `pkg/`

**Move to `pkg/crypto/keygen`**:
- `internal/common/crypto/keygen` (Key generation primitives could be reused by CA, Secrets, Vault)

**Move to `pkg/crypto/digests`**:
- `internal/common/crypto/digests` (HKDF key derivation - general-purpose)

**Move to `pkg/crypto/asn1`**:
- `internal/common/crypto/asn1` (ASN.1/DER/PEM utilities - needed by CA)

**Move to `pkg/crypto/certificate`**:
- `internal/common/crypto/certificate` (Certificate generation - needed by CA)

**Rationale**: These utilities implement cryptographic primitives that could be used by multiple service groups (CA, Secrets, Vault). Making them public (`pkg/`) enables reuse while maintaining clean boundaries.

---

### Strategy 3: Keep Truly Shared Utilities in `internal/common/`

**Retain in `internal/common/`**:
- `common/telemetry` (OpenTelemetry - all services need observability)
- `common/config` (Configuration loading - all services need config)
- `common/magic` (Shared constants - all domains use this)
- `common/apperr` (Application errors - used by KMS and Identity)
- `common/testutil` (Test utilities - used across domains)
- `common/util` (General utilities - JSON, YAML, UUID, random number generation)
- `common/container` (Testcontainers - used by KMS and Identity integration tests)

**Rationale**: These utilities provide infrastructure concerns (observability, configuration, testing) that span all service groups.

---

### Strategy 4: Validate Identity Domain Isolation

**Current State**: ✅ Identity only imports `internal/common/magic`

**golangci-lint depguard enforcement** (already in place):
```yaml
identity-domain-isolation:
  deny:
    - pkg: "cryptoutil/internal/server"  # KMS server
    - pkg: "cryptoutil/internal/client"  # KMS client
    - pkg: "cryptoutil/api"              # OpenAPI generated
    - pkg: "cryptoutil/cmd/cryptoutil"   # KMS CLI
    - pkg: "cryptoutil/internal/common/crypto"  # Use stdlib instead
    - pkg: "cryptoutil/internal/common/pool"     # KMS infrastructure
    - pkg: "cryptoutil/internal/common/container" # KMS infrastructure
    - pkg: "cryptoutil/internal/common/telemetry" # KMS infrastructure
    - pkg: "cryptoutil/internal/common/util"      # KMS infrastructure
```

**Additional CICD check**: `go run ./cmd/cicd go-check-identity-imports`

**Rationale**: Identity domain isolation is critical for independent deployment. Enforcement prevents regressions.

---

## Part 4: Migration Sequence

### Phase 1: Extract KMS-Specific Utilities
1. Move `internal/common/crypto/jose` → `internal/kms/crypto/jose`
2. Move `internal/common/pool` → `internal/kms/pool`
3. Update all KMS imports to use new paths
4. Update `golangci-lint` importas rules
5. Run full test suite: `go test ./...`
6. Run linter: `golangci-lint run`

**Impact**: LOW - Only KMS packages affected

---

### Phase 2: Promote General-Purpose Crypto to pkg/
1. Move `internal/common/crypto/keygen` → `pkg/crypto/keygen`
2. Move `internal/common/crypto/digests` → `pkg/crypto/digests`
3. Move `internal/common/crypto/asn1` → `pkg/crypto/asn1`
4. Move `internal/common/crypto/certificate` → `pkg/crypto/certificate`
5. Update KMS imports to use `pkg/crypto/*`
6. Update `golangci-lint` importas rules
7. Run full test suite: `go test ./...`
8. Run linter: `golangci-lint run`

**Impact**: MEDIUM - KMS packages + future CA packages

---

### Phase 3: Reorganize KMS Packages
1. Move `internal/server` → `internal/kms/server`
2. Move `internal/client` → `internal/kms/client`
3. Move `cmd/cryptoutil` → `cmd/kms`
4. Move `api/server` → `api/kms/server`
5. Move `api/client` → `api/kms/client`
6. Update all imports across codebase
7. Update `golangci-lint` importas rules
8. Run full test suite: `go test ./...`
9. Run linter: `golangci-lint run`

**Impact**: HIGH - Touches all KMS packages, requires comprehensive import updates

---

## Part 5: Validation Checklist

### Pre-Refactor Validation
- [ ] Run full test suite: `go test ./... -cover`
- [ ] Run linter: `golangci-lint run`
- [ ] Run CICD checks: `go run ./cmd/cicd go-check-circular-package-dependencies`
- [ ] Run CICD checks: `go run ./cmd/cicd go-check-identity-imports`
- [ ] Document current import graph: `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...`
- [ ] Capture baseline test coverage: `go test ./... -coverprofile=test-output/coverage_baseline.out`

### Post-Refactor Validation
- [ ] All tests pass: `go test ./... -cover`
- [ ] No new lint errors: `golangci-lint run`
- [ ] No circular dependencies: `go run ./cmd/cicd go-check-circular-package-dependencies`
- [ ] Identity isolation maintained: `go run ./cmd/cicd go-check-identity-imports`
- [ ] Test coverage unchanged or improved
- [ ] All GitHub workflows pass (Quality, Coverage, E2E, DAST, Load)
- [ ] Documentation updated (README.md, docs/README.md, import alias docs)

---

## Part 6: Cross-References

- **Service Groups Taxonomy**: `docs/01-refactor/service-groups.md` (Task 1)
- **Directory Blueprint**: `docs/01-refactor/blueprint.md` (Task 3 - to be created)
- **Import Alias Policy**: `docs/01-refactor/import-aliases.md` (Task 4 - to be created)
- **Identity Dependency Audit**: `docs/03-identityV2/dependency-graph.md`
- **golangci-lint Config**: `.golangci.yml` (importas and depguard rules)
- **CICD Identity Imports Check**: `internal/cmd/cicd/go_check_identity_imports/identityimports.go`

---

## Appendix: Package Dependency Data

**Full package import relationships**: `test-output/package-imports.txt` (85 packages analyzed)
**JSON metadata**: `test-output/go-list-all.json` (comprehensive package metadata from `go list -json ./...`)
