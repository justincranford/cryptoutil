---
name: coverage-analysis
description: "Analyze Go test coverage profiles to identify gaps and generate targeted test suggestions. Use after running go test -coverprofile to systematically categorize uncovered lines, identify error paths and seam injection opportunities, and prioritize which tests to write."
argument-hint: "[./internal/... or package path]"
metadata:
  domain: testing
---

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
3. **Third-party boundary** — library return errors that require internal library state manipulation (e.g. `jwk.Import`, `jwk.Set.Keys()` iterator errors)
4. **Unreachable** — structural barriers (os.Exit, shutdown hooks, exhaustive type switches)
5. **Coverage ceiling** — structurally unreachable; document exception with justification

**Ceiling formula**: `ceiling = (total_lines - unreachable_lines) / total_lines`. Set package target at `ceiling - 2%` (safety margin).

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

For packages with multiple seams, use a `saveRestoreSeams(t)` helper to save and restore all seam variables automatically via `t.Cleanup()`, preventing test pollution across parallel tests.

## Probability-Based Execution (when suggesting tests to write)

For expensive algorithm variant tests (RSA sizes, ECDSA curves, AES key sizes), apply probability gates to avoid running all variants on every test run:

| Gate | `TestProb` value | Use cases |
|------|-----------------|-----------|
| Always | 100 | Base algorithms (RSA-2048, AES-128, P-256) |
| Quarter | 25 | Important variants (RSA-3072, P-384) |
| Tenth | 10 | Redundant variants (RSA-4096, P-521) |

Apply this when coverage analysis reveals uncovered algorithm branches: add the appropriate probability gate rather than testing all variants unconditionally.

## References

Read [ARCHITECTURE.md Section 10.2.3 Coverage Targets](../../../docs/ARCHITECTURE.md#1023-coverage-targets) for per-package targets — apply these targets when categorizing uncovered lines and setting package-specific coverage ceiling exceptions.
Read [ARCHITECTURE.md Section 10.2.4 Test Seam Injection Pattern](../../../docs/ARCHITECTURE.md#1024-test-seam-injection-pattern) for unreachable code — use the seam injection pattern when suggesting how to cover structurally unreachable lines.

Read [ARCHITECTURE.md Section 10.2 Unit Testing Strategy](../../../docs/ARCHITECTURE.md#102-unit-testing-strategy) for probability-based test execution — when coverage gaps are in algorithm variant branches, apply `TestProbAlways/Quarter/Tenth` gates rather than unconstrained variant testing.
