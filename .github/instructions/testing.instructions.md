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
- Use constants for repeated test values if it improves clarity; prefer meaningful test data
