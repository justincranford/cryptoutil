# Quality Gates — Details

All items are **MANDATORY AND BLOCKING**.

Back to [quality-gates-summary.md](quality-gates-summary.md)

---

## QG-1: Linting — 2 goconst violations (exit code 1)

[↑ Summary item 1](quality-gates-summary.md#failing-quality-gates)

**Command:** `golangci-lint run`  
**Exit code:** 1 (FAIL)

### Violations

| File | Line | Linter | Message |
|------|------|---------|---------|
| `internal/apps/identity/notifications/service_test.go` | 27 | `goconst` | string `:memory:` has 2 occurrences, make it a constant |
| `internal/apps/template/service/client/user_auth_error_paths_test.go` | 39 | `goconst` | string `"user"` has 2 occurrences, make it a constant |

### Fix

Declare each repeated string as a named constant in the respective `_test.go` file:

```go
const testDSNMemory = ":memory:"  // notifications/service_test.go
const testUsername  = "user"      // user_auth_error_paths_test.go
```

---

## QG-2: Flaky Tests — 2 tests fail under concurrent load

[↑ Summary item 2](quality-gates-summary.md#failing-quality-gates)

**Observed during:** `go test ./... -shuffle=on`  
**Behavior:** Both tests pass when run in isolation (`-count=3`); intermittently fail in full parallel suite.

| Test | Package | Observed Error | Category |
|------|---------|----------------|----------|
| `TestCAServer_HandleOCSP` | `internal/apps/pki/ca/server` | Intermittent failure under concurrent load | Race / shared state |
| `TestPublicServerBase_StartAndShutdown` | `internal/apps/template/service/server` | `"0" is not greater than "0"` (port assertion) | Race: port not yet bound |

### Root Cause Hypothesis

- **`TestPublicServerBase_StartAndShutdown`**: Dynamic port 0 assignment race — test asserts `port > 0` before server completes bind.
- **`TestCAServer_HandleOCSP`**: Shared state or timing contention when CA server tests run concurrently across packages.

### Fix

- `TestPublicServerBase_StartAndShutdown`: Use `WaitForReady(ctx, timeout)` before asserting port value.
- `TestCAServer_HandleOCSP`: Audit for shared global/package-level state; ensure each parallel test instance has isolated DB and key material.

---

## QG-3: Infrastructure Coverage Below 98% — 51 packages

[↑ Summary item 3](quality-gates-summary.md#failing-quality-gates)

**Target:** ≥98%  
**Applies to:** `internal/shared/*`, `internal/cmd/cicd*`, `internal/apps/cicd/*`, `internal/apps/template/service/*`  
**Excluded:** test-helper sub-packages (`/testutil`, `/testing/*`, `/mocks`, `/keygenpooltest`)  
**Sorted by:** coverage ascending (lowest = highest urgency)

| Coverage | Covered | Total | Missing | Package |
|----------|---------|-------|---------|---------|
| 0.0% | 0 | 11 | 11 | `internal/apps/cicd/lint_go_mod` |
| 0.0% | 0 | 14 | 14 | `internal/apps/cicd/lint_golangci` |
| 0.0% | 0 | 16 | 16 | `internal/apps/cicd/lint_workflow` |
| 0.0% | 0 | 16 | 16 | `internal/shared/magic` |
| 0.0% | 0 | 56 | 56 | `internal/apps/cicd/lint_go/common` |
| 0.0% | 0 | 63 | 63 | `internal/shared/container` |
| 71.8% | 56 | 78 | 22 | `internal/cmd/cicd` |
| 86.5% | 498 | 576 | 78 | `internal/apps/template/service/config` |
| 88.4% | 987 | 1117 | 130 | `internal/shared/crypto/jose` |
| 88.9% | 511 | 575 | 64 | `internal/apps/template/service/server/businesslogic` |
| 95.0% | 474 | 499 | 25 | `internal/apps/template/service/server/realm` |
| 95.2% | 236 | 248 | 12 | `internal/apps/template/service/server/application` |
| 95.2% | 237 | 249 | 12 | `internal/apps/template/service/telemetry` |
| 95.2% | 257 | 270 | 13 | `internal/apps/template/service/server/listener` |
| 95.2% | 99 | 104 | 5 | `internal/apps/cicd/lint_go/circular_deps` |
| 95.3% | 221 | 232 | 11 | `internal/shared/crypto/certificate` |
| 95.3% | 81 | 85 | 4 | `internal/apps/cicd/lint_go/magic_usage` |
| 95.3% | 410 | 430 | 20 | `internal/apps/template/service/server/builder` |
| 95.4% | 206 | 216 | 10 | `internal/apps/template/service/cli` |
| 95.4% | 480 | 503 | 23 | `internal/apps/template/service/server/barrier` |
| 95.5% | 147 | 154 | 7 | `internal/apps/template/service/server` |
| 95.5% | 147 | 154 | 7 | `internal/apps/template/service/server/barrier/unsealkeysservice` |
| 95.5% | 85 | 89 | 4 | `internal/shared/util/files` |
| 95.5% | 107 | 112 | 5 | `internal/apps/cicd/lint_workflow/github_actions` |
| 95.8% | 113 | 118 | 5 | `internal/apps/template/service/server/tenant` |
| 95.8% | 68 | 71 | 3 | `internal/apps/cicd/lint_go/non_fips_algorithms` |
| 96.0% | 48 | 50 | 2 | `internal/apps/cicd/lint_go/insecure_skip_verify` |
| 96.2% | 100 | 104 | 4 | `internal/apps/cicd/lint_go/cgo_free_sqlite` |
| 96.3% | 79 | 82 | 3 | `internal/apps/cicd/lint_ports/legacy_ports` |
| 96.4% | 80 | 83 | 3 | `internal/apps/cicd/format_gotest/thelper` |
| 96.4% | 188 | 195 | 7 | `internal/shared/pool` |
| 96.6% | 56 | 58 | 2 | `internal/shared/apperr` |
| 96.6% | 56 | 58 | 2 | `internal/shared/database` |
| 96.6% | 171 | 177 | 6 | `internal/shared/crypto/tls` |
| 96.7% | 29 | 30 | 1 | `internal/apps/cicd/lint_go/magic_duplicates` |
| 96.8% | 30 | 31 | 1 | `internal/apps/cicd/lint_gotest/no_hardcoded_passwords` |
| 96.8% | 60 | 62 | 2 | `internal/apps/cicd/lint_go/no_unaliased_cryptoutil_imports` |
| 96.8% | 60 | 62 | 2 | `internal/shared/util/network` |
| 96.9% | 94 | 97 | 3 | `internal/shared/crypto/asn1` |
| 96.9% | 94 | 97 | 3 | `internal/shared/crypto/digests` |
| 97.1% | 33 | 34 | 1 | `internal/apps/cicd/lint_gotest/require_over_assert` |
| 97.1% | 34 | 35 | 1 | `internal/apps/cicd/lint_gotest/parallel_tests` |
| 97.2% | 314 | 323 | 9 | `internal/apps/template/service/server/repository` |
| 97.3% | 72 | 74 | 2 | `internal/apps/cicd/lint_compose/docker_secrets` |
| 97.3% | 73 | 75 | 2 | `internal/apps/cicd/lint_go/crypto_rand` |
| 97.5% | 77 | 79 | 2 | `internal/apps/cicd/lint_ports/host_port_ranges` |
| 97.5% | 119 | 122 | 3 | `internal/apps/cicd/lint_go_mod/outdated_deps` |
| 97.7% | 127 | 130 | 3 | `internal/apps/template/service/config/tls_generator` |
| 97.7% | 214 | 219 | 5 | `internal/shared/crypto/hash` |
| 97.8% | 174 | 178 | 4 | `internal/apps/template/service/server/apis` |
| 97.8% | 131 | 134 | 3 | `internal/apps/template/service/client` |

---

## QG-4: Production Coverage Below 95% — 63 packages

[↑ Summary item 4](quality-gates-summary.md#failing-quality-gates)

**Target:** ≥95%  
**Applies to:** `internal/apps/{pki,jose,cipher,sm,identity,cryptoutil}/*`  
**Sorted by:** coverage ascending (lowest = highest urgency)

| Coverage | Covered | Total | Missing | Package |
|----------|---------|-------|---------|---------|
| 0.0% | 0 | 30 | 30 | `internal/apps/identity/cmd/main/rs` |
| 0.0% | 0 | 34 | 34 | `internal/apps/sm/kms/cmd` |
| 0.0% | 0 | 36 | 36 | `internal/apps/identity/cmd/main/spa-rp` |
| 0.0% | 0 | 42 | 42 | `internal/apps/identity/rp` |
| 0.0% | 0 | 42 | 42 | `internal/apps/identity/spa` |
| 0.0% | 0 | 42 | 42 | `internal/apps/jose/ja` |
| 0.0% | 0 | 42 | 42 | `internal/apps/pki/ca` |
| 0.0% | 0 | 42 | 42 | `internal/apps/sm/kms` |
| 0.0% | 0 | 60 | 60 | `internal/apps/identity/cmd/main/idp` |
| 0.0% | 0 | 60 | 60 | `internal/apps/sm/kms/server` |
| 0.0% | 0 | 61 | 61 | `internal/apps/pki/ca/server/cmd` |
| 0.0% | 0 | 66 | 66 | `internal/apps/identity/cmd/main/authz` |
| 0.0% | 0 | 74 | 74 | `internal/apps/cipher/im/testing` |
| 0.0% | 0 | 101 | 101 | `internal/apps/identity/compose` |
| 0.0% | 0 | 102 | 102 | `internal/apps/identity/storage/fixtures` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/authz/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/idp/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/rp/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/rs/unified` |
| 0.0% | 0 | 110 | 110 | `internal/apps/identity/spa/unified` |
| 0.0% | 0 | 114 | 114 | `internal/apps/identity/server` |
| 0.0% | 0 | 187 | 187 | `internal/apps/identity/cmd` |
| 0.0% | 0 | 195 | 195 | `internal/apps/identity/cmd/main` |
| 0.0% | 0 | 285 | 285 | `internal/apps/identity/demo` |
| 0.0% | 0 | 330 | 330 | `internal/apps/sm/kms/server/handler` |
| 0.0% | 0 | 367 | 367 | `internal/apps/sm/kms/server/repository/orm` |
| 0.0% | 0 | 647 | 647 | `internal/apps/sm/kms/server/application` |
| 0.0% | 0 | 993 | 993 | `internal/apps/demo` |
| 5.9% | 1 | 17 | 16 | `internal/apps/cryptoutil` |
| 7.3% | 3 | 41 | 38 | `internal/apps/sm/kms/server/demo` |
| 13.4% | 20 | 149 | 129 | `internal/apps/identity/repository` |
| 19.3% | 45 | 233 | 188 | `internal/apps/identity/cmd/main/hardware-cred` |
| 21.3% | 10 | 47 | 37 | `internal/apps/identity/rp/server/config` |
| 31.0% | 13 | 42 | 29 | `internal/apps/cipher/im` |
| 32.1% | 17 | 53 | 36 | `internal/apps/identity/idp/server/config` |
| 32.8% | 19 | 58 | 39 | `internal/apps/identity/rs/server/config` |
| 36.8% | 225 | 612 | 387 | `internal/apps/sm/kms/server/businesslogic` |
| 37.7% | 20 | 53 | 33 | `internal/apps/identity/spa/server/config` |
| 38.8% | 19 | 49 | 30 | `internal/apps/identity/authz/server/config` |
| 39.7% | 125 | 315 | 190 | `internal/apps/identity/mfa` |
| 40.8% | 89 | 218 | 129 | `internal/apps/pki/ca/server` |
| 58.5% | 427 | 730 | 303 | `internal/apps/sm/kms/server/middleware` |
| 60.5% | 303 | 501 | 198 | `internal/apps/identity/idp` |
| 61.3% | 739 | 1206 | 467 | `internal/apps/identity/authz` |
| 61.5% | 91 | 148 | 57 | `internal/apps/identity/rs` |
| 65.8% | 48 | 73 | 25 | `internal/apps/identity/rp/server` |
| 66.7% | 54 | 81 | 27 | `internal/apps/identity/spa/server` |
| 69.1% | 112 | 162 | 50 | `internal/apps/identity/rs/server` |
| 72.4% | 118 | 163 | 45 | `internal/apps/identity/idp/server` |
| 73.0% | 127 | 174 | 47 | `internal/apps/identity/authz/server` |
| 75.6% | 201 | 266 | 65 | `internal/apps/identity/idp/auth` |
| 77.4% | 1041 | 1345 | 304 | `internal/apps/identity/idp/userauth` |
| 77.8% | 486 | 625 | 139 | `internal/apps/identity/repository/orm` |
| 79.6% | 133 | 167 | 34 | `internal/apps/pki/ca/cli` |
| 81.8% | 27 | 33 | 6 | `internal/apps/pki/ca/demo` |
| 83.3% | 70 | 84 | 14 | `internal/apps/cipher/im/server/apis` |
| 85.5% | 447 | 523 | 76 | `internal/apps/identity/authz/clientauth` |
| 85.7% | 66 | 77 | 11 | `internal/apps/identity/process` |
| 86.5% | 90 | 104 | 14 | `internal/apps/cipher/im/server` |
| 89.0% | 65 | 73 | 8 | `internal/apps/pki/ca/bootstrap` |
| 91.2% | 552 | 605 | 53 | `internal/apps/pki/ca/api/handler` |
| 92.2% | 83 | 90 | 7 | `internal/apps/pki/ca/intermediate` |
| 93.5% | 43 | 46 | 3 | `internal/apps/cipher/im/server/config` |

---

## QG-5: Zero Coverage — Infrastructure Packages With No Tests

[↑ Summary item 5](quality-gates-summary.md#failing-quality-gates)

These infrastructure/utility packages (≥98% target) have 0% coverage and **no test files**.
Packages that are themselves test-infrastructure helpers are excluded.

| Statements | Package | Context |
|------------|---------|---------|
| 63 | `internal/shared/container` | Testcontainers helpers; invoked indirectly by integration tests |
| 56 | `internal/apps/cicd/lint_go/common` | Shared utilities for all lint_go sub-linters; no test file |
| 16 | `internal/apps/cicd/lint_workflow` | Top-level entry function; sub-package `github_actions` is at 95.5% |
| 16 | `internal/shared/magic` | Constants-only; may be exempt from coverage requirement |
| 14 | `internal/apps/cicd/lint_golangci` | Top-level entry function; sub-package `golangci_config` is at 98.9% |
| 11 | `internal/apps/cicd/lint_go_mod` | Top-level entry function; sub-package `outdated_deps` is at 97.5% |

---

## QG-6: Mutation Testing — Configuration Mismatch and Incomplete Scope

[↑ Summary item 6](quality-gates-summary.md#failing-quality-gates)

### Configuration Mismatch

| Config File | Setting | Current Value | Architecture Requirement |
|-------------|---------|---------------|--------------------------|
| `.gremlins.yml` | `threshold.efficacy` | **85%** | **≥95% mandatory, ≥98% ideal** |
| `.gremlins.yaml` | `threshold-efficacy` | **70.0%** | **≥95% mandatory** |
| `.gremlins.yaml` | `threshold-mcover` | **60.0%** | Implied ≥95% |

Both config files must be updated to enforce the architecture-mandated ≥95% efficacy threshold.

### Available Mutation Data

Source: `test-output/phase7/mutation-report.txt`

| Package | Killed | Lived | Not Covered | Timed Out | Efficacy | Mutator Coverage | Status |
|---------|--------|-------|-------------|-----------|----------|------------------|--------|
| `internal/cmd/cicd/lint_deployments` | 127 | 2 | 1 | 42 | **98.45%** | 99.23% | ✅ PASS |

**All other packages:** no mutation testing has been performed.

### Required Actions

1. Fix `.gremlins.yml` → set `threshold.efficacy: 95` (≥95% mandatory).
2. Fix `.gremlins.yaml` → set `threshold-efficacy: 95.0` and `threshold-mcover: 95.0`.
3. Run `gremlins unleash --tags=!integration` for each package that currently meets ≥95% code coverage.
4. Fix any packages where test efficacy falls below 95%.
