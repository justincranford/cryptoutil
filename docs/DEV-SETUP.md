# Development Environment Setup

This guide covers setting up a complete development environment for the cryptoutil project on Windows, Linux, and macOS systems.

## Table of Contents

- [Prerequisites Overview](#prerequisites-overview)
- [Why Each Tool is Required](#why-each-tool-is-required)
- [Platform-Specific Setup](#platform-specific-setup)
  - [Windows](#windows)
  - [Linux](#linux)
  - [macOS](#macos)
- [Common Setup Steps](#common-setup-steps)
- [IDE Configuration](#ide-configuration)
- [Project Setup](#project-setup)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites Overview

### Minimum Version Requirements

| Tool | Version | Purpose | Required For |
|------|---------|---------|--------------|
| **Go** | 1.25.5+ | Primary language | All development |
| **Docker** | 24+ | Containerization | PostgreSQL, E2E tests, deployments |
| **Docker Compose** | v2+ | Container orchestration | Multi-service deployments |
| **Python** | 3.14+ | Pre-commit hooks, utilities | Code quality automation |
| **Node.js** | 24.11.1+ LTS | Spell checking, markdown linting | Pre-commit hooks |
| **Java** | 21 LTS | Gatling load tests | Performance testing (optional) |
| **Maven** | 3.9+ | Java build tool | Gatling tests (optional) |
| **Git** | 2.40+ | Version control | All development |

### Go Development Tools

| Tool | Installation | Purpose |
|------|-------------|---------|
| **golangci-lint** | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.7.2` | Comprehensive linting (includes 50+ linters) |
| **gofumpt** | `go install mvdan.cc/gofumpt@latest` | Strict Go formatting (gofmt superset) |
| **goimports** | `go install golang.org/x/tools/cmd/goimports@latest` | Auto-import organization |
| **gopls** | `go install golang.org/x/tools/gopls@latest` | Go language server for VS Code |
| **staticcheck** | `go install honnef.co/go/tools/cmd/staticcheck@latest` | Advanced static analysis |
| **govulncheck** | `go install golang.org/x/vuln/cmd/govulncheck@latest` | Go vulnerability scanning |
| **gremlins** | `go install github.com/go-gremlins/gremlins/cmd/gremlins@latest` | Mutation testing |
| **oapi-codegen** | `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest` | OpenAPI code generation |

### Security & Testing Tools

| Tool | Installation | Purpose |
|------|-------------|---------|
| **Trivy** | Platform-specific | Container vulnerability scanning |
| **Nuclei** | Platform-specific | Dynamic security scanning (DAST) |
| **act** | Platform-specific | Local GitHub Actions testing |

### Node.js Tools

| Tool | Installation | Purpose |
|------|-------------|---------|
| **cspell** | `npm install -g cspell` | Spell checking in code/docs |
| **markdownlint-cli2** | `npm install -g markdownlint-cli2` | Markdown linting with auto-fix |

### Python Tools

| Tool | Installation | Purpose |
|------|-------------|---------|
| **pre-commit** | `pip install pre-commit` | Git hooks framework |
| **pytest** | `pip install pytest` | Python testing (if writing Python) |

---

## Why Each Tool is Required

### Core Development (Go 1.25.5+)

**Why this version?** Go 1.25.5 is required for:

- Generic type improvements used in internal packages
- Performance optimizations in crypto packages
- Bug fixes in `modernc.org/sqlite` CGO-free driver

**Project dependencies requiring Go 1.25.5+:**

```text
github.com/gofiber/fiber/v2       # HTTP framework
gorm.io/gorm                       # ORM for PostgreSQL/SQLite
modernc.org/sqlite                 # CGO-free SQLite driver
go.opentelemetry.io/otel          # Observability (traces, metrics, logs)
github.com/go-jose/go-jose/v4     # JOSE (JWK, JWS, JWE) operations
github.com/testcontainers/testcontainers-go  # Integration testing
```

### Docker & Docker Compose

**Why required?**

- **PostgreSQL**: Primary production database runs in containers
- **E2E Tests**: Integration tests use testcontainers-go
- **Observability Stack**: OpenTelemetry Collector, Grafana LGTM in containers
- **CI/CD Parity**: Local testing matches GitHub Actions environment

### Python 3.14+

**Why required?**

- **pre-commit hooks**: Framework for automated code quality checks
- **Custom CICD tools**: Some utility scripts use Python
- **Type hints**: Python 3.14+ required for `pyproject.toml` configuration

### Node.js 24.11.1+ LTS

**Why required?**

- **cspell**: Catches spelling errors in code comments and documentation
- **markdownlint-cli2**: Enforces consistent markdown formatting

### Java 21 LTS (Optional - Load Testing Only)

**Why required?**

- **Gatling**: Industry-standard load testing tool (Scala/Java-based)
- **Maven**: Build tool for Gatling test projects
- **Location**: `test/load/` directory

**Skip if:** Not running performance/load tests.

### golangci-lint v2.7.2+

**Why this specific tool and version?**

- Runs 50+ linters in a single pass (efficient CI/CD)
- Built-in gofumpt and goimports with `--fix` flag
- v2.x has breaking config changes from v1.x (see `.golangci.yml`)
- Pre-commit hooks depend on this version

### Gremlins (Mutation Testing)

**Why required?**

- Validates test quality by introducing code mutations
- Ensures tests actually catch bugs (not just coverage theater)
- CI/CD workflow `ci-mutation.yml` requires this tool

### Security Tools (Trivy, Nuclei)

**Why required?**

- **Trivy**: Scans container images for CVEs before deployment
- **Nuclei**: Dynamic security testing (DAST) for API endpoints
- Used by CI/CD workflows `ci-sast.yml` and `ci-dast.yml`

---

## Platform-Specific Setup

### Windows

#### 1. Install Core Prerequisites

**Go 1.25.5+**

```powershell
# Download and install from https://golang.org/dl/
# Or use winget:
winget install --id Google.Go --source winget

# Verify installation
go version
# Expected: go version go1.25.5 windows/amd64
```

**Docker Desktop**

```powershell
# Download and install from https://www.docker.com/products/docker-desktop
# Or use winget:
winget install --id Docker.DockerDesktop --source winget

# Verify installation
docker --version
docker compose version
```

**Python 3.14+**

```powershell
# Download and install from https://python.org/downloads/
# Or use winget:
winget install --id Python.Python.3.14 --source winget

# Verify installation (may need to restart terminal)
python --version
# Expected: Python 3.14.x
pip --version
```

**Node.js 24.11.1+ LTS**

```powershell
# Option 1: Download from https://nodejs.org/
# Option 2: Use winget
winget install --id OpenJS.NodeJS.LTS --source winget

# Option 3: Use nvm-windows for version management
# Download nvm-windows from: https://github.com/coreybutler/nvm-windows/releases
nvm install 24.11.1
nvm use 24.11.1

# Verify installation
node --version
npm --version
```

**Java 21 LTS (Optional - For Gatling Load Tests)**

```powershell
# Download from https://adoptium.net/temurin/releases/?version=21
# Or use winget:
winget install --id EclipseAdoptium.Temurin.21.JDK --source winget

# Set JAVA_HOME environment variable
[System.Environment]::SetEnvironmentVariable("JAVA_HOME", "C:\Program Files\Eclipse Adoptium\jdk-21.0.x.x-hotspot", "User")

# Verify installation (restart terminal)
java -version
# Expected: openjdk version "21.0.x"
```

**Maven 3.9+ (Optional - For Gatling Load Tests)**

```powershell
# Download from https://maven.apache.org/download.cgi
# Or use winget:
winget install --id Apache.Maven --source winget

# Or use the Maven Wrapper included in test/load/ (recommended)
cd test\load
.\mvnw --version
```

**Git**

```powershell
# Usually pre-installed, otherwise:
winget install --id Git.Git --source winget

# Verify
git --version
```

#### 2. Install Go Development Tools

```powershell
# Install golangci-lint (comprehensive linting)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.7.2

# Install gofumpt (strict Go formatting)
go install mvdan.cc/gofumpt@latest

# Install goimports (import organization)
go install golang.org/x/tools/cmd/goimports@latest

# Install gopls (Go language server)
go install golang.org/x/tools/gopls@latest

# Install staticcheck (advanced static analysis)
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck (vulnerability scanning)
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install gremlins (mutation testing)
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

# Install oapi-codegen (OpenAPI code generation)
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Verify all tools
golangci-lint --version
gofumpt --version
gopls version
staticcheck --version
govulncheck --version
gremlins --version
```

#### 3. Install Node.js Tools

```powershell
# Install cspell (spell checking)
npm install -g cspell

# Install markdownlint-cli2 (markdown linting)
npm install -g markdownlint-cli2

# Verify
cspell --version
markdownlint-cli2 --version
```

#### 4. Install Python Tools

```powershell
# Install pre-commit
pip install pre-commit

# Verify
pre-commit --version
```

#### 5. Install Security & Testing Tools

```powershell
# Install Trivy (container vulnerability scanning)
# Download from: https://github.com/aquasecurity/trivy/releases
# Extract and add to PATH

# Install Nuclei (DAST scanning)
# Download from: https://github.com/projectdiscovery/nuclei/releases
# Extract and add to PATH

# Install act (local GitHub Actions testing)
# Download from: https://github.com/nektos/act/releases
# Extract and add to PATH

# Verify
trivy --version
nuclei --version
act --version
```

#### 6. Configure PowerShell Execution Policy

**Important Security Requirement:** PowerShell's default execution policy prevents scripts from running.

```powershell
# Check current policy
Get-ExecutionPolicy -List

# Set policy to allow local scripts (recommended for development)
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force

# Verify
Get-ExecutionPolicy -Scope CurrentUser
# Expected: RemoteSigned
```

#### 7. Configure Go Environment Variables (Performance Optimization)

```powershell
# Set Go temp directory (avoids antivirus scanning delays)
setx GOTMPDIR "C:\Temp\go-tmp"

# Set Go cache directory
setx GOCACHE "C:\Temp\go-tmp"

# Verify (restart terminal required)
go env GOTMPDIR
go env GOCACHE
```

---

### Linux

#### 1. Install Core Prerequisites

**Ubuntu/Debian:**

```bash
# Update package list
sudo apt update

# Install Go 1.25.5+
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version

# Install Docker
sudo apt install -y docker.io docker-compose-plugin
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Verify (logout/login required for group membership)
docker --version
docker compose version

# Install Python 3.14+
sudo apt install -y python3 python3-pip python3-venv

# Verify
python3 --version

# Install Node.js 24.x LTS
curl -fsSL https://deb.nodesource.com/setup_24.x | sudo -E bash -
sudo apt install -y nodejs

# Verify
node --version
npm --version

# Install Git
sudo apt install -y git

# Verify
git --version
```

**Java 21 LTS (Optional - For Gatling Load Tests)**

```bash
# Install Java 21
sudo apt install -y openjdk-21-jdk

# Set JAVA_HOME
echo 'export JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64' >> ~/.bashrc
source ~/.bashrc

# Verify
java -version

# Maven is included via Maven Wrapper in test/load/
# Or install globally:
sudo apt install -y maven
```

**Fedora/RHEL/CentOS:**

```bash
# Install Go
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

# Install Docker
sudo dnf install -y docker docker-compose-plugin
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Install Python
sudo dnf install -y python3 python3-pip

# Install Node.js
curl -fsSL https://rpm.nodesource.com/setup_24.x | sudo bash -
sudo dnf install -y nodejs

# Install Git
sudo dnf install -y git

# Install Java 21 (optional)
sudo dnf install -y java-21-openjdk-devel
```

#### 2. Install Go Development Tools

```bash
# Install all Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.7.2
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/gopls@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Verify
golangci-lint --version
```

#### 3. Install Node.js Tools

```bash
npm install -g cspell markdownlint-cli2
```

#### 4. Install Python Tools

```bash
pip3 install pre-commit
```

#### 5. Install Security & Testing Tools

```bash
# Install Trivy
sudo apt install -y wget apt-transport-https gnupg lsb-release
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/trivy.list
sudo apt update
sudo apt install -y trivy

# Install Nuclei
go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest

# Install act (local GitHub Actions)
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

---

### macOS

#### 1. Install Core Prerequisites

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go 1.25.5+
brew install go

# Install Docker Desktop
brew install --cask docker

# Install Python 3.14+
brew install python@3.14

# Install Node.js 24.x LTS
brew install node@24

# Install Git
brew install git

# Install Java 21 (optional - for Gatling)
brew install openjdk@21
sudo ln -sfn /opt/homebrew/opt/openjdk@21/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-21.jdk

# Verify all
go version
docker --version
python3 --version
node --version
git --version
java -version  # if installed
```

#### 2. Install Go Development Tools

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.7.2
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/gopls@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

#### 3. Install Node.js & Python Tools

```bash
npm install -g cspell markdownlint-cli2
pip3 install pre-commit
```

#### 4. Install Security & Testing Tools

```bash
brew install trivy nuclei act
```

---

## Common Setup Steps

### 1. Clone the Repository

```bash
git clone https://github.com/justincranford/cryptoutil.git
cd cryptoutil
```

### 2. Initialize Go Modules

```bash
go mod tidy
go mod download
```

### 3. Install Pre-commit Hooks

```bash
pre-commit install
pre-commit install --hook-type commit-msg
pre-commit install --hook-type pre-push
```

### 4. Generate OpenAPI Code

```bash
go generate ./...
```

### 5. Build and Test

```bash
# Build all packages
go build ./...

# Run tests
go test ./... -cover

# Run linting
golangci-lint run --timeout=10m
```

### 6. Verify Pre-commit Hooks

```bash
# Run all hooks on all files
pre-commit run --all-files
```

---

## IDE Configuration

### VS Code Setup

1. **Install VS Code**: Download from <https://code.visualstudio.com/>

2. **Install Go Extension**: Search "Go" by Google in Extensions

3. **gopls Configuration**: The project includes optimized `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.formatOnSave": true,
  "gopls": {
    "formatting.gofumpt": true,
    "ui.inlayhint.hints": {
      "assignVariableTypes": true,
      "parameterNames": true,
      "rangeVariableTypes": true
    }
  }
}
```

### Troubleshooting VS Code

```bash
# If gopls not found, add to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Clear gopls cache if slow
rm -rf $(go env GOPATH)/pkg/mod/cache/gopls

# Restart language server in VS Code
# Ctrl+Shift+P → "Go: Restart Language Server"
```

---

## Project Setup

### Development with SQLite (Recommended)

```bash
# SQLite uses in-memory database - no setup required
go run ./cmd/cryptoutil/main.go --dev
```

### Development with PostgreSQL

```bash
# Start PostgreSQL with Docker Compose
cd deployments/compose
docker compose up -d postgres

# Run with PostgreSQL config
go run ./cmd/cryptoutil/main.go --config=configs/dev/postgresql.yml
```

### Access the Application

- **API Documentation**: <https://localhost:8080/ui/swagger>
- **Health Checks**: <https://localhost:9090/admin/api/v1/livez>, <https://localhost:9090/admin/api/v1/readyz>
- **Grafana**: <http://localhost:3000> (admin/admin)

---

## Verification

### Run Full Verification

```bash
# Check all tool versions
echo "=== Tool Versions ==="
go version
docker --version
docker compose version
python3 --version
node --version
golangci-lint --version
gofumpt --version
pre-commit --version
cspell --version

# Optional tools
java -version 2>/dev/null || echo "Java not installed (optional)"
trivy --version 2>/dev/null || echo "Trivy not installed (optional)"
nuclei --version 2>/dev/null || echo "Nuclei not installed (optional)"
act --version 2>/dev/null || echo "Act not installed (optional)"

# Project verification
echo "=== Project Build ==="
go build ./...

echo "=== Tests ==="
go test ./... -cover -short

echo "=== Linting ==="
golangci-lint run --timeout=10m

echo "=== Pre-commit Hooks ==="
pre-commit run --all-files
```

---

## Troubleshooting

### Common Issues

**Go tools not found in PATH:**

```bash
export PATH=$PATH:$(go env GOPATH)/bin
# Add to shell profile for persistence
```

**Docker permission issues (Linux):**

```bash
sudo usermod -aG docker $USER
# Logout and login again
```

**Pre-commit hooks not working:**

```bash
pre-commit install --force
pre-commit clean
```

**PowerShell execution policy reset after reboot (Windows):**

```powershell
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force
```

**testcontainers-go fails on Windows:**

```text
Error: "testcontainers-go does not support rootless Docker"
Solution: Ensure Docker Desktop is running (not just installed)
```

### Getting Help

- Check [README.md](../README.md) for application documentation
- Check GitHub Issues for known problems
- Run diagnostic commands shown in Verification section

---

## Gatling Load Tests (Optional)

For performance testing, the project includes Gatling tests in `test/load/`.

### Prerequisites

- Java 21 LTS
- Maven 3.9+ (or use included Maven Wrapper)

### Running Load Tests

```bash
cd test/load

# Using Maven Wrapper (recommended - no Maven install needed)
./mvnw gatling:test

# Or with Maven installed
mvn gatling:test
```

### JDK Configuration for Gatling

Create `.mvn/jvm.config` in `test/load/` directory:

```properties
--java-home
/path/to/your/jdk21
```

Or set `JAVA_HOME`:

```bash
export JAVA_HOME=/path/to/jdk21
```

See [test/load/README.md](../test/load/README.md) for detailed Gatling documentation.

---

## Security Testing (Optional)

### Manual Nuclei Scanning

```bash
# Start services first
cd deployments/compose
docker compose up -d

# Wait for services to be ready (30-60 seconds)
sleep 60

# Quick scan
nuclei -target https://localhost:8080/ -severity info,low

# Comprehensive scan
nuclei -target https://localhost:8080/ -severity medium,high,critical

# Update templates before scanning
nuclei -update-templates
```

### Container Vulnerability Scanning

```bash
# Scan Docker images
trivy image cryptoutil:latest
```

---

This setup guide ensures you have all the tools needed for cryptoutil development. The project includes automated scripts and configurations to make the setup process as smooth as possible across all supported platforms.
