---
description: "Instructions for testing"
applyTo: "**"
---
# Testing Instructions

- Run automated tests with `go test ./... -cover`
- Use test utilities in keygenpooltest for pool and key generation tests
- Use manual tests via Swagger UI (`go run main.go` and open http://localhost:8080/swagger)
- Ensure coverage for all key types and pool configurations
- Always use `docker compose` form (not `docker-compose`) when running with Docker Compose
- When making code changes, ensure all tests are updated and working before completion
- When making code changes, run linters and fix all issues before completion (`golangci-lint run`)
