# Requirement Clarifications - 002-cryptoutil

**Date**: December 17, 2025
**Context**: Fresh clarifications after archiving 001-cryptoutil
**Status**: 🎯 MVP Quality Requirements

---

## Overview

This document clarifies ambiguous requirements from PLAN.md and TASKS.md, addressing questions that arose during 001-cryptoutil implementation.

---

## Phase 1: Test Performance Optimization

### Q1.1: What exactly is the "≤12s per package" target?

**Answer**: ALL !integration test packages MUST execute in ≤12 seconds when run with `go test ./pkg` (excluding integration tests tagged with `//go:build integration`).

**Rationale**: More aggressive than 001-cryptoutil (was ≤15s). Faster feedback loops essential for development velocity.

**Measurement**: `go test -json -v ./... 2>&1 | tee test-output/baseline-timing-002.txt` → parse JSON output → flag packages >12s.

### Q1.2: Can we skip algorithm variants entirely to meet timing target?

**Answer**: NO. Use probabilistic execution patterns:

- **Base algorithms**: `TestProbAlways` (100% execution) - NEVER skip
- **Important variants**: `TestProbQuarter` (25% execution) - Run 1 in 4 times
- **Less critical variants**: `TestProbTenth` (10% execution) - Run 1 in 10 times

**Rationale**: Statistical sampling ensures bugs eventually caught without running all variants every time.

### Q1.3: What if optimization causes coverage to drop?

**Answer**: Coverage loss is NOT acceptable. If optimization drops coverage, revert optimization and find alternative approach.

**Process**:

1. Run coverage baseline BEFORE optimization
2. Apply optimization
3. Run coverage again
4. If coverage dropped, REVERT and try different optimization
5. Only proceed if coverage maintained or improved

---

## Phase 2: Coverage Targets

### Q2.1: What does "NO EXCEPTIONS" mean for 95%+ coverage?

**Answer**: Coverage < 95% (production) or < 100% (infra/util) = BLOCKING issue. NO rationalization allowed.

**Examples of FORBIDDEN rationalizations**:

- ❌ "This package is mostly error handling" → Add error path tests
- ❌ "This is just a thin wrapper" → Still needs 95% coverage
- ❌ "We improved from 60% to 70%" → Still 25 points below target, NOT success

**Only valid outcome**: Coverage ≥ target.

### Q2.2: How do we handle OS-specific code (e.g., sysinfo.go)?

**Answer**: Use build tags and test on ALL supported platforms in CI/CD. If certain OS APIs untestable, document in coverage report with justification, but aim for ≥95% on testable portions.

**Process**:

1. Identify OS-specific code
2. Use `//go:build windows` or `//go:build linux` tags
3. Test on both Windows and Linux CI/CD runners
4. Document any untestable OS API wrappers with justification
5. Ensure testable portions ≥95%

### Q2.3: What if main() functions can't reach 95%?

**Answer**: main() functions MUST be thin wrappers calling testable internalMain():

```go
func main() {
    os.Exit(internalMain(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

func internalMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    // All logic here - fully testable with mocks
    return 0
}
```

**Coverage calculation**: main() at 0% is acceptable IF internalMain() ≥95%.

### Q2.4: How do we identify RED lines in HTML coverage reports?

**Answer**: Use `go tool cover -html=./test-output/coverage_pkg.out -o ./test-output/coverage_pkg.html`:

1. Open HTML file in browser
2. RED lines = uncovered code
3. GREEN lines = covered code
4. Write tests ONLY for RED lines (targeted, not trial-and-error)

### Q2.5: Can we use integration tests to meet coverage targets?

**Answer**: YES, but unit tests preferred. Integration tests count toward coverage, but:

- Integration tests tagged with `//go:build integration`
- Integration tests excluded from ≤12s timing target
- Unit tests provide faster feedback, prefer unit tests where possible

---

## Phase 3: CI/CD Workflow Fixes

### Q3.1: Should we fix all 5 workflows at once or incrementally?

**Answer**: Incrementally, in priority order:

1. **ci-quality** (quick win, unblocks merges)
2. **ci-fuzz** + **ci-load** together (same root cause: otel collector)
3. **ci-dast** (optimize startup, increase timeout)
4. **ci-mutation** last (requires most work: parallelization)

**Rationale**: Quick wins first, complex fixes last.

### Q3.2: What's the target timeout for ci-mutation after parallelization?

**Answer**: 15 minutes per job, ~20 minutes total (parallel execution).

**Process**:

1. Split packages into groups (4-6 packages per job)
2. Run groups in parallel using GitHub Actions matrix
3. Each job timeout: 15 minutes
4. Total workflow time: max(all jobs) + aggregation ~20 minutes

### Q3.3: Should we increase timeouts or optimize service startup for ci-dast?

**Answer**: BOTH.

