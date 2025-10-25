# Pre-commit Hooks Pipeline

This document describes the design intent, optimal ordering, and detailed configuration of the cryptoutil project's pre-commit hooks pipeline.

## Design Intent

The pre-commit hooks pipeline is designed with a **fail-fast, progressive validation** philosophy that optimizes for both developer experience and CI/CD performance.

### Core Principles

1. **Fail Fast**: Quick, inexpensive checks run first to catch obvious issues early
2. **Progressive Validation**: Each stage builds on the previous, ensuring code quality accumulates
3. **Auto-Fix First**: Formatting and import fixes happen before validation to reduce noise
4. **Dependency Management**: Module dependencies are cleaned before expensive linting operations
5. **Build Validation**: Code compilation is verified after linting passes
6. **Custom Checks**: Project-specific rules run after basic validation
7. **Specialized Linting**: File-type specific checks run last for comprehensive coverage

### Performance Optimization

The ordering minimizes redundant work and maximizes parallel processing potential:
- Auto-fixing tools run before their validation counterparts
- Expensive operations (golangci-lint) run after cheap fixes
- Related tools are grouped to share repository contexts
- File-type specific tools are logically sequenced

## Current Pipeline Ordering

### High-Level Flow

```
1. Generic File Checks → 2. Go Auto-Fixes → 3. Dependency Mgmt → 4. Go Validation → 5. Build Check → 6. Custom Rules → 7. Specialized Linting → 8. Commit Validation
```

### Detailed Stage Breakdown

| Stage | Purpose | Tools | Rationale |
|-------|---------|-------|-----------|
| **1. Generic File Checks** | Universal file validation | pre-commit-hooks | Fast, catches basic issues first |
| **2. Go Auto-Fixes** | Code formatting and imports | gofumpt, goimports | Fix before validate to reduce noise |
| **3. Dependency Management** | Module cleanup | go mod tidy | Clean state before expensive linting |
| **4. Go Validation** | Comprehensive Go linting | golangci-lint | Full analysis after fixes applied |
| **5. Build Validation** | Compilation verification | go build | Ensure code compiles after linting |
| **6. Custom Rules** | Project-specific checks | test-patterns | Business logic validation |
| **7. Specialized Linting** | File-type specific checks | actionlint, hadolint, shellcheck, bandit | Targeted validation by file type |
| **8. Commit Validation** | Message format checking | commitizen | Final gate before push |

## Tool Details and Configuration

### 1. Generic File Checks (pre-commit-hooks)

**Purpose**: Fast, universal validation of file formatting and basic syntax.

**Tools Included**:
- `end-of-file-fixer`: Ensures files end with newline
- `trailing-whitespace`: Removes trailing whitespace
- `fix-byte-order-marker`: Removes UTF-8 BOM
- `check-yaml`: Validates YAML syntax
- `check-json`: Validates JSON syntax (excludes VS Code settings)
- `check-added-large-files`: Prevents large file commits
- `check-merge-conflict`: Detects merge conflict markers
- `detect-aws-credentials`: Security check for AWS keys
- `detect-private-key`: Security check for private keys
- `check-case-conflict`: Prevents case conflicts
- `check-illegal-windows-names`: Prevents Windows illegal filenames
- `check-toml`: Validates TOML syntax
- `check-symlinks`: Checks for broken symlinks
- `check-executables-have-shebangs`: Ensures scripts have shebangs
- `check-shebang-scripts-are-executable`: Ensures shebang scripts are executable
- `check-vcs-permalinks`: Validates VCS permalinks
- `forbid-new-submodules`: Prevents new git submodules
- `pretty-format-json`: Auto-formats JSON files
- `mixed-line-ending`: Fixes mixed line endings

**Configuration**: [../.pre-commit-config.yaml](../.pre-commit-config.yaml) under `repos[0].hooks`

**Customization**: Standard pre-commit-hooks configuration options available.

### 2. Go Formatting (gofumpt)

**Purpose**: Strict Go code formatting with enhanced rules beyond standard `gofmt`.

