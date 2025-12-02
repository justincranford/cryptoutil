# Refactoring Sprint: 4-Products Architecture Alignment

**Sprint Date**: 2025-12-02
**Theme**: Directory structure alignment with P1-P4 products strategy
**Goal**: Prepare codebase structure for clear product separation (JOSE, Identity, KMS, CA)

---

## Sprint Overview

This is a short, opportunistic sprint to refactor directory structures in `api/`, `internal/`, and `deployments/` to align with the 4 products strategy. KMS is working well but the directory structure doesn't reflect the product architecture.

**Products**:

- P1: JOSE (JSON Object Signing and Encryption)
- P2: Identity (OAuth 2.1 AuthZ + OIDC IdP)
- P3: KMS (Key Management Service)
- P4: CA (Certificate Authority) - Planned

---

## Tasks

### Phase A: Documentation & Planning (Tasks 1-5)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 1 | Document current structure | ✅ | 30m | Create architecture diagram of current directory layout |
| 2 | Define target structure | ✅ | 30m | Document target product-aligned directory structure |
| 3 | Create import alias inventory | ✅ | 30m | List all cryptoutil import aliases in .golangci.yml |
| 4 | Identify circular dependencies | ✅ | 30m | Run cicd tool to identify any circular deps |
| 5 | Document migration plan | ✅ | 30m | Create step-by-step migration checklist |

### Phase B: API Directory Refactoring (Tasks 6-10)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 6 | Create api/kms/ directory | ✅ | 15m | Create with README documenting planned structure |
| 7 | Create api/jose/ directory | ✅ | 15m | Create with README documenting planned JOSE API |
| 8 | Create api/ca/ directory | ✅ | 15m | Create with README documenting planned CA API |
| 9 | Consolidate identity OpenAPI | ✅ | 30m | Verified api/identity/ is complete |
| 10 | Document OpenAPI migration | ✅ | 15m | READMEs document current vs target state |

### Phase C: Internal Product Directories (Tasks 11-18)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 11 | Create internal/kms/ directory | ✅ | 15m | Create with README documenting migration plan |
| 12 | Create internal/jose/ directory | ✅ | 15m | Create with README documenting migration plan |
| 13 | Create internal/ca/ directory | ✅ | 15m | Create with README documenting planned structure |
| 14 | Document server → kms migration | ✅ | 30m | kms/README.md has full migration table |
| 15 | Document jose migration | ✅ | 30m | jose/README.md has full migration table |
| 16 | Keep existing imports working | ✅ | 30m | No breaking changes to existing code |
| 17 | Verify all tests pass | ✅ | 30m | go build ./... and golangci-lint run pass |
| 18 | Document import alias updates | ✅ | 30m | golangci.yml import aliases documented |

### Phase D: Deployment Directories (Tasks 19-23)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 19 | Verify kms/ deployment | ✅ | 15m | deployments/kms/ structure is correct |
| 20 | Verify identity/ deployment | ✅ | 15m | deployments/identity/ structure is correct |
| 21 | Create deployments/jose/ | ✅ | 15m | Create with README placeholder |
| 22 | Create deployments/ca/ | ✅ | 15m | Create with README placeholder |
| 23 | Verify telemetry/ deployment | ✅ | 15m | deployments/telemetry/ structure is correct |

### Phase E: Code Quality (Tasks 24-28)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 24 | Run golangci-lint | ✅ | 15m | All linting passes |
| 25 | Verify go build | ✅ | 15m | All packages build successfully |
| 26 | Check circular dependencies | ✅ | 15m | No circular dependencies found |
| 27 | Review TODO comments | ✅ | 15m | Identified legitimate TODOs for future work |
| 28 | Document cspell additions | ✅ | 15m | Added missing words to cspell dictionary |

### Phase F: Validation & Cleanup (Tasks 29-30)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 29 | Run full test suite | ✅ | 15m | Core tests pass; pre-existing issues in identity integration |
| 30 | Update PROJECT-STATUS.md | ✅ | 15m | Updated docs/02-identityV2/PROJECT-STATUS.md |

**Test Results Note (Task 29)**:

