# Mutation Testing Analysis and Fixes

## Overview

This document analyzes mutation testing results from the CI - Mutation Testing workflow and provides prioritized tasks to address mutation testing failures and performance issues.

**Workflow**: [CI - Mutation Testing](https://github.com/justincranford/cryptoutil/actions/runs/20085840091/job/57626095356?pr=10)
**Status**: In Progress (Running for 33+ minutes as of analysis)
**Tool**: Gremlins v0.6.0
**Timeout**: 45 minutes configured
**Target Thresholds**: 70% efficacy, 60% mutant coverage (per .gremlins.yaml)

---

## Known Issues

### 1. **CRITICAL**: Excessive Runtime (33+ minutes and counting)

**Impact**: HIGH - Blocks CI/CD pipeline, wastes compute resources
**Symptom**: Mutation testing step has been running for 33+ minutes and is still in progress
**Expected Runtime**: 5-15 minutes for reasonable test suite
**Current Timeout**: 45 minutes

**Root Causes**:

- Testing ALL packages without package-level parallelization
- Gremlins may be running with sequential mutation execution
- Large number of mutants across entire codebase
- No incremental or selective mutation testing

**Proposed Solutions** (Priority Order):

#### High Priority

1. **Enable package-level parallelization**
   - Run gremlins on individual packages concurrently using Go workflow matrix
   - Example: Split into `crypto`, `server`, `repository`, `identity` jobs
   - Reduces total time from sequential to parallel (4x-8x speedup potential)

2. **Reduce mutation scope for CI**
   - Focus on critical paths: `internal/kms/server/barrier`, `internal/identity/oauth`, crypto packages
   - Run full mutation suite on schedule/nightly, not on every PR
   - Use `--target` flag to limit mutation scope per run

3. **Optimize gremlins configuration**
   - Current: `workers: 2, test-cpu: 1, timeout-coefficient: 3`
   - Recommended: `workers: 4, test-cpu: 2, timeout-coefficient: 2`
   - Add `--dry-run` option to estimate mutation count before execution

#### Medium Priority

1. **Incremental mutation testing**
   - Only run mutations on changed packages (git diff detection)
   - Cache mutation results and only re-run affected mutants
   - Use `--diff` mode if gremlins supports it

2. **Smarter test selection**
   - Use `--test-pattern` to run only relevant tests for each mutation
   - Current config runs all tests for every mutation
   - Filter tests based on coverage data

### 2. **Tool Stability**: Known Gremlins Panic

**Impact**: MEDIUM - May cause random failures
**Symptom**: `panic: error, this is temporary` in executor.go:165
**Status**: Documented in `docs/todos-gremlins.md`

**Proposed Solutions**:

1. Monitor for panic errors in workflow logs
2. Add retry logic with backoff for transient failures
3. Evaluate alternative tools: `go-mutesting`, `go-mutate`
4. Consider making mutation testing "recommended" not "mandatory" until tooling stabilizes

### 3. **Configuration Inefficiency**: Disabled Mutation Operators

**Impact**: LOW - Missing mutation coverage
**Current State**: Only 5 operators enabled, 6 operators disabled

**Disabled Operators** (from .gremlins.yaml):

- `invert-assignments: false`
- `invert-bitwise: false`
- `invert-bwassign: false`
- `invert-logical: false`
- `invert-loopctrl: false`
- `remove-self-assignments: false`

**Proposed Solutions**:

1. Enable operators incrementally: Start with `invert-logical`, `invert-loopctrl`
2. Monitor impact on runtime and noise level
3. Document rationale for enabled/disabled operators

---

## Performance Optimization Recommendations

### Immediate Actions (This PR/Iteration)

1. **Split mutation testing by package** - Update workflow to use matrix strategy:

   ```yaml
   strategy:
     matrix:
       package:
         - ./internal/kms/server/barrier
         - ./internal/kms/server/businesslogic
         - ./internal/identity/oauth
         - ./pkg/crypto/keygen
     fail-fast: false
   steps:
     - name: Run mutation tests for ${{ matrix.package }}
       run: gremlins unleash ${{ matrix.package }} --tags=!integration
   ```

2. **Add timeout per package** - Individual package timeout of 10 minutes instead of 45 minutes total

3. **Add early termination** - Stop if threshold not met on critical packages

### Short-term Actions (Next Sprint)

1. **Implement differential mutation testing**
   - Script to detect changed packages via `git diff`
   - Only run mutations on packages with code changes
   - Full suite on `main` branch, incremental on PRs

2. **Add mutation result caching**
   - Cache gremlins results per package + commit SHA
   - Only re-run if package code changes
   - GitHub Actions cache or artifact reuse

### Long-term Actions (Future Iterations)

1. **Evaluate alternative tools**
   - Benchmark `go-mutesting` vs `gremlins`
   - Consider language-agnostic mutation tools (universalmutator, mutatest)
   - Document trade-offs and migration path

2. **Integrate with coverage reports**
   - Link mutation kills to test coverage
   - Identify code covered but not mutation-tested
   - Prioritize mutations for low-coverage areas

---

## Mutation Testing Best Practices

### Package Selection Strategy

**Critical Packages** (MUST have ≥80% mutation score):

- `internal/kms/server/barrier` - Key unsealing and encryption
- `internal/identity/oauth` - OAuth2/OIDC authentication
- `pkg/crypto/keygen` - Cryptographic key generation
- `internal/kms/server/businesslogic` - Business logic validation

**High Priority Packages** (Target ≥70%):

- `internal/kms/server/repository` - Database persistence
- `internal/identity/session` - Session management
- `internal/common/config` - Configuration parsing

**Medium Priority Packages** (Target ≥60%):

- `internal/kms/server/application` - Application initialization
- `internal/identity/middleware` - Authentication middleware
- Test utilities and helpers

**Low Priority / Excluded**:

- Generated code (`api/`, `*_gen.go`)
- Mock implementations
- Vendor dependencies
- Integration test helpers

### Configuration Tuning Guidelines

**For Fast Feedback** (PR checks):

```yaml
workers: 4
test-cpu: 2
timeout-coefficient: 2
threshold-efficacy: 70.0
threshold-mcover: 60.0
```

**For Comprehensive Analysis** (Nightly/Weekly):

```yaml
workers: 2
test-cpu: 1
timeout-coefficient: 3
threshold-efficacy: 80.0
threshold-mcover: 70.0
```

---

## Action Items (Prioritized)

### P0 - Critical (Block CI/CD)

- [X] Create this analysis document
- [ ] Implement package-level matrix strategy in ci-mutation.yml
- [ ] Add per-package 10-minute timeout
- [ ] Test matrix approach with 2-3 critical packages first

### P1 - High (Performance)

- [ ] Profile gremlins execution to identify bottlenecks
- [ ] Add differential mutation testing script
- [ ] Configure mutation result caching
- [ ] Document expected runtime per package

### P2 - Medium (Quality)

- [ ] Enable `invert-logical` and `invert-loopctrl` operators
- [ ] Review mutation survivors and add missing test cases
- [ ] Create package-level mutation score tracking
- [ ] Add mutation testing metrics to docs/

### P3 - Low (Future Enhancements)

- [ ] Evaluate alternative mutation testing tools
- [ ] Research mutation testing visualization tools
- [ ] Consider commercial mutation testing services (Stryker, PIT)

---

## Questions for Stakeholder

1. **Timeout Tolerance**: Is 45 minutes acceptable for mutation testing in CI, or should we aim for <10 minutes?
2. **Scope Trade-off**: Should we run full mutation suite on every PR, or only on critical packages?
3. **Tool Stability**: Given gremlins instability, should we defer mutation testing requirement until tooling improves?
4. **Quality vs Speed**: Prefer comprehensive mutation coverage (slow) or fast feedback (limited scope)?

---

## Monitoring and Metrics

### Success Criteria

- Mutation testing completes in <15 minutes per PR
- Zero panics/crashes from gremlins tool
- ≥70% efficacy, ≥60% mutant coverage achieved
- No CI/CD pipeline blocking due to mutation testing

### Metrics to Track

- Runtime per package (target: <5 minutes each)
- Number of mutants generated/killed/survived
- Mutation score trend over time
- Test execution time with/without mutations

---

## References

- Gremlins Documentation: <https://gremlins.dev/>
- Gremlins GitHub: <https://github.com/go-gremlins/gremlins>
- Mutation Testing Best Practices: <https://mutation-testing.com/>
- Go Mutation Testing Alternatives: go-mutesting, go-mutate
- Current Configuration: `.gremlins.yaml`
- Known Issues: `docs/todos-gremlins.md`

---

**Last Updated**: 2025-12-10
**Next Review**: After mutation testing workflow completes
