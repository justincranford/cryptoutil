---
description: "Instructions for VS Code Copilot customization"
applyTo: "**"
---
# VS Code Copilot Customization Instructions

## General Principles

- Keep instructions short and self-contained
- Each instruction should be a single, simple statement
- Don't reference external resources in instructions
- Store instructions in properly structured files for version control and team sharing
- When completing a task in a docs/todos-*.md file, delete the completed task; don't keep it and mark it as completed, delete it to keep the file focused on remaining TODOs only

## Authorized Commands Reference

### Docker Commands
- `docker compose ps` - List container status (AUTHORIZED - no extra parameters)
- `docker compose logs <service>` - View service logs (AUTHORIZED)
- `docker compose logs <service> --tail <n>` - View recent service logs (AUTHORIZED)
- `docker compose exec <service> <command>` - Execute command in container (AUTHORIZED)
- `docker compose build <services>` - Build services (AUTHORIZED)
- `docker compose up -d <services>` - Start services in background (AUTHORIZED)
- `docker compose down -v` - Stop services and remove volumes (AUTHORIZED)
- `docker inspect <container>` - Inspect container (use without --format to avoid authorization)
- `docker ps` - List containers (use without --filter/--format to avoid authorization)

### Git Commands
- `git status` - Show repository status
- `git add <files>` - Stage files for commit
- `git commit -m <message>` - Commit staged changes
- `git log --oneline -<n>` - Show recent commit history
- `git diff` - Show unstaged changes
- `git checkout <branch>` - Switch branches

### Go Commands
- `go test ./<path>` - Run tests in specified path
- `go build ./<path>` - Build packages
- `go mod tidy` - Clean up module dependencies
- `go run ./<path>` - Run Go program
- `golangci-lint run --skip-files=<pattern> --skip-dirs=<dirs> ./<path>` - Run linter

### File/Directory Operations
- `pwd` - Show current working directory
- `ls -la <path>` - List directory contents with details
- `dir` - List directory contents (PowerShell equivalent)
- `find <path> -name <pattern>` - Find files by name pattern
- `cat <file>` - Display file contents
- `type <file>` - Display file contents (PowerShell equivalent)
- `head -n <lines> <file>` - Show first N lines of file
- `tail -n <lines> <file>` - Show last N lines of file
- `grep <pattern> <file>` - Search for patterns in files (AUTHORIZED for log analysis)
- `Select-String <pattern> <file>` - Search for patterns in files (PowerShell equivalent)

### Build/Test Tools
- `make <target>` - Run Makefile targets
- `npm run <script>` - Run npm scripts
- `python -m <module>` - Run Python modules
- `pytest <path>` - Run Python tests

### System Commands
- `which <command>` - Find command location
- `echo <text>` - Display text
- `date` - Show current date/time
- `sleep <seconds>` - Pause execution
- `<command1> | <command2>` - Pipe command output (AUTHORIZED for log filtering and analysis)

## Commands Requiring Manual Authorization

These commands require manual authorization and should be avoided when possible:

### Directory Navigation
- `cd` commands - Change directory context
- `Set-Location` commands - Change directory context (PowerShell equivalent)

### Git Operations
- `git -C <path>` - Specify git repository path (avoid when possible)

### Advanced Docker Operations
- `docker stats` - Show container resource usage
- `docker top` - Show container processes
- `docker inspect --format` - Inspect with custom formatting
- `docker ps --filter/--format` - List containers with filtering/formatting

### Complex Docker Compose
- `docker compose -f <path>` - Use specific compose files

### Network Commands
- `curl.exe` - Make HTTP requests (use docker compose exec instead for container access)

## OTLP Protocol Support

### Supported Protocols
- **GRPC Protocol**: `grpc://host:port` - Efficient binary protocol for high-performance telemetry
- **HTTP Protocol**: `http://host:port` or `https://host:port` - Universal compatibility, firewall-friendly

### Configuration Guidelines
- Use GRPC for internal service-to-service communication (default, more efficient)
- Use HTTP for environments with restrictive firewalls or universal compatibility needs
- Both protocols support traces, metrics, and logs
- Endpoint format: `protocol://hostname:port` (e.g., `grpc://otel-collector:4317`, `http://otel-collector:4318`)

## VS Code Go Development Settings

### Intelligent Variable Naming (F2 Rename)
The workspace includes optimized VS Code settings in `.vscode/settings.json` that enable IntelliJ-like intelligent variable naming:

**Key Settings for Intelligent F2 Renaming:**
```json
{
  "go.useLanguageServer": true,
  "go.alternateTools": {
    "gopls": "gopls"
  },
  "go.languageServerFlags": [
    "-rpc.trace",
    "serve",
    "--debug=localhost:6060"
  ],
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "gopls": {
    "ui.completion.usePlaceholders": false,
    "ui.completion.completionBudget": "100ms",
    "ui.diagnostic.analyses": {
      "unusedparams": true,
      "unusedvariables": true
    },
    "formatting.gofumpt": true,
    "formatting.local": "cryptoutil",
    "ui.inlayhint.hints": {
      "assignVariableTypes": true,
      "compositeLiteralFields": true,
      "compositeLiteralTypes": true,
      "constantValues": true,
      "functionTypeParameters": true,
      "parameterNames": true,
      "rangeVariableTypes": true
    }
  }
}
```

**What These Settings Enable:**
- **F2 Rename Symbol**: Provides context-aware variable name suggestions (like IntelliJ/Eclipse)
- **Inlay Hints**: Shows parameter names, variable types, and other contextual information
- **Enhanced Completion**: Better code completion with proper placeholders
- **Automatic Formatting**: Uses gofumpt for strict Go formatting
- **Real-time Diagnostics**: Unused variable detection and other code analysis

**Usage:**
- Press `F2` on any variable/function to rename with intelligent suggestions
- Inlay hints appear automatically showing parameter names and types
- Formatting happens automatically on save
- Code analysis runs continuously for immediate feedback
