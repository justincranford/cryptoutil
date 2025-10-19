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

## Test File Organization

Follow Go testing file naming conventions for proper organization:

| Test Type | File Suffix | Purpose | Example |
|-----------|-------------|---------|---------|
| Unit Tests | `_test.go` | Blackbox/whitebox testing of functions | `calculator_test.go` |
| Benchmarks | `_bench_test.go` | Performance testing | `calculator_bench_test.go` |
| Fuzz Tests | `_fuzz_test.go` | Property-based testing | `calculator_fuzz_test.go` |
| Integration | `_integration_test.go` | End-to-end with real dependencies | `api_integration_test.go` |

**File Separation Rules:**
- Keep unit tests, benchmarks, and fuzz tests in separate files
- Use descriptive names that indicate the test focus
- Group related tests by functionality within each file type

- When creating test names, test code, test utilities, and test data (e.g. unique database names), ensure they are concurrency safe for parallel testing; use UUIDv7 suffixes for uniqueness instead of counters or timestamps which can reset or collide during parallel execution
- UUIDv7 provides time-ordered uniqueness and randomness, making it ideal for concurrent testing while maintaining deterministic ordering for debugging
- Design tests, utilities, and data for robustness: avoid brittleness and brittle patterns that could cause intermittent failures
