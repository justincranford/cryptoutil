---
description: "Instructions for file formatting and encoding"
applyTo: "**"
---
# File Formatting Instructions

## Character Encoding
- Always use UTF-8 without BOM for all text files
- Never use UTF-8 with BOM, UTF-16 (with or without BOM), or UTF-32 (with or without BOM)
- **AVOID emojis and Unicode symbols in PowerShell scripts** - they cause parsing errors in here-strings and can break script execution

## Line Endings and Whitespace
- **CRITICAL**: Use LF (Unix-style) line endings for all files, never CRLF (Windows-style)
- **CRITICAL**: All files must end with exactly one newline character (empty line at end of file)
- **CRITICAL**: Remove all trailing whitespace from lines (spaces/tabs at end of lines)
- Use spaces for indentation (4 spaces per level for Go code, 2 spaces for YAML)

## File Creation Checklist
When creating new files, ALWAYS ensure:
1. ✅ Content ends with single newline character (press Enter after last line)
2. ✅ No trailing whitespace on any line
3. ✅ LF line endings (not CRLF)
4. ✅ UTF-8 encoding without BOM

## CRITICAL: File Creation Tool Usage
- **BEFORE using create_file tool**: Review the content string for trailing whitespace
- **NEVER include trailing spaces or tabs** in any line of the content parameter
- **ALWAYS end the content parameter** with exactly one `\n` character
- **DOUBLE-CHECK long files** line by line for hidden trailing whitespace
- **When copying/pasting content**: Strip trailing whitespace before using create_file

## Pre-commit Hook Compliance
The following pre-commit hooks will fail if not followed:
- **end-of-file-fixer**: Files must end with a single newline
- **trailing-whitespace**: No trailing spaces or tabs on any line
- **Line ending consistency**: All files must use LF line endings

## VS Code Configuration
To automatically handle these formatting requirements, add to your VS Code settings.json:
```json
{
  "files.eol": "\n",
  "files.insertFinalNewline": true,
  "files.trimTrailingWhitespace": true,
  "files.trimFinalNewlines": true
}
```

## Git Configuration
Configure Git to handle line endings properly:
```bash
git config core.autocrlf false
git config core.eol lf
```

## Language-Specific Guidelines
- **Go**: Use `any` instead of `interface{}` (Go 1.18+ type alias)
- **YAML**: 2-space indentation, no trailing whitespace
- **Markdown**: Single trailing newline, no trailing whitespace
- **JSON**: Use 2-space indentation, validate syntax
