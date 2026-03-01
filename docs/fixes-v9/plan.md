# Fixes v9 - Quality Review Passes, Agent Semantics, ARCHITECTURE.md Optimization, Skills, Toolchain Modernization

**Status**: Planning → Ready for Execution
**Created**: 2026-03-01
**Last Updated**: 2026-03-01 (quizme-v1 + quizme-v2 + quizme-v3 + quizme-v4 answers merged)

## Quality Mandate - MANDATORY

- ✅ **Correctness**: ALL changes must be accurate and semantically correct
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: lint-docs, build, lint, tests must all pass
- ✅ **Efficiency**: Optimized for clarity and maintainability, NOT speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation
- ❌ **Premature Completion**: NEVER mark tasks complete without objective evidence

**ALL issues are blockers. Resources (time, tokens) are NEVER a constraint.**

---

## Executive Summary

Eleven phases of improvement, all to be implemented:

1. **Phase 1: Quality Review Passes** — Rewrite so each pass checks ALL 8 quality attributes; min 3, max 5; continue to pass 4 whenever pass 3 finds ANY issue
2. **Phase 2: Agent Semantics** — beast-mode dual: generic principles + Go-specific Quality Gate examples (labeled); other agents confirmed domain-specific
3. **Phase 3: ARCHITECTURE.md Optimization** — Consolidate duplications, fill omissions (skills section, agent/skill/instruction matrix); target <4,000 lines
4. **Phase 4: doc-sync Propagation** — Add 12.7, 11.4, B.6 @source blocks to doc-sync.agent.md
5. **Phase 5: Copilot Skills** — Create 13 skills + infrastructure (12 from quizme-v2 + `new-service` from quizme-v3 S4-Item4)
6. **Phase 6: Pre-commit/Pre-push Linters** — Add 8 approved hooks: ruff-check, ruff-format, checkov, sqlfluff, taplo, pyproject-fmt, validate-pyproject, editorconfig-checker
7. **Phase 7: Python Toolchain Modernization** — All 5 items confirmed YES: remove black+isort+flake8, add ruff, migrate to uvx
8. **Phase 8: Java Toolchain Additions** — Add 6 tools to load tests: Spotless, Checkstyle, Error Prone, NullAway, maven-enforcer, JaCoCo (mandatory + ≥95% coverage threshold)
9. **Phase 9: lint-deployments Error Messages** — Analyze and improve all validator error messages for clarity and actionability (Q4 from quizme-v2)
10. **Phase 10: skeleton-template Improvements** — Research-driven improvements to naming, content, placeholder detection, auto-discovery (Q6 from quizme-v2)
11. **Phase 11: Validation** — lint-docs, build, lint, tests; 3–5 review passes

---

## Phase 1: Quality Review Passes Rework

### Current State
Review passes in ARCHITECTURE.md Section 2.5 and all @source targets check ONE attribute per pass (Pass 1=Completeness, Pass 2=Correctness, Pass 3=Quality).

### Target State
Each review pass checks ALL 8 quality attributes. Min 3, max 5. Continue to pass 4 whenever pass 3 finds ANY issue.

**8 Attributes (per pass)**:
1. ✅ Correctness — code/docs correct, no regressions
2. ✅ Completeness — all tasks/steps/items addressed
3. ✅ Thoroughness — evidence-based, all edge cases
4. ✅ Reliability — build, lint, test, coverage all pass
5. ✅ Efficiency — optimized for maintainability, not speed
6. ✅ Accuracy — root cause addressed, not symptoms
7. ❌ NO Time Pressure — not rushed
8. ❌ NO Premature Completion — evidence required before marking complete

**Continuation rule**: Continue to pass 4 when pass 3 finds ANY issue. Pass 5 if pass 4 still finds issues. Diminishing returns = done.
**Scope**: ALL work types — code, docs, config, tests, infrastructure, deployments.

### Files to Update
1. `docs/ARCHITECTURE.md` Section 2.5 — canonical
2. `.github/instructions/01-02.beast-mode.instructions.md` — @source
3. `.github/instructions/06-01.evidence-based.instructions.md` — @source
4. `.github/agents/beast-mode.agent.md`
5. `.github/agents/doc-sync.agent.md`
6. `.github/agents/fix-workflows.agent.md`
7. `.github/agents/implementation-execution.agent.md`
8. `.github/agents/implementation-planning.agent.md`

