# Cryptoutil Testing Infrastructure TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 17, 2025
**Status**: Testing infrastructure improvements completed - fuzz and benchmark testing implemented for cryptographic operations. Test file organization audit and migration completed.

---

## 🟡 MEDIUM - Testing Infrastructure Improvements

### Task T4: Implement Coverage Trend Analysis
- **Description**: Add coverage trend analysis to CI workflow to track coverage changes over time
- **Current State**: Basic coverage collection implemented, trend analysis not yet added
- **Proposed Implementation**:
  - Calculate current coverage percentage from `go tool cover -func` output
  - Download previous run's coverage data from artifacts
  - Compare current vs previous coverage and calculate difference
  - Display trend indicators (📈 increased, 📉 decreased, ➡️ unchanged, 📊 baseline)
  - Store current coverage for next run comparison
  - Show trend in GitHub Actions summary with visual indicators
- **Files**: `.github/workflows/ci-coverage.yml`
- **Expected Outcome**: Track coverage improvements/declines over time, provide data for coverage decisions
- **Priority**: Medium - Testing metrics enhancement
- **Dependencies**: ci-coverage.yml workflow completion

```

### Implementation Priority Recommendations

1. **High Priority**: External unit tests (`*_test.go`) - Establish API contracts
2. **Medium Priority**: Internal unit tests (`*_internal_test.go`) - Cover complex internals
3. **Medium Priority**: Integration tests (`*_integration_test.go`) - Validate real dependencies
4. **Low Priority**: Benchmarks (`*_bench_test.go`) - Performance optimization
5. **Low Priority**: Fuzz tests (`*_fuzz_test.go`) - Advanced property testing
6. **Optional**: E2E tests (`e2e/`) - Full system validation

### Current Project Assessment

- **Existing**: Mix of internal/external test patterns
- **testpackage linter**: Currently configured to allow internal testing
- **Recommendation**: Gradually migrate toward external testing for better API design
