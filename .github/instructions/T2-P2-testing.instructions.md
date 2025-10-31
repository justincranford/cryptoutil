---
description: "Instructions for testing"
applyTo: "**"
---
# Testing Instructions

- Run `go test ./... -cover` for automated tests
- Use testify require methods for assertions
- Use manual tests via Swagger UI (see README)
- Ensure coverage for all key types and pool configs
- Update/fix tests and run linters before committing (`golangci-lint run`)
- Script testing: always test scripts after add/update tests, verify help/params, test functional/error/cross-platform paths, document results (see README for details)
- Test directories may contain non-Go performance testing tools (e.g., Java Gatling in `/test/gatling/`)
- Use constants for repeated test values if it improves clarity; prefer meaningful test data
- When updating dependencies: run `go test ./...` first to confirm code and tests work before attempting updates; only after tests pass, update one dependency at a time and repeat `go test ./...` to iterate on fixing any issues caused by the update

## Dependency Management Best Practices

- **Automated Checks**: Use `go-update-direct-dependencies` in pre-commit hooks and CI/CD workflows for efficient, focused dependency updates
- **Avoid**: `go-update-all-dependencies` in automated contexts as it can cause unnecessary updates and potential compatibility issues
- **Manual Updates**: Use `go-update-all-dependencies` only for intentional comprehensive dependency refreshes during major version updates or maintenance windows

## Test File Organization

Follow Go testing file naming conventions for proper organization:

| Test Type | File Suffix | Purpose | Example |
|-----------|-------------|---------|---------|
| Unit Tests | `_test.go` | Blackbox/whitebox testing of functions | `calculator_test.go` |
| Benchmarks | `_bench_test.go` | Performance testing | `calculator_bench_test.go` |
| Fuzz Tests | `_fuzz_test.go` | Property-based testing | `calculator_fuzz_test.go` |
| Integration | `_integration_test.go` | Component interaction testing | `api_integration_test.go` |
| E2E | `*_test.go` with `//go:build e2e` | Full system end-to-end testing | `e2e_test.go` |

**File Separation Rules:**
- Keep unit tests, benchmarks, and fuzz tests in separate files
- Use descriptive names that indicate the test focus
- Group related tests by functionality within each file type

- When creating test names, test code, test utilities, and test data (e.g. unique database names), ensure they are concurrency safe for parallel testing; use UUIDv7 suffixes for uniqueness instead of counters or timestamps which can reset or collide during parallel execution
- UUIDv7 provides time-ordered uniqueness and randomness, making it ideal for concurrent testing while maintaining deterministic ordering for debugging
- Design tests, utilities, and data for robustness: avoid brittleness and brittle patterns that could cause intermittent failures

## Copilot Testing Guidelines

When testing linting of code samples or validating regex patterns during chat sessions:

- **Create permanent tests in `internal/cmd/cicd/cicd_test.go`** instead of one-off temporary test files
- This ensures test coverage persists across chat sessions and serves as regression testing
- Examples: regex validation tests, linting pattern tests, code transformation tests
- All tests in `cicd_test.go` execute automatically during Go test runs
- Use descriptive test names that indicate the validation purpose (e.g., `TestEnforceTestPatterns_RegexValidation`)

### cicd Utility Testing Patterns

When adding or updating the cicd utility (`internal/cmd/cicd/cicd.go`):

- **Always implement programmatic tests** in `internal/cmd/cicd/cicd_test.go`
- **Test pattern**: Write generated code to temporary file (use `t.TempDir()`), run lint/check function against it, assert results programmatically
- **Avoid interactive prompts**: This prevents unwanted prompts during Copilot-assisted sessions
- **Avoid ephemeral shell commands**: Don't use piping PowerShell Get-Content replace commands - prefer programmatic edits via tests or scripted tools
- **Rationale**: Ephemeral shell patterns trigger unwanted prompts and premium LLM watch requests
