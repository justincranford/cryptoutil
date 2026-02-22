# Quality Gates — Details (fixes-v5)

Deep analysis of changes made during fixes-v4 session. All findings are **MANDATORY AND BLOCKING**.

Back to [quality-gates-summary.md](quality-gates-summary.md)

---

## HIGH Priority

### F-1.1: poll.go — nil conditionFn panic

- **File**: `internal/shared/util/poll/poll.go` line 20
- **Severity**: HIGH
- **Issue**: `Until()` does not check for nil `conditionFn`. Passing nil causes nil pointer dereference panic at runtime.
- **Fix**: Add `if conditionFn == nil { return fmt.Errorf("poll condition function must not be nil") }` at function entry.

### F-1.6: poll_test.go — missing edge case tests

- **File**: `internal/shared/util/poll/poll_test.go`
- **Severity**: HIGH
- **Issue**: Several edge cases not tested:
  - nil conditionFn (panics — no test)
  - Zero timeout (returns instantly with timeout error)
  - Negative timeout
  - Zero interval (CPU spin loop)
  - Negative interval
  - Context deadline exceeded (vs. canceled — different error type)
  - Condition function that uses the passed `ctx` parameter
- **Fix**: Add table-driven test cases for each. Coverage target: ≥98% (infrastructure package).

### F-2.5: SPA testmain — inconsistent patterns

- **File**: `internal/apps/identity/spa/server/testmain_test.go`
- **Severity**: HIGH
- **Issues**:
  1. `serverReadyTimeout = 30 * time.Second` vs 10s used by all other identity services — inconsistency.
  2. `shutdownTimeout = 10 * time.Second` vs 5s used everywhere else — inconsistency.
  3. `waitForReady()` returns `bool` instead of `error`, swallowing the poll error. All other identity testmain files return `error`.
  4. Missing `SetReady(true)` call — all other identity services call `testServer.SetReady(true)` after readiness.
  5. Missing `cryptoutilSharedMagic` import.
- **Fix**: Align with authz/idp/rs/rp patterns: return `error`, add `SetReady(true)`, use consistent timeouts.

### F-2.9: PKI subtests missing t.Parallel()

- **File**: `internal/apps/pki/ca/server/public_server_highcov_test.go` line 238
- **Severity**: HIGH
- **Issue**: `TestCAServer_HealthEndpoints_EdgeCases` subtests do NOT call `t.Parallel()`. Per ARCHITECTURE.md Section 10.2, `t.Parallel()` is REQUIRED on all test functions and subtests.
- **Fix**: Add `t.Parallel()` to each subtest.

### F-2.11: server_highcov_test.go still uses time.Sleep

- **File**: `internal/apps/pki/ca/server/server_highcov_test.go` lines 30, 57, 106, 117
- **Severity**: HIGH
- **Issue**: 4 instances of `time.Sleep(100ms)` and `time.Sleep(500ms)` for server readiness. These are exactly the pattern `poll.Until()` was created to replace. This file was missed during migration.
- **Fix**: Replace all `time.Sleep` waits with `poll.Until()`.

---

## MEDIUM Priority

### F-1.2: poll.go — no zero/negative timeout validation

- **File**: `internal/shared/util/poll/poll.go` line 20
- **Severity**: MEDIUM
- **Issue**: Zero or negative `timeout` causes the loop body to never execute. Returns generic "poll timed out after 0s" which is confusing.
- **Fix**: Return explicit error like `"poll timeout must be positive"` for `timeout <= 0`.

### F-1.3: poll.go — no zero/negative interval validation

- **File**: `internal/shared/util/poll/poll.go` line 20
- **Severity**: MEDIUM
- **Issue**: Zero `interval` causes a busy-wait spin loop (CPU hazard). Negative `interval` causes `time.After` to fire immediately, also a spin loop.
- **Fix**: Return explicit error for `interval <= 0`.

### F-2.1–2.4: Identity testmain local constants duplicate magic

- **Files**:
  - `internal/apps/identity/authz/server/testmain_test.go` lines 70-72
  - `internal/apps/identity/idp/server/testmain_test.go` lines 70-72
  - `internal/apps/identity/rs/server/testmain_test.go` lines 70-72
  - `internal/apps/identity/rp/server/testmain_test.go` lines 24-26
- **Severity**: MEDIUM
- **Issue**: Declare local `readyTimeout = 10s`, `checkInterval = 100ms`, `shutdownTimeout = 5s` — identical values to `magic.TestPollReadyTimeout`, `magic.TestPollReadyInterval`, `magic.TestServerShutdownTimeout` but not referencing magic.
- **Fix**: Replace with `cryptoutilSharedMagic.TestPollReadyTimeout`, etc.

### F-2.6: jose-ja testmain — inline const inside function

- **File**: `internal/apps/jose/ja/testmain_test.go` line 53
- **Severity**: MEDIUM
- **Issue**: Declares `pollTimeout = 5s` and `pollInterval = 100ms` as local consts inside `TestMain`. Non-standard Go style. `pollTimeout` is 5s vs magic's 10s — inconsistent.
- **Fix**: Use magic constants at package level.

### F-2.7: PKI highcov — inline 5*time.Second magic values

