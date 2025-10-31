---
description: "Instructions for platform-specific tooling: PowerShell, scripts, Docker pre-pull"
applyTo: "**"
---
# Platform-Specific Tooling Instructions

## PowerShell Development

### Execution
- Use `powershell -NoProfile -ExecutionPolicy Bypass -File script.ps1` (process-scoped)
- Alternative: `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass`
- NEVER change `-Scope LocalMachine` or `-Scope CurrentUser`

### Syntax
- Use `;` for chaining, `$env:VAR` for environment variables
- Bash equivalents: `| Select-Object -First 10` (head), `| Select-String` (grep)
- Avoid emojis/Unicode in here-strings
- No `&&` or `||` in v5.1 — use `;` or `if ($LASTEXITCODE -eq 0) { cmd2 }`

### Common Errors
- Switch defaults: avoid `[switch]$Param = $true`
- String interpolation: use single quotes `'literal $var'` for literals
- Variable expansion: use `${variable}` for clarity
- File slicing: read to array first: `$lines = Get-Content file; $lines[10..20]`

## Cross-Platform Script Development

### Language Preferences

**ALWAYS prefer Go for new scripts** when:
- Cross-platform compatibility needed
- Complex logic or data processing required
- Performance-critical operations
- Need for static compilation and distribution
- Structured error handling and testing required

**PowerShell/Bash appropriate** when:
- Platform-specific system administration tasks
- Quick automation of platform-specific workflows
- Simple file operations or environment setup

### Go Script Guidelines
- Place scripts in `scripts/` directory
- Use `cmd/` subdirectory for main entry points
- Follow standard Go project layout
- Use Go modules for dependency management
- Implement proper error handling with context
- Use structured logging (slog package)
- Provide CLI interface with flag package

### PowerShell/Bash Script Pairings
- Always provide both .ps1 and .sh versions
- Follow existing patterns (lint.ps1/lint.sh model)
- Ensure consistent functionality and parameter naming
- Provide consistent help/usage information
- Ensure executable permissions for Unix scripts

### Script Testing (CRITICAL)
- **ALWAYS test scripts before committing**
- Test both PowerShell and Bash versions on their platforms
- For PowerShell: Test with execution policy restrictions
- Test help/usage: `script.ps1 -Help` and `script.sh --help`
- Test error conditions and edge cases
- Verify cleanup and resource management
- Test with different parameter combinations

## Docker Image Pre-Pull Optimization

### Pattern for CI/CD Workflows

**Add pre-pull step BEFORE first Docker usage**:

```yaml
- name: Pre-pull Docker images (parallel)
  run: |
    echo "🐳 Pre-pulling all Docker images in parallel..."

    IMAGES=(
      "postgres:18"
      "ghcr.io/zaproxy/zaproxy:stable"
      "alpine:latest"
    )

    for image in "${IMAGES[@]}"; do
      echo "Pulling $image..."
      docker pull "$image" &
    done

    wait

    echo "✅ All images pre-pulled successfully"
```

### Common Images by Workflow

**DAST Workflow**:
- `postgres:18`
- `ghcr.io/zaproxy/zaproxy:stable`

**E2E Workflow**:
- `postgres:18`
- `alpine:3.19`
- `golang:1.25.1`

**Quality Workflow**:
- `alpine:3.19`
- `golang:1.25.1`

### Benefits
- **Parallel downloads**: All images download simultaneously
- **Faster workflows**: Reduces total pull time by 50-80%
- **Better diagnostics**: Clear separation of pull failures vs runtime failures
- **Cached layers**: Docker BuildKit cache benefits from having base images present

### When to Skip
- Workflow uses only GitHub Actions (no direct Docker commands)
- Single image used once (minimal benefit)
- Images already cached by previous steps
