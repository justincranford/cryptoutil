# CryptoUtil Project - Deep Technical Analysis

**Generated:** 2025-10-26  
**Purpose:** Comprehensive project analysis for AI chat session context and developer onboarding

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Technology Stack](#technology-stack)
4. [Project Structure](#project-structure)
5. [Artifact Generation Analysis](#artifact-generation-analysis)
6. [Workflow System](#workflow-system)
7. [Testing Strategy](#testing-strategy)
8. [Security Implementation](#security-implementation)
9. [Observability & Monitoring](#observability--monitoring)
10. [Build & Deployment](#build--deployment)
11. [Development Tools](#development-tools)
12. [Dependencies](#dependencies)

---

## Executive Summary

**CryptoUtil** is an enterprise-grade, production-ready Key Management System (KMS) and cryptographic service implemented in Go 1.25.1. The project demonstrates mature software engineering practices with comprehensive testing, security scanning, observability, and automated quality gates.

### Key Characteristics

- **Domain:** Cryptographic Key Management & Operations
- **Language:** Go 1.25.1 (strict version pinning across all configs)
- **Architecture:** Layered hexagonal architecture with clear separation of concerns
- **Database:** Dual-backend support (PostgreSQL production, SQLite development)
- **API:** OpenAPI 3.0.3 driven with auto-generated handlers
- **Security:** NIST FIPS 140-3 compliant algorithms only
- **Deployment:** Docker Compose with Kubernetes-ready health endpoints
- **Testing:** 6-layer pyramid (unit, integration, E2E, fuzz, mutation, load)
- **CI/CD:** 5 independent workflows (quality, E2E, DAST, SAST, robust)

---

## Architecture Overview

### Layered Architecture

```
┌─────────────────────────────────────────────────────┐
│  API Layer (OpenAPI Generated)                      │
│  - Browser Context (/browser/api/v1/*)             │
│  - Service Context (/service/api/v1/*)             │
│  - Swagger UI (/ui/swagger/*)                      │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│  Handler Layer (HTTP Request Handling)              │
│  - Request validation                               │
│  - CORS/CSRF middleware (browser context)          │
│  - Rate limiting (IP-based)                        │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│  Business Logic Layer                               │
│  - Elastic Key Management                          │
│  - Material Key Management                         │
│  - Crypto Operations (encrypt/decrypt/sign/verify) │
│  - Barrier System (unseal → root → intermediate)  │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│  Repository Layer (GORM ORM)                        │
│  - PostgreSQL (production)                         │
│  - SQLite (development/testing)                    │
│  - Migration management                            │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│  Common Utilities                                    │
│  - Crypto primitives (RSA, ECDSA, EdDSA, AES)     │
│  - Key generation pools (concurrent)               │
│  - UUIDv7 generation                               │
│  - Datetime utilities                              │
└─────────────────────────────────────────────────────┘
```

### Dual API Context Design

The application exposes identical operations through two contexts optimized for different clients:

| Context Path | Target Clients | Security Features |
|-------------|----------------|-------------------|
| `/browser/api/v1/*` | Web browsers | CORS, CSRF tokens, CSP headers, comprehensive security headers |
| `/service/api/v1/*` | Service-to-service | Core security only, optimized for M2M communication |

Both contexts share the same business logic but apply different middleware stacks.

### Port Architecture

- **8080 (HTTPS):** Public API (browser + service contexts, Swagger UI)
- **9090 (HTTP):** Private admin API (health checks `/livez`, `/readyz`, graceful shutdown `/shutdown`)
- **5432:** PostgreSQL database
- **4317:** OpenTelemetry Collector gRPC endpoint
- **4318:** OpenTelemetry Collector HTTP endpoint
- **3000:** Grafana UI

---

## Technology Stack

### Core Technologies

- **Language:** Go 1.25.1 (module: `cryptoutil`)
- **Web Framework:** Fiber v2.52.9 (high-performance HTTP framework)
- **Database ORM:** GORM v1.31.0 with drivers for PostgreSQL and SQLite
- **OpenAPI:** oapi-codegen v2.5.0 for code generation from OpenAPI 3.0.3 specs
- **Cryptography:**
  - `golang.org/x/crypto` (standard library extensions)
  - `github.com/cloudflare/circl` (post-quantum and modern crypto)
  - `github.com/lestrrat-go/jwx/v3` (JWE/JWS support)

### Observability Stack

- **Tracing:** OpenTelemetry SDK v1.38.0
- **Metrics:** OpenTelemetry SDK Metrics v1.38.0
- **Logging:** OpenTelemetry SDK Log v0.14.0 + slog
- **Exporters:** OTLP gRPC/HTTP dual support
- **Collectors:**
  - `opentelemetry-collector-contrib` (sidecar for processing)
  - `grafana/otel-lgtm` (Grafana, Loki, Tempo, Prometheus all-in-one)

### Testing Tools

- **Unit/Integration:** `testify` v1.11.1 (assertions and test suites)
- **E2E:** `testcontainers-go` v0.39.0 (Docker-based integration testing)
- **Fuzz Testing:** Go native fuzz testing (Go 1.18+)
- **Mutation Testing:** Gremlins v1.0.0+ (test quality validation)
- **Load Testing:** Gatling 3.13.3 (Java-based HTTP load testing)
- **Security Scanning:**
  - OWASP ZAP (DAST)
  - Nuclei (vulnerability scanning)
  - gosec (Go security analyzer)
  - golangci-lint (comprehensive linting with security checks)
  - Trivy (container vulnerability scanning)
  - Docker Scout (advanced container security)

### Development Tools

- **Code Generation:** `oapi-codegen` (OpenAPI to Go code)
- **Formatting:** `gofumpt` (strict Go formatter, superset of gofmt)
- **Import Management:** `goimports` (automatic import organization)
- **Linting:** `golangci-lint` v1.63.4+ (40+ linters aggregated)
- **Pre-commit:** Python-based pre-commit framework with Git hooks
- **Local Workflow Testing:** `act` (GitHub Actions local runner)

---

## Project Structure

### Standard Go Project Layout

```
cryptoutil/
├── cmd/                          # Main applications
│   └── cryptoutil/
│       └── main.go               # Application entry point
├── internal/                     # Private application code
│   ├── client/                   # Client utilities
│   ├── cmd/                      # Command implementations
│   ├── common/                   # Shared utilities
│   │   ├── apperr/               # Application error types
│   │   ├── config/               # Configuration management
│   │   ├── container/            # Service container/DI
│   │   ├── crypto/               # Cryptographic primitives
│   │   │   ├── digests/          # Hash functions
│   │   │   ├── keygen/           # Key generation
│   │   │   ├── random/           # CSPRNG
│   │   │   └── ...
│   │   ├── magic/                # Magic numbers/constants
│   │   ├── pool/                 # Object pooling
│   │   ├── telemetry/            # OpenTelemetry setup
│   │   └── util/                 # Utility functions
│   ├── e2e/                      # End-to-end tests
│   └── server/                   # Server implementation
│       ├── application/          # Application services
│       ├── barrier/              # Barrier/unsealing system
│       ├── businesslogic/        # Business logic layer
│       ├── handler/              # HTTP handlers
│       └── repository/           # Data access layer
├── api/                          # OpenAPI specifications
│   ├── openapi_spec_components.yaml
│   ├── openapi_spec_paths.yaml
│   ├── client/                   # Generated client code
│   ├── model/                    # Generated model code
│   └── server/                   # Generated server code
├── configs/                      # Configuration files
│   └── test/                     # Test configurations
├── deployments/                  # Deployment configurations
│   ├── Dockerfile                # Multi-stage Docker build
│   └── compose/                  # Docker Compose configs
│       ├── compose.yml           # Main compose file
│       ├── cryptoutil/           # Service configs + secrets
│       ├── grafana-otel-lgtm/    # Observability stack
│       ├── otel/                 # OTEL collector config
│       └── postgres/             # PostgreSQL configs
├── docs/                         # Documentation
│   ├── README.md                 # Extended documentation
│   ├── DEEP-ANALYSIS.md          # This file
│   ├── pre-commit-hooks.md       # Hook documentation
│   └── todos-*.md                # Task tracking
├── scripts/                      # Build and utility scripts
│   ├── cicd/                     # CI/CD validation library
│   └── github-workflows/         # Workflow utilities
│       └── workflow/
│           └── workflow.go
├── test/                         # Test resources
│   ├── e2e/                      # E2E test artifacts
│   └── load/                     # Gatling load tests
├── .github/                      # GitHub configurations
│   ├── workflows/                # CI/CD workflows
│   │   ├── ci-quality.yml        # Code quality + build
│   │   ├── ci-e2e.yml            # End-to-end testing
│   │   ├── ci-dast.yml           # Dynamic security testing
│   │   ├── ci-sast.yml           # Static security testing
│   │   └── ci-robust.yml         # Robustness testing
│   ├── instructions/             # Copilot instruction files
│   └── copilot-instructions.md   # Main Copilot guidance
└── [Temporary Artifact Directories]
    ├── dast-reports/             # DAST scan outputs
    ├── workflow-reports/         # Workflow execution logs
    ├── e2e-artifacts/            # E2E test outputs
    ├── test-results/             # Generic test results
    └── coverage.out/.html        # Coverage reports
```

---

## Artifact Generation Analysis

### Current Artifact Locations (Scattered)

This section maps ALL temporary files/directories created by various project components:

#### 1. Main Application Runtime Artifacts

**Location:** Various (no consistent pattern)

- `cryptoutil` (binary in project root) - Built by `go build`
- `nohup.out` (background process logs) - Created by nohup in DAST workflows
- `.env` (ignored, but potential artifact)

#### 2. Go Test Artifacts

**Location:** Project root + per-package

- `coverage.out` - Coverage profile (project root)
- `coverage.html` - HTML coverage report (project root)
- `*.test` - Compiled test binaries (per-package, e.g., `crypto.test`)

#### 3. DAST Security Scan Artifacts

**Location:** `dast-reports/` (partially consolidated)

**Files Created:**
- `nuclei.log` - Nuclei vulnerability scan log
- `nuclei.sarif` - Nuclei results in SARIF format
- `nuclei-templates.version` - Nuclei templates version info
- `zap-report.html` - ZAP full scan HTML report
- `zap-report.md` - ZAP full scan Markdown report
- `zap-report.xml` - ZAP full scan XML report
- `zap-report.json` - ZAP full scan JSON report
- `zap-api-report.html` - ZAP API scan HTML report
- `zap-api-report.json` - ZAP API scan JSON report
- `response-headers.txt` - HTTP security headers baseline
- `cryptoutil.stdout` - Application stdout during DAST
- `cryptoutil.stderr` - Application stderr during DAST
- `system-info.txt` - System diagnostics
- `artifact-listing.txt` - File manifest
- `act-status.txt` - DAST execution summary
- `container-logs/` - Docker container logs
  - `*.log` - Individual container logs
  - `containers.txt` - Container listing
  - `docker-info.txt` - Docker system info
  - `docker-df.txt` - Docker disk usage

#### 4. GitHub Workflow Execution Artifacts

**Location:** `workflow-reports/` (consolidated by cmd/workflow)

**Files Created:**
- `{workflow-name}-{timestamp}.log` - Individual workflow logs
- `{workflow-name}-analysis-{timestamp}.md` - Workflow analysis reports
- `combined-{timestamp}.log` - Combined execution log

#### 5. E2E Test Artifacts

**Location:** Multiple locations (inconsistent)

- `e2e-artifacts/{run-number}/e2e-service-logs/` - Service logs from E2E runs
- `e2e-service-logs.txt` - Combined service logs (root level during workflow)
- `e2e-container-logs/` - Individual container logs (root level during workflow)
  - `{container-name}.log` - Per-container logs
  - `collection-summary.txt` - Log collection summary
  - `errors.txt` - Collection errors
- `test/e2e/e2e-reports/` - E2E test reports (in test directory)
- `internal/cmd/e2e/e2e-reports/` - E2E reports (in internal directory)

#### 6. Load Test Artifacts (Gatling)

**Location:** `test/load/target/` (Java Maven convention)

- `target/gatling/` - Gatling simulation results
  - HTML reports with timestamps
  - `simulation.log` files
- `target/classes/` - Compiled Java classes

#### 7. Mutation Testing Artifacts

**Location:** Project root (no consistent pattern)

- `mutation-{package-name}.json` - Gremlins mutation test results

#### 8. Pre-commit Hook Cache

**Location:** User home directory (Windows-specific)

- `C:\Users\{username}\.cache\pre-commit\` - Pre-commit tool cache

#### 9. Static Analysis Artifacts

**Location:** Project root (various)

- `trivy-image.sarif` - Trivy container scan results
- `docker-scout-cves.sarif` - Docker Scout CVE results
- `sbom.spdx.json` - Software Bill of Materials

#### 10. Build Artifacts (Docker)

**Location:** Docker layer cache (not in repo)

- Docker images with tags from `metadata-action`
- Build cache managed by Docker Buildx

#### 11. Generated Code (NOT temporary, committed)

**Location:** `api/` directory

- `api/client/openapi_gen_client.go`
- `api/model/openapi_gen_model.go`
- `api/server/openapi_gen_server.go`

These are **committed** to version control and regenerated via `go generate`.

---

### Current .gitignore Coverage

```gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib
cryptoutil$

# Test artifacts
*.test
coverage.out
coverage.html

# DAST reports
dast-reports/

# Workflow reports
workflow-reports/

# E2E reports
test/e2e/e2e-reports/

# Other
.env
.idea/
output/
*.pem
*.der
```

**Gaps in Coverage:**
- `e2e-artifacts/` (root level)
- `e2e-service-logs.txt` (root level)
- `e2e-container-logs/` (root level)
- `internal/cmd/e2e/e2e-reports/` (internal directory)
- `test-results/` (root level)
- `mutation-*.json` files (root level)
- `nohup.out` (root level)
- `trivy-*.sarif` (root level)
- `docker-scout-*.sarif` (root level)
- `sbom.*.json` (root level)
- `test/load/target/` (Gatling build artifacts)

---

## Workflow System

### 5 Independent CI/CD Workflows

#### 1. ci-quality.yml - Code Quality & Build

**Purpose:** Code formatting, linting, building, container security scanning, SBOM generation, mutation testing

**Jobs:**
- `code-quality` - gofumpt formatting check, golangci-lint, GitHub Actions version validation
- `build` - Binary build, Docker image build/push, Trivy scan, Docker Scout analysis
- `sbom` - Generate SPDX SBOM, upload to dependency graph
- `mutation-testing` - Run Gremlins on high-coverage packages (main branch only)

**Artifacts Generated:**
- Docker images pushed to GHCR
- `trivy-image.sarif`
- `docker-scout-cves.sarif`
- `sbom.spdx.json`
- `mutation-*.json` files

**Triggers:** Push to main/develop, PRs (excluding docs/markdown)

#### 2. ci-e2e.yml - End-to-End Testing

**Purpose:** Full system integration testing with Docker Compose

**Jobs:**
- `e2e` - Start PostgreSQL, build images, start services, run E2E tests, collect logs

**Artifacts Generated:**
- `e2e-service-logs.txt` (combined logs)
- `e2e-container-logs/` directory with individual container logs
- `e2e-container-logs/collection-summary.txt`

**Services Tested:**
- 3x cryptoutil instances (ports 8080, 8081, 8082)
- PostgreSQL database
- OpenTelemetry Collector
- Grafana-OTEL-LGTM stack

**Triggers:** Push to main/develop, PRs (excluding docs/markdown)

#### 3. ci-dast.yml - Dynamic Application Security Testing

**Purpose:** Runtime security scanning of running application

**Jobs:**
- `dast-security-scan` - Start app, run OWASP ZAP (full + API scans), run Nuclei vulnerability scan

**Scan Profiles:**
- `quick` - 3-5 minutes (recent CVEs, basic misconfigurations)
- `full` - 10-15 minutes (comprehensive coverage) **[Default]**
- `deep` - 20-25 minutes (all templates)

**Artifacts Generated:**
- `dast-reports/nuclei.log`
- `dast-reports/nuclei.sarif`
- `dast-reports/zap-report.{html,json,xml,md}`
- `dast-reports/zap-api-report.{html,json}`
- `dast-reports/response-headers.txt`
- `dast-reports/cryptoutil.{stdout,stderr}`
- `dast-reports/system-info.txt`
- `dast-reports/container-logs/`

**Triggers:** Push to main, PRs to main, weekly schedule (Sundays 2 AM UTC), manual dispatch

#### 4. ci-sast.yml - Static Application Security Testing

**Purpose:** Source code security analysis without execution

**Jobs:**
- `sast` - Run gosec Go security scanner, golangci-lint with security-focused linters

**Artifacts Generated:**
- SARIF reports uploaded to GitHub Security tab

**Triggers:** Push to main/develop, PRs (excluding docs/markdown)

#### 5. ci-robust.yml - Robustness Testing

**Purpose:** Concurrency, race conditions, fuzz testing, benchmarking

**Jobs:**
- `race-detection` - Run tests with `-race` flag
- `fuzz-testing` - Run Go fuzz tests with 1-minute duration
- `benchmarks` - Run performance benchmarks

**Artifacts Generated:**
- Benchmark results (inline in workflow logs)

**Triggers:** Push to main/develop, PRs (excluding docs/markdown)

---

### Workflow Testing Tool: cmd/workflow

**Location:** `cmd/workflow/main.go` (calls `internal/cmd/workflow/workflow.go`)

**Purpose:** Execute GitHub Actions workflows locally using `act` with comprehensive monitoring and reporting

**Features:**
- Sequential workflow execution with detailed logging
- Real-time output streaming to console and files
- Automatic log analysis and task result extraction
- Workflow-specific result parsing (DAST, E2E, SAST, Robust, Quality)
- Executive summary reports in Markdown format
- Execution metrics (duration, memory usage, CPU time approximation)
- Combined logs for entire execution session

**Supported Workflows:**
- `e2e` - End-to-End Testing
- `dast` - Dynamic Application Security Testing
- `sast` - Static Application Security Testing
- `robust` - Robustness Testing
- `quality` - Code Quality

**Usage:**
```bash
go run ./cmd/workflow -workflows=quality,e2e
go run ./cmd/workflow -workflows=dast -dry-run
go run ./cmd/workflow -list
```

**Output Files:**
- `workflow-reports/{workflow}-{timestamp}.log` - Full workflow output
- `workflow-reports/{workflow}-analysis-{timestamp}.md` - Analysis report
- `workflow-reports/combined-{timestamp}.log` - All workflows combined

---

## Testing Strategy

### 6-Layer Testing Pyramid

#### Layer 1: Unit Tests
- **Tool:** Go standard testing + testify
- **Coverage:** 70%+ target
- **Location:** `*_test.go` files co-located with source
- **Pattern:** Table-driven tests with comprehensive edge cases
- **Execution:** `go test ./...`

#### Layer 2: Integration Tests
- **Tool:** Go testing + testcontainers
- **Coverage:** Database interactions, external service integrations
- **Location:** `*_integration_test.go` files
- **Pattern:** Real database connections using test containers
- **Execution:** `go test -tags=integration ./...`

#### Layer 3: End-to-End Tests
- **Tool:** Go testing + Docker Compose + testify suites
- **Coverage:** Full system with 3 instances + dependencies
- **Location:** `internal/cmd/e2e/` directory
- **Pattern:** HTTP API calls with full request/response validation
- **Execution:** `go test -tags=e2e -v ./internal/cmd/e2e/`
- **Duration:** ~10-15 minutes with full stack startup

#### Layer 4: Fuzz Testing
- **Tool:** Go native fuzz testing (Go 1.18+)
- **Coverage:** Cryptographic functions, parsers, input validators
- **Location:** `*_fuzz_test.go` files
- **Pattern:** Property-based testing with random inputs
- **Execution:** `go test -fuzz=^FuzzXXX$ -fuzztime=5s ./path`

#### Layer 5: Mutation Testing
- **Tool:** Gremlins mutation testing
- **Coverage:** High-coverage packages to validate test quality
- **Location:** Results saved to `mutation-*.json`
- **Pattern:** Introduce bugs, verify tests catch them
- **Execution:** `gremlins unleash ./internal/common/util/datetime/`
- **Threshold:** 70% efficacy, 60% mutation coverage

#### Layer 6: Load Testing
- **Tool:** Gatling (Java-based)
- **Coverage:** HTTP API performance under load
- **Location:** `test/load/` directory
- **Pattern:** Scala DSL for load scenarios
- **Execution:** `./mvnw gatling:test` (from test/load/)

---

### Security Testing (DAST + SAST)

#### Dynamic Analysis (DAST)
- **OWASP ZAP Full Scan:** Comprehensive web app security testing
- **OWASP ZAP API Scan:** OpenAPI-driven API security testing
- **Nuclei:** Template-based vulnerability scanning (CVEs, misconfigurations)
- **Manual Header Analysis:** Security header validation

#### Static Analysis (SAST)
- **gosec:** Go-specific security analyzer (SQL injection, crypto misuse, etc.)
- **golangci-lint:** Security-focused linters (G101-G602)
- **Trivy:** Container vulnerability scanning
- **Docker Scout:** Advanced container security analysis

---

## Security Implementation

### Cryptographic Standards

**NIST FIPS 140-3 Compliance:**
- RSA ≥ 2048 bits
- AES ≥ 128 bits (typically 256)
- ECDSA with NIST curves (P-256, P-384, P-521)
- EdDSA (Ed25519, Ed448)
- SHA-256, SHA-384, SHA-512
- HMAC-SHA256/384/512

**Hierarchical Key Architecture:**
```
Unseal Secrets (Shamir's Secret Sharing or simple)
    ↓
Root Key (encrypted at rest)
    ↓
Intermediate Keys (encrypted by root)
    ↓
Content Keys (encrypted by intermediate)
```

### Security Features

#### API Security
- **IP Allowlisting:** Individual IPs + CIDR blocks
- **Rate Limiting:** Per-IP rate limits (100 req/sec browser, 25 req/sec service)
- **CORS:** Configurable origins for browser APIs
- **CSRF Protection:** Secure token handling with `_csrf` cookie
- **CSP Headers:** Content Security Policy for XSS prevention
- **Security Headers:** Comprehensive Helmet.js-equivalent headers

#### Secret Management
- **Docker Secrets:** Mounted to `/run/secrets/` (preferred)
- **Kubernetes Secrets:** File-based or direct reference (preferred)
- **Environment Variables:** Explicitly avoided for secrets in production

#### TLS Configuration
- **Minimum Version:** TLS 1.2+
- **Certificate Validation:** Full chain validation, no `InsecureSkipVerify`
- **Self-signed Certs:** Only for local development

---

## Observability & Monitoring

### OpenTelemetry Architecture

**Push-Based Telemetry Flow:**
```
cryptoutil instances (OTLP gRPC:4317 or HTTP:4318)
    ↓
OpenTelemetry Collector Contrib (sidecar)
    ↓
Grafana-OTEL-LGTM (OTLP HTTP:14318)
```

**Pull-Based Infrastructure Monitoring:**
```
Grafana-OTEL-LGTM (Prometheus) → OpenTelemetry Collector Contrib (HTTP:8888/metrics)
```

### Telemetry Types

1. **Traces** - Distributed tracing of HTTP requests and crypto operations
2. **Metrics** - Performance metrics, request counts, error rates
3. **Logs** - Structured logging with slog + OpenTelemetry integration

### Health Endpoints

- **`/livez`** - Liveness probe (process alive)
- **`/readyz`** - Readiness probe (ready to serve traffic)
- **`/shutdown`** - Graceful shutdown endpoint (POST)

---

## Build & Deployment

### Multi-Stage Docker Build

**File:** `deployments/Dockerfile`

**Stages:**
1. **Builder Stage:** Go build with static linking
2. **Runtime Stage:** Minimal Alpine image with CA certificates

**Build Args:**
- `APP_VERSION` - Application version/commit SHA
- `VCS_REF` - Git commit reference
- `BUILD_DATE` - ISO 8601 build timestamp

### Docker Compose Stack

**File:** `deployments/compose/compose.yml`

**Services:**
- `cryptoutil_sqlite` - SQLite instance (port 8080)
- `cryptoutil_postgres_1` - PostgreSQL instance 1 (port 8081)
- `cryptoutil_postgres_2` - PostgreSQL instance 2 (port 8082)
- `postgres` - PostgreSQL database (port 5432)
- `opentelemetry-collector` - OTEL collector sidecar (ports 4317, 4318, 8888, etc.)
- `grafana-otel-lgtm` - Observability stack (port 3000)

**Secrets Management:**
- Unseal secrets mounted from `deployments/compose/cryptoutil/` directory
- PostgreSQL credentials stored as Docker secrets

---

## Development Tools

### Pre-commit Hooks

**Framework:** Python pre-commit (v3.5.0+)

**Hooks:**
1. **File Fixers:** Trailing whitespace, EOF newline, YAML formatting
2. **Syntax Validators:** YAML, JSON, GitHub Actions, Dockerfiles
3. **Go Tools:**
   - `gofumpt` - Strict formatting
   - `goimports` - Import organization
   - `errcheck` - Error handling validation
   - `go build` - Compilation check
4. **Linting:** `golangci-lint run` with auto-fix for supported linters

**Installation:**
```bash
pip install pre-commit
pre-commit install
```

Or use automated setup scripts:
```bash
# Windows
.\scripts\setup-pre-commit.ps1
# or
.\scripts\setup-pre-commit.bat
```

### Code Quality Tools

- **gofumpt:** Stricter Go formatter (superset of gofmt)
- **goimports:** Automatic import management
- **golangci-lint:** Aggregates 40+ linters
  - Auto-fixable: wsl, gofmt, goimports, godot, goconst, importas, revive
  - Manual-only: errcheck, gosimple, govet, staticcheck, gosec, etc.

---

## Dependencies

### Major Dependencies (go.mod)

**Web Framework:**
- `github.com/gofiber/fiber/v2` v2.52.9
- `github.com/gofiber/contrib/otelfiber` v1.0.10 (OTEL integration)
- `github.com/gofiber/swagger` v1.1.1

**Database:**
- `gorm.io/gorm` v1.31.0
- `gorm.io/driver/postgres` v1.6.0
- `gorm.io/driver/sqlite` v1.6.0
- `modernc.org/sqlite` v1.39.1 (pure Go SQLite)

**Cryptography:**
- `golang.org/x/crypto` v0.43.0
- `github.com/cloudflare/circl` v1.6.1
- `github.com/lestrrat-go/jwx/v3` v3.0.11 (JWE/JWS)

**OpenTelemetry:**
- `go.opentelemetry.io/otel` v1.38.0
- `go.opentelemetry.io/otel/sdk` v1.38.0
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` v1.38.0
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` v1.38.0
- `go.opentelemetry.io/contrib/bridges/otelslog` v0.13.0

**Configuration:**
- `github.com/spf13/viper` v1.21.0
- `github.com/spf13/pflag` v1.0.10
- `github.com/goccy/go-yaml` v1.18.0

**Testing:**
- `github.com/stretchr/testify` v1.11.1
- `github.com/testcontainers/testcontainers-go` v0.39.0

**OpenAPI:**
- `github.com/getkin/kin-openapi` v0.133.0
- `github.com/oapi-codegen/runtime` v1.1.2
- `github.com/oapi-codegen/fiber-middleware` v1.0.2

**Tool Dependencies:**
- `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` (code generation)

---

## Appendix: Quick Reference

### Common Commands

```bash
# Build
go build -o cryptoutil ./cmd/cryptoutil

# Test
go test ./... -cover
go test -tags=e2e -v ./internal/cmd/e2e/
go test -fuzz=^FuzzXXX$ -fuzztime=5s ./path

# Lint
golangci-lint run --fix

# Format
gofumpt -extra -w .

# Security Scan
# (Now handled by cmd/workflow)

# DAST Scan
# (Now handled by cmd/workflow)

# Mutation Testing
# (Now handled by cmd/workflow)

# Workflow Testing
# (Now handled by cmd/workflow)
act workflow_dispatch -W .github/workflows/ci-sast.yml

# (Now handled by cmd/workflow)
act push -W .github/workflows/ci-robust.yml

# (Now handled by cmd/workflow)
act push -W .github/workflows/ci-quality.yml

go run ./cmd/workflow -workflows=quality,e2e

# Docker Compose
cd deployments/compose
docker compose up -d
docker compose logs -f cryptoutil_postgres_1
docker compose down -v
```

### Key Files

- `go.mod` - Go module definition (Go 1.25.1)
- `.golangci.yml` - Linter configuration (40+ linters)
- `.pre-commit-config.yaml` - Pre-commit hook configuration
- `.gofumpt.toml` - Formatter configuration
- `.gremlins.yaml` - Mutation testing configuration
- `.nuclei-ignore` - Nuclei false positive suppressions
- `.zap/rules.tsv` - ZAP scanning rules
- `deployments/compose/compose.yml` - Docker Compose configuration
- `api/openapi_spec_{components,paths}.yaml` - OpenAPI specifications

---

**Document Version:** 1.0  
**Last Updated:** 2025-10-26  
**Maintainer:** CryptoUtil Development Team
