# CICD-Enforceable Directive Analysis

**Date:** 2026-01-31
**Purpose:** Identify copilot instruction directives that can be deterministically enforced via `internal/cmd/cicd` formatters/linters

---

## Executive Summary - Numbered Enforcement Opportunities

### Existing CICD Capabilities (Already Implemented)

| # | Directive | CICD Command | Linter/Formatter | Instruction File |
|---|-----------|--------------|------------------|------------------|
| 1 | CGO_ENABLED=0 (CGO-free SQLite) | `lint-go` | `cgo-free-sqlite` | 03-03.golang, 03-04.database |
| 2 | FIPS-approved algorithms only | `lint-go` | `non-fips-algorithms` | 02-07.cryptography |
| 3 | Cryptoutil import aliases | `lint-go` | `no-unaliased-cryptoutil-imports` | 03-03.golang |
| 4 | `time.Now().UTC()` enforcement | `format-go` | `enforce-time-now-utc` | 03-02.testing |
| 5 | `any` over `interface{}` | `format-go` | `enforce-any` | 03-01.coding |
| 6 | Loop variable capture | `format-go` | `copyloopvar` | 03-01.coding |
| 7 | `t.Helper()` in test helpers | `format-go-test` | `thelper` | 03-02.testing |
| 8 | UUIDv7 in tests | `lint-go-test` | `test-patterns` | 03-02.testing |
| 9 | Bind address safety (127.0.0.1) | `lint-go-test` | `bind-address-safety` | 03-06.security |
| 10 | Admin port exposure (9090) | `lint-compose` | (built-in) | 04-02.docker |
| 11 | UTF-8 encoding | `lint-text` | `utf8` | 03-07.linting |
| 12 | GitHub Actions validation | `lint-workflow` | `github-actions` | 04-01.github |
| 13 | Outdated dependencies | `lint-go-mod` | `outdated-deps` | 02-04.versions |
| 14 | Circular dependencies | `lint-go` | `circular-deps` | 03-03.golang |

### NEW Enforcement Opportunities (Proposed)

| # | Directive | Proposed Command | Proposed Linter | Instruction File | Priority |
|---|-----------|------------------|-----------------|------------------|----------|
| 15 | Docker secrets pattern | `lint-compose` | `docker-secrets` | 03-06.security | HIGH |
| 16 | `require` over `assert` in tests | `lint-go-test` | `testify-require` | 03-02.testing | HIGH |
| 17 | `t.Parallel()` in tests | `lint-go-test` | `t-parallel` | 03-02.testing | HIGH |
| 18 | Table-driven test pattern | `lint-go-test` | `table-driven-tests` | 03-02.testing | MEDIUM |
| 19 | Hardcoded test passwords | `lint-go-test` | `no-hardcoded-passwords` | 02-02.service-template | HIGH |
| 20 | `crypto/rand` over `math/rand` | `lint-go` | `crypto-rand` | 02-07.cryptography | HIGH |
| 21 | File size limits (500 lines) | `lint-go` | `file-size-limit` | 03-01.coding | MEDIUM |
| 22 | Magic values in magic/ | `lint-go` | `magic-location` | 03-03.golang | LOW |
| 23 | No `localhost` bind in Go | `lint-go` | `explicit-ipv4-bind` | 03-06.security | MEDIUM |
| 24 | TLS 1.3+ minimum | `lint-go` | `tls-version` | 02-09.pki | MEDIUM |
| 25 | Test file size limits | `lint-go-test` | `test-file-size` | 03-02.testing | LOW |
| 26 | No inline env vars in compose | `lint-compose` | `no-inline-env` | 04-02.docker | HIGH |
| 27 | GORM over raw database/sql | `lint-go` | `gorm-required` | 03-04.database | LOW |
| 28 | No `InsecureSkipVerify: true` | `lint-go` | `tls-verify` | 02-09.pki | HIGH |

---

## Detailed Analysis by Instruction File

### 01-01.terminology.instructions.md

**Enforceable Directives**: None (terminology reference only)

---

### 01-02.beast-mode.instructions.md

**Enforceable Directives**: None (behavioral directive for AI agents)

---

### 02-01.architecture.instructions.md

**Enforceable Directives**: None (architecture reference)

---

