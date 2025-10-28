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
- **NEVER USE GitKraken in GitHub Copilot chat** - GitKraken is a GUI tool for Git operations; use terminal git commands instead for all version control operations
- **NEVER use curl in chat sessions** - curl is not installed in Windows PowerShell or Alpine container images; use PowerShell Invoke-WebRequest or docker compose exec instead
- **NEVER use -SkipCertificateCheck in PowerShell commands** - this parameter only exists in PowerShell 6+; use `[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}` for PowerShell 5.1
- **ALWAYS use HTTPS 127.0.0.1:9090 for admin APIs** (/shutdown, /livez, /readyz) - these are private server endpoints, not public server endpoints
- **ALWAYS rely on golangci-lint exclusions defined in .golangci-lint.yml** - never use --skip-files or --skip-dirs command line flags
- **ALWAYS run Go fuzz tests from project root** - never use `cd` commands before `go test -fuzz` (causes module detection failures)
- **ALWAYS use PowerShell `;` for command chaining** - never use bash `&&` syntax (PowerShell 5.1 doesn't support it)
- **STOP MODIFYING THE DOCKER COMPOSE SECRETS** - Docker Compose secrets in `deployments/compose/` are carefully configured for cryptographic interoperability; NEVER create, modify, or delete secret files as this breaks the cryptographic key hierarchy and causes test failures
- **PREFER SWITCH STATEMENTS WHERE POSSIBLE** - Use `switch variable { case value: ... }` over `if/else if` chains for cleaner, more maintainable code; when switch is not possible, prefer `if/elseif/else` pattern over separate `if` or `if/else` statements
- **ALWAYS declare values as constants near the top of the file** to proactively mitigate "mnd" (magic number detector) linter errors in .golangci-lint.yml:
  - **HTTP status codes**: `http.StatusOK`, `http.StatusNotFound`, `http.StatusForbidden` instead of `200`, `404`, `403`
  - **Durations**: `timeout = 30 * time.Second`, `delay = 100 * time.Millisecond` instead of inline `30000`, `100`
  - **Special strings**: `statusPass = "PASS"`, `statusFail = "FAIL"`, `trendNoChange = "→ No change"` instead of inline strings
  - **Special numbers**: `rsaKeySize2048 = 2048`, `aes256KeySize = 256`, `bufferSize1KB = 1024` instead of magic numbers
  - **Pool sizes**: `poolMin = 3`, `poolMax = 9` instead of inline numbers in pool configurations
  - **Port numbers**: `defaultPort = 8080`, `adminPort = 9090` instead of `8080`, `9090`
  - **Common counts/limits**: `maxRetries = 3`, `minArgs = 2` instead of `3`, `2`
  - **Percentage values**: `tolerance5Percent = 0.05`, `halfValue = 0.5` instead of `0.05`, `0.5`
  - **File permissions**: `fileMode = 0o600`, `dirMode = 0o755` instead of `0o600`, `0o755`
- **ALWAYS use `golangci-lint run --fix` FIRST before any manual lint error fixing** - this ensures all auto-fixable issues are resolved automatically before attempting manual fixes
- **ALWAYS use `golangci-lint run --fix` for auto-fixing** - this single command handles formatting (gofumpt), imports (goimports), and all other auto-fixable linters in one pass
- **Pre-commit hooks automatically run `golangci-lint run --fix`** - no need to run standalone gofumpt or goimports commands separately
- **CRITICAL: gofumpt is ALWAYS preferred over gofmt in ALL SITUATIONS** - gofumpt is a stricter superset of gofmt with additional formatting rules; NEVER use `gofmt` directly, even if you see a "gofmt" lint error from any tool including golangci-lint; always use `golangci-lint run --fix` which applies gofumpt formatting
- **CRITICAL: cicd.go contains linting tools, cicd_test.go contains deliberate lint violations for testing purposes** - cicd.go implements linting functionality (gofumpter, enforce-test-patterns, etc.) and should be linted normally; cicd_test.go contains interface{} patterns for testing gofumpter functionality and other deliberate violations to validate cicd commands work correctly - these files MUST be excluded from directory walks in cicd commands (gofumpter, enforce-test-patterns) to prevent the tool from modifying its own test patterns or checking itself; all other validations (imports, dependencies, linting) should still be performed on both files
- **ALWAYS exclude internal/cicd/cicd.go and internal/cicd/cicd_test.go from directory walks** in cicd commands (gofumpter, enforce-test-patterns) to prevent the tool from modifying its own test patterns or checking itself
- **ALWAYS use UTF-8 without BOM for ALL text files** - this is enforced by cicd enforce-file-encoding command; PowerShell's default UTF-16 LE encoding breaks Docker secrets and PostgreSQL initialization
- **PowerShell file creation**: `$utf8NoBom = New-Object System.Text.UTF8Encoding $false; [System.IO.File]::WriteAllText("file.txt", "content", $utf8NoBom)` (creates UTF-8 without BOM)
- **NEVER use `Out-File`, `Set-Content`, or `>` redirection** for text files - they default to UTF-16 LE with BOM; use `[System.IO.File]::WriteAllText()` with UTF8Encoding(false) instead

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
- `golangci-lint run` - Run linter (relies on exclusions in .golangci-lint.yml)
- `go test -fuzz=. -fuzztime=5s ./<path>` - Run fuzz tests for 5 seconds (ALWAYS run from project root, NEVER use cd commands)
- `go test -run=FuzzXXX -fuzz=FuzzXXX -fuzztime=5s ./<path>` - Run specific fuzz test (use PowerShell `;` not bash `&&` for command chaining)

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
- `./mvnw <goals>` - Run Maven goals with Maven Wrapper (for Java Gatling tests)
- `mvnw.cmd <goals>` - Run Maven goals with Maven Wrapper on Windows

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
- `tee` - Output redirection to files (E2E tests already log automatically)

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

## Fuzz Testing Guidelines

### CRITICAL: Unique Fuzz Test Naming Convention
- **ALL Fuzz* test function names MUST be unique and MUST NOT be substrings of any other fuzz test names**
- This ensures cross-platform compatibility without quotes or regex in `-fuzz` parameters
- **Example Problem**: `FuzzHKDF` conflicts with `FuzzHKDFwithSHA256` (substring match)
- **Solution**: Use `FuzzHKDFAllVariants` instead of `FuzzHKDF`
- **Why**: Go's `-fuzz` parameter does partial matching; unique names eliminate ambiguity
- **Cross-Platform**: Unquoted parameters work identically on Windows and Linux: `go test -fuzz=FuzzXXX`

### Common Mistakes to Avoid
- **NEVER do this**: `cd internal/common/crypto/keygen; go test -fuzz=.` (causes "go.mod file not found" errors)
- **NEVER do this**: `go test -fuzz=. && other-command` (PowerShell 5.1 doesn't support `&&`)
- **NEVER do this**: Running fuzz tests from subdirectories (breaks Go module detection)
- **NEVER do this**: Using quotes or regex when test names are unique: `-fuzz="^FuzzXXX$"` (causes cross-platform issues)
- **NEVER do this**: Creating fuzz test names that are substrings of other fuzz test names

### Correct Fuzz Test Execution
- **ALWAYS do this**: Run from project root: `go test -fuzz=FuzzSpecificTest -fuzztime=5s ./internal/common/crypto/keygen`
- **ALWAYS do this**: Use PowerShell `;` for chaining: `go test -fuzz=FuzzXXX -fuzztime=5s ./path; echo "Done"`
- **ALWAYS do this**: Specify full package paths: `./internal/common/crypto/digests`
- **ALWAYS do this**: Use unquoted, unique test names: `-fuzz=FuzzGenerateRSAKeyPair` (no quotes needed)

### Fuzz Test Patterns
- **Specific fuzz test**: `go test -fuzz=FuzzXXX -fuzztime=5s ./<package>` (most common, no quotes)
- **All fuzz tests in package**: `go test -fuzz=. -fuzztime=5s ./<package>` (only if package has 1 fuzz test)
- **Quick verification**: Use `-fuzztime=5s` for fast feedback during development
- **Cross-platform compatibility**: Avoid quotes and regex; ensure unique test names instead
