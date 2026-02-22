# Quality Gates Summary

All items below are **MANDATORY AND BLOCKING**. None may be deferred.

See [quality-gates-details.md](quality-gates-details.md) for full per-package data.

---

## Failing Quality Gates

1. **[QG-1: Linting — 2 goconst violations (exit code 1)](#qg-1-linting--2-goconst-violations-exit-code-1)** — `golangci-lint run` exits 1; must be zero issues before commit.

2. **[QG-2: Flaky Tests — 1 test fails under concurrent load](#qg-2-flaky-tests--1-test-fails-under-concurrent-load)** — One test observed failing during full `go test ./... -shuffle=on` runs; passes in isolation, indicating a race/timing defect.

3. **[QG-3: Infrastructure Coverage Below 98% — 50 packages](#qg-3-infrastructure-coverage-below-98--50-packages)** — `internal/shared/*`, `internal/cmd/cicd/*`, `internal/apps/cicd/*`, and `internal/apps/template/service/*` require ≥98%. `internal/shared/magic/` excluded (constants only). Ranging from 0% to 97.9%. All fixes must use table-driven tests.

4. **[QG-4: Production Coverage Below 95% — 63 packages (17 to implement, 46 WON'T IMPLEMENT)](#qg-4-production-coverage-below-95--63-packages-17-to-implement-46-wont-implement)** — `internal/apps/{pki,jose,cipher,sm,identity,cryptoutil}/*` require ≥95%. Identity and pki-ca deferred pending service-template migration. All fixes must use table-driven tests.

5. **[QG-6: Mutation Testing — Configuration Mismatch and Incomplete Scope](#qg-6-mutation-testing--configuration-mismatch-and-incomplete-scope)** — Two conflicting gremlins configs with thresholds well below ≥95% mandatory. Only 1 package (`lint_deployments`) has mutation data. Consolidate to `.gremlins.yaml` and raise thresholds.

---

## Passing Quality Gates

- `go build ./...` — clean
- `go build -tags e2e,integration ./...` — clean
- `go run ./cmd/cicd lint-deployments validate-all` — all 65 validators pass
- `go test ./...` — all tests pass when run without `-shuffle=on` interference (see QG-2 for shuffle failures)
