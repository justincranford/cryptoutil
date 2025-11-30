# Pre-commit, CI/CD, and Linting Speed & Coverage Analysis

**Generated**: 2025-11-30
**Purpose**: Deep analysis of filtering opportunities and additional checks for `.pre-commit-config.yaml`, `cicd.go`, and `.golangci.yml`

---

## Executive Summary

This document analyzes three key tooling configurations:

1. **`.pre-commit-config.yaml`** - Git hooks for pre-commit and pre-push validation
2. **`internal/cmd/cicd/cicd.go`** - Custom Go-based CI/CD command runner
3. **`.golangci.yml`** - golangci-lint v2 configuration

---

## Part 1: Speed Optimizations

### 1.1 Pre-commit-config.yaml Filtering Opportunities

#### Current Pre-commit Stage Hooks (Fast - Keep on pre-commit)

| Hook | Time | Files Filter | Status |
|------|------|--------------|--------|
| end-of-file-fixer | <1s | All (exclude binary) | ✅ Optimal |
| trailing-whitespace | <1s | All | ✅ Optimal |
| fix-byte-order-marker | <1s | All (exclude binary) | ✅ Optimal |
| check-yaml | <1s | `*.yaml, *.yml` | ✅ Optimal |
| check-json | <1s | `*.json` | ✅ Optimal |
| check-added-large-files | <1s | All | ✅ Optimal |
| check-merge-conflict | <1s | All | ✅ Optimal |
| detect-aws-credentials | <1s | All | ✅ Optimal |
| detect-private-key | <1s | All | ✅ Optimal |
| check-case-conflict | <1s | All | ✅ Optimal |
| check-illegal-windows-names | <1s | All | ✅ Optimal |
| check-toml | <1s | `*.toml` | ✅ Optimal |
| check-xml | <1s | `*.xml` | ✅ Optimal |
| check-symlinks | <1s | All | ✅ Optimal |
| check-executables-have-shebangs | <1s | Scripts | ✅ Optimal |
| check-vcs-permalinks | <1s | All | ✅ Optimal |
| pretty-format-json | <2s | `*.json` | ✅ Optimal |
| mixed-line-ending | <1s | All | ✅ Optimal |
| markdownlint-cli2 | 2-5s | `*.md` | ✅ Optimal |
| check-todo-severity | <1s | `*.go` | ✅ Optimal |

#### Current Pre-push Stage Hooks (Slower - Appropriate for pre-push)

| Hook | Time | Files Filter | Status |
|------|------|--------------|--------|
| golangci-lint-full | 30-120s | `*.go` | ✅ Appropriate |
| go-build | 10-60s | `*.go, go.mod, go.sum` | ✅ Appropriate |
| identity-progressive-validation | 30-120s | All Go | ✅ Appropriate |
| cspell | 5-30s | Multiple | ✅ Moved to pre-push |
| github-workflow-lint | 5-15s | `.github/workflows/*.yml` | ✅ Appropriate |

#### Recommendations for Speed Improvements

**HIGH IMPACT - Move to pre-push:**

1. **`golangci-lint` incremental** - Currently on pre-commit with `--new-from-rev=HEAD~1`. Consider removing entirely since `golangci-lint-full` runs on pre-push.
   - Savings: 5-30s per commit
   - Risk: Low - full validation catches everything on push

2. **`go-fix-all`** - Currently on pre-commit, runs Go auto-fixes.
   - Recommendation: Move to pre-push or make manual
   - Savings: 5-15s per commit
   - Risk: Medium - fixes may need to be staged

3. **`cicd-enforce-internal`** - Runs UTF-8 enforcement, test patterns, any type enforcement.
   - Recommendation: Add file filter `files: '\.go$'` to skip non-Go files
   - Savings: 2-5s per commit

**MEDIUM IMPACT - Add file filters:**

4. **`actionlint`** - Already has files filter ✅
5. **`hadolint-docker`** - Already has files filter ✅
6. **`shellcheck`** - Already has files filter ✅
7. **`bandit`** - Already has files filter ✅

**LOW IMPACT - Configuration tweaks:**

8. Add `require_serial: false` to all hooks that don't have it for parallel execution
9. Add `--cache` to golangci-lint (already using file-based cache via config)

### 1.2 cicd.go Filtering Opportunities

#### Current Command Performance

| Command | Time | Filtering | Status |
|---------|------|-----------|--------|
| `all-enforce-utf8` | 2-5s | Lists all files | ⚠️ Could filter |
| `go-enforce-test-patterns` | 2-5s | Lists all files | ⚠️ Could filter to `_test.go` |
| `go-enforce-any` | 2-5s | Lists all files | ⚠️ Could filter to `*.go` |
| `github-workflow-lint` | 5-15s | Lists all files | ⚠️ Could filter to `.github/workflows/` |
| `go-fix-*` commands | 5-15s | Lists all files | ⚠️ Could filter to `*.go` |
| `go-check-circular-package-dependencies` | 1-3s | Static analysis | ✅ Optimal |
| `go-check-identity-imports` | 1-3s | Static analysis | ✅ Optimal |
| `identity-progressive-validation` | 30-120s | Runs tests | ✅ Optimal |

