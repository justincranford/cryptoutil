# Configuration Reference

## Overview

cryptoutil uses a hierarchical configuration system that supports multiple input sources with proper validation and sensible defaults. Configuration can be provided through YAML files, command-line parameters, and environment-specific settings.

## Configuration Sources (Priority Order)

1. **Command-line parameters** (highest priority)
2. **YAML configuration files**
3. **Default values** (lowest priority)

## Configuration File Format

### Basic YAML Structure

```yaml
# config.yaml - Complete configuration example

# Server Binding Configuration
bind_public_protocol: "https"              # http | https
bind_public_address: "0.0.0.0"            # Listen address for public API
bind_public_port: 8080                     # Public API port

bind_private_protocol: "http"              # http | https 
bind_private_address: "127.0.0.1"         # Listen address for management API
bind_private_port: 9090                    # Management API port

# API Context Paths
browser_api_context_path: "/browser/api/v1"  # Browser client API context
service_api_context_path: "/service/api/v1"  # Service client API context

# TLS Configuration
tls_public_dns_names:                      # DNS names for public certificate
  - "localhost"
  - "cryptoutil.example.com"
tls_public_ip_addresses:                   # IP addresses for public certificate
  - "127.0.0.1"
  - "::1"
  - "192.168.1.100"
tls_private_dns_names:                     # DNS names for private certificate
  - "localhost"
tls_private_ip_addresses:                  # IP addresses for private certificate
  - "127.0.0.1"
  - "::1"

# Security Configuration
allowed_ips:                               # Individual allowed IP addresses
  - "127.0.0.1"
  - "::1"
  - "192.168.1.100"
allowed_cidrs:                             # Allowed CIDR blocks
  - "10.0.0.0/8"
  - "192.168.0.0/16"
  - "172.16.0.0/12"
ip_rate_limit: 100                         # Requests per second per IP

# CORS Configuration (Browser API only)
cors_allowed_origins: "http://localhost:3000,https://app.example.com"
cors_allowed_methods: "GET,POST,PUT,DELETE,OPTIONS"
cors_allowed_headers: "Content-Type,Authorization,X-CSRF-Token"
cors_max_age: 86400                        # CORS preflight cache time (seconds)

# CSRF Configuration (Browser API only)
csrf_token_name: "csrf_token"              # CSRF cookie name
csrf_token_same_site: "Strict"             # None | Lax | Strict
csrf_token_max_age: "1h"                   # Token validity duration
csrf_token_cookie_secure: true             # Require HTTPS for cookie
csrf_token_cookie_http_only: true          # HttpOnly cookie flag
csrf_token_cookie_session_only: false      # Session-only cookie
csrf_token_single_use_token: false         # Single-use tokens

# Database Configuration
database_container: "postgres"             # Container name for database
database_url: "postgres://user:pass@host:5432/cryptoutil?sslmode=require"
database_init_total_timeout: "60s"        # Total timeout for DB initialization
database_init_retry_wait: "5s"            # Wait between DB connection retries

# Observability Configuration
log_level: "INFO"                          # ALL | TRACE | DEBUG | CONFIG | INFO | NOTICE | WARN | ERROR | FATAL | OFF
verbose_mode: false                        # Enable verbose logging
dev_mode: false                            # Development mode flag
otlp: true                                 # Enable OpenTelemetry OTLP export
otlp_console: false                        # Enable console telemetry output
otlp_scope: "cryptoutil"                   # OpenTelemetry scope name

# Key Management Configuration
unseal_mode: "shamir"                      # simple | shamir | system
unseal_files:                              # Unseal key file paths
  - "/run/secrets/unseal_1of5"
  - "/run/secrets/unseal_2of5"
  - "/run/secrets/unseal_3of5"
```

## Configuration Parameters

### Server Binding

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `bind_public_protocol` | string | `"https"` | Protocol for public API (http/https) |
| `bind_public_address` | string | `"localhost"` | Listen address for public API |
| `bind_public_port` | uint16 | `8080` | Port for public API |
| `bind_private_protocol` | string | `"http"` | Protocol for management API (http/https) |
| `bind_private_address` | string | `"localhost"` | Listen address for management API |
| `bind_private_port` | uint16 | `9090` | Port for management API |

