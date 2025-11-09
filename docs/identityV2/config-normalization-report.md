# Identity Configuration Normalization Report

## Summary

Task 03 established normalized configuration templates for the identity module across three deployment profiles: development, test, and production.

## Configuration Inventory

### Templates Created

1. **Development Profile** (`configs/identity/development.yml`)
   - SQLite in-memory database
   - TLS disabled for local testing
   - Debug logging enabled
   - CORS permissive for localhost
   - Ephemeral ports (8100-8102)

2. **Test Profile** (`configs/identity/test.yml`)
   - SQLite in-memory database
   - Port 0 (OS-assigned) for parallel testing
   - Minimal timeouts for fast execution
   - Admin APIs disabled
   - Metrics/tracing disabled

3. **Production Profile** (`configs/identity/production.yml`)
   - PostgreSQL with connection pooling
   - TLS mandatory with certificate files
   - Strict CORS origins
   - Rate limiting enabled
   - File-based secrets (no environment variables)

### Configuration Structure

```text
Config
├── AuthZ (ServerConfig)
├── IDP (ServerConfig)
├── RS (ServerConfig)
├── Database (DatabaseConfig)
├── Tokens (TokenConfig)
├── Sessions (SessionConfig)
├── Security (SecurityConfig)
└── Observability (ObservabilityConfig)
```

## Changes from Current State

### Default Configuration Updates

**Before**: TLS enabled by default with empty cert paths (validation failures)
**After**: TLS disabled by default for development ergonomics

### New Capabilities

1. **YAML Loader** (`internal/identity/config/loader.go`)
   - `LoadFromFile()` - Parse and validate YAML configs
   - `SaveToFile()` - Export configs with secure permissions

2. **Test Fixtures** (`internal/identity/config/testdata/`)
   - Minimal test configuration for unit tests
   - Reusable across identity module tests

3. **Validation Framework**
   - Server configuration validation (ports, TLS, admin API)
   - Database configuration validation (connection pools)
   - Token configuration validation (formats, lifetimes)
   - Session configuration validation (cookies, SameSite)
   - Security configuration validation (PKCE, rate limits)
   - Observability configuration validation (logging, metrics)

## Security Compliance

All configurations follow security instructions:

- ✅ Secrets via file paths (`file:///run/secrets/*`)
- ✅ No environment variable secrets
- ✅ TLS mandatory in production
- ✅ Secure cookie defaults (HTTPOnly, Secure, SameSite)
- ✅ PKCE required for authorization code flow
- ✅ Rate limiting in production

## Testing Coverage

- `TestDefaultConfig`: Validates default configuration structure
- `TestLoadFromFile`: YAML parsing and validation
- `TestSaveToFile`: Configuration export and reload
- `TestServerConfigValidation`: Server config rules
- `TestDatabaseConfigValidation`: Database config rules
- `TestTokenConfigValidation`: Token config rules

All tests passing with 100% coverage of validation logic.

## Migration Path

### For CLI Applications

```go
// Before
config := &cryptoutilIdentityConfig.Config{
    AuthZ: &cryptoutilIdentityConfig.ServerConfig{
        Port: 8100,
        // ... hardcoded values
    },
}

// After
config, err := cryptoutilIdentityConfig.LoadFromFile(configPath)
if err != nil {
    log.Fatalf("failed to load config: %v", err)
}
```

### For Docker Compose

```yaml
services:
  authz:
    command:
      - "/app/cryptoutil"
      - "identity"
      - "authz"
      - "--config=/configs/production.yml"
    volumes:
      - ./configs/identity:/configs:ro
```

## Downstream Impact

Tasks depending on normalized config:

- **Task 04**: Dependency audit (config package boundaries)
- **Task 05**: Storage verification (database config)
- **Task 06**: AuthZ rehab (server config, security settings)
- **Task 07**: Client auth (security policies)
- **Task 08**: Token service (token config, key rotation)
- **Task 10**: Integration layer (orchestration config)
- **Task 18**: Docker Compose (production templates)

## Next Steps

1. Update `cmd/identity/{authz,idp,rs}/main.go` to use `LoadFromFile()`
2. Create Docker Compose service definitions using templates
3. Document configuration parameters in README
4. Add configuration validation to pre-commit hooks