#### Recommendations for cicd.go Speed

**HIGH IMPACT:**

1. **Add file extension filtering to `ListAllFiles`**:
   ```go
   // Current: Lists all files
   allFiles, err = cryptoutilFiles.ListAllFiles(".")
   
   // Improved: Filter by extension
   allFiles, err = cryptoutilFiles.ListAllFilesWithExtensions(".", []string{".go"})
   ```

2. **Implement parallel execution for independent commands**:
   ```go
   // Run go-check-circular-package-dependencies and go-check-identity-imports in parallel
   ```

3. **Add early exit for unchanged files**:
   - If no `.go` files changed since last run, skip Go-specific checks
   - Use git status or file modification times

**MEDIUM IMPACT:**

4. **Add caching for expensive operations**:
   - Cache circular dependency graph
   - Cache import analysis results
   - Invalidate on `go.mod` or package structure changes

5. **Add `--changed-only` flag**:
   - Only process files changed since HEAD~1 (like golangci-lint)

### 1.3 golangci.yml Filtering Opportunities

#### Current Configuration Analysis

| Setting | Current | Optimized | Savings |
|---------|---------|-----------|---------|
| `timeout` | 10m | Keep | N/A |
| `concurrency` | 0 (auto) | Keep | Already optimal |
| `max-issues-per-linter` | 50 | Keep | Limits output |
| `max-same-issues` | 5 | Keep | Limits duplicates |

#### Recommendations for golangci-lint Speed

**HIGH IMPACT:**

1. **Add exclude patterns for generated code**:
   ```yaml
   issues:
     exclude-dirs:
       - api/client
       - api/model
       - api/server
       - api/idp
       - api/authz
   ```
   - Savings: 10-20% of lint time
   - Already have `goheader` exclusion, extend to all linters

2. **Use `--new-from-rev` in CI**:
   - For PR checks, only lint changed files
   - Savings: 50-80% in incremental runs

**MEDIUM IMPACT:**

3. **Disable expensive linters in pre-commit**:
   - Create `.golangci.precommit.yml` with faster subset
   - Disable: `gosec`, `prealloc`, `wrapcheck` (slower)
   - Keep: `gofumpt`, `govet`, `errcheck`, `staticcheck` (fast, high value)

4. **Add skip-dirs for test output**:
   ```yaml
   run:
     skip-dirs:
       - test-output
       - test-reports
       - workflow-reports
   ```

---

## Part 2: Additional Linting & Formatting

### 2.1 Pre-commit-config.yaml - Additional Checks

#### HIGH Quality Impact (Enable First)

| Tool | Quality Impact | Performance | Recommendation |
|------|---------------|-------------|----------------|
| **secrets-check** (gitleaks) | 🔴 Critical | 2-5s | Enable on pre-push |
| **yaml-lint** (yamllint) | 🟠 High | 1-2s | Enable on pre-commit |
| **docker-compose-check** | 🟠 High | 1-2s | Enable on pre-commit |
| **openapi-lint** | 🟠 High | 2-5s | Enable on pre-push |

**Recommended Additions:**

```yaml
# Secrets scanning (Critical for crypto project)
- repo: https://github.com/gitleaks/gitleaks
  rev: v8.18.4
  hooks:
    - id: gitleaks
      name: Scan for secrets
      stages: [pre-push]

# Stricter YAML validation
- repo: https://github.com/adrienverge/yamllint
  rev: v1.35.1
  hooks:
    - id: yamllint
      name: YAML strict lint
      args: [-d, relaxed]
      files: '\.(yaml|yml)$'
      stages: [pre-commit]

# Docker Compose validation
- repo: https://github.com/IamTheFij/docker-pre-commit
  rev: v3.0.1
  hooks:
    - id: docker-compose-check
      files: 'compose.*\.ya?ml$|docker-compose.*\.ya?ml$'
      stages: [pre-commit]
```

#### MEDIUM Quality Impact

| Tool | Quality Impact | Performance | Recommendation |
|------|---------------|-------------|----------------|
| **codespell** | 🟡 Medium | 2-5s | Already have cspell |
| **typos** | 🟡 Medium | 1-3s | Faster than cspell, consider replacement |
| **editorconfig-checker** | 🟡 Medium | 1-2s | Enable on pre-commit |
| **go-license-check** | 🟡 Medium | 2-5s | Enable on pre-push |

**Recommended Additions:**

