# coverage-analysis

Analyze Go test coverage profiles to identify gaps and generate targeted test suggestions.

## Purpose

Use after running `go test -coverprofile` to systematically categorize uncovered
lines and prioritize which tests to write.

## Workflow

```bash
# 1. Generate coverage profile
mkdir -p test-output/coverage-analysis/
go test -coverprofile=test-output/coverage-analysis/coverage.out ./...

# 2. Generate HTML report
go tool cover -html=test-output/coverage-analysis/coverage.out -o test-output/coverage-analysis/coverage.html

# 3. Show function-level breakdown
go tool cover -func=test-output/coverage-analysis/coverage.out | sort -k3 -n | head -30

# 4. Show total
go tool cover -func=test-output/coverage-analysis/coverage.out | tail -1
```

## Coverage Targets

| Package Type | Minimum | Examples |
|--------------|---------|---------|
| Production | 95% | internal/{jose,identity,kms,ca} |
| Infrastructure/Utility | 98% | internal/apps/cicd/*, internal/shared/*, pkg/* |
| Generated Code | Excluded | api/*_gen.go |
| Magic Constants | Excluded | internal/shared/magic/ |

## Gap Categories

When analyzing uncovered (RED) lines:

1. **Error paths** — `if err != nil { ... }` branches not exercised
2. **Edge cases** — nil input, empty slice, boundary values
3. **Unreachable** — structural barriers (os.Exit, shutdown hooks)
4. **Coverage ceiling** — structurally unreachable, document exception

## Test Seam Pattern (for unreachable paths)

```go
// Production code
var osExit = os.Exit  // seam replaceable in tests

// Test code
func TestShutdownError(t *testing.T) {
    orig := osExit
    defer func() { osExit = orig }()
    var code int
    osExit = func(c int) { code = c }
    // trigger shutdown path
    require.Equal(t, 1, code)
}
```

## References

See [ARCHITECTURE.md Section 10.2.3 Coverage Targets](../../docs/ARCHITECTURE.md#1023-coverage-targets) for per-package targets.
See [ARCHITECTURE.md Section 10.2.4 Test Seam Injection Pattern](../../docs/ARCHITECTURE.md#1024-test-seam-injection-pattern) for unreachable code.
