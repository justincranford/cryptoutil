# CLI Strategy Framework

## Overview

This document defines the CLI patterns for service APIs and administration across the cryptoutil multi-service repository. It establishes consistent command structure, flag conventions, output expectations, and shared helper packages.

**Cross-references:**

- [Service Groups Taxonomy](./service-groups.md) - Defines 43 service groups
- [Group Directory Blueprint](./blueprint.md) - Defines target directory structure
- [Import Alias Policy](./import-aliases.md) - Import alias conventions

---

## Command Structure

### Top-Level Command Pattern

**Pattern:** `cryptoutil <service-group> <subcommand> [flags]`

**Examples:**

```bash
# KMS operations
cryptoutil kms server start --config configs/kms/production.yml
cryptoutil kms key generate --type rsa --size 2048
cryptoutil kms barrier unseal --secret-file /run/secrets/unseal_1of5.secret

# Identity operations
cryptoutil identity authz server start --config configs/identity/authz.yml
cryptoutil identity idp server start --config configs/identity/idp.yml
cryptoutil identity rs server start --config configs/identity/rs.yml

# CA operations (future)
cryptoutil ca server start --config configs/ca/production.yml
cryptoutil ca cert issue --profile webserver --cn example.com
cryptoutil ca cert revoke --serial 0x1234567890abcdef --reason keyCompromise
```

### Backward Compatibility Aliases

**Legacy command:** `cryptoutil server` → **New command:** `cryptoutil kms server`

```bash
# Both commands work during transition period
cryptoutil server start --config configs/kms/production.yml  # Legacy (deprecated)
cryptoutil kms server start --config configs/kms/production.yml  # New (canonical)
```

**Deprecation strategy:**

- Warn users when legacy command used: "DEPRECATED: Use 'cryptoutil kms server' instead"
- Support legacy commands for 12 months after refactor completion
- Remove legacy aliases in major version release (v2.0.0)

---

## Flag Conventions

### Global Flags (All Commands)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | service-specific | Path to YAML configuration file |
| `--verbose` | bool | false | Enable verbose logging (DEBUG level) |
| `--quiet` | bool | false | Suppress non-error output |
| `--output-format` | string | `text` | Output format: `text`, `json`, `yaml` |
| `--log-level` | string | `INFO` | Log level: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `--help` | bool | false | Show help message |

**Examples:**

```bash
# Enable verbose output
cryptoutil kms key generate --type rsa --verbose

# JSON output for automation
cryptoutil kms key list --output-format json

# Custom log level
cryptoutil kms server start --log-level DEBUG
```

### Server Subcommand Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--bind-address` | string | `127.0.0.1` | Server bind address |
| `--port` | int | service-specific | Server port (8080 for KMS, 9090 for AuthZ, etc.) |
| `--tls-cert` | string | required | Path to TLS certificate file |
| `--tls-key` | string | required | Path to TLS private key file |
| `--dev` | bool | false | Development mode (SQLite in-memory, relaxed security) |
| `--config` | string | required | Path to server configuration YAML file |

**Examples:**

```bash
# Production server with TLS
cryptoutil kms server start \
  --config configs/kms/production.yml \
  --bind-address 0.0.0.0 \
  --port 8080 \
  --tls-cert /etc/certs/server.crt \
  --tls-key /etc/certs/server.key

# Development server (SQLite in-memory)
cryptoutil kms server start --dev
```

### Key Management Flags (KMS)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | required | Key type: `rsa`, `ecdsa`, `ecdh`, `ed25519`, `ed448`, `aes`, `hmac` |
| `--size` | int | type-specific | Key size (RSA: 2048/3072/4096, AES: 128/192/256) |
| `--curve` | string | type-specific | EC curve: `p256`, `p384`, `p521` (for ECDSA/ECDH) |
| `--name` | string | generated | Human-readable key name |
| `--export` | bool | false | Export key material (requires unseal) |
| `--export-format` | string | `pem` | Export format: `pem`, `der`, `jwk` |

**Examples:**

```bash
# Generate RSA key
cryptoutil kms key generate --type rsa --size 4096 --name production-signing-key

# Generate ECDSA key
cryptoutil kms key generate --type ecdsa --curve p384 --name ecdsa-p384-key

# Export key to JWK format
cryptoutil kms key export --name production-signing-key --format jwk
```

