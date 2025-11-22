# Shared Utilities Extraction Plan

## Overview

This document audits code duplication across the cryptoutil repository, identifies utilities suitable for `internal/common` or `pkg` packages, and provides a staged extraction plan with risk assessment.

**Cross-references:**
- [Dependency Analysis](./dependency-analysis.md) - Identifies coupling risks (KMS → common utilities)
- [Group Directory Blueprint](./blueprint.md) - Defines target package locations
- [Import Alias Policy](./import-aliases.md) - Import alias conventions for new packages

---

## Current Utility Organization

### internal/common/ Structure

```
internal/common/
├── apperr/              # Application error definitions (stays)
├── config/              # Configuration loading (stays)
├── container/           # Dependency injection (MOVE to kms/container)
├── crypto/              # Cryptographic operations (SPLIT: kms + pkg)
│   ├── asn1/           → pkg/crypto/asn1 (general-purpose)
│   ├── certificate/    → pkg/crypto/certificate (general-purpose)
│   ├── digests/        → pkg/crypto/digests (general-purpose)
│   ├── jose/           → kms/crypto/jose (KMS-specific)
│   ├── keygen/         → pkg/crypto/keygen (general-purpose)
│   └── keygenpooltest/ → DELETE (test-only, not production code)
├── magic/              # Magic constants (stays, grows with new groups)
├── pool/               # Worker pools (MOVE to kms/pool)
├── telemetry/          # Observability (MOVE to kms/telemetry)
├── testutil/           # Test utilities (stays)
└── util/               # General utilities (AUDIT for duplication)
    ├── combinations/   # Permutation generation (stays - general-purpose)
    ├── datetime/       # Time utilities (stays - general-purpose)
    ├── files/          # File operations (stays - used by CICD + tests)
    ├── network/        # Network utilities (stays - used by server + tests)
    ├── sysinfo/        # System information (stays - general-purpose)
    └── thread/         # Concurrency helpers (stays - general-purpose)
```

---

## Duplication Audit Results

### Pattern 1: UUID Generation Duplication

**Location 1:** `internal/common/util/uuid.go`
```go
func GenerateUUIDv7() (*googleUuid.UUID, error) {
    uuid, err := googleUuid.NewV7()
    if err != nil {
        return nil, fmt.Errorf("failed to generate UUIDv7: %w", err)
    }
    return &uuid, nil
}
```

**Location 2:** `internal/identity/domain/nullable_uuid.go`
```go
// Similar UUID generation with null handling
func NewNullableUUID() NullableUUID {
    uuid, _ := googleUuid.NewV7()
    return NullableUUID{UUID: uuid, Valid: true}
}
```

**Assessment:**
- **Duplication level:** Moderate (different null-handling semantics)
- **Action:** Keep separate - identity domain needs NullableUUID wrapper
- **Rationale:** NullableUUID is domain-specific (identity only), not general-purpose

---

### Pattern 2: Error Wrapping Duplication

**Location 1:** `internal/server/repository/orm/business_entities.go`
```go
func (r *ORMRepository) toAppErr(err error, operation string) error {
    if err == nil {
        return nil
    }
    // ... extensive error mapping logic ...
}
```

**Location 2:** `internal/identity/repository/orm/user_repository.go`
```go
func toAppErr(err error, operation string) error {
    if err == nil {
        return nil
    }
    // ... similar error mapping logic ...
}
```

**Assessment:**
- **Duplication level:** High (very similar GORM error → HTTP status code mapping)
- **Action:** Extract to `internal/common/repository/gorm_errors.go`
- **Rationale:** Both KMS and Identity use GORM with similar error handling patterns
- **Risk:** Medium (changes error handling behavior if extraction introduces bugs)

**Extraction plan:**
```go
// internal/common/repository/gorm_errors.go

// ToAppErr maps GORM errors to application errors with HTTP status codes
func ToAppErr(err error, operation string, entityType string) error {
    if err == nil {
        return nil
    }

    // Check for GORM-specific errors
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return &apperr.Error{
            Type:    apperr.NotFound,
            Message: fmt.Sprintf("%s not found: %s", entityType, operation),
            Cause:   err,
        }
    }

    // Duplicate key violation (both PostgreSQL and SQLite)
    if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
        return &apperr.Error{
            Type:    apperr.Conflict,
            Message: fmt.Sprintf("%s already exists: %s", entityType, operation),
            Cause:   err,
        }
    }

    // Foreign key constraint violation
    if strings.Contains(err.Error(), "foreign key constraint") {
        return &apperr.Error{
            Type:    apperr.BadRequest,
            Message: fmt.Sprintf("invalid %s reference: %s", entityType, operation),
            Cause:   err,
        }
    }

    // Default: internal server error
    return &apperr.Error{
        Type:    apperr.Internal,
        Message: fmt.Sprintf("database error during %s: %s", entityType, operation),
        Cause:   err,
    }
}
```

