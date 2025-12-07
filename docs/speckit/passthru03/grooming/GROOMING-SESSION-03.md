# Grooming Session 03: CI/CD Reliability & Deferred Work Completion

## Overview
- **Focus Area**: CI/CD workflow reliability fixes and completion of deferred Iteration 2 work
- **Related Spec Section**: Infrastructure testing, CI/CD workflows, mutation testing
- **Prerequisites**: Understanding of Go testing patterns, GitHub Actions, Docker Compose, concurrent test execution

## Questions

### Q1: What is the CRITICAL rule about test concurrency in cryptoutil?
A) Tests must run with `-p=1` for deterministic results
B) Tests can run sequentially or concurrently based on preference
C) NEVER use `-p=1`, ALWAYS use concurrent execution with `-shuffle=on`
D) Concurrency only for integration tests, not unit tests

**Answer**: C
**Explanation**: Test concurrency is mandatory to achieve fastest execution and reveal production concurrency bugs. Sequential testing hides race conditions.

---

### Q2: Which workflow has a 100% failure rate that must be fixed in Phase 1?
A) ci-quality.yml
B) ci-race.yml
C) ci-coverage.yml
D) ci-benchmark.yml

**Answer**: B
**Explanation**: The ci-race.yml workflow has 100% failure rate due to DATA RACE in CA handler test at line 1502, requiring immediate attention.

---

### Q3: What is the correct test execution command for cryptoutil?
A) `go test ./... -p=1`
B) `go test ./... -cover -shuffle=on`
C) `go test ./... -parallel=1`
D) `go test ./... -race -p=1`

**Answer**: B
**Explanation**: Correct execution uses concurrent packages with shuffle for randomized test order to reveal timing-dependent bugs.

---

### Q4: What is the target coverage increase for Identity ORM in Iteration 3?
A) 67.5% → 85%
B) 67.5% → 90%
C) 67.5% → 95%
D) 67.5% → 100%

**Answer**: C
**Explanation**: Constitution requires ≥95% coverage for production code, so Identity ORM must reach 95% from current 67.5%.

---

### Q5: Which test failure is specifically identified in consent_decision_repository_test.go?
A) Race condition at line 150
B) Coverage gap at line 160
C) Test failure at line 160
D) Linting error at line 160

**Answer**: C
**Explanation**: The ITER3-003 task specifically identifies a test failure in consent_decision_repository_test.go at line 160 that needs fixing.

---

### Q6: What is the estimated total effort for Iteration 3?
A) ~20 hours (0.5 week sprint)
B) ~40 hours (1 week sprint)
C) ~60 hours (1.5 week sprint)
D) ~80 hours (2 week sprint)

**Answer**: B
**Explanation**: The plan estimates ~40 hours total effort across 4 phases, equivalent to a 1-week sprint.

---

### Q7: How many workflows currently fail out of the total CI/CD workflows?
A) 5 out of 11 workflows
B) 6 out of 11 workflows
C) 8 out of 11 workflows
D) 3 out of 11 workflows

**Answer**: C
**Explanation**: The plan identifies 8 failing workflows that need fixes to achieve 100% pass rate from current 27%.

---

### Q8: What is the preferred tool for running workflow tests locally?
A) `act` command directly
B) `cmd/workflow` tool for faster iteration
C) GitHub CLI (`gh`)
D) Docker Compose manual testing

**Answer**: B
**Explanation**: The plan recommends using `cmd/workflow` tool for fast local iteration and workflow debugging.

---

### Q9: Which testing methodology enhancement involves ≥80% score per package?
A) Fuzz testing coverage
B) Integration test coverage
C) Mutation testing (gremlins)
D) Benchmark test performance

**Answer**: C
**Explanation**: Mutation testing requires ≥80% gremlins score per package to validate test quality and effectiveness.

---

### Q10: What is the proper approach for fixing race conditions?
A) Use `-p=1` to avoid races entirely
B) Add `time.Sleep()` to fix timing issues
C) Fix with mutex/sync primitives while keeping `t.Parallel()`
D) Remove `t.Parallel()` from affected tests

