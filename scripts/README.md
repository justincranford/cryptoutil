# Scripts Directory

This directory contains utility scripts for development, testing, building, and security scanning of the cryptoutil project.

## Development Setup

### Pre-commit Hooks Setup

**Windows Batch:**
```batch
.\scripts\setup-pre-commit.bat
```

**Windows PowerShell:**
```powershell
.\scripts\setup-pre-commit.ps1
```

These scripts install pre-commit hooks that run automated code quality checks on every commit.

## Build Scripts

### build.ps1

Docker build script with mandatory version tagging and proper build arguments.

```powershell
.\scripts\build.ps1 -AppVersion v1.0.0
```

**Features:**
- Validates mandatory app version parameter
- Sets VCS_REF to current git commit hash
- Sets BUILD_DATE to current timestamp
- Builds cryptoutil Docker image with proper tagging

## Performance Testing

See `scripts/perf/README.doc` for detailed performance testing documentation and usage.

## Security Scripts

### security-scan.ps1 / security-scan.sh

Runs comprehensive security scans locally (same as CI/CD pipeline).

```powershell
# PowerShell - All scans
.\scripts\security-scan.ps1

# PowerShell - Static analysis only
.\scripts\security-scan.ps1 -StaticOnly

# PowerShell - Skip Docker scans
.\scripts\security-scan.ps1 -SkipDocker

# Bash
./scripts/security-scan.sh --skip-docker
```

**Scans included:**
- **Static Analysis**: staticcheck, golangci-lint
- **Vulnerability Scanning**: govulncheck, Trivy filesystem
- **Container Security**: Trivy image scan, Docker Scout

### dast.ps1 / dast.sh

Runs Dynamic Application Security Testing (DAST) with OWASP ZAP and Nuclei.

```powershell
# PowerShell - Default settings
.\scripts\dast.ps1

# PowerShell - Custom config and port
.\scripts\dast.ps1 -Config "configs/test/config.yml" -Port 9090

# Bash
./scripts/dast.sh -c configs/test/config.yml -p 9090
```

**Features:**
- Starts cryptoutil application automatically
- Runs OWASP ZAP full scan and API scan
- Runs Nuclei security scans
- Generates comprehensive reports

### run-act-dast.ps1

Advanced script for running GitHub Actions DAST workflows locally with `act`.

```powershell
# Quick scan (3-5 minutes)
.\scripts\run-act-dast.ps1

# Full scan (10-15 minutes)
.\scripts\run-act-dast.ps1 -ScanProfile full -Timeout 900

# Deep scan (20-25 minutes)
.\scripts\run-act-dast.ps1 -ScanProfile deep -Timeout 1500
```

**Features:**
- Automated background execution
- Real-time progress monitoring
- Automatic completion detection
- Comprehensive result analysis
- Artifact verification

## Utility Scripts

### cicd_utils.go

Go utility for CI/CD dependency and version checking.

```bash
# Check Go dependency versions (direct dependencies only)
go run scripts/cicd_utils.go go-update-direct-dependencies

# Check Go dependency versions (all dependencies) - not commonly done in Go, but util supports it
go run scripts/cicd_utils.go go-update-all-dependencies

# Check GitHub Actions versions
go run scripts/cicd_utils.go github-action-versions

# Check for circular dependencies in Go packages
go run scripts/cicd_utils.go go-check-circular-package-dependencies

# Check all Go dependencies, GitHub Actions versions, and circular dependencies in a single invocation
go run scripts/cicd_utils.go go-update-direct-dependencies github-action-versions go-check-circular-package-dependencies
```

### count_tokens.py

Token counting utility using tiktoken for estimating AI model costs.

```bash
# Count tokens for instruction files
python .\scripts\count_tokens.py --model gpt-4o --glob ".github/instructions/*.md" --as-message system

# Count tokens for single file
python .\scripts\count_tokens.py --file .github/copilot-instructions.md --as-message none --model gpt-4o
```

## Mutation Testing Scripts

### mutation-test.ps1 / mutation-test.sh

Runs mutation testing using Gremlins to ensure test quality.

```powershell
# PowerShell - Default settings
.\scripts\mutation-test.ps1

# PowerShell - Custom timeout and workers
.\scripts\mutation-test.ps1 -Timeout 10m -Workers 4

# Bash
./scripts/mutation-test.sh -t 10m -w 4
```

**Features:**
- Automated mutation testing with Gremlins
- Configurable timeout and worker count
- Test quality validation
- Cross-platform support

## Configuration Files

### errcheck_excludes.txt

Configuration file for errcheck tool exclusions used in pre-commit hooks.

## Documentation

### README-run-act-dast.md

Detailed documentation for the `run-act-dast.ps1` script with usage examples and troubleshooting.

## Cross-Platform Support

- **Windows**: PowerShell scripts (`.ps1`) with batch file alternatives (`.bat`)
- **Linux/macOS**: Bash scripts (`.sh`)
- **Go utilities**: Cross-platform (`.go`)
- **Python utilities**: Cross-platform (`.py`)

## Execution Policy (Windows)

For PowerShell scripts on Windows, use one of these approaches:

**Process-scoped (recommended for one-time execution):**
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\script.ps1
```

**Session-scoped:**
```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass; .\scripts\script.ps1
```

## Prerequisites

Most scripts will install required tools automatically if missing:
- `gremlins` (mutation testing)
- `staticcheck`, `golangci-lint`, `govulncheck` (security scanning)
- `trivy`, `docker-scout` (container scanning)
- `act` (GitHub Actions local testing)
- `k6` (performance testing)

## Integration with CI/CD

These scripts mirror the functionality available in GitHub Actions workflows:
- Security scanning matches `.github/workflows/security-scan.yml`
- DAST testing matches `.github/workflows/dast.yml`
- Mutation testing can be run locally for development validation
- Performance testing can be integrated into CI/CD pipelines