---

## Phase 2: Agent Semantic Analysis

### Decisions (from quizme-v1 Q5+Q6)

**beast-mode**: Dual structure:
- Generic section: continuous execution principles, applicable to ANY work type
- Labeled section: "Quality Gate Commands (Go Projects)" — go build, golangci-lint, go test, golangci-lint --build-tags e2e,integration
- Review passes section → updated via Phase 1

**Other agents**: All confirmed correctly domain-specific — NO scope changes.

---

## Phase 3: ARCHITECTURE.md Optimization

**Target**: <4,000 lines (from 4,445). Do NOT sacrifice clarity, completeness, correctness.

### 3.1 Quality Attributes → Consolidate to Section 11.1, @propagate to 1.3 + 2.5
### 3.2 CLI Patterns → Canonical in **Section 4.4.7**, Section 9.1 cross-references
### 3.3 Port Assignments → Canonical in **Section 3.4**, **DELETE Appendix B.1 + B.2**, resequence remaining B.# numbering, update all cross-references
### 3.4 Infrastructure Blocker Escalation → Keep in BOTH 13.7 and 2.5 (no change)
### 3.5 Skills Section → Add new section covering: VS Code Skills overview, `.github/skills/` organization, naming convention (`SKILLNAME.md`), skill catalogue table (name, purpose, link), reference to VS Code docs
### 3.6 Agent/Skill/Instruction Matrix → Add concise 4-row decision matrix to Section 2.1 Agent Orchestration; cross-reference to skills section
### 3.7 Review Pass Count → Sweep for stale "3 passes" mentions, update to "minimum 3, maximum 5"

**Expected net savings**: ~140–175 lines from deletions, ~50–65 lines added for skills/matrix.

---

## Phase 4: doc-sync Agent Propagation

### Missing @source blocks to add to doc-sync.agent.md
1. **Section 12.7** Documentation Propagation Strategy — core to doc-sync purpose
2. **Section 11.4** Documentation Standards — doc quality requirements
3. **Appendix B.6** Instruction File Reference — note: verify/update B.# after Phase 3.3 resequencing

---

## Phase 5: Copilot Skills

### Infrastructure (Q0 → A: Create early, before first skill)
- Create `.github/skills/` directory
- Create `README.md` in `.github/skills/` describing skill conventions
- Create `SKILL-TEMPLATE.md` as reference template

### Skills to Create (all YES from quizme-v2)

All skills use flat naming: `SKILLNAME.md` in `.github/skills/`.

#### Group A: Test Generation (Q1–Q3: all YES)
| Skill | Purpose |
|-------|---------|
| `test-table-driven` | Generate table-driven Go tests (t.Parallel, UUIDv7 data, require over assert, subtests) |
| `test-fuzz-gen` | Generate `_fuzz_test.go` (15s fuzz time, corpus examples, build tags) |
| `test-benchmark-gen` | Generate `_bench_test.go` (mandatory for crypto, reset timer pattern) |

#### Group B: Infrastructure / Deployment (Q5: YES; Q4: NO skill, Q6: research → Phase 10)
| Skill | Purpose |
|-------|---------|
| `migration-create` | Create numbered golang-migrate SQL files (correct range: template 1001-1999, domain 2001+, paired up/down) |

Note: `compose-validator` skill deferred (Q4 = B: fix lint-deployments error messages first — Phase 9). `service-scaffold` deferred to Phase 10 planning.

#### Group C: Code Quality (Q7–Q8: both YES)
| Skill | Purpose |
|-------|---------|
| `coverage-analysis` | Analyze coverprofile output, categorize uncovered lines, generate targeted test suggestions |
| `fips-audit` | Detect FIPS 140-3 violations + provide fix guidance (goes beyond cicd lint-go non-fips-algorithms) |

#### Group D: Documentation (Q9–Q10: both YES)
| Skill | Purpose |
|-------|---------|
| `propagation-check` | Detect @propagate/@source drift + generate corrected @source block content |
| `openapi-codegen` | Generate three oapi-codegen configs (server/model/client) + OpenAPI spec skeleton for any service |

