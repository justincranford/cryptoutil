# Workflow Runtime Analysis

> **Generated**: 2025-07-17
> **Repository**: [justincranford/cryptoutil](https://github.com/justincranford/cryptoutil)
> **Branch**: main
> **Sample size**: Last 100 runs per workflow (on `main` branch)

## Table of Contents

- [1. Workflow Catalog](#1-workflow-catalog)
- [2. Historical Runtime Statistics](#2-historical-runtime-statistics)
  - [2.1 Success Runs](#21-success-runs)
  - [2.2 Failure Runs](#22-failure-runs)
  - [2.3 Overall Pass/Fail Ratios](#23-overall-passfail-ratios)
- [3. Current Storage Footprint](#3-current-storage-footprint)
- [4. Optimization Recommendations](#4-optimization-recommendations)
  - [4.1 CI - Race Condition Detection](#41-ci---race-condition-detection)
  - [4.2 CI - Coverage Collection](#42-ci---coverage-collection)
  - [4.3 CI - Mutation Testing](#43-ci---mutation-testing)
  - [4.4 CI - End-to-End Testing](#44-ci---end-to-end-testing)
  - [4.5 CI - DAST Security Testing](#45-ci---dast-security-testing)
  - [4.6 CI - Fuzz Testing](#46-ci---fuzz-testing)
  - [4.7 CI - Quality Testing](#47-ci---quality-testing)
  - [4.8 CI - Load Testing](#48-ci---load-testing)
  - [4.9 CI - SAST Security Testing](#49-ci---sast-security-testing)
  - [4.10 CI - Identity Validation](#410-ci---identity-validation)
  - [4.11 CI - Benchmark Testing](#411-ci---benchmark-testing)
  - [4.12 CI - GitLeaks Secrets Scan](#412-ci---gitleaks-secrets-scan)
  - [4.13 CICD - Lint Deployments](#413-cicd---lint-deployments)

---

## 1. Workflow Catalog

| # | Workflow | File | Lines | Link |
|---|---------|------|-------|------|
| 1 | CI - Race Condition Detection | `ci-race.yml` | 138 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-race.yml) |
| 2 | CI - Coverage Collection | `ci-coverage.yml` | 310 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-coverage.yml) |
| 3 | CI - Mutation Testing | `ci-mutation.yml` | 149 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-mutation.yml) |
| 4 | CI - End-to-End Testing | `ci-e2e.yml` | 257 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-e2e.yml) |
| 5 | CI - DAST Security Testing | `ci-dast.yml` | 870 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-dast.yml) |
| 6 | CI - Fuzz Testing | `ci-fuzz.yml` | 332 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-fuzz.yml) |
| 7 | CI - Quality Testing | `ci-quality.yml` | 590 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-quality.yml) |
| 8 | CI - Load Testing | `ci-load.yml` | 382 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-load.yml) |
| 9 | CI - SAST Security Testing | `ci-sast.yml` | 476 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-sast.yml) |
| 10 | CI - Identity Validation | `ci-identity-validation.yml` | 129 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-identity-validation.yml) |
| 11 | CI - Benchmark Testing | `ci-benchmark.yml` | 178 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-benchmark.yml) |
| 12 | CI - GitLeaks Secrets Scan | `ci-gitleaks.yml` | 86 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/ci-gitleaks.yml) |
| 13 | CICD - Lint Deployments | `cicd-lint-deployments.yml` | 68 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/cicd-lint-deployments.yml) |
| 14 | Release - Automated Pipeline | `release.yml` | 364 | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/release.yml) |

**Non-CI workflows** (not analyzed for optimization):

| Workflow | File | Link |
|---------|------|------|
| Copilot coding agent | `copilot-setup-steps.yml` | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/copilot-setup-steps.yml) |
| Dependency Graph | `dependency-graph.yml` | [Workflow](https://github.com/justincranford/cryptoutil/actions/workflows/dependency-graph.yml) |

---

## 2. Historical Runtime Statistics

Data collected via `gh run list --workflow <name> --limit 100 --branch main`.

### 2.1 Success Runs

| # | Workflow | ✓ Count | Min | Max | Avg | Median Est. |
|---|---------|---------|-----|-----|-----|-------------|
| 1 | CI - Race Condition Detection | 9 | 53.4m | 79.6m | **65.0m** | ~64m |
| 2 | CI - Coverage Collection | 10 | 12.1m | 13.8m | **12.8m** | ~12.8m |
| 3 | CI - Mutation Testing | 31 | 8.2m | 12.0m | **9.6m** | ~9.4m |
| 4 | CI - End-to-End Testing | 11 | 7.7m | 24.5m | **9.4m** | ~8.5m |
| 5 | CI - DAST Security Testing | 16 | 7.0m | 13.7m | **8.2m** | ~7.8m |
| 6 | CI - Fuzz Testing | 30 | 3.1m | 19.5m | **5.6m** | ~4.8m |
| 7 | CI - Quality Testing | 5 | 5.2m | 5.7m | **5.5m** | ~5.5m |
| 8 | CI - Load Testing | 8 | 5.0m | 7.1m | **5.4m** | ~5.3m |
| 9 | CI - SAST Security Testing | 15 | 3.7m | 7.4m | **4.2m** | ~4.0m |
| 10 | CI - Identity Validation | 4 | 3.7m | 4.2m | **4.0m** | ~4.0m |
| 11 | CI - Benchmark Testing | 37 | 0.4m | 9.5m | **1.5m** | ~1.2m |
| 12 | CI - GitLeaks Secrets Scan | 99 | 0.3m | 10.3m | **0.8m** | ~0.5m |
| 13 | CICD - Lint Deployments | 24 | 0.3m | 2.6m | **0.6m** | ~0.5m |

**Totals**: Successful CI suite run ≈ **122m** cumulative (all 13 workflows in parallel ≈ **65m** wall-clock, gated by Race CI).

### 2.2 Failure Runs

| # | Workflow | ✗ Count | Min | Max | Avg |
|---|---------|---------|-----|-----|-----|
| 1 | CI - Race Condition Detection | 91 | 0.8m | 78.4m | 15.2m |
| 2 | CI - Coverage Collection | 90 | 0.8m | 20.4m | 4.0m |
| 3 | CI - Mutation Testing | 69 | 0.8m | 11.3m | 1.7m |
| 4 | CI - End-to-End Testing | 89 | 0.5m | 9.0m | 1.8m |
| 5 | CI - DAST Security Testing | 84 | 5.0m | 11.8m | 5.5m |
| 6 | CI - Fuzz Testing | 70 | 0.4m | 17.8m | 1.0m |
| 7 | CI - Quality Testing | 95 | 0.6m | 20.6m | 3.0m |
| 8 | CI - Load Testing | 92 | 0.4m | 10.9m | 2.3m |
| 9 | CI - SAST Security Testing | 85 | 0.0m | 20.0m | 4.2m |
| 10 | CI - Identity Validation | 96 | 0.0m | 12.3m | 2.5m |
| 11 | CI - Benchmark Testing | 63 | 0.4m | 0.8m | 0.5m |
| 12 | CI - GitLeaks Secrets Scan | 1 | 0.4m | 0.4m | 0.4m |
| 13 | CICD - Lint Deployments | 21 | 0.3m | 0.9m | 0.4m |

**Note**: High failure counts reflect iterative development with frequent intermediate pushes during active coding sessions. Most failures are early-abort (path-ignore triggering, linting failures, compilation errors) — not flaky tests.

### 2.3 Overall Pass/Fail Ratios

| # | Workflow | Total | ✓ Pass | ✗ Fail | Pass Rate |
|---|---------|-------|--------|--------|-----------|
| 1 | CI - Race Condition Detection | 100 | 9 | 91 | 9% |
| 2 | CI - Coverage Collection | 100 | 10 | 90 | 10% |
| 3 | CI - Mutation Testing | 100 | 31 | 69 | 31% |
| 4 | CI - End-to-End Testing | 100 | 11 | 89 | 11% |
| 5 | CI - DAST Security Testing | 100 | 16 | 84 | 16% |
| 6 | CI - Fuzz Testing | 100 | 30 | 70 | 30% |
| 7 | CI - Quality Testing | 100 | 5 | 95 | 5% |
| 8 | CI - Load Testing | 100 | 8 | 92 | 8% |
| 9 | CI - SAST Security Testing | 100 | 15 | 85 | 15% |
| 10 | CI - Identity Validation | 100 | 4 | 96 | 4% |
| 11 | CI - Benchmark Testing | 100 | 37 | 63 | 37% |
| 12 | CI - GitLeaks Secrets Scan | 100 | 99 | 1 | 99% |
| 13 | CICD - Lint Deployments | 45 | 24 | 21 | 53% |

---

## 3. Current Storage Footprint

| Resource | Count / Size |
|----------|-------------|
| Total workflow runs | 8,822 |
| Total artifacts | 6,998 |
| Active caches | 128 |
| Cache size | 2.31 GB (of 10 GB GitHub limit) |
| Repository size | 143 MB |

**GitHub Actions storage limits** (Free tier):

| Resource | Limit |
|----------|-------|
| Artifact & log retention | 90 days (configurable 1-90) |
| Cache total | 10 GB per repo |
| Cache eviction | LRU, entries >7 days unused |
| Workflow run logs | Retained with runs |
| Minutes/month | 2,000 (Free), 3,000 (Pro) |

---

## 4. Optimization Recommendations

### 4.1 CI - Race Condition Detection

> **Current avg success**: 65.0m | **Current config**: `go test -race -timeout=25m -count=2 ./internal/... ./scripts/...` + PostgreSQL service
> [ci-race.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-race.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Reduce `-count=2` to `-count=1`** — A single race-detected pass still catches most races; -count=2 nearly doubles execution time | -45% | ~36m | Low | Low — fewer statistical samples per run, but practical detection remains high |
| 2 | **Add concurrency group** — Cancel in-flight runs when a newer push arrives (`concurrency: { group: ${{ github.workflow }}-${{ github.ref }}, cancel-in-progress: true }`) | Saves wasted minutes | ~65m (per run) | Low | None |
| 3 | **Package sharding via matrix strategy** — Split `./internal/...` into 4 matrix shards (e.g., `internal/apps`, `internal/shared`, `internal/cmd`, `scripts`) running in parallel | -60% wall-clock | ~26m | Medium | Medium — shard imbalance if packages are uneven |
| 4 | **Use larger runner** (`ubuntu-latest-4-core` or `ubuntu-latest-8-core`) — Race detector is CPU-bound; more cores = faster ThreadSanitizer instrumentation | -25% to -50% | ~33-49m | Low (cost) | None — requires paid plan |
| 5 | **Target only changed packages** — Use `dorny/paths-filter` or `tj-actions/changed-files` to compute affected Go packages and test only those | -30% to -90% | Variable | Medium | Medium — may miss transitive race conditions |
| 6 | **Remove PostgreSQL service** — If race tests can use SQLite in-memory (`--dev` mode), eliminate ~1-2m container startup | -3% | ~63m | Medium | Low — need to verify all tests work without PostgreSQL |
| 7 | **Cache Go build artifacts** — Cache `~/.cache/go-build` and `~/go/pkg/mod` to avoid recompilation (if not already cached by `actions/setup-go`) | -5% | ~62m | Low | None |
| 8 | **Exclude known-stable packages** — Skip race testing for pure-config, constants, and generated code packages (`magic/`, `api/`, `model/`) | -15% | ~55m | Low | Low — constants cannot have races |
| 9 | **Run only on PRs** (not every push to main) — Race detection on PR validation only; main branch inherits PR results | -50% runs (halves total CI minutes) | 65m per run | Low | Low — delayed detection of merge-introduced races |
| 10 | **Schedule as nightly job** — Move race detection to a cron schedule (`schedule: cron: '0 3 * * *'`) instead of per-push | -90% runs | 65m per nightly | Low | Medium — races discovered next morning, not at push time |

**Recommended combination**: Options 1 + 2 + 8 → ~30m success runs with minimal risk.

---

### 4.2 CI - Coverage Collection

> **Current avg success**: 12.8m | **Current config**: Multiple `go test -coverprofile` runs + PostgreSQL
> [ci-coverage.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-coverage.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded coverage runs on new push | Saves wasted minutes | 12.8m per run | Low | None |
| 2 | **Single `go test -coverprofile` invocation** — Use `-coverpkg=./...` with a single test command instead of multiple runs | -15% | ~10.9m | Medium | Low — merged profiles may differ slightly |
| 3 | **Use larger runner** (4-core) — Parallel test execution benefits from more cores | -20% | ~10.2m | Low (cost) | None |
| 4 | **Matrix split by package group** — Run coverage in 3 parallel matrix jobs (apps, shared, cmd) | -50% wall-clock | ~6.4m | Medium | Low — need final merge step |
| 5 | **Cache test dependencies** — Pre-warm GORM, testcontainer images, Go build cache | -10% | ~11.5m | Low | None |
| 6 | **Target changed packages for PRs** — Full coverage on main, incremental on PRs | -30% on PRs | ~9.0m | Medium | Low — PR coverage is per-change focus |
| 7 | **Reduce artifact retention** — Change from 7 days to 3 days for coverage profiles (save storage, not runtime) | 0% runtime | 12.8m | Low | None |
| 8 | **Remove PostgreSQL if possible** — Use SQLite for coverage tests that don't require PostgreSQL-specific behavior | -8% | ~11.8m | Medium | Medium — cross-DB behavior differences |
| 9 | **Pre-compile test binaries** — `go test -c` and cache binaries between workflow runs | -5% | ~12.2m | Medium | Low |
| 10 | **Deduplicate with quality workflow** — Share compiled artifacts between Coverage and Quality workflows via cache | -5% | ~12.2m | High | Medium — cache key coordination |

**Recommended combination**: Options 1 + 2 + 4 → ~6m wall-clock with parallel matrix.

---

### 4.3 CI - Mutation Testing

> **Current avg success**: 9.6m | **Current config**: `gremlins unleash` (sequential, all packages) + PostgreSQL
> [ci-mutation.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-mutation.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded mutation runs | Saves wasted minutes | 9.6m per run | Low | None |
| 2 | **Cache gremlins binary** — Currently `go install` from source each run; cache the compiled binary | -10% | ~8.6m | Low | None |
| 3 | **Matrix parallelization by package** — Split packages into 4 shards; each runs gremlins on its subset | -60% wall-clock | ~3.8m | Medium | Low — gremlins supports package targeting |
| 4 | **Use larger runner** (4-core) — Gremlins generates and runs mutants; more cores = faster execution | -25% | ~7.2m | Low (cost) | None |
| 5 | **Target only changed packages** — Run mutation testing only on packages with source changes | -40% to -80% | Variable | Medium | Medium — may miss cross-package mutation escapes |
| 6 | **Reduce mutation scope** — Exclude test-only packages, generated code, constants from mutation | -15% | ~8.2m | Low | None — non-productive mutations eliminated |
| 7 | **Run mutation on PR only** (not on push to main) — Save CI minutes; main inherits PR validation | -50% runs | 9.6m per run | Low | Low |
| 8 | **Pre-compile test binaries** — Cache `go test -c` outputs for gremlins to reuse | -10% | ~8.6m | Medium | Low |
| 9 | **Lower timeout from 60m to 20m** — Current timeout is 3x the gremlins step timeout (45m); tighten to 2x avg | 0% runtime, faster fail | 9.6m | Low | Low |
| 10 | **Schedule as weekly job** — Mutation testing is slower-feedback; nightly or weekly is acceptable | -85% runs | 9.6m per run | Low | Medium — delayed mutation feedback |

**Recommended combination**: Options 1 + 2 + 3 → ~4m wall-clock with matrix and cached binary.

---

### 4.4 CI - End-to-End Testing

> **Current avg success**: 9.4m | **Current config**: Docker Compose build + service startup + E2E test suite
> [ci-e2e.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-e2e.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded E2E runs | Saves wasted minutes | 9.4m per run | Low | None |
| 2 | **Cache Docker build layers** — Use `docker/build-push-action` with GitHub Actions cache backend | -20% | ~7.5m | Medium | Low |
| 3 | **Pre-pull Docker images in parallel** — Already has `docker-images-pull` action; verify all images listed | -5% | ~8.9m | Low | None |
| 4 | **Use larger runner** (4-core) — Docker builds and test execution benefit from more cores | -20% | ~7.5m | Low (cost) | None |
| 5 | **Parallelize independent E2E suites** — If E2E tests are modular, run service tests in matrix | -40% wall-clock | ~5.6m | High | Medium — shared state complications |
| 6 | **Health check optimization** — Reduce `start-period`, `interval`, and `retries` for faster service readiness detection | -5% | ~8.9m | Low | Low — risk of premature health success |
| 7 | **Cache Go binary** — Build once, cache the binary; skip `go build` on subsequent runs if sources unchanged | -15% | ~8.0m | Medium | Low |
| 8 | **Run on PR only** (not every push to main) — E2E is expensive; PR-gating is sufficient | -50% runs | 9.4m per run | Low | Low |
| 9 | **Optimize compose startup order** — Ensure `depends_on` with `service_healthy` minimizes idle wait | -5% | ~8.9m | Low | None |
| 10 | **Reduce artifact retention** — Currently uploading compose logs; reduce retention from 7 to 1 day | 0% runtime (saves storage) | 9.4m | Low | None |

**Recommended combination**: Options 1 + 2 + 6 → ~7m with Docker layer caching.

---

### 4.5 CI - DAST Security Testing

> **Current avg success**: 8.2m | **Current config**: Docker Compose + Nuclei scanner + multiple scan stages (870 lines)
> [ci-dast.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-dast.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded DAST runs | Saves wasted minutes | 8.2m per run | Low | None |
| 2 | **Cache Nuclei templates** — Nuclei downloads templates on each run; cache `~/.nuclei-templates` | -10% | ~7.4m | Low | Low — templates may need updates |
| 3 | **Parallelize scan targets** — Split DAST scans into matrix (e.g., by severity level or endpoint group) | -30% wall-clock | ~5.7m | Medium | Low |
| 4 | **Use larger runner** (4-core) — Nuclei scanning benefits from more cores for concurrent requests | -15% | ~7.0m | Low (cost) | None |
| 5 | **Reduce scan scope** — Target only critical/high severity templates; skip informational scans | -25% | ~6.2m | Low | Medium — may miss medium-severity findings |
| 6 | **Cache Docker build layers** — Reuse service container images from prior builds | -15% | ~7.0m | Medium | Low |
| 7 | **Optimize service startup** — Faster health checks, smaller compose configuration | -5% | ~7.8m | Low | None |
| 8 | **Run on PR only** (not every push) — DAST is heavyweight; PR gating is sufficient | -50% runs | 8.2m per run | Low | Low |
| 9 | **Schedule as nightly/weekly** — DAST findings rarely change between commits | -90% runs | 8.2m per run | Low | Medium — delayed vulnerability detection |
| 10 | **Reduce artifact retention** — DAST outputs (SARIF, logs) to 3 days instead of 30 | 0% runtime (saves storage) | 8.2m | Low | None |

**Recommended combination**: Options 1 + 2 + 9 → nightly DAST with cached templates.

---

### 4.6 CI - Fuzz Testing

> **Current avg success**: 5.6m | **Current config**: Multiple fuzz targets with 15s+ fuzz time per target
> [ci-fuzz.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-fuzz.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded fuzz runs | Saves wasted minutes | 5.6m per run | Low | None |
| 2 | **Cache fuzz corpus** — Persist `testdata/fuzz/` corpus between runs for faster convergence | -15% | ~4.8m | Medium | None — strictly beneficial |
| 3 | **Matrix parallelization** — Run fuzz targets in parallel matrix jobs | -50% wall-clock | ~2.8m | Medium | Low |
| 4 | **Reduce fuzz duration** — Lower per-target fuzz time (e.g., 10s instead of 15s for CI; longer for nightly) | -30% | ~3.9m | Low | Low — reduced exploration |
| 5 | **Use larger runner** (4-core) — Fuzz engines benefit from parallelism | -20% | ~4.5m | Low (cost) | None |
| 6 | **Target only changed packages** — Fuzz only packages with source changes | -40% | Variable | Medium | Medium — may miss cross-package issues |
| 7 | **Run on PR only** (not every push) | -50% runs | 5.6m per run | Low | Low |
| 8 | **Schedule nightly with extended duration** — Short fuzz on PR (5s), extended nightly (60s) | -60% on PRs | ~2.2m PR / 15m nightly | Medium | None — best of both |
| 9 | **Pre-compile fuzz binaries** — `go test -c -fuzz=.` and cache compiled binaries | -10% | ~5.0m | Medium | Low |
| 10 | **Exclude stable/trivial fuzz targets** — Skip fuzz targets for formatters, validators that are well-tested | -20% | ~4.5m | Low | Low |

**Recommended combination**: Options 1 + 2 + 8 → 2m PR fuzz + 15m nightly deep fuzz.

---

### 4.7 CI - Quality Testing

> **Current avg success**: 5.5m | **Current config**: `golangci-lint` (2 runs: fix + check) + custom cicd linters + unit tests + PostgreSQL
> [ci-quality.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-quality.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded quality runs | Saves wasted minutes | 5.5m per run | Low | None |
| 2 | **Cache golangci-lint results** — `golangci-lint-action` has built-in cache; ensure `skip-cache: false` | -10% | ~5.0m | Low | None |
| 3 | **Single golangci-lint pass** — Currently runs `--fix` then re-runs check; in CI fix is unnecessary (no commit) — run check only | -20% | ~4.4m | Low | None — fix is only useful locally |
| 4 | **Use larger runner** (4-core) — golangci-lint heavily parallelizes; more cores = faster | -25% | ~4.1m | Low (cost) | None |
| 5 | **Parallel matrix: lint vs test** — Split linting and testing into separate matrix jobs | -30% wall-clock | ~3.9m | Medium | None |
| 6 | **Incremental linting** — Use `only-new-issues: true` for PRs (already available in golangci-lint-action) | -15% on PRs | ~4.7m | Low | Low — may hide pre-existing issues |
| 7 | **Reduce cicd lint scope** — Skip slow cicd linters for unchanged file types | -10% | ~5.0m | Medium | Low |
| 8 | **Remove duplicate Go test run** — If Coverage CI already runs tests, quality workflow can skip `go test` | -25% | ~4.1m | Low | Low — rely on Coverage workflow |
| 9 | **Pre-compile and cache cicd binary** — `go run ./cmd/cicd` compiles each time; pre-build and cache | -5% | ~5.2m | Medium | None |
| 10 | **Target changed files only** — Run linters only on changed files (golangci-lint supports `--new-from-rev`) | -20% on PRs | ~4.4m | Low | Low — may miss indirect issues |

**Recommended combination**: Options 1 + 3 + 5 → ~3.5m with single lint pass and parallel matrix.

---

### 4.8 CI - Load Testing

> **Current avg success**: 5.4m | **Current config**: Docker Compose + Java/Maven/Gatling load tests
> [ci-load.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-load.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded load test runs | Saves wasted minutes | 5.4m per run | Low | None |
| 2 | **Cache Maven dependencies** — Cache `~/.m2/repository` to avoid re-downloading Gatling + deps | -15% | ~4.6m | Low | None |
| 3 | **Cache Docker build layers** — Reuse service container builds | -10% | ~4.9m | Medium | Low |
| 4 | **Use larger runner** (4-core) — Gatling + service under test benefit from more CPU | -20% | ~4.3m | Low (cost) | None |
| 5 | **Reduce load test duration** — Shorter simulation for CI (smoke test); full load on schedule | -30% | ~3.8m | Low | Low — reduced confidence |
| 6 | **Pre-pull Docker images** — Ensure all images pre-pulled in parallel | -5% | ~5.1m | Low | None |
| 7 | **Run on PR only** (not every push) | -50% runs | 5.4m per run | Low | Low |
| 8 | **Schedule as nightly** — Load testing results don't change significantly per-commit | -90% runs | 5.4m per run | Low | Medium — delayed regression detection |
| 9 | **Pre-compile Gatling scenarios** — Maven compile once, cache target/ directory | -10% | ~4.9m | Medium | Low |
| 10 | **Reduce artifact retention** — Load test reports to 3 days | 0% runtime (saves storage) | 5.4m | Low | None |

**Recommended combination**: Options 1 + 2 + 5 → ~3.5m with Maven cache and shorter smoke tests.

---

### 4.9 CI - SAST Security Testing

> **Current avg success**: 4.2m | **Current config**: Multiple SAST tools (gosec, trivy, etc.) + 476 lines
> [ci-sast.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-sast.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded SAST runs | Saves wasted minutes | 4.2m per run | Low | None |
| 2 | **Cache vulnerability databases** — Trivy DB, gosec rules, etc. | -15% | ~3.6m | Low | Low — databases may become stale |
| 3 | **Matrix parallelization** — Run each SAST tool (gosec, trivy, semgrep) as separate matrix job | -40% wall-clock | ~2.5m | Medium | None |
| 4 | **Use larger runner** (4-core) — Scanning tools benefit from parallelism | -15% | ~3.6m | Low (cost) | None |
| 5 | **Target changed files only** — Incremental scanning on PRs, full scan on main | -30% on PRs | ~2.9m | Medium | Medium — may miss issues in unchanged files |
| 6 | **Reduce scan scope** — Exclude generated code, vendor, test files from SAST | -10% | ~3.8m | Low | Low |
| 7 | **Run on PR only** (not every push) | -50% runs | 4.2m per run | Low | Low |
| 8 | **Schedule as nightly** — SAST findings change infrequently | -80% runs | 4.2m per run | Low | Medium |
| 9 | **Consolidate SAST tools** — If tools overlap (e.g., gosec + trivy both check Go), reduce to non-overlapping set | -20% | ~3.4m | Medium | Medium — reduced coverage |
| 10 | **Reduce artifact retention** — SARIF outputs to 7 days (currently 30) | 0% runtime (saves storage) | 4.2m | Low | None |

**Recommended combination**: Options 1 + 2 + 3 → ~2.2m with parallel matrix and cached DBs.

---

### 4.10 CI - Identity Validation

> **Current avg success**: 4.0m | **Current config**: Custom cicd identity import validation
> [ci-identity-validation.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-identity-validation.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded validation runs | Saves wasted minutes | 4.0m per run | Low | None |
| 2 | **Cache compiled cicd binary** — Avoid `go run` recompilation each time | -15% | ~3.4m | Medium | None |
| 3 | **Use larger runner** (4-core) | -15% | ~3.4m | Low (cost) | None |
| 4 | **Path filtering** — Only run when `internal/apps/identity/` changes | -70% runs | 4.0m per run | Low | None — identity-specific validation |
| 5 | **Merge into Quality workflow** — Identity validation is a lint check; combine with existing quality CI | -100% (eliminated) | 0m | Medium | None — reduces workflow count |
| 6 | **Pre-compile cicd binary** — Share cached binary with Quality workflow | -10% | ~3.6m | Medium | None |
| 7 | **Reduce Go setup time** — Ensure Go module cache reused | -5% | ~3.8m | Low | None |
| 8 | **Run on PR only** | -50% runs | 4.0m per run | Low | Low |
| 9 | **Combine with lint-deployments** — Merge validation workflows to reduce overhead | -50% | ~2.0m | Medium | Low |
| 10 | **Skip if no identity code changes** — More granular path filter than workflow-level | -80% | 4.0m per run | Medium | None |

**Recommended combination**: Option 4 + 5 → Fold into Quality workflow with path filtering.

---

### 4.11 CI - Benchmark Testing

> **Current avg success**: 1.5m | **Current config**: `go test -bench` runs
> [ci-benchmark.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-benchmark.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded benchmark runs | Saves wasted minutes | 1.5m per run | Low | None |
| 2 | **Cache Go build** — Ensure cached compilation | -10% | ~1.4m | Low | None |
| 3 | **Target changed packages only** — Benchmark only modified packages | -30% | ~1.1m | Medium | Low |
| 4 | **Run on PR only** (not every push) | -50% runs | 1.5m per run | Low | Low |
| 5 | **Schedule as nightly** — Benchmarks need consistent environments for comparison; CI runners are variable | -90% runs | 1.5m per run | Low | Low — nightly is actually better for benchmarks |
| 6 | **Use consistent runner** — Pin to specific runner type for reproducible benchmarks | 0% runtime | 1.5m | Low | None — improves data quality |
| 7 | **Reduce benchmark iterations** — `-benchtime=1s` instead of default `-benchtime=1s -count=5` | -40% | ~0.9m | Low | Low — less statistical confidence |
| 8 | **Store results in GitHub Pages/artifact** — Enable regression comparison across runs | 0% runtime | 1.5m | Medium | None |
| 9 | **Exclude non-performance code** — Skip benchmarks for config, main, utilities | -15% | ~1.3m | Low | None |
| 10 | **Merge into coverage workflow** — Benchmarks could run alongside coverage in same pipeline | -100% (eliminated) | 0m separate | Medium | Low |

**Recommended combination**: Options 1 + 5 → Nightly scheduled benchmarks with concurrency group.

---

### 4.12 CI - GitLeaks Secrets Scan

> **Current avg success**: 0.8m | **Current config**: Gitleaks scan (86 lines, simplest workflow)
> [ci-gitleaks.yml](https://github.com/justincranford/cryptoutil/actions/workflows/ci-gitleaks.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded scans | Saves wasted minutes | 0.8m per run | Low | None |
| 2 | **Incremental scanning** — Scan only new commits since last scan (gitleaks has `--log-opts` for commit range) | -50% | ~0.4m | Low | Low — may miss secrets in older commits |
| 3 | **Cache gitleaks binary** — Avoid downloading each run | -10% | ~0.7m | Low | None |
| 4 | **Run on PR only** — Sufficient for pre-merge validation | -50% runs | 0.8m per run | Low | Low |
| 5 | **Use pre-commit hook** — Catch secrets locally before push; CI becomes verification | 0% CI runtime | 0.8m | Medium | None — defense in depth |
| 6 | **Reduce scan depth** — Limit to last N commits (e.g., `--depth=50`) for faster scans | -30% | ~0.6m | Low | Low — old secrets already committed |
| 7 | **Pin gitleaks version** — Avoid version download variance | -5% | ~0.8m | Low | None |
| 8 | **Merge into quality workflow** — Add gitleaks step to existing quality CI | -100% (eliminated) | 0m separate | Medium | None |
| 9 | **Already near-optimal** — At 0.8m avg, this is one of the fastest workflows | N/A | 0.8m | N/A | N/A |
| 10 | **Skip on docs-only changes** — Path filter already exists; verify it covers all non-code paths | -5% | ~0.8m | Low | None |

**Recommended combination**: Already efficient. Option 1 + 2 for marginal improvement.

---

### 4.13 CICD - Lint Deployments

> **Current avg success**: 0.6m | **Current config**: Custom Go-based deployment validators
> [cicd-lint-deployments.yml](https://github.com/justincranford/cryptoutil/actions/workflows/cicd-lint-deployments.yml)

| # | Optimization | Expected Improvement | New Est. Runtime | Effort | Risk |
|---|-------------|---------------------|-----------------|--------|------|
| 1 | **Add concurrency group** — Cancel superseded lint runs | Saves wasted minutes | 0.6m per run | Low | None |
| 2 | **Path filtering** — Only trigger on `deployments/` and `configs/` changes | -80% runs | 0.6m per run | Low | None |
| 3 | **Cache compiled cicd binary** — Pre-compile `go run ./cmd/cicd` | -20% | ~0.5m | Medium | None |
| 4 | **Merge into quality workflow** — Add as a step in the existing Quality CI | -100% (eliminated) | 0m separate | Medium | None |
| 5 | **Pre-build cicd binary as artifact** — Shared across workflows | -15% | ~0.5m | Medium | Low |
| 6 | **Already near-optimal** — At 0.6m avg, minimal room for improvement | N/A | 0.6m | N/A | N/A |
| 7 | **Run locally via pre-commit** — Already supported; CI becomes verification layer | 0% CI runtime | 0.6m | Low | None |
| 8 | **Use lighter-weight runner** — Could run on Alpine action container | -10% | ~0.5m | Medium | Low |
| 9 | **Skip Go setup** — Use pre-compiled binary instead of full Go installation | -25% | ~0.5m | High | Medium |
| 10 | **Reduce checkout depth** — `fetch-depth: 1` for faster checkout | -5% | ~0.6m | Low | None |

**Recommended combination**: Already efficient. Options 1 + 2 for targeted triggers.

---

## Summary: Top 5 Highest-Impact Changes Across All Workflows

| Priority | Change | Affected Workflows | Impact |
|----------|--------|-------------------|--------|
| 1 | **Add concurrency groups** to all workflows | All 13 | Prevents wasted CI minutes from superseded runs |
| 2 | **Reduce Race CI `-count=2` to `-count=1`** | Race | -45% on slowest workflow (65m → 36m) |
| 3 | **Move DAST/Load/Benchmark to nightly schedule** | DAST, Load, Benchmark | -90% run frequency for low-change-sensitivity workflows |
| 4 | **Matrix parallelization** for Race, Coverage, Mutation | Race, Coverage, Mutation | -50-60% wall-clock for top 3 slowest workflows |
| 5 | **Merge small workflows** (GitLeaks, Identity, CICD Lint) into Quality | GitLeaks, Identity, CICD Lint | Eliminate 3 workflow overheads (runner spin-up, checkout, Go setup) |

**Estimated impact of all Priority 1-5 changes**:

| Metric | Before | After |
|--------|--------|-------|
| Wall-clock (all parallel) | ~65m | ~26m |
| Cumulative CI minutes per push | ~122m | ~55m |
| Workflow count | 13 | 10 |
| Wasted minutes (superseded runs) | High | Near zero |