### Certificate Operations Flags (CA, future)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | required | Certificate profile: `webserver`, `codesigning`, `email`, `clientauth` |
| `--cn` | string | required | Common Name (CN) for certificate subject |
| `--san` | []string | empty | Subject Alternative Names (DNS, IP, email) |
| `--validity` | duration | `365d` | Certificate validity period |
| `--serial` | string | required | Certificate serial number (for revocation) |
| `--reason` | string | required | Revocation reason (for revoke command) |

**Examples:**

```bash
# Issue webserver certificate with SANs
cryptoutil ca cert issue \
  --profile webserver \
  --cn example.com \
  --san DNS:www.example.com,DNS:api.example.com,IP:192.168.1.100 \
  --validity 730d

# Revoke certificate
cryptoutil ca cert revoke \
  --serial 0x1234567890abcdef \
  --reason keyCompromise
```

---

## Output Formats

### Text Output (Default)

**Human-readable output for interactive use:**

```
$ cryptoutil kms key generate --type rsa --size 2048
✓ RSA key generated successfully
  Name:       rsa-2048-20250121-abc123
  Type:       RSA
  Size:       2048 bits
  Created:    2025-01-21T10:30:45Z
  Fingerprint: SHA256:1234567890abcdef...
```

**Characteristics:**

- Color-coded status (✓ green, ✗ red, ⚠ yellow)
- Emojis for visual clarity (✓, ✗, ⚠, 🔑, 📜, 🔒)
- Indented key-value pairs
- ISO 8601 timestamps
- Human-readable sizes (KB, MB, GB)

### JSON Output

**Machine-readable output for automation:**

```bash
cryptoutil kms key generate --type rsa --size 2048 --output-format json
```

```json
{
  "status": "success",
  "data": {
    "name": "rsa-2048-20250121-abc123",
    "type": "RSA",
    "size": 2048,
    "created_at": "2025-01-21T10:30:45Z",
    "fingerprint": "SHA256:1234567890abcdef..."
  },
  "metadata": {
    "command": "kms key generate",
    "duration_ms": 125,
    "version": "1.0.0"
  }
}
```

**Characteristics:**

- Snake_case field names
- ISO 8601 timestamps
- Consistent structure: `status`, `data`, `metadata`
- Error details in `error` field (when status != success)

### YAML Output

**Configuration-friendly output:**

```bash
cryptoutil kms key list --output-format yaml
```

```yaml
status: success
data:
  - name: rsa-2048-20250121-abc123
    type: RSA
    size: 2048
    created_at: "2025-01-21T10:30:45Z"
    fingerprint: "SHA256:1234567890abcdef..."
  - name: ecdsa-p384-20250121-def456
    type: ECDSA
    curve: P-384
    created_at: "2025-01-21T11:15:20Z"
    fingerprint: "SHA256:fedcba0987654321..."
metadata:
  command: kms key list
  duration_ms: 45
  version: "1.0.0"
```

**Characteristics:**

- Snake_case field names
- Quoted timestamps
- Easy copy/paste into config files

---

## Error Handling

### Error Output Format

**Text format (stderr):**

```
✗ Error: Failed to generate RSA key
  Cause: insufficient entropy
  Solution: Ensure /dev/urandom is accessible
  Exit code: 1
```

**JSON format (stdout):**

```json
{
  "status": "error",
  "error": {
    "message": "Failed to generate RSA key",
    "cause": "insufficient entropy",
    "solution": "Ensure /dev/urandom is accessible",
    "code": "ERR_INSUFFICIENT_ENTROPY"
  },
  "metadata": {
    "command": "kms key generate",
    "duration_ms": 12,
    "version": "1.0.0"
  }
}
```

### Exit Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Misuse of command (invalid flags, missing arguments) |
| 3 | Configuration error (invalid config file, missing required settings) |
| 4 | Authentication/authorization error |
| 5 | Resource not found (key, certificate, etc.) |
| 6 | Network error (connection refused, timeout) |
| 7 | Cryptographic operation failed |
| 8 | Database error |

**Examples:**

```bash
# Success
$ cryptoutil kms key generate --type rsa
echo $?
0

# Misuse (missing required flag)
$ cryptoutil kms key generate
echo $?
2

# Resource not found
$ cryptoutil kms key export --name nonexistent-key
echo $?
5
```

---

## Shared CLI Helper Package

### Package Structure

```
internal/common/cli/
├── flags.go           # Global flag definitions
├── output.go          # Output formatting (text, JSON, YAML)
├── errors.go          # Error handling and exit codes
├── config.go          # Configuration file loading
├── context.go         # Context management (timeout, cancellation)
└── validation.go      # Input validation helpers
```

### Helper Functions

#### Flag Management

