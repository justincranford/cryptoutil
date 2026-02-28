# Lessons Learned - Details

Full explanations of lessons from quality enforcement and refactoring sessions.

---

## Lesson 1: Per-Function t.Parallel() Enforcement

**Category**: Testing

**Problem**: Package-level t.Parallel() checks (e.g., "at least one test in the package calls t.Parallel()") are insufficient. A package passes even if only one test calls t.Parallel() while dozens of others don't.

**Solution**: Enforce t.Parallel() at the per-function level. Every test function and every subtest (t.Run()) MUST call t.Parallel() explicitly.

**Impact**: Fixed 1123 files. Tests now run concurrently across the entire suite, exposing real race conditions.

**Pattern**:
```go
func TestFoo(t *testing.T) {
    t.Parallel()
    tests := []struct{ name string }{...}
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            // test body
        })
    }
}
```

---

## Lesson 2: magic_usage Linter - Literal-Use vs Const-Redefine

**Category**: Linting

**Problem**: The magic_usage linter has two violation categories: literal-use (using a raw string/number that matches a magic constant) and const-redefine (redefining a constant in a non-magic package). These must be treated differently.

**Solution**: Only literal-use violations are errors (blocking). Const-redefine violations are informational warnings only. Fixing all 188 literal-use violations required replacing raw literals with `cryptoutilSharedMagic.XXX` references throughout the codebase.

**Impact**: Fixed 37 files. The linter now correctly differentiates between the two violation types.

---

## Lesson 3: URLPrefixLocalhostHTTPS Magic Constant

**Category**: Code Quality

**Problem**: E2E TestMain files contained hardcoded `https://127.0.0.1` URL strings instead of using the canonical magic constant.

**Solution**: Replace all hardcoded `https://127.0.0.1` occurrences with `cryptoutilSharedMagic.URLPrefixLocalhostHTTPS`.

**Impact**: Fixed 7 e2e test files. Better maintainability - if the base URL prefix changes, only the constant needs updating.

---

## Lesson 4: NEVER Write Go Files via Shell Heredocs

**Category**: Tooling

**Problem**: Shell heredocs (`cat > file.go << 'EOF'`) strip leading tabs from Go source files. Go syntax requires actual tab characters for indentation. A heredoc-written .go file will have spaces instead of tabs, causing `go build` parse errors.

**Root Cause Example**:
```
func (m *MergedMigrationsFS) Open(name string) (fs.File, error) {
// After heredoc: closing brace is missing, or indentation is wrong
```

**Solution**: ALWAYS use Python writes for Go files, using explicit `\t` escape sequences.

---

## Lesson 5: Use Python Writes for Go Files

**Category**: Tooling

**Problem**: Go files require exact tab-based indentation. Shell heredocs and other text-writing approaches strip tabs.

**Solution**: Use Python with explicit `\t` characters:
```bash
python3 -c "
content = 'func foo() {\n\tbar()\n}\n'
open('file.go', 'w').write(content)
"
```

**Key**: Every tab in Go code must be written as `\t` in the Python string. Verify with `go build ./...` after every write.

---

## Lesson 6: Extract MergedMigrationsFS into Shared Utility

**Category**: Architecture

**Problem**: The MergedMigrationsFS pattern (merging template migrations with domain migrations for golang-migrate compatibility) was duplicated ~330 lines across 4 service packages: jose-ja, skeleton-template, sm-im, pki-ca.

**Solution**: Extract to `internal/apps/template/service/server/repository/migrations_merged.go` as a shared utility. Each domain package now calls `NewMergedMigrationsFS(MigrationsFS)` instead of reimplementing the pattern from scratch.

**Impact**: 5 files changed, 330 duplicate lines removed.

**Note**: The `builder` package is DIFFERENT from the `repository` package. The shared utility lives in `repository/`, not `builder/`.

---

## Lesson 7: Always Run go build After File Writes

**Category**: Workflow

**Problem**: File writes can silently corrupt Go source files (tab stripping, encoding issues, truncation). Without a build check, broken files are committed.

**Solution**: After EVERY file write in a Go package, immediately run:
```bash
go build ./path/to/package/...
```

If the build fails, check: (1) Did tabs get stripped? (2) Is the file encoding correct? (3) Are braces balanced?

---

## Lesson 8: All Magic Values in internal/shared/magic/

**Category**: Linting

**Problem**: Magic values (string literals, numeric constants) scattered across packages create maintenance burden and magic_usage linter violations.

**Solution**: ALL magic constants MUST be declared in `internal/shared/magic/magic_*.go` files. No package-local magic files. The mnd linter enforces this.

**Key Points**:
- File naming: `magic_<domain>.go` (e.g., `magic_network.go`, `magic_cicd.go`)
- The `internal/shared/magic/` package is excluded from coverage and mutation targets (constants only, no executable logic)
- Never duplicate a constant - if it exists in magic/, use `cryptoutilSharedMagic.ConstantName`