---

### Pattern 3: Configuration Loading Duplication

**Location 1:** `internal/server/config/config.go` (KMS)
```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &cfg, nil
}
```

**Location 2:** `internal/identity/config/config.go` (Identity)
```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &cfg, nil
}
```

**Assessment:**
- **Duplication level:** High (identical code, different Config structs)
- **Action:** Extract to `internal/common/config/loader.go`
- **Rationale:** Generic YAML config loading pattern used by all service groups
- **Risk:** Low (straightforward utility function with no side effects)

**Extraction plan:**
```go
// internal/common/config/loader.go

// LoadYAML loads a YAML configuration file into the provided target struct
func LoadYAML[T any](path string) (*T, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }

    var cfg T
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
    }

    return &cfg, nil
}

// SaveYAML saves a configuration struct to a YAML file
func SaveYAML[T any](path string, cfg *T) error {
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    if err := os.WriteFile(path, data, 0o600); err != nil {
        return fmt.Errorf("failed to write config file %s: %w", path, err)
    }

    return nil
}
```

**Usage (post-extraction):**
```go
// internal/server/config/config.go (KMS)
func LoadConfig(path string) (*Config, error) {
    return config.LoadYAML[Config](path)
}

// internal/identity/config/config.go (Identity)
func LoadConfig(path string) (*Config, error) {
    return config.LoadYAML[Config](path)
}
```

---

### Pattern 4: HTTP Client Creation Duplication

**Location 1:** `internal/test/e2e/http_utils.go`
```go
func CreateInsecureHTTPClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true, // For self-signed certs in tests
            },
        },
        Timeout: 30 * time.Second,
    }
}
```

**Location 2:** Similar patterns in integration tests across packages

**Assessment:**
- **Duplication level:** Low-Medium (test-only code)
- **Action:** Consolidate in `internal/common/testutil/http.go`
- **Rationale:** E2E and integration tests need HTTP clients for self-signed cert testing
- **Risk:** Low (test-only utility, not production code)

**Extraction plan:**
```go
// internal/common/testutil/http.go

// CreateTestHTTPClient creates an HTTP client configured for testing with self-signed certificates
func CreateTestHTTPClient(timeout time.Duration) *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true, // ONLY for tests with self-signed certs
            },
        },
        Timeout: timeout,
    }
}
```

---

### Pattern 5: Context Timeout Duplication

**Location 1:** Multiple files use similar context timeout patterns
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

**Location 2:** Scattered across server, repository, test files

**Assessment:**
- **Duplication level:** Low (standard Go idiom)
- **Action:** No extraction needed - keep inline for clarity
- **Rationale:** Standard context pattern, extraction adds no value
- **Risk:** N/A

---

## Utilities Suitable for pkg/ Promotion

### Crypto Primitives (General-Purpose)

**Promote to pkg/crypto/:**

| Package | Current Location | Target Location | Rationale |
|---------|------------------|-----------------|-----------|
| keygen | `internal/common/crypto/keygen` | `pkg/crypto/keygen` | General key generation (RSA, ECDSA, EdDSA, AES, HMAC) - useful for CA, Identity, external tools |
| digests | `internal/common/crypto/digests` | `pkg/crypto/digests` | Hash functions (SHA-256, SHA-512, BLAKE2b) - useful for password hashing, data integrity |
| asn1 | `internal/common/crypto/asn1` | `pkg/crypto/asn1` | ASN.1 encoding/decoding - useful for certificate parsing, PKI operations |
| certificate | `internal/common/crypto/certificate` | `pkg/crypto/certificate` | X.509 certificate operations - useful for CA, TLS configuration, certificate validation |

**Benefits:**
- External tools can import cryptoutil's crypto primitives without internal dependencies
- CA service will need keygen, certificate, asn1
- Identity may need digests for password hashing (Argon2, bcrypt, scrypt)
- Promotes code reuse across service groups

**Risks:**
- **Medium:** Breaking change for existing code (import path changes)
- **Mitigation:** Use compatibility shims during migration (see blueprint.md Phase 2)

---

### DO NOT Promote (KMS-Specific)