1. **Optimize startup first**: Reduce database migration overhead, parallelize unseal operations
2. **Increase timeout second**: Add exponential backoff retry logic
3. **Target**: Service startup <30s, timeout 60s with retries

**Rationale**: Optimization improves production performance, timeout provides CI/CD resilience.

---

## Phase 4: Mutation Testing QA

### Q4.1: What does "98%+ efficacy" mean?

**Answer**: Mutation efficacy = (killed mutants / total mutants) × 100%.

**Example**:

- Total mutants: 100
- Killed mutants: 98
- Lived mutants: 2
- Efficacy: 98%

**Target**: ≥98% efficacy per package.

### Q4.2: How do we identify lived mutants?

**Answer**: gremlins output shows which mutants survived:

```
Mutant #42: Lived
  File: foo.go:123
  Original: if x > 0
  Mutated:  if x >= 0
  Reason: No test failed with this mutation
```

**Process**:

1. Run `gremlins unleash ./pkg`
2. Review output for "Lived" mutants
3. Write tests targeting specific lived mutants
4. Re-run gremlins, verify mutant now "Killed"

### Q4.3: Can we skip mutation testing for low-risk packages?

**Answer**: NO. ALL packages MUST achieve ≥98% efficacy, no exceptions.

**Rationale**: Even "low-risk" packages can introduce bugs. Mutation testing validates test quality.

### Q4.4: What if we can't reach 98% efficacy?

**Answer**: Analyze lived mutants, write targeted tests. If truly unreachable (e.g., linter false positives), document justification and aim for ≥95% minimum.

**Process**:

1. Identify lived mutants
2. Determine if mutant represents real bug risk
3. If yes: Write test to kill mutant
4. If no: Document why mutant is false positive
5. Repeat until ≥98% or ≥95% (with justification)

---

## Phase 5: Hash Refactoring

### Q5.1: What's the difference between Low/High Entropy?

**Answer**:

- **Low Entropy**: User-provided input (passwords, passphrases) - requires key derivation (PBKDF2)
- **High Entropy**: System-generated input (crypto keys, tokens) - requires key stretching (HKDF)

### Q5.2: What's the difference between Random/Deterministic?

**Answer**:

- **Random**: Uses salt (different hash each time) - secure for passwords
- **Deterministic**: No salt (same hash each time) - useful for key derivation

### Q5.3: How do we select hash version (SHA256/384/512)?

**Answer**: By input size:

- **v1 (SHA256)**: 0-31 bytes input
- **v2 (SHA384)**: 32-47 bytes input
- **v3 (SHA512)**: 48+ bytes input

**Rationale**: Larger inputs benefit from larger output digests (collision resistance).

### Q5.4: How does version-aware Verify work?

**Answer**: Hash output includes version metadata: `{version}$hash_data$salt`

**Process**:

1. Parse version from hash output
2. Use same parameters (version, algorithm, iterations) for verification
3. Compare computed hash with stored hash
4. Return true if match, false otherwise

**Example**:

```
Input: "mysecretpassword"
HashWithLatest() → "{1}$pbkdf2_hmac_sha256_output$salt"
Verify("mysecretpassword", "{1}$pbkdf2_hmac_sha256_output$salt") → true
```

---

## Phase 6: Service Template Extraction

### Q6.1: What's the difference between /browser and /service API paths?

**Answer**:

- **`/browser/**`**: Browser-to-service (session-based, OAuth 2.1 Authorization Code + PKCE flow)
- **`/service/**`**: Service-to-service (token-based, OAuth 2.1 Client Credentials flow)

**Same OpenAPI spec, different middleware**:

- /browser: CORS, CSRF, CSP, session cookies
- /service: mTLS, API key, bearer token

### Q6.2: Why 127.0.0.1:9090 for admin endpoints?

**Answer**: Security - admin endpoints MUST NOT be externally accessible.

- **127.0.0.1**: Localhost-only binding (not 0.0.0.0)
- **:9090**: Standard port for admin operations
- **NOT externally routable**: Docker/K8s expose only public endpoint

### Q6.3: Should barrier services be part of the template?

**Answer**: YES, but optional per service.

- **Barrier-enabled template**: For services needing encrypted-at-rest (KMS, CA)
- **Barrier-free template**: For services without encryption requirements (IdP, SPA)

**Template provides both**: Service chooses which to use.

### Q6.4: How do we validate template completeness?

**Answer**: Phase 7 (Learn-PS demonstration service).

**Process**:

1. Extract template from SM-KMS (Phase 6)
2. Use template to build Learn-PS Pet Store (Phase 7)
3. If Learn-PS requires custom infrastructure code → template incomplete → iterate
4. Only proceed if Learn-PS built ENTIRELY with template

---

## Phase 7: Learn-PS Demonstration Service

### Q7.1: Why Pet Store as demonstration?

