---
description: "Instructions for testing"
applyTo: "**"
---
# Testing Instructions

- Run automated tests with `go test ./... -cover`
- Use test utilities in keygenpooltest for pool and key generation tests
- Use testify require methods for assertions (require.NoError, require.True, require.Equal, etc.)
- Use manual tests via Swagger UI (`go run main.go` and open http://localhost:8080/swagger)
- Ensure coverage for all key types and pool configurations
- Always use `docker compose` form (not `docker-compose`) when running with Docker Compose
- When making code changes, ensure all tests are updated and working before completion
- When making code changes, run linters and fix all issues before completion (`golangci-lint run`)
- Run `golangci-lint run` before committing changes
- Address any new violations immediately

## Script Testing Requirements

- **ALWAYS test scripts after creation or modification** before declaring them complete
- **NEVER assume scripts work without verification** - syntax errors are common and must be caught
- Test PowerShell scripts with execution policy bypass: `powershell -ExecutionPolicy Bypass -File script.ps1 -Help`
- Test Bash scripts with: `bash script.sh --help` or `./script.sh --help`
- Verify help output displays correctly and all command-line parameters work as expected
- **Test functional execution paths**: Run scripts with different parameter combinations to verify actual functionality, not just help text
- **Validate output quality**: Ensure scripts produce expected files, reports, or results in the correct format
- **Test error conditions**: Verify scripts handle missing dependencies, invalid parameters, and edge cases gracefully
- **Cross-platform validation**: Test both PowerShell and Bash versions when creating cross-platform scripts
- Fix all syntax errors, parsing issues, and runtime errors before completion
- If script modifications are made, re-test immediately to verify fixes work
- **Document testing results**: Include evidence of successful testing in commit messages or documentation

## Test Constants and Code Quality

- Use constants for repeated string values in tests when they improve readability and maintainability
- Consider appending randomness to test constants to ensure test isolation (e.g., `testPrefix + uuid.New().String()`)
- Prefer meaningful test data over generic constants when it makes tests more understandable
- Balance DRY principles with test clarity - sometimes duplication in tests is acceptable for readability
- Ensure all test resources are properly cleaned up
