# Fixes v9 - Tasks

**Status**: 0 of N tasks complete
**Created**: 2026-03-01
**Last Updated**: 2026-03-01 (quizme-v1 + quizme-v2 + quizme-v3 + quizme-v4 answers merged)

## Quality Mandate - MANDATORY

Every task: ALL 8 quality attributes verified each review pass. All issues are blockers. Resources never a constraint.

---

## Phase 1: Quality Review Passes Rework

### Task 1.1: Update ARCHITECTURE.md Section 2.5 (canonical source)
- [ ] Rewrite review passes: each pass checks ALL 8 attributes
- [ ] Add continuation rule: "If pass 3 finds ANY issue, continue to pass 4"
- [ ] Add pass 4–5 rules and "diminishing returns" stop condition
- [ ] Change phrasing to "minimum 3, maximum 5"
- [ ] Confirm @propagate markers correctly placed
- **Files**: `docs/ARCHITECTURE.md` Section 2.5

### Task 1.2: Update beast-mode.instructions.md @source
- [ ] Update @source block to match new Section 2.5 verbatim
- [x] Run lint-docs to verify chain
- **Files**: `.github/instructions/01-02.beast-mode.instructions.md`

### Task 1.3: Update evidence-based.instructions.md @source
- [ ] Update @source block to match new Section 2.5 verbatim
- **Files**: `.github/instructions/06-01.evidence-based.instructions.md`

### Task 1.4–1.8: Update all 5 agent files
- [ ] `beast-mode.agent.md` — update review passes section
- [ ] `doc-sync.agent.md` — update review passes section
- [ ] `fix-workflows.agent.md` — update review passes section
- [ ] `implementation-execution.agent.md` — update review passes section
- [ ] `implementation-planning.agent.md` — update review passes section

### Task 1.9: Sweep for stale "3 passes" mentions
- [ ] `grep -rn "3 review passes\|3 sequential\|exactly 3\|three review" .github/ docs/ARCHITECTURE.md`
- [ ] Update all stale mentions

---

## Phase 2: Agent Semantic Analysis

### Task 2.1: beast-mode.agent.md dual labeling
- [ ] Add "Quality Gate Commands (Go Projects)" label/heading before Go-specific examples
- [ ] Verify generic execution principles use generic language ("build", "lint", "test")
- [ ] No Go-isms in generic sections
- **Files**: `.github/agents/beast-mode.agent.md`

### Task 2.2: Confirm other agents correctly scoped (read-only verify, no changes)
- [ ] doc-sync, fix-workflows, implementation-execution, implementation-planning — confirm domain-specific
- [ ] Document: confirmed, no scope changes needed

---

## Phase 3: ARCHITECTURE.md Optimization

### Task 3.1: Consolidate quality attributes → Section 11.1
- [x] Add @propagate marker in Section 11.1 around full quality attributes list
- [x] Replace quality attribute lists in 1.3 + 2.5 with @source blocks
- [ ] Run lint-docs to verify chain
- **Expected**: ~40–60 lines saved

### Task 3.2: CLI patterns canonical → Section 4.4.7
- [x] Section 4.4.7 → KEEP all CLI content (canonical)
- [x] Section 9.1 → remove duplicated HOW content; add cross-reference to 4.4.7
- [x] Check all instruction files referencing 9.1 for CLI patterns — update cross-references if needed
- **Expected**: ~30–50 lines saved

### Task 3.3: Port assignments → Section 3.4, DELETE Appendix B.1 + B.2
- [x] Read Section 3.4 — ensure it has complete service AND database port tables
- [x] Merge any rows from B.1/B.2 not yet in 3.4
- [x] DELETE Appendix B.1 (Service Port Assignments)
- [x] DELETE Appendix B.2 (Database Port Assignments)
- [x] Resequence: old B.3 → B.1, B.4 → B.2, etc.
- [x] Update TOC and all cross-references to old B.1/B.2 → Section 3.4
- [x] Update all cross-references to old B.3+ → new B.# numbers
- **Expected**: ~80–120 lines saved

### Task 3.4: Verify infrastructure blocker in both 13.7 and 2.5 (no deletion)
- [x] Read both sections — confirm consistent, no contradictions

### Task 3.5: Add new Copilot Skills section to ARCHITECTURE.md
- [x] Determine section number (new subsection in 2.X area)
- [x] Write section: VS Code Skills overview, `.github/skills/` organization, `SKILLNAME.md` naming, skill catalogue table (name, purpose, link), reference to VS Code docs
- [x] Add to TOC
- **Adds**: ~30–50 lines

