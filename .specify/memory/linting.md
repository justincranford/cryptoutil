# Linting and Code Quality Standards - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/03-07.linting.instructions.md`

## MANDATORY: Zero Linting Errors Policy

### Enforcement Scope

**ALL code must pass linting - NO EXCEPTIONS**:
- Production code (internal/*, pkg/*)
- Test code (*_test.go, *_bench_test.go, *_fuzz_test.go)
- Demo applications (cmd/demo/*)
- Example code (examples/*)
- Utility scripts (scripts/*, tools/*)
- Configuration files (*.yaml, *.json)

**NEVER use `//nolint:` directives except for documented linter bugs**:
- Documented bug: Must reference GitHub issue in comment
- Example: `//nolint:errcheck // See https://github.com/golangci/golangci-lint/issues/XXXX`

**Rationale**: 
- Linting errors in tests/demos/examples teach bad patterns
- "This is just a demo" is NOT a valid exception
- All code represents project quality standards

---

## Core Quality Rules

### File Encoding

**MANDATORY: UTF-8 without BOM for ALL text files**

**Enforcement**: `cicd all-enforce-utf8` pre-commit hook

```bash
# Check all text files for UTF-8 encoding
go run ./cmd/cicd all-enforce-utf8
```

**Why**:
- BOM (Byte Order Mark) breaks Go compiler on some platforms
- UTF-8 without BOM is universal standard
- Cross-platform consistency (Windows, Linux, macOS)

### Code Style Conventions

**Type Declarations**:
- ✅ Use `any` (not `interface{}`)
- Exception: See `03-01.coding.instructions.md` for format_go self-modification protection

**Formatters**:
- ✅ Use `gofumpt` (stricter than `gofmt`)
- ✅ `golangci-lint run --fix` applies gofumpt automatically

**Indentation**:
- Go code: 4 spaces (tab width = 4)
- YAML/JSON: 2 spaces
- Markdown: 2 spaces for nested lists

---

## golangci-lint v2 Configuration

### Version Requirements

**Current Version**: v2.6.2 (minimum required)
**See**: `02-04.versions.instructions.md` for version consistency requirements

### v2 Breaking Changes

**Configuration Key Changes**:
```yaml
# OLD (v1.x)
linters-settings:
  wsl:
    force-err-cuddling: true

# NEW (v2.x)
linters-settings:
  wsl_v5:  # Note: wsl → wsl_v5
    # force-err-cuddling removed (always enabled)
```

**Removed Settings** (no longer needed):
- `wsl.force-err-cuddling` - Always enabled in v2
- `misspell.ignore-words` - Replaced by allowlist
- `wrapcheck.ignoreSigs` - Replaced by ignorePackageGlobs

**Built-in Formatters**:
- gofumpt integration: `--fix` applies gofumpt automatically
- goimports integration: `--fix` organizes imports automatically

### Running golangci-lint

**ALWAYS run with --fix FIRST**:

```bash
# Step 1: Auto-fix all auto-fixable issues
golangci-lint run --fix

# Step 2: Check for remaining manual fixes
golangci-lint run

# Step 3: Fix ALL issues before committing
git add -A
git commit -m "style: fix linting issues"
```

**Why --fix First**:
- Handles formatting (gofumpt, goimports) automatically
- Fixes auto-fixable linters (wsl, godot, goconst, importas, copyloopvar)
- Reduces manual work (only manual-fix linters remain)

---

## Critical Linter Rules

### wsl (whitespace linter)

**NEVER use `//nolint:wsl` - restructure code instead**

**Rules**:
- Group related statements without blank lines
- Add blank lines between different statement types
- Cuddling rules enforced (related statements stay together)

**Examples**:

```go
// ✅ CORRECT: Related statements grouped
if err := validate(input); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

user, err := repo.Create(ctx, input)
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ WRONG: Blank line separates related error check
if err := validate(input); err != nil {

    return fmt.Errorf("validation failed: %w", err)
}

// ❌ WRONG: No blank line between unrelated statements
if err := validate(input); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
user, err := repo.Create(ctx, input)
```

**Fix Strategy**: Restructure code to group related logic, not suppress linter

### godot (comment period linter)

**ALWAYS end comments with periods**

```go
// ✅ CORRECT: Comment ends with period.
// Package cryptoutil provides encryption utilities.
package cryptoutil

// ✅ CORRECT: Multi-line comment ends with period.
// NewCipher creates a new cipher instance.
// It validates parameters and initializes internal state.
func NewCipher() {}

// ❌ WRONG: Missing period
// Package cryptoutil provides encryption utilities
package cryptoutil

// ❌ WRONG: Multi-line comment missing period
// NewCipher creates a new cipher instance
// It validates parameters and initializes internal state
func NewCipher() {}
```

**Auto-Fix**: `golangci-lint run --fix` adds periods automatically

### mnd (magic number detector)

**Declare magic values as constants**

**Storage Locations**:
- Shared constants: `internal/common/magic/magic_*.go`
- Package-specific constants: `internal/<package>/magic*.go`

**Examples**:

```go
// ✅ CORRECT: Named constant
import cryptoutilMagic "cryptoutil/internal/shared/magic"

server.Start(cryptoutilMagic.DefaultHTTPPort)  // 8080

// ❌ WRONG: Inline magic number
server.Start(8080)
```

**Magic Constant Files**:
- `magic_network.go` - Ports, timeouts, buffer sizes
- `magic_database.go` - Connection pool sizes, query timeouts
- `magic_cryptography.go` - Key sizes, iteration counts
- `magic_testing.go` - Test probabilities, timeouts

**See**: `03-03.golang.instructions.md` for magic values organization

### importas (import alias enforcement)

**Pattern: Enforce consistent import aliases across codebase**

**Configuration** (`.golangci.yml`):

```yaml
linters-settings:
  importas:
    alias:
      # Crypto aliases
      - pkg: crypto/rand
        alias: crand
      # UUID aliases
      - pkg: github.com/google/uuid
        alias: googleUuid
      # JOSE aliases
      - pkg: github.com/go-jose/go-jose/v4
        alias: jose
      - pkg: github.com/go-jose/go-jose/v4/jwt
        alias: joseJwt
```

**See**: `03-03.golang.instructions.md` for complete import alias conventions

---

## Linter Categories

### Auto-Fixable Linters (--fix support)

**Formatters**:
- `gofmt` - Standard Go formatting
- `gofumpt` - Stricter gofmt (preferred)
- `goimports` - Organize imports

**Style Linters**:
- `wsl` - Whitespace linting (cuddling rules)
- `godot` - Comment period enforcement
- `goconst` - Detect repeated strings → constants
- `importas` - Import alias consistency
- `copyloopvar` - Fix loop variable capture issues
- `testpackage` - Enforce `_test` package naming
- `revive` - Configurable linting rules

**Workflow**: Run `golangci-lint run --fix` before committing

### Manual-Fix Linters

**Error Handling**:
- `errcheck` - Check all error returns
- `wrapcheck` - Wrap external errors
- `errorlint` - Error handling best practices

**Code Quality**:
- `gosimple` - Simplify code
- `govet` - Go vet checks
- `ineffassign` - Detect ineffectual assignments
- `staticcheck` - Static analysis
- `unused` - Detect unused code

**Security**:
- `gosec` - Security issues (G401, G501, etc.)

**Testing**:
- `thelper` - Test helper function conventions
- `tparallel` - t.Parallel() usage
- `noctx` - HTTP requests must use context

**Dependencies**:
- `gomodguard` - Block/allow specific dependencies
- `prealloc` - Preallocate slices for performance

**HTTP**:
- `bodyclose` - Ensure HTTP response bodies closed

**Style**:
- `stylecheck` - Go style guide enforcement

**Workflow**: Fix manually, verify with `golangci-lint run`

---

## Domain Isolation Enforcement

### Identity Module Import Restrictions

**Rule**: `internal/identity/*` CANNOT import `internal/server/*`, `internal/client/*`, `api/*`

**Enforcement**: Custom check via cicd tool

```bash
# Verify identity domain isolation
go run ./cmd/cicd go-check-identity-imports
```

**Rationale**:
- Identity is domain layer (business logic, repositories, models)
- Server is presentation layer (HTTP handlers, middleware)
- Client is infrastructure layer (HTTP clients, external APIs)
- Domain layer MUST NOT depend on presentation/infrastructure layers

**Allowed Identity Imports**:
- ✅ `internal/identity/*` → `internal/identity/*` (same domain)
- ✅ `internal/identity/*` → `internal/shared/*` (shared utilities)
- ✅ `internal/identity/*` → `pkg/*` (public libraries)
- ❌ `internal/identity/*` → `internal/server/*` (presentation layer)
- ❌ `internal/identity/*` → `api/*` (generated API code)

---

## Batch Lint Fixing Strategy

### Using multi_replace_string_in_file Tool

**Pattern: Group similar fixes for efficiency**

**Example Batch Operations**:

1. **Copyright Header Fixes** (up to 10 files):
```json
{
  "replacements": [
    {"filePath": "file1.go", "oldString": "...", "newString": "..."},
    {"filePath": "file2.go", "oldString": "...", "newString": "..."}
  ]
}
```

2. **godot Comment Fixes** (up to 10 files):
```json
{
  "replacements": [
    {"filePath": "file1.go", "oldString": "// Comment without period", "newString": "// Comment without period."},
    {"filePath": "file2.go", "oldString": "// Another comment", "newString": "// Another comment."}
  ]
}
```

**Workflow**:
1. Run `golangci-lint run` to identify issues
2. Group similar issues (e.g., all godot fixes, all copyright headers)
3. Apply up to 10 related replacements per tool call
4. Verify with `golangci-lint run --fix` after batch edits
5. Repeat for next batch of issues

---

## Secret Detection

### detect-secrets vs gosec

**Preference**: Use `gosec` (part of golangci-lint)

**detect-secrets** (optional):
- Inline allowlist: `// pragma: allowlist secret`
- Used for additional secret scanning beyond gosec

**gosec** (preferred):
- G401: Detect weak cryptographic algorithms (MD5, DES)
- G501: Import blocklist (blacklist packages)
- G505: Detect weak random number generation (math/rand)

**Example Suppressions** (gosec):

```go
// #nosec G401 -- MD5 used for non-cryptographic checksums only
hash := md5.New()

// WRONG: Suppressing security issues without justification
// #nosec G401
hash := md5.New()
```

**Rationale**: Suppressions MUST include justification

---

## Pre-commit Hook Documentation Maintenance

### Files Requiring Doc Updates

**When modifying these files, update `docs/pre-commit-hooks.md`**:

1. `.pre-commit-config.yaml` - Hook configuration
2. `.golangci.yml` - Linter settings
3. `internal/cmd/cicd/cicd.go` - CICD tool logic

**Documentation Includes**:
- Hook descriptions and purpose
- Linter rules and enforcement
- Troubleshooting common issues
- Version compatibility notes

**Update Workflow**:
1. Modify `.pre-commit-config.yaml` or `.golangci.yml`
2. Update `docs/pre-commit-hooks.md` with changes
3. Commit both files together

**Example Commit**:

```bash
git add .pre-commit-config.yaml docs/pre-commit-hooks.md
git commit -m "ci: update pre-commit hooks and documentation"
```

---

## Key Takeaways

1. **Zero Exceptions**: ALL code must pass linting (production, tests, demos, examples)
2. **Always Run --fix First**: `golangci-lint run --fix` before manual fixes
3. **wsl Restructure**: Never suppress wsl, restructure code to group related logic
4. **godot Periods**: Always end comments with periods (auto-fixable)
5. **Magic Constants**: Declare in `internal/common/magic/` or package-specific `magic*.go`
6. **Domain Isolation**: Identity cannot import server/client/api layers (enforced by cicd check)
7. **Batch Fixes**: Use `multi_replace_string_in_file` for efficiency (up to 10 similar fixes)
8. **UTF-8 without BOM**: All text files must be UTF-8 encoded without Byte Order Mark