#### Group E: Scaffolding (Q11–Q12: both YES + skill-scaffold: missing from v2, user identified)
| Skill | Purpose |
|-------|---------|
| `agent-scaffold` | Create conformant `.github/agents/NAME.agent.md` with all mandatory sections |
| `instruction-scaffold` | Create conformant `.github/instructions/NN-NN.name.instructions.md` |
| `skill-scaffold` | Create conformant `.github/skills/NAME.md` (3rd Copilot customization type — was missing from quizme-v2) |
| `new-service` | Guide service creation from skeleton-template: copy, rename, register, migrate, test (quizme-v3 S4-Item4: create skill, not agent) |

### Three Copilot Customization Types
VS Code Copilot has exactly 3 customization file types:
1. **Instructions** — `.github/copilot-instructions.md` + `.github/instructions/*.instructions.md` (always loaded, passive context)
2. **Agents** — `.github/agents/*.agent.md` (on-demand, `/agent-name` invocation, tools + handoffs)
3. **Skills** — `.github/skills/*.md` (on-demand, `#skill-name` reference in chat)

All 3 scaffolding skills cover all 3 types.

### After Skills Created (Q15 → YES)
Update relevant agents' `skills:` frontmatter to reference applicable skills.

### ARCHITECTURE.md Skills Catalogue (Q14 → YES, concise)
Skills section added in Phase 3.5 will include catalogue table.

---

## Phase 6: Pre-commit / Pre-push Linter Additions

**Status**: Decisions confirmed from quizme-v3. Ready for implementation.

### Approved Hooks to Add

| Hook | Tool | Action |
|------|------|--------|
| ruff check | astral-sh/ruff-pre-commit | Python linting (replacing flake8+isort+pyupgrade) |
| ruff format | astral-sh/ruff-pre-commit | Python formatting (replacing black) |
| checkov | bridgecrewio/checkov | Deep Docker/compose security scanning |
| sqlfluff | sqlfluff-pre-commit | SQL migration file linting, dialect=postgres |
| taplo | CommaNet/taplo-pre-commit | TOML file formatting |
| pyproject-fmt | tox-dev/pyproject-fmt | Auto-format `pyproject.toml` |
| validate-pyproject | abravalheri/validate-pyproject | Schema validation for `pyproject.toml` |
| editorconfig-checker | editorconfig-checker | All-file editorconfig compliance (requires `.editorconfig` first) |

### Current Hooks Inventory (already in place)
**Go**: golangci-lint (full + incremental), go build, cicd custom lint (lint-docs, lint-go, lint-text, lint-workflow, lint-deployments, lint-ports, etc.), go mod tidy
**Python**: bandit (security)
**Other**: gitleaks (secrets), yamllint, actionlint (GitHub Actions YAML), hadolint (Dockerfile), shellcheck, markdownlint-cli2, commitizen (conventional commits), pre-commit-hooks (yaml, json, toml, xml, end-of-file, trailing-whitespace, merge-conflict, etc.)

---

## Phase 7: Python Toolchain Modernization

### Current State
`pyproject.toml` installs: `black`, `isort`, `flake8`, `mypy`, `bandit`, `pytest` + plugins

### Ruff (replaces multiple tools)
**Ruff** is a Python linter AND formatter written in Rust. 10–100x faster than black/flake8.

**Ruff replaces** (rules enabled per project needs):
| Replaced Tool | Ruff Rule Set | Equivalence |
|---------------|--------------|-------------|
| `flake8` | E, W (pycodestyle), F (pyflakes) | Full replacement |
| `isort` | I (isort) | Full replacement |
| `black` | `ruff format` | Style-compatible replacement |
| `pyupgrade` | UP (pyupgrade) | Full replacement |
| `flake8-bugbear` | B (bugbear) | Full replacement |
| `flake8-simplify` | SIM | Full replacement |
| `pydocstyle` | D (pydocstyle) | Full replacement |
| `mccabe` | C90 (complexity) | Full replacement |
| `autoflake` | F401 (unused imports) | Full replacement |

**Ruff does NOT replace**:
- `mypy` — static type checking (ruff has some UP rules but not full type analysis)
- `bandit` — security linting (ruff has partial S rules, bandit is more comprehensive for security)
- `pytest` — test runner

### uvx (replaces pip-installed CLI tools)
**uvx** runs Python CLI tools in isolated ephemeral environments without `pip install`.