### Task 3.6: Add agent/skill/instruction matrix to Section 2.1
- [x] Add concise 4-row decision table: Instructions / Agents / Skills — scope, trigger, best for
- [x] Cross-reference to skills section
- **Adds**: ~10–15 lines

### Task 3.7: Review pass count sweep (any missed occurrences)
- [x] `grep -n "3 passes\|3 review\|exactly 3\|minimum 3" docs/ARCHITECTURE.md`
- [x] Update remaining stale mentions

### Task 3.8: Verify line count target
- [x] `wc -l docs/ARCHITECTURE.md` — record before and after
- [x] Target: <4,000 lines
- [x] If not reached: identify additional prose candidates (no semantic loss)

---

## Phase 4: doc-sync Agent Propagation

### Task 4.1: Add Section 12.7 Documentation Propagation Strategy
- [x] Read ARCHITECTURE.md Section 12.7 — identify @propagate markers
- [x] Add @source block for 12.7 into doc-sync.agent.md
- [x] Add cross-reference header

### Task 4.2: Add Section 11.4 Documentation Standards
- [x] Read ARCHITECTURE.md Section 11.4 — identify @propagate markers
- [x] Add @source block for 11.4 into doc-sync.agent.md

### Task 4.3: Add Appendix B.6 Instruction File Reference
- [x] Complete Task 3.3 first (Appendix B resequencing)
- [x] Verify new B.# number for Instruction File Reference
- [x] Add @source block or cross-reference into doc-sync.agent.md

### Task 4.4: Verify Section 2.5 reference is current (after Phase 1)
- [x] Confirm doc-sync Section 2.5 reference is up to date

---

## Phase 5: Copilot Skills

### Task 5.0: Create .github/skills/ infrastructure
- [x] Create `.github/skills/` directory
- [x] Create `.github/skills/README.md` (naming convention, how to reference skills)
- [x] Create `.github/skills/SKILL-TEMPLATE.md` (reference template for new skills)
- [x] Add skills section to ARCHITECTURE.md if not done in Phase 3 (coordinate with 3.5)

### Task 5.1: Group A — Test Generation Skills
- [x] Create `.github/skills/test-table-driven.md`
- [x] Create `.github/skills/test-fuzz-gen.md`
- [x] Create `.github/skills/test-benchmark-gen.md`
- [x] Each skill: conventions, examples, required imports, project-specific rules

### Task 5.2: Group B — Infrastructure Skills
- [x] Create `.github/skills/migration-create.md` (numbered SQL files: template 1001-1999, domain 2001+)

### Task 5.3: Group C — Code Quality Skills
- [x] Create `.github/skills/coverage-analysis.md` (coverprofile analysis, categorize RED lines, test suggestions)
- [x] Create `.github/skills/fips-audit.md` (detect violations + fix guidance, not just detection)

### Task 5.4: Group D — Documentation Skills
- [x] Create `.github/skills/propagation-check.md` (detect drift, generate corrected @source text)
- [x] Create `.github/skills/openapi-codegen.md` (3 config files + OpenAPI spec skeleton for any service)

### Task 5.5: Group E — Scaffolding Skills (all 4: covers all 3 Copilot types + new-service)
- [x] Create `.github/skills/agent-scaffold.md` (creates .github/agents/NAME.agent.md with mandatory sections)
- [x] Create `.github/skills/instruction-scaffold.md` (creates .github/instructions/NN-NN.name.instructions.md)
- [x] Create `.github/skills/skill-scaffold.md` (creates .github/skills/NAME.md — 3rd type, was missing from v2)
- [x] Create `.github/skills/new-service.md` (guides service creation from skeleton-template; replaces new-service.agent.md per quizme-v3 S4-Item4)

### Task 5.6: Update ARCHITECTURE.md skills catalogue
- [x] Add all 13 skills to the catalogue table in Phase 3.5 section

### Task 5.7: Update relevant agents with skills: frontmatter
- [x] After all skills created, identify which agents benefit from skills: references
- [x] Update agent YAML frontmatter accordingly

---

## Phase 6: Pre-commit / Pre-push Linter Additions

**Status**: Decisions confirmed from quizme-v3. Ready for implementation.

### Task 6.1: Add checkov (IaC/Container Security)
- [x] Add `bridgecrewio/checkov` hook to `.pre-commit-config.yaml`
- [x] Configure to scan `deployments/` and Dockerfiles
- [x] Verify runs cleanly on existing codebase (`pre-commit run checkov --all-files`)

### Task 6.2: Add sqlfluff (SQL Migration Linting)
- [x] Add `sqlfluff-pre-commit` hook to `.pre-commit-config.yaml`
- [x] Create `.sqlfluff` config: `dialect = postgres`, consistent SQL style rules
- [x] Verify all existing `.sql` migration files pass (`pre-commit run sqlfluff-lint --all-files`)

