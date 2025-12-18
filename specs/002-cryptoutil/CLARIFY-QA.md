# Requirement Validation Questions - 002-cryptoutil

**Date**: December 17, 2025
**Context**: Multiple choice questions to validate understanding of PLAN.md, TASKS.md, and clarify.md
**Purpose**: Ensure shared understanding of requirements before Phase 1 implementation
**Format**: 100 questions, 4 options each, 1 correct answer

---

## Instructions

- Read each question carefully
- Select the ONE best answer (A, B, C, or D)
- Refer to PLAN.md, TASKS.md, and clarify.md for authoritative answers
- Questions cover all 7 phases + cross-cutting concerns

---

## Phase 1: Test Performance Optimization (Questions 1-15)

### 1. What is the target test execution time per package for !integration tests?

A) ≤10 seconds
B) ≤12 seconds
C) ≤15 seconds
D) ≤20 seconds

**Answer**: B (clarify.md Q1.1)

### 2. Which probability constant should be used for base algorithms (e.g., RSA2048, AES256)?

A) TestProbNever (0%)
B) TestProbTenth (10%)
C) TestProbQuarter (25%)
D) TestProbAlways (100%)

**Answer**: D (clarify.md Q1.2)

### 3. What command is used to measure test timing baseline?

A) `go test -v ./...`
B) `go test -bench=. ./...`
C) `go test -json -v ./... 2>&1 | tee test-output/baseline-timing-002.txt`
D) `go test -coverprofile=coverage.out ./...`

**Answer**: C (clarify.md Q1.1)

### 4. If optimization causes coverage to drop, what should you do?

A) Accept coverage loss as trade-off for speed
B) Revert optimization and find alternative approach
C) Skip coverage checks for optimized packages
D) Document coverage loss and proceed

**Answer**: B (clarify.md Q1.3)

### 5. Which packages are exempt from the ≤12s timing target?

A) All packages in internal/
B) Packages with //go:build integration tag
C) Packages with mutation testing
D) No packages are exempt

**Answer**: B (clarify.md Q1.1)

### 6. What is the minimum probability for "important variants" (e.g., AES192, RSA3072)?

A) TestProbAlways (100%)
B) TestProbQuarter (25%)
C) TestProbTenth (10%)
D) TestProbNever (0%)

**Answer**: B (clarify.md Q1.2)

### 7. How should test timing results be parsed?

A) Manual review of text output
B) Parse JSON output from go test -json
C) Use third-party timing tools
D) Estimate based on CI/CD logs

**Answer**: B (clarify.md Q1.1)

### 8. What happens if a package can't reach ≤12s target after optimization?

A) Mark as BLOCKING issue, continue other work
B) Increase target to ≤15s for that package
C) Skip timing requirements for that package
D) Remove tests to meet target

**Answer**: A (clarify.md Q9.1)

### 9. Should benchmarks be included in the ≤12s timing target?

A) Yes, benchmarks count toward timing
B) No, only tests count toward timing
C) Only crypto benchmarks count
D) Benchmarks are not required

**Answer**: B (benchmarks are run separately with go test -bench, not included in regular test runs)

### 10. What is the primary reason for ≤12s target vs ≤15s in 001-cryptoutil?

A) Arbitrary reduction
B) Faster feedback loops for development velocity
C) GitHub Actions runner limitations
D) Reduce CI/CD costs

**Answer**: B (clarify.md Q1.1)

### 11. How do you verify that probabilistic execution is working correctly?

A) Check logs for "skipped test" messages
B) Run tests multiple times, observe different tests executing
C) Use go test -v to see test selection
D) Probabilistic execution doesn't need verification

**Answer**: B (probabilistic execution should show different tests executing in different runs)

### 12. What if removing algorithm variants reduces mutation score below 98%?

A) Accept reduced mutation score
B) Revert variant removal and find different optimization
C) Lower mutation target for that package
D) Skip mutation testing for optimized packages

**Answer**: B (clarify.md Q1.3 principle applies to all quality metrics)

### 13. Which tool is used to identify slow test packages?

A) golangci-lint
B) gremlins
C) go test -json with custom parsing
D) GitHub Actions built-in profiling

