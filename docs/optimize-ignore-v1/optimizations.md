# Ignore And Exclusion Optimization v1

## Scope

This document analyzes and optimizes:

- `.gitignore`
- `.dockerignore`
- `.nuclei-ignore`
- `.vscode/settings.json` (`files.exclude`, `search.exclude`, `files.watcherExclude`)
- Other present/missing ignore files that can reduce context size and editor/agent workload

Primary goal: reduce unnecessary file indexing, search space, filesystem watch load, build context transfer, and long-running Copilot chat context pressure in large sessions (especially under `docs/framework-v2/`).

## Baseline Findings

## What is already good

- `.gitignore` already excludes most coverage/test artifacts and reports (`coverage*`, `test-output/`, `workflow-reports/`, `*.exe`, etc.).
- `.dockerignore` excludes many heavy directories (`docs/`, `test/`, `deployments/`, reports, testdata).
- VS Code excludes already include key artifact patterns such as `test-output`, `coverage_*`, `*_coverage`, `*.test`, and Python cache directories.
- `test/load/.gitignore` exists, which is good for nested load-test artifacts.

## High-impact gaps discovered

- Workspace root contains many generated artifacts (`coverage_*`, `*.test.exe`, helper outputs), so exclusion quality is critical.
- `.dockerignore` has duplicates (`.vscode/`, `.idea/`, `env/`, `test/`) and mixed root-prefixed patterns (`./...`) that add maintenance overhead.
- VS Code exclusions appear to miss some high-churn paths:
  - `workflow-reports/**`
  - `load-reports/**`
  - `e2e-reports/**`
  - `dast-reports/**`
  - `.semgrep/**`
  - `.zap/**`
  - `.cicd/**`
  - some root coverage variants like `all_coverage`
- `.gitignore` does not explicitly include cache dirs such as `.ruff_cache/`, `.mypy_cache/`.
- No workspace `.ignore`/`.rgignore` exists to further reduce ripgrep-based scans.
- `.git/info/exclude` is present but default-only (not used for local-machine-only generated files).

## Research Notes (Why these changes matter)

- VS Code uses `files.exclude`, `search.exclude`, and watcher exclusions to reduce Explorer rendering, search scope, and file watch overhead.
- VS Code Search can also honor ignore files (`.gitignore`), and this affects scan breadth.
- Docker build context is transferred recursively; `.dockerignore` directly reduces transfer size and build latency.
- Nuclei supports exclusion controls, but overusing static ignore entries can hide meaningful findings; precision and review cadence are important.

## Decision Checklist (Semantic Groups)

Use this as a choose-in/choose-out checklist. Each item is independent unless noted.

## Group A: `.gitignore` Hygiene (Git status noise, disk churn, context cleanup)

- [x] Add Python and scanner cache ignores.
  - Benefit: less accidental churn and status scans.
  - Candidate additions:

```gitignore
.ruff_cache/
.mypy_cache/
```

- [x] Add optional local scanner output ignores without hiding tracked rule files.
  - Keep tracked rule files (`.semgrep/rules/go-testing.yml`, `.zap/rules.tsv`) intact.
  - Candidate pattern style:

```gitignore
# Keep rules tracked, ignore runtime output folders/files if they appear.
.semgrep/tmp/
.semgrep/results/
.zap/reports/
.zap/session/
```

- [x] Add generic temp suffixes if these are truly ephemeral in this repo.
  - Use only if you confirm no intended tracked `*.tmp`/`*.bak` artifacts.

```gitignore
*.tmp
*.bak
```

- [x] Keep root helper script ignore policy explicit.
  - You currently ignore all root-level `/*.py` as temp helpers. Keep this if intentional.
  - If you want to allow future real root Python tools, replace with narrower patterns.

## Group B: `.dockerignore` Simplification And Build Context Performance

- [x] Remove duplicate entries.
  - Duplicates found: `.vscode/`, `.idea/`, `env/`, `test/`.
  - Benefit: lower maintenance and fewer pattern conflicts.

- [x] Normalize root patterns (remove `./` prefixes unless needed for clarity).
  - Example normalization:

```dockerignore
*.swp
*.log
*.tmp
*.bak
.DS_Store
Thumbs.db
coverage/
test/
docs/
scripts/
```

- [x] Consider excluding additional root heavy artifacts explicitly.
  - Candidate additions:

```dockerignore
workflow-reports/
test-output/
all_coverage
coverage_*
*.test.exe
```

- [x] Keep `.git` strategy intentional.
  - You currently do not ignore `.git` (commented out), likely for metadata generation.
  - Keep as-is if build uses commit metadata.
  - If not needed, ignoring `.git` can significantly reduce context for local/remote builders.

## Group C: VS Code Exclusions (Explorer/Search/Watcher CPU and memory)

- [x] Add missing report directories to all three maps:
  - `files.exclude`
  - `search.exclude`
  - `files.watcherExclude`

```json
"**/workflow-reports/**": true,
"**/load-reports/**": true,
"**/e2e-reports/**": true,
"**/dast-reports/**": true
```

