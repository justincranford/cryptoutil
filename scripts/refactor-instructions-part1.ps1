# Complete Instruction Files Reorganization Script
# This script creates refactored versions of all instruction files

$utf8NoBom = New-Object System.Text.UTF8Encoding $false

Write-Host "===== INSTRUCTION FILES REORGANIZATION =====" -ForegroundColor Cyan
Write-Host ""

# Define all refactored content
$files = @{
    "01-01.copilot-customization.instructions.md" = @"
---
description: "Instructions for VS Code Copilot customization and critical restrictions"
applyTo: "**"
---
# VS Code Copilot Customization Instructions

## General Principles

- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing

## CRITICAL: Tool and Command Restrictions

### Git Operations
- **ALWAYS use terminal git commands** (git status, git add, git commit, git push)
- **NEVER USE GitKraken MCP Server tools (mcp_gitkraken_*)** in GitHub Copilot chat sessions
- **GitKraken is ONLY for manual GUI operations** - never automated in chat

### Language/Shell Restrictions in Chat Sessions
- **NEVER use python** - not installed in Windows PowerShell or Alpine container images
- **NEVER use bash** - not available in Windows PowerShell
- **NEVER use powershell.exe** - not needed when already in PowerShell
- **NEVER use -SkipCertificateCheck** in PowerShell commands - only exists in PowerShell 6+
  - Alternative for PS 5.1: ``[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {`$true}``

## Critical Project Rules

- **ALWAYS use HTTPS 127.0.0.1:9090 for admin APIs** (/shutdown, /livez, /readyz)
- **ALWAYS run Go fuzz tests from project root** - never use `cd` commands before `go test -fuzz`
- **ALWAYS use PowerShell `;` for command chaining** - never use bash `&&` syntax (PS 5.1 doesn't support it)
- **STOP MODIFYING DOCKER COMPOSE SECRETS** - carefully configured for security; never create, modify, or delete
- **PREFER SWITCH STATEMENTS** over `if/else if/else` chains for cleaner code

## VS Code Integration

- See `.vscode/settings.json` for Go extension configuration
- Press `F2` on variables/functions for intelligent, context-aware rename suggestions
- Inlay hints show parameter names and types for better context
"@

    "02-01.coding.instructions.md" = @"
---
description: "Instructions for coding patterns and standards"
applyTo: "**"
---
# Coding Instructions

## Code Patterns

### Default Values
- **ALWAYS declare default values as named variables** rather than inline literals
- Example: ``var defaultConfigFiles = []string{}``
- Follows established pattern in config.go

### Pass-through Calls
- **Prefer same parameter and return value order** as helper functions
- Maintains API consistency and reduces confusion

## Conditional Statement Chaining

### CRITICAL: Pattern for Mutually Exclusive Conditions

**ALWAYS prefer chained if/else if/else for mutually exclusive conditions:**
``````go
if ctx == nil {
    return nil, fmt.Errorf("context cannot be nil")
} else if logger == nil {
    return nil, fmt.Errorf("logger cannot be nil")
} else if description == "" {
    return nil, fmt.Errorf("description cannot be empty")
}
``````

**Avoid separate if statements for mutually exclusive conditions:**
``````go
// DON'T DO THIS for mutually exclusive conditions
if ctx == nil {
    return nil, fmt.Errorf("context cannot be nil")
}
if logger == nil {
    return nil, fmt.Errorf("logger cannot be nil")
}
``````

### When NOT to Chain
- Independent conditions (not mutually exclusive)
- Error accumulation patterns
- Cases with early returns that don't overlap

## Switch Statements

- **PREFER switch statements** over `if/else if/else` chains when possible
- Pattern: ``switch variable { case value: ... }``
- When switch not possible, prefer `if/else if/else` over separate `if` statements
"@
}

# Write each file
foreach ($fileName in $files.Keys) {
    $filePath = "c:\Dev\Projects\cryptoutil\.github\instructions\$fileName"
    Write-Host "Refactoring: $fileName" -ForegroundColor Yellow
    [System.IO.File]::WriteAllText($filePath, $files[$fileName], $utf8NoBom)
}

Write-Host ""
Write-Host "===== REORGANIZATION COMPLETE =====" -ForegroundColor Green
Write-Host "Files have been refactored and saved with UTF-8 (no BOM) encoding"
