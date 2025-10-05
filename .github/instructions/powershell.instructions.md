---
description: "Instructions for PowerShell usage on Windows"
applyTo: "**"
---
# PowerShell Instructions

## Execution

- Use `powershell -NoProfile -ExecutionPolicy Bypass -File script.ps1` (process-scoped, no policy change)
- Alternative: `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass; .\script.ps1`
- Never change `-Scope LocalMachine` or `-Scope CurrentUser`

## Syntax

- Use `;` for chaining, `$env:VAR` for env vars
- Bash equivalents: `| Select-Object -First 10` (head), `| Select-String` (grep)
- Avoid emojis/Unicode in here-strings
- No `&&` or `||` in v5.1 — use `;` or `if ($LASTEXITCODE -eq 0) { cmd2 }`

## Common Errors

- Switch defaults: avoid `[switch]$Param = $true`
- String interpolation: use single quotes `'literal $var'` or here-strings `@' ... '@` for complex content
- Variable expansion: use `${variable}` for clarity
- File slicing: read to array first: `$lines = Get-Content file; $lines[10..20]`
