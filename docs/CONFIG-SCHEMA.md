# Config File Schema Reference

**Source of Truth**: [validate_schema.go](/internal/cmd/cicd/lint_deployments/validate_schema.go)

This document describes the flat kebab-case YAML config file schema used by `config-*.yml` files in `configs/`. The schema is hardcoded in Go and validated by `cicd lint-deployments validate-schema`.

## Key Format

Config keys use **flat kebab-case** (e.g., `bind-public-address`), NOT nested YAML (e.g., `server.bind_address`).

## Schema Fields

### Public Server Configuration

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `bind-public-protocol` | string | Yes | Public server protocol (MUST be `https`) |
| `bind-public-address` | string | Yes | Public server bind address (`0.0.0.0` for containers, `127.0.0.1` for local) |
| `bind-public-port` | int | Yes | Public server bind port (1-65535) |

### Admin Server Configuration

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `bind-private-protocol` | string | Yes | Admin server protocol (MUST be `https`) |
| `bind-private-address` | string | Yes | Admin server bind address (MUST be `127.0.0.1`) |
| `bind-private-port` | int | Yes | Admin server bind port (typically `9090`) |

### TLS Configuration

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `tls-public-mode` | string | Yes | TLS certificate mode for public endpoint (`auto-generate` or `manual`) |
| `tls-private-mode` | string | Yes | TLS certificate mode for admin endpoint (`auto-generate` or `manual`) |

### OTLP Telemetry

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `otlp` | bool | Yes | Enable OTLP telemetry export |
| `otlp-service` | string | No | OTLP service name (required when `otlp: true`) |
| `otlp-environment` | string | No | OTLP environment label (`development`, `production`, `ci`) |
| `otlp-endpoint` | string | No | OTLP collector endpoint (required when `otlp: true`) |

### CORS Configuration

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `cors-max-age` | int | No | CORS preflight cache duration in seconds |
| `cors-allowed-origins` | string[] | No | Allowed CORS origins |

### Session Configuration

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `browser-session-algorithm` | string | No | Browser session token format (`JWS`, `JWE`, `Opaque`) |
| `browser-session-jws-algorithm` | string | No | Browser JWS signing algorithm (`HS256`, `HS384`, `HS512`) |
| `browser-session-jwe-algorithm` | string | No | Browser JWE encryption algorithm |
| `service-session-algorithm` | string | No | Service session token format (`JWS`, `JWE`, `Opaque`) |
| `service-session-jws-algorithm` | string | No | Service JWS signing algorithm (`HS256`, `HS384`, `HS512`) |
| `service-session-jwe-algorithm` | string | No | Service JWE encryption algorithm |

### Database Configuration

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `database-url` | string | No | Database connection string (prefer `file:///run/secrets/` reference or `sqlite://` for dev) |

## Validation Rules

1. **YAML Syntax**: File must parse as valid YAML.
2. **Bind Address Format**: Must be valid IPv4 (via `net.ParseIP`).
3. **Port Range**: 1-65535 inclusive.
4. **Protocol**: Must be `https` (TLS required).
5. **Admin Bind Policy**: `bind-private-address` MUST be `127.0.0.1`.
6. **Secret References**: `database-url` must use `file:///run/secrets/` or `sqlite://` (never inline `postgres://`).
7. **OTLP Consistency**: When `otlp: true`, `otlp-service` and `otlp-endpoint` are required.

## Cross-References

- [ARCHITECTURE.md Section 12.4.8](/docs/ARCHITECTURE.md#1248-config-file-content-validation) - Config validation
- [ARCHITECTURE.md Section 12.5](/docs/ARCHITECTURE.md#125-config-file-architecture) - Config file architecture
