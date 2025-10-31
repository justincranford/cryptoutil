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

## CRITICAL: Git Operations

- **NEVER USE GitKraken MCP Server tools (mcp_gitkraken_*) in GitHub Copilot chat sessions**
- **ALWAYS use terminal git commands** (git status, git add, git commit, git push)
- **GitKraken is ONLY for manual GUI operations** - never automated in chat
- **All version control operations MUST use PowerShell git commands**
- **NEVER use python in chat sessions** - python is not installed in Windows PowerShell or Alpine container images; use PowerShell-native commands instead
- **NEVER use bash in chat sessions** - bash is not available in Windows PowerShell; use PowerShell syntax and commands instead
- **NEVER use powershell.exe in chat sessions** - powershell.exe is not needed when already in PowerShell; use native PowerShell commands instead
- **NEVER use -SkipCertificateCheck in PowerShell commands** - this parameter only exists in PowerShell 6+; use `[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}` for PowerShell 5.1

## Curl/Wget Command Usage Rules

**CRITICAL: Context-specific curl/wget restrictions**

### ❌ BANNED in Copilot Chat Sessions (Local Windows PowerShell)
- **NEVER use curl or wget in chat sessions** - not installed in Windows PowerShell
- **Alternative**: Use PowerShell `Invoke-WebRequest` or `docker compose exec <service> wget ...`
- **Why**: Local Windows dev environment doesn't have curl/wget in PATH
- **Exception**: Examples in documentation showing CI/CD usage are acceptable if clearly marked as workflow-only

### ✅ ALLOWED in GitHub Actions Workflows (.github/workflows/*.yml)
- **curl and wget are both available** in GitHub Actions Ubuntu runners and act containers
- **Preferred**: Use `curl -skf` for HTTPS with self-signed certificates (more reliable than wget)
- **Pattern**: `curl -skf --connect-timeout 10 --max-time 15 "$url" -o /tmp/response.json`
- **Why**: GitHub runners and act containers have both tools preinstalled

### ✅ ALLOWED in Docker Compose Healthchecks (compose.yml)
- **Preferred**: Use `wget` (available in Alpine/busybox containers)
- **Fallback**: Use `curl` if container includes it (e.g., grafana-otel-lgtm)
- **Pattern**: `test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]`
- **Why**: Most containers use Alpine base which includes wget via busybox
- **Exception**: Grafana-based containers may only have curl available

### Decision Tree for curl/wget Usage

```
Is this a Copilot chat command for local execution?
├─ YES → ❌ NEVER use curl/wget
│         ✅ Use: Invoke-WebRequest or docker compose exec
│
└─ NO → Is this a GitHub Actions workflow (.github/workflows/*.yml)?
         ├─ YES → ✅ Use curl or wget (curl preferred for HTTPS)
         │
         └─ NO → Is this a Docker Compose healthcheck?
                  ├─ YES → ✅ Use wget (preferred) or curl (fallback)
                  │         Check container base image:
                  │         - Alpine/busybox → wget available
                  │         - Grafana → curl available
                  │
                  └─ NO → Is this Dockerfile or container build context?
                           └─ YES → ✅ Use wget (Alpine base images have busybox wget)
```

### Summary Table

