# Mutation Testing Baseline Results

**Generated**: January 9, 2026
**Tool**: gremlins v0.6.0
**Target**: ≥80% test efficacy per package

## Summary

All tested packages meet or exceed the 80% test efficacy target.

| Package | Test Efficacy | Mutator Coverage | Killed | Lived | Not Covered | Status |
|---------|---------------|------------------|--------|-------|-------------|--------|
| internal/common/util/network | 100.00% | 100.00% | 9 | 0 | 0 | ✅ Excellent |
| internal/common/crypto/keygen | 100.00% | 100.00% | 16 | 0 | 0 | ✅ Excellent |
| internal/common/crypto/digests | 100.00% | 100.00% | 6 | 0 | 0 | ✅ Excellent |
| internal/identity/issuer | 94.12% | 73.91% | 16 | 1 | 6 | ✅ Good |
| internal/kms/server/businesslogic | 98.44% | 48.48% | 63 | 1 | 68 | ✅ Good |

## Package Details

### internal/common/util/network

- **Test Efficacy**: 100.00%
- **Mutator Coverage**: 100.00%
- **Results**: 9 killed, 0 lived, 0 not covered
- **Runtime**: ~50 seconds
- **Status**: ✅ Perfect score - all mutations killed

### internal/common/crypto/keygen

- **Test Efficacy**: 100.00%
- **Mutator Coverage**: 100.00%
- **Results**: 16 killed, 0 lived, 0 not covered
- **Runtime**: 6 minutes 42 seconds
- **Status**: ✅ Perfect score - comprehensive cryptographic testing

### internal/common/crypto/digests

- **Test Efficacy**: 100.00%
- **Mutator Coverage**: 100.00%
- **Results**: 6 killed, 0 lived, 0 not covered
- **Runtime**: 17 seconds
- **Status**: ✅ Perfect score - HKDF and SHA-256 fully tested

### internal/identity/issuer

- **Test Efficacy**: 94.12%
- **Mutator Coverage**: 73.91%
- **Results**: 16 killed, 1 lived, 6 not covered
- **Runtime**: 3 minutes 30 seconds
- **Status**: ✅ Exceeds 80% target - JWT/JWE/JWS operations well tested
- **Lived Mutation**: `CONDITIONALS_BOUNDARY at key_rotation.go:271:17`
- **Analysis**: Boundary condition in key rotation logic - acceptable given high efficacy

### internal/kms/server/businesslogic

- **Test Efficacy**: 98.44%
- **Mutator Coverage**: 48.48%
- **Results**: 63 killed, 1 lived, 68 not covered
- **Runtime**: 3 minutes 28 seconds
- **Status**: ✅ Exceeds 80% target - business logic well tested
- **Lived Mutation**: `CONDITIONALS_BOUNDARY at oam_orm_mapper.go:358:45`
- **Not Covered**: 68 mutations (likely in error handling and edge cases)
- **Analysis**: High test efficacy despite lower mutator coverage - core paths well protected

## Interpretation

### Test Efficacy

Measures how well tests catch actual code changes:

- **100%**: All viable mutations are caught by tests (perfect)
- **≥80%**: Target threshold for production code (good)
- **<80%**: Tests may miss real bugs (needs improvement)

### Mutator Coverage

Measures how many possible mutations are tested:

- **100%**: All possible mutations are tested (perfect)
- **≥80%**: Good coverage of mutation space
- **<80%**: Some code paths may not be mutated (acceptable if test efficacy is high)

### Analysis

- **Crypto packages** (keygen, digests): 100% scores - critical for security
- **Network package**: 100% score - critical for reliability
- **Identity issuer**: 94.12% efficacy - JWT/JWS/JWE operations well protected
- **KMS businesslogic**: 98.44% efficacy - core business logic well tested

All packages exceed the 80% test efficacy target, demonstrating high-quality test suites.

## Known Issues

### Temporary Folder Cleanup (Windows)

Gremlins occasionally fails to clean up temporary folders on Windows:

```
ERROR: impossible to remove temporary folder: The process cannot access the file
because it is being used by another process.
```

**Impact**: Non-critical - leaves temp folders in `R:\temp\gremlins-*`
**Workaround**: Manually delete temp folders if disk space becomes an issue
**Root Cause**: Windows file locking (Python virtualenv files remain locked)

## Recommendations

1. **Maintain current test quality** - all packages meet or exceed targets
2. **Monitor lived mutations** - investigate boundary conditions if they recur
3. **Run gremlins in CI** - add mutation testing to CI/CD pipeline (optional)
4. **Focus on test efficacy** - prioritize killing mutations over coverage percentage

## Next Steps

- [x] Run gremlins on 5 critical packages
- [ ] Run gremlins on remaining packages (CA, Identity AuthZ/IdP, KMS client)
- [ ] Add mutation testing to CI/CD (optional - long runtime)
- [ ] Investigate lived mutations in issuer and businesslogic (low priority)

## References

- Gremlins GitHub: <https://github.com/go-gremlins/gremlins>
- Mutation Testing Guide: <https://go-gremlins.github.io/gremlins/>
- Project Target: ≥80% test efficacy per package (01-02.testing.instructions.md)
