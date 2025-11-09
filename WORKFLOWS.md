# GitHub Actions Workflow Review

This document captures a snapshot of the ten workflows under `.github/workflows/` and outlines actionable improvements, refactors, and potential additions. Recommendations are based on current workflow definitions (as of 2025-11-08) and on existing project instructions in `.github/instructions`.

## Cross-Cutting Opportunities

- **Introduce concurrency controls**: Add `concurrency: { group: "${{ github.workflow }}-${{ github.ref }}", cancel-in-progress: true }` to every workflow to prevent redundant runs when developers push multiple commits in quick succession.
- **Harmonize permissions blocks**: Several workflows omit an explicit `permissions:` stanza (e.g., `ci-benchmark.yml`, `ci-fuzz.yml`). Add least-privilege permissions (`contents: read`) and opt-in scopes (`security-events: write`) only where required.
- **Shared path filters**: The identical `paths-ignore` arrays could be centralized via YAML anchors or sourced from a reusable workflow invoked with `workflow_call`. This reduces drift when folders change.
- **Reusable setup actions**: `./.github/actions/go-setup` already enforces `go mod verify` and `go mod tidy`. Extend it to expose a cache key (e.g., `steps.go-setup.outputs.cache-hit`) so downstream steps can skip repeated `go build`/`docker build` tasks when nothing changed.
- **Artifact retention policy**: Most artifacts are stored for a single day. Consider lengthening retention for performance/security reports (benchmarks, DAST, load) to 7 days so trends can be inspected without re-running jobs.
- **Dependabot fast-exit**: For heavy workflows (DAST, load, fuzz, E2E), guard with `if: github.actor != 'dependabot[bot]'` to avoid burning CI minutes on dependency PRs that cannot address failures.
- **Notification & summarization**: Leverage reusable summary helpers (or `actions/github-script`) to emit consistent markdown tables across workflows. Today each job writes summaries manually; a helper keeps formatting uniform and eases future updates.

## Workflow-Specific Recommendations

### `ci-benchmark.yml` (CI - Benchmark Testing)

- Add a `benchstat` comparison step that reads the last successful benchmark artifact (store in `workflow-reports/benchmarks/last.json`) to detect regressions automatically.
- Upload the raw `go test -bench` output in machine-readable form (JSON via `go test -json -bench`) so future tooling can consume it.
- Gate benchmark runs behind an input or label for PRs to avoid spending time on unlabelled contributions.

### `ci-coverage.yml` (CI - Coverage Collection)

- Add `-coverpkg=./...` so cross-package invocations count toward coverage metrics.
- Enforce a minimum coverage threshold (e.g., fail if total coverage < configured value) and surface the threshold in the step summary.
- Feed `coverage-func.txt` into an HTML summary or comment bot, highlighting functions below target coverage, to shorten feedback loops.

### `ci-dast.yml` (CI - DAST Security Testing)

- Cache the Nuclei templates directory (`~/.local/share/nuclei-templates`) using `actions/cache@v4`; this trims scan time by 1-2 minutes per run.
- Split the Nuclei and ZAP stages into parallel jobs (`needs:`) so deep scans finish faster, especially for the scheduled Sunday run.
- Emit SARIF per tool (Nuclei, ZAP) and upload to the Security tab individually so findings can be triaged without unzipping artifacts.

### `ci-e2e.yml` (CI - End-to-End Testing)

- Convert the Docker compose lifecycle to a reusable action that wraps `docker compose up/down` with diagnostics; currently three separate composed actions duplicate log handling.
- Capture container exit codes after tests to highlight which service failed instead of relying solely on Go test output.
- Add a matrix for database backend (SQLite vs PostgreSQL) by wiring compose overrides, mirroring local developer modes.

### `ci-fuzz.yml` (CI - Fuzz Testing)

- Persist the fuzz corpus (`testdata/fuzz/...`) as an artifact so newly discovered inputs can seed future runs.
- Add an optional nightly schedule with longer `fuzztime` values to increase depth without slowing PR feedback.
- Ensure the `fuzz-test` composite action surfaces crashes as SARIF or junit for richer reporting in the GitHub UI.

### `ci-gitleaks.yml` (CI - Secrets Scan)

- Configure `gitleaks` to pull a managed baseline from a repository artifact so known test secrets can be tracked explicitly instead of relying on inline allowlists.
- Upload findings both as SARIF and as markdown summary tables to make remediation easier for contributors unfamiliar with SARIF viewers.

### `ci-load.yml` (CI - Load Testing)

- Parameterize Gatling profiles via matrix to run quick smoke tests on PRs and reserve heavy profiles (`stress`) for nightly schedules.
- Capture JVM heap and GC logs while Gatling runs, storing them alongside HTML reports for deeper analysis during incidents.
- After tests, post-process `results/**/index.html` into a markdown summary with 95th/99th percentile latencies surfaced.

### `ci-quality.yml` (CI - Quality Testing)

- Replace the manual `docker buildx create --use` shell snippet with `docker/setup-buildx-action@v3` for clarity and native logging.
- Add SBOM attestation (e.g., `syft attest` + `cosign attest`) after the build step to tie SBOMs directly to pushed images.
- Introduce a matrix that builds on multiple architectures (amd64, arm64) if cross-platform support is required—`docker buildx build` already supports it with limited changes.

### `ci-race.yml` (CI - Race Detection)

- Enable `-race -count=1 -run '^Test'` across targeted packages and add a slack/Teams notification hook when race detection fails, since these issues are time-sensitive.
- Optionally restrict the workflow to run on main + nightly schedule to control cost; expose a manual dispatch for regressions.

### `ci-sast.yml` (CI - Static Application Security Testing)

- Consolidate the multiple SAST tools into a matrix (e.g., `bandit`, `semgrep`, `gosec`) to run in parallel and reduce total time.
- Ensure every SARIF upload uses `if: always() && hashFiles('*.sarif') != ''` to avoid false negatives when a tool finds nothing (some jobs still rely on default behavior).

## Potential New Workflows

1. **Release Promotion Workflow**: Triggered on tag creation to build multi-arch images, sign them with Cosign, attach SBOMs, and publish changelog notes.
2. **Nightly Dependency Audit**: Uses `go list -u -m` and `npm/yarn audit` (for SPA assets) to surface upcoming upgrades via slack or issue comments.
3. **Performance Baseline Drift**: Nightly job that runs benchmarks/load tests, compares against stored baselines, and files an issue when degradation exceeds thresholds.
4. **Docs Link Checker**: Lightweight workflow that runs on documentation-only PRs to keep README and docs/ links valid without invoking full CI.
5. **Workflow Linting Gate**: A small workflow that runs `actionlint` on PRs touching `.github/workflows/**`, preventing syntax or path regressions before the heavier jobs run.

Implementing the above changes will reduce mean time to feedback, cut wasted CI cycles, and improve security/observability parity between local and hosted environments.