**Examples**:
```yaml
# Production binding
bind_public_address: "0.0.0.0"    # Listen on all interfaces
bind_private_address: "127.0.0.1" # Management API localhost only

# Development binding
bind_public_address: "localhost"   # Local development
bind_private_address: "localhost"  # Local management
```

### API Context Paths

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `browser_api_context_path` | string | `"/browser/api/v1"` | Base path for browser API endpoints |
| `service_api_context_path` | string | `"/service/api/v1"` | Base path for service API endpoints |

**Usage**:
- Browser API: Full security middleware (CORS, CSRF, CSP)
- Service API: Streamlined for machine-to-machine communication

### TLS Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `tls_public_dns_names` | []string | `["localhost"]` | DNS names for public TLS certificate |
| `tls_public_ip_addresses` | []string | `["127.0.0.1", "::1", "::ffff:127.0.0.1"]` | IP addresses for public certificate |
| `tls_private_dns_names` | []string | `["localhost"]` | DNS names for private TLS certificate |
| `tls_private_ip_addresses` | []string | `["127.0.0.1", "::1", "::ffff:127.0.0.1"]` | IP addresses for private certificate |

**Certificate Generation**:
- Automatic certificate generation for development
- Subject Alternative Names (SAN) include all DNS names and IPs
- RSA 2048-bit certificates with SHA-256 signatures

### Security Configuration

#### IP Access Control

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `allowed_ips` | []string | `[]` | Individual IP addresses allowed to access API |
| `allowed_cidrs` | []string | `[]` | CIDR blocks allowed to access API |
| `ip_rate_limit` | uint16 | `100` | Maximum requests per second per IP |

**Examples**:
```yaml
# Strict production access
allowed_ips:
  - "203.0.113.10"     # Specific public IP
  - "2001:db8::1"      # IPv6 address

allowed_cidrs:
  - "10.0.0.0/8"       # Private network
  - "192.168.1.0/24"   # Local subnet

# Development access
allowed_cidrs:
  - "0.0.0.0/0"        # Allow all (development only!)
```

#### CORS Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `cors_allowed_origins` | string | `"http://localhost:3000"` | Comma-separated allowed origins |
| `cors_allowed_methods` | string | `"GET,POST,PUT,DELETE,OPTIONS"` | Allowed HTTP methods |
| `cors_allowed_headers` | string | `"Content-Type,Authorization,X-CSRF-Token"` | Allowed request headers |
| `cors_max_age` | uint16 | `86400` | Preflight cache duration (seconds) |

**Origin Patterns**:
```yaml
# Development
cors_allowed_origins: "http://localhost:3000,http://localhost:3001"

# Production
cors_allowed_origins: "https://app.example.com,https://admin.example.com"

# Multiple environments
cors_allowed_origins: "https://*.example.com,https://app.dev.example.com"
```

#### CSRF Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `csrf_token_name` | string | `"csrf_token"` | Name of CSRF token cookie |
| `csrf_token_same_site` | string | `"Strict"` | SameSite cookie attribute |
| `csrf_token_max_age` | duration | `"1h"` | Token validity duration |
| `csrf_token_cookie_secure` | bool | `true` | Require HTTPS for cookie |
| `csrf_token_cookie_http_only` | bool | `true` | HttpOnly cookie flag |
| `csrf_token_cookie_session_only` | bool | `false` | Session-only cookie |
| `csrf_token_single_use_token` | bool | `false` | Single-use tokens |

**SameSite Options**:
- `"None"`: Cross-site requests allowed (requires Secure flag)
- `"Lax"`: Cross-site requests on navigation only
- `"Strict"`: No cross-site requests (recommended)

### Database Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `database_container` | string | `"postgres"` | Database container name for health checks |
| `database_url` | string | (required) | Database connection URL |
| `database_init_total_timeout` | duration | `"60s"` | Total timeout for database initialization |
| `database_init_retry_wait` | duration | `"5s"` | Wait between connection retries |