**Configuration**:
```yaml
- id: gofumpt
  name: gofumpt (Go formatting)
  entry: gofumpt
  args: [-extra, -w]
  language: system
  files: '\.go$'
  exclude: '_gen\.go$|\.pb\.go$|vendor/|api/client|api/model|api/server|test/'
```

**Key Parameters**:
- `-extra`: Enables additional formatting rules
- `-w`: Write changes to files (auto-fix mode)
- `exclude`: Skips generated code and test directories

**Integration**: Works with VS Code settings in [../.vscode/settings.json](../.vscode/settings.json):
```json
{
  "go.formatTool": "gofumpt",
  "gopls": {
    "formatting.gofumpt": true
  }
}
```

**Documentation**: [gofumpt GitHub](https://github.com/mvdan/gofumpt)

### 3. Go Import Organization (goimports)

**Purpose**: Automatically organizes and formats Go import statements.

**Configuration**:
```yaml
- id: goimports
  name: goimports (Go import organization)
  entry: goimports
  args: [-w]
  language: system
  files: '\.go$'
  exclude: '_gen\.go$|\.pb\.go$|vendor/|api/client|api/model|api/server|test/'
```

**Key Parameters**:
- `-w`: Write changes to files (auto-fix mode)
- `exclude`: Skips generated code and test directories

**Integration**: Works with golangci-lint import organization checks.

**Documentation**: [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)

### 4. Go Module Tidy (go mod tidy)

**Purpose**: Cleans up Go module dependencies by removing unused modules and adding missing ones.

**Configuration**:
```yaml
- id: go-mod-tidy
  name: go mod tidy
  entry: go
  args: [mod, tidy]
  language: system
  pass_filenames: false
  files: go\.mod$
```

**Key Parameters**:
- `files: go\.mod$`: Only runs when go.mod changes
- `pass_filenames: false`: Operates on entire module

**Rationale**: Running before golangci-lint ensures dependency analysis is accurate.

**Documentation**: [Go Modules](https://go.dev/ref/mod)

### 5. Go Linting (golangci-lint)

**Purpose**: Comprehensive Go code linting with 40+ built-in linters.

**Configuration**:
```yaml
- id: golangci-lint
  name: golangci-lint (full suite)
  entry: golangci-lint
  args: [run, --timeout=10m]
  language: system
  pass_filenames: false
  stages: [pre-commit]
```

**Key Parameters**:
- `--timeout=10m`: Prevents hanging on complex codebases
- `stages: [pre-commit]`: Runs on commit (not just push)

**Integration**: Uses [../.golangci.yml](../.golangci.yml) for detailed configuration including:
- Enabled/disabled linters
- Custom settings per linter
- Exclusion rules for generated code
- Severity levels and output formatting

**Enabled Linters**: errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports, revive, stylecheck, gosec, godot, noctx, wrapcheck, thelper, tparallel, testpackage, gomodguard, gomoddirectives, prealloc, bodyclose, copyloopvar, goconst, importas, mnd, wsl, nlreturn, goheader, errorlint

**Documentation**: [golangci-lint](https://golangci-lint.run/)

### 6. Go Build Check (go build)

**Purpose**: Verifies that Go code compiles successfully.

**Configuration**:
```yaml
- id: go-build
  name: go build
  entry: go
  args: [build, ./...]
  language: system
  pass_filenames: false
```

**Key Parameters**:
- `./...`: Builds all packages recursively
- `pass_filenames: false`: Operates on entire project

**Rationale**: Runs after linting to ensure code quality doesn't break compilation.

**Documentation**: [Go Build](https://go.dev/cmd/go/#hdr-Compile_packages_and_dependencies)

### 7. Test Pattern Enforcement

**Purpose**: Enforces project-specific testing patterns and conventions.

**Configuration**:
```yaml
- id: test-patterns
  name: Enforce test patterns (UUIDv7, testify assertions)
  entry: go
  args: [run, scripts/cicd_checks.go, enforce-test-patterns]
  language: system
  pass_filenames: false
  files: go\.mod$
```

**Key Parameters**:
- `files: go\.mod$`: Only runs when go.mod changes (not per test file)
- Custom script: [../scripts/cicd_checks.go](../scripts/cicd_checks.go)

**Enforced Patterns**:
- UUIDv7 usage for test uniqueness
- testify assertion patterns
- Test file organization conventions

**Documentation**: See [../scripts/cicd_checks.go](../scripts/cicd_checks.go) for implementation details.

### 8. GitHub Actions Linting (actionlint)

**Purpose**: Lints GitHub Actions workflow files for syntax and best practices.

**Configuration**:
```yaml
- repo: https://github.com/rhysd/actionlint
  rev: v1.7.8
  hooks:
    - id: actionlint
```

**Key Features**:
- Validates workflow syntax
- Checks for deprecated features
- Ensures security best practices
- Validates action versions and permissions

**Documentation**: [actionlint](https://github.com/rhysd/actionlint)

### 9. Dockerfile Linting (hadolint)

**Purpose**: Lints Dockerfiles for best practices and security issues.

**Configuration**:
```yaml
- repo: https://github.com/hadolint/hadolint
  rev: v2.14.0
  hooks:
    - id: hadolint-docker
```

**Key Features**:
- Dockerfile best practices
- Security vulnerability detection
- Performance optimization suggestions
- Multi-stage build validation

**Documentation**: [hadolint](https://github.com/hadolint/hadolint)

### 10. Shell Script Linting (shellcheck)

**Purpose**: Lints shell scripts for common issues and best practices.

**Configuration**:
```yaml
- repo: https://github.com/koalaman/shellcheck-precommit
  rev: v0.10.0
  hooks:
    - id: shellcheck
      name: Shell script linting
      args: [--severity=warning, --exclude=SC1111]
```

**Key Parameters**:
- `--severity=warning`: Only show warnings and above
- `--exclude=SC1111`: Exclude specific rule (SC1111 is about dynamic paths)

**Documentation**: [shellcheck](https://www.shellcheck.net/)

### 11. Python Security Linting (bandit)

**Purpose**: Security linting for Python code in scripts and utilities.

**Configuration**:
```yaml
- repo: https://github.com/PyCQA/bandit
  rev: 1.8.0
  hooks:
    - id: bandit
      name: Python security linting
      files: '\.py$'
```

**Key Features**:
- Detects common security issues in Python code
- Configurable severity levels
- Integration with safety vulnerability database

**Documentation**: [bandit](https://bandit.readthedocs.io/)

### 12. Conventional Commit Validation (commitizen)

**Purpose**: Enforces conventional commit message formatting.

**Configuration**:
```yaml
- repo: https://github.com/commitizen-tools/commitizen
  rev: v3.29.1
  hooks:
    - id: commitizen
      name: Check conventional commit message
      stages: [commit-msg]
```

**Key Features**:
- Validates commit message format
- Supports conventional commits specification
- Configurable commit types and scopes

**Documentation**: [commitizen](https://commitizen-tools.github.io/commitizen/)

## Customization and Maintenance

### Modifying the Pipeline

To modify hook ordering or configuration:

1. Edit [../.pre-commit-config.yaml](../.pre-commit-config.yaml)
2. Test changes: `pre-commit run --all-files`
3. Update this documentation if significant changes are made

### Performance Tuning

- Adjust timeouts for slower systems
- Consider disabling expensive checks for rapid development
- Use `pre-commit run --files <specific-files>` for targeted testing

### Troubleshooting

**Common Issues**:
- **Hook failures**: Run `pre-commit run --all-files` to identify issues
- **Cache issues**: Clear cache with `pre-commit clean`
- **Version conflicts**: Update hooks with `pre-commit autoupdate`

**Debugging**: Use `pre-commit run <hook-id> --verbose` for detailed output.

## Integration with Development Workflow

This pipeline integrates with:
- **VS Code**: Settings in [../.vscode/settings.json](../.vscode/settings.json)
- **CI/CD**: GitHub Actions workflows in [../.github/workflows/](../.github/workflows/)
- **Scripts**: Utility scripts in [../scripts/](../scripts/)
- **Instructions**: Copilot guidance in [../.github/instructions/](../.github/instructions/)

The pipeline ensures consistent code quality across local development, CI/CD, and team contributions.</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\docs\pre-commit-hooks.md