**Replaces**:
```bash
pip install ruff && ruff check .    →   uvx ruff check .
pip install mypy && mypy .          →   uvx mypy .
pip install black && black .        →   uvx black .      (after ruff migration: uvx ruff format .)
```

**Benefits**: No global pip pollution, no venv management, always uses pinned version, deterministic.

### Migration Plan
1. **Remove from pyproject.toml**: `black`, `isort`, `flake8`
2. **Add to pyproject.toml**: `ruff>=0.9.0`
3. **Add `[tool.ruff]` config** in pyproject.toml: line-length=200, target-version=py314, enable E/W/F/I/B/UP/SIM/C90 rule sets
4. **Add `[tool.ruff.format]`** config: equivalent to black settings
5. **Update pre-commit hooks**: Replace black/isort/flake8 hooks with `astral-sh/ruff-pre-commit` hooks
6. **Update CI/CD workflows**: Replace `black --check` + `isort --check` + `flake8` with `ruff check` + `ruff format --check`
7. **Switch pyproject.toml tool installs to uvx**: Update scripts/commands to use `uvx ruff`, `uvx mypy`

---

## Phase 8: Java Toolchain Additions (Gatling Load Tests)

**Scope**: `test/load/pom.xml` — Gatling Java simulation tests.

### Current State
Already has: spotbugs + findsecbugs, owasp-dependency-check, versions-maven-plugin, maven-compiler-plugin

### Decisions (from quizme-v3)

| Tool | Action |
|------|--------|
| Spotless + google-java-format | Add `spotless-maven-plugin` with google-java-format, phase=validate |
| Checkstyle | Add `maven-checkstyle-plugin` with Google rules, phase=validate |
| Error Prone | Add to `maven-compiler-plugin` config as annotation processor |
| NullAway | Add as Error Prone plugin (requires Error Prone) |
| maven-enforcer | Add with `dependencyConvergence` + `requireJavaVersion` + `requireMavenVersion` rules |
| JaCoCo | Add with high coverage thresholds matching Go standards (≥95%); absolutely mandatory, failing build on threshold breach |

**JaCoCo note**: User specified "absolutely mandatory, with high threshold like Go coverage thresholds" — configure with ≥95% line coverage minimum, failing build on threshold breach.

### Reference Table (Java ↔ Go equivalents)

**Note**: "Ruff for Java" does not exist — ruff is a Python-only tool. These are Java-specific equivalents:

| Tool | Purpose | Equivalent to |
|------|---------|---------------|
| **google-java-format** (via Spotless) | Java code formatter — consistent style | ruff format (Python) |
| **Checkstyle** | Java style enforcement (Google/Sun style guide) | golangci-lint stylecheck (Go) |
| **Error Prone** | Compile-time bug detection via annotation processor | go vet (Go) |
| **NullAway** | Null safety analysis, works with Error Prone | N/A (Go has nil checks) |
| **maven-enforcer** | Enforce Maven version constraints, dependency convergence | go.mod (Go modules) |
| **JaCoCo** | Java code coverage | go test -coverprofile (Go) |

---

## Phase 9: lint-deployments Error Message Improvements

**Decision from quizme-v2 Q4 (answer B)**: NO to compose-validator skill. Instead, make the underlying `cicd lint-deployments` validators self-explanatory.

### Analysis Scope
Review all 8 validators' error messages for:
- **Clarity**: Does the error message identify the exact file + line?
- **Actionability**: Does it tell the user WHAT to fix and HOW?
- **Context**: Does it explain WHY the rule exists?
- **Consistency**: Same format across all validators?

### 8 Validators to Review
1. `ValidateNaming` — kebab-case directory/file naming
2. `ValidateKebabCase` — YAML keys and compose service names
3. `ValidateSchema` — service template config schema
4. `ValidateTemplatePattern` — template naming + placeholders
5. `ValidatePorts` — PORT range enforcement
6. `ValidateTelemetry` — OTLP endpoint consistency
7. `ValidateAdmin` — admin 127.0.0.1:9090 bind policy
8. `ValidateSecrets` — inline secret detection + Docker secrets

### Target Error Message Format
```
[VALIDATOR] path/to/file.yml: <concise problem description>
  Expected: <what it should be>
  Found:    <what it is>
  Fix:      <specific action to take>
  See:      ARCHITECTURE.md Section X.Y for rule rationale
```

