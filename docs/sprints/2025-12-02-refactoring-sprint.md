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
| 3 | Create import alias inventory | ⬜ | 30m | List all cryptoutil import aliases in .golangci.yml |
| 4 | Identify circular dependencies | ⬜ | 30m | Run cicd tool to identify any circular deps |
| 5 | Document migration plan | ⬜ | 30m | Create step-by-step migration checklist |

### Phase B: API Directory Refactoring (Tasks 6-10)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 6 | Create api/kms/ directory | ⬜ | 15m | Move KMS OpenAPI specs from api/ root |
| 7 | Create openapi_spec_kms.yaml | ⬜ | 30m | Rename/reorganize KMS OpenAPI spec files |
| 8 | Update api/generate.go for KMS | ⬜ | 30m | Update oapi-codegen config for new paths |
| 9 | Consolidate identity OpenAPI | ⬜ | 30m | Ensure api/identity/ is complete |
| 10 | Add api/jose/ placeholder | ⬜ | 15m | Create placeholder for future JOSE API |

### Phase C: Internal Server → KMS Migration (Tasks 11-18)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 11 | Create internal/kms/ directory | ⬜ | 15m | Create new KMS product directory |
| 12 | Move handler to kms/ | ⬜ | 30m | Move internal/server/handler/ → internal/kms/handler/ |
| 13 | Move businesslogic to kms/ | ⬜ | 30m | Move internal/server/businesslogic/ → internal/kms/service/ |
| 14 | Move repository to kms/ | ⬜ | 30m | Move internal/server/repository/ → internal/kms/repository/ |
| 15 | Move barrier to kms/ | ⬜ | 30m | Move internal/server/barrier/ → internal/kms/barrier/ |
| 16 | Move middleware to kms/ | ⬜ | 30m | Move internal/server/middleware/ → internal/kms/middleware/ |
| 17 | Move application to kms/ | ⬜ | 30m | Move internal/server/application/ → internal/kms/application/ |
| 18 | Update all KMS imports | ⬜ | 1h | Update import paths throughout codebase |

### Phase D: Common → Infra Reorganization (Tasks 19-23)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 19 | Rename common to infra | ⬜ | 30m | Evaluate renaming internal/common/ → internal/infra/ |
| 20 | Move crypto utilities | ⬜ | 30m | Organize internal/common/crypto/ structure |
| 21 | Consolidate magic packages | ⬜ | 30m | Review magic value organization across packages |
| 22 | Update config package | ⬜ | 30m | Ensure config supports multi-product deployments |
| 23 | Update telemetry package | ⬜ | 30m | Ensure telemetry supports multi-product deployments |

### Phase E: Deployments Refactoring (Tasks 24-28)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 24 | Rename deployments/kms/ | ⬜ | 15m | Verify kms/ deployment structure is correct |
| 25 | Consolidate compose files | ⬜ | 30m | Review compose file organization |
| 26 | Update Dockerfile references | ⬜ | 30m | Update Dockerfile paths if needed |
| 27 | Create deployments/jose/ | ⬜ | 15m | Create placeholder for JOSE deployment |
| 28 | Create deployments/ca/ | ⬜ | 15m | Create placeholder for CA deployment |

### Phase F: Validation & Cleanup (Tasks 29-30)

| # | Task | Status | LOE | Description |
|---|------|--------|-----|-------------|
| 29 | Run full test suite | ⬜ | 15m | Verify all tests pass after refactoring |
| 30 | Update PROJECT-STATUS.md | ⬜ | 15m | Document refactoring completion |

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
**Target Completion**: 2025-12-02

| Phase | Tasks | Completed | Status |
|-------|-------|-----------|--------|
| A: Planning | 5 | 2 | 🔄 In Progress |
| B: API | 5 | 0 | ⬜ Not Started |
| C: KMS | 8 | 0 | ⬜ Not Started |
| D: Infra | 5 | 0 | ⬜ Not Started |
| E: Deploy | 5 | 0 | ⬜ Not Started |
| F: Validate | 2 | 0 | ⬜ Not Started |
| **Total** | **30** | **2** | **7%** |

---

## Notes

- KMS is working extremely well - be careful not to break it
- Focus on directory structure, not major code changes
- Regular commits after each logical group of changes
- Run tests frequently to catch regressions early

---

*Sprint Created: 2025-12-02*