**Answer**: C
**Explanation**: Race conditions must be fixed with proper synchronization primitives while maintaining concurrent test execution to verify thread safety.

---

### Q11: Which deferred feature from Iteration 2 involves RFC 6960 compliance?
A) EST Handler implementation
B) JOSE Docker integration
C) CA OCSP Handler
D) Unified E2E test suite

**Answer**: C
**Explanation**: CA OCSP Handler implementation requires RFC 6960 compliance for certificate status checking functionality.

---

### Q12: What is the correct approach for E2E test debugging?
A) Simplify tests to avoid complexity
B) Add diagnostic logging + health checks + retries with exponential backoff
C) Run tests sequentially to avoid timing issues
D) Mock external dependencies

**Answer**: B
**Explanation**: E2E debugging requires comprehensive diagnostic logging, health checks, and resilient retry mechanisms for service startup timing.

---

### Q13: How should test data isolation be achieved in concurrent tests?
A) Use hardcoded test data values
B) Use UUIDv7 for unique test data + dynamic ports + TestMain pattern
C) Use separate test databases per test
D) Use sequential test execution only

**Answer**: B
**Explanation**: Test data isolation requires UUIDv7 for uniqueness, dynamic port allocation, and TestMain for shared dependencies.

---

### Q14: What is the primary goal of Phase 1 in Iteration 3?
A) Complete all deferred features
B) Fix critical CI/CD workflow failures blocking pipeline
C) Enhance test methodologies
D) Update documentation

**Answer**: B
**Explanation**: Phase 1 focuses on fixing 5 critical workflow failures that are blocking the CI/CD pipeline with 100% failure rates.

---

### Q15: Which GitHub workflow command should be used for DAST scanning?
A) `go run ./cmd/workflow -workflows=dast`
B) `go run ./cmd/workflow -workflows=dast -inputs="scan_profile=quick"`
C) `act -j dast`
D) `gh workflow run dast.yml`

**Answer**: B
**Explanation**: The workflow tool requires specific input parameters like scan_profile for DAST workflow execution.

---

### Q16: What is the correct approach for handling slow test packages?
A) Skip slow tests in CI/CD
B) Use selective test execution for local dev, full suite for CI/CD
C) Always run full test suite regardless of speed
D) Split slow packages into separate repositories

**Answer**: B
**Explanation**: The plan balances development speed with CI/CD completeness by using selective execution locally and full suites in CI.

---

### Q17: Which file documents the slow test packages requiring special handling?
A) SLOW-TESTS.md
B) SLOW-TEST-PACKAGES.md
C) docs/SLOW-PACKAGES.md
D) test-output/slow-tests.log

**Answer**: B
**Explanation**: The specs directory contains SLOW-TEST-PACKAGES.md documenting packages that require longer test execution times.

---

### Q18: What is the minimum mutation testing score required per package?
A) ≥70% gremlins score
B) ≥75% gremlins score
C) ≥80% gremlins score
D) ≥85% gremlins score

**Answer**: C
**Explanation**: Quality requirements mandate ≥80% gremlins score per package for mutation testing to ensure test effectiveness.

---

### Q19: Which Docker Compose issue is causing E2E/DAST/Load workflow failures?
A) Port conflicts between services
B) Service startup failures and timing issues
C) Memory limitations
D) Network connectivity problems

**Answer**: B
**Explanation**: The plan identifies Docker Compose startup failures and timing issues as the primary cause of E2E/DAST/Load workflow failures.

---

### Q20: What is the correct validation approach after fixing workflow issues?
A) Test only the fixed workflows
B) Run full workflow matrix locally before pushing
C) Rely on CI/CD to catch remaining issues
D) Test workflows individually in isolation

**Answer**: B
**Explanation**: Risk mitigation requires running the full workflow matrix locally to prevent regressions before pushing changes.

---

