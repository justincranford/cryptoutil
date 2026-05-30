# Test Performance Report

Created: 2026-05-17

## Scope

This report summarizes the full-suite Go test run, the suite-only failures that appeared before the targeted fixes, and the main opportunities to reduce end-to-end test time.

Validation inputs used:

- `go test ./... -shuffle=on`
- Targeted reruns of `internal/apps-framework/service/server`, `internal/apps/sm-kms/server`, and `internal/shared/telemetry`
- Isolated reruns of the previously failing tests

## Executive Result

The full Go test suite passes after two suite-stability fixes:

- `internal/apps-framework/service/server` had a suite-only timeout in `TestPublicServerBase_StartAndMakeRequest` and a hang-prone `TestPublicServerBase_ErrChanPath`.
- `internal/apps/sm-kms/server` had a suite-only timeout in `TestHTTPPost/admin_shutdown_endpoint`.

Both failures reproduced only under suite load. Each test passed in isolation, which points to contention, startup delay, or an overly brittle lifecycle test rather than a deterministic functional bug.

## Package Hotspots

The slowest packages in the post-fix suite run were:

| Package | Time | Notes |
|---|---:|---|
| `cryptoutil/internal/apps-framework/service/server/application` | 80.755s | Heavy server lifecycle, TLS, and application wiring coverage |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_deployments` | 78.053s | Large lint pass over deployment artifacts |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/circular_deps` | 73.169s | Fitness checks are expensive by design |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/cgo_free_sqlite` | 67.360s | Static analysis plus broad project scanning |
| `cryptoutil/internal/apps/sm-kms/server` | 66.732s | Full service bootstrap, TLS, DB, barrier, and shutdown tests |
| `cryptoutil/internal/apps-framework/service/server/barrier` | 65.807s | Crypto-heavy barrier setup and teardown |
| `cryptoutil/internal/apps-framework/tls` | 58.143s | TLS generation and validation work |
| `cryptoutil/internal/apps/identity/mfa` | 50.773s | Integration-heavy auth flows |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness` | 49.863s | Aggregate fitness suite cost |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/check_skeleton_placeholders` | 40.665s | Repository-wide scan and validation |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/config_overlay_freshness` | 36.300s | Large file-family validation |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/gen_config_initialisms` | 36.122s | Large generated-config scan |
| `cryptoutil/internal/apps-tools/cicd_lint/docs_validation` | 35.354s | Broad docs consistency checks |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/database_key_uniformity` | 31.505s | Global repository scanning |
| `cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/dockerfile_labels` | 31.503s | Global repository scanning |
| `cryptoutil/internal/apps/identity-authz/server/apis` | 31.085s | API-level coverage with realistic setup |

Interpretation:

- The suite is dominated by a small number of integration-style packages and lint-fitness passes, not by ordinary unit tests.
- The heaviest product packages are the ones that bring up real TLS, databases, or application servers.
- `lint_fitness` is expected to be slower than ordinary tests, but it should stay clearly separated from quick feedback loops.

## Test Family / Class Hotspots

These are the groups that most strongly influenced runtime.

| Test Family | Package | Evidence | What it means |
|---|---|---|---|
| Telemetry service startup and flush | `cryptoutil/internal/shared/telemetry` | `TestTelemetryService_OTLPEnabledWithGRPC`, `TestTelemetryService_OTLPEnabledWithGRPCS`, `TestTelemetryService_GRPCEndpoint` all ran at about 13.5s each | Real exporter startup and forced flush/shutdown dominate the package |
| Telemetry retry/failure paths | `cryptoutil/internal/shared/telemetry` | `TestCheckSidecarHealthWithRetry_AllRetriesFail` ran at 10.00s; `TestTelemetryService_InvalidEndpoint` ran at 10.00s | The retry budget is the runtime, not the code path itself |
| Public server lifecycle | `cryptoutil/internal/apps-framework/service/server` | `TestPublicServerBase_StartAndMakeRequest` passed in isolation in 0.21s but failed in-suite at 11.90s before the fix | The test was brittle under suite contention |
| Err-channel coverage | `cryptoutil/internal/apps-framework/service/server` | `TestPublicServerBase_ErrChanPath` was the test still running when the 10-minute package timeout hit | The real shutdown path was too race-prone for a coverage test |
| SM-IM shutdown helper | `cryptoutil/internal/apps/sm-kms/server` | `TestHTTPPost/admin_shutdown_endpoint` failed in-suite after 5.04s but passed in isolation in 0.16s | The request timeout was too tight for suite load |

## Test-Level Notes

Observed behavior from the targeted reruns:

- `internal/shared/telemetry`: `TestParseLogLevel_AllLevels` is not the problem. It passed 10/10 times in a repeat run and completed in 0.023s when isolated.
- `internal/shared/telemetry`: the expensive tests are the ones that start exporters or wait out retry budgets. The reported log noise is a side effect of exercising failure and flush paths, not a correctness issue.
- `internal/apps-framework/service/server`: the err-channel test should not depend on the real Fiber shutdown sequence. The suite-only hang came from a coverage test that was too close to production shutdown timing.
- `internal/apps/sm-kms/server`: the shutdown endpoint is functionally correct, but the test allowed only 5 seconds for a path that can stretch under parallel suite load.

## Why The Suite Was Slow

The suite had three main cost centers:

1. Real server startup and TLS setup.
   Packages such as `internal/apps-framework/service/server/application`, `internal/apps-framework/service/server/barrier`, and `internal/apps/sm-kms/server` start real listeners, build TLS material, and initialize application state. Those tests are correct, but they are expensive.

2. Lint-fitness scans over large file sets.
   `lint_fitness` packages are repo-wide by design. They provide strong guardrails, but they are not good candidates for fast inner-loop feedback.

3. Retry-based or shutdown-based tests.
   Telemetry tests and server shutdown tests intentionally wait. Under a busy suite, those waits become the main runtime.

## Opportunities To Speed Up Overall Test Time

### High Priority

- Replace real network listeners with in-memory handler tests wherever the behavior under test is only routing or response shape.
- Keep shutdown and err-channel tests deterministic by stubbing the listener lifecycle instead of relying on production shutdown timing.
- Separate fast unit tests from heavy server/bootstrap tests so quick feedback does not wait on TLS and DB initialization.

### Medium Priority

- Reuse shared fixtures through `TestMain` where it is safe to do so, especially for TLS material and reusable DB setup.
- Reduce repeated generation of expensive TLS and crypto fixtures in packages that only need a single shared instance.
- Tighten retry loops so that failure-path tests use explicit, bounded budgets instead of generic package-level timeouts.

### Lower Priority

- Consider a separate fast-check workflow that skips `lint_fitness` and other repo-wide scans for inner-loop work.
- Consider caching or partitioning the largest lint-fitness groups if they become a recurring bottleneck for CI turn-around.

## Bottom Line

The suite is healthy now, but it is still expensive because it mixes fast unit coverage with real server bootstrap, TLS, and repo-wide static-analysis passes. The largest performance gains will come from keeping those heavyweight paths isolated and making lifecycle tests deterministic.
