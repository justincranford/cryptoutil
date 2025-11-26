# I1: Configuration Infrastructure

**Component**: Configuration Management
**Current Location**: `internal/*/config/*`, scattered across products
**Target Location**: `internal/infra/configuration`
**Status**: PLANNING

---

## Overview

Configuration infrastructure provides centralized configuration management for all cryptoutil products, supporting multiple configuration sources (files, environment variables, secrets), validation, and runtime reloading.

### Purpose

- **Centralized Configuration**: Single source of truth for application settings
- **Multi-Source Support**: YAML files, environment variables, Docker secrets, Kubernetes ConfigMaps
- **Validation**: Schema validation, type checking, required fields enforcement
- **Security**: Secrets management integration, encryption at rest
- **Runtime Flexibility**: Hot reload support for non-critical settings

---

## Current State Analysis

### Existing Configuration Patterns

**KMS Server** (`internal/server/config/config.go`):
- YAML-based configuration with CLI flag overrides
- Unseal secrets from Docker secrets (`/run/secrets/*`)
- TLS certificate configuration
- Database DSN configuration
- IP allowlisting and CORS configuration
- Structured config validation with defaults

**Identity Server** (`internal/identity/config/config.go`):
- Similar YAML + CLI pattern
- OAuth2.1/OIDC specific configuration
- Client registration and credential management
- Token lifecycle configuration
- Session management configuration

**Common Patterns Across Products**:
```go
// Config loading pattern
config, err := LoadConfig(configFile)
if err != nil {
    return err
}

// Validation pattern
if err := config.Validate(); err != nil {
    return fmt.Errorf("invalid config: %w", err)
}

// Merge pattern (file + CLI flags)
config = MergeConfig(fileConfig, cliConfig)
```

### Problems with Current Approach

1. **Code Duplication**: Each product reimplements config loading, validation, merging
2. **Inconsistent Validation**: Different validation rules across products
3. **Secret Handling**: Mixed approaches (files, env vars, hardcoded paths)
4. **No Hot Reload**: Configuration changes require service restart
5. **Limited Observability**: No metrics on config changes or validation failures

---

## Target Architecture

### Package Structure

```
internal/infra/configuration/
├── config.go              # Core configuration types and interfaces
├── loader.go              # Config loading from multiple sources
├── validator.go           # Schema validation and type checking
├── merger.go              # Configuration merging logic
├── secrets.go             # Secrets management integration
├── watcher.go             # File watching for hot reload
├── provider.go            # Configuration provider interface
└── providers/             # Provider implementations
    ├── file.go            # YAML/JSON file provider
    ├── env.go             # Environment variable provider
    ├── secret.go          # Docker/Kubernetes secrets provider
    └── consul.go          # Consul KV provider (future)
```

### Core Interfaces

```go
// ConfigProvider loads configuration from a specific source
type ConfigProvider interface {
    // Load loads configuration into the provided struct
    Load(ctx context.Context, target interface{}) error

    // Watch watches for configuration changes
    Watch(ctx context.Context) (<-chan ConfigChange, error)

    // Name returns the provider name
    Name() string
}

// ConfigValidator validates configuration structs
type ConfigValidator interface {
    // Validate validates the configuration
    Validate(ctx context.Context, config interface{}) error

    // ValidateField validates a specific field
    ValidateField(ctx context.Context, fieldName string, value interface{}) error
}

// ConfigManager manages configuration lifecycle
type ConfigManager interface {
    // Load loads configuration from all providers
    Load(ctx context.Context, target interface{}) error

    // Reload reloads configuration from all providers
    Reload(ctx context.Context) error

    // Watch watches for configuration changes
    Watch(ctx context.Context) (<-chan ConfigChange, error)

    // Get gets a configuration value by path
    Get(path string) (interface{}, error)

    // Set sets a configuration value by path
    Set(path string, value interface{}) error
}
```

### Configuration Flow

