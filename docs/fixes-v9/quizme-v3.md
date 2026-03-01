# Quizme v3 — Deferred Decisions

**Purpose**: Numbered candidate lists for gated phases. Review each item and mark your answer.
**Format**: Each item has A (YES/include), B (NO/skip), C (DEFER) options + open **Answer:** line.

---

## Section 0 — skill-scaffold (Missing 3rd Copilot Customization Type)

The three VS Code Copilot customization file types are:
1. **Instructions** (`.github/instructions/*.instructions.md`) — always-on project context
2. **Agents** (`.github/agents/*.agent.md`) — slash-command autonomous workflows
3. **Skills** (`.github/skills/*.md`) — reusable capability modules referenced by agents

Phase 5 includes `agent-scaffold` and `instruction-scaffold` but **skills themselves** are also a customization type. A `skill-scaffold` skill would create new `SKILLNAME.md` files in `.github/skills/` following the proper template.

### Q0: Create skill-scaffold skill?

**A)** YES — Create `.github/skills/skill-scaffold.md` that generates new skill files with correct frontmatter, trigger phrases, content structure, and self-reference to the skills catalog. Covers all 3 Copilot customization types.
**B)** NO — The SKILL-TEMPLATE.md reference file is sufficient; no need for a skill that creates skills.
**C)**

**Answer:**

---

## Section 1 — Pre-commit / Pre-push Linter Candidates

Current hooks: gitleaks, yamllint, actionlint, hadolint, shellcheck, bandit, markdownlint-cli2, commitizen, and many custom Go local hooks (lint-go, format-go, lint-go-test, format-go-test, lint-go-mod, lint-golangci, lint-compose, lint-ports, lint-workflow, lint-deployments, lint-docs, validate-utf8, enforce-yaml).

Each candidate below: **A** = Add to pre-commit, **B** = Skip/not needed, **C** = Defer.

### Candidate 1: govulncheck (Go CVE Scanning)