### Task 6.3: Add taplo (TOML Formatter)
- [x] Add `CommaNet/taplo-pre-commit` hook to `.pre-commit-config.yaml`, hook id: `taplo-format`
- [x] Verify any TOML files format cleanly

### Task 6.4: Add pyproject-fmt (pyproject.toml Normalizer)
- [x] Add `tox-dev/pyproject-fmt` hook to `.pre-commit-config.yaml`
- [x] Verify `pyproject.toml` formats cleanly (run after Phase 7 ruff migration)

### Task 6.5: Add validate-pyproject (Schema Validation)
- [x] Add `abravalheri/validate-pyproject` hook to `.pre-commit-config.yaml`
- [x] Verify `pyproject.toml` passes schema validation

### Task 6.6: Create .editorconfig + add editorconfig-checker (quizme-v4 YES)
- [x] Create `.editorconfig` in project root with rules matching current editor standards (indent_style=space, indent_size=4 for Go, end_of_line=lf, charset=utf-8, trim_trailing_whitespace=true, insert_final_newline=true)
- [x] Add `editorconfig-checker/editorconfig-checker` hook to `.pre-commit-config.yaml`
- [x] Verify all existing files pass (`pre-commit run editorconfig-checker --all-files`); fix any violations

**Note**: ruff-check and ruff-format hooks implemented in Phase 7 (Python migration)

---

## Phase 7: Python Toolchain Modernization

**Status**: Decisions confirmed from quizme-v3. Ready for implementation.

### Task 7.1: Remove redundant Python tools from pyproject.toml
(Confirmed for removal per quizme-v3)
- [x] Remove `black` (replaced by `ruff format`)
- [x] Remove `isort` (replaced by ruff I rules)
- [x] Remove `flake8` (replaced by ruff E/W/F rules)

### Task 7.2: Add ruff to pyproject.toml
- [x] Add `ruff>=0.9.0` to dependencies
- [x] Add `[tool.ruff]` section: `line-length = 200`, `target-version = "py314"`, `select = ["E", "W", "F", "I", "B", "UP", "SIM", "C90"]`
- [x] Add `[tool.ruff.format]` section (black-compatible settings)
- [x] Review/adjust rule set based on existing code

### Task 7.3: Update pre-commit hooks
- [x] Remove: black hook, isort hook, flake8 hook
- [x] Add: `astral-sh/ruff-pre-commit` ruff (lint) hook
- [x] Add: `astral-sh/ruff-pre-commit` ruff-format hook
- [x] Verify all Python files pass ruff check + ruff format

### Task 7.4: Update CI/CD workflows for Python
- [x] Replace `black --check` / `isort --check` / `flake8` with `ruff check` + `ruff format --check`
- [x] Update `.github/workflows/*.yml` as applicable

### Task 7.5: Migrate to uvx for CLI tool execution
- [x] Update scripts that run ruff → `uvx ruff`
- [x] Update scripts that run mypy → `uvx mypy`
- [x] Update any pip-install-then-run patterns to uvx
- [x] Verify `uv` / `uvx` available in CI/CD (add to workflow setup if needed)

---

## Phase 8: Java Toolchain Additions (Gatling Load Tests)

**Status**: Decisions confirmed from quizme-v3. Ready for implementation.

### Task 8.1: Add Spotless + google-java-format
- [ ] Add `spotless-maven-plugin` to `test/load/pom.xml`
- [ ] Configure: google-java-format, phase=`validate`, apply on `mvn spotless:apply`
- [ ] Run `cd test/load && mvn spotless:check` — verify all `.java` files pass

### Task 8.2: Add Checkstyle
- [ ] Add `maven-checkstyle-plugin` to `test/load/pom.xml`
- [ ] Configure: Google Checkstyle rules, phase=`validate`, fail on violations
- [ ] Run `cd test/load && mvn checkstyle:check` — verify pass

### Task 8.3: Add Error Prone + NullAway
- [ ] Add Error Prone annotation processor to `maven-compiler-plugin` config
- [ ] Add NullAway as Error Prone plugin
- [ ] Run `cd test/load && mvn compile` — verify zero Error Prone violations

### Task 8.4: Add maven-enforcer-plugin
- [ ] Add `maven-enforcer-plugin` to `test/load/pom.xml`
- [ ] Rules: `dependencyConvergence`, `requireJavaVersion` (21+), `requireMavenVersion` (3.9+)
- [ ] Run `cd test/load && mvn enforcer:enforce` — verify pass