```yaml
# EditorConfig enforcement
- repo: https://github.com/editorconfig-checker/editorconfig-checker.python
  rev: 3.2.1
  hooks:
    - id: editorconfig-checker
      name: Check EditorConfig
      exclude: '^(test-output/|vendor/|\.git/)'
      stages: [pre-commit]

# Faster spell checker (alternative to cspell)
- repo: https://github.com/crate-ci/typos
  rev: v1.28.4
  hooks:
    - id: typos
      name: Check typos (fast)
      stages: [pre-commit]  # Fast enough for pre-commit
```

#### LOW Quality Impact

| Tool | Quality Impact | Performance | Recommendation |
|------|---------------|-------------|----------------|
| **check-ast** | 🟢 Low | <1s | Already covered by Go compiler |
| **debug-statements** | 🟢 Low | <1s | Go doesn't have debug statements |
| **file-contents-sorter** | 🟢 Low | <1s | Optional, for sorted imports/deps |

### 2.2 cicd.go - Additional Checks

#### HIGH Quality Impact (Implement First)

| Check | Quality Impact | Performance | Description |
|-------|---------------|-------------|-------------|
| **go-dead-code** | 🔴 Critical | 5-10s | Find unreachable code |
| **go-security-audit** | 🔴 Critical | 10-30s | Deep security scan |
| **go-doc-coverage** | 🟠 High | 2-5s | Ensure exported symbols have docs |
| **go-test-coverage-gate** | 🟠 High | 30-60s | Fail if coverage drops |

**Recommended Additions to cicd.go:**

```go
// New commands to add:
cmdGoDeadCode              = "go-dead-code"           // Find unreachable code
cmdGoSecurityAudit         = "go-security-audit"      // Deep gosec scan
cmdGoDocCoverage           = "go-doc-coverage"        // Check doc comments
cmdGoTestCoverageGate      = "go-test-coverage-gate"  // Enforce coverage threshold
cmdGoModuleGraph           = "go-module-graph"        // Visualize dependencies
cmdGoVulnCheck             = "go-vuln-check"          // Check for known vulns
```

#### MEDIUM Quality Impact

| Check | Quality Impact | Performance | Description |
|-------|---------------|-------------|-------------|
| **go-struct-align** | 🟡 Medium | 2-5s | Optimize struct memory layout |
| **go-interface-check** | 🟡 Medium | 2-5s | Verify interface implementations |
| **go-error-wrap-check** | 🟡 Medium | 2-5s | Ensure errors have context |
| **go-test-race** | 🟡 Medium | 60-120s | Run tests with race detector |

#### LOW Quality Impact

| Check | Quality Impact | Performance | Description |
|-------|---------------|-------------|-------------|
| **go-line-count** | 🟢 Low | <1s | Warn on long files |
| **go-func-count** | 🟢 Low | <1s | Warn on files with many funcs |
| **go-comment-density** | 🟢 Low | 1-2s | Check comment ratio |

### 2.3 golangci.yml - Additional Linters

#### HIGH Quality Impact (Enable First)

| Linter | Quality Impact | Performance | Description |
|--------|---------------|-------------|-------------|
| **exhaustive** | 🔴 Critical | 2-5s | Check switch exhaustiveness |
| **nilnil** | 🔴 Critical | 1-2s | Check nil returns |
| **nilerr** | 🔴 Critical | 1-2s | Check nil error returns |
| **exhaustruct** | 🟠 High | 3-5s | Require all struct fields |
| **musttag** | 🟠 High | 2-3s | Require struct field tags |
| **makezero** | 🟠 High | 1-2s | Check slice initialization |

**Recommended Additions to .golangci.yml:**

```yaml
linters:
  enable:
    # Existing (keep all)
    # NEW - High Impact
    - exhaustive     # Check switch exhaustiveness
    - nilnil         # Check nil returns
    - nilerr         # Check nil error returns
    - makezero       # Check slice initialization
    # NEW - Medium Impact
    - forcetypeassert # Require type assertion checks
    - gochecknoinits  # Discourage init() functions
    - promlinter      # Prometheus metrics naming
    - tagalign        # Align struct tags
```

#### MEDIUM Quality Impact

| Linter | Quality Impact | Performance | Description |
|--------|---------------|-------------|-------------|
| **forcetypeassert** | 🟡 Medium | 1-2s | Require type assertion checks |
| **gochecknoinits** | 🟡 Medium | <1s | Discourage init() functions |
| **promlinter** | 🟡 Medium | 1-2s | Prometheus metrics naming |
| **tagalign** | 🟡 Medium | 1-2s | Align struct tags |
| **tagliatelle** | 🟡 Medium | 1-2s | Struct tag naming |
| **usestdlibvars** | 🟡 Medium | 1-2s | Use stdlib vars (http.MethodGet) |
| **whitespace** | 🟡 Medium | <1s | Whitespace checks |