**Database URL Formats**:
```yaml
# PostgreSQL
database_url: "postgres://user:pass@host:5432/dbname?sslmode=require"

# PostgreSQL with connection pool
database_url: "postgres://user:pass@host:5432/dbname?pool_max_conns=25&pool_min_conns=5"

# SQLite (development)
database_url: "sqlite:./data/cryptoutil.db"

# In-memory SQLite (testing)
database_url: "sqlite::memory:"
```

### Observability Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `log_level` | string | `"INFO"` | Logging level |
| `verbose_mode` | bool | `false` | Enable verbose logging |
| `dev_mode` | bool | `false` | Development mode flag |
| `otlp` | bool | `true` | Enable OpenTelemetry OTLP export |
| `otlp_console` | bool | `false` | Enable console telemetry output |
| `otlp_scope` | string | `"cryptoutil"` | OpenTelemetry scope name |

**Log Levels**:
- `ALL`: All log messages
- `TRACE`: Very detailed debugging
- `DEBUG`: Debugging information
- `CONFIG`: Configuration information
- `INFO`: General information
- `NOTICE`: Important information
- `WARN`: Warning messages
- `ERROR`: Error messages
- `FATAL`: Critical errors
- `OFF`: No logging

### Key Management Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `unseal_mode` | string | `"simple"` | Unseal key mode |
| `unseal_files` | []string | (required) | Paths to unseal key files |

**Unseal Modes**:

1. **Simple Mode**: Single unseal key
   ```yaml
   unseal_mode: "simple"
   unseal_files: ["/path/to/unseal.key"]
   ```

2. **Shamir Secret Sharing**: M-of-N key sharing
   ```yaml
   unseal_mode: "shamir"
   unseal_files:
     - "/path/to/unseal_1of5.key"
     - "/path/to/unseal_2of5.key" 
     - "/path/to/unseal_3of5.key"
   ```

3. **System Fingerprinting**: Hardware-based unsealing
   ```yaml
   unseal_mode: "system"
   unseal_files: ["/path/to/system.fingerprint"]
   ```

## Command-Line Parameters

### Available Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--help` | `-h` | bool | Show help message |
| `--config` | `-y` | string | Path to configuration file |
| `--log-level` | `-l` | string | Log level (ALL, TRACE, DEBUG, etc.) |
| `--verbose` | `-v` | bool | Enable verbose logging |
| `--dev` | `-d` | bool | Enable development mode |
| `--bind-public-protocol` | `-t` | string | Public API protocol (http/https) |
| `--bind-public-address` | `-a` | string | Public API listen address |
| `--bind-public-port` | `-p` | uint16 | Public API port |
| `--bind-private-protocol` | `-T` | string | Private API protocol (http/https) |
| `--bind-private-address` | `-A` | string | Private API listen address |
| `--bind-private-port` | `-P` | uint16 | Private API port |

### Usage Examples

```bash
# Basic usage with config file
./cryptoutil --config=production.yaml

# Override specific settings
./cryptoutil --config=base.yaml --log-level=DEBUG --verbose

# Development mode with overrides
./cryptoutil --dev --bind-public-port=8080 --bind-private-port=9090

# Production with custom network binding
./cryptoutil --config=prod.yaml \
  --bind-public-address=0.0.0.0 \
  --bind-private-address=127.0.0.1
```

## Environment-Specific Configurations

### Development Configuration

```yaml
# configs/development.yaml
dev_mode: true
verbose_mode: true
log_level: "DEBUG"

bind_public_protocol: "http"
bind_private_protocol: "http"

# Relaxed security for development
allowed_cidrs: ["0.0.0.0/0"]
ip_rate_limit: 1000
csrf_token_cookie_secure: false

# In-memory database
database_url: "sqlite::memory:"

# Console telemetry
otlp_console: true
otlp: false
```

### Production Configuration