**Answer**: C (clarify.md Q1.1)

### 14. What is the correct order for test optimization steps?

A) Optimize → Measure coverage → Revert if dropped
B) Measure baseline → Optimize → Measure coverage → Revert if dropped
C) Optimize → Measure timing → Commit if faster
D) Measure baseline → Optimize → Commit immediately

**Answer**: B (clarify.md Q1.3)

### 15. Should table-driven tests be refactored to individual test functions to reduce timing?

A) Yes, smaller functions execute faster
B) No, keep table-driven tests, optimize within them
C) Only for tests >5 seconds
D) Only for crypto tests

**Answer**: B (instruction files mandate table-driven tests; optimize within table structure)

---

## Phase 2: Coverage Targets (Questions 16-30)

### 16. What is the MINIMUM acceptable coverage for production packages?

A) 80%
B) 85%
C) 90%
D) 95%

**Answer**: D (clarify.md Q2.1)

### 17. What is the MINIMUM acceptable coverage for infrastructure packages (internal/cmd/cicd/*)?

A) 80%
B) 90%
C) 95%
D) 100%

**Answer**: D (clarify.md Q2.1)

### 18. What does "NO EXCEPTIONS" mean for coverage enforcement?

A) Aim for target, accept 90%+ as success
B) Coverage < target = BLOCKING issue, no rationalization allowed
C) Coverage targets are guidelines, not hard requirements
D) Exception allowed for hard-to-test code

**Answer**: B (clarify.md Q2.1)

### 19. How should OS-specific code (e.g., sysinfo.go) be handled for coverage?

A) Exempt from coverage requirements
B) Use build tags, test on ALL supported platforms
C) Test on one platform only
D) Document as untestable, skip

**Answer**: B (clarify.md Q2.2)

### 20. What is the correct pattern for main() functions?

A) All logic in main(), accept 0% coverage
B) Thin main() wrapper calling testable internalMain()
C) No coverage requirements for main()
D) Use integration tests for main() coverage

**Answer**: B (clarify.md Q2.3)

### 21. How do you identify which lines need test coverage?

A) Read source code and guess
B) Use go tool cover -html to find RED lines
C) Ask CI/CD to report uncovered lines
D) Manually track coverage during development

**Answer**: B (clarify.md Q2.4)

### 22. What if a package improves from 60% to 85% coverage?

A) ✅ Success - 25 percentage point improvement
B) ⚠️ In progress - continue to 95%
C) ❌ BLOCKING - Still 10 points below target
D) Acceptable for first iteration

**Answer**: C (clarify.md Q2.1 - improvement is NOT success, only meeting target counts)

### 23. Can integration tests count toward coverage targets?

A) No, only unit tests count
B) Yes, but unit tests preferred
C) Only for packages without unit tests
D) Only for infrastructure packages

**Answer**: B (clarify.md Q2.5)

### 24. What is the correct coverage verification workflow?

A) Write tests → Check coverage
B) Generate baseline → Identify gaps → Write tests → Verify improvement
C) Write tests → Commit → Check CI/CD
D) Guess which code needs tests → Write tests

**Answer**: B (clarify.md Q2.4, instruction files mandate baseline analysis)

### 25. What if main() function can't delegate to internalMain() due to architecture constraints?

A) Accept 0% main() coverage, proceed
B) Refactor architecture to enable delegation
C) Use integration tests for main() coverage
D) Lower coverage target for that package

**Answer**: B (clarify.md Q2.3 - delegation is MANDATORY)

### 26. How should error paths be tested for coverage?

A) Error paths are low priority, skip if time-constrained
B) ALL error paths MUST be tested for 95%+ coverage
C) Test only "important" error paths
D) Document error paths as untestable

**Answer**: B (clarify.md Q2.1 - NO EXCEPTIONS applies to all code paths)

### 27. What if a package has 94.8% coverage after extensive test writing?

A) ✅ Round up to 95%, mark complete
B) ❌ BLOCKING - Continue until ≥95.0%
C) ⚠️ Acceptable - Focus on other packages
D) Document as "close enough"

**Answer**: B (clarify.md Q2.1 - NO EXCEPTIONS means exact target, no rounding)

