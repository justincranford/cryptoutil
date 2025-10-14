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
# Install pre-commit hooks (runs gofumpt + goimports automatically)
pip install pre-commit
pre-commit install

# Set consistent cache location (Windows)
setx PRE_COMMIT_HOME "C:\Users\%USERNAME%\.cache\pre-commit"

# Test the setup
pre-commit run --all-files
```

**Automated Setup (Recommended):**
```sh
# Windows Batch
.\scripts\setup-pre-commit.bat

# Windows PowerShell
.\scripts\setup-pre-commit.ps1
```

**What Gets Formatted Automatically:**
- `gofumpt` - Stricter Go code formatting (better than standard `gofmt`)
- `goimports` - Automatic import organization and formatting
- Trailing whitespace removal
- File ending fixes

**Manual Formatting (if needed):**
```sh
gofumpt -extra -w .        # Format all Go files
goimports -w .      # Organize imports
```

### Code Generation
```sh
# Install oapi-codegen
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1

# Generate OpenAPI code
go generate ./...
```

The generate command runs oapi-codegen using configurations in [api/generate.go](api/generate.go) to create:
- `api/model/` - Data models
- `api/server/` - HTTP handlers
- `api/client/` - Go client

### Linting & Formatting

#### Automated Formatting (Recommended)
```sh
# Install pre-commit for automatic formatting on every commit
pip install pre-commit
pre-commit install

# Run formatting on all files manually
pre-commit run --all-files
```

#### Manual Tools
```sh
# Install tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
go install mvdan.cc/gofumpt@v0.7.0
go install golang.org/x/tools/cmd/goimports@latest

# Run linters and formatters
golangci-lint run
gofumpt -extra -l -w .
goimports -l -w .
```

### Project Structure
```
├── cmd/                    # Main applications
│   ├── cryptoutil/         # Main server application
│   └── pgtest/             # PostgreSQL integration tests
├── internal/               # Private application code
│   ├── server/             # HTTP server and business logic
│   ├── common/             # Shared utilities (crypto, config, etc.)
│   └── openapi/            # Generated API code
├── api/                    # OpenAPI specifications
├── configs/                # Configuration templates
├── deployments/            # Docker and deployment files
│   ├── Dockerfile          # Container image definition
│   └── compose/            # Docker Compose setup
│       ├── compose.yml     # Docker Compose configuration
│       ├── cryptoutil/     # Application secrets and configs
│       └── postgres/       # Database secrets
└── docs/                   # Additional documentation
```

### Dependency and Action Version Checking

This project includes automated tools to check for outdated dependencies and GitHub Actions versions, helping maintain security and compatibility.

#### Automated Execution (CI/CD)
Version checks run automatically in GitHub Actions on:
- Push to `main` branch
- Pull requests
- Manual workflow dispatch

#### Manual Execution
```sh
# Check Go dependency versions
go run scripts/cicd_utils.go go-dependency-versions

# Check GitHub Actions versions in workflows
go run scripts/cicd_utils.go github-action-versions

