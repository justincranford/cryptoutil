# GitHub Storage Cleanup Best Practices

> **Repository**: [justincranford/cryptoutil](https://github.com/justincranford/cryptoutil)

## Table of Contents

- [1. Storage Categories](#1-storage-categories)
- [2. Current State](#2-current-state)
- [3. Workflow Runs](#3-workflow-runs)
- [4. Artifacts](#4-artifacts)
- [5. Caches](#5-caches)
- [6. Logs](#6-logs)
- [7. Container Registry](#7-container-registry)
- [8. Retention Policies](#8-retention-policies)
- [9. Automated Cleanup Strategy](#9-automated-cleanup-strategy)
- [10. Go Cleanup Scripts](#10-go-cleanup-scripts)

---

## 1. Storage Categories

GitHub Actions generates storage in five categories:

| Category | Location | Default Retention | Configurable? | Counted Against Quota? |
|----------|----------|-------------------|---------------|----------------------|
| **Workflow Run Logs** | Actions > Workflow > Run | 90 days (Free/Pro), 400 days (Enterprise) | Yes (repo settings) | Yes (storage) |
| **Artifacts** | `actions/upload-artifact` | 90 days (default) | Yes (per-upload `retention-days`) | Yes (storage quota) |
| **Caches** | `actions/cache` / `actions/setup-go` | 7 days unused, 10 GB total | Evicted LRU | Yes (10 GB per repo) |
| **Container Registry** | `ghcr.io` | Indefinite | Manual/API cleanup | Yes (packages storage) |
| **Git Repository** | `.git` | Indefinite | `git gc`, shallow clones | Yes (repo size) |

## 2. Current State

| Resource | Current Value | Limit |
|----------|--------------|-------|
| Workflow runs | 8,822 | Unlimited (counted by storage) |
| Artifacts | 6,998 | 500 MB free storage |
| Active caches | 128 | 10 GB (currently 2.31 GB used) |
| Repository size | 143 MB | 5 GB soft limit |

## 3. Workflow Runs

### What consumes space

Each workflow run stores: job logs, step outputs, annotations, `GITHUB_STEP_SUMMARY` content, and attached artifacts.

### Cleanup strategies

| Strategy | Method | Scope |
|----------|--------|-------|
| **Repository settings** | Settings > Actions > General > "Artifact and log retention" | Applies to new runs |
| **API deletion** | `gh api -X DELETE repos/{owner}/{repo}/actions/runs/{run_id}` | Individual run |
| **Bulk deletion** | Script iterating over `gh run list` | Filtered by age/status/workflow |
| **Concurrency groups** | `concurrency: { cancel-in-progress: true }` | Prevents accumulation |

### Best practices

1. **Set retention to 7 days** (repo Settings > Actions > General) — 90 days is never needed
2. **Delete failed runs older than 3 days** — Failed runs from development iterations have no archival value
3. **Keep only last 3 successful runs** per workflow — For audit trail and comparison
4. **Add concurrency groups** to all workflows — Prevents superseded run accumulation
5. **Never delete runs from release tags** — May be needed for audit/compliance

## 4. Artifacts

### What consumes space

- Coverage profiles (`.out` files)
- Mutation test results (`.gremlins/`)
- SARIF security reports
- Benchmark results
- Docker Compose logs
- Load test reports (Gatling)
- Build outputs

### Cleanup strategies

| Strategy | Method | Impact |
|----------|--------|--------|
| **Reduce `retention-days`** per artifact | `upload-artifact@v5` with `retention-days: 3` | Prevents long-term accumulation |
| **API deletion** | `gh api -X DELETE repos/{owner}/{repo}/actions/artifacts/{id}` | Individual artifact |
| **Bulk deletion** | Script iterating over `gh api repos/{owner}/{repo}/actions/artifacts` | Filtered by age/size |
| **Skip non-essential uploads** | Remove `upload-artifact` for logs that can be found in run output | Zero storage |

### Recommended retention by type

| Artifact Type | Retention | Rationale |
|---------------|-----------|-----------|
| Coverage profiles | 1 day | Only needed for immediate PR review |
| Mutation results | 1 day | Same as coverage |
| SARIF reports | 7 days | Security compliance; also stored in Code Scanning |
| Benchmark results | 1 day | Comparison window |
| Benchmark baseline | 7 days | Comparison across runs |
| Docker/compose logs | 1 day | Only useful for debugging recent failures |
| Load test results | 1 day | Comparison window |
| Gatling HTML reports | 1 day | Detailed perf analysis for recent runs only |
| Build outputs | 1 day | Rebuilt on each run |

## 5. Caches

### What consumes space

- Go module cache (`~/go/pkg/mod`)
- Go build cache (`~/.cache/go-build`)
- golangci-lint cache
- Docker layer cache
- Maven dependencies (`.m2/repository`)

### Current state

128 active caches using 2.31 GB of 10 GB limit (23% utilized).

### Cleanup strategies

| Strategy | Method | Impact |
|----------|--------|--------|
| **Eviction (automatic)** | GitHub evicts caches unused for 7+ days, LRU when over 10 GB | Automatic |
| **API deletion** | `gh api -X DELETE repos/{owner}/{repo}/actions/caches/{cache_id}` | Individual cache |
| **Key-based deletion** | `gh cache delete --key <key>` | Pattern-based |
| **Branch cleanup** | Delete caches for merged/deleted branches | Reclaims orphaned caches |

### Best practices

1. **Use precise cache keys** — Include `hashFiles('go.sum')` to avoid stale caches
2. **Clean after branch merge** — Caches from feature branches persist; script to delete on merge
3. **Monitor cache hit rate** — Low hit rate = bad cache key; high miss = wasted upload time
4. **Limit cache scope** — Don't cache everything; only cache downloads (modules, DBs) not build outputs

## 6. Logs

### What consumes space

Workflow run logs are retained with the run. They cannot be managed independently of runs.

### Cleanup strategies

1. **Same as workflow runs** — Deleting a run deletes its logs
2. **Reduce log verbosity** — Use `--quiet` flags, suppress commands with `set +x`
3. **Avoid large step outputs** — Don't `cat` large files in workflow steps
4. **Use `GITHUB_STEP_SUMMARY`** — Summary output is lighter than full log output

## 7. Container Registry

If using `ghcr.io` (GitHub Container Registry):

| Strategy | Method |
|----------|--------|
| **Delete old image versions** | `gh api -X DELETE /user/packages/container/{package}/versions/{version_id}` |
| **Retention policy** | Package settings > "Expire versions older than X days" |
| **Untagged image cleanup** | Delete images without tags (orphaned layers) |

## 8. Retention Policies

### Recommended repository settings

Navigate to: **Settings > Actions > General > Artifact and log retention**

| Setting | Current (default) | Recommended |
|---------|-------------------|-------------|
| Artifact retention | 90 days | **7 days** |
| Log retention | 90 days | **7 days** |

### Per-workflow artifact overrides

```yaml
- uses: actions/upload-artifact@v5.0.0
  with:
    retention-days: 3  # Override repo-level default
```

### Retention schedule

| Resource | Dev/Feature branches | Main branch | Release tags |
|----------|---------------------|-------------|--------------|
| Workflow runs | 3 days | 7 days | Indefinite |
| Artifacts | 1 day | 1-7 days | 7 days |
| Caches | Auto-evict | Keep | N/A |

## 9. Automated Cleanup Strategy

### Tier 1: Preventive (reduce generation)

1. Add `concurrency: { group: ${{ github.workflow }}-${{ github.ref }}, cancel-in-progress: true }` to all workflows
2. Set `retention-days` on all `upload-artifact` steps
3. Reduce repo-level retention in Settings
4. Move infrequent workflows to nightly/weekly schedule

### Tier 2: Reactive (delete existing)

1. Delete workflow runs older than 7 days (script)
2. Delete all artifacts older than 7 days (script)
3. Delete caches for deleted/merged branches (script)
4. Delete failed workflow runs older than 3 days (script)

### Tier 3: Monitoring

1. Track cache utilization (API: `repos/{owner}/{repo}/actions/cache/usage`)
2. Track artifact count growth (API: `repos/{owner}/{repo}/actions/artifacts`)
3. Track workflow run accumulation (API: `repos/{owner}/{repo}/actions/runs`)
4. Alert when cache exceeds 7 GB (70% of 10 GB limit)

## 10. Go Cleanup Scripts

See `internal/apps/cicd/cleanup_github/` for Go-based automation scripts.

### Available commands

| Command | Description |
|---------|-------------|
| `go run ./cmd/cicd cleanup-runs` | Delete old workflow runs (configurable age threshold) |
| `go run ./cmd/cicd cleanup-artifacts` | Delete expired artifacts |
| `go run ./cmd/cicd cleanup-caches` | Delete stale/orphaned caches |
| `go run ./cmd/cicd cleanup-all` | Run all cleanup commands |

### Prerequisites

- `gh` CLI authenticated (`gh auth status`)
- Repository write permissions (for deletion)
- **Dry-run mode** enabled by default (pass `--confirm` to execute)

### Scheduling

Run cleanup weekly as a scheduled workflow:

```yaml
name: Maintenance - Storage Cleanup
on:
  schedule:
    - cron: '0 4 * * 0'  # Every Sunday at 4 AM UTC
  workflow_dispatch:

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6.0.0
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - run: go run ./cmd/cicd cleanup-all --confirm --max-age-days=7
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