### 28. Should coverage targets be enforced in CI/CD?

A) No, coverage is development-time concern
B) Yes, coverage < target should FAIL CI/CD builds
C) Yes, but only report, don't fail builds
D) Only for main branch, not PRs

**Answer**: B (clarify.md Q2.1 - BLOCKING means CI/CD enforcement required)

### 29. What is the utility code coverage target (e.g., internal/shared/*)?

A) 80%
B) 90%
C) 95%
D) 100%

**Answer**: D (clarify.md Q2.1)

### 30. How do you handle "unreachable code" for coverage?

A) Document as unreachable, exempt from coverage
B) Remove unreachable code (dead code elimination)
C) Add tests to reach the code
D) Lower coverage target for that package

**Answer**: B (unreachable code = dead code; should be removed, not exempted)

---

## Phase 3: CI/CD Workflow Fixes (Questions 31-45)

### 31. How many CI/CD workflows are currently failing?

A) 3
B) 4
C) 5
D) 6

**Answer**: C (clarify.md Q3.1: quality, mutations, fuzz, dast, load)

### 32. What is the priority order for fixing CI/CD workflows?

A) Alphabetical order
B) Longest runtime first
C) ci-quality → ci-fuzz/ci-load → ci-dast → ci-mutation
D) ci-mutation → ci-quality → ci-dast → ci-fuzz/ci-load

**Answer**: C (clarify.md Q3.1)

### 33. What is the target timeout for ci-mutation after parallelization?

A) 10 minutes per job, 15 minutes total
B) 15 minutes per job, 20 minutes total
C) 20 minutes per job, 25 minutes total
D) 30 minutes per job, 45 minutes total

**Answer**: B (clarify.md Q3.2)

### 34. What is the root cause of ci-fuzz and ci-load failures?

A) GitHub Actions runner timeout
B) OTEL collector configuration issues
C) PostgreSQL connection failures
D) TLS certificate validation errors

**Answer**: B (clarify.md Q3.1: "same root cause: otel collector")

### 35. Should we increase ci-dast timeout or optimize service startup?

A) Increase timeout only
B) Optimize startup only
C) BOTH - optimize first, then increase timeout
D) Neither - disable ci-dast temporarily

**Answer**: C (clarify.md Q3.3)

### 36. What is the target service startup time for ci-dast?

A) <15s
B) <30s
C) <45s
D) <60s

**Answer**: B (clarify.md Q3.3)

### 37. How should mutation testing be parallelized?

A) Run all packages in single job
B) Split packages into groups, run groups in parallel using matrix
C) Run each package in separate workflow
D) Use go test -parallel flag

**Answer**: B (clarify.md Q3.2)

### 38. Why is ci-quality the highest priority fix?

A) Fastest to fix
B) Quick win, unblocks merges
C) Required for other workflow fixes
D) Most frequently failing

**Answer**: B (clarify.md Q3.1)

### 39. How many packages should be in each ci-mutation matrix job?

A) 1 package per job
B) 2-3 packages per job
C) 4-6 packages per job
D) All packages in one job

**Answer**: C (clarify.md Q3.2)

### 40. What is the expected total runtime for parallelized ci-mutation?

A) 10 minutes
B) 15 minutes
C) ~20 minutes (max of all jobs + aggregation)
D) 30 minutes

**Answer**: C (clarify.md Q3.2)

### 41. Should CI/CD workflow fixes be committed incrementally or in one big PR?

A) One big PR with all fixes
B) Incrementally, one workflow per commit
C) Incrementally, one fix per PR
D) All fixes in single commit

**Answer**: B (clarify.md Q3.1: "Incrementally, in priority order")

### 42. What retry logic should be implemented for ci-dast?

A) No retries, increase timeout only
B) Exponential backoff retry logic
C) Fixed interval retries (every 10s)
D) Retry until success (no timeout)

**Answer**: B (clarify.md Q3.3)

### 43. What is the ci-dast timeout target after optimization?

A) 30s with no retries
B) 45s with exponential backoff
C) 60s with retries
D) 120s with no retries

**Answer**: C (clarify.md Q3.3)

### 44. Which workflow should be fixed LAST?

