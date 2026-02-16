# Config File Schema

## Overview

Config files use flat kebab-case YAML keys that map to viper/pflag definitions. They are NOT nested YAML.

## File Naming Convention

| Pattern | Description | Example |
|---------|-------------|---------|
| `config.yml` | Default/common config | `configs/cipher/im/config.yml` |
| `config-pg-N.yml` | PostgreSQL instance N | `configs/cipher/im/config-pg-1.yml` |
| `config-sqlite.yml` | SQLite instance | `configs/cipher/im/config-sqlite.yml` |
| `PRODUCT-SERVICE-server.yml` | Service-specific | `configs/ca/ca-server.yml` |

## Server Settings

### Public Server

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `bind-public-protocol` | string | `https` | Yes |
| `bind-public-address` | string | IPv4 address (`0.0.0.0`, `127.0.0.1`) | Yes |
| `bind-public-port` | integer | 1-65535 | Yes |

### Admin Server

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `bind-private-protocol` | string | `https` | Yes |
| `bind-private-address` | string | MUST be `127.0.0.1` | Yes |
| `bind-private-port` | integer | 1-65535 (typically `9090`) | Yes |

**Policy**: `bind-private-address` MUST always be `127.0.0.1` (admin never exposed outside container).

## TLS Configuration

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `tls-public-mode` | string | `auto`, `manual` | Yes |
| `tls-private-mode` | string | `auto`, `manual` | Yes |

## Telemetry (OTLP)

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `otlp` | boolean | `true`, `false` | Yes |
| `otlp-service` | string | Service name (e.g., `cipher-im-pg-1`) | When `otlp: true` |
| `otlp-environment` | string | `development`, `production`, `ci` | When `otlp: true` |
| `otlp-endpoint` | string | HTTP URL (e.g., `http://host:4317`) | When `otlp: true` |

## CORS Configuration

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `cors-max-age` | integer | Seconds (e.g., `3600`) | No |
| `cors-allowed-origins` | string[] | HTTP/HTTPS URLs | No |

## Session Configuration

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `browser-session-algorithm` | string | `JWS`, `JWE`, `Opaque` | No |
| `browser-session-jws-algorithm` | string | `HS256`, `HS384`, `HS512` | When JWS |
| `browser-session-jwe-algorithm` | string | `dir+A256GCM` | When JWE |
| `service-session-algorithm` | string | `JWS`, `JWE`, `Opaque` | No |
| `service-session-jws-algorithm` | string | `HS256`, `HS384`, `HS512` | When JWS |
| `service-session-jwe-algorithm` | string | `dir+A256GCM` | When JWE |

## Database Configuration

Database URLs are passed via Docker secrets, NOT in config files. Config files reference:

| Key | Type | Valid Values | Required |
|-----|------|-------------|----------|
| `database-url` | string | `file:///run/secrets/postgres_url.secret` OR `sqlite://...` | Yes |

## Validation Rules

1. **YAML Syntax**: File must parse as valid YAML
2. **Bind Address Format**: Must be valid IPv4 (`0.0.0.0`, `127.0.0.1`, etc.)
3. **Port Range**: 1-65535 inclusive
4. **Admin Bind Policy**: `bind-private-address` MUST be `127.0.0.1`
5. **Secret References**: Database URLs should use `file:///run/secrets/` pattern (not inline credentials)
6. **Protocol**: Must be `https` (TLS required)
