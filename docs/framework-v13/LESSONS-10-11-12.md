# Lessons Learned — Framework v10, v11, v12 (Consolidated)

**Created**: 2026-06-29
**Purpose**: Consolidate all lessons learned from Framework v10 (Canonical Template Registry), v11 (PKI-Init Cert Structure), and v12 (PostgreSQL mTLS + Private App mTLS Trust) into a single prioritized reference for v13 planning and execution.

---

## Executive Summary

Three framework versions produced 102 tasks, 18 formal design decisions, and significant infrastructure. The lessons cluster into five themes: (1) Docker-dependent verification must not be deferred — it accumulates as unpaid debt, (2) lessons must be captured during each phase, not after, (3) scope management via quizme rounds works but needs limits, (4) seam injection and table-driven tests are proven patterns that should be mandatory, and (5) cross-document consistency requires automated enforcement. The 25 lessons below are prioritized by impact on v13 execution quality.

1. [Docker verification MUST be in-scope, never deferred](#1-docker-verification-must-be-in-scope-never-deferred)
2. [Capture lessons during each phase, not after](#2-capture-lessons-during-each-phase-not-after)
3. [Every config change needs runtime verification](#3-every-config-change-needs-runtime-verification)
4. [Deferred work must be explicitly assigned to a future version](#4-deferred-work-must-be-explicitly-assigned-to-a-future-version)
5. [E2E tests are mandatory for CLI entry points with productionNew* functions](#5-e2e-tests-are-mandatory-for-cli-entry-points-with-productionnew-functions)
6. [Struct-field seam injection is the proven pattern for testability](#6-struct-field-seam-injection-is-the-proven-pattern-for-testability)
7. [The internalMain pattern raises coverage ceilings](#7-the-internalmain-pattern-raises-coverage-ceilings)
8. [Table-driven tests with atomic counter injection eliminate mock libraries](#8-table-driven-tests-with-atomic-counter-injection-eliminate-mock-libraries)
9. [quizme rounds should be limited to 2 per plan](#9-quizme-rounds-should-be-limited-to-2-per-plan)
10. [Project-wide decisions belong in ENG-HANDBOOK.md, not plan.md](#10-project-wide-decisions-belong-in-eng-handbookmd-not-planmd)
11. [Derive directory/file counts from patterns, never estimate](#11-derive-directoryfile-counts-from-patterns-never-estimate)
12. [When accepting shared-identity decisions, trace the full CA signing chain](#12-when-accepting-shared-identity-decisions-trace-the-full-ca-signing-chain)
13. [os.Stat before os.ReadDir for cross-platform compatibility](#13-osstat-before-osreaddir-for-cross-platform-compatibility)
14. [PKCS#12 Modern (SHA-256/AES-256-CBC) is unconditionally preferred](#14-pkcs12-modern-sha-256aes-256-cbc-is-unconditionally-preferred)
15. [Comment category numbers at generation call sites](#15-comment-category-numbers-at-generation-call-sites)
16. [Inject all I/O dependencies as function fields from the start](#16-inject-all-io-dependencies-as-function-fields-from-the-start)
17. [Truststores never contain .key files — keystores do](#17-truststores-never-contain-key-files--keystores-do)
18. [SAME-AS-DIR-NAME convention eliminates secondary naming decisions](#18-same-as-dir-name-convention-eliminates-secondary-naming-decisions)
19. [stripQueryParam before appending DSN parameters](#19-stripqueryparam-before-appending-dsn-parameters)
20. [Template-compliance and lint-deployments must run after every deployment change](#20-template-compliance-and-lint-deployments-must-run-after-every-deployment-change)
21. [ENG-HANDBOOK.md edits require lint-docs verification](#21-eng-handbookmd-edits-require-lint-docs-verification)
22. [Estimation bias: documentation/verification phases take 50% of estimated time](#22-estimation-bias-documentationverification-phases-take-50-of-estimated-time)
23. [Track estimated vs actual hours per phase](#23-track-estimated-vs-actual-hours-per-phase)
24. [Standardize task status notation across all plans](#24-standardize-task-status-notation-across-all-plans)
25. [Named Docker volumes only — never bind mounts for certs](#25-named-docker-volumes-only--never-bind-mounts-for-certs)

---

## Details

### 1. Docker verification MUST be in-scope, never deferred

**Source**: v12 Phases 3, 6, 9, 10.5
**Priority**: CRITICAL

v12 wrote PostgreSQL TLS/mTLS configuration across 6+ files but deferred all Docker-based verification. The result: 38/43 tasks "complete" but zero runtime confidence. Configuration-only changes without Docker verification are untested hypotheses, not completed work.

**Rule for v13**: Every phase that modifies Docker Compose files, config files consumed by containers, or cert mount paths MUST include a Docker Compose verification step within the same phase. Docker Desktop must be running. If Docker is unavailable, the phase is blocked — not complete.

---

### 2. Capture lessons during each phase, not after

**Source**: v10 lessons.md (empty), v12 lessons.md (sparse)
**Priority**: CRITICAL

v10 completed 33/33 tasks but captured zero lessons — the lessons.md file has only empty placeholders. v12 captured lessons but in a terse 2-4 bullet format without root cause analysis. v11 captured detailed lessons with "What Worked / What Didn't Work / Root Causes / Patterns" structure and is the only version where lessons are genuinely useful.

**Rule for v13**: Update lessons.md at the end of each phase using v11's 4-section structure. Mark a phase as "complete" only after lessons are written. The lessons.md update is not a separate "Knowledge Propagation" phase — it is part of each phase's definition of done.

---

### 3. Every config change needs runtime verification

**Source**: v12 PostgreSQL TLS configuration
**Priority**: HIGH

PostgreSQL TLS involves postgresql.conf (`ssl_*` settings), pg_hba.conf (hostssl rules with `clientcert=verify-full`), GORM DSN parameters (`sslmode`, `sslcert`, `sslkey`, `sslrootcert`), and Docker volume mounts. Each file is independently correct, but the integrated chain was never tested. Common failure modes: wrong cert file paths after Docker volume mounting, permission errors inside containers, HBA rule ordering issues, GORM parameter mismatches.

**Rule for v13**: For any multi-file configuration change, include an integration verification step that exercises the full chain in a running environment.

---

### 4. Deferred work must be explicitly assigned to a future version

**Source**: v12 deferred tasks, v11 deferred mutation/race testing
**Priority**: HIGH

v12's 5 deferred tasks say "requires Docker" but are not assigned to v13 or any other version. v11's deferred mutation/race testing says "deferred to Linux CI/CD" but no CI/CD workflow was created. Unassigned deferrals become permanent gaps.

**Rule for v13**: Any deferred work must name the target version and be tracked in that version's plan.md as an explicit prerequisite phase.

---

### 5. E2E tests are mandatory for CLI entry points with productionNew* functions

**Source**: v11 Phase 3 lessons (pki-init CLI bugs)
**Priority**: HIGH

v11's `productionNewTelemetryService` and `productionNewGenerator` functions are bypassed by unit tests (which use stub injection). Three bugs were only found during E2E testing: missing LogLevel field, missing OTLPEndpoint field, and a cert validity off-by-one. Unit tests with stubs provide structural coverage but cannot catch initialization-time configuration errors.

**Rule for v13**: Every CLI entry point that constructs production dependencies (telemetry, database connections, TLS config) MUST have at least one E2E smoke test that exercises the full initialization path.

---

### 6. Struct-field seam injection is the proven pattern for testability

**Source**: v11 Phase 2 lessons (Generator), v12 Phase 10 lessons (applyAdminMTLS)
**Priority**: HIGH

v11's Generator has 8 function fields for I/O dependencies, all injected via struct fields. v12's admin mTLS uses `osReadFileFn` seam injection. Both achieved high coverage with zero mock libraries and zero `//nolint` directives. This pattern is codified in `03-02.testing.instructions.md §10.2.4`.

**Rule for v13**: All new code with I/O dependencies MUST use struct-field seam injection. No mock libraries. No package-level var injection (except `os.Exit`).

---

### 7. The internalMain pattern raises coverage ceilings

**Source**: v11 Phase 5 coverage ceiling analysis (92.4% vs 95% target)
**Priority**: MEDIUM

v11 accepted 92.4% coverage because `productionNew*` functions are only exercisable via E2E. The `internalMain` pattern (thin main() delegates to testable `internalMain(args, stdin, stdout, stderr)`) could wrap the pki-init CLI logic, making initialization testable in unit tests.

**Rule for v13**: New CLI entry points MUST use the `internalMain` pattern from the start. Existing CLI entry points should be migrated when touched.

---

### 8. Table-driven tests with atomic counter injection eliminate mock libraries

**Source**: v11 Phase 5 lessons (48 unit tests, 14 categories)
**Priority**: MEDIUM

v11's pki-init tests use `sync/atomic` int32 counters in stub functions to verify call counts. Pattern: `if atomic.AddInt32(&callCount, 1) == wantFailAt { return error }`. This verifies "function called exactly N times" without mockery or testify/mock.

**Rule for v13**: Use atomic counters for call-count verification. Never introduce external mock libraries.

---

### 9. quizme rounds should be limited to 2 per plan

**Source**: v10 (4 rounds, 32 questions, 18 decisions)
**Priority**: MEDIUM

v10 required 4 quizme rounds before implementation started. Some decisions were updated 3 times across rounds. While design questions prevented architectural mistakes, the overhead was disproportionate for a template registry.

**Rule for v13**: Maximum 2 quizme rounds (initial + clarification). If a third round seems needed, the plan scope is too large — split it.

---

### 10. Project-wide decisions belong in ENG-HANDBOOK.md, not plan.md

**Source**: v10 Decision 18 ("Docker Compose profiles BANNED")
**Priority**: MEDIUM

v10's Decision 18 bans Docker Compose profiles. This is a project-wide policy, not a v10-specific decision. Embedding it in a plan.md means it's discovered only by reading that specific plan.

**Rule for v13**: Any decision that applies beyond the current plan MUST be added to ENG-HANDBOOK.md. Plan-specific decisions reference the ENG-HANDBOOK.md section.

---

### 11. Derive directory/file counts from patterns, never estimate

**Source**: v11 Phase 1 lessons (120 → 86 directory count)
**Priority**: MEDIUM

v11's initial count estimate (120 directories) was wrong. After applying design decisions (removing leaf cert truststores, shared postgres identity), the count dropped to 86 for PS-ID scope. The correct approach: expand the parametric pattern (`{a,b} × {1,2}`) and count the results.

**Rule for v13**: All directory/file count claims MUST be derived from pattern expansion, with the expansion formula shown. Never state a count without showing how it was calculated.

---

### 12. When accepting shared-identity decisions, trace the full CA signing chain

**Source**: v11 Phase 1 lessons (Q4 gap discovered in quizme-v2)
**Priority**: MEDIUM

v11's Decision Q1 (postgres instances share identity) was accepted without tracing: "Which CA signs the shared cert? How do all recipients trust it?" This gap was caught only in quizme-v2, after Phase 1 was marked complete.

**Rule for v13**: When a design decision involves shared identities or certificates, immediately trace: (1) Which CA issues the cert? (2) Which truststores contain that CA? (3) Is the trust chain complete for all consumers?

---

### 13. os.Stat before os.ReadDir for cross-platform compatibility

**Source**: v11 Phase 2 lessons (Windows os.ReadDir bug)
**Priority**: MEDIUM

`os.ReadDir` on a non-existent path returns an error on Windows (not an empty slice as on Unix). Always check `os.Stat` first. This is a documented Go behavior difference, not a bug.

**Rule for v13**: All file/directory existence checks MUST use `os.Stat` with `errors.Is(err, fs.ErrNotExist)` before attempting reads.

---

### 14. PKCS#12 Modern (SHA-256/AES-256-CBC) is unconditionally preferred

**Source**: v11 Phase 2 lessons
**Priority**: LOW

`pkcs12.Modern.Encode` / `pkcs12.Modern.EncodeTrustStore` from `software.sslmate.com/src/go-pkcs12` uses SHA-256/AES-256-CBC. Never use `pkcs12.Legacy` (3DES). CGO-free, FIPS-aligned.

---

### 15. Comment category numbers at generation call sites

**Source**: v11 Phase 2 lessons
**Priority**: LOW

When implementing a multi-category generate function (like pki-init's 14 categories), add `// Cat N: <name>` comments at each invocation site. Reviewers can cross-reference tls-structure.md without mentally mapping the code.

---

### 16. Inject all I/O dependencies as function fields from the start

**Source**: v11 Phase 2 lessons (getRealmsForPSIDFn retro-fitted)
**Priority**: LOW

v11's `getRealmsForPSIDFn` was initially hardcoded to read registry.yaml from disk. Only during test writing was this discovered to be untestable. Retro-fitting required touching the `Generate` method signature.

**Rule**: For any function that does I/O (filesystem, network, registry), inject it as a function field from the start — do not wait until test-writing reveals it.

---

### 17. Truststores never contain .key files — keystores do

**Source**: v11 Phase 1 lessons
**Priority**: LOW

Keystores contain `.crt + .key + .p12` (leaf cert + private key). Truststores contain `.crt + .p12` only (CA cert chain, never private keys). This is exploitable in code: `writeKeystore(path, kp, cert, chain)` vs `writeTruststore(path, certs)` — different function signatures, no ambiguity.

---

### 18. SAME-AS-DIR-NAME convention eliminates secondary naming decisions

**Source**: v11 Phase 2 lessons
**Priority**: LOW

Files inside each directory are named identically to the directory name (e.g., `public-global-root-https-server-ca/public-global-root-https-server-ca.crt`). Eliminates the need for a second naming convention. Discoverable by consumers from directory path alone.

---

### 19. stripQueryParam before appending DSN parameters

**Source**: v12 Phase 4 lessons
**Priority**: LOW

When modifying PostgreSQL DSN strings (adding `sslmode`, `sslcert`, etc.), call `stripQueryParam` to remove any existing value before appending the new one. pgx uses first-value-wins semantics, so duplicate params indicate a code smell even though they might "work."

---

### 20. Template-compliance and lint-deployments must run after every deployment change

**Source**: v12 Phase 7-8 lessons
**Priority**: LOW

v12 discovered that lint-deployments covers 54 validators (not 8). Any deployment file change can trigger failures across validators that seem unrelated. Always run `go run ./cmd/cicd-lint lint-deployments` before committing deployment changes.

---

### 21. ENG-HANDBOOK.md edits require lint-docs verification

**Source**: v12 Phase 11 lessons
**Priority**: LOW

`replace_string_in_file` on ENG-HANDBOOK.md can silently delete section headings if the `oldString` includes the heading but the `newString` omits it. `lint-docs` catches broken anchors caused by deleted headings. Always run `go run ./cmd/cicd-lint lint-docs` after any ENG-HANDBOOK.md edit.

---

### 22. Estimation bias: documentation/verification phases take 50% of estimated time

**Source**: v11 Phase 4 (3h estimated → 0.5h actual), Phase 6 (2h → 0.5h)
**Priority**: LOW

Documentation and verification phases are consistently over-estimated because they assume new artifacts will be needed. In practice, they typically verify existing state or add small doc sections.

**Rule for v13**: Estimate doc/verification phases at 50% of implementation phase estimates, unless lessons explicitly identify new artifacts needed.

---

### 23. Track estimated vs actual hours per phase

**Source**: v12 (no actuals recorded)
**Priority**: LOW

v12's plan.md included time estimates per phase but tasks.md recorded no actual durations. Without actuals, estimation calibration is impossible.

**Rule for v13**: tasks.md should track estimated vs actual hours per phase to enable calibration for v14+.

---

### 24. Standardize task status notation across all plans

**Source**: v10/v11/v12 inconsistent notation
**Priority**: LOW

v10 uses `✅` only. v11 uses `✅` and `⚠️ PARTIAL`. v12 uses `✅`, `⏳ DEFERRED`, `☐ TODO`. Inconsistency makes cross-version scanning difficult.

**Standard for v13+**: `✅ COMPLETE`, `🔄 IN-PROGRESS`, `⏳ DEFERRED (reason)`, `☐ TODO`, `❌ BLOCKED (reason)`.

---

### 25. Named Docker volumes only — never bind mounts for certs

**Source**: v11 Phase 3 lessons, v12 Phase 7 lessons
**Priority**: LOW

Named volumes (`{PS-ID}-certs`) are Docker-native, portable, lifecycle-managed by Compose, and enforce least-privilege via read-only mounts. Bind mounts require host directory preparation and are host-path-dependent. Rules CO-21/CO-22 in deployment-templates.md enforce this.