---

## Phase 10: skeleton-template Improvements

**Decision from quizme-v2 Q6 (answer = modified D)**: Research skeleton-template and improve naming, content, placeholder detection, and automatic agent discovery.

### Research Findings

**Current state**: `internal/apps/skeleton/template/` + `cmd/skeleton-template/main.go` + deployment at `deployments/skeleton/` — functions as a running skeleton service.

**Best practices for service scaffolding templates:**

#### 10.1 Naming and Discoverability — quizme-v3 decisions
- ❌ SKIP: `SCAFFOLDING.md` in project root — user: "no more doc bloat and doc sprawl" (quizme-v3 S4-Item1 answer C/B)
- ✅ YES: Template comment headers in skeleton source files (quizme-v3 S4-Item2 answer A)
- ✅ YES: `cicd validate-skeleton` placeholder detection lint rule (quizme-v4 A — YES)
- ❌ SKIP agent, ADD SKILL: `new-service.agent.md` replaced by `new-service` skill in Phase 5 Group E (quizme-v3 S4-Item4 answer C: "make it a skill")

#### 10.2 Placeholder Detection — YES (quizme-v4 A)
- `cicd validate-skeleton` command: scan non-skeleton directories for unreplaced `skeleton`/`Skeleton`/`SKELETON` strings
- Add new validator in `cmd/cicd/` following same pattern as existing validators
- Add pre-commit hook for validate-skeleton

#### 10.3 Content Improvements (unchanged from research)
- Add: Example domain model, repository, service, handler in skeleton-template as comments
- Add: Step-by-step MIGRATION.md: "How to create a new service from this template"
- Add: Example OpenAPI paths/components for CRUD patterns (as commented-out reference)

#### 10.4 Implementation Plan (from decisions above)
**Phase 10 scope** (approved items including quizme-v4):
1. Add template comment headers to skeleton source files (quizme-v3 Item 2 YES)
2. Create `new-service` skill in `.github/skills/new-service.md` (moved to Phase 5 Group E)
3. Add commented example patterns (domain/repo/service/handler) — content improvement
4. Add MIGRATION.md step-by-step guide inside `internal/apps/skeleton/` (scope: skeleton dir only, not root)
5. Add `cicd validate-skeleton` placeholder detection lint rule + pre-commit hook (quizme-v4 Item 1 YES)

---

## Decisions Log (from quizme-v1.md + quizme-v2.md)

### From quizme-v1
| Q | Decision | Phase |
|---|----------|-------|
| Q1 | Each pass checks ALL 8 attributes | Phase 1 |
| Q2 | Min 3, max 5 passes | Phase 1 |
| Q3 | Continue on ANY issue in pass 3 | Phase 1 |
| Q4 | ALL work types | Phase 1 |
| Q5 | KEEP Go examples, label them | Phase 2 |
| Q6 | Dual: generic execution + Go QG examples | Phase 2 |
| Q7 | Other agents stay domain-specific | Phase 2 |
| Q8 | Quality attrs consolidate to 11.1, @propagate | Phase 3.1 |
| Q9 | CLI patterns canonical → **4.4.7** | Phase 3.2 |
| Q10 | Ports → **3.4** only, DELETE B.1+B.2, resequence | Phase 3.3 |
| Q11 | Infra blocker keep in BOTH 13.7 and 2.5 | Phase 3.4 |
| Q12 | Add new skills section to ARCHITECTURE.md | Phase 3.5 |
| Q13 | Add agent/skill/instruction matrix to 2.1 | Phase 3.6 |
| Q14 | doc-sync: add 12.7, 11.4, B.6 | Phase 4 |
| Q15 | Propagate content into doc-sync | Phase 4 |
| Q16 | YES create skills, all candidates in quizme-v2 | Phase 5 |
| Q22 | Target <4,000 lines | Phase 3 |
| Q23 | Embrace propagation automation | All |
| Q24 | Continue See X.Y cross-references | All |
| Q25 | Time/tokens NEVER a constraint | All |