### Q21: Which phase focuses on JOSE Docker Integration completion?
A) Phase 1 (Critical CI/CD Fixes)
B) Phase 2 (Deferred Work Completion)
C) Phase 3 (Test Methodology Enhancements)
D) Phase 4 (Documentation Cleanup)

**Answer**: B
**Explanation**: Phase 2 focuses on completing deferred Iteration 2 features including JOSE Docker Integration.

---

### Q22: What is the estimated effort for EST Handler implementation?
A) 2 hours
B) 4 hours
C) 6 hours
D) 8 hours

**Answer**: B
**Explanation**: The plan allocates 4 hours for EST Handler implementation (RFC 7030) as part of deferred work completion.

---

### Q23: How should health checks be implemented for Docker Compose services?
A) HTTP health endpoints only
B) TCP connection checks only
C) Retries + exponential backoff + comprehensive health endpoints
D) No health checks needed

**Answer**: C
**Explanation**: Robust health check implementation requires retries, exponential backoff, and comprehensive health endpoint coverage.

---

### Q24: Which testing approach is preferred for real dependencies vs mocks?
A) Always use mocks for predictability
B) Always use real dependencies for authenticity
C) Real dependencies preferred, mocks only for hard-to-reach corner cases
D) Equal preference between mocks and real dependencies

**Answer**: C
**Explanation**: Testing instructions prefer real dependencies (test containers, in-memory services) with mocks reserved for truly external or hard-to-reach scenarios.

---

### Q25: What is the correct pattern for TestMain usage?
A) Start dependencies once per test function
B) Start dependencies once per package, reuse across tests
C) Start dependencies once per test case
D) Avoid TestMain, use setup/teardown per test

**Answer**: B
**Explanation**: TestMain pattern starts shared dependencies (PostgreSQL containers, services) once per package with all tests reusing resources.

---

### Q26: Which coverage targets apply to different code categories?
A) 90% for all code categories
B) 95% production, 95% infrastructure, 95% utility
C) 95% production, 100% infrastructure, 100% utility
D) 80% production, 90% infrastructure, 100% utility

**Answer**: C
**Explanation**: Quality requirements specify different targets: ≥95% production, ≥100% infrastructure (cicd), ≥100% utility code.

---

### Q27: What is the primary focus of Phase 3 (Test Methodology Enhancements)?
A) Fix failing tests
B) Add benchmarks, fuzz tests, property-based tests
C) Improve test performance
D) Reduce test complexity

**Answer**: B
**Explanation**: Phase 3 enhances test methodologies by adding comprehensive benchmark tests, fuzz tests, and property-based testing capabilities.

---

### Q28: How long should fuzz tests run at minimum?
A) 5 seconds per test
B) 10 seconds per test
C) 15 seconds per test
D) 30 seconds per test

**Answer**: C
**Explanation**: Testing instructions specify minimum 15 seconds fuzz time per test to achieve adequate coverage of input space.

---

### Q29: Which command pattern should be used for fuzz test execution?
A) `go test -fuzz=FuzzXXX -fuzztime=15s`
B) `go test -fuzz="FuzzXXX" -fuzztime=15s ./path`
C) `go test -fuzz=FuzzXXX -fuzztime=15s ./path`
D) `go test -fuzz FuzzXXX -fuzztime 15s ./path`

**Answer**: C
**Explanation**: Correct fuzz test execution uses unquoted names without quotes and must be run from project root with path specification.

---

### Q30: What is the primary purpose of property-based testing in cryptoutil?
A) Test performance characteristics
B) Test mathematical properties and invariants (e.g., encrypt(decrypt(x)) == x)
C) Test error handling
D) Test concurrency behavior

**Answer**: B
**Explanation**: Property-based testing validates mathematical properties and invariants, particularly important for cryptographic operations.

---

### Q31: Which library should be used for property-based testing?
A) testify/require
B) gofuzz
C) gopter
D) go-fuzz

**Answer**: C
**Explanation**: Testing instructions recommend gopter library for property-based testing with invariant validation.

---

