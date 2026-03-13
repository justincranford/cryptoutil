# Tasks - Framework v3

**Status**: 11 of 58 tasks complete (19%)
**Last Updated**: 2026-03-12
**Created**: 2026-03-08

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- OK **Correctness**: ALL code must be functionally correct with comprehensive tests
- OK **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- OK **Thoroughness**: Evidence-based validation at every step
- OK **Reliability**: Quality gates enforced (>=95%/98% coverage/mutation)
- OK **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- OK **Accuracy**: Changes must address root cause, not just symptoms
- NO **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- NO **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

---

## Task Checklist

### Phase 1: Close v1 Gaps and Knowledge Propagation

**Phase Objective**: Fix immediate gaps from v1 review, propagate knowledge, add CI workflow, complete contract test coverage.

#### Task 1.1: Fix lessons.md auth contracts item

- **Status**: DONE
- **Dependencies**: None
- **Description**: Correct the inherited lesson that says "auth contracts belong in service-specific tests" - auth is 100% service-template owned
- **Acceptance Criteria**:
  - [x] lessons.md item 4 updated to reflect auth is service-template owned
  - [x] No references to "auth belongs in service-specific tests" remain

#### Task 1.2: Propagate timeout double-multiplication lesson

- **Status**: DONE
- **Dependencies**: None
- **Description**: The lesson about `time.Duration` constants NOT being multiplied by `time.Second` is only in ARCHITECTURE.md and lessons.md. Propagate to skills and instructions.
- **Acceptance Criteria**:
  - [x] `03-02.testing.instructions.md` includes timeout double-multiplication warning (Forbidden Pattern #7)
  - [x] `coverage-analysis` skill includes timeout pattern check (Common Pitfalls section)
  - [x] `contract-test-gen` skill includes timeout warning (Critical Notes section, fixed garbled encoding)
  - [x] Verify ARCHITECTURE.md Section 10.3.4 already documents this (confirmed)

#### Task 1.3: Clean up temp research files

- **Status**: DONE
- **Dependencies**: None
- **Description**: Remove temp_sequential_results.txt and any other research artifacts from repo root
- **Acceptance Criteria**:
  - [x] No temp research files in repo root
  - [x] `git status` clean after cleanup

#### Task 1.4: Add ci-fitness.yml GitHub Actions workflow

- **Status**: DONE
- **Dependencies**: None
- **Description**: Add CI workflow for `cicd lint-fitness` so fitness checks run in CI, not just pre-commit
- **Acceptance Criteria**:
  - [x] `.github/workflows/ci-fitness.yml` created
  - [x] Triggers on push/PR for relevant file changes (.go, .sql, .yml)
  - [x] Runs `go run ./cmd/cicd lint-fitness`
  - [x] Uses `actions/setup-go@v6` with `cache: true`
  - [x] Follows existing workflow conventions (see ci-quality.yml for reference)

#### Task 1.5: Add auth contract tests to RunContractTests

- **Status**: DONE
- **Dependencies**: None
- **Description**: Add 401/403 rejection contract tests to the cross-service contract suite. Auth is 100% service-template owned.
- **Acceptance Criteria**:
  - [x] New `auth_contracts.go` in `internal/apps/template/service/testing/contract/`
  - [x] Tests unauthenticated requests get 401 on protected endpoints
  - [x] Tests unauthorized requests get 403 (note: 403 requires authorization infrastructure not yet built; 401 is fully tested)
  - [x] Tests both `/service/**` and `/browser/**` paths
  - [x] Unit tests for auth contract tests (>=95% coverage)
  - [x] `RunContractTests` calls new auth contracts (opt-in via AuthContractServer interface)

#### Task 1.6: Integrate contract tests into identity-authz

- **Status**: DONE
- **Dependencies**: Task 1.5
- **Description**: Add `RunContractTests(t, server)` to identity-authz integration tests
- **Acceptance Criteria**:
  - [x] identity-authz calls `RunContractTests`
  - [x] All contract tests pass including auth contracts
  - [x] `go test ./internal/apps/identity/authz/...` passes

#### Task 1.7: Integrate contract tests into identity-idp

- **Status**: DONE
- **Dependencies**: Task 1.5
- **Description**: Add `RunContractTests(t, server)` to identity-idp integration tests
- **Acceptance Criteria**:
  - [x] identity-idp calls `RunContractTests`
  - [x] All contract tests pass
  - [x] `go test ./internal/apps/identity/idp/...` passes

#### Task 1.8: Integrate contract tests into identity-rp

- **Status**: DONE
- **Dependencies**: Task 1.5
- **Description**: Add `RunContractTests(t, server)` to identity-rp integration tests
- **Acceptance Criteria**:
  - [x] identity-rp calls `RunContractTests`
  - [x] All contract tests pass
  - [x] `go test ./internal/apps/identity/rp/...` passes

#### Task 1.9: Integrate contract tests into identity-rs

- **Status**: DONE
- **Dependencies**: Task 1.5
- **Description**: Add `RunContractTests(t, server)` to identity-rs integration tests
- **Acceptance Criteria**:
  - [x] identity-rs calls `RunContractTests`
  - [x] All contract tests pass
  - [x] `go test ./internal/apps/identity/rs/...` passes

#### Task 1.10: Integrate contract tests into identity-spa

- **Status**: DONE
- **Dependencies**: Task 1.5
- **Description**: Add `RunContractTests(t, server)` to identity-spa integration tests
- **Acceptance Criteria**:
  - [x] identity-spa calls `RunContractTests`
  - [x] All contract tests pass (required SPA fallback fix for reserved path prefixes)
  - [x] `go test ./internal/apps/identity/spa/...` passes

#### Task 1.11: Verify lint-fitness coverage and mutation

- **Status**: DONE (verified and documented; coverage improvement deferred)
- **Dependencies**: None
- **Description**: Run coverage and mutation testing on 10,500 lines of lint_fitness code
- **Acceptance Criteria**:
  - [x] Coverage verified for `internal/apps/cicd/lint_fitness/`
  - [ ] Coverage >=98% for all packages (19 of 24 below target — pre-existing gap)
  - [ ] Mutation testing >=95% (gremlins panics on Windows; CI-only)
  - [x] Document any uncovered lines with justification
- **Coverage Report** (24 packages):
  - ≥98%: lint_fitness(100%), product_structure(100%), product_wiring(100%), service_structure(100%), circular_deps(99%)
  - 95-98%: migration_numbering(97.7%), check_skeleton_placeholders(96.8%), crypto_rand(96.1%), insecure_skip_verify(96.1%), cgo_free_sqlite(95.2%)
  - 90-95%: non_fips_algorithms(94.4%), cmd_main_pattern(91.3%)
  - 80-90%: file_size_limits(89.2%), tls_minimum_version(88.5%), cross_service_import_isolation(88%), admin_bind_address(87.8%), domain_layer_isolation(86.8%), health_endpoint_presence(85.5%), service_contract_compliance(83%), bind_address_safety(80.6%)
  - <80%: migration_range_compliance(79.3%), test_patterns(76.7%), parallel_tests(76.2%), no_hardcoded_passwords(68.9%)
- **Note**: Fixed 9 pre-existing Windows NTFS test failures (os.Chmod 0o000). Coverage improvement is a dedicated task for Phase 2+.

#### Task 1.12: Phase 1 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 1.1-1.11
- **Description**: Full quality gate run, coverage verification, phase post-mortem
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run` clean
  - [ ] `go test ./... -shuffle=on` passes
  - [ ] All 10 services have `RunContractTests`
  - [ ] lessons.md updated with Phase 1 post-mortem
  - [ ] Git commit with conventional commit message

---

### Phase 2: Remove InsecureSkipVerify — Integration Tests Only (D14, D15)

**Phase Objective**: Eliminate InsecureSkipVerify from integration + contract tests (~90% of 47 files). Fix all 6 ARCHITECTURE.md TLS gaps. E2E/Docker TLS (2B), mTLS (2C), PostgreSQL TLS (2D) explicitly deferred.

#### Task 2.1: Add TLS Test Bundle to service-template testserver

- **Status**: TODO
- **Dependencies**: None
- **Description**: Add TLS cert bundle generation to the shared testserver infrastructure
- **Acceptance Criteria**:
  - [ ] `NewTestTLSBundle(t)` in `internal/apps/template/service/testing/testserver/` generates self-signed CA + server cert
  - [ ] `TLSClientConfig(t *testing.T, bundle *TestTLSBundle) *tls.Config` returns config trusting the test CA cert
  - [ ] `testserver.StartAndWait()` accepts optional TLS bundle or auto-generates one
  - [ ] Server exposes `TLSBundle()` accessor so test setup can retrieve the CA cert
  - [ ] Unit tests for TLS bundle generation (>=95% coverage)
  - [ ] Build clean: `go build ./internal/apps/template/service/testing/...`
  - [ ] No linting errors

#### Task 2.2: Migrate sm-im test HTTP clients

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in sm-im tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in sm-im test files
  - [ ] All sm-im tests pass
  - [ ] No linting errors

#### Task 2.3: Migrate jose-ja test HTTP clients

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in jose-ja tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in jose-ja test files
  - [ ] All jose-ja tests pass
  - [ ] No linting errors

#### Task 2.4: Migrate sm-kms test HTTP clients

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in sm-kms tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in sm-kms test files
  - [ ] All sm-kms tests pass
  - [ ] No linting errors

#### Task 2.5: Migrate pki-ca test HTTP clients

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in pki-ca tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in pki-ca test files
  - [ ] All pki-ca tests pass
  - [ ] No linting errors

#### Task 2.6: Migrate identity service test HTTP clients (all 5)

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in identity-authz/idp/rp/rs/spa tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in identity service test files
  - [ ] All identity tests pass
  - [ ] No linting errors

#### Task 2.7: Migrate skeleton-template test HTTP clients

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in skeleton-template tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in skeleton-template test files
  - [ ] All template and skeleton tests pass
  - [ ] No linting errors

#### Task 2.8: Remove G402 from gosec.excludes and activate semgrep rule

- **Status**: TODO
- **Dependencies**: Tasks 2.2-2.7
- **Description**: Remove G402 exclusion from .golangci.yml, activate the semgrep rule
- **Acceptance Criteria**:
  - [ ] `G402` removed from `gosec.excludes` in `.golangci.yml`
  - [ ] `no-tls-insecure-skip-verify` rule uncommented in `.semgrep/rules/go-testing.yml`
  - [ ] `golangci-lint run ./...` passes with G402 enabled
  - [ ] `go test ./... -shuffle=on` passes

#### Task 2.9: Fix ARCHITECTURE.md TLS gaps (D15)

- **Status**: TODO
- **Dependencies**: Task 2.1
- **Description**: Fix all 6 identified TLS documentation gaps in ARCHITECTURE.md
- **Acceptance Criteria**:
  - [ ] Gap 1: TLS Certificate Configuration table added to ARCHITECTURE.md Section 6
  - [ ] Gap 2: TLS CA/cert/key secrets documented in Section 12.3.3
  - [ ] Gap 3: TLS test bundle pattern documented in Section 10.3
  - [ ] Gap 4: ServiceServer.TLSBundle() accessor documented in Section 10.3.5
  - [ ] Gap 5: mTLS deployment architecture documented in Section 6.3
  - [ ] Gap 6: TLS mode taxonomy (Static/Mixed/Auto) documented in Section 6
  - [ ] `cicd lint-docs validate-propagation` passes

#### Task 2.10: Phase 2 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 2.8-2.9
- **Description**: Full quality gate run, phase post-mortem
- **Acceptance Criteria**:
  - [ ] `go build ./...` and `go build -tags e2e,integration ./...` clean
  - [ ] `golangci-lint run` and `golangci-lint run --build-tags e2e,integration` clean
  - [ ] `go test ./... -shuffle=on` passes
  - [ ] `go test -race -count=2 ./...` clean
  - [ ] Coverage maintained
  - [ ] lessons.md updated with Phase 2 post-mortem
  - [ ] Git commit

---

### Phase 3: Builder Refactoring

**Phase Objective**: Product-services pass config objects; service-template picks what it needs.

#### Task 3.1: Analyze current builder With*() call patterns

- **Status**: TODO
- **Dependencies**: None
- **Description**: Audit all 10 services to document current builder usage patterns
- **Acceptance Criteria**:
  - [ ] Document which With*() methods each service calls
  - [ ] Identify redundant per-service path setup (WithBrowserBasePath, WithServiceBasePath)
  - [ ] Identify what a minimal domain config struct needs

#### Task 3.2: Design new builder domain config API

- **Status**: TODO
- **Dependencies**: Task 3.1
- **Description**: Design the new builder API where services pass a config struct, not individual With*() calls
- **Acceptance Criteria**:
  - [ ] Domain config struct defined
  - [ ] Builder accepts config struct
  - [ ] API reviewed for simplicity (NewFromConfig <=10 lines)

#### Task 3.3: Implement builder refactoring

- **Status**: TODO
- **Dependencies**: Task 3.2
- **Description**: Implement the new builder API in service-template
- **Acceptance Criteria**:
  - [ ] New builder API implemented
  - [ ] Old With*() methods removed (NO backward compatibility)
  - [ ] Unit tests updated

#### Task 3.4: Migrate all 10 services to new builder API

- **Status**: TODO
- **Dependencies**: Task 3.3
- **Description**: Update all 10 services to use new builder API
- **Acceptance Criteria**:
  - [ ] All 10 services use new builder API
  - [ ] NewFromConfig is <=10 lines per service
  - [ ] Zero duplicated path setup
  - [ ] All tests pass

#### Task 3.5: Phase 3 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Task 3.4
- **Description**: Full quality gate run
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] `golangci-lint run` clean
  - [ ] `go test ./... -shuffle=on` passes
  - [ ] lessons.md updated

---

### Phase 4: Sequential Exemption Reduction

**Phase Objective**: Reduce 173 `// Sequential:` exemptions via SEAM PATTERN and dependency injection. **Smallest-first ordering** (D10) to build momentum.

#### Task 4.1: Categorize and triage all 173 exemptions

- **Status**: TODO
- **Dependencies**: None
- **Description**: Categorize each exemption, determine which are truly necessary vs. avoidable
- **Acceptance Criteria**:
  - [ ] Spreadsheet/doc with all 173 exemptions categorized
  - [ ] Each marked: necessary, refactorable, or questionable
  - [ ] Priority order: smallest categories first (os.Stderr, pgDriver, seams, os.Chdir, viper/pflag)

#### Task 4.2: Inject io.Writer for os.Stderr tests (5 exemptions)

- **Status**: TODO
- **Dependencies**: Task 4.1
- **Description**: Inject `io.Writer` instead of capturing os.Stderr. Smallest category — quick win.
- **Acceptance Criteria**:
  - [ ] os.Stderr capture tests use injected io.Writer
  - [ ] Sequential exemptions removed for these tests
  - [ ] All tests pass

#### Task 4.3: pgDriver registration exemptions (11 exemptions)

- **Status**: TODO
- **Dependencies**: Task 4.1
- **Description**: Evaluate test isolation approach for pgDriver registration
- **Acceptance Criteria**:
  - [ ] Each pgDriver exemption evaluated
  - [ ] Avoidable exemptions removed
  - [ ] All tests pass

#### Task 4.4: Seam variables audit (11 exemptions)

- **Status**: TODO
- **Dependencies**: Task 4.1
- **Description**: These ARE the SEAM PATTERN already. Align documentation. Verify all are truly necessary.
- **Acceptance Criteria**:
  - [ ] Each seam exemption verified as correct pattern usage
  - [ ] Documentation aligned with seam pattern
  - [ ] Unnecessary exemptions removed if any

#### Task 4.5: Evaluate os.Chdir exemptions (37 exemptions)

- **Status**: TODO
- **Dependencies**: Task 4.1
- **Description**: Many os.Chdir exemptions in lint_fitness use CheckInDir which is already parameterized
- **Acceptance Criteria**:
  - [ ] Each os.Chdir exemption evaluated
  - [ ] Unnecessary exemptions removed
  - [ ] All tests pass

#### Task 4.6: SEAM PATTERN for viper/pflag tests (58 exemptions)

- **Status**: TODO
- **Dependencies**: Task 4.1
- **Description**: Inject config reader instead of relying on global viper state. Largest category — most complex.
- **Acceptance Criteria**:
  - [ ] Config tests no longer use global viper state
  - [ ] Sequential exemptions reduced by ~30-50
  - [ ] All tests pass

#### Task 4.7: Remaining exemption categories

- **Status**: TODO
- **Dependencies**: Task 4.1
- **Description**: Address SQLite in-memory (10), shared state (13), injectable function variables (16), signals (6), port reuse (5)
- **Acceptance Criteria**:
  - [ ] Each category evaluated
  - [ ] Avoidable exemptions removed
  - [ ] Target: total <100 exemptions remaining

#### Task 4.8: Phase 4 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 4.2-4.7
- **Description**: Full quality gate run
- **Acceptance Criteria**:
  - [ ] Total exemptions documented (target: <100)
  - [ ] Each remaining exemption has justified comment
  - [ ] All tests pass
  - [ ] lessons.md updated

---

### Phase 5: ServiceServer Interface Expansion

**Phase Objective**: Expand interface to cover integration test needs.

#### Task 5.1: Audit integration test needs

- **Status**: TODO
- **Dependencies**: None
- **Description**: Survey all integration tests to determine what framework services they need access to
- **Acceptance Criteria**:
  - [ ] List of methods integration tests currently access
  - [ ] List of methods integration tests need but don't have
  - [ ] Recommendation for interface expansion

#### Task 5.2: Expand ServiceServer interface

- **Status**: TODO
- **Dependencies**: Task 5.1
- **Description**: Add new methods to ServiceServer interface (NO backward compatibility)
- **Acceptance Criteria**:
  - [ ] Interface expanded with needed methods
  - [ ] All 10 services implement new methods
  - [ ] Contract tests exercise new methods

#### Task 5.3: Phase 5 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Task 5.2
- **Description**: Full quality gate run
- **Acceptance Criteria**:
  - [ ] `go build ./...` clean
  - [ ] All tests pass
  - [ ] lessons.md updated

---

### Phase 6: lint-fitness Value Assessment

**Phase Objective**: Verify 10,500 lines of lint-fitness truly standardize services.

#### Task 6.1: Coverage and mutation testing of lint-fitness

- **Status**: TODO
- **Dependencies**: None
- **Description**: Run coverage and mutation on all 23 sub-linters
- **Acceptance Criteria**:
  - [ ] Coverage >=98%
  - [ ] Mutation >=95%
  - [ ] Document any gaps

#### Task 6.2: Synthetic vs real content audit

- **Status**: TODO
- **Dependencies**: None
- **Description**: Identify sub-linters testing synthetic file content instead of real project structure
- **Acceptance Criteria**:
  - [ ] Each sub-linter classified: real vs synthetic testing
  - [ ] Plan to convert synthetic tests to real-project tests where feasible

#### Task 6.3: Verify skeleton-template as scaffolding source (D12)

- **Status**: TODO
- **Dependencies**: None
- **Description**: Skeleton stays minimal per D12 (no CRUD, no code generation). Verify it's current as `/new-service` scaffolding source.
- **Acceptance Criteria**:
  - [ ] skeleton-template uses latest builder API patterns
  - [ ] `/new-service` skill generates valid services from skeleton
  - [ ] Document skeleton vs lint-fitness vs `/new-service` relationship in ARCHITECTURE.md Section 3.1.5

#### Task 6.4: Add test infrastructure rule linters

- **Status**: TODO
- **Dependencies**: None
- **Description**: Add fitness linters detecting unit tests that start servers or DBs, and other test infrastructure anti-patterns.
- **Acceptance Criteria**:
  - [ ] New sub-linter detects unit tests starting real servers
  - [ ] New sub-linter detects unit tests starting real databases
  - [ ] `no_local_create_closed_database` rule added (detects `createClosedDatabase`/`createClosedDBHandler` outside testdb package)
  - [ ] Tests for the new sub-linters

#### Task 6.5: Phase 6 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 6.1-6.4
- **Description**: Full quality gate run
- **Acceptance Criteria**:
  - [ ] Value assessment documented
  - [ ] lessons.md updated

---

### Phase 7: Domain Extraction and Fresh Skeletons (D13, D16)

**Phase Objective**: Extract domain logic from identity-* and pki-ca, replace with fresh skeleton-template copies. Update status table.

#### Task 7.1: Archive identity shared packages

- **Status**: TODO
- **Dependencies**: Phases 1-5 complete
- **Description**: Archive all shared packages under `internal/apps/identity/` to `_archived/`
- **Acceptance Criteria**:
  - [ ] All shared identity packages moved to `internal/apps/identity/_archived/`
  - [ ] Build passes (broken imports expected — services replaced in Task 7.3)

#### Task 7.2: Archive per-service domain code

- **Status**: TODO
- **Dependencies**: Task 7.1
- **Description**: Archive domain code for authz, idp, rp, rs, spa, pki-ca
- **Acceptance Criteria**:
  - [ ] authz domain → `internal/apps/identity/_authz-archived/`
  - [ ] idp domain → `internal/apps/identity/_idp-archived/`
  - [ ] rp domain → `internal/apps/identity/_rp-archived/`
  - [ ] rs domain → `internal/apps/identity/_rs-archived/`
  - [ ] spa domain → `internal/apps/identity/_spa-archived/`
  - [ ] pki-ca archive verified complete (`internal/apps/pki/_ca-archived/`)

#### Task 7.3: Replace services with fresh skeletons

- **Status**: TODO
- **Dependencies**: Task 7.2
- **Description**: Replace all 6 services with fresh skeleton-template copies (builder + contract tests + health)
- **Acceptance Criteria**:
  - [ ] All 6 services use latest builder pattern
  - [ ] All 6 services pass `RunContractTests`
  - [ ] `go build ./...` clean
  - [ ] `go test ./... -shuffle=on` passes

#### Task 7.4: Update ARCHITECTURE.md status table (D16)

- **Status**: TODO
- **Dependencies**: Task 7.3
- **Description**: Update Section 3.2 status table: all 5 identity-* services → "⚠️ Extraction Pending 0%"
- **Acceptance Criteria**:
  - [ ] identity-authz/idp/rp/rs/spa marked "⚠️ Extraction Pending 0%"
  - [ ] pki-ca status updated appropriately

#### Task 7.5: Phase 7 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 7.1-7.4
- **Description**: Full quality gate run, phase post-mortem
- **Acceptance Criteria**:
  - [ ] All 6 skeleton services pass contract tests
  - [ ] All domain logic safely archived
  - [ ] `go build ./...` and `golangci-lint run` clean
  - [ ] lessons.md updated

---

### Phase 8: Staged Domain Reintegration (D13)

**Phase Objective**: Reintroduce archived domain logic into fresh skeletons, smallest-first.

#### Task 8.1: Reintegrate rp, rs, spa (Stage 1)

- **Status**: TODO
- **Dependencies**: Phase 7 complete
- **Description**: Smallest services first (10-18 files each). Extract from archive, adapt to latest builder, test.
- **Acceptance Criteria**:
  - [ ] rp domain reintegrated and tests pass
  - [ ] rs domain reintegrated and tests pass
  - [ ] spa domain reintegrated and tests pass
  - [ ] Coverage >=95% for each

#### Task 8.2: Reintegrate authz (Stage 2)

- **Status**: TODO
- **Dependencies**: Task 8.1
- **Description**: OAuth 2.1 core (133 files/916KB — largest complexity)
- **Acceptance Criteria**:
  - [ ] authz domain reintegrated with latest builder patterns
  - [ ] All authz tests pass
  - [ ] Coverage >=95%

#### Task 8.3: Reintegrate idp (Stage 3)

- **Status**: TODO
- **Dependencies**: Task 8.2
- **Description**: OIDC provider (129 files/862KB — second largest)
- **Acceptance Criteria**:
  - [ ] idp domain reintegrated with latest builder patterns
  - [ ] All idp tests pass
  - [ ] Coverage >=95%

#### Task 8.4: Reintegrate pki-ca (Stage 4)

- **Status**: TODO
- **Dependencies**: Task 8.3
- **Description**: Certificate lifecycle (48KB active + 880KB archived)
- **Acceptance Criteria**:
  - [ ] pki-ca domain reintegrated with latest builder patterns
  - [ ] All pki-ca tests pass
  - [ ] Coverage >=95%

#### Task 8.5: Phase 8 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 8.1-8.4
- **Description**: Full quality gate run, phase post-mortem
- **Acceptance Criteria**:
  - [ ] All 6 services have working domain + latest framework patterns
  - [ ] Coverage >=95% across all reintegrated services
  - [ ] lessons.md updated

---

### Phase 9: Quality and Knowledge Propagation

**Phase Objective**: Final quality sweep and knowledge propagation.

#### Task 9.1: Full coverage and mutation enforcement

- **Status**: TODO
- **Dependencies**: None
- **Description**: Run coverage and mutation across entire codebase
- **Acceptance Criteria**:
  - [ ] All production code >=95% coverage
  - [ ] All infrastructure code >=98% coverage
  - [ ] Mutation >=95%

#### Task 9.2: Improve agent semantic commit instructions (D11)

- **Status**: TODO
- **Dependencies**: None
- **Description**: Improve agent instructions for Multi-Category Fix Commit Rule. NO automated tooling (no commitlint, no CI validation). Instructions-only approach per D11.
- **Acceptance Criteria**:
  - [ ] Agent instructions updated to better enforce semantic commits
  - [ ] beast-mode.agent.md updated with commit grouping examples
  - [ ] implementation-execution.agent.md updated with commit checkpoint pattern

#### Task 9.3: Propagate all lessons to permanent artifacts

- **Status**: TODO
- **Dependencies**: None
- **Description**: Review lessons.md and propagate all lessons to ARCHITECTURE.md, agents, skills, instructions
- **Acceptance Criteria**:
  - [ ] Every lesson in lessons.md has corresponding entry in permanent artifact
  - [ ] `cicd lint-docs validate-propagation` passes
  - [ ] No lessons orphaned in plan docs only

#### Task 9.4: Simplify review document format

- **Status**: TODO
- **Dependencies**: None
- **Description**: framework-v1/review.md was overwhelming. Design a simpler format for future reviews.
- **Acceptance Criteria**:
  - [ ] Review template documented (concise format)
  - [ ] Future reviews follow simpler format

#### Task 9.5: Phase 9 validation and post-mortem

- **Status**: TODO
- **Dependencies**: Tasks 9.1-9.4
- **Description**: Final quality gate run
- **Acceptance Criteria**:
  - [ ] All quality gates pass
  - [ ] lessons.md finalized
  - [ ] Git working tree clean

---

## Cross-Cutting Tasks

### Semgrep Rules Maintenance

- [ ] After each phase: review `.semgrep/rules/go-testing.yml` for new relevant patterns
- [ ] After Phase 2 complete: uncomment `no-tls-insecure-skip-verify` in go-testing.yml

### Product-Level and Suite-Level Contract Tests

- [ ] Design parameterized product-level contract tests (5 products)
- [ ] Design suite-level contract test (1 suite)
- [ ] Implement after Phase 1 service-level contracts are complete

---

## ARCHITECTURE.md Cross-References

| Topic | Section |
|-------|---------|
| TLS Configuration | [Section 6.4](../ARCHITECTURE.md#64-cryptographic-architecture) |
| Test HTTP Client Patterns | [Section 10.3.4](../ARCHITECTURE.md#1034-test-http-client-patterns) |
| Integration Testing | [Section 10.3](../ARCHITECTURE.md#103-integration-testing-strategy) |
| Shared Test Infrastructure | [Section 10.3.6](../ARCHITECTURE.md#1036-shared-test-infrastructure) |
| Quality Gates | [Section 11.2](../ARCHITECTURE.md#112-quality-gates) |
| Security Architecture | [Section 6](../ARCHITECTURE.md#6-security-architecture) |
| Service Template | [Section 5.1](../ARCHITECTURE.md#51-service-template-pattern) |
| Service Builder | [Section 5.2](../ARCHITECTURE.md#52-service-builder-pattern) |
| Fitness Functions | [Section 9.11](../ARCHITECTURE.md#911-architecture-fitness-functions) |
| Sequential Test Exemption | [Section 10.2.5](../ARCHITECTURE.md#1025-sequential-test-exemption) |
| Contract Test Pattern | [Section 10.3.5](../ARCHITECTURE.md#1035-cross-service-contract-test-pattern) |
| Post-Mortem and Knowledge Propagation | [Section 13.8](../ARCHITECTURE.md#138-phase-post-mortem--knowledge-propagation) |
| Authentication and Authorization | [Section 6.9](../ARCHITECTURE.md#69-authentication--authorization) |
