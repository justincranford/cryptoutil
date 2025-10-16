# cryptoutil

cryptoutil is a production-ready embedded Key Management System (KMS) and cryptographic service with enterprise-grade security features. It implements a hierarchical cryptographic architecture following NIST FIPS 140-3 standards.

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
- **Per-IP rate limiting** with configurable thresholds
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
Grafana-OTEL-LGTM (Prometheus) → OpenTelemetry Collector Contrib (HTTP:8889/metrics)
```
- **Purpose**: Monitor collector health and performance
- **Protocol**: Prometheus scraping - pull-based
- **Data**: Collector throughput, error rates, queue depths, resource usage

**Why Both Flows?** The collector both **receives application telemetry** (from cryptoutil) and **exposes its own metrics** (for monitoring). This provides complete observability of both your application and the telemetry pipeline itself.

**Services:**
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: Integrated into Grafana-OTEL-LGTM stack
- **Loki**: Integrated log aggregation
- **Tempo**: Integrated trace storage
- **OpenTelemetry Collector**: Receives telemetry from cryptoutil services

### 🏗️ Production Ready
- **Database support**: PostgreSQL (production), SQLite (development/testing)
- **Container deployment**: Docker Compose with secret management
- **Configuration management**: YAML files + CLI parameters
- **Graceful shutdown**: Signal handling and connection draining

## Quick Start

### Prerequisites
- Go 1.25.1+
- Docker and Docker Compose (for PostgreSQL)

### Running with Docker Compose
```sh
# Start full stack: PostgreSQL, cryptoutil, and observability
cd deployments/compose
docker compose up -d

# View logs
docker compose logs -f cryptoutil_postgres

# Access services
# Grafana UI: http://localhost:3000 (admin/admin)
# cryptoutil API: http://localhost:8081 (PostgreSQL) or http://localhost:8080 (SQLite)
# Swagger UI: http://localhost:8081/ui/swagger or http://localhost:8080/ui/swagger
```

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
go run main.go --config=./deployments/compose/cryptoutil/postgresql.yml

# Or run with SQLite (development mode)
go run main.go --dev --config=./deployments/compose/cryptoutil/sqlite.yml
```

### API Access
- **Swagger UI**: http://localhost:8081/ui/swagger (PostgreSQL) or http://localhost:8080/ui/swagger (SQLite)
- **Browser API**: http://localhost:8081/browser/api/v1/* or http://localhost:8080/browser/api/v1/*
- **Service API**: http://localhost:8081/service/api/v1/* or http://localhost:8080/service/api/v1/*
- **Health Checks**: http://localhost:9090/livez, http://localhost:9090/readyz
- **Grafana UI**: http://localhost:3000 (admin/admin)
- **OpenTelemetry Collector Metrics**: http://localhost:8888/metrics, http://localhost:8889/metrics

### Example API Usage
```sh
# Get CSRF token (for browser API)
curl http://localhost:8080/browser/api/v1/csrf-token

# Create an elastic key (service API)
curl -X POST http://localhost:8080/service/api/v1/elastickey \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key", "algorithm": "RSA", "provider": "CRYPTOUTIL"}'

# Encrypt data
curl -X POST http://localhost:8080/service/api/v1/crypto/encrypt \
  -H "Content-Type: application/json" \
  -d '{"elasticKeyId": "key-id", "plaintext": "SGVsbG8gV29ybGQ="}'
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
ip_rate_limit: 100

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
- **Rate Limiting**: Set conservative `ip_rate_limit` (10-100 requests/second per IP)
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
start http://localhost:8080/ui/swagger      # Windows
open http://localhost:8080/ui/swagger       # macOS
xdg-open http://localhost:8080/ui/swagger   # Linux

# Test with curl (service API - no CSRF needed)
curl -X GET http://localhost:8080/service/api/v1/elastickeys

# Test health endpoints
curl http://localhost:9090/livez
curl http://localhost:9090/readyz
```

### Integration Testing
```sh
# Run with test containers
go run cmd/pgtest/main.go  # PostgreSQL integration tests
```

### Mutation Testing