| Package | Current Location | Reason to Keep in internal/ |
|---------|------------------|----------------------------|
| jose | `internal/common/crypto/jose` | KMS-specific JOSE operations (JWE/JWS wrapping for barrier keys) |
| pool | `internal/common/pool` | KMS worker pools for concurrent key generation |
| telemetry | `internal/common/telemetry` | KMS observability patterns (may be KMS-specific) |
| container | `internal/common/container` | KMS dependency injection (service-specific) |

**Target locations:**
- `jose` → `internal/kms/crypto/jose`
- `pool` → `internal/kms/pool`
- `telemetry` → `internal/kms/telemetry`
- `container` → `internal/kms/container`

---

## Utilities to Keep in internal/common/

### Truly Shared (Multi-Service)

| Package | Location | Rationale |
|---------|----------|-----------|
| apperr | `internal/common/apperr` | Application error types used by all service groups (KMS, Identity, CA) |
| config | `internal/common/config` | Configuration primitives (YAML loading, validation) |
| magic | `internal/common/magic` | Magic constants (network, buffers, timeouts) used across services |
| testutil | `internal/common/testutil` | Test helpers (temp files, HTTP clients for tests) |
| util | `internal/common/util` | General utilities (UUID, YAML/JSON, byte operations) |

**Rationale:**
- Used by 2+ service groups
- No service-specific business logic
- Internal-only (not suitable for public API)

---

## Refactoring Backlog

### High Priority (Phase 2 - KMS Extraction)

1. **Extract GORM error mapping** to `internal/common/repository/gorm_errors.go`
   - **Files:** `internal/server/repository/orm/business_entities.go`, `internal/identity/repository/orm/user_repository.go`
   - **Effort:** 2-3 hours
   - **Risk:** Medium (error handling changes)
   - **Dependencies:** None
   - **Tests:** Add unit tests for all error mapping scenarios

2. **Extract generic config loader** to `internal/common/config/loader.go`
   - **Files:** `internal/server/config/config.go`, `internal/identity/config/config.go`
   - **Effort:** 1-2 hours
   - **Risk:** Low (pure function, no side effects)
   - **Dependencies:** None
   - **Tests:** Test YAML loading, validation, file not found scenarios

3. **Consolidate test HTTP client** in `internal/common/testutil/http.go`
   - **Files:** `internal/test/e2e/http_utils.go`, integration test files
   - **Effort:** 1 hour
   - **Risk:** Low (test-only code)
   - **Dependencies:** None
   - **Tests:** Test client creation, timeout configuration

### Medium Priority (Phase 2 - Crypto Promotion)

