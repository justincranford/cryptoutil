---
description: "Instructions for cross-platform script development"
applyTo: "scripts/**"
---
# Cross-Platform Script Instructions

## When to Use Scripts
- Use scripts for complex, multi-step development workflows (lint, test, build, mutation testing, etc.)
- Scripts provide consistent cross-platform developer experience with parameter handling
- Note: For Docker Compose, prefer direct command directives over separate script files

## Script Development Guidelines
- Always provide both PowerShell (.ps1) and Bash (.sh) versions of scripts
- Follow existing script patterns in the scripts/ directory (lint.ps1/lint.sh model)
- Ensure consistent functionality and parameter naming across both versions
- Use platform-appropriate syntax:
  - PowerShell: `param()`, `Write-Host`, `$env:VAR`, `-ForegroundColor`
  - Bash: `while [[ $# -gt 0 ]]`, `echo`, `$VAR`, ANSI color codes
- Provide consistent help/usage information in both versions
- Ensure scripts have executable permissions for Unix-like systems
- Use descriptive commit messages when adding script pairs
- Document both script versions in README.md and relevant documentation

## Script Testing and Validation
- **ALWAYS test scripts before committing** - never commit untested scripts
- Test both PowerShell and Bash versions on their respective platforms
- For PowerShell scripts, test with execution policy restrictions:
  - Test with `powershell -ExecutionPolicy Bypass -File script.ps1`
  - Include `#Requires -Version 5.1` directive for minimum version
  - Document execution policy requirements in script comments
- Test help/usage functions: `script.ps1 -Help` and `script.sh --help`
- Test error conditions and edge cases (missing dependencies, invalid parameters)
- Verify script cleanup and resource management (process termination, file cleanup)
- Test with different parameter combinations and configurations
- Ensure scripts handle interruption gracefully (Ctrl+C)
