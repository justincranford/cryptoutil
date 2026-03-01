# Quizme v4 — Deferred Decisions

**Purpose**: Items deferred from quizme-v3. Review each and mark your answer.
**Format**: **A** = YES/include, **B** = NO/skip, **C** = DEFER again.

---

## Section 1 — Pre-commit Linter Deferrals (from quizme-v3)

### Candidate 1: trivy (Container / Dependency Vulnerability)

Scans container images and Go/Python/Java dependencies for CVEs. Broader than govulncheck (multi-language, OS packages in containers).
- **Trigger**: `aquasecurity/trivy-action` or `repo: local`, runs `trivy fs --exit-code 1 --severity HIGH,CRITICAL .`
- **Cost**: Moderate (~10–30s); downloads vuln DB on first run
- **Previous concern**: Too slow for pre-commit

**A)** YES — Add trivy filesystem scan pre-commit hook
**B)** NO — govulncheck + OWASP sufficient; trivy too slow for pre-commit
**C)**

**Answer:**

### Candidate 2: semgrep (Multi-language SAST)

Pattern-based static analysis for Go, Python, Java, and Dockerfiles with curated rule sets.
- **Cost**: Slow (~30–120s for large codebases). Better as CI-only tool.
- **Note**: Already have `ci-sast` workflow — semgrep would add pre-commit gate

**A)** YES — Add semgrep as pre-commit hook
**B)** NO — Too slow for pre-commit; CI/CD (ci-sast workflow) is sufficient
**C)**

**Answer:**

### Candidate 3: codespell (Typo Detection)

Finds common typos in code comments, strings, and docs across all text files.
- **Trigger**: `codespell-project/codespell`, very fast (~1–2s)
- **Notable**: Low false-positive rate; only known common misspellings

**A)** YES — Add codespell pre-commit hook
**B)** NO — Not a significant problem; not worth the noise
**C)**

**Answer:**

### Candidate 4: editorconfig-checker (EditorConfig Compliance)

Validates all files conform to `.editorconfig` rules (indent, line endings, charset).
- **Prerequisite**: `.editorconfig` must exist in project root
- **Note**: Project may not have `.editorconfig` yet — would need to create it

**A)** YES — Create `.editorconfig` and add editorconfig-checker hook
**B)** NO — Editor settings enforced by existing formatters; not needed
**C)**

**Answer:**

---

## Section 2 — Java Toolchain Deferral (from quizme-v3)

### Tool 1: ArchUnit (Architecture Rule Enforcement)

Tests Java architectural constraints: package dependencies, layer isolation, naming conventions.
- **Plugin**: ArchUnit via JUnit test class
- **Effort**: Highest of Java options (requires writing architecture test class)
- **Current need**: Enforce Gatling simulation class conventions, prevent test utility leakage
- **Previous concern**: High effort; defer until other tools are in place (Spotless, Checkstyle, Error Prone, JaCoCo all now approved)

With Phase 8 toolchain approved and stable, ArchUnit becomes the remaining gap.

**A)** YES — Write ArchUnit test class for `test/load/` architecture rules
**B)** NO — Load test architecture is simple; enforcement overhead not justified
**C)**

**Answer:**

---

## Section 3 — skeleton-template Deferral (from quizme-v3)

### Item 1: CICD placeholder detection lint rule

Add `cicd validate-skeleton` lint check: scan non-skeleton directories for unreplaced "skeleton", "Skeleton", "SKELETON" strings.
- **Effort**: ~2–4 hours (Go validator in `cmd/cicd/` + pre-commit hook)
- **Value**: Catches copy-paste-forgot-to-rename errors automatically; complements template comments (Task 10.2)
- **Note**: Task 10.2 (template comment headers) will be done before this; question is whether automated CI detection is also needed

**A)** YES — Add placeholder detection as `cicd validate-skeleton` lint rule + pre-commit hook
**B)** NO — Template comments (Task 10.2) + code review are sufficient
**C)**

**Answer:**

