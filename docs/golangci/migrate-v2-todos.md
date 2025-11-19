# golangci-lint v2 Migration - TODO Steps

## Strategy Overview

Migrate from mixed v1/v2 configuration to clean v2 configuration in 12 incremental steps.

**Approach:**
1. Keep first 12 lines (header comments + version declaration)
2. Create fresh minimal v2 config from reference/docs/schemas
3. Migrate content from `.golangci.yml.backup` in order of execution speed (fastest first)
4. Commit with verify after each step to catch issues early
5. Document problems in `migrate-v2-problems.md` after each step

## Migration Steps (Fast → Slow)

### Step 1: Core Configuration ✅
- `version: "2"`
- Basic `run:` settings (timeout, exit-code, tests)
- Skip directories/files patterns
- **Rationale**: Foundation for all linters, no execution cost

### Step 2: Output Configuration ✅
- `output:` section (formats, sort-results)
- **Rationale**: No execution cost, only affects display

### Step 3: Fast Essential Linters ✅
- Enable: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`
- **Rationale**: Fast analyzers, catch critical bugs

### Step 4: Fast Code Quality Linters ✅
- Enable: `revive`, `godot`, `copyloopvar`, `goconst`, `importas`
- **Rationale**: Fast static checks, improve maintainability

### Step 5: Security & Error Handling ✅
- Enable: `gosec`, `noctx`, `wrapcheck`, `errorlint`
- Settings: `gosec` excludes, `wrapcheck` ignoreSigs
- **Rationale**: Security-critical but still fast

### Step 6: Testing Quality Linters ✅
- Enable: `thelper`, `tparallel`, `testpackage`, `gomodguard`, `gomoddirectives`
- Settings: `thelper`, `testpackage`
- **Rationale**: Test-focused, run only on test files

### Step 7: Performance & Style Linters ✅
- Enable: `prealloc`, `bodyclose`, `mnd`, `wsl`, `nlreturn`
- Settings: `mnd` (ignored-numbers, ignored-functions)
- **Rationale**: Performance hints, whitespace rules

### Step 8: Maintainability & Headers ✅
- Enable: `goheader`, `depguard`
- Settings: `goheader`, `depguard`, `lll`
- **Rationale**: Project-specific rules

### Step 9: Linter-Specific Settings ✅
- Settings: `errcheck`, `gocyclo`, `goconst`, `dupl`, `copyloopvar`
- Settings: `misspell`, `godot`, `godox`, `wsl`
- **Rationale**: Configure already-enabled linters

### Step 10: Import Alias Configuration ✅
- Settings: Complete `importas.alias` list
- **Rationale**: Large but fast string matching

### Step 11: Exclusions Configuration ✅
- `linters.exclusions.generated: lax`
- `linters.exclusions.paths`: IDE, VCS, non-Go, generated, artifacts
- `linters.exclusions.rules`: Test files, cicd.go/cicd_test.go
- **Rationale**: Reduces linter workload

### Step 12: Issue Management & Severity ✅
- `issues:` section (uniq-by-line, max-issues, max-same-issues)
- `severity:` section (default: error, rules for warning/info)
- **Rationale**: Final tuning of output quality

## Post-Migration Validation

After all 12 steps:
1. Run `golangci-lint run --timeout=10m` on full codebase
2. Compare output with backup config
3. Review `migrate-v2-problems.md` for patterns
4. Update instruction files if needed
