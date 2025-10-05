---
description: "Instructions for PowerShell usage on Windows"
applyTo: "**"
---
# PowerShell Instructions

## Execution policy: preferred invocation

- ALWAYS prefer a one-shot, process-scoped bypass when running bundled helper scripts. This avoids permanently weakening machine policies and works reliably on systems where script execution is restricted.

```powershell
# Recommended (one-shot, no persistent policy change)
powershell -NoProfile -ExecutionPolicy Bypass -File script.ps1 -ScanProfile quick -Timeout 900
```

- Alternative (session-scoped, safe for interactive runs):

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 900
```

Notes:
- The first form launches a new PowerShell process with ExecutionPolicy bypassed for that process only. It's the safest and most repeatable approach for automation and CI helpers.
- Avoid changing `-Scope LocalMachine` or `-Scope CurrentUser` unless you understand the security implications.

## Scripting best-practices

- Use PowerShell syntax for Windows terminal commands (not Bash).
- Use `;` for chaining, `\` for paths, and `$env:VAR` for environment variables.
- Use `| Select-Object -First 10` for head-like behavior and `| Select-String` for grep-like searches.
- Avoid emojis or complex Unicode in here-strings — they can cause parsing or encoding issues.

## Common mistakes to avoid

- Switch parameter defaults: avoid `[switch]$All = $true`; prefer explicit logic to set defaults.
- Here-strings: avoid complex Unicode characters inside `@"..."@`.
- Variable expansion in paths: use `${variable}` for clarity, e.g. `"${PWD}\${OutputDir}"`.
- Prefer here-strings over complex backtick concatenation for multi-line text.
- Validate script parameters (test help and parameter validation) before use.
- Use proper error handling and exit codes in scripts.
---
descript- **ALWAYS use execution policy bypass for PowerShell scripts**: `powershell -ExecutionPolicy Bypass -File script.ps1` (never run scripts directly with `.\script.ps1`)
- **NEVER use emojis or Unicode symbols in PowerShell scripts** - they cause parsing errors in here-strings and break script execution

## PowerShell Scripting Common Mistakes to Avoid

- **Switch parameter defaults**: Never set `[switch]$All = $true` - use logic to determine defaults instead
- **Here-string content**: Avoid complex Unicode characters in here-strings (`@"..."@`) - they cause parsing errors
- **Variable expansion in paths**: Use `${variable}` syntax for complex variable expansion (e.g., `"${PWD}/${OutputDir}"`)
- **String concatenation**: Prefer here-strings over complex string concatenation with backticks for multi-line content
- **Parameter validation**: Always test script parameters, especially help functionality, before completion
- **Error handling**: Use proper PowerShell error handling patterns and exit codesn: "Instructions for PowerShell usage on Windows"
applyTo: "**"
---
# PowerShell Instructions

- Use PowerShell syntax for Windows terminal commands (not Bash)
- Use `;` for chaining, `\` for paths, `$env:VAR` for env vars
- Use `| Select-Object -First 10` for head, `| Select-String` for grep
- Always use execution policy bypass: `powershell -ExecutionPolicy Bypass -File script.ps1`
- Never use emojis/unicode in scripts (see formatting instructions)
- Avoid: `[switch]$All = $true`, complex Unicode in here-strings, improper variable expansion, string concat with backticks, untested params, poor error handling
