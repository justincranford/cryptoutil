---
description: "Instructions for testing"
applyTo: "**"
---
# Testing Instructions

- Run automated tests with `go test ./... -cover`.
- Use test utilities in keygenpooltest for pool and key generation tests.
- Use manual tests via Swagger UI (`go run main.go` and open http://localhost:8080/swagger).
- Ensure coverage for all key types and pool configurations.
