# Lessons Learned - fixes-v7

**Source**: 11 phases, 220 tasks, 28+ commits across fixes-v7 plan execution.

## What Worked

### 1. Phased Commit Strategy
Splitting large changes into atomic commits per phase made bisect and review trivial.
The cipher→sm-im rename (200+ files) was cleanly split into code (130 files),
deployment (63 files), and docs (4 files) — each independently reviewable and revertable.

### 2. Test Seam Injection Pattern
Package-level seam variables with `saveRestoreSeams(t)` / `resetSeams()` helpers
enabled testing of error paths in third-party library wrappers without mocks or
interfaces. This pushed jose package coverage from 90% to 95.3% by making Set(),
Import(), Marshal(), PublicKey(), and keygen error paths testable.

### 3. Coverage Profiling Before Setting Targets
Running `go tool cover -html=coverage.out` and categorizing uncovered lines into
"structurally testable" vs "unreachable" BEFORE setting targets prevented chasing
impossible goals. The JWX-COV-CEILING.md analysis proved that 95% was achievable
but 100% was not (4 JWE OKP statements are truly unreachable).

### 4. Evidence-Based Task Completion
Requiring `go build`, `golangci-lint run`, and `go test` evidence before marking
any task complete caught issues early. The quality mandate in plan.md prevented
"LGTM" commits that would have failed CI/CD.

### 5. Table-Driven Tests with t.Parallel()
Converting standalone test functions to table-driven tests with t.Parallel() on
both parent and subtests improved both test quality and execution time. The jose
seam injection tests run 5 keygen error subtests in parallel.

### 6. Deployment Validators as Code
Having 62 deployment validators that run via `cicd lint-deployments validate-all`
caught config issues (port conflicts, missing secrets, naming violations) immediately
during the cipher→sm-im rename. Without validators, those issues would have reached CI/CD.

## What Didn't Work

### 1. Leaving Infrastructure Blockers Unresolved Across Plans
The OTel Docker socket issue (`resourcedetection` with `docker` detector) blocked
E2E tests across fixes-v1, v6, AND v7. Each plan documented it as "pre-existing"
and moved on. Three plans later, it's still unresolved. Infrastructure blockers
should be fixed immediately, not deferred.

### 2. Plan Proliferation (fixes-v1 through v7)
Seven plan iterations meant significant overlap, stale references, and confusion
about what was actually incomplete. The fixes-v7 "consolidation" itself took a
full phase just to audit and merge prior plans. A single living plan with phases
would have been more maintainable.

### 3. Large tasks.md Files
At 896 lines, tasks.md became unwieldy. Finding specific tasks required grep.
Phases should each have their own task file, or tasks should be broken into
smaller per-phase documents.

### 4. Coverage Targets Without Structural Analysis
Setting "≥95% coverage" as a blanket target without analyzing which packages have
structural ceilings (third-party library wrappers, error-only branches in generated
code) created frustration. Coverage targets should be per-package with documented
justifications for exceptions.

### 5. Amending Commits Instead of Incremental
Early phases occasionally used `git commit --amend` to "clean up" commits. This
lost context for bisect and made it impossible to understand the evolution of
changes within a phase. Incremental commits with conventional prefixes are always
better.

## Start Doing

### 1. Fix Infrastructure Blockers First
Before any feature/quality work, resolve all infrastructure issues (Docker config,
CI/CD pipeline, OTel, etc.). These block downstream validation and compound across
plans.

### 2. Per-Package Coverage Targets
Document expected coverage ceiling per package based on structural analysis. Track
in a coverage-targets.md or similar living document. Accept documented exceptions
rather than blanket percentages.

### 3. Single Living Plan
Maintain ONE plan.md that evolves. Archive completed phases by moving them to an
archive section within the same file (or a dated archive file). Never create
fixes-vN+1 — instead, add new phases to the existing plan.

### 4. Pre-Change Impact Analysis for Renames
Before large renames, run `grep -r "old-name"` across all file types (Go, YAML,
Dockerfile, Markdown, workflows) and document the full blast radius. The cipher→sm-im
rename touched 200+ files; knowing the count upfront enabled accurate phasing.

### 5. Test Seams as First-Class Pattern
Document the test seam injection pattern in ARCHITECTURE.md Section 10 (Testing).
It's a repeatable pattern that other packages (pki-ca, identity services) will need.

## Stop Doing

### 1. Deferring E2E Blockers
Never mark a plan "complete" with known E2E blockers. Either fix the blocker or
explicitly descope E2E from the plan's acceptance criteria.

### 2. Creating Multiple Plan Versions
fixes-v1 through v7 is seven iterations of the same work. Use one plan with phases.

### 3. Using tasks.md for Both Tracking AND Documentation
tasks.md mixed task checkboxes with detailed technical analysis, root cause
descriptions, and implementation notes. Keep tasks.md as a pure checklist; put
analysis in separate documents (or in the plan itself).

### 4. Blanket Coverage Mandates Without Analysis
"≥95% for all production code" doesn't account for structural ceilings,
generated code, or third-party library wrappers. Set targets after analysis.

### 5. Trusting Line Numbers in Multi-Step Edits
When making multiple edits to the same file, line numbers shift after each edit.
Always re-read the file between edits or use pattern-based matching instead of
line-number-based sed commands.

## ARCHITECTURE.md Gaps Identified

Based on fixes-v7 experience, the following ARCHITECTURE.md sections need updates:

1. **Section 10 (Testing)**: No documentation of test seam injection pattern.
   This was the key technique for pushing coverage past structural ceilings.

2. **Section 9.3/9.4 (Observability)**: No mention of OTel collector Docker socket
   requirement or the `resourcedetection` processor configuration. This caused
   a cross-plan blocker.

3. **Section 12.7 (Documentation Propagation)**: Mapping table has only 5 entries
   but there are 148 cross-references across 18 instruction files. The mapping
   is severely incomplete.

4. **Section 10.2 (Unit Testing)**: No guidance on coverage ceiling analysis
   methodology. Teams need to know how to identify structurally untestable code.

5. **Section 7 (Data Architecture)**: No mention of `saveRestoreSeams` pattern for
   database test isolation. The pattern could apply to repository testing.

6. **Section 11 (Quality)**: Coverage targets are blanket percentages (≥95%/≥98%).
   Should reference per-package analysis with documented exceptions.

7. **Section 13 (Development Practices)**: No guidance on plan lifecycle management.
   fixes-v1 through v7 proliferation shows the need for a "single living plan" pattern.
