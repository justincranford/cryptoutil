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

## Performance Testing

See `scripts/perf/README.doc` for detailed performance testing documentation and usage.

## Security Scripts

## Utility Scripts

### cicd_checks.go

Go utility for CI/CD dependency and version checking.

```bash
# Check Go dependency versions (direct dependencies only)
go run scripts/cicd_checks.go go-update-direct-dependencies

# Check Go dependency versions (all dependencies) - not commonly done in Go, but util supports it
go run scripts/cicd_checks.go go-update-all-dependencies

# Check GitHub Actions versions
go run scripts/cicd_checks.go github-action-versions

# Check for circular dependencies in Go packages
go run scripts/cicd_checks.go go-check-circular-package-dependencies

# Check all Go dependencies, GitHub Actions versions, and circular dependencies in a single invocation
go run scripts/cicd_checks.go go-update-direct-dependencies github-action-versions go-check-circular-package-dependencies
```

### count_tokens.py

Token counting utility using tiktoken for estimating AI model costs.

```bash
# Count tokens for instruction files
python .\scripts\count_tokens.py --model gpt-4o --glob ".github/instructions/*.md" --as-message system

# Count tokens for single file
python .\scripts\count_tokens.py --file .github/copilot-instructions.md --as-message none --model gpt-4o
```

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
- `act` (GitHub Actions local testing)

## Integration with CI/CD

These scripts mirror the functionality available in GitHub Actions workflows:
- Security scanning and DAST testing are handled by `run_github_workflow_locally.go` for local testing and GitHub Actions workflows for CI/CD
- Performance testing can be integrated into CI/CD pipelines
