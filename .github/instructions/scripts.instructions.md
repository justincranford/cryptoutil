---
description: "Instructions for cross-platform script development"
applyTo: "scripts/**"
---
# Cross-Platform Script Instructions

## Language Preferences

**ALWAYS prefer Go for new scripts when possible**

### When to Use Go for Scripts:
- Cross-platform compatibility needed (Windows, Linux, macOS)
- Complex logic or data processing required
- Integration with existing Go codebase
- Performance-critical operations
- Need for static compilation and distribution
- Structured error handling and testing required

### When PowerShell/Bash May Still Be Appropriate:
- Windows-specific system administration tasks
- Quick automation of platform-specific workflows
- Integration with platform-specific APIs
- Existing ecosystem dependencies
- Simple file operations or environment setup

## When to Use Scripts
- Use scripts for complex, multi-step development workflows (lint, test, build, mutation testing, etc.)
- Scripts provide consistent cross-platform developer experience with parameter handling
- Note: For Docker Compose, prefer direct command directives over separate script files

## Script Development Guidelines

### Go Scripts
- Place scripts in `scripts/` directory with appropriate subdirectories
- Use `cmd/` subdirectory for main script entry points
- Follow standard Go project layout within script directories
- Use Go modules for dependency management
- Implement proper error handling with context
- Use structured logging (slog package)
- Provide CLI interface with flag package
- Include comprehensive README files
- Support testing with Go's testing package

### PowerShell/Bash Scripts
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
- Test both PowerShell and Bash versions on their respective platforms (or Go binaries on all platforms)
- For PowerShell scripts, test with execution policy restrictions:
  - Test with `powershell -ExecutionPolicy Bypass -File script.ps1`
  - Include `#Requires -Version 5.1` directive for minimum version
  - Document execution policy requirements in script comments
- Test help/usage functions: `script.ps1 -Help` and `script.sh --help`
- Test error conditions and edge cases (missing dependencies, invalid parameters)
- Verify script cleanup and resource management (process termination, file cleanup)
- Test with different parameter combinations and configurations
- Ensure scripts handle interruption gracefully (Ctrl+C)
