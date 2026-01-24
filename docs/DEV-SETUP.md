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

## Prerequisites

- **Go 1.25.5+** - The project requires Go 1.25.5 or later
- **Docker & Docker Compose** - Required for PostgreSQL database and containerized testing
- **Git** - Version control (usually pre-installed on most systems)

### Development Tools

- **Python 3.8+** - Required for utility scripts and pre-commit hooks
- **pip** - Python package manager (usually comes with Python)
- **VS Code** - Recommended IDE with Go extension

## Platform-Specific Setup

### Windows

#### 1. Install Core Prerequisites

**Go 1.25.5+**

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

# Install staticcheck (advanced static analysis)
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck (vulnerability scanning)
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install cspell (spell checking)
npm install -g cspell
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

#### 4. Configure PowerShell Execution Policy

**Important Security Requirement:** PowerShell's default execution policy prevents scripts from running, including Python virtual environment activation.

```powershell
# Check current policy
Get-ExecutionPolicy -List

# Set policy to allow local scripts (recommended for development)
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force

# Alternative: Maximum security (requires all scripts to be signed)
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy AllSigned -Force
```

**Security Note:** `RemoteSigned` allows locally created scripts while blocking unsigned downloads. This is the industry standard for development environments. The VS Code integrated terminal is configured with the same `RemoteSigned` policy for consistency.

#### 5. Configure Go Environment Variables

**Performance Optimization:** Go stores compiled executables and build cache in user directories that may trigger antivirus scanning. Configure Go to use your custom temp directory for faster development workflows.

```powershell
# Set Go temp directory for compilation artifacts
setx GOTMPDIR "F:\go-tmp"

# Set Go cache directory for build artifacts and executables
setx GOCACHE "F:\go-tmp"

# Verify settings (restart terminal or new session required)
go env GOTMPDIR
go env GOCACHE
```

**Why This Matters:**

- `GOTMPDIR`: Controls where Go stores temporary compilation files
- `GOCACHE`: Controls where Go stores build cache and compiled executables
- Setting both to your excluded temp directory prevents antivirus scanning delays
- First run after changing may be slower (cache rebuild), subsequent runs are fast

**Expected Performance:**

- **Before**: `cicd` commands take 30-60+ seconds (antivirus scanning)
- **After**: Same commands complete in 2-5 seconds (excluded directory)

### Linux

#### 1. Install Core Prerequisites

**Ubuntu/Debian:**

```bash
# Update package list
sudo apt update

# Install Go 1.25.5+
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
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
```bash
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
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

# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install cspell
npm install -g cspell
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

# Install Go 1.25.5+
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

# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install cspell
npm install -g cspell
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

## Pre-commit Hooks

Install pre-commit hooks:

```bash
pip install pre-commit
pre-commit install
```

### Markdownlint Setup

Install markdownlint-cli2 for markdown formatting:

```bash
npm install -g markdownlint-cli2
```

Pre-commit hook automatically fixes markdown issues.

## Editor Setup

#### Linux

**1. Install pre-commit:**

```bash
# Install pre-commit globally
pip3 install pre-commit
```

**2. Install pre-commit hooks:**

```bash
# Navigate to project root
cd ~/cryptoutil

# Install the hooks
pre-commit install
```

**3. Test the setup:**

```bash
# Run all hooks on all files to verify installation
pre-commit run --all-files
```

#### macOS

**1. Install pre-commit:**

```bash
# Install pre-commit globally
pip3 install pre-commit
```

**2. Install pre-commit hooks:**

```bash
# Navigate to project root
cd ~/cryptoutil

# Install the hooks
pre-commit install
```

**3. Test the setup:**

```bash
# Run all hooks on all files to verify installation
pre-commit run --all-files
```

**Pre-commit Hook Details:**

- **Automatic execution**: Hooks run automatically on `git commit`
- **Manual execution**: Run `pre-commit run --all-files` to check all files
- **Selective execution**: Run `pre-commit run <hook-name>` for specific hooks
- **Cache location**: Configured to avoid antivirus interference on Windows
- **Performance**: First run may be slower (cache building), subsequent runs are fast

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
cspell --version

