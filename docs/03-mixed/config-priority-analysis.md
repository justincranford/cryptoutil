# Configuration Priority Order Analysis

## Current Implementation Review

### Viper Configuration Loading Sequence

Based on analysis of `internal/common/config/config.go` (lines 700-850), the current configuration precedence is:

1. **Command-line flags** (pflag) - Highest priority
2. **Configuration YAML files** (viper.ReadInConfig / viper.MergeInConfig)
3. **Configuration profiles** (dev, stg, prod, test) - Applied if flag not explicitly set
4. **Default values** (registered in code)

### Current Priority Order

```go
// 1. Register default values
pflag.StringP(setting.name, setting.shorthand, defaultValue, setting.usage)

// 2. Bind flags to Viper
viper.BindPFlags(pflag.CommandLine)

// 3. Load config files (if provided via --config flag)
viper.SetConfigFile(configFiles[0])
viper.ReadInConfig()   // First config file

// 4. Merge additional config files
viper.MergeInConfig()  // Subsequent config files

// 5. Apply profile settings (only if not already set)
if profileName != "" && profileConfig, exists := profiles[profileName]; exists {
    for key, value := range profileConfig {
        if !viper.IsSet(key) {  // ← Only sets if not already configured
            viper.Set(key, value)
        }
    }
}

// 6. Read final values
s.DatabaseURL = viper.GetString(databaseURL.name)  // etc.
```

## Analysis: Current vs Required Priority Order

### Required Priority Order (from Task INF5)

1. **Docker/Kubernetes secrets** (credentials and sensitive settings)
2. **Configuration YAML files** (non-sensitive settings)
3. **Command parameters** (first fallback to override 1 or 2)
4. **Environment variables** (last fallback to override 1, 2, or 3)

### Current Viper Priority Order

1. **Command parameters** (--flag) - HIGHEST
2. **Config YAML files** (--config)
3. **Profiles** (--profile dev/stg/prod/test)
4. **Defaults** (hardcoded in code) - LOWEST

### Issues Identified

#### ❌ ISSUE 1: Environment variables NOT supported

**Problem**: Viper is NOT configured to read from environment variables

**Current State**:
- No `viper.AutomaticEnv()` call
- No `viper.BindEnv()` calls
- Environment variables are NOT part of precedence chain

**Impact**: Cannot override configuration via environment variables (required for 12-factor apps)

**Required Fix**: Add environment variable support with proper precedence

#### ❌ ISSUE 2: Secrets handling via file:// URLs

**Current Implementation**:
```go
if strings.HasPrefix(s.DatabaseURL, "file://") {
    filePath := strings.TrimPrefix(s.DatabaseURL, "file://")
    content, err := os.ReadFile(filePath)
    s.DatabaseURL = strings.TrimSpace(string(content))
}
```

**Problem**: Only `DatabaseURL` supports `file://` prefix for Docker/K8s secrets

**Impact**: Other sensitive settings (TLS certs, keys, passwords) cannot use Docker/K8s secrets

**Required Fix**: Extend `file://` support to all sensitive settings

#### ✅ CORRECT: Command flags have highest priority

**Current**: Command flags override config files ✅

**Why correct**: Matches requirement "Command parameters as first fallback to override 1 or 2"

#### ⚠️ PARTIAL: Config file precedence

**Current**: Config files override profile defaults ✅

**Missing**: No distinction between:
- Sensitive data (should come from Docker/K8s secrets)
- Non-sensitive data (should come from YAML files)

## Recommended Configuration Priority Order

### Target Precedence (Highest to Lowest)

```
1. Command-line flags (--flag)             ← HIGHEST (current: ✅ correct)
2. Environment variables (ENV_VAR)         ← (current: ❌ not implemented)
3. Docker/K8s secrets (file:// URLs)       ← (current: ⚠️ partial - only database-url)
4. Config YAML files (--config)            ← (current: ✅ correct position)
5. Profile defaults (--profile dev/stg)    ← (current: ✅ correct position)
6. Hardcoded defaults                      ← LOWEST (current: ✅ correct)
```

### Rationale for Precedence

**Command flags > Env vars**:
- Allows operators to override any config temporarily
- Useful for debugging and emergency interventions

**Env vars > Secrets**:
- Env vars can override secrets for non-production environments
- Supports containerized development workflows

**Secrets > Config files**:
- Secrets contain sensitive data (passwords, keys)
- Config files contain non-sensitive settings (ports, timeouts)
- Secrets should override generic config file values

**Config files > Profiles**:
- Config files are deployment-specific
- Profiles are generic templates

## Implementation Plan

### Step 1: Add Environment Variable Support

```go
// Enable automatic environment variable reading
viper.SetEnvPrefix("CRYPTOUTIL")  // Prefix: CRYPTOUTIL_
viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))  // database-url → DATABASE_URL
viper.AutomaticEnv()

// Example: --database-url can be overridden by CRYPTOUTIL_DATABASE_URL
```

### Step 2: Extend file:// Support to All Sensitive Settings

**Sensitive settings requiring file:// support**:
- `database-url` (currently supported ✅)
- `tls-public-cert-file` (need to add)
- `tls-public-key-file` (need to add)
- `tls-private-cert-file` (need to add)
- `tls-private-key-file` (need to add)
- `unseal-files` (already supports file paths ✅)

