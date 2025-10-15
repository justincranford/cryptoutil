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

## Configuration Management

### Application Configuration (Production/CI Deployments)
- **ALWAYS use config files** for production and CI application deployments
- **Example**: `cryptoutil server start --config configs/production/config.yml`
- **Database**: Config files should specify actual database connections (PostgreSQL for production)
- **CI/CD Pattern**: Copy and modify base config files for different environments rather than using environment variables
- **Why**: Config files are version-controlled, documented, and prevent environment variable naming mistakes

### Environment Variables (When Necessary)
- **ONLY use for exceptional cases** when Docker/Kubernetes secrets or config files cannot be used
- **NEVER use environment variables for secrets in production** - always prefer Docker secrets or Kubernetes secrets
- Application uses Viper with `CRYPTOUTIL_` prefix: `CRYPTOUTIL_DATABASE_URL`, `CRYPTOUTIL_LOG_LEVEL`, etc.
- **NEVER use non-standard environment variable names** like `POSTGRES_URL` - they will be ignored by the application
- **Use sparingly**: Only for emergency overrides or local development when secrets infrastructure isn't available
- Check `config.go` for the exact setting names and their corresponding environment variables

### Test Configuration (Development/Testing)
- **Tests ALWAYS use SQLite in-memory databases** regardless of config file database-url settings
- When running `cryptoutil server start --dev`, the application automatically switches to SQLite for development/testing
- **Config files in tests**: Can specify any database URL (even PostgreSQL) - tests will ignore it and use SQLite
- **CI/CD test workflows should NOT include PostgreSQL services** - tests use SQLite automatically for isolation and speed
- **Why**: Ensures test isolation, faster execution, and eliminates database setup complexity in CI/CD

## Go Module Caching Best Practices

### Use `cache: true` on `setup-go` Action
- **Preferred**: `cache: true` on `actions/setup-go@v6`
- **Why**: Automatic, self-healing, prevents tar extraction conflicts
- **Avoid**: Manual `actions/cache@v4` for Go modules (brittle, requires workarounds)

### Cache Key Strategy
- Use `go.sum` hash for cache invalidation
- Include OS in key for cross-platform compatibility
- Consider dependency count for large monorepos

### Troubleshooting
- Cache misses: Check `go.sum` changes
- Cache corruption: Let `setup-go` handle it automatically
- Performance issues: Monitor cache hit rates in workflow logs

## Build Flags and Linking

### Static Linking Requirement
- **ALWAYS use static linking** for both CI and Docker builds to ensure maximum portability
- Use `-extldflags '-static'` in ldflags for static linking
- Validate static linking in Docker builds with `ldd` check

### Debug Symbols vs Size Trade-offs
- **Performance and diagnostics prioritized over binary size**
- **CI builds**: Use `-s -extldflags '-static'` (strip symbol table but keep DWARF debug symbols with `-w` removed)
  - Static linking for maximum portability across CI environments
  - Retains debug symbols for troubleshooting test failures and CI diagnostics
  - Smaller than full debug build but still debuggable
- **Docker builds**: Use `-s -extldflags '-static'` (strip symbol table but keep DWARF debug symbols)
  - Static linking for container portability
  - Debug symbols retained for production troubleshooting
- **NEVER use `-w`** in either context (removes DWARF debug symbols, hurts diagnostics)

### Flag Explanations
- `-s`: Strip symbol table (reduces size, keeps DWARF debug symbols)
- `-w`: Strip DWARF debug symbols (breaks debugging, never use)
- `-extldflags '-static'`: Force static linking with external linker