4. **Promote crypto packages to pkg/**
   - **Packages:** keygen, digests, asn1, certificate
   - **Effort:** 4-6 hours (import updates across codebase)
   - **Risk:** Medium (breaking change, import path updates)
   - **Dependencies:** Requires blueprint.md Phase 2 migration strategy
   - **Tests:** Run full test suite after promotion

5. **Move KMS-specific utilities** to `internal/kms/`
   - **Packages:** jose, pool, telemetry, container
   - **Effort:** 4-6 hours (import updates, tests)
   - **Risk:** Medium (coupling changes)
   - **Dependencies:** Requires KMS domain extraction (Task 11)
   - **Tests:** Run KMS test suite after moves

### Low Priority (Post-Refactor Cleanup)

6. **Audit util/ for additional duplication**
   - **Focus:** uuid.go, random.go, yml_json.go
   - **Effort:** 2-3 hours
   - **Risk:** Low (mostly discovery work)
   - **Dependencies:** None
   - **Outcome:** Identify candidates for future extraction

7. **Delete keygenpooltest/**
   - **Rationale:** Test-only code, not production package
   - **Effort:** 15 minutes
   - **Risk:** None (not used in production)
   - **Dependencies:** Verify no production code imports it
   - **Tests:** Remove from coverage reports

---

## Risk Assessment

### Extraction Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Breaking import paths during extraction | Medium | Use compatibility shims (blueprint.md Phase 2) |
| Changing error handling behavior | Medium | Extensive unit tests before/after extraction |
| Test failures after package moves | Low-Medium | Run full test suite after each move |
| Dependency cycles created by extraction | Low | Follow dependency analysis (dependency-analysis.md) |
| Coverage regression after consolidation | Low | Compare coverage before/after with diff |

### Migration Strategy Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Multiple developers editing same files | Medium | Coordinate via feature branch, PR reviews |
| CI/CD failures during migration | Medium | Use feature flags, incremental rollout |
| Rollback complexity if issues found | High | Tag commits, maintain compatibility shims |
| Documentation drift during changes | Low | Update docs in same PR as code changes |

---

## Staged Extraction Plan

### Stage 1: Extract Shared Utilities (1 week)

**Goal:** Consolidate duplicated code in `internal/common/`

**Tasks:**
1. Extract GORM error mapping → `internal/common/repository/gorm_errors.go`
2. Extract generic config loader → `internal/common/config/loader.go`
3. Consolidate test HTTP client → `internal/common/testutil/http.go`
4. Run full test suite, verify coverage unchanged
5. Commit with message: "refactor: extract shared utilities to internal/common"

**Validation:**
- [ ] `go test ./... --count=1 -timeout=10m` passes
- [ ] `golangci-lint run ./...` passes
- [ ] Coverage diff shows ±0% change
- [ ] All workflows pass via `go run ./cmd/workflow -workflows=all`

---

### Stage 2: Promote Crypto to pkg/ (1 week)

**Goal:** Make general-purpose crypto available as public API

**Tasks:**
1. Create `pkg/crypto/` directory structure
2. Move keygen → `pkg/crypto/keygen`
3. Move digests → `pkg/crypto/digests`
4. Move asn1 → `pkg/crypto/asn1`
5. Move certificate → `pkg/crypto/certificate`
6. Update import paths across codebase (use find-replace with importas validation)
7. Update `.golangci.yml` importas rules
8. Run full test suite, verify coverage unchanged
9. Commit with message: "refactor: promote general-purpose crypto to pkg/"

**Validation:**
- [ ] `go test ./... --count=1 -timeout=10m` passes
- [ ] `golangci-lint run ./...` passes (importas enforcement)
- [ ] Coverage diff shows ±0% change
- [ ] No references to old import paths remain

---

### Stage 3: Extract KMS-Specific Utilities (concurrent with Task 11)

**Goal:** Move KMS-coupled utilities to `internal/kms/`

**Tasks:**
1. Create `internal/kms/` subdirectories (crypto/jose, pool, telemetry, container)
2. Move jose → `internal/kms/crypto/jose`
3. Move pool → `internal/kms/pool`
4. Move telemetry → `internal/kms/telemetry`
5. Move container → `internal/kms/container`
6. Update import paths in KMS code
7. Update `.golangci.yml` importas rules
8. Run KMS test suite
9. Commit with message: "refactor: extract KMS-specific utilities to internal/kms"

**Validation:**
- [ ] `go test ./internal/kms/... --count=1 -timeout=10m` passes
- [ ] `go test ./internal/server/... --count=1 -timeout=10m` passes
- [ ] `golangci-lint run ./...` passes
- [ ] Coverage for KMS packages unchanged

---

## Validation Checklist

### Pre-Extraction

- [ ] Identify all files importing packages to be moved
- [ ] Generate import dependency graph: `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...`
- [ ] Document current test coverage: `go test ./... -coverprofile=test-output/coverage_baseline.out`
- [ ] Tag baseline: `git tag extraction-baseline-$(date +%Y%m%d)`

### During Extraction

- [ ] Move one package at a time
- [ ] Update imports immediately after each move
- [ ] Run package-specific tests after each move
- [ ] Commit after each successful move
- [ ] Update importas rules incrementally

### Post-Extraction

- [ ] Run full test suite: `go test ./... --count=1 -timeout=10m`
- [ ] Compare coverage: baseline vs post-extraction
- [ ] Run all workflows: `go run ./cmd/workflow -workflows=all`
- [ ] Verify linters pass: `golangci-lint run ./...`
- [ ] Update documentation (README, import-aliases.md, blueprint.md)
- [ ] Remove compatibility shims (after 8-week grace period)

---

## Cross-References

- **Dependency Analysis:** [docs/01-refactor/dependency-analysis.md](./dependency-analysis.md)
- **Group Directory Blueprint:** [docs/01-refactor/blueprint.md](./blueprint.md)
- **Import Alias Policy:** [docs/01-refactor/import-aliases.md](./import-aliases.md)
- **KMS Migration Plan:** Task 11 (Code Migration Phase 2)

---

## Notes

- **Duplication is not always bad:** Some duplication is acceptable if extraction adds complexity
- **Context timeout patterns:** Keep inline - standard Go idiom, no value in extraction
- **Test-only code:** Consolidate in `internal/common/testutil/` for reuse
- **GORM error mapping:** High-value extraction - used by both KMS and Identity
- **Config loading:** Generic YAML loader reduces boilerplate across service groups
- **Crypto promotion:** Enables external tools and CA service to use cryptoutil's crypto primitives