**Implementation**:
```go
func resolveFileURL(value string) (string, error) {
    if strings.HasPrefix(value, "file://") {
        filePath := strings.TrimPrefix(value, "file://")
        content, err := os.ReadFile(filePath)
        if err != nil {
            return "", fmt.Errorf("failed to read from file %s: %w", filePath, err)
        }
        return strings.TrimSpace(string(content)), nil
    }
    return value, nil
}

// Apply to all sensitive settings
s.DatabaseURL, err = resolveFileURL(viper.GetString(databaseURL.name))
s.TLSPublicCertFile, err = resolveFileURL(viper.GetString(tlsPublicCertFile.name))
s.TLSPublicKeyFile, err = resolveFileURL(viper.GetString(tlsPublicKeyFile.name))
// ... etc
```

### Step 3: Updated Configuration Load Sequence

```go
// 1. Register defaults
pflag.StringP(...)

// 2. Bind flags
viper.BindPFlags(pflag.CommandLine)

// 3. Configure environment variables
viper.SetEnvPrefix("CRYPTOUTIL")
viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
viper.AutomaticEnv()

// 4. Load config files
viper.ReadInConfig()
viper.MergeInConfig()

// 5. Apply profile defaults (if not set)
if !viper.IsSet(key) {
    viper.Set(key, value)
}

// 6. Resolve file:// URLs for secrets
s.DatabaseURL, err = resolveFileURL(viper.GetString(databaseURL.name))
```

### Step 4: Testing Configuration Precedence

**Test scenarios**:
1. **Command flag overrides all**: `--database-url=value1` + `CRYPTOUTIL_DATABASE_URL=value2` + `config.yml` → value1
2. **Env var overrides config**: `CRYPTOUTIL_DATABASE_URL=value2` + `config.yml` → value2
3. **Secret file overrides config**: `database-url: file:///run/secrets/db_url` + `config.yml` → secret content
4. **Config file overrides profile**: `config.yml` + `--profile dev` → config.yml value
5. **Profile overrides defaults**: `--profile dev` → profile value

## Documentation Updates Required

### 1. README.md - Configuration Section

```markdown
## Configuration Priority Order

cryptoutil uses a layered configuration system with the following precedence (highest to lowest):

1. **Command-line flags** - Highest priority, overrides all other sources
   - Example: `--database-url=postgres://...`

2. **Environment variables** - Prefix: `CRYPTOUTIL_`
   - Example: `CRYPTOUTIL_DATABASE_URL=postgres://...`
   - Hyphenated flags become underscored: `database-url` → `DATABASE_URL`

3. **Docker/Kubernetes secrets** - Use `file://` URLs
   - Example: `database-url: file:///run/secrets/database_url`
   - Supports: database-url, tls-*-cert-file, tls-*-key-file

4. **Configuration YAML files** - Specified via `--config`
   - Example: `--config=configs/production/config.yml`

5. **Configuration profiles** - Pre-defined templates
   - Example: `--profile=dev` (dev, stg, prod, test)

6. **Hardcoded defaults** - Lowest priority

### Using Docker Secrets

```yaml
# docker-compose.yml
services:
  cryptoutil:
    command: ["server", "start", "--database-url=file:///run/secrets/database_url"]
    secrets:
      - database_url

secrets:
  database_url:
    file: ./secrets/database_url.secret
```

### Using Kubernetes Secrets

```yaml
# deployment.yaml
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: cryptoutil
      args: ["server", "start", "--database-url=file:///run/secrets/database-url"]
      volumeMounts:
        - name: secrets
          mountPath: /run/secrets
          readOnly: true
  volumes:
    - name: secrets
      secret:
        secretName: cryptoutil-secrets
```
```

### 2. Architecture Instructions Update

Add to `.github/instructions/01-03.golang.instructions.md`:

```markdown
## Configuration Management Best Practices

### Priority Order (Highest to Lowest)
1. Command-line flags (--flag)
2. Environment variables (CRYPTOUTIL_*)
3. Docker/Kubernetes secrets (file:// URLs)
4. Configuration YAML files (--config)
5. Profile defaults (--profile)
6. Hardcoded defaults

### Sensitive Data Handling
- **ALWAYS use Docker secrets** for production credentials
- **NEVER commit secrets** to config files or code
- **Use file:// URLs** for database-url, TLS certs, TLS keys

### Environment Variables
- Prefix: `CRYPTOUTIL_`
- Hyphenated flags → Underscored env vars
- Example: `--database-url` → `CRYPTOUTIL_DATABASE_URL`
```

## Verification Checklist

- [ ] Add `viper.AutomaticEnv()` support
- [ ] Add `viper.SetEnvPrefix("CRYPTOUTIL")`
- [ ] Add `viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))`
- [ ] Extend `file://` support to all sensitive settings
- [ ] Create `resolveFileURL()` helper function
- [ ] Update README.md with configuration priority section
- [ ] Update architecture instructions
- [ ] Add configuration precedence tests
- [ ] Test Docker Compose with secrets
- [ ] Test Kubernetes deployment with secrets
- [ ] Document environment variable naming conventions

## Priority

**High** - Affects production security and deployment flexibility

## Timeline

**Implementation**: 4 hours
**Testing**: 2 hours
**Documentation**: 2 hours
**Total**: 8 hours (1 day)