```go
// internal/common/cli/flags.go

// RegisterGlobalFlags registers global flags (--verbose, --output-format, etc.)
func RegisterGlobalFlags(flagSet *flag.FlagSet, cfg *Config)

// RegisterServerFlags registers server-specific flags (--bind-address, --port, etc.)
func RegisterServerFlags(flagSet *flag.FlagSet, cfg *ServerConfig)

// RegisterKeyManagementFlags registers key operation flags (--type, --size, etc.)
func RegisterKeyManagementFlags(flagSet *flag.FlagSet, cfg *KeyConfig)
```

#### Output Formatting

```go
// internal/common/cli/output.go

// OutputFormatter interface for consistent output
type OutputFormatter interface {
    Success(data any) error
    Error(err error) error
    Info(message string) error
}

// NewFormatter creates formatter based on --output-format flag
func NewFormatter(format string, writer io.Writer) OutputFormatter

// TextFormatter for human-readable output
type TextFormatter struct{ /* ... */ }

// JSONFormatter for machine-readable JSON output
type JSONFormatter struct{ /* ... */ }

// YAMLFormatter for configuration-friendly YAML output
type YAMLFormatter struct{ /* ... */ }
```

#### Error Handling

```go
// internal/common/cli/errors.go

// CLIError represents a CLI-specific error with exit code
type CLIError struct {
    Message  string
    Cause    error
    Solution string
    Code     string
    ExitCode int
}

// ExitWithError formats error and exits with appropriate code
func ExitWithError(formatter OutputFormatter, err *CLIError)

// WrapError wraps standard error as CLIError
func WrapError(err error, exitCode int) *CLIError
```

#### Configuration Loading

```go
// internal/common/cli/config.go

// LoadConfig loads YAML configuration file
func LoadConfig(path string, target any) error

// MergeFlags merges command-line flags into loaded config
func MergeFlags(config any, flags *flag.FlagSet) error

// ValidateConfig validates configuration after loading
func ValidateConfig(config any) error
```

---

## CLI Implementation Examples

### KMS Server Start Command

```go
// cmd/kms/main.go (future)

func main() {
    // Create flag set
    flags := flag.NewFlagSet("kms", flag.ExitOnError)

    // Register flags
    cfg := &cli.Config{}
    cli.RegisterGlobalFlags(flags, cfg)
    cli.RegisterServerFlags(flags, &cfg.Server)

    // Parse flags
    flags.Parse(os.Args[2:])

    // Load configuration file
    if err := cli.LoadConfig(cfg.ConfigFile, cfg); err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 3))
    }

    // Merge flags into config (flags override file values)
    if err := cli.MergeFlags(cfg, flags); err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 3))
    }

    // Validate config
    if err := cli.ValidateConfig(cfg); err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 3))
    }

    // Create output formatter
    formatter := cli.NewFormatter(cfg.OutputFormat, os.Stdout)

    // Start server
    server := kms.NewServer(cfg)
    if err := server.Start(context.Background()); err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 1))
    }
}
```

### Identity AuthZ Server Start Command

```go
// cmd/identity/authz/main.go

func main() {
    // Similar pattern as KMS server
    flags := flag.NewFlagSet("authz", flag.ExitOnError)

    cfg := &cli.Config{}
    cli.RegisterGlobalFlags(flags, cfg)
    cli.RegisterServerFlags(flags, &cfg.Server)

    flags.Parse(os.Args[2:])

    // Load, merge, validate config
    if err := cli.LoadConfig(cfg.ConfigFile, cfg); err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 3))
    }

    // Start AuthZ server
    server := authz.NewServer(cfg)
    if err := server.Start(context.Background()); err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 1))
    }
}
```

### CA Certificate Issuance Command

```go
// cmd/ca/main.go (future)

func main() {
    flags := flag.NewFlagSet("ca", flag.ExitOnError)

    cfg := &cli.Config{}
    cli.RegisterGlobalFlags(flags, cfg)
    cli.RegisterCertificateFlags(flags, &cfg.Certificate)

    flags.Parse(os.Args[2:])

    formatter := cli.NewFormatter(cfg.OutputFormat, os.Stdout)

    // Issue certificate
    cert, err := ca.IssueCertificate(cfg.Certificate)
    if err != nil {
        cli.ExitWithError(formatter, cli.WrapError(err, 7))
    }

    // Output certificate
    formatter.Success(cert)
}
```

---

## Configuration File Structure

### KMS Configuration

