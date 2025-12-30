# Known Issues

## Test Flakiness: Full Suite Timeout (2025-12-30)

**Status**: Documented, workaround identified

**Description**:
When running the entire `./internal/learn/server` test suite, `TestHandleLoginUser_EmptyCredentials` hangs for ~10 minutes until timeout. However, the test passes when run in isolation.

**Evidence**:
- Full suite run: Test hangs at HTTP client request, never completes (timeout after 10m0s)
- Isolated run: `go test -run TestHandleLoginUser_EmptyCredentials -timeout 60s` → PASS in 0.20s
- Output files: `test-output/phase_0.3_baseline.txt` (full suite), `test-output/isolated_test.txt` (isolated)

**Root Cause Hypothesis**:
Test interaction issue - likely one of:
1. Resource exhaustion: Too many servers/goroutines created simultaneously across 99 test cases
2. Cleanup races: Server shutdown while other tests connecting
3. SQLite contention: WAL mode with heavy concurrent load
4. HTTP client/TLS handshake deadlock under high concurrency

**Workaround**:
Run tests in smaller batches or with reduced parallelism:
```bash
# Run by file (stable)
go test ./internal/learn/server -run TestHandleLoginUser -v

# Run with limited parallelism
go test ./internal/learn/server -v -parallel=4

# Run specific packages
go test ./internal/learn/server -run "TestHandle(Login|Register)" -v
```

**Impact**:
- CI/CD workflows: May need to split test execution into multiple jobs
- Local development: Run specific test files/patterns instead of full suite
- Coverage reporting: Still accurate (individual tests pass)

**Next Steps**:
1. ✅ Document issue (this file)
2. ⏸️ Proceed with Phase 0.3 refactoring (organization, not test fixes)
3. 📋 Create Phase 0.4 task: "Fix test suite flakiness and resource management"
4. 📋 Investigate: TestMain resource lifecycle, server cleanup, connection pooling
5. 📋 Consider: Shared server pattern instead of per-test servers

**Related**:
- Phase 0.3: Refactor internal/learn/server/ files (safe to proceed)
- SERVICE-TEMPLATE.md: Test quality standards require stable execution
- test-output/phase_0.3_baseline.txt: Full diagnostic goroutine dump

**Last Updated**: 2025-12-30
