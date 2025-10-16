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

## Linter Compliance

### Automatic Fixing with --fix

**ALWAYS attempt automatic fixing first** before manually fixing individual issues:

```bash
# Fix all auto-fixable linting issues across the entire codebase
golangci-lint run --fix

# Fix specific linters only
golangci-lint run --enable-only=wsl,gofmt,goimports --fix

# Fix issues in specific files
golangci-lint run --fix path/to/file.go
```

**After fixing golangci-lint errors, ALWAYS run gofumpt to auto-fix formatting:**

```bash
# Format all Go files in the project
gofumpt -w .

# Format specific files
gofumpt -w path/to/file.go
```

### CRITICAL: wsl and gofumpt Conflict Recognition

**RECOGNIZE IMMEDIATELY**: When `wsl` linter complains about "assignments should only be cuddled with other assignments" but `gofumpt` removes blank lines, this is a known conflict:

- `wsl` wants assignments grouped without blank lines for consistency
- `gofumpt` removes "unnecessary" blank lines as stricter formatting
- **SOLUTION**: Use `//nolint:wsl // gofumpt removes blank line required by wsl linter` comment
- **PATTERN**: Place comment inline with the assignment (not on separate line, as gofumpt removes separate-line comments)
- **EXAMPLES**: See `internal/common/util/random.go` and `internal/client/client_test_util.go`

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
- **gosec**: Security-focused static analysis (manual review required)
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
