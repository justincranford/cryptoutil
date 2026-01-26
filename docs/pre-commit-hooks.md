# Pre-Commit Hooks Documentation

This document describes the pre-commit hooks configured for the cryptoutil project.

## Overview

Pre-commit hooks are automated checks that run before git commits to enforce code quality standards, prevent common errors, and maintain consistent formatting.

**Installation**:
```powershell
# Install pre-commit framework
pip install pre-commit

# Install hooks for this repository
pre-commit install

# Run all hooks manually
pre-commit run --all-files
```

## Hook Categories

### 1. Generic File Checks

Lightweight, fast checks for common file issues:

- **end-of-file-fixer**: Ensures files end with newline
- **trailing-whitespace**: Removes trailing whitespace
- **fix-byte-order-marker**: Ensures UTF-8 encoding without BOM
- **check-yaml**: Validates YAML syntax
- **check-json**: Validates JSON syntax (excludes VS Code JSONC files with comments)
- **check-added-large-files**: Prevents files >1MB from being committed
- **check-merge-conflict**: Detects merge conflict markers
- **detect-aws-credentials**: Prevents AWS credential leaks
- **detect-private-key**: Prevents private key commits
- **check-case-conflict**: Detects case-sensitive naming conflicts
- **check-illegal-windows-names**: Validates filenames for Windows compatibility
- **check-toml/xml**: Syntax validation
- **check-symlinks**: Validates symlinks
- **check-executables-have-shebangs**: Ensures scripts have shebangs
- **check-shebang-scripts-are-executable**: Verifies execute permissions
- **pretty-format-json**: Auto-formats JSON files

### 2. Go Linting and Formatting

**golangci-lint** (incremental on commit, full on push):
- Runs 50+ linters including gofumpt, goimports, staticcheck, gosec
- Auto-fixes issues with --fix flag on commit
- Full validation before push
- Configured in .golangci.yml

### 3. Custom CI/CD Checks

**lint-go**:
- Detects circular dependencies
- Enforces FIPS-approved algorithms only
- Checks import restrictions

**lint-compose**:
- Prevents accidental exposure of admin port 9090 in Docker Compose

### 4. Go Formatters (format-go)

Automated code transformations to enforce project standards:

#### 4.1 enforce-any

**Purpose**: Replace interface{} with any for Go 1.18+ readability.

**Pattern**:
```go
// Before
func Process(data interface{}) error { ... }

// After
func Process(data any) error { ...}
```

**Self-Exclusion**: Skips internal/cmd/cicd/format_go/ package to prevent self-modification.

#### 4.2 enforce-time-now-utc

**Purpose**: Replace time.Now() with time.Now().UTC() to prevent SQLite/GORM timezone test failures.

**Pattern**:
```go
// Before
now := time.Now()
if session.ExpiresAt.Before(time.Now()) { ... }

// After
now := time.Now().UTC()
if session.ExpiresAt.Before(time.Now().UTC()) { ... }
```

**Rationale**:
- SQLite stores DATETIME in UTC
- Go's time.Now() respects local timezone (PST, EST, etc.)
- Tests fail when comparing local time against UTC database timestamps
- Formatter enforces UTC standardization across all code

**Implementation**:
- Two-pass AST traversal to detect already-wrapped calls
- Skips time.Now().UTC() (already correct)
- Adds .UTC() suffix to unwrapped time.Now() calls
- Self-exclusion for internal/cmd/cicd/format_go/ package

**Evidence**:
- Auto-applied 1394 fixes across 273 files during initial deployment
- Prevents recurring LLM agent mistakes with timezone handling

**Tests**:
- 8 comprehensive test cases in enforce_time_now_utc_test.go
- Coverage: 92.9% (enforceTimeNowUTC), 94.2% (processGoFileForTimeNowUTC)

**Self-Exclusion**: Skips internal/cmd/cicd/format_go/ to prevent formatter from modifying itself.

**Reference**: See .github/instructions/03-02.testing.instructions.md for complete SQLite DateTime UTC guidelines.

### 5. Test Pattern Enforcement (lint-go-test)

**test-patterns**:
- Enforces table-driven test structure
- Validates TestMain pattern for heavyweight resources
- Checks test file organization

**bind-address-safety**:
- Detects 0.0.0.0 in test files (triggers Windows Firewall prompts)
- Enforces 127.0.0.1 for loopback-only test bindings

### 6. Manual Hooks (--hook-stage manual)

**autoupdate-all-hooks**:
- Updates pre-commit hook versions
- Run manually: pre-commit run autoupdate-all-hooks --hook-stage manual

**lint-go-mod**:
- Checks for Go dependency updates
- Expensive (network calls), run manually before releases

**lint-workflow**:
- GitHub Actions workflow linting
- Only runs when workflow files change

## Hook Execution Order

1. **pre-commit** (on git commit):
   - Generic file checks (fast)
   - golangci-lint --fix (incremental, auto-fixes)
   - Custom CI/CD checks (lint-go, lint-compose)
   - **cicd-enforce-internal** (UTF-8, test patterns, **format-go**)
     - Runs enforce-any (interface{} -> any)
     - Runs enforce-time-now-utc (time.Now() -> time.Now().UTC())
   - go mod tidy
   - TODO severity check

2. **pre-push** (on git push):
   - golangci-lint (full validation, no auto-fix)

## Common Workflows

### First-Time Setup

```powershell
# Install framework
pip install pre-commit

# Install hooks
pre-commit install

# Test on all files
pre-commit run --all-files
```

### Manual Hook Execution

```powershell
# Run specific hook
pre-commit run golangci-lint --all-files
pre-commit run format-go --all-files

# Run manual hooks
pre-commit run autoupdate-all-hooks --hook-stage manual
pre-commit run lint-go-mod --hook-stage manual
```

### Bypass Hooks (Emergency Only)

```powershell
# Skip all hooks (NOT RECOMMENDED)
git commit --no-verify -m "emergency fix"

# Bypass specific hook by editing .pre-commit-config.yaml temporarily
```

## Troubleshooting

### Hooks Run Slow

- Use incremental golangci-lint (default on commit)
- Run pre-commit run --all-files only when needed
- Manual hooks (autoupdate, lint-go-mod) run separately

### Hook Fails with "modified files"

- This is **EXPECTED** behavior for formatter hooks (format-go, golangci-lint --fix)
- Hooks auto-fix files, then return error to prevent accidental commit of unfixed code
- **Solution**: Review auto-fixes, then re-run git add and git commit

### False Positives (bind-address-safety)

- Hook detects 0.0.0.0 in test files
- **Legitimate use cases**: Test table data, config assertion tests
- **Solution**: Verify usage is test data (not actual bind), use --no-verify if confirmed safe

## Maintenance

### Adding New Formatters

1. Create formatter in internal/cmd/cicd/format_go/<formatter_name>.go
2. Register in internal/cmd/cicd/format_go/format_go.go
3. Add tests in <formatter_name>_test.go
4. No changes needed to .pre-commit-config.yaml (format-go hook runs all registered formatters)
5. Update this documentation

### Updating Hook Versions

```powershell
# Auto-update all hooks
pre-commit run autoupdate-all-hooks --hook-stage manual

# Review changes
git diff .pre-commit-config.yaml

# Commit updates
git add .pre-commit-config.yaml
git commit -m "ci(hooks): update pre-commit hook versions"
```

## References

- Pre-commit framework: <https://pre-commit.com/>
- golangci-lint: <https://golangci-lint.run/>
- Project CI/CD tools: internal/cmd/cicd/
- SQLite UTC testing: .github/instructions/03-02.testing.instructions.md
