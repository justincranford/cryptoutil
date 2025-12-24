# Cross-Platform Tooling and Scripts - Complete Specifications

**Version**: 1.0
**Last Updated**: 2025-12-24
**Referenced by**: `.github/instructions/05-01.cross-platform.instructions.md`

## autoapprove Wrapper - Security Tool

### Purpose

**Bypasses VS Code Copilot's hardcoded safety blockers for loopback network commands**

**Problem**: VS Code Copilot blocks HTTPS curl commands to 127.0.0.1/localhost by default, even for local development servers

**Solution**: `autoapprove` wrapper allows loopback-only commands after security validation

### Usage Pattern (Local Chat Agent Sessions Only)

**Allowed Commands**:
```bash
# Health checks
autoapprove curl https://127.0.0.1:9090/admin/v1/livez
autoapprove curl https://localhost:8080/ui/swagger/doc.json

# Testing
autoapprove go test ./...
autoapprove golangci-lint run

# Database queries
autoapprove psql -h 127.0.0.1 -U postgres -d cryptoutil

# Docker commands
autoapprove docker exec cryptoutil-sqlite wget https://127.0.0.1:8080/api
```

###Security Restrictions

**ONLY allows loopback addresses**:
- ✅ `127.0.0.1` (IPv4 loopback)
- ✅ `::1` (IPv6 loopback)
- ✅ `localhost` (resolves to loopback)
- ❌ External IPs, hostnames, or public URLs

**Logging**:
- Creates timestamped directories in `./test-output/autoapprove/`
- Logs all executed commands with timestamps
- Preserves command output for debugging
- Audit trail for security review

**Pattern**:
```
./test-output/autoapprove/
  2025-12-24-143022-curl-127-0-0-1-9090/
    command.txt     # Full command executed
    stdout.log      # Command stdout
    stderr.log      # Command stderr
    exit_code.txt   # Exit code
```

---

## HTTP Commands by Environment - Decision Matrix

| Environment | Tool | Example | Rationale |
|-------------|------|---------|-----------|
| Local Chat (VS Code) | `autoapprove curl` | `autoapprove curl -s http://127.0.0.1:8080/api` | Bypasses Copilot safety blockers |
| GitHub Actions | `curl` directly | `curl -s http://127.0.0.1:8080/api` | No safety blockers in CI/CD |
| Docker healthchecks | `wget` | `wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/v1/livez` | Available in Alpine, lighter than curl |

**NEVER use `Invoke-WebRequest`**:
- ❌ PowerShell cmdlet blocked by VS Code safety overrides
- ❌ Slower than curl/wget
- ❌ More verbose syntax
- ✅ Use `autoapprove curl` instead

---

## Cross-Platform Script Language Preference

### MANDATORY Preference Order

**1. Go (Highest Preference)**:
- ✅ Cross-platform (Windows, Linux, macOS) out-of-box
- ✅ Static compilation (single binary, no runtime dependencies)
- ✅ Complex logic support (control flow, error handling, concurrency)
- ✅ Performance-critical operations (fast execution)
- ✅ Integration with Go codebase (import internal packages)

**Use Cases**:
- CI/CD utilities (`cmd/cicd/*`)
- Code generation tools
- Workflow automation
- Integration testing utilities
- Build pipeline scripts

**2. Python (Medium Preference)**:
- ✅ When Go not suitable (rapid prototyping, one-off scripts)
- ✅ Python ecosystem integration (libraries not available in Go)
- ✅ Data processing (CSV, JSON, XML parsing)
- ✅ System administration tasks

**Use Cases**:
- Quick prototypes
- Data analysis scripts
- One-off migrations
- Python-specific tooling (pyproject.toml management)

**3. PowerShell/Bash (BANNED for New Scripts)**:
- ❌ NEVER create new PowerShell scripts (Windows-only)
- ❌ NEVER create new Bash scripts (Unix-only)
- ❌ Existing scripts are LEGACY ONLY (do not extend)
- ✅ Convert to Go when modifying existing shell scripts

**Why Banned**:
- Platform-specific (breaks cross-platform compatibility)
- Shell syntax differences (PowerShell vs Bash)
- Hard to test and maintain
- Limited error handling
- No type safety

---

## Authorized Chat Session Commands

### Git Commands (Auto-Approved)

**Read-Only Operations**:
- `git status` - Check working tree state
- `git log --oneline` - View commit history
- `git diff` - View uncommitted changes
- `git diff HEAD~1` - Compare with previous commit

**Staging and Committing**:
- `git add -A` - Stage all changes
- `git commit -m "message"` - Commit with message
- `git commit --no-verify -m "message"` - Skip pre-commit hooks (local iteration)