A) ci-quality
B) ci-fuzz
C) ci-dast
D) ci-mutation

**Answer**: D (clarify.md Q3.1: "ci-mutation last (requires most work)")

### 45. What should happen if ci-quality fix causes regression in other workflows?

A) Revert ci-quality fix
B) Continue fixing other workflows
C) Investigate regression, fix root cause
D) Disable failing workflows temporarily

**Answer**: C (regressions must be investigated and fixed, not ignored or reverted without understanding)

---

## Phase 4: Mutation Testing QA (Questions 46-60)

### 46. What is the target mutation efficacy per package?

A) ≥80%
B) ≥90%
C) ≥95%
D) ≥98%

**Answer**: D (clarify.md Q4.1)

### 47. What does "mutation efficacy" mean?

A) (killed mutants / total tests) × 100%
B) (killed mutants / total mutants) × 100%
C) (lived mutants / total mutants) × 100%
D) (coverage / mutation score) × 100%

**Answer**: B (clarify.md Q4.1)

### 48. How do you identify which mutants survived (lived)?

A) Check gremlins output for "Lived" status
B) Run go test -v and check for failures
C) Review coverage report
D) Analyze CI/CD logs

**Answer**: A (clarify.md Q4.2)

### 49. Can low-risk packages skip mutation testing?

A) Yes, focus on high-risk packages only
B) No, ALL packages MUST achieve ≥98% efficacy
C) Yes, if coverage ≥95%
D) Yes, if integration tests exist

**Answer**: B (clarify.md Q4.3)

### 50. What if a package can't reach 98% efficacy?

A) Accept 95% minimum with documentation
B) Skip mutation testing for that package
C) Lower target to 90%
D) Mark as BLOCKING, continue other work

**Answer**: A (clarify.md Q4.4: "aim for ≥95% minimum" with justification)

### 51. What is the process for killing lived mutants?

A) Ignore lived mutants, focus on coverage
B) Write targeted tests for specific lived mutants
C) Increase mutation timeout
D) Disable mutation types that produce lived mutants

**Answer**: B (clarify.md Q4.2)

### 52. Which tool is used for mutation testing?

A) go test -mutate
B) gofuzz
C) gremlins
D) gobench

**Answer**: C (instruction files specify gremlins)

### 53. What is the correct command to run mutation testing?

A) `go test -mutate ./pkg`
B) `gremlins unleash ./pkg`
C) `mutation-test ./pkg`
D) `go test -fuzz=FuzzAll ./pkg`

**Answer**: B (clarify.md Q4.2)

### 54. Should integration tests be included in mutation testing?

A) Yes, include all tests
B) No, exclude integration tests with `--tags=!integration`
C) Only for packages without unit tests
D) Only for infrastructure packages

**Answer**: B (instruction files specify excluding integration tests for gremlins)

### 55. What if gremlins panics on Windows?

A) Skip mutation testing on Windows
B) Use CI/CD (Linux-based) for mutation testing
C) Report issue to gremlins maintainers and skip
D) Disable mutation testing entirely

**Answer**: B (clarify.md Q8.3)

### 56. How many workers should be configured for gremlins?

A) 1 worker (sequential)
B) 2 workers
C) 4 workers
D) 8 workers

**Answer**: C (instruction files: `.gremlins.yaml` workers: 4)

### 57. What is the timeout coefficient for gremlins?

A) 1x (no timeout multiplier)
B) 2x (timeout multiplier)
C) 3x (timeout multiplier)
D) 5x (timeout multiplier)

**Answer**: B (instruction files: `.gremlins.yaml` timeout-coefficient: 2)

### 58. Should mutation testing be run locally or in CI/CD?

A) Locally only (faster feedback)
B) CI/CD only (Linux compatibility)
C) Both (local for quick checks, CI/CD for authoritative results)
D) Locally on Linux, CI/CD on Windows

**Answer**: B (clarify.md Q8.3 specifies CI/CD due to Windows panic)

### 59. What if a mutant is a linter false positive?

A) Document justification, aim for ≥95% minimum
B) Disable linter for that code
C) Ignore lived mutant
D) Rewrite code to avoid false positive

**Answer**: A (clarify.md Q4.4)

