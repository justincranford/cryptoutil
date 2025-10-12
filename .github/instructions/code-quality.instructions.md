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

- **godot**: All comments must end with a period (`.`). This includes package comments, function comments, and inline comments
- **goconst**: Avoid repeating string literals. Use named constants for strings that appear 3+ times in the same file
- **errcheck**: Always check error return values from functions. Never ignore errors with `_` unless explicitly documented why the error can be safely ignored. Don't use `//nolint:errcheck` to suppress legitimate error handling requirements

## Code Patterns

- **Default Values**: Always declare default values as named variables (e.g., `var defaultConfigFiles = []string{}`) rather than inline literals, following the established pattern in config.go
