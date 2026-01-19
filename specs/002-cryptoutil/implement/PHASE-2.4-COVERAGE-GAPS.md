# Phase 2.4: JOSE-JA Coverage Gaps

## Status: IN PROGRESS

## Current Coverage Analysis

From test run output:

```
cryptoutil/internal/apps/jose/ja/domain                  coverage: 100.0% ✅
cryptoutil/internal/apps/jose/ja/repository              coverage: 81.6%  ❌ (target: ≥98%)
cryptoutil/internal/apps/jose/ja/server                  coverage: 0.0%   ✅ (no code)
cryptoutil/internal/apps/jose/ja/server/apis             coverage: 100.0% ✅
cryptoutil/internal/apps/jose/ja/server/config           coverage: 61.9%  ❌ (target: ≥98%)
```

## Coverage Gaps to Address

### Priority 1: Repository Layer (81.6% → 98%)

**Gap**: Missing 16.4% coverage on infrastructure code
**Target**: ≥98% (infrastructure code requirement)

**Likely uncovered areas**:
- Error handling edge cases
- Database constraint violations (unique, foreign key)
- Transaction rollback scenarios
- Empty result set edge cases
- Nil parameter handling

**Action**: Generate HTML coverage report to identify exact uncovered lines
```bash
go test ./internal/apps/jose/ja/repository/... -coverprofile=coverage_repo.out
go tool cover -html=coverage_repo.out -o coverage_repo.html
```

### Priority 2: Config Layer (61.9% → 98%)

**Gap**: Missing 36.1% coverage on configuration validation
**Target**: ≥98% (infrastructure code requirement)

**Likely uncovered areas**:
- Missing config field validation
- Invalid value ranges not tested
- Default value application logic
- Config file loading errors
- Environment variable parsing

**Action**: Review config validation logic and add missing test cases

## Next Steps

1. Generate HTML coverage reports for repository and config packages
2. Identify exact uncovered lines (RED in HTML report)
3. Write targeted tests for uncovered lines only
4. Re-run coverage until ≥98% achieved
5. Run mutation testing (gremlins) - target ≥98%
6. Update git commit with final coverage evidence

## Evidence Requirements

- [ ] Repository coverage ≥98%
- [ ] Config coverage ≥98%
- [ ] Mutation score ≥98%
- [ ] All tests passing
- [ ] Git commit with evidence

## Time Estimate

- Coverage gap analysis: 15 minutes
- Write missing tests: 30-45 minutes
- Mutation testing: 30-60 minutes
- **Total**: 1-2 hours
