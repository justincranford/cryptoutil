# Quality Gates Summary — fixes-v5

Deep analysis of changes made during fixes-v4 session. Identifies deficiencies in recently changed/created files.

See [quality-gates-details.md](quality-gates-details.md) for per-finding data.

---

## HIGH Priority — Fix Immediately (5 findings)

1. **[F-1.1: poll.go — nil conditionFn panic](#f-11-pollgo--nil-conditionfn-panic)** — Passing nil causes nil pointer dereference panic.

2. **[F-1.6: poll_test.go — missing edge case tests](#f-16-poll_testgo--missing-edge-case-tests)** — 7+ untested edge cases (nil conditionFn, zero/negative timeout, zero interval, context deadline, etc.).

3. **[F-2.5: SPA testmain — inconsistent patterns](#f-25-spa-testmain--inconsistent-patterns)** — Wrong `waitForReady` signature (returns bool instead of error), missing `SetReady(true)`, inconsistent timeout values (30s/10s vs 10s/5s elsewhere).

4. **[F-2.9: PKI subtests missing t.Parallel()](#f-29-pki-subtests-missing-tparallel)** — `TestCAServer_HealthEndpoints_EdgeCases` subtests violate mandatory `t.Parallel()` requirement.

5. **[F-2.11: server_highcov_test.go still uses time.Sleep](#f-211-server_highcov_testgo-still-uses-timesleep)** — 4 instances of `time.Sleep` for server readiness — was not migrated to `poll.Until()`.

---

## MEDIUM Priority — Fix Soon (13 findings)

6. **[F-1.2: poll.go — no zero/negative timeout validation](#f-12-pollgo--no-zeronegative-timeout-validation)**
7. **[F-1.3: poll.go — no zero/negative interval validation](#f-13-pollgo--no-zeronegative-interval-validation)** — Zero interval causes CPU spin loop.
8. **[F-2.1–2.4: Identity testmain local constants duplicate magic](#f-21-24-identity-testmain-local-constants-duplicate-magic)** — 4 files declare readyTimeout/checkInterval/shutdownTimeout locally instead of using magic constants.
9. **[F-2.6: jose-ja testmain — inline const inside function](#f-26-jose-ja-testmain--inline-const-inside-function)** — Non-standard Go style, different timeout value (5s vs 10s).
10. **[F-2.7: PKI highcov — inline 5*time.Second magic values](#f-27-pki-highcov--inline-5timesecond-magic-values)** — 4 occurrences of `5 * time.Second` not referencing magic constants.
11. **[F-2.12–2.14: Demo files — time.Sleep for server startup](#f-212-214-demo-files--timesleep-for-server-startup)** — 3 files use `time.Sleep` instead of `poll.Until()` for server readiness.
12. **[F-3.2: magic_testing.go — TestNegativeDuration not a Duration type](#f-32-magic_testinggo--testnegativeduration-not-a-duration-type)** — Untyped integer constant `-1` is misleading.
13. **[F-4.1: identity healthcheck/poller.go — duplicate polling implementation](#f-41-identity-healthcheckpollergo--duplicate-polling-implementation)** — Separate polling loop with exponential backoff; violates DRY.
14. **[F-4.2: Demo package — 20+ scattered constants](#f-42-demo-package--20-scattered-constants)** — Duration/timeout constants not in magic package.

---

## LOW Priority — Fix When Convenient (8 findings)

15. **[F-1.4: poll.go — timeout error not wrapped with sentinel](#f-14-pollgo--timeout-error-not-wrapped-with-sentinel)**
16. **[F-1.5: poll.go — context not checked before first conditionFn call](#f-15-pollgo--context-not-checked-before-first-conditionfn-call)**
17. **[F-2.8: PKI highcov — missing copyright header](#f-28-pki-highcov--missing-copyright-header)**
18. **[F-2.10: PKI highcov — duplicated HTTP client creation](#f-210-pki-highcov--duplicated-http-client-creation)**
19. **[F-3.3: magic_testing.go — redundant overlapping timeout names](#f-33-magic_testinggo--redundant-overlapping-timeout-names)**
20. **[F-3.4: magic_testing.go — nolint:stylecheck without bug reference](#f-34-magic_testinggo--nolintstylecheck-without-bug-reference)**
21. **[F-4.4: identity/rp uses server_test package inconsistently](#f-44-identityrp-uses-server_test-package-inconsistently)**
22. **[F-1.7: Copyright header inconsistency across files](#f-17-copyright-header-inconsistency-across-files)**
