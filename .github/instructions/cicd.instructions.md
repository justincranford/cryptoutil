---
description: "Instructions for CI/CD workflow configuration"
applyTo: ".github/workflows/*.yml"
---
# CI/CD Workflow Instructions

## Go Version Consistency
- **ALWAYS use the same Go version as specified in go.mod** for all CI/CD workflows
- Current project Go version: **1.25.1** (check go.mod file)
- Set `GO_VERSION: '1.25.1'` in workflow environment variables
- Use `go-version: ${{ env.GO_VERSION }}` in setup-go actions

## Version Management
- When updating Go version, update ALL workflow files consistently:
  - `.github/workflows/ci.yml`
  - `.github/workflows/dast.yml`  
  - Any other workflows using Go
- Verify go.mod version matches CI/CD workflows before committing

## Best Practices
- Use environment variables for version consistency across jobs
- Pin to specific patch versions (e.g., '1.25.1', not '1.25' or '^1.25')
- Test locally with the same Go version used in CI/CD
- Update Docker base images to match Go version when applicable