# Check both (run both commands)
go run scripts/cicd_utils.go go-dependency-versions
go run scripts/cicd_utils.go github-action-versions
```

#### What Gets Checked
- **Go Dependencies**: Compares current versions against latest available on Go module proxy
- **GitHub Actions**: Scans `.github/workflows/*.yml` files for action references and checks against latest releases/tags

#### Action Version Exceptions

For cases where the latest action version has breaking changes or compatibility issues, you can configure exceptions in `.github/workflows-outdated-action-exemptions.json`:

```json
{
  "exceptions": {
    "actions/checkout": {
      "allowed_versions": ["v4.1.7"],
      "reason": "Known stable version, v5.0.0 has Node.js compatibility issues in some environments"
    },
    "docker/build-push-action": {
      "allowed_versions": ["v5.4.0"],
      "reason": "Compatibility with current Docker daemon version"
    }
  }
}
```

The check script will skip validation for actions listed in the exceptions file and display them as "exempted" rather than outdated.

## Architecture Overview

### API Context Separation
- **Browser Context** (`/browser/api/v1/*`): Full browser security stack
  - CORS headers for cross-origin requests
  - CSRF token validation
  - Content Security Policy (CSP)
  - XSS protection headers
- **Service Context** (`/service/api/v1/*`): Streamlined for services
  - No browser-specific middleware
  - Optimized for machine-to-machine communication
- **Management Interface** (`localhost:9090`): Private operations
  - Health checks (`/livez`, `/readyz`)
  - Graceful shutdown (`/shutdown`)

### Security Layers
1. **Network Security**: IP allowlisting, rate limiting
2. **Transport Security**: TLS with auto-generated certificates
3. **Application Security**: CORS, CSRF, CSP, security headers
4. **Cryptographic Security**: FIPS 140-3 algorithms, hierarchical keys
5. **Operational Security**: Audit logging, secure failure modes

### Key Management Hierarchy
```
┌─────────────────┐
│   Unseal Keys   │ ← System initialization
└─────────────────┘
         │
┌─────────────────┐
│   Root Keys     │ ← Encrypted by unseal keys
└─────────────────┘
         │
┌─────────────────┐
│Intermediate Keys│ ← Encrypted by root keys
└─────────────────┘
         │
┌─────────────────┐
│ Content Keys    │ ← Material encryption keys
└─────────────────┘
```

## Production Deployment

### Docker Compose (Recommended)
```sh
cd deployments/compose
docker compose up -d
```

### Manual Docker Build
```sh
# Build with proper versioning (recommended)
docker build \
  --build-arg VCS_REF=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg APP_VERSION=dev \
  -t cryptoutil \
  -f deployments/Dockerfile .

# Or use the simple build (with default values)
DOCKER_BUILDKIT=1 docker build -t cryptoutil -f deployments/Dockerfile .
```

**Runtime Metadata Files:**
- `/app/.build-params`: Build arguments used (APP_VERSION, VCS_REF, BUILD_DATE)
- `/app/.vcs-ref`: Actual git commit hash from build time
- `/app/.build-date`: Actual build timestamp (ISO 8601 format)
- `/app/.app-version`: Application version string

**Mandatory Build Arguments:**
- `APP_VERSION`: Application version (e.g., "v1.0.0") - **MANDATORY**
- `VCS_REF`: Git commit hash (use `$(git rev-parse HEAD)`) - **MANDATORY**
- `BUILD_DATE`: ISO 8601 timestamp (use `$(date -u +"%Y-%m-%dT%H:%M:%SZ")`) - **MANDATORY**

**Dockerfile enforces all mandatory arguments internally - build will fail with clear error and usage if missing.**

**Debugging:** Alpine base image provides shell access for troubleshooting.

This deploys:
- **PostgreSQL**: Persistent database with encrypted storage
- **cryptoutil**: Production-configured server with secrets management
- **Health Monitoring**: Automatic health checks and restarts

### Container Architecture
```
┌─────────────────────┐    ┌─────────────────────┐
│   cryptoutil        │    │    PostgreSQL       │
│   Port 8080 (HTTPS) │◄──►│   Port 5432         │
│   Port 9090 (HTTP)  │    │   Persistent Volume │
└─────────────────────┘    └─────────────────────┘
         │
         ▼
┌─────────────────────┐
│   Docker Secrets    │
│   • Database URL    │
│   • Unseal Keys     │
│   • Configuration   │
└─────────────────────┘
```

### Secret Management
```sh
# Create database secrets
echo "cryptoutil" > postgres/postgres_database.secret
echo "cryptoutil_user" > postgres/postgres_username.secret
echo "$(openssl rand -base64 32)" > postgres/postgres_password.secret

# Create unseal key secrets (M-of-N sharing)
for i in {1..5}; do
  openssl rand -base64 64 > cryptoutil/cryptoutil_unseal_${i}of5.secret
done
```

### Health Monitoring
```sh
# Check application health
curl http://localhost:9090/livez    # Liveness probe
curl http://localhost:9090/readyz   # Readiness probe

# Graceful shutdown
curl -X POST http://localhost:9090/shutdown
```

### Kubernetes Deployment
```yaml
# Basic Kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cryptoutil
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cryptoutil
  template:
    spec:
      containers:
      - name: cryptoutil
        image: cryptoutil:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        livenessProbe:
          httpGet:
            path: /livez
            port: 9090
        readinessProbe:
          httpGet:
            path: /readyz
            port: 9090
```


## Tooling: Local GitHub Actions Testing with act

You can run GitHub Actions workflows locally using [`act`](https://github.com/nektos/act):

### Install act (Windows)

1. Download the latest release from [nektos/act releases](https://github.com/nektos/act/releases).
2. Extract `act.exe` from the zip file.
3. Move `act.exe` to a directory in your PATH (e.g., `C:\Program Files\act`).
4. Ensure [Docker Desktop](https://www.docker.com/products/docker-desktop/) is running.
5. Open a new terminal and verify:
  ```powershell
  act --version
  ```

### Run a Workflow Locally


From the project root, run:
```powershell
# By default, DAST scans use TARGET_URL=https://host.docker.internal:8080 for Docker compatibility:
act -j dast-security-scan

The workflow now waits for either the admin API (9090, loopback) or the client API (8080, HTTPS, 0.0.0.0) to be ready before running DAST scans. This ensures ZAP can scan as soon as the externally accessible API is up, even if the admin API is not exposed.
```
This will execute the `dast-security-scan` job from `.github/workflows/dast.yml` using Docker containers.

> **Note:** The default DAST scan target is now `https://host.docker.internal:8080` for compatibility with Docker and act. If you need to override the target (e.g., in CI), set the `TARGET_URL` environment variable or use the workflow_dispatch input. Some steps may require adjustment for local paths, secrets, or service compatibility. See the [act documentation](https://github.com/nektos/act#usage) for advanced usage and troubleshooting.

## Documentation

- [Project Overview](docs/README.md) - Comprehensive architectural deep dive

## Contributing

1. **Install pre-commit hooks**: `pip install pre-commit && pre-commit install`
2. Follow the project layout in [.github/instructions/project-layout.instructions.md](.github/instructions/project-layout.instructions.md)
3. Use the coding standards in [.github/instructions/](.github/instructions/)
4. Ensure all tests pass: `go test ./... -cover`
5. Code formatting is **automatic** via pre-commit hooks (gofumpt + goimports)
6. Manual linting (if needed): `golangci-lint run`

**Note:** Code formatting (`gofumpt` + `goimports`) is enforced automatically on commit and verified in CI.

## License

See [LICENSE](LICENSE) file for details.
