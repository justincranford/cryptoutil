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

### 🏗️ Production Ready
- **Database support**: PostgreSQL (production), SQLite (development/testing)
- **Container deployment**: Docker Compose with secret management
- **Configuration management**: YAML files + CLI parameters
- **Graceful shutdown**: Signal handling and connection draining

## Quick Start

### Prerequisites
- Go 1.24+
- Docker and Docker Compose (for PostgreSQL)

### Running with Docker Compose
```sh
# Start PostgreSQL and cryptoutil
cd deployments/compose
docker compose up -d

# View logs
docker compose logs -f cryptoutil
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
- **Swagger UI**: http://localhost:8080/ui/swagger
- **Browser API**: http://localhost:8080/browser/api/v1/*
- **Service API**: http://localhost:8080/service/api/v1/*
- **Health Checks**: http://localhost:9090/livez, http://localhost:9090/readyz

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
- **IP Allowlisting**: Configure `allowed_ips` and `allowed_cidrs`
- **Rate Limiting**: Set `ip_rate_limit` (requests per second per IP)
- **CORS**: Configure origins, methods, headers for browser clients
- **CSRF**: Token-based protection for browser API context
- **TLS**: Automatic certificate generation for HTTPS endpoints

## Testing

### Automated Tests
```sh
# Run all tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Open coverage report
start coverage.html  # Windows
open coverage.html   # macOS
```

### Manual Testing
```sh
# Start server
go run main.go --dev --verbose

# Test with Swagger UI (includes CSRF handling)
start http://localhost:8080/ui/swagger

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

## Development

### Code Generation
```sh
# Install oapi-codegen
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Generate OpenAPI code
go generate ./...
```

The generate command runs oapi-codegen using configurations in [internal/openapi/generate.go](internal/openapi/generate.go) to create:
- `internal/openapi/model/` - Data models
- `internal/openapi/server/` - HTTP handlers
- `internal/openapi/client/` - Go client

### Linting & Formatting
```sh
# Install tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest

# Run linters
golangci-lint run
gofumpt -l -w .
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
└── docs/                   # Additional documentation
```

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

## Advanced Configuration

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

### Security Configuration Best Practices

#### Network Security
- **IP Allowlisting**: Configure `allowed_ips` and `allowed_cidrs` for production
- **Rate Limiting**: Set conservative `ip_rate_limit` (10-100 requests/second per IP)
- **TLS**: Always use HTTPS in production (`bind_public_protocol: "https"`)

#### Application Security
- **CORS**: Configure specific origins, avoid wildcards in production
- **CSRF**: Use `csrf_token_cookie_secure: true` and `csrf_token_same_site: "Strict"`
- **Database**: Always use `sslmode=require` for PostgreSQL connections

## Production Deployment

### Docker Compose (Recommended)
```sh
cd deployments/compose
docker compose up -d
```

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

## Deployment

### Docker Compose (Recommended)
```sh
cd deployments/compose
docker compose up -d
```

This starts:
- PostgreSQL database with persistent storage
- cryptoutil server with production configuration
- Automatic secret management via Docker secrets

### Configuration Files
- `postgresql.yml` - Production PostgreSQL setup
- `sqlite.yml` - Development SQLite setup
- Secrets managed via `deployments/compose/cryptoutil/*.secret`

### Health Monitoring
```sh
# Check application health
curl http://localhost:9090/livez    # Liveness probe
curl http://localhost:9090/readyz   # Readiness probe

# Graceful shutdown
curl -X POST http://localhost:9090/shutdown
```

### Kubernetes Deployment
The application includes Kubernetes-ready features:
- Health check endpoints for probes
- Graceful shutdown handling
- Structured logging for log aggregation
- OpenTelemetry metrics for monitoring

## Documentation

- [Project Overview](docs/README.md) - Comprehensive architectural deep dive

## Contributing

1. Follow the project layout in [.github/instructions/project-layout.instructions.md](.github/instructions/project-layout.instructions.md)
2. Use the coding standards in [.github/instructions/](.github/instructions/)
3. Ensure all tests pass: `go test ./... -cover`
4. Run linters: `golangci-lint run && gofumpt -l -w .`

## License

See [LICENSE](LICENSE) file for details.
