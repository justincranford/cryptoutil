# Cryptoutil Project

## Project Overview

**Cryptoutil** is a sophisticated, production-ready **embedded Key Management System (KMS)** written in Go that implements a hierarchical cryptographic architecture with enterprise-grade security features. The project follows modern software engineering practices and cryptographic standards.

## Key Architecture Components

### 1. **Core Design Philosophy**

- **FIPS 140-3 Compliance**: Only uses NIST-approved algorithms (RSA ≥2048, AES ≥128, NIST curves, EdDSA)
- **Defense in Depth**: Multi-layered security with barrier system, unsealing mechanisms, and encrypted key storage
- **API-First Design**: OpenAPI-driven development with automatic code generation
- **Cloud-Native**: Containerized deployment with Docker Compose and Kubernetes readiness

### 2. **Cryptographic Architecture**

The system implements a sophisticated multi-tier key hierarchy:

**Barrier System (Vault-like)**:

- **Unseal Keys**: Root-level keys for system initialization
- **Root Keys**: Master keys encrypted by unseal keys
- **Intermediate Keys**: Secondary encryption layer
- **Content Keys**: Material key encryption keys

**Key Types Supported**:

- **Elastic Keys**: Logical key containers with metadata and policies
- **Material Keys**: Actual cryptographic keys (versioned within Elastic Keys)
- **Algorithm Support**: RSA (2048-4096), ECDSA/ECDH (P-256/384/521), EdDSA (Ed25519), AES (128/192/256), HMAC

### 3. **JWE/JWS Implementation**

- Full **JSON Web Encryption** and **JSON Web Signature** support
- Comprehensive algorithm combinations (75+ supported)
- Key wrapping vs. direct encryption modes
- Standards-compliant JWK (JSON Web Key) management

### 4. **Performance & Scalability**

**Key Generation Pools**:

- Pre-generated key pools for different algorithms
- Concurrent key generation with configurable pool sizes
- Optimized for high-throughput operations

**Database Support**:

- PostgreSQL for production
- SQLite for development/testing
- GORM ORM with migration support
- Transaction-based operations

### 5. **Security Features**

**Multi-Layered Network Security**:

- IP allowlisting (individual IPs and CIDR blocks)
- Per-IP rate limiting with configurable thresholds
- DDoS protection through request throttling
- Automatic blocking of excessive requests

**Browser Security Stack**:

- CORS configuration for cross-origin resource sharing
- CSRF protection with secure token handling
- Content Security Policy (CSP) for XSS prevention
- Comprehensive security headers (X-Frame-Options, HSTS, etc.)
- Sophisticated Swagger UI integration with automatic CSRF token injection

**Transport & Application Security**:

- TLS 1.3 with auto-generated certificates for development
- Certificate validation and management
- Secure cookie handling with HttpOnly and Secure flags
- Request/response validation middleware

**Operational Security**:

- Multiple unseal modes (simple keys, shared secrets, system fingerprinting)
- M-of-N secret sharing for high availability
- Encrypted storage of all sensitive material at rest
- Comprehensive audit logging with structured events
- Graceful degradation and secure failure modes
- Docker secrets integration for production deployments

#### Security Architecture Detail

**Multi-Layer Security Model**:

```
┌─────────────────────────────────────────────────────────────┐
│                    Network Security                         │
│  • IP Allowlisting (Individual IPs + CIDR blocks)          │
│  • Rate Limiting (Per-IP throttling)                       │
│  • DDoS Protection                                         │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                  Transport Security                         │
│  • TLS 1.3 with auto-generated certificates               │
│  • Certificate validation and management                   │
│  • Secure cipher suites                                   │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                 Application Security                        │
│  • CORS (Cross-Origin Resource Sharing)                   │
│  • CSRF (Cross-Site Request Forgery) Protection           │
│  • CSP (Content Security Policy)                          │
│  • XSS Protection Headers                                 │
│  • Security Headers (Helmet.js equivalent)                │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                Cryptographic Security                       │
│  • FIPS 140-3 Approved Algorithms                         │
│  • Hierarchical Key Management (Barrier System)           │
│  • Encrypted Key Storage                                  │
│  • Key Versioning and Rotation                            │
└─────────────────────────────────────────────────────────────┘
                               │
┌─────────────────────────────────────────────────────────────┐
│                 Operational Security                        │
│  • Comprehensive Audit Logging                            │
│  • Secure Failure Modes                                   │
│  • Graceful Degradation                                   │
│  • Secret Management                                      │
└─────────────────────────────────────────────────────────────┘
```