### 60. How do you verify that mutation testing improvements are real?

A) Check mutation score increased
B) Re-run gremlins, verify mutant now "Killed"
C) Check coverage increased
D) Run integration tests

**Answer**: B (clarify.md Q4.2)

---

## Phase 5: Hash Refactoring (Questions 61-70)

### 61. How many hash types are in the new architecture?

A) 2
B) 3
C) 4
D) 8

**Answer**: C (clarify.md Q5.1: Low/High × Random/Deterministic = 4 types)

### 62. What is the difference between Low Entropy and High Entropy hashing?

A) Output size (Low=16 bytes, High=32 bytes)
B) Input type (Low=user-provided, High=system-generated)
C) Algorithm (Low=PBKDF2, High=HKDF)
D) Security level (Low=less secure, High=more secure)

**Answer**: B (clarify.md Q5.1)

### 63. What is the difference between Random and Deterministic hashing?

A) Algorithm selection
B) Output size
C) Salt usage (Random=with salt, Deterministic=no salt)
D) Security level

**Answer**: C (clarify.md Q5.2)

### 64. How many hash versions are supported?

A) 1 (SHA256 only)
B) 2 (SHA256, SHA512)
C) 3 (SHA256, SHA384, SHA512)
D) 4 (SHA256, SHA384, SHA512, SHA3-256)

**Answer**: C (clarify.md Q5.3)

### 65. How is hash version selected?

A) User-specified parameter
B) By input size (0-31 bytes=v1, 32-47 bytes=v2, 48+ bytes=v3)
C) By security requirements
D) Random selection

**Answer**: B (clarify.md Q5.3)

### 66. What algorithm is used for Low Entropy Random hashing?

A) HKDF
B) PBKDF2
C) bcrypt
D) scrypt

**Answer**: B (clarify.md Q5.1: Low Entropy requires key derivation with PBKDF2)

### 67. What algorithm is used for High Entropy Deterministic hashing?

A) PBKDF2
B) HKDF (without salt)
C) SHA256 direct
D) bcrypt

**Answer**: B (clarify.md Q5.1: High Entropy requires key stretching with HKDF)

### 68. How does version-aware Verify work?

A) Always use latest version for verification
B) Parse version from hash output, use same parameters
C) Try all versions until one matches
D) User specifies version as parameter

**Answer**: B (clarify.md Q5.4)

### 69. What is the format of hash output?

A) `hash_data:salt`
B) `{version}$hash_data$salt`
C) `v1-hash_data-salt`
D) `hash_data_v1_salt`

**Answer**: B (clarify.md Q5.4)

### 70. Why is larger output digest used for larger inputs?

A) Performance optimization
B) Collision resistance
C) Compatibility requirements
D) Arbitrary design choice

**Answer**: B (clarify.md Q5.3)

---

## Phase 6: Service Template Extraction (Questions 71-85)

### 71. Which service is used as the source for template extraction?

A) identity-authz
B) identity-idp
C) jose-ja (JOSE)
D) sm-kms (KMS)

**Answer**: D (clarify.md Q6.3: "Extract template from SM-KMS")

### 72. How many PRODUCT-SERVICE instances will use the template?

A) 4
B) 6
C) 8
D) 10

**Answer**: C (clarify.md introduction: 8 instances including sm-kms)

### 73. What are the two API path prefixes?

A) /api and /ui
B) /public and /private
C) /browser and /service
D) /client and /server

**Answer**: C (clarify.md Q6.1)

### 74. Which authentication flow is used for /browser paths?

A) Client Credentials flow
B) Authorization Code + PKCE flow
C) Implicit flow
D) Resource Owner Password Credentials flow

**Answer**: B (clarify.md Q6.1)

### 75. Which authentication flow is used for /service paths?

A) Client Credentials flow
B) Authorization Code + PKCE flow
C) Implicit flow
D) Resource Owner Password Credentials flow

**Answer**: A (clarify.md Q6.1)

### 76. What port is used for admin endpoints?

A) 8080
B) 8443
C) 9090
D) 9443

**Answer**: C (clarify.md Q6.2)

### 77. Why must admin endpoints use 127.0.0.1 binding?

