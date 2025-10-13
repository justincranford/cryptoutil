---
description: "Instructions for consistent linting exclusions across pre-commit, CI/CD, and scripts"
applyTo: "**"
---
# Linting Exclusions Instructions

## Standard Exclusions

Always exclude these file types and directories from linting operations:

### Generated Code
- **`_gen.go`** - Auto-generated Go files (oapi-codegen, etc.)
- **`.pb.go`** - Protocol buffer generated files
- **`api/`** - OpenAPI generated client/server code

### Dependencies
- **`vendor/`** - Vendored Go dependencies

### Build Artifacts
- **`.exe`, `.dll`, `.so`, `.dylib`** - Compiled binaries
- **`*.key`, `*.crt`, `*.pem`** - Certificates and keys

### IDE/Editor Files
- **`.vscode/`** - VS Code settings (JSONC files with comments)

## Application Context

### Pre-commit Hooks
Use consistent regex patterns in `.pre-commit-config.yaml`:
```yaml
exclude: '_gen\.go$|\.pb\.go$|vendor/|api/'
```

### CI/CD Workflows
Apply exclusions in GitHub Actions and other CI systems:
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    args: --timeout=10m
    skip-files: '.*_gen\.go$|.*\.pb\.go$'
    skip-dirs: vendor,api
```

### Local Scripts
Include exclusions in custom linting scripts:
```bash
# Bash script
golangci-lint run --skip-files='.*_gen\.go$|.*\.pb\.go$' --skip-dirs=vendor,api

# PowerShell script
golangci-lint run --skip-files='.*_gen\.go$|.*\.pb\.go$' --skip-dirs=vendor,api
```

### IDE Integration
Configure VS Code and other editors to ignore excluded files:
```json
// .vscode/settings.json
{
  "go.lintFlags": ["--skip-files=.*_gen\\.go$|.*\\.pb\\.go$", "--skip-dirs=vendor,api"],
  "go.vetFlags": ["--skip-files=.*_gen\\.go$|.*\\.pb\\.go$", "--skip-dirs=vendor,api"]
}
```

## Tool-Specific Patterns

### golangci-lint
```yaml
# .golangci.yml
issues:
  exclude-dirs:
    - vendor
    - api
  exclude-files:
    - ".*\\.pb\\.go$"
    - ".*_gen\\.go$"
```

### errcheck
```txt
# scripts/errcheck_excludes.txt
# Exclude patterns for errcheck tool
```

### Pre-commit Hooks
```yaml
# .pre-commit-config.yaml
exclude: '_gen\.go$|\.pb\.go$|vendor/|api/'
```

## Validation

### Pre-commit Testing
```bash
pre-commit run --all-files
```

### CI/CD Testing
Ensure workflows skip excluded files by checking build logs for unexpected linting of generated code.

### Local Testing
```bash
# Test exclusions work
find . -name "*.go" | grep -E '(_gen\.go$|\.pb\.go$)' | head -5
# Should return generated files that are properly excluded
```

## Maintenance

- **Review exclusions** when adding new code generators or build tools
- **Update patterns** if new file extensions are introduced
- **Test exclusions** after changes to ensure generated code isn't linted
- **Document exceptions** for any non-standard exclusions with clear rationale

## Rationale

Consistent exclusions prevent:
- False positive linting errors on generated code
- Wasted CI/CD time linting uneditable files
- Developer confusion from irrelevant warnings
- Build failures due to generated code quality issues</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\.github\instructions\linting-exclusions.instructions.md