# Check security tools
trivy --version
act --version
```

## IDE Configuration

### VS Code Setup

1. **Install VS Code**
   - Download from <https://code.visualstudio.com/>

2. **Install Go Extension**
   - Open VS Code
   - Go to Extensions (Ctrl+Shift+X)
   - Search for "Go" by Google
   - Install the extension

3. **gopls Installation and Configuration**

   **gopls** is the official Go language server that powers VS Code's Go extension, providing intelligent code completion, navigation, and refactoring.

   **Installation:**

   ```bash
   # Install gopls (Go language server)
   go install golang.org/x/tools/gopls@latest

   # Verify installation
   gopls version
   # Expected output: gopls v0.X.X (or latest version)
   ```

   **VS Code Configuration:**

   The project's `.vscode/settings.json` is pre-configured with optimal gopls settings:

   ```json
   {
     "go.useLanguageServer": true,
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

   **Key Features Enabled:**
   - **Auto-import**: Automatically adds/removes imports on save
   - **Inlay Hints**: Shows type information inline for better readability
   - **gofumpt Integration**: Stricter formatting than standard gofmt
   - **Intelligent Refactoring**: F2 rename symbol with cross-file awareness

   **Troubleshooting:**

   1. **gopls not found**: Ensure `$(go env GOPATH)/bin` is in your PATH
      ```bash
      # Windows (PowerShell)
      $env:PATH += ";$(go env GOPATH)\bin"

      # Linux/macOS (Bash)
      export PATH=$PATH:$(go env GOPATH)/bin
      ```

   2. **Slow performance**: Clear gopls cache and restart
      ```bash
      # Clear gopls cache
      rm -rf $(go env GOPATH)/pkg/mod/cache/gopls

      # Restart VS Code
      # Or reload window: Ctrl+Shift+P → "Developer: Reload Window"
      ```

   3. **Import errors**: Run `go mod tidy` and restart gopls
      ```bash
      go mod tidy
      # Restart gopls in VS Code: Ctrl+Shift+P → "Go: Restart Language Server"
      ```

4. **Workspace Settings**
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

# Configure Go SDK to 1.25.5+

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

- **API Documentation**: <http://localhost:8080/ui/swagger>
- **Health Checks**: <http://localhost:9090/admin/v1/livez>, <http://localhost:9090/admin/v1/readyz>
- **Grafana**: <http://localhost:3000> (admin/admin)

### 4. Documentation Maintenance

**Keep these files up-to-date as the project evolves:**

- **`.github/instructions/`** - **CRITICAL: All Copilot instruction files** that guide AI-assisted development:
  - `copilot-customization.instructions.md` - VS Code Copilot behavior and tool usage
  - `code-quality.instructions.md` - Code standards and linting compliance
  - `testing.instructions.md` - Testing patterns and coverage requirements
  - `architecture.instructions.md` - Application architecture and design patterns
  - `security.instructions.md` - Security implementation guidelines
  - `commits.instructions.md` - Conventional commit message standards
  - `formatting.instructions.md` - Code formatting and encoding standards
  - `project-layout.instructions.md` - Go project structure conventions
  - And all other `.instructions.md` files for specific domains
- **`README.md`** - Main project documentation and usage examples
- **`docs/README.md`** - Deep-dive technical documentation and architecture details
- **`docs/DEV-SETUP.md`** - This development setup guide
- **`.vscode/settings.json`** - VS Code workspace configuration and Go language server settings
- **`.github/workflows/`** - CI/CD pipeline configurations and GitHub Actions
- **`scripts/`** - All utility scripts and their documentation

**When making changes:**

- Update tool versions and installation instructions when dependencies change
- Document new setup requirements or configuration options
- Keep platform-specific instructions current across Windows, Linux, and macOS
- Test setup instructions on clean systems periodically
- Update troubleshooting section with new common issues and solutions
- **Especially important**: Keep Copilot instruction files current as development practices evolve

## Verification

### Run Full Test Suite

```bash
# Run all tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run linting
golangci-lint run --timeout=10m

# Run security scans (via GitHub Actions workflows)
# Use: go run ./cmd/workflow -workflows=dast
```

### Manual Security Testing with Nuclei

[Nuclei](https://github.com/projectdiscovery/nuclei) is a fast, template-based vulnerability scanner included in the development environment for manual security testing.

#### Prerequisites: Start cryptoutil Services

Before running nuclei scans, start the cryptoutil services:

```bash
# Navigate to compose directory
cd deployments/compose

# Clean shutdown with volume removal
docker compose down -v

# Start all services
docker compose up -d

# Verify services are ready (may take 30-60 seconds)
curl -k https://localhost:8080/ui/swagger/doc.json  # SQLite instance
curl -k https://localhost:8081/ui/swagger/doc.json  # PostgreSQL instance 1
curl -k https://localhost:8082/ui/swagger/doc.json  # PostgreSQL instance 2
```

#### Manual Nuclei Scan Examples

**Service Endpoints:**

- **cryptoutil-sqlite**: `https://localhost:8080/` (SQLite backend, development instance)
- **cryptoutil-postgres-1**: `https://localhost:8081/` (PostgreSQL backend, production-like instance)
- **cryptoutil-postgres-2**: `https://localhost:8082/` (PostgreSQL backend, production-like instance)

**Basic Security Scans:**

```bash
# Quick scan - Info and Low severity (fast, ~5-10 seconds)
nuclei -target https://localhost:8080/ -severity info,low
nuclei -target https://localhost:8081/ -severity info,low
nuclei -target https://localhost:8082/ -severity info,low

# Standard scan - Medium, High, and Critical severity (~10-30 seconds)
nuclei -target https://localhost:8080/ -severity medium,high,critical
nuclei -target https://localhost:8081/ -severity medium,high,critical
nuclei -target https://localhost:8082/ -severity medium,high,critical

# Full scan - All severity levels (~20-60 seconds)
nuclei -target https://localhost:8080/ -severity info,low,medium,high,critical
```

**Targeted Vulnerability Scans:**

```bash
# CVE scanning (recent and historical vulnerabilities)
nuclei -target https://localhost:8080/ -tags cves -severity high,critical

# Security misconfigurations
nuclei -target https://localhost:8080/ -tags security-misconfiguration

# Information disclosure and exposure
nuclei -target https://localhost:8080/ -tags exposure,misc

# Technology detection and fingerprinting
nuclei -target https://localhost:8080/ -tags tech-detect
```

**Performance-Optimized Scans:**

```bash
# High-performance scanning (adjust concurrency and rate limiting as needed)
nuclei -target https://localhost:8080/ -c 25 -rl 100 -severity high,critical

# Conservative scanning (lower resource usage)
nuclei -target https://localhost:8080/ -c 10 -rl 25 -severity medium,high,critical
```

**Batch Scanning Script (PowerShell - Windows):**

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

**Batch Scanning Script (Bash - Linux/macOS):**

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

#### Nuclei Template Management

```bash
# Update nuclei templates to latest version
nuclei -update-templates

# Check current template version
nuclei -templates-version

# List available templates (shows first 20)
nuclei -tl | head -20

# Search for specific template types
nuclei -tl | grep -i "http"     # Linux/macOS
nuclei -tl | findstr http       # Windows PowerShell
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

### Test Key Features

```bash
# Test API endpoints
curl http://localhost:8080/service/api/v1/elastickeys

# Test health endpoints
curl http://localhost:9090/admin/v1/livez
curl http://localhost:9090/admin/v1/readyz

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

**PowerShell execution policy reset after reboot:**

```powershell
# If scripts stop working after system reboot
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force

# Verify the policy is set
Get-ExecutionPolicy -List
```

### Getting Help

- Check the main [README.md](../README.md) for application-specific documentation
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
