# Quality Gates Summary

All items below are **MANDATORY AND BLOCKING**. None may be deferred.

See [quality-gates-details.md](quality-gates-details.md) for full per-package data.

---

## Failing Quality Gates

1. **[QG-1: Linting — 2 goconst violations (exit code 1)](#qg-1-linting--2-goconst-violations-exit-code-1)** — `golangci-lint run` exits 1; must be zero issues before commit.

2. **[QG-2: Flaky Tests — 2 tests fail under concurrent load](#qg-2-flaky-tests--2-tests-fail-under-concurrent-load)** — Two tests observed failing during full `go test ./... -shuffle=on` runs; pass in isolation, indicating race/timing defects.

3. **[QG-3: Infrastructure Coverage Below 98% — 51 packages](#qg-3-infrastructure-coverage-below-98--51-packages)** — `internal/shared/*`, `internal/cmd/cicd/*`, `internal/apps/cicd/*`, and `internal/apps/template/service/*` require ≥98%. Ranging from 0% to 97.9%.

4. **[QG-4: Production Coverage Below 95% — 63 packages](#qg-4-production-coverage-below-95--63-packages)** — `internal/apps/{pki,jose,cipher,sm,identity,cryptoutil}/*` require ≥95%. Ranging from 0% to 93.5%.

5. **[QG-5: Zero Coverage — Infrastructure Packages With No Tests](#qg-5-zero-coverage--infrastructure-packages-with-no-tests)** — Several infrastructure packages have 0% coverage with real non-trivial code and no test files at all.

6. **[QG-6: Mutation Testing — Configuration Mismatch and Incomplete Scope](#qg-6-mutation-testing--configuration-mismatch-and-incomplete-scope)** — Gremlins threshold set to 85% in `.gremlins.yml` vs. architecture requirement of ≥95%. Only 1 package (`lint_deployments`) has mutation data. All other packages: untested.

---

## Passing Quality Gates

- `go build ./...` — clean
- `go build -tags e2e,integration ./...` — clean
- `go run ./cmd/cicd lint-deployments validate-all` — all 65 validators pass
- `go test ./...` — all tests pass when run without `-shuffle=on` interference (see QG-2 for shuffle failures)
