---
description: "Instructions for consistent linting exclusions across pre-commit, CI/CD, and scripts"
applyTo: "**"
---
# Linting Exclusions Instructions

## Standard Exclusions

Always exclude these from linting operations:

### Generated Code
- `_gen.go` - Auto-generated Go files
- `.pb.go` - Protocol buffer files
- `api/` - OpenAPI generated code

### Dependencies
- `vendor/` - Vendored dependencies

### Build Artifacts
- `.exe`, `.dll`, `.so`, `.dylib` - Binaries
- `*.key`, `*.crt`, `*.pem` - Certificates/keys

### IDE Files
- `.vscode/` - VS Code settings

## Application

Use regex pattern: `'_gen\.go$|\.pb\.go$|vendor/|api/'`

### Pre-commit
```yaml
exclude: '_gen\.go$|\.pb\.go$|vendor/|api/'
```

### CI/CD
```yaml
skip-files: '.*_gen\.go$|.*\.pb\.go$'
skip-dirs: vendor,api
```

### Scripts
```bash
golangci-lint run --skip-files='.*_gen\.go$|.*\.pb\.go$' --skip-dirs=vendor,api
```

### golangci-lint Config
```yaml
issues:
  exclude-dirs: [vendor, api]
  exclude-files: [".*\\.pb\\.go$", ".*_gen\\.go$"]
```

## Maintenance

- Update exclusions when adding new generators
- Test exclusions after changes: `pre-commit run --all-files`</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\.github\instructions\linting-exclusions.instructions.md