- **File**: `internal/apps/pki/ca/server/public_server_highcov_test.go` lines 49, 75, 104, 159, 229
- **Severity**: MEDIUM
- **Issue**: `Timeout: 5 * time.Second` for HTTP clients and shutdown contexts, repeated 4+ times without named constants.
- **Fix**: Use `cryptoutilSharedMagic.TestServerShutdownTimeout` or `cryptoutilSharedMagic.TimeoutHTTPHealthRequest`.

### F-2.12–2.14: Demo files — time.Sleep for server startup

- **Files**:
  - `internal/apps/demo/kms.go` line 159
  - `internal/apps/demo/integration.go` lines 355, 397
  - `internal/apps/demo/identity.go` line 297
- **Severity**: MEDIUM
- **Issue**: `time.Sleep` for server startup delays. Should use `poll.Until()` for server readiness instead of fixed waits.
- **Fix**: Replace with `poll.Until()` readiness checks where server publishes a port or health endpoint.

### F-3.2: magic_testing.go — TestNegativeDuration not a Duration type

- **File**: `internal/shared/magic/magic_testing.go` line 148
- **Severity**: MEDIUM
- **Issue**: `TestNegativeDuration = -1` is untyped integer constant, NOT `time.Duration`. Misleading name suggests it should be a duration.
- **Fix**: Change to `TestNegativeDuration time.Duration = -1 * time.Nanosecond` or rename to `TestNegativeValue`.

### F-4.1: identity healthcheck/poller.go — duplicate polling implementation

- **File**: `internal/apps/identity/healthcheck/poller.go` line 72
- **Severity**: MEDIUM
- **Issue**: `Poller.Poll()` implements its own polling loop with exponential backoff. Functionally similar to `poll.Until()` but with backoff. Two polling implementations violates DRY.
- **Fix**: Extend `poll.Until()` to optionally support exponential backoff, or document why separate implementation is needed.

### F-4.2: Demo package — 20+ scattered constants

- **Files**: `internal/apps/demo/integration.go` lines 47-67, `identity.go` lines 23-42
- **Severity**: MEDIUM
- **Issue**: Per ARCHITECTURE.md, "ALL magic constants MUST be consolidated in `internal/shared/magic/`". Demo packages have 20+ duration/timeout/credential constants locally.
- **Fix**: Move reusable constants to `magic/magic_demo.go` or reference existing magic constants.

---

## LOW Priority

### F-1.4: poll.go — timeout error not wrapped with sentinel

- **File**: `internal/shared/util/poll/poll.go` line 37
- **Severity**: LOW
- **Issue**: `fmt.Errorf("poll timed out after %v", timeout)` is not a sentinel error. Cannot be matched with `errors.Is()`.
- **Fix**: Define `var ErrTimeout = errors.New("poll timed out")` and return `fmt.Errorf("%w after %v", ErrTimeout, timeout)`.

### F-1.5: poll.go — context not checked before first conditionFn call

- **File**: `internal/shared/util/poll/poll.go` line 23
- **Severity**: LOW
- **Issue**: Context is only checked in `select` after `conditionFn` runs. If context is already canceled before `Until()`, conditionFn executes once unnecessarily.
- **Fix**: Check `ctx.Done()` before calling `conditionFn` in each iteration.

### F-2.8: PKI highcov — missing copyright header

- **File**: `internal/apps/pki/ca/server/public_server_highcov_test.go` line 1
- **Severity**: LOW
- **Issue**: No copyright header. Most other project files have one.

### F-2.10: PKI highcov — duplicated HTTP client creation

- **File**: `internal/apps/pki/ca/server/public_server_highcov_test.go` lines 47-55
- **Severity**: LOW
- **Issue**: Identical `http.Client` with `InsecureSkipVerify: true` created in every test function. Could use shared helper.

### F-3.3: magic_testing.go — redundant overlapping timeout names

- **File**: `internal/shared/magic/magic_testing.go` lines 80-100
- **Severity**: LOW
- **Issue**: `TimeoutTestServerReady = 30s` vs `TestPollReadyTimeout = 10s`; `TimeoutTestServerReadyRetryDelay = 500ms` vs `TestPollReadyInterval = 100ms`; `TestDefaultServerShutdownTimeout = 1min` vs `TestServerShutdownTimeout = 5s`. Naming inconsistency.

### F-3.4: magic_testing.go — nolint:stylecheck without bug reference

- **File**: `internal/shared/magic/magic_testing.go` lines 72-97
- **Severity**: LOW
- **Issue**: Multiple `//nolint:stylecheck // established API name` directives. Per project standards, `//nolint:` should only be used for documented linter bugs with GitHub issue reference.

### F-4.4: identity/rp uses server_test package inconsistently

- **File**: `internal/apps/identity/rp/server/testmain_test.go` line 5
- **Severity**: LOW
- **Issue**: Uses `package server_test` while all other identity services use `package server`. Causes `//nolint:gochecknoglobals` suppressions.

### F-1.7: Copyright header inconsistency across files

- **Files**: Various modified files
- **Severity**: LOW
- **Issue**: No consistent copyright header standard. Some have SPDX, some don't, some missing entirely.
