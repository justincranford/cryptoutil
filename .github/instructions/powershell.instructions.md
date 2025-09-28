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

- Use context clues to determine shell type before generating terminal commands
- Use PowerShell syntax for Windows terminal commands, not Bash
- Command chaining: use `;` (not `&&`)
- File paths: use `\` backslashes (not `/` forward slashes)
- Variables: use `$env:VARIABLE` (not `$VARIABLE`)
- Pipe to head: use `| Select-Object -First 10` (not `| head -10`)
- Grep equivalent: use `| Select-String "pattern"` (not `| grep "pattern"`)
- **ALWAYS use execution policy bypass for PowerShell scripts**: `powershell -ExecutionPolicy Bypass -File script.ps1` (never run scripts directly with `.\script.ps1`)
- **NEVER use emojis or Unicode symbols in PowerShell scripts** - they cause parsing errors in here-strings and break script execution
