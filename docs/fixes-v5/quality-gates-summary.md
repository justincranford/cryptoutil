# Quality Gates Summary — fixes-v5

Deep analysis of changes made during fixes-v4 session. Identifies deficiencies in recently changed/created files.

See [quality-gates-details.md](quality-gates-details.md) for per-finding data.

---

## HIGH Priority — Fix Immediately (5 findings)

Ordered by ease of fix (easiest first) within severity.

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 1 | F-1.1 | Easy | [poll.go — nil conditionFn panic](#f-11-pollgo--nil-conditionfn-panic) — 1-line guard |
| 2 | F-2.9 | Easy | [PKI subtests missing t.Parallel()](#f-29-pki-subtests-missing-tparallel) — add t.Parallel() calls |
| 3 | F-2.11 | Medium | [server_highcov_test.go still uses time.Sleep](#f-211-server_highcov_testgo-still-uses-timesleep) — replace 4 sleeps |
| 4 | F-2.5 | Medium | [SPA testmain — inconsistent patterns](#f-25-spa-testmain--inconsistent-patterns) — align with other services |
| 5 | F-1.6 | Medium | [poll_test.go — missing edge case tests](#f-16-poll_testgo--missing-edge-case-tests) — write 7+ test cases |

---

## MEDIUM Priority — Fix Soon (9 findings)

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 6 | F-1.2 | Easy | [poll.go — no zero/negative timeout validation](#f-12-pollgo--no-zeronegative-timeout-validation) — 1 guard |
| 7 | F-1.3 | Easy | [poll.go — no zero/negative interval validation](#f-13-pollgo--no-zeronegative-interval-validation) — 1 guard |
| 8 | F-3.2 | Easy | [magic_testing.go — TestNegativeDuration not a Duration type](#f-32-magic_testinggo--testnegativeduration-not-a-duration-type) — 1-line type change |
| 9 | F-2.1–2.4 | Easy | [Identity testmain local constants duplicate magic](#f-21-24-identity-testmain-local-constants-duplicate-magic) — 4-file find-replace |
| 10 | F-2.6 | Easy | [jose-ja testmain — inline const inside function](#f-26-jose-ja-testmain--inline-const-inside-function) — move consts |
| 11 | F-2.7 | Easy | [PKI highcov — inline 5*time.Second magic values](#f-27-pki-highcov--inline-5timesecond-magic-values) — replace with magic |
| 12 | F-2.12–2.14 | Medium | [Demo files — time.Sleep for server startup](#f-212-214-demo-files--timesleep-for-server-startup) — 3 files |
| 13 | F-4.2 | Medium | [Demo package — 20+ scattered constants](#f-42-demo-package--20-scattered-constants) — move to magic |
| 14 | F-4.1 | Complex | [identity healthcheck/poller.go — duplicate polling](#f-41-identity-healthcheckpollergo--duplicate-polling-implementation) — architectural decision |

---

## LOW Priority — Fix When Convenient (8 findings)

| # | ID | Ease | Finding |
|---|-----|------|---------|
| 15 | F-2.8 | Easy | [PKI highcov — missing copyright header](#f-28-pki-highcov--missing-copyright-header) |
| 16 | F-3.4 | Easy | [magic_testing.go — nolint:stylecheck without bug reference](#f-34-magic_testinggo--nolintstylecheck-without-bug-reference) |
| 17 | F-1.4 | Easy | [poll.go — timeout error not wrapped with sentinel](#f-14-pollgo--timeout-error-not-wrapped-with-sentinel) |
| 18 | F-1.5 | Easy | [poll.go — context not checked before first conditionFn call](#f-15-pollgo--context-not-checked-before-first-conditionfn-call) |
| 19 | F-2.10 | Easy | [PKI highcov — duplicated HTTP client creation](#f-210-pki-highcov--duplicated-http-client-creation) |
| 20 | F-4.4 | Easy | [identity/rp uses server_test package inconsistently](#f-44-identityrp-uses-server_test-package-inconsistently) |
| 21 | F-3.3 | Medium | [magic_testing.go — redundant overlapping timeout names](#f-33-magic_testinggo--redundant-overlapping-timeout-names) |
| 22 | F-1.7 | Complex | [Copyright header inconsistency across files](#f-17-copyright-header-inconsistency-across-files) |
