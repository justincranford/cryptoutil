---
description: "Instructions for PowerShell usage on Windows"
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