**Answer**: Simple, well-understood domain. Avoids cryptographic complexity while demonstrating template features.

**Pet Store scope**:

- CRUD operations (pets, orders, customers)
- Pagination, filtering, sorting
- Business logic (inventory, order processing)
- Database interactions (PostgreSQL + SQLite)

**NOT in scope**: Cryptographic operations, barrier services.

### Q7.2: Must Learn-PS achieve same quality targets as main services?

**Answer**: YES. Learn-PS is customer-facing demonstration.

**Quality targets**:

- ≤12s test execution
- 95%+ coverage
- 98%+ mutation efficacy
- All CI/CD workflows passing

**Rationale**: Learn-PS proves template produces production-quality services.

### Q7.3: Should Learn-PS be deployed to production?

**Answer**: NO. Learn-PS is demonstration/education only.

**Deployment targets**:

- Docker Compose (local development)
- Kubernetes manifests (reference implementation)

**NOT deployed**: No production infrastructure, no monitoring, no SLAs.

### Q7.4: What's the video demonstration scope?

**Answer**: 15-20 minute walkthrough:

1. **Part 1 (5 min)**: Service startup, Docker Compose, health checks
2. **Part 2 (5 min)**: API usage examples (curl, Postman, generated SDK)
3. **Part 3 (5 min)**: Code walkthrough (template usage, customization points)
4. **Part 4 (5 min)**: Customization tips (add new endpoints, change business logic)

**Format**: Screen recording with voiceover, uploaded to YouTube, linked from README.md.

---

## Cross-Cutting Concerns

### Q8.1: How do we prevent 002-cryptoutil DETAILED.md from growing to 3710 lines like 001?

**Answer**: Discipline and concise timeline entries.

**Guidelines**:

- Timeline entries ≤50 lines each
- Summarize work, don't copy-paste code
- Link to commits for detailed changes
- Archive to 003-cryptoutil if exceeds 2000 lines

### Q8.2: Should we use Windows or Linux for development?

**Answer**: BOTH, but prefer Linux for CI/CD.

**Windows**: Local development, IDE support, Windows-specific testing
**Linux**: CI/CD, Docker, Kubernetes, production parity

**Process**:

- Develop on Windows (VS Code, Copilot)
- Test on both Windows and Linux (CI/CD)
- Deploy from Linux (Docker, Kubernetes)

### Q8.3: How do we handle gremlins Windows panic?

**Answer**: Use CI/CD for mutation testing (Linux-based).

**Process**:

1. Local development on Windows (unit tests, coverage)
2. Push to GitHub
3. CI/CD runs mutation testing on Linux (gremlins works)
4. Review mutation results from CI/CD artifacts

**Permanent fix**: Track gremlins upstream issue, re-evaluate after v0.7.0 release.

---

## Success Criteria Clarifications

### Q9.1: What's the definition of "BLOCKING issue"?

**Answer**: Work cannot proceed to next task until BLOCKING issue resolved.

**Examples**:

- Coverage <95% (production) or <100% (infra/util) → BLOCKING
- Test execution >12s → BLOCKING
- CI/CD workflow failure → BLOCKING
- Mutation efficacy <98% → BLOCKING

**Resolution**: Fix issue, verify fix, only then proceed.

### Q9.2: When is a phase considered "complete"?

**Answer**: When ALL tasks in phase meet success criteria.

**Process**:

1. Complete all tasks in phase
2. Verify success criteria for each task
3. Run full verification (all tests, coverage, mutations, CI/CD)
4. Only mark phase complete if ALL criteria met

### Q9.3: Can we skip phases?

**Answer**: NO. Phases are sequential dependencies.

**Rationale**:

- P1 (test perf) → enables fast P2-P7 development
- P2 (coverage) → prerequisite for P4 (mutations)
- P3 (CI/CD) → quality gates must pass before P4-P7
- P4 (mutations) → validates test quality before P5-P6 refactoring
- P5 (hashes) → simpler refactoring, proves discipline for P6
- P6 (template) → prerequisite for P7 (Learn-PS)
- P7 (Learn-PS) → validates template from P6

**Only proceed if prerequisites met**.

---

## Conclusion

This clarification document addresses ambiguities from PLAN.md and TASKS.md. Key clarifications:

1. **≤12s target**: ALL !integration packages, NO exceptions
2. **95%+ coverage**: NO EXCEPTIONS, BLOCKING until met
3. **98%+ mutations**: Per package, NO exceptions
4. **CI/CD fixes**: Incremental, priority order
5. **Hash architecture**: 4 types, 3 versions, version-aware Verify
6. **Service template**: Dual HTTPS, dual paths, barrier optional
7. **Learn-PS**: Production quality, video demonstration, NOT deployed

**Next Steps**: Proceed with Phase 1 (test performance optimization), use these clarifications to resolve ambiguities during implementation.