### 02-02.service-template.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| No hardcoded passwords | ✅ Yes | ❌ No | `lint-go-test: no-hardcoded-passwords` |
| TestMain pattern usage | ⚠️ Partial | ❌ No | Heuristic detection possible |
| E2E testing helpers | ❌ No | - | Documentation only |

**Proposed Enhancement #19**: Detect hardcoded passwords in test files
```go
// Patterns to detect:
// password := "test123"
// Password: "hardcoded"
// secret := "mysecret"
```

---

### 02-03.https-ports.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Port 0 in tests | ✅ Yes | ✅ Yes | `lint-go-test: bind-address-safety` |
| 127.0.0.1 bind in tests | ✅ Yes | ✅ Yes | `lint-go-test: bind-address-safety` |
| No 0.0.0.0 in non-container code | ✅ Yes | ✅ Yes | `lint-go-test: bind-address-safety` |

**Status**: Already fully enforced by existing linter

---

### 02-04.versions.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Go version consistency | ✅ Yes | ⚠️ Partial | `lint-go-mod: outdated-deps` |
| Dependency updates | ✅ Yes | ✅ Yes | `lint-go-mod: outdated-deps` |

**Status**: Partially enforced - could add go.mod version consistency check

---

### 02-05.observability.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| OTLP configuration | ❌ No | - | Runtime config, not static |
| Structured logging | ⚠️ Partial | ❌ No | Could detect unstructured logging |

**Recommendation**: No immediate enforcement needed

---

### 02-06.openapi.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| OpenAPI 3.0.3 version | ✅ Yes | ❌ No | `lint-openapi` (new command) |
| Strict server generation | ⚠️ Partial | ❌ No | Check oapi-codegen config |

**Recommendation**: LOW priority - OpenAPI validation better handled by dedicated tools

---

### 02-07.cryptography.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| FIPS-approved algorithms | ✅ Yes | ✅ Yes | `lint-go: non-fips-algorithms` |
| crypto/rand over math/rand | ✅ Yes | ❌ No | `lint-go: crypto-rand` |
| No bcrypt/scrypt/Argon2 | ✅ Yes | ✅ Yes | `lint-go: non-fips-algorithms` |

**Proposed Enhancement #20**: Detect `math/rand` imports in crypto-sensitive code
```go
// Detect:
// import "math/rand"
// rand.Read() from math/rand
```

---

### 02-08.hashes.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| PBKDF2/HKDF selection | ❌ No | - | Runtime decision |
| Version-prefixed hashes | ❌ No | - | Runtime format |

**Recommendation**: No static enforcement possible

---

### 02-09.pki.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| TLS 1.3+ minimum | ✅ Yes | ❌ No | `lint-go: tls-version` |
| No InsecureSkipVerify | ✅ Yes | ❌ No | `lint-go: tls-verify` |
| CA/Browser Forum compliance | ❌ No | - | Certificate content validation |

**Proposed Enhancement #24**: Detect TLS version configuration below 1.3
```go
// Detect:
// MinVersion: tls.VersionTLS12
// tls.VersionTLS10, tls.VersionTLS11
```

**Proposed Enhancement #28**: Detect InsecureSkipVerify: true
```go
// Detect:
// InsecureSkipVerify: true
```

---

### 02-10.authn.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Session storage (SQL only) | ❌ No | - | Architecture decision |
| MFA step-up timing | ❌ No | - | Runtime logic |

**Recommendation**: No static enforcement possible

---

### 03-01.coding.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| File size limits (500 lines) | ✅ Yes | ❌ No | `lint-go: file-size-limit` |
| Default values as named vars | ⚠️ Partial | ❌ No | Heuristic only |
| `any` over `interface{}` | ✅ Yes | ✅ Yes | `format-go: enforce-any` |
| format_go self-protection | ✅ Yes | ✅ Yes | Built into enforce-any |

**Proposed Enhancement #21**: Enforce 500-line file limit
```go
// Error if file exceeds 500 lines (configurable)
```

---