```yaml
# configs/kms/production.yml

server:
  name: kms-production
  bind_address: 0.0.0.0
  port: 8080
  tls_cert_file: /etc/certs/kms.crt
  tls_key_file: /etc/certs/kms.key

database:
  type: postgresql
  dsn: postgres://user:pass@localhost:5432/kms?sslmode=require

unseal:
  mode: 3-of-5
  secret_files:
    - /run/secrets/unseal_1of5.secret
    - /run/secrets/unseal_2of5.secret
    - /run/secrets/unseal_3of5.secret
    - /run/secrets/unseal_4of5.secret
    - /run/secrets/unseal_5of5.secret

logging:
  level: INFO
  format: json
  output: stdout

telemetry:
  enabled: true
  otlp_endpoint: http://otel-collector:4317
  service_name: kms-production
```

### Identity Configuration

```yaml
# configs/identity/authz.yml

server:
  name: authz-production
  bind_address: 0.0.0.0
  port: 9090
  tls_cert_file: /etc/certs/authz.crt
  tls_key_file: /etc/certs/authz.key

database:
  type: postgresql
  dsn: postgres://user:pass@localhost:5432/identity?sslmode=require

tokens:
  access_token_format: jws
  issuer: https://authz.example.com
  audience: https://api.example.com

logging:
  level: INFO
  format: json
  output: stdout
```

---

## Testing Strategy

### CLI Integration Tests

```go
// internal/common/cli/cli_test.go

func TestKMSServerStart(t *testing.T) {
    t.Parallel()

    // Create test config
    cfg := &cli.Config{
        Server: cli.ServerConfig{
            BindAddress: "127.0.0.1",
            Port:        8080,
            TLSCertFile: "/tmp/test.crt",
            TLSKeyFile:  "/tmp/test.key",
        },
    }

    // Write config to temp file
    configPath := writeTestConfig(t, cfg)

    // Run CLI command
    cmd := exec.Command("cryptoutil", "kms", "server", "start", "--config", configPath)
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify server started
    assert.Contains(t, string(output), "Server started successfully")
}
```

### Output Format Tests

```go
// internal/common/cli/output_test.go

func TestJSONFormatter(t *testing.T) {
    t.Parallel()

    formatter := cli.NewFormatter("json", &bytes.Buffer{})

    data := map[string]any{
        "status": "success",
        "data":   map[string]string{"key": "value"},
    }

    err := formatter.Success(data)
    require.NoError(t, err)

    // Verify JSON output
    var result map[string]any
    json.Unmarshal(buffer.Bytes(), &result)
    assert.Equal(t, "success", result["status"])
}
```

---

## Validation Checklist

### CLI Framework Review

- [ ] Command structure follows `cryptoutil <service-group> <subcommand>` pattern
- [ ] Global flags available on all commands (--verbose, --output-format, --config)
- [ ] Server commands support common flags (--bind-address, --port, --tls-cert, --tls-key)
- [ ] Output formatters support text, JSON, and YAML formats
- [ ] Error handling includes actionable error messages with solutions
- [ ] Exit codes follow standard conventions (0=success, 1-8=specific errors)
- [ ] Configuration files use YAML format with snake_case fields
- [ ] Shared CLI helper package provides reusable components

### Prototype Skeleton Builds

- [ ] `cmd/kms/main.go` skeleton compiles and runs `--help`
- [ ] `cmd/identity/authz/main.go` skeleton compiles and runs `--help`
- [ ] `cmd/ca/main.go` skeleton compiles and runs `--help` (future)
- [ ] `internal/common/cli/` package compiles with no errors
- [ ] CLI integration tests pass

---

## Cross-References

- **Service Groups Taxonomy:** [docs/01-refactor/service-groups.md](./service-groups.md)
- **Group Directory Blueprint:** [docs/01-refactor/blueprint.md](./blueprint.md)
- **Import Alias Policy:** [docs/01-refactor/import-aliases.md](./import-aliases.md)
- **Current CLI Implementation:** `cmd/cryptoutil/main.go`, `internal/cmd/cryptoutil/cryptoutil.go`
- **Identity CLI Examples:** `cmd/identity/authz/main.go`, `cmd/identity/idp/main.go`

---

## Notes

- **Consistency:** All service groups follow same CLI patterns (flags, output, errors)
- **Backward compatibility:** Legacy `cryptoutil server` command supported during transition
- **Configuration first:** Prefer YAML config files over environment variables
- **Output flexibility:** Support text (human), JSON (automation), YAML (config) formats
- **Error clarity:** Provide actionable error messages with solutions
- **Testing:** CLI integration tests validate command execution and output formats