### Q32: What is the correct approach for benchmark tests in cryptographic operations?
A) Unit benchmarks only
B) Happy path benchmarks only
C) Benchmarks for both happy and sad paths
D) Integration benchmarks only

**Answer**: C
**Explanation**: Cryptographic operations require benchmarks for both successful operations (happy path) and error conditions (sad path).

---

### Q33: How should test naming be handled for fuzz tests?
A) Any descriptive name
B) Names can be substrings of other test names
C) Fuzz test names MUST be unique, NOT substrings of others
D) Names must start with "Fuzz" prefix only

**Answer**: C
**Explanation**: Fuzz test names must be unique and cannot be substrings of other test names to avoid execution conflicts.

---

### Q34: What is the correct approach for handling DELETE-ME files in Phase 4?
A) Delete all DELETE-ME files immediately
B) Review content, extract valuable information, then delete
C) Rename DELETE-ME files to archive format
D) Leave DELETE-ME files for future reference

**Answer**: B
**Explanation**: Phase 4 documentation cleanup involves processing DELETE-ME files by extracting valuable content before deletion.

---

### Q35: Which diagnostic command helps with workflow failures?
A) `go test -v ./...`
B) `gh run list --status failure --limit 5`
C) `docker compose logs`
D) `golangci-lint run`

**Answer**: B
**Explanation**: GitHub CLI commands like `gh run list` provide workflow diagnostics for failure analysis and debugging.

---

### Q36: What is the estimated timeline for Phase 1 (Critical CI/CD Fixes)?
A) Days 1-2, ~16 hours
B) Days 1-3, ~20 hours
C) Days 2-3, ~12 hours
D) Day 1 only, ~8 hours

**Answer**: A
**Explanation**: Phase 1 spans days 1-2 with approximately 16 hours of estimated effort for critical workflow fixes.

---

### Q37: Which risk mitigation strategy is recommended for E2E startup timing?
A) Increase service resource limits
B) Use sequential service startup
C) Add retries + exponential backoff to health checks
D) Reduce number of services

**Answer**: C
**Explanation**: Timing issues should be addressed with robust retry mechanisms and exponential backoff rather than avoiding concurrency.

---

### Q38: What is the correct approach for coverage test performance in CI?
A) Run all tests with coverage always
B) Use selective test execution for local dev, full coverage in CI
C) Skip coverage tests in CI for speed
D) Run coverage tests only weekly

**Answer**: B
**Explanation**: Balance development speed with CI completeness by using selective execution locally and comprehensive coverage in CI.

---

### Q39: Which file type should contain benchmark test code?
A) `_test.go` files
B) `_bench_test.go` files
C) `_benchmark.go` files
D) `_perf_test.go` files

**Answer**: B
**Explanation**: Testing instructions specify `_bench_test.go` suffix for benchmark test files to organize test types clearly.

---

### Q40: What is the primary objective of workflow verification in ITER3-005?
A) Test only the modified workflows
B) Verify all 11 workflows pass after fixes
C) Test workflows individually
D) Focus on critical workflows only

**Answer**: B
**Explanation**: Complete verification requires testing all 11 workflows to ensure no regressions were introduced during fixes.

---

### Q41: Which testing pattern should be avoided for cryptographic operations?
A) Table-driven tests
B) Fuzz testing
C) Sequential execution only
D) Property-based testing

**Answer**: C
**Explanation**: Cryptographic operations benefit from concurrent testing to reveal race conditions and timing issues that sequential testing would hide.

---

### Q42: What is the correct approach for handling test containers in TestMain?
A) Start new container per test
B) Start container once per package, reuse across tests
C) Use external test services only
D) Avoid containers in tests

**Answer**: B
**Explanation**: TestMain pattern starts PostgreSQL containers once per package with all tests creating orthogonal data for isolation.

---

### Q43: Which mutation testing command excludes integration tests?
A) `gremlins unleash`
B) `gremlins unleash --exclude-integration`
C) `gremlins unleash --tags=!integration`
D) `gremlins unleash --unit-only`

