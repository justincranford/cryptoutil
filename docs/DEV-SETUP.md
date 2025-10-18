# Development Environment Setup

This guide covers setting up a complete development environment for the cryptoutil project on Windows, Linux, and macOS systems.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Platform-Specific Setup](#platform-specific-setup)
  - [Windows](#windows)
  - [Linux](#linux)
  - [macOS](#macos)
- [Common Setup Steps](#common-setup-steps)
- [IDE Configuration](#ide-configuration)
- [Project Setup](#project-setup)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Core Requirements (All Platforms)
- **Go 1.25.1+** - The project requires Go 1.25.1 or later
- **Docker & Docker Compose** - Required for PostgreSQL database and containerized testing
- **Git** - Version control (usually pre-installed on most systems)

### Development Tools
- **Python 3.8+** - Required for utility scripts and pre-commit hooks
- **pip** - Python package manager (usually comes with Python)
- **VS Code** - Recommended IDE with Go extension

## Platform-Specific Setup

### Windows

#### 1. Install Core Prerequisites

**Go 1.25.1+**
```powershell
# Download and install from https://golang.org/dl/
# Or use winget:
winget install --id Google.Go --source winget
```

**Docker Desktop**
```powershell
# Download and install from https://www.docker.com/products/docker-desktop
# Or use winget:
winget install --id Docker.DockerDesktop --source winget
```

**Python 3.8+**
```powershell
# Download and install from https://python.org/downloads/
# Or use winget:
winget install --id Python.Python.3 --source winget
```

**Git**
```powershell
# Usually pre-installed, otherwise:
winget install --id Git.Git --source winget
```

#### 2. Install Go Development Tools

```powershell
# Install golangci-lint (Go linting)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt (strict Go formatting)
go install mvdan.cc/gofumpt@latest

# Install goimports (import organization)
go install golang.org/x/tools/cmd/goimports@latest

# Install errcheck (error checking)
go install github.com/kisielk/errcheck@latest

# Install staticcheck (advanced static analysis)
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck (vulnerability scanning)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

#### 3. Install Security & Testing Tools

```powershell
# Install Trivy (container vulnerability scanning)
# Download from: https://github.com/aquasecurity/trivy/releases
# Add to PATH

# Install act (GitHub Actions local testing)
# Download from: https://github.com/nektos/act/releases
# Add to PATH

# Install Gremlins (mutation testing) - installed automatically by scripts
```

#### 4. Set Environment Variables

```powershell
# Add Go bin to PATH (usually done automatically by Go installer)
# Add Docker Desktop to PATH
# Add Python Scripts to PATH
```

### Linux

#### 1. Install Core Prerequisites

**Ubuntu/Debian:**
```bash
# Update package list
sudo apt update

# Install Go 1.25.1+
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Install Docker
sudo apt install -y docker.io docker-compose
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Install Python 3.8+
sudo apt install -y python3 python3-pip

# Install Git
sudo apt install -y git
```

**Fedora/RHEL/CentOS:**
```bash
# Install Go
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Install Docker
sudo dnf install -y docker docker-compose
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Install Python
sudo dnf install -y python3 python3-pip

# Install Git
sudo dnf install -y git
```

#### 2. Install Go Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt
go install mvdan.cc/gofumpt@latest

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Install errcheck
go install github.com/kisielk/errcheck@latest

# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
```

#### 3. Install Security & Testing Tools

```bash
# Install Trivy
sudo apt install -y wget apt-transport-https
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt update
sudo apt install -y trivy

# Install act (GitHub Actions)
curl -s https://api.github.com/repos/nektos/act/releases/latest | grep "browser_download_url.*linux-amd64" | cut -d '"' -f 4 | wget -qi -
sudo mv act /usr/local/bin/act
sudo chmod +x /usr/local/bin/act

# Gremlins installed automatically by project scripts
```

### macOS

#### 1. Install Core Prerequisites

**Using Homebrew (Recommended):**
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go 1.25.1+
brew install go

# Install Docker Desktop
brew install --cask docker

# Install Python 3.8+
brew install python

# Install Git
brew install git
```

**Manual Installation:**
```bash
# Go - Download from https://golang.org/dl/
# Docker Desktop - Download from https://www.docker.com/products/docker-desktop
# Python - Download from https://python.org/downloads/
# Git - Usually pre-installed
```

#### 2. Install Go Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt
go install mvdan.cc/gofumpt@latest

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Install errcheck
go install github.com/kisielk/errcheck@latest

# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
```

#### 3. Install Security & Testing Tools

```bash
# Install Trivy
brew install trivy

# Install act (GitHub Actions)
brew install act

# Gremlins installed automatically by project scripts
```

## Common Setup Steps

### 1. Clone the Repository

```bash
# Clone the project
git clone https://github.com/justincranford/cryptoutil.git
cd cryptoutil

# Initialize Go modules
go mod tidy
```

### 2. Setup Pre-commit Hooks

**Windows:**
```powershell
# Using PowerShell script (recommended)
.\scripts\setup-pre-commit.ps1

# Or using batch file
.\scripts\setup-pre-commit.bat
```

**Linux/macOS:**
```bash
# Install pre-commit
pip3 install pre-commit

# Install hooks
pre-commit install

# Test setup
pre-commit run --all-files
```

### 3. Generate OpenAPI Code

```bash
# Generate OpenAPI client/server code
go generate ./...
```

### 4. Verify Installation

```bash
# Check Go version
go version

# Check Docker
docker --version
docker compose version

# Check Python
python3 --version

# Check Go tools
golangci-lint --version
gofumpt --version
goimports --version

# Check security tools
trivy --version
act --version
```

## IDE Configuration

### VS Code Setup

1. **Install VS Code**
   - Download from https://code.visualstudio.com/

2. **Install Go Extension**
   - Open VS Code
   - Go to Extensions (Ctrl+Shift+X)
   - Search for "Go" by Google
   - Install the extension

3. **Workspace Settings**
   The project includes optimized VS Code settings in `.vscode/settings.json` that provide:
   - Intelligent Go language server configuration
   - Automatic formatting and linting
   - Enhanced code completion and inlay hints
   - F2 rename symbol support for intelligent variable naming

4. **Key VS Code Settings Applied:**
   ```json
   {
     "go.useLanguageServer": true,
     "go.formatOnSave": true,
     "go.lintOnSave": "package",
     "go.vetOnSave": "package",
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

### Alternative IDEs

**GoLand/IntelliJ IDEA:**
- Install Go plugin
- Import project as Go module
- Configure Go SDK to 1.25.1+

**Vim/Neovim:**
- Install vim-go plugin
- Configure gopls integration

## Project Setup

### 1. Environment Configuration

**Development with SQLite (Recommended for development):**
```bash
# Copy SQLite config
cp configs/test/config.yml configs/dev/config.yml

# Edit as needed for your environment
```

**Development with PostgreSQL:**
```bash
# Start PostgreSQL with Docker Compose
cd deployments/compose
docker compose up -d postgres

# Use PostgreSQL config
cp deployments/compose/cryptoutil/postgresql.yml configs/dev/config.yml
```

### 2. Build and Test

```bash
# Build the project
go build ./...

# Run tests
go test ./... -cover

# Run with development config
go run main.go --dev --config=configs/dev/config.yml
```

### 3. Access the Application

- **API Documentation**: http://localhost:8080/ui/swagger
- **Health Checks**: http://localhost:9090/livez, http://localhost:9090/readyz
- **Grafana**: http://localhost:3000 (admin/admin)

## Verification

### Run Full Test Suite

```bash
# Run all tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run linting
golangci-lint run --timeout=10m

# Run security scans
./scripts/security-scan.sh  # Linux/macOS
.\scripts\security-scan.ps1 # Windows
```

### Test Key Features

```bash
# Test API endpoints
curl http://localhost:8080/service/api/v1/elastickeys

# Test health endpoints
curl http://localhost:9090/livez
curl http://localhost:9090/readyz

# Test with Swagger UI
open http://localhost:8080/ui/swagger
```

## Troubleshooting

### Common Issues

**Go tools not found in PATH:**
```bash
# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin
# Add to shell profile for persistence
```

**Docker permission issues (Linux):**
```bash
sudo usermod -aG docker $USER
# Logout and login again, or run: newgrp docker
```

**Pre-commit hooks not working:**
```bash
# Reinstall hooks
pre-commit install --force

# Clear cache
pre-commit clean
```

**VS Code Go extension issues:**
```bash
# Restart gopls
# In VS Code: Ctrl+Shift+P → "Go: Restart Language Server"

# Check gopls version
gopls version
```

**Python/pip issues:**
```bash
# Ensure python3 and pip3 are used
python3 --version
pip3 --version

# Upgrade pip
pip3 install --upgrade pip
```

### Getting Help

- Check the main [README.md](../README.md) for application-specific documentation
- Review [scripts/README.md](../scripts/README.md) for available utility scripts
- Check GitHub Issues for known problems
- Run diagnostic commands:
  ```bash
  # Check all tool versions
  go version && docker --version && python3 --version && golangci-lint --version

  # Test pre-commit
  pre-commit run --all-files

  # Test Go build
  go build ./...
  ```

### Performance Tips

- **Use SQLite for development** - faster than PostgreSQL for local development
- **Enable Docker BuildKit** - add `export DOCKER_BUILDKIT=1` to your shell profile
- **Use pre-commit hooks** - they run automatically and catch issues early
- **Keep Go modules tidy** - run `go mod tidy` regularly

---

This setup guide ensures you have all the tools needed for cryptoutil development. The project includes automated scripts and configurations to make the setup process as smooth as possible across all supported platforms.</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\docs\DEV-SETUP.md
