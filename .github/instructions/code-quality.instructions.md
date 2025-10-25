---
description: "Instructions for code quality and maintenance standards"
applyTo: "**"
---
# Code Quality Instructions

- Implement proper resource cleanup (defer statements for HTTP bodies, files, etc.)
- Maintain clear function boundaries (avoid high cyclomatic complexity)
- Remove unused code and parameters
- Validate input parameters in mapper and utility functions
- Wrap all external package errors with context using fmt.Errorf and %w verb to satisfy wrapcheck linter
- Use Go context for HTTP requests and long-running operations to satisfy noctx linter (http.NewRequestWithContext, t.Context() in tests)
- Follow maintenance guidelines in files: immediately remove completed/obsolete tasks from actionable lists

## Pre-commit Hook Documentation Maintenance

**CRITICAL**: When modifying any of these files, review and update `docs/pre-commit-hooks.md`:
- `.pre-commit-config.yaml` - Hook ordering, tool configuration, exclusions
- `.golangci.yml` - Linter configuration, enabled/disabled linters, severity settings
- `.gofumpt.toml` - Go formatting rules and module path
- `.gremlins.yaml` - Mutation testing configuration (if it affects pre-commit indirectly)
- `scripts/cicd_checks.go` - Test pattern enforcement logic
- `scripts/setup-pre-commit.ps1` / `setup-pre-commit.bat` - Setup script changes
- `.vscode/settings.json` - VS Code integration with pre-commit tools

**What to update in docs/pre-commit-hooks.md**:
- Tool ordering changes in pipeline flow diagram and stage breakdown table
- New/removed tools in configuration sections
- Parameter changes in tool-specific sections (args, exclusions, files patterns)
- Integration changes (VS Code settings, golangci-lint config references)
- New troubleshooting guidance for common issues

## Linter Compliance

### Automatic Fixing with --fix

**ALWAYS use golangci-lint --fix as the primary auto-fix tool** - it runs all auto-fixable linters including formatting, imports, and code quality fixes in one command:

```bash
# Fix all auto-fixable linting issues across the entire codebase
golangci-lint run --fix

# Fix specific linters only
golangci-lint run --enable-only=wsl,gofmt,goimports,gofumpt --fix

# Fix issues in specific files
golangci-lint run --fix path/to/file.go
```

**The pre-commit hooks automatically run `golangci-lint run --fix`** so formatting and imports are handled during commit. Manual usage is only needed for:
- Quick fixes during development
- Testing specific linter configurations
- Debugging linting issues

### CRITICAL: wsl and gofumpt Conflict Resolution

**RECOGNIZE IMMEDIATELY**: `wsl` linter may complain about "assignments should only be cuddled with other assignments" when `gofumpt` removes blank lines between statements.

**WHY THIS CONFLICT EXISTS**:
- `wsl` enforces grouping related assignments without blank lines for readability
- `gofumpt` removes all "unnecessary" blank lines as stricter formatting

**PREFERRED SOLUTIONS** (in order):
1. **Restructure code** to avoid the conflict (preferred)
2. **Accept gofumpt formatting** and suppress wsl with `//nolint:wsl`
3. **Use alternative formatting** that satisfies both tools

**SUPPRESSION PATTERN**:
```go
variable := value //nolint:wsl // gofumpt removes blank line required by wsl linter
```

**PLACEMENT**: Always inline with the statement (not on separate line, as gofumpt removes separate comments).

**WHEN TO USE**: Only when code restructuring would harm readability or when the conflict is unavoidable. Avoid unnecessary nolint usage.

### CRITICAL: godot Comment Period Requirements

**ALWAYS end comments with periods** to satisfy the godot linter:

- **Package comments**: `// Package cryptoutil provides cryptographic utilities.`
- **Function comments**: `// NewCipher creates a new cipher instance.`
- **Variable comments**: `// defaultTimeout is the default request timeout.`
- **Constant group comments**: `// HTTP status codes.`
- **Struct field comments**: `// Name is the user's full name.`

**Pattern for constant groups:**
```go
const (
    // HTTP status codes.
    statusOK    = 200
    statusError = 500

    // Database connection settings.
    maxConnections = 10
    timeoutSeconds = 30
)
```

**NEVER use incomplete sentences** in comments - always ensure they form complete thoughts ending with periods.

**Linters that support automatic fixing:**
- **wsl**: Whitespace consistency (blank lines between statements)
- **gofmt**: Go code formatting
- **goimports**: Import organization and formatting
- **godot**: Adds missing periods to documentation comments
- **goconst**: Creates named constants for repeated strings
- **importas**: Fixes import aliases to match configured rules
- **copyloopvar**: Fixes loop variable capture issues
- **testpackage**: Renames test packages to follow conventions
- **revive**: Fixes various style and code quality issues

**Linters that require manual fixing:**
- **errcheck**: Always check error return values from functions. Never ignore errors with `_` unless explicitly documented why the error can be safely ignored. Don't use `//nolint:errcheck` to suppress legitimate error handling requirements
- **gosimple**: Suggest code simplifications (manual review required)
- **govet**: Reports suspicious constructs (manual review required)
- **ineffassign**: Detect ineffectual assignments (manual review required)
- **staticcheck**: Advanced static analysis checks (manual review required)
- **unused**: Find unused constants, variables, functions and types (manual review required)
- **gosec**: Security-focused static analysis (manual review required). **USE THIS INSTEAD OF detect-secrets** - detect-secrets is not installed and incorrectly flags legitimate test code referencing cryptographic keys/materials

### detect-secrets Inline Allowlisting

**WHEN USING detect-secrets (despite recommendation to use gosec instead)**, use inline allowlisting comments to suppress false positives:

**Preferred: Same-line allowlisting (recommended for Go):**
```go
intermediateCASubject, err := CreateCASubject(rootCASubject, rootCASubject.KeyMaterial.PrivateKey, "Round Trip Intermediate CA", subjectsKeyPairs[1], 10*365*cryptoutilDateTime.Days1, 1) // pragma: allowlist secret
```

**Alternative: Next-line allowlisting:**
```go
// pragma: allowlist nextline secret
const secret = "hunter2";
```

**Supported comment styles:**
- `# pragma: allowlist secret` (Python/shell)
- `// pragma: allowlist secret` (Go/JavaScript/C++) - **PREFERRED for Go code**
- `/* pragma: allowlist secret */` (multi-line comments)

**WHEN TO USE**: Only for legitimate test code or configuration that must reference cryptographic materials. Prefer gosec over detect-secrets for security scanning.
- **noctx**: Check for missing context in HTTP requests (manual review required)
- **wrapcheck**: Check error wrapping consistency (manual review required)
- **thelper**: Check test helper functions (manual review required)
- **tparallel**: Check for parallel test issues (manual review required)
- **gomodguard**: Prevent importing blocked modules (manual review required)
- **prealloc**: Find slice declarations that could pre-allocate (manual review required)
- **bodyclose**: Check for HTTP response body closure (manual review required)
- **errorlint**: Find code that will cause problems with Go 1.13 error wrapping (manual review required)
- **stylecheck**: Go style guide compliance (manual review required)

## Code Patterns

- **Default Values**: Always declare default values as named variables (e.g., `var defaultConfigFiles = []string{}`) rather than inline literals, following the established pattern in config.go
- **Pass-through Calls**: When making pass-through calls to helper functions, prefer using the same parameter order and return value order as the helper function to maintain API consistency and reduce confusion