Scans Go source code against the Go vulnerability database. Finds known CVEs in imported dependencies.
- **Trigger**: `repo: local`, runs `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- **Cost**: Fast (~2–5s for most codebases)
- **Overlap**: Complements OWASP check; govulncheck is Go-specific and more precise than generic SCA tools
- **Why add**: Catches transitive CVEs early, before CI/CD

**A)** YES — Add `govulncheck ./...` as pre-commit hook
**B)** NO — Already covered by other scanning; prefer CI-only
**C)**

**Answer:**

### Candidate 2: ruff check (Python linting — part of Phase 7)

Replaces flake8 + isort + pyupgrade + bugbear + many others. Phase 7 migrates to ruff.
- **Trigger**: `astral-sh/ruff-pre-commit`, hook id: `ruff`, `args: [--fix]`
- **Cost**: Extremely fast (~10–50ms for typical Python files)
- **Replaces**: flake8 hook
- **Why add**: Part of Python modernization; required if removing flake8 pre-commit hook

**A)** YES — Add ruff check pre-commit hook (and remove flake8 hook)
**B)** NO — Keep flake8; do not migrate to ruff for linting
**C)**

**Answer:**

### Candidate 3: ruff format (Python formatting — part of Phase 7)

Replaces black formatter. Black-compatible, ~10–50x faster.
- **Trigger**: `astral-sh/ruff-pre-commit`, hook id: `ruff-format`
- **Cost**: Extremely fast
- **Replaces**: black hook
- **Why add**: Part of Python modernization; required if removing black pre-commit hook

**A)** YES — Add ruff format pre-commit hook (and remove black hook)
**B)** NO — Keep black; do not migrate to ruff format
**C)**

**Answer:**

### Candidate 4: checkov (Dockerfile / Compose / IaC Security)

Static analysis for Dockerfiles, docker-compose files, and IaC. Checks for security misconfigurations.
- **Trigger**: `bridgecrewio/checkov`, runs against `deployments/` and `Dockerfile`
- **Cost**: Moderate (~5–15s)
- **Overlap**: Hadolint covers Dockerfile best practices; checkov adds security layer
- **Why add**: Catches secrets in env vars, privileged containers, missing healthchecks, etc.

**A)** YES — Add checkov for Dockerfile + Compose security scanning
**B)** NO — Hadolint + existing validators sufficient
**C)**

**Answer:**

### Candidate 5: trivy (Container / Dependency Vulnerability)

Scans container images and Go/Python/Java dependencies for CVEs.
- **Trigger**: `aquasecurity/trivy-action` or `repo: local`, runs `trivy fs --exit-code 1 --severity HIGH,CRITICAL .`
- **Cost**: Moderate (~10–30s); downloads vuln DB on first run
- **Overlap**: Complements govulncheck; broader (supports all languages, OS packages in containers)
- **Why add**: catches OS-level and multi-language CVEs in one tool

**A)** YES — Add trivy filesystem scan pre-commit hook
**B)** NO — govulncheck + OWASP sufficient; trivy too slow for pre-commit
**C)**

**Answer:**

### Candidate 6: semgrep (Multi-language SAST)

Pattern-based static analysis. Checks Go, Python, Java, and Dockerfiles with curated rule sets.
- **Trigger**: `returntocorp/semgrep-action` or `repo: local`
- **Cost**: Slow (~30–120s for large codebases). Better as CI-only tool.
- **Overlap**: Complements golangci-lint (catch patterns linters miss)
- **Why add**: Detects hardcoded secrets, insecure patterns, OWASP Top 10 issues

**A)** YES — Add semgrep as pre-commit hook
**B)** NO — Too slow for pre-commit; run in CI/CD only (ci-sast workflow)
**C)**

**Answer:**

### Candidate 7: sqlfluff (SQL Migration Linting)

Lints SQL files in `internal/**/migrations/*.sql`. Enforces style consistency.
- **Trigger**: `repo: local`, runs `sqlfluff lint --dialect=postgres ...`
- **Cost**: Moderate (~1–3s per SQL file)
- **Config needed**: `.sqlfluff` file with `dialect = postgres`, style rules
- **Why add**: SQL migrations are rarely linted; catches syntax errors, style inconsistencies

**A)** YES — Add sqlfluff for SQL migration files
**B)** NO — SQL migrations are simple enough; not worth adding tooling complexity
**C)**

**Answer:**

### Candidate 8: vale (Prose Linting for .md Documentation)

Lints Markdown documentation for prose style, terminology consistency, and writing quality.
- **Trigger**: `repo: local`, runs `vale docs/`
- **Cost**: Fast (~1–3s)
- **Config needed**: `.vale.ini` + styles directory (Vale style packages)
- **Special consideration**: Our terminology rules (authn vs auth, MUST/SHOULD/MAY) could be encoded as Vale rules

**A)** YES — Add vale prose linting for docs/
**B)** NO — markdownlint-cli2 is sufficient for .md files; prose style not needed
**C)**

**Answer:**

### Candidate 9: codespell (Typo Detection)

Finds common typos in all text files (Go source, Markdown, YAML, etc.).
- **Trigger**: `codespell-project/codespell`, `args: [--skip=".git,*.sum,*.lock", --ignore-words-list="teh,sot,ot"]`
- **Cost**: Fast (~1–2s)
- **Low false-positive rate**: codespell only flags common known typos, not unusual words
- **Why add**: Finds typos that slip through code review

**A)** YES — Add codespell typo detection
**B)** NO — Typos not a significant problem in this project
**C)**

**Answer:**

### Candidate 10: taplo (TOML Formatting)

Formats and validates TOML files (`.taplo.toml`, if any TOML files exist).
- **Trigger**: `CommaNet/taplo-pre-commit`, hook id: `taplo-format`
- **Cost**: Very fast (~<1s)
- **Applicability**: This project uses YAML primarily; TOML files minimal (if any)
- **Why add**: Keeps TOML files consistently formatted; may not apply yet

**A)** YES — Add taplo TOML formatter
**B)** NO — No significant TOML files in this project; skip
**C)**

**Answer:**

### Candidate 11: pyproject-fmt (pyproject.toml Formatter)

Normalizes and formats `pyproject.toml` — sorts sections, normalizes formatting.
- **Trigger**: `tox-dev/pyproject-fmt`, hook id: `pyproject-fmt`
- **Cost**: Very fast (~<1s)
- **Why add**: Keeps pyproject.toml consistently formatted; particularly useful after ruff migration

**A)** YES — Add pyproject-fmt for pyproject.toml
**B)** NO — Manual formatting of pyproject.toml is sufficient
**C)**

**Answer:**

### Candidate 12: validate-pyproject (pyproject.toml Schema Validation)

Validates `pyproject.toml` against the official schema — catches missing required fields, unknown keys.
- **Trigger**: `abravalheri/validate-pyproject`, hook id: `validate-pyproject`
- **Cost**: Very fast (~<1s)
- **Why add**: Prevents broken pyproject.toml from reaching CI

**A)** YES — Add validate-pyproject schema validation
**B)** NO — pyproject.toml rarely changes; schema validation not needed
**C)**

**Answer:**

### Candidate 13: editorconfig-checker (EditorConfig Compliance)

Verifies all files comply with `.editorconfig` settings (indentation, line endings, charset, etc.).
- **Trigger**: `editorconfig-checker/editorconfig-checker`, hook id: `editorconfig-checker`
- **Cost**: Fast (~1–3s)
- **Prerequisite**: `.editorconfig` file must exist in root
- **Why add**: Enforces consistent editor settings across all file types and contributors

**A)** YES — Add editorconfig-checker
**B)** NO — Editor settings enforced via golangci-lint and gofumpt for Go; sufficient
**C)**

**Answer:**

---

## Section 2 — Python ruff + uvx Migration Items

Each item is a specific migration step. **A** = Do it, **B** = Skip/keep current, **C** = Defer.

### Item 1: Remove black from pyproject.toml

Remove `black` from dev dependencies and `[tool.black]` config section.
Ruff format provides identical functionality (Black-compatible output).

**A)** YES — Remove black, replace with ruff format
**B)** NO — Keep black alongside ruff (dual-tool approach)
**C)**

**Answer:**

### Item 2: Remove isort from pyproject.toml

Remove `isort` from dev dependencies and `[tool.isort]` config section.
Ruff's `I` rule category replaces isort completely with identical output.

**A)** YES — Remove isort, replace with ruff --select I
**B)** NO — Keep isort alongside ruff (dual-tool approach)
**C)**

**Answer:**

### Item 3: Remove flake8 from pyproject.toml

Remove `flake8` from dev dependencies and `[tool.flake8]` config section.
Ruff E/W/F rules replace flake8 + most common plugins.

**A)** YES — Remove flake8, replace with ruff E/W/F rules
**B)** NO — Keep flake8 alongside ruff (dual-tool approach)
**C)**

**Answer:**

### Item 4: Add ruff to pyproject.toml with full config

Add ruff with `[tool.ruff]` and `[tool.ruff.format]` sections:
- `line-length = 200`, `target-version = "py314"`
- `select = ["E", "W", "F", "I", "B", "UP", "SIM", "C90"]`
- `format.quote-style = "double"` (black-compatible)

**A)** YES — Add ruff with full configuration
**B)** NO — Use ruff with only default config (minimal)
**C)**

**Answer:**

### Item 5: Migrate to uvx for running Python CLI tools in scripts/CI

Replace patterns like `pip install X && X` or `python -m X` with `uvx X`.
Affects: ruff invocations, mypy, bandit, pytest in docs/scripts and CI workflows.

**A)** YES — Migrate all scripted Python tool invocations to uvx
**B)** NO — Keep current pip-based invocations
**C)**

**Answer:**

---

## Section 3 — Java Toolchain Additions (test/load/)

Each item is a Maven plugin addition to `test/load/pom.xml`. **A** = Add, **B** = Skip, **C** = Defer.

### Tool 1: Spotless + google-java-format (Java Code Formatting)

Auto-formats Java files to Google Java Format (2-space indent, sorted imports).
- **Plugin**: `com.diffplug.spotless:spotless-maven-plugin`
- **Phase**: `validate` (formats on `mvn validate`, fails if unformatted)
- **Equivalent to**: `ruff format` / `gofumpt` for Java
- **Why**: Enforces consistent formatting; eliminates style review comments

**A)** YES — Add Spotless + google-java-format
**B)** NO — Gatling load test code formatting not a priority
**C)**

**Answer:**

### Tool 2: Checkstyle (Java Style Enforcement)

Enforces Java coding standards — naming, Javadoc, import order, line length.
- **Plugin**: `org.apache.maven.plugins:maven-checkstyle-plugin`
- **Config**: Google Checkstyle rules (or Sun style)
- **Phase**: `validate`
- **Equivalent to**: `golangci-lint stylecheck` for Java

**A)** YES — Add Checkstyle with Google rules
**B)** NO — Spotless formatting is sufficient; style enforcement too strict for load tests
**C)**

**Answer:**

### Tool 3: PMD (Java Static Analysis)

Finds code smells, dead code, unused variables, empty catch blocks, overly complex methods.
- **Plugin**: `org.apache.maven.plugins:maven-pmd-plugin`
- **Equivalent to**: `golangci-lint staticcheck` for Java
- **Why**: Catches real bugs: empty catch blocks, unused imports, NPE-prone patterns

**A)** YES — Add PMD static analysis
**B)** NO — SpotBugs already covers enough; PMD overlap too high
**C)**

**Answer:**

### Tool 4: Error Prone (Java Compile-Time Bug Detection)

Google's compiler plugin that catches known bug patterns at compile time.
- **Plugin**: Error Prone via `maven-compiler-plugin` configuration
- **Equivalent to**: `go vet` for Java (compiler-integrated)
- **Why**: Catches real bugs that pass compilation: null dereference, API misuse, thread safety

**A)** YES — Add Error Prone to maven-compiler-plugin
**B)** NO — SpotBugs already provides runtime analysis; compiler-time check not needed
**C)**

**Answer:**

### Tool 5: NullAway (Null Safety — requires Error Prone)

Null-safety checker built on Error Prone. Enforces @NonNull / @Nullable annotations.
- **Plugin**: NullAway annotation processor via Error Prone
- **Prerequisite**: Tool 4 (Error Prone) must be approved
- **Why**: Eliminates NPEs at compile time — most common Java production bug

**A)** YES — Add NullAway (requires Error Prone approved above)
**B)** NO — NPE prevention via code review; annotation overhead not worth it
**C)**

**Answer:**

### Tool 6: maven-enforcer-plugin (Dependency + Build Rules)

Enforces build constraints: dependency convergence, minimum Maven/Java versions, banned dependencies.
- **Plugin**: `org.apache.maven.plugins:maven-enforcer-plugin`
- **Rules**: `dependencyConvergence`, `requireMavenVersion`, `requireJavaVersion`
- **Equivalent to**: Go module checks in `lint-go-mod`
- **Why**: Prevents silent version conflicts in transitive dependencies

**A)** YES — Add maven-enforcer with convergence + version rules
**B)** NO — Versions plugin already provides dependency management; enforcer redundant
**C)**

**Answer:**

### Tool 7: JaCoCo (Java Code Coverage)

Generates coverage reports for Java load tests.
- **Plugin**: `org.jacoco:jacoco-maven-plugin`
- **Phase**: `test` / `verify`
- **Why**: Provides coverage visibility for load test simulation code

**A)** YES — Add JaCoCo coverage reporting
**B)** NO — Load test coverage not meaningful; Gatling already shows scenario coverage
**C)**

**Answer:**

### Tool 8: ArchUnit (Architecture Rule Enforcement)

Tests Java architectural constraints: package dependencies, layer isolation, naming conventions.
- **Plugin**: ArchUnit via JUnit test
- **Why**: Enforces that Gatling load test code stays in proper layers; prevents test util leakage
- **Effort**: Highest of all options (requires writing ArchUnit test class)

**A)** YES — Add ArchUnit architecture tests
**B)** NO — Load test architecture is simple; ArchUnit enforcement overkill
**C)**

**Answer:**

---

## Section 4 — skeleton-template Improvement Items

Each item: **A** = Do it now (Phase 10), **B** = Skip, **C** = Defer to later.

### Item 1: SCAFFOLDING.md in project root

Step-by-step guide: copy skeleton-template → your-service-name, what to rename, what to register.
- **Effort**: ~1–2 hours
- **Value**: Onboarding time reduced from "figure it out" to following a guide

**A)** YES — Write SCAFFOLDING.md in root
**B)** NO — README.md links to architecture doc; SCAFFOLDING.md redundant
**C)**

**Answer:**

### Item 2: Add template comment headers to skeleton source files

Add `// TEMPLATE: Rename skeleton → your-service-name before use. See SCAFFOLDING.md.` to key files.
- **Effort**: ~30 minutes
- **Value**: Prevents copy-paste errors where "skeleton" strings remain unreplaced

**A)** YES — Add template comments to source files
**B)** NO — SCAFFOLDING.md and placeholder detection (Item 3) is sufficient
**C)**

**Answer:**

### Item 3: CICD placeholder detection lint rule

Add `cicd validate-skeleton` lint check: scan non-skeleton directories for unreplaced "skeleton" strings.
- **Effort**: ~2–4 hours (Go validator + pre-commit hook)
- **Value**: Catches copy-paste-forgot-to-rename errors automatically

**A)** YES — Add placeholder detection CI check
**B)** NO — Template comments (Item 2) + code review sufficient
**C)**

**Answer:**

### Item 4: new-service.agent.md — VS Code agent to guide service creation

Create `.github/agents/new-service.agent.md` that guides: copy, rename, register, migrate, test.
- **Effort**: ~2–3 hours
- **Value**: One `/new-service` command to walk through all steps

**A)** YES — Create new-service.agent.md
**B)** NO — SCAFFOLDING.md guide is sufficient; agent too much overhead
**C)**

**Answer:**