### 03-02.testing.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Table-driven tests | ✅ Yes | ❌ No | `lint-go-test: table-driven-tests` |
| app.Test() for handlers | ⚠️ Partial | ❌ No | Detect httptest.Server in tests |
| TestMain pattern | ⚠️ Partial | ❌ No | Heuristic detection |
| t.Parallel() everywhere | ✅ Yes | ❌ No | `lint-go-test: t-parallel` |
| UUIDv7 in tests | ✅ Yes | ✅ Yes | `lint-go-test: test-patterns` |
| require over assert | ✅ Yes | ❌ No | `lint-go-test: testify-require` |
| time.Now().UTC() | ✅ Yes | ✅ Yes | `format-go: enforce-time-now-utc` |
| t.Helper() in helpers | ✅ Yes | ✅ Yes | `format-go-test: thelper` |
| No hardcoded UUIDs | ✅ Yes | ✅ Yes | `lint-go-test: test-patterns` |
| Test file size | ✅ Yes | ❌ No | `lint-go-test: test-file-size` |
| Coverage targets | ❌ No | - | CI/CD workflow check |
| Mutation testing | ❌ No | - | Separate gremlins tool |

**Proposed Enhancement #16**: Enforce `require` over `assert`
```go
// Detect:
// assert.NoError(t, err)  → require.NoError(t, err)
// assert.Equal(t, a, b)   → require.Equal(t, a, b)
```

**Proposed Enhancement #17**: Enforce `t.Parallel()` in test functions
```go
// Detect test functions missing t.Parallel()
func TestSomething(t *testing.T) {
    // Missing t.Parallel() at start
}
```

**Proposed Enhancement #18**: Detect non-table-driven test patterns
```go
// Detect multiple TestX_Variant functions that should be consolidated
func TestCreate_Success(t *testing.T) { ... }
func TestCreate_Failure(t *testing.T) { ... }
func TestCreate_Empty(t *testing.T) { ... }
// → Should be single TestCreate with table
```

---

### 03-03.golang.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| CGO_ENABLED=0 | ✅ Yes | ✅ Yes | `lint-go: cgo-free-sqlite` |
| Import aliases | ✅ Yes | ✅ Yes | `lint-go: no-unaliased-cryptoutil-imports` |
| Magic values location | ⚠️ Partial | ❌ No | `lint-go: magic-location` |
| Circular dependencies | ✅ Yes | ✅ Yes | `lint-go: circular-deps` |

**Proposed Enhancement #22**: Detect magic values outside magic/ directories
```go
// Detect hardcoded values that should be in magic/:
// const defaultPort = 8080
// var maxRetries = 3
```

---

### 03-04.database.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| GORM over raw sql | ⚠️ Partial | ❌ No | `lint-go: gorm-required` |
| UUID as text type | ⚠️ Partial | ❌ No | Struct tag analysis |
| serializer:json | ⚠️ Partial | ❌ No | Struct tag analysis |
| CGO-free SQLite | ✅ Yes | ✅ Yes | `lint-go: cgo-free-sqlite` |

**Proposed Enhancement #27**: Warn on raw database/sql imports outside KMS
```go
// Detect in non-KMS packages:
// import "database/sql"
// sql.Open(...)
```

---

### 03-05.sqlite-gorm.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| modernc.org/sqlite driver | ✅ Yes | ✅ Yes | `lint-go: cgo-free-sqlite` |
| WAL mode configuration | ❌ No | - | Runtime config |
| MaxOpenConns settings | ❌ No | - | Runtime config |

**Status**: Already enforced by CGO-free SQLite check

---

### 03-06.security.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| 127.0.0.1 in tests | ✅ Yes | ✅ Yes | `lint-go-test: bind-address-safety` |
| Docker secrets pattern | ✅ Yes | ❌ No | `lint-compose: docker-secrets` |
| crypto/rand only | ✅ Yes | ❌ No | `lint-go: crypto-rand` |
| No InsecureSkipVerify | ✅ Yes | ❌ No | `lint-go: tls-verify` |

**Proposed Enhancement #15**: Enforce Docker secrets in compose files
```yaml
# Detect inline credentials:
environment:
  POSTGRES_PASSWORD: mypassword  # ← VIOLATION
  
# Require:
secrets:
  - postgres_password
environment:
  POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
```

**Proposed Enhancement #23**: Detect `localhost` in Go bind addresses
```go
// Detect:
// net.Listen("tcp", "localhost:8080")  → should be 127.0.0.1:8080
```

---

### 03-07.linting.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| UTF-8 encoding | ✅ Yes | ✅ Yes | `lint-text: utf8` |
| Zero linting errors | ✅ Yes | ✅ Yes | golangci-lint |
| Domain isolation | ✅ Yes | ✅ Yes | `lint-go: circular-deps` |

**Status**: Already fully enforced

---