This project uses [Gremlins](https://github.com/go-gremlins/gremlins) for mutation testing to validate test suite quality by introducing small code changes and verifying tests catch them.

#### Prerequisites
```bash
# Install Gremlins
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
```

#### Automated Execution (CI/CD)
Mutation testing runs automatically on the `main` branch after all tests pass, focusing on high-coverage packages.

#### Manual Execution
```sh
# Linux/macOS - Test all high-coverage packages
./scripts/mutation-test.sh

# Windows PowerShell
.\scripts\mutation-test.ps1

# Test specific package
./scripts/mutation-test.sh --target ./internal/common/util/datetime/
.\scripts\mutation-test.ps1 -Target "./internal/common/util/datetime/"

# Dry run (analyze without testing)
./scripts/mutation-test.sh --dry-run
.\scripts\mutation-test.ps1 -DryRun
```

#### Direct Gremlins Commands
```bash
# Dry run (analyze without testing)
gremlins unleash --dry-run ./internal/common/util/datetime/

# Full mutation testing on a specific package
gremlins unleash ./internal/common/util/datetime/

# Test multiple packages with custom settings
gremlins unleash ./internal/common/util/... --workers 2 --timeout-coefficient 3
```

**Configuration**: Mutation testing uses `.gremlins.yaml` with quality thresholds (70% efficacy, 60% coverage) and enabled mutation operators (arithmetic, conditionals, increment/decrement, etc.).
### DAST Security Testing

This project uses **Dynamic Application Security Testing (DAST)** to identify runtime vulnerabilities in the running application. DAST complements static analysis by testing the application from the outside, simulating real-world attack scenarios.

#### Tools Used
- **OWASP ZAP**: Comprehensive web application security scanner (currently disabled, planned re-enablement)
- **Nuclei**: Fast, template-based vulnerability scanner for CVE detection and security misconfigurations

#### Dual API Context Testing
cryptoutil exposes identical OpenAPI operations under two distinct context paths with different security middleware:

| Context Path | Intended Clients | Security Features |
|--------------|------------------|-------------------|
| `/browser/api/v1/*` | Browser/web clients | CORS, CSRF protection, CSP, comprehensive security headers |
| `/service/api/v1/*` | Service-to-service clients | Core security only (no browser-specific headers) |

**Testing Note**: Always test both API contexts as they have different security header configurations.

#### Automated Execution (CI/CD)
DAST runs automatically in GitHub Actions on:
- Push to `main` branch
- Pull requests
- Weekly scheduled scans (Sundays)
- Manual workflow dispatch

**Scan Profiles**: Quick (2-3 min), Full (8-10 min), Deep (15-20 min) with different coverage levels.

#### Manual Execution
```sh
# Linux/macOS - Complete DAST scan
./scripts/dast.sh

# Windows PowerShell (use -ExecutionPolicy Bypass if needed)
.\scripts\dast.ps1

# Custom configuration and port
./scripts/dast.sh --config configs/test/config.yml --port 9090
.\scripts\dast.ps1 -Config "configs/test/config.yml" -Port 9090

# Skip ZAP, run only Nuclei
./scripts/dast.sh --skip-zap
.\scripts\dast.ps1 -SkipZap

# Custom output directory
./scripts/dast.sh --output-dir security-reports
.\scripts\dast.ps1 -OutputDir "security-reports"
```

#### Local Testing with act
```powershell
# Quick scan (2-3 minutes)
act workflow_dispatch -j dast-security-scan --input scan_profile=quick

# Full scan (8-10 minutes)
act workflow_dispatch -j dast-security-scan --input scan_profile=full

# Deep scan (15-20 minutes)
act workflow_dispatch -j dast-security-scan --input scan_profile=deep
```

### Comprehensive Security Scanning
```sh
# Linux/macOS - Run all security scans locally
./scripts/security-scan.sh

# Windows PowerShell
.\scripts\security-scan.ps1

# Run specific scan types
./scripts/security-scan.sh --static-only              # Static analysis only
./scripts/security-scan.sh --vuln-only               # Vulnerability scans only
./scripts/security-scan.sh --container-only          # Container security only

# Windows equivalents
.\scripts\security-scan.ps1 -StaticOnly
.\scripts\security-scan.ps1 -VulnOnly
.\scripts\security-scan.ps1 -ContainerOnly

# Custom output directory and Docker image
./scripts/security-scan.sh --output-dir reports --image-tag cryptoutil:dev
.\scripts\security-scan.ps1 -OutputDir "reports" -ImageTag "cryptoutil:dev"

# Skip Docker-based scans (if Docker unavailable)
./scripts/security-scan.sh --skip-docker
.\scripts\security-scan.ps1 -SkipDocker
```

**Security Tools Included:**
- **Staticcheck**: Go static analysis and lint checking
- **golangci-lint**: Comprehensive Go linting with multiple analyzers
- **govulncheck**: Official Go vulnerability database scanning
- **Trivy**: File system and container vulnerability scanning
- **Docker Scout**: Advanced container security analysis and recommendations

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

**Automated Setup (Recommended):**
```sh
# Windows Batch
.\scripts\setup-pre-commit.bat

# Windows PowerShell
.\scripts\setup-pre-commit.ps1
```

**What Gets Checked Automatically:**
- **File formatting**: End-of-file fixes, trailing whitespace removal
- **Syntax validation**: YAML, JSON, GitHub Actions workflows, Dockerfiles
- **Go tools**: `gofumpt` (strict formatting), `goimports` (import organization), `errcheck` (error checking), `go build`
- **Security**: Large file prevention, merge conflict detection
- **Linting**: `golangci-lint` with automatic fixes for supported linters (e.g., WSL whitespace consistency: `golangci-lint run --enable-only=wsl --fix`)
  - **Auto-fixable linters**: wsl, gofmt, goimports, godot, goconst, importas, copyloopvar, testpackage, revive
  - **Manual-only linters**: errcheck, gosimple, govet, ineffassign, staticcheck, unused, gosec, noctx, wrapcheck, thelper, tparallel, gomodguard, prealloc, bodyclose, errorlint, stylecheck
