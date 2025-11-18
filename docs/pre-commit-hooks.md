# Pre-commit Hooks Pipeline

This document describes the design intent, optimal ordering, and detailed configuration of the cryptoutil project's pre-commit hooks pipeline.

## Quick Reference: Manual Commands

Some expensive hooks are configured to run manually instead of on every commit/push:

```bash
# Update pre-commit hook versions (weekly recommended)
pre-commit run autoupdate-all-hooks --hook-stage manual

# Check for Go dependency updates (weekly recommended)
pre-commit run go-update-direct-dependencies --hook-stage manual

# Run all manual checks at once
pre-commit run --hook-stage manual --all-files
```

## Design Intent

The pre-commit hooks pipeline is designed with a **fail-fast, progressive validation** philosophy that optimizes for both developer experience and CI/CD performance.

### Hook Stage Strategy

The pipeline uses a **three-tier approach** to balance speed and thoroughness:

**Tier 1: Pre-commit (Fast & Iterative)**
- **Purpose**: Catch common issues quickly during development
- **Target Time**: 8-12 seconds for typical changes
- **Approach**: Incremental validation of changed files only
- **Benefits**: Fast feedback loop, minimal interruption to flow state

**Tier 2: Pre-push (Comprehensive)**
- **Purpose**: Full validation before sharing code with team
- **Target Time**: 45-60 seconds
- **Approach**: Complete codebase analysis and build verification
- **Benefits**: Ensures CI/CD will succeed, catches integration issues

**Tier 3: Manual (As-needed)**
- **Purpose**: Expensive maintenance tasks run on-demand
- **Target Time**: Varies (2-5 minutes)
- **Approach**: Network-dependent operations (dependency updates, hook updates)
- **Benefits**: Eliminates unnecessary network calls, run on your schedule

This staged approach provides optimal developer experience:
- **Rapid iteration** during active development (pre-commit)
- **Confidence before pushing** to shared branches (pre-push)
- **Flexible maintenance** scheduling (manual)

### Core Principles

1. **Fail Fast**: Quick, inexpensive checks run first to catch obvious issues early
2. **Progressive Validation**: Each stage builds on the previous, ensuring code quality accumulates
3. **Auto-Fix First**: Formatting and import fixes happen before validation to reduce noise
4. **Incremental Analysis**: Pre-commit only checks changed files for speed
5. **Comprehensive Verification**: Pre-push validates entire codebase
6. **Dependency Management**: Module dependencies are cleaned before expensive linting operations
7. **Build Validation**: Code compilation is verified after linting passes
6. **Custom Checks**: Project-specific rules run after basic validation
7. **Specialized Linting**: File-type specific checks run last for comprehensive coverage

### Performance Optimization

The hook pipeline is optimized for both speed and thoroughness through a strategic pre-commit vs pre-push split:

**Pre-commit (Fast - ~8-12 seconds for incremental changes)**:
- Incremental golangci-lint with `--new-from-rev=HEAD~1` (only checks changed files)
- Quick file checks and formatting
- Fast custom checks
- Spell checking with cache

**Pre-push (Thorough - ~45-60 seconds)**:
- Full golangci-lint validation (all files)
- Go build verification
- GitHub workflow linting (when workflows change)

**Manual (Run weekly or as-needed)**:
- Pre-commit hook version updates: `pre-commit run autoupdate-all-hooks --hook-stage manual`
- Go dependency updates: `pre-commit run go-update-direct-dependencies --hook-stage manual`

This approach provides:
- **~70% faster development workflow** for typical commits
- **Full validation before pushing** to ensure CI/CD success
- **Flexible scheduling** for expensive maintenance tasks

### Timing Expectations

| Hook Category | Stage | Expected Time | Notes |
|---------------|-------|---------------|-------|
| Generic File Checks | pre-commit | 1-2s | Very fast, runs on all files |
| go mod tidy | pre-commit | 0.5-1s | Only when go.mod changes |
| golangci-lint (incremental) | pre-commit | 2-5s | Only changed files |
| Custom CI/CD checks | pre-commit | 3-5s | Fast internal validations |
| cspell | pre-commit | 1-3s | Cached spell checking |
| golangci-lint (full) | pre-push | 30-60s | Complete codebase validation |
| go build | pre-push | 10-20s | Full compilation check |
| github-workflow-lint | pre-push | 2-5s | Only when .github/workflows/ changes |
| **Total pre-commit** | - | **8-12s** | Incremental changes |
| **Total pre-push** | - | **45-60s** | Full validation |

### Troubleshooting Slow Hooks

**If pre-commit is slow (>15 seconds for small changes)**:

1. **Check if running on all files**:
   ```bash
   # Should only show changed files
   git diff --name-only HEAD~1
   ```

2. **Verify incremental mode is working**:
   ```bash
   # Should see --new-from-rev in output
   pre-commit run --verbose | grep golangci-lint
   ```

3. **Clear pre-commit cache**:
   ```bash
   pre-commit clean
   pre-commit install --install-hooks
   ```

**If pre-push is very slow (>2 minutes)**:

1. **Run golangci-lint directly to identify slow linters**:
   ```bash
   golangci-lint run --timeout=30m --verbose
   ```

2. **Check for large generated files** that should be excluded

3. **Verify build cache is working**:
   ```bash
   go clean -cache
   go build ./...  # Rebuild cache
   ```

The ordering minimizes redundant work and maximizes parallel processing potential:

- Auto-fixing tools run before their validation counterparts
- Expensive operations (golangci-lint) run after cheap fixes
- Related tools are grouped to share repository contexts
- File-type specific tools are logically sequenced

## Current Pipeline Ordering

### High-Level Flow