### From quizme-v2
| Q | Decision | Phase |
|---|----------|-------|
| Q0 | A — Create .github/skills/ infra early | Phase 5 |
| Q1 | A — test-table-driven skill YES | Phase 5, Group A |
| Q2 | A — test-fuzz-gen skill YES | Phase 5, Group A |
| Q3 | A — test-benchmark-gen skill YES | Phase 5, Group A |
| Q4 | B — NO compose-validator skill; analyze/fix lint-deployments instead | Phase 9 |
| Q5 | A — migration-create skill YES | Phase 5, Group B |
| Q6 | D (modified) — research skeleton-template best practices | Phase 10 |
| Q7 | A — coverage-analysis skill YES | Phase 5, Group C |
| Q8 | A — fips-audit skill YES (with fix guidance) | Phase 5, Group C |
| Q9 | A — propagation-check skill YES | Phase 5, Group D |
| Q10 | A — openapi-codegen skill YES (all services) | Phase 5, Group D |
| Q11 | A — agent-scaffold skill YES | Phase 5, Group E |
| Q12 | A — instruction-scaffold skill YES | Phase 5, Group E |
| Q13 | A — flat SKILLNAME.md naming | Phase 5 |
| Q14 | A — YES catalogue table in ARCHITECTURE.md | Phase 3.5 |
| Q15 | A — YES update agents with skills: refs | Phase 5 |
| +   | skill-scaffold skill (user identified missing 3rd type) | Phase 5, Group E |

### From quizme-v3
| Q | Decision | Phase |
|---|----------|-------|
| S0-Q0 | A — skill-scaffold skill YES | Phase 5, Group E |
| S1-C1 | B — govulncheck SKIP | Phase 6 |
| S1-C2 | A — ruff check YES | Phase 6 + 7 |
| S1-C3 | A — ruff format YES | Phase 6 + 7 |
| S1-C4 | A — checkov YES | Phase 6 |
| S1-C5 | C — trivy DEFER | quizme-v4 |
| S1-C6 | C — semgrep DEFER | quizme-v4 |
| S1-C7 | A — sqlfluff YES | Phase 6 |
| S1-C8 | B — vale SKIP | Phase 6 |
| S1-C9 | C — codespell DEFER | quizme-v4 |
| S1-C10 | A — taplo YES | Phase 6 |
| S1-C11 | A — pyproject-fmt YES | Phase 6 |
| S1-C12 | A — validate-pyproject YES | Phase 6 |
| S1-C13 | C — editorconfig-checker DEFER | quizme-v4 |
| S2-Item1 | A — remove black YES | Phase 7 |
| S2-Item2 | A — remove isort YES | Phase 7 |
| S2-Item3 | A — remove flake8 YES | Phase 7 |
| S2-Item4 | A — add ruff full config YES | Phase 7 |
| S2-Item5 | A — migrate to uvx YES | Phase 7 |
| S3-Tool1 | A — Spotless+google-java-format YES | Phase 8 |
| S3-Tool2 | A — Checkstyle YES | Phase 8 |
| S3-Tool3 | B — PMD SKIP (SpotBugs preferred) | Phase 8 |
| S3-Tool4 | A — Error Prone YES | Phase 8 |
| S3-Tool5 | A — NullAway YES | Phase 8 |
| S3-Tool6 | A — maven-enforcer YES | Phase 8 |
| S3-Tool7 | A — JaCoCo YES (mandatory + high threshold ≥95%) | Phase 8 |
| S3-Tool8 | C — ArchUnit DEFER | quizme-v4 |
| S4-Item1 | B — SCAFFOLDING.md SKIP (no doc bloat) | Phase 10 |
| S4-Item2 | A — template comment headers YES | Phase 10 |
| S4-Item3 | blank — CICD placeholder detection DEFER | quizme-v4 |
| S4-Item4 | C — new-service.agent.md SKIP; create skill instead | Phase 5 Group E |

### From quizme-v4
| Q | Decision | Phase |
|---|----------|-------|
| S1-C1 | C — trivy DEFER again | quizme-v5 |
| S1-C2 | C — semgrep DEFER again | quizme-v5 |
| S1-C3 | B — codespell SKIP | Phase 6 |
| S1-C4 | A — editorconfig-checker YES (create .editorconfig first) | Phase 6 |
| S2-Tool1 | B — ArchUnit SKIP (load test arch is simple; overkill) | Phase 8 |
| S3-Item1 | A — CICD placeholder detection YES | Phase 10 |

### Phases Deferred to quizme-v5
- Phase 6 deferred: trivy, semgrep
