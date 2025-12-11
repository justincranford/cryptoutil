# Gremlins Mutation Testing - ARCHIVED

## Status: RESOLVED - Gremlins working in CI/CD

**Archived Date**: 2025-12-11
**Resolution**: Gremlins v0.6.0 works successfully in ci-mutation workflow
**Evidence**: Run 20121960342 completed mutation testing (742s runtime)
**Action**: Refer to specs/001-cryptoutil/MUTATION-TESTING-BASELINE.md for current results

---

## Historical Issue (Now Resolved)

### Issue

Gremlins v0.6.0 crashes with panic "error, this is temporary" in executor.go:165
This is a known issue with the gremlins mutation testing tool.

### Installation Status

✅ Gremlins installed successfully via `go install github.com/go-gremlins/gremlins/cmd/gremlins@latest`

✅ Configuration exists at `.gremlins.yaml` with reasonable defaults:

- Threshold: 70% efficacy, 60% mutant coverage
- Workers: 2, Test CPU: 1, Timeout coefficient: 3
- Operators: arithmetic, conditionals, increment-decrement, invert-negatives
- Excludes: generated code, mocks, vendor, external dependencies

❌ Execution fails on all packages tested

### Test Command Used

```powershell
gremlins unleash --tags=!integration ./internal/common/magic
```

### Error Output

```
Starting...
Gathering coverage... done in 4.2946832s
panic: error, this is temporary

goroutine 21 [running]:
github.com/go-gremlins/gremlins/internal/engine.(*mutantExecutor).Start(0xc000282000, 0xc0000ae400)
        R:/temp/go-mod-cache/github.com/go-gremlins/gremlins@v0.6.0/internal/engine/executor.go:165 +0x55e
```

### Resolution Options

1. **Wait for upstream fix**: Monitor <https://github.com/go-gremlins/gremlins/issues>
2. **Try alternative tool**: Consider <https://github.com/zimmski/go-mutesting>
3. **Manual mutation testing**: Create mutation test cases manually for critical paths
4. **Defer requirement**: Update constitution to make mutation testing "recommended" not "mandatory" until tooling stabilizes

### Constitution Impact

Constitution v2.0.0 requires:

- ≥80% gremlins score per package (mandatory)

**RECOMMENDATION**: Add amendment to Constitution v2.0.1 making mutation testing "recommended" until tooling is stable.

### Workaround for Iteration 2

Since mutation testing is blocked:

1. Document this limitation in specs/002-cryptoutil/CLARIFICATIONS.md
2. Add to EXECUTIVE-SUMMARY.md known issues
3. Focus on achieving 95/100/100 coverage targets with comprehensive unit tests
4. Manual mutation testing for critical crypto operations (test with wrong keys, corrupted ciphertext, etc.)

### Next Steps

- [ ] File issue at <https://github.com/go-gremlins/gremlins/issues>
- [ ] Research alternative mutation testing tools (go-mutesting, etc.)
- [ ] Propose constitution amendment v2.0.1 to make mutation testing recommended not mandatory
- [ ] Document workaround approach for iteration 2/3