A) Performance optimization
B) Security - NOT externally accessible
C) Docker compatibility
D) Kubernetes requirement

**Answer**: B (clarify.md Q6.2)

### 78. Which middlewares are specific to /browser paths?

A) mTLS, API key, bearer token
B) CORS, CSRF, CSP, session cookies
C) Rate limiting, IP allowlist
D) OAuth 2.0 validation

**Answer**: B (clarify.md Q6.1)

### 79. Should barrier services be part of the template?

A) No, barrier services are KMS-specific
B) Yes, but optional per service
C) Yes, required for all services
D) No, barrier services deprecated

**Answer**: B (clarify.md Q6.3)

### 80. How is template completeness validated?

A) Code review by architects
B) Integration testing
C) Phase 7 (Learn-PS builds entirely with template)
D) CI/CD workflow passing

**Answer**: C (clarify.md Q6.4)

### 81. What happens if Learn-PS requires custom infrastructure code?

A) Document as template limitation
B) Build custom code for Learn-PS
C) Template incomplete, iterate on template
D) Skip Phase 7

**Answer**: C (clarify.md Q6.4)

### 82. How many public HTTPS endpoints does each service have?

A) 1 (public API)
B) 2 (public API + admin API)
C) 3 (browser API + service API + admin API)
D) 4 (browser API + service API + admin API + metrics)

**Answer**: B (actually C - public API has dual paths /browser and /service, plus admin API = 3 logical endpoints on 2 ports)

### 83. What is the relationship between /browser and /service APIs?

A) Different APIs with different functionality
B) Same OpenAPI spec, different middleware
C) /browser is deprecated, use /service
D) /service is subset of /browser

**Answer**: B (clarify.md Q6.1)

### 84. Which services need barrier-enabled template?

A) All 8 services
B) Only sm-kms
C) sm-kms and pki-ca (services needing encrypted-at-rest)
D) identity-authz, identity-idp only

**Answer**: C (clarify.md Q6.3)

### 85. What if template extraction reduces KMS service code coverage below 95%?

A) Accept coverage loss, prioritize template
B) Revert template extraction, maintain coverage
C) Template extraction MUST maintain coverage
D) Lower coverage target for KMS

**Answer**: C (clarify.md Q1.3 principle: quality metrics must be maintained)

---

## Phase 7: Learn-PS Demonstration (Questions 86-95)

### 86. What does "Learn-PS" stand for?

A) Learn-PostScript
B) Learn-PowerShell
C) Learn-Pet Store
D) Learn-Production Service

**Answer**: C (clarify.md Q7.1)

### 87. Why was Pet Store chosen as the demonstration domain?

A) Pet stores are popular
B) Simple, well-understood domain without cryptographic complexity
C) Existing Pet Store specifications available
D) Integration with existing services

**Answer**: B (clarify.md Q7.1)

### 88. Should Learn-PS achieve the same quality targets as main services?

A) No, it's just a demonstration
B) Yes, Learn-PS is customer-facing demonstration
C) Only coverage requirements apply
D) Only CI/CD requirements apply

**Answer**: B (clarify.md Q7.2)

### 89. Should Learn-PS be deployed to production?

A) Yes, as demonstration for customers
B) No, demonstration/education only
C) Yes, but without monitoring
D) Yes, to test template in production

**Answer**: B (clarify.md Q7.3)

### 90. What is the target video demonstration length?

A) 5-10 minutes
B) 10-15 minutes
C) 15-20 minutes
D) 20-30 minutes

**Answer**: C (clarify.md Q7.4)

### 91. What should the video demonstration cover?

A) Only code walkthrough
B) Only API usage examples
C) Service startup + API usage + code walkthrough + customization tips
D) Only service startup and deployment

**Answer**: C (clarify.md Q7.4: 4 parts)

### 92. What are the Learn-PS deployment targets?

A) Production Kubernetes cluster
B) Docker Compose (local) + Kubernetes manifests (reference)
C) Docker Compose only
D) GitHub Pages static site

**Answer**: B (clarify.md Q7.3)

### 93. Which operations should Learn-PS support?

A) Only read operations (GET)
B) CRUD operations + pagination + filtering + sorting
C) CRUD operations only
D) Read operations + business logic

