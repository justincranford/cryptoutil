---
name: coverage-analysis
description: "Analyze Go test coverage profiles to identify gaps and generate targeted test suggestions. Use after running go test -coverprofile to systematically categorize uncovered lines, identify error paths and seam injection opportunities, and prioritize which tests to write."
argument-hint: "[./internal/... or package path]"
---

Analyze test coverage gaps from a coverprofile and suggest specific tests to add.

**Full Copilot original**: [.github/skills/coverage-analysis/SKILL.md](.github/skills/coverage-analysis/SKILL.md)

## Key Rules

- Store ALL coverage artifacts in `test-output/coverage-analysis/` (never project root)
- Target ≥95% for production code, ≥98% for infrastructure/utility code
- Focus on RED lines in HTML report (uncovered code), not green
- Categorize uncovered lines: error paths, shutdown hooks, external integrations
- Document coverage ceiling analysis for structural barriers (ceiling − 2% buffer)
- `internal/shared/magic/` excluded (constants only, no executable logic)

## Coverage Targets

- Production code: **≥95%**
- Infrastructure/utility code: **≥98%**

## Usage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out -covermode=atomic ./...

# View HTML report
go tool cover -html=coverage.out -o coverage.html

# Check specific package
go test -coverprofile=pkg.out -covermode=atomic ./internal/apps/sm-kms/...
go tool cover -func=pkg.out | grep -v 100.0
```

Then provide the output and ask: "Analyze this coverage report for gaps."

## Gap Categories

When analyzing, categorize uncovered lines:

1. **Error paths** — `if err != nil` blocks not hit by tests; add error injection tests
2. **Edge cases** — boundary conditions, empty inputs, nil inputs
3. **Third-party boundary** — external library error returns; use injectable functions
4. **Unreachable** — code that genuinely cannot be reached; apply Test Seam Pattern
5. **Coverage ceiling** — infrastructure code (init, main) that cannot be unit tested; exclude via `//nolint:gocover`

## Test Seam Pattern (for unreachable paths)

```go
// Production code — injectable for testing:
var sqlOpenFn = sql.Open

func initDB(dsn string) (*sql.DB, error) {
    return sqlOpenFn("sqlite", dsn)
}
```

```go
// Test code:
func TestInitDB_OpenError(t *testing.T) {
    t.Parallel()
    original := sqlOpenFn
    sqlOpenFn = func(_, _ string) (*sql.DB, error) {
        return nil, errors.New("injected error")
    }
    t.Cleanup(func() { sqlOpenFn = original })

    _, err := initDB(":memory:")
    require.Error(t, err)
}
```

## Common Pitfalls

- Timeout double-multiplication: `timeout * time.Second` when timeout is already a duration
- Missing `DisableKeepAlives: true` in HTTP clients for test cleanup
- Probability-based tests: wrap expensive variants with `if testing.Short() { t.Skip() }`