- [x] Add scanner/cache folders to all three maps.

```json
"**/.semgrep/**": true,
"**/.zap/**": true,
"**/.cicd/**": true,
"**/.mypy_cache/**": true
```

- [x] Add root-heavy coverage aliases where needed.

```json
"**/all_coverage": true,
"**/coverage/**": true
```

- [ ] Keep source folders visible (`docs/`, `internal/`, etc.) but consider optional deep-doc exclusion during focused sessions.
  - For long agent sessions unrelated to framework docs, temporarily exclude:

```json
"**/docs/framework-v2/**": true
```

- Use this only when not actively editing those docs.

## Group D: VS Code/Copilot Behavior Settings (not ignore files, but high leverage)

- [x] Add search settings that honor ignore files consistently.

```json
"search.useIgnoreFiles": true,
"search.useGlobalIgnoreFiles": true,
"search.useParentIgnoreFiles": true,
"search.followSymlinks": false
```

- [ ] Optionally reduce Explorer load by respecting `.gitignore` in tree view.

```json
"explorer.excludeGitIgnore": true
```

- [ ] For large autonomous sessions, disable semantic code search auto-expansion if not needed.

```json
"github.copilot.chat.codesearch.enabled": false
```

- Keep this optional. It can reduce context churn but may reduce automatic file discovery quality.

## Group E: Nuclei Ignore Strategy (`.nuclei-ignore`)

- [ x Keep ignore list minimal and reasoned.
  - Current entries are mostly informational/expected-dev findings.
  - Add ownership and review cadence comments.

- [x] Add expiry metadata to each suppression (manual policy).

```text
http-missing-security-headers  # reason: false positive in this environment; owner: security; review-by: 2026-06-30
```

- [x] Prefer CLI filter strategy for broad classes where possible (`-severity`, `-exclude-tags`) rather than growing static ignore list forever.

## Group F: Missing Ignore Files You Can Add

- [ ] Add `.ignore` at repo root (for ripgrep-family tooling).
  - Useful for fast local scans and tool calls that honor `.ignore`.

```ignore
test-output/
workflow-reports/
coverage*
coverage_*
*_coverage
all_coverage
*.test.exe
```

- [x] Optionally add `.rgignore` only if you need ripgrep-specific behavior that differs from `.ignore`.
  - If not needed, prefer only one to avoid policy drift.

- [ ] Use `.git/info/exclude` for machine-local noise (do not commit).
  - Good for personal temp outputs that should not become team policy.

## Group G: Remove/Keep Decisions (Important)

- [x] Keep tracked rule/config files inside ignored-looking folders:
  - `.semgrep/rules/go-testing.yml`
  - `.zap/rules.tsv`
  - `test-output/phase9-unsealkeys-migration/migration-analysis.md`

- [x] Do not blanket-ignore entire `docs/`.

- [x] Do not over-ignore security findings globally in `.nuclei-ignore`; prefer explicit, reviewed exceptions.

## Suggested Implementation Order

- [ ] 1. Clean and normalize `.dockerignore` duplicates/pattern style.
- [ ] 2. Add missing VS Code exclusions to all three maps.
- [ ] 3. Add `.ignore` for ripgrep/tool scan reduction.
- [ ] 4. Add targeted `.gitignore` cache entries (`.ruff_cache/`, `.mypy_cache/`).
- [ ] 5. Add Nuclei suppression governance comments (owner/review-by).
- [ ] 6. Optionally tune Copilot/search behavior settings.

## Concrete Example Bundle (Conservative)

This is a safe starting set with high benefit and low risk.

### `.gitignore`

```gitignore
.ruff_cache/
.mypy_cache/
```

### `.dockerignore`

```dockerignore
# remove duplicates and normalize style
workflow-reports/
test-output/
all_coverage
coverage_*
*.test.exe
```

### `.vscode/settings.json` (add to all relevant exclude maps)

```json
"**/workflow-reports/**": true,
"**/load-reports/**": true,
"**/e2e-reports/**": true,
"**/dast-reports/**": true,
"**/.semgrep/**": true,
"**/.zap/**": true,
"**/.cicd/**": true,
"**/all_coverage": true,
"**/coverage/**": true
```

### `.ignore` (new file)

```ignore
test-output/
workflow-reports/
coverage*
coverage_*
*_coverage
all_coverage
*.test.exe
```

## Validation Checklist After Changes

- [ ] `git status --porcelain` remains clean for expected files.
- [ ] VS Code Search excludes artifact folders by default.
- [ ] CPU usage from file watching decreases during large test/report generation.
- [ ] Copilot long-running sessions attach fewer irrelevant artifact files.
- [ ] Docker build context transfer size decreases (watch `transferring context` logs).

## Notes For Your Framework v2 Session

When actively working the `docs/framework-v2/` planning flow, keep that folder visible. For unrelated coding sessions, temporarily excluding `docs/framework-v2/**` can materially reduce search and agent context churn.