**Security Configuration Examples**:

```yaml
# Network Security
allowed_ips: ["127.0.0.1", "::1", "192.168.1.100"]
allowed_cidrs: ["10.0.0.0/8", "192.168.0.0/16"]
browser_ip_rate_limit: 100
service_ip_rate_limit: 25

# CORS Configuration (Browser API)
cors_allowed_origins: "https://app.example.com,https://admin.example.com"
cors_allowed_methods: "GET,POST,PUT,DELETE,OPTIONS"
cors_allowed_headers: "Content-Type,Authorization,X-CSRF-Token"

# CSRF Configuration (Browser API)
csrf_token_name: "csrf_token"
csrf_token_same_site: "Strict"  # None | Lax | Strict
csrf_token_cookie_secure: true
csrf_token_single_use_token: false

# TLS Configuration
bind_public_protocol: "https"
tls_public_dns_names: ["cryptoutil.example.com"]
tls_public_ip_addresses: ["192.168.1.100"]
```

**Security Headers Applied**:

```http
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: same-origin
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Permissions-Policy: camera=(), microphone=(), geolocation=(), payment=()
Content-Security-Policy: default-src 'none'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; ...
```

**Multi-Layered Network Security**:

- IP allowlisting (individual IPs and CIDR blocks)
- Per-IP rate limiting with configurable thresholds
- DDoS protection through request throttling
- Automatic blocking of excessive requests

**Browser Security Stack**:

- CORS configuration for cross-origin resource sharing
- CSRF protection with secure token handling
- Content Security Policy (CSP) for XSS prevention
- Comprehensive security headers (X-Frame-Options, HSTS, etc.)
- Sophisticated Swagger UI integration with automatic CSRF token injection

**Operational Security**:

- Multiple unseal modes (simple keys, shared secrets, system fingerprinting)
- M-of-N secret sharing for high availability
- Encrypted storage of all sensitive material at rest
- Comprehensive audit logging with structured events
- Graceful degradation and secure failure modes
- Docker secrets integration for production deployments

### 6. **Observability & Monitoring**

**OpenTelemetry Integration**:

- Distributed tracing with correlation across API contexts
- Metrics collection (request rates, latencies, error rates)
- Structured logging with slog
- OTLP export support for production monitoring
- Prometheus-compatible metrics

**Health Checks & Management**:

- Kubernetes-ready health endpoints (`/livez`, `/readyz`)
- Private management interface (port 9090)
- Graceful shutdown handling with proper connection draining
- Docker health checks for container orchestration

### 7. **API Design & Context Architecture**

**Dual-Context API Architecture**:

- **Browser Context** (`/browser/api/v1/*`): Full browser security (CORS, CSRF, CSP)
- **Service Context** (`/service/api/v1/*`): Streamlined for service-to-service
- **Management Interface**: Private health checks and administrative operations

**OpenAPI-First Development**:

- Comprehensive schemas for all operations (Elastic Keys, Material Keys, Crypto Operations)
- Auto-generated client/server code with oapi-codegen
- Built-in Swagger UI with sophisticated CSRF token handling
- Strict request/response validation middleware

### 8. **Development & Testing**

**Code Quality**:

- Comprehensive test coverage
- Test containers for integration testing
- Proper error handling and validation
- Structured configuration management

**Security Testing Strategy**:

- **Multi-Tool Approach**: Comprehensive security scanning with Staticcheck, govulncheck, Trivy, and Docker Scout
- **Local Development Integration**: Cross-platform security scan scripts (Windows PowerShell and Linux/macOS Bash)
- **CI/CD Security Pipeline**: Automated security scanning with SARIF reports and artifact generation
- **DAST Integration**: Dynamic Application Security Testing with OWASP ZAP and Nuclei
- **Targeted Scan Types**: Static analysis only, vulnerability scans only, and container security only modes
- **Risk-Based Scanning**: Execute security scans before commits for high-risk changes (crypto code, API endpoints, dependencies)
- **Compliance Reporting**: Generate security summary reports for review meetings and compliance documentation

### Manual Nuclei Vulnerability Scanning

