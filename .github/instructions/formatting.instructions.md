---
description: "Instructions for file formatting and encoding"
applyTo: "**"
---
# Formatting Instructions

- Always use UTF-8 without BOM, single newline at EOF, no trailing whitespace
- When creating files: UTF-8 without BOM, single newline at EOF, no trailing whitespace
- Indent: 4 spaces (Go), 2 spaces (YAML, JSON, Markdown, Dockerfile)
- Avoid emojis/unicode in PowerShell scripts
- Use VS Code settings and Git config for auto-formatting (see README)
- Language specifics: Go—use `any` not `interface{}