**Answer**: B (clarify.md Q7.1)

### 94. Should Learn-PS include barrier services?

A) Yes, to demonstrate template completeness
B) No, not in scope
C) Only if time permits
D) Yes, required for all services

**Answer**: B (clarify.md Q7.1: "NOT in scope: Cryptographic operations, barrier services")

### 95. What is the Learn-PS mutation efficacy target?

A) 80% (demonstration quality)
B) 90% (reduced for demo)
C) 95% (same as production)
D) 98% (same as production)

**Answer**: D (clarify.md Q7.2: "YES. Learn-PS is customer-facing demonstration" with same quality targets)

---

## Cross-Cutting Concerns (Questions 96-100)

### 96. What should you do if DETAILED.md exceeds 2000 lines?

A) Continue editing, no line limit
B) Archive to 003-cryptoutil and start fresh
C) Split into multiple files
D) Compress timeline entries

**Answer**: B (clarify.md Q8.1)

### 97. Which platform should be preferred for CI/CD?

A) Windows (local development parity)
B) Linux (production parity)
C) macOS (Apple Silicon performance)
D) Mix of all platforms

**Answer**: B (clarify.md Q8.2)

### 98. What is the definition of "BLOCKING issue"?

A) Issue that slows down development
B) Work cannot proceed to next task until resolved
C) Issue requiring manager approval
D) Critical security vulnerability

**Answer**: B (clarify.md Q9.1)

### 99. Can phases be completed out of order?

A) Yes, work on any phase independently
B) No, phases are sequential dependencies
C) Only P1 and P2 must be sequential
D) Only P6 and P7 must be sequential

**Answer**: B (clarify.md Q9.2)

### 100. When is a phase considered "complete"?

A) When most tasks in phase meet success criteria
B) When timeline entry added for phase completion
C) When ALL tasks in phase meet ALL success criteria
D) When phase owner approves completion

**Answer**: C (clarify.md Q9.2)

---

## Answer Key

### Phase 1 (1-15)

- 1.B, 2.D, 3.C, 4.B, 5.B, 6.B, 7.B, 8.A, 9.B, 10.B, 11.B, 12.B, 13.C, 14.B, 15.B

### Phase 2 (16-30)

- 16.D, 17.D, 18.B, 19.B, 20.B, 21.B, 22.C, 23.B, 24.B, 25.B, 26.B, 27.B, 28.B, 29.D, 30.B

### Phase 3 (31-45)

- 31.C, 32.C, 33.B, 34.B, 35.C, 36.B, 37.B, 38.B, 39.C, 40.C, 41.B, 42.B, 43.C, 44.D, 45.C

### Phase 4 (46-60)

- 46.D, 47.B, 48.A, 49.B, 50.A, 51.B, 52.C, 53.B, 54.B, 55.B, 56.C, 57.B, 58.B, 59.A, 60.B

### Phase 5 (61-70)

- 61.C, 62.B, 63.C, 64.C, 65.B, 66.B, 67.B, 68.B, 69.B, 70.B

### Phase 6 (71-85)

- 71.D, 72.C, 73.C, 74.B, 75.A, 76.C, 77.B, 78.B, 79.B, 80.C, 81.C, 82.C, 83.B, 84.C, 85.C

### Phase 7 (86-95)

- 86.C, 87.B, 88.B, 89.B, 90.C, 91.C, 92.B, 93.B, 94.B, 95.D

### Cross-Cutting (96-100)

- 96.B, 97.B, 98.B, 99.B, 100.C

---

## Scoring Guide

- **95-100**: Excellent understanding, ready to proceed with implementation
- **90-94**: Good understanding, review clarify.md sections for questions missed
- **85-89**: Adequate understanding, review PLAN.md and clarify.md before starting
- **80-84**: Concerning gaps, mandatory review of all specification documents
- **<80**: BLOCKING - Do not proceed with implementation until understanding improved

---

## Next Steps

1. Review answers against PLAN.md, TASKS.md, and clarify.md
2. Identify areas of misunderstanding
3. Re-read relevant sections until 95%+ score achieved
4. Begin Phase 1 implementation with confidence in shared understanding
