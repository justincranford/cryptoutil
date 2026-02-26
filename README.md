# cryptoutil

[![CICD - Lint Deployments](https://github.com/justincranford/cryptoutil/actions/workflows/cicd-lint-deployments.yml/badge.svg)](https://github.com/justincranford/cryptoutil/actions/workflows/cicd-lint-deployments.yml)

## Introduction

**cryptoutil** is a production-ready suite of four cryptographic products, designed with enterprise-grade security, **FIPS 140-3** standards compliance, and Zero-Trust principles:

1. **Private Key Infrastructure (PKI)** - X.509 certificate management with EST, SCEP, OCSP, and CRL support
2. **JSON Object Signing and Encryption (JOSE)** - JWK/JWS/JWE/JWT cryptographic operations
3. **Secrets Manager (SM)** - Elastic key management service with hierarchical key barriers; includes Instant Messenger (IM) with encryption-at-rest
4. **Identity** - OAuth 2.1, OIDC 1.0, WebAuthn, and Passkeys authentication and authorization

### Project Background

The project began as a standalone **Key Management Service (KMS)**. As part of that design, I implemented **CA** and **JOSE** components, and the next logical step was to add an **Identity** component.

- **CA** => Issue TLS certificates to protect **data-in-transit**.
- **JOSE** => Issue JSON (JWEs) to protect **data-at-rest**.
- **Identity** => Provide multi-tenant **authentication** (AuthN) and flexible **authorization** (AuthZ).
**

### Evolution of the Architecture

Instead of keeping Identity internal, I chose to build it as an **independent external service**. It wanted to run on its own, not just as a dependency of the KMS.

I also decided to refactor the **CA** and **JOSE** components into external services. Both the KMS and Identity depend on them, but those components are widely useful on their own.

### Why This Exists

This project is **for fun**. I’ve worked on all four types of systems professionally, and I genuinely enjoy the challenge of building my own versions.

This opportunity is also a great learning experience, especially with respect to using **LLM agents** for Spec-Driven Development (SDD).

Finally, this project has given me a much greater appreciation for the breadth of work that goes into delivering modern, enterprise-ready security products.

## Key Features

### 🔐 Cryptographic Standards

- **FIPS 140-3 Compliance**: Only uses NIST-approved algorithms (RSA ≥2048, AES ≥128, NIST curves, EdDSA)
- **Key Generation**: RSA, ECDSA, ECDH, EdDSA, AES, HMAC, UUIDv7 with concurrent key pools
- **JWE/JWS Support**: Full JSON Web Encryption and Signature implementation
- **Hierarchical Key Management**: Multi-tier barrier system (unseal → root → intermediate → content keys)

### 🌐 API Architecture

- **Dual Context Design**:
  - **Browser API** (`/browser/api/v1/*`) - Full browser security (CORS, CSRF, CSP)
  - **Service API** (`/service/api/v1/*`) - Optimized for service-to-service communication
- **Management Interface** (`localhost:9090`) - Private health checks and graceful shutdown
- **OpenAPI-Driven**: Auto-generated handlers, models, and interactive Swagger UI

#### Context Paths Hierarchy

```
cryptoutil Server Applications
│
├── 🌐 Public Fiber App (Port 8080 - HTTPS)
│   │
│   ├── 📋 Swagger UI Routes
│   │   ├── GET /ui/swagger/doc.json              # OpenAPI spec JSON
│   │   └── GET /ui/swagger/*                     # Swagger UI interface
│   │
│   ├── 🔒 CSRF Token Route
│   │   └── GET /browser/api/v1/csrf-token        # Get CSRF token for browser clients
│   │
│   ├── 🌐 Browser API Context (/browser/api/v1)  # For browser clients with CORS/CSRF
│   │   ├── POST   /browser/api/v1/elastickey           # Create elastic key
│   │   ├── GET    /browser/api/v1/elastickey/{id}      # Get elastic key by ID
│   │   ├── GET    /browser/api/v1/elastickeys          # Find elastic keys (filtered)
│   │   ├── PUT    /browser/api/v1/elastickey/{id}      # Update elastic key
│   │   ├── DELETE /browser/api/v1/elastickey/{id}      # Delete elastic key
│   │   ├── POST   /browser/api/v1/materialkey          # Create material key
│   │   ├── GET    /browser/api/v1/materialkey/{id}     # Get material key by ID
│   │   ├── GET    /browser/api/v1/materialkeys         # Find material keys (filtered)
│   │   ├── PUT    /browser/api/v1/materialkey/{id}     # Update material key
│   │   ├── DELETE /browser/api/v1/materialkey/{id}     # Delete material key
│   │   ├── POST   /browser/api/v1/crypto/encrypt       # Encrypt operation
│   │   ├── POST   /browser/api/v1/crypto/decrypt       # Decrypt operation
│   │   ├── POST   /browser/api/v1/crypto/sign          # Sign operation
│   │   ├── POST   /browser/api/v1/crypto/verify        # Verify operation
│   │   └── POST   /browser/api/v1/crypto/generate      # Generate operation
│   │
│   └── 🔧 Service API Context (/service/api/v1)  # For service clients without browser middleware
│       ├── POST   /service/api/v1/elastickey           # Create elastic key
│       ├── GET    /service/api/v1/elastickey/{id}      # Get elastic key by ID
│       ├── GET    /service/api/v1/elastickeys          # Find elastic keys (filtered)
│       ├── PUT    /service/api/v1/elastickey/{id}      # Update elastic key
│       ├── DELETE /service/api/v1/elastickey/{id}      # Delete elastic key
│       ├── POST   /service/api/v1/materialkey          # Create material key
│       ├── GET    /service/api/v1/materialkey/{id}     # Get material key by ID
│       ├── GET    /service/api/v1/materialkeys         # Find material keys (filtered)
│       ├── PUT    /service/api/v1/materialkey/{id}     # Update material key
│       ├── DELETE /service/api/v1/materialkey/{id}     # Delete material key
│       ├── POST   /service/api/v1/crypto/encrypt       # Encrypt operation
│       ├── POST   /service/api/v1/crypto/decrypt       # Decrypt operation
│       ├── POST   /service/api/v1/crypto/sign          # Sign operation
│       ├── POST   /service/api/v1/crypto/verify        # Verify operation
│       └── POST   /service/api/v1/crypto/generate      # Generate operation
│
└── 🔐 Private Fiber App (Port 9090 - HTTP)
    ├── 🩺 Health Check Routes
    │   ├── GET  /livez                              # Liveness probe (Kubernetes)
    │   └── GET  /readyz                             # Readiness probe (Kubernetes)
    │
    └── 🛑 Management Routes
        └── POST /shutdown                           # Graceful shutdown endpoint
```

### 🛡️ Security Features

- **Multi-layered IP allowlisting** (individual IPs + CIDR blocks)
- **Per-IP rate limiting** with separate thresholds for browser vs service APIs (100 req/sec browser, 25 req/sec service)
- **CSRF protection** with secure token handling for browser clients
- **Content Security Policy (CSP)** for XSS prevention
- **Comprehensive security headers** (Helmet.js equivalent)
- **Encrypted key storage** with barrier system protection

### 📊 Observability & Monitoring

- **OpenTelemetry integration** (traces, metrics, logs via OTLP)
- **Structured logging** with slog
- **Kubernetes-ready health endpoints** (`/livez`, `/readyz`)
- **Performance metrics** for cryptographic operations
- **Grafana-OTEL-LGTM stack**: Integrated Grafana, Loki, Tempo, and Prometheus
- **Telemetry forwarding architecture**: cryptoutil services → OpenTelemetry Collector → Grafana-OTEL-LGTM

#### Telemetry Architecture

**Dual Telemetry Flows for Complete Observability:**

**Application Telemetry (Push-based):**

```
cryptoutil services (OTLP GRPC:4317 or HTTP:4318) → OpenTelemetry Collector Contrib → Grafana-OTEL-LGTM (OTLP HTTP:4318)
```

- **Purpose**: Business application traces, logs, and metrics
- **Protocol**: OTLP (OpenTelemetry Protocol) - push-based with automatic protocol detection
- **Data**: Crypto operations, API calls, business logic telemetry
- **Configuration**: Set `otlp-endpoint` with protocol prefix (`grpc://` or `http://`) for automatic exporter selection
- **Data**: Crypto operations, API calls, business logic telemetry

**Infrastructure Telemetry (Pull-based):**

```
Grafana-OTEL-LGTM (Prometheus) → OpenTelemetry Collector Contrib (HTTP:8888/metrics)
```

- **Purpose**: Monitor collector health and performance
- **Protocol**: Prometheus scraping - pull-based
- **Data**: Collector throughput, error rates, queue depths, resource usage

**Why Both Flows?** The collector both **receives application telemetry** (from cryptoutil) and **exposes its own metrics** (for monitoring). This provides complete observability of both your application and the telemetry pipeline itself.

#### Port Architecture

**OpenTelemetry Collector Ports:**

- **4317**: OTLP gRPC receiver (application telemetry ingress)
- **4318**: OTLP HTTP receiver (application telemetry ingress)
- **8888**: OpenTelemetry self-metrics (Prometheus, internal scraping)
- **8889**: OpenTelemetry received-metrics (Prometheus, for re-export)
- **13133**: Health check extension (container health monitoring)
- **1777**: pprof (performance profiling)
- **55679**: zPages (debugging UI)

**Application Ports:**

- **3000**: Grafana UI
- **5432**: PostgreSQL database
- **8080**: cryptoutil public API (HTTPS)
- **8081-8082**: Additional cryptoutil instances in Docker Compose (HTTPS)
- **8060**: JOSE Authority Server (HTTPS)
- **8050**: Certificate Authority Server (HTTPS)
- **9090**: cryptoutil private admin API (health checks, graceful shutdown) on all instances
- **14317**: Grafana OTLP gRPC receiver (telemetry ingress)
- **14318**: Grafana OTLP HTTP receiver (telemetry ingress)

**Services:**

- **Grafana**: <http://localhost:3000> (admin/admin)
- **Prometheus**: Integrated into Grafana-OTEL-LGTM stack
- **Loki**: Integrated log aggregation
- **Tempo**: Integrated trace storage
- **OpenTelemetry Collector**: Receives telemetry from cryptoutil services
- **JOSE Authority Server**: <https://localhost:8060> (cryptographic operations)
- **Certificate Authority**: <https://localhost:8050> (X.509 certificate management)

### 🏗️ Production Ready

- **Database support**: PostgreSQL (production), SQLite (development/testing)
- **Container deployment**: Docker Compose with secret management
- **Configuration management**: YAML files + CLI parameters
- **Graceful shutdown**: Signal handling and connection draining

## Quick Start

### Prerequisites

- Go 1.25.7+
- Docker Desktop (required for integration tests using testcontainers)
- Docker and Docker Compose (for PostgreSQL)

### Automation Tools

- **ci-identity-validation**: Automated PR validation (50% manual QA reduction)
- **markdownlint-cli2**: Automated markdown formatting

See [docs/TOOLS.md](docs/TOOLS.md) for detailed tool documentation.

### Identity System: Unified CLI (One-Liner Bootstrap)

The Identity system provides a unified command-line interface for managing OAuth 2.1 services:

```powershell
# Build unified CLI
go build -o bin/identity.exe ./cmd/identity-unified

# Build multi-product launcher (includes identity services)
go build -o bin/cryptoutil.exe ./cmd/cryptoutil

# Start all services with one command
./bin/identity start --profile demo

# Check service health
./bin/identity health

# Start individual identity services
./bin/cryptoutil identity authz --config /path/to/config.yml
./bin/cryptoutil identity idp --config /path/to/config.yml
./bin/cryptoutil identity rs --config /path/to/config.yml
````

**Available Profiles:**

- `demo`: All services (AuthZ, IdP, RS) with SQLite in-memory
- `authz-only`: Authorization Server only
- `authz-idp`: AuthZ + IdP without Resource Server
- `full-stack`: All services with PostgreSQL
- `ci`: Minimal config for CI/CD testing

For comprehensive usage, see [Unified CLI Guide](docs/02-identityV2/historical/unified-cli-guide.md).

### KMS Server: Running with Docker Compose

**Security Requirement**: All services use Docker secrets for credentials (NO inline environment variables). See [Docker Secrets Pattern Guide](docs/docker-secrets-pattern.md) for comprehensive documentation.

```sh
# Start full stack: PostgreSQL, cryptoutil, and observability
cd deployments/cryptoutil-suite
docker compose up -d

# View logs
docker compose logs -f cryptoutil-postgres-1

# Access services
# Grafana UI: http://localhost:3000 (admin/admin)
# cryptoutil API: https://localhost:8081 (PostgreSQL instance 1), https://localhost:8082 (PostgreSQL instance 2), or https://localhost:8080 (SQLite)
# Swagger UI: https://localhost:8081/ui/swagger, https://localhost:8082/ui/swagger, or https://localhost:8080/ui/swagger
```

#### Docker Secrets Example

All deployment configurations use Docker secrets for sensitive data:

```yaml
services:
  cryptoutil-postgres-1:
    image: cryptoutil:latest
    secrets:
      - postgres_url.secret
      - unseal_1of5.secret
      - unseal_2of5.secret
      - unseal_3of5.secret
    command:
      - -u
      - file:///run/secrets/postgres_url.secret
      - --unseal-key
      - file:///run/secrets/unseal_1of5.secret
      - --unseal-key
      - file:///run/secrets/unseal_2of5.secret
      - --unseal-key
      - file:///run/secrets/unseal_3of5.secret

secrets:
  postgres_url.secret:
    file: ./secrets/postgres_url.secret
  unseal_1of5.secret:
    file: ./secrets/unseal_1of5.secret
  unseal_2of5.secret:
    file: ./secrets/unseal_2of5.secret
  unseal_3of5.secret:
    file: ./secrets/unseal_3of5.secret
```

For complete examples, migration steps, and troubleshooting, see the [Docker Secrets Pattern Guide](docs/docker-secrets-pattern.md).

### Running with Go (Development)

```sh
# Clone and setup
git clone https://github.com/justincranford/cryptoutil
cd cryptoutil
go mod tidy

# Generate OpenAPI code
go generate ./...

# Run with PostgreSQL
docker compose up -d postgres
go run main.go --config=./configs/sm/config-pg-1.yml

# Or run with SQLite (development mode)
go run main.go --dev --config=./configs/sm/config-sqlite-1.yml
```

### API Access

#### KMS Server APIs

- **Swagger UI**: <https://localhost:8081/ui/swagger> (PostgreSQL instance 1), <https://localhost:8082/ui/swagger> (PostgreSQL instance 2), or <https://localhost:8080/ui/swagger> (SQLite)
- **OpenAPI Spec JSON**: <https://localhost:8081/ui/swagger/doc.json>, <https://localhost:8082/ui/swagger/doc.json>, or <https://localhost:8080/ui/swagger/doc.json>
- **Browser API**: <https://localhost:8081/browser/api/v1/*>, <https://localhost:8082/browser/api/v1/*>, or <https://localhost:8080/browser/api/v1/*>
- **Service API**: <https://localhost:8081/service/api/v1/*>, <https://localhost:8082/service/api/v1/*>, or <https://localhost:8080/service/api/v1/*>
- **Health Checks**: <https://localhost:9090/admin/v1/livez>, <https://localhost:9090/admin/v1/readyz>

#### Identity System APIs

- **AuthZ Service** (OAuth 2.1 Authorization Server):
  - **Base URL**: <https://localhost:8080>
  - **Swagger UI**: <https://localhost:8080/ui/swagger>
  - **OpenAPI Spec**: <https://localhost:8080/ui/swagger/doc.json>
  - **OAuth 2.1 Endpoints**: `/oauth2/v1/authorize`, `/oauth2/v1/token`, `/oauth2/v1/introspect`, `/oauth2/v1/revoke`
  - **Health**: `/health`
  - **Documentation**: See [OpenAPI Guide](docs/02-identityV2/historical/openapi-guide.md) for detailed API documentation

- **IdP Service** (OpenID Connect Identity Provider):
  - **Base URL**: <https://localhost:8081>
  - **Swagger UI**: <https://localhost:8081/ui/swagger>
  - **OpenAPI Spec**: <https://localhost:8081/ui/swagger/doc.json>
  - **OIDC Endpoints**: `/oidc/v1/login`, `/oidc/v1/consent`, `/oidc/v1/userinfo`, `/oidc/v1/logout`
  - **Health**: `/health`
  - **Documentation**: See [OpenAPI Guide](docs/02-identityV2/historical/openapi-guide.md) for detailed API documentation

- **RS Service** (OAuth 2.1 Resource Server):
  - **Base URL**: <https://localhost:8082>
  - **Swagger UI**: <https://localhost:8082/ui/swagger>
  - **OpenAPI Spec**: <https://localhost:8082/ui/swagger/doc.json>
  - **API Endpoints**: `/api/v1/public/health`, `/api/v1/protected/resource`, `/api/v1/admin/*`
  - **Health**: `/api/v1/public/health`
  - **Documentation**: See [OpenAPI Guide](docs/02-identityV2/historical/openapi-guide.md) for detailed API documentation

#### JOSE Authority Server APIs

- **JOSE Authority Service** (JOSE Cryptographic Operations):
  - **Base URL**: <https://localhost:8060>
  - **Swagger UI**: <https://localhost:8060/ui/swagger>
  - **OpenAPI Spec**: <https://localhost:8060/ui/swagger/doc.json>
  - **API Endpoints**:
    - `/jose/v1/sign` - Sign data with JWS
    - `/jose/v1/verify` - Verify JWS signatures
    - `/jose/v1/encrypt` - Encrypt data with JWE
    - `/jose/v1/decrypt` - Decrypt JWE data
    - `/jose/v1/keys` - JWKS key management
  - **Health**: `/health`

#### Certificate Authority APIs

- **CA Service** (X.509 Certificate Authority):
  - **Base URL**: <https://localhost:8050>
  - **Swagger UI**: <https://localhost:8050/ui/swagger>
  - **OpenAPI Spec**: <https://localhost:8050/ui/swagger/doc.json>
  - **API Endpoints**:
    - `/ca/v1/certificates` - Certificate lifecycle management
    - `/ca/v1/csr` - Certificate signing request operations
    - `/ca/v1/revoke` - Certificate revocation
    - `/ca/v1/crl` - Certificate revocation list
    - `/ca/v1/ocsp` - Online Certificate Status Protocol
  - **Health**: `/health`

#### Observability

- **Grafana UI**: <http://localhost:3000> (admin/admin)
- **OpenTelemetry Collector**:
  - **OTLP gRPC**: <http://localhost:4317> (receive telemetry from applications)
  - **OTLP HTTP**: <http://localhost:4318> (receive telemetry from applications)
  - **Self-metrics**: <http://localhost:8888/metrics> (OpenTelemetry Prometheus format)
  - **Received-metrics**: <http://localhost:8889/metrics> (OpenTelemetry Prometheus format, for re-export)
  - **Health Check**: <http://127.0.0.1:13133/>
  - **pprof**: <http://localhost:1777> (performance profiling)
  - **zPages**: <http://localhost:55679> (debugging UI)

### Example API Usage

```sh
# Get CSRF token (for browser API)
curl -k https://localhost:8080/browser/api/v1/csrf-token

# Create an elastic key (service API)
curl -k -X POST https://localhost:8080/service/api/v1/elastickey \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key", "algorithm": "A256GCM/A256KW", "provider": "Internal", "description": "Test key"}'

# Encrypt data
curl -k -X POST https://localhost:8080/elastickey/{elasticKeyID}/encrypt \
  -H "Content-Type: text/plain" \
  -d 'Hello World'
```

## Configuration

cryptoutil uses hierarchical configuration supporting multiple sources:

### Configuration Files (YAML)

```yaml
# Example: postgresql.yml
bind_public_address: "0.0.0.0"
bind_public_port: 8080
bind_private_address: "127.0.0.1"
bind_private_port: 9090
browser_api_context_path: "/browser/api/v1"
service_api_context_path: "/service/api/v1"
database_url: "postgres://user:pass@localhost:5432/cryptoutil"
allowed_ips: ["127.0.0.1", "::1"]
allowed_cidrs: ["10.0.0.0/8", "192.168.0.0/16"]
browser_ip_rate_limit: 100
service_ip_rate_limit: 25

# Security Configuration
cors_allowed_origins: "https://app.example.com"
csrf_token_name: "csrf_token"
csrf_token_same_site: "Strict"
csrf_token_cookie_secure: true

# Unseal Configuration
unseal_mode: "shamir"  # simple | shamir | system
unseal_files:
  - "/run/secrets/unseal_1of5"
  - "/run/secrets/unseal_2of5"
  - "/run/secrets/unseal_3of5"
```

### Command Line Parameters

```sh
# Key configuration options
go run main.go \
  --config=config.yaml \
  --dev \
  --verbose \
  --bind-public-port=8080 \
  --bind-private-port=9090 \
  --database-url="postgres://..." \
  --log-level=DEBUG
```

### Security Configuration

- **IP Allowlisting**: Configure `allowed_ips` and `allowed_cidrs` for production
- **Rate Limiting**: Set conservative `browser_ip_rate_limit` (10-100 requests/second per IP) and `service_ip_rate_limit` (10-25 requests/second per IP)
- **CORS**: Configure specific origins, avoid wildcards in production
- **CSRF**: Use `csrf_token_cookie_secure: true` and `csrf_token_same_site: "Strict"`
- **TLS**: Always use HTTPS in production (`bind_public_protocol: "https"`)
- **Database**: Always use `sslmode=require` for PostgreSQL connections

## Testing

### Automated Tests

```sh
# Run all tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Open coverage report in browser
start coverage.html  # Windows
open coverage.html   # macOS
xdg-open coverage.html  # Linux
```

### Manual Testing

```sh
# Start server
go run main.go --dev --verbose

# Test with Swagger UI (includes CSRF handling)
# Open in browser:
start https://localhost:8080/ui/swagger      # Windows
open https://localhost:8080/ui/swagger       # macOS
xdg-open https://localhost:8080/ui/swagger   # Linux

# Test with curl (service API - no CSRF needed)
curl -k -X GET https://localhost:8080/service/api/v1/elastickeys

# Test health endpoints
curl -k https://localhost:9090/admin/v1/livez
curl -k https://localhost:9090/admin/v1/readyz
```

### Integration Testing

```sh
# Run with test containers
go run cmd/pgtest/main.go  # PostgreSQL integration tests
```

### Fuzz Testing

This project uses **Go fuzz testing** for property-based testing to find edge cases and potential crashes in cryptographic functions.

#### Automated Execution (CI/CD)

Fuzz tests run automatically in GitHub Actions on:

- Push to `main` branch
- Pull requests

#### Fuzz Test Coverage

**Key Generation Functions (7 fuzz tests):**

- `FuzzGenerateRSAKeyPair` - RSA key pair generation
- `FuzzGenerateECDSAKeyPair` - ECDSA key pair generation
- `FuzzGenerateECDHKeyPair` - ECDH key pair generation
- `FuzzGenerateEdDSAKeyPair` - EdDSA key pair generation
- `FuzzGenerateAESKey` - AES key generation
- `FuzzGenerateAESHSKey` - AES-HS key generation
- `FuzzGenerateHMACKey` - HMAC key generation

**Digest Functions (9 fuzz tests):**

- `FuzzHKDF` - HMAC-based Key Derivation Function
- `FuzzHKDFWithSHA256` - HKDF with SHA-256
- `FuzzHKDFWithSHA384` - HKDF with SHA-384
- `FuzzHKDFWithSHA512` - HKDF with SHA-512
- `FuzzSHA256` - SHA-256 hashing
- `FuzzSHA384` - SHA-384 hashing
- `FuzzSHA512` - SHA-512 hashing
- `FuzzSHA3_256` - SHA3-256 hashing
- `FuzzSHA3_512` - SHA3-512 hashing

#### Manual Execution

```sh
# Run specific fuzz tests (use regex anchors for exact matching)
go test -fuzz=^FuzzHKDF$ -fuzztime=5s ./internal/common/crypto/digests/
go test -fuzz=^FuzzGenerateRSAKeyPair$ -fuzztime=5s ./internal/common/crypto/keygen/

# Run fuzz tests for 1 minute (CI/CD duration)
go test -fuzz=^FuzzHKDF$ -fuzztime=1m ./internal/common/crypto/digests/

# Quick verification during development
go test -fuzz=^FuzzHKDF$ -fuzztime=5s ./internal/common/crypto/digests/
```

**Important**: Always use regex anchors (`^FuzzXXX$`) for exact function matching when multiple fuzz tests exist in the same package, as Go's `-fuzz` flag uses prefix matching by default.

#### Fuzz Test Organization

- **File naming**: `*_fuzz_test.go` (separate from unit tests)
- **Function naming**: `FuzzXXX` where XXX describes the function being fuzzed
- **Duration**: 5 seconds for development, 1 minute for CI/CD
- **Coverage**: Focus on cryptographic functions with complex input handling

### Docker Compose Cleanup

After testing with Docker Compose, always clean up properly to ensure clean state for subsequent tests:

```sh
# Stop services and remove containers/networks
docker compose down

# Stop services, remove containers/networks, AND volumes (for complete cleanup)
docker compose down --volumes

# Remove everything including images (use with caution)
docker compose down --volumes --rmi all
```

**Always use `docker compose down --volumes` before starting new tests** after `compose.yml` changes to ensure:

- No state interference from previous tests
- Fresh database state for each test execution
- Proper resource management in CI/CD environments

## CI/CD Testing Pipeline

This project implements a comprehensive automated testing pipeline with 6 specialized workflows covering quality assurance, security testing, robustness validation, and integration testing.

### CI/CD Workflow Overview

The CI/CD pipeline is organized into 6 specialized workflows, each with different service orchestration and connectivity verification approaches:

| Workflow | File | Services Started | Connectivity Verification | Purpose |
|----------|------|------------------|---------------------------|---------|
| **Quality** | `ci-quality.yml` | None | N/A | Code quality, linting, formatting, container builds |
| **SAST** | `ci-sast.yml` | None | N/A | Static security analysis (Staticcheck, Govulncheck, Trivy, CodeQL) |
| **Robustness** | `ci-robust.yml` | None | N/A | Concurrency tests, race detection, fuzz tests, benchmarks |
| **DAST** | `ci-dast.yml` | Standalone app + PostgreSQL | Bash/curl verification | Dynamic security scanning (ZAP, Nuclei) |
| **E2E** | `ci-e2e.yml` | Full Docker Compose stack | Go test infrastructure | Complete system integration testing |
| **Load** | `ci-load.yml` | Full Docker Compose stack | Bash/curl verification | Performance testing with Gatling |

#### Service Connectivity Verification Techniques

Three distinct techniques are used across the workflows for service readiness verification:

**1. No Verification Required** (ci-quality.yml, ci-sast.yml, ci-robust.yml)

- **Approach**: Static analysis and unit tests don't start services
- **Rationale**: No runtime connectivity needed for code analysis or isolated tests

**2. Go-Based E2E Infrastructure** (ci-e2e.yml)

- **Approach**: Go test suite orchestrates Docker Compose with comprehensive health checks
- **Components**:
  - `infrastructure.go`: Docker Compose orchestration and health monitoring
  - `docker_health.go`: Service health status parsing and validation
  - `http_utils.go`: HTTP client creation with TLS configuration
- **Health Verification Flow**:
  1. Docker health checks: `docker compose ps --format json` → parse service health status
  2. HTTP connectivity: Go HTTP client with `InsecureSkipVerify` for self-signed certs
  3. Multi-endpoint verification: Public APIs, health endpoints, Swagger UI, OTEL collector, Grafana
- **Key Features**:
  - Handles 3 service types: standalone jobs, services with native health checks, services with healthcheck jobs
  - Exponential backoff with configurable timeouts
  - Comprehensive error reporting with detailed health status

**3. Bash/Curl Verification** (ci-dast.yml, ci-load.yml)

- **Approach**: Shell scripts with curl for HTTPS endpoints with self-signed certificates
- **Pattern**:

  ```bash
  # CORRECT: curl with -s (silent), -k (insecure/skip cert verification), -f (fail on HTTP errors)
  curl -skf --connect-timeout 10 --max-time 15 "$url" -o /tmp/response.json

  # INCORRECT: wget does not reliably verify HTTPS with self-signed certs
  wget --no-check-certificate --spider "$url"  # ❌ FAILS
  ```

- **Health Verification Flow**:
  1. Retry loop with exponential backoff (max 30 attempts, 5s max backoff)
  2. Verify response body is non-empty (successful connection indicator)
  3. Check all cryptoutil instances: `https://127.0.0.1:8080`, `8081`, `8082`
- **Key Features**:
  - Works with self-signed TLS certificates (`-k` flag)
  - Verifies actual response data (not just HTTP status codes)
  - Exponential backoff prevents overwhelming services during startup

**Common Mistakes to Avoid**:

- ❌ Using `wget` for HTTPS with self-signed certs (unreliable)
- ❌ Using `localhost` in workflows (use `127.0.0.1` for explicit IPv4, see localhost-vs-ip.instructions.md)
- ❌ Checking only HTTP status codes without verifying response body
- ❌ Missing exponential backoff (hammers services during startup)
- ❌ Insufficient timeouts for containerized environments (use 10-15s)

### Quality Assurance (ci-quality.yml)

Automated code quality and container security validation:

#### Mutation Testing

This project uses **Gremlins mutation testing** to assess test suite quality by introducing artificial faults and measuring test detection rates.

**Automated Execution (CI/CD):**

- Runs automatically on push to `main` branch and pull requests
- Targets high-coverage packages with mutation operators
- Generates mutation testing reports and coverage metrics

**Manual Execution:**

```sh
# Run mutation tests on specific packages
gremlins unleash --paths=./internal/common/crypto/keygen/
gremlins unleash --paths=./internal/common/crypto/digests/
```

#### Container Security & SBOM

Container vulnerability scanning and software bill of materials generation:

**Automated Analysis (CI/CD):**

- **Docker Scout**: Container image vulnerability scanning
- **SBOM Generation**: Software Bill of Materials creation using Syft
- **Security Policy**: Automated security gate for container deployments

### Security Testing

#### Static Application Security Testing (SAST) (ci-sast.yml)

Comprehensive static analysis using multiple security-focused tools:

**Automated SAST Pipeline (CI/CD):**

- **Staticcheck**: Advanced static analysis for Go code quality and security
- **Govulncheck**: Official Go vulnerability database scanning
- **Trivy**: Container and filesystem vulnerability scanning
- **CodeQL**: Semantic code analysis (Go, JavaScript, Python)

**Manual SAST Execution:**

```sh
# Run individual SAST tools
staticcheck ./...
govulncheck ./...
trivy filesystem .
```

#### Dynamic Application Security Testing (DAST) (ci-dast.yml)

Runtime security testing with active vulnerability scanning:

**Automated DAST Pipeline (CI/CD):**

- **OWASP ZAP**: Comprehensive web application security scanning
- **Nuclei**: Fast template-based vulnerability detection
- Scans all cryptoutil service instances (SQLite, PostgreSQL instances)

**DAST Scan Profiles:**

- **Quick**: Basic security checks (3-5 minutes)
- **Full**: Comprehensive scanning (10-15 minutes)
- **Deep**: Exhaustive security assessment (20-25 minutes)

### Robustness Testing (ci-robust.yml)

Thread safety and performance validation:

#### Concurrency & Race Detection

Thread safety validation with Go's race detector:

**Automated Testing (CI/CD):**

- Runs all tests with race detection enabled (`-race` flag)
- Identifies potential race conditions and data races
- Validates thread-safe cryptographic operations

#### Benchmark Testing

Performance regression detection and optimization validation:

**Automated Testing (CI/CD):**

- Executes performance benchmarks across all packages
- Measures cryptographic operation throughput and latency
- Detects performance regressions between commits

### Integration Testing

#### Load Testing (ci-load.yml)

Performance validation using **Gatling** load testing framework:

**Automated Load Testing (CI/CD):**

- Runs comprehensive load tests against cryptoutil APIs
- Measures response times, throughput, and error rates
- Generates detailed performance reports and metrics

**Manual Load Testing:**

```sh
# Run load tests from project root
cd test/load
mvnw gatling:test
# Results available in target/gatling/
```

#### End-to-End Testing (ci-e2e.yml)

Full system validation using Docker Compose orchestration:

**Automated E2E Pipeline (CI/CD):**

- Deploys complete cryptoutil stack (PostgreSQL, services, observability)
- Executes comprehensive API test suites
- Validates service-to-service communication and data flow
- Generates E2E test reports and failure analysis

## Security Testing

### Manual Nuclei Vulnerability Scanning

The project includes comprehensive vulnerability scanning using [Nuclei](https://github.com/projectdiscovery/nuclei), a fast, template-based vulnerability scanner.

#### Prerequisites: Start cryptoutil Services

Before running nuclei scans, start the cryptoutil services using Docker Compose:

```sh
# Clean up any existing containers and volumes
docker compose -f ./deployments/compose/compose.yml down -v

# Start all services (PostgreSQL, cryptoutil instances, observability stack)
docker compose -f ./deployments/compose/compose.yml up -d

# Wait for services to be ready (check health endpoints)
curl -k https://localhost:8080/ui/swagger/doc.json  # SQLite instance
curl -k https://localhost:8081/ui/swagger/doc.json  # PostgreSQL instance 1
curl -k https://localhost:8082/ui/swagger/doc.json  # PostgreSQL instance 2
```

#### Manual Nuclei Scan Examples

**Quick Security Scan (Info/Low severity, fast):**

```sh
# Scan all three cryptoutil instances
nuclei -target https://localhost:8080/ -severity info,low
nuclei -target https://localhost:8081/ -severity info,low
nuclei -target https://localhost:8082/ -severity info,low
```

**Comprehensive Security Scan (All severities):**

```sh
# Full vulnerability assessment
nuclei -target https://localhost:8080/ -severity info,low,medium,high,critical
nuclei -target https://localhost:8081/ -severity info,low,medium,high,critical
nuclei -target https://localhost:8082/ -severity info,low,medium,high,critical
```

**Targeted Scan Examples:**

```sh
# Scan for specific vulnerability types
nuclei -target https://localhost:8080/ -tags cves,vulnerabilities
nuclei -target https://localhost:8081/ -tags security-misconfiguration,exposure
nuclei -target https://localhost:8082/ -tags tech-detect,misc

# Scan with custom concurrency and rate limiting
nuclei -target https://localhost:8080/ -c 10 -rl 50 -severity high,critical
```

**Batch Scan All Services:**

```sh
# PowerShell: Scan all three services sequentially
foreach ($port in @(8080, 8081, 8082)) {
    Write-Host "Scanning https://localhost:$port/" -ForegroundColor Green
    nuclei -target "https://localhost:$port/" -severity medium,high,critical
}
```

#### Service Endpoints

| Service | Port | Backend | Purpose |
|---------|------|---------|---------|
| cryptoutil-sqlite | 8080 | SQLite | Development/testing instance |
| cryptoutil-postgres-1 | 8081 | PostgreSQL | Production-like instance #1 |
| cryptoutil-postgres-2 | 8082 | PostgreSQL | Production-like instance #2 |

#### Expected Results

- **✅ No vulnerabilities found**: Indicates proper security implementation
- **⚠️ Findings detected**: Review and address security issues
- **🔄 Scan duration**: 5-30 seconds per service depending on scan profile

#### Troubleshooting

**Nuclei templates not found:**

```sh
# Install/update nuclei templates
nuclei -update-templates
```

**Services not responding:**

```sh
# Check service health
curl -k https://localhost:8080/ui/swagger/doc.json
curl -k https://localhost:9090/admin/v1/livez  # Admin health endpoint
```

**Clean restart:**

```sh
docker compose -f ./deployments/compose/compose.yml down -v
docker compose -f ./deployments/compose/compose.yml up -d
```

## Development

### Automated Code Formatting

This project uses **automated code formatting** that runs on every commit. The formatting is enforced in CI/CD.

**Setup (Required for Contributors):**

```sh
# Install pre-commit hooks (runs comprehensive code quality checks)
pip install pre-commit
# If pip is not in PATH, use:
# python -m pip install pre-commit

pre-commit install
# If pre-commit is not in PATH, use:
# python -m pre_commit install

# Set consistent cache location (Windows)
setx PRE_COMMIT_HOME "C:\Users\%USERNAME%\.cache\pre-commit"

# Test the setup
pre-commit run --all-files
# If pre-commit is not in PATH, use:
# python -m pre_commit run --all-files
```

**Setup Instructions:**
See [docs/DEV-SETUP.md](docs/DEV-SETUP.md) for comprehensive pre-commit setup instructions covering Windows, Linux, and macOS.

**What Gets Checked Automatically:**

- **File formatting**: End-of-file fixes, trailing whitespace removal
- **Syntax validation**: YAML, JSON, GitHub Actions workflows, Dockerfiles
- **Go tools**: `gofumpt` (strict formatting), `goimports` (import organization), `errcheck` (error checking), `go build`
- **Security**: Large file prevention, merge conflict detection
- **Linting**: `golangci-lint` with automatic fixes for supported linters (e.g., WSL whitespace consistency: `golangci-lint run --enable-only=wsl --fix`)
  - **Auto-fixable linters**: wsl, gofmt, goimports, godot, goconst, importas, copyloopvar, testpackage, revive
  - **Manual-only linters**: errcheck, gosimple, govet, ineffassign, staticcheck, unused, gosec, noctx, wrapcheck, thelper, tparallel, gomodguard, prealloc, bodyclose, errorlint, stylecheck