**Answer**: C
**Explanation**: Gremlins mutation testing uses build tags to exclude integration tests while focusing on unit test effectiveness.

---

### Q44: What is the primary benefit of concurrent test execution?
A) Faster test execution only
B) Better test organization
C) Fastest execution + reveals production concurrency bugs
D) Easier test debugging

**Answer**: C
**Explanation**: Concurrent testing provides dual benefits: fastest execution and revealing race conditions that affect production code.

---

### Q45: How should dynamic port allocation be implemented for test servers?
A) Use hardcoded test ports
B) Use port 0 pattern with actual port extraction
C) Use random port generation
D) Use sequential port assignment

**Answer**: B
**Explanation**: Port 0 pattern allows OS to assign dynamic ports, with actual ports extracted from listeners for test URL construction.

---

### Q46: What is the correct approach for test data uniqueness?
A) Use hardcoded unique values
B) Use incremental counters
C) Use UUIDv7 for runtime-generated unique values
D) Use timestamps for uniqueness

**Answer**: C
**Explanation**: UUIDv7 provides thread-safe, process-safe unique identifiers ideal for concurrent test execution.

---

### Q47: Which phase has the highest estimated effort in Iteration 3?
A) Phase 1 (Critical CI/CD Fixes) - ~16 hours
B) Phase 2 (Deferred Work Completion) - ~15 hours
C) Phase 3 (Test Methodology Enhancements) - ~6 hours
D) Phase 4 (Documentation Cleanup) - ~3 hours

**Answer**: A
**Explanation**: Phase 1 has the highest effort estimate at ~16 hours for critical workflow fixes, followed by Phase 2 at ~15 hours.

---

### Q48: What should be the priority order for fixing workflow failures?
A) Start with easiest fixes first
B) Fix race detection, then coverage, then E2E/DAST/Load, then verify all
C) Fix all workflows simultaneously
D) Focus on most critical business workflows

**Answer**: B
**Explanation**: The implementation order prioritizes race detection fix, coverage improvements, Docker issues, then comprehensive verification.

---

### Q49: Which files should contain mutation testing baseline reports?
A) `test-output/` directory
B) `specs/` directory
C) `docs/` directory
D) Root project directory

**Answer**: B
**Explanation**: Mutation testing baseline reports should be tracked in specs/ directory to monitor improvements over iterations.

---

### Q50: What is the success criteria for Iteration 3 completion?
A) 95% code coverage achieved
B) All workflows passing + deferred features complete + enhanced testing
C) Documentation cleanup finished
D) All race conditions fixed

**Answer**: B
**Explanation**: Iteration 3 success requires achieving 100% workflow pass rate, completing deferred features, and implementing enhanced testing methodologies.

---

## Answer Summary

| Q# | Answer | Q# | Answer | Q# | Answer | Q# | Answer | Q# | Answer |
|----|--------|----|--------|----|--------|----|--------|----|--------|
| 1  | C      | 11 | C      | 21 | B      | 31 | C      | 41 | C      |
| 2  | B      | 12 | B      | 22 | B      | 32 | C      | 42 | B      |
| 3  | B      | 13 | B      | 23 | C      | 33 | C      | 43 | C      |
| 4  | C      | 14 | B      | 24 | C      | 34 | B      | 44 | C      |
| 5  | C      | 15 | B      | 25 | B      | 35 | B      | 45 | B      |
| 6  | B      | 16 | B      | 26 | C      | 36 | A      | 46 | C      |
| 7  | C      | 17 | B      | 27 | B      | 37 | C      | 47 | A      |
| 8  | B      | 18 | C      | 28 | C      | 38 | B      | 48 | B      |
| 9  | C      | 19 | B      | 29 | C      | 39 | B      | 49 | B      |
| 10 | C      | 20 | B      | 30 | B      | 40 | B      | 50 | B      |

---

*Grooming Session Version: 3.0.0*
*Generated for: cryptoutil Iteration 3*
*Focus: CI/CD reliability fixes and deferred work completion*
