---
description: "Instructions for file formatting and encoding"
applyTo: "**"
---
# Formatting Instructions

- Always use UTF-8 (no BOM), LF endings, single newline at EOF, no trailing whitespace
- When creating files: check for trailing whitespace, LF endings, single newline, UTF-8
- Indent: 4 spaces (Go), 2 spaces (YAML, JSON, Markdown, Dockerfile)
- Avoid emojis/unicode in PowerShell scripts
- Use VS Code settings and Git config for auto-formatting (see README)
- Pre-commit hooks enforce: single newline, no trailing whitespace, LF endings
- Language specifics: Go—use `any` not `interface{}