[Nuclei](https://github.com/projectdiscovery/nuclei) is a fast, template-based vulnerability scanner that uses YAML templates to detect vulnerabilities, misconfigurations, and security issues.

#### Prerequisites: Start cryptoutil Services

Before running nuclei scans, ensure the cryptoutil services are running:

```sh
# Navigate to suite deployment directory
cd deployments/cryptoutil-suite

# Clean shutdown with volume removal
docker compose down -v

# Start all services
docker compose up -d

# Verify services are ready
curl -k https://localhost:8080/ui/swagger/doc.json  # SQLite instance (port 8080)
curl -k https://localhost:8081/ui/swagger/doc.json  # PostgreSQL instance 1 (port 8081)
curl -k https://localhost:8082/ui/swagger/doc.json  # PostgreSQL instance 2 (port 8082)
```

#### Manual Nuclei Scan Commands

**Service Configuration:**

- **cryptoutil-sqlite**: `https://localhost:8080/` (SQLite backend, development instance)
- **cryptoutil-postgres-1**: `https://localhost:8081/` (PostgreSQL backend, production-like instance)
- **cryptoutil-postgres-2**: `https://localhost:8082/` (PostgreSQL backend, production-like instance)

**Basic Security Scans:**

```sh
# Quick scan - Info and Low severity issues only (fast, ~5-10 seconds)
nuclei -target https://localhost:8080/ -severity info,low
nuclei -target https://localhost:8081/ -severity info,low
nuclei -target https://localhost:8082/ -severity info,low

# Standard scan - Medium, High, and Critical severity (comprehensive, ~10-30 seconds)
nuclei -target https://localhost:8080/ -severity medium,high,critical
nuclei -target https://localhost:8081/ -severity medium,high,critical
nuclei -target https://localhost:8082/ -severity medium,high,critical

# Full scan - All severity levels (thorough, ~20-60 seconds)
nuclei -target https://localhost:8080/ -severity info,low,medium,high,critical
nuclei -target https://localhost:8081/ -severity info,low,medium,high,critical
nuclei -target https://localhost:8082/ -severity info,low,medium,high,critical
```

**Targeted Vulnerability Scans:**

```sh
# CVE scanning (recent and historical vulnerabilities)
nuclei -target https://localhost:8080/ -tags cves -severity high,critical

# Security misconfigurations
nuclei -target https://localhost:8080/ -tags security-misconfiguration

# Information disclosure and exposure
nuclei -target https://localhost:8080/ -tags exposure,misc

# Technology detection and fingerprinting
nuclei -target https://localhost:8080/ -tags tech-detect

# Default credentials and weak authentication
nuclei -target https://localhost:8080/ -tags default-logins
```

**Performance-Optimized Scans:**

```sh
# High-performance scanning (adjust concurrency and rate limiting as needed)
nuclei -target https://localhost:8080/ -c 25 -rl 100 -severity high,critical

# Conservative scanning (lower resource usage)
nuclei -target https://localhost:8080/ -c 10 -rl 25 -severity medium,high,critical
```

**Batch Scanning Script (PowerShell):**

```powershell
# Scan all three cryptoutil instances
$targets = @(
    "https://localhost:8080/",  # SQLite instance
    "https://localhost:8081/",  # PostgreSQL instance 1
    "https://localhost:8082/"   # PostgreSQL instance 2
)

foreach ($target in $targets) {
    Write-Host "🔍 Scanning $target" -ForegroundColor Green
    nuclei -target $target -severity medium,high,critical
    Write-Host "✅ Completed scanning $target" -ForegroundColor Green
    Write-Host ""
}
```

**Batch Scanning Script (Bash):**

```bash
# Scan all three cryptoutil instances
targets=(
    "https://localhost:8080/"  # SQLite instance
    "https://localhost:8081/"  # PostgreSQL instance 1
    "https://localhost:8082/"  # PostgreSQL instance 2
)

for target in "${targets[@]}"; do
    echo "🔍 Scanning $target"
    nuclei -target "$target" -severity medium,high,critical
    echo "✅ Completed scanning $target"
    echo ""
done
```

#### Nuclei Configuration and Troubleshooting

**Template Management:**

```sh
# Update nuclei templates to latest version
nuclei -update-templates

# Check current template version
nuclei -templates-version

# List available templates (shows first 20)
nuclei -tl | head -20

# Search for specific template types
nuclei -tl | grep -i "http"
nuclei -tl | findstr http  # Windows PowerShell
```

**Common Issues and Solutions:**

**Templates not found:**

```sh
# Force template update
nuclei -update-templates

# Check template directory
dir C:\Users\%USERNAME%\nuclei-templates  # Windows
ls ~/nuclei-templates                     # Linux/macOS
```

**Services not responding:**

```sh
# Check service health
curl -k https://localhost:8080/ui/swagger/doc.json
curl -k https://localhost:9090/admin/v1/livez  # Admin health endpoint

# Check Docker containers
docker compose -f ./deployments/cryptoutil-suite/compose.yml ps

# View service logs
docker compose -f ./deployments/cryptoutil-suite/compose.yml logs sm-kms
```

**Clean restart:**

```sh
# Complete cleanup and restart
cd deployments/cryptoutil-suite
docker compose down -v
docker compose up -d

# Wait for services to be ready
sleep 30
curl -k https://localhost:8080/ui/swagger/doc.json
```

#### Interpreting Scan Results

**Expected Results:**

- **✅ "No results found"**: Indicates no vulnerabilities detected - good security posture
- **⚠️ Vulnerabilities found**: Review findings and address security issues
- **🔄 Scan performance**: Typically 5-60 seconds per service depending on scan profile

**Common False Positives to Ignore:**

- Some generic web server detections that don't apply to cryptoutil's security model
- Default credential checks (cryptoutil uses proper authentication)
- Generic misconfiguration checks that don't apply to the custom security implementation

**Security Validation:**

- Regular nuclei scanning helps validate the effectiveness of security controls
- Compare scan results across different instances (SQLite vs PostgreSQL)
- Use findings to improve security configurations and threat models

**Build & Deployment**:

- Multi-stage Docker builds
- Docker Compose for local development
- Production-ready container images
- Secret management through Docker secrets

## Recent Enhancements (September 2025)

### Advanced Security Architecture

- **Dual API Context Design**: Separate browser and service API paths with context-appropriate middleware
- **Enhanced CSRF Protection**: Sophisticated token handling with Swagger UI integration
- **Content Security Policy**: Comprehensive CSP implementation with development/production modes
- **IP Access Control**: Granular IP allowlisting with both individual IPs and CIDR block support
- **Rate Limiting**: Per-IP throttling with configurable thresholds and logging

### Production-Ready Features

- **Container Orchestration**: Complete Docker Compose setup with PostgreSQL and secret management
- **Health Monitoring**: Kubernetes-ready health endpoints with proper liveness/readiness probes
- **Configuration Management**: Hierarchical YAML configuration with CLI parameter overrides
- **Graceful Shutdown**: Signal-based shutdown with proper connection draining and resource cleanup
- **Observability**: Enhanced OpenTelemetry integration with structured logging and distributed tracing

## Architectural Strengths

1. **Enterprise Security**: The barrier system provides vault-like security with proper key hierarchy and unsealing mechanisms

2. **Standards Compliance**: Full adherence to cryptographic standards (JWE, JWS, JWK) and FIPS 140-3 requirements

3. **Scalability**: Key generation pools and efficient database design support high-throughput operations

4. **Operability**: Comprehensive observability, health checks, and graceful degradation

5. **Developer Experience**: OpenAPI-first design, comprehensive testing, and clear documentation

6. **Flexibility**: Multiple deployment modes, database backends, and unsealing strategies

## Potential Areas for Enhancement

1. **Distributed Deployment**: Currently designed as a single-node system; could benefit from clustering support

2. **Key Rotation**: While versioning is supported, automated key rotation policies could be enhanced

3. **Hardware Security Modules**: Could integrate with HSMs for even higher security assurance

4. **Performance Optimization**: Additional caching layers for frequently accessed keys

## Use Cases

This system is well-suited for:

- **Microservices Security**: Providing cryptographic services to distributed applications
- **Data Protection**: Encrypting sensitive data at rest and in transit
- **Digital Signatures**: Document signing and verification workflows
- **Compliance**: Meeting regulatory requirements for key management
- **Development**: Providing crypto-as-a-service for development teams

## Swagger UI Demo Guide

### Quick Start

1. **Start the KMS server with Docker Compose:**

   ```bash
   cd deployments/sm-kms
   docker compose --profile dev up -d
   ```

2. **Access Swagger UI:**

   Open `https://localhost:8080/ui/swagger/` in your browser.

3. **Authenticate with demo credentials:**
   - Username: `demo`
   - Password: `demo-password`

### Demo Flow: Key Pool → Key → Encrypt → Decrypt

#### Step 1: Create a Key Pool

1. In Swagger UI, expand **POST /api/browser/v1/pools**
2. Click "Try it out"
3. Enter request body:

   ```json
   {
     "algorithm": "A256GCM",
     "name": "demo-encryption-pool"
   }
   ```

4. Click "Execute"
5. Note the `pool_id` from the response.

#### Step 2: Create an Encryption Key

1. Expand **POST /api/browser/v1/keys**
2. Click "Try it out"
3. Enter request body (replace `{pool_id}` with actual ID):

   ```json
   {
     "pool_id": "{pool_id}"
   }
   ```

4. Click "Execute"
5. Note the `key_id` from the response.

#### Step 3: Encrypt Data

1. Expand **POST /api/browser/v1/keys/{keyId}/encrypt**
2. Click "Try it out"
3. Enter the `key_id` in the path parameter
4. Enter request body:

   ```json
   {
     "plaintext": "SGVsbG8gV29ybGQh"
   }
   ```

   Note: `SGVsbG8gV29ybGQh` is "Hello World!" base64-encoded.

5. Click "Execute"
6. Note the `ciphertext` from the response.

#### Step 4: Decrypt Data

1. Expand **POST /api/browser/v1/keys/{keyId}/decrypt**
2. Click "Try it out"
3. Enter the same `key_id` in the path parameter
4. Enter request body (replace with actual ciphertext):

   ```json
   {
     "ciphertext": "{ciphertext_from_step_3}"
   }
   ```

5. Click "Execute"
6. Verify the `plaintext` matches `SGVsbG8gV29ybGQh`.

### CSRF Token Handling

The Swagger UI automatically handles CSRF tokens:

1. Before executing any modifying request (POST, PUT, DELETE), the UI fetches a CSRF token.
2. The token is included in the `X-CSRF-Token` header automatically.
3. If you see 403 errors, refresh the page to get a new CSRF token.

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Connection refused | Ensure Docker containers are running: `docker compose ps` |
| 401 Unauthorized | Verify demo credentials: `demo` / `demo-password` |
| 403 Forbidden | CSRF token expired; refresh the page |
| Certificate error | Add `--insecure` flag in curl, or accept certificate in browser |
| TLS handshake error | Ensure TLS 1.3 is supported by your client |

## Docker Compose Architecture

The project includes a comprehensive multi-service Docker Compose setup for local development, testing, and observability:

### Service Architecture Diagram

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        Docker Compose Network (cryptoutil-network)                     │
│                                                                                        │
│  ┌─────────────────────────┐   ┌─────────────────────────┐   ┌───────────────────────┐ │
│  │    cryptoutil-sqlite    │   │   cryptoutil-postgres1  │   │  cryptoutil-postgres2 │ │
│  │ Port:      0.0.0.0:8080 │   │ Port:      0.0.0.0:8081 │   │ Port:    0.0.0.0:8082 │ │
│  │ Admin:   127.0.0.1:9090 │   │ Admin:   127.0.0.1:9090 │   │ Admin: 127.0.0.1:9090 │ │
│  │ Backend: SQLite         │   │ Backend: PostgreSQL     │   │ Backend: PostgreSQL   │ │
│  └─────────┬───────────────┘   └──────────┬────────────┬─┘   └─┬─────────┬───────────┘ │
│            │                              │            └───┬───┘         │             │
│            │                              │                │             │             │
│            └──────────────────────────────┴──────────────────────────────┘             │
│                                           │                │                           │
│                                           ▼                ▼                           │
│  ┌────────────────────┐      ┌────────────────────┐     ┌──────────────┐               |
│  │ OTEL Healthcheck   │─────>│  OTEL Collector    │     │  PostgreSQL  │               |
│  │ (Alpine Sidecar)   │      │  GRPC: 4317        │     │ Port: 5432   │               |
│  └────────────────────┘      │  HTTP: 4318        │     │ Database: DB │               |
│                              │  Metrics: 8070     │     │ User: USR    │               |
│                              │  Health: 13133     │     └──────────────┘               |
│                              │  pprof: 1777       │                                    |
│                              │  zPages: 55679     │                                    |
│                              └─────────┬──────────┘                                    |
│                                        │                                               |
│                                        ▼                                               │
│                              ┌────────────────────┐                                    │
│                              │  Grafana OTEL LGTM │                                    │
│                              │  UI: 3000          │                                    │
│                              │  OTLP GRPC: 14317  │                                    │
│                              │  OTLP HTTP: 14318  │                                    │
│                              └────────────────────┘                                    │
│                                                                                        │
└────────────────────────────────────────────────────────────────────────────────────────┘

Dependencies Flow:
1. postgres → cryptoutil-postgres-1 → cryptoutil-postgres-2
2. opentelemetry-collector-contrib → healthcheck-opentelemetry-collector-contrib
3. grafana-otel-lgtm → opentelemetry-collector-contrib
4. cryptoutil-sqlite (independent of postgres)

Telemetry Flow:
cryptoutil services → OTEL Collector (4317/4318) → Grafana LGTM (14317/14318)
OTEL Collector self-metrics → Prometheus scraping (8070)

**Health Checks:**
- cryptoutil services: wget https://127.0.0.1:9090/admin/v1/livez
- postgres: pg_isready -U USR -d DB
- grafana: curl http://127.0.0.1:3000/api/health
- otel-collector: External via healthcheck sidecar (ping + wget http://otel:13133/)
```

### Port Mapping Reference

| Service | Public Port(s) | Admin Port | Protocol | Purpose |
|---------|---------------|------------|----------|---------|
| cryptoutil-sqlite | 8080 | 9090 | HTTPS | SQLite backend instance |
| cryptoutil-postgres-1 | 8081 | 9090 | HTTPS | PostgreSQL backend instance #1 |
| cryptoutil-postgres-2 | 8082 | 9090 | HTTPS | PostgreSQL backend instance #2 |
| postgres | 5432 | - | TCP | PostgreSQL database |
| opentelemetry-collector | 4317 (GRPC), 4318 (HTTP) | 8070 (metrics), 13133 (health) | OTLP | Telemetry collection |
| grafana-otel-lgtm | 3000 | 14317 (GRPC), 14318 (HTTP) | HTTP | Observability stack |

### Resource Allocation

| Service | Memory Limit | Memory Reserved | CPU Limit | CPU Reserved |
|---------|-------------|----------------|-----------|--------------|
| cryptoutil_* | 256M | 128M | - | - |
| postgres | 512M | 256M | - | - |
| opentelemetry-collector | 256M | 128M | 0.25 | 0.1 |
| grafana-otel-lgtm | 512M | 256M | 0.5 | 0.25 |

### Security Architecture

**Docker Secrets (Best Practice Implementation):**

- Database URLs: `cryptoutil_database_url.secret`
- Unseal Keys: 5-of-5 Shamir secret shares
- No environment variables for sensitive data
- Secrets mounted to `/run/secrets/` in containers

**Network Isolation:**

- All services communicate via `cryptoutil-network` bridge
- No direct host network exposure except mapped ports
- Service-to-service DNS resolution

**Volume Management:**

- `postgres_data`: Persistent PostgreSQL storage
- `grafana_data`: Persistent Grafana configuration and dashboards
- Named volumes for data persistence across restarts

## Testing Documentation

### End-to-End (E2E) Testing

Comprehensive E2E testing validates complete workflows across all services (KMS, CA, JOSE, Identity). Tests use real service deployments with Docker Compose, real databases (PostgreSQL, SQLite), and real telemetry infrastructure.

**E2E Test Suites**:

- **OAuth Workflows** (`oauth_workflow_test.go`) - OAuth 2.1 authorization code + PKCE, client credentials
- **KMS Workflows** (`kms_workflow_test.go`) - Encrypt/decrypt, sign/verify, key rotation
- **CA Workflows** (`ca_workflow_test.go`) - Certificate lifecycle, OCSP, CRL distribution
- **JOSE Workflows** (`jose_workflow_test.go`) - JWT signing/verification, JWKS endpoint, JWK rotation

**Running E2E Tests**:

```powershell
# Deploy all services (SUITE-level)
docker compose -f ./deployments/cryptoutil-suite/compose.yml up -d

# Or deploy individual services (SERVICE-level)
docker compose -f ./deployments/pki-ca/compose.yml up -d
docker compose -f ./deployments/jose-ja/compose.yml up -d

# Run E2E tests
go test -tags=e2e -v -timeout=30m ./internal/test/e2e/

# Cleanup
docker compose -f ./deployments/cryptoutil-suite/compose.yml down -v
```

**Detailed Documentation**: See [E2E-TESTING.md](./E2E-TESTING.md) for:

- Test architecture and infrastructure
- Complete test descriptions with step-by-step workflows
- Local execution instructions
- CI/CD integration
- Troubleshooting guide
- Development guidelines

## Conclusion

Cryptoutil represents a mature, well-architected cryptographic service that successfully balances security, performance, and usability. The codebase demonstrates strong software engineering practices, comprehensive security measures, and production readiness. It's particularly notable for its adherence to cryptographic standards and its sophisticated key management hierarchy.

The Docker Compose architecture provides a complete local development and testing environment with proper observability, multi-instance testing capabilities, and production-like configuration management.
