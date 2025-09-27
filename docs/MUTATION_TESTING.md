# Mutation Testing with Gremlins

This project uses [Gremlins](https://github.com/go-gremlins/gremlins) for mutation testing to validate the quality and effectiveness of our test suite.

## What is Mutation Testing?

Mutation testing works by introducing small changes (mutations) to your code and verifying that your test suite catches these changes. If a test suite fails when a mutation is introduced, the mutation is "killed" (good). If the test suite still passes, the mutation "lived" (indicates a potential gap in testing).

## How It Works

1. **Coverage Analysis**: Gremlins first analyzes which parts of code are covered by tests
2. **Mutation Generation**: Creates mutations only in covered code (no point testing uncovered code)
3. **Test Execution**: Runs your test suite against each mutation
4. **Result Classification**:
   - **KILLED**: Test suite detected the mutation (✅ good)
   - **LIVED**: Test suite didn't detect the mutation (⚠️ potential issue)  
   - **TIMED OUT**: Tests took too long (often means mutation was caught but slowly)
   - **NOT COVERED**: Code not covered by tests (skipped)
   - **NOT VIABLE**: Mutation broke compilation

## Configuration

Mutation testing is configured in `.gremlins.yaml`:

```yaml
# Quality thresholds
threshold-efficacy: 70.0    # Min % of killed mutants
threshold-mcover: 60.0      # Min % of mutant coverage

# Performance settings  
workers: 2                  # Parallel workers
timeout-coefficient: 3      # Timeout multiplier

# Enabled mutation operators
arithmetic-base: true       # +, -, *, /, %
conditionals-boundary: true # <, <=, >, >=  
conditionals-negation: true # Invert conditions
increment-decrement: true   # ++, --
invert-negatives: true      # Positive/negative values
```

## Running Mutation Testing

### Prerequisites

```bash
# Install Gremlins
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
```

### Manual Execution

```bash
# Dry run (analyze without testing)
gremlins unleash --dry-run ./internal/common/util/datetime/

# Full mutation testing on a specific package
gremlins unleash ./internal/common/util/datetime/

# Test multiple packages with custom settings
gremlins unleash ./internal/common/util/... --workers 2 --timeout-coefficient 3
```

### Automated Execution

Mutation testing runs automatically in GitHub Actions on the `main` branch after all tests pass. It focuses on packages with high test coverage:

- `./internal/common/util/datetime/` (100% coverage)
- `./internal/common/util/thread/` (100% coverage)  
- `./internal/common/util/sysinfo/` (84.9% coverage)

## Interpreting Results

### Good Results
```
Killed: 45, Lived: 2, Not covered: 5
Timed out: 3, Not viable: 0, Skipped: 0
Test efficacy: 90.43%
Mutator coverage: 91.67%
```

- **High kill rate** (45/47 = 95.7% of testable mutations killed)
- **Good efficacy** (90.43% > 70% threshold)
- **Acceptable coverage** (91.67% > 60% threshold)

### Problem Indicators
```
Killed: 12, Lived: 18, Not covered: 20
Test efficacy: 40.00%
```

- **High lived mutations** (18 survived) → Test gaps
- **Low efficacy** (40% < 70% threshold) → Need better assertions
- **High not covered** → Need more test coverage

## Common Mutation Types

| Type | Description | Example |
|------|-------------|---------|
| `ARITHMETIC_BASE` | Change math operators | `a + b` → `a - b` |
| `CONDITIONALS_BOUNDARY` | Change comparison operators | `x < 10` → `x <= 10` |  
| `CONDITIONALS_NEGATION` | Invert boolean conditions | `if (valid)` → `if (!valid)` |
| `INCREMENT_DECREMENT` | Change ++/-- operators | `i++` → `i--` |
| `INVERT_NEGATIVES` | Change positive/negative | `return -value` → `return value` |

## Best Practices

### Focus on High-Value Areas
- Run on packages with high test coverage first
- Prioritize business logic over utility functions
- Target complex algorithms and edge cases

### Interpreting LIVED Mutations
When mutations survive:
1. **Missing test cases** - Add tests for the specific condition
2. **Weak assertions** - Use more specific assertions instead of just checking "no error"
3. **Dead code** - Remove unreachable or redundant code
4. **False positives** - Some mutations may be semantically equivalent

### Performance Optimization
- Use `--workers 2` for parallel execution
- Increase `--timeout-coefficient` if getting false timeouts
- Use `--dry-run` for quick analysis
- Focus on specific packages rather than entire codebase

## Troubleshooting

### Common Issues

**Timeouts on Windows**
```
TIMED OUT mutations
```
- Increase `timeout-coefficient` in `.gremlins.yaml`
- Reduce `workers` count
- Windows file system can be slower than Linux

**File Permission Errors**
```
ERROR: impossible to remove temporary folder
```
- This is cosmetic - results are still valid
- Caused by Windows file locking
- Doesn't affect mutation testing functionality

**No Mutations Found**
```
Runnable: 0, Not covered: 100
```
- Package has no test coverage
- Add unit tests first before mutation testing

## Integration with CI/CD

Mutation testing runs automatically on:
- Push to `main` branch
- After all other CI jobs pass
- Results are uploaded as artifacts
- Summary appears in GitHub Action summary

The CI focuses on high-coverage packages to provide valuable feedback without excessive build times.

## Further Reading

- [Gremlins Documentation](https://gremlins.dev/)
- [Mutation Testing Introduction](https://pedrorijo.com/blog/intro-mutation/)
- [PITest (Java mutation testing inspiration)](https://pitest.org/)