**File Operations**:
- `git checkout <branch>` - Switch branches
- `git checkout <hash> -- <file>` - Restore file from commit
- `git mv <old> <new>` - Rename files

### Go Commands (Auto-Approved)

**Testing**:
- `go test ./...` - Run all tests
- `go test ./pkg -v` - Run specific package tests
- `go test -race ./...` - Race detector
- `go test -fuzz=FuzzFunc` - Fuzz testing

**Building**:
- `go build ./...` - Build all packages
- `go build -o bin/app ./cmd/app` - Build specific binary

**Maintenance**:
- `go mod tidy` - Clean dependencies
- `go run ./cmd/tool` - Run tool directly
- `golangci-lint run` - Lint codebase
- `golangci-lint run --fix` - Auto-fix issues

### Docker Commands (Auto-Approved)

**Container Management**:
- `docker compose ps` - List containers
- `docker compose logs <service>` - View logs
- `docker compose exec <service> <cmd>` - Execute in container
- `docker compose build` - Build images
- `docker compose up -d` - Start services
- `docker compose down` - Stop services

**Inspection**:
- `docker inspect <container>` - Inspect container
- `docker ps` - List running containers
- `docker stats` - Container resource usage

### File Operations (Auto-Approved)

**Navigation and Listing**:
- `pwd` - Print working directory
- `ls -la` - List files (Unix)
- `dir` - List files (Windows)

**File Content**:
- `cat <file>` - Display file (Unix)
- `type <file>` - Display file (Windows)
- `head -n 20 <file>` - First 20 lines
- `tail -n 20 <file>` - Last 20 lines

**Search**:
- `grep "pattern" <file>` - Search file (Unix)
- `Get-Content <file> | Select-String "pattern"` - Search file (PowerShell)

**Directory Creation**:
- `mkdir -p path/to/dir` - Create directories (Unix)
- `New-Item -ItemType Directory -Path "path\to\dir"` - Create directories (PowerShell)

### Commands Requiring Manual Authorization

**Directory Changes**:
- `cd <path>` - Change directory (Unix)
- `Set-Location <path>` - Change directory (PowerShell)

**Network Commands** (without autoapprove):
- `curl <url>` - HTTP request (requires autoapprove for loopback)
- `wget <url>` - HTTP download (requires autoapprove for loopback)
- `Invoke-WebRequest <url>` - BANNED (use autoapprove curl instead)

---

## Docker Image Pre-Pull Action

### Purpose

**Parallelize image downloads to reduce workflow startup time**

**Problem**: Sequential image pulls in workflows slow down startup (5-10 minutes for multiple images)

**Solution**: GitHub Actions composite action `.github/actions/docker-images-pull` downloads images in parallel

### Usage Pattern

```yaml
- name: Pre-pull Docker images
  uses: ./.github/actions/docker-images-pull
  with:
    images: |
      postgres:18
      alpine:3.19
      golang:1.25.5-alpine
      otel/opentelemetry-collector-contrib:latest
```

**Benefits**:
- **Parallelization**: All images download simultaneously
- **Time Savings**: 5-10 minute sequential pulls → 1-2 minute parallel pulls
- **Caching**: GitHub Actions cache layer speeds up subsequent runs
- **Reliability**: Retries on transient network failures

**When to Use**:
- Workflows that use multiple Docker images
- E2E tests requiring full stack (app + postgres + otel + grafana)
- Matrix workflows with different Go versions

---

## PowerShell Syntax Notes

### Command Chaining

**Use semicolon (`;`) to chain commands** (NOT `&&`):

```powershell
# ✅ CORRECT (PowerShell)
git add -A; git commit -m "message"; git status

# ❌ WRONG (Bash syntax doesn't work in PowerShell)
git add -A && git commit -m "message" && git status
```

**Why**: PowerShell uses `;` for command chaining, `&&` is Bash-specific

### Unix Utilities Not Available

**PowerShell does not include Unix utilities by default**:
- ❌ `sed` - Stream editor (NOT available)
- ❌ `awk` - Text processing (NOT available)
- ❌ `grep` - Use `Select-String` instead

**Alternatives**:
- ✅ `git diff -- <file>` - View file diffs (cross-platform)
- ✅ `git show <commit> -- <file>` - View file at commit (cross-platform)
- ✅ `git grep 'pattern'` - Search repository content (cross-platform)
- ✅ `Get-Content file | Select-String 'pattern'` - Grep-like search (PowerShell)

**Recommendation**: Prefer Git built-in commands over Unix utilities for cross-platform compatibility

---

## Cross-References

**Related Documentation**:
- Git workflow: `.specify/memory/git.md`
- Docker configuration: `.specify/memory/docker.md`
- Security: `.specify/memory/security.md`
- GitHub workflows: `.specify/memory/github.md`