```
1. Application Start
   ├── ConfigManager.Load(ctx, &config)
   │   ├── FileProvider.Load()        # Load from YAML/JSON files
   │   ├── EnvProvider.Load()         # Load from environment variables
   │   ├── SecretProvider.Load()      # Load from Docker/K8s secrets
   │   └── Merge all sources (priority: secrets > env > file > defaults)
   ├── ConfigValidator.Validate(ctx, config)
   │   ├── Check required fields
   │   ├── Type checking
   │   ├── Range validation
   │   └── Cross-field validation
   └── Return validated config

2. Runtime Updates
   ├── ConfigManager.Watch(ctx)
   │   ├── FileProvider.Watch()       # File system watching
   │   ├── SecretProvider.Watch()     # Secret updates
   │   └── Emit ConfigChange events
   ├── Application receives ConfigChange
   ├── ConfigManager.Reload(ctx)
   └── Application applies new config
```

---

## Migration Plan

### Phase 1: Extract Common Configuration (Week 1)

**Goal**: Create base configuration infrastructure without breaking existing code

**Tasks**:
1. Create `internal/infra/configuration/` package
2. Implement `ConfigProvider` interface and basic providers (file, env, secret)
3. Implement `ConfigValidator` with struct tag-based validation
4. Implement `ConfigManager` for multi-source loading and merging
5. Add comprehensive tests (≥95% coverage)

**Success Criteria**:
- All tests passing
- No changes to existing product code yet
- Documentation complete with usage examples

### Phase 2: Migrate KMS Server (Week 2)

**Goal**: Replace KMS server config code with infra configuration

**Tasks**:
1. Update `internal/server/config/config.go` to use `ConfigManager`
2. Add struct tags for validation rules
3. Configure providers (file: `configs/kms/config.yml`, secrets: `/run/secrets/*`)
4. Remove custom config loading/validation code
5. Run full KMS test suite to verify no regressions

**Success Criteria**:
- All KMS tests passing
- Config loading behavior unchanged
- Code reduction in `internal/server/config/`

### Phase 3: Migrate Identity Server (Week 3)

**Goal**: Replace Identity server config code with infra configuration

**Tasks**:
1. Update `internal/identity/config/config.go` to use `ConfigManager`
2. Add struct tags for validation rules
3. Configure providers (file: `configs/identity/config.yml`, secrets: Docker secrets)
4. Remove custom config loading/validation code
5. Run full Identity test suite to verify no regressions

**Success Criteria**:
- All Identity tests passing
- Config loading behavior unchanged
- Code reduction in `internal/identity/config/`

### Phase 4: Add Hot Reload (Week 4)

**Goal**: Enable runtime configuration updates for non-critical settings

**Tasks**:
1. Implement file watching in `watcher.go`
2. Add `ConfigChange` event handling in `ConfigManager`
3. Document which settings support hot reload vs require restart
4. Add integration tests for hot reload scenarios
5. Update product configs to specify reload behavior

**Success Criteria**:
- Hot reload working for logging levels, timeouts, feature flags
- Sensitive settings (TLS, database) still require restart
- Clear documentation on reload behavior

---

## Configuration Examples

### File Provider (YAML)

```yaml
# configs/kms/config.yml
server:
  bind_address: "127.0.0.1"
  public_port: 8080
  admin_port: 9090
  tls:
    cert_file: "/app/tls_cert.pem"
    key_file: "/app/tls_key.pem"

database:
  type: "postgres"
  dsn: "file:///run/secrets/database_url"  # Load from secret
  max_open_conns: 10
  max_idle_conns: 5

logging:
  level: "info"  # Hot reload supported
  format: "json"

security:
  allowed_ips:
    - "127.0.0.1"
    - "::1"
  allowed_cidrs:
    - "10.0.0.0/8"
  cors_origins:
    - "https://localhost:8080"
```

### Struct Tags for Validation

