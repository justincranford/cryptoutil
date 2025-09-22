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

## Test Constants and Code Quality

- Use constants for repeated string values in tests when they improve readability and maintainability
- Consider appending randomness to test constants to ensure test isolation (e.g., `testPrefix + uuid.New().String()`)
- Prefer meaningful test data over generic constants when it makes tests more understandable
- Balance DRY principles with test clarity - sometimes duplication in tests is acceptable for readability
- Ensure all test resources are properly cleaned up