```yaml
# configs/production.yaml
dev_mode: false
verbose_mode: false
log_level: "INFO"

bind_public_protocol: "https"
bind_private_protocol: "https"
bind_public_address: "0.0.0.0"
bind_private_address: "127.0.0.1"

# Strict security
allowed_cidrs: ["10.0.0.0/8"]
ip_rate_limit: 10
csrf_token_cookie_secure: true
csrf_token_same_site: "Strict"

# Production database
database_url: "postgres://user:pass@db:5432/cryptoutil?sslmode=require"

# OpenTelemetry
otlp: true
otlp_console: false
```

### Docker Configuration

```yaml
# configs/docker.yaml
bind_public_address: "0.0.0.0"
bind_private_address: "0.0.0.0"

# Database from environment/secrets
database_url: "postgres://cryptoutil_user:$(cat /run/secrets/db_password)@postgres:5432/cryptoutil"

# Unseal keys from Docker secrets
unseal_mode: "shamir"
unseal_files:
  - "/run/secrets/unseal_1of5"
  - "/run/secrets/unseal_2of5"
  - "/run/secrets/unseal_3of5"
```

## Configuration Validation

### Startup Validation

The system validates configuration at startup:

```go
// Example validation errors
2025/09/12 10:30:00 FATAL Configuration validation failed:
  - bind_public_port: must be between 1 and 65535
  - database_url: invalid postgres connection string
  - unseal_files: file not found: /path/to/unseal.key
  - allowed_cidrs: invalid CIDR notation: 192.168.1/24
```

### Required vs Optional Parameters

**Required Parameters**:
- `database_url`
- `unseal_files` (at least one file)

**Optional Parameters** (have defaults):
- All binding and port configurations
- Security settings (have safe defaults)
- Logging and telemetry settings

### Configuration Security

#### Sensitive Parameter Handling

Sensitive parameters are redacted in logs:

```bash
# Visible in logs
2025/09/12 10:30:00 INFO Configuration loaded:
  bind_public_address: "0.0.0.0"
  bind_public_port: 8080
  log_level: "INFO"

# Redacted in logs
  database_url: "[REDACTED]"
  unseal_files: "[REDACTED]"
```

**Exception**: In development mode with verbose logging, sensitive values may be shown for debugging.

#### Secret File References

Support for reading secrets from files:

```yaml
# Direct value
database_url: "postgres://user:password@host:5432/db"

# File reference (Docker secrets pattern)
database_url_file: "/run/secrets/database_url"
```

## Configuration Best Practices

### Security Recommendations

1. **Principle of Least Privilege**:
   ```yaml
   # Restrictive IP allowlisting
   allowed_cidrs: ["10.0.0.0/8"]  # Internal network only
   ip_rate_limit: 10              # Conservative rate limiting
   ```

2. **TLS in Production**:
   ```yaml
   bind_public_protocol: "https"   # Always HTTPS in production
   csrf_token_cookie_secure: true  # Secure cookies only
   ```

3. **Database Security**:
   ```yaml
   database_url: "postgres://user:pass@host:5432/db?sslmode=require"
   ```

### Performance Recommendations

1. **Rate Limiting**:
   ```yaml
   # Balanced rate limiting
   ip_rate_limit: 100  # Allow reasonable burst traffic
   ```

2. **Database Connections**:
   ```yaml
   database_url: "postgres://user:pass@host:5432/db?pool_max_conns=25&pool_min_conns=5"
   ```

3. **Logging**:
   ```yaml
   log_level: "INFO"     # Avoid DEBUG in production
   verbose_mode: false   # Disable verbose mode in production
   ```

### Operational Recommendations

1. **Configuration Management**:
   - Version control configuration files
   - Use environment-specific configurations
   - Validate configurations in CI/CD

2. **Secret Management**:
   - Use Docker secrets or Kubernetes secrets
   - Rotate unseal keys regularly
   - Never commit secrets to version control

3. **Monitoring**:
   ```yaml
   otlp: true           # Enable telemetry export
   otlp_console: false  # Disable console output in production
   ```

This comprehensive configuration reference ensures proper setup and operation of cryptoutil across all deployment scenarios while maintaining security and performance best practices.