```go
type ServerConfig struct {
    BindAddress string   `config:"bind_address" validate:"required,ip"`
    PublicPort  int      `config:"public_port" validate:"required,min=1,max=65535"`
    AdminPort   int      `config:"admin_port" validate:"required,min=1,max=65535"`
    TLS         TLSConfig `config:"tls" validate:"required"`
}

type TLSConfig struct {
    CertFile string `config:"cert_file" validate:"required,file"`
    KeyFile  string `config:"key_file" validate:"required,file"`
}

type DatabaseConfig struct {
    Type         string `config:"type" validate:"required,oneof=sqlite postgres"`
    DSN          string `config:"dsn" validate:"required"`
    MaxOpenConns int    `config:"max_open_conns" validate:"min=1,max=100"`
    MaxIdleConns int    `config:"max_idle_conns" validate:"min=1,max=100"`
}

type LoggingConfig struct {
    Level  string `config:"level" validate:"required,oneof=trace debug info warn error" hot-reload:"true"`
    Format string `config:"format" validate:"required,oneof=json text" hot-reload:"true"`
}
```

### Usage in Products

```go
// Load configuration
cfg := &ServerConfig{}
manager := configuration.NewManager(
    configuration.WithFileProvider("configs/kms/config.yml"),
    configuration.WithEnvProvider("CRYPTOUTIL"),
    configuration.WithSecretProvider("/run/secrets"),
    configuration.WithValidator(),
)

if err := manager.Load(ctx, cfg); err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Watch for changes (optional)
changes, err := manager.Watch(ctx)
if err != nil {
    return fmt.Errorf("failed to watch config: %w", err)
}

go func() {
    for change := range changes {
        if change.Field.HotReload {
            // Apply hot reload
            logger.Info("config updated", "field", change.Field.Name, "value", change.NewValue)
        } else {
            // Requires restart
            logger.Warn("config change requires restart", "field", change.Field.Name)
        }
    }
}()
```

---

## Testing Strategy

### Unit Tests

- **Provider Tests**: Test each provider independently (file, env, secret)
- **Validator Tests**: Test validation rules (required, type, range, custom)
- **Merger Tests**: Test configuration merging logic and priority
- **Watcher Tests**: Test file watching and change detection

### Integration Tests

- **Multi-Source Loading**: Test loading from file + env + secrets simultaneously
- **Hot Reload**: Test runtime configuration updates
- **Error Handling**: Test invalid configs, missing files, permission errors
- **Product Integration**: Test with actual KMS and Identity configs

### Performance Tests

- **Load Time**: Measure config loading time for various sizes
- **Reload Time**: Measure hot reload latency
- **Memory Usage**: Monitor memory overhead of configuration management
- **Watch Overhead**: Measure CPU/memory cost of file watching

---

## Dependencies

### Infrastructure Components

- **I5. Telemetry**: Logging, metrics for config operations
- **I16. Security**: Secret encryption, access control

### External Libraries

- **gopkg.in/yaml.v3**: YAML parsing
- **github.com/go-playground/validator/v10**: Struct validation
- **github.com/fsnotify/fsnotify**: File system watching

---

## Success Metrics

- **Code Reduction**: 50% reduction in product-specific config code
- **Test Coverage**: ≥95% for configuration infrastructure
- **Performance**: Config loading <100ms for typical configs
- **Hot Reload Latency**: <1s from file change to application update
- **Zero Regressions**: All existing product tests still passing

---

## Future Enhancements

### Distributed Configuration (I10. Messaging)

- Consul/etcd integration for distributed config management
- Configuration versioning and rollback
- Configuration diff and audit logging

### Configuration UI (I12. Documentation)

- Web UI for configuration management
- Visual configuration editor
- Configuration history and change tracking

### Advanced Validation

- Cross-product configuration validation
- Dependency checking (product A requires setting X in product B)
- Performance impact analysis for configuration changes

---

**Status**: PLANNING
**Next Steps**: Implement Phase 1 (extract common configuration infrastructure)
**Owner**: Infrastructure Team