#### LOW Quality Impact (Optional)

| Linter | Quality Impact | Performance | Description |
|--------|---------------|-------------|-------------|
| **gochecknoglobals** | 🟢 Low | <1s | Discourage global vars |
| **ireturn** | 🟢 Low | 1-2s | Accept interfaces, return concrete |
| **varnamelen** | 🟢 Low | 1-2s | Variable name length |
| **funlen** | 🟢 Low | <1s | Function length (300 soft limit) |
| **cyclop** | 🟢 Low | 1-2s | Alternative cyclomatic complexity |

---

## Part 3: Implementation Priority Matrix

### Phase 1: Quick Wins (Week 1)

| Change | Type | Impact | Effort | Files |
|--------|------|--------|--------|-------|
| Move go-fix-all to pre-push | Speed | High | Low | `.pre-commit-config.yaml` |
| Add file filters to cicd commands | Speed | High | Medium | `cicd.go`, `common/*.go` |
| Enable exhaustive linter | Quality | High | Low | `.golangci.yml` |
| Enable nilnil/nilerr linters | Quality | High | Low | `.golangci.yml` |
| Add gitleaks to pre-push | Quality | Critical | Low | `.pre-commit-config.yaml` |

### Phase 2: Medium-Term (Week 2-3)

| Change | Type | Impact | Effort | Files |
|--------|------|--------|--------|-------|
| Add --changed-only to cicd | Speed | High | Medium | `cicd.go` |
| Add skip-dirs to golangci | Speed | Medium | Low | `.golangci.yml` |
| Add yamllint hook | Quality | High | Low | `.pre-commit-config.yaml` |
| Implement go-doc-coverage | Quality | High | Medium | `cicd.go` |
| Enable makezero linter | Quality | High | Low | `.golangci.yml` |

### Phase 3: Long-Term (Week 4+)

| Change | Type | Impact | Effort | Files |
|--------|------|--------|--------|-------|
| Parallel command execution | Speed | High | High | `cicd.go` |
| Create fast golangci preset | Speed | Medium | Medium | New file |
| Add go-vuln-check command | Quality | High | Medium | `cicd.go` |
| Add coverage gating | Quality | High | Medium | `cicd.go` |
| Enable exhaustruct (selective) | Quality | Medium | High | `.golangci.yml` |

---

## Appendix A: Performance Benchmarks

### Current Pre-commit Timing (Typical)

```
end-of-file-fixer............Passed  0.2s
trailing-whitespace..........Passed  0.3s
check-yaml...................Passed  0.4s
check-json...................Passed  0.2s
golangci-lint (incremental)..Passed  8.5s  ← Bottleneck
go-fix-all...................Passed  5.2s  ← Bottleneck
markdownlint-cli2............Passed  2.1s
check-todo-severity..........Passed  0.3s
cicd-checks-internal.........Passed  3.5s
cicd-enforce-internal........Passed  4.2s
--------------------------------------------
TOTAL                              ~25s
```

### Current Pre-push Timing (Typical)

```
golangci-lint-full...........Passed  45.3s ← Bottleneck
go-build.....................Passed  12.1s
identity-progressive-validation.Passed  67.2s ← Bottleneck
cspell.......................Passed  8.5s
github-workflow-lint.........Passed  6.3s
--------------------------------------------
TOTAL                              ~140s
```

### Projected After Optimizations

**Pre-commit: ~15s** (40% improvement)
- Remove incremental golangci-lint (covered by pre-push)
- Move go-fix-all to pre-push
- Add file filters to cicd commands

**Pre-push: ~100s** (30% improvement)
- Add skip-dirs for generated code
- Add --changed-only for incremental checks
- Parallelize independent commands

---

## Appendix B: Disabled Linters Rationale

### Currently Disabled

| Linter | Reason | Reconsider? |
|--------|--------|-------------|
| `dupl` | False positives in table-driven tests | No |
| `gocyclo` | Replaced by `cyclop` | No |
| `godox` | Conflicts with TODO tracking | No |

### Intentionally Not Enabled

| Linter | Reason | Reconsider? |
|--------|--------|-------------|
| `gochecknoglobals` | Too strict for magic constants | No |
| `ireturn` | Conflicts with interface design patterns | No |
| `varnamelen` | Too strict, subjective | No |
| `exhaustruct` | Too strict for partial struct init | Maybe |

---

## Appendix C: References

- [golangci-lint v2 Configuration](https://golangci-lint.run/usage/configuration/)
- [pre-commit Hooks](https://pre-commit.com/hooks.html)
- [Go Security Checklist](https://github.com/securego/gosec)
- [Pre-commit Performance Tips](https://pre-commit.com/#performance)

