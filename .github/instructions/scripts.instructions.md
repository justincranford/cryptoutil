---
description: "Instructions for cross-platform script development"
applyTo: "scripts/**"
---
# Cross-Platform Script Instructions

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