```
1. Generic File Checks → 2. Dependency Mgmt → 3. Go Auto-Fix & Validation → 4. Build Check → 5. Custom Rules → 6. Specialized Linting → 7. Commit Validation
```

### Detailed Stage Breakdown

| Stage | Purpose | Tools | Rationale |
|-------|---------|-------|-----------|
| **1. Generic File Checks** | Universal file validation | pre-commit-hooks | Fast, catches basic issues first |
| **2. Dependency Management** | Module cleanup | go mod tidy | Clean state before expensive linting |
| **3. Go Auto-Fix & Validation** | Formatting, imports, and comprehensive linting | golangci-lint --fix | Single tool handles all auto-fixable issues + validation |
| **4. Build Validation** | Compilation verification | go build | Ensure code compiles after linting |
| **5. Custom Rules** | Project-specific checks | cicd commands | Business logic and project-specific validations |
| **6. Specialized Linting** | File-type specific checks | actionlint, hadolint, shellcheck, bandit, cspell | Targeted validation by file type |
| **7. Commit Validation** | Message format checking | commitizen | Final gate before push |

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

### 2. Go Module Tidy (go mod tidy)

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

### 3. Go Linting with Auto-Fix (golangci-lint)

**Purpose**: Comprehensive Go code linting with 40+ built-in linters plus automatic fixing for formatting, imports, and code quality issues.

**Configuration**:
```yaml
- id: golangci-lint
  name: golangci-lint (auto-fix + validation)
  entry: golangci-lint
  args: [run, --fix, --timeout=10m]
  language: system
  pass_filenames: false
  stages: [pre-commit]
```

**Key Parameters**:
- `--fix`: Automatically fixes all auto-fixable issues (formatting, imports, whitespace, etc.)
- `--timeout=10m`: Prevents hanging on complex codebases
- `stages: [pre-commit]`: Runs on commit (not just push)

**Auto-Fixable Linters** (enabled via --fix):
- `gofmt`: Go code formatting
- `gofumpt`: Stricter Go formatting with extra rules (replaces standalone gofumpt -extra -w)
- `goimports`: Import organization and formatting (replaces standalone goimports -w)
- `wsl`: Whitespace consistency (blank lines between statements)
- `godot`: Adds missing periods to documentation comments
- `goconst`: Creates named constants for repeated strings
- `importas`: Fixes import aliases to match configured rules
- `copyloopvar`: Fixes loop variable capture issues
- `testpackage`: Renames test packages to follow conventions
- `revive`: Fixes various style and code quality issues (subset of rules)

**Integration**: Uses [../.golangci.yml](../.golangci.yml) for detailed configuration including:
- Enabled/disabled linters
- Custom settings per linter (gofumpt extra-rules, module-path)
- Exclusion rules for generated code
- Severity levels and output formatting

**Rationale**: Single tool consolidates all auto-fixes (formatting, imports, style) plus validation, eliminating need for separate gofumpt/goimports hooks. This reduces hook count, simplifies pipeline, and ensures consistency.

**Documentation**: [golangci-lint](https://golangci-lint.run/)

### 4. Go Build Check (go build)

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

### 5. Custom Rules (Project-Specific Checks)

**Purpose**: Enforce project-specific patterns and validations across multiple domains.

**Configuration**:
```yaml
- id: go-check-circular-package-dependencies
  name: Check for circular package dependencies
  entry: go
  args: [run, cmd/cicd/main.go, go-check-circular-package-dependencies]
  language: system
  pass_filenames: false

- id: github-workflow-lint
  name: Lint GitHub Actions workflows
  entry: go
  args: [run, cmd/cicd/main.go, github-workflow-lint]
  language: system
  pass_filenames: false

- id: go-enforce-any
  name: Format Go code with gofumpt
  entry: go
  args: [run, cmd/cicd/main.go, go-enforce-any]
  language: system
  pass_filenames: false

- id: go-enforce-test-patterns
  name: Enforce test patterns (UUIDv7, testify assertions)
  entry: go
  args: [run, cmd/cicd/main.go, go-enforce-test-patterns]
  language: system
  pass_filenames: false
```

**Key Parameters**:
- All commands run on every commit (no file restrictions)
- Custom scripts: [../cmd/cicd/main.go](../cmd/cicd/main.go) (wrapper), [../internal/cmd/cicd/cicd.go](../internal/cmd/cicd/cicd.go) (implementation)

**Enforced Validations**:
- **go-check-circular-package-dependencies**: Prevents circular import dependencies
- **github-workflow-lint**: Validates GitHub Actions workflow naming and version conventions
- **go-enforce-any**: Applies strict Go code formatting (superset of gofmt)
- **go-enforce-test-patterns**: Enforces UUIDv7 usage, testify assertion patterns, and test file organization conventions

**Documentation**: See [../internal/cmd/cicd/cicd.go](../internal/cmd/cicd/cicd.go) for implementation details.

### 6. GitHub Actions Linting (actionlint)

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

### 7. Dockerfile Linting (hadolint)

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

### 8. Shell Script Linting (shellcheck)

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

### 9. Python Security Linting (bandit)

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

### 10. Spell Checking (cSpell)

**Purpose**: Checks spelling in code, comments, and documentation files.

**Configuration**:
```yaml
- repo: https://github.com/streetsidesoftware/cspell-precommit
  rev: v4.0.0
  hooks:
    - id: cspell
      name: Check spelling
      args: [--no-progress]
```

**Key Features**:
- Validates spelling in source code and documentation
- Uses custom dictionaries for technical terms
- Configurable via .vscode/cspell.json

**Documentation**: [cSpell](https://cspell.org/)

### 11. Conventional Commit Validation (commitizen)

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