### 03-08.server-builder.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Builder pattern usage | ❌ No | - | Architecture decision |
| Migration versioning | ❌ No | - | File naming convention |

**Recommendation**: No static enforcement possible

---

### 04-01.github.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| GitHub Actions validation | ✅ Yes | ✅ Yes | `lint-workflow: github-actions` |
| Variable expansion ${VAR} | ⚠️ Partial | ❌ No | Regex pattern matching |
| Test-containers pattern | ❌ No | - | Code pattern |

**Status**: Partially enforced by workflow linter

---

### 04-02.docker.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Admin port (9090) not exposed | ✅ Yes | ✅ Yes | `lint-compose` |
| Docker secrets pattern | ✅ Yes | ❌ No | `lint-compose: docker-secrets` |
| No inline env vars | ✅ Yes | ❌ No | `lint-compose: no-inline-env` |
| 127.0.0.1 in containers | ⚠️ Partial | ❌ No | Compose file analysis |

**Proposed Enhancement #26**: Detect inline environment credentials
```yaml
# Detect:
environment:
  PASSWORD: secret123
  API_KEY: abc123
```

---

### 05-01.cross-platform.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| autoapprove usage | ❌ No | - | Editor-specific |
| Go over PowerShell/Bash | ❌ No | - | Human decision |

**Recommendation**: No static enforcement possible

---

### 05-02.git.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Conventional commits | ⚠️ Partial | ❌ No | Git hook validation |
| Incremental commits | ❌ No | - | Human workflow |

**Recommendation**: Consider pre-commit hook for commit message validation

---

### 05-03.dast.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Variable expansion ${VAR} | ⚠️ Partial | ❌ No | Workflow file analysis |

**Recommendation**: LOW priority - covered in GitHub workflow validation

---

### 06-01.evidence-based.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| Evidence checklist | ❌ No | - | Human verification |
| Quality gates | ⚠️ Partial | - | CI/CD workflow checks |

**Recommendation**: No static enforcement - workflow-level enforcement

---

### 07-01.testmain-integration-pattern.instructions.md

| Directive | Enforceable | Existing | Proposed |
|-----------|-------------|----------|----------|
| TestMain pattern | ⚠️ Partial | ❌ No | Heuristic detection |
| app.Test() usage | ⚠️ Partial | ❌ No | Detect httptest.Server misuse |
| No GORM mocking | ⚠️ Partial | ❌ No | Detect mock struct patterns |

**Recommendation**: MEDIUM priority - complex heuristic analysis needed

---

## Implementation Priority Matrix

### HIGH Priority (Immediate Impact, Clear Detection)

| # | Enhancement | Estimated LOE | Impact |
|---|-------------|---------------|--------|
| 15 | Docker secrets pattern | 2-3 hours | Security compliance |
| 16 | testify require over assert | 1-2 hours | Test reliability |
| 17 | t.Parallel() enforcement | 2-3 hours | Race condition detection |
| 20 | crypto/rand over math/rand | 1 hour | Security |
| 26 | No inline env vars in compose | 1-2 hours | Security |
| 28 | No InsecureSkipVerify | 1 hour | Security |

### MEDIUM Priority (Good Value, Moderate Complexity)

| # | Enhancement | Estimated LOE | Impact |
|---|-------------|---------------|--------|
| 18 | Table-driven test pattern | 3-4 hours | Code quality |
| 19 | No hardcoded passwords | 2 hours | Security |
| 21 | File size limits | 1 hour | Maintainability |
| 23 | Explicit IPv4 bind | 1-2 hours | Cross-platform |
| 24 | TLS version minimum | 1-2 hours | Security |

### LOW Priority (Marginal Benefit, Complex Implementation)

| # | Enhancement | Estimated LOE | Impact |
|---|-------------|---------------|--------|
| 22 | Magic values location | 4+ hours | Organization |
| 25 | Test file size limits | 1 hour | Maintainability |
| 27 | GORM required | 3+ hours | Consistency |

---

## Recommended Next Steps

1. **Phase 1 (Security Focus)**: Implement #15, #20, #26, #28 - All security-related, clear detection
2. **Phase 2 (Testing Quality)**: Implement #16, #17 - Improve test reliability
3. **Phase 3 (Code Quality)**: Implement #18, #19, #21 - Code organization

**Total New Linters Proposed**: 14
**Recommendation**: Enhance existing commands rather than create new commands where possible
