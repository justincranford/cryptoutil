# CICD Formatters and Linters Summary

**Created**: 2026-02-01
**Purpose**: Complete inventory of internal/cmd/cicd/ tools

---

## Directory Structure

```
internal/cmd/cicd/
├── cicd.go                    # Main entry point
├── common/                    # Shared utilities
├── format_go/                 # Go code formatting
├── format_gotest/             # Test code formatting
├── lint_compose/              # Docker Compose linting
├── lint_go/                   # Go code pattern linting
├── lint_golangci/             # .golangci.yml validation
├── lint_gomod/                # go.mod linting
├── lint_gotest/               # Go test pattern linting
├── lint_text/                 # Text file linting
└── lint_workflow/             # GitHub workflow linting
```

---

## NEW in Previous Session (Document Organization and Cleanup Tasks)

### 1. lint_compose/dockersecrets.go
**Purpose**: Detect inline credentials in Docker Compose files

**Patterns Detected**:
- `POSTGRES_PASSWORD:` inline (should use `POSTGRES_PASSWORD_FILE`)
- `API_KEY:` or `SECRET:` inline (should use Docker secrets)
- `PASSWORD=` in environment (should mount from `/run/secrets/`)

**Correct Pattern**:
```yaml
secrets:
  postgres_password:
    file: ./secrets/postgres_password.secret

services:
  app:
    secrets:
      - postgres_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
```

---

### 2. lint_go/cryptopatterns.go
**Purpose**: Detect insecure cryptographic patterns

**Patterns Detected**:
- `math/rand` import (should use `crypto/rand`)
- `InsecureSkipVerify: true` (TLS verification disabled)
- `rand.Seed(` (math/rand seeding)

**Correct Patterns**:
```go
import crand "crypto/rand"  // ✅ Use crypto/rand

tlsConfig := &tls.Config{
    InsecureSkipVerify: false,  // ✅ Always verify certificates
}
```

---

### 3. lint_gotest/requirepatterns.go
**Purpose**: Enforce test code patterns

**Patterns Detected**:
- `assert.` usage (should use `require.`)
- Missing `t.Parallel()` in tests
- Non-table-driven tests with multiple similar functions
- Hardcoded passwords like `password := "test123"`

**Correct Patterns**:
```go
func TestSomething(t *testing.T) {
    t.Parallel()  // ✅ Required
    
    tests := []struct{...}{...}  // ✅ Table-driven
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            require.NoError(t, err)  // ✅ require over assert
        })
    }
}

password := googleUuid.NewV7().String()  // ✅ Dynamic test data
```

---

### 4. lint_golangci/golangci_config.go
**Purpose**: Validate .golangci.yml against v2 schema

**Patterns Detected**:
- v1 config keys (e.g., `wsl:` instead of `wsl_v5:`)
- Deprecated options (`force-err-cuddling`, `ignore-words`)
- Missing required sections

**Correct v2 Structure**:
```yaml
linters-settings:
  wsl_v5:           # ✅ v2 key (not wsl:)
    # force-err-cuddling removed in v2
  misspell:
    locale: US      # ✅ Instead of ignore-words
```

---

## EXISTING CICD Commands (For Holistic Review)

### format_go/
**Files**: `enforce_any.go`, `format.go`, `filter.go`, `time_now_utc.go`
**Purpose**: Format Go code
- Convert `interface{}` to `any`
- Enforce `time.Now().UTC()` pattern
- Self-exclusion patterns for format_go package

### format_gotest/
**Files**: Format test code patterns
**Purpose**: Standardize test code formatting

### lint_compose/
**Files**: `compose.go`, `dockersecrets.go` (NEW)
**Purpose**: Lint Docker Compose files
- Validate compose file structure
- Detect inline credentials

### lint_go/
**Files**: `imports.go`, `cryptopatterns.go` (NEW)
**Purpose**: Lint Go code patterns
- Domain isolation (identity cannot import server/api)
- Crypto pattern enforcement

### lint_golangci/
**Files**: `golangci_config.go` (NEW)
**Purpose**: Validate golangci-lint configuration
- Schema validation
- v1→v2 migration detection

### lint_gomod/
**Files**: `gomod.go`
**Purpose**: Lint go.mod files
- Dependency validation
- Version consistency

### lint_gotest/
**Files**: `requirepatterns.go` (NEW), other test linters
**Purpose**: Lint test patterns
- require vs assert
- t.Parallel enforcement
- Table-driven test enforcement

### lint_text/
**Files**: `utf8.go`, `whitespace.go`
**Purpose**: Lint text files
- UTF-8 without BOM
- Trailing whitespace

### lint_workflow/
**Files**: `workflow.go`
**Purpose**: Lint GitHub workflow files
- YAML validation
- Action version pinning

---

## Proposed: lint_coverage/ (Phase 11)

**Purpose**: Detect leftover coverage files outside test-output/

**Patterns to Detect**:
- `*.out` files in root/internal
- `*coverage*.html` files outside test-output/
- `cover.out` files in source directories

**See**: `quizme-v1.md` for configuration decisions

---

## Running CICD Commands

```bash
# Run all linters
go run ./cmd/cicd lint-all

# Run specific linter
go run ./cmd/cicd lint-go
go run ./cmd/cicd lint-compose
go run ./cmd/cicd lint-gotest
go run ./cmd/cicd lint-golangci

# Run formatters
go run ./cmd/cicd format-go
go run ./cmd/cicd format-gotest
```

---

## Integration Points

- **Pre-commit hooks**: `.pre-commit-config.yaml`
- **CI/CD workflows**: `.github/workflows/ci-quality.yml`
- **Magic constants**: `internal/shared/magic/magic_cicd.go`