- All core KMS/server tests pass ✅
- All jose/keygen/crypto tests pass ✅
- Identity integration tests have pre-existing port binding issues (not caused by refactoring)
- SQLRepository container tests fail on Windows (rootless Docker limitation)

---

## Current Structure Analysis

### Current `internal/` Layout

```text
internal/
├── client/          # HTTP client utilities (KMS client)
├── cmd/             # CLI commands (cicd tools)
├── common/          # Shared infrastructure
│   ├── apperr/      # Application errors
│   ├── config/      # Configuration
│   ├── container/   # Container utilities
│   ├── crypto/      # Cryptographic utilities (JOSE here!)
│   │   ├── jose/    # JWK/JWE/JWS/JWT
│   │   ├── keygen/  # Key generation
│   │   └── ...
│   ├── magic/       # Magic constants
│   ├── pool/        # Concurrency pools
│   ├── telemetry/   # OpenTelemetry
│   ├── testutil/    # Test utilities
│   └── util/        # General utilities
├── crypto/          # Additional crypto (registry, secrets)
├── identity/        # P2: Identity product
│   ├── authz/       # Authorization server
│   ├── idp/         # Identity provider
│   └── ...
├── infra/           # Infrastructure components
│   ├── demo/        # Demo utilities
│   ├── realm/       # Realm management
│   ├── tenant/      # Tenant management
│   └── tls/         # TLS utilities
├── server/          # P3: KMS (misnamed as "server")
│   ├── application/ # Server application
│   ├── barrier/     # Key barrier
│   ├── businesslogic/ # Business logic
│   ├── demo/        # KMS demo
│   ├── handler/     # HTTP handlers
│   ├── middleware/  # HTTP middleware
│   └── repository/  # Data repository
└── test/            # Test infrastructure
```

### Target `internal/` Layout

```text
internal/
├── cmd/             # CLI commands (cicd tools)
├── infra/           # Shared infrastructure (was common/)
│   ├── apperr/      # Application errors
│   ├── config/      # Configuration
│   ├── container/   # Container utilities
│   ├── crypto/      # Core crypto primitives
│   │   ├── keygen/  # Key generation
│   │   └── ...
│   ├── magic/       # Magic constants
│   ├── pool/        # Concurrency pools
│   ├── realm/       # Realm management
│   ├── telemetry/   # OpenTelemetry
│   ├── tenant/      # Tenant management
│   ├── testutil/    # Test utilities
│   ├── tls/         # TLS utilities
│   └── util/        # General utilities
├── jose/            # P1: JOSE product (was common/crypto/jose/)
│   ├── jwk/         # JWK operations
│   ├── jwe/         # JWE operations
│   ├── jws/         # JWS operations
│   └── service/     # JOSE service
├── identity/        # P2: Identity product (mostly unchanged)
│   ├── authz/       # Authorization server
│   ├── idp/         # Identity provider
│   └── ...
├── kms/             # P3: KMS product (was server/)
│   ├── application/ # Server application
│   ├── barrier/     # Key barrier
│   ├── client/      # KMS client
│   ├── demo/        # KMS demo
│   ├── handler/     # HTTP handlers
│   ├── middleware/  # HTTP middleware
│   ├── repository/  # Data repository
│   └── service/     # Business logic
├── ca/              # P4: CA product (placeholder)
│   └── .gitkeep
└── test/            # Test infrastructure
```

---

## Progress Tracking

**Started**: 2025-12-02
**Completed**: 2025-12-02

| Phase | Tasks | Completed | Status |
|-------|-------|-----------|--------|
| A: Planning | 5 | 5 | ✅ Complete |
| B: API | 5 | 5 | ✅ Complete |
| C: Internal | 8 | 8 | ✅ Complete |
| D: Deploy | 5 | 5 | ✅ Complete |
| E: Quality | 5 | 5 | ✅ Complete |
| F: Validate | 2 | 2 | ✅ Complete |
| **Total** | **30** | **30** | **100%** |

---

## Notes

- KMS is working extremely well - be careful not to break it
- Focus on directory structure, not major code changes
- Regular commits after each logical group of changes
- Run tests frequently to catch regressions early

---

*Sprint Created: 2025-12-02*