| Environment | curl | wget | Preferred | Alternative |
|-------------|------|------|-----------|-------------|
| Copilot Chat (Local PowerShell) | ❌ | ❌ | Invoke-WebRequest | docker compose exec |
| GitHub Workflows (ci-*.yml) | ✅ | ✅ | curl -skf | wget |
| Act Local Testing | ✅ | ✅ | curl -skf | wget |
| Compose Healthchecks (Alpine) | ❌ | ✅ | wget | External healthcheck |
| Compose Healthchecks (Grafana) | ✅ | ❌ | curl -f | External healthcheck |
| Dockerfile (Alpine base) | ❌ | ✅ | wget | apk add curl |
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
- **CRITICAL: cicd.go contains linting tools, cicd_test.go contains deliberate lint violations for testing purposes** - cicd.go implements linting functionality (go-enforce-any, go-enforce-test-patterns, etc.) and should be linted normally; cicd_test.go contains interface{} patterns for testing go-enforce-any functionality and other deliberate violations to validate cicd commands work correctly - these files MUST be excluded from directory walks in cicd commands (go-enforce-any, go-enforce-test-patterns) to prevent the tool from modifying its own test patterns or checking itself; all other validations (imports, dependencies, linting) should still be performed on both files
- **ALWAYS exclude internal/cmd/cicd/cicd.go and internal/cmd/cicd/cicd_test.go from directory walks** in cicd commands (go-enforce-any, go-enforce-test-patterns) to prevent the tool from modifying its own test patterns or checking itself
- **ALWAYS use UTF-8 without BOM for ALL text files** - this is enforced by cicd all-enforce-utf8 command; PowerShell's default UTF-16 LE encoding breaks Docker secrets and PostgreSQL initialization
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

## Git Workflow

### Commit and Push Strategy
- **Commit vs Push**: Commit frequently for logical units of work; push only when ready for CI/CD and peer review
- **Pre-push hooks**: Run automatically before push to enforce quality gates (dependency checks, linting)
- **Dependency management**: Update dependencies incrementally with test validation between updates
- **Branch strategy**: Use feature branches for development; merge to main only after CI passes
- **Commit hygiene**: Keep commits atomic and focused; use conventional commit messages
- **Push readiness**: Ensure all pre-commit checks pass before pushing; resolve any hook failures

### Conventional Commits
- Follow [Conventional Commits](https://www.conventionalcommits.org/) spec
- Format: `<type>[optional scope]: <description>`
- Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
- Use imperative mood, lowercase, no period at end
- For breaking changes: add `!` after type or use `BREAKING CHANGE:`
- Keep subject ≤72 chars; explain what/why in body if needed

## Terminal Command Auto-Approval

### Pattern Checking Workflow
When executing terminal commands through Copilot:
1. **Check Pattern Match**: Verify if command matches `chat.tools.terminal.autoApprove` patterns in `.vscode/settings.json`
2. **Track Unmatched**: Maintain list of unmatched commands during session
3. **End-of-Session Review**: Ask user if they'd like to add new auto-approve patterns
4. **Pattern Recommendations**:
   - **Auto-Enable (true)**: Safe, informational, build commands
   - **Auto-Disable (false)**: Destructive, dangerous, system-altering commands

### Auto-Enable Candidates
- Read-only operations (status, list, inspect, logs, history)
- Build and test commands (build, test, format, lint)
- Safe informational commands (version, info, df)
- Development workflow commands (fetch, status, diff)

### Auto-Disable Candidates
- Destructive operations (rm, delete, prune, reset, kill)
- Network operations (push, pull from remotes)
- System modifications (install, update, edit configurations)
- File system changes (create, update, delete files/directories)
- Container execution (exec, run interactive containers)

### Pattern Format
- Use established regex: `"/^command (sub1|sub2)/": true|false`
- Group related subcommands with alternation `(cmd1|cmd2|cmd3)`
- Use `^` for start anchor and appropriate word boundaries
- Include comments explaining security rationale

## TODO List Maintenance

### Critical Requirements
- **Delete completed tasks immediately** - don't mark as done, remove from file
- Review files for completed items before ending sessions
- Historical context belongs in commit messages, not TODO lists
- Always ensure files contain ONLY active, actionable tasks

### Implementation Guidelines
- **Large cleanups**: Use create_file to rewrite entire file with only active tasks
- **Failed replacements**: Create clean version in new file, then replace original
- **Avoid complex replace_string_in_file**: Large text blocks often fail due to whitespace mismatches
- **Validate compliance**: Check file contains only active tasks