### Task 8.5: Add JaCoCo (MANDATORY, high threshold)
- [ ] Add `jacoco-maven-plugin` to `test/load/pom.xml`
- [ ] Configure: `prepare-agent` goal + `report` goal + `check` goal with `≥95%` line coverage threshold
- [ ] Coverage threshold is MANDATORY — build MUST fail below 95% (user: "absolutely mandatory, with high threshold like Go coverage thresholds")
- [ ] Run `cd test/load && mvn verify` — verify coverage report generated and threshold passes
- [ ] Add CI/CD workflow step to upload JaCoCo coverage report as artifact



---

## Phase 9: lint-deployments Error Message Improvements

### Task 9.1: Audit all validator error messages
- [ ] Run `go run ./cmd/cicd lint-deployments validate-all` against real violations (create test cases)
- [ ] Collect ALL error message outputs from all 8 validators
- [ ] Grade each: Clarity (1–5), Actionability (1–5), Context (1–5)

### Task 9.2: Standardize error message format
- [ ] Design target format: `[VALIDATOR] path: description | Expected: X | Found: Y | Fix: Z | See: Arch Section N.N`
- [ ] Implement format in all validators that score below threshold

### Task 9.3: Add "See ARCHITECTURE.md Section" references to errors
- [ ] Each validator error → reference the ARCHITECTURE.md section explaining the rule
- [ ] ValidateAdmin → Section 5.3 Dual HTTPS; ValidateSecrets → Section 12.6; ValidatePorts → Section 3.4; etc.

### Task 9.4: Tests for error message quality
- [ ] Update validator tests to assert specific error message content (not just error presence)
- [ ] Verify test coverage ≥98% maintained

---

## Phase 10: skeleton-template Improvements

### Task 10.1: Add template comment headers to skeleton source files
- [ ] Add `// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use` to key skeleton source files
- [ ] Target files: `internal/apps/skeleton/template/*.go` and `cmd/skeleton-template/main.go`
- [ ] Verify comments are clear and discoverable

### Task 10.2: Add CICD placeholder detection lint rule
- [ ] Add new `validate-skeleton` validator in `internal/apps/cicd/lint_skeleton/` (following existing validator pattern)
- [ ] Validator scans all non-skeleton directories for unreplaced `skeleton`/`Skeleton`/`SKELETON` strings in `.go` source files
- [ ] Register command as `cicd lint-skeleton` in the CICD command dispatcher
- [ ] Add pre-commit hook for `cicd lint-skeleton`
- [ ] Add tests for the validator (≥98% coverage)
- [ ] Verify runs cleanly (`go run ./cmd/cicd lint-skeleton`)

### Task 10.3: Add example domain pattern to skeleton-template (commented out)
- [ ] Add example: entity model (GORM), repository, service, handler — as commented reference
- [ ] Shows correct patterns without running code

---

## Phase 11: Validation

### Task 11.1: Build
- [ ] `go build ./...` — clean
- [ ] `go build -tags e2e,integration ./...` — clean

### Task 11.2: Lint
- [ ] `golangci-lint run --fix && golangci-lint run` — 0 issues
- [ ] `golangci-lint run --build-tags e2e,integration --fix` — 0 issues

### Task 11.3: Tests
- [ ] `go test ./... -shuffle=on -count=1` — exit 0

### Task 11.4: Documentation validation
- [ ] `go run ./cmd/cicd lint-docs` — all propagation verified
- [ ] `go run ./cmd/cicd lint-text` — UTF-8 clean
- [ ] `go run ./cmd/cicd lint-deployments validate-all` — all validators pass

### Task 11.5: Python validation (after Phase 7)
- [ ] `uvx ruff check .` — 0 issues
- [ ] `uvx ruff format --check .` — 0 issues
- [ ] `uvx mypy .` — 0 type errors

### Task 11.6: Review passes (1 of minimum 3)

**Pass 1** — ALL 8 quality attributes:
- [ ] Correctness, Completeness, Thoroughness, Reliability, Efficiency, Accuracy, NO Time Pressure, NO Premature Completion

**Pass 2** — ALL 8 quality attributes (repeat)
- [ ] (Same 8 attributes; fresh perspective)

**Pass 3** — ALL 8 quality attributes
- [ ] (Same 8)
- [ ] ANY issue found? → Continue to Pass 4

**Pass 4** (if ANY issue in Pass 3):
- [ ] (Same 8)

**Pass 5** (if Pass 4 still has issues):
- [ ] (Same 8) — Diminishing returns → complete

### Task 11.7: Git commit
- [ ] `git add -A`
- [ ] `git commit -m "feat: v9 improvements (review passes, arch optimization, skills, toolchain)"`
- [ ] `git push`
